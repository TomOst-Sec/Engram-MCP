package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/rust"
)

// RustParser extracts symbols from Rust source files using tree-sitter.
type RustParser struct{}

// NewRustParser creates a new Rust parser.
func NewRustParser() *RustParser { return &RustParser{} }

func (p *RustParser) Language() string     { return "rust" }
func (p *RustParser) Extensions() []string { return []string{".rs"} }

func (p *RustParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, rust.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	p.extractNodes(filePath, source, root, &symbols, "")
	return symbols, nil
}

func (p *RustParser) extractNodes(filePath string, source []byte, node *sitter.Node, symbols *[]Symbol, implType string) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "function_item":
			*symbols = append(*symbols, p.parseFunction(filePath, source, child, implType))
		case "struct_item":
			*symbols = append(*symbols, p.parseStruct(filePath, source, child))
		case "enum_item":
			*symbols = append(*symbols, p.parseEnum(filePath, source, child))
		case "trait_item":
			*symbols = append(*symbols, p.parseTrait(filePath, source, child))
		case "impl_item":
			p.parseImpl(filePath, source, child, symbols)
		case "use_declaration":
			*symbols = append(*symbols, p.parseUse(filePath, source, child))
		case "macro_definition":
			*symbols = append(*symbols, p.parseMacro(filePath, source, child))
		case "mod_item":
			p.parseMod(filePath, source, child, symbols)
		case "attribute_item":
			// Skip standalone attributes
		}
	}
}

func (p *RustParser) parseFunction(filePath string, source []byte, node *sitter.Node, implType string) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := p.buildSignature(source, node)

	symType := "function"
	if strings.HasPrefix(name, "test_") {
		symType = "test"
	}
	// Check for #[test] attribute
	if p.hasTestAttribute(source, node) {
		symType = "test"
	}

	fullName := name
	if implType != "" {
		fullName = implType + "." + name
		symType = "method"
		if strings.HasPrefix(name, "test_") || p.hasTestAttribute(source, node) {
			symType = "test"
		}
	}

	return Symbol{
		Name:      fullName,
		Type:      symType,
		Language:  "rust",
		Signature: sig,
		Docstring: getRustDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashRustBody(source, node),
	}
}

func (p *RustParser) parseStruct(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "type",
		Language:  "rust",
		Signature: "struct " + name,
		Docstring: getRustDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashRustBody(source, node),
	}
}

func (p *RustParser) parseEnum(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "enum",
		Language:  "rust",
		Signature: "enum " + name,
		Docstring: getRustDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashRustBody(source, node),
	}
}

func (p *RustParser) parseTrait(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "interface",
		Language:  "rust",
		Signature: "trait " + name,
		Docstring: getRustDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashRustBody(source, node),
	}
}

func (p *RustParser) parseImpl(filePath string, source []byte, node *sitter.Node, symbols *[]Symbol) {
	// Get the type being implemented
	typeNode := node.ChildByFieldName("type")
	if typeNode == nil {
		return
	}
	typeName := typeNode.Content(source)

	// Extract methods from the impl body
	body := node.ChildByFieldName("body")
	if body == nil {
		return
	}
	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		if child.Type() == "function_item" {
			*symbols = append(*symbols, p.parseFunction(filePath, source, child, typeName))
		}
	}
}

func (p *RustParser) parseUse(filePath string, source []byte, node *sitter.Node) Symbol {
	content := node.Content(source)
	// Strip the "use " prefix and trailing ";"
	name := strings.TrimPrefix(content, "use ")
	name = strings.TrimSuffix(name, ";")

	return Symbol{
		Name:      name,
		Type:      "import",
		Language:  "rust",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

func (p *RustParser) parseMacro(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "macro",
		Language:  "rust",
		Signature: "macro_rules! " + name,
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashRustBody(source, node),
	}
}

func (p *RustParser) parseMod(filePath string, source []byte, node *sitter.Node, symbols *[]Symbol) {
	// Check if this is a #[cfg(test)] mod
	isTestMod := p.hasTestCfgAttribute(source, node)

	body := node.ChildByFieldName("body")
	if body == nil {
		return
	}

	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		switch child.Type() {
		case "function_item":
			sym := p.parseFunction(filePath, source, child, "")
			if isTestMod || p.hasTestAttribute(source, child) {
				sym.Type = "test"
			}
			*symbols = append(*symbols, sym)
		case "use_declaration":
			*symbols = append(*symbols, p.parseUse(filePath, source, child))
		}
	}
}

func (p *RustParser) buildSignature(source []byte, node *sitter.Node) string {
	content := node.Content(source)
	// Truncate at the body block opening brace
	if idx := strings.Index(content, "{\n"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	if idx := strings.Index(content, " {"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	return content
}

func (p *RustParser) hasTestAttribute(source []byte, node *sitter.Node) bool {
	prev := node.PrevNamedSibling()
	if prev == nil {
		return false
	}
	if prev.Type() == "attribute_item" {
		content := prev.Content(source)
		if strings.Contains(content, "#[test]") {
			return true
		}
	}
	return false
}

func (p *RustParser) hasTestCfgAttribute(source []byte, node *sitter.Node) bool {
	prev := node.PrevNamedSibling()
	if prev == nil {
		return false
	}
	if prev.Type() == "attribute_item" {
		content := prev.Content(source)
		if strings.Contains(content, "cfg(test)") {
			return true
		}
	}
	return false
}

// getRustDocstring extracts /// doc comments preceding a node.
func getRustDocstring(source []byte, node *sitter.Node) string {
	var comments []string
	cur := node.PrevNamedSibling()

	for cur != nil && cur.Type() == "line_comment" {
		content := cur.Content(source)
		if !strings.HasPrefix(content, "///") {
			break
		}
		// Check adjacency
		next := cur.NextNamedSibling()
		if next != nil && int(next.StartPoint().Row)-int(cur.EndPoint().Row) > 1 {
			break
		}
		line := strings.TrimPrefix(content, "///")
		line = strings.TrimPrefix(line, " ")
		comments = append([]string{line}, comments...)

		prev := cur.PrevNamedSibling()
		if prev == nil || prev.Type() != "line_comment" {
			break
		}
		if int(cur.StartPoint().Row)-int(prev.EndPoint().Row) > 1 {
			break
		}
		cur = prev
	}

	// Also check for attribute items between doc comments and node
	if len(comments) == 0 {
		// The node might have attribute(s) between doc comments and it
		cur = node.PrevNamedSibling()
		for cur != nil && cur.Type() == "attribute_item" {
			cur = cur.PrevNamedSibling()
		}
		if cur != nil && cur.Type() == "line_comment" {
			for cur != nil && cur.Type() == "line_comment" {
				content := cur.Content(source)
				if !strings.HasPrefix(content, "///") {
					break
				}
				line := strings.TrimPrefix(content, "///")
				line = strings.TrimPrefix(line, " ")
				comments = append([]string{line}, comments...)

				prev := cur.PrevNamedSibling()
				if prev == nil || prev.Type() != "line_comment" {
					break
				}
				if int(cur.StartPoint().Row)-int(prev.EndPoint().Row) > 1 {
					break
				}
				cur = prev
			}
		}
	}

	if len(comments) == 0 {
		return ""
	}
	return strings.Join(comments, "\n")
}

func hashRustBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
