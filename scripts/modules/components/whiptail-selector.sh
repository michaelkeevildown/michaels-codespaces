#!/bin/bash

# Whiptail Component Selector Module
# Provides Ubuntu installer-style menu for component selection

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/registry.sh"

# Check if whiptail is available
check_whiptail() {
    if command -v whiptail >/dev/null 2>&1; then
        return 0
    elif command -v dialog >/dev/null 2>&1; then
        # dialog can be used as a fallback
        return 0
    else
        return 1
    fi
}

# Get dialog command (whiptail or dialog)
get_dialog_cmd() {
    if command -v whiptail >/dev/null 2>&1; then
        echo "whiptail"
    elif command -v dialog >/dev/null 2>&1; then
        echo "dialog"
    else
        echo ""
    fi
}

# Whiptail selection
whiptail_select() {
    local dialog_cmd=$(get_dialog_cmd)
    
    if [ -z "$dialog_cmd" ]; then
        echo "Whiptail or dialog not available" >&2
        return 1
    fi
    
    # Ensure components are registered
    register_components
    
    # Build checklist options
    local options=()
    local i=1
    
    while IFS= read -r component; do
        local name=$(get_component_info "$component" "name")
        local desc=$(get_component_info "$component" "description")
        
        if [ -n "$name" ]; then
            # Add component to options (tag, item, status)
            # All components are selected by default (ON)
            options+=("$component" "$name - $desc" "ON")
        fi
        ((i++))
    done < <(list_components)
    
    # Calculate dialog dimensions
    local height=$((${#options[@]}/3 + 10))
    [ $height -gt 20 ] && height=20
    local width=78
    local list_height=$((height - 8))
    
    # Show checklist dialog
    local selected
    if selected=$($dialog_cmd \
        --title "Component Selection" \
        --checklist "Select components to install (use SPACE to toggle, ENTER to confirm):" \
        $height $width $list_height \
        "${options[@]}" \
        3>&1 1>&2 2>&3); then
        
        # Return selected components (removing quotes)
        echo "$selected" | tr -d '"'
    else
        # User cancelled
        return 1
    fi
}

# Preset selection with whiptail
whiptail_preset_select() {
    local dialog_cmd=$(get_dialog_cmd)
    
    if [ -z "$dialog_cmd" ]; then
        return 1
    fi
    
    # For our simplified setup, we just offer all or none
    local choice
    if choice=$($dialog_cmd \
        --title "Quick Selection" \
        --menu "Choose a preset or select Custom to choose individual components:" \
        15 60 3 \
        "all" "All components (GitHub CLI, Claude Code, Claude Flow)" \
        "custom" "Custom selection" \
        "none" "No components" \
        3>&1 1>&2 2>&3); then
        
        case "$choice" in
            all)
                echo "github-cli claude claude-flow"
                return 0
                ;;
            custom)
                # Return special value to indicate custom selection
                echo "CUSTOM"
                return 0
                ;;
            none)
                echo ""
                return 0
                ;;
        esac
    else
        # User cancelled
        return 1
    fi
}

# Main selection function
whiptail_component_selection() {
    # First show preset menu
    local preset_result=$(whiptail_preset_select)
    local exit_code=$?
    
    if [ $exit_code -ne 0 ]; then
        # User cancelled
        return 1
    fi
    
    if [ "$preset_result" = "CUSTOM" ]; then
        # Show component selection
        whiptail_select
    else
        # Return preset result
        echo "$preset_result"
    fi
}

# Export functions
export -f check_whiptail
export -f whiptail_select
export -f whiptail_preset_select
export -f whiptail_component_selection