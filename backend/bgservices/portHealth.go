package bgservices

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/mascarenhasmelson/gomotz/utils"

	"github.com/jackc/pgx/v4/pgxpool"
)

var service utils.Service
var workerLimit = make(chan struct{}, 200)

// func checkPort(ip string, port int) bool {
// 	address := fmt.Sprintf("%s:%d", ip, port)
// 	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
// 	if err != nil {
// 		return false
// 	}
// 	conn.Close()
// 	return true
// }
func CheckPort(host string, port int) bool {
	ips, err := net.LookupIP(host)
	if err != nil {
		log.Printf("DNS resolution failed for %s: %v", host, err)
		return false
	}
	var dstIP net.IP
	for _, ip := range ips {
		if ip4 := ip.To4(); ip4 != nil {
			dstIP = ip4
			break
		}
	}
	if dstIP == nil {
		log.Printf("No IPv4 address found for host: %s", host)
		return false
	}

	srcIP := utils.LocalIP()
	srcPort := layers.TCPPort(43591)

	ip := &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Protocol: layers.IPProtocolTCP,
	}

	conn, err := net.ListenPacket("ip4:tcp", "0.0.0.0")
	if err != nil {
		log.Printf("Failed to create raw socket: %v", err)
		return false
	}
	defer conn.Close()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	tcp := &layers.TCP{
		SrcPort: srcPort,
		DstPort: layers.TCPPort(port),
		Seq:     1105024978,
		SYN:     true,
		Window:  14600,
	}
	tcp.SetNetworkLayerForChecksum(ip)
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf, opts, tcp); err != nil {
		log.Printf("Failed to serialize packet: %v", err)
		return false
	}
	resultChan := make(chan bool, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 65535)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		for {
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				resultChan <- false
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
			tcpResponse := tcpLayer.(*layers.TCP)
			if tcpResponse.DstPort == srcPort &&
				tcpResponse.SrcPort == layers.TCPPort(port) {
				if tcpResponse.SYN && tcpResponse.ACK {
					log.Printf("Port %d on %s is OPEN", port, host)
					resultChan <- true
					return
				} else if tcpResponse.RST {
					log.Printf("Port %d on %s is CLOSED", port, host)
					resultChan <- false
					return
				}
			}
		}
	}()
	_, err = conn.WriteTo(buf.Bytes(), &net.IPAddr{IP: dstIP})
	if err != nil {
		log.Printf("Failed to send packet: %v", err)
		return false
	}
	select {
	case result := <-resultChan:
		wg.Wait()
		return result
	case <-time.After(2 * time.Second):
		log.Printf("Timeout checking port %d on %s", port, host)
		return false
	}
}
func updateStatus(ctx context.Context, db *pgxpool.Pool, id int, online bool) {
	if online {
		_, _ = db.Exec(ctx,
			`UPDATE services SET online = $1, last_seen = NOW() WHERE id = $2`,
			true, id)
	} else {
		_, _ = db.Exec(ctx,
			`UPDATE services SET online = $1 WHERE id = $2`,
			false, id)
	}
}

func StartPortMonitor(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("Port monitoring service started...")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rows, err := pool.Query(ctx,
				`SELECT id, host(remote_ip) AS remote_ip, remote_port FROM services`)
			if err != nil {
				fmt.Println("DB error:", err)
				continue
			}
			for rows.Next() {
				var s utils.Service
				if err := rows.Scan(&s.ID, &s.RemoteIP, &s.RemotePort); err != nil {
					fmt.Println("Scan error:", err)
					continue
				}
				go func(s utils.Service) {
					workerLimit <- struct{}{}
					defer func() { <-workerLimit }()
					online := CheckPort(s.RemoteIP, s.RemotePort)
					updateStatus(ctx, pool, s.ID, online)
				}(s)
			}
			rows.Close()
		case <-ctx.Done():
			fmt.Println("Port monitor stopped.")
			return
		}
	}
}
