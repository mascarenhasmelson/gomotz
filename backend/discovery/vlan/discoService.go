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
	if strings.HasPrefix(rawName, "{") || strings.HasPrefix(rawName, `{\"`) {
		name := extractFromJSONPrefix(rawName)
		if name != "" {
			return name
		}
	}
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
	end := strings.Index(raw, "}")
	if end < 0 {
		return ""
	}
	jsonPart := raw[:end+1]
	if name := parseJSONForName(jsonPart); name != "" {
		return name
	}
	unescaped := strings.ReplaceAll(jsonPart, `\"`, `"`)
	if name := parseJSONForName(unescaped); name != "" {
		return name
	}
	inner := strings.Trim(jsonPart, `{}\"`)
	for _, pair := range strings.Split(inner, ",") {
		pair = strings.TrimSpace(pair)
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
	if name, ok := m["name"].(string); ok && name != "" {
		return name
	}
	return ""
}
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

func splitKV(s string) []string {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) == 2 {
		return parts
	}
	return nil
}
