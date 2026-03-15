package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/ruby"
)

// RubyParser extracts symbols from Ruby source files using tree-sitter.
type RubyParser struct{}

// NewRubyParser creates a new Ruby parser.
func NewRubyParser() *RubyParser { return &RubyParser{} }

func (p *RubyParser) Language() string     { return "ruby" }
func (p *RubyParser) Extensions() []string { return []string{".rb"} }

func (p *RubyParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, ruby.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	p.extractNodes(filePath, source, root, "", &symbols)
	return symbols, nil
}

func (p *RubyParser) extractNodes(filePath string, source []byte, node *sitter.Node, className string, symbols *[]Symbol) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "call":
			if sym, ok := p.parseCall(filePath, source, child, className); ok {
				*symbols = append(*symbols, sym)
			}
		case "class":
			*symbols = append(*symbols, p.parseClass(filePath, source, child)...)
		case "module":
			*symbols = append(*symbols, p.parseModule(filePath, source, child)...)
		case "method":
			*symbols = append(*symbols, p.parseMethod(filePath, source, child, className))
		case "singleton_method":
			*symbols = append(*symbols, p.parseSingletonMethod(filePath, source, child, className))
		case "assignment":
			if sym, ok := p.parseConstant(filePath, source, child); ok {
				*symbols = append(*symbols, sym)
			}
		}
	}
}

func (p *RubyParser) parseCall(filePath string, source []byte, node *sitter.Node, className string) (Symbol, bool) {
	// Check the method name of the call
	fnNode := node.ChildByFieldName("method")
	if fnNode == nil {
		// For bare calls like `require 'json'`, the first child is an identifier
		if node.NamedChildCount() > 0 {
			first := node.NamedChild(0)
			if first.Type() == "identifier" {
				name := first.Content(source)
				return p.handleCallByName(filePath, source, node, name, className)
			}
		}
		return Symbol{}, false
	}
	name := fnNode.Content(source)
	return p.handleCallByName(filePath, source, node, name, className)
}

func (p *RubyParser) handleCallByName(filePath string, source []byte, node *sitter.Node, name string, className string) (Symbol, bool) {
	switch name {
	case "require", "require_relative":
		return Symbol{
			Name:      node.Content(source),
			Type:      "import",
			Language:  "ruby",
			StartLine: int(node.StartPoint().Row) + 1,
			EndLine:   int(node.EndPoint().Row) + 1,
			FilePath:  filePath,
		}, true
	case "include", "extend":
		return Symbol{
			Name:      node.Content(source),
			Type:      "import",
			Language:  "ruby",
			StartLine: int(node.StartPoint().Row) + 1,
			EndLine:   int(node.EndPoint().Row) + 1,
			FilePath:  filePath,
		}, true
	}
	return Symbol{}, false
}

func (p *RubyParser) parseClass(filePath string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	className := nameNode.Content(source)

	sig := "class " + className
	superclass := node.ChildByFieldName("superclass")
	if superclass != nil {
		sig += " " + superclass.Content(source)
	}

	symbols := []Symbol{{
		Name:      className,
		Type:      "class",
		Language:  "ruby",
		Signature: sig,
		Docstring: getRubyDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashRubyBody(source, node),
	}}

	// Extract members from class body
	body := node.ChildByFieldName("body")
	if body != nil {
		p.extractNodes(filePath, source, body, className, &symbols)
	}
	return symbols
}

func (p *RubyParser) parseModule(filePath string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	moduleName := nameNode.Content(source)

	symbols := []Symbol{{
		Name:      moduleName,
		Type:      "type",
		Language:  "ruby",
		Signature: "module " + moduleName,
		Docstring: getRubyDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashRubyBody(source, node),
	}}

	// Extract methods from module body
	body := node.ChildByFieldName("body")
	if body != nil {
		p.extractNodes(filePath, source, body, moduleName, &symbols)
	}
	return symbols
}

func (p *RubyParser) parseMethod(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	params := node.ChildByFieldName("parameters")
	sig := "def " + name
	if params != nil {
		sig += params.Content(source)
	}

	symType := "method"
	if strings.HasPrefix(name, "test_") || strings.HasPrefix(name, "Test") {
		symType = "test"
	}

	fullName := name
	if className != "" {
		fullName = className + "#" + name
	}

	return Symbol{
		Name:      fullName,
		Type:      symType,
		Language:  "ruby",
		Signature: sig,
		Docstring: getRubyDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashRubyBody(source, node),
	}
}

func (p *RubyParser) parseSingletonMethod(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	params := node.ChildByFieldName("parameters")
	sig := "def self." + name
	if params != nil {
		sig += params.Content(source)
	}

	fullName := name
	if className != "" {
		fullName = className + "." + name
	}

	return Symbol{
		Name:      fullName,
		Type:      "function",
		Language:  "ruby",
		Signature: sig,
		Docstring: getRubyDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashRubyBody(source, node),
	}
}

func (p *RubyParser) parseConstant(filePath string, source []byte, node *sitter.Node) (Symbol, bool) {
	// Check if the left side is a constant (ALL_CAPS)
	left := node.ChildByFieldName("left")
	if left == nil || left.Type() != "constant" {
		return Symbol{}, false
	}
	name := left.Content(source)

	return Symbol{
		Name:      name,
		Type:      "type",
		Language:  "ruby",
		Signature: node.Content(source),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}, true
}

// getRubyDocstring extracts # comments preceding a node.
func getRubyDocstring(source []byte, node *sitter.Node) string {
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
		line = strings.TrimPrefix(line, "#")
		line = strings.TrimSpace(line)
		if line != "" {
			comments = append([]string{line}, comments...)
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

func hashRubyBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
