# TASK-019: `get_conventions` MCP Tool — Pattern Inference and Convention Detection

**Priority:** P1
**Assigned:** bravo
**Milestone:** M2: Core Features
**Dependencies:** TASK-003, TASK-005
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 8 from GOALS.md. The `get_conventions` tool analyzes the codebase to infer team conventions and returns them as structured rules. It detects naming patterns (camelCase vs snake_case), error handling style, test structure, import ordering, and more. This is one of Engram's killer features — the AI coding agent automatically learns how the team writes code and follows the same patterns. The `conventions` table already exists in the schema (TASK-003).

## Specification
Create the `internal/conventions` package for convention analysis and `internal/tools/conventions` for the MCP tool.

### Convention Analyzer: `internal/conventions/analyzer.go`

```go
type Analyzer struct {
    store    *storage.Store
    repoRoot string
}

type Convention struct {
    Pattern     string   // e.g., "snake_case function names"
    Description string   // human-readable description
    Category    string   // naming, error_handling, testing, imports, structure, documentation
    Confidence  float64  // 0.0-1.0, based on consistency across codebase
    Examples    []string // 3-5 code examples showing the pattern
    Language    string   // language scope (or "all" for cross-language)
}

type AnalysisResult struct {
    Conventions []Convention
    Duration    time.Duration
}

func New(store *storage.Store, repoRoot string) *Analyzer

// Analyze scans the code_index and infers conventions.
func (a *Analyzer) Analyze(ctx context.Context) (*AnalysisResult, error)

// GetConventions retrieves previously stored conventions, optionally filtered.
func (a *Analyzer) GetConventions(ctx context.Context, language string, category string) ([]Convention, error)
```

### Convention Categories and Detection Logic

**1. Naming Conventions (`naming`)**
Query code_index for symbol_name values grouped by language and symbol_type:
- Count snake_case vs camelCase vs PascalCase for functions
- Count snake_case vs camelCase vs PascalCase for types/classes
- If >80% follow one pattern → high confidence
- If 60-80% → medium confidence
- If <60% → low confidence, don't report

Detection: use regex patterns:
- snake_case: `^[a-z][a-z0-9]*(_[a-z0-9]+)*$`
- camelCase: `^[a-z][a-zA-Z0-9]*$`
- PascalCase: `^[A-Z][a-zA-Z0-9]*$`

**2. Error Handling (`error_handling`)**
Query code_index signatures:
- Go: count functions returning `error` vs not → "Go functions return error as last value"
- Python: scan for `try/except` patterns in function bodies (approximate via signature/docstring)
- TypeScript/JavaScript: check for `throws` in JSDoc, `Promise<>` return types

**3. Test Structure (`testing`)**
Query code_index where symbol_type = 'test':
- Detect test naming: `Test_` prefix, `test_` prefix, `_test` suffix
- Detect test file naming: `*_test.go`, `*.test.ts`, `test_*.py`
- Count test-to-code ratio per language

**4. Import Organization (`imports`)**
Query code_index where symbol_type = 'import':
- Detect grouping patterns (stdlib vs third-party vs local)
- Detect absolute vs relative imports (JS/TS/Python)

**5. Documentation (`documentation`)**
Query code_index docstring field:
- Calculate percentage of functions with docstrings per language
- Detect docstring style (JSDoc, Python docstrings, Go comments, Javadoc)

**6. File Organization (`structure`)**
Analyze file paths from code_index:
- Detect common directory patterns (src/internal/cmd for Go, components/hooks/utils for React, etc.)
- Detect file naming conventions (kebab-case, PascalCase, camelCase)

### Storage

Store detected conventions in the `conventions` table:
```go
func StoreConventions(store *storage.Store, conventions []Convention) error
```
- Clear existing conventions before storing (full replace each analysis)
- Store examples as JSON array text

### MCP Tool: `internal/tools/conventions/register.go`

```go
func RegisterTools(server *engmcp.Server, store *storage.Store, repoRoot string)
```

**Tool: `get_conventions`**
- Input: `language` (optional filter), `category` (optional filter)
- Output: JSON array of conventions with pattern, description, confidence, examples
- If no conventions stored, run analysis on-the-fly
- Return conventions sorted by confidence descending

