package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/michaelkeevildown/mcs/pkg/utils"
)

var (
	dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// CleanupMCSContainers removes all MCS-related containers
func (c *Client) CleanupMCSContainers(ctx context.Context) error {
	// List all containers
	containers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	removedCount := 0
	failedCount := 0

	for _, container := range containers {
		// Check if it's an MCS container
		if isMCSContainer(container) {
			// Stop container if running
			if container.State == "running" {
				if err := c.StopContainer(ctx, container.ID); err != nil {
					fmt.Printf("Warning: Failed to stop container %s: %v\n", container.ID[:12], err)
				}
			}

			// Remove container
			if err := c.cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
				fmt.Printf("Warning: Failed to remove container %s: %v\n", container.ID[:12], err)
				failedCount++
			} else {
				removedCount++
			}
		}
	}

	if failedCount > 0 {
		return fmt.Errorf("removed %d containers, failed to remove %d", removedCount, failedCount)
	}

	return nil
}

// RemoveAllMCSContainers removes all containers with MCS prefix or related images
func (c *Client) RemoveAllMCSContainers(ctx context.Context) error {
	// Create filter for MCS containers
	filterArgs := filters.NewArgs()
	filterArgs.Add("name", "mcs-")
	
	// List containers with filter
	containers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	// Also check for containers using MCS images
	allContainers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err == nil {
		for _, container := range allContainers {
			if strings.Contains(container.Image, "michaelkeevildown/claude-coder") {
				containers = append(containers, container)
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	uniqueContainers := []types.Container{}
	for _, container := range containers {
		if !seen[container.ID] {
			seen[container.ID] = true
			uniqueContainers = append(uniqueContainers, container)
		}
	}

	// Remove containers
	for _, container := range uniqueContainers {
		// Force remove (stops if running)
		err := c.cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		})
		if err != nil {
			fmt.Printf("Warning: Failed to remove container %s: %v\n", container.Names[0], err)
		}
	}

	return nil
}

// UninstallDocker uninstalls Docker from the system
func UninstallDocker(progress *ui.Progress) error {
	platform := utils.GetPlatform()
	
	switch platform.OS {
	case "darwin":
		return uninstallDockerMacOS(progress)
	case "linux":
		return uninstallDockerLinux(progress)
	default:
		return fmt.Errorf("unsupported platform: %s", platform.OS)
	}
}

func uninstallDockerMacOS(progress *ui.Progress) error {
	appPath := "/Applications/Docker.app"
	if _, err := os.Stat(appPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Docker Desktop not installed
		}
		return err
	}

	fmt.Println(dimStyle.Render("  → Stopping Docker Desktop..."))
	// Stop Docker Desktop
	exec.Command("osascript", "-e", "quit app \"Docker\"").Run()
	time.Sleep(2 * time.Second)

	fmt.Println(dimStyle.Render("  → Removing Docker.app..."))
	// Remove application
	if err := os.RemoveAll(appPath); err != nil {
		return fmt.Errorf("failed to remove Docker.app: %w", err)
	}

	// Remove Docker data directories
	homeDir := utils.GetHomeDir()
	dockerDirs := []string{
		filepath.Join(homeDir, ".docker"),
		filepath.Join(homeDir, "Library/Containers/com.docker.docker"),
		filepath.Join(homeDir, "Library/Application Support/Docker Desktop"),
		filepath.Join(homeDir, "Library/Group Containers/group.com.docker"),
	}

	fmt.Println(dimStyle.Render("  → Removing Docker data directories..."))
	for _, dir := range dockerDirs {
		if _, err := os.Stat(dir); err == nil {
			os.RemoveAll(dir)
		}
	}

	return nil
}

func uninstallDockerLinux(progress *ui.Progress) error {
	// Detect package manager
	pm := utils.DetectPackageManager()
	
	// Stop the spinner before running sudo commands
	if progress != nil {
		progress.Stop()
	}

	fmt.Println(dimStyle.Render("  → Detecting package manager..."))
	
	switch pm {
	case utils.APT:
		fmt.Println(dimStyle.Render("  → Removing Docker packages (apt)..."))
		// Remove Docker packages
		cmd := exec.Command("sudo", "apt-get", "remove", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

		fmt.Println(dimStyle.Render("  → Purging Docker configuration..."))
		// Purge Docker packages
		cmd = exec.Command("sudo", "apt-get", "purge", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

		// Clean up
		cmd = exec.Command("sudo", "apt-get", "autoremove", "-y")
		cmd.Run()

	case utils.YUM:
		fmt.Println(dimStyle.Render("  → Removing Docker packages (yum)..."))
		cmd := exec.Command("sudo", "yum", "remove", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

	case utils.DNF:
		fmt.Println(dimStyle.Render("  → Removing Docker packages (dnf)..."))
		cmd := exec.Command("sudo", "dnf", "remove", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

	default:
		return fmt.Errorf("unsupported package manager or package manager not found")
	}

	fmt.Println(dimStyle.Render("  → Removing Docker data directories..."))
	// Remove Docker data
	dockerDataDirs := []string{
		"/var/lib/docker",
		"/var/lib/containerd",
		"/etc/docker",
		"/var/run/docker.sock",
	}

	for _, dir := range dockerDataDirs {
		cmd := exec.Command("sudo", "rm", "-rf", dir)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	// Remove Docker group
	exec.Command("sudo", "groupdel", "docker").Run()

	// Resume the spinner after sudo commands are done
	if progress != nil {
		progress.Resume()
	}

	return nil
}

// isMCSContainer checks if a container is MCS-related
func isMCSContainer(container types.Container) bool {
	// Check by name prefix
	for _, name := range container.Names {
		if strings.HasPrefix(name, "/mcs-") {
			return true
		}
	}
	
	// Check by image
	if strings.Contains(container.Image, "michaelkeevildown/claude-coder") {
		return true
	}
	
	// Check by labels
	if label, ok := container.Labels["mcs.managed"]; ok && label == "true" {
		return true
	}
	
	return false
}

// GetMCSContainerCount returns the number of MCS containers
func (c *Client) GetMCSContainerCount(ctx context.Context) (int, error) {
	containers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return 0, err
	}

	count := 0
	for _, container := range containers {
		if isMCSContainer(container) {
			count++
		}
	}

	return count, nil
}

// CleanupDockerResources cleans up unused Docker resources
func (c *Client) CleanupDockerResources(ctx context.Context) error {
	// Prune containers
	_, err := c.cli.ContainersPrune(ctx, filters.Args{})
	if err != nil {
		return fmt.Errorf("failed to prune containers: %w", err)
	}

	// Prune images
	_, err = c.cli.ImagesPrune(ctx, filters.Args{})
	if err != nil {
		return fmt.Errorf("failed to prune images: %w", err)
	}

	// Prune volumes
	_, err = c.cli.VolumesPrune(ctx, filters.Args{})
	if err != nil {
		return fmt.Errorf("failed to prune volumes: %w", err)
	}

	// Prune networks
	_, err = c.cli.NetworksPrune(ctx, filters.Args{})
	if err != nil {
		return fmt.Errorf("failed to prune networks: %w", err)
	}

	return nil
}