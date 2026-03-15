package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// openDatabase detects the repo root, constructs the database path,
// and opens the storage layer. Returns the store, repo root, and db path.
func openDatabase() (*storage.Store, string, string, error) {
	repoRoot, err := detectRepoRoot()
	if err != nil {
		return nil, "", "", fmt.Errorf("could not detect repository root: %w", err)
	}

	dbDir := databaseDir(repoRoot)
	dbPath := filepath.Join(dbDir, "engram.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, repoRoot, dbPath, fmt.Errorf("no Engram database found. Run 'engram index' first")
	}

	store, err := storage.Open(dbPath)
	if err != nil {
		return nil, repoRoot, dbPath, fmt.Errorf("failed to open database at %s: %w", dbPath, err)
	}

	return store, repoRoot, dbPath, nil
}

// formatSize formats a byte count as a human-readable string.
func formatSize(bytes int64) string {
	switch {
	case bytes >= 1 << 20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1 << 10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
