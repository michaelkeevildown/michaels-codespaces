package codespace

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/michaelkeevildown/mcs/internal/components"
	"github.com/michaelkeevildown/mcs/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodespace_Structure(t *testing.T) {
	tests := []struct {
		name     string
		cs       Codespace
		validate func(t *testing.T, cs Codespace)
	}{
		{
			name: "complete codespace",
			cs: Codespace{
				Name:               "test-project",
				Repository:         "https://github.com/user/repo.git",
				Path:               "/home/user/codespaces/test-project",
				Status:             "running",
				CreatedAt:          time.Now(),
				VSCodeURL:          "http://localhost:8080",
				AppURL:             "http://localhost:3000",
				Components:         []string{"go", "postgres"},
				Language:           "go",
				Password:           "testpass123",
				VSCodePort:         8080,
				DockerfileChecksum: "abc123",
			},
			validate: func(t *testing.T, cs Codespace) {
				assert.Equal(t, "test-project", cs.Name)
				assert.Equal(t, "https://github.com/user/repo.git", cs.Repository)
				assert.Equal(t, "/home/user/codespaces/test-project", cs.Path)
				assert.Equal(t, "running", cs.Status)
				assert.Equal(t, "http://localhost:8080", cs.VSCodeURL)
				assert.Equal(t, "http://localhost:3000", cs.AppURL)
				assert.Equal(t, []string{"go", "postgres"}, cs.Components)
				assert.Equal(t, "go", cs.Language)
				assert.Equal(t, "testpass123", cs.Password)
				assert.Equal(t, 8080, cs.VSCodePort)
				assert.Equal(t, "abc123", cs.DockerfileChecksum)
				assert.NotZero(t, cs.CreatedAt)
			},
		},
		{
			name: "minimal codespace",
			cs: Codespace{
				Name:       "minimal",
				Repository: "git@github.com:user/minimal.git",
				Path:       "/tmp/minimal",
				Status:     "stopped",
				CreatedAt:  time.Now(),
			},
			validate: func(t *testing.T, cs Codespace) {
				assert.Equal(t, "minimal", cs.Name)
				assert.Equal(t, "git@github.com:user/minimal.git", cs.Repository)
				assert.Equal(t, "/tmp/minimal", cs.Path)
				assert.Equal(t, "stopped", cs.Status)
				assert.Empty(t, cs.VSCodeURL)
				assert.Empty(t, cs.AppURL)
				assert.Empty(t, cs.Components)
				assert.Empty(t, cs.Language)
				assert.Empty(t, cs.Password)
				assert.Zero(t, cs.VSCodePort)
				assert.NotZero(t, cs.CreatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.cs)
		})
	}
}

func TestCreateOptions_GetPath(t *testing.T) {
	tests := []struct {
		name         string
		opts         CreateOptions
		expectedPath func() string
	}{
		{
			name: "simple name",
			opts: CreateOptions{
				Name: "my-project",
			},
			expectedPath: func() string {
				return filepath.Join(utils.GetHomeDir(), "codespaces", "my-project")
			},
		},
		{
			name: "name with special chars",
			opts: CreateOptions{
				Name: "user-repo_name",
			},
			expectedPath: func() string {
				return filepath.Join(utils.GetHomeDir(), "codespaces", "user-repo_name")
			},
		},
		{
			name: "empty name still generates path",
			opts: CreateOptions{
				Name: "",
			},
			expectedPath: func() string {
				return filepath.Join(utils.GetHomeDir(), "codespaces", "")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.opts.GetPath()
			expectedPath := tt.expectedPath()
			assert.Equal(t, expectedPath, path)
		})
	}
}

func TestCreateOptions_Validate(t *testing.T) {
	validRepo := &utils.Repository{
		URL:   "https://github.com/user/repo.git",
		Owner: "user",
		Name:  "repo",
	}

	tests := []struct {
		name        string
		opts        CreateOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid options",
			opts: CreateOptions{
				Name:       "test-project",
				Repository: validRepo,
			},
			expectError: false,
		},
		{
			name: "missing name",
			opts: CreateOptions{
				Name:       "",
				Repository: validRepo,
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "missing repository",
			opts: CreateOptions{
				Name:       "test-project",
				Repository: nil,
			},
			expectError: true,
			errorMsg:    "repository is required",
		},
		{
			name: "missing both",
			opts: CreateOptions{
				Name:       "",
				Repository: nil,
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "with components",
			opts: CreateOptions{
				Name:       "test-project",
				Repository: validRepo,
				Components: []components.Component{
					{ID: "go", Name: "Go", Installer: "go.sh"},
					{ID: "postgres", Name: "PostgreSQL", Installer: "postgres.sh"},
				},
			},
			expectError: false,
		},
		{
			name: "with progress callback",
			opts: CreateOptions{
				Name:       "test-project",
				Repository: validRepo,
				Progress: func(msg string) {
					// Progress callback
				},
			},
			expectError: false,
		},
		{
			name: "with clone depth",
			opts: CreateOptions{
				Name:       "test-project",
				Repository: validRepo,
				CloneDepth: 10,
			},
			expectError: false,
		},
		{
			name: "with no-start flag",
			opts: CreateOptions{
				Name:       "test-project",
				Repository: validRepo,
				NoStart:    true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProgressFunc(t *testing.T) {
	t.Run("progress callback", func(t *testing.T) {
		var messages []string
		progress := func(msg string) {
			messages = append(messages, msg)
		}

		opts := CreateOptions{
			Name:       "test",
			Repository: &utils.Repository{URL: "test.git"},
			Progress:   progress,
		}

		// Simulate progress calls
		if opts.Progress != nil {
			opts.Progress("Step 1")
			opts.Progress("Step 2")
			opts.Progress("Step 3")
		}

		assert.Equal(t, []string{"Step 1", "Step 2", "Step 3"}, messages)
	})

	t.Run("nil progress callback", func(t *testing.T) {
		opts := CreateOptions{
			Name:       "test",
			Repository: &utils.Repository{URL: "test.git"},
			Progress:   nil,
		}

		// Should not panic
		if opts.Progress != nil {
			opts.Progress("This won't be called")
		}
	})
}

func TestManager_NewManager(t *testing.T) {
	t.Run("creates manager with correct base directory", func(t *testing.T) {
		manager := NewManager()
		assert.NotNil(t, manager)
		expectedBaseDir := filepath.Join(utils.GetHomeDir(), "codespaces")
		assert.Equal(t, expectedBaseDir, manager.baseDir)
	})
}

func TestCodespaceStatus(t *testing.T) {
	validStatuses := []string{"running", "stopped", "error", "created"}
	
	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			cs := Codespace{
				Name:   "test",
				Status: status,
			}
			assert.Equal(t, status, cs.Status)
		})
	}
}

func TestCodespaceComponents(t *testing.T) {
	t.Run("empty components", func(t *testing.T) {
		cs := Codespace{
			Name:       "test",
			Components: []string{},
		}
		assert.Empty(t, cs.Components)
		assert.Len(t, cs.Components, 0)
	})

	t.Run("single component", func(t *testing.T) {
		cs := Codespace{
			Name:       "test",
			Components: []string{"go"},
		}
		assert.Len(t, cs.Components, 1)
		assert.Contains(t, cs.Components, "go")
	})

	t.Run("multiple components", func(t *testing.T) {
		cs := Codespace{
			Name:       "test",
			Components: []string{"go", "postgres", "redis"},
		}
		assert.Len(t, cs.Components, 3)
		assert.Contains(t, cs.Components, "go")
		assert.Contains(t, cs.Components, "postgres")
		assert.Contains(t, cs.Components, "redis")
	})
}