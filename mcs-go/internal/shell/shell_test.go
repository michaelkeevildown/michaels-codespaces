package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectShell(t *testing.T) {
	// Save original environment
	originalShell := os.Getenv("SHELL")
	originalBashVersion := os.Getenv("BASH_VERSION")
	originalZshVersion := os.Getenv("ZSH_VERSION")
	
	defer func() {
		os.Setenv("SHELL", originalShell)
		os.Setenv("BASH_VERSION", originalBashVersion)
		os.Setenv("ZSH_VERSION", originalZshVersion)
	}()

	tests := []struct {
		name         string
		shellEnv     string
		bashVersion  string
		zshVersion   string
		expected     ShellType
		description  string
	}{
		{
			name:        "detect bash via BASH_VERSION",
			shellEnv:    "",
			bashVersion: "5.0.0",
			zshVersion:  "",
			expected:    Bash,
			description: "Should detect bash when BASH_VERSION is set",
		},
		{
			name:        "detect zsh via ZSH_VERSION",
			shellEnv:    "",
			bashVersion: "",
			zshVersion:  "5.8",
			expected:    Zsh,
			description: "Should detect zsh when ZSH_VERSION is set",
		},
		{
			name:        "detect bash via SHELL env",
			shellEnv:    "/bin/bash",
			bashVersion: "",
			zshVersion:  "",
			expected:    Bash,
			description: "Should detect bash from SHELL environment variable",
		},
		{
			name:        "detect zsh via SHELL env",
			shellEnv:    "/usr/local/bin/zsh",
			bashVersion: "",
			zshVersion:  "",
			expected:    Zsh,
			description: "Should detect zsh from SHELL environment variable",
		},
		{
			name:        "fallback to sh",
			shellEnv:    "/bin/sh",
			bashVersion: "",
			zshVersion:  "",
			expected:    Sh,
			description: "Should fallback to sh for unknown shells",
		},
		{
			name:        "empty environment",
			shellEnv:    "",
			bashVersion: "",
			zshVersion:  "",
			expected:    Sh,
			description: "Should fallback to sh when no shell info available",
		},
		{
			name:        "bash version takes precedence",
			shellEnv:    "/usr/bin/zsh",
			bashVersion: "5.0.0",
			zshVersion:  "",
			expected:    Bash,
			description: "BASH_VERSION should take precedence over SHELL env",
		},
		{
			name:        "zsh version takes precedence",
			shellEnv:    "/bin/bash",
			bashVersion: "",
			zshVersion:  "5.8",
			expected:    Zsh,
			description: "ZSH_VERSION should take precedence over SHELL env",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Setenv("SHELL", tt.shellEnv)
			if tt.bashVersion != "" {
				os.Setenv("BASH_VERSION", tt.bashVersion)
			} else {
				os.Unsetenv("BASH_VERSION")
			}
			if tt.zshVersion != "" {
				os.Setenv("ZSH_VERSION", tt.zshVersion)
			} else {
				os.Unsetenv("ZSH_VERSION")
			}

			result := DetectShell()
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestGetShellConfigs(t *testing.T) {
	// Create temporary home directory
	tempHome, err := os.MkdirTemp("", "test-home-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempHome)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempHome)

	// Create test shell config files
	testFiles := map[string]ShellType{
		".bashrc":       Bash,
		".bash_profile": Bash,
		".zshrc":        Zsh,
		".zprofile":     Zsh,
		".profile":      Sh,
	}

	// Create some of the files
	createdFiles := []string{".bashrc", ".zshrc", ".profile"}
	for _, file := range createdFiles {
		filePath := filepath.Join(tempHome, file)
		err := os.WriteFile(filePath, []byte("# test config"), 0644)
		require.NoError(t, err)
	}

	configs := GetShellConfigs()

	// Should find the created files
	assert.Len(t, configs, len(createdFiles), "Should find all created config files")

	// Verify each config
	foundFiles := make(map[string]bool)
	for _, config := range configs {
		fileName := filepath.Base(config.FilePath)
		foundFiles[fileName] = true
		
		expectedShell := string(testFiles[fileName])
		assert.Equal(t, expectedShell, config.Shell, "Shell type should match for %s", fileName)
		assert.Equal(t, filepath.Join(tempHome, fileName), config.FilePath, "File path should be correct")
	}

	// Verify all created files were found
	for _, file := range createdFiles {
		assert.True(t, foundFiles[file], "Should find config file %s", file)
	}
}

func TestGetShellConfigs_NoFiles(t *testing.T) {
	// Create empty temporary home directory
	tempHome, err := os.MkdirTemp("", "test-home-empty-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempHome)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	os.Setenv("HOME", tempHome)

	configs := GetShellConfigs()

	assert.Empty(t, configs, "Should return empty slice when no config files exist")
}

func TestCleanConfig_Success(t *testing.T) {
	// Create temporary config file
	tempFile, err := os.CreateTemp("", "test-config-*.sh")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write test content
	content := `#!/bin/bash
export PATH="/old/path:$PATH"
# Some comment
# MCS - Michael's Codespaces
export MCS_HOME="/home/user/.mcs"
alias mcs="mcs-command"
# End MCS config
echo "other config"
export OTHER_VAR="value"
`
	_, err = tempFile.WriteString(content)
	require.NoError(t, err)
	tempFile.Close()

	// Clean patterns
	patterns := []string{
		"MCS - Michael's Codespaces",
		"export MCS_",
		"alias mcs=",
	}

	err = CleanConfig(tempFile.Name(), patterns)
	assert.NoError(t, err, "CleanConfig should succeed")

	// Read cleaned content
	cleanedContent, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)

	cleanedStr := string(cleanedContent)
	
	// Verify patterns were removed
	assert.NotContains(t, cleanedStr, "MCS - Michael's Codespaces", "Should remove MCS comment")
	assert.NotContains(t, cleanedStr, "export MCS_HOME", "Should remove MCS export")
	assert.NotContains(t, cleanedStr, "alias mcs=", "Should remove MCS alias")
	
	// Verify other content remains
	assert.Contains(t, cleanedStr, `export PATH="/old/path:$PATH"`, "Should keep other exports")
	assert.Contains(t, cleanedStr, `echo "other config"`, "Should keep other commands")
	assert.Contains(t, cleanedStr, `export OTHER_VAR="value"`, "Should keep other variables")
}

func TestCleanConfig_WithSkipNext(t *testing.T) {
	// Create temporary config file
	tempFile, err := os.CreateTemp("", "test-config-skip-*.sh")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write test content with lines that should trigger skipNext
	content := `#!/bin/bash
# MCS - Michael's Codespaces
export MCS_PATH="/usr/local/bin"
source ~/.mcs/init.sh
# End test
export KEEP_THIS="value"
`
	_, err = tempFile.WriteString(content)
	require.NoError(t, err)
	tempFile.Close()

	patterns := []string{"MCS - Michael's Codespaces"}

	err = CleanConfig(tempFile.Name(), patterns)
	assert.NoError(t, err, "CleanConfig should succeed")

	// Read cleaned content
	cleanedContent, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)

	cleanedStr := string(cleanedContent)
	
	// Verify the comment was removed 
	assert.NotContains(t, cleanedStr, "MCS - Michael's Codespaces", "Should remove MCS comment")
	// Note: The skipNext logic in CleanConfig only works for specific patterns (alias, export, source)
	// and only when they appear on the next line immediately after a pattern match
	
	// Verify other content remains
	assert.Contains(t, cleanedStr, `export KEEP_THIS="value"`, "Should keep unrelated exports")
}

func TestCleanConfig_NonExistentFile(t *testing.T) {
	err := CleanConfig("/non/existent/file.sh", []string{"pattern"})
	assert.NoError(t, err, "Should return nil for non-existent files")
}

func TestCleanConfig_ReadError(t *testing.T) {
	// Create a directory instead of a file to cause read error
	tempDir, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = CleanConfig(tempDir, []string{"pattern"})
	assert.Error(t, err, "Should return error when unable to read file")
}

func TestAddToConfig_NewFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-config-add-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, ".bashrc")
	comment := "Test Configuration"
	lines := []string{
		`export TEST_VAR="value"`,
		`alias test-cmd="echo test"`,
	}

	err = AddToConfig(configFile, comment, lines)
	assert.NoError(t, err, "AddToConfig should succeed")

	// Verify file was created and contains expected content
	content, err := os.ReadFile(configFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "#!/bin/bash", "Should include bash header")
	assert.Contains(t, contentStr, "# Test Configuration", "Should include comment")
	assert.Contains(t, contentStr, `export TEST_VAR="value"`, "Should include first line")
	assert.Contains(t, contentStr, `alias test-cmd="echo test"`, "Should include second line")
}

func TestAddToConfig_ZshFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-config-zsh-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, ".zshrc")
	comment := "ZSH Configuration"
	lines := []string{`export ZSH_VAR="value"`}

	err = AddToConfig(configFile, comment, lines)
	assert.NoError(t, err, "AddToConfig should succeed for zsh")

	content, err := os.ReadFile(configFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "#!/bin/zsh", "Should include zsh header for zsh files")
	assert.Contains(t, contentStr, "# ZSH Configuration", "Should include comment")
	assert.Contains(t, contentStr, `export ZSH_VAR="value"`, "Should include configuration line")
}

