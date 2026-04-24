<template>
  <div class="tcp-monitor">
    <!-- Monitor Configuration Form -->
    <div class="config-card">
      <div class="card-header">
        <h2>{{ editingMonitorId ? 'Edit Monitor' : 'Monitor Configuration' }}</h2>
        <button v-if="editingMonitorId" class="btn btn-secondary" @click="cancelEdit">
          Cancel Edit
        </button>
      </div>

      <div class="config-form">
        <!-- Friendly Name -->
        <div class="form-row">
          <div class="form-group">
            <label>Friendly Name</label>
            <input 
              type="text" 
              v-model="config.friendly_name" 
              placeholder="e.g., Production Web Server"
              class="form-input"
            />
          </div>
        </div>

        <!-- Hostname and Port -->
        <div class="form-row dual">
          <div class="form-group">
            <label>Hostname</label>
            <input 
              type="text" 
              v-model="config.hostname" 
              placeholder="e.g., example.com or 192.168.1.1"
              class="form-input"
            />
          </div>
          <div class="form-group">
            <label>Port</label>
            <input 
              type="number" 
              v-model="config.port" 
              placeholder="80"
              min="1"
              max="65535"
              class="form-input"
            />
          </div>
        </div>

        <!-- Heartbeat Interval -->
        <div class="form-row">
          <div class="form-group heartbeat-group">
            <label>Heartbeat Interval</label>
            <div class="heartbeat-interval-container">
              <div class="heartbeat-visualization">
                <div class="heartbeat-bar">
                  <div 
                    class="heartbeat-fill" 
                    :style="{ width: (config.heartbeat_interval / 300) * 100 + '%' }"
                  ></div>
                </div>
                <div class="heartbeat-markers">
                  <span class="marker">1m</span>
                  <span class="marker">5m</span>
                  <span class="marker">10m</span>
                  <span class="marker">30m</span>
                  <span class="marker">1h</span>
                </div>
              </div>

              <div class="heartbeat-input-group">
                <div class="heartbeat-number-input">
                  <input 
                    type="number" 
                    v-model="config.heartbeat_interval" 
                    min="1"
                    max="3600"
                    class="heartbeat-number-field"
                  />
                  <span class="heartbeat-unit">seconds</span>
                </div>
                <div class="heartbeat-display">
                  <span class="heartbeat-value">{{ formatHeartbeatTime(config.heartbeat_interval) }}</span>
                  <span class="heartbeat-badge" :class="getHeartbeatCategory(config.heartbeat_interval)">
                    {{ getHeartbeatCategory(config.heartbeat_interval) }}
                  </span>
                </div>
              </div>

              <div class="heartbeat-description">
                <span class="description-icon">⏱️</span>
                <span class="description-text">
                  Check every <strong>{{ config.heartbeat_interval }} seconds</strong> 
                  ({{ formatHeartbeatTime(config.heartbeat_interval) }})
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Retries -->
        <div class="form-row">
          <div class="form-group">
            <label>Retries</label>
            <div class="retries-container">
              <input 
                type="number" 
                v-model="config.retries" 
                min="0"
                max="10"
                class="form-input retries-input"
              />
              <div class="retries-visualization">
                <div 
                  v-for="n in 5" 
                  :key="n"
                  class="retry-dot"
                  :class="{ active: n <= config.retries }"
                ></div>
              </div>
              <span class="input-hint">Maximum retries before marking as down</span>
            </div>
          </div>
        </div>

        <!-- Heartbeat Retry Interval -->
        <div class="form-row">
          <div class="form-group">
            <label>Heartbeat Retry Interval</label>
            <div class="interval-input">
              <input 
                type="number" 
                v-model="config.heartbeat_retry_interval" 
                min="1"
                class="form-input interval-field"
              />
              <span class="interval-unit">seconds</span>
            </div>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="form-actions">
          <button class="btn btn-primary" @click="saveMonitor" :disabled="isLoading">
            <span class="btn-icon">💾</span>
            {{ isLoading ? 'Saving...' : (editingMonitorId ? 'Update Monitor' : 'Save Monitor') }}
          </button>
          <button class="btn btn-secondary" @click="testConnection" :disabled="!config.hostname || !config.port">
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

    <!-- WebSocket Connection Status -->
    <div class="ws-status" :class="wsConnected ? 'connected' : 'disconnected'">
      <span class="ws-indicator"></span>
      <span class="ws-text">{{ wsConnected ? 'Real-time updates active' : 'Connecting to real-time updates...' }}</span>
    </div>

    <!-- Active Monitors List -->
    <div class="monitors-list-card">
      <div class="card-header">
        <div class="header-left">
          <h2>Active Monitors</h2>
          <div class="sort-controls">
            <button 
              class="sort-btn" 
              :class="{ active: sortBy === 'status' }"
              @click="toggleSort('status')"
            >
              <span>Status</span>
              <span class="sort-icon">{{ getSortIcon('status') }}</span>
            </button>
            <button 
              class="sort-btn" 
              :class="{ active: sortBy === 'friendly_name' }"
              @click="toggleSort('friendly_name')"
            >
              <span>Name</span>
              <span class="sort-icon">{{ getSortIcon('friendly_name') }}</span>
            </button>
            <button 
              class="sort-btn" 
              :class="{ active: sortBy === 'hostname' }"
              @click="toggleSort('hostname')"
            >
              <span>Host</span>
              <span class="sort-icon">{{ getSortIcon('hostname') }}</span>
            </button>
          </div>
        </div>
        <div class="header-actions">
          <span class="monitor-count">
            <span class="up-count">🟢 Up: {{ upCount }}</span>
            <span class="down-count">🔴 Down: {{ downCount }}</span>
            <span class="total-count">Total: {{ monitors.length }}</span>
          </span>
          <button class="btn btn-icon-only" @click="fetchMonitors" :disabled="isLoading" title="Refresh">
            <span>↻</span>
          </button>
        </div>
      </div>

      <div class="monitors-table-container">
        <div class="monitors-table">
          <table>
            <thead>
              <tr>
                <th @click="toggleSort('status')" class="sortable-header">
                  Status
                  <span class="sort-icon">{{ getSortIcon('status') }}</span>
                </th>
                <th @click="toggleSort('friendly_name')" class="sortable-header">
                  Friendly Name
                  <span class="sort-icon">{{ getSortIcon('friendly_name') }}</span>
                </th>
                <th @click="toggleSort('hostname')" class="sortable-header">
                  Host:Port
                  <span class="sort-icon">{{ getSortIcon('hostname') }}</span>
                </th>
                <th>Heartbeat</th>
                <th>Retries</th>
                <th>Last Check</th>
                <th>Response Time</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="monitor in sortedMonitors" :key="monitor.id" class="monitor-row">
                <td data-label="Status">
                  <span class="status-badge" :class="monitor.status">
                    {{ monitor.status === 'up' ? '🟢 UP' : '🔴 down' }}
                  </span>
                </td>
                <td data-label="Friendly Name" class="friendly-name">{{ monitor.friendly_name }}</td>
                <td data-label="Host:Port" class="host-port">{{ monitor.hostname }}:{{ monitor.port }}</td>
                <td data-label="Heartbeat" class="heartbeat-cell">
                  <div class="heartbeat-info">
                    <span class="heartbeat-interval-badge">{{ monitor.heartbeat_interval }}s</span>
                    <span class="heartbeat-interval-human">{{ formatHeartbeatTime(monitor.heartbeat_interval) }}</span>
                  </div>
                  <div class="heartbeat-timeline">
                    <div 
                      v-for="i in 12" 
                      :key="i"
                      class="heartbeat-block"
                      :class="getHeartbeatBlockClass(monitor, i)"
                    ></div>
                  </div>
                </td>
                <td data-label="Retries">
                  <div class="retries-badge">
                    <span class="retries-count">{{ monitor.retries }}</span>
                    <span class="retries-max">/{{ monitor.retries || 3 }}</span>
                  </div>
                </td>
                <td data-label="Last Check" class="last-check-cell">
                  <span class="last-check-time">{{ formatLastCheck(monitor.last_check) }}</span>
                  <span class="last-check-timestamp">{{ formatTimestamp(monitor.last_check) }}</span>
                  <div class="status-indicator" :class="monitor.status"></div>
                </td>
                <td data-label="Response Time" class="response-time-cell">
                  <span v-if="monitor.last_response_ms" class="response-time" :class="getResponseTimeClass(monitor.last_response_ms)">
                    {{ monitor.last_response_ms }}ms
                  </span>
                  <span v-else class="response-time na">N/A</span>
                </td>
                <td data-label="Actions" class="actions">
                  <!-- <button class="action-btn" @click="editMonitor(monitor)" title="Edit">
                    ✏️
                  </button> -->
                  <button class="action-btn" @click="deleteMonitor(monitor)" title="Delete">
                    🗑️
                  </button>
                </td>
              </tr>
              <tr v-if="monitors.length === 0 && !isLoading">
                <td colspan="8" class="empty-state">
                  <div class="empty-icon">   </div>
                  <p>No monitors configured yet</p>
                  <p class="empty-hint">Create your first monitor using the form above</p>
                </td>
              </tr>
              <tr v-if="isLoading && monitors.length === 0">
                <td colspan="8" class="loading-state">
                  <div class="loading-spinner"></div>
                  <p>Loading monitors...</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
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
              <h4>{{ testResult.status === 'success' ? 'Connection Successful' : 'Connection Failed' }}</h4>
              <p>{{ testResult.message }}</p>
              <div class="result-meta" v-if="testResult.details">
                <div class="meta-item">
                  <span class="meta-label">Host:</span>
                  <span class="meta-value">{{ testResult.details.host }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">Port:</span>
                  <span class="meta-value">{{ testResult.details.port }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">Response Time:</span>
                  <span class="meta-value">{{ testResult.details.responseTime }}ms</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="closeTestModal">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8082'
const WS_BASE_URL = API_BASE_URL.replace('http', 'ws').replace('https', 'wss')

export default {
  name: 'TCPMonitor',
  
  setup() {
    // State
    const monitors = ref([])
    const isLoading = ref(false)
    const editingMonitorId = ref(null)
    const ws = ref(null)
    const wsConnected = ref(false)
    
    // Configuration state
    const config = reactive({
      friendly_name: '',
      hostname: '',
      port: 80,
      heartbeat_interval: 60,
      retries: 2,
      heartbeat_retry_interval: 5
    })

    // Sorting state
    const sortBy = ref('status')
    const sortDirection = ref('desc')

    // Test modal state
    const showTestModal = ref(false)
    const testResult = ref({
      status: 'success',
      message: '',
      details: {
        host: '',
        port: 80,
        responseTime: 0
      }
    })

    // Computed properties
    const upCount = computed(() => {
      return monitors.value.filter(m => m.status === 'up').length
    })

    const downCount = computed(() => {
      return monitors.value.filter(m => m.status === 'down').length
    })

    const sortedMonitors = computed(() => {
      const sorted = [...monitors.value]
      
      sorted.sort((a, b) => {
        let comparison = 0
        
        switch (sortBy.value) {
          case 'status':
            comparison = (a.status === 'up' ? -1 : 1) - (b.status === 'up' ? -1 : 1)
            break
          case 'friendly_name':
            comparison = (a.friendly_name || '').localeCompare(b.friendly_name || '')
            break
          case 'hostname':
            comparison = (a.hostname || '').localeCompare(b.hostname || '')
            break
          default:
            return 0
        }
        
        return sortDirection.value === 'asc' ? comparison : -comparison
      })
      
      return sorted
    })

    // WebSocket Connection
    const connectWebSocket = () => {
      const wsUrl = `${WS_BASE_URL}/v1/ws/monitors`
      console.log('Connecting to WebSocket:', wsUrl)
      
      ws.value = new WebSocket(wsUrl)

      ws.value.onopen = () => {
        console.log('  Monitor WebSocket connected')
        wsConnected.value = true
      }

      ws.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('📨 WebSocket message:', data.type)
          
          if (data.type === 'initial_state') {
            monitors.value = (data.monitors || []).map(m => ({
              ...m,
              status: m.status === 'up' ? 'up' : 'down',
              last_check: m.last_checked_at || null,
              last_response_ms: m.last_response_ms || null
            }))
            console.log(`Loaded ${monitors.value.length} monitors via WebSocket`)
          } else if (data.type === 'port_monitor_update') {
            const idx = monitors.value.findIndex(m => m.id === data.monitor_id)
            if (idx !== -1) {
              monitors.value[idx] = {
                ...monitors.value[idx],
                status: data.status === 'up' ? 'up' : 'down',
                last_tcp_status: data.tcp_status,
                last_response_ms: data.response_ms,
                last_check: data.checked_at
              }
              console.log(`Updated monitor ${data.monitor_id}: ${data.status}`)
            }
          } else if (data.type === 'monitor_created') {
            monitors.value.push({
              ...data.monitor,
              status: data.monitor.status === 'up' ? 'up' : 'down',
              last_check: data.monitor.last_checked_at || null,
              last_response_ms: data.monitor.last_response_ms || null
            })
            console.log(`Monitor created: ${data.monitor.friendly_name}`)
          } else if (data.type === 'monitor_updated') {
            const idx = monitors.value.findIndex(m => m.id === data.monitor.id)
            if (idx !== -1) {
              monitors.value[idx] = {
                ...data.monitor,
                status: data.monitor.status === 'up' ? 'up' : 'down',
                last_check: data.monitor.last_checked_at || null,
                last_response_ms: data.monitor.last_response_ms || null
              }
              console.log(`Monitor updated: ${data.monitor.friendly_name}`)
            }
          } else if (data.type === 'monitor_deleted') {
            monitors.value = monitors.value.filter(m => m.id !== data.monitor_id)
            console.log(`Monitor deleted: ${data.monitor_id}`)
          }
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e)
        }
      }

      ws.value.onclose = () => {
        console.log('   Monitor WebSocket disconnected')
        wsConnected.value = false
        setTimeout(connectWebSocket, 3000)
      }

      ws.value.onerror = (err) => {
        console.error('Monitor WebSocket error:', err)
        wsConnected.value = false
      }
    }

    // API Methods
    const fetchMonitors = async () => {
      isLoading.value = true
      try {
        const response = await fetch(`${API_BASE_URL}/v1/api/monitors`)
        
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }
        
        const data = await response.json()
        monitors.value = (data || []).map(monitor => ({
          ...monitor,
          status: monitor.status === 'up' ? 'up' : 'down',
          last_check: monitor.last_check || monitor.last_checked_at || new Date().toISOString(),
          last_response_ms: monitor.last_response_ms || null
        }))
        
        console.log('Loaded monitors via REST API:', monitors.value.length)
      } catch (error) {
        console.error('Error fetching monitors:', error)
        // alert(`Failed to load monitors: ${error.message}`)
      } finally {
        isLoading.value = false
      }
    }

    const saveMonitor = async () => {
      if (!config.friendly_name) {
        alert('Please enter a friendly name')
        return
      }
      if (!config.hostname) {
        alert('Please enter a hostname')
        return
      }
      if (!config.port || config.port < 1 || config.port > 65535) {
        alert('Please enter a valid port number (1-65535)')
        return
      }

      isLoading.value = true
      
      try {
        const payload = {
          friendly_name: config.friendly_name,
          hostname: config.hostname,
          port: parseInt(config.port),
          heartbeat_interval: parseInt(config.heartbeat_interval),
          retries: parseInt(config.retries),
          heartbeat_retry_interval: parseInt(config.heartbeat_retry_interval)
        }

        let response
        if (editingMonitorId.value) {
          response = await fetch(`${API_BASE_URL}/v1/api/monitors/${editingMonitorId.value}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
          })
        } else {
          response = await fetch(`${API_BASE_URL}/v1/api/monitors`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
          })
        }

        if (!response.ok) {
          const error = await response.json()
          throw new Error(error.message || `HTTP ${response.status}`)
        }

        alert(editingMonitorId.value ? 'Monitor updated successfully' : 'Monitor created successfully')
        resetForm()
        
        if (!wsConnected.value) {
          await fetchMonitors()
        }
      } catch (error) {
        console.error('Error saving monitor:', error)
        alert(`Failed to save monitor: ${error.message}`)
      } finally {
        isLoading.value = false
      }
    }

    const deleteMonitor = async (monitor) => {
      if (!confirm(`Are you sure you want to delete "${monitor.friendly_name}"?`)) {
        return
      }

      isLoading.value = true
      
      try {
        const response = await fetch(`${API_BASE_URL}/v1/api/monitors/${monitor.id}`, {
          method: 'DELETE'
        })

        if (!response.ok) {
          const error = await response.json()
          throw new Error(error.message || `HTTP ${response.status}`)
        }

        alert(`Monitor "${monitor.friendly_name}" deleted successfully`)
        
        if (!wsConnected.value) {
          await fetchMonitors()
        }
        
        if (editingMonitorId.value === monitor.id) {
          resetForm()
        }
      } catch (error) {
        console.error('Error deleting monitor:', error)
        alert(`Failed to delete monitor: ${error.message}`)
      } finally {
        isLoading.value = false
      }
    }

    const editMonitor = (monitor) => {
      editingMonitorId.value = monitor.id
      config.friendly_name = monitor.friendly_name
      config.hostname = monitor.hostname
      config.port = monitor.port
      config.heartbeat_interval = monitor.heartbeat_interval
      config.retries = monitor.retries
      config.heartbeat_retry_interval = monitor.heartbeat_retry_interval || 5
      
      document.querySelector('.config-card').scrollIntoView({ behavior: 'smooth' })
    }

    const cancelEdit = () => {
      resetForm()
    }

    const testConnection = async () => {
      if (!config.hostname || !config.port) {
        alert('Please enter hostname and port first')
        return
      }

      showTestModal.value = true
      testResult.value = {
        status: 'testing',
        message: 'Testing connection...',
        details: { host: config.hostname, port: config.port, responseTime: 0 }
      }

      try {
        const startTime = Date.now()
        const response = await fetch(`${API_BASE_URL}/v1/api/monitors/test`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ hostname: config.hostname, port: parseInt(config.port) })
        })

        const responseTime = Date.now() - startTime

        if (response.ok) {
          const data = await response.json()
          testResult.value = {
            status: 'success',
            message: data.message || 'Successfully connected to host:port',
            details: { host: config.hostname, port: config.port, responseTime }
          }
        } else {
          const error = await response.json()
          testResult.value = {
            status: 'error',
            message: error.message || 'Connection failed',
            details: { host: config.hostname, port: config.port, responseTime }
          }
        }
      } catch (error) {
        testResult.value = {
          status: 'error',
          message: error.message || 'Connection failed',
          details: { host: config.hostname, port: config.port, responseTime: 0 }
        }
      }
    }

    const resetForm = () => {
      editingMonitorId.value = null
      config.friendly_name = ''
      config.hostname = ''
      config.port = 80
      config.heartbeat_interval = 60
      config.retries = 2
      config.heartbeat_retry_interval = 5
    }

    // Helper Methods
    const formatHeartbeatTime = (seconds) => {
      if (!seconds) return '0s'
      const hours = Math.floor(seconds / 3600)
      const minutes = Math.floor((seconds % 3600) / 60)
      const remainingSeconds = seconds % 60
      
      if (hours > 0) return `${hours}h ${minutes}m`
      if (minutes > 0) return `${minutes}m ${remainingSeconds}s`
      return `${seconds}s`
    }

    const getHeartbeatCategory = (seconds) => {
      if (seconds <= 60) return 'very-fast'
      if (seconds <= 300) return 'fast'
      if (seconds <= 900) return 'medium'
      if (seconds <= 1800) return 'slow'
      return 'very-slow'
    }

    const formatTimestamp = (timestamp) => {
      if (!timestamp) return 'Never'
      return new Date(timestamp).toLocaleString()
    }

    const formatLastCheck = (timestamp) => {
      if (!timestamp) return 'Never'
      const date = new Date(timestamp)
      const now = new Date()
      const diffMs = now - date
      const diffSeconds = Math.floor(diffMs / 1000)
      const diffMinutes = Math.floor(diffSeconds / 60)
      const diffHours = Math.floor(diffMinutes / 60)
      const diffDays = Math.floor(diffHours / 24)
      
      if (diffDays > 0) return `${diffDays}d ago`
      if (diffHours > 0) return `${diffHours}h ago`
      if (diffMinutes > 0) return `${diffMinutes}m ago`
      if (diffSeconds > 0) return `${diffSeconds}s ago`
      return 'Just now'
    }

    const getResponseTimeClass = (ms) => {
      if (ms < 50) return 'fast'
      if (ms < 200) return 'normal'
      if (ms < 500) return 'slow'
      return 'very-slow'
    }

    const getHeartbeatBlockClass = (monitor, index) => {
      if (!monitor.last_check) return ''
      
      const now = Date.now()
      const lastCheck = new Date(monitor.last_check).getTime()
      const interval = monitor.heartbeat_interval * 1000
      const blockTime = lastCheck - (index * interval)
      const timeSinceBlock = now - blockTime
      
      if (monitor.status === 'up') {
        if (timeSinceBlock < interval * 2) return 'success'
        if (timeSinceBlock < interval * 4) return 'warning'
      }
      return 'danger'
    }

    const toggleSort = (field) => {
      if (sortBy.value === field) {
        sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
      } else {
        sortBy.value = field
        sortDirection.value = 'asc'
      }
    }

    const getSortIcon = (field) => {
      if (sortBy.value !== field) return '↕️'
      return sortDirection.value === 'asc' ? '↑' : '↓'
    }

    const closeTestModal = () => {
      showTestModal.value = false
    }

    // Lifecycle
    onMounted(() => {
      fetchMonitors()
      connectWebSocket()
    })

    onUnmounted(() => {
      if (ws.value) ws.value.close()
    })

    return {
      monitors,
      isLoading,
      editingMonitorId,
      config,
      upCount,
      downCount,
      sortedMonitors,
      showTestModal,
      testResult,
      sortBy,
      sortDirection,
      wsConnected,
      formatHeartbeatTime,
      getHeartbeatCategory,
      formatTimestamp,
      formatLastCheck,
      getResponseTimeClass,
      getHeartbeatBlockClass,
      toggleSort,
      getSortIcon,
      fetchMonitors,
      saveMonitor,
      deleteMonitor,
      editMonitor,
      cancelEdit,
      testConnection,
      resetForm,
      closeTestModal
    }
  }
}
</script>

<style scoped>
.tcp-monitor {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  min-height: 100vh;
  color: #e2e8f0;
}

/* Cards */
.config-card,
.monitors-list-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  margin-bottom: 24px;
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  background: rgba(15, 23, 42, 0.4);
  flex-wrap: wrap;
  gap: 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 24px;
  flex-wrap: wrap;
}

.card-header h2 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: #f8fafc;
}

/* Sort Controls */
.sort-controls {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.sort-btn {
  background: #1e293b;
  border: 1px solid #334155;
  color: #94a3b8;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.sort-btn:hover {
  background: #2d3748;
  border-color: #3b82f6;
  color: #60a5fa;
}

.sort-btn.active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: white;
}

.sort-icon {
  font-size: 0.9rem;
}

.sortable-header {
  cursor: pointer;
  user-select: none;
}

.sortable-header:hover {
  color: #60a5fa;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.monitor-count {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 0.9rem;
  background: #1e293b;
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid #334155;
}

.up-count {
  color: #34d399;
  display: flex;
  align-items: center;
  gap: 4px;
}

.down-count {
  color: #f87171;
  display: flex;
  align-items: center;
  gap: 4px;
}

.total-count {
  color: #94a3b8;
  padding-left: 8px;
  border-left: 1px solid #334155;
}

/* Form Styles */
.config-form {
  padding: 24px;
}

.form-row {
  margin-bottom: 20px;
}

.form-row.dual {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 0.9rem;
  font-weight: 500;
  color: #cbd5e1;
}

.form-input {
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 10px 14px;
  color: #e2e8f0;
  font-size: 0.95rem;
  transition: all 0.2s;
  width: 100%;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

/* Heartbeat Styles */
.heartbeat-interval-container {
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 12px;
  padding: 16px;
}

.heartbeat-visualization {
  margin-bottom: 16px;
}

.heartbeat-bar {
  height: 8px;
  background: #1e293b;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 4px;
}

.heartbeat-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #8b5cf6);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.heartbeat-markers {
  display: flex;
  justify-content: space-between;
  font-size: 0.7rem;
  color: #64748b;
  padding: 0 4px;
}

.marker {
  position: relative;
}

.marker::before {
  content: '';
  position: absolute;
  top: -12px;
  left: 50%;
  transform: translateX(-50%);
  width: 2px;
  height: 4px;
  background: #334155;
}

.heartbeat-input-group {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.heartbeat-number-input {
  display: flex;
  align-items: center;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  overflow: hidden;
}

.heartbeat-number-field {
  width: 80px;
  padding: 8px;
  background: #1e293b;
  border: none;
  color: #e2e8f0;
  font-size: 0.95rem;
  text-align: center;
}

.heartbeat-number-field:focus {
  outline: none;
}

.heartbeat-unit {
  padding: 8px;
  background: #0f172a;
  color: #94a3b8;
  font-size: 0.85rem;
  border-left: 1px solid #334155;
}

.heartbeat-display {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.heartbeat-value {
  font-size: 1.1rem;
  font-weight: 600;
  color: #60a5fa;
}

.heartbeat-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
}

.heartbeat-badge.very-fast {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.heartbeat-badge.fast {
  background: rgba(59, 130, 246, 0.1);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.heartbeat-badge.medium {
  background: rgba(245, 158, 11, 0.1);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.heartbeat-badge.slow {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.heartbeat-description {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  background: #1e293b;
  border-radius: 6px;
  font-size: 0.85rem;
  color: #94a3b8;
}

/* Retries */
.retries-container {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.retries-input {
  width: 80px;
}

.retries-visualization {
  display: flex;
  gap: 4px;
}

.retry-dot {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #1e293b;
  border: 2px solid #334155;
  transition: all 0.2s ease;
}

.retry-dot.active {
  background: #3b82f6;
  border-color: #3b82f6;
  box-shadow: 0 0 10px rgba(59, 130, 246, 0.3);
}

.input-hint {
  font-size: 0.8rem;
  color: #64748b;
  flex: 1;
}

/* Interval input */
.interval-input {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.interval-field {
  width: 80px;
}

.interval-unit {
  color: #94a3b8;
  font-size: 0.9rem;
}

/* Form Actions */
.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
  flex-wrap: wrap;
}

.btn {
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 0.9rem;
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
  transform: translateY(-1px);
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

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.btn-icon-only {
  padding: 8px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 8px;
  color: #94a3b8;
  cursor: pointer;
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-icon-only:hover:not(:disabled) {
  background: #2d3748;
  color: #60a5fa;
  transform: rotate(90deg);
}

.btn-icon {
  font-size: 1rem;
}

/* Monitors Table */
.monitors-list-card {
  display: flex;
  flex-direction: column;
  max-height: 600px;
}

.monitors-table-container {
  overflow-y: auto;
  overflow-x: auto;
  flex: 1;
}

.monitors-table-container::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.monitors-table-container::-webkit-scrollbar-track {
  background: #0f172a;
}

.monitors-table-container::-webkit-scrollbar-thumb {
  background: #334155;
  border-radius: 4px;
}

.monitors-table {
  min-width: 1000px;
  padding: 0 24px 24px 24px;
}

.monitors-table table {
  width: 100%;
  border-collapse: collapse;
}

.monitors-table table thead {
  position: sticky;
  top: 0;
  background: rgba(30, 41, 59, 0.95);
  z-index: 10;
  backdrop-filter: blur(4px);
}

.monitors-table table thead th {
  background: rgba(30, 41, 59, 0.95);
  padding: 16px 8px 12px 8px;
  text-align: left;
  color: #94a3b8;
  font-weight: 500;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.monitors-table table tbody td {
  padding: 12px 8px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.05);
  color: #e2e8f0;
}

.monitor-row:hover {
  background: rgba(59, 130, 246, 0.1);
}

/* Status Badge */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
}

.status-badge.up {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.status-badge.down {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.friendly-name {
  font-weight: 500;
  color: #f8fafc;
}

.host-port {
  font-family: monospace;
  color: #60a5fa;
}

/* Heartbeat Timeline */
.heartbeat-cell {
  min-width: 150px;
}

.heartbeat-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.heartbeat-interval-badge {
  padding: 4px 8px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 600;
  color: #60a5fa;
}

.heartbeat-interval-human {
  font-size: 0.75rem;
  color: #94a3b8;
}

.heartbeat-timeline {
  display: flex;
  gap: 2px;
  margin-top: 4px;
}

.heartbeat-block {
  flex: 1;
  height: 4px;
  border-radius: 2px;
  background: #1e293b;
}

.heartbeat-block.success {
  background: #10b981;
}

.heartbeat-block.warning {
  background: #f59e0b;
}

.heartbeat-block.danger {
  background: #ef4444;
}

/* Retries Badge */
.retries-badge {
  display: inline-flex;
  align-items: center;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 12px;
  padding: 2px 8px;
}

.retries-count {
  font-weight: 600;
  color: #f8fafc;
}

.retries-max {
  color: #64748b;
  font-size: 0.8rem;
  margin-left: 2px;
}

/* Last Check Cell */
.last-check-cell {
  position: relative;
  padding-right: 20px !important;
}

.last-check-time {
  display: block;
  font-weight: 500;
  color: #f8fafc;
}

.last-check-timestamp {
  display: block;
  font-size: 0.7rem;
  color: #64748b;
}

.status-indicator {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-indicator.up {
  background: #10b981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
  animation: pulse 2s infinite;
}

.status-indicator.down {
  background: #ef4444;
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.5);
}

/* Response Time */
.response-time-cell {
  font-family: monospace;
}

.response-time {
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
}

.response-time.fast {
  color: #34d399;
  background: rgba(16, 185, 129, 0.1);
}

.response-time.normal {
  color: #60a5fa;
  background: rgba(59, 130, 246, 0.1);
}

.response-time.slow {
  color: #fbbf24;
  background: rgba(245, 158, 11, 0.1);
}

.response-time.very-slow {
  color: #f87171;
  background: rgba(239, 68, 68, 0.1);
}

.response-time.na {
  color: #64748b;
}

/* Actions */
.actions {
  display: flex;
  gap: 6px;
}

.action-btn {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 1.1rem;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #1e293b;
  color: #60a5fa;
  transform: scale(1.1);
}

/* WebSocket Status */
.ws-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  margin-bottom: 16px;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 8px;
  font-size: 0.8rem;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.ws-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.ws-status.connected .ws-indicator {
  background: #10b981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
}

.ws-status.disconnected .ws-indicator {
  background: #ef4444;
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.5);
}

.ws-status.connected .ws-text {
  color: #34d399;
}

.ws-status.disconnected .ws-text {
  color: #f87171;
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 60px 24px;
  color: #94a3b8;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-5px); }
}

.empty-state p {
  margin: 8px 0;
  font-size: 1rem;
}

.empty-hint {
  font-size: 0.9rem;
  color: #64748b;
}

/* Loading State */
.loading-state {
  text-align: center;
  padding: 60px 24px;
  color: #94a3b8;
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
  width: 450px;
  max-width: 90vw;
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: #f8fafc;
}

.close-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 24px;
  cursor: pointer;
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
  padding: 20px;
}

/* Test Result */
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

.result-icon {
  font-size: 32px;
}

.result-details h4 {
  margin: 0 0 8px 0;
  font-size: 1rem;
  color: #f8fafc;
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
  padding: 16px 20px 20px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* Responsive */
@media (max-width: 1024px) {
  .tcp-monitor {
    padding: 16px;
  }
}

@media (max-width: 768px) {
  .form-row.dual {
    grid-template-columns: 1fr;
  }

  .form-actions {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }

  .heartbeat-input-group {
    flex-direction: column;
    align-items: stretch;
  }

  .heartbeat-display {
    justify-content: space-between;
  }

  .retries-container {
    flex-direction: column;
    align-items: flex-start;
  }

  .retries-visualization {
    width: 100%;
    justify-content: space-between;
  }

  .retry-dot {
    flex: 1;
  }

  .monitors-table {
    min-width: 800px;
  }

  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-left {
    width: 100%;
    justify-content: space-between;
  }
}
</style>