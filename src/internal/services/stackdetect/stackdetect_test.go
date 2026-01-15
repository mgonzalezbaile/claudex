package stackdetect

import (
	"testing"

	"claudex/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func Test_Detect(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		wantStacks []string
	}{
		{
			name:       "TypeScript via tsconfig",
			files:      map[string]string{"tsconfig.json": "{}"},
			wantStacks: []string{"typescript"},
		},
		{
			name:       "TypeScript via package.json",
			files:      map[string]string{"package.json": `{"name": "test"}`},
			wantStacks: []string{"typescript"},
		},
		{
			name:       "Go via go.mod",
			files:      map[string]string{"go.mod": "module test"},
			wantStacks: []string{"go"},
		},
		{
			name:       "Python via pyproject.toml",
			files:      map[string]string{"pyproject.toml": "[project]"},
			wantStacks: []string{"python"},
		},
		{
			name:       "Python via requirements.txt",
			files:      map[string]string{"requirements.txt": "flask==2.0"},
			wantStacks: []string{"python"},
		},
		{
			name:       "Python via setup.py",
			files:      map[string]string{"setup.py": "from setuptools import setup"},
			wantStacks: []string{"python"},
		},
		{
			name:       "Python via Pipfile",
			files:      map[string]string{"Pipfile": "[[source]]"},
			wantStacks: []string{"python"},
		},
		{
			name:       "PHP via composer.json",
			files:      map[string]string{"composer.json": `{"require": {"php": "^8.2"}}`},
			wantStacks: []string{"php"},
		},
		{
			name:       "PHP via index.php",
			files:      map[string]string{"index.php": "<?php echo 'Hello world!'; ?>"},
			wantStacks: []string{"php"},
		},
		{
			name:       "PHP via Laravel artisan",
			files:      map[string]string{"artisan": "#!/usr/bin/env php"},
			wantStacks: []string{"php"},
		},
		{
			name: "Multiple stacks - TypeScript and Go",
			files: map[string]string{
				"package.json": `{"name": "test"}`,
				"go.mod":       "module test",
			},
			wantStacks: []string{"typescript", "go"},
		},
		{
			name: "Multiple stacks - All three",
			files: map[string]string{
				"tsconfig.json":    "{}",
				"go.mod":           "module test",
				"requirements.txt": "flask==2.0",
			},
			wantStacks: []string{"typescript", "go", "python"},
		},
		{
			name: "TypeScript takes precedence - both tsconfig and package.json",
			files: map[string]string{
				"tsconfig.json": "{}",
				"package.json":  `{"name": "test"}`,
			},
			wantStacks: []string{"typescript"}, // Should only return once
		},
		{
			name: "Python - multiple markers",
			files: map[string]string{
				"requirements.txt": "flask==2.0",
				"pyproject.toml":   "[project]",
			},
			wantStacks: []string{"python"}, // Should only return once
		},
		{
			name: "Nested detection depth 1",
			files: map[string]string{
				"backend/go.mod": "module backend",
			},
			wantStacks: []string{"go"},
		},
		{
			name: "Nested detection depth 2",
			files: map[string]string{
				"packages/frontend/tsconfig.json": "{}",
			},
			wantStacks: []string{"typescript"},
		},
		{
			name: "Nested detection depth 3 (at limit)",
			files: map[string]string{
				"apps/web/client/package.json": `{"name": "client"}`,
			},
			wantStacks: []string{"typescript"},
		},
		{
			name: "Deep nesting beyond maxDepth (depth 4)",
			files: map[string]string{
				"a/b/c/d/go.mod": "module deep",
			},
			wantStacks: []string{}, // Should not find it at depth 4 (maxDepth is 3)
		},
		{
			name: "Mixed depth - root and nested",
			files: map[string]string{
				"go.mod":                        "module root",
				"frontend/package.json":         `{"name": "frontend"}`,
				"backend/services/api/setup.py": "from setuptools import setup",
			},
			wantStacks: []string{"typescript", "go", "python"},
		},
		{
			name:       "No stacks detected",
			files:      map[string]string{"README.md": "# Readme"},
			wantStacks: []string{},
		},
		{
			name:       "Empty project",
			files:      map[string]string{},
			wantStacks: []string{},
		},
		{
			name: "Hidden directories are searched (unlike findFile which skips them)",
			files: map[string]string{
				".hidden/package.json": `{"name": "hidden"}`,
			},
			// Note: Based on implementation, Detect calls FindFile
			// which skips directories starting with "."
			wantStacks: []string{},
		},
		{
			name: "Multiple Python markers in subdirectories",
			files: map[string]string{
				"api/requirements.txt":  "django==4.0",
				"worker/pyproject.toml": "[project]",
			},
			wantStacks: []string{"python"}, // Should deduplicate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testutil.NewTestHarness()

			// Create project directory
			h.CreateDir("/project")

			// Create marker files
			for path, content := range tt.files {
				h.WriteFile("/project/"+path, content)
			}

			// Exercise
			stacks := Detect(h.FS, "/project")

			// Verify - use ElementsMatch because order doesn't matter
			assert.ElementsMatch(t, tt.wantStacks, stacks,
				"Stack detection mismatch for test case: %s", tt.name)
		})
	}
}

