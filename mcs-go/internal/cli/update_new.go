package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/michaelkeevildown/mcs/internal/update"
	"github.com/michaelkeevildown/mcs/internal/version"
	"github.com/michaelkeevildown/mcs/pkg/utils"
	"github.com/spf13/cobra"
)

// UpdateNewCommand creates an enhanced update command
func UpdateNewCommand() *cobra.Command {
	var (
		fromSource bool
		dev        bool
		check      bool
	)
	
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update MCS to the latest version",
		Long: `Update MCS to the latest version.

By default, this will download the latest stable release from GitHub.
Use --source to build from source instead.
Use --dev to get the latest development build.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if check {
				return checkForUpdates()
			}
			
			if fromSource {
				return updateFromSource()
			}
			
			if dev {
				return updateToDev()
			}
			
			return updateToLatest()
		},
	}
	
	cmd.Flags().BoolVar(&fromSource, "source", false, "Update by building from source")
	cmd.Flags().BoolVar(&dev, "dev", false, "Update to latest development build")
	cmd.Flags().BoolVar(&check, "check", false, "Check for updates without installing")
	
	return cmd
}

func checkForUpdates() error {
	progress := utils.NewProgressIndicator()
	progress.Start("Checking for updates")
	
	currentVersion := version.Info()
	
	latest, err := update.GetLatestRelease()
	if err != nil {
		progress.Fail("Failed to check for updates")
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	
	if update.IsNewerVersion(currentVersion, latest.TagName) {
		progress.Success(fmt.Sprintf("Update available: %s → %s", currentVersion, latest.TagName))
		fmt.Println("\nRun 'mcs update' to install the latest version")
	} else {
		progress.Success("You are running the latest version")
	}
	
	// Also check if running a dev build
	if version.IsDevBuild() {
		fmt.Println("\nNote: You are running a development build.")
		fmt.Println("Use 'mcs update --dev' to get the latest dev build")
		fmt.Println("Use 'mcs update' to switch to the latest stable release")
	}
	
	return nil
}

func updateToLatest() error {
	progress := utils.NewProgressIndicator()
	progress.Start("Checking for updates")
	
	currentVersion := version.Info()
	
	latest, err := update.GetLatestRelease()
	if err != nil {
		progress.Fail("Failed to check for updates")
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	
	if !update.IsNewerVersion(currentVersion, latest.TagName) && !version.IsDevBuild() {
		progress.Success("Already running the latest version")
		return nil
	}
	
	progress.Update(fmt.Sprintf("Downloading MCS %s", latest.TagName))
	
	if err := update.DownloadAndInstall(latest); err != nil {
		progress.Fail("Failed to download update")
		return fmt.Errorf("failed to install update: %w", err)
	}
	
	progress.Success(fmt.Sprintf("Successfully updated to %s", latest.TagName))
	
	fmt.Println("\nUpdate complete! Please run 'mcs version' to verify.")
	
	return nil
}

func updateToDev() error {
	progress := utils.NewProgressIndicator()
	progress.Start("Checking for development updates")
	
	devRelease, err := update.GetDevRelease()
	if err != nil {
		progress.Fail("Failed to check for dev updates")
		return fmt.Errorf("failed to fetch dev release: %w", err)
	}
	
	progress.Update("Downloading latest development build")
	
	if err := update.DownloadAndInstall(devRelease); err != nil {
		progress.Fail("Failed to download dev update")
		return fmt.Errorf("failed to install dev update: %w", err)
	}
	
	progress.Success("Successfully updated to latest development build")
	
	fmt.Println("\n⚠️  Warning: You are now running a development build")
	fmt.Println("Development builds may be unstable. Use 'mcs update' to return to stable.")
	
	return nil
}

func updateFromSource() error {
	progress := utils.NewProgressIndicator()
	
	// Check if we have the source repository
	mcsHome := os.Getenv("MCS_HOME")
	if mcsHome == "" {
		home, _ := os.UserHomeDir()
		mcsHome = filepath.Join(home, ".mcs")
	}
	
	if _, err := os.Stat(filepath.Join(mcsHome, ".git")); os.IsNotExist(err) {
		progress.Fail("Source repository not found")
		fmt.Println("\nTo update from source, you need to have cloned the repository.")
		fmt.Println("Run the installer with --source flag:")
		fmt.Println("  curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/main/mcs-go/install.sh | bash -s -- --source")
		return fmt.Errorf("source repository not found at %s", mcsHome)
	}
	
	progress.Start("Updating from source")
	
	if err := update.BuildFromSource(); err != nil {
		progress.Fail("Failed to build from source")
		return err
	}
	
	progress.Success("Successfully updated from source")
	
	fmt.Println("\nUpdate complete! Please run 'mcs version' to verify.")
	
	return nil
}