package monitorsrv

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/mascarenhasmelson/gomotz/bgservices"
	"github.com/mascarenhasmelson/gomotz/utils"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

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
