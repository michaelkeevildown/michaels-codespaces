package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// CloneOptions holds options for cloning a repository
type CloneOptions struct {
	URL      string
	Path     string
	Branch   string
	Depth    int
	Progress func(string)
	Auth     transport.AuthMethod
}

// ProgressWriter implements io.Writer for progress updates
type ProgressWriter struct {
	callback func(string)
	lastLine string
}

func (w *ProgressWriter) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != w.lastLine {
			w.lastLine = line
			if w.callback != nil {
				// Parse git progress output
				if strings.Contains(line, "Counting objects:") ||
					strings.Contains(line, "Compressing objects:") ||
					strings.Contains(line, "Receiving objects:") ||
					strings.Contains(line, "Resolving deltas:") {
					w.callback(line)
				}
			}
		}
	}
	return len(p), nil
}

// Clone clones a repository with progress tracking
func Clone(ctx context.Context, opts CloneOptions) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(opts.Path), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Determine the URL to use for cloning
	cloneURL := opts.URL
	
	// If we have a GitHub token and this is a GitHub SSH URL, convert to HTTPS
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		if strings.HasPrefix(opts.URL, "git@github.com:") {
			// Convert git@github.com:user/repo.git to https://github.com/user/repo.git
			cloneURL = strings.Replace(opts.URL, "git@github.com:", "https://github.com/", 1)
			fmt.Printf("Using GitHub token authentication (converted SSH to HTTPS)\n")
		}
	}

	// Setup clone options
	cloneOpts := &git.CloneOptions{
		URL:      cloneURL,
		Progress: &ProgressWriter{callback: opts.Progress},
	}

	// Set branch if specified
	if opts.Branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(opts.Branch)
	}

	// Set depth for shallow clone
	if opts.Depth > 0 {
		cloneOpts.Depth = opts.Depth
	}

	// Set authentication
	if opts.Auth != nil {
		cloneOpts.Auth = opts.Auth
	} else {
		// Try to auto-detect auth method
		cloneOpts.Auth = detectAuthMethod(cloneURL)
	}

	// Clone the repository
	_, err := git.PlainCloneContext(ctx, opts.Path, false, cloneOpts)
	if err != nil {
		// Clean up on failure
		os.RemoveAll(opts.Path)
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

// detectAuthMethod attempts to detect the appropriate auth method
func detectAuthMethod(url string) transport.AuthMethod {
	// First, check if we have a GitHub token - this is preferred over SSH
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		// Check if this is a GitHub URL (SSH or HTTPS)
		if strings.Contains(url, "github.com") {
			return &http.BasicAuth{
				Username: "token",
				Password: token,
			}
		}
	}

	// GitLab token
	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		if strings.Contains(url, "gitlab") {
			return &http.BasicAuth{
				Username: "oauth2",
				Password: token,
			}
		}
	}

	// Fall back to SSH for git@ URLs if no token is available
	if strings.HasPrefix(url, "git@") || strings.Contains(url, "ssh://") {
		// Try to use SSH key from default location
		homeDir, _ := os.UserHomeDir()
		sshKeyPath := filepath.Join(homeDir, ".ssh", "id_rsa")
		
		// Check for id_ed25519 if id_rsa doesn't exist
		if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
			sshKeyPath = filepath.Join(homeDir, ".ssh", "id_ed25519")
		}

		if _, err := os.Stat(sshKeyPath); err == nil {
			auth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
			if err == nil {
				return auth
			}
			// Log the error for debugging
			fmt.Printf("Warning: Failed to load SSH key from %s: %v\n", sshKeyPath, err)
		}
	}

	return nil
}

// ValidateRepository checks if a URL points to a valid repository
func ValidateRepository(ctx context.Context, url string) error {
	// Create a temporary directory for validation
	tempDir, err := os.MkdirTemp("", "mcs-validate-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Convert SSH URL to HTTPS if we have a GitHub token
	validateURL := url
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		if strings.HasPrefix(url, "git@github.com:") {
			validateURL = strings.Replace(url, "git@github.com:", "https://github.com/", 1)
		}
	}

	// Try to list references without cloning
	remote := git.NewRemote(nil, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{validateURL},
	})

	auth := detectAuthMethod(validateURL)
	_, err = remote.ListContext(ctx, &git.ListOptions{
		Auth: auth,
	})

	if err != nil {
		return fmt.Errorf("repository validation failed: %w", err)
	}

	return nil
}

// GetDefaultBranch determines the default branch of a repository
func GetDefaultBranch(ctx context.Context, url string) (string, error) {
	// Convert SSH URL to HTTPS if we have a GitHub token
	checkURL := url
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		if strings.HasPrefix(url, "git@github.com:") {
			checkURL = strings.Replace(url, "git@github.com:", "https://github.com/", 1)
		}
	}

	remote := git.NewRemote(nil, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{checkURL},
	})

	auth := detectAuthMethod(checkURL)
	refs, err := remote.ListContext(ctx, &git.ListOptions{
		Auth: auth,
	})

	if err != nil {
		return "", fmt.Errorf("failed to list references: %w", err)
	}

	// Look for HEAD reference
	for _, ref := range refs {
		if ref.Name().String() == "HEAD" {
			target := ref.Target().Short()
			return strings.TrimPrefix(target, "refs/heads/"), nil
		}
	}

	// Fallback to common defaults
	for _, ref := range refs {
		name := ref.Name().String()
		if name == "refs/heads/main" || name == "refs/heads/master" {
			return strings.TrimPrefix(name, "refs/heads/"), nil
		}
	}

	return "main", nil
}