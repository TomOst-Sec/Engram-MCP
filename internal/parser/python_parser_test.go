package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPythonParserSymbolCount(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.py")
	require.NoError(t, err)

	p := NewPythonParser()
	symbols, err := p.Parse("testdata/sample.py", source)
	require.NoError(t, err)

	// Expected: 3 imports, 2 classes (UserModel, AdminUser), ~6 methods, 2 functions + 1 test
	assert.GreaterOrEqual(t, len(symbols), 10, "expected at least 10 symbols, got %d", len(symbols))
}

func TestPythonParserFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.py")
	require.NoError(t, err)

	p := NewPythonParser()
	symbols, err := p.Parse("testdata/sample.py", source)
	require.NoError(t, err)

	pr := findByName(symbols, "process_request")
	require.NotNil(t, pr, "process_request not found")
	assert.Equal(t, "function", pr.Type)
	assert.Equal(t, "python", pr.Language)
	assert.Contains(t, pr.Signature, "def process_request")
	assert.Contains(t, pr.Signature, "path")
	assert.Greater(t, pr.StartLine, 0)
	assert.GreaterOrEqual(t, pr.EndLine, pr.StartLine)
	assert.NotEmpty(t, pr.BodyHash)
}

func TestPythonParserDocstrings(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.py")
	require.NoError(t, err)

	p := NewPythonParser()
	symbols, err := p.Parse("testdata/sample.py", source)
	require.NoError(t, err)

	pr := findByName(symbols, "process_request")
	require.NotNil(t, pr)
	assert.Contains(t, pr.Docstring, "Process an incoming HTTP request")
}

func TestPythonParserClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.py")
	require.NoError(t, err)

	p := NewPythonParser()
	symbols, err := p.Parse("testdata/sample.py", source)
	require.NoError(t, err)

	user := findByName(symbols, "UserModel")
	require.NotNil(t, user, "UserModel class not found")
	assert.Equal(t, "class", user.Type)
	assert.Contains(t, user.Docstring, "Represents a user")

	admin := findByName(symbols, "AdminUser")
	require.NotNil(t, admin, "AdminUser class not found")
	assert.Equal(t, "class", admin.Type)
	assert.Contains(t, admin.Signature, "AdminUser")
}

func TestPythonParserMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.py")
	require.NoError(t, err)

	p := NewPythonParser()
	symbols, err := p.Parse("testdata/sample.py", source)
	require.NoError(t, err)

	methods := filterByType(symbols, "method")
	assert.GreaterOrEqual(t, len(methods), 4, "expected at least 4 methods")

	// Methods should be qualified with class name
	displayName := findByName(symbols, "UserModel.display_name")
	require.NotNil(t, displayName, "UserModel.display_name not found")
	assert.Equal(t, "method", displayName.Type)
	assert.Contains(t, displayName.Docstring, "display name")
}

func TestPythonParserImports(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.py")
	require.NoError(t, err)

	p := NewPythonParser()
	symbols, err := p.Parse("testdata/sample.py", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 3, "expected at least 3 imports")
}

func TestPythonParserTestFunctions(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.py")
	require.NoError(t, err)

	p := NewPythonParser()
	symbols, err := p.Parse("testdata/sample.py", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.Len(t, tests, 1, "expected 1 test function")
	assert.Equal(t, "test_process_request", tests[0].Name)
}

func TestPythonParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/sample.py")
	require.NoError(t, err)

	p := NewPythonParser()
	symbols, err := p.Parse("testdata/sample.py", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}
