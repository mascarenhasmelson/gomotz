<template>
  <div class="keyword-monitor">
    <!-- Connection Status -->
    <div class="connection-bar" :class="connectionStatus">
      <span class="status-indicator"></span>
      <span class="status-text">{{ connectionMessage }}</span>
      <button v-if="connectionStatus === 'disconnected'" @click="connectWebSocket" class="reconnect-btn">
        Reconnect
      </button>
    </div>

    <!-- Add New Monitor Form -->
    <div class="add-monitor-card">
      <div class="card-header">
        <h2>Add New HTTP/HTTPS Monitor</h2>
      </div>
      
      <div class="add-monitor-form">
        <!-- Basic Information -->
        <div class="form-section">
          <h3>Basic Information</h3>
          <div class="form-row">
            <div class="form-group">
              <label for="name">Friendly Name <span class="required">*</span></label>
              <input
                type="text"
                id="name"
                v-model="newMonitor.name"
                placeholder="e.g., Google API"
                class="form-input"
                :disabled="isAdding"
              />
            </div>
            
            <div class="form-group">
              <label for="url">URL <span class="required">*</span></label>
              <input
                type="url"
                id="url"
                v-model="newMonitor.url"
                placeholder="https://api.example.com/endpoint"
                class="form-input"
                :disabled="isAdding"
                @keyup.enter="addMonitor"
              />
            </div>
          </div>
        </div>

        <!-- HTTP Method and Authentication -->
        <div class="form-section">
          <h3>Request Configuration</h3>
          <div class="form-row">
            <div class="form-group">
              <label for="method">HTTP Method</label>
              <select
                id="method"
                v-model="newMonitor.method"
                class="form-select"
                :disabled="isAdding"
                @change="onMethodChange"
              >
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="PATCH">PATCH</option>
                <option value="DELETE">DELETE</option>
                <option value="HEAD">HEAD</option>
                <option value="OPTIONS">OPTIONS</option>
              </select>
            </div>
            
            <div class="form-group">
              <label for="authType">Authentication</label>
              <select
                id="authType"
                v-model="newMonitor.authType"
                class="form-select"
                :disabled="isAdding"
                @change="onAuthTypeChange"
              >
                <option value="none">None</option>
                <option value="basic">Basic Auth</option>
                <option value="bearer">Bearer Token</option>
                <option value="oauth2">OAuth2</option>
                <option value="digest">Digest Auth</option>
              </select>
            </div>
          </div>

          <!-- Basic Auth Fields -->
          <div v-if="newMonitor.authType === 'basic'" class="auth-fields">
            <div class="form-row">
              <div class="form-group">
                <label for="username">Username</label>
                <input
                  type="text"
                  id="username"
                  v-model="newMonitor.username"
                  placeholder="username"
                  class="form-input"
                />
              </div>
              <div class="form-group">
                <label for="password">Password</label>
                <input
                  type="password"
                  id="password"
                  v-model="newMonitor.password"
                  placeholder="••••••••"
                  class="form-input"
                />
              </div>
            </div>
          </div>

          <!-- Bearer Token Field -->
          <div v-if="newMonitor.authType === 'bearer'" class="auth-fields">
            <div class="form-row">
              <div class="form-group">
                <label for="token">Bearer Token</label>
                <input
                  type="text"
                  id="token"
                  v-model="newMonitor.token"
                  placeholder="eyJhbGciOiJIUzI1NiIs..."
                  class="form-input"
                />
              </div>
            </div>
          </div>

          <!-- OAuth2 Fields -->
          <div v-if="newMonitor.authType === 'oauth2'" class="auth-fields">
            <div class="form-row">
              <div class="form-group">
                <label for="oauthToken">OAuth2 Token</label>
                <input
                  type="text"
                  id="oauthToken"
                  v-model="newMonitor.oauthToken"
                  placeholder="OAuth2 token"
                  class="form-input"
                />
              </div>
              <div class="form-group">
                <label for="tokenType">Token Type</label>
                <select
                  id="tokenType"
                  v-model="newMonitor.tokenType"
                  class="form-select"
                >
                  <option value="Bearer">Bearer</option>
                  <option value="Basic">Basic</option>
                  <option value="Digest">Digest</option>
                </select>
              </div>
            </div>
          </div>
        </div>

        <!-- Headers Section -->
        <div class="form-section">
          <div class="section-header" @click="showHeaders = !showHeaders">
            <h3>HTTP Headers</h3>
            <button class="toggle-btn" type="button">
              <span :class="{ 'rotated': showHeaders }">▼</span>
            </button>
          </div>
          
          <div v-if="showHeaders" class="headers-section">
            <div class="headers-list">
              <div v-for="(header, index) in newMonitor.headers" :key="index" class="header-row">
                <input
                  type="text"
                  v-model="header.key"
                  placeholder="Header name"
                  class="form-input header-key"
                />
                <input
                  type="text"
                  v-model="header.value"
                  placeholder="Header value"
                  class="form-input header-value"
                />
                <button class="remove-btn" @click="removeHeader(index)" type="button">✕</button>
              </div>
            </div>
            <button class="add-header-btn" @click="addHeader" type="button">+ Add Header</button>
            
            <div class="preset-headers">
              <span class="preset-label">Presets:</span>
              <button class="preset-btn" @click="addContentTypeHeader('json')" type="button">JSON</button>
              <button class="preset-btn" @click="addContentTypeHeader('xml')" type="button">XML</button>
              <button class="preset-btn" @click="addContentTypeHeader('form')" type="button">Form</button>
              <button class="preset-btn" @click="addAuthHeader('basic')" type="button">Basic Auth</button>
              <button class="preset-btn" @click="addAuthHeader('bearer')" type="button">Bearer</button>
            </div>
          </div>
        </div>

        <!-- Body Section -->
        <div v-if="['POST', 'PUT', 'PATCH'].includes(newMonitor.method)" class="form-section">
          <div class="section-header" @click="showBody = !showBody">
            <h3>Request Body</h3>
            <button class="toggle-btn" type="button">
              <span :class="{ 'rotated': showBody }">▼</span>
            </button>
          </div>
          
          <div v-if="showBody" class="body-section">
            <div class="form-row">
              <div class="form-group">
                <label for="contentType">Content Type</label>
                <select
                  id="contentType"
                  v-model="newMonitor.contentType"
                  class="form-select"
                  @change="onContentTypeChange"
                >
                  <option value="application/json">application/json</option>
                  <option value="application/xml">application/xml</option>
                  <option value="application/x-www-form-urlencoded">application/x-www-form-urlencoded</option>
                  <option value="multipart/form-data">multipart/form-data</option>
                  <option value="text/plain">text/plain</option>
                </select>
              </div>
              
              <div class="form-group">
                <label for="encoding">Body Encoding</label>
                <select
                  id="encoding"
                  v-model="newMonitor.bodyEncoding"
                  class="form-select"
                >
                  <option value="utf-8">UTF-8</option>
                  <option value="ascii">ASCII</option>
                  <option value="base64">Base64</option>
                </select>
              </div>
            </div>

            <!-- JSON Body Editor -->
            <div v-if="newMonitor.contentType === 'application/json'" class="json-editor">
              <div class="editor-tabs">
                <button 
                  class="tab-btn" 
                  :class="{ active: jsonEditMode === 'form' }"
                  @click="jsonEditMode = 'form'"
                >Form</button>
                <button 
                  class="tab-btn" 
                  :class="{ active: jsonEditMode === 'raw' }"
                  @click="jsonEditMode = 'raw'"
                >Raw JSON</button>
              </div>

              <div v-if="jsonEditMode === 'form'" class="json-form">
                <div v-for="(field, index) in newMonitor.jsonFields" :key="index" class="json-field-row">
                  <input
                    type="text"
                    v-model="field.key"
                    placeholder="Field name"
                    class="form-input json-key"
                  />
                  <select v-model="field.type" class="form-select json-type">
                    <option value="string">String</option>
                    <option value="number">Number</option>
                    <option value="boolean">Boolean</option>
                    <option value="object">Object</option>
                    <option value="array">Array</option>
                    <option value="null">Null</option>
                  </select>
                  <input
                    v-if="field.type === 'string'"
                    type="text"
                    v-model="field.value"
                    placeholder="Value"
                    class="form-input json-value"
                  />
                  <input
                    v-else-if="field.type === 'number'"
                    type="number"
                    v-model="field.value"
                    placeholder="Value"
                    class="form-input json-value"
                  />
                  <select
                    v-else-if="field.type === 'boolean'"
                    v-model="field.value"
                    class="form-select json-value"
                  >
                    <option :value="true">true</option>
                    <option :value="false">false</option>
                  </select>
                  <input
                    v-else-if="field.type === 'null'"
                    type="text"
                    value="null"
                    disabled
                    class="form-input json-value"
                  />
                  <textarea
                    v-else
                    v-model="field.value"
                    placeholder="JSON value"
                    class="form-textarea json-value"
                    rows="2"
                  ></textarea>
                  <button class="remove-btn" @click="removeJsonField(index)" type="button">✕</button>
                </div>
                <button class="add-field-btn" @click="addJsonField" type="button">+ Add Field</button>
              </div>

              <div v-else class="json-raw">
                <textarea
                  v-model="newMonitor.rawJsonBody"
                  placeholder='{"key": "value"}'
                  class="form-textarea json-textarea"
                  rows="6"
                ></textarea>
                <button class="format-btn" @click="formatJson" type="button">Format JSON</button>
              </div>
            </div>

            <!-- XML Body Editor -->
            <div v-else-if="newMonitor.contentType === 'application/xml'" class="xml-editor">
              <textarea
                v-model="newMonitor.xmlBody"
                placeholder='<root><element>value</element></root>'
                class="form-textarea xml-textarea"
                rows="6"
              ></textarea>
              <button class="format-btn" @click="formatXml" type="button">Format XML</button>
            </div>

            <!-- Form URL Encoded Editor -->
            <div v-else-if="newMonitor.contentType === 'application/x-www-form-urlencoded'" class="form-data-editor">
              <div v-for="(field, index) in newMonitor.formFields" :key="index" class="form-field-row">
                <input
                  type="text"
                  v-model="field.key"
                  placeholder="Field name"
                  class="form-input form-key"
                />
                <input
                  type="text"
                  v-model="field.value"
                  placeholder="Field value"
                  class="form-input form-value"
                />
                <button class="remove-btn" @click="removeFormField(index)" type="button">✕</button>
              </div>
              <button class="add-field-btn" @click="addFormField" type="button">+ Add Field</button>
            </div>

            <!-- Multipart Form Data -->
            <div v-else-if="newMonitor.contentType === 'multipart/form-data'" class="multipart-editor">
              <div v-for="(field, index) in newMonitor.multipartFields" :key="index" class="multipart-field-row">
                <input
                  type="text"
                  v-model="field.key"
                  placeholder="Field name"
                  class="form-input multipart-key"
                />
                <select v-model="field.type" class="form-select multipart-type">
                  <option value="text">Text</option>
                  <option value="file">File</option>
                </select>
                <input
                  v-if="field.type === 'text'"
                  type="text"
                  v-model="field.value"
                  placeholder="Value"
                  class="form-input multipart-value"
                />
                <input
                  v-else
                  type="file"
                  @change="handleFileSelect(index, $event)"
                  class="form-input multipart-file"
                />
                <button class="remove-btn" @click="removeMultipartField(index)" type="button">✕</button>
              </div>
              <button class="add-field-btn" @click="addMultipartField" type="button">+ Add Field</button>
            </div>

            <!-- Plain Text Editor -->
            <div v-else>
              <textarea
                v-model="newMonitor.textBody"
                placeholder="Request body"
                class="form-textarea"
                rows="4"
              ></textarea>
            </div>
          </div>
        </div>

        <!-- Response Validation -->
        <div class="form-section">
          <h3>Response Validation</h3>
          <div class="form-row">
            <div class="form-group">
              <label for="keyword">Keyword to Search <span class="required">*</span></label>
              <input
                type="text"
                id="keyword"
                v-model="newMonitor.keyword"
                placeholder="e.g., success, OK, 200"
                class="form-input"
                :disabled="isAdding"
              />
            </div>
            
            <div class="form-group">
              <label for="expectedStatus">Expected Status Code</label>
              <select
                id="expectedStatus"
                v-model="newMonitor.expectedStatus"
                class="form-select"
                :disabled="isAdding"
              >
                <option :value="200">200 OK</option>
                <option :value="201">201 Created</option>
                <option :value="202">202 Accepted</option>
                <option :value="204">204 No Content</option>
                <option :value="301">301 Moved Permanently</option>
                <option :value="302">302 Found</option>
                <option :value="304">304 Not Modified</option>
                <option :value="400">400 Bad Request</option>
                <option :value="401">401 Unauthorized</option>
                <option :value="403">403 Forbidden</option>
                <option :value="404">404 Not Found</option>
                <option :value="500">500 Internal Server</option>
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="newMonitor.caseSensitive">
                <span>Case Sensitive Keyword</span>
              </label>
            </div>
            <div class="form-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="newMonitor.invertKeyword">
                <span>Invert (Fail if keyword exists)</span>
              </label>
            </div>
            <div class="form-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="newMonitor.useRegex">
                <span>Use Regular Expression</span>
              </label>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="responsePath">JSON/XML Path (optional)</label>
              <input
                type="text"
                id="responsePath"
                v-model="newMonitor.responsePath"
                placeholder="$.data.status or /root/element"
                class="form-input"
              />
              <span class="input-hint">Use JSONPath or XPath to extract specific field</span>
            </div>
          </div>
        </div>

        <!-- Monitoring Settings -->
        <div class="form-section">
          <h3>Monitoring Settings</h3>
          <div class="form-row">
            <div class="form-group heartbeat-group">
              <label>Heartbeat Interval</label>
              <div class="heartbeat-interval-container">
                <div class="heartbeat-visualization">
                  <div class="heartbeat-bar">
                    <div 
                      class="heartbeat-fill" 
                      :style="{ width: (newMonitor.interval / 300) * 100 + '%' }"
                    ></div>
                  </div>
                  <div class="heartbeat-markers">
                    <span class="marker">1m</span>
                    <span class="marker">5m</span>
                    <span class="marker">10m</span>
                    <span class="marker">30m</span>
                    <span class="marker">1h</span>
                  </div>
                </div>

                <div class="heartbeat-input-group">
                  <div class="heartbeat-number-input">
                    <input 
                      type="number" 
                      v-model="newMonitor.interval" 
                      min="5"
                      max="3600"
                      step="5"
                      class="heartbeat-number-field"
                    />
                    <span class="heartbeat-unit">seconds</span>
                  </div>
                  <div class="heartbeat-display">
                    <span class="heartbeat-value">{{ formatHeartbeatTime(newMonitor.interval) }}</span>
                    <span class="heartbeat-badge" :class="getHeartbeatCategory(newMonitor.interval)">
                      {{ getHeartbeatCategory(newMonitor.interval) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="timeout">Timeout (seconds)</label>
              <input
                type="number"
                id="timeout"
                v-model="newMonitor.timeout"
                min="1"
                max="30"
                step="1"
                class="form-input small-input"
              />
            </div>
            <div class="form-group">
              <label for="retries">Retries</label>
              <input
                type="number"
                id="retries"
                v-model="newMonitor.retries"
                min="0"
                max="5"
                class="form-input small-input"
              />
            </div>
            <div class="form-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="newMonitor.followRedirects">
                <span>Follow Redirects</span>
              </label>
            </div>
          </div>
        </div>

        <!-- Test Button -->
        <div class="test-section">
          <button class="btn test-btn" @click="testConnection" :disabled="!newMonitor.url || !isConnected">
            <span class="btn-icon">🔍</span>
            Test Configuration
          </button>
        </div>

        <!-- Action Buttons -->
        <div class="form-actions">
          <button class="btn btn-primary" @click="addMonitor" :disabled="!newMonitor.url || !newMonitor.keyword || !newMonitor.name || isAdding || !isConnected">
            <span class="btn-icon">➕</span>
            {{ isAdding ? 'Adding...' : 'Add Monitor' }}
          </button>
          <button class="btn btn-danger" @click="resetForm">
            <span class="btn-icon">↺</span>
            Reset
          </button>
        </div>
      </div>
    </div>

    <!-- Monitors Grid -->
    <div class="monitors-grid">
      <div v-for="monitor in monitors" :key="monitor.id" class="monitor-card" :class="monitor.status">
        <div class="card-header">
          <div class="header-left">
            <div class="status-indicator" :class="monitor.status"></div>
            <div class="title-section">
              <h3 class="monitor-name">{{ monitor.name }}</h3>
              <div class="monitor-meta">
                <span class="monitor-method">{{ monitor.method }}</span>
                <a :href="monitor.url" target="_blank" class="monitor-url">{{ truncateUrl(monitor.url) }}</a>
              </div>
            </div>
          </div>
          <div class="header-actions">
            <span class="keyword-badge">"{{ truncateKeyword(monitor.keyword) }}"</span>
            <button class="icon-btn" @click="editMonitor(monitor)" title="Edit Monitor">✏️</button>
            <button class="icon-btn delete" @click="deleteMonitor(monitor)" title="Delete Monitor">🗑️</button>
          </div>
        </div>

        <div class="heartbeat-timeline">
          <div 
            v-for="(beat, index) in monitor.heartbeats" 
            :key="index"
            class="heartbeat-block"
            :class="beat.status"
            :title="getHeartbeatTitle(beat)"
          ></div>
          <div v-if="!monitor.heartbeats || monitor.heartbeats.length === 0" class="no-data">
            No data yet
          </div>
        </div>

        <div class="stats-grid">
          <div class="stat-item">
            <span class="stat-label">Status</span>
            <span class="stat-value" :class="getStatusCodeClass(monitor.lastStatusCode)">
              {{ monitor.lastStatusCode || 'N/A' }}
            </span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Keyword Found</span>
            <span class="stat-value" :class="monitor.lastKeywordFound ? 'success' : 'error'">
              {{ monitor.lastKeywordFound ? '✅' : '❌' }}
            </span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Response Time</span>
            <span class="stat-value" :class="getLatencyClass(monitor.lastResponseTime)">
              {{ monitor.lastResponseTime || 0 }}ms
            </span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Uptime (24h)</span>
            <span class="stat-value" :class="getUptimeClass(monitor.stats?.uptime)">
              {{ monitor.stats?.uptime || 0 }}%
            </span>
          </div>
        </div>

        <div class="request-info">
          <span class="request-badge" :class="monitor.authType">{{ monitor.authType !== 'none' ? monitor.authType : 'No Auth' }}</span>
          <span class="content-badge" v-if="monitor.contentType">{{ monitor.contentType.split('/')[1] }}</span>
          <span class="headers-count" v-if="monitor.headers?.length">{{ monitor.headers.length }} headers</span>
        </div>

        <div class="footer-info">
          <div class="info-item">
            <span class="info-label">Last Check:</span>
            <span class="info-value">{{ formatTime(monitor.lastCheck) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Next Check:</span>
            <span class="info-value">{{ formatNextCheck(monitor.nextCheck) }}</span>
          </div>
        </div>

        <div v-if="monitor.lastError" class="error-banner">
          <span class="error-icon">⚠️</span>
          <span class="error-text">{{ monitor.lastError }}</span>
        </div>
      </div>

      <div v-if="monitors.length === 0 && !loading" class="empty-state">
        <div class="empty-icon">🔍</div>
        <h3>No Monitors Added</h3>
        <p>Add your first HTTP/HTTPS endpoint to start monitoring</p>
      </div>

      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <p>Loading monitors...</p>
      </div>
    </div>

    <!-- Test Connection Modal -->
    <div v-if="showTestModal" class="modal-overlay" @click.self="closeTestModal">
      <div class="modal-content large">
        <div class="modal-header">
          <h3>Test Results</h3>
          <button class="close-btn" @click="closeTestModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="test-result" :class="testResult.status">
            <div class="result-icon">{{ testResult.status === 'success' ? '✅' : '❌' }}</div>
            <div class="result-details">
              <h4>{{ testResult.message }}</h4>
              
              <div class="result-tabs">
                <button class="result-tab" :class="{ active: activeResultTab === 'summary' }" @click="activeResultTab = 'summary'">Summary</button>
                <button class="result-tab" :class="{ active: activeResultTab === 'headers' }" @click="activeResultTab = 'headers'">Headers</button>
                <button class="result-tab" :class="{ active: activeResultTab === 'body' }" @click="activeResultTab = 'body'">Response Body</button>
                <button class="result-tab" :class="{ active: activeResultTab === 'request' }" @click="activeResultTab = 'request'">Request</button>
              </div>

              <div v-if="activeResultTab === 'summary'" class="result-meta">
                <div class="meta-item"><span class="meta-label">URL:</span><span class="meta-value">{{ testResult.data?.url }}</span></div>
                <div class="meta-item"><span class="meta-label">Status Code:</span><span class="meta-value" :class="getStatusCodeClass(testResult.data?.statusCode)">{{ testResult.data?.statusCode }} {{ testResult.data?.statusText }}</span></div>
                <div class="meta-item"><span class="meta-label">Response Time:</span><span class="meta-value">{{ testResult.data?.responseTime }}ms</span></div>
                <div class="meta-item"><span class="meta-label">Content Length:</span><span class="meta-value">{{ testResult.data?.contentLength }} bytes</span></div>
                <div class="meta-item"><span class="meta-label">Content Type:</span><span class="meta-value">{{ testResult.data?.contentType }}</span></div>
                <div class="meta-item"><span class="meta-label">Keyword Found:</span><span class="meta-value">{{ testResult.data?.keywordFound ? '✅' : '❌' }}</span></div>
              </div>

              <div v-if="activeResultTab === 'headers'" class="headers-result">
                <div v-for="(value, key) in testResult.data?.headers" :key="key" class="header-line">
                  <span class="header-key">{{ key }}:</span>
                  <span class="header-value">{{ value }}</span>
                </div>
              </div>

              <div v-if="activeResultTab === 'body'" class="body-result">
                <div class="body-controls">
                  <button class="format-btn" @click="formatResponseBody" v-if="testResult.data?.contentType?.includes('json')">Format JSON</button>
                  <button class="copy-btn" @click="copyResponseBody">Copy to Clipboard</button>
                </div>
                <pre class="response-body">{{ testResult.data?.body }}</pre>
              </div>

              <div v-if="activeResultTab === 'request'" class="request-result">
                <div class="request-section">
                  <h4>Request Headers</h4>
                  <div v-for="(value, key) in testResult.data?.requestHeaders" :key="key" class="header-line">
                    <span class="header-key">{{ key }}:</span>
                    <span class="header-value">{{ value }}</span>
                  </div>
                </div>
                <div v-if="testResult.data?.requestBody" class="request-section">
                  <h4>Request Body</h4>
                  <pre class="request-body">{{ testResult.data.requestBody }}</pre>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-primary" @click="addFromTest" v-if="testResult.status === 'success'">Add Monitor</button>
          <button class="btn btn-secondary" @click="closeTestModal">Close</button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="closeDeleteModal">
      <div class="modal-content small">
        <div class="modal-header">
          <h3>Delete Monitor</h3>
          <button class="close-btn" @click="closeDeleteModal">&times;</button>
        </div>
        <div class="modal-body">
          <p>Are you sure you want to delete monitor for <strong>{{ monitorToDelete?.name }}</strong>?</p>
          <p class="warning">This action cannot be undone.</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-danger" @click="confirmDelete">Delete</button>
          <button class="btn btn-secondary" @click="closeDeleteModal">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive, onMounted, onUnmounted } from 'vue'

const WS_URL = 'ws://localhost:8082/ws/keyword'

export default {
  name: 'KeywordMonitor',
  
  setup() {
    const monitors = ref([])
    const loading = ref(true)
    const isAdding = ref(false)
    const isConnected = ref(false)
    const connectionStatus = ref('disconnected')
    const connectionMessage = ref('Disconnected')
    
    const showHeaders = ref(false)
    const showBody = ref(false)
    const jsonEditMode = ref('form')
    const activeResultTab = ref('summary')
    
    const newMonitor = reactive({
      name: '',
      url: '',
      method: 'GET',
      authType: 'none',
      username: '',
      password: '',
      token: '',
      oauthToken: '',
      tokenType: 'Bearer',
      headers: [],
      contentType: 'application/json',
      bodyEncoding: 'utf-8',
      jsonFields: [],
      rawJsonBody: '',
      xmlBody: '',
      formFields: [],
      multipartFields: [],
      textBody: '',
      keyword: '',
      expectedStatus: 200,
      caseSensitive: false,
      invertKeyword: false,
      useRegex: false,
      responsePath: '',
      interval: 60,
      timeout: 10,
      retries: 2,
      followRedirects: true
    })

    const testResult = ref({ status: 'pending', message: '', data: null })
    const showTestModal = ref(false)
    const showDeleteModal = ref(false)
    const monitorToDelete = ref(null)

    let ws = null
    let reconnectTimer = null
    const reconnectAttempts = ref(0)
    const maxReconnectAttempts = 5

    const connectWebSocket = () => {
      connectionStatus.value = 'connecting'
      connectionMessage.value = 'Connecting...'
      
      ws = new WebSocket(WS_URL)
      
      ws.onopen = () => {
        isConnected.value = true
        connectionStatus.value = 'connected'
        connectionMessage.value = 'Connected'
        reconnectAttempts.value = 0
        ws.send(JSON.stringify({ type: 'getMonitors' }))
      }
      
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          handleWebSocketMessage(data)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }
      
      ws.onerror = () => {
        isConnected.value = false
        connectionStatus.value = 'error'
        connectionMessage.value = 'Connection error'
      }
      
      ws.onclose = () => {
        isConnected.value = false
        connectionStatus.value = 'disconnected'
        connectionMessage.value = 'Disconnected'
        
        if (reconnectAttempts.value < maxReconnectAttempts) {
          reconnectAttempts.value++
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.value), 30000)
          connectionMessage.value = `Reconnecting in ${delay/1000}s...`
          
          if (reconnectTimer) clearTimeout(reconnectTimer)
          reconnectTimer = setTimeout(connectWebSocket, delay)
        }
      }
    }

    const handleWebSocketMessage = (data) => {
      switch (data.type) {
        case 'monitors':
          monitors.value = data.monitors || []
          loading.value = false
          break
        case 'monitorAdded':
          monitors.value.push(data.monitor)
          isAdding.value = false
          resetForm()
          break
        case 'monitorUpdated':
          const index = monitors.value.findIndex(m => m.id === data.monitor.id)
          if (index !== -1) monitors.value[index] = data.monitor
          break
        case 'monitorDeleted':
          monitors.value = monitors.value.filter(m => m.id !== data.id)
          break
        case 'checkUpdate':
          updateMonitorWithCheck(data.monitorId, data.check)
          break
        case 'testResult':
          testResult.value = { status: data.success ? 'success' : 'error', message: data.message, data: data.data }
          break
      }
    }

    const updateMonitorWithCheck = (monitorId, check) => {
      const monitor = monitors.value.find(m => m.id === monitorId)
      if (!monitor) return
      
      monitor.lastCheck = check.timestamp
      monitor.lastStatusCode = check.statusCode
      monitor.lastResponseTime = check.responseTime
      monitor.lastKeywordFound = check.keywordFound
      monitor.lastError = check.error
      monitor.status = check.error ? 'offline' : (!check.keywordFound || check.statusCode !== monitor.expectedStatus) ? 'warning' : 'online'
      
      if (!monitor.heartbeats) monitor.heartbeats = []
      monitor.heartbeats.push({ status: monitor.status, timestamp: check.timestamp, responseTime: check.responseTime })
      if (monitor.heartbeats.length > 24) monitor.heartbeats = monitor.heartbeats.slice(-24)
      
      updateMonitorStats(monitor)
      monitor.nextCheck = new Date(Date.now() + (monitor.interval * 1000)).toISOString()
    }

    const updateMonitorStats = (monitor) => {
      const heartbeats = monitor.heartbeats || []
      const total = heartbeats.length
      if (total === 0) {
        monitor.stats = { uptime: 0, avgResponseTime: 0 }
        return
      }
      const online = heartbeats.filter(h => h.status === 'online').length
      const responseTimes = heartbeats.filter(h => h.responseTime).map(h => h.responseTime)
      monitor.stats = {
        uptime: Math.round((online / total) * 100 * 10) / 10,
        avgResponseTime: responseTimes.length ? Math.round(responseTimes.reduce((a, b) => a + b, 0) / responseTimes.length) : 0
      }
    }

    const addHeader = () => newMonitor.headers.push({ key: '', value: '' })
    const removeHeader = (index) => newMonitor.headers.splice(index, 1)

    const addContentTypeHeader = (type) => {
      const types = { json: 'application/json', xml: 'application/xml', form: 'application/x-www-form-urlencoded' }
      newMonitor.headers.push({ key: 'Content-Type', value: types[type] })
    }

    const addAuthHeader = (type) => {
      if (type === 'basic') newMonitor.authType = 'basic'
      else if (type === 'bearer') {
        newMonitor.authType = 'bearer'
        newMonitor.headers.push({ key: 'Authorization', value: 'Bearer {token}' })
      }
    }

    const addJsonField = () => newMonitor.jsonFields.push({ key: '', type: 'string', value: '' })
    const removeJsonField = (index) => newMonitor.jsonFields.splice(index, 1)

    const formatJson = () => {
      try {
        newMonitor.rawJsonBody = JSON.stringify(JSON.parse(newMonitor.rawJsonBody), null, 2)
      } catch (e) {
        alert('Invalid JSON')
      }
    }

    const addFormField = () => newMonitor.formFields.push({ key: '', value: '' })
    const removeFormField = (index) => newMonitor.formFields.splice(index, 1)

    const addMultipartField = () => newMonitor.multipartFields.push({ key: '', type: 'text', value: '' })
    const removeMultipartField = (index) => newMonitor.multipartFields.splice(index, 1)

    const handleFileSelect = (index, event) => {
      const file = event.target.files[0]
      if (file) newMonitor.multipartFields[index].fileName = file.name
    }

    const formatXml = () => {}

    const onMethodChange = () => {
      if (['POST', 'PUT', 'PATCH'].includes(newMonitor.method)) showBody.value = true
    }

    const onAuthTypeChange = () => {
      newMonitor.username = newMonitor.password = newMonitor.token = newMonitor.oauthToken = ''
    }

    const onContentTypeChange = () => {
      newMonitor.jsonFields = []
      newMonitor.rawJsonBody = ''
      newMonitor.xmlBody = ''
      newMonitor.formFields = []
      newMonitor.multipartFields = []
      newMonitor.textBody = ''
    }

    const buildRequestBody = () => {
      if (newMonitor.contentType === 'application/json') {
        if (jsonEditMode.value === 'form') {
          const obj = {}
          newMonitor.jsonFields.forEach(field => {
            if (field.key) {
              if (field.type === 'number') obj[field.key] = Number(field.value)
              else if (field.type === 'boolean') obj[field.key] = field.value === 'true'
              else if (field.type === 'null') obj[field.key] = null
              else if (field.type === 'object' || field.type === 'array') {
                try { obj[field.key] = JSON.parse(field.value) } catch { obj[field.key] = field.value }
              } else obj[field.key] = field.value
            }
          })
          return JSON.stringify(obj)
        } else return newMonitor.rawJsonBody
      } else if (newMonitor.contentType === 'application/xml') return newMonitor.xmlBody
      else if (newMonitor.contentType === 'application/x-www-form-urlencoded') {
        const params = new URLSearchParams()
        newMonitor.formFields.forEach(field => { if (field.key) params.append(field.key, field.value) })
        return params.toString()
      } else return newMonitor.textBody
    }

    const addMonitor = () => {
      if (!newMonitor.url || !newMonitor.keyword || !newMonitor.name || !isConnected.value) return
      isAdding.value = true
      ws.send(JSON.stringify({
        type: 'addMonitor',
        monitor: {
          name: newMonitor.name,
          url: newMonitor.url,
          method: newMonitor.method,
          authType: newMonitor.authType,
          username: newMonitor.username,
          password: newMonitor.password,
          token: newMonitor.token,
          oauthToken: newMonitor.oauthToken,
          tokenType: newMonitor.tokenType,
          headers: newMonitor.headers.filter(h => h.key && h.value),
          contentType: newMonitor.contentType,
          bodyEncoding: newMonitor.bodyEncoding,
          body: buildRequestBody(),
          keyword: newMonitor.keyword,
          expectedStatus: newMonitor.expectedStatus,
          caseSensitive: newMonitor.caseSensitive,
          invertKeyword: newMonitor.invertKeyword,
          useRegex: newMonitor.useRegex,
          responsePath: newMonitor.responsePath,
          interval: newMonitor.interval,
          timeout: newMonitor.timeout,
          retries: newMonitor.retries,
          followRedirects: newMonitor.followRedirects
        }
      }))
    }

    const testConnection = () => {
      if (!newMonitor.url || !isConnected.value) return
      ws.send(JSON.stringify({
        type: 'testConnection',
        config: {
          url: newMonitor.url,
          method: newMonitor.method,
          authType: newMonitor.authType,
          username: newMonitor.username,
          password: newMonitor.password,
          token: newMonitor.token,
          oauthToken: newMonitor.oauthToken,
          tokenType: newMonitor.tokenType,
          headers: newMonitor.headers.filter(h => h.key && h.value),
          contentType: newMonitor.contentType,
          bodyEncoding: newMonitor.bodyEncoding,
          body: buildRequestBody(),
          keyword: newMonitor.keyword,
          expectedStatus: newMonitor.expectedStatus,
          caseSensitive: newMonitor.caseSensitive,
          invertKeyword: newMonitor.invertKeyword,
          useRegex: newMonitor.useRegex,
          responsePath: newMonitor.responsePath,
          timeout: newMonitor.timeout,
          followRedirects: newMonitor.followRedirects
        }
      }))
      showTestModal.value = true
      testResult.value = { status: 'pending', message: 'Testing...', data: null }
    }

    const addFromTest = () => { addMonitor(); closeTestModal() }
    const editMonitor = (monitor) => { Object.assign(newMonitor, monitor); monitorToDelete.value = monitor; confirmDelete(true) }
    const deleteMonitor = (monitor) => { monitorToDelete.value = monitor; showDeleteModal.value = true }
    const confirmDelete = (skipConfirm = false) => {
      if (!monitorToDelete.value) return
      if (!skipConfirm) closeDeleteModal()
      ws.send(JSON.stringify({ type: 'deleteMonitor', id: monitorToDelete.value.id }))
      monitorToDelete.value = null
    }
    const closeDeleteModal = () => { showDeleteModal.value = false; monitorToDelete.value = null }
    const resetForm = () => {
      newMonitor.name = ''; newMonitor.url = ''; newMonitor.method = 'GET'; newMonitor.authType = 'none'
      newMonitor.username = ''; newMonitor.password = ''; newMonitor.token = ''; newMonitor.oauthToken = ''
      newMonitor.tokenType = 'Bearer'; newMonitor.headers = []; newMonitor.contentType = 'application/json'
      newMonitor.bodyEncoding = 'utf-8'; newMonitor.jsonFields = []; newMonitor.rawJsonBody = ''
      newMonitor.xmlBody = ''; newMonitor.formFields = []; newMonitor.multipartFields = []; newMonitor.textBody = ''
      newMonitor.keyword = ''; newMonitor.expectedStatus = 200; newMonitor.caseSensitive = false
      newMonitor.invertKeyword = false; newMonitor.useRegex = false; newMonitor.responsePath = ''
      newMonitor.interval = 60; newMonitor.timeout = 10; newMonitor.retries = 2; newMonitor.followRedirects = true
      showHeaders.value = false; showBody.value = false; jsonEditMode.value = 'form'
    }
    const closeTestModal = () => { showTestModal.value = false; activeResultTab.value = 'summary' }
    const formatResponseBody = () => {}
    const copyResponseBody = () => {
      if (testResult.value.data?.body) navigator.clipboard.writeText(testResult.value.data.body)
    }

    const formatHeartbeatTime = (seconds) => seconds < 60 ? `${seconds}s` : `${Math.floor(seconds / 60)}m`
    const getHeartbeatCategory = (seconds) => seconds <= 60 ? 'very-fast' : seconds <= 300 ? 'fast' : seconds <= 900 ? 'medium' : seconds <= 1800 ? 'slow' : 'very-slow'
    const getHeartbeatTitle = (beat) => `${new Date(beat.timestamp).toLocaleString()} - ${beat.responseTime ? beat.responseTime + 'ms' : 'No response'}`
    const getUptimeClass = (uptime) => uptime >= 99 ? 'excellent' : uptime >= 95 ? 'good' : uptime >= 90 ? 'fair' : 'poor'
    const getLatencyClass = (latency) => !latency ? 'offline' : latency < 100 ? 'excellent' : latency < 300 ? 'good' : latency < 500 ? 'fair' : 'poor'
    const getStatusCodeClass = (code) => !code ? '' : code >= 200 && code < 300 ? 'success' : code >= 300 && code < 400 ? 'warning' : 'error'
    const formatTime = (timestamp) => {
      if (!timestamp) return 'Never'
      const diff = Math.floor((new Date() - new Date(timestamp)) / 1000)
      return diff < 60 ? 'Just now' : diff < 3600 ? `${Math.floor(diff / 60)}m ago` : diff < 86400 ? `${Math.floor(diff / 3600)}h ago` : new Date(timestamp).toLocaleDateString()
    }
    const formatNextCheck = (timestamp) => {
      if (!timestamp) return 'Not scheduled'
      const diff = Math.floor((new Date(timestamp) - new Date()) / 1000)
      return diff < 0 ? 'Now' : diff < 60 ? `in ${diff}s` : diff < 3600 ? `in ${Math.floor(diff / 60)}m` : diff < 86400 ? `in ${Math.floor(diff / 3600)}h` : `in ${Math.floor(diff / 86400)}d`
    }
    const truncateUrl = (url) => url.length > 40 ? url.substring(0, 37) + '...' : url
    const truncateKeyword = (keyword) => keyword.length > 20 ? keyword.substring(0, 17) + '...' : keyword

    onMounted(() => connectWebSocket())
    onUnmounted(() => { if (ws) ws.close(); if (reconnectTimer) clearTimeout(reconnectTimer) })

    return {
      monitors, loading, isAdding, isConnected, connectionStatus, connectionMessage,
      showHeaders, showBody, jsonEditMode, activeResultTab, newMonitor, testResult,
      showTestModal, showDeleteModal, monitorToDelete,
      connectWebSocket, addHeader, removeHeader, addContentTypeHeader, addAuthHeader,
      addJsonField, removeJsonField, formatJson, addFormField, removeFormField,
      addMultipartField, removeMultipartField, handleFileSelect, formatXml,
      onMethodChange, onAuthTypeChange, onContentTypeChange, addMonitor, testConnection,
      addFromTest, editMonitor, deleteMonitor, confirmDelete, closeDeleteModal,
      resetForm, closeTestModal, formatResponseBody, copyResponseBody,
      formatHeartbeatTime, getHeartbeatCategory, getHeartbeatTitle, getUptimeClass,
      getLatencyClass, getStatusCodeClass, formatTime, formatNextCheck,
      truncateUrl, truncateKeyword
    }
  }
}
</script>

