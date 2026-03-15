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
