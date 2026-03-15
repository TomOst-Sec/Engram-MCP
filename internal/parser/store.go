package parser

import (
	"fmt"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// StoreSymbols writes parsed symbols to the code_index table.
func StoreSymbols(store *storage.Store, fileHash string, symbols []Symbol) error {
	db := store.DB()
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO code_index
		(file_path, file_hash, language, symbol_name, symbol_type, signature, docstring, start_line, end_line, body_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, s := range symbols {
		_, err := stmt.Exec(s.FilePath, fileHash, s.Language, s.Name, s.Type, s.Signature, s.Docstring, s.StartLine, s.EndLine, s.BodyHash)
		if err != nil {
			return fmt.Errorf("insert symbol %s: %w", s.Name, err)
		}
	}

	return tx.Commit()
}

// DeleteFileSymbols removes all symbols for a file path (for re-indexing).
func DeleteFileSymbols(store *storage.Store, filePath string) error {
	_, err := store.DB().Exec("DELETE FROM code_index WHERE file_path = ?", filePath)
	return err
}

// GetFileHash retrieves the stored hash for a file to check if re-indexing is needed.
func GetFileHash(store *storage.Store, filePath string) (string, error) {
	var hash string
	err := store.DB().QueryRow("SELECT file_hash FROM code_index WHERE file_path = ? LIMIT 1", filePath).Scan(&hash)
	if err != nil {
		return "", err
	}
	return hash, nil
}
