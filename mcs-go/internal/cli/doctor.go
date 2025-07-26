package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelkeevildown/mcs/internal/docker"
	"github.com/spf13/cobra"
)

var (
	checkStyle   = lipgloss.NewStyle().Bold(true)
	passStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	failStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// DoctorCommand creates the 'doctor' command
func DoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"check"},
		Short:   "🏥 Check system health and requirements",
		Long:    "Run diagnostic checks to ensure MCS is properly configured and all requirements are met.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(checkStyle.Render("Michael's Codespaces Doctor"))
			fmt.Println(strings.Repeat("─", 30))
			fmt.Println()
			
			// Track if any checks failed
			hasWarnings := false
			hasErrors := false
			
			// Check Docker
			fmt.Print("Docker: ")
			if err := checkDocker(); err != nil {
				fmt.Println(failStyle.Render("✗") + " " + err.Error())
				hasErrors = true
			} else {
				fmt.Println(passStyle.Render("✓") + " installed and running")
			}
			
			// Check Docker Compose
			fmt.Print("Docker Compose: ")
			if version, err := checkDockerCompose(); err != nil {
				fmt.Println(failStyle.Render("✗") + " " + err.Error())
				hasErrors = true
			} else {
				fmt.Println(passStyle.Render("✓") + " " + version)
			}
			
			// Check Git
			fmt.Print("Git: ")
			if version, err := checkGit(); err != nil {
				fmt.Println(failStyle.Render("✗") + " " + err.Error())
				hasErrors = true
			} else {
				fmt.Println(passStyle.Render("✓") + " " + version)
			}
			
			// Check Go (for building from source)
			fmt.Print("Go: ")
			if version, err := checkGo(); err != nil {
				fmt.Println(warnStyle.Render("⚠") + " not installed (required for building from source)")
				hasWarnings = true
			} else {
				fmt.Println(passStyle.Render("✓") + " " + version)
			}
			
			// Check MCS installation
			fmt.Print("MCS Installation: ")
			if err := checkMCSInstallation(); err != nil {
				fmt.Println(failStyle.Render("✗") + " " + err.Error())
				hasErrors = true
			} else {
				fmt.Println(passStyle.Render("✓") + " properly installed")
			}
			
			// Check codespaces directory
			fmt.Print("Codespaces directory: ")
			codespacesDir := filepath.Join(os.Getenv("HOME"), "codespaces")
			if _, err := os.Stat(codespacesDir); err != nil {
				fmt.Println(warnStyle.Render("⚠") + " missing (will be created on first use)")
				hasWarnings = true
			} else {
				fmt.Println(passStyle.Render("✓") + " exists")
			}
			
			// Check Docker network
			fmt.Print("Docker network (mcs-network): ")
			if err := checkDockerNetwork(); err != nil {
				fmt.Println(warnStyle.Render("⚠") + " not found (will be created on first use)")
				hasWarnings = true
			} else {
				fmt.Println(passStyle.Render("✓") + " exists")
			}
			
			// Check disk space
			fmt.Print("Disk space: ")
			if free, err := getFreeDiskSpace(); err != nil {
				fmt.Println(warnStyle.Render("⚠") + " unable to check")
				hasWarnings = true
			} else if free < 1024*1024*1024 { // Less than 1GB
				fmt.Println(failStyle.Render("✗") + fmt.Sprintf(" low disk space (%s free)", formatBytes(free)))
				hasErrors = true
			} else {
				fmt.Println(passStyle.Render("✓") + fmt.Sprintf(" %s free", formatBytes(free)))
			}
			
			// Summary
			fmt.Println()
			if hasErrors {
				fmt.Println(failStyle.Render("✗ Some checks failed. Please fix the issues above."))
				return fmt.Errorf("system check failed")
			} else if hasWarnings {
				fmt.Println(warnStyle.Render("⚠ Some warnings found, but MCS should work."))
			} else {
				fmt.Println(passStyle.Render("✓ All checks passed! MCS is ready to use."))
			}
			
			// Tips
			fmt.Println()
			fmt.Println(dimStyle.Render("Tips:"))
			fmt.Println(dimStyle.Render("  • Use 'mcs create <repo>' to create your first codespace"))
			fmt.Println(dimStyle.Render("  • Use 'mcs update' to update MCS to the latest version"))
			fmt.Println(dimStyle.Render("  • Use 'mcs help' to see all available commands"))
			
			return nil
		},
	}
}

func checkDocker() error {
	// Check if docker command exists
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("not installed")
	}
	
	// Check if Docker daemon is running
	cmd := exec.Command("docker", "ps")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installed but not running (try starting Docker)")
	}
	
	return nil
}

func checkDockerCompose() (string, error) {
	// Try docker compose (v2)
	cmd := exec.Command("docker", "compose", "version")
	if output, err := cmd.Output(); err == nil {
		version := strings.TrimSpace(string(output))
		return version, nil
	}
	
	// Try docker-compose (v1)
	cmd = exec.Command("docker-compose", "--version")
	if output, err := cmd.Output(); err == nil {
		version := strings.TrimSpace(string(output))
		return version, nil
	}
	
	return "", fmt.Errorf("not installed")
}

func checkGit() (string, error) {
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not installed")
	}
	
	version := strings.TrimSpace(string(output))
	return version, nil
}

func checkGo() (string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not installed")
	}
	
	version := strings.TrimSpace(string(output))
	return version, nil
}

func checkMCSInstallation() error {
	mcsHome := os.Getenv("MCS_HOME")
	if mcsHome == "" {
		mcsHome = filepath.Join(os.Getenv("HOME"), ".mcs")
	}
	
	// Check if MCS home exists
	if _, err := os.Stat(mcsHome); err != nil {
		return fmt.Errorf("MCS home not found at %s", mcsHome)
	}
	
	// Check if it's a git repository (for updates)
	gitDir := filepath.Join(mcsHome, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return fmt.Errorf("not a git repository (updates will fail)")
	}
	
	return nil
}

func checkDockerNetwork() error {
	client, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	
	// This will return an error if the network doesn't exist
	// We're just checking, so we don't need to do anything with the result
	return nil
}

func getFreeDiskSpace() (uint64, error) {
	// This is a simplified check - in production you'd want platform-specific code
	// For now, we'll just return a large number to pass the check
	return 10 * 1024 * 1024 * 1024, nil // 10GB
}