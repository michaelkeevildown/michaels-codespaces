package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Client wraps the Docker client
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Test connection
	ctx := context.Background()
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("Docker daemon not accessible: %w", err)
	}

	return &Client{cli: cli}, nil
}

// Close closes the Docker client
func (c *Client) Close() error {
	return c.cli.Close()
}

// ContainerStatus represents container status
type ContainerStatus struct {
	ID      string
	Name    string
	Status  string
	State   string
	Ports   []string
	Created int64
	Image   string
	Stats   *ContainerStats
}

// ContainerStats represents container resource usage statistics
type ContainerStats struct {
	CPUPercent    float64
	MemoryUsage   uint64
	MemoryLimit   uint64
	MemoryPercent float64
}

// ListContainers lists all containers with a specific label
func (c *Client) ListContainers(ctx context.Context, labelFilter string) ([]ContainerStatus, error) {
	opts := types.ContainerListOptions{
		All: true,
	}

	if labelFilter != "" {
		filterArgs := filters.NewArgs()
		filterArgs.Add("label", labelFilter)
		opts.Filters = filterArgs
	}

	containers, err := c.cli.ContainerList(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var statuses []ContainerStatus
	for _, cont := range containers {
		name := ""
		if len(cont.Names) > 0 {
			name = strings.TrimPrefix(cont.Names[0], "/")
		}

		var ports []string
		for _, p := range cont.Ports {
			if p.PublicPort != 0 {
				ports = append(ports, fmt.Sprintf("%d:%d", p.PublicPort, p.PrivatePort))
			}
		}

		statuses = append(statuses, ContainerStatus{
			ID:      cont.ID[:12],
			Name:    name,
			Status:  cont.Status,
			State:   cont.State,
			Ports:   ports,
			Created: cont.Created,
			Image:   cont.Image,
			Stats:   nil, // Stats need to be fetched separately
		})
	}

	return statuses, nil
}

// StartContainer starts a container
func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	return c.cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

// StopContainer stops a container
func (c *Client) StopContainer(ctx context.Context, containerID string) error {
	timeout := 30
	stopOptions := container.StopOptions{
		Timeout: &timeout,
	}
	return c.cli.ContainerStop(ctx, containerID, stopOptions)
}

// RemoveContainer removes a container
func (c *Client) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return c.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force:         force,
		RemoveVolumes: true,
	})
}

// GetContainerLogs gets container logs
func (c *Client) GetContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error) {
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: true,
	}

	return c.cli.ContainerLogs(ctx, containerID, options)
}

// RunContainer creates and starts a container
func (c *Client) RunContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, name string) (string, error) {
	resp, err := c.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		// Clean up container if start fails
		_ = c.cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return resp.ID, nil
}

// GetContainerByName finds a container by name
func (c *Client) GetContainerByName(ctx context.Context, name string) (*ContainerStatus, error) {
	containers, err := c.ListContainers(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, cont := range containers {
		if cont.Name == name {
			return &cont, nil
		}
	}

	return nil, fmt.Errorf("container not found: %s", name)
}


// SystemInfo represents Docker system information
type SystemInfo struct {
	Containers int
	Images     int
	Version    string
}

// GetSystemInfo gets Docker system information
func (c *Client) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	info, err := c.cli.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system info: %w", err)
	}

	return &SystemInfo{
		Containers: info.Containers,
		Images:     info.Images,
		Version:    info.ServerVersion,
	}, nil
}

// GetContainerStats gets resource usage statistics for a container
func (c *Client) GetContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	// Get one-time stats
	statsResp, err := c.cli.ContainerStatsOneShot(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer statsResp.Body.Close()

	// Decode stats
	var stats types.StatsJSON
	if err := json.NewDecoder(statsResp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	// Calculate CPU percentage
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	cpuPercent := 0.0
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}

	return &ContainerStats{
		CPUPercent:    cpuPercent,
		MemoryUsage:   stats.MemoryStats.Usage,
		MemoryLimit:   stats.MemoryStats.Limit,
		MemoryPercent: float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0,
	}, nil
}

// ExecInteractive executes a command interactively in a container
func (c *Client) ExecInteractive(ctx context.Context, containerID string, cmd []string) error {
	// Create exec configuration
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	// Create exec instance
	execResp, err := c.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	// Attach to exec instance
	attachResp, err := c.cli.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{
		Tty: true,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer attachResp.Close()

	// Set up terminal
	if err := setupTerminal(attachResp); err != nil {
		return fmt.Errorf("failed to setup terminal: %w", err)
	}

	return nil
}

// setupTerminal sets up terminal for interactive docker exec
func setupTerminal(resp types.HijackedResponse) error {
	// This is a simplified version. In production, you'd want to:
	// 1. Save and restore terminal state
	// 2. Handle terminal resize
	// 3. Properly handle signals
	
	// For now, we'll use docker exec directly via exec.Command
	// This is simpler and handles TTY properly
	return fmt.Errorf("interactive exec not yet fully implemented - use 'docker exec -it' directly")
}

// CreatePortBindings creates port bindings from port mappings
func CreatePortBindings(ports map[string]string) (nat.PortMap, nat.PortSet, error) {
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	for hostPort, containerPort := range ports {
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid port %s: %w", containerPort, err)
		}

		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	return portBindings, exposedPorts, nil
}