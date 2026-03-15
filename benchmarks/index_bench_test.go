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

func BenchmarkIndexSmallRepo(b *testing.B) {
	repoDir := b.TempDir()
	if err := testdata.GenerateRepo(repoDir, 20); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		dbPath := filepath.Join(b.TempDir(), "bench.db")
		store, err := storage.Open(dbPath)
		if err != nil {
			b.Fatal(err)
		}
		registry := parser.NewDefaultRegistry()
		cfg := config.Default()
		idx := indexer.New(store, registry, nil, cfg, repoDir)
		_, err = idx.IndexAll(context.Background(), true, false)
		if err != nil {
			b.Fatal(err)
		}
		store.Close()
	}
}

func BenchmarkIndexMediumRepo(b *testing.B) {
	repoDir := b.TempDir()
	if err := testdata.GenerateRepo(repoDir, 100); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		dbPath := filepath.Join(b.TempDir(), "bench.db")
		store, err := storage.Open(dbPath)
		if err != nil {
			b.Fatal(err)
		}
		registry := parser.NewDefaultRegistry()
		cfg := config.Default()
		idx := indexer.New(store, registry, nil, cfg, repoDir)
		_, err = idx.IndexAll(context.Background(), true, false)
		if err != nil {
			b.Fatal(err)
		}
		store.Close()
	}
}

func BenchmarkIncrementalIndex(b *testing.B) {
	repoDir := b.TempDir()
	if err := testdata.GenerateRepo(repoDir, 50); err != nil {
		b.Fatal(err)
	}

	dbPath := filepath.Join(b.TempDir(), "bench.db")
	store, err := storage.Open(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

	registry := parser.NewDefaultRegistry()
	cfg := config.Default()
	idx := indexer.New(store, registry, nil, cfg, repoDir)

	// Full index first
	_, err = idx.IndexAll(context.Background(), true, false)
	if err != nil {
		b.Fatal(err)
	}

	// Benchmark single file re-index
	source := []byte("package test\nfunc Updated() {}\n")
	b.ResetTimer()
	for b.Loop() {
		idx.IndexFile(context.Background(), "pkg/module0/handler0.go", source)
	}
}
