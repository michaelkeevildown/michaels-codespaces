#!/bin/bash

# MCS Go Installation Script
# Installs the Go version of Michael's Codespaces
# Philosophy: Build from source by default, with pre-built binary as fallback

set -e

# Colors
if [[ -t 1 ]]; then
    RED=$(printf '\033[31m')
    GREEN=$(printf '\033[32m')
    YELLOW=$(printf '\033[33m')
    BLUE=$(printf '\033[34m')
    BOLD=$(printf '\033[1m')
    RESET=$(printf '\033[0m')
else
    RED='' GREEN='' YELLOW='' BLUE='' BOLD='' RESET=''
fi

# Helper functions
info() {
    printf "${BLUE}==>${RESET} ${BOLD}%s${RESET}\n" "$1"
}

success() {
    printf "${GREEN}✓${RESET} %s\n" "$1"
}

warning() {
    printf "${YELLOW}⚠${RESET}  %s\n" "$1"
}

error() {
    printf "${RED}✗${RESET} %s\n" "$1" >&2
}

# Display MCS ASCII art header
show_header() {
    local header_color="${BLUE}${BOLD}"
    
    printf "${header_color}"
    cat << 'EOF'
  __  __  ___ ___ 
 |  \/  |/ __/ __|
 | |\/| | (__\__ \
 |_|  |_|\___|___/
                  
Michael's Codespaces
EOF
    printf "${RESET}\n\n"
}

# Configuration
MCS_HOME="${MCS_HOME:-$HOME/.mcs}"
REPO_URL="${MCS_REPO_URL:-https://github.com/michaelkeevildown/michaels-codespaces.git}"
BRANCH="${MCS_BRANCH:-main}"
INSTALL_MODE="${MCS_INSTALL_MODE:-auto}"  # auto, source, release, dev

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    case "$OS" in
        linux|darwin)
            PLATFORM="${OS}-${ARCH}"
            ;;
        *)
            error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
}

