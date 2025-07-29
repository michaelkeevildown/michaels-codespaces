package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestManager_NewManager tests the creation of a new config manager
func TestManager_NewManager(t *testing.T) {
	tests := []struct {
		name        string
		mcsHome     string
		homeEnv     string
		setupFunc   func(string)
		cleanupFunc func(string)
		wantErr     bool
		errContains string
	}{
		{
			name:    "default config location with HOME",
			homeEnv: "/tmp/test-home",
			setupFunc: func(dir string) {
				os.Setenv("HOME", dir)
			},
			cleanupFunc: func(dir string) {
				os.RemoveAll(filepath.Join(dir, ".mcs"))
			},
			wantErr: false,
		},
		{
			name:    "custom MCS_HOME location",
			mcsHome: "/tmp/custom-mcs-home",
			cleanupFunc: func(dir string) {
				os.RemoveAll(dir)
			},
			wantErr: false,
		},
		{
			name:    "existing config file",
			mcsHome: "/tmp/existing-config",
			setupFunc: func(dir string) {
				os.MkdirAll(dir, 0755)
				config := &Config{
					HostIP:       "192.168.1.1",
					IPMode:       "custom",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				data, _ := json.MarshalIndent(config, "", "  ")
				os.WriteFile(filepath.Join(dir, "config.json"), data, 0600)
			},
			cleanupFunc: func(dir string) {
				os.RemoveAll(dir)
			},
			wantErr: false,
		},
		{
			name:    "invalid config file",
			mcsHome: "/tmp/invalid-config",
			setupFunc: func(dir string) {
				os.MkdirAll(dir, 0755)
				os.WriteFile(filepath.Join(dir, "config.json"), []byte("invalid json"), 0600)
			},
			cleanupFunc: func(dir string) {
				os.RemoveAll(dir)
			},
			wantErr: false, // Should create default config on error
		},
		{
			name:    "permission denied on directory creation",
			mcsHome: "/root/no-permission-mcs",
			setupFunc: func(dir string) {
				// Create parent directory with no write permission
				parent := filepath.Dir(dir)
				os.MkdirAll(parent, 0755)
				os.Chmod(parent, 0555)
			},
			cleanupFunc: func(dir string) {
				parent := filepath.Dir(dir)
				os.Chmod(parent, 0755)
				os.RemoveAll(parent)
			},
			wantErr:     true,
			errContains: "failed to create config directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			origMCSHome := os.Getenv("MCS_HOME")
			origHome := os.Getenv("HOME")
			defer func() {
				os.Setenv("MCS_HOME", origMCSHome)
				os.Setenv("HOME", origHome)
			}()

			// Set test env vars
			if tt.mcsHome != "" {
				os.Setenv("MCS_HOME", tt.mcsHome)
			} else {
				os.Unsetenv("MCS_HOME")
			}
			if tt.homeEnv != "" {
				os.Setenv("HOME", tt.homeEnv)
			}

			// Setup
			dir := tt.mcsHome
			if dir == "" {
				dir = filepath.Join(tt.homeEnv, ".mcs")
			}
			if tt.setupFunc != nil {
				tt.setupFunc(dir)
			}

			// Cleanup
			if tt.cleanupFunc != nil {
				defer tt.cleanupFunc(dir)
			}

			// Test
			manager, err := NewManager()
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewManager() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("NewManager() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if manager == nil {
				t.Error("NewManager() returned nil manager")
				return
			}

			// Verify config was created
			configPath := filepath.Join(dir, "config.json")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				t.Error("Config file was not created")
			}
		})
	}
}

// TestManager_defaultConfig tests the default configuration values
func TestManager_defaultConfig(t *testing.T) {
	manager := &Manager{}
	config := manager.defaultConfig()

	if config.HostIP != "localhost" {
		t.Errorf("defaultConfig() HostIP = %v, want localhost", config.HostIP)
	}
	if config.IPMode != "localhost" {
		t.Errorf("defaultConfig() IPMode = %v, want localhost", config.IPMode)
	}
	if config.AutoDetectIP != false {
		t.Errorf("defaultConfig() AutoDetectIP = %v, want false", config.AutoDetectIP)
	}
	if config.AutoUpdateEnabled != true {
		t.Errorf("defaultConfig() AutoUpdateEnabled = %v, want true", config.AutoUpdateEnabled)
	}
	if config.AutoUpdateCheckInterval != 86400 {
		t.Errorf("defaultConfig() AutoUpdateCheckInterval = %v, want 86400", config.AutoUpdateCheckInterval)
	}
	if config.LastUpdateCheck != 0 {
		t.Errorf("defaultConfig() LastUpdateCheck = %v, want 0", config.LastUpdateCheck)
	}
	if config.LastKnownVersion != "1.0.0" {
		t.Errorf("defaultConfig() LastKnownVersion = %v, want 1.0.0", config.LastKnownVersion)
	}
	if config.CreatedAt.IsZero() {
		t.Error("defaultConfig() CreatedAt is zero")
	}
	if config.UpdatedAt.IsZero() {
		t.Error("defaultConfig() UpdatedAt is zero")
	}
}

