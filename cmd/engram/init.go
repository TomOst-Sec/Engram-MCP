package main

import (
	"context"
	"fmt"
	"os"

	"github.com/TomOst-Sec/colony-project/internal/config"
	"github.com/TomOst-Sec/colony-project/internal/embeddings"
	"github.com/TomOst-Sec/colony-project/internal/indexer"
	"github.com/TomOst-Sec/colony-project/internal/parser"
	"github.com/TomOst-Sec/colony-project/internal/storage"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize Engram for this project",
		Long:  "Index the repository and print connection instructions for AI coding tools.",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Fprintf(os.Stderr, "Engram v%s — Persistent memory for AI coding agents\n\n", version)

	// 1. Detect repo root
	repoRoot, err := detectRepoRoot()
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	// 2. Load config
	cfg, err := config.Load(repoRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// 3. Open storage
	store, err := storage.Open(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer store.Close()

	// 4. Create parser registry
	registry := parser.NewDefaultRegistry()

	// 5. Create embedder (optional)
	var emb *embeddings.Embedder
	emb, err = embeddings.New(cfg.EmbeddingModel)
	if err != nil {
		emb = nil
	}

	// 6. Run indexer
	fmt.Fprintf(os.Stderr, "Indexing %s...\n", repoRoot)
	idx := indexer.New(store, registry, emb, cfg, repoRoot)
	stats, err := idx.IndexAll(context.Background(), false, false)
	if err != nil {
		return fmt.Errorf("indexing failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "  %d files indexed, %d symbols extracted\n", stats.FilesProcessed, stats.SymbolsExtracted)
	fmt.Fprintf(os.Stderr, "  Database: %s\n\n", cfg.DatabasePath)

	// 7. Print connection instructions
	fmt.Println("Connect your AI tool:")
	fmt.Println()
	fmt.Println("  Claude Code:  claude mcp add engram -- engram serve")
	fmt.Println("  Cursor:       Add to .cursor/mcp.json (see docs/cursor.md)")
	fmt.Println("  Codex CLI:    codex --mcp-server \"engram serve\"")
	fmt.Println()
	fmt.Println("Ready! Your AI coding tool now has persistent memory.")

	return nil
}
