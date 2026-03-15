package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	recallType  string
	recallLimit int
	recallSince string
)

func newRecallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recall <query>",
		Short: "Search past memories",
		Long:  "Search memories from past coding sessions using natural language.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runRecall,
	}
	cmd.Flags().StringVarP(&recallType, "type", "t", "", "Filter by memory type (decision, bugfix, refactor, learning, convention)")
	cmd.Flags().IntVarP(&recallLimit, "limit", "n", 10, "Maximum number of results")
	cmd.Flags().StringVar(&recallSince, "since", "", "Only memories after this date (YYYY-MM-DD)")
	return cmd
}

func runRecall(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	store, _, _, err := openDatabase()
	if err != nil {
		return err
	}
	defer store.Close()

	start := time.Now()

	sqlQuery := "SELECT m.id, m.content, m.type, m.tags, m.related_files, m.created_at FROM memories_fts fts JOIN memories m ON m.id = fts.rowid WHERE memories_fts MATCH ? AND m.deleted_at IS NULL"
	sqlArgs := []any{query}

	if recallType != "" {
		sqlQuery += " AND m.type = ?"
		sqlArgs = append(sqlArgs, recallType)
	}
	if recallSince != "" {
		sqlQuery += " AND m.created_at >= datetime(?)"
		sqlArgs = append(sqlArgs, recallSince)
	}

	sqlQuery += " ORDER BY rank LIMIT ?"
	sqlArgs = append(sqlArgs, recallLimit)

	rows, err := store.DB().Query(sqlQuery, sqlArgs...)
	if err != nil {
		return fmt.Errorf("recall query failed: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var content, memType, createdAt string
		var tags, relatedFiles *string
		if err := rows.Scan(&id, &content, &memType, &tags, &relatedFiles, &createdAt); err != nil {
			return fmt.Errorf("reading result: %w", err)
		}
		if count > 0 {
			fmt.Fprintln(os.Stdout)
		}
		fmt.Fprintf(os.Stdout, "  #%d  [%s]  %s\n", id, memType, createdAt)
		lines := strings.SplitN(content, "\n", 2)
		fmt.Fprintf(os.Stdout, "  %s\n", lines[0])
		if tags != nil && *tags != "" {
			fmt.Fprintf(os.Stdout, "  Tags: %s\n", *tags)
		}
		if relatedFiles != nil && *relatedFiles != "" {
			fmt.Fprintf(os.Stdout, "  Files: %s\n", *relatedFiles)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating results: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Fprintf(os.Stdout, "\nFound %d memories for %q (%.1fms)\n", count, query, float64(elapsed.Microseconds())/1000.0)
	return nil
}
