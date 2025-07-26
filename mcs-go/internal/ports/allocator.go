package ports

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/michaelkeevildown/mcs/pkg/utils"
)

// PortRegistry manages port allocations
type PortRegistry struct {
	mu        sync.Mutex
	file      string
	allocations map[int]Allocation
}

// Allocation represents a port allocation
type Allocation struct {
	Port       int       `json:"port"`
	Codespace  string    `json:"codespace"`
	Service    string    `json:"service"`
	AllocatedAt time.Time `json:"allocated_at"`
}

// DefaultRanges defines default port ranges for different services
var DefaultRanges = map[string]PortRange{
	"vscode": {Start: 8080, End: 8099},
	"app":    {Start: 3000, End: 3099},
	"api":    {Start: 5000, End: 5099},
	"db":     {Start: 5432, End: 5532},
}

// PortRange defines a range of ports
type PortRange struct {
	Start int
	End   int
}

// NewPortRegistry creates a new port registry
func NewPortRegistry() (*PortRegistry, error) {
	mcsDir := utils.GetMCSDir()
	portFile := filepath.Join(mcsDir, "ports.json")

	registry := &PortRegistry{
		file:        portFile,
		allocations: make(map[int]Allocation),
	}

	// Load existing allocations
	if err := registry.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load port registry: %w", err)
	}

	return registry, nil
}

// load loads allocations from disk
func (r *PortRegistry) load() error {
	data, err := os.ReadFile(r.file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &r.allocations)
}

// save saves allocations to disk
func (r *PortRegistry) save() error {
	// Ensure directory exists
	dir := filepath.Dir(r.file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(r.allocations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal allocations: %w", err)
	}

	return os.WriteFile(r.file, data, 0644)
}

// AllocatePort allocates a port for a service
func (r *PortRegistry) AllocatePort(codespace, service string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get port range for service
	portRange, ok := DefaultRanges[service]
	if !ok {
		portRange = PortRange{Start: 10000, End: 20000}
	}

	// Find an available port
	port, err := r.findAvailablePort(portRange)
	if err != nil {
		return 0, err
	}

	// Record allocation
	r.allocations[port] = Allocation{
		Port:        port,
		Codespace:   codespace,
		Service:     service,
		AllocatedAt: time.Now(),
	}

	// Save to disk
	if err := r.save(); err != nil {
		delete(r.allocations, port)
		return 0, fmt.Errorf("failed to save allocation: %w", err)
	}

	return port, nil
}

// findAvailablePort finds an available port in the given range
func (r *PortRegistry) findAvailablePort(portRange PortRange) (int, error) {
	// Create a random starting point
	rand.Seed(time.Now().UnixNano())
	start := portRange.Start + rand.Intn(portRange.End-portRange.Start)

	// Try ports in order from random start
	for i := 0; i <= portRange.End-portRange.Start; i++ {
		port := portRange.Start + ((start-portRange.Start+i) % (portRange.End-portRange.Start+1))
		
		// Check if port is already allocated
		if _, allocated := r.allocations[port]; allocated {
			continue
		}

		// Check if port is available on system
		if isPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d", portRange.Start, portRange.End)
}

// isPortAvailable checks if a port is available on the system
func isPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// ReleasePort releases a port allocation
func (r *PortRegistry) ReleasePort(port int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.allocations, port)
	return r.save()
}

// ReleaseCodespacePorts releases all ports for a codespace
func (r *PortRegistry) ReleaseCodespacePorts(codespace string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find all ports for this codespace
	var toRelease []int
	for port, alloc := range r.allocations {
		if alloc.Codespace == codespace {
			toRelease = append(toRelease, port)
		}
	}

	// Release them
	for _, port := range toRelease {
		delete(r.allocations, port)
	}

	return r.save()
}

// GetCodespacePorts returns all ports allocated to a codespace
func (r *PortRegistry) GetCodespacePorts(codespace string) []Allocation {
	r.mu.Lock()
	defer r.mu.Unlock()

	var ports []Allocation
	for _, alloc := range r.allocations {
		if alloc.Codespace == codespace {
			ports = append(ports, alloc)
		}
	}

	return ports
}

// AllocateCodespacePorts allocates standard ports for a codespace
func (r *PortRegistry) AllocateCodespacePorts(codespace string) (map[string]int, error) {
	ports := make(map[string]int)

	// Allocate VS Code port
	vsCodePort, err := r.AllocatePort(codespace, "vscode")
	if err != nil {
		return nil, fmt.Errorf("failed to allocate VS Code port: %w", err)
	}
	ports["vscode"] = vsCodePort

	// Allocate app port
	appPort, err := r.AllocatePort(codespace, "app")
	if err != nil {
		// Rollback VS Code port
		r.ReleasePort(vsCodePort)
		return nil, fmt.Errorf("failed to allocate app port: %w", err)
	}
	ports["app"] = appPort

	return ports, nil
}

// GetAllAllocations returns all current port allocations
func (r *PortRegistry) GetAllAllocations() map[int]Allocation {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Return a copy
	result := make(map[int]Allocation)
	for k, v := range r.allocations {
		result[k] = v
	}

	return result
}