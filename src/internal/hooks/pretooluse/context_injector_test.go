package pretooluse

import (
	"strings"
	"testing"

	"claudex/internal/hooks/shared"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_NonTaskTool(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()
	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "test-session-123",
		},
		ToolName: "Read",
		ToolInput: map[string]interface{}{
			"file_path": "/some/path/file.txt",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	assert.Nil(t, output.HookSpecificOutput.UpdatedInput)
}

func TestHandler_NoSessionFolder(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()
	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "nonexistent-session",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "Do some work",
			"description":   "Task description",
			"subagent_type": "researcher",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	assert.Nil(t, output.HookSpecificOutput.UpdatedInput)
}

func TestHandler_EmptyPrompt(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "",
			"description":   "Task description",
			"subagent_type": "researcher",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	assert.Nil(t, output.HookSpecificOutput.UpdatedInput)
}

func TestHandler_MissingPrompt(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"description":   "Task description",
			"subagent_type": "researcher",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	assert.Nil(t, output.HookSpecificOutput.UpdatedInput)
}

func TestHandler_SuccessfulInjection(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder with files
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create some files in the session folder
	afero.WriteFile(fs, sessionPath+"/research-topic.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/execution-plan.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/notes.txt", []byte("content"), 0644)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	originalPrompt := "Please analyze the codebase"
	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        originalPrompt,
			"description":   "Task description",
			"subagent_type": "researcher",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	// Verify updated input contains all original fields
	assert.Equal(t, "Task description", output.HookSpecificOutput.UpdatedInput["description"])
	assert.Equal(t, "researcher", output.HookSpecificOutput.UpdatedInput["subagent_type"])

	// Verify prompt was modified
	modifiedPrompt, ok := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	require.True(t, ok)
	assert.Contains(t, modifiedPrompt, "## SESSION CONTEXT (CRITICAL)")
	assert.Contains(t, modifiedPrompt, sessionPath)
	assert.Contains(t, modifiedPrompt, "MANDATORY RULES")
	assert.Contains(t, modifiedPrompt, "## ORIGINAL REQUEST")
	assert.Contains(t, modifiedPrompt, originalPrompt)

	// Verify session files are listed
	assert.Contains(t, modifiedPrompt, "- execution-plan.md")
	assert.Contains(t, modifiedPrompt, "- notes.txt")
	assert.Contains(t, modifiedPrompt, "- research-topic.md")
}

func TestHandler_EmptySessionFolder(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create empty session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	originalPrompt := "Please analyze the codebase"
	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt": originalPrompt,
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	assert.Contains(t, modifiedPrompt, "### Session Folder Contents:")
	assert.Contains(t, modifiedPrompt, "(empty)")
}

func TestHandler_WithDocPaths(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)
	env.Set("CLAUDEX_DOC_PATHS", "research-{topic}.md:execution-plan-{feature}.md:analysis-{component}.md")

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	originalPrompt := "Please analyze the codebase"
	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt": originalPrompt,
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	assert.Contains(t, modifiedPrompt, "**Root Documentation Entry Points:**")
	assert.Contains(t, modifiedPrompt, "- research-{topic}.md")
	assert.Contains(t, modifiedPrompt, "- execution-plan-{feature}.md")
	assert.Contains(t, modifiedPrompt, "- analysis-{component}.md")
}

func TestHandler_PatternMatchSessionFolder(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder with pattern-based name
	sessionPath := "./.claudex/sessions/golang-hooks-rewrite-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create a file in the session
	afero.WriteFile(fs, sessionPath+"/research.md", []byte("content"), 0644)

	// Don't set CLAUDEX_SESSION_PATH - let it use pattern matching

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	originalPrompt := "Please analyze the codebase"
	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt": originalPrompt,
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	assert.Contains(t, modifiedPrompt, "## SESSION CONTEXT (CRITICAL)")
	assert.Contains(t, modifiedPrompt, "- research.md")
}

func TestHandler_PreservesAllToolInputFields(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "Original prompt",
			"description":   "Task description",
			"subagent_type": "researcher",
			"custom_field":  "custom_value",
			"numeric_field": 42,
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	// Verify all fields are preserved
	assert.Equal(t, "Task description", output.HookSpecificOutput.UpdatedInput["description"])
	assert.Equal(t, "researcher", output.HookSpecificOutput.UpdatedInput["subagent_type"])
	assert.Equal(t, "custom_value", output.HookSpecificOutput.UpdatedInput["custom_field"])
	assert.Equal(t, 42, output.HookSpecificOutput.UpdatedInput["numeric_field"])

	// Verify only prompt is modified
	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	assert.Contains(t, modifiedPrompt, "## SESSION CONTEXT")
	assert.Contains(t, modifiedPrompt, "Original prompt")
}

