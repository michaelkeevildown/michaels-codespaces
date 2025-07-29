package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestFullWorkflow tests the complete configuration workflow
func TestFullWorkflow(t *testing.T) {
	// Save original environment
	origMCSHome := os.Getenv("MCS_HOME")
	origHome := os.Getenv("HOME")
	defer func() {
		os.Setenv("MCS_HOME", origMCSHome)
		os.Setenv("HOME", origHome)
	}()
	
	// Setup test environment
	tempDir := t.TempDir()
	os.Setenv("MCS_HOME", tempDir)
	
	// 1. Create new manager (should create default config)
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}
	
	// 2. Verify default config
	config := manager.Get()
	if config.HostIP != "localhost" {
		t.Errorf("Default HostIP = %v, want localhost", config.HostIP)
	}
	if config.AutoUpdateEnabled != true {
		t.Errorf("Default AutoUpdateEnabled = %v, want true", config.AutoUpdateEnabled)
	}
	
	// 3. Update network settings
	if err := manager.SetHostIP("192.168.1.100"); err != nil {
		t.Fatalf("SetHostIP() failed: %v", err)
	}
	if err := manager.SetIPMode("custom"); err != nil {
		t.Fatalf("SetIPMode() failed: %v", err)
	}
	
	// 4. Update auto-update settings
	if err := manager.SetAutoUpdateEnabled(false); err != nil {
		t.Fatalf("SetAutoUpdateEnabled() failed: %v", err)
	}
	if err := manager.SetAutoUpdateCheckInterval(7200); err != nil {
		t.Fatalf("SetAutoUpdateCheckInterval() failed: %v", err)
	}
	if err := manager.SetLastUpdateCheck(time.Now().Unix()); err != nil {
		t.Fatalf("SetLastUpdateCheck() failed: %v", err)
	}
	if err := manager.SetLastKnownVersion("2.0.0"); err != nil {
		t.Fatalf("SetLastKnownVersion() failed: %v", err)
	}
	
	// 5. Set authentication
	if err := manager.SetGitHubToken("test-token-123"); err != nil {
		t.Fatalf("SetGitHubToken() failed: %v", err)
	}
	
	// 6. Create second manager instance (should load existing config)
	manager2, err := NewManager()
	if err != nil {
		t.Fatalf("Second NewManager() failed: %v", err)
	}
	
	// 7. Verify persistence
	config2 := manager2.Get()
	if config2.HostIP != "192.168.1.100" {
		t.Errorf("Persisted HostIP = %v, want 192.168.1.100", config2.HostIP)
	}
	if config2.IPMode != "custom" {
		t.Errorf("Persisted IPMode = %v, want custom", config2.IPMode)
	}
	if config2.AutoUpdateEnabled != false {
		t.Errorf("Persisted AutoUpdateEnabled = %v, want false", config2.AutoUpdateEnabled)
	}
	if config2.GitHubToken != "test-token-123" {
		t.Errorf("Persisted GitHubToken = %v, want test-token-123", config2.GitHubToken)
	}
	
	// 8. Test update check logic
	shouldCheck := manager2.ShouldCheckForUpdate()
	if shouldCheck != false {
		t.Errorf("ShouldCheckForUpdate() = %v, want false (auto-update disabled)", shouldCheck)
	}
	
	// Enable auto-update and test again
	if err := manager2.SetAutoUpdateEnabled(true); err != nil {
		t.Fatalf("Re-enabling auto-update failed: %v", err)
	}
	shouldCheck = manager2.ShouldCheckForUpdate()
	if shouldCheck != false {
		t.Errorf("ShouldCheckForUpdate() = %v, want false (recent check)", shouldCheck)
	}
}

