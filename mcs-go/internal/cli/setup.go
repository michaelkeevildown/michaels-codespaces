package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/michaelkeevildown/mcs/internal/config"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/michaelkeevildown/mcs/pkg/utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

	// Check and install dependencies FIRST (includes Git)
	if !skipDeps {
		if err := checkAndInstallDependencies(); err != nil {
			return fmt.Errorf("dependency check failed: %w", err)
		}
	}

	// Setup shell integration
	if err := setupShellIntegration(); err != nil {
		// Non-fatal
		fmt.Println(warningStyle.Render("⚠️  Shell integration setup failed"))
	}

	// Clone/update MCS source for component installers (needs Git)
	if err := setupMCSSource(); err != nil {
		// Non-fatal
		fmt.Println(warningStyle.Render("⚠️  Could not clone MCS source"))
	}

	// Configure GitHub authentication (after MCS source is available)
	if !skipGitHub {
		if err := configureGitHub(); err != nil {
			// Don't fail setup if GitHub config fails
			fmt.Println(warningStyle.Render("⚠️  GitHub configuration skipped"))
			fmt.Println("You can configure it later with: mcs setup --skip-deps")
		}
	}

	// Configure network access
	if err := configureNetworkAccess(); err != nil {
		// Non-fatal
		fmt.Println(warningStyle.Render("⚠️  Network configuration failed"))
		fmt.Println("You can configure it later with: mcs update-ip")
	}

	// Success message
	fmt.Println()
	fmt.Println(successStyle.Render("✅ MCS setup complete!"))
	fmt.Println()
	
	// Note about PATH being already set by installer
	if bootstrap {
		fmt.Println("MCS is now available in your PATH!")
		fmt.Println()
		fmt.Println("You can now run:")
		fmt.Println("  • mcs doctor    - Verify your setup")
		fmt.Println("  • mcs create    - Create a new codespace")
		fmt.Println("  • mcs --help    - See all available commands")
	} else {
		fmt.Println("Next steps:")
		fmt.Println("  1. Reload your shell or run: source ~/.bashrc")
		fmt.Println("  2. Verify setup: mcs doctor")
		fmt.Println("  3. Create a codespace: mcs create my-project")
	}
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
			if getUserConfirmation("Would you like to install Docker? [Y/n]") {
				fmt.Println() // Add newline before progress starts
				if err := installDockerLinux(); err != nil {
					return fmt.Errorf("failed to install Docker: %w", err)
				}
			}
		} else {
			fmt.Println("Please install Docker Desktop from: https://www.docker.com/products/docker-desktop")
		}
	} else {
		// Docker command exists, but check if it's actually working
		fmt.Println(infoStyle.Render("Docker found, checking if it's working..."))
		cmd := exec.Command("docker", "version")
		if err := cmd.Run(); err != nil {
			fmt.Println(warningStyle.Render("⚠️  Docker is installed but not working properly"))
			fmt.Println("Please ensure Docker daemon is running")
			if runtime.GOOS == "linux" {
				fmt.Println("Try: sudo systemctl start docker")
			}
			// Don't fail setup, just warn
		} else {
			fmt.Println(successStyle.Render("✓ Docker is installed and working"))
		}
	}

	// Check Git
	if !commandExists("git") {
		fmt.Println(warningStyle.Render("⚠️  Git not found"))
		if runtime.GOOS == "linux" {
			fmt.Println(infoStyle.Render("📦 Installing Git..."))
			// First update package list
			updateCmd := exec.Command("sudo", "apt-get", "update", "-qq")
			if err := updateCmd.Run(); err != nil {
				fmt.Println(warningStyle.Render("Failed to update package list"))
			}
			
			// Install Git
			cmd := exec.Command("sudo", "apt-get", "install", "-y", "git")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to install Git: %w", err)
			}
			fmt.Println(successStyle.Render("✓ Git installed successfully"))
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
			// Verify it works
			if username := verifyGitHubToken(string(token)); username != "" {
				fmt.Println(successStyle.Render("✓ GitHub token already configured"))
				fmt.Printf("%s  Authenticated as: %s%s%s\n", infoStyle.Render("ℹ"), boldStyle.Render(""), username, "")
				return nil
			}
		}
	}

	// Check environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		fmt.Println("Using GitHub token from environment variable...")
		token = strings.TrimSpace(token)
		if username := verifyGitHubToken(token); username != "" {
			if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
				return err
			}
			fmt.Println(successStyle.Render("✓ GitHub token saved"))
			fmt.Printf("%s  Authenticated as: %s%s%s\n", infoStyle.Render("ℹ"), boldStyle.Render(""), username, "")
			return nil
		}
	}

	// Interactive setup with box
	fmt.Println()
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println(headerStyle.Render("GitHub Personal Access Token Setup"))
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println()
	fmt.Println("To create codespaces, you need a GitHub token.")
	fmt.Println()
	fmt.Println(boldStyle.Render("Quick Setup:"))
	fmt.Println()
	fmt.Println("1. " + infoStyle.Render("Open this URL:"))
	fmt.Println("   " + urlStyle.Render("https://github.com/settings/tokens/new"))
	fmt.Println()
	fmt.Println("2. " + infoStyle.Render("Configure token:"))
	fmt.Println("   • " + boldStyle.Render("Note:") + " MCS - " + getHostname())
	fmt.Println("   • " + boldStyle.Render("Expiration:") + " 90 days (recommended)")
	fmt.Println()
	fmt.Println("   " + boldStyle.Render("Select scopes - Check these boxes:"))
	fmt.Println("   ✓ " + boldStyle.Render("repo") + " (Full control of private repositories)")
	fmt.Println("   ✓ " + boldStyle.Render("workflow") + " (Update GitHub Action workflows)")
	fmt.Println("   ✓ " + boldStyle.Render("write:packages") + " (Upload packages to GitHub Package Registry)")
	fmt.Println()
	fmt.Println("3. " + infoStyle.Render("Generate & copy token") + " (starts with ghp_)")
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println()

	// Continuous loop until valid token or explicit skip
	for {
		fmt.Println(boldStyle.Render("Ready to paste your token?"))
		fmt.Println("• Make sure you checked: repo, workflow, write:packages")
		fmt.Println("• Token should start with 'ghp_' and be 40 characters long")
		fmt.Println()
		fmt.Println(boldStyle.Render("Paste your GitHub token:"))
		fmt.Println("Tip: The token will be hidden as you type (like a password)")
		fmt.Print("> ")
		os.Stdout.Sync()

		// Read password-style (hidden input)
		token := readHiddenInput()
		
		// Handle empty input
		if token == "" {
			fmt.Println()
			fmt.Println(warningStyle.Render("Token is required for creating codespaces."))
			fmt.Println()
			fmt.Println("If you haven't created your token yet:")
			fmt.Println("1. Open: " + urlStyle.Render("https://github.com/settings/tokens/new"))
			fmt.Println("2. Check the 3 scopes mentioned above")
			fmt.Println("3. Click 'Generate token' and copy it")
			fmt.Println()
			fmt.Print("Do you want to skip token setup for now? [y/N] ")
			os.Stdout.Sync()
			
			if getUserConfirmation("Do you want to skip token setup for now? [y/N]") {
				fmt.Println(infoStyle.Render("Skipping token setup. You'll need to set it before creating codespaces."))
				return nil
			} else {
				fmt.Println("Let's try again...")
				fmt.Println()
				continue
			}
		}

		// Show masked token
		if len(token) >= 10 {
			masked := fmt.Sprintf("%s%s%s", token[:7], strings.Repeat("*", 32), token[len(token)-3:])
			fmt.Println()
			fmt.Println(successStyle.Render("✓") + " Token captured: " + masked)
			fmt.Printf("  Length: %d characters\n", len(token))
		}

		// Validate token format
		if !strings.HasPrefix(token, "ghp_") || len(token) != 40 {
			fmt.Println()
			fmt.Println(errorStyle.Render("Invalid token format. GitHub tokens start with 'ghp_' followed by 36 characters."))
			fmt.Println(infoStyle.Render("Example: ghp_A1b2C3d4E5f6G7h8I9j0K1L2M3N4O5P6Q7R8"))
			fmt.Println()
			continue
		}

		// Save token
		if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
			fmt.Println(errorStyle.Render("Failed to save token: " + err.Error()))
			continue
		}

		fmt.Println(successStyle.Render("✓ GitHub token saved successfully!"))
		
		// Verify token
		fmt.Println(infoStyle.Render("Verifying token with GitHub..."))
		if username := verifyGitHubToken(token); username != "" {
			fmt.Println(successStyle.Render("✓ Token verified - authentication working!"))
			fmt.Printf("%s  Authenticated as: %s%s%s\n", infoStyle.Render("ℹ"), boldStyle.Render(""), username, "")
			return nil
		} else {
			// Token verification failed
			fmt.Println(errorStyle.Render("Token verification failed"))
			os.Remove(tokenFile) // Remove invalid token
			fmt.Println()
			fmt.Print("Retry with a different token? [Y/n] ")
			if !getUserConfirmation("Retry with a different token? [Y/n]") {
				fmt.Println(warningStyle.Render("Continuing without GitHub authentication"))
				return nil
			}
			continue
		}
	}
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
	// Check if git is available first
	if !commandExists("git") {
		fmt.Println()
		fmt.Println(warningStyle.Render("⚠️  Git not available, skipping MCS source setup"))
		fmt.Println("You can set it up later after installing Git")
		return nil
	}

	fmt.Println()
	fmt.Println(infoStyle.Render("📦 Setting up MCS source..."))

	mcsHome := filepath.Join(os.Getenv("HOME"), ".mcs")
	sourceDir := filepath.Join(mcsHome, "source")
	
	// Create source directory if it doesn't exist
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}
	
	// Check if already cloned
	if _, err := os.Stat(filepath.Join(sourceDir, ".git")); err == nil {
		// Update existing
		cmd := exec.Command("git", "pull", "origin", "main")
		cmd.Dir = sourceDir
		if err := cmd.Run(); err != nil {
			return err
		}
		fmt.Println(successStyle.Render("✓ MCS source updated"))
		return nil
	}

	// Clone repository
	repoURL := "https://github.com/michaelkeevildown/michaels-codespaces.git"
	cmd := exec.Command("git", "clone", repoURL, sourceDir)
	
	if err := cmd.Run(); err != nil {
		// Try with token if available
		tokenFile := filepath.Join(os.Getenv("HOME"), "codespaces", "auth", "tokens", "github.token")
		if token, err := os.ReadFile(tokenFile); err == nil && len(token) > 0 {
			authURL := fmt.Sprintf("https://token:%s@github.com/michaelkeevildown/michaels-codespaces.git", string(token))
			cmd = exec.Command("git", "clone", authURL, sourceDir)
			if err := cmd.Run(); err != nil {
				return err
			}
			// Remove token from URL
			cmd = exec.Command("git", "remote", "set-url", "origin", repoURL)
			cmd.Dir = sourceDir
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

func getUserConfirmation(prompt string) bool {
	// Display the prompt first!
	fmt.Print(prompt + " ")
	os.Stdout.Sync() // Ensure prompt is displayed before reading input
	
	var reader *bufio.Reader
	
	// Check if stdin is a terminal
	if !term.IsTerminal(int(syscall.Stdin)) {
		// stdin is not a terminal (e.g., piped input)
		// Try to open /dev/tty directly to read from the actual terminal
		tty, err := os.Open("/dev/tty")
		if err != nil {
			// Can't get user input in non-interactive mode
			// Default to NO for safety (don't auto-install things)
			fmt.Println("n (non-interactive mode)")
			return false
		}
		defer tty.Close()
		reader = bufio.NewReader(tty)
	} else {
		// Normal interactive mode
		reader = bufio.NewReader(os.Stdin)
	}
	
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	
	// Check if user wants to proceed (default is NO for skip prompts)
	var result bool
	if strings.Contains(prompt, "[y/N]") {
		// Default is NO
		result = response == "y" || response == "yes"
	} else {
		// Default is YES for other prompts
		result = response == "y" || response == "yes" || response == ""
	}
	
	// Simply display what was selected on the same line
	// This fixes the cursor positioning issue
	if response == "" {
		if strings.Contains(prompt, "[y/N]") {
			fmt.Println("n") // Show default NO
		} else {
			fmt.Println("y") // Show default YES
		}
	}
	
	return result
}

func getHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func pathContains(dir string) bool {
	path := os.Getenv("PATH")
	return strings.Contains(path, dir)
}

// readHiddenInput reads password-style input (hidden from terminal)
func readHiddenInput() string {
	var fd int
	var ttyFile *os.File
	
	// Check if stdin is a terminal
	if !term.IsTerminal(int(syscall.Stdin)) {
		// stdin is not a terminal (e.g., piped input)
		// Try to open /dev/tty directly to read from the actual terminal
		var err error
		ttyFile, err = os.Open("/dev/tty")
		if err != nil {
			// Can't get user input in non-interactive mode
			fmt.Println()
			fmt.Println(errorStyle.Render("Cannot read token in non-interactive mode"))
			fmt.Println(infoStyle.Render("Please run the installer interactively or set GITHUB_TOKEN environment variable"))
			return ""
		}
		defer ttyFile.Close()
		fd = int(ttyFile.Fd())
	} else {
		// Normal interactive mode
		fd = int(syscall.Stdin)
	}

	// Read password-style from the appropriate file descriptor
	bytePassword, err := term.ReadPassword(fd)
	if err != nil {
		// If password reading fails, try normal reading
		var reader *bufio.Reader
		if ttyFile != nil {
			reader = bufio.NewReader(ttyFile)
		} else {
			reader = bufio.NewReader(os.Stdin)
		}
		input, _ := reader.ReadString('\n')
		return strings.TrimSpace(input)
	}
	
	return strings.TrimSpace(string(bytePassword))
}

// verifyGitHubToken verifies a GitHub token and returns the username if valid
func verifyGitHubToken(token string) string {
	token = strings.TrimSpace(token)
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Create request
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return ""
	}
	
	// Set headers
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	// Check status code
	if resp.StatusCode != 200 {
		return ""
	}
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	
	// Extract username from JSON response
	// Simple extraction without importing encoding/json
	bodyStr := string(body)
	loginStart := strings.Index(bodyStr, `"login":"`)
	if loginStart == -1 {
		return ""
	}
	loginStart += len(`"login":"`)
	loginEnd := strings.Index(bodyStr[loginStart:], `"`)
	if loginEnd == -1 {
		return ""
	}
	
	return bodyStr[loginStart : loginStart+loginEnd]
}

func installDockerLinux() error {
	progress := ui.NewProgress()
	progress.Start("Installing Docker")

	// Create context with timeout for the entire installation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Set non-interactive environment for all apt commands
	env := append(os.Environ(),
		"DEBIAN_FRONTEND=noninteractive",
		"NEEDRESTART_MODE=a",
		"NEEDRESTART_SUSPEND=1",
	)

	// Update package list with progress
	progress.Update("Updating package list...")
	fmt.Println(dimStyle.Render("This may take a moment..."))
	
	cmd := exec.CommandContext(ctx, "sudo", "apt-get", "update", "-qq")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			progress.Fail("Package update timed out")
			return fmt.Errorf("apt-get update timed out after 10 minutes")
		}
		progress.Fail("Failed to update package list")
		return err
	}

	// Install prerequisites
	progress.Update("Installing prerequisites...")
	cmd = exec.CommandContext(ctx, "sudo", "apt-get", "install", "-y", "-qq",
		"--no-install-recommends",
		"ca-certificates", "curl", "gnupg", "lsb-release", "gpg")
	cmd.Env = env
	// Show output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			progress.Fail("Prerequisites installation timed out")
			return fmt.Errorf("prerequisites installation timed out")
		}
		progress.Fail("Failed to install prerequisites")
		return err
	}

	// Add Docker's GPG key
	progress.Update("Adding Docker GPG key")
	
	// Create keyrings directory
	cmd = exec.Command("sudo", "mkdir", "-p", "/etc/apt/keyrings")
	if err := cmd.Run(); err != nil {
		progress.Fail("Failed to create keyrings directory")
		return err
	}

	// Download GPG key with timeout and retry
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if attempt > 1 {
			progress.Update(fmt.Sprintf("Adding Docker GPG key (attempt %d/3)", attempt))
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}

		// Try to download GPG key with timeout
		// First, let's download the key to a temp file to separate curl and gpg operations
		tempFile := "/tmp/docker-gpg-key.asc"
		
		// Download the GPG key
		curlCmd := exec.Command("curl", "-fsSL", "--max-time", "30", "--connect-timeout", "10",
			"-o", tempFile, "https://download.docker.com/linux/ubuntu/gpg")
		
		var curlStderr strings.Builder
		curlCmd.Stderr = &curlStderr
		
		if err := curlCmd.Run(); err != nil {
			lastErr = fmt.Errorf("attempt %d: curl failed: %v (stderr: %s)", attempt, err, curlStderr.String())
			// Clean up temp file if it exists
			os.Remove(tempFile)
			continue
		}
		
		// Check if file was downloaded
		if _, err := os.Stat(tempFile); err != nil {
			lastErr = fmt.Errorf("attempt %d: downloaded file not found: %v", attempt, err)
			continue
		}
		
		// Now process the GPG key
		// Use gpg with explicit flags to avoid hanging
		gpgCmd := exec.Command("sudo", "sh", "-c",
			fmt.Sprintf("gpg --batch --yes --dearmor < %s > /etc/apt/keyrings/docker.gpg", tempFile))
		
		var gpgStderr strings.Builder
		gpgCmd.Stderr = &gpgStderr
		
		if err := gpgCmd.Run(); err != nil {
			// If gpg fails, try using cp directly (key might already be in binary format)
			cpCmd := exec.Command("sudo", "cp", tempFile, "/etc/apt/keyrings/docker.gpg")
			if cpErr := cpCmd.Run(); cpErr != nil {
				lastErr = fmt.Errorf("attempt %d: gpg processing failed: %v (stderr: %s), direct copy also failed: %v", 
					attempt, err, gpgStderr.String(), cpErr)
				os.Remove(tempFile)
				continue
			}
		}
		
		// Clean up temp file
		os.Remove(tempFile)
		
		// Verify the key was created
		if _, err := os.Stat("/etc/apt/keyrings/docker.gpg"); err != nil {
			lastErr = fmt.Errorf("attempt %d: GPG key file not created: %v", attempt, err)
			continue
		}
		// Success
		lastErr = nil
		break
	}

	if lastErr != nil {
		progress.Fail("Failed to add Docker GPG key after 3 attempts")
		fmt.Println(errorStyle.Render("Error details: " + lastErr.Error()))
		fmt.Println(infoStyle.Render("Troubleshooting tips:"))
		fmt.Println("  • Check your internet connection")
		fmt.Println("  • Try running: curl -fsSL https://download.docker.com/linux/ubuntu/gpg")
		fmt.Println("  • Ensure you can access download.docker.com")
		return lastErr
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
	progress.Update("Installing Docker packages...")
	fmt.Println(dimStyle.Render("Note: This may remove old Docker versions if present"))
	
	// Update package list again after adding Docker repo
	cmd = exec.CommandContext(ctx, "sudo", "apt-get", "update", "-qq")
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			progress.Fail("Package update timed out")
			return fmt.Errorf("apt-get update timed out")
		}
		progress.Fail("Failed to update package list")
		return err
	}

	// Install Docker with non-interactive flags
	cmd = exec.CommandContext(ctx, "sudo", "apt-get", "install", "-y",
		"--no-install-recommends",
		"--allow-downgrades",
		"--allow-remove-essential",
		"--allow-change-held-packages",
		"-o", "Dpkg::Options::=--force-confdef",
		"-o", "Dpkg::Options::=--force-confold",
		"docker-ce", "docker-ce-cli", "containerd.io", 
		"docker-buildx-plugin", "docker-compose-plugin")
	cmd.Env = env
	// Show output so we can see what's happening
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			progress.Fail("Docker installation timed out")
			fmt.Println(errorStyle.Render("Installation timed out after 10 minutes"))
			fmt.Println(infoStyle.Render("You may need to run 'sudo apt-get -f install' to fix any issues"))
			return fmt.Errorf("Docker installation timed out")
		}
		progress.Fail("Failed to install Docker")
		return err
	}

	// Start Docker
	progress.Update("Starting Docker service")
	cmd = exec.CommandContext(ctx, "sudo", "systemctl", "start", "docker")
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			progress.Fail("Starting Docker timed out")
			return fmt.Errorf("starting Docker service timed out")
		}
		progress.Fail("Failed to start Docker")
		return err
	}

	cmd = exec.CommandContext(ctx, "sudo", "systemctl", "enable", "docker")
	cmd.Run() // Don't fail if this doesn't work

	// Add user to docker group
	progress.Update("Adding user to docker group")
	user := os.Getenv("USER")
	cmd = exec.CommandContext(ctx, "sudo", "usermod", "-aG", "docker", user)
	if err := cmd.Run(); err != nil {
		progress.Fail("Failed to add user to docker group")
		fmt.Println(warningStyle.Render("You may need to log out and back in for docker permissions"))
	}

	progress.Success("Docker installed successfully")
	return nil
}

