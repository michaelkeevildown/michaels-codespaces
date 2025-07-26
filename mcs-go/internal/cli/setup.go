package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/spf13/cobra"
)

// SetupCommand creates the 'setup' command
func SetupCommand() *cobra.Command {
	var bootstrap bool
	var skipDeps bool
	var skipGitHub bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "🛠️  Setup MCS and configure environment",
		Long: `Complete MCS setup including:
- Installing system dependencies (Docker, Git)
- Configuring GitHub authentication
- Setting up shell integration
- Creating necessary directories`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(bootstrap, skipDeps, skipGitHub)
		},
	}

	cmd.Flags().BoolVar(&bootstrap, "bootstrap", false, "Run in bootstrap mode (called by installer)")
	cmd.Flags().BoolVar(&skipDeps, "skip-deps", false, "Skip dependency installation")
	cmd.Flags().BoolVar(&skipGitHub, "skip-github", false, "Skip GitHub configuration")

	return cmd
}

func runSetup(bootstrap, skipDeps, skipGitHub bool) error {
	// Header
	if bootstrap {
		fmt.Println(headerStyle.Render("MCS Setup"))
		fmt.Println(strings.Repeat("═", 50))
		fmt.Println()
	}

	// Create necessary directories
	if err := createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Check and install dependencies
	if !skipDeps {
		if err := checkAndInstallDependencies(); err != nil {
			return fmt.Errorf("dependency check failed: %w", err)
		}
	}

	// Configure GitHub authentication
	if !skipGitHub {
		if err := configureGitHub(); err != nil {
			// Don't fail setup if GitHub config fails
			fmt.Println(warningStyle.Render("⚠️  GitHub configuration skipped"))
			fmt.Println("You can configure it later with: mcs setup --skip-deps")
		}
	}

	// Setup shell integration
	if err := setupShellIntegration(); err != nil {
		// Non-fatal
		fmt.Println(warningStyle.Render("⚠️  Shell integration setup failed"))
	}

	// Clone/update MCS source for component installers
	if err := setupMCSSource(); err != nil {
		// Non-fatal
		fmt.Println(warningStyle.Render("⚠️  Could not clone MCS source"))
	}

	// Success message
	fmt.Println()
	fmt.Println(successStyle.Render("✅ MCS setup complete!"))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Reload your shell or run: source ~/.bashrc")
	fmt.Println("  2. Verify setup: mcs doctor")
	fmt.Println("  3. Create a codespace: mcs create my-project")
	fmt.Println()

	return nil
}

func createDirectories() error {
	dirs := []string{
		filepath.Join(os.Getenv("HOME"), ".mcs", "bin"),
		filepath.Join(os.Getenv("HOME"), ".mcs", "config"),
		filepath.Join(os.Getenv("HOME"), ".mcs", "cache"),
		filepath.Join(os.Getenv("HOME"), "codespaces"),
		filepath.Join(os.Getenv("HOME"), "codespaces", "auth", "tokens"),
		filepath.Join(os.Getenv("HOME"), "codespaces", "shared"),
		filepath.Join(os.Getenv("HOME"), "codespaces", "backups"),
	}

	fmt.Println(infoStyle.Render("📁 Creating directories..."))
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	fmt.Println(successStyle.Render("✓ Directories created"))
	return nil
}

func checkAndInstallDependencies() error {
	fmt.Println()
	fmt.Println(infoStyle.Render("🔍 Checking dependencies..."))

	// Check Docker
	if !commandExists("docker") {
		fmt.Println(warningStyle.Render("Docker not found"))
		if runtime.GOOS == "linux" {
			fmt.Println("Would you like to install Docker? [Y/n] ")
			if getUserConfirmation() {
				if err := installDockerLinux(); err != nil {
					return fmt.Errorf("failed to install Docker: %w", err)
				}
			}
		} else {
			fmt.Println("Please install Docker Desktop from: https://www.docker.com/products/docker-desktop")
		}
	} else {
		fmt.Println(successStyle.Render("✓ Docker found"))
	}

	// Check Git
	if !commandExists("git") {
		fmt.Println(warningStyle.Render("Git not found"))
		if runtime.GOOS == "linux" {
			fmt.Println("Installing Git...")
			cmd := exec.Command("sudo", "apt-get", "install", "-y", "git")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to install Git: %w", err)
			}
		} else {
			fmt.Println("Please install Git from: https://git-scm.com/downloads")
		}
	} else {
		fmt.Println(successStyle.Render("✓ Git found"))
	}

	return nil
}

