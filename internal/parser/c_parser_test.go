package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCParserSymbolCount(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.c")
	require.NoError(t, err)

	p := NewCParser()
	symbols, err := p.Parse("testdata/sample.c", source)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(symbols), 10, "expected at least 10 symbols, got %d", len(symbols))
}

func TestCParserFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.c")
	require.NoError(t, err)

	p := NewCParser()
	symbols, err := p.Parse("testdata/sample.c", source)
	require.NoError(t, err)

	add := findByName(symbols, "add")
	require.NotNil(t, add, "add function not found")
	assert.Equal(t, "function", add.Type)
	assert.Equal(t, "c", add.Language)
	assert.Contains(t, add.Signature, "int")
	assert.Contains(t, add.Signature, "add")
	assert.Contains(t, add.Docstring, "Adds two integers")

	process := findByName(symbols, "process")
	require.NotNil(t, process)
	assert.Contains(t, process.Signature, "void")
	assert.Contains(t, process.Signature, "const char")
}

func TestCParserStructs(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.c")
	require.NoError(t, err)

	p := NewCParser()
	symbols, err := p.Parse("testdata/sample.c", source)
	require.NoError(t, err)

	point := findByName(symbols, "Point")
	require.NotNil(t, point, "Point typedef struct not found")
	assert.Equal(t, "type", point.Type)

	user := findByName(symbols, "User")
	require.NotNil(t, user, "User struct not found")
}

func TestCParserEnums(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.c")
	require.NoError(t, err)

	p := NewCParser()
	symbols, err := p.Parse("testdata/sample.c", source)
	require.NoError(t, err)

	color := findByName(symbols, "Color")
	require.NotNil(t, color, "Color enum not found")
	assert.Equal(t, "enum", color.Type)
}

func TestCParserMacros(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.c")
	require.NoError(t, err)

	p := NewCParser()
	symbols, err := p.Parse("testdata/sample.c", source)
	require.NoError(t, err)

	maxSize := findByName(symbols, "MAX_SIZE")
	require.NotNil(t, maxSize, "MAX_SIZE macro not found")
	assert.Equal(t, "macro", maxSize.Type)

	square := findByName(symbols, "SQUARE")
	require.NotNil(t, square, "SQUARE macro not found")
	assert.Equal(t, "macro", square.Type)
}

func TestCParserIncludes(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.c")
	require.NoError(t, err)

	p := NewCParser()
	symbols, err := p.Parse("testdata/sample.c", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 2, "expected at least 2 includes")
}

func TestCParserTypedefs(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.c")
	require.NoError(t, err)

	p := NewCParser()
	symbols, err := p.Parse("testdata/sample.c", source)
	require.NoError(t, err)

	cb := findByName(symbols, "callback_fn")
	require.NotNil(t, cb, "callback_fn typedef not found")
	assert.Equal(t, "type", cb.Type)
}

func TestCParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.c")
	require.NoError(t, err)

	p := NewCParser()
	symbols, err := p.Parse("testdata/sample.c", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}

func TestCRegistryRouting(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewCParser())

	p, ok := reg.ParserFor("main.c")
	assert.True(t, ok)
	assert.Equal(t, "c", p.Language())

	p, ok = reg.ParserFor("utils.h")
	assert.True(t, ok)
	assert.Equal(t, "c", p.Language())
}
