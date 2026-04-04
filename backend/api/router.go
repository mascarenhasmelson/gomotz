package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mascarenhasmelson/gomotz/bgservices"
	"github.com/mascarenhasmelson/gomotz/discovery/vlan"
	ws "github.com/mascarenhasmelson/gomotz/discovery/vlan"
	"github.com/mascarenhasmelson/gomotz/utils"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4/pgxpool"
)

const defaultScanIntervalSeconds = 30

type Router struct {
	ctx         context.Context
	pool        *pgxpool.Pool
	db          *vlan.PostgresDB
	scanManager *vlan.VLANScanManager
	wsHub       *ws.Hub
	mux         *http.ServeMux
	upgrader    websocket.Upgrader
}

var wsMu sync.Mutex

func NewRouter(ctx context.Context, pool *pgxpool.Pool, database *vlan.PostgresDB, scanManager *vlan.VLANScanManager, wsHub *ws.Hub) *http.ServeMux {
	r := &Router{
		ctx:         ctx,
		pool:        pool,
		db:          database,
		scanManager: scanManager,
		wsHub:       wsHub,
		mux:         http.NewServeMux(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	r.routes()
	return r.mux
}

func (r *Router) routes() {
	r.mux.HandleFunc("/v1/api/services", r.servicesHandler)
	r.mux.HandleFunc("/v1/api/services/", r.serviceHandler)
	r.mux.HandleFunc("/v1/api/services/isp", r.ispHandler)
	// r.mux.HandleFunc("/v1/api/scan", handleSynScan)
	r.mux.HandleFunc("/scan", handleSynScan)
	r.mux.HandleFunc("/v1/api/tcpCheck", TcpCheckHandler)
	r.mux.HandleFunc("/v1/api/dnsCheck", r.dnsQueryHandler)
	r.mux.HandleFunc("/v1/api/traceroute", handleTracerouteWebSocket)
	r.mux.HandleFunc("/v1/api/icmp", PingHandler)
	r.mux.HandleFunc("/v1/api/httpsCheck", handleHTTPS)

	r.mux.HandleFunc("/v1/api/vlans", r.handleVLANs)
	r.mux.HandleFunc("/v1/api/vlans/", r.handleVLANWithID)
	r.mux.HandleFunc("/v1/api/devices", r.getAllDevices)
	r.mux.HandleFunc("/v1/api/scans/", r.handleScans)
	r.mux.HandleFunc("/v1/api/vendors", r.listVendors)
	r.mux.HandleFunc("/v1/api/vendors/stats", r.getVendorStats)
	r.mux.HandleFunc("/v1/api/vendors/cleanup", r.cleanupOldVendors)
	r.mux.HandleFunc("/v1/api/status", r.getStatus)
	r.mux.HandleFunc("/ws", r.handleWebSocket)
	r.mux.HandleFunc("/v1/api/conflicts", r.getConflicts)
	r.mux.HandleFunc("/v1/api/status/detailed", r.getDetailedStatus)
	r.mux.HandleFunc("/v1/api/discovery/status", r.getDiscoveryStatus)
	r.mux.HandleFunc("/v1/api/discovery/trigger", r.triggerDiscovery)

}
func (r *Router) getDetailedStatus(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	statuses := r.scanManager.GetAllStatuses()

	result := make([]map[string]interface{}, 0, len(statuses))
	for vlanID, scanner := range statuses {
		status := map[string]interface{}{
			"vlan_id":        vlanID,
			"vlan_name":      scanner.Config.VLANName,
			"status":         scanner.Status,
			"is_running":     scanner.IsRunning,
			"last_scan_time": scanner.LastScanTime,
			"host_count":     scanner.Scanner.GetHostCount(),
		}

		// Add interface status
		cidr, _ := r.scanManager.GetCIDRFromConfig(scanner.Config)
		if cidr != "" {
			iface, err := r.scanManager.DetectInterfaceForCIDR(cidr, vlanID)
			if err != nil {
				status["interface_status"] = "down"
				status["interface_error"] = err.Error()
			} else {
				status["interface_status"] = "up"
				status["interface_name"] = iface
			}
		}

		result = append(result, status)
	}

	respondJSON(w, http.StatusOK, result)
}

type CreateVLANRequest struct {
	VLANId              int    `json:"vlan_id"`
	VLANName            string `json:"vlan_name"`
	NetworkMode         string `json:"network_mode"` // "static" or "dhcp"
	IPAddress           string `json:"ip_address,omitempty"`
	CIDRNotation        string `json:"cidr_notation,omitempty"`
	DefaultGateway      string `json:"default_gateway,omitempty"`
	EnableMonitoring    bool   `json:"enable_monitoring"`
	ScanIntervalSeconds int    `json:"scan_interval_seconds,omitempty"` // 0 → use default
}

func (r *Router) getDiscoveryStatus(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	hostnameDiscovery := r.scanManager.GetHostnameDiscovery()

	status := make(map[string]interface{})

	for ifaceName, hd := range hostnameDiscovery {
		stats := hd.GetStats()
		cache := hd.GetAllCached()

		status[ifaceName] = map[string]interface{}{
			"cache_size":         len(cache),
			"total_scans":        stats.TotalScans,
			"ssdp_discoveries":   stats.SSDPDiscoveries,
			"mdns_discoveries":   stats.MDNSDiscoveries,
			"last_scan_time":     stats.LastScanTime,
			"last_scan_duration": stats.LastScanDuration.String(),
			"cached_hostnames":   cache,
		}
	}

	respondJSON(w, http.StatusOK, status)
}

func (r *Router) triggerDiscovery(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if req.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	iface := req.URL.Query().Get("interface")

	triggered := r.scanManager.TriggerDiscoveryScan(iface)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"triggered":  triggered,
		"message":    "Hostname discovery scan triggered",
		"interfaces": triggered,
	})
}
func (r *Router) getConflicts(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	query := `
		SELECT 
			vlan_id,
			host(ip_address) AS ip_address,
			mac_address,
			COALESCE(hostname, '') AS hostname,
			COALESCE(vendor, '') AS vendor,
			device_status,
			last_seen
		FROM discovered_devices
		WHERE device_status = 'conflict'
		ORDER BY last_seen DESC
	`

	rows, err := r.pool.Query(r.ctx, query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type Conflict struct {
		VLANId       int       `json:"vlan_id"`
		IPAddress    string    `json:"ip_address"`
		MACAddress   string    `json:"mac_address"`
		Hostname     string    `json:"hostname"`
		Vendor       string    `json:"vendor"`
		DeviceStatus string    `json:"device_status"`
		LastSeen     time.Time `json:"last_seen"`
	}

	conflicts := []Conflict{}
	for rows.Next() {
		var c Conflict
		if err := rows.Scan(&c.VLANId, &c.IPAddress, &c.MACAddress,
			&c.Hostname, &c.Vendor, &c.DeviceStatus, &c.LastSeen); err != nil {
			continue
		}
		conflicts = append(conflicts, c)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"conflicts": conflicts,
		"count":     len(conflicts),
	})
}
func (r *Router) handleVLANs(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	switch req.Method {
	case http.MethodGet:
		r.listVLANs(w, req)
	case http.MethodPost:
		r.createVLAN(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleVLANWithID(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/api/vlans/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "VLAN ID required", http.StatusBadRequest)
		return
	}

	vlanID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid VLAN ID", http.StatusBadRequest)
		return
	}

	if len(parts) > 1 {
		switch parts[1] {
		case "devices":
			r.getDevicesByVLAN(w, req, vlanID)
		// case "start":
		// 	r.startScan(w, req, vlanID)
		// case "stop":
		// 	r.stopScan(w, req, vlanID)
		default:
			http.Error(w, "Unknown sub-route", http.StatusNotFound)
		}
		return
	}

	switch req.Method {
	case http.MethodGet:
		r.getVLAN(w, req, vlanID)
	case http.MethodPut:
		r.updateVLAN(w, req, vlanID)
	case http.MethodDelete:
		r.deleteVLAN(w, req, vlanID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) listVLANs(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vlans, err := r.db.GetAllVLANs(r.ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, vlans)
}

func (r *Router) getVLAN(w http.ResponseWriter, req *http.Request, vlanID int) {
	w.Header().Set("Content-Type", "application/json")
	v, err := r.db.GetVLANNetwork(r.ctx, vlanID)
	if err != nil {
		respondError(w, http.StatusNotFound, "VLAN not found")
		return
	}
	respondJSON(w, http.StatusOK, v)
}

func (r *Router) createVLAN(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var vlanReq CreateVLANRequest
	if err := json.NewDecoder(req.Body).Decode(&vlanReq); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if vlanReq.VLANId < 1 || vlanReq.VLANId > 4094 {
		respondError(w, http.StatusBadRequest, "VLAN ID must be between 1 and 4094")
		return
	}
	if vlanReq.NetworkMode != "static" && vlanReq.NetworkMode != "dhcp" {
		respondError(w, http.StatusBadRequest, "Network mode must be 'static' or 'dhcp'")
		return
	}

	// FIX: respect caller-supplied interval; fall back to default if absent or invalid.
	scanInterval := vlanReq.ScanIntervalSeconds
	if scanInterval <= 0 {
		scanInterval = defaultScanIntervalSeconds
	}

	v := &utils.VLANNetwork{
		VLANId:              vlanReq.VLANId,
		VLANName:            vlanReq.VLANName,
		NetworkMode:         vlanReq.NetworkMode,
		MonitoringEnabled:   vlanReq.EnableMonitoring,
		ScanIntervalSeconds: scanInterval,
	}

	if vlanReq.NetworkMode == "static" {
		if vlanReq.IPAddress == "" || vlanReq.CIDRNotation == "" {
			respondError(w, http.StatusBadRequest, "Static mode requires ip_address and cidr_notation")
			return
		}
		cidrFull, err := calculateCIDRFull(vlanReq.IPAddress, vlanReq.CIDRNotation)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid IP/CIDR: %v", err))
			return
		}
		v.IPAddress = &vlanReq.IPAddress
		v.CIDRNotation = &vlanReq.CIDRNotation
		v.CIDRFull = &cidrFull
		if vlanReq.DefaultGateway != "" {
			v.DefaultGateway = &vlanReq.DefaultGateway
		}
	}

	if err := r.db.CreateVLANNetwork(r.ctx, v); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			respondError(w, http.StatusConflict, "VLAN ID already exists")
		} else {
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if v.MonitoringEnabled {
		if err := r.scanManager.StartVLANScan(v); err != nil {
			respondError(w, http.StatusInternalServerError,
				fmt.Sprintf("VLAN created but scan failed to start: %v", err))
			return
		}
	}

	respondJSON(w, http.StatusCreated, v)
}

func (r *Router) updateVLAN(w http.ResponseWriter, req *http.Request, vlanID int) {
	w.Header().Set("Content-Type", "application/json")

	var vlanReq CreateVLANRequest
	if err := json.NewDecoder(req.Body).Decode(&vlanReq); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	existing, err := r.db.GetVLANNetwork(r.ctx, vlanID)
	if err != nil {
		respondError(w, http.StatusNotFound, "VLAN not found")
		return
	}

	scanInterval := vlanReq.ScanIntervalSeconds
	if scanInterval <= 0 {
		scanInterval = existing.ScanIntervalSeconds
		if scanInterval <= 0 {
			scanInterval = defaultScanIntervalSeconds
		}
	}

	v := &utils.VLANNetwork{
		VLANId:              vlanID,
		VLANName:            vlanReq.VLANName,
		NetworkMode:         vlanReq.NetworkMode,
		MonitoringEnabled:   vlanReq.EnableMonitoring,
		ScanIntervalSeconds: scanInterval,
	}

	if vlanReq.NetworkMode == "static" {
		cidrFull, err := calculateCIDRFull(vlanReq.IPAddress, vlanReq.CIDRNotation)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid IP/CIDR: %v", err))
			return
		}
		v.IPAddress = &vlanReq.IPAddress
		v.CIDRNotation = &vlanReq.CIDRNotation
		v.CIDRFull = &cidrFull
		if vlanReq.DefaultGateway != "" {
			v.DefaultGateway = &vlanReq.DefaultGateway
		}
	}

	if err := r.db.UpdateVLANNetwork(r.ctx, v); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	r.scanManager.StopVLANScan(vlanID)

	if v.MonitoringEnabled {
		if err := r.scanManager.StartVLANScan(v); err != nil {
			log.Printf("[VLAN %d] Updated but failed to start: %v (will retry automatically)",
				vlanID, err)
		}
	}

	respondJSON(w, http.StatusOK, v)
}

