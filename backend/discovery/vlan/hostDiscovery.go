// package vlan

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"sync"
// 	"time"
// )

// // ============================================
// // HOSTNAME DISCOVERY
// // Runs mDNS and SSDP discovery on a single network interface.
// // Shared across all VLANs that use the same physical interface.
// // ============================================

// type HostnameDiscovery struct {
// 	iface       *InterfaceInfo
// 	ssdpScanner *SSDPScanner
// 	mdnsScanner *MDNSScanner
// 	mu          sync.RWMutex
// 	cache       map[string]string
// 	stats       DiscoveryStats
// 	ctx         context.Context
// 	cancel      context.CancelFunc
// }

// type DiscoveryStats struct {
// 	TotalScans       int
// 	SSDPDiscoveries  int
// 	MDNSDiscoveries  int
// 	LastScanTime     time.Time
// 	LastScanDuration time.Duration
// }

// func NewHostnameDiscovery(interfaceName string) (*HostnameDiscovery, error) {
// 	iface, err := GetInterfaceByName(interfaceName)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get interface %s: %w", interfaceName, err)
// 	}

// 	ssdpScanner, err := NewSSDP(iface)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create SSDP scanner: %w", err)
// 	}

// 	mdnsScanner, err := NewMDNS(iface)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create mDNS scanner: %w", err)
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())

// 	hd := &HostnameDiscovery{
// 		iface:       iface,
// 		ssdpScanner: ssdpScanner,
// 		mdnsScanner: mdnsScanner,
// 		cache:       make(map[string]string),
// 		ctx:         ctx,
// 		cancel:      cancel,
// 	}

// 	log.Printf("[DISCOVERY] Initialized on %s (%s)", interfaceName, iface.IPv4.String())
// 	return hd, nil
// }

// // ScanSubnet runs SSDP and mDNS concurrently and populates the hostname cache.
// func (h *HostnameDiscovery) ScanSubnet() error {
// 	startTime := time.Now()
// 	results := make(chan *Device, 256)
// 	var wg sync.WaitGroup

// 	ctx, cancel := context.WithTimeout(h.ctx, 10*time.Second)
// 	defer cancel()

// 	ssdpCount, mdnsCount := 0, 0

// 	// ->ADD: Verbose logging
// 	log.Printf("[DISCOVERY] ========== Starting scan on %s ==========", h.iface.Interface.Name)
// 	log.Printf("[DISCOVERY] Interface IP: %s", h.iface.IPv4.String())
// 	log.Printf("[DISCOVERY] Timeout: 10 seconds")

// 	wg.Add(2)

// 	// SSDP scanner
// 	go func() {
// 		defer wg.Done()
// 		log.Printf("[SSDP] Starting scanner...")
// 		if err := h.ssdpScanner.Scan(ctx, results); err != nil &&
// 			err != context.DeadlineExceeded && err != context.Canceled {
// 			log.Printf("[SSDP] ❌ Scan error: %v", err)
// 		} else {
// 			log.Printf("[SSDP] ->Scan completed")
// 		}
// 	}()

// 	// mDNS scanner
// 	go func() {
// 		defer wg.Done()
// 		log.Printf("[mDNS] Starting scanner...")
// 		if err := h.mdnsScanner.Scan(ctx, results); err != nil &&
// 			err != context.DeadlineExceeded && err != context.Canceled {
// 			log.Printf("[mDNS] ❌ Scan error: %v", err)
// 		} else {
// 			log.Printf("[mDNS] ->Scan completed")
// 		}
// 	}()

// 	go func() {
// 		wg.Wait()
// 		close(results)
// 		log.Printf("[DISCOVERY] Both scanners finished, processing results...")
// 	}()

// 	discovered := 0
// 	for device := range results {
// 		if device.IP == nil || device.Name == "" {
// 			log.Printf("[%s] ⚠️  Skipping device: IP=%v, Name=%s",
// 				device.Proto, device.IP, device.Name)
// 			continue
// 		}

