package vlan

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/mascarenhasmelson/gomotz/utils"
)

// ============================================
// VLAN SCAN MANAGER
// ============================================

type VLANScanManager struct {
	mu       sync.RWMutex
	scanners map[int]*VLANScanner
	db       *PostgresDB
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	// One HostnameDiscovery per physical interface, shared across VLANs.
	hostnameDiscovery map[string]*HostnameDiscovery

	// pendingVLANs holds VLANs whose interface was not found at startup.
	// The retry loop periodically attempts to start them.
	pendingVLANs map[int]*utils.VLANNetwork
	pendingMu    sync.Mutex
}

type VLANScanner struct {
	VLANId       int
	Scanner      *ARPScanner
	Config       *utils.VLANNetwork
	Status       string
	LastScanTime time.Time
	Cancel       context.CancelFunc
	IsRunning    bool
}

func NewVLANScanManager(database *PostgresDB) *VLANScanManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &VLANScanManager{
		scanners:          make(map[int]*VLANScanner),
		hostnameDiscovery: make(map[string]*HostnameDiscovery),
		pendingVLANs:      make(map[int]*utils.VLANNetwork),
		db:                database,
		ctx:               ctx,
		cancel:            cancel,
	}
}

// ============================================
// SCAN LIFECYCLE
// ============================================
func (m *VLANScanManager) GetHostnameDiscovery() map[string]*HostnameDiscovery {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*HostnameDiscovery)
	for k, v := range m.hostnameDiscovery {
		result[k] = v
	}
	return result
}

// TriggerDiscoveryScan manually triggers a discovery scan on specified interface
func (m *VLANScanManager) TriggerDiscoveryScan(targetInterface string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	triggered := 0

	for ifaceName, hd := range m.hostnameDiscovery {
		// ✅ FIXED: Compare ifaceName to targetInterface, not interfaceName to interfaceName
		if targetInterface != "" && ifaceName != targetInterface {
			continue
		}
		go hd.ScanSubnet()
		triggered++
		log.Printf("[DISCOVERY] Manual scan triggered on %s", ifaceName)
	}

	return triggered
}
func (m *VLANScanManager) StartVLANScan(config *utils.VLANNetwork) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if vs, exists := m.scanners[config.VLANId]; exists && vs.IsRunning {
		return fmt.Errorf("scan already running for VLAN %d", config.VLANId)
	}

	return m.startVLANScanLocked(config)
}

