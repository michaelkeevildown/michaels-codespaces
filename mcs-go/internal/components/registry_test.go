package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByID(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expected    *Component
		expectError bool
	}{
		{
			name: "valid claude component",
			id:   "claude",
			expected: &Component{
				ID:          "claude",
				Name:        "Claude Code",
				Description: "Anthropic's Claude AI coding assistant - your AI pair programmer",
				Emoji:       "🤖",
				Selected:    true,
				Installer:   "claude.sh",
				Requires:    []string{"nodejs"},
			},
			expectError: false,
		},
		{
			name: "valid claude-flow component",
			id:   "claude-flow",
			expected: &Component{
				ID:          "claude-flow",
				Name:        "Claude Flow",
				Description: "AI swarm orchestration and workflow automation",
				Emoji:       "🌊",
				Selected:    true,
				Installer:   "claude-flow.sh",
				DependsOn:   []string{"claude"},
				Requires:    []string{"nodejs"},
			},
			expectError: false,
		},
		{
			name: "valid github-cli component",
			id:   "github-cli",
			expected: &Component{
				ID:          "github-cli",
				Name:        "GitHub CLI",
				Description: "Command-line interface for GitHub with token authentication",
				Emoji:       "🐙",
				Selected:    true,
				Installer:   "github-cli.sh",
				Requires:    []string{},
			},
			expectError: false,
		},
		{
			name:        "non-existent component",
			id:          "non-existent",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "empty id",
			id:          "",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetByID(tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "component not found")
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.Description, result.Description)
				assert.Equal(t, tt.expected.Emoji, result.Emoji)
				assert.Equal(t, tt.expected.Selected, result.Selected)
				assert.Equal(t, tt.expected.Installer, result.Installer)
				assert.Equal(t, tt.expected.DependsOn, result.DependsOn)
				assert.Equal(t, tt.expected.Requires, result.Requires)
			}
		})
	}
}

func TestGetSelected(t *testing.T) {
	// Store original registry to restore later
	originalRegistry := make([]Component, len(Registry))
	copy(originalRegistry, Registry)
	defer func() {
		Registry = originalRegistry
	}()

	tests := []struct {
		name              string
		setupRegistry     func()
		expectedCount     int
		expectedIDs       []string
		expectedNames     []string
	}{
		{
			name: "all components selected (default state)",
			setupRegistry: func() {
				for i := range Registry {
					Registry[i].Selected = true
				}
			},
			expectedCount: 3,
			expectedIDs:   []string{"claude", "claude-flow", "github-cli"},
			expectedNames: []string{"Claude Code", "Claude Flow", "GitHub CLI"},
		},
		{
			name: "no components selected",
			setupRegistry: func() {
				for i := range Registry {
					Registry[i].Selected = false
				}
			},
			expectedCount: 0,
			expectedIDs:   []string{},
			expectedNames: []string{},
		},
		{
			name: "only claude selected",
			setupRegistry: func() {
				for i := range Registry {
					Registry[i].Selected = false
				}
				Registry[0].Selected = true // claude
			},
			expectedCount: 1,
			expectedIDs:   []string{"claude"},
			expectedNames: []string{"Claude Code"},
		},
		{
			name: "claude and github-cli selected",
			setupRegistry: func() {
				for i := range Registry {
					Registry[i].Selected = false
				}
				Registry[0].Selected = true // claude
				Registry[2].Selected = true // github-cli
			},
			expectedCount: 2,
			expectedIDs:   []string{"claude", "github-cli"},
			expectedNames: []string{"Claude Code", "GitHub CLI"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupRegistry()

			result := GetSelected()

			assert.Len(t, result, tt.expectedCount)

			if tt.expectedCount > 0 {
				var actualIDs []string
				var actualNames []string
				for _, comp := range result {
					actualIDs = append(actualIDs, comp.ID)
					actualNames = append(actualNames, comp.Name)
					assert.True(t, comp.Selected, "Component %s should be selected", comp.ID)
				}
				assert.ElementsMatch(t, tt.expectedIDs, actualIDs)
				assert.ElementsMatch(t, tt.expectedNames, actualNames)
			}
		})
	}
}

