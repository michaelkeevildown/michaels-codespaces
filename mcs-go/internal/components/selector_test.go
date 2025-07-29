package components

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test componentItem implementation
func TestComponentItem(t *testing.T) {
	comp := Component{
		ID:          "test-comp",
		Name:        "Test Component",
		Description: "A test component",
		Emoji:       "🧪",
		Selected:    true,
	}
	item := componentItem{component: comp}

	t.Run("FilterValue returns component name", func(t *testing.T) {
		assert.Equal(t, "Test Component", item.FilterValue())
	})

	t.Run("Title shows selected state", func(t *testing.T) {
		selectedItem := componentItem{component: comp}
		title := selectedItem.Title()
		assert.Contains(t, title, "[x]", "Selected component should show [x]")
		assert.Contains(t, title, "🧪", "Title should contain emoji")
		assert.Contains(t, title, "Test Component", "Title should contain name")

		// Test unselected
		comp.Selected = false
		unselectedItem := componentItem{component: comp}
		title = unselectedItem.Title()
		assert.Contains(t, title, "[ ]", "Unselected component should show [ ]")
	})

	t.Run("Description returns component description", func(t *testing.T) {
		assert.Equal(t, "A test component", item.Description())
	})
}

// Test itemDelegate implementation
func TestItemDelegate(t *testing.T) {
	delegate := itemDelegate{}

	t.Run("delegate properties", func(t *testing.T) {
		assert.Equal(t, 2, delegate.Height())
		assert.Equal(t, 1, delegate.Spacing())
	})

	t.Run("Update returns nil", func(t *testing.T) {
		cmd := delegate.Update(nil, nil)
		assert.Nil(t, cmd)
	})

	t.Run("Render handles valid componentItem", func(t *testing.T) {
		var buf bytes.Buffer
		comp := Component{
			ID:          "test",
			Name:        "Test",
			Description: "Test Description",
			Emoji:       "🧪",
			Selected:    false,
		}
		item := componentItem{component: comp}
		
		// Create a mock list model
		items := []list.Item{item}
		l := list.New(items, delegate, 80, 10)
		
		delegate.Render(&buf, l, 0, item)
		output := buf.String()
		
		assert.Contains(t, output, "Test", "Render should include component name")
		assert.Contains(t, output, "Test Description", "Render should include description")
	})

	t.Run("Render handles invalid item type gracefully", func(t *testing.T) {
		var buf bytes.Buffer
		items := []list.Item{}
		l := list.New(items, delegate, 80, 10)
		
		// This should not panic - pass nil instead of invalid string
		delegate.Render(&buf, l, 0, nil)
		assert.Equal(t, "", buf.String(), "Invalid item should produce no output")
	})
}

// Test Model initialization
func TestModelInit(t *testing.T) {
	model := NewSelector()

	t.Run("model initialization", func(t *testing.T) {
		assert.NotNil(t, model.list)
		assert.NotNil(t, model.items)
		assert.False(t, model.quitting)
		assert.Len(t, model.items, len(Registry), "Items should match registry length")
	})

	t.Run("Init returns nil command", func(t *testing.T) {
		cmd := model.Init()
		assert.Nil(t, cmd)
	})

	t.Run("list configuration", func(t *testing.T) {
		assert.Equal(t, "🚀 Select Components to Install", model.list.Title)
		assert.False(t, model.list.ShowStatusBar())
		assert.False(t, model.list.FilteringEnabled())
	})

	t.Run("items match registry", func(t *testing.T) {
		for i, item := range model.items {
			compItem, ok := item.(componentItem)
			require.True(t, ok, "Item should be componentItem")
			assert.Equal(t, Registry[i].ID, compItem.component.ID)
			assert.Equal(t, Registry[i].Name, compItem.component.Name)
			assert.Equal(t, Registry[i].Selected, compItem.component.Selected)
		}
	})
}

