<template>
  <div class="ping-monitor">
    <!-- Header -->
    <div class="dashboard-header">
      <h1>Ping Monitor</h1>
      <p class="subtitle">Monitor host availability with real-time WebSocket updates</p>
    </div>

    <!-- Connection Status -->
    <div class="connection-bar" :class="connectionStatus">
      <span class="status-indicator"></span>
      <span class="status-text">{{ connectionMessage }}</span>
      <button v-if="connectionStatus === 'disconnected'" @click="connectWebSocket" class="reconnect-btn">
        Reconnect
      </button>
    </div>

    <!-- Add New Monitor Form -->
    <div class="add-monitor-card">
      <div class="card-header">
        <h2>Add New Monitor</h2>
      </div>
      
      <div class="add-monitor-form">
        <div class="form-row">
          <div class="form-group">
            <label for="hostname">Hostname / IP Address <span class="required">*</span></label>
            <input
              type="text"
              id="hostname"
              v-model="newMonitor.host"
              placeholder="e.g., google.com or 8.8.8.8"
              class="form-input"
              :disabled="isAdding"
              @keyup.enter="addMonitor"
            />
          </div>
          
          <div class="form-group">
            <label for="friendlyName">Friendly Name</label>
            <input
              type="text"
              id="friendlyName"
              v-model="newMonitor.name"
              placeholder="e.g., Google DNS"
              class="form-input"
              :disabled="isAdding"
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label for="interval">Check Interval (seconds)</label>
            <select
              id="interval"
              v-model="newMonitor.interval"
              class="form-select"
              :disabled="isAdding"
            >
              <option :value="30">30 seconds</option>
              <option :value="60">1 minute</option>
              <option :value="120">2 minutes</option>
              <option :value="300">5 minutes</option>
              <option :value="600">10 minutes</option>
            </select>
          </div>
          
          <div class="form-group">
            <label for="threshold">Latency Threshold (ms)</label>
            <input
              type="number"
              id="threshold"
              v-model="newMonitor.threshold"
              min="50"
              max="1000"
              step="10"
              value="200"
              class="form-input"
              :disabled="isAdding"
            />
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="form-actions">
          <button class="btn btn-primary" @click="addMonitor" :disabled="!newMonitor.host || isAdding || !isConnected">
            <span class="btn-icon">➕</span>
            {{ isAdding ? 'Adding...' : 'Add Monitor' }}
          </button>
          <button class="btn btn-secondary" @click="testConnection" :disabled="!newMonitor.host || !isConnected">
            <span class="btn-icon">🔍</span>
            Test Connection
          </button>
          <button class="btn btn-danger" @click="resetForm">
            <span class="btn-icon">↺</span>
            Reset
          </button>
        </div>
      </div>
    </div>

    <!-- Monitors Grid -->
    <div class="monitors-grid">
      <div v-for="monitor in monitors" :key="monitor.id" class="monitor-card" :class="monitor.status">
        <!-- Card Header -->
        <div class="card-header">
          <div class="header-left">
            <div class="status-indicator" :class="monitor.status"></div>
            <div class="title-section">
              <h3 class="monitor-name">{{ monitor.name || monitor.host }}</h3>
              <span class="monitor-host">{{ monitor.host }}</span>
              <span class="monitor-id">ID: {{ monitor.id }}</span>
            </div>
          </div>
          <div class="header-actions">
            <button class="icon-btn" @click="editMonitor(monitor)" title="Edit Monitor">
              ✏️
            </button>
            <button class="icon-btn delete" @click="deleteMonitor(monitor)" title="Delete Monitor">
              🗑️
            </button>
          </div>
        </div>

        <!-- Heartbeat Timeline -->
        <div class="heartbeat-timeline">
          <div 
            v-for="(beat, index) in monitor.heartbeats" 
            :key="index"
            class="heartbeat-block"
            :class="getHeartbeatClass(beat)"
            :title="getHeartbeatTitle(beat)"
          ></div>
          <div v-if="!monitor.heartbeats || monitor.heartbeats.length === 0" class="no-data">
            No data yet
          </div>
        </div>

        <!-- Stats Grid -->
        <div class="stats-grid">
          <div class="stat-item">
            <span class="stat-label">Uptime (24h)</span>
            <span class="stat-value" :class="getUptimeClass(monitor.stats.uptime)">
              {{ monitor.stats.uptime }}%
            </span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Avg Latency</span>
            <span class="stat-value" :class="getLatencyClass(monitor.stats.avgLatency)">
              {{ monitor.stats.avgLatency }}ms
            </span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Packet Loss</span>
            <span class="stat-value" :class="getPacketLossClass(monitor.stats.packetLoss)">
              {{ monitor.stats.packetLoss }}%
            </span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Last Check</span>
            <span class="stat-value">{{ formatTime(monitor.lastCheck) }}</span>
          </div>
        </div>

        <!-- Response Time -->
        <div class="response-time" :class="getLatencyClass(monitor.lastLatency)">
          {{ monitor.lastLatency }}ms
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="monitors.length === 0 && !loading" class="empty-state">
        <div class="empty-icon">📡</div>
        <h3>No Monitors Added</h3>
        <p>Add your first host to start monitoring</p>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <p>Loading monitors...</p>
      </div>
    </div>

    <!-- Test Connection Modal -->
    <div v-if="showTestModal" class="modal-overlay" @click.self="closeTestModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Connection Test Results</h3>
          <button class="close-btn" @click="closeTestModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="test-result" :class="testResult.status">
            <div class="result-icon">
              {{ testResult.status === 'success' ? '✅' : '❌' }}
            </div>
            <div class="result-details">
              <h4>{{ testResult.status === 'success' ? 'Host is reachable' : 'Host is unreachable' }}</h4>
              <p>{{ testResult.message }}</p>
              <div class="result-meta" v-if="testResult.data">
                <div class="meta-item">
                  <span class="meta-label">IP Address:</span>
                  <span class="meta-value">{{ testResult.data.ip || 'N/A' }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">Packets:</span>
                  <span class="meta-value">{{ testResult.data.packetsSent }}/{{ testResult.data.packetsReceived }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">Avg Latency:</span>
                  <span class="meta-value">{{ testResult.data.avgLatency || 0 }}ms</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">Min/Max:</span>
                  <span class="meta-value">{{ testResult.data.minLatency || 0 }}/{{ testResult.data.maxLatency || 0 }}ms</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">Packet Loss:</span>
                  <span class="meta-value">{{ testResult.data.packetLoss || 0 }}%</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="addFromTest" v-if="testResult.status === 'success'">
            Add to Monitors
          </button>
          <button class="btn btn-secondary" @click="closeTestModal">Close</button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="closeDeleteModal">
      <div class="modal-content small">
        <div class="modal-header">
          <h3>Delete Monitor</h3>
          <button class="close-btn" @click="closeDeleteModal">&times;</button>
        </div>
        <div class="modal-body">
          <p>Are you sure you want to delete monitor for <strong>{{ monitorToDelete?.host }}</strong>?</p>
          <p class="warning">This action cannot be undone.</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-danger" @click="confirmDelete">Delete</button>
          <button class="btn btn-secondary" @click="closeDeleteModal">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive, onMounted, onUnmounted } from 'vue'

// WebSocket URL - adjust based on your environment
const WS_URL = 'ws://localhost:8082/ws/ping'

export default {
  name: 'PingMonitor',
  
  setup() {
    // State
    const monitors = ref([])
    const loading = ref(true)
    const isAdding = ref(false)
    const isConnected = ref(false)
    const connectionStatus = ref('disconnected')
    const connectionMessage = ref('Disconnected')
    
    const newMonitor = reactive({
      host: '',
      name: '',
      interval: 60,
      threshold: 200
    })

    const testResult = ref({
      status: 'success',
      message: '',
      data: null
    })

    const showTestModal = ref(false)
    const showDeleteModal = ref(false)
    const monitorToDelete = ref(null)

    // WebSocket connection
    let ws = null
    let reconnectTimer = null
    const reconnectAttempts = ref(0)
    const maxReconnectAttempts = 5

    // Connect to WebSocket
    const connectWebSocket = () => {
      connectionStatus.value = 'connecting'
      connectionMessage.value = 'Connecting...'
      
      ws = new WebSocket(WS_URL)
      
      ws.onopen = () => {
        console.log('WebSocket connected')
        isConnected.value = true
        connectionStatus.value = 'connected'
        connectionMessage.value = 'Connected'
        reconnectAttempts.value = 0
        
        // Request initial data
        ws.send(JSON.stringify({
          type: 'getMonitors'
        }))
      }
      
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          handleWebSocketMessage(data)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }
      
      ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        isConnected.value = false
        connectionStatus.value = 'error'
        connectionMessage.value = 'Connection error'
      }
      
      ws.onclose = () => {
        console.log('WebSocket disconnected')
        isConnected.value = false
        connectionStatus.value = 'disconnected'
        connectionMessage.value = 'Disconnected'
        
        // Attempt to reconnect
        if (reconnectAttempts.value < maxReconnectAttempts) {
          reconnectAttempts.value++
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.value), 30000)
          connectionMessage.value = `Reconnecting in ${delay/1000}s...`
          
          if (reconnectTimer) clearTimeout(reconnectTimer)
          reconnectTimer = setTimeout(connectWebSocket, delay)
        } else {
          connectionMessage.value = 'Failed to connect'
        }
      }
    }

    // Handle WebSocket messages
    const handleWebSocketMessage = (data) => {
      switch (data.type) {
        case 'monitors':
          // Initial list of monitors
          monitors.value = data.monitors || []
          loading.value = false
          break
          
        case 'monitorAdded':
          // New monitor added
          monitors.value.push(data.monitor)
          isAdding.value = false
          resetForm()
          break
          
        case 'monitorUpdated':
          // Monitor updated
          const index = monitors.value.findIndex(m => m.id === data.monitor.id)
          if (index !== -1) {
            monitors.value[index] = data.monitor
          }
          break
          
        case 'monitorDeleted':
          // Monitor deleted
          monitors.value = monitors.value.filter(m => m.id !== data.id)
          break
          
        case 'pingUpdate':
          // Real-time ping update
          updateMonitorWithPing(data.monitorId, data.ping)
          break
          
        case 'testResult':
          // Test connection result
          testResult.value = {
            status: data.success ? 'success' : 'error',
            message: data.message,
            data: data.data
          }
          break
          
        case 'error':
          console.error('Server error:', data.message)
          alert(`Error: ${data.message}`)
          break
      }
    }

    // Update monitor with new ping data
    const updateMonitorWithPing = (monitorId, ping) => {
      const monitor = monitors.value.find(m => m.id === monitorId)
      if (!monitor) return
      
      // Update last check
      monitor.lastCheck = ping.timestamp
      monitor.lastLatency = ping.latency || 0
      
      // Update status based on latency and threshold
      if (!ping.success) {
        monitor.status = 'offline'
      } else if (ping.latency > monitor.threshold) {
        monitor.status = 'warning'
      } else {
        monitor.status = 'online'
      }
      
      // Add to heartbeats
      if (!monitor.heartbeats) monitor.heartbeats = []
      monitor.heartbeats.push({
        status: monitor.status,
        latency: ping.latency,
        timestamp: ping.timestamp
      })
      
      // Keep last 24 heartbeats
      if (monitor.heartbeats.length > 24) {
        monitor.heartbeats = monitor.heartbeats.slice(-24)
      }
      
      // Update stats
      updateMonitorStats(monitor)
    }

    // Calculate monitor statistics
    const updateMonitorStats = (monitor) => {
      const heartbeats = monitor.heartbeats || []
      const total = heartbeats.length
      
      if (total === 0) {
        monitor.stats = {
          uptime: 0,
          avgLatency: 0,
          packetLoss: 0
        }
        return
      }
      
      const online = heartbeats.filter(h => h.status === 'online').length
      const warning = heartbeats.filter(h => h.status === 'warning').length
      const offline = heartbeats.filter(h => h.status === 'offline').length
      const latencies = heartbeats.filter(h => h.latency > 0).map(h => h.latency)
      
      monitor.stats = {
        uptime: Math.round(((online + warning) / total) * 100 * 10) / 10,
        avgLatency: latencies.length ? Math.round(latencies.reduce((a, b) => a + b, 0) / latencies.length) : 0,
        packetLoss: Math.round((offline / total) * 100 * 10) / 10
      }
    }

    // Add new monitor
    const addMonitor = () => {
      if (!newMonitor.host || !isConnected.value) return
      
      isAdding.value = true
      
      ws.send(JSON.stringify({
        type: 'addMonitor',
        monitor: {
          host: newMonitor.host,
          name: newMonitor.name || newMonitor.host,
          interval: newMonitor.interval,
          threshold: newMonitor.threshold
        }
      }))
    }

    // Test connection
    const testConnection = () => {
      if (!newMonitor.host || !isConnected.value) return
      
      ws.send(JSON.stringify({
        type: 'testConnection',
        host: newMonitor.host
      }))
      
      // Modal will show when response arrives
      showTestModal.value = true
      testResult.value = {
        status: 'pending',
        message: 'Testing connection...',
        data: null
      }
    }

    // Add from test
    const addFromTest = () => {
      addMonitor()
      closeTestModal()
    }

    // Edit monitor
    const editMonitor = (monitor) => {
      newMonitor.host = monitor.host
      newMonitor.name = monitor.name === monitor.host ? '' : monitor.name
      newMonitor.interval = monitor.interval
      newMonitor.threshold = monitor.threshold
      
      // Delete old monitor
      monitorToDelete.value = monitor
      confirmDelete(true) // Skip confirmation for edit
    }

    // Delete monitor
    const deleteMonitor = (monitor) => {
      monitorToDelete.value = monitor
      showDeleteModal.value = true
    }

    // Confirm delete
    const confirmDelete = (skipConfirm = false) => {
      if (!monitorToDelete.value) return
      
      if (!skipConfirm) {
        closeDeleteModal()
      }
      
      ws.send(JSON.stringify({
        type: 'deleteMonitor',
        id: monitorToDelete.value.id
      }))
      
      monitorToDelete.value = null
    }

    // Close delete modal
    const closeDeleteModal = () => {
      showDeleteModal.value = false
      monitorToDelete.value = null
    }

    // Reset form
    const resetForm = () => {
      newMonitor.host = ''
      newMonitor.name = ''
      newMonitor.interval = 60
      newMonitor.threshold = 200
    }

    // Close test modal
    const closeTestModal = () => {
      showTestModal.value = false
    }

    // Helper methods
    const getHeartbeatClass = (beat) => {
      return beat.status || 'offline'
    }

    const getHeartbeatTitle = (beat) => {
      const time = new Date(beat.timestamp).toLocaleString()
      const latency = beat.latency ? `${beat.latency}ms` : 'No response'
      return `${time} - ${latency}`
    }

    const getUptimeClass = (uptime) => {
      if (uptime >= 99) return 'excellent'
      if (uptime >= 95) return 'good'
      if (uptime >= 90) return 'fair'
      return 'poor'
    }

    const getLatencyClass = (latency) => {
      if (latency === 0 || latency === null) return 'offline'
      if (latency < 50) return 'excellent'
      if (latency < 100) return 'good'
      if (latency < 200) return 'fair'
      return 'poor'
    }

    const getPacketLossClass = (loss) => {
      if (loss === 0) return 'excellent'
      if (loss < 1) return 'good'
      if (loss < 5) return 'fair'
      return 'poor'
    }

    const formatTime = (timestamp) => {
      if (!timestamp) return 'Never'
      const date = new Date(timestamp)
      const now = new Date()
      const diff = Math.floor((now - date) / 1000)
      
      if (diff < 60) return 'Just now'
      if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
      if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
      return date.toLocaleDateString()
    }

    // Lifecycle
    onMounted(() => {
      connectWebSocket()
    })

    onUnmounted(() => {
      if (ws) {
        ws.close()
      }
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
      }
    })

    return {
      monitors,
      loading,
      isAdding,
      isConnected,
      connectionStatus,
      connectionMessage,
      newMonitor,
      testResult,
      showTestModal,
      showDeleteModal,
      monitorToDelete,
      connectWebSocket,
      addMonitor,
      testConnection,
      addFromTest,
      editMonitor,
      deleteMonitor,
      confirmDelete,
      closeDeleteModal,
      resetForm,
      closeTestModal,
      getHeartbeatClass,
      getHeartbeatTitle,
      getUptimeClass,
      getLatencyClass,
      getPacketLossClass,
      formatTime
    }
  }
}
</script>

