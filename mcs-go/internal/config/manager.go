package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config represents the MCS configuration
type Config struct {
	// Network settings
	HostIP       string `json:"host_ip"`
	IPMode       string `json:"ip_mode"` // localhost, auto, public, custom
	AutoDetectIP bool   `json:"auto_detect_ip"`

	// Auto-update settings
	AutoUpdateEnabled       bool   `json:"auto_update_enabled"`
	AutoUpdateCheckInterval int64  `json:"auto_update_check_interval"` // seconds
	LastUpdateCheck         int64  `json:"last_update_check"`          // unix timestamp
	LastKnownVersion        string `json:"last_known_version"`

	// Authentication
	GitHubToken string `json:"github_token,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Manager handles configuration persistence
type Manager struct {
	configPath string
	config     *Config
	mu         sync.RWMutex
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	configDir := os.Getenv("MCS_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".mcs")
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	
	m := &Manager{
		configPath: configPath,
	}

	// Load or create config
	if err := m.load(); err != nil {
		// Create default config if it doesn't exist
		m.config = m.defaultConfig()
		if err := m.save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	return m, nil
}

// defaultConfig returns the default configuration
func (m *Manager) defaultConfig() *Config {
	return &Config{
		HostIP:                  "localhost",
		IPMode:                  "localhost",
		AutoDetectIP:            false,
		AutoUpdateEnabled:       true,
		AutoUpdateCheckInterval: 86400, // 24 hours
		LastUpdateCheck:         0,
		LastKnownVersion:        "1.0.0",
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}
}

// load reads the configuration from disk
func (m *Manager) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	m.config = &config
	return nil
}

// save writes the configuration to disk
func (m *Manager) save() error {
	m.mu.RLock()
	m.config.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(m.config, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temp file first
	tempPath := m.configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, m.configPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Get returns the current configuration
func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.config
}

// GetHostIP returns the configured host IP
func (m *Manager) GetHostIP() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.HostIP
}

// SetHostIP sets the host IP
func (m *Manager) SetHostIP(ip string) error {
	m.mu.Lock()
	m.config.HostIP = ip
	m.mu.Unlock()
	return m.save()
}

// GetIPMode returns the IP mode
func (m *Manager) GetIPMode() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.IPMode
}

// SetIPMode sets the IP mode
func (m *Manager) SetIPMode(mode string) error {
	validModes := map[string]bool{
		"localhost": true,
		"auto":      true,
		"public":    true,
		"custom":    true,
	}
	
	if !validModes[mode] {
		return fmt.Errorf("invalid IP mode: %s", mode)
	}

	m.mu.Lock()
	m.config.IPMode = mode
	m.mu.Unlock()
	return m.save()
}

// IsAutoUpdateEnabled returns whether auto-update is enabled
func (m *Manager) IsAutoUpdateEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.AutoUpdateEnabled
}

// SetAutoUpdateEnabled enables or disables auto-update
func (m *Manager) SetAutoUpdateEnabled(enabled bool) error {
	m.mu.Lock()
	m.config.AutoUpdateEnabled = enabled
	m.mu.Unlock()
	return m.save()
}

// GetAutoUpdateCheckInterval returns the update check interval in seconds
func (m *Manager) GetAutoUpdateCheckInterval() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.AutoUpdateCheckInterval
}

// SetAutoUpdateCheckInterval sets the update check interval in seconds
func (m *Manager) SetAutoUpdateCheckInterval(seconds int64) error {
	if seconds < 3600 {
		return fmt.Errorf("interval must be at least 3600 seconds (1 hour)")
	}

	m.mu.Lock()
	m.config.AutoUpdateCheckInterval = seconds
	m.mu.Unlock()
	return m.save()
}

// GetLastUpdateCheck returns the last update check timestamp
func (m *Manager) GetLastUpdateCheck() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.LastUpdateCheck
}

// SetLastUpdateCheck sets the last update check timestamp
func (m *Manager) SetLastUpdateCheck(timestamp int64) error {
	m.mu.Lock()
	m.config.LastUpdateCheck = timestamp
	m.mu.Unlock()
	return m.save()
}

// GetLastKnownVersion returns the last known version
func (m *Manager) GetLastKnownVersion() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.LastKnownVersion
}

// SetLastKnownVersion sets the last known version
func (m *Manager) SetLastKnownVersion(version string) error {
	m.mu.Lock()
	m.config.LastKnownVersion = version
	m.mu.Unlock()
	return m.save()
}

// ShouldCheckForUpdate returns whether an update check is due
func (m *Manager) ShouldCheckForUpdate() bool {
	if !m.IsAutoUpdateEnabled() {
		return false
	}

	lastCheck := m.GetLastUpdateCheck()
	interval := m.GetAutoUpdateCheckInterval()
	
	if lastCheck == 0 {
		return true // Never checked
	}

	nextCheck := lastCheck + interval
	return time.Now().Unix() >= nextCheck
}

// GetGitHubToken returns the stored GitHub token
func (m *Manager) GetGitHubToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.GitHubToken
}

// SetGitHubToken stores the GitHub token
func (m *Manager) SetGitHubToken(token string) error {
	m.mu.Lock()
	m.config.GitHubToken = token
	m.mu.Unlock()
	return m.save()
}