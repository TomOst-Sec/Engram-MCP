package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJavaScriptParserSymbolCount(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.js")
	require.NoError(t, err)

	p := NewJavaScriptParser()
	symbols, err := p.Parse("testdata/sample.js", source)
	require.NoError(t, err)

	// Should have: 2 ES6 imports, 1 require, 2 classes + methods, functions, exports, tests
	assert.GreaterOrEqual(t, len(symbols), 8, "expected at least 8 symbols, got %d", len(symbols))
}

func TestJavaScriptParserFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.js")
	require.NoError(t, err)

	p := NewJavaScriptParser()
	symbols, err := p.Parse("testdata/sample.js", source)
	require.NoError(t, err)

	pd := findByName(symbols, "processData")
	require.NotNil(t, pd, "processData not found")
	assert.Equal(t, "function", pd.Type)
	assert.Equal(t, "javascript", pd.Language)
	assert.Contains(t, pd.Signature, "processData")
}

func TestJavaScriptParserArrowFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.js")
	require.NoError(t, err)

	p := NewJavaScriptParser()
	symbols, err := p.Parse("testdata/sample.js", source)
	require.NoError(t, err)

	fn := findByName(symbols, "greet")
	require.NotNil(t, fn, "greet arrow function not found")
	assert.Equal(t, "function", fn.Type)
	assert.Contains(t, fn.Signature, "greet")
}

func TestJavaScriptParserClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.js")
	require.NoError(t, err)

	p := NewJavaScriptParser()
	symbols, err := p.Parse("testdata/sample.js", source)
	require.NoError(t, err)

	db := findByName(symbols, "Database")
	require.NotNil(t, db, "Database class not found")
	assert.Equal(t, "class", db.Type)
	assert.Contains(t, db.Docstring, "database connection")

	cached := findByName(symbols, "CachedDatabase")
	require.NotNil(t, cached, "CachedDatabase class not found")
	assert.Equal(t, "class", cached.Type)
}

func TestJavaScriptParserMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.js")
	require.NoError(t, err)

	p := NewJavaScriptParser()
	symbols, err := p.Parse("testdata/sample.js", source)
	require.NoError(t, err)

	methods := filterByType(symbols, "method")
	assert.GreaterOrEqual(t, len(methods), 2, "expected at least 2 methods")

	connect := findByName(symbols, "Database.connect")
	require.NotNil(t, connect, "Database.connect not found")
	assert.Equal(t, "method", connect.Type)
}

func TestJavaScriptParserImportsES6(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.js")
	require.NoError(t, err)

	p := NewJavaScriptParser()
	symbols, err := p.Parse("testdata/sample.js", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 3, "expected at least 3 imports (2 ES6 + 1 require)")
}

func TestJavaScriptParserExports(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.js")
	require.NoError(t, err)

	p := NewJavaScriptParser()
	symbols, err := p.Parse("testdata/sample.js", source)
	require.NoError(t, err)

	exports := filterByType(symbols, "export")
	assert.GreaterOrEqual(t, len(exports), 1, "expected at least 1 export")
}

func TestJavaScriptParserTestFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.js")
	require.NoError(t, err)

	p := NewJavaScriptParser()
	symbols, err := p.Parse("testdata/sample.js", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.GreaterOrEqual(t, len(tests), 2, "expected at least 2 test functions")
}

func TestJavaScriptParserDocstrings(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.js")
	require.NoError(t, err)

	p := NewJavaScriptParser()
	symbols, err := p.Parse("testdata/sample.js", source)
	require.NoError(t, err)

	pd := findByName(symbols, "processData")
	require.NotNil(t, pd)
	assert.Contains(t, pd.Docstring, "Process incoming data")
}

func TestJavaScriptParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.js")
	require.NoError(t, err)

	p := NewJavaScriptParser()
	symbols, err := p.Parse("testdata/sample.js", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}

func TestJavaScriptRegistryRouting(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewJavaScriptParser())

	for _, ext := range []string{"app.js", "component.jsx", "module.mjs", "common.cjs"} {
		p, ok := reg.ParserFor(ext)
		assert.True(t, ok, "should find parser for %s", ext)
		assert.Equal(t, "javascript", p.Language())
	}
}
