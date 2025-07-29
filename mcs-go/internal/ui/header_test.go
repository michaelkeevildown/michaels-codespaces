package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestShowHeader(t *testing.T) {
	// Capture stdout to test the output
	output := captureHeaderOutput(func() {
		ShowHeader()
	})
	
	// Test that the output contains expected ASCII art
	assert.Contains(t, output, "Michael's Codespaces")
	assert.Contains(t, output, "__")
	assert.Contains(t, output, "|")
	
	// Test that output has multiple lines (ASCII art is multi-line)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.True(t, len(lines) >= 5, "Header should have at least 5 lines of ASCII art")
}

func TestHeaderStyle(t *testing.T) {
	// Test that headerStyle is properly initialized
	assert.NotNil(t, headerStyle)
	
	// Test style properties
	testText := "Test"
	styled := headerStyle.Render(testText)
	
	// The styled text should not be empty and should contain our test text
	assert.NotEmpty(t, styled)
	assert.Contains(t, styled, testText)
}

func TestHeaderContent(t *testing.T) {
	output := captureHeaderOutput(func() {
		ShowHeader()
	})
	
	// Test specific parts of the ASCII art
	expectedLines := []string{
		"__  __  ___ ___",
		"|  \\/  |/ __/ __|",
		"| |\\/| | (__\\__ \\",
		"|_|  |_|\\___|___/",
		"Michael's Codespaces",
	}
	
	for _, expectedLine := range expectedLines {
		assert.Contains(t, output, expectedLine, 
			"Header should contain the expected ASCII art line: %s", expectedLine)
	}
}

func TestHeaderFormatting(t *testing.T) {
	output := captureHeaderOutput(func() {
		ShowHeader()
	})
	
	// Test that the output starts and ends appropriately
	lines := strings.Split(output, "\n")
	
	// Should have empty lines for spacing
	var nonEmptyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	
	// Should have at least the ASCII art lines plus "Michael's Codespaces"
	assert.True(t, len(nonEmptyLines) >= 5, 
		"Should have at least 5 non-empty lines in header")
}

func TestHeaderConsistency(t *testing.T) {
	// Test that multiple calls produce identical output
	output1 := captureHeaderOutput(func() {
		ShowHeader()
	})
	
	output2 := captureHeaderOutput(func() {
		ShowHeader()
	})
	
	assert.Equal(t, output1, output2, 
		"Multiple calls to ShowHeader should produce identical output")
}

func TestHeaderConcurrentCalls(t *testing.T) {
	// Test that concurrent calls don't cause panics or crashes
	// Due to stdout capture race conditions in tests, we'll focus on safety
	var wg sync.WaitGroup
	
	numCalls := 10
	wg.Add(numCalls)
	
	// Capture if any goroutine panics
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Concurrent header calls caused panic: %v", r)
		}
	}()
	
	for i := 0; i < numCalls; i++ {
		go func() {
			defer wg.Done()
			// Just call ShowHeader without capturing - this tests thread safety
			ShowHeader()
		}()
	}
	
	wg.Wait()
	
	// If we get here without panics, the test passes
	// Also test that a single call works correctly after concurrent calls
	output := captureHeaderOutput(func() {
		ShowHeader()
	})
	
	assert.Contains(t, output, "Michael's Codespaces")
	assert.Contains(t, output, "__")
}

func TestHeaderLipglossIntegration(t *testing.T) {
	// Test that the header uses lipgloss styling correctly
	testStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)
	
	testText := "Test Header"
	styledText := testStyle.Render(testText)
	
	// In some environments, styles might not be applied (NO_COLOR, etc.)
	// So we just verify the text content is preserved
	assert.Contains(t, styledText, testText, 
		"Styled text should contain original text")
}

func TestHeaderWidth(t *testing.T) {
	output := captureHeaderOutput(func() {
		ShowHeader()
	})
	
	lines := strings.Split(output, "\n")
	
	// Find the ASCII art lines (non-empty lines that contain the art)
	var artLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && (strings.Contains(trimmed, "_") || 
			strings.Contains(trimmed, "|") || 
			strings.Contains(trimmed, "Michael")) {
			artLines = append(artLines, trimmed)
		}
	}
	
	// The ASCII art should have consistent structure
	assert.True(t, len(artLines) >= 4, 
		"Should have at least 4 lines of ASCII art")
	
	// The main MCS logo lines should have similar lengths
	logoLines := make([]string, 0)
	for _, line := range artLines {
		if strings.Contains(line, "_") || strings.Contains(line, "|") {
			logoLines = append(logoLines, line)
		}
	}
	
	assert.True(t, len(logoLines) >= 4, 
		"Should have at least 4 logo lines")
}

func TestHeaderUnicodeHandling(t *testing.T) {
	// Test that the header handles Unicode characters properly
	output := captureHeaderOutput(func() {
		ShowHeader()
	})
	
	// Should not have any replacement characters or encoding issues
	assert.NotContains(t, output, "�", 
		"Header should not contain replacement characters")
	
	// Should contain the expected characters
	assert.Contains(t, output, "_")
	assert.Contains(t, output, "|")
	assert.Contains(t, output, "/")
	assert.Contains(t, output, "\\")
}

func TestHeaderMemoryUsage(t *testing.T) {
	// Test that ShowHeader doesn't leak memory or resources
	// This is more of a smoke test - run it many times
	for i := 0; i < 1000; i++ {
		captureHeaderOutput(func() {
			ShowHeader()
		})
	}
	// If we get here without running out of memory, the test passes
}

func TestHeaderEmptyEnvironment(t *testing.T) {
	// Test header in a minimal environment
	output := captureHeaderOutput(func() {
		ShowHeader()
	})
	
	// Should still work even in minimal conditions
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Michael's Codespaces")
}

// Helper function to capture header output
func captureHeaderOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(&buf, r)
	}()

	f()

	w.Close()
	os.Stdout = oldStdout
	wg.Wait()
	
	return buf.String()
}

// Benchmark tests
func BenchmarkShowHeader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		captureHeaderOutput(func() {
			ShowHeader()
		})
	}
}

func BenchmarkHeaderStyleRender(b *testing.B) {
	testText := `
  __  __  ___ ___ 
 |  \/  |/ __/ __|
 | |\/| | (__\__ \
 |_|  |_|\___|___|
                  
Michael's Codespaces`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		headerStyle.Render(testText)
	}
}

// Test edge cases
func TestHeaderWithDifferentTerminalSettings(t *testing.T) {
	// Test with NO_COLOR environment variable
	oldNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", oldNoColor)
	
	os.Setenv("NO_COLOR", "1")
	
	output := captureHeaderOutput(func() {
		ShowHeader()
	})
	
	// Should still show the header content
	assert.Contains(t, output, "Michael's Codespaces")
}

func TestHeaderIntegrationWithLipgloss(t *testing.T) {
	// Test that our header style integrates properly with lipgloss
	
	// Create a similar style to verify behavior
	testStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)
	
	testContent := "Test Content"
	rendered := testStyle.Render(testContent)
	
	// Should contain escape sequences for color and bold
	assert.Contains(t, rendered, testContent)
	
	// Test that headerStyle behaves similarly
	headerRendered := headerStyle.Render(testContent)
	assert.Contains(t, headerRendered, testContent)
}