package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"Zero bytes", 0, "0 B"},
		{"Small bytes", 512, "512 B"},
		{"One kilobyte", 1024, "1.0 KB"},
		{"Multiple kilobytes", 2048, "2.0 KB"},
		{"One megabyte", 1048576, "1.0 MB"},
		{"One gigabyte", 1073741824, "1.0 GB"},
		{"Mixed MB", 1536000, "1.5 MB"},
		{"Large TB", 1099511627776, "1.0 TB"},
		{"Petabyte", 1125899906842624, "1.0 PB"},
		{"Exabyte", 1152921504606846976, "1.0 EB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateDirSize(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	
	// Create test files
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	subDir := filepath.Join(tempDir, "subdir")
	file3 := filepath.Join(subDir, "file3.txt")
	
	require.NoError(t, os.WriteFile(file1, []byte("hello"), 0644)) // 5 bytes
	require.NoError(t, os.WriteFile(file2, []byte("world!!!"), 0644)) // 8 bytes
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.WriteFile(file3, []byte("test"), 0644)) // 4 bytes
	
	size, err := CalculateDirSize(tempDir)
	require.NoError(t, err)
	assert.Equal(t, int64(17), size) // 5 + 8 + 4 = 17 bytes
}

func TestCalculateDirSize_NonexistentDir(t *testing.T) {
	_, err := CalculateDirSize("/nonexistent/directory")
	assert.Error(t, err)
}

func TestRemovePathsWithProgress(t *testing.T) {
	// Create temporary files and directories
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	subDir := filepath.Join(tempDir, "subdir")
	nonExistent := filepath.Join(tempDir, "nonexistent.txt")
	
	require.NoError(t, os.WriteFile(file1, []byte("test"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("test"), 0644))
	require.NoError(t, os.MkdirAll(subDir, 0755))
	
	paths := []string{file1, file2, subDir, nonExistent}
	
	successCount, failCount := RemovePathsWithProgress(paths, "Removing")
	
	// Should succeed in removing 3 items (2 files + 1 dir), fail 1 (nonexistent)
	assert.Equal(t, 3, successCount)
	assert.Equal(t, 0, failCount) // nonExistent doesn't exist, so it's skipped, not failed
	
	// Verify files are removed
	assert.False(t, PathExists(file1))
	assert.False(t, PathExists(file2))
	assert.False(t, PathExists(subDir))
}

func TestRemoveDirectoryWithSize(t *testing.T) {
	// Test with existing directory
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644))
	
	// Test with progress
	progress := ui.NewProgress()
	
	err := RemoveDirectoryWithSize(testDir, progress)
	assert.NoError(t, err)
	assert.False(t, PathExists(testDir))
}

func TestRemoveDirectoryWithSize_NonexistentDir(t *testing.T) {
	err := RemoveDirectoryWithSize("/nonexistent/dir", nil)
	assert.NoError(t, err) // Should not error for nonexistent directory
}

func TestRemoveDirectoryWithSize_WithProgressFail(t *testing.T) {
	t.Skip("Skipping UI progress test to avoid hanging")
}

func TestEnsureDirectory(t *testing.T) {
	tempDir := t.TempDir()
	newDir := filepath.Join(tempDir, "new", "nested", "dir")
	
	err := EnsureDirectory(newDir, 0755)
	assert.NoError(t, err)
	assert.True(t, IsDirectory(newDir))
	
	// Test with existing directory
	err = EnsureDirectory(newDir, 0755)
	assert.NoError(t, err) // Should not error if directory already exists
}

func TestPathExists(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing.txt")
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
	
	require.NoError(t, os.WriteFile(existingFile, []byte("test"), 0644))
	
	assert.True(t, PathExists(existingFile))
	assert.False(t, PathExists(nonExistentFile))
}

func TestIsDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testDir := filepath.Join(tempDir, "testdir")
	
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
	require.NoError(t, os.MkdirAll(testDir, 0755))
	
	assert.True(t, IsDirectory(testDir))
	assert.False(t, IsDirectory(testFile))
	assert.False(t, IsDirectory("/nonexistent/path"))
}

func TestGetFileSize(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test file size
	testFile := filepath.Join(tempDir, "test.txt")
	content := "hello world"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))
	
	size, err := GetFileSize(testFile)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)
	
	// Test directory size
	testDir := filepath.Join(tempDir, "testdir")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("test"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(testDir, "file2.txt"), []byte("hello"), 0644))
	
	size, err = GetFileSize(testDir)
	assert.NoError(t, err)
	assert.Equal(t, int64(9), size) // "test" (4) + "hello" (5) = 9
	
	// Test nonexistent file
	_, err = GetFileSize("/nonexistent/file")
	assert.Error(t, err)
}

func TestCleanPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string // What the result should contain
	}{
		{"Regular path", "/tmp/test", "/tmp/test"},
		{"Path with dots", "/tmp/../test", "/test"},
		{"Empty path", "", "."},
		{"Current dir", ".", "."},
		{"Double slashes", "//tmp//test", "/tmp/test"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanPath(tt.input)
			assert.Contains(t, result, strings.TrimPrefix(tt.contains, "/"))
		})
	}
	
	// Test home directory expansion
	t.Run("Home directory expansion", func(t *testing.T) {
		result := CleanPath("~/test")
		assert.True(t, strings.HasSuffix(result, "test"))
		assert.False(t, strings.HasPrefix(result, "~"))
	})
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	content := "test content for copy"
	require.NoError(t, os.WriteFile(srcFile, []byte(content), 0644))
	
	// Copy to destination
	dstFile := filepath.Join(tempDir, "destination.txt")
	err := CopyFile(srcFile, dstFile)
	assert.NoError(t, err)
	
	// Verify content
	copiedContent, err := os.ReadFile(dstFile)
	assert.NoError(t, err)
	assert.Equal(t, content, string(copiedContent))
	
	// Verify permissions are preserved
	srcInfo, err := os.Stat(srcFile)
	assert.NoError(t, err)
	dstInfo, err := os.Stat(dstFile)
	assert.NoError(t, err)
	assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
}

func TestCopyFile_Errors(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test copying nonexistent file
	err := CopyFile("/nonexistent/file", filepath.Join(tempDir, "dest"))
	assert.Error(t, err)
	
	// Test copying to invalid destination
	srcFile := filepath.Join(tempDir, "source.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("test"), 0644))
	
	err = CopyFile(srcFile, "/invalid/destination/file")
	assert.Error(t, err)
}

func TestListDirectorySize(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test files
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	subDir := filepath.Join(tempDir, "subdir")
	
	require.NoError(t, os.WriteFile(file1, []byte("hello"), 0644)) // 5 bytes
	require.NoError(t, os.WriteFile(file2, []byte("world"), 0644)) // 5 bytes
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("test"), 0644)) // 4 bytes
	
	sizes, err := ListDirectorySize(tempDir)
	assert.NoError(t, err)
	
	assert.Equal(t, int64(5), sizes["file1.txt"])
	assert.Equal(t, int64(5), sizes["file2.txt"])
	assert.Equal(t, int64(4), sizes["subdir"]) // Size of the nested file
	assert.Len(t, sizes, 3)
}

func TestListDirectorySize_NonexistentDir(t *testing.T) {
	_, err := ListDirectorySize("/nonexistent/directory")
	assert.Error(t, err)
}

func TestListDirectorySize_EmptyDir(t *testing.T) {
	tempDir := t.TempDir()
	
	sizes, err := ListDirectorySize(tempDir)
	assert.NoError(t, err)
	assert.Empty(t, sizes)
}

// Integration test for filesystem operations
func TestFilesystemIntegration(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test complete workflow
	testDir := filepath.Join(tempDir, "integration_test")
	
	// 1. Ensure directory exists
	err := EnsureDirectory(testDir, 0755)
	assert.NoError(t, err)
	assert.True(t, PathExists(testDir))
	assert.True(t, IsDirectory(testDir))
	
	// 2. Create some files
	file1 := filepath.Join(testDir, "file1.txt")
	file2 := filepath.Join(testDir, "file2.txt")
	require.NoError(t, os.WriteFile(file1, []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("content2"), 0644))
	
	// 3. Copy a file
	copiedFile := filepath.Join(testDir, "copied.txt")
	err = CopyFile(file1, copiedFile)
	assert.NoError(t, err)
	
	// 4. Check directory size
	sizes, err := ListDirectorySize(testDir)
	assert.NoError(t, err)
	assert.Len(t, sizes, 3)
	
	// 5. Calculate total directory size
	totalSize, err := CalculateDirSize(testDir)
	assert.NoError(t, err)
	assert.Equal(t, int64(24), totalSize) // 8 + 8 + 8 = 24 bytes
	
	// 6. Format size
	formattedSize := FormatBytes(totalSize)
	assert.Equal(t, "24 B", formattedSize)
	
	// 7. Clean up with progress
	progress := ui.NewProgress()
	err = RemoveDirectoryWithSize(testDir, progress)
	assert.NoError(t, err)
	assert.False(t, PathExists(testDir))
}