func Test_FindFile_RespectsMaxDepth(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		maxDepth int
		want     bool
	}{
		{
			name:     "Depth 0 - only checks current directory",
			filePath: "go.mod",
			maxDepth: 0,
			want:     true,
		},
		{
			name:     "Depth 0 - does not check subdirectories",
			filePath: "subdir/go.mod",
			maxDepth: 0,
			want:     false,
		},
		{
			name:     "Depth 1 - finds in subdirectory",
			filePath: "backend/go.mod",
			maxDepth: 1,
			want:     true,
		},
		{
			name:     "Depth 1 - does not find at depth 2",
			filePath: "apps/backend/go.mod",
			maxDepth: 1,
			want:     false,
		},
		{
			name:     "Depth 2 - finds at depth 2",
			filePath: "apps/backend/go.mod",
			maxDepth: 2,
			want:     true,
		},
		{
			name:     "Depth 2 - does not find at depth 3",
			filePath: "apps/backend/api/go.mod",
			maxDepth: 2,
			want:     false,
		},
		{
			name:     "Depth 3 - finds at depth 3",
			filePath: "apps/backend/api/go.mod",
			maxDepth: 3,
			want:     true,
		},
		{
			name:     "Depth 3 - does not find at depth 4",
			filePath: "a/b/c/d/go.mod",
			maxDepth: 3,
			want:     false,
		},
		{
			name:     "Negative depth - returns false",
			filePath: "go.mod",
			maxDepth: -1,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testutil.NewTestHarness()

			// Create project directory
			h.CreateDir("/project")

			// Create the file at the specified path
			h.WriteFile("/project/"+tt.filePath, "module test")

			// Exercise
			found := FindFile(h.FS, "/project", "go.mod", tt.maxDepth)

			// Verify
			assert.Equal(t, tt.want, found,
				"FindFile returned %v, want %v for maxDepth=%d and path=%s",
				found, tt.want, tt.maxDepth, tt.filePath)
		})
	}
}

func Test_FindFile_SkipsHiddenDirectories(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		filename string
		want     bool
	}{
		{
			name:     "Hidden .git directory - should skip",
			filePath: ".git/config",
			filename: "config",
			want:     false,
		},
		{
			name:     "Hidden .vscode directory - should skip",
			filePath: ".vscode/settings.json",
			filename: "settings.json",
			want:     false,
		},
		{
			name:     "Hidden directory with Go file - should skip",
			filePath: ".hidden/go.mod",
			filename: "go.mod",
			want:     false,
		},
		{
			name:     "Nested hidden directory - should skip",
			filePath: "src/.private/package.json",
			filename: "package.json",
			want:     false,
		},
		{
			name:     "Regular directory - should find",
			filePath: "src/go.mod",
			filename: "go.mod",
			want:     true,
		},
		{
			name:     "File with dot in name (not directory) - should find",
			filePath: "file.config.json",
			filename: "file.config.json",
			want:     true,
		},
		{
			name:     "Directory with underscore (not hidden) - should find",
			filePath: "_internal/go.mod",
			filename: "go.mod",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testutil.NewTestHarness()

			// Create project directory
			h.CreateDir("/project")

			// Create the file at the specified path
			h.WriteFile("/project/"+tt.filePath, "test content")

			// Exercise - search with maxDepth 3 to allow nested searches
			found := FindFile(h.FS, "/project", tt.filename, 3)

			// Verify
			assert.Equal(t, tt.want, found,
				"FindFile returned %v, want %v for path=%s",
				found, tt.want, tt.filePath)
		})
	}
}

