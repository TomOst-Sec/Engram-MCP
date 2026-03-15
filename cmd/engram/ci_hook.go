package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/TomOst-Sec/colony-project/internal/cihook"
	"github.com/TomOst-Sec/colony-project/internal/config"
	"github.com/TomOst-Sec/colony-project/internal/storage"
	"github.com/spf13/cobra"
)

var (
	ciSource string
	ciType   string
	ciRunID  string
	ciFile   string
)

func newCIHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ci-hook",
		Short: "Store CI/CD output as memories",
		Long:  "Parse CI/CD build output and store failures, warnings, and deployment info as Engram memories.",
		RunE:  runCIHook,
	}
	cmd.Flags().StringVarP(&ciSource, "source", "s", "generic", "CI system: github-actions, gitlab-ci, generic")
	cmd.Flags().StringVarP(&ciType, "type", "t", "learning", "Memory type: bugfix, learning, decision")
	cmd.Flags().StringVar(&ciRunID, "run-id", "", "CI run identifier (for dedup)")
	cmd.Flags().StringVarP(&ciFile, "file", "f", "", "Read from file instead of stdin")
	return cmd
}

func runCIHook(cmd *cobra.Command, args []string) error {
	// Validate source
	switch ciSource {
	case "github-actions", "gitlab-ci", "generic":
		// ok
	default:
		return fmt.Errorf("unknown source %q: must be github-actions, gitlab-ci, or generic", ciSource)
	}

	// Open input
	var input io.Reader
	if ciFile != "" {
		f, err := os.Open(ciFile)
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer f.Close()
		input = f
	} else {
		input = os.Stdin
	}

	// Parse CI output
	var events []cihook.CIEvent
	var err error
	switch ciSource {
	case "github-actions":
		events, err = cihook.ParseGitHubActions(input)
	case "gitlab-ci":
		events, err = cihook.ParseGitLabCI(input)
	default:
		events, err = cihook.ParseGeneric(input)
	}
	if err != nil {
		return fmt.Errorf("parsing CI output: %w", err)
	}

	if len(events) == 0 {
		fmt.Fprintln(os.Stderr, "No CI events detected.")
		return nil
	}

	// Open database
	repoRoot, err := detectRepoRoot()
	if err != nil {
		return fmt.Errorf("could not detect repository root: %w", err)
	}
	cfg, err := config.Load(repoRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	store, err := storage.Open(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer store.Close()

	// Check for dedup
	if ciRunID != "" {
		var count int
		err := store.DB().QueryRow("SELECT COUNT(*) FROM memories WHERE tags LIKE ?",
			"%ci-run:"+ciRunID+"%").Scan(&count)
		if err == nil && count > 0 {
			fmt.Fprintf(os.Stderr, "CI run %s already recorded (%d memories). Skipping.\n", ciRunID, count)
			return nil
		}
	}

	// Store events as memories
	counts := map[string]int{}
	for _, event := range events {
		tags := strings.Join(event.Tags, ",")
		if ciRunID != "" {
			tags += ",ci-run:" + ciRunID
		}
		files := strings.Join(event.Files, ",")

		_, err := store.DB().Exec(
			`INSERT INTO memories (content, summary, type, tags, related_files) VALUES (?, ?, ?, ?, ?)`,
			event.Details, event.Summary, ciType, tags, files,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to store memory: %v\n", err)
			continue
		}
		counts[event.Type]++
	}

	// Print summary
	total := 0
	parts := []string{}
	for typ, count := range counts {
		total += count
		parts = append(parts, fmt.Sprintf("%d %ss", count, typ))
	}
	fmt.Fprintf(os.Stderr, "Stored %d CI memories (%s)\n", total, strings.Join(parts, ", "))

	return nil
}
