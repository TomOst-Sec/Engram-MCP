package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/swift"
)

// SwiftParser extracts symbols from Swift source files using tree-sitter.
type SwiftParser struct{}

// NewSwiftParser creates a new Swift parser.
func NewSwiftParser() *SwiftParser { return &SwiftParser{} }

func (p *SwiftParser) Language() string     { return "swift" }
func (p *SwiftParser) Extensions() []string { return []string{".swift"} }

func (p *SwiftParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, swift.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	p.extractNodes(filePath, source, root, &symbols, "")
	return symbols, nil
}

func (p *SwiftParser) extractNodes(filePath string, source []byte, node *sitter.Node, symbols *[]Symbol, className string) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "import_declaration":
			*symbols = append(*symbols, p.parseImport(filePath, source, child))
		case "function_declaration":
			*symbols = append(*symbols, p.parseFunction(filePath, source, child, className))
		case "class_declaration":
			// Swift tree-sitter uses class_declaration for class, struct, enum, and extension
			content := child.Content(source)
			kind := p.detectSwiftDeclKind(content)
			switch kind {
			case "struct":
				*symbols = append(*symbols, p.parseStruct(filePath, source, child))
			case "enum":
				*symbols = append(*symbols, p.parseEnum(filePath, source, child))
			case "extension":
				p.parseExtension(filePath, source, child, symbols)
			default:
				*symbols = append(*symbols, p.parseClass(filePath, source, child)...)
			}
		case "protocol_declaration":
			*symbols = append(*symbols, p.parseProtocol(filePath, source, child))
		}
	}
}

func (p *SwiftParser) detectSwiftDeclKind(content string) string {
	trimmed := strings.TrimSpace(content)
	// Check for keyword at start (after optional access modifiers)
	words := strings.Fields(trimmed)
	for _, w := range words {
		switch w {
		case "struct":
			return "struct"
		case "enum":
			return "enum"
		case "extension":
			return "extension"
		case "class":
			return "class"
		}
	}
	return "class"
}

func (p *SwiftParser) parseImport(filePath string, source []byte, node *sitter.Node) Symbol {
	content := node.Content(source)
	name := strings.TrimPrefix(content, "import ")

	return Symbol{
		Name:      name,
		Type:      "import",
		Language:  "swift",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

func (p *SwiftParser) parseFunction(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := p.buildSignature(source, node)

	symType := "function"
	if className != "" {
		symType = "method"
	}
	if strings.HasPrefix(name, "test") {
		symType = "test"
	}

	fullName := name
	if className != "" {
		fullName = className + "." + name
	}

	return Symbol{
		Name:      fullName,
		Type:      symType,
		Language:  "swift",
		Signature: sig,
		Docstring: getSwiftDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashSwiftBody(source, node),
	}
}

func (p *SwiftParser) parseClass(filePath string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	className := nameNode.Content(source)

	sig := p.buildTypeSignature(source, node, "class")

	symbols := []Symbol{{
		Name:      className,
		Type:      "class",
		Language:  "swift",
		Signature: sig,
		Docstring: getSwiftDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashSwiftBody(source, node),
	}}

	// Extract members from class body
	body := node.ChildByFieldName("body")
	if body == nil {
		// Try finding class_body node
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "class_body" {
				body = child
				break
			}
		}
	}
	if body != nil {
		p.extractClassMembers(filePath, source, body, &symbols, className)
	}

	return symbols
}

func (p *SwiftParser) parseStruct(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "type",
		Language:  "swift",
		Signature: p.buildTypeSignature(source, node, "struct"),
		Docstring: getSwiftDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashSwiftBody(source, node),
	}
}

func (p *SwiftParser) parseProtocol(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "interface",
		Language:  "swift",
		Signature: "protocol " + name,
		Docstring: getSwiftDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashSwiftBody(source, node),
	}
}

func (p *SwiftParser) parseEnum(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "enum",
		Language:  "swift",
		Signature: "enum " + name,
		Docstring: getSwiftDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashSwiftBody(source, node),
	}
}

func (p *SwiftParser) parseExtension(filePath string, source []byte, node *sitter.Node, symbols *[]Symbol) {
	// Get the type being extended — may be name field, type_identifier, or user_type
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "user_type" || child.Type() == "type_identifier" {
				nameNode = child
				break
			}
		}
	}
	if nameNode == nil {
		return
	}
	typeName := nameNode.Content(source)

	// Extract methods from the extension body
	body := node.ChildByFieldName("body")
	if body == nil {
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "class_body" || child.Type() == "extension_body" {
				body = child
				break
			}
		}
	}
	if body != nil {
		p.extractClassMembers(filePath, source, body, symbols, typeName)
	}
}

func (p *SwiftParser) extractClassMembers(filePath string, source []byte, body *sitter.Node, symbols *[]Symbol, className string) {
	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		switch child.Type() {
		case "function_declaration":
			*symbols = append(*symbols, p.parseFunction(filePath, source, child, className))
		case "init_declaration":
			*symbols = append(*symbols, p.parseInit(filePath, source, child, className))
		case "deinit_declaration":
			// skip deinit for now
		case "class_declaration":
			*symbols = append(*symbols, p.parseClass(filePath, source, child)...)
		}
	}
}

func (p *SwiftParser) parseInit(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	sig := p.buildSignature(source, node)

	fullName := "init"
	if className != "" {
		fullName = className + ".init"
	}

	return Symbol{
		Name:      fullName,
		Type:      "constructor",
		Language:  "swift",
		Signature: sig,
		Docstring: getSwiftDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashSwiftBody(source, node),
	}
}

func (p *SwiftParser) buildSignature(source []byte, node *sitter.Node) string {
	content := node.Content(source)
	if idx := strings.Index(content, "{\n"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	if idx := strings.Index(content, " {"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	return content
}

func (p *SwiftParser) buildTypeSignature(source []byte, node *sitter.Node, keyword string) string {
	content := node.Content(source)
	if idx := strings.Index(content, "{"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	return keyword + " " + content
}

// getSwiftDocstring extracts /// or /** */ comments preceding a node.
func getSwiftDocstring(source []byte, node *sitter.Node) string {
	var comments []string
	cur := node.PrevNamedSibling()

	for cur != nil && cur.Type() == "comment" {
		content := cur.Content(source)
		if strings.HasPrefix(content, "///") {
			line := strings.TrimPrefix(content, "///")
			line = strings.TrimPrefix(line, " ")
			comments = append([]string{line}, comments...)
		} else if strings.HasPrefix(content, "/**") {
			// Block doc comment
			raw := strings.TrimPrefix(content, "/**")
			raw = strings.TrimSuffix(raw, "*/")
			lines := strings.Split(raw, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				line = strings.TrimPrefix(line, "* ")
				line = strings.TrimPrefix(line, "*")
				if line != "" {
					comments = append(comments, line)
				}
			}
			break
		} else {
			break
		}

		prev := cur.PrevNamedSibling()
		if prev == nil || prev.Type() != "comment" {
			break
		}
		if int(cur.StartPoint().Row)-int(prev.EndPoint().Row) > 1 {
			break
		}
		cur = prev
	}

	if len(comments) == 0 {
		return ""
	}
	return strings.Join(comments, "\n")
}

func hashSwiftBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
