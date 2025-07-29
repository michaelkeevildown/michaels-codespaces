package codespace

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test codespace with metadata
func createTestCodespace(t *testing.T, baseDir, name string) *Codespace {
	cs := &Codespace{
		Name:               name,
		Repository:         "https://github.com/test/repo.git",
		Path:               filepath.Join(baseDir, name),
		CreatedAt:          time.Now(),
		VSCodeURL:          "http://localhost:8080",
		AppURL:             "http://localhost:3000",
		Components:         []string{"go", "postgres"},
		Language:           "go",
		Password:           "testpass123",
		DockerfileChecksum: "abc123",
	}

	// Create directory structure
	metadataDir := filepath.Join(cs.Path, ".mcs")
	err := os.MkdirAll(metadataDir, 0755)
	require.NoError(t, err)

	// Save metadata
	metadata := Metadata{
		Name:               cs.Name,
		Repository:         cs.Repository,
		Path:               cs.Path,
		CreatedAt:          cs.CreatedAt,
		VSCodeURL:          cs.VSCodeURL,
		AppURL:             cs.AppURL,
		Components:         cs.Components,
		Language:           cs.Language,
		Password:           cs.Password,
		DockerfileChecksum: cs.DockerfileChecksum,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)

	metadataPath := filepath.Join(metadataDir, "metadata.json")
	err = os.WriteFile(metadataPath, data, 0644)
	require.NoError(t, err)

	return cs
}

func TestManager_SaveMetadata(t *testing.T) {
	tempDir := t.TempDir()

	// Create manager with temporary base directory
	manager := &Manager{
		baseDir: filepath.Join(tempDir, "codespaces"),
	}

	t.Run("save new metadata", func(t *testing.T) {
		cs := &Codespace{
			Name:               "test-project",
			Repository:         "https://github.com/test/repo.git",
			Path:               filepath.Join(tempDir, "codespaces", "test-project"),
			CreatedAt:          time.Now(),
			VSCodeURL:          "http://localhost:8080",
			AppURL:             "http://localhost:3000",
			Components:         []string{"go", "postgres"},
			Language:           "go",
			Password:           "secure123",
			DockerfileChecksum: "checksum123",
		}

		// Create base directory
		err := os.MkdirAll(cs.Path, 0755)
		require.NoError(t, err)

		err = manager.SaveMetadata(cs)
		require.NoError(t, err)

		// Verify metadata file exists
		metadataPath := filepath.Join(cs.Path, ".mcs", "metadata.json")
		assert.FileExists(t, metadataPath)

		// Read and verify content
		data, err := os.ReadFile(metadataPath)
		require.NoError(t, err)

		var metadata Metadata
		err = json.Unmarshal(data, &metadata)
		require.NoError(t, err)

		assert.Equal(t, cs.Name, metadata.Name)
		assert.Equal(t, cs.Repository, metadata.Repository)
		assert.Equal(t, cs.Path, metadata.Path)
		assert.Equal(t, cs.VSCodeURL, metadata.VSCodeURL)
		assert.Equal(t, cs.AppURL, metadata.AppURL)
		assert.Equal(t, cs.Components, metadata.Components)
		assert.Equal(t, cs.Language, metadata.Language)
		assert.Equal(t, cs.Password, metadata.Password)
		assert.Equal(t, cs.DockerfileChecksum, metadata.DockerfileChecksum)
		assert.WithinDuration(t, cs.CreatedAt, metadata.CreatedAt, time.Second)
	})

	t.Run("overwrite existing metadata", func(t *testing.T) {
		cs := &Codespace{
			Name:       "existing-project",
			Repository: "https://github.com/test/old.git",
			Path:       filepath.Join(tempDir, "codespaces", "existing-project"),
			CreatedAt:  time.Now().Add(-time.Hour),
			Language:   "python",
		}

		// Create directory and save initial metadata
		err := os.MkdirAll(filepath.Join(cs.Path, ".mcs"), 0755)
		require.NoError(t, err)
		err = manager.SaveMetadata(cs)
		require.NoError(t, err)

		// Update and save again
		cs.Repository = "https://github.com/test/new.git"
		cs.Language = "go"
		cs.Components = []string{"docker"}

		err = manager.SaveMetadata(cs)
		require.NoError(t, err)

		// Verify updated metadata
		metadata, err := manager.loadMetadata(cs.Name)
		require.NoError(t, err)

		assert.Equal(t, "https://github.com/test/new.git", metadata.Repository)
		assert.Equal(t, "go", metadata.Language)
		assert.Equal(t, []string{"docker"}, metadata.Components)
	})

	t.Run("permission error", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Cannot test permission errors as root")
		}

		cs := &Codespace{
			Name: "no-permission",
			Path: "/root/no-access/codespace",
		}

		err := manager.SaveMetadata(cs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create metadata directory")
	})
}

