package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
)

// CParser extracts symbols from C source files using tree-sitter.
type CParser struct{}

// NewCParser creates a new C parser.
func NewCParser() *CParser { return &CParser{} }

func (p *CParser) Language() string     { return "c" }
func (p *CParser) Extensions() []string { return []string{".c", ".h"} }

func (p *CParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, c.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		switch child.Type() {
		case "preproc_include":
			symbols = append(symbols, p.parseInclude(filePath, source, child))
		case "preproc_def":
			symbols = append(symbols, p.parseMacro(filePath, source, child))
		case "preproc_function_def":
			symbols = append(symbols, p.parseMacro(filePath, source, child))
		case "function_definition":
			symbols = append(symbols, p.parseFunction(filePath, source, child))
		case "declaration":
			symbols = append(symbols, p.parseDeclaration(filePath, source, child)...)
		case "type_definition":
			symbols = append(symbols, p.parseTypedef(filePath, source, child))
		case "struct_specifier":
			if sym, ok := p.parseStruct(filePath, source, child); ok {
				symbols = append(symbols, sym)
			}
		case "enum_specifier":
			if sym, ok := p.parseEnum(filePath, source, child); ok {
				symbols = append(symbols, sym)
			}
		}
	}
	return symbols, nil
}

func (p *CParser) parseInclude(filePath string, source []byte, node *sitter.Node) Symbol {
	path := ""
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "system_lib_string" || child.Type() == "string_literal" {
			path = child.Content(source)
			break
		}
	}
	return Symbol{
		Name:      "#include " + path,
		Type:      "import",
		Language:  "c",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

func (p *CParser) parseMacro(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "macro",
		Language:  "c",
		Signature: strings.TrimRight(node.Content(source), "\n"),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

func (p *CParser) parseFunction(filePath string, source []byte, node *sitter.Node) Symbol {
	declarator := node.ChildByFieldName("declarator")
	if declarator == nil {
		return Symbol{}
	}

	name := p.getFunctionName(source, declarator)
	sig := p.buildCFuncSignature(source, node)

	return Symbol{
		Name:      name,
		Type:      "function",
		Language:  "c",
		Signature: sig,
		Docstring: getCDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCBody(source, node),
	}
}

func (p *CParser) parseDeclaration(filePath string, source []byte, node *sitter.Node) []Symbol {
	var symbols []Symbol
	// Check for struct/enum declarations within a declaration
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "struct_specifier":
			if sym, ok := p.parseStruct(filePath, source, child); ok {
				symbols = append(symbols, sym)
			}
		case "enum_specifier":
			if sym, ok := p.parseEnum(filePath, source, child); ok {
				symbols = append(symbols, sym)
			}
		}
	}
	return symbols
}

func (p *CParser) parseTypedef(filePath string, source []byte, node *sitter.Node) Symbol {
	// Find the typedef name (last identifier/type_identifier before semicolon)
	name := ""
	for i := int(node.NamedChildCount()) - 1; i >= 0; i-- {
		child := node.NamedChild(i)
		if child.Type() == "type_identifier" || child.Type() == "identifier" {
			name = child.Content(source)
			break
		}
		// For function pointer typedefs, check function_declarator
		if child.Type() == "function_declarator" {
			inner := child.ChildByFieldName("declarator")
			if inner != nil && inner.Type() == "parenthesized_declarator" {
				for j := 0; j < int(inner.NamedChildCount()); j++ {
					sub := inner.NamedChild(j)
					if sub.Type() == "pointer_declarator" {
						name = strings.TrimPrefix(sub.Content(source), "*")
						break
					}
				}
			}
			break
		}
	}

	if name == "" {
		return Symbol{}
	}

	return Symbol{
		Name:      name,
		Type:      "type",
		Language:  "c",
		Signature: strings.TrimSuffix(node.Content(source), ";"),
		Docstring: getCDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCBody(source, node),
	}
}

