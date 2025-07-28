package utils

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

// NetworkInterface represents a network interface with its addresses
type NetworkInterface struct {
	Name      string
	IP        string
	IsDefault bool
	Type      string // "localhost", "local", "external", "domain"
}

// GetAvailableNetworkAddresses returns all available network addresses
func GetAvailableNetworkAddresses() ([]NetworkInterface, error) {
	var addresses []NetworkInterface

	// Always include localhost
	addresses = append(addresses, NetworkInterface{
		Name:      "Localhost (this machine only)",
		IP:        "127.0.0.1",
		IsDefault: true,
		Type:      "localhost",
	})

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return addresses, err
	}

	// Find all non-loopback addresses
	for _, iface := range interfaces {
		// Skip down interfaces
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Skip loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip if not IPv4
			if ip == nil || ip.To4() == nil {
				continue
			}

			// Skip link-local addresses
			if ip.IsLinkLocalUnicast() {
				continue
			}

			// Determine if this is a common local network
			ipStr := ip.String()
			name := fmt.Sprintf("%s (%s)", iface.Name, ipStr)
			
			// Make the name more user-friendly
			if strings.HasPrefix(ipStr, "192.168.") {
				name = fmt.Sprintf("Local Network - %s (%s)", iface.Name, ipStr)
			} else if strings.HasPrefix(ipStr, "10.") {
				name = fmt.Sprintf("Private Network - %s (%s)", iface.Name, ipStr)
			} else if strings.HasPrefix(ipStr, "172.") {
				// Check if it's in the 172.16.0.0/12 range
				parts := strings.Split(ipStr, ".")
				if len(parts) >= 2 {
					secondOctet := 0
					fmt.Sscanf(parts[1], "%d", &secondOctet)
					if secondOctet >= 16 && secondOctet <= 31 {
						name = fmt.Sprintf("Private Network - %s (%s)", iface.Name, ipStr)
					}
				}
			}

			addresses = append(addresses, NetworkInterface{
				Name: name,
				IP:   ipStr,
				Type: "local",
			})
		}
	}

	// Add placeholders for future features
	addresses = append(addresses, NetworkInterface{
		Name: "External IP (coming soon)",
		IP:   "",
		Type: "external",
	})

	addresses = append(addresses, NetworkInterface{
		Name: "Custom Domain with Wildcard Cert (coming soon)",
		IP:   "",
		Type: "domain",
	})

	return addresses, nil
}

// GetLocalNetworkIP returns the primary local network IP address
func GetLocalNetworkIP() string {
	// Get all network interfaces
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	// Sort addresses to prioritize common local networks
	var candidates []string
	
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipStr := ipnet.IP.String()
				candidates = append(candidates, ipStr)
			}
		}
	}

	// Sort to prioritize common local network ranges
	sort.Slice(candidates, func(i, j int) bool {
		// Prioritize 192.168.x.x
		if strings.HasPrefix(candidates[i], "192.168.") && !strings.HasPrefix(candidates[j], "192.168.") {
			return true
		}
		if !strings.HasPrefix(candidates[i], "192.168.") && strings.HasPrefix(candidates[j], "192.168.") {
			return false
		}
		
		// Then 10.x.x.x
		if strings.HasPrefix(candidates[i], "10.") && !strings.HasPrefix(candidates[j], "10.") {
			return true
		}
		if !strings.HasPrefix(candidates[i], "10.") && strings.HasPrefix(candidates[j], "10.") {
			return false
		}
		
		// Then 172.16-31.x.x
		return candidates[i] < candidates[j]
	})

	if len(candidates) > 0 {
		return candidates[0]
	}

	return "127.0.0.1"
}

// ValidateIP checks if a string is a valid IP address
func ValidateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsPrivateIP checks if an IP is in a private range
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}