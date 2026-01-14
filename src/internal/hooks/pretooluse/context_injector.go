package pretooluse

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"claudex"
	"claudex/internal/hooks/shared"
	"claudex/internal/services/session"
	"claudex/internal/services/stackdetect"

	"github.com/spf13/afero"
)

// Handler processes PreToolUse hook events
// It injects session context into Task tool invocations
type Handler struct {
	fs     afero.Fs
	env    shared.Environment
	logger *shared.Logger
}

// NewHandler creates a new Handler instance
func NewHandler(fs afero.Fs, env shared.Environment, logger *shared.Logger) *Handler {
	return &Handler{
		fs:     fs,
		env:    env,
		logger: logger,
	}
}

// Handle processes PreToolUse events
// Returns updatedInput for Task tools with session context injected
// Returns allow with no modification for non-Task tools
func (h *Handler) Handle(input *shared.PreToolUseInput) (*shared.HookOutput, error) {
	// Log the tool being invoked
	if h.logger != nil {
		_ = h.logger.Logf("Processing PreToolUse for tool: %s", input.ToolName)
	}

	// Only modify Task tool invocations
	if input.ToolName != "Task" {
		if h.logger != nil {
			_ = h.logger.Logf("Tool %s is not Task, passing through unchanged", input.ToolName)
		}
		return &shared.HookOutput{
			HookSpecificOutput: shared.HookSpecificOutput{
				HookEventName:      "PreToolUse",
				PermissionDecision: "allow",
			},
		}, nil
	}

	// Find session folder
	sessionPath, err := session.FindSessionFolder(h.fs, h.env, input.SessionID)
	if err != nil {
		// No session found - return allow without modification
		if h.logger != nil {
			_ = h.logger.Logf("No session folder found: %v", err)
		}
		return &shared.HookOutput{
			HookSpecificOutput: shared.HookSpecificOutput{
				HookEventName:      "PreToolUse",
				PermissionDecision: "allow",
			},
		}, nil
	}

	if h.logger != nil {
		_ = h.logger.Logf("Session folder found: %s", sessionPath)
	}

	// Get the original prompt
	originalPrompt, ok := input.ToolInput["prompt"].(string)
	if !ok || originalPrompt == "" {
		if h.logger != nil {
			_ = h.logger.LogInfo("No prompt found in tool_input, passing through unchanged")
		}
		return &shared.HookOutput{
			HookSpecificOutput: shared.HookSpecificOutput{
				HookEventName:      "PreToolUse",
				PermissionDecision: "allow",
			},
		}, nil
	}

	// Check if this is an Explore agent - they get specialized context
	subagentType, _ := input.ToolInput["subagent_type"].(string)
	if strings.EqualFold(subagentType, "Explore") {
		if h.logger != nil {
			_ = h.logger.Logf("Explore agent detected, injecting MCP/LSP instructions")
		}

		exploreContext := h.buildExploreContext()
		modifiedPrompt := fmt.Sprintf("%s\n\n---\n\n## ORIGINAL REQUEST\n\n%s", exploreContext, originalPrompt)

		updatedInput := make(map[string]interface{})
		for k, v := range input.ToolInput {
			updatedInput[k] = v
		}
		updatedInput["prompt"] = modifiedPrompt

		return &shared.HookOutput{
			HookSpecificOutput: shared.HookSpecificOutput{
				HookEventName:      "PreToolUse",
				PermissionDecision: "allow",
				UpdatedInput:       updatedInput,
			},
		}, nil
	}

	// Check if this is a Plan agent - they get planning context + stack skills
	if strings.EqualFold(subagentType, "Plan") {
		if h.logger != nil {
			_ = h.logger.Logf("Plan agent detected, injecting planning context + stack skills")
		}

		// Detect tech stacks
		stacks := stackdetect.Detect(h.fs, input.CWD)

		planContext := h.buildPlanContext(stacks)
		modifiedPrompt := fmt.Sprintf("%s\n\n---\n\n## ORIGINAL REQUEST\n\n%s", planContext, originalPrompt)

		updatedInput := make(map[string]interface{})
		for k, v := range input.ToolInput {
			updatedInput[k] = v
		}
		updatedInput["prompt"] = modifiedPrompt

		return &shared.HookOutput{
			HookSpecificOutput: shared.HookSpecificOutput{
				HookEventName:      "PreToolUse",
				PermissionDecision: "allow",
				UpdatedInput:       updatedInput,
			},
		}, nil
	}

	// Get doc paths from environment
	docPathsStr := h.env.Get("CLAUDEX_DOC_PATHS")
	var docPaths []string
	if docPathsStr != "" {
		docPaths = strings.Split(docPathsStr, ":")
	}

	// Build session context
	sessionContext, err := h.buildSessionContext(sessionPath, docPaths, input.CWD)
	if err != nil {
		if h.logger != nil {
			_ = h.logger.LogError(fmt.Errorf("failed to build session context: %w", err))
		}
		// On error, pass through without modification
		return &shared.HookOutput{
			HookSpecificOutput: shared.HookSpecificOutput{
				HookEventName:      "PreToolUse",
				PermissionDecision: "allow",
			},
		}, nil
	}

	// Build the modified prompt
	modifiedPrompt := fmt.Sprintf("%s\n\n---\n\n## ORIGINAL REQUEST\n\n%s", sessionContext, originalPrompt)

	// Create updated input with modified prompt
	updatedInput := make(map[string]interface{})
	for k, v := range input.ToolInput {
		updatedInput[k] = v
	}
	updatedInput["prompt"] = modifiedPrompt

	if h.logger != nil {
		_ = h.logger.Logf("Injected session context into Task tool prompt")
	}

	return &shared.HookOutput{
		HookSpecificOutput: shared.HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
			UpdatedInput:       updatedInput,
		},
	}, nil
}

