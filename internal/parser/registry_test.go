package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryRegisterAndRoute(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewGoParser())
	reg.Register(NewPythonParser())

	p, ok := reg.ParserFor("main.go")
	assert.True(t, ok)
	assert.Equal(t, "go", p.Language())

	p, ok = reg.ParserFor("script.py")
	assert.True(t, ok)
	assert.Equal(t, "python", p.Language())

	p, ok = reg.ParserFor("stub.pyi")
	assert.True(t, ok)
	assert.Equal(t, "python", p.Language())
}

func TestRegistryUnknownExtension(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewGoParser())

	_, ok := reg.ParserFor("main.rs")
	assert.False(t, ok)
}

func TestRegistryParseFile(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewGoParser())

	source := []byte(`package main
func Hello() {}
`)
	symbols, err := reg.ParseFile("hello.go", source)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(symbols), 1)
}

func TestRegistryParseFileUnknownExt(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.ParseFile("main.rs", []byte("fn main() {}"))
	assert.Error(t, err)
}

func TestRegistrySupportedLanguages(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewGoParser())
	reg.Register(NewPythonParser())

	langs := reg.SupportedLanguages()
	assert.Equal(t, []string{"go", "python"}, langs)
}
