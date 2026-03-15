package config

// Default returns a Config with all default values populated.
func Default() *Config {
	return &Config{
		WALMode:            true,
		Transport:          "stdio",
		HTTPAddr:           ":3333",
		Languages:          []string{"go", "python", "typescript", "rust", "java", "csharp"},
		IgnorePatterns:     []string{"vendor/", "node_modules/", ".git/", "bin/", "dist/"},
		MaxFileSize:        1048576, // 1MB
		EmbeddingModel:     "builtin",
		OllamaEndpoint:     "http://localhost:11434",
		OllamaModel:        "nomic-embed-text",
		EmbeddingBatchSize: 32,
	}
}