// buildSessionContext creates the markdown context block
func (h *Handler) buildSessionContext(sessionPath string, docPaths []string, projectRoot string) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("## SESSION CONTEXT (CRITICAL)\n\n")
	sb.WriteString("You are working within an active Claudex session. ")
	sb.WriteString("ALL documentation, plans, and artifacts MUST be created in the session folder.\n\n")
	sb.WriteString(fmt.Sprintf("**Session Folder (Absolute Path)**: `%s`\n\n", sessionPath))

	// Mandatory rules
	sb.WriteString("### MANDATORY RULES for Documentation:\n")
	sb.WriteString("1. ✅ ALWAYS save documentation to the session folder above\n")
	sb.WriteString("2. ✅ Use absolute paths when creating files (Write/Edit tools)\n")
	sb.WriteString("3. ✅ Before exploring the codebase, check the session folder for existing context\n")
	sb.WriteString("4. ❌ NEVER save documentation to project root or arbitrary locations\n")
	sb.WriteString("5. ❌ NEVER use relative paths for documentation files\n\n")

	// Check for session-overview.md - if exists, use pointer; otherwise fallback to enumeration
	overviewPath := filepath.Join(sessionPath, "session-overview.md")
	overviewExists, err := afero.Exists(h.fs, overviewPath)
	if err != nil {
		return "", fmt.Errorf("failed to check for session-overview.md: %w", err)
	}

	sb.WriteString("### Session Folder Contents:\n")
	if overviewExists {
		// Pointer-based approach: just reference the overview file
		sb.WriteString(fmt.Sprintf("- %s\n", overviewPath))
	} else {
		// Fallback to file enumeration for backward compatibility
		files, err := h.listSessionFiles(sessionPath)
		if err != nil {
			return "", fmt.Errorf("failed to list session files: %w", err)
		}

		if len(files) == 0 {
			sb.WriteString("(empty)\n")
		} else {
			for _, file := range files {
				sb.WriteString(fmt.Sprintf("- %s\n", file))
			}
		}
	}

	// Add activation procedure for documentation loading
	sb.WriteString("\n### ACTIVATION PROCEDURE (Execute on Session Start)\n\n")
	sb.WriteString("Before beginning any task work, execute this mandatory 3-step loading sequence:\n\n")

	sb.WriteString("**STEP 1: Load Session Context**\n")
	sb.WriteString(fmt.Sprintf("- Read `%s/session-overview.md` using the Read tool\n", sessionPath))

	sb.WriteString("**STEP 2: Load Root Doc Files**\n")
	sb.WriteString("- Read ALL files listed under \"Root Documentation Entry Points\" below\n")
	sb.WriteString("- Use Read tool for each file (do NOT use Glob/Grep for discovery)\n")

	sb.WriteString("**STEP 3: Recursive Index Traversal (Task-Driven)**\n")
	sb.WriteString("- Each doc file contains links to other doc files in subdirectories\n")
	sb.WriteString("- CRITICAL: Load only the files that are directly related and relevant to the task at hand\n")

	// Add doc paths as root entry points
	if len(docPaths) > 0 {
		sb.WriteString("**Root Documentation Entry Points:**\n")
		for _, docPath := range docPaths {
			if docPath != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", docPath))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// listSessionFiles returns markdown list of files in session folder
func (h *Handler) listSessionFiles(sessionPath string) ([]string, error) {
	// Read directory contents
	entries, err := afero.ReadDir(h.fs, sessionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session directory: %w", err)
	}

	// Collect file names (exclude directories)
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	// Sort alphabetically for consistent output
	sort.Strings(files)

	return files, nil
}

// hasIndexMdFiles checks if any index.md files exist in the project directory tree
func (h *Handler) hasIndexMdFiles(projectRoot string) bool {
	// Empty project root - graceful degradation
	if projectRoot == "" {
		return false
	}

	// Use afero.Walk to traverse directory tree
	found := false
	afero.Walk(h.fs, projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Continue walking even if we encounter errors
			return nil
		}

		// Check if this is an index.md file
		if !info.IsDir() && info.Name() == "index.md" {
			found = true
			// Early exit - we found one
			return filepath.SkipDir
		}

		return nil
	})

	return found
}