<style scoped>
.ping-monitor {
  min-height: 100vh;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  color: #e2e8f0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  padding: 24px;
}

/* Connection Bar */
.connection-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  margin-bottom: 24px;
  border-radius: 8px;
  font-size: 0.95rem;
}

.connection-bar.connected {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
  color: #34d399;
}

.connection-bar.connecting {
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
  color: #fbbf24;
}

.connection-bar.disconnected {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
}

.connection-bar.error {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.connected .status-indicator {
  background: #34d399;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
  animation: pulse 2s infinite;
}

.connecting .status-indicator {
  background: #fbbf24;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.reconnect-btn {
  margin-left: auto;
  padding: 6px 12px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s;
}

.reconnect-btn:hover {
  background: #dc2626;
  transform: translateY(-1px);
}

/* Dashboard Header */
.dashboard-header {
  margin-bottom: 24px;
}

.dashboard-header h1 {
  font-size: 2rem;
  font-weight: 700;
  margin: 0 0 8px 0;
  background: linear-gradient(135deg, #60a5fa 0%, #a78bfa 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.subtitle {
  color: #94a3b8;
  font-size: 1rem;
  margin: 0;
}

/* Add Monitor Card */
.add-monitor-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  margin-bottom: 32px;
  overflow: hidden;
}

.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  background: rgba(15, 23, 42, 0.4);
}

