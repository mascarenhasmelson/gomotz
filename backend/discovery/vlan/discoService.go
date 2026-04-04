package vlan

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	hashimdns "github.com/hashicorp/mdns"
)

//
// ====================== CORE TYPES ======================
//

type Device struct {
	IP    net.IP
	Name  string
	Meta  map[string]string
	Proto string
}

func NewDevice(ip net.IP) *Device {
	return &Device{
		IP:   ip,
		Meta: make(map[string]string),
	}
}

func (d *Device) AddMeta(k, v string) {
	if d.Meta == nil {
		d.Meta = make(map[string]string)
	}
	d.Meta[k] = v
}

type Scanner interface {
	Name() string
	Scan(ctx context.Context, out chan<- *Device) error
}

type InterfaceInfo struct {
	Interface *net.Interface
	IPv4      net.IP
}

// ✅ Helper to get interface by name
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

type Logger interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
}

type NoOpLogger struct{}

func (NoOpLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {}

type Option func(any) error

func WithLogger(logger Logger) Option {
	return func(s any) error {
		if logger == nil {
			return errors.New("logger nil")
		}
		switch v := s.(type) {
		case *MDNSScanner:
			v.logger = logger
		case *SSDPScanner:
			v.logger = logger
		}
		return nil
	}
}

//
// ====================== MDNS ======================
//

const mdnsQuery = "_services._dns-sd._udp"

type MDNSScanner struct {
	iface     *InterfaceInfo
	logger    Logger
	queryFunc func(*hashimdns.QueryParam) error
}

func NewMDNS(iface *InterfaceInfo, opts ...Option) (*MDNSScanner, error) {
	if iface == nil {
		return nil, errors.New("interface info is nil")
	}
	s := &MDNSScanner{
		iface:     iface,
		logger:    NoOpLogger{},
		queryFunc: hashimdns.Query,
	}
	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *MDNSScanner) Name() string { return "mdns" }

func extractCleanHostname(rawName string) string {
	// Handle mDNS service discovery format
	// Example: "{\"nm\":\"Kaimunchi\",\"as\":\"[8194]\",\"ip\":\"0\"}._mi-connect._udp.local."

	// Try to parse JSON metadata
	if strings.HasPrefix(rawName, "{") {
		endJSON := strings.Index(rawName, "}")
		if endJSON > 0 {
			jsonPart := rawName[:endJSON+1]

			// Parse the JSON to extract the name
			var metadata map[string]interface{}
			if err := json.Unmarshal([]byte(jsonPart), &metadata); err == nil {
				if nm, ok := metadata["nm"].(string); ok && nm != "" {
					return nm
				}
			}
		}
	}

	// Remove service type suffixes
	// Example: "MyDevice._device-info._tcp.local." → "MyDevice"
	cleanName := rawName

	// Common mDNS service patterns to remove
	serviceSuffixes := []string{
		"._device-info._tcp.local.",
		"._mi-connect._udp.local.",
		"._airplay._tcp.local.",
		"._raop._tcp.local.",
		"._homekit._tcp.local.",
		"._hap._tcp.local.",
		"._companion-link._tcp.local.",
		"._sleep-proxy._udp.local.",
		"._services._dns-sd._udp.local.",
		".local.",
		"._tcp.local.",
		"._udp.local.",
	}

	for _, suffix := range serviceSuffixes {
		if idx := strings.Index(cleanName, suffix); idx > 0 {
			cleanName = cleanName[:idx]
			break
		}
	}

	// Remove any remaining ._* patterns
	if idx := strings.Index(cleanName, "._"); idx > 0 {
		cleanName = cleanName[:idx]
	}

	// Trim whitespace and dots
	cleanName = strings.Trim(cleanName, ". \t\r\n")

	// If empty after cleaning, return empty string
	if cleanName == "" || cleanName == "{}" {
		return ""
	}

	return cleanName
}

func (s *MDNSScanner) Scan(ctx context.Context, out chan<- *Device) error {
	entries := make(chan *hashimdns.ServiceEntry, 256)
	errCh := make(chan error, 1)

	go func() {
		p := hashimdns.DefaultParams(mdnsQuery)
		p.Entries = entries
		p.Interface = s.iface.Interface
		p.DisableIPv6 = true

		p.Logger = log.Default()
		p.Logger.SetOutput(io.Discard)

		if dl, ok := ctx.Deadline(); ok {
			p.Timeout = time.Until(dl)
		}

		if err := s.queryFunc(p); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case e, ok := <-entries:
			if !ok {
				select {
				case err := <-errCh:
					return err
				default:
					return nil
				}
			}

			if e.AddrV4 == nil {
				continue
			}

			d := NewDevice(e.AddrV4)

			// ✅ Clean the hostname before using it
			cleanName := extractCleanHostname(e.Name)
			if cleanName != "" {
				d.Name = cleanName
			} else {
				// Skip devices with no meaningful name
				continue
			}

			d.Proto = "mdns"

			// Parse info fields
			for _, f := range e.InfoFields {
				if kv := splitKV(f); kv != nil {
					d.AddMeta(kv[0], kv[1])
				} else {
					d.AddMeta(f, "true")
				}
			}

			s.logger.Log(ctx, slog.LevelDebug, "mdns device",
				"name", d.Name, "ip", d.IP.String())

			select {
			case out <- d:
			case <-ctx.Done():
				return ctx.Err()
			}

		case err := <-errCh:
			return err
		}
	}
}

