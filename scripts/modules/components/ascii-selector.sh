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

# Simple ASCII selection without complex terminal manipulation
simple_ascii_select() {
    echo_debug "ASCII-SELECTOR: simple_ascii_select called" >&2
    
    # Ensure components are registered
    echo_debug "ASCII-SELECTOR: Calling register_components" >&2
    register_components
    
    # Build component list
    local component_ids=()
    local component_names=()
    local component_descs=()
    
    while IFS= read -r component; do
        local name=$(get_component_info "$component" "name")
        local desc=$(get_component_info "$component" "description")
        
        if [ -n "$name" ]; then
            component_ids+=("$component")
            component_names+=("$name")
            component_descs+=("$desc")
        fi
    done < <(list_components)
    
    echo_debug "ASCII-SELECTOR: Found ${#component_ids[@]} components" >&2
    
    if [ ${#component_ids[@]} -eq 0 ]; then
        echo "No components available" >&2
        return 1
    fi
    
    # Display components with ASCII checkboxes (all pre-selected)
    # Output to terminal (using /dev/tty to bypass command substitution)
    {
        echo "Select components to install:"
        echo ""
        
        for i in "${!component_ids[@]}"; do
            echo "[x] $((i+1)). ${component_names[$i]}"
            echo "    ${component_descs[$i]}"
            echo ""
        done
        
        echo "All components are selected by default."
        echo -n "Press Enter to confirm, or 'n' to skip component installation: "
    } >/dev/tty
    
    # Read user input from terminal
    local response
    read -r response </dev/tty
    
    if [[ "$response" =~ ^[nN] ]]; then
        echo_debug "ASCII-SELECTOR: User chose to skip components" >&2
        echo ""
        return 0
    else
        echo_debug "ASCII-SELECTOR: User confirmed all components" >&2
        # Return all component IDs
        echo "${component_ids[@]}"
        return 0
    fi
}

# ASCII selection (fallback to simple version)
ascii_component_select() {
    echo_debug "ASCII-SELECTOR: ascii_component_select called" >&2
    
    # Use the simpler ASCII selection that works reliably
    simple_ascii_select
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