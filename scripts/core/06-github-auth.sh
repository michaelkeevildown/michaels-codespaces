#!/bin/bash

# GitHub authentication setup - Token-based for security
# Uses Personal Access Tokens instead of SSH keys

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../utils/colors.sh"

echo_step "🔐 Configuring GitHub access..."

# Create directories for auth
mkdir -p ~/codespaces/auth/tokens
mkdir -p ~/codespaces/auth/git-config

# Check for existing token
TOKEN_FILE="$HOME/codespaces/auth/tokens/github.token"
if [ -f "$TOKEN_FILE" ] && [ -s "$TOKEN_FILE" ]; then
    echo_success "GitHub token already configured"
else
    # Check environment variable first
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        echo_debug "Using GitHub token from environment"
        echo "$GITHUB_TOKEN" > "$TOKEN_FILE"
        chmod 600 "$TOKEN_FILE"
        echo_success "GitHub token saved from environment"
    else
        # Create a placeholder and instructions
        echo_info "GitHub Personal Access Token required for codespace creation"
        
        # Create instructions file
        cat > ~/codespaces/auth/tokens/README.md << 'EOF'
# GitHub Authentication Setup

To use Michael's Codespaces, you need a GitHub Personal Access Token.

## Quick Setup (Do this now!)

### Step 1: Create Your Token

**Click this link:** https://github.com/settings/tokens/new

Or manually navigate to:
1. GitHub.com → Click your profile picture (top right)
2. Settings → Developer settings (bottom of left sidebar)
3. Personal access tokens → Tokens (classic) → Generate new token

### Step 2: Configure Your Token

On the token creation page:

1. **Note**: Enter "Michael's Codespaces - $(hostname)"
2. **Expiration**: Select "90 days" (recommended for security)
3. **Select scopes** - Check these boxes:
   - ✅ `repo` (Full control of private repositories)
   - ✅ `workflow` (Update GitHub Action workflows)
   - ✅ `write:packages` (Upload packages to GitHub Package Registry)
4. Scroll down and click the green **"Generate token"** button
5. **IMPORTANT**: Copy your token immediately! (looks like: ghp_xxxxxxxxxxxx)

### Step 3: Save Your Token

Run this command with your token:
```bash
echo "ghp_YOUR_TOKEN_HERE" > ~/codespaces/auth/tokens/github.token
chmod 600 ~/codespaces/auth/tokens/github.token
```

Example:
```bash
echo "ghp_A1b2C3d4E5f6G7h8I9j0K1L2M3N4O5P6Q7R8" > ~/codespaces/auth/tokens/github.token
chmod 600 ~/codespaces/auth/tokens/github.token
```

## Alternative: Environment Variable

You can also set it as an environment variable:
```bash
export GITHUB_TOKEN="ghp_YOUR_TOKEN_HERE"
```

Add to ~/.bashrc or ~/.zshrc to make it permanent:
```bash
echo 'export GITHUB_TOKEN="ghp_YOUR_TOKEN_HERE"' >> ~/.zshrc
```

## Verify Token Is Set

Check if your token is saved:
```bash
ls -la ~/codespaces/auth/tokens/github.token
# Should show: -rw------- 1 user user 40 date time github.token
```

## Security Best Practices

- **Never share your token** - Treat it like a password
- **Never commit tokens** to git repositories
- **Rotate regularly** - Create new tokens every 90 days
- **Use minimum scopes** - Only check the permissions you need
- **Revoke if compromised** - https://github.com/settings/tokens

## Troubleshooting

Token not working? Check:
1. Token hasn't expired
2. Token has correct scopes (repo, workflow, write:packages)
3. File permissions are 600 (read-only by you)
4. No extra spaces or newlines in the token file

Need to revoke/regenerate? Go to:
https://github.com/settings/tokens
EOF
        
        echo_warning "GitHub token not found"
        echo ""
        echo_step "Action required:"
        echo "1. Open: ${COLOR_CYAN}${COLOR_UNDERLINE}https://github.com/settings/tokens/new${COLOR_RESET}"
        echo "2. Create a token with 'repo', 'workflow', and 'write:packages' scopes"
        echo "3. Save it to: ~/codespaces/auth/tokens/github.token"
        echo ""
        echo_info "Detailed instructions saved to: ~/codespaces/auth/tokens/README.md"
    fi
fi

# Set up git config if not already done
if ! git config --global user.name &>/dev/null; then
    echo_debug "Setting default git config..."
    git config --global user.name "${USER}"
    git config --global user.email "${USER}@$(hostname)"
    git config --global init.defaultBranch main
    git config --global pull.rebase false
fi

# Configure git to use token authentication
if [ -f "$TOKEN_FILE" ]; then
    # Set up git credential helper to use our token
    git config --global credential.helper store
    git config --global credential.https://github.com.username token
    
    # Note: We don't store the token in git-credentials for security
    # It will be used by our scripts when needed
fi

# Copy git config for containers
if [ -f ~/.gitconfig ]; then
    cp ~/.gitconfig ~/codespaces/auth/git-config/.gitconfig 2>/dev/null || true
fi

echo_success "GitHub authentication configured"
if [ ! -f "$TOKEN_FILE" ] || [ ! -s "$TOKEN_FILE" ]; then
    echo_caveats
    echo "You need to set up a GitHub Personal Access Token before creating codespaces."
    echo "See: ~/codespaces/auth/tokens/README.md for instructions"
fi

# Exit successfully
exit 0