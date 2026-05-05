package utils

import (
	"encoding/json"
	"time"
)

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

//	type Home struct {
//		Status      string `json:"status"`
//		PublicIP    string `json:"publicip"`
//		ISPInfo     string `json:"ispinfo"`
//		Interstatus string `json:"interstatus"`
//	}
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

// ---------------------These from tcpcheck-----------------
type TCPCheckRequest struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Timeout int    `json:"timeout"` // seconds
}

type TCPCheckResponse struct {
	Success      bool   `json:"success"`
	Status       string `json:"status"` // open, closed, filtered, error
	Host         string `json:"host"`
	Port         int    `json:"port"`
	ResponseTime int64  `json:"responseTime"` // ms
	Message      string `json:"message"`
}

//-----------------ending Tcpcheck-------------------------

// -----------------DNS check-------------------------------
type Request struct {
	Domain string `json:"domain"`
	Type   string `json:"type"`
	Server string `json:"server,omitempty"`
}

type RRResponse struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	TTL   uint32 `json:"ttl"`
	Value string `json:"value"`
}

type APIResponse struct {
	Domain   string       `json:"domain"`
	Type     string       `json:"type"`
	Answers  []RRResponse `json:"answers"`
	Status   string       `json:"status"`
	Server   string       `json:"server,omitempty"`
	Duration string       `json:"duration,omitempty"`
}

//----------------------DNS end-------------------------------

//----------------------Traceroute---------------------------

type TracerouteRequest struct {
	Target       string `json:"target"`
	MaxHops      int    `json:"maxHops"`
	ProbesPerHop int    `json:"probesPerHop"`
	Timeout      int    `json:"timeout"`
	Protocol     string `json:"protocol"`
}

type TracerouteMessage struct {
	Type     string     `json:"type"`
	Hop      *HopResult `json:"hop,omitempty"`
	TargetIP string     `json:"ip,omitempty"`
	Progress *Progress  `json:"progress,omitempty"`
	Complete *Complete  `json:"complete,omitempty"`
	Error    string     `json:"message,omitempty"`
	Status   string     `json:"status,omitempty"`
}

type HopResult struct {
	Hop        int     `json:"hop"`
	IP         string  `json:"ip,omitempty"`
	Hostname   string  `json:"hostname,omitempty"`
	RTT        float64 `json:"rtt,omitempty"`
	IsTarget   bool    `json:"isTarget,omitempty"`
	Status     string  `json:"status,omitempty"`
	ASN        string  `json:"asn,omitempty"`
	Country    string  `json:"country,omitempty"`
	Location   string  `json:"location,omitempty"`
	ISP        string  `json:"isp,omitempty"`
	ReverseDNS string  `json:"reverseDns,omitempty"`
}

type Progress struct {
	CurrentHop int `json:"currentHop"`
	TotalHops  int `json:"totalHops"`
}

type Complete struct {
	Reached   bool        `json:"reached"`
	TargetIP  string      `json:"targetIp"`
	Hops      []HopResult `json:"hops"`
	TotalTime float64     `json:"totalTime"`
}

//---------------------------Tracerouteend------------------------------

//---------------------------ICMP------------------------------

type PingRequest struct {
	Target   string  `json:"target"`
	Count    int     `json:"count"`
	Size     int     `json:"size"`
	Timeout  float64 `json:"timeout"`
	Interval float64 `json:"interval"`
}

type PingMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

//-------------------------------------------------------------
//---------------------------HTTPS-----------------------------

type HTTPSCheckRequest struct {
	URL              string `json:"url"`
	Timeout          int    `json:"timeout"`
	CheckCertificate bool   `json:"checkCertificate"`
	CheckRedirects   bool   `json:"checkRedirects"`
}

type CertificateInfo struct {
	Subject       string `json:"subject"`
	Issuer        string `json:"issuer"`
	ValidFrom     string `json:"validFrom"`
	ValidUntil    string `json:"validUntil"`
	DaysRemaining int    `json:"daysRemaining"`
}

