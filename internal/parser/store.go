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

// StoreReferences writes parsed references to the code_references table.
func StoreReferences(store *storage.Store, refs []Reference) error {
	if len(refs) == 0 {
		return nil
	}
	db := store.DB()
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO code_references
		(file_path, to_name, kind, from_func, line, context, param_name, table_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	// Also prepare table_membership inserts for initializer refs with a table name
	tmStmt, err := tx.Prepare(`INSERT INTO table_membership
		(table_var, func_name, file_path, slot_index, line)
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare table_membership insert: %w", err)
	}
	defer tmStmt.Close()

	slotCounters := map[string]int{} // table_var -> next slot index

	for _, r := range refs {
		paramName := interface{}(nil)
		if r.ParamName != "" {
			paramName = r.ParamName
		}
		tableName := interface{}(nil)
		if r.TableName != "" {
			tableName = r.TableName
		}
		_, err := stmt.Exec(r.FilePath, r.ToName, r.Kind, r.FromFunc, r.Line, r.Context, paramName, tableName)
		if err != nil {
			return fmt.Errorf("insert reference %s: %w", r.ToName, err)
		}

		// Populate table_membership for initializer entries with a known table name
		if r.Kind == "initializer" && r.TableName != "" {
			idx := slotCounters[r.TableName]
			slotCounters[r.TableName] = idx + 1
			if _, err := tmStmt.Exec(r.TableName, r.ToName, r.FilePath, idx, r.Line); err != nil {
				return fmt.Errorf("insert table_membership %s.%s: %w", r.TableName, r.ToName, err)
			}
		}
	}

	return tx.Commit()
}

// DeleteFileReferences removes all references and table memberships for a file path (for re-indexing).
func DeleteFileReferences(store *storage.Store, filePath string) error {
	if _, err := store.DB().Exec("DELETE FROM code_references WHERE file_path = ?", filePath); err != nil {
		return err
	}
	_, err := store.DB().Exec("DELETE FROM table_membership WHERE file_path = ?", filePath)
	return err
}
