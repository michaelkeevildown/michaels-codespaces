package utils

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// Repository represents a Git repository
type Repository struct {
	URL      string
	Host     string // github.com, gitlab.com, etc.
	Owner    string
	Name     string
	IsLocal  bool
}

// ParseRepository parses various repository formats
func ParseRepository(input string) (*Repository, error) {
	// Check if it's a local path
	if input == "." || strings.HasPrefix(input, "./") || strings.HasPrefix(input, "/") {
		absPath, err := filepath.Abs(input)
		if err != nil {
			return nil, err
		}
		return &Repository{
			URL:     absPath,
			Name:    filepath.Base(absPath),
			IsLocal: true,
		}, nil
	}

	// Handle short format (e.g., "facebook/react")
	if !strings.Contains(input, "://") && !strings.HasPrefix(input, "git@") {
		parts := strings.Split(input, "/")
		if len(parts) == 2 {
			return &Repository{
				URL:   fmt.Sprintf("https://github.com/%s", input),
				Host:  "github.com",
				Owner: parts[0],
				Name:  strings.TrimSuffix(parts[1], ".git"),
			}, nil
		}
	}

	// Handle git@ URLs
	if strings.HasPrefix(input, "git@") {
		// Convert git@github.com:user/repo.git to https://github.com/user/repo
		input = strings.Replace(input, ":", "/", 1)
		input = strings.Replace(input, "git@", "https://", 1)
		input = strings.TrimSuffix(input, ".git")
	}

	// Parse as URL
	u, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	// Extract parts from path
	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid repository path: %s", u.Path)
	}

	return &Repository{
		URL:   strings.TrimSuffix(input, ".git"),
		Host:  u.Host,
		Owner: pathParts[0],
		Name:  strings.TrimSuffix(pathParts[1], ".git"),
	}, nil
}