type HTTPSCheckResponse struct {
	HTTPSSupported bool             `json:"httpsSupported"`
	StatusCode     int              `json:"statusCode"`
	TLSVersion     string           `json:"tlsVersion,omitempty"`
	Cipher         string           `json:"cipher,omitempty"`
	HSTSEnabled    bool             `json:"hstsEnabled"`
	Certificate    *CertificateInfo `json:"certificate,omitempty"`
	ResponseTime   int64            `json:"responseTime"`
	Error          string           `json:"error,omitempty"`
}

// ----------------------------------------------------------
type Error struct {
	Message string
}

// ------------------------VLAN---------------------------------
type VLANNetwork struct {
	ID                  int       `json:"id"`
	VLANId              int       `json:"vlan_id"`
	VLANName            string    `json:"vlan_name"`
	NetworkMode         string    `json:"network_mode"`
	IPAddress           *string   `json:"ip_address,omitempty"`
	CIDRNotation        *string   `json:"cidr_notation,omitempty"`
	CIDRFull            *string   `json:"cidr_full,omitempty"`
	DefaultGateway      *string   `json:"default_gateway,omitempty"`
	MonitoringEnabled   bool      `json:"monitoring_enabled"`
	ScanIntervalSeconds int       `json:"scan_interval_seconds"`
	CreatedAt           time.Time `json:"created_at"`
	InterfaceName       string    `json:"interface_name"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type DiscoveredDevice struct {
	ID           int       `json:"id"`
	NetworkId    int       `json:"network_id"`
	IPAddress    string    `json:"ip_address"`
	MACAddress   string    `json:"mac_address"`
	Hostname     string    `json:"hostname,omitempty"`
	Vendor       string    `json:"vendor,omitempty"`
	DeviceStatus string    `json:"device_status"`
	FirstSeen    time.Time `json:"first_seen"`
	LastSeen     time.Time `json:"last_seen"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DeviceNotification struct {
	EventType  string `json:"event_type"`
	NetworkId  int    `json:"network_id"`
	IPAddress  string `json:"ip_address"`
	MACAddress string `json:"mac_address"`
	Hostname   string `json:"hostname"`
	Vendor     string `json:"vendor"`
	OldStatus  string `json:"old_status,omitempty"`
	NewStatus  string `json:"new_status,omitempty"`
	Status     string `json:"status,omitempty"`
	Severity   string `json:"severity"`
	Message    string `json:"message,omitempty"`
	LastSeen   string `json:"last_seen,omitempty"`
	FirstSeen  string `json:"first_seen,omitempty"`
}

// last try fix
//
//	type DeviceNotification struct {
//		EventType   string `json:"event_type"`
//		VLANId      int    `json:"vlan_id"`
//		IPAddress   string `json:"ip_address"`
//		MACAddress  string `json:"mac_address,omitempty"`
//		Hostname    string `json:"hostname,omitempty"`
//		Vendor      string `json:"vendor,omitempty"`
//		Status      string `json:"status,omitempty"`
//		OldStatus   string `json:"old_status,omitempty"`
//		NewStatus   string `json:"new_status,omitempty"`
//		FirstSeen   string `json:"first_seen,omitempty"`
//		LastSeen    string `json:"last_seen,omitempty"`
//	}

type CustomTime struct {
	time.Time
}

// UnmarshalJSON handles both RFC3339 and PostgreSQL timestamp formats
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	// Remove quotes
	s = s[1 : len(s)-1]

	// Try different formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999999", // PostgreSQL timestamp
		"2006-01-02T15:04:05",        // PostgreSQL timestamp without microseconds
		"2006-01-02 15:04:05.999999", // PostgreSQL timestamp with space
		"2006-01-02 15:04:05",        // PostgreSQL timestamp with space, no microseconds
	}

	var err error
	for _, format := range formats {
		ct.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}

	return err
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.Time.Format(time.RFC3339))
}

