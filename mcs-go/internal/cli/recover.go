package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelkeevildown/mcs/internal/codespace"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/spf13/cobra"
)

var (
	recoverHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	recoverInfoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	recoverURLStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
)

// RecoverCommand creates the 'recover' command
func RecoverCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "recover <name>",
		Short: "🔑 Recover codespace credentials",
		Long:  "Quickly recover and display the credentials for a codespace.",
		Args:  cobra.ExactArgs(1),
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
			
			// Display credentials
			fmt.Println()
			fmt.Println(recoverHeaderStyle.Render("Codespace Credentials Recovery"))
			fmt.Println(strings.Repeat("━", 50))
			fmt.Println()
			
			fmt.Printf("%s %s\n", recoverInfoStyle.Render("Codespace:"), name)
			fmt.Printf("%s %s\n", recoverInfoStyle.Render("Repository:"), cs.Repository)
			fmt.Println()
			
			fmt.Println(recoverHeaderStyle.Render("VS Code Access:"))
			fmt.Printf("  URL: %s\n", recoverURLStyle.Render(cs.VSCodeURL))
			if cs.Password != "" {
				fmt.Printf("  Password: %s\n", recoverURLStyle.Render(cs.Password))
			}
			fmt.Println()
			
			fmt.Println(recoverHeaderStyle.Render("Quick Connect:"))
			fmt.Printf("  1. Open: %s\n", cs.VSCodeURL)
			if cs.Password != "" {
				fmt.Printf("  2. Enter password: %s\n", cs.Password)
			}
			fmt.Println()
			fmt.Println(strings.Repeat("━", 50))
			fmt.Println()
			
			// Show status
			if cs.Status != "running" {
				fmt.Printf("%s Codespace is not running. Start it with: mcs start %s\n", 
					recoverInfoStyle.Render("Note:"), name)
			}
			
			return nil
		},
	}
}