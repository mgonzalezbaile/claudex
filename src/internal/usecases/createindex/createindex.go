// Package createindex provides the usecase for generating index.md documentation
// for any directory using Claude. It scans the directory structure, finds nearby
// index.md files for style reference, and invokes Claude to generate contextually
// relevant documentation.
package createindex

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"claudex/internal/services/commander"
	"claudex/internal/services/env"

	"github.com/spf13/afero"
)

// CreateIndexUseCase orchestrates the index.md generation workflow
type CreateIndexUseCase struct {
	fs  afero.Fs
	cmd commander.Commander
	env env.Environment
}

// New creates a new CreateIndexUseCase instance with the given dependencies
func New(fs afero.Fs, cmd commander.Commander, env env.Environment) *CreateIndexUseCase {
	return &CreateIndexUseCase{
		fs:  fs,
		cmd: cmd,
		env: env,
	}
}

// codeExtensions defines the file extensions that should be included in the scan
var codeExtensions = []string{
	".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".rs",
	".java", ".kt", ".swift", ".c", ".cpp", ".h", ".rb", ".php",
}

// Execute generates an index.md file for the specified directory path.
// It validates the directory exists, scans for code files, finds nearby index.md
// for style reference, builds a prompt, and invokes Claude to generate the content.
//
// Parameters:
//   - dirPath: The directory path where index.md should be created
//
// Returns an error if the generation fails.
func (uc *CreateIndexUseCase) Execute(dirPath string) error {
	// 1. Validate directory exists
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	info, err := uc.fs.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", absPath)
		}
		return fmt.Errorf("failed to access path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is a file, not a directory: %s", absPath)
	}

	// 2. Scan directory for code files
	fileListing, err := uc.scanDirectory(absPath)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// 3. Find nearby index.md for style reference
	styleReference, err := uc.findStyleReference(absPath)
	if err != nil {
		// Non-fatal - we can proceed without a style reference
		styleReference = "(No nearby index.md found for style reference)"
	}

	// 4. Build prompt
	prompt := uc.buildPrompt(absPath, fileListing, styleReference)

	// 5. Invoke Claude with haiku model - Claude will create the file directly
	outputPath := filepath.Join(absPath, "index.md")
	if err := uc.invokeClaudeSync(prompt, outputPath); err != nil {
		return fmt.Errorf("failed to generate index.md: %w", err)
	}

	// 6. Display success message
	fmt.Printf("✓ Created index.md at: %s\n", outputPath)

	return nil
}

// scanDirectory scans the directory and returns a formatted listing of code files
func (uc *CreateIndexUseCase) scanDirectory(dirPath string) (string, error) {
	var files []string

	err := afero.Walk(uc.fs, dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the directory itself
		if path == dirPath {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			relPath = filepath.Base(path)
		}

		// Check if it's a directory
		if info.IsDir() {
			// Check if subdirectory has an index.md
			indexPath := filepath.Join(path, "index.md")
			if _, err := uc.fs.Stat(indexPath); err == nil {
				files = append(files, fmt.Sprintf("%s/ (has index.md)", relPath))
			} else {
				files = append(files, relPath+"/")
			}
			return nil
		}

		// Check if it's a code file
		ext := filepath.Ext(path)
		for _, codeExt := range codeExtensions {
			if ext == codeExt {
				files = append(files, relPath)
				break
			}
		}

		// Also include existing index.md if present
		if filepath.Base(path) == "index.md" {
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "(No code files found)", nil
	}

	return strings.Join(files, "\n"), nil
}

// findStyleReference looks for nearby index.md files in parent or sibling directories
func (uc *CreateIndexUseCase) findStyleReference(dirPath string) (string, error) {
	// Try parent directory first
	parentDir := filepath.Dir(dirPath)
	parentIndex := filepath.Join(parentDir, "index.md")

	content, err := afero.ReadFile(uc.fs, parentIndex)
	if err == nil {
		return string(content), nil
	}

	// Try sibling directories
	entries, err := afero.ReadDir(uc.fs, parentDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		siblingIndex := filepath.Join(parentDir, entry.Name(), "index.md")
		content, err := afero.ReadFile(uc.fs, siblingIndex)
		if err == nil {
			return string(content), nil
		}
	}

	return "", fmt.Errorf("no style reference found")
}

// buildPrompt constructs the Claude prompt for index.md generation
func (uc *CreateIndexUseCase) buildPrompt(dirPath, fileListing, styleReference string) string {
	return fmt.Sprintf(`Create an index.md documentation file for the directory: %s

FILES IN DIRECTORY:
%s

STYLE REFERENCE (from nearby index.md):
%s

Requirements:
- Create a lightweight documentation pointer that helps developers understand this directory
- Include a title based on the package/directory name
- Write a 1-2 sentence summary of the directory's purpose
- List key files with brief descriptions (if relevant)
- If subdirectories have index.md files, use markdown links: [subdir/](./subdir/index.md)
- Match the style and tone of the reference index.md
- Use the Write tool to create the file directly at the target path
- Do NOT output the content to stdout - write it to the file`, dirPath, fileListing, styleReference)
}

// invokeClaudeSync invokes Claude synchronously to create the index.md file directly
func (uc *CreateIndexUseCase) invokeClaudeSync(prompt string, outputPath string) error {
	// Recursion guard: check if we're already inside a hook invocation
	if uc.env.Get("CLAUDE_HOOK_INTERNAL") == "1" {
		return fmt.Errorf("recursion guard: already inside Claude hook invocation")
	}

	// Add output path to prompt so Claude knows where to write
	fullPrompt := fmt.Sprintf("%s\n\nWrite the index.md file to: %s", prompt, outputPath)

	// Create command with haiku model for cost efficiency
	cmd := exec.Command("claude", "-p", fullPrompt, "--model", "haiku")

	// Set recursion guard in environment for this command
	cmd.Env = append(os.Environ(), "CLAUDE_HOOK_INTERNAL=1")

	// Run the command - Claude will write the file directly using the Write tool
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude invocation failed: %w", err)
	}

	return nil
}
