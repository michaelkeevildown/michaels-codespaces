package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelkeevildown/mcs/internal/backup"
	"github.com/michaelkeevildown/mcs/internal/docker"
	"github.com/michaelkeevildown/mcs/internal/shell"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/michaelkeevildown/mcs/pkg/utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	// Destroy-specific styles
	destroyWarningStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	destroyInfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// DestroyCommand creates the 'destroy' command
func DestroyCommand() *cobra.Command {
	var (
		force      bool
		keepDocker bool
		skipBackup bool
	)

	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "💣 Completely remove MCS and all codespaces",
		Long: `Completely remove MCS, all codespaces, and optionally Docker.
		
⚠️  WARNING: This is a destructive operation!

This will:
- Stop and remove ALL codespaces and containers
- Delete ALL codespace data in ~/codespaces
- Remove MCS installation completely
- Optionally uninstall Docker
- Remove all shell configurations

Use 'mcs cleanup' for a soft removal that preserves codespaces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Println(destroyWarningStyle.Render("⚠️  DESTRUCTIVE OPERATION WARNING ⚠️"))
				fmt.Println()
				fmt.Println("This will permanently delete:")
				fmt.Println("  • All codespaces and their data")
				fmt.Println("  • All running containers")
				fmt.Println("  • MCS installation")
				if !keepDocker {
					fmt.Println("  • Docker installation")
				}
				fmt.Println()
				fmt.Println(destroyWarningStyle.Render("This action cannot be undone!"))
				fmt.Println()
				fmt.Print("Type 'destroy' to confirm: ")
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
						// Default to NO/cancel for safety
						fmt.Println("(non-interactive mode)")
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
				response = strings.TrimSpace(response)

				// Show what was typed to fix cursor positioning
				if response == "" {
					fmt.Println() // Just move to next line if nothing typed
				}

				if response != "destroy" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			return performDestroy(keepDocker, skipBackup)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&keepDocker, "keep-docker", false, "Keep Docker installed")
	cmd.Flags().BoolVar(&skipBackup, "skip-backup", false, "Skip backup of codespaces")

	return cmd
}

func performDestroy(keepDocker, skipBackup bool) error {
	progress := ui.NewProgress()
	
	// Header
	fmt.Println()
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println(headerStyle.Render("MCS Destruction Process"))
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println()
	
	progress.Start("Beginning destruction sequence")

	// 1. Create backup if requested
	if !skipBackup {
		fmt.Println()
		fmt.Println(infoStyle.Render("📦 Creating Backup"))
		fmt.Println(strings.Repeat("─", 50))
		fmt.Println(dimStyle.Render("Backing up your codespace data..."))
		fmt.Println()
		
		// Create backup using BackupManager
		backupManager := backup.NewBackupManager()
		homeDir := os.Getenv("HOME")
		codespacesDir := filepath.Join(homeDir, "codespaces")
		
		backupID, err := backupManager.CreateQuick(codespacesDir, backup.BackupTypeDestroy)
		if err != nil {
			// Don't fail, just warn
			fmt.Println(warningStyle.Render("⚠️  Failed to create backup: " + err.Error()))
			fmt.Println(dimStyle.Render("   Continuing without backup..."))
		} else {
			progress.Update("Created backup")
			fmt.Println(successStyle.Render("✓ Backup created: " + backupID))
			fmt.Println(dimStyle.Render("  Your data is safe in ~/.mcs.backup and can be restored later"))
		}
	}

	// 2. Stop all running containers
	dockerClient, err := docker.NewClient()
	if err == nil {
		fmt.Println()
		fmt.Println(infoStyle.Render("🐳 Container Cleanup"))
		fmt.Println(strings.Repeat("─", 50))
		
		progress.Update("Checking for running containers")

		containers, err := dockerClient.ListContainers(context.Background(), "")
		if err == nil {
			mcsContainerCount := 0
			for _, container := range containers {
				if strings.Contains(container.Name, "mcs-") ||
					strings.Contains(container.Name, "michaelkeevildown/claude-coder") {
					mcsContainerCount++
				}
			}
			
			if mcsContainerCount > 0 {
				fmt.Printf(dimStyle.Render("Found %d MCS container(s) to remove\n"), mcsContainerCount)
				fmt.Println()
				
				progress.Update("Stopping containers")
				for _, container := range containers {
					// Only stop MCS containers
					if strings.Contains(container.Name, "mcs-") ||
						strings.Contains(container.Name, "michaelkeevildown/claude-coder") {
						shortID := container.ID[:12]
						fmt.Printf(dimStyle.Render("  → Stopping %s (%s)..."), container.Name, shortID)
						if err := dockerClient.StopContainer(context.Background(), container.ID); err != nil {
							fmt.Printf(" %s\n", errorStyle.Render("failed"))
							fmt.Printf(dimStyle.Render("    Error: %v\n"), err)
						} else {
							fmt.Printf(" %s\n", successStyle.Render("✓"))
						}
					}
				}
			} else {
				fmt.Println(dimStyle.Render("No active containers found"))
			}
		}
		progress.Update("All containers stopped")

		// 3. Remove all containers
		fmt.Println()
		progress.Update("Removing containers")
		if err := dockerClient.RemoveAllMCSContainers(context.Background()); err != nil {
			fmt.Println(warningStyle.Render("⚠️  Some containers could not be removed"))
			fmt.Println(dimStyle.Render("   You may need to remove them manually"))
		} else {
			fmt.Println(successStyle.Render("✓ All containers removed"))
		}
	}

	// 4. Remove codespaces directory
	homeDir := os.Getenv("HOME")
	codespacesDir := filepath.Join(homeDir, "codespaces")
	
	fmt.Println()
	fmt.Println(infoStyle.Render("📁 Removing Data"))
	fmt.Println(strings.Repeat("─", 50))
	
	if _, err := os.Stat(codespacesDir); err == nil {
		// Try to get directory size
		dirSize, _ := utils.CalculateDirSize(codespacesDir)
		sizeStr := utils.FormatBytes(dirSize)
		fmt.Printf(dimStyle.Render("Removing codespace data (%s)...\n"), sizeStr)
		
		progress.Update("Removing codespaces directory")
		if err := os.RemoveAll(codespacesDir); err != nil {
			progress.Fail("Failed to remove codespaces directory")
			return fmt.Errorf("failed to remove %s: %w", codespacesDir, err)
		}
		fmt.Println(successStyle.Render("✓ Removed ~/codespaces"))
	} else {
		fmt.Println(dimStyle.Render("No codespaces directory found"))
	}

	// 5. Remove MCS installation
	fmt.Println()
	fmt.Println(dimStyle.Render("Removing MCS installation files..."))
	
	mcsHome := os.Getenv("MCS_HOME")
	if mcsHome == "" {
		mcsHome = filepath.Join(homeDir, ".mcs")
	}

	// Remove MCS home directory
	if err := utils.RemoveDirectoryWithSize(mcsHome, progress); err != nil {
		fmt.Println(warningStyle.Render("⚠️  Failed to remove MCS directory"))
	}

	// Remove MCS binary from PATH locations
	fmt.Println()
	fmt.Println(dimStyle.Render("Removing MCS binaries..."))
	pathLocations := []string{
		"/usr/local/bin/mcs",
		filepath.Join(homeDir, ".local/bin/mcs"),
		filepath.Join(homeDir, "bin/mcs"),
	}

	utils.RemovePathsWithProgress(pathLocations, "→ Removing")

	// 6. Clean shell configurations
	fmt.Println()
	fmt.Println(infoStyle.Render("🐚 Shell Configuration"))
	fmt.Println(strings.Repeat("─", 50))
	
	progress.Update("Cleaning shell configurations")
	shellConfigs := []string{
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".bash_profile"),
		filepath.Join(homeDir, ".profile"),
	}

	fmt.Println(dimStyle.Render("Cleaning shell configuration files..."))
	cleanedCount := 0
	patterns := []string{
		"Michael's Codespaces",
		"MCS aliases",
		"/.mcs/bin",
		"# Codespace:",
		"mcs completion",
		"export MCS_",
	}
	
	for _, configFile := range shellConfigs {
		if _, err := os.Stat(configFile); err == nil {
			fmt.Printf(dimStyle.Render("  → Cleaning %s..."), filepath.Base(configFile))
			if err := shell.CleanConfig(configFile, patterns); err != nil {
				fmt.Printf(" %s\n", errorStyle.Render("failed"))
			} else {
				fmt.Printf(" %s\n", successStyle.Render("✓"))
				cleanedCount++
			}
		}
	}
	
	if cleanedCount == 0 {
		fmt.Println(dimStyle.Render("No shell configurations found"))
	}

	// 7. Remove helper scripts
	fmt.Println()
	fmt.Println(dimStyle.Render("Removing helper scripts..."))
	scripts := []string{
		filepath.Join(homeDir, "monitor-system.sh"),
		filepath.Join(homeDir, ".mcs-install.sh"),
	}

	for _, script := range scripts {
		if _, err := os.Stat(script); err == nil {
			fmt.Printf(dimStyle.Render("  → Removing %s..."), filepath.Base(script))
			if err := os.Remove(script); err != nil {
				fmt.Printf(" %s\n", errorStyle.Render("failed"))
			} else {
				fmt.Printf(" %s\n", successStyle.Render("✓"))
			}
		}
	}

	// 8. Optionally uninstall Docker
	if !keepDocker {
		fmt.Println()
		fmt.Println(infoStyle.Render("🐳 Docker Uninstallation"))
		fmt.Println(strings.Repeat("─", 50))
		
		progress.Update("Preparing to uninstall Docker")
		fmt.Println(dimStyle.Render("Removing Docker installation..."))
		fmt.Println()
		
		if err := docker.UninstallDocker(progress); err != nil {
			fmt.Println(warningStyle.Render("⚠️  Failed to uninstall Docker completely"))
			fmt.Println(dimStyle.Render("   You may need to uninstall Docker manually"))
			fmt.Printf(dimStyle.Render("   Error: %v\n"), err)
		} else {
			fmt.Println(successStyle.Render("✓ Docker uninstalled successfully"))
		}
	}

	progress.Success("Destruction complete!")

	// Final summary
	fmt.Println()
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println(destroyWarningStyle.Render("💥 MCS Destruction Complete"))
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println()
	
	fmt.Println(boldStyle.Render("Successfully removed:"))
	fmt.Println()
	fmt.Println("  " + successStyle.Render("✓") + " All codespaces and containers")
	fmt.Println("  " + successStyle.Render("✓") + " All codespace data")
	fmt.Println("  " + successStyle.Render("✓") + " MCS installation files")
	fmt.Println("  " + successStyle.Render("✓") + " Shell configurations")
	fmt.Println("  " + successStyle.Render("✓") + " Helper scripts")
	if !keepDocker {
		fmt.Println("  " + successStyle.Render("✓") + " Docker installation")
	}
	fmt.Println()

	if !skipBackup {
		fmt.Println(strings.Repeat("─", 50))
		fmt.Println(infoStyle.Render("💾 Backup Information"))
		fmt.Println()
		fmt.Println("Your codespace data was backed up to:")
		fmt.Println(boldStyle.Render("  ~/.mcs.backup"))
		fmt.Println()
		fmt.Println(dimStyle.Render("You can safely delete this backup when no longer needed."))
		fmt.Println()
	}

	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()
	fmt.Println(headerStyle.Render("Thank you for using MCS! 👋"))
	fmt.Println()
	fmt.Println(dimStyle.Render("If you change your mind, you can reinstall with:"))
	fmt.Println(dimStyle.Render("  curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/main/mcs-go/install.sh | bash"))
	fmt.Println()

	return nil
}


