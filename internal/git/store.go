package git

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// UpsertFileHistory stores or updates git context for a file.
func UpsertFileHistory(store *storage.Store, fh *FileHistory) error {
	coChanged, err := json.Marshal(fh.CoChangedFiles)
	if err != nil {
		return fmt.Errorf("marshaling co_changed_files: %w", err)
	}

	tx, err := store.DB().Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing entry for this file (file-level, symbol_name = '')
	_, err = tx.Exec(`DELETE FROM git_context WHERE file_path = ? AND (symbol_name = '' OR symbol_name IS NULL)`, fh.FilePath)
	if err != nil {
		return fmt.Errorf("deleting old history for %s: %w", fh.FilePath, err)
	}

	_, err = tx.Exec(
		`INSERT INTO git_context
		(file_path, symbol_name, last_author, last_commit_hash, last_commit_message, last_modified, change_frequency, co_changed_files, updated_at)
		VALUES (?, '', ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		fh.FilePath, fh.LastAuthor, fh.LastCommitHash, fh.LastCommitMessage,
		fh.LastModified, fh.ChangeFrequency, string(coChanged),
	)
	if err != nil {
		return fmt.Errorf("inserting file history for %s: %w", fh.FilePath, err)
	}

	return tx.Commit()
}

// GetFileHistory retrieves git context for a file.
func GetFileHistory(store *storage.Store, filePath string) (*FileHistory, error) {
	fh := &FileHistory{FilePath: filePath}
	var coChangedJSON sql.NullString
	var lastModified sql.NullTime

	err := store.DB().QueryRow(
		`SELECT last_author, last_commit_hash, last_commit_message, last_modified, change_frequency, co_changed_files
		FROM git_context WHERE file_path = ?`, filePath,
	).Scan(&fh.LastAuthor, &fh.LastCommitHash, &fh.LastCommitMessage, &lastModified, &fh.ChangeFrequency, &coChangedJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying file history for %s: %w", filePath, err)
	}

	if lastModified.Valid {
		fh.LastModified = lastModified.Time
	}

	if coChangedJSON.Valid && coChangedJSON.String != "" {
		if err := json.Unmarshal([]byte(coChangedJSON.String), &fh.CoChangedFiles); err != nil {
			return nil, fmt.Errorf("parsing co_changed_files for %s: %w", filePath, err)
		}
	}

	return fh, nil
}

// GetHotspots retrieves files ordered by change_frequency DESC.
func GetHotspots(store *storage.Store, limit int) ([]FileHistory, error) {
	rows, err := store.DB().Query(
		`SELECT file_path, last_author, last_commit_hash, last_commit_message, last_modified, change_frequency, co_changed_files
		FROM git_context
		WHERE change_frequency > 0
		ORDER BY change_frequency DESC
		LIMIT ?`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("querying hotspots: %w", err)
	}
	defer rows.Close()

	var results []FileHistory
	for rows.Next() {
		var fh FileHistory
		var coChangedJSON sql.NullString
		var lastModified sql.NullTime

		if err := rows.Scan(&fh.FilePath, &fh.LastAuthor, &fh.LastCommitHash,
			&fh.LastCommitMessage, &lastModified, &fh.ChangeFrequency, &coChangedJSON); err != nil {
			return nil, fmt.Errorf("scanning hotspot row: %w", err)
		}

		if lastModified.Valid {
			fh.LastModified = lastModified.Time
		}

		if coChangedJSON.Valid && coChangedJSON.String != "" {
			_ = json.Unmarshal([]byte(coChangedJSON.String), &fh.CoChangedFiles)
		}

		results = append(results, fh)
	}

	return results, rows.Err()
}
