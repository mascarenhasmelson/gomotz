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
import { ref, computed, onMounted, onUnmounted } from 'vue';
// const API_URL = import.meta.env.VITE_API_URL;
const API_URL = "http://192.168.20.17:8082";
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
      isWaitingForResponse: true, // New flag to track if we're waiting for response
      
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
      this.isWaitingForResponse = true // Reset waiting flag
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
      
      // Start progress simulation based on timeout
      const timeoutValue = this.options.customTimeout ? this.options.timeoutValue * 1000 : 5000
      this.startProgressSimulation(timeoutValue)
      
      try {
        const response = await fetch(`${API_URL}/v1/tcpCheck`, {
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
        
        // If we get a response (open port), immediately update progress and show results
        if (data && data.status === 'open') {
          this.isWaitingForResponse = false
          this.progress = 100
          this.currentStep = 4
        }
        
        this.processResult(data)
        
      } catch (err) {
        console.error('TCP check error:', err)
        // Only handle as error if we haven't already processed a result
        if (this.isWaitingForResponse) {
          this.isWaitingForResponse = false
          this.handleError(err)
        }
      } finally {
        this.isChecking = false
        this.cleanup()
      }
    },
    
    startProgressSimulation(timeoutMs) {
      const startTime = Date.now()
      const timeoutSeconds = timeoutMs / 1000
      
      this.progressInterval = setInterval(() => {
        const elapsedMs = Date.now() - this.testStartTime
        this.elapsedTime = (elapsedMs / 1000).toFixed(1)
        
        // If we're still waiting for response, continue progress simulation
        if (this.isWaitingForResponse) {
          // Calculate progress based on elapsed time vs timeout
          const progressPercent = Math.min((elapsedMs / timeoutMs) * 100, 95)
          this.progress = Math.round(progressPercent)
          
          // Update steps based on progress
          if (elapsedMs < 500) {
            this.currentStep = 1 // DNS Lookup
          } else if (elapsedMs < 1000) {
            this.currentStep = 2 // Sending SYN
          } else if (elapsedMs < 2000) {
            this.currentStep = 3 // Waiting Response
          } else {
            this.currentStep = 3 // Still waiting
          }
          
          // If we've reached timeout, show that we're waiting for response
          if (elapsedMs >= timeoutMs) {
            this.currentStep = 3 // Still in waiting response
          }
        }
      }, 100)
    },
    
    processResult(data) {
      const duration = Date.now() - this.testStartTime
      let status = 'closed'
      let title = 'Port Closed'
      let message = 'No response received (connection refused or timeout)'
      let responseTime = duration
      
      if (data && data.status === 'open') {
        status = 'open'
        title = 'Port Open'
        message = 'Response received - Port is open and accepting connections'
        responseTime = data.responseTime || duration
      } else if (data && data.status === 'filtered') {
        status = 'filtered'
        title = 'Port Filtered'
        message = 'Port may be filtered by firewall (no response)'
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
          ttl: '64',
          timeout: this.options.customTimeout ? this.options.timeoutValue : 5
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
            message: 'An open port may be a security risk. Ensure this service is intended to be publicly accessible and properly secured.'
          }
        case 'closed':
          return {
            title: 'Standard Security',
            message: 'Closed ports are secure by default. No action required.'
          }
        case 'filtered':
          return {
            title: 'Firewall Detected',
            message: 'A firewall appears to be blocking connections. Check your firewall rules if this port should be accessible.'
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
        details: {
          error: err.message,
          timeout: this.options.customTimeout ? this.options.timeoutValue : 5
        },
        banner: '',
        recommendation: {
          title: 'Troubleshooting',
          message: 'Check your network connection, firewall settings, and ensure the host is reachable.'
        }
      }
      
      this.showResults = true
      this.progress = 100
    },
    
    stopTest() {
      this.isChecking = false
      this.isWaitingForResponse = false
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
/* Dark Mode Theme */
.tcp-check-page {
  padding: 24px;
  width: 100%;
  min-height: 100vh;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  color: #e2e8f0;
}

.tcp-check-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

/* Form */
.check-form {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  margin-bottom: 30px;
  border: 1px solid rgba(148, 163, 184, 0.1);
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
  color: #cbd5e1;
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
  border: 1px solid #334155;
  border-radius: 10px;
  font-size: 1rem;
  font-family: 'Monaco', 'Courier New', monospace;
  transition: all 0.3s;
  background: #0f172a;
  color: #e2e8f0;
}

.host-input:focus, .port-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

.host-input::placeholder, .port-input::placeholder {
  color: #64748b;
}

.host-input:disabled, .port-input:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  background: #1e293b;
}

.suggestions {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 8px;
  margin-top: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  z-index: 1000;
  overflow: hidden;
}

.suggestion-item {
  padding: 12px 20px;
  cursor: pointer;
  transition: all 0.2s;
  color: #cbd5e1;
  border-bottom: 1px solid #334155;
}

.suggestion-item:hover {
  background: #2d3748;
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
  color: #94a3b8;
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
  border: 1px solid #334155;
  border-radius: 6px;
  background: #0f172a;
  color: #cbd5e1;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.port-btn:hover {
  border-color: #3b82f6;
  transform: translateY(-2px);
  background: #1e293b;
}

.port-btn.active {
  background: #3b82f6;
  border-color: #3b82f6;
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
  background: #1e293b;
  border-radius: 10px;
  border: 1px solid #334155;
}

.quick-tests h4 {
  margin: 0 0 15px 0;
  color: #cbd5e1;
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
  border: 1px solid #334155;
  border-radius: 8px;
  background: #0f172a;
  color: #cbd5e1;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 10px;
}

.quick-btn:hover {
  border-color: #3b82f6;
  transform: translateY(-2px);
  background: #1e293b;
}

.quick-btn.active {
  background: #3b82f6;
  border-color: #3b82f6;
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
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  flex: 1;
}

.check-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(59, 130, 246, 0.3);
}

