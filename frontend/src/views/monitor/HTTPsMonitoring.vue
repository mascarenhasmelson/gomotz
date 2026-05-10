<template>
  <div class="ssl-monitor">
    <!-- Header -->
    <div class="header-row">
      <h2>SSL Certificates</h2>
      <div class="header-actions">
        <button @click="exportCSV" class="btn btn-secondary">
          ⬇️ Download CSV
        </button>
        <button @click="showAddModal = true" class="btn btn-primary">
          ➕ Add Domain
        </button>
        <button @click="fetchMonitors" class="btn btn-icon" title="Refresh">↻</button>
      </div>
    </div>

    <!-- Connection Status -->
    <!-- <div class="connection-bar" :class="wsConnected ? 'connected' : 'disconnected'">
      <span class="dot"></span>
      {{ wsConnected ? 'Live updates active' : 'Disconnected — updates paused' }}
    </div> -->

    <!-- Stats Summary -->
    <div class="stats-row">
      <div class="stat-card valid">
        <span class="stat-count">{{ countByStatus('valid') }}</span>
        <span class="stat-label">Valid</span>
      </div>
      <div class="stat-card warning">
        <span class="stat-count">{{ countByStatus('warning') }}</span>
        <span class="stat-label">Expiring Soon</span>
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

    <!-- Filters -->
    <div class="filter-row">
      <input
        v-model="search"
        type="text"
        placeholder="Search domain or issuer..."
        class="search-input"
      />
      <select v-model="filterStatus" class="filter-select">
        <option value="">All Status</option>
        <option value="valid">Valid</option>
        <option value="warning">Warning</option>
        <option value="critical">Critical</option>
        <option value="expired">Expired</option>
        <option value="error">Error</option>
        <option value="pending">Pending</option>
      </select>
    </div>

    <!-- Table -->
    <div class="table-container">
      <table class="ssl-table">
        <thead>
          <tr>
            <th @click="toggleSort('status')" class="sortable">
              Status <span>{{ getSortIcon('status') }}</span>
            </th>
            <th @click="toggleSort('domain')" class="sortable">
              Domain <span>{{ getSortIcon('domain') }}</span>
            </th>
            <th @click="toggleSort('issuer')" class="sortable">
              Issuer <span>{{ getSortIcon('issuer') }}</span>
            </th>
            <th @click="toggleSort('valid_until')" class="sortable">
              Expiry <span>{{ getSortIcon('valid_until') }}</span>
            </th>
            <th @click="toggleSort('days_remaining')" class="sortable">
              Remaining Days <span>{{ getSortIcon('days_remaining') }}</span>
            </th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="monitor in filteredMonitors"
            :key="monitor.id"
            class="monitor-row"
            :class="monitor.status"
          >
            <!-- Status -->
            <td>
              <div class="status-cell">
                <span class="status-dot" :class="monitor.status"></span>
                <span class="status-text" :class="monitor.status">
                  {{ statusLabel(monitor.status) }}
                </span>
              </div>
            </td>

            <!-- Domain -->
            <td>
              <div class="domain-cell">
                <span class="domain-name">{{ monitor.friendly_name || monitor.domain }}</span>
                <span class="domain-host">{{ monitor.domain }}</span>
              </div>
            </td>

            <!-- Issuer -->
            <td class="issuer-cell">
              {{ monitor.issuer || '—' }}
            </td>

            <!-- Expiry -->
            <td class="expiry-cell">
              <span v-if="monitor.valid_until">
                {{ formatExpiry(monitor.valid_until) }}
              </span>
              <span v-else class="na">—</span>
            </td>

            <!-- Days Remaining -->
            <td>
              <div v-if="monitor.days_remaining !== null && monitor.days_remaining !== undefined"
                   class="days-cell" :class="getDaysClass(monitor)">
                <span class="days-count">{{ monitor.days_remaining }}</span>
                <span class="days-label">days</span>
                <div class="days-bar">
                  <div
                    class="days-fill"
                    :class="monitor.status"
                    :style="{ width: getDaysPercent(monitor) + '%' }"
                  ></div>
                </div>
              </div>
              <span v-else class="na">—</span>
            </td>

            <!-- Actions -->
            <td class="actions-cell">
              <button @click="viewDetails(monitor)" class="action-btn" title="View Details">
                👁️
              </button>
              <button @click="refreshMonitor(monitor)" class="action-btn" title="Refresh Now">
                🔄
              </button>
              <!-- <button @click="editMonitor(monitor)" class="action-btn" title="Edit">
                ✏️
              </button> -->
              <button @click="deleteMonitor(monitor)" class="action-btn danger" title="Delete">
                🗑️
              </button>
            </td>
          </tr>

          <!-- Empty -->
          <tr v-if="filteredMonitors.length === 0 && !loading">
            <td colspan="6" class="empty-state">
              <div>🔒</div>
              <p>No SSL monitors configured</p>
            </td>
          </tr>

          <!-- Loading -->
          <tr v-if="loading">
            <td colspan="6" class="loading-state">
              <div class="spinner"></div>
              <p>Loading...</p>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add/Edit Modal -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ editingId ? 'Edit SSL Monitor' : 'Add SSL Monitor' }}</h3>
          <button @click="closeModal" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Domain <span class="required">*</span></label>
            <input v-model="form.domain" type="text"
              placeholder="e.g., google.com"
              class="form-input" :disabled="!!editingId"/>
            <span class="hint">Port 443 is used by default</span>
          </div>
          <div class="form-group">
            <label>Friendly Name</label>
            <input v-model="form.friendly_name" type="text"
              placeholder="e.g., Google" class="form-input"/>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Port</label>
              <input v-model.number="form.port" type="number"
                min="1" max="65535" class="form-input"/>
            </div>
            <div class="form-group">
              <label>Check Interval</label>
              <select v-model.number="form.check_interval" class="form-input">
                <option :value="3600">1 hour</option>
                <option :value="21600">6 hours</option>
                <option :value="43200">12 hours</option>
                <option :value="86400">24 hours</option>
              </select>
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Warning threshold (days)</label>
              <input v-model.number="form.warning_days" type="number"
                min="1" class="form-input"/>
            </div>
            <div class="form-group">
              <label>Critical threshold (days)</label>
              <input v-model.number="form.critical_days" type="number"
                min="1" class="form-input"/>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="testCertificate" class="btn btn-secondary"
            :disabled="!form.domain || testing">
            {{ testing ? 'Testing...' : '🔍 Test Certificate' }}
          </button>
          <button @click="saveMonitor" class="btn btn-primary"
            :disabled="!form.domain || saving">
            {{ saving ? 'Saving...' : (editingId ? 'Update' : 'Add Monitor') }}
          </button>
        </div>

        <!-- Test Result -->
        <div v-if="testResultData" class="test-result" :class="testResultData.success ? 'success' : 'error'">
          <div v-if="testResultData.success">
            <div class="tr-row">
              <span>Status:</span>
              <span :class="testResultData.status">{{ statusLabel(testResultData.status) }}</span>
            </div>
            <div class="tr-row">
              <span>Issuer:</span>
              <span>{{ testResultData.issuer }}</span>
            </div>
            <div class="tr-row">
              <span>Expires:</span>
              <span>{{ formatExpiry(testResultData.valid_until) }}</span>
            </div>
            <div class="tr-row">
              <span>Days Remaining:</span>
              <span :class="getDaysClassByDays(testResultData.days_remaining)">
                {{ testResultData.days_remaining }} days
              </span>
            </div>
          </div>
          <div v-else class="error-msg">
            ❌ {{ testResultData.error }}
          </div>
        </div>
      </div>
    </div>

    <!-- Detail Modal -->
    <div v-if="detailMonitor" class="modal-overlay" @click.self="detailMonitor = null">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Certificate Details — {{ detailMonitor.domain }}</h3>
          <button @click="detailMonitor = null" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div class="detail-grid">
            <div class="detail-item">
              <label>Status</label>
              <span class="status-text" :class="detailMonitor.status">
                {{ statusLabel(detailMonitor.status) }}
              </span>
            </div>
            <div class="detail-item">
              <label>Domain</label>
              <span>{{ detailMonitor.domain }}</span>
            </div>
            <div class="detail-item">
              <label>Subject (CN)</label>
              <span>{{ detailMonitor.subject || '—' }}</span>
            </div>
            <div class="detail-item">
              <label>Issuer</label>
              <span>{{ detailMonitor.issuer || '—' }}</span>
            </div>
            <div class="detail-item">
              <label>Valid From</label>
              <span>{{ formatExpiry(detailMonitor.valid_from) }}</span>
            </div>
            <div class="detail-item">
              <label>Valid Until</label>
              <span>{{ formatExpiry(detailMonitor.valid_until) }}</span>
            </div>
            <div class="detail-item">
              <label>Days Remaining</label>
              <span :class="getDaysClassByDays(detailMonitor.days_remaining)">
                {{ detailMonitor.days_remaining ?? '—' }} days
              </span>
            </div>
            <div class="detail-item">
              <label>Last Checked</label>
              <span>{{ formatLastChecked(detailMonitor.last_checked_at) }}</span>
            </div>
            <div class="detail-item">
              <label>Warning Threshold</label>
              <span>{{ detailMonitor.warning_days }} days</span>
            </div>
            <div class="detail-item">
              <label>Critical Threshold</label>
              <span>{{ detailMonitor.critical_days }} days</span>
            </div>
            <div class="detail-item full" v-if="detailMonitor.error_message">
              <label>Error</label>
              <span class="error-text">{{ detailMonitor.error_message }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirm Modal -->
    <div v-if="deleteTarget" class="modal-overlay" @click.self="deleteTarget = null">
      <div class="modal-content small">
        <div class="modal-header">
          <h3>Delete Monitor</h3>
          <button @click="deleteTarget = null" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <p>Delete SSL monitor for <strong>{{ deleteTarget.domain }}</strong>?</p>
          <p class="warning-text">This action cannot be undone.</p>
        </div>
        <div class="modal-footer">
          <button @click="confirmDelete" class="btn btn-danger">Delete</button>
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
  name: 'SSLMonitor',
  setup() {
    const monitors = ref([])
    const loading = ref(true)
    const saving = ref(false)
    const testing = ref(false)
    const wsConnected = ref(false)
    const showAddModal = ref(false)
    const editingId = ref(null)
    const detailMonitor = ref(null)
    const deleteTarget = ref(null)
    const testResultData = ref(null)
    const search = ref('')
    const filterStatus = ref('')
    const sortBy = ref('days_remaining')
    const sortDir = ref('asc')
    let ws = null

    const form = reactive({
      domain: '',
      friendly_name: '',
      port: 443,
      check_interval: 3600,
      warning_days: 30,
      critical_days: 7
    })
    const filteredMonitors = computed(() => {
      let list = [...monitors.value]
      if (search.value) {
        const q = search.value.toLowerCase()
        list = list.filter(m =>
          (m.domain || '').toLowerCase().includes(q) ||
          (m.friendly_name || '').toLowerCase().includes(q) ||
          (m.issuer || '').toLowerCase().includes(q)
        )
      }
      if (filterStatus.value) {
        list = list.filter(m => m.status === filterStatus.value)
      }
      list.sort((a, b) => {
        let va = a[sortBy.value]
        let vb = b[sortBy.value]
        if (va === null || va === undefined) va = 999999
        if (vb === null || vb === undefined) vb = 999999
        if (typeof va === 'string') va = va.toLowerCase()
        if (typeof vb === 'string') vb = vb.toLowerCase()
        return sortDir.value === 'asc'
          ? (va < vb ? -1 : va > vb ? 1 : 0)
          : (va > vb ? -1 : va < vb ? 1 : 0)
      })

      return list
    })

    const countByStatus = (status) =>
      monitors.value.filter(m => m.status === status).length
    const connectWebSocket = () => {
      ws = new WebSocket(`${WS_BASE_URL}/v1/api/ws/ssl`)

      ws.onopen = () => {
        wsConnected.value = true
        console.log('✅ SSL WebSocket connected')
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          handleWSMessage(data)
        } catch (e) {
          console.error('WS parse error:', e)
        }
      }

      ws.onclose = () => {
        wsConnected.value = false
        console.log('🔌 SSL WebSocket disconnected')
        setTimeout(connectWebSocket, 3000)
      }

      ws.onerror = (e) => console.error('SSL WS error:', e)
    }

    const handleWSMessage = (data) => {
      switch (data.type) {
        case 'initial_state':
          monitors.value = data.monitors || []
          loading.value = false
          console.log(`📋 Loaded ${monitors.value.length} SSL monitors via WebSocket`)
          break

        case 'ssl_monitor_update':
          const idx = monitors.value.findIndex(m => m.id === data.monitor_id)
          if (idx !== -1) {
            monitors.value[idx] = {
              ...monitors.value[idx],
              status: data.status,
              days_remaining: data.days_remaining,
              issuer: data.issuer,
              valid_until: data.valid_until,
              last_checked_at: data.checked_at,
              error_message: data.error || null
            }
            console.log(`📊 SSL monitor ${data.monitor_id} updated: ${data.status}`)
          }
          break

        case 'ssl_monitor_created':
          monitors.value.push(data.monitor)
          console.log(`➕ SSL monitor created: ${data.monitor.domain}`)
          break

        case 'ssl_monitor_updated':
          const updateIdx = monitors.value.findIndex(m => m.id === data.monitor.id)
          if (updateIdx !== -1) monitors.value[updateIdx] = data.monitor
          console.log(`✏️ SSL monitor updated: ${data.monitor.domain}`)
          break

        case 'ssl_monitor_deleted':
          monitors.value = monitors.value.filter(m => m.id !== data.monitor_id)
          console.log(`🗑️ SSL monitor deleted: ${data.monitor_id}`)
          break
      }
    }
    const fetchMonitors = async () => {
      loading.value = true
      try {
        const res = await fetch(`${API_BASE_URL}/v1/api/ssl`)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        monitors.value = await res.json() || []
        console.log(`📡 Fetched ${monitors.value.length} SSL monitors via REST`)
      } catch (e) {
        console.error('Fetch error:', e)
      } finally {
        loading.value = false
      }
    }

    const saveMonitor = async () => {
      if (!form.domain) return
      saving.value = true
      try {
        const url = editingId.value
          ? `${API_BASE_URL}/v1/api/ssl/${editingId.value}`
          : `${API_BASE_URL}/v1/api/ssl`
        const method = editingId.value ? 'PUT' : 'POST'

        const payload = {
          domain: form.domain,
          friendly_name: form.friendly_name || form.domain,
          port: form.port,
          check_interval: form.check_interval,
          warning_days: form.warning_days,
          critical_days: form.critical_days
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
        closeModal()
      } catch (e) {
        alert(`Failed: ${e.message}`)
      } finally {
        saving.value = false
      }
    }

    const testCertificate = async () => {
      if (!form.domain) return
      testing.value = true
      testResultData.value = null
      try {
        const res = await fetch(`${API_BASE_URL}/v1/api/ssl/test`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            domain: form.domain,
            port: form.port || 443
          })
        })
        testResultData.value = await res.json()
        console.log('Test result:', testResultData.value)
      } catch (e) {
        testResultData.value = { success: false, error: e.message }
      } finally {
        testing.value = false
      }
    }

    const refreshMonitor = async (monitor) => {
      try {
        const res = await fetch(`${API_BASE_URL}/v1/api/ssl/test`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ domain: monitor.domain, port: monitor.port })
        })
        const data = await res.json()
        if (data.success) {
          const idx = monitors.value.findIndex(m => m.id === monitor.id)
          if (idx !== -1) {
            monitors.value[idx] = {
              ...monitors.value[idx],
              status: data.status,
              issuer: data.issuer,
              valid_until: data.valid_until,
              days_remaining: data.days_remaining,
              last_checked_at: new Date().toISOString()
            }
          }
          console.log(`🔄 Refreshed ${monitor.domain}: ${data.status}`)
        }
      } catch (e) {
        console.error('Refresh error:', e)
      }
    }

    const editMonitor = (monitor) => {
      editingId.value = monitor.id
      form.domain = monitor.domain
      form.friendly_name = monitor.friendly_name || ''
      form.port = monitor.port
      form.check_interval = monitor.check_interval
      form.warning_days = monitor.warning_days
      form.critical_days = monitor.critical_days
      testResultData.value = null
      showAddModal.value = true
    }

    const deleteMonitor = (monitor) => {
      deleteTarget.value = monitor
    }

    const confirmDelete = async () => {
      if (!deleteTarget.value) return
      try {
        const res = await fetch(`${API_BASE_URL}/v1/api/ssl/${deleteTarget.value.id}`, {
          method: 'DELETE'
        })
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        deleteTarget.value = null
      } catch (e) {
        alert(`Failed: ${e.message}`)
      }
    }

    const viewDetails = (monitor) => {
      detailMonitor.value = monitor
    }

    const closeModal = () => {
      showAddModal.value = false
      editingId.value = null
      testResultData.value = null
      Object.assign(form, {
        domain: '', friendly_name: '', port: 443,
        check_interval: 3600, warning_days: 30, critical_days: 7
      })
    }

    const exportCSV = () => {
      const headers = ['ID', 'Domain', 'Friendly Name', 'Issuer', 'Expiry', 'Days Remaining', 'Status']
      const rows = monitors.value.map(m => [
        m.id,
        m.domain,
        m.friendly_name || '',
        m.issuer || '',
        m.valid_until ? new Date(m.valid_until).toLocaleString() : '',
        m.days_remaining ?? '',
        m.status
      ])

      const csv = [headers, ...rows]
        .map(r => r.map(v => `"${String(v).replace(/"/g, '""')}"`).join(','))
        .join('\n')

      const blob = new Blob([csv], { type: 'text/csv' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `ssl-certificates-${new Date().toISOString().slice(0, 10)}.csv`
      a.click()
      URL.revokeObjectURL(url)
    }
    const statusLabel = (status) => {
      const map = {
        valid: '✅ Valid',
        warning: '⚠️ Warning',
        critical: '🔴 Critical',
        expired: '❌ Expired',
        error: '⚠️ Error',
        pending: '⏳ Pending'
      }
      return map[status] || status
    }

    const formatExpiry = (ts) => {
      if (!ts) return '—'
      return new Date(ts).toLocaleString('en-US', {
        year: 'numeric', month: 'short', day: 'numeric',
        hour: '2-digit', minute: '2-digit'
      })
    }

    const formatLastChecked = (ts) => {
      if (!ts) return 'Never'
      const diff = Math.floor((Date.now() - new Date(ts)) / 1000)
      if (diff < 60) return 'Just now'
      if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
      if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
      return `${Math.floor(diff / 86400)}d ago`
    }

    const getDaysPercent = (monitor) => {
      const days = monitor.days_remaining
      if (days === null || days === undefined || days <= 0) return 0
      return Math.min(100, Math.round((days / 365) * 100))
    }

    const getDaysClass = (monitor) => {
      return getDaysClassByDays(monitor.days_remaining)
    }

    const getDaysClassByDays = (days) => {
      if (days === null || days === undefined || days <= 0) return 'expired'
      if (days <= 7) return 'critical'
      if (days <= 30) return 'warning'
      return 'valid'
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
      fetchMonitors()
      connectWebSocket()
    })

    onUnmounted(() => {
      if (ws) ws.close()
    })

    return {
      monitors, loading, saving, testing, wsConnected,
      showAddModal, editingId, detailMonitor, deleteTarget,
      testResultData, search, filterStatus, sortBy, sortDir,
      form, filteredMonitors,
      countByStatus, fetchMonitors, saveMonitor, testCertificate,
      refreshMonitor, editMonitor, deleteMonitor, confirmDelete,
      viewDetails, closeModal, exportCSV,
      statusLabel, formatExpiry, formatLastChecked,
      getDaysPercent, getDaysClass, getDaysClassByDays,
      toggleSort, getSortIcon
    }
  }
}
</script>