func configureGitHub() error {
	fmt.Println()
	fmt.Println(infoStyle.Render("🔐 Configuring GitHub authentication..."))

	tokenFile := filepath.Join(os.Getenv("HOME"), "codespaces", "auth", "tokens", "github.token")

	// Check if token already exists
	if _, err := os.Stat(tokenFile); err == nil {
		// Try to validate existing token
		if token, err := os.ReadFile(tokenFile); err == nil && len(token) > 0 {
			fmt.Println(successStyle.Render("✓ GitHub token already configured"))
			return nil
		}
	}

	// Check environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		fmt.Println("Using GitHub token from environment variable...")
		if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
			return err
		}
		fmt.Println(successStyle.Render("✓ GitHub token saved"))
		return nil
	}

	// Interactive setup
	fmt.Println()
	fmt.Println("To create codespaces, you need a GitHub Personal Access Token.")
	fmt.Println()
	fmt.Println("Steps:")
	fmt.Println("1. Open: " + urlStyle.Render("https://github.com/settings/tokens/new"))
	fmt.Println("2. Set a note: 'MCS - " + getHostname() + "'")
	fmt.Println("3. Select scopes:")
	fmt.Println("   ✓ repo (Full control of private repositories)")
	fmt.Println("   ✓ workflow (Update GitHub Action workflows)")
	fmt.Println("   ✓ write:packages (Upload packages)")
	fmt.Println("4. Click 'Generate token' and copy it")
	fmt.Println()
	fmt.Print("Paste your token (or press Enter to skip): ")

	reader := bufio.NewReader(os.Stdin)
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)

	if token == "" {
		fmt.Println(warningStyle.Render("Skipping GitHub configuration"))
		return nil
	}

	// Validate token format
	if !strings.HasPrefix(token, "ghp_") || len(token) != 40 {
		return fmt.Errorf("invalid token format")
	}

	// Save token
	if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
		return err
	}

	fmt.Println(successStyle.Render("✓ GitHub token saved"))
	return nil
}

func setupShellIntegration() error {
	fmt.Println()
	fmt.Println(infoStyle.Render("🐚 Setting up shell integration..."))

	mcsHome := filepath.Join(os.Getenv("HOME"), ".mcs")
	binPath := filepath.Join(mcsHome, "bin")

	// Detect shell
	shell := os.Getenv("SHELL")
	var rcFile string

	if strings.Contains(shell, "zsh") {
		rcFile = filepath.Join(os.Getenv("HOME"), ".zshrc")
	} else if strings.Contains(shell, "bash") {
		rcFile = filepath.Join(os.Getenv("HOME"), ".bashrc")
	} else {
		fmt.Println(warningStyle.Render("Unknown shell, skipping PATH setup"))
		return nil
	}

	// Check if PATH already contains MCS
	if pathContains(binPath) {
		fmt.Println(successStyle.Render("✓ MCS already in PATH"))
		return nil
	}

	// Add to shell config
	pathLine := fmt.Sprintf("\n# MCS - Michael's Codespaces\nexport PATH=\"%s:$PATH\"\n", binPath)
	
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(pathLine); err != nil {
		return err
	}

	fmt.Println(successStyle.Render("✓ Added MCS to PATH in " + filepath.Base(rcFile)))
	return nil
}

