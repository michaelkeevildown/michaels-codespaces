package codespace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/michaelkeevildown/mcs/internal/docker"
	"github.com/michaelkeevildown/mcs/internal/ports"
)

// Metadata represents codespace metadata stored on disk
type Metadata struct {
	Name       string    `json:"name"`
	Repository string    `json:"repository"`
	Path       string    `json:"path"`
	CreatedAt  time.Time `json:"created_at"`
	VSCodeURL  string    `json:"vscode_url"`
	AppURL     string    `json:"app_url"`
	Components []string  `json:"components"`
	Language   string    `json:"language"`
	Password   string    `json:"password"`
}

// SaveMetadata saves codespace metadata
func (m *Manager) SaveMetadata(cs *Codespace) error {
	metadataPath := filepath.Join(cs.Path, ".mcs", "metadata.json")
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(metadataPath), 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	metadata := Metadata{
		Name:       cs.Name,
		Repository: cs.Repository,
		Path:       cs.Path,
		CreatedAt:  cs.CreatedAt,
		VSCodeURL:  cs.VSCodeURL,
		AppURL:     cs.AppURL,
		Components: cs.Components,
		Language:   cs.Language,
		Password:   cs.Password,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return os.WriteFile(metadataPath, data, 0644)
}

// loadMetadata loads codespace metadata
func (m *Manager) loadMetadata(name string) (*Metadata, error) {
	codespaceDir := filepath.Join(m.baseDir, name)
	metadataPath := filepath.Join(codespaceDir, ".mcs", "metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// List returns all codespaces
func (m *Manager) List(ctx context.Context) ([]Codespace, error) {
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Codespace{}, nil
		}
		return nil, fmt.Errorf("failed to read codespaces directory: %w", err)
	}

	// Get Docker client for status
	dockerClient, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	// Get all containers
	containers, err := dockerClient.ListContainers(ctx, "mcs.managed=true")
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// Create container map for quick lookup
	containerMap := make(map[string]*docker.ContainerStatus)
	for i := range containers {
		containerMap[containers[i].Name] = &containers[i]
	}

	var codespaces []Codespace
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Try to load metadata
		metadata, err := m.loadMetadata(entry.Name())
		if err != nil {
			// Skip directories without metadata
			continue
		}

		cs := Codespace{
			Name:       metadata.Name,
			Repository: metadata.Repository,
			Path:       metadata.Path,
			CreatedAt:  metadata.CreatedAt,
			VSCodeURL:  metadata.VSCodeURL,
			AppURL:     metadata.AppURL,
			Components: metadata.Components,
			Language:   metadata.Language,
			Password:   metadata.Password,
			Status:     "stopped",
		}

		// Check container status
		containerName := fmt.Sprintf("%s-dev", cs.Name)
		if container, ok := containerMap[containerName]; ok {
			if container.State == "running" {
				cs.Status = "running"
			}
		}

		codespaces = append(codespaces, cs)
	}

	return codespaces, nil
}

// Get returns a specific codespace
func (m *Manager) Get(ctx context.Context, name string) (*Codespace, error) {
	metadata, err := m.loadMetadata(name)
	if err != nil {
		return nil, fmt.Errorf("codespace not found: %s", name)
	}

	cs := &Codespace{
		Name:       metadata.Name,
		Repository: metadata.Repository,
		Path:       metadata.Path,
		CreatedAt:  metadata.CreatedAt,
		VSCodeURL:  metadata.VSCodeURL,
		AppURL:     metadata.AppURL,
		Components: metadata.Components,
		Language:   metadata.Language,
		Password:   metadata.Password,
		Status:     "stopped",
	}

	// Check container status
	dockerClient, err := docker.NewClient()
	if err != nil {
		return cs, nil // Return codespace without status
	}
	defer dockerClient.Close()

	containerName := fmt.Sprintf("%s-dev", name)
	container, err := dockerClient.GetContainerByName(ctx, containerName)
	if err == nil && container.State == "running" {
		cs.Status = "running"
	}

	return cs, nil
}

// Start starts a codespace
func (m *Manager) Start(ctx context.Context, name string) error {
	// Verify codespace exists
	metadata, err := m.loadMetadata(name)
	if err != nil {
		return fmt.Errorf("codespace not found: %s", name)
	}

	// Create Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	// Check if container already exists
	containerName := fmt.Sprintf("%s-dev", name)
	container, err := dockerClient.GetContainerByName(ctx, containerName)
	
	if err == nil {
		// Container exists, just start it
		if container.State != "running" {
			return dockerClient.StartContainer(ctx, container.ID)
		}
		return nil // Already running
	}

	// Container doesn't exist, use docker-compose to create and start
	composeExecutor := docker.NewComposeExecutor(metadata.Path)
	
	// Run docker-compose up -d
	if err := composeExecutor.Up(ctx, true); err != nil {
		return fmt.Errorf("failed to start container with docker-compose: %w", err)
	}

	return nil
}

// Stop stops a codespace
func (m *Manager) Stop(ctx context.Context, name string) error {
	// Verify codespace exists
	if _, err := m.loadMetadata(name); err != nil {
		return fmt.Errorf("codespace not found: %s", name)
	}

	// Create Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	// Get container
	containerName := fmt.Sprintf("%s-dev", name)
	container, err := dockerClient.GetContainerByName(ctx, containerName)
	if err != nil {
		return fmt.Errorf("container not found: %s", containerName)
	}

	// Stop container
	return dockerClient.StopContainer(ctx, container.ID)
}

// Remove removes a codespace
func (m *Manager) Remove(ctx context.Context, name string, force bool) error {
	// Verify codespace exists
	metadata, err := m.loadMetadata(name)
	if err != nil {
		return fmt.Errorf("codespace not found: %s", name)
	}

	// Create Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	// Try to remove container
	containerName := fmt.Sprintf("%s-dev", name)
	container, err := dockerClient.GetContainerByName(ctx, containerName)
	if err == nil {
		// Stop container if running
		if container.State == "running" && !force {
			return fmt.Errorf("container is running, use --force to remove")
		}

		// Remove container
		if err := dockerClient.RemoveContainer(ctx, container.ID, force); err != nil {
			return fmt.Errorf("failed to remove container: %w", err)
		}
	}

	// Release ports
	portRegistry, err := ports.NewPortRegistry()
	if err == nil {
		portRegistry.ReleaseCodespacePorts(name)
	}

	// Remove directory
	if err := os.RemoveAll(metadata.Path); err != nil {
		return fmt.Errorf("failed to remove codespace directory: %w", err)
	}

	return nil
}

// GetLogs gets logs for a codespace
func (m *Manager) GetLogs(ctx context.Context, name string, follow bool) (string, error) {
	// Verify codespace exists
	if _, err := m.loadMetadata(name); err != nil {
		return "", fmt.Errorf("codespace not found: %s", name)
	}

	// Create Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		return "", fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	// Get container
	containerName := fmt.Sprintf("%s-dev", name)
	container, err := dockerClient.GetContainerByName(ctx, containerName)
	if err != nil {
		return "", fmt.Errorf("container not found: %s", containerName)
	}

	// Get logs
	reader, err := dockerClient.GetContainerLogs(ctx, container.ID, follow)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}
	defer reader.Close()

	// TODO: Implement proper log streaming
	return "Log streaming not yet implemented", nil
}