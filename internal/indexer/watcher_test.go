package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TomOst-Sec/colony-project/internal/config"
	"github.com/TomOst-Sec/colony-project/internal/parser"
	"github.com/TomOst-Sec/colony-project/internal/storage"
)

func setupWatcherTest(t *testing.T) (string, *storage.Store, *Indexer, *config.Config) {
	t.Helper()
	repoRoot := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("open storage: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	registry := parser.NewDefaultRegistry()
	cfg := config.Default()
	cfg.RepoRoot = repoRoot
	idx := New(store, registry, nil, cfg, repoRoot)
	return repoRoot, store, idx, cfg
}

func TestNewWatcher(t *testing.T) {
	_, _, idx, cfg := setupWatcherTest(t)
	w := NewWatcher(idx, cfg.RepoRoot, cfg)
	if w == nil {
		t.Fatal("NewWatcher returned nil")
	}
}

func TestWatcherDetectsFileCreate(t *testing.T) {
	repoRoot, store, idx, cfg := setupWatcherTest(t)
	w := NewWatcher(idx, repoRoot, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start watcher in background
	errCh := make(chan error, 1)
	go func() { errCh <- w.Start(ctx) }()

	// Give watcher time to start
	time.Sleep(200 * time.Millisecond)

	// Create a Go file
	goFile := filepath.Join(repoRoot, "main.go")
	if err := os.WriteFile(goFile, []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Wait for debounce + indexing
	time.Sleep(500 * time.Millisecond)

	// Verify symbols were indexed
	var count int
	err := store.DB().QueryRow("SELECT COUNT(*) FROM code_index WHERE file_path = 'main.go'").Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count == 0 {
		t.Error("watcher did not index created file")
	}

	cancel()
}

func TestWatcherDetectsFileModification(t *testing.T) {
	repoRoot, store, idx, cfg := setupWatcherTest(t)

	// Create file before watcher starts
	goFile := filepath.Join(repoRoot, "lib.go")
	if err := os.WriteFile(goFile, []byte("package lib\nfunc Original() {}\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Index it first
	source, _ := os.ReadFile(goFile)
	if _, err := idx.IndexFile(context.Background(), "lib.go", source); err != nil {
		t.Fatalf("initial index: %v", err)
	}

	w := NewWatcher(idx, repoRoot, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { w.Start(ctx) }()
	time.Sleep(200 * time.Millisecond)

	// Modify the file
	if err := os.WriteFile(goFile, []byte("package lib\nfunc Modified() {}\nfunc Another() {}\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Check that symbols were updated
	var name string
	err := store.DB().QueryRow("SELECT symbol_name FROM code_index WHERE file_path = 'lib.go' AND symbol_type != 'import' LIMIT 1").Scan(&name)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if name == "Original" {
		t.Error("watcher did not re-index modified file — still shows old symbol")
	}

	cancel()
}

func TestWatcherDetectsFileDeletion(t *testing.T) {
	repoRoot, store, idx, cfg := setupWatcherTest(t)

	// Create and index a file
	goFile := filepath.Join(repoRoot, "delete_me.go")
	if err := os.WriteFile(goFile, []byte("package main\nfunc DeleteMe() {}\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	source, _ := os.ReadFile(goFile)
	if _, err := idx.IndexFile(context.Background(), "delete_me.go", source); err != nil {
		t.Fatalf("index: %v", err)
	}

	w := NewWatcher(idx, repoRoot, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { w.Start(ctx) }()
	time.Sleep(200 * time.Millisecond)

	// Delete the file
	os.Remove(goFile)
	time.Sleep(500 * time.Millisecond)

	// Verify symbols were removed
	var count int
	err := store.DB().QueryRow("SELECT COUNT(*) FROM code_index WHERE file_path = 'delete_me.go'").Scan(&count)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 0 {
		t.Errorf("watcher did not remove symbols on file delete, count=%d", count)
	}

	cancel()
}

func TestWatcherIgnoresGitDir(t *testing.T) {
	repoRoot, store, idx, cfg := setupWatcherTest(t)

	// Create .git directory
	gitDir := filepath.Join(repoRoot, ".git")
	os.MkdirAll(gitDir, 0755)

	w := NewWatcher(idx, repoRoot, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { w.Start(ctx) }()
	time.Sleep(200 * time.Millisecond)

	// Create a Go file inside .git
	gitFile := filepath.Join(gitDir, "hooks.go")
	os.WriteFile(gitFile, []byte("package hooks\nfunc PreCommit() {}\n"), 0644)
	time.Sleep(500 * time.Millisecond)

	// Verify nothing was indexed
	var count int
	store.DB().QueryRow("SELECT COUNT(*) FROM code_index").Scan(&count)
	if count != 0 {
		t.Errorf("watcher indexed file in .git directory, count=%d", count)
	}

	cancel()
}

func TestWatcherIgnoresUnsupportedExtensions(t *testing.T) {
	repoRoot, store, idx, cfg := setupWatcherTest(t)
	w := NewWatcher(idx, repoRoot, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { w.Start(ctx) }()
	time.Sleep(200 * time.Millisecond)

	// Create a .txt file (no parser registered)
	txtFile := filepath.Join(repoRoot, "notes.txt")
	os.WriteFile(txtFile, []byte("just notes"), 0644)
	time.Sleep(500 * time.Millisecond)

	var count int
	store.DB().QueryRow("SELECT COUNT(*) FROM code_index").Scan(&count)
	if count != 0 {
		t.Errorf("watcher indexed unsupported file extension, count=%d", count)
	}

	cancel()
}

func TestWatcherStopClean(t *testing.T) {
	_, _, idx, cfg := setupWatcherTest(t)
	w := NewWatcher(idx, cfg.RepoRoot, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- w.Start(ctx) }()

	time.Sleep(100 * time.Millisecond)
	w.Stop()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Stop returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("watcher did not stop within 2 seconds")
	}
}
