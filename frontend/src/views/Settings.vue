<template>
  <div class="network-dashboard">
    <!-- Header -->
    <div class="dashboard-header">
      <h1>Network Configuration</h1>
      <p class="subtitle">Configure network interfaces and monitor networks</p>
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
          <div 
            v-for="iface in interfaces" 
            :key="iface.interface"
            class="interface-item"
            :class="{ 
              'selected': selectedInterface && selectedInterface.interface === iface.interface
            }"
            @click="selectInterface(iface)"
          >
            <div class="interface-indicator" :class="iface.is_monitored ? 'monitored' : 'unmonitored'"></div>
            <div class="interface-info">
              <div class="interface-name">
                <span class="name">{{ iface.interface }}</span>
                <span v-if="iface.is_vlan" class="vlan-badge">VLAN</span>
                <span class="status-badge" :class="iface.is_monitored ? 'monitored' : 'unmonitored'">
                  {{ iface.is_monitored ? 'Monitored' : 'Not Monitored' }}
                </span>
              </div>
              <div class="interface-details">
                <span class="ip">{{ iface.ipv4 || 'No IP' }}</span>
                <span class="mac">{{ iface.mac }}</span>
              </div>
              <div class="interface-network" v-if="iface.network_name">
                <span class="network-label">Network:</span>
                <span class="network-name">{{ iface.network_name }}</span>
              </div>
            </div>
            <div class="interface-actions">
              <label class="monitor-checkbox" @click.stop>
                <input 
                  type="checkbox" 
                  v-model="iface.is_monitored"
                  @change="toggleInterfaceMonitor(iface)"
                >
                <span class="checkmark"></span>
                <span class="monitor-label">Monitor</span>
              </label>
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

      <!-- Right Panel - Interface Details -->
      <div class="config-panel">
        <div class="panel-header">
          <h2>{{ selectedInterface ? `Interface Details: ${selectedInterface.interface}` : 'Interface Details' }}</h2>
          <button v-if="selectedInterface" class="close-btn" @click="selectedInterface = null">×</button>
        </div>

        <div class="config-form">
          <!-- Interface Details (when selected) -->
          <div v-if="selectedInterface" class="info-section">
            <div class="info-item">
              <span class="info-label">Interface</span>
              <span class="info-value">{{ selectedInterface.interface }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">IP Address</span>
              <span class="info-value ip">{{ selectedInterface.ipv4 || 'No IP' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">MAC Address</span>
              <span class="info-value mac">{{ selectedInterface.mac }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">CIDR</span>
              <span class="info-value">{{ selectedInterface.cidr ? '/' + selectedInterface.cidr : 'N/A' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Default Gateway</span>
              <span class="info-value">{{ selectedInterface.default_gateway || 'N/A' }}</span>
            </div>
            <div class="info-item" v-if="selectedInterface.network_name">
              <span class="info-label">Network Name</span>
              <span class="info-value network-name">{{ selectedInterface.network_name }}</span>
            </div>
           
            <div class="info-item">
              <span class="info-label">Monitor Status</span>
              <span class="info-value">
                <span class="monitor-status" :class="{ 'enabled': selectedInterface.is_monitored }">
                  {{ selectedInterface.is_monitored ? 'Enabled' : 'Disabled' }}
                </span>
              </span>
            </div>
          </div>

          <div v-else class="no-selection-message">
            <div class="message-icon">👆</div>
            <p>Select a network interface from the left panel to view details</p>
          </div>

          <!-- Network Creation Form (shown when interface has no network) -->
          <div v-if="selectedInterface && !selectedInterface.network_name" class="form-section">
            <h3>Create Network on {{ selectedInterface.interface }}</h3>
            
            <div class="form-group">
              <label for="networkName">Network Name <span class="required">*</span></label>
              <input 
                type="text" 
                id="networkName"
                v-model="networkConfig.name"
                placeholder="e.g., homewifi, lan, guest"
                class="form-input"
              />
              <span class="input-hint">Friendly name for this network</span>
            </div>

            <div class="form-group">
              <label for="scanInterval">Scan Interval (seconds)</label>
              <input 
                type="number" 
                id="scanInterval"
                v-model="networkConfig.scan_interval"
                min="10"
                max="3600"
                placeholder="30"
                class="form-input"
              />
              <span class="input-hint">How often to scan this network (10-3600 seconds)</span>
            </div>

            <div class="form-group monitor-group">
              <label class="monitor-checkbox">
                <input 
                  type="checkbox" 
                  v-model="networkConfig.monitoring_enabled"
                >
                <span class="checkmark"></span>
                <span class="monitor-label">Enable monitoring for this network</span>
              </label>
            </div>

            <div class="form-actions">
              <button class="btn btn-primary" @click="createNetwork" :disabled="!networkConfig.name || updating">
                <span class="btn-icon">➕</span>
                {{ updating ? 'Creating...' : 'Create Network' }}
              </button>
              <button class="btn btn-secondary" @click="resetNetworkForm">
                Reset
              </button>
            </div>
          </div>

          <!-- Existing Network Configuration (shown when interface has a network) -->
          <div v-if="selectedInterface && selectedInterface.network_name" class="form-section">
            <h3>Network Configuration</h3>
            <div class="network-info-card">
              <div class="network-header">
                <span class="network-name-large">{{ selectedInterface.network_name }}</span>
                <span class="network-status" :class="{ 'active': selectedInterface.is_monitored }">
                  {{ selectedInterface.is_monitored ? 'Monitoring Active' : 'Monitoring Disabled' }}
                </span>
              </div>
              <div class="network-details">
                <div class="network-detail-item">
                  <span class="detail-label">Interface:</span>
                  <span class="detail-value">{{ selectedInterface.interface }}</span>
                </div>
                <!-- <div class="network-detail-item">
                  <span class="detail-label">Network ID:</span>
                  <span class="detail-value">{{ selectedInterface.network_db_id }}</span>
                </div> -->
                <div class="network-detail-item">
                  <span class="detail-label">IP Range:</span>
                  <span class="detail-value">{{ selectedInterface.ipv4 }}/{{ selectedInterface.cidr }}</span>
                </div>
                <div class="network-detail-item">
                  <span class="detail-label">Gateway:</span>
                  <span class="detail-value">{{ selectedInterface.default_gateway || 'N/A' }}</span>
                </div>
                <!-- <div class="network-detail-item">
                  <span class="detail-label">Scan Interval:</span>
                  <span class="detail-value">{{ selectedInterface.scan_interval || 30 }} seconds</span>
                </div> -->
              </div>
              <div class="network-actions">
                <button class="btn btn-warning" @click="toggleMonitoring" :disabled="updating">
                  {{ updating ? 'Updating...' : (selectedInterface.is_monitored ? 'Stop Monitoring' : 'Start Monitoring') }}
                </button>
                <button class="btn btn-danger" @click="deleteNetwork" :disabled="updating">
                  🗑️ Delete Network
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Delete Network</h3>
          <button class="modal-close" @click="showDeleteModal = false">×</button>
        </div>
        <div class="modal-body">
          <p>Are you sure you want to delete network <strong>"{{ selectedInterface?.network_name }}"</strong>?</p>
          <p class="warning-text">This action cannot be undone. All devices in this network will no longer be monitored.</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteModal = false">Cancel</button>
          <button class="btn btn-danger" @click="confirmDeleteNetwork" :disabled="deleting">
            {{ deleting ? 'Deleting...' : 'Delete Network' }}
          </button>
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

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8082'

export default {
  name: 'NetworkDashboard',
  
  setup() {
    // State
    const interfaces = ref([])
    const selectedInterface = ref(null)
    const isLoading = ref(false)
    const updating = ref(false)
    const deleting = ref(false)
    const showDeleteModal = ref(false)
    
    // Network Configuration
    const networkConfig = reactive({
      name: '',
      scan_interval: 30,
      monitoring_enabled: true
    })

    // Notification
    const notification = reactive({
      show: false,
      type: 'success',
      message: ''
    })

    // Computed
    const isNetworkFormValid = computed(() => {
      return networkConfig.name && networkConfig.name.trim().length > 0
    })

    // API Methods
    const fetchInterfaces = async () => {
      isLoading.value = true
      
      try {
        const response = await fetch(`${API_BASE_URL}/v1/api/interfaces`)
        
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }
        
        const data = await response.json()
        interfaces.value = (data || []).map(iface => ({
          ...iface,
          is_monitored: iface.is_monitored || false
        }))
        
        showNotification('Network interfaces loaded successfully', 'success')
        console.log('Loaded interfaces:', interfaces.value)
      } catch (error) {
        console.error('Error fetching interfaces:', error)
        showNotification(`Failed to load interfaces: ${error.message}`, 'error')
        interfaces.value = []
      } finally {
        isLoading.value = false
      }
    }

    const selectInterface = (iface) => {
      selectedInterface.value = iface
      resetNetworkForm()
    }

    const toggleInterfaceMonitor = async (iface) => {
      const originalState = iface.is_monitored
      updating.value = true
      
      try {
        const response = await fetch(`${API_BASE_URL}/v1/api/interfaces/${iface.interface}/monitor`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ 
            name: iface.network_name || iface.interface,
            enabled: iface.is_monitored 
          })
        })
        
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }
        
        const result = await response.json()
        showNotification(`Monitoring ${iface.is_monitored ? 'enabled' : 'disabled'} for ${iface.interface}`, 'success')
        
        // Refresh interface details
        await fetchInterfaces()
        if (selectedInterface.value && selectedInterface.value.interface === iface.interface) {
          const updated = interfaces.value.find(i => i.interface === iface.interface)
          if (updated) selectedInterface.value = updated
        }
      } catch (error) {
        console.error('Error updating monitor state:', error)
        iface.is_monitored = originalState
        showNotification(`Failed to update monitoring state: ${error.message}`, 'error')
      } finally {
        updating.value = false
      }
    }

    const createNetwork = async () => {
      if (!selectedInterface.value) {
        showNotification('Please select an interface first', 'error')
        return
      }
      
      if (!networkConfig.name) {
        showNotification('Please enter a network name', 'error')
        return
      }
      
      updating.value = true
      
      try {
        const payload = {
          name: networkConfig.name,
          scan_interval: networkConfig.scan_interval,
          monitoring_enabled: networkConfig.monitoring_enabled
        }

        const response = await fetch(`${API_BASE_URL}/v1/api/interfaces/${selectedInterface.value.interface}/monitor`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(payload)
        })

        if (!response.ok) {
          const error = await response.json()
          throw new Error(error.message || `HTTP ${response.status}`)
        }

        const result = await response.json()
        showNotification(`Network "${networkConfig.name}" created successfully on ${selectedInterface.value.interface}`, 'success')
        
        // Refresh interfaces
        await fetchInterfaces()
        
        // Update selected interface with new data
        const updated = interfaces.value.find(i => i.interface === selectedInterface.value.interface)
        if (updated) selectedInterface.value = updated
        
        resetNetworkForm()
      } catch (error) {
        console.error('Error creating network:', error)
        showNotification(`Failed to create network: ${error.message}`, 'error')
      } finally {
        updating.value = false
      }
    }

    const toggleMonitoring = async () => {
      if (!selectedInterface.value) return
      
      const newState = !selectedInterface.value.is_monitored
      updating.value = true
      
      try {
        const response = await fetch(`${API_BASE_URL}/v1/api/interfaces/${selectedInterface.value.interface}/monitor`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ 
            name: selectedInterface.value.network_name,
            enabled: newState
          })
        })

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }

        showNotification(`Monitoring ${newState ? 'enabled' : 'disabled'} for ${selectedInterface.value.interface}`, 'success')
        
        // Refresh interfaces
        await fetchInterfaces()
        
        // Update selected interface
        const updated = interfaces.value.find(i => i.interface === selectedInterface.value.interface)
        if (updated) selectedInterface.value = updated
      } catch (error) {
        console.error('Error toggling monitoring:', error)
        showNotification(`Failed to update monitoring: ${error.message}`, 'error')
      } finally {
        updating.value = false
      }
    }

    const deleteNetwork = () => {
      if (!selectedInterface.value || !selectedInterface.value.network_db_id) {
        showNotification('No network found to delete', 'error')
        return
      }
      showDeleteModal.value = true
    }

    const confirmDeleteNetwork = async () => {
      if (!selectedInterface.value || !selectedInterface.value.network_db_id) return
      
      deleting.value = true
      
      try {
        const response = await fetch(`${API_BASE_URL}/v1/api/vlans/${selectedInterface.value.network_db_id}`, {
          method: 'DELETE'
        })

        if (!response.ok) {
          const error = await response.json()
          throw new Error(error.message || `HTTP ${response.status}`)
        }

        const result = await response.json()
        showNotification(`Network "${selectedInterface.value.network_name}" deleted successfully`, 'success')
        
        // Close modal
        showDeleteModal.value = false
        
        // Refresh interfaces
        await fetchInterfaces()
        
        // Clear selected interface if it was deleted
        const updated = interfaces.value.find(i => i.interface === selectedInterface.value.interface)
        if (updated) {
          selectedInterface.value = updated
        } else {
          selectedInterface.value = null
        }
      } catch (error) {
        console.error('Error deleting network:', error)
        showNotification(`Failed to delete network: ${error.message}`, 'error')
      } finally {
        deleting.value = false
      }
    }

    const resetNetworkForm = () => {
      networkConfig.name = ''
      networkConfig.scan_interval = 30
      networkConfig.monitoring_enabled = true
    }

    const showNotification = (message, type = 'success') => {
      notification.message = message
      notification.type = type
      notification.show = true
      
      setTimeout(() => {
        notification.show = false
      }, 4000)
    }

    // Initialize
    onMounted(() => {
      fetchInterfaces()
    })

    return {
      interfaces,
      selectedInterface,
      isLoading,
      updating,
      deleting,
      showDeleteModal,
      networkConfig,
      notification,
      isNetworkFormValid,
      fetchInterfaces,
      selectInterface,
      toggleInterfaceMonitor,
      createNetwork,
      toggleMonitoring,
      deleteNetwork,
      confirmDeleteNetwork,
      resetNetworkForm
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
}

