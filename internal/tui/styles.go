package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	primaryColor   = lipgloss.Color("12")  // Blue
	secondaryColor = lipgloss.Color("10")  // Green
	errorColor     = lipgloss.Color("9")   // Red
	warningColor   = lipgloss.Color("11")  // Yellow
	subtleColor    = lipgloss.Color("241") // Gray
	highlightColor = lipgloss.Color("13")  // Magenta

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Subtitle style
	subtitleStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			Italic(true)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			MarginTop(1)

	// Selected item style
	selectedStyle = lipgloss.NewStyle().
			Foreground(highlightColor).
			Bold(true)

	// Regular item style
	itemStyle = lipgloss.NewStyle()

	// Box style for forms
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// Input field style
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(subtleColor)

	// Button styles
	activeButtonStyle = lipgloss.NewStyle().
				Background(primaryColor).
				Foreground(lipgloss.Color("0")).
				Padding(0, 2).
				Bold(true)

	inactiveButtonStyle = lipgloss.NewStyle().
				Background(subtleColor).
				Foreground(lipgloss.Color("15")).
				Padding(0, 2)

	// Table header style
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(primaryColor)

	// Table cell style
	tableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Status badges
	statusActiveStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	statusInactiveStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	// Error message style
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Padding(1)
)

// Common help text
const (
	navigationHelp = "↑/↓ or w/s: Navigate • Space: Select • Backspace/Esc: Back • Ctrl-C twice: Quit"
	formHelp       = "Tab: Next field • Shift-Tab: Previous field • Enter: Submit • Esc: Cancel"
)
