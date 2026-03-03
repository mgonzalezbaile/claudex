package setup

import (
	"strings"
	"testing"

	"claudex/internal/testutil"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_Execute_CreatesStructure verifies that the complete directory
// structure is created with hooks, agents, and settings files from the config directory.
func Test_Execute_CreatesStructure(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create config directory with optional hooks (profiles are now embedded)
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
		"hooks/session-end.sh":       "#!/bin/bash\necho end",
		"hooks/subagent-stop.sh":     "#!/bin/bash\necho stop",
		"hooks/pre-tool-use.sh":      "#!/bin/bash\necho pre",
		"hooks/post-tool-use.sh":     "#!/bin/bash\necho post",
		"hooks/auto-doc-updater.sh":  "#!/bin/bash\necho doc",
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

	// Verify - agents copied from embedded profiles
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/team-lead.md")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/commands/agents/team-lead.md")

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
// Settings are always merged to add missing hooks while preserving customizations.
func Test_Execute_RespectsNoOverwrite(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create config directory with optional hooks (profiles are now embedded)
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
	})

	// Create existing .claude directory with custom agent and settings
	h.CreateDir("/project/.claude/agents")
	h.CreateDir("/project/.claude/commands/agents")
	h.WriteFile("/project/.claude/agents/custom-agent.md", "# My Custom Agent\nCustom content here")
	h.WriteFile("/project/.claude/agents/team-lead.md", "# Old Team Lead\nOld content")

	// Create existing settings with custom permissions and a custom hook
	existingSettings := `{
  "permissions": {
    "allow": ["Bash(custom:*)"],
    "deny": ["Write"],
    "ask": []
  },
  "hooks": {
    "PostToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/my-custom-hook.sh"
          }
        ]
      }
    ]
  }
}`
	h.WriteFile("/project/.claude/settings.local.json", existingSettings)

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
	testutil.AssertFileContains(t, h.FS, "/project/.claude/agents/team-lead.md", "Old Team Lead")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/agents/team-lead.md", "Old content")

	// Verify - settings merged: custom permissions preserved
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", `"allow": [`)
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", `"Bash(custom:*)"`)
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", `"deny": [`)
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", `"Write"`)

	// Verify - settings merged: custom hook preserved
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "my-custom-hook.sh")

	// Verify - settings merged: template hooks added
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "notification-hook.sh")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "session-end.sh")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "Notification")
}

// Test_Execute_GeneratesEngineerProfiles verifies that engineer
// profiles are correctly assembled from role + skill templates for detected stacks.
func Test_Execute_GeneratesEngineerProfiles(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create optional hooks config (roles and skills are now embedded)
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
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

	// Read file content as bytes (need larger buffer for full content)
	buf := make([]byte, 16384)
	n, err := content.Read(buf)
	require.NoError(t, err)
	contentStr := string(buf[:n])

	// Verify frontmatter generated
	assert.Contains(t, contentStr, "name: principal-engineer-typescript")
	assert.Contains(t, contentStr, "description: Use this agent when you need a Principal TypeScript Engineer")

	// Verify role content included from embedded profiles
	assert.Contains(t, contentStr, "Principal Software Engineer specializing in TypeScript")
	assert.NotContains(t, contentStr, "{Stack}") // Placeholder should be replaced

	// Verify skill content included from embedded profiles
	assert.Contains(t, contentStr, "# TypeScript Skill")
	assert.Contains(t, contentStr, "<skill_expertise>")

	// Verify principal-engineer.md alias created (copy of primary stack)
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer.md")
	aliasContent, err := h.FS.Open("/project/.claude/agents/principal-engineer.md")
	require.NoError(t, err)
	defer aliasContent.Close()

	aliasBuf := make([]byte, 16384)
	aliasN, err := aliasContent.Read(aliasBuf)
	require.NoError(t, err)
	aliasContentStr := string(aliasBuf[:aliasN])

	// Alias should have same content as principal-engineer-typescript
	assert.Contains(t, aliasContentStr, "Principal Software Engineer specializing in TypeScript")
	assert.Contains(t, aliasContentStr, "# TypeScript Skill")
	assert.Contains(t, aliasContentStr, "<skill_expertise>")

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

	// Create optional hooks config (roles and skills are now embedded)
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
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

	buf := make([]byte, 16384)
	n, err := aliasContent.Read(buf)
	require.NoError(t, err)
	contentStr := string(buf[:n])

	// Should contain TypeScript content from embedded profiles (first stack)
	assert.Contains(t, contentStr, "Principal Software Engineer specializing in TypeScript")
	assert.Contains(t, contentStr, "# TypeScript Skill")
	assert.Contains(t, contentStr, "<skill_expertise>")
}

