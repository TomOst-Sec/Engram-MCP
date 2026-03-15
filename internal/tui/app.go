package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

var tabNames = []string{"Status", "Memories", "Conventions", "Architecture"}

// App is the main TUI model.
type App struct {
	store    *storage.Store
	repoRoot string

	tabs      []string
	activeTab int

	statusPanel   StatusModel
	memoriesPanel MemoriesModel

	width, height int
}

// NewApp creates a new TUI app.
func NewApp(store *storage.Store, repoRoot string) App {
	return App{
		store:         store,
		repoRoot:      repoRoot,
		tabs:          tabNames,
		statusPanel:   NewStatusModel(store, repoRoot),
		memoriesPanel: NewMemoriesModel(store),
	}
}

// TabCount returns the number of tabs.
func (a App) TabCount() int {
	return len(a.tabs)
}

// ActiveTab returns the active tab index.
func (a App) ActiveTab() int {
	return a.activeTab
}

// Init initializes all panels.
func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.statusPanel.Init(),
		a.memoriesPanel.Init(),
	)
}

// Update handles all messages.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tea.KeyMsg:
		// Global keys (unless a panel has focus for input)
		switch {
		case key.Matches(msg, keys.Quit) && !a.memoriesPanel.searching:
			return a, tea.Quit
		case key.Matches(msg, keys.NextTab):
			a.activeTab = (a.activeTab + 1) % len(a.tabs)
			return a, nil
		case key.Matches(msg, keys.PrevTab):
			a.activeTab = (a.activeTab - 1 + len(a.tabs)) % len(a.tabs)
			return a, nil
		case msg.String() == "1":
			a.activeTab = 0
			return a, nil
		case msg.String() == "2":
			a.activeTab = 1
			return a, nil
		case msg.String() == "3":
			a.activeTab = 2
			return a, nil
		case msg.String() == "4":
			a.activeTab = 3
			return a, nil
		}
	}

	// Route to active panel
	var cmd tea.Cmd
	switch a.activeTab {
	case 0:
		a.statusPanel, cmd = a.statusPanel.Update(msg)
	case 1:
		a.memoriesPanel, cmd = a.memoriesPanel.Update(msg)
	}

	return a, cmd
}

// View renders the entire TUI.
func (a App) View() string {
	var b strings.Builder

	// Tab bar
	for i, tab := range a.tabs {
		if i == a.activeTab {
			b.WriteString(activeTabStyle.Render(tab))
		} else {
			b.WriteString(inactiveTabStyle.Render(tab))
		}
		b.WriteString(" ")
	}
	b.WriteString("\n\n")

	// Active panel
	switch a.activeTab {
	case 0:
		b.WriteString(a.statusPanel.View())
	case 1:
		b.WriteString(a.memoriesPanel.View())
	case 2:
		b.WriteString(panelStyle.Render(labelStyle.Render("Conventions panel — coming soon")))
	case 3:
		b.WriteString(panelStyle.Render(labelStyle.Render("Architecture panel — coming soon")))
	}

	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("q: quit | tab: switch | 1-4: jump"))

	return b.String()
}
