package git

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// HistoryAnalyzer shells out to git log and git blame to extract
// decision context for every indexed file, storing results in the
// git_context table.
type HistoryAnalyzer struct {
	store    *storage.Store
	repoRoot string
}

// FileHistory holds git history context for a single file.
type FileHistory struct {
	FilePath          string
	LastAuthor        string
	LastCommitHash    string
	LastCommitMessage string
	LastModified      time.Time
	ChangeFrequency   int
	CoChangedFiles    []string
}

// AnalysisStats holds summary statistics from a full repository analysis.
type AnalysisStats struct {
	FilesAnalyzed    int
	FilesSkipped     int
	TotalCommits     int
	HottestFile      string
	HottestFrequency int
	Duration         time.Duration
}

// New creates a new HistoryAnalyzer.
func New(store *storage.Store, repoRoot string) *HistoryAnalyzer {
	return &HistoryAnalyzer{
		store:    store,
		repoRoot: repoRoot,
	}
}

// AnalyzeAll scans git history for all indexed files and populates the git_context table.
func (h *HistoryAnalyzer) AnalyzeAll(ctx context.Context) (*AnalysisStats, error) {
	start := time.Now()
	stats := &AnalysisStats{}

	// Get distinct indexed file paths
	rows, err := h.store.DB().Query(`SELECT DISTINCT file_path FROM code_index ORDER BY file_path`)
	if err != nil {
		return nil, fmt.Errorf("querying indexed files: %w", err)
	}
	var filePaths []string
	for rows.Next() {
		var fp string
		if err := rows.Scan(&fp); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scanning file path: %w", err)
		}
		filePaths = append(filePaths, fp)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating indexed files: %w", err)
	}

	// Build commit cache to avoid redundant git diff-tree calls
	commitCache := make(map[string][]string) // commitHash -> files

	for _, fp := range filePaths {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		fh, err := h.AnalyzeFile(ctx, fp)
		if err != nil {
			stats.FilesSkipped++
			continue
		}
		if fh == nil || fh.ChangeFrequency == 0 {
			stats.FilesSkipped++
			continue
		}

		// Compute co-changed files using commit cache
		coChanged, uniqueCommits, err := h.computeCoChanged(ctx, fp, commitCache)
		if err == nil {
			fh.CoChangedFiles = coChanged
		}
		stats.TotalCommits += uniqueCommits

		if err := UpsertFileHistory(h.store, fh); err != nil {
			return nil, fmt.Errorf("storing history for %s: %w", fp, err)
		}

		stats.FilesAnalyzed++
		if fh.ChangeFrequency > stats.HottestFrequency {
			stats.HottestFrequency = fh.ChangeFrequency
			stats.HottestFile = fh.FilePath
		}
	}

	stats.Duration = time.Since(start)
	return stats, nil
}

// AnalyzeFile analyzes git history for a single file.
func (h *HistoryAnalyzer) AnalyzeFile(ctx context.Context, filePath string) (*FileHistory, error) {
	output, err := h.RunGit(ctx, "log", "--format=%H|%an|%s|%aI", "--follow", "--", filePath)
	if err != nil {
		return nil, fmt.Errorf("git log for %s: %w", filePath, err)
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return &FileHistory{FilePath: filePath}, nil
	}

	lines := strings.Split(output, "\n")
	fh := &FileHistory{
		FilePath:        filePath,
		ChangeFrequency: len(lines),
	}

	// Parse most recent commit (first line)
	if len(lines) > 0 {
		parts := strings.SplitN(lines[0], "|", 4)
		if len(parts) >= 4 {
			fh.LastCommitHash = parts[0]
			fh.LastAuthor = parts[1]
			fh.LastCommitMessage = parts[2]
			if t, err := time.Parse(time.RFC3339, parts[3]); err == nil {
				fh.LastModified = t
			}
		}
	}

	return fh, nil
}

// GetHotspots returns files sorted by change frequency (most-changed first).
func (h *HistoryAnalyzer) GetHotspots(ctx context.Context, limit int) ([]FileHistory, error) {
	return GetHotspots(h.store, limit)
}

// GetCoChangedFiles returns files that frequently change together with the given file.
func (h *HistoryAnalyzer) GetCoChangedFiles(ctx context.Context, filePath string, limit int) ([]string, error) {
	fh, err := GetFileHistory(h.store, filePath)
	if err != nil {
		return nil, err
	}
	if fh == nil {
		return nil, nil
	}
	if limit > 0 && len(fh.CoChangedFiles) > limit {
		return fh.CoChangedFiles[:limit], nil
	}
	return fh.CoChangedFiles, nil
}

// computeCoChanged finds files that frequently change alongside the given file.
// Uses a cache to avoid redundant git diff-tree calls.
func (h *HistoryAnalyzer) computeCoChanged(ctx context.Context, filePath string, cache map[string][]string) ([]string, int, error) {
	// Get commit hashes for this file, limited to last 100 commits
	output, err := h.RunGit(ctx, "log", "--format=%H", "-n", "100", "--follow", "--", filePath)
	if err != nil {
		return nil, 0, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return nil, 0, nil
	}

	hashes := strings.Split(output, "\n")
	coCount := make(map[string]int) // file -> co-occurrence count

	for _, hash := range hashes {
		hash = strings.TrimSpace(hash)
		if hash == "" {
			continue
		}

		files, ok := cache[hash]
		if !ok {
			diffOutput, err := h.RunGit(ctx, "diff-tree", "--no-commit-id", "--name-only", "-r", hash)
			if err != nil {
				continue
			}
			files = strings.Split(strings.TrimSpace(diffOutput), "\n")
			cache[hash] = files
		}

		for _, f := range files {
			f = strings.TrimSpace(f)
			if f != "" && f != filePath {
				coCount[f]++
			}
		}
	}

	// Sort by frequency, take top 10
	type pair struct {
		file  string
		count int
	}
	var pairs []pair
	for f, c := range coCount {
		pairs = append(pairs, pair{f, c})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})

	limit := min(10, len(pairs))

	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		result[i] = pairs[i].file
	}

	return result, len(hashes), nil
}