type ScanLog struct {
	ID              int        `json:"id"`
	NetworkId       int        `json:"network_id"`
	ScanStartedAt   time.Time  `json:"scan_started_at"`
	ScanCompletedAt *time.Time `json:"scan_completed_at,omitempty"`
	TotalIPsScanned int        `json:"total_ips_scanned"`
	DevicesFound    int        `json:"devices_found"`
	DevicesNew      int        `json:"devices_new"`
	DevicesOffline  int        `json:"devices_offline"`
	ScanDurationMs  int        `json:"scan_duration_ms"`
	ScanStatus      string     `json:"scan_status"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}
type MACVendor struct {
	ID             int       `json:"id"`
	OUI            string    `json:"oui"`
	VendorName     string    `json:"vendor_name"`
	FetchedFromAPI bool      `json:"fetched_from_api"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastSeen       time.Time `json:"last_seen"`
}

//----------------------------------VLAN END-----------------------------------

//---------------------------------TCP monitoring----------------------------------

type PortMonitor struct {
	ID                     int        `json:"id"`
	FriendlyName           string     `json:"friendly_name"`
	Hostname               string     `json:"hostname"`
	Port                   int        `json:"port"`
	HeartbeatInterval      int        `json:"heartbeat_interval"`
	Retries                int        `json:"retries"`
	HeartbeatRetryInterval int        `json:"heartbeat_retry_interval"`
	Status                 string     `json:"status"`
	LastTCPStatus          *string    `json:"last_tcp_status,omitempty"`
	LastCheckedAt          *time.Time `json:"last_checked_at,omitempty"`
	LastResponseMs         *int       `json:"last_response_ms,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type PortMonitorLog struct {
	ID           int       `json:"id"`
	MonitorID    int       `json:"monitor_id"`
	Status       string    `json:"status"`
	ResponseMs   *int      `json:"response_ms,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CheckedAt    time.Time `json:"checked_at"`
}

type CreatePortMonitorRequest struct {
	FriendlyName           string `json:"friendly_name"`
	Hostname               string `json:"hostname"`
	Port                   int    `json:"port"`
	HeartbeatInterval      int    `json:"heartbeat_interval"`
	Retries                int    `json:"retries"`
	HeartbeatRetryInterval int    `json:"heartbeat_retry_interval"`
}

//-------------------------snmp

type SNMPMonitor struct {
	ID                int        `json:"id"`
	FriendlyName      string     `json:"friendly_name"`
	Hostname          string     `json:"hostname"`
	Port              int        `json:"port"`
	CommunityString   string     `json:"community_string"`
	OID               string     `json:"oid"`
	SNMPVersion       string     `json:"snmp_version"`
	PollingInterval   int        `json:"polling_interval"`
	Timeout           int        `json:"timeout"`
	Retries           int        `json:"retries"`
	ExpectedValueType string     `json:"expected_value_type"`
	Status            string     `json:"status"`
	LastValue         *string    `json:"last_value,omitempty"`
	LastCheckedAt     *time.Time `json:"last_checked_at,omitempty"`
	LastResponseMs    *int       `json:"last_response_ms,omitempty"`
	ErrorMessage      *string    `json:"error_message,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type SNMPMonitorLog struct {
	ID           int       `json:"id"`
	MonitorID    int       `json:"monitor_id"`
	Status       string    `json:"status"`
	Value        *string   `json:"value,omitempty"`
	ResponseMs   *int      `json:"response_ms,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CheckedAt    time.Time `json:"checked_at"`
}

type CreateSNMPMonitorRequest struct {
	FriendlyName      string `json:"friendly_name"`
	Hostname          string `json:"hostname"`
	Port              int    `json:"port"`
	CommunityString   string `json:"community_string"`
	OID               string `json:"oid"`
	SNMPVersion       string `json:"snmp_version"`
	PollingInterval   int    `json:"polling_interval"`
	Timeout           int    `json:"timeout"`
	Retries           int    `json:"retries"`
	ExpectedValueType string `json:"expected_value_type"`
}

// ----------------------ping

type PingMonitor struct {
	ID               int        `json:"id"`
	FriendlyName     string     `json:"friendly_name"`
	Hostname         string     `json:"hostname"`
	CheckInterval    int        `json:"check_interval"`
	LatencyThreshold int        `json:"latency_threshold"`
	Timeout          int        `json:"timeout"`
	Status           string     `json:"status"`
	LastLatencyMs    *int       `json:"last_latency_ms,omitempty"`
	LastCheckedAt    *time.Time `json:"last_checked_at,omitempty"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type PingMonitorLog struct {
	ID           int       `json:"id"`
	MonitorID    int       `json:"monitor_id"`
	Status       string    `json:"status"`
	LatencyMs    *int      `json:"latency_ms,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	CheckedAt    time.Time `json:"checked_at"`
}