// startVLANScanLocked does the actual work. Caller must hold m.mu.Lock().
func (m *VLANScanManager) startVLANScanLocked(config *utils.VLANNetwork) error {
	cidr, err := m.GetCIDRFromConfig(config)
	if err != nil {
		return err
	}

	scanInterval := time.Duration(config.ScanIntervalSeconds) * time.Second

	interfaceName, err := m.DetectInterfaceForCIDR(cidr, config.VLANId)
	if err != nil {
		log.Printf("[VLAN %d] Interface not found, marking all devices offline: %v",
			config.VLANId, err)
		m.markAllDevicesOffline(config.VLANId)
		return fmt.Errorf("failed to detect interface for VLAN %d: %w", config.VLANId, err)
	}

	// Shared mDNS/SSDP discovery per interface.
	if _, exists := m.hostnameDiscovery[interfaceName]; !exists {
		hd, err := NewHostnameDiscovery(interfaceName)
		if err != nil {
			log.Printf("[VLAN %d] Hostname discovery init failed on %s: %v",
				config.VLANId, interfaceName, err)
		} else {
			m.hostnameDiscovery[interfaceName] = hd
			go hd.Start(5 * scanInterval)
			log.Printf("[VLAN %d] Hostname discovery started on %s (shared)",
				config.VLANId, interfaceName)
		}
	}

	arpScanner, err := NewARPScanner(cidr, scanInterval)
	if err != nil {
		return fmt.Errorf("failed to create ARP scanner for VLAN %d: %w", config.VLANId, err)
	}

	arpScanner.SetVendorLookup(NewDatabaseVendorLookup(m.db))

	if hd, ok := m.hostnameDiscovery[interfaceName]; ok {
		arpScanner.SetHostnameDiscovery(hd)
	}

	arpScanner.OnARPEvent = m.buildEventCallback(config.VLANId)

	ctx, cancel := context.WithCancel(m.ctx)
	vs := &VLANScanner{
		VLANId:    config.VLANId,
		Scanner:   arpScanner,
		Config:    config,
		Status:    "running",
		Cancel:    cancel,
		IsRunning: true,
	}
	m.scanners[config.VLANId] = vs

	arpScanner.Start()

	log.Printf("[VLAN %d] Scan started — CIDR: %s, interval: %ds",
		config.VLANId, cidr, config.ScanIntervalSeconds)

	m.wg.Add(1)
	go m.monitorVLAN(ctx, vs)

	return nil
}
func (m *VLANScanManager) markAllDevicesOffline(vlanID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE discovered_devices
		SET device_status = 'offline'
		WHERE vlan_id = $1
		  AND device_status != 'offline'
	`

	result, err := m.db.GetPool().Exec(ctx, query, vlanID)
	if err != nil {
		log.Printf("[VLAN %d] Failed to mark devices offline: %v", vlanID, err)
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("[VLAN %d] Marked %d devices as offline (interface down)",
			vlanID, rowsAffected)
	}
}
func (m *VLANScanManager) StopVLANScan(vlanID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pendingMu.Lock()
	delete(m.pendingVLANs, vlanID)
	m.pendingMu.Unlock()

	vs, exists := m.scanners[vlanID]
	if !exists {
		return fmt.Errorf("no scanner found for VLAN %d", vlanID)
	}

	if !vs.IsRunning {
		return fmt.Errorf("scanner not running for VLAN %d", vlanID)
	}

	vs.Cancel()

	cidr, _ := m.GetCIDRFromConfig(vs.Config) // ✅ CHANGED
	if cidr != "" {
		ifaceName, _ := m.DetectInterfaceForCIDR(cidr, vlanID) // ✅ CHANGED
		if ifaceName != "" && !m.interfaceStillInUse(ifaceName, vlanID) {
			if hd, ok := m.hostnameDiscovery[ifaceName]; ok {
				hd.Stop()
				delete(m.hostnameDiscovery, ifaceName)
				log.Printf("[VLAN %d] Hostname discovery stopped on %s", vlanID, ifaceName)
			}
		}
	}

	log.Printf("[VLAN %d] Scan stopped", vlanID)
	return nil
}
func (m *VLANScanManager) RecoverFromRestart() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configs, err := m.db.GetEnabledVLANs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled VLANs: %w", err)
	}

	log.Printf("Recovering %d enabled VLAN scan(s)...", len(configs))

	for _, config := range configs {
		if err := m.StartVLANScan(config); err != nil {
			log.Printf("[VLAN %d] Start failed: %v — adding to retry queue", config.VLANId, err)
			m.queueForRetry(config)
			continue
		}
		log.Printf("[VLAN %d] ✅ Recovered: %s", config.VLANId, config.VLANName)
		time.Sleep(500 * time.Millisecond)
	}

	// ✅ ALWAYS start infinite retry loop
	m.wg.Add(1)
	go m.retryPendingVLANs()
	log.Printf("[RETRY] ♾️  Infinite retry loop started (runs forever, checks every 30s)")

	return nil
}
func (m *VLANScanManager) queueForRetry(config *utils.VLANNetwork) {
	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()
	m.pendingVLANs[config.VLANId] = config
}

func (m *VLANScanManager) retryPendingVLANs() {
	defer m.wg.Done()

	// ✅ Retry every 30 seconds - FOREVER
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Printf("[RETRY] ♾️  Infinite retry loop started (checks every 30s)")

	retryCount := 0

	for {
		select {
		case <-m.ctx.Done():
			log.Printf("[RETRY] Context cancelled, exiting retry loop")
			return

		case <-ticker.C:
			retryCount++

			m.pendingMu.Lock()
			pendingCount := len(m.pendingVLANs)

			if pendingCount == 0 {
				m.pendingMu.Unlock()
				// ✅ No pending VLANs, just wait for next tick
				if retryCount%10 == 0 {
					log.Printf("[RETRY] Heartbeat #%d - No pending VLANs", retryCount)
				}
				continue
			}

			// Copy pending list
			toRetry := make([]*utils.VLANNetwork, 0, pendingCount)
			vlanIDs := make([]int, 0, pendingCount)
			for vlanID, config := range m.pendingVLANs {
				toRetry = append(toRetry, config)
				vlanIDs = append(vlanIDs, vlanID)
			}
			m.pendingMu.Unlock()

			log.Printf("[RETRY] Attempt #%d - Trying to start %d VLAN(s): %v",
				retryCount, len(toRetry), vlanIDs)

			for _, config := range toRetry {
				// ✅ Check if VLAN still exists in database before trying
				dbConfig, err := m.db.GetVLANNetwork(context.Background(), config.VLANId)
				if err != nil {
					// VLAN deleted from database - remove from retry queue
					log.Printf("[RETRY] VLAN %d deleted from database, removing from retry queue",
						config.VLANId)

					m.pendingMu.Lock()
					delete(m.pendingVLANs, config.VLANId)
					m.pendingMu.Unlock()
					continue
				}

				// ✅ Check if monitoring is disabled
				if !dbConfig.MonitoringEnabled {
					log.Printf("[RETRY] VLAN %d monitoring disabled, removing from retry queue",
						config.VLANId)

					m.pendingMu.Lock()
					delete(m.pendingVLANs, config.VLANId)
					m.pendingMu.Unlock()
					continue
				}

				// ✅ Try to start
				m.mu.Lock()
				err = m.startVLANScanLocked(dbConfig) // Use fresh config from DB
				m.mu.Unlock()

				if err != nil {
					log.Printf("[RETRY] VLAN %d still unavailable: %v (will retry in 30s)",
						config.VLANId, err)
					continue
				}

				// Success — remove from pending
				m.pendingMu.Lock()
				delete(m.pendingVLANs, config.VLANId)
				m.pendingMu.Unlock()

				log.Printf("[RETRY] ✅ VLAN %d started successfully", config.VLANId)
			}

			// Log current pending count
			m.pendingMu.Lock()
			remaining := len(m.pendingVLANs)
			m.pendingMu.Unlock()

			if remaining > 0 {
				log.Printf("[RETRY] %d VLAN(s) still pending, will retry in 30s", remaining)
			} else {
				log.Printf("[RETRY] ✅ All pending VLANs started successfully")
			}
		}
	}
}

func (m *VLANScanManager) Shutdown() {
	log.Println("Shutting down all VLAN scanners...")

	m.mu.Lock()
	for vlanID, vs := range m.scanners {
		if vs.IsRunning {
			vs.Cancel()
			log.Printf("  Stopped VLAN %d", vlanID)
		}
	}
	m.mu.Unlock()

	for ifaceName, hd := range m.hostnameDiscovery {
		hd.Stop()
		log.Printf("  Stopped hostname discovery on %s", ifaceName)
	}

	m.cancel()
	m.wg.Wait()
	log.Println("All scanners stopped.")
}

func (m *VLANScanManager) buildEventCallback(vlanID int) ARPEventCallback {
	return func(event ARPEvent, host *Host, oldHost *Host) {
		switch event {
		case EventNewDevice:
			log.Printf("[VLAN %d]  New device: %s (%s)", vlanID, host.IP, host.MAC)

		case EventIPChange:
			oldIP := ""
			if oldHost != nil {
				oldIP = oldHost.IP.String()
			}
			log.Printf("[VLAN %d]  IP change: %s — %s → %s", vlanID, host.MAC, oldIP, host.IP)

		case EventIPConflict:
			oldMAC := ""
			if oldHost != nil {
				oldMAC = oldHost.MAC.String()
			}
			log.Printf("[VLAN %d]   IP CONFLICT + IP CHANGE: %s (old MAC: %s → conflict)",
				vlanID, host.IP, oldMAC)

		case EventIPConflictND:
			log.Printf("[VLAN %d]  DUPLICATE IP DETECTED: %s — Multiple MACs claiming this IP!",
				vlanID, host.IP)
			fmt.Printf("\n")
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf(" DUPLICATE IP CONFLICT DETECTED\n")
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("VLAN:     %d\n", vlanID)
			fmt.Printf("IP:       %s\n", host.IP)
			fmt.Printf("Status:   CONFLICT - Multiple devices using same IP\n")
			fmt.Printf("Action:   Investigate immediately\n")
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("\n")

		case EventMACChange:
			oldMAC, oldIP := "", ""
			if oldHost != nil {
				oldMAC = oldHost.MAC.String()
				oldIP = oldHost.IP.String()
			}
			log.Printf("[VLAN %d]  MAC+IP change: %s (old MAC: %s, old IP: %s)",
				vlanID, host.IP, oldMAC, oldIP)

		case EventMACChangeND:
			oldMAC := ""
			if oldHost != nil {
				oldMAC = oldHost.MAC.String()
			}
			log.Printf("[VLAN %d]  MAC change + new device: %s (old MAC: %s)",
				vlanID, host.IP, oldMAC)

		case EventCameOnline:
			log.Printf("[VLAN %d]  Gratuitous ARP (still online): %s (%s)",
				vlanID, host.IP, host.MAC)

		case EventCameOnline2:
			log.Printf("[VLAN %d]  Came online: %s (%s)", vlanID, host.IP, host.MAC)

		case EventWentOffline:
			log.Printf("[VLAN %d]   Went offline: %s (%s)", vlanID, host.IP, host.MAC)
		}

		m.saveHostToDB(vlanID, host)
	}
}

// func (m *VLANScanManager) monitorVLAN(ctx context.Context, vs *VLANScanner) {
// 	defer m.wg.Done()

// 	// ✅ Health check every 30 seconds
// 	ticker := time.NewTicker(30 * time.Second)
// 	defer ticker.Stop()

// 	consecutiveFailures := 0

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			vs.Scanner.Stop()
// 			vs.IsRunning = false
// 			vs.Status = "stopped"
// 			return

// 		case <-ticker.C:
// 			vs.LastScanTime = time.Now()

// 			// ✅ Check if VLAN still exists in database
// 			dbConfig, err := m.db.GetVLANNetwork(context.Background(), vs.VLANId)
// 			if err != nil {
// 				// VLAN deleted from database - stop completely
// 				log.Printf("[VLAN %d] ⚠️  VLAN deleted from database, stopping scanner", vs.VLANId)

// 				vs.Scanner.Stop()
// 				vs.IsRunning = false
// 				vs.Status = "deleted"

// 				// Remove from scanners map
// 				m.mu.Lock()
// 				delete(m.scanners, vs.VLANId)
// 				m.mu.Unlock()

// 				// Remove from pending if exists
// 				m.pendingMu.Lock()
// 				delete(m.pendingVLANs, vs.VLANId)
// 				m.pendingMu.Unlock()

// 				log.Printf("[VLAN %d] ✅ Cleanup complete", vs.VLANId)
// 				return
// 			}

// 			// ✅ Check if monitoring is disabled
// 			if !dbConfig.MonitoringEnabled {
// 				log.Printf("[VLAN %d] ⚠️  Monitoring disabled in database, stopping scanner", vs.VLANId)

// 				vs.Scanner.Stop()
// 				vs.IsRunning = false
// 				vs.Status = "disabled"

// 				m.mu.Lock()
// 				delete(m.scanners, vs.VLANId)
// 				m.mu.Unlock()

// 				m.pendingMu.Lock()
// 				delete(m.pendingVLANs, vs.VLANId)
// 				m.pendingMu.Unlock()

// 				log.Printf("[VLAN %d] ✅ Stopped due to disabled monitoring", vs.VLANId)
// 				return
// 			}

// 			// ✅ Check if interface is still available
// 			cidr, _ := m.GetCIDRFromConfig(vs.Config)
// 			if cidr != "" {
// 				_, err := m.DetectInterfaceForCIDR(cidr, vs.VLANId)
// 				// if err != nil {
// 				// 	// Interface is down
// 				// 	consecutiveFailures++
// 				// 	log.Printf("[VLAN %d] Interface check failed (%d consecutive failures): %v",
// 				// 		vs.VLANId, consecutiveFailures, err)

// 				// 	// ✅ After 3 failures, mark offline and re-queue (NO MAX, keeps trying)
// 				// 	if consecutiveFailures == 199 {
// 				// 		log.Printf("[VLAN %d] ⚠️  Interface lost, marking devices offline and re-queuing",
// 				// 			vs.VLANId)

// 				// 		// Mark all devices offline
// 				// 		m.markAllDevicesOffline(vs.VLANId)

// 				// 		// Stop the scanner
// 				// 		vs.Scanner.Stop()
// 				// 		vs.IsRunning = false
// 				// 		vs.Status = "interface_down"

// 				// 		// ✅ Re-queue for continuous retry
// 				// 		m.requeueForRetry(vs.Config, vs.VLANId)

// 				// 		// Exit this monitor goroutine
// 				// 		log.Printf("[VLAN %d] Monitor exited, retry loop will continue every 30s", vs.VLANId)
// 				// 		return
// 				// 	}
// 				// }
// 				if err != nil {
// 					consecutiveFailures++

// 					log.Printf("[VLAN %d] Interface DOWN (%d failures): %v",
// 						vs.VLANId, consecutiveFailures, err)

// 					if vs.IsRunning {
// 						log.Printf("[VLAN %d] ⚠️ Stopping scanner (interface down)", vs.VLANId)

// 						vs.Scanner.Stop()
// 						vs.IsRunning = false
// 						vs.Status = "interface_down"

// 						m.markAllDevicesOffline(vs.VLANId)
// 					}

// 					continue
// 				} else {
// 					// Interface is up
// 					// if consecutiveFailures > 0 {
// 					// 	log.Printf("[VLAN %d] ✅ Interface recovered after %d failures",
// 					// 		vs.VLANId, consecutiveFailures)
// 					// }
// 					// consecutiveFailures = 0
// 					// vs.Status = "running"
// 					if !vs.IsRunning {
// 						log.Printf("[VLAN %d] ✅ Interface recovered, recreating scanner", vs.VLANId)

// 						cidr, err := m.GetCIDRFromConfig(vs.Config)
// 						if err != nil {
// 							log.Printf("[VLAN %d] CIDR error: %v", vs.VLANId, err)
// 							continue
// 						}

// 						scanInterval := time.Duration(vs.Config.ScanIntervalSeconds) * time.Second

// 						ifaceName, err := m.DetectInterfaceForCIDR(cidr, vs.VLANId)
// 						if err != nil {
// 							log.Printf("[VLAN %d] Interface still not ready: %v", vs.VLANId, err)
// 							continue
// 						}

// 						arpScanner, err := NewARPScanner(cidr, scanInterval)
// 						if err != nil {
// 							log.Printf("[VLAN %d] Failed to recreate scanner: %v", vs.VLANId, err)
// 							continue
// 						}

// 						arpScanner.SetVendorLookup(NewDatabaseVendorLookup(m.db))

// 						if hd, ok := m.hostnameDiscovery[ifaceName]; ok {
// 							arpScanner.SetHostnameDiscovery(hd)
// 						}

// 						arpScanner.OnARPEvent = m.buildEventCallback(vs.VLANId)

// 						// Replace scanner
// 						vs.Scanner = arpScanner

// 						arpScanner.Start()
// 						vs.IsRunning = true
// 					}

// 					if consecutiveFailures > 0 {
// 						log.Printf("[VLAN %d] Interface recovered after %d failures",
// 							vs.VLANId, consecutiveFailures)
// 					}

// 					consecutiveFailures = 0
// 					vs.Status = "running"
// 				}
// 			}

// 			// Save current hosts to database
// 			for _, host := range vs.Scanner.GetHosts() {
// 				m.saveHostToDB(vs.VLANId, host)
// 			}
// 		}
// 	}
// }

func (m *VLANScanManager) monitorVLAN(ctx context.Context, vs *VLANScanner) {
	defer m.wg.Done()

	ticker := time.NewTicker(40 * time.Second) // ✅ changed to 40s
	defer ticker.Stop()

	consecutiveFailures := 0

	for {
		select {
		case <-ctx.Done():
			if vs.IsRunning {
				vs.Scanner.Stop()
			}
			vs.IsRunning = false
			vs.Status = "stopped"
			return

		case <-ticker.C:
			vs.LastScanTime = time.Now()

			// ================================
			// ✅ 1. Validate VLAN from DB
			// ================================
			dbConfig, err := m.db.GetVLANNetwork(context.Background(), vs.VLANId)
			if err != nil {
				log.Printf("[VLAN %d] ❌ Deleted from DB, stopping", vs.VLANId)
				m.cleanupVLAN(vs.VLANId, "deleted")
				return
			}

			if !dbConfig.MonitoringEnabled {
				log.Printf("[VLAN %d] ⚠️ Monitoring disabled", vs.VLANId)
				m.cleanupVLAN(vs.VLANId, "disabled")
				return
			}

			// ================================
			// ✅ 2. Detect Interface
			// ================================
			cidr, err := m.GetCIDRFromConfig(vs.Config)
			if err != nil || cidr == "" {
				log.Printf("[VLAN %d] CIDR error: %v", vs.VLANId, err)
				continue
			}

			ifaceName, err := m.DetectInterfaceForCIDR(cidr, vs.VLANId)

			// ================================
			// 🔴 3. Interface DOWN
			// ================================
			if err != nil {
				consecutiveFailures++

				log.Printf("[VLAN %d] 🔴 Interface DOWN (%d): %v",
					vs.VLANId, consecutiveFailures, err)

				if vs.IsRunning {
					log.Printf("[VLAN %d] Stopping scanner", vs.VLANId)

					vs.Scanner.Stop()
					vs.IsRunning = false
					vs.Status = "interface_down"

					// 🔥 CRITICAL: sync DB
					m.markAllDevicesOffline(vs.VLANId)
				}

				continue
			}

			// ================================
			// 🟢 4. Interface UP
			// ================================
			if !vs.IsRunning {
				log.Printf("[VLAN %d] 🟢 Interface recovered, recreating scanner", vs.VLANId)

				if err := m.recreateScanner(vs, cidr, ifaceName); err != nil {
					log.Printf("[VLAN %d] Failed to recreate scanner: %v", vs.VLANId, err)
					continue
				}

				vs.IsRunning = true
			}

			if consecutiveFailures > 0 {
				log.Printf("[VLAN %d] ✅ Recovered after %d failures",
					vs.VLANId, consecutiveFailures)
			}

			consecutiveFailures = 0
			vs.Status = "running"

			// ================================
			// ✅ 5. Persist Hosts
			// ================================
			for _, host := range vs.Scanner.GetHosts() {
				m.saveHostToDB(vs.VLANId, host)
			}
		}
	}
}
func (m *VLANScanManager) recreateScanner(vs *VLANScanner, cidr, ifaceName string) error {
	scanInterval := time.Duration(vs.Config.ScanIntervalSeconds) * time.Second

	arpScanner, err := NewARPScanner(cidr, scanInterval)
	if err != nil {
		return err
	}

	arpScanner.SetVendorLookup(NewDatabaseVendorLookup(m.db))

	if hd, ok := m.hostnameDiscovery[ifaceName]; ok {
		arpScanner.SetHostnameDiscovery(hd)
	}

	arpScanner.OnARPEvent = m.buildEventCallback(vs.VLANId)

	vs.Scanner = arpScanner
	arpScanner.Start()

	return nil
}
func (m *VLANScanManager) cleanupVLAN(vlanID int, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if vs, exists := m.scanners[vlanID]; exists {
		if vs.IsRunning {
			vs.Scanner.Stop()
		}
		vs.IsRunning = false
		vs.Status = reason
		delete(m.scanners, vlanID)
	}

	m.pendingMu.Lock()
	delete(m.pendingVLANs, vlanID)
	m.pendingMu.Unlock()

	log.Printf("[VLAN %d] ✅ Cleaned up (%s)", vlanID, reason)
}

/////////////////////////////////////////////

// ✅ NEW: Re-queue a VLAN for retry after interface failure
func (m *VLANScanManager) requeueForRetry(config *utils.VLANNetwork, vlanID int) {
	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()

	// Add to pending queue (or update if already there)
	m.pendingVLANs[vlanID] = config

	// Remove from active scanners
	m.mu.Lock()
	delete(m.scanners, vlanID)
	m.mu.Unlock()

	log.Printf("[VLAN %d] ✅ Added to infinite retry queue (30s interval)", vlanID)
}

func (m *VLANScanManager) ensureRetryLoopRunning() {
	select {
	case <-m.ctx.Done():
		return
	default:
		// Start a new retry loop goroutine
		m.wg.Add(1)
		go m.retryPendingVLANs()
	}
}

// ============================================
// STATUS
// ============================================

func (m *VLANScanManager) GetAllStatuses() map[int]*VLANScanner {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[int]*VLANScanner, len(m.scanners))
	for k, v := range m.scanners {
		out[k] = v
	}
	return out
}

// GetPendingVLANs returns the list of VLANs still waiting for their interface.
func (m *VLANScanManager) GetPendingVLANs() []int {
	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()
	ids := make([]int, 0, len(m.pendingVLANs))
	for id := range m.pendingVLANs {
		ids = append(ids, id)
	}
	return ids
}

func (m *VLANScanManager) GetScannerStatus(vlanID int) (*VLANScanner, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	vs, exists := m.scanners[vlanID]
	if !exists {
		return nil, fmt.Errorf("scanner not found for VLAN %d", vlanID)
	}
	return vs, nil
}

// ============================================
// HELPERS
// ============================================

func (m *VLANScanManager) interfaceStillInUse(ifaceName string, excludeVLAN int) bool {
	for id, vs := range m.scanners {
		if id == excludeVLAN || !vs.IsRunning {
			continue
		}
		cidr, _ := m.GetCIDRFromConfig(vs.Config)
		if cidr == "" {
			continue
		}
		name, _ := m.DetectInterfaceForCIDR(cidr, id)
		if name == ifaceName {
			return true
		}
	}
	return false
}

func (m *VLANScanManager) DetectInterfaceForCIDR(cidr string, vlanID int) (string, error) {
	_, targetNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR %s: %w", cidr, err)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get interfaces: %w", err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ifaceIPNet, ok := addr.(*net.IPNet)
			if !ok || ifaceIPNet.IP.To4() == nil {
				continue
			}
			if targetNet.Contains(ifaceIPNet.IP) {
				return iface.Name, nil
			}
		}
	}
	return "", fmt.Errorf("no interface found with IP in network %s", cidr)
}

func (m *VLANScanManager) GetCIDRFromConfig(config *utils.VLANNetwork) (string, error) {
	switch config.NetworkMode {
	case "static":
		if config.CIDRFull == nil || *config.CIDRFull == "" {
			return "", fmt.Errorf("static mode requires CIDR for VLAN %d", config.VLANId)
		}
		return *config.CIDRFull, nil
	case "dhcp":
		return "", fmt.Errorf("DHCP auto-detection not implemented for VLAN %d", config.VLANId)
	default:
		return "", fmt.Errorf("invalid network mode '%s' for VLAN %d",
			config.NetworkMode, config.VLANId)
	}
}

func (m *VLANScanManager) saveHostToDB(vlanID int, host *Host) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ✅ Detect conflict state
	deviceStatus := string(host.Status)
	if isBroadcastMAC(host.MAC) {
		deviceStatus = "conflict" // Mark as conflict instead of "online"
	}

	device := &utils.DiscoveredDevice{
		VLANId:       vlanID,
		IPAddress:    host.IP.String(),
		MACAddress:   host.MAC.String(),
		Hostname:     host.Hostname,
		Vendor:       host.Vendor,
		DeviceStatus: deviceStatus,
		FirstSeen:    host.FirstSeen,
		LastSeen:     host.LastSeen,
	}

	if err := m.db.UpsertDevice(ctx, device); err != nil {
		log.Printf("[VLAN %d] Failed to save device %s: %v", vlanID, host.IP, err)
	}
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
