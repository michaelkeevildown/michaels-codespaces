package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// PackageManager represents different package managers
type PackageManager string

const (
	APT  PackageManager = "apt"
	YUM  PackageManager = "yum"
	DNF  PackageManager = "dnf"
	BREW PackageManager = "brew"
	NONE PackageManager = "none"
)

// Platform represents OS and architecture information
type Platform struct {
	OS   string
	Arch string
	Name string // Human-readable name like "linux-amd64"
}

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to HOME environment variable
		home = os.Getenv("HOME")
		if home == "" && runtime.GOOS == "windows" {
			home = os.Getenv("USERPROFILE")
		}
		if home == "" {
			home = "/tmp" // Last resort
		}
	}
	return home
}

// GetMCSDir returns the MCS configuration directory
func GetMCSDir() string {
	return filepath.Join(GetHomeDir(), ".mcs")
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// DetectPackageManager detects the available package manager on the system
func DetectPackageManager() PackageManager {
	// Check for macOS first
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("brew"); err == nil {
			return BREW
		}
		return NONE
	}
	
	// Check for Linux package managers
	if runtime.GOOS == "linux" {
		// Check in order of preference
		managers := []struct {
			cmd string
			pm  PackageManager
		}{
			{"apt-get", APT},
			{"dnf", DNF},
			{"yum", YUM},
		}
		
		for _, mgr := range managers {
			if _, err := exec.LookPath(mgr.cmd); err == nil {
				return mgr.pm
			}
		}
	}
	
	return NONE
}

// GetPlatform returns the current platform information
func GetPlatform() Platform {
	os := runtime.GOOS
	arch := runtime.GOARCH
	
	// Normalize architecture names
	switch arch {
	case "x86_64":
		arch = "amd64"
	case "aarch64":
		arch = "arm64"
	}
	
	name := fmt.Sprintf("%s-%s", os, arch)
	
	return Platform{
		OS:   os,
		Arch: arch,
		Name: name,
	}
}

// IsRoot checks if the current process is running as root/administrator
func IsRoot() bool {
	if runtime.GOOS == "windows" {
		// On Windows, check if running as administrator
		// This is a simplified check
		return false
	}
	
	// On Unix-like systems, check if UID is 0
	return os.Geteuid() == 0
}

// RunCommand runs a command and returns the output
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunCommandWithEnv runs a command with custom environment variables
func RunCommandWithEnv(env []string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// CommandExists checks if a command exists in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetEnvOrDefault returns an environment variable value or a default
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsWSL detects if running inside Windows Subsystem for Linux
func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	
	// Check for WSL-specific files or environment
	if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
		return true
	}
	
	// Check kernel version for Microsoft
	if data, err := os.ReadFile("/proc/version"); err == nil {
		return strings.Contains(strings.ToLower(string(data)), "microsoft")
	}
	
	return false
}

// GetSystemInfo returns basic system information
func GetSystemInfo() map[string]string {
	info := make(map[string]string)
	
	info["os"] = runtime.GOOS
	info["arch"] = runtime.GOARCH
	info["go_version"] = runtime.Version()
	info["num_cpu"] = fmt.Sprintf("%d", runtime.NumCPU())
	
	if hostname, err := os.Hostname(); err == nil {
		info["hostname"] = hostname
	}
	
	if IsWSL() {
		info["environment"] = "wsl"
	}
	
	return info
}

// RequireSudo checks if a command requires sudo and prepends it if needed
func RequireSudo(args []string) []string {
	if IsRoot() {
		return args // Already root, no sudo needed
	}
	
	if runtime.GOOS == "windows" {
		return args // No sudo on Windows
	}
	
	// Check if sudo is available
	if !CommandExists("sudo") {
		return args // No sudo available
	}
	
	// Prepend sudo
	return append([]string{"sudo"}, args...)
}

// EnsureExecutable ensures a file has executable permissions
func EnsureExecutable(path string) error {
	if runtime.GOOS == "windows" {
		return nil // Windows doesn't use Unix permissions
	}
	
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	
	mode := info.Mode()
	if mode&0111 == 0 {
		// Add executable permission for user
		return os.Chmod(path, mode|0100)
	}
	
	return nil
}