.check-button:disabled {
  opacity: 0.7;
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

.clear-button {
  background: #1e293b;
  color: #cbd5e1;
  border: 1px solid #334155;
}

.clear-button:hover:not(:disabled) {
  background: #2d3748;
  transform: translateY(-2px);
}

.advanced-toggle-btn {
  background: #0f172a;
  color: #cbd5e1;
  border: 1px solid #334155;
}

.advanced-toggle-btn:hover {
  border-color: #3b82f6;
  transform: translateY(-2px);
}

.button-icon {
  font-size: 1.1rem;
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

/* Advanced Options */
.advanced-options {
  margin: 20px 0;
  padding: 25px;
  background: #1e293b;
  border-radius: 10px;
  border: 1px solid #334155;
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
  color: #cbd5e1;
  cursor: pointer;
}

.option-group input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: #3b82f6;
}

.option-control {
  display: flex;
  align-items: center;
  gap: 15px;
}

.timeout-slider, .retry-slider {
  flex: 1;
  height: 6px;
  background: #334155;
  border-radius: 3px;
  outline: none;
  -webkit-appearance: none;
}

.timeout-slider::-webkit-slider-thumb,
.retry-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 20px;
  height: 20px;
  background: #3b82f6;
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid white;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
}

.slider-value {
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  color: #94a3b8;
  min-width: 80px;
}

/* Error Message */
.error-message {
  margin-top: 20px;
  padding: 15px;
  background: rgba(239, 68, 68, 0.1);
  color: #fca5a5;
  border-radius: 8px;
  border-left: 4px solid #ef4444;
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
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  margin-bottom: 30px;
  border: 1px solid rgba(148, 163, 184, 0.1);
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
  color: #cbd5e1;
  font-size: 1.1rem;
}

.status-icon {
  font-size: 1.3rem;
}

