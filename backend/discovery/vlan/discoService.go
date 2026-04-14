// package vlan

// import (
// 	"bufio"
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"log"
// 	"log/slog"
// 	"net"
// 	"net/textproto"
// 	"net/url"
// 	"strings"
// 	"time"

// 	hashimdns "github.com/hashicorp/mdns"
// )

// //
// // ====================== CORE TYPES ======================
// //

// type Device struct {
// 	IP    net.IP
// 	Name  string
// 	Meta  map[string]string
// 	Proto string
// }

// func NewDevice(ip net.IP) *Device {
// 	return &Device{
// 		IP:   ip,
// 		Meta: make(map[string]string),
// 	}
// }

// func (d *Device) AddMeta(k, v string) {
// 	if d.Meta == nil {
// 		d.Meta = make(map[string]string)
// 	}
// 	d.Meta[k] = v
// }

// type Scanner interface {
// 	Name() string
// 	Scan(ctx context.Context, out chan<- *Device) error
// }

// type InterfaceInfo struct {
// 	Interface *net.Interface
// 	IPv4      net.IP
// }

// // ->Helper to get interface by name
// func GetInterfaceByName(name string) (*InterfaceInfo, error) {
// 	iface, err := net.InterfaceByName(name)
// 	if err != nil {
// 		return nil, fmt.Errorf("interface not found: %w", err)
// 	}

// 	addrs, err := iface.Addrs()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get addresses: %w", err)
// 	}

// 	for _, addr := range addrs {
// 		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
// 			return &InterfaceInfo{
// 				Interface: iface,
// 				IPv4:      ipnet.IP.To4(),
// 			}, nil
// 		}
// 	}

// 	return nil, fmt.Errorf("no IPv4 address found on %s", name)
// }

// type Logger interface {
// 	Log(ctx context.Context, level slog.Level, msg string, args ...any)
// }

// type NoOpLogger struct{}

// func (NoOpLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {}

// type Option func(any) error

// func WithLogger(logger Logger) Option {
// 	return func(s any) error {
// 		if logger == nil {
// 			return errors.New("logger nil")
// 		}
// 		switch v := s.(type) {
// 		case *MDNSScanner:
// 			v.logger = logger
// 		case *SSDPScanner:
// 			v.logger = logger
// 		}
// 		return nil
// 	}
// }

// //
// // ====================== MDNS ======================
// //

// const mdnsQuery = "_services._dns-sd._udp"

// type MDNSScanner struct {
// 	iface     *InterfaceInfo
// 	logger    Logger
// 	queryFunc func(*hashimdns.QueryParam) error
// }

// func NewMDNS(iface *InterfaceInfo, opts ...Option) (*MDNSScanner, error) {
// 	if iface == nil {
// 		return nil, errors.New("interface info is nil")
// 	}
// 	s := &MDNSScanner{
// 		iface:     iface,
// 		logger:    NoOpLogger{},
// 		queryFunc: hashimdns.Query,
// 	}
// 	for _, o := range opts {
// 		if err := o(s); err != nil {
// 			return nil, err
// 		}
// 	}
// 	return s, nil
// }

// func (s *MDNSScanner) Name() string { return "mdns" }

// func extractCleanHostname(rawName string) string {
// 	// Handle mDNS service discovery format
// 	// Example: "{\"nm\":\"Kaimunchi\",\"as\":\"[8194]\",\"ip\":\"0\"}._mi-connect._udp.local."

// 	// Try to parse JSON metadata
// 	if strings.HasPrefix(rawName, "{") {
// 		endJSON := strings.Index(rawName, "}")
// 		if endJSON > 0 {
// 			jsonPart := rawName[:endJSON+1]

// 			// Parse the JSON to extract the name
// 			var metadata map[string]interface{}
// 			if err := json.Unmarshal([]byte(jsonPart), &metadata); err == nil {
// 				if nm, ok := metadata["nm"].(string); ok && nm != "" {
// 					return nm
// 				}
// 			}
// 		}
// 	}

// 	// Remove service type suffixes
// 	// Example: "MyDevice._device-info._tcp.local." → "MyDevice"
// 	cleanName := rawName

// 	// Common mDNS service patterns to remove
// 	serviceSuffixes := []string{
// 		"._device-info._tcp.local.",
// 		"._mi-connect._udp.local.",
// 		"._airplay._tcp.local.",
// 		"._raop._tcp.local.",
// 		"._homekit._tcp.local.",
// 		"._hap._tcp.local.",
// 		"._companion-link._tcp.local.",
// 		"._sleep-proxy._udp.local.",
// 		"._services._dns-sd._udp.local.",
// 		".local.",
// 		"._tcp.local.",
// 		"._udp.local.",
// 	}

// 	for _, suffix := range serviceSuffixes {
// 		if idx := strings.Index(cleanName, suffix); idx > 0 {
// 			cleanName = cleanName[:idx]
// 			break
// 		}
// 	}

