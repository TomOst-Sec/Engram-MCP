# TASK-005: Tree-Sitter Parser Framework + Go and Python Grammars

**Priority:** P1
**Assigned:** bravo
**Milestone:** M1: MVP
**Dependencies:** TASK-003
**Status:** done
**Created:** 2026-03-15
**Author:** atlas

## Context
Engram's core value proposition is deep understanding of codebases through AST (Abstract Syntax Tree) parsing. Tree-sitter provides fast, incremental parsing across many languages. This task creates the parser framework — a pluggable system where each language has a grammar and a set of extraction queries — and implements the first two languages: Go and Python. This forms the foundation for all code indexing.

## Specification
Create the `internal/parser` package with a tree-sitter-based AST parser.

### Dependencies
- `github.com/smacker/go-tree-sitter` — Go bindings for tree-sitter
- `github.com/smacker/go-tree-sitter/golang` — Go grammar
- `github.com/smacker/go-tree-sitter/python` — Python grammar

### Core Types
```go
// Symbol represents an extracted code symbol from AST parsing
type Symbol struct {
    Name       string   // symbol name (e.g., "HandleRequest", "UserModel")
    Type       string   // function, method, type, class, interface, import, export, test
    Language   string   // go, python, etc.
    Signature  string   // full signature (e.g., "func HandleRequest(w http.ResponseWriter, r *http.Request) error")
    Docstring  string   // leading comment/docstring
    StartLine  int      // 1-based line number
    EndLine    int      // 1-based line number
    FilePath   string   // relative to repo root
    BodyHash   string   // SHA256 of the symbol body text
}

// Parser interface — one implementation per language
type Parser interface {
    Language() string
    Extensions() []string                     // e.g., [".go"] or [".py", ".pyi"]
    Parse(filePath string, source []byte) ([]Symbol, error)
}

// Registry manages available parsers
type Registry struct { ... }
func NewRegistry() *Registry
func (r *Registry) Register(p Parser)
func (r *Registry) ParserFor(filePath string) (Parser, bool)   // match by extension
func (r *Registry) ParseFile(filePath string, source []byte) ([]Symbol, error)
func (r *Registry) SupportedLanguages() []string
```

### Go Parser Extractions
Parse Go source files and extract:
- **Functions:** top-level `func name(...)` — capture name, full signature, docstring (preceding comment), start/end lines
- **Methods:** `func (recv Type) name(...)` — capture receiver, name, full signature
- **Types:** `type Name struct/interface/...` — capture name, kind (struct/interface/alias), docstring
- **Interfaces:** `type Name interface { ... }` — capture name, method list
- **Imports:** all import paths
- **Test functions:** `func TestXxx(t *testing.T)` — mark as type "test"

### Python Parser Extractions
Parse Python source files and extract:
- **Functions:** `def name(...)` at module level — capture name, signature (params + type hints), docstring (first string literal in body)
- **Methods:** `def name(self, ...)` inside class — capture class name + method name, signature
- **Classes:** `class Name(bases):` — capture name, base classes, docstring
- **Imports:** `import x`, `from x import y` — capture all import paths
- **Test functions:** `def test_xxx(...)` or functions in files matching `test_*.py` — mark as type "test"

### Storage Integration
```go
// StoreSymbols writes parsed symbols to the code_index table
func StoreSymbols(store *storage.Store, fileHash string, symbols []Symbol) error

// DeleteFileSymbols removes all symbols for a file path (for re-indexing)
func DeleteFileSymbols(store *storage.Store, filePath string) error

// GetFileHash retrieves the stored hash for a file to check if re-indexing is needed
func GetFileHash(store *storage.Store, filePath string) (string, error)
```

