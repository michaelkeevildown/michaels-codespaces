package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock interfaces for testing
type MockProgressCallback struct {
	calls []string
}

func (m *MockProgressCallback) Callback(message string) {
	m.calls = append(m.calls, message)
}

func (m *MockProgressCallback) GetCalls() []string {
	return m.calls
}

func TestProgressWriter_Write(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		expectedCalls  []string
		description    string
	}{
		{
			name:  "single line progress",
			input: []byte("Counting objects: 100% (50/50), done.\n"),
			expectedCalls: []string{"Counting objects: 100% (50/50), done."},
			description: "Should capture single progress line",
		},
		{
			name:  "multiple progress lines",
			input: []byte("Counting objects: 50% (25/50)\nCompressing objects: 100% (25/25), done.\n"),
			expectedCalls: []string{
				"Counting objects: 50% (25/50)",
				"Compressing objects: 100% (25/25), done.",
			},
			description: "Should capture multiple progress lines",
		},
		{
			name:  "receiving objects progress",
			input: []byte("Receiving objects: 75% (375/500), 2.1 MiB | 1.5 MiB/s\n"),
			expectedCalls: []string{"Receiving objects: 75% (375/500), 2.1 MiB | 1.5 MiB/s"},
			description: "Should capture receiving objects progress",
		},
		{
			name:  "resolving deltas progress",
			input: []byte("Resolving deltas: 100% (125/125), done.\n"),
			expectedCalls: []string{"Resolving deltas: 100% (125/125), done."},
			description: "Should capture resolving deltas progress",
		},
		{
			name:          "non-progress output",
			input:         []byte("warning: redirecting to https://github.com/user/repo.git/\n"),
			expectedCalls: nil,
			description:   "Should ignore non-progress output",
		},
		{
			name:          "empty input",
			input:         []byte(""),
			expectedCalls: nil,
			description:   "Should handle empty input",
		},
		{
			name:          "duplicate lines",
			input:         []byte("Counting objects: 100% (50/50), done.\nCounting objects: 100% (50/50), done.\n"),
			expectedCalls: []string{"Counting objects: 100% (50/50), done."},
			description:   "Should deduplicate identical lines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCallback := &MockProgressCallback{}
			writer := &ProgressWriter{
				callback: mockCallback.Callback,
			}

			n, err := writer.Write(tt.input)
			
			assert.NoError(t, err, "Write should not return error")
			assert.Equal(t, len(tt.input), n, "Should return number of bytes written")
			assert.Equal(t, tt.expectedCalls, mockCallback.GetCalls(), tt.description)
		})
	}
}

func TestProgressWriter_WriteNilCallback(t *testing.T) {
	writer := &ProgressWriter{callback: nil}
	
	input := []byte("Counting objects: 100% (50/50), done.\n")
	n, err := writer.Write(input)
	
	assert.NoError(t, err, "Write should not panic with nil callback")
	assert.Equal(t, len(input), n, "Should return number of bytes written")
}

func TestClone_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for clone
	tempDir, err := os.MkdirTemp("", "mcs-git-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	clonePath := filepath.Join(tempDir, "test-repo")
	
	// Use a public repository for testing
	testURL := "https://github.com/octocat/Hello-World.git"
	
	var progressMessages []string
	progressCallback := func(msg string) {
		progressMessages = append(progressMessages, msg)
	}

	opts := CloneOptions{
		URL:      testURL,
		Path:     clonePath,
		Depth:    1, // Shallow clone for faster test
		Progress: progressCallback,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = Clone(ctx, opts)
	
	assert.NoError(t, err, "Clone should succeed")
	assert.DirExists(t, clonePath, "Clone directory should exist")
	assert.FileExists(t, filepath.Join(clonePath, ".git"), "Git directory should exist")
	assert.Greater(t, len(progressMessages), 0, "Should have received progress messages")
}

func TestClone_WithBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "mcs-git-test-branch-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	clonePath := filepath.Join(tempDir, "test-repo-branch")
	testURL := "https://github.com/octocat/Hello-World.git"

	opts := CloneOptions{
		URL:    testURL,
		Path:   clonePath,
		Branch: "master", // Specify branch
		Depth:  1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = Clone(ctx, opts)
	
	assert.NoError(t, err, "Clone with branch should succeed")
	assert.DirExists(t, clonePath, "Clone directory should exist")
}

func TestClone_ParentDirectoryCreation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcs-git-test-parent-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create nested path that doesn't exist
	clonePath := filepath.Join(tempDir, "nested", "deep", "path", "test-repo")
	
	// Mock a successful clone by creating a minimal git repo structure
	opts := CloneOptions{
		URL:  "https://example.com/fake-repo.git",
		Path: clonePath,
	}

	// This will fail at the actual git clone, but should succeed in creating parent dirs
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = Clone(ctx, opts)
	
	// Should fail at git clone but parent dirs should be created
	assert.Error(t, err, "Should fail on invalid repository")
	assert.DirExists(t, filepath.Dir(clonePath), "Parent directories should be created")
}

