<template>
  <div class="dns-lookup-page">
    <!-- Main Container -->
    <div class="dns-container">
      <!-- Input Form -->
      <div class="input-card">
        <form @submit.prevent="handleSubmit">
          <div class="input-group">
            <div class="relative">
              <input
                v-model="domain"
                type="text"
                id="domain"
                placeholder="yourdomain.com"
                class="domain-input"
                :disabled="isLoading"
              />
              <label for="domain" class="input-label">
                Name
              </label>
            </div>
          </div>
          
          <!-- Record Type Selection -->
          <div class="type-selection">
            <div class="radio-group">
              <!-- A -->
              <input v-model="recordType" value="A" id="A" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="A" class="record-btn">A</label>

              <!-- AAAA -->
              <input v-model="recordType" value="AAAA" id="AAAA" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="AAAA" class="record-btn">AAAA</label>

              <!-- AFSDB -->
              <input v-model="recordType" value="AFSDB" id="AFSDB" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="AFSDB" class="record-btn">AFSDB</label>

              <!-- ATMA -->
              <input v-model="recordType" value="ATMA" id="ATMA" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="ATMA" class="record-btn">ATMA</label>

              <!-- CAA -->
              <input v-model="recordType" value="CAA" id="CAA" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="CAA" class="record-btn">CAA</label>

              <!-- CERT -->
              <input v-model="recordType" value="CERT" id="CERT" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="CERT" class="record-btn">CERT</label>

              <!-- CNAME -->
              <input v-model="recordType" value="CNAME" id="CNAME" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="CNAME" class="record-btn">CNAME</label>

              <!-- DHCID -->
              <input v-model="recordType" value="DHCID" id="DHCID" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="DHCID" class="record-btn">DHCID</label>

              <!-- DNAME -->
              <input v-model="recordType" value="DNAME" id="DNAME" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="DNAME" class="record-btn">DNAME</label>

              <!-- DNSKEY -->
              <input v-model="recordType" value="DNSKEY" id="DNSKEY" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="DNSKEY" class="record-btn">DNSKEY</label>

              <!-- DS -->
              <input v-model="recordType" value="DS" id="DS" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="DS" class="record-btn">DS</label>

              <!-- HINFO -->
              <input v-model="recordType" value="HINFO" id="HINFO" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="HINFO" class="record-btn">HINFO</label>

              <!-- ISDN -->
              <input v-model="recordType" value="ISDN" id="ISDN" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="ISDN" class="record-btn">ISDN</label>

              <!-- LOC -->
              <input v-model="recordType" value="LOC" id="LOC" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="LOC" class="record-btn">LOC</label>

              <!-- MB -->
              <input v-model="recordType" value="MB" id="MB" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="MB" class="record-btn">MB</label>

              <!-- MG -->
              <input v-model="recordType" value="MG" id="MG" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="MG" class="record-btn">MG</label>

              <!-- MINFO -->
              <input v-model="recordType" value="MINFO" id="MINFO" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="MINFO" class="record-btn">MINFO</label>

              <!-- MR -->
              <input v-model="recordType" value="MR" id="MR" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="MR" class="record-btn">MR</label>

              <!-- MX -->
              <input v-model="recordType" value="MX" id="MX" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="MX" class="record-btn">MX</label>

              <!-- NAPTR -->
              <input v-model="recordType" value="NAPTR" id="NAPTR" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="NAPTR" class="record-btn">NAPTR</label>

              <!-- NS -->
              <input v-model="recordType" value="NS" id="NS" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="NS" class="record-btn">NS</label>

              <!-- NSAP-PTR -->
              <input v-model="recordType" value="NSAP-PTR" id="NSAP-PTR" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="NSAP-PTR" class="record-btn">NSAP-PTR</label>

              <!-- NSEC -->
              <input v-model="recordType" value="NSEC" id="NSEC" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="NSEC" class="record-btn">NSEC</label>

              <!-- NSEC3 -->
              <input v-model="recordType" value="NSEC3" id="NSEC3" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="NSEC3" class="record-btn">NSEC3</label>

              <!-- PTR -->
              <input v-model="recordType" value="PTR" id="PTR" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="PTR" class="record-btn">PTR</label>

              <!-- RP -->
              <input v-model="recordType" value="RP" id="RP" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="RP" class="record-btn">RP</label>

              <!-- RRSIG -->
              <input v-model="recordType" value="RRSIG" id="RRSIG" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="RRSIG" class="record-btn">RRSIG</label>

              <!-- RT -->
              <input v-model="recordType" value="RT" id="RT" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="RT" class="record-btn">RT</label>

              <!-- SOA -->
              <input v-model="recordType" value="SOA" id="SOA" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="SOA" class="record-btn">SOA</label>

              <!-- SRV -->
              <input v-model="recordType" value="SRV" id="SRV" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="SRV" class="record-btn">SRV</label>

              <!-- TLSA -->
              <input v-model="recordType" value="TLSA" id="TLSA" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="TLSA" class="record-btn">TLSA</label>

              <!-- TXT -->
              <input v-model="recordType" value="TXT" id="TXT" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="TXT" class="record-btn">TXT</label>

              <!-- X25 -->
              <input v-model="recordType" value="X25" id="X25" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="X25" class="record-btn">X25</label>
            </div>
            
            <div class="all-records-btn">
              <input v-model="recordType" value="ALL" id="ALL" type="radio" class="hidden-radio" :disabled="isLoading" />
              <label for="ALL" class="all-btn">ALL THE RECORDS</label>
            </div>
          </div>
        </form>
      </div>

      <!-- Loading Indicator -->
      <div v-if="isLoading" class="loading-indicator">
        <div class="loading-spinner"></div>
        <p>Querying DNS records...</p>
      </div>

      <!-- Error Message -->
      <div v-if="error" class="error-message">
        {{ error }}
      </div>

      <!-- Results Section -->
      <div v-for="(result, index) in results" :key="index" class="result-card">
        <div class="result-header">
          <span class="result-badge">{{ result.type.toUpperCase() }} answers</span>
        </div>
        
        <div v-if="result.answers && result.answers.length > 0">
          <!-- Table for large screens -->
          <table class="answers-table desktop-view">
            <thead>
              <tr class="table-header">
                <th class="header-cell">name</th>
                <th class="header-cell">TTL</th>
                <th class="header-cell whitespace-nowrap">record type</th>
                <th class="header-cell">value</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(answer, ansIndex) in result.answers" :key="ansIndex" class="table-row">
                <td class="table-cell">{{ answer.name }}</td>
                <td class="table-cell">{{ answer.ttl }}</td>
                <td class="table-cell type-cell">{{ answer.type }}</td>
                <td class="table-cell value-cell">{{ answer.value }}</td>
              </tr>
            </tbody>
          </table>

          <!-- List for small screens -->
          <div v-for="(answer, ansIndex) in result.answers" :key="ansIndex" class="mobile-view">
            <ul class="answer-list">
              <li><span class="list-label">Name:</span> {{ answer.name }}</li>
              <li><span class="list-label">TTL:</span> {{ answer.ttl }}</li>
              <li><span class="list-label">Record Type:</span> {{ answer.type }}</li>
              <li><span class="list-label">Value:</span> {{ answer.value }}</li>
            </ul>
            <div v-if="ansIndex < result.answers.length - 1" class="list-divider"></div>
          </div>
        </div>
        
        <div v-else class="no-results">
          &lt;no result&gt;
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, onUnmounted } from 'vue';
// const API_URL = import.meta.env.VITE_API_URL;
const API_URL = "http://192.168.20.17:8082";
export default {
  name: 'DnsLookup',

  data() {
    return {
      domain: '',
      recordType: 'A',
      results: [],
      isLoading: false,
      error: null,
      allRecordTypes: [
        'A', 'AAAA', 'AFSDB', 'ATMA', 'CAA', 'CERT', 'CNAME', 'DHCID',
        'DNAME', 'DNSKEY', 'DS', 'HINFO', 'ISDN', 'LOC', 'MB', 'MG',
        'MINFO', 'MR', 'MX', 'NAPTR', 'NS', 'NSAP-PTR', 'NSEC', 'NSEC3',
        'PTR', 'RP', 'RRSIG', 'RT', 'SOA', 'SRV', 'TLSA', 'TXT', 'X25'
      ]
    }
  },

  mounted() {
    const hash = window.location.hash.substring(1)
    if (hash) {
      const parts = hash.split('|')
      if (parts.length === 2) {
        this.domain = parts[0]
        this.recordType = parts[1].toUpperCase()
        this.performLookup()
      }
    }
  },

  methods: {
    handleSubmit() {
      if (!this.domain.trim()) {
        this.error = 'Please enter a domain name'
        return
      }
      this.performLookup()
    },

    async performLookup() {
      if (!this.domain.trim()) {
        this.results = []
        return
      }

      this.isLoading = true
      this.error = null
      this.results = []

      try {
        const types = this.recordType === 'ALL' 
          ? this.allRecordTypes 
          : [this.recordType]

        // Query types sequentially instead of all at once
        for (const type of types) {
          const response = await this.dnsQuery(this.domain, type)
          if (response && response.answers && response.answers.length > 0) {
            this.results.push({
              type: type,
              answers: response.answers,
              timestamp: new Date().toISOString()
            })
          } else if (this.recordType !== 'ALL') {
            // Only show empty result for single queries, not for ALL
            this.results.push({
              type: type,
              answers: [],
              timestamp: new Date().toISOString()
            })
          }
        }

        // Update URL hash
        window.location.hash = `${this.domain}|${this.recordType}`
      } catch (error) {
        console.error('DNS lookup error:', error)
        this.error = `Failed to query DNS: ${error.message}`
      } finally {
        this.isLoading = false
      }
    },

 async dnsQuery(domain, type) {
  try {
    // Using POST request as per your Go API
    const requestBody = {
      domain: domain,
      type: type,
      server: '1.1.1.1' // Default DNS server
    }

    console.log('Making DNS query:', { domain, type }) // Debug log
    //  const response = await fetch(`${API_URL}/v1/services/isp`);
    const response = await fetch(`${API_URL}/v1/api/dnsCheck`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      body: JSON.stringify(requestBody)
    })

    console.log('Response status:', response.status) // Debug log
    
    if (!response.ok) {
      let errorText = 'Unknown error'
      try {
        const errorData = await response.json()
        errorText = errorData.error || JSON.stringify(errorData)
      } catch (e) {
        errorText = await response.text()
      }
      throw new Error(`HTTP ${response.status}: ${errorText}`)
    }

    const data = await response.json()
    console.log('API Response data:', data) // Debug log - check this in browser console
    
    // Format response based on your Go API structure
    const formatted = this.formatResponse(data, domain, type)
    console.log('Formatted response:', formatted) // Debug log
    
    return formatted
  } catch (error) {
    console.error(`DNS query failed (${type}):`, error)
    return { answers: [] }
  }
},

 formatResponse(data, domain, type) {
  // Format response from your Go API - UPDATED
  let answers = []
  
  // Check the actual structure from your Go API
  if (data.answers && Array.isArray(data.answers)) {
    // This matches your Go API's APIResponse struct
    answers = data.answers.map(answer => ({
      name: answer.name || domain,
      ttl: answer.ttl || answer.TTL || 0,
      type: answer.type || type,
      value: answer.value || ''
    }))
  } else if (data.Answers && Array.isArray(data.Answers)) {
    // Alternative format
    answers = data.Answers.map(answer => ({
      name: answer.name || domain,
      ttl: answer.ttl || answer.TTL || 0,
      type: answer.type || type,
      value: answer.value || answer.data || ''
    }))
  } else if (Array.isArray(data)) {
    // Direct array format
    answers = data.map(item => ({
      name: item.name || domain,
      ttl: item.ttl || item.TTL || 0,
      type: item.type || type,
      value: item.value || item.data || ''
    }))
  } else if (data.result && Array.isArray(data.result)) {
    // Another possible format
    answers = data.result.map(item => ({
      name: item.name || domain,
      ttl: item.ttl || 0,
      type: item.type || type,
      value: item.value || ''
    }))
  }
  
  return {
    domain: domain,
    type: type,
    answers: answers
  }
}
  },

  watch: {
    recordType() {
      if (this.domain) {
        this.performLookup()
      }
    }
  }
}
</script>

