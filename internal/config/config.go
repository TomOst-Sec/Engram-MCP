package config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds all Engram configuration values.
type Config struct {
	// Storage
	DatabasePath string `json:"database_path"`
	WALMode      bool   `json:"wal_mode"`

	// Server
	Transport string `json:"transport"`
	HTTPAddr  string `json:"http_addr"`
	HTTPToken string `json:"http_token"`

	// Indexing
	Languages      []string `json:"languages"`
	IgnorePatterns []string `json:"ignore_patterns"`
	MaxFileSize    int64    `json:"max_file_size"`

	// Embeddings
	EmbeddingModel     string `json:"embedding_model"`
	OllamaEndpoint     string `json:"ollama_endpoint"`
	OllamaModel        string `json:"ollama_model"`
	EmbeddingBatchSize int    `json:"embedding_batch_size"`

	// Multi-repo
	AdditionalRepos []RepoRef `json:"additional_repos"`

	// Derived (not in JSON)
	RepoRoot string `json:"-"`
	RepoHash string `json:"-"`
}

// RepoRef references an additional repository for multi-repo support.
type RepoRef struct {
	Path string `json:"path"` // relative or absolute path
	Name string `json:"name"` // display name for search results
}

// Load loads configuration with precedence: env vars > engram.json > defaults.
func Load(repoRoot string) (*Config, error) {
	cfg := Default()

	// Try loading from engram.json using raw map to detect which keys are present
	jsonPath := filepath.Join(repoRoot, "engram.json")
	if data, err := os.ReadFile(jsonPath); err == nil {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
		applyRawOverrides(cfg, raw)
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	// Set derived fields
	cfg.RepoRoot = repoRoot
	cfg.RepoHash = RepoHash(repoRoot)

	// Set default database path if not specified
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = filepath.Join(DatabaseDir(repoRoot), "engram.db")
	}

	return cfg, nil
}

// LoadFromFile reads and parses an engram.json file.
// Returns only the values present in the file (no defaults applied).
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// Validate checks the config for invalid values.
func (c *Config) Validate() error {
	if c.Transport != "stdio" && c.Transport != "http" {
		return fmt.Errorf("invalid transport %q: must be \"stdio\" or \"http\"", c.Transport)
	}
	if c.MaxFileSize < 0 {
		return fmt.Errorf("invalid max_file_size %d: must be non-negative", c.MaxFileSize)
	}
	return nil
}

// RepoHash returns a consistent 12-character hex string derived from the repo root path.
func RepoHash(repoRoot string) string {
	h := sha256.Sum256([]byte(repoRoot))
	return fmt.Sprintf("%x", h[:6])
}

// DatabaseDir returns the Engram data directory for a given repo root: ~/.engram/<repo-hash>/
func DatabaseDir(repoRoot string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}
	return filepath.Join(homeDir, ".engram", RepoHash(repoRoot))
}

// applyRawOverrides applies values from a raw JSON map onto cfg,
// only overriding fields that are explicitly present in the JSON.
func applyRawOverrides(cfg *Config, raw map[string]json.RawMessage) {
	if v, ok := raw["database_path"]; ok {
		json.Unmarshal(v, &cfg.DatabasePath)
	}
	if v, ok := raw["wal_mode"]; ok {
		json.Unmarshal(v, &cfg.WALMode)
	}
	if v, ok := raw["transport"]; ok {
		json.Unmarshal(v, &cfg.Transport)
	}
	if v, ok := raw["http_addr"]; ok {
		json.Unmarshal(v, &cfg.HTTPAddr)
	}
	if v, ok := raw["http_token"]; ok {
		json.Unmarshal(v, &cfg.HTTPToken)
	}
	if v, ok := raw["languages"]; ok {
		json.Unmarshal(v, &cfg.Languages)
	}
	if v, ok := raw["ignore_patterns"]; ok {
		json.Unmarshal(v, &cfg.IgnorePatterns)
	}
	if v, ok := raw["max_file_size"]; ok {
		json.Unmarshal(v, &cfg.MaxFileSize)
	}
	if v, ok := raw["embedding_model"]; ok {
		json.Unmarshal(v, &cfg.EmbeddingModel)
	}
	if v, ok := raw["ollama_endpoint"]; ok {
		json.Unmarshal(v, &cfg.OllamaEndpoint)
	}
	if v, ok := raw["ollama_model"]; ok {
		json.Unmarshal(v, &cfg.OllamaModel)
	}
	if v, ok := raw["embedding_batch_size"]; ok {
		json.Unmarshal(v, &cfg.EmbeddingBatchSize)
	}
	if v, ok := raw["additional_repos"]; ok {
		json.Unmarshal(v, &cfg.AdditionalRepos)
	}
}

// applyEnvOverrides reads ENGRAM_* environment variables and applies them.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("ENGRAM_DATABASE_PATH"); v != "" {
		cfg.DatabasePath = v
	}
	if v := os.Getenv("ENGRAM_TRANSPORT"); v != "" {
		cfg.Transport = v
	}
	if v := os.Getenv("ENGRAM_HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := os.Getenv("ENGRAM_HTTP_TOKEN"); v != "" {
		cfg.HTTPToken = v
	}
	if v := os.Getenv("ENGRAM_EMBEDDING_MODEL"); v != "" {
		cfg.EmbeddingModel = v
	}
	if v := os.Getenv("ENGRAM_OLLAMA_ENDPOINT"); v != "" {
		cfg.OllamaEndpoint = v
	}
	if v := os.Getenv("ENGRAM_OLLAMA_MODEL"); v != "" {
		cfg.OllamaModel = v
	}
	if v := os.Getenv("ENGRAM_MAX_FILE_SIZE"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.MaxFileSize = n
		}
	}
	if v := os.Getenv("ENGRAM_WAL_MODE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.WALMode = b
		}
	}
	if v := os.Getenv("ENGRAM_EMBEDDING_BATCH_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.EmbeddingBatchSize = n
		}
	}
}
