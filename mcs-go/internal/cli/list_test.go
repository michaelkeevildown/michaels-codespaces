package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/michaelkeevildown/mcs/internal/codespace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockListManager for testing list command
type MockListManager struct {
	mock.Mock
}

func (m *MockListManager) List(ctx context.Context) ([]codespace.Codespace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]codespace.Codespace), args.Error(1)
}

func TestListCommand(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name           string
		flags          map[string]string
		setupMocks     func(*MockListManager)
		expectedError  bool
		errorContains  string
		checkOutput    bool
		outputContains []string
		outputNotContains []string
	}{
		{
			name: "List running codespaces only (default)",
			setupMocks: func(m *MockListManager) {
				m.On("List", mock.Anything).Return([]codespace.Codespace{
					{
						Name:       "project-one",
						Status:     "running",
						Repository: "https://github.com/user/project-one",
						VSCodeURL:  "http://localhost:8443",
						AppURL:     "http://localhost:3000",
						CreatedAt:  now.Add(-2 * time.Hour),
					},
					{
						Name:       "project-two",
						Status:     "stopped",
						Repository: "git@github.com:user/project-two.git",
						VSCodeURL:  "http://localhost:8444",
						AppURL:     "http://localhost:3001",
						CreatedAt:  now.Add(-24 * time.Hour),
					},
				}, nil)
			},
			checkOutput: true,
			outputContains: []string{
				"NAME",
				"STATUS",
				"REPOSITORY",
				"PORTS",
				"CREATED",
				"project-one",
				"● running",
				"user/project-one",
				"8443, 3000",
				"2h ago",
			},
			outputNotContains: []string{
				"project-two", // Should be filtered out
			},
		},
		{
			name: "List all codespaces with --all flag",
			flags: map[string]string{
				"all": "true",
			},
			setupMocks: func(m *MockListManager) {
				m.On("List", mock.Anything).Return([]codespace.Codespace{
					{
						Name:       "project-one",
						Status:     "running",
						Repository: "https://github.com/user/project-one",
						VSCodeURL:  "http://localhost:8443",
						AppURL:     "http://localhost:3000",
						CreatedAt:  now.Add(-30 * time.Minute),
					},
					{
						Name:       "project-two",
						Status:     "stopped",
						Repository: "git@github.com:user/project-two.git",
						VSCodeURL:  "http://localhost:8444",
						AppURL:     "http://localhost:3001",
						CreatedAt:  now.Add(-3 * 24 * time.Hour),
					},
				}, nil)
			},
			checkOutput: true,
			outputContains: []string{
				"project-one",
				"● running",
				"project-two",
				"○ stopped",
				"30m ago",
				"3d ago",
			},
		},
		{
			name: "Empty list - no codespaces",
			setupMocks: func(m *MockListManager) {
				m.On("List", mock.Anything).Return([]codespace.Codespace{}, nil)
			},
			checkOutput: true,
			outputContains: []string{
				"No running codespaces found. Use --all to see all codespaces.",
			},
		},
		{
			name: "Empty list with --all flag",
			flags: map[string]string{
				"all": "true",
			},
			setupMocks: func(m *MockListManager) {
				m.On("List", mock.Anything).Return([]codespace.Codespace{}, nil)
			},
			checkOutput: true,
			outputContains: []string{
				"No codespaces found.",
			},
		},
		{
			name: "Simple format",
			flags: map[string]string{
				"format": "simple",
			},
			setupMocks: func(m *MockListManager) {
				m.On("List", mock.Anything).Return([]codespace.Codespace{
					{
						Name:   "project-one",
						Status: "running",
					},
					{
						Name:   "project-two",
						Status: "stopped",
					},
				}, nil)
			},
			checkOutput: true,
			outputContains: []string{
				"project-one (running)",
			},
			outputNotContains: []string{
				"project-two", // Filtered out in default view
				"NAME",        // No table headers in simple format
				"STATUS",
			},
		},
		{
			name: "JSON format (not implemented)",
			flags: map[string]string{
				"format": "json",
			},
			setupMocks: func(m *MockListManager) {
				m.On("List", mock.Anything).Return([]codespace.Codespace{
					{
						Name:   "project-one",
						Status: "running",
					},
				}, nil)
			},
			expectedError: true,
			errorContains: "JSON format not yet implemented",
		},
		{
			name: "List error",
			setupMocks: func(m *MockListManager) {
				m.On("List", mock.Anything).Return(nil, errors.New("failed to list"))
			},
			expectedError: true,
			errorContains: "failed to list codespaces",
		},
		{
			name: "Long repository URLs are truncated",
			setupMocks: func(m *MockListManager) {
				m.On("List", mock.Anything).Return([]codespace.Codespace{
					{
						Name:       "project",
						Status:     "running",
						Repository: "https://github.com/very-long-organization-name/very-long-repository-name-that-exceeds-the-limit",
						VSCodeURL:  "http://localhost:8443",
						AppURL:     "http://localhost:3000",
						CreatedAt:  now,
					},
				}, nil)
			},
			checkOutput: true,
			outputContains: []string{
				"very-long-organization-name/very-long-r...", // Should be truncated
			},
		},
		{
			name: "Different time formats",
			setupMocks: func(m *MockListManager) {
				m.On("List", mock.Anything).Return([]codespace.Codespace{
					{
						Name:      "project-1",
						Status:    "running",
						CreatedAt: now.Add(-30 * time.Second), // just now
					},
					{
						Name:      "project-2",
						Status:    "running",
						CreatedAt: now.Add(-45 * time.Minute), // 45m ago
					},
					{
						Name:      "project-3",
						Status:    "running",
						CreatedAt: now.Add(-5 * time.Hour), // 5h ago
					},
					{
						Name:      "project-4",
						Status:    "running",
						CreatedAt: now.Add(-5 * 24 * time.Hour), // 5d ago
					},
					{
						Name:      "project-5",
						Status:    "running",
						CreatedAt: now.Add(-30 * 24 * time.Hour), // Jan 02 format
					},
				}, nil)
			},
			checkOutput: true,
			outputContains: []string{
				"just now",
				"45m ago",
				"5h ago",
				"5d ago",
			},
		},
		{
			name: "Ports display for stopped codespace",
			flags: map[string]string{
				"all": "true",
			},
			setupMocks: func(m *MockListManager) {
				m.On("List", mock.Anything).Return([]codespace.Codespace{
					{
						Name:       "stopped-project",
						Status:     "stopped",
						Repository: "user/repo",
						VSCodeURL:  "http://localhost:8443",
						AppURL:     "http://localhost:3000",
						CreatedAt:  now,
					},
				}, nil)
			},
			checkOutput: true,
			outputContains: []string{
				"stopped-project",
				"○ stopped",
				"-", // Ports should show dash for stopped codespaces
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(MockListManager)
			if tt.setupMocks != nil {
				tt.setupMocks(mockManager)
			}

			cmd := ListCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Set flags
			for flag, value := range tt.flags {
				cmd.Flags().Set(flag, value)
			}

			err := cmd.Execute()

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.checkOutput {
				output := buf.String()
				for _, expected := range tt.outputContains {
					assert.Contains(t, output, expected, "Output should contain: %s", expected)
				}
				for _, notExpected := range tt.outputNotContains {
					assert.NotContains(t, output, notExpected, "Output should not contain: %s", notExpected)
				}
			}

			mockManager.AssertExpectations(t)
		})
	}
}

