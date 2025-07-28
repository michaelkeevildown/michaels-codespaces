package cli

import (
	"bufio"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// TerminalReader wraps a reader and its underlying file (if any) for proper cleanup
type TerminalReader struct {
	reader *bufio.Reader
	file   *os.File // nil for stdin, non-nil for /dev/tty
}

// ReadString reads a line from the terminal
func (tr *TerminalReader) ReadString(delim byte) (string, error) {
	return tr.reader.ReadString(delim)
}

// Close closes the underlying file if it's not stdin
func (tr *TerminalReader) Close() error {
	if tr.file != nil {
		return tr.file.Close()
	}
	return nil
}

// getTerminalReader returns a terminal-aware reader for user input.
// It handles cases where stdin is not a terminal (e.g., piped input)
// by attempting to read directly from /dev/tty.
// Caller must call Close() on the returned reader when done.
func getTerminalReader() (*TerminalReader, error) {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(syscall.Stdin)) {
		// stdin is not a terminal (e.g., piped input)
		// Try to open /dev/tty directly to read from the actual terminal
		tty, err := os.Open("/dev/tty")
		if err != nil {
			return nil, err
		}
		return &TerminalReader{
			reader: bufio.NewReader(tty),
			file:   tty,
		}, nil
	}
	
	// Normal interactive mode
	return &TerminalReader{
		reader: bufio.NewReader(os.Stdin),
		file:   nil, // stdin doesn't need to be closed
	}, nil
}

// readUserInput is a convenience function that reads a line of input from the terminal
// and handles non-interactive scenarios. Returns the trimmed input and whether input was read.
func readUserInput() (string, bool) {
	reader, err := getTerminalReader()
	if err != nil {
		// Can't get user input in non-interactive mode
		return "", false
	}
	defer reader.Close()
	
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", false
	}
	
	return strings.TrimSpace(input), true
}