# Check for required tools
check_requirements() {
    local missing=()
    local docker_missing=false
    
    # Check for Go (optional - for building from source)
    if ! command -v go >/dev/null 2>&1; then
        GO_AVAILABLE=false
    else
        GO_AVAILABLE=true
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    fi
    
    # Check for Docker (will be installed by MCS setup if missing)
    if ! command -v docker >/dev/null 2>&1; then
        docker_missing=true
        warning "Docker not found - MCS setup will install it for you"
    fi
    
    # Check for Git (required for cloning)
    if ! command -v git >/dev/null 2>&1; then
        missing+=("git")
    fi
    
    # Only fail if we're missing Git (which we need to clone the repo)
    if [ ${#missing[@]} -gt 0 ]; then
        error "Missing required tools: ${missing[*]}"
        error "Please install them before running this installer"
        exit 1
    fi
    
    # Set a flag for Docker installation
    DOCKER_MISSING=$docker_missing
}

# Clone or update repository
clone_or_update_repo() {
    # Check if we need authentication
    local clone_url="$REPO_URL"
    
    # If GitHub token is provided, use it
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        info "Using GitHub token for authentication..."
        # Extract repo path from URL
        repo_path=$(echo "$REPO_URL" | sed 's|https://github.com/||')
        clone_url="https://token:${GITHUB_TOKEN}@github.com/${repo_path}"
    fi
    
    if [ -d "$MCS_HOME/.git" ]; then
        info "Updating existing MCS installation..."
        cd "$MCS_HOME"
        
        # Update remote URL if token is provided
        if [ -n "${GITHUB_TOKEN:-}" ]; then
            git remote set-url origin "$clone_url" 2>/dev/null || true
        fi
        
        git pull origin "$BRANCH" || {
            warning "Failed to update repository. Continuing with existing code..."
        }
        
        # Reset remote URL to remove token
        if [ -n "${GITHUB_TOKEN:-}" ]; then
            git remote set-url origin "$REPO_URL" 2>/dev/null || true
        fi
    else
        info "Cloning MCS repository..."
        # Backup existing directory if it exists
        if [ -d "$MCS_HOME" ]; then
            warning "Backing up existing MCS directory..."
            
            # Try to use Go backup command if mcs is already installed
            if command -v mcs >/dev/null 2>&1 && mcs backup --help >/dev/null 2>&1; then
                info "Using MCS backup system..."
                if mcs backup create --type install --source "$MCS_HOME" --description "Pre-update backup"; then
                    info "Backup created successfully"
                else
                    # Fallback to manual backup
                    warning "MCS backup failed, using manual backup..."
                    backup_dir="$HOME/.mcs.backup"
                    mkdir -p "$backup_dir"
                    backup_id="install-$(date +%Y%m%d_%H%M%S)"
                    mv "$MCS_HOME" "$backup_dir/$backup_id"
                    info "Manual backup created: $backup_dir/$backup_id"
                fi
            else
                # No mcs command available, use manual backup
                backup_dir="$HOME/.mcs.backup"
                mkdir -p "$backup_dir"
                backup_id="install-$(date +%Y%m%d_%H%M%S)"
                mv "$MCS_HOME" "$backup_dir/$backup_id"
                info "Manual backup created: $backup_dir/$backup_id"
            fi
        fi
        
        # Clone with branch support
        info "Cloning from: $REPO_URL"
        info "Branch: $BRANCH"
        if git clone -b "$BRANCH" "$clone_url" "$MCS_HOME"; then
            # Remove token from remote URL after successful clone
            if [ -n "${GITHUB_TOKEN:-}" ]; then
                cd "$MCS_HOME"
                git remote set-url origin "$REPO_URL"
            fi
        else
            # If clone fails, check if it's a public repo
            warning "Clone failed. Trying without authentication..."
            if ! git clone -b "$BRANCH" "$REPO_URL" "$MCS_HOME"; then
                error "Failed to clone repository"
                error ""
                error "If this is a private repository, please set GITHUB_TOKEN:"
                error "  export GITHUB_TOKEN='your-github-token'"
                error "  curl -fsSL install-script-url | bash"
                exit 1
            fi
        fi
    fi
}

# Setup PATH configuration
setup_path_config() {
    local configs_updated=0
    local path_line="export PATH=\"$BIN_DIR:\$PATH\""
    local mcs_comment="# MCS - Michael's Codespaces"
    
    # Function to add PATH to a config file
    add_to_config() {
        local config_file="$1"
        local file_desc="$2"
        
        if [ -f "$config_file" ] || [ "$config_file" = "$HOME/.bashrc" ] || [ "$config_file" = "$HOME/.profile" ]; then
            if ! grep -q "$BIN_DIR" "$config_file" 2>/dev/null; then
                info "Adding MCS to PATH in $file_desc..."
                {
                    echo ""
                    echo "$mcs_comment"
                    echo "$path_line"
                } >> "$config_file"
                configs_updated=$((configs_updated + 1))
                success "Updated $config_file"
            else
                info "MCS already in PATH in $file_desc"
            fi
        fi
    }
    
    # Detect current shell
    local current_shell=""
    if [ -n "$BASH_VERSION" ]; then
        current_shell="bash"
    elif [ -n "$ZSH_VERSION" ]; then
        current_shell="zsh"
    else
        case "$SHELL" in
            */bash) current_shell="bash" ;;
            */zsh) current_shell="zsh" ;;
            */sh) current_shell="sh" ;;
        esac
    fi
    
    # Update shell-specific configs
    case "$current_shell" in
        bash)
            add_to_config "$HOME/.bashrc" "~/.bashrc"
            # Also update .profile for login shells
            add_to_config "$HOME/.profile" "~/.profile"
            ;;
        zsh)
            add_to_config "$HOME/.zshrc" "~/.zshrc"
            add_to_config "$HOME/.zprofile" "~/.zprofile"
            ;;
        *)
            # Fallback: update common files
            add_to_config "$HOME/.profile" "~/.profile"
            add_to_config "$HOME/.bashrc" "~/.bashrc"
            ;;
    esac
    
    # Also check for .bash_profile on macOS
    if [ -f "$HOME/.bash_profile" ]; then
        add_to_config "$HOME/.bash_profile" "~/.bash_profile"
    fi
    
    if [ $configs_updated -gt 0 ]; then
        echo ""
        success "PATH configuration updated in $configs_updated file(s)"
        echo ""
        warning "MCS has been added to your PATH, but won't be available until you:"
        echo ""
        echo "  Option 1: Start a new terminal session"
        echo "  Option 2: Run one of these commands:"
        case "$current_shell" in
            bash) echo "    source ~/.bashrc" ;;
            zsh) echo "    source ~/.zshrc" ;;
            *) echo "    source ~/.profile" ;;
        esac
        echo ""
    else
        success "MCS is already in your PATH configuration"
    fi
    
    # For this session, PATH is already set by the main script
    info "MCS is available in the current session at: $BIN_DIR/mcs"
}

