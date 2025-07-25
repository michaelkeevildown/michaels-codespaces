# Interactive Component Selection

When creating a codespace with `mcs create`, you'll now get an interactive component selector by default.

## How it works:

### Interactive Mode (when terminal supports it):
```
┌─ Select Components ─────────────────────────────────────────┐
│ ○ GitHub CLI           Command-line interface for GitHub    │
│ ○ Claude CLI           Anthropic's Claude AI assistant CLI  │
│ ● Claude Flow          AI orchestration and workflow tool   │  ← Selected with spacebar
│ ○ Docker in Docker     Run Docker inside containers         │
│ ○ AWS CLI              Amazon Web Services command-line...  │
│ ○ Terraform            Infrastructure as Code tool          │
│ ○ Node.js Tools        npm, yarn, pnpm package managers     │
│ ○ Python Tools         pip, poetry, virtualenv tools        │
│ ○ Kubernetes Tools     kubectl, helm, k9s                   │
│ ○ Database Clients     PostgreSQL, MySQL, MongoDB clients   │
│ ○ VS Code Extensions   Popular extensions pre-installed     │
│ ○ Git Tools            git-flow, git-lfs, hub               │
├─────────────────────────────────────────────────────────────┤
│ [Space] Toggle  [a] All  [n] None  [Enter] Confirm  [q] Cancel │
└─────────────────────────────────────────────────────────────┘
Selected: 1 component(s)
```

**Controls:**
- **↑/↓** or **j/k**: Navigate up/down
- **Space**: Toggle selection on/off
- **a**: Select all components
- **n**: Deselect all components
- **Enter**: Confirm selection
- **q** or **ESC**: Cancel

### Simple Mode (fallback when no TTY):
```
Available components:

   1) AWS CLI              - Amazon Web Services command-line tools
   2) Claude CLI           - Anthropic's Claude AI assistant CLI
   3) Claude Flow          - AI orchestration and workflow tool
   4) Database Clients     - PostgreSQL, MySQL, MongoDB clients
   5) Docker in Docker     - Run Docker inside containers
   6) Git Tools            - git-flow, git-lfs, hub
   7) GitHub CLI           - Command-line interface for GitHub
   8) Kubernetes Tools     - kubectl, helm, k9s
   9) Node.js Tools        - npm, yarn, pnpm package managers
  10) Python Tools         - pip, poetry, virtualenv tools
  11) Terraform            - Infrastructure as Code tool
  12) VS Code Extensions   - Popular extensions pre-installed

Presets:
   a) AI Development (GitHub CLI, Claude, Claude Flow)
   f) Full Stack (All tools)
   m) Minimal (GitHub CLI, Git tools)
   d) DevOps (Docker, AWS, Terraform, K8s)
   n) None (skip component installation)

Select components (comma-separated numbers), preset letter, or press Enter for AI Development:
```

## Usage Examples:

### Default (interactive):
```bash
mcs create https://github.com/user/repo.git
# Interactive selector appears automatically
```

### Skip component selection:
```bash
mcs create https://github.com/user/repo.git --no-interactive
```

### Use a preset:
```bash
mcs create https://github.com/user/repo.git --preset ai-dev
```

### Specify components directly:
```bash
mcs create https://github.com/user/repo.git --components github-cli,claude,claude-flow
```

## Features:

1. **Smart Dependencies**: When you select a component with dependencies (like Claude Flow depends on Claude), the dependencies are automatically selected.

2. **Visual Feedback**: Selected components show a green dot (●) instead of an empty circle (○).

3. **Dependency Display**: When navigating to a component with dependencies, they're shown below the menu.

4. **Automatic Fallback**: If the terminal doesn't support interactive mode (like in CI/CD), it automatically falls back to the simple numbered list.

5. **Default Preset**: Pressing Enter without selection defaults to the AI Development preset (GitHub CLI, Claude, Claude Flow).