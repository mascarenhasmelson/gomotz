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

var hostnameCallbackRegistered bool
var hostnameCallbackRegisteredOnce sync.Once

type VLANScanManager struct {
	mu                sync.RWMutex
	scanners          map[int]*VLANScanner
	db                *PostgresDB
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	hostnameDiscovery map[string]*HostnameDiscovery
	pendingVLANs      map[int]*utils.VLANNetwork
	pendingMu         sync.Mutex
	parentInterface   string
	dbScanner         *DBARPScanner
}

type VLANScanner struct {
	NetworkId    int
	Scanner      *ARPScanner
	Config       *utils.VLANNetwork
	Status       string
	LastScanTime time.Time
	Cancel       context.CancelFunc
	IsRunning    bool
}

func (m *VLANScanManager) StopVLANScanByInterface(interfaceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for vlanID, vs := range m.scanners {
		if vs.Config.InterfaceName == interfaceName {
			m.pendingMu.Lock()
			delete(m.pendingVLANs, vlanID)
			m.pendingMu.Unlock()
			vs.Cancel()

			if !m.interfaceStillInUse(interfaceName, vlanID) {
				if hd, ok := m.hostnameDiscovery[interfaceName]; ok {
					hd.Stop()
					delete(m.hostnameDiscovery, interfaceName)
					log.Printf("[%s] Hostname discovery stopped", interfaceName)
				}
			}

			log.Printf("[%s] Scan stopped", interfaceName)
			return nil
		}
	}
	return fmt.Errorf("no scanner found for interface %s", interfaceName)
}

func NewVLANScanManager(database *PostgresDB, parentInterface string) *VLANScanManager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &VLANScanManager{
		parentInterface:   parentInterface,
		scanners:          make(map[int]*VLANScanner),
		hostnameDiscovery: make(map[string]*HostnameDiscovery),
		pendingVLANs:      make(map[int]*utils.VLANNetwork),
		db:                database,
		ctx:               ctx,
		cancel:            cancel,
	}
	m.dbScanner = NewDBARPScanner(database, 2*time.Minute)
	return m
}
func (m *VLANScanManager) GetHostnameDiscovery() map[string]*HostnameDiscovery {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]*HostnameDiscovery)
	for k, v := range m.hostnameDiscovery {
		result[k] = v
	}
	return result
}

