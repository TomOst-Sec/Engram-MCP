package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCPPParserSymbolCount(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.cpp")
	require.NoError(t, err)

	p := NewCPPParser()
	symbols, err := p.Parse("testdata/sample.cpp", source)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(symbols), 8, "expected at least 8 symbols, got %d", len(symbols))
}

func TestCPPParserClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.cpp")
	require.NoError(t, err)

	p := NewCPPParser()
	symbols, err := p.Parse("testdata/sample.cpp", source)
	require.NoError(t, err)

	vec := findByName(symbols, "Vector2D")
	require.NotNil(t, vec, "Vector2D class not found")
	assert.Equal(t, "class", vec.Type)
	assert.Equal(t, "cpp", vec.Language)
	assert.Contains(t, vec.Signature, "Vector2D")
	assert.Contains(t, vec.Signature, "Printable")
	assert.Contains(t, vec.Docstring, "2D vector class")
}

func TestCPPParserMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.cpp")
	require.NoError(t, err)

	p := NewCPPParser()
	symbols, err := p.Parse("testdata/sample.cpp", source)
	require.NoError(t, err)

	mag := findByName(symbols, "Vector2D.magnitude")
	require.NotNil(t, mag, "Vector2D.magnitude not found")
	assert.Equal(t, "method", mag.Type)
	assert.Contains(t, mag.Signature, "double")
	assert.Contains(t, mag.Signature, "const")
}

func TestCPPParserConstructors(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.cpp")
	require.NoError(t, err)

	p := NewCPPParser()
	symbols, err := p.Parse("testdata/sample.cpp", source)
	require.NoError(t, err)

	ctor := findByName(symbols, "Vector2D.Vector2D")
	require.NotNil(t, ctor, "Vector2D constructor not found")
	assert.Equal(t, "constructor", ctor.Type)
}

func TestCPPParserTemplates(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.cpp")
	require.NoError(t, err)

	p := NewCPPParser()
	symbols, err := p.Parse("testdata/sample.cpp", source)
	require.NoError(t, err)

	maxVal := findByName(symbols, "max_val")
	require.NotNil(t, maxVal, "max_val template function not found")
	assert.Equal(t, "function", maxVal.Type)
	assert.Contains(t, maxVal.Signature, "template")
}

func TestCPPParserEnums(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.cpp")
	require.NoError(t, err)

	p := NewCPPParser()
	symbols, err := p.Parse("testdata/sample.cpp", source)
	require.NoError(t, err)

	dir := findByName(symbols, "Direction")
	require.NotNil(t, dir, "Direction enum not found")
	assert.Equal(t, "enum", dir.Type)
}

func TestCPPParserIncludes(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.cpp")
	require.NoError(t, err)

	p := NewCPPParser()
	symbols, err := p.Parse("testdata/sample.cpp", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 2, "expected at least 2 includes")
}

func TestCPPParserUsing(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.cpp")
	require.NoError(t, err)

	p := NewCPPParser()
	symbols, err := p.Parse("testdata/sample.cpp", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	names := symbolNames(imports)
	found := false
	for _, n := range names {
		if n == "using namespace std" {
			found = true
		}
	}
	assert.True(t, found, "using namespace std not found in imports")
}

func TestCPPParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.cpp")
	require.NoError(t, err)

	p := NewCPPParser()
	symbols, err := p.Parse("testdata/sample.cpp", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}

func TestCPPRegistryRouting(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewCPPParser())

	for _, ext := range []string{"main.cpp", "util.hpp", "lib.cc", "lib.hh", "lib.cxx", "lib.hxx"} {
		p, ok := reg.ParserFor(ext)
		assert.True(t, ok, "%s should route to C++ parser", ext)
		assert.Equal(t, "cpp", p.Language())
	}
}
