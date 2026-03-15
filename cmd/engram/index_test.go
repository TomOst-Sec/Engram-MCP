package main

import (
	"strings"
	"testing"
)

func TestIndexCmdRegistered(t *testing.T) {
	root := newRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "index" {
			found = true
			break
		}
	}
	if !found {
		t.Error("index command not registered on root command")
	}
}

func TestIndexCmdHelp(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"index", "--help"})

	out := new(strings.Builder)
	root.SetOut(out)

	err := root.Execute()
	if err != nil {
		t.Fatalf("execute help: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "index") {
		t.Error("help output should contain 'index'")
	}
	if !strings.Contains(output, "--force") {
		t.Error("help output should contain '--force'")
	}
	if !strings.Contains(output, "--verbose") {
		t.Error("help output should contain '--verbose'")
	}
}
