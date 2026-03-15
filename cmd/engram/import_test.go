package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportCmdRegistered(t *testing.T) {
	root := newRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "import" {
			found = true
			break
		}
	}
	assert.True(t, found, "import command should be registered")
}

func TestImportCmdHelp(t *testing.T) {
	root := newRootCmd()
	for _, cmd := range root.Commands() {
		if cmd.Name() == "import" {
			assert.Contains(t, cmd.Short, "Import")
			assert.Contains(t, cmd.Long, "JSON")
			return
		}
	}
	t.Fatal("import command not found")
}
