package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedInstallersExist(t *testing.T) {
	// Test that all embedded installers are available
	installers := []string{"claude", "claude-flow", "github-cli"}
	
	for _, installer := range installers {
		content, ok := installerMap[installer]
		assert.True(t, ok, "Installer %s should exist in installerMap", installer)
		assert.NotEmpty(t, content, "Installer %s should have content", installer)
	}
}

func TestGetInstaller(t *testing.T) {
	tests := []struct {
		name        string
		componentID string
		shouldExist bool
	}{
		{"claude installer", "claude", true},
		{"claude-flow installer", "claude-flow", true},
		{"github-cli installer", "github-cli", true},
		{"non-existent installer", "nonexistent", false},
		{"empty component ID", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := GetInstaller(tt.componentID)
			
			if tt.shouldExist {
				assert.NoError(t, err, "Should not return error for valid component")
				assert.NotEmpty(t, content, "Should return non-empty content")
				
				// Check that content looks like a shell script
				assert.True(t, strings.Contains(content, "#!/") || 
							strings.Contains(content, "set -e") ||
							strings.Contains(content, "echo") ||
							strings.Contains(content, "curl") ||
							strings.Contains(content, "wget"),
					"Content should look like a shell script")
			} else {
				assert.Error(t, err, "Should return error for invalid component")
				assert.Empty(t, content, "Should return empty content on error")
				assert.Contains(t, err.Error(), "installer not found", 
					"Error message should indicate installer not found")
			}
		})
	}
}

func TestGetInstallerContentFormat(t *testing.T) {
	// Test each installer has proper shell script format
	for componentID := range installerMap {
		t.Run(componentID, func(t *testing.T) {
			content, err := GetInstaller(componentID)
			require.NoError(t, err)
			require.NotEmpty(t, content)
			
			lines := strings.Split(content, "\n")
			assert.True(t, len(lines) > 1, "Installer should have multiple lines")
			
			// Should have some shell script characteristics
			hasShellIndicators := false
			for _, line := range lines[:10] { // Check first 10 lines
				if strings.Contains(line, "#!/") ||
				   strings.Contains(line, "set -") ||
				   strings.Contains(line, "echo") ||
				   strings.Contains(line, "curl") ||
				   strings.Contains(line, "wget") ||
				   strings.Contains(line, "export") {
					hasShellIndicators = true
					break
				}
			}
			assert.True(t, hasShellIndicators, 
				"Installer %s should have shell script indicators", componentID)
		})
	}
}

func TestExtractInstallers(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Test successful extraction
	err := ExtractInstallers(tempDir)
	assert.NoError(t, err, "ExtractInstallers should not return error")
	
	// Verify all installer files were created
	expectedFiles := []string{"claude.sh", "claude-flow.sh", "github-cli.sh"}
	for _, filename := range expectedFiles {
		filePath := filepath.Join(tempDir, filename)
		
		// Check file exists
		_, err := os.Stat(filePath)
		assert.NoError(t, err, "File %s should exist", filename)
		
		// Check file is executable
		info, err := os.Stat(filePath)
		require.NoError(t, err)
		assert.True(t, info.Mode()&0111 != 0, "File %s should be executable", filename)
		
		// Check file content matches embedded content
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		
		componentID := strings.TrimSuffix(filename, ".sh")
		expectedContent, err := GetInstaller(componentID)
		require.NoError(t, err)
		
		assert.Equal(t, expectedContent, string(content), 
			"File content should match embedded content for %s", filename)
	}
}

func TestExtractInstallersToNonExistentDirectory(t *testing.T) {
	// Test extraction to a directory that doesn't exist
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "nonexistent", "nested", "path")
	
	err := ExtractInstallers(nonExistentDir)
	assert.NoError(t, err, "Should create nested directories")
	
	// Verify directory was created
	info, err := os.Stat(nonExistentDir)
	assert.NoError(t, err, "Directory should be created")
	assert.True(t, info.IsDir(), "Path should be a directory")
	
	// Verify files were extracted
	files, err := os.ReadDir(nonExistentDir)
	assert.NoError(t, err)
	assert.Equal(t, len(installerMap), len(files), "Should have all installer files")
}

func TestExtractInstallersToReadOnlyDirectory(t *testing.T) {
	// Create a read-only directory
	tempDir := t.TempDir()
	readOnlyDir := filepath.Join(tempDir, "readonly")
	err := os.Mkdir(readOnlyDir, 0444) // Read-only
	require.NoError(t, err)
	
	// Attempt to extract installers
	err = ExtractInstallers(readOnlyDir)
	assert.Error(t, err, "Should return error for read-only directory")
	assert.Contains(t, err.Error(), "failed to write installer", 
		"Error should indicate write failure")
}

func TestExtractInstallersOverwriteExisting(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a dummy file first
	dummyFile := filepath.Join(tempDir, "claude.sh")
	err := os.WriteFile(dummyFile, []byte("dummy content"), 0644)
	require.NoError(t, err)
	
	// Extract installers (should overwrite)
	err = ExtractInstallers(tempDir)
	assert.NoError(t, err, "Should overwrite existing files")
	
	// Verify the file was overwritten with correct content
	content, err := os.ReadFile(dummyFile)
	require.NoError(t, err)
	
	expectedContent, err := GetInstaller("claude")
	require.NoError(t, err)
	
	assert.Equal(t, expectedContent, string(content), 
		"File should be overwritten with correct content")
}

