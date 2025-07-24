#!/bin/bash

# GitHub Authentication Module
# Handles all GitHub authentication operations

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AUTH_DIR="$HOME/codespaces/auth"
TOKENS_DIR="$AUTH_DIR/tokens"
SSH_DIR="$AUTH_DIR/ssh"

# Source utilities if available
if [ -f "$HOME/codespaces/scripts/utils/colors.sh" ]; then
    source "$HOME/codespaces/scripts/utils/colors.sh"
else
    echo_info() { echo "ℹ️  $1"; }
    echo_success() { echo "✅ $1"; }
    echo_warning() { echo "⚠️  $1"; }
    echo_error() { echo "❌ $1"; }
fi

# Check if GitHub token exists and is valid
check_github_token() {
    local token_file="$TOKENS_DIR/github.token"
    
    if [ ! -f "$token_file" ]; then
        return 1
    fi
    
    local token=$(cat "$token_file" 2>/dev/null)
    
    # Validate token format
    if [[ ! "$token" =~ ^gh[ps]_[a-zA-Z0-9]{36,}$ ]]; then
        return 1
    fi
    
    # Optionally validate with GitHub API
    if [ "$1" == "--validate-api" ]; then
        if ! curl -s -H "Authorization: token $token" https://api.github.com/user >/dev/null 2>&1; then
            return 1
        fi
    fi
    
    return 0
}

# Get GitHub token
get_github_token() {
    local token_file="$TOKENS_DIR/github.token"
    
    if check_github_token; then
        cat "$token_file"
    else
        echo ""
    fi
}

# Set GitHub token
set_github_token() {
    local token="$1"
    
    # Validate token format
    if [[ ! "$token" =~ ^gh[ps]_[a-zA-Z0-9]{36,}$ ]]; then
        echo_error "Invalid GitHub token format"
        return 1
    fi
    
    # Create directory if needed
    mkdir -p "$TOKENS_DIR"
    
    # Save token with secure permissions
    echo "$token" > "$TOKENS_DIR/github.token"
    chmod 600 "$TOKENS_DIR/github.token"
    
    echo_success "GitHub token saved successfully"
}

# Setup SSH key for GitHub
setup_github_ssh() {
    local key_name="${1:-id_ed25519}"
    local key_path="$SSH_DIR/$key_name"
    
    # Create SSH directory
    mkdir -p "$SSH_DIR"
    
    # Check if key already exists
    if [ -f "$key_path" ]; then
        echo_info "SSH key already exists at $key_path"
        return 0
    fi
    
    # Generate new SSH key
    echo_info "Generating new SSH key..."
    read -p "Enter your email for SSH key: " email
    
    ssh-keygen -t ed25519 -C "$email" -f "$key_path" -N ""
    
    # Set proper permissions
    chmod 600 "$key_path"
    chmod 644 "$key_path.pub"
    
    echo_success "SSH key generated successfully"
    echo ""
    echo "Public key:"
    cat "$key_path.pub"
    
    return 0
}

# Convert repository URL to authenticated URL
convert_to_auth_url() {
    local repo_url="$1"
    local use_token="${2:-true}"
    
    # If token authentication is disabled, return original URL
    if [ "$use_token" != "true" ]; then
        echo "$repo_url"
        return
    fi
    
    local token=$(get_github_token)
    
    # If no token, return original URL
    if [ -z "$token" ]; then
        echo "$repo_url"
        return
    fi
    
    # Convert SSH URL to HTTPS with token
    if [[ "$repo_url" =~ ^git@github\.com:(.+)$ ]]; then
        local repo_path="${BASH_REMATCH[1]}"
        echo "https://${token}@github.com/${repo_path}"
    # Add token to HTTPS URL
    elif [[ "$repo_url" =~ ^https://github\.com/(.+)$ ]]; then
        local repo_path="${BASH_REMATCH[1]}"
        echo "https://${token}@github.com/${repo_path}"
    else
        # Non-GitHub URL, return as-is
        echo "$repo_url"
    fi
}

# Validate GitHub repository access
validate_repo_access() {
    local repo_url="$1"
    
    # Extract owner and repo from URL
    local owner=""
    local repo=""
    
    if [[ "$repo_url" =~ github\.com[:/]([^/]+)/([^/\.]+)(\.git)?$ ]]; then
        owner="${BASH_REMATCH[1]}"
        repo="${BASH_REMATCH[2]}"
    else
        return 1
    fi
    
    local token=$(get_github_token)
    
    if [ -n "$token" ]; then
        # Check with token
        if curl -s -H "Authorization: token $token" \
            "https://api.github.com/repos/$owner/$repo" >/dev/null 2>&1; then
            return 0
        fi
    else
        # Check public access
        if curl -s "https://api.github.com/repos/$owner/$repo" | grep -q '"private":false'; then
            return 0
        fi
    fi
    
    return 1
}

# Interactive GitHub setup
interactive_github_setup() {
    echo_info "GitHub Authentication Setup"
    echo "=========================="
    echo ""
    
    # SSH Key Setup
    echo "1. SSH Key Setup"
    read -p "Do you want to set up an SSH key for GitHub? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        setup_github_ssh
        echo ""
        echo "Add this key to GitHub:"
        echo "https://github.com/settings/ssh/new"
        echo ""
        read -p "Press Enter after adding the key to GitHub..."
    fi
    
    # Personal Access Token
    echo ""
    echo "2. Personal Access Token (recommended)"
    echo "This allows enhanced features like:"
    echo "  - Private repository access"
    echo "  - API rate limit increases"
    echo "  - Repository creation"
    echo ""
    echo "Create a token at: https://github.com/settings/tokens/new"
    echo "Required scopes: repo, workflow"
    echo ""
    read -sp "Paste your token here (or press Enter to skip): " token
    echo
    
    if [ -n "$token" ]; then
        if set_github_token "$token"; then
            echo_success "GitHub authentication configured successfully!"
        else
            echo_error "Failed to save GitHub token"
        fi
    fi
}

# Export functions for use by other scripts
export -f check_github_token
export -f get_github_token
export -f set_github_token
export -f convert_to_auth_url
export -f validate_repo_access