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

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mascarenhasmelson/gomotz/bgservices"
	"github.com/mascarenhasmelson/gomotz/discovery/vlan"
	ws "github.com/mascarenhasmelson/gomotz/discovery/vlan"
	"github.com/mascarenhasmelson/gomotz/monitorsrv"
	"github.com/mascarenhasmelson/gomotz/utils"
)

const defaultScanIntervalSeconds = 30

type Router struct {
	ctx         context.Context
	pool        *pgxpool.Pool
	db          *vlan.PostgresDB
	monitorDB   *monitorsrv.PostgresDB
	scanManager *vlan.VLANScanManager
	portMonitor *monitorsrv.PortMonitorService
	snmpMonitor *monitorsrv.SNMPMonitorService
	pingMonitor *monitorsrv.PingMonitorService
	wsHub       *ws.Hub
	mux         *http.ServeMux
	upgrader    websocket.Upgrader
}
type MonitorRequest struct {
	Name         string `json:"name"`
	ScanInterval int    `json:"scan_interval"`
}

var wsMu sync.Mutex

func NewRouter(ctx context.Context, pool *pgxpool.Pool, database *vlan.PostgresDB, monitorDB *monitorsrv.PostgresDB, scanManager *vlan.VLANScanManager, portMonitor *monitorsrv.PortMonitorService, snmpMonitor *monitorsrv.SNMPMonitorService, pingMonitor *monitorsrv.PingMonitorService, wsHub *ws.Hub) *http.ServeMux {
	r := &Router{
		ctx:         ctx,
		pool:        pool,
		db:          database,
		monitorDB:   monitorDB,
		scanManager: scanManager,
		portMonitor: portMonitor,
		snmpMonitor: snmpMonitor,
		pingMonitor: pingMonitor,
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

	r.mux.HandleFunc("/v1/api/interfaces", r.getInterfaces)
	r.mux.HandleFunc("/v1/api/interfaces/", r.handleInterfaceAction)

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
	r.mux.HandleFunc("/v1/api/monitors", r.handlePortMonitors)
	r.mux.HandleFunc("/v1/api/monitors/", r.handlePortMonitorWithID)
	r.mux.HandleFunc("/v1/ws/monitors", r.handleMonitorWebSocket)
	r.mux.HandleFunc("/v1/api/monitors/test", r.testPortConnection)

	r.mux.HandleFunc("/v1/api/snmp", r.handleSNMPMonitors)
	r.mux.HandleFunc("/v1/api/snmp/test", r.testSNMPConnection)
	r.mux.HandleFunc("/v1/api/snmp/", r.handleSNMPMonitorWithID)
	r.mux.HandleFunc("/v1/api/ws/snmp", r.handleSNMPWebSocket)

	r.mux.HandleFunc("/v1/api/ping", r.handlePingMonitors)
	r.mux.HandleFunc("/v1/api/ping/test", r.testPingConnection)
	r.mux.HandleFunc("/v1/api/ping/", r.handlePingMonitorWithID)
	r.mux.HandleFunc("/v1/api/ws/ping", r.handlePingWebSocket)

}
func (r *Router) handlePingMonitors(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if r.pingMonitor == nil {
		respondError(w, http.StatusServiceUnavailable, "ping monitor service not initialized")
		return
	}

	switch req.Method {
	case http.MethodGet:
		monitors, err := r.monitorDB.GetAllPingMonitors(r.ctx)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if monitors == nil {
			monitors = []*utils.PingMonitor{}
		}
		respondJSON(w, http.StatusOK, monitors)

	case http.MethodPost:
		var body utils.CreatePingMonitorRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if body.Hostname == "" {
			respondError(w, http.StatusBadRequest, "hostname required")
			return
		}
		if body.FriendlyName == "" {
			body.FriendlyName = body.Hostname
		}
		if body.CheckInterval <= 0 {
			body.CheckInterval = 60
		}
		if body.LatencyThreshold <= 0 {
			body.LatencyThreshold = 200
		}
		if body.Timeout <= 0 {
			body.Timeout = 3
		}

		m := &utils.PingMonitor{
			FriendlyName:     body.FriendlyName,
			Hostname:         body.Hostname,
			CheckInterval:    body.CheckInterval,
			LatencyThreshold: body.LatencyThreshold,
			Timeout:          body.Timeout,
		}
		if err := r.monitorDB.CreatePingMonitor(r.ctx, m); err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				respondError(w, http.StatusConflict, "monitor for this hostname already exists")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		r.pingMonitor.StartMonitor(m)

		if payload, err := json.Marshal(map[string]interface{}{
			"type": "ping_monitor_created", "monitor": m,
		}); err == nil {
			r.wsHub.Broadcast <- payload
		}

		respondJSON(w, http.StatusCreated, m)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) handlePingMonitorWithID(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(req.URL.Path, "/v1/api/ping/")
	parts := strings.Split(path, "/")

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid monitor ID")
		return
	}
	if len(parts) > 1 && parts[1] == "logs" {
		limit := 50
		if s := req.URL.Query().Get("limit"); s != "" {
			if l, err := strconv.Atoi(s); err == nil && l > 0 {
				limit = l
			}
		}
		logs, err := r.monitorDB.GetPingMonitorLogs(r.ctx, id, limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if logs == nil {
			logs = []*utils.PingMonitorLog{}
		}
		respondJSON(w, http.StatusOK, logs)
		return
	}

	switch req.Method {
	case http.MethodGet:
		m, err := r.monitorDB.GetPingMonitorByID(r.ctx, id)
		if err != nil {
			respondError(w, http.StatusNotFound, "monitor not found")
			return
		}
		respondJSON(w, http.StatusOK, m)

	case http.MethodPut:
		var body utils.CreatePingMonitorRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if body.CheckInterval <= 0 {
			body.CheckInterval = 60
		}
		if body.LatencyThreshold <= 0 {
			body.LatencyThreshold = 200
		}
		if body.Timeout <= 0 {
			body.Timeout = 3
		}

		m := &utils.PingMonitor{
			ID:               id,
			FriendlyName:     body.FriendlyName,
			Hostname:         body.Hostname,
			CheckInterval:    body.CheckInterval,
			LatencyThreshold: body.LatencyThreshold,
			Timeout:          body.Timeout,
		}
		if err := r.monitorDB.UpdatePingMonitor(r.ctx, m); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		r.pingMonitor.StopMonitor(id)
		r.pingMonitor.StartMonitor(m)

		if payload, err := json.Marshal(map[string]interface{}{
			"type": "ping_monitor_updated", "monitor": m,
		}); err == nil {
			r.wsHub.Broadcast <- payload
		}

		respondJSON(w, http.StatusOK, m)

	case http.MethodDelete:
		m, err := r.monitorDB.GetPingMonitorByID(r.ctx, id)
		if err != nil {
			respondError(w, http.StatusNotFound, "monitor not found")
			return
		}

		r.pingMonitor.StopMonitor(id)

		if err := r.monitorDB.DeletePingMonitor(r.ctx, id); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if payload, err := json.Marshal(map[string]interface{}{
			"type": "ping_monitor_deleted", "monitor_id": id, "name": m.FriendlyName,
		}); err == nil {
			r.wsHub.Broadcast <- payload
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "deleted", "monitor_id": id, "name": m.FriendlyName,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) testPingConnection(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Hostname string `json:"hostname"`
		Timeout  int    `json:"timeout"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if body.Hostname == "" {
		respondError(w, http.StatusBadRequest, "hostname required")
		return
	}
	if body.Timeout <= 0 {
		body.Timeout = 3
	}

	latencyMs, err := r.pingMonitor.PingOnce(body.Hostname, body.Timeout)
	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success":  false,
			"hostname": body.Hostname,
			"error":    err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"hostname":   body.Hostname,
		"latency_ms": latencyMs,
	})
}

func (r *Router) handlePingWebSocket(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Printf("Ping WebSocket upgrade error: %v", err)
		return
	}

	// Send initial state on connect
	monitors, err := r.monitorDB.GetAllPingMonitors(r.ctx)
	if err == nil {
		if monitors == nil {
			monitors = []*utils.PingMonitor{}
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"type": "initial_state", "monitors": monitors,
		})
		conn.WriteMessage(websocket.TextMessage, payload)
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
func (r *Router) handleSNMPMonitors(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case http.MethodGet:
		monitors, err := r.monitorDB.GetAllSNMPMonitors(r.ctx)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if monitors == nil {
			monitors = []*utils.SNMPMonitor{}
		}
		respondJSON(w, http.StatusOK, monitors)

	case http.MethodPost:
		var body utils.CreateSNMPMonitorRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if body.FriendlyName == "" || body.Hostname == "" || body.OID == "" {
			respondError(w, http.StatusBadRequest, "friendly_name, hostname and oid required")
			return
		}

		// Defaults
		if body.Port == 0 {
			body.Port = 161
		}
		if body.CommunityString == "" {
			body.CommunityString = "public"
		}
		if body.SNMPVersion == "" {
			body.SNMPVersion = "v2c"
		}
		if body.PollingInterval <= 0 {
			body.PollingInterval = 60
		}
		if body.Timeout <= 0 {
			body.Timeout = 5
		}
		if body.ExpectedValueType == "" {
			body.ExpectedValueType = "Integer"
		}

		m := &utils.SNMPMonitor{
			FriendlyName:      body.FriendlyName,
			Hostname:          body.Hostname,
			Port:              body.Port,
			CommunityString:   body.CommunityString,
			OID:               body.OID,
			SNMPVersion:       body.SNMPVersion,
			PollingInterval:   body.PollingInterval,
			Timeout:           body.Timeout,
			Retries:           body.Retries,
			ExpectedValueType: body.ExpectedValueType,
		}
		if err := r.monitorDB.CreateSNMPMonitor(r.ctx, m); err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				respondError(w, http.StatusConflict, "monitor for this hostname:port:oid already exists")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		r.snmpMonitor.StartMonitor(m)

		if payload, err := json.Marshal(map[string]interface{}{
			"type": "snmp_monitor_created", "monitor": m,
		}); err == nil {
			r.wsHub.Broadcast <- payload
		}

		respondJSON(w, http.StatusCreated, m)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleSNMPMonitorWithID(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(req.URL.Path, "/v1/api/snmp/")
	parts := strings.Split(path, "/")

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid monitor ID")
		return
	}
	if len(parts) > 1 && parts[1] == "logs" {
		limit := 50
		if s := req.URL.Query().Get("limit"); s != "" {
			if l, err := strconv.Atoi(s); err == nil && l > 0 {
				limit = l
			}
		}
		logs, err := r.monitorDB.GetSNMPMonitorLogs(r.ctx, id, limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if logs == nil {
			logs = []*utils.SNMPMonitorLog{}
		}
		respondJSON(w, http.StatusOK, logs)
		return
	}

	switch req.Method {
	case http.MethodGet:
		m, err := r.monitorDB.GetSNMPMonitorByID(r.ctx, id)
		if err != nil {
			respondError(w, http.StatusNotFound, "monitor not found")
			return
		}
		respondJSON(w, http.StatusOK, m)

	case http.MethodPut:
		var body utils.CreateSNMPMonitorRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if body.Port == 0 {
			body.Port = 161
		}
		if body.PollingInterval <= 0 {
			body.PollingInterval = 60
		}
		if body.Timeout <= 0 {
			body.Timeout = 5
		}

		m := &utils.SNMPMonitor{
			ID:                id,
			FriendlyName:      body.FriendlyName,
			Hostname:          body.Hostname,
			Port:              body.Port,
			CommunityString:   body.CommunityString,
			OID:               body.OID,
			SNMPVersion:       body.SNMPVersion,
			PollingInterval:   body.PollingInterval,
			Timeout:           body.Timeout,
			Retries:           body.Retries,
			ExpectedValueType: body.ExpectedValueType,
		}
		if err := r.monitorDB.UpdateSNMPMonitor(r.ctx, m); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		r.snmpMonitor.StopMonitor(id)
		r.snmpMonitor.StartMonitor(m)

		if payload, err := json.Marshal(map[string]interface{}{
			"type": "snmp_monitor_updated", "monitor": m,
		}); err == nil {
			r.wsHub.Broadcast <- payload
		}

		respondJSON(w, http.StatusOK, m)

	case http.MethodDelete:
		m, err := r.monitorDB.GetSNMPMonitorByID(r.ctx, id)
		if err != nil {
			respondError(w, http.StatusNotFound, "monitor not found")
			return
		}

		r.snmpMonitor.StopMonitor(id)

		if err := r.monitorDB.DeleteSNMPMonitor(r.ctx, id); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if payload, err := json.Marshal(map[string]interface{}{
			"type": "snmp_monitor_deleted", "monitor_id": id, "name": m.FriendlyName,
		}); err == nil {
			r.wsHub.Broadcast <- payload
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "deleted", "monitor_id": id, "name": m.FriendlyName,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) testSNMPConnection(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body utils.CreateSNMPMonitorRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if body.Port == 0 {
		body.Port = 161
	}
	if body.CommunityString == "" {
		body.CommunityString = "public"
	}
	if body.SNMPVersion == "" {
		body.SNMPVersion = "v2c"
	}
	if body.Timeout <= 0 {
		body.Timeout = 5
	}
	if body.ExpectedValueType == "" {
		body.ExpectedValueType = "Integer"
	}

	m := &utils.SNMPMonitor{
		Hostname:          body.Hostname,
		Port:              body.Port,
		CommunityString:   body.CommunityString,
		OID:               body.OID,
		SNMPVersion:       body.SNMPVersion,
		Timeout:           body.Timeout,
		ExpectedValueType: body.ExpectedValueType,
	}

	value, responseMs, err := r.snmpMonitor.PollOnce(m)
	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success":     false,
			"error":       err.Error(),
			"hostname":    body.Hostname,
			"port":        body.Port,
			"oid":         body.OID,
			"response_ms": 0,
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"value":        value,
		"hostname":     body.Hostname,
		"port":         body.Port,
		"oid":          body.OID,
		"snmp_version": body.SNMPVersion,
		"response_ms":  responseMs,
	})
}

func (r *Router) handleSNMPWebSocket(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Printf("SNMP WebSocket upgrade error: %v", err)
		return
	}

	// Send initial state on connect
	monitors, err := r.monitorDB.GetAllSNMPMonitors(r.ctx)
	if err == nil {
		if monitors == nil {
			monitors = []*utils.SNMPMonitor{}
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"type": "initial_state", "monitors": monitors,
		})
		conn.WriteMessage(websocket.TextMessage, payload)
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
func (r *Router) testPortConnection(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Hostname string `json:"hostname"`
		Port     int    `json:"port"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if body.Hostname == "" || body.Port == 0 {
		respondError(w, http.StatusBadRequest, "hostname and port required")
		return
	}

	resp := bgservices.TcpCheck(utils.TCPCheckRequest{
		Host:    body.Hostname,
		Port:    body.Port,
		Timeout: 10,
	})

	if resp.Success {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success":       true,
			"status":        resp.Status,
			"message":       resp.Message,
			"host":          resp.Host,
			"port":          resp.Port,
			"response_time": resp.ResponseTime,
		})
	} else {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success":       false,
			"status":        resp.Status,
			"message":       resp.Message,
			"host":          resp.Host,
			"port":          resp.Port,
			"response_time": resp.ResponseTime,
		})
	}
}
func (r *Router) handleMonitorWebSocket(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Printf("Monitor WebSocket upgrade error: %v", err)
		return
	}

	//  Send current state immediately on connect
	monitors, err := r.monitorDB.GetAllPortMonitors(r.ctx) //  was r.GetAllPortMonitors
	if err == nil {
		if monitors == nil {
			monitors = []*utils.PortMonitor{}
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"type":     "initial_state",
			"monitors": monitors,
		})
		conn.WriteMessage(websocket.TextMessage, payload)
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

