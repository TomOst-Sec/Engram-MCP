package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/TomOst-Sec/colony-project/internal/config"
	"github.com/TomOst-Sec/colony-project/internal/parser"
	"github.com/TomOst-Sec/colony-project/internal/storage"
)

func setupTestEnv(t *testing.T) (string, *storage.Store, *parser.Registry, *config.Config) {
	t.Helper()

	// Create temp directory as fake repo root
	repoRoot := t.TempDir()

	// Create database
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	// Create parser registry
	registry := parser.NewDefaultRegistry()

	// Create config
	cfg := config.Default()
	cfg.RepoRoot = repoRoot

	return repoRoot, store, registry, cfg
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestIndexAllGoFiles(t *testing.T) {
	repoRoot, store, registry, cfg := setupTestEnv(t)

	// Write sample Go files
	writeFile(t, repoRoot, "main.go", `package main

func main() {}

func helper(x int) string {
	return ""
}
`)
	writeFile(t, repoRoot, "lib/utils.go", `package lib

// Add adds two numbers.
func Add(a, b int) int {
	return a + b
}

type Config struct {
	Name string
}
`)

	idx := New(store, registry, nil, cfg, repoRoot)
	stats, err := idx.IndexAll(context.Background(), false, false)
	if err != nil {
		t.Fatalf("IndexAll: %v", err)
	}

	if stats.FilesProcessed != 2 {
		t.Errorf("FilesProcessed = %d, want 2", stats.FilesProcessed)
	}
	if stats.SymbolsExtracted == 0 {
		t.Error("SymbolsExtracted = 0, want > 0")
	}

	// Verify symbols are in the database
	var count int
	err = store.DB().QueryRow("SELECT COUNT(*) FROM code_index").Scan(&count)
	if err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count == 0 {
		t.Error("no symbols in code_index")
	}
}

func TestIndexAllSkipsGitDir(t *testing.T) {
	repoRoot, store, registry, cfg := setupTestEnv(t)

	// Write a .git directory with a Go file inside
	writeFile(t, repoRoot, ".git/hooks/pre-commit.go", `package hooks
func preCommit() {}
`)
	writeFile(t, repoRoot, "main.go", `package main
func main() {}
`)

	idx := New(store, registry, nil, cfg, repoRoot)
	stats, err := idx.IndexAll(context.Background(), false, false)
	if err != nil {
		t.Fatalf("IndexAll: %v", err)
	}

	if stats.FilesProcessed != 1 {
		t.Errorf("FilesProcessed = %d, want 1 (should skip .git/)", stats.FilesProcessed)
	}
}

func TestIndexAllSkipsVendor(t *testing.T) {
	repoRoot, store, registry, cfg := setupTestEnv(t)

	writeFile(t, repoRoot, "vendor/lib/lib.go", `package lib
func VendorFunc() {}
`)
	writeFile(t, repoRoot, "main.go", `package main
func main() {}
`)

	idx := New(store, registry, nil, cfg, repoRoot)
	stats, err := idx.IndexAll(context.Background(), false, false)
	if err != nil {
		t.Fatalf("IndexAll: %v", err)
	}

	if stats.FilesProcessed != 1 {
		t.Errorf("FilesProcessed = %d, want 1 (should skip vendor/)", stats.FilesProcessed)
	}
}

func TestIndexAllSkipsLargeFiles(t *testing.T) {
	repoRoot, store, registry, cfg := setupTestEnv(t)

	// Set very small max file size
	cfg.MaxFileSize = 50

	writeFile(t, repoRoot, "small.go", `package main
func small() {}
`)
	// Write a file larger than 50 bytes
	writeFile(t, repoRoot, "large.go", `package main

// This file is intentionally larger than the MaxFileSize limit of 50 bytes.
func largeFunction(a, b, c, d, e, f int) (int, error) {
	return 0, nil
}
`)

	idx := New(store, registry, nil, cfg, repoRoot)
	stats, err := idx.IndexAll(context.Background(), false, false)
	if err != nil {
		t.Fatalf("IndexAll: %v", err)
	}

	if stats.FilesProcessed != 1 {
		t.Errorf("FilesProcessed = %d, want 1 (should skip large file)", stats.FilesProcessed)
	}
}

func TestIndexAllRelativePaths(t *testing.T) {
	repoRoot, store, registry, cfg := setupTestEnv(t)

	writeFile(t, repoRoot, "pkg/handler.go", `package pkg
func Handle() {}
`)

	idx := New(store, registry, nil, cfg, repoRoot)
	_, err := idx.IndexAll(context.Background(), false, false)
	if err != nil {
		t.Fatalf("IndexAll: %v", err)
	}

	var filePath string
	err = store.DB().QueryRow("SELECT file_path FROM code_index LIMIT 1").Scan(&filePath)
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	if filepath.IsAbs(filePath) {
		t.Errorf("file_path is absolute: %s, want relative", filePath)
	}
	if filePath != filepath.Join("pkg", "handler.go") {
		t.Errorf("file_path = %q, want %q", filePath, filepath.Join("pkg", "handler.go"))
	}
}

func TestIndexAllIncremental(t *testing.T) {
	repoRoot, store, registry, cfg := setupTestEnv(t)

	writeFile(t, repoRoot, "main.go", `package main
func main() {}
`)

	idx := New(store, registry, nil, cfg, repoRoot)

	// First indexing
	stats1, err := idx.IndexAll(context.Background(), false, false)
	if err != nil {
		t.Fatalf("IndexAll (1st): %v", err)
	}
	if stats1.FilesProcessed == 0 {
		t.Fatal("first run should process files")
	}

	// Second indexing — file unchanged, should be skipped
	stats2, err := idx.IndexAll(context.Background(), false, false)
	if err != nil {
		t.Fatalf("IndexAll (2nd): %v", err)
	}
	if stats2.FilesProcessed != 0 {
		t.Errorf("second run FilesProcessed = %d, want 0 (unchanged)", stats2.FilesProcessed)
	}
	if stats2.FilesSkipped == 0 {
		t.Error("second run should have skipped files")
	}
}

func TestIndexAllForce(t *testing.T) {
	repoRoot, store, registry, cfg := setupTestEnv(t)

	writeFile(t, repoRoot, "main.go", `package main
func main() {}
`)

	idx := New(store, registry, nil, cfg, repoRoot)

	// First indexing
	_, err := idx.IndexAll(context.Background(), false, false)
	if err != nil {
		t.Fatalf("IndexAll (1st): %v", err)
	}

	// Force re-index — should process all files again
	stats2, err := idx.IndexAll(context.Background(), true, false)
	if err != nil {
		t.Fatalf("IndexAll (force): %v", err)
	}
	if stats2.FilesProcessed == 0 {
		t.Error("force re-index should process files")
	}
}

func TestIndexFile(t *testing.T) {
	repoRoot, store, registry, cfg := setupTestEnv(t)

	source := []byte(`package main

func Hello() string {
	return "hello"
}

type Server struct {
	Port int
}
`)
	idx := New(store, registry, nil, cfg, repoRoot)
	count, err := idx.IndexFile(context.Background(), "main.go", source)
	if err != nil {
		t.Fatalf("IndexFile: %v", err)
	}
	if count == 0 {
		t.Error("IndexFile returned 0 symbols")
	}

	// Verify in database
	var dbCount int
	err = store.DB().QueryRow("SELECT COUNT(*) FROM code_index WHERE file_path = 'main.go'").Scan(&dbCount)
	if err != nil {
		t.Fatalf("count query: %v", err)
	}
	if dbCount != count {
		t.Errorf("db count = %d, IndexFile returned %d", dbCount, count)
	}
}
