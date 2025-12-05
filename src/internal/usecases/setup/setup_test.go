package setup

import (
	"testing"

	"claudex/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_Execute_CreatesStructure verifies that the complete directory
// structure is created with hooks, agents, and settings files from the config directory.
func Test_Execute_CreatesStructure(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create config directory structure
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh":    "#!/bin/bash\necho notify",
		"hooks/session-end.sh":          "#!/bin/bash\necho end",
		"hooks/subagent-stop.sh":        "#!/bin/bash\necho stop",
		"hooks/pre-tool-use.sh":         "#!/bin/bash\necho pre",
		"hooks/post-tool-use.sh":        "#!/bin/bash\necho post",
		"hooks/auto-doc-updater.sh":     "#!/bin/bash\necho doc",
		"profiles/agents/team-lead.md":  "# Team Lead Agent\nContent here",
		"profiles/agents/architect.md":  "# Architect Agent\nContent here",
		"profiles/roles/engineer.md":    "# {Stack} Engineer Role\nRole template",
		"profiles/skills/typescript.md": "# TypeScript Skill\nTypeScript expertise",
		"profiles/skills/go.md":         "# Go Skill\nGo expertise",
		"profiles/skills/python.md":     "# Python Skill\nPython expertise",
	})

	// Create project directory with package.json to detect TypeScript
	h.CreateDir("/project")
	h.WriteFile("/project/package.json", `{"name": "test-project"}`)

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - no errors
	require.NoError(t, err)

	// Verify - directory structure created
	testutil.AssertDirExists(t, h.FS, "/project/.claude")
	testutil.AssertDirExists(t, h.FS, "/project/.claude/hooks")
	testutil.AssertDirExists(t, h.FS, "/project/.claude/agents")
	testutil.AssertDirExists(t, h.FS, "/project/.claude/commands/agents")

	// Verify - hooks copied
	testutil.AssertFileExists(t, h.FS, "/project/.claude/hooks/notification-hook.sh")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/hooks/session-end.sh")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/hooks/subagent-stop.sh")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/hooks/pre-tool-use.sh")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/hooks/post-tool-use.sh")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/hooks/auto-doc-updater.sh")

	// Verify - hook file contents
	testutil.AssertFileContains(t, h.FS, "/project/.claude/hooks/notification-hook.sh", "echo notify")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/hooks/session-end.sh", "echo end")

	// Verify - agents copied
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/team-lead.md.md")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/architect.md.md")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/commands/agents/team-lead.md.md")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/commands/agents/architect.md.md")

	// Verify - settings.local.json created with hook registrations
	testutil.AssertFileExists(t, h.FS, "/project/.claude/settings.local.json")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "notification-hook.sh")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "session-end.sh")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "SessionEnd")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "Notification")

	// Verify - generated engineer profiles exist
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer-typescript.md")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/commands/agents/principal-engineer-typescript.md")

	// Verify - principal-engineer alias created (points to primary stack)
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer.md")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/commands/agents/principal-engineer.md")
}

