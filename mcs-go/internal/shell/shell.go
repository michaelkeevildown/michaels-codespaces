package shell

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents shell configuration
type Config struct {
	FilePath string
	Shell    string
}

// ShellType represents different shell types
type ShellType string

const (
	Bash ShellType = "bash"
	Zsh  ShellType = "zsh"
	Sh   ShellType = "sh"
)

// DetectShell detects the current shell
func DetectShell() ShellType {
	// Check environment variables first
	if os.Getenv("BASH_VERSION") != "" {
		return Bash
	}
	if os.Getenv("ZSH_VERSION") != "" {
		return Zsh
	}
	
	// Check SHELL environment variable
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "bash") {
		return Bash
	}
	if strings.Contains(shell, "zsh") {
		return Zsh
	}
	
	// Default to sh
	return Sh
}

// GetShellConfigs returns all shell configuration files
func GetShellConfigs() []Config {
	homeDir := os.Getenv("HOME")
	configs := []Config{}
	
	// Common shell config files
	shellFiles := map[string]ShellType{
		".bashrc":       Bash,
		".bash_profile": Bash,
		".zshrc":        Zsh,
		".zprofile":     Zsh,
		".profile":      Sh,
	}
	
	for file, shellType := range shellFiles {
		path := filepath.Join(homeDir, file)
		if _, err := os.Stat(path); err == nil {
			configs = append(configs, Config{
				FilePath: path,
				Shell:    string(shellType),
			})
		}
	}
	
	return configs
}

// CleanConfig removes lines matching patterns from a shell configuration file
func CleanConfig(configFile string, patterns []string) error {
	// Read the file
	content, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, nothing to clean
		}
		return err
	}
	
	lines := strings.Split(string(content), "\n")
	cleanedLines := []string{}
	skipNext := false
	
	for i, line := range lines {
		// Check if line matches any pattern
		shouldSkip := false
		for _, pattern := range patterns {
			if strings.Contains(line, pattern) {
				shouldSkip = true
				
				// Check if we should skip the next line too
				if i+1 < len(lines) {
					nextLine := strings.TrimSpace(lines[i+1])
					if strings.HasPrefix(nextLine, "alias ") ||
						strings.HasPrefix(nextLine, "export ") ||
						strings.HasPrefix(nextLine, "source ") {
						skipNext = true
					}
				}
				break
			}
		}
		
		if skipNext {
			skipNext = false
			continue
		}
		
		if !shouldSkip {
			cleanedLines = append(cleanedLines, line)
		}
	}
	
	// Write back the cleaned content
	cleanedContent := strings.Join(cleanedLines, "\n")
	return os.WriteFile(configFile, []byte(cleanedContent), 0644)
}

// AddToConfig adds lines to a shell configuration file
func AddToConfig(configFile string, comment string, lines []string) error {
	// Ensure file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create file with basic shell header
		header := "#!/bin/bash\n" // Safe default
		if strings.Contains(configFile, "zsh") {
			header = "#!/bin/zsh\n"
		}
		if err := os.WriteFile(configFile, []byte(header), 0644); err != nil {
			return err
		}
	}
	
	// Read existing content
	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	
	// Check if any of the lines already exist
	existingContent := string(content)
	for _, line := range lines {
		if strings.Contains(existingContent, line) {
			return nil // Already configured
		}
	}
	
	// Append new configuration
	file, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	
	// Add spacing
	writer.WriteString("\n")
	
	// Add comment
	if comment != "" {
		writer.WriteString(fmt.Sprintf("# %s\n", comment))
	}
	
	// Add lines
	for _, line := range lines {
		writer.WriteString(line + "\n")
	}
	
	return writer.Flush()
}

// RemoveMCSConfig removes all MCS-related configuration from shell files
func RemoveMCSConfig() error {
	patterns := []string{
		"Michael's Codespaces",
		"MCS aliases",
		"/.mcs/bin",
		"# Codespace:",
		"mcs completion",
		"export MCS_",
		"alias mcs=",
	}
	
	configs := GetShellConfigs()
	errors := []error{}
	
	for _, config := range configs {
		if err := CleanConfig(config.FilePath, patterns); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", config.FilePath, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to clean some configs: %v", errors)
	}
	
	return nil
}

// AddMCSToPath adds MCS binary directory to PATH in shell configs
func AddMCSToPath(binDir string) (int, error) {
	comment := "MCS - Michael's Codespaces"
	pathLine := fmt.Sprintf("export PATH=\"%s:$PATH\"", binDir)
	
	configs := GetShellConfigs()
	updated := 0
	
	for _, config := range configs {
		// Check if already in PATH
		content, err := os.ReadFile(config.FilePath)
		if err == nil && strings.Contains(string(content), binDir) {
			continue // Already configured
		}
		
		if err := AddToConfig(config.FilePath, comment, []string{pathLine}); err == nil {
			updated++
		}
	}
	
	return updated, nil
}

// SourceConfig returns the command to source a shell config file
func SourceConfig(shell ShellType) string {
	homeDir := os.Getenv("HOME")
	
	switch shell {
	case Bash:
		return fmt.Sprintf("source %s/.bashrc", homeDir)
	case Zsh:
		return fmt.Sprintf("source %s/.zshrc", homeDir)
	default:
		return fmt.Sprintf("source %s/.profile", homeDir)
	}
}

// GetCompletionScript generates shell completion script for the given shell
func GetCompletionScript(shell ShellType, command string) string {
	switch shell {
	case Bash:
		return fmt.Sprintf("source <(%s completion bash)", command)
	case Zsh:
		return fmt.Sprintf("source <(%s completion zsh)", command)
	default:
		return ""
	}
}

// IsLoginShell checks if the current shell is a login shell
func IsLoginShell() bool {
	// Check if $0 starts with '-'
	if len(os.Args) > 0 && strings.HasPrefix(os.Args[0], "-") {
		return true
	}
	
	// Check common login shell indicators
	return os.Getenv("SHLVL") == "1"
}