// 	// Remove any remaining ._* patterns
// 	if idx := strings.Index(cleanName, "._"); idx > 0 {
// 		cleanName = cleanName[:idx]
// 	}

// 	// Trim whitespace and dots
// 	cleanName = strings.Trim(cleanName, ". \t\r\n")

// 	// If empty after cleaning, return empty string
// 	if cleanName == "" || cleanName == "{}" {
// 		return ""
// 	}

// 	return cleanName
// }

// func (s *MDNSScanner) Scan(ctx context.Context, out chan<- *Device) error {
// 	entries := make(chan *hashimdns.ServiceEntry, 256)
// 	errCh := make(chan error, 1)

// 	go func() {
// 		p := hashimdns.DefaultParams(mdnsQuery)
// 		p.Entries = entries
// 		p.Interface = s.iface.Interface
// 		p.DisableIPv6 = true

// 		p.Logger = log.Default()
// 		p.Logger.SetOutput(io.Discard)

// 		if dl, ok := ctx.Deadline(); ok {
// 			p.Timeout = time.Until(dl)
// 		}

// 		if err := s.queryFunc(p); err != nil {
// 			errCh <- err
// 		}
// 		close(errCh)
// 	}()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()

// 		case e, ok := <-entries:
// 			if !ok {
// 				select {
// 				case err := <-errCh:
// 					return err
// 				default:
// 					return nil
// 				}
// 			}

// 			if e.AddrV4 == nil {
// 				continue
// 			}

// 			d := NewDevice(e.AddrV4)

// 			// ->Clean the hostname before using it
// 			cleanName := extractCleanHostname(e.Name)
// 			if cleanName != "" {
// 				d.Name = cleanName
// 			} else {
// 				// Skip devices with no meaningful name
// 				continue
// 			}

// 			d.Proto = "mdns"

// 			// Parse info fields
// 			for _, f := range e.InfoFields {
// 				if kv := splitKV(f); kv != nil {
// 					d.AddMeta(kv[0], kv[1])
// 				} else {
// 					d.AddMeta(f, "true")
// 				}
// 			}

// 			s.logger.Log(ctx, slog.LevelDebug, "mdns device",
// 				"name", d.Name, "ip", d.IP.String())

// 			select {
// 			case out <- d:
// 			case <-ctx.Done():
// 				return ctx.Err()
// 			}

// 		case err := <-errCh:
// 			return err
// 		}
// 	}
// }

// //
// // ====================== SSDP ======================
// //

// const ssdpAddr = "239.255.255.250:1900"

// type SSDPScanner struct {
// 	iface  *InterfaceInfo
// 	logger Logger
// }

// func NewSSDP(iface *InterfaceInfo, opts ...Option) (*SSDPScanner, error) {
// 	if iface == nil {
// 		return nil, errors.New("interface info is nil")
// 	}
// 	s := &SSDPScanner{
// 		iface:  iface,
// 		logger: NoOpLogger{},
// 	}
// 	for _, o := range opts {
// 		if err := o(s); err != nil {
// 			return nil, err
// 		}
// 	}
// 	return s, nil
// }

// func (s *SSDPScanner) Name() string { return "ssdp" }

// func (s *SSDPScanner) Scan(ctx context.Context, out chan<- *Device) error {
// 	maddr, err := net.ResolveUDPAddr("udp4", ssdpAddr)
// 	if err != nil {
// 		return fmt.Errorf("resolve ssdp addr: %w", err)
// 	}

// 	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
// 		IP:   s.iface.IPv4,
// 		Port: 0,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("listen udp: %w", err)
// 	}
// 	defer conn.Close()

// 	s.logger.Log(ctx, slog.LevelDebug, "sending ssdp m-search",
// 		"from", conn.LocalAddr().String())

// 	if err := sendSSDP(conn, maddr); err != nil {
// 		return err
// 	}

// 	// Set deadline from context
// 	if dl, ok := ctx.Deadline(); ok {
// 		conn.SetReadDeadline(dl)
// 	} else {
// 		// Default timeout if no deadline
// 		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
// 	}

// 	buf := make([]byte, 8192)

// 	for {
// 		if ctx.Err() != nil {
// 			return ctx.Err()
// 		}

// 		n, src, err := conn.ReadFromUDP(buf)
// 		if err != nil {
// 			if ne, ok := err.(net.Error); ok && ne.Timeout() {
// 				return nil
// 			}
// 			return fmt.Errorf("read ssdp: %w", err)
// 		}
// 		handleSSDP(out, src, buf[:n])
// 	}
// }

// //
// // ====================== SSDP HELPERS ======================
// //

