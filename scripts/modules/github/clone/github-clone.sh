#!/bin/bash

# GitHub Clone Module
# Handles repository cloning with authentication and error handling

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AUTH_MODULE="$SCRIPT_DIR/../auth/github-auth.sh"

# Source authentication module
if [ -f "$AUTH_MODULE" ]; then
    source "$AUTH_MODULE"
fi

# Source utilities
if [ -f "$HOME/codespaces/scripts/utils/colors.sh" ]; then
    source "$HOME/codespaces/scripts/utils/colors.sh"
else
    echo_info() { echo "ℹ️  $1"; }
    echo_success() { echo "✅ $1"; }
    echo_warning() { echo "⚠️  $1"; }
    echo_error() { echo "❌ $1"; }
    echo_debug() { [ "${DEBUG:-0}" -eq 1 ] && echo "🔍 $1"; }
fi

# Clone repository with authentication and progress
clone_repository() {
    local repo_url="$1"
    local target_dir="$2"
    local branch="${3:-}"
    local depth="${4:-}"
    
    # Validate inputs
    if [ -z "$repo_url" ] || [ -z "$target_dir" ]; then
        echo_error "Usage: clone_repository <repo_url> <target_dir> [branch] [depth]"
        return 1
    fi
    
    # Check if target directory already exists
    if [ -d "$target_dir" ] && [ -n "$(ls -A "$target_dir" 2>/dev/null)" ]; then
        echo_warning "Target directory is not empty: $target_dir"
        return 1
    fi
    
    # Create target directory
    mkdir -p "$target_dir"
    
    # Get authenticated URL
    local auth_url=$(convert_to_auth_url "$repo_url")
    echo_debug "Using URL: ${auth_url//*@github.com/***@github.com}"
    
    # Build clone command
    local clone_cmd="git clone"
    
    # Add branch if specified
    if [ -n "$branch" ]; then
        clone_cmd="$clone_cmd --branch $branch"
    fi
    
    # Add depth if specified
    if [ -n "$depth" ]; then
        clone_cmd="$clone_cmd --depth $depth"
    fi
    
    # Add progress flag
    clone_cmd="$clone_cmd --progress"
    
    # Clone with error handling
    echo_info "Cloning repository..."
    
    # Use a temporary file for stderr to parse progress
    local temp_err=$(mktemp)
    
    if $clone_cmd "$auth_url" "$target_dir" 2>"$temp_err"; then
        echo_success "Repository cloned successfully"
        
        # Remove authentication from remote URL for security
        if [ -d "$target_dir/.git" ]; then
            cd "$target_dir"
            git remote set-url origin "$repo_url"
            cd - >/dev/null
        fi
        
        rm -f "$temp_err"
        return 0
    else
        local exit_code=$?
        local error_msg=$(cat "$temp_err")
        rm -f "$temp_err"
        
        # Parse common error messages
        if echo "$error_msg" | grep -q "Authentication failed"; then
            echo_error "Authentication failed. Please check your credentials."
            echo_info "Run: ~/codespaces/shared/scripts/setup-github-auth.sh"
        elif echo "$error_msg" | grep -q "Could not resolve host"; then
            echo_error "Network error: Could not resolve host"
        elif echo "$error_msg" | grep -q "Permission denied"; then
            echo_error "Permission denied. You may not have access to this repository."
        elif echo "$error_msg" | grep -q "Repository not found"; then
            echo_error "Repository not found or you don't have access."
        else
            echo_error "Clone failed: $error_msg"
        fi
        
        # Clean up failed clone
        rm -rf "$target_dir"
        
        return $exit_code
    fi
}

# Clone with retry logic
clone_with_retry() {
    local repo_url="$1"
    local target_dir="$2"
    local max_retries="${3:-3}"
    local retry_delay="${4:-5}"
    
    local attempt=1
    
    while [ $attempt -le $max_retries ]; do
        echo_info "Clone attempt $attempt of $max_retries"
        
        if clone_repository "$repo_url" "$target_dir"; then
            return 0
        fi
        
        if [ $attempt -lt $max_retries ]; then
            echo_warning "Clone failed, retrying in $retry_delay seconds..."
            sleep $retry_delay
        fi
        
        ((attempt++))
    done
    
    echo_error "Failed to clone after $max_retries attempts"
    return 1
}

# Check if repository has .devcontainer configuration
check_devcontainer() {
    local repo_dir="$1"
    
    if [ -f "$repo_dir/.devcontainer.json" ]; then
        echo "$repo_dir/.devcontainer.json"
    elif [ -f "$repo_dir/.devcontainer/devcontainer.json" ]; then
        echo "$repo_dir/.devcontainer/devcontainer.json"
    else
        echo ""
    fi
}

# Parse .devcontainer.json for image information
parse_devcontainer_image() {
    local devcontainer_file="$1"
    
    if [ ! -f "$devcontainer_file" ]; then
        echo ""
        return
    fi
    
    # Try to extract image using various methods
    # This is a simple parser - for production, use jq
    local image=""
    
    # Check for "image" field
    image=$(grep -E '"image"[[:space:]]*:[[:space:]]*"[^"]*"' "$devcontainer_file" | \
            sed -E 's/.*"image"[[:space:]]*:[[:space:]]*"([^"]*)".*$/\1/' | \
            head -n1)
    
    if [ -n "$image" ]; then
        echo "$image"
        return
    fi
    
    # Check for build.dockerfile
    local dockerfile=$(grep -E '"dockerfile"[[:space:]]*:[[:space:]]*"[^"]*"' "$devcontainer_file" | \
                      sed -E 's/.*"dockerfile"[[:space:]]*:[[:space:]]*"([^"]*)".*$/\1/' | \
                      head -n1)
    
    if [ -n "$dockerfile" ]; then
        echo "dockerfile:$dockerfile"
        return
    fi
    
    echo ""
}

# Export functions
export -f clone_repository
export -f clone_with_retry
export -f check_devcontainer
export -f parse_devcontainer_image