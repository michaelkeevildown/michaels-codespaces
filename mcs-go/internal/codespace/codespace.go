package codespace

import (
	"path/filepath"
	"time"

	"github.com/michaelkeevildown/mcs/internal/components"
	"github.com/michaelkeevildown/mcs/pkg/utils"
)

// Codespace represents a development environment
type Codespace struct {
	Name       string
	Repository string
	Path       string
	Status     string // running, stopped, error
	CreatedAt  time.Time
	VSCodeURL  string
	AppURL     string
	Components []string
	Language   string
	Password   string
	VSCodePort int
}

// ProgressFunc is a callback for reporting progress
type ProgressFunc func(message string)

// CreateOptions holds options for creating a codespace
type CreateOptions struct {
	Name       string
	Repository *utils.Repository
	Components []components.Component
	NoStart    bool
	CloneDepth int
	Progress   ProgressFunc
}

// GetPath returns the full path for the codespace
func (o CreateOptions) GetPath() string {
	home := utils.GetHomeDir()
	return filepath.Join(home, "codespaces", o.Name)
}

// Manager handles codespace operations
type Manager struct {
	baseDir string
}

// NewManager creates a new codespace manager
func NewManager() *Manager {
	return &Manager{
		baseDir: filepath.Join(utils.GetHomeDir(), "codespaces"),
	}
}

