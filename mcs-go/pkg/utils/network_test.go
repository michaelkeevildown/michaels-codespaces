package utils

import (
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkInterface_Structure(t *testing.T) {
	ni := NetworkInterface{
		Name:      "Test Interface",
		IP:        "192.168.1.100",
		IsDefault: true,
		Type:      "local",
	}
	
	assert.Equal(t, "Test Interface", ni.Name)
	assert.Equal(t, "192.168.1.100", ni.IP)
	assert.True(t, ni.IsDefault)
	assert.Equal(t, "local", ni.Type)
}

func TestGetAvailableNetworkAddresses(t *testing.T) {
	addresses, err := GetAvailableNetworkAddresses()
	require.NoError(t, err)
	require.NotEmpty(t, addresses)
	
	// Should always include localhost
	found := false
	for _, addr := range addresses {
		if addr.IP == "127.0.0.1" && addr.Type == "localhost" {
			found = true
			assert.True(t, addr.IsDefault)
			assert.Equal(t, "Localhost (this machine only)", addr.Name)
			break
		}
	}
	assert.True(t, found, "Localhost should always be included")
	
	// Should include placeholder entries for future features
	hasExternal := false
	hasDomain := false
	for _, addr := range addresses {
		if addr.Type == "external" {
			hasExternal = true
			assert.Contains(t, addr.Name, "External IP")
			assert.Empty(t, addr.IP)
		}
		if addr.Type == "domain" {
			hasDomain = true
			assert.Contains(t, addr.Name, "Custom Domain")
			assert.Empty(t, addr.IP)
		}
	}
	assert.True(t, hasExternal, "Should include external IP placeholder")
	assert.True(t, hasDomain, "Should include custom domain placeholder")
}

func TestGetAvailableNetworkAddresses_Types(t *testing.T) {
	addresses, err := GetAvailableNetworkAddresses()
	require.NoError(t, err)
	
	// Check that we have expected types
	types := make(map[string]bool)
	for _, addr := range addresses {
		types[addr.Type] = true
		
		// Validate type-specific properties
		switch addr.Type {
		case "localhost":
			assert.Equal(t, "127.0.0.1", addr.IP)
			assert.True(t, addr.IsDefault)
		case "local":
			assert.NotEmpty(t, addr.IP)
			assert.True(t, ValidateIP(addr.IP))
			assert.False(t, addr.IsDefault)
		case "external", "domain":
			assert.Empty(t, addr.IP)
			assert.False(t, addr.IsDefault)
		}
	}
	
	// Should have at least localhost and placeholders
	assert.True(t, types["localhost"])
	assert.True(t, types["external"])
	assert.True(t, types["domain"])
}

func TestGetAvailableNetworkAddresses_NameFormatting(t *testing.T) {
	addresses, err := GetAvailableNetworkAddresses()
	require.NoError(t, err)
	
	for _, addr := range addresses {
		if addr.Type == "local" && addr.IP != "" {
			// Check naming patterns for different IP ranges
			if strings.HasPrefix(addr.IP, "192.168.") {
				assert.Contains(t, addr.Name, "Local Network")
			} else if strings.HasPrefix(addr.IP, "10.") {
				assert.Contains(t, addr.Name, "Private Network")
			} else if strings.HasPrefix(addr.IP, "172.") {
				// Check if it's in the valid private range (172.16-31.x.x)
				parts := strings.Split(addr.IP, ".")
				if len(parts) >= 2 {
					var secondOctet int
					if _, err := sscanf(parts[1], "%d", &secondOctet); err == nil {
						if secondOctet >= 16 && secondOctet <= 31 {
							assert.Contains(t, addr.Name, "Private Network")
						}
					}
				}
			}
			
			// All local addresses should include the IP in parentheses
			assert.Contains(t, addr.Name, addr.IP)
		}
	}
}

func TestGetLocalNetworkIP(t *testing.T) {
	ip := GetLocalNetworkIP()
	assert.NotEmpty(t, ip)
	assert.True(t, ValidateIP(ip))
	
	// Should return a valid IP, could be localhost if no network interfaces
	parsedIP := net.ParseIP(ip)
	require.NotNil(t, parsedIP)
	
	// Should be an IPv4 address
	assert.NotNil(t, parsedIP.To4())
}

func TestGetLocalNetworkIP_Prioritization(t *testing.T) {
	ip := GetLocalNetworkIP()
	
	// Test that the function returns a consistent result
	ip2 := GetLocalNetworkIP()
	assert.Equal(t, ip, ip2, "Should return consistent results")
	
	// If we have multiple interfaces, should prioritize appropriately
	// This test is environment-dependent, so we just verify it's valid
	assert.True(t, ValidateIP(ip))
}

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Valid IPv4", "192.168.1.1", true},
		{"Valid IPv4 localhost", "127.0.0.1", true},
		{"Valid IPv4 zero", "0.0.0.0", true},
		{"Valid IPv6", "2001:db8::1", true},
		{"Valid IPv6 localhost", "::1", true},
		{"Invalid format", "192.168.1", false},
		{"Invalid octets", "192.168.300.1", false},
		{"Empty string", "", false},
		{"Non-numeric", "abc.def.ghi.jkl", false},
		{"Partial IP", "192.168.", false},
		{"Too many octets", "192.168.1.1.1", false},
		{"Negative numbers", "192.168.-1.1", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateIP(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		// Class A private (10.0.0.0/8)
		{"Private 10.0.0.1", "10.0.0.1", true},
		{"Private 10.255.255.255", "10.255.255.255", true},
		{"Private 10.123.45.67", "10.123.45.67", true},
		
		// Class B private (172.16.0.0/12)
		{"Private 172.16.0.1", "172.16.0.1", true},
		{"Private 172.31.255.255", "172.31.255.255", true},
		{"Private 172.20.1.1", "172.20.1.1", true},
		
		// Class C private (192.168.0.0/16)
		{"Private 192.168.0.1", "192.168.0.1", true},
		{"Private 192.168.255.255", "192.168.255.255", true},
		{"Private 192.168.1.100", "192.168.1.100", true},
		
		// Localhost (127.0.0.0/8)
		{"Localhost 127.0.0.1", "127.0.0.1", true},
		{"Localhost 127.255.255.255", "127.255.255.255", true},
		{"Localhost 127.1.2.3", "127.1.2.3", true},
		
		// Public IPs
		{"Public 8.8.8.8", "8.8.8.8", false},
		{"Public 1.1.1.1", "1.1.1.1", false},
		{"Public 172.15.0.1", "172.15.0.1", false}, // Just outside private range
		{"Public 172.32.0.1", "172.32.0.1", false}, // Just outside private range
		{"Public 9.255.255.255", "9.255.255.255", false}, // Just outside 10.x range
		{"Public 11.0.0.1", "11.0.0.1", false}, // Just outside 10.x range
		{"Public 192.167.1.1", "192.167.1.1", false}, // Just outside 192.168.x range
		{"Public 192.169.1.1", "192.169.1.1", false}, // Just outside 192.168.x range
		
		// Invalid IPs
		{"Invalid IP", "invalid.ip", false},
		{"Empty IP", "", false},
		{"Malformed IP", "192.168.1", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPrivateIP(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPrivateIP_EdgeCases(t *testing.T) {
	// Test boundary cases for 172.16.0.0/12 range
	boundaryTests := []struct {
		ip       string
		expected bool
	}{
		{"172.15.255.255", false}, // Just before range
		{"172.16.0.0", true},      // Start of range
		{"172.31.255.255", true},  // End of range
		{"172.32.0.0", false},     // Just after range
	}
	
	for _, tt := range boundaryTests {
		t.Run("Boundary "+tt.ip, func(t *testing.T) {
			result := IsPrivateIP(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test with mock network interfaces (integration-style test)
func TestNetworkIntegration(t *testing.T) {
	// Test the complete workflow
	addresses, err := GetAvailableNetworkAddresses()
	require.NoError(t, err)
	
	// Find a local IP
	var localIP string
	for _, addr := range addresses {
		if addr.Type == "local" && addr.IP != "" {
			localIP = addr.IP
			break
		}
	}
	
	// Test the local IP we found
	if localIP != "" {
		assert.True(t, ValidateIP(localIP))
		// Most local IPs should be private (unless running on public cloud)
		// We don't assert this since test environment varies
	}
	
	// Test getting primary local IP
	primaryIP := GetLocalNetworkIP()
	assert.True(t, ValidateIP(primaryIP))
	
	// Verify localhost is always available
	localhostFound := false
	for _, addr := range addresses {
		if addr.IP == "127.0.0.1" {
			localhostFound = true
			assert.True(t, IsPrivateIP(addr.IP))
			assert.True(t, ValidateIP(addr.IP))
			break
		}
	}
	assert.True(t, localhostFound)
}

// Benchmark tests for performance-critical functions
func BenchmarkValidateIP(b *testing.B) {
	testIPs := []string{
		"192.168.1.1",
		"invalid.ip",
		"10.0.0.1",
		"2001:db8::1",
		"",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ip := range testIPs {
			ValidateIP(ip)
		}
	}
}

func BenchmarkIsPrivateIP(b *testing.B) {
	testIPs := []string{
		"192.168.1.1",
		"8.8.8.8",
		"10.0.0.1", 
		"172.16.1.1",
		"127.0.0.1",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ip := range testIPs {
			IsPrivateIP(ip)
		}
	}
}

func BenchmarkGetLocalNetworkIP(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetLocalNetworkIP()
	}
}

// Helper function to simulate sscanf since Go doesn't have it built-in
func sscanf(s, format string, a ...interface{}) (int, error) {
	// Simple implementation for our test case
	if format == "%d" && len(a) == 1 {
		if ptr, ok := a[0].(*int); ok {
			var val int
			n, err := parseIntSimple(s, &val)
			if err == nil {
				*ptr = val
			}
			return n, err
		}
	}
	return 0, nil
}

func parseIntSimple(s string, result *int) (int, error) {
	val := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, assert.AnError
		}
		val = val*10 + int(r-'0')
	}
	*result = val
	return 1, nil
}