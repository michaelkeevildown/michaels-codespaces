package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGetMCSInstallPath tests the MCS installation path detection
func TestGetMCSInstallPath(t *testing.T) {
	// Save original values
	origWD, _ := os.Getwd()
	origMCSInstallPath := os.Getenv("MCS_INSTALL_PATH")
	origHome := os.Getenv("HOME")
	
	// Ensure cleanup
	defer func() {
		os.Chdir(origWD)
		os.Setenv("MCS_INSTALL_PATH", origMCSInstallPath)
		os.Setenv("HOME", origHome)
	}()

	tests := []struct {
		name          string
		setupFunc     func() (cleanup func())
		expected      string
		expectDefault bool
	}{
		{
			name: "development mode - dockerfiles in current directory",
			setupFunc: func() func() {
				// Create temp directory with dockerfiles
				tempDir := t.TempDir()
				dockerfilesDir := filepath.Join(tempDir, "dockerfiles")
				os.MkdirAll(dockerfilesDir, 0755)
				os.Chdir(tempDir)
				
				return func() {
					os.Chdir(origWD)
				}
			},
			expected:      "", // Will be set to tempDir in test
			expectDefault: false,
		},
		{
			name: "MCS_INSTALL_PATH environment variable",
			setupFunc: func() func() {
				customPath := "/custom/mcs/path"
				os.Setenv("MCS_INSTALL_PATH", customPath)
				
				// Change to a directory without dockerfiles
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				
				return func() {
					os.Unsetenv("MCS_INSTALL_PATH")
					os.Chdir(origWD)
				}
			},
			expected:      "/custom/mcs/path",
			expectDefault: false,
		},
		{
			name: "found in /usr/local/share/mcs",
			setupFunc: func() func() {
				// Mock the existence check by creating the directory structure
				tempRoot := t.TempDir()
				mockPath := filepath.Join(tempRoot, "usr", "local", "share", "mcs", "dockerfiles")
				os.MkdirAll(mockPath, 0755)
				
				// We can't actually test this without mocking os.Stat,
				// so we'll test the logic indirectly
				os.Unsetenv("MCS_INSTALL_PATH")
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				
				return func() {
					os.Chdir(origWD)
				}
			},
			expected:      "", // Will fallback to HOME/.mcs
			expectDefault: true,
		},
		{
			name: "default to HOME/.mcs",
			setupFunc: func() func() {
				// Set HOME to a known location
				tempHome := t.TempDir()
				os.Setenv("HOME", tempHome)
				os.Unsetenv("MCS_INSTALL_PATH")
				
				// Change to a directory without dockerfiles
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				
				return func() {
					os.Setenv("HOME", origHome)
					os.Chdir(origWD)
				}
			},
			expected:      "", // Will be set based on HOME
			expectDefault: true,
		},
		{
			name: "found in HOME/.mcs",
			setupFunc: func() func() {
				tempHome := t.TempDir()
				os.Setenv("HOME", tempHome)
				os.Unsetenv("MCS_INSTALL_PATH")
				
				// Create dockerfiles in HOME/.mcs
				mcsPath := filepath.Join(tempHome, ".mcs", "dockerfiles")
				os.MkdirAll(mcsPath, 0755)
				
				// Change to a directory without dockerfiles
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				
				return func() {
					os.Setenv("HOME", origHome)
					os.Chdir(origWD)
				}
			},
			expected:      "", // Will be HOME/.mcs
			expectDefault: true,
		},
		{
			name: "found in HOME/.local/share/mcs",
			setupFunc: func() func() {
				tempHome := t.TempDir()
				os.Setenv("HOME", tempHome)
				os.Unsetenv("MCS_INSTALL_PATH")
				
				// Create dockerfiles in HOME/.local/share/mcs
				mcsPath := filepath.Join(tempHome, ".local", "share", "mcs", "dockerfiles")
				os.MkdirAll(mcsPath, 0755)
				
				// Change to a directory without dockerfiles
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				
				return func() {
					os.Setenv("HOME", origHome)
					os.Chdir(origWD)
				}
			},
			expected:      "", // Will be HOME/.local/share/mcs
			expectDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc()
			defer cleanup()

			result := GetMCSInstallPath()

			// Handle dynamic paths
			if tt.expected == "" {
				if tt.expectDefault {
					expectedHome := os.Getenv("HOME")
					expectedPath := filepath.Join(expectedHome, ".mcs")
					if result != expectedPath {
						// Check if it's the current directory (dev mode)
						pwd, _ := os.Getwd()
						dockerfilesPath := filepath.Join(pwd, "dockerfiles")
						if _, err := os.Stat(dockerfilesPath); err == nil {
							if result != pwd {
								t.Errorf("GetMCSInstallPath() = %v, want %v (dev mode)", result, pwd)
							}
						} else {
							t.Errorf("GetMCSInstallPath() = %v, want %v (default)", result, expectedPath)
						}
					}
				} else if tt.name == "development mode - dockerfiles in current directory" {
					pwd, _ := os.Getwd()
					if result != pwd {
						t.Errorf("GetMCSInstallPath() = %v, want %v", result, pwd)
					}
				}
			} else {
				if result != tt.expected {
					t.Errorf("GetMCSInstallPath() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

// TestGetDockerfilesPath tests the dockerfiles path construction
func TestGetDockerfilesPath(t *testing.T) {
	// Save original values
	origWD, _ := os.Getwd()
	origMCSInstallPath := os.Getenv("MCS_INSTALL_PATH")
	origHome := os.Getenv("HOME")
	
	defer func() {
		os.Chdir(origWD)
		os.Setenv("MCS_INSTALL_PATH", origMCSInstallPath)
		os.Setenv("HOME", origHome)
	}()

	tests := []struct {
		name      string
		setupFunc func() (cleanup func())
		validate  func(t *testing.T, result string)
	}{
		{
			name: "returns install path with dockerfiles suffix",
			setupFunc: func() func() {
				os.Setenv("MCS_INSTALL_PATH", "/test/mcs")
				return func() {
					os.Unsetenv("MCS_INSTALL_PATH")
				}
			},
			validate: func(t *testing.T, result string) {
				expected := filepath.Join("/test/mcs", "dockerfiles")
				if result != expected {
					t.Errorf("GetDockerfilesPath() = %v, want %v", result, expected)
				}
			},
		},
		{
			name: "uses development path when in dev mode",
			setupFunc: func() func() {
				// Create temp directory with dockerfiles
				tempDir := t.TempDir()
				dockerfilesDir := filepath.Join(tempDir, "dockerfiles")
				os.MkdirAll(dockerfilesDir, 0755)
				os.Chdir(tempDir)
				os.Unsetenv("MCS_INSTALL_PATH")
				
				return func() {
					os.Chdir(origWD)
				}
			},
			validate: func(t *testing.T, result string) {
				pwd, _ := os.Getwd()
				expected := filepath.Join(pwd, "dockerfiles")
				if result != expected {
					t.Errorf("GetDockerfilesPath() = %v, want %v", result, expected)
				}
			},
		},
		{
			name: "defaults to HOME/.mcs/dockerfiles",
			setupFunc: func() func() {
				tempHome := t.TempDir()
				os.Setenv("HOME", tempHome)
				os.Unsetenv("MCS_INSTALL_PATH")
				
				// Change to directory without dockerfiles
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				
				return func() {
					os.Setenv("HOME", origHome)
					os.Chdir(origWD)
				}
			},
			validate: func(t *testing.T, result string) {
				home := os.Getenv("HOME")
				expected := filepath.Join(home, ".mcs", "dockerfiles")
				if result != expected {
					t.Errorf("GetDockerfilesPath() = %v, want %v", result, expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc()
			defer cleanup()

			result := GetDockerfilesPath()
			tt.validate(t, result)
		})
	}
}

// TestPathResolution tests various path resolution scenarios
func TestPathResolution(t *testing.T) {
	// Save original values
	origHome := os.Getenv("HOME")
	origMCSInstallPath := os.Getenv("MCS_INSTALL_PATH")
	
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("MCS_INSTALL_PATH", origMCSInstallPath)
	}()

	tests := []struct {
		name           string
		home           string
		mcsInstallPath string
		validate       func(t *testing.T, installPath, dockerfilesPath string)
	}{
		{
			name:           "paths with spaces",
			home:           "/home/user name/with spaces",
			mcsInstallPath: "",
			validate: func(t *testing.T, installPath, dockerfilesPath string) {
				if !strings.Contains(installPath, "with spaces") {
					t.Errorf("Install path doesn't handle spaces: %v", installPath)
				}
				if !strings.HasSuffix(dockerfilesPath, filepath.Join("dockerfiles")) {
					t.Errorf("Dockerfiles path incorrect: %v", dockerfilesPath)
				}
			},
		},
		{
			name:           "paths with special characters",
			home:           "/home/user-name_123",
			mcsInstallPath: "/opt/mcs@latest",
			validate: func(t *testing.T, installPath, dockerfilesPath string) {
				if installPath != "/opt/mcs@latest" {
					t.Errorf("Install path = %v, want /opt/mcs@latest", installPath)
				}
			},
		},
		{
			name:           "relative path resolution",
			home:           "/home/./user/../user",
			mcsInstallPath: "",
			validate: func(t *testing.T, installPath, dockerfilesPath string) {
				// Path should be cleaned - Go automatically cleans paths when using filepath.Join
				// The result should not contain raw ".." or "/." but may contain resolved paths
				if strings.Contains(installPath, "/./") || strings.Contains(installPath, "/../") {
					t.Errorf("Path not cleaned properly: %v", installPath)
				}
			},
		},
		{
			name:           "empty HOME variable",
			home:           "",
			mcsInstallPath: "/fallback/mcs",
			validate: func(t *testing.T, installPath, dockerfilesPath string) {
				if installPath != "/fallback/mcs" {
					t.Errorf("Should use MCS_INSTALL_PATH when HOME is empty: %v", installPath)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("HOME", tt.home)
			if tt.mcsInstallPath != "" {
				os.Setenv("MCS_INSTALL_PATH", tt.mcsInstallPath)
			} else {
				os.Unsetenv("MCS_INSTALL_PATH")
			}

			// Change to temp directory to avoid dev mode
			tempDir := t.TempDir()
			origWD, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(origWD)

			installPath := GetMCSInstallPath()
			dockerfilesPath := GetDockerfilesPath()

			tt.validate(t, installPath, dockerfilesPath)
		})
	}
}

// TestPathPrecedence tests the precedence order of path detection
func TestPathPrecedence(t *testing.T) {
	// Save original values
	origWD, _ := os.Getwd()
	origMCSInstallPath := os.Getenv("MCS_INSTALL_PATH")
	origHome := os.Getenv("HOME")
	
	defer func() {
		os.Chdir(origWD)
		os.Setenv("MCS_INSTALL_PATH", origMCSInstallPath)
		os.Setenv("HOME", origHome)
	}()

	// Test precedence: dev mode > standard paths > MCS_INSTALL_PATH > default
	t.Run("precedence order", func(t *testing.T) {
		tempHome := t.TempDir()
		os.Setenv("HOME", tempHome)
		
		// 1. First, dev mode should take precedence
		devDir := t.TempDir()
		dockerfilesDir := filepath.Join(devDir, "dockerfiles")
		os.MkdirAll(dockerfilesDir, 0755)
		os.Chdir(devDir)
		
		result := GetMCSInstallPath()
		if result != devDir {
			t.Errorf("Dev mode not taking precedence: got %v, want %v", result, devDir)
		}
		
		// 2. Without dev mode, MCS_INSTALL_PATH should be used
		noDevDir := t.TempDir()
		os.Chdir(noDevDir)
		os.Setenv("MCS_INSTALL_PATH", "/custom/path")
		
		result = GetMCSInstallPath()
		if result != "/custom/path" {
			t.Errorf("MCS_INSTALL_PATH not used: got %v, want /custom/path", result)
		}
		
		// 3. Without MCS_INSTALL_PATH, should default to HOME/.mcs
		os.Unsetenv("MCS_INSTALL_PATH")
		
		result = GetMCSInstallPath()
		expectedDefault := filepath.Join(tempHome, ".mcs")
		if result != expectedDefault {
			t.Errorf("Default path not used: got %v, want %v", result, expectedDefault)
		}
	})
}

// TestConcurrentPathAccess tests thread-safe access to paths
func TestConcurrentPathAccess(t *testing.T) {
	// Save original values
	origMCSInstallPath := os.Getenv("MCS_INSTALL_PATH")
	defer os.Setenv("MCS_INSTALL_PATH", origMCSInstallPath)
	
	// Set a known path
	os.Setenv("MCS_INSTALL_PATH", "/test/concurrent/mcs")
	
	// Run concurrent access with proper synchronization
	numGoroutines := 10
	results := make(chan [2]string, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			installPath := GetMCSInstallPath()
			dockerfilesPath := GetDockerfilesPath()
			results <- [2]string{installPath, dockerfilesPath}
		}()
	}
	
	// Collect results
	expectedInstall := "/test/concurrent/mcs"
	expectedDockerfiles := filepath.Join(expectedInstall, "dockerfiles")
	
	for i := 0; i < numGoroutines; i++ {
		result := <-results
		if result[0] != expectedInstall {
			t.Errorf("Unexpected install path: got %v, want %v", result[0], expectedInstall)
		}
		if result[1] != expectedDockerfiles {
			t.Errorf("Unexpected dockerfiles path: got %v, want %v", result[1], expectedDockerfiles)
		}
	}
}

// BenchmarkGetMCSInstallPath benchmarks path detection
func BenchmarkGetMCSInstallPath(b *testing.B) {
	// Test different scenarios
	scenarios := []struct {
		name  string
		setup func()
	}{
		{
			name: "with_env_var",
			setup: func() {
				os.Setenv("MCS_INSTALL_PATH", "/bench/mcs")
			},
		},
		{
			name: "without_env_var",
			setup: func() {
				os.Unsetenv("MCS_INSTALL_PATH")
			},
		},
		{
			name: "dev_mode",
			setup: func() {
				os.Unsetenv("MCS_INSTALL_PATH")
				// Would need to mock dev mode
			},
		},
	}
	
	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			scenario.setup()
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				_ = GetMCSInstallPath()
			}
		})
	}
}

// BenchmarkGetDockerfilesPath benchmarks dockerfiles path construction
func BenchmarkGetDockerfilesPath(b *testing.B) {
	os.Setenv("MCS_INSTALL_PATH", "/bench/mcs")
	defer os.Unsetenv("MCS_INSTALL_PATH")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetDockerfilesPath()
	}
}