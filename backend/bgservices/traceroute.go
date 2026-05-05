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

	rawConn, err := net.ListenPacket("ip4:1", "0.0.0.0") // prtocol 1 icmp
	if err != nil {
		SendTracerouteError(conn, fmt.Sprintf("Failed to create ICMP socket (requires root privileges): %v", err))
		return
	}
	defer rawConn.Close()
	ipConn := ipv4.NewPacketConn(rawConn)
	if ipConn == nil {
		SendTracerouteError(conn, "Failed to create IPv4 packet connection")
		return
	}
	SendTracerouteStatus(conn, "Starting ICMP traceroute...")
	for ttl := 1; ttl <= req.MaxHops; ttl++ {
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
		if err := ipConn.SetTTL(ttl); err != nil {
			log.Printf("Failed to set TTL to %d: %v", ttl, err)
			continue
		}
		var rtts []float64
		var ip net.IP

		for probe := 0; probe < req.ProbesPerHop; probe++ {
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
			if err := rawConn.SetReadDeadline(time.Now().Add(time.Duration(req.Timeout) * time.Millisecond)); err != nil {
				log.Printf("Failed to set read deadline: %v", err)
				continue
			}
			start := time.Now()
			_, err = ipConn.WriteTo(msgBytes, nil, &net.IPAddr{IP: target})
			if err != nil {
				log.Printf("Failed to send ICMP packet: %v", err)
				continue
			}
			recvBuf := make([]byte, 1500)
			n, _, peer, err := ipConn.ReadFrom(recvBuf)
			if err != nil {
				continue
			}
			recvMsg, err := icmp.ParseMessage(1, recvBuf[:n])
			if err != nil {
				log.Printf("Failed to parse ICMP message: %v", err)
				continue
			}
			rtt := float64(time.Since(start).Microseconds()) / 1000.0

			switch recvMsg.Type {
			case ipv4.ICMPTypeTimeExceeded:
				rtts = append(rtts, rtt)
				if ip == nil {
					ip = peer.(*net.IPAddr).IP
				}
				hopResult.Status = "success"

			case ipv4.ICMPTypeEchoReply:
				rtts = append(rtts, rtt)
				if ip == nil {
					ip = peer.(*net.IPAddr).IP
				}
				hopResult.Status = "success"
				hopResult.IsTarget = true
			}
		}
		if ip != nil {
			hopResult.IP = ip.String()
			hostnameChan := make(chan string, 1)
			go func() {
				names, err := net.LookupAddr(ip.String())
				if err == nil && len(names) > 0 {
					hostname := strings.TrimSuffix(names[0], ".")
					hostnameChan <- hostname
				} else {
					hostnameChan <- ""
				}
			}()
			select {
			case hostname := <-hostnameChan:
				if hostname != "" {
					hopResult.Hostname = hostname
					hopResult.ReverseDNS = hostname
				}
			case <-time.After(1 * time.Second):
			}
			if len(rtts) > 0 {
				var sum float64
				for _, rtt := range rtts {
					sum += rtt
				}
				hopResult.RTT = sum / float64(len(rtts))
			}
			enrichHopInfo(&hopResult)
		}
		mu.Lock()
		hops = append(hops, hopResult)
		mu.Unlock()
		SendTracerouteMessage(conn, utils.TracerouteMessage{
			Type: "hopResult",
			Hop:  &hopResult,
		})
		if hopResult.IsTarget {
			SendTracerouteStatus(conn, "Destination reached!")
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
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

func enrichHopInfo(hop *utils.HopResult) {
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
