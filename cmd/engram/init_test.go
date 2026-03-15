package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCmdRegistered(t *testing.T) {
	root := newRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "init" {
			found = true
			break
		}
	}
	assert.True(t, found, "init command should be registered")
}

func TestInitCmdHelp(t *testing.T) {
	root := newRootCmd()
	for _, cmd := range root.Commands() {
		if cmd.Name() == "init" {
			assert.Contains(t, cmd.Short, "Initialize")
			assert.Contains(t, cmd.Long, "connection instructions")
			return
		}
	}
	t.Fatal("init command not found")
}
