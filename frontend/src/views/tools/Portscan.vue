<template>
  <div class="portscan-page">
    <div class="portscan-container">
      <!-- Scan Form -->
      <div class="scan-form">
        <div class="form-group">
          <label for="target-input">Target IP / Host:</label>
          <div class="input-with-button">
            <input
              v-model="scanTarget"
              id="target-input"
              type="text"
              placeholder="Enter IP address or hostname (e.g., 1.1.1.1, google.com)"
              :disabled="isScanning"
              @keyup.enter="startScan"
              class="target-input"
            />
            <button 
              @click="startScan" 
              :disabled="!scanTarget || isScanning" 
              class="scan-button"
            >
              <span v-if="!isScanning">Scan</span>
              <span v-else class="scanning-text">
                <span class="spinner"></span> Scanning...
              </span>
            </button>
            <button 
              v-if="isScanning"
              @click="stopScan" 
              class="stop-button"
            >
              Stop Scan
            </button>
          </div>
          
          <!-- Loading Animation -->
          <div v-if="isScanning" class="loading-section">
            <div class="progress-container">
              <div class="progress-bar" :style="{ width: scanProgress + '%' }"></div>
            </div>
            
            <div class="loading-info">
              <div class="loading-stats">
                <div class="loading-stat">
                  <span class="stat-icon">⏱️</span>
                  <span class="stat-text">Time: {{ elapsedTime }}s</span>
                </div>
                <div class="loading-stat">
                  <span class="stat-icon">📊</span>
                  <span class="stat-text">Progress: {{ Math.round(scanProgress) }}%</span>
                </div>
                <div class="loading-stat">
                  <span class="stat-icon">🔍</span>
                  <span class="stat-text">Status: {{ scanStatus }}</span>
                </div>
                <div class="loading-stat">
                  <span class="stat-icon">🚪</span>
                  <span class="stat-text">Ports Found: {{ openPorts.length }}</span>
                </div>
                <div class="loading-stat">
                  <span class="stat-icon">⚡</span>
                  <span class="stat-text">Rate: {{ scanRate }} ports/sec</span>
                </div>
              </div>
              
              <div class="port-visualization">
                <div class="vis-title">Scanning Progress (1-65535)</div>
                <div class="port-range-info">
                  <div class="range-start">Port 1</div>
                  <div class="range-progress">
                    <div class="range-bar">
                      <div class="range-fill" :style="{ width: scanProgress + '%' }"></div>
                    </div>
                  </div>
                  <div class="range-end">Port 65535</div>
                </div>
                <div class="current-port">
                  <div class="current-range">
                    Currently scanning: <strong>{{ currentPortRange }}</strong>
                  </div>
                  <div class="ports-found">
                    Open ports found: 
                    <span class="open-ports-list">
                      <span 
                        v-for="(port, index) in openPorts" 
                        :key="port.port"
                        class="open-port-badge"
                        :title="`Port ${port.port} - ${port.service}`"
                      >
                        {{ port.port }}
                        <span v-if="index < openPorts.length - 1">, </span>
                      </span>
                      <span v-if="openPorts.length === 0">None yet</span>
                    </span>
                  </div>
                </div>
                
                <!-- Real-time port visualization -->
                <div class="port-grid">
                  <div 
                    v-for="n in 100" 
                    :key="n"
                    class="port-block"
                    :class="getPortBlockStatus(n)"
                    :title="`Ports ${(n-1)*656 + 1} - ${n*656}`"
                    @mouseover="hoverBlock = n"
                    @mouseleave="hoverBlock = null"
                  >
                    <div v-if="hoverBlock === n" class="block-tooltip">
                      Ports {{ (n-1)*656 + 1 }} - {{ n*656 }}
                    </div>
                  </div>
                </div>
                
                <div class="vis-legend">
                  <div class="legend-item"><span class="dot pending"></span>Pending</div>
                  <div class="legend-item"><span class="dot scanning"></span>Scanning</div>
                  <div class="legend-item"><span class="dot open"></span>Open Found</div>
                  <div class="legend-item"><span class="dot completed"></span>Completed</div>
                </div>
              </div>
              
              <!-- Real-time open ports display -->
              <div v-if="openPorts.length > 0" class="realtime-ports">
                <h3>Open Ports Found:</h3>
                <div class="ports-grid">
                  <div 
                    v-for="port in openPorts" 
                    :key="port.port"
                    class="port-bubble"
                    :class="getPortBubbleClass(port.port)"
                  >
                    <div class="port-number">{{ port.port }}</div>
                    <div class="port-service">{{ port.service }}</div>
                    <div class="port-status">OPEN</div>
                  </div>
                </div>
              </div>
              
              <!-- <div class="scanning-tips">
                <div class="tip-icon">💡</div>
                <div class="tip-content">
                  <strong>SYN Scan in Progress</strong>
                  <p>Sending SYN packets to all 65535 ports and listening for SYN-ACK responses.</p>
                  <ul>
                    <li>Ports scanned: {{ scannedPortsCount }}/65535</li>
                    <li>Open ports found: {{ openPorts.length }}</li>
                    <li>Estimated time remaining: {{ estimatedTimeRemaining }}s</li>
                    <li>Packet rate: ~6,500 packets/second</li>
                  </ul>
                </div>
              </div> -->
            </div>
          </div>
          
          <p v-if="error" class="error-message">{{ error }}</p>
          <p v-if="connectionStatus" class="connection-status" :class="connectionStatusClass">
            {{ connectionStatus }}
          </p>
        </div>
      </div>

      <!-- Results Section -->
      <div v-if="showResults && !isScanning" class="results-section">
        <div class="results-header">
          <h2>Scan Results</h2>
          <div class="results-meta">
            <span class="target-display">{{ scanTarget }}</span>
            <span class="scan-time">{{ formattedScanTime }}</span>
          </div>
        </div>

        <!-- Port List -->
        <div class="port-results">
          <div v-if="openPorts.length === 0" class="no-ports">
            <p>No open ports found on target</p>
            <p class="scan-summary">Scanned all 65535 ports (1-65535) using SYN scan</p>
          </div>
          
          <div v-else>
            <div class="results-summary">
              <h3>Found {{ openPorts.length }} open port(s) out of 65535</h3>
              <div class="summary-stats">
                <span>Open: {{ openPorts.length }}</span>
                <span>Closed: {{ 65535 - openPorts.length }}</span>
                <span>Scan Rate: {{ scanRate }} ports/sec</span>
                <span>Duration: {{ scanDuration }}s</span>
              </div>
            </div>
            
            <div class="port-list">
              <div 
                v-for="(port, index) in sortedOpenPorts" 
                :key="index" 
                class="port-item"
              >
                <div class="port-number">{{ port.port }}</div>
                <div class="port-status open">OPEN</div>
                <div class="port-service">{{ port.service }}</div>
                <div class="port-protocol">{{ port.protocol || 'TCP' }}</div>
                <div class="port-description">{{ port.description }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Statistics -->
        <div class="scan-stats">
          <div class="stat-item">
            <span class="stat-label">Total Ports Scanned:</span>
            <span class="stat-value">65535</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Scan Duration:</span>
            <span class="stat-value">{{ scanDuration }} seconds</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Open Ports Found:</span>
            <span class="stat-value">{{ openPorts.length }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Average Scan Rate:</span>
            <span class="stat-value">{{ scanRate }} ports/sec</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Status:</span>
            <span class="stat-value success">Completed</span>
          </div>
        </div>
        
        <!-- Export Results -->
        <div class="export-section">
          <button @click="exportResults('json')" class="export-btn">
            Export as JSON
          </button>
          <button @click="exportResults('csv')" class="export-btn">
            Export as CSV
          </button>
          <button @click="copyResults" class="export-btn">
            Copy to Clipboard
          </button>
        </div>
      </div>

      <!-- Recent Scans -->
      <div v-if="recentScans.length > 0" class="recent-scans">
        <div class="recent-header">
          <h3>Recent Scans</h3>
          <button @click="clearHistory" class="clear-history-btn">Clear History</button>
        </div>
        <div class="scans-list">
          <div 
            v-for="(scan, index) in recentScans" 
            :key="index" 
            class="scan-item"
            @click="loadScan(scan)"
          >
            <div class="scan-info-main">
              <span class="scan-target">{{ scan.target }}</span>
              <span class="scan-ports">{{ scan.openPorts.length }} open ports</span>
            </div>
            <div class="scan-info-secondary">
              <span class="scan-time">{{ scan.time }}</span>
              <span class="scan-duration">{{ scan.duration }}s</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PortScan',
  data() {
    return {
      scanTarget: 'example.com',
      isScanning: false,
      openPorts: [],
      scanStartTime: null,
      scanDuration: 0,
      error: null,
      recentScans: [],
      showResults: false,
      scanProgress: 0,
      elapsedTime: 0,
      scanStatus: 'Initializing SYN scan...',
      scanInterval: null,
      ws: null,
      scannedPortsCount: 0,
      currentPortRange: '1-65535',
      scanRate: 0,
      connectionStatus: '',
      connectionStatusClass: '',
      hoverBlock: null,
      estimatedTimeRemaining: 0,
      scanStartTimestamp: 0,
      portsScannedLastSecond: 0,
      rateUpdateInterval: null
    }
  },
  computed: {
    formattedScanTime() {
      if (!this.scanStartTime) return '';
      return new Date(this.scanStartTime).toLocaleString();
    },
    sortedOpenPorts() {
      return [...this.openPorts].sort((a, b) => a.port - b.port);
    }
  },
  mounted() {
    const savedScans = localStorage.getItem('portScanHistory');
    if (savedScans) {
      this.recentScans = JSON.parse(savedScans);
    }
  },
  beforeUnmount() {
    this.cleanup();
  },
  methods: {
    getPortBlockStatus(blockIndex) {
      const totalBlocks = 100;
      const blockProgress = (blockIndex / totalBlocks) * 100;
      
      if (this.scanProgress >= blockProgress) {
        // Check if any ports in this block are open
        const startPort = (blockIndex - 1) * 656 + 1;
        const endPort = blockIndex * 656;
        const hasOpenPort = this.openPorts.some(port => 
          port.port >= startPort && port.port <= endPort
        );
        
        if (hasOpenPort) return 'open';
        return 'completed';
      } else if (blockProgress - this.scanProgress <= 1) {
        return 'scanning';
      }
      return 'pending';
    },
    
    getPortBubbleClass(port) {
      if (port <= 1023) return 'well-known';
      if (port <= 49151) return 'registered';
      return 'dynamic';
    },
    
    getServiceByPort(port) {
      // Common port services
      const commonPorts = {
        20: 'FTP Data',
        21: 'FTP Control',
        22: 'SSH',
        23: 'Telnet',
        25: 'SMTP',
        53: 'DNS',
        80: 'HTTP',
        110: 'POP3',
        119: 'NNTP',
        123: 'NTP',
        143: 'IMAP',
        161: 'SNMP',
        194: 'IRC',
        443: 'HTTPS',
        465: 'SMTPS',
        587: 'SMTP Submission',
        993: 'IMAPS',
        995: 'POP3S',
        3306: 'MySQL',
        3389: 'RDP',
        5432: 'PostgreSQL',
        5900: 'VNC',
        8080: 'HTTP Proxy',
      };
      
      return commonPorts[port] || 'Unknown Service';
    },
    
    getPortDescription(port) {
      if (port < 1024) return 'Well-known port (System/Privileged)';
      if (port < 49152) return 'Registered port (User/Application)';
      return 'Dynamic/private port (Ephemeral)';
    },
    
    cleanup() {
      // Clear intervals
      if (this.scanInterval) {
        clearInterval(this.scanInterval);
        this.scanInterval = null;
      }
      
      if (this.rateUpdateInterval) {
        clearInterval(this.rateUpdateInterval);
        this.rateUpdateInterval = null;
      }
      
      // Close WebSocket connection
      if (this.ws) {
        this.ws.close();
        this.ws = null;
      }
    },
    
    updateScanRate() {
      if (this.scanStartTimestamp && this.scannedPortsCount > 0) {
        const elapsed = (Date.now() - this.scanStartTimestamp) / 1000;
        this.scanRate = Math.round(this.scannedPortsCount / elapsed);
        
        // Calculate estimated time remaining
        if (this.scanRate > 0) {
          const remainingPorts = 65535 - this.scannedPortsCount;
          this.estimatedTimeRemaining = Math.round(remainingPorts / this.scanRate);
        }
      }
    },
    
    async startScan() {
      if (!this.scanTarget.trim()) {
        this.error = 'Please enter a target IP or hostname';
        return;
      }

      // Reset state
      this.isScanning = true;
      this.error = null;
      this.openPorts = [];
      this.showResults = false;
      this.scanStartTime = new Date();
      this.scanStartTimestamp = Date.now();
      
      this.scanProgress = 0;
      this.elapsedTime = 0;
      this.scannedPortsCount = 0;
      this.scanRate = 0;
      this.estimatedTimeRemaining = 0;
      this.scanStatus = 'Starting SYN port scan...';
      this.connectionStatus = 'Connecting to scanner...';
      this.connectionStatusClass = 'connecting';
      
      // Clean up any existing connections
      this.cleanup();
      
      // Start WebSocket connection
      this.connectWebSocket();
      
      // Start elapsed time counter
      this.scanInterval = setInterval(() => {
        this.elapsedTime = Math.floor((Date.now() - this.scanStartTimestamp) / 1000);
        this.updateScanRate();
      }, 1000);
      
      // Update progress based on time (since backend doesn't send progress)
      // Your scan takes 27-28 seconds for 65535 ports
      this.progressSimulationInterval = setInterval(() => {
        if (this.scanProgress < 100) {
          // 27 seconds for full scan = ~3.7% per second
          this.scanProgress += 3.7;
          this.scannedPortsCount = Math.floor((this.scanProgress / 100) * 65535);
          
          // Update current port range
          const currentPort = Math.floor(this.scanProgress / 100 * 65535);
          const rangeSize = 1000;
          const rangeStart = Math.floor(currentPort / rangeSize) * rangeSize + 1;
          const rangeEnd = Math.min(rangeStart + rangeSize - 1, 65535);
          this.currentPortRange = `${rangeStart}-${rangeEnd}`;
          
          // Update status based on progress
          if (this.scanProgress < 1.5) {
            this.scanStatus = 'Resolving hostname...';
          } else if (this.scanProgress < 15) {
            this.scanStatus = 'Scanning well-known ports (1-1023)...';
          } else if (this.scanProgress < 75) {
            this.scanStatus = 'Scanning registered ports (1024-49151)...';
          } else {
            this.scanStatus = 'Scanning dynamic ports (49152-65535)...';
          }
          
          if (this.scanProgress >= 100) {
            this.scanProgress = 100;
            this.scannedPortsCount = 65535;
            clearInterval(this.progressSimulationInterval);
          }
        }
      }, 1000);
    },
    
    connectWebSocket() {
      // WebSocket connection to backend
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${wsProtocol}//${window.location.hostname}:8082/v1/scan`;
      
      this.ws = new WebSocket(wsUrl);
      
      this.ws.onopen = () => {
        this.connectionStatus = 'Connected, starting scan...';
        this.connectionStatusClass = 'connected';
        
        // Send scan request
        this.ws.send(JSON.stringify({
          action: 'startScan',
          target: this.scanTarget.trim(),
          scanType: 'syn',
          ports: '1-65535'
        }));
      };
      
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          
          switch (data.type) {
            case 'openPort':
              this.handleOpenPort(data.port);
              break;
              
            case 'progress':
              // If backend sends progress, use it instead of simulation
              if (data.progress !== undefined) {
                this.scanProgress = data.progress;
                this.scannedPortsCount = data.scanned || Math.floor((data.progress / 100) * 65535);
              }
              break;
              
            case 'status':
              this.scanStatus = data.message;
              break;
              
            case 'complete':
              this.handleScanComplete(data);
              break;
              
            case 'error':
              this.error = data.message;
              this.connectionStatus = 'Error: ' + data.message;
              this.connectionStatusClass = 'error';
              this.isScanning = false;
              break;
          }
        } catch (err) {
          console.error('Error parsing WebSocket message:', err);
        }
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket Error:', error);
        this.error = 'Failed to connect to scanner';
        this.connectionStatus = 'Connection failed';
        this.connectionStatusClass = 'error';
        this.isScanning = false;
      };
      
      this.ws.onclose = () => {
        this.connectionStatus = 'Connection closed';
        this.connectionStatusClass = 'disconnected';
        this.ws = null;
        
        // If still scanning, try to reconnect
        if (this.isScanning) {
          this.connectionStatus = 'Reconnecting...';
          this.connectionStatusClass = 'connecting';
          setTimeout(() => this.connectWebSocket(), 3000);
        }
      };
    },
    
    handleOpenPort(portNumber) {
      // Check if port already exists
      if (!this.openPorts.some(p => p.port === portNumber)) {
        const portInfo = {
          port: portNumber,
          service: this.getServiceByPort(portNumber),
          protocol: 'TCP',
          description: this.getPortDescription(portNumber)
        };
        
        this.openPorts.push(portInfo);
        
        // Sort ports numerically
        this.openPorts.sort((a, b) => a.port - b.port);
        
        // Update UI immediately
        this.$forceUpdate();
      }
    },
    
    handleScanComplete(data) {
      this.scanDuration = ((Date.now() - this.scanStartTimestamp) / 1000).toFixed(2);
      this.scanProgress = 100;
      this.scannedPortsCount = 65535;
      this.scanStatus = 'Scan completed successfully';
      this.connectionStatus = 'Scan completed';
      this.connectionStatusClass = 'completed';
      
      // Clear simulation interval
      if (this.progressSimulationInterval) {
        clearInterval(this.progressSimulationInterval);
      }
      
      // Add any remaining ports from complete data
      if (data.openPorts && Array.isArray(data.openPorts)) {
        data.openPorts.forEach(portNumber => {
          this.handleOpenPort(portNumber);
        });
      }
      
      this.showResults = true;
      this.isScanning = false;
      
      // Save to history
      this.saveToHistory();
      
      // Cleanup
      this.cleanup();
    },
    
    stopScan() {
      if (confirm('Stop the current scan?')) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
          this.ws.send(JSON.stringify({ action: 'stopScan' }));
        }
        
        this.isScanning = false;
        this.scanStatus = 'Scan stopped by user';
        this.connectionStatus = 'Scan stopped';
        this.connectionStatusClass = 'stopped';
        
        // Show partial results
        if (this.openPorts.length > 0) {
          this.scanDuration = ((Date.now() - this.scanStartTimestamp) / 1000).toFixed(2);
          this.showResults = true;
          this.saveToHistory();
        }
        
        this.cleanup();
      }
    },
    
    saveToHistory() {
      const scanRecord = {
        target: this.scanTarget.trim(),
        openPorts: this.openPorts,
        time: new Date().toLocaleString(),
        duration: this.scanDuration,
        timestamp: Date.now(),
        totalPorts: this.scannedPortsCount,
        scanType: 'SYN Scan'
      };

      this.recentScans.unshift(scanRecord);
      if (this.recentScans.length > 10) {
        this.recentScans = this.recentScans.slice(0, 10);
      }

      localStorage.setItem('portScanHistory', JSON.stringify(this.recentScans));
    },
    
    exportResults(format) {
      const data = {
        target: this.scanTarget.trim(),
        timestamp: this.formattedScanTime,
        duration: this.scanDuration,
        totalPorts: 65535,
        openPorts: this.openPorts,
        scanType: 'SYN Scan'
      };
      
      let content, mimeType, filename;
      
      if (format === 'json') {
        content = JSON.stringify(data, null, 2);
        mimeType = 'application/json';
        filename = `portscan-${this.scanTarget.replace(/[^a-z0-9]/gi, '-')}-${Date.now()}.json`;
      } else if (format === 'csv') {
        const headers = ['Port', 'Service', 'Protocol', 'Description'];
        const rows = this.openPorts.map(port => 
          [port.port, port.service, port.protocol, port.description].map(cell => 
            `"${cell}"`
          ).join(',')
        );
        content = [headers.join(','), ...rows].join('\n');
        mimeType = 'text/csv';
        filename = `portscan-${this.scanTarget.replace(/[^a-z0-9]/gi, '-')}-${Date.now()}.csv`;
      }
      
      const blob = new Blob([content], { type: mimeType });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    },
    
    copyResults() {
      const text = `Port Scan Results for ${this.scanTarget}\n` +
                   `Scanned on: ${this.formattedScanTime}\n` +
                   `Duration: ${this.scanDuration} seconds\n` +
                   `Open Ports (${this.openPorts.length}):\n` +
                   this.openPorts.map(p => 
                     `  ${p.port} - ${p.service} (${p.protocol})`
                   ).join('\n');
      
      navigator.clipboard.writeText(text).then(() => {
        alert('Results copied to clipboard!');
      }).catch(err => {
        console.error('Failed to copy:', err);
        alert('Failed to copy results');
      });
    },
    
    loadScan(scan) {
      this.scanTarget = scan.target;
      this.openPorts = scan.openPorts;
      this.scanDuration = scan.duration;
      this.showResults = true;
      window.scrollTo({ top: 0, behavior: 'smooth' });
    },
    
    clearHistory() {
      if (confirm('Clear all scan history?')) {
        this.recentScans = [];
        localStorage.removeItem('portScanHistory');
      }
    }
  }
}
</script>

