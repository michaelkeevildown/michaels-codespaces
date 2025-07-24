#!/bin/bash

# Advanced color and formatting utilities - Homebrew-inspired

# Detect color support
if [[ -t 1 ]] && [[ -n "${TERM:-}" ]] && [[ "${TERM:-}" != "dumb" ]]; then
    # Basic colors
    export COLOR_RESET='\033[0m'
    export COLOR_BOLD='\033[1m'
    export COLOR_DIM='\033[2m'
    export COLOR_UNDERLINE='\033[4m'
    export COLOR_BLINK='\033[5m'
    export COLOR_REVERSE='\033[7m'
    export COLOR_HIDDEN='\033[8m'
    
    # Foreground colors
    export COLOR_BLACK='\033[30m'
    export COLOR_RED='\033[31m'
    export COLOR_GREEN='\033[32m'
    export COLOR_YELLOW='\033[33m'
    export COLOR_BLUE='\033[34m'
    export COLOR_MAGENTA='\033[35m'
    export COLOR_CYAN='\033[36m'
    export COLOR_WHITE='\033[37m'
    export COLOR_GRAY='\033[90m'
    export COLOR_BRIGHT_RED='\033[91m'
    export COLOR_BRIGHT_GREEN='\033[92m'
    export COLOR_BRIGHT_YELLOW='\033[93m'
    export COLOR_BRIGHT_BLUE='\033[94m'
    export COLOR_BRIGHT_MAGENTA='\033[95m'
    export COLOR_BRIGHT_CYAN='\033[96m'
    export COLOR_BRIGHT_WHITE='\033[97m'
    
    # Background colors
    export BG_BLACK='\033[40m'
    export BG_RED='\033[41m'
    export BG_GREEN='\033[42m'
    export BG_YELLOW='\033[43m'
    export BG_BLUE='\033[44m'
    export BG_MAGENTA='\033[45m'
    export BG_CYAN='\033[46m'
    export BG_WHITE='\033[47m'
    
    # Unicode symbols
    export SYMBOL_CHECK='✓'
    export SYMBOL_CROSS='✗'
    export SYMBOL_ARROW='▶'
    export SYMBOL_BULLET='•'
    export SYMBOL_ELLIPSIS='…'
    export SYMBOL_INFO='ℹ'
    export SYMBOL_WARNING='⚠'
    export SYMBOL_STAR='★'
    export SYMBOL_ROCKET='🚀'
    export SYMBOL_PACKAGE='📦'
    export SYMBOL_LOCK='🔒'
    export SYMBOL_KEY='🔑'
    export SYMBOL_GEAR='⚙'
    export SYMBOL_LIGHTNING='⚡'
    export SYMBOL_FIRE='🔥'
else
    # No color support
    export COLOR_RESET=''
    export COLOR_BOLD=''
    export COLOR_DIM=''
    export COLOR_UNDERLINE=''
    export COLOR_BLINK=''
    export COLOR_REVERSE=''
    export COLOR_HIDDEN=''
    export COLOR_BLACK=''
    export COLOR_RED=''
    export COLOR_GREEN=''
    export COLOR_YELLOW=''
    export COLOR_BLUE=''
    export COLOR_MAGENTA=''
    export COLOR_CYAN=''
    export COLOR_WHITE=''
    export COLOR_GRAY=''
    export COLOR_BRIGHT_RED=''
    export COLOR_BRIGHT_GREEN=''
    export COLOR_BRIGHT_YELLOW=''
    export COLOR_BRIGHT_BLUE=''
    export COLOR_BRIGHT_MAGENTA=''
    export COLOR_BRIGHT_CYAN=''
    export COLOR_BRIGHT_WHITE=''
    export BG_BLACK=''
    export BG_RED=''
    export BG_GREEN=''
    export BG_YELLOW=''
    export BG_BLUE=''
    export BG_MAGENTA=''
    export BG_CYAN=''
    export BG_WHITE=''
    export SYMBOL_CHECK='[OK]'
    export SYMBOL_CROSS='[FAIL]'
    export SYMBOL_ARROW='=>'
    export SYMBOL_BULLET='*'
    export SYMBOL_ELLIPSIS='...'
    export SYMBOL_INFO='[i]'
    export SYMBOL_WARNING='[!]'
    export SYMBOL_STAR='*'
    export SYMBOL_ROCKET=''
    export SYMBOL_PACKAGE=''
    export SYMBOL_LOCK=''
    export SYMBOL_KEY=''
    export SYMBOL_GEAR=''
    export SYMBOL_LIGHTNING=''
    export SYMBOL_FIRE=''
fi

# Legacy color variables for backward compatibility
RED=$COLOR_RED
GREEN=$COLOR_GREEN
YELLOW=$COLOR_YELLOW
BLUE=$COLOR_BLUE
PURPLE=$COLOR_MAGENTA
CYAN=$COLOR_CYAN
NC=$COLOR_RESET

# Terminal width detection
TERM_WIDTH=$(tput cols 2>/dev/null || echo 80)

# Progress indicators
SPINNER_FRAMES=("⠋" "⠙" "⠹" "⠸" "⠼" "⠴" "⠦" "⠧" "⠇" "⠏")
SPINNER_PID=""

# Output functions with Homebrew-style formatting
echo_step() {
    printf "${COLOR_BLUE}==>${COLOR_RESET} ${COLOR_BOLD}%s${COLOR_RESET}\n" "$1"
}

echo_substep() {
    printf "  ${COLOR_GREEN}${SYMBOL_ARROW}${COLOR_RESET} %s\n" "$1"
}

echo_info() {
    printf "${COLOR_BLUE}${SYMBOL_INFO}${COLOR_RESET}  %s\n" "$1"
}