<style scoped>
.ssl-monitor {
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

/* Connection bar */
.connection-bar {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 16px; border-radius: 8px;
  font-size: 0.85rem; margin-bottom: 20px;
}
.connection-bar.connected {
  background: rgba(16,185,129,0.1);
  border: 1px solid rgba(16,185,129,0.3); color: #34d399;
}
.connection-bar.disconnected {
  background: rgba(239,68,68,0.1);
  border: 1px solid rgba(239,68,68,0.3); color: #f87171;
}
.dot {
  width: 8px; height: 8px; border-radius: 50%;
  background: currentColor;
}

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
  display: flex; flex-direction: column; gap: 4px;
}
.stat-count { font-size: 2rem; font-weight: 700; }
.stat-label { font-size: 0.75rem; color: #94a3b8; text-transform: uppercase; }
.stat-card.valid .stat-count { color: #34d399; }
.stat-card.warning .stat-count { color: #fbbf24; }
.stat-card.critical .stat-count { color: #f87171; }
.stat-card.expired .stat-count { color: #6b7280; }
.stat-card.error .stat-count { color: #94a3b8; }

/* Filters */
.filter-row {
  display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap;
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

/* Table */
.table-container {
  background: rgba(30,41,59,0.8);
  border-radius: 16px; border: 1px solid rgba(148,163,184,0.1);
  overflow-x: auto;
}
.ssl-table { width: 100%; border-collapse: collapse; min-width: 800px; }
.ssl-table thead th {
  padding: 14px 16px;
  background: rgba(15,23,42,0.6);
  color: #94a3b8; font-size: 0.8rem;
  text-transform: uppercase; letter-spacing: 0.5px;
  text-align: left; border-bottom: 1px solid #334155;
}
.ssl-table thead th.sortable { cursor: pointer; user-select: none; }
.ssl-table thead th.sortable:hover { color: #60a5fa; }

.ssl-table tbody tr { border-bottom: 1px solid rgba(148,163,184,0.05); }
.ssl-table tbody tr:hover { background: rgba(59,130,246,0.05); }
.ssl-table tbody td { padding: 14px 16px; }

/* Status cell */
.status-cell { display: flex; align-items: center; gap: 8px; }
.status-dot {
  width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0;
}
.status-dot.valid { background: #10b981; box-shadow: 0 0 6px rgba(16,185,129,0.5); }
.status-dot.warning { background: #f59e0b; }
.status-dot.critical { background: #ef4444; box-shadow: 0 0 6px rgba(239,68,68,0.5); }
.status-dot.expired { background: #6b7280; }
.status-dot.error { background: #94a3b8; }
.status-dot.pending { background: #475569; }

.status-text.valid { color: #34d399; }
.status-text.warning { color: #fbbf24; }
.status-text.critical { color: #f87171; }
.status-text.expired { color: #9ca3af; }
.status-text.error { color: #94a3b8; }
.status-text.pending { color: #64748b; }

/* Domain cell */
.domain-cell { display: flex; flex-direction: column; }
.domain-name { font-weight: 600; color: #f8fafc; }
.domain-host { font-size: 0.8rem; color: #94a3b8; font-family: monospace; }

/* Issuer */
.issuer-cell { color: #94a3b8; font-size: 0.9rem; max-width: 250px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* Expiry */
.expiry-cell { font-family: monospace; font-size: 0.85rem; color: #cbd5e1; }

/* Days */
.days-cell { display: flex; flex-direction: column; gap: 4px; }
.days-count { font-size: 1.2rem; font-weight: 700; }
.days-label { font-size: 0.7rem; color: #94a3b8; }
.days-cell.valid .days-count { color: #34d399; }
.days-cell.warning .days-count { color: #fbbf24; }
.days-cell.critical .days-count { color: #f87171; }
.days-cell.expired .days-count { color: #9ca3af; }

.days-bar {
  height: 4px; background: #1e293b;
  border-radius: 2px; overflow: hidden; width: 80px;
}
.days-fill {
  height: 100%; border-radius: 2px; transition: width 0.3s;
}
.days-fill.valid { background: #10b981; }
.days-fill.warning { background: #f59e0b; }
.days-fill.critical { background: #ef4444; }
.days-fill.expired { background: #6b7280; }

.na { color: #475569; }

/* Actions */
.actions-cell { display: flex; gap: 6px; }
.action-btn {
  background: transparent; border: none;
  color: #64748b; cursor: pointer; padding: 5px 8px;
  border-radius: 4px; font-size: 1rem; transition: all 0.2s;
}
.action-btn:hover { background: #1e293b; color: #60a5fa; }
.action-btn.danger:hover { background: rgba(239,68,68,0.1); color: #f87171; }

/* Empty / Loading */
.empty-state, .loading-state {
  text-align: center; padding: 60px; color: #94a3b8;
}
.empty-state div { font-size: 48px; margin-bottom: 12px; opacity: 0.5; }
.spinner {
  width: 36px; height: 36px;
  border: 3px solid #1e293b; border-top-color: #3b82f6;
  border-radius: 50%; animation: spin 1s linear infinite;
  margin: 0 auto 12px;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* Buttons */
.btn {
  padding: 8px 16px; border-radius: 8px; font-size: 0.9rem;
  font-weight: 500; cursor: pointer; transition: all 0.2s;
  border: none; display: flex; align-items: center; gap: 6px;
}
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover:not(:disabled) { background: #2563eb; }
.btn-secondary { background: #1e293b; color: #cbd5e1; border: 1px solid #334155; }
.btn-secondary:hover:not(:disabled) { background: #2d3748; }
.btn-danger { background: #ef4444; color: white; }
.btn-danger:hover { background: #dc2626; }
.btn-icon {
  padding: 8px; background: #1e293b;
  border: 1px solid #334155; border-radius: 8px;
  color: #94a3b8; cursor: pointer; font-size: 1.1rem;
}
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
  box-shadow: 0 20px 40px rgba(0,0,0,0.5);
}
.modal-content.small { width: 400px; }
.modal-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 20px 24px; border-bottom: 1px solid #334155;
}
.modal-header h3 { margin: 0; color: #f8fafc; font-size: 1.1rem; }
.close-btn {
  background: transparent; border: none; color: #94a3b8;
  font-size: 24px; cursor: pointer; padding: 0;
  width: 32px; height: 32px; display: flex;
  align-items: center; justify-content: center; border-radius: 6px;
}
.close-btn:hover { background: #ef4444; color: white; }
.modal-body { padding: 24px; }
.modal-footer {
  display: flex; justify-content: flex-end; gap: 10px;
  padding: 16px 24px; border-top: 1px solid #334155;
}

/* Form */
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 16px; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.form-group label { color: #cbd5e1; font-size: 0.9rem; font-weight: 500; }
.form-input {
  background: #0f172a; border: 1px solid #334155;
  border-radius: 8px; padding: 10px 12px; color: #e2e8f0; font-size: 0.95rem;
}
.form-input:focus { outline: none; border-color: #3b82f6; }
.form-input:disabled { opacity: 0.5; }
.hint { font-size: 0.75rem; color: #64748b; }
.required { color: #ef4444; }

/* Test result */
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
  font-size: 0.9rem;
}
.tr-row:last-child { border: none; }
.tr-row span:first-child { color: #94a3b8; }
.error-msg { color: #f87171; }

/* Detail grid */
.detail-grid {
  display: grid; grid-template-columns: 1fr 1fr; gap: 16px;
}
.detail-item {
  display: flex; flex-direction: column; gap: 4px;
}
.detail-item.full { grid-column: 1 / -1; }
.detail-item label {
  font-size: 0.75rem; color: #94a3b8;
  text-transform: uppercase; letter-spacing: 0.5px;
}
.detail-item span { color: #f8fafc; font-size: 0.95rem; }
.error-text { color: #f87171; }
.warning-text { color: #fbbf24; font-size: 0.85rem; }

/* Responsive */
@media (max-width: 768px) {
  .ssl-monitor { padding: 16px; }
  .stats-row { flex-wrap: wrap; }
  .stat-card { min-width: calc(50% - 6px); }
  .form-row { grid-template-columns: 1fr; }
  .detail-grid { grid-template-columns: 1fr; }
  .header-actions { flex-wrap: wrap; }
  .issuer-cell { max-width: 150px; }
}
</style>