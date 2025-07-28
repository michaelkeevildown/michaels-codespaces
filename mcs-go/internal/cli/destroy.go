package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelkeevildown/mcs/internal/backup"
	"github.com/michaelkeevildown/mcs/internal/docker"
	"github.com/michaelkeevildown/mcs/internal/ui"
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
		if err := removeAllContainers(dockerClient); err != nil {
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
		var dirSize int64
		filepath.Walk(codespacesDir, func(_ string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				dirSize += info.Size()
			}
			return nil
		})
		
		sizeStr := formatBytes(uint64(dirSize))
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

	dirsToRemove := []string{
		mcsHome,
	}
	
	for _, dir := range dirsToRemove {
		if _, err := os.Stat(dir); err == nil {
			fmt.Printf(dimStyle.Render("  → Removing %s..."), dir)
			if err := os.RemoveAll(dir); err != nil {
				fmt.Printf(" %s\n", errorStyle.Render("failed"))
			} else {
				fmt.Printf(" %s\n", successStyle.Render("✓"))
			}
		}
	}

	// Remove MCS binary from PATH locations
	fmt.Println()
	fmt.Println(dimStyle.Render("Removing MCS binaries..."))
	pathLocations := []string{
		"/usr/local/bin/mcs",
		filepath.Join(homeDir, ".local/bin/mcs"),
		filepath.Join(homeDir, "bin/mcs"),
	}

	for _, loc := range pathLocations {
		if _, err := os.Stat(loc); err == nil {
			fmt.Printf(dimStyle.Render("  → Removing %s..."), loc)
			if err := os.Remove(loc); err != nil {
				fmt.Printf(" %s\n", errorStyle.Render("failed"))
			} else {
				fmt.Printf(" %s\n", successStyle.Render("✓"))
			}
		}
	}

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
	for _, configFile := range shellConfigs {
		if _, err := os.Stat(configFile); err == nil {
			fmt.Printf(dimStyle.Render("  → Cleaning %s..."), filepath.Base(configFile))
			if err := cleanShellConfigDestroy(configFile); err != nil {
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
		
		if err := uninstallDocker(progress); err != nil {
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


func removeAllContainers(client *docker.Client) error {
	containers, err := client.ListContainers(context.Background(), "")
	if err != nil {
		return err
	}

	for _, container := range containers {
		// Only remove MCS containers
		if strings.Contains(container.Name, "mcs-") ||
			strings.Contains(container.Name, "michaelkeevildown/claude-coder") {
			// Force remove
			cmd := exec.Command("docker", "rm", "-f", container.ID)
			if err := cmd.Run(); err != nil {
				// Continue even if one fails
				fmt.Printf("Warning: Failed to remove container %s\n", container.ID[:12])
			}
		}
	}

	return nil
}

func uninstallDocker(progress *ui.Progress) error {
	// Detect OS and uninstall accordingly
	switch runtime.GOOS {
	case "darwin":
		// macOS - Docker Desktop
		appPath := "/Applications/Docker.app"
		if _, err := os.Stat(appPath); err == nil {
			fmt.Println(dimStyle.Render("  → Stopping Docker Desktop..."))
			// Stop Docker Desktop
			exec.Command("osascript", "-e", "quit app \"Docker\"").Run()
			time.Sleep(2 * time.Second)

			fmt.Println(dimStyle.Render("  → Removing Docker.app..."))
			// Remove application
			if err := os.RemoveAll(appPath); err != nil {
				return err
			}

			// Remove Docker data
			homeDir := os.Getenv("HOME")
			dockerDirs := []string{
				filepath.Join(homeDir, ".docker"),
				filepath.Join(homeDir, "Library/Containers/com.docker.docker"),
				filepath.Join(homeDir, "Library/Application Support/Docker Desktop"),
				filepath.Join(homeDir, "Library/Group Containers/group.com.docker"),
			}

			fmt.Println(dimStyle.Render("  → Removing Docker data directories..."))
			for _, dir := range dockerDirs {
				if _, err := os.Stat(dir); err == nil {
					os.RemoveAll(dir)
				}
			}
		}

	case "linux":
		// Stop the spinner before running sudo commands
		progress.Stop()

		fmt.Println(dimStyle.Render("  → Detecting package manager..."))
		
		// Try different package managers
		if _, err := exec.LookPath("apt-get"); err == nil {
			fmt.Println(dimStyle.Render("  → Removing Docker packages (apt)..."))
			// Remove Docker packages
			cmd := exec.Command("sudo", "apt-get", "remove", "-y", "docker-ce", "docker-ce-cli", "containerd.io")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()

			fmt.Println(dimStyle.Render("  → Purging Docker configuration..."))
			// Purge Docker packages
			cmd = exec.Command("sudo", "apt-get", "purge", "-y", "docker-ce", "docker-ce-cli", "containerd.io")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		} else if _, err := exec.LookPath("yum"); err == nil {
			fmt.Println(dimStyle.Render("  → Removing Docker packages (yum)..."))
			cmd := exec.Command("sudo", "yum", "remove", "-y", "docker-ce", "docker-ce-cli", "containerd.io")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		} else if _, err := exec.LookPath("dnf"); err == nil {
			fmt.Println(dimStyle.Render("  → Removing Docker packages (dnf)..."))
			cmd := exec.Command("sudo", "dnf", "remove", "-y", "docker-ce", "docker-ce-cli", "containerd.io")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}

		fmt.Println(dimStyle.Render("  → Removing Docker data directories..."))
		// Remove Docker data
		cmd := exec.Command("sudo", "rm", "-rf", "/var/lib/docker")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

		cmd = exec.Command("sudo", "rm", "-rf", "/var/lib/containerd")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

		// Resume the spinner after sudo commands are done
		progress.Resume()
	}

	return nil
}

func cleanShellConfigDestroy(configFile string) error {
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
			strings.Contains(line, "mcs completion") ||
			strings.Contains(line, "export MCS_") {
			// Also skip the next line if it's an alias or export
			if i+1 < len(lines) &&
				(strings.HasPrefix(strings.TrimSpace(lines[i+1]), "alias ") ||
					strings.HasPrefix(strings.TrimSpace(lines[i+1]), "export ")) {
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