// 		ipStr := device.IP.String()

// 		h.mu.Lock()
// 		if _, exists := h.cache[ipStr]; !exists {
// 			log.Printf("[%s] ->NEW: %s → %s", device.Proto, ipStr, device.Name)
// 		} else {
// 			log.Printf("[%s] -> UPDATE: %s → %s", device.Proto, ipStr, device.Name)
// 		}
// 		h.cache[ipStr] = device.Name
// 		h.mu.Unlock()

// 		discovered++
// 		switch device.Proto {
// 		case "ssdp":
// 			ssdpCount++
// 		case "mdns":
// 			mdnsCount++
// 		}
// 	}

// 	duration := time.Since(startTime)

// 	h.mu.Lock()
// 	h.stats.TotalScans++
// 	h.stats.SSDPDiscoveries += ssdpCount
// 	h.stats.MDNSDiscoveries += mdnsCount
// 	h.stats.LastScanTime = time.Now()
// 	h.stats.LastScanDuration = duration
// 	h.mu.Unlock()

// 	// ->ALWAYS log, even if 0
// 	log.Printf("[DISCOVERY] ========== Scan complete on %s ==========", h.iface.Interface.Name)
// 	log.Printf("[DISCOVERY] Found: %d devices (SSDP: %d, mDNS: %d)",
// 		discovered, ssdpCount, mdnsCount)
// 	log.Printf("[DISCOVERY] Duration: %v", duration.Round(time.Millisecond))
// 	log.Printf("[DISCOVERY] Total cache size: %d", len(h.cache))
// 	log.Printf("[DISCOVERY] ================================================")

// 	return nil
// }

// // GetHostname returns the cached hostname for the given IP, or "".
// func (h *HostnameDiscovery) GetHostname(ip string) string {
// 	h.mu.RLock()
// 	defer h.mu.RUnlock()
// 	return h.cache[ip]
// }

// func (h *HostnameDiscovery) GetStats() DiscoveryStats {
// 	h.mu.RLock()
// 	defer h.mu.RUnlock()
// 	return h.stats
// }

// func (h *HostnameDiscovery) GetCacheSize() int {
// 	h.mu.RLock()
// 	defer h.mu.RUnlock()
// 	return len(h.cache)
// }

// func (h *HostnameDiscovery) GetAllCached() map[string]string {
// 	h.mu.RLock()
// 	defer h.mu.RUnlock()
// 	out := make(map[string]string, len(h.cache))
// 	for k, v := range h.cache {
// 		out[k] = v
// 	}
// 	return out
// }

// func (h *HostnameDiscovery) ClearCache() {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()
// 	h.cache = make(map[string]string)
// }

// // Start runs ScanSubnet periodically until Stop is called.
// func (h *HostnameDiscovery) Start(interval time.Duration) {
// 	log.Printf("[DISCOVERY] Started (interval: %v)", interval)

// 	if err := h.ScanSubnet(); err != nil {
// 		log.Printf("[DISCOVERY] Initial scan failed: %v", err)
// 	}

// 	ticker := time.NewTicker(interval)
// 	defer ticker.Stop()

// 	scanCount := 0
// 	for {
// 		select {
// 		case <-h.ctx.Done():
// 			log.Printf("[DISCOVERY] Stopped (scans: %d, cache: %d)",
// 				scanCount, h.GetCacheSize())
// 			return

// 		case <-ticker.C:
// 			scanCount++
// 			if err := h.ScanSubnet(); err != nil {
// 				log.Printf("[DISCOVERY] Scan #%d failed: %v", scanCount, err)
// 			}
// 			if scanCount%10 == 0 {
// 				s := h.GetStats()
// 				log.Printf("[DISCOVERY] Stats — scans: %d, SSDP: %d, mDNS: %d, cache: %d",
// 					s.TotalScans, s.SSDPDiscoveries, s.MDNSDiscoveries, h.GetCacheSize())
// 			}
// 		}
// 	}
// }

