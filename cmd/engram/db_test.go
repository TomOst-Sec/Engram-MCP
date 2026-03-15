package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenDatabaseNoGitDir(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	_, _, _, err = openDatabase()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not detect repository root")
}

func TestFormatSize(t *testing.T) {
	assert.Equal(t, "0 B", formatSize(0))
	assert.Equal(t, "512 B", formatSize(512))
	assert.Equal(t, "1.0 KB", formatSize(1024))
	assert.Equal(t, "1.5 KB", formatSize(1536))
	assert.Equal(t, "1.0 MB", formatSize(1<<20))
	assert.Equal(t, "4.2 MB", formatSize(4404019))
}
