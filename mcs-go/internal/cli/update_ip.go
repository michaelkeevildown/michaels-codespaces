package cli

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelkeevildown/mcs/internal/config"
	"github.com/spf13/cobra"
)

var (
	ipHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	ipInfoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	ipValueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

// UpdateIPCommand creates the 'update-ip' command
func UpdateIPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-ip [mode]",
		Aliases: []string{"ip"},
		Short:   "🌐 Configure network access mode",
		Long: `Configure how to access your codespaces (localhost, auto-detect, public IP, or custom).
		
Available modes:
  localhost  - Use localhost (127.0.0.1)
  auto       - Auto-detect local network IP
  public     - Use public IP address
  custom     - Set a specific IP address
  
You can also use flags for direct configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle flags first
			if showFlag, _ := cmd.Flags().GetBool("show"); showFlag {
				return showCurrentConfig()
			}

			if autoFlag, _ := cmd.Flags().GetBool("auto"); autoFlag {
				return updateIPMode("auto", "")
			}

			if publicFlag, _ := cmd.Flags().GetBool("public"); publicFlag {
				return updateIPMode("public", "")
			}

			if localhostFlag, _ := cmd.Flags().GetBool("localhost"); localhostFlag {
				return updateIPMode("localhost", "")
			}

			if ipFlag, _ := cmd.Flags().GetString("ip"); ipFlag != "" {
				return updateIPMode("custom", ipFlag)
			}

			// Handle positional argument
			if len(args) > 0 {
				mode := args[0]
				switch mode {
				case "localhost", "auto", "public":
					return updateIPMode(mode, "")
				case "show":
					return showCurrentConfig()
				default:
					// Assume it's an IP address
					if net.ParseIP(mode) != nil {
						return updateIPMode("custom", mode)
					}
					return fmt.Errorf("invalid mode or IP address: %s", mode)
				}
			}

			// No arguments, show interactive menu
			return interactiveIPConfig()
		},
	}

	// Add flags
	cmd.Flags().Bool("show", false, "Show current IP configuration")
	cmd.Flags().BoolP("auto", "a", false, "Auto-detect local IP")
	cmd.Flags().BoolP("public", "p", false, "Use public IP")
	cmd.Flags().BoolP("localhost", "l", false, "Use localhost")
	cmd.Flags().StringP("ip", "i", "", "Set specific IP address")

	return cmd
}

func showCurrentConfig() error {
	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println()
	fmt.Println(ipHeaderStyle.Render("Current Network Configuration"))
	fmt.Println(strings.Repeat("─", 29))
	
	conf := cfg.Get()
	fmt.Printf("%s %s\n", ipInfoStyle.Render("Mode:"), ipValueStyle.Render(conf.IPMode))
	fmt.Printf("%s %s\n", ipInfoStyle.Render("Current IP:"), ipValueStyle.Render(conf.HostIP))
	
	// Show detected IPs
	fmt.Println()
	fmt.Println(ipHeaderStyle.Render("Available IPs"))
	fmt.Println(strings.Repeat("─", 13))
	
	// Local IP
	localIP := getLocalIP()
	fmt.Printf("%s %s\n", ipInfoStyle.Render("Local IP:"), localIP)
	
	// Public IP (would need external service)
	fmt.Printf("%s %s\n", ipInfoStyle.Render("Public IP:"), dimStyle.Render("(requires external service)"))
	
	fmt.Println()
	return nil
}

func updateIPMode(mode, customIP string) error {
	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var newIP string
	
	switch mode {
	case "localhost":
		newIP = "127.0.0.1"
	case "auto":
		newIP = getLocalIP()
		if newIP == "" {
			return fmt.Errorf("failed to detect local IP")
		}
	case "public":
		// This would need an external service
		return fmt.Errorf("public IP detection not yet implemented")
	case "custom":
		if customIP == "" {
			return fmt.Errorf("custom IP address required")
		}
		if net.ParseIP(customIP) == nil {
			return fmt.Errorf("invalid IP address: %s", customIP)
		}
		newIP = customIP
	default:
		return fmt.Errorf("invalid mode: %s", mode)
	}

	// Update configuration
	if err := cfg.SetIPMode(mode); err != nil {
		return err
	}
	if err := cfg.SetHostIP(newIP); err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("✅ Network configuration updated\n")
	fmt.Printf("   Mode: %s\n", ipValueStyle.Render(mode))
	fmt.Printf("   IP: %s\n", ipValueStyle.Render(newIP))
	
	// TODO: Update all existing codespace URLs
	fmt.Println()
	fmt.Println(dimStyle.Render("Note: Existing codespace URLs will be updated on next start"))
	
	return nil
}

func interactiveIPConfig() error {
	fmt.Println()
	fmt.Println(ipHeaderStyle.Render("Configure Network Access"))
	fmt.Println(strings.Repeat("─", 24))
	fmt.Println()
	fmt.Println("How would you like to access your codespaces?")
	fmt.Println()
	fmt.Println("1) Localhost only (127.0.0.1)")
	fmt.Println("2) Auto-detect local network IP")
	fmt.Println("3) Use public IP address")
	fmt.Println("4) Set custom IP address")
	fmt.Println("5) Show current configuration")
	fmt.Println()
	fmt.Print("Select option [1-5]: ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return updateIPMode("localhost", "")
	case "2":
		return updateIPMode("auto", "")
	case "3":
		return updateIPMode("public", "")
	case "4":
		fmt.Print("Enter IP address: ")
		ip, _ := reader.ReadString('\n')
		ip = strings.TrimSpace(ip)
		return updateIPMode("custom", ip)
	case "5":
		return showCurrentConfig()
	default:
		return fmt.Errorf("invalid option: %s", choice)
	}
}

func getLocalIP() string {
	// Get all network interfaces
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	// Find the first non-loopback IPv4 address
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}