func TestHandler_SessionFolderWithDirectories(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder with both files and directories
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath+"/subdir", 0755)
	require.NoError(t, err)

	// Create files
	afero.WriteFile(fs, sessionPath+"/file1.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/file2.txt", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/subdir/nested.md", []byte("content"), 0644)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt": "Test prompt",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	// Should list files but not directories
	assert.Contains(t, modifiedPrompt, "- file1.md")
	assert.Contains(t, modifiedPrompt, "- file2.txt")
	assert.NotContains(t, modifiedPrompt, "- subdir")
	assert.NotContains(t, modifiedPrompt, "- nested.md") // Should not list files in subdirs
}

func TestBuildSessionContext(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create files in alphabetical order to test sorting
	afero.WriteFile(fs, sessionPath+"/zebra.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/apple.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/banana.md", []byte("content"), 0644)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	docPaths := []string{"research-{topic}.md", "execution-plan-{feature}.md"}

	// Act
	context, err := handler.buildSessionContext(sessionPath, docPaths, "")

	// Assert
	require.NoError(t, err)
	assert.Contains(t, context, "## SESSION CONTEXT (CRITICAL)")
	assert.Contains(t, context, sessionPath)
	assert.Contains(t, context, "MANDATORY RULES")
	assert.Contains(t, context, "### Session Folder Contents:")

	// Verify files are sorted alphabetically
	appleIdx := strings.Index(context, "- apple.md")
	bananaIdx := strings.Index(context, "- banana.md")
	zebraIdx := strings.Index(context, "- zebra.md")
	assert.True(t, appleIdx < bananaIdx)
	assert.True(t, bananaIdx < zebraIdx)

	// Verify doc paths
	assert.Contains(t, context, "**Root Documentation Entry Points:**")
	assert.Contains(t, context, "- research-{topic}.md")
	assert.Contains(t, context, "- execution-plan-{feature}.md")
}

func TestListSessionFiles(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	sessionPath := "/workspace/.claudex/sessions/test-session"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create mixed files and directories
	afero.WriteFile(fs, sessionPath+"/file1.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/file2.txt", []byte("content"), 0644)
	fs.MkdirAll(sessionPath+"/subdir", 0755)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	// Act
	files, err := handler.listSessionFiles(sessionPath)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, len(files))
	assert.Contains(t, files, "file1.md")
	assert.Contains(t, files, "file2.txt")
	assert.NotContains(t, files, "subdir")
}

func TestListSessionFiles_NonexistentDirectory(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	// Act
	files, err := handler.listSessionFiles("/nonexistent/path")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, files)
}

func TestBuildSessionContext_WithOverview(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	sessionPath := "/workspace/.claudex/sessions/test-session"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create session-overview.md and other files
	afero.WriteFile(fs, sessionPath+"/session-overview.md", []byte("overview content"), 0644)
	afero.WriteFile(fs, sessionPath+"/other-file.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/another-file.md", []byte("content"), 0644)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	// Act
	context, err := handler.buildSessionContext(sessionPath, nil, "")

	// Assert
	require.NoError(t, err)
	assert.Contains(t, context, "### Session Folder Contents:")
	// Should contain absolute path to session-overview.md
	assert.Contains(t, context, sessionPath+"/session-overview.md")
	// Should NOT contain other file names (pointer-based approach)
	assert.NotContains(t, context, "- other-file.md")
	assert.NotContains(t, context, "- another-file.md")
}

func TestBuildSessionContext_WithoutOverview(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	sessionPath := "/workspace/.claudex/sessions/test-session"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create files WITHOUT session-overview.md
	afero.WriteFile(fs, sessionPath+"/file1.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/file2.md", []byte("content"), 0644)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	// Act
	context, err := handler.buildSessionContext(sessionPath, nil, "")

	// Assert
	require.NoError(t, err)
	assert.Contains(t, context, "### Session Folder Contents:")
	// Should fallback to file enumeration (filenames only, not full paths)
	assert.Contains(t, context, "- file1.md")
	assert.Contains(t, context, "- file2.md")
	// Should NOT contain absolute paths (fallback mode)
	assert.NotContains(t, context, sessionPath+"/file1.md")
}

