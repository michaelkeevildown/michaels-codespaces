#!/bin/bash

# ASCII Component Selector Module
# Uses the ascii-select utility for component selection

set -e

# Define echo_debug if not already defined
if ! type -t echo_debug >/dev/null 2>&1; then
    echo_debug() {
        [ "${DEBUG:-0}" -eq 1 ] && echo "🔍 $1" >&2
    }
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo_debug "ASCII-SELECTOR: SCRIPT_DIR=$SCRIPT_DIR" >&2

source "$SCRIPT_DIR/registry.sh"

# Source the ascii-select utility
UTILS_DIR="$SCRIPT_DIR/../../utils"
echo_debug "ASCII-SELECTOR: Looking for ascii-select.sh at: $UTILS_DIR/ascii-select.sh" >&2

if [ -f "$UTILS_DIR/ascii-select.sh" ]; then
    echo_debug "ASCII-SELECTOR: Found ascii-select.sh, sourcing..." >&2
    source "$UTILS_DIR/ascii-select.sh"
    echo_debug "ASCII-SELECTOR: After sourcing, ascii_select is: $(type -t ascii_select || echo 'not found')" >&2
else
    echo "Error: ascii-select.sh not found at $UTILS_DIR/ascii-select.sh" >&2
    # Return false for check function instead of exiting
    check_ascii_select() { return 1; }
    ascii_component_select() { return 1; }
    ascii_component_selection() { return 1; }
fi

# Check if ascii-select is available
check_ascii_select() {
    echo_debug "ASCII-SELECTOR: check_ascii_select called" >&2
    # Check if the ascii_select function is available
    local func_type=$(type -t ascii_select 2>/dev/null)
    echo_debug "ASCII-SELECTOR: type -t ascii_select returned: '$func_type'" >&2
    
    if [ "$func_type" = "function" ]; then
        echo_debug "ASCII-SELECTOR: ascii_select is available as a function" >&2
        return 0
    else
        echo_debug "ASCII-SELECTOR: ascii_select is NOT available" >&2
        return 1
    fi
}

# ASCII selection
ascii_component_select() {
    echo_debug "ASCII-SELECTOR: ascii_component_select called" >&2
    
    # Ensure components are registered
    echo_debug "ASCII-SELECTOR: Calling register_components" >&2
    register_components
    
    # Build component list with descriptions
    local components=()
    echo_debug "ASCII-SELECTOR: Building component list..." >&2
    
    while IFS= read -r component; do
        echo_debug "ASCII-SELECTOR: Processing component: '$component'" >&2
        local name=$(get_component_info "$component" "name")
        local desc=$(get_component_info "$component" "description")
        echo_debug "ASCII-SELECTOR: Component name='$name', desc='$desc'" >&2
        
        if [ -n "$name" ]; then
            components+=("${component}|${name} - ${desc}")
            echo_debug "ASCII-SELECTOR: Added to array: '${component}|${name} - ${desc}'" >&2
        fi
    done < <(list_components)
    
    # Debug: log what we found
    echo_debug "ASCII-SELECTOR: Found ${#components[@]} components total" >&2
    
    # Check if we have components
    if [ ${#components[@]} -eq 0 ]; then
        echo "No components available" >&2
        return 1
    fi
    
    # Debug: Show what we're about to display
    [ "${DEBUG:-0}" -eq 1 ] && {
        echo "DEBUG: About to show ASCII selection with components:" >&2
        for comp in "${components[@]}"; do
            echo "DEBUG:   $comp" >&2
        done
    }
    
    # Show ASCII selection
    echo_debug "ASCII-SELECTOR: About to call ascii_select with ${#components[@]} components" >&2
    echo_debug "ASCII-SELECTOR: ascii_select command: ascii_select --with-descriptions --preselect '1,2,3' --style simple --delimiter ' ' 'Select components to install:' '${components[@]}'" >&2
    
    local selected
    if selected=$(ascii_select \
        --with-descriptions \
        --preselect "1,2,3" \
        --style simple \
        --delimiter " " \
        "Select components to install:" \
        "${components[@]}" 2>&1); then
        
        [ "${DEBUG:-0}" -eq 1 ] && echo "DEBUG: Raw selection output: '$selected'" >&2
        
        # Extract just the component IDs from the selection
        # The selected output will be in format: "component_id|description component_id|description"
        local component_ids=""
        for item in $selected; do
            local id=$(echo "$item" | cut -d'|' -f1)
            if [ -n "$id" ]; then
                component_ids="$component_ids $id"
            fi
        done
        
        # Return selected component IDs (trimmed)
        local result="${component_ids## }"
        [ "${DEBUG:-0}" -eq 1 ] && echo "DEBUG: Returning component IDs: '$result'" >&2
        echo "$result"
        return 0
    else
        local exit_code=$?
        [ "${DEBUG:-0}" -eq 1 ] && echo "DEBUG: ASCII select failed with exit code: $exit_code" >&2
        # User cancelled or error occurred
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
    # Debug output
    echo_debug "ASCII-SELECTOR: ascii_component_selection called" >&2
    echo_debug "ASCII-SELECTOR: DEBUG=$DEBUG" >&2
    
    # Directly show component selection with all pre-selected
    # This simplifies the flow and avoids the hanging preset menu
    echo_debug "ASCII-SELECTOR: Calling ascii_component_select" >&2
    local result
    if result=$(ascii_component_select); then
        echo_debug "ASCII-SELECTOR: ascii_component_select succeeded, result='$result'" >&2
        echo "$result"
        return 0
    else
        local exit_code=$?
        echo_debug "ASCII-SELECTOR: ascii_component_select failed with exit code $exit_code" >&2
        return $exit_code
    fi
}

# Export functions
export -f check_ascii_select
export -f ascii_component_select
export -f ascii_preset_select
export -f ascii_component_selection