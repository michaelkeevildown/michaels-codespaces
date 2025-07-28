package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/michaelkeevildown/mcs/internal/docker"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/spf13/cobra"
)

// ExecCommand creates the 'exec' command
func ExecCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "exec <name> [command...]",
		Short: "🖥️  Execute a command in a codespace",
		Long: `Execute a command inside a running codespace container.
If no command is specified, an interactive shell will be started.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.ShowHeader()
			ctx := context.Background()
			name := args[0]
			
			// Create Docker client to check container status
			dockerClient, err := docker.NewClient()
			if err != nil {
				return fmt.Errorf("failed to create Docker client: %w", err)
			}
			defer dockerClient.Close()
			
			// Get container
			containerName := fmt.Sprintf("%s-dev", name)
			container, err := dockerClient.GetContainerByName(ctx, containerName)
			if err != nil {
				return fmt.Errorf("codespace '%s' not found", name)
			}
			
			// Check if container is running
			if container.State != "running" {
				return fmt.Errorf("codespace '%s' is not running. Use 'mcs start %s' first", name, name)
			}
			
			// Build docker exec command
			dockerArgs := []string{"exec", "-it", containerName}
			
			// Add command or default to bash
			if len(args) > 1 {
				dockerArgs = append(dockerArgs, args[1:]...)
			} else {
				dockerArgs = append(dockerArgs, "/bin/bash")
			}
			
			// Execute docker command directly
			dockerCmd := exec.Command("docker", dockerArgs...)
			dockerCmd.Stdin = os.Stdin
			dockerCmd.Stdout = os.Stdout
			dockerCmd.Stderr = os.Stderr
			
			// Run and wait
			if err := dockerCmd.Run(); err != nil {
				// Check if it's an exit code from the command
				if exitErr, ok := err.(*exec.ExitError); ok {
					if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
						os.Exit(status.ExitStatus())
					}
				}
				return fmt.Errorf("exec failed: %w", err)
			}
			
			return nil
		},
	}
}