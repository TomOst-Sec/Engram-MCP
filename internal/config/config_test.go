package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultReturnsCorrectValues(t *testing.T) {
	cfg := Default()

	if cfg.WALMode != true {
		t.Errorf("WALMode: got %v, want true", cfg.WALMode)
	}
	if cfg.Transport != "stdio" {
		t.Errorf("Transport: got %q, want %q", cfg.Transport, "stdio")
	}
	if cfg.HTTPAddr != ":3333" {
		t.Errorf("HTTPAddr: got %q, want %q", cfg.HTTPAddr, ":3333")
	}
	if cfg.MaxFileSize != 1048576 {
		t.Errorf("MaxFileSize: got %d, want %d", cfg.MaxFileSize, 1048576)
	}
	if cfg.EmbeddingModel != "builtin" {
		t.Errorf("EmbeddingModel: got %q, want %q", cfg.EmbeddingModel, "builtin")
	}
	if cfg.OllamaEndpoint != "http://localhost:11434" {
		t.Errorf("OllamaEndpoint: got %q, want %q", cfg.OllamaEndpoint, "http://localhost:11434")
	}
	if cfg.OllamaModel != "nomic-embed-text" {
		t.Errorf("OllamaModel: got %q, want %q", cfg.OllamaModel, "nomic-embed-text")
	}
	if cfg.EmbeddingBatchSize != 32 {
		t.Errorf("EmbeddingBatchSize: got %d, want %d", cfg.EmbeddingBatchSize, 32)
	}
	if len(cfg.Languages) != 6 {
		t.Errorf("Languages: got %d items, want 6", len(cfg.Languages))
	}
	if len(cfg.IgnorePatterns) != 5 {
		t.Errorf("IgnorePatterns: got %d items, want 5", len(cfg.IgnorePatterns))
	}
}

func TestLoadWithJSONFileOverridesFields(t *testing.T) {
	dir := t.TempDir()
	jsonCfg := map[string]interface{}{
		"transport":    "http",
		"http_addr":    ":9999",
		"max_file_size": 2097152,
	}
	data, err := json.Marshal(jsonCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "engram.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Overridden fields
	if cfg.Transport != "http" {
		t.Errorf("Transport: got %q, want %q", cfg.Transport, "http")
	}
	if cfg.HTTPAddr != ":9999" {
		t.Errorf("HTTPAddr: got %q, want %q", cfg.HTTPAddr, ":9999")
	}
	if cfg.MaxFileSize != 2097152 {
		t.Errorf("MaxFileSize: got %d, want %d", cfg.MaxFileSize, 2097152)
	}

	// Non-overridden fields should keep defaults
	if cfg.WALMode != true {
		t.Errorf("WALMode should remain default true, got %v", cfg.WALMode)
	}
	if cfg.EmbeddingModel != "builtin" {
		t.Errorf("EmbeddingModel should remain default %q, got %q", "builtin", cfg.EmbeddingModel)
	}
	if cfg.OllamaEndpoint != "http://localhost:11434" {
		t.Errorf("OllamaEndpoint should remain default, got %q", cfg.OllamaEndpoint)
	}
}

func TestLoadWithoutJSONFileReturnsDefaults(t *testing.T) {
	dir := t.TempDir() // empty dir, no engram.json

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	defaults := Default()
	if cfg.Transport != defaults.Transport {
		t.Errorf("Transport: got %q, want default %q", cfg.Transport, defaults.Transport)
	}
	if cfg.WALMode != defaults.WALMode {
		t.Errorf("WALMode: got %v, want default %v", cfg.WALMode, defaults.WALMode)
	}
	if cfg.MaxFileSize != defaults.MaxFileSize {
		t.Errorf("MaxFileSize: got %d, want default %d", cfg.MaxFileSize, defaults.MaxFileSize)
	}
	if cfg.EmbeddingModel != defaults.EmbeddingModel {
		t.Errorf("EmbeddingModel: got %q, want default %q", cfg.EmbeddingModel, defaults.EmbeddingModel)
	}
	if cfg.EmbeddingBatchSize != defaults.EmbeddingBatchSize {
		t.Errorf("EmbeddingBatchSize: got %d, want default %d", cfg.EmbeddingBatchSize, defaults.EmbeddingBatchSize)
	}
	// Derived fields should be set
	if cfg.RepoRoot != dir {
		t.Errorf("RepoRoot: got %q, want %q", cfg.RepoRoot, dir)
	}
	if cfg.RepoHash == "" {
		t.Error("RepoHash should not be empty")
	}
}

func TestValidateAcceptsValidConfig(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() returned error for valid config: %v", err)
	}

	cfg.Transport = "http"
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() returned error for transport=http: %v", err)
	}
}

func TestValidateRejectsInvalidTransport(t *testing.T) {
	cfg := Default()

	cfg.Transport = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should reject transport=\"invalid\"")
	}

	cfg.Transport = ""
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should reject transport=\"\"")
	}
}

func TestValidateRejectsNegativeMaxFileSize(t *testing.T) {
	cfg := Default()
	cfg.MaxFileSize = -1

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should reject MaxFileSize=-1")
	}
}

func TestRepoHashDeterminism(t *testing.T) {
	// Same path produces same hash
	hash1 := RepoHash("/some/repo/path")
	hash2 := RepoHash("/some/repo/path")
	if hash1 != hash2 {
		t.Errorf("RepoHash not deterministic: %q != %q", hash1, hash2)
	}

	// Hash is 12 hex chars
	if len(hash1) != 12 {
		t.Errorf("RepoHash length: got %d, want 12", len(hash1))
	}

	// Different paths produce different hashes
	hash3 := RepoHash("/different/repo/path")
	if hash1 == hash3 {
		t.Errorf("Different paths should produce different hashes: both got %q", hash1)
	}
}

func TestEnvVariableOverridesDefault(t *testing.T) {
	dir := t.TempDir()

	t.Setenv("ENGRAM_DATABASE_PATH", "/custom/db/path.db")
	defer os.Unsetenv("ENGRAM_DATABASE_PATH")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.DatabasePath != "/custom/db/path.db" {
		t.Errorf("DatabasePath: got %q, want %q", cfg.DatabasePath, "/custom/db/path.db")
	}
}

func TestDatabaseDirContainsRepoHash(t *testing.T) {
	dir := DatabaseDir("/some/repo")
	hash := RepoHash("/some/repo")

	if !filepath.IsAbs(dir) {
		t.Errorf("DatabaseDir should return absolute path, got %q", dir)
	}
	if filepath.Base(dir) != hash {
		t.Errorf("DatabaseDir should end with repo hash %q, got dir %q", hash, dir)
	}
}