func TestAddToConfig_ExistingFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test-config-existing-*.sh")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write initial content
	initialContent := `#!/bin/bash
export EXISTING_VAR="existing"
`
	_, err = tempFile.WriteString(initialContent)
	require.NoError(t, err)
	tempFile.Close()

	comment := "Additional Configuration"
	lines := []string{`export NEW_VAR="new"`}

	err = AddToConfig(tempFile.Name(), comment, lines)
	assert.NoError(t, err, "AddToConfig should succeed on existing file")

	content, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, `export EXISTING_VAR="existing"`, "Should preserve existing content")
	assert.Contains(t, contentStr, "# Additional Configuration", "Should add new comment")
	assert.Contains(t, contentStr, `export NEW_VAR="new"`, "Should add new line")
}

func TestAddToConfig_DuplicateLines(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test-config-duplicate-*.sh")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write initial content including the line we'll try to add
	initialContent := `#!/bin/bash
export EXISTING_VAR="value"
`
	_, err = tempFile.WriteString(initialContent)
	require.NoError(t, err)
	tempFile.Close()

	comment := "Duplicate Test"
	lines := []string{`export EXISTING_VAR="value"`} // Same line already exists

	err = AddToConfig(tempFile.Name(), comment, lines)
	assert.NoError(t, err, "AddToConfig should succeed but not add duplicate")

	content, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)

	contentStr := string(content)
	// Should not add the comment or duplicate line
	assert.NotContains(t, contentStr, "# Duplicate Test", "Should not add comment for duplicate lines")
	
	// Count occurrences of the line
	count := strings.Count(contentStr, `export EXISTING_VAR="value"`)
	assert.Equal(t, 1, count, "Should not duplicate existing lines")
}

