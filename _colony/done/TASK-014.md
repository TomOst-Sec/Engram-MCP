# TASK-014: Rust and Java Tree-Sitter Grammars

**Priority:** P1
**Assigned:** bravo
**Milestone:** M1: MVP
**Dependencies:** TASK-005
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
Engram's MVP targets 6 languages: Go, Python, TypeScript, JavaScript, Rust, and Java. The parser framework (TASK-005) and TS/JS grammars (TASK-009) give us 4 languages. This task adds Rust and Java to complete the MVP language set. Both are popular in backend/systems development and represent important user segments for Engram.

## Specification
Add Rust and Java parsers to the `internal/parser` package, following the exact patterns established by the existing Go, Python, TypeScript, and JavaScript parsers.

### Dependencies
- `github.com/smacker/go-tree-sitter/rust` — Rust grammar
- `github.com/smacker/go-tree-sitter/java` — Java grammar

### Rust Parser Extractions
Parse `.rs` files and extract:
- **Functions:** `fn name(params) -> ReturnType` — capture name, full signature with type annotations, doc comments (`///` and `//!`)
- **Methods:** `fn name(&self, params)` inside `impl` blocks — capture impl type + method name, full signature, `pub` visibility
- **Structs:** `struct Name { ... }` — capture name, fields (pub/private), doc comments, derive macros
- **Enums:** `enum Name { Variant1, Variant2(Type) }` — capture name, variants
- **Traits:** `trait Name { ... }` — capture name, method signatures, doc comments
- **Impl blocks:** `impl Trait for Type` — capture type, trait (if any), method list
- **Imports:** `use path::to::module` and `use path::{item1, item2}` — capture all use paths
- **Test functions:** `#[test] fn test_xxx()` and functions inside `#[cfg(test)] mod tests` — mark as type "test"
- **Macros:** `macro_rules! name` — capture name, mark as type "macro"

### Java Parser Extractions
Parse `.java` files and extract:
- **Methods:** in class bodies — capture class + method name, full signature with type annotations, access modifier (public/private/protected), Javadoc (`/** ... */`)
- **Classes:** `class Name extends Base implements Interface1, Interface2` — capture name, base class, interfaces, Javadoc, access modifier
- **Interfaces:** `interface Name extends Other` — capture name, method signatures, Javadoc
- **Enums:** `enum Name { ... }` — capture name, constants
- **Imports:** `import package.Class` and `import package.*` — capture all import paths
- **Annotations:** `@Annotation` on classes/methods — include in symbol metadata
- **Test methods:** `@Test void testXxx()` or methods in classes matching `*Test.java` — mark as type "test"
- **Constructors:** `ClassName(params)` — capture as type "constructor"

### Registration
Both parsers must register with the parser registry. If there's a `NewDefaultRegistry()` function, add them there. Otherwise, add them where the existing parsers are registered.

## Acceptance Criteria
- [ ] Rust parser extracts functions, methods, structs, enums, traits, impl blocks, imports, tests, and macros from .rs files
- [ ] Java parser extracts methods, classes, interfaces, enums, imports, annotations, tests, and constructors from .java files
- [ ] Registry routes `.rs` files to Rust parser and `.java` files to Java parser
- [ ] Symbols include accurate start/end line numbers (1-based)
- [ ] Rust parser captures `///` doc comments
- [ ] Java parser captures Javadoc `/** ... */` comments
- [ ] Symbols include full type-annotated signatures
- [ ] Test functions are correctly identified in both languages
- [ ] All existing parser tests still pass (no regressions)
- [ ] All new tests pass

## Implementation Steps
1. `go get github.com/smacker/go-tree-sitter/rust github.com/smacker/go-tree-sitter/java`
2. Create `internal/parser/rust_parser.go` — Rust parser implementing Parser interface
3. Create `internal/parser/java_parser.go` — Java parser implementing Parser interface
4. Create `internal/parser/testdata/sample.rs` — Rust test fixture with:
   - Public and private functions with type annotations and lifetimes
   - Struct with derive macros and fields
   - Enum with tuple and struct variants
   - Trait definition with method signatures
   - Impl block (inherent and trait impl)
   - Use statements (simple and nested)
   - `#[test]` functions and `#[cfg(test)]` module
   - `///` doc comments
   - `macro_rules!` definition
5. Create `internal/parser/testdata/Sample.java` — Java test fixture with:
   - Public class with extends and implements
   - Methods with various access modifiers and return types
   - Interface with method signatures
   - Enum with constants
   - Import statements
   - `@Test` annotated methods
   - Constructor
   - Javadoc comments
   - Nested/inner class
6. Create `internal/parser/rust_parser_test.go` — parse sample.rs, assert all symbols
7. Create `internal/parser/java_parser_test.go` — parse Sample.java, assert all symbols
8. Register both parsers in the registry
9. Run ALL parser tests (go, python, ts, js, rust, java) to ensure no regressions

## Testing Requirements
- Unit test: Rust parser extracts functions with correct signatures and line numbers
- Unit test: Rust parser captures struct fields and derive macros
- Unit test: Rust parser identifies trait definitions and impl blocks
- Unit test: Rust parser captures `///` doc comments
- Unit test: Rust parser identifies `#[test]` functions
- Unit test: Java parser extracts classes with inheritance info
- Unit test: Java parser captures method signatures with access modifiers
- Unit test: Java parser captures Javadoc comments
- Unit test: Java parser identifies `@Test` methods
- Unit test: Java parser handles constructors
- Unit test: Registry routes .rs → Rust, .java → Java
- Regression test: existing Go, Python, TS, JS parser tests still pass

## Files to Create/Modify
- `internal/parser/rust_parser.go` — Rust tree-sitter parser
- `internal/parser/java_parser.go` — Java tree-sitter parser
- `internal/parser/rust_parser_test.go` — Rust parser tests
- `internal/parser/java_parser_test.go` — Java parser tests
- `internal/parser/testdata/sample.rs` — Rust test fixture
- `internal/parser/testdata/Sample.java` — Java test fixture
- `internal/parser/registry.go` — register Rust + Java parsers (or wherever parsers are registered)

## Notes
- Follow the EXACT same patterns as the existing parsers. Study `go_parser.go` and `typescript_parser.go` to understand the tree-sitter query patterns, symbol extraction logic, and test structure.
- Rust tree-sitter node types: `function_item`, `impl_item`, `struct_item`, `enum_item`, `trait_item`, `use_declaration`, `macro_definition`. Check the tree-sitter-rust grammar for exact node type names.
- Java tree-sitter node types: `class_declaration`, `method_declaration`, `interface_declaration`, `enum_declaration`, `import_declaration`, `constructor_declaration`. Check the tree-sitter-java grammar.
- Rust lifetimes (`'a`) in signatures should be preserved as-is.
- Java generic types (`List<String>`) should be preserved in signatures.
- The Java test fixture file should be named `Sample.java` (capital S) to follow Java conventions.
- Do NOT modify any existing parser files beyond adding registration entries. Only add new files.

---
## Completion Notes
- **Completed by:** bravo-2
- **Date:** 2026-03-15 16:07:55
- **Branch:** task/014

---
## Completion Notes
- **Completed by:** bravo-2
- **Date:** 2026-03-15 16:15:00
- **Branch:** task/014
