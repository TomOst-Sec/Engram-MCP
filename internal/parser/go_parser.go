package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

// GoParser extracts symbols from Go source files using tree-sitter.
type GoParser struct{}

// NewGoParser creates a new Go parser.
func NewGoParser() *GoParser { return &GoParser{} }

func (p *GoParser) Language() string     { return "go" }
func (p *GoParser) Extensions() []string { return []string{".go"} }

func (p *GoParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, golang.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		switch child.Type() {
		case "function_declaration":
			symbols = append(symbols, p.parseFunction(filePath, source, child))
		case "method_declaration":
			symbols = append(symbols, p.parseMethod(filePath, source, child))
		case "type_declaration":
			symbols = append(symbols, p.parseTypeDecl(filePath, source, child)...)
		case "import_declaration":
			symbols = append(symbols, p.parseImports(filePath, source, child)...)
		}
	}
	return symbols, nil
}

func (p *GoParser) parseFunction(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	name := nameNode.Content(source)

	sig := node.Content(source)
	// Truncate signature at the body opening brace
	if idx := strings.Index(sig, " {"); idx > 0 {
		sig = sig[:idx]
	}

	symType := "function"
	if strings.HasPrefix(name, "Test") {
		symType = "test"
	}

	return Symbol{
		Name:      name,
		Type:      symType,
		Language:  "go",
		Signature: sig,
		Docstring: getGoDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashBody(source, node),
	}
}

func (p *GoParser) parseMethod(filePath string, source []byte, node *sitter.Node) Symbol {
	nameNode := node.ChildByFieldName("name")
	name := nameNode.Content(source)

	sig := node.Content(source)
	if idx := strings.Index(sig, " {"); idx > 0 {
		sig = sig[:idx]
	}

	return Symbol{
		Name:      name,
		Type:      "method",
		Language:  "go",
		Signature: sig,
		Docstring: getGoDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashBody(source, node),
	}
}

func (p *GoParser) parseTypeDecl(filePath string, source []byte, node *sitter.Node) []Symbol {
	var symbols []Symbol
	for i := 0; i < int(node.NamedChildCount()); i++ {
		spec := node.NamedChild(i)
		if spec.Type() != "type_spec" {
			continue
		}
		nameNode := spec.ChildByFieldName("name")
		if nameNode == nil {
			continue
		}
		name := nameNode.Content(source)

		typeNode := spec.ChildByFieldName("type")
		symType := "type"
		if typeNode != nil && typeNode.Type() == "interface_type" {
			symType = "interface"
		}

		symbols = append(symbols, Symbol{
			Name:      name,
			Type:      symType,
			Language:  "go",
			Signature: "type " + name + " " + typeNodeKind(typeNode, source),
			Docstring: getGoDocstring(source, node),
			StartLine: int(node.StartPoint().Row) + 1,
			EndLine:   int(node.EndPoint().Row) + 1,
			FilePath:  filePath,
			BodyHash:  hashBody(source, node),
		})
	}
	return symbols
}

func (p *GoParser) parseImports(filePath string, source []byte, node *sitter.Node) []Symbol {
	var symbols []Symbol
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "import_spec" {
			path := child.ChildByFieldName("path")
			if path == nil {
				continue
			}
			importPath := strings.Trim(path.Content(source), "\"")
			symbols = append(symbols, Symbol{
				Name:      importPath,
				Type:      "import",
				Language:  "go",
				StartLine: int(child.StartPoint().Row) + 1,
				EndLine:   int(child.EndPoint().Row) + 1,
				FilePath:  filePath,
			})
		} else if child.Type() == "import_spec_list" {
			for j := 0; j < int(child.NamedChildCount()); j++ {
				spec := child.NamedChild(j)
				if spec.Type() != "import_spec" {
					continue
				}
				path := spec.ChildByFieldName("path")
				if path == nil {
					continue
				}
				importPath := strings.Trim(path.Content(source), "\"")
				symbols = append(symbols, Symbol{
					Name:      importPath,
					Type:      "import",
					Language:  "go",
					StartLine: int(spec.StartPoint().Row) + 1,
					EndLine:   int(spec.EndPoint().Row) + 1,
					FilePath:  filePath,
				})
			}
		}
	}
	return symbols
}

func typeNodeKind(n *sitter.Node, source []byte) string {
	if n == nil {
		return ""
	}
	switch n.Type() {
	case "struct_type":
		return "struct"
	case "interface_type":
		return "interface"
	default:
		return n.Content(source)
	}
}

// getGoDocstring extracts the comment immediately preceding a node.
func getGoDocstring(source []byte, node *sitter.Node) string {
	prev := node.PrevNamedSibling()
	if prev == nil || prev.Type() != "comment" {
		return ""
	}
	// Check this comment is directly above (no blank line gap)
	if int(node.StartPoint().Row)-int(prev.EndPoint().Row) > 1 {
		return ""
	}

	// Collect consecutive comment lines above the node
	var comments []string
	cur := prev
	for cur != nil && cur.Type() == "comment" {
		comments = append([]string{strings.TrimPrefix(cur.Content(source), "// ")}, comments...)
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

func hashBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
