package docker

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

// TestIsMCSContainer tests the isMCSContainer function
func TestIsMCSContainer_Simple(t *testing.T) {
	tests := []struct {
		name      string
		container types.Container
		expected  bool
	}{
		{
			name: "MCS container by name prefix",
			container: types.Container{
				Names: []string{"/mcs-test-project", "/alias"},
			},
			expected: true,
		},
		{
			name: "MCS container by image",
			container: types.Container{
				Names: []string{"/some-container"},
				Image: "michaelkeevildown/claude-coder:latest",
			},
			expected: true,
		},
		{
			name: "MCS container by label",
			container: types.Container{
				Names: []string{"/labeled-container"},
				Labels: map[string]string{
					"mcs.managed": "true",
				},
			},
			expected: true,
		},
		{
			name: "Non-MCS container",
			container: types.Container{
				Names: []string{"/nginx-container"},
				Image: "nginx:latest",
			},
			expected: false,
		},
		{
			name: "Container with mcs.managed=false",
			container: types.Container{
				Names: []string{"/test-container"},
				Labels: map[string]string{
					"mcs.managed": "false",
				},
			},
			expected: false,
		},
		{
			name: "Container with empty names",
			container: types.Container{
				Names: []string{},
				Image: "test:latest",
			},
			expected: false,
		},
		{
			name: "Container with partial mcs name",
			container: types.Container{
				Names: []string{"/my-mcs-container"}, // Contains mcs but doesn't start with /mcs-
				Image: "test:latest",
			},
			expected: false,
		},
		{
			name: "Container with partial claude-coder image",
			container: types.Container{
				Names: []string{"/test-container"},
				Image: "other/claude-coder:latest", // Contains claude-coder but not michaelkeevildown
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMCSContainer(tt.container)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsMCSContainer_EdgeCases tests edge cases for the isMCSContainer function
func TestIsMCSContainer_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		container types.Container
		expected  bool
	}{
		{
			name: "Container with nil labels",
			container: types.Container{
				Names:  []string{"/test-container"},
				Labels: nil, // nil labels map
			},
			expected: false,
		},
		{
			name: "Container with multiple names, some matching",
			container: types.Container{
				Names: []string{"/regular-name", "/mcs-project", "/another-name"},
			},
			expected: true,
		},
		{
			name: "Container with mixed case image",
			container: types.Container{
				Names: []string{"/test-container"},
				Image: "michaelkeevildown/CLAUDE-CODER:latest", // Mixed case
			},
			expected: false, // Case sensitive matching, should not match
		},
		{
			name: "Container with image tag variations",
			container: types.Container{
				Names: []string{"/test-container"},
				Image: "michaelkeevildown/claude-coder", // No tag
			},
			expected: true,
		},
		{
			name: "Container with registry prefix",
			container: types.Container{
				Names: []string{"/test-container"},
				Image: "docker.io/michaelkeevildown/claude-coder:latest",
			},
			expected: true, // Should still match the core image name
		},
		{
			name: "Empty container",
			container: types.Container{
				Names:  []string{},
				Image:  "",
				Labels: nil,
			},
			expected: false,
		},
		{
			name: "Container with whitespace in name",
			container: types.Container{
				Names: []string{"/mcs-test project"}, // Space in name
			},
			expected: true, // Should still match the prefix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMCSContainer(tt.container)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsMCSContainer_Performance tests performance of the isMCSContainer function
func TestIsMCSContainer_Performance(t *testing.T) {
	// Create a variety of containers to test
	containers := []types.Container{
		{Names: []string{"/mcs-test-1"}},
		{Names: []string{"/regular-container-1"}},
		{Image: "michaelkeevildown/claude-coder:latest"},
		{Image: "nginx:latest"},
		{Labels: map[string]string{"mcs.managed": "true"}},
		{Labels: map[string]string{"other.label": "value"}},
		{Names: []string{"/mcs-test-2"}, Image: "test:latest"},
		{Names: []string{"/regular-2"}, Image: "redis:latest"},
		{Names: []string{"/mcs-test-3"}, Labels: map[string]string{"mcs.managed": "true"}},
		{Names: []string{"/regular-3"}, Labels: map[string]string{"other": "label"}},
	}

	// Test that the function handles multiple containers efficiently
	mcsCount := 0
	for _, container := range containers {
		if isMCSContainer(container) {
			mcsCount++
		}
	}

	// We expect 5 MCS containers in our test set (updated count)
	assert.Equal(t, 5, mcsCount)
}

// Benchmark the isMCSContainer function
func BenchmarkIsMCSContainer_Simple(b *testing.B) {
	containers := []types.Container{
		{
			Names: []string{"/mcs-test-project"},
		},
		{
			Image: "michaelkeevildown/claude-coder:latest",
			Names: []string{"/claude-container"},
		},
		{
			Names: []string{"/regular-container"},
			Labels: map[string]string{
				"mcs.managed": "true",
			},
		},
		{
			Names: []string{"/nginx-container"},
			Image: "nginx:latest",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, container := range containers {
			isMCSContainer(container)
		}
	}
}

// Test string matching logic
func TestStringMatching(t *testing.T) {
	t.Run("Name prefix matching", func(t *testing.T) {
		testCases := []struct {
			name     string
			expected bool
		}{
			{"/mcs-test", true},
			{"/mcs-", true},
			{"/mcs-project-123", true},
			{"/mcss-test", false}, // Extra 's'
			{"/test-mcs", false},  // mcs not at start
			{"mcs-test", false},   // No leading slash
			{"/MCS-test", false},  // Wrong case
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				container := types.Container{Names: []string{tc.name}}
				result := isMCSContainer(container)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("Image substring matching", func(t *testing.T) {
		testCases := []struct {
			image    string
			expected bool
		}{
			{"michaelkeevildown/claude-coder:latest", true},
			{"michaelkeevildown/claude-coder", true},
			{"docker.io/michaelkeevildown/claude-coder:v1.0", true},
			{"registry.com/michaelkeevildown/claude-coder:tag", true},
			{"other/claude-coder:latest", false},    // Wrong prefix
			{"michaelkeevildown/other:latest", false}, // Wrong suffix
			{"claude-coder:latest", false},          // No prefix
			{"", false},                             // Empty
		}

		for _, tc := range testCases {
			t.Run(tc.image, func(t *testing.T) {
				container := types.Container{
					Names: []string{"/test"},
					Image: tc.image,
				}
				result := isMCSContainer(container)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("Label exact matching", func(t *testing.T) {
		testCases := []struct {
			labels   map[string]string
			expected bool
		}{
			{map[string]string{"mcs.managed": "true"}, true},
			{map[string]string{"mcs.managed": "false"}, false},
			{map[string]string{"mcs.managed": "True"}, false}, // Case sensitive
			{map[string]string{"MCS.managed": "true"}, false}, // Case sensitive key
			{map[string]string{"other.label": "true"}, false}, // Wrong key
			{map[string]string{"mcs.managed": "yes"}, false},  // Wrong value
			{map[string]string{}, false},                      // Empty
			{nil, false},                                      // Nil
		}

		for i, tc := range testCases {
			t.Run(string(rune(i)), func(t *testing.T) {
				container := types.Container{
					Names:  []string{"/test"},
					Labels: tc.labels,
				}
				result := isMCSContainer(container)
				assert.Equal(t, tc.expected, result)
			})
		}
	})
}

// Test boundary conditions
func TestIsMCSContainer_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name      string
		container types.Container
		expected  bool
	}{
		{
			name: "Exactly '/mcs-' name",
			container: types.Container{
				Names: []string{"/mcs-"},
			},
			expected: true,
		},
		{
			name: "Very long MCS name",
			container: types.Container{
				Names: []string{"/mcs-" + string(make([]byte, 1000))}, // Very long name
			},
			expected: true,
		},
		{
			name: "Multiple labels with MCS label",
			container: types.Container{
				Names: []string{"/test"},
				Labels: map[string]string{
					"app.name":    "test",
					"version":     "1.0",
					"mcs.managed": "true",
					"other.key":   "value",
				},
			},
			expected: true,
		},
		{
			name: "All criteria match",
			container: types.Container{
				Names: []string{"/mcs-claude-project"},
				Image: "michaelkeevildown/claude-coder:latest",
				Labels: map[string]string{
					"mcs.managed": "true",
				},
			},
			expected: true, // Should match on any criterion
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMCSContainer(tt.container)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test that validates the function logic flow
func TestIsMCSContainer_LogicFlow(t *testing.T) {
	t.Run("Check name first", func(t *testing.T) {
		// Container with MCS name should return true immediately
		container := types.Container{
			Names: []string{"/mcs-test"},
			Image: "other:latest", // Non-MCS image
			Labels: map[string]string{
				"mcs.managed": "false", // Even with false managed label
			},
		}
		assert.True(t, isMCSContainer(container))
	})

	t.Run("Check image second", func(t *testing.T) {
		// Container with non-MCS name but MCS image should return true
		container := types.Container{
			Names: []string{"/other-name"},
			Image: "michaelkeevildown/claude-coder:latest", // MCS image
			Labels: map[string]string{
				"mcs.managed": "false", // Even with false managed label
			},
		}
		assert.True(t, isMCSContainer(container))
	})

	t.Run("Check labels last", func(t *testing.T) {
		// Container with non-MCS name and image but MCS label should return true
		container := types.Container{
			Names: []string{"/other-name"},
			Image: "other:latest", // Non-MCS image
			Labels: map[string]string{
				"mcs.managed": "true", // MCS managed label
			},
		}
		assert.True(t, isMCSContainer(container))
	})

	t.Run("All checks fail", func(t *testing.T) {
		// Container that fails all checks should return false
		container := types.Container{
			Names: []string{"/other-name"},
			Image: "other:latest",
			Labels: map[string]string{
				"mcs.managed": "false",
			},
		}
		assert.False(t, isMCSContainer(container))
	})
}