# TASK-011: `get_architecture` MCP Tool — Module Map and Import Graph

**Priority:** P1
**Assigned:** alpha
**Milestone:** M1: MVP
**Dependencies:** TASK-003, TASK-005
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
The `get_architecture` tool gives AI coding agents a high-level understanding of a project's structure — what modules exist, what each one does, how they depend on each other, and what they export. When an AI agent needs to understand "how is this project organized?" or "what module handles authentication?", this tool provides the answer. It reads from the `architecture` and `code_index` tables populated by the indexer.

## Specification
Create the `internal/tools/architecture` package implementing the `get_architecture` MCP tool.

### MCP Tool Definition
**Name:** `get_architecture`
**Description:** "Returns a structured map of the project's architecture: modules, their responsibilities, dependencies, and key exports."

**Input Schema:**
```json
{
    "type": "object",
    "properties": {
        "module": {
            "type": "string",
            "description": "Focus on a specific module/directory path (returns detailed view of that module)"
        },
        "depth": {
            "type": "integer",
            "description": "Directory depth for module detection (default: 2, e.g., 'internal/auth' is depth 2)"
        },
        "include_exports": {
            "type": "boolean",
            "description": "Include exported symbols per module (default: false, set true for detailed view)"
        }
    },
    "required": []
}
```

### Architecture Analysis Logic

The tool builds the architecture from data already in SQLite (populated by the indexer):

1. **Module Detection:**
   - Query `code_index` table for all unique file paths
   - Group files by directory up to `depth` levels (default 2)
   - Each unique directory group = one module
   - Example: files in `internal/auth/handler.go`, `internal/auth/middleware.go` → module `internal/auth`

2. **Module Description:**
   - Infer module purpose from directory name + symbol names + docstrings
   - Simple heuristic: join the directory name with first sentence of the first docstring found
   - Store in `architecture` table if not already present

3. **Dependency Graph:**
   - Query `code_index` where `symbol_type = 'import'` for each module
   - Map import paths to module paths within the project
   - External imports (outside the project) listed separately
   - Build directed dependency graph: module A → imports from module B

4. **Module Exports:**
   - For Go: exported symbols start with uppercase
   - For Python: symbols not starting with `_`
   - For TypeScript/JavaScript: symbols in `export` statements
   - Count per module

5. **Complexity Score:**
   - Simple heuristic: `(number of symbols * 0.5) + (number of imports * 0.3) + (number of files * 0.2)`
   - Normalized to 0.0–10.0 scale

### Response Format
```json
{
    "project_root": "/path/to/repo",
    "total_modules": 8,
    "total_files": 42,
    "modules": [
        {
            "name": "internal/auth",
            "path": "internal/auth",
            "description": "Authentication and authorization handling",
            "files": 4,
            "symbols": 23,
            "exports": ["HandleLogin", "AuthMiddleware", "ValidateToken"],
            "dependencies": ["internal/storage", "internal/config"],
            "external_dependencies": ["github.com/golang-jwt/jwt"],
            "complexity_score": 6.2
        }
    ],
    "dependency_graph": {
        "internal/auth": ["internal/storage", "internal/config"],
        "internal/mcp": ["internal/tools", "internal/config"],
        "internal/tools": ["internal/storage", "internal/embeddings"]
    }
}
```

When `module` parameter is specified, return detailed view of just that module:
```json
{
    "module": {
        "name": "internal/auth",
        "path": "internal/auth",
        "description": "Authentication and authorization handling",
        "files": [
            {"path": "internal/auth/handler.go", "symbols": 8},
            {"path": "internal/auth/middleware.go", "symbols": 5}
        ],
        "exports": [
            {"name": "HandleLogin", "type": "function", "signature": "func HandleLogin(...) error", "file": "handler.go", "line": 42},
            {"name": "AuthMiddleware", "type": "function", "signature": "func AuthMiddleware(...) http.Handler", "file": "middleware.go", "line": 15}
        ],
        "dependencies": ["internal/storage", "internal/config"],
        "dependents": ["internal/mcp"],
        "complexity_score": 6.2
    }
}
```