func configureNetworkAccess() error {
	fmt.Println()
	fmt.Println(infoStyle.Render("🌐 Configuring network access..."))
	
	// Load config manager
	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Check if already configured
	currentConfig := cfg.Get()
	// Skip only if we've moved away from the defaults
	isDefault := currentConfig.IPMode == "localhost" && currentConfig.HostIP == "localhost"
	if !isDefault {
		fmt.Println(successStyle.Render("✓ Network access already configured"))
		fmt.Printf("%s  Mode: %s, IP: %s\n", infoStyle.Render("ℹ"), currentConfig.IPMode, currentConfig.HostIP)
		return nil
	}
	
	// Get available network addresses
	addresses, err := utils.GetAvailableNetworkAddresses()
	if err != nil {
		fmt.Println(warningStyle.Render("⚠️  Could not detect network addresses"))
		return err
	}
	
	// Show network configuration options
	fmt.Println()
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println(headerStyle.Render("Network Access Configuration"))
	fmt.Println(strings.Repeat("═", 50))
	fmt.Println()
	fmt.Println("How would you like to access your codespaces?")
	fmt.Println()
	
	// Display available options
	validOptions := make(map[string]utils.NetworkInterface)
	optionNum := 1
	
	for _, addr := range addresses {
		if addr.Type == "localhost" || addr.Type == "local" {
			fmt.Printf("%d. %s\n", optionNum, addr.Name)
			validOptions[fmt.Sprintf("%d", optionNum)] = addr
			optionNum++
		}
	}
	
	// Show future options as disabled
	for _, addr := range addresses {
		if addr.Type == "external" || addr.Type == "domain" {
			fmt.Printf("%s. %s\n", dimStyle.Render(fmt.Sprintf("%d", optionNum)), dimStyle.Render(addr.Name))
			optionNum++
		}
	}
	
	fmt.Println()
	fmt.Println(infoStyle.Render("ℹ️  Choose how codespaces will be accessible:"))
	fmt.Println("   • Localhost = Only from this machine")
	fmt.Println("   • Local Network = From any device on your network")
	fmt.Println()
	fmt.Print("Select option (default: 1): ")
	os.Stdout.Sync()
	
	// Read user choice with terminal-aware logic
	choice, ok := readUserInput()
	if !ok {
		// Can't get user input in non-interactive mode
		// Default to localhost for safety
		fmt.Println("1 (non-interactive mode)")
		choice = "1"
	}
	
	// Default to localhost if no choice
	if choice == "" {
		choice = "1"
	}
	
	// Get selected option
	selected, ok := validOptions[choice]
	if !ok {
		fmt.Println(warningStyle.Render("Invalid option, defaulting to localhost"))
		selected = validOptions["1"]
	}
	
	// Configure based on selection
	var ipMode string
	var hostIP string
	
	if selected.Type == "localhost" {
		ipMode = "localhost"
		hostIP = "127.0.0.1"
	} else if selected.Type == "local" {
		ipMode = "specific"
		hostIP = selected.IP
	}
	
	// Save configuration
	if err := cfg.SetIPMode(ipMode); err != nil {
		return err
	}
	if err := cfg.SetHostIP(hostIP); err != nil {
		return err
	}
	
	fmt.Println()
	fmt.Println(successStyle.Render("✓ Network access configured"))
	fmt.Printf("%s  Codespaces will be accessible at: %s%s%s\n", 
		infoStyle.Render("ℹ"), 
		boldStyle.Render("http://"), 
		boldStyle.Render(hostIP), 
		boldStyle.Render(":PORT"))
	
	if selected.Type == "local" {
		fmt.Println()
		fmt.Println(warningStyle.Render("⚠️  Security Note:"))
		fmt.Println("   Codespaces will be accessible from other devices on your network.")
		fmt.Println("   Make sure you trust all devices on your network.")
	}
	
	return nil
}

