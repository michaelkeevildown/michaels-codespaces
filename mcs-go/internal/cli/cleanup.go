package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/spf13/cobra"
)

// CleanupCommand creates the 'cleanup' command
func CleanupCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "cleanup",
		Aliases: []string{"clean"},
		Short:   "🧹 Remove MCS installation (keep codespaces)",
		Long: `Perform a soft cleanup of MCS installation.
		
This will:
- Remove MCS binary and installation files
- Remove shell aliases from .zshrc/.bashrc
- Remove PATH entries
- Keep all codespaces and containers running`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Println("⚠️  This will remove MCS installation files.")
				fmt.Println("   Your codespaces will be preserved and containers will keep running.")
				fmt.Println()
				fmt.Print("Are you sure? [y/N] ")
				os.Stdout.Sync() // Ensure prompt is displayed

				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				
				// Show what was selected to fix cursor positioning
				if response == "" {
					fmt.Println("n") // Default is NO
				}
				
				if response != "y" && response != "yes" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			return performCleanup()
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	
	return cmd
}

func performCleanup() error {
	progress := ui.NewProgress()
	progress.Start("Cleaning up MCS installation")

	mcsHome := os.Getenv("MCS_HOME")
	if mcsHome == "" {
		mcsHome = filepath.Join(os.Getenv("HOME"), ".mcs")
	}

	// Remove MCS installation directory
	if err := os.RemoveAll(mcsHome); err != nil {
		progress.Fail("Failed to remove MCS directory")
		return fmt.Errorf("failed to remove %s: %w", mcsHome, err)
	}
	progress.Update("Removed MCS installation directory")

	// Clean shell configurations
	homeDir := os.Getenv("HOME")
	shellConfigs := []string{
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".bash_profile"),
	}

	for _, configFile := range shellConfigs {
		if err := cleanShellConfig(configFile); err != nil {
			// Don't fail on shell config errors, just warn
			fmt.Printf("Warning: Failed to clean %s: %v\n", configFile, err)
		}
	}
	progress.Update("Cleaned shell configurations")

	// Remove monitor script if it exists
	monitorScript := filepath.Join(homeDir, "monitor-system.sh")
	if _, err := os.Stat(monitorScript); err == nil {
		os.Remove(monitorScript)
	}

	progress.Success("MCS installation removed!")
	
	fmt.Println()
	fmt.Println("✓ MCS installation files removed")
	fmt.Println("✓ Your codespaces in ~/codespaces are preserved")
	fmt.Println("✓ All running containers are still active")
	fmt.Println("✓ Docker remains installed")
	fmt.Println()
	fmt.Println("To reinstall MCS, run the installer again:")
	fmt.Println("  curl -fsSL https://raw.githubusercontent.com/yourusername/mcs/main/mcs-go/install.sh | bash")
	fmt.Println()
	fmt.Println("To completely remove everything including codespaces, use: mcs destroy")

	return nil
}

func cleanShellConfig(configFile string) error {
	// Read the file
	content, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, nothing to clean
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	cleanedLines := []string{}
	skipNext := false

	for i, line := range lines {
		// Skip MCS-related lines
		if strings.Contains(line, "Michael's Codespaces") ||
			strings.Contains(line, "MCS aliases") ||
			strings.Contains(line, "/.mcs/bin") ||
			strings.Contains(line, "# Codespace:") ||
			strings.Contains(line, "mcs completion") {
			// Also skip the next line if it's an alias
			if i+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i+1]), "alias ") {
				skipNext = true
			}
			continue
		}

		if skipNext {
			skipNext = false
			continue
		}

		cleanedLines = append(cleanedLines, line)
	}

	// Write back the cleaned content
	cleanedContent := strings.Join(cleanedLines, "\n")
	return os.WriteFile(configFile, []byte(cleanedContent), 0644)
}