<template>
  <div class="lan-scanner">

    <div class="top-bar">
      <div class="controls-bar">
        <div class="vlan-filter">
          <label for="network-select">Network:</label>
          <select
            id="network-select"
            v-model="selectedNetworkId"
            class="vlan-select"
            :disabled="!connected"
          >
            <option value="all">All Networks</option>
            <option v-for="network in networkList" :key="network.id" :value="network.id">
              {{ network.vlan_name }} ({{ network.interface_name }})
            </option>
          </select>
        </div>

        <div class="status-filter">
          <label>Status:</label>
          <div class="status-buttons">
            <button 
              @click="selectedStatus = 'all'" 
              class="status-btn"
              :class="{ active: selectedStatus === 'all' }"
            >
              All ({{ totalDevicesCount }})
            </button>
            <button 
              @click="selectedStatus = 'online'" 
              class="status-btn online-btn"
              :class="{ active: selectedStatus === 'online' }"
            >
              🟢 Online ({{ onlineCount }})
            </button>
            <button 
              @click="selectedStatus = 'offline'" 
              class="status-btn offline-btn"
              :class="{ active: selectedStatus === 'offline' }"
            >
              🔴 Offline ({{ offlineCount }})
            </button>
            <button 
              @click="selectedStatus = 'conflict'" 
              class="status-btn conflict-btn"
              :class="{ active: selectedStatus === 'conflict' }"
            >
              ⚠️ Conflict ({{ conflictCount }})
            </button>
          </div>
        </div>

        <div class="search-bar">
          <span class="search-icon">🔍</span>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by IP, MAC, hostname, or vendor..."
            class="search-input"
          />
        </div>

        <div class="connection-status" :class="{ connected, disconnected: !connected }">
          {{ connected ? '🟢 Connected' : '🔴 Disconnected' }}
        </div>
      </div>

      <div class="stats-bar">
        <div class="stat-item">
          <span class="stat-label">Total Devices:</span>
          <span class="stat-value">{{ totalDevicesCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Online:</span>
          <span class="stat-value online">{{ onlineCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Offline:</span>
          <span class="stat-value offline">{{ offlineCount }}</span>
        </div>
        <div class="stat-item" v-if="conflictCount > 0">
          <span class="stat-label conflict-label">⚠️ Conflicts:</span>
          <span class="stat-value conflict">{{ conflictCount }}</span>
        </div>
        <div class="stat-item" v-if="displayedDevicesCount < filteredDevicesCount">
          <span class="stat-label">Showing:</span>
          <span class="stat-value">{{ displayedDevicesCount }} / {{ filteredDevicesCount }}</span>
        </div>
      </div>

      <div v-if="conflictCount > 0" class="conflict-alert">
        <div class="alert-icon">🚨</div>
        <div class="alert-content">
          <strong>IP Conflict Detected!</strong>
          <span>{{ conflictCount }} device(s) have duplicate IP addresses. Please investigate immediately.</span>
        </div>
        <button @click="showConflictsPanel = !showConflictsPanel" class="alert-btn">
          {{ showConflictsPanel ? 'Hide Details' : 'View Details' }}
        </button>
      </div>
    </div>

    <div class="main-content">
      <div class="device-list-panel">
        <div class="panel-header">
          <h2>Network Devices</h2>
          <div class="header-actions">
            <button @click="refreshData" class="refresh-btn" title="Refresh devices">
              🔄
            </button>
            <span class="device-count">{{ filteredDevicesCount }} devices</span>
          </div>
        </div>

        <div v-if="connectionError" class="error-message">
          <span class="error-icon">⚠️</span>
          <span>{{ connectionError }}</span>
          <button @click="reconnect" class="retry-btn">Reconnect</button>
        </div>

        <div v-if="loading && devices.length === 0" class="loading">
          <div class="loading-spinner"></div>
          <p>Loading devices...</p>
        </div>

        <div 
          v-else 
          class="devices-list"
          ref="scrollContainer"
          @scroll="handleScroll"
        >
          <div
            v-for="device in displayedDevices"
            :key="deviceKey(device)"
            class="device-card"
            :class="{
              online: device.status === 'online',
              offline: device.status === 'offline',
              conflict: device.status === 'conflict',
              new: device.isNew,
              selected: selectedDevice && deviceKey(selectedDevice) === deviceKey(device)
            }"
            @click="selectDevice(device)"
          >
            <div class="status-indicator" :class="device.status">
              <span v-if="device.status === 'online'">🟢</span>
              <span v-else-if="device.status === 'offline'">🔴</span>
              <span v-else-if="device.status === 'conflict'">⚠️</span>
              <span v-else>❓</span>
            </div>

            <div class="device-content">
              <div class="device-header">
                <div class="device-name-section">
                  <span class="device-name">{{ device.hostname || 'Unknown' }}</span>
                  <span v-if="device.network_name" class="device-vlan">{{ device.network_name }}</span>
                  <span class="status-badge" :class="device.status">
                    {{ device.status === 'online' ? '🟢 ONLINE' : 
                       device.status === 'offline' ? '🔴 OFFLINE' : 
                       device.status === 'conflict' ? '⚠️ CONFLICT' : '❓ UNKNOWN' }}
                  </span>
                  <span v-if="device.status === 'conflict'" class="conflict-badge">DUPLICATE IP</span>
                </div>
              </div>

              <div class="device-details">
                <div class="detail-row">
                  <span class="detail-label">IP:</span>
                  <span class="detail-value ip">{{ device.ip_address }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">MAC:</span>
                  <span class="detail-value mac">{{ device.mac_address }}</span>
                </div>
                <div class="detail-row" v-if="device.vendor">
                  <span class="detail-label">Vendor:</span>
                  <span class="detail-value vendor">{{ device.vendor }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Last Seen:</span>
                  <span class="detail-value last-seen">{{ formatTime(device.last_seen) }}</span>
                </div>
              </div>
            </div>

            <div v-if="device.isNew" class="new-badge">NEW</div>
            <div v-if="device.status === 'conflict'" class="conflict-icon">⚠️</div>
          </div>

          <div v-if="hasMore && isLoadingMore" class="loading-more">
            <div class="loading-spinner-small"></div>
            <span>Loading more devices...</span>
          </div>

          <div v-if="!hasMore && filteredDevicesCount > 0" class="end-message">
            <span>✨ No more devices to load</span>
          </div>

          <div v-if="filteredDevicesCount === 0 && !loading" class="no-devices">
            <div class="no-devices-icon">📡</div>
            <h3>No Devices Found</h3>
            <p v-if="searchQuery">No devices match your search criteria</p>
            <p v-else-if="selectedNetworkId !== 'all'">No devices in selected network</p>
            <p v-else-if="selectedStatus !== 'all'">No {{ selectedStatus }} devices found</p>
            <p v-else>{{ connected ? 'Waiting for devices...' : 'Unable to connect to server' }}</p>
          </div>
        </div>
      </div>

      <div class="device-details-panel" :class="{ 'has-device': selectedDevice }">
        <div class="panel-header">
          <h2>Device Details</h2>
          <button v-if="selectedDevice" @click="clearSelection" class="clear-btn">×</button>
        </div>

        <div v-if="!selectedDevice" class="no-selection">
          <div class="no-selection-icon">👆</div>
          <h3>No Device Selected</h3>
          <p>Click on any device from the list to view details</p>
        </div>

        <div v-else class="selected-device-details">
          <div class="device-status-banner" :class="selectedDevice.status">
            <span class="status-icon">
              <span v-if="selectedDevice.status === 'online'">🟢</span>
              <span v-else-if="selectedDevice.status === 'offline'">🔴</span>
              <span v-else-if="selectedDevice.status === 'conflict'">⚠️</span>
            </span>
            <span class="status-text">
              {{ selectedDevice.status === 'online' ? 'Online' : 
                 selectedDevice.status === 'offline' ? 'Offline' : 
                 selectedDevice.status === 'conflict' ? 'IP CONFLICT' : 'Unknown' }}
            </span>
          </div>

          <div class="details-content">
            <div v-if="selectedDevice.status === 'conflict'" class="conflict-warning">
              <div class="warning-icon">🚨</div>
              <div class="warning-text">
                <strong>IP Address Conflict Detected!</strong>
                <p>Multiple devices are using the same IP address: {{ selectedDevice.ip_address }}</p>
                <p>This can cause network connectivity issues and should be resolved immediately.</p>
              </div>
            </div>

            <div class="detail-section">
              <h3>Basic Information</h3>
              <div class="detail-grid">
                <div class="detail-item">
                  <span class="detail-label">Hostname</span>
                  <span class="detail-value">{{ selectedDevice.hostname || 'Unknown' }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">IP Address</span>
                  <span class="detail-value ip">{{ selectedDevice.ip_address }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">MAC Address</span>
                  <div class="value-with-copy">
                    <span class="detail-value mac">{{ selectedDevice.mac_address }}</span>
                    <button @click="copyToClipboard(selectedDevice.mac_address)" class="copy-btn" title="Copy MAC">📋</button>
                  </div>
                </div>
                <div class="detail-item" v-if="selectedDevice.vendor">
                  <span class="detail-label">Vendor</span>
                  <span class="detail-value">{{ selectedDevice.vendor }}</span>
                </div>
                <div class="detail-item" v-if="selectedDevice.network_name">
                  <span class="detail-label">Network</span>
                  <span class="detail-value vlan-badge">{{ selectedDevice.network_name }}</span>
                </div>
                <div class="detail-item" v-if="selectedDevice.interface_name">
                  <span class="detail-label">Interface</span>
                  <span class="detail-value">{{ selectedDevice.interface_name }}</span>
                </div>
              </div>
            </div>

            <div class="detail-section">
              <h3>Timing Information</h3>
              <div class="detail-grid">
                <div class="detail-item">
                  <span class="detail-label">First Seen</span>
                  <span class="detail-value">{{ formatFullTime(selectedDevice.first_seen) }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">Last Seen</span>
                  <span class="detail-value">{{ formatFullTime(selectedDevice.last_seen) }}</span>
                </div>
                <div class="detail-item" v-if="selectedDevice.last_seen">
                  <span class="detail-label">Last Seen (Relative)</span>
                  <span class="detail-value">{{ formatTime(selectedDevice.last_seen) }}</span>
                </div>
              </div>
            </div>

            <div class="action-buttons">
              <button @click="copyToClipboard(selectedDevice.ip_address)" class="action-btn copy-ip">📋 Copy IP</button>
              <button @click="copyToClipboard(selectedDevice.mac_address)" class="action-btn copy-mac">📋 Copy MAC</button>
              <button @click="pingDevice(selectedDevice)" class="action-btn ping">📡 Ping</button>

              <button @click="scanPort(selectedDevice)" class="action-btn scan-port" title="Coming soon">🔍 Scan Port</button>
              <button @click="showConflictsForIP(selectedDevice)" class="action-btn conflict-info" v-if="selectedDevice.status === 'conflict'">⚠️ View Conflicts</button>
            </div>
            <div class="future-release-note">
              <span class="note-icon">🚀</span>
              <span class="note-text">Scan Port and Ping feature coming in future release</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showConflictsPanel" class="conflicts-panel" :class="{ open: showConflictsPanel }">
      <div class="conflicts-panel-header">
        <h3>⚠️ IP Conflict Details</h3>
        <button @click="showConflictsPanel = false" class="close-panel-btn">×</button>
      </div>
      <div class="conflicts-panel-content">
        <div v-if="conflictsLoading" class="loading-conflicts">
          <div class="loading-spinner-small"></div>
          <span>Loading conflicts...</span>
        </div>
        <div v-else-if="conflictsData.length === 0" class="no-conflicts">
          <span>✅ No active IP conflicts detected</span>
        </div>
        <div v-else>
          <div v-for="conflict in conflictsData" :key="`${conflict.network_id}:${conflict.ip_address}`" class="conflict-item">
            <div class="conflict-header">
              <span class="conflict-ip">🚨 {{ conflict.ip_address }}</span>
              <span class="conflict-vlan">{{ conflict.network_name || `Network ${conflict.network_id}` }}</span>
            </div>
            <div class="conflict-details">
              <div>MAC: {{ conflict.mac_address }}</div>
              <div>Hostname: {{ conflict.hostname || 'Unknown' }}</div>
              <div>Vendor: {{ conflict.vendor || 'Unknown' }}</div>
              <div>Last Seen: {{ formatFullTime(conflict.last_seen) }}</div>
            </div>
            <div class="conflict-actions">
              <button @click="investigateConflict(conflict)" class="investigate-btn">🔍 Investigate</button>
              <button @click="resolveConflict(conflict)" class="resolve-btn">✅ Mark Resolved</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
 const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8082'
class WebSocketManager {
  constructor() {
    this.ws = null
    this.isConnected = false
    this.isConnecting = false
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 10
    this.messageHandlers = []
  }

  static getInstance() {
    if (!this.instance) {
      this.instance = new WebSocketManager()
    }
    return this.instance
  }

  connect(url) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.log('WebSocket already connected')
      return
    }

    if (this.isConnecting) {
      console.log('WebSocket already connecting...')
      return
    }

    this.isConnecting = true
    this.ws = new WebSocket(url)

    this.ws.onopen = () => {
      console.log('✅ WebSocket connected')
      this.isConnected = true
      this.isConnecting = false
      this.reconnectAttempts = 0
    }

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        this.messageHandlers.forEach(handler => handler(data))
      } catch (e) {
        console.error('WebSocket message parse error:', e)
      }
    }

    this.ws.onclose = () => {
      console.log('🔌 WebSocket closed')
      this.isConnected = false
      this.isConnecting = false
      this.reconnect()
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
  }

  reconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('Max reconnection attempts reached')
      return
    }

    const delay = Math.min(3000 * Math.pow(2, this.reconnectAttempts), 30000)
    console.log(`Reconnecting in ${delay/1000}s... (attempt ${this.reconnectAttempts + 1})`)
    
    setTimeout(() => {
      this.reconnectAttempts++
      if (this.ws) {
        this.ws.close()
      }
      this.connect(this.ws?.url)
    }, delay)
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.isConnected = false
    this.isConnecting = false
  }

  onMessage(handler) {
    this.messageHandlers.push(handler)
  }

  offMessage(handler) {
    const index = this.messageHandlers.indexOf(handler)
    if (index > -1) {
      this.messageHandlers.splice(index, 1)
    }
  }
}

export default {
  name: 'LanScanner',

  setup() {
    const devices = ref([])
    const networks = ref([])
    const loading = ref(true)
    const connected = ref(false)
    const connectionError = ref(null)
    const selectedNetworkId = ref('all')
    const selectedStatus = ref('all')
    const searchQuery = ref('')
    const selectedDevice = ref(null)
    
    const scrollContainer = ref(null)
    const itemsPerLoad = ref(50)
    const currentDisplayLimit = ref(50)
    const isLoadingMore = ref(false)
    const hasMore = ref(true)
    const conflictsData = ref([])
    const conflictsLoading = ref(false)
    const showConflictsPanel = ref(false)
    let refreshInterval = null
    let messageHandler = null
    const apiUrl = (path) => {
      const cleanPath = path.startsWith('/') ? path : `/${path}`
      return `${API_BASE_URL}${cleanPath}`
    }

    const getWebSocketUrl = () => {
      const protocol = API_BASE_URL.startsWith('https') ? 'wss' : 'ws'
      const host = API_BASE_URL.replace(/^https?:\/\//, '')
      return `${protocol}://${host}/ws`
    }

    const wsManager = WebSocketManager.getInstance()

    const networkList = computed(() => networks.value)

    const normalizeDevice = (raw, networkInfo = null) => {
      let status = 'offline'
      
      if (raw.device_status) {
        status = raw.device_status
      } else if (raw.status) {
        status = raw.status
      }
      
      if (!['online', 'offline', 'conflict'].includes(status)) {
        status = 'offline'
      }
      
      return {
        id: raw.id,
        network_id: raw.network_id,
        network_name: networkInfo?.vlan_name || null,
        interface_name: networkInfo?.interface_name || null,
        ip_address: raw.ip_address,
        mac_address: raw.mac_address,
        hostname: raw.hostname || '',
        vendor: raw.vendor || '',
        status: status,
        first_seen: raw.first_seen,
        last_seen: raw.last_seen,
        created_at: raw.created_at,
        updated_at: raw.updated_at,
        isNew: false
      }
    }

    const deviceKey = (d) => `${d.network_id}:${d.ip_address}`

    const totalDevicesCount = computed(() => devices.value.length)
    const onlineCount = computed(() => devices.value.filter(d => d.status === 'online').length)
    const offlineCount = computed(() => devices.value.filter(d => d.status === 'offline').length)
    const conflictCount = computed(() => devices.value.filter(d => d.status === 'conflict').length)

    const filteredDevicesList = computed(() => {
      let list = [...devices.value]

      if (selectedNetworkId.value !== 'all') {
        list = list.filter(d => String(d.network_id) === String(selectedNetworkId.value))
      }

      if (selectedStatus.value !== 'all') {
        list = list.filter(d => d.status === selectedStatus.value)
      }

      if (searchQuery.value) {
        const q = searchQuery.value.toLowerCase()
        list = list.filter(d =>
          (d.ip_address && d.ip_address.toLowerCase().includes(q)) ||
          (d.mac_address && d.mac_address.toLowerCase().includes(q)) ||
          (d.hostname && d.hostname.toLowerCase().includes(q)) ||
          (d.vendor && d.vendor.toLowerCase().includes(q))
        )
      }

      list.sort((a, b) => {
        const pa = a.ip_address?.split('.').map(Number) ?? []
        const pb = b.ip_address?.split('.').map(Number) ?? []
        for (let i = 0; i < 4; i++) {
          if ((pa[i] ?? 0) !== (pb[i] ?? 0)) return (pa[i] ?? 0) - (pb[i] ?? 0)
        }
        return 0
      })

      return list
    })

    const filteredDevicesCount = computed(() => filteredDevicesList.value.length)
    const displayedDevices = computed(() => filteredDevicesList.value.slice(0, currentDisplayLimit.value))
    const displayedDevicesCount = computed(() => displayedDevices.value.length)
    const fetchNetworks = async () => {
      try {
        const res = await fetch(apiUrl('/v1/api/vlans'))
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = await res.json()
        if (Array.isArray(data)) {
          networks.value = data
          console.log(`📡 Loaded ${networks.value.length} networks:`, networks.value.map(n => n.vlan_name))
        }
        return true
      } catch (err) {
        console.error('Failed to fetch networks:', err)
        return false
      }
    }
    const fetchDevicesForNetwork = async (network) => {
      try {
        const res = await fetch(apiUrl(`/v1/api/vlans/${network.id}/devices`))
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = await res.json()
        if (Array.isArray(data)) {
          return data.map(device => normalizeDevice(device, network))
        }
        return []
      } catch (err) {
        console.error(`Failed to fetch devices for network ${network.id} (${network.vlan_name}):`, err)
        return []
      }
    }

    const fetchAllDevices = async () => {
      loading.value = true
      try {
        const networksLoaded = await fetchNetworks()
        if (!networksLoaded || networks.value.length === 0) {
          console.log('No networks found')
          devices.value = []
          return
        }
        const devicePromises = networks.value.map(network => fetchDevicesForNetwork(network))
        const devicesArrays = await Promise.all(devicePromises)
        const allDevices = devicesArrays.flat()
        devices.value = allDevices
        resetInfiniteScroll()
        console.log(` Loaded ${devices.value.length} devices from ${networks.value.length} networks`)
      } catch (err) {
        console.error('Failed to fetch all devices:', err)
        connectionError.value = `Failed to fetch devices: ${err.message}`
      } finally {
        loading.value = false
      }
    }

    const fetchConflicts = async () => {
      conflictsLoading.value = true
      try {
        const res = await fetch(apiUrl('/v1/api/conflicts'))
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = await res.json()
        
        if (data && data.conflicts) {
          conflictsData.value = data.conflicts
          conflictsData.value.forEach(conflict => {
            const network = networks.value.find(n => n.id === conflict.network_id)
            if (network) {
              conflict.network_name = network.vlan_name
            }
          })
        }
      } catch (err) {
        console.error('Failed to fetch conflicts:', err)
      } finally {
        conflictsLoading.value = false
      }
    }

    const refreshData = async () => {
      console.log('🔄 Refreshing all data...')
      await fetchAllDevices()
      await fetchConflicts()
    }

    const refreshConflicts = () => fetchConflicts()
    const showConflictsForIP = (device) => { showConflictsPanel.value = true }

    const investigateConflict = (conflict) => {
      const key = `${conflict.network_id}:${conflict.ip_address}`
      const device = devices.value.find(d => deviceKey(d) === key)
      if (device) {
        selectDevice(device)
        showConflictsPanel.value = false
      }
    }

    const resolveConflict = (conflict) => {
      conflictsData.value = conflictsData.value.filter(c => 
        !(c.network_id === conflict.network_id && c.ip_address === conflict.ip_address)
      )
      const key = `${conflict.network_id}:${conflict.ip_address}`
      const device = devices.value.find(d => deviceKey(d) === key)
      if (device && device.status === 'conflict') {
        device.status = 'offline'
      }
    }

    const resetInfiniteScroll = () => {
      currentDisplayLimit.value = itemsPerLoad.value
      hasMore.value = filteredDevicesCount.value > itemsPerLoad.value
      isLoadingMore.value = false
      if (scrollContainer.value) scrollContainer.value.scrollTop = 0
    }

    const handleScroll = async (event) => {
      if (isLoadingMore.value || !hasMore.value) return
      const container = event.target
      const scrollPosition = container.scrollTop + container.clientHeight
      const scrollHeight = container.scrollHeight
      if (scrollHeight - scrollPosition < 200) await loadMore()
    }

    const loadMore = async () => {
      if (isLoadingMore.value || !hasMore.value) return
      isLoadingMore.value = true
      await new Promise(resolve => setTimeout(resolve, 100))
      const newLimit = Math.min(currentDisplayLimit.value + itemsPerLoad.value, filteredDevicesCount.value)
      currentDisplayLimit.value = newLimit
      hasMore.value = currentDisplayLimit.value < filteredDevicesCount.value
      isLoadingMore.value = false
    }

    const upsertDevice = (incoming, isNewDevice = false) => {
      const key = `${incoming.network_id}:${incoming.ip_address}`
      const idx = devices.value.findIndex(d => deviceKey(d) === key)
      const normalizedDevice = normalizeDevice(incoming, {
        vlan_name: incoming.network_name,
        interface_name: incoming.interface_name
      })
      normalizedDevice.isNew = isNewDevice

      if (idx === -1) {
        devices.value.unshift(normalizedDevice)
        if (isNewDevice) {
          setTimeout(() => {
            const i = devices.value.findIndex(d => deviceKey(d) === key)
            if (i !== -1) devices.value[i].isNew = false
          }, 3000)
        }
      } else {
        const existing = devices.value[idx]
        let finalStatus = normalizedDevice.status
        if (existing.status === 'conflict' && normalizedDevice.status !== 'conflict') {
          finalStatus = 'conflict'
        }
        devices.value[idx] = {
          ...existing,
          ...normalizedDevice,
          status: finalStatus,
          first_seen: existing.first_seen || normalizedDevice.first_seen,
          isNew: false,
        }
        if (selectedDevice.value && deviceKey(selectedDevice.value) === key) {
          selectedDevice.value = devices.value[idx]
        }
      }
      if (filteredDevicesCount.value !== currentDisplayLimit.value) {
        hasMore.value = currentDisplayLimit.value < filteredDevicesCount.value
      }
    }

    const handleWebSocketMessage = (data) => {
      console.log('📨 WebSocket message:', data.event_type || 'device_update')
      let networkName = null
      let interfaceName = null
      if (data.network_id) {
        const network = networks.value.find(n => n.id === data.network_id)
        if (network) {
          networkName = network.vlan_name
          interfaceName = network.interface_name
        }
      }
      
      switch (data.event_type) {
        case 'new_device':
          upsertDevice({
            network_id: data.network_id,
            network_name: networkName || data.network_name,
            interface_name: interfaceName,
            ip_address: data.ip_address,
            mac_address: data.mac_address,
            hostname: data.hostname,
            vendor: data.vendor,
            status: data.status || 'online',
            first_seen: data.first_seen || new Date().toISOString(),
            last_seen: data.last_seen || new Date().toISOString(),
          }, true)
          break
        case 'status_change':
          upsertDevice({
            network_id: data.network_id,
            network_name: networkName || data.network_name,
            interface_name: interfaceName,
            ip_address: data.ip_address,
            mac_address: data.mac_address,
            hostname: data.hostname,
            vendor: data.vendor,
            status: data.new_status,
            last_seen: data.last_seen || new Date().toISOString(),
          }, false)
          break
        case 'went_offline':
          upsertDevice({
            network_id: data.network_id,
            network_name: networkName || data.network_name,
            interface_name: interfaceName,
            ip_address: data.ip_address,
            mac_address: data.mac_address,
            hostname: data.hostname,
            vendor: data.vendor,
            status: 'offline',
            last_seen: data.last_seen || new Date().toISOString(),
          }, false)
          break
        case 'came_online':
          upsertDevice({
            network_id: data.network_id,
            network_name: networkName || data.network_name,
            interface_name: interfaceName,
            ip_address: data.ip_address,
            mac_address: data.mac_address,
            hostname: data.hostname,
            vendor: data.vendor,
            status: 'online',
            last_seen: data.last_seen || new Date().toISOString(),
          }, false)
          break
        case 'ip_conflict':
          console.log('🚨 IP CONFLICT detected:', data.ip_address)
          if (conflictCount.value === 0) showConflictsPanel.value = true
          fetchConflicts()
          upsertDevice({
            network_id: data.network_id,
            network_name: networkName || data.network_name,
            interface_name: interfaceName,
            ip_address: data.ip_address,
            mac_address: data.mac_address || 'ff:ff:ff:ff:ff:ff',
            hostname: data.hostname,
            vendor: data.vendor,
            status: 'conflict',
            last_seen: new Date().toISOString(),
          }, false)
          break
        default:
          if (data.ip_address && data.mac_address) {
            upsertDevice({
              network_id: data.network_id,
              network_name: networkName || data.network_name,
              interface_name: interfaceName,
              ip_address: data.ip_address,
              mac_address: data.mac_address,
              hostname: data.hostname,
              vendor: data.vendor,
              status: data.status || data.device_status || 'offline',
              first_seen: data.first_seen,
              last_seen: data.last_seen,
            }, false)
          }
      }
    }

    const formatTime = (ts) => {
      if (!ts) return 'Never'
      try {
        const date = new Date(ts)
        if (isNaN(date.getTime())) return 'Invalid date'
        
        const now = new Date()
        const diffMs = now - date
        const diffSeconds = Math.floor(diffMs / 1000)
        const diffMinutes = Math.floor(diffSeconds / 60)
        const diffHours = Math.floor(diffMinutes / 60)
        const diffDays = Math.floor(diffHours / 24)
        const diffWeeks = Math.floor(diffDays / 7)
        const diffMonths = Math.floor(diffDays / 30)
        const diffYears = Math.floor(diffDays / 365)
        
        if (diffSeconds < 60) return 'Just now'
        if (diffMinutes < 60) return `${diffMinutes}m ago`
        if (diffHours < 24) return `${diffHours}h ago`
        if (diffDays < 7) return `${diffDays}d ago`
        if (diffDays < 30) return `${diffWeeks}w ago`
        if (diffDays < 365) return `${diffMonths}mo ago`
        return `${diffYears}y ago`
      } catch (e) {
        return 'Invalid date'
      }
    }

    const formatFullTime = (ts) => {
      if (!ts) return 'Never'
      try {
        const date = new Date(ts)
        if (isNaN(date.getTime())) return 'Invalid date'
        return date.toLocaleString('en-US', {
          year: 'numeric',
          month: 'short',
          day: 'numeric',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
          hour12: true
        })
      } catch (e) {
        return 'Invalid date'
      }
    }

    const copyToClipboard = (text) => {
      if (text) navigator.clipboard.writeText(text).catch(console.error)
    }

    const selectDevice = (d) => { selectedDevice.value = d }
    const clearSelection = () => { selectedDevice.value = null }
    const pingDevice = (device) => alert(`Ping ${device.ip_address} will be available in a future release.\n\nThis feature will allow you to scan open ports on this device.`)
    const scanPort = (device) => {
      alert(`🔍 Port scanning for ${device.ip_address} will be available in a future release.\n\nThis feature will allow you to scan open ports on this device.`)
    }
    
    const reconnect = () => {
      wsManager.disconnect()
      setTimeout(() => {
        wsManager.connect(getWebSocketUrl())
        refreshData()
      }, 100)
    }

    const startPeriodicRefresh = () => {
      if (refreshInterval) clearInterval(refreshInterval)
      refreshInterval = setInterval(() => {
        if (connected.value) {
          console.log('🔄 Periodic refresh checking for updates...')
          refreshData()
        }
      }, 30000)
    }

    const stopPeriodicRefresh = () => {
      if (refreshInterval) {
        clearInterval(refreshInterval)
        refreshInterval = null
      }
    }
    const connectionInterval = setInterval(() => {
      connected.value = wsManager.isConnected
    }, 1000)

    watch([searchQuery, selectedNetworkId, selectedStatus], () => resetInfiniteScroll())

    onMounted(async () => {
      console.log('🚀 Component mounted, initializing...')
      messageHandler = handleWebSocketMessage
      wsManager.onMessage(messageHandler)
      wsManager.connect(getWebSocketUrl())
      await refreshData()
      
      startPeriodicRefresh()
    })

    onBeforeUnmount(() => {
      console.log('🧹 Cleaning up...')
      stopPeriodicRefresh()
      clearInterval(connectionInterval)
      if (messageHandler) {
        wsManager.offMessage(messageHandler)
      }
    })

    return {
      devices,
      networks,
      networkList,
      loading,
      connected,
      connectionError,
      selectedNetworkId,
      selectedStatus,
      searchQuery,
      selectedDevice,
      totalDevicesCount,
      onlineCount,
      offlineCount,
      conflictCount,
      filteredDevicesCount,
      displayedDevices,
      displayedDevicesCount,
      scrollContainer,
      hasMore,
      isLoadingMore,
      conflictsData,
      conflictsLoading,
      showConflictsPanel,
      deviceKey,
      formatTime,
      formatFullTime,
      copyToClipboard,
      selectDevice,
      clearSelection,
      pingDevice,
      scanPort,
      reconnect,
      handleScroll,
      refreshData,
      refreshConflicts,
      showConflictsForIP,
      investigateConflict,
      resolveConflict,
    }
  }
}
</script>



<style scoped>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.lan-scanner {
  display: flex;
  flex-direction: column;
  height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  color: #e2e8f0;
  overflow: hidden;
}

.top-bar {
  padding: 20px 20px 10px 20px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
}

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
  cursor: pointer;
  transition: all 0.2s;
}

.vlan-select:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59,130,246,0.2);
}

.status-filter {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-filter label {
  font-size: 14px;
  color: #cbd5e1;
  font-weight: 500;
}

.status-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.status-btn {
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #334155;
  background: #0f172a;
  color: #94a3b8;
}

.status-btn:hover {
  background: #1e293b;
  transform: translateY(-1px);
}

.status-btn.active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: white;
}

.status-btn.online-btn.active {
  background: #10b981;
  border-color: #10b981;
}

.status-btn.offline-btn.active {
  background: #ef4444;
  border-color: #ef4444;
}

.status-btn.conflict-btn {
  background: #f59e0b;
  color: #1e293b;
}

.status-btn.conflict-btn.active {
  background: #ef4444;
  color: white;
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
  box-shadow: 0 0 0 3px rgba(59,130,246,0.2);
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
  background: rgba(16,185,129,0.1);
  border-color: rgba(16,185,129,0.3);
}

.connection-status.disconnected {
  color: #f87171;
  background: rgba(239,68,68,0.1);
  border-color: rgba(239,68,68,0.3);
}

.stats-bar {
  display: flex;
  gap: 25px;
  padding: 12px 20px;
  background: rgba(30,41,59,0.8);
  backdrop-filter: blur(10px);
  border-radius: 12px;
  border: 1px solid rgba(148,163,184,0.1);
  flex-wrap: wrap;
  margin-bottom: 15px;
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

.stat-value.conflict {
  color: #f59e0b;
}

.conflict-label {
  color: #f59e0b;
}

.conflict-alert {
  padding: 15px 20px;
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 15px;
  animation: shake 0.5s ease-in-out;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-5px); }
  75% { transform: translateX(5px); }
}