func TestListCommand_Aliases(t *testing.T) {
	cmd := ListCommand()
	assert.Contains(t, cmd.Aliases, "ls")
}

func TestListCommand_Flags(t *testing.T) {
	cmd := ListCommand()

	// Check that all expected flags exist
	assert.NotNil(t, cmd.Flags().Lookup("all"))
	assert.NotNil(t, cmd.Flags().Lookup("format"))

	// Check short flags
	allFlag := cmd.Flags().ShorthandLookup("a")
	assert.NotNil(t, allFlag)
	assert.Equal(t, "all", allFlag.Name)

	// Check default values
	format, _ := cmd.Flags().GetString("format")
	assert.Equal(t, "table", format)

	showAll, _ := cmd.Flags().GetBool("all")
	assert.False(t, showAll)
}

func TestTruncateRepo(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{
			input:    "https://github.com/user/repo",
			maxLen:   20,
			expected: "user/repo",
		},
		{
			input:    "git@github.com:user/repo.git",
			maxLen:   20,
			expected: "user/repo",
		},
		{
			input:    "user/very-long-repository-name-that-exceeds-limit",
			maxLen:   20,
			expected: "user/very-long-re...",
		},
		{
			input:    "short",
			maxLen:   20,
			expected: "short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateRepo(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractPort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "http://localhost:8443",
			expected: "8443",
		},
		{
			input:    "https://example.com:3000",
			expected: "3000",
		},
		{
			input:    "http://localhost",
			expected: "",
		},
		{
			input:    "invalid-url",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractPort(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Just now",
			input:    now.Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "Minutes ago",
			input:    now.Add(-45 * time.Minute),
			expected: "45m ago",
		},
		{
			name:     "Hours ago",
			input:    now.Add(-5 * time.Hour),
			expected: "5h ago",
		},
		{
			name:     "Days ago",
			input:    now.Add(-3 * 24 * time.Hour),
			expected: "3d ago",
		},
		{
			name:     "More than a week",
			input:    now.Add(-30 * 24 * time.Hour),
			expected: now.Add(-30 * 24 * time.Hour).Format("Jan 02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTime(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatPorts(t *testing.T) {
	tests := []struct {
		name     string
		cs       codespace.Codespace
		expected string
	}{
		{
			name: "Running with both ports",
			cs: codespace.Codespace{
				Status:    "running",
				VSCodeURL: "http://localhost:8443",
				AppURL:    "http://localhost:3000",
			},
			expected: "8443, 3000",
		},
		{
			name: "Running with VSCode port only",
			cs: codespace.Codespace{
				Status:    "running",
				VSCodeURL: "http://localhost:8443",
				AppURL:    "",
			},
			expected: "8443",
		},
		{
			name: "Stopped codespace",
			cs: codespace.Codespace{
				Status:    "stopped",
				VSCodeURL: "http://localhost:8443",
				AppURL:    "http://localhost:3000",
			},
			expected: "-",
		},
		{
			name: "Running with no ports",
			cs: codespace.Codespace{
				Status:    "running",
				VSCodeURL: "",
				AppURL:    "",
			},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPorts(tt.cs)
			assert.Equal(t, tt.expected, result)
		})
	}
}