// unicast arp request for offline and online update for avoid broadcast flooding
// subnet broadcast every 30sec
// inmemory arp cache
// arp db fetch and request
// passive and active-scan
// inspired by arpwatch - https://ee.lbl.gov/downloads/arpwatch/
package vlan

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/mascarenhasmelson/gomotz/utils"
	"github.com/mdlayher/arp"
	"github.com/mdlayher/ethernet"
)

type ARPEvent int

const (
	EventNewDevice    ARPEvent = 2
	EventIPChange     ARPEvent = 1
	EventIPConflict   ARPEvent = 3
	EventIPConflictND ARPEvent = 4
	EventMACChange    ARPEvent = 5
	EventMACChangeND  ARPEvent = 6
	EventCameOnline1  ARPEvent = 7
	EventCameOnline2  ARPEvent = 8
	EventWentOffline  ARPEvent = 9
)

type DBARPScanner struct {
	db       *PostgresDB
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	interval time.Duration
}

type ARPEventCallback func(event ARPEvent, host *Host, oldHost *Host)
type HostStatus string

const (
	StatusOnline  HostStatus = "online"
	StatusOffline HostStatus = "offline"
	StatusNew     HostStatus = "new"
)

type Host struct {
	IP        net.IP
	MAC       net.HardwareAddr
	Hostname  string
	Vendor    string
	FirstSeen time.Time
	LastSeen  time.Time
	Status    HostStatus
	flag      bool
}

type VendorLookup interface {
	GetVendorByOUI(ctx context.Context, oui string) (*utils.MACVendor, error)
	SaveVendor(ctx context.Context, vendor *utils.MACVendor) error
	UpdateVendorLastSeen(ctx context.Context, oui string) error
}

type ARPScanner struct {
	iface             *net.Interface
	subnet            *net.IPNet
	localIP           net.IP
	localAddr         netip.Addr
	localMAC          net.HardwareAddr
	client            *arp.Client
	HostMap           map[string]*Host
	HostMutex         sync.RWMutex
	scanInterval      time.Duration
	replyWindow       time.Duration
	fullSweepInterval time.Duration
	dupCheckInterval  time.Duration
	pendingConflicts  map[string][]net.HardwareAddr
	conflictMu        sync.Mutex
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	hostnameDiscovery *HostnameDiscovery
	vendorLookup      VendorLookup
	httpClient        *http.Client
	OnARPEvent        ARPEventCallback
}

func NewARPScanner(subnetCIDR string, scanInterval time.Duration) (*ARPScanner, error) {
	_, subnet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}
	iface, localIP, err := findInterfaceForSubnet(subnet)
	if err != nil {
		return nil, err
	}
	localAddr, ok := netip.AddrFromSlice(localIP.To4())
	if !ok {
		return nil, fmt.Errorf("failed to convert local IP to netip.Addr")
	}
	client, err := arp.Dial(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to create ARP client: %w", err)
	}
	log.Printf("[ARP] Interface: %s | IP: %s | Subnet: %s", iface.Name, localIP, subnet)
	ctx, cancel := context.WithCancel(context.Background())

	return &ARPScanner{
		iface:             iface,
		subnet:            subnet,
		localIP:           localIP,
		localAddr:         localAddr,
		localMAC:          iface.HardwareAddr,
		client:            client,
		HostMap:           make(map[string]*Host),
		scanInterval:      scanInterval,
		replyWindow:       2 * time.Second,
		fullSweepInterval: scanInterval * 10,
		dupCheckInterval:  scanInterval * 5,
		pendingConflicts:  make(map[string][]net.HardwareAddr),
		ctx:               ctx,
		cancel:            cancel,
		httpClient:        &http.Client{Timeout: 5 * time.Second},
	}, nil
}

func (s *ARPScanner) SetVendorLookup(v VendorLookup) {
	s.vendorLookup = v
}
func (s *ARPScanner) SetHostnameDiscovery(h *HostnameDiscovery) {
	s.hostnameDiscovery = h
}

func (s *ARPScanner) Start() {
	log.Printf("[ARP] Starting for %s (scan: %v, full sweep: %v, dup check: %v)",
		s.subnet, s.scanInterval, s.fullSweepInterval, s.dupCheckInterval)
	s.wg.Add(1)
	go s.receiver()
	time.Sleep(200 * time.Millisecond)
	s.wg.Add(1)
	go s.sender()
}

