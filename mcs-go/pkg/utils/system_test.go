package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageManager_Constants(t *testing.T) {
	assert.Equal(t, PackageManager("apt"), APT)
	assert.Equal(t, PackageManager("yum"), YUM)
	assert.Equal(t, PackageManager("dnf"), DNF)
	assert.Equal(t, PackageManager("brew"), BREW)
	assert.Equal(t, PackageManager("none"), NONE)
}

func TestPlatform_Structure(t *testing.T) {
	p := Platform{
		OS:   "linux",
		Arch: "amd64",
		Name: "linux-amd64",
	}
	
	assert.Equal(t, "linux", p.OS)
	assert.Equal(t, "amd64", p.Arch)
	assert.Equal(t, "linux-amd64", p.Name)
}

func TestGetHomeDir(t *testing.T) {
	home := GetHomeDir()
	assert.NotEmpty(t, home)
	
	// Should be an absolute path
	assert.True(t, filepath.IsAbs(home))
	
	// Directory should exist
	info, err := os.Stat(home)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGetHomeDir_Fallback(t *testing.T) {
	// Test fallback behavior by temporarily unsetting HOME
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")
	
	// Clear environment variables
	os.Unsetenv("HOME")
	os.Unsetenv("USERPROFILE")
	
	// Mock os.UserHomeDir to return an error
	defer func() {
		os.Setenv("HOME", originalHome)
		if originalUserProfile != "" {
			os.Setenv("USERPROFILE", originalUserProfile)
		}
	}()
	
	home := GetHomeDir()
	assert.NotEmpty(t, home)
	// In worst case, should fallback to /tmp or be able to determine some directory
}

func TestGetMCSDir(t *testing.T) {
	mcsDir := GetMCSDir()
	assert.NotEmpty(t, mcsDir)
	
	// Should be under home directory
	home := GetHomeDir()
	assert.True(t, strings.HasPrefix(mcsDir, home))
	
	// Should end with .mcs
	assert.True(t, strings.HasSuffix(mcsDir, ".mcs"))
	
	// Should be an absolute path
	assert.True(t, filepath.IsAbs(mcsDir))
}

func TestEnsureDir(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test", "nested", "directory")
	
	// Directory shouldn't exist initially
	_, err := os.Stat(testDir)
	assert.True(t, os.IsNotExist(err))
	
	// Create it
	err = EnsureDir(testDir)
	assert.NoError(t, err)
	
	// Should exist now
	info, err := os.Stat(testDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
	
	// Should not error if called again
	err = EnsureDir(testDir)
	assert.NoError(t, err)
}

func TestDetectPackageManager(t *testing.T) {
	pm := DetectPackageManager()
	
	// Should return a valid package manager type
	validPMs := []PackageManager{APT, YUM, DNF, BREW, NONE}
	found := false
	for _, validPM := range validPMs {
		if pm == validPM {
			found = true
			break
		}
	}
	assert.True(t, found, "Should return a valid package manager")
	
	// On macOS, should prefer BREW if available
	if runtime.GOOS == "darwin" {
		if CommandExists("brew") {
			assert.Equal(t, BREW, pm)
		} else {
			assert.Equal(t, NONE, pm)
		}
	}
	
	// On Linux, should detect appropriate package manager
	if runtime.GOOS == "linux" {
		// At least one of these should be true based on the detection logic
		if pm != NONE {
			assert.Contains(t, []PackageManager{APT, DNF, YUM}, pm)
		}
	}
}

func TestGetPlatform(t *testing.T) {
	platform := GetPlatform()
	
	// OS should match runtime
	assert.Equal(t, runtime.GOOS, platform.OS)
	
	// Architecture should be normalized
	expectedArch := runtime.GOARCH
	if expectedArch == "x86_64" {
		expectedArch = "amd64"
	} else if expectedArch == "aarch64" {
		expectedArch = "arm64"
	}
	assert.Equal(t, expectedArch, platform.Arch)
	
	// Name should be formatted correctly
	expectedName := platform.OS + "-" + platform.Arch
	assert.Equal(t, expectedName, platform.Name)
	
	// Should be non-empty
	assert.NotEmpty(t, platform.OS)
	assert.NotEmpty(t, platform.Arch)
	assert.NotEmpty(t, platform.Name)
}

func TestGetPlatform_ArchNormalization(t *testing.T) {
	platform := GetPlatform()
	
	// Make sure we don't have the old architecture names
	assert.NotEqual(t, "x86_64", platform.Arch)
	assert.NotEqual(t, "aarch64", platform.Arch)
	
	// Should be one of the expected values
	validArchs := []string{"amd64", "arm64", "386", "arm"}
	assert.Contains(t, validArchs, platform.Arch)
}

func TestIsRoot(t *testing.T) {
	isRoot := IsRoot()
	
	// On Windows, should always return false for this simple implementation
	if runtime.GOOS == "windows" {
		assert.False(t, isRoot)
	} else {
		// On Unix-like systems, test current user
		// In most test environments, we're not root
		actualUID := os.Geteuid()
		expectedRoot := (actualUID == 0)
		assert.Equal(t, expectedRoot, isRoot)
	}
}

func TestRunCommand(t *testing.T) {
	// Test with a simple command that should work everywhere
	var cmd string
	var expectedSubstring string
	
	if runtime.GOOS == "windows" {
		cmd = "echo"
		expectedSubstring = "hello"
	} else {
		cmd = "echo"
		expectedSubstring = "hello"
	}
	
	output, err := RunCommand(cmd, "hello")
	assert.NoError(t, err)
	assert.Contains(t, strings.ToLower(output), expectedSubstring)
}

func TestRunCommand_Error(t *testing.T) {
	// Test with a command that doesn't exist
	_, err := RunCommand("nonexistent-command-12345")
	assert.Error(t, err)
}

func TestRunCommandWithEnv(t *testing.T) {
	// Test with custom environment variable
	env := []string{"TEST_VAR=test_value"}
	
	var cmd string
	var args []string
	
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo", "%TEST_VAR%"}
	} else {
		cmd = "sh"
		args = []string{"-c", "echo $TEST_VAR"}
	}
	
	output, err := RunCommandWithEnv(env, cmd, args...)
	assert.NoError(t, err)
	assert.Contains(t, output, "test_value")
}

func TestCommandExists(t *testing.T) {
	// Test with a command that should exist
	var existingCmd string
	if runtime.GOOS == "windows" {
		existingCmd = "cmd"
	} else {
		existingCmd = "sh"
	}
	
	assert.True(t, CommandExists(existingCmd))
	
	// Test with a command that shouldn't exist
	assert.False(t, CommandExists("nonexistent-command-12345"))
}

func TestGetEnvOrDefault(t *testing.T) {
	// Test with existing environment variable
	key := "PATH"
	value := GetEnvOrDefault(key, "default")
	assert.NotEqual(t, "default", value)
	assert.NotEmpty(t, value)
	
	// Test with non-existing environment variable
	key = "NONEXISTENT_ENV_VAR_12345"
	value = GetEnvOrDefault(key, "default_value")
	assert.Equal(t, "default_value", value)
	
	// Test with empty environment variable
	originalValue := os.Getenv("TEST_EMPTY_VAR")
	os.Setenv("TEST_EMPTY_VAR", "")
	defer func() {
		if originalValue != "" {
			os.Setenv("TEST_EMPTY_VAR", originalValue)
		} else {
			os.Unsetenv("TEST_EMPTY_VAR")
		}
	}()
	
	value = GetEnvOrDefault("TEST_EMPTY_VAR", "default_for_empty")
	assert.Equal(t, "default_for_empty", value)
}

func TestIsWSL(t *testing.T) {
	isWSL := IsWSL()
	
	// On non-Linux systems, should always be false
	if runtime.GOOS != "linux" {
		assert.False(t, isWSL)
		return
	}
	
	// On Linux, check if we can detect WSL correctly
	// This test is environment-dependent, so we mainly test that it doesn't panic
	assert.IsType(t, false, isWSL) // Just checking it returns a boolean
}

func TestGetSystemInfo(t *testing.T) {
	info := GetSystemInfo()
	
	// Should contain basic system information
	assert.Equal(t, runtime.GOOS, info["os"])
	assert.Equal(t, runtime.GOARCH, info["arch"])
	assert.Equal(t, runtime.Version(), info["go_version"])
	assert.Contains(t, info, "num_cpu")
	
	// num_cpu should be a valid number
	assert.NotEmpty(t, info["num_cpu"])
	
	// Should include hostname if available
	if hostname, err := os.Hostname(); err == nil {
		assert.Equal(t, hostname, info["hostname"])
	}
	
	// If running in WSL, should include environment info
	if IsWSL() {
		assert.Equal(t, "wsl", info["environment"])
	}
}

func TestRequireSudo(t *testing.T) {
	testArgs := []string{"some", "command", "args"}
	
	if runtime.GOOS == "windows" {
		// On Windows, should return args unchanged
		result := RequireSudo(testArgs)
		assert.Equal(t, testArgs, result)
	} else {
		result := RequireSudo(testArgs)
		
		if IsRoot() {
			// If already root, should return unchanged
			assert.Equal(t, testArgs, result)
		} else if CommandExists("sudo") {
			// If not root and sudo exists, should prepend sudo
			expected := append([]string{"sudo"}, testArgs...)
			assert.Equal(t, expected, result)
		} else {
			// If sudo doesn't exist, should return unchanged
			assert.Equal(t, testArgs, result)
		}
	}
}

func TestEnsureExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		// On Windows, should be a no-op
		err := EnsureExecutable("any-path")
		assert.NoError(t, err)
		return
	}
	
	// On Unix-like systems, test actual functionality
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "testfile")
	
	// Create a file without executable permissions
	err := os.WriteFile(testFile, []byte("#!/bin/sh\necho hello"), 0644)
	require.NoError(t, err)
	
	// Initially shouldn't be executable
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0), info.Mode()&0111)
	
	// Make it executable
	err = EnsureExecutable(testFile)
	assert.NoError(t, err)
	
	// Should now be executable
	info, err = os.Stat(testFile)
	require.NoError(t, err)
	assert.NotEqual(t, os.FileMode(0), info.Mode()&0100) // At least user executable
}