<style scoped>
/* Dark Mode Theme */
.keyword-monitor {
  min-height: 100vh;
  background: linear-gradient(135deg, #0a0c10 0%, #1a1e24 100%);
  color: #e2e8f0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  padding: 24px;
}

/* Dashboard Header */
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
}

.subtitle {
  color: #94a3b8;
  font-size: 1rem;
  margin: 0;
}

/* Connection Bar */
.connection-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  margin-bottom: 24px;
  border-radius: 8px;
  font-size: 0.95rem;
}

.connection-bar.connected {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
  color: #34d399;
}

.connection-bar.connecting {
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
  color: #fbbf24;
}

.connection-bar.disconnected {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.connected .status-indicator {
  background: #34d399;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
  animation: pulse 2s infinite;
}

.connecting .status-indicator {
  background: #fbbf24;
  animation: pulse 1s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.reconnect-btn {
  margin-left: auto;
  padding: 6px 12px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s;
}

.reconnect-btn:hover {
  background: #dc2626;
  transform: translateY(-1px);
}

/* Add Monitor Card */
.add-monitor-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  margin-bottom: 32px;
  overflow: hidden;
}

.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  background: rgba(15, 23, 42, 0.4);
}

.card-header h2 {
  margin: 0;
  color: #f8fafc;
  font-size: 1.25rem;
  font-weight: 600;
}

