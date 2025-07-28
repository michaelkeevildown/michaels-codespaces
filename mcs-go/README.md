# MCS - Michael's Codespaces (Go Version)

🚀 A delightful CLI for managing AI-powered development environments.

## Features

- 🎨 **Beautiful CLI** - Interactive component selection with Bubble Tea
- 🚀 **Fast & Reliable** - Single Go binary, no dependencies
- 🤖 **AI-Optimized** - Built for running multiple AI agents
- 📦 **Components** - Claude, Claude Flow, GitHub CLI, and more
- 🎯 **Smart Defaults** - Detects project type and suggests configuration
- ✨ **Delightful UX** - Progress indicators, emoji, helpful error messages

## Installation

### Prerequisites

- Docker (for running codespaces)
- Git
- Go 1.21 or later (optional - for building from source)

### Quick Install

```bash
# Install with one command
curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/main/mcs-go/install.sh | bash

# Or for a specific branch:
curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/main/mcs-go/install.sh | MCS_BRANCH=feature-branch bash
```

The installer will:
- Build from Go source if Go is available
- Or download a pre-built binary as fallback
- Set up your PATH automatically
- Configure shell completions

### Building from Source (Alternative)

If you prefer to build manually:

```bash
# Clone the repository
git clone https://github.com/michaelkeevildown/michaels-codespaces.git
cd michaels-codespaces/mcs-go

# Download dependencies
go mod download

# Build and install
make install

# Add to PATH (if not already done)
export PATH="$HOME/.mcs/bin:$PATH"
```

## Usage

### Create a Codespace

```bash
# From GitHub repository
mcs create facebook/react

# From current directory
mcs create .

# With custom name
mcs create torvalds/linux --name kernel-dev
```

### List Codespaces

```bash
mcs list
```

### Start/Stop Codespaces

```bash
mcs start my-project
mcs stop my-project
```

### Remove a Codespace

```bash
mcs remove my-project
```

## Component Selection

When creating a codespace, you'll see an interactive component selector:

```
🚀 Select Components to Install

> [x] 🤖 Claude Code
      Anthropic's Claude AI coding assistant - your AI pair programmer

  [x] 🌊 Claude Flow  
      AI swarm orchestration and workflow automation

  [x] 🐙 GitHub CLI
      Command-line interface for GitHub with token authentication

[Space: Toggle] [Enter: Confirm] [q: Cancel]
```

## Development

### Project Structure

```
mcs-go/
├── cmd/mcs/          # CLI entry point
├── internal/         # Internal packages
│   ├── cli/          # Command implementations
│   ├── codespace/    # Core codespace logic
│   ├── components/   # Component registry and selector
│   ├── docker/       # Docker integration
│   └── ui/           # Terminal UI components
├── pkg/              # Public packages
│   ├── errors/       # Error handling
│   └── utils/        # Utilities
└── Makefile          # Build automation
```

### Building

```bash
# Standard build
make build

# Development build (with race detector)
make dev

# Build for all platforms
make build-all

# Run tests
make test

# Clean build artifacts
make clean
```

### Adding New Components

Edit `internal/components/registry.go`:

```go
{
    ID:          "new-component",
    Name:        "New Component", 
    Description: "Description of your component",
    Emoji:       "🎯",
    Selected:    false,
    Installer:   "new-component.sh",
}
```

## Architecture

MCS is built with:

- **[Cobra](https://github.com/spf13/cobra)** - CLI framework
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - Terminal UI
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** - Styling
- **[Docker SDK](https://github.com/docker/docker)** - Container management
- **[go-git](https://github.com/go-git/go-git)** - Git operations

## Philosophy

From the WHY.md:

> Control your infrastructure. Run AI agents without constraints.
> Your hardware, your rules, your freedom.

MCS embodies this philosophy by:
- Being a single binary with no runtime dependencies
- Working on any Linux system with Docker
- Giving you complete control over your development environment
- Making it delightful to use

## License

MIT