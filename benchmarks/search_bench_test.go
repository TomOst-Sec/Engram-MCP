package benchmarks

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/TomOst-Sec/colony-project/benchmarks/testdata"
	"github.com/TomOst-Sec/colony-project/internal/config"
	"github.com/TomOst-Sec/colony-project/internal/indexer"
	"github.com/TomOst-Sec/colony-project/internal/parser"
	"github.com/TomOst-Sec/colony-project/internal/storage"
)

func setupIndexedStore(b *testing.B, fileCount int) *storage.Store {
	b.Helper()

	repoDir := b.TempDir()
	if err := testdata.GenerateRepo(repoDir, fileCount); err != nil {
		b.Fatal(err)
	}

	dbPath := filepath.Join(b.TempDir(), "bench.db")
	store, err := storage.Open(dbPath)
	if err != nil {
		b.Fatal(err)
	}

	registry := parser.NewDefaultRegistry()
	cfg := config.Default()
	idx := indexer.New(store, registry, nil, cfg, repoDir)
	if _, err := idx.IndexAll(context.Background(), true, false); err != nil {
		b.Fatal(err)
	}

	return store
}

func BenchmarkSearchFTS5(b *testing.B) {
	store := setupIndexedStore(b, 100)
	defer store.Close()

	b.ResetTimer()
	for b.Loop() {
		rows, err := store.DB().Query(
			`SELECT ci.symbol_name, ci.file_path, rank
			 FROM code_index_fts fts
			 JOIN code_index ci ON ci.id = fts.rowid
			 WHERE code_index_fts MATCH ?
			 ORDER BY rank
			 LIMIT 10`, "handler")
		if err != nil {
			b.Fatal(err)
		}
		for rows.Next() {
			var name, path string
			var rank float64
			rows.Scan(&name, &path, &rank)
		}
		rows.Close()
	}
}

func BenchmarkSearchFTS5Cached(b *testing.B) {
	store := setupIndexedStore(b, 100)
	defer store.Close()

	// Warm the cache with a first query
	store.DB().Query(
		`SELECT ci.symbol_name FROM code_index_fts fts JOIN code_index ci ON ci.id = fts.rowid WHERE code_index_fts MATCH ? LIMIT 10`,
		"service",
	)

	b.ResetTimer()
	for b.Loop() {
		rows, err := store.DB().Query(
			`SELECT ci.symbol_name, ci.file_path, rank
			 FROM code_index_fts fts
			 JOIN code_index ci ON ci.id = fts.rowid
			 WHERE code_index_fts MATCH ?
			 ORDER BY rank
			 LIMIT 10`, "service")
		if err != nil {
			b.Fatal(err)
		}
		for rows.Next() {
			var name, path string
			var rank float64
			rows.Scan(&name, &path, &rank)
		}
		rows.Close()
	}
}
