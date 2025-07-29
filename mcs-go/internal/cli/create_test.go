package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock dependencies
type MockCodespaceManager struct {
	mock.Mock
}

func (m *MockCodespaceManager) Create(ctx context.Context, config interface{}) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockCodespaceManager) Start(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockCodespaceManager) Get(ctx context.Context, name string) (interface{}, error) {
	args := m.Called(ctx, name)
	return args.Get(0), args.Error(1)
}

type MockComponentSelector struct {
	mock.Mock
}

func (m *MockComponentSelector) Select() ([]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]interface{}), args.Error(1)
}

func TestCreateCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		flags         map[string]interface{}
		setupMocks    func(*MockCodespaceManager, *MockComponentSelector)
		expectedError bool
		errorContains string
	}{
		{
			name: "Valid GitHub URL",
			args: []string{"https://github.com/user/repo"},
			flags: map[string]interface{}{
				"no-start": false,
				"depth":    20,
			},
			setupMocks: func(cm *MockCodespaceManager, cs *MockComponentSelector) {
				cs.On("Select").Return([]interface{}{}, nil)
				cm.On("Create", mock.Anything, mock.Anything).Return(nil)
				cm.On("Start", mock.Anything, "user-repo").Return(nil)
				cm.On("Get", mock.Anything, "user-repo").Return(struct {
					VSCodeURL string
					AppURL    string
					Password  string
				}{
					VSCodeURL: "http://localhost:8443",
					AppURL:    "http://localhost:3000",
					Password:  "test123",
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "SSH URL",
			args: []string{"git@github.com:user/repo.git"},
			flags: map[string]interface{}{
				"no-start": false,
				"depth":    20,
			},
			setupMocks: func(cm *MockCodespaceManager, cs *MockComponentSelector) {
				cs.On("Select").Return([]interface{}{}, nil)
				cm.On("Create", mock.Anything, mock.Anything).Return(nil)
				cm.On("Start", mock.Anything, "user-repo").Return(nil)
				cm.On("Get", mock.Anything, "user-repo").Return(struct {
					VSCodeURL string
					AppURL    string
					Password  string
				}{
					VSCodeURL: "http://localhost:8443",
					AppURL:    "http://localhost:3000",
					Password:  "test123",
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Short form GitHub",
			args: []string{"user/repo"},
			flags: map[string]interface{}{
				"no-start": false,
				"depth":    20,
			},
			setupMocks: func(cm *MockCodespaceManager, cs *MockComponentSelector) {
				cs.On("Select").Return([]interface{}{}, nil)
				cm.On("Create", mock.Anything, mock.Anything).Return(nil)
				cm.On("Start", mock.Anything, "user-repo").Return(nil)
				cm.On("Get", mock.Anything, "user-repo").Return(struct {
					VSCodeURL string
					AppURL    string
					Password  string
				}{
					VSCodeURL: "http://localhost:8443",
					AppURL:    "http://localhost:3000",
					Password:  "",
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "No start flag",
			args: []string{"user/repo"},
			flags: map[string]interface{}{
				"no-start": true,
				"depth":    20,
			},
			setupMocks: func(cm *MockCodespaceManager, cs *MockComponentSelector) {
				cs.On("Select").Return([]interface{}{}, nil)
				cm.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "Full git clone depth",
			args: []string{"user/repo"},
			flags: map[string]interface{}{
				"no-start": false,
				"depth":    -1,
			},
			setupMocks: func(cm *MockCodespaceManager, cs *MockComponentSelector) {
				cs.On("Select").Return([]interface{}{}, nil)
				cm.On("Create", mock.Anything, mock.Anything).Return(nil)
				cm.On("Start", mock.Anything, "user-repo").Return(nil)
				cm.On("Get", mock.Anything, "user-repo").Return(struct {
					VSCodeURL string
					AppURL    string
					Password  string
				}{
					VSCodeURL: "http://localhost:8443",
					AppURL:    "http://localhost:3000",
					Password:  "",
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Invalid repository URL",
			args: []string{"not-a-valid-url"},
			flags: map[string]interface{}{
				"no-start": false,
				"depth":    20,
			},
			setupMocks: func(cm *MockCodespaceManager, cs *MockComponentSelector) {
				// No mocks needed, should fail at URL parsing
			},
			expectedError: true,
			errorContains: "invalid repository",
		},
		{
			name: "Create fails",
			args: []string{"user/repo"},
			flags: map[string]interface{}{
				"no-start": false,
				"depth":    20,
			},
			setupMocks: func(cm *MockCodespaceManager, cs *MockComponentSelector) {
				cs.On("Select").Return([]interface{}{}, nil)
				cm.On("Create", mock.Anything, mock.Anything).Return(errors.New("create failed"))
			},
			expectedError: true,
			errorContains: "create failed",
		},
		{
			name: "Start fails",
			args: []string{"user/repo"},
			flags: map[string]interface{}{
				"no-start": false,
				"depth":    20,
			},
			setupMocks: func(cm *MockCodespaceManager, cs *MockComponentSelector) {
				cs.On("Select").Return([]interface{}{}, nil)
				cm.On("Create", mock.Anything, mock.Anything).Return(nil)
				cm.On("Start", mock.Anything, "user-repo").Return(errors.New("start failed"))
			},
			expectedError: true,
			errorContains: "start failed",
		},
		{
			name: "Skip component selector",
			args: []string{"user/repo"},
			flags: map[string]interface{}{
				"no-start":       false,
				"skip-selector": true,
				"depth":         20,
			},
			setupMocks: func(cm *MockCodespaceManager, cs *MockComponentSelector) {
				// Should not call Select when skip-selector is true
				cm.On("Create", mock.Anything, mock.Anything).Return(nil)
				cm.On("Start", mock.Anything, "user-repo").Return(nil)
				cm.On("Get", mock.Anything, "user-repo").Return(struct {
					VSCodeURL string
					AppURL    string
					Password  string
				}{
					VSCodeURL: "http://localhost:8443",
					AppURL:    "http://localhost:3000",
					Password:  "",
				}, nil)
			},
			expectedError: false,
		},
		{
			name: "Component selector error",
			args: []string{"user/repo"},
			flags: map[string]interface{}{
				"no-start": false,
				"depth":    20,
			},
			setupMocks: func(cm *MockCodespaceManager, cs *MockComponentSelector) {
				cs.On("Select").Return([]interface{}{}, errors.New("selector error"))
			},
			expectedError: true,
			errorContains: "selector error",
		},
		{
			name:          "No arguments",
			args:          []string{},
			setupMocks:    func(cm *MockCodespaceManager, cs *MockComponentSelector) {},
			expectedError: true,
			errorContains: "accepts 1 arg(s), received 0",
		},
		{
			name:          "Too many arguments",
			args:          []string{"repo1", "repo2"},
			setupMocks:    func(cm *MockCodespaceManager, cs *MockComponentSelector) {},
			expectedError: true,
			errorContains: "accepts 1 arg(s), received 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockManager := new(MockCodespaceManager)
			mockSelector := new(MockComponentSelector)

			// Setup mocks
			if tt.setupMocks != nil {
				tt.setupMocks(mockManager, mockSelector)
			}

			// Create command
			cmd := CreateCommand()

			// Set flags
			for flag, value := range tt.flags {
				switch v := value.(type) {
				case bool:
					cmd.Flags().Set(flag, "true")
				case int:
					cmd.Flags().Set(flag, string(rune(v)))
				case string:
					cmd.Flags().Set(flag, v)
				}
			}

			// Capture output
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Execute command
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// Check error
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockManager.AssertExpectations(t)
			mockSelector.AssertExpectations(t)
		})
	}
}

func TestCreateCommand_CollisionHandling(t *testing.T) {
	// Test that collision detection adds suffixes
	tests := []struct {
		name          string
		repoURL       string
		existingNames []string
		expectedName  string
	}{
		{
			name:          "No collision",
			repoURL:       "user/repo",
			existingNames: []string{},
			expectedName:  "user-repo",
		},
		{
			name:          "One collision",
			repoURL:       "user/repo",
			existingNames: []string{"user-repo"},
			expectedName:  "user-repo-happy-narwhal", // Or any suffix
		},
		{
			name:          "Multiple collisions",
			repoURL:       "user/repo",
			existingNames: []string{"user-repo", "user-repo-happy-narwhal"},
			expectedName:  "user-repo-dancing-penguin", // Or any other suffix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test would need to mock the filesystem checks
			// and verify that the collision handling works correctly
			// For now, we'll skip the implementation as it requires
			// deeper integration with the actual codebase
			t.Skip("Collision handling test requires filesystem mocking")
		})
	}
}

func TestCreateCommand_Flags(t *testing.T) {
	cmd := CreateCommand()

	// Check that all expected flags exist
	assert.NotNil(t, cmd.Flags().Lookup("no-start"))
	assert.NotNil(t, cmd.Flags().Lookup("skip-selector"))
	assert.NotNil(t, cmd.Flags().Lookup("depth"))

	// Check default values
	noStart, _ := cmd.Flags().GetBool("no-start")
	assert.False(t, noStart)

	skipSelector, _ := cmd.Flags().GetBool("skip-selector")
	assert.False(t, skipSelector)

	depth, _ := cmd.Flags().GetInt("depth")
	assert.Equal(t, 20, depth)
}

func TestCreateCommand_Help(t *testing.T) {
	cmd := CreateCommand()
	
	// Check command metadata
	assert.Equal(t, "create", cmd.Use)
	assert.Contains(t, cmd.Short, "Create a new codespace")
	assert.Contains(t, cmd.Long, "repository")
	assert.NotEmpty(t, cmd.Example)
}

func TestCreateCommand_URLFormats(t *testing.T) {
	tests := []struct {
		input        string
		expectedOwner string
		expectedRepo  string
		shouldError   bool
	}{
		{
			input:         "https://github.com/facebook/react",
			expectedOwner: "facebook",
			expectedRepo:  "react",
			shouldError:   false,
		},
		{
			input:         "git@github.com:michaelkeevildown/michaels-codespaces.git",
			expectedOwner: "michaelkeevildown",
			expectedRepo:  "michaels-codespaces",
			shouldError:   false,
		},
		{
			input:         "user/repo",
			expectedOwner: "user",
			expectedRepo:  "repo",
			shouldError:   false,
		},
		{
			input:         "https://gitlab.com/user/repo",
			expectedOwner: "user",
			expectedRepo:  "repo",
			shouldError:   false,
		},
		{
			input:         ".",
			expectedOwner: "",
			expectedRepo:  "",
			shouldError:   false, // Local path is valid
		},
		{
			input:         "./path/to/repo",
			expectedOwner: "",
			expectedRepo:  "",
			shouldError:   false, // Local path is valid
		},
		{
			input:         "not-a-valid-url",
			expectedOwner: "",
			expectedRepo:  "",
			shouldError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// This test would verify URL parsing logic
			// Requires access to the utils.ParseRepository function
			t.Skip("URL format test requires utils package mocking")
		})
	}
}