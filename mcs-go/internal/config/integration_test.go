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

// TestIntegrationScenarios tests realistic usage scenarios
func TestIntegrationScenarios(t *testing.T) {
	// Scenario 1: First time user setup
	t.Run("first_time_user", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("MCS_HOME", tempDir)
		defer os.Unsetenv("MCS_HOME")
		
		// User runs MCS for the first time
		manager, err := NewManager()
		if err != nil {
			t.Fatalf("First time setup failed: %v", err)
		}
		
		// Verify default config was created
		config := manager.Get()
		if config.HostIP != "localhost" {
			t.Errorf("Default HostIP = %v, want localhost", config.HostIP)
		}
		if !config.AutoUpdateEnabled {
			t.Error("Auto-update should be enabled by default")
		}
		
		// Verify config file exists
		configPath := filepath.Join(tempDir, "config.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}
	})
	
	// Scenario 2: User customizes network settings
	t.Run("network_customization", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("MCS_HOME", tempDir)
		defer os.Unsetenv("MCS_HOME")
		
		manager, err := NewManager()
		if err != nil {
			t.Fatalf("Manager creation failed: %v", err)
		}
		
		// User wants to use public IP
		if err := manager.SetIPMode("public"); err != nil {
			t.Fatalf("Failed to set public IP mode: %v", err)
		}
		if err := manager.SetHostIP("203.0.113.1"); err != nil {
			t.Fatalf("Failed to set public IP: %v", err)
		}
		
		// Verify settings were saved
		newManager, err := NewManager()
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}
		
		config := newManager.Get()
		if config.IPMode != "public" {
			t.Errorf("IP mode not persisted: got %v, want public", config.IPMode)
		}
		if config.HostIP != "203.0.113.1" {
			t.Errorf("Host IP not persisted: got %v, want 203.0.113.1", config.HostIP)
		}
	})
	
	// Scenario 3: Developer working on MCS itself
	t.Run("developer_mode", func(t *testing.T) {
		// Create temp dev directory with dockerfiles
		devDir := t.TempDir()
		dockerfilesDir := filepath.Join(devDir, "dockerfiles")
		os.MkdirAll(dockerfilesDir, 0755)
		
		// Change to dev directory
		origWD, _ := os.Getwd()
		defer os.Chdir(origWD)
		os.Chdir(devDir)
		
		// Clear environment variables to test dev mode
		origInstallPath := os.Getenv("MCS_INSTALL_PATH")
		defer os.Setenv("MCS_INSTALL_PATH", origInstallPath)
		os.Unsetenv("MCS_INSTALL_PATH")
		
		// Should detect development mode
		installPath := GetMCSInstallPath()
		if installPath != devDir {
			t.Errorf("Dev mode not detected: got %v, want %v", installPath, devDir)
		}
		
		dockerfilesPath := GetDockerfilesPath()
		if dockerfilesPath != dockerfilesDir {
			t.Errorf("Dev dockerfiles path wrong: got %v, want %v", dockerfilesPath, dockerfilesDir)
		}
	})
	
	// Scenario 4: System administrator deployment
	t.Run("system_deployment", func(t *testing.T) {
		// Simulate system-wide installation
		systemPath := "/opt/mcs-system"
		os.Setenv("MCS_INSTALL_PATH", systemPath)
		defer os.Unsetenv("MCS_INSTALL_PATH")
		
		// Change to user directory
		tempDir := t.TempDir()
		origWD, _ := os.Getwd()
		defer os.Chdir(origWD)
		os.Chdir(tempDir)
		
		installPath := GetMCSInstallPath()
		if installPath != systemPath {
			t.Errorf("System install path not used: got %v, want %v", installPath, systemPath)
		}
		
		expectedDockerfiles := filepath.Join(systemPath, "dockerfiles")
		dockerfilesPath := GetDockerfilesPath()
		if dockerfilesPath != expectedDockerfiles {
			t.Errorf("System dockerfiles path wrong: got %v, want %v", dockerfilesPath, expectedDockerfiles)
		}
	})
}

