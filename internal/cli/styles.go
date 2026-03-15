package cli

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	Primary   = lipgloss.Color("#7C3AED") // Purple
	Secondary = lipgloss.Color("#06B6D4") // Cyan
	Success   = lipgloss.Color("#22C55E") // Green
	Warning   = lipgloss.Color("#EAB308") // Yellow
	Error     = lipgloss.Color("#EF4444") // Red
	Muted     = lipgloss.Color("#6B7280") // Gray
)

// Styles
var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary)

	Subtitle = lipgloss.NewStyle().
			Foreground(Secondary)

	Label = lipgloss.NewStyle().
		Foreground(Muted)

	Value = lipgloss.NewStyle().
		Bold(true)

	FilePath = lipgloss.NewStyle().
			Foreground(Secondary).
			Underline(true)

	SymbolName = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	SuccessText = lipgloss.NewStyle().
			Foreground(Success)

	ErrorText = lipgloss.NewStyle().
			Foreground(Error)

	WarningText = lipgloss.NewStyle().
			Foreground(Warning)
)
