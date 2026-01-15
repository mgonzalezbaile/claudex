package rangeupdater

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"claudex/internal/services/doctracking"
	"claudex/internal/services/lock"

	"github.com/spf13/afero"
)

// Mock implementations for testing

type mockGitService struct {
	currentSHA     string
	changedFiles   []string
	validateResult bool
	mergeBase      string
	validateError  error
	changedError   error
	mergeBaseError error
}

func (m *mockGitService) GetCurrentSHA() (string, error) {
	return m.currentSHA, nil
}

func (m *mockGitService) GetChangedFiles(base, head string) ([]string, error) {
	if m.changedError != nil {
		return nil, m.changedError
	}
	return m.changedFiles, nil
}

func (m *mockGitService) ValidateCommit(sha string) (bool, error) {
	return m.validateResult, m.validateError
}

func (m *mockGitService) GetMergeBase(branch string) (string, error) {
	if m.mergeBaseError != nil {
		return "", m.mergeBaseError
	}
	return m.mergeBase, nil
}

type mockLockService struct {
	isLocked     bool
	acquireFails bool
	fs           afero.Fs
}

func newMockLockService() *mockLockService {
	return &mockLockService{
		fs: afero.NewMemMapFs(),
	}
}

func (m *mockLockService) Acquire(path string) (*lock.Lock, error) {
	if m.acquireFails {
		return nil, fmt.Errorf("failed to acquire lock")
	}
	// Use the real lock.FileLock to create a proper Lock instance
	realLockSvc := lock.New(m.fs)
	return realLockSvc.Acquire(path)
}

func (m *mockLockService) IsLocked(path string) (bool, error) {
	return m.isLocked, nil
}

type mockFile struct{}

func (m *mockFile) Close() error                                  { return nil }
func (m *mockFile) Read(p []byte) (n int, err error)              { return 0, io.EOF }
func (m *mockFile) ReadAt(p []byte, off int64) (n int, err error) { return 0, io.EOF }
func (m *mockFile) Seek(offset int64, whence int) (int64, error)  { return 0, nil }
func (m *mockFile) Write(p []byte) (n int, err error)             { return len(p), nil }
func (m *mockFile) WriteAt(p []byte, off int64) (n int, err error) {
	return len(p), nil
}
func (m *mockFile) Name() string { return "mockfile" }
func (m *mockFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}
func (m *mockFile) Readdirnames(n int) ([]string, error) { return nil, nil }
func (m *mockFile) Stat() (os.FileInfo, error)           { return nil, nil }
func (m *mockFile) Sync() error                          { return nil }
func (m *mockFile) Truncate(size int64) error            { return nil }
func (m *mockFile) WriteString(s string) (ret int, err error) {
	return len(s), nil
}

type mockTrackingService struct {
	tracking    doctracking.DocUpdateTracking
	writeError  error
	writeCalled bool
}

func (m *mockTrackingService) Read() (doctracking.DocUpdateTracking, error) {
	return m.tracking, nil
}

func (m *mockTrackingService) Write(tracking doctracking.DocUpdateTracking) error {
	if m.writeError != nil {
		return m.writeError
	}
	m.tracking = tracking
	m.writeCalled = true
	return nil
}

func (m *mockTrackingService) Initialize(headSHA string) error {
	m.tracking = doctracking.DocUpdateTracking{
		LastProcessedCommit: headSHA,
		UpdatedAt:           time.Now().Format(time.RFC3339),
		StrategyVersion:     "v1",
	}
	return nil
}

type mockCommander struct {
	output []byte
	err    error
}

func (m *mockCommander) Run(name string, args ...string) ([]byte, error) {
	return m.output, m.err
}

func (m *mockCommander) Start(name string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error {
	return nil
}

type mockEnvironment struct {
	vars map[string]string
}

func (m *mockEnvironment) Get(key string) string {
	return m.vars[key]
}

func (m *mockEnvironment) Set(key, value string) {
	if m.vars == nil {
		m.vars = make(map[string]string)
	}
	m.vars[key] = value
}

// Test cases

func TestRangeUpdater_Run_FirstRun_Initializes(t *testing.T) {
	fs := afero.NewMemMapFs()
	sessionPath := "/session"
	fs.MkdirAll(sessionPath, 0755)

	gitSvc := &mockGitService{
		currentSHA: "abc123",
	}
	lockSvc := newMockLockService()
	trackingSvc := &mockTrackingService{}
	cmdr := &mockCommander{}
	env := &mockEnvironment{vars: make(map[string]string)}

	config := RangeUpdaterConfig{
		SessionPath:   sessionPath,
		DefaultBranch: "main",
	}

	updater := New(config, gitSvc, lockSvc, trackingSvc, cmdr, fs, env)
	result, err := updater.Run()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", result.Status)
	}

	if trackingSvc.tracking.LastProcessedCommit != "abc123" {
		t.Errorf("expected tracking to be initialized with 'abc123', got '%s'", trackingSvc.tracking.LastProcessedCommit)
	}
}

