package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/michaelkeevildown/mcs/internal/codespace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCodespaceManager for testing
type MockManager struct {
	mock.Mock
}

func (m *MockManager) Start(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockManager) Stop(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockManager) Remove(ctx context.Context, name string, force bool) error {
	args := m.Called(ctx, name, force)
	return args.Error(0)
}

func (m *MockManager) Get(ctx context.Context, name string) (*codespace.Codespace, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*codespace.Codespace), args.Error(1)
}

// MockDockerClient for testing
type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDockerClient) GetContainerByName(ctx context.Context, name string) (*Container, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Container), args.Error(1)
}

func (m *MockDockerClient) StopContainer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDockerClient) RemoveContainer(ctx context.Context, id string, force bool) error {
	args := m.Called(ctx, id, force)
	return args.Error(0)
}

// Container struct for testing
type Container struct {
	ID    string
	State string
}

// MockComposeExecutor for testing
type MockComposeExecutor struct {
	mock.Mock
}

func (m *MockComposeExecutor) Build(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockComposeExecutor) Up(ctx context.Context, detached bool) error {
	args := m.Called(ctx, detached)
	return args.Error(0)
}

func TestStartCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		setupMocks    func(*MockManager)
		expectedError bool
		errorContains string
		checkOutput   bool
		outputContains []string
	}{
		{
			name: "Start success with URLs",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager) {
				m.On("Start", mock.Anything, "test-codespace").Return(nil)
				m.On("Get", mock.Anything, "test-codespace").Return(&codespace.Codespace{
					Name:      "test-codespace",
					VSCodeURL: "http://localhost:8443",
					AppURL:    "http://localhost:3000",
					Password:  "test123",
				}, nil)
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"Started test-codespace",
				"VS Code: http://localhost:8443",
				"App: http://localhost:3000",
				"Password: test123",
			},
		},
		{
			name: "Start success without password",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager) {
				m.On("Start", mock.Anything, "test-codespace").Return(nil)
				m.On("Get", mock.Anything, "test-codespace").Return(&codespace.Codespace{
					Name:      "test-codespace",
					VSCodeURL: "http://localhost:8443",
					AppURL:    "http://localhost:3000",
					Password:  "",
				}, nil)
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"Started test-codespace",
				"VS Code: http://localhost:8443",
				"App: http://localhost:3000",
			},
		},
		{
			name: "Start failure",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager) {
				m.On("Start", mock.Anything, "test-codespace").Return(errors.New("start failed"))
			},
			expectedError: true,
			errorContains: "start failed",
		},
		{
			name: "Get info failure after start",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager) {
				m.On("Start", mock.Anything, "test-codespace").Return(nil)
				m.On("Get", mock.Anything, "test-codespace").Return(nil, errors.New("not found"))
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"Started test-codespace",
			},
		},
		{
			name:          "No arguments",
			args:          []string{},
			setupMocks:    func(m *MockManager) {},
			expectedError: true,
			errorContains: "accepts 1 arg(s), received 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(MockManager)
			if tt.setupMocks != nil {
				tt.setupMocks(mockManager)
			}

			cmd := StartCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

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
					assert.Contains(t, output, expected)
				}
			}

			mockManager.AssertExpectations(t)
		})
	}
}

func TestStopCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		setupMocks    func(*MockManager)
		expectedError bool
		errorContains string
		checkOutput   bool
		outputContains []string
	}{
		{
			name: "Stop success",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager) {
				m.On("Stop", mock.Anything, "test-codespace").Return(nil)
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"Stopped test-codespace",
			},
		},
		{
			name: "Stop failure",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager) {
				m.On("Stop", mock.Anything, "test-codespace").Return(errors.New("stop failed"))
			},
			expectedError: true,
			errorContains: "stop failed",
		},
		{
			name:          "No arguments",
			args:          []string{},
			setupMocks:    func(m *MockManager) {},
			expectedError: true,
			errorContains: "accepts 1 arg(s), received 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(MockManager)
			if tt.setupMocks != nil {
				tt.setupMocks(mockManager)
			}

			cmd := StopCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

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
					assert.Contains(t, output, expected)
				}
			}

			mockManager.AssertExpectations(t)
		})
	}
}

func TestRestartCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		setupMocks    func(*MockManager)
		expectedError bool
		errorContains string
		checkOutput   bool
		outputContains []string
	}{
		{
			name: "Restart success",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager) {
				m.On("Stop", mock.Anything, "test-codespace").Return(nil)
				m.On("Start", mock.Anything, "test-codespace").Return(nil)
				m.On("Get", mock.Anything, "test-codespace").Return(&codespace.Codespace{
					Name:      "test-codespace",
					VSCodeURL: "http://localhost:8443",
					AppURL:    "http://localhost:3000",
					Password:  "",
				}, nil)
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"Stopped test-codespace",
				"Restarted test-codespace",
				"VS Code: http://localhost:8443",
				"App: http://localhost:3000",
			},
		},
		{
			name: "Restart with stop failure (not found) continues to start",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager) {
				m.On("Stop", mock.Anything, "test-codespace").Return(errors.New("not found"))
				m.On("Start", mock.Anything, "test-codespace").Return(nil)
				m.On("Get", mock.Anything, "test-codespace").Return(&codespace.Codespace{
					Name:      "test-codespace",
					VSCodeURL: "http://localhost:8443",
					AppURL:    "http://localhost:3000",
					Password:  "",
				}, nil)
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"Restarted test-codespace",
			},
		},
		{
			name: "Restart with start failure",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager) {
				m.On("Stop", mock.Anything, "test-codespace").Return(nil)
				m.On("Start", mock.Anything, "test-codespace").Return(errors.New("start failed"))
			},
			expectedError: true,
			errorContains: "start failed",
		},
		{
			name:          "No arguments",
			args:          []string{},
			setupMocks:    func(m *MockManager) {},
			expectedError: true,
			errorContains: "accepts 1 arg(s), received 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(MockManager)
			if tt.setupMocks != nil {
				tt.setupMocks(mockManager)
			}

			cmd := RestartCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

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
					assert.Contains(t, output, expected)
				}
			}

			mockManager.AssertExpectations(t)
		})
	}
}

func TestRemoveCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		flags         map[string]string
		userInput     string
		setupMocks    func(*MockManager)
		expectedError bool
		errorContains string
		checkOutput   bool
		outputContains []string
	}{
		{
			name: "Remove with confirmation yes",
			args: []string{"test-codespace"},
			userInput: "y\n",
			setupMocks: func(m *MockManager) {
				m.On("Remove", mock.Anything, "test-codespace", false).Return(nil)
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"This will permanently delete the codespace 'test-codespace'",
				"Are you sure?",
				"Removed test-codespace",
			},
		},
		{
			name: "Remove with confirmation no",
			args: []string{"test-codespace"},
			userInput: "n\n",
			setupMocks: func(m *MockManager) {
				// Remove should not be called
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"This will permanently delete the codespace 'test-codespace'",
				"Are you sure?",
				"Cancelled.",
			},
		},
		{
			name: "Remove with force flag",
			args: []string{"test-codespace"},
			flags: map[string]string{
				"force": "true",
			},
			setupMocks: func(m *MockManager) {
				m.On("Remove", mock.Anything, "test-codespace", true).Return(nil)
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"Removed test-codespace",
			},
		},
		{
			name: "Remove failure",
			args: []string{"test-codespace"},
			flags: map[string]string{
				"force": "true",
			},
			setupMocks: func(m *MockManager) {
				m.On("Remove", mock.Anything, "test-codespace", true).Return(errors.New("remove failed"))
			},
			expectedError: true,
			errorContains: "remove failed",
		},
		{
			name: "Remove with default (empty) input",
			args: []string{"test-codespace"},
			userInput: "\n",
			setupMocks: func(m *MockManager) {
				// Remove should not be called (default is NO)
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"Cancelled.",
			},
		},
		{
			name:          "No arguments",
			args:          []string{},
			setupMocks:    func(m *MockManager) {},
			expectedError: true,
			errorContains: "accepts 1 arg(s), received 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(MockManager)
			if tt.setupMocks != nil {
				tt.setupMocks(mockManager)
			}

			cmd := RemoveCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			// Set flags
			for flag, value := range tt.flags {
				cmd.Flags().Set(flag, value)
			}

			// Mock stdin for user input
			if tt.userInput != "" {
				oldStdin := cmd.InOrStdin()
				cmd.SetIn(strings.NewReader(tt.userInput))
				defer cmd.SetIn(oldStdin)
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
					assert.Contains(t, output, expected)
				}
			}

			mockManager.AssertExpectations(t)
		})
	}
}

func TestRebuildCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		setupMocks    func(*MockManager, *MockDockerClient, *MockComposeExecutor)
		expectedError bool
		errorContains string
		checkOutput   bool
		outputContains []string
	}{
		{
			name: "Rebuild success with running container",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager, d *MockDockerClient, c *MockComposeExecutor) {
				m.On("Get", mock.Anything, "test-codespace").Return(&codespace.Codespace{
					Name:      "test-codespace",
					Path:      "/path/to/codespace",
					VSCodeURL: "http://localhost:8443",
					AppURL:    "http://localhost:3000",
					Password:  "",
				}, nil)
				d.On("Close").Return(nil)
				c.On("Build", mock.Anything).Return(nil)
				d.On("GetContainerByName", mock.Anything, "test-codespace-dev").Return(&Container{
					ID:    "container123",
					State: "running",
				}, nil)
				d.On("StopContainer", mock.Anything, "container123").Return(nil)
				d.On("RemoveContainer", mock.Anything, "container123", true).Return(nil)
				c.On("Up", mock.Anything, true).Return(nil)
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"Rebuilding image for test-codespace",
				"Image rebuilt successfully",
				"Stopped old container",
				"Removed old container",
				"Rebuilt and restarted test-codespace",
				"VS Code: http://localhost:8443",
				"App: http://localhost:3000",
			},
		},
		{
			name: "Rebuild success without existing container",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager, d *MockDockerClient, c *MockComposeExecutor) {
				m.On("Get", mock.Anything, "test-codespace").Return(&codespace.Codespace{
					Name:      "test-codespace",
					Path:      "/path/to/codespace",
					VSCodeURL: "http://localhost:8443",
					AppURL:    "http://localhost:3000",
					Password:  "",
				}, nil)
				d.On("Close").Return(nil)
				c.On("Build", mock.Anything).Return(nil)
				d.On("GetContainerByName", mock.Anything, "test-codespace-dev").Return(nil, errors.New("not found"))
				c.On("Up", mock.Anything, true).Return(nil)
			},
			expectedError: false,
			checkOutput:   true,
			outputContains: []string{
				"Rebuilding image for test-codespace",
				"Image rebuilt successfully",
				"Rebuilt and restarted test-codespace",
			},
		},
		{
			name: "Codespace not found",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager, d *MockDockerClient, c *MockComposeExecutor) {
				m.On("Get", mock.Anything, "test-codespace").Return(nil, errors.New("not found"))
			},
			expectedError: true,
			errorContains: "codespace not found",
		},
		{
			name: "Build failure",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockManager, d *MockDockerClient, c *MockComposeExecutor) {
				m.On("Get", mock.Anything, "test-codespace").Return(&codespace.Codespace{
					Name: "test-codespace",
					Path: "/path/to/codespace",
				}, nil)
				d.On("Close").Return(nil)
				c.On("Build", mock.Anything).Return(errors.New("build failed"))
			},
			expectedError: true,
			errorContains: "failed to rebuild image",
		},
		{
			name:          "No arguments",
			args:          []string{},
			setupMocks:    func(m *MockManager, d *MockDockerClient, c *MockComposeExecutor) {},
			expectedError: true,
			errorContains: "accepts 1 arg(s), received 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockManager := new(MockManager)
			mockDocker := new(MockDockerClient)
			mockCompose := new(MockComposeExecutor)

			if tt.setupMocks != nil {
				tt.setupMocks(mockManager, mockDocker, mockCompose)
			}

			cmd := RebuildCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

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
					assert.Contains(t, output, expected)
				}
			}

			mockManager.AssertExpectations(t)
			mockDocker.AssertExpectations(t)
			mockCompose.AssertExpectations(t)
		})
	}
}

func TestUpdateCommand(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]string
		expectedError bool
		errorContains string
	}{
		{
			name: "Update check only",
			flags: map[string]string{
				"check": "true",
			},
			expectedError: true,
			errorContains: "source not found", // Expected as we're not in actual MCS environment
		},
		{
			name:          "Full update",
			flags:         map[string]string{},
			expectedError: true,
			errorContains: "source not found", // Expected as we're not in actual MCS environment
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := UpdateCommand()
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
		})
	}
}

func TestCommandHelp(t *testing.T) {
	commands := []struct {
		name     string
		cmd      func() *cobra.Command
		useText  string
		shortText string
	}{
		{
			name:      "start",
			cmd:       StartCommand,
			useText:   "start <name>",
			shortText: "Start a codespace",
		},
		{
			name:      "stop",
			cmd:       StopCommand,
			useText:   "stop <name>",
			shortText: "Stop a codespace",
		},
		{
			name:      "restart",
			cmd:       RestartCommand,
			useText:   "restart <name>",
			shortText: "Restart a codespace",
		},
		{
			name:      "rebuild",
			cmd:       RebuildCommand,
			useText:   "rebuild <name>",
			shortText: "Rebuild and recreate a codespace container",
		},
		{
			name:      "remove",
			cmd:       RemoveCommand,
			useText:   "remove <name>",
			shortText: "Remove a codespace",
		},
		{
			name:      "update",
			cmd:       UpdateCommand,
			useText:   "update",
			shortText: "Update MCS to the latest version",
		},
	}

	for _, tt := range commands {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd()
			assert.Equal(t, tt.useText, cmd.Use)
			assert.Contains(t, cmd.Short, tt.shortText)
		})
	}
}

func TestRemoveCommandAliases(t *testing.T) {
	cmd := RemoveCommand()
	assert.Contains(t, cmd.Aliases, "rm")
	assert.Contains(t, cmd.Aliases, "delete")
}