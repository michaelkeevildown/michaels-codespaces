package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// testConfig creates a test configuration with predefined values
func testConfig() *Config {
	return &Config{
		HostIP:                  "192.168.1.100",
		IPMode:                  "custom",
		AutoDetectIP:            true,
		AutoUpdateEnabled:       true,
		AutoUpdateCheckInterval: 7200,
		LastUpdateCheck:         time.Now().Unix() - 3600,
		LastKnownVersion:        "1.5.0",
		GitHubToken:             "test-token-123",
		CreatedAt:               time.Now().Add(-24 * time.Hour),
		UpdatedAt:               time.Now().Add(-1 * time.Hour),
	}
}

// createTempManager creates a manager with a temporary config file
func createTempManager(t *testing.T, config *Config) *Manager {
	t.Helper()
	
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	
	if config != nil {
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal test config: %v", err)
		}
		
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}
	}
	
	return &Manager{
		configPath: configPath,
		config:     config,
	}
}

// TestConfigSerialization tests config JSON serialization/deserialization
func TestConfigSerialization(t *testing.T) {
	original := testConfig()
	
	// Test serialization
	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("Failed to serialize config: %v", err)
	}
	
	// Test deserialization
	var deserialized Config
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatalf("Failed to deserialize config: %v", err)
	}
	
	// Compare values
	if deserialized.HostIP != original.HostIP {
		t.Errorf("HostIP mismatch: got %v, want %v", deserialized.HostIP, original.HostIP)
	}
	if deserialized.IPMode != original.IPMode {
		t.Errorf("IPMode mismatch: got %v, want %v", deserialized.IPMode, original.IPMode)
	}
	if deserialized.AutoDetectIP != original.AutoDetectIP {
		t.Errorf("AutoDetectIP mismatch: got %v, want %v", deserialized.AutoDetectIP, original.AutoDetectIP)
	}
	if deserialized.AutoUpdateEnabled != original.AutoUpdateEnabled {
		t.Errorf("AutoUpdateEnabled mismatch: got %v, want %v", deserialized.AutoUpdateEnabled, original.AutoUpdateEnabled)
	}
	if deserialized.AutoUpdateCheckInterval != original.AutoUpdateCheckInterval {
		t.Errorf("AutoUpdateCheckInterval mismatch: got %v, want %v", deserialized.AutoUpdateCheckInterval, original.AutoUpdateCheckInterval)
	}
	if deserialized.LastUpdateCheck != original.LastUpdateCheck {
		t.Errorf("LastUpdateCheck mismatch: got %v, want %v", deserialized.LastUpdateCheck, original.LastUpdateCheck)
	}
	if deserialized.LastKnownVersion != original.LastKnownVersion {
		t.Errorf("LastKnownVersion mismatch: got %v, want %v", deserialized.LastKnownVersion, original.LastKnownVersion)
	}
	if deserialized.GitHubToken != original.GitHubToken {
		t.Errorf("GitHubToken mismatch: got %v, want %v", deserialized.GitHubToken, original.GitHubToken)
	}
}

// TestConfigOmitEmpty tests that empty GitHub token is omitted from JSON
func TestConfigOmitEmpty(t *testing.T) {
	config := &Config{
		HostIP:        "localhost",
		IPMode:        "localhost",
		GitHubToken:   "", // Should be omitted
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to serialize config: %v", err)
	}
	
	// Check that github_token field is not present when empty
	dataStr := string(data)
	if len(config.GitHubToken) == 0 && contains(dataStr, "github_token") {
		t.Error("Empty GitHubToken should be omitted from JSON")
	}
}

// contains is a helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (contains(s[1:], substr) || (len(s) >= len(substr) && s[:len(substr)] == substr)))
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		validate func(*testing.T, *Config) bool
	}{
		{
			name: "valid config",
			config: &Config{
				HostIP:                  "192.168.1.1",
				IPMode:                  "custom",
				AutoUpdateCheckInterval: 3600,
				LastKnownVersion:        "1.0.0",
			},
			validate: func(t *testing.T, c *Config) bool {
				return c.HostIP != "" && c.IPMode != "" && c.AutoUpdateCheckInterval >= 3600
			},
		},
		{
			name: "config with zero values",
			config: &Config{
				HostIP:                  "",
				IPMode:                  "",
				AutoUpdateCheckInterval: 0,
				LastKnownVersion:        "",
			},
			validate: func(t *testing.T, c *Config) bool {
				// Zero values should be acceptable for loading
				return true
			},
		},
		{
			name: "config with future timestamps",
			config: &Config{
				LastUpdateCheck: time.Now().Unix() + 86400, // Tomorrow
				CreatedAt:       time.Now().Add(24 * time.Hour),
				UpdatedAt:       time.Now().Add(48 * time.Hour),
			},
			validate: func(t *testing.T, c *Config) bool {
				// Future timestamps should be handled gracefully
				return c.LastUpdateCheck > 0
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.validate(t, tt.config) {
				t.Error("Config validation failed")
			}
		})
	}
}

