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

	// Post-processing: refine references across all files
	if err := idx.RefineReferences(ctx, verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: reference refinement failed: %v\n", err)
	}

	stats.Duration = time.Since(start)
	return stats, nil
}

// RefineReferences runs cross-file post-processing on references:
// 1. Type-aware filtering: remove callback_arg entries that aren't functions
// 2. Parameter flow tracking: link call arguments to callee parameter names
// 3. Indirect call resolution: link pointer_assign + local call to actual function
func (idx *Indexer) RefineReferences(ctx context.Context, verbose bool) error {
	db := idx.store.DB()

	// === Improvement 1: Type-aware filtering ===
	// Delete callback_arg refs where to_name is not a known function symbol
	if _, err := db.ExecContext(ctx, `
		DELETE FROM code_references
		WHERE kind = 'callback_arg'
		AND to_name NOT IN (SELECT DISTINCT symbol_name FROM code_index WHERE symbol_type = 'function')
	`); err != nil {
		return fmt.Errorf("type-aware filtering: %w", err)
	}

	// === Improvement 2: Parameter flow tracking ===
	// For each callback_arg (now only functions), find the callee's signature
	// and match argument position to parameter name
	if err := idx.resolveParamAliases(ctx); err != nil {
		return fmt.Errorf("param alias resolution: %w", err)
	}

	// === Improvement 3: Indirect call resolution ===
	// For each function scope, find pointer_assign + call pairs
	if err := idx.resolveIndirectCalls(ctx); err != nil {
		return fmt.Errorf("indirect call resolution: %w", err)
	}

	return nil
}

// resolveParamAliases upgrades callback_arg refs to param_alias by matching
// argument position to the callee's parameter name.
func (idx *Indexer) resolveParamAliases(ctx context.Context) error {
	db := idx.store.DB()

	// Get all remaining callback_arg refs (they're all confirmed function names now)
	rows, err := db.QueryContext(ctx, `
		SELECT cr.id, cr.to_name, cr.from_func, cr.file_path, cr.line
		FROM code_references cr WHERE cr.kind = 'callback_arg'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type cbRef struct {
		id       int64
		toName   string
		fromFunc string
		filePath string
		line     int
	}
	var refs []cbRef
	for rows.Next() {
		var r cbRef
		if err := rows.Scan(&r.id, &r.toName, &r.fromFunc, &r.filePath, &r.line); err != nil {
			return err
		}
		refs = append(refs, r)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, r := range refs {
		// Find the call on the same line from the same function to get the callee name
		var calleeName string
		err := db.QueryRowContext(ctx, `
			SELECT to_name FROM code_references
			WHERE kind = 'call' AND from_func = ? AND line = ? AND file_path = ?
			LIMIT 1`,
			r.fromFunc, r.line, r.filePath,
		).Scan(&calleeName)
		if err != nil {
			continue
		}

		// Look up callee's signature
		var signature string
		err = db.QueryRowContext(ctx, `
			SELECT signature FROM code_index
			WHERE symbol_name = ? AND symbol_type = 'function' LIMIT 1`,
			calleeName,
		).Scan(&signature)
		if err != nil {
			continue
		}

		// Parse parameter names from signature and find the position of this argument
		params := parseParamNames(signature)
		// Find the argument index: count callback_arg refs on the same line before this one
		var argIndex int
		db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM code_references
			WHERE kind IN ('callback_arg', 'param_alias')
			AND from_func = ? AND line = ? AND file_path = ? AND id < ?`,
			r.fromFunc, r.line, r.filePath, r.id,
		).Scan(&argIndex)

		if argIndex < len(params) {
			db.ExecContext(ctx, `
				UPDATE code_references SET kind = 'param_alias', param_name = ?
				WHERE id = ?`, params[argIndex], r.id)
		}
	}

	return nil
}

// parseParamNames extracts parameter names from a C function signature.
// E.g. "void traced_dispatch(const char *name, math_op *table, int size)" -> ["name", "table", "size"]
func parseParamNames(sig string) []string {
	// Find the parameter list between ( and )
	start := strings.Index(sig, "(")
	end := strings.LastIndex(sig, ")")
	if start < 0 || end <= start {
		return nil
	}
	paramStr := sig[start+1 : end]
	if strings.TrimSpace(paramStr) == "" || strings.TrimSpace(paramStr) == "void" {
		return nil
	}

	parts := strings.Split(paramStr, ",")
	var names []string
	for _, part := range parts {
		name := extractLastIdentifier(strings.TrimSpace(part))
		names = append(names, name)
	}
	return names
}

// extractLastIdentifier gets the parameter name from a C parameter declaration.
// E.g. "const char *name" -> "name", "int size" -> "size", "int (*callback)(int, int)" -> "callback"
func extractLastIdentifier(param string) string {
	// Handle function pointer: int (*callback)(int, int)
	if idx := strings.Index(param, "(*"); idx >= 0 {
		rest := param[idx+2:]
		end := strings.Index(rest, ")")
		if end > 0 {
			return rest[:end]
		}
	}

	// Strip array brackets
	if idx := strings.Index(param, "["); idx >= 0 {
		param = param[:idx]
	}

	// Take the last token, stripping pointer stars
	param = strings.TrimRight(param, " ")
	tokens := strings.Fields(param)
	if len(tokens) == 0 {
		return ""
	}
	last := tokens[len(tokens)-1]
	return strings.TrimLeft(last, "*&")
}