### Package Structure
```go
// internal/tools/architecture/architecture.go
type ArchitectureTool struct {
    store    *storage.Store
    repoRoot string
}

func NewArchitectureTool(store *storage.Store, repoRoot string) *ArchitectureTool
func (t *ArchitectureTool) Definition() mcp.Tool
func (t *ArchitectureTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

// internal/tools/architecture/analysis.go
func DetectModules(store *storage.Store, depth int) ([]Module, error)
func BuildDependencyGraph(store *storage.Store, modules []Module) (map[string][]string, error)
func ComputeComplexity(symbols int, imports int, files int) float64

// internal/tools/architecture/register.go
func RegisterTools(server *mcp.Server, store *storage.Store, repoRoot string)
```

## Acceptance Criteria
- [ ] `get_architecture` tool is registered as an MCP tool with correct schema
- [ ] Without parameters, returns full project module map
- [ ] Module detection correctly groups files by directory path at specified depth
- [ ] Dependency graph correctly maps import statements to project modules
- [ ] External dependencies are listed separately from internal ones
- [ ] `module` parameter returns detailed view of a single module
- [ ] `include_exports` parameter controls whether exported symbols are listed
- [ ] Complexity score is computed and within 0.0–10.0 range
- [ ] Works correctly when code_index is empty (returns empty module list, no error)
- [ ] All tests pass

## Implementation Steps
1. Create `internal/tools/architecture/architecture.go` — ArchitectureTool struct, Definition, Handle
2. Create `internal/tools/architecture/analysis.go` — DetectModules, BuildDependencyGraph, ComputeComplexity
3. Create `internal/tools/architecture/register.go` — RegisterTools function
4. Create `internal/tools/architecture/architecture_test.go`:
   - Set up test DB with sample code_index data (multiple modules, import relationships)
   - Test: DetectModules groups files correctly
   - Test: BuildDependencyGraph maps imports to modules
   - Test: ComputeComplexity returns value in 0.0–10.0 range
   - Test: Full tool handler returns correct JSON structure
   - Test: module parameter filters to single module
   - Test: empty database returns empty module list (no error)
5. Create `internal/tools/architecture/analysis_test.go`:
   - Test: DetectModules with depth=1 vs depth=2
   - Test: BuildDependencyGraph separates internal vs external deps
   - Test: ComputeComplexity normalization
6. Run all tests

## Testing Requirements
- Unit test: DetectModules with 10 files across 3 directories produces 3 modules
- Unit test: DetectModules with depth=1 groups at top-level only
- Unit test: BuildDependencyGraph with Go imports maps "internal/auth" → imports "internal/storage"
- Unit test: External imports (e.g., "github.com/spf13/cobra") listed separately
- Unit test: ComputeComplexity(10 symbols, 5 imports, 3 files) returns score in range
- Unit test: Full handler returns JSON with modules[], dependency_graph{}, total_modules
- Unit test: Handler with module="internal/auth" returns only that module's details
- Unit test: Empty database → empty response (no panic)

## Files to Create/Modify
- `internal/tools/architecture/architecture.go` — ArchitectureTool implementation
- `internal/tools/architecture/analysis.go` — module detection, dependency graph, complexity scoring
- `internal/tools/architecture/register.go` — MCP registration helper
- `internal/tools/architecture/architecture_test.go` — tool handler tests
- `internal/tools/architecture/analysis_test.go` — analysis function tests

## Notes
- Module detection is simple directory grouping, not a complex heuristic. For Go, `internal/auth/` is obviously a module. For JS/TS with `src/components/Button/`, the depth parameter controls granularity.
- Import path resolution is language-specific:
  - **Go:** import paths are absolute module paths (e.g., `github.com/user/repo/internal/auth`). Strip the module prefix to get the relative module path.
  - **Python:** `from internal.auth import handler` → module path `internal/auth`
  - **TypeScript/JS:** `import { x } from '../auth/handler'` → resolve relative to file location
- For MVP, focus on Go import resolution (since the Engram codebase is Go). TS/Python resolution can be improved later.
- The `architecture` table in SQLite can cache computed module data. Query code_index for raw data, write aggregated results to architecture table.
- Do NOT read files from disk in this tool. All data comes from the code_index table (populated by the indexer pipeline). This keeps the tool fast (<200ms).