func (p *CParser) parseStruct(filePath string, source []byte, node *sitter.Node) (Symbol, bool) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}, false
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "type",
		Language:  "c",
		Signature: "struct " + name,
		Docstring: getCDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCBody(source, node),
	}, true
}

func (p *CParser) parseEnum(filePath string, source []byte, node *sitter.Node) (Symbol, bool) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}, false
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "enum",
		Language:  "c",
		Signature: "enum " + name,
		Docstring: getCDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCBody(source, node),
	}, true
}

func (p *CParser) getFunctionName(source []byte, declarator *sitter.Node) string {
	// function_declarator has identifier child
	nameNode := declarator.ChildByFieldName("declarator")
	if nameNode != nil && nameNode.Type() == "identifier" {
		return nameNode.Content(source)
	}
	// Fallback: look for identifier child
	for i := 0; i < int(declarator.NamedChildCount()); i++ {
		child := declarator.NamedChild(i)
		if child.Type() == "identifier" {
			return child.Content(source)
		}
	}
	return declarator.Content(source)
}

func (p *CParser) buildCFuncSignature(source []byte, node *sitter.Node) string {
	// Build signature from type + declarator (without body)
	content := node.Content(source)
	// Truncate at the opening brace
	if idx := strings.Index(content, " {"); idx > 0 {
		return content[:idx]
	}
	if idx := strings.Index(content, "\n{"); idx > 0 {
		return content[:idx]
	}
	return content
}

// getCDocstring extracts comments preceding a node.
func getCDocstring(source []byte, node *sitter.Node) string {
	prev := node.PrevNamedSibling()
	if prev == nil || prev.Type() != "comment" {
		return ""
	}
	if int(node.StartPoint().Row)-int(prev.EndPoint().Row) > 1 {
		return ""
	}

	var comments []string
	cur := prev
	for cur != nil && cur.Type() == "comment" {
		line := cur.Content(source)
		// Strip comment markers
		line = strings.TrimPrefix(line, "///")
		line = strings.TrimPrefix(line, "//")
		line = strings.TrimPrefix(line, "/*")
		line = strings.TrimSuffix(line, "*/")
		line = strings.TrimSpace(line)
		// Clean up multi-line comment lines
		lines := strings.Split(line, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			l = strings.TrimPrefix(l, "* ")
			l = strings.TrimPrefix(l, "*")
			l = strings.TrimSpace(l)
			if l != "" {
				comments = append([]string{l}, comments...)
			}
		}
		next := cur.PrevNamedSibling()
		if next == nil || next.Type() != "comment" {
			break
		}
		if int(cur.StartPoint().Row)-int(next.EndPoint().Row) > 1 {
			break
		}
		cur = next
	}
	return strings.Join(comments, "\n")
}

func hashCBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}

// ParseReferences walks the AST to extract all identifier references from
// function bodies, initializer lists, and top-level expressions. This enables
// tracking direct calls, function pointer assignments, callback arguments,
// and dispatch table entries.
func (p *CParser) ParseReferences(filePath string, source []byte) ([]Reference, error) {
	root := sitter.Parse(source, c.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	lines := strings.Split(string(source), "\n")
	getLine := func(line int) string {
		if line >= 1 && line <= len(lines) {
			return strings.TrimSpace(lines[line-1])
		}
		return ""
	}

	var refs []Reference

	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		switch child.Type() {
		case "function_definition":
			funcName := p.getFuncDefName(source, child)
			body := child.ChildByFieldName("body")
			if body != nil {
				p.extractRefsFromNode(source, body, filePath, funcName, getLine, &refs)
			}
		case "declaration":
			// Top-level declarations: dispatch tables, global pointer assignments
			p.extractRefsFromTopDecl(source, child, filePath, getLine, &refs)
		}
	}

	return refs, nil
}

// getFuncDefName extracts the function name from a function_definition node.
func (p *CParser) getFuncDefName(source []byte, node *sitter.Node) string {
	decl := node.ChildByFieldName("declarator")
	if decl == nil {
		return "<unknown>"
	}
	return p.getFunctionName(source, decl)
}