// func sendSSDP(conn *net.UDPConn, addr *net.UDPAddr) error {
// 	req := fmt.Sprintf(
// 		"M-SEARCH * HTTP/1.1\r\n"+
// 			"HOST: %s\r\n"+
// 			"MAN: \"ssdp:discover\"\r\n"+
// 			"MX: 2\r\n"+
// 			"ST: ssdp:all\r\n"+
// 			"USER-AGENT: vlan-scanner/1.0\r\n\r\n",
// 		ssdpAddr,
// 	)
// 	_, err := conn.WriteToUDP([]byte(req), addr)
// 	if err != nil {
// 		return fmt.Errorf("send m-search: %w", err)
// 	}
// 	return nil
// }

// // func handleSSDP(out chan<- *Device, src *net.UDPAddr, payload []byte) {
// // 	loc, server := parseHeaders(payload)

// // 	ip := src.IP
// // 	if ip == nil && loc != "" {
// // 		ip = ipFromLocation(loc)
// // 	}
// // 	if ip == nil {
// // 		return
// // 	}

// // 	d := NewDevice(ip)
// // 	// ->Fixed: Use server as name if available
// // 	if server != "" {
// // 		d.Name = server
// // 	}
// // 	d.Proto = "ssdp"

// // 	if loc != "" {
// // 		d.AddMeta("location", loc)
// // 	}
// // 	if server != "" {
// // 		d.AddMeta("server", server)
// // 	}

// //		select {
// //		case out <- d:
// //		default:
// //		}
// //	}
// func handleSSDP(out chan<- *Device, src *net.UDPAddr, payload []byte) {
// 	loc, server := parseHeaders(payload)

// 	ip := src.IP
// 	if ip == nil && loc != "" {
// 		ip = ipFromLocation(loc)
// 	}
// 	if ip == nil {
// 		return
// 	}

// 	d := NewDevice(ip)

// 	// ->Clean server name
// 	if server != "" {
// 		// Remove common prefixes/suffixes
// 		cleanServer := strings.TrimSpace(server)

// 		// Remove HTTP server version info
// 		// Example: "UPnP/1.0 Smart TV" → "Smart TV"
// 		if idx := strings.Index(cleanServer, " "); idx > 0 {
// 			parts := strings.Split(cleanServer, " ")
// 			// If first part looks like a version (contains /), skip it
// 			if strings.Contains(parts[0], "/") && len(parts) > 1 {
// 				cleanServer = strings.Join(parts[1:], " ")
// 			}
// 		}

// 		d.Name = cleanServer
// 	} else {
// 		// Skip devices with no name
// 		return
// 	}

// 	d.Proto = "ssdp"

// 	if loc != "" {
// 		d.AddMeta("location", loc)
// 	}
// 	if server != "" {
// 		d.AddMeta("server", server)
// 	}

// 	select {
// 	case out <- d:
// 	default:
// 	}
// }

// func parseHeaders(b []byte) (string, string) {
// 	// Ensure buffer ends with CRLFCRLF
// 	if !bytes.HasSuffix(b, []byte("\r\n\r\n")) {
// 		b = append(append([]byte{}, b...), []byte("\r\n\r\n")...)
// 	}

// 	tr := textproto.NewReader(bufio.NewReader(bytes.NewReader(b)))
// 	// Skip status line
// 	_, _ = tr.ReadLine()

// 	h, err := tr.ReadMIMEHeader()
// 	if err != nil {
// 		return "", ""
// 	}

// 	return strings.TrimSpace(h.Get("Location")), strings.TrimSpace(h.Get("Server"))
// }

// func ipFromLocation(loc string) net.IP {
// 	u, err := url.Parse(loc)
// 	if err != nil {
// 		return nil
// 	}
// 	host := u.Host
// 	if h, _, err := net.SplitHostPort(host); err == nil {
// 		host = h
// 	}
// 	return net.ParseIP(host)
// }

// //
// // ====================== UTILS ======================
// //

// func splitKV(s string) []string {
// 	parts := strings.SplitN(s, "=", 2)
// 	if len(parts) == 2 {
// 		return parts
// 	}
// 	return nil
// }

// package vlan

// import (
// 	"bufio"
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net"
// 	"net/textproto"
// 	"strings"
// 	"time"

// 	hashimdns "github.com/hashicorp/mdns"
// 	"golang.org/x/net/ipv4"
// )

// // ============================================
// // CORE TYPES
// // ============================================

// type Device struct {
// 	IP    net.IP
// 	Name  string
// 	Meta  map[string]string
// 	Proto string
// }

// func NewDevice(ip net.IP) *Device {
// 	return &Device{
// 		IP:   ip,
// 		Meta: make(map[string]string),
// 	}
// }

// func (d *Device) AddMeta(k, v string) {
// 	if d.Meta == nil {
// 		d.Meta = make(map[string]string)
// 	}
// 	d.Meta[k] = v
// }

