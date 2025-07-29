package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyDirectory(t *testing.T) {
	// Create source directory structure
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	dstDir := filepath.Join(tempDir, "destination")
	
	// Create source structure
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	
	// Create files in source
	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "file2.txt")
	require.NoError(t, os.WriteFile(file1, []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("content2"), 0755)) // Different permissions
	
	// Create subdirectory with file
	subDir := filepath.Join(srcDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	subFile := filepath.Join(subDir, "subfile.txt")
	require.NoError(t, os.WriteFile(subFile, []byte("subcontent"), 0600))
	
	// Copy the directory
	err := CopyDirectory(srcDir, dstDir)
	assert.NoError(t, err)
	
	// Verify destination structure exists
	assert.True(t, PathExists(dstDir))
	assert.True(t, IsDirectory(dstDir))
	
	// Verify files were copied
	dstFile1 := filepath.Join(dstDir, "file1.txt")
	dstFile2 := filepath.Join(dstDir, "file2.txt")
	assert.True(t, PathExists(dstFile1))
	assert.True(t, PathExists(dstFile2))
	
	// Verify file contents
	content1, err := os.ReadFile(dstFile1)
	assert.NoError(t, err)
	assert.Equal(t, "content1", string(content1))
	
	content2, err := os.ReadFile(dstFile2)
	assert.NoError(t, err)
	assert.Equal(t, "content2", string(content2))
	
	// Verify subdirectory was copied
	dstSubDir := filepath.Join(dstDir, "subdir")
	assert.True(t, PathExists(dstSubDir))
	assert.True(t, IsDirectory(dstSubDir))
	
	// Verify subdirectory file
	dstSubFile := filepath.Join(dstSubDir, "subfile.txt")
	assert.True(t, PathExists(dstSubFile))
	subContent, err := os.ReadFile(dstSubFile)
	assert.NoError(t, err)
	assert.Equal(t, "subcontent", string(subContent))
}

func TestCopyDirectory_PreservesPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping permission test in short mode")
	}
	
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	dstDir := filepath.Join(tempDir, "destination")
	
	// Create source directory with specific permissions
	require.NoError(t, os.MkdirAll(srcDir, 0750))
	
	// Create file with specific permissions
	file := filepath.Join(srcDir, "test.txt")
	require.NoError(t, os.WriteFile(file, []byte("test"), 0600))
	
	// Copy directory
	err := CopyDirectory(srcDir, dstDir)
	assert.NoError(t, err)
	
	// Check directory permissions
	srcInfo, err := os.Stat(srcDir)
	require.NoError(t, err)
	dstInfo, err := os.Stat(dstDir)
	require.NoError(t, err)
	assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
	
	// Check file permissions (via CopyFile which preserves permissions)
	srcFileInfo, err := os.Stat(file)
	require.NoError(t, err)
	dstFile := filepath.Join(dstDir, "test.txt")
	dstFileInfo, err := os.Stat(dstFile)
	require.NoError(t, err)
	assert.Equal(t, srcFileInfo.Mode(), dstFileInfo.Mode())
}

func TestCopyDirectory_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "empty_source")
	dstDir := filepath.Join(tempDir, "empty_destination")
	
	// Create empty source directory
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	
	// Copy empty directory
	err := CopyDirectory(srcDir, dstDir)
	assert.NoError(t, err)
	
	// Verify destination exists and is empty
	assert.True(t, PathExists(dstDir))
	assert.True(t, IsDirectory(dstDir))
	
	entries, err := os.ReadDir(dstDir)
	assert.NoError(t, err)
	assert.Empty(t, entries)
}

func TestCopyDirectory_DeepNesting(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	dstDir := filepath.Join(tempDir, "destination")
	
	// Create deeply nested structure
	deepPath := srcDir
	for i := 0; i < 5; i++ {
		deepPath = filepath.Join(deepPath, "level"+string(rune('0'+i)))
	}
	require.NoError(t, os.MkdirAll(deepPath, 0755))
	
	// Add a file at the deepest level
	deepFile := filepath.Join(deepPath, "deep.txt")
	require.NoError(t, os.WriteFile(deepFile, []byte("deep content"), 0644))
	
	// Copy directory
	err := CopyDirectory(srcDir, dstDir)
	assert.NoError(t, err)
	
	// Verify deep structure was copied
	expectedDeepPath := dstDir
	for i := 0; i < 5; i++ {
		expectedDeepPath = filepath.Join(expectedDeepPath, "level"+string(rune('0'+i)))
	}
	assert.True(t, PathExists(expectedDeepPath))
	
	expectedDeepFile := filepath.Join(expectedDeepPath, "deep.txt")
	assert.True(t, PathExists(expectedDeepFile))
	
	content, err := os.ReadFile(expectedDeepFile)
	assert.NoError(t, err)
	assert.Equal(t, "deep content", string(content))
}

