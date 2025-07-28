# MCS Docker Images

This directory contains Dockerfiles for building MCS development environments with different language support.

## Image Variants

### Base Images (without Node.js)
- `mcs/code-server-base` - Base image with VS Code (code-server) only
- `mcs/code-server-python` - Python development environment
- `mcs/code-server-go` - Go development environment

### Images with Node.js
- `mcs/code-server-node` - Node.js development environment
- `mcs/code-server-python-node` - Python + Node.js
- `mcs/code-server-go-node` - Go + Node.js
- `mcs/code-server-full` - All languages (Python, Go, Node.js, Rust, Java)

## Why Multiple Images?

MCS now intelligently selects the appropriate image based on:
1. **Detected project language** - Analyzes your repository
2. **Selected components** - Components like Claude Code require Node.js

This approach provides:
- **Smaller images** for projects that don't need Node.js
- **Faster startup times**
- **Flexibility** - Users can opt out of components they don't need

## Building Images

```bash
# Build all images
./build.sh

# Build and push to registry
./build.sh --push
```

## Image Selection Logic

When creating a codespace, MCS:
1. Detects the primary language (Go, Python, Node.js, etc.)
   - Checks root directory first for language files (go.mod, package.json, etc.)
   - Also scans subdirectories for monorepos and nested projects
   - Supports common patterns like `mcs-go/`, `backend/`, `api/`, etc.
2. Checks which components are selected
3. Determines if Node.js is required (for Claude Code, Claude Flow)
4. Selects the optimal image:
   - Language-only image if no Node.js components
   - Language+Node image if Node.js components selected
   - Full image for complex multi-language projects

## Customization

To add support for a new language:
1. Create `Dockerfile.{language}`
2. Create `Dockerfile.{language}-node` variant
3. Update `languageImages` map in `internal/docker/compose.go`
4. Run `./build.sh` to build the new images