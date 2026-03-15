package git

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/TomOst-Sec/colony-project/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertAndGetFileHistory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	now := time.Now().Truncate(time.Second)
	fh := &FileHistory{
		FilePath:          "internal/auth/handler.go",
		LastAuthor:        "alice",
		LastCommitHash:    "abc123def456",
		LastCommitMessage: "Fix session token expiry",
		LastModified:      now,
		ChangeFrequency:   15,
		CoChangedFiles:    []string{"internal/auth/middleware.go", "tests/auth_test.go"},
	}

	err = UpsertFileHistory(store, fh)
	require.NoError(t, err)

	got, err := GetFileHistory(store, "internal/auth/handler.go")
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "alice", got.LastAuthor)
	assert.Equal(t, "abc123def456", got.LastCommitHash)
	assert.Equal(t, "Fix session token expiry", got.LastCommitMessage)
	assert.Equal(t, 15, got.ChangeFrequency)
	assert.Len(t, got.CoChangedFiles, 2)
	assert.Equal(t, "internal/auth/middleware.go", got.CoChangedFiles[0])
}

func TestGetFileHistoryNotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	got, err := GetFileHistory(store, "nonexistent.go")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestUpsertOverwrites(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	fh1 := &FileHistory{
		FilePath:        "file.go",
		LastAuthor:      "alice",
		ChangeFrequency: 5,
		CoChangedFiles:  []string{},
	}
	require.NoError(t, UpsertFileHistory(store, fh1))

	fh2 := &FileHistory{
		FilePath:        "file.go",
		LastAuthor:      "bob",
		ChangeFrequency: 10,
		CoChangedFiles:  []string{"other.go"},
	}
	require.NoError(t, UpsertFileHistory(store, fh2))

	got, err := GetFileHistory(store, "file.go")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "bob", got.LastAuthor)
	assert.Equal(t, 10, got.ChangeFrequency)
}

func TestGetHotspots(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	files := []FileHistory{
		{FilePath: "low.go", LastAuthor: "a", ChangeFrequency: 2, CoChangedFiles: []string{}},
		{FilePath: "high.go", LastAuthor: "b", ChangeFrequency: 42, CoChangedFiles: []string{}},
		{FilePath: "mid.go", LastAuthor: "c", ChangeFrequency: 15, CoChangedFiles: []string{}},
	}
	for _, fh := range files {
		require.NoError(t, UpsertFileHistory(store, &fh))
	}

	results, err := GetHotspots(store, 10)
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, "high.go", results[0].FilePath)
	assert.Equal(t, 42, results[0].ChangeFrequency)
	assert.Equal(t, "mid.go", results[1].FilePath)
	assert.Equal(t, "low.go", results[2].FilePath)
}

func TestGetHotspotsLimit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	for i := 0; i < 5; i++ {
		fh := &FileHistory{
			FilePath:        fmt.Sprintf("file%d.go", i),
			ChangeFrequency: i + 1,
			CoChangedFiles:  []string{},
		}
		require.NoError(t, UpsertFileHistory(store, fh))
	}

	results, err := GetHotspots(store, 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestCoChangedFilesJSONRoundTrip(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	defer store.Close()

	fh := &FileHistory{
		FilePath:       "main.go",
		LastAuthor:     "dev",
		CoChangedFiles: []string{"a.go", "b.go", "c.go"},
	}
	require.NoError(t, UpsertFileHistory(store, fh))

	got, err := GetFileHistory(store, "main.go")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, []string{"a.go", "b.go", "c.go"}, got.CoChangedFiles)
}
