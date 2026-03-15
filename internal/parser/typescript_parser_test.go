package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeScriptParserSymbolCount(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.ts")
	require.NoError(t, err)

	p := NewTypeScriptParser()
	symbols, err := p.Parse("testdata/sample.ts", source)
	require.NoError(t, err)

	// Should have: 3 imports, 1 interface, 1 type, 1 enum, 2 classes + methods,
	// functions, arrow funcs, exports, tests
	assert.GreaterOrEqual(t, len(symbols), 10, "expected at least 10 symbols, got %d", len(symbols))
}

func TestTypeScriptParserFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.ts")
	require.NoError(t, err)

	p := NewTypeScriptParser()
	symbols, err := p.Parse("testdata/sample.ts", source)
	require.NoError(t, err)

	hr := findByName(symbols, "handleRequest")
	require.NotNil(t, hr, "handleRequest not found")
	assert.Equal(t, "function", hr.Type)
	assert.Equal(t, "typescript", hr.Language)
	assert.Contains(t, hr.Signature, "handleRequest")
	assert.Contains(t, hr.Signature, "Request")
}

func TestTypeScriptParserArrowFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.ts")
	require.NoError(t, err)

	p := NewTypeScriptParser()
	symbols, err := p.Parse("testdata/sample.ts", source)
	require.NoError(t, err)

	fn := findByName(symbols, "formatName")
	require.NotNil(t, fn, "formatName arrow function not found")
	assert.Equal(t, "function", fn.Type)
	assert.Contains(t, fn.Signature, "formatName")
}

func TestTypeScriptParserClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.ts")
	require.NoError(t, err)

	p := NewTypeScriptParser()
	symbols, err := p.Parse("testdata/sample.ts", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user, "User class not found")
	assert.Equal(t, "class", user.Type)
	assert.Contains(t, user.Docstring, "User class implementing")

	admin := findByName(symbols, "AdminUser")
	require.NotNil(t, admin, "AdminUser class not found")
	assert.Equal(t, "class", admin.Type)
}

func TestTypeScriptParserInterfaces(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.ts")
	require.NoError(t, err)

	p := NewTypeScriptParser()
	symbols, err := p.Parse("testdata/sample.ts", source)
	require.NoError(t, err)

	ui := findByName(symbols, "UserInterface")
	require.NotNil(t, ui, "UserInterface not found")
	assert.Equal(t, "interface", ui.Type)
	assert.Contains(t, ui.Signature, "interface UserInterface")
	assert.Contains(t, ui.Docstring, "Represents a user")
}

func TestTypeScriptParserTypeAliases(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.ts")
	require.NoError(t, err)

	p := NewTypeScriptParser()
	symbols, err := p.Parse("testdata/sample.ts", source)
	require.NoError(t, err)

	status := findByName(symbols, "Status")
	require.NotNil(t, status, "Status type not found")
	assert.Equal(t, "type", status.Type)
	assert.Contains(t, status.Signature, "Status")
}

func TestTypeScriptParserEnums(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.ts")
	require.NoError(t, err)

	p := NewTypeScriptParser()
	symbols, err := p.Parse("testdata/sample.ts", source)
	require.NoError(t, err)

	role := findByName(symbols, "Role")
	require.NotNil(t, role, "Role enum not found")
	assert.Equal(t, "enum", role.Type)
}

func TestTypeScriptParserImports(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.ts")
	require.NoError(t, err)

	p := NewTypeScriptParser()
	symbols, err := p.Parse("testdata/sample.ts", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 3, "expected at least 3 imports")
}

func TestTypeScriptParserTestFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.ts")
	require.NoError(t, err)

	p := NewTypeScriptParser()
	symbols, err := p.Parse("testdata/sample.ts", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.GreaterOrEqual(t, len(tests), 2, "expected at least 2 test functions")
}

func TestTypeScriptParserDocstrings(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.ts")
	require.NoError(t, err)

	p := NewTypeScriptParser()
	symbols, err := p.Parse("testdata/sample.ts", source)
	require.NoError(t, err)

	hr := findByName(symbols, "handleRequest")
	require.NotNil(t, hr)
	assert.Contains(t, hr.Docstring, "Process an incoming request")
}

func TestTypeScriptParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.ts")
	require.NoError(t, err)

	p := NewTypeScriptParser()
	symbols, err := p.Parse("testdata/sample.ts", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}

func TestTypeScriptRegistryRouting(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewTypeScriptParser())

	p, ok := reg.ParserFor("app.ts")
	assert.True(t, ok)
	assert.Equal(t, "typescript", p.Language())

	p, ok = reg.ParserFor("component.tsx")
	assert.True(t, ok)
	assert.Equal(t, "typescript", p.Language())
}
