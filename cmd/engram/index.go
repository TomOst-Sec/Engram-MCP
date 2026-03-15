package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/TomOst-Sec/colony-project/internal/cli"
	"github.com/TomOst-Sec/colony-project/internal/config"
	"github.com/TomOst-Sec/colony-project/internal/conventions"
	"github.com/TomOst-Sec/colony-project/internal/embeddings"
	"github.com/TomOst-Sec/colony-project/internal/git"
	"github.com/TomOst-Sec/colony-project/internal/indexer"
	"github.com/TomOst-Sec/colony-project/internal/parser"
	"github.com/TomOst-Sec/colony-project/internal/storage"
	"github.com/spf13/cobra"
)

var (
	indexForce   bool
	indexVerbose bool
	indexWatch   bool
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
	cmd.Flags().BoolVar(&indexWatch, "watch", false, "Watch for file changes and re-index automatically")
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
	fmt.Fprintln(os.Stderr, cli.Title.Render(fmt.Sprintf("Engram indexing %s...", repoRoot)))

	idx := indexer.New(store, registry, emb, cfg, repoRoot)
	stats, err := idx.IndexAll(context.Background(), indexForce, indexVerbose)
	if err != nil {
		return fmt.Errorf("indexing failed: %w", err)
	}

	// 7. Analyze git history
	ctx := context.Background()
	fmt.Fprintln(os.Stderr, "Analyzing git history...")
	gitAnalyzer := git.New(store, repoRoot)
	historyStats, historyErr := gitAnalyzer.AnalyzeAll(ctx)
	if historyErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: git history analysis failed: %v\n", historyErr)
	} else {
		fmt.Fprintf(os.Stderr, "Git history: %d files analyzed, hottest: %s (%d changes)\n",
			historyStats.FilesAnalyzed, historyStats.HottestFile, historyStats.HottestFrequency)
	}

	// 8. Analyze conventions
	fmt.Fprintln(os.Stderr, "Detecting conventions...")
	convAnalyzer := conventions.New(store, repoRoot)
	convResult, convErr := convAnalyzer.Analyze(ctx)
	if convErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: convention analysis failed: %v\n", convErr)
	} else {
		fmt.Fprintf(os.Stderr, "Conventions: %d patterns detected\n", len(convResult.Conventions))
	}

	// 9. Print summary
	fmt.Fprintf(os.Stderr, "\n%s\n", cli.SuccessText.Render("Index complete:"))
	fmt.Fprintf(os.Stderr, "  Files processed:     %d\n", stats.FilesProcessed)
	fmt.Fprintf(os.Stderr, "  Files skipped:       %d\n", stats.FilesSkipped)
	fmt.Fprintf(os.Stderr, "  Symbols extracted:   %d\n", stats.SymbolsExtracted)
	if emb != nil {
		fmt.Fprintf(os.Stderr, "  Embeddings:          %d\n", stats.EmbeddingsGenerated)
	} else {
		fmt.Fprintf(os.Stderr, "  Embeddings:          unavailable — install ONNX model for semantic search\n")
	}
	if historyErr == nil {
		fmt.Fprintf(os.Stderr, "  Git history:         %d files analyzed\n", historyStats.FilesAnalyzed)
	}
	if convErr == nil {
		fmt.Fprintf(os.Stderr, "  Conventions:         %d patterns detected\n", len(convResult.Conventions))
	}
	fmt.Fprintf(os.Stderr, "  Duration:            %s\n", stats.Duration.Round(100*1e6))
	fmt.Fprintf(os.Stderr, "  Database:            %s\n", cfg.DatabasePath)

	if len(stats.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "  Errors:              %d\n", len(stats.Errors))
		for _, e := range stats.Errors {
			fmt.Fprintf(os.Stderr, "    - %s\n", e)
		}
	}

	// 10. Start watcher if --watch flag is set
	if indexWatch {
		fmt.Fprintln(os.Stderr, "\nWatching for file changes (Ctrl+C to stop)...")
		w := indexer.NewWatcher(idx, repoRoot, cfg)
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		if err := w.Start(ctx); err != nil && ctx.Err() == nil {
			return fmt.Errorf("watcher failed: %w", err)
		}
		fmt.Fprintln(os.Stderr, "Watcher stopped.")
	}

	return nil
}