// extractRefsFromNode recursively walks an AST subtree extracting references.
func (p *CParser) extractRefsFromNode(source []byte, node *sitter.Node, filePath, fromFunc string, getLine func(int) string, refs *[]Reference) {
	switch node.Type() {
	case "call_expression":
		p.extractCallRef(source, node, filePath, fromFunc, getLine, refs)
		// Also walk arguments for callbacks passed as args
		args := node.ChildByFieldName("arguments")
		if args != nil {
			for j := 0; j < int(args.NamedChildCount()); j++ {
				arg := args.NamedChild(j)
				if arg.Type() == "identifier" {
					line := int(arg.StartPoint().Row) + 1
					*refs = append(*refs, Reference{
						ToName:   arg.Content(source),
						Kind:     "callback_arg",
						FromFunc: fromFunc,
						FilePath: filePath,
						Line:     line,
						Context:  getLine(line),
					})
				} else if arg.Type() == "unary_expression" {
					// &func_name
					p.extractAddressOf(source, arg, filePath, fromFunc, getLine, refs)
				} else {
					p.extractRefsFromNode(source, arg, filePath, fromFunc, getLine, refs)
				}
			}
		}
		return // don't double-walk children

	case "init_declarator":
		// variable = value — check for function pointer assignments
		value := node.ChildByFieldName("value")
		if value != nil {
			if value.Type() == "identifier" {
				line := int(value.StartPoint().Row) + 1
				*refs = append(*refs, Reference{
					ToName:   value.Content(source),
					Kind:     "pointer_assign",
					FromFunc: fromFunc,
					FilePath: filePath,
					Line:     line,
					Context:  getLine(line),
				})
			} else if value.Type() == "unary_expression" {
				p.extractAddressOf(source, value, filePath, fromFunc, getLine, refs)
			} else {
				p.extractRefsFromNode(source, value, filePath, fromFunc, getLine, refs)
			}
		}
		return

	case "assignment_expression":
		right := node.ChildByFieldName("right")
		if right != nil {
			if right.Type() == "identifier" {
				line := int(right.StartPoint().Row) + 1
				*refs = append(*refs, Reference{
					ToName:   right.Content(source),
					Kind:     "pointer_assign",
					FromFunc: fromFunc,
					FilePath: filePath,
					Line:     line,
					Context:  getLine(line),
				})
			} else if right.Type() == "unary_expression" {
				p.extractAddressOf(source, right, filePath, fromFunc, getLine, refs)
			} else {
				p.extractRefsFromNode(source, right, filePath, fromFunc, getLine, refs)
			}
		}
		return

	case "initializer_list":
		// Array/struct initializer — recurse to handle nested structs
		p.extractRefsFromInitList(source, node, filePath, fromFunc, "", getLine, refs)
		return
	}

	// Recurse into children
	for j := 0; j < int(node.NamedChildCount()); j++ {
		p.extractRefsFromNode(source, node.NamedChild(j), filePath, fromFunc, getLine, refs)
	}
}

// extractCallRef extracts the function name from a call_expression.
func (p *CParser) extractCallRef(source []byte, node *sitter.Node, filePath, fromFunc string, getLine func(int) string, refs *[]Reference) {
	fn := node.ChildByFieldName("function")
	if fn == nil {
		return
	}
	line := int(fn.StartPoint().Row) + 1

	switch fn.Type() {
	case "identifier":
		*refs = append(*refs, Reference{
			ToName:   fn.Content(source),
			Kind:     "call",
			FromFunc: fromFunc,
			FilePath: filePath,
			Line:     line,
			Context:  getLine(line),
		})
	case "subscript_expression":
		// dispatch_table[i](...) — extract the table name
		obj := fn.ChildByFieldName("argument")
		if obj == nil {
			// Try first named child
			for k := 0; k < int(fn.NamedChildCount()); k++ {
				c := fn.NamedChild(k)
				if c.Type() == "identifier" {
					obj = c
					break
				}
			}
		}
		if obj != nil && obj.Type() == "identifier" {
			*refs = append(*refs, Reference{
				ToName:   obj.Content(source),
				Kind:     "call",
				FromFunc: fromFunc,
				FilePath: filePath,
				Line:     line,
				Context:  getLine(line),
			})
		}
	}
}

