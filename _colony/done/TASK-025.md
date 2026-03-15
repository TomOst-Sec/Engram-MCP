# TASK-025: `engram index` Git History Integration

**Priority:** P1
**Assigned:** alpha
**Milestone:** M2: Core Features
**Dependencies:** TASK-013, TASK-018
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
TASK-013 created `engram index` and TASK-018 created the git history analyzer. Currently they're disconnected — indexing doesn't analyze git history. This task integrates them: when `engram index` runs, it also analyzes git history and stores results in the `git_context` table. This also adds convention analysis (TASK-019) to the index pipeline so a single `engram index` populates all tables.

## Specification

### Modify `cmd/engram/index.go`

After the existing indexing completes, add:

```go
// After indexer.IndexAll():

// Analyze git history
fmt.Fprintln(os.Stderr, "Analyzing git history...")
analyzer := git.New(store, repoRoot)
historyStats, err := analyzer.AnalyzeAll(ctx)
if err != nil {
    fmt.Fprintf(os.Stderr, "Warning: git history analysis failed: %v\n", err)
    // Continue — git history is optional
} else {
    fmt.Fprintf(os.Stderr, "Git history: %d files analyzed, hottest: %s (%d changes)\n",
        historyStats.FilesAnalyzed, historyStats.HottestFile, historyStats.HottestFrequency)
}

// Analyze conventions (if conventions package exists)
fmt.Fprintln(os.Stderr, "Detecting conventions...")
convAnalyzer := conventions.New(store, repoRoot)
convResult, err := convAnalyzer.Analyze(ctx)
if err != nil {
    fmt.Fprintf(os.Stderr, "Warning: convention analysis failed: %v\n", err)
} else {
    fmt.Fprintf(os.Stderr, "Conventions: %d patterns detected\n", len(convResult.Conventions))
}
```

### Updated Index Output
```
Engram indexing /path/to/repo...
Scanning files... found 342 supported files
Indexing... 342/342 files processed
Git history: 342 files analyzed, hottest: internal/api/router.go (42 changes)
Detecting conventions... 8 patterns detected

Index complete:
  Files processed:     342
  Symbols extracted:   2,847
  Git history:         342 files analyzed
  Conventions:         8 patterns detected
  Duration:            18.3s
  Database:            ~/.engram/a1b2c3d4e5f6/engram.db (5.1 MB)
```

### Error Handling
- Git history analysis failure should NOT fail the index command
- Convention analysis failure should NOT fail the index command
- Both are "best effort" enhancements to the core indexing

### Import Additions
```go
import (
    "github.com/TomOst-Sec/colony-project/internal/git"
    "github.com/TomOst-Sec/colony-project/internal/conventions"
)
```

## Acceptance Criteria
- [ ] `engram index` runs git history analysis after indexing
- [ ] `engram index` runs convention analysis after indexing
- [ ] Git history results are stored in git_context table
- [ ] Convention results are stored in conventions table
- [ ] Git history failure doesn't abort the index command
- [ ] Convention analysis failure doesn't abort the index command
- [ ] Output shows git history and convention stats
- [ ] `engram index --force` also re-analyzes git history and conventions
- [ ] All tests pass

## Implementation Steps
1. Read `internal/git/analyzer.go` to confirm AnalyzeAll signature and AnalysisStats fields
2. Read `internal/conventions/analyzer.go` to confirm Analyze signature and AnalysisResult fields
3. Add imports to `cmd/engram/index.go`
4. Add git history analysis after indexer.IndexAll()
5. Add convention analysis after git history
6. Update summary output to include new stats
7. Update `cmd/engram/index_test.go`:
   - Test: index command still builds
   - Test: index completes even if git analysis would fail (graceful degradation)
8. Run all tests: `go test ./cmd/engram/`

## Testing Requirements
- Unit test: index command builds with new imports
- Integration test: index on this repo also populates git_context table
- Unit test: git history failure is non-fatal (warning only)

## Files to Create/Modify
- `cmd/engram/index.go` — add git history + convention analysis after indexing

## Notes
- Check the actual function signatures in `internal/git/analyzer.go` and `internal/conventions/analyzer.go`. The examples above are approximations — match what actually exists.
- If the conventions package doesn't exist yet (TASK-019 still in progress), wrap the convention analysis in a build tag or just import it and let the build fail if it's not ready. The task dependency system should ensure TASK-019 is done first.
- Keep the changes minimal — this is a wiring task, not a feature task.

---
## Completion Notes
- **Completed by:** alpha-3
- **Date:** 2026-03-15 17:17:41
- **Branch:** task/025
