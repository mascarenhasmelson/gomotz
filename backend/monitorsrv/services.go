package monitorsrv

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mascarenhasmelson/gomotz/bgservices"
	"github.com/mascarenhasmelson/gomotz/utils"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type SSLMonitorService struct {
	db        *PostgresDB
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	monitors  map[int]*sslWorker
	wg        sync.WaitGroup
	broadcast BroadcastFunc
}

type sslWorker struct {
	monitor *utils.SSLMonitor
	cancel  context.CancelFunc
}

type BroadcastFunc func(payload []byte)

type PortMonitorService struct {
	db        *PostgresDB
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	monitors  map[int]*portMonitorWorker
	wg        sync.WaitGroup
	broadcast BroadcastFunc //  instead of *vlan.Hub
}

type portMonitorWorker struct {
	monitor *utils.PortMonitor
	cancel  context.CancelFunc
}

type PingMonitorService struct {
	db        *PostgresDB
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	monitors  map[int]*pingWorker
	wg        sync.WaitGroup
	broadcast BroadcastFunc
}

type pingWorker struct {
	monitor *utils.PingMonitor
	cancel  context.CancelFunc
}

func NewPingMonitorService(db *PostgresDB, broadcast BroadcastFunc) *PingMonitorService {
	ctx, cancel := context.WithCancel(context.Background())
	return &PingMonitorService{
		db:        db,
		ctx:       ctx,
		cancel:    cancel,
		monitors:  make(map[int]*pingWorker),
		broadcast: broadcast,
	}
}

func (s *PingMonitorService) RecoverFromDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitors, err := s.db.GetAllPingMonitors(ctx)
	if err != nil {
		return err
	}

	log.Printf("[PING MONITOR] Recovering %d monitor(s)", len(monitors))
	for _, m := range monitors {
		s.StartMonitor(m)
	}
	return nil
}

func (s *PingMonitorService) StartMonitor(m *utils.PingMonitor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w, exists := s.monitors[m.ID]; exists {
		w.cancel()
	}

	ctx, cancel := context.WithCancel(s.ctx)
	s.monitors[m.ID] = &pingWorker{monitor: m, cancel: cancel}

	s.wg.Add(1)
	go s.runWorker(ctx, m)

	log.Printf("[PING MONITOR] Started %d — %s (%s) every %ds",
		m.ID, m.FriendlyName, m.Hostname, m.CheckInterval)
}

func (s *PingMonitorService) StopMonitor(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w, exists := s.monitors[id]; exists {
		w.cancel()
		delete(s.monitors, id)
		log.Printf("[PING MONITOR] Stopped monitor %d", id)
	}
}

func (s *PingMonitorService) Shutdown() {
	s.cancel()
	s.wg.Wait()
	log.Println("[PING MONITOR] All monitors stopped")
}

func (s *PingMonitorService) GetActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.monitors)
}

func (s *PingMonitorService) runWorker(ctx context.Context, m *utils.PingMonitor) {
	defer s.wg.Done()

	// Ping immediately on start
	s.pingAndSave(m)

	ticker := time.NewTicker(time.Duration(m.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.pingAndSave(m)
		}
	}
}

func (s *PingMonitorService) pingAndSave(m *utils.PingMonitor) {
	latencyMs, err := s.doPing(m.Hostname, m.Timeout)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var status string
	var latencyPtr *int
	var errMsgPtr *string

	if err != nil {
		status = "down"
		errMsg := err.Error()
		errMsgPtr = &errMsg
		log.Printf("[PING MONITOR] %s (%s) → down [%s]",
			m.FriendlyName, m.Hostname, err.Error())
	} else {
		latencyPtr = &latencyMs
		//  warning if latency exceeds threshold
		if latencyMs > m.LatencyThreshold {
			status = "warning"
		} else {
			status = "up"
		}
		log.Printf("[PING MONITOR] %s (%s) → %s [%dms] threshold:%dms",
			m.FriendlyName, m.Hostname, status, latencyMs, m.LatencyThreshold)
	}

	if err := s.db.UpdatePingMonitorStatus(ctx, m.ID, status, latencyPtr, errMsgPtr); err != nil {
		log.Printf("[PING MONITOR] Failed to update status %d: %v", m.ID, err)
	}

	if err := s.db.InsertPingMonitorLog(ctx, &utils.PingMonitorLog{
		MonitorID:    m.ID,
		Status:       status,
		LatencyMs:    latencyPtr,
		ErrorMessage: errMsgPtr,
	}); err != nil {
		log.Printf("[PING MONITOR] Failed to insert log %d: %v", m.ID, err)
	}

	//Broadcast realtime update
	if s.broadcast != nil {
		latency := 0
		if latencyPtr != nil {
			latency = *latencyPtr
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"type":          "ping_monitor_update",
			"monitor_id":    m.ID,
			"friendly_name": m.FriendlyName,
			"hostname":      m.Hostname,
			"status":        status,
			"latency_ms":    latency,
			"threshold":     m.LatencyThreshold,
			"error": func() string {
				if errMsgPtr != nil {
					return *errMsgPtr
				}
				return ""
			}(),
			"checked_at": time.Now().UTC(),
		})
		s.broadcast(payload)
	}
}

