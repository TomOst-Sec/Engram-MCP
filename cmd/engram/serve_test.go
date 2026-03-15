package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeCommandRegistered(t *testing.T) {
	root := newRootCmd()
	cmd, _, err := root.Find([]string{"serve"})
	require.NoError(t, err)
	assert.Equal(t, "serve", cmd.Name())
}

func TestServeHelpOutput(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"serve", "--help"})

	err := root.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "MCP server")
	assert.Contains(t, output, "stdio")
}

func TestServeInvalidTransport(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"serve", "--transport", "invalid"})

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown transport")
}

func TestServeHTTPTransportNotImplemented(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"serve", "--transport", "http"})

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestServeInvalidLogLevel(t *testing.T) {
	root := newRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"serve", "--log-level", "trace"})

	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown log level")
}

func TestVersionIsNonEmpty(t *testing.T) {
	assert.NotEmpty(t, version)
}

func TestDetectRepoRoot(t *testing.T) {
	root, err := detectRepoRoot()
	require.NoError(t, err)
	assert.NotEmpty(t, root)

	gitDir := filepath.Join(root, ".git")
	info, err := os.Stat(gitDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestDetectRepoRootFromNonGitDir(t *testing.T) {
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	_, err = detectRepoRoot()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no .git directory found")
}

func TestGoModulePath(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/example/project\n\ngo 1.21\n"), 0644)
	require.NoError(t, err)

	modPath := goModulePath(tmpDir)
	assert.Equal(t, "github.com/example/project", modPath)
}

func TestGoModulePathNoGoMod(t *testing.T) {
	tmpDir := t.TempDir()
	modPath := goModulePath(tmpDir)
	assert.Equal(t, "", modPath)
}

func TestDatabaseDir(t *testing.T) {
	dir := databaseDir("/some/repo/path")
	assert.NotEmpty(t, dir)
	assert.Contains(t, dir, ".engram")

	dir2 := databaseDir("/some/repo/path")
	assert.Equal(t, dir, dir2)

	dir3 := databaseDir("/other/repo")
	assert.NotEqual(t, dir, dir3)
}