func TestBuildSessionContext_WithIndexMdHint(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	sessionPath := "/workspace/.claudex/sessions/test-session"
	projectRoot := "/workspace/project"

	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create project structure with nested index.md
	err = fs.MkdirAll(projectRoot+"/src/internal", 0755)
	require.NoError(t, err)
	afero.WriteFile(fs, projectRoot+"/src/internal/index.md", []byte("index content"), 0644)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	// Act
	context, err := handler.buildSessionContext(sessionPath, nil, projectRoot)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, context, "### ACTIVATION PROCEDURE (Execute on Session Start)")
	assert.Contains(t, context, "**STEP 1: Load Session Context**")
	assert.Contains(t, context, "**STEP 2: Load Root Doc Files**")
	assert.Contains(t, context, "**STEP 3: Recursive Index Traversal (Task-Driven)**")
}

func TestBuildSessionContext_NoIndexMdHint(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	sessionPath := "/workspace/.claudex/sessions/test-session"
	projectRoot := "/workspace/project"

	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create project structure WITHOUT any index.md files
	err = fs.MkdirAll(projectRoot+"/src/internal", 0755)
	require.NoError(t, err)
	afero.WriteFile(fs, projectRoot+"/src/main.go", []byte("code"), 0644)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	// Act
	context, err := handler.buildSessionContext(sessionPath, nil, projectRoot)

	// Assert
	require.NoError(t, err)
	assert.NotContains(t, context, "### Codebase Navigation:")
	assert.NotContains(t, context, "index.md files")
}

func TestHasIndexMdFiles_Found(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	projectRoot := "/workspace/project"
	err := fs.MkdirAll(projectRoot+"/src/internal/hooks", 0755)
	require.NoError(t, err)

	// Create nested index.md file
	afero.WriteFile(fs, projectRoot+"/src/internal/hooks/index.md", []byte("content"), 0644)
	afero.WriteFile(fs, projectRoot+"/src/main.go", []byte("code"), 0644)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	// Act
	found := handler.hasIndexMdFiles(projectRoot)

	// Assert
	assert.True(t, found)
}

func TestHasIndexMdFiles_NotFound(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	projectRoot := "/workspace/project"
	err := fs.MkdirAll(projectRoot+"/src/internal", 0755)
	require.NoError(t, err)

	// Create files but NO index.md
	afero.WriteFile(fs, projectRoot+"/src/main.go", []byte("code"), 0644)
	afero.WriteFile(fs, projectRoot+"/README.md", []byte("readme"), 0644)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	// Act
	found := handler.hasIndexMdFiles(projectRoot)

	// Assert
	assert.False(t, found)
}

func TestHasIndexMdFiles_EmptyProjectRoot(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	// Act
	found := handler.hasIndexMdFiles("")

	// Assert
	assert.False(t, found, "Empty project root should return false for graceful degradation")
}

func TestHandler_ExploreAgent_InjectsMCPLSPContext(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	originalPrompt := "Explore the authentication flow in the codebase"
	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        originalPrompt,
			"description":   "Exploration task",
			"subagent_type": "Explore",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	// Verify prompt was modified with Explore-specific context
	modifiedPrompt, ok := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	require.True(t, ok)

	// Should contain Explore-specific instructions
	assert.Contains(t, modifiedPrompt, "## EXPLORE AGENT ENHANCEMENTS")
	assert.Contains(t, modifiedPrompt, "### LSP Tool (PREFERRED for code navigation)")
	assert.Contains(t, modifiedPrompt, "goToDefinition")
	assert.Contains(t, modifiedPrompt, "findReferences")
	assert.Contains(t, modifiedPrompt, "workspaceSymbol")
	assert.Contains(t, modifiedPrompt, "### Context7 MCP (for library documentation)")
	assert.Contains(t, modifiedPrompt, "mcp__context7__resolve-library-id")
	assert.Contains(t, modifiedPrompt, "mcp__context7__query-docs")
	assert.Contains(t, modifiedPrompt, "### Sequential Thinking MCP (for complex analysis)")
	assert.Contains(t, modifiedPrompt, "mcp__sequential-thinking__sequentialthinking")
	assert.Contains(t, modifiedPrompt, "### Exploration Best Practices")

	// Should contain original request
	assert.Contains(t, modifiedPrompt, "## ORIGINAL REQUEST")
	assert.Contains(t, modifiedPrompt, originalPrompt)

	// Verify all original fields are preserved
	assert.Equal(t, "Exploration task", output.HookSpecificOutput.UpdatedInput["description"])
	assert.Equal(t, "Explore", output.HookSpecificOutput.UpdatedInput["subagent_type"])
}

