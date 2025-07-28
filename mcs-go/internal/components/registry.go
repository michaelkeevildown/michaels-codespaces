package components

import (
	"fmt"
)

// Component represents an installable component
type Component struct {
	ID          string
	Name        string
	Description string
	Emoji       string
	Selected    bool
	Installer   string
	DependsOn   []string
	Requires    []string // System requirements (e.g., "nodejs", "python")
}

// Registry holds all available components
var Registry = []Component{
	{
		ID:          "claude",
		Name:        "Claude Code",
		Description: "Anthropic's Claude AI coding assistant - your AI pair programmer",
		Emoji:       "🤖",
		Selected:    true,
		Installer:   "claude.sh",
		Requires:    []string{"nodejs"},
	},
	{
		ID:          "claude-flow",
		Name:        "Claude Flow",
		Description: "AI swarm orchestration and workflow automation",
		Emoji:       "🌊",
		Selected:    true,
		Installer:   "claude-flow.sh",
		DependsOn:   []string{"claude"},
		Requires:    []string{"nodejs"},
	},
	{
		ID:          "github-cli",
		Name:        "GitHub CLI",
		Description: "Command-line interface for GitHub with token authentication",
		Emoji:       "🐙",
		Selected:    true,
		Installer:   "github-cli.sh",
		Requires:    []string{}, // No special requirements
	},
}

// GetByID returns a component by its ID
func GetByID(id string) (*Component, error) {
	for _, c := range Registry {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("component not found: %s", id)
}

// GetSelected returns all selected components
func GetSelected() []Component {
	var selected []Component
	for _, c := range Registry {
		if c.Selected {
			selected = append(selected, c)
		}
	}
	return selected
}

// GetSelectedIDs returns IDs of selected components
func GetSelectedIDs() []string {
	var ids []string
	for _, c := range Registry {
		if c.Selected {
			ids = append(ids, c.ID)
		}
	}
	return ids
}

// GetSystemRequirements returns unique system requirements for selected components
func GetSystemRequirements() []string {
	reqMap := make(map[string]bool)
	
	for _, c := range Registry {
		if c.Selected {
			for _, req := range c.Requires {
				reqMap[req] = true
			}
		}
	}
	
	var requirements []string
	for req := range reqMap {
		requirements = append(requirements, req)
	}
	
	return requirements
}