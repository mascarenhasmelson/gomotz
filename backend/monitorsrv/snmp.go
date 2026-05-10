package monitorsrv

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/mascarenhasmelson/gomotz/utils"
)

type SNMPMonitorService struct {
	db        *PostgresDB
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	monitors  map[int]*snmpWorker
	wg        sync.WaitGroup
	broadcast BroadcastFunc
}

type snmpWorker struct {
	monitor *utils.SNMPMonitor
	cancel  context.CancelFunc
}

func NewSNMPMonitorService(db *PostgresDB, broadcast BroadcastFunc) *SNMPMonitorService {
	ctx, cancel := context.WithCancel(context.Background())
	return &SNMPMonitorService{
		db:        db,
		ctx:       ctx,
		cancel:    cancel,
		monitors:  make(map[int]*snmpWorker),
		broadcast: broadcast,
	}
}

func (s *SNMPMonitorService) RecoverFromDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitors, err := s.db.GetAllEnabledSNMPMonitors(ctx)
	if err != nil {
		return err
	}

	log.Printf("[SNMP MONITOR] Recovering %d monitor(s)", len(monitors))
	for _, m := range monitors {
		s.StartMonitor(m)
	}
	return nil
}

func (s *SNMPMonitorService) StartMonitor(m *utils.SNMPMonitor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w, exists := s.monitors[m.ID]; exists {
		w.cancel()
	}

	ctx, cancel := context.WithCancel(s.ctx)
	s.monitors[m.ID] = &snmpWorker{monitor: m, cancel: cancel}

	s.wg.Add(1)
	go s.runWorker(ctx, m)

	log.Printf("[SNMP MONITOR] Started %d — %s (%s:%d) OID:%s every %ds",
		m.ID, m.FriendlyName, m.Hostname, m.Port, m.OID, m.PollingInterval)
}

func (s *SNMPMonitorService) StopMonitor(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w, exists := s.monitors[id]; exists {
		w.cancel()
		delete(s.monitors, id)
		log.Printf("[SNMP MONITOR] Stopped monitor %d", id)
	}
}

func (s *SNMPMonitorService) Shutdown() {
	s.cancel()
	s.wg.Wait()
	log.Println("[SNMP MONITOR] All monitors stopped")
}

func (s *SNMPMonitorService) GetActiveCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.monitors)
}

func (s *SNMPMonitorService) runWorker(ctx context.Context, m *utils.SNMPMonitor) {
	defer s.wg.Done()
	s.pollAndSave(m)

	ticker := time.NewTicker(time.Duration(m.PollingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.pollAndSave(m)
		}
	}
}

