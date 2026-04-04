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
          
          <!-- Circular Loading Animation -->
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
              
              <!-- Circular Progress Visualization -->
              <div class="circular-progress-section">
                <div class="circular-progress-container">
                  <div class="circular-progress">
                    <svg class="circular-progress-svg" viewBox="0 0 100 100">
                      <!-- Background circle -->
                      <circle
                        class="circular-bg"
                        cx="50"
                        cy="50"
                        r="45"
                        fill="none"
                        stroke="#1e293b"
                        stroke-width="8"
                      />
                      <!-- Progress circle -->
                      <circle
                        class="circular-fill"
                        cx="50"
                        cy="50"
                        r="45"
                        fill="none"
                        stroke="url(#gradient)"
                        stroke-width="8"
                        stroke-linecap="round"
                        :stroke-dasharray="283"
                        :stroke-dashoffset="283 - (283 * scanProgress / 100)"
                        transform="rotate(-90 50 50)"
                      />
                      <defs>
                        <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="0%">
                          <stop offset="0%" stop-color="#3b82f6" />
                          <stop offset="100%" stop-color="#8b5cf6" />
                        </linearGradient>
                      </defs>
                    </svg>
                    <div class="circular-progress-text">
                      <div class="circular-percentage">{{ Math.round(scanProgress) }}%</div>
                      <div class="circular-label">Complete</div>
                    </div>
                  </div>
                  
                  <div class="circular-stats">
                    <div class="circular-stat-item">
                      <span class="stat-label">Scanned</span>
                      <span class="stat-number">{{ scannedPortsCount.toLocaleString() }}</span>
                    </div>
                    <div class="circular-stat-item">
                      <span class="stat-label">Total</span>
                      <span class="stat-number">65,535</span>
                    </div>
                    <div class="circular-stat-item">
                      <span class="stat-label">Remaining</span>
                      <span class="stat-number">{{ (65535 - scannedPortsCount).toLocaleString() }}</span>
                    </div>
                  </div>
                </div>
                
                <div class="current-scan-info">
                  <div class="info-row">
                    <span class="info-label">Current Range:</span>
                    <span class="info-value">{{ currentPortRange }}</span>
                  </div>
                  <div class="info-row">
                    <span class="info-label">Est. Time Left:</span>
                    <span class="info-value">{{ estimatedTimeRemaining }} seconds</span>
                  </div>
                </div>
              </div>
              
              <!-- Real-time open ports table during scan -->
              <div v-if="openPorts.length > 0" class="realtime-ports-section">
                <div class="section-header">
                  <h3>Open Ports Found ({{ openPorts.length }})</h3>
                  <span class="badge">{{ getPortCategories() }}</span>
                </div>
                <div class="table-container">
                  <table class="ports-table">
                    <thead>
                      <tr>
                        <th>Port</th>
                        <th>Service</th>
                        <th>Protocol</th>
                        <th>Category</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr 
                        v-for="port in sortedOpenPorts" 
                        :key="port.port"
                        class="port-row"
                        :class="getPortRowClass(port.port)"
                      >
                        <td class="port-number-cell">{{ port.port }}</td>
                        <td class="port-service-cell">{{ port.service }}</td>
                        <td class="port-protocol-cell">{{ port.protocol }}</td>
                        <td class="port-category-cell">{{ getPortCategory(port.port) }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
              
              <!-- Empty state when no ports found yet -->
              <div v-else class="no-ports-yet">
                <div class="no-ports-icon">🔍</div>
                <p>No open ports discovered yet. Scanning in progress...</p>
                <div class="scanning-pulse">
                  <span class="pulse-dot"></span>
                  <span class="pulse-dot"></span>
                  <span class="pulse-dot"></span>
                </div>
              </div>
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
          <h2>Scan Results for {{ scanTarget }}</h2>
          <div class="results-meta">
            <span class="scan-time">{{ formattedScanTime }}</span>
            <span class="scan-duration-badge">{{ scanDuration }}s</span>
          </div>
        </div>

        <!-- Summary Cards -->
        <div class="summary-cards">
          <div class="summary-card">
            <div class="summary-icon"></div>
            <div class="summary-content">
              <div class="summary-value">{{ scannedPortsCount.toLocaleString() }}</div>
              <div class="summary-label">Ports Scanned</div>
            </div>
          </div>
          <div class="summary-card">
            <div class="summary-icon"></div>
            <div class="summary-content">
              <div class="summary-value">{{ openPorts.length }}</div>
              <div class="summary-label">Open Ports</div>
            </div>
          </div>
          <div class="summary-card">
            <div class="summary-icon"></div>
            <div class="summary-content">
              <div class="summary-value">{{ scanRate }}</div>
              <div class="summary-label">Scan Rate (ports/s)</div>
            </div>
          </div>
          <div class="summary-card">
            <div class="summary-icon"></div>
            <div class="summary-content">
              <div class="summary-value">{{ getWellKnownCount() }}</div>
              <div class="summary-label">Well-Known Ports</div>
            </div>
          </div>
        </div>

        <!-- Port List Table -->
        <div class="port-results">
          <div v-if="openPorts.length === 0" class="no-ports">
            <div class="no-ports-icon">🔌</div>
            <h3>No Open Ports Found</h3>
            <p>The target {{ scanTarget }} has no open ports in the range 1-65535</p>
            <p class="scan-summary">This could mean the host is down, firewalled, or has no services running</p>
          </div>
          
          <div v-else>
            <div class="results-table-container">
              <table class="results-table">
                <thead>
                  <tr>
                    <th>Port</th>
                    <th>Service</th>
                    <th>Protocol</th>
                    <th>Category</th>
                    <th>Description</th>
                  </tr>
                </thead>
                <tbody>
                  <tr 
                    v-for="port in sortedOpenPorts" 
                    :key="port.port"
                    class="result-row"
                    :class="getPortRowClass(port.port)"
                  >
                    <td class="port-number-cell">{{ port.port }}</td>
                    <td class="port-service-cell">{{ port.service }}</td>
                    <td class="port-protocol-cell">{{ port.protocol }}</td>
                    <td class="port-category-cell">{{ getPortCategory(port.port) }}</td>
                    <td class="port-description-cell">{{ port.description }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Port Distribution -->
            <div class="port-distribution">
              <h4>Port Distribution</h4>
              <div class="distribution-bars">
                <div class="distribution-item">
                  <span class="dist-label">Well-Known (1-1023)</span>
                  <div class="dist-bar-container">
                    <div class="dist-bar well-known" :style="{ width: (getWellKnownCount() / openPorts.length * 100) + '%' }"></div>
                  </div>
                  <span class="dist-count">{{ getWellKnownCount() }}</span>
                </div>
                <div class="distribution-item">
                  <span class="dist-label">Registered (1024-49151)</span>
                  <div class="dist-bar-container">
                    <div class="dist-bar registered" :style="{ width: (getRegisteredCount() / openPorts.length * 100) + '%' }"></div>
                  </div>
                  <span class="dist-count">{{ getRegisteredCount() }}</span>
                </div>
                <div class="distribution-item">
                  <span class="dist-label">Dynamic (49152-65535)</span>
                  <div class="dist-bar-container">
                    <div class="dist-bar dynamic" :style="{ width: (getDynamicCount() / openPorts.length * 100) + '%' }"></div>
                  </div>
                  <span class="dist-count">{{ getDynamicCount() }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Export Actions -->
        <div class="export-section">
          <button @click="exportResults('json')" class="export-btn json">
            <span class="btn-icon">📋</span> Export JSON
          </button>
          <button @click="exportResults('csv')" class="export-btn csv">
            <span class="btn-icon">📊</span> Export CSV
          </button>
          <button @click="copyResults" class="export-btn copy">
            <span class="btn-icon">📑</span> Copy to Clipboard
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
import { ref, computed, onMounted, onUnmounted } from 'vue';

const API_URL = "http://192.168.20.17:8082";

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
      progressSimulationInterval: null,
      ws: null,
      scannedPortsCount: 0,
      currentPortRange: '1-65535',
      scanRate: 0,
      connectionStatus: '',
      connectionStatusClass: '',
      scanStartTimestamp: 0,
      estimatedTimeRemaining: 0
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
    getPortRowClass(port) {
      if (port <= 1023) return 'well-known';
      if (port <= 49151) return 'registered';
      return 'dynamic';
    },
    
    getPortCategory(port) {
      if (port <= 1023) return 'Well-Known';
      if (port <= 49151) return 'Registered';
      return 'Dynamic';
    },
    
    getPortCategories() {
      const wellKnown = this.openPorts.filter(p => p.port <= 1023).length;
      const registered = this.openPorts.filter(p => p.port > 1023 && p.port <= 49151).length;
      const dynamic = this.openPorts.filter(p => p.port > 49151).length;
      return `${wellKnown} Well-Known, ${registered} Registered, ${dynamic} Dynamic`;
    },
    
    getWellKnownCount() {
      return this.openPorts.filter(p => p.port <= 1023).length;
    },
    
    getRegisteredCount() {
      return this.openPorts.filter(p => p.port > 1023 && p.port <= 49151).length;
    },
    
    getDynamicCount() {
      return this.openPorts.filter(p => p.port > 49151).length;
    },
    
    getServiceByPort(port) {
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
        135: 'RPC',
        137: 'NetBIOS-NS',
        138: 'NetBIOS-DGM',
        139: 'NetBIOS-SSN',
        143: 'IMAP',
        161: 'SNMP',
        162: 'SNMP Trap',
        389: 'LDAP',
        443: 'HTTPS',
        445: 'SMB',
        465: 'SMTPS',
        514: 'Syslog',
        515: 'LPD',
        587: 'SMTP Submission',
        631: 'IPP',
        993: 'IMAPS',
        995: 'POP3S',
        1433: 'MSSQL',
        3306: 'MySQL',
        3389: 'RDP',
        5432: 'PostgreSQL',
        5900: 'VNC',
        6379: 'Redis',
        8080: 'HTTP Proxy',
        8443: 'HTTPS Alt',
        27017: 'MongoDB'
      };
      return commonPorts[port] || 'Unknown';
    },
    
    getPortDescription(port) {
      if (port < 1024) return 'Well-known port (system services)';
      if (port < 49152) return 'Registered port (user applications)';
      return 'Dynamic/private port (ephemeral)';
    },
    
    cleanup() {
      if (this.scanInterval) {
        clearInterval(this.scanInterval);
        this.scanInterval = null;
      }
      if (this.progressSimulationInterval) {
        clearInterval(this.progressSimulationInterval);
        this.progressSimulationInterval = null;
      }
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
      
      this.cleanup();
      this.connectWebSocket();
      
      this.scanInterval = setInterval(() => {
        this.elapsedTime = Math.floor((Date.now() - this.scanStartTimestamp) / 1000);
        this.updateScanRate();
      }, 1000);
      
      // Progress simulation
      this.progressSimulationInterval = setInterval(() => {
        if (this.scanProgress < 100) {
          this.scanProgress += 3.7;
          this.scannedPortsCount = Math.floor((this.scanProgress / 100) * 65535);
          
          const currentPort = Math.floor(this.scanProgress / 100 * 65535);
          const rangeSize = 1000;
          const rangeStart = Math.floor(currentPort / rangeSize) * rangeSize + 1;
          const rangeEnd = Math.min(rangeStart + rangeSize - 1, 65535);
          this.currentPortRange = `${rangeStart}-${rangeEnd}`;
          
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
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      // const wsUrl = `${wsProtocol}//${window.location.hostname}:8082/v1/api/scan`;
      const wsUrl = `${wsProtocol}//${window.location.hostname}:8082/scan`;
      
      this.ws = new WebSocket(wsUrl);
      
      this.ws.onopen = () => {
        this.connectionStatus = 'Connected, starting scan...';
        this.connectionStatusClass = 'connected';
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
        if (this.isScanning) {
          this.connectionStatus = 'Reconnecting...';
          this.connectionStatusClass = 'connecting';
          setTimeout(() => this.connectWebSocket(), 3000);
        }
      };
    },
    
    handleOpenPort(portNumber) {
      if (!this.openPorts.some(p => p.port === portNumber)) {
        this.openPorts.push({
          port: portNumber,
          service: this.getServiceByPort(portNumber),
          protocol: 'TCP',
          description: this.getPortDescription(portNumber)
        });
        this.openPorts.sort((a, b) => a.port - b.port);
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
      
      if (this.progressSimulationInterval) {
        clearInterval(this.progressSimulationInterval);
      }
      
      if (data.openPorts && Array.isArray(data.openPorts)) {
        data.openPorts.forEach(portNumber => {
          this.handleOpenPort(portNumber);
        });
      }
      
      this.showResults = true;
      this.isScanning = false;
      this.saveToHistory();
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
        openPorts: this.openPorts.map(p => ({
          port: p.port,
          service: p.service,
          protocol: p.protocol
        })),
        scanType: 'SYN Scan'
      };
      
      let content, mimeType, filename;
      
      if (format === 'json') {
        content = JSON.stringify(data, null, 2);
        mimeType = 'application/json';
        filename = `portscan-${this.scanTarget.replace(/[^a-z0-9]/gi, '-')}-${Date.now()}.json`;
      } else if (format === 'csv') {
        const headers = ['Port', 'Service', 'Protocol'];
        const rows = this.openPorts.map(port => 
          [port.port, port.service, port.protocol].join(',')
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
      const text = `Port Scan Results for ${this.scanTarget}
================================
Scan Time: ${this.formattedScanTime}
Duration: ${this.scanDuration} seconds
Total Ports Scanned: 65535
Open Ports Found: ${this.openPorts.length}

Open Ports:
${this.openPorts.map(p => `  ${p.port}/tcp - ${p.service}`).join('\n')}`;
      
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
  padding: 24px;
  width: 100%;
  min-height: 100vh;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  color: #e2e8f0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.portscan-container {
  max-width: 1400px;
  margin: 0 auto;
}

/* Scan Form */
.scan-form {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  margin-bottom: 24px;
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
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.target-input {
  flex: 1;
  min-width: 300px;
  padding: 14px 18px;
  border: 1px solid #334155;
  border-radius: 10px;
  font-size: 1rem;
  font-family: 'Monaco', 'Courier New', monospace;
  transition: all 0.3s;
  background: #0f172a;
  color: #e2e8f0;
}

.target-input::placeholder {
  color: #64748b;
}

.target-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

.target-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.scan-button, .stop-button {
  padding: 14px 28px;
  border: none;
  border-radius: 10px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  white-space: nowrap;
}

.scan-button {
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  color: white;
}

.scan-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(59, 130, 246, 0.3);
}

.scan-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.stop-button {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}

.stop-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(239, 68, 68, 0.3);
}

.scanning-text {
  display: flex;
  align-items: center;
  gap: 10px;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Loading Section */
.loading-section {
  margin-top: 24px;
  padding: 24px;
  background: #0f172a;
  border-radius: 12px;
  border: 1px solid #334155;
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.progress-container {
  height: 8px;
  background: #1e293b;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 24px;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #8b5cf6);
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
  gap: 24px;
}

.loading-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
}

.loading-stat {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #1e293b;
  border-radius: 8px;
  border: 1px solid #334155;
}

.stat-icon {
  font-size: 1.2rem;
}

.stat-text {
  font-weight: 500;
  color: #cbd5e1;
  font-size: 0.9rem;
}

/* Circular Progress Section */
.circular-progress-section {
  padding: 20px;
  background: #1e293b;
  border-radius: 12px;
  border: 1px solid #334155;
}

.circular-progress-container {
  display: flex;
  align-items: center;
  gap: 40px;
  flex-wrap: wrap;
  justify-content: center;
}

.circular-progress {
  position: relative;
  width: 180px;
  height: 180px;
}

.circular-progress-svg {
  width: 100%;
  height: 100%;
  transform: rotate(0deg);
}

.circular-bg {
  stroke: #0f172a;
}

.circular-fill {
  transition: stroke-dashoffset 0.3s ease;
  filter: drop-shadow(0 0 8px rgba(59, 130, 246, 0.3));
}

.circular-progress-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
}

.circular-percentage {
  font-size: 2rem;
  font-weight: 700;
  color: #f8fafc;
  line-height: 1.2;
}

.circular-label {
  font-size: 0.8rem;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.circular-stats {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  min-width: 250px;
}

.circular-stat-item {
  text-align: center;
  padding: 12px;
  background: #0f172a;
  border-radius: 8px;
  border: 1px solid #334155;
}

.circular-stat-item .stat-label {
  display: block;
  font-size: 0.75rem;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 4px;
}

.circular-stat-item .stat-number {
  display: block;
  font-size: 1.2rem;
  font-weight: 600;
  color: #60a5fa;
}

.current-scan-info {
  margin-top: 20px;
  padding: 16px;
  background: #0f172a;
  border-radius: 8px;
  border: 1px solid #334155;
  display: flex;
  justify-content: space-around;
  flex-wrap: wrap;
  gap: 20px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.info-label {
  color: #94a3b8;
  font-size: 0.9rem;
}

.info-value {
  color: #60a5fa;
  font-weight: 600;
  font-family: 'Monaco', 'Courier New', monospace;
  background: #1e293b;
  padding: 4px 12px;
  border-radius: 20px;
  border: 1px solid #334155;
}

/* Real-time Ports Table */
.realtime-ports-section {
  margin-top: 20px;
  background: #1e293b;
  border-radius: 12px;
  border: 1px solid #334155;
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: #0f172a;
  border-bottom: 1px solid #334155;
}

.section-header h3 {
  margin: 0;
  color: #f8fafc;
  font-size: 1.1rem;
}

.badge {
  background: #3b82f6;
  color: white;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
}

.table-container {
  max-height: 300px;
  overflow-y: auto;
}

.ports-table, .results-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.ports-table th, .results-table th {
  position: sticky;
  top: 0;
  background: #0f172a;
  padding: 12px 16px;
  text-align: left;
  color: #94a3b8;
  font-weight: 600;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid #334155;
  z-index: 10;
}

.ports-table td, .results-table td {
  padding: 10px 16px;
  border-bottom: 1px solid #1e293b;
}

.port-row, .result-row {
  transition: background-color 0.2s;
}

.port-row:hover, .result-row:hover {
  background: #2d3748;
}

.port-row.well-known, .result-row.well-known {
  background: rgba(16, 185, 129, 0.05);
}

.port-row.registered, .result-row.registered {
  background: rgba(59, 130, 246, 0.05);
}

.port-row.dynamic, .result-row.dynamic {
  background: rgba(139, 92, 246, 0.05);
}

.port-number-cell {
  font-family: 'Monaco', 'Courier New', monospace;
  font-weight: 600;
  color: #f8fafc;
}

.port-service-cell {
  color: #cbd5e1;
  font-weight: 500;
}

.port-protocol-cell {
  color: #94a3b8;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.8rem;
}

.port-category-cell {
  font-size: 0.8rem;
  padding: 4px 8px;
  border-radius: 4px;
  display: inline-block;
}

.well-known .port-category-cell {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
}

.registered .port-category-cell {
  background: rgba(59, 130, 246, 0.1);
  color: #60a5fa;
}

.dynamic .port-category-cell {
  background: rgba(139, 92, 246, 0.1);
  color: #c4b5fd;
}

.port-description-cell {
  color: #94a3b8;
  font-size: 0.8rem;
  font-style: italic;
}

.no-ports-yet {
  text-align: center;
  padding: 40px;
  color: #64748b;
  background: #1e293b;
  border-radius: 8px;
  border: 1px dashed #334155;
}

.no-ports-icon {
  font-size: 3rem;
  margin-bottom: 16px;
  opacity: 0.5;
}

.scanning-pulse {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 20px;
}

.pulse-dot {
  width: 12px;
  height: 12px;
  background: #3b82f6;
  border-radius: 50%;
  animation: pulse 1.5s ease-in-out infinite;
}

.pulse-dot:nth-child(2) {
  animation-delay: 0.2s;
}

.pulse-dot:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
    opacity: 0.5;
  }
  50% {
    transform: scale(1.2);
    opacity: 1;
  }
}

/* Results Section */
.results-section {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 24px;
  border: 1px solid rgba(148, 163, 184, 0.1);
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
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #334155;
}

.results-header h2 {
  margin: 0;
  color: #f8fafc;
  font-size: 1.5rem;
}

.results-meta {
  display: flex;
  align-items: center;
  gap: 16px;
}

.scan-time {
  color: #94a3b8;
  font-size: 0.9rem;
}

.scan-duration-badge {
  background: #1e293b;
  color: #60a5fa;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 0.9rem;
  font-weight: 600;
  border: 1px solid #334155;
}

/* Summary Cards */
.summary-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.summary-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #0f172a;
  border-radius: 12px;
  border: 1px solid #334155;
  transition: all 0.2s;
}

.summary-card:hover {
  transform: translateY(-2px);
  border-color: #3b82f6;
}

.summary-icon {
  font-size: 2rem;
}

.summary-content {
  flex: 1;
}

.summary-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #f8fafc;
  line-height: 1.2;
  margin-bottom: 4px;
}

.summary-label {
  color: #94a3b8;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Results Table */
.results-table-container {
  max-height: 400px;
  overflow-y: auto;
  border-radius: 8px;
  border: 1px solid #334155;
  background: #0f172a;
  margin-bottom: 24px;
}

.results-table {
  width: 100%;
  border-collapse: collapse;
}

.results-table th {
  position: sticky;
  top: 0;
  background: #0f172a;
  padding: 14px 16px;
  text-align: left;
  color: #94a3b8;
  font-weight: 600;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid #334155;
  z-index: 10;
}

.results-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #1e293b;
}

/* Port Distribution */
.port-distribution {
  margin-top: 24px;
  padding: 20px;
  background: #0f172a;
  border-radius: 12px;
  border: 1px solid #334155;
}

.port-distribution h4 {
  margin: 0 0 16px 0;
  color: #f8fafc;
  font-size: 1rem;
}

.distribution-bars {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.distribution-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dist-label {
  min-width: 120px;
  color: #cbd5e1;
  font-size: 0.9rem;
}

.dist-bar-container {
  flex: 1;
  height: 24px;
  background: #1e293b;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid #334155;
}

.dist-bar {
  height: 100%;
  transition: width 0.3s ease;
}

.dist-bar.well-known {
  background: linear-gradient(90deg, #10b981, #059669);
}

.dist-bar.registered {
  background: linear-gradient(90deg, #3b82f6, #2563eb);
}

.dist-bar.dynamic {
  background: linear-gradient(90deg, #8b5cf6, #7c3aed);
}

.dist-count {
  min-width: 60px;
  color: #f8fafc;
  font-weight: 600;
  text-align: right;
}

/* No Ports State */
.no-ports {
  text-align: center;
  padding: 60px 40px;
  background: #0f172a;
  border-radius: 12px;
  border: 1px dashed #334155;
}

.no-ports-icon {
  font-size: 4rem;
  margin-bottom: 20px;
  opacity: 0.5;
}

.no-ports h3 {
  color: #f8fafc;
  margin-bottom: 10px;
  font-size: 1.3rem;
}

.no-ports p {
  color: #94a3b8;
  margin-bottom: 8px;
}

.scan-summary {
  font-size: 0.9rem;
  color: #64748b;
}

/* Export Section */
.export-section {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  flex-wrap: wrap;
}

.export-btn {
  padding: 12px 20px;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.export-btn.json {
  background: #3b82f6;
  color: white;
}

.export-btn.json:hover {
  background: #2563eb;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.export-btn.csv {
  background: #10b981;
  color: white;
}

.export-btn.csv:hover {
  background: #059669;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}

.export-btn.copy {
  background: #1e293b;
  color: #cbd5e1;
  border: 1px solid #334155;
}

.export-btn.copy:hover {
  background: #2d3748;
  transform: translateY(-2px);
}

.btn-icon {
  font-size: 1.1rem;
}

/* Recent Scans */
.recent-scans {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 24px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.recent-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.recent-header h3 {
  margin: 0;
  color: #f8fafc;
  font-size: 1.2rem;
}

.clear-history-btn {
  padding: 8px 16px;
  background: #1e293b;
  color: #f87171;
  border: 1px solid #334155;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.clear-history-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: #ef4444;
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
  padding: 16px 20px;
  background: #0f172a;
  border-radius: 10px;
  border: 1px solid #334155;
  cursor: pointer;
  transition: all 0.2s;
}

.scan-item:hover {
  background: #1e293b;
  border-color: #3b82f6;
  transform: translateX(5px);
}

.scan-info-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.scan-target {
  font-weight: 600;
  color: #f8fafc;
  font-family: 'Monaco', 'Courier New', monospace;
}

.scan-ports {
  color: #94a3b8;
  font-size: 0.8rem;
}

.scan-info-secondary {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.scan-time {
  color: #94a3b8;
  font-size: 0.8rem;
}

.scan-duration {
  background: #1e293b;
  color: #60a5fa;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-family: 'Monaco', 'Courier New', monospace;
  border: 1px solid #334155;
}

/* Error and Connection Status */
.error-message {
  margin-top: 16px;
  padding: 12px 16px;
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border-radius: 8px;
  border-left: 4px solid #ef4444;
}

.connection-status {
  margin-top: 12px;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
}

.connection-status.connecting {
  background: rgba(245, 158, 11, 0.1);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.connection-status.connected {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.connection-status.error {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.connection-status.disconnected {
  background: rgba(148, 163, 184, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(148, 163, 184, 0.3);
}

/* Scrollbar */
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

/* Responsive */
@media (max-width: 768px) {
  .portscan-container {
    padding: 12px;
  }
  
  .input-with-button {
    flex-direction: column;
  }
  
  .target-input {
    width: 100%;
    min-width: auto;
  }
  
  .scan-button, .stop-button {
    width: 100%;
  }
  
  .results-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .results-meta {
    width: 100%;
    justify-content: space-between;
  }
  
  .summary-cards {
    grid-template-columns: 1fr;
  }
  
  .export-section {
    flex-direction: column;
  }
  
  .export-btn {
    width: 100%;
    justify-content: center;
  }
  
  .circular-progress-container {
    flex-direction: column;
    gap: 20px;
  }
  
  .circular-stats {
    width: 100%;
  }
  
  .distribution-item {
    flex-wrap: wrap;
  }
  
  .dist-label {
    min-width: 100%;
  }
  
  .scan-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .scan-info-secondary {
    align-items: flex-start;
    width: 100%;
    flex-direction: row;
    justify-content: space-between;
  }
  
  .current-scan-info {
    flex-direction: column;
    align-items: center;
  }
}

@media (max-width: 480px) {
  .loading-stats {
    grid-template-columns: 1fr;
  }
  
  .circular-stats {
    grid-template-columns: 1fr;
  }
}
</style>