// type InterfaceInfo struct {
// 	Interface *net.Interface
// 	IPv4      net.IP
// }

// // ->Get interface by name
// func GetInterfaceByName(name string) (*InterfaceInfo, error) {
// 	iface, err := net.InterfaceByName(name)
// 	if err != nil {
// 		return nil, fmt.Errorf("interface not found: %w", err)
// 	}

// 	addrs, err := iface.Addrs()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get addresses: %w", err)
// 	}

// 	for _, addr := range addrs {
// 		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
// 			return &InterfaceInfo{
// 				Interface: iface,
// 				IPv4:      ipnet.IP.To4(),
// 			}, nil
// 		}
// 	}

// 	return nil, fmt.Errorf("no IPv4 address found on %s", name)
// }

// // ============================================
// // MDNS SCANNER
// // ============================================

// const mdnsQuery = "_services._dns-sd._udp"

// type MDNSScanner struct {
// 	iface *InterfaceInfo
// }

// func NewMDNS(iface *InterfaceInfo) (*MDNSScanner, error) {
// 	if iface == nil {
// 		return nil, fmt.Errorf("interface info is nil")
// 	}
// 	return &MDNSScanner{iface: iface}, nil
// }

// func (s *MDNSScanner) Name() string { return "mdns" }

// // ->Continuous listener that runs forever
// func (s *MDNSScanner) Scan(ctx context.Context, out chan<- *Device) error {
// 	const window = 10 * time.Second

// 	for {
// 		if ctx.Err() != nil {
// 			return ctx.Err()
// 		}

// 		entriesCh := make(chan *hashimdns.ServiceEntry, 256)

// 		params := hashimdns.DefaultParams(mdnsQuery)
// 		params.Entries = entriesCh
// 		params.Interface = s.iface.Interface
// 		params.DisableIPv6 = true
// 		params.Timeout = window
// 		params.Logger = log.New(io.Discard, "", 0)

// 		// Run query in background
// 		go func() {
// 			hashimdns.Query(params)
// 			close(entriesCh)
// 		}()

// 		// Drain results
// 	drain:
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return ctx.Err()
// 			case entry, ok := <-entriesCh:
// 				if !ok {
// 					break drain
// 				}
// 				if entry.AddrV4 == nil {
// 					continue
// 				}

// 				d := NewDevice(entry.AddrV4)
// 				d.Proto = "mdns"

// 				// ->Clean the hostname
// 				cleanName := cleanMDNSHostname(entry.Name)
// 				if cleanName == "" {
// 					continue // Skip if no meaningful name
// 				}
// 				d.Name = cleanName

// 				// Add metadata
// 				d.AddMeta("host", entry.Host)
// 				d.AddMeta("port", fmt.Sprintf("%d", entry.Port))
// 				for _, f := range entry.InfoFields {
// 					if kv := splitKV(f); kv != nil {
// 						d.AddMeta(kv[0], kv[1])
// 					} else {
// 						d.AddMeta(f, "true")
// 					}
// 				}

// 				select {
// 				case out <- d:
// 				case <-ctx.Done():
// 					return ctx.Err()
// 				}
// 			}
// 		}
// 		// Immediately restart next window
// 	}
// }

// // ->Clean mDNS hostname
// func cleanMDNSHostname(rawName string) string {
// 	// Handle JSON metadata format
// 	// Example: {"nm":"Kaimunchi","as":"[8194]","ip":"4"}._mi-connect._udp.local.
// 	if strings.HasPrefix(rawName, "{") {
// 		endJSON := strings.Index(rawName, "}")
// 		if endJSON > 0 {
// 			jsonPart := rawName[:endJSON+1]
// 			var metadata map[string]interface{}
// 			if err := json.Unmarshal([]byte(jsonPart), &metadata); err == nil {
// 				if nm, ok := metadata["nm"].(string); ok && nm != "" {
// 					return nm
// 				}
// 			}
// 		}
// 	}

// 	// Remove service type suffixes
// 	cleanName := rawName
// 	serviceSuffixes := []string{
// 		"._device-info._tcp.local.",
// 		"._mi-connect._udp.local.",
// 		"._airplay._tcp.local.",
// 		"._raop._tcp.local.",
// 		"._homekit._tcp.local.",
// 		"._hap._tcp.local.",
// 		"._companion-link._tcp.local.",
// 		"._sleep-proxy._udp.local.",
// 		"._services._dns-sd._udp.local.",
// 		"._myapp._tcp.local.",
// 		".local.",
// 		"._tcp.local.",
// 		"._udp.local.",
// 	}

// 	for _, suffix := range serviceSuffixes {
// 		if idx := strings.Index(cleanName, suffix); idx > 0 {
// 			cleanName = cleanName[:idx]
// 			break
// 		}
// 	}

