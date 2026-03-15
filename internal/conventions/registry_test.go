package conventions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePack(t *testing.T) {
	data := []byte(`{
		"name": "test-pack",
		"version": "1.0.0",
		"description": "Test conventions",
		"author": "test",
		"conventions": [
			{"pattern": "test_pattern", "description": "A test", "category": "naming", "confidence": 0.9, "language": "go"}
		]
	}`)

	pack, err := ParsePack(data)
	require.NoError(t, err)
	assert.Equal(t, "test-pack", pack.Name)
	assert.Equal(t, "1.0.0", pack.Version)
	assert.Len(t, pack.Conventions, 1)
}

func TestParsePackMissingName(t *testing.T) {
	data := []byte(`{"conventions": [{"pattern": "x", "description": "y", "category": "z", "confidence": 0.5}]}`)
	_, err := ParsePack(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestParsePackNoConventions(t *testing.T) {
	data := []byte(`{"name": "empty", "conventions": []}`)
	_, err := ParsePack(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no conventions")
}

func TestMergeConventions(t *testing.T) {
	local := []Convention{
		{Pattern: "snake_case", Language: "go", Confidence: 0.9, Description: "local snake_case"},
	}
	community := []Convention{
		{Pattern: "snake_case", Language: "go", Confidence: 1.0, Description: "community snake_case"}, // conflict — local wins
		{Pattern: "PascalCase components", Language: "typescript", Confidence: 1.0, Description: "community pascal"},
	}

	merged := MergeConventions(local, community)
	assert.Len(t, merged, 2)
	// Local wins on conflict
	assert.Equal(t, "local snake_case", merged[0].Description)
	// Community added for non-conflicting
	assert.Equal(t, "community pascal", merged[1].Description)
}

func TestRegistryListEmpty(t *testing.T) {
	dir := t.TempDir()
	registry := NewPackRegistry(dir)
	packs, err := registry.List()
	require.NoError(t, err)
	assert.Len(t, packs, 0)
}

func TestRegistryInstallAndList(t *testing.T) {
	dir := t.TempDir()
	registry := NewPackRegistry(dir)

	// Write a pack file manually
	packData := []byte(`{"name":"test","version":"1.0","description":"Test","conventions":[{"pattern":"p","description":"d","category":"c","confidence":0.5}]}`)
	err := os.WriteFile(filepath.Join(dir, "test.json"), packData, 0644)
	require.NoError(t, err)

	packs, err := registry.List()
	require.NoError(t, err)
	assert.Len(t, packs, 1)
	assert.Equal(t, "test", packs[0].Name)
}

func TestRegistryRemove(t *testing.T) {
	dir := t.TempDir()
	registry := NewPackRegistry(dir)

	packData := []byte(`{"name":"test","version":"1.0","description":"Test","conventions":[{"pattern":"p","description":"d","category":"c","confidence":0.5}]}`)
	os.WriteFile(filepath.Join(dir, "test.json"), packData, 0644)

	err := registry.Remove("test")
	assert.NoError(t, err)

	packs, _ := registry.List()
	assert.Len(t, packs, 0)
}

func TestRegistryRemoveNotInstalled(t *testing.T) {
	dir := t.TempDir()
	registry := NewPackRegistry(dir)
	err := registry.Remove("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed")
}
