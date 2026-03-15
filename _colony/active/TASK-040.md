# TASK-040: Benchmark Suite — Performance Validation

**Priority:** P2
**Assigned:** alpha
**Milestone:** M3: Polish & Growth
**Dependencies:** TASK-013, TASK-010
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
GOALS.md requires published benchmarks for index time, query latency, and memory usage. This task creates a Go benchmark suite that validates the performance constraints: <200ms tool responses, <100ms cached search, <30s full index for 100K LOC, <500ms incremental update. Results are printed in a table format for documentation.

## Specification

### Benchmark Package: `benchmarks/`

Create Go benchmark tests that measure:

**1. Indexing Performance**
```go
// BenchmarkIndexSmallRepo benchmarks indexing a ~1K LOC repo
// BenchmarkIndexMediumRepo benchmarks indexing a ~10K LOC repo
// Target: <30s for 100K LOC (linear extrapolation)
```
Generate synthetic test repos with realistic Go/Python/TypeScript files.

**2. Search Latency**
```go
// BenchmarkSearchFTS5 benchmarks FTS5-only search
// BenchmarkSearchHybrid benchmarks FTS5 + vector similarity
// Target: <200ms for repos up to 100K LOC, <100ms cached
```

**3. Memory Usage**
```go
// BenchmarkMemoryUsage measures RSS during steady-state serving
// Target: <200MB for 100K LOC repo
```

**4. Incremental Re-Index**
```go
// BenchmarkIncrementalIndex benchmarks single-file re-index
// Target: <500ms per file change
```

**5. MCP Tool Responses**
```go
// BenchmarkSearchCodeTool benchmarks the full search_code MCP tool path
// BenchmarkRememberTool benchmarks memory storage
// BenchmarkRecallTool benchmarks memory retrieval
// BenchmarkGetArchitectureTool benchmarks architecture query
// Target: <200ms each
```

### Test Repo Generator: `benchmarks/testdata/generate.go`

```go
// GenerateRepo creates a synthetic repository with the specified number of files.
func GenerateRepo(dir string, fileCount int, avgLOCPerFile int) error
```

Generates realistic files:
- 40% Go files with functions, types, imports
- 30% Python files with classes, functions
- 30% TypeScript files with components, interfaces

### Results Output

```
Engram Benchmark Results
========================

Indexing:
  1K LOC:    0.8s  (target: <30s extrapolated)
  10K LOC:   4.2s  (target: <30s extrapolated)

Search (10K LOC index):
  FTS5 only:   12ms  (target: <200ms) ✓
  Hybrid:      45ms  (target: <200ms) ✓
  Cached:      3ms   (target: <100ms) ✓

MCP Tools:
  search_code:      42ms  (target: <200ms) ✓
  remember:         8ms   (target: <200ms) ✓
  recall:           15ms  (target: <200ms) ✓
  get_architecture: 23ms  (target: <200ms) ✓

Memory:
  Idle server:      45MB
  10K LOC loaded:   78MB  (target: <200MB) ✓
```

### Makefile Integration
Add to Makefile:
```makefile
.PHONY: bench
bench:
	go test -tags sqlite_fts5 -bench=. -benchmem ./benchmarks/...
```

## Acceptance Criteria
- [ ] Benchmark tests exist for indexing, search, MCP tools, and memory
- [ ] Benchmarks generate synthetic test repos (no external data needed)
- [ ] `go test -bench=. ./benchmarks/` runs all benchmarks
- [ ] Results printed in readable format with pass/fail against targets
- [ ] Makefile has `bench` target
- [ ] All benchmarks complete without error
- [ ] All regular tests still pass

## Implementation Steps
1. Create `benchmarks/` directory
2. Create `benchmarks/testdata/generate.go` — test repo generator
3. Create `benchmarks/index_bench_test.go` — indexing benchmarks
4. Create `benchmarks/search_bench_test.go` — search benchmarks
5. Create `benchmarks/tools_bench_test.go` — MCP tool benchmarks
6. Create `benchmarks/memory_bench_test.go` — memory usage
7. Update Makefile with bench target
8. Run benchmarks, document results

## Files to Create/Modify
- `benchmarks/testdata/generate.go` — synthetic repo generator
- `benchmarks/index_bench_test.go` — indexing benchmarks
- `benchmarks/search_bench_test.go` — search benchmarks
- `benchmarks/tools_bench_test.go` — MCP tool benchmarks
- `benchmarks/memory_bench_test.go` — memory usage
- `Makefile` — add bench target

## Notes
- Use Go's standard `testing.B` for benchmarks. No external benchmark frameworks.
- Synthetic repos should be created in `testing.TempDir()` — cleaned up automatically.
- Memory benchmarks can use `runtime.ReadMemStats()` for heap measurements.
- The benchmark suite should be self-contained — no dependencies on external repos or data.
- Keep generated file sizes realistic: 50-200 LOC per file, mix of functions and types.
