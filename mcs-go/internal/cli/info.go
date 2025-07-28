package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/michaelkeevildown/mcs/internal/codespace"
	"github.com/michaelkeevildown/mcs/internal/docker"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/spf13/cobra"
)


// InfoCommand creates the 'info' command
func InfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "info <name>",
		Aliases: []string{"show"},
		Short:   "ℹ️  Show detailed codespace information",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.ShowHeader()
			ctx := context.Background()
			name := args[0]
			
			// Create manager
			manager := codespace.NewManager()
			
			// Get codespace
			cs, err := manager.Get(ctx, name)
			if err != nil {
				return fmt.Errorf("codespace '%s' not found", name)
			}
			
			// Create Docker client for container stats
			dockerClient, err := docker.NewClient()
			if err == nil {
				defer dockerClient.Close()
			}
			
			// Display information
			fmt.Println()
			fmt.Println(sectionStyle.Render("Codespace Information"))
			fmt.Println(strings.Repeat("━", 50))
			fmt.Println()
			
			// Basic info
			fmt.Printf("%s %s\n", infoStyle.Render("Name:"), cs.Name)
			fmt.Printf("%s %s\n", infoStyle.Render("Repository:"), cs.Repository)
			fmt.Printf("%s %s\n", infoStyle.Render("Language:"), cs.Language)
			fmt.Printf("%s %s\n", infoStyle.Render("Created:"), cs.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("%s %s\n", infoStyle.Render("Location:"), cs.Path)
			
			// Components
			if len(cs.Components) > 0 {
				fmt.Printf("%s %s\n", infoStyle.Render("Components:"), strings.Join(cs.Components, ", "))
			}
			
			fmt.Println()
			fmt.Println(sectionStyle.Render("Access Details"))
			
			// Access details
			fmt.Printf("%s %s\n", infoStyle.Render("VS Code URL:"), urlStyle.Render(cs.VSCodeURL))
			if cs.Password != "" {
				fmt.Printf("%s %s\n", infoStyle.Render("Password:"), urlStyle.Render(cs.Password))
			}
			if cs.AppURL != "" {
				fmt.Printf("%s %s\n", infoStyle.Render("App URL:"), urlStyle.Render(cs.AppURL))
			}
			
			// Container status
			fmt.Println()
			fmt.Println(sectionStyle.Render("Container Status"))
			
			if dockerClient != nil {
				containerName := fmt.Sprintf("%s-dev", name)
				container, err := dockerClient.GetContainerByName(ctx, containerName)
				if err != nil {
					fmt.Printf("%s %s\n", infoStyle.Render("Status:"), stoppedStyle.Render("○ Not found"))
				} else {
					if container.State == "running" {
						fmt.Printf("%s %s\n", infoStyle.Render("Status:"), runningStyle.Render("● Running"))
						
						// Get container stats
						stats, err := dockerClient.GetContainerStats(ctx, container.ID)
						if err == nil {
							fmt.Printf("%s %.2f%%\n", infoStyle.Render("CPU Usage:"), stats.CPUPercent)
							fmt.Printf("%s %s / %s (%.1f%%)\n", 
								infoStyle.Render("Memory:"), 
								formatBytes(stats.MemoryUsage),
								formatBytes(stats.MemoryLimit),
								(float64(stats.MemoryUsage)/float64(stats.MemoryLimit))*100,
							)
						}
						
						// Uptime
						if container.Created > 0 {
							uptime := time.Since(time.Unix(container.Created, 0))
							fmt.Printf("%s %s\n", infoStyle.Render("Uptime:"), formatDuration(uptime))
						}
					} else {
						fmt.Printf("%s %s\n", infoStyle.Render("Status:"), stoppedStyle.Render("○ Stopped"))
					}
					
					// Container ID
					fmt.Printf("%s %s\n", infoStyle.Render("Container ID:"), container.ID)
				}
			}
			
			// Disk usage
			fmt.Println()
			fmt.Println(sectionStyle.Render("Disk Usage"))
			
			if info, err := os.Stat(cs.Path); err == nil && info.IsDir() {
				size, _ := getDirSize(cs.Path)
				fmt.Printf("%s %s\n", infoStyle.Render("Total Size:"), formatBytes(uint64(size)))
			}
			
			// Quick actions
			fmt.Println()
			fmt.Println(sectionStyle.Render("Quick Actions"))
			if cs.Status == "running" {
				fmt.Printf("  %s\n", infoStyle.Render("mcs exec "+name+" bash"))
				fmt.Printf("  %s\n", infoStyle.Render("mcs logs -f "+name))
				fmt.Printf("  %s\n", infoStyle.Render("mcs stop "+name))
			} else {
				fmt.Printf("  %s\n", infoStyle.Render("mcs start "+name))
				fmt.Printf("  %s\n", infoStyle.Render("mcs remove "+name))
			}
			fmt.Println()
			
			return nil
		},
	}
}

// formatBytes formats bytes into human readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatDuration formats a duration into a human readable format
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// getDirSize calculates the total size of a directory
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}