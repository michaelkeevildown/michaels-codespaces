package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/michaelkeevildown/mcs/internal/shell"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
			// Show beautiful header
			ui.ShowHeader()
			
			if !force {
				fmt.Println("⚠️  This will remove MCS installation files.")
				fmt.Println("   Your codespaces will be preserved and containers will keep running.")
				fmt.Println()
				fmt.Print("Are you sure? [y/N] ")
				os.Stdout.Sync() // Ensure prompt is displayed

				// Read user choice with terminal-aware logic
				var reader *bufio.Reader
				var response string
				
				// Check if stdin is a terminal
				if !term.IsTerminal(int(syscall.Stdin)) {
					// stdin is not a terminal (e.g., piped input)
					// Try to open /dev/tty directly to read from the actual terminal
					tty, err := os.Open("/dev/tty")
					if err != nil {
						// Can't get user input in non-interactive mode
						// Default to NO for safety
						fmt.Println("n (non-interactive mode)")
						fmt.Println("Cancelled.")
						return nil
					}
					defer tty.Close()
					reader = bufio.NewReader(tty)
				} else {
					// Normal interactive mode
					reader = bufio.NewReader(os.Stdin)
				}
				
				response, _ = reader.ReadString('\n')
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

	patterns := []string{
		"Michael's Codespaces",
		"MCS aliases",
		"/.mcs/bin",
		"# Codespace:",
		"mcs completion",
	}
	
	for _, configFile := range shellConfigs {
		if err := shell.CleanConfig(configFile, patterns); err != nil {
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

