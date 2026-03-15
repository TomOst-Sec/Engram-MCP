package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusCommandRegistered(t *testing.T) {
	root := newRootCmd()
	cmd, _, err := root.Find([]string{"status"})
	require.NoError(t, err)
	assert.Equal(t, "status", cmd.Name())
}

func TestStatusHelpOutput(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"status", "--help"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "status")
	assert.Contains(t, output, "index")
}
