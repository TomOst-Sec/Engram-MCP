# TASK-016: C# Tree-Sitter Grammar

**Priority:** P1
**Assigned:** bravo
**Milestone:** M1: MVP
**Dependencies:** TASK-005
**Status:** queued
**Created:** 2026-03-15
**Author:** atlas

## Context
Engram's MVP targets 6 languages. With Go, Python, TypeScript, JavaScript (TASK-005, TASK-009), and Rust + Java (TASK-014), we have the first 6 covered. However, the GOALS.md specifically lists C# as one of the initial 6: "TypeScript, JavaScript, Python, Go, Rust, Java, C#". This task adds C# as the 7th language for a complete coverage of the most common languages in AI-assisted development. C# is critical for the .NET developer segment.

## Specification
Add a C# parser to the `internal/parser` package.

### Dependencies
- `github.com/smacker/go-tree-sitter/csharp` — C# grammar

### C# Parser Extractions
Parse `.cs` files and extract:
- **Methods:** inside class/struct bodies — capture class + method name, full signature with type annotations, access modifier (public/private/protected/internal), XML doc comments (`/// <summary>`)
- **Classes:** `class Name : Base, IInterface` — capture name, base class, interfaces, access modifier, XML doc comments
- **Structs:** `struct Name : IInterface` — capture name, interfaces, access modifier
- **Interfaces:** `interface IName : IOther` — capture name, extended interfaces, method signatures, XML doc comments
- **Enums:** `enum Name { ... }` — capture name, members
- **Properties:** `public Type Name { get; set; }` — capture name, type, accessors
- **Namespaces:** `namespace Name.Space { ... }` — capture as module context for symbol naming
- **Imports:** `using Namespace;` and `using static Class;` — capture all using directives
- **Test methods:** `[Test]`, `[TestMethod]`, `[Fact]`, `[Theory]` attributed methods — mark as type "test"
- **Constructors:** `ClassName(params)` — capture as type "constructor"
- **Records:** `record Name(params)` (C# 9+) — capture name, parameters

### Symbol Naming
C# uses namespaces extensively. Symbol names should include the class context:
- Method: `ClassName.MethodName`
- Nested class: `OuterClass.InnerClass`
- Property: `ClassName.PropertyName`

## Acceptance Criteria
- [ ] C# parser extracts methods, classes, structs, interfaces, enums, properties, namespaces, imports, tests, and constructors from .cs files
- [ ] Registry routes `.cs` files to C# parser
- [ ] Symbols include accurate start/end line numbers (1-based)
- [ ] C# parser captures XML doc comments (`/// <summary>`)
- [ ] C# parser captures access modifiers in method/class metadata
- [ ] Test methods with `[Test]`, `[Fact]`, `[TestMethod]` are identified
- [ ] Signatures include full type information (generics, nullable, etc.)
- [ ] All existing parser tests still pass (no regressions)
- [ ] All new tests pass

## Implementation Steps
1. `go get github.com/smacker/go-tree-sitter/csharp`
2. Create `internal/parser/csharp_parser.go` — C# parser implementing Parser interface
3. Create `internal/parser/testdata/Sample.cs` — C# test fixture with:
   - Namespace declaration
   - Public class with inheritance
   - Methods with various access modifiers
   - Properties with getters/setters
   - Interface definition
   - Struct definition
   - Enum definition
   - Using statements
   - `[Test]` and `[Fact]` attributed methods
   - Constructor
   - XML doc comments (`/// <summary>`)
   - Generics (`List<T>`, `Dictionary<TKey, TValue>`)
   - Record type
4. Create `internal/parser/csharp_parser_test.go` — parse Sample.cs, assert all symbols
5. Register C# parser in the registry
6. Run ALL parser tests to ensure no regressions

## Testing Requirements
- Unit test: C# parser extracts classes with correct inheritance info
- Unit test: C# parser captures methods with access modifiers and return types
- Unit test: C# parser captures XML doc comments
- Unit test: C# parser identifies `[Test]` and `[Fact]` attributed methods as test type
- Unit test: C# parser handles properties with get/set
- Unit test: C# parser captures constructors
- Unit test: C# parser handles using statements
- Unit test: C# parser handles generics in signatures
- Unit test: Registry routes .cs → C#
- Regression test: all existing parser tests still pass

## Files to Create/Modify
- `internal/parser/csharp_parser.go` — C# tree-sitter parser
- `internal/parser/csharp_parser_test.go` — C# parser tests
- `internal/parser/testdata/Sample.cs` — C# test fixture
- `internal/parser/registry.go` — register C# parser

## Notes
- Follow the exact same patterns as the Rust and Java parsers (TASK-014). The structure should be nearly identical — only the tree-sitter grammar and extraction queries differ.
- C# tree-sitter node types: `class_declaration`, `method_declaration`, `interface_declaration`, `struct_declaration`, `enum_declaration`, `property_declaration`, `namespace_declaration`, `using_directive`, `constructor_declaration`, `record_declaration`.
- C# XML doc comments start with `///`. Collect all consecutive `///` lines above a symbol as the docstring.
- C# access modifiers are part of the declaration node (child nodes of type `modifier`). Extract `public`, `private`, `protected`, `internal`, `static`, `abstract`, `virtual`, `override`.
- Generic type parameters should be preserved in signatures: `public List<T> GetItems<T>() where T : IComparable`.
- Do NOT modify any existing parser files beyond adding the registration entry.
