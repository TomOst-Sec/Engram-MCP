package parser

import (
	"fmt"
	"path/filepath"
	"sort"
)

// Registry manages available parsers, routing files to the correct parser by extension.
type Registry struct {
	parsers map[string]Parser // extension -> parser
}

// NewRegistry creates a new empty Registry.
func NewRegistry() *Registry {
	return &Registry{parsers: make(map[string]Parser)}
}

// Register adds a parser to the registry, mapping all its extensions.
func (r *Registry) Register(p Parser) {
	for _, ext := range p.Extensions() {
		r.parsers[ext] = p
	}
}

// ParserFor returns the parser that handles files with the given path's extension.
func (r *Registry) ParserFor(filePath string) (Parser, bool) {
	ext := filepath.Ext(filePath)
	p, ok := r.parsers[ext]
	return p, ok
}

// ParseFile parses a file using the appropriate parser based on extension.
func (r *Registry) ParseFile(filePath string, source []byte) ([]Symbol, error) {
	p, ok := r.ParserFor(filePath)
	if !ok {
		return nil, fmt.Errorf("no parser registered for %s", filepath.Ext(filePath))
	}
	return p.Parse(filePath, source)
}

// SupportedLanguages returns a sorted list of language names from all registered parsers.
func (r *Registry) SupportedLanguages() []string {
	seen := make(map[string]bool)
	for _, p := range r.parsers {
		seen[p.Language()] = true
	}
	langs := make([]string, 0, len(seen))
	for lang := range seen {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}
