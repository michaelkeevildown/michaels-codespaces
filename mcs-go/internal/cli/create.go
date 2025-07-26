package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

	// Removed --name flag as names are now auto-generated from repository
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

	// Generate secure password
	password := utils.GenerateSecurePassword()
	
	// Create codespace object
	cs := &codespace.Codespace{
		Name:       opts.Name,
		Repository: opts.Repository.URL,
		Path:       opts.GetPath(),
		Status:     "running",
		CreatedAt:  time.Now(),
		VSCodeURL:  fmt.Sprintf("http://localhost:8080"),
		AppURL:     fmt.Sprintf("http://localhost:3000"),
		Password:   password,
		VSCodePort: 8080,
	}

	return cs, nil
}

func showSuccess(cs *codespace.Codespace) {
	fmt.Println()
	fmt.Println(successStyle.Render("✨ Codespace created successfully!"))
	fmt.Println()
	
	// Create the perfect box format
	boxTop := "╭─────────────────────────────────────────────╮"
	boxMid := "│                                             │"
	boxBot := "╰─────────────────────────────────────────────╯"
	
	// Format content with proper spacing
	nameLine := fmt.Sprintf("│  📍 Name: %-33s │", cs.Name)
	vscLine := fmt.Sprintf("│  🔗 VS Code: %-30s │", cs.VSCodeURL)
	pwdLine := fmt.Sprintf("│  🔑 Password: %-29s │", cs.Password)
	pathLine := fmt.Sprintf("│  📂 Path: %-33s │", cs.Path)
	
	// Print the box
	fmt.Println(boxTop)
	fmt.Println(boxMid)
	fmt.Println(nameLine)
	fmt.Println(vscLine)
	fmt.Println(pwdLine)
	fmt.Println(pathLine)
	fmt.Println(boxMid)
	fmt.Println(boxBot)
	fmt.Println()
	
	// Show helpful tips
	fmt.Println("💡 Tips:")
	fmt.Printf("  • Open VS Code: mcs open %s\n", cs.Name)
	fmt.Printf("  • View logs: mcs logs %s\n", cs.Name)
	fmt.Printf("  • Stop when done: mcs stop %s\n", cs.Name)
	fmt.Println()
}