//
// ====================== SSDP ======================
//

const ssdpAddr = "239.255.255.250:1900"

type SSDPScanner struct {
	iface  *InterfaceInfo
	logger Logger
}

func NewSSDP(iface *InterfaceInfo, opts ...Option) (*SSDPScanner, error) {
	if iface == nil {
		return nil, errors.New("interface info is nil")
	}
	s := &SSDPScanner{
		iface:  iface,
		logger: NoOpLogger{},
	}
	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *SSDPScanner) Name() string { return "ssdp" }

func (s *SSDPScanner) Scan(ctx context.Context, out chan<- *Device) error {
	maddr, err := net.ResolveUDPAddr("udp4", ssdpAddr)
	if err != nil {
		return fmt.Errorf("resolve ssdp addr: %w", err)
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   s.iface.IPv4,
		Port: 0,
	})
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer conn.Close()

	s.logger.Log(ctx, slog.LevelDebug, "sending ssdp m-search",
		"from", conn.LocalAddr().String())

	if err := sendSSDP(conn, maddr); err != nil {
		return err
	}

	// Set deadline from context
	if dl, ok := ctx.Deadline(); ok {
		conn.SetReadDeadline(dl)
	} else {
		// Default timeout if no deadline
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	}

	buf := make([]byte, 8192)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				return nil
			}
			return fmt.Errorf("read ssdp: %w", err)
		}
		handleSSDP(out, src, buf[:n])
	}
}

//
// ====================== SSDP HELPERS ======================
//

func sendSSDP(conn *net.UDPConn, addr *net.UDPAddr) error {
	req := fmt.Sprintf(
		"M-SEARCH * HTTP/1.1\r\n"+
			"HOST: %s\r\n"+
			"MAN: \"ssdp:discover\"\r\n"+
			"MX: 2\r\n"+
			"ST: ssdp:all\r\n"+
			"USER-AGENT: vlan-scanner/1.0\r\n\r\n",
		ssdpAddr,
	)
	_, err := conn.WriteToUDP([]byte(req), addr)
	if err != nil {
		return fmt.Errorf("send m-search: %w", err)
	}
	return nil
}

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
// 	// ✅ Fixed: Use server as name if available
// 	if server != "" {
// 		d.Name = server
// 	}
// 	d.Proto = "ssdp"

// 	if loc != "" {
// 		d.AddMeta("location", loc)
// 	}
// 	if server != "" {
// 		d.AddMeta("server", server)
// 	}

//		select {
//		case out <- d:
//		default:
//		}
//	}
func handleSSDP(out chan<- *Device, src *net.UDPAddr, payload []byte) {
	loc, server := parseHeaders(payload)

	ip := src.IP
	if ip == nil && loc != "" {
		ip = ipFromLocation(loc)
	}
	if ip == nil {
		return
	}

	d := NewDevice(ip)

	// ✅ Clean server name
	if server != "" {
		// Remove common prefixes/suffixes
		cleanServer := strings.TrimSpace(server)

		// Remove HTTP server version info
		// Example: "UPnP/1.0 Smart TV" → "Smart TV"
		if idx := strings.Index(cleanServer, " "); idx > 0 {
			parts := strings.Split(cleanServer, " ")
			// If first part looks like a version (contains /), skip it
			if strings.Contains(parts[0], "/") && len(parts) > 1 {
				cleanServer = strings.Join(parts[1:], " ")
			}
		}

		d.Name = cleanServer
	} else {
		// Skip devices with no name
		return
	}

	d.Proto = "ssdp"

	if loc != "" {
		d.AddMeta("location", loc)
	}
	if server != "" {
		d.AddMeta("server", server)
	}

	select {
	case out <- d:
	default:
	}
}

func parseHeaders(b []byte) (string, string) {
	// Ensure buffer ends with CRLFCRLF
	if !bytes.HasSuffix(b, []byte("\r\n\r\n")) {
		b = append(append([]byte{}, b...), []byte("\r\n\r\n")...)
	}

	tr := textproto.NewReader(bufio.NewReader(bytes.NewReader(b)))
	// Skip status line
	_, _ = tr.ReadLine()

	h, err := tr.ReadMIMEHeader()
	if err != nil {
		return "", ""
	}

	return strings.TrimSpace(h.Get("Location")), strings.TrimSpace(h.Get("Server"))
}

func ipFromLocation(loc string) net.IP {
	u, err := url.Parse(loc)
	if err != nil {
		return nil
	}
	host := u.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return net.ParseIP(host)
}

//
// ====================== UTILS ======================
//

func splitKV(s string) []string {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) == 2 {
		return parts
	}
	return nil
}
