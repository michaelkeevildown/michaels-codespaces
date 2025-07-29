package version

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		gitCommit   string
		gitDirty    string
		expected    string
	}{
		{
			name:     "release version",
			version:  "1.2.3",
			expected: "1.2.3",
		},
		{
			name:      "dev version with commit",
			version:   "dev",
			gitCommit: "abcdef1234567890",
			expected:  "dev-abcdef12",
		},
		{
			name:      "dev version with short commit",
			version:   "dev",
			gitCommit: "abc123",
			expected:  "dev-abc123",
		},
		{
			name:      "dev version with commit and dirty flag",
			version:   "dev",
			gitCommit: "abcdef1234567890",
			gitDirty:  "true",
			expected:  "dev-abcdef12-dirty",
		},
		{
			name:      "dev version without commit",
			version:   "dev",
			gitCommit: "unknown", // Explicitly set to unknown
			expected:  "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origVersion := Version
			origGitCommit := GitCommit
			origGitDirty := GitDirty
			
			// Set test values
			Version = tt.version
			GitCommit = tt.gitCommit
			if tt.gitDirty != "" {
				GitDirty = tt.gitDirty
			} else {
				GitDirty = ""
			}
			
			// Test
			result := Info()
			assert.Equal(t, tt.expected, result)
			
			// Restore original values
			Version = origVersion
			GitCommit = origGitCommit
			GitDirty = origGitDirty
		})
	}
}

func TestDetailedInfo(t *testing.T) {
	// Save original values
	origVersion := Version
	origGitCommit := GitCommit
	origGitTag := GitTag
	origBuildTime := BuildTime
	
	// Set test values
	Version = "1.2.3"
	GitCommit = "abcdef1234567890"
	GitTag = "v1.2.3"
	BuildTime = "2023-01-01T12:00:00Z"
	
	result := DetailedInfo()
	
	// Test that all expected fields are present
	assert.Contains(t, result, "Version:")
	assert.Contains(t, result, "1.2.3")
	assert.Contains(t, result, "Git commit:")
	assert.Contains(t, result, "abcdef1234567890")
	assert.Contains(t, result, "Git tag:")
	assert.Contains(t, result, "v1.2.3")
	assert.Contains(t, result, "Built:")
	assert.Contains(t, result, "2023-01-01T12:00:00Z")
	assert.Contains(t, result, "Go version:")
	assert.Contains(t, result, runtime.Version())
	assert.Contains(t, result, "OS/Arch:")
	assert.Contains(t, result, runtime.GOOS+"/"+runtime.GOARCH)
	
	// Test that it's properly formatted with newlines
	lines := strings.Split(result, "\n")
	assert.True(t, len(lines) >= 6, "Should have at least 6 lines of info")
	
	// Restore original values
	Version = origVersion
	GitCommit = origGitCommit
	GitTag = origGitTag
	BuildTime = origBuildTime
}

func TestDetailedInfoWithMissingFields(t *testing.T) {
	// Save original values
	origVersion := Version
	origGitCommit := GitCommit
	origGitTag := GitTag
	origBuildTime := BuildTime
	
	// Set test values with some missing fields
	Version = "dev"
	GitCommit = "unknown"
	GitTag = ""
	BuildTime = "unknown"
	
	result := DetailedInfo()
	
	// Should still contain basic info
	assert.Contains(t, result, "Version:")
	assert.Contains(t, result, "dev")
	assert.Contains(t, result, "Go version:")
	assert.Contains(t, result, "OS/Arch:")
	
	// Should not contain unknown/empty fields
	assert.NotContains(t, result, "Git commit: unknown")
	assert.NotContains(t, result, "Git tag:")
	assert.NotContains(t, result, "Built: unknown")
	
	// Restore original values
	Version = origVersion
	GitCommit = origGitCommit
	GitTag = origGitTag
	BuildTime = origBuildTime
}

