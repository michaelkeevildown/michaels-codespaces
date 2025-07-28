package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/michaelkeevildown/mcs/internal/codespace"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/spf13/cobra"
)


// ListCommand creates the 'list' command
func ListCommand() *cobra.Command {
	var (
		showAll bool
		format  string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "📋 List all codespaces",
		Long:    "List all codespaces with their current status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.ShowHeader()
			ctx := context.Background()
			
			// Create manager
			manager := codespace.NewManager()
			
			// List codespaces
			codespaces, err := manager.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list codespaces: %w", err)
			}

			// Filter if not showing all
			if !showAll {
				var filtered []codespace.Codespace
				for _, cs := range codespaces {
					if cs.Status == "running" {
						filtered = append(filtered, cs)
					}
				}
				codespaces = filtered
			}

			// Handle empty list
			if len(codespaces) == 0 {
				if showAll {
					fmt.Println("No codespaces found.")
				} else {
					fmt.Println("No running codespaces found. Use --all to see all codespaces.")
				}
				return nil
			}

			// Display based on format
			switch format {
			case "json":
				return displayJSON(codespaces)
			case "simple":
				return displaySimple(codespaces)
			default:
				return displayTable(codespaces)
			}
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all codespaces (including stopped)")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, simple, json)")

	return cmd
}

func displayTable(codespaces []codespace.Codespace) error {
	// Calculate column widths
	colWidths := struct {
		name       int
		status     int
		repository int
		ports      int
		created    int
	}{
		name:       4, // "NAME"
		status:     10, // "STATUS" + icon
		repository: 10, // "REPOSITORY"
		ports:      5,  // "PORTS"
		created:    7,  // "CREATED"
	}
	
	// Prepare data for display
	type row struct {
		name       string
		status     string
		statusRaw  string
		repository string
		ports      string
		created    string
	}
	
	rows := make([]row, len(codespaces))
	
	// Process each codespace and calculate max widths
	for i, cs := range codespaces {
		r := row{
			name:       cs.Name,
			statusRaw:  cs.Status,
			repository: truncateRepo(cs.Repository, 40),
			ports:      formatPorts(cs),
			created:    formatTime(cs.CreatedAt),
		}
		
		// Format status separately to handle color codes
		if cs.Status == "running" {
			r.status = "● running"
		} else {
			r.status = "○ stopped"
		}
		
		// Update max widths
		if len(r.name) > colWidths.name {
			colWidths.name = len(r.name)
		}
		if len(r.repository) > colWidths.repository {
			colWidths.repository = len(r.repository)
		}
		if len(r.ports) > colWidths.ports {
			colWidths.ports = len(r.ports)
		}
		if len(r.created) > colWidths.created {
			colWidths.created = len(r.created)
		}
		
		rows[i] = r
	}
	
	// Add padding
	colWidths.name += 2
	colWidths.status += 2
	colWidths.repository += 2
	colWidths.ports += 2
	colWidths.created += 2
	
	// Print header
	header := fmt.Sprintf("%-*s%-*s%-*s%-*s%-*s",
		colWidths.name, "NAME",
		colWidths.status, "STATUS",
		colWidths.repository, "REPOSITORY",
		colWidths.ports, "PORTS",
		colWidths.created, "CREATED",
	)
	fmt.Println(headerStyle.Render(header))
	
	// Print separator line
	separator := strings.Repeat("─", len(header))
	fmt.Println(dividerStyle.Render(separator))
	
	// Print rows
	for _, r := range rows {
		// Apply color to status
		var statusDisplay string
		if r.statusRaw == "running" {
			statusDisplay = runningStyle.Render(r.status)
		} else {
			statusDisplay = stoppedStyle.Render(r.status)
		}
		
		// Format the row
		fmt.Printf("%-*s%-*s%-*s%-*s%-*s\n",
			colWidths.name, r.name,
			colWidths.status + getColorCodeWidth(statusDisplay, r.status), statusDisplay,
			colWidths.repository, r.repository,
			colWidths.ports, r.ports,
			colWidths.created, r.created,
		)
	}
	
	return nil
}

// getColorCodeWidth calculates the difference between rendered string length and visible length
func getColorCodeWidth(rendered, plain string) int {
	// This accounts for ANSI color codes that don't contribute to visible width
	return len(rendered) - len(plain)
}

func displaySimple(codespaces []codespace.Codespace) error {
	for _, cs := range codespaces {
		fmt.Printf("%s (%s)\n", cs.Name, cs.Status)
	}
	return nil
}

func displayJSON(codespaces []codespace.Codespace) error {
	// TODO: Implement JSON output
	return fmt.Errorf("JSON format not yet implemented")
}


func truncateRepo(repo string, maxLen int) string {
	// Remove common prefixes
	repo = strings.TrimPrefix(repo, "https://github.com/")
	repo = strings.TrimPrefix(repo, "git@github.com:")
	repo = strings.TrimSuffix(repo, ".git")
	
	if len(repo) > maxLen {
		return repo[:maxLen-3] + "..."
	}
	return repo
}

func formatPorts(cs codespace.Codespace) string {
	if cs.Status != "running" {
		return "-"
	}
	
	// Extract port from URL
	vsCodePort := extractPort(cs.VSCodeURL)
	appPort := extractPort(cs.AppURL)
	
	if vsCodePort != "" && appPort != "" {
		return fmt.Sprintf("%s, %s", vsCodePort, appPort)
	} else if vsCodePort != "" {
		return vsCodePort
	}
	return "-"
}

func extractPort(url string) string {
	parts := strings.Split(url, ":")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

func formatTime(t time.Time) string {
	duration := time.Since(t)
	
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(duration.Hours()/24))
	default:
		return t.Format("Jan 02")
	}
}