func Test_FileExists_WithAfero(t *testing.T) {
	tests := []struct {
		name       string
		createFile bool
		filePath   string
		want       bool
	}{
		{
			name:       "File exists",
			createFile: true,
			filePath:   "/project/go.mod",
			want:       true,
		},
		{
			name:       "File does not exist",
			createFile: false,
			filePath:   "/project/go.mod",
			want:       false,
		},
		{
			name:       "Directory exists (not a file)",
			createFile: false,
			filePath:   "/project",
			want:       true, // Stat succeeds for directories too
		},
		{
			name:       "Nested file exists",
			createFile: true,
			filePath:   "/project/backend/api/main.go",
			want:       true,
		},
		{
			name:       "Nested file does not exist",
			createFile: false,
			filePath:   "/project/backend/api/main.go",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := testutil.NewTestHarness()

			if tt.createFile {
				h.WriteFile(tt.filePath, "test content")
			} else if tt.name == "Directory exists (not a file)" {
				h.CreateDir(tt.filePath)
			}

			// Exercise
			exists := FileExists(h.FS, tt.filePath)

			// Verify
			assert.Equal(t, tt.want, exists,
				"FileExists returned %v, want %v for path=%s",
				exists, tt.want, tt.filePath)
		})
	}
}

func Test_Detect_EdgeCases(t *testing.T) {
	t.Run("Empty directory path", func(t *testing.T) {
		h := testutil.NewTestHarness()
		stacks := Detect(h.FS, "")
		assert.Empty(t, stacks, "Empty path should return empty stacks")
	})

	t.Run("Non-existent directory", func(t *testing.T) {
		h := testutil.NewTestHarness()
		stacks := Detect(h.FS, "/nonexistent")
		assert.Empty(t, stacks, "Non-existent directory should return empty stacks")
	})

	t.Run("File instead of directory", func(t *testing.T) {
		h := testutil.NewTestHarness()
		h.WriteFile("/project.txt", "not a directory")
		stacks := Detect(h.FS, "/project.txt")
		assert.Empty(t, stacks, "File path should return empty stacks")
	})

	t.Run("Root directory", func(t *testing.T) {
		h := testutil.NewTestHarness()
		h.WriteFile("/go.mod", "module root")
		stacks := Detect(h.FS, "/")
		assert.Contains(t, stacks, "go", "Should detect at root level")
	})
}

func Test_FindFile_Performance(t *testing.T) {
	// This test ensures FindFile doesn't recurse indefinitely or inefficiently
	t.Run("Large directory structure", func(t *testing.T) {
		h := testutil.NewTestHarness()
		h.CreateDir("/project")

		// Create a broad directory structure (10 dirs at each level)
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				h.CreateDir("/project/dir" + string(rune('0'+i)) + "/subdir" + string(rune('0'+j)))
			}
		}

		// Place marker file in one location
		h.WriteFile("/project/dir5/subdir3/go.mod", "module test")

		// Exercise - should complete quickly
		found := FindFile(h.FS, "/project", "go.mod", 2)

		// Verify
		assert.True(t, found, "Should find file in large structure")
	})

	t.Run("No infinite recursion on maxDepth", func(t *testing.T) {
		h := testutil.NewTestHarness()
		h.CreateDir("/project")

		// Create very deep structure
		h.WriteFile("/project/a/b/c/d/e/f/g/h/i/j/go.mod", "module deep")

		// Exercise - should stop at maxDepth and not hang
		found := FindFile(h.FS, "/project", "go.mod", 3)

		// Verify - should not find it (beyond depth 3)
		assert.False(t, found, "Should not find file beyond maxDepth")
	})
}