func (m *VLANScanManager) TriggerDiscoveryScan(targetInterface string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	triggered := 0

	for ifaceName, hd := range m.hostnameDiscovery {
		if targetInterface != "" && ifaceName != targetInterface {
			continue
		}
		go hd.Start()
		triggered++
		log.Printf("[DISCOVERY] Manual scan triggered on %s", ifaceName)
	}

	return triggered
}
func (m *VLANScanManager) StartVLANScan(config *utils.VLANNetwork) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if vs, exists := m.scanners[config.ID]; exists && vs.IsRunning {
		return fmt.Errorf("scan already running for network %d (%s)", config.ID, config.InterfaceName)
	}

	return m.startVLANScanLocked(config)
}
func (m *VLANScanManager) startVLANScanLocked(config *utils.VLANNetwork) error {
	cidr, err := m.GetCIDRFromConfig(config)
	if err != nil {
		return err
	}

	scanInterval := time.Duration(config.ScanIntervalSeconds) * time.Second
	interfaceName, err := m.DetectInterfaceForCIDR(cidr, config.ID)
	if err != nil {
		log.Printf("[Network %d] Interface not found: %v", config.ID, err)
		m.markAllDevicesOffline(config.ID)
		return fmt.Errorf("failed to detect interface for network %d: %w", config.ID, err)
	}

	if _, exists := m.hostnameDiscovery[interfaceName]; !exists {
		hd, err := NewHostnameDiscovery(interfaceName)
		if err != nil {
			log.Printf("[Network %d] Hostname discovery init failed: %v", config.ID, err)
		} else {
			m.hostnameDiscovery[interfaceName] = hd
			go hd.Start()
			m.registerHostnameCallbackOnce()
		}
	}
	arpScanner, err := NewARPScanner(cidr, scanInterval)
	if err != nil {
		return fmt.Errorf("failed to create ARP scanner for network %d: %w", config.ID, err)
	}
	arpScanner.SetVendorLookup(NewDatabaseVendorLookup(m.db))
	if hd, ok := m.hostnameDiscovery[interfaceName]; ok {
		arpScanner.SetHostnameDiscovery(hd)
	}
	arpScanner.OnARPEvent = m.buildEventCallback(config.ID)

	ctx, cancel := context.WithCancel(m.ctx)
	vs := &VLANScanner{
		NetworkId: config.ID,
		Scanner:   arpScanner,
		Config:    config,
		Status:    "running",
		Cancel:    cancel,
		IsRunning: true,
	}
	m.scanners[config.ID] = vs
	arpScanner.Start()
	log.Printf("[Network %d] Scan started — interface: %s, CIDR: %s, interval: %ds",
		config.ID, interfaceName, cidr, config.ScanIntervalSeconds)
	m.wg.Add(1)
	go m.monitorVLAN(ctx, vs)
	return nil
}
func (m *VLANScanManager) markAllDevicesOffline(networkID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `
        UPDATE discovered_devices
        SET device_status = 'offline'
        WHERE network_id = $1      
          AND device_status NOT IN ('offline', 'conflict')
    `
	result, err := m.db.GetPool().Exec(ctx, query, networkID)
	if err != nil {
		log.Printf("[Network %d] Failed to mark devices offline: %v", networkID, err)
		return
	}
	if result.RowsAffected() > 0 {
		log.Printf("[Network %d] Marked %d devices as offline", networkID, result.RowsAffected())
	}
}
func (m *VLANScanManager) StopVLANScan(networkID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pendingMu.Lock()
	delete(m.pendingVLANs, networkID)
	m.pendingMu.Unlock()
	vs, exists := m.scanners[networkID]
	if !exists {
		return fmt.Errorf("no scanner found for network %d", networkID)
	}
	if !vs.IsRunning {
		return fmt.Errorf("scanner not running for network %d", networkID)
	}
	vs.Cancel()
	cidr, _ := m.GetCIDRFromConfig(vs.Config)
	if cidr != "" {
		ifaceName, _ := m.DetectInterfaceForCIDR(cidr, networkID)
		if ifaceName != "" && !m.interfaceStillInUse(ifaceName, networkID) {
			if hd, ok := m.hostnameDiscovery[ifaceName]; ok {
				hd.Stop()
				delete(m.hostnameDiscovery, ifaceName)
			}
		}
	}
	log.Printf("[Network %d] Scan stopped", networkID)
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
		log.Printf("[VLAN %d] ->Recovered: %s", config.VLANId, config.VLANName)
		time.Sleep(500 * time.Millisecond)
	}
	m.dbScanner.Start()
	log.Printf("[DB-ARP] ->DB-based unicast ARP scanner started")
	m.wg.Add(1)
	go m.retryPendingVLANs()
	log.Printf("[RETRY] ->  Infinite retry loop started (runs forever, checks every 30s)")

	return nil
}
func (m *VLANScanManager) queueForRetry(config *utils.VLANNetwork) {
	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()
	m.pendingVLANs[config.VLANId] = config
}
func (m *VLANScanManager) retryPendingVLANs() {
	defer m.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	log.Printf("[RETRY] ->  Infinite retry loop started (checks every 30s)")
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
				if retryCount%10 == 0 {
					log.Printf("[RETRY] arping  #%d - No pending VLANs", retryCount)
				}
				continue
			}
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
				dbConfig, err := m.db.GetNetworkByID(context.Background(), config.ID)
				if err != nil {
					log.Printf("[RETRY] VLAN %d deleted from database, removing from retry queue",
						config.VLANId)

					m.pendingMu.Lock()
					delete(m.pendingVLANs, config.VLANId)
					m.pendingMu.Unlock()
					continue
				}

				if !dbConfig.MonitoringEnabled {
					log.Printf("[RETRY] VLAN %d monitoring disabled, removing from retry queue",
						config.VLANId)

					m.pendingMu.Lock()
					delete(m.pendingVLANs, config.VLANId)
					m.pendingMu.Unlock()
					continue
				}

				m.mu.Lock()
				err = m.startVLANScanLocked(dbConfig)
				m.mu.Unlock()

				if err != nil {
					log.Printf("[RETRY] VLAN %d still unavailable: %v (will retry in 30s)",
						config.VLANId, err)
					continue
				}
				m.pendingMu.Lock()
				delete(m.pendingVLANs, config.VLANId)
				m.pendingMu.Unlock()
				log.Printf("[RETRY] ->VLAN %d started successfully", config.VLANId)
			}
			m.pendingMu.Lock()
			remaining := len(m.pendingVLANs)
			m.pendingMu.Unlock()

			if remaining > 0 {
				log.Printf("[RETRY] %d VLAN(s) still pending, will retry in 30s", remaining)
			} else {
				log.Printf("[RETRY] ->All pending VLANs started successfully")
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
	if m.dbScanner != nil {
		m.dbScanner.Stop()
		log.Printf("  Stopped DB-based ARP scanner")
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

		case EventCameOnline1:
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

func (m *VLANScanManager) monitorVLAN(ctx context.Context, vs *VLANScanner) {
	defer m.wg.Done()

	ticker := time.NewTicker(40 * time.Second) // ->changed to 40s
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
			dbConfig, err := m.db.GetNetworkByID(context.Background(), vs.Config.ID)
			if err != nil {
				log.Printf("[VLAN %d]  Deleted from DB, stopping", vs.NetworkId)
				m.cleanupVLAN(vs.NetworkId, "deleted")
				return
			}

			if !dbConfig.MonitoringEnabled {
				log.Printf("[VLAN %d]  Monitoring disabled", vs.NetworkId)
				m.cleanupVLAN(vs.NetworkId, "disabled")
				return
			}

			cidr, err := m.GetCIDRFromConfig(vs.Config)
			if err != nil || cidr == "" {
				log.Printf("[VLAN %d] CIDR error: %v", vs.NetworkId, err)
				continue
			}

			ifaceName, err := m.DetectInterfaceForCIDR(cidr, vs.NetworkId)
			if err != nil {
				consecutiveFailures++
				log.Printf("[VLAN %d]  Interface DOWN (%d): %v",
					vs.NetworkId, consecutiveFailures, err)

				if vs.IsRunning {
					log.Printf("[VLAN %d] Stopping scanner", vs.NetworkId)
					vs.Scanner.Stop()
					vs.IsRunning = false
					vs.Status = "interface_down"
					m.markAllDevicesOffline(vs.NetworkId)
				}

				continue
			}
			if !vs.IsRunning {
				log.Printf("[VLAN %d]  Interface recovered, recreating scanner", vs.NetworkId)

				if err := m.recreateScanner(vs, cidr, ifaceName); err != nil {
					log.Printf("[VLAN %d] Failed to recreate scanner: %v", vs.NetworkId, err)
					continue
				}

				vs.IsRunning = true
			}
			if consecutiveFailures > 0 {
				log.Printf("[VLAN %d] ->Recovered after %d failures",
					vs.NetworkId, consecutiveFailures)
			}

			consecutiveFailures = 0
			vs.Status = "running"
			for _, host := range vs.Scanner.GetHosts() {
				m.saveHostToDB(vs.NetworkId, host)
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

	arpScanner.OnARPEvent = m.buildEventCallback(vs.NetworkId)

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

	log.Printf("[VLAN %d] ->Cleaned up (%s)", vlanID, reason)
}

func (m *VLANScanManager) requeueForRetry(config *utils.VLANNetwork, vlanID int) {
	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()
	m.pendingVLANs[vlanID] = config
	m.mu.Lock()
	delete(m.scanners, vlanID)
	m.mu.Unlock()

	log.Printf("[VLAN %d] ->Added to infinite retry queue (30s interval)", vlanID)
}

func (m *VLANScanManager) ensureRetryLoopRunning() {
	select {
	case <-m.ctx.Done():
		return
	default:
		m.wg.Add(1)
		go m.retryPendingVLANs()
	}
}

func (m *VLANScanManager) GetAllStatuses() map[int]*VLANScanner {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[int]*VLANScanner, len(m.scanners))
	for k, v := range m.scanners {
		out[k] = v
	}
	return out
}
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
			return "", fmt.Errorf("static mode requires CIDR for network %d (%s)",
				config.ID, config.InterfaceName)
		}
		return *config.CIDRFull, nil

	case "auto", "dhcp":

		if config.CIDRFull != nil && *config.CIDRFull != "" {
			return *config.CIDRFull, nil
		}

		if config.InterfaceName == "" {
			return "", fmt.Errorf("no interface name for network %d", config.ID)
		}

		cidr, err := m.detectCIDRFromInterface(config.InterfaceName)
		if err != nil {
			return "", fmt.Errorf("could not detect CIDR for interface %s: %w",
				config.InterfaceName, err)
		}

		log.Printf("[Network %d] Live-detected CIDR for %s: %s",
			config.ID, config.InterfaceName, cidr)
		return cidr, nil

	default:
		return "", fmt.Errorf("invalid network mode '%s' for network %d",
			config.NetworkMode, config.ID)
	}
}
func (m *VLANScanManager) detectCIDRFromInterface(interfaceName string) (string, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", fmt.Errorf("interface %s not found: %w", interfaceName, err)
	}

	if iface.Flags&net.FlagUp == 0 {
		return "", fmt.Errorf("interface %s is down", interfaceName)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("failed to get addresses for %s: %w", interfaceName, err)
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.To4() == nil {
			continue
		}
		_, network, err := net.ParseCIDR(ipNet.String())
		if err != nil {
			continue
		}
		return network.String(), nil
	}

	return "", fmt.Errorf("no IPv4 address found on interface %s", interfaceName)
}
func (m *VLANScanManager) saveHostToDB(networkID int, host *Host) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deviceStatus := string(host.Status)
	if isBroadcastMAC(host.MAC) {
		deviceStatus = "conflict"
	}

	device := &utils.DiscoveredDevice{
		NetworkId:    networkID,
		IPAddress:    host.IP.String(),
		MACAddress:   host.MAC.String(),
		Hostname:     host.Hostname,
		Vendor:       host.Vendor,
		DeviceStatus: deviceStatus,
		FirstSeen:    host.FirstSeen,
		LastSeen:     host.LastSeen,
	}

	if err := m.db.UpsertDevice(ctx, device); err != nil {
		log.Printf("[Network %d] Failed to save device %s: %v", networkID, host.IP, err)
	}
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func (m *VLANScanManager) registerHostnameCallbackOnce() {
	hostnameCallbackRegisteredOnce.Do(func() {
		RegisterHostnameCallback(func(ip, hostname, proto string) {
			m.onHostnameDiscovered(ip, hostname, proto)
		})
		log.Printf("[DISCOVERY] Hostname→DB callback registered")
	})
}
func (m *VLANScanManager) onHostnameDiscovered(ip, hostname, proto string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for vlanID, vs := range m.scanners {
		if !vs.IsRunning {
			continue
		}
		vs.Scanner.HostMutex.Lock()
		host, exists := vs.Scanner.HostMap[ip]
		if exists && host.Hostname != hostname {
			host.Hostname = hostname
			log.Printf("[%s] Updated hostname for %s → %s (VLAN %d)",
				proto, ip, hostname, vlanID)
			hostCopy := copyHost(host)
			vs.Scanner.HostMutex.Unlock()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			deviceStatus := string(hostCopy.Status)
			if isBroadcastMAC(hostCopy.MAC) {
				deviceStatus = "conflict"
			}

			device := &utils.DiscoveredDevice{
				NetworkId:    vlanID,
				IPAddress:    hostCopy.IP.String(),
				MACAddress:   hostCopy.MAC.String(),
				Hostname:     hostname,
				Vendor:       hostCopy.Vendor,
				DeviceStatus: deviceStatus,
				FirstSeen:    hostCopy.FirstSeen,
				LastSeen:     hostCopy.LastSeen,
			}
			if err := m.db.UpsertDevice(ctx, device); err != nil {
				log.Printf("[VLAN %d] Failed to save hostname update for %s: %v",
					vlanID, ip, err)
			}
			return
		}
		vs.Scanner.HostMutex.Unlock()
	}
}

func (m *VLANScanManager) GetParentInterface() string {
	return m.parentInterface
}
