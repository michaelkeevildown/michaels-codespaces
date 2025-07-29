package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/michaelkeevildown/mcs/internal/version"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	// Save original version info functions if any
	originalVersion := version.Info()
	originalDetailed := version.DetailedInfo()

	tests := []struct {
		name           string
		flags          map[string]string
		expectedOutput []string
		notExpected    []string
	}{
		{
			name: "Basic version",
			expectedOutput: []string{
				"mcs version",
			},
		},
		{
			name: "Detailed version",
			flags: map[string]string{
				"detailed": "true",
			},
			expectedOutput: []string{
				originalDetailed,
			},
		},
		{
			name: "Short flag -d",
			flags: map[string]string{
				"d": "true",
			},
			expectedOutput: []string{
				originalDetailed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := VersionCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Set flags
			for flag, value := range tt.flags {
				cmd.Flags().Set(flag, value)
			}

			// Execute command
			err := cmd.Execute()
			assert.NoError(t, err)

			output := buf.String()
			
			// Check expected output
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}

			// Check not expected
			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, output, notExpected, "Output should not contain: %s", notExpected)
			}
		})
	}
}

func TestVersionCommand_Help(t *testing.T) {
	cmd := VersionCommand()
	
	assert.Equal(t, "version", cmd.Use)
	assert.Contains(t, cmd.Short, "Display version information")
	assert.Contains(t, cmd.Long, "Display version information for MCS")
}

func TestVersionCommand_Flags(t *testing.T) {
	cmd := VersionCommand()
	
	// Check detailed flag
	detailedFlag := cmd.Flags().Lookup("detailed")
	assert.NotNil(t, detailedFlag)
	assert.Equal(t, "d", detailedFlag.Shorthand)
	assert.Equal(t, "Show detailed version information", detailedFlag.Usage)
	
	// Check default value
	detailed, _ := cmd.Flags().GetBool("detailed")
	assert.False(t, detailed)
}

func TestVersionOutput_Format(t *testing.T) {
	tests := []struct {
		name           string
		detailed       bool
		checkFormat    func(t *testing.T, output string)
	}{
		{
			name:     "Basic format",
			detailed: false,
			checkFormat: func(t *testing.T, output string) {
				// Should be in format: "mcs version X.Y.Z"
				assert.True(t, strings.HasPrefix(output, "mcs version"))
				assert.True(t, strings.HasSuffix(output, "\n"))
			},
		},
		{
			name:     "Detailed format",
			detailed: true,
			checkFormat: func(t *testing.T, output string) {
				// Detailed output depends on version.DetailedInfo()
				// Just check it's not empty
				assert.NotEmpty(t, output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := VersionCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			
			if tt.detailed {
				cmd.Flags().Set("detailed", "true")
			}
			
			err := cmd.Execute()
			assert.NoError(t, err)
			
			output := buf.String()
			tt.checkFormat(t, output)
		})
	}
}

// Mock version info for testing
type MockVersionInfo struct {
	version  string
	detailed string
}

func (m *MockVersionInfo) Info() string {
	return m.version
}

func (m *MockVersionInfo) DetailedInfo() string {
	return m.detailed
}

func TestVersionCommand_WithMockVersion(t *testing.T) {
	// This test demonstrates how version info could be mocked
	// if the version package supported dependency injection
	
	mockVersions := []struct {
		version  string
		detailed string
	}{
		{
			version:  "1.2.3",
			detailed: "MCS version 1.2.3\nGit commit: abc123\nBuild date: 2024-01-01",
		},
		{
			version:  "2.0.0-beta",
			detailed: "MCS version 2.0.0-beta\nGit commit: def456\nBuild date: 2024-02-01",
		},
		{
			version:  "dev",
			detailed: "MCS version dev\nGit commit: HEAD\nBuild date: unknown",
		},
	}

	for _, mock := range mockVersions {
		t.Run(mock.version, func(t *testing.T) {
			// This would test with mocked version info
			// Currently skipped as version package doesn't support injection
			t.Skip("Version mocking not implemented")
		})
	}
}

// Benchmark version command
func BenchmarkVersionCommand(b *testing.B) {
	cmd := VersionCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	for i := 0; i < b.N; i++ {
		buf.Reset()
		cmd.Execute()
	}
}

func BenchmarkVersionCommand_Detailed(b *testing.B) {
	cmd := VersionCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.Flags().Set("detailed", "true")
	
	for i := 0; i < b.N; i++ {
		buf.Reset()
		cmd.Execute()
	}
}

// Test edge cases
func TestVersionCommand_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "Extra arguments ignored",
			args:        []string{"extra", "args"},
			expectError: false,
		},
		{
			name:        "Invalid flag",
			args:        []string{"--invalid-flag"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := VersionCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test concurrent execution
func TestVersionCommand_Concurrent(t *testing.T) {
	// Version command should be safe to run concurrently
	cmd := VersionCommand()
	
	// Run multiple instances concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			err := cmd.Execute()
			assert.NoError(t, err)
			done <- true
		}()
	}
	
	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}