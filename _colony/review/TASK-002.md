# TASK-002: Configuration System — engram.json Loading, Defaults, Validation

**Priority:** P0
**Assigned:** bravo
**Milestone:** M1: MVP
**Dependencies:** TASK-001
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
Engram needs a configuration system that loads settings from an `engram.json` file in the project root, applies sensible defaults, and validates values. Every component (storage, MCP server, parser, embeddings) will read from this config. This must be in place early so other tasks can reference configuration values instead of hardcoding them.

## Specification
Create the `internal/config` package that:

1. **Defines the Config struct:**
```go
type Config struct {
    // Storage
    DatabasePath string `json:"database_path"` // default: ~/.engram/<repo-hash>/engram.db
    WALMode      bool   `json:"wal_mode"`      // default: true

    // Server
    Transport    string `json:"transport"`      // "stdio" (default) or "http"
    HTTPAddr     string `json:"http_addr"`      // default: ":3333"
    HTTPToken    string `json:"http_token"`     // bearer token for HTTP transport

    // Indexing
    Languages    []string `json:"languages"`    // default: ["go","python","typescript","rust","java","csharp"]
    IgnorePatterns []string `json:"ignore_patterns"` // default: ["vendor/","node_modules/",".git/","bin/","dist/"]
    MaxFileSize  int64  `json:"max_file_size"`  // default: 1MB (1048576)

    // Embeddings
    EmbeddingModel   string `json:"embedding_model"`    // default: "builtin" (ONNX MiniLM)
    OllamaEndpoint   string `json:"ollama_endpoint"`    // default: "http://localhost:11434"
    OllamaModel      string `json:"ollama_model"`       // default: "nomic-embed-text"
    EmbeddingBatchSize int  `json:"embedding_batch_size"` // default: 32

    // Derived (not in JSON)
    RepoRoot     string `json:"-"`              // detected from git
    RepoHash     string `json:"-"`              // SHA256 of repo root path
}
```

2. **Load precedence:** CLI flags > environment variables (`ENGRAM_*`) > engram.json > defaults

3. **Functions to implement:**
   - `Load(repoRoot string) (*Config, error)` — main entry point
   - `Default() *Config` — returns config with all defaults applied
   - `LoadFromFile(path string) (*Config, error)` — reads and parses engram.json
   - `(c *Config) Validate() error` — checks for invalid values
   - `RepoHash(repoRoot string) string` — SHA256 of the absolute repo root path, truncated to 12 hex chars
   - `DatabaseDir(repoRoot string) string` — returns `~/.engram/<repo-hash>/`

## Acceptance Criteria
- [ ] `config.Default()` returns a Config with all defaults populated correctly
- [ ] `config.Load(repoRoot)` loads from engram.json if it exists, applies defaults for missing fields
- [ ] `config.Load(repoRoot)` works correctly when no engram.json exists (all defaults)
- [ ] `config.Validate()` returns error for invalid transport value (not "stdio" or "http")
- [ ] `config.Validate()` returns error for negative MaxFileSize
- [ ] `config.RepoHash()` returns consistent 12-char hex string for the same path
- [ ] `config.DatabaseDir()` returns `~/.engram/<hash>/` with the correct hash
- [ ] Environment variables with ENGRAM_ prefix override JSON values

## Implementation Steps
1. Create `internal/config/config.go` with the Config struct and all methods
2. Create `internal/config/defaults.go` with the Default() function and default constants
3. Create `internal/config/config_test.go` with comprehensive tests:
   - Test Default() returns correct values
   - Test Load() with a temp engram.json
   - Test Load() with no config file (defaults only)
   - Test Validate() with valid config
   - Test Validate() with invalid transport
   - Test RepoHash() determinism
   - Test environment variable override
4. Run `go test ./internal/config/` — all tests pass

## Testing Requirements
- Unit test: Default() returns correct default values for all fields
- Unit test: Load() with JSON file overrides specified fields, keeps defaults for others
- Unit test: Load() without JSON file returns all defaults
- Unit test: Validate() accepts valid config
- Unit test: Validate() rejects invalid transport ("invalid", "")
- Unit test: Validate() rejects invalid MaxFileSize (-1)
- Unit test: RepoHash() same path → same hash, different paths → different hashes
- Unit test: Environment variable ENGRAM_DATABASE_PATH overrides default

## Files to Create/Modify
- `internal/config/config.go` — Config struct, Load, LoadFromFile, Validate methods
- `internal/config/defaults.go` — Default() function, default constants
- `internal/config/config_test.go` — all tests

## Notes
- Use `os.UserHomeDir()` for the `~` expansion in DatabaseDir.
- Use `crypto/sha256` for RepoHash.
- For env var loading, use a simple pattern: `ENGRAM_DATABASE_PATH` maps to `Config.DatabasePath`. Use uppercase with underscores.
- Do NOT use viper or any third-party config library. Keep it simple with `encoding/json` + manual env var reads.
- The config system should work on Linux, macOS, and Windows (path separators, home dir detection).

---
## Completion Notes
- **Completed by:** bravo-2
- **Date:** 2026-03-15 15:02:11
- **Branch:** task/002