func TestCopyDirectory_NonexistentSource(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "nonexistent")
	dstDir := filepath.Join(tempDir, "destination")
	
	err := CopyDirectory(srcDir, dstDir)
	assert.Error(t, err)
	assert.False(t, PathExists(dstDir))
}

func TestCopyDirectory_SourceIsFile(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	dstDir := filepath.Join(tempDir, "destination")
	
	// Create source file instead of directory
	require.NoError(t, os.WriteFile(srcFile, []byte("test"), 0644))
	
	err := CopyDirectory(srcFile, dstDir)
	assert.Error(t, err)
}

func TestCopyDirectory_DestinationExists(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	dstDir := filepath.Join(tempDir, "destination")
	
	// Create source directory
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("test"), 0644))
	
	// Create destination directory (already exists)
	require.NoError(t, os.MkdirAll(dstDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dstDir, "existing.txt"), []byte("existing"), 0644))
	
	// Copy should still work (merging)
	err := CopyDirectory(srcDir, dstDir)
	assert.NoError(t, err)
	
	// Both files should exist
	assert.True(t, PathExists(filepath.Join(dstDir, "test.txt")))
	assert.True(t, PathExists(filepath.Join(dstDir, "existing.txt")))
}

func TestCopyDirectory_ReadOnlySource(t *testing.T) {
	if testing.Short() || os.Getuid() == 0 {
		t.Skip("Skipping read-only test (short mode or running as root)")
	}
	
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "readonly_source")
	dstDir := filepath.Join(tempDir, "destination")
	
	// Create source with read-only file
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	readOnlyFile := filepath.Join(srcDir, "readonly.txt")
	require.NoError(t, os.WriteFile(readOnlyFile, []byte("readonly"), 0400)) // Read-only
	
	// Copy should still work
	err := CopyDirectory(srcDir, dstDir)
	assert.NoError(t, err)
	
	// Verify read-only file was copied
	dstReadOnlyFile := filepath.Join(dstDir, "readonly.txt")
	assert.True(t, PathExists(dstReadOnlyFile))
	
	content, err := os.ReadFile(dstReadOnlyFile)
	assert.NoError(t, err)
	assert.Equal(t, "readonly", string(content))
}

func TestCopyDirectory_SymlinksHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping symlink test in short mode")
	}
	
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	dstDir := filepath.Join(tempDir, "destination")
	
	// Create source directory
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	
	// Create a regular file
	regularFile := filepath.Join(srcDir, "regular.txt")
	require.NoError(t, os.WriteFile(regularFile, []byte("regular"), 0644))
	
	// Create a symlink to the regular file
	symlinkFile := filepath.Join(srcDir, "symlink.txt")
	err := os.Symlink("regular.txt", symlinkFile)
	if err != nil {
		t.Skip("Cannot create symlinks on this system")
	}
	
	// Copy directory
	err = CopyDirectory(srcDir, dstDir)
	assert.NoError(t, err)
	
	// The symlink should be treated as a regular file by our copy function
	// since os.ReadDir() returns DirEntry.Type() which follows symlinks
	dstSymlink := filepath.Join(dstDir, "symlink.txt")
	assert.True(t, PathExists(dstSymlink))
	
	// Content should be copied (following the symlink)
	content, err := os.ReadFile(dstSymlink)
	assert.NoError(t, err)
	assert.Equal(t, "regular", string(content))
}

func TestCopyDirectory_LargeFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}
	
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	dstDir := filepath.Join(tempDir, "destination")
	
	// Create source directory
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	
	// Create a moderately large file (1MB)
	largeFile := filepath.Join(srcDir, "large.txt")
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	require.NoError(t, os.WriteFile(largeFile, largeContent, 0644))
	
	// Copy directory
	err := CopyDirectory(srcDir, dstDir)
	assert.NoError(t, err)
	
	// Verify large file was copied correctly
	dstLargeFile := filepath.Join(dstDir, "large.txt")
	assert.True(t, PathExists(dstLargeFile))
	
	copiedContent, err := os.ReadFile(dstLargeFile)
	assert.NoError(t, err)
	assert.Equal(t, largeContent, copiedContent)
	
	// Verify file sizes match
	srcInfo, err := os.Stat(largeFile)
	assert.NoError(t, err)
	dstInfo, err := os.Stat(dstLargeFile)
	assert.NoError(t, err)
	assert.Equal(t, srcInfo.Size(), dstInfo.Size())
}

