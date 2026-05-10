<template>
  <div class="domain-monitor">

    <div class="header-row">
      <h2>Domain Expiry Monitor</h2>
      <div class="header-actions">
        <button @click="exportCSV" class="btn btn-secondary">
          ⬇️ Download CSV
        </button>
        <button @click="openAddModal" class="btn btn-primary">
          ➕ Add Domain
        </button>
        <button @click="refreshData" class="btn btn-icon" title="Refresh">↻</button>
      </div>
    </div>


    <div class="connection-bar" :class="connectionStatus">
      <span class="dot"></span>
      <span>{{ connectionMessage }}</span>
      <button v-if="connectionStatus === 'disconnected'" @click="reconnectWebSocket" class="reconnect-btn">
        Reconnect
      </button>
    </div>


    <div class="stats-row">
      <div class="stat-card active">
        <span class="stat-count">{{ countByStatus('active') }}</span>
        <span class="stat-label">Active</span>
      </div>
      <div class="stat-card warning">
        <span class="stat-count">{{ countByStatus('warning') }}</span>
        <span class="stat-label">Warning</span>
      </div>
      <div class="stat-card critical">
        <span class="stat-count">{{ countByStatus('critical') }}</span>
        <span class="stat-label">Critical</span>
      </div>
      <div class="stat-card expired">
        <span class="stat-count">{{ countByStatus('expired') }}</span>
        <span class="stat-label">Expired</span>
      </div>
      <div class="stat-card error">
        <span class="stat-count">{{ countByStatus('error') }}</span>
        <span class="stat-label">Error</span>
      </div>
    </div>


    <div class="filter-row">
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Search domain or registrar..."
        class="search-input"
      />
      <select v-model="statusFilter" class="filter-select">
        <option value="">All Status</option>
        <option value="active">Active</option>
        <option value="warning">Warning</option>
        <option value="critical">Critical</option>
        <option value="expired">Expired</option>
        <option value="error">Error</option>
        <option value="pending">Pending</option>
      </select>
    </div>


    <div class="domains-grid">
      <div v-for="domain in filteredDomains" :key="domain.id" class="domain-card" :class="domain.status">

        <div class="card-header">
          <div class="header-left">
            <div class="status-indicator" :class="domain.status"></div>
            <div class="title-section">
              <h3 class="domain-name">{{ domain.friendly_name || domain.domain }}</h3>
              <span class="domain-host">{{ domain.domain }}</span>
            </div>
          </div>
          <div class="header-actions">
            <button class="icon-btn" @click="checkDomainNow(domain)" title="Check Now">
              🔄
            </button>
            <button class="icon-btn" @click="editDomain(domain)" title="Edit">
              ✏️
            </button>
            <button class="icon-btn delete" @click="confirmDelete(domain)" title="Delete">
              🗑️
            </button>
          </div>
        </div>


        <div class="expiry-progress">
          <div class="progress-bar">
            <div
              class="progress-fill"
              :class="domain.status"
              :style="{ width: getProgressPercent(domain) + '%' }"
            ></div>
          </div>
          <div class="expiry-days" :class="getExpiryClass(domain.days_remaining)">
            {{ domain.days_remaining ?? '?' }} days remaining
          </div>
        </div>


        <div class="expiry-date">
          <span class="label">Expires:</span>
          <span class="value">{{ formatDate(domain.expires_on) }}</span>
        </div>


        <div class="stats-grid">
          <div class="stat-item">
            <span class="stat-label">Registrar</span>
            <span class="stat-value">{{ domain.registrar || '—' }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Nameservers</span>
            <span class="stat-value">{{ domain.name_servers?.length || 0 }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Last Check</span>
            <span class="stat-value">{{ formatRelativeTime(domain.last_checked_at) }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Registered</span>
            <span class="stat-value">{{ formatDate(domain.registered_on) || '—' }}</span>
          </div>
        </div>
      </div>


      <div v-if="filteredDomains.length === 0 && !loading" class="empty-state">
        <div class="empty-icon">🌐</div>
        <h3>No Domains Added</h3>
        <p>Add your first domain to start monitoring expiry dates</p>
      </div>


      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <p>Loading domains...</p>
      </div>
    </div>


    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ isEditing ? 'Edit Domain' : 'Add Domain' }}</h3>
          <button @click="closeModal" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Domain Name <span class="required">*</span></label>
            <input 
              v-model="form.domain" 
              type="text"
              placeholder="e.g., example.com"
              class="form-input" 
              :disabled="isEditing"
            />
            <span class="hint">Enter the domain name without http:// or https://</span>
          </div>
          <div class="form-group">
            <label>Friendly Name</label>
            <input 
              v-model="form.friendly_name" 
              type="text"
              placeholder="e.g., My Website" 
              class="form-input"
            />
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Check Interval</label>
              <select v-model.number="form.check_interval" class="form-input">
                <option :value="86400">Daily</option>
                <option :value="43200">Every 12 hours</option>
                <option :value="21600">Every 6 hours</option>
                <option :value="3600">Every hour</option>
              </select>
            </div>
            <div class="form-group">
              <label>Warning (days)</label>
              <input 
                v-model.number="form.warning_days" 
                type="number" min="1" class="form-input"
              />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Critical (days)</label>
              <input 
                v-model.number="form.critical_days" 
                type="number" min="1" class="form-input"
              />
            </div>
            <div class="form-group">
              <label>Tags</label>
              <input 
                v-model="form.tags" 
                type="text"
                placeholder="production, critical" 
                class="form-input"
              />
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="testDomain" class="btn btn-secondary" :disabled="!form.domain || testing">
            {{ testing ? 'Testing...' : '🔍 Test Domain' }}
          </button>
          <button @click="saveDomain" class="btn btn-primary" :disabled="!form.domain || saving">
            {{ saving ? 'Saving...' : (isEditing ? 'Update' : 'Add Domain') }}
          </button>
        </div>


        <div v-if="testResult" class="test-result" :class="testResult.success ? 'success' : 'error'">
          <div v-if="testResult.success">
            <div class="tr-row"><span>Domain:</span><span>{{ testResult.domain }}</span></div>
            <div class="tr-row"><span>Expires:</span><span>{{ formatDate(testResult.expires_on) }}</span></div>
            <div class="tr-row"><span>Days:</span><span :class="getExpiryClass(testResult.days_remaining)">{{ testResult.days_remaining }} days</span></div>
            <div class="tr-row"><span>Registrar:</span><span>{{ testResult.registrar || 'Unknown' }}</span></div>
          </div>
          <div v-else class="error-msg">❌ {{ testResult.error }}</div>
        </div>
      </div>
    </div>


    <div v-if="selectedDomain" class="modal-overlay" @click.self="selectedDomain = null">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Domain Details — {{ selectedDomain.domain }}</h3>
          <button @click="selectedDomain = null" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div class="detail-grid">
            <div class="detail-item"><label>Status</label><span :class="selectedDomain.status">{{ getStatusLabel(selectedDomain.status) }}</span></div>
            <div class="detail-item"><label>Domain</label><span>{{ selectedDomain.domain }}</span></div>
            <div class="detail-item"><label>Friendly Name</label><span>{{ selectedDomain.friendly_name || '—' }}</span></div>
            <div class="detail-item"><label>Registrar</label><span>{{ selectedDomain.registrar || '—' }}</span></div>
            <div class="detail-item"><label>Registered</label><span>{{ formatDate(selectedDomain.registered_on) || '—' }}</span></div>
            <div class="detail-item"><label>Expires</label><span>{{ formatDate(selectedDomain.expires_on) }}</span></div>
            <div class="detail-item"><label>Days Remaining</label><span :class="getExpiryClass(selectedDomain.days_remaining)">{{ selectedDomain.days_remaining ?? '—' }} days</span></div>
            <div class="detail-item"><label>Last Checked</label><span>{{ formatRelativeTime(selectedDomain.last_checked_at) }}</span></div>
            <div class="detail-item full" v-if="selectedDomain.name_servers?.length"><label>Nameservers</label><span>{{ selectedDomain.name_servers.join(', ') }}</span></div>
            <div class="detail-item full" v-if="selectedDomain.error_message"><label>Error</label><span class="error-text">{{ selectedDomain.error_message }}</span></div>
          </div>
        </div>
      </div>
    </div>


    <div v-if="deleteTarget" class="modal-overlay" @click.self="deleteTarget = null">
      <div class="modal-content small">
        <div class="modal-header">
          <h3>Delete Domain</h3>
          <button @click="deleteTarget = null" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <p>Delete domain monitor for <strong>{{ deleteTarget.domain }}</strong>?</p>
          <p class="warning-text">This action cannot be undone.</p>
        </div>
        <div class="modal-footer">
          <button @click="performDelete" class="btn btn-danger">Delete</button>
          <button @click="deleteTarget = null" class="btn btn-secondary">Cancel</button>
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
  name: 'DomainMonitor',
  setup() {
    const domains = ref([])
    const loading = ref(true)
    const saving = ref(false)
    const testing = ref(false)
    const isEditing = ref(false)
    const editingId = ref(null)
    const showModal = ref(false)
    const selectedDomain = ref(null)
    const deleteTarget = ref(null)
    const testResult = ref(null)
    const searchQuery = ref('')
    const statusFilter = ref('')
    const sortBy = ref('days_remaining')
    const sortDir = ref('asc')
    const wsConnected = ref(false)
    const connectionStatus = ref('disconnected')
    const connectionMessage = ref('Disconnected')
    let ws = null
    let reconnectTimer = null
    let reconnectAttempts = ref(0)
    let pollInterval = null

    const form = reactive({
      domain: '',
      friendly_name: '',
      check_interval: 86400,
      warning_days: 30,
      critical_days: 15,
    })
    const filteredDomains = computed(() => {
      let list = [...domains.value]

      if (searchQuery.value) {
        const q = searchQuery.value.toLowerCase()
        list = list.filter(d =>
          (d.domain || '').toLowerCase().includes(q) ||
          (d.friendly_name || '').toLowerCase().includes(q) ||
          (d.registrar || '').toLowerCase().includes(q)
        )
      }

      if (statusFilter.value) {
        list = list.filter(d => d.status === statusFilter.value)
      }

      list.sort((a, b) => {
        let va = a[sortBy.value]
        let vb = b[sortBy.value]
        if (va === null || va === undefined) va = 999999
        if (vb === null || vb === undefined) vb = 999999
        return sortDir.value === 'asc' ? va - vb : vb - va
      })

      return list
    })

    const countByStatus = (status) => domains.value.filter(d => d.status === status).length
    const connectWebSocket = () => {
      if (ws && ws.readyState === WebSocket.OPEN) return

      const wsUrl = `${WS_BASE_URL}/v1/api/ws/domain`
      console.log('[WS] Connecting to:', wsUrl)

      connectionStatus.value = 'connecting'
      connectionMessage.value = 'Connecting...'

      ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        console.log('[WS] Domain WebSocket connected')
        wsConnected.value = true
        connectionStatus.value = 'connected'
        connectionMessage.value = 'Connected'
        reconnectAttempts.value = 0
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          console.log('[WS] Message received, type:', data.type, data)
          handleWebSocketMessage(data)
        } catch (e) {
          console.error('[WS] Parse error:', e)
        }
      }
      ws.onclose = () => {
        console.log('[WS] Domain WebSocket disconnected')
        wsConnected.value = false
        connectionStatus.value = 'disconnected'
        connectionMessage.value = 'Disconnected'

        if (reconnectAttempts.value < 10) {
          const delay = Math.min(3000 * Math.pow(2, reconnectAttempts.value), 30000)
          connectionMessage.value = `Reconnecting in ${Math.round(delay / 1000)}s...`
          reconnectTimer = setTimeout(() => {
            reconnectAttempts.value++
            connectWebSocket()
          }, delay)
        } else {
          connectionMessage.value = 'Failed to connect — refresh the page'
        }
      }

      ws.onerror = (error) => {
        console.error('[WS] Error:', error)
        connectionStatus.value = 'error'
        connectionMessage.value = 'Connection error'
      }
    }

    const reconnectWebSocket = () => {
      reconnectAttempts.value = 0
      if (reconnectTimer) clearTimeout(reconnectTimer)
      if (ws) ws.close()
      connectWebSocket()
    }
    const patchDomain = (id, fields) => {
      const idx = domains.value.findIndex(d => d.id === id)
      if (idx !== -1) {
        domains.value[idx] = { ...domains.value[idx], ...fields }
        console.log(`[WS] Patched domain id=${id}`, fields)
      } else {
        console.warn(`[WS] patchDomain: id=${id} not found in list`)
      }
    }

    const handleWebSocketMessage = (data) => {
      switch (data.type) {
        case 'initial_state':
          domains.value = data.domains || []
          loading.value = false
          console.log(`[WS] Loaded ${domains.value.length} domains`)
          break
        case 'domain_expiry_update':
          patchDomain(data.monitor_id, {
            status:          data.status,
            days_remaining:  data.days_remaining,
            registrar:       data.registrar       || undefined,
            expires_on:      data.expires_on       || undefined,
            last_checked_at: data.checked_at,
            error_message:   data.error            || null,
            warning_days:    data.warning_days,
            critical_days:   data.critical_days,
          })
          break

        case 'status_change':
          patchDomain(data.monitor_id, {
            status:          data.new_status,
            days_remaining:  data.days_remaining,
            expires_on:      data.expires_on       || undefined,
            registrar:       data.registrar        || undefined,
            last_checked_at: data.last_checked_at,
          })
          break

        default:
          console.log('[WS] Unhandled message type:', data.type, data)
      }
    }
    const fetchDomains = async () => {
      loading.value = true
      try {
        const res = await fetch(`${API_BASE_URL}/v1/api/domains`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        domains.value = await res.json() || []
        console.log(`[API] Fetched ${domains.value.length} domains`)
      } catch (e) {
        console.error('[API] Fetch error:', e)
      } finally {
        loading.value = false
      }
    }

    const refreshData = () => fetchDomains()

    const saveDomain = async () => {
      if (!form.domain) return
      saving.value = true
      try {
        const url = isEditing.value
          ? `${API_BASE_URL}/v1/api/domains/${editingId.value}`
          : `${API_BASE_URL}/v1/api/domains`
        const method = isEditing.value ? 'PUT' : 'POST'

        const payload = {
          domain:         form.domain,
          friendly_name:  form.friendly_name || form.domain,
          check_interval: form.check_interval,
          warning_days:   form.warning_days,
          critical_days:  form.critical_days,
        }

        const res = await fetch(url, {
          method,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        })

        if (!res.ok) {
          const err = await res.json()
          throw new Error(err.error || `HTTP ${res.status}`)
        }

        const saved = await res.json()

        if (isEditing.value) {
          patchDomain(saved.id, saved)
        } else {
          domains.value.unshift({
            ...saved,
            status: 'pending',
          })
        }

        closeModal()
      } catch (e) {
        console.error('[API] Save error:', e)
        alert(`Failed: ${e.message}`)
      } finally {
        saving.value = false
      }
    }

    const testDomain = async () => {
      if (!form.domain) return
      testing.value = true
      testResult.value = null
      try {
        const res = await fetch(`${API_BASE_URL}/v1/api/domains/test`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ domain: form.domain })
        })
        testResult.value = await res.json()
      } catch (e) {
        testResult.value = { success: false, error: e.message }
      } finally {
        testing.value = false
      }
    }
    const checkDomainNow = async (domain) => {
      try {
        const res = await fetch(
          `${API_BASE_URL}/v1/api/domains/${domain.id}/check`,
          { method: 'POST' }
        )
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        console.log(`[API] Check initiated for domain id=${domain.id}`)
      } catch (e) {
        console.error('[API] Check error:', e)
      }
    }

    const editDomain = (domain) => {
      isEditing.value = true
      editingId.value = domain.id
      form.domain = domain.domain
      form.friendly_name = domain.friendly_name || ''
      form.check_interval = domain.check_interval
      form.warning_days = domain.warning_days
      form.critical_days = domain.critical_days
      testResult.value = null
      showModal.value = true
    }

    const openAddModal = () => {
      isEditing.value = false
      editingId.value = null
      form.domain = ''
      form.friendly_name = ''
      form.check_interval = 86400
      form.warning_days = 30
      form.critical_days = 15
      testResult.value = null
      showModal.value = true
    }

    const confirmDelete = (domain) => {
      deleteTarget.value = domain
    }

    const performDelete = async () => {
      if (!deleteTarget.value) return
      const id = deleteTarget.value.id
      try {
        const res = await fetch(`${API_BASE_URL}/v1/api/domains/${id}`, {
          method: 'DELETE'
        })
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        domains.value = domains.value.filter(d => d.id !== id)
        deleteTarget.value = null
      } catch (e) {
        console.error('[API] Delete error:', e)
        alert(`Failed: ${e.message}`)
      }
    }

    const closeModal = () => {
      showModal.value = false
      isEditing.value = false
      editingId.value = null
      testResult.value = null
    }

    const viewDetails = (domain) => {
      selectedDomain.value = domain
    }
    const exportCSV = () => {
      const headers = ['ID', 'Domain', 'Friendly Name', 'Registrar', 'Expiry Date', 'Days Remaining', 'Status']
      const rows = domains.value.map(d => [
        d.id,
        d.domain,
        d.friendly_name || '',
        d.registrar || '',
        d.expires_on ? new Date(d.expires_on).toLocaleDateString() : '',
        d.days_remaining ?? '',
        d.status
      ])
      const csv = [headers, ...rows]
        .map(r => r.map(v => `"${String(v).replace(/"/g, '""')}"`).join(','))
        .join('\n')
      const blob = new Blob([csv], { type: 'text/csv' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `domains-${new Date().toISOString().slice(0, 10)}.csv`
      a.click()
      URL.revokeObjectURL(url)
    }

    const getStatusLabel = (status) => {
      const map = {
        active:  '✅ Active',
        warning: '⚠️ Warning',
        critical:'🔴 Critical',
        expired: '❌ Expired',
        error:   '⚠️ Error',
        pending: '⏳ Pending'
      }
      return map[status] || status
    }

    const formatDate = (ts) => {
      if (!ts) return '—'
      return new Date(ts).toLocaleDateString('en-US', {
        year: 'numeric', month: 'short', day: 'numeric'
      })
    }

    const formatRelativeTime = (ts) => {
      if (!ts) return 'Never'
      const diff = Math.floor((Date.now() - new Date(ts)) / 1000)
      if (diff < 60) return 'Just now'
      if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
      if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
      return `${Math.floor(diff / 86400)}d ago`
    }

    const getProgressPercent = (domain) => {
      const days = domain.days_remaining
      if (!days || days <= 0) return 0
      return Math.min(100, Math.round((days / 365) * 100))
    }

    const getExpiryClass = (days) => {
      if (!days || days <= 0) return 'expired'
      if (days <= 15) return 'critical'
      if (days <= 30) return 'warning'
      return 'active'
    }

    const toggleSort = (field) => {
      if (sortBy.value === field) {
        sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
      } else {
        sortBy.value = field
        sortDir.value = 'asc'
      }
    }

    const getSortIcon = (field) => {
      if (sortBy.value !== field) return '↕'
      return sortDir.value === 'asc' ? '↑' : '↓'
    }
    onMounted(() => {
      fetchDomains()
      connectWebSocket()

      // Fallback polling when WS is down
      pollInterval = setInterval(() => {
        if (!wsConnected.value) {
          console.log('[POLL] WS offline, polling REST API...')
          fetchDomains()
        }
      }, 30000)
    })

    onUnmounted(() => {
      if (ws) ws.close()
      if (reconnectTimer) clearTimeout(reconnectTimer)
      if (pollInterval) clearInterval(pollInterval)
    })

    return {
      domains, loading, saving, testing, isEditing, showModal,
      selectedDomain, deleteTarget, testResult, searchQuery, statusFilter,
      sortBy, sortDir, wsConnected, connectionStatus, connectionMessage,
      form, filteredDomains,
      countByStatus,
      fetchDomains, refreshData, saveDomain, testDomain, checkDomainNow,
      editDomain, openAddModal, confirmDelete, performDelete, closeModal,
      viewDetails, exportCSV, reconnectWebSocket,
      getStatusLabel, formatDate, formatRelativeTime, getProgressPercent,
      getExpiryClass, toggleSort, getSortIcon
    }
  }
}
</script>
<style scoped>
.domain-monitor {
  padding: 24px;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  min-height: 100vh;
  color: #e2e8f0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* Header */
.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}
.header-row h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  background: linear-gradient(135deg, #60a5fa, #a78bfa);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
.header-actions { display: flex; gap: 10px; align-items: center; }

/* Connection Bar */
.connection-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  margin-bottom: 20px;
  border-radius: 8px;
  font-size: 0.85rem;
}
.connection-bar.connected {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
  color: #34d399;
}
.connection-bar.disconnected, .connection-bar.error {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
}
.connection-bar.connecting {
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
  color: #fbbf24;
}
.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
  animation: pulse 2s infinite;
}
.reconnect-btn {
  margin-left: auto;
  padding: 4px 12px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.8rem;
}
.reconnect-btn:hover { background: #dc2626; }
@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.5; } }