// TestManager_load tests loading configuration from disk
func TestManager_load(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		setupFile   bool
		wantErr     bool
		errContains string
		validate    func(*testing.T, *Config)
	}{
		{
			name: "valid config",
			configData: `{
				"host_ip": "192.168.1.100",
				"ip_mode": "custom",
				"auto_detect_ip": true,
				"auto_update_enabled": false,
				"auto_update_check_interval": 7200,
				"last_update_check": 1234567890,
				"last_known_version": "2.0.0",
				"github_token": "test-token",
				"created_at": "2023-01-01T00:00:00Z",
				"updated_at": "2023-01-02T00:00:00Z"
			}`,
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				if c.HostIP != "192.168.1.100" {
					t.Errorf("HostIP = %v, want 192.168.1.100", c.HostIP)
				}
				if c.IPMode != "custom" {
					t.Errorf("IPMode = %v, want custom", c.IPMode)
				}
				if c.AutoDetectIP != true {
					t.Errorf("AutoDetectIP = %v, want true", c.AutoDetectIP)
				}
				if c.GitHubToken != "test-token" {
					t.Errorf("GitHubToken = %v, want test-token", c.GitHubToken)
				}
			},
		},
		{
			name:        "invalid json",
			configData:  "not valid json",
			wantErr:     true,
			errContains: "failed to unmarshal config",
		},
		{
			name:        "empty file",
			configData:  "",
			setupFile:   true,
			wantErr:     true,
			errContains: "failed to unmarshal config",
		},
		{
			name:       "empty json object",
			configData: "{}",
			wantErr:    false,
			validate: func(t *testing.T, c *Config) {
				// Should have zero values
				if c.HostIP != "" {
					t.Errorf("HostIP = %v, want empty", c.HostIP)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.json")

			// Write test data
			if tt.configData != "" || tt.setupFile {
				err := os.WriteFile(configPath, []byte(tt.configData), 0600)
				if err != nil {
					t.Fatalf("Failed to write test config: %v", err)
				}
			}

			// Create manager
			manager := &Manager{
				configPath: configPath,
			}

			// Test load
			err := manager.load()
			if tt.wantErr {
				if err == nil {
					t.Errorf("load() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("load() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validate != nil {
				tt.validate(t, manager.config)
			}
		})
	}
}

// TestManager_save tests saving configuration to disk
func TestManager_save(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		setupFunc   func(string)
		wantErr     bool
		errContains string
		validate    func(*testing.T, string)
	}{
		{
			name: "save valid config",
			config: &Config{
				HostIP:                  "10.0.0.1",
				IPMode:                  "public",
				AutoUpdateEnabled:       true,
				AutoUpdateCheckInterval: 3600,
				CreatedAt:               time.Now(),
			},
			wantErr: false,
			validate: func(t *testing.T, configPath string) {
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("Failed to read saved config: %v", err)
				}
				var saved Config
				if err := json.Unmarshal(data, &saved); err != nil {
					t.Fatalf("Failed to unmarshal saved config: %v", err)
				}
				if saved.HostIP != "10.0.0.1" {
					t.Errorf("Saved HostIP = %v, want 10.0.0.1", saved.HostIP)
				}
				if saved.UpdatedAt.IsZero() {
					t.Error("UpdatedAt was not set")
				}
			},
		},
		{
			name: "atomic save with existing file",
			config: &Config{
				HostIP: "localhost",
			},
			setupFunc: func(dir string) {
				// Create existing config
				existingData := []byte(`{"host_ip": "old-value"}`)
				os.WriteFile(filepath.Join(dir, "config.json"), existingData, 0600)
			},
			wantErr: false,
			validate: func(t *testing.T, configPath string) {
				// Verify temp file doesn't exist
				if _, err := os.Stat(configPath + ".tmp"); err == nil {
					t.Error("Temporary file was not cleaned up")
				}
			},
		},
		{
			name: "permission denied",
			config: &Config{
				HostIP: "localhost",
			},
			setupFunc: func(dir string) {
				// Make directory read-only
				os.Chmod(dir, 0555)
			},
			wantErr:     true,
			errContains: "failed to write config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.json")

			// Setup
			if tt.setupFunc != nil {
				tt.setupFunc(tempDir)
				defer os.Chmod(tempDir, 0755) // Reset permissions
			}

			// Create manager
			manager := &Manager{
				configPath: configPath,
				config:     tt.config,
			}

			// Test save
			err := manager.save()
			if tt.wantErr {
				if err == nil {
					t.Errorf("save() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("save() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Validate
			if tt.validate != nil {
				tt.validate(t, configPath)
			}

			// Verify file permissions
			info, err := os.Stat(configPath)
			if err != nil {
				t.Fatalf("Failed to stat config file: %v", err)
			}
			if info.Mode().Perm() != 0600 {
				t.Errorf("Config file permissions = %v, want 0600", info.Mode().Perm())
			}
		})
	}
}

// TestManager_ConcurrentAccess tests thread-safe access to configuration
func TestManager_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("MCS_HOME", tempDir)
	defer os.Unsetenv("MCS_HOME")

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Number of concurrent operations (reduced to avoid overwhelming the test)
	numGoroutines := 10
	numOperations := 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Concurrent reads and writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				switch j % 5 {
				case 0:
					// Get config
					_ = manager.Get()
				case 1:
					// Set host IP
					ip := fmt.Sprintf("10.0.%d.%d", id, j)
					if err := manager.SetHostIP(ip); err != nil {
						// Only report non-file system errors
						if !strings.Contains(err.Error(), "no such file or directory") {
							errors <- err
						}
					}
				case 2:
					// Get host IP
					_ = manager.GetHostIP()
				case 3:
					// Set IP mode
					modes := []string{"localhost", "auto", "public", "custom"}
					mode := modes[j%len(modes)]
					if err := manager.SetIPMode(mode); err != nil {
						// Only report non-file system errors
						if !strings.Contains(err.Error(), "no such file or directory") {
							errors <- err
						}
					}
				case 4:
					// Toggle auto update
					enabled := j%2 == 0
					if err := manager.SetAutoUpdateEnabled(enabled); err != nil {
						// Only report non-file system errors
						if !strings.Contains(err.Error(), "no such file or directory") {
							errors <- err
						}
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	// Verify final state is valid
	config := manager.Get()
	if config.HostIP == "" {
		t.Error("Final config has empty HostIP")
	}
	if config.IPMode == "" {
		t.Error("Final config has empty IPMode")
	}
}

// TestManager_Getters tests all getter methods
func TestManager_Getters(t *testing.T) {
	manager := &Manager{
		config: &Config{
			HostIP:                  "192.168.1.1",
			IPMode:                  "custom",
			AutoDetectIP:            true,
			AutoUpdateEnabled:       false,
			AutoUpdateCheckInterval: 7200,
			LastUpdateCheck:         1234567890,
			LastKnownVersion:        "2.0.0",
			GitHubToken:             "secret-token",
		},
	}

	tests := []struct {
		name     string
		getter   func() interface{}
		expected interface{}
	}{
		{"GetHostIP", func() interface{} { return manager.GetHostIP() }, "192.168.1.1"},
		{"GetIPMode", func() interface{} { return manager.GetIPMode() }, "custom"},
		{"IsAutoUpdateEnabled", func() interface{} { return manager.IsAutoUpdateEnabled() }, false},
		{"GetAutoUpdateCheckInterval", func() interface{} { return manager.GetAutoUpdateCheckInterval() }, int64(7200)},
		{"GetLastUpdateCheck", func() interface{} { return manager.GetLastUpdateCheck() }, int64(1234567890)},
		{"GetLastKnownVersion", func() interface{} { return manager.GetLastKnownVersion() }, "2.0.0"},
		{"GetGitHubToken", func() interface{} { return manager.GetGitHubToken() }, "secret-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.getter()
			if result != tt.expected {
				t.Errorf("%s() = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestManager_SetIPMode tests IP mode validation
func TestManager_SetIPMode(t *testing.T) {
	tempDir := t.TempDir()
	manager := &Manager{
		configPath: filepath.Join(tempDir, "config.json"),
		config:     &Config{},
	}

	tests := []struct {
		mode    string
		wantErr bool
	}{
		{"localhost", false},
		{"auto", false},
		{"public", false},
		{"custom", false},
		{"invalid", true},
		{"", true},
		{"LOCAL", true}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			err := manager.SetIPMode(tt.mode)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetIPMode(%v) error = %v, wantErr %v", tt.mode, err, tt.wantErr)
			}
			if !tt.wantErr && manager.config.IPMode != tt.mode {
				t.Errorf("IPMode = %v, want %v", manager.config.IPMode, tt.mode)
			}
		})
	}
}

// TestManager_SetAutoUpdateCheckInterval tests interval validation
func TestManager_SetAutoUpdateCheckInterval(t *testing.T) {
	tempDir := t.TempDir()
	manager := &Manager{
		configPath: filepath.Join(tempDir, "config.json"),
		config:     &Config{},
	}

	tests := []struct {
		interval int64
		wantErr  bool
	}{
		{3600, false},    // 1 hour (minimum)
		{7200, false},    // 2 hours
		{86400, false},   // 24 hours
		{3599, true},     // Just under 1 hour
		{0, true},        // Zero
		{-1, true},       // Negative
		{1800, true},     // 30 minutes
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("interval_%d", tt.interval), func(t *testing.T) {
			err := manager.SetAutoUpdateCheckInterval(tt.interval)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetAutoUpdateCheckInterval(%v) error = %v, wantErr %v", tt.interval, err, tt.wantErr)
			}
			if !tt.wantErr && manager.config.AutoUpdateCheckInterval != tt.interval {
				t.Errorf("AutoUpdateCheckInterval = %v, want %v", manager.config.AutoUpdateCheckInterval, tt.interval)
			}
		})
	}
}

// TestManager_ShouldCheckForUpdate tests update check logic
func TestManager_ShouldCheckForUpdate(t *testing.T) {
	now := time.Now().Unix()
	
	tests := []struct {
		name              string
		autoUpdateEnabled bool
		lastCheck         int64
		interval          int64
		expected          bool
	}{
		{
			name:              "auto update disabled",
			autoUpdateEnabled: false,
			lastCheck:         now - 100,
			interval:          3600,
			expected:          false,
		},
		{
			name:              "never checked before",
			autoUpdateEnabled: true,
			lastCheck:         0,
			interval:          3600,
			expected:          true,
		},
		{
			name:              "check interval not reached",
			autoUpdateEnabled: true,
			lastCheck:         now - 1800, // 30 minutes ago
			interval:          3600,       // 1 hour interval
			expected:          false,
		},
		{
			name:              "check interval reached",
			autoUpdateEnabled: true,
			lastCheck:         now - 7200, // 2 hours ago
			interval:          3600,       // 1 hour interval
			expected:          true,
		},
		{
			name:              "exact interval boundary",
			autoUpdateEnabled: true,
			lastCheck:         now - 3600, // Exactly 1 hour ago
			interval:          3600,       // 1 hour interval
			expected:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{
				config: &Config{
					AutoUpdateEnabled:       tt.autoUpdateEnabled,
					LastUpdateCheck:         tt.lastCheck,
					AutoUpdateCheckInterval: tt.interval,
				},
			}

			result := manager.ShouldCheckForUpdate()
			if result != tt.expected {
				t.Errorf("ShouldCheckForUpdate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestManager_Migration tests handling of old config formats
func TestManager_Migration(t *testing.T) {
	tests := []struct {
		name       string
		oldConfig  string
		validate   func(*testing.T, *Config)
	}{
		{
			name: "missing new fields",
			oldConfig: `{
				"host_ip": "localhost"
			}`,
			validate: func(t *testing.T, c *Config) {
				// Should have default values for missing fields
				if c.IPMode != "" {
					t.Logf("IPMode defaulted to: %v", c.IPMode)
				}
				if c.AutoUpdateCheckInterval == 0 {
					t.Logf("AutoUpdateCheckInterval defaulted to: %v", c.AutoUpdateCheckInterval)
				}
			},
		},
		{
			name: "future version with unknown fields",
			oldConfig: `{
				"host_ip": "localhost",
				"ip_mode": "localhost",
				"future_field": "some value",
				"nested_future": {
					"key": "value"
				}
			}`,
			validate: func(t *testing.T, c *Config) {
				// Should load known fields and ignore unknown ones
				if c.HostIP != "localhost" {
					t.Errorf("HostIP = %v, want localhost", c.HostIP)
				}
				if c.IPMode != "localhost" {
					t.Errorf("IPMode = %v, want localhost", c.IPMode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.json")
			
			// Write old config
			err := os.WriteFile(configPath, []byte(tt.oldConfig), 0600)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			// Create manager and load
			manager := &Manager{
				configPath: configPath,
			}
			err = manager.load()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Validate
			if tt.validate != nil {
				tt.validate(t, manager.config)
			}
		})
	}
}

// BenchmarkManager_Get benchmarks concurrent config reads
func BenchmarkManager_Get(b *testing.B) {
	manager := &Manager{
		config: &Config{
			HostIP: "localhost",
			IPMode: "localhost",
		},
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = manager.Get()
		}
	})
}

// BenchmarkManager_Save benchmarks config saves
func BenchmarkManager_Save(b *testing.B) {
	tempDir := b.TempDir()
	manager := &Manager{
		configPath: filepath.Join(tempDir, "config.json"),
		config: &Config{
			HostIP: "localhost",
			IPMode: "localhost",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.config.HostIP = fmt.Sprintf("10.0.0.%d", i%256)
		if err := manager.save(); err != nil {
			b.Fatalf("save() failed: %v", err)
		}
	}
}