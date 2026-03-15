# TASK-026: Swift + Kotlin Tree-Sitter Grammars

**Priority:** P1
**Assigned:** bravo
**Milestone:** M2: Core Features
**Dependencies:** TASK-005
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Milestone 2 adds the remaining 9 languages. Ruby + PHP are in TASK-021. This task adds Swift and Kotlin — important for mobile development (iOS and Android). Both follow the established parser pattern from existing Go, Python, TypeScript, JavaScript, Rust, and Java parsers.

## Specification
Add Swift and Kotlin parsers to `internal/parser/`.

### Dependencies
- `github.com/smacker/go-tree-sitter/swift` — Swift grammar
- `github.com/smacker/go-tree-sitter/kotlin` — Kotlin grammar (may be at a different import path — check go-tree-sitter docs)

### Swift Parser Extractions
Parse `.swift` files and extract:
- **Functions:** `func name(params) -> ReturnType` — capture name, full signature, doc comments (`///` and `/** */`)
- **Methods:** inside class/struct/enum bodies — capture type + method name, access level
- **Classes:** `class Name: Base, Protocol` — capture name, superclass, protocols
- **Structs:** `struct Name: Protocol` — capture name, protocols
- **Enums:** `enum Name { case a, b }` — capture name, cases
- **Protocols:** `protocol Name` — capture name, method signatures
- **Extensions:** `extension Type: Protocol` — capture type, added protocol
- **Imports:** `import Module` — capture module name
- **Properties:** `var name: Type` / `let name: Type` — capture in class/struct context
- **Test functions:** functions starting with `test` in `XCTestCase` subclasses — mark as type "test"
- **Initializers:** `init(params)` — capture as type "constructor"

### Kotlin Parser Extractions
Parse `.kt` and `.kts` files and extract:
- **Functions:** `fun name(params): ReturnType` — capture name, full signature, KDoc (`/** */`)
- **Methods:** inside class bodies — capture class + method name, visibility (public/private/internal/protected)
- **Classes:** `class Name : Base(), Interface` — capture name, superclass, interfaces, KDoc
- **Objects:** `object Name` and `companion object` — capture name
- **Data classes:** `data class Name(val x: Type)` — capture name, properties
- **Interfaces:** `interface Name` — capture name, method signatures
- **Sealed classes:** `sealed class Name` — capture name, subclasses
- **Imports:** `import package.Class` — capture import paths
- **Properties:** `val name: Type` / `var name: Type` — capture in class context
- **Test functions:** `@Test fun testName()` — mark as type "test"
- **Extensions:** `fun Type.name()` — capture as function with receiver type in name

### Registration
Add both parsers to `NewDefaultRegistry()` in `registry.go`.

## Acceptance Criteria
- [ ] Swift parser extracts functions, classes, structs, enums, protocols, imports from .swift files
- [ ] Kotlin parser extracts functions, classes, objects, data classes, interfaces, imports from .kt files
- [ ] Registry routes `.swift` → Swift, `.kt`/`.kts` → Kotlin
- [ ] Symbols include accurate line numbers and signatures
- [ ] Swift parser captures `///` doc comments
- [ ] Kotlin parser captures KDoc comments
- [ ] Test functions identified in both languages
- [ ] All existing parser tests still pass
- [ ] All new tests pass

## Implementation Steps
1. `go get github.com/smacker/go-tree-sitter/swift` (check if kotlin is available, try `github.com/smacker/go-tree-sitter/kotlin`)
2. Create `internal/parser/swift_parser.go`
3. Create `internal/parser/kotlin_parser.go`
4. Create `internal/parser/testdata/sample.swift` — test fixture with classes, structs, enums, protocols, functions, imports, test functions
5. Create `internal/parser/testdata/sample.kt` — test fixture with classes, data classes, objects, interfaces, functions, imports, @Test methods
6. Create `internal/parser/swift_parser_test.go`
7. Create `internal/parser/kotlin_parser_test.go`
8. Register both in `NewDefaultRegistry()` in `registry.go`
9. Run ALL parser tests

## Testing Requirements
- Unit test: Swift parser extracts functions with correct signatures
- Unit test: Swift parser captures class inheritance and protocol conformance
- Unit test: Swift parser identifies test functions
- Unit test: Kotlin parser extracts functions with return types
- Unit test: Kotlin parser captures data classes
- Unit test: Kotlin parser identifies @Test methods
- Unit test: Registry routes .swift and .kt correctly
- Regression test: all existing parser tests pass

## Files to Create/Modify
- `internal/parser/swift_parser.go` — Swift parser
- `internal/parser/kotlin_parser.go` — Kotlin parser
- `internal/parser/swift_parser_test.go` — tests
- `internal/parser/kotlin_parser_test.go` — tests
- `internal/parser/testdata/sample.swift` — fixture
- `internal/parser/testdata/sample.kt` — fixture
- `internal/parser/registry.go` — add registrations

## Notes
- Follow the exact patterns from existing parsers. Study `rust_parser.go` and `java_parser.go` (TASK-014) as the closest references.
- Swift tree-sitter node types: `function_declaration`, `class_declaration`, `struct_declaration`, `enum_declaration`, `protocol_declaration`, `import_declaration`, `property_declaration`.
- Kotlin tree-sitter node types: `function_declaration`, `class_declaration`, `object_declaration`, `interface_declaration`, `import_header`, `property_declaration`.
- If a go-tree-sitter grammar doesn't exist for a language, check for alternative packages or skip that language with a TODO comment. Do NOT block the entire task on one missing grammar.
- Kotlin files can be `.kt` or `.kts` (script). Register both extensions.
