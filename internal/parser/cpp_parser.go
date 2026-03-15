package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/cpp"
)

// CPPParser extracts symbols from C++ source files using tree-sitter.
type CPPParser struct{}

// NewCPPParser creates a new C++ parser.
func NewCPPParser() *CPPParser { return &CPPParser{} }

func (p *CPPParser) Language() string { return "cpp" }
func (p *CPPParser) Extensions() []string {
	return []string{".cpp", ".hpp", ".cc", ".hh", ".cxx", ".hxx"}
}

func (p *CPPParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, cpp.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	p.extractNodes(filePath, source, root, "", &symbols)
	return symbols, nil
}

func (p *CPPParser) extractNodes(filePath string, source []byte, node *sitter.Node, className string, symbols *[]Symbol) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "preproc_include":
			*symbols = append(*symbols, p.parseInclude(filePath, source, child))
		case "using_declaration":
			*symbols = append(*symbols, p.parseUsing(filePath, source, child))
		case "namespace_definition":
			p.parseNamespace(filePath, source, child, symbols)
		case "class_specifier":
			*symbols = append(*symbols, p.parseClass(filePath, source, child)...)
		case "struct_specifier":
			if sym, ok := p.parseStructSpecifier(filePath, source, child); ok {
				*symbols = append(*symbols, sym)
			}
		case "function_definition":
			*symbols = append(*symbols, p.parseFunction(filePath, source, child, className))
		case "template_declaration":
			*symbols = append(*symbols, p.parseTemplate(filePath, source, child, className)...)
		case "declaration":
			p.parseDeclaration(filePath, source, child, className, symbols)
		case "enum_specifier":
			if sym, ok := p.parseEnum(filePath, source, child); ok {
				*symbols = append(*symbols, sym)
			}
		}
	}
}

func (p *CPPParser) parseInclude(filePath string, source []byte, node *sitter.Node) Symbol {
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
		Language:  "cpp",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

func (p *CPPParser) parseUsing(filePath string, source []byte, node *sitter.Node) Symbol {
	content := strings.TrimSuffix(node.Content(source), ";")
	return Symbol{
		Name:      content,
		Type:      "import",
		Language:  "cpp",
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

func (p *CPPParser) parseNamespace(filePath string, source []byte, node *sitter.Node, symbols *[]Symbol) {
	// Recurse into namespace body
	body := node.ChildByFieldName("body")
	if body != nil {
		p.extractNodes(filePath, source, body, "", symbols)
	}
}

func (p *CPPParser) parseClass(filePath string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	className := nameNode.Content(source)

	sig := "class " + className
	// Check for base classes
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == "base_class_clause" {
			sig += " " + child.Content(source)
			break
		}
	}

	symbols := []Symbol{{
		Name:      className,
		Type:      "class",
		Language:  "cpp",
		Signature: sig,
		Docstring: getCDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCPPBody(source, node),
	}}

	// Extract members from field_declaration_list
	body := node.ChildByFieldName("body")
	if body != nil {
		p.extractClassMembers(filePath, source, body, className, &symbols)
	}
	return symbols
}

func (p *CPPParser) extractClassMembers(filePath string, source []byte, body *sitter.Node, className string, symbols *[]Symbol) {
	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		switch child.Type() {
		case "function_definition":
			*symbols = append(*symbols, p.parseClassMethod(filePath, source, child, className))
		case "declaration":
			// Could be a method declaration (prototype) inside the class
			if sym, ok := p.parseFieldDeclarationAsMethod(filePath, source, child, className); ok {
				*symbols = append(*symbols, sym)
			}
		case "template_declaration":
			syms := p.parseTemplate(filePath, source, child, className)
			*symbols = append(*symbols, syms...)
		}
	}
}

func (p *CPPParser) parseClassMethod(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	declarator := node.ChildByFieldName("declarator")
	if declarator == nil {
		return Symbol{}
	}
	name := p.getCPPFunctionName(source, declarator)

	sig := p.buildCPPMethodSignature(source, node)

	symType := "method"
	if name == className {
		symType = "constructor"
	} else if name == "~"+className {
		symType = "constructor"
	}

	return Symbol{
		Name:      className + "." + name,
		Type:      symType,
		Language:  "cpp",
		Signature: sig,
		Docstring: getCDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCPPBody(source, node),
	}
}

func (p *CPPParser) parseFieldDeclarationAsMethod(filePath string, source []byte, node *sitter.Node, className string) (Symbol, bool) {
	// Check if this declaration contains a function declarator (method prototype)
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "function_declarator" {
			name := p.getCPPFunctionName(source, child)
			sig := p.buildCPPMethodSignature(source, node)

			symType := "method"
			if name == className || name == "~"+className {
				symType = "constructor"
			}

			return Symbol{
				Name:      className + "." + name,
				Type:      symType,
				Language:  "cpp",
				Signature: sig,
				Docstring: getCDocstring(source, node),
				StartLine: int(node.StartPoint().Row) + 1,
				EndLine:   int(node.EndPoint().Row) + 1,
				FilePath:  filePath,
				BodyHash:  hashCPPBody(source, node),
			}, true
		}
	}
	return Symbol{}, false
}

