package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelkeevildown/mcs/internal/ui"
)

// Common styles for filesystem operations
var (
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CalculateDirSize calculates the total size of a directory
func CalculateDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// RemovePathsWithProgress removes files or directories with progress indication
func RemovePathsWithProgress(paths []string, prefix string) (int, int) {
	successCount := 0
	failCount := 0
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf(dimStyle.Render("  %s %s..."), prefix, path)
			
			// Determine if it's a file or directory
			info, _ := os.Stat(path)
			var err error
			if info.IsDir() {
				err = os.RemoveAll(path)
			} else {
				err = os.Remove(path)
			}
			
			if err != nil {
				fmt.Printf(" %s\n", errorStyle.Render("failed"))
				failCount++
			} else {
				fmt.Printf(" %s\n", successStyle.Render("✓"))
				successCount++
			}
		}
	}
	
	return successCount, failCount
}

// RemoveDirectoryWithSize removes a directory and reports its size
func RemoveDirectoryWithSize(dirPath string, progress *ui.Progress) error {
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to do
	}
	
	// Calculate directory size
	size, err := CalculateDirSize(dirPath)
	if err != nil {
		// Non-fatal, continue with removal
		size = 0
	}
	
	// Format size for display
	sizeStr := FormatBytes(size)
	fmt.Printf(dimStyle.Render("Removing directory (%s)...\n"), sizeStr)
	
	// Update progress
	if progress != nil {
		progress.Update(fmt.Sprintf("Removing %s", filepath.Base(dirPath)))
	}
	
	// Remove directory
	if err := os.RemoveAll(dirPath); err != nil {
		if progress != nil {
			progress.Fail(fmt.Sprintf("Failed to remove %s", filepath.Base(dirPath)))
		}
		return fmt.Errorf("failed to remove %s: %w", dirPath, err)
	}
	
	fmt.Printf(successStyle.Render("✓ Removed %s\n"), dirPath)
	return nil
}

// EnsureDirectory creates a directory if it doesn't exist
func EnsureDirectory(path string, perm os.FileMode) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, perm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	return nil
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetFileSize gets the size of a file or directory
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	
	if info.IsDir() {
		return CalculateDirSize(path)
	}
	
	return info.Size(), nil
}

// CleanPath cleans and expands a file path
func CleanPath(path string) string {
	// Expand home directory
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[1:])
		}
	}
	
	// Clean the path
	return filepath.Clean(path)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	// Get source file info for permissions
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	return os.WriteFile(dst, input, info.Mode())
}

// ListDirectorySize lists all items in a directory with their sizes
func ListDirectorySize(dirPath string) (map[string]int64, error) {
	sizes := make(map[string]int64)
	
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	
	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		size, err := GetFileSize(fullPath)
		if err != nil {
			continue // Skip items we can't size
		}
		sizes[entry.Name()] = size
	}
	
	return sizes, nil
}