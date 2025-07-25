# ASCII Select Utility

A reusable, lightweight selection utility for Michael's Codespaces that provides checkbox, radio button, and list selection interfaces without taking over the terminal screen.

## Overview

The `ascii-select` utility provides a clean, inline selection interface that stays within the normal terminal flow. Unlike full-screen TUI tools like whiptail, it displays selections as a simple list that doesn't clear the screen or require special terminal capabilities.

## Features

- **Multiple Selection Modes**
  - Checkbox (multiple selection)
  - Radio button (single selection)
  - List (numbered menu)

- **Visual Styles**
  - Simple (default) - Clean ASCII checkboxes
  - Fancy - Unicode characters and borders
  - Compact - Minimal space usage

- **Flexible Configuration**
  - Customizable titles and prompts
  - Item descriptions support
  - Preselection of default items
  - Min/max selection constraints
  - Custom output delimiters

## Usage

### Basic Usage

```bash
# Source the utility
source /path/to/ascii-select.sh

# Simple checkbox selection
selected=$(ascii_select "Select items:" item1 item2 item3)
echo "You selected: $selected"
```

### Command Line Options

```bash
ascii_select [options] "title" item1 item2 item3...
```

#### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--mode MODE` | Selection mode: checkbox, radio, list | checkbox |
| `--style STYLE` | Visual style: simple, fancy, compact | simple |
| `--border` | Show border (fancy style only) | false |
| `--min N` | Minimum selections required | 0 |
| `--max N` | Maximum selections allowed | 999 |
| `--preselect "1,3,5"` | Preselect items by number | none |
| `--delimiter "SEP"` | Output delimiter | space |
| `--with-descriptions` | Enable item descriptions | false |
| `--no-numbers` | Hide item numbers | false |
| `--prompt "TEXT"` | Custom input prompt | "Selection" |
| `--no-color` | Disable colors | false |

### Examples

#### Checkbox with Descriptions

```bash
result=$(ascii_select --with-descriptions \
    "Select components:" \
    "github-cli|GitHub command line interface" \
    "claude|AI-powered coding assistant" \
    "docker|Container runtime")
```

Output:
```
Select components:

[x] 1. github-cli - GitHub command line interface
[ ] 2. claude - AI-powered coding assistant
[x] 3. docker - Container runtime

Toggle: 1-3, All: a, None: n, Confirm: Enter
Selection: _
```

#### Radio Button Selection

```bash
choice=$(ascii_select --mode radio \
    "Choose environment:" \
    "Production" "Staging" "Development" "Local")
```

Output:
```
Choose environment:

( ) 1. Production
(●) 2. Staging
( ) 3. Development
( ) 4. Local

Select: 1-4, Confirm: Enter
Selection: _
```

#### Fancy Style with Border

```bash
result=$(ascii_select --style fancy --border \
    "Select features:" \
    "Feature A" "Feature B" "Feature C")
```

Output:
```
╭─ Select features ─────────────────────────────────╮
│ ☑ 1. Feature A                                    │
│ ☐ 2. Feature B                                    │
│ ☑ 3. Feature C                                    │
╰───────────────────────────────────────────────────╯
```

#### List Mode (Menu)

```bash
option=$(ascii_select --mode list \
    "Main Menu:" \
    "New Project" "Open Project" "Settings" "Exit")
```

#### With Constraints

```bash
# Require at least 2, maximum 4 selections
result=$(ascii_select --min 2 --max 4 \
    "Select 2-4 options:" \
    "Option 1" "Option 2" "Option 3" "Option 4" "Option 5")
```

## Integration in Scripts

### Component Selection Example

```bash
#!/bin/bash
source /path/to/ascii-select.sh

# Get available components
components=("github-cli" "claude" "docker" "nodejs")

# Let user select
selected=$(ascii_select \
    --preselect "1,2" \
    --style simple \
    "Select components to install:" \
    "${components[@]}")

# Process selections
for component in $selected; do
    echo "Installing $component..."
    # Installation logic here
done
```

### Menu System Example

```bash
#!/bin/bash
source /path/to/ascii-select.sh

while true; do
    choice=$(ascii_select --mode radio --style fancy \
        "Main Menu:" \
        "Create Codespace" \
        "List Codespaces" \
        "Update Settings" \
        "Exit")
    
    case "$choice" in
        "Create Codespace")
            # Create logic
            ;;
        "List Codespaces")
            # List logic
            ;;
        "Update Settings")
            # Settings logic
            ;;
        "Exit")
            break
            ;;
    esac
done
```

## User Interaction

### Checkbox Mode
- **Number keys (1-9)**: Toggle specific item
- **Space**: Toggle current item (if arrow navigation added)
- **a/A**: Select all items
- **n/N**: Deselect all items
- **Enter**: Confirm selection
- **q/Q**: Cancel selection

### Radio Mode
- **Number keys (1-9)**: Select specific item
- **Enter**: Confirm selection
- **q/Q**: Cancel selection

### List Mode
- **Number keys**: Select and confirm immediately
- **q/Q**: Cancel selection

## Return Values

- **Success (0)**: User confirmed selection
  - Returns selected items separated by delimiter
- **Failure (1)**: User cancelled or error occurred
  - No output on stdout

## Environment Variables

- `ASCI_COLOR`: Set to false to disable colors globally

## Best Practices

1. **Always check return code**:
   ```bash
   if result=$(ascii_select "Select:" item1 item2); then
       echo "Selected: $result"
   else
       echo "Selection cancelled"
   fi
   ```

2. **Use descriptions for clarity**:
   ```bash
   ascii_select --with-descriptions "Select:" \
       "opt1|Short description" \
       "opt2|Another description"
   ```

3. **Set appropriate constraints**:
   ```bash
   # For required selections
   ascii_select --min 1 "Select at least one:" item1 item2
   ```

4. **Choose the right mode**:
   - Use checkbox for multiple selections
   - Use radio for mutually exclusive choices
   - Use list for simple menus

## Troubleshooting

### No items displayed
- Ensure items are passed as separate arguments
- Check that the script is properly sourced

### Colors not working
- Check terminal supports colors (`tput colors`)
- Try without `--no-color` flag
- Ensure output is to a terminal (not piped)

### Selection not working
- Verify terminal is interactive (`[ -t 0 ]`)
- Check stdin is not redirected
- Try simpler style or mode

## Future Enhancements

- Arrow key navigation
- Search/filter for long lists
- Multi-column layout
- Nested/hierarchical selections
- Custom key bindings
- Progress indicators