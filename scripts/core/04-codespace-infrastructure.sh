#!/bin/bash

# Create codespace infrastructure and directories

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../utils/colors.sh"

echo_step "📁 Creating codespace infrastructure..."

# Create directory structure
DIRECTORIES=(
    "$HOME/codespaces"
    "$HOME/codespaces/shared"
    "$HOME/codespaces/shared/templates"
    "$HOME/codespaces/shared/scripts"
    "$HOME/codespaces/shared/docs"
    "$HOME/codespaces/auth"
    "$HOME/codespaces/auth/ssh"
    "$HOME/codespaces/auth/tokens"
    "$HOME/codespaces/auth/git-config"
    "$HOME/codespaces/backups"
    "$HOME/codespaces/scripts"
    "$HOME/codespaces/scripts/utils"
)

for dir in "${DIRECTORIES[@]}"; do
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir"
        echo_debug "Created: $dir"
    fi
done

# Copy utility scripts
echo_info "Installing utility scripts..."
cp -r "$SCRIPT_DIR/../utils" "$HOME/codespaces/scripts/"

# Create templates
echo_info "Creating templates..."

# README template for codespaces
cat > "$HOME/codespaces/shared/templates/README.template.md" << 'EOF'
# {{REPO_NAME}} Development Environment

## Quick Start

```bash
# Start development environment
start-{{SAFE_REPO_NAME}}

# Access VS Code
open http://localhost:{{VS_CODE_PORT}}

# View logs
logs-{{SAFE_REPO_NAME}}

# Stop environment
stop-{{SAFE_REPO_NAME}}
```

## Container Details

- **VS Code Port**: {{VS_CODE_PORT}}
- **App Port**: {{APP_PORT}} (if applicable)
- **Container**: {{CONTAINER_NAME}}
- **Created**: {{DATE}}

## Useful Commands

- `cd-{{SAFE_REPO_NAME}}` - Navigate to project directory
- `exec-{{SAFE_REPO_NAME}}` - Enter container shell
- `rebuild-{{SAFE_REPO_NAME}}` - Rebuild container
- `backup-{{SAFE_REPO_NAME}}` - Create backup

## Repository

{{REPO_URL}}
EOF

# Docker compose template
cat > "$HOME/codespaces/shared/templates/docker-compose.template.yml" << 'EOF'
services:
  {{SAFE_REPO_NAME}}-dev:
    image: codercom/code-server:latest
    container_name: {{CONTAINER_NAME}}
    restart: unless-stopped
    environment:
      - PASSWORD={{PASSWORD}}
      - TZ=${TZ:-UTC}
      - DOCKER_USER=${USER}
    ports:
      - "{{VS_CODE_PORT}}:8080"
      - "{{APP_PORT}}:3000"
    volumes:
      - ./src:/home/coder/{{SAFE_REPO_NAME}}
      - ./data:/home/coder/.local/share/code-server
      - ${HOME}/.ssh:/home/coder/.ssh:ro
      - ${HOME}/codespaces/auth/git-config:/home/coder/.gitconfig:ro
    networks:
      - codespace-network
    labels:
      - "codespace.repo={{REPO_NAME}}"
      - "codespace.created={{DATE}}"

networks:
  codespace-network:
    name: {{SAFE_REPO_NAME}}-network
    driver: bridge
EOF

# Environment template
cat > "$HOME/codespaces/shared/templates/env.template" << 'EOF'
# Codespace Configuration
REPO_NAME={{REPO_NAME}}
REPO_URL={{REPO_URL}}
CONTAINER_NAME={{CONTAINER_NAME}}
VS_CODE_PORT={{VS_CODE_PORT}}
APP_PORT={{APP_PORT}}
PASSWORD={{PASSWORD}}
CREATED={{DATE}}

# User Configuration
USER={{USER}}
TZ={{TZ}}
EOF

# Create shared scripts
echo_info "Creating shared scripts..."

# GitHub auth setup script
cat > "$HOME/codespaces/shared/scripts/setup-github-auth.sh" << 'EOF'
#!/bin/bash

source "$HOME/codespaces/scripts/utils/colors.sh"

