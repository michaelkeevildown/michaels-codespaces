package utils

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Command execution modes
type OutputMode int

const (
	OutputModeNormal OutputMode = iota
	OutputModeVerbose
	OutputModeQuiet
)

var (
	// Styles for command output
	commandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

// CommandRunner provides transparent command execution with progress
type CommandRunner struct {
	Mode           OutputMode
	ShowCommand    bool
	TimeoutWarning time.Duration // Show warning after this duration
}

// NewCommandRunner creates a new command runner with default settings
func NewCommandRunner() *CommandRunner {
	return &CommandRunner{
		Mode:           OutputModeNormal,
		ShowCommand:    true,
		TimeoutWarning: 30 * time.Second,
	}
}

// RunCommand executes a command with transparent output
func (cr *CommandRunner) RunCommand(name string, args ...string) error {
	return cr.RunCommandContext(context.Background(), name, args...)
}

// RunCommandContext executes a command with context and transparent output
func (cr *CommandRunner) RunCommandContext(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	
	// Show what command we're running (unless in quiet mode)
	if cr.ShowCommand && cr.Mode != OutputModeQuiet {
		cmdStr := name + " " + strings.Join(args, " ")
		fmt.Println(commandStyle.Render("→ Running: " + cmdStr))
		if cr.Mode == OutputModeVerbose {
			fmt.Println(dimStyle.Render("  Working directory: " + cmd.Dir))
		}
	}

	// In quiet mode, just run the command
	if cr.Mode == OutputModeQuiet {
		return cmd.Run()
	}

	// For normal and verbose modes, show output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Set up timeout warning
	warningTimer := time.NewTimer(cr.TimeoutWarning)
	defer warningTimer.Stop()

	// Channel to signal command completion
	done := make(chan error, 1)

	// Read output in a separate goroutine
	go func() {
		// Create multi-reader for both stdout and stderr
		reader := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(reader)
		
		for scanner.Scan() {
			line := scanner.Text()
			
			// In verbose mode, show all output
			if cr.Mode == OutputModeVerbose {
				fmt.Println(dimStyle.Render("  │ " + line))
			} else {
				// In normal mode, filter and format output
				formattedLine := cr.formatOutputLine(line)
				if formattedLine != "" {
					fmt.Println(formattedLine)
				}
			}
		}
	}()

	// Wait for command or timeout warning
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-warningTimer.C:
		// Show timeout warning but continue waiting
		fmt.Println()
		fmt.Println(warningStyle.Render("  ⏱  This is taking longer than expected..."))
		fmt.Println(dimStyle.Render("  • Check your internet connection"))
		fmt.Println(dimStyle.Render("  • The remote server might be slow"))
		fmt.Println(dimStyle.Render("  • Press Ctrl+C to cancel if stuck"))
		fmt.Println()
		
		// Wait for actual completion
		return <-done
	case <-ctx.Done():
		// Context cancelled
		cmd.Process.Kill()
		return ctx.Err()
	}
}

// formatOutputLine formats output lines for normal mode
func (cr *CommandRunner) formatOutputLine(line string) string {
	// Skip empty lines
	if strings.TrimSpace(line) == "" {
		return ""
	}

	// APT output formatting
	if strings.Contains(line, "Get:") || strings.Contains(line, "Hit:") {
		// Repository access
		return dimStyle.Render("  ↓ " + line)
	}
	if strings.Contains(line, "Reading package lists") {
		return infoStyle.Render("  📦 Reading package lists...")
	}
	if strings.Contains(line, "Building dependency tree") {
		return infoStyle.Render("  🌳 Building dependency tree...")
	}
	if strings.Contains(line, "packages can be upgraded") {
		return successStyle.Render("  ✓ " + line)
	}
	if strings.Contains(line, "Unpacking") || strings.Contains(line, "Setting up") {
		// Package installation progress
		parts := strings.Fields(line)
		if len(parts) > 1 {
			return dimStyle.Render("  📦 " + parts[0] + " " + parts[1] + "...")
		}
	}

	// Git output formatting
	if strings.Contains(line, "Cloning into") {
		return infoStyle.Render("  📂 " + line)
	}
	if strings.Contains(line, "remote:") {
		// Git remote output
		return dimStyle.Render("  ↓ " + strings.TrimPrefix(line, "remote: "))
	}
	if strings.Contains(line, "Receiving objects:") || strings.Contains(line, "Resolving deltas:") {
		// Git progress
		return dimStyle.Render("  ⟳ " + line)
	}

	// Docker output formatting
	if strings.Contains(line, "Pulling from") {
		return infoStyle.Render("  🐳 " + line)
	}
	if strings.Contains(line, "Pull complete") {
		return successStyle.Render("  ✓ " + line)
	}
	if strings.Contains(line, "Downloaded newer image") {
		return successStyle.Render("  ✓ " + line)
	}

	// Error detection
	if strings.Contains(strings.ToLower(line), "error") || 
	   strings.Contains(strings.ToLower(line), "failed") ||
	   strings.Contains(strings.ToLower(line), "unable to") {
		return errorStyle.Render("  ✗ " + line)
	}

	// Warning detection
	if strings.Contains(strings.ToLower(line), "warning") ||
	   strings.Contains(strings.ToLower(line), "warn") {
		return warningStyle.Render("  ⚠ " + line)
	}

	// Default: show important lines in normal mode
	if strings.Contains(line, "Installing") || 
	   strings.Contains(line, "Downloading") ||
	   strings.Contains(line, "Complete") ||
	   strings.Contains(line, "Done") {
		return dimStyle.Render("  • " + line)
	}

	// Skip other lines in normal mode
	return ""
}

// RunWithProgress runs a command and shows a custom progress message
func (cr *CommandRunner) RunWithProgress(message string, name string, args ...string) error {
	if cr.Mode != OutputModeQuiet {
		fmt.Println(infoStyle.Render("⚡ " + message))
	}
	return cr.RunCommand(name, args...)
}

// RunSilent runs a command silently (for backward compatibility)
func RunCommandSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// CommandExists checks if a command is available in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}