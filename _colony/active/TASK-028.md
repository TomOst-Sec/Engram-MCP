# TASK-028: Lua + Zig Tree-Sitter Grammars

**Priority:** P2
**Assigned:** bravo
**Milestone:** M2: Core Features
**Dependencies:** TASK-005
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Final language batch for Milestone 2. Lua and Zig complete the 15-language target from GOALS.md. Lua is widely used in game development (Love2D, Roblox), embedded systems (OpenResty/Nginx), and configuration (Neovim). Zig is a growing systems language alternative to C/C++. Both follow the established parser pattern.

## Specification
Add Lua and Zig parsers to `internal/parser/`.

### Dependencies
- `github.com/smacker/go-tree-sitter/lua` — Lua grammar (may not exist — check alternatives)
- Zig grammar — check `github.com/smacker/go-tree-sitter` for availability, or `github.com/tree-sitter-grammars/tree-sitter-zig` with Go bindings

### Lua Parser Extractions
Parse `.lua` files and extract:
- **Functions:** `function name(params)` and `local function name(params)` — capture name, params
- **Methods:** `function obj:method(params)` and `function obj.method(params)` — capture object + method name
- **Module returns:** `return { ... }` at end of file — detect as module exports
- **Requires:** `require("module")` and `require 'module'` — capture as import
- **Local variables (tables as classes):** `local ClassName = {}` followed by method definitions — capture as type "class"
- **Comments:** `---` prefix (LDoc/EmmyLua style) — capture as docstrings
- **Test functions:** functions containing "test" in name — mark as type "test"

### Zig Parser Extractions
Parse `.zig` files and extract:
- **Functions:** `fn name(params) ReturnType` and `pub fn name(params) ReturnType` — capture name, full signature, `///` doc comments
- **Structs:** `const Name = struct { ... }` — capture name, fields
- **Enums:** `const Name = enum { ... }` — capture name, values
- **Unions:** `const Name = union(enum) { ... }` — capture name
- **Constants:** `const name = ...` (non-type constants) — capture as type "type"
- **Imports:** `@import("module")` — capture as import
- **Test blocks:** `test "description" { ... }` — capture as type "test"
- **Errors:** `const Name = error { ... }` — capture as type "type"

### Registration
Register both parsers in `NewDefaultRegistry()`.

## Acceptance Criteria
- [ ] Lua parser extracts functions, methods, requires from .lua files
- [ ] Zig parser extracts functions, structs, enums, imports, test blocks from .zig files
- [ ] Registry routes `.lua` → Lua, `.zig` → Zig
- [ ] Symbols include accurate line numbers
- [ ] Lua parser captures `---` LDoc comments
- [ ] Zig parser captures `///` doc comments
- [ ] Test symbols identified in both languages
- [ ] All existing parser tests still pass
- [ ] All new tests pass

## Implementation Steps
1. Check if go-tree-sitter has Lua and Zig grammars. If not available as Go packages, see if alternative bindings exist. If no bindings exist for a language, implement a basic regex-based fallback parser and note it in the code.
2. Create `internal/parser/lua_parser.go`
3. Create `internal/parser/zig_parser.go`
4. Create `internal/parser/testdata/sample.lua` — Lua fixture with functions, methods, requires, module table
5. Create `internal/parser/testdata/sample.zig` — Zig fixture with functions, structs, enums, imports, tests
6. Create `internal/parser/lua_parser_test.go`
7. Create `internal/parser/zig_parser_test.go`
8. Register in `NewDefaultRegistry()`
9. Run ALL parser tests

## Testing Requirements
- Unit test: Lua parser extracts global and local functions
- Unit test: Lua parser captures method definitions (colon syntax)
- Unit test: Lua parser captures require statements
- Unit test: Zig parser extracts pub and private functions
- Unit test: Zig parser captures structs and enums
- Unit test: Zig parser captures test blocks
- Unit test: Zig parser captures @import
- Unit test: Registry routes .lua and .zig correctly
- Regression test: all existing parser tests pass

## Files to Create/Modify
- `internal/parser/lua_parser.go` — Lua parser
- `internal/parser/zig_parser.go` — Zig parser
- `internal/parser/lua_parser_test.go` — tests
- `internal/parser/zig_parser_test.go` — tests
- `internal/parser/testdata/sample.lua` — Lua fixture
- `internal/parser/testdata/sample.zig` — Zig fixture
- `internal/parser/registry.go` — add registrations

## Notes
- If tree-sitter grammars aren't available as Go bindings, implement a **regex-based fallback parser** that extracts function definitions and imports using regular expressions. This is explicitly allowed by GOALS.md's graceful degradation constraint: "If tree-sitter grammar is missing for a language, fall back to regex-based symbol extraction."
- For regex fallback: match `function\s+(\w+)` for Lua, `(pub\s+)?fn\s+(\w+)` for Zig. This won't be as accurate as tree-sitter but provides basic coverage.
- Lua has no formal class system. The `local ClassName = {}` pattern followed by `function ClassName:method()` is idiomatic. Detect this pattern heuristically.
- Zig's `test "name" { }` blocks are string-named, not function-named. Use the test description as the symbol name.
- With these two languages, Engram reaches 15 total: Go, Python, TypeScript, JavaScript, Rust, Java, C#, Ruby, PHP, Swift, Kotlin, C, C++, Lua, Zig.
