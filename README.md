# Claudex

## Features

### 🗂️ Persistent Sessions
Work across days, weeks, or months without losing context. Claudex sessions preserve all research, plans, and artifacts in organized folders—even when Claude's memory resets. Fork sessions to explore alternatives, or use **fresh memory** to start a new conversation while keeping everything you've built.

### 📝 Auto-Documentation
A background agent silently maintains a living overview of your session. Every decision, discovery, and milestone is captured automatically—no manual note-taking required. Pick up any project instantly, even after weeks away.

### 🤖 Parallel Agent Orchestration
A team-lead agent coordinates specialized researchers, architects, and engineers. Work gets planned with parallelization in mind, then multiple engineers execute simultaneously on independent tracks. Ship faster with systematic divide-and-conquer.

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Claude CLI](https://docs.anthropic.com/claude-code)

## Quick Start

```bash
git clone https://github.com/YOUR_USERNAME/claudex.git
cd claudex/claudex
make install
```

Add to your shell config if needed:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Usage

Navigate to your project directory and run:

```bash
cd /path/to/your/project
claudex
```

On first run, claudex creates a `.claude` folder with agent profiles and hooks. If a `.claude` folder already exists, files are merged (use `--no-overwrite` to preserve your existing files).

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

## Configuration

Create a `.claudex.toml` file in your project root to customize behavior:

```toml
# Documentation files always loaded into context
doc = ["docs/index.md"]

# Preserve existing .claude files during setup
no_overwrite = true

[features]
# Auto-documentation during session (default: true)
autodoc_session_progress = true

# Auto-documentation on session end (default: true)
autodoc_session_end = true

# Tool executions between doc updates (default: 5)
autodoc_frequency = 5
```

Environment variables override config values: `CLAUDEX_AUTODOC_SESSION_PROGRESS`, `CLAUDEX_AUTODOC_SESSION_END`, `CLAUDEX_AUTODOC_FREQUENCY`.

**Tip:** Keep `doc` files lightweight. Use an `index.md` with brief descriptions and file paths—Claude will load details on demand, saving context for actual work.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Credits

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
