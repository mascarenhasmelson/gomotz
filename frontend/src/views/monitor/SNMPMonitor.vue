<template>
  <div class="snmp-monitor">
    <!-- SNMP Configuration Form -->
    <div class="config-card">
      <div class="card-header">
        <h2>{{ editingMonitorId ? 'Edit SNMP Monitor' : 'SNMP Monitor Configuration' }}</h2>
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
              placeholder="e.g., Router Uptime"
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
              v-model="config.community_string" 
              placeholder="public"
              class="form-input"
            />
            <span class="input-hint">This string functions as a password to authenticate SNMP requests</span>
          </div>
        </div>

        <!-- OID -->
        <div class="form-row">
          <div class="form-group">
            <label>OID (Object Identifier)</label>
            <input 
              type="text" 
              v-model="config.oid" 
              placeholder="1.3.6.1.2.1.1.3.0"
              class="form-input"
            />
            <span class="input-hint">Enter the OID you want to monitor</span>
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>SNMP Version</label>
            <div class="select-wrapper">
              <select v-model="config.snmp_version" class="form-select">
                <option value="v1">SNMPv1</option>
                <option value="v2c">SNMPv2c</option>
                <!-- <option value="v3">SNMPv3</option> -->
              </select>
              <span class="select-arrow">▼</span>
            </div>
          </div>
        </div>

        <!-- SNMPv3 specific fields -->
        <div v-if="config.snmp_version === 'v3'" class="v3-fields">
          <div class="form-row dual">
            <div class="form-group">
              <label>Security Name</label>
              <input type="text" v-model="config.security_name" placeholder="username" class="form-input"/>
            </div>
            <div class="form-group">
              <label>Security Level</label>
              <select v-model="config.security_level" class="form-select">
                <option value="noAuthNoPriv">noAuthNoPriv</option>
                <option value="authNoPriv">authNoPriv</option>
                <option value="authPriv">authPriv</option>
              </select>
            </div>
          </div>
          
          <div v-if="config.security_level !== 'noAuthNoPriv'" class="form-row dual">
            <div class="form-group">
              <label>Auth Protocol</label>
              <select v-model="config.auth_protocol" class="form-select">
                <option value="MD5">MD5</option>
                <option value="SHA">SHA</option>
              </select>
            </div>
            <div class="form-group">
              <label>Auth Password</label>
              <input type="password" v-model="config.auth_password" placeholder="••••••••" class="form-input"/>
            </div>
          </div>
          
          <div v-if="config.security_level === 'authPriv'" class="form-row dual">
            <div class="form-group">
              <label>Privacy Protocol</label>
              <select v-model="config.priv_protocol" class="form-select">
                <option value="DES">DES</option>
                <option value="AES">AES</option>
              </select>
            </div>
            <div class="form-group">
              <label>Privacy Password</label>
              <input type="password" v-model="config.priv_password" placeholder="••••••••" class="form-input"/>
            </div>
          </div>
        </div>

        <!-- Polling Interval -->
        <div class="form-row">
          <div class="form-group">
            <label>Polling Interval</label>
            <div class="interval-input">
              <input type="number" v-model="config.polling_interval" min="5" class="form-input interval-field"/>
              <span class="interval-unit">seconds</span>
            </div>
          </div>
        </div>

        <!-- Timeout and Retries -->
        <div class="form-row dual">
          <div class="form-group">
            <label>Timeout</label>
            <input type="number" v-model="config.timeout" min="1" class="form-input"/>
          </div>
          <div class="form-group">
            <label>Retries</label>
            <input type="number" v-model="config.retries" min="0" max="10" class="form-input"/>
          </div>
        </div>

        <!-- Expected Value Type -->
        <div class="form-row">
          <div class="form-group">
            <label>Expected Value Type</label>
            <select v-model="config.expected_value_type" class="form-select">
              <option value="integer">Integer</option>
              <option value="string">String</option>
              <option value="oid">OID</option>
              <option value="counter">Counter</option>
              <option value="gauge">Gauge</option>
              <option value="timeticks">TimeTicks</option>
            </select>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="form-actions">
          <button class="btn btn-primary" @click="saveMonitor" :disabled="isLoading">
            <span class="btn-icon">💾</span>
            {{ isLoading ? 'Saving...' : (editingMonitorId ? 'Update Monitor' : 'Save Monitor') }}
          </button>
          <button class="btn btn-secondary" @click="testConnection" :disabled="isLoading">
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

    <!-- WebSocket Connection Status -->
    <div class="ws-status" :class="wsConnected ? 'connected' : 'disconnected'">
      <span class="ws-indicator"></span>
      <span class="ws-text">{{ wsConnected ? 'Real-time updates active' : 'Connecting to real-time updates...' }}</span>
    </div>

    <!-- Active SNMP Monitors List -->
    <div class="monitors-list-card">
      <div class="card-header">
        <div class="header-left">
          <h2>Active SNMP Monitors</h2>
          <div class="sort-controls">
            <button class="sort-btn" :class="{ active: sortBy === 'status' }" @click="toggleSort('status')">
              <span>Status</span>
              <span class="sort-icon">{{ getSortIcon('status') }}</span>
            </button>
            <button class="sort-btn" :class="{ active: sortBy === 'friendly_name' }" @click="toggleSort('friendly_name')">
              <span>Name</span>
              <span class="sort-icon">{{ getSortIcon('friendly_name') }}</span>
            </button>
            <button class="sort-btn" :class="{ active: sortBy === 'oid' }" @click="toggleSort('oid')">
              <span>OID</span>
              <span class="sort-icon">{{ getSortIcon('oid') }}</span>
            </button>
          </div>
        </div>
        <div class="header-actions">
          <span class="monitor-count">
            <span class="online-count">🟢 Up: {{ upCount }}</span>
            <span class="offline-count">🔴 Down: {{ downCount }}</span>
            <span class="pending-count">⏳ Pending: {{ pendingCount }}</span>
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
                <th @click="toggleSort('status')" class="sortable-header">Status</th>
                <th @click="toggleSort('friendly_name')" class="sortable-header">Friendly Name</th>
                <th>Host:Port</th>
                <th>Community</th>
                <th @click="toggleSort('oid')" class="sortable-header">OID</th>
                <th>Value</th>
                <th>Version</th>
                <th>Last Check</th>
                <th>Response Time</th>
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
                <td data-label="Friendly Name">{{ monitor.friendly_name }}</td>
                <td data-label="Host:Port">{{ monitor.hostname }}:{{ monitor.port }}</td>
                <td data-label="Community">{{ maskCommunity(monitor.community_string) }}</td>
                <td data-label="OID" class="oid">{{ monitor.oid }}</td>
                <td data-label="Value" :class="getValueClass(monitor)">
                  {{ monitor.last_value || '—' }}
                </td>
                <td data-label="Version">{{ monitor.snmp_version }}</td>
                <td data-label="Last Check">{{ formatLastCheck(monitor) }}</td>
                <td data-label="Response Time" class="response-time-cell">
                  <span v-if="monitor.last_response_ms" class="response-time" :class="getResponseTimeClass(monitor.last_response_ms)">
                    {{ monitor.last_response_ms }}ms
                  </span>
                  <span v-else class="response-time na">—</span>
                </td>
                <td data-label="Actions" class="actions">
                  <!-- <button class="action-btn" @click="editMonitor(monitor)" title="Edit">✏️</button> -->
                  <button class="action-btn" @click="deleteMonitor(monitor)" title="Delete">🗑️</button>
                </td>
              </tr>
              <tr v-if="monitors.length === 0 && !isLoading">
                <td colspan="10" class="empty-state">
                  <div class="empty-icon">   </div>
                  <p>No SNMP monitors configured yet</p>
                  <p class="empty-hint">Create your first SNMP monitor using the form above</p>
                </td>
              </tr>
              <tr v-if="isLoading && monitors.length === 0">
                <td colspan="10" class="loading-state">
                  <div class="loading-spinner"></div>
                  <p>Loading monitors...</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Delete SNMP Monitor</h3>
          <button class="close-btn" @click="showDeleteModal = false">&times;</button>
        </div>
        <div class="modal-body">
          <p>Are you sure you want to delete <strong>"{{ selectedMonitor?.friendly_name }}"</strong>?</p>
          <p class="warning-text">This action cannot be undone.</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteModal = false">Cancel</button>
          <button class="btn btn-danger" @click="confirmDelete" :disabled="deleting">
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </button>
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
            <div class="result-icon">{{ testResult.status === 'success' ? '✅' : '❌' }}</div>
            <div class="result-details">
              <h4>{{ testResult.status === 'success' ? 'SNMP Query Successful' : 'SNMP Query Failed' }}</h4>
              <p>{{ testResult.message }}</p>
              <div class="result-meta" v-if="testResult.details">
                <div class="meta-item"><span class="meta-label">Host:</span><span class="meta-value">{{ testResult.details.host }}</span></div>
                <div class="meta-item"><span class="meta-label">Port:</span><span class="meta-value">{{ testResult.details.port }}</span></div>
                <div class="meta-item"><span class="meta-label">OID:</span><span class="meta-value">{{ testResult.details.oid }}</span></div>
                <div class="meta-item"><span class="meta-label">Value:</span><span class="meta-value">{{ testResult.details.value }}</span></div>
                <div class="meta-item"><span class="meta-label">Response Time:</span><span class="meta-value">{{ testResult.details.responseTime }}ms</span></div>
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
  name: 'SNMPMonitor',
  
  setup() {
    const monitors = ref([])
    const isLoading = ref(false)
    const deleting = ref(false)
    const editingMonitorId = ref(null)
    const ws = ref(null)
    const wsConnected = ref(false)
    const showDeleteModal = ref(false)
    const selectedMonitor = ref(null)
    
    const config = reactive({
      friendly_name: '',
      hostname: '',
      port: 161,
      community_string: 'public',
      oid: '1.3.6.1.2.1.1.3.0',
      snmp_version: 'v2c',
      polling_interval: 60,
      timeout: 5,
      retries: 2,
      expected_value_type: 'timeticks',
      security_name: '',
      security_level: 'noAuthNoPriv',
      auth_protocol: 'MD5',
      auth_password: '',
      priv_protocol: 'AES',
      priv_password: ''
    })

    const sortBy = ref('status')
    const sortDirection = ref('desc')
    const showTestModal = ref(false)
    const testResult = ref({
      status: 'success',
      message: '',
      details: { host: '', port: 161, oid: '', value: '', responseTime: 0 }
    })

    const upCount = computed(() => monitors.value.filter(m => m.status === 'up').length)
    const downCount = computed(() => monitors.value.filter(m => m.status === 'down').length)
    const pendingCount = computed(() => monitors.value.filter(m => m.status === 'pending').length)

    const sortedMonitors = computed(() => {
      const sorted = [...monitors.value]
      sorted.sort((a, b) => {
        let comparison = 0
        switch (sortBy.value) {
          case 'status':
            const statusWeight = { 'up': 1, 'pending': 2, 'down': 3 }
            comparison = (statusWeight[a.status] || 4) - (statusWeight[b.status] || 4)
            break
          case 'friendly_name': 
            comparison = (a.friendly_name || '').localeCompare(b.friendly_name || '')
            break
          case 'oid': 
            comparison = (a.oid || '').localeCompare(b.oid || '')
            break
          default: return 0
        }
        return sortDirection.value === 'asc' ? comparison : -comparison
      })
      return sorted
    })
    const capitalizeValueType = (type) => {
      const map = {
        'integer': 'Integer',
        'string': 'String',
        'oid': 'OID',
        'counter': 'Counter',
        'gauge': 'Gauge',
        'timeticks': 'TimeTicks',
      }
      return map[type?.toLowerCase()] || type
    }
    const connectWebSocket = () => {
      const wsUrl = `${WS_BASE_URL}/v1/api/ws/snmp`
      ws.value = new WebSocket(wsUrl)

      ws.value.onopen = () => {
        console.log('  SNMP WebSocket connected')
        wsConnected.value = true
      }

      ws.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('📨 WebSocket message:', data.type)
          if (data.type === 'initial_state') {
            monitors.value = (data.monitors || []).map(m => ({
              ...m,
              status: m.status || 'pending',
              last_check: m.last_checked_at || null,
              last_value: m.last_value || null,
              last_response_ms: m.last_response_ms || null
            }))
            console.log(`Loaded ${monitors.value.length} monitors via WebSocket`)
          } 
          else if (data.type === 'snmp_monitor_update') {
            const idx = monitors.value.findIndex(m => m.id === data.monitor_id)
            if (idx !== -1) {
              monitors.value[idx] = {
                ...monitors.value[idx],
                status: data.status,
                last_value: data.value,
                last_response_ms: data.response_ms,
                last_checked_at: data.checked_at,
                last_check: data.checked_at
              }
              console.log(`Updated monitor ${data.monitor_id}: ${data.status}`)
            }
          } 
          else if (data.type === 'snmp_monitor_created') {
            monitors.value.push({
              ...data.monitor,
              status: data.monitor.status || 'pending',
              last_check: data.monitor.last_checked_at || null,
              last_value: data.monitor.last_value || null,
              last_response_ms: data.monitor.last_response_ms || null
            })
            console.log(`Monitor created: ${data.monitor.friendly_name}`)
          } 
          else if (data.type === 'snmp_monitor_updated') {
            const idx = monitors.value.findIndex(m => m.id === data.monitor.id)
            if (idx !== -1) {
              monitors.value[idx] = {
                ...data.monitor,
                status: data.monitor.status || 'pending',
                last_check: data.monitor.last_checked_at || null,
                last_value: data.monitor.last_value || null,
                last_response_ms: data.monitor.last_response_ms || null
              }
              console.log(`Monitor updated: ${data.monitor.friendly_name}`)
            }
          } 
          else if (data.type === 'snmp_monitor_deleted') {
            monitors.value = monitors.value.filter(m => m.id !== data.monitor_id)
            console.log(`Monitor deleted: ${data.monitor_id}`)
          }
        } catch (e) { 
          console.error('WebSocket parse error:', e) 
        }
      }

      ws.value.onclose = () => {
        console.log('   SNMP WebSocket disconnected')
        wsConnected.value = false
        setTimeout(connectWebSocket, 3000)
      }
      ws.value.onerror = (err) => console.error('WebSocket error:', err)
    }

    const fetchMonitors = async () => {
      isLoading.value = true
      try {
        const response = await fetch(`${API_BASE_URL}/v1/api/snmp`)
        if (!response.ok) throw new Error(`HTTP ${response.status}`)
        const data = await response.json()
        monitors.value = (data || []).map(m => ({
          ...m,
          status: m.status || 'pending',
          last_check: m.last_checked_at || null,
          last_value: m.last_value || null,
          last_response_ms: m.last_response_ms || null
        }))
        console.log('Fetched monitors:', monitors.value.length)
      } catch (error) {
        console.error('Error fetching monitors:', error)
      } finally { isLoading.value = false }
    }

    const saveMonitor = async () => {
      if (!config.friendly_name) { alert('Please enter a friendly name'); return }
      if (!config.hostname) { alert('Please enter a hostname'); return }
      if (!config.port || config.port < 1 || config.port > 65535) { alert('Please enter a valid port'); return }
      if (!config.oid) { alert('Please enter an OID'); return }

      isLoading.value = true
      try {
        const payload = {
          friendly_name: config.friendly_name,
          hostname: config.hostname,
          port: parseInt(config.port),
          community_string: config.community_string,
          oid: config.oid,
          snmp_version: config.snmp_version,
          polling_interval: parseInt(config.polling_interval),
          timeout: parseInt(config.timeout),
          retries: parseInt(config.retries),
          expected_value_type: capitalizeValueType(config.expected_value_type),
        }

        if (config.snmp_version === 'v3') {
          payload.security_name = config.security_name
          payload.security_level = config.security_level
          if (config.security_level !== 'noAuthNoPriv') payload.auth_protocol = config.auth_protocol
          if (config.auth_password) payload.auth_password = config.auth_password
          if (config.security_level === 'authPriv') payload.priv_protocol = config.priv_protocol
          if (config.priv_password) payload.priv_password = config.priv_password
        }

        let response
        if (editingMonitorId.value) {
          response = await fetch(`${API_BASE_URL}/v1/api/snmp/${editingMonitorId.value}`, {
            method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload)
          })
        } else {
          response = await fetch(`${API_BASE_URL}/v1/api/snmp`, {
            method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload)
          })
        }

        if (!response.ok) throw new Error(`HTTP ${response.status}`)
        alert(editingMonitorId.value ? 'Monitor updated successfully' : 'Monitor created successfully')
        resetForm()
        if (!wsConnected.value) await fetchMonitors()
      } catch (error) {
        console.error('Error saving monitor:', error)
        alert(`Failed to save monitor: ${error.message}`)
      } finally { isLoading.value = false }
    }

    const deleteMonitor = (monitor) => {
      selectedMonitor.value = monitor
      showDeleteModal.value = true
    }

    const confirmDelete = async () => {
      if (!selectedMonitor.value) return
      deleting.value = true
      try {
        const response = await fetch(`${API_BASE_URL}/v1/api/snmp/${selectedMonitor.value.id}`, { method: 'DELETE' })
        if (!response.ok) throw new Error(`HTTP ${response.status}`)
        alert(`Monitor "${selectedMonitor.value.friendly_name}" deleted successfully`)
        showDeleteModal.value = false
        if (!wsConnected.value) await fetchMonitors()
        if (editingMonitorId.value === selectedMonitor.value.id) resetForm()
      } catch (error) {
        console.error('Error deleting monitor:', error)
        alert(`Failed to delete monitor: ${error.message}`)
      } finally { deleting.value = false; selectedMonitor.value = null }
    }

    const editMonitor = (monitor) => {
      editingMonitorId.value = monitor.id
      config.friendly_name = monitor.friendly_name
      config.hostname = monitor.hostname
      config.port = monitor.port
      config.community_string = monitor.community_string || 'public'
      config.oid = monitor.oid
      config.snmp_version = monitor.snmp_version
      config.polling_interval = monitor.polling_interval
      config.timeout = monitor.timeout
      config.retries = monitor.retries
      config.expected_value_type = monitor.expected_value_type ? monitor.expected_value_type.toLowerCase() : 'timeticks'
      document.querySelector('.config-card').scrollIntoView({ behavior: 'smooth' })
    }

    const cancelEdit = () => resetForm()
    const testConnection = async () => {
      if (!config.hostname) { alert('Please enter a hostname'); return }
      if (!config.oid) { alert('Please enter an OID'); return }

      showTestModal.value = true
      testResult.value = { status: 'testing', message: 'Testing SNMP query...', details: {} }

      try {
        const payload = {
          hostname: config.hostname,
          port: parseInt(config.port),
          community_string: config.community_string,
          oid: config.oid,
          snmp_version: config.snmp_version,
          timeout: parseInt(config.timeout) || 5
        }

        const response = await fetch(`${API_BASE_URL}/v1/api/snmp/test`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        })

        const data = await response.json()
        if (data.success) {
          testResult.value = {
            status: 'success',
            message: 'SNMP query successful',
            details: {
              host: data.hostname || config.hostname,
              port: data.port || config.port,
              oid: data.oid || config.oid,
              value: data.value || 'N/A',
              responseTime: data.response_ms || 0
            }
          }
        } else {
          testResult.value = {
            status: 'error',
            message: data.error || 'SNMP query failed',
            details: {
              host: config.hostname,
              port: config.port,
              oid: config.oid,
              value: '—',
              responseTime: 0
            }
          }
        }
      } catch (error) {
        testResult.value = {
          status: 'error',
          message: error.message || 'Connection failed',
          details: { host: config.hostname, port: config.port, oid: config.oid, value: '—', responseTime: 0 }
        }
      }
    }

    const resetForm = () => {
      editingMonitorId.value = null
      config.friendly_name = ''
      config.hostname = ''
      config.port = 161
      config.community_string = 'public'
      config.oid = '1.3.6.1.2.1.1.3.0'
      config.snmp_version = 'v2c'
      config.polling_interval = 60
      config.timeout = 5
      config.retries = 2
      config.expected_value_type = 'timeticks'
      config.security_name = ''
      config.security_level = 'noAuthNoPriv'
      config.auth_protocol = 'MD5'
      config.auth_password = ''
      config.priv_protocol = 'AES'
      config.priv_password = ''
    }

    const maskCommunity = (community) => {
      if (!community) return '—'
      if (community.length <= 3) return '*'.repeat(community.length)
      return community[0] + '*'.repeat(community.length - 2) + community[community.length - 1]
    }
    const formatLastCheck = (monitor) => {
      const timestamp = monitor.last_checked_at || monitor.last_check
      if (!timestamp) return 'Never'
      const date = new Date(timestamp)
      const now = new Date()
      const diffMs = now - date
      const diffSecs = Math.floor(diffMs / 1000)
      const diffMins = Math.floor(diffMs / 60000)
      if (diffSecs < 10) return 'Just now'
      if (diffSecs < 60) return `${diffSecs}s ago`
      if (diffMins < 60) return `${diffMins}m ago`
      return date.toLocaleTimeString()
    }

    const getStatusIcon = (status) => {
      const icons = { 'up': '🟢 UP', 'down': '🔴 DOWN', 'pending': '⏳ PENDING' }
      return icons[status] || '⚪ UNKNOWN'
    }

    const getValueClass = (monitor) => {
      if (monitor.status === 'up') return 'value-normal'
      if (monitor.status === 'pending') return 'value-pending'
      return 'value-offline'
    }

    const getResponseTimeClass = (ms) => {
      if (ms < 50) return 'fast'
      if (ms < 200) return 'normal'
      if (ms < 500) return 'slow'
      return 'very-slow'
    }

    const toggleSort = (field) => {
      if (sortBy.value === field) sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
      else { sortBy.value = field; sortDirection.value = 'asc' }
    }

    const getSortIcon = (field) => {
      if (sortBy.value !== field) return '↕️'
      return sortDirection.value === 'asc' ? '↑' : '↓'
    }

    const closeTestModal = () => { showTestModal.value = false }

    onMounted(() => {
      fetchMonitors()
      connectWebSocket()
    })

    onUnmounted(() => { if (ws.value) ws.value.close() })

    return {
      monitors, isLoading, deleting, editingMonitorId, config, upCount, downCount, pendingCount,
      sortedMonitors, showTestModal, testResult, wsConnected, showDeleteModal, selectedMonitor,
      formatLastCheck, maskCommunity, getStatusIcon, getValueClass, getResponseTimeClass,
      toggleSort, getSortIcon, fetchMonitors, saveMonitor, deleteMonitor, confirmDelete,
      editMonitor, cancelEdit, testConnection, resetForm, closeTestModal
    }
  }
}
</script>