// TestPathDiscoveryWorkflow tests path discovery in different scenarios
func TestPathDiscoveryWorkflow(t *testing.T) {
	// Save original environment
	origInstallPath := os.Getenv("MCS_INSTALL_PATH")
	origHome := os.Getenv("HOME")
	origWD, _ := os.Getwd()
	defer func() {
		os.Setenv("MCS_INSTALL_PATH", origInstallPath)
		os.Setenv("HOME", origHome)
		os.Chdir(origWD)
	}()
	
	// Test 1: Development mode
	t.Run("development_mode_workflow", func(t *testing.T) {
		devDir := t.TempDir()
		dockerfilesDir := filepath.Join(devDir, "dockerfiles")
		os.MkdirAll(dockerfilesDir, 0755)
		os.Chdir(devDir)
		os.Unsetenv("MCS_INSTALL_PATH")
		
		installPath := GetMCSInstallPath()
		dockerfilesPath := GetDockerfilesPath()
		
		if installPath != devDir {
			t.Errorf("Dev mode install path = %v, want %v", installPath, devDir)
		}
		if dockerfilesPath != dockerfilesDir {
			t.Errorf("Dev mode dockerfiles path = %v, want %v", dockerfilesPath, dockerfilesDir)
		}
	})
	
	// Test 2: Custom installation path
	t.Run("custom_install_path_workflow", func(t *testing.T) {
		customPath := "/opt/mcs-custom"
		os.Setenv("MCS_INSTALL_PATH", customPath)
		
		// Change to directory without dockerfiles
		tempDir := t.TempDir()
		os.Chdir(tempDir)
		
		installPath := GetMCSInstallPath()
		dockerfilesPath := GetDockerfilesPath()
		
		if installPath != customPath {
			t.Errorf("Custom install path = %v, want %v", installPath, customPath)
		}
		expectedDockerfiles := filepath.Join(customPath, "dockerfiles")
		if dockerfilesPath != expectedDockerfiles {
			t.Errorf("Custom dockerfiles path = %v, want %v", dockerfilesPath, expectedDockerfiles)
		}
	})
	
	// Test 3: Default home directory fallback
	t.Run("home_directory_fallback", func(t *testing.T) {
		tempHome := t.TempDir()
		os.Setenv("HOME", tempHome)
		os.Unsetenv("MCS_INSTALL_PATH")
		
		// Change to directory without dockerfiles
		tempDir := t.TempDir()
		os.Chdir(tempDir)
		
		installPath := GetMCSInstallPath()
		dockerfilesPath := GetDockerfilesPath()
		
		expectedInstall := filepath.Join(tempHome, ".mcs")
		expectedDockerfiles := filepath.Join(expectedInstall, "dockerfiles")
		
		if installPath != expectedInstall {
			t.Errorf("Home fallback install path = %v, want %v", installPath, expectedInstall)
		}
		if dockerfilesPath != expectedDockerfiles {
			t.Errorf("Home fallback dockerfiles path = %v, want %v", dockerfilesPath, expectedDockerfiles)
		}
	})
}

// TestErrorRecovery tests error recovery scenarios
func TestErrorRecovery(t *testing.T) {
	tempDir := t.TempDir()
	
	t.Run("recover_from_corrupted_config", func(t *testing.T) {
		// Create corrupted config file
		configPath := filepath.Join(tempDir, "corrupted", "config.json")
		os.MkdirAll(filepath.Dir(configPath), 0755)
		os.WriteFile(configPath, []byte("corrupted json content"), 0600)
		
		// Set environment to use corrupted config
		os.Setenv("MCS_HOME", filepath.Dir(configPath))
		defer os.Unsetenv("MCS_HOME")
		
		// NewManager should recover by creating default config
		manager, err := NewManager()
		if err != nil {
			t.Fatalf("Failed to recover from corrupted config: %v", err)
		}
		
		// Should have default values
		config := manager.Get()
		if config.HostIP != "localhost" {
			t.Errorf("Recovery config HostIP = %v, want localhost", config.HostIP)
		}
		if config.IPMode != "localhost" {
			t.Errorf("Recovery config IPMode = %v, want localhost", config.IPMode)
		}
	})
	
	t.Run("handle_permission_errors_gracefully", func(t *testing.T) {
		// This test would require root permissions to create truly inaccessible directories
		// Instead, we test the error handling path
		manager := &Manager{
			configPath: "/root/inaccessible/config.json",
			config:     testConfig(),
		}
		
		err := manager.save()
		if err == nil {
			t.Error("Expected permission error, got nil")
		}
		
		// Error should contain relevant message
		if err != nil && !contains(err.Error(), "failed to write config") {
			t.Errorf("Expected write error, got: %v", err)
		}
	})
}

