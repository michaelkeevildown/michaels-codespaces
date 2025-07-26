package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)
)

// ShowHeader displays the MCS ASCII art header
func ShowHeader() {
	header := `
 __  __  ___ _____ 
|  \/  |/ __/ ____|
| |\/| | (__\__ \  
|_|  |_|\___|___/  
                   
Michael's Codespaces`

	fmt.Println(headerStyle.Render(header))
	fmt.Println()
}