func (r *Router) handlePortMonitors(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case http.MethodGet:
		monitors, err := r.monitorDB.GetAllPortMonitors(r.ctx) //
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if monitors == nil {
			monitors = []*utils.PortMonitor{}
		}
		respondJSON(w, http.StatusOK, monitors)

	case http.MethodPost:
		var body utils.CreatePortMonitorRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if body.FriendlyName == "" || body.Hostname == "" || body.Port == 0 {
			respondError(w, http.StatusBadRequest,
				"friendly_name, hostname and port required")
			return
		}
		if body.HeartbeatInterval <= 0 {
			body.HeartbeatInterval = 60
		}
		if body.HeartbeatRetryInterval <= 0 {
			body.HeartbeatRetryInterval = 60
		}

		m := &utils.PortMonitor{
			FriendlyName:           body.FriendlyName,
			Hostname:               body.Hostname,
			Port:                   body.Port,
			HeartbeatInterval:      body.HeartbeatInterval,
			Retries:                body.Retries,
			HeartbeatRetryInterval: body.HeartbeatRetryInterval,
		}
		if err := r.monitorDB.CreatePortMonitor(r.ctx, m); err != nil { //
			if strings.Contains(err.Error(), "duplicate") {
				respondError(w, http.StatusConflict,
					"monitor for this hostname:port already exists")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		r.portMonitor.StartMonitor(m)

		if payload, err := json.Marshal(map[string]interface{}{
			"type":    "monitor_created",
			"monitor": m,
		}); err == nil {
			r.wsHub.Broadcast <- payload
		}

		respondJSON(w, http.StatusCreated, m)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) handlePortMonitorWithID(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(req.URL.Path, "/v1/api/monitors/")
	parts := strings.Split(path, "/")

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid monitor ID")
		return
	}

	// GET /v1/api/monitors/1/logs
	if len(parts) > 1 && parts[1] == "logs" {
		limit := 50
		if s := req.URL.Query().Get("limit"); s != "" {
			if l, err := strconv.Atoi(s); err == nil && l > 0 {
				limit = l
			}
		}
		logs, err := r.monitorDB.GetPortMonitorLogs(r.ctx, id, limit) //
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if logs == nil {
			logs = []*utils.PortMonitorLog{}
		}
		respondJSON(w, http.StatusOK, logs)
		return
	}

	switch req.Method {
	case http.MethodGet:
		m, err := r.monitorDB.GetPortMonitorByID(r.ctx, id) //
		if err != nil {
			respondError(w, http.StatusNotFound, "monitor not found")
			return
		}
		respondJSON(w, http.StatusOK, m)

	case http.MethodPut:
		var body utils.CreatePortMonitorRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if body.HeartbeatInterval <= 0 {
			body.HeartbeatInterval = 60
		}
		if body.HeartbeatRetryInterval <= 0 {
			body.HeartbeatRetryInterval = 60
		}

		m := &utils.PortMonitor{
			ID:                     id,
			FriendlyName:           body.FriendlyName,
			Hostname:               body.Hostname,
			Port:                   body.Port,
			HeartbeatInterval:      body.HeartbeatInterval,
			Retries:                body.Retries,
			HeartbeatRetryInterval: body.HeartbeatRetryInterval,
		}
		if err := r.monitorDB.UpdatePortMonitor(r.ctx, m); err != nil { //
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		r.portMonitor.StopMonitor(id)
		r.portMonitor.StartMonitor(m)

		if payload, err := json.Marshal(map[string]interface{}{
			"type":    "monitor_updated",
			"monitor": m,
		}); err == nil {
			r.wsHub.Broadcast <- payload
		}

		respondJSON(w, http.StatusOK, m)

	case http.MethodDelete:
		m, err := r.monitorDB.GetPortMonitorByID(r.ctx, id) //
		if err != nil {
			respondError(w, http.StatusNotFound, "monitor not found")
			return
		}

		r.portMonitor.StopMonitor(id)

		if err := r.monitorDB.DeletePortMonitor(r.ctx, id); err != nil { //
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if payload, err := json.Marshal(map[string]interface{}{
			"type":       "monitor_deleted",
			"monitor_id": id,
			"name":       m.FriendlyName,
		}); err == nil {
			r.wsHub.Broadcast <- payload
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":     "deleted",
			"monitor_id": id,
			"name":       m.FriendlyName,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func (r *Router) handleInterfaceAction(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(req.URL.Path, "/v1/api/interfaces/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	interfaceName := parts[0]
	action := parts[1]

	if action != "monitor" && action != "unmonitor" {
		http.Error(w, "Invalid action (use 'monitor' or 'unmonitor')", http.StatusBadRequest)
		return
	}

	if req.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	detector := vlan.NewInterfaceDetector()
	interfaces, err := detector.GetAllInterfaces()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var targetInterface *vlan.DetectedInterface
	for _, iface := range interfaces {
		if iface.Name == interfaceName {
			targetInterface = &iface
			break
		}
	}

	if targetInterface == nil {
		respondError(w, http.StatusNotFound, "Interface not found")
		return
	}

	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	if action == "monitor" {
		//  Parse optional body
		var monitorReq MonitorRequest
		if req.Body != nil {
			json.NewDecoder(req.Body).Decode(&monitorReq) // all fields optional
		}

		vlanID := 0
		if targetInterface.VLANId != nil {
			vlanID = *targetInterface.VLANId
		}

		//  Name priority: user provided > auto-generated
		networkName := monitorReq.Name
		if networkName == "" {
			if targetInterface.IsVLAN {
				networkName = fmt.Sprintf("VLAN %d (%s)", vlanID, interfaceName)
			} else {
				networkName = fmt.Sprintf("%s Network", interfaceName)
			}
		}

		//  Scan interval priority: user provided > default
		scanInterval := monitorReq.ScanInterval
		if scanInterval <= 0 {
			scanInterval = 30
		}
		ipAddr := targetInterface.IPv4
		cidr := fmt.Sprintf("/%d", targetInterface.CIDR)
		cidrFull := ipAddr + cidr
		gateway := targetInterface.DefaultGateway
		vlanConfig := &utils.VLANNetwork{
			VLANId:              vlanID,
			InterfaceName:       interfaceName,
			VLANName:            networkName,
			NetworkMode:         "auto",
			IPAddress:           &ipAddr,
			CIDRNotation:        &cidr,
			CIDRFull:            &cidrFull,
			DefaultGateway:      &gateway,
			MonitoringEnabled:   true,
			ScanIntervalSeconds: scanInterval,
		}

		existing, _ := r.db.GetVLANNetworkByInterface(ctx, interfaceName)
		if existing != nil {
			existing.MonitoringEnabled = true
			//  Only override name/interval if user explicitly provided them
			if monitorReq.Name != "" {
				existing.VLANName = monitorReq.Name
			}
			if monitorReq.ScanInterval > 0 {
				existing.ScanIntervalSeconds = monitorReq.ScanInterval
			}
			if err := r.db.UpdateVLANNetworkByInterface(ctx, existing, interfaceName); err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			vlanConfig = existing
		} else {
			if err := r.db.CreateVLANNetworkByInterface(ctx, vlanConfig, interfaceName); err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		log.Printf("[Monitor] Starting scan for %s with DB id=%d name=%s",
			interfaceName, vlanConfig.ID, vlanConfig.VLANName)

		if err := r.scanManager.StartVLANScan(vlanConfig); err != nil {
			respondError(w, http.StatusInternalServerError,
				fmt.Sprintf("Failed to start monitoring: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":       "monitoring_started",
			"interface":    interfaceName,
			"network_id":   vlanConfig.ID,       //  frontend uses to fetch devices
			"network_name": vlanConfig.VLANName, //  frontend displays this
			"interface_type": func() string {
				if targetInterface.IsVLAN {
					return "vlan"
				}
				return "physical"
			}(),
			"cidr":          cidrFull,
			"scan_interval": vlanConfig.ScanIntervalSeconds,
		})

	} else if action == "unmonitor" {
		existing, err := r.db.GetVLANNetworkByInterface(ctx, interfaceName)
		if err != nil {
			respondError(w, http.StatusNotFound, "Network not found in database")
			return
		}
		existing.MonitoringEnabled = false
		if err := r.db.UpdateVLANNetworkByInterface(ctx, existing, interfaceName); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		r.scanManager.StopVLANScanByInterface(existing.InterfaceName)
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":       "monitoring_stopped",
			"interface":    interfaceName,
			"network_id":   existing.ID,
			"network_name": existing.VLANName,
		})
	}
}
func (r *Router) getInterfaces(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	if req.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	detector := vlan.NewInterfaceDetector()
	interfaces, err := detector.GetAllInterfaces()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get monitored networks from database
	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	vlans, _ := r.db.GetAllVLANs(ctx)

	//  Build map by interface_name instead of constructed name
	vlanMap := make(map[string]*utils.VLANNetwork)
	for _, v := range vlans {
		if v.InterfaceName != "" {
			vlanMap[v.InterfaceName] = v
		} else if v.VLANId > 0 {
			// Backward compatibility: construct name for old entries
			vlanMap[fmt.Sprintf("%s.%d", r.scanManager.GetParentInterface(), v.VLANId)] = v
		}
	}

	// Build response with monitoring status
	type InterfaceResponse struct {
		vlan.DetectedInterface
		IsMonitored  bool   `json:"is_monitored"`
		NetworkDBId  *int   `json:"network_db_id,omitempty"`
		NetworkName  string `json:"network_name,omitempty"`
		ScanInterval int    `json:"scan_interval,omitempty"`
	}

	var response []InterfaceResponse
	for _, iface := range interfaces {
		resp := InterfaceResponse{
			DetectedInterface: iface,
			IsMonitored:       false,
		}

		//  Check by interface name
		if v, exists := vlanMap[iface.Name]; exists && v.MonitoringEnabled {
			resp.IsMonitored = true
			resp.NetworkDBId = &v.ID
			resp.NetworkName = v.VLANName
			resp.ScanInterval = v.ScanIntervalSeconds
		}

		response = append(response, resp)
	}

	respondJSON(w, http.StatusOK, response)
}

// func (r *Router) getInterfaces(w http.ResponseWriter, req *http.Request) {
// 	if EnableCORS(&w, req) {
// 		return
// 	}

// 	if req.Method != http.MethodGet {
// 		http.Error(w, "GET only", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")

// 	detector := vlan.NewInterfaceDetector()
// 	interfaces, err := detector.GetAllInterfaces()
// 	if err != nil {
// 		respondError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	// Enrich with monitoring status from database
// 	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
// 	defer cancel()

// 	vlans, _ := r.db.GetAllVLANs(ctx)
// 	vlanMap := make(map[string]*utils.VLANNetwork)
// 	for _, v := range vlans {
// 		vlanMap[fmt.Sprintf("%s.%d", r.scanManager.GetParentInterface(), v.VLANId)] = v
// 	}

// 	// Build response with monitoring status
// 	type InterfaceResponse struct {
// 		vlan.DetectedInterface
// 		IsMonitored bool   `json:"is_monitored"`
// 		VLANDBId    *int   `json:"vlan_db_id,omitempty"`
// 		VLANName    string `json:"vlan_name,omitempty"`
// 	}

// 	var response []InterfaceResponse
// 	for _, iface := range interfaces {
// 		resp := InterfaceResponse{
// 			DetectedInterface: iface,
// 			IsMonitored:       false,
// 		}

// 		// Check if this interface is being monitored
// 		if v, exists := vlanMap[iface.Name]; exists && v.MonitoringEnabled {
// 			resp.IsMonitored = true
// 			resp.VLANDBId = &v.VLANId
// 			resp.VLANName = v.VLANName
// 		}

// 		response = append(response, resp)
// 	}

// 	respondJSON(w, http.StatusOK, response)
// }

// func (r *Router) handleInterfaceAction(w http.ResponseWriter, req *http.Request) {
// 	if EnableCORS(&w, req) {
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")

// 	// Extract interface name from path
// 	path := strings.TrimPrefix(req.URL.Path, "/v1/api/interfaces/")
// 	parts := strings.Split(path, "/")
// 	if len(parts) < 2 {
// 		http.Error(w, "Invalid path", http.StatusBadRequest)
// 		return
// 	}

// 	interfaceName := parts[0]
// 	action := parts[1]

// 	if action != "monitor" && action != "unmonitor" {
// 		http.Error(w, "Invalid action (use 'monitor' or 'unmonitor')", http.StatusBadRequest)
// 		return
// 	}

// 	if req.Method != http.MethodPost {
// 		http.Error(w, "POST only", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Get interface details
// 	detector := vlan.NewInterfaceDetector()
// 	interfaces, err := detector.GetAllInterfaces()
// 	if err != nil {
// 		respondError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	var targetInterface *vlan.DetectedInterface
// 	for _, iface := range interfaces {
// 		if iface.Name == interfaceName {
// 			targetInterface = &iface
// 			break
// 		}
// 	}

// 	if targetInterface == nil {
// 		respondError(w, http.StatusNotFound, "Interface not found")
// 		return
// 	}

// 	if action == "monitor" {
// 		// Start monitoring this interface
// 		vlanID := 0
// 		if targetInterface.VLANId != nil {
// 			vlanID = *targetInterface.VLANId
// 		} else {
// 			// Auto-generate VLAN ID for non-VLAN interfaces
// 			// Use last octet of IP or generate unique ID
// 			respondError(w, http.StatusBadRequest, "Only VLAN interfaces can be monitored")
// 			return
// 		}

// 		// Create VLAN config
// 		vlanConfig := &utils.VLANNetwork{
// 			VLANId:              vlanID,
// 			VLANName:            fmt.Sprintf("VLAN %d (%s)", vlanID, interfaceName),
// 			NetworkMode:         "dhcp", // Auto-detected interfaces use their existing config
// 			MonitoringEnabled:   true,
// 			ScanIntervalSeconds: 30,
// 		}

// 		// Check if already exists in database
// 		ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
// 		defer cancel()

// 		existing, _ := r.db.GetVLANNetwork(ctx, vlanID)
// 		if existing != nil {
// 			// Update to enable monitoring
// 			existing.MonitoringEnabled = true
// 			if err := r.db.UpdateVLANNetwork(ctx, existing); err != nil {
// 				respondError(w, http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 			vlanConfig = existing
// 		} else {
// 			// Create new VLAN entry
// 			ipAddr := targetInterface.IPv4
// 			cidr := fmt.Sprintf("/%d", targetInterface.CIDR)
// 			cidrFull := ipAddr + cidr
// 			gateway := targetInterface.DefaultGateway

// 			vlanConfig.IPAddress = &ipAddr
// 			vlanConfig.CIDRNotation = &cidr
// 			vlanConfig.CIDRFull = &cidrFull
// 			vlanConfig.DefaultGateway = &gateway

// 			if err := r.db.CreateVLANNetwork(ctx, vlanConfig); err != nil {
// 				respondError(w, http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 		}

// 		// Start scanning
// 		if err := r.scanManager.StartVLANScan(vlanConfig); err != nil {
// 			respondError(w, http.StatusInternalServerError,
// 				fmt.Sprintf("Failed to start monitoring: %v", err))
// 			return
// 		}

// 		respondJSON(w, http.StatusOK, map[string]interface{}{
// 			"status":    "monitoring_started",
// 			"interface": interfaceName,
// 			"vlan_id":   vlanID,
// 		})

// 	} else if action == "unmonitor" {
// 		// Stop monitoring
// 		if targetInterface.VLANId == nil {
// 			respondError(w, http.StatusBadRequest, "Not a VLAN interface")
// 			return
// 		}

// 		vlanID := *targetInterface.VLANId

// 		// Update database to disable monitoring
// 		ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
// 		defer cancel()

// 		existing, err := r.db.GetVLANNetwork(ctx, vlanID)
// 		if err != nil {
// 			respondError(w, http.StatusNotFound, "VLAN not found in database")
// 			return
// 		}

// 		existing.MonitoringEnabled = false
// 		if err := r.db.UpdateVLANNetwork(ctx, existing); err != nil {
// 			respondError(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		// Stop scanner
// 		r.scanManager.StopVLANScan(vlanID)

//			respondJSON(w, http.StatusOK, map[string]interface{}{
//				"status":    "monitoring_stopped",
//				"interface": interfaceName,
//				"vlan_id":   vlanID,
//			})
//		}
//	}
// func (r *Router) handleInterfaceAction(w http.ResponseWriter, req *http.Request) {
// 	if EnableCORS(&w, req) {
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")

// 	// Extract interface name from path
// 	path := strings.TrimPrefix(req.URL.Path, "/v1/api/interfaces/")
// 	parts := strings.Split(path, "/")
// 	if len(parts) < 2 {
// 		http.Error(w, "Invalid path", http.StatusBadRequest)
// 		return
// 	}

// 	interfaceName := parts[0]
// 	action := parts[1]

// 	if action != "monitor" && action != "unmonitor" {
// 		http.Error(w, "Invalid action (use 'monitor' or 'unmonitor')", http.StatusBadRequest)
// 		return
// 	}

// 	if req.Method != http.MethodPost {
// 		http.Error(w, "POST only", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Get interface details
// 	detector := vlan.NewInterfaceDetector()
// 	interfaces, err := detector.GetAllInterfaces()
// 	if err != nil {
// 		respondError(w, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	var targetInterface *vlan.DetectedInterface
// 	for _, iface := range interfaces {
// 		if iface.Name == interfaceName {
// 			targetInterface = &iface
// 			break
// 		}
// 	}

// 	if targetInterface == nil {
// 		respondError(w, http.StatusNotFound, "Interface not found")
// 		return
// 	}

// 	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
// 	defer cancel()

// 	if action == "monitor" {
// 		//  Allow monitoring ANY interface type (physical or VLAN)

// 		// Determine VLAN ID (0 for non-VLAN interfaces)
// 		vlanID := 0
// 		if targetInterface.VLANId != nil {
// 			vlanID = *targetInterface.VLANId
// 		}

// 		// Create network config
// 		ipAddr := targetInterface.IPv4
// 		cidr := fmt.Sprintf("/%d", targetInterface.CIDR)
// 		cidrFull := ipAddr + cidr
// 		gateway := targetInterface.DefaultGateway

// 		networkName := fmt.Sprintf("%s Network", interfaceName)
// 		if targetInterface.IsVLAN {
// 			networkName = fmt.Sprintf("VLAN %d (%s)", vlanID, interfaceName)
// 		}

// 		vlanConfig := &utils.VLANNetwork{
// 			VLANId:              vlanID,
// 			InterfaceName:       interfaceName, //  Set interface name
// 			VLANName:            networkName,
// 			NetworkMode:         "auto", // Auto-detected
// 			IPAddress:           &ipAddr,
// 			CIDRNotation:        &cidr,
// 			CIDRFull:            &cidrFull,
// 			DefaultGateway:      &gateway,
// 			MonitoringEnabled:   true,
// 			ScanIntervalSeconds: 30,
// 		}

// 		// Check if already exists in database (by interface name)
// 		existing, _ := r.db.GetVLANNetworkByInterface(ctx, interfaceName)
// 		if existing != nil {
// 			// Update to enable monitoring
// 			existing.MonitoringEnabled = true
// 			if err := r.db.UpdateVLANNetworkByInterface(ctx, existing, interfaceName); err != nil {
// 				respondError(w, http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 			vlanConfig = existing
// 		} else {
// 			// Create new entry
// 			if err := r.db.CreateVLANNetworkByInterface(ctx, vlanConfig, interfaceName); err != nil {
// 				respondError(w, http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 		}

// 		// Start scanning
// 		if err := r.scanManager.StartVLANScan(vlanConfig); err != nil {
// 			respondError(w, http.StatusInternalServerError,
// 				fmt.Sprintf("Failed to start monitoring: %v", err))
// 			return
// 		}

// 		respondJSON(w, http.StatusOK, map[string]interface{}{
// 			"status":    "monitoring_started",
// 			"interface": interfaceName,
// 			"interface_type": func() string {
// 				if targetInterface.IsVLAN {
// 					return "vlan"
// 				}
// 				return "physical"
// 			}(),
// 			"vlan_id": vlanID,
// 		})

// 	} else if action == "unmonitor" {
// 		// Stop monitoring (works for any interface type)
// 		existing, err := r.db.GetVLANNetworkByInterface(ctx, interfaceName)
// 		if err != nil {
// 			respondError(w, http.StatusNotFound, "Network not found in database")
// 			return
// 		}

// 		existing.MonitoringEnabled = false
// 		if err := r.db.UpdateVLANNetworkByInterface(ctx, existing, interfaceName); err != nil {
// 			respondError(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		// Stop scanner (by VLAN ID if it's a VLAN, otherwise by interface name)
// 		if existing.VLANId > 0 {
// 			r.scanManager.StopVLANScan(existing.VLANId)
// 		}

//			respondJSON(w, http.StatusOK, map[string]interface{}{
//				"status":    "monitoring_stopped",
//				"interface": interfaceName,
//			})
//		}
//	}
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
        network_id,
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
		NetworkId    int       `json:"network_id"`
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
		if err := rows.Scan(&c.NetworkId, &c.IPAddress, &c.MACAddress,
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

	path := strings.TrimPrefix(req.URL.Path, "/v1/api/vlans/")
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
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	vlans, err := r.db.GetAllVLANs(r.ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, vlans)
}

func (r *Router) getVLAN(w http.ResponseWriter, req *http.Request, id int) {
	if EnableCORS(&w, req) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	v, err := r.db.GetVLANNetwork(r.ctx, id)
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

func (r *Router) updateVLAN(w http.ResponseWriter, req *http.Request, id int) {
	w.Header().Set("Content-Type", "application/json")

	var vlanReq CreateVLANRequest
	if err := json.NewDecoder(req.Body).Decode(&vlanReq); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	existing, err := r.db.GetNetworkByID(r.ctx, id) //  lookup by id
	if err != nil {
		respondError(w, http.StatusNotFound, "Network not found")
		return
	}

	scanInterval := vlanReq.ScanIntervalSeconds
	if scanInterval <= 0 {
		scanInterval = existing.ScanIntervalSeconds
		if scanInterval <= 0 {
			scanInterval = defaultScanIntervalSeconds
		}
	}

	existing.VLANName = vlanReq.VLANName
	existing.NetworkMode = vlanReq.NetworkMode
	existing.MonitoringEnabled = vlanReq.EnableMonitoring
	existing.ScanIntervalSeconds = scanInterval

	if vlanReq.NetworkMode == "static" {
		cidrFull, err := calculateCIDRFull(vlanReq.IPAddress, vlanReq.CIDRNotation)
		if err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid IP/CIDR: %v", err))
			return
		}
		existing.IPAddress = &vlanReq.IPAddress
		existing.CIDRNotation = &vlanReq.CIDRNotation
		existing.CIDRFull = &cidrFull
		if vlanReq.DefaultGateway != "" {
			existing.DefaultGateway = &vlanReq.DefaultGateway
		}
	}

	if err := r.db.UpdateNetwork(r.ctx, existing); err != nil { //  new UpdateNetwork
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	r.scanManager.StopVLANScan(existing.VLANId)

	if existing.MonitoringEnabled {
		if err := r.scanManager.StartVLANScan(existing); err != nil {
			log.Printf("[Network %d] Updated but failed to start: %v", id, err)
		}
	}

	respondJSON(w, http.StatusOK, existing)
}

// deleteVLAN — use interface_name based delete
func (r *Router) deleteVLAN(w http.ResponseWriter, req *http.Request, id int) {
	w.Header().Set("Content-Type", "application/json")

	existing, err := r.db.GetNetworkByID(r.ctx, id) //  get first to find interface_name
	if err != nil {
		respondError(w, http.StatusNotFound, "Network not found")
		return
	}

	r.scanManager.StopVLANScan(existing.VLANId)

	if err := r.db.DeleteNetwork(r.ctx, existing.InterfaceName); //  delete by interface_name
	err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "deleted",
		"id":      strconv.Itoa(id),
		"message": "Network deleted successfully",
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
	if EnableCORS(&w, req) {
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
	if EnableCORS(&w, req) {
		return
	}
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
