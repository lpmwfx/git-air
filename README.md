# Git Air 🚀

A simple, fully automatic Git daemon service for Linux and macOS that handles auto-commits, auto-pushes, and auto-pulls.

## Features

- **🔍 Auto Discovery**: Automatically finds all Git repositories in current directory and subdirectories
- **📝 Auto Commit**: Automatically commits all changes with timestamp or AI-generated messages
- **🤖 AI Commit Messages**: Optional AI-generated commit messages via Gemini or Claude CLI (with fallback)
- **🚀 Multi-Remote Push**: Pushes to ALL configured remotes for each repository
- **📡 Inter-Project Communication**: Pulls updates from remotes every minute
- **📚 Monorepo Support**: Syncs submodules before committing main repository
- **🏠 Dev Server Ready**: Perfect for development servers with multiple projects
- **⏱️ Configurable Interval**: Set check interval from 30 seconds to 30 minutes

## Quick Start

### 1. Running Git Air in a Project Terminal

**For single project monitoring:**
```bash
# Navigate to your project directory
cd /path/to/your/project

# Download or copy git-air binary to project
cp /path/to/git-air ./

# Start git-air (will monitor current directory and subdirectories)
./git-air
```

**This will:**
- 🔍 Monitor the current project and any Git repositories in subdirectories
- 📝 Auto-commit changes every 30 seconds if detected
- 🚀 Push to all configured remotes immediately
- 📡 Pull updates from remotes every minute

**With AI commit messages:**
```bash
# Use Gemini AI to generate descriptive commit messages (requires gemini CLI)
./git-air -ai

# Use Claude AI for commit messages (requires claude CLI)
./git-air -claude

# Combine with custom interval
./git-air -i 2 -claude  # Check every 2 minutes with Claude AI commits
```

### 2. Running Git Air as Ubuntu Service (Dev Server)

**For development server with multiple projects:**

#### Step 1: Install Git Air
```bash
# Copy git-air to system location
sudo cp git-air /usr/local/bin/
sudo chmod +x /usr/local/bin/git-air
```

#### Step 2: Create systemd service
```bash
# Create service file
sudo nano /etc/systemd/system/git-air.service
```

**Add this content:**
```ini
[Unit]
Description=Git Air - Automatic Git synchronization service
After=network.target
Wants=network.target

[Service]
Type=simple
User=your-username
Group=your-username
WorkingDirectory=/home/your-username
ExecStart=/usr/local/bin/git-air
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

#### Step 3: Enable and start service
```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start at boot
sudo systemctl enable git-air

# Start service now
sudo systemctl start git-air

# Check service status
sudo systemctl status git-air

# View logs
journalctl -u git-air -f
```

**This will:**
- 🏠 **Scan entire HOME directory** for all Git repositories
- 🚀 **Auto-sync ALL projects** continuously
- 🔄 **Start automatically** on server boot
- 📊 **Log all operations** to system journal

## Installation

### Prerequisites
- Go 1.21 or higher
- Git installed and configured
- (Optional) Gemini CLI or Claude CLI for AI-generated commit messages

### Build from Source
```bash
git clone <repository-url>
cd git-air
go build -o git-air
```

## How It Works

1. **Repository Discovery**: Scans for all `.git` directories recursively
2. **Auto Commit**: When changes are detected, automatically stages and commits them
3. **Multi-Remote Push**: After successful commits, pushes to ALL configured remotes
4. **Inter-Project Communication**: Every minute, checks all remotes for updates and pulls them
5. **Monorepo Handling**: For repositories with submodules, syncs all submodules before committing main repo

## Use Cases

### 🖥️ **Primary Use Case: Development Server**
- **Place git-air in ~/home** on your development server
- **Manages all your projects automatically** - no manual setup per project
- **Multi-remote support** - pushes to origin, backup, mirror, etc.
- **Perfect for teams** - everyone's changes stay synchronized
- **Never lose work** - continuous auto-commits serve as safety net

### 🔄 **Other Use Cases**
- **Development Workflow**: Never lose work with automatic commits
- **Team Collaboration**: Keep repositories synchronized automatically  
- **CI/CD Integration**: Ensure changes are always pushed to trigger builds
- **Backup Strategy**: Automatic commits serve as frequent backups
- **Inter-project Communication**: Auto-sync between related repositories

## Architecture

The service consists of a simple, single-file Go application that:

- **Repository Scanner**: Discovers Git repositories recursively
- **Change Monitor**: Checks for uncommitted changes every 30 seconds
- **Git Operations**: Handles commits, pushes to all remotes, and pulls
- **Monorepo Support**: Syncs submodules before main repository commits
- **Inter-Project Sync**: Pulls from all remotes every minute

## Security Considerations

- Git Air only operates on local repositories using standard Git commands
- Uses existing Git configuration (credentials, remotes, etc.)
- Excludes common non-source directories (node_modules, vendor)
- No direct repository manipulation - relies on Git CLI

## License

MIT License - see LICENSE file for details
