#!/bin/bash

# ASCII Select Demo - Demonstrates various selection modes and styles

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/ascii-select.sh"

# Colors for demo
if [[ -t 1 ]]; then
    COLOR_RESET='\033[0m'
    COLOR_BOLD='\033[1m'
    COLOR_GREEN='\033[0;32m'
    COLOR_BLUE='\033[0;34m'
    COLOR_YELLOW='\033[0;33m'
else
    COLOR_RESET=''
    COLOR_BOLD=''
    COLOR_GREEN=''
    COLOR_BLUE=''
    COLOR_YELLOW=''
fi

demo_header() {
    echo ""
    echo -e "${COLOR_BOLD}${COLOR_BLUE}═══════════════════════════════════════════════════════${COLOR_RESET}"
    echo -e "${COLOR_BOLD}$1${COLOR_RESET}"
    echo -e "${COLOR_BOLD}${COLOR_BLUE}═══════════════════════════════════════════════════════${COLOR_RESET}"
    echo ""
}

demo_result() {
    echo ""
    echo -e "${COLOR_GREEN}Result:${COLOR_RESET} $1"
    echo ""
}

# Demo 1: Basic checkbox selection
demo_header "Demo 1: Basic Checkbox Selection"
echo "This demonstrates the default checkbox mode with simple style."
echo ""

result=$(ascii_select "Select your favorite programming languages:" \
    "Python" "JavaScript" "Go" "Rust" "Ruby" "Java" "C++" "Swift")

demo_result "$result"

# Demo 2: Checkbox with descriptions
demo_header "Demo 2: Checkbox with Descriptions"
echo "Components can have descriptions for more context."
echo ""

result=$(ascii_select --with-descriptions \
    "Select components to install:" \
    "github-cli|GitHub command line interface" \
    "claude|AI-powered coding assistant" \
    "claude-flow|Advanced AI workflow automation" \
    "docker|Container runtime and tools" \
    "nodejs|JavaScript runtime environment")

demo_result "$result"

# Demo 3: Radio button selection
demo_header "Demo 3: Radio Button Selection (Single Choice)"
echo "Radio mode allows only one selection."
echo ""

result=$(ascii_select --mode radio \
    "Choose your preferred editor:" \
    "VS Code" "Vim" "Emacs" "Sublime Text" "Atom" "IntelliJ IDEA")

demo_result "$result"

# Demo 4: Fancy style with border
demo_header "Demo 4: Fancy Style with Border"
echo "A more decorative presentation style."
echo ""

result=$(ascii_select --style fancy --border --with-descriptions \
    "Select deployment targets:" \
    "production|Live production environment" \
    "staging|Pre-production testing" \
    "development|Development environment" \
    "local|Local machine only")

demo_result "$result"

# Demo 5: Compact style
demo_header "Demo 5: Compact Style"
echo "Minimal space usage for simple selections."
echo ""

result=$(ascii_select --style compact \
    "Quick options:" \
    "Yes" "No" "Maybe")

demo_result "$result"

# Demo 6: List mode (numbered menu)
demo_header "Demo 6: List Mode (Simple Menu)"
echo "Traditional numbered menu selection."
echo ""

result=$(ascii_select --mode list \
    "Main Menu:" \
    "Create new project" "Open existing project" "Settings" "Help" "Exit")

demo_result "$result"

# Demo 7: Preselection
demo_header "Demo 7: Checkbox with Preselected Items"
echo "Items can be preselected by default."
echo ""

result=$(ascii_select --preselect "1,3,5" \
    "Select features to enable:" \
    "Auto-save" "Syntax highlighting" "Auto-complete" "Linting" "Format on save" "Git integration")

demo_result "$result"

# Demo 8: Min/Max constraints
demo_header "Demo 8: Selection Constraints"
echo "Enforce minimum and maximum selection limits."
echo ""

result=$(ascii_select --min 2 --max 4 \
    "Choose 2-4 pizza toppings:" \
    "Pepperoni" "Mushrooms" "Onions" "Sausage" "Bell Peppers" "Olives" "Pineapple" "Extra Cheese")

demo_result "$result"

# Demo 9: Different delimiters
demo_header "Demo 9: Custom Output Delimiter"
echo "Output can use different delimiters (comma-separated in this case)."
echo ""

result=$(ascii_select --delimiter "," \
    "Select tags:" \
    "urgent" "bug" "feature" "documentation" "help-wanted" "good-first-issue")

demo_result "$result"

# Demo 10: No color mode
demo_header "Demo 10: No Color Mode"
echo "For terminals that don't support colors."
echo ""

result=$(ascii_select --no-color \
    "Select options (no color):" \
    "Option A" "Option B" "Option C")

demo_result "$result"

# Final message
echo ""
echo -e "${COLOR_BOLD}${COLOR_GREEN}Demo Complete!${COLOR_RESET}"
echo ""
echo "The ascii_select function provides a flexible, reusable selection interface"
echo "that can be integrated anywhere in your application."
echo ""
echo "Usage in scripts:"
echo '  result=$(ascii_select [options] "title" item1 item2 ...)'
echo ""
echo "See ascii-select.sh for full documentation."
echo ""