func (p *CPPParser) parseFunction(filePath string, source []byte, node *sitter.Node, className string) Symbol {
	declarator := node.ChildByFieldName("declarator")
	if declarator == nil {
		return Symbol{}
	}
	name := p.getCPPFunctionName(source, declarator)
	sig := p.buildCPPFuncSignature(source, node, "")

	return Symbol{
		Name:      name,
		Type:      "function",
		Language:  "cpp",
		Signature: sig,
		Docstring: getCDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCPPBody(source, node),
	}
}

func (p *CPPParser) parseTemplate(filePath string, source []byte, node *sitter.Node, className string) []Symbol {
	// template_declaration wraps the actual declaration
	templatePrefix := ""
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == "template_parameter_list" {
			templatePrefix = "template" + child.Content(source) + " "
			break
		}
	}

	var symbols []Symbol
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "function_definition":
			sym := p.parseFunction(filePath, source, child, className)
			sym.Signature = templatePrefix + sym.Signature
			sym.Docstring = getCDocstring(source, node)
			symbols = append(symbols, sym)
		case "class_specifier":
			syms := p.parseClass(filePath, source, child)
			if len(syms) > 0 {
				syms[0].Signature = templatePrefix + syms[0].Signature
			}
			symbols = append(symbols, syms...)
		case "declaration":
			// Template method prototype
			if sym, ok := p.parseFieldDeclarationAsMethod(filePath, source, child, className); ok {
				sym.Signature = templatePrefix + sym.Signature
				symbols = append(symbols, sym)
			}
		}
	}
	return symbols
}

func (p *CPPParser) parseDeclaration(filePath string, source []byte, node *sitter.Node, className string, symbols *[]Symbol) {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "class_specifier":
			*symbols = append(*symbols, p.parseClass(filePath, source, child)...)
		case "struct_specifier":
			if sym, ok := p.parseStructSpecifier(filePath, source, child); ok {
				*symbols = append(*symbols, sym)
			}
		case "enum_specifier":
			if sym, ok := p.parseEnum(filePath, source, child); ok {
				*symbols = append(*symbols, sym)
			}
		}
	}
}

func (p *CPPParser) parseStructSpecifier(filePath string, source []byte, node *sitter.Node) (Symbol, bool) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}, false
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "type",
		Language:  "cpp",
		Signature: "struct " + name,
		Docstring: getCDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCPPBody(source, node),
	}, true
}

func (p *CPPParser) parseEnum(filePath string, source []byte, node *sitter.Node) (Symbol, bool) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}, false
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "enum",
		Language:  "cpp",
		Signature: "enum " + name,
		Docstring: getCDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashCPPBody(source, node),
	}, true
}

func (p *CPPParser) getCPPFunctionName(source []byte, declarator *sitter.Node) string {
	nameNode := declarator.ChildByFieldName("declarator")
	if nameNode != nil {
		switch nameNode.Type() {
		case "identifier", "field_identifier", "destructor_name", "operator_name":
			return nameNode.Content(source)
		case "qualified_identifier":
			return nameNode.Content(source)
		}
	}
	// Fallback: look for direct identifier children
	for i := 0; i < int(declarator.NamedChildCount()); i++ {
		child := declarator.NamedChild(i)
		if child.Type() == "identifier" || child.Type() == "field_identifier" {
			return child.Content(source)
		}
	}
	return ""
}

func (p *CPPParser) buildCPPFuncSignature(source []byte, node *sitter.Node, prefix string) string {
	content := node.Content(source)
	// Truncate at body
	if idx := strings.Index(content, " {"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	if idx := strings.Index(content, "\n{"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	return strings.TrimSuffix(content, ";")
}

func (p *CPPParser) buildCPPMethodSignature(source []byte, node *sitter.Node) string {
	content := node.Content(source)
	// Truncate at body (handling initializer lists too)
	if idx := strings.Index(content, " {"); idx > 0 {
		sig := content[:idx]
		// Remove initializer list
		if colonIdx := strings.Index(sig, " : "); colonIdx > 0 {
			// Only trim if this looks like an initializer list (after params)
			beforeColon := sig[:colonIdx]
			if strings.Contains(beforeColon, ")") {
				sig = beforeColon
			}
		}
		return strings.TrimSpace(sig)
	}
	if idx := strings.Index(content, "\n{"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	return strings.TrimSuffix(strings.TrimSpace(content), ";")
}

func hashCPPBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
