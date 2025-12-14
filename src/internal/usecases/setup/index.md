# Setup Usecase

Initializes .claude directory structure with hooks, agents, commands, and project-specific configuration. Detects project technology stacks (TypeScript, Go, Python, React Native) and generates appropriate principal-engineer agents by assembling role and skill templates.

## Key Files

- **setup.go** - Main setup workflow orchestration
- **agents.go** - Agent profile assembly from role and skill templates

## Key Types

- `SetupUseCase` - Orchestrates .claude directory setup workflow

## Stack Detection

The setup process automatically detects project technology stacks via marker files:

- **React Native**: app.json, react-native.config.js, metro.config.js
- **TypeScript/JavaScript**: tsconfig.json, package.json
- **Go**: go.mod
- **Python**: pyproject.toml, requirements.txt, setup.py, Pipfile

Detected stacks are used to generate corresponding principal-engineer agents.

## Usage

The `Execute` method creates the complete .claude directory structure:
1. Creates hooks/, agents/, and commands/agents/ directories
2. Copies hooks from ~/.config/claudex/hooks/
3. Copies agent profiles to both agents/ and commands/agents/
4. Detects project stacks using breadth-first file search (up to 3 levels deep)
5. Generates principal-engineer-{stack} agents for each detected stack
6. Creates principal-engineer.md alias pointing to primary stack's agent
7. Creates settings.local.json with hooks configuration
