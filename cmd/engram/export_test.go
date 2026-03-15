package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExportCmdRegistered(t *testing.T) {
	root := newRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "export" {
			found = true
			break
		}
	}
	assert.True(t, found, "export command should be registered")
}

func TestExportCmdHelp(t *testing.T) {
	root := newRootCmd()
	for _, cmd := range root.Commands() {
		if cmd.Name() == "export" {
			assert.Contains(t, cmd.Short, "Export")
			assert.Contains(t, cmd.Long, "JSON")
			return
		}
	}
	t.Fatal("export command not found")
}
