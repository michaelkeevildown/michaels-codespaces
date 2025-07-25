#!/bin/bash

# ASCII Component Selector Module
# Uses the ascii-select utility for component selection

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/registry.sh"

# Source the ascii-select utility
UTILS_DIR="$SCRIPT_DIR/../../utils"
if [ -f "$UTILS_DIR/ascii-select.sh" ]; then
    source "$UTILS_DIR/ascii-select.sh"
else
    echo "Error: ascii-select.sh not found" >&2
    exit 1
fi

# Check if ascii-select is available
check_ascii_select() {
    if command -v ascii_select >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# ASCII selection
ascii_component_select() {
    # Ensure components are registered
    register_components
    
    # Build component list with descriptions
    local components=()
    while IFS= read -r component; do
        local name=$(get_component_info "$component" "name")
        local desc=$(get_component_info "$component" "description")
        
        if [ -n "$name" ]; then
            components+=("${component}|${name} - ${desc}")
        fi
    done < <(list_components)
    
    # Check if we have components
    if [ ${#components[@]} -eq 0 ]; then
        echo "No components available" >&2
        return 1
    fi
    
    # Show ASCII selection
    local selected
    if selected=$(ascii_select \
        --with-descriptions \
        --preselect "1,2,3" \
        --style simple \
        --delimiter " " \
        "Select components to install:" \
        "${components[@]}"); then
        
        # Return selected component IDs
        echo "$selected"
        return 0
    else
        # User cancelled
        return 1
    fi
}

# Preset selection with ASCII
ascii_preset_select() {
    # Simple preset menu
    local choice
    if choice=$(ascii_select \
        --mode radio \
        --style simple \
        "Choose installation preset:" \
        "All components" \
        "Custom selection" \
        "No components"); then
        
        case "$choice" in
            "All components")
                echo "github-cli claude claude-flow"
                return 0
                ;;
            "Custom selection")
                # Return special value to indicate custom selection
                echo "CUSTOM"
                return 0
                ;;
            "No components")
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
ascii_component_selection() {
    # First show preset menu
    local preset_result=$(ascii_preset_select)
    local exit_code=$?
    
    if [ $exit_code -ne 0 ]; then
        # User cancelled
        return 1
    fi
    
    if [ "$preset_result" = "CUSTOM" ]; then
        # Show component selection
        ascii_component_select
    else
        # Return preset result
        echo "$preset_result"
    fi
}

# Export functions
export -f check_ascii_select
export -f ascii_component_select
export -f ascii_preset_select
export -f ascii_component_selection