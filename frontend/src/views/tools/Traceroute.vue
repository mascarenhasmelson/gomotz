<template>
  <div class="traceroute-page">
    <div class="traceroute-container">
      <!-- Header -->
      <div class="scan-header">
        <h1>Traceroute</h1>
        <p class="subtitle">Trace the network path to any host with detailed hop-by-hop analysis</p>
      </div>

      <!-- Scan Form -->
      <div class="scan-form">
        <div class="form-group">
          <label for="target-input">Target Host:</label>
          <div class="input-with-button">
            <input
              v-model="targetHost"
              id="target-input"
              type="text"
              placeholder="Enter IP address or hostname (e.g., google.com, 8.8.8.8)"
              :disabled="isTracing"
              @keyup.enter="startTrace"
              class="target-input"
            />
            <button 
              @click="startTrace" 
              :disabled="!targetHost || isTracing" 
              class="scan-button"
            >
              <span v-if="!isTracing">Start Trace</span>
              <span v-else class="scanning-text">
                <span class="spinner"></span> Tracing...
              </span>
            </button>
            <button 
              v-if="isTracing"
              @click="stopTrace" 
              class="stop-button"
            >
              Stop
            </button>
          </div>
          
          <!-- Advanced Options -->
          <div class="advanced-options">
            <button 
              @click="showAdvanced = !showAdvanced" 
              class="advanced-toggle"
            >
              {{ showAdvanced ? 'Hide' : 'Advanced' }} Options
              <span class="toggle-icon">{{ showAdvanced ? '▲' : '▼' }}</span>
            </button>
            
            <div v-if="showAdvanced" class="options-panel">
              <div class="options-grid">
                <div class="option-group">
                  <label>Max Hops:</label>
                  <input 
                    v-model="maxHops" 
                    type="range" 
                    min="1" 
                    max="64" 
                    step="1"
                    class="slider"
                  />
                  <span class="slider-value">{{ maxHops }}</span>
                </div>
                
                <div class="option-group">
                  <label>Probes per Hop:</label>
                  <div class="probe-buttons">
                    <button 
                      v-for="n in [1, 2, 3]" 
                      :key="n"
                      @click="probesPerHop = n"
                      :class="{ active: probesPerHop === n }"
                      class="probe-btn"
                    >
                      {{ n }}
                    </button>
                  </div>
                </div>
                
                <div class="option-group">
                  <label>Timeout (ms):</label>
                  <input 
                    v-model="timeout" 
                    type="range" 
                    min="500" 
                    max="5000" 
                    step="100"
                    class="slider"
                  />
                  <span class="slider-value">{{ timeout }}ms</span>
                </div>
                
                <div class="option-group">
                  <label>Protocol:</label>
                  <div class="protocol-buttons">
                    <button 
                      @click="protocol = 'ICMP'"
                      :class="{ active: protocol === 'ICMP' }"
                      class="protocol-btn"
                    >
                      ICMP
                    </button>
                    <button 
                      @click="protocol = 'TCP'"
                      :class="{ active: protocol === 'TCP' }"
                      class="protocol-btn"
                    >
                      TCP
                    </button>
                    <button 
                      @click="protocol = 'UDP'"
                      :class="{ active: protocol === 'UDP' }"
                      class="protocol-btn"
                    >
                      UDP
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Connection Status -->
        <div v-if="connectionStatus" class="connection-status" :class="connectionStatusClass">
          {{ connectionStatus }}
        </div>
        
        <!-- Error Message -->
        <div v-if="error" class="error-message">
          ⚠️ {{ error }}
        </div>
      </div>

      <!-- Real-time Visualization -->
      <div v-if="isTracing || traceResults.length > 0" class="visualization-section">
        <div class="vis-header">
          <h3>Network Path Visualization</h3>
          <div class="vis-stats">
            <span class="stat-badge">
              <span class="stat-icon">📍</span>
              {{ traceResults.length }} Hops
            </span>
            <span class="stat-badge">
              <span class="stat-icon">⏱️</span>
              {{ elapsedTime }}s
            </span>
            <span class="stat-badge">
              <span class="stat-icon">📊</span>
              {{ completedHops }}/{{ maxHops }}
            </span>
          </div>
        </div>
        
        <!-- Network Map -->
        <div class="network-map">
          <div class="map-container">
            <!-- Source Node -->
            <div class="network-node source">
              <div class="node-icon">🏠</div>
              <div class="node-label">You</div>
              <div class="node-ip">Local Network</div>
            </div>
            
            <!-- Network Path -->
            <div class="network-path">
              <div 
                v-for="(hop, index) in traceResults" 
                :key="index"
                class="path-segment"
                :class="getHopStatusClass(hop)"
              >
                <div class="hop-line">
                  <div class="line-progress" v-if="isTracing && index === traceResults.length - 1"></div>
                </div>
                
                <div class="hop-node">
                  <div class="hop-number">Hop {{ hop.hop }}</div>
                  <div class="hop-ip" :title="hop.ip">
                    {{ hop.hostname || hop.ip }}
                  </div>
                  <div class="hop-latency">
                    <span v-for="(rtt, i) in hop.rtts" :key="i" class="latency-badge">
                      {{ rtt }}ms
                    </span>
                  </div>
                  <div class="hop-info">
                    <span class="hop-as" v-if="hop.asn">AS{{ hop.asn }}</span>
                    <span class="hop-location" v-if="hop.location">{{ hop.location }}</span>
                  </div>
                </div>
              </div>
              
              <!-- Target Node -->
              <div class="path-segment" v-if="traceCompleted || isTracing">
                <div class="hop-line">
                  <div class="line-progress" v-if="isTracing && traceResults.length === 0"></div>
                </div>
                <div class="network-node target" :class="{ reached: traceCompleted }">
                  <div class="node-icon">{{ traceCompleted ? '🎯' : '🌐' }}</div>
                  <div class="node-label">{{ targetHost }}</div>
                  <div class="node-ip">{{ targetIp || 'Resolving...' }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Real-time Progress -->
        <div v-if="isTracing" class="trace-progress">
          <div class="progress-container">
            <div class="progress-bar" :style="{ width: traceProgress + '%' }"></div>
          </div>
          <div class="progress-info">
            <span>Probing hop {{ currentHop }} of {{ maxHops }}...</span>
            <span>{{ traceProgress.toFixed(1) }}% Complete</span>
          </div>
        </div>
      </div>

      <!-- Results Table -->
      <div v-if="traceResults.length > 0" class="results-section">
        <div class="results-header">
          <h2>Traceroute Results</h2>
          <div class="results-meta">
            <span class="target-display">{{ targetHost }}</span>
            <span class="scan-time">{{ formattedTraceTime }}</span>
            <button @click="exportResults" class="export-btn">
              Export Results
            </button>
          </div>
        </div>

        <div class="results-table">
          <table>
            <thead>
              <tr>
                <th>Hop</th>
                <th>IP Address</th>
                <th>Hostname</th>
                <th>Latency (ms)</th>
                <th>AS Number</th>
                <th>Location</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              <tr 
                v-for="(hop, index) in traceResults" 
                :key="index"
                :class="getRowClass(hop)"
              >
                <td class="hop-cell">
                  <div class="hop-number">{{ hop.hop }}</div>
                  <div class="hop-visual">
                    <div class="hop-dot" :class="getHopDotClass(hop)"></div>
                    <div class="hop-line-visual"></div>
                  </div>
                </td>
                <td class="ip-cell">
                  <div class="ip-address">{{ hop.ip }}</div>
                  <div v-if="hop.reverseDns" class="reverse-dns">
                    {{ hop.reverseDns }}
                  </div>
                </td>
                <td class="hostname-cell">
                  <div class="hostname">{{ hop.hostname || '-' }}</div>
                  <div v-if="hop.isp" class="isp">{{ hop.isp }}</div>
                </td>
                <td class="latency-cell">
                  <div class="latency-bars">
                    <div 
                      v-for="(rtt, i) in hop.rtts" 
                      :key="i"
                      class="latency-bar"
                      :style="{ height: getLatencyHeight(rtt) + 'px' }"
                      :title="`Probe ${i+1}: ${rtt}ms`"
                    ></div>
                  </div>
                  <div class="latency-stats">
                    <span v-if="hop.avgRtt" class="avg-latency">
                      Avg: {{ hop.avgRtt }}ms
                    </span>
                    <span v-if="hop.packetLoss > 0" class="packet-loss">
                      Loss: {{ hop.packetLoss }}%
                    </span>
                  </div>
                </td>
                <td class="as-cell">
                  <div v-if="hop.asn" class="as-badge">
                    AS{{ hop.asn }}
                  </div>
                  <div v-else class="as-unknown">Unknown</div>
                </td>
                <td class="location-cell">
                  <div v-if="hop.location" class="location-info">
                    <span class="location-flag">{{ getFlagEmoji(hop.country) }}</span>
                    <span class="location-text">{{ hop.location }}</span>
                  </div>
                  <div v-else>-</div>
                </td>
                <td class="status-cell">
                  <span class="status-badge" :class="getStatusClass(hop)">
                    {{ getStatusText(hop) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Summary Statistics -->
        <div class="summary-section">
          <h3>Path Summary</h3>
          <div class="summary-stats">
            <div class="stat-card">
              <div class="stat-icon">📈</div>
              <div class="stat-content">
                <div class="stat-value">{{ totalHops }}</div>
                <div class="stat-label">Total Hops</div>
              </div>
            </div>
            
            <div class="stat-card">
              <div class="stat-icon">⚡</div>
              <div class="stat-content">
                <div class="stat-value">{{ avgLatency }}ms</div>
                <div class="stat-label">Avg Latency</div>
              </div>
            </div>
            
            <div class="stat-card">
              <div class="stat-icon">📉</div>
              <div class="stat-content">
                <div class="stat-value">{{ maxLatency }}ms</div>
                <div class="stat-label">Max Latency</div>
              </div>
            </div>
            
            <div class="stat-card">
              <div class="stat-icon">🎯</div>
              <div class="stat-content">
                <div class="stat-value">{{ traceCompleted ? 'Reached' : 'Failed' }}</div>
                <div class="stat-label">Destination</div>
              </div>
            </div>
            
            <div class="stat-card">
              <div class="stat-icon">⏱️</div>
              <div class="stat-content">
                <div class="stat-value">{{ totalDuration }}s</div>
                <div class="stat-label">Total Time</div>
              </div>
            </div>
          </div>
          
          <!-- Path Analysis -->
          <div class="path-analysis">
            <h4>Path Analysis</h4>
            <div class="analysis-points">
              <div 
                v-for="(analysis, index) in pathAnalysis" 
                :key="index"
                class="analysis-point"
                :class="analysis.type"
              >
                <div class="analysis-icon">{{ analysis.icon }}</div>
                <div class="analysis-text">{{ analysis.text }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Recent Traces -->
      <div v-if="recentTraces.length > 0" class="recent-traces">
        <div class="recent-header">
          <h3>Recent Traces</h3>
          <button @click="clearHistory" class="clear-history-btn">
            Clear History
          </button>
        </div>
        <div class="traces-list">
          <div 
            v-for="(trace, index) in recentTraces" 
            :key="index"
            class="trace-item"
            @click="loadTrace(trace)"
          >
            <div class="trace-info-main">
              <span class="trace-target">{{ trace.target }}</span>
              <span class="trace-hops">{{ trace.hops }} hops</span>
              <span class="trace-status" :class="trace.reached ? 'success' : 'failed'">
                {{ trace.reached ? 'Reached' : 'Failed' }}
              </span>
            </div>
            <div class="trace-info-secondary">
              <span class="trace-time">{{ trace.time }}</span>
              <span class="trace-duration">{{ trace.duration }}s</span>
              <span class="trace-avg-latency">{{ trace.avgLatency }}ms avg</span>
            </div>
            <div class="trace-path-preview">
              <div 
                v-for="n in 8" 
                :key="n"
                class="path-dot"
                :class="getPathDotClass(trace, n)"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Traceroute',
  data() {
    return {
      targetHost: 'google.com',
      targetIp: '',
      isTracing: false,
      traceResults: [],
      error: null,
      connectionStatus: '',
      connectionStatusClass: '',
      recentTraces: [],
      maxHops: 30,
      probesPerHop: 3,
      timeout: 2000,
      protocol: 'ICMP',
      showAdvanced: false,
      traceStartTime: null,
      elapsedTime: 0,
      traceInterval: null,
      traceProgress: 0,
      currentHop: 1,
      traceCompleted: false,
      totalDuration: 0,
      ws: null,
      traceStartTimestamp: 0
    }
  },
  computed: {
    formattedTraceTime() {
      if (!this.traceStartTime) return '';
      return new Date(this.traceStartTime).toLocaleString();
    },
    completedHops() {
      return this.traceResults.filter(h => h.status !== 'pending').length;
    },
    totalHops() {
      return this.traceResults.length;
    },
    avgLatency() {
      if (this.traceResults.length === 0) return 0;
      const validLatencies = this.traceResults
        .filter(h => h.avgRtt)
        .map(h => h.avgRtt);
      if (validLatencies.length === 0) return 0;
      const sum = validLatencies.reduce((a, b) => a + b, 0);
      return Math.round(sum / validLatencies.length);
    },
    maxLatency() {
      if (this.traceResults.length === 0) return 0;
      const latencies = this.traceResults
        .filter(h => h.maxRtt)
        .map(h => h.maxRtt);
      return latencies.length > 0 ? Math.max(...latencies) : 0;
    },
    pathAnalysis() {
      const analysis = [];
      
      if (this.traceResults.length === 0) return analysis;
      
      // Check for local network hops
      const localHops = this.traceResults.filter(h => 
        h.ip && (h.ip.startsWith('192.168.') || 
                h.ip.startsWith('10.') || 
                h.ip.startsWith('172.16.'))
      );
      if (localHops.length > 0) {
        analysis.push({
          icon: '🏠',
          type: 'local',
          text: `${localHops.length} local network hops`
        });
      }
      
      // Check for latency spikes
      let hasSpike = false;
      for (let i = 1; i < this.traceResults.length; i++) {
        const prev = this.traceResults[i-1].avgRtt;
        const curr = this.traceResults[i].avgRtt;
        if (prev && curr && curr > prev * 2 && curr - prev > 50) {
          hasSpike = true;
          break;
        }
      }
      if (hasSpike) {
        analysis.push({
          icon: '⚠️',
          type: 'warning',
          text: 'Potential network congestion detected'
        });
      }
      
      // Check for packet loss
      const lossyHops = this.traceResults.filter(h => h.packetLoss > 20);
      if (lossyHops.length > 0) {
        analysis.push({
          icon: '📉',
          type: 'warning',
          text: `${lossyHops.length} hops with high packet loss`
        });
      }
      
      // Destination reached
      if (this.traceCompleted) {
        analysis.push({
          icon: '✅',
          type: 'success',
          text: `Destination reached in ${this.totalHops} hops`
        });
      }
      
      return analysis;
    }
  },
  mounted() {
    const savedTraces = localStorage.getItem('tracerouteHistory');
    if (savedTraces) {
      this.recentTraces = JSON.parse(savedTraces);
    }
  },
  beforeUnmount() {
    this.cleanup();
  },
  methods: {
    getFlagEmoji(countryCode) {
      if (!countryCode) return '🌐';
      const codePoints = countryCode
        .toUpperCase()
        .split('')
        .map(char => 127397 + char.charCodeAt());
      return String.fromCodePoint(...codePoints);
    },
    
    getLatencyHeight(rtt) {
      if (!rtt) return 5;
      const maxHeight = 40;
      const maxRtt = 500;
      return Math.min((rtt / maxRtt) * maxHeight, maxHeight);
    },
    
    getHopStatusClass(hop) {
      if (hop.status === 'timeout') return 'timeout';
      if (hop.status === 'error') return 'error';
      if (hop.rtts && hop.rtts.length > 0) return 'success';
      return 'pending';
    },
    
    getHopDotClass(hop) {
      if (hop.status === 'timeout') return 'timeout';
      if (hop.status === 'error') return 'error';
      if (hop.rtts && hop.rtts.length > 0) return hop.rtts.length === this.probesPerHop ? 'complete' : 'partial';
      return 'pending';
    },
    
    getRowClass(hop) {
      if (hop.status === 'timeout') return 'timeout-row';
      if (hop.status === 'error') return 'error-row';
      if (hop.isTarget) return 'target-row';
      return '';
    },
    
    getStatusClass(hop) {
      if (hop.status === 'timeout') return 'timeout';
      if (hop.status === 'error') return 'error';
      if (hop.rtts && hop.rtts.length === this.probesPerHop) return 'complete';
      if (hop.rtts && hop.rtts.length > 0) return 'partial';
      return 'pending';
    },
    
    getStatusText(hop) {
      if (hop.status === 'timeout') return 'Timeout';
      if (hop.status === 'error') return 'Error';
      if (hop.rtts && hop.rtts.length === this.probesPerHop) return 'Complete';
      if (hop.rtts && hop.rtts.length > 0) return 'Partial';
      return 'Pending';
    },
    
    getPathDotClass(trace, index) {
      const totalDots = 8;
      const progress = (trace.hops / trace.maxHops) * totalDots;
      if (index <= progress) {
        return trace.reached ? 'reached' : 'active';
      }
      return 'inactive';
    },
    
    cleanup() {
      if (this.traceInterval) {
        clearInterval(this.traceInterval);
        this.traceInterval = null;
      }
      
      if (this.ws) {
        this.ws.close();
        this.ws = null;
      }
    },
    
    async startTrace() {
      if (!this.targetHost.trim()) {
        this.error = 'Please enter a target hostname or IP address';
        return;
      }

      // Reset state
      this.isTracing = true;
      this.error = null;
      this.traceResults = [];
      this.traceCompleted = false;
      this.traceStartTime = new Date();
      this.traceStartTimestamp = Date.now();
      this.elapsedTime = 0;
      this.traceProgress = 0;
      this.currentHop = 1;
      this.targetIp = '';
      
      this.connectionStatus = 'Starting traceroute...';
      this.connectionStatusClass = 'connecting';
      
      // Clean up any existing connections
      this.cleanup();
      
      // Connect via WebSocket
      this.connectWebSocket();
      
      // Start elapsed time counter
      this.traceInterval = setInterval(() => {
        this.elapsedTime = Math.floor((Date.now() - this.traceStartTimestamp) / 1000);
        this.updateProgress();
      }, 1000);
    },
    
    connectWebSocket() {
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${wsProtocol}//${window.location.hostname}:8082/traceroute`;
      
      this.ws = new WebSocket(wsUrl);
      
      this.ws.onopen = () => {
        this.connectionStatus = 'Connected, starting trace...';
        this.connectionStatusClass = 'connected';
        
        // Send trace request
        this.ws.send(JSON.stringify({
          action: 'startTrace',
          target: this.targetHost.trim(),
          maxHops: this.maxHops,
          probesPerHop: this.probesPerHop,
          timeout: this.timeout,
          protocol: this.protocol
        }));
      };
      
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          
          switch (data.type) {
            case 'hopResult':
              this.handleHopResult(data.hop);
              break;
              
            case 'targetIp':
              this.targetIp = data.ip;
              break;
              
            case 'progress':
              this.currentHop = data.currentHop;
              break;
              
            case 'complete':
              this.handleTraceComplete(data);
              break;
              
            case 'error':
              this.error = data.message;
              this.connectionStatus = 'Error: ' + data.message;
              this.connectionStatusClass = 'error';
              this.isTracing = false;
              break;
              
            case 'status':
              this.connectionStatus = data.message;
              break;
          }
        } catch (err) {
          console.error('Error parsing WebSocket message:', err);
        }
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket Error:', error);
        this.error = 'Failed to connect to traceroute server';
        this.connectionStatus = 'Connection failed';
        this.connectionStatusClass = 'error';
        this.isTracing = false;
      };
      
      this.ws.onclose = () => {
        this.connectionStatus = 'Connection closed';
        this.connectionStatusClass = 'disconnected';
        this.ws = null;
      };
    },
    
    handleHopResult(hopData) {
      // Check if hop already exists
      const existingIndex = this.traceResults.findIndex(h => h.hop === hopData.hop);
      
      if (existingIndex !== -1) {
        // Update existing hop
        const existing = this.traceResults[existingIndex];
        
        // Add new RTT if not already present
        if (hopData.rtt && !existing.rtts.includes(hopData.rtt)) {
          existing.rtts.push(hopData.rtt);
          existing.avgRtt = Math.round(existing.rtts.reduce((a, b) => a + b, 0) / existing.rtts.length);
          existing.maxRtt = Math.max(...existing.rtts);
        }
        
        // Update other fields
        if (hopData.ip && !existing.ip) existing.ip = hopData.ip;
        if (hopData.hostname && !existing.hostname) existing.hostname = hopData.hostname;
        if (hopData.status) existing.status = hopData.status;
        
        // Update packet loss calculation
        if (this.probesPerHop > 0) {
          existing.packetLoss = Math.round(((this.probesPerHop - existing.rtts.length) / this.probesPerHop) * 100);
        }
      } else {
        // Add new hop
        const hop = {
          hop: hopData.hop,
          ip: hopData.ip || '',
          hostname: hopData.hostname || '',
          rtts: hopData.rtt ? [hopData.rtt] : [],
          avgRtt: hopData.rtt || 0,
          maxRtt: hopData.rtt || 0,
          status: hopData.status || (hopData.rtt ? 'success' : 'timeout'),
          packetLoss: hopData.rtt ? 0 : 100,
          asn: hopData.asn,
          country: hopData.country,
          location: hopData.location,
          isp: hopData.isp,
          reverseDns: hopData.reverseDns
        };
        
        this.traceResults.push(hop);
        this.traceResults.sort((a, b) => a.hop - b.hop);
      }
      
      // Check if this is the target
      if (hopData.isTarget) {
        this.traceCompleted = true;
      }
      
      this.$forceUpdate();
    },
    
    updateProgress() {
      if (this.maxHops > 0) {
        this.traceProgress = (this.currentHop / this.maxHops) * 100;
      }
    },
    
    handleTraceComplete(data) {
      this.totalDuration = ((Date.now() - this.traceStartTimestamp) / 1000).toFixed(2);
      this.traceProgress = 100;
      this.connectionStatus = 'Trace completed';
      this.connectionStatusClass = 'completed';
      
      // Add any final hops from complete data
      if (data.hops && Array.isArray(data.hops)) {
        data.hops.forEach(hop => {
          this.handleHopResult(hop);
        });
      }
      
      if (data.targetIp) {
        this.targetIp = data.targetIp;
      }
      
      this.isTracing = false;
      this.traceCompleted = data.reached || false;
      
      // Save to history
      this.saveToHistory(data);
      
      // Cleanup
      this.cleanup();
    },
    
    stopTrace() {
      if (confirm('Stop the current traceroute?')) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
          this.ws.send(JSON.stringify({ action: 'stopTrace' }));
        }
        
        this.isTracing = false;
        this.connectionStatus = 'Trace stopped by user';
        this.connectionStatusClass = 'stopped';
        
        // Save partial results
        if (this.traceResults.length > 0) {
          this.saveToHistory({
            hops: this.traceResults,
            reached: this.traceCompleted,
            targetIp: this.targetIp
          });
        }
        
        this.cleanup();
      }
    },
    
    saveToHistory(data) {
      const traceRecord = {
        target: this.targetHost.trim(),
        targetIp: this.targetIp,
        hops: this.traceResults.length,
        maxHops: this.maxHops,
        reached: this.traceCompleted,
        avgLatency: this.avgLatency,
        maxLatency: this.maxLatency,
        time: new Date().toLocaleString(),
        duration: this.totalDuration,
        timestamp: Date.now(),
        protocol: this.protocol,
        results: this.traceResults.slice()
      };

      this.recentTraces.unshift(traceRecord);
      if (this.recentTraces.length > 10) {
        this.recentTraces = this.recentTraces.slice(0, 10);
      }

      localStorage.setItem('tracerouteHistory', JSON.stringify(this.recentTraces));
    },
    
    exportResults() {
      const data = {
        target: this.targetHost.trim(),
        targetIp: this.targetIp,
        timestamp: this.formattedTraceTime,
        duration: this.totalDuration,
        protocol: this.protocol,
        maxHops: this.maxHops,
        probesPerHop: this.probesPerHop,
        reached: this.traceCompleted,
        hops: this.traceResults,
        summary: {
          totalHops: this.totalHops,
          avgLatency: this.avgLatency,
          maxLatency: this.maxLatency
        }
      };
      
      const content = JSON.stringify(data, null, 2);
      const blob = new Blob([content], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `traceroute-${this.targetHost.replace(/[^a-z0-9]/gi, '-')}-${Date.now()}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    },
    
    loadTrace(trace) {
      this.targetHost = trace.target;
      this.targetIp = trace.targetIp;
      this.traceResults = trace.results || [];
      this.traceCompleted = trace.reached;
      this.totalDuration = trace.duration;
      window.scrollTo({ top: 0, behavior: 'smooth' });
    },
    
    clearHistory() {
      if (confirm('Clear all traceroute history?')) {
        this.recentTraces = [];
        localStorage.removeItem('tracerouteHistory');
      }
    }
  }
}
</script>

<style scoped>
.traceroute-page {
  padding: 20px;
  width: 100%;
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.traceroute-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
}

/* Header */
.scan-header {
  text-align: center;
  margin-bottom: 40px;
  padding: 30px;
  background: white;
  border-radius: 15px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
}

.scan-header h1 {
  font-size: 2.5rem;
  color: #2d3748;
  margin: 0 0 10px 0;
  font-weight: 700;
  background: linear-gradient(135deg, #4299e1 0%, #3182ce 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.subtitle {
  color: #718096;
  font-size: 1.1rem;
  margin: 0;
}

/* Scan Form */
.scan-form {
  background: white;
  border-radius: 15px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  margin-bottom: 30px;
}

.form-group label {
  display: block;
  font-weight: 600;
  color: #4a5568;
  margin-bottom: 12px;
  font-size: 1.1rem;
}

.input-with-button {
  display: flex;
  gap: 15px;
  align-items: center;
  margin-bottom: 20px;
}

.target-input {
  flex: 1;
  padding: 16px 20px;
  border: 2px solid #e2e8f0;
  border-radius: 10px;
  font-size: 1rem;
  font-family: 'Monaco', 'Courier New', monospace;
  transition: all 0.3s;
  background: #f8fafc;
}

.target-input:focus {
  outline: none;
  border-color: #4299e1;
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
  background: white;
}

.scan-button {
  background: linear-gradient(135deg, #4299e1 0%, #3182ce 100%);
  color: white;
  border: none;
  padding: 16px 40px;
  border-radius: 10px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  min-width: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
}

.scan-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(66, 153, 225, 0.3);
}

.scan-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.stop-button {
  background: linear-gradient(135deg, #fc8181 0%, #e53e3e 100%);
  color: white;
  border: none;
  padding: 16px 25px;
  border-radius: 10px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  min-width: 100px;
  white-space: nowrap;
}

.stop-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(229, 62, 62, 0.3);
}

.scanning-text {
  display: flex;
  align-items: center;
  gap: 10px;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid white;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Advanced Options */
.advanced-options {
  margin: 20px 0;
}

.advanced-toggle {
  background: #f8fafc;
  border: 2px solid #e2e8f0;
  color: #4a5568;
  padding: 10px 20px;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: all 0.3s;
}

.advanced-toggle:hover {
  background: #edf2f7;
  border-color: #cbd5e0;
}

.toggle-icon {
  font-size: 0.8rem;
}

.options-panel {
  margin-top: 20px;
  padding: 25px;
  background: #f8fafc;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.options-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 25px;
}

.option-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.option-group label {
  font-weight: 500;
  color: #4a5568;
  font-size: 0.9rem;
}

.slider {
  width: 100%;
  height: 6px;
  background: #e2e8f0;
  border-radius: 3px;
  outline: none;
  -webkit-appearance: none;
}

.slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 20px;
  height: 20px;
  background: #4299e1;
  border-radius: 50%;
  cursor: pointer;
  border: 3px solid white;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
}

.slider-value {
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  color: #718096;
  text-align: right;
}

.probe-buttons, .protocol-buttons {
  display: flex;
  gap: 10px;
}

.probe-btn, .protocol-btn {
  flex: 1;
  padding: 10px;
  background: white;
  border: 2px solid #e2e8f0;
  border-radius: 6px;
  color: #4a5568;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.probe-btn.active, .protocol-btn.active {
  background: #4299e1;
  color: white;
  border-color: #3182ce;
}

.probe-btn:hover:not(.active), .protocol-btn:hover:not(.active) {
  border-color: #cbd5e0;
}

/* Connection Status */
.connection-status {
  margin-top: 15px;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  animation: fadeIn 0.3s ease-out;
}

.connection-status.connecting {
  background: #feebc8;
  color: #744210;
  border-left: 4px solid #ed8936;
}

.connection-status.connected {
  background: #c6f6d5;
  color: #22543d;
  border-left: 4px solid #38a169;
}

.connection-status.error {
  background: #fed7d7;
  color: #c53030;
  border-left: 4px solid #e53e3e;
}

.connection-status.disconnected {
  background: #e2e8f0;
  color: #4a5568;
  border-left: 4px solid #a0aec0;
}

.connection-status.completed {
  background: #bee3f8;
  color: #2c5282;
  border-left: 4px solid #3182ce;
}

.connection-status.stopped {
  background: #fed7e2;
  color: #97266d;
  border-left: 4px solid #d53f8c;
}

/* Error Message */
.error-message {
  margin-top: 15px;
  padding: 12px 16px;
  background: #fed7d7;
  color: #c53030;
  border-radius: 8px;
  border-left: 4px solid #e53e3e;
  font-size: 0.9rem;
  animation: shake 0.5s ease-out;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-5px); }
  75% { transform: translateX(5px); }
}

/* Visualization Section */
.visualization-section {
  background: white;
  border-radius: 15px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  margin-bottom: 30px;
  animation: slideInUp 0.5s ease-out;
}

@keyframes slideInUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

.vis-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 25px;
  padding-bottom: 20px;
  border-bottom: 2px solid #e2e8f0;
}

.vis-header h3 {
  margin: 0;
  color: #4a5568;
  font-size: 1.5rem;
}

.vis-stats {
  display: flex;
  gap: 15px;
}

.stat-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: #f8fafc;
  border-radius: 20px;
  font-size: 0.9rem;
  font-weight: 500;
  color: #4a5568;
  border: 1px solid #e2e8f0;
}

.stat-icon {
  font-size: 1rem;
}

/* Network Map */
.network-map {
  margin: 30px 0;
  padding: 25px;
  background: #f8fafc;
  border-radius: 12px;
  border: 2px solid #e2e8f0;
}

.map-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 40px;
}

.network-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 20px;
  border-radius: 12px;
  min-width: 150px;
  text-align: center;
  position: relative;
}

.network-node.source {
  background: linear-gradient(135deg, #9ae6b4 0%, #38a169 100%);
  color: white;
  box-shadow: 0 8px 20px rgba(56, 161, 105, 0.2);
}

.network-node.target {
  background: linear-gradient(135deg, #e2e8f0 0%, #cbd5e0 100%);
  color: #4a5568;
  border: 2px dashed #a0aec0;
}

.network-node.target.reached {
  background: linear-gradient(135deg, #feb2b2 0%, #e53e3e 100%);
  color: white;
  border: 2px solid #e53e3e;
  box-shadow: 0 8px 20px rgba(229, 62, 62, 0.2);
}

.node-icon {
  font-size: 2.5rem;
  margin-bottom: 5px;
}

.node-label {
  font-weight: 700;
  font-size: 1.2rem;
}

.node-ip {
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  opacity: 0.9;
}

.network-path {
  display: flex;
  flex-direction: column;
  gap: 30px;
  width: 100%;
  align-items: center;
}

.path-segment {
  display: flex;
  align-items: center;
  gap: 30px;
  width: 100%;
  max-width: 800px;
}

.hop-line {
  flex: 1;
  height: 4px;
  background: #e2e8f0;
  border-radius: 2px;
  position: relative;
  overflow: hidden;
}

.hop-line .line-progress {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  background: linear-gradient(90deg, #4299e1, #3182ce);
  border-radius: 2px;
  animation: progressLine 2s infinite linear;
}

@keyframes progressLine {
  0% { width: 0%; left: 0; }
  50% { width: 100%; left: 0; }
  100% { width: 0%; left: 100%; }
}

.hop-node {
  padding: 20px;
  background: white;
  border-radius: 10px;
  border: 2px solid #e2e8f0;
  min-width: 200px;
  text-align: center;
  transition: all 0.3s;
  position: relative;
}

.hop-node:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  border-color: #4299e1;
}

.path-segment.success .hop-node {
  border-color: #38a169;
}

.path-segment.timeout .hop-node {
  border-color: #ed8936;
}

.path-segment.error .hop-node {
  border-color: #e53e3e;
}

.hop-number {
  font-weight: 700;
  color: #4299e1;
  font-size: 1.1rem;
  margin-bottom: 5px;
}

.hop-ip {
  font-family: 'Monaco', 'Courier New', monospace;
  font-weight: 600;
  color: #4a5568;
  font-size: 1rem;
  margin-bottom: 10px;
  word-break: break-all;
}

.hop-latency {
  display: flex;
  gap: 5px;
  justify-content: center;
  margin-bottom: 10px;
}

.latency-badge {
  padding: 4px 8px;
  background: #bee3f8;
  color: #2c5282;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 600;
  font-family: 'Monaco', 'Courier New', monospace;
}

.hop-info {
  display: flex;
  gap: 10px;
  justify-content: center;
  font-size: 0.8rem;
  color: #718096;
}

.hop-as {
  background: #e9d8fd;
  color: #553c9a;
  padding: 2px 6px;
  border-radius: 3px;
}

.hop-location {
  background: #fed7e2;
  color: #97266d;
  padding: 2px 6px;
  border-radius: 3px;
}

/* Trace Progress */
.trace-progress {
  margin-top: 30px;
  padding: 20px;
  background: white;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
}

.progress-container {
  height: 10px;
  background: #e2e8f0;
  border-radius: 5px;
  overflow: hidden;
  margin-bottom: 15px;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #4299e1, #3182ce);
  border-radius: 5px;
  transition: width 0.5s ease;
  position: relative;
}

.progress-bar::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.4), transparent);
  animation: shimmer 2s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.progress-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.9rem;
  color: #718096;
}

/* Results Table */
.results-section {
  background: white;
  border-radius: 15px;
  padding: 30px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
  margin-bottom: 30px;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 25px;
  padding-bottom: 20px;
  border-bottom: 2px solid #e2e8f0;
}

.results-header h2 {
  margin: 0;
  color: #2d3748;
  font-size: 1.8rem;
}

.results-meta {
  display: flex;
  align-items: center;
  gap: 20px;
}

.target-display {
  font-family: 'Monaco', 'Courier New', monospace;
  font-weight: 600;
  color: #4a5568;
  font-size: 1.1rem;
}

.scan-time {
  color: #718096;
  font-size: 0.9rem;
}

.export-btn {
  padding: 10px 20px;
  background: #4299e1;
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.export-btn:hover {
  background: #3182ce;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(49, 130, 206, 0.2);
}

.results-table {
  overflow-x: auto;
  margin: 20px 0;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
}

table {
  width: 100%;
  border-collapse: collapse;
  min-width: 1000px;
}

thead {
  background: #f8fafc;
  border-bottom: 2px solid #e2e8f0;
}

th {
  padding: 16px 20px;
  text-align: left;
  font-weight: 600;
  color: #4a5568;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

tbody tr {
  border-bottom: 1px solid #e2e8f0;
  transition: all 0.3s;
}

tbody tr:hover {
  background: #f8fafc;
}

tbody tr.timeout-row {
  background: #fffaf0;
}

tbody tr.error-row {
  background: #fff5f5;
}

tbody tr.target-row {
  background: #f0fff4;
}

td {
  padding: 16px 20px;
  vertical-align: middle;
}

/* Table Cell Styles */
.hop-cell {
  display: flex;
  align-items: center;
  gap: 15px;
}

.hop-visual {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
}

.hop-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid #e2e8f0;
}

.hop-dot.pending {
  background: #e2e8f0;
  animation: pulse 2s infinite;
}

.hop-dot.partial {
  background: #feebc8;
  border-color: #ed8936;
}

.hop-dot.complete {
  background: #9ae6b4;
  border-color: #38a169;
}

.hop-dot.timeout {
  background: #fed7d7;
  border-color: #e53e3e;
}

.hop-dot.error {
  background: #fed7e2;
  border-color: #d53f8c;
}

.hop-line-visual {
  width: 2px;
  height: 20px;
  background: #e2e8f0;
}

.ip-cell, .hostname-cell {
  font-family: 'Monaco', 'Courier New', monospace;
}

.reverse-dns, .isp {
  font-size: 0.8rem;
  color: #718096;
  margin-top: 5px;
}

.latency-cell {
  min-width: 120px;
}

.latency-bars {
  display: flex;
  align-items: flex-end;
  gap: 4px;
  height: 40px;
  margin-bottom: 10px;
}

.latency-bar {
  width: 8px;
  background: #4299e1;
  border-radius: 2px;
  transition: height 0.3s ease;
}

.latency-stats {
  display: flex;
  gap: 10px;
  font-size: 0.8rem;
}

.avg-latency {
  color: #38a169;
  font-weight: 500;
}

.packet-loss {
  color: #e53e3e;
  font-weight: 500;
}

.as-badge {
  padding: 6px 12px;
  background: #e9d8fd;
  color: #553c9a;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
  text-align: center;
}

.as-unknown {
  color: #a0aec0;
  font-style: italic;
  font-size: 0.9rem;
}

.location-cell {
  min-width: 150px;
}

.location-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.location-flag {
  font-size: 1.2rem;
}

.location-text {
  font-size: 0.9rem;
  color: #4a5568;
}

.status-cell {
  min-width: 100px;
}

.status-badge {
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-badge.complete {
  background: #c6f6d5;
  color: #22543d;
}

.status-badge.partial {
  background: #feebc8;
  color: #744210;
}

.status-badge.pending {
  background: #e2e8f0;
  color: #4a5568;
}

.status-badge.timeout {
  background: #fed7d7;
  color: #c53030;
}

.status-badge.error {
  background: #fed7e2;
  color: #97266d;
}

/* Summary Section */
.summary-section {
  margin-top: 40px;
  padding: 30px;
  background: #f8fafc;
  border-radius: 12px;
  border: 1px solid #e2e8f0;
}

.summary-section h3 {
  margin: 0 0 25px 0;
  color: #4a5568;
  font-size: 1.5rem;
}

.summary-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 20px;
  background: white;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  font-size: 2rem;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 1.8rem;
  font-weight: 700;
  color: #2d3748;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 0.9rem;
  color: #718096;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.path-analysis {
  margin-top: 30px;
}

.path-analysis h4 {
  margin: 0 0 15px 0;
  color: #4a5568;
  font-size: 1.2rem;
}

.analysis-points {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
}

.analysis-point {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 15px;
  background: white;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  transition: all 0.3s;
}

.analysis-point:hover {
  transform: translateY(-3px);
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.1);
}

.analysis-point.local {
  border-left: 4px solid #38a169;
}

.analysis-point.warning {
  border-left: 4px solid #ed8936;
}

.analysis-point.success {
  border-left: 4px solid #4299e1;
}

.analysis-icon {
  font-size: 1.5rem;
}

.analysis-text {
  font-size: 0.9rem;
  color: #4a5568;
  font-weight: 500;
}

/* Recent Traces */
.recent-traces {
  background: white;
  border-radius: 15px;
  padding: 25px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.08);
}

.recent-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.recent-header h3 {
  margin: 0;
  color: #4a5568;
  font-size: 1.3rem;
}

.clear-history-btn {
  padding: 8px 16px;
  background: #fed7d7;
  color: #c53030;
  border: 1px solid #feb2b2;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.clear-history-btn:hover {
  background: #feb2b2;
}

.traces-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.trace-item {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 18px 20px;
  background: #f8fafc;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  cursor: pointer;
  transition: all 0.3s;
}

.trace-item:hover {
  background: #edf2f7;
  border-color: #cbd5e0;
  transform: translateX(5px);
}

.trace-info-main {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.trace-target {
  font-weight: 600;
  color: #4a5568;
  font-family: 'Monaco', 'Courier New', monospace;
}

.trace-hops {
  color: #718096;
  font-size: 0.9rem;
}

.trace-status {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 600;
}

.trace-status.success {
  background: #c6f6d5;
  color: #22543d;
}

.trace-status.failed {
  background: #fed7d7;
  color: #c53030;
}

.trace-info-secondary {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
  font-size: 0.85rem;
  color: #a0aec0;
}

.trace-path-preview {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
}

.path-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  transition: all 0.3s;
}

.path-dot.inactive {
  background: #e2e8f0;
}

.path-dot.active {
  background: #4299e1;
  animation: bounce 1s infinite;
}

.path-dot.reached {
  background: #38a169;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-5px); }
}

/* Responsive Design */
@media (max-width: 1200px) {
  .traceroute-container {
    padding: 10px;
  }
  
  .scan-header {
    padding: 20px;
  }
  
  .scan-header h1 {
    font-size: 2rem;
  }
  
  .input-with-button {
    flex-direction: column;
  }
  
  .target-input,
  .scan-button,
  .stop-button {
    width: 100%;
  }
  
  .path-segment {
    flex-direction: column;
    gap: 20px;
  }
  
  .hop-line {
    width: 4px;
    height: 30px;
    flex: none;
  }
  
  .summary-stats {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .vis-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }
  
  .vis-stats {
    width: 100%;
    flex-wrap: wrap;
  }
  
  .results-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }
  
  .results-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .options-grid {
    grid-template-columns: 1fr;
  }
  
  .trace-info-main,
  .trace-info-secondary {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .summary-stats {
    grid-template-columns: 1fr;
  }
  
  .analysis-points {
    flex-direction: column;
  }
}

@media (max-width: 480px) {
  .hop-node {
    min-width: auto;
    width: 100%;
  }
  
  .trace-progress {
    padding: 15px;
  }
  
  .stat-card {
    flex-direction: column;
    text-align: center;
    gap: 10px;
  }
}
</style>