func (s *ARPScanner) Stop() {
	s.cancel()
	s.wg.Wait()
	s.client.Close()
	s.HostMutex.Lock()
	for _, host := range s.HostMap {
		if host.Status != StatusOffline {
			host.Status = StatusOffline
		}
	}
	s.HostMutex.Unlock()

	log.Printf("[ARP] Stopped for %s", s.subnet)
}

func (s *ARPScanner) sender() {
	defer s.wg.Done()
	log.Printf("ARP 1 initial full subnet scan: %s", s.subnet)
	s.fullSubnetSweep()
	targetedTicker := time.NewTicker(s.scanInterval)
	fullSweepTicker := time.NewTicker(s.fullSweepInterval)
	dupCheckTicker := time.NewTicker(s.dupCheckInterval)
	defer targetedTicker.Stop()
	defer fullSweepTicker.Stop()
	defer dupCheckTicker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-targetedTicker.C:
			s.targetedScan()
		case <-fullSweepTicker.C:
			log.Printf("ARP 3  periodic full sweep (new device discovery): %s", s.subnet)
			s.fullSubnetSweep()
		case <-dupCheckTicker.C:
			s.duplicateIPSweep()
		}
	}
}

func (s *ARPScanner) fullSubnetSweep() {
	targets := s.generateTargetIPs()
	sent := 0
	for _, ip := range targets {
		if s.ctx.Err() != nil {
			return
		}
		if err := s.sendARPRequest(ip, ethernet.Broadcast); err == nil {
			sent++
		}
		time.Sleep(200 * time.Millisecond)
	}
	log.Printf("ARP sent %d/%d requests — waiting %v for replies",
		sent, len(targets), s.replyWindow)

	select {
	case <-time.After(s.replyWindow):
	case <-s.ctx.Done():
		return
	}

	s.logHostSummary("Full sweep complete")
}

func (s *ARPScanner) targetedScan() {
	s.HostMutex.Lock()
	hosts := make([]*Host, 0, len(s.HostMap))
	for _, h := range s.HostMap {
		h.flag = false
		hosts = append(hosts, h)
	}
	s.HostMutex.Unlock()
	if len(hosts) == 0 {
		return
	}
	for _, h := range hosts {
		if s.ctx.Err() != nil {
			return
		}
		if err := s.sendARPRequest(h.IP, cloneMAC(h.MAC)); err != nil {
			log.Printf("[ARP] Send failed for %s: %v", h.IP, err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	select {
	case <-time.After(s.replyWindow):
	case <-s.ctx.Done():
		return
	}
	s.checkOfflineByFlag()
}

func (s *ARPScanner) checkOfflineByFlag() {
	s.HostMutex.Lock()
	defer s.HostMutex.Unlock()
	for _, host := range s.HostMap {
		if host.flag {
			continue
		}
		if host.Status != StatusOnline {
			continue
		}
		if isBroadcastMAC(host.MAC) {
			continue
		}
		oldHost := copyHost(host)
		host.Status = StatusOffline
		log.Printf("[ARP] Went offline: %s (%s)", host.IP, host.MAC)
		if s.OnARPEvent != nil {
			go s.OnARPEvent(EventWentOffline, copyHost(host), oldHost)
		}
	}
}

func (s *ARPScanner) duplicateIPSweep() {
	s.HostMutex.RLock()
	knownIPs := make([]net.IP, 0, len(s.HostMap))
	for _, h := range s.HostMap {
		if h.Status == StatusOnline {
			knownIPs = append(knownIPs, cloneIP(h.IP))
		}
	}
	s.HostMutex.RUnlock()
	if len(knownIPs) == 0 {
		return
	}
	log.Printf("ARP Dup check sweep probing %d known IPs via broadcast", len(knownIPs))
	s.conflictMu.Lock()
	s.pendingConflicts = make(map[string][]net.HardwareAddr)
	s.conflictMu.Unlock()
	s.setDupSweepActive(true)

	for _, ip := range knownIPs {
		if s.ctx.Err() != nil {
			s.setDupSweepActive(false)
			return
		}
		s.sendARPRequest(ip, ethernet.Broadcast)
		time.Sleep(50 * time.Millisecond)
	}
	select {
	case <-time.After(s.replyWindow):
	case <-s.ctx.Done():
		s.setDupSweepActive(false)
		return
	}

	s.setDupSweepActive(false)
	s.evaluateDuplicates()
}
func (s *ARPScanner) evaluateDuplicates() {
	s.conflictMu.Lock()
	conflicts := s.pendingConflicts
	s.pendingConflicts = make(map[string][]net.HardwareAddr)
	s.conflictMu.Unlock()
	for ipStr, macs := range conflicts {
		unique := uniqueMACs(macs)
		if len(unique) < 2 {
			continue
		}
		log.Printf("[ARP] duplicate ip detected: %s is claimed by %d devices:", ipStr, len(unique))
		for _, mac := range unique {
			log.Printf("         MAC: %s", mac)
		}
		s.HostMutex.Lock()
		existing, exists := s.HostMap[ipStr]
		if exists {
			oldHost := copyHost(existing)
			existing.MAC = broadcastMAC()
			if s.OnARPEvent != nil {
				go s.OnARPEvent(EventIPConflictND, copyHost(existing), oldHost)
			}
		}
		s.HostMutex.Unlock()
	}
}

var dupSweepActive bool
var dupSweepMu sync.Mutex

func (s *ARPScanner) setDupSweepActive(v bool) {
	dupSweepMu.Lock()
	dupSweepActive = v
	dupSweepMu.Unlock()
}

func (s *ARPScanner) isDupSweepActive() bool {
	dupSweepMu.Lock()
	defer dupSweepMu.Unlock()
	return dupSweepActive
}

func (s *ARPScanner) receiver() {
	defer s.wg.Done()
	s.client.SetReadDeadline(time.Now().Add(time.Second))
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}
		packet, _, err := s.client.Read()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				s.client.SetReadDeadline(time.Now().Add(time.Second))
				continue
			}
			s.client.SetReadDeadline(time.Now().Add(time.Second))
			continue
		}
		if s.isDupSweepActive() {
			s.collectDupReply(packet)
		}
		s.processPacket(packet)
		s.client.SetReadDeadline(time.Now().Add(time.Second))
	}
}

