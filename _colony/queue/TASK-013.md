# TASK-013: `engram index` CLI Command — Full Repository Indexer

**Priority:** P0
**Assigned:** alpha
**Milestone:** M1: MVP
**Dependencies:** TASK-003, TASK-005, TASK-006
**Status:** queued
**Created:** 2026-03-15
**Author:** atlas

## Context
The `engram index` command is how users build the code intelligence database. It walks the entire repository, parses every supported source file through tree-sitter, generates embeddings for extracted symbols, and stores everything in SQLite. This is the command that must complete before `engram serve` becomes useful — without an index, search_code and get_architecture return empty results. Target: full index of a 50K LOC repo in <30 seconds.

## Specification
Create the `engram index` CLI subcommand and the `internal/indexer` package.

### CLI Command: `cmd/engram/index.go`
```go
var indexCmd = &cobra.Command{
    Use:   "index",
    Short: "Index the repository for code search",
    Long:  "Walk the repository, parse source files, generate embeddings, and build the search index.",
    RunE:  runIndex,
}
```

**Flags:**
- `--force` — Drop and rebuild the entire index (delete all code_index rows, re-parse everything)
- `--verbose` / `-v` — Print each file as it's processed

**Behavior:**
1. Detect repo root (same as serve command — reuse `detectRepoRoot()`)
2. Load config
3. Open storage
4. Create embedder (or NoOp)
5. Create parser registry with all available grammars
6. Run the indexer
7. Print summary statistics

### Indexer Package: `internal/indexer/indexer.go`

```go
type Indexer struct {
    store    *storage.Store
    registry *parser.Registry
    embedder *embeddings.Embedder  // may be nil
    config   *config.Config
    repoRoot string
}

type IndexStats struct {
    FilesProcessed  int
    FilesSkipped    int
    SymbolsExtracted int
    EmbeddingsGenerated int
    Duration        time.Duration
    Errors          []string
}

func New(store *storage.Store, registry *parser.Registry, embedder *embeddings.Embedder, cfg *config.Config, repoRoot string) *Indexer

// IndexAll performs a full repository index.
// If force is true, clears existing index data first.
func (idx *Indexer) IndexAll(ctx context.Context, force bool, verbose bool) (*IndexStats, error)

// IndexFile indexes a single file. Used by IndexAll and future --watch mode.
func (idx *Indexer) IndexFile(ctx context.Context, filePath string, source []byte) (int, error)  // returns symbol count
```

### IndexAll Algorithm

```
1. If force: DELETE FROM code_index; DELETE FROM code_index_fts (rebuild triggers handle FTS)
2. Walk repo root recursively:
   a. Skip directories matching config.IgnorePatterns (vendor/, node_modules/, .git/, bin/, dist/)
   b. Skip files larger than config.MaxFileSize (default 1MB)
   c. Skip files with no registered parser (unknown extension)
   d. For each supported file:
      i.   Read file contents
      ii.  Compute SHA256 hash of contents
      iii. Check if file_hash in code_index matches current hash
      iv.  If hash matches (unchanged): skip (incremental optimization)
      v.   If hash differs or file is new:
           - Delete old symbols for this file_path: parser.DeleteFileSymbols(store, filePath)
           - Parse file: registry.ParseFile(filePath, source) → []Symbol
           - Store symbols: parser.StoreSymbols(store, fileHash, symbols)
           - Generate embeddings for each symbol (if embedder available):
             * Text = symbol_name + " " + signature + " " + docstring
             * embedder.Embed(text) → vector
             * embeddings.UpdateCodeIndexEmbedding(store, symbolID, vector)
           - Track stats
3. Print summary to stderr
```

### File Walking Rules
- Use `filepath.WalkDir` (Go 1.16+, more efficient than `filepath.Walk`)
- Skip directories that start with `.` (hidden dirs)
- Skip directories matching any config.IgnorePatterns using `filepath.Match`
- All file paths stored relative to repoRoot
- Use `filepath.Rel(repoRoot, absPath)` for relative paths

