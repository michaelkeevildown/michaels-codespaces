#!/bin/bash

# Main setup orchestrator - Clean, dynamic UI
# Shows progress without cluttering the terminal

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Debug mode
if [ "${DEBUG:-0}" = "1" ]; then
    set -x
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

# Setup modules
SETUP_MODULES=(
    "01-system-packages.sh:System packages"
    "02-docker-setup.sh:Docker"
    "03-shell-setup.sh:Shell environment"
    "04-codespace-infrastructure.sh:Codespace infrastructure"
    "05-monitoring-tools.sh:Monitoring tools"
    "06-github-auth.sh:GitHub authentication"
)

# Function to run module with clean output
run_module_clean() {
    local module_file="$1"
    local module_name="$2"
    local module_num="$3"
    local total_modules="$4"
    
    # Create a temporary file for output
    local output_file="/tmp/mcs-setup-${module_file}.log"
    local status_file="/tmp/mcs-setup-${module_file}.status"
    
    # Start the module in background
    (
        "$SCRIPT_DIR/$module_file" > "$output_file" 2>&1
        echo $? > "$status_file"
    ) &
    
    local pid=$!
    
    # Show progress while module runs
    while kill -0 $pid 2>/dev/null; do
        for frame in "${SPINNER_FRAMES[@]}"; do
            printf "\r${COLOR_DIM}[%d/%d]${COLOR_RESET} ${COLOR_CYAN}%s${COLOR_RESET} Installing %s${SYMBOL_ELLIPSIS}   " \
                "$module_num" "$total_modules" "$frame" "$module_name"
            sleep 0.08
        done
    done
    
    # Check if module succeeded
    wait $pid
    local exit_code=$(cat "$status_file" 2>/dev/null || echo 1)
    rm -f "$status_file"
    
    if [ $exit_code -eq 0 ]; then
        printf "\r${COLOR_DIM}[%d/%d]${COLOR_RESET} ${COLOR_GREEN}${SYMBOL_CHECK}${COLOR_RESET} %s installed                     \n" \
            "$module_num" "$total_modules" "$module_name"
    else
        printf "\r${COLOR_DIM}[%d/%d]${COLOR_RESET} ${COLOR_RED}${SYMBOL_CROSS}${COLOR_RESET} %s failed                        \n" \
            "$module_num" "$total_modules" "$module_name"
        echo ""
        echo_error "Installation failed during: $module_name"
        echo_info "Showing last 20 lines of log:"
        echo_divider
        tail -20 "$output_file"
        echo_divider
        echo_info "Full log: $output_file"
        exit 1
    fi
}

# Function to run module with GitHub interaction
run_github_module() {
    local module_file="$1"
    local module_name="$2"
    local module_num="$3"
    local total_modules="$4"
    
    # Clear the progress line
    printf "\r%${TERM_WIDTH}s\r" ""
    
    # Show we're on GitHub step
    printf "${COLOR_DIM}[%d/%d]${COLOR_RESET} ${COLOR_BLUE}${SYMBOL_ARROW}${COLOR_RESET} %s\n\n" \
        "$module_num" "$total_modules" "$module_name"
    
    # Run the module interactively
    "$SCRIPT_DIR/$module_file"
    
    # Add spacing after
    echo ""
}

# Start setup
# Clear screen only if in terminal
if [ -t 1 ]; then
    clear 2>/dev/null || true
fi

echo ""
printf "${COLOR_BOLD}┌────────────────────────────────────┐${COLOR_RESET}\n"
printf "${COLOR_BOLD}│  Michael's Codespaces Installer    │${COLOR_RESET}\n"
printf "${COLOR_BOLD}└────────────────────────────────────┘${COLOR_RESET}\n"
echo ""

# Show what will be installed
echo_step "Installation Plan"
echo ""
for i in "${!SETUP_MODULES[@]}"; do
    IFS=':' read -r module_file module_desc <<< "${SETUP_MODULES[$i]}"
    printf "  ${COLOR_DIM}%d.${COLOR_RESET} %s\n" $((i+1)) "$module_desc"
done
echo ""

# Countdown
echo -n "Starting installation in "
for i in 3 2 1; do
    printf "${COLOR_YELLOW}%d${COLOR_RESET}" "$i"
    sleep 1
    if [ $i -gt 1 ]; then
        printf "\b"
    fi
done
printf "\b \n\n"

# Run each module
module_num=0
total_modules=${#SETUP_MODULES[@]}

for module_info in "${SETUP_MODULES[@]}"; do
    IFS=':' read -r module_file module_desc <<< "$module_info"
    ((module_num++))
    
    # GitHub auth needs to be interactive
    if [[ "$module_file" == "06-github-auth.sh" ]]; then
        run_github_module "$module_file" "$module_desc" "$module_num" "$total_modules"
    else
        run_module_clean "$module_file" "$module_desc" "$module_num" "$total_modules"
    fi
done

# Success message
echo ""
echo_success "${SYMBOL_CHECK} Installation completed successfully!"
echo ""

# Show summary with icons
echo_step "Summary"
echo ""
printf "  ${COLOR_GREEN}${SYMBOL_CHECK}${COLOR_RESET} System packages installed\n"
printf "  ${COLOR_GREEN}${SYMBOL_CHECK}${COLOR_RESET} Docker configured and running\n"
printf "  ${COLOR_GREEN}${SYMBOL_CHECK}${COLOR_RESET} Zsh with Oh My Zsh installed\n"
printf "  ${COLOR_GREEN}${SYMBOL_CHECK}${COLOR_RESET} Codespace infrastructure created\n"
printf "  ${COLOR_GREEN}${SYMBOL_CHECK}${COLOR_RESET} Monitoring tools installed\n"
printf "  ${COLOR_GREEN}${SYMBOL_CHECK}${COLOR_RESET} GitHub authentication configured\n"
echo ""