package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mainMenuModel struct {
	cursor int
	items  []menuItem
}

type menuItem struct {
	title       string
	description string
	view        ViewType
}

func newMainMenuModel() mainMenuModel {
	items := []menuItem{
		{"Projects", "Manage projects and workspaces", ViewProjects},
		{"Tasks", "View and manage tasks", ViewTasks},
		{"Roles", "Manage roles and job descriptions", ViewRoles},
		{"Users", "Manage users and access", ViewUsers},
		{"Config", "System configuration", ViewConfig},
		{"Workers", "Monitor workers and orchestrators", ViewWorkers},
	}

	return mainMenuModel{
		cursor: 0,
		items:  items,
	}
}

func (m mainMenuModel) Update(msg tea.Msg) (mainMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "w":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "s":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter", " ":
			// Navigate to selected view
			return m, func() tea.Msg {
				return changeViewMsg{view: m.items[m.cursor].view}
			}
		case "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m mainMenuModel) View(width, height int) string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Main Menu")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Menu items
	for i, item := range m.items {
		cursor := "  "
		itemStr := item.title
		descStr := subtitleStyle.Render(item.description)

		if m.cursor == i {
			cursor = "→ "
			itemStr = selectedStyle.Render(item.title)
		} else {
			itemStr = itemStyle.Render(item.title)
		}

		b.WriteString(cursor + itemStr + "\n")
		b.WriteString("  " + descStr + "\n\n")
	}

	// Help
	help := helpStyle.Render(navigationHelp)
	b.WriteString("\n")
	b.WriteString(help)

	// Center the content
	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		b.String())
}
