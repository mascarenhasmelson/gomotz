package utils

import "time"

type Service struct {
	ID           int        `json:"id"`
	Service_name string     `json:"service_name"`
	LocalIP      string     `json:"local_ip"`
	LocalPort    int        `json:"local_port"`
	RemoteIP     string     `json:"remote_ip"`
	RemotePort   int        `json:"remote_port"`
	Online       bool       `json:"online"`
	Lastseen     *time.Time `json:"last_seen"`
	PID          int        `json:"pid"`
}

// type Home struct {
// 	Status      string `json:"status"`
// 	PublicIP    string `json:"publicip"`
// 	ISPInfo     string `json:"ispinfo"`
// 	Interstatus string `json:"interstatus"`
// }
type ScanRequest struct {
	Target string `json:"target"`
}

type ScanMessage struct {
	Type      string `json:"type"`
	Port      int    `json:"port,omitempty"`
	Scanned   int    `json:"scanned,omitempty"`
	Message   string `json:"message,omitempty"`
	OpenPorts []int  `json:"openPorts,omitempty"`
}
type IPInfoRaw struct {
	IP  string `json:"ip"`
	Org string `json:"org"`
}

type Error struct {
	Message string
}
