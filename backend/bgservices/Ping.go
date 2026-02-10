package bgservices

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mascarenhasmelson/gomotz/utils"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var Pingupgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

func PerformPing(conn *websocket.Conn, req utils.PingRequest) {
	stats := struct {
		sent      int
		received  int
		latencies []float64
		mutex     sync.Mutex
	}{
		latencies: make([]float64, 0, req.Count),
	}

	// Send banner
	conn.WriteJSON(utils.PingMessage{
		Type:    "message",
		Message: fmt.Sprintf("PING %s %d(%d) bytes of data.", req.Target, req.Size, req.Size+28),
	})

	// Resolve host
	ips, err := net.LookupIP(req.Target)
	if err != nil || len(ips) == 0 {
		conn.WriteJSON(utils.PingMessage{Type: "message", Message: fmt.Sprintf("Could not resolve host: %s", req.Target)})
		return
	}
	dst := &net.IPAddr{IP: ips[0]}

	// Get local IP for comparison
	localIP := getLocalIP(dst.IP)
	if localIP == nil {
		conn.WriteJSON(utils.PingMessage{Type: "message", Message: "Could not determine local IP address"})
		return
	}

	// Create ICMP connection
	connICMP, err := icmp.ListenPacket("ip4:icmp", localIP.String())
	if err != nil {
		// Fallback to listening on all interfaces
		connICMP, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			conn.WriteJSON(utils.PingMessage{Type: "message", Message: fmt.Sprintf("Failed to create ICMP connection: %v", err)})
			return
		}
	}
	defer connICMP.Close()

	// Prepare data
	data := make([]byte, req.Size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	pid := os.Getpid() & 0xffff
	interval := time.Duration(req.Interval * float64(time.Second))

	// Perform pings
	for seq := 0; seq < req.Count; seq++ {
		stats.mutex.Lock()
		stats.sent++
		stats.mutex.Unlock()

		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   pid,
				Seq:  seq + 1,
				Data: data,
			},
		}

		b, err := msg.Marshal(nil)
		if err != nil {
			conn.WriteJSON(utils.PingMessage{
				Type:    "ping_result",
				Data:    map[string]interface{}{"seq": seq + 1, "status": "error"},
				Message: err.Error(),
			})
			time.Sleep(interval)
			continue
		}

		start := time.Now()
		if _, err := connICMP.WriteTo(b, dst); err != nil {
			conn.WriteJSON(utils.PingMessage{
				Type:    "ping_result",
				Data:    map[string]interface{}{"seq": seq + 1, "status": "error"},
				Message: err.Error(),
			})
			time.Sleep(interval)
			continue
		}

		// Set read deadline
		connICMP.SetReadDeadline(time.Now().Add(time.Duration(req.Timeout * float64(time.Second))))

		// Track attempts to receive response
		received := false

		for attempts := 0; attempts < 10; attempts++ { // Multiple attempts to filter out our own packets
			buf := make([]byte, 1500)
			n, peer, err := connICMP.ReadFrom(buf)
			if err != nil {
				// Timeout or error
				break
			}

			rm, err := icmp.ParseMessage(1, buf[:n])
			if err != nil {
				continue
			}

			// Skip echo requests (we're receiving our own packets)
			if rm.Type == ipv4.ICMPTypeEcho {
				continue
			}

			// Check if this is our echo reply
			if rm.Type == ipv4.ICMPTypeEchoReply {
				if echo, ok := rm.Body.(*icmp.Echo); ok {
					// Verify this is our packet by checking ID and sequence
					if echo.ID == pid && echo.Seq == seq+1 {
						// Also verify the peer is the destination
						if isSameIP(peer, dst.IP) {
							rtt := float64(time.Since(start).Microseconds()) / 1000

							result := map[string]interface{}{
								"seq":     seq + 1,
								"status":  "success",
								"latency": rtt,
								"ttl":     64,
								"from":    peer.String(),
							}

							conn.WriteJSON(utils.PingMessage{
								Type:    "ping_result",
								Data:    result,
								Message: fmt.Sprintf("%d bytes from %s: icmp_seq=%d ttl=64 time=%.1f ms", req.Size, peer.String(), seq+1, rtt),
							})

							stats.mutex.Lock()
							stats.received++
							stats.latencies = append(stats.latencies, rtt)
							stats.mutex.Unlock()
							received = true
							break
						}
					}
				}
			}
		}

		if !received {
			// Timeout or invalid response
			conn.WriteJSON(utils.PingMessage{
				Type:    "ping_result",
				Data:    map[string]interface{}{"seq": seq + 1, "status": "timeout"},
				Message: fmt.Sprintf("Request timeout for icmp_seq %d", seq+1),
			})
		}

		time.Sleep(interval)
	}

	// Send summary
	sendSummary(conn, req.Target, stats)
}

func sendSummary(conn *websocket.Conn, target string, stats struct {
	sent      int
	received  int
	latencies []float64
	mutex     sync.Mutex
}) {
	stats.mutex.Lock()
	defer stats.mutex.Unlock()

	packetLoss := 0.0
	if stats.sent > 0 {
		packetLoss = float64(stats.sent-stats.received) / float64(stats.sent) * 100
	}

	min, avg, max := 0.0, 0.0, 0.0
	if len(stats.latencies) > 0 {
		min, max = stats.latencies[0], stats.latencies[0]
		sum := 0.0
		for _, l := range stats.latencies {
			sum += l
			if l < min {
				min = l
			}
			if l > max {
				max = l
			}
		}
		avg = sum / float64(len(stats.latencies))
	}

	// Send summary
	summary := map[string]interface{}{
		"type":       "ping_summary",
		"target":     target,
		"sent":       stats.sent,
		"received":   stats.received,
		"packetLoss": packetLoss,
		"min":        min,
		"avg":        avg,
		"max":        max,
	}

	conn.WriteJSON(utils.PingMessage{
		Type: "ping_summary",
		Data: summary,
	})

	// Send statistics message
	statsMsg := fmt.Sprintf("--- %s ping statistics ---\n%d packets transmitted, %d received, %.1f%% packet loss\n", target, stats.sent, stats.received, packetLoss)
	if len(stats.latencies) > 0 {
		statsMsg += fmt.Sprintf("rtt min/avg/max = %.1f/%.1f/%.1f ms\n", min, avg, max)
	}

	conn.WriteJSON(utils.PingMessage{
		Type:    "message",
		Message: statsMsg,
	})
}

// Helper function to get local IP for destination
func getLocalIP(dst net.IP) net.IP {
	// Try to get local IP by attempting to connect to destination
	conn, err := net.Dial("udp", dst.String()+":53")
	if err != nil {
		// Fallback to getting any local IP
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return net.IPv4(0, 0, 0, 0)
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP
				}
			}
		}

		// If no non-loopback IP found, use loopback
		return net.IPv4(127, 0, 0, 1)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

// Helper function to compare IP addresses
func isSameIP(addr net.Addr, ip net.IP) bool {
	switch v := addr.(type) {
	case *net.IPAddr:
		return v.IP.Equal(ip)
	case *net.UDPAddr:
		return v.IP.Equal(ip)
	case *net.TCPAddr:
		return v.IP.Equal(ip)
	default:
		// Try to parse as string
		addrStr := addr.String()
		if host, _, err := net.SplitHostPort(addrStr); err == nil {
			return net.ParseIP(host).Equal(ip)
		}
		return net.ParseIP(addrStr).Equal(ip)
	}
}
