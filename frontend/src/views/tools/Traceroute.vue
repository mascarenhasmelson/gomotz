<template>
  <div class="traceroute-page">
    <div class="traceroute-container">

      <!-- Scan Form -->
      <div class="scan-form">
        <div class="form-group">
          <label for="target-input">Target Host:</label>
          <div class="input-with-button">
            <input
              v-model="targetHost"
              id="target-input"
              type="text"
              placeholder="Enter IP address or hostname (e.g., google.com, 8.8.8.8)"
              :disabled="isTracing"
              @keyup.enter="startTrace"
              class="target-input"
            />
            <button 
              @click="startTrace" 
              :disabled="!targetHost || isTracing" 
              class="scan-button"
            >
              <span v-if="!isTracing">Start Trace</span>
              <span v-else class="scanning-text">
                <span class="spinner"></span> Tracing...
              </span>
            </button>
            <button 
              v-if="isTracing"
              @click="stopTrace" 
              class="stop-button"
            >
              Stop
            </button>
          </div>
        </div>
        
        <!-- Connection Status -->
        <div v-if="connectionStatus" class="connection-status" :class="connectionStatusClass">
          {{ connectionStatus }}
        </div>
        
        <!-- Error Message -->
        <div v-if="error" class="error-message">
          ⚠️ {{ error }}
        </div>
      </div>

      <!-- Results Table -->
      <div v-if="traceResults.length > 0" class="results-section">
        <div class="results-header">
          <h2>Traceroute Results</h2>
          <div class="results-meta">
            <span class="target-display">{{ targetHost }}</span>
            <span class="scan-time">{{ formattedTraceTime }}</span>
          </div>
        </div>

        <div class="results-table">
          <table>
            <thead>
              <tr>
                <th>Hop</th>
                <th>IP Address</th>
                <th>Hostname</th>
                <th>Latency (ms)</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <tr 
                v-for="(hop, index) in traceResults" 
                :key="index"
              >
                <td class="hop-cell">
                  <div class="hop-number">{{ hop.hop }}</div>
                </td>
                <td class="ip-cell">
                  <div class="ip-address">{{ hop.ip }}</div>
                </td>
                <td class="hostname-cell">
                  <div class="hostname">{{ hop.hostname || '-' }}</div>
                </td>
                <td class="latency-cell">
                  <div class="latency-stats">
                    <span v-for="(rtt, i) in hop.rtts" :key="i" class="latency-badge">
                      {{ rtt }}ms
                    </span>
                  </div>
                </td>
                <td class="status-cell">
                  <span class="status-badge" :class="getStatusClass(hop)">
                    {{ getStatusText(hop) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Summary Statistics -->
        <div class="summary-section">
          <h3>Summary</h3>
          <div class="summary-stats">
            <div class="stat-card">
              <div class="stat-content">
                <div class="stat-value">{{ traceResults.length }}</div>
                <div class="stat-label">Total Hops</div>
              </div>
            </div>
            
            <div class="stat-card">
              <div class="stat-content">
                <div class="stat-value">{{ avgLatency }}ms</div>
                <div class="stat-label">Avg Latency</div>
              </div>
            </div>
            
            <div class="stat-card">
              <div class="stat-content">
                <div class="stat-value">{{ traceCompleted ? 'Reached' : 'Failed' }}</div>
                <div class="stat-label">Destination</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Traceroute',
  data() {
    return {
      targetHost: 'google.com',
      targetIp: '',
      isTracing: false,
      traceResults: [],
      error: null,
      connectionStatus: '',
      connectionStatusClass: '',
      maxHops: 30,
      probesPerHop: 3,
      traceStartTime: null,
      elapsedTime: 0,
      traceInterval: null,
      traceProgress: 0,
      currentHop: 1,
      traceCompleted: false,
      totalDuration: 0,
      ws: null,
      traceStartTimestamp: 0
    }
  },
  computed: {
    formattedTraceTime() {
      if (!this.traceStartTime) return '';
      return new Date(this.traceStartTime).toLocaleString();
    },
    avgLatency() {
      if (this.traceResults.length === 0) return 0;
      const validLatencies = this.traceResults
        .filter(h => h.avgRtt)
        .map(h => h.avgRtt);
      if (validLatencies.length === 0) return 0;
      const sum = validLatencies.reduce((a, b) => a + b, 0);
      return Math.round(sum / validLatencies.length);
    }
  },
  beforeUnmount() {
    this.cleanup();
  },
  methods: {
    getStatusClass(hop) {
      if (hop.status === 'timeout') return 'timeout';
      if (hop.status === 'error') return 'error';
      if (hop.rtts && hop.rtts.length === this.probesPerHop) return 'complete';
      if (hop.rtts && hop.rtts.length > 0) return 'partial';
      return 'pending';
    },
    
    getStatusText(hop) {
      if (hop.status === 'timeout') return 'Timeout';
      if (hop.status === 'error') return 'Error';
      if (hop.rtts && hop.rtts.length === this.probesPerHop) return 'Complete';
      if (hop.rtts && hop.rtts.length > 0) return 'Partial';
      return 'Pending';
    },
    
    cleanup() {
      if (this.traceInterval) {
        clearInterval(this.traceInterval);
        this.traceInterval = null;
      }
      
      if (this.ws) {
        this.ws.close();
        this.ws = null;
      }
    },
    
    async startTrace() {
      if (!this.targetHost.trim()) {
        this.error = 'Please enter a target hostname or IP address';
        return;
      }

      // Reset state
      this.isTracing = true;
      this.error = null;
      this.traceResults = [];
      this.traceCompleted = false;
      this.traceStartTime = new Date();
      this.traceStartTimestamp = Date.now();
      this.elapsedTime = 0;
      this.traceProgress = 0;
      this.currentHop = 1;
      this.targetIp = '';
      
      this.connectionStatus = 'Starting traceroute...';
      this.connectionStatusClass = 'connecting';
      
      // Clean up any existing connections
      this.cleanup();
      
      // Connect via WebSocket
      this.connectWebSocket();
      
      // Start elapsed time counter
      this.traceInterval = setInterval(() => {
        this.elapsedTime = Math.floor((Date.now() - this.traceStartTimestamp) / 1000);
        this.updateProgress();
      }, 1000);
    },
    
    connectWebSocket() {
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${wsProtocol}//${window.location.hostname}:8082/v1/traceroute`;
      
      this.ws = new WebSocket(wsUrl);
      
      this.ws.onopen = () => {
        this.connectionStatus = 'Connected, starting trace...';
        this.connectionStatusClass = 'connected';
        
        // Send trace request
        this.ws.send(JSON.stringify({
          action: 'startTrace',
          target: this.targetHost.trim(),
          maxHops: this.maxHops,
          probesPerHop: this.probesPerHop
        }));
      };
      
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          
          switch (data.type) {
            case 'hopResult':
              this.handleHopResult(data.hop);
              break;
              
            case 'targetIp':
              this.targetIp = data.ip;
              break;
              
            case 'progress':
              this.currentHop = data.currentHop;
              break;
              
            case 'complete':
              this.handleTraceComplete(data);
              break;
              
            case 'error':
              this.error = data.message;
              this.connectionStatus = 'Error: ' + data.message;
              this.connectionStatusClass = 'error';
              this.isTracing = false;
              break;
              
            case 'status':
              this.connectionStatus = data.message;
              break;
          }
        } catch (err) {
          console.error('Error parsing WebSocket message:', err);
        }
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket Error:', error);
        this.error = 'Failed to connect to traceroute server';
        this.connectionStatus = 'Connection failed';
        this.connectionStatusClass = 'error';
        this.isTracing = false;
      };
      
      this.ws.onclose = () => {
        this.connectionStatus = 'Connection closed';
        this.connectionStatusClass = 'disconnected';
        this.ws = null;
      };
    },
    
    handleHopResult(hopData) {
      // Check if hop already exists
      const existingIndex = this.traceResults.findIndex(h => h.hop === hopData.hop);
      
      if (existingIndex !== -1) {
        // Update existing hop
        const existing = this.traceResults[existingIndex];
        
        // Add new RTT if not already present
        if (hopData.rtt && !existing.rtts.includes(hopData.rtt)) {
          existing.rtts.push(hopData.rtt);
          existing.avgRtt = Math.round(existing.rtts.reduce((a, b) => a + b, 0) / existing.rtts.length);
        }
        
        // Update other fields
        if (hopData.ip && !existing.ip) existing.ip = hopData.ip;
        if (hopData.hostname && !existing.hostname) existing.hostname = hopData.hostname;
        if (hopData.status) existing.status = hopData.status;
        
      } else {
        // Add new hop
        const hop = {
          hop: hopData.hop,
          ip: hopData.ip || '',
          hostname: hopData.hostname || '',
          rtts: hopData.rtt ? [hopData.rtt] : [],
          avgRtt: hopData.rtt || 0,
          status: hopData.status || (hopData.rtt ? 'success' : 'timeout')
        };
        
        this.traceResults.push(hop);
        this.traceResults.sort((a, b) => a.hop - b.hop);
      }
      
      // Check if this is the target
      if (hopData.isTarget) {
        this.traceCompleted = true;
      }
    },
    
    updateProgress() {
      if (this.maxHops > 0) {
        this.traceProgress = (this.currentHop / this.maxHops) * 100;
      }
    },
    
    handleTraceComplete(data) {
      this.totalDuration = ((Date.now() - this.traceStartTimestamp) / 1000).toFixed(2);
      this.traceProgress = 100;
      this.connectionStatus = 'Trace completed';
      this.connectionStatusClass = 'completed';
      
      // Add any final hops from complete data
      if (data.hops && Array.isArray(data.hops)) {
        data.hops.forEach(hop => {
          this.handleHopResult(hop);
        });
      }
      
      if (data.targetIp) {
        this.targetIp = data.targetIp;
      }
      
      this.isTracing = false;
      this.traceCompleted = data.reached || false;
      
      // Cleanup
      this.cleanup();
    },
    
    stopTrace() {
      if (confirm('Stop the current traceroute?')) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
          this.ws.send(JSON.stringify({ action: 'stopTrace' }));
        }
        
        this.isTracing = false;
        this.connectionStatus = 'Trace stopped by user';
        this.connectionStatusClass = 'stopped';
        
        this.cleanup();
      }
    }
  }
}
</script>