// 	// Remove any remaining ._* patterns
// 	if idx := strings.Index(cleanName, "._"); idx > 0 {
// 		cleanName = cleanName[:idx]
// 	}

// 	// Trim whitespace and dots
// 	cleanName = strings.Trim(cleanName, ". \t\r\n")

// 	// If empty after cleaning, return empty
// 	if cleanName == "" || cleanName == "{}" {
// 		return ""
// 	}

// 	return cleanName
// }

// // ============================================
// // SSDP SCANNER
// // ============================================

// const (
// 	ssdpMulticast = "239.255.255.250"
// 	ssdpPort      = 1900
// )

// type SSDPScanner struct {
// 	iface *InterfaceInfo
// }

// func NewSSDP(iface *InterfaceInfo) (*SSDPScanner, error) {
// 	if iface == nil {
// 		return nil, fmt.Errorf("interface info is nil")
// 	}
// 	return &SSDPScanner{iface: iface}, nil
// }

// func (s *SSDPScanner) Name() string { return "ssdp" }

// // ->Listen for SSDP broadcasts on the correct interface
// func (s *SSDPScanner) Scan(ctx context.Context, out chan<- *Device) error {
// 	// ->Bind to 0.0.0.0:1900 to receive multicast
// 	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
// 		IP:   net.IPv4zero,
// 		Port: ssdpPort,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("cannot bind :1900 (try sudo): %w", err)
// 	}
// 	defer conn.Close()

// 	// ->Join multicast group on specific interface
// 	pc := ipv4.NewPacketConn(conn)
// 	if err := pc.JoinGroup(s.iface.Interface, &net.UDPAddr{
// 		IP: net.ParseIP(ssdpMulticast),
// 	}); err != nil {
// 		return fmt.Errorf("join multicast group: %w", err)
// 	}

// 	log.Printf("[SSDP] Joined multicast %s:%d on %s", ssdpMulticast, ssdpPort, s.iface.Interface.Name)

// 	// ->Send initial M-SEARCH
// 	if err := s.sendMSearch(conn); err != nil {
// 		log.Printf("[SSDP] M-SEARCH error: %v", err)
// 	}

// 	// ->Re-send M-SEARCH every 60 seconds
// 	go func() {
// 		ticker := time.NewTicker(60 * time.Second)
// 		defer ticker.Stop()
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return
// 			case <-ticker.C:
// 				if err := s.sendMSearch(conn); err != nil {
// 					log.Printf("[SSDP] M-SEARCH error: %v", err)
// 				}
// 			}
// 		}
// 	}()

// 	// ->Main read loop
// 	buf := make([]byte, 8192)
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		default:
// 		}

// 		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
// 		n, src, err := conn.ReadFromUDP(buf)
// 		if err != nil {
// 			if ne, ok := err.(net.Error); ok && ne.Timeout() {
// 				continue
// 			}
// 			log.Printf("[SSDP] Read error: %v", err)
// 			continue
// 		}

// 		d := s.parsePacket(src, buf[:n])
// 		if d == nil {
// 			continue
// 		}

// 		// Skip byebye messages
// 		if d.Meta["nts"] == "ssdp:byebye" {
// 			continue
// 		}

// 		select {
// 		case out <- d:
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		}
// 	}
// }

// // ->Send M-SEARCH to trigger device responses
// func (s *SSDPScanner) sendMSearch(conn *net.UDPConn) error {
// 	addr := &net.UDPAddr{
// 		IP:   net.ParseIP(ssdpMulticast),
// 		Port: ssdpPort,
// 	}

// 	msg := fmt.Sprintf(
// 		"M-SEARCH * HTTP/1.1\r\n"+
// 			"HOST: %s:%d\r\n"+
// 			"MAN: \"ssdp:discover\"\r\n"+
// 			"MX: 3\r\n"+
// 			"ST: ssdp:all\r\n"+
// 			"USER-AGENT: vlan-scanner/1.0\r\n"+
// 			"\r\n",
// 		ssdpMulticast, ssdpPort,
// 	)

// 	_, err := conn.WriteToUDP([]byte(msg), addr)
// 	return err
// }

// // ->Parse SSDP packet
// func (s *SSDPScanner) parsePacket(src *net.UDPAddr, buf []byte) *Device {
// 	data := buf
// 	if !bytes.HasSuffix(data, []byte("\r\n\r\n")) {
// 		data = append(append([]byte{}, data...), "\r\n\r\n"...)
// 	}

// 	br := bufio.NewReader(bytes.NewReader(data))
// 	tr := textproto.NewReader(br)

// 	firstLine, err := tr.ReadLine()
// 	if err != nil {
// 		return nil
// 	}

// 	hdr, _ := tr.ReadMIMEHeader()

// 	// Only care about NOTIFY and 200 OK responses
// 	isNotify := strings.HasPrefix(firstLine, "NOTIFY")
// 	isOK := strings.HasPrefix(firstLine, "HTTP/1.1 200")
// 	if !isNotify && !isOK {
// 		return nil
// 	}

