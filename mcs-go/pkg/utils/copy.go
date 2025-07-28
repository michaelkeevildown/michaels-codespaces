package utils

import (
	"os"
	"path/filepath"
)

// CopyDirectory copies a directory recursively
func CopyDirectory(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	
	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	
	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	
	return nil
}

