package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRubyParserSymbolCount(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rb")
	require.NoError(t, err)

	p := NewRubyParser()
	symbols, err := p.Parse("testdata/sample.rb", source)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(symbols), 15, "expected at least 15 symbols, got %d", len(symbols))
}

func TestRubyParserClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rb")
	require.NoError(t, err)

	p := NewRubyParser()
	symbols, err := p.Parse("testdata/sample.rb", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user, "User class not found")
	assert.Equal(t, "class", user.Type)
	assert.Equal(t, "ruby", user.Language)
	assert.Contains(t, user.Signature, "User")
	assert.Contains(t, user.Signature, "BaseModel")
	assert.Greater(t, user.StartLine, 0)
}

func TestRubyParserDocComments(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rb")
	require.NoError(t, err)

	p := NewRubyParser()
	symbols, err := p.Parse("testdata/sample.rb", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user)
	assert.Contains(t, user.Docstring, "handles user operations")
}

func TestRubyParserMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rb")
	require.NoError(t, err)

	p := NewRubyParser()
	symbols, err := p.Parse("testdata/sample.rb", source)
	require.NoError(t, err)

	greet := findByName(symbols, "User#greet")
	require.NotNil(t, greet, "User#greet not found")
	assert.Equal(t, "method", greet.Type)
	assert.Contains(t, greet.Signature, "greet")
	assert.Contains(t, greet.Docstring, "greeting string")

	init := findByName(symbols, "User#initialize")
	require.NotNil(t, init, "User#initialize not found")
	assert.Equal(t, "method", init.Type)
}

func TestRubyParserClassMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rb")
	require.NoError(t, err)

	p := NewRubyParser()
	symbols, err := p.Parse("testdata/sample.rb", source)
	require.NoError(t, err)

	find := findByName(symbols, "User.find")
	require.NotNil(t, find, "User.find class method not found")
	assert.Equal(t, "function", find.Type)
}

func TestRubyParserModules(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rb")
	require.NoError(t, err)

	p := NewRubyParser()
	symbols, err := p.Parse("testdata/sample.rb", source)
	require.NoError(t, err)

	mod := findByName(symbols, "Serializable")
	require.NotNil(t, mod, "Serializable module not found")
	assert.Equal(t, "type", mod.Type)
	assert.Contains(t, mod.Docstring, "JSON conversion")
}

func TestRubyParserImports(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rb")
	require.NoError(t, err)

	p := NewRubyParser()
	symbols, err := p.Parse("testdata/sample.rb", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 4, "expected at least 4 imports (require, require_relative, include, extend)")
}

func TestRubyParserTestMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rb")
	require.NoError(t, err)

	p := NewRubyParser()
	symbols, err := p.Parse("testdata/sample.rb", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.GreaterOrEqual(t, len(tests), 2, "expected at least 2 test methods")
}

func TestRubyParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rb")
	require.NoError(t, err)

	p := NewRubyParser()
	symbols, err := p.Parse("testdata/sample.rb", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}

func TestRubyRegistryRouting(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewRubyParser())

	p, ok := reg.ParserFor("app.rb")
	assert.True(t, ok, ".rb files should route to Ruby parser")
	assert.Equal(t, "ruby", p.Language())

	_, ok = reg.ParserFor("app.py")
	assert.False(t, ok)
}