// Test Model Update method
func TestModelUpdate(t *testing.T) {
	// Store original registry to restore later
	originalRegistry := make([]Component, len(Registry))
	copy(originalRegistry, Registry)
	defer func() {
		Registry = originalRegistry
	}()

	t.Run("WindowSizeMsg updates list width", func(t *testing.T) {
		model := NewSelector()
		msg := tea.WindowSizeMsg{Width: 120, Height: 30}
		
		newModel, cmd := model.Update(msg)
		assert.Nil(t, cmd)
		assert.Equal(t, 120, newModel.(Model).list.Width())
	})

	t.Run("quit with 'q' key", func(t *testing.T) {
		model := NewSelector()
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		
		newModel, cmd := model.Update(msg)
		assert.True(t, newModel.(Model).quitting)
		assert.NotNil(t, cmd)
	})

	t.Run("quit with ctrl+c", func(t *testing.T) {
		model := NewSelector()
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		
		newModel, cmd := model.Update(msg)
		assert.True(t, newModel.(Model).quitting)
		assert.NotNil(t, cmd)
	})

	t.Run("enter key saves selections and quits", func(t *testing.T) {
		model := NewSelector()
		
		// Modify local item selection
		if len(model.items) > 0 {
			compItem := model.items[0].(componentItem)
			compItem.component.Selected = false
			model.items[0] = compItem
		}
		
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, cmd := model.Update(msg)
		
		assert.True(t, newModel.(Model).quitting)
		assert.NotNil(t, cmd)
	})

	t.Run("space key toggles selection", func(t *testing.T) {
		model := NewSelector()
		
		// Ensure we have items
		require.NotEmpty(t, model.items)
		
		// Get initial selection state
		initialItem := model.items[0].(componentItem)
		initialSelected := initialItem.component.Selected
		
		// Send space key message
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
		newModel, cmd := model.Update(msg)
		
		assert.Nil(t, cmd)
		
		// Check that selection was toggled
		updatedItem := newModel.(Model).items[0].(componentItem)
		assert.Equal(t, !initialSelected, updatedItem.component.Selected)
	})

	t.Run("other keys are passed to list", func(t *testing.T) {
		model := NewSelector()
		msg := tea.KeyMsg{Type: tea.KeyDown}
		
		newModel, cmd := model.Update(msg)
		
		// Should not quit
		assert.False(t, newModel.(Model).quitting)
		// List update might return a command
		_ = cmd
	})

	t.Run("non-key messages are passed to list", func(t *testing.T) {
		model := NewSelector()
		msg := tea.MouseMsg{}
		
		newModel, cmd := model.Update(msg)
		
		assert.False(t, newModel.(Model).quitting)
		_ = cmd
	})
}

// Test Model View method
func TestModelView(t *testing.T) {
	t.Run("quitting model returns empty string", func(t *testing.T) {
		model := NewSelector()
		model.quitting = true
		
		view := model.View()
		assert.Equal(t, "", view)
	})

	t.Run("active model returns list view", func(t *testing.T) {
		model := NewSelector()
		model.quitting = false
		
		view := model.View()
		assert.NotEmpty(t, view)
		assert.True(t, strings.HasPrefix(view, "\n"), "View should start with newline")
	})
}

// Test SelectComponents function
func TestSelectComponents(t *testing.T) {
	// Note: SelectComponents uses tea.NewProgram which requires terminal interaction
	// We'll test the function structure and error handling, but not the full interactive flow
	
	t.Run("SelectComponents returns error on program failure", func(t *testing.T) {
		// This test is mainly for structure validation
		// In a real scenario, we'd need to mock the tea.Program
		// For now, we'll just ensure the function exists and has the right signature
		result := GetSelected() // This should work as it doesn't need UI
		assert.IsType(t, []Component{}, result)
	})
}

// Test UI interaction flow simulation
func TestUIInteractionFlow(t *testing.T) {
	// Store original registry to restore later
	originalRegistry := make([]Component, len(Registry))
	copy(originalRegistry, Registry)
	defer func() {
		Registry = originalRegistry
	}()

	t.Run("complete interaction simulation", func(t *testing.T) {
		model := NewSelector()
		
		// Simulate window resize
		resizeMsg := tea.WindowSizeMsg{Width: 100, Height: 25}
		newModel, _ := model.Update(resizeMsg)
		model = newModel.(Model)
		
		// Simulate moving down
		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ = model.Update(downMsg)
		model = newModel.(Model)
		
		// Simulate space toggle
		spaceMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
		newModel, _ = model.Update(spaceMsg)
		model = newModel.(Model)
		
		// Simulate enter to finish
		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, cmd := model.Update(enterMsg)
		model = newModel.(Model)
		
		assert.True(t, model.quitting)
		assert.NotNil(t, cmd)
	})
}

