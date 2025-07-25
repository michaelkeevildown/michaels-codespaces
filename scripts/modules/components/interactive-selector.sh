#!/bin/bash

# Interactive Component Selector Module
# Provides an interactive menu for selecting components

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/registry.sh"

# Source utilities
if [ -f "$HOME/codespaces/scripts/utils/colors.sh" ]; then
    source "$HOME/codespaces/scripts/utils/colors.sh"
else
    # Fallback color definitions
    COLOR_RESET='\033[0m'
    COLOR_BOLD='\033[1m'
    COLOR_GREEN='\033[0;32m'
    COLOR_BLUE='\033[0;34m'
    COLOR_CYAN='\033[0;36m'
    COLOR_GRAY='\033[0;90m'
    COLOR_REVERSE='\033[7m'
fi

# Interactive menu state
declare -a MENU_ITEMS
declare -a MENU_SELECTED
declare -i MENU_POSITION=0
declare -i MENU_SIZE=0

# Terminal handling
save_cursor() {
    printf '\033[s'
}

restore_cursor() {
    printf '\033[u'
}

clear_menu() {
    printf '\033[%dA' "$((MENU_SIZE + 4))"
    printf '\033[J'
}

# Initialize menu
init_menu() {
    MENU_ITEMS=()
    MENU_SELECTED=()
    MENU_POSITION=0
    
    # Ensure components are registered
    register_components
    
    # Get all components
    while IFS= read -r component; do
        MENU_ITEMS+=("$component")
        MENU_SELECTED+=(0)
    done < <(list_components)
    
    MENU_SIZE=${#MENU_ITEMS[@]}
}

# Draw menu header
draw_header() {
    echo -e "${COLOR_BOLD}┌─ Select Components ─────────────────────────────────────────┐${COLOR_RESET}"
}

# Draw menu footer
draw_footer() {
    echo -e "${COLOR_BOLD}├─────────────────────────────────────────────────────────────┤${COLOR_RESET}"
    echo -e "${COLOR_BOLD}│${COLOR_RESET} ${COLOR_CYAN}[Space]${COLOR_RESET} Toggle  ${COLOR_CYAN}[a]${COLOR_RESET} All  ${COLOR_CYAN}[n]${COLOR_RESET} None  ${COLOR_CYAN}[Enter]${COLOR_RESET} Confirm  ${COLOR_CYAN}[q]${COLOR_RESET} Cancel ${COLOR_BOLD}│${COLOR_RESET}"
    echo -e "${COLOR_BOLD}└─────────────────────────────────────────────────────────────┘${COLOR_RESET}"
}

# Draw menu item
draw_item() {
    local index=$1
    local component="${MENU_ITEMS[$index]}"
    local selected="${MENU_SELECTED[$index]}"
    local name=$(get_component_info "$component" "name")
    local desc=$(get_component_info "$component" "description")
    local deps=$(get_component_dependencies "$component")
    
    # Prepare selection indicator
    local indicator="○"
    if [ "$selected" -eq 1 ]; then
        indicator="${COLOR_GREEN}●${COLOR_RESET}"
    fi
    
    # Format description to fit
    local max_desc_len=45
    if [ ${#desc} -gt $max_desc_len ]; then
        desc="${desc:0:$((max_desc_len-3))}..."
    fi
    
    # Format line
    local line=$(printf "${COLOR_BOLD}│${COLOR_RESET} %s %-20s ${COLOR_GRAY}%-45s${COLOR_RESET} ${COLOR_BOLD}│${COLOR_RESET}" \
        "$indicator" "$name" "$desc")
    
    # Highlight if current position
    if [ "$index" -eq "$MENU_POSITION" ]; then
        echo -e "${COLOR_REVERSE}${line}${COLOR_RESET}"
        
        # Show dependencies if any
        if [ -n "$deps" ]; then
            restore_cursor
            printf '\033[%dB' "$((MENU_SIZE + 2))"
            printf '\033[K'
            echo -e "${COLOR_CYAN}Dependencies: ${deps}${COLOR_RESET}"
            restore_cursor
        fi
    else
        echo -e "$line"
    fi
}

# Draw the menu
draw_menu() {
    save_cursor
    draw_header
    
    for i in $(seq 0 $((MENU_SIZE - 1))); do
        draw_item "$i"
    done
    
    draw_footer
    
    # Show selected count
    local selected_count=0
    for sel in "${MENU_SELECTED[@]}"; do
        ((selected_count += sel))
    done
    echo -e "${COLOR_GRAY}Selected: $selected_count component(s)${COLOR_RESET}"
}

# Toggle selection
toggle_selection() {
    local index=$1
    if [ "${MENU_SELECTED[$index]}" -eq 0 ]; then
        MENU_SELECTED[$index]=1
        
        # Auto-select dependencies
        local component="${MENU_ITEMS[$index]}"
        local deps=$(get_component_dependencies "$component")
        if [ -n "$deps" ]; then
            for dep in $deps; do
                # Find dep index
                for i in $(seq 0 $((MENU_SIZE - 1))); do
                    if [ "${MENU_ITEMS[$i]}" == "$dep" ]; then
                        MENU_SELECTED[$i]=1
                        break
                    fi
                done
            done
        fi
    else
        MENU_SELECTED[$index]=0
    fi
}

# Select all
select_all() {
    for i in $(seq 0 $((MENU_SIZE - 1))); do
        MENU_SELECTED[$i]=1
    done
}

# Select none
select_none() {
    for i in $(seq 0 $((MENU_SIZE - 1))); do
        MENU_SELECTED[$i]=0
    done
}

# Handle key press
handle_key() {
    local key="$1"
    
    case "$key" in
        $'\x1b[A'|k) # Up arrow or k
            if [ "$MENU_POSITION" -gt 0 ]; then
                ((MENU_POSITION--))
            else
                MENU_POSITION=$((MENU_SIZE - 1))
            fi
            ;;
        $'\x1b[B'|j) # Down arrow or j
            if [ "$MENU_POSITION" -lt $((MENU_SIZE - 1)) ]; then
                ((MENU_POSITION++))
            else
                MENU_POSITION=0
            fi
            ;;
        ' ') # Space - toggle selection
            toggle_selection "$MENU_POSITION"
            ;;
        'a'|'A') # Select all
            select_all
            ;;
        'n'|'N') # Select none
            select_none
            ;;
        $'\n') # Enter - confirm
            return 0
            ;;
        'q'|'Q'|$'\x1b') # q or ESC - cancel
            return 1
            ;;
    esac
    
    return 2 # Continue
}