.card-header h2 {
  margin: 0;
  color: #f8fafc;
  font-size: 1.25rem;
  font-weight: 600;
}

.add-monitor-form {
  padding: 24px;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  color: #cbd5e1;
  font-size: 0.9rem;
  font-weight: 500;
}

.required {
  color: #ef4444;
  margin-left: 4px;
}

.form-input,
.form-select {
  padding: 10px 14px;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  color: #e2e8f0;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

.form-input::placeholder {
  color: #64748b;
}

.form-input:disabled,
.form-select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Form Actions */
.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  flex-wrap: wrap;
}

.btn {
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: #1e293b;
  color: #cbd5e1;
  border: 1px solid #334155;
}

.btn-secondary:hover:not(:disabled) {
  background: #2d3748;
  transform: translateY(-1px);
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover {
  background: #dc2626;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.btn-icon {
  font-size: 1.1rem;
}

/* Monitors Grid */
.monitors-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.monitor-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  padding: 20px;
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: relative;
}

.monitor-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.4);
}

.monitor-card.online {
  border-left: 4px solid #10b981;
}

.monitor-card.warning {
  border-left: 4px solid #f59e0b;
}

.monitor-card.offline {
  border-left: 4px solid #ef4444;
}

/* Card Header */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 0;
  border: none;
  background: transparent;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-section {
  display: flex;
  flex-direction: column;
}

