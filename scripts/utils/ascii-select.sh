#!/bin/bash

# ASCII Select - A reusable selection utility for Michael's Codespaces
# Provides checkbox, radio button, and list selection interfaces
# Usage: ascii_select [options] "title" item1 item2 item3...

set -e

# Default configuration
ASCI_MODE="checkbox"              # checkbox, radio, list
ASCI_STYLE="simple"              # simple, fancy, compact
ASCI_BORDER=false
ASCI_MIN_SELECT=0
ASCI_MAX_SELECT=999
ASCI_PRESELECT=""
ASCI_DELIMITER=" "               # Output delimiter
ASCI_WITH_DESC=false
ASCI_ITEM_DELIMITER="|"          # For items with descriptions
ASCI_SHOW_NUMBERS=true
ASCI_TITLE=""
ASCI_PROMPT="Selection"
ASCI_COLOR=true

# Colors (if enabled)
if [[ -t 1 ]] && [ "$ASCI_COLOR" = true ]; then
    COLOR_RESET='\033[0m'
    COLOR_BOLD='\033[1m'
    COLOR_DIM='\033[2m'
    COLOR_GREEN='\033[0;32m'
    COLOR_BLUE='\033[0;34m'
    COLOR_CYAN='\033[0;36m'
    COLOR_YELLOW='\033[0;33m'
    COLOR_GRAY='\033[0;90m'
else
    COLOR_RESET=''
    COLOR_BOLD=''
    COLOR_DIM=''
    COLOR_GREEN=''
    COLOR_BLUE=''
    COLOR_CYAN=''
    COLOR_YELLOW=''
    COLOR_GRAY=''
fi

# Parse command line options
parse_options() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --mode)
                ASCI_MODE="$2"
                shift 2
                ;;
            --style)
                ASCI_STYLE="$2"
                shift 2
                ;;
            --border)
                ASCI_BORDER=true
                shift
                ;;
            --no-border)
                ASCI_BORDER=false
                shift
                ;;
            --min)
                ASCI_MIN_SELECT="$2"
                shift 2
                ;;
            --max)
                ASCI_MAX_SELECT="$2"
                shift 2
                ;;
            --preselect)
                ASCI_PRESELECT="$2"
                shift 2
                ;;
            --delimiter)
                ASCI_DELIMITER="$2"
                shift 2
                ;;
            --with-descriptions)
                ASCI_WITH_DESC=true
                shift
                ;;
            --no-numbers)
                ASCI_SHOW_NUMBERS=false
                shift
                ;;
            --prompt)
                ASCI_PROMPT="$2"
                shift 2
                ;;
            --no-color)
                ASCI_COLOR=false
                COLOR_RESET=''
                COLOR_BOLD=''
                COLOR_DIM=''
                COLOR_GREEN=''
                COLOR_BLUE=''
                COLOR_CYAN=''
                COLOR_YELLOW=''
                COLOR_GRAY=''
                shift
                ;;
            --)
                shift
                break
                ;;
            -*)
                echo "Unknown option: $1" >&2
                return 1
                ;;
            *)
                break
                ;;
        esac
    done
    
    # First remaining argument is the title
    if [[ $# -gt 0 ]]; then
        ASCI_TITLE="$1"
        shift
    fi
    
    # Remaining arguments are items
    ASCI_ITEMS=("$@")
}

# Initialize selection state
init_selection() {
    ASCI_SELECTED=()
    local item_count=${#ASCI_ITEMS[@]}
    
    for ((i=0; i<item_count; i++)); do
        ASCI_SELECTED+=(0)
    done
    
    # Apply preselection
    if [ -n "$ASCI_PRESELECT" ]; then
        IFS=',' read -ra PRESELECTS <<< "$ASCI_PRESELECT"
        for sel in "${PRESELECTS[@]}"; do
            sel=$((sel - 1))  # Convert to 0-based
            if [ "$sel" -ge 0 ] && [ "$sel" -lt "$item_count" ]; then
                ASCI_SELECTED[$sel]=1
            fi
        done
    fi
}

# Get item name and description
get_item_parts() {
    local item="$1"
    local name=""
    local desc=""
    
    if [ "$ASCI_WITH_DESC" = true ]; then
        IFS="$ASCI_ITEM_DELIMITER" read -r name desc <<< "$item"
    else
        name="$item"
    fi
    
    echo "$name|$desc"
}

# Draw checkbox
draw_checkbox() {
    local selected="$1"
    
    case "$ASCI_STYLE" in
        fancy)
            if [ "$selected" -eq 1 ]; then
                echo -ne "${COLOR_GREEN}☑${COLOR_RESET}"
            else
                echo -ne "☐"
            fi
            ;;
        compact)
            if [ "$selected" -eq 1 ]; then
                echo -ne "[${COLOR_GREEN}x${COLOR_RESET}]"
            else
                echo -ne "[ ]"
            fi
            ;;
        *)  # simple
            if [ "$selected" -eq 1 ]; then
                echo -ne "[${COLOR_GREEN}x${COLOR_RESET}]"
            else
                echo -ne "[ ]"
            fi
            ;;
    esac
}

