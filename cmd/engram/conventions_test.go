package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConventionsCmdRegistered(t *testing.T) {
	root := newRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "conventions" {
			found = true
			// Check subcommands
			subCmds := cmd.Commands()
			subNames := make([]string, len(subCmds))
			for i, sc := range subCmds {
				subNames[i] = sc.Name()
			}
			assert.Contains(t, subNames, "list")
			assert.Contains(t, subNames, "add")
			assert.Contains(t, subNames, "remove")
			break
		}
	}
	assert.True(t, found, "conventions command should be registered")
}