.monitor-name {
  margin: 0;
  color: #f8fafc;
  font-size: 1.1rem;
  font-weight: 600;
}

.monitor-host {
  color: #94a3b8;
  font-size: 0.85rem;
}

.monitor-id {
  color: #64748b;
  font-size: 0.7rem;
  margin-top: 2px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.icon-btn {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 6px;
  border-radius: 6px;
  font-size: 1rem;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: #1e293b;
  color: #60a5fa;
}

.icon-btn.delete:hover {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
}

/* Heartbeat Timeline */
.heartbeat-timeline {
  display: flex;
  gap: 4px;
  height: 40px;
  align-items: center;
  background: #0f172a;
  border-radius: 8px;
  padding: 4px;
}

.heartbeat-block {
  flex: 1;
  height: 30px;
  border-radius: 4px;
  transition: all 0.2s;
  cursor: default;
}

.heartbeat-block:hover {
  transform: scaleY(1.2);
}

.heartbeat-block.online {
  background: #10b981;
}

.heartbeat-block.warning {
  background: #f59e0b;
}

.heartbeat-block.offline {
  background: #ef4444;
}

.no-data {
  flex: 1;
  text-align: center;
  color: #64748b;
  font-size: 0.85rem;
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 12px;
  background: #0f172a;
  border-radius: 8px;
  border: 1px solid #334155;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  font-size: 0.7rem;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 1rem;
  font-weight: 600;
}

.stat-value.excellent {
  color: #34d399;
}

.stat-value.good {
  color: #60a5fa;
}

.stat-value.fair {
  color: #fbbf24;
}

.stat-value.poor {
  color: #f87171;
}

/* Response Time */
.response-time {
  position: absolute;
  top: 20px;
  right: 70px;
  padding: 4px 10px;
  background: #1e293b;
  border-radius: 20px;
  font-weight: 600;
  font-size: 0.9rem;
  border: 1px solid #334155;
}

.response-time.excellent {
  color: #34d399;
}

.response-time.good {
  color: #60a5fa;
}

.response-time.fair {
  color: #fbbf24;
}

.response-time.poor {
  color: #f87171;
}

/* Empty State */
.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 80px 20px;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px dashed rgba(148, 163, 184, 0.2);
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 20px;
  opacity: 0.5;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.empty-state h3 {
  color: #f8fafc;
  margin-bottom: 8px;
}

.empty-state p {
  color: #94a3b8;
}

/* Loading State */
.loading-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #1e293b;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 16px auto;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-state p {
  color: #94a3b8;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #1e293b;
  border-radius: 16px;
  width: 500px;
  max-width: 90vw;
  max-height: 90vh;
  overflow-y: auto;
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5);
}

