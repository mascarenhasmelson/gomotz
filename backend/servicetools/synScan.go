package servicetools

import (
	"encoding/json"
	"net"
	"net/http"
	"port/utils"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/gorilla/websocket"
)

var wsMu sync.Mutex

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func GetUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

func UpgradeConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return Upgrader.Upgrade(w, r, nil)
}
func localIP(dst net.IP) net.IP {
	conn, _ := net.Dial("udp", dst.String()+":53")
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP
}

// var receivedMsg utils.ScanMessage

func wsSend(conn *websocket.Conn, msg utils.ScanMessage) {
	wsMu.Lock()
	defer wsMu.Unlock()
	b, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, b)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
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
		time.Sleep(5 * time.Second)
		wsMu.Lock()
		err := ws.WriteMessage(websocket.PingMessage, nil)
		wsMu.Unlock()
		if err != nil {
			return
		}
	}
}

func SynScan(ws *websocket.Conn, host string) {
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
	wsSend(ws, utils.ScanMessage{
		Type:    "status",
		Message: "Starting SYN scan",
	})
	srcIP := localIP(dstIP)
	srcPort := layers.TCPPort(43591)
	ip := &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Protocol: layers.IPProtocolTCP,
	}
	conn, err := net.ListenPacket("ip4:tcp", "0.0.0.0")
	if err != nil {
		wsSend(ws, utils.ScanMessage{Type: "error", Message: err.Error()})
		return
	}
	defer conn.Close()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	openPorts := make(map[layers.TCPPort]bool)
	var mu sync.Mutex
	var openList []int
	//sniffer
	go func() {
		buf := make([]byte, 65535)
		for {
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}
			if !addr.(*net.IPAddr).IP.Equal(dstIP) {
				continue
			}
			p := gopacket.NewPacket(buf[:n], layers.LayerTypeTCP, gopacket.Default)
			tcpLayer := p.Layer(layers.LayerTypeTCP)
			if tcpLayer == nil {
				continue
			}
			tcp := tcpLayer.(*layers.TCP)
			if tcp.DstPort != srcPort {
				continue
			}
			if tcp.SYN && tcp.ACK {
				mu.Lock()
				if !openPorts[tcp.SrcPort] {
					openPorts[tcp.SrcPort] = true
					openList = append(openList, int(tcp.SrcPort))
					wsSend(ws, utils.ScanMessage{
						Type: "open",
						Port: int(tcp.SrcPort),
					})
				}
				mu.Unlock()
			}
		}
	}()
	//send
	for port := 1; port <= 65535; port++ {
		tcp := &layers.TCP{
			SrcPort: srcPort,
			DstPort: layers.TCPPort(port),
			Seq:     1105024978,
			SYN:     true,
			Window:  14600,
		}
		tcp.SetNetworkLayerForChecksum(ip)
		buf := gopacket.NewSerializeBuffer()
		gopacket.SerializeLayers(buf, opts, tcp)
		conn.WriteTo(buf.Bytes(), &net.IPAddr{IP: dstIP})
		if port%1000 == 0 {
			wsSend(ws, utils.ScanMessage{
				Type:    "progress",
				Scanned: port,
			})
		}
		time.Sleep(150 * time.Microsecond)
	}
	time.Sleep(3 * time.Second)
	wsSend(ws, utils.ScanMessage{
		Type:      "complete",
		OpenPorts: openList,
		Message:   "Scan finished",
	})
}
