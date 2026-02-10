<template>
  <main class="dashboard">
    <!-- Header -->
    <div class="dashboard-header">
      <div class="header-content">
        <h1 class="dashboard-title">Network Dashboard</h1>
        <p class="dashboard-subtitle">Real-time network monitoring and insights</p>
      </div>
      <div class="header-actions">
        <div class="refresh-info" @click="fetchISPInfo">
          <span class="refresh-icon">↻</span>
          <span class="refresh-text">Last updated: {{ lastUpdateTime }}</span>
        </div>
      </div>
    </div>

    <!-- Main Dashboard Grid -->
    <div class="dashboard-grid">
      <!-- Public IP Card -->
      <div class="dashboard-card ip-card">
        <div class="card-header">
          <div class="card-icon">
            <span>🌐</span>
          </div>
          <div class="card-title">
            <h3>Public IP Address</h3>
            <p class="card-subtitle">Your external network identifier</p>
          </div>
        </div>
        <div class="card-content">
          <div class="ip-display">
            <div class="ip-address">{{ Public_IP }}</div>
            <div class="ip-type">IPv4</div>
          </div>
          <div class="copy-action">
            <button class="copy-btn" @click="copyToClipboard(Public_IP)">
              <span class="copy-icon">📋</span>
              Copy
            </button>
          </div>
        </div>
        <div class="card-footer">
          <div class="location-info">
            <span class="location-icon">📍</span>
            <span class="location-text">Detected from your network</span>
          </div>
        </div>
      </div>

      <!-- ISP Info Card -->
      <div class="dashboard-card isp-card">
        <div class="card-header">
          <div class="card-icon">
            <span>🏢</span>
          </div>
          <div class="card-title">
            <h3>Internet Service Provider</h3>
            <p class="card-subtitle">Your network provider details</p>
          </div>
        </div>
        <div class="card-content">
          <div class="isp-details">
            <div class="isp-name">{{ ISP_Info }}</div>
            <div class="isp-asn">AS{{ ASN || 'Unknown' }}</div>
          </div>
          <!-- <div class="isp-meta"> -->
            <!-- <div class="meta-item">
              <span class="meta-label">Country:</span>
              <span class="meta-value">{{ country || 'Unknown' }}</span>
            </div> -->
            <!-- <div class="meta-item">
              <span class="meta-label">Region:</span>
              <span class="meta-value">{{ region || 'Unknown' }}</span>
            </div> -->
          <!-- </div> -->
        </div>
      </div>

      <!-- Connection Status Card -->
      <div class="dashboard-card status-card" :class="connectionClass">
        <div class="card-header">
          <div class="card-icon">
            <span :class="statusIcon"></span>
          </div>
          <div class="card-title">
            <h3>Connection Status</h3>
            <p class="card-subtitle">Real-time network availability</p>
          </div>
        </div>
        <div class="card-content">
          <div class="status-display">
            <div class="status-indicator">
              <div class="status-pulse"></div>
              <div class="status-text">{{ Internet_Status }}</div>
            </div>
            <div class="latency-display" v-if="latency">
              <span class="latency-value">{{ latency }}ms</span>
              <span class="latency-label">Latency</span>
            </div>
          </div>
          <div class="uptime-stats">
            <div class="uptime-bar">
              <div class="uptime-fill" :style="{ width: uptimePercentage + '%' }"></div>
            </div>
            <div class="uptime-text">{{ uptimePercentage }}% Uptime</div>
          </div>
        </div>
        <div class="card-footer">
          <div class="status-history">
            <span class="history-icon">📈</span>
            <span class="history-text">Last check: {{ formatTime(lastCheck) }}</span>
          </div>
        </div>
      </div>

      <!-- Network Statistics -->
      <div class="dashboard-card stats-card">
        <div class="card-header">
          <div class="card-icon">
            <span>📊</span>
          </div>
          <div class="card-title">
            <h3>Network Statistics</h3>
            <p class="card-subtitle">Performance overview</p>
          </div>
        </div>
        <div class="card-content">
          <div class="stats-grid">
            <div class="stat-item">
              <div class="stat-value">{{ checksCount }}</div>
              <div class="stat-label">Total Checks</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">{{ successRate }}%</div>
              <div class="stat-label">Success Rate</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">{{ avgLatency }}ms</div>
              <div class="stat-label">Avg Latency</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">{{ responseTime }}ms</div>
              <div class="stat-label">Response Time</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="dashboard-card actions-card">
        <div class="card-header">
          <div class="card-icon">
            <span>⚡</span>
          </div>
          <div class="card-title">
            <h3>Quick Actions</h3>
            <p class="card-subtitle">Common network operations</p>
          </div>
        </div>
        <div class="card-content">
          <div class="actions-grid">
            <button class="action-btn" @click="runSpeedTest">
              <span class="action-icon">📶</span>
              <span class="action-text">Speed Test</span>
            </button>
            <button class="action-btn" @click="pingNetwork">
              <span class="action-icon">🏓</span>
              <span class="action-text">Ping Test</span>
            </button>
            <button class="action-btn" @click="runTraceroute">
              <span class="action-icon">🛣️</span>
              <span class="action-text">Traceroute</span>
            </button>
            <!-- <button class="action-btn" @click="flushDNS">
              <span class="action-icon">🧹</span>
              <span class="action-text">Flush DNS</span>
            </button> -->
          </div>
        </div>
      </div>

      <!-- Connection Timeline -->
      <div class="dashboard-card timeline-card">
        <div class="card-header">
          <div class="card-icon">
            <span>⏰</span>
          </div>
          <div class="card-title">
            <h3>Connection History</h3>
            <p class="card-subtitle">Recent network events</p>
          </div>
        </div>
        <div class="card-content">
          <div class="timeline">
            <div v-for="(event, index) in connectionHistory" :key="index" class="timeline-item">
              <div class="timeline-marker" :class="event.status"></div>
              <div class="timeline-content">
                <div class="timeline-time">{{ event.time }}</div>
                <div class="timeline-event">{{ event.message }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <div class="loading-spinner"></div>
      <p>Updating network information...</p>
    </div>

    <!-- Notification Toast -->
    <div v-if="showNotification" class="notification-toast">
      <div class="toast-content">
        <span class="toast-icon">✅</span>
        <span class="toast-message">{{ notificationMessage }}</span>
      </div>
    </div>
  </main>
</template>

<script>
import { ref, computed, onMounted, onUnmounted } from 'vue';
// const API_URL = import.meta.env.VITE_API_URL;
const API_URL = "http://192.168.20.17:8082";

export default {
  data() {
    return {
      Public_IP: "Loading...",
      ISP_Info: "Loading...",
      Internet_Status: "Checking...",
      ASN: "",
      country: "",
      region: "",
      city: "",
      latency: 0,
      loading: false,
      intervalId: null,
      checksCount: 0,
      successfulChecks: 0,
      lastCheck: new Date(),
      connectionHistory: [],
      showNotification: false,
      notificationMessage: "",
      responseTime: 0,
      startTime: Date.now(),
      uptimeData: []
    };
  },

  computed: {
    lastUpdateTime() {
      return this.formatTime(this.lastCheck);
    },
    
    connectionClass() {
      return this.Internet_Status === 'Online' ? 'online' : 'offline';
    },
    
    statusIcon() {
      return this.Internet_Status === 'Online' ? '✅' : '❌';
    },
    
    successRate() {
      return this.checksCount > 0 
        ? Math.round((this.successfulChecks / this.checksCount) * 100)
        : 0;
    },
    
    avgLatency() {
      if (this.uptimeData.length === 0) return 0;
      const sum = this.uptimeData.reduce((a, b) => a + b.latency, 0);
      return Math.round(sum / this.uptimeData.length);
    },
    
    uptimePercentage() {
      const onlineCount = this.uptimeData.filter(d => d.status === 'online').length;
      return this.uptimeData.length > 0 
        ? Math.round((onlineCount / this.uptimeData.length) * 100)
        : 100;
    }
  },

  methods: {
    async fetchISPInfo() {
      this.loading = true;
      this.checksCount++;
      
      const startTime = Date.now();
      
      try {
        const response = await fetch(`${API_URL}/v1/services/isp`);
        
        if (!response.ok) {
          throw new Error("Failed to fetch ISP info");
        }

        const data = await response.json();
        const endTime = Date.now();
        this.responseTime = endTime - startTime;

        this.Public_IP = data.ip || "Unknown";
        this.ISP_Info = data.org || "Unknown ISP";
        this.ASN = data.asn || "Unknown";
        // this.country = data.country || "Unknown";
        // this.region = data.region || "Unknown";
        this.city = data.city || "Unknown";
        this.Internet_Status = "Online";
        this.latency = Math.round(Math.random() * 50 + 10); // Simulated latency
        this.successfulChecks++;
        
        this.uptimeData.push({
          timestamp: new Date(),
          status: 'online',
          latency: this.latency
        });
        
        // Keep only last 100 records
        if (this.uptimeData.length > 100) {
          this.uptimeData.shift();
        }

        this.addHistoryEvent('online', `Connected successfully (${this.latency}ms)`);
        
      } catch (error) {
        console.error("Error fetching ISP info:", error);
        this.Internet_Status = "Offline";
        this.latency = 0;
        
        this.uptimeData.push({
          timestamp: new Date(),
          status: 'offline',
          latency: 0
        });
        
        this.addHistoryEvent('offline', 'Connection failed');
      } finally {
        this.lastCheck = new Date();
        this.loading = false;
      }
    },

    addHistoryEvent(status, message) {
      const now = new Date();
      const timeString = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
      
      this.connectionHistory.unshift({
        time: timeString,
        status: status,
        message: message
      });
      
      // Keep only last 10 events
      if (this.connectionHistory.length > 10) {
        this.connectionHistory.pop();
      }
    },

    formatTime(date) {
      if (!date) return 'Never';
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    },

    async copyToClipboard(text) {
      try {
        await navigator.clipboard.writeText(text);
        this.showNotificationMessage('IP address copied to clipboard!');
      } catch (err) {
        console.error('Failed to copy:', err);
        this.showNotificationMessage('Failed to copy IP address');
      }
    },

    showNotificationMessage(message) {
      this.notificationMessage = message;
      this.showNotification = true;
      
      setTimeout(() => {
        this.showNotification = false;
      }, 3000);
    },

    runSpeedTest() {
      this.showNotificationMessage('Starting speed test...');
       this.$router.push('/tools/speedtest');
    },

    pingNetwork() {
      this.showNotificationMessage('Pinging network...');
      this.$router.push('/tools/ping');
    },

    runTraceroute() {
      this.showNotificationMessage('Running traceroute...');
    this.$router.push('/tools/traceroute');
    },

  },

  mounted() {
    this.fetchISPInfo();
    const FIVE_HOURS = 5 * 60 * 60 * 1000;
    this.intervalId = setInterval(() => {
      this.fetchISPInfo();
    }, FIVE_HOURS); // Check every 30 seconds

    // Add initial history event
    this.addHistoryEvent('info', 'Dashboard initialized');
  },

  beforeUnmount() {
    if (this.intervalId) {
      clearInterval(this.intervalId);
    }
  }
};
</script>

<style scoped>
.dashboard {
  min-height: 100vh;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  padding: 30px;
  color: #e2e8f0;
}

/* Header */
.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 40px;
  padding-bottom: 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.header-content .dashboard-title {
  font-size: 2.5rem;
  font-weight: 700;
  background: linear-gradient(135deg, #60a5fa 0%, #818cf8 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin: 0;
}

.dashboard-subtitle {
  color: #94a3b8;
  margin: 8px 0 0;
  font-size: 1.1rem;
}

.refresh-info {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 20px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.refresh-info:hover {
  background: rgba(255, 255, 255, 0.1);
  transform: translateY(-2px);
}

.refresh-icon {
  font-size: 1.2rem;
  animation: spin 2s linear infinite;
}

.refresh-text {
  font-size: 0.9rem;
  color: #94a3b8;
}

/* Dashboard Grid */
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 25px;
  margin-top: 30px;
}

/* Card Styles */
.dashboard-card {
  background: rgba(30, 41, 59, 0.7);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 25px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.dashboard-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, #3b82f6, #8b5cf6);
  opacity: 0.7;
}

.dashboard-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
  border-color: rgba(255, 255, 255, 0.2);
}

/* Card Header */
.card-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 25px;
}

