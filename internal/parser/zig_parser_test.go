package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZigParserSymbolCount(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.zig")
	require.NoError(t, err)

	p := NewZigParser()
	symbols, err := p.Parse("testdata/sample.zig", source)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(symbols), 8, "expected at least 8 symbols, got %d", len(symbols))
}

func TestZigParserFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.zig")
	require.NoError(t, err)

	p := NewZigParser()
	symbols, err := p.Parse("testdata/sample.zig", source)
	require.NoError(t, err)

	add := findByName(symbols, "add")
	require.NotNil(t, add, "add function not found")
	assert.Equal(t, "function", add.Type)
	assert.Equal(t, "zig", add.Language)
	assert.Contains(t, add.Signature, "pub fn add")
	assert.Contains(t, add.Docstring, "Adds two numbers")

	helper := findByName(symbols, "helper")
	require.NotNil(t, helper, "helper function not found")
	assert.Equal(t, "function", helper.Type)
}

func TestZigParserStructs(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.zig")
	require.NoError(t, err)

	p := NewZigParser()
	symbols, err := p.Parse("testdata/sample.zig", source)
	require.NoError(t, err)

	point := findByName(symbols, "Point")
	require.NotNil(t, point, "Point struct not found")
	assert.Equal(t, "type", point.Type)
	assert.Contains(t, point.Docstring, "2D point")
}

func TestZigParserEnums(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.zig")
	require.NoError(t, err)

	p := NewZigParser()
	symbols, err := p.Parse("testdata/sample.zig", source)
	require.NoError(t, err)

	color := findByName(symbols, "Color")
	require.NotNil(t, color, "Color enum not found")
	assert.Equal(t, "enum", color.Type)
}

func TestZigParserImports(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.zig")
	require.NoError(t, err)

	p := NewZigParser()
	symbols, err := p.Parse("testdata/sample.zig", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 2, "expected at least 2 imports")
}

func TestZigParserTestBlocks(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.zig")
	require.NoError(t, err)

	p := NewZigParser()
	symbols, err := p.Parse("testdata/sample.zig", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.GreaterOrEqual(t, len(tests), 2, "expected at least 2 test blocks")
}

func TestZigParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.zig")
	require.NoError(t, err)

	p := NewZigParser()
	symbols, err := p.Parse("testdata/sample.zig", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}

func TestZigRegistryRouting(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewZigParser())

	p, ok := reg.ParserFor("main.zig")
	assert.True(t, ok)
	assert.Equal(t, "zig", p.Language())
}