<style scoped>
.portscan-page {
  padding: 20px;
  width: 100%;
  min-height: 100vh;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  color: #e2e8f0;
}

.portscan-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

/* Scan Form */
.scan-form {
  background: rgba(30, 41, 59, 0.8);
  border-radius: 15px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
  margin-bottom: 30px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  backdrop-filter: blur(10px);
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
}

.target-input {
  flex: 1;
  padding: 16px 20px;
  border: 2px solid rgba(148, 163, 184, 0.2);
  border-radius: 10px;
  font-size: 1rem;
  font-family: 'Monaco', 'Courier New', monospace;
  transition: all 0.3s;
  background: rgba(15, 23, 42, 0.7);
  color: #e2e8f0;
}

.target-input::placeholder {
  color: #94a3b8;
}

.target-input:focus {
  outline: none;
  border-color: #60a5fa;
  box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.1);
  background: rgba(15, 23, 42, 0.9);
}

.target-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.scan-button {
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
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
  box-shadow: 0 8px 20px rgba(59, 130, 246, 0.3);
}

.scan-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.scanning-text {
  display: flex;
  align-items: center;
  gap: 10px;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid white;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.stop-button {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  border: none;
  padding: 16px 25px;
  border-radius: 10px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  min-width: 120px;
  white-space: nowrap;
}

.stop-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(239, 68, 68, 0.3);
}

