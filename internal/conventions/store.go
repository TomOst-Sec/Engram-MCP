package conventions

import (
	"encoding/json"
	"fmt"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// StoreConventions clears existing conventions and stores new ones.
func StoreConventions(store *storage.Store, conventions []Convention) error {
	db := store.DB()
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing
	if _, err := tx.Exec("DELETE FROM conventions"); err != nil {
		return fmt.Errorf("clear conventions: %w", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO conventions
		(pattern, description, category, confidence, examples, language)
		VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, c := range conventions {
		examplesJSON, _ := json.Marshal(c.Examples)
		_, err := stmt.Exec(c.Pattern, c.Description, c.Category, c.Confidence, string(examplesJSON), c.Language)
		if err != nil {
			return fmt.Errorf("insert convention %s: %w", c.Pattern, err)
		}
	}

	return tx.Commit()
}

// GetConventions retrieves conventions from the database, optionally filtered.
func GetConventions(store *storage.Store, language string, category string) ([]Convention, error) {
	db := store.DB()

	query := "SELECT pattern, description, category, confidence, examples, language FROM conventions WHERE 1=1"
	var args []interface{}

	if language != "" {
		query += " AND (language = ? OR language = 'all')"
		args = append(args, language)
	}
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	query += " ORDER BY confidence DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query conventions: %w", err)
	}
	defer rows.Close()

	var conventions []Convention
	for rows.Next() {
		var c Convention
		var examplesJSON string
		if err := rows.Scan(&c.Pattern, &c.Description, &c.Category, &c.Confidence, &examplesJSON, &c.Language); err != nil {
			continue
		}
		if examplesJSON != "" {
			json.Unmarshal([]byte(examplesJSON), &c.Examples)
		}
		conventions = append(conventions, c)
	}

	return conventions, nil
}
