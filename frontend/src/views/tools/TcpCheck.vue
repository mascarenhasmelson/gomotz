<template>
  <div class="tcp-check-page">
    <div class="tcp-check-container">
    

      <!-- Main Form -->
      <div class="check-form">
        <div class="form-grid">
          <div class="form-group">
            <label for="host-input">
              <span class="label-icon">🌐</span>
              Host / IP Address
            </label>
            <div class="input-with-validation">
              <input
                v-model="host"
                id="host-input"
                type="text"
                placeholder="Enter domain name or IP address"
                :disabled="isChecking"
                class="host-input"
                @keyup.enter="checkTcp"
              />
              <div v-if="hostSuggestions.length && host" class="suggestions">
                <div 
                  v-for="suggestion in hostSuggestions" 
                  :key="suggestion"
                  @click="selectSuggestion(suggestion)"
                  class="suggestion-item"
                >
                  {{ suggestion }}
                </div>
              </div>
            </div>
          </div>

          <div class="form-group">
            <label for="port-input">
              <span class="label-icon">🔌</span>
              Port Number
            </label>
            <div class="port-input-container">
              <input
                v-model.number="port"
                id="port-input"
                type="number"
                min="1"
                max="65535"
                placeholder="1-65535"
                :disabled="isChecking"
                class="port-input"
                @keyup.enter="checkTcp"
              />
              <div class="common-ports">
                <div class="ports-header">Common Ports:</div>
                <div class="ports-grid">
                  <button 
                    v-for="commonPort in commonPorts" 
                    :key="commonPort.port"
                    @click="setPort(commonPort.port)"
                    class="port-btn"
                    :class="{ 'active': port === commonPort.port }"
                    :title="commonPort.service"
                  >
                    <span class="port-number">{{ commonPort.port }}</span>
                    <span class="port-service">{{ commonPort.service }}</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Quick Test Buttons -->
        <div class="quick-tests">
          <h4>Quick Tests:</h4>
          <div class="quick-buttons">
            <button 
              v-for="preset in presets" 
              :key="preset.label"
              @click="usePreset(preset)"
              class="quick-btn"
              :class="{ 'active': isActivePreset(preset) }"
            >
              <span class="quick-icon">{{ preset.icon }}</span>
              <span class="quick-label">{{ preset.label }}</span>
            </button>
          </div>
        </div>

        <!-- Form Actions -->
        <div class="form-actions">
          <button 
            @click="checkTcp" 
            :disabled="!canCheck || isChecking" 
            class="check-button"
          >
            <span v-if="!isChecking">
              <span class="button-icon">🔍</span>
              Test Connection
            </span>
            <span v-else>
              <span class="spinner"></span>
              Testing...
            </span>
          </button>
          
          <button 
            @click="stopTest" 
            v-if="isChecking"
            class="stop-button"
          >
            <span class="button-icon">⏹️</span>
            Stop Test
          </button>

          <button 
            @click="clearForm" 
            :disabled="isChecking"
            class="clear-button"
          >
            Clear
          </button>

          <button 
            @click="toggleAdvanced" 
            class="advanced-toggle-btn"
          >
            <span class="button-icon">{{ showAdvanced ? '▲' : '▼' }}</span>
            Advanced
          </button>
        </div>

        <!-- Advanced Options -->
        <div v-if="showAdvanced" class="advanced-options">
          <div class="advanced-grid">
            <div class="option-group">
              <label>
                <input type="checkbox" v-model="options.customTimeout" />
                Custom Timeout
              </label>
              <div class="option-control">
                <input 
                  v-model.number="options.timeoutValue" 
                  type="range" 
                  min="1" 
                  max="10" 
                  step="1"
                  :disabled="!options.customTimeout"
                  class="timeout-slider"
                />
                <span class="slider-value">{{ options.timeoutValue }}s</span>
              </div>
            </div>
            
            <div class="option-group">
              <label>
                <input type="checkbox" v-model="options.retry" />
                Auto Retry
              </label>
              <div class="option-control">
                <input 
                  v-model.number="options.retryCount" 
                  type="range" 
                  min="1" 
                  max="5" 
                  step="1"
                  :disabled="!options.retry"
                  class="retry-slider"
                />
                <span class="slider-value">{{ options.retryCount }} times</span>
              </div>
            </div>
            
            <div class="option-group">
              <label>
                <input type="checkbox" v-model="options.verbose" />
                Verbose Output
              </label>
            </div>
          </div>
        </div>

        <div v-if="error" class="error-message">
          <span class="error-icon">⚠️</span>
          {{ error }}
        </div>
      </div>

      <!-- Real-time Progress -->
      <div v-if="isChecking" class="progress-section">
        <div class="progress-info">
          <div class="progress-text">
            <span class="status-icon">🔄</span>
            Testing connection to {{ host }}:{{ port }}
          </div>
          <div class="progress-time">Elapsed: {{ elapsedTime }}s</div>
        </div>
        <div class="progress-container">
          <div class="progress-bar" :style="{ width: progress + '%' }">
            <div class="progress-indicator"></div>
          </div>
        </div>
        <div class="progress-steps">
          <div class="step" :class="{ 'active': currentStep >= 1 }">
            <div class="step-dot"></div>
            <div class="step-label">DNS Lookup</div>
          </div>
          <div class="step" :class="{ 'active': currentStep >= 2 }">
            <div class="step-dot"></div>
            <div class="step-label">Sending SYN</div>
          </div>
          <div class="step" :class="{ 'active': currentStep >= 3 }">
            <div class="step-dot"></div>
            <div class="step-label">Waiting Response</div>
          </div>
          <div class="step" :class="{ 'active': currentStep >= 4 }">
            <div class="step-dot"></div>
            <div class="step-label">Analyzing</div>
          </div>
        </div>
      </div>

      <!-- Results Section -->
      <div v-if="showResults" class="results-section">
        <div class="results-header">
          <div class="result-title">
            <h2>Test Results</h2>
            <div class="result-target">
              <span class="target-icon">🎯</span>
              {{ host }}:<span class="target-port">{{ port }}</span>
            </div>
          </div>
          <div class="result-actions">
            <button @click="copyResult" class="action-btn" title="Copy results">
              <span class="action-icon">📋</span>
              Copy
            </button>
            <button @click="exportResult" class="action-btn" title="Export results">
              <span class="action-icon">📤</span>
              Export
            </button>
            <button @click="saveResult" class="action-btn" title="Save to history">
              <span class="action-icon">💾</span>
              Save
            </button>
          </div>
        </div>

        <div class="result-card" :class="result.status">
          <div class="result-status">
            <div class="status-icon-large">
              <span v-if="result.status === 'open'">✅</span>
              <span v-else-if="result.status === 'closed'">❌</span>
              <span v-else-if="result.status === 'filtered'">🛡️</span>
              <span v-else>⏳</span>
            </div>
            <div class="status-content">
              <h3 class="status-title">{{ result.title }}</h3>
              <p class="status-message">{{ result.message }}</p>
              <div class="status-details">
                <span class="detail-item">
                  <span class="detail-label">Response Time:</span>
                  <span class="detail-value">{{ result.responseTime }} ms</span>
                </span>
                <span class="detail-item">
                  <span class="detail-label">Test Duration:</span>
                  <span class="detail-value">{{ result.duration }} ms</span>
                </span>
                <span class="detail-item">
                  <span class="detail-label">Timestamp:</span>
                  <span class="detail-value">{{ result.timestamp }}</span>
                </span>
              </div>
            </div>
          </div>

          <div v-if="result.details" class="technical-details">
            <div class="details-header" @click="toggleDetails">
              <span class="details-title">Technical Details</span>
              <span class="toggle-icon">{{ showDetails ? '▲' : '▼' }}</span>
            </div>
            <div v-if="showDetails" class="details-content">
              <div class="details-grid">
                <div class="detail-row">
                  <span class="detail-label">Protocol:</span>
                  <span class="detail-value">TCP</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">SYN Sent:</span>
                  <span class="detail-value">{{ result.details.synSent || 'Yes' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">SYN-ACK Received:</span>
                  <span class="detail-value">{{ result.details.synAckReceived || 'No' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Packet Size:</span>
                  <span class="detail-value">{{ result.details.packetSize || '60 bytes' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">TTL Used:</span>
                  <span class="detail-value">{{ result.details.ttl || '64' }}</span>
                </div>
              </div>
              
              <div v-if="result.banner" class="banner-section">
                <h4>Service Banner:</h4>
                <div class="banner-content">
                  <pre>{{ result.banner }}</pre>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Recommendation -->
        <div v-if="result.recommendation" class="recommendation">
          <div class="recommendation-icon">💡</div>
          <div class="recommendation-content">
            <strong>{{ result.recommendation.title }}</strong>
            <p>{{ result.recommendation.message }}</p>
          </div>
        </div>
      </div>

      <!-- History Sidebar -->
      <div v-if="showHistory" class="history-sidebar" :class="{ 'visible': showHistory }">
        <div class="history-header">
          <h3>Test History</h3>
          <button @click="toggleHistory" class="close-history" title="Close">
            <span>×</span>
          </button>
        </div>
        <div class="history-stats">
          <div class="history-stat">
            <span class="stat-value">{{ stats.total }}</span>
            <span class="stat-label">Total Tests</span>
          </div>
          <div class="history-stat">
            <span class="stat-value">{{ stats.open }}</span>
            <span class="stat-label">Open</span>
          </div>
          <div class="history-stat">
            <span class="stat-value">{{ stats.closed }}</span>
            <span class="stat-label">Closed</span>
          </div>
        </div>
        <div class="history-list">
          <div 
            v-for="(item, index) in history" 
            :key="index" 
            class="history-item"
            @click="loadHistoryItem(item)"
          >
            <div class="history-status" :class="item.status"></div>
            <div class="history-info">
              <div class="history-target">{{ item.host }}:{{ item.port }}</div>
              <div class="history-time">{{ item.time }}</div>
            </div>
            <div class="history-response">{{ item.responseTime }}ms</div>
            <div class="history-action">
              <button @click.stop="deleteHistoryItem(index)" class="delete-btn" title="Delete">
                ×
              </button>
            </div>
          </div>
        </div>
        <div v-if="history.length === 0" class="empty-history">
          <div class="empty-icon">📋</div>
          <p>No test history yet</p>
        </div>
        <div v-else class="history-actions">
          <button @click="clearHistory" class="clear-history-btn">
            Clear All History
          </button>
          <button @click="exportHistory" class="export-history-btn">
            Export History
          </button>
        </div>
      </div>

      <!-- Main History Button -->
      <button 
        @click="toggleHistory" 
        class="history-toggle-btn"
        :class="{ 'active': showHistory }"
      >
        <span class="history-icon">📊</span>
        <span class="history-text">History ({{ history.length }})</span>
      </button>

      <!-- Statistics -->
      <div v-if="stats.total > 0" class="stats-section">
        <h3>Statistics</h3>
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-icon">📈</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.total }}</div>
              <div class="stat-label">Total Tests</div>
            </div>
          </div>
          <div class="stat-card success">
            <div class="stat-icon">✅</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.open }}</div>
              <div class="stat-label">Open Ports</div>
            </div>
          </div>
          <div class="stat-card failed">
            <div class="stat-icon">❌</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.closed }}</div>
              <div class="stat-label">Closed Ports</div>
            </div>
          </div>
          <div class="stat-card warning">
            <div class="stat-icon">🛡️</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.filtered }}</div>
              <div class="stat-label">Filtered</div>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon">⚡</div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.averageTime }}ms</div>
              <div class="stat-label">Avg Response</div>
            </div>
          </div>
        </div>
        
        <div v-if="mostTested.length > 0" class="most-tested">
          <h4>Most Tested Targets:</h4>
          <div class="targets-list">
            <div 
              v-for="target in mostTested" 
              :key="target.host"
              class="target-item"
              @click="loadTarget(target)"
            >
              <span class="target-name">{{ target.host }}</span>
              <span class="target-count">{{ target.count }} tests</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'TcpCheck',
  data() {
    return {
      host: '1.1.1.1',
      port: 80,
      isChecking: false,
      showAdvanced: false,
      showHistory: false,
      showResults: false,
      showDetails: false,
      error: null,
      connectionStatus: '',
      currentStep: 0,
      progress: 0,
      elapsedTime: 0,
      progressInterval: null,
      ws: null,
      testStartTime: null,
      
      result: {
        status: '', // 'open', 'closed', 'filtered', 'error'
        title: '',
        message: '',
        responseTime: 0,
        timestamp: '',
        duration: 0,
        details: null,
        banner: '',
        recommendation: null
      },
      
      history: [],
      stats: {
        total: 0,
        open: 0,
        closed: 0,
        filtered: 0,
        averageTime: 0
      },
      
      options: {
        customTimeout: false,
        timeoutValue: 5,
        retry: false,
        retryCount: 3,
        verbose: false
      },
      
      presets: [
        { label: 'Web Server', icon: '🌐', host: 'google.com', port: 80 },
        { label: 'Secure Web', icon: '🔒', host: 'google.com', port: 443 },
        { label: 'DNS Server', icon: '📡', host: '8.8.8.8', port: 53 },
        { label: 'SSH Server', icon: '💻', host: 'github.com', port: 22 },
        { label: 'Email Server', icon: '📧', host: 'smtp.gmail.com', port: 587 },
        { label: 'MySQL', icon: '🗄️', host: 'localhost', port: 3306 }
      ],
      
      commonPorts: [
        { port: 21, service: 'FTP' },
        { port: 22, service: 'SSH' },
        { port: 25, service: 'SMTP' },
        { port: 53, service: 'DNS' },
        { port: 80, service: 'HTTP' },
        { port: 110, service: 'POP3' },
        { port: 143, service: 'IMAP' },
        { port: 443, service: 'HTTPS' },
        { port: 587, service: 'SMTP SSL' },
        { port: 993, service: 'IMAP SSL' },
        { port: 995, service: 'POP3 SSL' },
        { port: 3306, service: 'MySQL' },
        { port: 3389, service: 'RDP' },
        { port: 5432, service: 'PostgreSQL' },
        { port: 8080, service: 'HTTP Alt' },
        { port: 8443, service: 'HTTPS Alt' }
      ]
    }
  },
  computed: {
    canCheck() {
      return this.host.trim() && this.port > 0 && this.port <= 65535
    },
    
    hostSuggestions() {
      if (!this.host) return []
      const input = this.host.toLowerCase()
      const suggestions = [
        'google.com', 'github.com', 'cloudflare.com', 
        '1.1.1.1', '8.8.8.8', 'localhost', '127.0.0.1'
      ]
      return suggestions
        .filter(h => h.includes(input) && h !== this.host)
        .slice(0, 5)
    },
    
    isActivePreset() {
      return (preset) => {
        return this.host === preset.host && this.port == preset.port
      }
    },
    
    mostTested() {
      const counts = {}
      this.history.forEach(item => {
        const key = `${item.host}:${item.port}`
        counts[key] = (counts[key] || 0) + 1
      })
      
      return Object.entries(counts)
        .map(([key, count]) => {
          const [host, port] = key.split(':')
          return { host, port, count }
        })
        .sort((a, b) => b.count - a.count)
        .slice(0, 5)
    }
  },
  mounted() {
    this.loadHistory()
  },
  beforeUnmount() {
    this.cleanup()
  },
  methods: {
    usePreset(preset) {
      this.host = preset.host
      this.port = preset.port
      this.showResults = false
      this.error = null
    },
    
    setPort(portNumber) {
      this.port = portNumber
    },
    
    selectSuggestion(suggestion) {
      this.host = suggestion
    },
    
    toggleAdvanced() {
      this.showAdvanced = !this.showAdvanced
    },
    
    toggleHistory() {
      this.showHistory = !this.showHistory
    },
    
    toggleDetails() {
      this.showDetails = !this.showDetails
    },
    
    clearForm() {
      this.host = ''
      this.port = ''
      this.showResults = false
      this.error = null
      this.cleanup()
    },
    
    async checkTcp() {
      if (!this.canCheck) {
        this.error = 'Please enter a valid host and port (1-65535)'
        return
      }
      this.isChecking = true
      this.showResults = false
      this.error = null
      this.result = {
        status: '',
        title: '',
        message: '',
        responseTime: 0,
        timestamp: '',
        duration: 0,
        details: null,
        banner: ''
      }
      
      this.testStartTime = Date.now()
      this.elapsedTime = 0
      this.currentStep = 0
      this.progress = 0
      this.startProgressSimulation()
      try {
        const response = await fetch('http://localhost:8082/v1/tcpCheck', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            host: this.host.trim(),
            port: this.port,
            timeout: this.options.customTimeout ? this.options.timeoutValue : null
          })
        })
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }
        
        const data = await response.json()
        this.processResult(data)
        
      } catch (err) {
        console.error('TCP check error:', err)
        this.handleError(err)
      } finally {
        this.isChecking = false
        this.cleanup()
      }
    },
    
    startProgressSimulation() {
      this.progressInterval = setInterval(() => {
        this.elapsedTime = ((Date.now() - this.testStartTime) / 1000).toFixed(1)
        if (this.elapsedTime < 0.5) {
          this.currentStep = 1 
          this.progress = 25
        } else if (this.elapsedTime < 1) {
          this.currentStep = 2 
          this.progress = 50
        } else if (this.elapsedTime < 2) {
          this.currentStep = 3 
          this.progress = 75
        } else if (this.elapsedTime < 3) {
          this.currentStep = 4 
          this.progress = 95
        }
      }, 100)
    },
    
    processResult(data) {
      const duration = Date.now() - this.testStartTime
      let status = 'closed'
      let title = 'Port Closed'
      let message = 'No response received'
      let responseTime = duration
      
      if (data && data.status === 'open') {
        status = 'open'
        title = 'Port Open'
        message = 'response received - Port is open'
      } else if (data && data.status === 'filtered') {
        status = 'filtered'
        title = 'Port Filtered'
        message = 'Port may be filtered by firewall'
      }
      
      this.result = {
        status,
        title,
        message,
        responseTime: Math.round(responseTime),
        timestamp: new Date().toLocaleString(),
        duration: Math.round(duration),
        details: {
          synSent: 'Yes',
          synAckReceived: status === 'open' ? 'Yes' : 'No',
          packetSize: '60 bytes',
          ttl: '64'
        },
        banner: data.banner || '',
        recommendation: this.getRecommendation(status)
      }
      
      this.showResults = true
      this.progress = 100
      this.saveToHistory()
    },
    
    getRecommendation(status) {
      switch(status) {
        case 'open':
          return {
            title: 'Port Security Check',
            message: 'An open port may be a security risk. Consider firewall rules.'
          }
        case 'closed':
          return {
            title: 'Standard Security',
            message: 'Closed ports are secure by default.'
          }
        case 'filtered':
          return {
            title: 'Firewall Detected',
            message: 'A firewall appears to be blocking connections.'
          }
        default:
          return null
      }
    },
    
    handleError(err) {
      const duration = Date.now() - this.testStartTime
      
      this.result = {
        status: 'error',
        title: 'Test Failed',
        message: err.message || 'Connection test failed',
        responseTime: 0,
        timestamp: new Date().toLocaleString(),
        duration: Math.round(duration),
        details: null,
        banner: '',
        recommendation: {
          title: 'Troubleshooting',
          message: 'Check your network connection and ensure the host is reachable.'
        }
      }
      
      this.showResults = true
      this.progress = 100
    },
    
    stopTest() {
      this.isChecking = false
      this.cleanup()
      
      this.result = {
        status: 'stopped',
        title: 'Test Stopped',
        message: 'Test stopped by user',
        responseTime: 0,
        timestamp: new Date().toLocaleString(),
        duration: Date.now() - this.testStartTime,
        details: null
      }
      
      this.showResults = true
    },
    
    cleanup() {
      if (this.progressInterval) {
        clearInterval(this.progressInterval)
        this.progressInterval = null
      }
      
      if (this.ws) {
        this.ws.close()
        this.ws = null
      }
    },
    
    saveToHistory() {
      const historyItem = {
        host: this.host,
        port: this.port,
        status: this.result.status,
        responseTime: this.result.responseTime,
        time: new Date().toLocaleTimeString(),
        date: new Date().toLocaleDateString(),
        timestamp: Date.now()
      }
      
      this.history.unshift(historyItem)
      if (this.history.length > 50) {
        this.history = this.history.slice(0, 50)
      }
      
      this.updateStats()
      localStorage.setItem('tcpCheckHistory', JSON.stringify(this.history))
    },
    
    loadHistory() {
      const saved = localStorage.getItem('tcpCheckHistory')
      if (saved) {
        this.history = JSON.parse(saved)
        this.updateStats()
      }
    },
    
    loadHistoryItem(item) {
      this.host = item.host
      this.port = item.port
      this.showHistory = false
      this.showResults = false
    },
    
    loadTarget(target) {
      this.host = target.host
      this.port = parseInt(target.port)
      this.showResults = false
    },
    
    deleteHistoryItem(index) {
      this.history.splice(index, 1)
      this.updateStats()
      localStorage.setItem('tcpCheckHistory', JSON.stringify(this.history))
    },
    
    clearHistory() {
      if (confirm('Clear all test history?')) {
        this.history = []
        this.updateStats()
        localStorage.removeItem('tcpCheckHistory')
      }
    },
    
    updateStats() {
      const open = this.history.filter(h => h.status === 'open').length
      const closed = this.history.filter(h => h.status === 'closed').length
      const filtered = this.history.filter(h => h.status === 'filtered').length
      const total = this.history.length
      
      const avgTime = total > 0 
        ? Math.round(this.history.reduce((sum, h) => sum + h.responseTime, 0) / total)
        : 0
      
      this.stats = { total, open, closed, filtered, averageTime: avgTime }
    },
    
    copyResult() {
      const text = `
TCP Port Check Result
=====================
Target: ${this.host}:${this.port}
Status: ${this.result.title}
Message: ${this.result.message}
Response Time: ${this.result.responseTime}ms
Test Duration: ${this.result.duration}ms
Timestamp: ${this.result.timestamp}
${this.result.banner ? `Banner:\n${this.result.banner}` : ''}
      `.trim()
      
      navigator.clipboard.writeText(text)
        .then(() => {
          // Show success message
          alert('Result copied to clipboard!')
        })
        .catch(err => {
          console.error('Copy failed:', err)
          alert('Failed to copy result')
        })
    },
    
    exportResult() {
      const data = {
        target: `${this.host}:${this.port}`,
        result: this.result,
        timestamp: new Date().toISOString()
      }
      
      const dataStr = JSON.stringify(data, null, 2)
      const dataUri = 'data:application/json;charset=utf-8,' + encodeURIComponent(dataStr)
      
      const link = document.createElement('a')
      link.setAttribute('href', dataUri)
      link.setAttribute('download', `tcp-check-${this.host}-${this.port}-${Date.now()}.json`)
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    },
    
    saveResult() {
      this.saveToHistory()
      alert('Result saved to history!')
    },
    
    exportHistory() {
      const data = {
        history: this.history,
        stats: this.stats,
        exported: new Date().toISOString()
      }
      
      const dataStr = JSON.stringify(data, null, 2)
      const dataUri = 'data:application/json;charset=utf-8,' + encodeURIComponent(dataStr)
      
      const link = document.createElement('a')
      link.setAttribute('href', dataUri)
      link.setAttribute('download', `tcp-check-history-${Date.now()}.json`)
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    }
  }
}
</script>

