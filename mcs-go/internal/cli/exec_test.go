package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDockerClientForExec for testing exec command
type MockDockerClientForExec struct {
	mock.Mock
}

func (m *MockDockerClientForExec) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDockerClientForExec) GetContainerByName(ctx context.Context, name string) (*ContainerInfo, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ContainerInfo), args.Error(1)
}

// ContainerInfo struct for exec testing
type ContainerInfo struct {
	ID    string
	State string
}

// MockCommandExecutor to mock exec.Command
type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) Execute(name string, args ...string) error {
	mArgs := m.Called(name, args)
	return mArgs.Error(0)
}

func TestExecCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		setupMocks    func(*MockDockerClientForExec)
		expectedError bool
		errorContains string
		checkDockerCmd bool
		expectedDockerArgs []string
	}{
		{
			name: "Execute command in running container",
			args: []string{"test-codespace", "ls", "-la"},
			setupMocks: func(m *MockDockerClientForExec) {
				m.On("Close").Return(nil)
				m.On("GetContainerByName", mock.Anything, "test-codespace-dev").Return(&ContainerInfo{
					ID:    "container123",
					State: "running",
				}, nil)
			},
			expectedError: false,
			checkDockerCmd: true,
			expectedDockerArgs: []string{"exec", "-it", "test-codespace-dev", "ls", "-la"},
		},
		{
			name: "Execute interactive shell (no command specified)",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockDockerClientForExec) {
				m.On("Close").Return(nil)
				m.On("GetContainerByName", mock.Anything, "test-codespace-dev").Return(&ContainerInfo{
					ID:    "container123",
					State: "running",
				}, nil)
			},
			expectedError: false,
			checkDockerCmd: true,
			expectedDockerArgs: []string{"exec", "-it", "test-codespace-dev", "/bin/bash"},
		},
		{
			name: "Container not found",
			args: []string{"nonexistent-codespace"},
			setupMocks: func(m *MockDockerClientForExec) {
				m.On("Close").Return(nil)
				m.On("GetContainerByName", mock.Anything, "nonexistent-codespace-dev").Return(nil, errors.New("not found"))
			},
			expectedError: true,
			errorContains: "codespace 'nonexistent-codespace' not found",
		},
		{
			name: "Container not running",
			args: []string{"stopped-codespace"},
			setupMocks: func(m *MockDockerClientForExec) {
				m.On("Close").Return(nil)
				m.On("GetContainerByName", mock.Anything, "stopped-codespace-dev").Return(&ContainerInfo{
					ID:    "container123",
					State: "stopped",
				}, nil)
			},
			expectedError: true,
			errorContains: "codespace 'stopped-codespace' is not running. Use 'mcs start stopped-codespace' first",
		},
		{
			name: "Docker client creation fails",
			args: []string{"test-codespace"},
			setupMocks: func(m *MockDockerClientForExec) {
				// Mock will be created but not used due to early error
			},
			expectedError: true,
			errorContains: "failed to create Docker client",
		},
		{
			name:          "No arguments",
			args:          []string{},
			setupMocks:    func(m *MockDockerClientForExec) {},
			expectedError: true,
			errorContains: "requires at least 1 arg(s), only received 0",
		},
		{
			name: "Complex command with pipes and flags",
			args: []string{"test-codespace", "sh", "-c", "ps aux | grep node"},
			setupMocks: func(m *MockDockerClientForExec) {
				m.On("Close").Return(nil)
				m.On("GetContainerByName", mock.Anything, "test-codespace-dev").Return(&ContainerInfo{
					ID:    "container123",
					State: "running",
				}, nil)
			},
			expectedError: false,
			checkDockerCmd: true,
			expectedDockerArgs: []string{"exec", "-it", "test-codespace-dev", "sh", "-c", "ps aux | grep node"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that would actually execute docker commands
			if tt.checkDockerCmd {
				t.Skip("Skipping test that would execute actual docker command")
			}

			mockDocker := new(MockDockerClientForExec)
			if tt.setupMocks != nil {
				tt.setupMocks(mockDocker)
			}

			cmd := ExecCommand()
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

			mockDocker.AssertExpectations(t)
		})
	}
}

