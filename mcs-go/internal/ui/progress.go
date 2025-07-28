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
	paused      chan bool
	resumed     chan bool
	isRunning   bool
}

// NewProgress creates a new progress indicator
func NewProgress() *Progress {
	return &Progress{
		done:    make(chan bool),
		paused:  make(chan bool),
		resumed: make(chan bool),
	}
}

// Start begins showing progress for a task
func (p *Progress) Start(task string) {
	p.currentTask = task
	p.frame = 0
	p.isRunning = true

	// Clear the entire line and show initial state
	fmt.Printf("\r\033[K%s %s", spinnerStyle.Render(spinnerFrames[0]), task)

	// Start spinner in background
	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if p.isRunning {
					p.frame = (p.frame + 1) % len(spinnerFrames)
					fmt.Printf("\r\033[K%s %s", spinnerStyle.Render(spinnerFrames[p.frame]), p.currentTask)
				}
			case <-p.paused:
				// Clear the line when paused
				fmt.Printf("\r\033[K")
				p.isRunning = false
			case <-p.resumed:
				// Resume showing the spinner
				p.isRunning = true
				fmt.Printf("\r\033[K%s %s", spinnerStyle.Render(spinnerFrames[p.frame]), p.currentTask)
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
	fmt.Printf("\r\033[K%s %s\n", successMark, message)
}

// Fail marks the current task as failed
func (p *Progress) Fail(message string) {
	p.done <- true
	time.Sleep(50 * time.Millisecond) // Brief pause to ensure spinner stops
	fmt.Printf("\r\033[K%s %s\n", failMark, message)
}

// Update updates the current task message
func (p *Progress) Update(task string) {
	p.currentTask = task
	if p.isRunning {
		fmt.Printf("\r\033[K%s %s", spinnerStyle.Render(spinnerFrames[p.frame]), task)
	}
}

// Stop temporarily stops the spinner without marking it as complete
func (p *Progress) Stop() {
	if p.isRunning {
		p.paused <- true
		time.Sleep(50 * time.Millisecond) // Brief pause to ensure spinner stops
	}
}

// Resume restarts the spinner after it was stopped
func (p *Progress) Resume() {
	if !p.isRunning {
		p.resumed <- true
		time.Sleep(50 * time.Millisecond) // Brief pause to ensure spinner resumes
	}
}
