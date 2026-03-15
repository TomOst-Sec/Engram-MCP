package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// ConventionsModel displays detected code conventions.
type ConventionsModel struct {
	store *storage.Store

	conventions    []convention
	filtered       []convention
	cursor         int
	languageFilter string
	categoryFilter string

	width, height int
}

type convention struct {
	pattern     string
	description string
	category    string
	confidence  float64
	examples    []string
	language    string
}

// NewConventionsModel creates a new conventions panel.
func NewConventionsModel(store *storage.Store) ConventionsModel {
	return ConventionsModel{store: store}
}

// Init loads conventions data.
func (m ConventionsModel) Init() tea.Cmd {
	return m.loadConventions
}

type conventionsLoaded struct {
	conventions []convention
}

func (m ConventionsModel) loadConventions() tea.Msg {
	db := m.store.DB()
	rows, err := db.Query(
		`SELECT pattern, description, category, confidence, examples, language
		 FROM conventions ORDER BY category, confidence DESC`,
	)
	if err != nil {
		return conventionsLoaded{}
	}
	defer rows.Close()

	var convs []convention
	for rows.Next() {
		var c convention
		var examplesJSON string
		if rows.Scan(&c.pattern, &c.description, &c.category, &c.confidence, &examplesJSON, &c.language) == nil {
			json.Unmarshal([]byte(examplesJSON), &c.examples)
			convs = append(convs, c)
		}
	}
	return conventionsLoaded{conventions: convs}
}

// Update handles messages.
func (m ConventionsModel) Update(msg tea.Msg) (ConventionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case conventionsLoaded:
		m.conventions = msg.conventions
		m.applyFilter()

	case tea.KeyMsg:
		switch {
		case msg.String() == "l":
			m.cycleLanguageFilter()
			m.applyFilter()
		case msg.String() == "c":
			m.cycleCategoryFilter()
			m.applyFilter()
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m *ConventionsModel) applyFilter() {
	m.filtered = nil
	for _, c := range m.conventions {
		if m.languageFilter != "" && c.language != m.languageFilter {
			continue
		}
		if m.categoryFilter != "" && c.category != m.categoryFilter {
			continue
		}
		m.filtered = append(m.filtered, c)
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m *ConventionsModel) cycleLanguageFilter() {
	langs := m.uniqueLanguages()
	langs = append([]string{""}, langs...)
	for i, l := range langs {
		if l == m.languageFilter {
			m.languageFilter = langs[(i+1)%len(langs)]
			return
		}
	}
	m.languageFilter = ""
}

func (m *ConventionsModel) cycleCategoryFilter() {
	cats := m.uniqueCategories()
	cats = append([]string{""}, cats...)
	for i, c := range cats {
		if c == m.categoryFilter {
			m.categoryFilter = cats[(i+1)%len(cats)]
			return
		}
	}
	m.categoryFilter = ""
}

func (m ConventionsModel) uniqueLanguages() []string {
	seen := make(map[string]bool)
	var result []string
	for _, c := range m.conventions {
		if c.language != "" && !seen[c.language] {
			seen[c.language] = true
			result = append(result, c.language)
		}
	}
	return result
}

func (m ConventionsModel) uniqueCategories() []string {
	seen := make(map[string]bool)
	var result []string
	for _, c := range m.conventions {
		if !seen[c.category] {
			seen[c.category] = true
			result = append(result, c.category)
		}
	}
	return result
}

// View renders the conventions panel.
func (m ConventionsModel) View() string {
	var b strings.Builder

	header := titleStyle.Render("Conventions")
	if m.languageFilter != "" {
		header += "  " + selectedStyle.Render(fmt.Sprintf("[%s]", m.languageFilter))
	}
	if m.categoryFilter != "" {
		header += "  " + selectedStyle.Render(fmt.Sprintf("(%s)", m.categoryFilter))
	}
	header += "  " + statusBarStyle.Render("[l] language  [c] category")
	b.WriteString(header)
	b.WriteString("\n\n")

	if len(m.filtered) == 0 {
		b.WriteString(labelStyle.Render("No conventions detected. Run 'engram index' first."))
		b.WriteString("\n")
	}

	currentCategory := ""
	for i, c := range m.filtered {
		if c.category != currentCategory {
			if currentCategory != "" {
				b.WriteString("\n")
			}
			catName := c.category
			if len(catName) > 0 {
				catName = strings.ToUpper(catName[:1]) + catName[1:]
			}
			b.WriteString(lipgloss.NewStyle().Bold(true).Render(catName))
			b.WriteString("\n")
			currentCategory = c.category
		}

		indicator := "○"
		if c.confidence >= 0.8 {
			indicator = "✓"
		}

		prefix := "  "
		if i == m.cursor {
			prefix = selectedStyle.Render("> ")
		}

		pct := int(c.confidence * 100)
		line := fmt.Sprintf("%s %s %-40s (%d%%)  [%s]", prefix, indicator, c.pattern, pct, c.language)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render(fmt.Sprintf("%d conventions", len(m.filtered))))

	return panelStyle.Render(b.String())
}
