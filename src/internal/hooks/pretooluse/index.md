# hooks/pretooluse

Context injection hook that modifies Task tool prompts with session folder information.

## Key Files

- **context_injector.go** - Handler for PreToolUse events with session context injection
- **context_injector_test.go** - Test suite for context injection logic

## Key Types

- `Handler` - Processes PreToolUse events and injects session context into Task tool prompts

## Behavior

1. Only modifies `Task` tool invocations (all other tools pass through unchanged)
2. Detects agent type (subagent_type, case-insensitive) and provides specialized context:
   - **Explore agents** (subagent_type="Explore"): Receive LSP/MCP tool instructions only
   - **Plan agents** (subagent_type="Plan"): Receive planning context + detected tech stack skills
   - **Other agents**: Receive session context with documentation loading procedures
3. Finds session folder using `session.FindSessionFolder()`
4. Builds appropriate markdown context block:
   - For Explore agents: LSP (code navigation), Context7 (library docs), Sequential Thinking instructions
   - For Plan agents: MCP tools, execution plan structure, phase/track labeling, detected tech stack skills
   - For other agents: Session path, mandatory rules, activation procedure (3-step doc loading)
5. Plan agents detect tech stack (Go, TypeScript, etc.) and inject relevant skill guidance
6. Uses pointer-based approach: references `session-overview.md` if available; falls back to file enumeration
7. Injects context before original prompt using `UpdatedInput` field
8. Returns "allow" with modified prompt for Task tools

## Context Injection Formats

### For Standard Agents (Session Context)

```markdown
## SESSION CONTEXT (CRITICAL)

You are working within an active Claudex session.
ALL documentation, plans, and artifacts MUST be created in the session folder.

**Session Folder (Absolute Path)**: `{sessionPath}`

### MANDATORY RULES for Documentation:
1. ✅ ALWAYS save documentation to the session folder above
2. ✅ Use absolute paths when creating files (Write/Edit tools)
3. ✅ Before exploring the codebase, check the session folder for existing context
4. ❌ NEVER save documentation to project root or arbitrary locations
5. ❌ NEVER use relative paths for documentation files

### Session Folder Contents:
(Pointer-based approach when session-overview.md exists)
- {sessionPath}/session-overview.md

OR (File enumeration fallback):
- file1.md
- file2.md

### ACTIVATION PROCEDURE (Execute on Session Start)

Before beginning any task work, execute this mandatory 3-step loading sequence:

**STEP 1: Load Session Context**
- Read `{sessionPath}/session-overview.md` using the Read tool

**STEP 2: Load Root Doc Files**
- Read ALL files listed under "Root Documentation Entry Points" below
- Use Read tool for each file (do NOT use Glob/Grep for discovery)

**STEP 3: Recursive Index Traversal (Task-Driven)**
- Each doc file contains links to other doc files in subdirectories
- CRITICAL: Load only the files that are directly related and relevant to the task at hand

### Recommended File Names:
- {CLAUDEX_DOC_PATHS entries}

---

## ORIGINAL REQUEST

{original prompt}
```

### For Explore Agents (MCP/LSP Instructions)

```markdown
## EXPLORE AGENT ENHANCEMENTS

You have access to powerful tools for codebase exploration. Use them effectively.

### LSP Tool (PREFERRED for code navigation)
Use LSP instead of brute-force Glob/Grep when possible:
- `goToDefinition`: Jump to where a symbol is defined
- `findReferences`: Find all usages of a symbol
- `hover`: Get documentation and type info for a symbol
- `documentSymbol`: List all symbols in a file
- `workspaceSymbol`: Search symbols across the codebase
- `incomingCalls`/`outgoingCalls`: Trace call hierarchy

**Parameters**: `operation`, `filePath` (absolute), `line`, `character`

### Context7 MCP (for library documentation)
Before making assumptions about libraries/frameworks, query current docs:
1. `mcp__context7__resolve-library-id`: Get library ID (e.g., "redis" → "/redis/redis")
2. `mcp__context7__query-docs`: Query specific documentation

**Constraint**: Max 3 calls per question

### Sequential Thinking MCP (for complex analysis)
Use `mcp__sequential-thinking__sequentialthinking` for multi-step problem solving and trade-off analysis.

### Exploration Best Practices
1. Start with LSP `workspaceSymbol` to find entry points
2. Use `goToDefinition` to trace implementations
3. Use `findReferences` to understand usage patterns
4. Fall back to Glob/Grep only for pattern-based searches
5. Cite findings with file:line format

---

## ORIGINAL REQUEST

{original prompt}
```

### For Plan Agents (Planning Context + Tech Stack Skills)

```markdown
## PLAN AGENT ENHANCEMENTS

You are creating an execution plan. Use these tools and practices.

### MCP Tools (MANDATORY)

**Context7 MCP** - Query documentation for all libraries/frameworks:
1. `mcp__context7__resolve-library-id`: Get library ID
2. `mcp__context7__query-docs`: Query specific documentation

**Sequential Thinking MCP** - Use for parallelization analysis:
- Component boundary identification
- Dependency mapping (what blocks what)
- Shared contract discovery
- Parallel opportunity grouping (Track A/B/C)
- Sequential constraint justification

### Execution Plan Structure

**Phase Labeling** (MANDATORY):
- `### Phase N: [Name] (Parallel: X independent tracks)`
- `### Phase N: [Name] (Sequential)` with justification

**Track Groupings** for parallel phases:
```
Track A: [task1, task2]
Track B: [task3, task4]
```

**Architect Boundaries**:
- Define WHAT to build and HOW to approach it
- Code snippets: Max 15 lines for patterns, NOT full implementations
- Use file:line pointers when referencing existing code

### Detected Tech Stack Skills

When multiple tech stacks are detected, skill-specific guidance is injected for:
- **Go** - Go-specific patterns, best practices, and standard library usage
- **TypeScript** - TypeScript-specific patterns, type safety, and framework guidance
- (Other stacks detected based on project markers: go.mod, package.json, etc.)

---

## ORIGINAL REQUEST

{original prompt}
```

## Usage

Hook is invoked automatically by Claude Code before each tool execution. Executable located at `.claude/hooks/PreToolUse`.
