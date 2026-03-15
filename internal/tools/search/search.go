package search

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TomOst-Sec/colony-project/internal/embeddings"
	"github.com/TomOst-Sec/colony-project/internal/storage"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// SearchTool implements the search_code MCP tool.
type SearchTool struct {
	store    *storage.Store
	embedder *embeddings.Embedder // may be nil (graceful degradation)
	repoRoot string
}

// NewSearchTool creates a new SearchTool.
func NewSearchTool(store *storage.Store, embedder *embeddings.Embedder, repoRoot string) *SearchTool {
	return &SearchTool{
		store:    store,
		embedder: embedder,
		repoRoot: repoRoot,
	}
}

// Definition returns the MCP tool definition with schema.
func (t *SearchTool) Definition() mcpgo.Tool {
	return mcpgo.NewTool("search_code",
		mcpgo.WithDescription("Search the codebase using natural language or keyword queries. Returns ranked code snippets with file paths, line numbers, and relevance scores."),
		mcpgo.WithString("query",
			mcpgo.Required(),
			mcpgo.Description("Natural language or keyword search query"),
		),
		mcpgo.WithString("language",
			mcpgo.Description("Filter by programming language (e.g., 'go', 'python', 'typescript')"),
		),
		mcpgo.WithString("symbol_type",
			mcpgo.Description("Filter by symbol type"),
			mcpgo.Enum("function", "method", "type", "class", "interface", "import", "export", "test"),
		),
		mcpgo.WithString("directory",
			mcpgo.Description("Filter to symbols within this directory path"),
		),
		mcpgo.WithNumber("limit",
			mcpgo.Description("Maximum results to return (default: 10, max: 50)"),
		),
	)
}

// searchResponse is the JSON response structure for the search tool.
type searchResponse struct {
	Results      []searchResult `json:"results"`
	TotalMatches int            `json:"total_matches"`
	Query        string         `json:"query"`
	SearchMode   string         `json:"search_mode"`
}

type searchResult struct {
	FilePath   string  `json:"file_path"`
	SymbolName string  `json:"symbol_name"`
	SymbolType string  `json:"symbol_type"`
	Language   string  `json:"language"`
	Signature  string  `json:"signature"`
	StartLine  int     `json:"start_line"`
	EndLine    int     `json:"end_line"`
	Score      float64 `json:"score"`
	Context    string  `json:"context,omitempty"`
}

// Handle processes a search_code tool request.
func (t *SearchTool) Handle(ctx context.Context, request mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	args := request.GetArguments()

	query, _ := args["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query parameter is required and must be non-empty")
	}

	language, _ := args["language"].(string)
	symbolType, _ := args["symbol_type"].(string)
	directory, _ := args["directory"].(string)

	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	if limit > 50 {
		limit = 50
	}

	// Step 1: FTS5 keyword search
	ftsResults, err := t.searchFTS5(query, language, symbolType, directory)
	if err != nil {
		return nil, fmt.Errorf("FTS5 search failed: %w", err)
	}

	// Step 2: Vector similarity search (if embedder available)
	var vectorResults []VectorResult
	searchMode := "fts5"
	if t.embedder != nil {
		vectorResults, err = t.searchVector(query, language, symbolType, directory)
		if err != nil {
			// Non-fatal: log and continue with FTS5-only
			vectorResults = nil
		}
		if vectorResults != nil {
			searchMode = "hybrid"
		}
	}

	// Step 3: Normalize FTS5 scores and hybrid rank
	normalizedFTS := NormalizeFTSScores(ftsResults)
	ranked := HybridRank(normalizedFTS, vectorResults)

	// Step 4: Apply limit
	totalMatches := len(ranked)
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}

	// Step 5: Build response with context
	results := make([]searchResult, len(ranked))
	for i, r := range ranked {
		results[i] = searchResult{
			FilePath:   r.FilePath,
			SymbolName: r.SymbolName,
			SymbolType: r.SymbolType,
			Language:   r.Language,
			Signature:  r.Signature,
			StartLine:  r.StartLine,
			EndLine:    r.EndLine,
			Score:      r.Score,
			Context:    t.getContext(r.FilePath, r.StartLine, r.EndLine),
		}
	}

	resp := searchResponse{
		Results:      results,
		TotalMatches: totalMatches,
		Query:        query,
		SearchMode:   searchMode,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return mcpgo.NewToolResultText(string(data)), nil
}

// searchFTS5 performs FTS5 full-text search on the code_index_fts table.
func (t *SearchTool) searchFTS5(query, language, symbolType, directory string) ([]FTSResult, error) {
	sqlQuery := `SELECT ci.id, ci.file_path, ci.symbol_name, ci.symbol_type, ci.language,
		ci.signature, ci.start_line, ci.end_line, fts.rank
		FROM code_index_fts fts
		JOIN code_index ci ON ci.id = fts.rowid
		WHERE code_index_fts MATCH ?`
	args := []interface{}{query}

	if language != "" {
		sqlQuery += " AND ci.language = ?"
		args = append(args, language)
	}
	if symbolType != "" {
		sqlQuery += " AND ci.symbol_type = ?"
		args = append(args, symbolType)
	}
	if directory != "" {
		sqlQuery += " AND ci.file_path LIKE ?"
		args = append(args, directory+"%")
	}

	sqlQuery += " ORDER BY fts.rank LIMIT 100"

	rows, err := t.store.DB().Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []FTSResult
	for rows.Next() {
		var r FTSResult
		if err := rows.Scan(&r.ID, &r.FilePath, &r.SymbolName, &r.SymbolType,
			&r.Language, &r.Signature, &r.StartLine, &r.EndLine, &r.Rank); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// searchVector performs vector similarity search using embeddings.
func (t *SearchTool) searchVector(query, language, symbolType, directory string) ([]VectorResult, error) {
	queryEmb, err := t.embedder.Embed(query)
	if err != nil {
		return nil, err
	}
	if queryEmb == nil {
		return nil, nil // NoOp embedder
	}

	filter := embeddings.SearchFilter{
		Language:   language,
		SymbolType: symbolType,
		Directory:  directory,
	}

	embResults, err := embeddings.SearchByVector(t.store, queryEmb, 100, filter)
	if err != nil {
		return nil, err
	}

	results := make([]VectorResult, len(embResults))
	for i, r := range embResults {
		// Get additional fields from code_index
		var sig, lang string
		_ = t.store.DB().QueryRow(
			"SELECT COALESCE(signature, ''), language FROM code_index WHERE id = ?", r.ID,
		).Scan(&sig, &lang)

		results[i] = VectorResult{
			ID:         r.ID,
			FilePath:   r.FilePath,
			SymbolName: r.SymbolName,
			SymbolType: r.SymbolType,
			Language:   lang,
			Signature:  sig,
			StartLine:  r.StartLine,
			EndLine:    r.EndLine,
			Score:      r.Score,
		}
	}

	return results, nil
}

// getContext reads surrounding lines from a source file for context.
// Returns empty string if file cannot be read.
func (t *SearchTool) getContext(filePath string, startLine, endLine int) string {
	fullPath := filepath.Join(t.repoRoot, filePath)
	f, err := os.Open(fullPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	contextStart := startLine - 3
	if contextStart < 1 {
		contextStart = 1
	}
	contextEnd := endLine + 3

	scanner := bufio.NewScanner(f)
	var result string
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum >= contextStart && lineNum <= contextEnd {
			if result != "" {
				result += "\n"
			}
			result += scanner.Text()
		}
		if lineNum > contextEnd {
			break
		}
	}

	return result
}