// func (h *HostnameDiscovery) Stop() {
// 	log.Printf("[DISCOVERY] Stopping on %s...", h.iface.Interface.Name)
// 	h.cancel()
// }

// package vlan

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"sync"
// 	"time"
// )

// // ============================================
// // HOSTNAME DISCOVERY
// // ============================================

// type HostnameDiscovery struct {
// 	iface       *InterfaceInfo
// 	ssdpScanner *SSDPScanner
// 	mdnsScanner *MDNSScanner
// 	mu          sync.RWMutex
// 	cache       map[string]string
// 	stats       DiscoveryStats
// 	ctx         context.Context
// 	cancel      context.CancelFunc
// 	out         chan<- *Device
// }

// type DiscoveryStats struct {
// 	TotalScans       int
// 	SSDPDiscoveries  int
// 	MDNSDiscoveries  int
// 	LastScanTime     time.Time
// 	LastScanDuration time.Duration
// }

// // , out chan<- *Device
// func NewHostnameDiscovery(interfaceName string) (*HostnameDiscovery, error) {
// 	iface, err := GetInterfaceByName(interfaceName)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get interface %s: %w", interfaceName, err)
// 	}

// 	ssdpScanner, err := NewSSDP(iface)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create SSDP scanner: %w", err)
// 	}

// 	mdnsScanner, err := NewMDNS(iface)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create mDNS scanner: %w", err)
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())

// 	hd := &HostnameDiscovery{
// 		iface:       iface,
// 		ssdpScanner: ssdpScanner,
// 		mdnsScanner: mdnsScanner,
// 		cache:       make(map[string]string),
// 		ctx:         ctx,
// 		cancel:      cancel,
// 		// out:         out,
// 	}

// 	log.Printf("[DISCOVERY] Initialized on %s (%s)", interfaceName, iface.IPv4.String())
// 	return hd, nil
// }

// // ->Start continuous discovery (runs forever)
// // interval time.Duration
// func (h *HostnameDiscovery) Start() {
// 	log.Printf("[DISCOVERY] ->  Started continuous discovery on %s", h.iface.Interface.Name)

// 	results := make(chan *Device, 256)

// 	// ->Start both scanners as continuous listeners
// 	go func() {
// 		if err := h.ssdpScanner.Scan(h.ctx, results); err != nil &&
// 			err != context.Canceled {
// 			log.Printf("[SSDP] Scanner error: %v", err)
// 		}
// 	}()

// 	go func() {
// 		if err := h.mdnsScanner.Scan(h.ctx, results); err != nil &&
// 			err != context.Canceled {
// 			log.Printf("[mDNS] Scanner error: %v", err)
// 		}
// 	}()

// 	// ->Process results continuously
// 	discovered := 0
// 	for {
// 		select {
// 		case <-h.ctx.Done():
// 			log.Printf("[DISCOVERY] Stopped on %s (discovered: %d)",
// 				h.iface.Interface.Name, discovered)
// 			return

// 		case device, ok := <-results:
// 			if !ok {
// 				return
// 			}
// 			// if h.out != nil {
// 			// 	select {
// 			// 	case h.out <- device:
// 			// 	default:
// 			// 		log.Printf("[WARN] DB channel full, dropping %s", device.IP)
// 			// 	}
// 			// }
// 			if device.IP == nil || device.Name == "" {
// 				continue
// 			}

// 			ipStr := device.IP.String()

// 			h.mu.Lock()
// 			if _, exists := h.cache[ipStr]; !exists {
// 				log.Printf("[%s] ->Discovered: %s → %s", device.Proto, ipStr, device.Name)
// 				discovered++
// 			}
// 			h.cache[ipStr] = device.Name
// 			h.mu.Unlock()

// 			// Update stats
// 			h.mu.Lock()
// 			switch device.Proto {
// 			case "ssdp":
// 				h.stats.SSDPDiscoveries++
// 			case "mdns":
// 				h.stats.MDNSDiscoveries++
// 			}
// 			h.stats.LastScanTime = time.Now()
// 			h.mu.Unlock()
// 		}
// 	}
// }