// buildExploreContext creates the Explore-specific context with MCP/LSP instructions
func (h *Handler) buildExploreContext() string {
	var sb strings.Builder

	sb.WriteString("## EXPLORE AGENT ENHANCEMENTS\n\n")
	sb.WriteString("You have access to powerful tools for codebase exploration. Use them effectively.\n\n")

	// LSP Instructions
	sb.WriteString("### LSP Tool (PREFERRED for code navigation)\n")
	sb.WriteString("Use LSP instead of brute-force Glob/Grep when possible:\n")
	sb.WriteString("- `goToDefinition`: Jump to where a symbol is defined\n")
	sb.WriteString("- `findReferences`: Find all usages of a symbol\n")
	sb.WriteString("- `hover`: Get documentation and type info for a symbol\n")
	sb.WriteString("- `documentSymbol`: List all symbols in a file\n")
	sb.WriteString("- `workspaceSymbol`: Search symbols across the codebase\n")
	sb.WriteString("- `incomingCalls`/`outgoingCalls`: Trace call hierarchy\n\n")
	sb.WriteString("**Parameters**: `operation`, `filePath` (absolute), `line`, `character`\n\n")

	// Context7 MCP Instructions
	sb.WriteString("### Context7 MCP (for library documentation)\n")
	sb.WriteString("Before making assumptions about libraries/frameworks, query current docs:\n")
	sb.WriteString("1. `mcp__context7__resolve-library-id`: Get library ID (e.g., \"redis\" → \"/redis/redis\")\n")
	sb.WriteString("2. `mcp__context7__query-docs`: Query specific documentation\n")
	sb.WriteString("**Constraint**: Max 3 calls per question\n\n")

	// Sequential Thinking MCP Instructions
	sb.WriteString("### Sequential Thinking MCP (for complex analysis)\n")
	sb.WriteString("Use `mcp__sequential-thinking__sequentialthinking` for:\n")
	sb.WriteString("- Multi-step problem solving\n")
	sb.WriteString("- Trade-off analysis\n")
	sb.WriteString("- Complex architectural decisions\n\n")

	// Best Practices
	sb.WriteString("### Exploration Best Practices\n")
	sb.WriteString("1. Start with LSP `workspaceSymbol` to find entry points\n")
	sb.WriteString("2. Use `goToDefinition` to trace implementations\n")
	sb.WriteString("3. Use `findReferences` to understand usage patterns\n")
	sb.WriteString("4. Fall back to Glob/Grep only for pattern-based searches\n")
	sb.WriteString("5. Cite findings with file:line format\n")

	return sb.String()
}

