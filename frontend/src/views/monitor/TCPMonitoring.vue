<template>
  <div class="tcp-monitor">
 

    <!-- Monitor Configuration Form -->
    <div class="config-card">
      <div class="card-header">
        <h2>Monitor Configuration</h2>
      </div>

      <div class="config-form">
        <!-- Friendly Name -->
        <div class="form-row">
          <div class="form-group">
            <label>Friendly Name</label>
            <input 
              type="text" 
              v-model="config.friendlyName" 
              placeholder="New Monitor"
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
              placeholder="161"
              min="1"
              max="65535"
              class="form-input"
            />
          </div>
        </div>

        <!-- Heartbeat Interval - Uptime Kuma Style -->
        <div class="form-row">
          <div class="form-group heartbeat-group">
            <label>Heartbeat Interval</label>
            <div class="heartbeat-interval-container">
              <!-- Heartbeat Visualization Bar -->
              <div class="heartbeat-visualization">
                <div class="heartbeat-bar">
                  <div 
                    class="heartbeat-fill" 
                    :style="{ width: (config.heartbeatInterval / 300) * 100 + '%' }"
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

              <!-- Heartbeat Input and Display -->
              <div class="heartbeat-input-group">
                <div class="heartbeat-number-input">
                  <input 
                    type="number" 
                    v-model="config.heartbeatInterval" 
                    min="1"
                    max="3600"
                    class="heartbeat-number-field"
                  />
                  <span class="heartbeat-unit">seconds</span>
                </div>
                <div class="heartbeat-display">
                  <span class="heartbeat-value">{{ formatHeartbeatTime(config.heartbeatInterval) }}</span>
                  <span class="heartbeat-badge" :class="getHeartbeatCategory(config.heartbeatInterval)">
                    {{ getHeartbeatCategory(config.heartbeatInterval) }}
                  </span>
                </div>
              </div>

              <!-- Heartbeat Description -->
              <div class="heartbeat-description">
                <span class="description-icon">⏱️</span>
                <span class="description-text">
                  Check every <strong>{{ config.heartbeatInterval }} seconds</strong> 
                  ({{ formatHeartbeatTime(config.heartbeatInterval) }})
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
                v-model="config.retryInterval" 
                min="1"
                class="form-input interval-field"
              />
              <span class="interval-unit">seconds</span>
            </div>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="form-actions">
          <button class="btn btn-primary" @click="saveMonitor">
            <span class="btn-icon">💾</span>
            Save Monitor
          </button>
          <button class="btn btn-secondary" @click="testConnection">
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

    <!-- Active Monitors List - Scrollable with Sorting -->
    <div class="monitors-list-card">
      <div class="card-header">
        <div class="header-left">
          <h2>Active Monitors</h2>
          <div class="sort-controls">
            <button 
              class="sort-btn" 
              :class="{ active: sortBy === 'status' }"
              @click="toggleSort('status')"
              title="Sort by status"
            >
              <span>Sort by Status</span>
              <span class="sort-icon">{{ getSortIcon('status') }}</span>
            </button>
            <button 
              class="sort-btn" 
              :class="{ active: sortBy === 'name' }"
              @click="toggleSort('name')"
              title="Sort by name"
            >
              <span>Sort by Name</span>
              <span class="sort-icon">{{ getSortIcon('name') }}</span>
            </button>
            <button 
              class="sort-btn" 
              :class="{ active: sortBy === 'host' }"
              @click="toggleSort('host')"
              title="Sort by host"
            >
              <span>Sort by Host</span>
              <span class="sort-icon">{{ getSortIcon('host') }}</span>
            </button>
          </div>
        </div>
        <div class="header-actions">
          <span class="monitor-count">
            <span class="online-count">🟢 {{ onlineCount }}</span>
            <span class="offline-count">🔴 {{ offlineCount }}</span>
            <span class="total-count">Total: {{ monitors.length }}</span>
          </span>
          <button class="btn btn-icon-only" @click="refreshMonitors" title="Refresh">
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
                <th @click="toggleSort('name')" class="sortable-header">
                  Friendly Name
                  <span class="sort-icon">{{ getSortIcon('name') }}</span>
                </th>
                <th @click="toggleSort('host')" class="sortable-header">
                  Host:Port
                  <span class="sort-icon">{{ getSortIcon('host') }}</span>
                </th>
                <th>Heartbeat</th>
                <th>Retries</th>
                <th>Last Heartbeat</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="monitor in sortedMonitors" :key="monitor.id" class="monitor-row">
                <td data-label="Status">
                  <span class="status-badge" :class="monitor.status">
                    {{ monitor.status === 'online' ? '🟢' : '🔴' }}
                  </span>
                </td>
                <td data-label="Friendly Name" class="friendly-name">{{ monitor.friendlyName }}</td>
                <td data-label="Host:Port" class="host-port">{{ monitor.hostname }}:{{ monitor.port }}</td>
                <td data-label="Heartbeat" class="heartbeat-cell">
                  <div class="heartbeat-info">
                    <span class="heartbeat-interval-badge">{{ monitor.interval }}s</span>
                    <span class="heartbeat-interval-human">{{ formatHeartbeatTime(monitor.interval) }}</span>
                  </div>
                  <!-- Heartbeat Timeline Visualization -->
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
                    <span class="retries-max">/3</span>
                  </div>
                </td>
                <td data-label="Last Heartbeat" class="last-heartbeat-cell">
                  <span class="last-heartbeat-time">{{ formatLastHeartbeat(monitor.lastCheck) }}</span>
                  <span class="last-heartbeat-timestamp">{{ formatTimestamp(monitor.lastCheck) }}</span>
                  <div class="heartbeat-status-indicator" :class="monitor.status"></div>
                </td>
                <td data-label="Actions" class="actions">
                  <button class="action-btn" @click="editMonitor(monitor)" title="Edit">
                    ✏️
                  </button>
                  <button class="action-btn" @click="pauseMonitor(monitor)" title="Pause">
                    ⏸️
                  </button>
                  <button class="action-btn" @click="deleteMonitor(monitor)" title="Delete">
                    🗑️
                  </button>
                </td>
              </tr>
              <tr v-if="monitors.length === 0">
                <td colspan="7" class="empty-state">
                  <div class="empty-icon">📡</div>
                  <p>No monitors configured yet</p>
                  <p class="empty-hint">Create your first monitor using the form above</p>
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
import { ref, reactive, computed } from 'vue'

