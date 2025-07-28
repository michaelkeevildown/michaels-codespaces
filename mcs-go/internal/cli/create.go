package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/michaelkeevildown/mcs/internal/codespace"
	"github.com/michaelkeevildown/mcs/internal/components"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/michaelkeevildown/mcs/pkg/utils"
	"github.com/spf13/cobra"
)


// CreateCommand creates the 'create' command
func CreateCommand() *cobra.Command {
	var (
		noStart      bool
		skipSelector bool
	)

	cmd := &cobra.Command{
		Use:   "create <repository-url>",
		Short: "🚀 Create a new codespace",
		Long: `Create a new codespace from a Git repository.

The repository can be specified as:
  - Full URL: https://github.com/user/repo
  - SSH URL: git@github.com:user/repo.git
  - Short form: user/repo (assumes GitHub)
  - Local path: . or ./path/to/repo

Codespace names are automatically generated from the repository owner and name.
If a collision occurs, a random suffix (e.g., 'happy-narwhal') will be added.`,
		Example: `  # Create from GitHub repository
  mcs create facebook/react
  # Creates: facebook-react
  
  # Create from SSH URL
  mcs create git@github.com:michaelkeevildown/michaels-codespaces.git
  # Creates: michaelkeevildown-michaels-codespaces
  
  # Create from current directory
  mcs create .`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			repoURL := args[0]

			// Show beautiful header
			ui.ShowHeader()

			// Create progress indicator
			progress := ui.NewProgress()
			
			// Parse repository URL
			progress.Start("Analyzing repository")
			repo, err := utils.ParseRepository(repoURL)
			if err != nil {
				progress.Fail("Invalid repository URL")
				return fmt.Errorf("invalid repository: %w", err)
			}
			progress.Success("Repository validated")

			// Generate unique name with collision detection
			baseDir := filepath.Join(utils.GetHomeDir(), "codespaces")
			checkExists := func(name string) bool {
				_, err := os.Stat(filepath.Join(baseDir, name))
				return !os.IsNotExist(err)
			}
			
			name := utils.GenerateUniqueCodespaceName(repo.Owner, repo.Name, checkExists)

			fmt.Println()
			fmt.Printf("📦 Creating codespace: %s\n", name)
			fmt.Printf("📁 Repository: %s\n", repo.URL)
			fmt.Println()

			// Component selection
			var selectedComponents []components.Component
			if !skipSelector {
				progress.Start("Component selection")
				selectedComponents, err = components.SelectComponents()
				if err != nil {
					progress.Fail("Component selection cancelled")
					return err
				}
				progress.Success(fmt.Sprintf("Selected %d components", len(selectedComponents)))
			} else {
				selectedComponents = components.GetSelected()
			}

			// Create codespace options
			opts := codespace.CreateOptions{
				Name:       name,
				Repository: repo,
				Components: selectedComponents,
				NoStart:    noStart,
			}

			// Create codespace with progress tracking
			cs, err := createWithProgress(ctx, opts, progress)
			if err != nil {
				return err
			}

			// Show success message
			showSuccess(cs)
			
			return nil
		},
	}

	// Note: --name flag has been removed - names are auto-generated from repository
	cmd.Flags().BoolVar(&noStart, "no-start", false, "Don't start the codespace after creation")
	cmd.Flags().BoolVar(&skipSelector, "skip-selector", false, "Skip component selection (use defaults)")

	return cmd
}

func createWithProgress(ctx context.Context, opts codespace.CreateOptions, progress *ui.Progress) (*codespace.Codespace, error) {
	// Create a new codespace manager
	manager := codespace.NewManager()
	
	// Track the last message to show success for each step
	var lastMessage string
	
	// Add progress callback to options
	opts.Progress = func(message string) {
		// If we had a previous message, mark it as successful
		if lastMessage != "" {
			progress.Success(lastMessage)
			// Create a new progress instance for the next step
			progress = ui.NewProgress()
		}
		
		// Update the progress UI with the new step
		progress.Start(message)
		lastMessage = message
	}
	
	// Call the actual Create method which does all the work
	cs, err := manager.Create(ctx, opts)
	if err != nil {
		progress.Fail("Failed to create codespace")
		return nil, err
	}
	
	// Mark the last step as successful
	if lastMessage != "" {
		progress.Success(lastMessage)
	}
	
	return cs, nil
}

func showSuccess(cs *codespace.Codespace) {
	fmt.Println()
	fmt.Println("✨ Codespace created successfully!")
	fmt.Println()
	
	// Calculate the maximum width needed
	lines := []struct {
		icon  string
		label string
		value string
	}{
		{"📍", "Name", cs.Name},
		{"🔗", "VS Code", cs.VSCodeURL},
		{"🔑", "Password", cs.Password},
		{"📂", "Path", cs.Path},
	}
	
	// Find the longest line
	maxWidth := 0
	for _, line := range lines {
		// Calculate: "│  icon label: value  │"
		lineWidth := 2 + len(line.icon) + 1 + len(line.label) + 2 + len(line.value) + 2 + 1
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}
	
	// Ensure minimum width
	if maxWidth < 50 {
		maxWidth = 50
	}
	
	// Create box elements
	boxLine := strings.Repeat("─", maxWidth-2)
	boxTop := "╭" + boxLine + "╮"
	boxBot := "╰" + boxLine + "╯"
	
	// Helper function to pad line to box width
	padLine := func(content string) string {
		padding := maxWidth - len(content) - 2
		if padding < 0 {
			padding = 0
		}
		return "│ " + content + strings.Repeat(" ", padding) + " │"
	}
	
	// Print the box
	fmt.Println(boxTop)
	fmt.Println(padLine(""))
	
	// Print each line
	for _, line := range lines {
		if line.value != "" { // Only show lines with values
			content := fmt.Sprintf("%s %s: %s", line.icon, line.label, line.value)
			fmt.Println(padLine(content))
		}
	}
	
	fmt.Println(padLine(""))
	fmt.Println(boxBot)
	fmt.Println()
	
	// Show helpful tips
	fmt.Println("💡 Tips:")
	fmt.Printf("  • Open VS Code: mcs open %s\n", cs.Name)
	fmt.Printf("  • View logs: mcs logs %s\n", cs.Name)
	fmt.Printf("  • Stop when done: mcs stop %s\n", cs.Name)
	fmt.Println()
}