func (r *Router) deleteVLAN(w http.ResponseWriter, req *http.Request, vlanID int) {
	w.Header().Set("Content-Type", "application/json")

	r.scanManager.StopVLANScan(vlanID)

	if err := r.db.DeleteVLANNetwork(r.ctx, vlanID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[VLAN %d] Deleted from database", vlanID)

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "deleted",
		"vlan_id": strconv.Itoa(vlanID),
		"message": "VLAN deleted successfully",
	})
}

// func (r *Router) startScan(w http.ResponseWriter, req *http.Request, vlanID int) {
// 	if EnableCORS(&w, req) {
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")

// 	config, err := r.db.GetVLANNetwork(r.ctx, vlanID)
// 	if err != nil {
// 		respondError(w, http.StatusNotFound, "VLAN not found")
// 		return
// 	}

// 	if err := r.scanManager.StartVLANScan(config); err != nil {
// 		respondError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	respondJSON(w, http.StatusOK, map[string]string{"status": "started"})
// }

// func (r *Router) stopScan(w http.ResponseWriter, req *http.Request, vlanID int) {
// 	if EnableCORS(&w, req) {
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")

// 	if err := r.scanManager.StopVLANScan(vlanID); err != nil {
// 		respondError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	respondJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
// }

// ============================================
// DEVICE HANDLERS
// ============================================