.card-icon {
  font-size: 2rem;
  background: rgba(255, 255, 255, 0.1);
  width: 60px;
  height: 60px;
  border-radius: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-title h3 {
  margin: 0;
  font-size: 1.3rem;
  font-weight: 600;
}

.card-subtitle {
  margin: 4px 0 0;
  color: #94a3b8;
  font-size: 0.9rem;
}

/* IP Card Specific */
.ip-card .ip-display {
  background: rgba(255, 255, 255, 0.05);
  padding: 20px;
  border-radius: 12px;
  margin-bottom: 20px;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.ip-address {
  font-family: 'Courier New', monospace;
  font-size: 1.8rem;
  font-weight: 600;
  color: #60a5fa;
  margin-bottom: 8px;
}

.ip-type {
  display: inline-block;
  padding: 4px 12px;
  background: rgba(96, 165, 250, 0.2);
  color: #60a5fa;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
}

.copy-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: white;
  border: none;
  border-radius: 10px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.copy-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(59, 130, 246, 0.4);
}

/* Status Card */
.status-card {
  border: 2px solid;
}

.status-card.online {
  border-color: rgba(34, 197, 94, 0.3);
  background: rgba(34, 197, 94, 0.05);
}

.status-card.offline {
  border-color: rgba(239, 68, 68, 0.3);
  background: rgba(239, 68, 68, 0.05);
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.status-pulse {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.status-card.online .status-pulse {
  background: #10b981;
  box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
}

.status-card.offline .status-pulse {
  background: #ef4444;
  box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7);
}

.status-text {
  font-size: 1.5rem;
  font-weight: 600;
}

.status-card.online .status-text {
  color: #10b981;
}

.status-card.offline .status-text {
  color: #ef4444;
}

.latency-display {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.latency-value {
  font-size: 1.8rem;
  font-weight: 700;
  color: #fbbf24;
}

.latency-label {
  color: #94a3b8;
  font-size: 0.9rem;
}

.uptime-bar {
  height: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
  overflow: hidden;
  margin: 15px 0;
}

.uptime-fill {
  height: 100%;
  background: linear-gradient(90deg, #10b981, #34d399);
  border-radius: 3px;
  transition: width 1s ease;
}

.uptime-text {
  text-align: center;
  color: #94a3b8;
  font-size: 0.9rem;
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.stat-item {
  text-align: center;
  padding: 20px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  transition: all 0.3s;
}

.stat-item:hover {
  background: rgba(255, 255, 255, 0.1);
  transform: translateY(-3px);
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  color: #60a5fa;
  margin-bottom: 4px;
}

.stat-label {
  color: #94a3b8;
  font-size: 0.9rem;
}

/* Actions Grid */
.actions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
}

.action-btn {
  padding: 20px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  color: #e2e8f0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  transition: all 0.3s;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  transform: translateY(-3px);
  border-color: rgba(59, 130, 246, 0.3);
}

.action-icon {
  font-size: 1.8rem;
}

.action-text {
  font-size: 0.9rem;
  font-weight: 600;
}

/* Timeline */
.timeline {
  position: relative;
  padding-left: 30px;
}

.timeline::before {
  content: '';
  position: absolute;
  left: 15px;
  top: 0;
  bottom: 0;
  width: 2px;
  background: rgba(255, 255, 255, 0.1);
}

.timeline-item {
  display: flex;
  align-items: flex-start;
  margin-bottom: 20px;
  position: relative;
}

.timeline-marker {
  position: absolute;
  left: -30px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.timeline-marker.online {
  background: #10b981;
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.5);
}

.timeline-marker.offline {
  background: #ef4444;
  box-shadow: 0 0 10px rgba(239, 68, 68, 0.5);
}

.timeline-marker.info {
  background: #60a5fa;
  box-shadow: 0 0 10px rgba(96, 165, 250, 0.5);
}

.timeline-time {
  font-size: 0.8rem;
  color: #94a3b8;
  margin-bottom: 4px;
}

.timeline-event {
  font-size: 0.9rem;
  color: #e2e8f0;
}

/* Card Footer */
.card-footer {
  margin-top: 20px;
  padding-top: 15px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  color: #94a3b8;
  font-size: 0.85rem;
}

.location-info, .status-history {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Loading Overlay */
.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(5px);
}

.loading-spinner {
  width: 60px;
  height: 60px;
  border: 4px solid rgba(255, 255, 255, 0.1);
  border-top-color: #60a5fa;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 20px;
}

.loading-overlay p {
  color: #94a3b8;
  font-size: 1.1rem;
}

/* Notification Toast */
.notification-toast {
  position: fixed;
  bottom: 30px;
  right: 30px;
  background: rgba(30, 41, 59, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 15px 20px;
  display: flex;
  align-items: center;
  gap: 15px;
  z-index: 1000;
  animation: slideIn 0.3s ease;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.toast-icon {
  font-size: 1.2rem;
  color: #10b981;
}

.toast-message {
  color: #e2e8f0;
  font-weight: 500;
}

/* Animations */
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(16, 185, 129, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0);
  }
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Responsive Design */
@media (max-width: 1200px) {
  .dashboard-grid {
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  }
}

@media (max-width: 768px) {
  .dashboard {
    padding: 20px;
  }
  
  .dashboard-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 20px;
  }
  
  .dashboard-grid {
    grid-template-columns: 1fr;
  }
  
  .dashboard-title {
    font-size: 2rem;
  }
  
  .stats-grid,
  .actions-grid {
    grid-template-columns: 1fr;
  }
}

/* Custom scrollbar */
.dashboard ::-webkit-scrollbar {
  width: 8px;
}

.dashboard ::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
}

.dashboard ::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}

.dashboard ::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>