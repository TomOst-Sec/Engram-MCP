package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"
)

// PHPParser extracts symbols from PHP source files using tree-sitter.
type PHPParser struct{}

// NewPHPParser creates a new PHP parser.
func NewPHPParser() *PHPParser { return &PHPParser{} }

func (p *PHPParser) Language() string     { return "php" }
func (p *PHPParser) Extensions() []string { return []string{".php"} }

func (p *PHPParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, php.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	p.extractNodes(filePath, source, root, &symbols)
	return symbols, nil
}

func (p *PHPParser) extractNodes(filePath string, source []byte, node *sitter.Node, symbols *[]Symbol) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "namespace_use_declaration":
			*symbols = append(*symbols, p.parseUseDecl(filePath, source, child))
		case "class_declaration":
			*symbols = append(*symbols, p.parseClass(filePath, source, child)...)
		case "interface_declaration":
			*symbols = append(*symbols, p.parseInterface(filePath, source, child)...)
		case "trait_declaration":
			*symbols = append(*symbols, p.parseTrait(filePath, source, child)...)
		case "function_definition":
			*symbols = append(*symbols, p.parseFunction(filePath, source, child))
		case "enum_declaration":
			*symbols = append(*symbols, p.parseEnum(filePath, source, child))
		}
	}
}

func (p *PHPParser) parseUseDecl(filePath string, source []byte, node *sitter.Node) Symbol {
	return Symbol{
		Name:      strings.TrimSuffix(node.Content(source), ";"),
		Type:      "import",
		Language:  "php",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

func (p *PHPParser) parseClass(filePath string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	className := nameNode.Content(source)

	sig := p.buildPHPClassSignature(source, node, "class")

	symbols := []Symbol{{
		Name:      className,
		Type:      "class",
		Language:  "php",
		Signature: sig,
		Docstring: getPHPDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashPHPBody(source, node),
	}}

	body := node.ChildByFieldName("body")
	if body != nil {
		p.extractClassMembers(filePath, source, body, className, &symbols)
	}
	return symbols
}

func (p *PHPParser) parseInterface(filePath string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nameNode.Content(source)

	symbols := []Symbol{{
		Name:      name,
		Type:      "interface",
		Language:  "php",
		Signature: "interface " + name,
		Docstring: getPHPDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashPHPBody(source, node),
	}}

	body := node.ChildByFieldName("body")
	if body != nil {
		p.extractClassMembers(filePath, source, body, name, &symbols)
	}
	return symbols
}

func (p *PHPParser) parseTrait(filePath string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nameNode.Content(source)

	symbols := []Symbol{{
		Name:      name,
		Type:      "trait",
		Language:  "php",
		Signature: "trait " + name,
		Docstring: getPHPDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashPHPBody(source, node),
	}}

	body := node.ChildByFieldName("body")
	if body != nil {
		p.extractClassMembers(filePath, source, body, name, &symbols)
	}
	return symbols
}

func (p *PHPParser) parseFunction(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := p.buildPHPFuncSignature(source, node, "")

	return Symbol{
		Name:      name,
		Type:      "function",
		Language:  "php",
		Signature: sig,
		Docstring: getPHPDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashPHPBody(source, node),
	}
}

func (p *PHPParser) parseEnum(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "enum",
		Language:  "php",
		Signature: "enum " + name,
		Docstring: getPHPDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashPHPBody(source, node),
	}
}

func (p *PHPParser) extractClassMembers(filePath string, source []byte, body *sitter.Node, className string, symbols *[]Symbol) {
	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		if child.Type() == "method_declaration" {
			*symbols = append(*symbols, p.parseMethod(filePath, source, child, className))
		}
	}
}

func (p *PHPParser) parseMethod(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := p.buildPHPFuncSignature(source, node, "")

	symType := "method"
	if name == "__construct" {
		symType = "constructor"
	} else if strings.HasPrefix(name, "test") || strings.HasPrefix(name, "Test") {
		symType = "test"
	}

	return Symbol{
		Name:      className + "::" + name,
		Type:      symType,
		Language:  "php",
		Signature: sig,
		Docstring: getPHPDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashPHPBody(source, node),
	}
}

func (p *PHPParser) buildPHPClassSignature(source []byte, node *sitter.Node, keyword string) string {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return keyword
	}
	sig := keyword + " " + nameNode.Content(source)

	// Check for extends
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Type() {
		case "base_clause":
			sig += " " + child.Content(source)
		case "class_interface_clause":
			sig += " " + child.Content(source)
		}
	}
	return sig
}

func (p *PHPParser) buildPHPFuncSignature(source []byte, node *sitter.Node, prefix string) string {
	var parts []string

	// Get visibility modifier
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == "visibility_modifier" {
			parts = append(parts, child.Content(source))
			break
		}
	}

	parts = append(parts, "function")

	nameNode := node.ChildByFieldName("name")
	if nameNode != nil {
		parts = append(parts, nameNode.Content(source))
	}

	sig := strings.Join(parts, " ")

	params := node.ChildByFieldName("parameters")
	if params != nil {
		sig += params.Content(source)
	}

	// Get return type
	retType := node.ChildByFieldName("return_type")
	if retType != nil {
		sig += ": " + retType.Content(source)
	} else {
		// Try to find return type manually (colon + type after params)
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child == nil {
				continue
			}
			switch child.Type() {
			case "primitive_type", "named_type", "nullable_type", "union_type":
				sig += ": " + child.Content(source)
			}
		}
	}

	return sig
}

// getPHPDocstring extracts PHPDoc (/** ... */) or // comments preceding a node.
func getPHPDocstring(source []byte, node *sitter.Node) string {
	prev := node.PrevNamedSibling()
	if prev == nil || prev.Type() != "comment" {
		return ""
	}
	if int(node.StartPoint().Row)-int(prev.EndPoint().Row) > 1 {
		return ""
	}

	raw := prev.Content(source)

	// Handle PHPDoc style
	if strings.HasPrefix(raw, "/**") {
		raw = strings.TrimPrefix(raw, "/**")
		raw = strings.TrimSuffix(raw, "*/")
		lines := strings.Split(raw, "\n")
		var cleaned []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			line = strings.TrimPrefix(line, "* ")
			line = strings.TrimPrefix(line, "*")
			if line != "" {
				cleaned = append(cleaned, line)
			}
		}
		return strings.Join(cleaned, "\n")
	}

	// Handle // style
	raw = strings.TrimPrefix(raw, "//")
	return strings.TrimSpace(raw)
}

func hashPHPBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