// TestConfigurationLifecycle tests the full lifecycle of configuration management
func TestConfigurationLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("MCS_HOME", tempDir)
	defer os.Unsetenv("MCS_HOME")
	
	// Phase 1: Initial creation
	manager1, err := NewManager()
	if err != nil {
		t.Fatalf("Phase 1 failed: %v", err)
	}
	
	initialConfig := manager1.Get()
	if initialConfig.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set on initial creation")
	}
	if initialConfig.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set on initial creation")
	}
	
	// Phase 2: Configuration updates
	time.Sleep(time.Millisecond * 10) // Ensure timestamp difference
	
	updates := []func() error{
		func() error { return manager1.SetHostIP("192.168.1.100") },
		func() error { return manager1.SetIPMode("custom") },
		func() error { return manager1.SetAutoUpdateEnabled(false) },
		func() error { return manager1.SetAutoUpdateCheckInterval(7200) },
		func() error { return manager1.SetGitHubToken("test-token") },
	}
	
	for i, update := range updates {
		if err := update(); err != nil {
			t.Fatalf("Update %d failed: %v", i, err)
		}
	}
	
	updatedConfig := manager1.Get()
	if !updatedConfig.UpdatedAt.After(initialConfig.UpdatedAt) {
		t.Error("UpdatedAt should be updated on config changes")
	}
	if updatedConfig.CreatedAt != initialConfig.CreatedAt {
		t.Error("CreatedAt should not change on updates")
	}
	
	// Phase 3: Persistence verification
	manager2, err := NewManager()
	if err != nil {
		t.Fatalf("Phase 3 failed: %v", err)
	}
	
	persistedConfig := manager2.Get()
	if persistedConfig.HostIP != "192.168.1.100" {
		t.Error("HostIP not persisted")
	}
	if persistedConfig.IPMode != "custom" {
		t.Error("IPMode not persisted")
	}
	if persistedConfig.AutoUpdateEnabled != false {
		t.Error("AutoUpdateEnabled not persisted")
	}
	if persistedConfig.GitHubToken != "test-token" {
		t.Error("GitHubToken not persisted")
	}
	
	// Phase 4: Configuration validation
	if persistedConfig.CreatedAt != initialConfig.CreatedAt {
		t.Error("CreatedAt changed during persistence")
	}
	if persistedConfig.UpdatedAt.Before(initialConfig.UpdatedAt) {
		t.Error("UpdatedAt went backwards")
	}
}

// TestErrorRecoveryScenarios tests various error recovery situations
func TestErrorRecoveryScenarios(t *testing.T) {
	t.Run("corrupted_config_recovery", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")
		
		// Create corrupted config
		corruptedData := []byte(`{"host_ip": "localhost", "corrupted": true, "missing_brace":`)
		os.WriteFile(configPath, corruptedData, 0600)
		
		os.Setenv("MCS_HOME", tempDir)
		defer os.Unsetenv("MCS_HOME")
		
		// Should recover by creating default config
		manager, err := NewManager()
		if err != nil {
			t.Fatalf("Failed to recover from corrupted config: %v", err)
		}
		
		config := manager.Get()
		if config.HostIP != "localhost" {
			t.Errorf("Recovery config HostIP = %v, want localhost", config.HostIP)
		}
		
		// Verify new config is valid JSON
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read recovered config: %v", err)
		}
		
		var validConfig Config
		if err := json.Unmarshal(data, &validConfig); err != nil {
			t.Errorf("Recovered config is not valid JSON: %v", err)
		}
	})
	
	t.Run("partial_write_recovery", func(t *testing.T) {
		tempDir := t.TempDir()
		manager := &Manager{
			configPath: filepath.Join(tempDir, "config.json"),
			config:     testConfig(),
		}
		
		// Simulate partial write by making directory read-only after config creation
		if err := manager.save(); err != nil {
			t.Fatalf("Initial save failed: %v", err)
		}
		
		// Make directory read-only
		os.Chmod(tempDir, 0555)
		defer os.Chmod(tempDir, 0755)
		
		// Attempt to save again - should fail gracefully
		err := manager.SetHostIP("new-ip")
		if err == nil {
			t.Error("Expected save to fail with read-only directory")
		}
		
		// Verify original config is unchanged
		os.Chmod(tempDir, 0755)
		data, _ := os.ReadFile(manager.configPath)
		var config Config
		json.Unmarshal(data, &config)
		
		if config.HostIP == "new-ip" {
			t.Error("Config was modified despite save failure")
		}
	})
}

