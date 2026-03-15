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

var (
	indexForce   bool
	indexVerbose bool
)

func newIndexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Index the repository for code search",
		Long:  "Walk the repository, parse source files, generate embeddings, and build the search index.",
		RunE:  runIndex,
	}
	cmd.Flags().BoolVar(&indexForce, "force", false, "Drop and rebuild the entire index")
	cmd.Flags().BoolVarP(&indexVerbose, "verbose", "v", false, "Print each file as it's processed")
	return cmd
}

func runIndex(cmd *cobra.Command, args []string) error {
	// 1. Detect repo root
	repoRoot, err := detectRepoRoot()
	if err != nil {
		return fmt.Errorf("could not detect repository root: %w", err)
	}

	// 2. Load config
	cfg, err := config.Load(repoRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// 3. Open storage
	store, err := storage.Open(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer store.Close()

	// 4. Create embedder (graceful degradation)
	var emb *embeddings.Embedder
	emb, err = embeddings.New(cfg.EmbeddingModel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: embeddings unavailable (%v)\n", err)
		emb = nil
	}

	// 5. Create parser registry
	registry := parser.NewDefaultRegistry()

	// 6. Run the indexer
	fmt.Fprintf(os.Stderr, "Engram indexing %s...\n", repoRoot)

	idx := indexer.New(store, registry, emb, cfg, repoRoot)
	stats, err := idx.IndexAll(context.Background(), indexForce, indexVerbose)
	if err != nil {
		return fmt.Errorf("indexing failed: %w", err)
	}

	// 7. Print summary
	fmt.Fprintf(os.Stderr, "\nIndex complete:\n")
	fmt.Fprintf(os.Stderr, "  Files processed:     %d\n", stats.FilesProcessed)
	fmt.Fprintf(os.Stderr, "  Files skipped:       %d\n", stats.FilesSkipped)
	fmt.Fprintf(os.Stderr, "  Symbols extracted:   %d\n", stats.SymbolsExtracted)
	if emb != nil {
		fmt.Fprintf(os.Stderr, "  Embeddings:          %d\n", stats.EmbeddingsGenerated)
	} else {
		fmt.Fprintf(os.Stderr, "  Embeddings:          unavailable — install ONNX model for semantic search\n")
	}
	fmt.Fprintf(os.Stderr, "  Duration:            %s\n", stats.Duration.Round(100*1e6))
	fmt.Fprintf(os.Stderr, "  Database:            %s\n", cfg.DatabasePath)

	if len(stats.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "  Errors:              %d\n", len(stats.Errors))
		for _, e := range stats.Errors {
			fmt.Fprintf(os.Stderr, "    - %s\n", e)
		}
	}

	return nil
}