/* Loading Animation */
.loading-section {
  margin-top: 25px;
  padding: 25px;
  background: rgba(15, 23, 42, 0.7);
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.progress-container {
  height: 10px;
  background: rgba(148, 163, 184, 0.1);
  border-radius: 5px;
  overflow: hidden;
  margin-bottom: 30px;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #8b5cf6);
  border-radius: 5px;
  transition: width 0.3s ease;
  position: relative;
}

.progress-bar::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
  animation: shimmer 2s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.loading-info {
  display: flex;
  flex-direction: column;
  gap: 25px;
}

.loading-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
  padding: 20px;
  background: rgba(30, 41, 59, 0.8);
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.loading-stat {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: rgba(15, 23, 42, 0.7);
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.stat-icon {
  font-size: 1.2rem;
}

.stat-text {
  font-weight: 500;
  color: #cbd5e1;
  font-size: 0.95rem;
}

/* Port Visualization */
.port-visualization {
  padding: 25px;
  background: rgba(30, 41, 59, 0.8);
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.vis-title {
  font-weight: 600;
  color: #e2e8f0;
  margin-bottom: 20px;
  text-align: center;
  font-size: 1.1rem;
}

.port-range-info {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.range-start, .range-end {
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  color: #94a3b8;
  white-space: nowrap;
}

.range-progress {
  flex: 1;
}

.range-bar {
  height: 6px;
  background: rgba(148, 163, 184, 0.1);
  border-radius: 3px;
  overflow: hidden;
}

.range-fill {
  height: 100%;
  background: linear-gradient(90deg, #10b981, #0ea5e9);
  border-radius: 3px;
  transition: width 0.3s ease;
}

.current-port {
  text-align: center;
  margin: 15px 0;
  padding: 10px;
  background: rgba(15, 23, 42, 0.7);
  border-radius: 8px;
  font-family: 'Monaco', 'Courier New', monospace;
  color: #cbd5e1;
}

.open-ports-list {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 5px;
  margin-left: 10px;
}

.open-port-badge {
  padding: 4px 8px;
  background: rgba(254, 226, 226, 0.1);
  color: #fecaca;
  border-radius: 4px;
  font-size: 0.9rem;
  font-weight: 600;
  font-family: 'Monaco', 'Courier New', monospace;
  border: 1px solid rgba(254, 202, 202, 0.2);
}

/* Port Grid */
.port-grid {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 4px;
  margin: 20px 0;
  padding: 10px;
  background: rgba(15, 23, 42, 0.7);
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.port-block {
  aspect-ratio: 1;
  border-radius: 3px;
  cursor: default;
  position: relative;
  transition: all 0.3s;
}

.port-block.pending {
  background: rgba(148, 163, 184, 0.1);
}

.port-block.scanning {
  background: linear-gradient(45deg, rgba(139, 92, 246, 0.7), rgba(99, 102, 241, 0.7));
  animation: pulse 2s infinite;
}

.port-block.completed {
  background: rgba(16, 185, 129, 0.3);
}

.port-block.open {
  background: rgba(239, 68, 68, 0.3);
  animation: highlight 1.5s infinite alternate;
}

@keyframes pulse {
  0%, 100% { opacity: 0.7; }
  50% { opacity: 1; }
}

@keyframes highlight {
  from { transform: scale(1); }
  to { transform: scale(1.1); }
}

.block-tooltip {
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background: #1e293b;
  color: #e2e8f0;
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 0.8rem;
  white-space: nowrap;
  z-index: 1000;
  margin-bottom: 5px;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.block-tooltip::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border-width: 5px;
  border-style: solid;
  border-color: #1e293b transparent transparent transparent;
}

.vis-legend {
  display: flex;
  justify-content: center;
  gap: 20px;
  flex-wrap: wrap;
  margin-top: 20px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
  color: #94a3b8;
}

.dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.dot.pending { background: rgba(148, 163, 184, 0.3); }
.dot.scanning { background: rgba(139, 92, 246, 0.7); }
.dot.completed { background: rgba(16, 185, 129, 0.5); }
.dot.open { background: rgba(239, 68, 68, 0.5); }

/* Real-time Ports */
.realtime-ports {
  margin: 20px 0;
  padding: 20px;
  background: rgba(30, 41, 59, 0.8);
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.realtime-ports h3 {
  margin: 0 0 15px 0;
  color: #e2e8f0;
  font-size: 1.2rem;
}

.ports-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.port-bubble {
  padding: 12px 15px;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  min-width: 80px;
  animation: slideIn 0.5s ease-out;
  backdrop-filter: blur(10px);
}

.port-bubble.well-known {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.2), rgba(5, 150, 105, 0.3));
  border: 2px solid rgba(16, 185, 129, 0.4);
}

.port-bubble.registered {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(37, 99, 235, 0.3));
  border: 2px solid rgba(59, 130, 246, 0.4);
}

.port-bubble.dynamic {
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.2), rgba(124, 58, 237, 0.3));
  border: 2px solid rgba(139, 92, 246, 0.4);
}

.port-bubble .port-number {
  font-size: 1.4rem;
  font-weight: 700;
  font-family: 'Monaco', 'Courier New', monospace;
  color: white;
}

.port-bubble .port-service {
  font-size: 0.8rem;
  font-weight: 500;
  text-align: center;
  color: #cbd5e1;
}

.port-bubble .port-status {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 2px 8px;
  background: rgba(255, 255, 255, 0.1);
  color: #86efac;
  border-radius: 10px;
  backdrop-filter: blur(5px);
}

/* Scanning Tips */
.scanning-tips {
  display: flex;
  gap: 20px;
  padding: 25px;
  background: rgba(146, 64, 14, 0.1);
  border-radius: 10px;
  border: 1px solid rgba(251, 191, 36, 0.2);
  animation: slideIn 0.5s ease;
}

@keyframes slideIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.tip-icon {
  font-size: 2.5rem;
  flex-shrink: 0;
}

.tip-content {
  flex: 1;
}

.tip-content strong {
  display: block;
  color: #fbbf24;
  margin-bottom: 10px;
  font-size: 1.1rem;
}

.tip-content p {
  margin: 0 0 15px 0;
  color: #fde68a;
  line-height: 1.5;
}

.tip-content ul {
  margin: 0;
  padding-left: 20px;
  color: #fde68a;
}

.tip-content li {
  margin-bottom: 5px;
  font-size: 0.9rem;
}

/* Connection Status */
.connection-status {
  margin-top: 10px;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  border-left: 4px solid;
}

.connection-status.connecting {
  background: rgba(251, 191, 36, 0.1);
  color: #fbbf24;
  border-color: #f59e0b;
}

.connection-status.connected {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border-color: #10b981;
}

.connection-status.error {
  background: rgba(239, 68, 68, 0.1);
  color: #fca5a5;
  border-color: #ef4444;
}

.connection-status.disconnected {
  background: rgba(148, 163, 184, 0.1);
  color: #94a3b8;
  border-color: #64748b;
}

.connection-status.completed {
  background: rgba(59, 130, 246, 0.1);
  color: #93c5fd;
  border-color: #3b82f6;
}

.connection-status.stopped {
  background: rgba(236, 72, 153, 0.1);
  color: #f9a8d4;
  border-color: #ec4899;
}

.error-message {
  color: #fca5a5;
  margin-top: 10px;
  font-size: 0.9rem;
  padding: 8px 12px;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 6px;
  border-left: 4px solid #ef4444;
}

/* Results Section */
.results-section {
  background: rgba(30, 41, 59, 0.8);
  border-radius: 15px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
  margin-bottom: 30px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  backdrop-filter: blur(10px);
  animation: slideInUp 0.5s ease-out;
}

@keyframes slideInUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 25px;
  padding-bottom: 20px;
  border-bottom: 2px solid rgba(148, 163, 184, 0.2);
}

.results-header h2 {
  margin: 0;
  color: #e2e8f0;
  font-size: 1.8rem;
}

.results-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 5px;
}