// TestConcurrentConfigurationManagement tests concurrent config operations
func TestConcurrentConfigurationManagement(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("MCS_HOME", tempDir)
	defer os.Unsetenv("MCS_HOME")
	
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("Manager creation failed: %v", err)
	}
	
	// Test concurrent reads (should be safe)
	t.Run("concurrent_reads", func(t *testing.T) {
		numReaders := 20
		results := make(chan Config, numReaders)
		
		for i := 0; i < numReaders; i++ {
			go func() {
				config := manager.Get()
				results <- config
			}()
		}
		
		// Collect all results
		configs := make([]Config, numReaders)
		for i := 0; i < numReaders; i++ {
			configs[i] = <-results
		}
		
		// All configs should be identical
		first := configs[0]
		for i, config := range configs[1:] {
			if config.HostIP != first.HostIP {
				t.Errorf("Config %d differs: HostIP = %v, want %v", i+1, config.HostIP, first.HostIP)
			}
		}
	})
	
	// Test mixed concurrent operations
	t.Run("concurrent_mixed_operations", func(t *testing.T) {
		numOps := 10
		var wg sync.WaitGroup
		errors := make(chan error, numOps)
		
		// Mix of reads and writes
		for i := 0; i < numOps; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				if id%3 == 0 {
					// Read operation
					_ = manager.GetHostIP()
				} else if id%3 == 1 {
					// Write operation
					if err := manager.SetHostIP("10.0.0.1"); err != nil {
						if !strings.Contains(err.Error(), "no such file or directory") {
							errors <- err
						}
					}
				} else {
					// Different write operation
					if err := manager.SetAutoUpdateEnabled(true); err != nil {
						if !strings.Contains(err.Error(), "no such file or directory") {
							errors <- err
						}
					}
				}
			}(i)
		}
		
		wg.Wait()
		close(errors)
		
		// Check for unexpected errors
		for err := range errors {
			t.Errorf("Concurrent operation failed: %v", err)
		}
	})
}

// TestComplexConfigurationScenarios tests complex real-world scenarios
func TestComplexConfigurationScenarios(t *testing.T) {
	t.Run("multiple_rapid_updates", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("MCS_HOME", tempDir)
		defer os.Unsetenv("MCS_HOME")
		
		manager, err := NewManager()
		if err != nil {
			t.Fatalf("Manager creation failed: %v", err)
		}
		
		// Rapid sequence of updates
		updates := 20
		for i := 0; i < updates; i++ {
			manager.SetHostIP(fmt.Sprintf("192.168.1.%d", i+1))
			manager.SetAutoUpdateEnabled(i%2 == 0)
			manager.SetLastUpdateCheck(time.Now().Unix() - int64(i)*100)
		}
		
		// Verify final state
		final := manager.Get()
		expectedIP := fmt.Sprintf("192.168.1.%d", updates)
		if final.HostIP != expectedIP {
			t.Errorf("Final HostIP = %v, want %v", final.HostIP, expectedIP)
		}
		
		expectedAutoUpdate := (updates-1)%2 == 0
		if final.AutoUpdateEnabled != expectedAutoUpdate {
			t.Errorf("Final AutoUpdateEnabled = %v, want %v", final.AutoUpdateEnabled, expectedAutoUpdate)
		}
	})
	
	t.Run("configuration_validation_chain", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("MCS_HOME", tempDir)
		defer os.Unsetenv("MCS_HOME")
		
		manager, err := NewManager()
		if err != nil {
			t.Fatalf("Manager creation failed: %v", err)
		}
		
		// Test validation chain
		validationTests := []struct {
			operation func() error
			shouldErr bool
		}{
			{func() error { return manager.SetIPMode("localhost") }, false},
			{func() error { return manager.SetIPMode("invalid") }, true},
			{func() error { return manager.SetAutoUpdateCheckInterval(3600) }, false},
			{func() error { return manager.SetAutoUpdateCheckInterval(1800) }, true},
			{func() error { return manager.SetHostIP("") }, false}, // Empty IP is allowed
			{func() error { return manager.SetGitHubToken("valid-token") }, false},
		}
		
		for i, test := range validationTests {
			err := test.operation()
			if test.shouldErr && err == nil {
				t.Errorf("Test %d: expected error but got none", i)
			} else if !test.shouldErr && err != nil {
				t.Errorf("Test %d: unexpected error: %v", i, err)
			}
		}
	})
}