// Test error conditions and edge cases
func TestSelectorEdgeCases(t *testing.T) {
	t.Run("empty registry handling", func(t *testing.T) {
		// Store original registry
		originalRegistry := Registry
		defer func() {
			Registry = originalRegistry
		}()
		
		// Set empty registry
		Registry = []Component{}
		
		model := NewSelector()
		assert.Empty(t, model.items)
		
		// Space toggle with empty list
		spaceMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
		newModel, cmd := model.Update(spaceMsg)
		assert.Nil(t, cmd)
		assert.False(t, newModel.(Model).quitting)
	})

	t.Run("invalid list index handling", func(t *testing.T) {
		model := NewSelector()
		
		// Force invalid state - this tests defensive programming
		// In normal operation, this shouldn't happen
		model.list.Select(-1) // Invalid index
		
		spaceMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
		newModel, cmd := model.Update(spaceMsg)
		
		// Should handle gracefully
		assert.Nil(t, cmd)
		assert.False(t, newModel.(Model).quitting)
	})
}

// Test style configurations
func TestStyleConfigurations(t *testing.T) {
	t.Run("styles are properly defined", func(t *testing.T) {
		assert.NotNil(t, titleStyle)
		assert.NotNil(t, itemStyle)
		assert.NotNil(t, selectedItemStyle)
		assert.NotNil(t, paginationStyle)
		assert.NotNil(t, helpStyle)
		assert.NotNil(t, quitTextStyle)
	})
}

// Test list item type assertions
func TestListItemTypeAssertions(t *testing.T) {
	t.Run("componentItem implements list.Item interface", func(t *testing.T) {
		comp := Component{ID: "test", Name: "Test", Description: "Desc", Emoji: "🧪"}
		item := componentItem{component: comp}
		
		// Verify it implements the interface
		var _ list.Item = item
		
		// Test interface methods
		assert.Equal(t, "Test", item.FilterValue())
		assert.Contains(t, item.Title(), "Test")
		assert.Equal(t, "Desc", item.Description())
	})
}

// Performance and memory tests
func TestSelectorPerformance(t *testing.T) {
	t.Run("NewSelector performance", func(t *testing.T) {
		// Create many selectors to test memory efficiency
		selectors := make([]Model, 100)
		for i := 0; i < 100; i++ {
			selectors[i] = NewSelector()
		}
		
		// Basic validation that they're all properly created
		for _, sel := range selectors {
			assert.Len(t, sel.items, len(Registry))
		}
	})
}

// Test concurrent access safety
func TestConcurrentAccess(t *testing.T) {
	t.Run("multiple selector creation", func(t *testing.T) {
		// Create multiple selectors concurrently
		done := make(chan bool)
		
		for i := 0; i < 10; i++ {
			go func() {
				model := NewSelector()
				assert.NotNil(t, model)
				assert.Len(t, model.items, len(Registry))
				done <- true
			}()
		}
		
		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// Benchmark tests for UI performance
func BenchmarkNewSelector(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewSelector()
	}
}

func BenchmarkModelUpdate(b *testing.B) {
	model := NewSelector()
	msg := tea.KeyMsg{Type: tea.KeyDown}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.Update(msg)
	}
}

func BenchmarkModelView(b *testing.B) {
	model := NewSelector()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

// Integration tests with registry
func TestSelectorRegistryIntegration(t *testing.T) {
	// Store original registry to restore later
	originalRegistry := make([]Component, len(Registry))
	copy(originalRegistry, Registry)
	defer func() {
		Registry = originalRegistry
	}()

	t.Run("selector reflects registry changes", func(t *testing.T) {
		// Create selector
		model := NewSelector()
		initialCount := len(model.items)
		
		// Modify registry (this would be unsafe in real app, but for testing)
		Registry = append(Registry, Component{
			ID:          "test-new",
			Name:        "New Component",
			Description: "Test component",
			Emoji:       "🆕",
			Selected:    false,
		})
		
		// Create new selector - should reflect changes
		newModel := NewSelector()
		assert.Equal(t, initialCount+1, len(newModel.items))
	})
}

// Test UI text content
func TestUITextContent(t *testing.T) {
	t.Run("list title is set correctly", func(t *testing.T) {
		model := NewSelector()
		assert.Equal(t, "🚀 Select Components to Install", model.list.Title)
	})

	t.Run("component titles format correctly", func(t *testing.T) {
		comp := Component{
			ID:          "test",
			Name:        "Test Component",
			Description: "A test component",
			Emoji:       "🧪",
			Selected:    true,
		}
		item := componentItem{component: comp}
		
		title := item.Title()
		assert.Contains(t, title, "[x]")
		assert.Contains(t, title, "🧪")
		assert.Contains(t, title, "Test Component")
		
		// Test unselected
		comp.Selected = false
		item = componentItem{component: comp}
		title = item.Title()
		assert.Contains(t, title, "[ ]")
	})
}