// TestConfigurationMigration tests handling of configuration migrations
func TestConfigurationMigration(t *testing.T) {
	tests := []struct {
		name       string
		oldConfig  string
		expectLoad bool
		validate   func(*testing.T, *Config)
	}{
		{
			name: "v1.0_config_format",
			oldConfig: `{
				"host_ip": "localhost"
			}`,
			expectLoad: true,
			validate: func(t *testing.T, c *Config) {
				if c.HostIP != "localhost" {
					t.Errorf("Migrated HostIP = %v, want localhost", c.HostIP)
				}
				// New fields should have zero values
				if c.IPMode != "" {
					t.Logf("IPMode defaulted to: %v", c.IPMode)
				}
			},
		},
		{
			name: "config_with_extra_fields",
			oldConfig: `{
				"host_ip": "192.168.1.1",
				"ip_mode": "custom",
				"deprecated_field": "should_be_ignored",
				"nested_deprecated": {
					"old_setting": true
				}
			}`,
			expectLoad: true,
			validate: func(t *testing.T, c *Config) {
				if c.HostIP != "192.168.1.1" {
					t.Errorf("Migrated HostIP = %v, want 192.168.1.1", c.HostIP)
				}
				if c.IPMode != "custom" {
					t.Errorf("Migrated IPMode = %v, want custom", c.IPMode)
				}
				// Extra fields should be ignored (no error)
			},
		},
		{
			name: "config_with_wrong_types",
			oldConfig: `{
				"host_ip": 12345,
				"auto_update_enabled": "true",
				"auto_update_check_interval": "3600"
			}`,
			expectLoad: false,
			validate: func(t *testing.T, c *Config) {
				// Should fail to load due to type mismatches
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.json")
			
			// Write old config format
			os.WriteFile(configPath, []byte(tt.oldConfig), 0600)
			
			manager := &Manager{configPath: configPath}
			err := manager.load()
			
			if tt.expectLoad && err != nil {
				t.Errorf("Expected successful load, got error: %v", err)
			} else if !tt.expectLoad && err == nil {
				t.Error("Expected load error, got success")
			}
			
			if tt.expectLoad && manager.config != nil {
				tt.validate(t, manager.config)
			}
		})
	}
}

// TestPerformanceCharacteristics tests performance under various conditions
func TestPerformanceCharacteristics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}
	
	t.Run("large_config_values", func(t *testing.T) {
		manager := createTempManager(t, testConfig())
		
		// Test with large values
		largeIP := make([]byte, 10000)
		for i := range largeIP {
			largeIP[i] = '1'
		}
		
		start := time.Now()
		err := manager.SetHostIP(string(largeIP))
		duration := time.Since(start)
		
		if err != nil {
			t.Errorf("Failed to handle large IP: %v", err)
		}
		
		if duration > time.Millisecond*100 {
			t.Errorf("Large config save took too long: %v", duration)
		}
		
		// Verify it was saved and can be loaded
		start = time.Now()
		retrievedIP := manager.GetHostIP()
		duration = time.Since(start)
		
		if retrievedIP != string(largeIP) {
			t.Error("Large IP not preserved")
		}
		
		if duration > time.Millisecond*10 {
			t.Errorf("Large config read took too long: %v", duration)
		}
	})
	
	t.Run("frequent_updates", func(t *testing.T) {
		manager := createTempManager(t, testConfig())
		
		start := time.Now()
		for i := 0; i < 100; i++ {
			if i%2 == 0 {
				manager.SetHostIP("192.168.1.1")
			} else {
				manager.SetHostIP("10.0.0.1")
			}
		}
		duration := time.Since(start)
		
		avgPerUpdate := duration / 100
		if avgPerUpdate > time.Millisecond*10 {
			t.Errorf("Average update time too slow: %v", avgPerUpdate)
		}
	})
}