func (r *Router) getDevicesByVLAN(w http.ResponseWriter, req *http.Request, vlanID int) {
	if EnableCORS(&w, req) { // FIX: was missing CORS
		return
	}
	w.Header().Set("Content-Type", "application/json")

	devices, err := r.db.GetDevicesByVLAN(r.ctx, vlanID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, devices)
}

func (r *Router) getAllDevices(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	devices, err := r.db.GetAllDevices(r.ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, devices)
}

// ============================================
// SCAN LOG HANDLERS
// ============================================

func (r *Router) handleScans(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	path := strings.TrimPrefix(req.URL.Path, "/api/scans/")
	vlanID, err := strconv.Atoi(path)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid VLAN ID")
		return
	}
	r.getScanLogs(w, req, vlanID)
}

func (r *Router) getScanLogs(w http.ResponseWriter, req *http.Request, vlanID int) {
	w.Header().Set("Content-Type", "application/json")

	limit := 20
	if s := req.URL.Query().Get("limit"); s != "" {
		if l, err := strconv.Atoi(s); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := r.db.GetScanLogsByVLAN(r.ctx, vlanID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, logs)
}

// ============================================
// VENDOR HANDLERS
// ============================================

func (r *Router) listVendors(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	vendors, err := r.db.GetAllVendors(r.ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, vendors)
}

func (r *Router) getVendorStats(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	stats, err := r.db.GetVendorStats(r.ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

func (r *Router) cleanupOldVendors(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	daysOld := 90
	if s := req.URL.Query().Get("days"); s != "" {
		if d, err := strconv.Atoi(s); err == nil && d > 0 {
			daysOld = d
		}
	}

	deleted, err := r.db.DeleteOldVendors(r.ctx, daysOld)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"deleted": deleted, "days": daysOld})
}

// ============================================
// STATUS & WEBSOCKET
// ============================================

func (r *Router) getStatus(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	statuses := r.scanManager.GetAllStatuses()

	result := make([]map[string]interface{}, 0, len(statuses))
	for vlanID, scanner := range statuses {
		status := map[string]interface{}{
			"vlan_id":        vlanID,
			"vlan_name":      scanner.Config.VLANName,
			"status":         scanner.Status,
			"is_running":     scanner.IsRunning,
			"last_scan_time": scanner.LastScanTime,
			"host_count":     scanner.Scanner.GetHostCount(),
		}

		cidr, _ := r.scanManager.GetCIDRFromConfig(scanner.Config)
		if cidr != "" {
			iface, err := r.scanManager.DetectInterfaceForCIDR(cidr, vlanID)
			if err != nil {
				status["interface_status"] = "down"
				status["interface_error"] = err.Error()
			} else {
				status["interface_status"] = "up"
				status["interface_name"] = iface
			}
		}

		result = append(result, status)
	}

	respondJSON(w, http.StatusOK, result)
}

func (r *Router) handleWebSocket(w http.ResponseWriter, req *http.Request) {
	if r.wsHub == nil {
		log.Println("WebSocket hub is nil")
		http.Error(w, "WebSocket service not available", http.StatusServiceUnavailable)
		return
	}

	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &vlan.Client{
		Hub:  r.wsHub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	client.Hub.Register <- client
	go client.WritePump()
	go client.ReadPump()
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	if EnableCORS(&w, r) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req utils.HTTPSCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(utils.HTTPSCheckResponse{Error: "Invalid JSON"})
		return
	}
	if req.URL == "" {
		json.NewEncoder(w).Encode(utils.HTTPSCheckResponse{Error: "URL required"})
		return
	}
	if !strings.HasPrefix(req.URL, "https://") {
		req.URL = "https://" + req.URL
	}

	timeout := 10 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	start := time.Now()
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	resp, err := client.Get(req.URL)
	if err != nil {
		json.NewEncoder(w).Encode(utils.HTTPSCheckResponse{HTTPSSupported: false, Error: err.Error()})
		return
	}
	defer resp.Body.Close()

	res := utils.HTTPSCheckResponse{
		HTTPSSupported: true,
		StatusCode:     resp.StatusCode,
		ResponseTime:   time.Since(start).Milliseconds(),
	}

	if resp.TLS != nil {
		res.TLSVersion = bgservices.TLSVersion(resp.TLS.Version)
		res.Cipher = tls.CipherSuiteName(resp.TLS.CipherSuite)
		if req.CheckCertificate && len(resp.TLS.PeerCertificates) > 0 {
			c := resp.TLS.PeerCertificates[0]
			res.Certificate = &utils.CertificateInfo{
				Subject:       c.Subject.CommonName,
				Issuer:        c.Issuer.CommonName,
				ValidFrom:     c.NotBefore.String(),
				ValidUntil:    c.NotAfter.String(),
				DaysRemaining: int(time.Until(c.NotAfter).Hours() / 24),
			}
		}
	}
	if resp.Header.Get("Strict-Transport-Security") != "" {
		res.HSTSEnabled = true
	}
	json.NewEncoder(w).Encode(res)
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	if EnableCORS(&w, r) {
		return
	}
	conn, err := bgservices.Pingupgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	conn.WriteJSON(utils.PingMessage{Type: "connected", Message: "WebSocket connected"})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var req utils.PingRequest
		if err := json.Unmarshal(message, &req); err != nil {
			conn.WriteJSON(utils.PingMessage{Type: "error", Message: "Invalid request format"})
			continue
		}
		if req.Count <= 0 || req.Count > 100 {
			req.Count = 4
		}
		if req.Size < 32 || req.Size > 65500 {
			req.Size = 56
		}
		if req.Timeout <= 0 || req.Timeout > 30 {
			req.Timeout = 2
		}
		if req.Interval <= 0 || req.Interval > 5 {
			req.Interval = 1
		}
		bgservices.PerformPing(conn, req)
	}
}

