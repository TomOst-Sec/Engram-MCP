# TASK-044: CI/CD Memory Hook — `engram ci-hook`

**Priority:** P2
**Assigned:** alpha
**Milestone:** M4: Ecosystem
**Dependencies:** TASK-003, TASK-007
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 20 from GOALS.md. `engram ci-hook` reads CI/CD output (test failures, build errors, deployment logs) and stores them as memories. When a developer starts a new session, the AI knows "the last 3 builds failed because of a timeout in the payment service." Supports GitHub Actions, GitLab CI, and generic stdin parsing.

## Specification

### CLI Command: `cmd/engram/ci_hook.go`

```go
var ciHookCmd = &cobra.Command{
    Use:   "ci-hook",
    Short: "Store CI/CD output as memories",
    Long:  "Parse CI/CD build output and store failures, warnings, and deployment info as Engram memories.",
    RunE:  runCIHook,
}
```

**Flags:**
- `--source` / `-s` — CI system: "github-actions", "gitlab-ci", "generic" (default: "generic")
- `--type` / `-t` — Memory type: "bugfix", "learning", "decision" (default: "learning")
- `--run-id` — CI run identifier (optional, for dedup)
- `--file` / `-f` — Read from file instead of stdin

**Behavior:**
1. Read CI output from stdin (pipe) or file
2. Parse based on source format
3. Extract: failures, errors, warnings, test results
4. Store each finding as a memory with type and tags
5. Print summary: "Stored 3 CI memories (2 failures, 1 warning)"

### Parser Package: `internal/cihook/`

```go
type CIEvent struct {
    Type    string   // "failure", "warning", "success", "deployment"
    Summary string   // one-line description
    Details string   // full context
    Tags    []string // auto-generated tags
    Files   []string // related files (if detectable)
}

// ParseGitHubActions parses GitHub Actions log output.
func ParseGitHubActions(input io.Reader) ([]CIEvent, error)

// ParseGitLabCI parses GitLab CI log output.
func ParseGitLabCI(input io.Reader) ([]CIEvent, error)

// ParseGeneric parses generic build output (look for ERROR, FAIL, WARNING patterns).
func ParseGeneric(input io.Reader) ([]CIEvent, error)
```

### GitHub Actions Parser
Detect patterns:
- `##[error]` prefixed lines → failure
- `FAIL` in test output → test failure with package name
- `Error:` lines → build error
- `warning:` lines → warning
- Exit code non-zero → overall failure

### Generic Parser
Regex-based pattern matching:
- Lines matching `(?i)(error|fail|fatal)` → failure
- Lines matching `(?i)(warn|warning)` → warning
- Lines matching `(?i)(deploy|deployed|released)` → deployment

### Usage Examples

```bash
# GitHub Actions — pipe workflow log
gh run view 12345 --log | engram ci-hook --source github-actions

# GitLab CI — pipe job output
gitlab-ci-job-output | engram ci-hook --source gitlab-ci

# Generic — pipe any build output
make test 2>&1 | engram ci-hook

# From file
engram ci-hook --file build-output.log
```

## Acceptance Criteria
- [ ] `engram ci-hook` reads from stdin and stores memories
- [ ] `engram ci-hook --file build.log` reads from file
- [ ] GitHub Actions parser detects ##[error] and FAIL lines
- [ ] Generic parser detects ERROR/FAIL/WARNING patterns
- [ ] Each finding stored as a memory with appropriate type and tags
- [ ] `--run-id` prevents duplicate memories from same CI run
- [ ] Summary output shows count of stored memories by type
- [ ] All tests pass

## Implementation Steps
1. Create `internal/cihook/parser.go` — CIEvent struct, ParseGeneric
2. Create `internal/cihook/github_actions.go` — GitHub Actions parser
3. Create `internal/cihook/gitlab_ci.go` — GitLab CI parser
4. Create `cmd/engram/ci_hook.go` — ci-hook subcommand
5. Register in `cmd/engram/main.go`
6. Create `internal/cihook/parser_test.go`:
   - Test: Generic parser extracts ERROR lines
   - Test: Generic parser extracts FAIL lines
   - Test: Generic parser ignores normal output
7. Create `internal/cihook/github_actions_test.go`:
   - Test: Parses ##[error] lines
   - Test: Parses test FAIL output
8. Create `cmd/engram/ci_hook_test.go`:
   - Test: ci-hook command registered
   - Test: --source flag accepts valid values
9. Run all tests

## Files to Create/Modify
- `internal/cihook/parser.go` — base parser types + generic parser
- `internal/cihook/github_actions.go` — GitHub Actions parser
- `internal/cihook/gitlab_ci.go` — GitLab CI parser
- `internal/cihook/parser_test.go` — parser tests
- `internal/cihook/github_actions_test.go` — GHA parser tests
- `cmd/engram/ci_hook.go` — CLI command
- `cmd/engram/ci_hook_test.go` — CLI tests
- `cmd/engram/main.go` — register command

## Notes
- CI output can be very large (thousands of lines). Parse line-by-line with `bufio.Scanner`, don't load all into memory.
- The `--run-id` dedup: before storing, check if a memory with tag `ci-run:<run-id>` exists. If yes, skip.
- For GitHub Actions, the `##[error]` annotation format is documented at https://docs.github.com/en/actions/reference/workflow-commands
- The generic parser should be conservative — only extract lines that clearly indicate problems. False positives (storing normal log lines as "errors") would pollute the memory.
- Tags should include: `ci`, source name, and any detected package/module names.

---
## Completion Notes
- **Completed by:** alpha-1
- **Date:** 2026-03-15 18:01:03
- **Branch:** task/044
