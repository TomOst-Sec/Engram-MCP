package main

import (
	"errors"
	"fmt"
	"os"

	mcpmcp "github.com/TomOst-Sec/colony-project/internal/mcp"
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

	// Create MCP server
	server := mcpmcp.New("engram", version)
	mcpmcp.RegisterBuiltinTools(server)

	fmt.Fprintf(os.Stderr, "Engram MCP server starting (version %s, transport: %s)\n", version, transport)

	// Start stdio transport (blocks until stdin closes or signal received)
	if err := server.ServeStdio(); err != nil {
		fmt.Fprintln(os.Stderr, "Shutting down...")
		server.Shutdown()
		return err
	}

	fmt.Fprintln(os.Stderr, "Shutting down...")
	return server.Shutdown()
}