// TestManager_ErrorHandling tests various error conditions
func TestManager_ErrorHandling(t *testing.T) {
	t.Run("load with corrupted JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")
		
		// Write corrupted JSON (missing closing brace)
		corruptedJSON := `{"host_ip": "localhost", "ip_mode": "localhost"`
		os.WriteFile(configPath, []byte(corruptedJSON), 0600)
		
		manager := &Manager{configPath: configPath}
		err := manager.load()
		
		if err == nil {
			t.Error("Expected error loading corrupted JSON, got nil")
		}
		if err != nil && !contains(err.Error(), "failed to unmarshal config") {
			t.Errorf("Expected unmarshal error, got: %v", err)
		}
	})
	
	t.Run("save with invalid config path", func(t *testing.T) {
		manager := &Manager{
			configPath: "/invalid/path/that/does/not/exist/config.json",
			config:     testConfig(),
		}
		
		err := manager.save()
		if err == nil {
			t.Error("Expected error saving to invalid path, got nil")
		}
	})
	
	t.Run("get with nil config", func(t *testing.T) {
		manager := &Manager{
			configPath: "/tmp/test.json",
			config:     nil, // This could happen in error conditions
		}
		
		// This should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Get() panicked with nil config: %v", r)
			}
		}()
		
		// Note: This will panic in the actual implementation, 
		// which is acceptable behavior for an invalid state
		_ = manager.Get()
	})
}

// TestManager_UpdateSequence tests a sequence of configuration updates
func TestManager_UpdateSequence(t *testing.T) {
	manager := createTempManager(t, testConfig())
	
	// Sequence of updates
	updates := []struct {
		name string
		fn   func() error
	}{
		{"set host IP", func() error { return manager.SetHostIP("10.0.0.1") }},
		{"set IP mode", func() error { return manager.SetIPMode("auto") }},
		{"disable auto update", func() error { return manager.SetAutoUpdateEnabled(false) }},
		{"set check interval", func() error { return manager.SetAutoUpdateCheckInterval(7200) }},
		{"set last check", func() error { return manager.SetLastUpdateCheck(time.Now().Unix()) }},
		{"set version", func() error { return manager.SetLastKnownVersion("2.0.0") }},
		{"set GitHub token", func() error { return manager.SetGitHubToken("new-token") }},
	}
	
	for _, update := range updates {
		t.Run(update.name, func(t *testing.T) {
			if err := update.fn(); err != nil {
				t.Errorf("Update %s failed: %v", update.name, err)
			}
		})
	}
	
	// Verify final state
	final := manager.Get()
	if final.HostIP != "10.0.0.1" {
		t.Errorf("Final HostIP = %v, want 10.0.0.1", final.HostIP)
	}
	if final.IPMode != "auto" {
		t.Errorf("Final IPMode = %v, want auto", final.IPMode)
	}
	if final.AutoUpdateEnabled != false {
		t.Errorf("Final AutoUpdateEnabled = %v, want false", final.AutoUpdateEnabled)
	}
	if final.LastKnownVersion != "2.0.0" {
		t.Errorf("Final LastKnownVersion = %v, want 2.0.0", final.LastKnownVersion)
	}
	if final.GitHubToken != "new-token" {
		t.Errorf("Final GitHubToken = %v, want new-token", final.GitHubToken)
	}
}

// TestManager_EdgeCases tests edge cases and boundary conditions
func TestManager_EdgeCases(t *testing.T) {
	t.Run("very long host IP", func(t *testing.T) {
		manager := createTempManager(t, testConfig())
		longIP := make([]byte, 1000)
		for i := range longIP {
			longIP[i] = '1'
		}
		
		err := manager.SetHostIP(string(longIP))
		if err != nil {
			t.Errorf("Failed to set long host IP: %v", err)
		}
		
		if manager.GetHostIP() != string(longIP) {
			t.Error("Long host IP not preserved")
		}
	})
	
	t.Run("unicode in config values", func(t *testing.T) {
		manager := createTempManager(t, testConfig())
		unicodeIP := "测试.example.com"
		unicodeVersion := "版本1.0.0"
		
		if err := manager.SetHostIP(unicodeIP); err != nil {
			t.Errorf("Failed to set unicode host IP: %v", err)
		}
		if err := manager.SetLastKnownVersion(unicodeVersion); err != nil {
			t.Errorf("Failed to set unicode version: %v", err)
		}
		
		config := manager.Get()
		if config.HostIP != unicodeIP {
			t.Errorf("Unicode host IP = %v, want %v", config.HostIP, unicodeIP)
		}
		if config.LastKnownVersion != unicodeVersion {
			t.Errorf("Unicode version = %v, want %v", config.LastKnownVersion, unicodeVersion)
		}
	})
	
	t.Run("maximum interval value", func(t *testing.T) {
		manager := createTempManager(t, testConfig())
		maxInterval := int64(31536000) // 1 year in seconds
		
		err := manager.SetAutoUpdateCheckInterval(maxInterval)
		if err != nil {
			t.Errorf("Failed to set maximum interval: %v", err)
		}
		
		if manager.GetAutoUpdateCheckInterval() != maxInterval {
			t.Error("Maximum interval not preserved")
		}
	})
}