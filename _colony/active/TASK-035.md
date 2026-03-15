# TASK-035: Multi-Repo Support — Cross-Repository Search and Context

**Priority:** P2
**Assigned:** alpha
**Milestone:** M3: Polish & Growth
**Dependencies:** TASK-002, TASK-003
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 17 from GOALS.md. Developers working on microservice architectures or monorepos need Engram to understand multiple related repositories. This task adds multi-repo support: `engram.json` can reference additional repos, the MCP server merges indexes, and search/architecture tools work across repos.

## Specification

### Config Extension

Add to `engram.json`:
```json
{
  "additional_repos": [
    {
      "path": "../api-service",
      "name": "api"
    },
    {
      "path": "../shared-lib",
      "name": "shared"
    }
  ]
}
```

### Config Changes: `internal/config/config.go`

```go
type RepoRef struct {
    Path string `json:"path"` // relative or absolute path
    Name string `json:"name"` // display name for search results
}

// Add to Config struct:
AdditionalRepos []RepoRef `json:"additional_repos"`
```

### Multi-Store: `internal/storage/multi.go`

```go
type MultiStore struct {
    primary    *Store     // main repo
    additional []*Store   // additional repos
    names      []string   // repo display names
}

func NewMultiStore(primary *Store, additional []*Store, names []string) *MultiStore

// SearchCode searches across all stores and merges results.
func (ms *MultiStore) SearchCode(query string, filters map[string]string, limit int) ([]SearchResult, error)

// Close closes all stores.
func (ms *MultiStore) Close() error
```

### Search Result Enrichment

Search results from additional repos include the repo name prefix:
```
[api] internal/handlers/auth.go:42  HandleLogin  function
[shared] pkg/auth/token.go:15      ValidateToken  function
internal/server/main.go:30          StartServer   function (primary repo, no prefix)
```

### Serve Integration

When `engram serve` starts with `additional_repos` configured:
1. Open primary store (existing behavior)
2. For each additional repo: resolve path, compute hash, open store
3. Pass MultiStore to tool registrations instead of single Store

### Architecture Cross-Repo

The `get_architecture` tool shows inter-repo dependencies when multiple repos are configured:
```json
{
  "repos": [
    { "name": "primary", "modules": [...] },
    { "name": "api", "modules": [...] },
    { "name": "shared", "modules": [...] }
  ],
  "cross_repo_dependencies": [
    { "from": "api/handlers", "to": "shared/auth", "type": "import" }
  ]
}
```

## Acceptance Criteria
- [ ] Config parses `additional_repos` field correctly
- [ ] MultiStore opens multiple SQLite databases
- [ ] Search across multiple repos returns merged, ranked results
- [ ] Results from additional repos are labeled with repo name
- [ ] Primary repo results appear without prefix
- [ ] MultiStore.Close() closes all connections
- [ ] Single-repo mode still works when no additional_repos configured
- [ ] All tests pass

## Implementation Steps
1. Add `RepoRef` struct and `AdditionalRepos` field to config
2. Create `internal/storage/multi.go` — MultiStore implementation
3. Create `internal/storage/multi_test.go`:
   - Test: MultiStore searches across two stores
   - Test: Results include repo name prefix
   - Test: Close() closes all stores
   - Test: Single-store mode (no additional repos) works
4. Update config tests to handle additional_repos parsing
5. Run all tests

## Testing Requirements
- Unit test: Config parses additional_repos JSON correctly
- Unit test: MultiStore.SearchCode merges results from two stores
- Unit test: Results sorted by relevance across repos
- Unit test: Repo name prefix applied to additional repo results
- Unit test: MultiStore.Close closes all stores without error
- Unit test: Empty additional_repos doesn't break existing behavior

## Files to Create/Modify
- `internal/config/config.go` — add RepoRef struct and AdditionalRepos field
- `internal/config/config_test.go` — add additional_repos test
- `internal/storage/multi.go` — MultiStore implementation
- `internal/storage/multi_test.go` — MultiStore tests

## Notes
- Do NOT modify serve.go to use MultiStore yet. That wiring will be a separate task. This task creates the MultiStore abstraction and config support.
- Each additional repo has its own database at `~/.engram/<hash>/engram.db`. The hash is based on the repo's absolute path (resolved from the relative path in config).
- Multi-repo search should interleave results by relevance, not show all results from one repo then another.
- For MVP multi-repo, focus on search. Architecture cross-repo analysis can be added later.
- Relative paths in `additional_repos` are resolved relative to the primary repo root, not the cwd.