// // GetHostname returns cached hostname
// func (h *HostnameDiscovery) GetHostname(ip string) string {
// 	h.mu.RLock()
// 	defer h.mu.RUnlock()
// 	return h.cache[ip]
// }

// func (h *HostnameDiscovery) GetStats() DiscoveryStats {
// 	h.mu.RLock()
// 	defer h.mu.RUnlock()
// 	return h.stats
// }

// func (h *HostnameDiscovery) GetCacheSize() int {
// 	h.mu.RLock()
// 	defer h.mu.RUnlock()
// 	return len(h.cache)
// }

// func (h *HostnameDiscovery) GetAllCached() map[string]string {
// 	h.mu.RLock()
// 	defer h.mu.RUnlock()
// 	out := make(map[string]string, len(h.cache))
// 	for k, v := range h.cache {
// 		out[k] = v
// 	}
// 	return out
// }

// func (h *HostnameDiscovery) ClearCache() {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()
// 	h.cache = make(map[string]string)
// }

// func (h *HostnameDiscovery) Stop() {
// 	log.Printf("[DISCOVERY] Stopping on %s...", h.iface.Interface.Name)
// 	h.cancel()
// }

package vlan

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// ============================================
// HOSTNAME DISCOVERY
// ============================================

type HostnameDiscovery struct {
	iface       *InterfaceInfo
	ssdpScanner *SSDPScanner
	mdnsScanner *MDNSScanner
	mu          sync.RWMutex
	cache       map[string]string
	stats       DiscoveryStats
	ctx         context.Context
	cancel      context.CancelFunc
}

type DiscoveryStats struct {
	TotalScans       int
	SSDPDiscoveries  int
	MDNSDiscoveries  int
	LastScanTime     time.Time
	LastScanDuration time.Duration
}

