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
	},
	{
		ID:          "claude-flow",
		Name:        "Claude Flow",
		Description: "AI swarm orchestration and workflow automation",
		Emoji:       "🌊",
		Selected:    true,
		Installer:   "claude-flow.sh",
		DependsOn:   []string{"claude"},
	},
	{
		ID:          "github-cli",
		Name:        "GitHub CLI",
		Description: "Command-line interface for GitHub with token authentication",
		Emoji:       "🐙",
		Selected:    true,
		Installer:   "github-cli.sh",
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