func handleTracerouteWebSocket(w http.ResponseWriter, r *http.Request) {
	if EnableCORS(&w, r) {
		return
	}
	conn, err := bgservices.TracerouteUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Traceroute WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	var req utils.TracerouteRequest
	if err := conn.ReadJSON(&req); err != nil {
		bgservices.SendTracerouteError(conn, "Invalid request format")
		return
	}
	if req.Target == "" {
		bgservices.SendTracerouteError(conn, "Target is required")
		return
	}
	if req.MaxHops <= 0 || req.MaxHops > 64 {
		req.MaxHops = 30
	}
	if req.ProbesPerHop <= 0 || req.ProbesPerHop > 5 {
		req.ProbesPerHop = 3
	}
	if req.Timeout <= 0 || req.Timeout > 10000 {
		req.Timeout = 2000
	}
	if req.Protocol == "" {
		req.Protocol = "ICMP"
	}

	bgservices.SendTracerouteStatus(conn, "Resolving target hostname...")
	ips, err := net.LookupIP(req.Target)
	if err != nil {
		bgservices.SendTracerouteError(conn, fmt.Sprintf("DNS resolution failed: %v", err))
		return
	}

	var targetIP net.IP
	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			targetIP = ip4
			break
		}
	}
	if targetIP == nil {
		bgservices.SendTracerouteError(conn, "No IPv4 address found for target")
		return
	}

	bgservices.SendTracerouteMessage(conn, utils.TracerouteMessage{
		Type:     "targetIp",
		TargetIP: targetIP.String(),
	})

	switch req.Protocol {
	case "ICMP":
		bgservices.PerformICMPTraceroute(conn, targetIP, req)
	default:
		bgservices.SendTracerouteError(conn, "Unsupported protocol")
	}
}