<style scoped>
/* Dark Mode Theme */
.traceroute-page {
  padding: 20px;
  width: 100%;
  min-height: 100vh;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  color: #e2e8f0;
}

.traceroute-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

/* Scan Form */
.scan-form {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  margin-bottom: 30px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.form-group label {
  display: block;
  font-weight: 600;
  color: #cbd5e1;
  margin-bottom: 12px;
  font-size: 1.1rem;
}

.input-with-button {
  display: flex;
  gap: 15px;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.target-input {
  flex: 1;
  padding: 16px 20px;
  border: 1px solid #334155;
  border-radius: 10px;
  font-size: 1rem;
  font-family: 'Monaco', 'Courier New', monospace;
  transition: all 0.3s;
  background: #0f172a;
  color: #e2e8f0;
}

.target-input:focus {
  outline: none;
  border-color: #4299e1;
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.2);
  background: #1a1e2c;
}

.target-input::placeholder {
  color: #64748b;
}

.target-input:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  background: #1e293b;
}

.scan-button {
  background: linear-gradient(135deg, #4299e1 0%, #3182ce 100%);
  color: white;
  border: none;
  padding: 16px 40px;
  border-radius: 10px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  min-width: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
}

.scan-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(66, 153, 225, 0.3);
}

.scan-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.stop-button {
  background: linear-gradient(135deg, #fc8181 0%, #e53e3e 100%);
  color: white;
  border: none;
  padding: 16px 25px;
  border-radius: 10px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  min-width: 100px;
  white-space: nowrap;
}

.stop-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(229, 62, 62, 0.3);
}

