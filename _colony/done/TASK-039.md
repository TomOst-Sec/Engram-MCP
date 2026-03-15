# TASK-039: Community Convention Modules — Registry and `engram conventions add`

**Priority:** P2
**Assigned:** bravo
**Milestone:** M3: Polish & Growth
**Dependencies:** TASK-019
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 19 from GOALS.md. Community convention packs are shareable JSON files containing coding conventions. `engram conventions add react-typescript` downloads a pack from a community GitHub repo and merges it with locally-inferred conventions. Local conventions always win on conflict.

## Specification

### Convention Pack Format

```json
{
  "name": "react-typescript",
  "version": "1.0.0",
  "description": "React + TypeScript conventions for modern frontend projects",
  "author": "community",
  "conventions": [
    {
      "pattern": "PascalCase component names",
      "description": "React components use PascalCase file and function names",
      "category": "naming",
      "confidence": 1.0,
      "language": "typescript",
      "examples": ["UserProfile.tsx", "LoginForm.tsx"]
    }
  ]
}
```

### CLI Commands

**`engram conventions list`** — List installed convention packs
**`engram conventions add <pack-name>`** — Download and install a convention pack
**`engram conventions remove <pack-name>`** — Remove a convention pack

### Convention Registry: `internal/conventions/registry.go`

```go
type PackRegistry struct {
    store    *storage.Store
    cacheDir string  // ~/.engram/packs/
}

func NewPackRegistry(store *storage.Store) *PackRegistry

// Install downloads a convention pack from the community repo.
func (r *PackRegistry) Install(ctx context.Context, name string) error

// Remove deletes an installed convention pack.
func (r *PackRegistry) Remove(name string) error

// List returns all installed packs.
func (r *PackRegistry) List() ([]PackInfo, error)

// MergeConventions merges community conventions with local ones.
// Local conventions always win on conflict (same pattern + language).
func (r *PackRegistry) MergeConventions(local []Convention) []Convention
```

### Pack Source
Packs are JSON files hosted at:
`https://raw.githubusercontent.com/TomOst-Sec/engram-conventions/main/packs/<name>.json`

For MVP, the community repo doesn't need to exist yet. The code should handle 404 gracefully: "Convention pack 'xyz' not found. Browse available packs at https://github.com/TomOst-Sec/engram-conventions"

### Storage
Installed packs stored as JSON files in `~/.engram/packs/`:
```
~/.engram/packs/
├── react-typescript.json
├── go-clean-arch.json
└── rails-standard.json
```

## Acceptance Criteria
- [ ] `engram conventions add <name>` downloads and stores a pack
- [ ] `engram conventions remove <name>` removes an installed pack
- [ ] `engram conventions list` shows installed packs
- [ ] MergeConventions combines community + local, local wins on conflict
- [ ] 404 from community repo returns helpful error
- [ ] Pack JSON validation (required fields)
- [ ] All tests pass

## Implementation Steps
1. Create `internal/conventions/registry.go` — PackRegistry
2. Create `internal/conventions/pack.go` — Pack struct, validation, JSON parsing
3. Create `cmd/engram/conventions.go` — conventions subcommand with list/add/remove
4. Create tests for registry, pack parsing, and merge logic
5. Register conventions command in main.go
6. Run all tests

## Files to Create/Modify
- `internal/conventions/registry.go` — pack registry
- `internal/conventions/pack.go` — pack struct and validation
- `internal/conventions/registry_test.go` — registry tests
- `cmd/engram/conventions.go` — CLI subcommand
- `cmd/engram/conventions_test.go` — CLI tests
- `cmd/engram/main.go` — register command

## Notes
- Use `net/http` for downloading packs. This is one of the ONLY network calls Engram makes.
- Cache downloaded packs locally. Don't re-download on every `engram serve` start.
- The merge logic is simple: for each community convention, check if a local convention exists with the same pattern+language. If yes, keep local. If no, add community.
- Convention packs are additive only — they can't remove or override local conventions.

---
## Completion Notes
- **Completed by:** bravo-1
- **Date:** 2026-03-15 18:09:32
- **Branch:** task/039
