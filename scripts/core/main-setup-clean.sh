#!/bin/bash

# Main setup orchestrator - Homebrew-style reliable UI
# Simple, direct execution with clean formatting

set -e

# Error handler
error_handler() {
    local line_no=$1
    local bash_lineno=$2
    local last_command=$3
    echo ""
    echo "ERROR: Script failed at line $line_no"
    echo "Command: $last_command"
    echo "Exit code: $?"
    echo ""
    echo "Debug info:"
    echo "  NONINTERACTIVE: ${NONINTERACTIVE:-not set}"
    echo "  Terminal: stdin=$([ -t 0 ] && echo "yes" || echo "no"), stdout=$([ -t 1 ] && echo "yes" || echo "no")"
    exit 1
}

# Set error trap
trap 'error_handler ${LINENO} ${BASH_LINENO} "$BASH_COMMAND"' ERR

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Debug mode
if [ "${DEBUG:-0}" = "1" ]; then
    set -x
    echo "DEBUG: Script starting"
    echo "DEBUG: NONINTERACTIVE=${NONINTERACTIVE:-not set}"
    echo "DEBUG: Terminal detection: stdin=$([ -t 0 ] && echo "yes" || echo "no"), stdout=$([ -t 1 ] && echo "yes" || echo "no")"
fi

# Source utilities
if [ ! -f "$SCRIPT_DIR/../utils/colors.sh" ]; then
    echo "Error: Cannot find colors.sh at $SCRIPT_DIR/../utils/colors.sh"
    exit 1
fi
source "$SCRIPT_DIR/../utils/colors.sh"

if [ ! -f "$SCRIPT_DIR/../utils/checks.sh" ]; then
    echo "Error: Cannot find checks.sh at $SCRIPT_DIR/../utils/checks.sh"
    exit 1
fi
source "$SCRIPT_DIR/../utils/checks.sh"

# Homebrew-style output functions
ohai() {
    printf "${COLOR_BLUE}==>${COLOR_RESET} ${COLOR_BOLD}%s${COLOR_RESET}\n" "$@"
}

opoo() {
    printf "${COLOR_YELLOW}Warning${COLOR_RESET}: %s\n" "$@"
}

onoe() {
    printf "${COLOR_RED}Error${COLOR_RESET}: %s\n" "$@" >&2
}

# Setup modules
SETUP_MODULES=(
    "01-system-packages.sh:System packages"
    "02-docker-setup.sh:Docker"
    "03-shell-setup.sh:Shell environment"
    "04-codespace-infrastructure.sh:Codespace infrastructure"
    "05-monitoring-tools.sh:Monitoring tools"
    "06-github-auth.sh:GitHub authentication"
)

# Homebrew-style module execution
run_module() {
    local module_file="$1"
    local module_name="$2"
    local module_num="$3"
    local total_modules="$4"
    
    # Show what we're doing
    ohai "[$module_num/$total_modules] $module_name"
    
    # Run the module directly
    if (cd "$SCRIPT_DIR" && "./$module_file"); then
        echo "${COLOR_GREEN}${SYMBOL_CHECK}${COLOR_RESET} $module_name completed"
        echo ""
    else
        onoe "$module_name failed"
        exit 1
    fi
}

# Start setup
echo ""
printf "${COLOR_BOLD}Michael's Codespaces Installer${COLOR_RESET}\n"
echo "=============================="
echo ""

# Debug environment
if [ "${DEBUG:-0}" = "1" ]; then
    echo "DEBUG: Environment check"
    echo "  PWD: $PWD"
    echo "  SCRIPT_DIR: $SCRIPT_DIR"
    echo "  PATH: $PATH"
    echo "  Shell: $SHELL"
    echo ""
fi

# Show what will be installed
ohai "This script will install:"
for i in "${!SETUP_MODULES[@]}"; do
    IFS=':' read -r module_file module_desc <<< "${SETUP_MODULES[$i]}"
    printf "  ${COLOR_DIM}•${COLOR_RESET} %s\n" "$module_desc"
done
echo ""

# Simple confirmation
if [ "${NONINTERACTIVE:-}" != "1" ]; then
    # Check if we can actually read from stdin
    if [ -t 0 ]; then
        printf "Press ${COLOR_BOLD}RETURN${COLOR_RESET} to continue or any other key to abort: "
        # Use timeout to prevent hanging
        if command -v timeout >/dev/null 2>&1; then
            timeout 60 bash -c 'read -r REPLY; echo "$REPLY"' || REPLY=""
        else
            read -r REPLY || REPLY=""
        fi
        
        if [ -n "$REPLY" ]; then
            onoe "Installation aborted"
            exit 1
        fi
        echo ""
    else
        # No terminal available, run non-interactively
        ohai "No terminal detected, running non-interactively"
    fi
else
    ohai "Running in non-interactive mode"
fi

# Run each module
module_num=0
total_modules=${#SETUP_MODULES[@]}

for module_info in "${SETUP_MODULES[@]}"; do
    IFS=':' read -r module_file module_desc <<< "$module_info"
    ((module_num++))
    
    # Check if module file exists and is executable
    if [ ! -f "$SCRIPT_DIR/$module_file" ]; then
        onoe "Module not found: $SCRIPT_DIR/$module_file"
        exit 1
    fi
    
    if [ ! -x "$SCRIPT_DIR/$module_file" ]; then
        chmod +x "$SCRIPT_DIR/$module_file"
    fi
    
    # Run the module
    run_module "$module_file" "$module_desc" "$module_num" "$total_modules"
done

# Success message
echo ""
printf "${COLOR_GREEN}${SYMBOL_CHECK}${COLOR_RESET} ${COLOR_BOLD}Installation completed successfully!${COLOR_RESET}\n"
echo ""

ohai "Michael's Codespaces is now installed!"
echo ""
echo "Next steps:"
echo "  ${COLOR_DIM}•${COLOR_RESET} Logout and login again for Docker permissions"
echo "  ${COLOR_DIM}•${COLOR_RESET} Run ${COLOR_BOLD}mcs doctor${COLOR_RESET} to check system health"
echo "  ${COLOR_DIM}•${COLOR_RESET} Create your first codespace with ${COLOR_BOLD}mcs create <repo-url>${COLOR_RESET}"
echo ""