func setupMCSSource() error {
	fmt.Println()
	fmt.Println(infoStyle.Render("📦 Setting up MCS source..."))

	mcsHome := filepath.Join(os.Getenv("HOME"), ".mcs")
	
	// Check if already cloned
	if _, err := os.Stat(filepath.Join(mcsHome, ".git")); err == nil {
		// Update existing
		cmd := exec.Command("git", "pull", "origin", "main")
		cmd.Dir = mcsHome
		if err := cmd.Run(); err != nil {
			return err
		}
		fmt.Println(successStyle.Render("✓ MCS source updated"))
		return nil
	}

	// Clone repository
	repoURL := "https://github.com/michaelkeevildown/michaels-codespaces.git"
	cmd := exec.Command("git", "clone", repoURL, mcsHome)
	
	if err := cmd.Run(); err != nil {
		// Try with token if available
		tokenFile := filepath.Join(os.Getenv("HOME"), "codespaces", "auth", "tokens", "github.token")
		if token, err := os.ReadFile(tokenFile); err == nil && len(token) > 0 {
			authURL := fmt.Sprintf("https://token:%s@github.com/michaelkeevildown/michaels-codespaces.git", string(token))
			cmd = exec.Command("git", "clone", authURL, mcsHome)
			if err := cmd.Run(); err != nil {
				return err
			}
			// Remove token from URL
			cmd = exec.Command("git", "remote", "set-url", "origin", repoURL)
			cmd.Dir = mcsHome
			cmd.Run()
		} else {
			return err
		}
	}

	fmt.Println(successStyle.Render("✓ MCS source cloned"))
	return nil
}

// Helper functions

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func getUserConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes" || response == ""
}

func getHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func pathContains(dir string) bool {
	path := os.Getenv("PATH")
	return strings.Contains(path, dir)
}

func installDockerLinux() error {
	progress := ui.NewProgress()
	progress.Start("Installing Docker")

	// Install prerequisites
	cmd := exec.Command("sudo", "apt-get", "update")
	if err := cmd.Run(); err != nil {
		progress.Fail("Failed to update package list")
		return err
	}

	cmd = exec.Command("sudo", "apt-get", "install", "-y", 
		"ca-certificates", "curl", "gnupg", "lsb-release")
	if err := cmd.Run(); err != nil {
		progress.Fail("Failed to install prerequisites")
		return err
	}

	// Add Docker's GPG key
	progress.Update("Adding Docker GPG key")
	cmds := [][]string{
		{"sudo", "mkdir", "-p", "/etc/apt/keyrings"},
		{"sh", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg"},
	}

	for _, args := range cmds {
		cmd = exec.Command(args[0], args[1:]...)
		if err := cmd.Run(); err != nil {
			progress.Fail("Failed to add Docker GPG key")
			return err
		}
	}

	// Add Docker repository
	progress.Update("Adding Docker repository")
	cmd = exec.Command("sh", "-c", 
		`echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null`)
	if err := cmd.Run(); err != nil {
		progress.Fail("Failed to add Docker repository")
		return err
	}

	// Install Docker
	progress.Update("Installing Docker packages")
	cmd = exec.Command("sudo", "apt-get", "update")
	if err := cmd.Run(); err != nil {
		progress.Fail("Failed to update package list")
		return err
	}

	cmd = exec.Command("sudo", "apt-get", "install", "-y",
		"docker-ce", "docker-ce-cli", "containerd.io", 
		"docker-buildx-plugin", "docker-compose-plugin")
	if err := cmd.Run(); err != nil {
		progress.Fail("Failed to install Docker")
		return err
	}

	// Start Docker
	progress.Update("Starting Docker service")
	cmd = exec.Command("sudo", "systemctl", "start", "docker")
	if err := cmd.Run(); err != nil {
		progress.Fail("Failed to start Docker")
		return err
	}

	cmd = exec.Command("sudo", "systemctl", "enable", "docker")
	cmd.Run() // Don't fail if this doesn't work

	// Add user to docker group
	progress.Update("Adding user to docker group")
	user := os.Getenv("USER")
	cmd = exec.Command("sudo", "usermod", "-aG", "docker", user)
	if err := cmd.Run(); err != nil {
		progress.Fail("Failed to add user to docker group")
		fmt.Println(warningStyle.Render("You may need to log out and back in for docker permissions"))
	}

	progress.Success("Docker installed successfully")
	return nil
}

