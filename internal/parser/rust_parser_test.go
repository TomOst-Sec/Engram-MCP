package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRustParserLanguage(t *testing.T) {
	p := NewRustParser()
	assert.Equal(t, "rust", p.Language())
	assert.Equal(t, []string{".rs"}, p.Extensions())
}

func TestRustParserFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rs")
	require.NoError(t, err)

	p := NewRustParser()
	symbols, err := p.Parse("testdata/sample.rs", source)
	require.NoError(t, err)

	// process_data is a public function
	pd := findByName(symbols, "process_data")
	require.NotNil(t, pd, "process_data not found")
	assert.Equal(t, "function", pd.Type)
	assert.Equal(t, "rust", pd.Language)
	assert.Contains(t, pd.Signature, "process_data")
	assert.Contains(t, pd.Signature, "'a")
	assert.Greater(t, pd.StartLine, 0)
	assert.GreaterOrEqual(t, pd.EndLine, pd.StartLine)
	assert.NotEmpty(t, pd.BodyHash)

	// helper_function is a private function
	hf := findByName(symbols, "helper_function")
	require.NotNil(t, hf, "helper_function not found")
	assert.Equal(t, "function", hf.Type)
	assert.Contains(t, hf.Signature, "helper_function")
}

func TestRustParserDocstrings(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rs")
	require.NoError(t, err)

	p := NewRustParser()
	symbols, err := p.Parse("testdata/sample.rs", source)
	require.NoError(t, err)

	pd := findByName(symbols, "process_data")
	require.NotNil(t, pd)
	assert.Contains(t, pd.Docstring, "Processes input data")

	user := findByName(symbols, "User")
	require.NotNil(t, user)
	assert.Contains(t, user.Docstring, "user in the system")
}

func TestRustParserStructs(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rs")
	require.NoError(t, err)

	p := NewRustParser()
	symbols, err := p.Parse("testdata/sample.rs", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user, "User struct not found")
	assert.Equal(t, "type", user.Type)
	assert.Equal(t, "rust", user.Language)
	assert.Contains(t, user.Signature, "struct User")
	assert.Greater(t, user.StartLine, 0)
}

func TestRustParserEnums(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rs")
	require.NoError(t, err)

	p := NewRustParser()
	symbols, err := p.Parse("testdata/sample.rs", source)
	require.NoError(t, err)

	ae := findByName(symbols, "AppError")
	require.NotNil(t, ae, "AppError enum not found")
	assert.Equal(t, "enum", ae.Type)
	assert.Contains(t, ae.Signature, "enum AppError")
}

func TestRustParserTraits(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rs")
	require.NoError(t, err)

	p := NewRustParser()
	symbols, err := p.Parse("testdata/sample.rs", source)
	require.NoError(t, err)

	pr := findByName(symbols, "Printable")
	require.NotNil(t, pr, "Printable trait not found")
	assert.Equal(t, "interface", pr.Type)
	assert.Contains(t, pr.Signature, "trait Printable")
}

func TestRustParserImplBlocks(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rs")
	require.NoError(t, err)

	p := NewRustParser()
	symbols, err := p.Parse("testdata/sample.rs", source)
	require.NoError(t, err)

	// Methods from impl User
	newUser := findByName(symbols, "User.new")
	require.NotNil(t, newUser, "User.new not found")
	assert.Equal(t, "method", newUser.Type)
	assert.Contains(t, newUser.Signature, "new")

	greet := findByName(symbols, "User.greet")
	require.NotNil(t, greet, "User.greet not found")
	assert.Equal(t, "method", greet.Type)

	// Methods from impl Printable for User
	format := findByName(symbols, "User.format")
	require.NotNil(t, format, "User.format not found")
	assert.Equal(t, "method", format.Type)
}

func TestRustParserImports(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rs")
	require.NoError(t, err)

	p := NewRustParser()
	symbols, err := p.Parse("testdata/sample.rs", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 2, "expected at least 2 import symbols")
}

func TestRustParserTestFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rs")
	require.NoError(t, err)

	p := NewRustParser()
	symbols, err := p.Parse("testdata/sample.rs", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.GreaterOrEqual(t, len(tests), 2, "expected at least 2 test functions")
}

func TestRustParserMacros(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rs")
	require.NoError(t, err)

	p := NewRustParser()
	symbols, err := p.Parse("testdata/sample.rs", source)
	require.NoError(t, err)

	cm := findByName(symbols, "create_map")
	require.NotNil(t, cm, "create_map macro not found")
	assert.Equal(t, "macro", cm.Type)
}

func TestRustParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.rs")
	require.NoError(t, err)

	p := NewRustParser()
	symbols, err := p.Parse("testdata/sample.rs", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}