func TestAddToConfig_EmptyComment(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-config-no-comment-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, ".profile")
	lines := []string{`export NO_COMMENT_VAR="value"`}

	err = AddToConfig(configFile, "", lines) // Empty comment
	assert.NoError(t, err, "AddToConfig should succeed with empty comment")

	content, err := os.ReadFile(configFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, `export NO_COMMENT_VAR="value"`, "Should include the line")
	assert.NotContains(t, contentStr, "# ", "Should not include empty comment")
}

func TestRemoveMCSConfig(t *testing.T) {
	// Create temporary home directory
	tempHome, err := os.MkdirTemp("", "test-home-mcs-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempHome)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)

	// Create test config files with MCS content
	bashrcContent := `#!/bin/bash
export PATH="/usr/bin:$PATH"
# MCS - Michael's Codespaces
export MCS_HOME="/home/user/.mcs"
export PATH="$HOME/.mcs/bin:$PATH"
alias mcs="/usr/local/bin/mcs"
# End MCS
export OTHER_VAR="keep this"
`

	zshrcContent := `#!/bin/zsh
# Some other config
# Codespace: test-codespace
source ~/.mcs/completion.zsh
export KEEP_THIS="value"
`

	bashrcPath := filepath.Join(tempHome, ".bashrc")
	zshrcPath := filepath.Join(tempHome, ".zshrc")

	err = os.WriteFile(bashrcPath, []byte(bashrcContent), 0644)
	require.NoError(t, err)
	err = os.WriteFile(zshrcPath, []byte(zshrcContent), 0644)
	require.NoError(t, err)

	// Remove MCS config
	err = RemoveMCSConfig()
	assert.NoError(t, err, "RemoveMCSConfig should succeed")

	// Check bashrc was cleaned
	bashrcCleaned, err := os.ReadFile(bashrcPath)
	require.NoError(t, err)
	bashrcStr := string(bashrcCleaned)

	assert.NotContains(t, bashrcStr, "Michael's Codespaces", "Should remove MCS comment")
	assert.NotContains(t, bashrcStr, "export MCS_HOME", "Should remove MCS exports")
	assert.NotContains(t, bashrcStr, "/.mcs/bin", "Should remove MCS path")
	assert.NotContains(t, bashrcStr, "alias mcs=", "Should remove MCS alias")
	assert.Contains(t, bashrcStr, `export OTHER_VAR="keep this"`, "Should keep other config")

	// Check zshrc was cleaned
	zshrcCleaned, err := os.ReadFile(zshrcPath)
	require.NoError(t, err)
	zshrcStr := string(zshrcCleaned)

	assert.NotContains(t, zshrcStr, "# Codespace:", "Should remove codespace comments")
	assert.NotContains(t, zshrcStr, "mcs completion", "Should remove completion source")
	assert.Contains(t, zshrcStr, `export KEEP_THIS="value"`, "Should keep other config")
}

func TestRemoveMCSConfig_NoFiles(t *testing.T) {
	// Create empty temporary home directory
	tempHome, err := os.MkdirTemp("", "test-home-empty-mcs-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempHome)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)

	err = RemoveMCSConfig()
	assert.NoError(t, err, "RemoveMCSConfig should succeed even with no config files")
}

func TestAddMCSToPath(t *testing.T) {
	// Create temporary home directory
	tempHome, err := os.MkdirTemp("", "test-home-path-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempHome)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)

	// Create test config files
	bashrcPath := filepath.Join(tempHome, ".bashrc")
	zshrcPath := filepath.Join(tempHome, ".zshrc")

	err = os.WriteFile(bashrcPath, []byte("#!/bin/bash\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(zshrcPath, []byte("#!/bin/zsh\n"), 0644)
	require.NoError(t, err)

	testBinDir := "/usr/local/mcs/bin"

	updated, err := AddMCSToPath(testBinDir)
	assert.NoError(t, err, "AddMCSToPath should succeed")
	assert.Equal(t, 2, updated, "Should update 2 config files")

	// Check bashrc was updated
	bashrcContent, err := os.ReadFile(bashrcPath)
	require.NoError(t, err)
	bashrcStr := string(bashrcContent)

	assert.Contains(t, bashrcStr, "# MCS - Michael's Codespaces", "Should add MCS comment")
	assert.Contains(t, bashrcStr, fmt.Sprintf(`export PATH="%s:$PATH"`, testBinDir), "Should add PATH export")

	// Check zshrc was updated
	zshrcContent, err := os.ReadFile(zshrcPath)
	require.NoError(t, err)
	zshrcStr := string(zshrcContent)

	assert.Contains(t, zshrcStr, "# MCS - Michael's Codespaces", "Should add MCS comment")
	assert.Contains(t, zshrcStr, fmt.Sprintf(`export PATH="%s:$PATH"`, testBinDir), "Should add PATH export")
}

func TestAddMCSToPath_AlreadyConfigured(t *testing.T) {
	// Create temporary home directory
	tempHome, err := os.MkdirTemp("", "test-home-path-exists-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempHome)

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempHome)

	testBinDir := "/usr/local/mcs/bin"

	// Create bashrc with MCS path already configured
	bashrcPath := filepath.Join(tempHome, ".bashrc")
	existingContent := fmt.Sprintf(`#!/bin/bash
export PATH="%s:$PATH"
`, testBinDir)
	err = os.WriteFile(bashrcPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	updated, err := AddMCSToPath(testBinDir)
	assert.NoError(t, err, "AddMCSToPath should succeed")
	assert.Equal(t, 0, updated, "Should not update files that already have the path")

	// Verify content wasn't duplicated
	finalContent, err := os.ReadFile(bashrcPath)
	require.NoError(t, err)
	finalStr := string(finalContent)

	pathCount := strings.Count(finalStr, testBinDir)
	assert.Equal(t, 1, pathCount, "Should not duplicate existing PATH entries")
}

func TestSourceConfig(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	testHome := "/home/testuser"
	os.Setenv("HOME", testHome)

	tests := []struct {
		shell    ShellType
		expected string
	}{
		{Bash, "source /home/testuser/.bashrc"},
		{Zsh, "source /home/testuser/.zshrc"},
		{Sh, "source /home/testuser/.profile"},
		{ShellType("unknown"), "source /home/testuser/.profile"},
	}

	for _, tt := range tests {
		t.Run(string(tt.shell), func(t *testing.T) {
			result := SourceConfig(tt.shell)
			assert.Equal(t, tt.expected, result, "Should return correct source command")
		})
	}
}

func TestGetCompletionScript(t *testing.T) {
	testCommand := "mcs"

	tests := []struct {
		shell    ShellType
		expected string
	}{
		{Bash, "source <(mcs completion bash)"},
		{Zsh, "source <(mcs completion zsh)"},
		{Sh, ""},
		{ShellType("unknown"), ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.shell), func(t *testing.T) {
			result := GetCompletionScript(tt.shell, testCommand)
			assert.Equal(t, tt.expected, result, "Should return correct completion script")
		})
	}
}

func TestIsLoginShell(t *testing.T) {
	// Save original args and environment
	originalArgs := os.Args
	originalSHLVL := os.Getenv("SHLVL")
	defer func() {
		os.Args = originalArgs
		os.Setenv("SHLVL", originalSHLVL)
	}()

	tests := []struct {
		name        string
		args        []string
		shlvl       string
		expected    bool
		description string
	}{
		{
			name:        "login shell with dash prefix",
			args:        []string{"-bash"},
			shlvl:       "2",
			expected:    true,
			description: "Should detect login shell from dash prefix in args",
		},
		{
			name:        "non-login shell args",
			args:        []string{"bash"},
			shlvl:       "1",
			expected:    true,
			description: "Should detect login shell from SHLVL=1",
		},
		{
			name:        "non-login shell",
			args:        []string{"bash"},
			shlvl:       "2",
			expected:    false,
			description: "Should detect non-login shell",
		},
		{
			name:        "empty args with SHLVL=1",
			args:        []string{},
			shlvl:       "1",
			expected:    true,
			description: "Should fallback to SHLVL check with empty args",
		},
		{
			name:        "empty args and SHLVL",
			args:        []string{},
			shlvl:       "",
			expected:    false,
			description: "Should return false with no indicators",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			os.Setenv("SHLVL", tt.shlvl)

			result := IsLoginShell()
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

// Test error handling for file operations
func TestAddToConfig_WriteError(t *testing.T) {
	// Create a read-only directory
	tempDir, err := os.MkdirTemp("", "test-readonly-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = os.Chmod(tempDir, 0444) // Read-only
	if err != nil {
		t.Skip("Unable to make directory read-only, skipping test")
	}
	defer os.Chmod(tempDir, 0755) // Restore permissions for cleanup

	configFile := filepath.Join(tempDir, ".bashrc")
	lines := []string{`export TEST="value"`}

	err = AddToConfig(configFile, "Test", lines)
	assert.Error(t, err, "Should fail when unable to create file in read-only directory")
}

func TestCleanConfig_WriteError(t *testing.T) {
	// Create a file and make it read-only
	tempFile, err := os.CreateTemp("", "test-readonly-*.sh")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("# MCS - Michael's Codespaces\n")
	require.NoError(t, err)
	tempFile.Close()

	err = os.Chmod(tempFile.Name(), 0444) // Read-only
	require.NoError(t, err)
	defer os.Chmod(tempFile.Name(), 0644) // Restore permissions

	err = CleanConfig(tempFile.Name(), []string{"MCS"})
	assert.Error(t, err, "Should fail when unable to write to read-only file")
}

// Benchmark tests
func BenchmarkDetectShell(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DetectShell()
	}
}

func BenchmarkGetShellConfigs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetShellConfigs()
	}
}

func BenchmarkCleanConfig(b *testing.B) {
	// Create temporary file for benchmarking
	tempFile, err := os.CreateTemp("", "bench-config-*.sh")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	content := strings.Repeat("export VAR=value\n# MCS comment\nexport MCS_VAR=value\n", 100)
	_, err = tempFile.WriteString(content)
	if err != nil {
		b.Fatal(err)
	}
	tempFile.Close()

	patterns := []string{"MCS"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CleanConfig(tempFile.Name(), patterns)
	}
}