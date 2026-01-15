# Setup Usecase

Initializes .claude directory structure with hooks, agents, commands, and project-specific configuration. Detects project technology stacks (TypeScript, Go, Python, React Native, PHP) and generates appropriate principal-engineer agents by assembling embedded role and skill templates.

## Key Files

- **setup.go** - Main setup workflow orchestration
- **agents.go** - Agent profile assembly from embedded role, skill, and common design principle templates

## Key Types

- `SetupUseCase` - Orchestrates .claude directory setup workflow

## Embedded Profiles

Profiles (roles, skills, and agents) are embedded in the binary using Go's embed feature. This enables the setup to work without requiring a separate `make install` step or configuration directory at `~/.config/claudex/profiles`, making `npm install -g @claudex/cli` seamless.

## Stack Detection

The setup process automatically detects project technology stacks via marker files:

- **React Native**: app.json, react-native.config.js, metro.config.js
- **TypeScript/JavaScript**: tsconfig.json, package.json
- **Go**: go.mod
- **Python**: pyproject.toml, requirements.txt, setup.py, Pipfile
- **PHP**: composer.json, index.php, artisan

Detected stacks are used to generate corresponding principal-engineer agents.

## Usage

The `Execute` method creates the complete .claude directory structure:
1. Creates hooks/, agents/, and commands/agents/ directories
2. Installs hooks with dual-path support:
   - Primary: Copies hooks from ~/.config/claudex/hooks/ (for `make install` users)
   - Fallback: Installs hooks from embedded FS (for npm install users)
3. Copies agent profiles from embedded FS to both agents/ and commands/agents/
4. Detects project stacks using breadth-first file search (up to 3 levels deep)
5. Generates principal-engineer-{stack} agents for each detected stack by combining:
   - Engineer role template (profiles/roles/engineer.md)
   - Stack-specific skill (profiles/skills/{stack}.md, e.g., typescript.md)
   - Common design principles skill (profiles/skills/software-design-principles.md)
6. Creates principal-engineer.md alias pointing to primary stack's agent
7. Creates settings.local.json with hooks configuration

## Agent Assembly

The `AssembleEngineerAgent` function combines three profile components:

1. **Frontmatter** - Auto-generated with agent name, description, model, and color
2. **Engineer Role** - Base role template customized with stack display name (TypeScript, Go, Python, React Native)
3. **Stack Skill** - Stack-specific capabilities (e.g., TypeScript-specific patterns)
4. **Design Principles Skill** - Common design principles applied across all principal engineers (fail-fast, type safety, make illegal states unrepresentable, etc.)
