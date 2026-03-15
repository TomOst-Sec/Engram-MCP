package git

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/TomOst-Sec/colony-project/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunGitVersion(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Use this repo's root as repoRoot
	h := New(store, ".")
	out, err := h.RunGit(context.Background(), "--version")
	require.NoError(t, err)
	assert.Contains(t, out, "git version")
}

func TestRunGitInvalidCommand(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	h := New(store, t.TempDir())
	_, err = h.RunGit(context.Background(), "not-a-real-command")
	assert.Error(t, err)
}

func TestRunGitCancelledContext(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	h := New(store, ".")
	_, err = h.RunGit(ctx, "--version")
	assert.Error(t, err)
}
