<!-- src/views/Ping.vue -->
<template>
  <div class="ping-tool">
    <div class="dashboard-header">
      <h1 class="page-title">ICMP Ping Tool</h1>
      <p class="page-subtitle">Real ICMP ping using Go backend</p>
    </div>

    <div class="dashboard-content">
      <!-- Main Ping Tool Section -->
      <div class="ping-tool-section">
        <div class="tool-card">
          <div class="tool-header">
            <h2>ICMP Ping</h2>
            <p class="tool-description">Enter an IP address or domain name to perform real ICMP ping</p>
          </div>

          <!-- Input Section -->
          <div class="input-section">
            <div class="input-group">
              <div class="input-with-button">
                <input
                  v-model="pingAddress"
                  type="text"
                  placeholder="Enter IP address or domain (e.g., 8.8.8.8 or google.com)"
                  class="ping-input"
                  @keyup.enter="performPing"
                  :disabled="isPinging"
                />
                <button class="ping-button" @click="performPing" :disabled="isPinging || !pingAddress.trim()">
                  <span v-if="!isPinging">Start Ping</span>
                  <span v-else class="pinging-indicator">
                    <svg class="spinner" viewBox="0 0 50 50" width="20" height="20">
                      <circle cx="25" cy="25" r="20" fill="none" stroke="currentColor" stroke-width="5"></circle>
                    </svg>
                    Pinging...
                  </span>
                </button>
              </div>
              
              <div class="quick-options">
                <span class="quick-label">Quick targets:</span>
                <div class="quick-buttons">
                  <button
                    v-for="option in quickOptions"
                    :key="option.address"
                    class="quick-option"
                    @click="selectQuickOption(option)"
                    :title="`Ping ${option.name} (${option.address})`"
                    :disabled="isPinging"
                  >
                    {{ option.name }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Connection Status -->
          <div class="connection-status" v-if="backendStatus !== null">
            <!-- <div class="status-indicator" :class="backendStatus.connected ? 'connected' : 'disconnected'"> -->
              <!-- <span class="status-dot"></span>
              <span class="status-text">
                {{ backendStatus.connected ? 'Backend Connected' : 'Backend Disconnected' }}
              </span>
              <span v-if="!backendStatus.connected" class="retry-link" @click="checkBackendConnection">
                Retry
              </span> -->
            <!-- </div> -->
          </div>

          <!-- Ping Results -->
          <div class="results-section" v-if="pingResults.length > 0">
            <div class="results-header">
              <h3>Ping Results</h3>
          <div class="results-controls">
  <button class="control-btn" @click="clearResults" title="Clear all results">
    <svg viewBox="0 0 24 24" width="16" height="16">
      <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z" fill="currentColor"/>
    </svg>
    Clear
  </button>
  <button class="control-btn" @click="exportResults" title="Export results">
    <svg viewBox="0 0 24 24" width="16" height="16">
      <path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z" fill="currentColor"/>
    </svg>
    Export
  </button>
</div>
            </div>

            <!-- Summary Statistics -->
            <div class="summary-card" v-if="currentPingStats">
              <div class="summary-grid">
                <div class="summary-item">
                  <div class="summary-label">Target</div>
                  <div class="summary-value">{{ currentPingStats.target }}</div>
                </div>
                <div class="summary-item">
                  <div class="summary-label">Packets Sent</div>
                  <div class="summary-value">{{ currentPingStats.sent }}</div>
                </div>
                <div class="summary-item">
                  <div class="summary-label">Packets Received</div>
                  <div class="summary-value">{{ currentPingStats.received }}</div>
                </div>
                <div class="summary-item">
                  <div class="summary-label">Packet Loss</div>
                  <div class="summary-value" :class="getPacketLossClass(currentPingStats.packetLoss)">
                    {{ currentPingStats.packetLoss }}%
                  </div>
                </div>
                <div class="summary-item">
                  <div class="summary-label">Min Time</div>
                  <div class="summary-value" :class="getLatencyClass(currentPingStats.min)">
                    {{ currentPingStats.min }} ms
                  </div>
                </div>
                <div class="summary-item">
                  <div class="summary-label">Avg Time</div>
                  <div class="summary-value" :class="getLatencyClass(currentPingStats.avg)">
                    {{ currentPingStats.avg }} ms
                  </div>
                </div>
                <div class="summary-item">
                  <div class="summary-label">Max Time</div>
                  <div class="summary-value" :class="getLatencyClass(currentPingStats.max)">
                    {{ currentPingStats.max }} ms
                  </div>
                </div>
              </div>
            </div>

            <!-- Real-time Ping Output -->
            <div class="ping-output">
              <div class="output-header">
                <h4>Real-time Output</h4>
                <button class="copy-btn" @click="copyOutput" title="Copy output to clipboard">
                  <svg viewBox="0 0 24 24" width="16" height="16">
                    <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z" fill="currentColor"/>
                  </svg>
                  Copy
                </button>
              </div>
              <div class="output-content">
                <pre ref="outputContent">{{ pingOutput }}</pre>
              </div>
            </div>

            <!-- Individual Ping Results -->
            <div class="individual-results">
              <h4>Individual Ping Results</h4>
              <div class="results-table">
                <div class="table-header">
                  <div class="table-cell">Seq</div>
                  <div class="table-cell">Status</div>
                  <div class="table-cell">Latency</div>
                  <div class="table-cell">TTL</div>
                  <div class="table-cell">Time</div>
                </div>
                <div class="table-body">
                  <div v-for="(result, index) in pingResults" :key="index" class="table-row">
                    <div class="table-cell">{{ result.seq }}</div>
                    <div class="table-cell">
                      <span class="status-badge" :class="result.status">
                        {{ result.status === 'success' ? 'Success' : 'Timeout' }}
                      </span>
                    </div>
                    <div class="table-cell" :class="getLatencyClass(result.latency)">
                      {{ result.latency }} ms
                    </div>
                    <div class="table-cell">{{ result.ttl }}</div>
                    <div class="table-cell">{{ formatTime(result.timestamp) }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Advanced Options -->
          <div class="advanced-section">
            <div class="advanced-card">
              <div class="advanced-header" @click="showAdvanced = !showAdvanced">
                <h3>Advanced Options</h3>
                <svg class="toggle-icon" viewBox="0 0 24 24" width="20" height="20" :class="{ rotated: showAdvanced }">
                  <path d="M7.41 8.59L12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z" fill="currentColor"/>
                </svg>
              </div>
              
              <div class="advanced-content" v-if="showAdvanced">
                <div class="options-grid">
                  <div class="option-group">
                    <label for="packetCount">Packet Count</label>
                    <input
                      id="packetCount"
                      v-model.number="advancedOptions.packetCount"
                      type="number"
                      min="1"
                      max="100"
                      class="option-input"
                      :disabled="isPinging"
                    />
                  </div>
                  <div class="option-group">
                    <label for="packetSize">Packet Size (bytes)</label>
                    <input
                      id="packetSize"
                      v-model.number="advancedOptions.packetSize"
                      type="number"
                      min="32"
                      max="65500"
                      class="option-input"
                      :disabled="isPinging"
                    />
                  </div>
                  <div class="option-group">
                    <label for="timeout">Timeout (seconds)</label>
                    <input
                      id="timeout"
                      v-model.number="advancedOptions.timeout"
                      type="number"
                      min="1"
                      max="30"
                      step="0.5"
                      class="option-input"
                      :disabled="isPinging"
                    />
                  </div>
                  <div class="option-group">
                    <label for="interval">Interval (seconds)</label>
                    <input
                      id="interval"
                      v-model.number="advancedOptions.interval"
                      type="number"
                      min="0.1"
                      max="5"
                      step="0.1"
                      class="option-input"
                      :disabled="isPinging"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Ping History -->
        <div class="history-section" v-if="pingHistory.length > 0">
          <div class="history-card">
            <div class="history-header">
              <h3>Recent Pings</h3>
              <button class="clear-history" @click="clearHistory">
                Clear History
              </button>
            </div>
            <div class="history-list">
              <div
                v-for="(item, index) in pingHistory"
                :key="index"
                class="history-item"
                @click="selectFromHistory(item)"
              >
                <div class="history-target">
                  <span class="history-address">{{ item.address }}</span>
                  <span class="history-hostname" v-if="item.hostname">({{ item.hostname }})</span>
                </div>
                <div class="history-info">
                  <span class="history-status" :class="item.status">
                    {{ item.status === 'success' ? '✓' : '✗' }}
                  </span>
                  <span class="history-latency" v-if="item.stats">
                    {{ item.stats.avg }}ms avg
                  </span>
                  <span class="history-time">{{ formatTime(item.timestamp) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'

// State
const pingAddress = ref('')
const pingResults = ref([])
const pingOutput = ref('')
const isPinging = ref(false)
const showAdvanced = ref(false)
const pingHistory = ref([])
const backendStatus = ref(null)
const currentPingStats = ref(null)
const websocket = ref(null)

// Quick options
const quickOptions = ref([
  { name: 'Google DNS', address: '8.8.8.8' },
  { name: 'Cloudflare DNS', address: '1.1.1.1' },
  { name: 'OpenDNS', address: '208.67.222.222' },
  { name: 'Quad9 DNS', address: '9.9.9.9' },
])

// Advanced options (matching backend defaults)
const advancedOptions = ref({
  packetCount: 4,
  packetSize: 56,
  timeout: 2,
  interval: 1
})

// Computed properties
const successfulPings = computed(() => {
  return pingResults.value.filter(r => r.status === 'success').length
})

const failedPings = computed(() => {
  return pingResults.value.filter(r => r.status === 'timeout').length
})

const outputContent = ref(null)

// WebSocket setup
const setupWebSocket = () => {
  // Close existing connection if any
  if (websocket.value && websocket.value.readyState === WebSocket.OPEN) {
    websocket.value.close()
  }

  const ws = new WebSocket('ws://localhost:8082/v1/icmp')
  
  ws.onopen = () => {
    console.log('WebSocket connected')
    backendStatus.value = { connected: true }
    // showNotification('Connected to ping backend')
  }
  
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      processPingData(data)
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error)
    }
  }
  
  ws.onerror = (error) => {
    console.error('WebSocket error:', error)
    backendStatus.value = { connected: false }
    showNotification('Backend connection error', 'error')
  }
  
  ws.onclose = () => {
    console.log('WebSocket disconnected')
    backendStatus.value = { connected: false }
    isPinging.value = false
  }
  
  websocket.value = ws
  return ws
}

// Methods
const selectQuickOption = (option) => {
  pingAddress.value = option.address
  performPing()
}

const checkBackendConnection = () => {
  if (!websocket.value || websocket.value.readyState !== WebSocket.OPEN) {
    setupWebSocket()
  }
}

const performPing = async () => {
  if (!pingAddress.value.trim() || isPinging.value) return
  
  const address = pingAddress.value.trim()
  
  // Validate input
  if (!isValidHost(address)) {
    showNotification('Please enter a valid IP address or domain name', 'error')
    return
  }
  
  // Reset state
  pingResults.value = []
  pingOutput.value = ''
  currentPingStats.value = null
  isPinging.value = true
  
  // Setup WebSocket if not connected
  if (!websocket.value || websocket.value.readyState !== WebSocket.OPEN) {
    setupWebSocket()
    
    // Wait for connection with timeout
    try {
      await new Promise((resolve, reject) => {
        const maxWaitTime = 5000 // 5 seconds
        const startTime = Date.now()
        
        const checkInterval = setInterval(() => {
          if (websocket.value && websocket.value.readyState === WebSocket.OPEN) {
            clearInterval(checkInterval)
            resolve()
          } else if (Date.now() - startTime > maxWaitTime) {
            clearInterval(checkInterval)
            reject(new Error('Connection timeout'))
          }
        }, 100)
      })
    } catch (error) {
      isPinging.value = false
      showNotification('Failed to connect to backend', 'error')
      return
    }
  }
  
  try {
    const payload = {
      target: address,
      count: advancedOptions.value.packetCount,
      size: advancedOptions.value.packetSize,
      timeout: advancedOptions.value.timeout,
      interval: advancedOptions.value.interval
    }
    
    websocket.value.send(JSON.stringify(payload))
    
  } catch (error) {
    console.error('Ping failed:', error)
    pingOutput.value += `\nError: ${error.message}\n`
    showNotification('Ping failed', 'error')
    isPinging.value = false
  }
}

const processPingData = (data) => {
  switch (data.type) {
    case 'message':
      if (data.message) {
        pingOutput.value += data.message + '\n'
        
        // Scroll output to bottom
        if (outputContent.value) {
          setTimeout(() => {
            outputContent.value.scrollTop = outputContent.value.scrollHeight
          }, 0)
        }
      }
      break
      
    case 'ping_result':
      if (data.data) {
        const result = {
          seq: data.data.seq,
          status: data.data.status,
          latency: data.data.latency || 0,
          ttl: data.data.ttl || 0,
          from: data.data.from,
          timestamp: new Date()
        }
        
        pingResults.value.push(result)
      }
      
      if (data.message) {
        pingOutput.value += data.message + '\n'
        
        // Scroll output to bottom
        if (outputContent.value) {
          setTimeout(() => {
            outputContent.value.scrollTop = outputContent.value.scrollHeight
          }, 0)
        }
      }
      break
      
    case 'ping_summary':
      if (data.data) {
        currentPingStats.value = {
          target: data.data.target,
          sent: data.data.sent,
          received: data.data.received,
          packetLoss: data.data.packetLoss,
          min: data.data.min,
          avg: data.data.avg,
          max: data.data.max
        }
        
        // Add to history when we have summary
        addToHistory(pingAddress.value)
        
        // IMPORTANT: Reset pinging state after receiving summary
        isPinging.value = false
      }
      break
      
    case 'connected':
      console.log('WebSocket connected message received')
      break
      
    case 'error':
      showNotification(data.message || 'An error occurred', 'error')
      isPinging.value = false
      break
  }
}

const clearResults = () => {
  pingResults.value = []
  pingOutput.value = ''
  currentPingStats.value = null
}

const exportResults = () => {
  const data = {
    exportDate: new Date().toISOString(),
    target: pingAddress.value,
    results: pingResults.value,
    summary: currentPingStats.value,
    output: pingOutput.value
  }
  
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `ping-${pingAddress.value.replace(/[^a-z0-9]/gi, '-')}-${new Date().toISOString().split('T')[0]}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  
  showNotification('Results exported successfully')
}

const copyOutput = () => {
  const text = `Ping results for ${pingAddress.value}:\n\n${pingOutput.value}`
  
  navigator.clipboard.writeText(text)
    .then(() => showNotification('Output copied to clipboard'))
    .catch(() => showNotification('Failed to copy output', 'error'))
}

const addToHistory = (address) => {
  if (!currentPingStats.value) return
  
  const historyItem = {
    address: address,
    timestamp: new Date(),
    stats: currentPingStats.value,
    status: currentPingStats.value.packetLoss < 100 ? 'success' : 'failed'
  }
  
  pingHistory.value.unshift(historyItem)
  
  // Keep only last 20 items
  if (pingHistory.value.length > 20) {
    pingHistory.value = pingHistory.value.slice(0, 20)
  }
  
  // Save to localStorage
  try {
    localStorage.setItem('pingHistory', JSON.stringify(pingHistory.value))
  } catch (error) {
    console.error('Failed to save history:', error)
  }
}

const clearHistory = () => {
  pingHistory.value = []
  try {
    localStorage.removeItem('pingHistory')
    showNotification('History cleared')
  } catch (error) {
    console.error('Failed to clear history:', error)
  }
}

const selectFromHistory = (item) => {
  pingAddress.value = item.address
  performPing()
}

const isValidHost = (host) => {
  // Simple validation for IP address or domain
  const ipRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
  const domainRegex = /^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/
  
  // Also accept localhost
  return ipRegex.test(host) || domainRegex.test(host) || host === 'localhost'
}

const formatTime = (date) => {
  const now = new Date()
  const diff = now - new Date(date)
  const minutes = Math.floor(diff / 60000)
  
  if (minutes < 1) return 'Just now'
  if (minutes < 60) return `${minutes}m ago`
  if (minutes < 1440) return `${Math.floor(minutes / 60)}h ago`
  return `${Math.floor(minutes / 1440)}d ago`
}

const getLatencyClass = (latency) => {
  if (!latency || latency === 0) return 'offline'
  if (latency < 50) return 'good'
  if (latency < 100) return 'warning'
  return 'poor'
}

const getPacketLossClass = (loss) => {
  if (loss === 0) return 'good'
  if (loss < 10) return 'warning'
  return 'poor'
}

const showNotification = (message, type = 'success') => {
  // Remove any existing notifications
  const existingNotifications = document.querySelectorAll('.notification')
  existingNotifications.forEach(notification => {
    if (notification.parentNode) {
      notification.parentNode.removeChild(notification)
    }
  })
  
  const notification = document.createElement('div')
  notification.className = `notification ${type}`
  notification.textContent = message
  notification.style.cssText = `
    position: fixed;
    top: 20px;
    right: 20px;
    background: ${type === 'error' ? '#ef4444' : '#10b981'};
    color: white;
    padding: 12px 16px;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
    z-index: 9999;
    animation: slideIn 0.3s ease;
  `
  
  document.body.appendChild(notification)
  
  setTimeout(() => {
    notification.style.animation = 'slideOut 0.3s ease'
    setTimeout(() => {
      if (notification.parentNode) {
        document.body.removeChild(notification)
      }
    }, 300)
  }, 3000)
}

// Lifecycle
onMounted(() => {
  // Load history from localStorage
  try {
    const savedHistory = localStorage.getItem('pingHistory')
    if (savedHistory) {
      pingHistory.value = JSON.parse(savedHistory)
    }
  } catch (error) {
    console.error('Failed to load ping history:', error)
  }
  
  // Setup WebSocket connection
  setupWebSocket()
  
  // Check connection periodically
  const interval = setInterval(() => {
    if (!websocket.value || websocket.value.readyState !== WebSocket.OPEN) {
      setupWebSocket()
    }
  }, 30000)
  
  onUnmounted(() => {
    clearInterval(interval)
    
    // Close WebSocket connection
    if (websocket.value && websocket.value.readyState === WebSocket.OPEN) {
      websocket.value.close()
    }
  })
})
</script>

<style scoped>
.ping-tool {
  min-height: 100vh;
  background: #f8fafc;
}

.dashboard-header {
  background: white;
  padding: 2rem;
  border-bottom: 1px solid #e2e8f0;
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  color: #1e293b;
  margin: 0 0 0.5rem 0;
}

.page-subtitle {
  color: #64748b;
  margin: 0;
}

.dashboard-content {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

/* Tool Card */
.ping-tool-section {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.tool-card {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  border: 1px solid #e2e8f0;
}

.tool-header h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 0.5rem 0;
}

.tool-description {
  color: #64748b;
  margin: 0 0 1.5rem 0;
}

/* Input Section */
.input-section {
  margin-bottom: 1.5rem;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.input-with-button {
  display: flex;
  gap: 1rem;
}

.ping-input {
  flex: 1;
  padding: 1rem;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 1rem;
  transition: all 0.2s;
}

.ping-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.ping-input:disabled {
  background: #f8fafc;
  cursor: not-allowed;
}

.ping-button {
  padding: 1rem 2rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  min-width: 140px;
}

.ping-button:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-1px);
}

.ping-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.pinging-indicator {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Quick Options */
.quick-options {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.quick-label {
  color: #64748b;
  font-size: 0.875rem;
  font-weight: 500;
}

.quick-buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.quick-option {
  padding: 0.5rem 1rem;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  color: #475569;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-option:hover:not(:disabled) {
  background: #e2e8f0;
  transform: translateY(-1px);
}

.quick-option:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Connection Status */
.connection-status {
  margin: 1rem 0;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  font-size: 0.875rem;
}

.status-indicator.connected {
  background: #dcfce7;
  color: #16a34a;
}

.status-indicator.disconnected {
  background: #fee2e2;
  color: #dc2626;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-indicator.connected .status-dot {
  background: #16a34a;
  box-shadow: 0 0 8px rgba(22, 163, 74, 0.5);
  animation: pulse 2s infinite;
}

.status-indicator.disconnected .status-dot {
  background: #dc2626;
}

.retry-link {
  margin-left: 0.5rem;
  text-decoration: underline;
  cursor: pointer;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* Results Section */
.results-section {
  margin-top: 2rem;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.results-header h3 {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.results-controls {
  display: flex;
  gap: 0.5rem;
}

.control-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  color: #475569;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.control-btn:hover:not(:disabled) {
  background: #e2e8f0;
}

.control-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Summary Card */
.summary-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.summary-label {
  font-size: 0.75rem;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.summary-value {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1e293b;
}

.summary-value.good {
  color: #16a34a;
}

.summary-value.warning {
  color: #d97706;
}

.summary-value.poor {
  color: #dc2626;
}

/* Ping Output */
.ping-output {
  background: #1e293b;
  border-radius: 8px;
  margin-bottom: 1.5rem;
  overflow: hidden;
}

.output-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: #0f172a;
  border-bottom: 1px solid #334155;
}

.output-header h4 {
  font-size: 0.875rem;
  font-weight: 600;
  color: #cbd5e1;
  margin: 0;
}

.copy-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.25rem 0.5rem;
  background: #334155;
  border: 1px solid #475569;
  border-radius: 4px;
  color: #cbd5e1;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.copy-btn:hover {
  background: #475569;
}

.output-content {
  padding: 1rem;
  max-height: 300px;
  overflow-y: auto;
}

.output-content pre {
  margin: 0;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 0.875rem;
  line-height: 1.5;
  color: #94a3b8;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* Individual Results */
.individual-results {
  margin-top: 1.5rem;
}

.individual-results h4 {
  font-size: 1rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0 0 1rem 0;
}

.results-table {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
}

.table-header {
  display: grid;
  grid-template-columns: 80px 100px 100px 80px 1fr;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

.table-body {
  max-height: 300px;
  overflow-y: auto;
}

.table-row {
  display: grid;
  grid-template-columns: 80px 100px 100px 80px 1fr;
  border-bottom: 1px solid #f1f5f9;
}

.table-row:last-child {
  border-bottom: none;
}

.table-row:hover {
  background: #f8fafc;
}

.table-cell {
  padding: 0.75rem 1rem;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
}

.table-header .table-cell {
  font-weight: 600;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.table-body .table-cell {
  color: #475569;
}

.table-body .table-cell.good {
  color: #16a34a;
  font-weight: 600;
}

.table-body .table-cell.warning {
  color: #d97706;
  font-weight: 600;
}

.table-body .table-cell.poor {
  color: #dc2626;
  font-weight: 600;
}

.table-body .table-cell.offline {
  color: #94a3b8;
}

.status-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.status-badge.success {
  background: #dcfce7;
  color: #16a34a;
}

.status-badge.timeout {
  background: #fee2e2;
  color: #dc2626;
}

/* Advanced Section */
.advanced-section {
  margin-top: 1.5rem;
}

.advanced-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
}

.advanced-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: #f1f5f9;
  cursor: pointer;
  border-bottom: 1px solid #e2e8f0;
}

.advanced-header h3 {
  font-size: 1rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.toggle-icon {
  transition: transform 0.3s ease;
}

.toggle-icon.rotated {
  transform: rotate(180deg);
}

.advanced-content {
  padding: 1.5rem;
}

.options-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.option-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.option-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.option-input {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
}

.option-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.option-input:disabled {
  background: #f8fafc;
  cursor: not-allowed;
}

/* History Section */
.history-section {
  margin-top: 1.5rem;
}

.history-card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

.history-header h3 {
  font-size: 1rem;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}

.clear-history {
  padding: 0.25rem 0.75rem;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 4px;
  color: #64748b;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.clear-history:hover {
  background: #fee2e2;
  color: #dc2626;
  border-color: #fee2e2;
}

.history-list {
  max-height: 300px;
  overflow-y: auto;
}

.history-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1.5rem;
  border-bottom: 1px solid #f1f5f9;
  cursor: pointer;
  transition: all 0.2s;
}

.history-item:hover {
  background: #f8fafc;
}

.history-item:last-child {
  border-bottom: none;
}

.history-target {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.history-address {
  font-weight: 500;
  color: #1e293b;
}

.history-hostname {
  font-size: 0.875rem;
  color: #64748b;
}

.history-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.history-status {
  font-size: 0.875rem;
  font-weight: 600;
}

.history-status.success {
  color: #16a34a;
}

.history-status.failed {
  color: #dc2626;
}

.history-latency {
  font-size: 0.875rem;
  color: #64748b;
}

.history-time {
  font-size: 0.75rem;
  color: #94a3b8;
}

/* Animations */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes slideOut {
  from {
    opacity: 1;
    transform: translateX(0);
  }
  to {
    opacity: 0;
    transform: translateX(100%);
  }
}

/* Responsive */
@media (max-width: 768px) {
  .dashboard-content {
    padding: 1rem;
  }
  
  .tool-card {
    padding: 1.5rem;
  }
  
  .input-with-button {
    flex-direction: column;
  }
  
  .ping-button {
    width: 100%;
  }
  
  .quick-options {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .quick-buttons {
    width: 100%;
  }
  
  .quick-option {
    flex: 1;
    min-width: 0;
    text-align: center;
  }
  
  .summary-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .table-header,
  .table-row {
    grid-template-columns: 60px 80px 80px 60px 1fr;
  }
  
  .table-cell {
    padding: 0.5rem;
    font-size: 0.75rem;
  }
  
  .options-grid {
    grid-template-columns: 1fr;
  }
}
</style>