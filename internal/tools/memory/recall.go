package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// RecallTool searches memories from past coding sessions.
type RecallTool struct {
	store *storage.Store
}

// NewRecallTool creates a new RecallTool.
func NewRecallTool(store *storage.Store) *RecallTool {
	return &RecallTool{store: store}
}

// Definition returns the MCP tool definition with input schema.
func (t *RecallTool) Definition() mcpgo.Tool {
	return mcpgo.NewTool("recall",
		mcpgo.WithDescription("Search memories from past coding sessions"),
		mcpgo.WithString("query",
			mcpgo.Required(),
			mcpgo.Description("Natural language search query"),
		),
		mcpgo.WithString("type",
			mcpgo.Description("Filter by memory type"),
			mcpgo.Enum("decision", "bugfix", "refactor", "learning", "convention"),
		),
		mcpgo.WithArray("tags",
			mcpgo.Description("Filter by tags (AND logic)"),
		),
		mcpgo.WithNumber("limit",
			mcpgo.Description("Maximum results to return (default: 10)"),
		),
		mcpgo.WithString("since",
			mcpgo.Description("Only memories after this ISO date"),
		),
	)
}

// Handle processes a recall tool call.
func (t *RecallTool) Handle(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return toolError("query is required"), nil
	}
	if query == "" {
		return toolError("query must not be empty"), nil
	}

	limit := request.GetInt("limit", 10)
	if limit <= 0 {
		limit = 10
	}

	memType := request.GetString("type", "")
	since := request.GetString("since", "")

	// Build the SQL query
	var conditions []string
	var args []any

	// FTS5 match
	conditions = append(conditions, "memories_fts MATCH ?")
	args = append(args, query)

	// Soft-delete filter
	conditions = append(conditions, "m.deleted_at IS NULL")

	// Type filter
	if memType != "" {
		conditions = append(conditions, "m.type = ?")
		args = append(args, memType)
	}

	// Date filter
	if since != "" {
		conditions = append(conditions, "m.created_at > datetime(?)")
		args = append(args, since)
	}

	where := strings.Join(conditions, " AND ")

	// Count total matches
	countSQL := fmt.Sprintf(
		`SELECT COUNT(*) FROM memories_fts JOIN memories m ON memories_fts.rowid = m.id WHERE %s`,
		where,
	)
	var totalMatches int
	if err := t.store.DB().QueryRowContext(ctx, countSQL, args...).Scan(&totalMatches); err != nil {
		return nil, fmt.Errorf("counting memories: %w", err)
	}

	// Fetch results ranked by FTS5 relevance
	selectSQL := fmt.Sprintf(
		`SELECT m.id, m.content, m.type, m.tags, m.related_files, m.created_at, rank
		 FROM memories_fts
		 JOIN memories m ON memories_fts.rowid = m.id
		 WHERE %s
		 ORDER BY rank
		 LIMIT ?`,
		where,
	)
	args = append(args, limit)

	rows, err := t.store.DB().QueryContext(ctx, selectSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("searching memories: %w", err)
	}
	defer rows.Close()

	type memoryResult struct {
		ID             int64    `json:"id"`
		Content        string   `json:"content"`
		Type           string   `json:"type"`
		Tags           []string `json:"tags"`
		RelatedFiles   []string `json:"related_files"`
		CreatedAt      string   `json:"created_at"`
		RelevanceScore float64  `json:"relevance_score"`
	}

	var memories []memoryResult
	for rows.Next() {
		var m memoryResult
		var tagsJSON, filesJSON string
		var rank float64
		if err := rows.Scan(&m.ID, &m.Content, &m.Type, &tagsJSON, &filesJSON, &m.CreatedAt, &rank); err != nil {
			return nil, fmt.Errorf("scanning memory row: %w", err)
		}
		_ = json.Unmarshal([]byte(tagsJSON), &m.Tags)
		_ = json.Unmarshal([]byte(filesJSON), &m.RelatedFiles)
		if m.Tags == nil {
			m.Tags = []string{}
		}
		if m.RelatedFiles == nil {
			m.RelatedFiles = []string{}
		}
		// FTS5 rank is negative (closer to 0 = better); normalize to 0..1
		m.RelevanceScore = -rank
		memories = append(memories, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating memory rows: %w", err)
	}

	if memories == nil {
		memories = []memoryResult{}
	}

	response := map[string]any{
		"memories":      memories,
		"total_matches": totalMatches,
		"query":         query,
	}
	data, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("marshaling response: %w", err)
	}

	return mcpgo.NewToolResultText(string(data)), nil
}
