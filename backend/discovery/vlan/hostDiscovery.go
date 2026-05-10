package vlan

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

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
			h.notifyHostnameUpdate(ipStr, device.Name, device.Proto)
		}
	}
}

var _ = (*HostnameDiscovery)(nil)

type HostnameUpdateFunc func(ip, hostname, proto string)
type hostnameUpdateCallback struct {
	mu  sync.RWMutex
	fns []HostnameUpdateFunc
}

var globalHostnameCallbacks = &hostnameUpdateCallback{}

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

func isValidDiscoveryName(name string) bool {
	if name == "" || len(name) > 128 {
		return false
	}
	if (strings.HasPrefix(name, "{") && strings.Contains(name, ":")) ||
		strings.HasPrefix(name, `{\"`) {
		return false
	}
	special := 0
	for _, r := range name {
		if r == '{' || r == '}' || r == '[' || r == ']' || r == '\\' || r == '"' {
			special++
		}
	}
	return special <= 2
}

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