# Get selected components
get_selected_components() {
    local selected=()
    
    for i in $(seq 0 $((MENU_SIZE - 1))); do
        if [ "${MENU_SELECTED[$i]}" -eq 1 ]; then
            selected+=("${MENU_ITEMS[$i]}")
        fi
    done
    
    printf '%s\n' "${selected[@]}"
}

# Main interactive selection
interactive_select() {
    # Check if terminal supports interactive mode
    if [ ! -t 0 ] || [ ! -t 1 ]; then
        echo "Terminal does not support interactive mode" >&2
        echo "Use --components or --preset flags instead" >&2
        return 1
    fi
    
    # Initialize
    init_menu
    
    # Check if we have components to display
    if [ ${#MENU_ITEMS[@]} -eq 0 ]; then
        echo "No components available" >&2
        return 1
    fi
    
    # Clear screen space
    printf '\n%.0s' {1..20}
    printf '\033[20A'
    
    # Disable cursor
    printf '\033[?25l'
    
    # Set up trap to restore terminal on exit
    trap 'printf "\033[?25h"; stty echo 2>/dev/null || true' EXIT INT TERM
    
    # Disable echo
    stty -echo 2>/dev/null || true
    
    # Main loop
    while true; do
        draw_menu
        
        # Read single character
        IFS= read -r -n 1 key
        
        # Handle escape sequences
        if [[ $key == $'\x1b' ]]; then
            read -r -n 2 -t 0.1 key2 || true
            key="${key}${key2}"
        fi
        
        # Clear menu for redraw
        clear_menu
        
        # Handle key
        handle_key "$key"
        local result=$?
        
        if [ $result -eq 0 ]; then
            # Confirmed
            break
        elif [ $result -eq 1 ]; then
            # Cancelled
            printf '\033[?25h'
            stty echo
            return 1
        fi
    done
    
    # Restore cursor and echo
    printf '\033[?25h'
    stty echo
    
    # Clear menu one last time
    clear_menu
    
    # Return selected components
    get_selected_components
}

# Select components from preset
select_preset() {
    echo -e "${COLOR_BOLD}Available Presets:${COLOR_RESET}"
    echo ""
    
    local i=1
    declare -A preset_map
    
    while IFS='|' read -r preset_name preset_desc; do
        echo -e "  ${COLOR_CYAN}$i)${COLOR_RESET} ${COLOR_BOLD}$preset_name${COLOR_RESET}"
        echo -e "     ${COLOR_GRAY}$preset_desc${COLOR_RESET}"
        preset_map[$i]="$preset_name"
        ((i++))
    done < <(list_presets)
    
    echo ""
    echo -n "Select preset (1-$((i-1)), or press Enter to customize): "
    read -r selection
    
    if [ -z "$selection" ]; then
        return 1
    fi
    
    if [ -n "${preset_map[$selection]}" ]; then
        local preset="${preset_map[$selection]}"
        echo -e "${COLOR_GREEN}Loading preset: $preset${COLOR_RESET}"
        load_preset "$preset"
        return 0
    else
        echo -e "${COLOR_RED}Invalid selection${COLOR_RESET}"
        return 1
    fi
}

# Export functions
export -f interactive_select
export -f select_preset
export -f get_selected_components