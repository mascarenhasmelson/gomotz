package bgservices

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mascarenhasmelson/gomotz/utils"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var TracerouteUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func PerformICMPTraceroute(conn *websocket.Conn, target net.IP, req utils.TracerouteRequest) {
	startTime := time.Now()
	var hops []utils.HopResult
	var mu sync.Mutex

	rawConn, err := net.ListenPacket("ip4:1", "0.0.0.0") // Protocol 1 = ICMP
	if err != nil {
		SendTracerouteError(conn, fmt.Sprintf("Failed to create ICMP socket (requires root privileges): %v", err))
		return
	}
	defer rawConn.Close()

	// Get IPv4 packet connection for setting TTL
	ipConn := ipv4.NewPacketConn(rawConn)
	if ipConn == nil {
		SendTracerouteError(conn, "Failed to create IPv4 packet connection")
		return
	}

	SendTracerouteStatus(conn, "Starting ICMP traceroute...")

	for ttl := 1; ttl <= req.MaxHops; ttl++ {
		// Send progress update
		SendTracerouteMessage(conn, utils.TracerouteMessage{
			Type: "progress",
			Progress: &utils.Progress{
				CurrentHop: ttl,
				TotalHops:  req.MaxHops,
			},
		})

		hopResult := utils.HopResult{
			Hop:    ttl,
			Status: "timeout",
		}

		// Set TTL for this hop
		if err := ipConn.SetTTL(ttl); err != nil {
			log.Printf("Failed to set TTL to %d: %v", ttl, err)
			continue
		}

		// Send multiple probes per hop
		var rtts []float64
		var ip net.IP

		for probe := 0; probe < req.ProbesPerHop; probe++ {
			// Create ICMP echo request
			msg := icmp.Message{
				Type: ipv4.ICMPTypeEcho,
				Code: 0,
				Body: &icmp.Echo{
					ID:   os.Getpid() & 0xffff,
					Seq:  ttl*req.ProbesPerHop + probe,
					Data: []byte("TRACEROUTE_PROBE"),
				},
			}

			msgBytes, err := msg.Marshal(nil)
			if err != nil {
				log.Printf("Failed to marshal ICMP message: %v", err)
				continue
			}

			// Set read deadline for this probe
			if err := rawConn.SetReadDeadline(time.Now().Add(time.Duration(req.Timeout) * time.Millisecond)); err != nil {
				log.Printf("Failed to set read deadline: %v", err)
				continue
			}

			// Send packet
			start := time.Now()
			_, err = ipConn.WriteTo(msgBytes, nil, &net.IPAddr{IP: target})
			if err != nil {
				log.Printf("Failed to send ICMP packet: %v", err)
				continue
			}

			// Receive response
			recvBuf := make([]byte, 1500)
			n, _, peer, err := ipConn.ReadFrom(recvBuf)
			if err != nil {
				// Timeout or other error - no response for this probe
				continue
			}

			// Parse response
			recvMsg, err := icmp.ParseMessage(1, recvBuf[:n])
			if err != nil {
				log.Printf("Failed to parse ICMP message: %v", err)
				continue
			}

			// Calculate RTT
			rtt := float64(time.Since(start).Microseconds()) / 1000.0 // Convert to milliseconds

			switch recvMsg.Type {
			case ipv4.ICMPTypeTimeExceeded:
				// Intermediate hop
				rtts = append(rtts, rtt)
				if ip == nil {
					ip = peer.(*net.IPAddr).IP
				}
				hopResult.Status = "success"

			case ipv4.ICMPTypeEchoReply:
				// Destination reached
				rtts = append(rtts, rtt)
				if ip == nil {
					ip = peer.(*net.IPAddr).IP
				}
				hopResult.Status = "success"
				hopResult.IsTarget = true
			}
		}

		// Process hop results
		if ip != nil {
			hopResult.IP = ip.String()

			// Get hostname (non-blocking with timeout)
			hostnameChan := make(chan string, 1)
			go func() {
				names, err := net.LookupAddr(ip.String())
				if err == nil && len(names) > 0 {
					// Remove trailing dot from hostname
					hostname := strings.TrimSuffix(names[0], ".")
					hostnameChan <- hostname
				} else {
					hostnameChan <- ""
				}
			}()

			// Wait for hostname lookup with timeout
			select {
			case hostname := <-hostnameChan:
				if hostname != "" {
					hopResult.Hostname = hostname
					hopResult.ReverseDNS = hostname
				}
			case <-time.After(1 * time.Second):
				// Hostname lookup timed out
			}

			// Calculate average RTT if we have measurements
			if len(rtts) > 0 {
				var sum float64
				for _, rtt := range rtts {
					sum += rtt
				}
				hopResult.RTT = sum / float64(len(rtts))
			}

			// Enrich hop info before sending
			enrichHopInfo(&hopResult)
		}

		// Add hop to results
		mu.Lock()
		hops = append(hops, hopResult)
		mu.Unlock()

		// Send hop result to client
		SendTracerouteMessage(conn, utils.TracerouteMessage{
			Type: "hopResult",
			Hop:  &hopResult,
		})

		// Check if we reached the target
		if hopResult.IsTarget {
			SendTracerouteStatus(conn, "Destination reached!")
			break
		}

		// Small delay between hops to avoid flooding
		time.Sleep(50 * time.Millisecond)
	}

	// Send completion
	SendTracerouteMessage(conn, utils.TracerouteMessage{
		Type: "complete",
		Complete: &utils.Complete{
			Reached:   len(hops) > 0 && hops[len(hops)-1].IsTarget,
			TargetIP:  target.String(),
			Hops:      hops,
			TotalTime: time.Since(startTime).Seconds(),
		},
	})
}

