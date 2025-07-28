package dockerfiles

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

// Note: go:embed directives must use _ for the embed package to work with non-Go files

//go:embed Dockerfile.base.txt
var dockerfileBase string

//go:embed Dockerfile.node.txt
var dockerfileNode string

//go:embed Dockerfile.python.txt
var dockerfilePython string

//go:embed Dockerfile.python-node.txt
var dockerfilePythonNode string

//go:embed Dockerfile.go.txt
var dockerfileGo string

//go:embed Dockerfile.go-node.txt
var dockerfileGoNode string

//go:embed Dockerfile.full.txt
var dockerfileFull string

//go:embed README.md
var dockerfileReadme string

// dockerfileMap maps dockerfile names to their content
var dockerfileMap = map[string]string{
	"Dockerfile.base":        dockerfileBase,
	"Dockerfile.node":        dockerfileNode,
	"Dockerfile.python":      dockerfilePython,
	"Dockerfile.python-node": dockerfilePythonNode,
	"Dockerfile.go":          dockerfileGo,
	"Dockerfile.go-node":     dockerfileGoNode,
	"Dockerfile.full":        dockerfileFull,
	"README.md":              dockerfileReadme,
}

// ExtractDockerfiles extracts all dockerfiles to the specified directory
func ExtractDockerfiles(targetDir string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create dockerfiles directory: %w", err)
	}

	// Write each dockerfile
	for filename, content := range dockerfileMap {
		path := filepath.Join(targetDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	return nil
}

// IsExtracted checks if dockerfiles have been extracted
func IsExtracted(targetDir string) bool {
	// Check if at least the base dockerfile exists
	baseFile := filepath.Join(targetDir, "Dockerfile.base")
	_, err := os.Stat(baseFile)
	return err == nil
}