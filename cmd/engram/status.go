package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/TomOst-Sec/colony-project/internal/cli"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Engram index status",
		Long:  "Display statistics about the current code index, memories, and database.",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	store, repoRoot, dbPath, err := openDatabase()
	if err != nil {
		return err
	}
	defer store.Close()

	db := store.DB()

	dbInfo, err := os.Stat(dbPath)
	var dbSize string
	if err == nil {
		dbSize = formatSize(dbInfo.Size())
	} else {
		dbSize = "unknown"
	}

	fmt.Fprintln(os.Stdout, cli.Title.Render(fmt.Sprintf("Engram v%s", version)))
	fmt.Fprintf(os.Stdout, "%s %s\n", cli.Label.Render("Repository:"), repoRoot)
	fmt.Fprintf(os.Stdout, "%s %s (%s)\n\n", cli.Label.Render("Database:  "), dbPath, dbSize)

	var fileCount, symbolCount int
	row := db.QueryRow("SELECT COUNT(DISTINCT file_path), COUNT(*) FROM code_index")
	if err := row.Scan(&fileCount, &symbolCount); err != nil {
		fileCount, symbolCount = 0, 0
	}

	type langCount struct {
		lang  string
		count int
	}
	var languages []langCount
	rows, err := db.Query("SELECT language, COUNT(DISTINCT file_path) FROM code_index GROUP BY language ORDER BY COUNT(DISTINCT file_path) DESC")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var lc langCount
			if rows.Scan(&lc.lang, &lc.count) == nil {
				languages = append(languages, lc)
			}
		}
	}

	var embeddingCount int
	row = db.QueryRow("SELECT COUNT(*) FROM code_index WHERE embedding IS NOT NULL")
	if err := row.Scan(&embeddingCount); err != nil {
		embeddingCount = 0
	}

	var lastIndexed *string
	row = db.QueryRow("SELECT MAX(updated_at) FROM code_index")
	row.Scan(&lastIndexed)

	fmt.Fprintln(os.Stdout, cli.Subtitle.Render("Index:"))
	fmt.Fprintf(os.Stdout, "  Files indexed:    %d\n", fileCount)
	fmt.Fprintf(os.Stdout, "  Symbols:          %d\n", symbolCount)

	if len(languages) > 0 {
		parts := make([]string, len(languages))
		for i, lc := range languages {
			parts[i] = fmt.Sprintf("%s (%d)", lc.lang, lc.count)
		}
		fmt.Fprintf(os.Stdout, "  Languages:        %s\n", strings.Join(parts, ", "))
	}

	if symbolCount > 0 {
		pct := embeddingCount * 100 / symbolCount
		fmt.Fprintf(os.Stdout, "  Embeddings:       %d / %d (%d%%)\n", embeddingCount, symbolCount, pct)
	}
	if lastIndexed != nil {
		fmt.Fprintf(os.Stdout, "  Last indexed:     %s\n", *lastIndexed)
	}

	var memoryCount int
	row = db.QueryRow("SELECT COUNT(*) FROM memories WHERE deleted_at IS NULL")
	if err := row.Scan(&memoryCount); err != nil {
		memoryCount = 0
	}

	fmt.Fprintf(os.Stdout, "\n%s\n", cli.Subtitle.Render("Memories:"))
	fmt.Fprintf(os.Stdout, "  Total:            %d\n", memoryCount)

	if memoryCount > 0 {
		var memTypes []string
		typeRows, err := db.Query("SELECT type, COUNT(*) FROM memories WHERE deleted_at IS NULL GROUP BY type ORDER BY COUNT(*) DESC")
		if err == nil {
			defer typeRows.Close()
			for typeRows.Next() {
				var mtype string
				var cnt int
				if typeRows.Scan(&mtype, &cnt) == nil {
					memTypes = append(memTypes, fmt.Sprintf("%s (%d)", mtype, cnt))
				}
			}
		}
		if len(memTypes) > 0 {
			fmt.Fprintf(os.Stdout, "  By type:          %s\n", strings.Join(memTypes, ", "))
		}

		var oldest, newest *string
		db.QueryRow("SELECT MIN(created_at) FROM memories WHERE deleted_at IS NULL").Scan(&oldest)
		db.QueryRow("SELECT MAX(created_at) FROM memories WHERE deleted_at IS NULL").Scan(&newest)
		if oldest != nil {
			fmt.Fprintf(os.Stdout, "  Oldest:           %s\n", *oldest)
		}
		if newest != nil {
			fmt.Fprintf(os.Stdout, "  Newest:           %s\n", *newest)
		}
	}

	var conventionCount, archCount int
	db.QueryRow("SELECT COUNT(*) FROM conventions").Scan(&conventionCount)
	db.QueryRow("SELECT COUNT(*) FROM architecture").Scan(&archCount)

	fmt.Fprintf(os.Stdout, "\nConventions:        %d patterns detected\n", conventionCount)
	fmt.Fprintf(os.Stdout, "Architecture:       %d modules mapped\n", archCount)

	return nil
}