// doPing sends a single ICMP echo and returns latency in ms
func (s *PingMonitorService) doPing(target string, timeoutSecs int) (int, error) {
	ipAddr, err := resolveTarget(target)
	if err != nil {
		return 0, err
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "")
	if err != nil {
		return 0, fmt.Errorf("listen error: %w", err)
	}
	defer conn.Close()

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("gomotz-ping"),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return 0, err
	}

	start := time.Now()

	if _, err := conn.WriteTo(msgBytes, ipAddr); err != nil {
		return 0, fmt.Errorf("write error: %w", err)
	}

	reply := make([]byte, 1500)
	deadline := time.Duration(timeoutSecs) * time.Second
	if err := conn.SetReadDeadline(time.Now().Add(deadline)); err != nil {
		return 0, err
	}

	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		return 0, fmt.Errorf("timeout or unreachable: %w", err)
	}

	latencyMs := int(time.Since(start).Milliseconds())

	parsedMsg, err := icmp.ParseMessage(1, reply[:n])
	if err != nil {
		return 0, fmt.Errorf("parse error: %w", err)
	}

	switch parsedMsg.Type {
	case ipv4.ICMPTypeEchoReply:
		return latencyMs, nil
	case ipv4.ICMPTypeDestinationUnreachable:
		return 0, fmt.Errorf("destination unreachable")
	default:
		return 0, fmt.Errorf("unexpected ICMP type: %v", parsedMsg.Type)
	}
}

// PingOnce does a single ping — used for test connection
func (s *PingMonitorService) PingOnce(hostname string, timeoutSecs int) (int, error) {
	return s.doPing(hostname, timeoutSecs)
}

func resolveTarget(target string) (*net.IPAddr, error) {
	ipAddr, err := net.ResolveIPAddr("ip4", target)
	if err == nil {
		return ipAddr, nil
	}
	ips, err := net.LookupIP(target)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("failed to resolve: %s", target)
	}
	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			return &net.IPAddr{IP: v4}, nil
		}
	}
	return nil, fmt.Errorf("no IPv4 found for: %s", target)
}

func NewPortMonitorService(db *PostgresDB, broadcast BroadcastFunc) *PortMonitorService { //
	ctx, cancel := context.WithCancel(context.Background())
	return &PortMonitorService{
		db:        db,
		ctx:       ctx,
		cancel:    cancel,
		monitors:  make(map[int]*portMonitorWorker),
		broadcast: broadcast, //
	}
}

func (s *PortMonitorService) RecoverFromDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitors, err := s.db.GetAllPortMonitors(ctx)
	if err != nil {
		return err
	}

	log.Printf("[PORT MONITOR] Recovering %d monitor(s)", len(monitors))
	for _, m := range monitors {
		s.StartMonitor(m)
	}
	return nil
}

func (s *PortMonitorService) StartMonitor(m *utils.PortMonitor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w, exists := s.monitors[m.ID]; exists {
		w.cancel()
	}

	ctx, cancel := context.WithCancel(s.ctx)
	s.monitors[m.ID] = &portMonitorWorker{monitor: m, cancel: cancel}

	s.wg.Add(1)
	go s.runWorker(ctx, m)

	log.Printf("[PORT MONITOR] Started monitor %d — %s (%s:%d) every %ds",
		m.ID, m.FriendlyName, m.Hostname, m.Port, m.HeartbeatInterval)
}

func (s *PortMonitorService) StopMonitor(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w, exists := s.monitors[id]; exists {
		w.cancel()
		delete(s.monitors, id)
		log.Printf("[PORT MONITOR] Stopped monitor %d", id)
	}
}

func (s *PortMonitorService) Shutdown() {
	s.cancel()
	s.wg.Wait()
	log.Println("[PORT MONITOR] All monitors stopped")
}

func (s *PortMonitorService) GetActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.monitors)
}

func (s *PortMonitorService) runWorker(ctx context.Context, m *utils.PortMonitor) {
	defer s.wg.Done()

	s.checkAndSave(m)

	ticker := time.NewTicker(time.Duration(m.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndSave(m)
		}
	}
}

