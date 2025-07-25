# Michael's Codespaces - Self-Hosted Development Environments

## Overview

This repository provides a complete solution for creating self-hosted development environments that replicate GitHub Codespaces functionality on your own infrastructure. It enables you to spin up isolated, browser-accessible VS Code environments for any GitHub repository, with full persistence and security.

## Core Objectives

- **Complete Environment Setup**: Automatically provision Ubuntu VMs with all required development tools
- **Repository Isolation**: Each GitHub repository runs in its own Docker container with complete isolation
- **Browser-Based Development**: Access VS Code through any web browser from anywhere
- **Persistence**: All code and container data persists across restarts
- **Security**: Secure SSH access, isolated containers, and proper authentication
- **Flexibility**: Support for both local VMs (with Cloudflare tunnels) and cloud VMs (with public IPs)

## Architecture

### Base System
- **Host OS**: Ubuntu (VM or bare metal)
- **Container Runtime**: Docker with docker-compose
- **Development Environment**: VS Code Server in containers
- **Shell**: Zsh with Oh My Zsh
- **Networking**: Port-based isolation with optional Cloudflare tunnels

### Directory Structure
```
~/codespaces/
├── shared/           # Shared resources across all codespaces
│   ├── templates/    # Template files for new codespaces
│   ├── scripts/      # Management and utility scripts
│   └── docs/         # Documentation
├── auth/             # Authentication and credentials
│   ├── ssh/          # SSH keys
│   ├── tokens/       # GitHub tokens
│   └── git-config/   # Git configuration
├── backups/          # Backup storage
└── [repo-name]/      # Individual repository codespaces
    ├── docker-compose.yml
    ├── .env
    ├── src/          # Cloned repository
    └── data/         # Persistent data
```

## Workflow

### Initial Setup
1. **Base System Installation** (`setup.sh`)
   - Updates Ubuntu packages
   - Installs Docker and docker-compose
   - Configures Zsh with helpful aliases
   - Creates directory structure
   - Sets up monitoring tools

2. **GitHub Authentication** (`setup-github-auth.sh`)
   - Configures SSH keys
   - Stores GitHub personal access tokens
   - Sets up git configuration

### Creating a Codespace
1. **Repository Setup** (`setup-repo-codespace.sh`)
   - Pass a GitHub repository URL: `./setup-repo-codespace.sh git@github.com:user/repo.git`
   - Creates isolated Docker container
   - Clones repository into container
   - Configures VS Code Server
   - Sets up port forwarding
   - Creates management aliases

2. **Access Methods**
   - **Local VM**: Uses Cloudflare tunnel for external access
   - **Cloud VM**: Direct access via public IP and assigned port
   - VS Code accessible at `http://localhost:[assigned-port]`

### Management Commands
- `list-codespaces` - List all available codespaces
- `start-[repo-name]` - Start a specific codespace
- `stop-[repo-name]` - Stop a specific codespace
- `logs-[repo-name]` - View container logs
- `cd-[repo-name]` - Navigate to codespace directory
- `monitor-system.sh` - System and Docker status overview

## Key Features

### Security
- Each repository runs in an isolated Docker container
- SSH key-based authentication for GitHub
- Secure token storage
- Network isolation between codespaces

### Persistence
- Code persists in mounted volumes
- Container data preserved across restarts
- Automatic backup capabilities
- Git integration maintains version control

### Scalability
- Multiple codespaces can run simultaneously
- Resource limits configurable per container
- Easy to add/remove repositories
- Minimal overhead for inactive codespaces

### Convenience
- Browser-based access from any device
- Pre-configured development environment
- Automatic dependency installation
- Integrated terminal and debugging

## Use Cases

1. **Personal Development Machine**
   - Run on laptop/desktop VM
   - Access your code from any browser
   - Keep development environments clean and isolated

2. **Team Development Server**
   - Deploy on cloud VM (EC2, Hetzner, etc.)
   - Multiple developers access different repositories
   - Consistent development environments

3. **Client Project Isolation**
   - Separate containers for each client project
   - No cross-contamination of dependencies
   - Easy to archive/restore projects

## Requirements

- Ubuntu 20.04+ (VM or bare metal)
- Minimum 4GB RAM (8GB+ recommended)
- 20GB+ free disk space
- Internet connection for GitHub access
- (Optional) Cloudflare account for tunnel access

## Next Steps

After running `setup.sh`:
1. Configure GitHub authentication
2. Create your first repository codespace
3. Access VS Code in your browser
4. Start coding with full isolation and persistence

This solution provides the flexibility and power of GitHub Codespaces while maintaining full control over your infrastructure and data.