# Draw radio button
draw_radio() {
    local selected="$1"
    
    case "$ASCI_STYLE" in
        fancy)
            if [ "$selected" -eq 1 ]; then
                echo -ne "${COLOR_GREEN}◉${COLOR_RESET}"
            else
                echo -ne "○"
            fi
            ;;
        compact)
            if [ "$selected" -eq 1 ]; then
                echo -ne "(${COLOR_GREEN}*${COLOR_RESET})"
            else
                echo -ne "( )"
            fi
            ;;
        *)  # simple
            if [ "$selected" -eq 1 ]; then
                echo -ne "(${COLOR_GREEN}●${COLOR_RESET})"
            else
                echo -ne "( )"
            fi
            ;;
    esac
}

# Draw the selection interface
draw_interface() {
    # Clear previous output (count lines to clear)
    if [ -n "$ASCI_LAST_LINE_COUNT" ]; then
        for ((i=0; i<ASCI_LAST_LINE_COUNT; i++)); do
            echo -ne "\033[1A\033[2K"
        done
    fi
    
    local line_count=0
    
    # Title
    if [ -n "$ASCI_TITLE" ]; then
        case "$ASCI_STYLE" in
            fancy)
                if [ "$ASCI_BORDER" = true ]; then
                    echo -e "${COLOR_BOLD}╭─ $ASCI_TITLE ─────────────────────────────────╮${COLOR_RESET}"
                    ((line_count++))
                else
                    echo -e "${COLOR_BOLD}$ASCI_TITLE${COLOR_RESET}"
                    ((line_count++))
                fi
                ;;
            *)
                echo -e "${COLOR_BOLD}$ASCI_TITLE${COLOR_RESET}"
                ((line_count++))
                ;;
        esac
        
        if [ "$ASCI_STYLE" != "compact" ]; then
            echo ""
            ((line_count++))
        fi
    fi
    
    # Items
    local i=1
    local selected_count=0
    for idx in "${!ASCI_ITEMS[@]}"; do
        local item="${ASCI_ITEMS[$idx]}"
        local selected="${ASCI_SELECTED[$idx]}"
        IFS="|" read -r name desc <<< "$(get_item_parts "$item")"
        
        [ "$selected" -eq 1 ] && ((selected_count++))
        
        case "$ASCI_STYLE" in
            fancy)
                if [ "$ASCI_BORDER" = true ]; then
                    echo -n -e "${COLOR_BOLD}│${COLOR_RESET} "
                fi
                ;;
        esac
        
        # Draw selector
        case "$ASCI_MODE" in
            checkbox)
                draw_checkbox "$selected"
                ;;
            radio)
                draw_radio "$selected"
                ;;
            list)
                if [ "$selected" -eq 1 ]; then
                    echo -ne "${COLOR_GREEN}→${COLOR_RESET}"
                else
                    echo -ne " "
                fi
                ;;
        esac
        
        # Item number
        if [ "$ASCI_SHOW_NUMBERS" = true ]; then
            echo -ne " ${COLOR_CYAN}$i.${COLOR_RESET}"
        fi
        
        # Item name
        echo -ne " $name"
        
        # Description
        if [ -n "$desc" ] && [ "$ASCI_STYLE" != "compact" ]; then
            echo -ne " ${COLOR_GRAY}- $desc${COLOR_RESET}"
        fi
        
        # Border end
        case "$ASCI_STYLE" in
            fancy)
                if [ "$ASCI_BORDER" = true ]; then
                    # Pad to border
                    local line_len=$(echo -ne "$(draw_checkbox 0) $i. $name - $desc" | wc -c)
                    local pad_len=$((48 - line_len))
                    [ $pad_len -gt 0 ] && printf "%${pad_len}s"
                    echo -ne " ${COLOR_BOLD}│${COLOR_RESET}"
                fi
                ;;
        esac
        
        echo ""
        ((line_count++))
        ((i++))
    done
    
    # Bottom border
    case "$ASCI_STYLE" in
        fancy)
            if [ "$ASCI_BORDER" = true ]; then
                echo -e "${COLOR_BOLD}╰─────────────────────────────────────────────────╯${COLOR_RESET}"
                ((line_count++))
            fi
            ;;
    esac
    
    # Instructions
    if [ "$ASCI_STYLE" != "compact" ]; then
        echo ""
        ((line_count++))
        
        case "$ASCI_MODE" in
            checkbox)
                local inst="Toggle: 1-${#ASCI_ITEMS[@]}, All: a, None: n, Confirm: Enter"
                if [ "$ASCI_MIN_SELECT" -gt 0 ]; then
                    inst="$inst ${COLOR_YELLOW}(min: $ASCI_MIN_SELECT)${COLOR_RESET}"
                fi
                if [ "$ASCI_MAX_SELECT" -lt 999 ]; then
                    inst="$inst ${COLOR_YELLOW}(max: $ASCI_MAX_SELECT)${COLOR_RESET}"
                fi
                echo -e "${COLOR_DIM}$inst${COLOR_RESET}"
                ;;
            radio)
                echo -e "${COLOR_DIM}Select: 1-${#ASCI_ITEMS[@]}, Confirm: Enter${COLOR_RESET}"
                ;;
            list)
                echo -e "${COLOR_DIM}Select: 1-${#ASCI_ITEMS[@]}${COLOR_RESET}"
                ;;
        esac
        ((line_count++))
    fi
    
    # Prompt
    echo -ne "$ASCI_PROMPT: "
    ((line_count++))
    
    # Store line count for next clear
    ASCI_LAST_LINE_COUNT=$line_count
}

