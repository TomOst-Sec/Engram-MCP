package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/TomOst-Sec/colony-project/internal/tui"
)

func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive dashboard",
		Long:  "Launch an interactive terminal UI for browsing index status, memories, conventions, and architecture.",
		RunE:  runTUI,
	}
}

func runTUI(cmd *cobra.Command, args []string) error {
	store, repoRoot, _, err := openDatabase()
	if err != nil {
		return err
	}
	defer store.Close()

	app := tui.NewApp(store, repoRoot)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}
