package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKotlinParserLanguage(t *testing.T) {
	p := NewKotlinParser()
	assert.Equal(t, "kotlin", p.Language())
	assert.Equal(t, []string{".kt", ".kts"}, p.Extensions())
}

func TestKotlinParserFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	pd := findByName(symbols, "processData")
	require.NotNil(t, pd, "processData not found")
	assert.Equal(t, "function", pd.Type)
	assert.Equal(t, "kotlin", pd.Language)
	assert.Contains(t, pd.Signature, "processData")
	assert.Greater(t, pd.StartLine, 0)
	assert.NotEmpty(t, pd.BodyHash)

	hf := findByName(symbols, "helperFunction")
	require.NotNil(t, hf, "helperFunction not found")
	assert.Equal(t, "function", hf.Type)
}

func TestKotlinParserDocstrings(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	pd := findByName(symbols, "processData")
	require.NotNil(t, pd)
	assert.Contains(t, pd.Docstring, "Processes input data")

	prefs := findByName(symbols, "UserPreferences")
	require.NotNil(t, prefs)
	assert.Contains(t, prefs.Docstring, "data class for user preferences")
}

func TestKotlinParserClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user, "User class not found")
	assert.Equal(t, "class", user.Type)
	// Note: docstring for the first class after import_list may not be captured
	// due to Kotlin tree-sitter grammar embedding the comment in import_list
}

func TestKotlinParserDataClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	prefs := findByName(symbols, "UserPreferences")
	require.NotNil(t, prefs, "UserPreferences data class not found")
	assert.Equal(t, "class", prefs.Type)
}

func TestKotlinParserInterfaces(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	ur := findByName(symbols, "UserRepository")
	require.NotNil(t, ur, "UserRepository interface not found")
	assert.Equal(t, "interface", ur.Type)
}

func TestKotlinParserObjects(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	ac := findByName(symbols, "AppConfig")
	require.NotNil(t, ac, "AppConfig object not found")
	assert.Equal(t, "type", ac.Type)
}

func TestKotlinParserSealedClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	result := findByName(symbols, "Result")
	require.NotNil(t, result, "Result sealed class not found")
	assert.Equal(t, "class", result.Type)
}

func TestKotlinParserMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	gdn := findByName(symbols, "User.getDisplayName")
	require.NotNil(t, gdn, "User.getDisplayName not found")
	assert.Equal(t, "method", gdn.Type)
}

func TestKotlinParserImports(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 2, "expected at least 2 imports")
}

func TestKotlinParserTestFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.GreaterOrEqual(t, len(tests), 2, "expected at least 2 test methods")
}

func TestKotlinParserExtensionFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	slug := findByName(symbols, "String.toSlug")
	require.NotNil(t, slug, "String.toSlug extension not found")
	assert.Equal(t, "function", slug.Type)
}

func TestKotlinParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.kt")
	require.NoError(t, err)

	p := NewKotlinParser()
	symbols, err := p.Parse("testdata/sample.kt", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}
