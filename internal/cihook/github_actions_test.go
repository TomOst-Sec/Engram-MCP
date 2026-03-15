package cihook

import (
	"strings"
	"testing"
)

func TestGHAParserDetectsErrorAnnotations(t *testing.T) {
	input := `2026-03-15T10:00:00.000Z ##[error]Process completed with exit code 1.
2026-03-15T10:00:01.000Z ##[error]src/auth.go:15: undefined: validateToken`
	events, err := ParseGitHubActions(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseGitHubActions: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	for _, e := range events {
		if e.Type != "failure" {
			t.Errorf("event type = %q, want 'failure'", e.Type)
		}
	}
}

func TestGHAParserDetectsTestFailures(t *testing.T) {
	input := `=== RUN   TestLogin
--- FAIL: TestLogin (0.50s)
    auth_test.go:42: expected 200, got 401
FAIL	github.com/example/auth	0.55s`
	events, err := ParseGitHubActions(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseGitHubActions: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected test failure events")
	}
	foundTest := false
	foundPkg := false
	for _, e := range events {
		if strings.Contains(e.Summary, "Test failure") {
			foundTest = true
		}
		if strings.Contains(e.Summary, "Package test failure") {
			foundPkg = true
		}
	}
	if !foundTest {
		t.Error("expected test failure event from --- FAIL line")
	}
	if !foundPkg {
		t.Error("expected package failure event from FAIL line")
	}
}

func TestGHAParserDetectsWarnings(t *testing.T) {
	input := `##[warning]The following actions uses node12 which is deprecated`
	events, err := ParseGitHubActions(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseGitHubActions: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != "warning" {
		t.Errorf("type = %q, want 'warning'", events[0].Type)
	}
}

func TestGHAParserIgnoresNormalOutput(t *testing.T) {
	input := `=== RUN   TestLogin
--- PASS: TestLogin (0.01s)
PASS
ok  	github.com/example/auth	0.05s`
	events, err := ParseGitHubActions(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseGitHubActions: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events for passing tests, got %d", len(events))
	}
}
