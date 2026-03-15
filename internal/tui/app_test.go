package tui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

func setupTestStore(t *testing.T) *storage.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })
	return store
}

func TestAppInitializesWithCorrectTabCount(t *testing.T) {
	store := setupTestStore(t)
	app := NewApp(store, "/test/repo")
	assert.Equal(t, 4, app.TabCount())
	assert.Equal(t, 0, app.ActiveTab())
}

func TestTabSwitchingForward(t *testing.T) {
	store := setupTestStore(t)
	app := NewApp(store, "/test/repo")

	// Tab forward
	msg := tea.KeyMsg{Type: tea.KeyTab}
	model, _ := app.Update(msg)
	app = model.(App)
	assert.Equal(t, 1, app.ActiveTab())

	model, _ = app.Update(msg)
	app = model.(App)
	assert.Equal(t, 2, app.ActiveTab())

	model, _ = app.Update(msg)
	app = model.(App)
	assert.Equal(t, 3, app.ActiveTab())

	// Wraps around
	model, _ = app.Update(msg)
	app = model.(App)
	assert.Equal(t, 0, app.ActiveTab())
}

func TestTabSwitchingBackward(t *testing.T) {
	store := setupTestStore(t)
	app := NewApp(store, "/test/repo")

	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	model, _ := app.Update(msg)
	app = model.(App)
	assert.Equal(t, 3, app.ActiveTab()) // wraps to last
}

func TestNumericTabJump(t *testing.T) {
	store := setupTestStore(t)
	app := NewApp(store, "/test/repo")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	model, _ := app.Update(msg)
	app = model.(App)
	assert.Equal(t, 1, app.ActiveTab())

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}}
	model, _ = app.Update(msg)
	app = model.(App)
	assert.Equal(t, 3, app.ActiveTab())
}

func TestStatusPanelRenders(t *testing.T) {
	store := setupTestStore(t)

	// Seed some data
	store.DB().Exec(
		`INSERT INTO code_index (file_path, file_hash, language, symbol_name, symbol_type, start_line, end_line)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"main.go", "hash1", "go", "main", "function", 1, 5,
	)

	m := NewStatusModel(store, "/test/repo")
	// Load stats directly
	stats := m.loadStats()
	m, _ = m.Update(stats)

	view := m.View()
	assert.Contains(t, view, "Index Status")
	assert.Contains(t, view, "/test/repo")
	assert.Contains(t, view, "1") // file count
}

func TestMemoriesPanelRenders(t *testing.T) {
	store := setupTestStore(t)

	// Seed memories
	store.DB().Exec(
		`INSERT INTO memories (content, type, tags) VALUES (?, ?, ?)`,
		"Decided to use SQLite", "decision", `["database"]`,
	)

	m := NewMemoriesModel(store)
	items := m.loadMemories()
	m, _ = m.Update(items)

	view := m.View()
	assert.Contains(t, view, "Memories")
	assert.Contains(t, view, "Decided to use SQLite")
	assert.Contains(t, view, "decision")
}

func TestMemoriesTypeFilter(t *testing.T) {
	store := setupTestStore(t)

	store.DB().Exec(`INSERT INTO memories (content, type) VALUES (?, ?)`, "Decision 1", "decision")
	store.DB().Exec(`INSERT INTO memories (content, type) VALUES (?, ?)`, "Bug fix 1", "bugfix")

	m := NewMemoriesModel(store)
	items := m.loadMemories()
	m, _ = m.Update(items)

	assert.Len(t, m.filtered, 2)

	m.typeFilter = "decision"
	m.applyFilter()
	assert.Len(t, m.filtered, 1)
	assert.Equal(t, "Decision 1", m.filtered[0].content)
}

func TestAppViewRenders(t *testing.T) {
	store := setupTestStore(t)
	app := NewApp(store, "/test/repo")

	view := app.View()
	assert.Contains(t, view, "Status")
	assert.Contains(t, view, "Memories")
	assert.Contains(t, view, "quit")
}
