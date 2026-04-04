package vlan

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ============================================
// HOSTNAME DISCOVERY
// Runs mDNS and SSDP discovery on a single network interface.
// Shared across all VLANs that use the same physical interface.
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

// ScanSubnet runs SSDP and mDNS concurrently and populates the hostname cache.
func (h *HostnameDiscovery) ScanSubnet() error {
	startTime := time.Now()
	results := make(chan *Device, 256)
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(h.ctx, 10*time.Second)
	defer cancel()

	ssdpCount, mdnsCount := 0, 0

	// ✅ ADD: Verbose logging
	log.Printf("[DISCOVERY] ========== Starting scan on %s ==========", h.iface.Interface.Name)
	log.Printf("[DISCOVERY] Interface IP: %s", h.iface.IPv4.String())
	log.Printf("[DISCOVERY] Timeout: 10 seconds")

	wg.Add(2)

	// SSDP scanner
	go func() {
		defer wg.Done()
		log.Printf("[SSDP] Starting scanner...")
		if err := h.ssdpScanner.Scan(ctx, results); err != nil &&
			err != context.DeadlineExceeded && err != context.Canceled {
			log.Printf("[SSDP] ❌ Scan error: %v", err)
		} else {
			log.Printf("[SSDP] ✅ Scan completed")
		}
	}()

	// mDNS scanner
	go func() {
		defer wg.Done()
		log.Printf("[mDNS] Starting scanner...")
		if err := h.mdnsScanner.Scan(ctx, results); err != nil &&
			err != context.DeadlineExceeded && err != context.Canceled {
			log.Printf("[mDNS] ❌ Scan error: %v", err)
		} else {
			log.Printf("[mDNS] ✅ Scan completed")
		}
	}()

	go func() {
		wg.Wait()
		close(results)
		log.Printf("[DISCOVERY] Both scanners finished, processing results...")
	}()

	discovered := 0
	for device := range results {
		if device.IP == nil || device.Name == "" {
			log.Printf("[%s] ⚠️  Skipping device: IP=%v, Name=%s",
				device.Proto, device.IP, device.Name)
			continue
		}

		ipStr := device.IP.String()

		h.mu.Lock()
		if _, exists := h.cache[ipStr]; !exists {
			log.Printf("[%s] ✅ NEW: %s → %s", device.Proto, ipStr, device.Name)
		} else {
			log.Printf("[%s] 🔄 UPDATE: %s → %s", device.Proto, ipStr, device.Name)
		}
		h.cache[ipStr] = device.Name
		h.mu.Unlock()

		discovered++
		switch device.Proto {
		case "ssdp":
			ssdpCount++
		case "mdns":
			mdnsCount++
		}
	}

	duration := time.Since(startTime)

	h.mu.Lock()
	h.stats.TotalScans++
	h.stats.SSDPDiscoveries += ssdpCount
	h.stats.MDNSDiscoveries += mdnsCount
	h.stats.LastScanTime = time.Now()
	h.stats.LastScanDuration = duration
	h.mu.Unlock()

	// ✅ ALWAYS log, even if 0
	log.Printf("[DISCOVERY] ========== Scan complete on %s ==========", h.iface.Interface.Name)
	log.Printf("[DISCOVERY] Found: %d devices (SSDP: %d, mDNS: %d)",
		discovered, ssdpCount, mdnsCount)
	log.Printf("[DISCOVERY] Duration: %v", duration.Round(time.Millisecond))
	log.Printf("[DISCOVERY] Total cache size: %d", len(h.cache))
	log.Printf("[DISCOVERY] ================================================")

	return nil
}

// GetHostname returns the cached hostname for the given IP, or "".
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

// Start runs ScanSubnet periodically until Stop is called.
func (h *HostnameDiscovery) Start(interval time.Duration) {
	log.Printf("[DISCOVERY] Started (interval: %v)", interval)

	if err := h.ScanSubnet(); err != nil {
		log.Printf("[DISCOVERY] Initial scan failed: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	scanCount := 0
	for {
		select {
		case <-h.ctx.Done():
			log.Printf("[DISCOVERY] Stopped (scans: %d, cache: %d)",
				scanCount, h.GetCacheSize())
			return

		case <-ticker.C:
			scanCount++
			if err := h.ScanSubnet(); err != nil {
				log.Printf("[DISCOVERY] Scan #%d failed: %v", scanCount, err)
			}
			if scanCount%10 == 0 {
				s := h.GetStats()
				log.Printf("[DISCOVERY] Stats — scans: %d, SSDP: %d, mDNS: %d, cache: %d",
					s.TotalScans, s.SSDPDiscoveries, s.MDNSDiscoveries, h.GetCacheSize())
			}
		}
	}
}

func (h *HostnameDiscovery) Stop() {
	log.Printf("[DISCOVERY] Stopping on %s...", h.iface.Interface.Name)
	h.cancel()
}
