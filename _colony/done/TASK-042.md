# TASK-042: GoReleaser + Package Distribution

**Priority:** P2
**Assigned:** alpha
**Milestone:** M3: Polish & Growth
**Dependencies:** TASK-001
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
GOALS.md specifies distribution via GoReleaser for cross-compilation, Homebrew tap, AUR package, Scoop bucket, and GitHub Releases with .deb/.rpm. This task sets up the GoReleaser configuration and packaging infrastructure so releases can be built with a single `goreleaser release` command.

## Specification

### GoReleaser Config: `.goreleaser.yml`

```yaml
version: 2
project_name: engram

before:
  hooks:
    - go mod tidy

builds:
  - main: ./cmd/engram
    binary: engram
    env:
      - CGO_ENABLED=1
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    flags:
      - -tags=sqlite_fts5
    ldflags:
      - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}}

archives:
  - format: tar.gz
    name_template: "engram-{{ .Os }}-{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^atlas:"
      - "^audit:"

nfpms:
  - id: engram
    package_name: engram
    vendor: TomOst-Sec
    homepage: https://github.com/TomOst-Sec/colony-project
    description: "Persistent memory MCP server for AI coding agents"
    license: MIT
    formats:
      - deb
      - rpm

brews:
  - repository:
      owner: TomOst-Sec
      name: homebrew-tap
    homepage: https://github.com/TomOst-Sec/colony-project
    description: "Persistent memory MCP server for AI coding agents"
    license: MIT
```

### Version Injection

Update `cmd/engram/main.go`:
```go
var (
    version = "dev"
    commit  = "none"
)
```

GoReleaser injects the actual version via ldflags at build time.

### Makefile Updates

```makefile
.PHONY: release-dry
release-dry:
	goreleaser release --snapshot --clean --skip=publish

.PHONY: release
release:
	goreleaser release --clean
```

### GitHub Actions: `.github/workflows/release.yml`

```yaml
name: Release
on:
  push:
    tags:
      - "v*"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Acceptance Criteria
- [ ] `.goreleaser.yml` exists with correct build config
- [ ] GoReleaser config compiles for linux/darwin x amd64/arm64
- [ ] Version is injected via ldflags
- [ ] `goreleaser release --snapshot --clean` succeeds (dry run)
- [ ] nfpms config generates .deb and .rpm descriptions
- [ ] Homebrew tap config is correct
- [ ] GitHub Actions release workflow exists
- [ ] Makefile has release-dry and release targets
- [ ] All tests still pass

## Implementation Steps
1. Create `.goreleaser.yml`
2. Update `cmd/engram/main.go` — add version/commit vars, use in root command Version field
3. Create `.github/workflows/release.yml` — release automation
4. Update `Makefile` — add release-dry and release targets
5. Test: `goreleaser check` (if goreleaser installed) or validate YAML structure
6. Run all tests

## Files to Create/Modify
- `.goreleaser.yml` — GoReleaser config (create new)
- `.github/workflows/release.yml` — release CI (create new)
- `cmd/engram/main.go` — add version/commit vars
- `Makefile` — add release targets

## Notes
- CGO is required for SQLite (mattn/go-sqlite3). This means cross-compilation needs cross-compilers (zig cc or platform-specific gcc). GoReleaser handles this with the right env vars.
- Windows build may not work initially due to CGO cross-compilation complexity. That's OK — target linux + darwin first.
- The Homebrew tap repo (TomOst-Sec/homebrew-tap) doesn't need to exist for the config to be correct.
- AUR and Scoop packages are lower priority — add comments in .goreleaser.yml noting they'll be added later.
- Do NOT run `goreleaser release` — just create the config and dry-run.

---
## Completion Notes
- **Completed by:** alpha-3
- **Date:** 2026-03-15 18:03:23
- **Branch:** task/042