func NewHostnameDiscovery(interfaceName string) (*HostnameDiscovery, error) {
	iface, err := GetInterfaceByName(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface %s: %w", interfaceName, err)
	}

	ssdpScanner, err := NewSSDP(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSDP scanner: %w", err)
	}

	mdnsScanner, err := NewMDNS(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to create mDNS scanner: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	hd := &HostnameDiscovery{
		iface:       iface,
		ssdpScanner: ssdpScanner,
		mdnsScanner: mdnsScanner,
		cache:       make(map[string]string),
		ctx:         ctx,
		cancel:      cancel,
	}
	log.Printf("[DISCOVERY] Initialized on %s (%s)", interfaceName, iface.IPv4.String())
	return hd, nil
}

// Start runs continuous mDNS and SSDP discovery.
// Discovered hostnames are stored in the cache and made available via GetHostname.
// Also saves directly to the ARP scanner's host map via the saveFunc callback —
// this is the fix for "mDNS discoveries not saved to database": previously the
// hostname was only cached in memory; the DB update only happened when the ARP
// scanner's next periodic sync fired (up to 30s later). Now every discovery
// immediately calls saveFunc so the DB is updated in real time.
func (h *HostnameDiscovery) Start() {
	log.Printf("[DISCOVERY] Started continuous discovery on %s", h.iface.Interface.Name)

	results := make(chan *Device, 256)

	go func() {
		if err := h.ssdpScanner.Scan(h.ctx, results); err != nil &&
			err != context.Canceled {
			log.Printf("[SSDP] Scanner exited: %v", err)
		}
	}()

	go func() {
		if err := h.mdnsScanner.Scan(h.ctx, results); err != nil &&
			err != context.Canceled {
			log.Printf("[mDNS] Scanner exited: %v", err)
		}
	}()

	for {
		select {
		case <-h.ctx.Done():
			log.Printf("[DISCOVERY] Stopped on %s", h.iface.Interface.Name)
			return

		case device, ok := <-results:
			if !ok {
				return
			}
			if device.IP == nil || device.Name == "" {
				continue
			}

			// FIX: validate the name is clean before caching.
			// The scanner already calls cleanMDNSHostname, but if for any
			// reason a raw JSON string slips through, reject it here.
			if !isValidDiscoveryName(device.Name) {
				log.Printf("[%s] Skipping invalid name for %s: %q",
					device.Proto, device.IP, device.Name)
				continue
			}

			ipStr := device.IP.String()

			h.mu.Lock()
			isNew := h.cache[ipStr] != device.Name
			if isNew {
				log.Printf("[%s] Discovered: %s → %s", device.Proto, ipStr, device.Name)
			}
			h.cache[ipStr] = device.Name

			switch device.Proto {
			case "ssdp":
				h.stats.SSDPDiscoveries++
			case "mdns":
				h.stats.MDNSDiscoveries++
			}
			h.stats.LastScanTime = time.Now()
			h.mu.Unlock()

			// Notify any registered update callbacks (e.g. to update the DB).
			h.notifyHostnameUpdate(ipStr, device.Name, device.Proto)
		}
	}
}

// hostnameUpdateCallbacks is called when a hostname is discovered or updated.
// Register with OnHostnameDiscovered to get real-time DB updates.
var _ = (*HostnameDiscovery)(nil) // compile check

type HostnameUpdateFunc func(ip, hostname, proto string)

// onHostnameUpdate is an optional callback registered by the scanner manager
// to persist hostname updates to the DB immediately rather than waiting for
// the next periodic sync.
type hostnameUpdateCallback struct {
	mu  sync.RWMutex
	fns []HostnameUpdateFunc
}

var globalHostnameCallbacks = &hostnameUpdateCallback{}

// RegisterHostnameCallback registers a function to be called whenever a
// hostname is discovered via mDNS or SSDP. The VLANScanManager uses this
// to write hostname updates to the DB immediately.
func RegisterHostnameCallback(fn HostnameUpdateFunc) {
	globalHostnameCallbacks.mu.Lock()
	defer globalHostnameCallbacks.mu.Unlock()
	globalHostnameCallbacks.fns = append(globalHostnameCallbacks.fns, fn)
}

func (h *HostnameDiscovery) notifyHostnameUpdate(ip, hostname, proto string) {
	globalHostnameCallbacks.mu.RLock()
	fns := make([]HostnameUpdateFunc, len(globalHostnameCallbacks.fns))
	copy(fns, globalHostnameCallbacks.fns)
	globalHostnameCallbacks.mu.RUnlock()

	for _, fn := range fns {
		go fn(ip, hostname, proto)
	}
}

// isValidDiscoveryName rejects raw JSON strings and other garbage that
// occasionally slip through the mDNS hostname cleaner.
func isValidDiscoveryName(name string) bool {
	if name == "" || len(name) > 128 {
		return false
	}
	// Reject anything that looks like raw JSON
	if (strings.HasPrefix(name, "{") && strings.Contains(name, ":")) ||
		strings.HasPrefix(name, `{\"`) {
		return false
	}
	// Reject strings with too many special characters
	special := 0
	for _, r := range name {
		if r == '{' || r == '}' || r == '[' || r == ']' || r == '\\' || r == '"' {
			special++
		}
	}
	return special <= 2
}

// ============================================
// ACCESSORS
// ============================================

func (h *HostnameDiscovery) GetHostname(ip string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.cache[ip]
}

func (h *HostnameDiscovery) GetStats() DiscoveryStats {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.stats
}

func (h *HostnameDiscovery) GetCacheSize() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.cache)
}

func (h *HostnameDiscovery) GetAllCached() map[string]string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make(map[string]string, len(h.cache))
	for k, v := range h.cache {
		out[k] = v
	}
	return out
}

func (h *HostnameDiscovery) ClearCache() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cache = make(map[string]string)
}

func (h *HostnameDiscovery) Stop() {
	log.Printf("[DISCOVERY] Stopping on %s...", h.iface.Interface.Name)
	h.cancel()
}
