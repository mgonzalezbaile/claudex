# Claudex

A modern, interactive session manager for Claude Code with AI-powered agent profiles.

## Features

- 🎨 Beautiful TUI built with Bubble Tea
- 📋 Session management - create, resume, and organize work sessions
- 🤖 AI-generated session names based on descriptions
- 🎭 Agent profiles - specialized Claude configurations for different tasks
- ⚡ Ephemeral mode for quick, unsaved sessions
- 🔍 Fuzzy search for sessions and profiles

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Claude CLI](https://docs.anthropic.com/claude-code)

## Quick Start

```bash
git clone https://github.com/YOUR_USERNAME/claudex.git
cd claudex/claudex
make install
```

This installs:
- Profiles and hooks to `~/.config/claudex/`
- Binary to `~/.local/bin/claudex`

Add to your shell config if needed:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Usage

```bash
claudex
```

The TUI will guide you through:
1. Session selection (new, ephemeral, or existing)
2. Profile selection (choose agent type)
3. Launch Claude with your selections

### Keyboard Controls

- `↑/↓` - Navigate
- `Enter` - Select
- `/` - Fuzzy search
- `q` or `Ctrl+C` - Quit

## Agent Profiles

Claudex includes specialized agent profiles:

| Profile | Purpose |
|---------|---------|
| `team-lead` | Strategic planning and orchestration |
| `architect` | System design and architecture |
| `researcher` | Deep analysis and investigation |
| `principal-engineer-{stack}` | Implementation (TypeScript, Python, Go) |
| `prompt-engineer` | Prompt design and optimization |

Profiles are automatically assembled based on your project's technology stack.

## Project Structure

```
claudex/
├── claudex/              # Main application
│   ├── main.go           # TUI application
│   ├── profiles/         # Agent profile definitions
│   │   ├── agents/       # Pre-built agents
│   │   ├── roles/        # Role templates
│   │   └── skills/       # Stack-specific skills
│   ├── .claude/hooks/    # Claude Code hooks
│   └── scripts/          # Installation scripts
├── LICENSE               # MIT License
└── README.md             # This file
```

## Development

### Building from source

```bash
cd claudex/claudex
make build      # Build binary
make run        # Build and run
make clean      # Clean artifacts
```

### Installation targets

```bash
make install          # Install to ~/.config/claudex and ~/.local/bin
make uninstall        # Remove installation
make install-project  # Install to current project's .claude/
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Credits

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