### Output Format (to stderr)
```
Engram indexing /path/to/repo...
Scanning files... found 342 supported files (Go: 120, Python: 85, TypeScript: 137)
Indexing... [=============>      ] 67% (229/342)
Generating embeddings... [========>           ] 42% (1,203/2,847 symbols)

Index complete:
  Files processed:     342
  Files skipped:       1,204 (unsupported: 890, unchanged: 314)
  Symbols extracted:   2,847
  Embeddings:          2,847 (or "unavailable — install ONNX model for semantic search")
  Duration:            12.3s
  Database:            ~/.engram/a1b2c3d4e5f6/engram.db (4.2 MB)
```

Progress bars are optional for MVP — a simple counter per 100 files is sufficient.

## Acceptance Criteria
- [ ] `engram index` walks the repository and indexes all supported source files
- [ ] `engram index` skips files matching ignore patterns (vendor/, node_modules/, .git/)
- [ ] `engram index` skips files exceeding max file size
- [ ] `engram index` uses file hashes for incremental re-indexing (unchanged files are skipped)
- [ ] `engram index --force` rebuilds the entire index from scratch
- [ ] `engram index` generates embeddings for each symbol (when embedder is available)
- [ ] `engram index` prints summary statistics to stderr
- [ ] `engram index` stores file paths relative to repo root
- [ ] After indexing, `search_code` tool returns results from the indexed files
- [ ] `engram index` completes without error on this project's codebase
- [ ] All tests pass

## Implementation Steps
1. Create `internal/indexer/indexer.go` — Indexer struct, New, IndexAll, IndexFile
2. Create `cmd/engram/index.go` — index subcommand, runIndex function, --force and --verbose flags
3. Register index command in `cmd/engram/main.go`: `cmd.AddCommand(newIndexCmd())`
4. Create `internal/indexer/indexer_test.go`:
   - Test: IndexAll on a temp directory with sample Go/Python files produces correct symbol count
   - Test: Incremental indexing skips unchanged files (index twice, second run processes 0 files)
   - Test: --force flag causes full re-index (all files processed on second run)
   - Test: Ignore patterns skip vendor/ and node_modules/
   - Test: Large files are skipped
   - Test: File paths are stored relative to repo root
5. Create `cmd/engram/index_test.go`:
   - Test: index command is registered on root command
   - Test: --help output describes indexing
6. Run all tests

## Testing Requirements
- Unit test: IndexAll on temp dir with 5 Go files extracts expected symbol count
- Unit test: IndexAll skips .git/ directory
- Unit test: IndexAll skips files > MaxFileSize
- Unit test: IndexFile parses a Go file and stores symbols in code_index table
- Unit test: Incremental: index twice, second run returns FilesSkipped > 0 and FilesProcessed = 0
- Unit test: Force flag: index twice with force=true, second run has FilesProcessed > 0
- Unit test: Relative paths stored (no absolute paths in code_index.file_path)

## Files to Create/Modify
- `internal/indexer/indexer.go` — Indexer struct, New, IndexAll, IndexFile
- `internal/indexer/indexer_test.go` — indexer tests
- `cmd/engram/index.go` — index CLI subcommand
- `cmd/engram/index_test.go` — CLI tests
- `cmd/engram/main.go` — add newIndexCmd() to root command (one line)

## Notes
- Reuse `detectRepoRoot()` from serve.go. If it's not exported, either export it to a shared location or duplicate it in index.go (simple function, duplication is acceptable for now).
- The parser registry needs to be populated with all available parsers. Create a helper like `parser.NewDefaultRegistry()` that registers Go, Python, TypeScript, JavaScript parsers. If this doesn't exist yet, add it to `internal/parser/registry.go`.
- Embedding generation is the slowest part. Use EmbedBatch if available for better throughput. Process symbols in batches of 32 (config.EmbeddingBatchSize).
- For the FTS5 content sync, symbols stored via `parser.StoreSymbols` should automatically populate the FTS index via SQLite triggers (set up in TASK-003). Verify this works.
- Do NOT implement progress bars with terminal escape codes for MVP. Simple line-by-line output is fine: `fmt.Fprintf(os.Stderr, "Indexing %s (%d symbols)...\n", filePath, len(symbols))`
- The indexer should be robust against individual file failures — log the error, skip the file, continue with the next one. Don't abort the entire index because one file has a parse error.
