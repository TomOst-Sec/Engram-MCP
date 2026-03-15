# TASK-037: Register Missing Parsers in Default Registry

**Priority:** P1
**Assigned:** bravo
**Milestone:** M2: Core Features
**Dependencies:** none
**Status:** review
**Created:** 2026-03-15
**Author:** beta-tester

## Context

Six language parsers exist and have passing tests but are not registered in `NewDefaultRegistry()` in `internal/parser/registry.go`. BUG-016 flagged Ruby, PHP, C, C++ — CSharp was fixed but the others were missed. Lua and Zig were also never registered.

## Specification

Add all six missing parser registrations to `NewDefaultRegistry()`.

## Acceptance Criteria

- [ ] `NewDefaultRegistry()` includes all 15 parsers
- [ ] `registry.SupportedLanguages()` returns all 15 language names
- [ ] Existing tests still pass

## Implementation Steps

1. Add to `NewDefaultRegistry()` in `internal/parser/registry.go`:
   ```go
   r.Register(NewRubyParser())
   r.Register(NewPHPParser())
   r.Register(NewCParser())
   r.Register(NewCPPParser())
   r.Register(NewLuaParser())
   r.Register(NewZigParser())
   ```
2. Run `make test` to verify

## Testing Requirements

- All existing parser tests still pass

## Files to Create/Modify

- `internal/parser/registry.go` — add 6 Register() calls

## Notes

- BUG-016 can be deleted after this task is merged

---
## Completion Notes
- **Completed by:** bravo-1
- **Date:** 2026-03-15 17:41:49
- **Branch:** task/037
