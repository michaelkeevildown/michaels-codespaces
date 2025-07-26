package docker

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/michaelkeevildown/mcs/internal/components"
)

// ComposeConfig holds configuration for docker-compose generation
type ComposeConfig struct {
	ContainerName string
	CodespaceName string
	Image         string
	Password      string
	Ports         map[string]string // host:container
	Environment   map[string]string
	Labels        map[string]string
	Components    []components.Component
	WorkingDir    string
}

// Default images for different languages
var languageImages = map[string]string{
	"python":     "codercom/code-server:latest",
	"node":       "codercom/code-server:latest",
	"go":         "codercom/code-server:latest",
	"rust":       "codercom/code-server:latest",
	"java":       "codercom/code-server:latest",
	"php":        "codercom/code-server:latest",
	"ruby":       "codercom/code-server:latest",
	"default":    "codercom/code-server:latest",
}

const dockerComposeTemplate = `services:
  {{ .ContainerName }}:
    image: {{ .Image }}
    container_name: {{ .ContainerName }}
    restart: unless-stopped
    environment:
      - PASSWORD={{ .Password }}
      - TZ=${TZ:-UTC}
      - DOCKER_USER=${USER}
      {{- range $key, $value := .Environment }}
      - {{ $key }}={{ $value }}
      {{- end }}
    ports:
      {{- range $host, $container := .Ports }}
      - "{{ $host }}:{{ $container }}"
      {{- end }}
    volumes:
      - ./src:/home/coder/{{ .CodespaceName }}
      - ./data:/home/coder/.local/share/code-server
      - ./config:/home/coder/.config
      - ./logs:/home/coder/logs
      - ${HOME}/.ssh:/home/coder/.ssh:ro
      - ${HOME}/.gitconfig:/home/coder/.gitconfig:ro
      {{- if .Components }}
      - ./components:/home/coder/.components:ro
      - ./init:/docker-entrypoint-initdb.d:ro
      {{- end }}
    labels:
      {{- range $key, $value := .Labels }}
      {{ $key }}: "{{ $value }}"
      {{- end }}
    working_dir: /home/coder/{{ .CodespaceName }}
    command: >
      sh -c "
        {{- if .Components }}
        # Install components
        if [ -f /docker-entrypoint-initdb.d/init.sh ]; then
          echo 'Installing components...' &&
          chmod +x /docker-entrypoint-initdb.d/init.sh &&
          /docker-entrypoint-initdb.d/init.sh
        fi &&
        {{- end }}
        # Start code-server
        code-server --bind-addr 0.0.0.0:8080 --auth password
      "
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

networks:
  default:
    name: mcs-network
    external: true
`

const initScriptTemplate = `#!/bin/bash
# Component installation script
set -e

echo "🚀 Installing MCS components..."

# Create component directory
mkdir -p /home/coder/.mcs/components

{{- range .Components }}
{{- if .Selected }}

echo "📦 Installing {{ .Name }}..."
if [ -f /home/coder/.components/{{ .Installer }} ]; then
    chmod +x /home/coder/.components/{{ .Installer }}
    /home/coder/.components/{{ .Installer }}
    echo "✅ {{ .Name }} installed successfully"
else
    echo "⚠️  Installer not found for {{ .Name }}"
fi
{{- end }}
{{- end }}

echo "✨ Component installation complete!"
`

// GenerateDockerCompose generates a docker-compose.yml file
func GenerateDockerCompose(config ComposeConfig) ([]byte, error) {
	// Set defaults
	if config.Image == "" {
		config.Image = "codercom/code-server:latest"
	}

	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}

	// Add standard labels
	config.Labels["mcs.codespace"] = config.CodespaceName
	config.Labels["mcs.managed"] = "true"

	// Parse template
	tmpl, err := template.New("docker-compose").Parse(dockerComposeTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateInitScript generates the component installation script
func GenerateInitScript(components []components.Component) ([]byte, error) {
	tmpl, err := template.New("init-script").Parse(initScriptTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse init script template: %w", err)
	}

	data := struct {
		Components []components.Component
	}{
		Components: components,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute init script template: %w", err)
	}

	return buf.Bytes(), nil
}

// GetImageForLanguage returns the appropriate Docker image for a language
func GetImageForLanguage(language string) string {
	if image, ok := languageImages[strings.ToLower(language)]; ok {
		return image
	}
	return languageImages["default"]
}

// GenerateEnvFile generates a .env file for the codespace
func GenerateEnvFile(config ComposeConfig) []byte {
	var lines []string

	lines = append(lines, "# MCS Codespace Environment")
	lines = append(lines, fmt.Sprintf("CODESPACE_NAME=%s", config.CodespaceName))
	lines = append(lines, fmt.Sprintf("PASSWORD=%s", config.Password))
	lines = append(lines, "")

	// Add custom environment variables
	for key, value := range config.Environment {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	return []byte(strings.Join(lines, "\n") + "\n")
}

// CreateDockerNetwork ensures the MCS network exists
func (c *Client) CreateDockerNetwork(ctx context.Context) error {
	// Check if network already exists
	networks, err := c.cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	for _, network := range networks {
		if network.Name == "mcs-network" {
			return nil // Network already exists
		}
	}

	// Create network
	_, err = c.cli.NetworkCreate(ctx, "mcs-network", types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "bridge",
		Labels: map[string]string{
			"mcs.managed": "true",
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	return nil
}