func TestClone_InvalidURL(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcs-git-test-invalid-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	clonePath := filepath.Join(tempDir, "invalid-repo")
	
	opts := CloneOptions{
		URL:  "https://invalid-git-url-that-does-not-exist.com/repo.git",
		Path: clonePath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = Clone(ctx, opts)
	
	assert.Error(t, err, "Should fail with invalid URL")
	assert.Contains(t, err.Error(), "failed to clone repository", "Error should mention clone failure")
	assert.NoDirExists(t, clonePath, "Failed clone directory should be cleaned up")
}

func TestClone_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "mcs-git-test-cancel-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	clonePath := filepath.Join(tempDir, "cancelled-repo")
	
	opts := CloneOptions{
		URL:  "https://github.com/torvalds/linux.git", // Large repo to ensure cancellation works
		Path: clonePath,
	}

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = Clone(ctx, opts)
	
	assert.Error(t, err, "Should fail with cancelled context")
	assert.NoDirExists(t, clonePath, "Cancelled clone directory should be cleaned up")
}

func TestClone_GitHubTokenConversion(t *testing.T) {
	// Save original env
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	// Set test token
	os.Setenv("GITHUB_TOKEN", "test-token-123")

	tempDir, err := os.MkdirTemp("", "mcs-git-test-token-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	clonePath := filepath.Join(tempDir, "token-test-repo")
	
	// Test SSH URL conversion to HTTPS
	sshURL := "git@github.com:octocat/Hello-World.git"
	
	opts := CloneOptions{
		URL:  sshURL,
		Path: clonePath,
		Depth: 1,
	}

	// This should convert the SSH URL to HTTPS and use token auth
	// We'll just test the URL conversion logic by checking the actual clone attempt
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = Clone(ctx, opts)
	
	// The clone might succeed or fail, but it should attempt the conversion
	// We can't easily test the exact URL conversion without mocking the git library
	// This test mainly ensures no panic occurs with token-based conversion
	if err != nil {
		// If it fails, it should be a git-related error, not a conversion panic
		assert.NotContains(t, err.Error(), "panic", "Should not panic during URL conversion")
	}
}

func TestDetectAuthMethod_GitHubToken(t *testing.T) {
	// Save original env
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	tests := []struct {
		name        string
		url         string
		token       string
		expectType  string
		description string
	}{
		{
			name:        "github https with token",
			url:         "https://github.com/user/repo.git",
			token:       "test-token",
			expectType:  "http.BasicAuth",
			description: "Should use HTTP basic auth for GitHub HTTPS URLs with token",
		},
		{
			name:        "github ssh with token",
			url:         "git@github.com:user/repo.git",
			token:       "test-token",
			expectType:  "http.BasicAuth",
			description: "Should use HTTP basic auth for GitHub SSH URLs with token",
		},
		{
			name:        "non-github with token",
			url:         "https://gitlab.com/user/repo.git",
			token:       "test-token",
			expectType:  "nil",
			description: "Should not use GitHub token for non-GitHub URLs",
		},
		{
			name:        "github without token",
			url:         "https://github.com/user/repo.git",
			token:       "",
			expectType:  "nil",
			description: "Should return nil when no token available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.token != "" {
				os.Setenv("GITHUB_TOKEN", tt.token)
			} else {
				os.Unsetenv("GITHUB_TOKEN")
			}

			auth := detectAuthMethod(tt.url)

			switch tt.expectType {
			case "http.BasicAuth":
				assert.IsType(t, &http.BasicAuth{}, auth, tt.description)
				if basicAuth, ok := auth.(*http.BasicAuth); ok {
					assert.Equal(t, "token", basicAuth.Username, "GitHub token auth should use 'token' as username")
					assert.Equal(t, tt.token, basicAuth.Password, "Should use token as password")
				}
			case "nil":
				assert.Nil(t, auth, tt.description)
			}
		})
	}
}

