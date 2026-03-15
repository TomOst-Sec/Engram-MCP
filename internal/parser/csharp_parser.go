package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/csharp"
)

// CSharpParser extracts symbols from C# source files using tree-sitter.
type CSharpParser struct{}

// NewCSharpParser creates a new C# parser.
func NewCSharpParser() *CSharpParser { return &CSharpParser{} }

func (p *CSharpParser) Language() string     { return "csharp" }
func (p *CSharpParser) Extensions() []string { return []string{".cs"} }

func (p *CSharpParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, csharp.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	p.extractNodes(filePath, source, root, "", &symbols)
	return symbols, nil
}

func (p *CSharpParser) extractNodes(filePath string, source []byte, node *sitter.Node, className string, symbols *[]Symbol) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "using_directive":
			*symbols = append(*symbols, p.parseUsing(filePath, source, child))

		case "namespace_declaration", "file_scoped_namespace_declaration":
			// Recurse into namespace body
			body := child.ChildByFieldName("body")
			if body != nil {
				p.extractNodes(filePath, source, body, className, symbols)
			}
			// Also check direct children for file-scoped namespaces
			p.extractNodes(filePath, source, child, className, symbols)

		case "class_declaration":
			*symbols = append(*symbols, p.parseClass(filePath, source, child, className)...)

		case "interface_declaration":
			*symbols = append(*symbols, p.parseInterface(filePath, source, child))

		case "struct_declaration":
			*symbols = append(*symbols, p.parseStruct(filePath, source, child)...)

		case "enum_declaration":
			*symbols = append(*symbols, p.parseEnum(filePath, source, child))

		case "record_declaration":
			*symbols = append(*symbols, p.parseRecord(filePath, source, child))
		}
	}
}

func (p *CSharpParser) parseUsing(filePath string, source []byte, node *sitter.Node) Symbol {
	return Symbol{
		Name:      strings.TrimSuffix(node.Content(source), ";"),
		Type:      "import",
		Language:  "csharp",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

func (p *CSharpParser) parseClass(filePath string, source []byte, node *sitter.Node, parentClass string) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	className := nameNode.Content(source)

	sig := p.buildClassSignature(source, node, "class")

	symbols := []Symbol{{
		Name:      className,
		Type:      "class",
		Language:  "csharp",
		Signature: sig,
		Docstring: getCSharpDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCSharpBody(source, node),
	}}

	// Extract members from class body
	body := node.ChildByFieldName("body")
	if body == nil {
		return symbols
	}

	p.extractClassMembers(filePath, source, body, className, &symbols)
	return symbols
}

func (p *CSharpParser) parseInterface(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := p.buildClassSignature(source, node, "interface")

	return Symbol{
		Name:      name,
		Type:      "interface",
		Language:  "csharp",
		Signature: sig,
		Docstring: getCSharpDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCSharpBody(source, node),
	}
}

func (p *CSharpParser) parseStruct(filePath string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nameNode.Content(source)

	sig := p.buildClassSignature(source, node, "struct")

	symbols := []Symbol{{
		Name:      name,
		Type:      "struct",
		Language:  "csharp",
		Signature: sig,
		Docstring: getCSharpDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCSharpBody(source, node),
	}}

	body := node.ChildByFieldName("body")
	if body == nil {
		return symbols
	}

	p.extractClassMembers(filePath, source, body, name, &symbols)
	return symbols
}

func (p *CSharpParser) parseEnum(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "enum",
		Language:  "csharp",
		Signature: "enum " + name,
		Docstring: getCSharpDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCSharpBody(source, node),
	}
}

func (p *CSharpParser) parseRecord(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	// Build record signature including parameters
	sig := "record " + name
	params := node.ChildByFieldName("parameters")
	if params != nil {
		sig += params.Content(source)
	}

	return Symbol{
		Name:      name,
		Type:      "record",
		Language:  "csharp",
		Signature: sig,
		Docstring: getCSharpDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCSharpBody(source, node),
	}
}

func (p *CSharpParser) extractClassMembers(filePath string, source []byte, body *sitter.Node, className string, symbols *[]Symbol) {
	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		switch child.Type() {
		case "method_declaration":
			*symbols = append(*symbols, p.parseMethod(filePath, source, child, className))
		case "constructor_declaration":
			*symbols = append(*symbols, p.parseConstructor(filePath, source, child, className))
		case "property_declaration":
			*symbols = append(*symbols, p.parseProperty(filePath, source, child, className))
		case "class_declaration":
			// Nested class
			*symbols = append(*symbols, p.parseClass(filePath, source, child, className)...)
		}
	}
}

func (p *CSharpParser) parseMethod(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := p.buildMethodSignature(source, node)

	symType := "method"
	if hasTestAttribute(source, node) {
		symType = "test"
	}

	return Symbol{
		Name:      className + "." + name,
		Type:      symType,
		Language:  "csharp",
		Signature: sig,
		Docstring: getCSharpDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCSharpBody(source, node),
	}
}