func (s *ARPScanner) collectDupReply(packet *arp.Packet) {
	senderIP := net.IP(packet.SenderIP.AsSlice())
	senderMAC := net.HardwareAddr(packet.SenderHardwareAddr)
	if !s.subnet.Contains(senderIP) || senderIP.Equal(s.localIP) {
		return
	}
	ipStr := senderIP.String()
	s.conflictMu.Lock()
	defer s.conflictMu.Unlock()

	existing := s.pendingConflicts[ipStr]
	for _, m := range existing {
		if m.String() == senderMAC.String() {
			return
		}
	}
	s.pendingConflicts[ipStr] = append(existing, cloneMAC(senderMAC))
}

func (s *ARPScanner) processPacket(packet *arp.Packet) {
	senderIP := net.IP(packet.SenderIP.AsSlice())
	senderMAC := net.HardwareAddr(packet.SenderHardwareAddr)
	targetIP := net.IP(packet.TargetIP.AsSlice())
	if !s.subnet.Contains(senderIP) || senderIP.Equal(s.localIP) {
		return
	}
	isGarp := senderIP.Equal(targetIP)
	s.HostMutex.Lock()
	defer s.HostMutex.Unlock()
	ipStr := senderIP.String()
	existing, exists := s.HostMap[ipStr]
	if !exists {
		//new ip
		oldIP := s.removeMACFromOtherIP(senderMAC, ipStr)

		newHost := &Host{
			IP:        cloneIP(senderIP),
			MAC:       cloneMAC(senderMAC),
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
			Status:    StatusNew,
			flag:      true,
		}
		s.HostMap[ipStr] = newHost
		if oldIP != "" {
			log.Printf("[ARP] IP change: MAC %s moved from %s to %s", senderMAC, oldIP, senderIP)
			if s.OnARPEvent != nil {
				go s.OnARPEvent(EventIPChange, copyHost(newHost), &Host{IP: net.ParseIP(oldIP)})
			}
		} else {
			log.Printf("[ARP] New device: %s (%s)", senderIP, senderMAC)
			if s.OnARPEvent != nil {
				go s.OnARPEvent(EventNewDevice, copyHost(newHost), nil)
			}
		}
		go s.resolveHostnameAsync(cloneIP(senderIP), newHost)
		go s.resolveVendorAsync(cloneMAC(senderMAC), newHost)
		return
	}
	existing.flag = true
	existing.LastSeen = time.Now()
	macUnchanged := existing.MAC.String() == senderMAC.String()
	if macUnchanged {
		if isGarp {
			if existing.Status == StatusOnline {

				if s.OnARPEvent != nil {
					go s.OnARPEvent(EventCameOnline1, copyHost(existing), nil)
				}
			} else {
				// was offline garp came back online.
				oldHost := copyHost(existing)
				existing.Status = StatusOnline
				log.Printf("[ARP] Back online (gratuitous ARP): %s", senderIP)
				if s.OnARPEvent != nil {
					go s.OnARPEvent(EventCameOnline2, copyHost(existing), oldHost)
				}
			}
		} else if existing.Status != StatusOnline {
			oldHost := copyHost(existing)
			existing.Status = StatusOnline
			log.Printf("[ARP] Came online: %s (%s)", senderIP, senderMAC)
			if s.OnARPEvent != nil {
				go s.OnARPEvent(EventCameOnline2, copyHost(existing), oldHost)
			}
		}
		return
	}
	oldIP := s.removeMACFromOtherIP(senderMAC, ipStr)
	if existing.Status == StatusOnline {
		oldHost := copyHost(existing)
		existing.MAC = broadcastMAC()
		if oldIP != "" {
			log.Printf("[ARP] IP conflict + IP change: MAC %s on %s (was at %s)",
				senderMAC, senderIP, oldIP)
			if s.OnARPEvent != nil {
				go s.OnARPEvent(EventIPConflict, copyHost(existing), oldHost)
			}
		} else {
			log.Printf("[ARP] IP conflict new device: MAC %s on %s", senderMAC, senderIP)
			if s.OnARPEvent != nil {
				go s.OnARPEvent(EventIPConflictND, copyHost(existing), oldHost)
			}
		}
	} else {
		oldHost := copyHost(existing)
		existing.IP = cloneIP(senderIP)
		existing.MAC = cloneMAC(senderMAC)
		existing.Status = StatusOnline
		existing.LastSeen = time.Now()
		if oldIP != "" {
			log.Printf("[ARP] MAC change + IP change: MAC %s on %s (was at %s)",
				senderMAC, senderIP, oldIP)
			if s.OnARPEvent != nil {
				go s.OnARPEvent(EventMACChange, copyHost(existing), oldHost)
			}
		} else {
			log.Printf("[ARP] MAC change + new device: MAC %s on %s", senderMAC, senderIP)
			if s.OnARPEvent != nil {
				go s.OnARPEvent(EventMACChangeND, copyHost(existing), oldHost)
			}
		}
		go s.resolveVendorAsync(cloneMAC(senderMAC), existing)
	}
}

