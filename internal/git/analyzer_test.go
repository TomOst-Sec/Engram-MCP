package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/TomOst-Sec/colony-project/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRepo creates a temp git repo with known commits for testing.
func setupTestRepo(t *testing.T) (string, *storage.Store) {
	t.Helper()
	repoDir := t.TempDir()

	// Initialize git repo
	runCmd(t, repoDir, "git", "init")
	runCmd(t, repoDir, "git", "config", "user.email", "test@test.com")
	runCmd(t, repoDir, "git", "config", "user.name", "Test User")

	// Create files and commits
	writeFile(t, repoDir, "main.go", "package main\nfunc main() {}\n")
	writeFile(t, repoDir, "helper.go", "package main\nfunc helper() {}\n")
	runCmd(t, repoDir, "git", "add", ".")
	runCmd(t, repoDir, "git", "commit", "-m", "Initial commit")

	// Second commit touching both files
	writeFile(t, repoDir, "main.go", "package main\nfunc main() {\n\thelper()\n}\n")
	writeFile(t, repoDir, "helper.go", "package main\nfunc helper() string { return \"\" }\n")
	runCmd(t, repoDir, "git", "add", ".")
	runCmd(t, repoDir, "git", "commit", "-m", "Wire helper into main")

	// Third commit touching only main.go
	writeFile(t, repoDir, "main.go", "package main\nimport \"fmt\"\nfunc main() {\n\tfmt.Println(helper())\n}\n")
	runCmd(t, repoDir, "git", "add", ".")
	runCmd(t, repoDir, "git", "commit", "-m", "Add fmt output")

	// Open storage and seed code_index
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)

	// Seed code_index with the files we committed
	_, err = store.DB().Exec(
		`INSERT INTO code_index (file_path, file_hash, language, symbol_name, symbol_type, start_line, end_line) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"main.go", "hash1", "go", "main", "function", 1, 5,
	)
	require.NoError(t, err)
	_, err = store.DB().Exec(
		`INSERT INTO code_index (file_path, file_hash, language, symbol_name, symbol_type, start_line, end_line) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"helper.go", "hash2", "go", "helper", "function", 1, 2,
	)
	require.NoError(t, err)

	return repoDir, store
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "command %s %v failed: %s", name, args, string(out))
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0644))
}

func TestAnalyzeFileOnRealRepo(t *testing.T) {
	repoDir, store := setupTestRepo(t)
	defer store.Close()

	h := New(store, repoDir)
	fh, err := h.AnalyzeFile(context.Background(), "main.go")
	require.NoError(t, err)
	require.NotNil(t, fh)

	assert.Equal(t, "main.go", fh.FilePath)
	assert.Equal(t, 3, fh.ChangeFrequency) // 3 commits touch main.go
	assert.Equal(t, "Test User", fh.LastAuthor)
	assert.Equal(t, "Add fmt output", fh.LastCommitMessage)
	assert.NotEmpty(t, fh.LastCommitHash)
	assert.False(t, fh.LastModified.IsZero())
}

func TestAnalyzeFileCommitCount(t *testing.T) {
	repoDir, store := setupTestRepo(t)
	defer store.Close()

	h := New(store, repoDir)
	fh, err := h.AnalyzeFile(context.Background(), "helper.go")
	require.NoError(t, err)
	require.NotNil(t, fh)

	assert.Equal(t, 2, fh.ChangeFrequency) // 2 commits touch helper.go
}

func TestAnalyzeFileNonExistent(t *testing.T) {
	repoDir, store := setupTestRepo(t)
	defer store.Close()

	h := New(store, repoDir)
	fh, err := h.AnalyzeFile(context.Background(), "nonexistent.go")
	require.NoError(t, err)
	require.NotNil(t, fh)
	assert.Equal(t, 0, fh.ChangeFrequency)
}

func TestAnalyzeAllPopulatesGitContext(t *testing.T) {
	repoDir, store := setupTestRepo(t)
	defer store.Close()

	h := New(store, repoDir)
	stats, err := h.AnalyzeAll(context.Background())
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, 2, stats.FilesAnalyzed)
	assert.Equal(t, "main.go", stats.HottestFile)
	assert.Equal(t, 3, stats.HottestFrequency)
	assert.Greater(t, stats.Duration.Nanoseconds(), int64(0))

	// Verify data was stored
	mainFH, err := GetFileHistory(store, "main.go")
	require.NoError(t, err)
	require.NotNil(t, mainFH)
	assert.Equal(t, 3, mainFH.ChangeFrequency)

	helperFH, err := GetFileHistory(store, "helper.go")
	require.NoError(t, err)
	require.NotNil(t, helperFH)
	assert.Equal(t, 2, helperFH.ChangeFrequency)
}

func TestAnalyzeAllCoChangedFiles(t *testing.T) {
	repoDir, store := setupTestRepo(t)
	defer store.Close()

	h := New(store, repoDir)
	_, err := h.AnalyzeAll(context.Background())
	require.NoError(t, err)

	mainFH, err := GetFileHistory(store, "main.go")
	require.NoError(t, err)
	require.NotNil(t, mainFH)
	assert.Contains(t, mainFH.CoChangedFiles, "helper.go")
}

func TestGetHotspotsViaAnalyzer(t *testing.T) {
	repoDir, store := setupTestRepo(t)
	defer store.Close()

	h := New(store, repoDir)
	_, err := h.AnalyzeAll(context.Background())
	require.NoError(t, err)

	hotspots, err := h.GetHotspots(context.Background(), 10)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(hotspots), 2)
	assert.Equal(t, "main.go", hotspots[0].FilePath) // most changed
}

func TestGetCoChangedFilesViaAnalyzer(t *testing.T) {
	repoDir, store := setupTestRepo(t)
	defer store.Close()

	h := New(store, repoDir)
	_, err := h.AnalyzeAll(context.Background())
	require.NoError(t, err)

	coChanged, err := h.GetCoChangedFiles(context.Background(), "main.go", 10)
	require.NoError(t, err)
	assert.Contains(t, coChanged, "helper.go")
}

func TestGetCoChangedFilesLimit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Seed with many co-changed files
	fh := &FileHistory{
		FilePath:       "target.go",
		LastAuthor:     "dev",
		CoChangedFiles: []string{"a.go", "b.go", "c.go", "d.go", "e.go"},
	}
	require.NoError(t, UpsertFileHistory(store, fh))

	h := New(store, ".")
	coChanged, err := h.GetCoChangedFiles(context.Background(), "target.go", 2)
	require.NoError(t, err)
	assert.Len(t, coChanged, 2)
}

func TestAnalyzeAllEmptyCodeIndex(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	h := New(store, t.TempDir())
	stats, err := h.AnalyzeAll(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, stats.FilesAnalyzed)
}

// Ensure unused import doesn't cause issues
var _ = fmt.Sprintf