func (p *CSharpParser) parseConstructor(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	sig := p.buildConstructorSignature(source, node)

	return Symbol{
		Name:      className + "." + className,
		Type:      "constructor",
		Language:  "csharp",
		Signature: sig,
		Docstring: getCSharpDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCSharpBody(source, node),
	}
}

func (p *CSharpParser) parseProperty(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := p.buildPropertySignature(source, node)

	return Symbol{
		Name:      className + "." + name,
		Type:      "property",
		Language:  "csharp",
		Signature: sig,
		Docstring: getCSharpDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCSharpBody(source, node),
	}
}

func (p *CSharpParser) buildClassSignature(source []byte, node *sitter.Node, keyword string) string {
	mods := p.getModifiers(source, node)
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return keyword
	}

	sig := ""
	if mods != "" {
		sig = mods + " "
	}
	sig += keyword + " " + nameNode.Content(source)

	// Check for base types (base_list is a child type, not a field)
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == "base_list" {
			sig += " " + child.Content(source)
			break
		}
	}

	return sig
}

func (p *CSharpParser) buildMethodSignature(source []byte, node *sitter.Node) string {
	mods := p.getModifiers(source, node)
	nameNode := node.ChildByFieldName("name")
	params := node.ChildByFieldName("parameters")

	// Try both "type" and "returns" field names for the return type
	retType := node.ChildByFieldName("type")
	if retType == nil {
		retType = node.ChildByFieldName("returns")
	}

	var parts []string
	if mods != "" {
		parts = append(parts, mods)
	}
	if retType != nil {
		parts = append(parts, retType.Content(source))
	}
	if nameNode != nil {
		name := nameNode.Content(source)
		// Include type parameters if present
		typeParams := node.ChildByFieldName("type_parameters")
		if typeParams != nil {
			name += typeParams.Content(source)
		}
		parts = append(parts, name)
	}

	sig := strings.Join(parts, " ")
	if params != nil {
		sig += params.Content(source)
	}

	// Include constraints
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "type_parameter_constraints_clause" {
			sig += " " + child.Content(source)
		}
	}

	return sig
}

func (p *CSharpParser) buildConstructorSignature(source []byte, node *sitter.Node) string {
	mods := p.getModifiers(source, node)
	nameNode := node.ChildByFieldName("name")
	params := node.ChildByFieldName("parameters")

	sig := ""
	if mods != "" {
		sig = mods + " "
	}
	if nameNode != nil {
		sig += nameNode.Content(source)
	}
	if params != nil {
		sig += params.Content(source)
	}
	return sig
}

func (p *CSharpParser) buildPropertySignature(source []byte, node *sitter.Node) string {
	mods := p.getModifiers(source, node)
	typeNode := node.ChildByFieldName("type")
	nameNode := node.ChildByFieldName("name")

	var parts []string
	if mods != "" {
		parts = append(parts, mods)
	}
	if typeNode != nil {
		parts = append(parts, typeNode.Content(source))
	}
	if nameNode != nil {
		parts = append(parts, nameNode.Content(source))
	}

	return strings.Join(parts, " ")
}

func (p *CSharpParser) getModifiers(source []byte, node *sitter.Node) string {
	var mods []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		if child.Type() == "modifier" {
			mods = append(mods, child.Content(source))
		}
	}
	return strings.Join(mods, " ")
}

// hasTestAttribute checks if a method has [Test], [Fact], [TestMethod], or [Theory] attributes.
func hasTestAttribute(source []byte, node *sitter.Node) bool {
	// Check preceding sibling for attribute_list
	prev := node.PrevNamedSibling()
	if prev != nil && prev.Type() == "attribute_list" {
		if isTestAttrList(source, prev) {
			return true
		}
	}

	// Also check children of the node for inline attributes
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == "attribute_list" {
			if isTestAttrList(source, child) {
				return true
			}
		}
	}
	return false
}

var testAttributeNames = map[string]bool{
	"Test":       true,
	"Fact":       true,
	"Theory":     true,
	"TestMethod": true,
}

func isTestAttrList(source []byte, attrList *sitter.Node) bool {
	for i := 0; i < int(attrList.NamedChildCount()); i++ {
		attr := attrList.NamedChild(i)
		if attr.Type() == "attribute" {
			nameNode := attr.ChildByFieldName("name")
			if nameNode != nil {
				name := nameNode.Content(source)
				if testAttributeNames[name] {
					return true
				}
			}
		}
	}
	return false
}

// getCSharpDocstring extracts XML doc comments (///) preceding a node.
func getCSharpDocstring(source []byte, node *sitter.Node) string {
	prev := node.PrevNamedSibling()
	if prev == nil || prev.Type() != "comment" {
		return ""
	}

	// Check comment is directly above
	if int(node.StartPoint().Row)-int(prev.EndPoint().Row) > 1 {
		return ""
	}

	// Collect consecutive comment lines
	var comments []string
	cur := prev
	for cur != nil && cur.Type() == "comment" {
		line := cur.Content(source)
		line = strings.TrimPrefix(line, "///")
		line = strings.TrimPrefix(line, "//")
		line = strings.TrimSpace(line)
		// Strip XML tags
		line = stripXMLTags(line)
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

func stripXMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func hashCSharpBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