func (s *PortMonitorService) checkAndSave(m *utils.PortMonitor) {
	resp := bgservices.TcpCheck(utils.TCPCheckRequest{
		Host:    m.Hostname,
		Port:    m.Port,
		Timeout: 10,
	})

	//  Derive status from tcp status, not just resp.Success
	status := tcpStatusToMonitorStatus(resp.Status)

	// Retries only if down
	if status == "down" && m.Retries > 0 {
		for i := 0; i < m.Retries; i++ {
			time.Sleep(time.Duration(m.HeartbeatRetryInterval) * time.Second)
			retry := bgservices.TcpCheck(utils.TCPCheckRequest{
				Host:    m.Hostname,
				Port:    m.Port,
				Timeout: 10,
			})
			retryStatus := tcpStatusToMonitorStatus(retry.Status)
			if retryStatus == "up" {
				status = "up"
				resp = retry
				break
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var responseMs *int
	if resp.ResponseTime > 0 {
		ms := int(resp.ResponseTime)
		responseMs = &ms
	}

	if err := s.db.UpdatePortMonitorStatus(ctx, m.ID, status, resp.Status, responseMs); err != nil {
		log.Printf("[PORT MONITOR] Failed to update status for %d: %v", m.ID, err)
	}

	var errMsg *string
	if resp.Message != "" && status == "down" {
		errMsg = &resp.Message
	}

	if err := s.db.InsertPortMonitorLog(ctx, &utils.PortMonitorLog{
		MonitorID:    m.ID,
		Status:       status,
		ResponseMs:   responseMs,
		ErrorMessage: errMsg,
	}); err != nil {
		log.Printf("[PORT MONITOR] Failed to insert log for %d: %v", m.ID, err)
	}

	log.Printf("[PORT MONITOR] %s (%s:%d) → %s [%s] (%dms)",
		m.FriendlyName, m.Hostname, m.Port, status, resp.Status,
		func() int {
			if responseMs != nil {
				return *responseMs
			}
			return 0
		}())

	if s.broadcast != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"type":          "port_monitor_update",
			"monitor_id":    m.ID,
			"friendly_name": m.FriendlyName,
			"hostname":      m.Hostname,
			"port":          m.Port,
			"status":        status,
			"tcp_status":    resp.Status,
			"response_ms": func() int {
				if responseMs != nil {
					return *responseMs
				}
				return 0
			}(),
			"message":    resp.Message,
			"checked_at": time.Now().UTC(),
		})
		s.broadcast(payload)
	}
}

// Single place to define the mapping
func tcpStatusToMonitorStatus(tcpStatus string) string {
	switch tcpStatus {
	case "open":
		return "up"
	case "closed", "filtered", "error":
		return "down"
	default:
		return "down"
	}
}

func NewSSLMonitorService(db *PostgresDB, broadcast BroadcastFunc) *SSLMonitorService {
	ctx, cancel := context.WithCancel(context.Background())
	return &SSLMonitorService{
		db:        db,
		ctx:       ctx,
		cancel:    cancel,
		monitors:  make(map[int]*sslWorker),
		broadcast: broadcast,
	}
}

func (s *SSLMonitorService) RecoverFromDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitors, err := s.db.GetAllSSLMonitors(ctx)
	if err != nil {
		return err
	}

	log.Printf("[SSL MONITOR] Recovering %d monitor(s)", len(monitors))
	for _, m := range monitors {
		s.StartMonitor(m)
	}
	return nil
}

func (s *SSLMonitorService) StartMonitor(m *utils.SSLMonitor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w, exists := s.monitors[m.ID]; exists {
		w.cancel()
	}

	ctx, cancel := context.WithCancel(s.ctx)
	s.monitors[m.ID] = &sslWorker{monitor: m, cancel: cancel}

	s.wg.Add(1)
	go s.runWorker(ctx, m)

	log.Printf("[SSL MONITOR] Started %d — %s every %ds",
		m.ID, m.Domain, m.CheckInterval)
}

func (s *SSLMonitorService) StopMonitor(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w, exists := s.monitors[id]; exists {
		w.cancel()
		delete(s.monitors, id)
		log.Printf("[SSL MONITOR] Stopped monitor %d", id)
	}
}

func (s *SSLMonitorService) Shutdown() {
	s.cancel()
	s.wg.Wait()
	log.Println("[SSL MONITOR] All monitors stopped")
}

func (s *SSLMonitorService) runWorker(ctx context.Context, m *utils.SSLMonitor) {
	defer s.wg.Done()

	// Check immediately on start
	s.checkAndSave(m)

	ticker := time.NewTicker(time.Duration(m.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndSave(m)
		}
	}
}

