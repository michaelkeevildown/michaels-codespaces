#!/bin/bash

# GitHub authentication setup
# Integrated into main installation flow

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../utils/colors.sh"

echo_step "🔐 Setting up GitHub Authentication..."

# Function to test SSH connection
test_github_ssh() {
    ssh -T git@github.com 2>&1 | grep -q "successfully authenticated"
}

# Check if GitHub SSH is already configured
if test_github_ssh; then
    echo_success "GitHub SSH authentication already configured!"
    return 0
fi

echo ""
echo_info "GitHub authentication is required for creating codespaces."
echo_info "Let's set this up now (it only takes a minute)."
echo ""

# Prompt to continue or skip
read -p "Setup GitHub authentication now? [Y/n] " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    echo_warning "Skipping GitHub setup. You can run setup-github-auth.sh later."
    return 0
fi

# Check for existing SSH key
SSH_KEY_PATH="$HOME/.ssh/id_ed25519"
if [ ! -f "$SSH_KEY_PATH" ]; then
    echo_info "Creating SSH key..."
    read -p "Enter your email for SSH key: " email
    ssh-keygen -t ed25519 -C "$email" -f "$SSH_KEY_PATH" -N ""
    echo_success "SSH key created"
fi

# Add GitHub's SSH key fingerprint automatically
echo_info "Adding GitHub to known hosts..."
mkdir -p ~/.ssh
ssh-keyscan -t ed25519 github.com >> ~/.ssh/known_hosts 2>/dev/null

# Display the public key
echo ""
echo_info "Your SSH public key:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat "$SSH_KEY_PATH.pub"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Provide instructions
echo_step "Add this key to your GitHub account:"
echo "1. Copy the SSH key above"
echo "2. Open: https://github.com/settings/ssh/new"
echo "3. Give it a title (e.g., 'Ubuntu Codespace')"
echo "4. Paste the key and click 'Add SSH key'"
echo ""

# Wait for user to add the key
read -p "Press Enter after adding the key to GitHub..."

# Test SSH connection
echo ""
echo_info "Testing GitHub SSH connection..."
if test_github_ssh; then
    echo_success "GitHub SSH authentication successful!"
else
    echo_warning "SSH test failed. You may need to check your key setup."
    echo_info "You can run ~/codespaces/shared/scripts/setup-github-auth.sh later to retry."
fi

# Git configuration
echo ""
echo_step "Configuring Git..."

# Check if git is already configured
if git config --global user.name &>/dev/null && git config --global user.email &>/dev/null; then
    echo_info "Git already configured:"
    echo "  Name: $(git config --global user.name)"
    echo "  Email: $(git config --global user.email)"
    read -p "Keep current configuration? [Y/n] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        echo_success "Git configuration kept"
        return 0
    fi
fi

# Configure git
read -p "Enter your Git username: " git_name
read -p "Enter your Git email: " git_email

git config --global user.name "$git_name"
git config --global user.email "$git_email"
git config --global init.defaultBranch main
git config --global pull.rebase false

# Save for codespaces
mkdir -p ~/codespaces/auth/git-config
echo "$git_name" > ~/codespaces/auth/git-config/name
echo "$git_email" > ~/codespaces/auth/git-config/email

# Create .gitconfig for containers
cat > ~/codespaces/auth/git-config/.gitconfig << EOF
[user]
    name = $git_name
    email = $git_email
[init]
    defaultBranch = main
[pull]
    rebase = false
[core]
    editor = vim
EOF

echo_success "Git configuration complete!"

# Optional: Personal Access Token
echo ""
echo_info "GitHub Personal Access Token (optional)"
echo "A token enables enhanced features like creating repos and managing issues."
read -p "Would you like to set up a token? [y/N] " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo_step "Create a Personal Access Token:"
    echo "1. Open: https://github.com/settings/tokens/new"
    echo "2. Give it a name (e.g., 'Ubuntu Codespace')"
    echo "3. Select scopes: repo, workflow, write:packages"
    echo "4. Generate token and copy it"
    echo ""
    read -sp "Paste your token here: " token
    echo ""
    
    if [ -n "$token" ]; then
        mkdir -p ~/codespaces/auth/tokens
        echo "$token" > ~/codespaces/auth/tokens/github.token
        chmod 600 ~/codespaces/auth/tokens/github.token
        echo_success "Token saved securely"
    fi
fi

echo ""
echo_success "GitHub authentication setup complete!"