<template>
  <div class="tcp-monitor">

    <!-- Monitor Configuration Form -->
    <div class="config-card">

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

        <!-- Heartbeat Interval -->
        <div class="form-row">
          <div class="form-group">
            <label>Heartbeat Interval <span class="label-hint">(Check every 60 seconds)</span></label>
            <div class="interval-input">
              <input 
                type="number" 
                v-model="config.heartbeatInterval" 
                min="1"
                class="form-input interval-field"
              />
              <span class="interval-unit">seconds</span>
              <span class="interval-example">1 minute</span>
            </div>
          </div>
        </div>

        <!-- Retries -->
        <div class="form-row">
          <div class="form-group">
            <label>Retries</label>
            <input 
              type="number" 
              v-model="config.retries" 
              min="0"
              max="10"
              class="form-input retries-input"
            />
            <span class="input-hint">Maximum retries before the service is marked as down and a notification is sent</span>
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
                <th>Interval</th>
                <th>Retries</th>
                <th>Last Check</th>
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
                <td data-label="Interval">{{ monitor.interval }}s</td>
                <td data-label="Retries">{{ monitor.retries }}</td>
                <td data-label="Last Check" class="last-check">{{ formatLastCheck(monitor.lastCheck) }}</td>
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
    const sortDirection = ref('desc') // 'asc' or 'desc'

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

    // Computed properties for counts
    const onlineCount = computed(() => {
      return monitors.value.filter(m => m.status === 'online').length
    })

    const offlineCount = computed(() => {
      return monitors.value.filter(m => m.status === 'offline').length
    })

    // Sorted monitors computed property
    const sortedMonitors = computed(() => {
      const sorted = [...monitors.value]
      
      sorted.sort((a, b) => {
        let comparison = 0
        
        switch (sortBy.value) {
          case 'status':
            // Sort by status (online first by default)
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

    // Methods
    const toggleSort = (field) => {
      if (sortBy.value === field) {
        // Toggle direction if same field
        sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
      } else {
        // New field, set to default direction
        sortBy.value = field
        sortDirection.value = field === 'status' ? 'desc' : 'asc'
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

    const formatLastCheck = (timestamp) => {
      const date = new Date(timestamp)
      const now = new Date()
      const diffMs = now - date
      const diffMins = Math.floor(diffMs / 60000)
      
      if (diffMins < 1) return 'Just now'
      if (diffMins < 60) return `${diffMins}m ago`
      return date.toLocaleTimeString()
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
      toggleSort,
      getSortIcon,
      saveMonitor,
      testConnection,
      resetForm,
      editMonitor,
      pauseMonitor,
      deleteMonitor,
      refreshMonitors,
      formatLastCheck,
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

.label-hint {
  font-size: 0.8rem;
  color: #64748b;
  font-weight: 400;
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

.form-input::placeholder {
  color: #475569;
}

/* Interval input */
.interval-input {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.interval-field {
  width: 100px;
}

.interval-unit {
  color: #94a3b8;
  font-size: 0.9rem;
}

.interval-example {
  background: #1e293b;
  color: #94a3b8;
  padding: 4px 10px;
  border-radius: 16px;
  font-size: 0.85rem;
  border: 1px solid #334155;
}

.retries-input {
  width: 100px;
  margin-bottom: 6px;
}

.input-hint {
  font-size: 0.8rem;
  color: #64748b;
  line-height: 1.4;
  max-width: 400px;
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

.btn-icon {
  font-size: 1.1rem;
}

/* Monitors List Card - Scrollable */
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

/* Table Container - Scrollable */
.monitors-table-container {
  overflow-y: auto;
  overflow-x: auto;
  flex: 1;
  scrollbar-width: thin;
  scrollbar-color: #334155 #0f172a;
}

/* Custom scrollbar styling */
.monitors-table-container::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.monitors-table-container::-webkit-scrollbar-track {
  background: #0f172a;
  border-radius: 4px;
}

.monitors-table-container::-webkit-scrollbar-thumb {
  background: #334155;
  border-radius: 4px;
}

.monitors-table-container::-webkit-scrollbar-thumb:hover {
  background: #475569;
}

/* Ensure table takes full width */
.monitors-table {
  min-width: 800px;
  padding: 0 24px 24px 24px;
}

/* Keep thead sticky */
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

.monitors-table table thead::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 1px;
  background: rgba(148, 163, 184, 0.2);
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

.status-badge {
  display: inline-block;
  font-size: 1.2rem;
}

.friendly-name {
  font-weight: 500;
  color: #f8fafc;
}

.host-port {
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  color: #60a5fa;
}

.last-check {
  color: #94a3b8;
  font-size: 0.9rem;
}

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
  min-width: 600px;
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

.modal-footer {
  display: flex;
  justify-content: flex-end;
  padding: 16px 20px 20px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
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

/* Responsive */
@media (max-width: 1024px) {
  .monitors-list-card {
    max-height: 400px;
  }
  
  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}

@media (max-width: 768px) {
  .tcp-monitor {
    padding: 16px;
  }

  .form-row.dual {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .form-actions {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }

  .interval-input {
    flex-direction: column;
    align-items: flex-start;
  }

  .interval-field {
    width: 100%;
  }

  .monitors-list-card {
    max-height: 400px;
  }

  .monitors-table-container {
    overflow-x: auto;
  }

  .monitors-table {
    min-width: 600px;
    padding: 0 16px 16px 16px;
  }

  .monitors-table table thead {
    position: sticky;
    top: 0;
  }

  /* Mobile card view */
  .monitors-table table,
  .monitors-table table thead,
  .monitors-table table tbody,
  .monitors-table table tr,
  .monitors-table table td {
    display: block;
  }

  .monitors-table table thead {
    display: none;
  }

  .monitor-row {
    display: block;
    border: 1px solid rgba(148, 163, 184, 0.1);
    border-radius: 8px;
    margin-bottom: 12px;
    padding: 12px;
    background: rgba(15, 23, 42, 0.4);
  }

  .monitors-table table tbody td {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid rgba(148, 163, 184, 0.05);
    text-align: right;
  }

  .monitors-table table tbody td:last-child {
    border-bottom: none;
  }

  .monitors-table table tbody td::before {
    content: attr(data-label);
    font-weight: 500;
    color: #94a3b8;
    text-align: left;
  }

  .actions {
    justify-content: flex-end;
  }
  
  .monitor-count {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .total-count {
    padding-left: 0;
    border-left: none;
  }
}
</style>