// 	d := NewDevice(src.IP)
// 	d.Proto = "ssdp"

// 	// Extract server name and clean it
// 	server := strings.TrimSpace(hdr.Get("Server"))
// 	if server != "" {
// 		d.Name = cleanSSDPServer(server)
// 	}

// 	// Add metadata
// 	d.AddMeta("usn", strings.TrimSpace(hdr.Get("Usn")))
// 	d.AddMeta("nt", strings.TrimSpace(hdr.Get("Nt")))
// 	d.AddMeta("nts", strings.TrimSpace(hdr.Get("Nts")))
// 	d.AddMeta("location", strings.TrimSpace(hdr.Get("Location")))
// 	d.AddMeta("server", server)

// 	if d.Name == "" {
// 		// Try to extract from location or other fields
// 		location := hdr.Get("Location")
// 		if location != "" {
// 			d.Name = extractDeviceNameFromURL(location)
// 		}
// 	}

// 	// Skip if no meaningful name
// 	if d.Name == "" {
// 		return nil
// 	}

// 	return d
// }

// // ->Clean SSDP server string
// func cleanSSDPServer(server string) string {
// 	cleanServer := strings.TrimSpace(server)

// 	// Remove UPnP version prefix
// 	// Example: "UPnP/1.0 Smart TV" → "Smart TV"
// 	if idx := strings.Index(cleanServer, " "); idx > 0 {
// 		parts := strings.Split(cleanServer, " ")
// 		// If first part looks like a version (contains /), skip it
// 		if strings.Contains(parts[0], "/") && len(parts) > 1 {
// 			cleanServer = strings.Join(parts[1:], " ")
// 		}
// 	}

// 	// Remove common suffixes
// 	cleanServer = strings.TrimSuffix(cleanServer, " UPnP/1.0")
// 	cleanServer = strings.TrimSuffix(cleanServer, " DLNADOC/1.50")

// 	return cleanServer
// }

// // ->Extract device name from location URL
// func extractDeviceNameFromURL(location string) string {
// 	// Try to extract hostname from URL
// 	// http://192.168.1.100:8080/description.xml → return empty (just IP)
// 	// http://smart-tv.local/device.xml → return "smart-tv"

// 	if !strings.HasPrefix(location, "http") {
// 		return ""
// 	}

// 	// Simple extraction - could be enhanced
// 	parts := strings.Split(location, "/")
// 	if len(parts) > 2 {
// 		hostPort := parts[2]
// 		host := strings.Split(hostPort, ":")[0]

// 		// Skip if it's just an IP address
// 		if net.ParseIP(host) != nil {
// 			return ""
// 		}

// 		// Remove .local suffix
// 		host = strings.TrimSuffix(host, ".local")
// 		return host
// 	}

// 	return ""
// }

// // ============================================
// // UTILS
// // ============================================

//	func splitKV(s string) []string {
//		parts := strings.SplitN(s, "=", 2)
//		if len(parts) == 2 {
//			return parts
//		}
//		return nil
//	}
package vlan

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"strings"
	"time"
	"unicode"

	hashimdns "github.com/hashicorp/mdns"
	"golang.org/x/net/ipv4"
)

// ============================================
// CORE TYPES
// ============================================

type Device struct {
	IP    net.IP
	Name  string
	Meta  map[string]string
	Proto string
}

func NewDevice(ip net.IP) *Device {
	return &Device{IP: ip, Meta: make(map[string]string)}
}

func (d *Device) AddMeta(k, v string) {
	if d.Meta == nil {
		d.Meta = make(map[string]string)
	}
	d.Meta[k] = v
}

type InterfaceInfo struct {
	Interface *net.Interface
	IPv4      net.IP
}

func GetInterfaceByName(name string) (*InterfaceInfo, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("interface not found: %w", err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			return &InterfaceInfo{
				Interface: iface,
				IPv4:      ipnet.IP.To4(),
			}, nil
		}
	}
	return nil, fmt.Errorf("no IPv4 address found on %s", name)
}

// ============================================
// MDNS SCANNER
// ============================================

const mdnsQuery = "_services._dns-sd._udp"

type MDNSScanner struct {
	iface *InterfaceInfo
}

func NewMDNS(iface *InterfaceInfo) (*MDNSScanner, error) {
	if iface == nil {
		return nil, fmt.Errorf("interface info is nil")
	}
	return &MDNSScanner{iface: iface}, nil
}

func (s *MDNSScanner) Name() string { return "mdns" }

