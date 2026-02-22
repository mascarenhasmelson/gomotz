<template>
  <div class="vlan-dashboard">
    <!-- Header -->
    <div class="dashboard-header">
      <h1>VLAN Configuration</h1>
      <p class="subtitle">Configure and manage VLAN interfaces with DHCP or static IP addressing</p>
    </div>

    <!-- Main Content -->
    <div class="dashboard-content">
      <!-- Left Panel - VLAN List -->
      <div class="vlan-list-panel">
        <div class="panel-header">
          <h2>Discovered VLANs</h2>
          <div class="header-actions">
            <button class="refresh-btn" @click="scanVLANs" :disabled="isScanning">
              <span class="refresh-icon" :class="{ 'spinning': isScanning }">↻</span>
              {{ isScanning ? 'Scanning...' : 'Scan VLANs' }}
            </button>
          </div>
        </div>

        <!-- VLAN Scanning Status -->
        <div v-if="isScanning" class="scanning-status">
          <div class="scanning-progress">
            <div class="progress-bar" :style="{ width: scanProgress + '%' }"></div>
          </div>
          <p>Discovering VLANs on network... {{ scanProgress }}%</p>
        </div>

        <!-- VLAN List -->
        <div class="vlan-list">
          <div 
            v-for="vlan in vlans" 
            :key="vlan.id"
            class="vlan-item"
            :class="{ 'selected': selectedVlan && selectedVlan.id === vlan.id }"
            @click="selectVLAN(vlan)"
          >
            <div class="vlan-indicator" :style="{ backgroundColor: getVLANColor(vlan.id) }"></div>
            <div class="vlan-info">
              <div class="vlan-name">
                <span class="vlan-id">VLAN {{ vlan.id }}</span>
                <span class="vlan-status" :class="vlan.status">{{ vlan.status }}</span>
              </div>
              <div class="vlan-details">
                <span class="vlan-subnet">{{ vlan.subnet || 'No subnet' }}</span>
                <span class="vlan-devices">{{ vlan.deviceCount }} devices</span>
              </div>
            </div>
            <div class="vlan-actions">
              <button class="action-btn" @click.stop="configureVLAN(vlan)" title="Configure">
                ⚙️
              </button>
            </div>
          </div>

          <!-- Empty State -->
          <div v-if="vlans.length === 0 && !isScanning" class="empty-state">
            <div class="empty-icon">🌐</div>
            <h3>No VLANs Discovered</h3>
            <p>Click the "Scan VLANs" button to discover VLANs on your network</p>
          </div>
        </div>
      </div>

      <!-- Right Panel - VLAN Configuration -->
      <div class="config-panel" v-if="selectedVlan">
        <div class="panel-header">
          <h2>Configure VLAN {{ selectedVlan.id }}</h2>
          <button class="close-btn" @click="selectedVlan = null">×</button>
        </div>

        <div class="config-form">
          <!-- VLAN Info -->
          <div class="info-section">
            <div class="info-item">
              <span class="info-label">VLAN ID</span>
              <span class="info-value">{{ selectedVlan.id }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Status</span>
              <span class="info-value status" :class="selectedVlan.status">{{ selectedVlan.status }}</span>
            </div>
          </div>

          <!-- IP Configuration Method -->
          <div class="form-section">
            <h3>IP Configuration</h3>
            <div class="config-methods">
              <label class="method-radio">
                <input 
                  type="radio" 
                  v-model="configMethod" 
                  value="dhcp"
                  @change="onConfigMethodChange"
                >
                <span class="radio-label">DHCP</span>
                <span class="method-desc">Automatically obtain IP from DHCP server</span>
              </label>
              
              <label class="method-radio">
                <input 
                  type="radio" 
                  v-model="configMethod" 
                  value="static"
                  @change="onConfigMethodChange"
                >
                <span class="radio-label">Static IP</span>
                <span class="method-desc">Manually configure IP address and routing</span>
              </label>
            </div>
          </div>

          <!-- Static IP Configuration -->
          <div v-if="configMethod === 'static'" class="form-section">
            <h3>Static IP Settings</h3>
            
            <div class="form-group">
              <label for="ipAddress">IP Address</label>
              <input 
                type="text" 
                id="ipAddress"
                v-model="staticConfig.ipAddress"
                placeholder="e.g., 192.168.1.10"
                class="form-input"
                @input="validateIP"
              />
              <span v-if="ipError" class="error-text">{{ ipError }}</span>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label for="subnetMask">Subnet Mask</label>
                <select 
                  id="subnetMask"
                  v-model="staticConfig.subnetMask"
                  class="form-select"
                >
                  <option value="/32">/32 (255.255.255.255)</option>
                  <option value="/31">/31 (255.255.255.254)</option>
                  <option value="/30">/30 (255.255.255.252)</option>
                  <option value="/29">/29 (255.255.255.248)</option>
                  <option value="/28">/28 (255.255.255.240)</option>
                  <option value="/27">/27 (255.255.255.224)</option>
                  <option value="/26">/26 (255.255.255.192)</option>
                  <option value="/25">/25 (255.255.255.128)</option>
                  <option value="/24">/24 (255.255.255.0)</option>
                  <option value="/23">/23 (255.255.254.0)</option>
                  <option value="/22">/22 (255.255.252.0)</option>
                  <option value="/21">/21 (255.255.248.0)</option>
                  <option value="/20">/20 (255.255.240.0)</option>
                  <option value="/19">/19 (255.255.224.0)</option>
                  <option value="/18">/18 (255.255.192.0)</option>
                  <option value="/17">/17 (255.255.128.0)</option>
                  <option value="/16">/16 (255.255.0.0)</option>
                  <option value="/15">/15 (255.254.0.0)</option>
                  <option value="/14">/14 (255.252.0.0)</option>
                  <option value="/13">/13 (255.248.0.0)</option>
                  <option value="/12">/12 (255.240.0.0)</option>
                  <option value="/11">/11 (255.224.0.0)</option>
                  <option value="/10">/10 (255.192.0.0)</option>
                  <option value="/9">/9 (255.128.0.0)</option>
                  <option value="/8">/8 (255.0.0.0)</option>
                </select>
              </div>

              <div class="form-group">
                <label for="cidr">CIDR Notation</label>
                <input 
                  type="text" 
                  id="cidr"
                  v-model="staticConfig.cidr"
                  placeholder="e.g., 24"
                  class="form-input"
                  @input="updateFromCIDR"
                />
              </div>
            </div>

            <div class="form-group">
              <label for="gateway">Default Gateway</label>
              <input 
                type="text" 
                id="gateway"
                v-model="staticConfig.gateway"
                placeholder="e.g., 192.168.1.1"
                class="form-input"
                @input="validateGateway"
              />
              <span v-if="gatewayError" class="error-text">{{ gatewayError }}</span>
            </div>

            <div class="form-group">
              <label for="dnsServers">DNS Servers (optional)</label>
              <input 
                type="text" 
                id="dnsServers"
                v-model="staticConfig.dnsServers"
                placeholder="e.g., 8.8.8.8, 8.8.4.4"
                class="form-input"
              />
            </div>

            <div class="network-preview">
              <h4>Network Configuration Preview</h4>
              <div class="preview-grid">
                <div class="preview-item">
                  <span class="preview-label">Network Address:</span>
                  <span class="preview-value">{{ calculateNetworkAddress() }}</span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">Broadcast Address:</span>
                  <span class="preview-value">{{ calculateBroadcastAddress() }}</span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">Usable Hosts:</span>
                  <span class="preview-value">{{ calculateUsableHosts() }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- VLAN Sniffing Section -->
          <!-- <div class="form-section">
            <div class="section-header" @click="toggleSniffing">
              <h3>VLAN Sniffing</h3>
              <button class="toggle-btn">
                <span :class="{ 'rotated': showSniffing }">▼</span>
              </button>
            </div>
            
            <div v-if="showSniffing" class="sniffing-content">
              <p class="sniffing-desc">Enable VLAN sniffing to automatically detect VLANs on the network</p>
              
              <label class="toggle-switch">
                <input 
                  type="checkbox" 
                  v-model="sniffingEnabled"
                  @change="toggleSniffingMode"
                >
                <span class="toggle-slider"></span>
                <span class="toggle-label">Enable VLAN Sniffing</span>
              </label>

              <div v-if="sniffingEnabled" class="sniffing-results">
                <div class="sniffing-header">
                  <h4>Detected VLANs</h4>
                  <button class="small-btn" @click="refreshSniffing" :disabled="isSniffing">
                    {{ isSniffing ? 'Sniffing...' : 'Refresh' }}
                  </button>
                </div>

                <div class="sniffing-list">
                  <div 
                    v-for="vlan in sniffedVlans" 
                    :key="vlan.id"
                    class="sniffed-item"
                    @click="selectSniffedVLAN(vlan)"
                  >
                    <div class="sniffed-indicator" :style="{ backgroundColor: getVLANColor(vlan.id) }"></div>
                    <div class="sniffed-info">
                      <span class="sniffed-id">VLAN {{ vlan.id }}</span>
                      <span class="sniffed-name">{{ vlan.name || 'Unnamed' }}</span>
                    </div>
                    <div class="sniffed-status" :class="vlan.status">{{ vlan.status }}</div>
                  </div>

                  <div v-if="sniffedVlans.length === 0 && !isSniffing" class="sniffing-empty">
                    <p>No VLANs detected yet. Click "Refresh" to start sniffing.</p>
                  </div>

                  <div v-if="isSniffing" class="sniffing-loading">
                    <div class="spinner-small"></div>
                    <p>Sniffing network traffic...</p>
                  </div>
                </div>
              </div>
            </div>
          </div> -->

          <!-- Action Buttons -->
          <div class="form-actions">
            <button class="btn btn-primary" @click="applyConfiguration">
              <span class="btn-icon">✓</span>
              Apply Configuration
            </button>
            <button class="btn btn-secondary" @click="testConfiguration">
              <span class="btn-icon">🔍</span>
              Test Connection
            </button>
            <button class="btn btn-danger" @click="resetConfiguration">
              <span class="btn-icon">↺</span>
              Reset
            </button>
          </div>
        </div>
      </div>

      <!-- Empty State for Right Panel -->
      <div v-else class="config-panel empty-panel">
        <div class="empty-selection">
          <div class="empty-icon">👈</div>
          <h3>Select a VLAN</h3>
          <p>Choose a VLAN from the list to configure its settings</p>
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
import { ref, reactive, onMounted, watch } from 'vue'

export default {
  name: 'VLANDashboard',
  
  setup() {
    // State
    const vlans = ref([])
    const selectedVlan = ref(null)
    const configMethod = ref('dhcp')
    const isScanning = ref(false)
    const scanProgress = ref(0)
    const showSniffing = ref(false)
    const sniffingEnabled = ref(false)
    const isSniffing = ref(false)
    const sniffedVlans = ref([])
    
    // Static configuration
    const staticConfig = reactive({
      ipAddress: '',
      subnetMask: '/24',
      cidr: '24',
      gateway: '',
      dnsServers: ''
    })

    // Validation errors
    const ipError = ref('')
    const gatewayError = ref('')

    // Notification
    const notification = reactive({
      show: false,
      type: 'success',
      message: ''
    })

    // Sample VLAN data (replace with actual backend data)
    const sampleVlans = [
      { id: 10, name: 'Management', subnet: '192.168.10.0/24', status: 'active', deviceCount: 12 },
      { id: 20, name: 'Data', subnet: '192.168.20.0/24', status: 'active', deviceCount: 25 },
      { id: 30, name: 'Voice', subnet: '192.168.30.0/24', status: 'active', deviceCount: 8 },
      { id: 40, name: 'Guest', subnet: '192.168.40.0/24', status: 'inactive', deviceCount: 0 },
      { id: 50, name: 'DMZ', subnet: '10.0.50.0/24', status: 'active', deviceCount: 5 },
      { id: 100, name: 'Storage', subnet: '172.16.100.0/24', status: 'active', deviceCount: 3 }
    ]

    // Methods
    const scanVLANs = async () => {
      isScanning.value = true
      scanProgress.value = 0
      
      // Simulate scanning progress
      const interval = setInterval(() => {
        if (scanProgress.value < 100) {
          scanProgress.value += 10
        } else {
          clearInterval(interval)
          // Load sample VLANs after scan
          vlans.value = sampleVlans
          isScanning.value = false
          showNotification('VLAN scan completed', 'success')
        }
      }, 500)
    }

    const selectVLAN = (vlan) => {
      selectedVlan.value = vlan
      // Reset configuration
      configMethod.value = 'dhcp'
      resetStaticConfig()
    }

    const configureVLAN = (vlan) => {
      selectVLAN(vlan)
    }

    const getVLANColor = (id) => {
      const colors = [
        '#3b82f6', '#10b981', '#f59e0b', '#ef4444', 
        '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16'
      ]
      return colors[id % colors.length]
    }

    const onConfigMethodChange = () => {
      if (configMethod.value === 'static') {
        // Pre-fill with some defaults
        if (!staticConfig.ipAddress) {
          staticConfig.ipAddress = '192.168.1.10'
        }
        if (!staticConfig.gateway) {
          staticConfig.gateway = '192.168.1.1'
        }
      }
    }

    const validateIP = () => {
      const ipRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
      if (staticConfig.ipAddress && !ipRegex.test(staticConfig.ipAddress)) {
        ipError.value = 'Invalid IP address format'
      } else {
        ipError.value = ''
      }
    }

    const validateGateway = () => {
      const ipRegex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
      if (staticConfig.gateway && !ipRegex.test(staticConfig.gateway)) {
        gatewayError.value = 'Invalid gateway address format'
      } else {
        gatewayError.value = ''
      }
    }

    const updateFromCIDR = () => {
      const cidr = parseInt(staticConfig.cidr)
      if (cidr >= 1 && cidr <= 32) {
        staticConfig.subnetMask = `/${cidr}`
      }
    }

    const calculateNetworkAddress = () => {
      if (!staticConfig.ipAddress || !staticConfig.subnetMask) return 'N/A'
      // Simple calculation - in production use proper IP math
      const mask = parseInt(staticConfig.subnetMask.substring(1))
      const ipParts = staticConfig.ipAddress.split('.')
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

    const calculateBroadcastAddress = () => {
      if (!staticConfig.ipAddress || !staticConfig.subnetMask) return 'N/A'
      const mask = parseInt(staticConfig.subnetMask.substring(1))
      const ipParts = staticConfig.ipAddress.split('.')
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

    const calculateUsableHosts = () => {
      if (!staticConfig.subnetMask) return 'N/A'
      const mask = parseInt(staticConfig.subnetMask.substring(1))
      if (mask === 31 || mask === 32) return '0'
      return Math.pow(2, 32 - mask) - 2
    }

    const resetStaticConfig = () => {
      staticConfig.ipAddress = ''
      staticConfig.subnetMask = '/24'
      staticConfig.cidr = '24'
      staticConfig.gateway = ''
      staticConfig.dnsServers = ''
      ipError.value = ''
      gatewayError.value = ''
    }

    const toggleSniffing = () => {
      showSniffing.value = !showSniffing.value
    }

    const toggleSniffingMode = async () => {
      if (sniffingEnabled.value) {
        await refreshSniffing()
      }
    }

    const refreshSniffing = async () => {
      isSniffing.value = true
      
      // Simulate sniffing
      setTimeout(() => {
        sniffedVlans.value = [
          { id: 10, name: 'Management', status: 'active' },
          { id: 20, name: 'Data', status: 'active' },
          { id: 30, name: 'Voice', status: 'active' },
          { id: 55, name: 'IoT', status: 'new' },
          { id: 66, name: 'Security', status: 'new' }
        ]
        isSniffing.value = false
        showNotification('VLAN sniffing completed', 'success')
      }, 2000)
    }

    const selectSniffedVLAN = (vlan) => {
      // Check if VLAN already exists
      const existingVlan = vlans.value.find(v => v.id === vlan.id)
      if (existingVlan) {
        selectVLAN(existingVlan)
      } else {
        // Add new VLAN to list
        const newVlan = {
          id: vlan.id,
          name: vlan.name,
          status: 'new',
          deviceCount: 0
        }
        vlans.value.push(newVlan)
        selectVLAN(newVlan)
        showNotification(`VLAN ${vlan.id} added to configuration`, 'success')
      }
      showSniffing.value = false
    }

    const applyConfiguration = async () => {
      if (configMethod.value === 'static') {
        if (ipError.value || gatewayError.value) {
          showNotification('Please fix configuration errors', 'error')
          return
        }
        if (!staticConfig.ipAddress || !staticConfig.gateway) {
          showNotification('Please fill in all required fields', 'error')
          return
        }
      }

      // Simulate API call
      const config = {
        vlanId: selectedVlan.value.id,
        method: configMethod.value,
        ...(configMethod.value === 'static' && {
          ipAddress: staticConfig.ipAddress,
          subnetMask: staticConfig.subnetMask,
          gateway: staticConfig.gateway,
          dnsServers: staticConfig.dnsServers.split(',').map(s => s.trim())
        })
      }

      console.log('Applying configuration:', config)
      
      // Simulate success
      showNotification(`Configuration applied to VLAN ${selectedVlan.value.id}`, 'success')
    }

    const testConfiguration = () => {
      showNotification('Testing connection...', 'success')
      setTimeout(() => {
        showNotification('Connection test successful', 'success')
      }, 2000)
    }

    const resetConfiguration = () => {
      configMethod.value = 'dhcp'
      resetStaticConfig()
      showNotification('Configuration reset', 'success')
    }

    const showNotification = (message, type = 'success') => {
      notification.message = message
      notification.type = type
      notification.show = true
      
      setTimeout(() => {
        notification.show = false
      }, 3000)
    }

    // Watch for VLAN selection changes
    watch(selectedVlan, (newVlan) => {
      if (newVlan) {
        // Could load existing config for this VLAN
        console.log('Loading config for VLAN:', newVlan.id)
      }
    })

    // Initial load - REMOVED auto-scan
    onMounted(() => {
      // No auto-scan - wait for user to click button
      console.log('VLAN Dashboard ready - click Scan to discover VLANs')
    })

    return {
      vlans,
      selectedVlan,
      configMethod,
      isScanning,
      scanProgress,
      showSniffing,
      sniffingEnabled,
      isSniffing,
      sniffedVlans,
      staticConfig,
      ipError,
      gatewayError,
      notification,
      scanVLANs,
      selectVLAN,
      configureVLAN,
      getVLANColor,
      onConfigMethodChange,
      validateIP,
      validateGateway,
      updateFromCIDR,
      calculateNetworkAddress,
      calculateBroadcastAddress,
      calculateUsableHosts,
      toggleSniffing,
      toggleSniffingMode,
      refreshSniffing,
      selectSniffedVLAN,
      applyConfiguration,
      testConfiguration,
      resetConfiguration
    }
  }
}
</script>

<style scoped>
/* Dark Mode Theme */
.vlan-dashboard {
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
.vlan-list-panel,
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

.vlan-list-panel {
  flex: 1;
  min-width: 350px;
  max-width: 450px;
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

/* Scanning Status */
.scanning-status {
  padding: 16px 24px;
  background: #0f172a;
  border-bottom: 1px solid #334155;
}

.scanning-progress {
  height: 4px;
  background: #1e293b;
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: 8px;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #8b5cf6);
  transition: width 0.3s ease;
}

.scanning-status p {
  margin: 0;
  color: #94a3b8;
  font-size: 0.9rem;
}

/* VLAN List */
.vlan-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.vlan-item {
  display: flex;
  align-items: center;
  padding: 16px;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 12px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.vlan-item:hover {
  background: #1e293b;
  border-color: #3b82f6;
  transform: translateX(4px);
}

.vlan-item.selected {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.1);
  box-shadow: 0 0 20px rgba(59, 130, 246, 0.2);
}

.vlan-indicator {
  width: 4px;
  height: 40px;
  border-radius: 2px;
  margin-right: 16px;
}

.vlan-info {
  flex: 1;
}

.vlan-name {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.vlan-id {
  font-weight: 600;
  color: #f8fafc;
}

.vlan-status {
  font-size: 0.75rem;
  padding: 2px 8px;
  border-radius: 12px;
  text-transform: uppercase;
}

.vlan-status.active {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.vlan-status.inactive {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.vlan-status.new {
  background: rgba(59, 130, 246, 0.1);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.vlan-details {
  display: flex;
  gap: 12px;
  font-size: 0.85rem;
  color: #94a3b8;
}

.vlan-actions {
  opacity: 0;
  transition: opacity 0.2s;
}

.vlan-item:hover .vlan-actions {
  opacity: 1;
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
}

/* Empty States */
.empty-state,
.empty-selection {
  text-align: center;
  padding: 60px 24px;
  color: #94a3b8;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 20px;
  opacity: 0.5;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.empty-state h3,
.empty-selection h3 {
  color: #f8fafc;
  margin-bottom: 8px;
}

.empty-panel {
  display: flex;
  align-items: center;
  justify-content: center;
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

.info-value.status.active {
  color: #34d399;
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

.config-methods {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.method-radio {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.method-radio:hover {
  background: #2d3748;
  border-color: #3b82f6;
}

.method-radio input[type="radio"] {
  margin-right: 12px;
  accent-color: #3b82f6;
  width: 16px;
  height: 16px;
}

.radio-label {
  font-weight: 600;
  color: #f8fafc;
  margin-right: 12px;
  min-width: 80px;
}

.method-desc {
  color: #94a3b8;
  font-size: 0.9rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
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

.form-input,
.form-select {
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

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

.form-input::placeholder {
  color: #64748b;
}

.error-text {
  display: block;
  color: #f87171;
  font-size: 0.8rem;
  margin-top: 4px;
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
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
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

/* VLAN Sniffing */
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  padding: 8px 0;
}

.toggle-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  font-size: 1rem;
  padding: 4px;
}

.toggle-btn span {
  display: inline-block;
  transition: transform 0.3s;
}

.toggle-btn span.rotated {
  transform: rotate(180deg);
}

.sniffing-content {
  margin-top: 16px;
}

.sniffing-desc {
  color: #94a3b8;
  font-size: 0.9rem;
  margin-bottom: 16px;
}

.toggle-switch {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  margin-bottom: 20px;
}

.toggle-switch input {
  display: none;
}

.toggle-slider {
  position: relative;
  width: 48px;
  height: 24px;
  background: #1e293b;
  border-radius: 12px;
  border: 1px solid #334155;
  transition: all 0.2s;
}

.toggle-slider::before {
  content: '';
  position: absolute;
  width: 20px;
  height: 20px;
  background: #94a3b8;
  border-radius: 50%;
  top: 2px;
  left: 2px;
  transition: all 0.2s;
}

input:checked + .toggle-slider {
  background: #3b82f6;
  border-color: #3b82f6;
}

input:checked + .toggle-slider::before {
  transform: translateX(24px);
  background: white;
}

.toggle-label {
  color: #cbd5e1;
  font-size: 0.95rem;
}

.sniffing-results {
  margin-top: 16px;
}

.sniffing-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.sniffing-header h4 {
  margin: 0;
  color: #cbd5e1;
}

.small-btn {
  padding: 4px 12px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #cbd5e1;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
}

.small-btn:hover:not(:disabled) {
  background: #2d3748;
  border-color: #3b82f6;
}

.small-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.sniffing-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #334155;
  border-radius: 8px;
}

.sniffed-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid #334155;
  cursor: pointer;
  transition: all 0.2s;
}

.sniffed-item:hover {
  background: #1e293b;
}

.sniffed-item:last-child {
  border-bottom: none;
}

.sniffed-indicator {
  width: 3px;
  height: 30px;
  border-radius: 1.5px;
  margin-right: 12px;
}

.sniffed-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.sniffed-id {
  font-weight: 600;
  color: #f8fafc;
  font-size: 0.9rem;
}

.sniffed-name {
  color: #94a3b8;
  font-size: 0.8rem;
}

.sniffed-status {
  font-size: 0.7rem;
  padding: 2px 8px;
  border-radius: 12px;
  text-transform: uppercase;
}

.sniffed-status.active {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.sniffed-status.new {
  background: rgba(59, 130, 246, 0.1);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.sniffing-empty {
  padding: 24px;
  text-align: center;
  color: #64748b;
}

.sniffing-loading {
  padding: 24px;
  text-align: center;
}

.spinner-small {
  width: 24px;
  height: 24px;
  border: 2px solid #1e293b;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 12px auto;
}

/* Form Actions */
.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
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

.btn-primary:hover {
  background: #2563eb;
  transform: translateY(-2px);
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
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.btn-icon {
  font-size: 1.1rem;
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
  
  .vlan-list-panel,
  .config-panel {
    max-width: 100%;
    min-width: 100%;
  }
  
  .config-panel {
    margin-top: 20px;
  }
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .form-actions {
    flex-direction: column;
  }
  
  .btn {
    width: 100%;
    justify-content: center;
  }
}
</style>