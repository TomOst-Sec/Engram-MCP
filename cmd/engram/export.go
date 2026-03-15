package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	exportTables string
	exportPretty bool
)

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [file]",
		Short: "Export database to JSON",
		Long:  "Export all Engram data (memories, code index, conventions, git context) to JSON format.",
		RunE:  runExport,
	}
	cmd.Flags().StringVarP(&exportTables, "tables", "t", "", "Comma-separated tables to export (default: all)")
	cmd.Flags().BoolVar(&exportPretty, "pretty", false, "Pretty-print JSON output")
	return cmd
}

// ExportData is the top-level JSON structure for export.
type ExportData struct {
	Version    int                       `json:"version"`
	ExportedAt string                    `json:"exported_at"`
	RepoRoot   string                    `json:"repo_root,omitempty"`
	Tables     map[string][]ExportRow    `json:"tables"`
}

// ExportRow represents a generic database row as key-value pairs.
type ExportRow map[string]interface{}

var allTables = []string{"memories", "code_index", "conventions", "git_context", "architecture"}

func runExport(cmd *cobra.Command, args []string) error {
	repoRoot, _ := detectRepoRoot()

	store, _, _, err := openDatabase()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer store.Close()

	// Determine which tables to export
	tables := allTables
	if exportTables != "" {
		tables = strings.Split(exportTables, ",")
		for i := range tables {
			tables[i] = strings.TrimSpace(tables[i])
		}
	}

	data := ExportData{
		Version:    1,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		RepoRoot:   repoRoot,
		Tables:     make(map[string][]ExportRow),
	}

	db := store.DB()
	for _, table := range tables {
		rows, err := db.Query("SELECT * FROM " + table)
		if err != nil {
			// Table might not exist or be empty — skip silently
			data.Tables[table] = []ExportRow{}
			continue
		}

		cols, err := rows.Columns()
		if err != nil {
			rows.Close()
			continue
		}

		var tableRows []ExportRow
		for rows.Next() {
			values := make([]interface{}, len(cols))
			valuePtrs := make([]interface{}, len(cols))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				continue
			}

			row := ExportRow{}
			for i, col := range cols {
				// Skip embedding blob columns
				if col == "embedding" {
					if values[i] != nil {
						row["has_embedding"] = true
					} else {
						row["has_embedding"] = false
					}
					continue
				}
				val := values[i]
				// Convert []byte to string for JSON
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			}
			tableRows = append(tableRows, row)
		}
		rows.Close()

		if tableRows == nil {
			tableRows = []ExportRow{}
		}
		data.Tables[table] = tableRows
	}

	// Marshal JSON
	var jsonBytes []byte
	if exportPretty {
		jsonBytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(data)
	}
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	// Write to file or stdout
	if len(args) > 0 {
		if err := os.WriteFile(args[0], jsonBytes, 0644); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Exported to %s\n", args[0])
	} else {
		os.Stdout.Write(jsonBytes)
		os.Stdout.Write([]byte("\n"))
	}

	return nil
}