// Test_Execute_XDGConfigHome verifies that XDG_CONFIG_HOME
// is respected when looking for the claudex config directory.
func Test_Execute_XDGConfigHome(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")
	h.Env.Set("XDG_CONFIG_HOME", "/custom/config")

	// Create optional hooks in custom XDG location (profiles are now embedded)
	h.SetupConfigDir("/custom/config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
	})

	// Create project
	h.CreateDir("/project")
	h.WriteFile("/project/package.json", `{"name": "test"}`)

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - no errors
	require.NoError(t, err)

	// Verify - structure created using XDG_CONFIG_HOME for hooks, profiles from embedded
	testutil.AssertDirExists(t, h.FS, "/project/.claude")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/team-lead.md")
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

	// Create optional hooks config (roles and skills are now embedded)
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
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
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer-php.md")

	// Verify - principal-engineer alias created (first default: typescript)
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer.md")
}

// Test_Execute_GeneratesAbsoluteHookPaths verifies that hook paths
// in settings.local.json are replaced with absolute paths.
func Test_Execute_GeneratesAbsoluteHookPaths(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create optional hooks config
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
	})

	// Create project
	h.CreateDir("/project")
	h.WriteFile("/project/package.json", `{"name": "test"}`)

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - no errors
	require.NoError(t, err)

	// Verify - settings file exists
	testutil.AssertFileExists(t, h.FS, "/project/.claude/settings.local.json")

	// Verify - all hook paths are absolute (not relative)
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/notification-hook.sh")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/session-end.sh")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/subagent-stop.sh")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/pre-tool-use.sh")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/post-tool-use.sh")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/auto-doc-updater.sh")

	// Verify - no relative paths remain
	content, err := afero.ReadFile(h.FS, "/project/.claude/settings.local.json")
	require.NoError(t, err)
	assert.NotContains(t, string(content), `".claude/hooks/`)
}

// Test_Execute_RunningTwiceDoesNotCorruptPaths verifies that running
// setup multiple times does not create duplicate hooks or corrupted paths.
func Test_Execute_RunningTwiceDoesNotCorruptPaths(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create optional hooks config
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
		"hooks/auto-doc-updater.sh":  "#!/bin/bash\necho doc",
	})

	// Create project
	h.CreateDir("/project")
	h.WriteFile("/project/package.json", `{"name": "test"}`)

	// Create usecase
	uc := New(h.FS, h.Env)

	// Run setup first time
	err := uc.Execute("/project", false)
	require.NoError(t, err)

	// Read settings after first run
	firstRunContent, err := afero.ReadFile(h.FS, "/project/.claude/settings.local.json")
	require.NoError(t, err)

	// Verify first run has absolute paths
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/notification-hook.sh")
	assert.NotContains(t, string(firstRunContent), `".claude/hooks/`)

	// Run setup second time
	err = uc.Execute("/project", false)
	require.NoError(t, err)

	// Read settings after second run
	secondRunContent, err := afero.ReadFile(h.FS, "/project/.claude/settings.local.json")
	require.NoError(t, err)

	// Verify no corrupted paths like "/project/.claude/hooks//project/.claude/hooks/"
	assert.NotContains(t, string(secondRunContent), "/project/.claude/hooks//project/.claude/hooks/")
	assert.NotContains(t, string(secondRunContent), "//project/")

	// Verify absolute paths are still present and correct
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/notification-hook.sh")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/auto-doc-updater.sh")

	// Verify no relative paths remain
	assert.NotContains(t, string(secondRunContent), `".claude/hooks/`)

	// Count occurrences of notification-hook.sh to ensure no duplicates
	notificationCount := strings.Count(string(secondRunContent), "notification-hook.sh")
	if notificationCount != 1 {
		t.Errorf("expected notification-hook.sh to appear once, got %d occurrences", notificationCount)
	}
}

// Test_Execute_IncludesCommonSkill verifies that the software-design-principles
// common skill is included in all generated engineer agents.
func Test_Execute_IncludesCommonSkill(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create optional hooks config
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
	})

	// Create project with package.json (TypeScript marker)
	h.CreateDir("/project")
	h.WriteFile("/project/package.json", `{"name": "test-project"}`)

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - no errors
	require.NoError(t, err)

	// Verify - principal-engineer-typescript.md exists
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/principal-engineer-typescript.md")

	// Read and verify content includes common skill
	content, err := h.FS.Open("/project/.claude/agents/principal-engineer-typescript.md")
	require.NoError(t, err)
	defer content.Close()

	buf := make([]byte, 32768) // Larger buffer for full content including common skill
	n, err := content.Read(buf)
	require.NoError(t, err)
	contentStr := string(buf[:n])

	// Verify common skill content is included
	assert.Contains(t, contentStr, "# Software Design Principles")
	assert.Contains(t, contentStr, "Fail-fast over silent fallbacks")
	assert.Contains(t, contentStr, "Make illegal states unrepresentable")
}

