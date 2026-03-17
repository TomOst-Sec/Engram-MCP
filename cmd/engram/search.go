package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TomOst-Sec/colony-project/internal/cli"
	"github.com/spf13/cobra"
)

var (
	searchLanguage string
	searchType     string
	searchLimit    int
)

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the codebase",
		Long:  "Search indexed code using natural language or keywords. Uses FTS5 full-text search.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runSearch,
	}
	cmd.Flags().StringVarP(&searchLanguage, "language", "l", "", "Filter by language (e.g., go, python, typescript)")
	cmd.Flags().StringVarP(&searchType, "type", "t", "", "Filter by symbol type (e.g., function, class, type)")
	cmd.Flags().IntVarP(&searchLimit, "limit", "n", 10, "Maximum number of results")
	return cmd
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	store, _, _, err := openDatabase()
	if err != nil {
		return err
	}
	defer store.Close()

	start := time.Now()

	// Step 1: FTS5 symbol search
	sqlQuery := "SELECT ci.file_path, ci.symbol_name, ci.symbol_type, ci.language, ci.signature, ci.start_line FROM code_index_fts fts JOIN code_index ci ON ci.id = fts.rowid WHERE code_index_fts MATCH ?"
	sqlArgs := []any{query}

	if searchLanguage != "" {
		sqlQuery += " AND ci.language = ?"
		sqlArgs = append(sqlArgs, searchLanguage)
	}
	if searchType != "" {
		sqlQuery += " AND ci.symbol_type = ?"
		sqlArgs = append(sqlArgs, searchType)
	}

	sqlQuery += " ORDER BY rank LIMIT ?"
	sqlArgs = append(sqlArgs, searchLimit)

	rows, err := store.DB().Query(sqlQuery, sqlArgs...)
	if err != nil {
		return fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var filePath, symbolName, symbolType, language, signature string
		var startLine int
		if err := rows.Scan(&filePath, &symbolName, &symbolType, &language, &signature, &startLine); err != nil {
			return fmt.Errorf("reading result: %w", err)
		}
		if count > 0 {
			fmt.Fprintln(os.Stdout)
		}
		fmt.Fprintf(os.Stdout, "  %s  %s  %s\n",
			cli.FilePath.Render(fmt.Sprintf("%s:%d", filePath, startLine)),
			cli.SymbolName.Render(symbolName),
			cli.Label.Render(symbolType))
		if signature != "" {
			fmt.Fprintf(os.Stdout, "  %s\n", signature)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating results: %w", err)
	}

	// Step 2: Reference search — find where the query term is called/used.
	// Extract simple identifier from query for exact reference matching.
	refName := extractIdentifier(query)
	if refName != "" {
		refRows, refErr := store.DB().Query(
			`SELECT file_path, to_name, kind, from_func, line, context
			 FROM code_references WHERE to_name = ? ORDER BY file_path, line LIMIT ?`,
			refName, searchLimit)
		if refErr == nil {
			defer refRows.Close()
			refCount := 0
			for refRows.Next() {
				var filePath, toName, kind, fromFunc, context string
				var line int
				if err := refRows.Scan(&filePath, &toName, &kind, &fromFunc, &line, &context); err != nil {
					break
				}
				if refCount == 0 {
					fmt.Fprintf(os.Stdout, "\n  %s\n", cli.Label.Render(fmt.Sprintf("References to %q:", refName)))
				}
				fmt.Fprintf(os.Stdout, "  %s  %s  %s\n",
					cli.FilePath.Render(fmt.Sprintf("%s:%d", filePath, line)),
					cli.SymbolName.Render(fmt.Sprintf("%s → %s", fromFunc, toName)),
					cli.Label.Render(kind))
				if context != "" {
					fmt.Fprintf(os.Stdout, "  %s\n", context)
				}
				refCount++
				count++
			}
		}
	}

	elapsed := time.Since(start)
	fmt.Fprintf(os.Stdout, "\n%s\n", cli.Label.Render(fmt.Sprintf("Found %d results for %q (%.1fms)", count, query, float64(elapsed.Microseconds())/1000.0)))
	return nil
}

// extractIdentifier pulls a simple C identifier from a search query.
// For queries like "function f", "where is f called", etc., it returns "f".
// For simple identifiers like "f" or "dispatch_table", returns as-is.
func extractIdentifier(query string) string {
	// Remove common natural language prefixes
	for _, prefix := range []string{
		"where is ", "where does ", "who calls ", "callers of ",
		"references to ", "usages of ", "uses of ",
		"function ", "func ",
	} {
		lower := strings.ToLower(query)
		if strings.HasPrefix(lower, prefix) {
			query = query[len(prefix):]
			break
		}
	}
	// Remove common suffixes
	for _, suffix := range []string{
		" called", " used", " invoked", " referenced",
	} {
		lower := strings.ToLower(query)
		if strings.HasSuffix(lower, suffix) {
			query = query[:len(query)-len(suffix)]
			break
		}
	}
	query = strings.TrimSpace(query)

	// Validate it's a C identifier
	if len(query) == 0 {
		return ""
	}
	for i, ch := range query {
		if i == 0 {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_') {
				return ""
			}
		} else {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
				return ""
			}
		}
	}
	return query
}
