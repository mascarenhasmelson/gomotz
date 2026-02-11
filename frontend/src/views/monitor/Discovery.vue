<template>
  <div class="lan-scanner">
    <!-- Header -->
    

    <!-- Controls Bar -->
    <div class="controls-bar">
      <!-- VLAN Filter -->
      <div class="vlan-filter">
        <label for="vlan-select">VLAN:</label>
        <select 
          id="vlan-select" 
          v-model="selectedVlan" 
          @change="filterByVlan"
          class="vlan-select"
          :disabled="!connected"
        >
          <option value="all">All VLANs</option>
          <option v-for="vlan in vlanList" :key="vlan" :value="vlan">
            VLAN {{ vlan }}
          </option>
        </select>
      </div>

      <!-- Search Bar -->
      <div class="search-bar">
        <span class="search-icon">🔍</span>
        <input 
          v-model="searchQuery" 
          type="text" 
          placeholder="Search by IP, MAC, hostname, or vendor..."
          class="search-input"
        />
      </div>

      <!-- Connection Status -->
      <div class="connection-status" :class="{ connected, disconnected: !connected }">
        {{ connected ? '🟢 Connected' : '🔴 Disconnected' }}
      </div>
    </div>

    <!-- Stats Bar -->
    <div class="stats-bar">
      <div class="stat-item">
        <span class="stat-label">Total Devices:</span>
        <span class="stat-value">{{ filteredDevices.length }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">Online:</span>
        <span class="stat-value online">{{ onlineCount }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">Offline:</span>
        <span class="stat-value offline">{{ offlineCount }}</span>
      </div>
      <div class="stat-item" v-if="lastUpdateTime">
        <span class="stat-label">Last Update:</span>
        <span class="stat-value">{{ lastUpdateTime }}</span>
      </div>
    </div>

    <!-- Connection Error -->
    <div v-if="connectionError" class="error-message">
      <span class="error-icon">⚠️</span>
      <span>{{ connectionError }}</span>
      <button @click="reconnect" class="retry-btn">Reconnect</button>
    </div>

    <!-- Loading State -->
    <div v-if="loading && devices.length === 0" class="loading">
      <div class="loading-spinner"></div>
      <p>Connecting to server...</p>
    </div>

    <!-- Device List -->
    <div v-else class="devices-list">
      <div 
        v-for="device in paginatedDevices" 
        :key="device.mac" 
        class="device-card"
        :class="{ online: device.status === 'online', offline: device.status === 'offline', new: device.isNew }"
        @click="showDeviceDetails(device)"
      >
        <!-- Status Indicator -->
        <div class="status-indicator" :class="device.status">
          {{ device.status === 'online' ? '🟢' : '🔴' }}
        </div>

        <!-- Device Content -->
        <div class="device-content">
          <div class="device-header">
            <div class="device-name-section">
              <span class="device-name">{{ device.hostname || 'Generic' }}</span>
              <span v-if="device.deviceType" class="device-type">{{ device.deviceType }}</span>
            </div>
            <span class="device-vlan" v-if="device.vlan">VLAN {{ device.vlan }}</span>
          </div>

          <div class="device-details">
            <div class="detail-row">
              <span class="detail-label">IP:</span>
              <span class="detail-value ip">{{ device.ip }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">MAC:</span>
              <span class="detail-value mac">{{ formatMAC(device.mac) }}</span>
              <button @click.stop="copyToClipboard(device.mac)" class="copy-btn" title="Copy MAC address">
                📋
              </button>
            </div>
            <div class="detail-row" v-if="device.vendor">
              <span class="detail-label">Vendor:</span>
              <span class="detail-value vendor">{{ device.vendor }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Last Seen:</span>
              <span class="detail-value last-seen">{{ formatTime(device.lastSeen) }}</span>
            </div>
          </div>
        </div>

        <!-- New Badge -->
        <div v-if="device.isNew" class="new-badge">NEW</div>
      </div>

      <!-- Pagination -->
      <div v-if="filteredDevices.length > 0" class="pagination">
        <button 
          @click="currentPage--" 
          :disabled="currentPage === 1"
          class="pagination-btn"
        >
          ←
        </button>
        <span class="page-info">
          Page {{ currentPage }} of {{ totalPages }}
        </span>
        <button 
          @click="currentPage++" 
          :disabled="currentPage === totalPages"
          class="pagination-btn"
        >
          →
        </button>
        <select v-model="itemsPerPage" class="items-per-page">
          <option :value="10">10 per page</option>
          <option :value="25">25 per page</option>
          <option :value="50">50 per page</option>
        </select>
      </div>

      <!-- No Devices Message -->
      <div v-if="filteredDevices.length === 0 && !loading" class="no-devices">
        <div class="no-devices-icon">📡</div>
        <h3>No Devices Found</h3>
        <p v-if="searchQuery">No devices match your search criteria</p>
        <p v-else-if="selectedVlan !== 'all'">No devices in VLAN {{ selectedVlan }}</p>
        <p v-else>{{ connected ? 'Waiting for devices...' : 'Unable to connect to server' }}</p>
      </div>
    </div>

    <!-- Device Details Modal -->
    <div v-if="selectedDevice" class="modal-overlay" @click.self="closeModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ selectedDevice.hostname || 'Generic Device' }}</h3>
          <button @click="closeModal" class="close-btn">&times;</button>
        </div>
        <div class="modal-body">
          <div class="device-details-grid">
            <div class="detail-row">
              <span class="detail-label">Status:</span>
              <span class="detail-value" :class="selectedDevice.status">
                {{ selectedDevice.status === 'online' ? '🟢 Online' : '🔴 Offline' }}
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">IP Address:</span>
              <span class="detail-value ip">{{ selectedDevice.ip }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">MAC Address:</span>
              <span class="detail-value mac">{{ formatMAC(selectedDevice.mac) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Hostname:</span>
              <span class="detail-value">{{ selectedDevice.hostname || 'Generic' }}</span>
            </div>
            <div class="detail-row" v-if="selectedDevice.deviceType">
              <span class="detail-label">Device Type:</span>
              <span class="detail-value">{{ selectedDevice.deviceType }}</span>
            </div>
            <div class="detail-row" v-if="selectedDevice.vendor">
              <span class="detail-label">Vendor:</span>
              <span class="detail-value">{{ selectedDevice.vendor }}</span>
            </div>
            <div class="detail-row" v-if="selectedDevice.vlan">
              <span class="detail-label">VLAN:</span>
              <span class="detail-value">VLAN {{ selectedDevice.vlan }}</span>
            </div>
            <div class="detail-row" v-if="selectedDevice.ports && selectedDevice.ports.length">
              <span class="detail-label">Open Ports:</span>
              <span class="detail-value">
                <span v-for="port in selectedDevice.ports" :key="port" class="port-badge">
                  {{ port }}
                </span>
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">First Seen:</span>
              <span class="detail-value">{{ formatFullTime(selectedDevice.firstSeen) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Last Seen:</span>
              <span class="detail-value">{{ formatFullTime(selectedDevice.lastSeen) }}</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="copyToClipboard(selectedDevice.mac)" class="modal-btn copy">
            📋 Copy MAC
          </button>
          <button @click="copyToClipboard(selectedDevice.ip)" class="modal-btn copy">
            📋 Copy IP
          </button>
          <button @click="closeModal" class="modal-btn close">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'

export default {
  name: 'LanScanner',
  
  setup() {
    // State
    const devices = ref([])
    const loading = ref(true)
    const connected = ref(false)
    const connectionError = ref(null)
    const selectedVlan = ref('all')
    const searchQuery = ref('')
    const currentPage = ref(1)
    const itemsPerPage = ref(25)
    const lastUpdateTime = ref(null)
    const selectedDevice = ref(null)
    const vlanList = ref([])
    
    // WebSocket connection
    let ws = null
    const wsReconnectAttempts = ref(0)
    const maxReconnectAttempts = ref(10)
    
    // Computed properties
    const filteredDevices = computed(() => {
      let filtered = [...devices.value]
      
      // Filter by VLAN
      if (selectedVlan.value !== 'all') {
        filtered = filtered.filter(d => d.vlan == selectedVlan.value)
      }
      
      // Filter by search query
      if (searchQuery.value) {
        const query = searchQuery.value.toLowerCase()
        filtered = filtered.filter(d => 
          d.ip?.toLowerCase().includes(query) ||
          d.mac?.toLowerCase().includes(query) ||
          d.hostname?.toLowerCase().includes(query) ||
          d.vendor?.toLowerCase().includes(query) ||
          d.deviceType?.toLowerCase().includes(query)
        )
      }
      
      // Sort by IP address
      filtered.sort((a, b) => {
        const ipA = a.ip.split('.').map(Number)
        const ipB = b.ip.split('.').map(Number)
        for (let i = 0; i < 4; i++) {
          if (ipA[i] !== ipB[i]) return ipA[i] - ipB[i]
        }
        return 0
      })
      
      return filtered
    })
    
    const onlineCount = computed(() => {
      return devices.value.filter(d => d.status === 'online').length
    })
    
    const offlineCount = computed(() => {
      return devices.value.filter(d => d.status === 'offline').length
    })
    
    const totalPages = computed(() => {
      return Math.ceil(filteredDevices.value.length / itemsPerPage.value)
    })
    
    const paginatedDevices = computed(() => {
      const start = (currentPage.value - 1) * itemsPerPage.value
      const end = start + itemsPerPage.value
      return filteredDevices.value.slice(start, end)
    })
    
    // Methods
    const connectWebSocket = () => {
      loading.value = true
      connectionError.value = null
      
      ws = new WebSocket('ws://localhost:8082/ws/lan-scanner')
      
      ws.onopen = () => {
        console.log('WebSocket connected')
        connected.value = true
        loading.value = false
        wsReconnectAttempts.value = 0
        connectionError.value = null
      }
      
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          handleWebSocketMessage(data)
          lastUpdateTime.value = new Date().toLocaleTimeString()
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }
      
      ws.onclose = () => {
        console.log('WebSocket disconnected')
        connected.value = false
        loading.value = false
        
        if (wsReconnectAttempts.value < maxReconnectAttempts.value) {
          setTimeout(() => {
            wsReconnectAttempts.value++
            connectWebSocket()
          }, 3000)
        } else {
          connectionError.value = 'Maximum reconnection attempts reached'
        }
      }
      
      ws.onerror = () => {
        connectionError.value = 'Failed to connect to device server'
        connected.value = false
        loading.value = false
      }
    }
    
    const handleWebSocketMessage = (data) => {
      switch (data.type) {
        case 'device_discovered':
          addOrUpdateDevice({ ...data.device, isNew: true })
          break
          
        case 'device_update':
          addOrUpdateDevice({ ...data.device, isUpdated: true })
          break
          
        case 'devices_batch':
          if (data.devices && Array.isArray(data.devices)) {
            devices.value = data.devices.map(device => ({
              ...device,
              isNew: false,
              isUpdated: false
            }))
          }
          break
          
        case 'vlan_devices':
          if (data.vlan && data.devices) {
            if (selectedVlan.value == data.vlan) {
              devices.value = data.devices.map(device => ({
                ...device,
                isNew: false,
                isUpdated: false
              }))
            }
          }
          break
      }
      
      // Update VLAN list
      updateVlanList()
    }
    
    const addOrUpdateDevice = (device) => {
      if (!device.mac) return
      
      const index = devices.value.findIndex(d => d.mac === device.mac)
      
      if (index === -1) {
        // New device
        device.firstSeen = device.firstSeen || new Date().toISOString()
        device.lastSeen = new Date().toISOString()
        devices.value.unshift(device)
        
        // Remove new flag after 3 seconds
        setTimeout(() => {
          const idx = devices.value.findIndex(d => d.mac === device.mac)
          if (idx !== -1) {
            devices.value[idx].isNew = false
          }
        }, 3000)
      } else {
        // Update existing device
        devices.value[index] = {
          ...devices.value[index],
          ...device,
          lastSeen: new Date().toISOString(),
          firstSeen: devices.value[index].firstSeen || device.firstSeen || new Date().toISOString()
        }
        
        // Remove updated flag after 3 seconds
        setTimeout(() => {
          const idx = devices.value.findIndex(d => d.mac === device.mac)
          if (idx !== -1) {
            devices.value[idx].isUpdated = false
          }
        }, 3000)
      }
      
      updateVlanList()
    }
    
    const updateVlanList = () => {
      const vlans = new Set()
      devices.value.forEach(device => {
        if (device.vlan) vlans.add(device.vlan)
      })
      vlanList.value = Array.from(vlans).sort((a, b) => a - b)
    }
    
    const filterByVlan = () => {
      if (selectedVlan.value !== 'all' && ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'get_vlan_devices',
          vlan: parseInt(selectedVlan.value)
        }))
      }
      currentPage.value = 1
    }
    
    const formatMAC = (mac) => {
      if (!mac) return 'Unknown'
      const cleanMac = mac.replace(/[^a-fA-F0-9]/g, '')
      if (cleanMac.length === 12) {
        return cleanMac.match(/.{2}/g).join(':').toLowerCase()
      }
      return mac.toLowerCase()
    }
    
    const formatTime = (timestamp) => {
      if (!timestamp) return 'Never'
      const date = new Date(timestamp)
      const now = new Date()
      const diffMs = now - date
      const diffMins = Math.floor(diffMs / 60000)
      
      if (diffMins < 1) return 'Just now'
      if (diffMins < 60) return `${diffMins}m ago`
      if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`
      return date.toLocaleDateString()
    }
    
    const formatFullTime = (timestamp) => {
      if (!timestamp) return 'Never'
      return new Date(timestamp).toLocaleString()
    }
    
    const copyToClipboard = (text) => {
      navigator.clipboard.writeText(text)
        .then(() => {
          // Show temporary tooltip or notification
          console.log('Copied to clipboard:', text)
        })
        .catch(err => console.error('Failed to copy:', err))
    }
    
    const showDeviceDetails = (device) => {
      selectedDevice.value = device
    }
    
    const closeModal = () => {
      selectedDevice.value = null
    }
    
    const reconnect = () => {
      wsReconnectAttempts.value = 0
      if (ws) {
        ws.close()
      }
      connectWebSocket()
    }
    
    // Watchers
    watch(searchQuery, () => {
      currentPage.value = 1
    })
    
    watch(selectedVlan, () => {
      currentPage.value = 1
    })
    
    watch(itemsPerPage, () => {
      currentPage.value = 1
    })
    
    watch(filteredDevices, () => {
      if (currentPage.value > totalPages.value) {
        currentPage.value = totalPages.value || 1
      }
    })
    
    // Lifecycle
    onMounted(() => {
      connectWebSocket()
    })
    
    onBeforeUnmount(() => {
      if (ws) {
        ws.close()
      }
    })
    
    return {
      devices,
      loading,
      connected,
      connectionError,
      selectedVlan,
      searchQuery,
      currentPage,
      itemsPerPage,
      lastUpdateTime,
      selectedDevice,
      vlanList,
      filteredDevices,
      onlineCount,
      offlineCount,
      totalPages,
      paginatedDevices,
      filterByVlan,
      formatMAC,
      formatTime,
      formatFullTime,
      copyToClipboard,
      showDeviceDetails,
      closeModal,
      reconnect
    }
  }
}
</script>

<style scoped>
/* Dark Mode Theme */
.lan-scanner {
  padding: 30px;
  max-width: 1200px;
  margin: 0 auto;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  min-height: 100vh;
  color: #e2e8f0;
}

/* Header */
.scanner-header {
  margin-bottom: 25px;
}

.scanner-header h1 {
  font-size: 28px;
  font-weight: 700;
  margin: 0 0 5px 0;
  letter-spacing: -0.5px;
  background: linear-gradient(135deg, #60a5fa 0%, #a78bfa 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.subtitle {
  color: #94a3b8;
  font-size: 15px;
  margin: 0;
}

/* Controls Bar */
.controls-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px 20px;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  flex-wrap: wrap;
  gap: 15px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.2);
}

.vlan-filter {
  display: flex;
  align-items: center;
  gap: 10px;
}

.vlan-filter label {
  font-size: 14px;
  color: #cbd5e1;
  font-weight: 500;
}

.vlan-select {
  padding: 8px 14px;
  border: 1px solid #334155;
  border-radius: 8px;
  font-size: 14px;
  color: #e2e8f0;
  background: #0f172a;
  outline: none;
  min-width: 130px;
  transition: all 0.2s;
}

.vlan-select:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

.vlan-select:disabled {
  background-color: #1e293b;
  cursor: not-allowed;
  opacity: 0.7;
}

.search-bar {
  flex: 1;
  max-width: 400px;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: #64748b;
  font-size: 16px;
}

.search-input {
  width: 100%;
  padding: 10px 12px 10px 40px;
  border: 1px solid #334155;
  border-radius: 8px;
  font-size: 14px;
  color: #e2e8f0;
  background: #0f172a;
  outline: none;
  transition: all 0.2s;
}

.search-input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

.search-input::placeholder {
  color: #64748b;
}

.connection-status {
  font-size: 13px;
  padding: 6px 14px;
  border-radius: 20px;
  background: #1e293b;
  font-weight: 500;
  border: 1px solid #334155;
}

.connection-status.connected {
  color: #34d399;
  background: rgba(16, 185, 129, 0.1);
  border-color: rgba(16, 185, 129, 0.3);
}

.connection-status.disconnected {
  color: #f87171;
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
}

/* Stats Bar */
.stats-bar {
  display: flex;
  gap: 25px;
  margin-bottom: 25px;
  padding: 16px 20px;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  flex-wrap: wrap;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.2);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-label {
  font-size: 14px;
  color: #94a3b8;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: #f8fafc;
}

.stat-value.online {
  color: #34d399;
}

.stat-value.offline {
  color: #f87171;
}

/* Error Message */
.error-message {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 10px;
  padding: 14px 18px;
  margin-bottom: 25px;
  display: flex;
  align-items: center;
  gap: 12px;
  color: #fca5a5;
  font-size: 14px;
  backdrop-filter: blur(10px);
}

.error-icon {
  font-size: 18px;
}

.retry-btn {
  margin-left: auto;
  padding: 6px 16px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.retry-btn:hover {
  background: #dc2626;
  transform: translateY(-1px);
  box-shadow: 0 4px 8px rgba(239, 68, 68, 0.3);
}

/* Loading State */
.loading {
  text-align: center;
  padding: 60px 20px;
  color: #94a3b8;
  font-size: 15px;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #1e293b;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 15px auto;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Device List */
.devices-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.device-card {
  display: flex;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
  padding: 0;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  overflow: hidden;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.2);
}

.device-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  border-color: rgba(59, 130, 246, 0.3);
}

.device-card.online {
  border-left: 4px solid #10b981;
}

.device-card.offline {
  border-left: 4px solid #ef4444;
  opacity: 0.7;
}

.device-card.new {
  animation: highlight-new 2s ease-out;
}

@keyframes highlight-new {
  0% { background: rgba(16, 185, 129, 0.2); }
  100% { background: rgba(30, 41, 59, 0.8); }
}

.status-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  font-size: 20px;
  background: rgba(15, 23, 42, 0.6);
  border-right: 1px solid rgba(148, 163, 184, 0.1);
}

.device-content {
  flex: 1;
  padding: 18px 20px;
}

.device-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.device-name-section {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.device-name {
  font-weight: 600;
  font-size: 16px;
  color: #f8fafc;
}

.device-type {
  font-size: 12px;
  color: #94a3b8;
  background: #0f172a;
  padding: 3px 10px;
  border-radius: 12px;
  font-weight: 500;
  border: 1px solid #334155;
}

.device-vlan {
  font-size: 12px;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  padding: 4px 12px;
  border-radius: 12px;
  font-weight: 600;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.device-details {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 12px 20px;
}

.detail-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.detail-label {
  color: #94a3b8;
  min-width: 70px;
}

.detail-value {
  color: #e2e8f0;
  font-weight: 500;
}

.detail-value.ip {
  color: #60a5fa;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
}

.detail-value.mac {
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  color: #94a3b8;
}

.detail-value.vendor {
  color: #c4b5fd;
}

.detail-value.last-seen {
  color: #94a3b8;
  font-size: 12px;
}

.copy-btn {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  font-size: 14px;
  transition: all 0.2s;
  margin-left: 4px;
}

.copy-btn:hover {
  background: #1e293b;
  color: #60a5fa;
}

.new-badge {
  position: absolute;
  top: 12px;
  right: 12px;
  background: #10b981;
  color: white;
  font-size: 11px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 12px;
  letter-spacing: 0.5px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

/* Pagination */
.pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 16px;
  margin-top: 25px;
  padding-top: 20px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
  flex-wrap: wrap;
}

.pagination-btn {
  background: #0f172a;
  border: 1px solid #334155;
  color: #cbd5e1;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.pagination-btn:hover:not(:disabled) {
  background: #1e293b;
  border-color: #3b82f6;
  color: #60a5fa;
  transform: translateY(-1px);
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  color: #94a3b8;
  font-size: 14px;
}

.items-per-page {
  padding: 8px 12px;
  border: 1px solid #334155;
  border-radius: 8px;
  font-size: 13px;
  color: #e2e8f0;
  background: #0f172a;
  cursor: pointer;
  outline: none;
}

/* No Devices */
.no-devices {
  text-align: center;
  padding: 80px 20px;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px dashed rgba(148, 163, 184, 0.2);
}

.no-devices-icon {
  font-size: 48px;
  margin-bottom: 20px;
  opacity: 0.7;
}

.no-devices h3 {
  font-size: 20px;
  font-weight: 600;
  color: #f8fafc;
  margin-bottom: 10px;
}

.no-devices p {
  color: #94a3b8;
  font-size: 15px;
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
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background: #1e293b;
  border-radius: 16px;
  width: 550px;
  max-width: 90vw;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.1);
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #f8fafc;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: #94a3b8;
  cursor: pointer;
  padding: 0;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #ef4444;
  color: white;
}

.modal-body {
  padding: 24px;
}

.device-details-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-body .detail-row {
  display: flex;
  padding: 12px 0;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  margin: 0;
}

.modal-body .detail-row:last-child {
  border-bottom: none;
}

.modal-body .detail-label {
  width: 110px;
  font-size: 14px;
  color: #94a3b8;
}

.modal-body .detail-value {
  flex: 1;
  font-size: 14px;
  color: #e2e8f0;
  font-weight: 500;
}

.modal-body .detail-value.ip {
  color: #60a5fa;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
}

.modal-body .detail-value.mac {
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  color: #94a3b8;
}

.modal-body .detail-value.online {
  color: #34d399;
}

.modal-body .detail-value.offline {
  color: #f87171;
}

.port-badge {
  display: inline-block;
  background: #0f172a;
  color: #94a3b8;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  margin-right: 8px;
  margin-bottom: 8px;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  border: 1px solid #334155;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px 24px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
}

.modal-btn {
  padding: 10px 18px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #334155;
  background: #0f172a;
  color: #cbd5e1;
}

.modal-btn:hover {
  background: #1e293b;
  border-color: #475569;
  transform: translateY(-1px);
}

.modal-btn.copy {
  background: rgba(59, 130, 246, 0.1);
  color: #60a5fa;
  border-color: rgba(59, 130, 246, 0.3);
}

.modal-btn.copy:hover {
  background: rgba(59, 130, 246, 0.2);
  border-color: #3b82f6;
}

.modal-btn.close {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.modal-btn.close:hover {
  background: #2563eb;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

/* Responsive */
@media (max-width: 768px) {
  .lan-scanner {
    padding: 20px;
  }
  
  .controls-bar {
    flex-direction: column;
    align-items: stretch;
  }
  
  .search-bar {
    max-width: 100%;
  }
  
  .stats-bar {
    gap: 15px;
  }
  
  .device-details {
    grid-template-columns: 1fr;
  }
  
  .device-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .device-vlan {
    align-self: flex-start;
  }
  
  .status-indicator {
    width: 40px;
  }
  
  .pagination {
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .device-card {
    flex-direction: column;
  }
  
  .status-indicator {
    width: 100%;
    padding: 8px;
    border-right: none;
    border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  }
  
  .device-content {
    padding: 16px;
  }
}

/* Scrollbar Styling */
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