.progress-time {
  font-family: 'Monaco', 'Courier New', monospace;
  color: #94a3b8;
  font-size: 0.9rem;
}

.progress-container {
  height: 10px;
  background: #1e293b;
  border-radius: 5px;
  overflow: hidden;
  margin-bottom: 30px;
  border: 1px solid #334155;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #8b5cf6);
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
  background: #334155;
  transition: all 0.3s;
}

.step.active .step-dot {
  background: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

.step-label {
  font-size: 0.85rem;
  color: #94a3b8;
  text-align: center;
  font-weight: 500;
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
  animation: slideInUp 0.5s ease-out;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 25px;
  padding-bottom: 20px;
  border-bottom: 1px solid #334155;
}

.result-title h2 {
  margin: 0;
  color: #f8fafc;
  font-size: 1.8rem;
}

.result-target {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 10px;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 1.1rem;
  color: #cbd5e1;
}

.target-icon {
  font-size: 1.2rem;
}

.target-port {
  font-weight: 700;
  color: #60a5fa;
}

.result-actions {
  display: flex;
  gap: 10px;
}

.action-btn {
  padding: 10px 20px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 8px;
  color: #cbd5e1;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn:hover {
  background: #2d3748;
  border-color: #3b82f6;
  transform: translateY(-2px);
}

.action-icon {
  font-size: 1.1rem;
}

/* Result Card */
.result-card {
  border-radius: 12px;
  overflow: hidden;
  border: 2px solid transparent;
  transition: all 0.3s;
}

.result-card.open {
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.1);
}