<style scoped>
.tcp-check-page {
  padding: 20px;
  width: 100%;
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.tcp-check-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

/* Header */
.check-header {
  text-align: center;
  margin-bottom: 40px;
  padding: 30px;
  background: white;
  border-radius: 15px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
}

.check-header h1 {
  font-size: 2.5rem;
  color: #2d3748;
  margin: 0 0 10px 0;
  font-weight: 700;
  background: linear-gradient(135deg, #4299e1 0%, #3182ce 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.subtitle {
  color: #718096;
  font-size: 1.1rem;
  margin: 0;
}

/* Form */
.check-form {
  background: white;
  border-radius: 15px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  margin-bottom: 30px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 30px;
  margin-bottom: 25px;
}

@media (max-width: 768px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 12px;
  font-size: 1rem;
}

.label-icon {
  font-size: 1.2rem;
}

.input-with-validation {
  position: relative;
}

.host-input, .port-input {
  width: 100%;
  padding: 16px 20px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  font-size: 1rem;
  font-family: 'Monaco', 'Courier New', monospace;
  transition: all 0.3s;
  background: #f8fafc;
}

.host-input:focus, .port-input:focus {
  outline: none;
  border-color: #4299e1;
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
  background: white;
}

.suggestions {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  margin-top: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 1000;
  overflow: hidden;
}

.suggestion-item {
  padding: 12px 20px;
  cursor: pointer;
  transition: all 0.2s;
  border-bottom: 1px solid #f7fafc;
}

.suggestion-item:hover {
  background: #f8fafc;
}

.suggestion-item:last-child {
  border-bottom: none;
}

.port-input-container {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.common-ports {
  margin-top: 10px;
}

.ports-header {
  font-size: 0.9rem;
  color: #718096;
  margin-bottom: 10px;
  font-weight: 500;
}

.ports-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 8px;
}

@media (max-width: 1024px) {
  .ports-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 480px) {
  .ports-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

.port-btn {
  padding: 8px 6px;
  border: 2px solid #e2e8f0;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.port-btn:hover {
  border-color: #cbd5e0;
  transform: translateY(-2px);
}

.port-btn.active {
  background: #4299e1;
  border-color: #3182ce;
  color: white;
}

.port-number {
  font-weight: 700;
  font-size: 0.9rem;
  font-family: 'Monaco', 'Courier New', monospace;
}

.port-service {
  font-size: 0.7rem;
  opacity: 0.8;
}

/* Quick Tests */
.quick-tests {
  margin: 25px 0;
  padding: 20px;
  background: #f8fafc;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
}

.quick-tests h4 {
  margin: 0 0 15px 0;
  color: #4a5568;
  font-size: 1.1rem;
}

.quick-buttons {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

@media (max-width: 768px) {
  .quick-buttons {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .quick-buttons {
    grid-template-columns: 1fr;
  }
}

.quick-btn {
  padding: 12px 15px;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  background: white;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 10px;
}

.quick-btn:hover {
  border-color: #cbd5e0;
  transform: translateY(-2px);
}

.quick-btn.active {
  background: #4299e1;
  border-color: #3182ce;
  color: white;
}

.quick-icon {
  font-size: 1.2rem;
}

.quick-label {
  font-weight: 500;
  font-size: 0.9rem;
}

/* Form Actions */
.form-actions {
  display: flex;
  gap: 15px;
  margin: 25px 0;
  flex-wrap: wrap;
}

.check-button, .stop-button, .clear-button, .advanced-toggle-btn {
  padding: 16px 30px;
  border: none;
  border-radius: 10px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 1rem;
}

.check-button {
  background: linear-gradient(135deg, #4299e1 0%, #3182ce 100%);
  color: white;
  flex: 1;
}

.check-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(66, 153, 225, 0.3);
}

.check-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.stop-button {
  background: linear-gradient(135deg, #fc8181 0%, #e53e3e 100%);
  color: white;
}

.stop-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(229, 62, 62, 0.3);
}

.clear-button {
  background: #e2e8f0;
  color: #4a5568;
}

.clear-button:hover:not(:disabled) {
  background: #cbd5e0;
  transform: translateY(-2px);
}

.advanced-toggle-btn {
  background: #f8fafc;
  color: #4a5568;
  border: 2px solid #e2e8f0;
}

.advanced-toggle-btn:hover {
  border-color: #cbd5e0;
  transform: translateY(-2px);
}

.button-icon {
  font-size: 1.1rem;
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

/* Advanced Options */
.advanced-options {
  margin: 20px 0;
  padding: 25px;
  background: #f8fafc;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.advanced-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 25px;
}

.option-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.option-group label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 500;
  color: #4a5568;
  cursor: pointer;
}

.option-control {
  display: flex;
  align-items: center;
  gap: 15px;
}

.timeout-slider, .retry-slider {
  flex: 1;
  height: 6px;
  background: #e2e8f0;
  border-radius: 3px;
  outline: none;
  -webkit-appearance: none;
}

.timeout-slider::-webkit-slider-thumb,
.retry-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 20px;
  height: 20px;
  background: #4299e1;
  border-radius: 50%;
  cursor: pointer;
  border: 3px solid white;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
}

.slider-value {
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  color: #718096;
  min-width: 80px;
}

/* Error Message */
.error-message {
  margin-top: 20px;
  padding: 15px;
  background: #fed7d7;
  color: #c53030;
  border-radius: 8px;
  border-left: 4px solid #e53e3e;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.95rem;
}

.error-icon {
  font-size: 1.2rem;
}

/* Progress Section */
.progress-section {
  background: white;
  border-radius: 15px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  margin-bottom: 30px;
  animation: slideInUp 0.5s ease-out;
}

@keyframes slideInUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

.progress-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.progress-text {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 500;
  color: #4a5568;
  font-size: 1.1rem;
}

.status-icon {
  font-size: 1.3rem;
}

.progress-time {
  font-family: 'Monaco', 'Courier New', monospace;
  color: #718096;
  font-size: 0.9rem;
}

.progress-container {
  height: 10px;
  background: #e2e8f0;
  border-radius: 5px;
  overflow: hidden;
  margin-bottom: 30px;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #4299e1, #3182ce);
  border-radius: 5px;
  transition: width 0.3s ease;
  position: relative;
}

.progress-indicator {
  position: absolute;
  top: 0;
  right: 0;
  width: 3px;
  height: 100%;
  background: white;
  animation: progressPulse 1.5s infinite;
}

@keyframes progressPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.progress-steps {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  opacity: 0.5;
  transition: all 0.3s;
}

.step.active {
  opacity: 1;
}

.step-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #e2e8f0;
  transition: all 0.3s;
}

.step.active .step-dot {
  background: #4299e1;
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.2);
}

.step-label {
  font-size: 0.85rem;
  color: #718096;
  text-align: center;
  font-weight: 500;
}

/* Results Section */
.results-section {
  background: white;
  border-radius: 15px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  margin-bottom: 30px;
  animation: slideInUp 0.5s ease-out;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 25px;
  padding-bottom: 20px;
  border-bottom: 2px solid #e2e8f0;
}

.result-title h2 {
  margin: 0;
  color: #2d3748;
  font-size: 1.8rem;
}

.result-target {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 10px;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 1.1rem;
  color: #4a5568;
}

.target-icon {
  font-size: 1.2rem;
}

.target-port {
  font-weight: 700;
  color: #4299e1;
}

.result-actions {
  display: flex;
  gap: 10px;
}

.action-btn {
  padding: 10px 20px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #4a5568;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn:hover {
  background: #edf2f7;
  border-color: #cbd5e0;
  transform: translateY(-2px);
}

.action-icon {
  font-size: 1.1rem;
}

/* Result Card */
.result-card {
  border-radius: 12px;
  overflow: hidden;
  border: 2px solid #e2e8f0;
  transition: all 0.3s;
}

.result-card.open {
  border-color: #38a169;
  background: linear-gradient(135deg, #f0fff4 0%, #c6f6d5 100%);
}

.result-card.closed {
  border-color: #e53e3e;
  background: linear-gradient(135deg, #fff5f5 0%, #fed7d7 100%);
}

.result-card.filtered {
  border-color: #ed8936;
  background: linear-gradient(135deg, #fffaf0 0%, #feebc8 100%);
}

.result-card.error {
  border-color: #718096;
  background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
}

.result-status {
  display: flex;
  align-items: center;
  gap: 25px;
  padding: 30px;
}

@media (max-width: 768px) {
  .result-status {
    flex-direction: column;
    text-align: center;
    gap: 20px;
  }
}

.status-icon-large {
  font-size: 3.5rem;
  flex-shrink: 0;
}

.status-content {
  flex: 1;
}

.status-title {
  margin: 0 0 10px 0;
  color: #2d3748;
  font-size: 1.5rem;
}

.status-message {
  margin: 0 0 15px 0;
  color: #4a5568;
  font-size: 1.1rem;
  line-height: 1.5;
}

.status-details {
  display: flex;
  gap: 30px;
  flex-wrap: wrap;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.detail-label {
  font-size: 0.85rem;
  color: #718096;
  font-weight: 500;
}

.detail-value {
  font-family: 'Monaco', 'Courier New', monospace;
  font-weight: 600;
  color: #2d3748;
  font-size: 1.1rem;
}

/* Technical Details */
.technical-details {
  border-top: 1px solid rgba(0, 0, 0, 0.1);
  background: rgba(255, 255, 255, 0.5);
}

.details-header {
  padding: 20px 30px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  transition: all 0.3s;
}

.details-header:hover {
  background: rgba(0, 0, 0, 0.02);
}

.details-title {
  font-weight: 600;
  color: #4a5568;
  font-size: 1.1rem;
}

.toggle-icon {
  font-size: 1.2rem;
  color: #718096;
}

.details-content {
  padding: 0 30px 30px 30px;
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: white;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.banner-section {
  margin-top: 25px;
}

.banner-section h4 {
  margin: 0 0 15px 0;
  color: #4a5568;
  font-size: 1.1rem;
}

.banner-content {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 20px;
  max-height: 200px;
  overflow-y: auto;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  line-height: 1.4;
}

/* Recommendation */
.recommendation {
  margin-top: 25px;
  padding: 25px;
  background: linear-gradient(135deg, #bee3f8 0%, #90cdf4 100%);
  border-radius: 10px;
  border: 1px solid #63b3ed;
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

.recommendation-icon {
  font-size: 2rem;
  flex-shrink: 0;
}

.recommendation-content {
  flex: 1;
}

.recommendation-content strong {
  display: block;
  color: #2c5282;
  margin-bottom: 8px;
  font-size: 1.1rem;
}

.recommendation-content p {
  margin: 0;
  color: #2c5282;
  opacity: 0.9;
  line-height: 1.5;
}

/* History Sidebar */
.history-sidebar {
  position: fixed;
  top: 0;
  right: -400px;
  width: 400px;
  height: 100vh;
  background: white;
  box-shadow: -10px 0 30px rgba(0, 0, 0, 0.1);
  transition: right 0.3s ease;
  z-index: 1000;
  display: flex;
  flex-direction: column;
}

.history-sidebar.visible {
  right: 0;
}

.history-header {
  padding: 25px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid #e2e8f0;
}

.history-header h3 {
  margin: 0;
  color: #2d3748;
  font-size: 1.5rem;
}

.close-history {
  width: 40px;
  height: 40px;
  border: none;
  background: #f8fafc;
  border-radius: 50%;
  font-size: 1.5rem;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-history:hover {
  background: #e53e3e;
  color: white;
}

.history-stats {
  padding: 20px 25px;
  display: flex;
  gap: 20px;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

.history-stat {
  flex: 1;
  text-align: center;
}

.stat-value {
  display: block;
  font-size: 1.5rem;
  font-weight: 700;
  color: #4299e1;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 0.8rem;
  color: #718096;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.history-list {
  flex: 1;
  overflow-y: auto;
  padding: 10px 0;
}

.history-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px 25px;
  border-bottom: 1px solid #f7fafc;
  cursor: pointer;
  transition: all 0.3s;
}

.history-item:hover {
  background: #f8fafc;
}

.history-status {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.history-status.open {
  background: #38a169;
}

.history-status.closed {
  background: #e53e3e;
}

.history-status.filtered {
  background: #ed8936;
}

.history-status.error {
  background: #718096;
}

.history-info {
  flex: 1;
  min-width: 0;
}

.history-target {
  font-family: 'Monaco', 'Courier New', monospace;
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.history-time {
  font-size: 0.8rem;
  color: #a0aec0;
}

.history-response {
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  color: #718096;
  font-weight: 500;
}

.history-action {
  opacity: 0;
  transition: opacity 0.3s;
}

.history-item:hover .history-action {
  opacity: 1;
}

.delete-btn {
  width: 24px;
  height: 24px;
  border: none;
  background: #fed7d7;
  color: #c53030;
  border-radius: 50%;
  font-size: 1.2rem;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.delete-btn:hover {
  background: #feb2b2;
}

.empty-history {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #a0aec0;
  padding: 40px;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 20px;
  opacity: 0.5;
}

.history-actions {
  padding: 20px 25px;
  display: flex;
  gap: 10px;
  border-top: 1px solid #e2e8f0;
}

.clear-history-btn, .export-history-btn {
  flex: 1;
  padding: 12px;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.clear-history-btn {
  background: #fed7d7;
  color: #c53030;
}

.clear-history-btn:hover {
  background: #feb2b2;
}

.export-history-btn {
  background: #e2e8f0;
  color: #4a5568;
}

.export-history-btn:hover {
  background: #cbd5e0;
}

/* History Toggle Button */
.history-toggle-btn {
  position: fixed;
  top: 100px;
  right: 30px;
  padding: 12px 20px;
  background: #4299e1;
  color: white;
  border: none;
  border-radius: 25px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 10px;
  box-shadow: 0 4px 12px rgba(66, 153, 225, 0.3);
  z-index: 999;
}

.history-toggle-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(66, 153, 225, 0.4);
}

.history-toggle-btn.active {
  background: #3182ce;
}

.history-icon {
  font-size: 1.2rem;
}

/* Stats Section */
.stats-section {
  background: white;
  border-radius: 15px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  margin-bottom: 30px;
}

.stats-section h3 {
  margin: 0 0 25px 0;
  color: #4a5568;
  font-size: 1.5rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 25px;
  margin-bottom: 30px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 25px;
  background: #f8fafc;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
}

.stat-card.success {
  border-color: #38a169;
  background: linear-gradient(135deg, #f0fff4 0%, #c6f6d5 100%);
}

.stat-card.failed {
  border-color: #e53e3e;
  background: linear-gradient(135deg, #fff5f5 0%, #fed7d7 100%);
}

.stat-card.warning {
  border-color: #ed8936;
  background: linear-gradient(135deg, #fffaf0 0%, #feebc8 100%);
}

.stat-icon {
  font-size: 2rem;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 1.8rem;
  font-weight: 700;
  color: #2d3748;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 0.9rem;
  color: #718096;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.most-tested {
  margin-top: 30px;
  padding: 25px;
  background: #f8fafc;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
}

.most-tested h4 {
  margin: 0 0 15px 0;
  color: #4a5568;
  font-size: 1.2rem;
}

.targets-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.target-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: white;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  cursor: pointer;
  transition: all 0.3s;
}

.target-item:hover {
  background: #f8fafc;
  border-color: #cbd5e0;
  transform: translateX(5px);
}

.target-name {
  font-weight: 500;
  color: #4a5568;
}

.target-count {
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  color: #718096;
  background: #edf2f7;
  padding: 4px 10px;
  border-radius: 12px;
}

/* Responsive */
@media (max-width: 768px) {
  .tcp-check-container {
    padding: 10px;
  }
  
  .check-header {
    padding: 20px;
  }
  
  .check-header h1 {
    font-size: 2rem;
  }
  
  .history-sidebar {
    width: 100%;
    right: -100%;
  }
  
  .history-toggle-btn {
    right: 20px;
    padding: 10px 15px;
    font-size: 0.9rem;
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .form-actions {
    flex-direction: column;
  }
  
  .check-button, .stop-button, .clear-button, .advanced-toggle-btn {
    width: 100%;
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .quick-buttons {
    grid-template-columns: 1fr;
  }
  
  .ports-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>