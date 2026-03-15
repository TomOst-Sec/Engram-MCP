package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var importReplace bool

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import database from JSON",
		Long:  "Import Engram data from a JSON export file. Merges with existing data.",
		RunE:  runImport,
	}
	cmd.Flags().BoolVar(&importReplace, "replace", false, "Clear existing data before importing")
	return cmd
}

func runImport(cmd *cobra.Command, args []string) error {
	store, _, _, err := openDatabase()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer store.Close()

	// Read JSON from file or stdin
	var input io.Reader
	if len(args) > 0 {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer f.Close()
		input = f
	} else {
		input = os.Stdin
	}

	var data ExportData
	if err := json.NewDecoder(input).Decode(&data); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}

	if data.Version != 1 {
		return fmt.Errorf("unsupported export version: %d (expected 1)", data.Version)
	}

	db := store.DB()
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	counts := make(map[string]int)

	for table, rows := range data.Tables {
		if importReplace {
			if _, err := tx.Exec("DELETE FROM " + table); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not clear %s: %v\n", table, err)
			}
		}

		for _, row := range rows {
			// Skip has_embedding pseudo-field
			delete(row, "has_embedding")

			if len(row) == 0 {
				continue
			}

			cols := make([]string, 0, len(row))
			placeholders := make([]string, 0, len(row))
			values := make([]interface{}, 0, len(row))

			for col, val := range row {
				cols = append(cols, col)
				placeholders = append(placeholders, "?")
				values = append(values, val)
			}

			query := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES (%s)",
				table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))

			if _, err := tx.Exec(query, values...); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to insert into %s: %v\n", table, err)
				continue
			}
			counts[table]++
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Print summary
	fmt.Fprint(os.Stderr, "Imported:")
	for table, count := range counts {
		fmt.Fprintf(os.Stderr, " %d %s,", count, table)
	}
	fmt.Fprintln(os.Stderr)

	return nil
}
