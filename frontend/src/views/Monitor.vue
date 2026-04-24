<template>
  <div class="tools-layout">
    <aside class="tools-sidebar">
      
      
      <nav class="tools-nav">
        <router-link 
          v-for="tool in tools" 
          :key="tool.id"
          :to="tool.path" 
          class="tool-link"
          active-class="active"
        >
          <span class="tool-icon">{{ tool.icon }}</span>
          <span class="tool-name">{{ tool.name }}</span>
          <span v-if="tool.beta" class="beta-badge">BETA</span>
        </router-link>
      </nav>
    </aside>
    
    <main class="tools-content">
      <div class="scan-header">
        <h1>{{ currentToolName }}</h1>
        <p class="tool-description">{{ currentToolDescription }}</p>
      </div>
      <div class="content-wrapper">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const tools = ref([
  { id: 1, name: 'Device Monitoring', path: '/monitor/discovery', icon: '', beta: false },
  { id: 2, name: 'TCP Port Monitoring', path: '/monitor/TCPMonitoring', icon: '  ', beta: false },
  { id: 4, name: 'SNMP Monitor', path: '/monitor/SNMPMonitor', icon: '', beta: false },
  { id: 5, name: 'Ping Monitor', path: '/monitor/PingMonitor', icon: '   ', beta: false },
  // { id: 6, name: 'HTTP(s) Keyword', path: '/monitor/HTTPsKeyword', icon: '', beta: false },
  { id: 9, name: 'HTTP(s) Monitor', path: '/monitor/HTTPsMonitoring', icon: '', beta: false },
  { id: 10, name: 'LanWakeup', path: '/monitor/LANWakeup', icon: '  ', beta: false },
  { id: 11, name: 'Domain Expiry', path: '/monitor/domain-expiry', icon: '', beta: false },
])

const toolDescriptions = {
  'discovery': 'Scan network devices, discover IPs, MAC addresses and open ports',
  'TCPMonitoring': 'Check TCP port connectivity and response time',
  'SNMPMonitor': 'Monitor network devices via SNMP v1/v2 with custom OIDs',
  'PingMonitor': 'Monitor host availability, latency, and packet loss',
  'HTTPsKeyword': 'Monitor websites and APIs for specific keywords in responses',
  'HTTPsMonitoring': 'Monitor website HTTP/HTTPS availability and response time',
  'LANWakeup': 'Wake up devices on local network using Wake-on-LAN',
  'domain-expiry': 'Monitor domain expiration dates with WHOIS queries and alerts',
}


const currentToolName = computed(() => {
  if (route.path === '/monitor' || route.path === '/monitor/') {
    return 'Device Monitoring'
  }
  
  const tool = tools.value.find(t => route.path === t.path)
  return tool ? tool.name : 'Device Monitoring'
})

const currentToolDescription = computed(() => {

  if (route.path === '/monitor' || route.path === '/monitor/') {
    return toolDescriptions['discovery']
  }
  
  const pathSegment = route.path.split('/').pop()
  return toolDescriptions[pathSegment] || toolDescriptions['discovery']
})
</script>

<style scoped>
/* Keep all your existing styles exactly as they are */
.tools-layout {
  display: flex;
  height: 100vh;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  width: 100vw;
  overflow: hidden;
}

.tools-sidebar {
  width: 250px;
  flex-shrink: 0;
  background: rgba(15, 23, 42, 0.9);
  box-shadow: 4px 0 15px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgba(30, 41, 59, 0.5);
  z-index: 10;
  backdrop-filter: blur(10px);
}

.sidebar-header {
  padding: 30px 25px;
  border-bottom: 1px solid rgba(30, 41, 59, 0.5);
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.8) 0%, rgba(139, 92, 246, 0.8) 100%);
  color: white;
  position: relative;
  overflow: hidden;
}

.sidebar-header::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.1), transparent);
  animation: shimmer 3s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.sidebar-header h2 {
  margin: 0 0 8px 0;
  font-size: 1.5rem;
  font-weight: 700;
  position: relative;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.sidebar-header .subtitle {
  margin: 0;
  opacity: 0.9;
  font-size: 0.9rem;
  position: relative;
  color: rgba(255, 255, 255, 0.8);
}

.tools-nav {
  padding: 20px 15px;
  flex: 1;
  overflow-y: auto;
}

.tool-link {
  display: flex;
  align-items: center;
  padding: 14px 20px;
  margin-bottom: 8px;
  border-radius: 12px;
  text-decoration: none;
  color: #cbd5e1;
  font-weight: 500;
  transition: all 0.3s ease;
  position: relative;
  border: 2px solid transparent;
  background: rgba(30, 41, 59, 0.3);
  backdrop-filter: blur(5px);
}

.tool-link:hover {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1) 0%, rgba(30, 41, 59, 0.8) 100%);
  color: #60a5fa;
  transform: translateX(5px);
  border-color: rgba(96, 165, 250, 0.2);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.tool-link.active {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.9) 0%, rgba(37, 99, 235, 0.9) 100%);
  color: white;
  box-shadow: 0 4px 20px rgba(59, 130, 246, 0.4);
  border-color: rgba(96, 165, 250, 0.3);
}