<style scoped>
/* Keep all existing styles and add these new ones */

.pending-count {
  color: #fbbf24;
  display: flex;
  align-items: center;
  gap: 4px;
}

.status-badge.pending {
  background: rgba(245, 158, 11, 0.1);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.value-pending {
  color: #fbbf24;
  font-weight: 500;
}

.value-offline {
  color: #94a3b8;
}

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

/* Keep all other existing styles from the previous version */
.snmp-monitor {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  min-height: 100vh;
  color: #e2e8f0;
}

/* Cards */
.config-card, .monitors-list-card {
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

.sort-btn:hover, .sort-btn.active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: white;
}

.sort-icon { font-size: 0.9rem; }
.sortable-header { cursor: pointer; user-select: none; }
.sortable-header:hover { color: #60a5fa; }

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

.online-count { color: #34d399; display: flex; align-items: center; gap: 4px; }
.offline-count { color: #f87171; display: flex; align-items: center; gap: 4px; }
.total-count { color: #94a3b8; padding-left: 8px; border-left: 1px solid #334155; }

/* Form Styles */
.config-form { padding: 24px; }
.form-row { margin-bottom: 20px; }
.form-row.dual { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 0.9rem; font-weight: 500; color: #cbd5e1; }

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

.input-hint { font-size: 0.75rem; color: #64748b; margin-top: 4px; }

/* Select wrapper */
.select-wrapper { position: relative; width: 100%; }
.form-select {
  width: 100%;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 10px 14px;
  color: #e2e8f0;
  font-size: 0.95rem;
  cursor: pointer;
  appearance: none;
}
.select-arrow {
  position: absolute;
  right: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: #64748b;
  pointer-events: none;
}

/* V3 fields */
.v3-fields {
  margin-top: 16px;
  padding: 16px;
  background: rgba(15, 23, 42, 0.6);
  border-radius: 8px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.interval-input { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.interval-field { width: 100px; }
.interval-unit { color: #94a3b8; font-size: 0.9rem; }

/* Buttons */
.form-actions { display: flex; gap: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid rgba(148, 163, 184, 0.1); flex-wrap: wrap; }
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
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover:not(:disabled) { background: #2563eb; transform: translateY(-1px); box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3); }
.btn-secondary { background: #1e293b; color: #cbd5e1; border: 1px solid #334155; }
.btn-secondary:hover:not(:disabled) { background: #2d3748; transform: translateY(-1px); }
.btn-danger { background: #ef4444; color: white; }
.btn-danger:hover:not(:disabled) { background: #dc2626; transform: translateY(-1px); }
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
.btn-icon-only:hover:not(:disabled) { background: #2d3748; color: #60a5fa; transform: rotate(90deg); }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-icon { font-size: 1.1rem; }

/* Monitors Table */
.monitors-list-card { display: flex; flex-direction: column; max-height: 600px; }
.monitors-table-container { overflow-y: auto; overflow-x: auto; flex: 1; }
.monitors-table { min-width: 1100px; padding: 0 24px 24px 24px; }
.monitors-table table { width: 100%; border-collapse: collapse; }
.monitors-table table thead { position: sticky; top: 0; background: rgba(30, 41, 59, 0.95); z-index: 10; backdrop-filter: blur(4px); }
.monitors-table table thead th { background: rgba(30, 41, 59, 0.95); padding: 16px 8px 12px 8px; text-align: left; color: #94a3b8; font-weight: 500; font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.5px; }
.monitors-table table tbody td { padding: 12px 8px; border-bottom: 1px solid rgba(148, 163, 184, 0.05); color: #e2e8f0; }
.monitor-row:hover { background: rgba(59, 130, 246, 0.1); }

.status-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 600;
}
.status-badge.up { background: rgba(16, 185, 129, 0.1); color: #34d399; border: 1px solid rgba(16, 185, 129, 0.3); }
.status-badge.down { background: rgba(239, 68, 68, 0.1); color: #f87171; border: 1px solid rgba(239, 68, 68, 0.3); }

.oid { font-family: monospace; color: #c4b5fd; font-size: 0.85rem; }
.value-normal { color: #34d399; font-weight: 600; }
.actions { display: flex; gap: 6px; }
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
.action-btn:hover { background: #1e293b; color: #60a5fa; transform: scale(1.1); }

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
.ws-indicator { width: 8px; height: 8px; border-radius: 50%; animation: pulse 2s infinite; }
.ws-status.connected .ws-indicator { background: #10b981; box-shadow: 0 0 8px rgba(16, 185, 129, 0.5); }
.ws-status.disconnected .ws-indicator { background: #ef4444; box-shadow: 0 0 8px rgba(239, 68, 68, 0.5); }
.ws-status.connected .ws-text { color: #34d399; }
.ws-status.disconnected .ws-text { color: #f87171; }

/* Empty & Loading States */
.empty-state, .loading-state { text-align: center; padding: 60px 24px; color: #94a3b8; }
.empty-icon { font-size: 48px; margin-bottom: 16px; opacity: 0.5; animation: float 3s ease-in-out infinite; }
@keyframes float { 0%,100% { transform: translateY(0); } 50% { transform: translateY(-5px); } }
.loading-spinner { width: 40px; height: 40px; border: 3px solid #1e293b; border-top-color: #3b82f6; border-radius: 50%; animation: spin 1s linear infinite; margin: 0 auto 16px; }
@keyframes spin { to { transform: rotate(360deg); } }

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
.modal-header h3 { margin: 0; font-size: 1.1rem; color: #f8fafc; }
.close-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 24px;
  cursor: pointer;
  width: 32px;
  height: 32px;
  border-radius: 6px;
  transition: all 0.2s;
}
.close-btn:hover { background: #ef4444; color: white; }
.modal-body { padding: 20px; }
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px 20px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
}
.warning-text { color: #f87171; font-size: 0.85rem; margin-top: 8px; }

/* Test Result */
.test-result { display: flex; gap: 16px; padding: 16px; border-radius: 12px; }
.test-result.success { background: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.3); }
.test-result.error { background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); }
.result-icon { font-size: 32px; }
.result-details h4 { margin: 0 0 8px 0; font-size: 1rem; color: #f8fafc; }
.result-details p { margin: 0 0 12px 0; color: #94a3b8; font-size: 0.9rem; }
.result-meta { background: #0f172a; border-radius: 8px; padding: 12px; border: 1px solid #334155; }
.meta-item { display: flex; justify-content: space-between; margin-bottom: 6px; font-size: 0.9rem; }
.meta-label { color: #64748b; }
.meta-value { color: #e2e8f0; font-weight: 500; }

@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.5; } }

/* Responsive */
@media (max-width: 1024px) {
  .monitors-list-card { max-height: 400px; }
  .header-left { flex-direction: column; align-items: flex-start; gap: 12px; }
}
@media (max-width: 768px) {
  .snmp-monitor { padding: 16px; }
  .form-row.dual { grid-template-columns: 1fr; gap: 16px; }
  .form-actions { flex-direction: column; }
  .btn { width: 100%; justify-content: center; }
  .monitors-table { min-width: 800px; }
  .monitor-count { flex-direction: column; align-items: flex-start; gap: 8px; }
  .total-count { padding-left: 0; border-left: none; }
}
</style>