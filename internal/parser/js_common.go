package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// parseJSImport extracts an import symbol from an import_statement node.
func parseJSImport(filePath string, lang string, source []byte, node *sitter.Node) Symbol {
	return Symbol{
		Name:      node.Content(source),
		Type:      "import",
		Language:  lang,
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

// parseJSClass extracts class and method symbols from a class_declaration node.
func parseJSClass(filePath string, lang string, source []byte, node *sitter.Node) []Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	className := nameNode.Content(source)

	sig := buildClassSignature(source, node)

	symbols := []Symbol{{
		Name:      className,
		Type:      "class",
		Language:  lang,
		Signature: sig,
		Docstring: getJSDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJSBody(source, node),
	}}

	// Extract methods from class body
	body := node.ChildByFieldName("body")
	if body == nil {
		return symbols
	}
	for i := 0; i < int(body.NamedChildCount()); i++ {
		child := body.NamedChild(i)
		if child.Type() == "method_definition" {
			m := parseJSMethod(filePath, lang, source, child, className)
			symbols = append(symbols, m)
		}
	}
	return symbols
}

// parseJSMethod extracts a method symbol from a method_definition node.
func parseJSMethod(filePath string, lang string, source []byte, node *sitter.Node, className string) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	params := node.ChildByFieldName("parameters")
	sig := "method " + name
	if params != nil {
		sig = name + params.Content(source)
	}

	return Symbol{
		Name:      className + "." + name,
		Type:      "method",
		Language:  lang,
		Signature: sig,
		Docstring: getJSDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJSBody(source, node),
	}
}

// parseJSFunction extracts a function symbol from a function_declaration node.
func parseJSFunction(filePath string, lang string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := buildFuncSignature(source, node)
	symType := "function"
	if isTestName(name) {
		symType = "test"
	}

	return Symbol{
		Name:      name,
		Type:      symType,
		Language:  lang,
		Signature: sig,
		Docstring: getJSDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJSBody(source, node),
	}
}

// parseJSArrowFunction extracts a symbol from a lexical_declaration containing an arrow function.
// Returns ok=true if it was an arrow function, false otherwise.
func parseJSArrowFunction(filePath string, lang string, source []byte, node *sitter.Node) (Symbol, bool) {
	if node.NamedChildCount() == 0 {
		return Symbol{}, false
	}
	declarator := node.NamedChild(0)
	if declarator.Type() != "variable_declarator" {
		return Symbol{}, false
	}

	nameNode := declarator.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}, false
	}

	valueNode := declarator.ChildByFieldName("value")
	if valueNode == nil {
		return Symbol{}, false
	}

	// Check if value is an arrow function
	if valueNode.Type() != "arrow_function" {
		return Symbol{}, false
	}

	name := nameNode.Content(source)
	sig := "const " + name + " = "

	// Build arrow function signature
	params := valueNode.ChildByFieldName("parameters")
	if params != nil {
		sig += params.Content(source)
	}

	// Check for return type annotation (TypeScript)
	retType := valueNode.ChildByFieldName("return_type")
	if retType != nil {
		sig += " " + retType.Content(source)
	}

	sig += " => ..."

	return Symbol{
		Name:      name,
		Type:      "function",
		Language:  lang,
		Signature: sig,
		Docstring: getJSDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJSBody(source, node),
	}, true
}

// parseJSExportStatement extracts an export symbol.
func parseJSExportStatement(filePath string, lang string, source []byte, node *sitter.Node) Symbol {
	return Symbol{
		Name:      node.Content(source),
		Type:      "export",
		Language:  lang,
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
	}
}

// parseJSDescribe extracts test symbols from describe/it/test call expressions.
func parseJSDescribe(filePath string, lang string, source []byte, node *sitter.Node) []Symbol {
	var symbols []Symbol
	extractTestCalls(filePath, lang, source, node, &symbols)
	return symbols
}

func extractTestCalls(filePath string, lang string, source []byte, node *sitter.Node, symbols *[]Symbol) {
	if node.Type() == "call_expression" {
		fn := node.ChildByFieldName("function")
		if fn != nil {
			name := fn.Content(source)
			if name == "describe" || name == "it" || name == "test" {
				args := node.ChildByFieldName("arguments")
				testName := name
				if args != nil && args.NamedChildCount() > 0 {
					firstArg := args.NamedChild(0)
					if firstArg.Type() == "string" {
						testName = name + "(" + firstArg.Content(source) + ")"
					}
				}
				*symbols = append(*symbols, Symbol{
					Name:      testName,
					Type:      "test",
					Language:  lang,
					StartLine: int(node.StartPoint().Row) + 1,
					EndLine:   int(node.EndPoint().Row) + 1,
					FilePath:  filePath,
					BodyHash:  hashJSBody(source, node),
				})
			}
		}
	}
	// Recurse into children to find nested it/test calls
	for i := 0; i < int(node.NamedChildCount()); i++ {
		extractTestCalls(filePath, lang, source, node.NamedChild(i), symbols)
	}
}

func buildClassSignature(source []byte, node *sitter.Node) string {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return "class"
	}
	sig := "class " + nameNode.Content(source)

	// Look for extends/implements in the node's children
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		switch child.Type() {
		case "class_heritage":
			sig += " " + child.Content(source)
		}
	}
	return sig
}

func buildFuncSignature(source []byte, node *sitter.Node) string {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return "function"
	}
	sig := "function " + nameNode.Content(source)

	params := node.ChildByFieldName("parameters")
	if params != nil {
		sig += params.Content(source)
	}

	retType := node.ChildByFieldName("return_type")
	if retType != nil {
		sig += retType.Content(source)
	}
	return sig
}

// getJSDocstring extracts the JSDoc comment preceding a node.
func getJSDocstring(source []byte, node *sitter.Node) string {
	prev := node.PrevNamedSibling()
	if prev == nil || prev.Type() != "comment" {
		return ""
	}
	// Check comment is directly above
	if int(node.StartPoint().Row)-int(prev.EndPoint().Row) > 1 {
		return ""
	}

	raw := prev.Content(source)
	// Strip JSDoc markers
	raw = strings.TrimPrefix(raw, "/**")
	raw = strings.TrimSuffix(raw, "*/")
	raw = strings.TrimPrefix(raw, "//")

	// Clean up each line
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

func isTestName(name string) bool {
	return strings.HasPrefix(name, "test") || strings.HasPrefix(name, "Test") ||
		name == "describe" || name == "it"
}

func hashJSBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