func TestHandler_ExploreAgent_CaseInsensitive(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	testCases := []struct {
		name          string
		subagentType  string
		shouldTrigger bool
	}{
		{"lowercase", "explore", true},
		{"uppercase", "EXPLORE", true},
		{"mixed case", "ExPlOrE", true},
		{"proper case", "Explore", true},
		{"researcher", "researcher", false},
		{"architect", "Architect", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := &shared.PreToolUseInput{
				HookInput: shared.HookInput{
					SessionID: "abc123",
				},
				ToolName: "Task",
				ToolInput: map[string]interface{}{
					"prompt":        "Test prompt",
					"subagent_type": tc.subagentType,
				},
			}

			// Act
			output, err := handler.Handle(input)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

			modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)

			if tc.shouldTrigger {
				assert.Contains(t, modifiedPrompt, "## EXPLORE AGENT ENHANCEMENTS",
					"Expected Explore context for subagent_type: %s", tc.subagentType)
			} else {
				assert.NotContains(t, modifiedPrompt, "## EXPLORE AGENT ENHANCEMENTS",
					"Did not expect Explore context for subagent_type: %s", tc.subagentType)
			}
		})
	}
}

func TestHandler_ExploreAgent_NoSessionContext(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder with files
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create session files that would normally be listed
	afero.WriteFile(fs, sessionPath+"/research.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/plan.md", []byte("content"), 0644)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "Explore the codebase",
			"subagent_type": "Explore",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)

	// Should NOT contain session context markers
	assert.NotContains(t, modifiedPrompt, "## SESSION CONTEXT (CRITICAL)")
	assert.NotContains(t, modifiedPrompt, "MANDATORY RULES")
	assert.NotContains(t, modifiedPrompt, "### Session Folder Contents:")
	assert.NotContains(t, modifiedPrompt, sessionPath)

	// Should contain Explore-specific context
	assert.Contains(t, modifiedPrompt, "## EXPLORE AGENT ENHANCEMENTS")
}

func TestHandler_NonExploreAgent_StillGetsSessionContext(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder with files
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	afero.WriteFile(fs, sessionPath+"/research.md", []byte("content"), 0644)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "Research the authentication pattern",
			"subagent_type": "researcher",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)

	// Should contain session context (not Explore context)
	assert.Contains(t, modifiedPrompt, "## SESSION CONTEXT (CRITICAL)")
	assert.Contains(t, modifiedPrompt, "MANDATORY RULES")
	assert.Contains(t, modifiedPrompt, sessionPath)

	// Should NOT contain Explore-specific context
	assert.NotContains(t, modifiedPrompt, "## EXPLORE AGENT ENHANCEMENTS")
	assert.NotContains(t, modifiedPrompt, "### LSP Tool (PREFERRED for code navigation)")
}

func TestHandler_PlanAgent_InjectsPlanContext(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	originalPrompt := "Create an execution plan for adding authentication"
	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
			CWD:       "/workspace/project",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        originalPrompt,
			"description":   "Planning task",
			"subagent_type": "Plan",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	// Verify prompt was modified with Plan-specific context
	modifiedPrompt, ok := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	require.True(t, ok)

	// Should contain Plan-specific instructions
	assert.Contains(t, modifiedPrompt, "## PLAN AGENT ENHANCEMENTS")
	assert.Contains(t, modifiedPrompt, "### MCP Tools (MANDATORY)")
	assert.Contains(t, modifiedPrompt, "**Context7 MCP**")
	assert.Contains(t, modifiedPrompt, "mcp__context7__resolve-library-id")
	assert.Contains(t, modifiedPrompt, "mcp__context7__query-docs")
	assert.Contains(t, modifiedPrompt, "**Sequential Thinking MCP**")
	assert.Contains(t, modifiedPrompt, "Component boundary identification")
	assert.Contains(t, modifiedPrompt, "parallelization analysis")
	assert.Contains(t, modifiedPrompt, "### Execution Plan Structure")
	assert.Contains(t, modifiedPrompt, "**Phase Labeling**")
	assert.Contains(t, modifiedPrompt, "**Track Groupings**")
	assert.Contains(t, modifiedPrompt, "**Architect Boundaries**")

	// Should contain original request
	assert.Contains(t, modifiedPrompt, "## ORIGINAL REQUEST")
	assert.Contains(t, modifiedPrompt, originalPrompt)

	// Verify all original fields are preserved
	assert.Equal(t, "Planning task", output.HookSpecificOutput.UpdatedInput["description"])
	assert.Equal(t, "Plan", output.HookSpecificOutput.UpdatedInput["subagent_type"])
}

