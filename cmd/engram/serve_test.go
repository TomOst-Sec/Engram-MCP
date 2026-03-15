package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeCommandRegistered(t *testing.T) {
	root := newRootCmd()
	cmd, _, err := root.Find([]string{"serve"})
	require.NoError(t, err)
	assert.Equal(t, "serve", cmd.Name())
}

func TestServeHelpOutput(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"serve", "--help"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "MCP server")
	assert.Contains(t, output, "stdio")
}

func TestServeInvalidTransport(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"serve", "--transport", "invalid"})

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown transport")
}

func TestServeHTTPTransportNotImplemented(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"serve", "--transport", "http"})

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestServeInvalidLogLevel(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"serve", "--log-level", "trace"})

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown log level")
}

func TestVersionIsNonEmpty(t *testing.T) {
	assert.NotEmpty(t, version)
}