# Download from GitHub releases
download_from_release() {
    local release_type="${1:-latest}"  # latest, dev, or specific version
    
    info "Downloading MCS from GitHub releases..."
    
    # Determine the release URL
    local release_api_url
    local release_name
    
    if [ "$release_type" = "dev" ]; then
        release_api_url="https://api.github.com/repos/michaelkeevildown/michaels-codespaces/releases/tags/dev-latest"
        release_name="development"
    elif [ "$release_type" = "latest" ]; then
        release_api_url="https://api.github.com/repos/michaelkeevildown/michaels-codespaces/releases/latest"
        release_name="latest stable"
    else
        # Specific version
        release_api_url="https://api.github.com/repos/michaelkeevildown/michaels-codespaces/releases/tags/${release_type}"
        release_name="$release_type"
    fi
    
    info "Fetching $release_name release information..."
    
    # Get release data
    local release_data
    release_data=$(curl -fsSL "$release_api_url" 2>/dev/null) || {
        error "Failed to fetch release information"
        return 1
    }
    
    # Extract download URL for our platform
    local download_url
    local archive_name="mcs-${OS}-${ARCH}"
    
    if [ "$OS" = "windows" ]; then
        archive_name="${archive_name}.zip"
    else
        archive_name="${archive_name}.tar.gz"
    fi
    
    download_url=$(echo "$release_data" | grep -o "\"browser_download_url\": *\"[^\"]*${archive_name}\"" | cut -d'"' -f4 | head -1)
    
    if [ -z "$download_url" ]; then
        error "No release found for platform: ${OS}-${ARCH}"
        return 1
    fi
    
    # Download the archive
    local temp_dir=$(mktemp -d)
    local archive_path="$temp_dir/$archive_name"
    
    info "Downloading from: $download_url"
    if ! curl -fsSL "$download_url" -o "$archive_path"; then
        error "Failed to download release"
        rm -rf "$temp_dir"
        return 1
    fi
    
    # Extract the binary
    info "Extracting binary..."
    cd "$temp_dir"
    
    if [ "$OS" = "windows" ]; then
        unzip -q "$archive_name" || {
            error "Failed to extract archive"
            rm -rf "$temp_dir"
            return 1
        }
    else
        tar -xzf "$archive_name" || {
            error "Failed to extract archive"
            rm -rf "$temp_dir"
            return 1
        }
    fi
    
    # Find and install the binary
    local binary_name="mcs"
    if [ "$OS" = "windows" ]; then
        binary_name="mcs.exe"
    fi
    
    if [ -f "$binary_name" ]; then
        chmod +x "$binary_name"
        mv "$binary_name" "$BIN_DIR/mcs"
        success "Downloaded and installed MCS $release_name"
        
        # Get version info
        local version=$("$BIN_DIR/mcs" version 2>/dev/null || echo "unknown")
        info "Installed version: $version"
    else
        error "Binary not found in archive"
        rm -rf "$temp_dir"
        return 1
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
    return 0
}

# Build from source
build_from_source() {
    cd "$MCS_HOME/mcs-go"
    
    info "Downloading dependencies..."
    go mod download || {
        error "Failed to download Go dependencies"
        exit 1
    }
    
    info "Building MCS..."
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')
    go build -ldflags "-X main.version=$VERSION" \
        -o "$BIN_DIR/mcs" cmd/mcs/main.go || {
        error "Failed to build MCS"
        exit 1
    }
    
    # Ensure binary is executable
    chmod +x "$BIN_DIR/mcs" || {
        error "Failed to make MCS executable"
        exit 1
    }
    
    # Verify binary exists
    if [ ! -f "$BIN_DIR/mcs" ]; then
        error "MCS binary not found at $BIN_DIR/mcs after build"
        exit 1
    fi
    
    success "Built MCS from source (version: $VERSION)"
}

