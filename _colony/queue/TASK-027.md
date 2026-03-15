# TASK-027: C + C++ Tree-Sitter Grammars

**Priority:** P1
**Assigned:** bravo
**Milestone:** M2: Core Features
**Dependencies:** TASK-005
**Status:** queued
**Created:** 2026-03-15
**Author:** atlas

## Context
Continuing language expansion for Milestone 2. C and C++ are foundational systems languages used extensively in embedded, OS, game, and infrastructure development. They share some extraction patterns but have distinct parsers. Both grammars are well-supported by tree-sitter.

## Specification
Add C and C++ parsers to `internal/parser/`.

### Dependencies
- `github.com/smacker/go-tree-sitter/c` — C grammar
- `github.com/smacker/go-tree-sitter/cpp` — C++ grammar

### C Parser Extractions
Parse `.c` and `.h` files and extract:
- **Functions:** `returnType functionName(params)` — capture name, full signature with types, leading `//` or `/* */` comments
- **Function declarations:** in .h files, capture prototype signatures
- **Structs:** `struct Name { ... }` and `typedef struct { ... } Name` — capture name, fields
- **Enums:** `enum Name { ... }` and `typedef enum { ... } Name` — capture name, values
- **Typedefs:** `typedef oldType newName` — capture as type
- **Macros:** `#define NAME(...)` (function-like macros) — capture name, mark as type "macro"
- **Includes:** `#include <header>` and `#include "header"` — capture as type "import"
- **Global variables:** `extern Type name` — capture as type "type"

### C++ Parser Extractions
Parse `.cpp`, `.hpp`, `.cc`, `.hh`, `.cxx`, `.hxx` files and extract:
- Everything from C, plus:
- **Classes:** `class Name : public Base` — capture name, base classes, access specifiers, `///` or `/** */` comments
- **Methods:** inside class bodies — capture class + method name, virtual/override/const qualifiers
- **Namespaces:** `namespace Name { ... }` — capture as context for symbol naming
- **Templates:** `template<typename T> class/function` — preserve template parameters in signature
- **Constructors/Destructors:** `ClassName(...)` / `~ClassName()` — capture as constructor type
- **Operator overloads:** `operator+(...)` — capture as method
- **Enums:** `enum class Name { ... }` (scoped enums) — capture name
- **Using:** `using namespace std;` and `using Type = Other;` — capture as import

### Symbol Naming
- C: function names are global, no namespace prefix needed
- C++: `Namespace::Class::method` for fully qualified names

### Registration
Register both parsers in `NewDefaultRegistry()`.

## Acceptance Criteria
- [ ] C parser extracts functions, structs, enums, typedefs, macros, includes from .c/.h files
- [ ] C++ parser extracts classes, methods, namespaces, templates, constructors from .cpp/.hpp files
- [ ] Registry routes `.c`/`.h` → C parser, `.cpp`/`.hpp`/`.cc`/`.hh`/`.cxx`/`.hxx` → C++ parser
- [ ] Symbols include accurate line numbers and full signatures with types
- [ ] C parser handles typedef struct pattern
- [ ] C++ parser preserves template parameters in signatures
- [ ] Leading comments captured as docstrings
- [ ] All existing parser tests still pass
- [ ] All new tests pass

## Implementation Steps
1. `go get github.com/smacker/go-tree-sitter/c github.com/smacker/go-tree-sitter/cpp`
2. Create `internal/parser/c_parser.go` — C parser
3. Create `internal/parser/cpp_parser.go` — C++ parser
4. Create `internal/parser/testdata/sample.c` — C test fixture with functions, structs, enums, typedefs, macros, includes
5. Create `internal/parser/testdata/sample.h` — C header test fixture with prototypes and struct declarations
6. Create `internal/parser/testdata/sample.cpp` — C++ test fixture with classes, templates, namespaces, methods
7. Create `internal/parser/c_parser_test.go` — C parser tests
8. Create `internal/parser/cpp_parser_test.go` — C++ parser tests
9. Register in `NewDefaultRegistry()`
10. Run ALL parser tests

## Testing Requirements
- Unit test: C parser extracts functions with correct signatures
- Unit test: C parser handles typedef struct
- Unit test: C parser captures #define macros
- Unit test: C parser captures #include as imports
- Unit test: C++ parser extracts classes with inheritance
- Unit test: C++ parser captures methods with qualifiers (virtual, const, override)
- Unit test: C++ parser handles templates
- Unit test: C++ parser captures namespace context
- Unit test: Registry routes all C/C++ extensions correctly
- Regression test: all existing parser tests pass

## Files to Create/Modify
- `internal/parser/c_parser.go` — C parser
- `internal/parser/cpp_parser.go` — C++ parser
- `internal/parser/c_parser_test.go` — tests
- `internal/parser/cpp_parser_test.go` — tests
- `internal/parser/testdata/sample.c` — C fixture
- `internal/parser/testdata/sample.h` — C header fixture
- `internal/parser/testdata/sample.cpp` — C++ fixture
- `internal/parser/registry.go` — add registrations

## Notes
- Follow exact patterns from existing parsers. Study `go_parser.go` or `rust_parser.go`.
- C tree-sitter node types: `function_definition`, `declaration`, `struct_specifier`, `enum_specifier`, `type_definition`, `preproc_def`, `preproc_include`.
- C++ tree-sitter node types: `class_specifier`, `function_definition`, `namespace_definition`, `template_declaration`, `using_declaration`, `field_declaration`.
- `.h` files are ambiguous (could be C or C++). Route them to the C parser by default. This is a reasonable heuristic since C++ headers typically use `.hpp`.
- C function signatures often span multiple lines. Tree-sitter handles this — just extract the full declarator text.
- C++ templates can be complex (`template<typename T, typename U = int>`). Capture the full template prefix as part of the signature.
