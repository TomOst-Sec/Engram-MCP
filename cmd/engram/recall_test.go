package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecallCommandRegistered(t *testing.T) {
	root := newRootCmd()
	cmd, _, err := root.Find([]string{"recall"})
	require.NoError(t, err)
	assert.Equal(t, "recall", cmd.Name())
}

func TestRecallHelpOutput(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"recall", "--help"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "memories")
	assert.Contains(t, output, "--type")
	assert.Contains(t, output, "--limit")
	assert.Contains(t, output, "--since")
}

func TestRecallNoArgsReturnsError(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"recall"})

	err := root.Execute()
	assert.Error(t, err)
}
