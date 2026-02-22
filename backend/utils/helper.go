package utils

import "net"

func LocalIP() net.IP {

	if ip := localIPVia("1.1.1.1:53"); ip != nil {
		return ip
	}

	if ip := localIPVia("8.8.8.8:53"); ip != nil {
		return ip
	}

	return net.IPv4(127, 0, 0, 1)
}

func localIPVia(addr string) net.IP {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil
	}
	defer conn.Close()

	if udpAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		return udpAddr.IP
	}

	return nil
}
