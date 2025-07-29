package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/michaelkeevildown/mcs/internal/update"
	"github.com/spf13/cobra"
)

// CheckUpdatesCommand creates the 'check-updates' command
func CheckUpdatesCommand() *cobra.Command {
	var (
		verbose bool
		json    bool
	)

	cmd := &cobra.Command{
		Use:   "check-updates",
		Short: "🔍 Check for code-server and MCS image updates",
		Long: `Check for available updates to the code-server base image and MCS Docker images.

This command will:
  - Check the latest code-server release on GitHub
  - Compare with your local code-server version
  - Show all MCS images and their build dates
  - Recommend updates if available`,
		Example: `  # Check for updates
  mcs check-updates
  
  # Show detailed information
  mcs check-updates --verbose
  
  # Output in JSON format
  mcs check-updates --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Show header
			if !json {
				ui.ShowHeader()
			}

			// Create update checker
			checker := update.NewUpdateChecker()
			progress := ui.NewProgress()

			// Check for code-server updates
			if !json {
				progress.Start("Checking for code-server updates")
			}

			hasUpdate, latest, localVersion, err := checker.CheckForUpdates(ctx)
			if err != nil {
				if !json {
					progress.Fail("Failed to check updates")
				}
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			if !json {
				progress.Success("Update check completed")
				fmt.Println()
			}

			// Get MCS images info
			mcsImages, err := checker.GetMCSImages(ctx)
			if err != nil && !json {
				fmt.Printf("⚠️  Warning: Could not list MCS images: %v\n", err)
			}

			if json {
				// JSON output
				fmt.Printf(`{
  "code_server": {
    "current_version": "%s",
    "latest_version": "%s",
    "update_available": %t,
    "release_date": "%s",
    "release_url": "%s"
  },
  "mcs_images": [`, 
					localVersion, 
					latest.TagName, 
					hasUpdate,
					latest.PublishedAt.Format(time.RFC3339),
					latest.HTMLURL)

				for i, img := range mcsImages {
					if i > 0 {
						fmt.Print(",")
					}
					fmt.Printf(`
    {
      "repository": "%s",
      "tag": "%s",
      "created": "%s",
      "size": "%s"
    }`, img.Repository, img.Tag, img.Created.Format(time.RFC3339), img.Size)
				}

				fmt.Println(`
  ]
}`)
				return nil
			}

			// Regular output
			fmt.Println("📊 Code-Server Status")
			fmt.Println(strings.Repeat("─", 50))
			
			if localVersion == "unknown" {
				fmt.Printf("  Local Version:  ⚠️  Unable to determine\n")
			} else {
				fmt.Printf("  Local Version:  %s\n", localVersion)
			}
			fmt.Printf("  Latest Version: %s (released %s)\n", 
				latest.TagName, 
				latest.PublishedAt.Format("Jan 2, 2006"))

			if hasUpdate {
				fmt.Printf("  Status:         🆕 Update available!\n")
			} else {
				fmt.Printf("  Status:         ✅ Up to date\n")
			}

			if verbose && latest.HTMLURL != "" {
				fmt.Printf("  Release Notes:  %s\n", latest.HTMLURL)
			}

			fmt.Println()

			// Show MCS images
			if len(mcsImages) > 0 {
				fmt.Println("📦 MCS Docker Images")
				fmt.Println(strings.Repeat("─", 50))
				
				for _, img := range mcsImages {
					age := time.Since(img.Created)
					ageStr := formatAge(age)
					
					fmt.Printf("  • %-30s %s (%s)\n", 
						img.Repository + ":" + img.Tag,
						img.Size,
						ageStr)
				}
				fmt.Println()
			}

			// Show recommendations
			if hasUpdate {
				fmt.Println("💡 Recommendations:")
				fmt.Println("  • Run 'mcs update-images' to update to the latest version")
				fmt.Println("  • This will pull the latest code-server and rebuild MCS images")
				
				if verbose {
					fmt.Println()
					fmt.Println("📋 What's New:")
					// Show truncated release notes
					if latest.Body != "" {
						lines := strings.Split(latest.Body, "\n")
						for i, line := range lines {
							if i > 10 {
								fmt.Println("  ... (see full release notes online)")
								break
							}
							if line != "" {
								fmt.Printf("  %s\n", line)
							}
						}
					}
				}
			} else {
				fmt.Println("✨ Your code-server installation is up to date!")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	cmd.Flags().BoolVar(&json, "json", false, "Output in JSON format")

	return cmd
}

// formatAge formats a duration into a human-readable age string
func formatAge(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days > 0 {
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	
	hours := int(d.Hours())
	if hours > 0 {
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	
	minutes := int(d.Minutes())
	if minutes > 0 {
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	
	return "just now"
}