func TestHandler_PlanAgent_CaseInsensitive(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	testCases := []struct {
		name          string
		subagentType  string
		shouldTrigger bool
	}{
		{"lowercase", "plan", true},
		{"uppercase", "PLAN", true},
		{"mixed case", "PlAn", true},
		{"proper case", "Plan", true},
		{"researcher", "researcher", false},
		{"architect", "Architect", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := &shared.PreToolUseInput{
				HookInput: shared.HookInput{
					SessionID: "abc123",
					CWD:       "/workspace/project",
				},
				ToolName: "Task",
				ToolInput: map[string]interface{}{
					"prompt":        "Test prompt",
					"subagent_type": tc.subagentType,
				},
			}

			// Act
			output, err := handler.Handle(input)

			// Assert
			require.NoError(t, err)
			require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

			modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)

			if tc.shouldTrigger {
				assert.Contains(t, modifiedPrompt, "## PLAN AGENT ENHANCEMENTS",
					"Expected Plan context for subagent_type: %s", tc.subagentType)
			} else {
				assert.NotContains(t, modifiedPrompt, "## PLAN AGENT ENHANCEMENTS",
					"Did not expect Plan context for subagent_type: %s", tc.subagentType)
			}
		})
	}
}

func TestHandler_PlanAgent_DetectsGoStack(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create project with go.mod
	projectPath := "/workspace/project"
	err = fs.MkdirAll(projectPath, 0755)
	require.NoError(t, err)
	afero.WriteFile(fs, projectPath+"/go.mod", []byte("module test"), 0644)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
			CWD:       projectPath,
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "Create execution plan",
			"subagent_type": "Plan",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)

	// Should contain Plan context
	assert.Contains(t, modifiedPrompt, "## PLAN AGENT ENHANCEMENTS")

	// Should contain detected Go stack skills
	assert.Contains(t, modifiedPrompt, "### Detected Tech Stack Skills")
	assert.Contains(t, modifiedPrompt, "#### Go")
}

func TestHandler_PlanAgent_DetectsTypeScriptStack(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create project with package.json
	projectPath := "/workspace/project"
	err = fs.MkdirAll(projectPath, 0755)
	require.NoError(t, err)
	afero.WriteFile(fs, projectPath+"/package.json", []byte("{}"), 0644)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
			CWD:       projectPath,
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "Create execution plan",
			"subagent_type": "Plan",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)

	// Should contain Plan context
	assert.Contains(t, modifiedPrompt, "## PLAN AGENT ENHANCEMENTS")

	// Should contain detected TypeScript stack skills
	assert.Contains(t, modifiedPrompt, "### Detected Tech Stack Skills")
	assert.Contains(t, modifiedPrompt, "#### Typescript")
}

func TestHandler_PlanAgent_MultipleStacks(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create project with both go.mod and package.json
	projectPath := "/workspace/project"
	err = fs.MkdirAll(projectPath, 0755)
	require.NoError(t, err)
	afero.WriteFile(fs, projectPath+"/go.mod", []byte("module test"), 0644)
	afero.WriteFile(fs, projectPath+"/package.json", []byte("{}"), 0644)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
			CWD:       projectPath,
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "Create execution plan",
			"subagent_type": "Plan",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)

	// Should contain Plan context
	assert.Contains(t, modifiedPrompt, "## PLAN AGENT ENHANCEMENTS")

	// Should contain detected skills for both stacks
	assert.Contains(t, modifiedPrompt, "### Detected Tech Stack Skills")
	assert.Contains(t, modifiedPrompt, "#### Typescript")
	assert.Contains(t, modifiedPrompt, "#### Go")
}

func TestHandler_PlanAgent_NoSessionContext(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder with files
	sessionPath := "/workspace/.claudex/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create session files that would normally be listed
	afero.WriteFile(fs, sessionPath+"/research.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/plan.md", []byte("content"), 0644)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
			CWD:       "/workspace/project",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "Create execution plan",
			"subagent_type": "Plan",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)

	// Should NOT contain session context markers
	assert.NotContains(t, modifiedPrompt, "## SESSION CONTEXT (CRITICAL)")
	assert.NotContains(t, modifiedPrompt, "MANDATORY RULES")
	assert.NotContains(t, modifiedPrompt, "### Session Folder Contents:")
	assert.NotContains(t, modifiedPrompt, sessionPath)

	// Should contain Plan-specific context
	assert.Contains(t, modifiedPrompt, "## PLAN AGENT ENHANCEMENTS")
}
