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
	"github.com/mascarenhasmelson/gomotz/utils"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Router struct {
	ctx  context.Context
	pool *pgxpool.Pool
	mux  *http.ServeMux
}

func NewRouter(ctx context.Context, pool *pgxpool.Pool) *http.ServeMux {
	r := &Router{
		ctx:  ctx,
		pool: pool,
		mux:  http.NewServeMux(),
	}

	r.routes()
	return r.mux
}

var wsMu sync.Mutex

func (r *Router) routes() {
	r.mux.HandleFunc("/services", r.servicesHandler)
	r.mux.HandleFunc("/services/", r.serviceHandler)
	r.mux.HandleFunc("/v1/services/isp", r.ispHandler)
	r.mux.HandleFunc("/v1/scan", handleSynScan)
	r.mux.HandleFunc("/v1/tcpCheck", TcpCheckHandler)
	r.mux.HandleFunc("/v1/dnsCheck", r.dnsQueryHandler)
	r.mux.HandleFunc("/v1/traceroute", handleTracerouteWebSocket)
	r.mux.HandleFunc("/v1/icmp", PingHandler)
	r.mux.HandleFunc("/v1/httpsCheck", handleHTTPS)
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	if EnableCORS(&w, r) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	if r.Method != "POST" {
		http.Error(w, "POST only", 405)
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

	// Use timeout from request or default to 10 seconds
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
		json.NewEncoder(w).Encode(utils.HTTPSCheckResponse{
			HTTPSSupported: false,
			Error:          err.Error(),
		})
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

	// Send connection confirmation
	conn.WriteJSON(utils.PingMessage{Type: "connected", Message: "WebSocket connected"})

	for {
		// Read message from client
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		var req utils.PingRequest
		if err := json.Unmarshal(message, &req); err != nil {
			conn.WriteJSON(utils.PingMessage{Type: "error", Message: "Invalid request format"})
			continue
		}

		// Validate and set defaults
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

		// Perform ping
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

	// Read traceroute request
	var req utils.TracerouteRequest
	err = conn.ReadJSON(&req)
	if err != nil {
		log.Println("Error reading traceroute request:", err)
		bgservices.SendTracerouteError(conn, "Invalid request format")
		return
	}

	// Validate request
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

	// Resolve target IP
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

	// Send target IP to client
	bgservices.SendTracerouteMessage(conn, utils.TracerouteMessage{
		Type:     "targetIp",
		TargetIP: targetIP.String(),
	})

	// Start traceroute based on protocol
	switch req.Protocol {
	case "ICMP":
		bgservices.PerformICMPTraceroute(conn, targetIP, req)
	// case "UDP":
	// 	performUDPTraceroute(conn, targetIP, req)
	// case "TCP":
	// 	performTCPTraceroute(conn, targetIP, req)
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
