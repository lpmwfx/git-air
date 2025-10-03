# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Git Air is a single-file Go daemon service that automatically manages Git repositories through continuous synchronization. It auto-commits changes, pushes to multiple remotes, and pulls updates for inter-project communication.

**Primary use case**: Development servers managing multiple Git repositories from a home directory.

## Build & Run Commands

```bash
# Build the binary
go build -o git-air

# Show help screen
./git-air -h
./git-air --help

# Run with default settings (30 second check interval)
./git-air

# Run with custom interval (0.5-30 minutes)
./git-air -i 1        # Check every 1 minute
./git-air -i 5        # Check every 5 minutes
./git-air --interval 10  # Check every 10 minutes

# Force monorepo mode (auto-detects if not specified)
./git-air -mr
./git-air --monorepo

# Combine flags
./git-air -i 5 -mr    # 5 minute interval, force monorepo mode

# Run on development server (typically from home directory)
cd ~ && ./git-air -i 2
```

### Command-Line Flags

- `-h`, `--help`: Show help screen and exit
- `-i`, `--interval <minutes>`: Set check interval (0.5-30 minutes, default: 0.5)
- `-mr`, `--monorepo`: Force monorepo mode (auto-detects by default)
- `-ai`, `--ai-commits`: Use AI-generated commit messages via gemini CLI (requires gemini CLI installed)

## Architecture

### Single-File Design
The entire application is in `main.go` (~285 lines). This is intentional - the project follows a simple, monolithic approach.

### Core Flow
1. **Repository Discovery** (`findGitRepos`): Recursively scans for `.git` directories, excluding `node_modules` and `vendor`
2. **Main Loop**:
   - Every 30 seconds: Check all repos for changes, commit, and push to ALL remotes
   - Every 60 seconds: Pull from all remotes for inter-project communication
3. **Monorepo Handling**: Detects submodules via `.gitmodules` or nested `.git` directories, syncs submodules BEFORE committing parent repo

### Key Functions
- `processRepo()`: Main processing logic - handles monorepo sync, auto-commit (with optional AI), multi-remote push
- `generateAICommitMessage()`: Calls gemini CLI with git diff, 30s timeout, returns AI-generated message
- `isMonorepo()`: Detects if repo has submodules or nested repos
- `syncSubmodules()`: Updates submodules before main repo commit
- `pushToAllRemotes()`: Pushes to every configured remote (origin, backup, mirror, etc.)
- `pullFromRemotes()`: Pulls from all remotes for inter-project updates

### Multi-Remote Strategy
Unlike standard Git tools, Git Air pushes to **ALL** configured remotes, not just origin. This enables:
- Automatic backup to multiple locations
- Mirror synchronization
- Multi-location deployment

### Inter-Project Communication
The 60-second pull cycle enables projects to communicate via Git:
- Project A commits data/config changes
- Git Air pushes to remote
- Git Air pulls updates to Project B
- Project B sees changes automatically

## Development Notes

### No Dependencies
Uses only Go standard library (`os`, `exec`, `path/filepath`, `time`). Module declaration in `go.mod` specifies Go 1.21.

### Error Handling Philosophy
- **Validation**: Validates interval range (0.5-30 minutes) at startup, shows help and exits on invalid input
- **Discovery**: Silent failures for discovery (`return nil` in walk functions), skips inaccessible directories
- **Git Operations**: Boolean returns with visual feedback (✓ for success, ❌ for errors)
- **Directory Changes**: Explicit error checking with deferred restoration of working directory
- **Resilience**: Continues processing other repos if one fails
- **Recovery**: No explicit error recovery - relies on next cycle to retry failed operations
- **User Feedback**: Clear status messages with emojis for quick visual parsing

### Commit Messages

**Standard Mode (default):**
- Standard: `"auto commit - {timestamp}"`
- Monorepo: `"auto commit (monorepo) - {timestamp}"`
- Format: `2006-01-02 15:04:05`

**AI Mode (-ai flag):**
- Uses gemini CLI to generate descriptive commit messages
- Analyzes git diff and creates concise, imperative mood messages
- 30-second timeout per request
- Automatically falls back to timestamp commits on error
- Example: `"Add command-line flag parsing"` or `"Refactor error handling logic"`

### Directory Exclusions
Hardcoded exclusions in `findGitRepos()`:
- `node_modules/`
- `vendor/`
- `.git/` (skipped after detection)

## Testing Strategy

No automated tests exist (see TODO). Testing is done manually:
1. Create test repos with/without submodules
2. Configure multiple remotes
3. Run git-air and verify commit/push/pull behavior
4. Check monorepo submodule sync order

## Production Deployment

### Systemd Service (Ubuntu/Linux)
```bash
# Install binary
sudo cp git-air /usr/local/bin/
sudo chmod +x /usr/local/bin/git-air

# Create service file at /etc/systemd/system/git-air.service
# Set WorkingDirectory to user home directory
# Set User/Group to appropriate user

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable git-air
sudo systemctl start git-air

# Monitor
journalctl -u git-air -f
```

### Key Service Configuration
- `WorkingDirectory`: Set to home directory (e.g., `/home/username`) to scan all projects
- `Restart=always` with `RestartSec=10` for automatic recovery
- Logs to systemd journal

## Output & Logging

Git Air provides rich console output with emoji-based visual feedback:
- 🚀 = Starting operations or pushing
- 📡 = Inter-project communication (pulls)
- 📝 = Committing changes
- 📦 = Syncing submodules
- 📥 = Checking for updates
- ✓ = Success
- ❌ = Error/failure
- ⚠️ = Warning
- 💤 = Sleeping between cycles
- 🔄 = Check cycle number

Example output format:
```
🔄 Check cycle #3
📝 hugo-norsetinge [MONOREPO]: Auto committing changes...
  ✓ Committed changes in hugo-norsetinge
  🚀 Pushing to origin... ✓
  ✓ Successfully pushed to 1/1 remotes
```

## Known Limitations (from TODO)

- No smart merge strategies (uses default Git merge)
- No performance optimization for large repository sets
- No automated testing suite
