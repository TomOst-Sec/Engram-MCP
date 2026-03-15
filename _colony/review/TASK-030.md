# TASK-030: `npx engram init` Bootstrap — npm Wrapper Package

**Priority:** P1
**Assigned:** bravo
**Milestone:** M2: Core Features
**Dependencies:** none
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
Feature 12 from GOALS.md. The `npx engram init` command is the primary installation flow for most developers. It's an npm package that detects the platform/arch, downloads the correct Go binary from GitHub Releases, runs the initial index, and prints connection instructions. The entire flow from `npx engram init` to "your AI tool is connected" must take <60 seconds on a 50K LOC repo.

## Specification

### npm Package Structure

Create a `npm/` directory at the project root:

```
npm/
├── package.json
├── bin/
│   └── engram.js       — CLI entry point
├── lib/
│   ├── install.js      — Binary download logic
│   ├── platform.js     — Platform/arch detection
│   └── run.js          — Binary execution wrapper
└── README.md           — npm package README (brief, links to main README)
```

### package.json

```json
{
  "name": "engram",
  "version": "0.1.0",
  "description": "Persistent, intelligent memory for AI coding agents — MCP server",
  "bin": {
    "engram": "./bin/engram.js"
  },
  "scripts": {
    "postinstall": "node lib/install.js"
  },
  "keywords": ["mcp", "ai", "coding", "memory", "context", "cursor", "claude"],
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/TomOst-Sec/colony-project"
  }
}
```

### bin/engram.js

```javascript
#!/usr/bin/env node
const { spawn } = require('child_process');
const { getBinaryPath } = require('../lib/platform');

const binary = getBinaryPath();
const child = spawn(binary, process.argv.slice(2), { stdio: 'inherit' });
child.on('exit', (code) => process.exit(code || 0));
```

### lib/platform.js

```javascript
function getPlatform() {
  const platform = process.platform; // 'darwin', 'linux', 'win32'
  const arch = process.arch;         // 'x64', 'arm64'

  const platformMap = {
    'darwin-x64': 'darwin-amd64',
    'darwin-arm64': 'darwin-arm64',
    'linux-x64': 'linux-amd64',
    'linux-arm64': 'linux-arm64',
    'win32-x64': 'windows-amd64',
  };

  const key = `${platform}-${arch}`;
  return platformMap[key] || null;
}

function getBinaryName() {
  return process.platform === 'win32' ? 'engram.exe' : 'engram';
}

function getBinaryPath() {
  return path.join(__dirname, '..', 'bin', getBinaryName());
}
```

### lib/install.js (postinstall)

```javascript
const https = require('https');
const fs = require('fs');
const { getPlatform, getBinaryName } = require('./platform');

async function install() {
  const platform = getPlatform();
  if (!platform) {
    console.error('Unsupported platform:', process.platform, process.arch);
    process.exit(1);
  }

  const version = require('../package.json').version;
  const binaryName = getBinaryName();
  const url = `https://github.com/TomOst-Sec/colony-project/releases/download/v${version}/engram-${platform}`;

  console.log(`Downloading Engram ${version} for ${platform}...`);
  // Download binary to bin/ directory
  // Set executable permissions (chmod +x on unix)
  // Verify checksum if .sha256 file exists
}
```

### `npx engram init` Flow

When a user runs `npx engram init`, the engram.js wrapper calls the Go binary with `init` args. This requires adding an `init` subcommand to the Go CLI:

Add to `cmd/engram/init.go`:

```go
var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize Engram for this project",
    Long:  "Index the repository and print connection instructions for AI coding tools.",
    RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
    // 1. Run index
    // 2. Print connection instructions for Claude Code, Cursor, Codex
    // 3. Print "Ready! Your AI coding tool now has persistent memory."
}
```

**Output of `engram init`:**
```
Engram v0.1.0 — Persistent memory for AI coding agents

Indexing /path/to/project...
  342 files indexed, 2,847 symbols extracted
  Database: ~/.engram/a1b2c3/engram.db (4.2 MB)

Connect your AI tool:

  Claude Code:  claude mcp add engram -- engram serve
  Cursor:       Add to .cursor/mcp.json (see docs/cursor.md)
  Codex CLI:    codex --mcp-server "engram serve"

Ready! Your AI coding tool now has persistent memory.
```

## Acceptance Criteria
- [ ] `npm/` directory contains a valid npm package
- [ ] `npm/package.json` has correct name, version, bin, and postinstall script
- [ ] `npm/lib/platform.js` correctly detects all 5 platform/arch combinations
- [ ] `npm/lib/install.js` downloads the correct binary for the platform
- [ ] `npm/bin/engram.js` delegates to the Go binary
- [ ] `cmd/engram/init.go` — `engram init` command runs index then prints connection instructions
- [ ] `engram init` output includes instructions for Claude Code, Cursor, and Codex
- [ ] All Go tests pass

## Implementation Steps
1. Create `npm/` directory structure
2. Create `npm/package.json`
3. Create `npm/lib/platform.js` — platform/arch detection
4. Create `npm/lib/install.js` — binary download (postinstall)
5. Create `npm/bin/engram.js` — CLI wrapper
6. Create `npm/README.md` — brief npm package description
7. Create `cmd/engram/init.go` — init subcommand
8. Register init command in `cmd/engram/main.go`
9. Create `cmd/engram/init_test.go`:
   - Test: init command is registered
   - Test: --help output is correct
10. Run all Go tests

## Testing Requirements
- Unit test: init command registered on root command
- Manual test: `node npm/lib/platform.js` outputs correct platform
- Manual test: `engram init` prints connection instructions

## Files to Create/Modify
- `npm/package.json` — npm package config
- `npm/bin/engram.js` — CLI entry point
- `npm/lib/platform.js` — platform detection
- `npm/lib/install.js` — binary download
- `npm/lib/run.js` — binary execution
- `npm/README.md` — npm package README
- `cmd/engram/init.go` — init subcommand
- `cmd/engram/init_test.go` — init tests
- `cmd/engram/main.go` — register init command (one line)

## Notes
- The npm package does NOT need to actually work end-to-end (no GitHub Release exists yet). Just build the correct structure so it's ready when releases are set up.
- The `postinstall` download script should fail gracefully if the release URL doesn't exist — print a message saying "Binary not found, build from source: go install ..."
- Use `https` module directly (no dependencies). The npm package must have zero npm dependencies.
- The Go `init` subcommand reuses the existing indexing logic from index.go. Import and call the same functions.
- Do NOT add the npm/ directory to .gitignore. It should be committed.

---
## Completion Notes
- **Completed by:** bravo-1
- **Date:** 2026-03-15 17:25:45
- **Branch:** task/030