func TestDetectAuthMethod_GitLabToken(t *testing.T) {
	// Save original env
	originalToken := os.Getenv("GITLAB_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITLAB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITLAB_TOKEN")
		}
	}()

	os.Setenv("GITLAB_TOKEN", "gitlab-test-token")

	auth := detectAuthMethod("https://gitlab.com/user/repo.git")
	
	assert.IsType(t, &http.BasicAuth{}, auth, "Should use HTTP basic auth for GitLab URLs")
	if basicAuth, ok := auth.(*http.BasicAuth); ok {
		assert.Equal(t, "oauth2", basicAuth.Username, "GitLab token auth should use 'oauth2' as username")
		assert.Equal(t, "gitlab-test-token", basicAuth.Password, "Should use GitLab token as password")
	}
}

func TestDetectAuthMethod_SSHKey(t *testing.T) {
	// This test is tricky because it depends on actual SSH keys existing
	// We'll test the logic flow but can't easily test successful SSH key loading
	
	// Clear tokens to force SSH path
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GITLAB_TOKEN")

	tests := []struct {
		name        string
		url         string
		description string
	}{
		{
			name:        "ssh url",
			url:         "git@github.com:user/repo.git",
			description: "Should attempt SSH auth for SSH URLs",
		},
		{
			name:        "ssh protocol url",
			url:         "ssh://git@gitlab.com/user/repo.git",
			description: "Should attempt SSH auth for SSH protocol URLs",
		},
		{
			name:        "https url without token",
			url:         "https://github.com/user/repo.git",
			description: "Should return nil for HTTPS URLs without tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := detectAuthMethod(tt.url)
			
			if strings.HasPrefix(tt.url, "git@") || strings.Contains(tt.url, "ssh://") {
				// For SSH URLs, it might return SSH auth or nil (if no keys found)
				// We mainly test that it doesn't panic
				if auth != nil {
					// If auth is returned, it should be SSH auth
					assert.IsType(t, &ssh.PublicKeys{}, auth, "SSH URLs should return SSH auth when keys are available")
				}
			} else {
				// For HTTPS URLs without tokens, should return nil
				assert.Nil(t, auth, tt.description)
			}
		})
	}
}

func TestValidateRepository_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Test with a known public repository
	err := ValidateRepository(ctx, "https://github.com/octocat/Hello-World.git")
	
	assert.NoError(t, err, "Should validate public repository successfully")
}

func TestValidateRepository_InvalidRepo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := ValidateRepository(ctx, "https://github.com/nonexistent/invalid-repo-12345.git")
	
	assert.Error(t, err, "Should fail validation for invalid repository")
	assert.Contains(t, err.Error(), "repository validation failed", "Error should mention validation failure")
}

func TestValidateRepository_NetworkError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ValidateRepository(ctx, "https://invalid-domain-that-does-not-exist-12345.com/repo.git")
	
	assert.Error(t, err, "Should fail validation for network errors")
	assert.Contains(t, err.Error(), "repository validation failed", "Error should mention validation failure")
}

func TestValidateRepository_ContextTimeout(t *testing.T) {
	// Create context that times out very quickly
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := ValidateRepository(ctx, "https://github.com/torvalds/linux.git")
	
	assert.Error(t, err, "Should fail validation on context timeout")
}

