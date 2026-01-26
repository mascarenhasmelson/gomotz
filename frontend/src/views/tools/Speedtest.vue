<template>
  <div class="speedtest-container">
   
      

    <!-- Main Button -->
    <div class="button-container">
      <button 
        @click="startSpeedtest" 
        :disabled="!complete"
        :class="['start-button', { 'running': !complete, 'pulse': complete }]"
      >
        <span v-if="complete" class="button-content">
          <svg class="button-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          Start Speed Test
        </span>
        <span v-else class="button-content">
          <svg class="spinner" viewBox="0 0 50 50">
            <circle cx="25" cy="25" r="20" fill="none" stroke-width="5"></circle>
          </svg>
          Testing in Progress...
        </span>
      </button>
      
      <p v-if="complete && testTime > 0" class="last-test-info">
        Last test completed in {{ testTime }} seconds
      </p>
    </div>

    <!-- Speed Cards -->
    <div class="speed-cards">
      <div class="speed-card download-card">
        <div class="card-header">
          <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
          </svg>
          <h2>Download Speed</h2>
        </div>
        <div class="speed-value" :class="{ 'active': !complete }">
          {{ downloadSpeed || "0.00" }}<span class="unit">Mb/s</span>
        </div>
        <div class="card-footer">
          <div class="progress-container">
            <div 
              class="progress-bar" 
              :class="{ 'active': !complete }"
              :style="{ width: downloadProgress + '%' }"
            ></div>
          </div>
        </div>
      </div>

      <div class="speed-card upload-card">
        <div class="card-header">
          <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"/>
          </svg>
          <h2>Upload Speed</h2>
        </div>
        <div class="speed-value" :class="{ 'active': !complete }">
          {{ uploadSpeed || "0.00" }}<span class="unit">Mb/s</span>
        </div>
        <div class="card-footer">
          <div class="progress-container">
            <div 
              class="progress-bar" 
              :class="{ 'active': !complete }"
              :style="{ width: uploadProgress + '%' }"
            ></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Test Info -->
    <div class="test-info">
      <div class="info-item">
        <svg class="info-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <div>
          <div class="info-label">Test Duration</div>
          <div class="info-value">{{ testTime || "0" }} seconds</div>
        </div>
      </div>
      <div class="info-item">
        <svg class="info-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
        </svg>
        <div>
          <div class="info-label">Status</div>
          <div class="info-value status" :class="{ 'active': !complete }">
            {{ complete ? (testTime > 0 ? 'Complete' : 'Ready') : 'Testing...' }}
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="footer">
      <p>Powered by Measurement Lab (M-Lab)</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import ndt7 from "@m-lab/ndt7";

/* state */
const downloadSpeed = ref("0.00");
const uploadSpeed = ref("0.00");
const complete = ref(true);
const testTime = ref(0);
const downloadProgress = ref(0);
const uploadProgress = ref(0);

const startSpeedtest = () => {
  // Reset state
  downloadSpeed.value = "0.00";
  uploadSpeed.value = "0.00";
  testTime.value = 0;
  complete.value = false;
  downloadProgress.value = 0;
  uploadProgress.value = 0;

  const startTime = Date.now();
  let testStage = 'starting';

  ndt7.test(
    {
      userAcceptedDataPolicy: true,
      downloadworkerfile: "/ndt7-download-worker.js",
      uploadworkerfile: "/ndt7-upload-worker.js",
      metadata: {
        client_name: "gomotz",
      },
    },
    {
      downloadMeasurement(data) {
        if (data.Source === "client") {
          const speed = data.Data.MeanClientMbps.toFixed(2);
          downloadSpeed.value = speed;
          downloadProgress.value = Math.min(50, (speed / 100) * 100); // Cap at 50% during download
        }
      },

      downloadComplete(data) {
        const speed = data.LastClientMeasurement.MeanClientMbps.toFixed(2);
        downloadSpeed.value = speed;
        downloadProgress.value = 50; // Complete download phase
        testStage = 'uploading';
      },

      uploadMeasurement(data) {
        if (data.Source === "server") {
          const throughput = (
            (data.Data.TCPInfo.BytesReceived / data.Data.TCPInfo.ElapsedTime) *
            8
          ).toFixed(2);
          uploadSpeed.value = throughput;
          uploadProgress.value = Math.min(50, 50 + (throughput / 100) * 50); // 50-100% range
        }
      },

      uploadComplete(data) {
        const bytes = data.LastServerMeasurement.TCPInfo.BytesReceived;
        const elapsed = data.LastServerMeasurement.TCPInfo.ElapsedTime;
        const throughput = ((bytes * 8) / elapsed).toFixed(2);
        uploadSpeed.value = throughput;
        uploadProgress.value = 100; // Complete upload phase
      },

      error(err) {
        console.error("Speedtest error:", err.message);
        complete.value = true;
        downloadProgress.value = 0;
        uploadProgress.value = 0;
      },
    }
  ).then(() => {
    testTime.value = ((Date.now() - startTime) / 1000).toFixed(1);
    complete.value = true;
  });
};
</script>