**Tool Definition:**
```json
{
  "name": "get_conventions",
  "description": "Get inferred code conventions and team patterns. Returns naming styles, error handling patterns, test structure, and more.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "language": { "type": "string", "description": "Filter by language (go, python, typescript, etc.)" },
      "category": { "type": "string", "description": "Filter by category (naming, error_handling, testing, imports, structure, documentation)" }
    }
  }
}
```

## Acceptance Criteria
- [ ] Analyzer detects naming conventions (snake_case vs camelCase) per language with correct confidence scores
- [ ] Analyzer detects test naming patterns
- [ ] Analyzer detects documentation coverage percentage
- [ ] Analyzer detects import organization patterns
- [ ] Conventions are stored in the conventions table
- [ ] `get_conventions` MCP tool returns conventions as JSON
- [ ] `get_conventions` accepts language and category filters
- [ ] Confidence scoring: >80% consistency = high (>0.8), 60-80% = medium (0.6-0.8)
- [ ] Each convention includes 3-5 code examples
- [ ] All tests pass

## Implementation Steps
1. Create `internal/conventions/analyzer.go` — Analyzer struct, New, Analyze, GetConventions
2. Create `internal/conventions/naming.go` — naming convention detection (snake_case, camelCase, PascalCase)
3. Create `internal/conventions/testing_patterns.go` — test structure detection
4. Create `internal/conventions/documentation.go` — docstring coverage analysis
5. Create `internal/conventions/store.go` — StoreConventions, GetConventions storage operations
6. Create `internal/tools/conventions/tools.go` — get_conventions MCP tool handler
7. Create `internal/tools/conventions/register.go` — RegisterTools
8. Create `internal/conventions/analyzer_test.go`:
   - Test: naming detection correctly identifies snake_case-dominant codebase
   - Test: naming detection correctly identifies camelCase-dominant codebase
   - Test: confidence scoring is accurate (feed 90% snake_case → confidence > 0.8)
   - Test: conventions with <60% consistency are not reported
9. Create `internal/conventions/naming_test.go`:
   - Test: isSnakeCase, isCamelCase, isPascalCase regex patterns
10. Create `internal/tools/conventions/tools_test.go`:
    - Test: get_conventions returns valid JSON response
    - Test: language filter works correctly
    - Test: category filter works correctly
11. Run all tests

## Testing Requirements
- Unit test: isSnakeCase correctly classifies "my_function", "myFunction", "MyFunction"
- Unit test: isCamelCase correctly classifies names
- Unit test: isPascalCase correctly classifies names
- Unit test: Confidence calculation for 90% consistency returns >0.8
- Unit test: Confidence calculation for 50% consistency returns <0.6 (not reported)
- Unit test: StoreConventions + GetConventions round-trip through SQLite
- Unit test: MCP tool handler returns correct JSON structure
- Unit test: Language filter restricts results to specified language
- Unit test: Category filter restricts results to specified category
- Integration test: Analyze on this project's codebase returns at least one convention

## Files to Create/Modify
- `internal/conventions/analyzer.go` — main Analyzer
- `internal/conventions/naming.go` — naming pattern detection
- `internal/conventions/testing_patterns.go` — test pattern detection
- `internal/conventions/documentation.go` — documentation pattern detection
- `internal/conventions/store.go` — conventions table operations
- `internal/conventions/analyzer_test.go` — analyzer tests
- `internal/conventions/naming_test.go` — naming detection tests
- `internal/tools/conventions/tools.go` — MCP tool handler
- `internal/tools/conventions/register.go` — tool registration
- `internal/tools/conventions/tools_test.go` — MCP tool tests

## Notes
- The `conventions` table already exists in the schema: id, pattern, description, category, confidence, examples (TEXT), language, created_at, updated_at. The `examples` column stores a JSON array of strings.
- Follow the same MCP tool registration pattern as `internal/tools/search/register.go` and `internal/tools/memory/register.go`. Study those files for the exact pattern.
- The convention analysis queries the `code_index` table heavily. Use efficient SQL with GROUP BY and COUNT for pattern detection rather than loading all symbols into memory.
- Do NOT modify `cmd/engram/serve.go` in this task. The tool will be wired into the serve command in a later task.
- For naming detection, only analyze user-defined symbols (functions, methods, types, classes). Skip imports and built-in names.
- The confidence threshold for reporting is 60%. Below that, the codebase is too inconsistent to call it a "convention."