func TestManager_LoadMetadata(t *testing.T) {
	tempDir := t.TempDir()

	// Create manager with temporary base directory
	manager := &Manager{
		baseDir: filepath.Join(tempDir, "codespaces"),
	}
	baseDir := manager.baseDir

	t.Run("load existing metadata", func(t *testing.T) {
		// Create test codespace
		cs := createTestCodespace(t, baseDir, "test-load")

		// Load metadata
		metadata, err := manager.loadMetadata("test-load")
		require.NoError(t, err)

		assert.Equal(t, cs.Name, metadata.Name)
		assert.Equal(t, cs.Repository, metadata.Repository)
		assert.Equal(t, cs.Path, metadata.Path)
		assert.Equal(t, cs.VSCodeURL, metadata.VSCodeURL)
		assert.Equal(t, cs.AppURL, metadata.AppURL)
		assert.Equal(t, cs.Components, metadata.Components)
		assert.Equal(t, cs.Language, metadata.Language)
		assert.Equal(t, cs.Password, metadata.Password)
		assert.Equal(t, cs.DockerfileChecksum, metadata.DockerfileChecksum)
	})

	t.Run("metadata not found", func(t *testing.T) {
		_, err := manager.loadMetadata("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read metadata")
	})

	t.Run("corrupted metadata", func(t *testing.T) {
		// Create directory with corrupted metadata
		corruptDir := filepath.Join(baseDir, "corrupted")
		metadataPath := filepath.Join(corruptDir, ".mcs", "metadata.json")
		err := os.MkdirAll(filepath.Dir(metadataPath), 0755)
		require.NoError(t, err)

		// Write invalid JSON
		err = os.WriteFile(metadataPath, []byte("invalid json"), 0644)
		require.NoError(t, err)

		_, err = manager.loadMetadata("corrupted")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal metadata")
	})
}

func TestManager_List(t *testing.T) {
	tempDir := t.TempDir()

	// Create manager with temporary base directory
	manager := &Manager{
		baseDir: filepath.Join(tempDir, "codespaces"),
	}
	baseDir := manager.baseDir

	t.Run("list empty directory", func(t *testing.T) {
		ctx := context.Background()
		codespaces, err := manager.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, codespaces)
	})

	t.Run("list multiple codespaces", func(t *testing.T) {
		// Create test codespaces
		cs1 := createTestCodespace(t, baseDir, "project-1")
		cs2 := createTestCodespace(t, baseDir, "project-2")
		cs3 := createTestCodespace(t, baseDir, "project-3")

		// Create a directory without metadata (should be skipped)
		err := os.MkdirAll(filepath.Join(baseDir, "not-a-codespace"), 0755)
		require.NoError(t, err)

		// Create a file (should be skipped)
		err = os.WriteFile(filepath.Join(baseDir, "readme.txt"), []byte("test"), 0644)
		require.NoError(t, err)

		ctx := context.Background()
		codespaces, err := manager.List(ctx)
		require.NoError(t, err)

		// Should find exactly 3 codespaces
		assert.Len(t, codespaces, 3)

		// Verify codespace data
		names := make(map[string]bool)
		for _, cs := range codespaces {
			names[cs.Name] = true
			assert.NotEmpty(t, cs.Repository)
			assert.NotEmpty(t, cs.Path)
			assert.NotZero(t, cs.CreatedAt)
			assert.Equal(t, "stopped", cs.Status) // Default status without Docker
		}

		assert.True(t, names[cs1.Name])
		assert.True(t, names[cs2.Name])
		assert.True(t, names[cs3.Name])
	})

	t.Run("handle missing directory gracefully", func(t *testing.T) {
		// Remove the codespaces directory
		err := os.RemoveAll(baseDir)
		require.NoError(t, err)

		ctx := context.Background()
		codespaces, err := manager.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, codespaces)
	})
}

func TestManager_Get(t *testing.T) {
	tempDir := t.TempDir()

	// Create manager with temporary base directory
	manager := &Manager{
		baseDir: filepath.Join(tempDir, "codespaces"),
	}
	baseDir := manager.baseDir

	t.Run("get existing codespace", func(t *testing.T) {
		// Create test codespace
		expected := createTestCodespace(t, baseDir, "test-get")

		ctx := context.Background()
		cs, err := manager.Get(ctx, "test-get")
		require.NoError(t, err)

		assert.Equal(t, expected.Name, cs.Name)
		assert.Equal(t, expected.Repository, cs.Repository)
		assert.Equal(t, expected.Path, cs.Path)
		assert.Equal(t, expected.VSCodeURL, cs.VSCodeURL)
		assert.Equal(t, expected.AppURL, cs.AppURL)
		assert.Equal(t, expected.Components, cs.Components)
		assert.Equal(t, expected.Language, cs.Language)
		assert.Equal(t, expected.Password, cs.Password)
		assert.Equal(t, "stopped", cs.Status) // Default without Docker
	})

	t.Run("get non-existent codespace", func(t *testing.T) {
		ctx := context.Background()
		_, err := manager.Get(ctx, "does-not-exist")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "codespace not found")
	})
}