<style scoped>
.speedtest-container {
  max-width: 800px;
  margin: 2rem auto;
  padding: 2rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 24px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  color: white;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.header {
  text-align: center;
  margin-bottom: 2rem;
}

.header h1 {
  font-size: 2.5rem;
  font-weight: 700;
  margin: 0;
  background: linear-gradient(to right, #ffffff, #e0e7ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.subtitle {
  opacity: 0.9;
  margin-top: 0.5rem;
  font-size: 1.1rem;
}

.button-container {
  text-align: center;
  margin: 2rem 0;
}

.start-button {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 16px;
  padding: 1.25rem 3rem;
  font-size: 1.25rem;
  font-weight: 600;
  color: white;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  min-width: 250px;
  position: relative;
  overflow: hidden;
}

.start-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 15px 40px rgba(0, 0, 0, 0.3);
}

.start-button:active:not(:disabled) {
  transform: translateY(0);
}

.start-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.start-button.running {
  background: linear-gradient(135deg, #4a5568 0%, #2d3748 100%);
}

.start-button.pulse:not(:disabled)::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: 16px;
  animation: pulse 2s infinite;
  pointer-events: none;
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(102, 126, 234, 0.7);
  }
  70% {
    box-shadow: 0 0 0 20px rgba(102, 126, 234, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(102, 126, 234, 0);
  }
}

.button-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.button-icon {
  width: 24px;
  height: 24px;
}

.spinner {
  width: 24px;
  height: 24px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.last-test-info {
  margin-top: 1rem;
  opacity: 0.8;
  font-size: 0.9rem;
}

.speed-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
  margin: 2rem 0;
}

.speed-card {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 1.5rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  transition: transform 0.3s ease;
}

.speed-card:hover {
  transform: translateY(-4px);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 1.5rem;
}

.card-icon {
  width: 32px;
  height: 32px;
  opacity: 0.9;
}

.card-header h2 {
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0;
}

.speed-value {
  font-size: 3.5rem;
  font-weight: 800;
  line-height: 1;
  margin: 1rem 0;
  font-feature-settings: "tnum";
  font-variant-numeric: tabular-nums;
  transition: color 0.3s ease;
}

.speed-value.active {
  color: #a78bfa;
  animation: glow 1s ease-in-out infinite alternate;
}

@keyframes glow {
  from {
    text-shadow: 0 0 5px rgba(167, 139, 250, 0.5);
  }
  to {
    text-shadow: 0 0 20px rgba(167, 139, 250, 0.8);
  }
}

.unit {
  font-size: 1.5rem;
  font-weight: 600;
  opacity: 0.8;
  margin-left: 4px;
}

.card-footer {
  margin-top: 1.5rem;
}

.progress-container {
  height: 8px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #a78bfa, #8b5cf6);
  border-radius: 4px;
  width: 0%;
  transition: width 0.5s ease;
}

.progress-bar.active {
  animation: shimmer 2s infinite;
}

@keyframes shimmer {
  0% {
    background-position: -200% center;
  }
  100% {
    background-position: 200% center;
  }
}

.test-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin: 2rem 0;
  padding: 1.5rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 16px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.info-icon {
  width: 24px;
  height: 24px;
  opacity: 0.8;
}

.info-label {
  font-size: 0.875rem;
  opacity: 0.8;
  margin-bottom: 0.25rem;
}

.info-value {
  font-size: 1.25rem;
  font-weight: 600;
}

.status {
  color: #10b981;
}

.status.active {
  color: #fbbf24;
}

.footer {
  text-align: center;
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 0.875rem;
  opacity: 0.7;
}

/* Responsive design */
@media (max-width: 768px) {
  .speedtest-container {
    margin: 1rem;
    padding: 1.5rem;
  }
  
  .header h1 {
    font-size: 2rem;
  }
  
  .speed-cards {
    grid-template-columns: 1fr;
  }
  
  .speed-value {
    font-size: 2.75rem;
  }
  
  .start-button {
    width: 100%;
    padding: 1rem 2rem;
  }
}
</style>