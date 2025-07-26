package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/michaelkeevildown/mcs/internal/codespace"
	"github.com/michaelkeevildown/mcs/internal/docker"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/spf13/cobra"
)

// ResetPasswordCommand creates the 'reset-password' command
func ResetPasswordCommand() *cobra.Command {
	var force bool
	
	cmd := &cobra.Command{
		Use:   "reset-password <name>",
		Short: "🔐 Reset codespace password",
		Long:  "Generate a new password for VS Code access to a codespace.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			name := args[0]
			
			// Create manager
			manager := codespace.NewManager()
			
			// Get codespace
			cs, err := manager.Get(ctx, name)
			if err != nil {
				return fmt.Errorf("codespace '%s' not found", name)
			}
			
			// Confirm if not forced
			if !force {
				fmt.Printf("⚠️  This will change the password for codespace '%s'.\n", name)
				fmt.Print("Are you sure? [y/N] ")
				os.Stdout.Sync() // Ensure prompt is displayed
				
				var response string
				fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))
				
				// Show what was selected to fix cursor positioning
				if response == "" {
					fmt.Println("n") // Default is NO
				}
				
				if response != "y" && response != "yes" {
					fmt.Println("Cancelled.")
					return nil
				}
			}
			
			// Create progress indicator
			progress := ui.NewProgress()
			progress.Start("Resetting password")
			
			// Generate new password
			newPassword := generateSecurePassword()
			
			// Update .env file
			envPath := filepath.Join(cs.Path, ".env")
			if err := updateEnvFile(envPath, "PASSWORD", newPassword); err != nil {
				progress.Fail("Failed to update .env file")
				return err
			}
			
			// Update docker-compose.yml
			composePath := filepath.Join(cs.Path, "docker-compose.yml")
			if err := updateComposeFile(composePath, newPassword); err != nil {
				progress.Fail("Failed to update docker-compose.yml")
				return err
			}
			
			// Update metadata
			cs.Password = newPassword
			if err := manager.SaveMetadata(cs); err != nil {
				progress.Fail("Failed to update metadata")
				return err
			}
			
			// Check if container is running
			dockerClient, err := docker.NewClient()
			if err == nil {
				defer dockerClient.Close()
				
				containerName := fmt.Sprintf("%s-dev", name)
				container, err := dockerClient.GetContainerByName(ctx, containerName)
				if err == nil && container.State == "running" {
					progress.Update("Restarting container")
					
					// Stop container
					if err := dockerClient.StopContainer(ctx, container.ID); err != nil {
						progress.Fail("Failed to stop container")
						return err
					}
					
					// Start container with new password
					if err := manager.Start(ctx, name); err != nil {
						progress.Fail("Failed to restart container")
						return err
					}
				}
			}
			
			progress.Success("Password reset successfully!")
			
			// Show new credentials
			fmt.Println()
			fmt.Println(recoverHeaderStyle.Render("New Credentials"))
			fmt.Println(strings.Repeat("━", 50))
			fmt.Printf("URL: %s\n", recoverURLStyle.Render(cs.VSCodeURL))
			fmt.Printf("Password: %s\n", recoverURLStyle.Render(newPassword))
			fmt.Println(strings.Repeat("━", 50))
			fmt.Println()
			
			return nil
		},
	}
	
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force password reset without confirmation")
	
	return cmd
}

// generateSecurePassword generates a secure random password
func generateSecurePassword() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:16]
}

// updateEnvFile updates a key in the .env file
func updateEnvFile(path, key, value string) error {
	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}
	
	// Update the key
	lines := strings.Split(string(content), "\n")
	updated := false
	for i, line := range lines {
		if strings.HasPrefix(line, key+"=") {
			lines[i] = fmt.Sprintf("%s=%s", key, value)
			updated = true
		}
	}
	
	// Add key if not found
	if !updated {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}
	
	// Write back
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

// updateComposeFile updates the password in docker-compose.yml
func updateComposeFile(path, newPassword string) error {
	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read docker-compose.yml: %w", err)
	}
	
	// Replace PASSWORD environment variable
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "- PASSWORD=") {
			// Preserve indentation
			indent := strings.TrimRight(line, strings.TrimSpace(line))
			lines[i] = fmt.Sprintf("%s- PASSWORD=%s", indent, newPassword)
		}
	}
	
	// Write back
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}