package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPHPParserSymbolCount(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(symbols), 15, "expected at least 15 symbols, got %d", len(symbols))
}

func TestPHPParserClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user, "User class not found")
	assert.Equal(t, "class", user.Type)
	assert.Equal(t, "php", user.Language)
	assert.Contains(t, user.Signature, "User")
	assert.Contains(t, user.Signature, "BaseModel")
	assert.Contains(t, user.Signature, "Displayable")
}

func TestPHPParserDocComments(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user)
	assert.Contains(t, user.Docstring, "Represents a user in the system")
}

func TestPHPParserMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	greet := findByName(symbols, "User::greet")
	require.NotNil(t, greet, "User::greet not found")
	assert.Equal(t, "method", greet.Type)
	assert.Contains(t, greet.Signature, "public")
	assert.Contains(t, greet.Signature, "string")
	assert.Contains(t, greet.Docstring, "greeting message")
}

func TestPHPParserVisibilityModifiers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	greet := findByName(symbols, "User::greet")
	require.NotNil(t, greet)
	assert.Contains(t, greet.Signature, "public")

	validate := findByName(symbols, "User::validateEmail")
	require.NotNil(t, validate)
	assert.Contains(t, validate.Signature, "private")
}

func TestPHPParserConstructors(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	ctor := findByName(symbols, "User::__construct")
	require.NotNil(t, ctor, "User::__construct not found")
	assert.Equal(t, "constructor", ctor.Type)
}

func TestPHPParserInterfaces(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	iface := findByName(symbols, "Displayable")
	require.NotNil(t, iface, "Displayable interface not found")
	assert.Equal(t, "interface", iface.Type)
	assert.Contains(t, iface.Docstring, "display contract")
}

func TestPHPParserTraits(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	trait := findByName(symbols, "HasTimestamps")
	require.NotNil(t, trait, "HasTimestamps trait not found")
	assert.Equal(t, "trait", trait.Type)
}

func TestPHPParserFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	helper := findByName(symbols, "helperFunc")
	require.NotNil(t, helper, "helperFunc not found")
	assert.Equal(t, "function", helper.Type)
	assert.Contains(t, helper.Signature, "int")
	assert.Contains(t, helper.Docstring, "standalone helper")
}

func TestPHPParserEnums(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	role := findByName(symbols, "UserRole")
	require.NotNil(t, role, "UserRole enum not found")
	assert.Equal(t, "enum", role.Type)
}

func TestPHPParserImports(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 2, "expected at least 2 use declarations")
}

func TestPHPParserTestMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.GreaterOrEqual(t, len(tests), 2, "expected at least 2 test methods")
	names := symbolNames(tests)
	assert.Contains(t, names, "UserTest::testUserCreation")
	assert.Contains(t, names, "UserTest::testGreeting")
}

func TestPHPParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.php")
	require.NoError(t, err)

	p := NewPHPParser()
	symbols, err := p.Parse("testdata/sample.php", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}

func TestPHPRegistryRouting(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewPHPParser())

	p, ok := reg.ParserFor("index.php")
	assert.True(t, ok, ".php files should route to PHP parser")
	assert.Equal(t, "php", p.Language())

	_, ok = reg.ParserFor("index.rb")
	assert.False(t, ok)
}
