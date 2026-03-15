package history

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/mcp"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// HistoryTool exposes git history context via MCP.
type HistoryTool struct {
	store *storage.Store
}

// NewHistoryTool creates a new HistoryTool.
func NewHistoryTool(store *storage.Store) *HistoryTool {
	return &HistoryTool{store: store}
}

// Definition returns the MCP tool definition for get_history.
func (t *HistoryTool) Definition() mcpgo.Tool {
	return mcpgo.NewTool("get_history",
		mcpgo.WithDescription("Get git history context for files and symbols. Shows who last changed code, why (commit messages as decision context), change frequency (hotspots), and which files frequently change together."),
		mcpgo.WithString("file_path",
			mcpgo.Description("File path to get history for (relative to repo root)"),
		),
		mcpgo.WithString("mode",
			mcpgo.Required(),
			mcpgo.Description("Query mode: 'file' for single file history, 'hotspots' for most-changed files, 'cochanged' for files that change together"),
			mcpgo.Enum("file", "hotspots", "cochanged"),
		),
		mcpgo.WithNumber("limit",
			mcpgo.Description("Maximum number of results (default: 10)"),
		),
	)
}

// Handle dispatches on the mode parameter.
func (t *HistoryTool) Handle(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	args := request.GetArguments()

	mode, _ := args["mode"].(string)
	if mode == "" {
		return mcpgo.NewToolResultError("mode parameter is required"), nil
	}

	filePath, _ := args["file_path"].(string)

	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	switch mode {
	case "file":
		return t.handleFile(ctx, filePath)
	case "hotspots":
		return t.handleHotspots(ctx, limit)
	case "cochanged":
		return t.handleCochanged(ctx, filePath, limit)
	default:
		return mcpgo.NewToolResultError(fmt.Sprintf("unknown mode %q: must be file, hotspots, or cochanged", mode)), nil
	}
}

func (t *HistoryTool) handleFile(ctx context.Context, filePath string) (*mcpgo.CallToolResult, error) {
	if filePath == "" {
		return mcpgo.NewToolResultError("file_path is required for mode 'file'"), nil
	}

	var lastAuthor, commitHash, commitMessage sql.NullString
	var lastModified sql.NullTime
	var changeFrequency int
	var coChangedJSON sql.NullString

	err := t.store.DB().QueryRowContext(ctx,
		`SELECT last_author, last_commit_hash, last_commit_message, last_modified, change_frequency, co_changed_files
		FROM git_context WHERE file_path = ?`, filePath,
	).Scan(&lastAuthor, &commitHash, &commitMessage, &lastModified, &changeFrequency, &coChangedJSON)

	if err == sql.ErrNoRows {
		return mcpgo.NewToolResultText("No git history found for this file. Run 'engram index' to analyze git history."), nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying file history: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "File: %s\n\n", filePath)

	if lastModified.Valid {
		fmt.Fprintf(&sb, "Last modified: %s", lastModified.Time.Format("2006-01-02"))
	}
	if lastAuthor.Valid && lastAuthor.String != "" {
		fmt.Fprintf(&sb, " by %s", lastAuthor.String)
	}
	sb.WriteString("\n")

	if commitHash.Valid && commitHash.String != "" {
		shortHash := commitHash.String
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}
		msg := ""
		if commitMessage.Valid {
			msg = commitMessage.String
		}
		fmt.Fprintf(&sb, "Commit: %s \"%s\"\n", shortHash, msg)
	}

	fmt.Fprintf(&sb, "Change frequency: %d commits\n", changeFrequency)

	if coChangedJSON.Valid && coChangedJSON.String != "" {
		var coChanged []string
		if err := json.Unmarshal([]byte(coChangedJSON.String), &coChanged); err == nil && len(coChanged) > 0 {
			sb.WriteString("\nOften changed with:\n")
			for _, f := range coChanged {
				fmt.Fprintf(&sb, "  - %s\n", f)
			}
		}
	}

	return mcpgo.NewToolResultText(sb.String()), nil
}

func (t *HistoryTool) handleHotspots(ctx context.Context, limit int) (*mcpgo.CallToolResult, error) {
	rows, err := t.store.DB().QueryContext(ctx,
		`SELECT file_path, change_frequency, last_author
		FROM git_context
		WHERE change_frequency > 0
		ORDER BY change_frequency DESC
		LIMIT ?`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying hotspots: %w", err)
	}
	defer rows.Close()

	var sb strings.Builder
	sb.WriteString("Codebase Hotspots (most frequently changed files):\n\n")

	rank := 1
	found := false
	for rows.Next() {
		var filePath, lastAuthor string
		var freq int
		if err := rows.Scan(&filePath, &freq, &lastAuthor); err != nil {
			return nil, fmt.Errorf("scanning hotspot row: %w", err)
		}
		fmt.Fprintf(&sb, "%2d. %-50s — %d changes, last by %s\n", rank, filePath, freq, lastAuthor)
		rank++
		found = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating hotspots: %w", err)
	}

	if !found {
		return mcpgo.NewToolResultText("No git history found. Run 'engram index' to analyze git history."), nil
	}

	return mcpgo.NewToolResultText(sb.String()), nil
}

func (t *HistoryTool) handleCochanged(ctx context.Context, filePath string, limit int) (*mcpgo.CallToolResult, error) {
	if filePath == "" {
		return mcpgo.NewToolResultError("file_path is required for mode 'cochanged'"), nil
	}

	var coChangedJSON sql.NullString
	err := t.store.DB().QueryRowContext(ctx,
		`SELECT co_changed_files FROM git_context WHERE file_path = ?`, filePath,
	).Scan(&coChangedJSON)

	if err == sql.ErrNoRows {
		return mcpgo.NewToolResultText("No git history found for this file. Run 'engram index' to analyze git history."), nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying co-changed files: %w", err)
	}

	if !coChangedJSON.Valid || coChangedJSON.String == "" || coChangedJSON.String == "[]" {
		return mcpgo.NewToolResultText(fmt.Sprintf("No co-change data found for %s.", filePath)), nil
	}

	var coChanged []string
	if err := json.Unmarshal([]byte(coChangedJSON.String), &coChanged); err != nil {
		return nil, fmt.Errorf("parsing co-changed files: %w", err)
	}

	if limit > 0 && len(coChanged) > limit {
		coChanged = coChanged[:limit]
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Files that frequently change with %s:\n\n", filePath)
	for i, f := range coChanged {
		fmt.Fprintf(&sb, "%2d. %s\n", i+1, f)
	}

	return mcpgo.NewToolResultText(sb.String()), nil
}
