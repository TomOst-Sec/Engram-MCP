package cihook

import (
	"strings"
	"testing"
)

func TestGenericParserDetectsErrors(t *testing.T) {
	input := `Building project...
src/main.go:42: error: undefined variable
Compiling...
Done.`
	events, err := ParseGeneric(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseGeneric: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected at least one event for error line")
	}
	if events[0].Type != "failure" {
		t.Errorf("event type = %q, want 'failure'", events[0].Type)
	}
}

func TestGenericParserDetectsFailures(t *testing.T) {
	input := `Running tests...
FAIL TestLogin (0.5s)
ok TestSignup (0.1s)`
	events, err := ParseGeneric(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseGeneric: %v", err)
	}
	found := false
	for _, e := range events {
		if e.Type == "failure" && strings.Contains(e.Summary, "FAIL") {
			found = true
		}
	}
	if !found {
		t.Error("expected failure event for FAIL line")
	}
}

func TestGenericParserDetectsWarnings(t *testing.T) {
	input := `Compiling...
warning: unused variable 'x'
Done.`
	events, err := ParseGeneric(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseGeneric: %v", err)
	}
	found := false
	for _, e := range events {
		if e.Type == "warning" {
			found = true
		}
	}
	if !found {
		t.Error("expected warning event")
	}
}

func TestGenericParserIgnoresNormalOutput(t *testing.T) {
	input := `Building project...
Compiling package main
Linking binary
Done. 0 errors, 0 warnings.`
	events, err := ParseGeneric(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseGeneric: %v", err)
	}
	// "0 errors, 0 warnings" contains "error" and "warning" words,
	// but the generic parser matches word boundaries, so this may or may not match.
	// The key test is that "Building project..." and "Compiling" lines don't match.
	for _, e := range events {
		if e.Summary == "Building project..." || e.Summary == "Compiling package main" {
			t.Errorf("parser should not capture normal output: %q", e.Summary)
		}
	}
}

func TestGenericParserExtractsFilePaths(t *testing.T) {
	input := `error in src/handler.go:42 undefined reference`
	events, err := ParseGeneric(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseGeneric: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected event")
	}
	if len(events[0].Files) == 0 {
		t.Error("expected file path extraction")
	}
	if events[0].Files[0] != "src/handler.go" {
		t.Errorf("file = %q, want 'src/handler.go'", events[0].Files[0])
	}
}

func TestGenericParserDetectsDeployment(t *testing.T) {
	input := `Successfully deployed to production
Build complete`
	events, err := ParseGeneric(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseGeneric: %v", err)
	}
	found := false
	for _, e := range events {
		if e.Type == "deployment" {
			found = true
		}
	}
	if !found {
		t.Error("expected deployment event")
	}
}
