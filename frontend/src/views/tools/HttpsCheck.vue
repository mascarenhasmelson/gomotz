<template>
  <div class="https-checker-page">
    <div class="https-container">
      <!-- Input Form -->
      <div class="checker-card">
        <div class="input-group">
          <label for="url-input">Website URL</label>
          <div class="input-with-button">
            <input
              v-model="url"
              id="url-input"
              type="text"
              placeholder="Enter website URL (e.g., https://example.com)"
              :disabled="isChecking"
              @keyup.enter="checkHTTPS"
              class="url-input"
            />
            <button 
              @click="checkHTTPS" 
              :disabled="!url || isChecking" 
              class="check-button"
            >
              <span v-if="!isChecking">Check HTTPS</span>
              <span v-else class="checking-text">
                <span class="spinner"></span> Checking...
              </span>
            </button>
          </div>
          <p class="input-hint">Enter URL with or without https:// (e.g., google.com or https://google.com)</p>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="isChecking" class="loading-section">
        <div class="loading-content">
          <div class="loading-spinner-large"></div>
          <p>Analyzing HTTPS configuration...</p>
          <div class="loading-steps">
            <div class="step" :class="{ active: currentStep === 1, completed: currentStep > 1 }">
              <span class="step-icon">1</span>
              <span class="step-text">Resolving domain</span>
            </div>
            <div class="step" :class="{ active: currentStep === 2, completed: currentStep > 2 }">
              <span class="step-icon">2</span>
              <span class="step-text">Checking HTTPS</span>
            </div>
            <div class="step" :class="{ active: currentStep === 3, completed: currentStep > 3 }">
              <span class="step-icon">3</span>
              <span class="step-text">Analyzing SSL/TLS</span>
            </div>
            <div class="step" :class="{ active: currentStep === 4 }">
              <span class="step-icon">4</span>
              <span class="step-text">Getting results</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Error Message -->
      <div v-if="error" class="error-message">
        <span class="error-icon">⚠️</span>
        <div class="error-content">
          <strong>Error:</strong> {{ error }}
        </div>
      </div>

      <!-- Results Section -->
      <div v-if="result && !isChecking" class="results-section">
        <div class="results-header">
          <h2>HTTPS Check Results</h2>
          <div class="results-meta">
            <span class="target-url">{{ url }}</span>
            <span class="check-time">{{ formattedCheckTime }}</span>
          </div>
        </div>

        <!-- Overall Status -->
        <div class="status-card" :class="result.httpsSupported ? 'success' : 'error'">
          <div class="status-icon">
            <span v-if="result.httpsSupported">✅</span>
            <span v-else>❌</span>
          </div>
          <div class="status-content">
            <h3>{{ result.httpsSupported ? 'HTTPS is Supported' : 'HTTPS Not Supported' }}</h3>
            <p v-if="result.httpsSupported">This website supports secure HTTPS connections</p>
            <p v-else>This website does not support HTTPS or has configuration issues</p>
          </div>
        </div>

        <!-- Certificate Details -->
        <div v-if="result.certificate" class="details-grid">
          <div class="detail-card">
            <h4>Certificate Information</h4>
            <div class="detail-list">
              <div class="detail-item">
                <span class="detail-label">Issued To:</span>
                <span class="detail-value">{{ result.certificate.subject || 'N/A' }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Issued By:</span>
                <span class="detail-value">{{ result.certificate.issuer || 'N/A' }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Valid From:</span>
                <span class="detail-value">{{ formatDate(result.certificate.validFrom) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Valid Until:</span>
                <span class="detail-value">{{ formatDate(result.certificate.validUntil) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Days Remaining:</span>
                <span class="detail-value" :class="getDaysRemainingClass(result.certificate.daysRemaining)">
                  {{ result.certificate.daysRemaining || 0 }} days
                </span>
              </div>
            </div>
          </div>

          <!-- Security Details -->
          <div class="detail-card">
            <h4>Security Details</h4>
            <div class="detail-list">
              <div class="detail-item">
                <span class="detail-label">Protocol Version:</span>
                <span class="detail-value">{{ result.protocol || 'N/A' }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Cipher Suite:</span>
                <span class="detail-value">{{ result.cipher || 'N/A' }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">TLS Version:</span>
                <span class="detail-value">{{ result.tlsVersion || 'N/A' }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Key Exchange:</span>
                <span class="detail-value">{{ result.keyExchange || 'N/A' }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">Forward Secrecy:</span>
                <span class="detail-value" :class="{ good: result.forwardSecrecy, bad: !result.forwardSecrecy }">
                  {{ result.forwardSecrecy ? 'Supported' : 'Not Supported' }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Connection Details -->
        <div class="connection-details">
          <h4>Connection Details</h4>
          <div class="connection-grid">
            <div class="connection-item">
              <div class="connection-label">HTTP Redirect</div>
              <div class="connection-value" :class="{ good: result.httpRedirect, bad: !result.httpRedirect }">
                {{ result.httpRedirect ? 'Enabled' : 'Not Enabled' }}
              </div>
            </div>
            <div class="connection-item">
              <div class="connection-label">HSTS Enabled</div>
              <div class="connection-value" :class="{ good: result.hstsEnabled, bad: !result.hstsEnabled }">
                {{ result.hstsEnabled ? 'Yes' : 'No' }}
              </div>
            </div>
            <div class="connection-item">
              <div class="connection-label">Response Time</div>
              <div class="connection-value">{{ result.responseTime || 0 }} ms</div>
            </div>
            <div class="connection-item">
              <div class="connection-label">Status Code</div>
              <div class="connection-value" :class="getStatusCodeClass(result.statusCode)">
                {{ result.statusCode || 'N/A' }}
              </div>
            </div>
          </div>
        </div>

        <!-- Recommendations -->
        <div v-if="result.recommendations && result.recommendations.length > 0" class="recommendations">
          <h4>Recommendations</h4>
          <ul class="recommendation-list">
            <li v-for="(rec, index) in result.recommendations" :key="index" class="recommendation-item">
              <span class="rec-icon">💡</span>
              {{ rec }}
            </li>
          </ul>
        </div>

        <!-- Raw Response (Debug) -->
        <div v-if="showDebug" class="debug-section">
          <button @click="toggleDebug" class="debug-toggle">
            {{ showDebug ? 'Hide' : 'Show' }} Raw Response
          </button>
          <pre v-if="showDebug" class="debug-output">{{ JSON.stringify(result, null, 2) }}</pre>
        </div>
      </div>

      <!-- Recent Checks -->
      <div v-if="recentChecks.length > 0" class="recent-checks">
        <h3>Recent Checks</h3>
        <div class="checks-list">
          <div 
            v-for="(check, index) in recentChecks" 
            :key="index" 
            class="check-item"
            @click="loadRecentCheck(check)"
          >
            <div class="check-status" :class="check.httpsSupported ? 'success' : 'error'">
              {{ check.httpsSupported ? '✅' : '❌' }}
            </div>
            <div class="check-info">
              <div class="check-url">{{ check.url }}</div>
              <div class="check-time">{{ check.time }}</div>
            </div>
            <div class="check-actions">
              <button @click.stop="recheck(check.url)" class="recheck-btn">↻</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
// const API_URL = import.meta.env.VITE_API_URL;
const API_URL = "http://192.168.20.17:8082";
export default {
  name: 'HttpsChecker',
  data() {
    return {
      url: '',
      isChecking: false,
      error: null,
      result: null,
      checkStartTime: null,
      recentChecks: [],
      showDebug: false,
      currentStep: 1,
      stepInterval: null
    }
  },
  computed: {
    formattedCheckTime() {
      if (!this.checkStartTime) return '';
      return new Date(this.checkStartTime).toLocaleString();
    }
  },
  mounted() {
    // Load recent checks from localStorage
    const savedChecks = localStorage.getItem('httpsCheckHistory');
    if (savedChecks) {
      this.recentChecks = JSON.parse(savedChecks);
    }
  },
  beforeUnmount() {
    this.clearStepInterval();
  },
  methods: {
    async checkHTTPS() {
      if (!this.url.trim()) {
        this.error = 'Please enter a website URL';
        return;
      }

      this.isChecking = true;
      this.error = null;
      this.result = null;
      this.checkStartTime = new Date();
      this.currentStep = 1;
      this.showDebug = false;

      // Start step animation
      this.startStepAnimation();

      try {
        // Clean the URL
        const cleanUrl = this.cleanUrl(this.url.trim());
        
        // Call backend API
        const response = await fetch(`${API_URL}/v1/api/httpsCheck`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            url: cleanUrl,
            timeout: 10, // seconds
            checkRedirects: true,
            checkCertificate: true,
            checkSecurityHeaders: true
          })
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        
        // Process the response
        this.processResult(data, cleanUrl);
        
        // Save to recent checks
        this.saveToHistory(data, cleanUrl);
        
      } catch (err) {
        console.error('HTTPS check error:', err);
        this.error = `Failed to check HTTPS: ${err.message}`;
        this.result = {
          url: this.url,
          httpsSupported: false,
          error: err.message,
          timestamp: new Date().toISOString()
        };
      } finally {
        this.isChecking = false;
        this.clearStepInterval();
        this.currentStep = 4;
      }
    },

    cleanUrl(url) {
      // Remove protocol if present, we'll let backend handle it
      url = url.replace(/^https?:\/\//i, '');
      // Remove trailing slash
      url = url.replace(/\/$/, '');
      return url;
    },

    processResult(data, originalUrl) {
      // Format the result for display
      this.result = {
        url: originalUrl,
        httpsSupported: data.httpsSupported || false,
        certificate: data.certificate || null,
        protocol: data.protocol || null,
        cipher: data.cipher || null,
        tlsVersion: data.tlsVersion || null,
        keyExchange: data.keyExchange || null,
        forwardSecrecy: data.forwardSecrecy || false,
        httpRedirect: data.httpRedirect || false,
        hstsEnabled: data.hstsEnabled || false,
        responseTime: data.responseTime || 0,
        statusCode: data.statusCode || null,
        recommendations: data.recommendations || [],
        rawData: data
      };
    },

    saveToHistory(data, url) {
      const checkRecord = {
        url: url,
        httpsSupported: data.httpsSupported || false,
        time: new Date().toLocaleString(),
        timestamp: Date.now(),
        statusCode: data.statusCode,
        responseTime: data.responseTime
      };

      // Add to beginning of array
      this.recentChecks.unshift(checkRecord);
      
      // Keep only last 10 checks
      if (this.recentChecks.length > 10) {
        this.recentChecks = this.recentChecks.slice(0, 10);
      }

      // Save to localStorage
      localStorage.setItem('httpsCheckHistory', JSON.stringify(this.recentChecks));
    },

    loadRecentCheck(check) {
      this.url = check.url;
      this.checkHTTPS();
    },

    async recheck(url) {
      this.url = url;
      this.checkHTTPS();
    },

    startStepAnimation() {
      this.stepInterval = setInterval(() => {
        if (this.currentStep < 4) {
          this.currentStep++;
        } else {
          this.clearStepInterval();
        }
      }, 1000);
    },

    clearStepInterval() {
      if (this.stepInterval) {
        clearInterval(this.stepInterval);
        this.stepInterval = null;
      }
    },

    formatDate(dateString) {
      if (!dateString) return 'N/A';
      try {
        const date = new Date(dateString);
        return date.toLocaleDateString();
      } catch {
        return dateString;
      }
    },

    getDaysRemainingClass(days) {
      if (days === null || days === undefined) return '';
      if (days <= 0) return 'expired';
      if (days <= 7) return 'warning';
      if (days <= 30) return 'expiring';
      return 'good';
    },

    getStatusCodeClass(statusCode) {
      if (!statusCode) return '';
      if (statusCode >= 200 && statusCode < 300) return 'good';
      if (statusCode >= 300 && statusCode < 400) return 'redirect';
      return 'error';
    },

    toggleDebug() {
      this.showDebug = !this.showDebug;
    }
  }
}
</script>

<style scoped>
/* Dark Mode Theme */
.https-checker-page {
  padding: 0;
  width: 100%;
  min-height: 100vh;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  color: #e2e8f0;
}

.https-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 30px;
}

/* Checker Card */
.checker-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  margin-bottom: 30px;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.input-group label {
  display: block;
  font-weight: 600;
  color: #cbd5e1;
  margin-bottom: 12px;
  font-size: 1.1rem;
}

.input-with-button {
  display: flex;
  gap: 15px;
  align-items: center;
}

.url-input {
  flex: 1;
  padding: 16px 20px;
  border: 1px solid #334155;
  border-radius: 10px;
  font-size: 1rem;
  transition: all 0.3s;
  background: #0f172a;
  color: #e2e8f0;
  font-family: 'Monaco', 'Courier New', monospace;
}

.url-input:focus {
  outline: none;
  border-color: #10b981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.2);
}

.url-input::placeholder {
  color: #64748b;
}

.url-input:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  background: #1e293b;
}

.input-hint {
  color: #94a3b8;
  font-size: 0.9rem;
  margin-top: 8px;
  margin-left: 5px;
}

.check-button {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
  padding: 16px 40px;
  border-radius: 10px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  min-width: 140px;
  display: flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
}

.check-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(16, 185, 129, 0.3);
}

.check-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.checking-text {
  display: flex;
  align-items: center;
  gap: 10px;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Loading Section */
.loading-section {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 40px;
  margin-bottom: 30px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  text-align: center;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.loading-spinner-large {
  width: 60px;
  height: 60px;
  border: 4px solid #1e293b;
  border-top-color: #10b981;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 20px auto;
}

.loading-section p {
  color: #cbd5e1;
  font-size: 1.1rem;
  margin-bottom: 30px;
}

.loading-steps {
  display: flex;
  justify-content: center;
  gap: 40px;
  max-width: 600px;
  margin: 0 auto;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #64748b;
}

.step.active {
  color: #10b981;
}

.step.completed {
  color: #10b981;
}

.step-icon {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: #1e293b;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: #94a3b8;
  border: 1px solid #334155;
}

.step.active .step-icon {
  background: #10b981;
  color: white;
  animation: pulse 1.5s infinite;
  border-color: #10b981;
}

.step.completed .step-icon {
  background: #10b981;
  color: white;
  border-color: #10b981;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

.step-text {
  font-size: 0.9rem;
  font-weight: 500;
}

/* Error Message */
.error-message {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border-radius: 10px;
  padding: 20px;
  margin-bottom: 30px;
  display: flex;
  align-items: center;
  gap: 15px;
  border-left: 4px solid #ef4444;
  backdrop-filter: blur(10px);
}

.error-icon {
  font-size: 1.5rem;
}

.error-content {
  flex: 1;
}

.error-content strong {
  font-weight: 600;
}

/* Results Section */
.results-section {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 30px;
  margin-bottom: 30px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(148, 163, 184, 0.1);
  animation: slideIn 0.5s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid #334155;
}

.results-header h2 {
  margin: 0;
  color: #f8fafc;
  font-size: 1.8rem;
}

.results-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 5px;
}

.target-url {
  font-family: 'Monaco', 'Courier New', monospace;
  font-weight: 600;
  color: #60a5fa;
  font-size: 1.1rem;
}

.check-time {
  color: #94a3b8;
  font-size: 0.9rem;
}

/* Status Card */
.status-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 25px;
  border-radius: 12px;
  margin-bottom: 30px;
}

.status-card.success {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.status-card.error {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.status-icon {
  font-size: 2.5rem;
}

.status-content h3 {
  margin: 0 0 8px 0;
  color: #f8fafc;
  font-size: 1.5rem;
}

.status-content p {
  margin: 0;
  color: #94a3b8;
}

/* Details Grid */
.details-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 25px;
  margin-bottom: 30px;
}

@media (max-width: 768px) {
  .details-grid {
    grid-template-columns: 1fr;
  }
}

.detail-card {
  background: #0f172a;
  border-radius: 12px;
  padding: 25px;
  border: 1px solid #334155;
}

.detail-card h4 {
  margin: 0 0 20px 0;
  color: #f8fafc;
  font-size: 1.2rem;
}

.detail-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 15px;
}

.detail-label {
  font-weight: 500;
  color: #94a3b8;
  font-size: 0.95rem;
  min-width: 120px;
}

.detail-value {
  font-weight: 600;
  color: #e2e8f0;
  font-size: 0.95rem;
  text-align: right;
  font-family: 'Monaco', 'Courier New', monospace;
  word-break: break-word;
}

.detail-value.good {
  color: #34d399;
}

.detail-value.bad {
  color: #f87171;
}

.detail-value.warning {
  color: #fbbf24;
}

.detail-value.expiring {
  color: #fbbf24;
}

.detail-value.expired {
  color: #f87171;
  font-weight: 700;
}

/* Connection Details */
.connection-details {
  background: #0f172a;
  border-radius: 12px;
  padding: 25px;
  margin-bottom: 30px;
  border: 1px solid #334155;
}

.connection-details h4 {
  margin: 0 0 20px 0;
  color: #f8fafc;
  font-size: 1.2rem;
}

.connection-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.connection-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.connection-label {
  font-weight: 500;
  color: #94a3b8;
  font-size: 0.9rem;
}

.connection-value {
  font-weight: 600;
  color: #e2e8f0;
  font-size: 1.1rem;
  font-family: 'Monaco', 'Courier New', monospace;
}

.connection-value.good {
  color: #34d399;
}

.connection-value.bad {
  color: #f87171;
}

.connection-value.redirect {
  color: #fbbf24;
}

.connection-value.error {
  color: #f87171;
}

/* Recommendations */
.recommendations {
  background: rgba(245, 158, 11, 0.1);
  border-radius: 12px;
  padding: 25px;
  border: 1px solid rgba(245, 158, 11, 0.3);
  margin-bottom: 30px;
}

.recommendations h4 {
  margin: 0 0 15px 0;
  color: #fbbf24;
  font-size: 1.2rem;
}

.recommendation-list {
  margin: 0;
  padding-left: 0;
  list-style: none;
}

.recommendation-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
  color: #fde68a;
  line-height: 1.5;
}

.recommendation-item:last-child {
  margin-bottom: 0;
}

.rec-icon {
  font-size: 1.2rem;
  flex-shrink: 0;
  margin-top: 2px;
}

/* Debug Section */
.debug-section {
  margin-top: 30px;
}

.debug-toggle {
  background: #1e293b;
  color: #cbd5e1;
  border: 1px solid #334155;
  padding: 10px 20px;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
  margin-bottom: 15px;
}

.debug-toggle:hover {
  background: #2d3748;
  border-color: #3b82f6;
}

.debug-output {
  background: #0f172a;
  color: #94a3b8;
  padding: 20px;
  border-radius: 8px;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.85rem;
  overflow-x: auto;
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #334155;
}

/* Recent Checks */
.recent-checks {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 25px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.recent-checks h3 {
  margin: 0 0 20px 0;
  color: #f8fafc;
  font-size: 1.3rem;
}

.checks-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.check-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px 20px;
  background: #0f172a;
  border-radius: 10px;
  border: 1px solid #334155;
  cursor: pointer;
  transition: all 0.3s;
}

.check-item:hover {
  background: #1e293b;
  border-color: #3b82f6;
  transform: translateX(5px);
}

.check-status {
  font-size: 1.2rem;
  flex-shrink: 0;
}

.check-status.success {
  color: #34d399;
}

.check-status.error {
  color: #f87171;
}

.check-info {
  flex: 1;
}

.check-url {
  font-weight: 600;
  color: #f8fafc;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.9rem;
  margin-bottom: 4px;
}

.check-time {
  color: #94a3b8;
  font-size: 0.8rem;
}

.check-actions {
  flex-shrink: 0;
}

.recheck-btn {
  background: #1e293b;
  color: #94a3b8;
  border: 1px solid #334155;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  font-size: 1.2rem;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.recheck-btn:hover {
  background: #10b981;
  color: white;
  transform: rotate(90deg);
  border-color: #10b981;
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

/* Responsive */
@media (max-width: 768px) {
  .https-container {
    padding: 20px;
  }
  
  .input-with-button {
    flex-direction: column;
  }
  
  .url-input,
  .check-button {
    width: 100%;
  }
  
  .results-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }
  
  .results-meta {
    align-items: flex-start;
  }
  
  .loading-steps {
    flex-direction: column;
    gap: 20px;
  }
}
</style>