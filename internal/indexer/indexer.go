package indexer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TomOst-Sec/colony-project/internal/config"
	"github.com/TomOst-Sec/colony-project/internal/embeddings"
	"github.com/TomOst-Sec/colony-project/internal/parser"
	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// Indexer walks a repository, parses source files, and stores symbols in the database.
type Indexer struct {
	store    *storage.Store
	registry *parser.Registry
	embedder *embeddings.Embedder
	config   *config.Config
	repoRoot string
}

// IndexStats holds statistics from an indexing run.
type IndexStats struct {
	FilesProcessed      int
	FilesSkipped        int
	SymbolsExtracted    int
	EmbeddingsGenerated int
	Duration            time.Duration
	Errors              []string
}

// New creates a new Indexer.
func New(store *storage.Store, registry *parser.Registry, embedder *embeddings.Embedder, cfg *config.Config, repoRoot string) *Indexer {
	return &Indexer{
		store:    store,
		registry: registry,
		embedder: embedder,
		config:   cfg,
		repoRoot: repoRoot,
	}
}

// IndexAll performs a full repository index.
// If force is true, clears existing index data first.
func (idx *Indexer) IndexAll(ctx context.Context, force bool, verbose bool) (*IndexStats, error) {
	start := time.Now()
	stats := &IndexStats{}

	if force {
		if _, err := idx.store.DB().ExecContext(ctx, "DELETE FROM code_index"); err != nil {
			return nil, fmt.Errorf("clearing code_index: %w", err)
		}
	}

	err := filepath.WalkDir(idx.repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("walk error %s: %v", path, err))
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Skip directories
		if d.IsDir() {
			name := d.Name()
			// Skip hidden directories
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			// Skip directories matching ignore patterns
			for _, pattern := range idx.config.IgnorePatterns {
				trimmed := strings.TrimSuffix(pattern, "/")
				if name == trimmed {
					return filepath.SkipDir
				}
			}
			return nil
		}

		relPath, err := filepath.Rel(idx.repoRoot, path)
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("rel path %s: %v", path, err))
			return nil
		}

		// Skip files with no registered parser
		if _, ok := idx.registry.ParserFor(relPath); !ok {
			stats.FilesSkipped++
			return nil
		}

		// Skip files exceeding max size
		info, err := d.Info()
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("stat %s: %v", relPath, err))
			return nil
		}
		if info.Size() > idx.config.MaxFileSize {
			stats.FilesSkipped++
			return nil
		}

		// Read file
		source, err := os.ReadFile(path)
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("read %s: %v", relPath, err))
			return nil
		}

		// Compute hash for incremental indexing
		hash := fmt.Sprintf("%x", sha256.Sum256(source))

		// Check if file is unchanged
		existingHash, hashErr := parser.GetFileHash(idx.store, relPath)
		if hashErr == nil && existingHash == hash {
			stats.FilesSkipped++
			return nil
		}

		// Index the file
		if verbose {
			fmt.Fprintf(os.Stderr, "Indexing %s...\n", relPath)
		}

		symbolCount, err := idx.IndexFile(ctx, relPath, source)
		if err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("index %s: %v", relPath, err))
			return nil
		}

		stats.FilesProcessed++
		stats.SymbolsExtracted += symbolCount
		return nil
	})

	if err != nil {
		return stats, fmt.Errorf("walking repository: %w", err)
	}

	stats.Duration = time.Since(start)
	return stats, nil
}

// IndexFile indexes a single file. Returns the number of symbols extracted.
func (idx *Indexer) IndexFile(ctx context.Context, filePath string, source []byte) (int, error) {
	// Delete old symbols for this file
	if err := parser.DeleteFileSymbols(idx.store, filePath); err != nil {
		return 0, fmt.Errorf("deleting old symbols for %s: %w", filePath, err)
	}

	// Parse file
	symbols, err := idx.registry.ParseFile(filePath, source)
	if err != nil {
		return 0, fmt.Errorf("parsing %s: %w", filePath, err)
	}

	if len(symbols) == 0 {
		return 0, nil
	}

	// Compute file hash
	fileHash := fmt.Sprintf("%x", sha256.Sum256(source))

	// Store symbols
	if err := parser.StoreSymbols(idx.store, fileHash, symbols); err != nil {
		return 0, fmt.Errorf("storing symbols for %s: %w", filePath, err)
	}

	// Generate embeddings if embedder is available
	if idx.embedder != nil {
		if err := idx.generateEmbeddings(ctx, filePath, symbols); err != nil {
			// Log but don't fail — embeddings are best-effort
			fmt.Fprintf(os.Stderr, "Warning: embedding generation failed for %s: %v\n", filePath, err)
		}
	}

	return len(symbols), nil
}

func (idx *Indexer) generateEmbeddings(ctx context.Context, filePath string, symbols []parser.Symbol) error {
	// Build texts for embedding
	texts := make([]string, len(symbols))
	for i, s := range symbols {
		texts[i] = s.Name + " " + s.Signature + " " + s.Docstring
	}

	batchSize := idx.config.EmbeddingBatchSize
	if batchSize <= 0 {
		batchSize = 32
	}

	vectors, err := idx.embedder.EmbedBatch(texts, batchSize)
	if err != nil {
		return fmt.Errorf("embedding batch: %w", err)
	}

	// Look up symbol IDs and update embeddings
	rows, err := idx.store.DB().QueryContext(ctx,
		"SELECT id FROM code_index WHERE file_path = ? ORDER BY id",
		filePath,
	)
	if err != nil {
		return fmt.Errorf("querying symbol IDs: %w", err)
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		if i >= len(vectors) {
			break
		}
		var id int64
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("scanning symbol ID: %w", err)
		}
		if err := embeddings.UpdateCodeIndexEmbedding(idx.store, id, vectors[i]); err != nil {
			return fmt.Errorf("updating embedding for symbol %d: %w", id, err)
		}
		i++
	}

	return rows.Err()
}
