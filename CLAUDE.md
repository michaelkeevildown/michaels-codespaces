# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

Michael's Codespaces provides infrastructure for creating self-hosted GitHub Codespaces - isolated, browser-accessible VS Code development environments for GitHub repositories. It's designed to work on Ubuntu VMs with Docker containers providing isolation for each repository.

## Key Scripts and Commands

### Base System Setup
```bash
# Initial VM setup - run as regular user (not root)
./setup.sh

# After setup, logout and login again for Docker permissions
exit
ssh coder@<vm-ip>
```

### GitHub Authentication Setup
```bash
# Configure SSH keys and GitHub tokens
~/codespaces/shared/scripts/setup-github-auth.sh
```

### Creating Repository Codespaces
```bash
# Setup a new codespace for a repository (script needs to be created)
./setup-repo-codespace.sh git@github.com:user/repo.git
```

### System Monitoring
```bash
# Check system and Docker status
~/monitor-system.sh

# List all codespaces
list-codespaces

# Manage all codespaces
~/codespaces/manage-all.sh start-all
~/codespaces/manage-all.sh stop-all
```

## Architecture Overview

### Directory Structure
The system creates a structured directory hierarchy at `~/codespaces/`:

- **shared/** - Resources shared across all codespaces
  - templates/ - README templates for new codespaces
  - scripts/ - Management scripts (backup-all.sh, setup-github-auth.sh)
  - docs/ - Documentation
- **auth/** - Authentication and credentials
  - ssh/ - SSH keys
  - tokens/ - GitHub personal access tokens
  - git-config/ - Git configuration
- **backups/** - Backup storage
- **[repo-name]/** - Individual repository codespaces (created by setup-repo-codespace.sh)

### Key Components

1. **Base VM Requirements**: Ubuntu 20.04+, Docker, docker-compose, Zsh with Oh My Zsh
2. **Container Architecture**: Each repository runs in isolated Docker container with VS Code Server
3. **Networking**: Port-based isolation (8080-8089 for VS Code, 7680-7689 for other services)
4. **Persistence**: Code and data stored in mounted volumes

### Environment Considerations

- **Local VM Setup**: Requires Cloudflare tunnel for external access
- **Cloud VM Setup**: Direct access via public IP
- **Docker Configuration**: Custom daemon.json with log rotation and network pools

## Development Workflow

### Creating the Repository Setup Script
The main missing component is `setup-repo-codespace.sh` which should:
1. Accept a GitHub repository URL as parameter
2. Create directory structure under ~/codespaces/[repo-name]/
3. Generate docker-compose.yml for VS Code container
4. Clone the repository into the container
5. Configure port assignments and environment variables
6. Create management aliases (start-[repo], stop-[repo], etc.)

### Alias System
The setup creates numerous Zsh aliases for quick navigation and management:
- `cds` - Navigate to codespaces directory
- `start-all` / `stop-all` - Manage all codespaces
- Repository-specific aliases created by setup-repo-codespace.sh

## Security Notes

- Never run setup.sh as root
- GitHub tokens stored in ~/codespaces/auth/tokens/ with 600 permissions
- Each repository isolated in its own Docker container
- SSH keys managed separately from repository code