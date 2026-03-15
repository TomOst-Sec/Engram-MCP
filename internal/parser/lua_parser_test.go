package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLuaParserSymbolCount(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.lua")
	require.NoError(t, err)

	p := NewLuaParser()
	symbols, err := p.Parse("testdata/sample.lua", source)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(symbols), 8, "expected at least 8 symbols, got %d", len(symbols))
}

func TestLuaParserFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.lua")
	require.NoError(t, err)

	p := NewLuaParser()
	symbols, err := p.Parse("testdata/sample.lua", source)
	require.NoError(t, err)

	gf := findByName(symbols, "globalFunc")
	require.NotNil(t, gf, "globalFunc not found")
	assert.Equal(t, "function", gf.Type)
	assert.Equal(t, "lua", gf.Language)

	helper := findByName(symbols, "helper")
	require.NotNil(t, helper, "helper local function not found")
	assert.Equal(t, "function", helper.Type)
}

func TestLuaParserMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.lua")
	require.NoError(t, err)

	p := NewLuaParser()
	symbols, err := p.Parse("testdata/sample.lua", source)
	require.NoError(t, err)

	greet := findByName(symbols, "User:greet")
	require.NotNil(t, greet, "User:greet not found")
	assert.Equal(t, "method", greet.Type)
	assert.Contains(t, greet.Docstring, "greeting string")

	find := findByName(symbols, "User.find")
	require.NotNil(t, find, "User.find not found")
	assert.Equal(t, "function", find.Type)
}

func TestLuaParserRequires(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.lua")
	require.NoError(t, err)

	p := NewLuaParser()
	symbols, err := p.Parse("testdata/sample.lua", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 2, "expected at least 2 require imports")
}

func TestLuaParserDocComments(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.lua")
	require.NoError(t, err)

	p := NewLuaParser()
	symbols, err := p.Parse("testdata/sample.lua", source)
	require.NoError(t, err)

	newMethod := findByName(symbols, "User:new")
	require.NotNil(t, newMethod)
	assert.Contains(t, newMethod.Docstring, "new user instance")
}

func TestLuaParserTestFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.lua")
	require.NoError(t, err)

	p := NewLuaParser()
	symbols, err := p.Parse("testdata/sample.lua", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.GreaterOrEqual(t, len(tests), 2, "expected at least 2 test functions")
}

func TestLuaParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.lua")
	require.NoError(t, err)

	p := NewLuaParser()
	symbols, err := p.Parse("testdata/sample.lua", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}

func TestLuaRegistryRouting(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewLuaParser())

	p, ok := reg.ParserFor("init.lua")
	assert.True(t, ok)
	assert.Equal(t, "lua", p.Language())
}
