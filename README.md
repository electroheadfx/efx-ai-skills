# Efx-ai-skills

> A beautiful TUI (Terminal User Interface) for discovering, previewing, and managing AI agent skills across multiple providers.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)

![efx-ai-skills screenshot](https://via.placeholder.com/800x450?text=efx-ai-skills+TUI)

## ✨ Features

- 🔍 **Unified Search** - Search skills from skills.sh and playbooks.com registries
- 👀 **Beautiful Preview** - View skill documentation with rendered Markdown
- 📦 **Centralized Storage** - One skill library (`~/.agents/skills/`) shared across all providers
- 🔗 **Smart Linking** - Symlink skills to multiple providers (Claude, Cursor, Qoder, Windsurf, Copilot)
- 🎨 **Intuitive TUI** - Built with Charm's [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- ⚡ **Fast & Lightweight** - Single Go binary, no dependencies
- 🔄 **Sync Management** - Keep skills in sync across all providers
- ⌨️  **Keyboard-First** - Navigate everything with vim-style keybindings

## 🎯 Concept

**efx-ai-skills** solves the problem of managing AI agent skills across different providers. Instead of downloading the same skill multiple times or manually copying files, it provides:

1. **Central Storage** - All skills stored once in `~/.agents/skills/`
2. **Provider Linking** - Symlinks to provider directories (`~/.claude/skills/`, `~/.cursor/skills/`, etc.)
3. **Easy Discovery** - Search 100+ skills from multiple registries
4. **Visual Management** - See which providers have which skills at a glance

## 📥 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/electroheadfx/efx-ai-skills.git
cd efx-ai-skills

# Build the binary
make build

# Install to your PATH
sudo mv bin/efx-skills /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/electroheadfx/efx-ai-skills@latest
```

### Prerequisites

- Go 1.22 or later (for building from source)
- Terminal with Unicode support
- Supported OS: macOS, Linux

## 🚀 Quick Start

### Launch the TUI

```bash
# Start the main interface (status view)
efx-skills
```

### Search for Skills

```bash
# Search from the TUI
efx-skills search

# Or search with a query
efx-skills search "react"
```

### Preview a Skill

```bash
# Preview skill documentation
efx-skills preview vercel-labs/agent-skills/react-best-practices
```

### View Status

```bash
# Show provider status
efx-skills status
```

## 🎮 Usage

### Navigation

**Status View** (default)
- `s` - Open search
- `↑/↓` - Navigate providers
- `Enter` / `m` - Manage provider skills
- `c` - Open configuration
- `r` - Refresh status
- `q` - Quit

**Search View**
- Type to search across registries
- `↵` - Execute search (when focused on input)
- `Tab` - Toggle focus between input and results
- `↑/↓` or `j/k` - Navigate results
- `p` or `Enter` - Preview selected skill
- `i` - Install skill
- `←/→` - Page navigation
- `Esc` - Back to status

**Preview View**
- `j/k` or `↑/↓` - Scroll line by line
- `Space/b` - Page down/up
- `g/G` - Jump to top/bottom
- `Esc` - Back to search

### CLI Commands

```bash
# Search skills
efx-skills search "authentication"

# Preview a skill
efx-skills preview yoanbernabeu/grepai-skills/find-skills

# Install a skill
efx-skills install <skill-name> -p claude -p cursor

# List installed skills
efx-skills list

# Show provider status
efx-skills status

# Sync all providers
efx-skills sync

# Manage configuration
efx-skills config

# Show version
efx-skills --version
```

## 📁 Directory Structure

```
~/.agents/
├── skills/                    # Central skill storage
│   ├── find-skills/
│   │   └── SKILL.md
│   ├── grepai-installation/
│   │   └── SKILL.md
│   └── ...
└── .skill-lock.json          # Lock file

~/.claude/skills/             # Symlinks to central storage
~/.cursor/skills/             # Symlinks to central storage
~/.qoder/skills/              # Symlinks to central storage
~/.windsurf/skills/           # Symlinks to central storage
```

## 🎨 Supported Providers

**efx-ai-skills** can manage skills for the following AI coding assistants:

- **Claude** (`~/.claude/skills/`) - Anthropic's Claude Desktop
- **Cursor** (`~/.cursor/skills/`) - Cursor AI Editor
- **Qoder** (`~/.qoder/skills/`) - Qoder AI Assistant
- **Windsurf** (`~/.windsurf/skills/`) - Windsurf Editor
- **GitHub Copilot** (`~/.copilot/skills/`) - GitHub Copilot
- **Cline** (`~/.cline/skills/`) - Cline VSCode Extension
- **Roo Code** (`~/.roo-code/skills/`) - Roo Code Extension
- **OpenCode** (`~/.opencode/skills/`) - OpenCode Assistant
- **Continue** (`~/.continue/skills/`) - Continue.dev Extension

Each provider can be individually enabled/disabled in the configuration.

## 🎨 Screenshots

### Status View
```
┌─────────────────────────────────────────────────────────┐
│  efx-skills v0.1.0                                      │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Provider Status                                        │
│  ─────────────────────────────────────────────────────  │
│                                                         │
│  Provider       Skills    Status                        │
│  ─────────────────────────────────────────────────────  │
│  ● claude         28      ✓ synced                      │
│  ● cursor          0      ✓ synced                      │
│  ● qoder          28      ✓ synced                      │
│  ○ windsurf        -      not configured                │
│                                                         │
│  Total: 28 skills in ~/.agents/skills/                  │
│                                                         │
│  [s] search  [c] configure  [r] refresh  [q] quit      │
└─────────────────────────────────────────────────────────┘
```

## 🔧 Configuration

Configuration is stored in `~/.config/efx-skills/config.json`:

```json
{
  "registries": [
    {
      "name": "skills.sh",
      "url": "https://skills.sh/api/search",
      "enabled": true
    },
    {
      "name": "playbooks.com",
      "url": "https://playbooks.com/api/skills",
      "enabled": true
    }
  ],
  "repos": [
    "yoanbernabeu/grepai-skills"
  ],
  "providers": [
    "claude",
    "cursor",
    "qoder"
  ]
}
```

## 🏗️ Architecture

**efx-ai-skills** uses a centralized storage model with provider linking:

1. **Skills are downloaded once** to `~/.agents/skills/`
2. **Providers reference via symlinks** to their respective directories
3. **TUI manages the relationships** between skills and providers
4. **Lock file tracks state** for consistency

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Cobra](https://github.com/spf13/cobra) - CLI framework

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

## 👤 Author

**Laurent Marques** (efx)

© 2026 efx - Laurent Marques

## 🙏 Acknowledgments

- [Charmbracelet](https://github.com/charmbracelet) for the amazing TUI libraries
- [skills.sh](https://skills.sh) and [playbooks.com](https://playbooks.com) for skill registries
- All skill authors contributing to the AI agent ecosystem

---

**Efx-ai-skills** - Making AI agent skill management simple and beautiful ✨