func TestRangeUpdater_Run_NoNewCommits_Skips(t *testing.T) {
	fs := afero.NewMemMapFs()
	sessionPath := "/session"
	fs.MkdirAll(sessionPath, 0755)

	gitSvc := &mockGitService{
		currentSHA: "abc123",
	}
	lockSvc := newMockLockService()
	trackingSvc := &mockTrackingService{
		tracking: doctracking.DocUpdateTracking{
			LastProcessedCommit: "abc123", // Same as current
		},
	}
	cmdr := &mockCommander{}
	env := &mockEnvironment{vars: make(map[string]string)}

	config := RangeUpdaterConfig{
		SessionPath:   sessionPath,
		DefaultBranch: "main",
	}

	updater := New(config, gitSvc, lockSvc, trackingSvc, cmdr, fs, env)
	result, err := updater.Run()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "skipped" {
		t.Errorf("expected status 'skipped', got '%s'", result.Status)
	}

	if result.Reason != "no new commits since last update" {
		t.Errorf("unexpected reason: %s", result.Reason)
	}
}

func TestRangeUpdater_Run_Locked_SkipsWithLockStatus(t *testing.T) {
	fs := afero.NewMemMapFs()
	sessionPath := "/session"
	fs.MkdirAll(sessionPath, 0755)

	gitSvc := &mockGitService{
		currentSHA: "abc123",
	}
	lockSvc := newMockLockService()
	lockSvc.isLocked = true // Already locked
	trackingSvc := &mockTrackingService{}
	cmdr := &mockCommander{}
	env := &mockEnvironment{vars: make(map[string]string)}

	config := RangeUpdaterConfig{
		SessionPath:   sessionPath,
		DefaultBranch: "main",
	}

	updater := New(config, gitSvc, lockSvc, trackingSvc, cmdr, fs, env)
	result, err := updater.Run()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "locked" {
		t.Errorf("expected status 'locked', got '%s'", result.Status)
	}
}

func TestRangeUpdater_Run_SkipRules_AllMarkdown(t *testing.T) {
	fs := afero.NewMemMapFs()
	sessionPath := "/session"
	fs.MkdirAll(sessionPath, 0755)

	gitSvc := &mockGitService{
		currentSHA:     "def456",
		changedFiles:   []string{"docs/readme.md", "docs/guide.md"},
		validateResult: true,
	}
	lockSvc := newMockLockService()
	trackingSvc := &mockTrackingService{
		tracking: doctracking.DocUpdateTracking{
			LastProcessedCommit: "abc123",
		},
	}
	cmdr := &mockCommander{}
	env := &mockEnvironment{vars: make(map[string]string)}

	config := RangeUpdaterConfig{
		SessionPath:   sessionPath,
		DefaultBranch: "main",
	}

	updater := New(config, gitSvc, lockSvc, trackingSvc, cmdr, fs, env)
	result, err := updater.Run()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "skipped" {
		t.Errorf("expected status 'skipped', got '%s'", result.Status)
	}

	if result.Reason == "" || len(result.Reason) < 10 {
		t.Errorf("expected skip reason about markdown files, got: %s", result.Reason)
	}
}

func TestRangeUpdater_Run_SkipRules_EnvVar(t *testing.T) {
	fs := afero.NewMemMapFs()
	sessionPath := "/session"
	fs.MkdirAll(sessionPath, 0755)

	gitSvc := &mockGitService{
		currentSHA:     "def456",
		changedFiles:   []string{"src/foo.go"},
		validateResult: true,
	}
	lockSvc := newMockLockService()
	trackingSvc := &mockTrackingService{
		tracking: doctracking.DocUpdateTracking{
			LastProcessedCommit: "abc123",
		},
	}
	cmdr := &mockCommander{}
	env := &mockEnvironment{
		vars: map[string]string{
			"CLAUDEX_SKIP_DOCS": "1",
		},
	}

	config := RangeUpdaterConfig{
		SessionPath:   sessionPath,
		DefaultBranch: "main",
	}

	updater := New(config, gitSvc, lockSvc, trackingSvc, cmdr, fs, env)
	result, err := updater.Run()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "skipped" {
		t.Errorf("expected status 'skipped', got '%s'", result.Status)
	}

	if result.Reason == "" {
		t.Errorf("expected skip reason about environment variable")
	}
}