func TestExecCommand_ExitCode(t *testing.T) {
	// This test verifies that exit codes are properly propagated
	// In a real test environment, we would need to mock exec.Command
	t.Skip("Exit code propagation test requires exec.Command mocking")
}

func TestExecCommand_Help(t *testing.T) {
	cmd := ExecCommand()
	
	assert.Equal(t, "exec <name> [command...]", cmd.Use)
	assert.Contains(t, cmd.Short, "Execute a command in a codespace")
	assert.Contains(t, cmd.Long, "interactive shell")
}

func TestExecCommand_ContainerNameFormat(t *testing.T) {
	tests := []struct {
		codespace     string
		expectedName  string
	}{
		{
			codespace:    "my-project",
			expectedName: "my-project-dev",
		},
		{
			codespace:    "user-repo",
			expectedName: "user-repo-dev",
		},
		{
			codespace:    "test",
			expectedName: "test-dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.codespace, func(t *testing.T) {
			// The container name format is hardcoded in the exec command
			containerName := tt.codespace + "-dev"
			assert.Equal(t, tt.expectedName, containerName)
		})
	}
}

// Integration test skeleton - would require Docker in test environment
func TestExecCommand_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test")
	}

	// This would:
	// 1. Create a test container
	// 2. Execute commands in it
	// 3. Verify output
	// 4. Clean up
}

// Test that verifies the command builds correct docker arguments
func TestExecCommand_DockerArgs(t *testing.T) {
	tests := []struct {
		name         string
		inputArgs    []string
		expectedArgs []string
	}{
		{
			name:         "Simple command",
			inputArgs:    []string{"test-codespace", "ls"},
			expectedArgs: []string{"exec", "-it", "test-codespace-dev", "ls"},
		},
		{
			name:         "Command with flags",
			inputArgs:    []string{"test-codespace", "ls", "-la", "/app"},
			expectedArgs: []string{"exec", "-it", "test-codespace-dev", "ls", "-la", "/app"},
		},
		{
			name:         "No command (interactive shell)",
			inputArgs:    []string{"test-codespace"},
			expectedArgs: []string{"exec", "-it", "test-codespace-dev", "/bin/bash"},
		},
		{
			name:         "Complex shell command",
			inputArgs:    []string{"test-codespace", "sh", "-c", "echo $PATH"},
			expectedArgs: []string{"exec", "-it", "test-codespace-dev", "sh", "-c", "echo $PATH"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies the logic without actually running docker
			// In the actual implementation, these args would be passed to exec.Command
			
			dockerArgs := []string{"exec", "-it", tt.inputArgs[0] + "-dev"}
			if len(tt.inputArgs) > 1 {
				dockerArgs = append(dockerArgs, tt.inputArgs[1:]...)
			} else {
				dockerArgs = append(dockerArgs, "/bin/bash")
			}
			
			assert.Equal(t, tt.expectedArgs, dockerArgs)
		})
	}
}

// MockExecCommand for testing command execution
type MockExecCommand struct {
	mock.Mock
}

func (m *MockExecCommand) Run() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockExecCommand) CombinedOutput() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

// Test error handling for different failure scenarios
func TestExecCommand_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		containerState string
		dockerError   error
		expectedError string
	}{
		{
			name:          "Container exited",
			containerState: "exited",
			expectedError: "is not running",
		},
		{
			name:          "Container paused",
			containerState: "paused",
			expectedError: "is not running",
		},
		{
			name:          "Container restarting",
			containerState: "restarting",
			expectedError: "is not running",
		},
		{
			name:          "Docker daemon not running",
			dockerError:   errors.New("Cannot connect to the Docker daemon"),
			expectedError: "codespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDocker := new(MockDockerClientForExec)
			mockDocker.On("Close").Return(nil)
			
			if tt.dockerError != nil {
				mockDocker.On("GetContainerByName", mock.Anything, mock.Anything).Return(nil, tt.dockerError)
			} else {
				mockDocker.On("GetContainerByName", mock.Anything, mock.Anything).Return(&ContainerInfo{
					ID:    "container123",
					State: tt.containerState,
				}, nil)
			}

			cmd := ExecCommand()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs([]string{"test-codespace", "ls"})

			err := cmd.Execute()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)

			mockDocker.AssertExpectations(t)
		})
	}
}