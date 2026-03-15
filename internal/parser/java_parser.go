package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

// JavaParser extracts symbols from Java source files using tree-sitter.
type JavaParser struct{}

// NewJavaParser creates a new Java parser.
func NewJavaParser() *JavaParser { return &JavaParser{} }

func (p *JavaParser) Language() string     { return "java" }
func (p *JavaParser) Extensions() []string { return []string{".java"} }

func (p *JavaParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, java.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	p.extractNodes(filePath, source, root, &symbols, "")
	return symbols, nil
}

func (p *JavaParser) extractNodes(filePath string, source []byte, node *sitter.Node, symbols *[]Symbol, className string) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "import_declaration":
			*symbols = append(*symbols, p.parseImport(filePath, source, child))
		case "class_declaration":
			*symbols = append(*symbols, p.parseClass(filePath, source, child, className)...)
		case "interface_declaration":
			*symbols = append(*symbols, p.parseInterface(filePath, source, child))
		case "enum_declaration":
			*symbols = append(*symbols, p.parseEnumDecl(filePath, source, child))
		case "method_declaration":
			*symbols = append(*symbols, p.parseMethod(filePath, source, child, className))
		case "constructor_declaration":
			*symbols = append(*symbols, p.parseConstructor(filePath, source, child, className))
		}
	}
}

func (p *JavaParser) parseImport(filePath string, source []byte, node *sitter.Node) Symbol {
	content := node.Content(source)
	// Strip "import " prefix and trailing ";"
	name := strings.TrimPrefix(content, "import ")
	name = strings.TrimSuffix(name, ";")

	return Symbol{
		Name:      name,
		Type:      "import",
		Language:  "java",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

func (p *JavaParser) parseClass(filePath string, source []byte, node *sitter.Node, _ string) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	className := nameNode.Content(source)

	sig := p.buildClassSignature(source, node)

	symbols := []Symbol{{
		Name:      className,
		Type:      "class",
		Language:  "java",
		Signature: sig,
		Docstring: getJavadoc(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJavaBody(source, node),
	}}

	// Extract members from class body
	body := node.ChildByFieldName("body")
	if body == nil {
		return symbols
	}
	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		switch child.Type() {
		case "method_declaration":
			symbols = append(symbols, p.parseMethod(filePath, source, child, className))
		case "constructor_declaration":
			symbols = append(symbols, p.parseConstructor(filePath, source, child, className))
		case "class_declaration":
			symbols = append(symbols, p.parseClass(filePath, source, child, className)...)
		}
	}

	return symbols
}

func (p *JavaParser) parseInterface(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := "interface " + name
	// Check for extends
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == "extends_interfaces" {
			sig += " " + child.Content(source)
		}
	}

	return Symbol{
		Name:      name,
		Type:      "interface",
		Language:  "java",
		Signature: sig,
		Docstring: getJavadoc(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJavaBody(source, node),
	}
}

func (p *JavaParser) parseEnumDecl(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "enum",
		Language:  "java",
		Signature: "enum " + name,
		Docstring: getJavadoc(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJavaBody(source, node),
	}
}

func (p *JavaParser) parseMethod(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := p.buildMethodSignature(source, node)

	symType := "method"
	if p.hasTestAnnotation(source, node) {
		symType = "test"
	}

	fullName := name
	if className != "" {
		fullName = className + "." + name
	}

	return Symbol{
		Name:      fullName,
		Type:      symType,
		Language:  "java",
		Signature: sig,
		Docstring: getJavadoc(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJavaBody(source, node),
	}
}

func (p *JavaParser) parseConstructor(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := p.buildConstructorSignature(source, node)

	fullName := name
	if className != "" {
		fullName = className + "." + name
	}

	return Symbol{
		Name:      fullName,
		Type:      "constructor",
		Language:  "java",
		Signature: sig,
		Docstring: getJavadoc(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJavaBody(source, node),
	}
}

func (p *JavaParser) buildClassSignature(source []byte, node *sitter.Node) string {
	content := node.Content(source)
	// Find the opening brace and truncate
	if idx := strings.Index(content, "{"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	return content
}

func (p *JavaParser) buildMethodSignature(source []byte, node *sitter.Node) string {
	content := node.Content(source)
	if idx := strings.Index(content, "{"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	// For interface methods ending with ;
	if idx := strings.Index(content, ";"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	return content
}

func (p *JavaParser) buildConstructorSignature(source []byte, node *sitter.Node) string {
	content := node.Content(source)
	if idx := strings.Index(content, "{"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	return content
}

func (p *JavaParser) hasTestAnnotation(source []byte, node *sitter.Node) bool {
	// Check for marker_annotation siblings (Java annotations appear as children of the parent)
	// In tree-sitter-java, annotations are often part of the modifiers
	parent := node.Parent()
	if parent == nil {
		return false
	}

	// Find the method's position among siblings and check preceding annotations
	for i := 0; i < int(parent.NamedChildCount()); i++ {
		child := parent.NamedChild(i)
		if child == node {
			// Check preceding siblings for annotations
			if i > 0 {
				prev := parent.NamedChild(i - 1)
				if isTestAnnotationNode(source, prev) {
					return true
				}
			}
			break
		}
	}

	// Also check the node's own children for modifiers containing annotations
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == "modifiers" {
			for j := 0; j < int(child.NamedChildCount()); j++ {
				mod := child.NamedChild(j)
				if isTestAnnotationNode(source, mod) {
					return true
				}
			}
		}
	}

	return false
}

func isTestAnnotationNode(source []byte, node *sitter.Node) bool {
	if node == nil {
		return false
	}
	if node.Type() == "marker_annotation" || node.Type() == "annotation" {
		content := node.Content(source)
		return content == "@Test" || content == "@TestMethod" || content == "@Fact" || content == "@Theory"
	}
	return false
}

// getJavadoc extracts Javadoc comments (/** ... */) preceding a node.
func getJavadoc(source []byte, node *sitter.Node) string {
	prev := node.PrevNamedSibling()
	if prev == nil {
		return ""
	}

	// In tree-sitter-java, Javadoc is a block_comment or a comment node
	var commentNode *sitter.Node
	if prev.Type() == "block_comment" || prev.Type() == "comment" {
		commentNode = prev
	} else if prev.Type() == "line_comment" {
		return ""
	}

	if commentNode == nil {
		return ""
	}

	// Verify adjacency
	if int(node.StartPoint().Row)-int(commentNode.EndPoint().Row) > 1 {
		return ""
	}

	raw := commentNode.Content(source)
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

func hashJavaBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
