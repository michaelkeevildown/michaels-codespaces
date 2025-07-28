package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
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
	// Create a tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	// Print header
	fmt.Fprintln(w, headerStyle.Render("NAME\tSTATUS\tREPOSITORY\tPORTS\tCREATED"))
	
	// Print codespaces
	for _, cs := range codespaces {
		status := formatStatus(cs.Status)
		repo := truncateRepo(cs.Repository, 40)
		ports := formatPorts(cs)
		created := formatTime(cs.CreatedAt)
		
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			cs.Name,
			status,
			repo,
			ports,
			created,
		)
	}
	
	return w.Flush()
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

func formatStatus(status string) string {
	switch status {
	case "running":
		return runningStyle.Render("● running")
	case "stopped":
		return stoppedStyle.Render("○ stopped")
	default:
		return status
	}
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