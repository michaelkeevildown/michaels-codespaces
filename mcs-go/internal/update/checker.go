package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelkeevildown/mcs/internal/config"
)

var (
	updateAvailableStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
	updateInfoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
}

// CheckForUpdates checks if an update is available
func CheckForUpdates(currentVersion string) error {
	// Skip if MCS_NO_AUTO_UPDATE is set
	if os.Getenv("MCS_NO_AUTO_UPDATE") == "1" {
		return nil
	}

	// Load config
	cfg, err := config.NewManager()
	if err != nil {
		return nil // Silently fail
	}

	// Check if auto-update is enabled and if it's time to check
	if !cfg.ShouldCheckForUpdate() {
		return nil
	}

	// Update last check time
	cfg.SetLastUpdateCheck(time.Now().Unix())

	// Fetch latest release from GitHub
	latestVersion, err := fetchLatestVersion()
	if err != nil {
		return nil // Silently fail
	}

	// Store the latest known version
	cfg.SetLastKnownVersion(latestVersion)

	// Compare versions
	if isNewerVersion(currentVersion, latestVersion) {
		showUpdateNotification(currentVersion, latestVersion)
	}

	return nil
}

// fetchLatestVersion fetches the latest version from GitHub
func fetchLatestVersion() (string, error) {
	// GitHub API URL for latest release
	url := "https://api.github.com/repos/michaelkeevildown/mcs/releases/latest"

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch release: %s", resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	// Remove 'v' prefix if present
	version := strings.TrimPrefix(release.TagName, "v")
	return version, nil
}

// isNewerVersion compares two semantic versions
func isNewerVersion(current, latest string) bool {
	// Simple comparison - in production, use a proper semver library
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Skip if current is "dev"
	if current == "dev" {
		return false
	}

	// Parse versions
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// Compare major.minor.patch
	for i := 0; i < 3 && i < len(currentParts) && i < len(latestParts); i++ {
		var currentNum, latestNum int
		fmt.Sscanf(currentParts[i], "%d", &currentNum)
		fmt.Sscanf(latestParts[i], "%d", &latestNum)

		if latestNum > currentNum {
			return true
		} else if latestNum < currentNum {
			return false
		}
	}

	// If all parts are equal, check if latest has more parts
	return len(latestParts) > len(currentParts)
}

// showUpdateNotification displays an update notification
func showUpdateNotification(current, latest string) {
	fmt.Println()
	fmt.Println(updateAvailableStyle.Render("🚀 Update available!"))
	fmt.Printf("%s %s → %s\n", 
		updateInfoStyle.Render("Version:"),
		current,
		updateAvailableStyle.Render(latest))
	fmt.Println()
	fmt.Println("Update with: mcs update")
	fmt.Println("Disable auto-update checks with: mcs autoupdate off")
	fmt.Println()
}

// PerformUpdate performs the actual update
func PerformUpdate(checkOnly bool) error {
	mcsHome := os.Getenv("MCS_HOME")
	if mcsHome == "" {
		mcsHome = filepath.Join(os.Getenv("HOME"), ".mcs")
	}

	// Check if it's a source installation
	gitDir := filepath.Join(mcsHome, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Source installation - use git pull
		return updateFromSource(mcsHome, checkOnly)
	} else {
		// Binary installation
		return updateBinary(checkOnly)
	}
}

// updateFromSource updates a source installation
func updateFromSource(mcsHome string, checkOnly bool) error {
	// Fetch latest changes
	cmd := exec.Command("git", "fetch", "origin", "main")
	cmd.Dir = mcsHome
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch updates: %w", err)
	}

	// Check for updates
	cmd = exec.Command("git", "rev-list", "HEAD...origin/main", "--count")
	cmd.Dir = mcsHome
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	commitsBehind := strings.TrimSpace(string(output))
	if commitsBehind == "0" {
		fmt.Println("✅ MCS is up to date!")
		return nil
	}

	if checkOnly {
		fmt.Printf("📦 Update available: %s commits behind\n", commitsBehind)
		fmt.Println("Run 'mcs update' to update")
		return nil
	}

	// Pull latest changes
	fmt.Println("🔄 Updating MCS from source...")
	cmd = exec.Command("git", "pull", "origin", "main")
	cmd.Dir = mcsHome
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull updates: %w", err)
	}

	// Rebuild
	fmt.Println("🔨 Rebuilding MCS...")
	cmd = exec.Command("go", "build", "-o", "mcs", "./cmd/mcs")
	cmd.Dir = mcsHome
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to rebuild: %w", err)
	}

	// Install to PATH
	fmt.Println("📦 Installing updated binary...")
	srcBinary := filepath.Join(mcsHome, "mcs")
	destBinary := "/usr/local/bin/mcs"
	
	// Try without sudo first
	if err := copyFile(srcBinary, destBinary); err != nil {
		// Try with sudo
		cmd = exec.Command("sudo", "cp", srcBinary, destBinary)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install binary: %w", err)
		}
	}

	fmt.Println("✅ MCS has been updated successfully!")
	return nil
}

// updateBinary updates a binary installation
func updateBinary(checkOnly bool) error {
	// This would download the latest binary from GitHub releases
	// For now, we'll just inform the user
	if checkOnly {
		fmt.Println("📦 Checking for updates...")
		// Would check GitHub releases API
		fmt.Println("Binary update check not yet implemented")
		return nil
	}

	fmt.Println("Binary updates not yet implemented")
	fmt.Println("Please reinstall MCS using the install script:")
	fmt.Println("  curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/mcs/main/mcs-go/install.sh | bash")
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, 0755)
	if err != nil {
		return err
	}

	return nil
}