.tool-link.active::after {
  content: '';
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  width: 6px;
  height: 6px;
  background: rgba(255, 255, 255, 0.8);
  border-radius: 50%;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.tool-icon {
  font-size: 1.2rem;
  margin-right: 15px;
  width: 24px;
  text-align: center;
  filter: drop-shadow(0 2px 2px rgba(0, 0, 0, 0.2));
}

.tool-link.active .tool-icon {
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
}

.tool-name {
  flex: 1;
  font-size: 0.95rem;
}

.beta-badge {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  font-size: 0.7rem;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 600;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.tool-link.active .beta-badge {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border-color: rgba(255, 255, 255, 0.3);
}

.tools-content {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  padding: 40px;
  width: 100%;
}

.scan-header {
  text-align: center;
  margin-bottom: 40px;
}

.scan-header h1 {
  font-size: 2.5rem;
  margin: 0 0 10px 0;
  font-weight: 700;
  background: linear-gradient(135deg, #f472b6 0%, #c084fc 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  text-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
  position: relative;
  display: inline-block;
}

.scan-header h1::after {
  content: '';
  position: absolute;
  bottom: -5px;
  left: 25%;
  width: 50%;
  height: 3px;
  background: linear-gradient(90deg, transparent, #f472b6, transparent);
  border-radius: 2px;
}

.tool-description {
  color: #94a3b8;
  font-size: 1.1rem;
  margin: 0;
  max-width: 600px;
  margin: 0 auto;
  text-align: center;
  opacity: 0.9;
}

.content-wrapper {
  background: rgba(30, 41, 59, 0.8);
  border-radius: 20px;
  padding: 40px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
  min-height: 500px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
  backdrop-filter: blur(10px);
  position: relative;
  overflow: hidden;
}

.content-wrapper::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(244, 114, 182, 0.3), transparent);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

.tools-nav::-webkit-scrollbar {
  width: 6px;
}

.tools-nav::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.3);
  border-radius: 10px;
}

.tools-nav::-webkit-scrollbar-thumb {
  background: rgba(59, 130, 246, 0.5);
  border-radius: 10px;
}

.tools-nav::-webkit-scrollbar-thumb:hover {
  background: rgba(59, 130, 246, 0.7);
}

.tools-content::-webkit-scrollbar {
  width: 10px;
}

.tools-content::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.3);
  border-radius: 10px;
}

.tools-content::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.3);
  border-radius: 10px;
  border: 2px solid rgba(15, 23, 42, 0.3);
}

.tools-content::-webkit-scrollbar-thumb:hover {
  background: rgba(148, 163, 184, 0.5);
}

@media (max-width: 1024px) {
  .tools-sidebar {
    width: 220px;
  }
  .tools-content {
    padding: 30px;
  }
  .content-wrapper {
    padding: 30px;
  }
  .scan-header h1 {
    font-size: 2.2rem;
  }
}

@media (max-width: 768px) {
  .tools-layout {
    flex-direction: column;
    height: auto;
    min-height: 100vh;
  }
  .tools-sidebar {
    width: 100%;
    height: auto;
    border-right: none;
    border-bottom: 1px solid rgba(30, 41, 59, 0.5);
  }
  .sidebar-header {
    padding: 20px;
  }
  .tools-nav {
    display: flex;
    overflow-x: auto;
    overflow-y: hidden;
    padding: 15px;
    gap: 10px;
  }
  .tool-link {
    flex-shrink: 0;
    margin-bottom: 0;
    white-space: nowrap;
  }
  .tools-content {
    padding: 20px;
  }
  .content-wrapper {
    padding: 20px;
    min-height: 400px;
  }
  .scan-header h1 {
    font-size: 1.8rem;
  }
}

@media (max-width: 480px) {
  .sidebar-header h2 {
    font-size: 1.3rem;
  }
  .scan-header h1 {
    font-size: 1.5rem;
  }
  .tool-description {
    font-size: 1rem;
  }
  .content-wrapper {
    padding: 15px;
  }
}
</style>