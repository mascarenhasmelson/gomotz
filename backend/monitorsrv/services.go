package monitorsrv

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/mascarenhasmelson/gomotz/bgservices"
	"github.com/mascarenhasmelson/gomotz/utils"
)

type BroadcastFunc func(payload []byte)

type PortMonitorService struct {
	db        *PostgresDB
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	monitors  map[int]*portMonitorWorker
	wg        sync.WaitGroup
	broadcast BroadcastFunc // ✅ instead of *vlan.Hub
}

type portMonitorWorker struct {
	monitor *utils.PortMonitor
	cancel  context.CancelFunc
}

func NewPortMonitorService(db *PostgresDB, broadcast BroadcastFunc) *PortMonitorService { // ✅
	ctx, cancel := context.WithCancel(context.Background())
	return &PortMonitorService{
		db:        db,
		ctx:       ctx,
		cancel:    cancel,
		monitors:  make(map[int]*portMonitorWorker),
		broadcast: broadcast, // ✅
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

	// ✅ Derive status from tcp status, not just resp.Success
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

// ✅ Single place to define the mapping
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
