package parser

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/lua"
)

// LuaParser extracts symbols from Lua source files using tree-sitter.
type LuaParser struct{}

// NewLuaParser creates a new Lua parser.
func NewLuaParser() *LuaParser { return &LuaParser{} }

func (p *LuaParser) Language() string     { return "lua" }
func (p *LuaParser) Extensions() []string { return []string{".lua"} }

func (p *LuaParser) Parse(filePath string, source []byte) ([]Symbol, error) {
	root := sitter.Parse(source, lua.GetLanguage())
	if root == nil {
		return nil, fmt.Errorf("failed to parse %s", filePath)
	}

	var symbols []Symbol
	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		switch child.Type() {
		case "function_statement":
			symbols = append(symbols, p.parseFunction(filePath, source, child))
		case "variable_declaration":
			if sym, ok := p.parseVarDecl(filePath, source, child); ok {
				symbols = append(symbols, sym)
			}
		}
	}
	return symbols, nil
}

func (p *LuaParser) parseFunction(filePath string, source []byte, node *sitter.Node) Symbol {
	name := ""
	params := ""
	isMethod := false

	// Look for function_name or identifier
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "function_name":
			name = child.Content(source)
			// Check for colon syntax (method)
			for j := 0; j < int(child.ChildCount()); j++ {
				sub := child.Child(j)
				if sub != nil && sub.Type() == "table_colon" {
					isMethod = true
				}
			}
		case "identifier":
			if name == "" {
				name = child.Content(source)
			}
		case "parameter_list":
			params = "(" + child.Content(source) + ")"
		}
	}

	sig := "function " + name + params

	symType := "function"
	if isMethod {
		symType = "method"
	}
	if strings.Contains(name, "test_") || strings.HasPrefix(name, "test") {
		symType = "test"
	}

	return Symbol{
		Name:      name,
		Type:      symType,
		Language:  "lua",
		Signature: sig,
		Docstring: getLuaDocstring(source, node),
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		FilePath:  filePath,
		BodyHash:  hashLuaBody(source, node),
	}
}

func (p *LuaParser) parseVarDecl(filePath string, source []byte, node *sitter.Node) (Symbol, bool) {
	// Check for require() calls
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "function_call" {
			// Check if it's a require call
			for j := 0; j < int(child.NamedChildCount()); j++ {
				sub := child.NamedChild(j)
				if sub.Type() == "identifier" && sub.Content(source) == "require" {
					return Symbol{
						Name:      node.Content(source),
						Type:      "import",
						Language:  "lua",
						StartLine: int(node.StartPoint().Row) + 1,
						EndLine:   int(node.EndPoint().Row) + 1,
						FilePath:  filePath,
					}, true
				}
			}
		}
	}
	return Symbol{}, false
}

// getLuaDocstring extracts --- EmmyLua/LDoc style comments.
func getLuaDocstring(source []byte, node *sitter.Node) string {
	// Look for emmy_documentation child
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "emmy_documentation" {
			raw := child.Content(source)
			lines := strings.Split(raw, "\n")
			var cleaned []string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				line = strings.TrimPrefix(line, "---")
				line = strings.TrimSpace(line)
				if line != "" {
					cleaned = append(cleaned, line)
				}
			}
			return strings.Join(cleaned, "\n")
		}
	}
	return ""
}

func hashLuaBody(source []byte, node *sitter.Node) string {
	body := source[node.StartByte():node.EndByte()]
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h)
}
