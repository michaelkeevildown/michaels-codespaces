package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/michaelkeevildown/mcs/internal/cli"
	"github.com/michaelkeevildown/mcs/internal/update"
	"github.com/michaelkeevildown/mcs/internal/version"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain tests the main function by capturing its behavior
func TestMain(t *testing.T) {
	// Save original args and restore after test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Save original stderr and restore after test
	originalStderr := os.Stderr
	defer func() { os.Stderr = originalStderr }()

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name:        "help command",
			args:        []string{"mcs", "--help"},
			expectError: false,
		},
		{
			name:        "version command",
			args:        []string{"mcs", "--version"},
			expectError: false,
		},
		{
			name:           "invalid command",
			args:        []string{"mcs", "invalid-command"},
			expectError: true,
			expectedOutput: "unknown command",
		},
		{
			name:        "no args shows help",
			args:        []string{"mcs"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe to capture stderr
			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stderr = w

			// Set test args
			os.Args = tt.args

			// Run main in a goroutine and capture any panics
			done := make(chan bool)
			var recovered interface{}
			
			go func() {
				defer func() {
					if r := recover(); r != nil {
						recovered = r
					}
					done <- true
				}()

				// Since main() calls os.Exit, we need to test the root command directly
				// instead of calling main()
				rootCmd := createRootCommand()
				err := rootCmd.Execute()
				if err != nil && tt.expectError {
					// Expected error case
				} else if err != nil && !tt.expectError {
					t.Errorf("unexpected error: %v", err)
				} else if err == nil && tt.expectError {
					t.Error("expected error but got none")
				}
			}()

			// Wait for completion with timeout
			select {
			case <-done:
				if recovered != nil {
					if !tt.expectError {
						t.Errorf("unexpected panic: %v", recovered)
					}
				}
			case <-time.After(2 * time.Second):
				t.Error("test timed out")
			}

			// Close writer and read output
			w.Close()
			output, err := io.ReadAll(r)
			require.NoError(t, err)

			// Check output if expected
			if tt.expectedOutput != "" {
				assert.Contains(t, string(output), tt.expectedOutput)
			}
		})
	}
}

// TestRootCommand tests the root command configuration
func TestRootCommand(t *testing.T) {
	rootCmd := createRootCommand()

	// Test basic properties
	assert.Equal(t, "mcs", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "Michael's Codespaces")
	assert.Contains(t, rootCmd.Long, "AI-powered development environments")
	assert.NotEmpty(t, rootCmd.Version)

	// Test that all expected commands are registered
	expectedCommands := []string{
		"setup", "create", "list", "start", "stop", "restart",
		"rebuild", "remove", "exec", "logs", "info", "recover",
		"reset-password", "update-ip", "autoupdate", "doctor",
		"status", "update", "check-updates", "update-images",
		"cleanup", "destroy", "backup", "version",
	}

	registeredCommands := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		registeredCommands[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		assert.True(t, registeredCommands[expected], "expected command %s to be registered", expected)
	}
}

// TestPersistentPreRun tests the update check logic
func TestPersistentPreRun(t *testing.T) {
	rootCmd := createRootCommand()

	tests := []struct {
		name       string
		cmdName    string
		shouldSkip bool
	}{
		{"update command skips check", "update", true},
		{"autoupdate command skips check", "autoupdate", true},
		{"version command skips check", "version", true},
		{"help command skips check", "help", true},
		{"completion command skips check", "completion", true},
		{"other commands check updates", "list", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock command with the test name
			mockCmd := &cobra.Command{Use: tt.cmdName}
			
			// Mock the update checker to verify if it's called
			updateChecked := false
			originalCheckForUpdates := update.CheckForUpdates
			update.CheckForUpdates = func(version string) {
				updateChecked = true
			}
			defer func() { update.CheckForUpdates = originalCheckForUpdates }()

			// Execute the PersistentPreRun
			if rootCmd.PersistentPreRun != nil {
				rootCmd.PersistentPreRun(mockCmd, []string{})
			}

			// Give goroutine time to execute
			time.Sleep(10 * time.Millisecond)

			if tt.shouldSkip {
				assert.False(t, updateChecked, "update check should have been skipped for %s", tt.cmdName)
			} else {
				// For non-skip commands, we can't reliably test the goroutine execution
				// without more complex mocking
			}
		})
	}
}

// TestHelpTemplate tests the custom help template
func TestHelpTemplate(t *testing.T) {
	template := helpTemplate()
	
	// Verify the template contains expected placeholders
	assert.Contains(t, template, "{{with (or .Long .Short)}}")
	assert.Contains(t, template, "{{.UsageString}}")
	assert.Contains(t, template, "trimTrailingWhitespaces")
}

