package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/TomOst-Sec/colony-project/internal/storage"
)

// ArchitectureModel displays project module structure.
type ArchitectureModel struct {
	store *storage.Store

	modules  []module
	cursor   int
	expanded int // index of expanded module, -1 for none
	showDeps bool

	width, height int
}

type module struct {
	name         string
	path         string
	description  string
	dependencies []string
	exports      []string
	complexity   float64
	files        int
	symbols      int
}

// NewArchitectureModel creates a new architecture panel.
func NewArchitectureModel(store *storage.Store) ArchitectureModel {
	return ArchitectureModel{
		store:    store,
		expanded: -1,
	}
}

// Init loads architecture data.
func (m ArchitectureModel) Init() tea.Cmd {
	return m.loadModules
}

type modulesLoaded struct {
	modules []module
}

func (m ArchitectureModel) loadModules() tea.Msg {
	db := m.store.DB()

	// Query architecture table for modules
	rows, err := db.Query(
		`SELECT module_name, module_path, description, dependencies, exports, complexity_score
		 FROM architecture ORDER BY module_path`,
	)
	if err != nil {
		// Fall back to detecting modules from code_index
		return m.detectModulesFromIndex()
	}
	defer rows.Close()

	var mods []module
	for rows.Next() {
		var mod module
		var depsJSON, exportsJSON string
		if rows.Scan(&mod.name, &mod.path, &mod.description, &depsJSON, &exportsJSON, &mod.complexity) == nil {
			json.Unmarshal([]byte(depsJSON), &mod.dependencies)
			json.Unmarshal([]byte(exportsJSON), &mod.exports)
			mods = append(mods, mod)
		}
	}

	if len(mods) == 0 {
		return m.detectModulesFromIndex()
	}

	return modulesLoaded{modules: mods}
}

func (m ArchitectureModel) detectModulesFromIndex() modulesLoaded {
	db := m.store.DB()

	rows, err := db.Query(
		`SELECT
			CASE
				WHEN instr(file_path, '/') > 0
				THEN substr(file_path, 1, instr(substr(file_path, instr(file_path, '/')+1), '/')+instr(file_path, '/')-1)
				ELSE file_path
			END as module,
			COUNT(DISTINCT file_path) as file_count,
			COUNT(*) as symbol_count
		 FROM code_index
		 GROUP BY module
		 ORDER BY module`,
	)
	if err != nil {
		return modulesLoaded{}
	}
	defer rows.Close()

	var mods []module
	for rows.Next() {
		var mod module
		if rows.Scan(&mod.path, &mod.files, &mod.symbols) == nil {
			mod.name = mod.path
			mod.complexity = float64(mod.symbols) / 100.0
			if mod.complexity > 10 {
				mod.complexity = 10
			}
			mods = append(mods, mod)
		}
	}

	return modulesLoaded{modules: mods}
}

// Update handles messages.
func (m ArchitectureModel) Update(msg tea.Msg) (ArchitectureModel, tea.Cmd) {
	switch msg := msg.(type) {
	case modulesLoaded:
		m.modules = msg.modules

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.modules)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Enter) || msg.String() == "e":
			if m.expanded == m.cursor {
				m.expanded = -1
			} else {
				m.expanded = m.cursor
				m.showDeps = false
			}
		case msg.String() == "d":
			if m.expanded == m.cursor {
				m.showDeps = !m.showDeps
			} else {
				m.expanded = m.cursor
				m.showDeps = true
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the architecture panel.
func (m ArchitectureModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Architecture"))
	b.WriteString("  ")
	b.WriteString(statusBarStyle.Render("[e] expand  [d] show deps"))
	b.WriteString("\n\n")

	if len(m.modules) == 0 {
		b.WriteString(labelStyle.Render("No modules detected. Run 'engram index' first."))
		b.WriteString("\n")
	} else {
		b.WriteString(labelStyle.Render(fmt.Sprintf("Modules (%d detected)", len(m.modules))))
		b.WriteString("\n\n")
	}

	for i, mod := range m.modules {
		prefix := "  "
		if i == m.cursor {
			prefix = selectedStyle.Render("> ")
		}

		desc := mod.description
		if desc == "" {
			desc = mod.name
		}
		if len(desc) > 20 {
			desc = desc[:17] + "..."
		}

		line := fmt.Sprintf("%s%-25s [%-12s]  complexity: %.0f", prefix, mod.path, desc, mod.complexity)
		b.WriteString(line)
		b.WriteString("\n")

		if i == m.expanded {
			if m.showDeps && len(mod.dependencies) > 0 {
				b.WriteString("    Dependencies:\n")
				for _, dep := range mod.dependencies {
					b.WriteString(fmt.Sprintf("      → %s\n", dep))
				}
			} else if !m.showDeps && len(mod.exports) > 0 {
				b.WriteString("    Exports:\n")
				limit := min(8, len(mod.exports))
				for _, exp := range mod.exports[:limit] {
					b.WriteString(fmt.Sprintf("      • %s\n", exp))
				}
				if len(mod.exports) > limit {
					b.WriteString(fmt.Sprintf("      ... and %d more\n", len(mod.exports)-limit))
				}
			}
		}
	}

	return panelStyle.Render(b.String())
}