.target-display {
  font-family: 'Monaco', 'Courier New', monospace;
  font-weight: 600;
  color: #cbd5e1;
  font-size: 1.1rem;
}

.scan-time {
  color: #94a3b8;
  font-size: 0.9rem;
}

/* Port Results */
.port-results {
  margin: 20px 0;
}

.no-ports {
  text-align: center;
  padding: 60px 40px;
  color: #94a3b8;
  background: rgba(15, 23, 42, 0.7);
  border-radius: 10px;
  border: 2px dashed rgba(148, 163, 184, 0.2);
}

.no-ports p {
  margin: 0 0 10px 0;
  font-size: 1.2rem;
}

.scan-summary {
  font-size: 0.9rem !important;
  color: #64748b;
}

.results-summary {
  margin-bottom: 30px;
  padding: 20px;
  background: rgba(15, 23, 42, 0.7);
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.results-summary h3 {
  margin: 0 0 15px 0;
  color: #e2e8f0;
  font-size: 1.3rem;
}

.summary-stats {
  display: flex;
  gap: 30px;
  flex-wrap: wrap;
}

.summary-stats span {
  padding: 8px 16px;
  background: rgba(30, 41, 59, 0.8);
  color: #cbd5e1;
  border-radius: 6px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  font-weight: 500;
}

.port-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 20px;
}