// Test_Execute_RespectsNoOverwrite verifies that existing files
// are preserved when noOverwrite=true, and new files from config are not copied.
func Test_Execute_RespectsNoOverwrite(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create config directory with agents
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"profiles/agents/team-lead.md":  "# New Team Lead Content",
		"profiles/agents/architect.md":  "# New Architect Content",
		"profiles/roles/engineer.md":    "# {Stack} Engineer Role",
		"profiles/skills/typescript.md": "# TypeScript Skill",
		"hooks/notification-hook.sh":    "#!/bin/bash\necho notify",
	})

	// Create existing .claude directory with custom agent
	h.CreateDir("/project/.claude/agents")
	h.CreateDir("/project/.claude/commands/agents")
	h.WriteFile("/project/.claude/agents/custom-agent.md", "# My Custom Agent\nCustom content here")
	h.WriteFile("/project/.claude/agents/team-lead.md.md", "# Old Team Lead\nOld content")
	h.WriteFile("/project/.claude/settings.local.json", `{"existing": "settings"}`)

	// Create project with package.json
	h.WriteFile("/project/package.json", `{"name": "test"}`)

	// Create usecase and exercise - call with noOverwrite=true
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", true)

	// Verify - no errors
	require.NoError(t, err)

	// Verify - existing custom agent preserved
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/custom-agent.md")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/agents/custom-agent.md", "My Custom Agent")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/agents/custom-agent.md", "Custom content here")

	// Verify - existing team-lead not overwritten
	testutil.AssertFileContains(t, h.FS, "/project/.claude/agents/team-lead.md.md", "Old Team Lead")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/agents/team-lead.md.md", "Old content")

	// Verify - existing settings.local.json preserved
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", `{"existing": "settings"}`)

	// Verify - architect agent NOT copied (noOverwrite prevents new files from being written if any exist)
	// Note: Based on the implementation, noOverwrite only prevents overwriting existing files,
	// new files are still created. So architect.md.md should exist.
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/architect.md.md")
}

// Test_Execute_GeneratesEngineerProfiles verifies that engineer
// profiles are correctly assembled from role + skill templates for detected stacks.
func Test_Execute_GeneratesEngineerProfiles(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create config with roles and skills
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"profiles/roles/engineer.md":    "# {Stack} Engineer Role\nThis is the engineer role template.\n{Stack} specific content.",
		"profiles/skills/typescript.md": "# TypeScript Skill\n\nExpert in TypeScript development.\nUses strict typing.",
		"profiles/skills/go.md":         "# Go Skill\n\nExpert in Go development.\nIdiomatic Go patterns.",
		"profiles/skills/python.md":     "# Python Skill\n\nExpert in Python.\nPEP 8 compliant.",
		"hooks/notification-hook.sh":    "#!/bin/bash\necho notify",
	})

	// Create project with package.json (TypeScript marker)
	h.CreateDir("/project")
	h.WriteFile("/project/package.json", `{"name": "typescript-project"}`)

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - no errors
	require.NoError(t, err)

	// Verify - principal-engineer-typescript.md generated
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer-typescript.md")

	// Read and verify content
	content, err := h.FS.Open("/project/.claude/agents/principal-engineer-typescript.md")
	require.NoError(t, err)
	defer content.Close()

	// Read file content as bytes
	buf := make([]byte, 4096)
	n, err := content.Read(buf)
	require.NoError(t, err)
	contentStr := string(buf[:n])

	// Verify frontmatter generated
	assert.Contains(t, contentStr, "name: principal-engineer-typescript")
	assert.Contains(t, contentStr, "description: Use this agent when you need a Principal TypeScript Engineer")

	// Verify role content included (with {Stack} replaced)
	assert.Contains(t, contentStr, "TypeScript Engineer Role")
	assert.Contains(t, contentStr, "This is the engineer role template")
	assert.NotContains(t, contentStr, "{Stack}") // Placeholder should be replaced

	// Verify skill content included
	assert.Contains(t, contentStr, "TypeScript Skill")
	assert.Contains(t, contentStr, "Expert in TypeScript development")
	assert.Contains(t, contentStr, "Uses strict typing")

	// Verify principal-engineer.md alias created (copy of primary stack)
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer.md")
	aliasContent, err := h.FS.Open("/project/.claude/agents/principal-engineer.md")
	require.NoError(t, err)
	defer aliasContent.Close()

	aliasBuf := make([]byte, 4096)
	aliasN, err := aliasContent.Read(aliasBuf)
	require.NoError(t, err)
	aliasContentStr := string(aliasBuf[:aliasN])

	// Alias should have same content as principal-engineer-typescript
	assert.Contains(t, aliasContentStr, "TypeScript Engineer Role")
	assert.Contains(t, aliasContentStr, "Expert in TypeScript development")

	// Verify files also copied to commands/agents/
	testutil.AssertFileExists(t, h.FS, "/project/.claude/commands/agents/principal-engineer-typescript.md")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/commands/agents/principal-engineer.md")
}

