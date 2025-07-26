package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"github.com/michaelkeevildown/mcs/internal/docker"
	"github.com/spf13/cobra"
)

// LogsCommand creates the 'logs' command
func LogsCommand() *cobra.Command {
	var follow bool
	var tail string
	
	cmd := &cobra.Command{
		Use:   "logs <name>",
		Short: "📜 View codespace logs",
		Long:  "Display logs from a codespace container.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			name := args[0]
			
			// Create Docker client to check container
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
			
			// Build docker logs command
			dockerArgs := []string{"logs"}
			
			if follow {
				dockerArgs = append(dockerArgs, "-f")
			}
			
			if tail != "" {
				dockerArgs = append(dockerArgs, "--tail", tail)
			}
			
			dockerArgs = append(dockerArgs, containerName)
			
			// Execute docker command
			dockerCmd := exec.Command("docker", dockerArgs...)
			dockerCmd.Stdout = os.Stdout
			dockerCmd.Stderr = os.Stderr
			
			// Handle Ctrl+C gracefully when following logs
			if follow {
				// Set up signal handling
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt)
				
				go func() {
					<-sigChan
					// Kill the docker process
					if dockerCmd.Process != nil {
						dockerCmd.Process.Kill()
					}
				}()
			}
			
			// Run command
			if err := dockerCmd.Run(); err != nil {
				// Ignore error if it was interrupted
				if follow {
					return nil
				}
				return fmt.Errorf("failed to get logs: %w", err)
			}
			
			return nil
		},
	}
	
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	cmd.Flags().StringVar(&tail, "tail", "", "Number of lines to show from the end of the logs")
	
	return cmd
}