// Test_Execute_CommonSkillAfterStackSkill verifies that the common skill
// appears after the stack-specific skill in the assembled agent.
func Test_Execute_CommonSkillAfterStackSkill(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create optional hooks config
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
	})

	// Create project with package.json (TypeScript marker)
	h.CreateDir("/project")
	h.WriteFile("/project/package.json", `{"name": "test-project"}`)

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - no errors
	require.NoError(t, err)

	// Read content
	content, err := h.FS.Open("/project/.claude/agents/principal-engineer-typescript.md")
	require.NoError(t, err)
	defer content.Close()

	buf := make([]byte, 32768)
	n, err := content.Read(buf)
	require.NoError(t, err)
	contentStr := string(buf[:n])

	// Find positions of stack skill and common skill markers
	stackSkillPos := strings.Index(contentStr, "# TypeScript Skill")
	commonSkillPos := strings.Index(contentStr, "# Software Design Principles")

	// Verify both exist
	assert.NotEqual(t, -1, stackSkillPos, "Stack skill should be present")
	assert.NotEqual(t, -1, commonSkillPos, "Common skill should be present")

	// Verify common skill comes after stack skill
	assert.Greater(t, commonSkillPos, stackSkillPos, "Common skill should appear after stack skill")
}

// Test_Execute_MergesWithAbsolutePaths verifies that when merging settings,
// the resulting hook paths are absolute.
func Test_Execute_MergesWithAbsolutePaths(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create optional hooks config
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
	})

	// Create project with existing settings (custom hook with relative path)
	h.CreateDir("/project/.claude")
	existingSettings := `{
  "permissions": {
    "allow": ["Bash(custom:*)"],
    "deny": [],
    "ask": []
  },
  "hooks": {
    "PostToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/my-custom-hook.sh"
          }
        ]
      }
    ]
  }
}`
	h.WriteFile("/project/.claude/settings.local.json", existingSettings)
	h.WriteFile("/project/package.json", `{"name": "test"}`)

	// Create usecase and exercise
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Verify - no errors
	require.NoError(t, err)

	// Verify - custom hook preserved and converted to absolute path
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/my-custom-hook.sh")

	// Verify - template hooks added with absolute paths
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/notification-hook.sh")
	testutil.AssertFileContains(t, h.FS, "/project/.claude/settings.local.json", "/project/.claude/hooks/session-end.sh")

	// Verify - no relative paths remain
	content, err := afero.ReadFile(h.FS, "/project/.claude/settings.local.json")
	require.NoError(t, err)
	assert.NotContains(t, string(content), `".claude/hooks/`)
}

// Test_Execute_CleansUpDeprecatedAgents verifies that deprecated agent files
// (architect.md, researcher.md) are removed during setup (e.g., after upgrade).
func Test_Execute_CleansUpDeprecatedAgents(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.Env.Set("HOME", "/home/user")

	// Create optional hooks config
	h.SetupConfigDir("/home/user/.config/claudex", map[string]string{
		"hooks/notification-hook.sh": "#!/bin/bash\necho notify",
	})

	// Create project with deprecated agent files (simulating upgrade scenario)
	h.CreateDir("/project")
	h.WriteFile("/project/package.json", `{"name": "test"}`)
	h.FS.MkdirAll("/project/.claude/agents", 0755)
	h.FS.MkdirAll("/project/.claude/commands/agents", 0755)
	// Create deprecated agents that should be removed
	afero.WriteFile(h.FS, "/project/.claude/agents/architect.md", []byte("old architect"), 0644)
	afero.WriteFile(h.FS, "/project/.claude/agents/researcher.md", []byte("old researcher"), 0644)
	afero.WriteFile(h.FS, "/project/.claude/commands/agents/architect.md", []byte("old architect"), 0644)
	afero.WriteFile(h.FS, "/project/.claude/commands/agents/researcher.md", []byte("old researcher"), 0644)
	// Create a custom agent that should be preserved
	afero.WriteFile(h.FS, "/project/.claude/agents/custom-agent.md", []byte("custom agent"), 0644)

	// Act
	uc := New(h.FS, h.Env)
	err := uc.Execute("/project", false)

	// Assert - no errors
	require.NoError(t, err)

	// Assert - deprecated agents should be removed
	testutil.AssertNoFileExists(t, h.FS, "/project/.claude/agents/architect.md")
	testutil.AssertNoFileExists(t, h.FS, "/project/.claude/agents/researcher.md")
	testutil.AssertNoFileExists(t, h.FS, "/project/.claude/commands/agents/architect.md")
	testutil.AssertNoFileExists(t, h.FS, "/project/.claude/commands/agents/researcher.md")

	// Assert - valid agents should exist
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/team-lead.md")
	testutil.AssertFileExists(t, h.FS, "/project/.claude/commands/agents/team-lead.md")

	// Assert - custom user agents should be preserved
	testutil.AssertFileExists(t, h.FS, "/project/.claude/agents/custom-agent.md")
}
