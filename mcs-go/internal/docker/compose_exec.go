package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ComposeExecutor handles docker-compose operations
type ComposeExecutor struct {
	workDir string
}

// NewComposeExecutor creates a new compose executor
func NewComposeExecutor(workDir string) *ComposeExecutor {
	return &ComposeExecutor{
		workDir: workDir,
	}
}

// Up runs docker-compose up
func (ce *ComposeExecutor) Up(ctx context.Context, detached bool) error {
	args := []string{"up"}
	if detached {
		args = append(args, "-d")
	}
	
	return ce.runCompose(ctx, args...)
}

// Down runs docker-compose down
func (ce *ComposeExecutor) Down(ctx context.Context) error {
	return ce.runCompose(ctx, "down")
}

// Start runs docker-compose start
func (ce *ComposeExecutor) Start(ctx context.Context) error {
	return ce.runCompose(ctx, "start")
}

// Stop runs docker-compose stop
func (ce *ComposeExecutor) Stop(ctx context.Context) error {
	return ce.runCompose(ctx, "stop")
}

// Logs runs docker-compose logs
func (ce *ComposeExecutor) Logs(ctx context.Context, follow bool) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	
	return ce.runCompose(ctx, args...)
}

// runCompose executes a docker-compose command
func (ce *ComposeExecutor) runCompose(ctx context.Context, args ...string) error {
	// Check if docker-compose.yml exists
	composePath := filepath.Join(ce.workDir, "docker-compose.yml")
	if _, err := os.Stat(composePath); err != nil {
		return fmt.Errorf("docker-compose.yml not found: %w", err)
	}

	// Try docker compose first (Docker Compose V2)
	cmd := exec.CommandContext(ctx, "docker", append([]string{"compose"}, args...)...)
	cmd.Dir = ce.workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		// If docker compose fails, try docker-compose (V1)
		if strings.Contains(err.Error(), "docker compose is not a docker command") {
			cmd = exec.CommandContext(ctx, "docker-compose", args...)
			cmd.Dir = ce.workDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("docker-compose failed: %w", err)
			}
			return nil
		}
		return fmt.Errorf("docker compose failed: %w", err)
	}
	
	return nil
}

// IsComposeAvailable checks if docker-compose is available
func IsComposeAvailable() (bool, string) {
	// Check for Docker Compose V2
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err == nil {
		return true, "docker compose"
	}
	
	// Check for Docker Compose V1
	cmd = exec.Command("docker-compose", "--version")
	if err := cmd.Run(); err == nil {
		return true, "docker-compose"
	}
	
	return false, ""
}