package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwiftParserLanguage(t *testing.T) {
	p := NewSwiftParser()
	assert.Equal(t, "swift", p.Language())
	assert.Equal(t, []string{".swift"}, p.Extensions())
}

func TestSwiftParserFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.swift")
	require.NoError(t, err)

	p := NewSwiftParser()
	symbols, err := p.Parse("testdata/sample.swift", source)
	require.NoError(t, err)

	pd := findByName(symbols, "processData")
	require.NotNil(t, pd, "processData not found")
	assert.Equal(t, "function", pd.Type)
	assert.Equal(t, "swift", pd.Language)
	assert.Contains(t, pd.Signature, "processData")
	assert.Greater(t, pd.StartLine, 0)
	assert.GreaterOrEqual(t, pd.EndLine, pd.StartLine)
	assert.NotEmpty(t, pd.BodyHash)

	hf := findByName(symbols, "helperFunction")
	require.NotNil(t, hf, "helperFunction not found")
	assert.Equal(t, "function", hf.Type)
}

func TestSwiftParserDocstrings(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.swift")
	require.NoError(t, err)

	p := NewSwiftParser()
	symbols, err := p.Parse("testdata/sample.swift", source)
	require.NoError(t, err)

	pd := findByName(symbols, "processData")
	require.NotNil(t, pd)
	assert.Contains(t, pd.Docstring, "Processes input data")
}

func TestSwiftParserClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.swift")
	require.NoError(t, err)

	p := NewSwiftParser()
	symbols, err := p.Parse("testdata/sample.swift", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user, "User class not found")
	assert.Equal(t, "class", user.Type)
	assert.Contains(t, user.Docstring, "user in the system")
}

func TestSwiftParserStructs(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.swift")
	require.NoError(t, err)

	p := NewSwiftParser()
	symbols, err := p.Parse("testdata/sample.swift", source)
	require.NoError(t, err)

	pt := findByName(symbols, "Point")
	require.NotNil(t, pt, "Point struct not found")
	assert.Equal(t, "type", pt.Type)
	assert.Contains(t, pt.Docstring, "point in 2D")
}

func TestSwiftParserProtocols(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.swift")
	require.NoError(t, err)

	p := NewSwiftParser()
	symbols, err := p.Parse("testdata/sample.swift", source)
	require.NoError(t, err)

	pr := findByName(symbols, "Printable")
	require.NotNil(t, pr, "Printable protocol not found")
	assert.Equal(t, "interface", pr.Type)
}

func TestSwiftParserEnums(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.swift")
	require.NoError(t, err)

	p := NewSwiftParser()
	symbols, err := p.Parse("testdata/sample.swift", source)
	require.NoError(t, err)

	as := findByName(symbols, "AppState")
	require.NotNil(t, as, "AppState enum not found")
	assert.Equal(t, "enum", as.Type)
}

func TestSwiftParserMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.swift")
	require.NoError(t, err)

	p := NewSwiftParser()
	symbols, err := p.Parse("testdata/sample.swift", source)
	require.NoError(t, err)

	gdn := findByName(symbols, "User.getDisplayName")
	require.NotNil(t, gdn, "User.getDisplayName not found")
	assert.Equal(t, "method", gdn.Type)
}

func TestSwiftParserConstructors(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.swift")
	require.NoError(t, err)

	p := NewSwiftParser()
	symbols, err := p.Parse("testdata/sample.swift", source)
	require.NoError(t, err)

	ctor := findByName(symbols, "User.init")
	require.NotNil(t, ctor, "User.init not found")
	assert.Equal(t, "constructor", ctor.Type)
}

func TestSwiftParserImports(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.swift")
	require.NoError(t, err)

	p := NewSwiftParser()
	symbols, err := p.Parse("testdata/sample.swift", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 2, "expected at least 2 imports")
}

func TestSwiftParserTestFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.swift")
	require.NoError(t, err)

	p := NewSwiftParser()
	symbols, err := p.Parse("testdata/sample.swift", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.GreaterOrEqual(t, len(tests), 2, "expected at least 2 test functions")
}

func TestSwiftParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.swift")
	require.NoError(t, err)

	p := NewSwiftParser()
	symbols, err := p.Parse("testdata/sample.swift", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}
