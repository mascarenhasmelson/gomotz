<template>
  <div class="snmp-monitor">

    <!-- SNMP Configuration Form -->
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
              placeholder="e.g., 192.168.1.1 or switch.local"
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

        <!-- Community String -->
        <div class="form-row">
          <div class="form-group">
            <label>Community String</label>
            <input 
              type="text" 
              v-model="config.community" 
              placeholder="public"
              class="form-input"
            />
            <span class="input-hint">This string functions as a password to authenticate and control access to SNMP-enabled devices. Match it with your SNMP device's configuration.</span>
          </div>
        </div>

        <!-- OID -->
        <div class="form-row">
          <div class="form-group">
            <label>OID (Object Identifier)</label>
            <input 
              type="text" 
              v-model="config.oid" 
              placeholder="1.3.6.1.4.1.9.6.1.101"
              class="form-input"
            />
            <span class="input-hint">Enter the OID for the sensor or status you want to monitor. Use network management tools like MIB browsers or SNMP software if you're unsure about the OID.</span>
          </div>
        </div>

        <!-- SNMP Version -->
        <div class="form-row">
          <div class="form-group">
            <label>SNMP Version</label>
            <div class="select-wrapper">
              <select v-model="config.snmpVersion" class="form-select">
                <option value="v1">SNMPv1</option>
                <option value="v2c">SNMPv2c</option>
                <option value="v3">SNMPv3</option>
              </select>
              <span class="select-arrow">▼</span>
            </div>
          </div>
        </div>

        <!-- SNMPv3 specific fields (shown only when v3 is selected) -->
        <div v-if="config.snmpVersion === 'v3'" class="v3-fields">
          <div class="form-row dual">
            <div class="form-group">
              <label>Security Name</label>
              <input 
                type="text" 
                v-model="config.securityName" 
                placeholder="username"
                class="form-input"
              />
            </div>
            <div class="form-group">
              <label>Security Level</label>
              <div class="select-wrapper">
                <select v-model="config.securityLevel" class="form-select">
                  <option value="noAuthNoPriv">noAuthNoPriv</option>
                  <option value="authNoPriv">authNoPriv</option>
                  <option value="authPriv">authPriv</option>
                </select>
                <span class="select-arrow">▼</span>
              </div>
            </div>
          </div>
          
          <div v-if="config.securityLevel !== 'noAuthNoPriv'" class="form-row dual">
            <div class="form-group">
              <label>Auth Protocol</label>
              <div class="select-wrapper">
                <select v-model="config.authProtocol" class="form-select">
                  <option value="MD5">MD5</option>
                  <option value="SHA">SHA</option>
                </select>
                <span class="select-arrow">▼</span>
              </div>
            </div>
            <div class="form-group">
              <label>Auth Password</label>
              <input 
                type="password" 
                v-model="config.authPassword" 
                placeholder="••••••••"
                class="form-input"
              />
            </div>
          </div>
          
          <div v-if="config.securityLevel === 'authPriv'" class="form-row dual">
            <div class="form-group">
              <label>Privacy Protocol</label>
              <div class="select-wrapper">
                <select v-model="config.privProtocol" class="form-select">
                  <option value="DES">DES</option>
                  <option value="AES">AES</option>
                </select>
                <span class="select-arrow">▼</span>
              </div>
            </div>
            <div class="form-group">
              <label>Privacy Password</label>
              <input 
                type="password" 
                v-model="config.privPassword" 
                placeholder="••••••••"
                class="form-input"
              />
            </div>
          </div>
        </div>

        <!-- Polling Interval -->
        <div class="form-row">
          <div class="form-group">
            <label>Polling Interval <span class="label-hint">(Check every 60 seconds)</span></label>
            <div class="interval-input">
              <input 
                type="number" 
                v-model="config.interval" 
                min="5"
                class="form-input interval-field"
              />
              <span class="interval-unit">seconds</span>
              <span class="interval-example">1 minute</span>
            </div>
          </div>
        </div>

        <!-- Timeout and Retries -->
        <div class="form-row dual">
          <div class="form-group">
            <label>Timeout</label>
            <div class="interval-input">
              <input 
                type="number" 
                v-model="config.timeout" 
                min="1"
                class="form-input interval-field"
              />
              <span class="interval-unit">seconds</span>
            </div>
          </div>
          <div class="form-group">
            <label>Retries</label>
            <input 
              type="number" 
              v-model="config.retries" 
              min="0"
              max="10"
              class="form-input"
            />
          </div>
        </div>

        <!-- Value Type (for OID response) -->
        <div class="form-row">
          <div class="form-group">
            <label>Expected Value Type</label>
            <div class="select-wrapper">
              <select v-model="config.valueType" class="form-select">
                <option value="integer">Integer</option>
                <option value="string">String</option>
                <option value="oid">OID</option>
                <option value="counter">Counter</option>
                <option value="gauge">Gauge</option>
                <option value="timeticks">TimeTicks</option>
              </select>
              <span class="select-arrow">▼</span>
            </div>
          </div>
        </div>

        <!-- Thresholds -->
        <div class="form-row dual">
          <div class="form-group">
            <label>Warning Threshold</label>
            <input 
              type="text" 
              v-model="config.warningThreshold" 
              placeholder="e.g., >80"
              class="form-input"
            />
          </div>
          <div class="form-group">
            <label>Critical Threshold</label>
            <input 
              type="text" 
              v-model="config.criticalThreshold" 
              placeholder="e.g., >90"
              class="form-input"
            />
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
            Test SNMP Query
          </button>
          <button class="btn btn-danger" @click="resetForm">
            <span class="btn-icon">↺</span>
            Reset
          </button>
        </div>
      </div>
    </div>

    <!-- Active SNMP Monitors List -->
    <div class="monitors-list-card">
      <div class="card-header">
        <div class="header-left">
          <h2>Active SNMP Monitors</h2>
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
              :class="{ active: sortBy === 'oid' }"
              @click="toggleSort('oid')"
              title="Sort by OID"
            >
              <span>Sort by OID</span>
              <span class="sort-icon">{{ getSortIcon('oid') }}</span>
            </button>
          </div>
        </div>
        <div class="header-actions">
          <span class="monitor-count">
            <span class="online-count">🟢 {{ onlineCount }}</span>
            <span class="offline-count">🔴 {{ offlineCount }}</span>
            <span class="warning-count">⚠️ {{ warningCount }}</span>
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
                <th>Host:Port</th>
                <th>Community</th>
                <th @click="toggleSort('oid')" class="sortable-header">
                  OID
                  <span class="sort-icon">{{ getSortIcon('oid') }}</span>
                </th>
                <th>Value</th>
                <th>Version</th>
                <th>Last Check</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="monitor in sortedMonitors" :key="monitor.id" class="monitor-row">
                <td data-label="Status">
                  <span class="status-badge" :class="monitor.status">
                    {{ getStatusIcon(monitor.status) }}
                  </span>
                </td>
                <td data-label="Friendly Name" class="friendly-name">{{ monitor.friendlyName }}</td>
                <td data-label="Host:Port" class="host-port">{{ monitor.hostname }}:{{ monitor.port }}</td>
                <td data-label="Community" class="community">{{ maskCommunity(monitor.community) }}</td>
                <td data-label="OID" class="oid">{{ monitor.oid }}</td>
                <td data-label="Value" :class="getValueClass(monitor)">
                  {{ monitor.lastValue || '—' }}
                  <span v-if="monitor.unit" class="unit">{{ monitor.unit }}</span>
                </td>
                <td data-label="Version">{{ monitor.snmpVersion }}</td>
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
                <td colspan="9" class="empty-state">
                  <div class="empty-icon">📡</div>
                  <p>No SNMP monitors configured yet</p>
                  <p class="empty-hint">Create your first SNMP monitor using the form above</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Test SNMP Query Modal -->
    <div v-if="showTestModal" class="modal-overlay" @click.self="closeTestModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>SNMP Query Test Results</h3>
          <button class="close-btn" @click="closeTestModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="test-result" :class="testResult.status">
            <div class="result-icon">
              {{ testResult.status === 'success' ? '✅' : '❌' }}
            </div>
            <div class="result-details">
              <h4>{{ testResult.status === 'success' ? 'SNMP Query Successful' : 'SNMP Query Failed' }}</h4>
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
                  <span class="meta-label">OID:</span>
                  <span class="meta-value">{{ testResult.details.oid }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">Value:</span>
                  <span class="meta-value">{{ testResult.details.value }}</span>
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
  name: 'SNMPMonitor',
  
  setup() {
    // Configuration state
    const config = reactive({
      friendlyName: 'New Monitor',
      hostname: '',
      port: 161,
      community: 'public',
      oid: '1.3.6.1.4.1.9.6.1.101',
      snmpVersion: 'v2c',
      interval: 60,
      timeout: 5,
      retries: 2,
      valueType: 'integer',
      warningThreshold: '',
      criticalThreshold: '',
      // SNMPv3 fields
      securityName: '',
      securityLevel: 'noAuthNoPriv',
      authProtocol: 'MD5',
      authPassword: '',
      privProtocol: 'AES',
      privPassword: ''
    })

    // Sorting state
    const sortBy = ref('status')
    const sortDirection = ref('desc')

    // Monitors list with sample data
    const monitors = ref([
      {
        id: 1,
        friendlyName: 'Core Switch CPU Load',
        hostname: '192.168.1.1',
        port: 161,
        community: 'public',
        oid: '1.3.6.1.4.1.9.2.1.56.0',
        snmpVersion: 'v2c',
        status: 'warning',
        lastValue: '78',
        unit: '%',
        lastCheck: new Date(Date.now() - 45000).toISOString()
      },
      {
        id: 2,
        friendlyName: 'Router Uptime',
        hostname: '192.168.1.254',
        port: 161,
        community: 'monitor',
        oid: '1.3.6.1.2.1.1.3.0',
        snmpVersion: 'v2c',
        status: 'online',
        lastValue: '3456789',
        unit: 'ticks',
        lastCheck: new Date(Date.now() - 120000).toISOString()
      },
      {
        id: 3,
        friendlyName: 'Interface Status - Port 1',
        hostname: '192.168.1.1',
        port: 161,
        community: 'public',
        oid: '1.3.6.1.2.1.2.2.1.8.1',
        snmpVersion: 'v1',
        status: 'online',
        lastValue: '1',
        unit: '',
        lastCheck: new Date(Date.now() - 30000).toISOString()
      },
      {
        id: 4,
        friendlyName: 'Temperature Sensor',
        hostname: '192.168.1.100',
        port: 161,
        community: 'private',
        oid: '1.3.6.1.4.1.9.9.13.1.3.1.3.1',
        snmpVersion: 'v2c',
        status: 'critical',
        lastValue: '82',
        unit: '°C',
        lastCheck: new Date(Date.now() - 180000).toISOString()
      },
      {
        id: 5,
        friendlyName: 'Memory Usage',
        hostname: '192.168.1.1',
        port: 161,
        community: 'public',
        oid: '1.3.6.1.4.1.9.9.48.1.1.1.6.1',
        snmpVersion: 'v3',
        status: 'offline',
        lastValue: null,
        unit: '%',
        lastCheck: new Date(Date.now() - 360000).toISOString()
      }
    ])

    // Computed properties for counts
    const onlineCount = computed(() => {
      return monitors.value.filter(m => m.status === 'online').length
    })

    const offlineCount = computed(() => {
      return monitors.value.filter(m => m.status === 'offline').length
    })

    const warningCount = computed(() => {
      return monitors.value.filter(m => m.status === 'warning').length
    })

    const criticalCount = computed(() => {
      return monitors.value.filter(m => m.status === 'critical').length
    })

    // Sorted monitors computed property
    const sortedMonitors = computed(() => {
      const sorted = [...monitors.value]
      
      sorted.sort((a, b) => {
        let comparison = 0
        
        switch (sortBy.value) {
          case 'status':
            const statusWeight = { 'critical': 1, 'warning': 2, 'offline': 3, 'online': 4 }
            comparison = (statusWeight[a.status] || 5) - (statusWeight[b.status] || 5)
            break
          case 'name':
            comparison = a.friendlyName.localeCompare(b.friendlyName)
            break
          case 'oid':
            comparison = a.oid.localeCompare(b.oid)
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
      message: 'Successfully queried SNMP device',
      details: {
        host: '',
        port: 161,
        oid: '',
        value: '78 %',
        responseTime: 45
      }
    })

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

    const getStatusIcon = (status) => {
      const icons = {
        'online': '🟢',
        'offline': '🔴',
        'warning': '⚠️',
        'critical': '🔥',
        'pending': '⏳'
      }
      return icons[status] || '⚪'
    }

    const getValueClass = (monitor) => {
      if (monitor.status === 'critical') return 'value-critical'
      if (monitor.status === 'warning') return 'value-warning'
      if (monitor.status === 'online') return 'value-normal'
      return 'value-offline'
    }

    const maskCommunity = (community) => {
      if (!community) return '—'
      if (community.length <= 3) return '*'.repeat(community.length)
      return community.substring(0, 1) + '*'.repeat(community.length - 2) + community.substring(community.length - 1)
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
      if (!config.oid) {
        alert('Please enter an OID')
        return
      }

      const newMonitor = {
        id: Date.now(),
        friendlyName: config.friendlyName || 'New Monitor',
        hostname: config.hostname,
        port: config.port,
        community: config.community,
        oid: config.oid,
        snmpVersion: config.snmpVersion,
        status: 'pending',
        lastValue: null,
        lastCheck: new Date().toISOString()
      }

      monitors.value.push(newMonitor)
      resetForm()
    }

    const testConnection = () => {
      const success = Math.random() > 0.3
      testResult.value = {
        status: success ? 'success' : 'error',
        message: success 
          ? 'Successfully queried SNMP device'
          : 'SNMP query failed: No response from device',
        details: {
          host: config.hostname || '192.168.1.1',
          port: config.port || 161,
          oid: config.oid || '1.3.6.1.4.1.9.6.1.101',
          value: success ? '78 %' : '—',
          responseTime: success ? Math.floor(Math.random() * 100) + 20 : 0
        }
      }
      showTestModal.value = true
    }

    const resetForm = () => {
      config.friendlyName = 'New Monitor'
      config.hostname = ''
      config.port = 161
      config.community = 'public'
      config.oid = '1.3.6.1.4.1.9.6.1.101'
      config.snmpVersion = 'v2c'
      config.interval = 60
      config.timeout = 5
      config.retries = 2
      config.valueType = 'integer'
      config.warningThreshold = ''
      config.criticalThreshold = ''
      config.securityName = ''
      config.securityLevel = 'noAuthNoPriv'
      config.authProtocol = 'MD5'
      config.authPassword = ''
      config.privProtocol = 'AES'
      config.privPassword = ''
    }

    const editMonitor = (monitor) => {
      config.friendlyName = monitor.friendlyName
      config.hostname = monitor.hostname
      config.port = monitor.port
      config.community = monitor.community
      config.oid = monitor.oid
      config.snmpVersion = monitor.snmpVersion
      
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
        status: ['online', 'offline', 'warning', 'critical'][Math.floor(Math.random() * 4)],
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
      warningCount,
      criticalCount,
      showTestModal,
      testResult,
      sortBy,
      sortDirection,
      toggleSort,
      getSortIcon,
      getStatusIcon,
      getValueClass,
      maskCommunity,
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
.snmp-monitor {
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
  flex-wrap: wrap;
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

.warning-count {
  color: #fbbf24;
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

.input-hint {
  font-size: 0.8rem;
  color: #64748b;
  line-height: 1.4;
  margin-top: 4px;
}

/* Select wrapper */
.select-wrapper {
  position: relative;
  width: 100%;
}

.form-select {
  appearance: none;
  width: 100%;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 10px 14px;
  color: #e2e8f0;
  font-size: 0.95rem;
  cursor: pointer;
}

.select-arrow {
  position: absolute;
  right: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: #64748b;
  pointer-events: none;
  font-size: 0.8rem;
}

/* V3 fields */
.v3-fields {
  margin-top: 16px;
  padding: 16px;
  background: rgba(15, 23, 42, 0.6);
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.1);
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
  min-width: 1000px;
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

.community {
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  color: #94a3b8;
}

.oid {
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  color: #c4b5fd;
  font-size: 0.9rem;
}

.value-normal {
  color: #34d399;
  font-weight: 600;
}

.value-warning {
  color: #fbbf24;
  font-weight: 600;
}

.value-critical {
  color: #ef4444;
  font-weight: 600;
}

.value-offline {
  color: #94a3b8;
}

.unit {
  font-size: 0.8rem;
  color: #64748b;
  margin-left: 2px;
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
  width: 500px;
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
  .snmp-monitor {
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
    min-width: 800px;
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