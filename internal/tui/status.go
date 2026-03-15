package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// StatusModel shows index statistics.
type StatusModel struct {
	store    *storage.Store
	repoRoot string

	fileCount       int
	symbolCount     int
	languages       []langStat
	embeddingCount  int
	memoryCount     int
	conventionCount int
	gitContextCount int
	lastIndexed     string

	width, height int
}

type langStat struct {
	lang  string
	count int
}

// NewStatusModel creates a new status panel.
func NewStatusModel(store *storage.Store, repoRoot string) StatusModel {
	return StatusModel{
		store:    store,
		repoRoot: repoRoot,
	}
}

// Init loads initial data.
func (m StatusModel) Init() tea.Cmd {
	return m.loadStats
}

type statsLoaded struct {
	fileCount       int
	symbolCount     int
	languages       []langStat
	embeddingCount  int
	memoryCount     int
	conventionCount int
	gitContextCount int
	lastIndexed     string
}

func (m StatusModel) loadStats() tea.Msg {
	db := m.store.DB()

	var stats statsLoaded

	db.QueryRow("SELECT COUNT(DISTINCT file_path), COUNT(*) FROM code_index").
		Scan(&stats.fileCount, &stats.symbolCount)

	rows, err := db.Query("SELECT language, COUNT(DISTINCT file_path) FROM code_index GROUP BY language ORDER BY COUNT(DISTINCT file_path) DESC")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ls langStat
			if rows.Scan(&ls.lang, &ls.count) == nil {
				stats.languages = append(stats.languages, ls)
			}
		}
	}

	db.QueryRow("SELECT COUNT(*) FROM code_index WHERE embedding IS NOT NULL").Scan(&stats.embeddingCount)
	db.QueryRow("SELECT COUNT(*) FROM memories WHERE deleted_at IS NULL").Scan(&stats.memoryCount)
	db.QueryRow("SELECT COUNT(*) FROM conventions").Scan(&stats.conventionCount)
	db.QueryRow("SELECT COUNT(*) FROM git_context").Scan(&stats.gitContextCount)

	var lastIdx *string
	db.QueryRow("SELECT MAX(updated_at) FROM code_index").Scan(&lastIdx)
	if lastIdx != nil {
		stats.lastIndexed = *lastIdx
	}

	return stats
}

// Update handles messages.
func (m StatusModel) Update(msg tea.Msg) (StatusModel, tea.Cmd) {
	switch msg := msg.(type) {
	case statsLoaded:
		m.fileCount = msg.fileCount
		m.symbolCount = msg.symbolCount
		m.languages = msg.languages
		m.embeddingCount = msg.embeddingCount
		m.memoryCount = msg.memoryCount
		m.conventionCount = msg.conventionCount
		m.gitContextCount = msg.gitContextCount
		m.lastIndexed = msg.lastIndexed
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the status panel.
func (m StatusModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Index Status"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Repository:"), m.repoRoot))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Files indexed:"), valueStyle.Render(fmt.Sprintf("%d", m.fileCount))))
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Symbols:     "), valueStyle.Render(fmt.Sprintf("%d", m.symbolCount))))

	if len(m.languages) > 0 {
		parts := make([]string, len(m.languages))
		for i, ls := range m.languages {
			parts[i] = fmt.Sprintf("%s (%d)", ls.lang, ls.count)
		}
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Languages:   "), strings.Join(parts, ", ")))
	}

	if m.symbolCount > 0 {
		pct := m.embeddingCount * 100 / m.symbolCount
		b.WriteString(fmt.Sprintf("%s %d / %d (%d%%)\n", labelStyle.Render("Embeddings:  "), m.embeddingCount, m.symbolCount, pct))
	}

	if m.lastIndexed != "" {
		b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Last indexed:"), m.lastIndexed))
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Memories:    "), valueStyle.Render(fmt.Sprintf("%d", m.memoryCount))))
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Conventions: "), valueStyle.Render(fmt.Sprintf("%d patterns", m.conventionCount))))
	b.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Git history: "), valueStyle.Render(fmt.Sprintf("%d files", m.gitContextCount))))

	return panelStyle.Render(b.String())
}