echo_success() {
    printf "${COLOR_GREEN}${SYMBOL_CHECK}${COLOR_RESET} %s\n" "$1"
}

echo_warning() {
    printf "${COLOR_YELLOW}${SYMBOL_WARNING}${COLOR_RESET}  %s\n" "$1"
}

echo_error() {
    printf "${COLOR_RED}${SYMBOL_CROSS}${COLOR_RESET} %s\n" "$1" >&2
}

echo_debug() {
    if [[ "${DEBUG:-0}" == "1" ]] || [[ "${VERBOSE:-0}" == "1" ]]; then
        printf "${COLOR_GRAY}[DEBUG]${COLOR_RESET} ${COLOR_DIM}%s${COLOR_RESET}\n" "$1"
    fi
}

# Pretty print with indentation
echo_indent() {
    local level="${1:-1}"
    local message="$2"
    local indent=""
    for ((i=0; i<level; i++)); do
        indent="${indent}  "
    done
    printf "%s%s\n" "$indent" "$message"
}

# Box drawing
echo_box() {
    local title="$1"
    local width=${2:-$((TERM_WIDTH-4))}
    local line=$(printf '─%.0s' $(seq 1 $width))
    
    printf "\n${COLOR_BOLD}┌─ %s %s┐${COLOR_RESET}\n" "$title" "${line:0:$((width-${#title}-4))}"
}

echo_box_end() {
    local width=${1:-$((TERM_WIDTH-4))}
    local line=$(printf '─%.0s' $(seq 1 $width))
    printf "${COLOR_BOLD}└%s┘${COLOR_RESET}\n\n" "$line"
}

# Progress bar
progress_bar() {
    local current=$1
    local total=$2
    local width=${3:-50}
    local percentage=$((current * 100 / total))
    local filled=$((width * current / total))
    local empty=$((width - filled))
    
    printf "\r["
    printf "${COLOR_GREEN}%${filled}s${COLOR_RESET}" | tr ' ' '█'
    printf "%${empty}s" | tr ' ' '░'
    printf "] %3d%% " $percentage
}

# Spinner functions
start_spinner() {
    local message="${1:-Loading}"
    
    (while true; do
        for frame in "${SPINNER_FRAMES[@]}"; do
            printf "\r${COLOR_CYAN}%s${COLOR_RESET} %s${SYMBOL_ELLIPSIS}" "$frame" "$message"
            sleep 0.08
        done
    done) &
    
    SPINNER_PID=$!
}

stop_spinner() {
    if [[ -n "$SPINNER_PID" ]]; then
        kill "$SPINNER_PID" 2>/dev/null
        wait "$SPINNER_PID" 2>/dev/null
        printf "\r%*s\r" "${TERM_WIDTH}" ""
        SPINNER_PID=""
    fi
}

# Status line that overwrites itself
echo_status() {
    printf "\r${COLOR_DIM}%s${COLOR_RESET}%*s" "$1" $((TERM_WIDTH - ${#1} - 1)) ""
}

clear_status() {
    printf "\r%*s\r" "${TERM_WIDTH}" ""
}

# Formatted list
echo_list_item() {
    local marker="${1:-$SYMBOL_BULLET}"
    local text="$2"
    printf "  ${COLOR_CYAN}%s${COLOR_RESET} %s\n" "$marker" "$text"
}

# Table formatting
echo_table_header() {
    printf "${COLOR_BOLD}${COLOR_UNDERLINE}%-30s %-15s %-30s${COLOR_RESET}\n" "$1" "$2" "$3"
}

echo_table_row() {
    printf "%-30s %-15s %-30s\n" "$1" "$2" "$3"
}

# Divider line
echo_divider() {
    local char="${1:-─}"
    local width="${2:-$TERM_WIDTH}"
    printf "${COLOR_DIM}%${width}s${COLOR_RESET}\n" | tr ' ' "$char"
}

# Homebrew-style caveats section
echo_caveats() {
    echo ""
    printf "${COLOR_YELLOW}${COLOR_BOLD}==> Caveats${COLOR_RESET}\n"
}

# Task runner with progress
run_task() {
    local task_name="$1"
    local task_command="$2"
    
    start_spinner "$task_name"
    
    # Run command and capture output
    local output_file=$(mktemp)
    local exit_code=0
    
    eval "$task_command" > "$output_file" 2>&1 || exit_code=$?
    
    stop_spinner
    
    if [[ $exit_code -eq 0 ]]; then
        echo_success "$task_name"
    else
        echo_error "$task_name failed"
        if [[ "${VERBOSE:-0}" == "1" ]]; then
            cat "$output_file"
        fi
    fi
    
    rm -f "$output_file"
    return $exit_code
}

# Multi-step progress
PROGRESS_CURRENT=0
PROGRESS_TOTAL=0

init_progress() {
    PROGRESS_CURRENT=0
    PROGRESS_TOTAL=$1
}

step_progress() {
    local step_name="$1"
    PROGRESS_CURRENT=$((PROGRESS_CURRENT + 1))
    
    local percentage=$((PROGRESS_CURRENT * 100 / PROGRESS_TOTAL))
    printf "${COLOR_DIM}[%2d/%2d]${COLOR_RESET} %s\n" "$PROGRESS_CURRENT" "$PROGRESS_TOTAL" "$step_name"
}

# Loading animation for long operations
loading_dots() {
    local message="$1"
    local dots=""
    
    for i in {1..3}; do
        printf "\r%s%s   " "$message" "$dots"
        dots="${dots}."
        sleep 0.5
    done
    printf "\r%*s\r" "${TERM_WIDTH}" ""
}