// extractAddressOf handles &func_name expressions.
func (p *CParser) extractAddressOf(source []byte, node *sitter.Node, filePath, fromFunc string, getLine func(int) string, refs *[]Reference) {
	// unary_expression with operator "&" and operand identifier
	for j := 0; j < int(node.ChildCount()); j++ {
		child := node.Child(j)
		if child != nil && child.Type() == "identifier" {
			line := int(child.StartPoint().Row) + 1
			*refs = append(*refs, Reference{
				ToName:   child.Content(source),
				Kind:     "address_of",
				FromFunc: fromFunc,
				FilePath: filePath,
				Line:     line,
				Context:  getLine(line),
			})
		}
	}
}

// extractRefsFromInitList recursively walks an initializer_list (and nested
// initializer_lists for struct arrays like `{ { "name", f, 1 }, ... }`)
// extracting identifier references as "initializer" kind.
// tableName is the variable name of the table being initialized (e.g., "dispatch_table").
func (p *CParser) extractRefsFromInitList(source []byte, node *sitter.Node, filePath, fromFunc, tableName string, getLine func(int) string, refs *[]Reference) {
	for k := 0; k < int(node.NamedChildCount()); k++ {
		elem := node.NamedChild(k)
		switch elem.Type() {
		case "identifier":
			line := int(elem.StartPoint().Row) + 1
			*refs = append(*refs, Reference{
				ToName:    elem.Content(source),
				Kind:      "initializer",
				FromFunc:  fromFunc,
				FilePath:  filePath,
				Line:      line,
				Context:   getLine(line),
				TableName: tableName,
			})
		case "initializer_list":
			// Nested struct: { "name", func_ptr, flag }
			p.extractRefsFromInitList(source, elem, filePath, fromFunc, tableName, getLine, refs)
		case "unary_expression":
			p.extractAddressOf(source, elem, filePath, fromFunc, getLine, refs)
		}
	}
}

// getDeclaratorName extracts the variable name from an init_declarator node.
// Handles: `name`, `name[]`, `*name`, `name[SIZE]` etc.
func (p *CParser) getDeclaratorName(source []byte, initDecl *sitter.Node) string {
	decl := initDecl.ChildByFieldName("declarator")
	if decl == nil {
		return ""
	}
	// Walk down to find the innermost identifier
	return p.findIdentifier(source, decl)
}

func (p *CParser) findIdentifier(source []byte, node *sitter.Node) string {
	if node.Type() == "identifier" {
		return node.Content(source)
	}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		if name := p.findIdentifier(source, node.NamedChild(i)); name != "" {
			return name
		}
	}
	return ""
}

// extractRefsFromTopDecl extracts references from top-level declarations
// (dispatch tables, global function pointer assignments).
func (p *CParser) extractRefsFromTopDecl(source []byte, node *sitter.Node, filePath string, getLine func(int) string, refs *[]Reference) {
	for j := 0; j < int(node.NamedChildCount()); j++ {
		child := node.NamedChild(j)
		if child.Type() == "init_declarator" {
			value := child.ChildByFieldName("value")
			if value == nil {
				continue
			}

			// Extract the variable name from the declarator
			varName := p.getDeclaratorName(source, child)

			if value.Type() == "initializer_list" {
				// Dispatch table: static math_op table[] = { f, multiply, subtract };
				// or struct array: { { "name", f, 1 }, ... }
				p.extractRefsFromInitList(source, value, filePath, "<top-level>", varName, getLine, refs)
			} else if value.Type() == "identifier" {
				line := int(value.StartPoint().Row) + 1
				*refs = append(*refs, Reference{
					ToName:   value.Content(source),
					Kind:     "pointer_assign",
					FromFunc: "<top-level>",
					FilePath: filePath,
					Line:     line,
					Context:  getLine(line),
				})
			}
		}
	}
}
