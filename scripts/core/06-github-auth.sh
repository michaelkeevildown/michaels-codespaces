#!/bin/bash

# GitHub authentication setup - Token-based for security
# Uses Personal Access Tokens instead of SSH keys

set -e

# Enable debug mode if DEBUG or VERBOSE environment variable is set
[[ "${DEBUG:-0}" == "1" || "${VERBOSE:-0}" == "1" ]] && set -x

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../utils/colors.sh"

echo_step "🔐 Configuring GitHub access..."

# Show debug info if enabled
if [[ "${DEBUG:-0}" == "1" || "${VERBOSE:-0}" == "1" ]]; then
    echo_debug "Debug mode enabled"
    echo_debug "Script directory: $SCRIPT_DIR"
    echo_debug "Home directory: $HOME"
    echo_debug "Current user: $USER"
fi

# Create directories for auth
mkdir -p ~/codespaces/auth/tokens
mkdir -p ~/codespaces/auth/git-config

# Check for existing token
TOKEN_FILE="$HOME/codespaces/auth/tokens/github.token"
if [ -f "$TOKEN_FILE" ] && [ -s "$TOKEN_FILE" ]; then
    echo_success "GitHub token already configured"
    
    if [[ "${DEBUG:-0}" == "1" || "${VERBOSE:-0}" == "1" ]]; then
        echo_debug "Found existing token file: $TOKEN_FILE"
        echo_debug "File size: $(stat -f%z "$TOKEN_FILE" 2>/dev/null || stat -c%s "$TOKEN_FILE" 2>/dev/null || echo 'unknown') bytes"
    fi
    
    # Display authenticated user
    if token=$(cat "$TOKEN_FILE" 2>/dev/null); then
        if [[ "${DEBUG:-0}" == "1" || "${VERBOSE:-0}" == "1" ]]; then
            echo_debug "Token loaded: ${token:0:7}...${token: -3}"
        fi
        # Create temp file for response
        temp_response=$(mktemp /tmp/github-verify-XXXXXX)
        http_code=$(curl -s --max-time 10 --connect-timeout 5 -H "Authorization: token $token" -H "Accept: application/vnd.github.v3+json" -w "%{http_code}" -o "$temp_response" https://api.github.com/user 2>&1)
        http_code=$(echo -n "$http_code" | tr -d '[:space:]' | tr -d '[:cntrl:]')
        
        if [[ "$http_code" == "200" ]]; then
            if command -v jq >/dev/null 2>&1; then
                username=$(jq -r '.login' < "$temp_response" 2>/dev/null)
            else
                username=$(grep -o '"login":"[^"]*"' < "$temp_response" | cut -d'"' -f4 2>/dev/null)
            fi
            if [[ -n "$username" ]]; then
                printf "${COLOR_BLUE}${SYMBOL_INFO}${COLOR_RESET}  Authenticated as: ${COLOR_BOLD}%s${COLOR_RESET}\n" "$username"
            fi
        else
            echo_warning "Token found but verification failed. You may need to update it."
        fi
        rm -f "$temp_response"
    fi
else
    # Check environment variable first
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        echo_debug "Using GitHub token from environment"
        echo "$GITHUB_TOKEN" > "$TOKEN_FILE"
        chmod 600 "$TOKEN_FILE"
        echo_success "GitHub token saved from environment"
        # Verify it works
        temp_response=$(mktemp /tmp/github-verify-env-XXXXXX)
        http_code=$(curl -s --max-time 10 --connect-timeout 5 -H "Authorization: token $GITHUB_TOKEN" -H "Accept: application/vnd.github.v3+json" -w "%{http_code}" -o "$temp_response" https://api.github.com/user 2>&1)
        http_code=$(echo -n "$http_code" | tr -d '[:space:]' | tr -d '[:cntrl:]')
        
        if [[ "$http_code" == "200" ]]; then
            if command -v jq >/dev/null 2>&1; then
                username=$(jq -r '.login' < "$temp_response" 2>/dev/null)
            else
                username=$(grep -o '"login":"[^"]*"' < "$temp_response" | cut -d'"' -f4 2>/dev/null)
            fi
            if [[ -n "$username" ]]; then
                printf "${COLOR_BLUE}${SYMBOL_INFO}${COLOR_RESET}  Authenticated as: ${COLOR_BOLD}%s${COLOR_RESET}\n" "$username"
            fi
        fi
        rm -f "$temp_response"
    else
        # Interactive token setup
        echo_box "GitHub Personal Access Token Setup" 50
        echo ""
        echo "To create codespaces, you need a GitHub token."
        echo ""
        echo -e "${COLOR_BOLD}Quick Setup:${COLOR_RESET}"
        echo ""
        echo -e "1. ${COLOR_BLUE}Open this URL:${COLOR_RESET}"
        echo -e "   ${COLOR_CYAN}${COLOR_UNDERLINE}https://github.com/settings/tokens/new${COLOR_RESET}"
        echo ""
        echo -e "2. ${COLOR_BLUE}Configure token:${COLOR_RESET}"
        echo -e "   • ${COLOR_BOLD}Note:${COLOR_RESET} Michael's Codespaces - $(hostname)"
        echo -e "   • ${COLOR_BOLD}Expiration:${COLOR_RESET} 90 days (recommended)"
        echo ""
        echo -e "   ${COLOR_BOLD}Select scopes - Check these boxes:${COLOR_RESET}"
        echo -e "   ${COLOR_GREEN}✓${COLOR_RESET} ${COLOR_BOLD}repo${COLOR_RESET} (Full control of private repositories)"
        echo -e "   ${COLOR_GREEN}✓${COLOR_RESET} ${COLOR_BOLD}workflow${COLOR_RESET} (Update GitHub Action workflows)"  
        echo -e "   ${COLOR_GREEN}✓${COLOR_RESET} ${COLOR_BOLD}write:packages${COLOR_RESET} (Upload packages to GitHub Package Registry)"
        echo ""
        echo -e "   ${COLOR_GRAY}Note: The 'repo' scope includes:${COLOR_RESET}"
        echo -e "   ${COLOR_GRAY}• repo:status, repo_deployment, public_repo, repo:invite, security_events${COLOR_RESET}"
        echo ""
        echo -e "3. ${COLOR_BLUE}Generate & copy token${COLOR_RESET} ${COLOR_GRAY}(starts with ghp_)${COLOR_RESET}"
        echo_box_end 50
        
        # Prompt for token with better guidance
        echo ""
        printf "${COLOR_BOLD}Ready to paste your token?${COLOR_RESET}\n"
        printf "${COLOR_GRAY}• Make sure you checked: repo, workflow, write:packages${COLOR_RESET}\n"
        printf "${COLOR_GRAY}• Token should start with 'ghp_' and be 40 characters long${COLOR_RESET}\n"
        echo ""
        
        while true; do
            echo -e "${COLOR_BOLD}Paste your GitHub token:${COLOR_RESET}"
            echo -e "${COLOR_GRAY}Tip: The token will be hidden as you type (like a password)${COLOR_RESET}"
            echo -n "> "
            
            # Read token (hidden while typing/pasting)
            IFS= read -rs GITHUB_TOKEN_INPUT
            echo ""  # Add newline after hidden input
            
            # Trim any whitespace or newlines immediately
            GITHUB_TOKEN_INPUT=$(echo -n "$GITHUB_TOKEN_INPUT" | tr -d '[:space:]')
            
            # Show what we captured (masked)
            if [ -n "$GITHUB_TOKEN_INPUT" ]; then
                # Create masked version showing first 7 and last 3 chars
                masked_token="${GITHUB_TOKEN_INPUT:0:7}********************************${GITHUB_TOKEN_INPUT: -3}"
                echo -e "\n${COLOR_GREEN}✓${COLOR_RESET} Token captured: $masked_token"
                echo -e "  Length: ${#GITHUB_TOKEN_INPUT} characters"
            fi
            
            if [ -z "$GITHUB_TOKEN_INPUT" ]; then
                echo_warning "Token is required for creating codespaces."
                echo ""
                echo "If you haven't created your token yet:"
                echo "1. Open: ${COLOR_CYAN}https://github.com/settings/tokens/new${COLOR_RESET}"
                echo "2. Check the 3 scopes mentioned above"
                echo "3. Click 'Generate token' and copy it"
                echo ""
                read -p "Do you want to skip token setup for now? [y/N] " -n 1 -r
                echo ""
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    echo_info "Skipping token setup. You'll need to set it before creating codespaces."
                    break
                else
                    echo "Let's try again..."
                    echo ""
                    continue
                fi
            fi
            
            # Validate token format
            if [[ "${DEBUG:-0}" == "1" || "${VERBOSE:-0}" == "1" ]]; then
                echo_debug "Token format validation..."
                echo_debug "Token length: ${#GITHUB_TOKEN_INPUT} (expected: 40)"
                echo_debug "Token prefix: '${GITHUB_TOKEN_INPUT:0:4}' (expected: 'ghp_')"
                echo_debug "Token suffix: '...${GITHUB_TOKEN_INPUT: -4}'"
                echo_debug "Token sample: ${GITHUB_TOKEN_INPUT:0:10}...${GITHUB_TOKEN_INPUT: -10}"
                
                # Check specific issues
                if [[ ! "$GITHUB_TOKEN_INPUT" =~ ^ghp_ ]]; then
                    echo_debug "ERROR: Token doesn't start with 'ghp_'"
                elif [[ ${#GITHUB_TOKEN_INPUT} -ne 40 ]]; then
                    echo_debug "ERROR: Token length is ${#GITHUB_TOKEN_INPUT}, not 40"
                elif [[ ! "$GITHUB_TOKEN_INPUT" =~ ^ghp_[a-zA-Z0-9]+$ ]]; then
                    echo_debug "ERROR: Token contains invalid characters"
                fi
            fi
            
            if [[ "$GITHUB_TOKEN_INPUT" =~ ^ghp_[a-zA-Z0-9]{36}$ ]]; then
                # Save token with proper handling
                printf "%s" "$GITHUB_TOKEN_INPUT" > "$TOKEN_FILE"
                chmod 600 "$TOKEN_FILE"
                
                # Ensure file is written to disk
                sync
                
                # Verify the token was saved correctly
                if [ -f "$TOKEN_FILE" ]; then
                    saved_token=$(cat "$TOKEN_FILE" 2>/dev/null)
                    if [ "$saved_token" != "$GITHUB_TOKEN_INPUT" ]; then
                        echo_error "Token save verification failed!"
                        echo_debug "Expected: ${#GITHUB_TOKEN_INPUT} chars"
                        echo_debug "Saved: ${#saved_token} chars"
                        rm -f "$TOKEN_FILE"
                        continue
                    fi
                else
                    echo_error "Failed to create token file!"
                    continue
                fi
                
                echo_success "GitHub token saved successfully!"
                
                # Small delay to ensure file operations complete
                sleep 0.1
                
                # Test the token
                echo_info "Verifying token with GitHub..."
                
                # Debug: Show what we're about to do
                if [[ "${DEBUG:-0}" == "1" || "${VERBOSE:-0}" == "1" ]]; then
                    echo_debug "Making API request to https://api.github.com/user"
                    echo_debug "Token format: ${GITHUB_TOKEN_INPUT:0:7}...${GITHUB_TOKEN_INPUT: -4}"
                fi
                
                # Make API call with timeout - use separate variables for response and status
                # Create temp file for response body
                response_file=$(mktemp /tmp/github-auth-response.XXXXXX)
                
                # Make the API call and capture HTTP status code
                # Note: We don't use spinner here to avoid interfering with curl output
                http_code_raw=$(curl -s --max-time 10 --connect-timeout 5 \
                    -H "Authorization: token $GITHUB_TOKEN_INPUT" \
                    -H "Accept: application/vnd.github.v3+json" \
                    -w "%{http_code}" \
                    -o "$response_file" \
                    https://api.github.com/user 2>&1)
                
                # Clean the HTTP code - remove any whitespace, newlines, or control characters
                http_code=$(echo -n "$http_code_raw" | tr -d '[:space:]' | tr -d '[:cntrl:]')
                
                # Read the JSON response from file
                json_response=$(cat "$response_file" 2>/dev/null || echo "{}")
                
                # Debug: Show raw response
                if [[ "${DEBUG:-0}" == "1" || "${VERBOSE:-0}" == "1" ]]; then
                    echo_debug "Raw HTTP status code: '$http_code_raw' (length: ${#http_code_raw})"
                    echo_debug "Cleaned HTTP status code: '$http_code' (length: ${#http_code})"
                    echo_debug "Response file: $response_file"
                    echo_debug "Response length: ${#json_response} characters"
                    echo_debug "First 100 chars of response: '${json_response:0:100}'"
                fi
                
                # Clean up temp file
                rm -f "$response_file"
                
                # Debug: Show comparison
                if [[ "${DEBUG:-0}" == "1" || "${VERBOSE:-0}" == "1" ]]; then
                    echo_debug "Comparing HTTP code '$http_code' with '200'"
                    if [[ "$http_code" == "200" ]]; then
                        echo_debug "HTTP code matches 200!"
                    else
                        echo_debug "HTTP code does NOT match 200"
                    fi
                fi
                
                if [[ "$http_code" == "200" ]]; then
                    # Try to extract username
                    if command -v jq >/dev/null 2>&1; then
                        username=$(echo "$json_response" | jq -r '.login' 2>/dev/null)
                    else
                        username=$(echo "$json_response" | grep -o '"login":"[^"]*"' | cut -d'"' -f4)
                    fi
                    
                    if [[ -n "$username" ]]; then
                        echo_success "Token verified - authentication working!"
                        printf "${COLOR_BLUE}${SYMBOL_INFO}${COLOR_RESET}  Authenticated as: ${COLOR_BOLD}%s${COLOR_RESET}\n" "$username"
                        break
                    else
                        echo_error "Token verified but couldn't parse username from response"
                        if [[ "${DEBUG:-0}" == "1" || "${VERBOSE:-0}" == "1" ]]; then
                            echo_debug "HTTP Code: $http_code"
                            echo_debug "Response (first 500 chars): ${json_response:0:500}"
                            echo_info "Tip: Install 'jq' for better JSON parsing: sudo apt-get install jq"
                        fi
                        break  # Still break as token is valid
                    fi
                elif [[ "$http_code" == "401" ]]; then
                    echo_error "Token verification failed: Invalid or expired token"
                    rm -f "$TOKEN_FILE"
                    echo ""
                    continue
                elif [[ "$http_code" == "403" ]]; then
                    echo_error "Token verification failed: Insufficient permissions"
                    echo_info "Make sure your token has the required scopes: repo, workflow, write:packages"
                    rm -f "$TOKEN_FILE"
                    echo ""
                    continue
                elif [[ "$http_code" == "000" ]] || [[ -z "$http_code" ]]; then
                    echo_error "Token verification failed: Connection timeout or network error"
                    
                    # Additional debugging for connection issues
                    if [[ "${DEBUG:-0}" == "1" || "${VERBOSE:-0}" == "1" ]]; then
                        echo_debug "Attempting direct connection test..."
                        if curl -s --max-time 5 --connect-timeout 3 https://api.github.com/zen >/dev/null 2>&1; then
                            echo_debug "GitHub API is reachable"
                        else
                            echo_debug "Cannot reach GitHub API"
                        fi
                    fi
                    
                    echo_info "Please check your internet connection and try again"
                    echo_info "Token saved but not verified. Retry or continue with setup."
                    echo ""
                    # Don't delete token on network errors - it might be valid
                    read -p "Retry verification? [Y/n] " -n 1 -r
                    echo ""
                    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
                        # Add small delay before retry to avoid rate limiting
                        sleep 1
                        continue
                    else
                        # Accept the token even though we couldn't verify it
                        echo_warning "Continuing with unverified token"
                        break
                    fi
                else
                    echo_error "Token verification failed: HTTP $http_code"
                    if [[ "${DEBUG:-0}" == "1" || "${VERBOSE:-0}" == "1" ]]; then
                        echo_debug "Full response:"
                        echo "$json_response" | head -20
                        echo_info "To see full debug output, run: DEBUG=1 $0"
                    fi
                    rm -f "$TOKEN_FILE"
                    echo ""
                    continue
                fi
            else
                echo_error "Invalid token format. GitHub tokens start with 'ghp_' followed by 36 characters."
                echo_info "Example: ghp_A1b2C3d4E5f6G7h8I9j0K1L2M3N4O5P6Q7R8"
                
                if [[ "${DEBUG:-0}" != "1" && "${VERBOSE:-0}" != "1" ]]; then
                    echo ""
                    echo_info "Run with DEBUG=1 to see what's wrong with your token format"
                fi
                
                echo ""
                echo_info "Alternative: Save your token to a file and run:"
                echo "  echo 'YOUR_TOKEN' > ~/codespaces/auth/tokens/github.token"
                echo "  chmod 600 ~/codespaces/auth/tokens/github.token"
                echo ""
                continue
            fi
        done
        
        # Always create instructions file
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
3. **Select scopes** - Check EXACTLY these boxes:

   **✅ repo** - Full control of private repositories
   This automatically includes:
   • repo:status - Access commit status
   • repo_deployment - Access deployment status  
   • public_repo - Access public repositories
   • repo:invite - Access repository invitations
   • security_events - Read and write security events

   **✅ workflow** - Update GitHub Action workflows
   Required to work with GitHub Actions and CI/CD

   **✅ write:packages** - Upload packages to GitHub Package Registry
   Needed for publishing packages and container images

   ${COLOR_YELLOW}Important: Only check these 3 main scopes. The sub-scopes under 'repo' are included automatically.${COLOR_RESET}
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