// resolveIndirectCalls finds pointer_assign + call patterns within each function
// and creates indirect_call references linking the call to the actual function.
func (idx *Indexer) resolveIndirectCalls(ctx context.Context) error {
	db := idx.store.DB()

	// Get all functions that have pointer_assign refs
	funcRows, err := db.QueryContext(ctx, `
		SELECT DISTINCT from_func, file_path FROM code_references
		WHERE kind = 'pointer_assign' AND from_func != '<top-level>'
	`)
	if err != nil {
		return err
	}
	defer funcRows.Close()

	type funcKey struct{ name, file string }
	var funcs []funcKey
	for funcRows.Next() {
		var fk funcKey
		if err := funcRows.Scan(&fk.name, &fk.file); err != nil {
			return err
		}
		funcs = append(funcs, fk)
	}
	if err := funcRows.Err(); err != nil {
		return err
	}

	for _, fk := range funcs {
		// Get all refs in this function ordered by line
		rows, err := db.QueryContext(ctx, `
			SELECT id, to_name, kind, line, context FROM code_references
			WHERE from_func = ? AND file_path = ? ORDER BY line`,
			fk.name, fk.file)
		if err != nil {
			continue
		}

		// Build pointer assignment map: localVar -> assignedFuncName
		ptrMap := map[string]string{}
		type ref struct {
			id      int64
			toName  string
			kind    string
			line    int
			context string
		}
		var allRefs []ref

		for rows.Next() {
			var r ref
			if err := rows.Scan(&r.id, &r.toName, &r.kind, &r.line, &r.context); err != nil {
				break
			}
			allRefs = append(allRefs, r)
		}
		rows.Close()

		// Pass 1: collect pointer assignments
		for _, r := range allRefs {
			if r.kind == "pointer_assign" {
				// Find the variable name: look for a call to `r.toName` as
				// the assigned value — the variable is the init_declarator target.
				// Actually: for `chosen = f`, r.toName = "f". We need to find
				// the variable name. Check if there's an init_declarator or
				// assignment on the same line.
				// Heuristic: look at context for `varname = toName` pattern
				varName := extractAssignTarget(r.context, r.toName)
				if varName != "" {
					ptrMap[varName] = r.toName
				}
			}
		}

		// Pass 2: find calls to local pointer variables and resolve them
		for _, r := range allRefs {
			if r.kind == "call" {
				if resolvedFunc, ok := ptrMap[r.toName]; ok {
					// This call is through a function pointer — insert indirect_call ref
					db.ExecContext(ctx, `
						INSERT INTO code_references (file_path, to_name, kind, from_func, line, context)
						VALUES (?, ?, 'indirect_call', ?, ?, ?)`,
						fk.file, resolvedFunc, fk.name, r.line, r.context)
				}
			}
		}
	}

	return nil
}

// extractAssignTarget extracts the variable name from an assignment context line.
// E.g., "binary_op op = f;" -> "op", "chosen = f;" -> "chosen"
func extractAssignTarget(context, assignedValue string) string {
	// Look for pattern: something = assignedValue
	idx := strings.Index(context, "= "+assignedValue)
	if idx < 0 {
		idx = strings.Index(context, "="+assignedValue)
	}
	if idx < 0 {
		return ""
	}
	before := strings.TrimRight(context[:idx], " ")
	tokens := strings.Fields(before)
	if len(tokens) == 0 {
		return ""
	}
	return strings.TrimLeft(tokens[len(tokens)-1], "*&")
}

// IndexFile indexes a single file. Returns the number of symbols extracted.
func (idx *Indexer) IndexFile(ctx context.Context, filePath string, source []byte) (int, error) {
	// Delete old symbols and references for this file
	if err := parser.DeleteFileSymbols(idx.store, filePath); err != nil {
		return 0, fmt.Errorf("deleting old symbols for %s: %w", filePath, err)
	}
	if err := parser.DeleteFileReferences(idx.store, filePath); err != nil {
		return 0, fmt.Errorf("deleting old references for %s: %w", filePath, err)
	}

	// Parse file
	p, ok := idx.registry.ParserFor(filePath)
	if !ok {
		return 0, nil
	}
	symbols, err := p.Parse(filePath, source)
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

	// Extract and store references if parser supports it
	if rp, ok := p.(parser.ReferenceParser); ok {
		refs, err := rp.ParseReferences(filePath, source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: reference extraction failed for %s: %v\n", filePath, err)
		} else if len(refs) > 0 {
			if err := parser.StoreReferences(idx.store, refs); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: storing references failed for %s: %v\n", filePath, err)
			}
		}
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
