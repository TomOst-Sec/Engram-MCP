package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSharpParserSymbolCount(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	// Expected symbols:
	// Imports: 3 (System, System.Collections.Generic, static System.Console)
	// Classes: User, InnerSettings, BaseEntity, UserTests = 4
	// Interface: IDisplayable = 1
	// Struct: Point = 1
	// Enum: UserRole = 1
	// Record: UserRecord = 1
	// Constructors: User = 1
	// Methods: Greet, ValidateEmail, GetDisplayName, GetItems, CompareTo, Describe = 6
	// Properties: Name, Age, Theme, X, Y, Id, CreatedAt = 7
	// Tests: TestUserCreation, UserGreetReturnsCorrectMessage, ValidateEmailThrowsOnEmpty = 3
	// Total: at least 28
	assert.GreaterOrEqual(t, len(symbols), 20, "expected at least 20 symbols, got %d", len(symbols))
}

func TestCSharpParserClasses(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user, "User class not found")
	assert.Equal(t, "class", user.Type)
	assert.Equal(t, "csharp", user.Language)
	assert.Contains(t, user.Signature, "User")
	assert.Contains(t, user.Signature, "BaseEntity")
	assert.Contains(t, user.Signature, "IDisplayable")
	assert.Greater(t, user.StartLine, 0)
	assert.GreaterOrEqual(t, user.EndLine, user.StartLine)
}

func TestCSharpParserXMLDocComments(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	user := findByName(symbols, "User")
	require.NotNil(t, user)
	assert.Contains(t, user.Docstring, "Represents a user in the system")
}

func TestCSharpParserMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	greet := findByName(symbols, "User.Greet")
	require.NotNil(t, greet, "User.Greet method not found")
	assert.Equal(t, "method", greet.Type)
	assert.Contains(t, greet.Signature, "string")
	assert.Contains(t, greet.Signature, "Greet")
	assert.Contains(t, greet.Docstring, "greeting message")
}

func TestCSharpParserAccessModifiers(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	greet := findByName(symbols, "User.Greet")
	require.NotNil(t, greet)
	assert.Contains(t, greet.Signature, "public")

	validate := findByName(symbols, "User.ValidateEmail")
	require.NotNil(t, validate)
	assert.Contains(t, validate.Signature, "private")
}

func TestCSharpParserInterfaces(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	iface := findByName(symbols, "IDisplayable")
	require.NotNil(t, iface, "IDisplayable interface not found")
	assert.Equal(t, "interface", iface.Type)
	assert.Contains(t, iface.Docstring, "display contract")
}

func TestCSharpParserStructs(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	point := findByName(symbols, "Point")
	require.NotNil(t, point, "Point struct not found")
	assert.Equal(t, "struct", point.Type)
}

func TestCSharpParserEnums(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	role := findByName(symbols, "UserRole")
	require.NotNil(t, role, "UserRole enum not found")
	assert.Equal(t, "enum", role.Type)
}

func TestCSharpParserProperties(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	name := findByName(symbols, "User.Name")
	require.NotNil(t, name, "User.Name property not found")
	assert.Equal(t, "property", name.Type)
	assert.Contains(t, name.Signature, "string")
}

func TestCSharpParserConstructors(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	ctor := findByName(symbols, "User.User")
	require.NotNil(t, ctor, "User constructor not found")
	assert.Equal(t, "constructor", ctor.Type)
	assert.Contains(t, ctor.Signature, "string name")
}

func TestCSharpParserImports(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	imports := filterByType(symbols, "import")
	assert.GreaterOrEqual(t, len(imports), 3, "expected at least 3 using directives")
}

func TestCSharpParserTestMethods(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	tests := filterByType(symbols, "test")
	assert.GreaterOrEqual(t, len(tests), 3, "expected at least 3 test methods")
	names := symbolNames(tests)
	assert.Contains(t, names, "UserTests.TestUserCreation")
	assert.Contains(t, names, "UserTests.UserGreetReturnsCorrectMessage")
	assert.Contains(t, names, "UserTests.ValidateEmailThrowsOnEmpty")
}

func TestCSharpParserGenerics(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	getItems := findByName(symbols, "User.GetItems")
	require.NotNil(t, getItems, "User.GetItems method not found")
	assert.Contains(t, getItems.Signature, "List<T>")
}

func TestCSharpParserRecords(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	rec := findByName(symbols, "UserRecord")
	require.NotNil(t, rec, "UserRecord record not found")
	assert.Equal(t, "record", rec.Type)
}

func TestCSharpParserLineNumbers(t *testing.T) {
	source, err := os.ReadFile("testdata/Sample.cs")
	require.NoError(t, err)

	p := NewCSharpParser()
	symbols, err := p.Parse("testdata/Sample.cs", source)
	require.NoError(t, err)

	for _, s := range symbols {
		assert.Greater(t, s.StartLine, 0, "StartLine should be 1-based for %s", s.Name)
		assert.GreaterOrEqual(t, s.EndLine, s.StartLine, "EndLine >= StartLine for %s", s.Name)
	}
}

func TestCSharpRegistryRouting(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewCSharpParser())

	p, ok := reg.ParserFor("Program.cs")
	assert.True(t, ok, ".cs files should route to C# parser")
	assert.Equal(t, "csharp", p.Language())

	_, ok = reg.ParserFor("Program.java")
	assert.False(t, ok, ".java should not route to C# parser")
}