.modal-content.small {
  width: 400px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #334155;
}

.modal-header h3 {
  margin: 0;
  color: #f8fafc;
  font-size: 1.2rem;
}

.close-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 24px;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #ef4444;
  color: white;
}

.modal-body {
  padding: 24px;
}

.modal-body .warning {
  color: #f87171;
  font-size: 0.9rem;
  margin-top: 8px;
}

.test-result {
  display: flex;
  gap: 16px;
  padding: 16px;
  border-radius: 12px;
}

.test-result.success {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.test-result.error {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.test-result.pending {
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.result-icon {
  font-size: 32px;
}

.result-details h4 {
  margin: 0 0 8px 0;
  color: #f8fafc;
  font-size: 1rem;
}

.result-details p {
  margin: 0 0 12px 0;
  color: #94a3b8;
  font-size: 0.9rem;
}

.result-meta {
  background: #0f172a;
  border-radius: 8px;
  padding: 12px;
  border: 1px solid #334155;
}

.meta-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
  font-size: 0.9rem;
}

.meta-item:last-child {
  margin-bottom: 0;
}

.meta-label {
  color: #64748b;
}

.meta-value {
  color: #e2e8f0;
  font-weight: 500;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 24px 24px;
  border-top: 1px solid #334155;
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
  .ping-monitor {
    padding: 16px;
  }
  
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .form-actions {
    flex-direction: column;
  }
  
  .btn {
    width: 100%;
    justify-content: center;
  }
  
  .monitors-grid {
    grid-template-columns: 1fr;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .response-time {
    position: static;
    margin-top: 8px;
  }
  
  .test-result {
    flex-direction: column;
    text-align: center;
  }
}
</style>