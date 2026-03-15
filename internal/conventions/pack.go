package conventions

import (
	"encoding/json"
	"fmt"
)

// Pack represents a community convention pack.
type Pack struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Description  string       `json:"description"`
	Author       string       `json:"author"`
	Conventions  []Convention `json:"conventions"`
}

// PackInfo is a summary of an installed pack.
type PackInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Count       int    `json:"convention_count"`
}

// ParsePack reads and validates a convention pack from JSON.
func ParsePack(data []byte) (*Pack, error) {
	var pack Pack
	if err := json.Unmarshal(data, &pack); err != nil {
		return nil, fmt.Errorf("invalid pack JSON: %w", err)
	}
	if pack.Name == "" {
		return nil, fmt.Errorf("pack missing required field: name")
	}
	if len(pack.Conventions) == 0 {
		return nil, fmt.Errorf("pack has no conventions")
	}
	return &pack, nil
}

// MergeConventions combines community conventions with local ones.
// Local conventions always win on conflict (same pattern + language).
func MergeConventions(local []Convention, community []Convention) []Convention {
	// Build a set of local convention keys
	localKeys := make(map[string]bool)
	for _, c := range local {
		key := c.Pattern + "|" + c.Language
		localKeys[key] = true
	}

	// Start with all local conventions
	merged := make([]Convention, len(local))
	copy(merged, local)

	// Add community conventions that don't conflict
	for _, c := range community {
		key := c.Pattern + "|" + c.Language
		if !localKeys[key] {
			merged = append(merged, c)
		}
	}

	return merged
}
