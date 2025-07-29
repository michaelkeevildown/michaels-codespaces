package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/michaelkeevildown/mcs/internal/config"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/michaelkeevildown/mcs/internal/update"
	"github.com/spf13/cobra"
)

// UpdateImagesCommand creates the 'update-images' command
func UpdateImagesCommand() *cobra.Command {
	var (
		force    bool
		onlyPull bool
		skipPull bool
	)

	cmd := &cobra.Command{
		Use:   "update-images",
		Short: "🔄 Update code-server and rebuild MCS images",
		Long: `Update the code-server base image and rebuild all MCS Docker images.

This command will:
  - Pull the latest code-server image from Docker Hub
  - Rebuild all MCS images with the new base
  - Show progress and results
  - Optionally clean up old images`,
		Example: `  # Update everything (pull + rebuild)
  mcs update-images
  
  # Only pull the latest code-server image
  mcs update-images --only-pull
  
  # Skip pulling and just rebuild MCS images
  mcs update-images --skip-pull
  
  # Force update even if already up to date
  mcs update-images --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Show header
			ui.ShowHeader()

			// Create update checker
			checker := update.NewUpdateChecker()
			progress := ui.NewProgress()

			// Check for updates first (unless forced)
			if !force && !skipPull {
				progress.Start("Checking for updates")
				hasUpdate, latest, localVersion, err := checker.CheckForUpdates(ctx)
				if err != nil {
					progress.Fail("Failed to check updates")
					// Continue anyway since user wants to update
				} else {
					progress.Success("Update check completed")
					
					if !hasUpdate {
						fmt.Println()
						fmt.Printf("✅ Already up to date! (version %s)\n", localVersion)
						fmt.Println("   Use --force to rebuild anyway")
						return nil
					}
					
					fmt.Println()
					fmt.Printf("🆕 Update available: %s → %s\n", localVersion, latest.TagName)
				}
			}

			// Pull latest code-server image
			if !skipPull {
				fmt.Println()
				progress = ui.NewProgress()
				progress.Start("Pulling latest code-server image")
				
				err := checker.PullLatestCodeServer(ctx, func(msg string) {
					// Update progress message if needed
				})
				
				if err != nil {
					progress.Fail("Failed to pull image")
					return fmt.Errorf("failed to pull latest code-server: %w", err)
				}
				
				progress.Success("Successfully pulled latest code-server")
			}

			if onlyPull {
				fmt.Println()
				fmt.Println("✅ Image pull completed successfully!")
				return nil
			}

			// Get dockerfiles path
			dockerfilesPath := config.GetDockerfilesPath()
			if _, err := os.Stat(dockerfilesPath); os.IsNotExist(err) {
				return fmt.Errorf("dockerfiles directory not found at %s", dockerfilesPath)
			}

			// Check if build script exists
			buildScript := filepath.Join(dockerfilesPath, "build.sh")
			if _, err := os.Stat(buildScript); os.IsNotExist(err) {
				return fmt.Errorf("build script not found at %s", buildScript)
			}

			// Rebuild MCS images
			fmt.Println()
			progress = ui.NewProgress()
			progress.Start("Rebuilding MCS Docker images")
			
			err := checker.RebuildMCSImages(ctx, dockerfilesPath, func(msg string) {
				// Could update progress here if needed
			})
			
			if err != nil {
				progress.Fail("Failed to rebuild images")
				return fmt.Errorf("failed to rebuild MCS images: %w", err)
			}
			
			progress.Success("Successfully rebuilt all MCS images")

			// Get and display updated images
			fmt.Println()
			mcsImages, err := checker.GetMCSImages(ctx)
			if err == nil && len(mcsImages) > 0 {
				fmt.Println("📦 Updated MCS Images:")
				fmt.Println(strings.Repeat("─", 50))
				
				for _, img := range mcsImages {
					fmt.Printf("  ✅ %-30s %s\n", 
						img.Repository + ":" + img.Tag,
						img.Size)
				}
			}

			fmt.Println()
			fmt.Println("✨ Update completed successfully!")
			fmt.Println()
			fmt.Println("💡 Next steps:")
			fmt.Println("  • New codespaces will use the updated images")
			fmt.Println("  • To update existing codespaces:")
			fmt.Println("    1. Stop the codespace: mcs stop <name>")
			fmt.Println("    2. Rebuild it: mcs rebuild <name>")
			fmt.Println("    3. Start it again: mcs start <name>")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force update even if already up to date")
	cmd.Flags().BoolVar(&onlyPull, "only-pull", false, "Only pull the latest code-server image without rebuilding")
	cmd.Flags().BoolVar(&skipPull, "skip-pull", false, "Skip pulling and just rebuild MCS images")

	return cmd
}