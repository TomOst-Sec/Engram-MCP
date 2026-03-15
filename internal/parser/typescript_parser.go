package parser

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// TypeScriptParser extracts symbols from TypeScript source files using tree-sitter.
type TypeScriptParser struct{}

// NewTypeScriptParser creates a new TypeScript parser.
func NewTypeScriptParser() *TypeScriptParser { return &TypeScriptParser{} }

func (p *TypeScriptParser) Language() string     { return "typescript" }
func (p *TypeScriptParser) Extensions() []string { return []string{".ts", ".tsx"} }

func (p *TypeScriptParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, typescript.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		symbols = append(symbols, p.extractNode(filePath, source, child)...)
	}
	return symbols, nil
}

func (p *TypeScriptParser) extractNode(filePath string, source []byte, node *sitter.Node) []Symbol {
	switch node.Type() {
	case "import_statement":
		return []Symbol{parseJSImport(filePath, "typescript", source, node)}

	case "function_declaration":
		return []Symbol{parseJSFunction(filePath, "typescript", source, node)}

	case "class_declaration":
		return parseJSClass(filePath, "typescript", source, node)

	case "interface_declaration":
		return []Symbol{p.parseInterface(filePath, source, node)}

	case "type_alias_declaration":
		return []Symbol{p.parseTypeAlias(filePath, source, node)}

	case "enum_declaration":
		return []Symbol{p.parseEnum(filePath, source, node)}

	case "lexical_declaration":
		if s, ok := parseJSArrowFunction(filePath, "typescript", source, node); ok {
			return []Symbol{s}
		}

	case "export_statement":
		return p.parseExport(filePath, source, node)

	case "expression_statement":
		// Check for describe/it/test blocks
		if node.NamedChildCount() > 0 {
			child := node.NamedChild(0)
			if child.Type() == "call_expression" {
				return parseJSDescribe(filePath, "typescript", source, child)
			}
		}
	}
	return nil
}

func (p *TypeScriptParser) parseInterface(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	sig := "interface " + name
	// Check for extends
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && child.Type() == "extends_type_clause" {
			sig += " " + child.Content(source)
		}
	}

	return Symbol{
		Name:      name,
		Type:      "interface",
		Language:  "typescript",
		Signature: sig,
		Docstring: getJSDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJSBody(source, node),
	}
}

func (p *TypeScriptParser) parseTypeAlias(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "type",
		Language:  "typescript",
		Signature: strings.TrimSuffix(node.Content(source), ";"),
		Docstring: getJSDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJSBody(source, node),
	}
}

func (p *TypeScriptParser) parseEnum(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return Symbol{}
	}
	name := nameNode.Content(source)

	return Symbol{
		Name:      name,
		Type:      "enum",
		Language:  "typescript",
		Signature: "enum " + name,
		Docstring: getJSDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashJSBody(source, node),
	}
}

func (p *TypeScriptParser) parseExport(filePath string, source []byte, node *sitter.Node) []Symbol {
	var symbols []Symbol

	// The JSDoc comment is a sibling of the export_statement, not the inner declaration.
	exportDocstring := getJSDocstring(source, node)

	// Check for re-export with declaration (export class/function/etc.)
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		inner := p.extractNode(filePath, source, child)
		if len(inner) > 0 {
			for j := range inner {
				if child.Type() == "class_declaration" {
					inner[j].Type = "class"
				} else if child.Type() == "function_declaration" {
					inner[j].Type = "function"
				}
				// Inherit docstring from export statement if inner has none
				if inner[j].Docstring == "" && exportDocstring != "" {
					inner[j].Docstring = exportDocstring
				}
			}
			symbols = append(symbols, inner...)
			return symbols
		}
	}

	// Bare export statement like `export { name }`
	symbols = append(symbols, parseJSExportStatement(filePath, "typescript", source, node))
	return symbols
}
