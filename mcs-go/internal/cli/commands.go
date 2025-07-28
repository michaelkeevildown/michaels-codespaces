package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/michaelkeevildown/mcs/internal/codespace"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/spf13/cobra"
)

// StartCommand creates the 'start' command
func StartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start <name>",
		Short: "▶️  Start a codespace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show beautiful header
			ui.ShowHeader()
			
			ctx := context.Background()
			name := args[0]
			
			// Create progress indicator
			progress := ui.NewProgress()
			progress.Start(fmt.Sprintf("Starting codespace %s", name))
			
			// Create manager
			manager := codespace.NewManager()
			
			// Start codespace
			if err := manager.Start(ctx, name); err != nil {
				progress.Fail(fmt.Sprintf("Failed to start %s", name))
				return err
			}
			
			progress.Success(fmt.Sprintf("Started %s", name))
			
			// Get codespace info
			cs, err := manager.Get(ctx, name)
			if err == nil {
				fmt.Println()
				fmt.Printf("🔗 VS Code: %s\n", urlStyle.Render(cs.VSCodeURL))
				fmt.Printf("🌐 App: %s\n", urlStyle.Render(cs.AppURL))
				if cs.Password != "" {
					fmt.Printf("🔑 Password: %s\n", infoStyle.Render(cs.Password))
				}
			}
			
			return nil
		},
	}
}

// StopCommand creates the 'stop' command
func StopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <name>",
		Short: "⏸️  Stop a codespace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show beautiful header
			ui.ShowHeader()
			
			ctx := context.Background()
			name := args[0]
			
			// Create progress indicator
			progress := ui.NewProgress()
			progress.Start(fmt.Sprintf("Stopping codespace %s", name))
			
			// Create manager
			manager := codespace.NewManager()
			
			// Stop codespace
			if err := manager.Stop(ctx, name); err != nil {
				progress.Fail(fmt.Sprintf("Failed to stop %s", name))
				return err
			}
			
			progress.Success(fmt.Sprintf("Stopped %s", name))
			return nil
		},
	}
}

// RestartCommand creates the 'restart' command
func RestartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restart <name>",
		Short: "🔄 Restart a codespace",
		Long:  "Stop and then start a codespace.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show beautiful header
			ui.ShowHeader()
			
			ctx := context.Background()
			name := args[0]
			
			// Create manager
			manager := codespace.NewManager()
			
			// First stop
			progress := ui.NewProgress()
			progress.Start(fmt.Sprintf("Stopping codespace %s", name))
			
			if err := manager.Stop(ctx, name); err != nil {
				// If stop fails because it's not running, continue to start
				if !strings.Contains(err.Error(), "not found") {
					progress.Update(fmt.Sprintf("Stop failed, attempting to start %s anyway", name))
				}
			} else {
				progress.Success(fmt.Sprintf("Stopped %s", name))
			}
			
			// Then start
			progress = ui.NewProgress()
			progress.Start(fmt.Sprintf("Starting codespace %s", name))
			
			if err := manager.Start(ctx, name); err != nil {
				progress.Fail(fmt.Sprintf("Failed to start %s", name))
				return err
			}
			
			progress.Success(fmt.Sprintf("Restarted %s", name))
			
			// Get codespace info to show URLs
			cs, err := manager.Get(ctx, name)
			if err == nil {
				fmt.Println()
				fmt.Printf("🔗 VS Code: %s\n", urlStyle.Render(cs.VSCodeURL))
				fmt.Printf("🌐 App: %s\n", urlStyle.Render(cs.AppURL))
				if cs.Password != "" {
					fmt.Printf("🔑 Password: %s\n", infoStyle.Render(cs.Password))
				}
			}
			
			return nil
		},
	}
}

// RemoveCommand creates the 'remove' command
func RemoveCommand() *cobra.Command {
	var force bool
	
	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm", "delete"},
		Short:   "🗑️  Remove a codespace",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show beautiful header
			ui.ShowHeader()
			
			ctx := context.Background()
			name := args[0]
			
			// Confirm if not forced
			if !force {
				fmt.Printf("⚠️  This will permanently delete the codespace '%s' and all its data.\n", name)
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
			progress.Start(fmt.Sprintf("Removing codespace %s", name))
			
			// Create manager
			manager := codespace.NewManager()
			
			// Remove codespace
			if err := manager.Remove(ctx, name, force); err != nil {
				progress.Fail(fmt.Sprintf("Failed to remove %s", name))
				return err
			}
			
			progress.Success(fmt.Sprintf("Removed %s", name))
			return nil
		},
	}
	
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force removal without confirmation")
	
	return cmd
}


