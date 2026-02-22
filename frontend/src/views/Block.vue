<template>
  <div class="block-devices-dashboard">
    <!-- Header -->
    <div class="dashboard-header">
      <h1>Internet Access Control</h1>
      <p class="subtitle">Block or unblock internet access for specific devices on your network</p>
    </div>

    <!-- Stats Cards -->
    <div class="stats-cards">
      <div class="stat-card">
        <div class="stat-icon">📊</div>
        <div class="stat-content">
          <div class="stat-value">{{ totalDevices }}</div>
          <div class="stat-label">Total Devices</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">✅</div>
        <div class="stat-content">
          <div class="stat-value">{{ onlineDevices }}</div>
          <div class="stat-label">Online</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">🔴</div>
        <div class="stat-content">
          <div class="stat-value">{{ blockedDevices }}</div>
          <div class="stat-label">Blocked</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">🟢</div>
        <div class="stat-content">
          <div class="stat-value">{{ unblockedDevices }}</div>
          <div class="stat-label">Unblocked</div>
        </div>
      </div>
    </div>

    <!-- Filters and Search -->
    <div class="filters-bar">
      <div class="search-box">
        <span class="search-icon">🔍</span>
        <input 
          v-model="searchQuery" 
          type="text" 
          placeholder="Search by IP, MAC, hostname, or device type..."
          class="search-input"
        />
      </div>
      
      <div class="filter-options">
        <button 
          class="filter-btn" 
          :class="{ active: statusFilter === 'all' }"
          @click="statusFilter = 'all'"
        >
          All
        </button>
        <button 
          class="filter-btn" 
          :class="{ active: statusFilter === 'online' }"
          @click="statusFilter = 'online'"
        >
          Online
        </button>
        <button 
          class="filter-btn" 
          :class="{ active: statusFilter === 'offline' }"
          @click="statusFilter = 'offline'"
        >
          Offline
        </button>
        <button 
          class="filter-btn" 
          :class="{ active: statusFilter === 'blocked' }"
          @click="statusFilter = 'blocked'"
        >
          Blocked
        </button>
        <button 
          class="filter-btn" 
          :class="{ active: statusFilter === 'unblocked' }"
          @click="statusFilter = 'unblocked'"
        >
          Unblocked
        </button>
      </div>
      
      <div class="bulk-actions">
        <button class="bulk-btn" @click="selectAll" :disabled="!filteredDevices.length">
          Select All
        </button>
        <button class="bulk-btn" @click="clearSelection" :disabled="!selectedDevices.length">
          Clear Selection
        </button>
        <button 
          class="bulk-toggle-btn" 
          @click="bulkToggleDevices" 
          :disabled="!selectedDevices.length"
        >
          <span v-if="hasBlockedInSelection">🔓 Unblock Selected ({{ selectedDevices.length }})</span>
          <span v-else>🔒 Block Selected ({{ selectedDevices.length }})</span>
        </button>
      </div>
    </div>

    <!-- Devices Table -->
    <div class="devices-table-container">
      <table class="devices-table">
        <thead>
          <tr>
            <th width="40">
              <input 
                type="checkbox" 
                v-model="selectAllCheckbox" 
                @change="toggleSelectAll"
                :indeterminate="selectedDevices.length > 0 && selectedDevices.length < filteredDevices.length"
              >
            </th>
            <th>Status</th>
            <th>IP Address</th>
            <th>MAC Address</th>
            <th>Hostname</th>
            <th>Device Type</th>
            <th>Vendor</th>
            <th>Last Seen</th>
            <th>Access Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="device in paginatedDevices" :key="device.mac" class="device-row">
            <td>
              <input 
                type="checkbox" 
                v-model="selectedDevices" 
                :value="device"
                @change="updateSelection"
              >
            </td>
            <td>
              <span class="status-indicator" :class="device.status">
                {{ device.status === 'online' ? '🟢' : '🔴' }}
              </span>
            </td>
            <td class="ip-cell">{{ device.ip }}</td>
            <td class="mac-cell">{{ formatMAC(device.mac) }}</td>
            <td>{{ device.hostname || 'Unknown' }}</td>
            <td>
              <span class="device-type-badge">{{ device.deviceType || 'Generic' }}</span>
            </td>
            <td>{{ device.vendor || 'Unknown' }}</td>
            <td class="last-seen">{{ formatTime(device.lastSeen) }}</td>
            <td>
              <span class="access-badge" :class="device.blocked ? 'blocked' : 'unblocked'">
                {{ device.blocked ? '🔒 Blocked' : '🔓 Unblocked' }}
              </span>
            </td>
            <td class="actions-cell">
              <button 
                class="toggle-btn" 
                :class="device.blocked ? 'unblock' : 'block'"
                @click="toggleDeviceAccess(device)"
                :title="device.blocked ? 'Unblock Internet Access' : 'Block Internet Access'"
              >
                <span v-if="device.blocked">🔓 Unblock</span>
                <span v-else>🔒 Block</span>
              </button>
              <button 
                class="info-btn" 
                @click="showDeviceDetails(device)"
                title="Device Details"
              >
                ℹ️
              </button>
            </td>
          </tr>
          
          <!-- Empty State -->
          <tr v-if="filteredDevices.length === 0">
            <td colspan="10" class="empty-state">
              <div class="empty-icon">🔍</div>
              <h3>No Devices Found</h3>
              <p v-if="searchQuery">No devices match your search criteria</p>
              <p v-else>No devices available to display</p>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div class="pagination">
      <button 
        class="page-btn" 
        @click="currentPage--" 
        :disabled="currentPage === 1"
      >
        ← Previous
      </button>
      <span class="page-info">
        Page {{ currentPage }} of {{ totalPages }}
      </span>
      <button 
        class="page-btn" 
        @click="currentPage++" 
        :disabled="currentPage === totalPages"
      >
        Next →
      </button>
      <select v-model="itemsPerPage" class="items-per-page">
        <option :value="10">10 per page</option>
        <option :value="25">25 per page</option>
        <option :value="50">50 per page</option>
        <option :value="100">100 per page</option>
      </select>
    </div>

    <!-- Quick Block/Unblock Summary -->
    <div class="summary-section">
      <h3>Blocked Devices Summary</h3>
      <div class="summary-grid">
        <div v-for="(count, type) in blockedByType" :key="type" class="summary-item">
          <span class="summary-type">{{ type || 'Unknown' }}</span>
          <span class="summary-count">{{ count }} blocked</span>
        </div>
      </div>
    </div>

    <!-- Device Details Modal -->
    <div v-if="selectedDeviceDetails" class="modal-overlay" @click.self="closeModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Device Details</h3>
          <button class="close-btn" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="detail-grid">
            <div class="detail-row">
              <span class="detail-label">IP Address:</span>
              <span class="detail-value">{{ selectedDeviceDetails.ip }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">MAC Address:</span>
              <span class="detail-value mac">{{ formatMAC(selectedDeviceDetails.mac) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Hostname:</span>
              <span class="detail-value">{{ selectedDeviceDetails.hostname || 'Unknown' }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Device Type:</span>
              <span class="detail-value">{{ selectedDeviceDetails.deviceType || 'Generic' }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Vendor:</span>
              <span class="detail-value">{{ selectedDeviceDetails.vendor || 'Unknown' }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Status:</span>
              <span class="detail-value" :class="selectedDeviceDetails.status">
                {{ selectedDeviceDetails.status === 'online' ? '🟢 Online' : '🔴 Offline' }}
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Access:</span>
              <span class="detail-value" :class="selectedDeviceDetails.blocked ? 'blocked' : 'unblocked'">
                {{ selectedDeviceDetails.blocked ? '🔒 Blocked' : '🔓 Unblocked' }}
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">First Seen:</span>
              <span class="detail-value">{{ formatFullTime(selectedDeviceDetails.firstSeen) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Last Seen:</span>
              <span class="detail-value">{{ formatFullTime(selectedDeviceDetails.lastSeen) }}</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button 
            class="modal-toggle-btn" 
            :class="selectedDeviceDetails.blocked ? 'unblock' : 'block'"
            @click="toggleDeviceAccess(selectedDeviceDetails); closeModal()"
          >
            <span v-if="selectedDeviceDetails.blocked">🔓 Unblock Device</span>
            <span v-else>🔒 Block Device</span>
          </button>
          <button class="modal-btn close" @click="closeModal">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'

export default {
  name: 'BlockDevicesDashboard',
  
  setup() {
    // State
    const devices = ref([])
    const searchQuery = ref('')
    const statusFilter = ref('all')
    const currentPage = ref(1)
    const itemsPerPage = ref(25)
    const selectedDevices = ref([])
    const selectAllCheckbox = ref(false)
    const selectedDeviceDetails = ref(null)
    
    // Load devices from localStorage or API
    const loadDevices = () => {
      // Sample data - replace with your actual data source
      const sampleDevices = [
        { 
          ip: '192.168.1.10', 
          mac: '00:11:22:33:44:55', 
          hostname: 'john-laptop', 
          deviceType: 'Laptop', 
          vendor: 'Dell Inc.', 
          status: 'online', 
          blocked: false,
          firstSeen: new Date(Date.now() - 86400000).toISOString(),
          lastSeen: new Date().toISOString()
        },
        { 
          ip: '192.168.1.20', 
          mac: 'aa:bb:cc:dd:ee:ff', 
          hostname: 'living-room-tv', 
          deviceType: 'Smart TV', 
          vendor: 'Samsung', 
          status: 'online', 
          blocked: true,
          firstSeen: new Date(Date.now() - 172800000).toISOString(),
          lastSeen: new Date().toISOString()
        },
        { 
          ip: '192.168.1.30', 
          mac: '11:22:33:44:55:66', 
          hostname: 'kids-ipad', 
          deviceType: 'Tablet', 
          vendor: 'Apple Inc.', 
          status: 'offline', 
          blocked: false,
          firstSeen: new Date(Date.now() - 259200000).toISOString(),
          lastSeen: new Date(Date.now() - 3600000).toISOString()
        },
        { 
          ip: '192.168.1.40', 
          mac: '22:33:44:55:66:77', 
          hostname: 'guest-phone', 
          deviceType: 'Smartphone', 
          vendor: 'Google', 
          status: 'online', 
          blocked: true,
          firstSeen: new Date(Date.now() - 43200000).toISOString(),
          lastSeen: new Date().toISOString()
        },
        { 
          ip: '192.168.1.50', 
          mac: '33:44:55:66:77:88', 
          hostname: 'server-nas', 
          deviceType: 'NAS', 
          vendor: 'Synology', 
          status: 'online', 
          blocked: false,
          firstSeen: new Date(Date.now() - 604800000).toISOString(),
          lastSeen: new Date().toISOString()
        },
        { 
          ip: '192.168.1.60', 
          mac: '44:55:66:77:88:99', 
          hostname: 'office-printer', 
          deviceType: 'Printer', 
          vendor: 'HP Inc.', 
          status: 'offline', 
          blocked: false,
          firstSeen: new Date(Date.now() - 345600000).toISOString(),
          lastSeen: new Date(Date.now() - 7200000).toISOString()
        },
        { 
          ip: '192.168.1.70', 
          mac: '55:66:77:88:99:aa', 
          hostname: 'security-camera', 
          deviceType: 'IP Camera', 
          vendor: 'Hikvision', 
          status: 'online', 
          blocked: false,
          firstSeen: new Date(Date.now() - 1209600000).toISOString(),
          lastSeen: new Date().toISOString()
        },
        { 
          ip: '192.168.1.80', 
          mac: '66:77:88:99:aa:bb', 
          hostname: 'game-console', 
          deviceType: 'Gaming Console', 
          vendor: 'Sony', 
          status: 'online', 
          blocked: true,
          firstSeen: new Date(Date.now() - 864000000).toISOString(),
          lastSeen: new Date().toISOString()
        },
        { 
          ip: '192.168.1.90', 
          mac: '77:88:99:aa:bb:cc', 
          hostname: 'smart-speaker', 
          deviceType: 'IoT Device', 
          vendor: 'Amazon', 
          status: 'online', 
          blocked: false,
          firstSeen: new Date(Date.now() - 432000000).toISOString(),
          lastSeen: new Date().toISOString()
        },
        { 
          ip: '192.168.1.100', 
          mac: '88:99:aa:bb:cc:dd', 
          hostname: 'workstation', 
          deviceType: 'Desktop', 
          vendor: 'Custom', 
          status: 'offline', 
          blocked: false,
          firstSeen: new Date(Date.now() - 518400000).toISOString(),
          lastSeen: new Date(Date.now() - 14400000).toISOString()
        },
        { 
          ip: '192.168.1.110', 
          mac: '99:aa:bb:cc:dd:ee', 
          hostname: 'thermostat', 
          deviceType: 'IoT Device', 
          vendor: 'Nest', 
          status: 'online', 
          blocked: true,
          firstSeen: new Date(Date.now() - 388800000).toISOString(),
          lastSeen: new Date().toISOString()
        },
        { 
          ip: '192.168.1.120', 
          mac: 'aa:bb:cc:dd:ee:ff', 
          hostname: 'smart-lock', 
          deviceType: 'IoT Device', 
          vendor: 'August', 
          status: 'online', 
          blocked: false,
          firstSeen: new Date(Date.now() - 777600000).toISOString(),
          lastSeen: new Date().toISOString()
        }
      ]
      
      // Load from localStorage if available
      const savedDevices = localStorage.getItem('blockedDevices')
      if (savedDevices) {
        try {
          const parsed = JSON.parse(savedDevices)
          // Merge with sample data to ensure all fields exist
          devices.value = sampleDevices.map(sample => {
            const saved = parsed.find(d => d.mac === sample.mac)
            return { ...sample, ...saved }
          })
        } catch (e) {
          devices.value = sampleDevices
        }
      } else {
        devices.value = sampleDevices
      }
    }
    
    // Computed properties
    const filteredDevices = computed(() => {
      let filtered = [...devices.value]
      
      // Apply status filter
      if (statusFilter.value !== 'all') {
        filtered = filtered.filter(d => {
          if (statusFilter.value === 'online') return d.status === 'online'
          if (statusFilter.value === 'offline') return d.status === 'offline'
          if (statusFilter.value === 'blocked') return d.blocked === true
          if (statusFilter.value === 'unblocked') return d.blocked === false
          return true
        })
      }
      
      // Apply search
      if (searchQuery.value) {
        const query = searchQuery.value.toLowerCase()
        filtered = filtered.filter(d => 
          d.ip.toLowerCase().includes(query) ||
          d.mac.toLowerCase().includes(query) ||
          (d.hostname && d.hostname.toLowerCase().includes(query)) ||
          (d.deviceType && d.deviceType.toLowerCase().includes(query)) ||
          (d.vendor && d.vendor.toLowerCase().includes(query))
        )
      }
      
      return filtered
    })
    
    const totalDevices = computed(() => devices.value.length)
    const onlineDevices = computed(() => devices.value.filter(d => d.status === 'online').length)
    const blockedDevices = computed(() => devices.value.filter(d => d.blocked).length)
    const unblockedDevices = computed(() => devices.value.filter(d => !d.blocked).length)
    
    const totalPages = computed(() => Math.ceil(filteredDevices.value.length / itemsPerPage.value))
    
    const paginatedDevices = computed(() => {
      const start = (currentPage.value - 1) * itemsPerPage.value
      return filteredDevices.value.slice(start, start + itemsPerPage.value)
    })
    
    const blockedByType = computed(() => {
      const counts = {}
      devices.value.filter(d => d.blocked).forEach(d => {
        const type = d.deviceType || 'Unknown'
        counts[type] = (counts[type] || 0) + 1
      })
      return counts
    })
    
    const hasBlockedInSelection = computed(() => {
      return selectedDevices.value.some(d => d.blocked)
    })
    
    // Methods
    const formatMAC = (mac) => {
      if (!mac) return 'Unknown'
      return mac.toUpperCase()
    }
    
    const formatTime = (timestamp) => {
      if (!timestamp) return 'Never'
      const date = new Date(timestamp)
      const now = new Date()
      const diff = now - date
      
      if (diff < 60000) return 'Just now'
      if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
      if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`
      return date.toLocaleDateString()
    }
    
    const formatFullTime = (timestamp) => {
      if (!timestamp) return 'Never'
      return new Date(timestamp).toLocaleString()
    }
    
    const toggleDeviceAccess = (device) => {
      const action = device.blocked ? 'unblock' : 'block'
      const actionText = device.blocked ? 'Unblock' : 'Block'
      
      if (!confirm(`${actionText} internet access for ${device.hostname || device.ip}?`)) return
      
      const index = devices.value.findIndex(d => d.mac === device.mac)
      if (index !== -1) {
        devices.value[index].blocked = !device.blocked
        saveToLocalStorage()
        console.log(`${actionText}ing device:`, device.ip)
        alert(`✅ ${actionText}ed internet for ${device.hostname || device.ip}`)
      }
    }
    
    const bulkToggleDevices = () => {
      if (!selectedDevices.value.length) return
      
      const willBlock = !hasBlockedInSelection.value
      const action = willBlock ? 'Block' : 'Unblock'
      const count = selectedDevices.value.length
      
      if (!confirm(`${action} internet access for ${count} selected device${count > 1 ? 's' : ''}?`)) return
      
      selectedDevices.value.forEach(device => {
        const index = devices.value.findIndex(d => d.mac === device.mac)
        if (index !== -1) {
          devices.value[index].blocked = willBlock
        }
      })
      
      saveToLocalStorage()
      selectedDevices.value = []
      selectAllCheckbox.value = false
      alert(`✅ ${action}ed internet for ${count} device${count > 1 ? 's' : ''}`)
    }
    
    const selectAll = () => {
      selectedDevices.value = [...filteredDevices.value]
      selectAllCheckbox.value = true
    }
    
    const clearSelection = () => {
      selectedDevices.value = []
      selectAllCheckbox.value = false
    }
    
    const toggleSelectAll = () => {
      if (selectAllCheckbox.value) {
        selectAll()
      } else {
        clearSelection()
      }
    }
    
    const updateSelection = () => {
      selectAllCheckbox.value = selectedDevices.value.length === filteredDevices.value.length
    }
    
    const saveToLocalStorage = () => {
      localStorage.setItem('blockedDevices', JSON.stringify(devices.value))
    }
    
    const showDeviceDetails = (device) => {
      selectedDeviceDetails.value = device
    }
    
    const closeModal = () => {
      selectedDeviceDetails.value = null
    }
    
    // Watch for filter changes
    watch([searchQuery, statusFilter], () => {
      currentPage.value = 1
      selectedDevices.value = []
      selectAllCheckbox.value = false
    })
    
    watch(itemsPerPage, () => {
      currentPage.value = 1
    })
    
    // Initialize
    onMounted(() => {
      loadDevices()
    })
    
    return {
      devices,
      searchQuery,
      statusFilter,
      currentPage,
      itemsPerPage,
      selectedDevices,
      selectAllCheckbox,
      selectedDeviceDetails,
      filteredDevices,
      totalDevices,
      onlineDevices,
      blockedDevices,
      unblockedDevices,
      totalPages,
      paginatedDevices,
      blockedByType,
      hasBlockedInSelection,
      formatMAC,
      formatTime,
      formatFullTime,
      toggleDeviceAccess,
      bulkToggleDevices,
      selectAll,
      clearSelection,
      toggleSelectAll,
      updateSelection,
      showDeviceDetails,
      closeModal
    }
  }
}
</script>

<style scoped>
/* Dark Mode Theme */
.block-devices-dashboard {
  min-height: 100vh;
  width: 100%;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  color: #e2e8f0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  padding: 24px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}

.dashboard-header {
  margin-bottom: 20px;
  width: 100%;
}

.dashboard-header h1 {
  font-size: 2rem;
  font-weight: 700;
  margin: 0 0 4px 0;
  background: linear-gradient(135deg, #60a5fa 0%, #a78bfa 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  line-height: 1.2;
}

.subtitle {
  color: #94a3b8;
  font-size: 0.95rem;
  margin: 0;
}

/* Stats Cards */
.stats-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;
  width: 100%;
}

@media (max-width: 1200px) {
  .stats-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .stats-cards {
    grid-template-columns: 1fr;
  }
}

.stat-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 10px;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  transition: all 0.2s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  border-color: #3b82f6;
}

.stat-icon {
  font-size: 2rem;
  flex-shrink: 0;
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #f8fafc;
  line-height: 1.2;
  margin-bottom: 2px;
}

.stat-label {
  color: #94a3b8;
  font-size: 0.8rem;
}

/* Filters Bar */
.filters-bar {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 20px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  width: 100%;
  box-sizing: border-box;
}

.search-box {
  flex: 1;
  min-width: 250px;
  max-width: 350px;
  position: relative;
}

@media (max-width: 768px) {
  .search-box {
    max-width: 100%;
    width: 100%;
  }
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #64748b;
  font-size: 1rem;
}

.search-input {
  width: 100%;
  padding: 10px 12px 10px 36px;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #e2e8f0;
  font-size: 0.9rem;
  transition: all 0.2s;
  box-sizing: border-box;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
}

.filter-options {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  flex: 0 1 auto;
}

.filter-btn {
  padding: 8px 12px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #cbd5e1;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.filter-btn:hover {
  border-color: #3b82f6;
}

.filter-btn.active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: white;
}

.bulk-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
  flex-wrap: wrap;
}

@media (max-width: 768px) {
  .bulk-actions {
    margin-left: 0;
    width: 100%;
    justify-content: flex-start;
  }
}

.bulk-btn {
  padding: 8px 12px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #cbd5e1;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.bulk-btn:hover:not(:disabled) {
  border-color: #3b82f6;
}

.bulk-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.bulk-toggle-btn {
  padding: 8px 16px;
  background: #3b82f6;
  border: 1px solid #3b82f6;
  border-radius: 6px;
  color: white;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.bulk-toggle-btn:hover:not(:disabled) {
  background: #2563eb;
}

.bulk-toggle-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Devices Table */
.devices-table-container {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  overflow: auto;
  margin-bottom: 20px;
  width: 100%;
  flex: 1;
  min-height: 0;
}

.devices-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1000px;
}

.devices-table th {
  text-align: left;
  padding: 14px 12px;
  background: #0f172a;
  color: #94a3b8;
  font-weight: 500;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid #334155;
  position: sticky;
  top: 0;
  z-index: 10;
  white-space: nowrap;
}

.devices-table td {
  padding: 12px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  color: #e2e8f0;
  font-size: 0.9rem;
}

.devices-table tbody tr:hover {
  background: rgba(59, 130, 246, 0.1);
}

.devices-table input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #3b82f6;
  margin: 0;
}

.status-indicator {
  display: inline-block;
  font-size: 1rem;
}

.ip-cell {
  font-family: 'Monaco', 'Courier New', monospace;
  color: #60a5fa;
  font-size: 0.9rem;
  white-space: nowrap;
}

.mac-cell {
  font-family: 'Monaco', 'Courier New', monospace;
  color: #94a3b8;
  font-size: 0.85rem;
  white-space: nowrap;
}

.device-type-badge {
  display: inline-block;
  padding: 4px 8px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 4px;
  font-size: 0.8rem;
  color: #cbd5e1;
  white-space: nowrap;
}

.last-seen {
  color: #94a3b8;
  font-size: 0.85rem;
  white-space: nowrap;
}

.access-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
  white-space: nowrap;
}

.access-badge.blocked {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.access-badge.unblocked {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.actions-cell {
  display: flex;
  gap: 6px;
  white-space: nowrap;
}

.toggle-btn {
  padding: 6px 10px;
  border: none;
  border-radius: 4px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
  color: white;
  white-space: nowrap;
}

.toggle-btn.block {
  background: #ef4444;
}

.toggle-btn.block:hover {
  background: #dc2626;
}

.toggle-btn.unblock {
  background: #10b981;
}

.toggle-btn.unblock:hover {
  background: #059669;
}

.info-btn {
  padding: 6px 8px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 4px;
  color: #94a3b8;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
  min-width: 30px;
}

.info-btn:hover {
  border-color: #3b82f6;
  color: #60a5fa;
}

/* Pagination */
.pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

@media (max-width: 768px) {
  .pagination {
    justify-content: center;
  }
}

.page-btn {
  padding: 8px 14px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #cbd5e1;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.page-btn:hover:not(:disabled) {
  border-color: #3b82f6;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  color: #94a3b8;
  font-size: 0.9rem;
  white-space: nowrap;
}

.items-per-page {
  padding: 8px 10px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #e2e8f0;
  font-size: 0.9rem;
  cursor: pointer;
  outline: none;
}

/* Summary Section */
.summary-section {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 10px;
  padding: 20px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  width: 100%;
  box-sizing: border-box;
  margin-top: auto;
}

.summary-section h3 {
  margin: 0 0 16px 0;
  color: #f8fafc;
  font-size: 1.1rem;
  font-weight: 600;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: #0f172a;
  border-radius: 6px;
  border: 1px solid #334155;
}

.summary-type {
  color: #cbd5e1;
  font-size: 0.85rem;
}

.summary-count {
  color: #f87171;
  font-weight: 600;
  font-size: 0.85rem;
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
  padding: 20px;
}

.modal-content {
  background: #1e293b;
  border-radius: 10px;
  width: 500px;
  max-width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #334155;
  position: sticky;
  top: 0;
  background: #1e293b;
  z-index: 1;
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
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #ef4444;
  color: white;
}

.modal-body {
  padding: 20px;
}

.detail-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-row {
  display: flex;
  padding: 8px 0;
  border-bottom: 1px solid #334155;
}

.detail-label {
  width: 110px;
  color: #94a3b8;
  font-size: 0.9rem;
  flex-shrink: 0;
}

.detail-value {
  flex: 1;
  color: #e2e8f0;
  font-size: 0.9rem;
  word-break: break-word;
}

.detail-value.mac {
  font-family: 'Monaco', 'Courier New', monospace;
}

.detail-value.online {
  color: #34d399;
}

.detail-value.offline {
  color: #f87171;
}

.detail-value.blocked {
  color: #f87171;
}

.detail-value.unblocked {
  color: #34d399;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px 20px;
  border-top: 1px solid #334155;
  position: sticky;
  bottom: 0;
  background: #1e293b;
}

.modal-toggle-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
  color: white;
}

.modal-toggle-btn.block {
  background: #ef4444;
}

.modal-toggle-btn.block:hover {
  background: #dc2626;
}

.modal-toggle-btn.unblock {
  background: #10b981;
}

.modal-toggle-btn.unblock:hover {
  background: #059669;
}

.modal-btn {
  padding: 8px 16px;
  border: 1px solid #334155;
  border-radius: 6px;
  background: #1e293b;
  color: #cbd5e1;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
}

.modal-btn.close:hover {
  border-color: #3b82f6;
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #94a3b8;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
  opacity: 0.5;
}

.empty-state h3 {
  margin: 0 0 4px 0;
  color: #f8fafc;
  font-size: 1.2rem;
}

.empty-state p {
  margin: 0;
  font-size: 0.9rem;
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
</style>