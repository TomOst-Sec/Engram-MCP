package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoParserSymbolCount(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.go")
	require.NoError(t, err)

	p := NewGoParser()
	symbols, err := p.Parse("testdata/sample.go", source)
	require.NoError(t, err)

	// Expected: 2 imports, 2 types (User, Stringer), 3 functions (HandleRequest, helperFunc, TestHandleRequest),
	// 2 methods (String, Greet)
	// Total: 9
	assert.GreaterOrEqual(t, len(symbols), 9, "expected at least 9 symbols, got %d", len(symbols))
}

func TestGoParserFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.go")
	require.NoError(t, err)

	p := NewGoParser()
	symbols, err := p.Parse("testdata/sample.go", source)
	require.NoError(t, err)

	funcs := filterByType(symbols, "function")
	assert.Len(t, funcs, 2, "expected 2 functions: HandleRequest, helperFunc")

	hr := findByName(symbols, "HandleRequest")
	require.NotNil(t, hr, "HandleRequest not found")
	assert.Equal(t, "function", hr.Type)
	assert.Equal(t, "go", hr.Language)
	assert.Contains(t, hr.Signature, "HandleRequest")
	assert.Contains(t, hr.Signature, "http.ResponseWriter")
	assert.Greater(t, hr.StartLine, 0)
	assert.GreaterOrEqual(t, hr.EndLine, hr.StartLine)
	assert.NotEmpty(t, hr.BodyHash)
}

func TestGoParserDocstrings(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.go")
	require.NoError(t, err)

	p := NewGoParser()
	symbols, err := p.Parse("testdata/sample.go", source)
	require.NoError(t, err)

	hr := findByName(symbols, "HandleRequest")
	require.NotNil(t, hr)
	assert.Contains(t, hr.Docstring, "processes an incoming HTTP request")
}

func TestGoParserMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.go")
	require.NoError(t, err)

	p := NewGoParser()
	symbols, err := p.Parse("testdata/sample.go", source)
	require.NoError(t, err)

	methods := filterByType(symbols, "method")
	assert.Len(t, methods, 2, "expected 2 methods: String, Greet")

	str := findByName(symbols, "String")
	require.NotNil(t, str)
	assert.Equal(t, "method", str.Type)
	assert.Contains(t, str.Signature, "User")
}

func TestGoParserTypes(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.go")
	require.NoError(t, err)

	p := NewGoParser()
	symbols, err := p.Parse("testdata/sample.go", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user)
	assert.Equal(t, "type", user.Type)
	assert.Contains(t, user.Signature, "struct")

	stringer := findByName(symbols, "Stringer")
	require.NotNil(t, stringer)
	assert.Equal(t, "interface", stringer.Type)
}

func TestGoParserImports(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.go")
	require.NoError(t, err)

	p := NewGoParser()
	symbols, err := p.Parse("testdata/sample.go", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.Len(t, imports, 2, "expected 2 imports: fmt, net/http")
	names := symbolNames(imports)
	assert.Contains(t, names, "fmt")
	assert.Contains(t, names, "net/http")
}

func TestGoParserTestFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.go")
	require.NoError(t, err)

	p := NewGoParser()
	symbols, err := p.Parse("testdata/sample.go", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.Len(t, tests, 1, "expected 1 test function")
	assert.Equal(t, "TestHandleRequest", tests[0].Name)
}

func TestGoParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.go")
	require.NoError(t, err)

	p := NewGoParser()
	symbols, err := p.Parse("testdata/sample.go", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}

// Test helpers

func filterByType(symbols []Symbol, typ string) []Symbol {
	var result []Symbol
	for _, s := range symbols {
		if s.Type == typ {
			result = append(result, s)
		}
	}
	return result
}

func findByName(symbols []Symbol, name string) *Symbol {
	for _, s := range symbols {
		if s.Name == name {
			return &s
		}
	}
	return nil
}

func symbolNames(symbols []Symbol) []string {
	names := make([]string, len(symbols))
	for i, s := range symbols {
		names[i] = s.Name
	}
	return names
}