<style scoped>
/* Dark Mode Theme */
.dns-lookup-page {
  font-family: 'Roboto', 'Helvetica', 'Arial', sans-serif;
  min-height: 100vh;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  margin: 0;
  padding: 0;
  color: #e2e8f0;
}

.dns-container {
  max-width: 72rem;
  margin: 0 auto;
  padding: 20px;
  box-sizing: border-box;
}

/* Input Card */
.input-card {
  margin: 2rem auto 0 auto;
  width: 100%;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  padding: 30px;
  border-radius: 16px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.input-group {
  margin-bottom: 30px;
}

.relative {
  position: relative;
}

.domain-input {
  width: 100%;
  padding: 12px 0;
  font-size: 16px;
  background: transparent;
  border: none;
  border-bottom: 2px solid #334155;
  outline: none;
  transition: border-color 0.3s;
  font-family: inherit;
  box-sizing: border-box;
  color: #e2e8f0;
}

.domain-input:focus {
  border-bottom-color: #8b5cf6;
}

.domain-input:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  background-color: transparent;
}

.domain-input::placeholder {
  color: #64748b;
}

.input-label {
  font-size: 14px;
  color: #8b5cf6;
  font-weight: 500;
  position: absolute;
  left: 0;
  top: -24px;
}

/* Record Type Selection */
.type-selection {
  margin-top: 30px;
}