.port-item {
  background: rgba(15, 23, 42, 0.7);
  border-radius: 12px;
  padding: 25px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  border: 2px solid rgba(148, 163, 184, 0.1);
  transition: all 0.3s;
  text-align: center;
  backdrop-filter: blur(10px);
}

.port-item:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
  border-color: rgba(96, 165, 250, 0.3);
}

.port-number {
  font-size: 2.2rem;
  font-weight: 700;
  color: white;
  font-family: 'Monaco', 'Courier New', monospace;
  line-height: 1;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.port-status {
  padding: 6px 16px;
  border-radius: 20px;
  font-size: 0.9rem;
  font-weight: 600;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

.port-status.open {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.port-service {
  color: #cbd5e1;
  font-weight: 600;
  font-size: 1rem;
  margin-top: 5px;
}

.port-protocol {
  color: #94a3b8;
  font-size: 0.85rem;
  padding: 4px 10px;
  background: rgba(30, 41, 59, 0.8);
  border-radius: 4px;
  font-family: 'Monaco', 'Courier New', monospace;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.port-description {
  color: #64748b;
  font-size: 0.8rem;
  margin-top: 5px;
  font-style: italic;
}

/* Statistics */
.scan-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 25px;
  padding-top: 30px;
  margin-top: 30px;
  border-top: 2px solid rgba(148, 163, 184, 0.2);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 20px;
  background: rgba(15, 23, 42, 0.7);
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  backdrop-filter: blur(10px);
}

.stat-label {
  color: #94a3b8;
  font-size: 0.9rem;
  font-weight: 500;
}

.stat-value {
  font-weight: 700;
  font-size: 1.5rem;
  color: #e2e8f0;
}

.stat-value.success {
  color: #34d399;
}

/* Export Section */
.export-section {
  display: flex;
  gap: 15px;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid rgba(148, 163, 184, 0.2);
}

.export-btn {
  padding: 10px 20px;
  background: rgba(148, 163, 184, 0.1);
  color: #cbd5e1;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.export-btn:hover {
  background: rgba(148, 163, 184, 0.2);
  transform: translateY(-2px);
  color: #e2e8f0;
}

/* Recent Scans */
.recent-scans {
  background: rgba(30, 41, 59, 0.8);
  border-radius: 15px;
  padding: 25px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(148, 163, 184, 0.1);
  backdrop-filter: blur(10px);
}

.recent-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.recent-header h3 {
  margin: 0;
  color: #e2e8f0;
  font-size: 1.3rem;
}

.clear-history-btn {
  padding: 8px 16px;
  background: rgba(239, 68, 68, 0.1);
  color: #fca5a5;
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.clear-history-btn:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #fecaca;
}

.scans-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.scan-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18px 20px;
  background: rgba(15, 23, 42, 0.7);
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  cursor: pointer;
  transition: all 0.3s;
  backdrop-filter: blur(10px);
}

.scan-item:hover {
  background: rgba(30, 41, 59, 0.9);
  border-color: rgba(96, 165, 250, 0.3);
  transform: translateX(5px);
}

.scan-info-main {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.scan-target {
  font-weight: 600;
  color: #e2e8f0;
  font-family: 'Monaco', 'Courier New', monospace;
}

.scan-ports {
  color: #94a3b8;
  font-size: 0.85rem;
}

.scan-info-secondary {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 5px;
}

.scan-time {
  color: #94a3b8;
  font-size: 0.85rem;
}

.scan-duration {
  background: rgba(148, 163, 184, 0.1);
  color: #cbd5e1;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-family: 'Monaco', 'Courier New', monospace;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

/* Responsive */
@media (max-width: 768px) {
  .portscan-container {
    padding: 10px;
  }
  
  .scan-form {
    padding: 20px;
  }
  
  .input-with-button {
    flex-direction: column;
  }
  
  .target-input,
  .scan-button {
    width: 100%;
  }
  
  .results-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }
  
  .results-meta {
    align-items: flex-start;
  }
  
  .port-list {
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  }
  
  .scan-stats {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .scan-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .scan-info-secondary {
    align-items: flex-start;
  }
  
  .loading-stats {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 480px) {
  .scan-stats {
    grid-template-columns: 1fr;
  }
  
  .loading-stats {
    grid-template-columns: 1fr;
  }
  
  .export-section {
    flex-direction: column;
  }
  
  .export-btn {
    width: 100%;
  }
}
</style>