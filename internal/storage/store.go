package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Store wraps a SQLite database connection with schema management.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) a SQLite database at dbPath, applies any pending
// migrations, and configures pragmas for optimal performance.
// Parent directories are created automatically if they don't exist.
func Open(dbPath string) (*Store, error) {
	// Create parent directories
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating directory %s: %w", dir, err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	// Configure pragmas
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
		"PRAGMA auto_vacuum=INCREMENTAL",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("setting pragma %q: %w", p, err)
		}
	}

	// Apply migrations
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return &Store{db: db}, nil
}

// Close runs incremental vacuum and closes the database connection.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	// Run incremental vacuum before closing
	_, _ = s.db.Exec("PRAGMA incremental_vacuum")
	return s.db.Close()
}

// DB returns the underlying *sql.DB for advanced queries.
func (s *Store) DB() *sql.DB {
	return s.db
}