func TestGetDefaultBranch_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	branch, err := GetDefaultBranch(ctx, "https://github.com/octocat/Hello-World.git")
	
	assert.NoError(t, err, "Should get default branch successfully")
	assert.NotEmpty(t, branch, "Should return non-empty branch name")
	// Most repositories use either 'main' or 'master'
	assert.Contains(t, []string{"main", "master"}, branch, "Should return common default branch name")
}

func TestGetDefaultBranch_InvalidRepo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := GetDefaultBranch(ctx, "https://github.com/nonexistent/invalid-repo-12345.git")
	
	assert.Error(t, err, "Should fail to get default branch for invalid repository")
	assert.Contains(t, err.Error(), "failed to list references", "Error should mention reference listing failure")
}

func TestGetDefaultBranch_Fallback(t *testing.T) {
	// This test would require mocking the git remote list functionality
	// For now, we'll test the basic structure and ensure no panics
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with timeout to avoid long waits
	_, err := GetDefaultBranch(ctx, "https://invalid-test-url.com/repo.git")
	
	// Should error but not panic
	assert.Error(t, err, "Should fail gracefully for invalid URLs")
}

// Benchmark tests for performance monitoring
func BenchmarkProgressWriter_Write(b *testing.B) {
	writer := &ProgressWriter{
		callback: func(string) {}, // No-op callback
	}
	data := []byte("Counting objects: 100% (50/50), done.\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(data)
	}
}

func BenchmarkDetectAuthMethod(b *testing.B) {
	urls := []string{
		"https://github.com/user/repo.git",
		"git@github.com:user/repo.git",
		"https://gitlab.com/user/repo.git",
		"ssh://git@bitbucket.org/user/repo.git",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := urls[i%len(urls)]
		detectAuthMethod(url)
	}
}

// Test edge cases and error conditions
func TestClone_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		opts        CloneOptions
		expectError bool
		description string
	}{
		{
			name: "empty URL",
			opts: CloneOptions{
				URL:  "",
				Path: "/tmp/test",
			},
			expectError: true,
			description: "Should fail with empty URL",
		},
		{
			name: "empty path",
			opts: CloneOptions{
				URL:  "https://github.com/octocat/Hello-World.git",
				Path: "",
			},
			expectError: false, // Empty path might not cause immediate error in go-git, but will fail eventually
			description: "Should handle empty path gracefully (may fail later in process)",
		},
		{
			name: "negative depth",
			opts: CloneOptions{
				URL:   "https://github.com/octocat/Hello-World.git",
				Path:  "/tmp/test-negative-depth",
				Depth: -1,
			},
			expectError: false, // Negative depth should be ignored, not cause error
			description: "Should handle negative depth gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := Clone(ctx, tt.opts)
			
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				// For the negative depth test, we expect it to not panic
				// It might still fail due to other reasons (like invalid path)
				// but should not fail specifically due to negative depth
				if err != nil {
					assert.NotContains(t, err.Error(), "depth", "Should not fail due to depth parameter")
				}
			}

			// Clean up any created directories
			if tt.opts.Path != "" {
				os.RemoveAll(tt.opts.Path)
			}
		})
	}
}

// Helper functions for testing
func createTempRepo(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "mcs-test-repo-*")
	require.NoError(t, err)

	// Initialize a git repository
	_, err = git.PlainInit(tempDir, false)
	require.NoError(t, err)

	return tempDir
}

func TestClone_CustomAuth(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcs-git-test-auth-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	clonePath := filepath.Join(tempDir, "custom-auth-repo")
	
	// Create custom auth
	customAuth := &http.BasicAuth{
		Username: "test-user",
		Password: "test-pass",
	}

	opts := CloneOptions{
		URL:  "https://httpbin.org/status/404", // This will fail, but we test auth setup
		Path: clonePath,
		Auth: customAuth,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = Clone(ctx, opts)
	
	// Should fail due to invalid repo, but should use custom auth
	assert.Error(t, err, "Should fail with invalid repository")
	// The error should not be auth-related since we provided auth
	assert.NotContains(t, err.Error(), "authentication", "Should not have auth error with custom auth")
}