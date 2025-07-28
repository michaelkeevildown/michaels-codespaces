package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelkeevildown/mcs/internal/config"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/spf13/cobra"
)

var (
	updateHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	updateLabelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	enabledStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	disabledStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// AutoUpdateCommand creates the 'autoupdate' command
func AutoUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autoupdate [subcommand]",
		Short: "🔄 Configure automatic update checking",
		Long: `Configure automatic update checking for MCS.
		
Subcommands:
  status   - Show current auto-update configuration (default)
  on       - Enable automatic update checking
  off      - Disable automatic update checking  
  interval - Set update check interval in seconds`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ui.ShowHeader()
			// Default to status if no subcommand
			if len(args) == 0 {
				return showAutoUpdateStatus()
			}

			subcommand := args[0]
			switch subcommand {
			case "status":
				return showAutoUpdateStatus()
			case "on", "enable":
				return setAutoUpdateEnabled(true)
			case "off", "disable":
				return setAutoUpdateEnabled(false)
			case "interval":
				if len(args) < 2 {
					return fmt.Errorf("interval value required")
				}
				return setAutoUpdateInterval(args[1])
			case "help", "--help", "-h":
				return cmd.Help()
			default:
				return fmt.Errorf("unknown subcommand: %s", subcommand)
			}
		},
	}

	return cmd
}

func showAutoUpdateStatus() error {
	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	conf := cfg.Get()

	fmt.Println()
	fmt.Println(updateHeaderStyle.Render("Auto-update Status"))
	fmt.Println(strings.Repeat("━", 40))
	fmt.Println()

	// Status
	status := disabledStyle.Render("disabled")
	if conf.AutoUpdateEnabled {
		status = enabledStyle.Render("enabled")
	}
	fmt.Printf("%s %s\n", updateLabelStyle.Render("Status:"), status)

	// Interval
	hours := conf.AutoUpdateCheckInterval / 3600
	fmt.Printf("%s %d seconds (%d hours)\n", 
		updateLabelStyle.Render("Check interval:"), 
		conf.AutoUpdateCheckInterval,
		hours)

	// Last check
	if conf.LastUpdateCheck > 0 {
		lastCheck := time.Unix(conf.LastUpdateCheck, 0)
		fmt.Printf("%s %s\n", 
			updateLabelStyle.Render("Last check:"), 
			lastCheck.Format("2006-01-02 15:04:05"))

		// Next check
		if conf.AutoUpdateEnabled {
			nextCheck := lastCheck.Add(time.Duration(conf.AutoUpdateCheckInterval) * time.Second)
			if time.Now().Before(nextCheck) {
				timeUntil := time.Until(nextCheck).Round(time.Minute)
				fmt.Printf("%s in %s\n", updateLabelStyle.Render("Next check:"), timeUntil)
			} else {
				fmt.Printf("%s %s\n", updateLabelStyle.Render("Next check:"), "on next mcs command")
			}
		}
	} else {
		fmt.Printf("%s never\n", updateLabelStyle.Render("Last check:"))
		if conf.AutoUpdateEnabled {
			fmt.Printf("%s on next mcs command\n", updateLabelStyle.Render("Next check:"))
		}
	}

	// Last known version
	fmt.Printf("%s %s\n", updateLabelStyle.Render("Last known version:"), conf.LastKnownVersion)

	fmt.Println()
	fmt.Println(dimStyle.Render("Tip: Set MCS_NO_AUTO_UPDATE=1 to temporarily disable checks"))
	fmt.Println()

	return nil
}

func setAutoUpdateEnabled(enabled bool) error {
	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.SetAutoUpdateEnabled(enabled); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}

	fmt.Printf("✅ Auto-update %s\n", status)
	return nil
}

func setAutoUpdateInterval(intervalStr string) error {
	seconds, err := strconv.ParseInt(intervalStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid interval: %s", intervalStr)
	}

	if seconds < 3600 {
		return fmt.Errorf("interval must be at least 3600 seconds (1 hour)")
	}

	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.SetAutoUpdateCheckInterval(seconds); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	hours := seconds / 3600
	fmt.Printf("✅ Auto-update interval set to %d seconds (%d hours)\n", seconds, hours)
	return nil
}