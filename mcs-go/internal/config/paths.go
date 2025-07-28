package config

import (
	"os"
	"path/filepath"
)

// GetMCSInstallPath returns the path where MCS is installed
// This is used to locate dockerfiles and other resources
func GetMCSInstallPath() string {
	// First, check if we're running from the source directory (development)
	if _, err := os.Stat("dockerfiles"); err == nil {
		pwd, _ := os.Getwd()
		return pwd
	}
	
	// Check common installation paths
	paths := []string{
		"/usr/local/share/mcs",
		"/opt/mcs",
		filepath.Join(os.Getenv("HOME"), ".mcs"),
		filepath.Join(os.Getenv("HOME"), ".local/share/mcs"),
	}
	
	for _, path := range paths {
		dockerfilesPath := filepath.Join(path, "dockerfiles")
		if _, err := os.Stat(dockerfilesPath); err == nil {
			return path
		}
	}
	
	// If MCS_INSTALL_PATH is set, use that
	if installPath := os.Getenv("MCS_INSTALL_PATH"); installPath != "" {
		return installPath
	}
	
	// Default to ~/.mcs
	return filepath.Join(os.Getenv("HOME"), ".mcs")
}

// GetDockerfilesPath returns the path to the dockerfiles directory
func GetDockerfilesPath() string {
	return filepath.Join(GetMCSInstallPath(), "dockerfiles")
}