.add-monitor-form {
  padding: 24px;
}

/* Form Sections */
.form-section {
  background: #0f172a;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
  border: 1px solid #334155;
}

.form-section h3 {
  margin: 0 0 16px 0;
  color: #f8fafc;
  font-size: 1rem;
  font-weight: 600;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  margin-bottom: 16px;
}

.section-header h3 {
  margin: 0;
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

.auth-fields {
  margin-top: 16px;
  padding: 16px;
  background: #1e293b;
  border-radius: 8px;
  border: 1px solid #334155;
}

/* Form Layout */
.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  color: #cbd5e1;
  font-size: 0.9rem;
  font-weight: 500;
}

.required {
  color: #ef4444;
  margin-left: 4px;
}

.form-input,
.form-select,
.form-textarea {
  padding: 10px 14px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 8px;
  color: #e2e8f0;
  font-size: 0.95rem;
  transition: all 0.2s;
  font-family: inherit;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
}

.form-input::placeholder {
  color: #64748b;
}

.form-input:disabled,
.form-select:disabled,
.form-textarea:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-textarea {
  resize: vertical;
}

.small-input {
  max-width: 120px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  accent-color: #3b82f6;
  width: 16px;
  height: 16px;
}

.input-hint {
  font-size: 0.7rem;
  color: #64748b;
  margin-top: 4px;
}

