<template>
  <div class="lan-scanner">
    <!-- Top Bar -->
    <div class="top-bar">
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
    </div>

    <!-- Main Content Area - Split into two panels -->
    <div class="main-content">
      <!-- Left Panel - Device List -->
      <div class="device-list-panel">
        <div class="panel-header">
          <h2>Network Devices</h2>
          <span class="device-count">{{ filteredDevices.length }} devices</span>
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
            :class="{ 
              online: device.status === 'online', 
              offline: device.status === 'offline', 
              new: device.isNew,
              selected: selectedDevice && selectedDevice.mac === device.mac
            }"
            @click="selectDevice(device)"
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
                <!-- <span class="device-vlan" v-if="device.vlan">VLAN {{ device.vlan }}</span> -->
              </div>

              <div class="device-details">
                <div class="detail-row">
                  <span class="detail-label">IP:</span>
                  <span class="detail-value ip">{{ device.ip }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">MAC:</span>
                  <span class="detail-value mac">{{ formatMAC(device.mac) }}</span>
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
      </div>

      <!-- Right Panel - Selected Device Details (Fixed) -->
      <div class="device-details-panel" :class="{ 'has-device': selectedDevice }">
        <div class="panel-header">
          <h2>Device Details</h2>
          <button v-if="selectedDevice" @click="clearSelection" class="clear-btn">×</button>
        </div>

        <!-- Empty State -->
        <div v-if="!selectedDevice" class="no-selection">
          <div class="no-selection-icon">👆</div>
          <h3>No Device Selected</h3>
          <p>Click on any device from the list to view details</p>
        </div>

        <!-- Selected Device Details -->
        <div v-else class="selected-device-details">
          <div class="device-status-banner" :class="selectedDevice.status">
            <span class="status-icon">{{ selectedDevice.status === 'online' ? '🟢' : '🔴' }}</span>
            <span class="status-text">{{ selectedDevice.status === 'online' ? 'Online' : 'Offline' }}</span>
          </div>

          <div class="details-content">
            <div class="detail-section">
              <h3>Basic Information</h3>
              <div class="detail-grid">
                <div class="detail-item">
                  <span class="detail-label">Hostname:</span>
                  <span class="detail-value">{{ selectedDevice.hostname || 'Generic' }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">IP Address:</span>
                  <span class="detail-value ip">{{ selectedDevice.ip }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">MAC Address:</span>
                  <div class="value-with-copy">
                    <span class="detail-value mac">{{ formatMAC(selectedDevice.mac) }}</span>
                    <button @click="copyToClipboard(selectedDevice.mac)" class="copy-btn" title="Copy MAC">
                      📋
                    </button>
                  </div>
                </div>
                <div class="detail-item" v-if="selectedDevice.deviceType">
                  <span class="detail-label">Device Type:</span>
                  <span class="detail-value">{{ selectedDevice.deviceType }}</span>
                </div>
                <div class="detail-item" v-if="selectedDevice.vendor">
                  <span class="detail-label">Vendor:</span>
                  <span class="detail-value">{{ selectedDevice.vendor }}</span>
                </div>
                <div class="detail-item" v-if="selectedDevice.vlan">
                  <span class="detail-label">VLAN:</span>
                  <span class="detail-value vlan-badge">VLAN {{ selectedDevice.vlan }}</span>
                </div>
              </div>
            </div>

            <div class="detail-section">
              <h3>Timing Information</h3>
              <div class="detail-grid">
                <div class="detail-item">
                  <span class="detail-label">First Seen:</span>
                  <span class="detail-value">{{ formatFullTime(selectedDevice.firstSeen) }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">Last Seen:</span>
                  <span class="detail-value">{{ formatFullTime(selectedDevice.lastSeen) }}</span>
                </div>
              </div>
            </div>

            <div class="detail-section" v-if="selectedDevice.ports && selectedDevice.ports.length">
              <h3>Open Ports</h3>
              <div class="ports-container">
                <span v-for="port in selectedDevice.ports" :key="port" class="port-badge">
                  {{ port }}
                </span>
              </div>
            </div>

            <div class="action-buttons">
              <button @click="copyToClipboard(selectedDevice.ip)" class="action-btn copy-ip">
                📋 Copy IP
              </button>
              <button @click="copyToClipboard(selectedDevice.mac)" class="action-btn copy-mac">
                📋 Copy MAC
              </button>
              <button @click="pingDevice(selectedDevice)" class="action-btn ping">
                📡 Ping
              </button>
              <button @click="scanPorts(selectedDevice)" class="action-btn scan">
                🔍 Scan Ports
              </button>
            </div>
          </div>
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
      const wasSelected = selectedDevice.value && selectedDevice.value.mac === device.mac
      
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
        
        // Update selected device if it's the same
        if (wasSelected) {
          selectedDevice.value = devices.value[index]
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
          console.log('Copied to clipboard:', text)
        })
        .catch(err => console.error('Failed to copy:', err))
    }
    
    const selectDevice = (device) => {
      selectedDevice.value = device
    }
    
    const clearSelection = () => {
      selectedDevice.value = null
    }
    
    const pingDevice = (device) => {
      console.log('Pinging device:', device.ip)
      // Implement ping functionality
    }
    
    const scanPorts = (device) => {
      console.log('Scanning ports for device:', device.ip)
      // Implement port scan functionality
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
      
      // Clear selection if selected device is no longer in filtered list
      if (selectedDevice.value) {
        const stillExists = filteredDevices.value.some(d => d.mac === selectedDevice.value.mac)
        if (!stillExists) {
          selectedDevice.value = null
        }
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
      selectDevice,
      clearSelection,
      pingDevice,
      scanPorts,
      reconnect
    }
  }
}
</script>

<style scoped>
/* Dark Mode Theme */
.lan-scanner {
  display: flex;
  flex-direction: column;
  height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  color: #e2e8f0;
  overflow: hidden;
}

/* Top Bar */
.top-bar {
  padding: 20px 20px 10px 20px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
}

/* Controls Bar */
.controls-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
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
  padding: 12px 20px;
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

/* Main Content - Split Panels */
.main-content {
  display: flex;
  flex: 1;
  overflow: hidden;
  padding: 0 20px 20px 20px;
  gap: 20px;
  min-height: 0;
}

/* Left Panel - Device List */
.device-list-panel {
  flex: 1;
  background: rgba(30, 41, 59, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  min-width: 600px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  background: rgba(15, 23, 42, 0.6);
}

.panel-header h2 {
  margin: 0;
  font-size: 1.4rem;
  font-weight: 600;
  color: #f8fafc;
}

.device-count {
  font-size: 1rem;
  color: #94a3b8;
  background: #0f172a;
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid #334155;
  font-weight: 500;
}

/* Error Message */
.error-message {
  margin: 20px;
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.4);
  border-radius: 12px;
  padding: 18px 22px;
  display: flex;
  align-items: center;
  gap: 15px;
  color: #fca5a5;
  font-size: 15px;
}

.error-icon {
  font-size: 22px;
}

.retry-btn {
  margin-left: auto;
  padding: 8px 20px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.retry-btn:hover {
  background: #dc2626;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

/* Loading State */
.loading {
  text-align: center;
  padding: 80px 30px;
  color: #94a3b8;
  font-size: 16px;
}

.loading-spinner {
  width: 50px;
  height: 50px;
  border: 4px solid #1e293b;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 20px auto;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Device List Container */
.devices-list {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 9px;
}

/* Device Cards */
.device-card {
  display: flex;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 14px;
  padding: 0;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  overflow: hidden;
  min-height: 120px;
}

.device-card:hover {
  transform: translateX(4px);
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  background: rgba(15, 23, 42, 0.95);
}

.device-card.selected {
  border: 2px solid #3b82f6;
  background: rgba(59, 130, 246, 0.15);
  box-shadow: 0 0 20px rgba(59, 130, 246, 0.2);
}

.device-card.online {
  border-left: 6px solid #10b981;
}

.device-card.offline {
  border-left: 6px solid #ef4444;
  opacity: 0.8;
}

.device-card.new {
  animation: highlight-new 2s ease-out;
}

@keyframes highlight-new {
  0% { background: rgba(16, 185, 129, 0.3); }
  100% { background: rgba(15, 23, 42, 0.8); }
}

/* Status Indicator */
.status-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 60px;
  font-size: 24px;
  background: rgba(0, 0, 0, 0.3);
  border-right: 1px solid rgba(148, 163, 184, 0.1);
}

/* Device Content */
.device-content {
  flex: 1;
  padding: 12px 15px;
}

.device-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.device-name-section {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.device-name {
  font-weight: 700;
  font-size: 16px;
  color: #f8fafc;
}

.device-type {
  font-size: 11px;
  color: #94a3b8;
  background: #0f172a;
  padding: 3px 8px;
  border-radius: 12px;
  border: 1px solid #334155;
  font-weight: 500;
}

.device-vlan {
  font-size: 11px;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  padding: 3px 8px;
  border-radius: 12px;
  font-weight: 600;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  white-space: nowrap;
}

/* Device Details Grid */
.device-details {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 6px 8px;
  margin-top: 4px;
}

.detail-row {
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: 12px;
  background: rgba(0, 0, 0, 0.2);
  padding: 4px 6px;
  border-radius: 6px;
  border: 1px solid rgba(148, 163, 184, 0.05);
  min-width: 0;
  width: 100%;
}

.detail-label {
  color: #94a3b8;
  min-width: 35px;
  font-size: 11px;
  font-weight: 300;
  white-space: nowrap;
  flex-shrink: 0;
}

.detail-value {
  color: #e2e8f0;
  font-weight: 500;
  font-size: 11px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  min-width: 0;
}

.detail-value.ip {
  color: #60a5fa;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  font-size: 11px;
  font-weight: 600;
}

.detail-value.mac {
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  color: #94a3b8;
  font-size: 11px;
  letter-spacing: 0.2px;
  overflow: visible;
  text-overflow: clip;
  white-space: nowrap;
}

.detail-value.vendor {
  color: #c4b5fd;
  font-size: 11px;
  font-weight: 600;
}

.detail-value.last-seen {
  color: #94a3b8;
  font-size: 11px;
}

/* Make MAC row span both columns for better visibility */
.detail-row.mac-row {
  grid-column: span 2;
  width: 100%;
  background: rgba(15, 23, 42, 0.9);
}

/* New Badge */
.new-badge {
  position: absolute;
  top: 12px;
  right: 12px;
  background: #10b981;
  color: white;
  font-size: 12px;
  font-weight: 700;
  padding: 4px 12px;
  border-radius: 20px;
  letter-spacing: 0.5px;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

/* Pagination */
.pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 15px;
  margin-top: 20px;
  padding: 16px 20px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
  background: rgba(15, 23, 42, 0.4);
}

.pagination-btn {
  background: #0f172a;
  border: 1px solid #334155;
  color: #cbd5e1;
  padding: 10px 18px;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  min-width: 45px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pagination-btn:hover:not(:disabled) {
  background: #1e293b;
  border-color: #3b82f6;
  color: #60a5fa;
  transform: translateY(-2px);
}

.pagination-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-info {
  color: #94a3b8;
  font-size: 15px;
  font-weight: 500;
}

.items-per-page {
  padding: 10px 14px;
  border: 1px solid #334155;
  border-radius: 10px;
  font-size: 14px;
  color: #e2e8f0;
  background: #0f172a;
  cursor: pointer;
  outline: none;
  font-weight: 500;
}

.items-per-page:hover {
  border-color: #3b82f6;
}

/* No Devices */
.no-devices {
  text-align: center;
  padding: 80px 30px;
  background: rgba(15, 23, 42, 0.6);
  border-radius: 14px;
  margin: 20px;
}

.no-devices-icon {
  font-size: 64px;
  margin-bottom: 25px;
  opacity: 0.7;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.no-devices h3 {
  font-size: 22px;
  font-weight: 600;
  color: #f8fafc;
  margin-bottom: 15px;
}

.no-devices p {
  color: #94a3b8;
  font-size: 16px;
  line-height: 1.5;
  max-width: 400px;
  margin: 0 auto;
}

/* Right Panel - Device Details - COMPACT VERSION */
.device-details-panel {
  width: 450px;
  min-width: 450px;
  background: rgba(30, 41, 59, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  transition: all 0.3s ease;
}

.device-details-panel.has-device {
  border-color: #3b82f6;
  box-shadow: 0 0 30px rgba(59, 130, 246, 0.3);
}

/* No Selection State - Compact */
.no-selection {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  text-align: center;
}

.no-selection-icon {
  font-size: 60px;
  margin-bottom: 20px;
  opacity: 0.7;
  animation: bounce 2s infinite;
}

.no-selection h3 {
  font-size: 18px;
  font-weight: 600;
  color: #f8fafc;
  margin-bottom: 10px;
}

.no-selection p {
  color: #94a3b8;
  font-size: 14px;
  max-width: 250px;
  line-height: 1.4;
}

/* Selected Device Details - Compact */
.selected-device-details {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.device-status-banner {
  padding: 12px 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.device-status-banner.online {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
  border-bottom: 1px solid rgba(16, 185, 129, 0.3);
}

.device-status-banner.offline {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
  border-bottom: 1px solid rgba(239, 68, 68, 0.3);
}

.status-icon {
  font-size: 20px;
}

.status-text {
  font-size: 14px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 600;
}

.details-content {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-section {
  background: rgba(15, 23, 42, 0.8);
  border-radius: 12px;
  padding: 14px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.detail-section h3 {
  margin: 0 0 12px 0;
  font-size: 13px;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 600;
  border-bottom: 1px solid rgba(148, 163, 184, 0.2);
  padding-bottom: 6px;
}

.detail-grid {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.detail-item .detail-label {
  font-size: 11px;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.3px;
  font-weight: 500;
}

.detail-item .detail-value {
  font-size: 14px;
  color: #f8fafc;
  font-weight: 500;
  line-height: 1.3;
}

.detail-item .detail-value.ip {
  color: #60a5fa;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  font-size: 15px;
  font-weight: 600;
}

.detail-item .detail-value.mac {
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  color: #94a3b8;
  font-size: 14px;
  letter-spacing: 0.3px;
}

.value-with-copy {
  display: flex;
  align-items: center;
  gap: 8px;
}

.copy-btn {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 14px;
  transition: all 0.2s;
}

.copy-btn:hover {
  background: #1e293b;
  color: #60a5fa;
}

.vlan-badge {
  display: inline-block;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 13px;
  font-weight: 600;
  width: fit-content;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.3);
}

/* Open Ports Section - Compact */
.ports-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.port-badge {
  display: inline-block;
  background: #0f172a;
  color: #94a3b8;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 13px;
  font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
  border: 1px solid #334155;
  transition: all 0.2s;
  font-weight: 500;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.port-badge:hover {
  background: #1e293b;
  border-color: #3b82f6;
  color: #60a5fa;
  transform: translateY(-2px);
}

/* Action Buttons - Compact */
.action-buttons {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  margin-top: 16px;
}

.action-btn {
  padding: 10px 8px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #334155;
  background: #0f172a;
  color: #cbd5e1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.action-btn:hover {
  background: #1e293b;
  border-color: #475569;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.action-btn.copy-ip:hover {
  background: rgba(59, 130, 246, 0.2);
  border-color: #3b82f6;
  color: #60a5fa;
}

.action-btn.copy-mac:hover {
  background: rgba(139, 92, 246, 0.2);
  border-color: #8b5cf6;
  color: #c4b5fd;
}

.action-btn.ping:hover {
  background: rgba(16, 185, 129, 0.2);
  border-color: #10b981;
  color: #34d399;
}

.action-btn.scan:hover {
  background: rgba(245, 158, 11, 0.2);
  border-color: #f59e0b;
  color: #fbbf24;
}

/* Scrollbar Styling */
.devices-list::-webkit-scrollbar,
.selected-device-details::-webkit-scrollbar {
  width: 10px;
}

.devices-list::-webkit-scrollbar-track,
.selected-device-details::-webkit-scrollbar-track {
  background: #0f172a;
  border-radius: 5px;
}

.devices-list::-webkit-scrollbar-thumb,
.selected-device-details::-webkit-scrollbar-thumb {
  background: #334155;
  border-radius: 5px;
  border: 2px solid #0f172a;
}

.devices-list::-webkit-scrollbar-thumb:hover,
.selected-device-details::-webkit-scrollbar-thumb:hover {
  background: #475569;
}

/* Responsive adjustments */
@media (max-width: 1400px) {
  .device-list-panel {
    min-width: 500px;
  }
}

@media (max-width: 1200px) {
  .device-list-panel {
    min-width: 450px;
  }
  
  .device-name {
    font-size: 15px;
  }
  
  .device-details-panel {
    width: 400px;
    min-width: 400px;
  }
}

@media (max-width: 1024px) {
  .main-content {
    flex-direction: column;
  }
  
  .device-list-panel {
    min-width: 100%;
    height: 500px;
  }
  
  .device-details-panel {
    width: 100%;
    min-width: 100%;
    height: 500px;
  }
  
  .device-details {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
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
  
  .device-list-panel {
    height: 400px;
  }
  
  .device-card {
    min-height: 100px;
  }
  
  .status-indicator {
    width: 50px;
    font-size: 20px;
  }
  
  .device-details {
    grid-template-columns: 1fr;
  }
  
  .pagination {
    flex-wrap: wrap;
    justify-content: center;
  }
  
  .device-details-panel {
    height: 450px;
  }
  
  .detail-item .detail-value {
    font-size: 13px;
  }
  
  .detail-item .detail-value.ip,
  .detail-item .detail-value.mac {
    font-size: 13px;
  }
  
  .action-buttons {
    grid-template-columns: 1fr;
  }
}
</style>