.result-card.closed {
  border-color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

.result-card.filtered {
  border-color: #f59e0b;
  background: rgba(245, 158, 11, 0.1);
}

.result-card.error {
  border-color: #6b7280;
  background: rgba(107, 114, 128, 0.1);
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
  color: #f8fafc;
  font-size: 1.5rem;
}

.status-message {
  margin: 0 0 15px 0;
  color: #cbd5e1;
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
  color: #94a3b8;
  font-weight: 500;
}

.detail-value {
  font-family: 'Monaco', 'Courier New', monospace;
  font-weight: 600;
  color: #f8fafc;
  font-size: 1.1rem;
}

/* Technical Details */
.technical-details {
  border-top: 1px solid #334155;
  background: rgba(15, 23, 42, 0.4);
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
  background: rgba(59, 130, 246, 0.1);
}

.details-title {
  font-weight: 600;
  color: #cbd5e1;
  font-size: 1.1rem;
}

.toggle-icon {
  font-size: 1.2rem;
  color: #94a3b8;
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
  background: #0f172a;
  border-radius: 8px;
  border: 1px solid #334155;
}

.banner-section {
  margin-top: 25px;
}

.banner-section h4 {
  margin: 0 0 15px 0;
  color: #cbd5e1;
  font-size: 1.1rem;
}

.banner-content {
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 20px;
  max-height: 200px;
  overflow-y: auto;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  line-height: 1.4;
  color: #94a3b8;
}

/* Recommendation */
.recommendation {
  margin-top: 25px;
  padding: 25px;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 10px;
  border: 1px solid rgba(59, 130, 246, 0.3);
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
  color: #60a5fa;
  margin-bottom: 8px;
  font-size: 1.1rem;
}

.recommendation-content p {
  margin: 0;
  color: #94a3b8;
  line-height: 1.5;
}

/* History Sidebar */
.history-sidebar {
  position: fixed;
  top: 0;
  right: -450px;
  width: 450px;
  height: 100vh;
  background: #1e293b;
  box-shadow: -10px 0 30px rgba(0, 0, 0, 0.3);
  transition: right 0.3s ease;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  border-left: 1px solid #334155;
}

.history-sidebar.visible {
  right: 0;
}

.history-header {
  padding: 25px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #334155;
}

.history-header h3 {
  margin: 0;
  color: #f8fafc;
  font-size: 1.5rem;
}

.close-history {
  width: 40px;
  height: 40px;
  border: none;
  background: #0f172a;
  border-radius: 50%;
  font-size: 1.5rem;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  border: 1px solid #334155;
}

.close-history:hover {
  background: #ef4444;
  color: white;
}

.history-stats {
  padding: 20px 25px;
  display: flex;
  gap: 20px;
  background: #0f172a;
  border-bottom: 1px solid #334155;
}

.history-stat {
  flex: 1;
  text-align: center;
}

.history-stat .stat-value {
  display: block;
  font-size: 1.5rem;
  font-weight: 700;
  color: #60a5fa;
  margin-bottom: 5px;
}

.history-stat .stat-label {
  font-size: 0.8rem;
  color: #94a3b8;
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
  border-bottom: 1px solid #334155;
  cursor: pointer;
  transition: all 0.3s;
}

.history-item:hover {
  background: #2d3748;
}

.history-status {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.history-status.open {
  background: #10b981;
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.5);
}

.history-status.closed {
  background: #ef4444;
  box-shadow: 0 0 10px rgba(239, 68, 68, 0.5);
}

.history-status.filtered {
  background: #f59e0b;
  box-shadow: 0 0 10px rgba(245, 158, 11, 0.5);
}

.history-status.error {
  background: #6b7280;
  box-shadow: 0 0 10px rgba(107, 114, 128, 0.5);
}

.history-info {
  flex: 1;
  min-width: 0;
}

.history-target {
  font-family: 'Monaco', 'Courier New', monospace;
  font-weight: 600;
  color: #f8fafc;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.history-time {
  font-size: 0.8rem;
  color: #94a3b8;
}

.history-response {
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  color: #60a5fa;
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
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
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
  background: #ef4444;
  color: white;
}

.empty-history {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #64748b;
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
  border-top: 1px solid #334155;
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
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.clear-history-btn:hover {
  background: #ef4444;
  color: white;
}

.export-history-btn {
  background: #0f172a;
  color: #cbd5e1;
  border: 1px solid #334155;
}

.export-history-btn:hover {
  background: #1e293b;
  border-color: #3b82f6;
  color: #60a5fa;
}

/* History Toggle Button */
.history-toggle-btn {
  position: fixed;
  top: 100px;
  right: 30px;
  padding: 12px 20px;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  border: none;
  border-radius: 25px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 10px;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  z-index: 999;
}

.history-toggle-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(59, 130, 246, 0.4);
}

.history-toggle-btn.active {
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
}

.history-icon {
  font-size: 1.2rem;
}

/* Stats Section */
.stats-section {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  margin-bottom: 30px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.stats-section h3 {
  margin: 0 0 25px 0;
  color: #f8fafc;
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
  background: #0f172a;
  border-radius: 12px;
  border: 1px solid #334155;
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-5px);
  border-color: #3b82f6;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
}

.stat-card.success {
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.1);
}

.stat-card.failed {
  border-color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

.stat-card.warning {
  border-color: #f59e0b;
  background: rgba(245, 158, 11, 0.1);
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
  color: #f8fafc;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 0.9rem;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.most-tested {
  margin-top: 30px;
  padding: 25px;
  background: #0f172a;
  border-radius: 10px;
  border: 1px solid #334155;
}

.most-tested h4 {
  margin: 0 0 15px 0;
  color: #cbd5e1;
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
  background: #1e293b;
  border-radius: 8px;
  border: 1px solid #334155;
  cursor: pointer;
  transition: all 0.3s;
}

.target-item:hover {
  background: #2d3748;
  border-color: #3b82f6;
  transform: translateX(5px);
}

.target-name {
  font-weight: 500;
  color: #cbd5e1;
}

.target-count {
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  color: #60a5fa;
  background: #0f172a;
  padding: 4px 10px;
  border-radius: 12px;
  border: 1px solid #334155;
}

/* Responsive */
@media (max-width: 768px) {
  .tcp-check-container {
    padding: 10px;
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
</style>