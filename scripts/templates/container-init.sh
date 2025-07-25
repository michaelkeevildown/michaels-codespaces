#!/bin/bash

# Container Initialization Script
# Runs on container startup to install and configure selected components

set -e

# Container init configuration
INIT_DIR="/home/coder/.codespace-init"
MANIFEST_FILE="$INIT_DIR/components.manifest"
LOG_FILE="$INIT_DIR/init.log"
INSTALLERS_DIR="/opt/codespace/components/installers"
INIT_MARKER="$INIT_DIR/.initialized"

# Logging functions
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

log_error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $*" | tee -a "$LOG_FILE" >&2
}

log_success() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS: $*" | tee -a "$LOG_FILE"
}

# Create initialization directory
init_setup() {
    mkdir -p "$INIT_DIR"
    touch "$LOG_FILE"
    
    log "Container initialization started"
    log "User: $(whoami)"
    log "Home: $HOME"
    log "Working directory: $(pwd)"
}

# Check if already initialized
check_initialized() {
    if [ -f "$INIT_MARKER" ]; then
        log "Container already initialized on $(cat "$INIT_MARKER")"
        return 0
    fi
    return 1
}

# Read component manifest
read_manifest() {
    if [ ! -f "$MANIFEST_FILE" ]; then
        log "No component manifest found at $MANIFEST_FILE"
        return 1
    fi
    
    log "Reading component manifest..."
    cat "$MANIFEST_FILE"
}

# Install a single component
install_component() {
    local component="$1"
    local installer="$INSTALLERS_DIR/${component}.sh"
    
    log "Installing component: $component"
    
    if [ ! -f "$installer" ]; then
        log_error "Installer not found: $installer"
        return 1
    fi
    
    # Make installer executable
    chmod +x "$installer"
    
    # Run installation
    if bash "$installer" install >> "$LOG_FILE" 2>&1; then
        log_success "Installed: $component"
        
        # Run configuration
        if bash "$installer" configure >> "$LOG_FILE" 2>&1; then
            log_success "Configured: $component"
        else
            log_error "Configuration failed: $component"
        fi
        
        # Run verification
        if bash "$installer" verify >> "$LOG_FILE" 2>&1; then
            log_success "Verified: $component"
        else
            log_error "Verification failed: $component"
        fi
    else
        log_error "Installation failed: $component"
        return 1
    fi
}

# Install all components from manifest
install_components() {
    local components
    
    if ! components=$(read_manifest); then
        log "No components to install"
        return 0
    fi
    
    log "Components to install: $components"
    
    # Install each component
    local failed=0
    for component in $components; do
        if ! install_component "$component"; then
            ((failed++))
        fi
    done
    
    if [ $failed -gt 0 ]; then
        log_error "$failed component(s) failed to install"
        return 1
    else
        log_success "All components installed successfully"
        return 0
    fi
}

# Post-installation setup
post_install_setup() {
    log "Running post-installation setup..."
    
    # Update PATH if needed
    if [ -d "/usr/local/bin" ] && [[ ":$PATH:" != *":/usr/local/bin:"* ]]; then
        export PATH="/usr/local/bin:$PATH"
        echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
    fi
    
    # Source bashrc to pick up new aliases and functions
    if [ -f ~/.bashrc ]; then
        source ~/.bashrc
    fi
    
    # Create workspace symlink if needed
    local workspace="/home/coder/workspace"
    if [ ! -e "$workspace" ] && [ -d "/home/coder/$(basename "$(pwd)")" ]; then
        ln -s "/home/coder/$(basename "$(pwd)")" "$workspace"
        log "Created workspace symlink"
    fi
    
    # Set up git safe directory
    if command -v git >/dev/null 2>&1; then
        git config --global --add safe.directory /home/coder/workspace 2>/dev/null || true
        git config --global --add safe.directory "$(pwd)" 2>/dev/null || true
    fi
    
    log_success "Post-installation setup completed"
}

# Display initialization summary
display_summary() {
    echo ""
    echo "=== Container Initialization Summary ==="
    echo ""
    
    # Check installed components
    if [ -f "$MANIFEST_FILE" ]; then
        echo "Installed Components:"
        while read -r component; do
            local name=""
            local installer="$INSTALLERS_DIR/${component}.sh"
            
            if [ -f "$installer" ]; then
                name=$(bash "$installer" metadata | grep "^name=" | cut -d= -f2)
            fi
            
            echo "  • ${name:-$component}"
        done < "$MANIFEST_FILE"
        echo ""
    fi
    
    # Show useful commands
    echo "Useful Commands:"
    
    if command -v gh >/dev/null 2>&1; then
        echo "  • gh auth status - Check GitHub authentication"
    fi
    
    if command -v claude >/dev/null 2>&1; then
        echo "  • claude --help - Claude AI assistant"
    fi
    
    if command -v claude-flow >/dev/null 2>&1; then
        echo "  • claude-flow --help - AI orchestration tool"
    fi
    
    echo ""
    echo "Log file: $LOG_FILE"
    echo ""
}

# Main initialization function
main() {
    # Setup
    init_setup
    
    # Check if already initialized
    if check_initialized; then
        log "Skipping initialization - already completed"
        return 0
    fi
    
    # Install components
    if install_components; then
        # Run post-install setup
        post_install_setup
        
        # Mark as initialized
        date > "$INIT_MARKER"
        log_success "Container initialization completed"
        
        # Display summary
        display_summary
    else
        log_error "Container initialization failed"
        return 1
    fi
}

# Run main function
main "$@"