.alert-icon {
  font-size: 28px;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

.alert-content {
  flex: 1;
  color: white;
}

.alert-content strong {
  display: block;
  font-size: 16px;
  margin-bottom: 4px;
}

.alert-btn {
  padding: 8px 20px;
  background: rgba(255,255,255,0.2);
  border: 1px solid rgba(255,255,255,0.3);
  border-radius: 8px;
  color: white;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.alert-btn:hover {
  background: rgba(255,255,255,0.3);
  transform: translateY(-2px);
}

.main-content {
  display: flex;
  flex: 1;
  overflow: hidden;
  padding: 0 20px 20px 20px;
  gap: 20px;
  min-height: 0;
}

.device-list-panel {
  flex: 1;
  background: rgba(30,41,59,0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148,163,184,0.1);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 8px 16px rgba(0,0,0,0.3);
  min-width: 600px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(148,163,184,0.1);
  background: rgba(15,23,42,0.6);
}

.panel-header h2 {
  margin: 0;
  font-size: 1.4rem;
  font-weight: 600;
  color: #f8fafc;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.refresh-conflicts-btn {
  background: #334155;
  border: none;
  color: #94a3b8;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.refresh-conflicts-btn:hover {
  background: #475569;
  color: #f8fafc;
  transform: rotate(90deg);
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

.error-message {
  margin: 20px;
  background: rgba(239,68,68,0.15);
  border: 1px solid rgba(239,68,68,0.4);
  border-radius: 12px;
  padding: 18px 22px;
  display: flex;
  align-items: center;
  gap: 15px;
  color: #fca5a5;
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
}

.loading {
  text-align: center;
  padding: 80px 30px;
  color: #94a3b8;
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

.devices-list {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.device-card {
  display: flex;
  background: rgba(15,23,42,0.8);
  border: 1px solid rgba(148,163,184,0.1);
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  overflow: hidden;
  min-height: 130px;
}

.device-card:hover {
  transform: translateX(4px);
  border-color: rgba(59,130,246,0.4);
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
}

.device-card.selected {
  border: 2px solid #3b82f6;
  background: rgba(59,130,246,0.15);
}

.device-card.online {
  border-left: 6px solid #10b981;
}

.device-card.offline {
  border-left: 6px solid #ef4444;
  opacity: 0.85;
}

.device-card.conflict {
  border-left: 6px solid #f59e0b;
  background: rgba(245,158,11,0.1);
}

.device-card.new {
  animation: highlight-new 2s ease-out;
}

@keyframes highlight-new {
  0% { background: rgba(16,185,129,0.3); }
  100% { background: rgba(15,23,42,0.8); }
}

.status-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 60px;
  font-size: 28px;
  background: rgba(0,0,0,0.3);
  border-right: 1px solid rgba(148,163,184,0.1);
}

.status-indicator.online {
  background: rgba(16,185,129,0.2);
}

.status-indicator.offline {
  background: rgba(239,68,68,0.2);
}

.status-indicator.conflict {
  background: rgba(245,158,11,0.25);
  animation: pulse 1s infinite;
}

.device-content {
  flex: 1;
  padding: 12px 15px;
}

.device-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 10px;
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

.device-vlan {
  font-size: 11px;
  background: linear-gradient(135deg,#3b82f6,#8b5cf6);
  color: white;
  padding: 3px 8px;
  border-radius: 12px;
  font-weight: 600;
}

.status-badge {
  display: inline-block;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 12px;
  letter-spacing: 0.5px;
  white-space: nowrap;
}

.status-badge.online {
  background: rgba(16,185,129,0.2);
  color: #10b981;
  border: 1px solid rgba(16,185,129,0.3);
}

.status-badge.offline {
  background: rgba(239,68,68,0.2);
  color: #ef4444;
  border: 1px solid rgba(239,68,68,0.3);
}

.status-badge.conflict {
  background: rgba(245,158,11,0.2);
  color: #f59e0b;
  border: 1px solid rgba(245,158,11,0.3);
  animation: pulse 1s infinite;
}

.conflict-badge {
  font-size: 10px;
  background: #ef4444;
  color: white;
  padding: 2px 6px;
  border-radius: 12px;
  font-weight: 700;
  animation: pulse 1s infinite;
}

.device-details {
  display: grid;
  grid-template-columns: repeat(2,1fr);
  gap: 6px 8px;
  margin-top: 4px;
}

.detail-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  background: rgba(0,0,0,0.2);
  padding: 6px 8px;
  border-radius: 6px;
  border: 1px solid rgba(148,163,184,0.05);
}

.status-row {
  background: rgba(0, 0, 0, 0.3);
}

.status-row .detail-value {
  font-weight: 600;
}

.status-row .detail-value.online {
  color: #10b981;
}

.status-row .detail-value.offline {
  color: #ef4444;
}

.status-row .detail-value.conflict {
  color: #f59e0b;
  animation: pulse 1s infinite;
}

.detail-label {
  color: #94a3b8;
  min-width: 65px;
  font-size: 11px;
  font-weight: 500;
  white-space: nowrap;
}

.detail-value {
  color: #e2e8f0;
  font-weight: 500;
  font-size: 12px;
  flex: 1;
  word-break: break-all;
}

.detail-value.ip {
  color: #60a5fa;
  font-family: monospace;
  font-weight: 600;
}

.detail-value.mac {
  font-family: monospace;
  color: #94a3b8;
}

.detail-value.vendor {
  color: #c4b5fd;
  font-weight: 600;
}

.detail-value.last-seen {
  color: #94a3b8;
}

.new-badge {
  position: absolute;
  top: 12px;
  right: 12px;
  background: #10b981;
  color: white;
  font-size: 11px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 20px;
  z-index: 1;
}

.conflict-icon {
  position: absolute;
  top: 12px;
  right: 12px;
  font-size: 20px;
  animation: pulse 1s infinite;
  z-index: 1;
}

.loading-more {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 20px;
  color: #94a3b8;
}

.loading-spinner-small {
  width: 20px;
  height: 20px;
  border: 2px solid #1e293b;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.end-message {
  text-align: center;
  padding: 20px;
  color: #64748b;
  font-size: 13px;
  border-top: 1px solid rgba(148,163,184,0.1);
  margin-top: 10px;
}

.no-devices {
  text-align: center;
  padding: 80px 30px;
  background: rgba(15,23,42,0.6);
  border-radius: 14px;
  margin: 20px;
}

.no-devices-icon {
  font-size: 64px;
  margin-bottom: 25px;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%,100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.device-details-panel {
  width: 450px;
  min-width: 450px;
  background: rgba(30,41,59,0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148,163,184,0.1);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 8px 16px rgba(0,0,0,0.3);
  transition: all 0.3s ease;
}

.device-details-panel.has-device {
  border-color: #3b82f6;
  box-shadow: 0 0 30px rgba(59,130,246,0.3);
}

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
  animation: bounce 2s infinite;
}

@keyframes bounce {
  0%,100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

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
  background: rgba(16,185,129,0.2);
  color: #34d399;
  border-bottom: 1px solid rgba(16,185,129,0.3);
}

.device-status-banner.offline {
  background: rgba(239,68,68,0.2);
  color: #f87171;
  border-bottom: 1px solid rgba(239,68,68,0.3);
}

.device-status-banner.conflict {
  background: rgba(245,158,11,0.2);
  color: #f59e0b;
  border-bottom: 1px solid rgba(245,158,11,0.3);
}

.details-content {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.conflict-warning {
  background: rgba(239,68,68,0.15);
  border: 1px solid rgba(239,68,68,0.4);
  border-radius: 12px;
  padding: 15px;
  display: flex;
  gap: 12px;
}

.warning-icon {
  font-size: 24px;
}

.warning-text strong {
  color: #ef4444;
  display: block;
  margin-bottom: 8px;
}

.warning-text p {
  color: #94a3b8;
  font-size: 13px;
  margin: 4px 0;
}

.detail-section {
  background: rgba(15,23,42,0.8);
  border-radius: 12px;
  padding: 14px;
  border: 1px solid rgba(148,163,184,0.1);
}

.detail-section h3 {
  margin: 0 0 12px 0;
  font-size: 13px;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 600;
  border-bottom: 1px solid rgba(148,163,184,0.2);
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
  font-weight: 500;
}

.detail-item .detail-value {
  font-size: 14px;
  color: #f8fafc;
  font-weight: 500;
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
  transition: all 0.2s;
}

.copy-btn:hover {
  background: #1e293b;
  color: #60a5fa;
}

.vlan-badge {
  display: inline-block;
  background: linear-gradient(135deg,#3b82f6,#8b5cf6);
  color: white;
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 13px;
  font-weight: 600;
}

.clear-btn {
  background: transparent;
  border: 1px solid #334155;
  color: #94a3b8;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  font-size: 18px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.clear-btn:hover {
  background: #1e293b;
  color: #f8fafc;
}

.action-buttons {
  display: grid;
  grid-template-columns: repeat(2,1fr);
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
  transform: translateY(-1px);
}

.action-btn.conflict-info {
  background: #f59e0b;
  color: #1e293b;
}

.action-btn.conflict-info:hover {
  background: #ef4444;
  color: white;
}

.conflicts-panel {
  position: fixed;
  right: -500px;
  top: 0;
  width: 500px;
  height: 100vh;
  background: rgba(30,41,59,0.98);
  backdrop-filter: blur(20px);
  border-left: 1px solid rgba(148,163,184,0.2);
  box-shadow: -5px 0 25px rgba(0,0,0,0.3);
  transition: right 0.3s ease;
  z-index: 1000;
  display: flex;
  flex-direction: column;
}

.conflicts-panel.open {
  right: 0;
}

.conflicts-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(148,163,184,0.1);
  background: rgba(15,23,42,0.8);
}

.conflicts-panel-header h3 {
  margin: 0;
  color: #f59e0b;
}

.close-panel-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 28px;
  cursor: pointer;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-panel-btn:hover {
  background: rgba(148,163,184,0.1);
  color: #f8fafc;
}

.conflicts-panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.loading-conflicts {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px;
  color: #94a3b8;
}

.no-conflicts {
  text-align: center;
  padding: 40px;
  color: #10b981;
}

.conflict-item {
  background: rgba(15,23,42,0.8);
  border: 1px solid rgba(245,158,11,0.3);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
  transition: all 0.2s;
}

.conflict-item:hover {
  transform: translateX(-4px);
  border-color: #f59e0b;
  background: rgba(245,158,11,0.1);
}

.conflict-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(148,163,184,0.1);
}

.conflict-ip {
  font-size: 18px;
  font-weight: 700;
  color: #f59e0b;
  font-family: monospace;
}

.conflict-vlan {
  font-size: 12px;
  background: #3b82f6;
  color: white;
  padding: 4px 10px;
  border-radius: 20px;
}

.conflict-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-bottom: 12px;
  font-size: 13px;
  color: #94a3b8;
}

.conflict-actions {
  display: flex;
  gap: 10px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(148,163,184,0.1);
}

.investigate-btn, .resolve-btn {
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
  flex: 1;
}

.investigate-btn {
  background: #3b82f6;
  color: white;
}

.investigate-btn:hover {
  background: #2563eb;
  transform: translateY(-1px);
}

.resolve-btn {
  background: #10b981;
  color: white;
}

.resolve-btn:hover {
  background: #059669;
  transform: translateY(-1px);
}

.devices-list::-webkit-scrollbar,
.selected-device-details::-webkit-scrollbar,
.conflicts-panel-content::-webkit-scrollbar {
  width: 10px;
}

.devices-list::-webkit-scrollbar-track,
.selected-device-details::-webkit-scrollbar-track,
.conflicts-panel-content::-webkit-scrollbar-track {
  background: #0f172a;
  border-radius: 5px;
}

.devices-list::-webkit-scrollbar-thumb,
.selected-device-details::-webkit-scrollbar-thumb,
.conflicts-panel-content::-webkit-scrollbar-thumb {
  background: #334155;
  border-radius: 5px;
}

.devices-list::-webkit-scrollbar-thumb:hover,
.selected-device-details::-webkit-scrollbar-thumb:hover,
.conflicts-panel-content::-webkit-scrollbar-thumb:hover {
  background: #475569;
}

@media (max-width: 1200px) {
  .device-list-panel {
    min-width: 450px;
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
  .device-list-panel, .device-details-panel {
    min-width: 100%;
    width: 100%;
    height: 500px;
  }
  .conflicts-panel {
    width: 100%;
    right: -100%;
  }
  .conflict-details {
    grid-template-columns: 1fr;
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
  .status-buttons {
    flex-wrap: wrap;
  }
  .device-details {
    grid-template-columns: 1fr;
  }
  .action-buttons {
    grid-template-columns: 1fr;
  }
}
</style>