func (s *SSLMonitorService) checkAndSave(m *utils.SSLMonitor) {
	result, err := s.checkCertificate(m.Domain, m.Port, m.WarningDays, m.CriticalDays)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err != nil {
		errMsg := err.Error()
		log.Printf("[SSL MONITOR] %s → error [%s]", m.Domain, errMsg)

		s.db.UpdateSSLMonitorStatus(ctx, m.ID,
			"error", "", "", nil, nil, nil, &errMsg)

		s.db.InsertSSLMonitorLog(ctx, &utils.SSLMonitorLog{
			MonitorID:    m.ID,
			Status:       "error",
			ErrorMessage: &errMsg,
		})

		s.broadcastUpdate(m, "error", 0, "", nil, &errMsg)
		return
	}

	log.Printf("[SSL MONITOR] %s → %s [%d days] issuer: %s",
		m.Domain, result.Status, result.DaysRemaining, result.Issuer)

	s.db.UpdateSSLMonitorStatus(ctx, m.ID,
		result.Status, result.Issuer, result.Subject,
		&result.ValidFrom, &result.ValidUntil,
		&result.DaysRemaining, nil,
	)

	s.db.InsertSSLMonitorLog(ctx, &utils.SSLMonitorLog{
		MonitorID:     m.ID,
		Status:        result.Status,
		Issuer:        &result.Issuer,
		ValidUntil:    &result.ValidUntil,
		DaysRemaining: &result.DaysRemaining,
	})

	s.broadcastUpdate(m, result.Status, result.DaysRemaining, result.Issuer,
		&result.ValidUntil, nil)
}

type certResult struct {
	Status        string
	Issuer        string
	Subject       string
	ValidFrom     time.Time
	ValidUntil    time.Time
	DaysRemaining int
}

func (s *SSLMonitorService) checkCertificate(domain string, port, warningDays, criticalDays int) (*certResult, error) {
	// ✅ Clean domain — remove http/https prefix
	domain = cleanDomain(domain)

	addr := fmt.Sprintf("%s:%d", domain, port)

	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		addr,
		&tls.Config{
			ServerName:         domain,
			InsecureSkipVerify: false,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("TLS connect failed: %w", err)
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	// ✅ Use leaf certificate (first one)
	cert := certs[0]
	now := time.Now()
	daysRemaining := int(math.Ceil(cert.NotAfter.Sub(now).Hours() / 24))

	// ✅ Determine status based on thresholds
	var status string
	switch {
	case now.After(cert.NotAfter):
		status = "expired"
	case daysRemaining <= criticalDays:
		status = "critical"
	case daysRemaining <= warningDays:
		status = "warning"
	default:
		status = "valid"
	}

	// ✅ Extract issuer
	issuer := cert.Issuer.Organization
	issuerStr := "Unknown"
	if len(issuer) > 0 {
		issuerStr = strings.Join(issuer, ", ")
	} else if cert.Issuer.CommonName != "" {
		issuerStr = cert.Issuer.CommonName
	}

	return &certResult{
		Status:        status,
		Issuer:        issuerStr,
		Subject:       cert.Subject.CommonName,
		ValidFrom:     cert.NotBefore,
		ValidUntil:    cert.NotAfter,
		DaysRemaining: daysRemaining,
	}, nil
}

func (s *SSLMonitorService) broadcastUpdate(m *utils.SSLMonitor, status string, daysRemaining int,
	issuer string, validUntil *time.Time, errMsg *string,
) {
	if s.broadcast == nil {
		return
	}

	validUntilStr := ""
	if validUntil != nil {
		validUntilStr = validUntil.Format(time.RFC3339)
	}
	errStr := ""
	if errMsg != nil {
		errStr = *errMsg
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"type":           "ssl_monitor_update",
		"monitor_id":     m.ID,
		"domain":         m.Domain,
		"friendly_name":  m.FriendlyName,
		"status":         status,
		"days_remaining": daysRemaining,
		"issuer":         issuer,
		"valid_until":    validUntilStr,
		"warning_days":   m.WarningDays,
		"critical_days":  m.CriticalDays,
		"error":          errStr,
		"checked_at":     time.Now().UTC(),
	})
	s.broadcast(payload)
}

// CheckOnce does a single check — used for test connection
func (s *SSLMonitorService) CheckOnce(domain string, port, warningDays, criticalDays int) (*certResult, error) {
	return s.checkCertificate(domain, port, warningDays, criticalDays)
}

func cleanDomain(domain string) string {
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "www.")
	domain = strings.Split(domain, "/")[0] // remove path
	domain = strings.Split(domain, ":")[0] // remove port
	return strings.TrimSpace(domain)
}
