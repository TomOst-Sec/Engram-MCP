package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

// PythonParser extracts symbols from Python source files using tree-sitter.
type PythonParser struct{}

// NewPythonParser creates a new Python parser.
func NewPythonParser() *PythonParser { return &PythonParser{} }

func (p *PythonParser) Language() string     { return "python" }
func (p *PythonParser) Extensions() []string { return []string{".py", ".pyi"} }

func (p *PythonParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, python.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		switch child.Type() {
		case "function_definition":
			symbols = append(symbols, p.parseFunction(filePath, source, child))
		case "class_definition":
			symbols = append(symbols, p.parseClass(filePath, source, child)...)
		case "import_statement":
			symbols = append(symbols, p.parseImport(filePath, source, child))
		case "import_from_statement":
			symbols = append(symbols, p.parseFromImport(filePath, source, child))
		case "decorated_definition":
			// Handle decorated functions/classes
			for j := 0; j < int(child.NamedChildCount()); j++ {
				inner := child.NamedChild(j)
				switch inner.Type() {
				case "function_definition":
					symbols = append(symbols, p.parseFunction(filePath, source, inner))
				case "class_definition":
					symbols = append(symbols, p.parseClass(filePath, source, inner)...)
				}
			}
		}
	}
	return symbols, nil
}

func (p *PythonParser) parseFunction(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	name := nameNode.Content(source)

	sig := p.buildPythonSignature(source, node)

	symType := "function"
	if strings.HasPrefix(name, "test_") || strings.HasPrefix(name, "Test") {
		symType = "test"
	}

	return Symbol{
		Name:      name,
		Type:      symType,
		Language:  "python",
		Signature: sig,
		Docstring: getPythonDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashPythonBody(source, node),
	}
}

func (p *PythonParser) parseClass(filePath string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	className := nameNode.Content(source)

	// Get base classes
	var bases string
	superclass := node.ChildByFieldName("superclasses")
	if superclass != nil {
		bases = superclass.Content(source)
	}

	sig := "class " + className
	if bases != "" {
		sig += bases
	}

	symbols := []Symbol{{
		Name:      className,
		Type:      "class",
		Language:  "python",
		Signature: sig,
		Docstring: getPythonDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashPythonBody(source, node),
	}}

	// Extract methods from class body
	body := node.ChildByFieldName("body")
	if body == nil {
		return symbols
	}
	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		if child.Type() == "function_definition" {
			m := p.parseMethod(filePath, source, child, className)
			symbols = append(symbols, m)
		} else if child.Type() == "decorated_definition" {
			for j := 0; j < int(child.NamedChildCount()); j++ {
				inner := child.NamedChild(j)
				if inner.Type() == "function_definition" {
					m := p.parseMethod(filePath, source, inner, className)
					symbols = append(symbols, m)
				}
			}
		}
	}
	return symbols
}

func (p *PythonParser) parseMethod(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	nameNode := node.ChildByFieldName("name")
	name := nameNode.Content(source)

	sig := p.buildPythonSignature(source, node)

	return Symbol{
		Name:      className + "." + name,
		Type:      "method",
		Language:  "python",
		Signature: sig,
		Docstring: getPythonDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashPythonBody(source, node),
	}
}

func (p *PythonParser) parseImport(filePath string, source []byte, node *sitter.Node) Symbol {
	return Symbol{
		Name:      node.Content(source),
		Type:      "import",
		Language:  "python",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

func (p *PythonParser) parseFromImport(filePath string, source []byte, node *sitter.Node) Symbol {
	return Symbol{
		Name:      node.Content(source),
		Type:      "import",
		Language:  "python",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

func (p *PythonParser) buildPythonSignature(source []byte, node *sitter.Node) string {
	nameNode := node.ChildByFieldName("name")
	params := node.ChildByFieldName("parameters")
	retType := node.ChildByFieldName("return_type")

	sig := "def " + nameNode.Content(source)
	if params != nil {
		sig += params.Content(source)
	}
	if retType != nil {
		sig += " -> " + retType.Content(source)
	}
	return sig
}

// getPythonDocstring extracts the docstring from a function/class body.
func getPythonDocstring(source []byte, node *sitter.Node) string {
	body := node.ChildByFieldName("body")
	if body == nil || body.NamedChildCount() == 0 {
		return ""
	}
	first := body.NamedChild(0)
	if first.Type() != "expression_statement" {
		return ""
	}
	if first.NamedChildCount() == 0 {
		return ""
	}
	expr := first.NamedChild(0)
	if expr.Type() != "string" {
		return ""
	}
	raw := expr.Content(source)
	// Strip triple quotes
	raw = strings.TrimPrefix(raw, `"""`)
	raw = strings.TrimSuffix(raw, `"""`)
	raw = strings.TrimPrefix(raw, `'''`)
	raw = strings.TrimSuffix(raw, `'''`)
	return strings.TrimSpace(raw)
}

func hashPythonBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
