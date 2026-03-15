package main

import (
	"strings"
	"testing"
)

func TestCIHookCmdRegistered(t *testing.T) {
	root := newRootCmd()
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "ci-hook" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ci-hook command not registered on root command")
	}
}

func TestCIHookCmdHelp(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"ci-hook", "--help"})

	out := new(strings.Builder)
	root.SetOut(out)

	err := root.Execute()
	if err != nil {
		t.Fatalf("execute help: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "--source") {
		t.Error("help output should contain '--source'")
	}
	if !strings.Contains(output, "--file") {
		t.Error("help output should contain '--file'")
	}
	if !strings.Contains(output, "--run-id") {
		t.Error("help output should contain '--run-id'")
	}
}
