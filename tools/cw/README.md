# cw - Workspace Navigator

A fast, single-binary CLI tool to navigate your workspace and launch Claude in any directory.

## Features

- 🗂️ **Interactive folder navigation** - Browse directories with keyboard
- 📁 **Parent directory traversal** - Navigate up with `..`
- 🤖 **Claude integration** - Start Claude in any directory with "Run Claude Here"
- ⚡ **Cross-platform** - Works on Linux, macOS, Windows
- 🚀 **No dependencies** - Single standalone binary
- 🔧 **Configurable** - Set default workspace once, use everywhere

## Installation

### Linux / macOS

#### Option 1: Build from Source
```bash
cd tools/cw
go build -o cw ./cmd/cw
mv cw ~/.local/bin/
```

#### Option 2: From Parent Directory
```bash
cd ../../
go build -o cw ./tools/cw/cmd/cw
mv cw ~/.local/bin/
```

### Windows
```cmd
cd tools\cw
go build -o cw.exe .\cmd\cw
REM Move cw.exe to PATH directory
```

## Quick Start

### First Run - Set Your Workspace
```bash
cw -r ~/workspace
# Saves to ~/.cw
```

### Then Just Use It
```bash
cw
# Select folder → "Run Claude Here" → Claude opens in selected directory
```

## Usage

```bash
# Navigate and open Claude
cw

# Resume last Claude session
cw --resume

# Set different workspace
cw -r ~/projects

# Use with custom root (one-off)
CW_ROOT=/tmp cw

# Select account interactively
cw -a

# Use specific account
cw --account myaccount
```

## Configuration

### Set Default Workspace

**Linux / macOS:**
```bash
cw -r ~/workspace
# Saved to ~/.cw
```

**Windows (cmd):**
```cmd
cw -r %USERPROFILE%\workspace
REM Saved to %USERPROFILE%\.cw
```

**Windows (PowerShell):**
```powershell
cw -r $env:USERPROFILE\workspace
# Saved to $env:USERPROFILE\.cw
```

### Priority Order
1. `-r` / `--root` flag (one-time override)
2. `~/.cw` or `%USERPROFILE%\.cw` config file (persistent)
3. `$CW_ROOT` environment variable
4. Home directory (fallback)

## How It Works

1. **Start navigator** - Shows folders in configured root directory
2. **Navigate** - Use arrow keys, type to search
3. **Select folder** - Press Enter to select
4. **Run Claude** - Choose "Run Claude Here" or pick subfolder
5. **Claude starts** - In the selected directory with `cd` and `claude` command

### Keyboard Navigation
- `↑` / `↓` - Navigate
- `/` - Search
- `Enter` - Select
- `Esc` - Cancel

## Integration with Claude

```bash
# Pass arguments to Claude
cw --resume
# → Runs: cd /selected/path && claude --resume

# Any Claude CLI arguments work
cw --model claude-3-sonnet
# → Runs: cd /selected/path && claude --model claude-3-sonnet
```

## Requirements

- **Shell**: bash, zsh, sh (Linux/macOS) or cmd.exe, PowerShell (Windows)
- **Claude CLI** installed and in PATH (`claude` command)
- **Go 1.21+** (for building from source)

## Supported Shells

- **Linux/macOS**: bash, zsh, fish, sh, etc.
- **Windows**: cmd.exe, PowerShell, PowerShell Core, Git Bash

## Development

```bash
# Build locally
go build -o cw ./cmd/cw

# Test
./cw -r ~/workspace
./cw
```

## License

MIT