/* Heartbeat Interval */
.heartbeat-group {
  gap: 8px;
}

.heartbeat-interval-container {
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 12px;
  padding: 16px;
}

.heartbeat-visualization {
  margin-bottom: 16px;
}

.heartbeat-bar {
  height: 8px;
  background: #1e293b;
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 4px;
}

.heartbeat-fill {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #8b5cf6);
  transition: width 0.3s ease;
}

.heartbeat-markers {
  display: flex;
  justify-content: space-between;
  font-size: 0.7rem;
  color: #64748b;
  padding: 0 4px;
}

.marker {
  position: relative;
}

.marker::before {
  content: '';
  position: absolute;
  top: -12px;
  left: 50%;
  transform: translateX(-50%);
  width: 2px;
  height: 4px;
  background: #334155;
}

.heartbeat-input-group {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.heartbeat-number-input {
  display: flex;
  align-items: center;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  overflow: hidden;
}

.heartbeat-number-field {
  width: 80px;
  padding: 8px;
  background: #1e293b;
  border: none;
  color: #e2e8f0;
  font-size: 0.95rem;
  text-align: center;
}

.heartbeat-number-field:focus {
  outline: none;
}

.heartbeat-unit {
  padding: 8px;
  background: #0f172a;
  color: #94a3b8;
  font-size: 0.85rem;
  border-left: 1px solid #334155;
}

.heartbeat-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.heartbeat-value {
  font-size: 1rem;
  font-weight: 600;
  color: #60a5fa;
}

.heartbeat-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
}

