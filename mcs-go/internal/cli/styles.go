package cli

import "github.com/charmbracelet/lipgloss"

// Common styles used across CLI commands
var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	urlStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Underline(true)
	sectionStyle = lipgloss.NewStyle().Bold(true).Underline(true)
	boldStyle    = lipgloss.NewStyle().Bold(true)
	runningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	stoppedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))
)