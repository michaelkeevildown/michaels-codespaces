# MCS Go - Michael's Codespaces (Go Version)

A complete rewrite of Michael's Codespaces in Go, providing isolated, reproducible development environments optimized for AI agents and modern development workflows.

## 🚀 Quick Start

```bash
# Clone and install from source
git clone https://github.com/yourusername/mcs
cd mcs/mcs-go
./install.sh

# Create your first codespace
mcs create https://github.com/facebook/react

# List codespaces
mcs list

# Start a codespace
mcs start my-react-project
```

## 📦 Installation Philosophy

MCS follows a **source-first** approach:
- Primary method: Build from source for full transparency and control
- Fallback: Pre-built binaries for users without Go
- Updates via `git pull` + rebuild (same as shell version)

## ✨ Features

### Core Commands
- `mcs create <repo>` - Create a new codespace with interactive component selection
- `mcs list` - List all codespaces with beautiful table output
- `mcs start <name>` - Start a codespace
- `mcs stop <name>` - Stop a running codespace
- `mcs remove <name>` - Remove a codespace
- `mcs exec <name> [cmd]` - Execute commands in container
- `mcs logs <name>` - View container logs
- `mcs info <name>` - Show detailed information
- `mcs recover <name>` - Quick credential recovery
- `mcs reset-password <name>` - Reset VS Code password
- `mcs doctor` - System health check
- `mcs update` - Update MCS from source

### Key Improvements Over Shell Version
- **Single binary** - No bash dependencies
- **Beautiful TUI** - Interactive component selection with Bubble Tea
- **Better performance** - Compiled Go vs interpreted shell
- **Improved error handling** - Clear, actionable error messages
- **Cross-platform** - Works on Linux, macOS, and Windows

## 🛠️ Requirements

- Docker
- Git
- Go (for building from source, optional if using pre-built binaries)

## 🔄 Updating

```bash
# Update to latest version
mcs update

# Check for updates without installing
mcs update --check
```

## 🏗️ Architecture

```
mcs-go/
├── cmd/mcs/          # Main entry point
├── internal/
│   ├── cli/          # Command implementations
│   ├── codespace/    # Codespace management
│   ├── components/   # Component system
│   ├── docker/       # Docker integration
│   ├── git/          # Git operations
│   ├── ports/        # Port management
│   └── ui/           # Terminal UI components
├── assets/           # Embedded resources
└── install.sh        # Installation script
```

## 🎯 Design Principles

1. **Build from source** - Users compile locally for full control
2. **No magic** - Transparent operations, clear feedback
3. **User ownership** - Your hardware, your rules
4. **Minimal dependencies** - Just Docker and Git required
5. **Beautiful UX** - Terminal UI that's a joy to use

## 🤝 Contributing

This is a complete rewrite maintaining the philosophy of the original shell version while providing a better developer experience.

## 📄 License

Same as the original MCS project.