// buildPlanContext creates Plan-specific context with MCP tools and stack skills
func (h *Handler) buildPlanContext(stacks []string) string {
	var sb strings.Builder

	sb.WriteString("## PLAN AGENT ENHANCEMENTS\n\n")
	sb.WriteString("You are creating an execution plan. Use these tools and practices.\n\n")

	// MCP Tools (MANDATORY)
	sb.WriteString("### MCP Tools (MANDATORY)\n\n")
	sb.WriteString("**Context7 MCP** - Query documentation for all libraries/frameworks:\n")
	sb.WriteString("1. `mcp__context7__resolve-library-id`: Get library ID\n")
	sb.WriteString("2. `mcp__context7__query-docs`: Query specific documentation\n\n")
	sb.WriteString("**Sequential Thinking MCP** - Use for parallelization analysis:\n")
	sb.WriteString("- Component boundary identification\n")
	sb.WriteString("- Dependency mapping (what blocks what)\n")
	sb.WriteString("- Shared contract discovery\n")
	sb.WriteString("- Parallel opportunity grouping (Track A/B/C)\n")
	sb.WriteString("- Sequential constraint justification\n\n")

	// Execution Plan Structure
	sb.WriteString("### Execution Plan Structure\n\n")
	sb.WriteString("**Phase Labeling** (MANDATORY):\n")
	sb.WriteString("- `### Phase N: [Name] (Parallel: X independent tracks)`\n")
	sb.WriteString("- `### Phase N: [Name] (Sequential)` with justification\n\n")
	sb.WriteString("**Track Groupings** for parallel phases:\n")
	sb.WriteString("```\n")
	sb.WriteString("Track A: [task1, task2]\n")
	sb.WriteString("Track B: [task3, task4]\n")
	sb.WriteString("```\n\n")
	sb.WriteString("**Architect Boundaries**:\n")
	sb.WriteString("- Define WHAT to build and HOW to approach it\n")
	sb.WriteString("- Code snippets: Max 15 lines for patterns, NOT full implementations\n")
	sb.WriteString("- Use file:line pointers when referencing existing code\n\n")

	// Inject stack-specific skills
	if len(stacks) > 0 {
		sb.WriteString("### Detected Tech Stack Skills\n\n")
		for _, stack := range stacks {
			skillContent := h.loadSkillContent(stack)
			if skillContent != "" {
				sb.WriteString(fmt.Sprintf("#### %s\n\n", strings.Title(stack)))
				sb.WriteString(skillContent)
				sb.WriteString("\n\n")
			}
		}
	}

	return sb.String()
}

// loadSkillContent reads skill file from embedded profiles
func (h *Handler) loadSkillContent(stack string) string {
	skillPath := fmt.Sprintf("profiles/skills/%s.md", stack)
	content, err := fs.ReadFile(claudex.Profiles, skillPath)
	if err != nil {
		if h.logger != nil {
			_ = h.logger.Logf("Could not load skill for %s: %v", stack, err)
		}
		return ""
	}
	return string(content)
}

// HandleFromBuilder is a convenience wrapper that returns the built output
// This is useful for command-line integration
func (h *Handler) HandleFromBuilder(input *shared.PreToolUseInput, builder *shared.Builder) error {
	output, err := h.Handle(input)
	if err != nil {
		return err
	}
	return builder.BuildCustom(*output)
}