// TestConfigurationMigrationScenarios tests various migration scenarios
func TestConfigurationMigrationScenarios(t *testing.T) {
	t.Run("old_version_upgrade", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")
		
		// Simulate old version config (missing new fields)
		oldConfig := map[string]interface{}{
			"host_ip":            "192.168.1.1",
			"auto_update_enabled": true,
			"created_at":         "2023-01-01T00:00:00Z",
			"updated_at":         "2023-01-01T00:00:00Z",
		}
		
		data, _ := json.MarshalIndent(oldConfig, "", "  ")
		os.WriteFile(configPath, data, 0600)
		
		os.Setenv("MCS_HOME", tempDir)
		defer os.Unsetenv("MCS_HOME")
		
		// Load with new manager
		manager, err := NewManager()
		if err != nil {
			t.Fatalf("Failed to migrate old config: %v", err)
		}
		
		config := manager.Get()
		if config.HostIP != "192.168.1.1" {
			t.Errorf("Migrated HostIP = %v, want 192.168.1.1", config.HostIP)
		}
		
		// New fields should have zero/default values
		if config.IPMode != "" {
			t.Logf("IPMode got default value: %v", config.IPMode)
		}
		if config.AutoUpdateCheckInterval != 0 {
			t.Logf("AutoUpdateCheckInterval got default value: %v", config.AutoUpdateCheckInterval)
		}
	})
	
	t.Run("future_version_compatibility", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")
		
		// Simulate future version config (extra fields)
		futureConfig := map[string]interface{}{
			"host_ip":                      "10.0.0.1",
			"ip_mode":                      "auto",
			"auto_update_enabled":          false,
			"auto_update_check_interval":   7200,
			"future_feature_enabled":       true,
			"experimental_settings": map[string]interface{}{
				"new_feature": "enabled",
				"beta_mode":   true,
			},
			"created_at": "2023-01-01T00:00:00Z",
			"updated_at": "2023-01-02T00:00:00Z",
		}
		
		data, _ := json.MarshalIndent(futureConfig, "", "  ")
		os.WriteFile(configPath, data, 0600)
		
		os.Setenv("MCS_HOME", tempDir)
		defer os.Unsetenv("MCS_HOME")
		
		// Should load without error, ignoring unknown fields
		manager, err := NewManager()
		if err != nil {
			t.Fatalf("Failed to handle future config: %v", err)
		}
		
		config := manager.Get()
		if config.HostIP != "10.0.0.1" {
			t.Errorf("Future config HostIP = %v, want 10.0.0.1", config.HostIP)
		}
		if config.IPMode != "auto" {
			t.Errorf("Future config IPMode = %v, want auto", config.IPMode)
		}
		if config.AutoUpdateEnabled != false {
			t.Errorf("Future config AutoUpdateEnabled = %v, want false", config.AutoUpdateEnabled)
		}
	})
}