func (s *MDNSScanner) Scan(ctx context.Context, out chan<- *Device) error {
	const window = 10 * time.Second
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		entriesCh := make(chan *hashimdns.ServiceEntry, 256)
		params := hashimdns.DefaultParams(mdnsQuery)
		params.Entries = entriesCh
		params.Interface = s.iface.Interface
		params.DisableIPv6 = true
		params.Timeout = window
		params.Logger = log.New(io.Discard, "", 0)

		go func() {
			hashimdns.Query(params)
			close(entriesCh)
		}()

	drain:
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case entry, ok := <-entriesCh:
				if !ok {
					break drain
				}
				if entry.AddrV4 == nil {
					continue
				}

				// FIX: clean the hostname here, before the device leaves
				// the scanner. Uses the improved cleaner below.
				cleanName := cleanMDNSHostname(entry.Name)
				if cleanName == "" {
					continue
				}

				d := NewDevice(entry.AddrV4)
				d.Proto = "mdns"
				d.Name = cleanName

				d.AddMeta("host", entry.Host)
				d.AddMeta("port", fmt.Sprintf("%d", entry.Port))
				for _, f := range entry.InfoFields {
					if kv := splitKV(f); kv != nil {
						d.AddMeta(kv[0], kv[1])
					} else {
						d.AddMeta(f, "true")
					}
				}

				select {
				case out <- d:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}

func cleanMDNSHostname(rawName string) string {
	// ── Try JSON metadata prefix (cases 2 and 3) ─────────────────────────
	if strings.HasPrefix(rawName, "{") || strings.HasPrefix(rawName, `{\"`) {
		name := extractFromJSONPrefix(rawName)
		if name != "" {
			return name
		}
	}

	// ── Strip service type suffixes ───────────────────────────────────────
	cleanName := rawName
	suffixes := []string{
		"._device-info._tcp.local.",
		"._mi-connect._udp.local.",
		"._airplay._tcp.local.",
		"._raop._tcp.local.",
		"._homekit._tcp.local.",
		"._hap._tcp.local.",
		"._companion-link._tcp.local.",
		"._sleep-proxy._udp.local.",
		"._services._dns-sd._udp.local.",
		"._myapp._tcp.local.",
		".local.",
		"._tcp.local.",
		"._udp.local.",
	}
	for _, suffix := range suffixes {
		if idx := strings.Index(cleanName, suffix); idx > 0 {
			cleanName = cleanName[:idx]
			break
		}
	}

	// Remove any remaining ._* patterns
	if idx := strings.Index(cleanName, "._"); idx > 0 {
		cleanName = cleanName[:idx]
	}

	cleanName = strings.Trim(cleanName, ". \t\r\n")

	if cleanName == "" || cleanName == "{}" {
		return ""
	}
	return cleanName
}

func extractFromJSONPrefix(raw string) string {
	// Find the closing brace of the JSON object.
	// We need the FIRST '}' that closes the top-level object.
	end := strings.Index(raw, "}")
	if end < 0 {
		return ""
	}
	jsonPart := raw[:end+1]

	// ── Attempt 1: parse as-is (standard JSON) ───────────────────────────
	if name := parseJSONForName(jsonPart); name != "" {
		return name
	}

	// ── Attempt 2: unescape Go-style escaped quotes ───────────────────────
	// Replace \" with " so the string becomes valid JSON.
	unescaped := strings.ReplaceAll(jsonPart, `\"`, `"`)
	if name := parseJSONForName(unescaped); name != "" {
		return name
	}

	// ── Attempt 3: remove surrounding braces and parse key="value" pairs ─
	// Last resort for non-standard formats.
	inner := strings.Trim(jsonPart, `{}\"`)
	for _, pair := range strings.Split(inner, ",") {
		pair = strings.TrimSpace(pair)
		// Handle both: nm="Kaimunchi" and "nm":"Kaimunchi"
		pair = strings.ReplaceAll(pair, `\"`, `"`)
		pair = strings.Trim(pair, `"`)
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			kv = strings.SplitN(pair, "=", 2)
		}
		if len(kv) == 2 {
			k := strings.Trim(strings.TrimSpace(kv[0]), `"`)
			v := strings.Trim(strings.TrimSpace(kv[1]), `"`)
			if strings.EqualFold(k, "nm") && v != "" && isPrintableASCII(v) {
				return v
			}
		}
	}

	return ""
}

func parseJSONForName(s string) string {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return ""
	}
	if nm, ok := m["nm"].(string); ok && nm != "" {
		return nm
	}
	// Some devices use "name" instead of "nm"
	if name, ok := m["name"].(string); ok && name != "" {
		return name
	}
	return ""
}

// isPrintableASCII returns true if s contains only printable ASCII characters.
func isPrintableASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

const (
	ssdpMulticast = "239.255.255.250"
	ssdpPort      = 1900
)

type SSDPScanner struct {
	iface *InterfaceInfo
}

func NewSSDP(iface *InterfaceInfo) (*SSDPScanner, error) {
	if iface == nil {
		return nil, fmt.Errorf("interface info is nil")
	}
	return &SSDPScanner{iface: iface}, nil
}