# Main installation
main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dev)
                INSTALL_MODE="dev"
                shift
                ;;
            --source)
                INSTALL_MODE="source"
                shift
                ;;
            --release)
                INSTALL_MODE="release"
                shift
                ;;
            --version)
                INSTALL_MODE="version"
                INSTALL_VERSION="$2"
                shift 2
                ;;
            *)
                error "Unknown option: $1"
                echo "Usage: $0 [--dev|--source|--release|--version VERSION]"
                exit 1
                ;;
        esac
    done
    
    # Show beautiful header
    show_header
    
    info "Installing Michael's Codespaces (Go version)..."
    echo ""
    
    # Display installation mode
    case "$INSTALL_MODE" in
        dev)
            echo "Installation mode: Development (latest commit)"
            ;;
        source)
            echo "Installation mode: Build from source"
            ;;
        release)
            echo "Installation mode: Latest stable release"
            ;;
        version)
            echo "Installation mode: Specific version ($INSTALL_VERSION)"
            ;;
        auto)
            echo "Installation mode: Auto-detect (release → source)"
            ;;
    esac
    
    echo "Repository: $REPO_URL"
    if [ -n "$BRANCH" ] && [ "$BRANCH" != "main" ]; then
        echo "Branch: $BRANCH"
    fi
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        echo "Authentication: Using provided GitHub token"
    fi
    echo ""
    echo "Note: MCS will install and configure Docker if not present"
    echo ""
    
    # Detect platform
    detect_platform
    info "Detected platform: $PLATFORM"
    
    # Check requirements
    check_requirements
    
    # Set installation directory
    INSTALL_DIR="$MCS_HOME"
    BIN_DIR="$INSTALL_DIR/bin"
    
    # Create directories
    info "Creating installation directories..."
    mkdir -p "$BIN_DIR"
    mkdir -p "$HOME/codespaces"
    
    # Determine installation method based on mode
    local install_success=false
    
    case "$INSTALL_MODE" in
        auto)
            # Try release first, then source
            info "Attempting to download latest release..."
            if download_from_release "latest"; then
                install_success=true
            else
                warning "Release download failed, trying to build from source..."
                if [ "$GO_AVAILABLE" = true ]; then
                    clone_or_update_repo
                    if build_from_source; then
                        install_success=true
                    fi
                else
                    error "Go compiler not found and release download failed"
                fi
            fi
            ;;
            
        dev)
            # Download development build
            if download_from_release "dev"; then
                install_success=true
            else
                error "Failed to download development build"
            fi
            ;;
            
        release)
            # Download latest stable release
            if download_from_release "latest"; then
                install_success=true
            else
                error "Failed to download latest release"
            fi
            ;;
            
        version)
            # Download specific version
            if download_from_release "$INSTALL_VERSION"; then
                install_success=true
            else
                error "Failed to download version $INSTALL_VERSION"
            fi
            ;;
            
        source)
            # Build from source
            if [ "$GO_AVAILABLE" = true ]; then
                clone_or_update_repo
                if build_from_source; then
                    install_success=true
                fi
            else
                error "Go compiler not found - cannot build from source"
                echo ""
                echo "Install Go first:"
                echo "  - Ubuntu/Debian: sudo apt install golang-go"
                echo "  - macOS: brew install go"
                echo "  - Or download from: https://go.dev/dl/"
            fi
            ;;
    esac
    
    if [ "$install_success" = false ]; then
        error "Installation failed"
        exit 1
    fi
    
    # Create shell completion
    info "Setting up shell completion..."
    "$BIN_DIR/mcs" completion bash > "$INSTALL_DIR/mcs.bash" 2>/dev/null || true
    "$BIN_DIR/mcs" completion zsh > "$INSTALL_DIR/mcs.zsh" 2>/dev/null || true
    
    # PATH is already configured by setup_path_config()
    
    # Create update script
    info "Creating update script..."
    cat > "$BIN_DIR/mcs-update.sh" << 'EOF'
#!/bin/bash
# MCS Update Script
set -e

MCS_HOME="${MCS_HOME:-$HOME/.mcs}"
BIN_DIR="$MCS_HOME/bin"

echo "Updating MCS from source..."

# Pull latest changes
cd "$MCS_HOME"
git pull origin main || {
    echo "Failed to pull latest changes"
    exit 1
}

# Rebuild from source
cd "$MCS_HOME/mcs-go"
go mod download
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')
go build -ldflags "-X main.version=$VERSION" -o "$BIN_DIR/mcs" cmd/mcs/main.go || {
    echo "Failed to build MCS"
    exit 1
}

echo "Successfully updated MCS to version: $VERSION"

