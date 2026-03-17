package storage

import (
	"database/sql"
	"fmt"
)

// migration represents a schema migration to a specific version.
type migration struct {
	version    int
	statements []string
}

// allMigrations returns migrations in order. Each migration includes
// all statements needed to bring the schema to that version.
func allMigrations() []migration {
	var v1Stmts []string
	v1Stmts = append(v1Stmts, schemaV1...)
	v1Stmts = append(v1Stmts, ftsV1...)
	v1Stmts = append(v1Stmts, triggersV1...)
	v1Stmts = append(v1Stmts, indexesV1...)
	return []migration{
		{version: 1, statements: v1Stmts},
		{version: 2, statements: schemaV2},
		{version: 3, statements: schemaV3},
	}
}

// currentVersion returns the current schema version from the database.
// Returns 0 if the schema_version table doesn't exist yet.
func currentVersion(db *sql.DB) (int, error) {
	// Check if schema_version table exists
	var tableName string
	err := db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name='schema_version'`,
	).Scan(&tableName)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("checking schema_version table: %w", err)
	}

	var version int
	err = db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("reading schema version: %w", err)
	}
	return version, nil
}

// migrate applies any outstanding migrations to bring the schema up to date.
func migrate(db *sql.DB) error {
	current, err := currentVersion(db)
	if err != nil {
		return fmt.Errorf("getting current version: %w", err)
	}

	for _, m := range allMigrations() {
		if m.version <= current {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration v%d: %w", m.version, err)
		}

		for _, stmt := range m.statements {
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration v%d statement failed: %w\nSQL: %s", m.version, err, stmt)
			}
		}

		if _, err := tx.Exec(`INSERT INTO schema_version (version) VALUES (?)`, m.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("recording migration v%d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration v%d: %w", m.version, err)
		}
	}

	return nil
}
