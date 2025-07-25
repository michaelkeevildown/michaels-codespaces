# Component Selection Guide

When creating a codespace with `mcs create`, you'll be prompted to select components to install. The system supports three selection methods, automatically choosing the best one for your environment.

## Selection Methods

### 1. Whiptail Selection (Ubuntu Installer Style) - Default
If `whiptail` is installed, you'll get an Ubuntu installer-style interface:

```
┌─────────────── Component Selection ───────────────┐
│ Select components to install:                      │
│                                                    │
│    [X] github-cli    GitHub CLI - Command-line... │
│    [X] claude        Claude Code - Anthropic's... │
│    [X] claude-flow   Claude Flow - AI swarm...    │
│                                                    │
│       <Ok>                   <Cancel>              │
└────────────────────────────────────────────────────┘
```

**Controls:**
- Use ↑/↓ arrow keys to navigate
- Press SPACE to toggle selection
- Press TAB to move between components and buttons
- Press ENTER to confirm

### 2. Interactive Terminal Selection (Fallback)
If whiptail isn't available but you have a proper terminal:

```
┌─ Select Components ─────────────────────────────────────────┐
│ ● GitHub CLI           Command-line interface for GitHub    │
│ ● Claude Code          Anthropic's Claude AI coding assi... │
│ ● Claude Flow          AI swarm orchestration and workfl... │
├─────────────────────────────────────────────────────────────┤
│ [Space] Toggle  [a] All  [n] None  [Enter] Confirm  [q] Cancel │
└─────────────────────────────────────────────────────────────┘
```

### 3. Simple Text Selection (Final Fallback)
For non-interactive environments or when TTY isn't available:

```
Available components:

   1) GitHub CLI           - Command-line interface for GitHub with token authentication
   2) Claude Code          - Anthropic's Claude AI coding assistant (claude-code)
   3) Claude Flow          - AI swarm orchestration and workflow automation

Presets:
   a) All components (GitHub CLI, Claude Code, Claude Flow)
   n) None (skip component installation)

Select components (comma-separated numbers), preset letter, or press Enter for all: 
```

## Installing Whiptail

To get the best experience (Ubuntu-style menus), install whiptail:

### Ubuntu/Debian:
```bash
sudo apt-get install whiptail
```

### On the host system (recommended):
```bash
./scripts/utils/install-whiptail.sh
```

## Component Descriptions

### GitHub CLI
- Provides `gh` command for GitHub operations
- Automatically authenticates using your saved GitHub token
- Configures git with your GitHub username
- Essential for repository management

### Claude Code
- Anthropic's AI coding assistant (`claude-code`)
- Can work with or without API key
- Provides intelligent code completion and assistance
- Aliased as `claude`, `cc`, and `ccd` (debug mode)

### Claude Flow
- AI swarm orchestration tool
- Manages multiple AI agents for complex tasks
- Requires Node.js (automatically installed)
- Provides workflow automation capabilities

## Skip Component Selection

To skip the interactive selection:

```bash
# Install no components
mcs create https://github.com/user/repo.git --no-interactive

# Specify components directly
mcs create https://github.com/user/repo.git --components github-cli,claude

# Use all components
mcs create https://github.com/user/repo.git --preset ai-dev
```

## Troubleshooting

### "Falling back to simple selection"
This happens when:
- Whiptail is not installed
- Terminal doesn't support interactive mode (no TTY)
- SSH session without TTY allocation

**Solution:** Install whiptail or use `ssh -t` for TTY allocation

### Components not installing
Check the container logs:
```bash
docker logs <container-name>
cat ~/codespaces/<name>/logs/init.log
```

### Debug mode
Run with debug flag to see detailed selection process:
```bash
mcs create https://github.com/user/repo.git --debug
```