package bgservices

import (
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/mascarenhasmelson/gomotz/utils"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/gorilla/websocket"
)

var wsMu sync.Mutex

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsSend(ws *websocket.Conn, msg utils.ScanMessage) {
	wsMu.Lock()
	defer wsMu.Unlock()
	b, _ := json.Marshal(msg)
	ws.WriteMessage(websocket.TextMessage, b)
}

func HandleSynScanWS(w http.ResponseWriter, r *http.Request) {
	ws, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	var req utils.ScanRequest
	if err := ws.ReadJSON(&req); err != nil {
		return
	}

	go SynScan(ws, req.Target)
	for {
		time.Sleep(10 * time.Second)
		wsMu.Lock()
		err := ws.WriteMessage(websocket.PingMessage, nil)
		wsMu.Unlock()
		if err != nil {
			return
		}
	}
}

func SynScan(ws *websocket.Conn, host string) {

	wsSend(ws, utils.ScanMessage{Type: "status", Message: "Resolving host"})

	ips, err := net.LookupIP(host)
	if err != nil {
		wsSend(ws, utils.ScanMessage{Type: "error", Message: err.Error()})
		return
	}
	var dstIP net.IP
	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			dstIP = ip4
			break
		}
	}
	if dstIP == nil {
		wsSend(ws, utils.ScanMessage{Type: "error", Message: "No IPv4 address found"})
		return
	}
	srcIP := utils.LocalIP()
	ip := &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}
	conn, err := net.ListenPacket("ip4:tcp", "0.0.0.0")
	if err != nil {
		wsSend(ws, utils.ScanMessage{Type: "error", Message: err.Error()})
		return
	}
	defer conn.Close()
	if sc, ok := conn.(syscall.Conn); ok {
		if raw, err := sc.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_SNDBUF, 4*1024*1024)
			})
		}
	}
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	openPorts := make(map[layers.TCPPort]bool)
	var mu sync.Mutex
	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	wsSend(ws, utils.ScanMessage{
		Type:    "status",
		Message: "Starting SYN scan",
	})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go listenForResponses(ws, conn, dstIP, openPorts, &mu, &wg, stopChan)
	}
	const (
		startPort = 1
		endPort   = 65535
		workers   = 3
	)
	chunk := (endPort - startPort + 1) / workers
	var sendWG sync.WaitGroup

	for i := 0; i < workers; i++ {
		s := startPort + i*chunk
		e := s + chunk - 1
		if i == workers-1 {
			e = endPort
		}
		sendWG.Add(1)
		go sendSynChunk(ws, conn, ip, dstIP, s, e, opts, &sendWG)
	}
	sendWG.Wait()
	wsSend(ws, utils.ScanMessage{Type: "status", Message: "Packets sent, waiting for responses"})
	time.Sleep(7 * time.Second)
	close(stopChan)
	wg.Wait()
	mu.Lock()
	var ports []int
	for p := range openPorts {
		ports = append(ports, int(p))
	}
	mu.Unlock()
	sort.Ints(ports)
	wsSend(ws, utils.ScanMessage{
		Type:      "complete",
		OpenPorts: ports,
		Message:   "Scan finished",
	})
}

func listenForResponses(
	ws *websocket.Conn,
	conn net.PacketConn,
	dstIP net.IP,
	openPorts map[layers.TCPPort]bool,
	mu *sync.Mutex,
	wg *sync.WaitGroup,
	stopChan chan struct{},
) {
	defer wg.Done()

	buf := make([]byte, 65535)

	for {
		select {
		case <-stopChan:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				continue
			}

			ipAddr, ok := addr.(*net.IPAddr)
			if !ok || !ipAddr.IP.Equal(dstIP) {
				continue
			}

			p := gopacket.NewPacket(buf[:n], layers.LayerTypeTCP, gopacket.Default)
			tcpLayer := p.Layer(layers.LayerTypeTCP)
			if tcpLayer == nil {
				continue
			}

			tcp := tcpLayer.(*layers.TCP)
			if tcp.SYN && tcp.ACK {
				mu.Lock()
				if !openPorts[tcp.SrcPort] {
					openPorts[tcp.SrcPort] = true
					wsSend(ws, utils.ScanMessage{
						Type: "open",
						Port: int(tcp.SrcPort),
					})
				}
				mu.Unlock()
			}
		}
	}
}

func sendSynChunk(
	ws *websocket.Conn,
	conn net.PacketConn,
	ip *layers.IPv4,
	dstIP net.IP,
	startPort, endPort int,
	opts gopacket.SerializeOptions,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	srcPort := layers.TCPPort(40000 + rand.Intn(20000))
	batchSize := 50

	for port := startPort; port <= endPort; port += batchSize {
		bEnd := port + batchSize
		if bEnd > endPort {
			bEnd = endPort
		}

		for p := port; p < bEnd; p++ {
			tcp := &layers.TCP{
				SrcPort: srcPort,
				DstPort: layers.TCPPort(p),
				Seq:     rand.Uint32(),
				SYN:     true,
				Window:  14600,
			}
			tcp.SetNetworkLayerForChecksum(ip)

			buf := gopacket.NewSerializeBuffer()
			if err := gopacket.SerializeLayers(buf, opts, tcp); err == nil {
				conn.WriteTo(buf.Bytes(), &net.IPAddr{IP: dstIP})
			}
			time.Sleep(20 * time.Microsecond)
		}

		wsSend(ws, utils.ScanMessage{
			Type:    "progress",
			Scanned: bEnd,
		})

		time.Sleep(10 * time.Millisecond)
	}
}