func TestIsDevBuild(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"dev version", "dev", true},
		{"dev with commit", "dev-abc123", true},
		{"dev with commit and dirty", "dev-abc123-dirty", true},
		{"release version", "1.2.3", false},
		{"beta version", "1.2.3-beta", false},
		{"rc version", "1.2.3-rc1", false},
		{"empty version", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			origVersion := Version
			
			// Set test value
			Version = tt.version
			
			// Test
			result := IsDevBuild()
			assert.Equal(t, tt.expected, result)
			
			// Restore original value
			Version = origVersion
		})
	}
}

func TestIsPreRelease(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"beta version", "1.2.3-beta", true},
		{"beta with number", "1.2.3-beta1", true},
		{"rc version", "1.2.3-rc", true},
		{"rc with number", "1.2.3-rc1", true},
		{"alpha version", "1.2.3-alpha", true},
		{"alpha with number", "1.2.3-alpha2", true},
		{"release version", "1.2.3", false},
		{"dev version", "dev", false},
		{"dev with commit", "dev-abc123", false},
		{"empty version", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			origVersion := Version
			
			// Set test value
			Version = tt.version
			
			// Test
			result := IsPreRelease()
			assert.Equal(t, tt.expected, result)
			
			// Restore original value
			Version = origVersion
		})
	}
}

func TestBuildDate(t *testing.T) {
	tests := []struct {
		name      string
		buildTime string
		expectZero bool
	}{
		{
			name:      "valid RFC3339 time",
			buildTime: "2023-01-01T12:00:00Z",
			expectZero: false,
		},
		{
			name:      "valid RFC3339 time with timezone",
			buildTime: "2023-01-01T12:00:00-07:00",
			expectZero: false,
		},
		{
			name:      "unknown build time",
			buildTime: "unknown",
			expectZero: true,
		},
		{
			name:      "invalid time format",
			buildTime: "2023-01-01 12:00:00",
			expectZero: true,
		},
		{
			name:      "empty build time",
			buildTime: "",
			expectZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			origBuildTime := BuildTime
			
			// Set test value
			BuildTime = tt.buildTime
			
			// Test
			result := BuildDate()
			
			if tt.expectZero {
				assert.True(t, result.IsZero(), "Should return zero time")
			} else {
				assert.False(t, result.IsZero(), "Should return valid time")
				
				// For valid times, verify it parses correctly
				expected, err := time.Parse(time.RFC3339, tt.buildTime)
				assert.NoError(t, err)
				assert.Equal(t, expected, result)
			}
			
			// Restore original value
			BuildTime = origBuildTime
		})
	}
}

func TestVersionVariablesDefault(t *testing.T) {
	// Test that version variables have reasonable defaults
	// Note: These might be overridden at build time, so we test the fallbacks
	
	// Save originals
	origVersion := Version
	origGitCommit := GitCommit
	origBuildTime := BuildTime
	origGitTag := GitTag
	origGitDirty := GitDirty
	
	// Reset to defaults (what they should be without build-time injection)
	Version = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
	GitTag = ""
	GitDirty = ""
	
	// Test defaults
	assert.Equal(t, "dev", Version)
	assert.Equal(t, "unknown", GitCommit)
	assert.Equal(t, "unknown", BuildTime)
	assert.Equal(t, "", GitTag)
	assert.Equal(t, "", GitDirty)
	
	// Test that default Info() works
	info := Info()
	assert.Equal(t, "dev", info)
	
	// Test that default IsDevBuild() works
	assert.True(t, IsDevBuild())
	
	// Test that default IsPreRelease() works
	assert.False(t, IsPreRelease())
	
	// Test that default BuildDate() works
	buildDate := BuildDate()
	assert.True(t, buildDate.IsZero())
	
	// Restore originals
	Version = origVersion
	GitCommit = origGitCommit
	BuildTime = origBuildTime
	GitTag = origGitTag
	GitDirty = origGitDirty
}