func (s *SSDPScanner) Name() string { return "ssdp" }

func (s *SSDPScanner) Scan(ctx context.Context, out chan<- *Device) error {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: ssdpPort,
	})
	if err != nil {
		return fmt.Errorf("cannot bind :1900 (need root): %w", err)
	}
	defer conn.Close()

	pc := ipv4.NewPacketConn(conn)
	if err := pc.JoinGroup(s.iface.Interface, &net.UDPAddr{
		IP: net.ParseIP(ssdpMulticast),
	}); err != nil {
		return fmt.Errorf("join multicast group: %w", err)
	}
	log.Printf("[SSDP] Joined %s:%d on %s", ssdpMulticast, ssdpPort, s.iface.Interface.Name)

	if err := s.sendMSearch(conn); err != nil {
		log.Printf("[SSDP] Initial M-SEARCH error: %v", err)
	}

	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.sendMSearch(conn); err != nil {
					log.Printf("[SSDP] M-SEARCH error: %v", err)
				}
			}
		}
	}()

	buf := make([]byte, 8192)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn.SetReadDeadline(time.Now().Add(time.Second))
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			log.Printf("[SSDP] Read error: %v", err)
			continue
		}

		d := s.parsePacket(src, buf[:n])
		if d == nil {
			continue
		}
		if d.Meta["nts"] == "ssdp:byebye" {
			continue
		}

		select {
		case out <- d:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *SSDPScanner) sendMSearch(conn *net.UDPConn) error {
	addr := &net.UDPAddr{IP: net.ParseIP(ssdpMulticast), Port: ssdpPort}
	msg := fmt.Sprintf(
		"M-SEARCH * HTTP/1.1\r\n"+
			"HOST: %s:%d\r\n"+
			"MAN: \"ssdp:discover\"\r\n"+
			"MX: 3\r\n"+
			"ST: ssdp:all\r\n"+
			"USER-AGENT: vlan-scanner/1.0\r\n"+
			"\r\n",
		ssdpMulticast, ssdpPort,
	)
	_, err := conn.WriteToUDP([]byte(msg), addr)
	return err
}

func (s *SSDPScanner) parsePacket(src *net.UDPAddr, buf []byte) *Device {
	data := buf
	if !bytes.HasSuffix(data, []byte("\r\n\r\n")) {
		data = append(append([]byte{}, data...), "\r\n\r\n"...)
	}

	br := bufio.NewReader(bytes.NewReader(data))
	tr := textproto.NewReader(br)

	firstLine, err := tr.ReadLine()
	if err != nil {
		return nil
	}
	hdr, _ := tr.ReadMIMEHeader()

	isNotify := strings.HasPrefix(firstLine, "NOTIFY")
	isOK := strings.HasPrefix(firstLine, "HTTP/1.1 200")
	if !isNotify && !isOK {
		return nil
	}

	d := NewDevice(src.IP)
	d.Proto = "ssdp"

	server := strings.TrimSpace(hdr.Get("Server"))
	if server != "" {
		d.Name = cleanSSDPServer(server)
	}

	d.AddMeta("usn", strings.TrimSpace(hdr.Get("Usn")))
	d.AddMeta("nt", strings.TrimSpace(hdr.Get("Nt")))
	d.AddMeta("nts", strings.TrimSpace(hdr.Get("Nts")))
	d.AddMeta("location", strings.TrimSpace(hdr.Get("Location")))
	d.AddMeta("server", server)

	if d.Name == "" {
		if loc := hdr.Get("Location"); loc != "" {
			d.Name = extractDeviceNameFromURL(loc)
		}
	}
	if d.Name == "" {
		return nil
	}
	return d
}

func cleanSSDPServer(server string) string {
	s := strings.TrimSpace(server)
	// Remove "UPnP/1.0 " style version prefixes
	if idx := strings.Index(s, " "); idx > 0 {
		parts := strings.SplitN(s, " ", 2)
		if strings.Contains(parts[0], "/") {
			s = parts[1]
		}
	}
	s = strings.TrimSuffix(s, " UPnP/1.0")
	s = strings.TrimSuffix(s, " DLNADOC/1.50")
	return strings.TrimSpace(s)
}

func extractDeviceNameFromURL(location string) string {
	if !strings.HasPrefix(location, "http") {
		return ""
	}
	parts := strings.Split(location, "/")
	if len(parts) > 2 {
		hostPort := parts[2]
		host := strings.Split(hostPort, ":")[0]
		if net.ParseIP(host) != nil {
			return "" // just an IP, not a name
		}
		return strings.TrimSuffix(host, ".local")
	}
	return ""
}

// ============================================
// UTILS
// ============================================

func splitKV(s string) []string {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) == 2 {
		return parts
	}
	return nil
}