func (s *ARPScanner) removeMACFromOtherIP(mac net.HardwareAddr, skipIP string) string {
	for ipStr, host := range s.HostMap {
		if ipStr == skipIP {
			continue
		}
		if host.MAC.String() == mac.String() {
			delete(s.HostMap, ipStr)
			return ipStr
		}
	}
	return ""
}

func (s *ARPScanner) sendARPRequest(targetIP net.IP, dstMAC net.HardwareAddr) error {
	targetAddr, ok := netip.AddrFromSlice(targetIP.To4())
	if !ok {
		return fmt.Errorf("cannot convert %s to netip.Addr", targetIP)
	}
	pkt, err := arp.NewPacket(
		arp.OperationRequest,
		s.localMAC,
		s.localAddr,
		dstMAC,
		targetAddr,
	)
	if err != nil {
		return err
	}
	return s.client.WriteTo(pkt, dstMAC)
}

func (s *ARPScanner) generateTargetIPs() []net.IP {
	var ips []net.IP
	network := s.subnet.IP.Mask(s.subnet.Mask)
	broadcast := make(net.IP, len(network))
	for i := range network {
		broadcast[i] = network[i] | ^s.subnet.Mask[i]
	}
	for ip := incrementIP(network); s.subnet.Contains(ip); ip = incrementIP(ip) {
		if ip.Equal(network) || ip.Equal(broadcast) || ip.Equal(s.localIP) {
			continue
		}
		c := make(net.IP, len(ip))
		copy(c, ip)
		ips = append(ips, c)
	}
	return ips
}

