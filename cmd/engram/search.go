package main

import (
	"fmt"
	"os"
	"strings"
	"time"

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
		fmt.Fprintf(os.Stdout, "  %s:%d  %s  %s\n", filePath, startLine, symbolName, symbolType)
		if signature != "" {
			fmt.Fprintf(os.Stdout, "  %s\n", signature)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating results: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Fprintf(os.Stdout, "\nFound %d results for %q (%.1fms)\n", count, query, float64(elapsed.Microseconds())/1000.0)
	return nil
}
