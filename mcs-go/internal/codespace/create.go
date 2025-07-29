package codespace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/michaelkeevildown/mcs/internal/assets"
	"github.com/michaelkeevildown/mcs/internal/assets/dockerfiles"
	"github.com/michaelkeevildown/mcs/internal/components"
	"github.com/michaelkeevildown/mcs/internal/config"
	"github.com/michaelkeevildown/mcs/internal/docker"
	"github.com/michaelkeevildown/mcs/internal/git"
	"github.com/michaelkeevildown/mcs/internal/ports"
)

// Create creates a new codespace
func (m *Manager) Create(ctx context.Context, opts CreateOptions) (*Codespace, error) {
	// Helper function for progress reporting
	reportProgress := func(msg string) {
		if opts.Progress != nil {
			opts.Progress(msg)
		}
	}

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// Check if codespace already exists
	codespaceDir := opts.GetPath()
	if _, err := os.Stat(codespaceDir); err == nil {
		return nil, fmt.Errorf("codespace already exists: %s", opts.Name)
	}

	// Create directory structure
	reportProgress("Creating directory structure")
	if err := createDirectoryStructure(codespaceDir, len(opts.Components) > 0); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Clone repository
	reportProgress("Cloning repository")
	if err := cloneRepository(ctx, opts.Repository.URL, filepath.Join(codespaceDir, "src"), opts.CloneDepth); err != nil {
		// Clean up on failure
		os.RemoveAll(codespaceDir)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Detect language/framework
	reportProgress("Detecting project type")
	language := detectLanguage(filepath.Join(codespaceDir, "src"))

	// Allocate ports
	reportProgress("Allocating ports")
	portRegistry, err := ports.NewPortRegistry()
	if err != nil {
		os.RemoveAll(codespaceDir)
		return nil, fmt.Errorf("failed to create port registry: %w", err)
	}

	allocatedPorts, err := portRegistry.AllocateCodespacePorts(opts.Name)
	if err != nil {
		os.RemoveAll(codespaceDir)
		return nil, fmt.Errorf("failed to allocate ports: %w", err)
	}

	// Generate password
	password := generatePassword()

	// Prepare Docker configuration
	reportProgress("Generating Docker configuration")
	
	// Get image info based on language and components
	imageInfo := docker.GetImageInfo(language, opts.Components)
	
	// Check if dockerfiles are available
	dockerfilesPath := config.GetDockerfilesPath()
	var buildContext string
	var dockerfile string
	var finalImage string
	
	// Try to extract dockerfiles if not present
	if _, err := os.Stat(dockerfilesPath); os.IsNotExist(err) {
		// Try to extract embedded dockerfiles
		if err := extractEmbeddedDockerfiles(dockerfilesPath); err == nil {
			buildContext = dockerfilesPath
			dockerfile = imageInfo.Dockerfile
			finalImage = imageInfo.Image
		} else {
			// Extraction failed, use fallback image without build context
			finalImage = imageInfo.FallbackImage
			reportProgress("Using fallback image (dockerfiles not available)")
		}
	} else {
		// Dockerfiles directory exists
		buildContext = dockerfilesPath
		dockerfile = imageInfo.Dockerfile
		finalImage = imageInfo.Image
	}
	
	dockerConfig := docker.ComposeConfig{
		ContainerName: fmt.Sprintf("%s-dev", opts.Name),
		CodespaceName: opts.Name,
		Image:         finalImage,
		BuildContext:  buildContext,
		Dockerfile:    dockerfile,
		Password:      password,
		Ports: map[string]string{
			fmt.Sprintf("%d", allocatedPorts["vscode"]): "8080",
			fmt.Sprintf("%d", allocatedPorts["app"]):    "3000",
		},
		Environment: map[string]string{
			"CODESPACE_NAME": opts.Name,
			"REPO_URL":       opts.Repository.URL,
		},
		Labels: map[string]string{
			"codespace.repo":     opts.Repository.Name,
			"codespace.created":  time.Now().Format(time.RFC3339),
			"codespace.language": language,
		},
		Components: opts.Components,
		WorkingDir: codespaceDir,
	}

	// Generate docker-compose.yml
	composeContent, err := docker.GenerateDockerCompose(dockerConfig)
	if err != nil {
		portRegistry.ReleaseCodespacePorts(opts.Name)
		os.RemoveAll(codespaceDir)
		return nil, fmt.Errorf("failed to generate docker-compose: %w", err)
	}

	if err := os.WriteFile(filepath.Join(codespaceDir, "docker-compose.yml"), composeContent, 0644); err != nil {
		portRegistry.ReleaseCodespacePorts(opts.Name)
		os.RemoveAll(codespaceDir)
		return nil, fmt.Errorf("failed to write docker-compose.yml: %w", err)
	}

	// Generate .env file
	envContent := docker.GenerateEnvFile(dockerConfig)
	if err := os.WriteFile(filepath.Join(codespaceDir, ".env"), envContent, 0644); err != nil {
		portRegistry.ReleaseCodespacePorts(opts.Name)
		os.RemoveAll(codespaceDir)
		return nil, fmt.Errorf("failed to write .env file: %w", err)
	}

	// Setup components if any
	if len(opts.Components) > 0 {
		reportProgress("Setting up components")
		if err := setupComponents(codespaceDir, opts.Components); err != nil {
			portRegistry.ReleaseCodespacePorts(opts.Name)
			os.RemoveAll(codespaceDir)
			return nil, fmt.Errorf("failed to setup components: %w", err)
		}
	}

	// Create Docker network
	reportProgress("Creating Docker network")
	dockerClient, err := docker.NewClient()
	if err != nil {
		portRegistry.ReleaseCodespacePorts(opts.Name)
		os.RemoveAll(codespaceDir)
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	if err := dockerClient.CreateDockerNetwork(ctx); err != nil {
		portRegistry.ReleaseCodespacePorts(opts.Name)
		os.RemoveAll(codespaceDir)
		return nil, fmt.Errorf("failed to create Docker network: %w", err)
	}

	// Get configured host IP
	cfg, err := config.NewManager()
	if err != nil {
		portRegistry.ReleaseCodespacePorts(opts.Name)
		os.RemoveAll(codespaceDir)
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	hostIP := cfg.GetHostIP()
	if hostIP == "" {
		hostIP = "localhost"
	}

	// Calculate dockerfile checksum
	var dockerfileChecksum string
	if dockerfile != "" {
		dockerfileChecksum = dockerfiles.GetDockerfileChecksum(dockerfile)
	}

	// Create codespace object
	cs := &Codespace{
		Name:               opts.Name,
		Repository:         opts.Repository.URL,
		Path:               codespaceDir,
		Status:             "created",
		CreatedAt:          time.Now(),
		VSCodeURL:          fmt.Sprintf("http://%s:%d", hostIP, allocatedPorts["vscode"]),
		AppURL:             fmt.Sprintf("http://%s:%d", hostIP, allocatedPorts["app"]),
		Components:         components.GetSelectedIDs(),
		Language:           language,
		Password:           password,
		DockerfileChecksum: dockerfileChecksum,
	}

	// Save metadata
	reportProgress("Saving metadata")
	if err := m.SaveMetadata(cs); err != nil {
		portRegistry.ReleaseCodespacePorts(opts.Name)
		os.RemoveAll(codespaceDir)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	// Start container if requested
	if !opts.NoStart {
		reportProgress("Starting services")
		if err := m.Start(ctx, opts.Name); err != nil {
			// Don't fail creation if start fails
			fmt.Printf("Warning: Failed to start codespace: %v\n", err)
		} else {
			cs.Status = "running"
		}
	}

	return cs, nil
}

// createDirectoryStructure creates the codespace directory structure
func createDirectoryStructure(basePath string, hasComponents bool) error {
	dirs := []string{
		basePath,
		filepath.Join(basePath, "src"),
		filepath.Join(basePath, "data"),
		filepath.Join(basePath, "config"),
		filepath.Join(basePath, "logs"),
	}

	if hasComponents {
		dirs = append(dirs, 
			filepath.Join(basePath, "components"),
			filepath.Join(basePath, "init"),
		)
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// cloneRepository clones the repository with progress tracking
func cloneRepository(ctx context.Context, url, path string, depth int) error {
	cloneOpts := git.CloneOptions{
		URL:   url,
		Path:  path,
		Progress: func(msg string) {
			// Progress is now handled by the main progress tracker
		},
	}
	
	// Handle depth:
	// - 0 or negative: full clone (no depth limit)
	// - positive: shallow clone with specified depth
	// - not specified in CLI (0): defaults to 20
	if depth > 0 {
		cloneOpts.Depth = depth
	} else if depth == 0 {
		// Default to 20 commits if not specified
		cloneOpts.Depth = 20
	}
	// If depth < 0, don't set Depth field (full clone)

	err := git.Clone(ctx, cloneOpts)
	if err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	
	// Verify clone succeeded
	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		return fmt.Errorf("clone verification failed - .git directory not found: %w", err)
	}
	
	return nil
}

// detectLanguage attempts to detect the primary language of the project
func detectLanguage(projectPath string) string {
	// Check for language-specific files
	checks := map[string][]string{
		"python":     {"requirements.txt", "setup.py", "Pipfile", "pyproject.toml"},
		"node":       {"package.json", "yarn.lock", "package-lock.json"},
		"go":         {"go.mod", "go.sum"},
		"rust":       {"Cargo.toml", "Cargo.lock"},
		"java":       {"pom.xml", "build.gradle", "build.gradle.kts"},
		"php":        {"composer.json", "composer.lock"},
		"ruby":       {"Gemfile", "Gemfile.lock"},
		"dotnet":     {"*.csproj", "*.fsproj", "*.vbproj"},
	}

	// First, check root directory
	for language, files := range checks {
		for _, file := range files {
			// Handle glob patterns
			if strings.Contains(file, "*") {
				matches, _ := filepath.Glob(filepath.Join(projectPath, file))
				if len(matches) > 0 {
					return language
				}
			} else {
				if _, err := os.Stat(filepath.Join(projectPath, file)); err == nil {
					return language
				}
			}
		}
	}

	// If no language detected at root, check subdirectories
	// Common subdirectory patterns to check
	subdirPatterns := []string{
		"*",           // Check all immediate subdirectories
		"backend",     // Common backend directory
		"api",         // API directory
		"server",      // Server directory
		"app",         // Application directory
		"src",         // Source directory
		"*-go",        // Go-specific patterns like mcs-go
		"*-api",       // API subdirectories
		"*-backend",   // Backend subdirectories
		"services/*",  // Microservices pattern
		"packages/*",  // Monorepo pattern
	}

	for _, pattern := range subdirPatterns {
		var dirsToCheck []string
		
		if strings.Contains(pattern, "*") {
			// Handle glob patterns
			matches, _ := filepath.Glob(filepath.Join(projectPath, pattern))
			dirsToCheck = append(dirsToCheck, matches...)
		} else {
			// Direct directory path
			dirsToCheck = append(dirsToCheck, filepath.Join(projectPath, pattern))
		}

		for _, dir := range dirsToCheck {
			// Skip if not a directory
			info, err := os.Stat(dir)
			if err != nil || !info.IsDir() {
				continue
			}

			// Check for language files in this subdirectory
			for language, files := range checks {
				for _, file := range files {
					if strings.Contains(file, "*") {
						matches, _ := filepath.Glob(filepath.Join(dir, file))
						if len(matches) > 0 {
							return language
						}
					} else {
						if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
							return language
						}
					}
				}
			}
		}
	}

	return "generic"
}

// generatePassword generates a secure random password
func generatePassword() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:16]
}

// setupComponents prepares component installation
func setupComponents(codespaceDir string, selectedComponents []components.Component) error {
	componentsDir := filepath.Join(codespaceDir, "components")
	initDir := filepath.Join(codespaceDir, "init")

	// Extract embedded installer scripts
	if err := assets.ExtractInstallers(componentsDir); err != nil {
		return fmt.Errorf("failed to extract installers: %w", err)
	}

	// Update component Installer field to match extracted files
	for i := range selectedComponents {
		selectedComponents[i].Installer = fmt.Sprintf("%s.sh", selectedComponents[i].ID)
	}

	// Generate init script
	initScript, err := docker.GenerateInitScript(selectedComponents)
	if err != nil {
		return fmt.Errorf("failed to generate init script: %w", err)
	}

	initScriptPath := filepath.Join(initDir, "init.sh")
	if err := os.WriteFile(initScriptPath, initScript, 0755); err != nil {
		return fmt.Errorf("failed to write init script: %w", err)
	}

	return nil
}

// Validate validates create options
func (o CreateOptions) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("name is required")
	}
	if o.Repository == nil {
		return fmt.Errorf("repository is required")
	}
	return nil
}

// extractEmbeddedDockerfiles extracts embedded dockerfiles to the target directory
func extractEmbeddedDockerfiles(targetDir string) error {
	return dockerfiles.ExtractDockerfiles(targetDir)
}