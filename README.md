# Claudex

The Claude Experience you wish you had.

> **Important: Claude Pro/Max/Team subscription required**
>
> Claudex requires [Claude Code](https://docs.anthropic.com/claude-code), which is only available with a Claude Pro, Max, or Team subscription. The free Claude tier does not include Claude Code access.

## Features

### 🗂️ Persistent Sessions

Every task starts with a session — a folder that accumulates everything Claude produces:

```
.claudex/sessions/
└── api-refactor-abc123/
    ├── session-overview.md    ← Auto-maintained status & index
    ├── feature-description.md ← Manually added from Jira, Linear, etc.
    ├── research-findings.md   ← Research artifacts
    ├── execution-plan.md      ← Architecture decisions
    └── ...                    ← Your custom docs
```

**Why it matters:** Claude's context window fills up. When you clear it, Claude normally forgets everything. With claudex, the session folder persists — Claude reads `session-overview.md` on startup and catches up in seconds.

**Session modes:**
- **Resume** — Continue where you left off with full claude's conversation history
- **Fresh memory** — Clear claude's context window, keep all docs (Claude catches up via overview)
- **Fork** — Branch into a new task while cloning all the docs

### 📝 Auto-Documentation

A background agent silently maintains `session-overview.md` as you work—no manual note-taking:

```
┌───────────────────────────────────────────────────────────────────────┐
│  You work normally                                                    │
│       ↓                                                               │
│  Every few messages, claudex updates the session-overview.md document │
│       ↓                                                               │
│  Clear claude's context window                                        │
│       ↓                                                               │
│  Leverage full claude's potential                                     │
└───────────────────────────────────────────────────────────────────────┘
```

Example of auto-maintained `session-overview.md`:

```markdown
# Session: API Refactor

## Status
Phase 2 in progress - Authentication endpoints complete

## Key Decisions
- JWT over session cookies (see research-auth.md)
- Rate limiting at gateway level

## Documents
- [research-auth.md](./research-auth.md) — Auth strategy analysis
- [execution-plan.md](./execution-plan.md) — Implementation phases
```

Pick up any session instantly—even weeks later. Claude reads the overview, follows the pointers, and catches up in seconds.

### 📚 Auto-Updating Index Files

Keep your codebase documentation up-to-date automatically. On first run in a git repo, claudex offers to install a post-commit hook:

```
📝 Enable auto-docs update after git commits? [y/n/never]:
```

When enabled, after each commit:
1. Detects which files changed
2. Identifies affected `index.md` files
3. Spawns Claude (Haiku) to intelligently update them

```
┌─────────────────────────────────────────────────────────────────┐
│  git commit                                                      │
│       ↓                                                          │
│  Post-commit hook triggers claudex --update-docs                 │
│       ↓                                                          │
│  Claude reads existing index.md, explores surrounding context    │
│       ↓                                                          │
│  Makes thoughtful updates to keep docs relevant                  │
└─────────────────────────────────────────────────────────────────┘
```

**Manual trigger:** `claudex --update-docs`

**Skip for a commit:** `CLAUDEX_SKIP_DOCS=1 git commit -m "quick fix"`

### 🤖 Parallel Agent Orchestration

A team-lead agent coordinates specialists through a structured workflow:

```
┌─────────────────────────────────────────────────────┐
│  You describe what you need                         │
│       ↓                                             │
│  Explore agent investigates codebase & docs         │
│       ↓                                             │
│  Plan agent creates execution plan with phases      │
│       ↓                                             │
│  Engineers execute in parallel:                     │
│       ├── Track A: Auth service                     │
│       ├── Track B: API endpoints                    │
│       └── Track C: Database migrations              │
└─────────────────────────────────────────────────────┘
```

![Parallel Agent Execution](assets/agents-parallel-exec.png)

Work gets broken into independent tracks. Multiple engineers execute simultaneously: divide-and-conquer.

## Prerequisites

- **Claude Pro, Max, or Team subscription** — Required for Claude Code access
- [Claude Code CLI](https://docs.anthropic.com/claude-code) — Install via `npm install -g @anthropic-ai/claude-code`
- [Node.js 14+](https://nodejs.org/) — For npm installation
- [Go 1.21+](https://go.dev/dl/) — Only needed if building from source

### Recommended MCPs

On first run, claudex will prompt you to install these recommended MCPs for the best experience:

| MCP | Description |
|-----|-------------|
| [Sequential Thinking](https://github.com/modelcontextprotocol/servers/tree/main/src/sequentialthinking) | Structured reasoning through complex problems step-by-step |
| [Context7](https://github.com/upstash/context7) | Up-to-date documentation lookup for libraries and frameworks |

You can also configure them manually anytime with `claudex --setup-mcp` or `make install-mcp`.

## Installation

### npm (Recommended)

```bash
npm install -g @claudex/cli
```

This works on macOS and Linux without needing to clone the repository.

### From Source

```bash
git clone https://github.com/mgonzalezbaile/claudex.git
cd claudex
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
| `principal-engineer-{stack}` | Implementation (TypeScript, Python, Go, PHP) |
| `prompt-engineer` | Prompt design and optimization |

Profiles are automatically assembled based on your project's technology stack.

## Configuration

Claudex stores its artifacts in a `.claudex/` folder in your project root:

```
.claudex/
├── config.toml      # Configuration file (auto-created)
├── sessions/        # Session data
├── logs/            # Log files
└── preferences.json # User preferences
```

### Customizing Behavior

Edit `.claudex/config.toml` to customize behavior:

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

**Tip:** Keep `doc` files lightweight—they're passed to every agent. Use an index with brief descriptions and pointers:

```markdown
# Project Documentation Index

## Product
- [docs/product-overview.md](docs/product-overview.md) — Business goals, user personas, success metrics

## Technology
- [docs/architecture.md](docs/architecture.md) — System design, service boundaries, data flow
- [docs/tech-stack.md](docs/tech-stack.md) — Languages, frameworks, infrastructure choices

## Development
- [docs/coding-standards.md](docs/coding-standards.md) — Style guide, patterns, conventions
- [docs/testing-strategy.md](docs/testing-strategy.md) — Test types, coverage requirements
```

Claude reads the index, understands what's available, and loads detailed docs on demand—saving context for actual work.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Credits

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Readline](https://github.com/chzyer/readline) - Readline implementation
