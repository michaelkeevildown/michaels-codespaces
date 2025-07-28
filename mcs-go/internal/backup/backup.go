package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupType represents the type of backup being created
type BackupType string

const (
	BackupTypeDestroy BackupType = "destroy"
	BackupTypeInstall BackupType = "install"
	BackupTypeManual  BackupType = "manual"
)

// BackupInfo contains metadata about a backup
type BackupInfo struct {
	ID          string    `json:"id"`
	Type        BackupType `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	SourcePath  string    `json:"source_path"`
	Size        int64     `json:"size"`
	Description string    `json:"description,omitempty"`
}

// BackupManager handles all backup operations for MCS
type BackupManager struct {
	baseDir string
}

// NewBackupManager creates a new backup manager instance
func NewBackupManager() *BackupManager {
	homeDir := os.Getenv("HOME")
	return &BackupManager{
		baseDir: filepath.Join(homeDir, ".mcs.backup"),
	}
}

// Create creates a new backup of the specified directory
func (bm *BackupManager) Create(sourceDir string, backupType BackupType, description string) (string, error) {
	// Check if source directory exists
	sourceInfo, err := os.Stat(sourceDir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("source directory does not exist: %s", sourceDir)
	}
	if err != nil {
		return "", fmt.Errorf("failed to stat source directory: %w", err)
	}
	if !sourceInfo.IsDir() {
		return "", fmt.Errorf("source is not a directory: %s", sourceDir)
	}

	// Create base backup directory if it doesn't exist
	if err := os.MkdirAll(bm.baseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup ID with type prefix and microseconds for uniqueness
	timestamp := time.Now().Format("20060102_150405.000000")
	// Remove dots from timestamp for cleaner filenames
	timestamp = strings.ReplaceAll(timestamp, ".", "")
	backupID := fmt.Sprintf("%s-%s", backupType, timestamp)
	backupPath := filepath.Join(bm.baseDir, backupID)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup path: %w", err)
	}

	// Copy source directory to backup location
	sourceName := filepath.Base(sourceDir)
	destPath := filepath.Join(backupPath, sourceName)
	
	cmd := exec.Command("cp", "-r", sourceDir, destPath)
	if err := cmd.Run(); err != nil {
		// Clean up on failure
		os.RemoveAll(backupPath)
		return "", fmt.Errorf("failed to copy files: %w", err)
	}

	// Calculate backup size
	size, err := bm.calculateDirSize(backupPath)
	if err != nil {
		// Non-fatal error, continue
		size = 0
	}

	// Create metadata
	metadata := BackupInfo{
		ID:          backupID,
		Type:        backupType,
		Timestamp:   time.Now(),
		SourcePath:  sourceDir,
		Size:        size,
		Description: description,
	}

	// Save metadata
	metadataPath := filepath.Join(backupPath, "metadata.json")
	metadataFile, err := os.Create(metadataPath)
	if err != nil {
		return backupID, fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer metadataFile.Close()

	encoder := json.NewEncoder(metadataFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(metadata); err != nil {
		return backupID, fmt.Errorf("failed to write metadata: %w", err)
	}

	return backupID, nil
}

// List returns all available backups sorted by timestamp (newest first)
func (bm *BackupManager) List() ([]BackupInfo, error) {
	// Ensure backup directory exists
	if _, err := os.Stat(bm.baseDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil // No backups yet
	}

	entries, err := os.ReadDir(bm.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Read metadata
		metadataPath := filepath.Join(bm.baseDir, entry.Name(), "metadata.json")
		metadataFile, err := os.Open(metadataPath)
		if err != nil {
			// Skip backups without metadata
			continue
		}
		defer metadataFile.Close()

		var info BackupInfo
		decoder := json.NewDecoder(metadataFile)
		if err := decoder.Decode(&info); err != nil {
			metadataFile.Close()
			continue
		}
		metadataFile.Close()

		backups = append(backups, info)
	}

	// Sort by timestamp (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// Restore restores a backup to the specified target directory
func (bm *BackupManager) Restore(backupID string, targetDir string) error {
	backupPath := filepath.Join(bm.baseDir, backupID)
	
	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// Read metadata to get original structure
	metadataPath := filepath.Join(backupPath, "metadata.json")
	metadataFile, err := os.Open(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read backup metadata: %w", err)
	}
	defer metadataFile.Close()

	var info BackupInfo
	decoder := json.NewDecoder(metadataFile)
	if err := decoder.Decode(&info); err != nil {
		return fmt.Errorf("failed to parse backup metadata: %w", err)
	}

	// Find the backed up content (excluding metadata.json)
	entries, err := os.ReadDir(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() == "metadata.json" {
			continue
		}

		sourcePath := filepath.Join(backupPath, entry.Name())
		destPath := filepath.Join(targetDir, entry.Name())

		// Check if destination already exists
		if _, err := os.Stat(destPath); err == nil {
			return fmt.Errorf("destination already exists: %s", destPath)
		}

		// Copy the backup to target
		cmd := exec.Command("cp", "-r", sourcePath, destPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}
	}

	return nil
}

// Delete removes a backup
func (bm *BackupManager) Delete(backupID string) error {
	backupPath := filepath.Join(bm.baseDir, backupID)
	
	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// Remove the backup directory
	if err := os.RemoveAll(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	return nil
}

// CleanupOld removes old backups, keeping only the specified number of most recent backups
func (bm *BackupManager) CleanupOld(keepCount int) error {
	if keepCount < 0 {
		return fmt.Errorf("keepCount must be non-negative")
	}

	backups, err := bm.List()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	// If we have fewer backups than keepCount, nothing to do
	if len(backups) <= keepCount {
		return nil
	}

	// Backups are already sorted by timestamp (newest first)
	// Delete backups beyond keepCount
	for i := keepCount; i < len(backups); i++ {
		if err := bm.Delete(backups[i].ID); err != nil {
			// Log error but continue with other deletions
			fmt.Fprintf(os.Stderr, "Warning: failed to delete backup %s: %v\n", backups[i].ID, err)
		}
	}

	return nil
}

// GetBackupPath returns the full path to a backup
func (bm *BackupManager) GetBackupPath(backupID string) string {
	return filepath.Join(bm.baseDir, backupID)
}

// calculateDirSize calculates the total size of a directory
func (bm *BackupManager) calculateDirSize(path string) (int64, error) {
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

// FormatSize formats bytes into human-readable format
func FormatSize(bytes int64) string {
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

// CreateQuick is a convenience method for quick backups without description
func (bm *BackupManager) CreateQuick(sourceDir string, backupType BackupType) (string, error) {
	return bm.Create(sourceDir, backupType, "")
}

// GetLatestBackup returns the most recent backup of a specific type
func (bm *BackupManager) GetLatestBackup(backupType BackupType) (*BackupInfo, error) {
	backups, err := bm.List()
	if err != nil {
		return nil, err
	}

	for _, backup := range backups {
		if backup.Type == backupType {
			return &backup, nil
		}
	}

	return nil, fmt.Errorf("no backup found for type: %s", backupType)
}

// Exists checks if a backup exists
func (bm *BackupManager) Exists(backupID string) bool {
	backupPath := filepath.Join(bm.baseDir, backupID)
	_, err := os.Stat(backupPath)
	return err == nil
}

// CopyFile is a utility function to copy a single file
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// GetBackupTypeFromID extracts the backup type from a backup ID
func GetBackupTypeFromID(backupID string) BackupType {
	parts := strings.SplitN(backupID, "-", 2)
	if len(parts) > 0 {
		switch parts[0] {
		case string(BackupTypeDestroy):
			return BackupTypeDestroy
		case string(BackupTypeInstall):
			return BackupTypeInstall
		case string(BackupTypeManual):
			return BackupTypeManual
		}
	}
	return BackupTypeManual // Default
}