func incrementIP(ip net.IP) net.IP {
	ip = ip.To4()
	v := binary.BigEndian.Uint32(ip)
	v++
	out := make(net.IP, 4)
	binary.BigEndian.PutUint32(out, v)
	return out
}

func (s *ARPScanner) resolveHostnameAsync(ip net.IP, host *Host) {
	var hostname string
	if s.hostnameDiscovery != nil {
		if h := s.hostnameDiscovery.GetHostname(ip.String()); h != "" && isValidHostname(h) {
			hostname = h
		}
	}
	if hostname == "" {
		if h := resolveHostname(ip); h != "" && isValidHostname(h) {
			hostname = h
		}
	}
	if hostname == "" {
		return
	}
	s.HostMutex.Lock()
	host.Hostname = hostname
	s.HostMutex.Unlock()
	log.Printf("[HOSTNAME] %s → %s", ip, hostname)
}

func (s *ARPScanner) resolveVendorAsync(mac net.HardwareAddr, host *Host) {
	vendor := s.getVendor(mac)
	if vendor == "" {
		return
	}
	s.HostMutex.Lock()
	host.Vendor = vendor
	s.HostMutex.Unlock()
	log.Printf("[VENDOR] %s → %s", mac, vendor)
}

func (s *ARPScanner) getVendor(mac net.HardwareAddr) string {
	if s.vendorLookup == nil || len(mac) < 3 {
		return ""
	}
	oui := strings.ToUpper(strings.ReplaceAll(mac.String()[:8], ":", ""))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if v, err := s.vendorLookup.GetVendorByOUI(ctx, oui); err == nil && v != nil {
		s.vendorLookup.UpdateVendorLastSeen(ctx, oui)
		return v.VendorName
	}
	name, err := s.fetchVendorFromAPI(mac)
	if err != nil {
		return ""
	}
	_ = s.vendorLookup.SaveVendor(ctx, &utils.MACVendor{
		OUI:            oui,
		VendorName:     name,
		FetchedFromAPI: true,
		LastSeen:       time.Now(),
	})
	return name
}

func (s *ARPScanner) fetchVendorFromAPI(mac net.HardwareAddr) (string, error) {
	url := fmt.Sprintf("https://api.macvendors.com/%s",
		strings.ReplaceAll(mac.String(), ":", ""))
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (s *ARPScanner) GetHosts() []*Host {
	s.HostMutex.RLock()
	defer s.HostMutex.RUnlock()
	out := make([]*Host, 0, len(s.HostMap))
	for _, h := range s.HostMap {
		out = append(out, copyHost(h))
	}
	return out
}

func (s *ARPScanner) GetHostCount() int {
	s.HostMutex.RLock()
	defer s.HostMutex.RUnlock()
	return len(s.HostMap)
}

func (s *ARPScanner) logHostSummary(label string) {
	s.HostMutex.RLock()
	defer s.HostMutex.RUnlock()
	online := 0
	for _, h := range s.HostMap {
		if h.Status != StatusOffline {
			online++
		}
	}
	log.Printf("[ARP] %s — %d hosts in %s", label, online, s.subnet)
}

func uniqueMACs(macs []net.HardwareAddr) []net.HardwareAddr {
	seen := make(map[string]bool)
	var out []net.HardwareAddr
	for _, m := range macs {
		if !seen[m.String()] {
			seen[m.String()] = true
			out = append(out, m)
		}
	}
	return out
}

func cloneIP(ip net.IP) net.IP {
	c := make(net.IP, len(ip))
	copy(c, ip)
	return c
}

func cloneMAC(mac net.HardwareAddr) net.HardwareAddr {
	c := make(net.HardwareAddr, len(mac))
	copy(c, mac)
	return c
}

func copyHost(h *Host) *Host {
	c := *h
	c.IP = cloneIP(h.IP)
	c.MAC = cloneMAC(h.MAC)
	return &c
}

func broadcastMAC() net.HardwareAddr {
	return net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
}

func isBroadcastMAC(mac net.HardwareAddr) bool {
	if len(mac) != 6 {
		return false
	}
	for _, b := range mac {
		if b != 0xff {
			return false
		}
	}
	return true
}

func isValidHostname(s string) bool {
	if s == "" || len(s) > 255 || strings.HasPrefix(s, "{") {
		return false
	}
	special := 0
	for _, r := range s {
		if r < 32 || r > 126 {
			return false
		}
		if r == '{' || r == '}' || r == '[' || r == ']' || r == '\\' || r == '"' {
			special++
		}
	}
	return special <= 2
}

func findInterfaceForSubnet(subnet *net.IPNet) (*net.Interface, net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.To4() != nil && subnet.Contains(ip) {
				return &iface, ip, nil
			}
		}
	}
	return nil, nil, fmt.Errorf("no interface found for subnet %s", subnet)
}