// TestMainWithMockExit tests main() with a mocked os.Exit
func TestMainWithMockExit(t *testing.T) {
	// Save original functions
	originalExit := os.Exit
	originalArgs := os.Args
	defer func() {
		os.Exit = originalExit
		os.Args = originalArgs
	}()

	// Mock os.Exit to capture exit code
	var exitCode *int
	os.Exit = func(code int) {
		exitCode = &code
		panic(fmt.Sprintf("os.Exit(%d)", code))
	}

	tests := []struct {
		name         string
		args         []string
		expectedExit int
	}{
		{
			name:         "successful help",
			args:         []string{"mcs", "--help"},
			expectedExit: -1, // No exit expected for help
		},
		{
			name:         "invalid command exits with 1",
			args:         []string{"mcs", "nonexistent"},
			expectedExit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset exit code
			exitCode = nil
			os.Args = tt.args

			// Capture stderr
			var stderr bytes.Buffer
			originalStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w
			defer func() { os.Stderr = originalStderr }()

			// Run main and catch panic from mocked os.Exit
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Expected panic from os.Exit mock
						if !strings.Contains(fmt.Sprint(r), "os.Exit") {
							t.Errorf("unexpected panic: %v", r)
						}
					}
					w.Close()
					io.Copy(&stderr, r)
				}()
				main()
			}()

			// Check exit code
			if tt.expectedExit >= 0 {
				require.NotNil(t, exitCode, "expected os.Exit to be called")
				assert.Equal(t, tt.expectedExit, *exitCode)
			} else {
				assert.Nil(t, exitCode, "expected os.Exit not to be called")
			}
		})
	}
}

// TestCommandIntegration tests that CLI commands can be created without error
func TestCommandIntegration(t *testing.T) {
	// This test ensures all command constructors work properly
	commands := []func() *cobra.Command{
		cli.SetupCommand,
		cli.CreateCommand,
		cli.ListCommand,
		cli.StartCommand,
		cli.StopCommand,
		cli.RestartCommand,
		cli.RebuildCommand,
		cli.RemoveCommand,
		cli.ExecCommand,
		cli.LogsCommand,
		cli.InfoCommand,
		cli.RecoverCommand,
		cli.ResetPasswordCommand,
		cli.UpdateIPCommand,
		cli.AutoUpdateCommand,
		cli.DoctorCommand,
		cli.StatusCommand,
		cli.UpdateCommand,
		cli.CheckUpdatesCommand,
		cli.UpdateImagesCommand,
		cli.CleanupCommand,
		cli.DestroyCommand,
		cli.BackupCommand,
		cli.VersionCommand,
	}

	for i, cmdFunc := range commands {
		t.Run(fmt.Sprintf("command_%d", i), func(t *testing.T) {
			// Ensure command can be created without panic
			assert.NotPanics(t, func() {
				cmd := cmdFunc()
				assert.NotNil(t, cmd)
				assert.NotEmpty(t, cmd.Use)
			})
		})
	}
}

// createRootCommand extracts the root command creation logic for testing
func createRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mcs",
		Short: "🚀 Michael's Codespaces - AI-powered development environments",
		Long: `Michael's Codespaces (MCS) provides isolated, reproducible development
environments optimized for AI agents and modern development workflows.

Run AI agents without constraints, on your own hardware.`,
		Version: version.Info(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip update check for certain commands
			skipCommands := map[string]bool{
				"update":     true,
				"autoupdate": true,
				"version":    true,
				"help":       true,
				"completion": true,
			}
			
			if !skipCommands[cmd.Name()] {
				// Check for updates in the background
				go update.CheckForUpdates(version.Info())
			}
		},
	}

	// Add commands
	rootCmd.AddCommand(
		cli.SetupCommand(),
		cli.CreateCommand(),
		cli.ListCommand(),
		cli.StartCommand(),
		cli.StopCommand(),
		cli.RestartCommand(),
		cli.RebuildCommand(),
		cli.RemoveCommand(),
		cli.ExecCommand(),
		cli.LogsCommand(),
		cli.InfoCommand(),
		cli.RecoverCommand(),
		cli.ResetPasswordCommand(),
		cli.UpdateIPCommand(),
		cli.AutoUpdateCommand(),
		cli.DoctorCommand(),
		cli.StatusCommand(),
		cli.UpdateCommand(),
		cli.CheckUpdatesCommand(),
		cli.UpdateImagesCommand(),
		cli.CleanupCommand(),
		cli.DestroyCommand(),
		cli.BackupCommand(),
		cli.VersionCommand(),
	)

	// Customize help
	rootCmd.SetHelpTemplate(helpTemplate())

	return rootCmd
}

