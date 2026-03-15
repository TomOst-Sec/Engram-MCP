package architecture

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// Module represents a detected project module (directory grouping).
type Module struct {
	Name                 string         `json:"name"`
	Path                 string         `json:"path"`
	Description          string         `json:"description"`
	Files                int            `json:"files"`
	Symbols              int            `json:"symbols"`
	Exports              []string       `json:"exports,omitempty"`
	Dependencies         []string       `json:"dependencies"`
	ExternalDependencies []string       `json:"external_dependencies"`
	ComplexityScore      float64        `json:"complexity_score"`
	FileDetails          []FileDetail   `json:"files_detail,omitempty"`
	ExportDetails        []ExportDetail `json:"export_details,omitempty"`
	Dependents           []string       `json:"dependents,omitempty"`
}

// FileDetail provides per-file info for detailed module view.
type FileDetail struct {
	Path    string `json:"path"`
	Symbols int    `json:"symbols"`
}

// ExportDetail provides per-export info for detailed module view.
type ExportDetail struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Signature string `json:"signature,omitempty"`
	File      string `json:"file"`
	Line      int    `json:"line"`
}

// DetectModules groups files in code_index by directory path at the given depth.
func DetectModules(store *storage.Store, depth int) ([]Module, error) {
	if depth < 1 {
		depth = 2
	}

	rows, err := store.DB().Query(`
		SELECT file_path, COUNT(*) as symbol_count
		FROM code_index
		GROUP BY file_path
		ORDER BY file_path
	`)
	if err != nil {
		return nil, fmt.Errorf("querying code_index: %w", err)
	}
	defer rows.Close()

	type fileInfo struct {
		path    string
		symbols int
	}

	moduleFiles := make(map[string][]fileInfo)
	moduleSymbolCount := make(map[string]int)

	for rows.Next() {
		var filePath string
		var symbolCount int
		if err := rows.Scan(&filePath, &symbolCount); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		modulePath := truncatePath(filePath, depth)
		moduleFiles[modulePath] = append(moduleFiles[modulePath], fileInfo{
			path:    filePath,
			symbols: symbolCount,
		})
		moduleSymbolCount[modulePath] += symbolCount
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	var modules []Module
	for modPath, files := range moduleFiles {
		mod := Module{
			Name:    modPath,
			Path:    modPath,
			Files:   len(files),
			Symbols: moduleSymbolCount[modPath],
		}

		for _, f := range files {
			mod.FileDetails = append(mod.FileDetails, FileDetail{
				Path:    f.path,
				Symbols: f.symbols,
			})
		}

		modules = append(modules, mod)
	}

	return modules, nil
}

// truncatePath returns the directory path truncated to the given depth.
func truncatePath(filePath string, depth int) string {
	dir := filepath.Dir(filePath)
	parts := strings.Split(filepath.ToSlash(dir), "/")
	if len(parts) > depth {
		parts = parts[:depth]
	}
	return strings.Join(parts, "/")
}

// BuildDependencyGraph maps import statements to project modules.
// Returns internal deps per module and external deps per module.
func BuildDependencyGraph(store *storage.Store, modules []Module, goModulePath string) (map[string][]string, map[string][]string, error) {
	// Build a set of known module paths for quick lookup
	knownModules := make(map[string]bool)
	for _, m := range modules {
		knownModules[m.Path] = true
	}

	internalDeps := make(map[string][]string)
	externalDeps := make(map[string][]string)

	rows, err := store.DB().Query(`
		SELECT file_path, symbol_name
		FROM code_index
		WHERE symbol_type = 'import'
		ORDER BY file_path
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("querying imports: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var filePath, importPath string
		if err := rows.Scan(&filePath, &importPath); err != nil {
			return nil, nil, fmt.Errorf("scanning import row: %w", err)
		}

		// Determine which module this file belongs to
		sourceModule := ""
		for _, m := range modules {
			if strings.HasPrefix(filePath, m.Path+"/") || strings.HasPrefix(filePath, m.Path) {
				sourceModule = m.Path
				break
			}
		}
		if sourceModule == "" {
			continue
		}

		// Resolve import to a project module
		targetModule := resolveImport(importPath, goModulePath, knownModules)
		if targetModule != "" && targetModule != sourceModule {
			if !containsString(internalDeps[sourceModule], targetModule) {
				internalDeps[sourceModule] = append(internalDeps[sourceModule], targetModule)
			}
		} else if targetModule == "" && importPath != "" {
			// External dependency
			if !containsString(externalDeps[sourceModule], importPath) {
				externalDeps[sourceModule] = append(externalDeps[sourceModule], importPath)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterating imports: %w", err)
	}

	return internalDeps, externalDeps, nil
}

// resolveImport attempts to map an import path to a known project module.
func resolveImport(importPath, goModulePath string, knownModules map[string]bool) string {
	// Go imports: strip the module prefix to get relative path
	if goModulePath != "" && strings.HasPrefix(importPath, goModulePath+"/") {
		relPath := strings.TrimPrefix(importPath, goModulePath+"/")
		// Check if relPath or any prefix is a known module
		for modPath := range knownModules {
			if strings.HasPrefix(relPath, modPath) {
				return modPath
			}
		}
		return relPath
	}

	// Python imports: convert dots to slashes
	if strings.Contains(importPath, ".") && !strings.Contains(importPath, "/") {
		relPath := strings.ReplaceAll(importPath, ".", "/")
		for modPath := range knownModules {
			if strings.HasPrefix(relPath, modPath) {
				return modPath
			}
		}
	}

	return ""
}

// GetModuleExports returns exported symbols for a module.
func GetModuleExports(store *storage.Store, modulePath string) ([]ExportDetail, error) {
	rows, err := store.DB().Query(`
		SELECT symbol_name, symbol_type, signature, file_path, start_line, language
		FROM code_index
		WHERE file_path LIKE ? AND symbol_type != 'import'
		ORDER BY file_path, start_line
	`, modulePath+"/%")
	if err != nil {
		return nil, fmt.Errorf("querying exports: %w", err)
	}
	defer rows.Close()

	var exports []ExportDetail
	for rows.Next() {
		var name, symType, file, language string
		var signature *string
		var line int
		if err := rows.Scan(&name, &symType, &signature, &file, &line, &language); err != nil {
			return nil, fmt.Errorf("scanning export: %w", err)
		}

		if isExported(name, language) {
			ed := ExportDetail{
				Name: name,
				Type: symType,
				File: filepath.Base(file),
				Line: line,
			}
			if signature != nil {
				ed.Signature = *signature
			}
			exports = append(exports, ed)
		}
	}
	return exports, rows.Err()
}

// isExported determines if a symbol is exported based on language conventions.
func isExported(name, language string) bool {
	if name == "" {
		return false
	}
	switch language {
	case "go":
		return name[0] >= 'A' && name[0] <= 'Z'
	case "python":
		return !strings.HasPrefix(name, "_")
	case "typescript", "javascript":
		return true // exports tracked via symbol_type
	default:
		return name[0] >= 'A' && name[0] <= 'Z'
	}
}

// ComputeComplexity calculates a complexity score in the 0.0–10.0 range.
func ComputeComplexity(symbols, imports, files int) float64 {
	raw := float64(symbols)*0.5 + float64(imports)*0.3 + float64(files)*0.2
	// Normalize: cap at 10.0
	if raw > 10.0 {
		return 10.0
	}
	if raw < 0 {
		return 0.0
	}
	return raw
}

// FindDependents returns modules that depend on the given module.
func FindDependents(targetModule string, depGraph map[string][]string) []string {
	var dependents []string
	for mod, deps := range depGraph {
		for _, dep := range deps {
			if dep == targetModule {
				dependents = append(dependents, mod)
				break
			}
		}
	}
	return dependents
}

func containsString(s []string, v string) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}
	return false
}
