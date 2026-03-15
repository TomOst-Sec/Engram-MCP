package main

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TomOst-Sec/colony-project/internal/embeddings"
	mcpmcp "github.com/TomOst-Sec/colony-project/internal/mcp"
	"github.com/TomOst-Sec/colony-project/internal/storage"
	"github.com/TomOst-Sec/colony-project/internal/tools/architecture"
	"github.com/TomOst-Sec/colony-project/internal/tools/memory"
	searchtools "github.com/TomOst-Sec/colony-project/internal/tools/search"
	"github.com/spf13/cobra"
)

var (
	transport string
	logLevel  string
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Engram MCP server",
		Long:  "Start the Engram MCP server using stdio transport. AI coding tools connect to this server to access codebase intelligence.",
		RunE:  runServe,
	}
	cmd.Flags().StringVar(&transport, "transport", "stdio", "Transport type: stdio (default) or http")
	cmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level: debug, info, warn, error")
	return cmd
}

func runServe(cmd *cobra.Command, args []string) error {
	// Validate transport
	switch transport {
	case "stdio":
		// ok
	case "http":
		fmt.Fprintln(os.Stderr, "HTTP transport not yet implemented")
		return errors.New("HTTP transport not yet implemented")
	default:
		return fmt.Errorf("unknown transport %q: must be stdio or http", transport)
	}

	// Validate log level
	switch logLevel {
	case "debug", "info", "warn", "error":
		// ok
	default:
		return fmt.Errorf("unknown log level %q: must be debug, info, warn, or error", logLevel)
	}

	// 1. Detect repo root
	repoRoot, err := detectRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not detect repo root (%v), using cwd\n", err)
		repoRoot, _ = os.Getwd()
	}

	// 2. Open storage
	dbDir := databaseDir(repoRoot)
	dbPath := filepath.Join(dbDir, "engram.db")
	store, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database at %s: %w", dbPath, err)
	}
	defer store.Close()

	// 3. Create embedder (graceful degradation if unavailable)
	var emb *embeddings.Embedder
	emb, err = embeddings.New("") // empty = no model available yet
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: embeddings unavailable (%v), using FTS5-only search\n", err)
		emb = nil
	}

	// 4. Create MCP server and register tools
	server := mcpmcp.New("engram", version)
	mcpmcp.RegisterBuiltinTools(server)

	// Register all tools
	goMod := goModulePath(repoRoot)
	searchtools.RegisterTools(server, store, emb, repoRoot)
	memory.RegisterTools(server, store)
	architecture.RegisterTools(server, store, repoRoot, goMod)

	fmt.Fprintf(os.Stderr, "Engram MCP server starting (version %s, transport: %s, repo: %s)\n", version, transport, repoRoot)

	// Start stdio transport (blocks until stdin closes or signal received)
	if err := server.ServeStdio(); err != nil {
		fmt.Fprintln(os.Stderr, "Shutting down...")
		server.Shutdown()
		return err
	}

	fmt.Fprintln(os.Stderr, "Shutting down...")
	return server.Shutdown()
}

// detectRepoRoot walks up from the current directory looking for .git/.
func detectRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting cwd: %w", err)
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .git directory found")
		}
		dir = parent
	}
}

// goModulePath reads the go.mod file in the given directory and returns
// the module path. Returns empty string if go.mod doesn't exist or can't be parsed.
func goModulePath(repoRoot string) string {
	goModFile := filepath.Join(repoRoot, "go.mod")
	f, err := os.Open(goModFile)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}
	return ""
}

// databaseDir returns the directory for storing the Engram database.
// Uses ~/.engram/<repo-hash>/ to isolate per-project databases.
func databaseDir(repoRoot string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}

	hash := sha256.Sum256([]byte(repoRoot))
	hashStr := fmt.Sprintf("%x", hash[:6])

	return filepath.Join(home, ".engram", hashStr)
}