func TestInstallerMapIntegrity(t *testing.T) {
	// Test that installerMap has expected structure
	assert.NotEmpty(t, installerMap, "installerMap should not be empty")
	
	expectedKeys := []string{"claude", "claude-flow", "github-cli"}
	assert.Equal(t, len(expectedKeys), len(installerMap), 
		"installerMap should have expected number of entries")
	
	for _, key := range expectedKeys {
		content, exists := installerMap[key]
		assert.True(t, exists, "Key %s should exist in installerMap", key)
		assert.NotEmpty(t, content, "Content for %s should not be empty", key)
	}
}

func TestEmbeddedContentNotEmpty(t *testing.T) {
	// Test each embedded variable directly
	assert.NotEmpty(t, claudeInstaller, "claudeInstaller should not be empty")
	assert.NotEmpty(t, claudeFlowInstaller, "claudeFlowInstaller should not be empty")
	assert.NotEmpty(t, githubCLIInstaller, "githubCLIInstaller should not be empty")
}

func TestInstallerContentQuality(t *testing.T) {
	// Test that installer content has reasonable quality indicators
	for componentID, content := range installerMap {
		t.Run(componentID, func(t *testing.T) {
			// Should have reasonable length (not just a few characters)
			assert.True(t, len(content) > 50, 
				"Installer %s should have substantial content (>50 chars)", componentID)
			
			// Should not contain placeholder text
			assert.NotContains(t, content, "TODO", 
				"Installer %s should not contain TODO placeholders", componentID)
			assert.NotContains(t, content, "FIXME", 
				"Installer %s should not contain FIXME placeholders", componentID)
			
			// Should contain typical installation patterns
			hasInstallPatterns := strings.Contains(content, "install") ||
								  strings.Contains(content, "curl") ||
								  strings.Contains(content, "wget") ||
								  strings.Contains(content, "download") ||
								  strings.Contains(content, "setup")
			
			assert.True(t, hasInstallPatterns, 
				"Installer %s should contain installation patterns", componentID)
		})
	}
}

func TestExtractInstallersPermissions(t *testing.T) {
	tempDir := t.TempDir()
	
	err := ExtractInstallers(tempDir)
	require.NoError(t, err)
	
	// Check that all files have correct permissions (executable)
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sh") {
			filePath := filepath.Join(tempDir, file.Name())
			info, err := os.Stat(filePath)
			require.NoError(t, err)
			
			// Check that file is executable by owner
			assert.True(t, info.Mode()&0100 != 0, 
				"File %s should be executable by owner", file.Name())
			
			// Check that file is readable by owner
			assert.True(t, info.Mode()&0400 != 0, 
				"File %s should be readable by owner", file.Name())
		}
	}
}

func TestGetInstallerCaseVariations(t *testing.T) {
	// Test that componentID is case-sensitive
	validIDs := []string{"claude", "claude-flow", "github-cli"}
	invalidVariations := []string{
		"Claude", "CLAUDE", "Claude-Flow", "CLAUDE-FLOW",
		"Github-Cli", "GITHUB-CLI", "github_cli", "claude_flow",
	}
	
	// Valid IDs should work
	for _, id := range validIDs {
		content, err := GetInstaller(id)
		assert.NoError(t, err, "Valid ID %s should work", id)
		assert.NotEmpty(t, content, "Valid ID %s should return content", id)
	}
	
	// Invalid variations should fail
	for _, id := range invalidVariations {
		content, err := GetInstaller(id)
		assert.Error(t, err, "Invalid ID %s should fail", id)
		assert.Empty(t, content, "Invalid ID %s should return empty content", id)
	}
}

// Benchmark tests
func BenchmarkGetInstaller(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GetInstaller("claude")
	}
}

func BenchmarkExtractInstallers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tempDir := b.TempDir()
		_ = ExtractInstallers(tempDir)
	}
}

func TestInstallerMapConsistency(t *testing.T) {
	// Test that installerMap content matches embedded variables
	assert.Equal(t, claudeInstaller, installerMap["claude"], 
		"installerMap['claude'] should match claudeInstaller")
	assert.Equal(t, claudeFlowInstaller, installerMap["claude-flow"], 
		"installerMap['claude-flow'] should match claudeFlowInstaller")
	assert.Equal(t, githubCLIInstaller, installerMap["github-cli"], 
		"installerMap['github-cli'] should match githubCLIInstaller")
}

func TestExtractInstallersFilenameGeneration(t *testing.T) {
	tempDir := t.TempDir()
	
	err := ExtractInstallers(tempDir)
	require.NoError(t, err)
	
	// Verify filename generation logic
	for componentID := range installerMap {
		expectedFilename := componentID + ".sh"
		filePath := filepath.Join(tempDir, expectedFilename)
		
		_, err := os.Stat(filePath)
		assert.NoError(t, err, 
			"File %s should exist for component %s", expectedFilename, componentID)
	}
}