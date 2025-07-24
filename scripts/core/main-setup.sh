#!/bin/bash

# Main setup orchestrator - Homebrew-style
# This script calls all the modular components in the correct order

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source utilities
source "$SCRIPT_DIR/../utils/colors.sh"
source "$SCRIPT_DIR/../utils/checks.sh"

# Setup modules
SETUP_MODULES=(
    "01-system-packages.sh:Installing system packages"
    "02-docker-setup.sh:Setting up Docker"
    "03-shell-setup.sh:Configuring shell environment"
    "04-codespace-infrastructure.sh:Creating codespace infrastructure"
    "05-monitoring-tools.sh:Installing monitoring tools"
    "06-github-auth.sh:Configuring GitHub access"
)

# Start setup
echo ""
echo_box "${SYMBOL_ROCKET} Michael's Codespaces Setup"
echo ""

# Initialize progress
init_progress ${#SETUP_MODULES[@]}

# Run each module
for module_info in "${SETUP_MODULES[@]}"; do
    IFS=':' read -r module_file module_desc <<< "$module_info"
    
    step_progress "$module_desc"
    echo ""
    
    # Run the module and show output
    if ! "$SCRIPT_DIR/$module_file"; then
        echo_error "Failed during: $module_desc"
        exit 1
    fi
    
    echo ""
done

echo_box_end

# Success message
echo_success "${SYMBOL_CHECK} Setup completed successfully!"
echo ""

# Show summary
echo_step "Summary"
echo_list_item "•" "System packages installed"
echo_list_item "•" "Docker configured and running"
echo_list_item "•" "Zsh with Oh My Zsh installed"
echo_list_item "•" "Codespace infrastructure created"
echo_list_item "•" "Monitoring tools installed"
echo_list_item "•" "GitHub SSH key generated"
echo ""