func TestRangeUpdater_Run_UnreachableBase_UsesFallback(t *testing.T) {
	fs := afero.NewMemMapFs()
	sessionPath := "/session"
	fs.MkdirAll(sessionPath, 0755)

	gitSvc := &mockGitService{
		currentSHA:     "def456",
		changedFiles:   []string{"src/foo.go"},
		validateResult: false, // Base is unreachable
		mergeBase:      "fallback123",
	}
	lockSvc := newMockLockService()
	trackingSvc := &mockTrackingService{
		tracking: doctracking.DocUpdateTracking{
			LastProcessedCommit: "unreachable",
		},
	}
	cmdr := &mockCommander{}
	env := &mockEnvironment{vars: make(map[string]string)}

	config := RangeUpdaterConfig{
		SessionPath:   sessionPath,
		DefaultBranch: "main",
	}

	// Create an index.md to be affected
	fs.MkdirAll("/src", 0755)
	afero.WriteFile(fs, "/src/index.md", []byte("# Index"), 0644)

	updater := New(config, gitSvc, lockSvc, trackingSvc, cmdr, fs, env)
	result, err := updater.Run()

	// Should succeed with fallback
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", result.Status)
	}

	// Verify tracking was updated
	if !trackingSvc.writeCalled {
		t.Error("expected tracking to be updated")
	}
}

func TestRangeUpdater_Run_NoAffectedIndexes_UpdatesTracking(t *testing.T) {
	fs := afero.NewMemMapFs()
	sessionPath := "/session"
	fs.MkdirAll(sessionPath, 0755)

	gitSvc := &mockGitService{
		currentSHA:     "def456",
		changedFiles:   []string{"orphan/foo.go"}, // No parent index.md
		validateResult: true,
	}
	lockSvc := newMockLockService()
	trackingSvc := &mockTrackingService{
		tracking: doctracking.DocUpdateTracking{
			LastProcessedCommit: "abc123",
		},
	}
	cmdr := &mockCommander{}
	env := &mockEnvironment{vars: make(map[string]string)}

	config := RangeUpdaterConfig{
		SessionPath:   sessionPath,
		DefaultBranch: "main",
	}

	updater := New(config, gitSvc, lockSvc, trackingSvc, cmdr, fs, env)
	result, err := updater.Run()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", result.Status)
	}

	// Verify tracking was updated even though no indexes were affected
	if !trackingSvc.writeCalled {
		t.Error("expected tracking to be updated")
	}

	if trackingSvc.tracking.LastProcessedCommit != "def456" {
		t.Errorf("expected tracking to be updated to 'def456', got '%s'", trackingSvc.tracking.LastProcessedCommit)
	}
}