type CreatePingMonitorRequest struct {
	FriendlyName     string `json:"friendly_name"`
	Hostname         string `json:"hostname"`
	CheckInterval    int    `json:"check_interval"`
	LatencyThreshold int    `json:"latency_threshold"`
	Timeout          int    `json:"timeout"`
}

// -------------------ping end

// --------------certificate monitor
type SSLMonitor struct {
	ID            int        `json:"id"`
	Domain        string     `json:"domain"`
	FriendlyName  string     `json:"friendly_name,omitempty"`
	Port          int        `json:"port"`
	CheckInterval int        `json:"check_interval"`
	WarningDays   int        `json:"warning_days"`
	CriticalDays  int        `json:"critical_days"`
	Status        string     `json:"status"`
	Issuer        *string    `json:"issuer,omitempty"`
	Subject       *string    `json:"subject,omitempty"`
	ValidFrom     *time.Time `json:"valid_from,omitempty"`
	ValidUntil    *time.Time `json:"valid_until,omitempty"`
	DaysRemaining *int       `json:"days_remaining,omitempty"`
	LastCheckedAt *time.Time `json:"last_checked_at,omitempty"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type SSLMonitorLog struct {
	ID            int        `json:"id"`
	MonitorID     int        `json:"monitor_id"`
	Status        string     `json:"status"`
	Issuer        *string    `json:"issuer,omitempty"`
	ValidUntil    *time.Time `json:"valid_until,omitempty"`
	DaysRemaining *int       `json:"days_remaining,omitempty"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	CheckedAt     time.Time  `json:"checked_at"`
}

type CreateSSLMonitorRequest struct {
	Domain        string `json:"domain"`
	FriendlyName  string `json:"friendly_name"`
	Port          int    `json:"port"`
	CheckInterval int    `json:"check_interval"`
	WarningDays   int    `json:"warning_days"`
	CriticalDays  int    `json:"critical_days"`
}

// --------------certificate monitor end------------------------
// --------------domain expiry----------------------------------
type DomainExpiryMonitor struct {
	ID            int        `json:"id"`
	Domain        string     `json:"domain"`
	FriendlyName  string     `json:"friendly_name,omitempty"`
	CheckInterval int        `json:"check_interval"`
	WarningDays   int        `json:"warning_days"`
	CriticalDays  int        `json:"critical_days"`
	Status        string     `json:"status"`
	Registrar     *string    `json:"registrar,omitempty"`
	Registrant    *string    `json:"registrant,omitempty"`
	RegisteredOn  *time.Time `json:"registered_on,omitempty"`
	ExpiresOn     *time.Time `json:"expires_on,omitempty"`
	UpdatedOn     *time.Time `json:"updated_on,omitempty"`
	DaysRemaining *int       `json:"days_remaining,omitempty"`
	NameServers   []string   `json:"name_servers,omitempty"`
	LastCheckedAt *time.Time `json:"last_checked_at,omitempty"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type DomainExpiryLog struct {
	ID            int        `json:"id"`
	MonitorID     int        `json:"monitor_id"`
	Status        string     `json:"status"`
	Registrar     *string    `json:"registrar,omitempty"`
	ExpiresOn     *time.Time `json:"expires_on,omitempty"`
	DaysRemaining *int       `json:"days_remaining,omitempty"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	CheckedAt     time.Time  `json:"checked_at"`
}

type CreateDomainExpiryRequest struct {
	Domain        string `json:"domain"`
	FriendlyName  string `json:"friendly_name"`
	CheckInterval int    `json:"check_interval"`
	WarningDays   int    `json:"warning_days"`
	CriticalDays  int    `json:"critical_days"`
}

// whoisResult
type WhoisResult struct {
	Registrar    string
	Registrant   string
	RegisteredOn *time.Time
	ExpiresOn    *time.Time
	UpdatedOn    *time.Time
	NameServers  []string
}

//---------------------------------------------------------------------
