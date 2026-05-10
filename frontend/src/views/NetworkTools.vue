<template>
  <div class="tools-layout">
    <aside class="tools-sidebar">
      <!-- <div class="sidebar-header">
        <h2>Network Tools</h2>
        <p class="subtitle">Professional diagnostic tools</p>
      </div> -->
      
      <nav class="tools-nav">
        <router-link 
          v-for="tool in tools" 
          :key="tool.id"
          :to="tool.path" 
          class="tool-link"
          active-class="active"
        >
          <span class="tool-icon" v-html="tool.icon"></span>
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
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

// SVG Icons for each tool
const tools = ref([
  { 
    id: 1, 
    name: 'Portscan', 
    path: '/tools/portscan', 
    icon: `<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="12" cy="12" r="10"/>
      <line x1="12" y1="8" x2="12" y2="12"/>
      <line x1="12" y1="16" x2="12.01" y2="16"/>
      <path d="M4 4 L20 20"/>
    </svg>`,
    beta: false 
  },
  { 
    id: 2, 
    name: 'TCP Check', 
    path: '/tools/tcp-check', 
    icon: `<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <path d="M6.5 6.5L12 12l5.5-5.5"/>
      <path d="M6.5 17.5L12 12l5.5 5.5"/>
      <circle cx="12" cy="12" r="10"/>
    </svg>`,
    beta: false 
  },
  { 
    id: 3, 
    name: 'DNS Lookup', 
    path: '/tools/dns-lookup', 
    icon: `<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="12" cy="12" r="10"/>
      <path d="M2 12h20"/>
      <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
    </svg>`,
    beta: false 
  },
  { 
    id: 4, 
    name: 'Traceroute', 
    path: '/tools/traceroute', 
    icon: `<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="5" cy="12" r="2"/>
      <circle cx="12" cy="5" r="2"/>
      <circle cx="12" cy="19" r="2"/>
      <circle cx="19" cy="12" r="2"/>
      <path d="M7 7 L10 9"/>
      <path d="M14 15 L17 17"/>
      <path d="M14 9 L17 7"/>
      <path d="M7 17 L10 15"/>
      <path d="M10 11 L14 9"/>
      <path d="M14 15 L10 13"/>
    </svg>`,
    beta: false 
  },
  { 
    id: 5, 
    name: 'Ping', 
    path: '/tools/ping', 
    icon: `<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <path d="M12 2v4"/>
      <path d="M12 18v4"/>
      <path d="M4.9 4.9l3.3 3.3"/>
      <path d="M15.8 15.8l3.3 3.3"/>
      <path d="M2 12h4"/>
      <path d="M18 12h4"/>
      <circle cx="12" cy="12" r="5"/>
      <path d="M8 8 L16 16"/>
    </svg>`,
    beta: false 
  },
  { 
    id: 6, 
    name: 'HTTP(S) Check', 
    path: '/tools/httpsCheck', 
    icon: `<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
      <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
      <circle cx="12" cy="16" r="1.5"/>
      <path d="M12 22v-4"/>
    </svg>`,
    beta: false 
  },
  { 
    id: 7, 
    name: 'Speedtest', 
    path: '/tools/speedtest', 
    icon: `<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <path d="M3 12h4"/>
      <path d="M17 12h4"/>
      <path d="M8 8L6 10"/>
      <path d="M16 8l2 2"/>
      <path d="M10 20h4"/>
      <path d="M12 4v2"/>
      <path d="M12 10v4"/>
      <path d="M12 18v2"/>
      <circle cx="12" cy="12" r="3"/>
      <path d="M15 15l3 3"/>
      <path d="M9 15l-3 3"/>
    </svg>`,
    beta: false 
  },
])

const toolDescriptions = {
  'portscan': 'Scan open ports on network devices to identify services and potential vulnerabilities',
  'tcp-check': 'Verify TCP connectivity to specific ports and measure response times',
  'dns-lookup': 'Resolve domain names to IP addresses and view DNS records',
  'traceroute': 'Trace the network path to destination and identify routing hops',
  'ping': 'Check host availability, measure latency, and analyze packet loss',
  'httpsCheck': 'Check if a website supports HTTPS and analyze its SSL/TLS configuration',
  'speedtest': 'Test network bandwidth, upload/download speeds, and connection quality',
}

const currentToolName = computed(() => {
  if (route.path === '/tools' || route.path === '/tools/') {
    return 'Portscan'
  }
  
  const tool = tools.value.find(t => route.path === t.path)
  return tool ? tool.name : 'Portscan'
})

const currentToolDescription = computed(() => {
  if (route.path === '/tools' || route.path === '/tools/') {
    return toolDescriptions['portscan']
  }
  
  const pathSegment = route.path.split('/').pop()
  return toolDescriptions[pathSegment] || toolDescriptions['portscan']
})
</script>

<style scoped>
.tools-layout {
  display: flex;
  height: 100vh;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  width: 100vw;
  overflow: hidden;
}

/* Sidebar Styles */
.tools-sidebar {
  width: 280px;
  flex-shrink: 0;
  background: rgba(15, 23, 42, 0.95);
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
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(139, 92, 246, 0.15) 100%);
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
  background: linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.05), transparent);
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
  background: linear-gradient(135deg, #60a5fa, #a78bfa);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.sidebar-header .subtitle {
  margin: 0;
  opacity: 0.7;
  font-size: 0.85rem;
  position: relative;
  color: #94a3b8;
}

/* Navigation Styles */
.tools-nav {
  padding: 20px 15px;
  flex: 1;
  overflow-y: auto;
}

.tool-link {
  display: flex;
  align-items: center;
  padding: 12px 16px;
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
  gap: 12px;
}

.tool-link:hover {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(30, 41, 59, 0.8) 100%);
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
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  width: 6px;
  height: 6px;
  background: rgba(255, 255, 255, 0.8);
  border-radius: 50%;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: translateY(-50%) scale(1); }
  50% { opacity: 0.3; transform: translateY(-50%) scale(0.8); }
}

.tool-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  filter: drop-shadow(0 2px 2px rgba(0, 0, 0, 0.2));
  flex-shrink: 0;
}

.tool-icon svg {
  width: 100%;
  height: 100%;
}

.tool-link.active .tool-icon svg {
  stroke: white;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
}

.tool-name {
  flex: 1;
  font-size: 0.95rem;
}

.beta-badge {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  font-size: 0.65rem;
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

/* Main Content Styles */
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
  font-size: 1rem;
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

/* Animation for router transitions */
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

/* Scrollbar Styling */
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

/* Responsive Design */
@media (max-width: 1024px) {
  .tools-sidebar {
    width: 250px;
  }
  
  .tools-content {
    padding: 30px;
  }
  
  .content-wrapper {
    padding: 30px;
  }
  
  .scan-header h1 {
    font-size: 2rem;
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
    font-size: 0.9rem;
  }
  
  .content-wrapper {
    padding: 15px;
  }
}
</style>