.heartbeat-badge.very-fast {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.heartbeat-badge.fast {
  background: rgba(59, 130, 246, 0.1);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.heartbeat-badge.medium {
  background: rgba(245, 158, 11, 0.1);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.heartbeat-badge.slow {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

/* Headers Section */
.headers-list,
.json-form,
.form-data-editor,
.multipart-editor {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 16px;
}

.header-row,
.json-field-row,
.form-field-row,
.multipart-field-row {
  display: flex;
  gap: 10px;
  align-items: center;
}

.header-key,
.json-key,
.form-key,
.multipart-key {
  flex: 1;
}

.header-value,
.json-value,
.form-value,
.multipart-value {
  flex: 2;
}

.json-type {
  width: 100px;
}

.remove-btn {
  width: 30px;
  height: 30px;
  border-radius: 6px;
  border: none;
  background: #1e293b;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  font-size: 16px;
}

.remove-btn:hover {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
}

.add-header-btn,
.add-field-btn {
  padding: 8px 16px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #cbd5e1;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s;
  align-self: flex-start;
}

.add-header-btn:hover,
.add-field-btn:hover {
  background: #2d3748;
  border-color: #3b82f6;
}

.preset-headers {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.preset-label {
  color: #94a3b8;
  font-size: 0.8rem;
}

.preset-btn {
  padding: 4px 12px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 16px;
  color: #cbd5e1;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.preset-btn:hover {
  background: #2d3748;
  border-color: #3b82f6;
}

/* JSON Editor */
.editor-tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
  border-bottom: 1px solid #334155;
  padding-bottom: 10px;
}

.tab-btn {
  padding: 6px 16px;
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  font-size: 0.9rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.tab-btn:hover {
  color: #60a5fa;
}

.tab-btn.active {
  background: #3b82f6;
  color: white;
}

.json-textarea,
.xml-textarea {
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.85rem;
}

.format-btn {
  margin-top: 10px;
  padding: 6px 12px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #cbd5e1;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
}

.format-btn:hover {
  background: #2d3748;
  border-color: #3b82f6;
}

/* Test Section */
.test-section {
  margin: 20px 0;
}

.test-btn {
  width: 100%;
  justify-content: center;
  padding: 12px;
  background: #4f46e5;
  color: white;
}

.test-btn:hover:not(:disabled) {
  background: #4338ca;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(79, 70, 229, 0.3);
}

/* Form Actions */
.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  flex-wrap: wrap;
}

.btn {
  padding: 10px 20px;
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

.btn-danger:hover {
  background: #dc2626;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.btn-secondary {
  background: #1e293b;
  color: #cbd5e1;
  border: 1px solid #334155;
}

.btn-secondary:hover:not(:disabled) {
  background: #2d3748;
  transform: translateY(-1px);
}

.btn-icon {
  font-size: 1.1rem;
}

/* Monitors Grid */
.monitors-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.monitor-card {
  background: rgba(30, 41, 59, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.3);
  padding: 20px;
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: relative;
}

.monitor-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.4);
}

.monitor-card.online {
  border-left: 4px solid #10b981;
}

.monitor-card.warning {
  border-left: 4px solid #f59e0b;
}

.monitor-card.offline {
  border-left: 4px solid #ef4444;
}

/* Card Header */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.online .status-indicator {
  background: #10b981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
}

.warning .status-indicator {
  background: #f59e0b;
  box-shadow: 0 0 8px rgba(245, 158, 11, 0.5);
}

.offline .status-indicator {
  background: #ef4444;
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.5);
}

.title-section {
  display: flex;
  flex-direction: column;
}

.monitor-name {
  margin: 0;
  color: #f8fafc;
  font-size: 1.1rem;
  font-weight: 600;
}

.monitor-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.monitor-method {
  padding: 2px 8px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 4px;
  font-size: 0.7rem;
  font-weight: 600;
  color: #60a5fa;
}

.monitor-url {
  color: #94a3b8;
  font-size: 0.8rem;
  text-decoration: none;
}

.monitor-url:hover {
  color: #60a5fa;
  text-decoration: underline;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.keyword-badge {
  background: #1e293b;
  color: #60a5fa;
  padding: 4px 10px;
  border-radius: 16px;
  font-size: 0.8rem;
  font-weight: 500;
  border: 1px solid #334155;
}

.icon-btn {
  background: transparent;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 6px;
  border-radius: 6px;
  font-size: 1rem;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: #1e293b;
  color: #60a5fa;
}

.icon-btn.delete:hover {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
}

/* Heartbeat Timeline */
.heartbeat-timeline {
  display: flex;
  gap: 4px;
  height: 40px;
  align-items: center;
  background: #0f172a;
  border-radius: 8px;
  padding: 4px;
}

.heartbeat-block {
  flex: 1;
  height: 30px;
  border-radius: 4px;
  transition: all 0.2s;
  cursor: default;
}

.heartbeat-block:hover {
  transform: scaleY(1.2);
}

.heartbeat-block.online {
  background: #10b981;
}

.heartbeat-block.warning {
  background: #f59e0b;
}

.heartbeat-block.offline {
  background: #ef4444;
}

.no-data {
  flex: 1;
  text-align: center;
  color: #64748b;
  font-size: 0.85rem;
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 12px;
  background: #0f172a;
  border-radius: 8px;
  border: 1px solid #334155;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  font-size: 0.7rem;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 1rem;
  font-weight: 600;
}

.stat-value.excellent {
  color: #34d399;
}

.stat-value.good {
  color: #60a5fa;
}

.stat-value.fair {
  color: #fbbf24;
}

.stat-value.poor {
  color: #f87171;
}

.stat-value.success {
  color: #34d399;
}

.stat-value.error {
  color: #f87171;
}

/* Request Info */
.request-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.75rem;
}

.request-badge,
.content-badge,
.headers-count {
  padding: 2px 8px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 12px;
  color: #94a3b8;
}

.request-badge.basic {
  background: rgba(16, 185, 129, 0.1);
  color: #34d399;
  border-color: rgba(16, 185, 129, 0.3);
}

.request-badge.bearer {
  background: rgba(59, 130, 246, 0.1);
  color: #60a5fa;
  border-color: rgba(59, 130, 246, 0.3);
}

/* Footer Info */
.footer-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #334155;
  font-size: 0.8rem;
}

.info-item {
  display: flex;
  flex-direction: column;
}

.info-label {
  color: #64748b;
  font-size: 0.7rem;
}

.info-value {
  color: #cbd5e1;
  font-size: 0.8rem;
}

/* Error Banner */
.error-banner {
  padding: 8px 12px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #f87171;
  font-size: 0.8rem;
}

.error-icon {
  font-size: 1rem;
}

.error-text {
  flex: 1;
  word-break: break-word;
}

/* Empty State */
.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 80px 20px;
  background: rgba(30, 41, 59, 0.8);
  border-radius: 16px;
  border: 1px dashed rgba(148, 163, 184, 0.2);
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

.empty-state h3 {
  color: #f8fafc;
  margin-bottom: 8px;
}

.empty-state p {
  color: #94a3b8;
}

/* Loading State */
.loading-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #1e293b;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 16px auto;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-state p {
  color: #94a3b8;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #1e293b;
  border-radius: 16px;
  width: 500px;
  max-width: 90vw;
  max-height: 90vh;
  overflow-y: auto;
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5);
}

