package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/kotlin"
)

// KotlinParser extracts symbols from Kotlin source files using tree-sitter.
type KotlinParser struct{}

// NewKotlinParser creates a new Kotlin parser.
func NewKotlinParser() *KotlinParser { return &KotlinParser{} }

func (p *KotlinParser) Language() string     { return "kotlin" }
func (p *KotlinParser) Extensions() []string { return []string{".kt", ".kts"} }

func (p *KotlinParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, kotlin.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	p.extractNodes(filePath, source, root, &symbols, "")
	return symbols, nil
}

func (p *KotlinParser) extractNodes(filePath string, source []byte, node *sitter.Node, symbols *[]Symbol, className string) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "import_header", "import_list":
			p.parseImports(filePath, source, child, symbols)
		case "function_declaration":
			*symbols = append(*symbols, p.parseFunction(filePath, source, child, className))
		case "class_declaration":
			// Kotlin uses class_declaration for class, interface, enum, etc.
			if p.isInterfaceDecl(source, child) {
				*symbols = append(*symbols, p.parseInterface(filePath, source, child))
			} else {
				*symbols = append(*symbols, p.parseClass(filePath, source, child)...)
			}
		case "object_declaration":
			*symbols = append(*symbols, p.parseObject(filePath, source, child)...)
		}
	}
}

func (p *KotlinParser) parseImports(filePath string, source []byte, node *sitter.Node, symbols *[]Symbol) {
	if node.Type() == "import_header" {
		content := node.Content(source)
		name := strings.TrimPrefix(content, "import ")
		name = strings.TrimSpace(name)
		*symbols = append(*symbols, Symbol{
			Name:      name,
			Type:      "import",
			Language:  "kotlin",
			StartLine: int(node.StartPoint().Row) + 1,
			EndLine:   int(node.EndPoint().Row) + 1,
			FilePath:  filePath,
		})
		return
	}
	// import_list contains multiple import_header children
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "import_header" {
			p.parseImports(filePath, source, child, symbols)
		}
	}
}

func (p *KotlinParser) parseFunction(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	nameNode := p.findFunctionName(source, node)
	if nameNode == "" {
		return Symbol{}
	}
	name := nameNode

	sig := p.buildSignature(source, node)

	symType := "function"
	if className != "" {
		symType = "method"
	}
	if p.hasTestAnnotation(source, node) {
		symType = "test"
	}

	// Check for extension function (has receiver type)
	receiverType := p.getReceiverType(source, node)
	fullName := name
	if receiverType != "" {
		fullName = receiverType + "." + name
	} else if className != "" {
		fullName = className + "." + name
	}

	return Symbol{
		Name:      fullName,
		Type:      symType,
		Language:  "kotlin",
		Signature: sig,
		Docstring: getKotlinDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashKotlinBody(source, node),
	}
}

func (p *KotlinParser) findFunctionName(source []byte, node *sitter.Node) string {
	// Try field name first
	nameNode := node.ChildByFieldName("name")
	if nameNode != nil {
		return nameNode.Content(source)
	}
	// Look for simple_identifier child
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "simple_identifier" {
			return child.Content(source)
		}
	}
	return ""
}

func (p *KotlinParser) getReceiverType(source []byte, node *sitter.Node) string {
	// In Kotlin tree-sitter, extension functions have a receiver_type field or
	// the function signature contains "Type.name"
	content := node.Content(source)
	// Look for pattern: fun Type.name(
	if strings.HasPrefix(content, "fun ") {
		afterFun := content[4:]
		if dotIdx := strings.Index(afterFun, "."); dotIdx > 0 {
			parenIdx := strings.Index(afterFun, "(")
			if parenIdx > dotIdx {
				return strings.TrimSpace(afterFun[:dotIdx])
			}
		}
	}
	return ""
}