func resolveHostname(ip net.IP) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	names, err := net.DefaultResolver.LookupAddr(ctx, ip.String())
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

type DatabaseVendorLookup struct{ db *PostgresDB }

func NewDatabaseVendorLookup(db *PostgresDB) *DatabaseVendorLookup {
	return &DatabaseVendorLookup{db: db}
}
func (d *DatabaseVendorLookup) GetVendorByOUI(ctx context.Context, oui string) (*utils.MACVendor, error) {
	return d.db.GetVendorByOUI(ctx, oui)
}
func (d *DatabaseVendorLookup) SaveVendor(ctx context.Context, v *utils.MACVendor) error {
	return d.db.UpsertVendor(ctx, v)
}
func (d *DatabaseVendorLookup) UpdateVendorLastSeen(ctx context.Context, oui string) error {
	return d.db.UpdateVendorLastSeen(ctx, oui)
}

// db scan

func NewDBARPScanner(db *PostgresDB, interval time.Duration) *DBARPScanner {
	ctx, cancel := context.WithCancel(context.Background())
	return &DBARPScanner{
		db:       db,
		ctx:      ctx,
		cancel:   cancel,
		interval: interval,
	}
}

func (s *DBARPScanner) Start() {
	log.Printf("[DB-ARP] Starting DB-based unicast ARP scanner (interval: %v)", s.interval)
	s.wg.Add(1)
	go s.run()
}

func (s *DBARPScanner) Stop() {
	s.cancel()
	s.wg.Wait()
	log.Printf("[DB-ARP] Stopped")
}

func (s *DBARPScanner) run() {
	defer s.wg.Done()
	s.scanAllDevices()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.scanAllDevices()
		}
	}
}

func (s *DBARPScanner) scanAllDevices() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	devices, err := s.db.GetAllDevicesForARPScan(ctx)
	if err != nil {
		log.Printf("[DB-ARP] Failed to fetch devices: %v", err)
		return
	}

	if len(devices) == 0 {
		return
	}
	log.Printf("[DB-ARP] Scanning %d devices from database", len(devices))
	byInterface := make(map[string][]*ARPScanDevice)
	for _, d := range devices {
		byInterface[d.InterfaceName] = append(byInterface[d.InterfaceName], d)
	}
	var wg sync.WaitGroup
	for ifaceName, devs := range byInterface {
		wg.Add(1)
		go func(ifaceName string, devs []*ARPScanDevice) {
			defer wg.Done()
			s.scanInterface(ifaceName, devs)
		}(ifaceName, devs)
	}
	wg.Wait()

	log.Printf("[DB-ARP] Scan cycle complete — %d devices checked", len(devices))
}