.dashboard-header h1 {
  font-size: 2rem;
  font-weight: 700;
  margin: 0 0 8px 0;
  background: linear-gradient(135deg, #60a5fa 0%, #a78bfa 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
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
  min-width: 380px;
  max-width: 480px;
}

.config-panel {
  flex: 2;
  min-width: 520px;
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

.interface-indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 16px;
}

.interface-indicator.monitored {
  background: #10b981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
}

.interface-indicator.unmonitored {
  background: #64748b;
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

.vlan-badge {
  background: #8b5cf6;
  color: white;
  padding: 2px 6px;
  border-radius: 10px;
  font-size: 0.65rem;
  font-weight: 600;
}

.status-badge {
  font-size: 0.7rem;
  padding: 2px 8px;
  border-radius: 12px;
  text-transform: uppercase;
}

.status-badge.monitored {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.status-badge.unmonitored {
  background: rgba(107, 114, 128, 0.1);
  color: #9ca3af;
  border: 1px solid rgba(107, 114, 128, 0.3);
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
  font-family: monospace;
}

.interface-network {
  margin-top: 4px;
  font-size: 0.75rem;
  color: #94a3b8;
}

.network-label {
  color: #64748b;
  margin-right: 6px;
}

.network-name {
  color: #60a5fa;
  font-weight: 500;
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

/* Monitor Checkbox */
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
  font-size: 1rem;
  font-weight: 600;
  color: #f8fafc;
}

.info-value.ip {
  color: #60a5fa;
  font-family: monospace;
}

.info-value.mac {
  font-family: monospace;
  color: #94a3b8;
}

.info-value.network-name {
  color: #34d399;
  font-weight: 600;
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

.no-selection-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  color: #94a3b8;
}

.message-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
  animation: bounce 2s infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

.no-selection-message p {
  font-size: 0.95rem;
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

.required {
  color: #ef4444;
  margin-left: 4px;
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

.form-input::placeholder {
  color: #64748b;
}

.input-hint {
  display: block;
  font-size: 0.75rem;
  color: #64748b;
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

/* Form Actions */
.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
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
  flex: 1;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: #334155;
  color: #cbd5e1;
  flex: 1;
}

.btn-secondary:hover {
  background: #475569;
  transform: translateY(-1px);
}

.btn-warning {
  background: #f59e0b;
  color: #1e293b;
  flex: 1;
}

.btn-warning:hover:not(:disabled) {
  background: #d97706;
  transform: translateY(-1px);
}

.btn-warning:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #dc2626;
  transform: translateY(-1px);
}

.btn-danger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-icon {
  font-size: 1rem;
}

/* Network Info Card */
.network-info-card {
  background: #1e293b;
  border-radius: 12px;
  padding: 16px;
  border: 1px solid #334155;
}

.network-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #334155;
}

.network-name-large {
  font-size: 1.1rem;
  font-weight: 700;
  color: #60a5fa;
}

.network-status {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 600;
}

.network-status.active {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.network-status:not(.active) {
  background: rgba(107, 114, 128, 0.1);
  color: #9ca3af;
  border: 1px solid rgba(107, 114, 128, 0.3);
}

.network-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.network-detail-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 0.85rem;
}

.network-detail-item .detail-label {
  color: #94a3b8;
  min-width: 100px;
}

.network-detail-item .detail-value {
  color: #e2e8f0;
  font-family: monospace;
}

.network-actions {
  display: flex;
  gap: 12px;
  padding-top: 12px;
  border-top: 1px solid #334155;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.modal-content {
  background: #1e293b;
  border-radius: 16px;
  width: 90%;
  max-width: 450px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
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
  color: #f8fafc;
  font-size: 1.2rem;
}

.modal-close {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 24px;
  cursor: pointer;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.modal-close:hover {
  background: rgba(148, 163, 184, 0.1);
  color: #f8fafc;
}

.modal-body {
  padding: 24px;
}

.modal-body p {
  color: #e2e8f0;
  margin-bottom: 12px;
}

.warning-text {
  color: #f87171 !important;
  font-size: 0.85rem;
  margin-top: 8px;
}

.modal-footer {
  display: flex;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
}

/* Notifications */
.notification {
  position: fixed;
  top: 24px;
  right: 24px;
  padding: 14px 20px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  gap: 12px;
  animation: slideIn 0.3s ease;
  z-index: 1000;
  max-width: 420px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
}

.notification.success {
  background: rgba(16, 185, 129, 0.95);
  border: 1px solid rgba(16, 185, 129, 0.3);
  color: white;
}

.notification.error {
  background: rgba(239, 68, 68, 0.95);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: white;
}

.notification.info {
  background: rgba(59, 130, 246, 0.95);
  border: 1px solid rgba(59, 130, 246, 0.3);
  color: white;
}

.notification-icon {
  font-size: 1.2rem;
}

.notification-message {
  flex: 1;
  font-size: 0.9rem;
}

.notification-close {
  background: transparent;
  border: none;
  color: white;
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
  border-radius: 4px;
}

.notification-close:hover {
  opacity: 1;
  background: rgba(255, 255, 255, 0.1);
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

/* Empty State */
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
  
  .network-header {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }
  
  .network-detail-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  
  .network-detail-item .detail-label {
    min-width: auto;
  }
  
  .network-actions {
    flex-direction: column;
  }
}
</style>