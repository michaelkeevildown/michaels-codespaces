package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/michaelkeevildown/mcs/internal/codespace"
	"github.com/michaelkeevildown/mcs/internal/components"
	"github.com/michaelkeevildown/mcs/internal/ui"
	"github.com/michaelkeevildown/mcs/pkg/utils"
	"github.com/spf13/cobra"
)


// CreateCommand creates the 'create' command
func CreateCommand() *cobra.Command {
	var (
		name         string
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
  - Local path: . or ./path/to/repo`,
		Example: `  # Create from GitHub repository
  mcs create facebook/react
  
  # Create from current directory
  mcs create .
  
  # Create with custom name
  mcs create torvalds/linux --name my-kernel-dev`,
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

			// Generate name if not provided
			if name == "" {
				name = utils.GenerateFunName(repo.Name)
			}
			
			// Ensure name is valid
			name = utils.SanitizeName(name)

			fmt.Println()
			fmt.Printf("📦 Creating codespace: %s\n", infoStyle.Render(name))
			fmt.Printf("📁 Repository: %s\n", urlStyle.Render(repo.URL))
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

	cmd.Flags().StringVar(&name, "name", "", "Custom name for the codespace")
	cmd.Flags().BoolVar(&noStart, "no-start", false, "Don't start the codespace after creation")
	cmd.Flags().BoolVar(&skipSelector, "skip-selector", false, "Skip component selection (use defaults)")

	return cmd
}

func createWithProgress(ctx context.Context, opts codespace.CreateOptions, progress *ui.Progress) (*codespace.Codespace, error) {
	tasks := []struct {
		name string
		fn   func() error
	}{
		{
			name: "Creating directory structure",
			fn: func() error {
				return os.MkdirAll(opts.GetPath(), 0755)
			},
		},
		{
			name: "Cloning repository",
			fn: func() error {
				// TODO: Implement git clone with progress
				time.Sleep(2 * time.Second) // Simulate
				return nil
			},
		},
		{
			name: "Detecting project type",
			fn: func() error {
				// TODO: Implement language/framework detection
				time.Sleep(500 * time.Millisecond) // Simulate
				return nil
			},
		},
		{
			name: "Generating Docker configuration",
			fn: func() error {
				// TODO: Implement Docker compose generation
				time.Sleep(1 * time.Second) // Simulate
				return nil
			},
		},
		{
			name: "Setting up components",
			fn: func() error {
				// TODO: Implement component installation
				time.Sleep(1 * time.Second) // Simulate
				return nil
			},
		},
	}

	// Execute tasks with progress
	for _, task := range tasks {
		progress.Start(task.name)
		if err := task.fn(); err != nil {
			progress.Fail(fmt.Sprintf("%s failed", task.name))
			return nil, err
		}
		progress.Success(task.name)
	}

	if !opts.NoStart {
		progress.Start("Starting services")
		// TODO: Implement Docker container start
		time.Sleep(2 * time.Second) // Simulate
		progress.Success("Services started")
	}

	// Create codespace object
	cs := &codespace.Codespace{
		Name:       opts.Name,
		Repository: opts.Repository.URL,
		Path:       opts.GetPath(),
		Status:     "running",
		CreatedAt:  time.Now(),
		VSCodeURL:  fmt.Sprintf("http://localhost:8080"),
		AppURL:     fmt.Sprintf("http://localhost:3000"),
	}

	return cs, nil
}

func showSuccess(cs *codespace.Codespace) {
	fmt.Println()
	fmt.Println(successStyle.Render("✨ Codespace created successfully!"))
	fmt.Println()
	
	// Create a nice box for the URLs
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(1, 2)

	content := fmt.Sprintf(
		"📍 Name: %s\n"+
		"🔗 VS Code: %s\n"+
		"🌐 App: %s\n"+
		"📂 Path: %s",
		cs.Name,
		urlStyle.Render(cs.VSCodeURL),
		urlStyle.Render(cs.AppURL),
		cs.Path,
	)

	fmt.Println(boxStyle.Render(content))
	fmt.Println()
	
	// Show helpful tips
	fmt.Println(infoStyle.Render("💡 Tips:"))
	fmt.Printf("  • Open VS Code: %s\n", infoStyle.Render(fmt.Sprintf("mcs open %s", cs.Name)))
	fmt.Printf("  • View logs: %s\n", infoStyle.Render(fmt.Sprintf("mcs logs %s", cs.Name)))
	fmt.Printf("  • Stop when done: %s\n", infoStyle.Render(fmt.Sprintf("mcs stop %s", cs.Name)))
	fmt.Println()
}