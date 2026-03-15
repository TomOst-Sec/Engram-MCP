# TASK-021: Ruby + PHP Tree-Sitter Grammars

**Priority:** P1
**Assigned:** bravo
**Milestone:** M2: Core Features
**Dependencies:** TASK-005
**Status:** active
**Created:** 2026-03-15
**Author:** atlas

## Context
Milestone 2 targets the remaining 9 languages beyond the MVP 6. This task adds Ruby and PHP — two widely-used languages in web development. These follow the exact same pattern as the existing Go, Python, TypeScript, JavaScript parsers (TASK-005, TASK-009) and the Rust/Java parsers (TASK-014). The parser framework and registry are established; this task is pure pattern implementation.

## Specification
Add Ruby and PHP parsers to the `internal/parser` package, following the exact patterns established by existing parsers.

### Dependencies
- `github.com/smacker/go-tree-sitter/ruby` — Ruby grammar
- `github.com/smacker/go-tree-sitter/php` — PHP grammar

### Ruby Parser Extractions
Parse `.rb` files and extract:
- **Methods:** `def method_name(params)` — capture name, parameters, leading `#` comments
- **Class methods:** `def self.method_name` — capture as type "function" with `self.` prefix
- **Classes:** `class Name < Base` — capture name, superclass, leading `#` comments
- **Modules:** `module Name` — capture name as type "type" (modules are Ruby's namespaces/mixins)
- **Includes/Extends:** `include ModuleName`, `extend ModuleName` — capture as type "import"
- **Requires:** `require 'library'`, `require_relative 'file'` — capture as type "import"
- **Constants:** `CONSTANT_NAME = value` (all-caps names) — capture as type "type"
- **Blocks/Procs:** `define_method(:name)` — skip these for MVP, too complex
- **Test methods:** methods starting with `test_` or inside `describe`/`it` blocks (RSpec) — mark as type "test"
- **Attr accessors:** `attr_reader :name`, `attr_accessor :name` — capture as type "type" (property-like)

### PHP Parser Extractions
Parse `.php` files and extract:
- **Functions:** `function name($params): ReturnType` — capture name, full signature with type hints, PHPDoc (`/** ... */`)
- **Methods:** inside class bodies — capture class + method name, visibility (public/private/protected), return type, PHPDoc
- **Classes:** `class Name extends Base implements Interface1, Interface2` — capture name, parent, interfaces, PHPDoc
- **Interfaces:** `interface Name extends Other` — capture name, method signatures
- **Traits:** `trait Name` — capture name, methods
- **Enums:** `enum Name: string` (PHP 8.1+) — capture name, backed type
- **Namespaces:** `namespace App\Http\Controllers;` — capture as context for symbol naming
- **Imports:** `use App\Models\User;` and `use Trait;` — capture all use statements
- **Constants:** `const NAME = value` and `define('NAME', value)` — capture as type "type"
- **Test methods:** methods with `@test` PHPDoc tag or names starting with `test` in classes extending `TestCase` — mark as type "test"
- **Constructors:** `__construct($params)` — capture as type "constructor"

### Symbol Naming
- Ruby: `ClassName#method_name` for instance methods, `ClassName.method_name` for class methods
- PHP: `ClassName::methodName` for methods (following PHP convention)

### Registration
Register both parsers in the registry where existing parsers are registered.

## Acceptance Criteria
- [ ] Ruby parser extracts methods, classes, modules, requires, includes, and constants from .rb files
- [ ] PHP parser extracts functions, methods, classes, interfaces, traits, namespaces, use statements, and enums from .php files
- [ ] Registry routes `.rb` files to Ruby parser and `.php` files to PHP parser
- [ ] Symbols include accurate start/end line numbers (1-based)
- [ ] Ruby parser captures `#` comments as docstrings
- [ ] PHP parser captures PHPDoc `/** ... */` comments
- [ ] Test methods are correctly identified in both languages
- [ ] PHP parser captures visibility modifiers (public/private/protected)
- [ ] All existing parser tests still pass (no regressions)
- [ ] All new tests pass

## Implementation Steps
1. `go get github.com/smacker/go-tree-sitter/ruby github.com/smacker/go-tree-sitter/php`
2. Create `internal/parser/ruby_parser.go` — Ruby parser implementing Parser interface
3. Create `internal/parser/php_parser.go` — PHP parser implementing Parser interface
4. Create `internal/parser/testdata/sample.rb` — Ruby test fixture with:
   - Class with inheritance
   - Instance methods and class methods
   - Module definition with include
   - require and require_relative
   - `#` comments above methods
   - Constants (ALL_CAPS)
   - test_ prefixed methods
   - attr_reader/attr_accessor
5. Create `internal/parser/testdata/sample.php` — PHP test fixture with:
   - Namespace declaration
   - Class with extends and implements
   - Methods with visibility and return types
   - Interface definition
   - Trait definition
   - Use statements (namespace and trait)
   - PHPDoc comments
   - Enum (PHP 8.1 style)
   - Constructor (__construct)
   - test-prefixed methods
6. Create `internal/parser/ruby_parser_test.go` — parse sample.rb, assert all symbols
7. Create `internal/parser/php_parser_test.go` — parse sample.php, assert all symbols
8. Register both parsers in the registry
9. Run ALL parser tests (all languages) to ensure no regressions

## Testing Requirements
- Unit test: Ruby parser extracts methods with correct names and line numbers
- Unit test: Ruby parser captures classes with inheritance
- Unit test: Ruby parser captures module definitions
- Unit test: Ruby parser captures `#` leading comments as docstrings
- Unit test: Ruby parser identifies test_ methods
- Unit test: Ruby parser captures require statements
- Unit test: PHP parser extracts functions with type hints
- Unit test: PHP parser captures class methods with visibility modifiers
- Unit test: PHP parser captures interfaces and traits
- Unit test: PHP parser captures PHPDoc comments
- Unit test: PHP parser identifies test methods
- Unit test: PHP parser handles namespace context
- Unit test: Registry routes .rb → Ruby, .php → PHP
- Regression test: existing parser tests all still pass

## Files to Create/Modify
- `internal/parser/ruby_parser.go` — Ruby tree-sitter parser
- `internal/parser/php_parser.go` — PHP tree-sitter parser
- `internal/parser/ruby_parser_test.go` — Ruby parser tests
- `internal/parser/php_parser_test.go` — PHP parser tests
- `internal/parser/testdata/sample.rb` — Ruby test fixture
- `internal/parser/testdata/sample.php` — PHP test fixture
- `internal/parser/registry.go` — register Ruby + PHP parsers (or wherever parsers are registered)

## Notes
- Follow the EXACT same patterns as the existing parsers. Study `go_parser.go` and `python_parser.go` as reference.
- Ruby tree-sitter node types: `method`, `singleton_method`, `class`, `module`, `call` (for require/include), `assignment` (for constants). Check tree-sitter-ruby grammar for exact names.
- PHP tree-sitter node types: `function_definition`, `method_declaration`, `class_declaration`, `interface_declaration`, `trait_declaration`, `namespace_definition`, `use_declaration`, `enum_declaration`. Check tree-sitter-php grammar.
- The PHP grammar from smacker/go-tree-sitter may be at `github.com/smacker/go-tree-sitter/php` or `github.com/smacker/go-tree-sitter/php/php`. Check go.sum or documentation. If the import path differs, use whatever works with `go get`.
- Ruby uses `#` for single-line comments. Collect consecutive `#` lines above a symbol as the docstring.
- PHP files must start with `<?php` for tree-sitter to parse them. Include this in the test fixture.
- Do NOT modify any existing parser files beyond adding registration entries.