func (s *DBARPScanner) scanInterface(ifaceName string, devices []*ARPScanDevice) {
	// get the interface
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Printf("[DB-ARP] Interface %s not found: %v", ifaceName, err)
		s.markDevicesOffline(devices, "interface not found")
		return
	}
	if iface.Flags&net.FlagUp == 0 {
		log.Printf("[DB-ARP] Interface %s is down", ifaceName)
		s.markDevicesOffline(devices, "interface down")
		return
	}
	localIP, err := getInterfaceIPv4(iface)
	if err != nil {
		log.Printf("[DB-ARP] No IPv4 on %s: %v", ifaceName, err)
		return
	}
	localAddr, ok := netip.AddrFromSlice(localIP.To4())
	if !ok {
		return
	}
	client, err := arp.Dial(iface)
	if err != nil {
		log.Printf("[DB-ARP] ARP dial failed for %s: %v", ifaceName, err)
		return
	}
	defer client.Close()

	log.Printf("[DB-ARP] Scanning %d devices on %s", len(devices), ifaceName)

	//send unicast arp one by one
	replied := make(map[string]bool)
	var repliedMu sync.Mutex
	replyCtx, replyCancel := context.WithCancel(s.ctx)
	defer replyCancel()

	var replyWg sync.WaitGroup
	replyWg.Add(1)
	go func() {
		defer replyWg.Done()
		s.collectReplies(replyCtx, client, iface, replied, &repliedMu)
	}()

	for _, device := range devices {
		if s.ctx.Err() != nil {
			break
		}
		targetIP := net.ParseIP(device.IPAddress).To4()
		if targetIP == nil {
			continue
		}
		targetAddr, ok := netip.AddrFromSlice(targetIP)
		if !ok {
			continue
		}
		targetMAC, err := net.ParseMAC(device.MACAddress)
		if err != nil {
			targetMAC = ethernet.Broadcast
		}
		pkt, err := arp.NewPacket(
			arp.OperationRequest,
			iface.HardwareAddr,
			localAddr,
			targetMAC,
			targetAddr,
		)
		if err != nil {
			continue
		}
		if err := client.WriteTo(pkt, targetMAC); err != nil {
			log.Printf("[DB-ARP] Send failed for %s: %v", device.IPAddress, err)
		}

		time.Sleep(50 * time.Millisecond)
	}
	select {
	case <-time.After(3 * time.Second):
	case <-s.ctx.Done():
		replyCancel()
		replyWg.Wait()
		return
	}

	replyCancel()
	replyWg.Wait()
	s.updateDeviceStatuses(devices, replied)
}

func (s *DBARPScanner) collectReplies(
	ctx context.Context,
	client *arp.Client,
	iface *net.Interface,
	replied map[string]bool,
	mu *sync.Mutex,
) {
	client.SetReadDeadline(time.Now().Add(time.Second))

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		packet, _, err := client.Read()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				client.SetReadDeadline(time.Now().Add(time.Second))
				continue
			}
			return
		}

		if packet.Operation != arp.OperationReply {
			continue
		}

		senderIP := net.IP(packet.SenderIP.AsSlice()).String()

		mu.Lock()
		replied[senderIP] = true
		mu.Unlock()

		client.SetReadDeadline(time.Now().Add(time.Second))
	}
}

func (s *DBARPScanner) updateDeviceStatuses(devices []*ARPScanDevice, replied map[string]bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	onlineCount := 0
	offlineCount := 0

	for _, device := range devices {
		isOnline := replied[device.IPAddress]
		var newStatus string

		if isOnline {
			newStatus = "online"
			onlineCount++
		} else {

			if device.CurrentStatus == "conflict" {
				continue
			}
			newStatus = "offline"
			offlineCount++
		}
		if device.CurrentStatus == newStatus {
			continue
		}

		if err := s.db.UpdateDeviceStatusByNetworkAndMAC(ctx, device.NetworkID, device.MACAddress, newStatus); err != nil {
			log.Printf("[DB-ARP] Failed to update %s (%s): %v",
				device.IPAddress, device.MACAddress, err)
			continue
		}

		log.Printf("[DB-ARP] %s (%s) on %s: %s → %s",
			device.IPAddress, device.MACAddress, device.InterfaceName,
			device.CurrentStatus, newStatus)
	}

	log.Printf("[DB-ARP] Updated: %d online, %d offline", onlineCount, offlineCount)
}

func (s *DBARPScanner) markDevicesOffline(devices []*ARPScanDevice, reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, device := range devices {
		if device.CurrentStatus == "offline" || device.CurrentStatus == "conflict" {
			continue
		}
		if err := s.db.UpdateDeviceStatusByNetworkAndMAC(ctx, device.NetworkID, device.MACAddress, "offline"); err != nil {
			log.Printf("[DB-ARP] Failed to mark offline %s: %v", device.IPAddress, err)
		}
	}
	log.Printf("[DB-ARP] Marked %d devices offline (%s)", len(devices), reason)
}

func getInterfaceIPv4(iface *net.Interface) (net.IP, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip != nil && ip.To4() != nil {
			return ip.To4(), nil
		}
	}
	return nil, fmt.Errorf("no IPv4 address on %s", iface.Name)
}
