package vlan

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

// DetectedInterface represents an auto-detected network interface
type DetectedInterface struct {
	Name           string `json:"interface"`
	IPv4           string `json:"ipv4"`
	MAC            string `json:"mac"`
	CIDR           int    `json:"cidr"`
	DefaultGateway string `json:"default_gateway"`
	IsVLAN         bool   `json:"is_vlan"`
	ParentIface    string `json:"parent_interface,omitempty"`
	VLANId         *int   `json:"vlan_id,omitempty"`
}

// InterfaceDetector scans system interfaces
type InterfaceDetector struct{}

func NewInterfaceDetector() *InterfaceDetector {
	return &InterfaceDetector{}
}

// GetAllInterfaces returns all interfaces with default gateways
func (d *InterfaceDetector) GetAllInterfaces() ([]DetectedInterface, error) {
	// Get default routes
	defaultRoutes := d.getDefaultRoutes()

	if len(defaultRoutes) == 0 {
		log.Printf("[DETECTOR] ⚠️  No default routes found")
		return []DetectedInterface{}, nil
	}

	log.Printf("[DETECTOR] Found %d default route(s)", len(defaultRoutes))

	// Get all interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	var results []DetectedInterface

	for _, iface := range ifaces {
		//  Skip loopback interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		//  Skip interfaces that are down
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Get gateway for this interface
		gateway, hasGateway := defaultRoutes[iface.Name]
		if !hasGateway {
			// Skip interfaces without default gateway
			continue
		}

		// Get addresses
		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("[DETECTOR] Failed to get addresses for %s: %v", iface.Name, err)
			continue
		}

		if len(addrs) == 0 {
			log.Printf("[DETECTOR] No addresses found for %s", iface.Name)
			continue
		}

		// Find first IPv4 address
		var foundIPv4 bool
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			// Skip non-IPv4
			if ipNet.IP.To4() == nil {
				continue
			}

			ones, _ := ipNet.Mask.Size()
			mac := iface.HardwareAddr.String()
			if mac == "" {
				mac = "00:00:00:00:00:00"
			}

			detected := DetectedInterface{
				Name:           iface.Name,
				IPv4:           ipNet.IP.String(),
				MAC:            mac,
				CIDR:           ones,
				DefaultGateway: gateway,
				IsVLAN:         false,
			}

			// Check if it's a VLAN interface
			if d.isVLANInterface(iface.Name) {
				detected.IsVLAN = true
				parent, vlanID := d.parseVLANInterface(iface.Name)
				detected.ParentIface = parent
				if vlanID > 0 {
					detected.VLANId = &vlanID
				}
			}

			results = append(results, detected)
			foundIPv4 = true
			break // Only first IPv4 per interface
		}

		if !foundIPv4 {
			log.Printf("DETECTOR No IPv4 address found for %s", iface.Name)
		}
	}

	if len(results) == 0 {
		log.Printf("DETECTOR  No interfaces with IPv4 and default gateway found")
	} else {
		log.Printf("DETECTOR Found %d interface(s) with default gateway", len(results))
	}

	return results, nil
}

// getDefaultRoutes parses `ip route show` for default gateways
func (d *InterfaceDetector) getDefaultRoutes() map[string]string {
	routes := make(map[string]string)

	out, err := exec.Command("ip", "route", "show").Output()
	if err != nil {
		log.Printf("[DETECTOR] Failed to get routes: %v", err)
		return routes
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		// Format: default via <gateway> dev <iface> [metric <N>]
		if len(fields) >= 5 && fields[0] == "default" && fields[1] == "via" {
			gateway := fields[2]
			for i, f := range fields {
				if f == "dev" && i+1 < len(fields) {
					iface := fields[i+1]
					// Only add if not exists (keep first/best metric)
					if _, exists := routes[iface]; !exists {
						routes[iface] = gateway
						log.Printf("[DETECTOR] Route: %s -> %s", iface, gateway)
					}
					break
				}
			}
		}
	}

	return routes
}

// isVLANInterface checks if interface name suggests it's a VLAN
func (d *InterfaceDetector) isVLANInterface(name string) bool {
	// Check for pattern: <parent>.<number>
	// Examples: eth0.20, enp0s3.10, wlan0.100
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return false
	}

	// Check if second part is a valid VLAN ID (numeric, 1-4094)
	var vlanID int
	n, err := fmt.Sscanf(parts[1], "%d", &vlanID)
	return err == nil && n == 1 && vlanID > 0 && vlanID <= 4094
}

// parseVLANInterface extracts parent and VLAN ID
// eth0.20 -> ("eth0", 20)
func (d *InterfaceDetector) parseVLANInterface(name string) (string, int) {
	parts := strings.Split(name, ".")
	if len(parts) != 2 {
		return name, 0
	}

	var vlanID int
	n, err := fmt.Sscanf(parts[1], "%d", &vlanID)
	if err != nil || n != 1 || vlanID <= 0 || vlanID > 4094 {
		// Invalid VLAN ID (valid range: 1-4094)
		return name, 0
	}

	return parts[0], vlanID
}
