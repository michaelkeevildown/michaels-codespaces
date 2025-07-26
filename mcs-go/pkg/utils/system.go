package utils

import (
	"os"
	"path/filepath"
)

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to HOME environment variable
		home = os.Getenv("HOME")
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