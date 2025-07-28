package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBackupManager(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	testSourceDir := filepath.Join(tempDir, "test-source")
	
	// Create test source directory with some content
	if err := os.MkdirAll(filepath.Join(testSourceDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create test files
	testFile1 := filepath.Join(testSourceDir, "file1.txt")
	testFile2 := filepath.Join(testSourceDir, "subdir", "file2.txt")
	
	if err := os.WriteFile(testFile1, []byte("test content 1"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("test content 2"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create backup manager with custom base directory
	bm := &BackupManager{
		baseDir: filepath.Join(tempDir, ".mcs.backup"),
	}
	
	t.Run("Create Backup", func(t *testing.T) {
		backupID, err := bm.Create(testSourceDir, BackupTypeManual, "Test backup")
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}
		
		// Check backup ID format
		if len(backupID) < 10 {
			t.Errorf("Invalid backup ID format: %s", backupID)
		}
		
		// Check if backup was created
		if !bm.Exists(backupID) {
			t.Errorf("Backup does not exist after creation: %s", backupID)
		}
		
		// Check if metadata exists
		metadataPath := filepath.Join(bm.GetBackupPath(backupID), "metadata.json")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			t.Errorf("Metadata file not created")
		}
	})
	
	t.Run("List Backups", func(t *testing.T) {
		// Create another backup
		backupID2, err := bm.Create(testSourceDir, BackupTypeDestroy, "Another test backup")
		if err != nil {
			t.Fatalf("Failed to create second backup: %v", err)
		}
		
		// Wait a bit to ensure different timestamps
		time.Sleep(100 * time.Millisecond)
		
		// List backups
		backups, err := bm.List()
		if err != nil {
			t.Fatalf("Failed to list backups: %v", err)
		}
		
		if len(backups) < 2 {
			t.Errorf("Expected at least 2 backups, got %d", len(backups))
		}
		
		// Check if backups are sorted by timestamp (newest first)
		if len(backups) >= 2 {
			if backups[0].Timestamp.Before(backups[1].Timestamp) {
				t.Errorf("Backups not sorted correctly by timestamp")
			}
		}
		
		// Check backup types
		foundManual := false
		foundDestroy := false
		for _, b := range backups {
			if b.Type == BackupTypeManual {
				foundManual = true
			}
			if b.Type == BackupTypeDestroy {
				foundDestroy = true
			}
		}
		
		if !foundManual || !foundDestroy {
			t.Errorf("Not all backup types found in list")
		}
		
		// Store backup ID for restore test
		t.Logf("Created backup ID for restore test: %s", backupID2)
	})
	
	t.Run("Restore Backup", func(t *testing.T) {
		// Get latest backup
		backupInfo, err := bm.GetLatestBackup(BackupTypeDestroy)
		if err != nil {
			t.Fatalf("Failed to get latest backup: %v", err)
		}
		
		// Create restore target directory
		restoreDir := filepath.Join(tempDir, "restored")
		if err := os.MkdirAll(restoreDir, 0755); err != nil {
			t.Fatalf("Failed to create restore directory: %v", err)
		}
		
		// Restore backup
		if err := bm.Restore(backupInfo.ID, restoreDir); err != nil {
			t.Fatalf("Failed to restore backup: %v", err)
		}
		
		// Check if files were restored
		restoredFile1 := filepath.Join(restoreDir, "test-source", "file1.txt")
		restoredFile2 := filepath.Join(restoreDir, "test-source", "subdir", "file2.txt")
		
		if _, err := os.Stat(restoredFile1); os.IsNotExist(err) {
			t.Errorf("File1 not restored")
		}
		if _, err := os.Stat(restoredFile2); os.IsNotExist(err) {
			t.Errorf("File2 not restored")
		}
		
		// Check content
		content1, err := os.ReadFile(restoredFile1)
		if err != nil || string(content1) != "test content 1" {
			t.Errorf("Restored file1 has incorrect content")
		}
		
		content2, err := os.ReadFile(restoredFile2)
		if err != nil || string(content2) != "test content 2" {
			t.Errorf("Restored file2 has incorrect content")
		}
	})
	
	t.Run("Delete Backup", func(t *testing.T) {
		// Get a backup to delete
		backups, _ := bm.List()
		if len(backups) == 0 {
			t.Skip("No backups to delete")
		}
		
		backupToDelete := backups[0].ID
		
		// Delete the backup
		if err := bm.Delete(backupToDelete); err != nil {
			t.Fatalf("Failed to delete backup: %v", err)
		}
		
		// Check if backup still exists
		if bm.Exists(backupToDelete) {
			t.Errorf("Backup still exists after deletion")
		}
		
		// Check list doesn't include deleted backup
		newBackups, _ := bm.List()
		for _, b := range newBackups {
			if b.ID == backupToDelete {
				t.Errorf("Deleted backup still appears in list")
			}
		}
	})
	
	t.Run("Cleanup Old Backups", func(t *testing.T) {
		// Create a fresh backup manager for this test
		cleanupBm := &BackupManager{
			baseDir: filepath.Join(tempDir, ".mcs.backup.cleanup-test"),
		}
		
		// Check if test source still exists
		if _, err := os.Stat(testSourceDir); os.IsNotExist(err) {
			t.Logf("Test source directory doesn't exist, recreating...")
			// Recreate test source directory
			if err := os.MkdirAll(filepath.Join(testSourceDir, "subdir"), 0755); err != nil {
				t.Fatalf("Failed to recreate test directory: %v", err)
			}
			testFile := filepath.Join(testSourceDir, "test.txt")
			if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}
		
		// Create several backups
		for i := 0; i < 5; i++ {
			time.Sleep(10 * time.Millisecond) // Ensure different timestamps
			backupID, err := cleanupBm.CreateQuick(testSourceDir, BackupTypeManual)
			if err != nil {
				t.Fatalf("Failed to create backup %d: %v", i, err)
			}
			t.Logf("Created backup %d: %s", i, backupID)
		}
		
		// List backups before cleanup
		beforeBackups, _ := cleanupBm.List()
		beforeCount := len(beforeBackups)
		
		if beforeCount != 5 {
			t.Errorf("Expected 5 backups before cleanup, got %d", beforeCount)
		}
		
		// Keep only 2 most recent
		if err := cleanupBm.CleanupOld(2); err != nil {
			t.Fatalf("Failed to cleanup old backups: %v", err)
		}
		
		// Check remaining backups
		afterBackups, _ := cleanupBm.List()
		afterCount := len(afterBackups)
		
		if afterCount != 2 {
			t.Errorf("Expected 2 backups after cleanup, got %d", afterCount)
		}
		
		t.Logf("Cleaned up %d old backups", beforeCount-afterCount)
	})
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	
	for _, test := range tests {
		result := FormatSize(test.bytes)
		if result != test.expected {
			t.Errorf("FormatSize(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

func TestGetBackupTypeFromID(t *testing.T) {
	tests := []struct {
		id       string
		expected BackupType
	}{
		{"destroy-20250128_123456", BackupTypeDestroy},
		{"install-20250128_123456", BackupTypeInstall},
		{"manual-20250128_123456", BackupTypeManual},
		{"unknown-20250128_123456", BackupTypeManual}, // Default
		{"no-prefix", BackupTypeManual},                // Default
	}
	
	for _, test := range tests {
		result := GetBackupTypeFromID(test.id)
		if result != test.expected {
			t.Errorf("GetBackupTypeFromID(%s) = %s, expected %s", test.id, result, test.expected)
		}
	}
}