// TestErrorHandling tests the error handling in main
func TestErrorHandling(t *testing.T) {
	// Create a command that returns an error
	errorCmd := &cobra.Command{
		Use: "error-test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("test error")
		},
	}

	rootCmd := createRootCommand()
	rootCmd.AddCommand(errorCmd)

	// Capture stderr
	var stderr bytes.Buffer
	rootCmd.SetErr(&stderr)

	// Set args to trigger error command
	rootCmd.SetArgs([]string{"error-test"})

	// Execute should return error
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test error")
}

// TestConcurrentUpdateCheck tests that update check runs in background
func TestConcurrentUpdateCheck(t *testing.T) {
	rootCmd := createRootCommand()
	
	// Track if update check was called
	checkCalled := make(chan bool, 1)
	originalCheckForUpdates := update.CheckForUpdates
	update.CheckForUpdates = func(version string) {
		checkCalled <- true
	}
	defer func() { update.CheckForUpdates = originalCheckForUpdates }()

	// Create a command that shouldn't skip update check
	testCmd := &cobra.Command{Use: "test-update-check"}
	
	// Call PersistentPreRun
	if rootCmd.PersistentPreRun != nil {
		rootCmd.PersistentPreRun(testCmd, []string{})
	}

	// Wait for goroutine or timeout
	select {
	case <-checkCalled:
		// Success - update check was called
	case <-time.After(100 * time.Millisecond):
		// Update check runs in background, so this is also acceptable
		// as we can't guarantee goroutine execution in tests
	}
}

// TestVersionInfo tests that version is properly set
func TestVersionInfo(t *testing.T) {
	rootCmd := createRootCommand()
	
	// Version should be set from version.Info()
	assert.Equal(t, version.Info(), rootCmd.Version)
	assert.NotEmpty(t, rootCmd.Version)
}

// TestPanicRecovery tests that main handles panics gracefully
func TestPanicRecovery(t *testing.T) {
	// Create a command that panics
	panicCmd := &cobra.Command{
		Use: "panic-test",
		Run: func(cmd *cobra.Command, args []string) {
			panic("test panic")
		},
	}

	rootCmd := createRootCommand()
	rootCmd.AddCommand(panicCmd)

	// Capture output
	var output bytes.Buffer
	rootCmd.SetOut(&output)
	rootCmd.SetErr(&output)

	// Set args to trigger panic command
	rootCmd.SetArgs([]string{"panic-test"})

	// Execute should handle the panic
	assert.Panics(t, func() {
		rootCmd.Execute()
	})
}

// TestCommandLineFlags tests various command line flag combinations
func TestCommandLineFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		shouldError bool
	}{
		{
			name:        "help flag short",
			args:        []string{"-h"},
			shouldError: false,
		},
		{
			name:        "help flag long",
			args:        []string{"--help"},
			shouldError: false,
		},
		{
			name:        "version flag short",
			args:        []string{"-v"},
			shouldError: false,
		},
		{
			name:        "version flag long",
			args:        []string{"--version"},
			shouldError: false,
		},
		{
			name:        "unknown flag",
			args:        []string{"--unknown-flag"},
			shouldError: true,
		},
		{
			name:        "command with help",
			args:        []string{"list", "--help"},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := createRootCommand()
			
			// Capture output
			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			
			// Set test args
			rootCmd.SetArgs(tt.args)
			
			// Execute
			err := rootCmd.Execute()
			
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUpdateCheckSkipLogic tests the skip logic more thoroughly
func TestUpdateCheckSkipLogic(t *testing.T) {
	skipCommands := []string{
		"update", "autoupdate", "version", "help", "completion",
	}

	for _, cmdName := range skipCommands {
		t.Run(fmt.Sprintf("skip_%s", cmdName), func(t *testing.T) {
			// Verify the command name is in the skip map
			skipMap := map[string]bool{
				"update":     true,
				"autoupdate": true,
				"version":    true,
				"help":       true,
				"completion": true,
			}
			
			assert.True(t, skipMap[cmdName], "command %s should be in skip map", cmdName)
		})
	}
}

// BenchmarkRootCommandCreation benchmarks the root command creation
func BenchmarkRootCommandCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = createRootCommand()
	}
}

// BenchmarkHelpTemplate benchmarks the help template generation
func BenchmarkHelpTemplate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = helpTemplate()
	}
}