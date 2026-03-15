package parser

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

// JavaScriptParser extracts symbols from JavaScript source files using tree-sitter.
type JavaScriptParser struct{}

// NewJavaScriptParser creates a new JavaScript parser.
func NewJavaScriptParser() *JavaScriptParser { return &JavaScriptParser{} }

func (p *JavaScriptParser) Language() string     { return "javascript" }
func (p *JavaScriptParser) Extensions() []string { return []string{".js", ".jsx", ".mjs", ".cjs"} }

func (p *JavaScriptParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, javascript.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		symbols = append(symbols, p.extractNode(filePath, source, child)...)
	}
	return symbols, nil
}

func (p *JavaScriptParser) extractNode(filePath string, source []byte, node *sitter.Node) []Symbol {
	switch node.Type() {
	case "import_statement":
		return []Symbol{parseJSImport(filePath, "javascript", source, node)}

	case "function_declaration":
		return []Symbol{parseJSFunction(filePath, "javascript", source, node)}

	case "class_declaration":
		return parseJSClass(filePath, "javascript", source, node)

	case "lexical_declaration":
		// Check for arrow function or CommonJS require
		if s, ok := parseJSArrowFunction(filePath, "javascript", source, node); ok {
			return []Symbol{s}
		}
		// Check for require() calls
		if s, ok := p.parseRequire(filePath, source, node); ok {
			return []Symbol{s}
		}

	case "export_statement":
		return p.parseExport(filePath, source, node)

	case "expression_statement":
		// Check for module.exports or describe/it/test
		if node.NamedChildCount() > 0 {
			child := node.NamedChild(0)
			if child.Type() == "assignment_expression" {
				if s, ok := p.parseModuleExports(filePath, source, child); ok {
					return []Symbol{s}
				}
			}
			if child.Type() == "call_expression" {
				return parseJSDescribe(filePath, "javascript", source, child)
			}
		}
	}
	return nil
}

func (p *JavaScriptParser) parseRequire(filePath string, source []byte, node *sitter.Node) (Symbol, bool) {
	if node.NamedChildCount() == 0 {
		return Symbol{}, false
	}
	declarator := node.NamedChild(0)
	if declarator.Type() != "variable_declarator" {
		return Symbol{}, false
	}
	valueNode := declarator.ChildByFieldName("value")
	if valueNode == nil || valueNode.Type() != "call_expression" {
		return Symbol{}, false
	}
	fn := valueNode.ChildByFieldName("function")
	if fn == nil || fn.Content(source) != "require" {
		return Symbol{}, false
	}

	return Symbol{
		Name:      node.Content(source),
		Type:      "import",
		Language:  "javascript",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}, true
}

func (p *JavaScriptParser) parseModuleExports(filePath string, source []byte, node *sitter.Node) (Symbol, bool) {
	left := node.ChildByFieldName("left")
	if left == nil {
		return Symbol{}, false
	}
	leftText := left.Content(source)
	if !strings.HasPrefix(leftText, "module.exports") && !strings.HasPrefix(leftText, "exports.") {
		return Symbol{}, false
	}

	return Symbol{
		Name:      node.Content(source),
		Type:      "export",
		Language:  "javascript",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}, true
}

func (p *JavaScriptParser) parseExport(filePath string, source []byte, node *sitter.Node) []Symbol {
	// Check for exported declarations
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		inner := p.extractNode(filePath, source, child)
		if len(inner) > 0 {
			return inner
		}
	}
	// Bare export like `export { name }`
	return []Symbol{parseJSExportStatement(filePath, "javascript", source, node)}
}
