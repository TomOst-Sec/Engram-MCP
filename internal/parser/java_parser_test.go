package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJavaParserLanguage(t *testing.T) {
	p := NewJavaParser()
	assert.Equal(t, "java", p.Language())
	assert.Equal(t, []string{".java"}, p.Extensions())
}

func TestJavaParserClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.java")
	require.NoError(t, err)

	p := NewJavaParser()
	symbols, err := p.Parse("testdata/Sample.java", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user, "User class not found")
	assert.Equal(t, "class", user.Type)
	assert.Equal(t, "java", user.Language)
	assert.Contains(t, user.Signature, "class User")
	assert.Contains(t, user.Signature, "extends")
	assert.Greater(t, user.StartLine, 0)
	assert.GreaterOrEqual(t, user.EndLine, user.StartLine)
}

func TestJavaParserDocstrings(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.java")
	require.NoError(t, err)

	p := NewJavaParser()
	symbols, err := p.Parse("testdata/Sample.java", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user)
	assert.Contains(t, user.Docstring, "Represents a user")
}

func TestJavaParserMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.java")
	require.NoError(t, err)

	p := NewJavaParser()
	symbols, err := p.Parse("testdata/Sample.java", source)
	require.NoError(t, err)

	gdn := findByName(symbols, "User.getDisplayName")
	require.NotNil(t, gdn, "User.getDisplayName not found")
	assert.Equal(t, "method", gdn.Type)
	assert.Contains(t, gdn.Signature, "getDisplayName")
	assert.Contains(t, gdn.Docstring, "display name")

	ve := findByName(symbols, "User.validateEmail")
	require.NotNil(t, ve, "User.validateEmail not found")
	assert.Equal(t, "method", ve.Type)
}

func TestJavaParserConstructors(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.java")
	require.NoError(t, err)

	p := NewJavaParser()
	symbols, err := p.Parse("testdata/Sample.java", source)
	require.NoError(t, err)

	ctor := findByName(symbols, "User.User")
	require.NotNil(t, ctor, "User constructor not found")
	assert.Equal(t, "constructor", ctor.Type)
	assert.Contains(t, ctor.Docstring, "Creates a new User")
}

func TestJavaParserInterfaces(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.java")
	require.NoError(t, err)

	p := NewJavaParser()
	symbols, err := p.Parse("testdata/Sample.java", source)
	require.NoError(t, err)

	ur := findByName(symbols, "UserRepository")
	require.NotNil(t, ur, "UserRepository interface not found")
	assert.Equal(t, "interface", ur.Type)
	assert.Contains(t, ur.Signature, "interface UserRepository")
}

func TestJavaParserEnums(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.java")
	require.NoError(t, err)

	p := NewJavaParser()
	symbols, err := p.Parse("testdata/Sample.java", source)
	require.NoError(t, err)

	role := findByName(symbols, "UserRole")
	require.NotNil(t, role, "UserRole enum not found")
	assert.Equal(t, "enum", role.Type)
	assert.Contains(t, role.Signature, "enum UserRole")
}

func TestJavaParserImports(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.java")
	require.NoError(t, err)

	p := NewJavaParser()
	symbols, err := p.Parse("testdata/Sample.java", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 3, "expected at least 3 import symbols")
}

func TestJavaParserTestMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.java")
	require.NoError(t, err)

	p := NewJavaParser()
	symbols, err := p.Parse("testdata/Sample.java", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.GreaterOrEqual(t, len(tests), 2, "expected at least 2 test methods")
}

func TestJavaParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.java")
	require.NoError(t, err)

	p := NewJavaParser()
	symbols, err := p.Parse("testdata/Sample.java", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}

func TestJavaParserNestedClass(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.java")
	require.NoError(t, err)

	p := NewJavaParser()
	symbols, err := p.Parse("testdata/Sample.java", source)
	require.NoError(t, err)

	prefs := findByName(symbols, "Preferences")
	require.NotNil(t, prefs, "Preferences inner class not found")
	assert.Equal(t, "class", prefs.Type)
}