# Toggle selection
toggle_selection() {
    local idx=$1
    
    case "$ASCI_MODE" in
        checkbox)
            if [ "${ASCI_SELECTED[$idx]}" -eq 0 ]; then
                ASCI_SELECTED[$idx]=1
            else
                ASCI_SELECTED[$idx]=0
            fi
            ;;
        radio)
            # Clear all selections
            for i in "${!ASCI_SELECTED[@]}"; do
                ASCI_SELECTED[$i]=0
            done
            # Set this one
            ASCI_SELECTED[$idx]=1
            ;;
    esac
}

# Select all (checkbox only)
select_all() {
    if [ "$ASCI_MODE" = "checkbox" ]; then
        for i in "${!ASCI_SELECTED[@]}"; do
            ASCI_SELECTED[$i]=1
        done
    fi
}

# Select none
select_none() {
    for i in "${!ASCI_SELECTED[@]}"; do
        ASCI_SELECTED[$i]=0
    done
}

# Validate selection
validate_selection() {
    local count=0
    for sel in "${ASCI_SELECTED[@]}"; do
        [ "$sel" -eq 1 ] && ((count++))
    done
    
    if [ "$count" -lt "$ASCI_MIN_SELECT" ]; then
        echo -e "\n${COLOR_YELLOW}Please select at least $ASCI_MIN_SELECT item(s)${COLOR_RESET}" >&2
        return 1
    fi
    
    if [ "$count" -gt "$ASCI_MAX_SELECT" ]; then
        echo -e "\n${COLOR_YELLOW}Please select at most $ASCI_MAX_SELECT item(s)${COLOR_RESET}" >&2
        return 1
    fi
    
    return 0
}

# Get selected items
get_selected() {
    local selected=()
    
    for idx in "${!ASCI_ITEMS[@]}"; do
        if [ "${ASCI_SELECTED[$idx]}" -eq 1 ]; then
            IFS="|" read -r name desc <<< "$(get_item_parts "${ASCI_ITEMS[$idx]}")"
            selected+=("$name")
        fi
    done
    
    # Output with specified delimiter
    if [ ${#selected[@]} -gt 0 ]; then
        printf "%s" "${selected[0]}"
        for ((i=1; i<${#selected[@]}; i++)); do
            printf "%s%s" "$ASCI_DELIMITER" "${selected[$i]}"
        done
    fi
}

# Main selection function
ascii_select() {
    # Parse options and items
    parse_options "$@"
    
    # Initialize selection state
    init_selection
    
    # Check if we have items
    if [ ${#ASCI_ITEMS[@]} -eq 0 ]; then
        echo "No items to select from" >&2
        return 1
    fi
    
    # For list mode, just show and get single selection
    if [ "$ASCI_MODE" = "list" ]; then
        draw_interface
        read -r selection
        
        # Validate selection
        if [[ "$selection" =~ ^[0-9]+$ ]] && [ "$selection" -ge 1 ] && [ "$selection" -le "${#ASCI_ITEMS[@]}" ]; then
            local idx=$((selection - 1))
            IFS="|" read -r name desc <<< "$(get_item_parts "${ASCI_ITEMS[$idx]}")"
            echo "$name"
            return 0
        else
            return 1
        fi
    fi
    
    # Interactive selection loop
    while true; do
        draw_interface
        read -r input
        
        case "$input" in
            [0-9]*)
                # Number selection
                if [[ "$input" =~ ^[0-9]+$ ]] && [ "$input" -ge 1 ] && [ "$input" -le "${#ASCI_ITEMS[@]}" ]; then
                    toggle_selection $((input - 1))
                fi
                ;;
            a|A)
                select_all
                ;;
            n|N)
                select_none
                ;;
            "")
                # Enter pressed - confirm selection
                if validate_selection; then
                    echo ""  # New line after prompt
                    get_selected
                    return 0
                fi
                ;;
            q|Q)
                # Quit
                echo ""  # New line after prompt
                return 1
                ;;
        esac
    done
}

# Export the main function if sourced
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
    export -f ascii_select
else
    # If run directly, execute the function
    ascii_select "$@"
fi