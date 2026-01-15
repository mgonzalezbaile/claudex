// Package stackdetect provides technology stack detection for projects.
// It identifies project technologies (TypeScript, Go, Python, PHP) by scanning
// for marker files like tsconfig.json, go.mod, pyproject.toml, etc.
package stackdetect

import (
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Detect detects technology stacks based on marker files (searches up to 3 levels deep).
// It returns a list of detected stack identifiers such as "typescript", "go", "python", "php".
func Detect(fs afero.Fs, projectDir string) []string {
	var stacks []string

	// React Native detection (before TypeScript - RN projects also have package.json)
	if FindFile(fs, projectDir, "app.json", 3) ||
		FindFile(fs, projectDir, "react-native.config.js", 3) ||
		FindFile(fs, projectDir, "metro.config.js", 3) {
		stacks = append(stacks, "react-native")
	}

	// TypeScript detection
	if FindFile(fs, projectDir, "tsconfig.json", 3) {
		stacks = append(stacks, "typescript")
	} else if FindFile(fs, projectDir, "package.json", 3) {
		stacks = append(stacks, "typescript")
	}

	// Go detection
	if FindFile(fs, projectDir, "go.mod", 3) {
		stacks = append(stacks, "go")
	}

	// Python detection
	if FindFile(fs, projectDir, "pyproject.toml", 3) ||
		FindFile(fs, projectDir, "requirements.txt", 3) ||
		FindFile(fs, projectDir, "setup.py", 3) ||
		FindFile(fs, projectDir, "Pipfile", 3) {
		stacks = append(stacks, "python")
	}

	// PHP detection
	if FindFile(fs, projectDir, "composer.json", 3) ||
		FindFile(fs, projectDir, "index.php", 3) ||
		FindFile(fs, projectDir, "artisan", 3) {
		stacks = append(stacks, "php")
	}

	return stacks
}

// FindFile searches for a file in projectDir and subdirectories up to maxDepth.
// It performs a breadth-first search, skipping hidden directories (those starting with '.').
func FindFile(fs afero.Fs, dir string, filename string, maxDepth int) bool {
	if maxDepth < 0 {
		return false
	}

	// Check current directory
	if FileExists(fs, filepath.Join(dir, filename)) {
		return true
	}

	// Search subdirectories
	entries, err := afero.ReadDir(fs, dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			if FindFile(fs, filepath.Join(dir, entry.Name()), filename, maxDepth-1) {
				return true
			}
		}
	}

	return false
}

// FileExists checks if a file exists at the given path.
func FileExists(fs afero.Fs, path string) bool {
	_, err := fs.Stat(path)
	return err == nil
}
