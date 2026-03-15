package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

// HTTPConfig holds HTTP server configuration.
type HTTPConfig struct {
	Addr       string // listen address, e.g. ":3333"
	Token      string // bearer token for auth (empty = no auth)
	CORSOrigin string // CORS allowed origin (default "*")
}

// ServeHTTP starts the HTTP/SSE MCP transport with optional auth and CORS.
func (s *Server) ServeHTTP(cfg HTTPConfig) error {
	if cfg.CORSOrigin == "" {
		cfg.CORSOrigin = "*"
	}

	// Create SSE server from mcp-go (implements http.Handler)
	sseServer := mcpserver.NewSSEServer(s.mcpServer,
		mcpserver.WithBaseURL(fmt.Sprintf("http://localhost%s", cfg.Addr)),
		mcpserver.WithStaticBasePath("/mcp"),
	)

	// Build HTTP mux
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.healthHandler)

	// Mount SSE server with auth and CORS middleware
	wrapped := corsMiddleware(cfg.CORSOrigin,
		authMiddleware(cfg.Token, sseServer))
	mux.Handle("/mcp/", wrapped)

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      corsMiddleware(cfg.CORSOrigin, mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0, // SSE needs no write timeout
	}

	// Graceful shutdown on signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		sseServer.Shutdown(shutCtx)
		return httpServer.Shutdown(shutCtx)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	uptime := time.Since(s.startTime).Seconds()
	resp := map[string]any{
		"name":           "engram",
		"version":        s.version,
		"transport":      "http",
		"tools":          7,
		"uptime_seconds": int(uptime),
	}
	json.NewEncoder(w).Encode(resp)
}

// authMiddleware checks for a valid bearer token if configured.
func authMiddleware(token string, next http.Handler) http.Handler {
	if token == "" {
		return next // no auth required
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		expected := "Bearer " + token
		if auth != expected {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers.
func corsMiddleware(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