func TestInfoEdgeCases(t *testing.T) {
	// Save original values
	origVersion := Version
	origGitCommit := GitCommit
	origGitDirty := GitDirty
	
	t.Run("dev with unknown commit", func(t *testing.T) {
		Version = "dev"
		GitCommit = "unknown"
		GitDirty = ""
		
		result := Info()
		assert.Equal(t, "dev", result)
	})
	
	t.Run("dev with empty commit", func(t *testing.T) {
		Version = "dev"
		GitCommit = ""
		GitDirty = ""
		
		result := Info()
		// Empty commit is treated as unknown, so no suffix should be added
		assert.Equal(t, "dev", result)
	})
	
	t.Run("dev with dirty flag but no commit", func(t *testing.T) {
		Version = "dev"
		GitCommit = "unknown"
		GitDirty = "true"
		
		result := Info()
		assert.Equal(t, "dev", result)
	})
	
	t.Run("empty version", func(t *testing.T) {
		Version = ""
		GitCommit = "abc123"
		GitDirty = ""
		
		result := Info()
		assert.Equal(t, "", result)
	})
	
	// Restore original values
	Version = origVersion
	GitCommit = origGitCommit
	GitDirty = origGitDirty
}

func TestDetailedInfoFormatting(t *testing.T) {
	// Save originals
	origVersion := Version
	origGitCommit := GitCommit
	origGitTag := GitTag
	origBuildTime := BuildTime
	
	// Set comprehensive test values
	Version = "1.2.3-beta1"
	GitCommit = "abcdef1234567890123456789012345678901234"
	GitTag = "v1.2.3-beta1"
	BuildTime = "2023-12-25T10:30:45Z"
	
	result := DetailedInfo()
	lines := strings.Split(result, "\n")
	
	// Test formatting of each line
	var versionLine, commitLine, tagLine, builtLine, goLine, osLine string
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "Version:"):
			versionLine = line
		case strings.HasPrefix(line, "Git commit:"):
			commitLine = line
		case strings.HasPrefix(line, "Git tag:"):
			tagLine = line
		case strings.HasPrefix(line, "Built:"):
			builtLine = line
		case strings.HasPrefix(line, "Go version:"):
			goLine = line
		case strings.HasPrefix(line, "OS/Arch:"):
			osLine = line
		}
	}
	
	// Test line formats
	assert.Contains(t, versionLine, "Version:    1.2.3-beta1")
	assert.Contains(t, commitLine, "Git commit: abcdef1234567890123456789012345678901234")
	assert.Contains(t, tagLine, "Git tag:    v1.2.3-beta1")
	assert.Contains(t, builtLine, "Built:      2023-12-25T10:30:45Z")
	assert.Contains(t, goLine, "Go version:")
	assert.Contains(t, osLine, "OS/Arch:")
	
	// Restore originals
	Version = origVersion
	GitCommit = origGitCommit
	GitTag = origGitTag
	BuildTime = origBuildTime
}

func TestConcurrentAccess(t *testing.T) {
	// Test that version functions are safe for concurrent access
	const numGoroutines = 100
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			// Call all version functions
			Info()
			DetailedInfo()
			IsDevBuild()
			IsPreRelease()
			BuildDate()
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// If we get here without deadlock or panic, the test passes
}

// Benchmark tests
func BenchmarkInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Info()
	}
}

func BenchmarkDetailedInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DetailedInfo()
	}
}

func BenchmarkIsDevBuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsDevBuild()
	}
}

func BenchmarkIsPreRelease(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsPreRelease()
	}
}

func BenchmarkBuildDate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildDate()
	}
}

func TestVersionConsistency(t *testing.T) {
	// Test that repeated calls return consistent results
	for i := 0; i < 10; i++ {
		info1 := Info()
		info2 := Info()
		assert.Equal(t, info1, info2, "Info() should return consistent results")
		
		detailed1 := DetailedInfo()
		detailed2 := DetailedInfo()
		assert.Equal(t, detailed1, detailed2, "DetailedInfo() should return consistent results")
		
		isDev1 := IsDevBuild()
		isDev2 := IsDevBuild()
		assert.Equal(t, isDev1, isDev2, "IsDevBuild() should return consistent results")
		
		isPre1 := IsPreRelease()
		isPre2 := IsPreRelease()
		assert.Equal(t, isPre1, isPre2, "IsPreRelease() should return consistent results")
		
		buildDate1 := BuildDate()
		buildDate2 := BuildDate()
		assert.Equal(t, buildDate1, buildDate2, "BuildDate() should return consistent results")
	}
}