# Show changes
echo ""
echo "Recent changes:"
git log --oneline -10
EOF
    chmod +x "$BIN_DIR/mcs-update.sh"
    
    # Add to PATH temporarily for this session BEFORE testing
    export PATH="$BIN_DIR:$PATH"
    info "Added $BIN_DIR to PATH for current session"
    
    # Test installation
    info "Testing MCS installation..."
    
    # Check if binary exists
    if [ ! -f "$BIN_DIR/mcs" ]; then
        error "MCS binary not found at: $BIN_DIR/mcs"
        error "Installation failed - binary was not created"
        exit 1
    fi
    
    # Check if binary is executable
    if [ ! -x "$BIN_DIR/mcs" ]; then
        error "MCS binary is not executable"
        info "Attempting to fix permissions..."
        chmod +x "$BIN_DIR/mcs"
    fi
    
    # Verify PATH contains our directory
    if ! echo "$PATH" | grep -q "$BIN_DIR"; then
        error "Failed to add $BIN_DIR to PATH"
        error "Current PATH: $PATH"
        exit 1
    fi
    
    # Test running the binary (using --version flag for Cobra)
    if ! "$BIN_DIR/mcs" --version >/dev/null 2>&1; then
        error "MCS binary exists but failed to run"
        error "Trying to diagnose the issue..."
        
        # Try to get more information about the failure
        echo "Binary location: $BIN_DIR/mcs"
        echo "File info: $(file "$BIN_DIR/mcs" 2>&1 || echo 'file command not available')"
        echo "Permissions: $(ls -la "$BIN_DIR/mcs")"
        echo ""
        echo "Attempting to run with error output:"
        "$BIN_DIR/mcs" --version 2>&1 || true
        echo ""
        
        error "Installation completed but MCS test failed"
        echo ""
        error "TROUBLESHOOTING STEPS:"
        echo "1. Check if the binary works directly:"
        echo "   $BIN_DIR/mcs --version"
        echo ""
        echo "2. If that works, the issue is PATH related. Run:"
        echo "   export PATH=\"$BIN_DIR:\$PATH\""
        echo "   mcs --version"
        echo ""
        echo "3. If the binary doesn't work directly, possible causes:"
        echo "   - Architecture mismatch (built for wrong platform)"
        echo "   - Missing shared libraries (try: ldd $BIN_DIR/mcs)"
        echo "   - Go build issues (try rebuilding with: cd $MCS_HOME/mcs-go && go build -o $BIN_DIR/mcs cmd/mcs/main.go)"
        echo ""
        echo "4. For immediate use without PATH issues:"
        echo "   alias mcs='$BIN_DIR/mcs'"
        echo ""
        exit 1
    fi
    
    success "MCS installed successfully!"
    echo ""
    
    # Show header again for successful completion
    show_header
    
    info "Version: $("$BIN_DIR/mcs" --version 2>/dev/null || echo "unknown")"
    echo ""
    info "Installation details:"
    echo "  Location: $MCS_HOME"
    echo "  Binary: $BIN_DIR/mcs"
    if [ "$GO_AVAILABLE" = true ]; then
        echo "  Built from: source"
    else
        echo "  Built from: pre-built binary"
    fi
    echo ""
    
    # Check if mcs is available in PATH
    if command -v mcs >/dev/null 2>&1; then
        success "✓ MCS is available in your PATH for this session"
    else
        warning "⚠ MCS is not yet in your PATH"
        echo "  Run this command to use MCS immediately:"
        echo "    export PATH=\"$BIN_DIR:\$PATH\""
    fi
    echo ""
    
    # Setup PATH for future sessions
    setup_path_config
    
    # Run setup to configure network and other settings
    info "Running initial setup..."
    if [ "$DOCKER_MISSING" = true ]; then
        echo ""
        info "Setup will install Docker and configure your environment"
    fi
    echo ""
    if "$BIN_DIR/mcs" setup --bootstrap; then
        echo ""
        success "Setup completed!"
    else
        warning "Setup encountered issues. You can run 'mcs setup' later to reconfigure."
    fi
    
    echo ""
    info "Get started with:"
    echo "  mcs create github.com/user/repo"
    echo "  mcs list"
    echo "  mcs --help"
    echo ""
    info "Update MCS:"
    if [ "$GO_AVAILABLE" = true ]; then
        echo "  mcs update                    # Auto-update via git pull + rebuild"
        echo "  $BIN_DIR/mcs-update.sh       # Manual update script"
    else
        echo "  Install Go first, then run: mcs update"
    fi
}

# Run main function
main "$@"