echo_step "🔐 GitHub Authentication Setup"
echo "=============================="
echo ""

# Check for existing SSH key
if [ ! -f ~/.ssh/id_ed25519 ]; then
    echo_info "Creating SSH key..."
    read -p "Enter your email for SSH key: " email
    ssh-keygen -t ed25519 -C "$email" -f ~/.ssh/id_ed25519 -N ""
fi

echo ""
echo_info "Your SSH public key:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat ~/.ssh/id_ed25519.pub
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo_step "Add this key to GitHub:"
echo "1. Go to: https://github.com/settings/ssh/new"
echo "2. Give it a title (e.g., 'Ubuntu Codespace')"
echo "3. Paste the key above"
echo ""
read -p "Press Enter after adding the key to GitHub..."

# Test SSH connection
echo ""
echo_info "Testing SSH connection..."
if ssh -T git@github.com 2>&1 | grep -q "successfully authenticated"; then
    echo_success "SSH authentication successful!"
else
    echo_warning "SSH test failed. You may need to check your key setup."
fi

# Personal Access Token
echo ""
echo_step "GitHub Personal Access Token (optional but recommended)"
echo "This allows enhanced features like creating repos, managing issues, etc."
echo ""
echo "1. Go to: https://github.com/settings/tokens/new"
echo "2. Give it a name (e.g., 'Ubuntu Codespace')"
echo "3. Select scopes: repo, workflow, write:packages"
echo "4. Generate token"
echo ""
read -sp "Paste your token here (or press Enter to skip): " token
echo ""

if [ -n "$token" ]; then
    mkdir -p ~/codespaces/auth/tokens
    echo "$token" > ~/codespaces/auth/tokens/github.token
    chmod 600 ~/codespaces/auth/tokens/github.token
    echo_success "Token saved securely"
fi

# Git configuration
echo ""
echo_step "Git Configuration"
read -p "Enter your Git username: " git_name
read -p "Enter your Git email: " git_email

git config --global user.name "$git_name"
git config --global user.email "$git_email"
git config --global init.defaultBranch main
git config --global pull.rebase false

# Save for codespaces
echo "$git_name" > ~/codespaces/auth/git-config/name
echo "$git_email" > ~/codespaces/auth/git-config/email

# Create .gitconfig for containers
cat > ~/codespaces/auth/git-config/.gitconfig << EOG
[user]
    name = $git_name
    email = $git_email
[init]
    defaultBranch = main
[pull]
    rebase = false
[core]
    editor = vim
EOG

echo ""
echo_success "🎉 GitHub authentication setup complete!"
echo ""
echo "You can now create codespaces with:"
echo "  ~/setup-repo-codespace.sh git@github.com:user/repo.git"
EOF

chmod +x "$HOME/codespaces/shared/scripts/setup-github-auth.sh"

# Backup script
cat > "$HOME/codespaces/shared/scripts/backup-all.sh" << 'EOF'
#!/bin/bash

source "$HOME/codespaces/scripts/utils/colors.sh"

BACKUP_DIR="$HOME/codespaces/backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

echo_step "💾 Creating codespace backup..."

# Stop all containers first
echo_info "Stopping containers..."
cd ~/codespaces
for dir in */; do
    if [ -f "$dir/docker-compose.yml" ]; then
        docker-compose -f "$dir/docker-compose.yml" stop 2>/dev/null || true
    fi
done

# Create backup
echo_info "Creating backup archive..."
tar -czf "$BACKUP_DIR/codespaces-backup.tar.gz" \
    --exclude='*/node_modules' \
    --exclude='*/.git/objects' \
    --exclude='*/vendor' \
    --exclude='*/__pycache__' \
    -C "$HOME" codespaces/

# Get backup size
SIZE=$(du -h "$BACKUP_DIR/codespaces-backup.tar.gz" | cut -f1)

echo_success "Backup completed!"
echo "Location: $BACKUP_DIR/codespaces-backup.tar.gz"
echo "Size: $SIZE"
EOF

chmod +x "$HOME/codespaces/shared/scripts/backup-all.sh"

echo_success "Codespace infrastructure created successfully"