func TestManager_UpdateMetadataChecksum(t *testing.T) {
	tempDir := t.TempDir()

	// Create manager with temporary base directory
	manager := &Manager{
		baseDir: filepath.Join(tempDir, "codespaces"),
	}
	baseDir := manager.baseDir

	t.Run("update checksum", func(t *testing.T) {
		// Create test codespace
		cs := createTestCodespace(t, baseDir, "test-checksum")
		originalChecksum := cs.DockerfileChecksum

		// Update checksum
		newChecksum := "new-checksum-xyz"
		err := manager.updateMetadataChecksum("test-checksum", newChecksum)
		require.NoError(t, err)

		// Verify update
		metadata, err := manager.loadMetadata("test-checksum")
		require.NoError(t, err)

		assert.Equal(t, newChecksum, metadata.DockerfileChecksum)
		assert.NotEqual(t, originalChecksum, metadata.DockerfileChecksum)

		// Verify other fields unchanged
		assert.Equal(t, cs.Name, metadata.Name)
		assert.Equal(t, cs.Repository, metadata.Repository)
		assert.Equal(t, cs.Language, metadata.Language)
	})

	t.Run("update non-existent codespace", func(t *testing.T) {
		err := manager.updateMetadataChecksum("does-not-exist", "checksum")
		assert.Error(t, err)
	})
}

func TestManager_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()

	// Create manager with temporary base directory
	manager := &Manager{
		baseDir: filepath.Join(tempDir, "codespaces"),
	}

	// These tests verify the methods exist and handle basic error cases
	// Full integration testing would require Docker mocking

	t.Run("start non-existent codespace", func(t *testing.T) {
		ctx := context.Background()
		err := manager.Start(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "codespace not found")
	})

	t.Run("stop non-existent codespace", func(t *testing.T) {
		ctx := context.Background()
		err := manager.Stop(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "codespace not found")
	})

	t.Run("remove non-existent codespace", func(t *testing.T) {
		ctx := context.Background()
		err := manager.Remove(ctx, "non-existent", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "codespace not found")
	})

	t.Run("get logs for non-existent codespace", func(t *testing.T) {
		ctx := context.Background()
		_, err := manager.GetLogs(ctx, "non-existent", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "codespace not found")
	})
}

func TestMetadataStructure(t *testing.T) {
	t.Run("metadata JSON marshaling", func(t *testing.T) {
		original := Metadata{
			Name:               "test",
			Repository:         "https://github.com/test/repo.git",
			Path:               "/path/to/codespace",
			CreatedAt:          time.Now().Truncate(time.Second),
			VSCodeURL:          "http://localhost:8080",
			AppURL:             "http://localhost:3000",
			Components:         []string{"go", "postgres"},
			Language:           "go",
			Password:           "secure123",
			DockerfileChecksum: "abc123",
		}

		// Marshal to JSON
		data, err := json.MarshalIndent(original, "", "  ")
		require.NoError(t, err)

		// Unmarshal back
		var restored Metadata
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		// Verify fields
		assert.Equal(t, original.Name, restored.Name)
		assert.Equal(t, original.Repository, restored.Repository)
		assert.Equal(t, original.Path, restored.Path)
		assert.Equal(t, original.CreatedAt.Unix(), restored.CreatedAt.Unix())
		assert.Equal(t, original.VSCodeURL, restored.VSCodeURL)
		assert.Equal(t, original.AppURL, restored.AppURL)
		assert.Equal(t, original.Components, restored.Components)
		assert.Equal(t, original.Language, restored.Language)
		assert.Equal(t, original.Password, restored.Password)
		assert.Equal(t, original.DockerfileChecksum, restored.DockerfileChecksum)
	})

	t.Run("metadata with empty optional fields", func(t *testing.T) {
		minimal := Metadata{
			Name:      "minimal",
			CreatedAt: time.Now(),
		}

		data, err := json.Marshal(minimal)
		require.NoError(t, err)

		var restored Metadata
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		assert.Equal(t, minimal.Name, restored.Name)
		assert.Empty(t, restored.Repository)
		assert.Empty(t, restored.Path)
		assert.Empty(t, restored.VSCodeURL)
		assert.Empty(t, restored.AppURL)
		assert.Empty(t, restored.Components)
		assert.Empty(t, restored.Language)
		assert.Empty(t, restored.Password)
		assert.Empty(t, restored.DockerfileChecksum)
	})
}