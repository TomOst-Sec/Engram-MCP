package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// MemoriesModel provides an interactive memory browser.
type MemoriesModel struct {
	store *storage.Store

	memories []memoryItem
	filtered []memoryItem
	cursor   int

	searching bool
	search    textinput.Model
	typeFilter string

	width, height int
}

type memoryItem struct {
	id        int
	content   string
	memType   string
	createdAt string
}

// NewMemoriesModel creates a new memories panel.
func NewMemoriesModel(store *storage.Store) MemoriesModel {
	ti := textinput.New()
	ti.Placeholder = "Search memories..."
	ti.CharLimit = 100

	return MemoriesModel{
		store:  store,
		search: ti,
	}
}

// Init loads memories.
func (m MemoriesModel) Init() tea.Cmd {
	return m.loadMemories
}

type memoriesLoaded struct {
	memories []memoryItem
}

func (m MemoriesModel) loadMemories() tea.Msg {
	db := m.store.DB()

	rows, err := db.Query(
		`SELECT id, content, type, created_at FROM memories WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT 100`,
	)
	if err != nil {
		return memoriesLoaded{}
	}
	defer rows.Close()

	var items []memoryItem
	for rows.Next() {
		var mi memoryItem
		if rows.Scan(&mi.id, &mi.content, &mi.memType, &mi.createdAt) == nil {
			items = append(items, mi)
		}
	}
	return memoriesLoaded{memories: items}
}

// Update handles messages.
func (m MemoriesModel) Update(msg tea.Msg) (MemoriesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case memoriesLoaded:
		m.memories = msg.memories
		m.applyFilter()

	case tea.KeyMsg:
		if m.searching {
			switch {
			case key.Matches(msg, keys.Escape), key.Matches(msg, keys.Enter):
				m.searching = false
				m.search.Blur()
				m.applyFilter()
				return m, nil
			default:
				var cmd tea.Cmd
				m.search, cmd = m.search.Update(msg)
				m.applyFilter()
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, keys.Search):
			m.searching = true
			m.search.Focus()
			return m, m.search.Cursor.BlinkCmd()
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case msg.String() == "t":
			m.cycleTypeFilter()
			m.applyFilter()
		case key.Matches(msg, keys.Delete):
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				item := m.filtered[m.cursor]
				m.store.DB().Exec("UPDATE memories SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?", item.id)
				return m, m.loadMemories
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m *MemoriesModel) applyFilter() {
	query := strings.ToLower(m.search.Value())
	m.filtered = nil

	for _, mi := range m.memories {
		if m.typeFilter != "" && mi.memType != m.typeFilter {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(mi.content), query) {
			continue
		}
		m.filtered = append(m.filtered, mi)
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m *MemoriesModel) cycleTypeFilter() {
	types := []string{"", "decision", "bugfix", "refactor", "learning", "convention"}
	for i, t := range types {
		if t == m.typeFilter {
			m.typeFilter = types[(i+1)%len(types)]
			return
		}
	}
	m.typeFilter = ""
}

// View renders the memories panel.
func (m MemoriesModel) View() string {
	var b strings.Builder

	header := titleStyle.Render("Memories")
	if m.typeFilter != "" {
		header += "  " + selectedStyle.Render(fmt.Sprintf("[%s]", m.typeFilter))
	}
	header += "  " + statusBarStyle.Render("[/] search  [t] filter type  [d] delete")
	b.WriteString(header)
	b.WriteString("\n")

	if m.searching {
		b.WriteString(m.search.View())
		b.WriteString("\n")
	} else if m.search.Value() != "" {
		b.WriteString(labelStyle.Render(fmt.Sprintf("Search: %s", m.search.Value())))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if len(m.filtered) == 0 {
		b.WriteString(labelStyle.Render("No memories found."))
		b.WriteString("\n")
	}

	maxShow := 10
	if len(m.filtered) < maxShow {
		maxShow = len(m.filtered)
	}

	// Simple windowing around cursor
	start := 0
	if m.cursor >= maxShow {
		start = m.cursor - maxShow + 1
	}
	end := start + maxShow
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		mi := m.filtered[i]
		prefix := "  "
		if i == m.cursor {
			prefix = selectedStyle.Render("> ")
		}

		// Truncate content to first line, max 60 chars
		content := mi.content
		if idx := strings.IndexByte(content, '\n'); idx >= 0 {
			content = content[:idx]
		}
		if len(content) > 60 {
			content = content[:57] + "..."
		}

		line := fmt.Sprintf("%s#%d  [%s]  %s", prefix, mi.id, mi.memType, mi.createdAt)
		b.WriteString(line)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("   %s\n", content))
		b.WriteString("\n")
	}

	b.WriteString(statusBarStyle.Render(fmt.Sprintf("%d memories", len(m.filtered))))
	if m.typeFilter != "" {
		b.WriteString(statusBarStyle.Render(fmt.Sprintf(" | type: %s", m.typeFilter)))
	} else {
		b.WriteString(statusBarStyle.Render(" | showing all types"))
	}

	return panelStyle.Render(b.String())
}
