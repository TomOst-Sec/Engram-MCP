# TASK-043: Docker Image — Alpine-Based Container for CI/CD

**Priority:** P2
**Assigned:** bravo
**Milestone:** M3: Polish & Growth
**Dependencies:** TASK-001
**Status:** review
**Created:** 2026-03-15
**Author:** atlas

## Context
GOALS.md specifies a Docker image (alpine-based, ~30MB) for CI/CD integration. This allows teams to run Engram in pipelines without installing the Go binary. The image runs `engram serve` by default and can be used in GitHub Actions, GitLab CI, and other CI systems.

## Specification

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -tags sqlite_fts5 \
    -ldflags="-s -w" \
    -o /engram ./cmd/engram

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache sqlite-libs git ca-certificates

COPY --from=builder /engram /usr/local/bin/engram

# Create non-root user
RUN adduser -D -h /home/engram engram
USER engram
WORKDIR /workspace

# Default: start MCP server
ENTRYPOINT ["engram"]
CMD ["serve"]
```

### .dockerignore

```
.git
_colony
.worktrees
*.db
node_modules
vendor
```

### Docker Compose (optional, for team setups)

```yaml
# docker-compose.yml
version: "3.8"
services:
  engram:
    build: .
    volumes:
      - .:/workspace:ro
      - engram-data:/home/engram/.engram
    ports:
      - "3333:3333"
    command: ["serve", "--transport", "http", "--http-addr", "0.0.0.0:3333"]

volumes:
  engram-data:
```

### Usage Documentation

Add `docs/docker.md`:
```markdown
# Engram Docker

## Quick Start
docker build -t engram .
docker run -v $(pwd):/workspace engram index
docker run -v $(pwd):/workspace engram serve

## CI/CD Example (GitHub Actions)
- uses: docker://ghcr.io/TomOst-Sec/engram:latest
  with:
    args: index

## Team Setup (HTTP mode)
docker compose up -d
# Connect AI tools to http://localhost:3333
```

### Makefile Integration

```makefile
.PHONY: docker-build
docker-build:
	docker build -t engram:latest .

.PHONY: docker-test
docker-test: docker-build
	docker run --rm -v $(PWD):/workspace engram:latest status
```

## Acceptance Criteria
- [ ] `Dockerfile` exists with multi-stage build
- [ ] `docker build .` succeeds and produces a working image
- [ ] Image size is <50MB (target: ~30MB)
- [ ] `docker run engram status` works
- [ ] `docker run -v $(pwd):/workspace engram index` indexes the mounted repo
- [ ] `.dockerignore` excludes unnecessary files
- [ ] `docker-compose.yml` exists for team HTTP setup
- [ ] `docs/docker.md` has usage instructions
- [ ] Makefile has docker-build and docker-test targets
- [ ] Non-root user in container

## Implementation Steps
1. Create `Dockerfile` — multi-stage alpine build
2. Create `.dockerignore`
3. Create `docker-compose.yml` — team HTTP setup
4. Create `docs/docker.md` — usage documentation
5. Update `Makefile` — docker targets
6. Test: `docker build .` (if Docker available)
7. Run all Go tests (no Docker dependency)

## Files to Create/Modify
- `Dockerfile` — container build (create new)
- `.dockerignore` — build exclusions (create new)
- `docker-compose.yml` — team setup (create new)
- `docs/docker.md` — Docker documentation (create new)
- `Makefile` — add docker targets

## Notes
- Alpine needs `musl-dev` for CGO compilation and `sqlite-dev` for the SQLite C library.
- The runtime image needs `sqlite-libs` for the dynamic SQLite library and `git` for git history analysis.
- Mount the repo as read-only (`:ro`) for serve mode. Index mode needs write access to `~/.engram/`.
- The image should work with both stdio and HTTP transport. Stdio is default (for pipe-based MCP), HTTP is for team/remote setups.
- Don't publish the image to any registry — just create the Dockerfile and docs. Publishing is a future CI task.

---
## Completion Notes
- **Completed by:** bravo-1
- **Date:** 2026-03-15 18:13:38
- **Branch:** task/043