func TestGetSelectedIDs(t *testing.T) {
	// Store original registry to restore later
	originalRegistry := make([]Component, len(Registry))
	copy(originalRegistry, Registry)
	defer func() {
		Registry = originalRegistry
	}()

	tests := []struct {
		name          string
		setupRegistry func()
		expected      []string
	}{
		{
			name: "all components selected",
			setupRegistry: func() {
				for i := range Registry {
					Registry[i].Selected = true
				}
			},
			expected: []string{"claude", "claude-flow", "github-cli"},
		},
		{
			name: "no components selected",
			setupRegistry: func() {
				for i := range Registry {
					Registry[i].Selected = false
				}
			},
			expected: []string{},
		},
		{
			name: "mixed selection",
			setupRegistry: func() {
				Registry[0].Selected = true  // claude
				Registry[1].Selected = false // claude-flow
				Registry[2].Selected = true  // github-cli
			},
			expected: []string{"claude", "github-cli"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupRegistry()

			result := GetSelectedIDs()

			if len(tt.expected) == 0 {
				assert.Empty(t, result)
			} else {
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestGetSystemRequirements(t *testing.T) {
	// Store original registry to restore later
	originalRegistry := make([]Component, len(Registry))
	copy(originalRegistry, Registry)
	defer func() {
		Registry = originalRegistry
	}()

	tests := []struct {
		name          string
		setupRegistry func()
		expected      []string
	}{
		{
			name: "all components selected (default)",
			setupRegistry: func() {
				for i := range Registry {
					Registry[i].Selected = true
				}
			},
			expected: []string{"nodejs"}, // Only unique requirements
		},
		{
			name: "no components selected",
			setupRegistry: func() {
				for i := range Registry {
					Registry[i].Selected = false
				}
			},
			expected: []string{},
		},
		{
			name: "only github-cli selected (no requirements)",
			setupRegistry: func() {
				for i := range Registry {
					Registry[i].Selected = false
				}
				Registry[2].Selected = true // github-cli has no requirements
			},
			expected: []string{},
		},
		{
			name: "only claude selected",
			setupRegistry: func() {
				for i := range Registry {
					Registry[i].Selected = false
				}
				Registry[0].Selected = true // claude requires nodejs
			},
			expected: []string{"nodejs"},
		},
		{
			name: "claude and claude-flow selected (duplicate nodejs requirement)",
			setupRegistry: func() {
				for i := range Registry {
					Registry[i].Selected = false
				}
				Registry[0].Selected = true // claude requires nodejs
				Registry[1].Selected = true // claude-flow requires nodejs
			},
			expected: []string{"nodejs"}, // Should be deduplicated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupRegistry()

			result := GetSystemRequirements()

			if len(tt.expected) == 0 {
				assert.Empty(t, result)
			} else {
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestRegistryStructure(t *testing.T) {
	t.Run("registry is not empty", func(t *testing.T) {
		assert.NotEmpty(t, Registry, "Registry should contain components")
		assert.GreaterOrEqual(t, len(Registry), 3, "Registry should have at least 3 components")
	})

	t.Run("all components have required fields", func(t *testing.T) {
		for i, comp := range Registry {
			assert.NotEmpty(t, comp.ID, "Component %d should have non-empty ID", i)
			assert.NotEmpty(t, comp.Name, "Component %d should have non-empty Name", i)
			assert.NotEmpty(t, comp.Description, "Component %d should have non-empty Description", i)
			assert.NotEmpty(t, comp.Emoji, "Component %d should have non-empty Emoji", i)
			assert.NotEmpty(t, comp.Installer, "Component %d should have non-empty Installer", i)
			// Requires can be empty, so we don't check it
		}
	})

	t.Run("component IDs are unique", func(t *testing.T) {
		idMap := make(map[string]bool)
		for _, comp := range Registry {
			assert.False(t, idMap[comp.ID], "Component ID %s should be unique", comp.ID)
			idMap[comp.ID] = true
		}
	})

	t.Run("component names are unique", func(t *testing.T) {
		nameMap := make(map[string]bool)
		for _, comp := range Registry {
			assert.False(t, nameMap[comp.Name], "Component name %s should be unique", comp.Name)
			nameMap[comp.Name] = true
		}
	})

	t.Run("dependency validation", func(t *testing.T) {
		// Check that all dependencies reference valid component IDs
		for _, comp := range Registry {
			for _, dep := range comp.DependsOn {
				found := false
				for _, other := range Registry {
					if other.ID == dep {
						found = true
						break
					}
				}
				assert.True(t, found, "Component %s depends on non-existent component %s", comp.ID, dep)
			}
		}
	})

	t.Run("claude-flow depends on claude", func(t *testing.T) {
		claudeFlow, err := GetByID("claude-flow")
		require.NoError(t, err)
		assert.Contains(t, claudeFlow.DependsOn, "claude", "claude-flow should depend on claude")
	})
}

func TestComponentDefaults(t *testing.T) {
	t.Run("all components selected by default", func(t *testing.T) {
		for _, comp := range Registry {
			assert.True(t, comp.Selected, "Component %s should be selected by default", comp.ID)
		}
	})

	t.Run("claude and claude-flow require nodejs", func(t *testing.T) {
		claude, err := GetByID("claude")
		require.NoError(t, err)
		assert.Contains(t, claude.Requires, "nodejs")

		claudeFlow, err := GetByID("claude-flow")
		require.NoError(t, err)
		assert.Contains(t, claudeFlow.Requires, "nodejs")
	})

	t.Run("github-cli has no special requirements", func(t *testing.T) {
		githubCli, err := GetByID("github-cli")
		require.NoError(t, err)
		assert.Empty(t, githubCli.Requires, "github-cli should have no special requirements")
	})
}

// Benchmark tests for performance validation
func BenchmarkGetByID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GetByID("claude")
	}
}

func BenchmarkGetSelected(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetSelected()
	}
}

func BenchmarkGetSelectedIDs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetSelectedIDs()
	}
}

func BenchmarkGetSystemRequirements(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetSystemRequirements()
	}
}

// Test helper functions for component manipulation
func TestComponentCopy(t *testing.T) {
	t.Run("component struct is properly copyable", func(t *testing.T) {
		original := Registry[0]
		copy := original
		
		// Modify copy
		copy.Selected = !copy.Selected
		
		// Original should be unchanged
		assert.NotEqual(t, original.Selected, copy.Selected)
	})
}

func TestRegistryModification(t *testing.T) {
	// Store original registry to restore later
	originalRegistry := make([]Component, len(Registry))
	copy(originalRegistry, Registry)
	defer func() {
		Registry = originalRegistry
	}()

	t.Run("registry can be safely modified", func(t *testing.T) {
		// Modify registry
		Registry[0].Selected = false
		
		// Changes should be reflected in GetSelected
		selected := GetSelected()
		found := false
		for _, comp := range selected {
			if comp.ID == Registry[0].ID {
				found = true
				break
			}
		}
		assert.False(t, found, "Modified component should not appear in GetSelected")
	})
}

// Edge case tests
func TestEdgeCases(t *testing.T) {
	t.Run("case sensitivity in GetByID", func(t *testing.T) {
		_, err := GetByID("CLAUDE")
		assert.Error(t, err, "GetByID should be case sensitive")
		
		_, err = GetByID("Claude")
		assert.Error(t, err, "GetByID should be case sensitive")
	})

	t.Run("whitespace in GetByID", func(t *testing.T) {
		_, err := GetByID(" claude ")
		assert.Error(t, err, "GetByID should not handle whitespace")
	})
}