// func performUDPTraceroute(conn *websocket.Conn, target net.IP, req TracerouteRequest) {
// 	sendTracerouteStatus(conn, "UDP traceroute not yet implemented")
// 	SendTracerouteError(conn, "UDP protocol is not implemented yet. Please use ICMP.")
// }

// func performTCPTraceroute(conn *websocket.Conn, target net.IP, req TracerouteRequest) {
// 	sendTracerouteStatus(conn, "TCP traceroute not yet implemented")
// 	SendTracerouteError(conn, "TCP protocol is not implemented yet. Please use ICMP.")
// }

func enrichHopInfo(hop *utils.HopResult) {
	// Enrich hop info with ASN, location, ISP, etc.
	// Using ip-api.com (free tier with rate limits)

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,city,isp,as,org", hop.IP))
	if err != nil {
		log.Printf("Failed to enrich hop info for %s: %v", hop.IP, err)
		return
	}
	defer resp.Body.Close()

	var info struct {
		Status  string `json:"status"`
		AS      string `json:"as"`
		Country string `json:"country"`
		City    string `json:"city"`
		ISP     string `json:"isp"`
		Org     string `json:"org"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Printf("Failed to decode hop info for %s: %v", hop.IP, err)
		return
	}

	if info.Status == "success" {
		// Extract ASN from "AS" field (format: "AS15169 Google LLC")
		if info.AS != "" {
			parts := strings.Fields(info.AS)
			if len(parts) > 0 {
				hop.ASN = parts[0]
			}
		}
		hop.Country = info.Country
		hop.Location = info.City
		hop.ISP = info.ISP
	}
}

func SendTracerouteMessage(conn *websocket.Conn, msg utils.TracerouteMessage) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
		log.Printf("Failed to send WebSocket message: %v", err)
	}
}

func SendTracerouteError(conn *websocket.Conn, errMsg string) {
	SendTracerouteMessage(conn, utils.TracerouteMessage{
		Type:  "error",
		Error: errMsg,
	})
}

func SendTracerouteStatus(conn *websocket.Conn, status string) {
	SendTracerouteMessage(conn, utils.TracerouteMessage{
		Type:   "status",
		Status: status,
	})
}