/* Stats */
.stats-row {
  display: flex; gap: 12px; margin-bottom: 20px; flex-wrap: wrap;
}
.stat-card {
  flex: 1; min-width: 100px;
  background: rgba(30,41,59,0.8);
  border-radius: 12px; padding: 16px;
  text-align: center;
  border: 1px solid rgba(148,163,184,0.1);
}
.stat-count { font-size: 2rem; font-weight: 700; }
.stat-label { font-size: 0.7rem; color: #94a3b8; display: block; margin-top: 4px; }
.stat-card.active .stat-count { color: #34d399; }
.stat-card.warning .stat-count { color: #fbbf24; }
.stat-card.critical .stat-count { color: #f87171; }
.stat-card.expired .stat-count { color: #6b7280; }
.stat-card.error .stat-count { color: #94a3b8; }

/* Filters */
.filter-row {
  display: flex; gap: 12px; margin-bottom: 20px; flex-wrap: wrap;
}
.search-input, .filter-select {
  background: #0f172a; border: 1px solid #334155;
  border-radius: 8px; padding: 8px 12px; color: #e2e8f0;
  font-size: 0.9rem;
}
.search-input { flex: 1; min-width: 200px; }
.search-input:focus, .filter-select:focus {
  outline: none; border-color: #3b82f6;
}

/* Domains Grid */
.domains-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 20px;
}
.domain-card {
  background: rgba(30,41,59,0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148,163,184,0.1);
  padding: 20px;
  transition: all 0.3s ease;
  display: flex; flex-direction: column; gap: 16px;
  position: relative;
}
.domain-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 24px rgba(0,0,0,0.4);
}
.domain-card.active { border-left: 4px solid #10b981; }
.domain-card.warning { border-left: 4px solid #f59e0b; }
.domain-card.critical { border-left: 4px solid #ef4444; }
.domain-card.expired { border-left: 4px solid #6b7280; }
.domain-card.error { border-left: 4px solid #94a3b8; }

/* Card Header */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.status-indicator {
  width: 10px; height: 10px; border-radius: 50%;
}
.status-indicator.active { background: #10b981; box-shadow: 0 0 6px rgba(16,185,129,0.5); }
.status-indicator.warning { background: #f59e0b; }
.status-indicator.critical { background: #ef4444; box-shadow: 0 0 6px rgba(239,68,68,0.5); }
.status-indicator.expired { background: #6b7280; }
.status-indicator.error { background: #94a3b8; }
.title-section h3 { margin: 0; font-size: 1rem; font-weight: 600; color: #f8fafc; }
.domain-host { font-size: 0.8rem; color: #94a3b8; font-family: monospace; }

/* Expiry Progress */
.expiry-progress { display: flex; flex-direction: column; gap: 8px; }
.progress-bar { height: 8px; background: #1e293b; border-radius: 4px; overflow: hidden; }
.progress-fill { height: 100%; border-radius: 4px; transition: width 0.3s; }
.progress-fill.active { background: linear-gradient(90deg, #10b981, #34d399); }
.progress-fill.warning { background: linear-gradient(90deg, #f59e0b, #fbbf24); }
.progress-fill.critical { background: linear-gradient(90deg, #ef4444, #f87171); }
.expiry-days { font-size: 0.9rem; font-weight: 600; }
.expiry-days.active { color: #34d399; }
.expiry-days.warning { color: #fbbf24; }
.expiry-days.critical { color: #f87171; }
.expiry-days.expired { color: #6b7280; }

/* Expiry Date */
.expiry-date {
  display: flex;
  justify-content: space-between;
  padding: 8px 12px;
  background: #0f172a;
  border-radius: 8px;
  font-size: 0.85rem;
}
.expiry-date .label { color: #94a3b8; }
.expiry-date .value { color: #e2e8f0; font-family: monospace; }

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 12px;
  background: #0f172a;
  border-radius: 8px;
}
.stat-item { display: flex; flex-direction: column; gap: 4px; }
.stat-label { font-size: 0.7rem; color: #94a3b8; text-transform: uppercase; }
.stat-value { font-size: 0.9rem; font-weight: 500; color: #f8fafc; }

/* Buttons */
.icon-btn {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 1rem;
  transition: all 0.2s;
}
.icon-btn:hover { background: #1e293b; color: #60a5fa; }
.icon-btn.delete:hover { background: rgba(239,68,68,0.1); color: #f87171; }

.btn {
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  border: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn-secondary { background: #1e293b; color: #cbd5e1; border: 1px solid #334155; }
.btn-secondary:hover:not(:disabled) { background: #2d3748; }
.btn-danger { background: #ef4444; color: white; }
.btn-icon { padding: 8px; background: #1e293b; border: 1px solid #334155; border-radius: 8px; }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* Modal */
.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.8); backdrop-filter: blur(8px);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000;
}
.modal-content {
  background: #1e293b; border-radius: 16px;
  width: 560px; max-width: 95vw; max-height: 90vh; overflow-y: auto;
  border: 1px solid rgba(148,163,184,0.1);
}
.modal-content.small { width: 400px; }
.modal-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 20px 24px; border-bottom: 1px solid #334155;
}
.modal-header h3 { margin: 0; color: #f8fafc; }
.close-btn {
  background: transparent; border: none;
  color: #94a3b8; font-size: 24px; cursor: pointer;
  width: 32px; height: 32px; border-radius: 6px;
}
.close-btn:hover { background: #ef4444; color: white; }
.modal-body { padding: 24px; }
.modal-footer {
  display: flex; justify-content: flex-end; gap: 10px;
  padding: 16px 24px; border-top: 1px solid #334155;
}

/* Form */
.form-group { margin-bottom: 16px; }
.form-group label { display: block; margin-bottom: 6px; color: #cbd5e1; font-size: 0.9rem; }
.form-input {
  width: 100%; padding: 10px 12px;
  background: #0f172a; border: 1px solid #334155;
  border-radius: 8px; color: #e2e8f0; font-size: 0.95rem;
}
.form-input:focus { outline: none; border-color: #3b82f6; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 16px; }
.required { color: #ef4444; margin-left: 4px; }
.hint { font-size: 0.7rem; color: #64748b; display: block; margin-top: 4px; }

/* Test Result */
.test-result {
  margin: 16px 24px; padding: 16px; border-radius: 8px;
}
.test-result.success {
  background: rgba(16,185,129,0.1); border: 1px solid rgba(16,185,129,0.3);
}
.test-result.error {
  background: rgba(239,68,68,0.1); border: 1px solid rgba(239,68,68,0.3);
}
.tr-row {
  display: flex; justify-content: space-between;
  padding: 6px 0; border-bottom: 1px solid rgba(148,163,184,0.1);
}
.tr-row span:first-child { color: #94a3b8; }
.error-msg { color: #f87171; }

/* Detail Grid */
.detail-grid {
  display: grid; grid-template-columns: 1fr 1fr; gap: 16px;
}
.detail-item { display: flex; flex-direction: column; gap: 4px; }
.detail-item.full { grid-column: 1 / -1; }
.detail-item label { font-size: 0.7rem; color: #94a3b8; text-transform: uppercase; }
.detail-item span { color: #f8fafc; }
.error-text { color: #f87171; }
.warning-text { color: #fbbf24; font-size: 0.85rem; margin-top: 8px; }

/* Empty & Loading */
.empty-state, .loading-state {
  text-align: center; padding: 60px; color: #94a3b8;
}
.empty-icon { font-size: 48px; margin-bottom: 16px; opacity: 0.5; }
.loading-spinner {
  width: 40px; height: 40px;
  border: 3px solid #1e293b; border-top-color: #3b82f6;
  border-radius: 50%; animation: spin 1s linear infinite;
  margin: 0 auto 16px;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* Responsive */
@media (max-width: 768px) {
  .domain-monitor { padding: 16px; }
  .domains-grid { grid-template-columns: 1fr; }
  .form-row { grid-template-columns: 1fr; }
  .stats-row { flex-wrap: wrap; }
  .stat-card { min-width: calc(50% - 6px); }
  .detail-grid { grid-template-columns: 1fr; }
}
</style>