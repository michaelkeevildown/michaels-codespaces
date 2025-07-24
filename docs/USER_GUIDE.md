# Michael's Codespaces - User Guide

## Quick Start

### Installation

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/main/install.sh)"
```

### Create Your First Codespace

```bash
# Basic usage
mcs create git@github.com:facebook/react.git

# With custom name
mcs create https://github.com/nodejs/node.git --name my-node-dev
```

## Enhanced Features

### 1. Custom Docker Images

Use any Docker image for your development environment:

```bash
# Python development
mcs create git@github.com:user/python-app.git --image python:3.11-slim

# Node.js specific version
mcs create git@github.com:user/node-app.git --image node:18-alpine

# Custom company image
mcs create git@github.com:company/app.git --image company/dev-environment:latest
```

### 2. Language Presets

Automatically configure environments for specific languages:

```bash
# Node.js with proper volume mounts
mcs create git@github.com:user/app.git --language node

# Python with pip cache
mcs create git@github.com:user/app.git --language python

# Go with module cache
mcs create git@github.com:user/app.git --language go

# Rust with cargo cache
mcs create git@github.com:user/app.git --language rust

# Java with Maven cache
mcs create git@github.com:user/app.git --language java
```

### 3. Custom Port Mappings

Configure specific ports for your applications:

```bash
# Single custom port
mcs create git@github.com:user/app.git --ports "3000:3000"

# Multiple ports
mcs create git@github.com:user/app.git --ports "8080:8080,3000:3000,5432:5432"

# Different host and container ports
mcs create git@github.com:user/app.git --ports "8090:8080,3001:3000"
```

### 4. Environment Variables

Load environment variables from a file:

```bash
# Create .env file
cat > myapp.env << EOF
NODE_ENV=development
API_KEY=your-api-key
DATABASE_URL=postgresql://localhost:5432/myapp
EOF

# Use with codespace
mcs create git@github.com:user/app.git --env-file myapp.env
```

### 5. DevContainer Support

Automatically detects and uses `.devcontainer.json` configuration:

```bash
# If repository has .devcontainer.json, it will be used automatically
mcs create git@github.com:microsoft/vscode-remote-try-node.git
```

### 6. Advanced Options

```bash
# Don't start container immediately
mcs create git@github.com:user/app.git --no-start

# Force overwrite existing codespace
mcs create git@github.com:user/app.git --force

# Enable debug output
mcs create git@github.com:user/app.git --debug

# Combine multiple options
mcs create git@github.com:user/app.git \
  --name my-dev \
  --image node:18 \
  --ports "8080:8080,3000:3000" \
  --env-file .env \
  --debug
```

## Managing Codespaces

### List All Codespaces

```bash
mcs list

# Output:
# NAME                      STATUS    URL
# ----                      ------    ---
# facebook-react           running    http://localhost:8080
# nodejs-node              stopped    port 8081
```

### Get Detailed Information

```bash
mcs info facebook-react

# Output:
# Codespace: facebook-react
# ─────────────────────────
# Repository: https://github.com/facebook/react.git
# VS Code Port: 8080
# App Port: 7680
# Password: aBc123XyZ
# Created: 2024-01-15
# Status: running
# Resources: 512MB / 2.1%
# Location: /home/user/codespaces/facebook-react
# Source size: 125M
```

### Container Management

```bash
# Start a stopped codespace
mcs start facebook-react

# Stop a running codespace
mcs stop facebook-react

# Restart a codespace
mcs restart facebook-react

# View logs
mcs logs facebook-react

# Enter container shell
mcs exec facebook-react

# Run command in container
mcs exec facebook-react npm install
mcs exec facebook-react npm test
```

### Removal and Cleanup

```bash
# Remove a specific codespace
mcs remove facebook-react

# Clean up port registrations
mcs cleanup-ports

# Remove MCS but keep Docker
mcs cleanup

# Complete uninstall
mcs destroy
```

## GitHub Authentication

### Setup GitHub Access

```bash
# Interactive setup
~/codespaces/shared/scripts/setup-github-auth.sh

# This will:
# 1. Generate SSH key (optional)
# 2. Guide you to add it to GitHub
# 3. Setup personal access token (recommended)
# 4. Configure Git settings
```

### Working with Private Repositories

1. Ensure GitHub token is configured
2. Use standard repository URLs:
```bash
mcs create git@github.com:company/private-repo.git
```

## Port Management

The system automatically manages ports to avoid conflicts:

- VS Code ports: 8080-8089 (default range)
- Application ports: 7680-7689 (default range)
- Custom ports: Any available port via --ports

View port allocations:
```bash
# Integrated in mcs list output
mcs list

# Check system ports
mcs doctor
```

## Troubleshooting

### Common Issues

#### 1. Port Already in Use
```bash
# Use custom ports
mcs create git@github.com:user/app.git --ports "8090:8080"
```

#### 2. Authentication Failed
```bash
# Re-run GitHub setup
~/codespaces/shared/scripts/setup-github-auth.sh

# Check token
cat ~/codespaces/auth/tokens/github.token
```

#### 3. Container Won't Start
```bash
# Check logs
mcs logs codespace-name

# Validate docker-compose
cd ~/codespaces/codespace-name
docker-compose config

# Rebuild
docker-compose up -d --build
```

### Health Check

```bash
mcs doctor

# Output:
# Michael's Codespaces Doctor
# ─────────────────────────
# Docker: ✓ installed and running
# Docker Compose: ✓ installed
# Codespaces directory: ✓ exists
# Setup script: ✓ found and executable
# GitHub Token: ✓ configured
```

## Best Practices

1. **Use meaningful names**: `--name project-feature-branch`
2. **Configure GitHub auth**: For private repos and better rate limits
3. **Use language presets**: Optimized configurations
4. **Clean up regularly**: Remove unused codespaces
5. **Use .devcontainer.json**: For reproducible environments
6. **Backup important work**: Codespaces are temporary

## Advanced Usage

### Custom Aliases

After creating a codespace, aliases are automatically added:

```bash
cd-myproject     # Go to codespace directory
src-myproject    # Go to source code
start-myproject  # Start container
stop-myproject   # Stop container
logs-myproject   # View logs
exec-myproject   # Enter container
```

### Integration with VS Code

1. Start your codespace: `mcs start myproject`
2. Open browser: `http://localhost:8080`
3. Enter password from `mcs info myproject`
4. Install extensions and configure as needed

### Monitoring

```bash
# System overview
~/monitor-system.sh

# Watch specific codespace
watch -n 5 'docker stats --no-stream $(docker ps --format "table {{.Names}}" | grep myproject)'
```

## Getting Help

```bash
# General help
mcs help

# Command-specific help
mcs help create

# Check system
mcs doctor

# Report issues
https://github.com/michaelkeevildown/michaels-codespaces/issues
```