func (p *KotlinParser) parseClass(filePath string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		// Try simple_identifier
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "type_identifier" || child.Type() == "simple_identifier" {
				nameNode = child
				break
			}
		}
	}
	if nameNode == nil {
		return nil
	}
	className := nameNode.Content(source)

	sig := p.buildTypeSignature(source, node)

	symbols := []Symbol{{
		Name:      className,
		Type:      "class",
		Language:  "kotlin",
		Signature: sig,
		Docstring: getKotlinDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashKotlinBody(source, node),
	}}

	// Extract members from class body
	body := node.ChildByFieldName("body")
	if body == nil {
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

func (p *KotlinParser) parseObject(filePath string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "type_identifier" || child.Type() == "simple_identifier" {
				nameNode = child
				break
			}
		}
	}
	if nameNode == nil {
		return nil
	}
	name := nameNode.Content(source)

	symbols := []Symbol{{
		Name:      name,
		Type:      "type",
		Language:  "kotlin",
		Signature: "object " + name,
		Docstring: getKotlinDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashKotlinBody(source, node),
	}}

	return symbols
}

func (p *KotlinParser) parseInterface(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "type_identifier" || child.Type() == "simple_identifier" {
				nameNode = child
				break
			}
		}
	}
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "interface",
		Language:  "kotlin",
		Signature: "interface " + name,
		Docstring: getKotlinDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashKotlinBody(source, node),
	}
}

func (p *KotlinParser) extractClassMembers(filePath string, source []byte, body *sitter.Node, symbols *[]Symbol, className string) {
	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		switch child.Type() {
		case "function_declaration":
			*symbols = append(*symbols, p.parseFunction(filePath, source, child, className))
		case "class_declaration":
			*symbols = append(*symbols, p.parseClass(filePath, source, child)...)
		case "object_declaration":
			*symbols = append(*symbols, p.parseObject(filePath, source, child)...)
		case "companion_object":
			p.parseCompanionObject(filePath, source, child, symbols, className)
		}
	}
}

func (p *KotlinParser) parseCompanionObject(filePath string, source []byte, node *sitter.Node, symbols *[]Symbol, className string) {
	body := node.ChildByFieldName("body")
	if body == nil {
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "class_body" {
				body = child
				break
			}
		}
	}
	if body != nil {
		p.extractClassMembers(filePath, source, body, symbols, className)
	}
}

func (p *KotlinParser) isInterfaceDecl(source []byte, node *sitter.Node) bool {
	content := node.Content(source)
	words := strings.Fields(content)
	for _, w := range words {
		if w == "interface" {
			return true
		}
		if w == "class" || w == "{" {
			return false
		}
	}
	return false
}

func (p *KotlinParser) hasTestAnnotation(source []byte, node *sitter.Node) bool {
	// Check for @Test annotation — look at modifiers or preceding siblings
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "modifiers" {
			for j := 0; j < int(child.NamedChildCount()); j++ {
				mod := child.NamedChild(j)
				if mod.Type() == "annotation" {
					content := mod.Content(source)
					if content == "@Test" {
						return true
					}
				}
			}
		}
	}
	// Also check preceding sibling
	prev := node.PrevNamedSibling()
	if prev != nil && prev.Type() == "annotation" {
		content := prev.Content(source)
		if content == "@Test" {
			return true
		}
	}
	return false
}

func (p *KotlinParser) buildSignature(source []byte, node *sitter.Node) string {
	content := node.Content(source)
	if idx := strings.Index(content, "{\n"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	if idx := strings.Index(content, " {"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	// Single-expression function: fun name() = expr
	if idx := strings.Index(content, " ="); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	return content
}

func (p *KotlinParser) buildTypeSignature(source []byte, node *sitter.Node) string {
	content := node.Content(source)
	if idx := strings.Index(content, "{"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	return content
}

// getKotlinDocstring extracts KDoc (/** ... */) comments preceding a node.
func getKotlinDocstring(source []byte, node *sitter.Node) string {
	prev := node.PrevNamedSibling()
	if prev == nil {
		return ""
	}

	if prev.Type() != "multiline_comment" && prev.Type() != "comment" {
		return ""
	}

	if int(node.StartPoint().Row)-int(prev.EndPoint().Row) > 1 {
		return ""
	}

	raw := prev.Content(source)
	if !strings.HasPrefix(raw, "/**") {
		return ""
	}

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

func hashKotlinBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