## Acceptance Criteria
- [ ] Go parser extracts functions, methods, types, interfaces, imports, and test functions from a Go source file
- [ ] Python parser extracts functions, methods, classes, imports, and test functions from a Python source file
- [ ] Registry correctly routes `.go` files to Go parser and `.py` files to Python parser
- [ ] Symbols include accurate start/end line numbers (1-based)
- [ ] Symbols include docstrings when present
- [ ] Symbols include full signatures
- [ ] BodyHash is a consistent SHA256 of the symbol's source text
- [ ] StoreSymbols writes to SQLite code_index table successfully
- [ ] FTS5 index is searchable after StoreSymbols
- [ ] All tests pass

## Implementation Steps
1. `go get github.com/smacker/go-tree-sitter github.com/smacker/go-tree-sitter/golang github.com/smacker/go-tree-sitter/python`
2. Create `internal/parser/types.go` — Symbol struct, Parser interface
3. Create `internal/parser/registry.go` — Registry with Register, ParserFor, ParseFile
4. Create `internal/parser/go_parser.go` — Go language parser using tree-sitter
5. Create `internal/parser/python_parser.go` — Python language parser using tree-sitter
6. Create `internal/parser/store.go` — StoreSymbols, DeleteFileSymbols, GetFileHash
7. Create `internal/parser/testdata/sample.go` — sample Go file with functions, methods, types, interfaces, imports, tests
8. Create `internal/parser/testdata/sample.py` — sample Python file with functions, methods, classes, imports, tests
9. Create `internal/parser/go_parser_test.go` — parse sample.go, assert all symbols extracted correctly
10. Create `internal/parser/python_parser_test.go` — parse sample.py, assert all symbols extracted correctly
11. Create `internal/parser/registry_test.go` — test registration, routing, SupportedLanguages
12. Create `internal/parser/store_test.go` — test StoreSymbols writes correctly, FTS5 searchable
13. Run all tests

## Testing Requirements
- Unit test: Go parser on sample.go extracts expected number of symbols with correct names, types, line numbers
- Unit test: Go parser captures function signatures including params and return types
- Unit test: Go parser captures docstrings (preceding comments)
- Unit test: Python parser on sample.py extracts functions, classes, methods with correct attributes
- Unit test: Python parser captures docstrings (triple-quoted strings)
- Unit test: Registry routes .go to Go parser, .py to Python parser, returns false for .rs
- Unit test: StoreSymbols + FTS5 query returns matching symbol
- Golden test: parse sample.go → compare extracted symbols against expected JSON snapshot

## Files to Create/Modify
- `internal/parser/types.go` — Symbol struct, Parser interface
- `internal/parser/registry.go` — Registry struct and methods
- `internal/parser/go_parser.go` — Go tree-sitter parser
- `internal/parser/python_parser.go` — Python tree-sitter parser
- `internal/parser/store.go` — StoreSymbols, DeleteFileSymbols, GetFileHash
- `internal/parser/go_parser_test.go` — Go parser tests
- `internal/parser/python_parser_test.go` — Python parser tests
- `internal/parser/registry_test.go` — registry tests
- `internal/parser/store_test.go` — storage integration tests
- `internal/parser/testdata/sample.go` — test fixture
- `internal/parser/testdata/sample.py` — test fixture

## Notes
- tree-sitter queries use S-expression syntax. Study the Go and Python grammar node types. For Go, function declarations are `function_declaration`, methods are `method_declaration`, types are `type_declaration`. For Python, functions are `function_definition`, classes are `class_definition`.
- Use `sitter.NewParser()` and set the language with `parser.SetLanguage(golang.GetLanguage())`.
- Walk the tree with `tree.RootNode()` and iterate children, or use tree-sitter queries for targeted extraction.
- The sample test files should be realistic — include edge cases like multi-line signatures, multiple return values (Go), decorated functions (Python), nested classes, etc.
- For BodyHash, extract the source text of the node (using byte range from the tree) and SHA256 it.

---
## Completion Notes
- **Completed by:** bravo-2
- **Date:** 2026-03-15 15:12:24
- **Branch:** task/005

---
## Completion Notes
- **Completed by:** bravo-2
- **Date:** 2026-03-15 15:19:43
- **Branch:** task/005