.radio-group {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
}

.hidden-radio {
  display: none;
}

.record-btn {
  display: inline-block;
  margin: 4px;
  background-color: #1e293b;
  font-weight: 500;
  font-size: 14px;
  color: #cbd5e1;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;
  border: 1px solid #334155;
}

.record-btn:hover {
  background-color: #2d3748;
  border-color: #8b5cf6;
}

.hidden-radio:checked + .record-btn {
  color: white;
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);
  border-color: #8b5cf6;
}

.hidden-radio:checked + .record-btn:hover {
  background: linear-gradient(135deg, #7c3aed 0%, #6d28d9 100%);
}

.hidden-radio:disabled + .record-btn {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
}

.all-records-btn {
  text-align: center;
  margin-top: 16px;
}

.all-btn {
  display: inline-block;
  background-color: #1e293b;
  font-weight: 500;
  font-size: 14px;
  color: #cbd5e1;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #334155;
}

.all-btn:hover {
  background-color: #2d3748;
  border-color: #8b5cf6;
}

.hidden-radio:checked + .all-btn {
  color: white;
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);
  border-color: #8b5cf6;
}

.hidden-radio:checked + .all-btn:hover {
  background: linear-gradient(135deg, #7c3aed 0%, #6d28d9 100%);
}

.hidden-radio:disabled + .all-btn {
  opacity: 0.5;
  cursor: not-allowed;
  pointer-events: none;
}

/* Loading Indicator */
.loading-indicator {
  text-align: center;
  margin: 2rem auto;
  padding: 2rem;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #1e293b;
  border-top-color: #8b5cf6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 16px auto;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading-indicator p {
  color: #cbd5e1;
  font-weight: 500;
  margin: 0;
}

/* Error Message */
.error-message {
  margin: 2rem auto;
  padding: 16px;
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border-radius: 8px;
  border-left: 4px solid #ef4444;
  text-align: center;
  max-width: 600px;
  font-weight: 500;
}

/* Results Card */
.result-card {
  position: relative;
  margin: 2rem auto;
  width: 100%;
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  padding: 30px;
  border-radius: 16px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(148, 163, 184, 0.1);
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.result-header {
  position: absolute;
  left: 16px;
  top: -16px;
}

.result-badge {
  display: inline-block;
  border-radius: 20px;
  padding: 8px 20px;
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  color: white;
  font-weight: 600;
  font-size: 14px;
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

/* Answers Table */
.answers-table {
  width: 100%;
  font-family: 'Monaco', 'Courier New', monospace;
  border-collapse: collapse;
  font-size: 14px;
  color: #e2e8f0;
}

.desktop-view {
  display: table;
}

.table-header {
  border-bottom: 1px solid #334155;
}

.header-cell {
  padding: 16px;
  text-align: left;
  font-weight: 500;
  color: #94a3b8;
  background-color: #0f172a;
}

.table-row {
  border-bottom: 1px solid #334155;
}

.table-row:last-child {
  border-bottom: none;
}

.table-cell {
  padding: 12px 16px;
  vertical-align: top;
  color: #cbd5e1;
}

.type-cell {
  min-width: 100px;
  white-space: nowrap;
  color: #8b5cf6;
}

.value-cell {
  width: 100%;
  word-break: break-word;
  color: #60a5fa;
}

/* Mobile View */
.mobile-view {
  display: none;
}

.answer-list {
  font-family: 'Monaco', 'Courier New', monospace;
  padding: 16px 0;
  margin: 0;
  list-style: none;
  font-size: 14px;
}

.answer-list li {
  margin-bottom: 8px;
  line-height: 1.5;
  color: #cbd5e1;
}

.list-label {
  color: #8b5cf6;
  font-weight: 600;
  margin-right: 8px;
}

.list-divider {
  border-top: 1px solid #334155;
  margin: 16px 0;
}

.no-results {
  font-family: 'Monaco', 'Courier New', monospace;
  color: #64748b;
  padding: 20px;
  text-align: center;
  font-size: 14px;
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

/* Responsive Design */
@media (max-width: 768px) {
  .dns-container {
    padding: 10px;
  }
  
  .input-card, .result-card {
    padding: 20px;
  }
  
  .desktop-view {
    display: none;
  }
  
  .mobile-view {
    display: block;
  }
  
  .radio-group {
    justify-content: flex-start;
  }
  
  .record-btn, .all-btn {
    font-size: 13px;
    padding: 6px 12px;
  }
}

@media (min-width: 769px) {
  .desktop-view {
    display: table;
  }
  
  .mobile-view {
    display: none;
  }
}

/* Focus styles for accessibility */
.record-btn:focus-visible,
.all-btn:focus-visible,
.domain-input:focus-visible {
  outline: 2px solid #8b5cf6;
  outline-offset: 2px;
}

/* Utility classes */
.whitespace-nowrap {
  white-space: nowrap;
}
</style>