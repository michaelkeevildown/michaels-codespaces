package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	successMark   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
	failMark      = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗")
	spinnerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

// Progress represents a progress indicator
type Progress struct {
	currentTask string
	frame       int
	done        chan bool
}

// NewProgress creates a new progress indicator
func NewProgress() *Progress {
	return &Progress{
		done: make(chan bool),
	}
}

// Start begins showing progress for a task
func (p *Progress) Start(task string) {
	p.currentTask = task
	p.frame = 0
	
	// Clear the line and show initial state
	fmt.Printf("\r%s %s", spinnerStyle.Render(spinnerFrames[0]), task)
	
	// Start spinner in background
	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				p.frame = (p.frame + 1) % len(spinnerFrames)
				fmt.Printf("\r%s %s", spinnerStyle.Render(spinnerFrames[p.frame]), p.currentTask)
			case <-p.done:
				return
			}
		}
	}()
}

// Success marks the current task as successful
func (p *Progress) Success(message string) {
	p.done <- true
	time.Sleep(50 * time.Millisecond) // Brief pause to ensure spinner stops
	fmt.Printf("\r%s %s\n", successMark, message)
}

// Fail marks the current task as failed
func (p *Progress) Fail(message string) {
	p.done <- true
	time.Sleep(50 * time.Millisecond) // Brief pause to ensure spinner stops
	fmt.Printf("\r%s %s\n", failMark, message)
}