func (r *Router) dnsQueryHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	switch req.Method {
	case http.MethodPost:
		r.handleDNSPost(w, req)
	case http.MethodGet:
		r.handleDNSGet(w, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleSynScan(w http.ResponseWriter, r *http.Request) {
	ws, err := bgservices.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	var req utils.ScanRequest
	if err := ws.ReadJSON(&req); err != nil {
		return
	}

	go bgservices.SynScan(ws, req.Target)

	for {
		time.Sleep(5 * time.Second)
		wsMu.Lock()
		err := ws.WriteMessage(websocket.PingMessage, nil)
		wsMu.Unlock()
		if err != nil {
			return
		}
	}
}

func (r *Router) servicesHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	switch req.Method {
	case http.MethodGet:
		HandleFetchServices(r.ctx, w, r.pool)
	case http.MethodPost:
		HandleAddService(r.ctx, w, r.pool, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) serviceHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	idStr := strings.TrimPrefix(req.URL.Path, "/services/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}
	if req.Method == http.MethodDelete {
		HandleDeleteService(r.ctx, w, r.pool, id)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) ispHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	switch req.Method {
	case http.MethodGet:
		HandleGetISPInfo(w, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// ============================================
// HELPERS
// ============================================

func calculateCIDRFull(ipAddr, cidr string) (string, error) {
	cidr = strings.TrimPrefix(cidr, "/")
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipAddr)
	}
	prefixLen, err := strconv.Atoi(cidr)
	if err != nil || prefixLen < 0 || prefixLen > 32 {
		return "", fmt.Errorf("invalid CIDR notation: /%s", cidr)
	}
	_, ipNet, err := net.ParseCIDR(fmt.Sprintf("%s/%d", ipAddr, prefixLen))
	if err != nil {
		return "", err
	}
	return ipNet.String(), nil
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