.modal-content.large {
  width: 800px;
}

.modal-content.small {
  width: 400px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #334155;
  position: sticky;
  top: 0;
  background: #1e293b;
  z-index: 1;
}

.modal-header h3 {
  margin: 0;
  color: #f8fafc;
  font-size: 1.2rem;
}

.close-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  font-size: 24px;
  cursor: pointer;
  padding: 0;
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

.modal-body {
  padding: 24px;
}

.modal-body .warning {
  color: #f87171;
  font-size: 0.9rem;
  margin-top: 8px;
}

/* Test Result Modal */
.test-result {
  display: flex;
  gap: 16px;
  padding: 16px;
  border-radius: 12px;
}

.test-result.success {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.test-result.error {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.test-result.pending {
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.result-icon {
  font-size: 32px;
}

.result-details {
  flex: 1;
}

.result-details h4 {
  margin: 0 0 12px 0;
  color: #f8fafc;
  font-size: 1rem;
}

.result-tabs {
  display: flex;
  gap: 10px;
  margin: 20px 0;
  border-bottom: 1px solid #334155;
  padding-bottom: 10px;
}

.result-tab {
  padding: 8px 16px;
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  font-size: 0.9rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.result-tab:hover {
  color: #60a5fa;
}

.result-tab.active {
  background: #3b82f6;
  color: white;
}

.result-meta {
  background: #0f172a;
  border-radius: 8px;
  padding: 16px;
  border: 1px solid #334155;
  max-height: 400px;
  overflow-y: auto;
}

.meta-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #1e293b;
}

.meta-item:last-child {
  border-bottom: none;
}

.meta-label {
  color: #64748b;
  font-size: 0.85rem;
}

.meta-value {
  color: #e2e8f0;
  font-weight: 500;
}

.headers-result,
.body-result,
.request-result {
  max-height: 400px;
  overflow-y: auto;
  padding: 16px;
  background: #0f172a;
  border-radius: 8px;
  border: 1px solid #334155;
}

.headers-section,
.request-section {
  margin-bottom: 20px;
}

.headers-section h4,
.request-section h4 {
  margin: 0 0 10px 0;
  color: #f8fafc;
  font-size: 0.9rem;
}

.header-line {
  display: flex;
  padding: 6px 0;
  border-bottom: 1px solid #1e293b;
  font-size: 0.85rem;
}

.header-key {
  width: 200px;
  color: #94a3b8;
}

.header-value {
  flex: 1;
  color: #60a5fa;
  word-break: break-word;
}

.body-controls {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.copy-btn {
  padding: 6px 12px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #cbd5e1;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
}

.copy-btn:hover {
  background: #2d3748;
  border-color: #3b82f6;
}

.response-body,
.request-body {
  background: #0f172a;
  padding: 12px;
  border-radius: 6px;
  font-family: 'Monaco', 'Courier New', monospace;
  font-size: 0.8rem;
  color: #94a3b8;
  max-height: 300px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 24px 24px;
  border-top: 1px solid #334155;
  position: sticky;
  bottom: 0;
  background: #1e293b;
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
@media (max-width: 768px) {
  .keyword-monitor {
    padding: 16px;
  }
  
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
  
  .monitors-grid {
    grid-template-columns: 1fr;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .footer-info {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .header-row,
  .json-field-row,
  .form-field-row,
  .multipart-field-row {
    flex-direction: column;
    align-items: stretch;
  }
  
  .json-type {
    width: 100%;
  }
  
  .remove-btn {
    align-self: flex-end;
  }
  
  .result-tabs {
    flex-wrap: wrap;
  }
  
  .header-line {
    flex-direction: column;
  }
  
  .header-key {
    width: 100%;
  }
}
</style>