func TestCopyDirectory_SpecialCharacters(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	dstDir := filepath.Join(tempDir, "destination")
	
	// Create source directory
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	
	// Create files with special characters in names
	specialFiles := []string{
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
		"file.with.dots.txt",
		"file(with)parentheses.txt",
		"file[with]brackets.txt",
	}
	
	for i, filename := range specialFiles {
		filePath := filepath.Join(srcDir, filename)
		content := "content" + string(rune('0'+i))
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
	}
	
	// Copy directory
	err := CopyDirectory(srcDir, dstDir)
	assert.NoError(t, err)
	
	// Verify all special files were copied
	for i, filename := range specialFiles {
		dstFile := filepath.Join(dstDir, filename)
		assert.True(t, PathExists(dstFile), "File %s should exist", filename)
		
		content, err := os.ReadFile(dstFile)
		assert.NoError(t, err)
		expectedContent := "content" + string(rune('0'+i))
		assert.Equal(t, expectedContent, string(content))
	}
}

// Integration test combining CopyDirectory with other utils functions
func TestCopyDirectoryIntegration(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	dstDir := filepath.Join(tempDir, "destination")
	
	// Create source structure using other utils functions
	err := EnsureDirectory(srcDir, 0755)
	require.NoError(t, err)
	
	// Create some files
	file1 := filepath.Join(srcDir, "file1.txt")
	file2 := filepath.Join(srcDir, "file2.txt")
	require.NoError(t, os.WriteFile(file1, []byte("hello"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("world"), 0644))
	
	// Create subdirectory
	subDir := filepath.Join(srcDir, "subdir")
	err = EnsureDirectory(subDir, 0755)
	require.NoError(t, err)
	
	subFile := filepath.Join(subDir, "subfile.txt")
	require.NoError(t, os.WriteFile(subFile, []byte("sub"), 0644))
	
	// Get original size using utils function
	originalSize, err := CalculateDirSize(srcDir)
	require.NoError(t, err)
	
	// Copy the directory
	err = CopyDirectory(srcDir, dstDir)
	assert.NoError(t, err)
	
	// Verify copy using utils functions
	assert.True(t, PathExists(dstDir))
	assert.True(t, IsDirectory(dstDir))
	
	// Verify size matches
	copiedSize, err := CalculateDirSize(dstDir)
	assert.NoError(t, err)
	assert.Equal(t, originalSize, copiedSize)
	
	// Verify individual files using utils functions
	assert.True(t, PathExists(filepath.Join(dstDir, "file1.txt")))
	assert.True(t, PathExists(filepath.Join(dstDir, "file2.txt")))
	assert.True(t, PathExists(filepath.Join(dstDir, "subdir")))
	assert.True(t, PathExists(filepath.Join(dstDir, "subdir", "subfile.txt")))
	
	// List directory contents
	sizes, err := ListDirectorySize(dstDir)
	assert.NoError(t, err)
	assert.Len(t, sizes, 3) // file1.txt, file2.txt, subdir
	assert.Equal(t, int64(5), sizes["file1.txt"])
	assert.Equal(t, int64(5), sizes["file2.txt"])
	assert.Equal(t, int64(3), sizes["subdir"]) // Size of subfile.txt
}

// Benchmark tests
func BenchmarkCopyDirectory_Small(b *testing.B) {
	// Setup small directory
	tempDir := b.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	os.MkdirAll(srcDir, 0755)
	
	for i := 0; i < 10; i++ {
		filename := filepath.Join(srcDir, "file"+string(rune('0'+i))+".txt")
		os.WriteFile(filename, []byte("content"), 0644)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dstDir := filepath.Join(tempDir, "dst"+string(rune('0'+i)))
		CopyDirectory(srcDir, dstDir)
	}
}

func BenchmarkCopyDirectory_Deep(b *testing.B) {
	// Setup deep directory structure
	tempDir := b.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	
	// Create 5 levels deep with 3 files at each level
	current := srcDir
	for level := 0; level < 5; level++ {
		os.MkdirAll(current, 0755)
		for file := 0; file < 3; file++ {
			filename := filepath.Join(current, "file"+string(rune('0'+file))+".txt")
			os.WriteFile(filename, []byte("content at level "+string(rune('0'+level))), 0644)
		}
		current = filepath.Join(current, "level"+string(rune('0'+level+1)))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dstDir := filepath.Join(tempDir, "dst"+string(rune('0'+i)))
		CopyDirectory(srcDir, dstDir)
	}
}