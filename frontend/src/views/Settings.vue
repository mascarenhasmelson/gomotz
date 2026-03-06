<template>
  <div class="network-dashboard">
    <!-- Header -->
    <div class="dashboard-header">
      <h1>Network Configuration</h1>
      <p class="subtitle">Configure network interfaces and VLANs with static IP addressing</p>
    </div>

    <!-- Main Content -->
    <div class="dashboard-content">
      <!-- Left Panel - Network Interfaces -->
      <div class="interfaces-list-panel">
        <div class="panel-header">
          <h2>Network Interfaces</h2>
          <div class="header-actions">
            <button class="refresh-btn" @click="fetchInterfaces" :disabled="isLoading">
              <span class="refresh-icon" :class="{ 'spinning': isLoading }">↻</span>
              {{ isLoading ? 'Loading...' : 'Refresh Interfaces' }}
            </button>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="isLoading" class="loading-state">
          <div class="loading-spinner"></div>
          <p>Loading network interfaces...</p>
        </div>

        <!-- Interfaces List -->
        <div class="interfaces-list" v-if="!isLoading">
          <h4>Available Interfaces</h4>
          <div 
            v-for="iface in interfaces" 
            :key="iface.name"
            class="interface-item"
            :class="{ 
              'selected': selectedInterface && selectedInterface.name === iface.name,
              'default-route': iface.isDefault 
            }"
            @click="selectInterface(iface)"
          >
            <div class="interface-indicator" :class="iface.status"></div>
            <div class="interface-info">
              <div class="interface-name">
                <span class="name">{{ iface.name }}</span>
                <span v-if="iface.isDefault" class="default-badge">Default Route</span>
                <span class="status-badge" :class="iface.status">{{ iface.status }}</span>
              </div>
              <div class="interface-details">
                <span class="ip">{{ iface.ipAddress || 'No IP' }}</span>
                <span class="mac">{{ iface.mac }}</span>
              </div>
            </div>
            <div class="interface-actions">
              <label class="monitor-checkbox" @click.stop>
                <input 
                  type="checkbox" 
                  v-model="iface.monitorEnabled"
                  @change="toggleInterfaceMonitor(iface)"
                >
                <span class="checkmark"></span>
                <span class="monitor-label">Monitor</span>
              </label>
              <button class="action-btn" @click.stop="saveInterfaceConfig(iface)" title="Save Configuration">
                💾
              </button>
            </div>
          </div>

          <!-- Empty State -->
          <div v-if="interfaces.length === 0 && !isLoading" class="empty-state">
            <div class="empty-icon">🔌</div>
            <h3>No Interfaces Found</h3>
            <p>Unable to detect network interfaces</p>
          </div>
        </div>
      </div>

      <!-- Right Panel - VLAN Configuration -->
      <div class="config-panel">
        <div class="panel-header">
          <h2>{{ selectedInterface ? `VLAN on ${selectedInterface.name}` : 'VLAN Configuration' }}</h2>
          <button v-if="selectedInterface" class="close-btn" @click="selectedInterface = null">×</button>
        </div>

        <div class="config-form">
          <!-- Interface Info (when selected) -->
          <div v-if="selectedInterface" class="info-section">
            <div class="info-item">
              <span class="info-label">Interface</span>
              <span class="info-value">{{ selectedInterface.name }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Status</span>
              <span class="info-value status" :class="selectedInterface.status">{{ selectedInterface.status }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">IP Address</span>
              <span class="info-value">{{ selectedInterface.ipAddress || 'No IP' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Monitor</span>
              <span class="info-value">
                <span class="monitor-status" :class="{ 'enabled': selectedInterface.monitorEnabled }">
                  {{ selectedInterface.monitorEnabled ? 'Enabled' : 'Disabled' }}
                </span>
              </span>
            </div>
          </div>

          <!-- VLAN Creation Form -->
          <div class="form-section">
            <h3>Create VLAN Interface</h3>
            
            <div class="form-group">
              <label for="vlanId">VLAN ID</label>
              <input 
                type="number" 
                id="vlanId"
                v-model="vlanConfig.id"
                min="1"
                max="4094"
                placeholder="e.g., 10, 20, 100"
                class="form-input"
                :disabled="!selectedInterface"
              />
              <span class="input-hint">VLAN ID (1-4094)</span>
            </div>

            <div class="form-group">
              <label for="ipAddress">IP Address</label>
              <input 
                type="text" 
                id="ipAddress"
                v-model="vlanConfig.ipAddress"
                placeholder="e.g., 192.168.10.1"
                class="form-input"
                :disabled="!selectedInterface"
                @input="validateIP"
              />
              <span v-if="ipError" class="error-text">{{ ipError }}</span>
            </div>

            <!-- Simplified CIDR Selection -->
            <div class="form-group">
              <label for="cidr">CIDR Notation</label>
              <select 
                id="cidr"
                v-model="vlanConfig.cidr"
                class="cidr-select"
                :disabled="!selectedInterface"
              >
                <option value="/24">/24</option>
                <option value="/16">/16</option>
                <option value="/8">/8</option>
              </select>
              <span class="input-hint">Select CIDR notation</span>
            </div>

            <div class="form-group">
              <label for="gateway">Default Gateway</label>
              <input 
                type="text" 
                id="gateway"
                v-model="vlanConfig.gateway"
                placeholder="e.g., 192.168.10.1"
                class="form-input"
                :disabled="!selectedInterface"
                @input="validateGateway"
              />
              <span v-if="gatewayError" class="error-text">{{ gatewayError }}</span>
            </div>

            <!-- Monitor Checkbox for VLAN -->
            <div class="form-group monitor-group">
              <label class="monitor-checkbox">
                <input 
                  type="checkbox" 
                  v-model="vlanConfig.monitorEnabled"
                  :disabled="!selectedInterface"
                  @change="toggleVLANMonitor"
                >
                <span class="checkmark"></span>
                <span class="monitor-label">Enable monitoring for this VLAN</span>
              </label>
            </div>

            <!-- Network Preview -->
            <div v-if="vlanConfig.ipAddress && vlanConfig.cidr" class="network-preview">
              <h4>Network Preview</h4>
              <div class="preview-grid">
                <div class="preview-item">
                  <span class="preview-label">Network:</span>
                  <span class="preview-value">{{ calculateNetworkAddress(vlanConfig) }}</span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">Broadcast:</span>
                  <span class="preview-value">{{ calculateBroadcastAddress(vlanConfig) }}</span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">Hosts:</span>
                  <span class="preview-value">{{ calculateUsableHosts(vlanConfig) }}</span>
                </div>
              </div>
            </div>

            <!-- Action Buttons -->
            <div class="form-actions">
              <button class="btn btn-primary" @click="createVLAN" :disabled="!selectedInterface">
                <span class="btn-icon">➕</span>
                Create VLAN
              </button>
              <button class="btn btn-danger" @click="resetVLANForm">
                Reset
              </button>
            </div>
          </div>

          <!-- Existing VLANs on this Interface -->
          <div v-if="selectedInterface && existingVLANs.filter(v => v.parentInterface === selectedInterface.name).length > 0" class="form-section">
            <h3>Existing VLANs on {{ selectedInterface.name }}</h3>
            <div class="vlan-list-compact">
              <div v-for="vlan in existingVLANs.filter(v => v.parentInterface === selectedInterface.name)" :key="vlan.id" class="vlan-compact-item">
                <div class="vlan-compact-info">
                  <span class="vlan-compact-id">VLAN {{ vlan.id }}</span>
                  <span class="vlan-compact-ip">{{ vlan.ipAddress }}{{ vlan.cidr }}</span>
                  <span v-if="vlan.monitorEnabled" class="monitoring-badge" title="Monitoring enabled">📊</span>
                </div>
                <div class="vlan-compact-actions">
                  <label class="icon-checkbox" @click.stop>
                    <input 
                      type="checkbox" 
                      v-model="vlan.monitorEnabled"
                      @change="toggleVLANMonitor(vlan)"
                    >
                    <span class="checkmark-small"></span>
                  </label>
                  <button class="icon-btn" @click.stop="editVLAN(vlan)" title="Edit">✏️</button>
                  <button class="icon-btn delete" @click.stop="deleteVLAN(vlan)" title="Delete">🗑️</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Success/Error Notifications -->
    <div v-if="notification.show" class="notification" :class="notification.type">
      <span class="notification-icon">{{ notification.type === 'success' ? '✅' : '❌' }}</span>
      <span class="notification-message">{{ notification.message }}</span>
      <button class="notification-close" @click="notification.show = false">×</button>
    </div>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted } from 'vue'

// API Base URL - configure based on your environment
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

export default {
  name: 'NetworkDashboard',
  
  setup() {
    // State
    const interfaces = ref([])
    const selectedInterface = ref(null)
    const isLoading = ref(false)
    
    // VLAN Configuration - Simplified
    const vlanConfig = reactive({
      id: '',
      ipAddress: '',
      cidr: '/24',
      gateway: '',
      monitorEnabled: false
    })

    // Existing VLANs
    const existingVLANs = ref([])

    // Validation errors
    const ipError = ref('')
    const gatewayError = ref('')

    // Notification
    const notification = reactive({
      show: false,
      type: 'success',
      message: ''
    })

    // Computed properties
    const interfacesWithMonitor = computed(() => {
      return interfaces.value.map(iface => ({
        ...iface,
        monitorEnabled: iface.monitorEnabled || false
      }))
    })

    // API Methods
    const fetchInterfaces = async () => {
      isLoading.value = true
      
      try {
        const response = await fetch(`${API_BASE_URL}/network/interfaces`)
        
        if (!response.ok) {
          throw new Error('Failed to fetch network interfaces')
        }
        
        const data = await response.json()
        
        interfaces.value = (data.interfaces || []).map(iface => ({
          ...iface,
          monitorEnabled: iface.monitorEnabled || false
        }))
        
        // Fetch existing VLANs
        await fetchVLANs()
        
        showNotification('Network interfaces loaded successfully', 'success')
      } catch (error) {
        console.error('Error fetching interfaces:', error)
        showNotification('Failed to load network interfaces', 'error')
        
        // For development/demo - use sample data
        loadSampleData()
      } finally {
        isLoading.value = false
      }
    }

    const fetchVLANs = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/vlans`)
        
        if (!response.ok) {
          throw new Error('Failed to fetch VLANs')
        }
        
        const data = await response.json()
        existingVLANs.value = (data.vlans || []).map(vlan => ({
          ...vlan,
          monitorEnabled: vlan.monitorEnabled || false
        }))
      } catch (error) {
        console.error('Error fetching VLANs:', error)
        // For development - use sample data
        existingVLANs.value = [
          { id: 10, parentInterface: 'eth0', ipAddress: '192.168.10.1', cidr: '/24', gateway: '192.168.10.254', monitorEnabled: true },
          { id: 20, parentInterface: 'eth0', ipAddress: '192.168.20.1', cidr: '/24', gateway: '192.168.20.254', monitorEnabled: false }
        ]
      }
    }

    // Development sample data
    const loadSampleData = () => {
      interfaces.value = [
        { 
          name: 'eth0', 
          ipAddress: '192.168.1.100', 
          cidr: '/24',
          gateway: '192.168.1.1', 
          mac: '00:11:22:33:44:55', 
          status: 'up',
          isDefault: true,
          monitorEnabled: true
        },
        { 
          name: 'wlan0', 
          ipAddress: '10.0.0.5', 
          cidr: '/24',
          gateway: '10.0.0.1', 
          mac: 'aa:bb:cc:dd:ee:ff', 
          status: 'up',
          isDefault: false,
          monitorEnabled: false
        },
        { 
          name: 'docker0', 
          ipAddress: '172.17.0.1', 
          cidr: '/16',
          gateway: null, 
          mac: '00:11:22:33:44:66', 
          status: 'up',
          isDefault: false,
          monitorEnabled: false
        }
      ]
      
      existingVLANs.value = [
        { id: 10, parentInterface: 'eth0', ipAddress: '192.168.10.1', cidr: '/24', gateway: '192.168.10.254', monitorEnabled: true },
        { id: 20, parentInterface: 'eth0', ipAddress: '192.168.20.1', cidr: '/24', gateway: '192.168.20.254', monitorEnabled: false }
      ]
    }

    const selectInterface = (iface) => {
      selectedInterface.value = iface
    }

    const toggleInterfaceMonitor = async (iface) => {
      try {
        const response = await fetch(`${API_BASE_URL}/network/interfaces/${iface.name}/monitor`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ enabled: iface.monitorEnabled })
        })
        
        if (!response.ok) {
          throw new Error('Failed to update monitoring state')
        }
        
        showNotification(`Monitoring ${iface.monitorEnabled ? 'enabled' : 'disabled'} for ${iface.name}`, 'success')
      } catch (error) {
        console.error('Error updating monitor state:', error)
        // Revert checkbox state on error
        iface.monitorEnabled = !iface.monitorEnabled
        showNotification(`Failed to update monitoring state`, 'error')
      }
    }

    const toggleVLANMonitor = async (vlan = null) => {
      const targetVLAN = vlan || { id: vlanConfig.id, monitorEnabled: vlanConfig.monitorEnabled }
      
      if (!targetVLAN.id) {
        showNotification('No VLAN selected', 'error')
        return
      }

      try {
        const response = await fetch(`${API_BASE_URL}/vlans/${targetVLAN.id}/monitor`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ enabled: targetVLAN.monitorEnabled })
        })
        
        if (!response.ok) {
          throw new Error('Failed to update monitoring state')
        }
        
        showNotification(`Monitoring ${targetVLAN.monitorEnabled ? 'enabled' : 'disabled'} for VLAN ${targetVLAN.id}`, 'success')
      } catch (error) {
        console.error('Error updating VLAN monitor state:', error)
        // Revert checkbox state on error
        targetVLAN.monitorEnabled = !targetVLAN.monitorEnabled
        showNotification(`Failed to update monitoring state`, 'error')
      }
    }

    const saveInterfaceConfig = async (iface) => {
      try {
        const response = await fetch(`${API_BASE_URL}/network/interfaces/${iface.name}`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(iface)
        })
        
        if (!response.ok) {
          throw new Error('Failed to save configuration')
        }
        
        showNotification(`Configuration saved for ${iface.name}`, 'success')
      } catch (error) {
        console.error('Error saving interface config:', error)
        showNotification(`Failed to save configuration for ${iface.name}`, 'error')
      }
    }

    const validateIP = () => {
      const ipRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
      if (vlanConfig.ipAddress && !ipRegex.test(vlanConfig.ipAddress)) {
        ipError.value = 'Invalid IP address format'
      } else {
        ipError.value = ''
      }
    }

    const validateGateway = () => {
      const ipRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
      if (vlanConfig.gateway && !ipRegex.test(vlanConfig.gateway)) {
        gatewayError.value = 'Invalid gateway address format'
      } else {
        gatewayError.value = ''
      }
    }

    const calculateNetworkAddress = (config) => {
      if (!config.ipAddress || !config.cidr) return 'N/A'
      const mask = parseInt(config.cidr.substring(1))
      const ipParts = config.ipAddress.split('.')
      if (ipParts.length === 4) {
        const networkParts = ipParts.map((part, i) => {
          if (mask >= (i + 1) * 8) return part
          if (mask <= i * 8) return '0'
          const bits = mask - (i * 8)
          const maskValue = 256 - Math.pow(2, 8 - bits)
          return parseInt(part) & maskValue
        })
        return networkParts.join('.')
      }
      return 'N/A'
    }

    const calculateBroadcastAddress = (config) => {
      if (!config.ipAddress || !config.cidr) return 'N/A'
      const mask = parseInt(config.cidr.substring(1))
      const ipParts = config.ipAddress.split('.')
      if (ipParts.length === 4) {
        const broadcastParts = ipParts.map((part, i) => {
          if (mask >= (i + 1) * 8) return part
          if (mask <= i * 8) return '255'
          const bits = mask - (i * 8)
          const maskValue = 256 - Math.pow(2, 8 - bits)
          return (parseInt(part) & maskValue) + (Math.pow(2, 8 - bits) - 1)
        })
        return broadcastParts.join('.')
      }
      return 'N/A'
    }

    const calculateUsableHosts = (config) => {
      if (!config.cidr) return 'N/A'
      const mask = parseInt(config.cidr.substring(1))
      if (mask === 31 || mask === 32) return '0'
      return Math.pow(2, 32 - mask) - 2
    }

    const createVLAN = async () => {
      if (!selectedInterface.value) {
        showNotification('Please select an interface first', 'error')
        return
      }
      
      if (!vlanConfig.id || vlanConfig.id < 1 || vlanConfig.id > 4094) {
        showNotification('Please enter a valid VLAN ID (1-4094)', 'error')
        return
      }

      if (!vlanConfig.ipAddress) {
        showNotification('Please enter an IP address', 'error')
        return
      }

      if (ipError.value || gatewayError.value) {
        showNotification('Please fix configuration errors', 'error')
        return
      }

      try {
        const response = await fetch(`${API_BASE_URL}/vlans`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            id: vlanConfig.id,
            parentInterface: selectedInterface.value.name,
            ipAddress: vlanConfig.ipAddress,
            cidr: vlanConfig.cidr,
            gateway: vlanConfig.gateway,
            monitorEnabled: vlanConfig.monitorEnabled
          })
        })

        if (!response.ok) {
          throw new Error('Failed to create VLAN')
        }

        const newVLAN = await response.json()
        
        existingVLANs.value.push({
          ...newVLAN,
          monitorEnabled: newVLAN.monitorEnabled || false
        })
        showNotification(`VLAN ${vlanConfig.id} created successfully`, 'success')
        resetVLANForm()
      } catch (error) {
        console.error('Error creating VLAN:', error)
        showNotification('Failed to create VLAN', 'error')
      }
    }

    const resetVLANForm = () => {
      vlanConfig.id = ''
      vlanConfig.ipAddress = ''
      vlanConfig.cidr = '/24'
      vlanConfig.gateway = ''
      vlanConfig.monitorEnabled = false
      ipError.value = ''
      gatewayError.value = ''
    }

    const editVLAN = (vlan) => {
      vlanConfig.id = vlan.id
      vlanConfig.ipAddress = vlan.ipAddress
      vlanConfig.cidr = vlan.cidr
      vlanConfig.gateway = vlan.gateway
      vlanConfig.monitorEnabled = vlan.monitorEnabled || false
      
      // Find and select the parent interface
      const parentIface = interfaces.value.find(i => i.name === vlan.parentInterface)
      if (parentIface) {
        selectedInterface.value = parentIface
      }
      
      // Scroll to form
      document.querySelector('.config-panel').scrollIntoView({ behavior: 'smooth' })
    }

    const deleteVLAN = async (vlan) => {
      if (!confirm(`Delete VLAN ${vlan.id}?`)) return

      try {
        const response = await fetch(`${API_BASE_URL}/vlans/${vlan.id}`, {
          method: 'DELETE'
        })

        if (!response.ok) {
          throw new Error('Failed to delete VLAN')
        }

        existingVLANs.value = existingVLANs.value.filter(v => v.id !== vlan.id)
        showNotification(`VLAN ${vlan.id} deleted`, 'success')
      } catch (error) {
        console.error('Error deleting VLAN:', error)
        showNotification('Failed to delete VLAN', 'error')
      }
    }

    const showNotification = (message, type = 'success') => {
      notification.message = message
      notification.type = type
      notification.show = true
      
      setTimeout(() => {
        notification.show = false
      }, 3000)
    }

    // Initialize
    onMounted(() => {
      fetchInterfaces()
    })

    return {
      interfaces: interfacesWithMonitor,
      selectedInterface,
      isLoading,
      vlanConfig,
      existingVLANs,
      ipError,
      gatewayError,
      notification,
      fetchInterfaces,
      selectInterface,
      toggleInterfaceMonitor,
      toggleVLANMonitor,
      saveInterfaceConfig,
      validateIP,
      validateGateway,
      calculateNetworkAddress,
      calculateBroadcastAddress,
      calculateUsableHosts,
      createVLAN,
      resetVLANForm,
      editVLAN,
      deleteVLAN
    }
  }
}
</script>

<style scoped>
/* Dark Mode Theme */
.network-dashboard {
  min-height: 100vh;
  width: 100%;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  color: #e2e8f0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  padding: 24px;
  box-sizing: border-box;
}

.dashboard-header {
  margin-bottom: 24px;
  width: 100%;
}

.dashboard-header h1 {
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

.dashboard-content {
  display: flex;
  gap: 24px;
  height: calc(100vh - 120px);
  width: 100%;
}

/* Panels */
.interfaces-list-panel,
.config-panel {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.interfaces-list-panel {
  flex: 1;
  min-width: 400px;
  max-width: 500px;
}

.config-panel {
  flex: 2;
  min-width: 500px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  background: rgba(15, 23, 42, 0.4);
}

.panel-header h2 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: #f8fafc;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.refresh-btn {
  background: #1e293b;
  border: 1px solid #334155;
  color: #cbd5e1;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.refresh-btn:hover:not(:disabled) {
  background: #2d3748;
  border-color: #3b82f6;
  color: #60a5fa;
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.refresh-icon {
  display: inline-block;
  transition: transform 0.3s;
}

.refresh-icon.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
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
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #ef4444;
  color: white;
}

/* Loading State */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #94a3b8;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #1e293b;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

/* Interfaces List */
.interfaces-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.interfaces-list h4 {
  margin: 0 0 12px 16px;
  color: #94a3b8;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.interface-item {
  display: flex;
  align-items: center;
  padding: 16px;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 12px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.interface-item:hover {
  background: #1e293b;
  border-color: #3b82f6;
  transform: translateX(4px);
}

.interface-item.selected {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.1);
  box-shadow: 0 0 20px rgba(59, 130, 246, 0.2);
}

.interface-item.default-route {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.1);
}

.interface-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 16px;
}

.interface-indicator.up {
  background: #10b981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
}

.interface-indicator.down {
  background: #ef4444;
}

.interface-info {
  flex: 1;
}

.interface-name {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  flex-wrap: wrap;
}

.interface-name .name {
  font-weight: 600;
  color: #f8fafc;
}

.default-badge {
  background: #3b82f6;
  color: white;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.7rem;
  font-weight: 500;
}

.status-badge {
  font-size: 0.7rem;
  padding: 2px 8px;
  border-radius: 12px;
  text-transform: uppercase;
}

.status-badge.up {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.status-badge.down {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.interface-details {
  display: flex;
  gap: 16px;
  font-size: 0.85rem;
  color: #94a3b8;
}

.interface-details .ip {
  color: #60a5fa;
}

.interface-details .mac {
  font-family: 'Monaco', 'Courier New', monospace;
}

.interface-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  opacity: 0;
  transition: opacity 0.2s;
}

.interface-item:hover .interface-actions {
  opacity: 1;
}

/* Monitor Checkbox Styles */
.monitor-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  position: relative;
  padding-left: 28px;
  user-select: none;
}

.monitor-checkbox input {
  position: absolute;
  opacity: 0;
  cursor: pointer;
  height: 0;
  width: 0;
}

.checkmark {
  position: absolute;
  left: 0;
  height: 18px;
  width: 18px;
  background: #1e293b;
  border: 2px solid #334155;
  border-radius: 4px;
  transition: all 0.2s;
}

.monitor-checkbox:hover input ~ .checkmark {
  background: #2d3748;
  border-color: #3b82f6;
}

.monitor-checkbox input:checked ~ .checkmark {
  background: #3b82f6;
  border-color: #3b82f6;
}

.checkmark:after {
  content: "";
  position: absolute;
  display: none;
  left: 5px;
  top: 2px;
  width: 4px;
  height: 8px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.monitor-checkbox input:checked ~ .checkmark:after {
  display: block;
}

.monitor-label {
  font-size: 0.85rem;
  color: #cbd5e1;
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

/* Config Form */
.config-form {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.info-section {
  display: flex;
  gap: 24px;
  padding: 16px;
  background: #0f172a;
  border-radius: 12px;
  border: 1px solid #334155;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label {
  font-size: 0.8rem;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-value {
  font-size: 1.1rem;
  font-weight: 600;
  color: #f8fafc;
}

.info-value.status.up {
  color: #34d399;
}

.monitor-status {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
}

.monitor-status.enabled {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.monitor-status:not(.enabled) {
  background: rgba(107, 114, 128, 0.1);
  color: #9ca3af;
  border: 1px solid rgba(107, 114, 128, 0.3);
}

.form-section {
  margin-bottom: 24px;
  padding: 20px;
  background: #0f172a;
  border-radius: 12px;
  border: 1px solid #334155;
}

.form-section h3 {
  margin: 0 0 16px 0;
  color: #f8fafc;
  font-size: 1rem;
  font-weight: 600;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  font-size: 0.9rem;
  color: #cbd5e1;
  margin-bottom: 6px;
  font-weight: 500;
}

.form-input {
  width: 100%;
  padding: 10px 14px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 8px;
  color: #e2e8f0;
  font-size: 0.95rem;
  transition: all 0.2s;
  box-sizing: border-box;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

.form-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-input::placeholder {
  color: #64748b;
}

.input-hint {
  display: block;
  font-size: 0.8rem;
  color: #64748b;
  margin-top: 4px;
}

/* CIDR Select */
.cidr-select {
  width: 100%;
  padding: 10px 14px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 8px;
  color: #e2e8f0;
  font-size: 0.95rem;
  cursor: pointer;
  transition: all 0.2s;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%2394a3b8' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 40px;
}

.cidr-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

.cidr-select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.cidr-select option {
  background: #1e293b;
  color: #e2e8f0;
}

.error-text {
  display: block;
  color: #f87171;
  font-size: 0.8rem;
  margin-top: 4px;
}

/* Monitor Group */
.monitor-group {
  margin-top: 20px;
  padding: 12px;
  background: #1e293b;
  border-radius: 8px;
  border: 1px solid #334155;
}

/* Network Preview */
.network-preview {
  margin-top: 20px;
  padding: 16px;
  background: #1e293b;
  border-radius: 8px;
  border: 1px solid #334155;
}

.network-preview h4 {
  margin: 0 0 12px 0;
  color: #cbd5e1;
  font-size: 0.9rem;
}

.preview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
}

.preview-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.preview-label {
  font-size: 0.8rem;
  color: #94a3b8;
}

.preview-value {
  font-size: 0.95rem;
  color: #60a5fa;
  font-family: 'Monaco', 'Courier New', monospace;
  word-break: break-word;
}

/* VLAN Compact List */
.vlan-list-compact {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.vlan-compact-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: #1e293b;
  border-radius: 8px;
  border: 1px solid #334155;
}

.vlan-compact-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.vlan-compact-id {
  font-weight: 600;
  color: #60a5fa;
  font-size: 0.9rem;
}

.vlan-compact-ip {
  color: #94a3b8;
  font-size: 0.85rem;
}

.monitoring-badge {
  font-size: 1rem;
  color: #34d399;
}

.vlan-compact-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Small checkbox for VLAN items */
.icon-checkbox {
  position: relative;
  display: inline-block;
  width: 20px;
  height: 20px;
  cursor: pointer;
}

.icon-checkbox input {
  position: absolute;
  opacity: 0;
  cursor: pointer;
  height: 0;
  width: 0;
}

.checkmark-small {
  position: absolute;
  top: 0;
  left: 0;
  height: 18px;
  width: 18px;
  background: #1e293b;
  border: 2px solid #334155;
  border-radius: 4px;
  transition: all 0.2s;
}

.icon-checkbox:hover input ~ .checkmark-small {
  background: #2d3748;
  border-color: #3b82f6;
}

.icon-checkbox input:checked ~ .checkmark-small {
  background: #3b82f6;
  border-color: #3b82f6;
}

.checkmark-small:after {
  content: "";
  position: absolute;
  display: none;
  left: 4px;
  top: 1px;
  width: 4px;
  height: 8px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.icon-checkbox input:checked ~ .checkmark-small:after {
  display: block;
}

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

.icon-btn:hover {
  background: #2d3748;
  color: #60a5fa;
}

.icon-btn.delete:hover {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
}

/* Form Actions */
.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  flex-wrap: wrap;
}

.btn {
  padding: 12px 20px;
  border-radius: 8px;
  font-size: 0.95rem;
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

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.btn-icon {
  font-size: 1.1rem;
}

/* Empty States */
.empty-state {
  text-align: center;
  padding: 60px 20px;
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

.empty-state h3 {
  color: #f8fafc;
  margin-bottom: 8px;
  font-size: 1.1rem;
}

.empty-state p {
  font-size: 0.9rem;
}

/* Notifications */
.notification {
  position: fixed;
  top: 24px;
  right: 24px;
  padding: 16px 20px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 12px;
  animation: slideIn 0.3s ease;
  z-index: 1000;
  max-width: 400px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
}

.notification.success {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
  color: #34d399;
}

.notification.error {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
}

.notification-icon {
  font-size: 1.2rem;
}

.notification-message {
  flex: 1;
  font-size: 0.95rem;
}

.notification-close {
  background: transparent;
  border: none;
  color: currentColor;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.notification-close:hover {
  opacity: 1;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
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

/* Responsive */
@media (max-width: 1024px) {
  .dashboard-content {
    flex-direction: column;
    height: auto;
  }
  
  .interfaces-list-panel,
  .config-panel {
    max-width: 100%;
    min-width: 100%;
  }
  
  .config-panel {
    margin-top: 20px;
  }
}

@media (max-width: 768px) {
  .network-dashboard {
    padding: 16px;
  }
  
  .interface-actions {
    opacity: 1;
  }
  
  .interface-details {
    flex-direction: column;
    gap: 4px;
  }
  
  .info-section {
    flex-direction: column;
    gap: 12px;
  }
  
  .form-actions {
    flex-direction: column;
  }
  
  .btn {
    width: 100%;
    justify-content: center;
  }
  
  .vlan-compact-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .vlan-compact-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>