func TestResolveAffectedIndexes_MultipleIndexes(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create directory structure with indexes
	fs.MkdirAll("/pkg/a", 0755)
	fs.MkdirAll("/pkg/b", 0755)
	afero.WriteFile(fs, "/pkg/a/index.md", []byte("# A"), 0644)
	afero.WriteFile(fs, "/pkg/b/index.md", []byte("# B"), 0644)

	changedFiles := []string{
		"/pkg/a/foo.go",
		"/pkg/a/bar.go",
		"/pkg/b/baz.go",
	}

	indexes, err := ResolveAffectedIndexes(fs, changedFiles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(indexes) != 2 {
		t.Errorf("expected 2 affected indexes, got %d", len(indexes))
	}

	// Verify both indexes are present
	hasA := false
	hasB := false
	for _, idx := range indexes {
		if idx == "/pkg/a/index.md" {
			hasA = true
		}
		if idx == "/pkg/b/index.md" {
			hasB = true
		}
	}

	if !hasA || !hasB {
		t.Errorf("expected both /pkg/a/index.md and /pkg/b/index.md, got %v", indexes)
	}
}

func TestResolveAffectedIndexes_Deduplication(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create directory structure
	fs.MkdirAll("/pkg/a", 0755)
	afero.WriteFile(fs, "/pkg/a/index.md", []byte("# A"), 0644)

	changedFiles := []string{
		"/pkg/a/foo.go",
		"/pkg/a/bar.go",
		"/pkg/a/baz.go",
	}

	indexes, err := ResolveAffectedIndexes(fs, changedFiles)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(indexes) != 1 {
		t.Errorf("expected 1 affected index (deduplicated), got %d", len(indexes))
	}

	if indexes[0] != "/pkg/a/index.md" {
		t.Errorf("expected /pkg/a/index.md, got %s", indexes[0])
	}
}

func TestShouldSkip_EnvVar(t *testing.T) {
	env := &mockEnvironment{
		vars: map[string]string{
			"CLAUDEX_SKIP_DOCS": "1",
		},
	}

	skip, reason := ShouldSkip([]string{"foo.go"}, "", env)

	if !skip {
		t.Error("expected skip to be true for CLAUDEX_SKIP_DOCS=1")
	}

	if reason == "" {
		t.Error("expected reason to be set")
	}
}

func TestShouldSkip_CommitMessageTag(t *testing.T) {
	env := &mockEnvironment{vars: make(map[string]string)}

	skip, reason := ShouldSkip([]string{"foo.go"}, "fix: typo [skip-docs]", env)

	if !skip {
		t.Error("expected skip to be true for [skip-docs] tag")
	}

	if reason == "" {
		t.Error("expected reason to be set")
	}
}

func TestShouldSkip_AllMarkdownFiles(t *testing.T) {
	env := &mockEnvironment{vars: make(map[string]string)}

	skip, reason := ShouldSkip([]string{"docs/readme.md", "docs/guide.md"}, "", env)

	if !skip {
		t.Error("expected skip to be true for all markdown files")
	}

	if reason == "" {
		t.Error("expected reason to be set")
	}
}

func TestShouldSkip_MixedFiles_NoSkip(t *testing.T) {
	env := &mockEnvironment{vars: make(map[string]string)}

	skip, _ := ShouldSkip([]string{"src/foo.go", "docs/readme.md"}, "", env)

	if skip {
		t.Error("expected skip to be false for mixed file types")
	}
}

func TestHandleUnreachableBase_DefaultBranch(t *testing.T) {
	gitSvc := &mockGitService{
		mergeBase: "fallback123",
	}

	sha, err := HandleUnreachableBase(gitSvc, "main")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sha != "fallback123" {
		t.Errorf("expected 'fallback123', got '%s'", sha)
	}
}

func TestHandleUnreachableBase_FallsBackToMain(t *testing.T) {
	// Create a mock that will fail for "develop" but succeed for "main"
	gitSvc := &mockGitServiceWithCallback{
		mergeBaseCallback: func(branch string) (string, error) {
			if branch == "develop" {
				return "", fmt.Errorf("branch not found")
			}
			if branch == "main" {
				return "main-fallback", nil
			}
			return "", fmt.Errorf("branch not found")
		},
	}

	sha, err := HandleUnreachableBase(gitSvc, "develop")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sha != "main-fallback" {
		t.Errorf("expected 'main-fallback', got '%s'", sha)
	}
}

// mockGitServiceWithCallback allows custom behavior per call
type mockGitServiceWithCallback struct {
	mergeBaseCallback func(branch string) (string, error)
}

func (m *mockGitServiceWithCallback) GetCurrentSHA() (string, error) {
	return "", nil
}

func (m *mockGitServiceWithCallback) GetChangedFiles(base, head string) ([]string, error) {
	return nil, nil
}

func (m *mockGitServiceWithCallback) ValidateCommit(sha string) (bool, error) {
	return false, nil
}

func (m *mockGitServiceWithCallback) GetMergeBase(branch string) (string, error) {
	if m.mergeBaseCallback != nil {
		return m.mergeBaseCallback(branch)
	}
	return "", fmt.Errorf("not implemented")
}

func TestHandleUnreachableBase_AllFail_ReturnsError(t *testing.T) {
	gitSvc := &mockGitService{
		mergeBaseError: fmt.Errorf("no merge base found"),
	}

	_, err := HandleUnreachableBase(gitSvc, "main")

	if err == nil {
		t.Error("expected error when all fallback attempts fail")
	}
}

// Test helper to create temp directories
func createTempDir(t *testing.T) (string, func()) {
	dir, err := os.MkdirTemp("", "rangeupdater-test-*")
	if err != nil {
		t.Fatal(err)
	}
	return dir, func() { os.RemoveAll(dir) }
}
