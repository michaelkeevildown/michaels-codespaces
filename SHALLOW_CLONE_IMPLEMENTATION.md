# Shallow Clone Implementation for Large Repositories

## Overview

This implementation adds automatic shallow cloning for large repositories to prevent the `mcs create` command from appearing stuck when cloning repositories like Homebrew/homebrew-core.

## Changes Made

### 1. Enhanced Clone Module (`scripts/modules/github/clone/github-clone.sh`)

- Added a list of known large repositories that automatically use shallow clone:
  - homebrew/homebrew-core
  - homebrew/homebrew-cask
  - torvalds/linux
  - chromium/chromium
  - microsoft/vscode
  - llvm/llvm-project

- Modified `clone_repository()` function to:
  - Auto-detect large repositories and use `--depth 1` by default
  - Show real-time progress during cloning
  - Support `force_shallow` parameter
  - Add verbose output with `--progress --verbose` flags

- Updated `clone_with_retry()` to pass through all clone parameters

### 2. Updated Create Script (`scripts/core/create-codespace.sh`)

- Added new command-line options:
  - `--shallow`: Force shallow clone (depth=1)
  - `--depth <n>`: Specify clone depth (0 = full history)

- Modified to pass clone options through to the clone module

### 3. Updated MCS Command (`bin/mcs`)

- Added documentation for new clone options in help text
- Added examples showing how to use shallow cloning

## Usage Examples

### Automatic shallow clone for Homebrew (large repo):
```bash
mcs create https://github.com/Homebrew/homebrew-core.git
```
This will automatically use `--depth 1` and show a warning.

### Force shallow clone for any repository:
```bash
mcs create https://github.com/facebook/react.git --shallow
```

### Specify custom clone depth:
```bash
mcs create https://github.com/torvalds/linux.git --depth 10
```

### Force full clone of a large repository:
```bash
mcs create https://github.com/homebrew/homebrew-core.git --depth 0
```

## Testing on VM

1. **Update the branch on your VM:**
   ```bash
   cd ~/.michaels-codespaces
   git pull origin enhance-container-creation
   ```

2. **Test with Homebrew (should now use shallow clone automatically):**
   ```bash
   mcs create https://github.com/Homebrew/homebrew-core.git --force --debug
   ```
   
   You should see:
   - "Detected large repository. Using shallow clone (depth=1) for faster cloning."
   - Real-time progress updates during cloning
   - Much faster completion (minutes instead of hours)

3. **Test with verbose output:**
   ```bash
   mcs create https://github.com/Homebrew/homebrew-core.git --force --verbose
   ```

4. **Check the clone depth:**
   ```bash
   cd ~/codespaces/homebrew-homebrew-core/src
   git log --oneline | wc -l
   ```
   Should show only 1 commit for shallow clone.

## Benefits

1. **Faster cloning**: Homebrew-core goes from ~6GB full clone to ~200MB shallow clone
2. **Better user experience**: Real-time progress shown during clone
3. **Automatic optimization**: Large repos automatically use shallow clone
4. **User control**: Can override with `--depth` option

## Future Enhancements

The remaining todo item (repository size detection via GitHub API) could be implemented to:
- Show estimated repository size before cloning
- Warn users about large downloads
- Suggest shallow clone for repositories over a certain size

This would require authenticated API calls to get repository statistics.