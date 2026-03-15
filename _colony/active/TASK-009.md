# TASK-009: TypeScript and JavaScript Tree-Sitter Grammars

**Priority:** P1
**Assigned:** bravo
**Milestone:** M1: MVP
**Dependencies:** TASK-005
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Engram's MVP targets 6 languages. TASK-005 creates the parser framework with Go and Python support. This task adds TypeScript and JavaScript — two of the most popular languages in AI-assisted development and critical for Engram's target audience (developers using Cursor, Claude Code, etc. on web projects). TypeScript and JavaScript share many AST structures but have key differences (type annotations, interfaces, enums in TS).

## Specification
Add TypeScript and JavaScript parsers to the `internal/parser` package, following the patterns established in TASK-005.

### Dependencies
- `github.com/smacker/go-tree-sitter/typescript/typescript` — TypeScript grammar
- `github.com/smacker/go-tree-sitter/javascript` — JavaScript grammar

### TypeScript Parser Extractions
Parse `.ts` and `.tsx` files and extract:
- **Functions:** `function name(params): ReturnType` and arrow functions assigned to `const/let/var name = (...) =>` at module level — capture name, full signature (including type annotations), docstring (preceding JSDoc comment `/** ... */`)
- **Methods:** functions inside class bodies — capture class name + method name, full signature, access modifier (public/private/protected)
- **Classes:** `class Name extends Base implements Interface` — capture name, base class, implemented interfaces, docstring
- **Interfaces:** `interface Name extends Other` — capture name, extended interfaces, properties/methods
- **Types:** `type Name = ...` — capture name, type alias definition
- **Enums:** `enum Name { ... }` — capture name, members
- **Imports:** `import { x } from 'y'`, `import x from 'y'`, `import * as x from 'y'` — capture all import paths and imported names
- **Exports:** `export function/class/const/type/interface` and `export { x } from 'y'` — capture exported names
- **Test functions:** functions named `it(...)`, `test(...)`, `describe(...)`, or in files matching `*.test.ts`, `*.spec.ts` — mark as type "test"

### JavaScript Parser Extractions
Parse `.js`, `.jsx`, `.mjs`, `.cjs` files and extract:
- Same as TypeScript MINUS: interfaces, type aliases, enums, type annotations in signatures
- Additional: `module.exports` and `exports.name` patterns (CommonJS)

### Implementation Pattern
Follow the exact same pattern as the Go and Python parsers from TASK-005:
1. Implement the `Parser` interface
2. Use tree-sitter queries to extract symbols
3. Register with the parser Registry
4. Include comprehensive test fixtures

## Acceptance Criteria
- [ ] TypeScript parser extracts functions (regular + arrow), methods, classes, interfaces, types, enums, imports, exports, and test functions
- [ ] JavaScript parser extracts functions, methods, classes, imports (ES6 + CommonJS), exports, and test functions
- [ ] Registry routes `.ts` and `.tsx` files to TypeScript parser
- [ ] Registry routes `.js`, `.jsx`, `.mjs`, `.cjs` files to JavaScript parser
- [ ] Symbols include accurate start/end line numbers
- [ ] Symbols include JSDoc docstrings when present
- [ ] Symbols include full type-annotated signatures (TypeScript)
- [ ] All tests pass

## Implementation Steps
1. `go get github.com/smacker/go-tree-sitter/typescript/typescript github.com/smacker/go-tree-sitter/javascript`
2. Create `internal/parser/typescript_parser.go` — TypeScript parser implementation
3. Create `internal/parser/javascript_parser.go` — JavaScript parser implementation
4. Create `internal/parser/testdata/sample.ts` — TypeScript test fixture with:
   - Regular functions with type annotations
   - Arrow functions assigned to const
   - Classes with inheritance and access modifiers
   - Interfaces with extends
   - Type aliases
   - Enums
   - Various import styles
   - Export patterns
   - Test functions (describe/it blocks)
   - JSDoc comments
5. Create `internal/parser/testdata/sample.js` — JavaScript test fixture with:
   - Regular and arrow functions
   - Classes with extends
   - ES6 imports and CommonJS require/exports
   - Test functions
   - JSDoc comments
6. Create `internal/parser/typescript_parser_test.go` — parse sample.ts, assert all symbols
7. Create `internal/parser/javascript_parser_test.go` — parse sample.js, assert all symbols
8. Register both parsers in the registry (add to NewRegistry or a RegisterDefaults function)
9. Run all tests including existing Go/Python tests to ensure no regressions

## Testing Requirements
- Unit test: TypeScript parser extracts all symbol types from sample.ts with correct names and line numbers
- Unit test: TypeScript parser captures type annotations in function signatures
- Unit test: TypeScript parser captures JSDoc docstrings
- Unit test: TypeScript parser identifies test functions (describe/it/test)
- Unit test: JavaScript parser extracts functions, classes, imports (ES6 + CommonJS) from sample.js
- Unit test: JavaScript parser handles arrow functions assigned to variables
- Unit test: Registry routes .ts → TypeScript, .js → JavaScript
- Regression test: existing Go and Python parser tests still pass

## Files to Create/Modify
- `internal/parser/typescript_parser.go` — TypeScript tree-sitter parser
- `internal/parser/javascript_parser.go` — JavaScript tree-sitter parser
- `internal/parser/typescript_parser_test.go` — TypeScript parser tests
- `internal/parser/javascript_parser_test.go` — JavaScript parser tests
- `internal/parser/testdata/sample.ts` — TypeScript test fixture
- `internal/parser/testdata/sample.js` — JavaScript test fixture
- `internal/parser/registry.go` — register TS/JS parsers (modify existing file from TASK-005)

## Notes
- TypeScript and JavaScript share many tree-sitter node types. Consider extracting common logic into helper functions to avoid duplication. Both grammars use `function_declaration`, `class_declaration`, `arrow_function`, `import_statement`, etc.
- The TypeScript grammar in `smacker/go-tree-sitter` is at `typescript/typescript` (not just `typescript`). Make sure to use the correct import path.
- For arrow functions, the variable name is the symbol name: `const handler = (req: Request) => { ... }` → symbol name is "handler".
- TSX/JSX files may contain JSX elements. The parser should ignore JSX markup and only extract code symbols.
- CommonJS `module.exports = { ... }` should extract the object's properties as exported symbols.
- Do NOT modify any files created by TASK-005. Only add new files and register the new parsers in the registry.
