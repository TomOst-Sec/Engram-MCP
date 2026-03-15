package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchCommandRegistered(t *testing.T) {
	root := newRootCmd()
	cmd, _, err := root.Find([]string{"search"})
	require.NoError(t, err)
	assert.Equal(t, "search", cmd.Name())
}

func TestSearchHelpOutput(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"search", "--help"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Search")
	assert.Contains(t, output, "FTS5")
	assert.Contains(t, output, "--language")
	assert.Contains(t, output, "--type")
	assert.Contains(t, output, "--limit")
}

func TestSearchNoArgsReturnsError(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"search"})

	err := root.Execute()
	assert.Error(t, err)
}
