package conventions

import (
	"context"
	"fmt"
	"time"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// Convention represents an inferred code convention.
type Convention struct {
	Pattern     string   `json:"pattern"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Confidence  float64  `json:"confidence"`
	Examples    []string `json:"examples"`
	Language    string   `json:"language"`
}

// AnalysisResult holds the results of convention analysis.
type AnalysisResult struct {
	Conventions []Convention
	Duration    time.Duration
}

// Analyzer scans the code_index and infers conventions.
type Analyzer struct {
	store    *storage.Store
	repoRoot string
}

// New creates a new convention Analyzer.
func New(store *storage.Store, repoRoot string) *Analyzer {
	return &Analyzer{store: store, repoRoot: repoRoot}
}

// Analyze scans the code_index and infers conventions across all categories.
func (a *Analyzer) Analyze(ctx context.Context) (*AnalysisResult, error) {
	start := time.Now()
	var conventions []Convention

	naming, err := a.analyzeNaming(ctx)
	if err == nil {
		conventions = append(conventions, naming...)
	}

	testing, err := a.analyzeTesting(ctx)
	if err == nil {
		conventions = append(conventions, testing...)
	}

	docs, err := a.analyzeDocumentation(ctx)
	if err == nil {
		conventions = append(conventions, docs...)
	}

	return &AnalysisResult{
		Conventions: conventions,
		Duration:    time.Since(start),
	}, nil
}

// GetConventions retrieves stored conventions, optionally filtered.
func (a *Analyzer) GetConventions(ctx context.Context, language string, category string) ([]Convention, error) {
	return GetConventions(a.store, language, category)
}

func (a *Analyzer) analyzeNaming(ctx context.Context) ([]Convention, error) {
	db := a.store.DB()
	var conventions []Convention

	// Query function/method names per language
	rows, err := db.QueryContext(ctx,
		`SELECT language, symbol_name FROM code_index
		 WHERE symbol_type IN ('function', 'method')
		 AND symbol_name NOT LIKE '%import%'
		 AND symbol_name != ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	namesByLang := make(map[string][]string)
	for rows.Next() {
		var lang, name string
		if err := rows.Scan(&lang, &name); err != nil {
			continue
		}
		// Extract just the function name (strip class prefix)
		parts := []string{name}
		for _, sep := range []string{"::", ".", "#", ":"} {
			for i := len(name) - 1; i >= 0; i-- {
				if string(name[i]) == sep[:1] {
					parts = []string{name[i+len(sep):]}
					break
				}
			}
		}
		cleanName := parts[0]
		if cleanName != "" {
			namesByLang[lang] = append(namesByLang[lang], cleanName)
		}
	}

	for lang, names := range namesByLang {
		style, conf := DetectNamingStyle(names)
		if style == "" {
			continue
		}

		// Collect examples
		var examples []string
		count := 0
		for _, n := range names {
			if count >= 5 {
				break
			}
			switch style {
			case "snake_case":
				if IsSnakeCase(n) {
					examples = append(examples, n)
					count++
				}
			case "camelCase":
				if IsCamelCase(n) {
					examples = append(examples, n)
					count++
				}
			case "PascalCase":
				if IsPascalCase(n) {
					examples = append(examples, n)
					count++
				}
			}
		}

		conventions = append(conventions, Convention{
			Pattern:     style + " function names",
			Description: lang + " functions predominantly use " + style + " naming",
			Category:    "naming",
			Confidence:  conf,
			Examples:    examples,
			Language:    lang,
		})
	}

	return conventions, nil
}

func (a *Analyzer) analyzeTesting(ctx context.Context) ([]Convention, error) {
	db := a.store.DB()
	var conventions []Convention

	// Count tests and non-tests per language
	rows, err := db.QueryContext(ctx,
		`SELECT language,
		        SUM(CASE WHEN symbol_type = 'test' THEN 1 ELSE 0 END) as test_count,
		        COUNT(*) as total_count
		 FROM code_index
		 GROUP BY language
		 HAVING test_count > 0`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var lang string
		var testCount, totalCount int
		if err := rows.Scan(&lang, &testCount, &totalCount); err != nil {
			continue
		}

		ratio := float64(testCount) / float64(totalCount)
		conventions = append(conventions, Convention{
			Pattern:     "test coverage pattern",
			Description: lang + ": " + formatPercent(ratio) + " of symbols are tests",
			Category:    "testing",
			Confidence:  0.9, // factual observation
			Language:    lang,
		})
	}

	return conventions, nil
}

func (a *Analyzer) analyzeDocumentation(ctx context.Context) ([]Convention, error) {
	db := a.store.DB()
	var conventions []Convention

	rows, err := db.QueryContext(ctx,
		`SELECT language,
		        SUM(CASE WHEN docstring != '' THEN 1 ELSE 0 END) as documented,
		        COUNT(*) as total
		 FROM code_index
		 WHERE symbol_type IN ('function', 'method', 'class', 'type')
		 GROUP BY language
		 HAVING total > 5`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var lang string
		var documented, total int
		if err := rows.Scan(&lang, &documented, &total); err != nil {
			continue
		}

		coverage := float64(documented) / float64(total)
		if coverage >= 0.6 {
			conventions = append(conventions, Convention{
				Pattern:     "documentation coverage",
				Description: lang + ": " + formatPercent(coverage) + " of symbols have docstrings",
				Category:    "documentation",
				Confidence:  coverage,
				Language:    lang,
			})
		}
	}

	return conventions, nil
}

func formatPercent(f float64) string {
	return fmt.Sprintf("%.0f%%", f*100)
}