export default {
  name: 'TCPMonitor',
  
  setup() {
    // Configuration state
    const config = reactive({
      friendlyName: 'New Monitor',
      hostname: '',
      port: 161,
      heartbeatInterval: 60,
      retries: 0,
      retryInterval: 60
    })

    // Sorting state
    const sortBy = ref('status')
    const sortDirection = ref('desc')

    // Monitors list with sample data
    const monitors = ref([
      {
        id: 1,
        friendlyName: 'Production Web Server',
        hostname: '192.168.1.100',
        port: 80,
        interval: 60,
        retries: 2,
        status: 'online',
        lastCheck: new Date(Date.now() - 30000).toISOString()
      },
      {
        id: 2,
        friendlyName: 'Database Server',
        hostname: '192.168.1.101',
        port: 3306,
        interval: 120,
        retries: 3,
        status: 'online',
        lastCheck: new Date(Date.now() - 45000).toISOString()
      },
      {
        id: 3,
        friendlyName: 'Mail Server',
        hostname: 'mail.example.com',
        port: 25,
        interval: 300,
        retries: 1,
        status: 'offline',
        lastCheck: new Date(Date.now() - 120000).toISOString()
      },
      {
        id: 4,
        friendlyName: 'DNS Server',
        hostname: '8.8.8.8',
        port: 53,
        interval: 60,
        retries: 2,
        status: 'online',
        lastCheck: new Date(Date.now() - 15000).toISOString()
      },
      {
        id: 5,
        friendlyName: 'SSH Server',
        hostname: '192.168.1.200',
        port: 22,
        interval: 120,
        retries: 3,
        status: 'online',
        lastCheck: new Date(Date.now() - 60000).toISOString()
      },
      {
        id: 6,
        friendlyName: 'Redis Cache',
        hostname: 'cache.internal',
        port: 6379,
        interval: 60,
        retries: 2,
        status: 'offline',
        lastCheck: new Date(Date.now() - 300000).toISOString()
      }
    ])

    // Computed properties
    const onlineCount = computed(() => {
      return monitors.value.filter(m => m.status === 'online').length
    })

    const offlineCount = computed(() => {
      return monitors.value.filter(m => m.status === 'offline').length
    })

    const sortedMonitors = computed(() => {
      const sorted = [...monitors.value]
      
      sorted.sort((a, b) => {
        let comparison = 0
        
        switch (sortBy.value) {
          case 'status':
            comparison = (a.status === 'online' ? -1 : 1) - (b.status === 'online' ? -1 : 1)
            break
          case 'name':
            comparison = a.friendlyName.localeCompare(b.friendlyName)
            break
          case 'host':
            comparison = a.hostname.localeCompare(b.hostname)
            break
          default:
            return 0
        }
        
        return sortDirection.value === 'asc' ? comparison : -comparison
      })
      
      return sorted
    })

    // Test modal state
    const showTestModal = ref(false)
    const testResult = ref({
      status: 'success',
      message: 'Successfully connected to host:port',
      details: {
        host: '',
        port: 161,
        responseTime: 45
      }
    })

    // Helper Methods
    const formatHeartbeatTime = (seconds) => {
      if (!seconds) return '0s'
      
      const hours = Math.floor(seconds / 3600)
      const minutes = Math.floor((seconds % 3600) / 60)
      const remainingSeconds = seconds % 60
      
      if (hours > 0) {
        return `${hours}h ${minutes}m`
      } else if (minutes > 0) {
        return `${minutes}m ${remainingSeconds}s`
      } else {
        return `${seconds}s`
      }
    }

    const getHeartbeatCategory = (seconds) => {
      if (seconds <= 60) return 'very-fast'
      if (seconds <= 300) return 'fast'
      if (seconds <= 900) return 'medium'
      if (seconds <= 1800) return 'slow'
      return 'very-slow'
    }

    const formatTimestamp = (timestamp) => {
      const date = new Date(timestamp)
      return date.toLocaleTimeString()
    }

    const formatLastHeartbeat = (timestamp) => {
      const date = new Date(timestamp)
      const now = new Date()
      const diffMs = now - date
      const diffSeconds = Math.floor(diffMs / 1000)
      const diffMinutes = Math.floor(diffSeconds / 60)
      const diffHours = Math.floor(diffMinutes / 60)
      const diffDays = Math.floor(diffHours / 24)
      
      if (diffDays > 0) {
        return `${diffDays}d ago`
      } else if (diffHours > 0) {
        return `${diffHours}h ago`
      } else if (diffMinutes > 0) {
        return `${diffMinutes}m ago`
      } else {
        return 'Just now'
      }
    }

    const getHeartbeatBlockClass = (monitor, index) => {
      const now = Date.now()
      const lastCheck = new Date(monitor.lastCheck).getTime()
      const interval = monitor.interval * 1000
      
      // Simulate some recent heartbeats based on last check
      const blockTime = lastCheck - (index * interval * 2)
      const timeSinceBlock = now - blockTime
      
      if (monitor.status === 'online') {
        if (timeSinceBlock < interval * 3) {
          return 'success'
        } else if (timeSinceBlock < interval * 6) {
          return 'warning'
        }
      }
      
      return 'danger'
    }

    // Methods
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

    const saveMonitor = () => {
      if (!config.hostname) {
        alert('Please enter a hostname')
        return
      }
      if (!config.port || config.port < 1 || config.port > 65535) {
        alert('Please enter a valid port number (1-65535)')
        return
      }

      const newMonitor = {
        id: Date.now(),
        friendlyName: config.friendlyName || 'New Monitor',
        hostname: config.hostname,
        port: config.port,
        interval: config.heartbeatInterval,
        retries: config.retries,
        status: 'pending',
        lastCheck: new Date().toISOString()
      }

      monitors.value.push(newMonitor)
      resetForm()
    }

    const testConnection = () => {
      testResult.value = {
        status: Math.random() > 0.3 ? 'success' : 'error',
        message: Math.random() > 0.3 
          ? 'Successfully connected to host:port'
          : 'Connection refused: Unable to establish TCP connection',
        details: {
          host: config.hostname || 'test.example.com',
          port: config.port || 161,
          responseTime: Math.floor(Math.random() * 100) + 20
        }
      }
      showTestModal.value = true
    }

    const resetForm = () => {
      config.friendlyName = 'New Monitor'
      config.hostname = ''
      config.port = 161
      config.heartbeatInterval = 60
      config.retries = 0
      config.retryInterval = 60
    }

    const editMonitor = (monitor) => {
      config.friendlyName = monitor.friendlyName
      config.hostname = monitor.hostname
      config.port = monitor.port
      config.heartbeatInterval = monitor.interval
      config.retries = monitor.retries
      config.retryInterval = monitor.interval
      
      document.querySelector('.config-card').scrollIntoView({ behavior: 'smooth' })
    }

    const pauseMonitor = (monitor) => {
      const index = monitors.value.findIndex(m => m.id === monitor.id)
      if (index !== -1) {
        monitors.value[index].status = monitors.value[index].status === 'online' ? 'offline' : 'online'
      }
    }

    const deleteMonitor = (monitor) => {
      if (confirm(`Are you sure you want to delete "${monitor.friendlyName}"?`)) {
        monitors.value = monitors.value.filter(m => m.id !== monitor.id)
      }
    }

    const refreshMonitors = () => {
      monitors.value = monitors.value.map(monitor => ({
        ...monitor,
        status: Math.random() > 0.2 ? 'online' : 'offline',
        lastCheck: new Date().toISOString()
      }))
    }

    const closeTestModal = () => {
      showTestModal.value = false
    }

    return {
      config,
      monitors,
      sortedMonitors,
      onlineCount,
      offlineCount,
      showTestModal,
      testResult,
      sortBy,
      sortDirection,
      formatHeartbeatTime,
      getHeartbeatCategory,
      formatTimestamp,
      formatLastHeartbeat,
      getHeartbeatBlockClass,
      toggleSort,
      getSortIcon,
      saveMonitor,
      testConnection,
      resetForm,
      editMonitor,
      pauseMonitor,
      deleteMonitor,
      refreshMonitors,
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

/* Header */
.monitor-header {
  margin-bottom: 30px;
}

.monitor-header h1 {
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
  transition: all 0.2s;
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

.online-count {
  color: #34d399;
  display: flex;
  align-items: center;
  gap: 4px;
}

.offline-count {
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
  display: flex;
  align-items: center;
  gap: 8px;
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

/* Uptime Kuma Style Heartbeat */
.heartbeat-group {
  gap: 8px;
}

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

.description-icon {
  font-size: 1rem;
}

.description-text strong {
  color: #f8fafc;
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
  margin-bottom: 0;
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

.btn-primary:hover {
  background: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-secondary {
  background: #1e293b;
  color: #cbd5e1;
  border: 1px solid #334155;
}

.btn-secondary:hover {
  background: #2d3748;
  transform: translateY(-1px);
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover {
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

.btn-icon-only:hover {
  background: #2d3748;
  color: #60a5fa;
  transform: rotate(90deg);
}

/* Monitors List Card */
.monitors-list-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  margin-bottom: 24px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  max-height: 500px;
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
  transition: all 0.2s;
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

/* Last Heartbeat */
.last-heartbeat-cell {
  position: relative;
  padding-right: 20px !important;
}

.last-heartbeat-time {
  display: block;
  font-weight: 500;
  color: #f8fafc;
}

.last-heartbeat-timestamp {
  display: block;
  font-size: 0.7rem;
  color: #64748b;
}

.heartbeat-status-indicator {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.heartbeat-status-indicator.online {
  background: #10b981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
  animation: pulse 2s infinite;
}

.heartbeat-status-indicator.offline {
  background: #ef4444;
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.5);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
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

/* Responsive */
@media (max-width: 768px) {
  .tcp-monitor {
    padding: 16px;
  }

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
}
</style>