func (s *SNMPMonitorService) pollAndSave(m *utils.SNMPMonitor) {
	start := time.Now()

	value, err := s.pollSNMP(m)

	responseMs := int(time.Since(start).Milliseconds())
	responseMsPtr := &responseMs

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var status string
	var errMsgPtr *string
	var valuePtr *string

	if err != nil {
		status = "down"
		errMsg := err.Error()
		errMsgPtr = &errMsg
		responseMsPtr = nil

		log.Printf("[SNMP MONITOR] %s (%s:%d) OID:%s → down [%s]",
			m.FriendlyName, m.Hostname, m.Port, m.OID, err.Error())
	} else {
		status = "up"
		valuePtr = &value

		log.Printf("[SNMP MONITOR] %s (%s:%d) OID:%s → up [%s] (%dms)",
			m.FriendlyName, m.Hostname, m.Port, m.OID, value, responseMs)
	}

	if status == "down" && m.Retries > 0 {
		for i := 0; i < m.Retries; i++ {
			time.Sleep(2 * time.Second)
			retryVal, retryErr := s.pollSNMP(m)
			if retryErr == nil {
				status = "up"
				valuePtr = &retryVal
				errMsgPtr = nil
				responseMsPtr = &responseMs
				log.Printf("[SNMP MONITOR] %s recovered on retry %d", m.FriendlyName, i+1)
				break
			}
		}
	}
	valueStr := ""
	if valuePtr != nil {
		valueStr = *valuePtr
	}
	if err := s.db.UpdateSNMPMonitorStatus(ctx, m.ID, status, valueStr, responseMsPtr, errMsgPtr); err != nil {
		log.Printf("[SNMP MONITOR] Failed to update status %d: %v", m.ID, err)
	}

	if err := s.db.InsertSNMPMonitorLog(ctx, &utils.SNMPMonitorLog{
		MonitorID:    m.ID,
		Status:       status,
		Value:        valuePtr,
		ResponseMs:   responseMsPtr,
		ErrorMessage: errMsgPtr,
	}); err != nil {
		log.Printf("[SNMP MONITOR] Failed to insert log %d: %v", m.ID, err)
	}

	if s.broadcast != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"type":          "snmp_monitor_update",
			"monitor_id":    m.ID,
			"friendly_name": m.FriendlyName,
			"hostname":      m.Hostname,
			"port":          m.Port,
			"oid":           m.OID,
			"status":        status,
			"value":         valueStr,
			"response_ms":   responseMs,
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

func (s *SNMPMonitorService) pollSNMP(m *utils.SNMPMonitor) (string, error) {
	g := &gosnmp.GoSNMP{
		Target:    m.Hostname,
		Port:      uint16(m.Port),
		Community: m.CommunityString,
		Timeout:   time.Duration(m.Timeout) * time.Second,
		Retries:   0, // handled manually
	}

	switch m.SNMPVersion {
	case "v1":
		g.Version = gosnmp.Version1
	case "v2c":
		g.Version = gosnmp.Version2c
	default:
		g.Version = gosnmp.Version2c
	}

	if err := g.Connect(); err != nil {
		return "", fmt.Errorf("connect failed: %w", err)
	}
	defer g.Conn.Close()
	result, err := g.Get([]string{m.OID})
	if err != nil {
		return "", fmt.Errorf("SNMP get failed: %w", err)
	}
	if len(result.Variables) == 0 {
		return "", fmt.Errorf("no variables returned")
	}
	variable := result.Variables[0]
	if variable.Type == gosnmp.NoSuchObject || variable.Type == gosnmp.NoSuchInstance {
		return "", fmt.Errorf("OID not found: %s", m.OID)
	}

	return formatSNMPValue(variable, m.ExpectedValueType), nil
}

func formatSNMPValue(variable gosnmp.SnmpPDU, expectedType string) string {
	switch variable.Type {
	case gosnmp.OctetString:
		b, ok := variable.Value.([]byte)
		if ok {
			printable := true
			for _, c := range b {
				if c < 32 || c > 126 {
					printable = false
					break
				}
			}
			if printable {
				return string(b)
			}
			return fmt.Sprintf("%x", b)
		}
		return fmt.Sprintf("%v", variable.Value)
	case gosnmp.Integer:
		return fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value))

	case gosnmp.Counter32, gosnmp.Counter64:
		return fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value))
	case gosnmp.Gauge32:
		return fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value))
	case gosnmp.TimeTicks:
		ticks := gosnmp.ToBigInt(variable.Value).Uint64()
		seconds := ticks / 100
		days := seconds / 86400
		hours := (seconds % 86400) / 3600
		minutes := (seconds % 3600) / 60
		secs := seconds % 60
		if days > 0 {
			return fmt.Sprintf("%dd %02dh %02dm %02ds", days, hours, minutes, secs)
		}
		return fmt.Sprintf("%02dh %02dm %02ds", hours, minutes, secs)

	case gosnmp.ObjectIdentifier:
		return fmt.Sprintf("%v", variable.Value)

	case gosnmp.IPAddress:
		return fmt.Sprintf("%v", variable.Value)

	default:
		return fmt.Sprintf("%v", variable.Value)
	}
}

func (s *SNMPMonitorService) PollOnce(m *utils.SNMPMonitor) (string, int, error) {
	start := time.Now()
	value, err := s.pollSNMP(m)
	responseMs := int(time.Since(start).Milliseconds())
	return value, responseMs, err
}

func validateOID(oid string) bool {
	if len(oid) == 0 {
		return false
	}
	trimmed := strings.TrimPrefix(oid, ".")
	if len(trimmed) == 0 {
		return false
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
		}
	}

	return true
}
