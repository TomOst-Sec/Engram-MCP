# Engram Docker

## Quick Start

Build the image:
```bash
docker build -t engram .
```

Index a repository:
```bash
docker run -v $(pwd):/workspace engram index
```

Start the MCP server (stdio mode):
```bash
docker run -v $(pwd):/workspace engram serve
```

## CI/CD Example (GitHub Actions)

```yaml
jobs:
  index:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Index with Engram
        uses: docker://ghcr.io/TomOst-Sec/engram:latest
        with:
          args: index
```

## Team Setup (HTTP mode)

Use Docker Compose for team setups where multiple AI tools connect via HTTP:

```bash
docker compose up -d
# Connect AI tools to http://localhost:3333
```

This starts Engram in HTTP transport mode, accessible at `http://localhost:3333`.

## Environment Variables

- Mount your repository to `/workspace` (read-only is fine for `serve`)
- Engram data is stored in `/home/engram/.engram/` inside the container
- Use a named volume for persistence across container restarts

## Image Details

- Base: Alpine 3.19
- Size: ~30MB
- Includes: SQLite, Git, CA certificates
- Runs as non-root user `engram`