func TestEnsureExecutable_AlreadyExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		return // Skip on Windows
	}
	
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "testfile")
	
	// Create a file with executable permissions
	err := os.WriteFile(testFile, []byte("#!/bin/sh\necho hello"), 0755)
	require.NoError(t, err)
	
	// Should not error if already executable
	err = EnsureExecutable(testFile)
	assert.NoError(t, err)
}

func TestEnsureExecutable_NonexistentFile(t *testing.T) {
	err := EnsureExecutable("/nonexistent/file")
	assert.Error(t, err)
}

// Integration test for system operations
func TestSystemIntegration(t *testing.T) {
	// Test the complete workflow of system operations
	
	// 1. Get platform information
	platform := GetPlatform()
	assert.NotEmpty(t, platform.OS)
	assert.NotEmpty(t, platform.Arch)
	
	// 2. Get system info
	sysInfo := GetSystemInfo()
	assert.Equal(t, platform.OS, sysInfo["os"])
	assert.Equal(t, platform.Arch, sysInfo["arch"])
	
	// 3. Test directory operations
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "integration-test")
	
	err := EnsureDir(testDir)
	assert.NoError(t, err)
	_, err = os.Stat(testDir)
	assert.NoError(t, err)
	
	// 4. Test command execution
	if CommandExists("echo") {
		output, err := RunCommand("echo", "integration test")
		assert.NoError(t, err)
		assert.Contains(t, output, "integration test")
	}
	
	// 5. Test environment handling
	testValue := GetEnvOrDefault("NONEXISTENT_VAR", "default")
	assert.Equal(t, "default", testValue)
	
	// 6. Test package manager detection
	pm := DetectPackageManager()
	assert.NotEqual(t, PackageManager(""), pm) // Should return something
}

// Benchmark tests for performance-critical functions
func BenchmarkGetPlatform(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetPlatform()
	}
}

func BenchmarkGetSystemInfo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetSystemInfo()
	}
}

func BenchmarkCommandExists(b *testing.B) {
	commands := []string{"sh", "echo", "nonexistent", "ls"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, cmd := range commands {
			CommandExists(cmd)
		}
	}
}

func BenchmarkDetectPackageManager(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectPackageManager()
	}
}

// Test helper functions for edge cases
func TestSystemEdgeCases(t *testing.T) {
	// Test with empty command
	_, err := RunCommand("")
	assert.Error(t, err)
	
	// Test RequireSudo with nil args
	result := RequireSudo(nil)
	if runtime.GOOS == "windows" || IsRoot() || !CommandExists("sudo") {
		assert.Nil(t, result)
	} else {
		assert.Equal(t, []string{"sudo"}, result)
	}
	
	// Test GetEnvOrDefault with empty key
	value := GetEnvOrDefault("", "default")
	assert.Equal(t, "default", value)
}