.scanning-text {
  display: flex;
  align-items: center;
  gap: 10px;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Connection Status */
.connection-status {
  margin-top: 15px;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  border: 1px solid transparent;
}

.connection-status.connecting {
  background: rgba(237, 137, 54, 0.1);
  color: #fbbf24;
  border-color: rgba(237, 137, 54, 0.3);
}

.connection-status.connected {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border-color: rgba(16, 185, 129, 0.3);
}

.connection-status.error {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border-color: rgba(239, 68, 68, 0.3);
}

.connection-status.disconnected {
  background: rgba(107, 114, 128, 0.1);
  color: #9ca3af;
  border-color: rgba(107, 114, 128, 0.3);
}

.connection-status.completed {
  background: rgba(59, 130, 246, 0.1);
  color: #60a5fa;
  border-color: rgba(59, 130, 246, 0.3);
}

.connection-status.stopped {
  background: rgba(139, 92, 246, 0.1);
  color: #a78bfa;
  border-color: rgba(139, 92, 246, 0.3);
}

/* Error Message */
.error-message {
  margin-top: 15px;
  padding: 12px 16px;
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border-radius: 8px;
  border-left: 4px solid #ef4444;
  font-size: 0.9rem;
}

/* Results Section */
.results-section {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  margin-bottom: 30px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 25px;
  padding-bottom: 20px;
  border-bottom: 1px solid #334155;
}

.results-header h2 {
  margin: 0;
  color: #f8fafc;
  font-size: 1.8rem;
}

.results-meta {
  display: flex;
  align-items: center;
  gap: 20px;
}

.target-display {
  font-family: 'Monaco', 'Courier New', monospace;
  font-weight: 600;
  color: #60a5fa;
  font-size: 1.1rem;
}

.scan-time {
  color: #94a3b8;
  font-size: 0.9rem;
}

.results-table {
  overflow-x: auto;
  margin: 20px 0;
  border-radius: 10px;
  border: 1px solid #334155;
}

table {
  width: 100%;
  border-collapse: collapse;
  min-width: 800px;
}

thead {
  background: #0f172a;
  border-bottom: 1px solid #334155;
}

th {
  padding: 16px 20px;
  text-align: left;
  font-weight: 600;
  color: #94a3b8;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

tbody tr {
  border-bottom: 1px solid #334155;
  transition: all 0.3s;
}

tbody tr:hover {
  background: rgba(59, 130, 246, 0.1);
}

td {
  padding: 16px 20px;
  vertical-align: middle;
  color: #cbd5e1;
}

/* Table Cell Styles */
.hop-number {
  font-weight: 700;
  color: #60a5fa;
  font-size: 1.1rem;
}

.ip-cell, .hostname-cell {
  font-family: 'Monaco', 'Courier New', monospace;
}

.ip-address {
  color: #60a5fa;
}

.hostname {
  color: #94a3b8;
}

.latency-stats {
  display: flex;
  gap: 5px;
  flex-wrap: wrap;
}

.latency-badge {
  padding: 4px 8px;
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 600;
  font-family: 'Monaco', 'Courier New', monospace;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.status-badge {
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  display: inline-block;
}

.status-badge.complete {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.status-badge.partial {
  background: rgba(245, 158, 11, 0.1);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.status-badge.pending {
  background: rgba(107, 114, 128, 0.1);
  color: #9ca3af;
  border: 1px solid rgba(107, 114, 128, 0.3);
}

.status-badge.timeout {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.status-badge.error {
  background: rgba(139, 92, 246, 0.1);
  color: #a78bfa;
  border: 1px solid rgba(139, 92, 246, 0.3);
}

/* Summary Section */
.summary-section {
  margin-top: 40px;
  padding: 30px;
  background: #0f172a;
  border-radius: 12px;
  border: 1px solid #334155;
}

.summary-section h3 {
  margin: 0 0 25px 0;
  color: #f8fafc;
  font-size: 1.5rem;
}

.summary-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: #1e293b;
  border-radius: 10px;
  border: 1px solid #334155;
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-5px);
  border-color: #4299e1;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
}

.stat-content {
  text-align: center;
}

.stat-value {
  font-size: 1.8rem;
  font-weight: 700;
  color: #f8fafc;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 0.9rem;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Scrollbar Styling */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #0f172a;
}

::-webkit-scrollbar-thumb {
  background: #334155;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #475569;
}

/* Responsive Design */
@media (max-width: 768px) {
  .traceroute-container {
    padding: 10px;
  }
  
  .input-with-button {
    flex-direction: column;
  }
  
  .target-input,
  .scan-button,
  .stop-button {
    width: 100%;
  }
  
  .results-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }
  
  .results-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .summary-stats {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .summary-stats {
    grid-template-columns: 1fr;
  }
  
  .scan-header h1 {
    font-size: 2rem;
  }
}
</style>