// TestComprehensiveCoverage ensures we test all code paths
func TestComprehensiveCoverage(t *testing.T) {
	// Test all IP modes
	modes := []string{"localhost", "auto", "public", "custom"}
	manager := createTempManager(t, testConfig())
	
	for _, mode := range modes {
		t.Run("ip_mode_"+mode, func(t *testing.T) {
			if err := manager.SetIPMode(mode); err != nil {
				t.Errorf("Failed to set IP mode %s: %v", mode, err)
			}
			if manager.GetIPMode() != mode {
				t.Errorf("IP mode not set correctly: got %v, want %v", manager.GetIPMode(), mode)
			}
		})
	}
	
	// Test boundary values for interval
	intervals := []int64{3600, 7200, 86400, 604800} // 1h, 2h, 1d, 1week
	for _, interval := range intervals {
		t.Run("interval_boundary", func(t *testing.T) {
			if err := manager.SetAutoUpdateCheckInterval(interval); err != nil {
				t.Errorf("Failed to set interval %d: %v", interval, err)
			}
			if manager.GetAutoUpdateCheckInterval() != interval {
				t.Errorf("Interval not set correctly: got %v, want %v", manager.GetAutoUpdateCheckInterval(), interval)
			}
		})
	}
	
	// Test update check scenarios
	testCases := []struct {
		name           string
		enabled        bool
		lastCheck      int64
		interval       int64
		expectedResult bool
	}{
		{"disabled", false, 0, 3600, false},
		{"never_checked", true, 0, 3600, true},
		{"due_for_check", true, time.Now().Unix() - 7200, 3600, true},
		{"not_due", true, time.Now().Unix() - 1800, 3600, false},
	}
	
	for _, tc := range testCases {
		t.Run("update_check_"+tc.name, func(t *testing.T) {
			manager.SetAutoUpdateEnabled(tc.enabled)
			manager.SetLastUpdateCheck(tc.lastCheck)
			manager.SetAutoUpdateCheckInterval(tc.interval)
			
			result := manager.ShouldCheckForUpdate()
			if result != tc.expectedResult {
				t.Errorf("ShouldCheckForUpdate() = %v, want %v", result, tc.expectedResult)
			}
		})
	}
}

// TestEdgeCaseValidation tests edge cases that might cause issues
func TestEdgeCaseValidation(t *testing.T) {
	manager := createTempManager(t, testConfig())
	
	// Test empty string values
	emptyTests := []struct {
		name string
		fn   func() error
	}{
		{"empty_host_ip", func() error { return manager.SetHostIP("") }},
		{"empty_version", func() error { return manager.SetLastKnownVersion("") }},
		{"empty_token", func() error { return manager.SetGitHubToken("") }},
	}
	
	for _, test := range emptyTests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.fn(); err != nil {
				t.Errorf("Failed to handle empty value: %v", err)
			}
		})
	}
	
	// Test special characters
	specialChars := []string{
		"special!@#$%^&*()",
		"unicode测试",
		"newline\ntest",
		"tab\ttest",
		"quote\"test",
		"backslash\\test",
	}
	
	for _, special := range specialChars {
		t.Run("special_chars", func(t *testing.T) {
			if err := manager.SetHostIP(special); err != nil {
				t.Errorf("Failed to handle special characters %q: %v", special, err)
			}
			if manager.GetHostIP() != special {
				t.Errorf("Special characters not preserved: got %q, want %q", manager.GetHostIP(), special)
			}
		})
	}
}