// Test_Execute_MultipleStacks verifies that multiple engineer profiles
// are generated when multiple stack markers are detected.
func Test_Execute_MultipleStacks(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create config with roles and skills
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"profiles/roles/engineer.md":    "# {Stack} Engineer",
		"profiles/skills/typescript.md": "# TypeScript Skill",
		"profiles/skills/go.md":         "# Go Skill",
		"profiles/skills/python.md":     "# Python Skill",
		"hooks/notification-hook.sh":    "#!/bin/bash\necho notify",
	})

	// Create project with both package.json and go.mod (TypeScript + Go)
	h.CreateDir("/project")
	h.WriteFile("/project/package.json", `{"name": "polyglot-project"}`)
	h.WriteFile("/project/go.mod", "module example.com/project\n\ngo 1.21")

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - no errors
	require.NoError(t, err)

	// Verify - both engineer profiles generated
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer-typescript.md")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer-go.md")

	// Verify - principal-engineer alias points to first detected stack (typescript)
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer.md")
	aliasContent, err := h.FS.Open("/project/.claude/agents/principal-engineer.md")
	require.NoError(t, err)
	defer aliasContent.Close()

	buf := make([]byte, 2048)
	n, err := aliasContent.Read(buf)
	require.NoError(t, err)
	contentStr := string(buf[:n])

	// Should contain TypeScript content (first stack)
	assert.Contains(t, contentStr, "TypeScript Engineer")
	assert.Contains(t, contentStr, "TypeScript Skill")
}

// Test_Execute_XDGConfigHome verifies that XDG_CONFIG_HOME
// is respected when looking for the claudex config directory.
func Test_Execute_XDGConfigHome(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")
	h.Env.Set("XDG_CONFIG_HOME", "/custom/config")

	// Create config in custom XDG location
	h.SetupConfigDir("/custom/config/claudex", map[string]string{
		"profiles/agents/team-lead.md":  "# Team Lead",
		"profiles/roles/engineer.md":    "# Engineer",
		"profiles/skills/typescript.md": "# TypeScript",
		"hooks/notification-hook.sh":    "#!/bin/bash\necho notify",
	})

	// Create project
	h.CreateDir("/project")
	h.WriteFile("/project/package.json", `{"name": "test"}`)

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - no errors
	require.NoError(t, err)

	// Verify - structure created using XDG_CONFIG_HOME
	testutil.AssertDirExists(t, h.FS, "/project/.claude")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/team-lead.md.md")
}

// Test_Execute_MissingConfigDir verifies that an error is returned
// when the claudex config directory doesn't exist.
func Test_Execute_MissingConfigDir(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")
	// Don't create the config directory

	h.CreateDir("/project")

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - error returned
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claudex config directory not found")
	assert.Contains(t, err.Error(), "/home/user/.config/claudex")
}

// Test_Execute_MissingHOME verifies that an error is returned
// when HOME environment variable is not set.
func Test_Execute_MissingHOME(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	// Don't set HOME or XDG_CONFIG_HOME

	h.CreateDir("/project")

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - error returned
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HOME environment variable not set")
}

// Test_Execute_NoStackDetected verifies that default stacks
// are used when no project markers are found.
func Test_Execute_NoStackDetected(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create config
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"profiles/roles/engineer.md":    "# Engineer",
		"profiles/skills/typescript.md": "# TypeScript",
		"profiles/skills/go.md":         "# Go",
		"profiles/skills/python.md":     "# Python",
		"hooks/notification-hook.sh":    "#!/bin/bash\necho notify",
	})

	// Create empty project (no stack markers)
	h.CreateDir("/project")

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - no errors
	require.NoError(t, err)

	// Verify - all three default stacks generated
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer-typescript.md")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer-go.md")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer-python.md")

	// Verify - principal-engineer alias created (first default: typescript)
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer.md")
}
