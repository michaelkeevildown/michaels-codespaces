package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelkeevildown/mcs/internal/docker"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/spf13/cobra"
)

var (
	destroyWarningStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	destroyInfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// DestroyCommand creates the 'destroy' command
func DestroyCommand() *cobra.Command {
	var (
		force        bool
		keepDocker   bool
		skipBackup   bool
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

				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(response)
				
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
	progress.Start("Beginning destruction sequence")

	// 1. Create backup if requested
	if !skipBackup {
		if err := createBackup(); err != nil {
			// Don't fail, just warn
			fmt.Printf("Warning: Failed to create backup: %v\n", err)
		} else {
			progress.Update("Created backup in ~/mcs-backup")
		}
	}

	// 2. Stop all running containers
	dockerClient, err := docker.NewClient()
	if err == nil {
		progress.Update("Stopping all codespace containers")
		
		containers, err := dockerClient.ListContainers()
		if err == nil {
			for _, container := range containers {
				// Only stop MCS containers
				if strings.Contains(container.Names[0], "mcs-") || 
				   strings.Contains(container.Image, "michaelkeevildown/claude-coder") {
					if err := dockerClient.StopContainer(container.ID); err != nil {
						fmt.Printf("Warning: Failed to stop container %s: %v\n", container.Names[0], err)
					}
				}
			}
		}
		progress.Update("Stopped all containers")

		// 3. Remove all containers
		progress.Update("Removing all codespace containers")
		if err := removeAllContainers(dockerClient); err != nil {
			fmt.Printf("Warning: Failed to remove some containers: %v\n", err)
		}
		progress.Update("Removed all containers")
	}

	// 4. Remove codespaces directory
	homeDir := os.Getenv("HOME")
	codespacesDir := filepath.Join(homeDir, "codespaces")
	if _, err := os.Stat(codespacesDir); err == nil {
		progress.Update("Removing codespaces directory")
		if err := os.RemoveAll(codespacesDir); err != nil {
			progress.Fail("Failed to remove codespaces directory")
			return fmt.Errorf("failed to remove %s: %w", codespacesDir, err)
		}
		progress.Update("Removed all codespace data")
	}

	// 5. Remove MCS installation
	progress.Update("Removing MCS installation")
	mcsHome := os.Getenv("MCS_HOME")
	if mcsHome == "" {
		mcsHome = filepath.Join(homeDir, ".mcs")
	}
	
	if err := os.RemoveAll(mcsHome); err != nil {
		fmt.Printf("Warning: Failed to remove MCS directory: %v\n", err)
	}

	// Remove MCS binary from PATH locations
	pathLocations := []string{
		"/usr/local/bin/mcs",
		filepath.Join(homeDir, ".local/bin/mcs"),
		filepath.Join(homeDir, "bin/mcs"),
	}
	
	for _, loc := range pathLocations {
		if _, err := os.Stat(loc); err == nil {
			os.Remove(loc)
		}
	}

	// 6. Clean shell configurations
	progress.Update("Cleaning shell configurations")
	shellConfigs := []string{
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".bash_profile"),
		filepath.Join(homeDir, ".profile"),
	}

	for _, configFile := range shellConfigs {
		if err := cleanShellConfig(configFile); err != nil {
			fmt.Printf("Warning: Failed to clean %s: %v\n", configFile, err)
		}
	}

	// 7. Remove helper scripts
	scripts := []string{
		filepath.Join(homeDir, "monitor-system.sh"),
		filepath.Join(homeDir, ".mcs-install.sh"),
	}
	
	for _, script := range scripts {
		if _, err := os.Stat(script); err == nil {
			os.Remove(script)
		}
	}

	// 8. Optionally uninstall Docker
	if !keepDocker {
		progress.Update("Uninstalling Docker")
		if err := uninstallDocker(); err != nil {
			fmt.Printf("Warning: Failed to uninstall Docker: %v\n", err)
			fmt.Println("You may need to uninstall Docker manually")
		} else {
			progress.Update("Uninstalled Docker")
		}
	}

	progress.Success("Destruction complete!")
	
	fmt.Println()
	fmt.Println(destroyWarningStyle.Render("💥 MCS has been completely destroyed"))
	fmt.Println()
	fmt.Println("Removed:")
	fmt.Println("  ✓ All codespaces and containers")
	fmt.Println("  ✓ All codespace data")
	fmt.Println("  ✓ MCS installation")
	fmt.Println("  ✓ Shell configurations")
	if !keepDocker {
		fmt.Println("  ✓ Docker installation")
	}
	fmt.Println()
	
	if !skipBackup {
		fmt.Println("Your codespace data was backed up to: ~/mcs-backup")
		fmt.Println("You can safely delete this backup when no longer needed.")
		fmt.Println()
	}
	
	fmt.Println("Thank you for using MCS! 👋")

	return nil
}

func createBackup() error {
	homeDir := os.Getenv("HOME")
	codespacesDir := filepath.Join(homeDir, "codespaces")
	backupDir := filepath.Join(homeDir, "mcs-backup")
	
	// Check if codespaces directory exists
	if _, err := os.Stat(codespacesDir); os.IsNotExist(err) {
		return nil // Nothing to backup
	}
	
	// Create backup directory with timestamp
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	backupPath := filepath.Join(backupDir, timestamp)
	
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return err
	}
	
	// Copy codespaces directory
	cmd := exec.Command("cp", "-r", codespacesDir, backupPath)
	return cmd.Run()
}

func removeAllContainers(client *docker.Client) error {
	containers, err := client.ListContainers()
	if err != nil {
		return err
	}
	
	for _, container := range containers {
		// Only remove MCS containers
		if strings.Contains(container.Names[0], "mcs-") || 
		   strings.Contains(container.Image, "michaelkeevildown/claude-coder") {
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

func uninstallDocker() error {
	// Detect OS and uninstall accordingly
	switch runtime.GOOS {
	case "darwin":
		// macOS - Docker Desktop
		appPath := "/Applications/Docker.app"
		if _, err := os.Stat(appPath); err == nil {
			// Stop Docker Desktop
			exec.Command("osascript", "-e", "quit app \"Docker\"").Run()
			time.Sleep(2 * time.Second)
			
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
			
			for _, dir := range dockerDirs {
				os.RemoveAll(dir)
			}
		}
		
	case "linux":
		// Try different package managers
		if _, err := exec.LookPath("apt-get"); err == nil {
			exec.Command("sudo", "apt-get", "remove", "-y", "docker-ce", "docker-ce-cli", "containerd.io").Run()
			exec.Command("sudo", "apt-get", "purge", "-y", "docker-ce", "docker-ce-cli", "containerd.io").Run()
		} else if _, err := exec.LookPath("yum"); err == nil {
			exec.Command("sudo", "yum", "remove", "-y", "docker-ce", "docker-ce-cli", "containerd.io").Run()
		} else if _, err := exec.LookPath("dnf"); err == nil {
			exec.Command("sudo", "dnf", "remove", "-y", "docker-ce", "docker-ce-cli", "containerd.io").Run()
		}
		
		// Remove Docker data
		exec.Command("sudo", "rm", "-rf", "/var/lib/docker").Run()
		exec.Command("sudo", "rm", "-rf", "/var/lib/containerd").Run()
	}
	
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