// UpdateCommand creates the 'update' command
func UpdateCommand() *cobra.Command {
	var check bool
	
	cmd := &cobra.Command{
		Use:   "update",
		Short: "🔄 Update MCS to the latest version",
		Long:  "Update MCS by pulling latest changes from git and rebuilding from source.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show beautiful header
			ui.ShowHeader()
			
			return updateMCS(check)
		},
	}
	
	cmd.Flags().BoolVar(&check, "check", false, "Only check for updates without installing")
	
	return cmd
}

func updateMCS(checkOnly bool) error {
	progress := ui.NewProgress()
	
	// Get MCS home directory
	mcsHome := os.Getenv("MCS_HOME")
	if mcsHome == "" {
		mcsHome = filepath.Join(os.Getenv("HOME"), ".mcs")
	}
	
	// Check if running from source installation
	// Source code is now in ~/.mcs/source
	sourceDir := filepath.Join(mcsHome, "source")
	gitDir := filepath.Join(sourceDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Check old location for backward compatibility
		oldGitDir := filepath.Join(mcsHome, ".git")
		if _, err := os.Stat(oldGitDir); err == nil {
			return fmt.Errorf("MCS source in wrong location. Run 'mcs setup' to fix")
		}
		return fmt.Errorf("MCS source not found. Run 'mcs setup' to clone source repository")
	}
	
	if checkOnly {
		progress.Start("Checking for updates")
		
		// Fetch latest changes
		cmd := exec.Command("git", "fetch", "origin", "main")
		cmd.Dir = sourceDir
		if err := cmd.Run(); err != nil {
			progress.Fail("Failed to check for updates")
			return fmt.Errorf("failed to fetch updates: %w", err)
		}
		
		// Check if we're behind
		cmd = exec.Command("git", "rev-list", "--count", "HEAD..origin/main")
		cmd.Dir = sourceDir
		output, err := cmd.Output()
		if err != nil {
			progress.Fail("Failed to check update status")
			return fmt.Errorf("failed to check update status: %w", err)
		}
		
		count := strings.TrimSpace(string(output))
		if count == "0" {
			progress.Success("You are running the latest version")
		} else {
			progress.Success(fmt.Sprintf("Updates available: %s commits behind", count))
			fmt.Println("\nRun 'mcs update' to install updates")
		}
		return nil
	}
	
	progress.Start("Updating MCS from source")
	
	// Get current version before update
	cmdVersion := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	cmdVersion.Dir = sourceDir
	currentVersion, _ := cmdVersion.Output()
	
	// Pull latest changes
	progress.Update("Pulling latest changes")
	cmd := exec.Command("git", "pull", "origin", "main")
	cmd.Dir = sourceDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		progress.Fail("Failed to pull updates")
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, output)
	}
	
	// Check if Go is available
	if _, err := exec.LookPath("go"); err != nil {
		progress.Fail("Go compiler not found")
		return fmt.Errorf("Go is required to build from source. Please install Go and try again")
	}
	
	// Change to mcs-go directory
	mcsGoDir := filepath.Join(sourceDir, "mcs-go")
	
	// Download dependencies
	progress.Update("Downloading dependencies")
	cmd = exec.Command("go", "mod", "download")
	cmd.Dir = mcsGoDir
	if err := cmd.Run(); err != nil {
		progress.Fail("Failed to download dependencies")
		return fmt.Errorf("go mod download failed: %w", err)
	}
	
	// Build new binary
	progress.Update("Building MCS")
	newVersion := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	newVersion.Dir = sourceDir
	versionOutput, _ := newVersion.Output()
	version := strings.TrimSpace(string(versionOutput))
	
	binPath := filepath.Join(mcsHome, "bin", "mcs")
	cmd = exec.Command("go", "build", 
		"-ldflags", fmt.Sprintf("-X main.version=%s", version),
		"-o", binPath,
		"cmd/mcs/main.go")
	cmd.Dir = mcsGoDir
	
	if output, err := cmd.CombinedOutput(); err != nil {
		progress.Fail("Failed to build MCS")
		return fmt.Errorf("build failed: %w\nOutput: %s", err, output)
	}
	
	progress.Success("Successfully updated MCS")
	
	// Show what changed if there were updates
	if string(currentVersion) != version {
		fmt.Println("\nChanges in this update:")
		cmd = exec.Command("git", "log", "--oneline", "--no-decorate",
			fmt.Sprintf("%s..%s", strings.TrimSpace(string(currentVersion)), version))
		cmd.Dir = sourceDir
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Printf("  • %s\n", line)
				}
			}
		}
	}
	
	fmt.Printf("\nUpdated to version: %s\n", version)
	fmt.Println("\nPlease restart any running MCS commands to use the new version.")
	
	return nil
}