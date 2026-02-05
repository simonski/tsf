package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonski/task/internal/cli"
)

type rolesModel struct {
	roles   []Role
	cursor  int
	loading bool
}

func newRolesModel() rolesModel {
	return rolesModel{loading: true}
}

func (m rolesModel) loadRoles(client *cli.Client) tea.Cmd {
	return func() tea.Msg {
		data, err := client.Request("GET", "/api/v1/roles", nil)
		if err != nil {
			return errorMsg{err: err}
		}
		var roles []Role
		if err := json.Unmarshal(data, &roles); err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: roles}
	}
}

func (m rolesModel) Update(msg tea.Msg, client *cli.Client) (rolesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		if roles, ok := msg.data.([]Role); ok {
			m.roles = roles
			m.loading = false
		}
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		switch msg.String() {
		case "up", "w":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "s":
			if m.cursor < len(m.roles)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.roles) > 0 {
				return m, func() tea.Msg {
					return changeViewMsg{view: ViewRoleForm, data: &m.roles[m.cursor]}
				}
			}
		case "n":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewRoleForm}
			}
		case "d":
			if len(m.roles) > 0 {
				return m, m.deleteRole(client, m.roles[m.cursor].ID)
			}
		case "r":
			m.loading = true
			return m, m.loadRoles(client)
		case "esc", "backspace":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewMainMenu}
			}
		}
	}
	return m, nil
}

func (m rolesModel) deleteRole(client *cli.Client, roleID string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Request("DELETE", fmt.Sprintf("/api/v1/roles/%s", roleID), nil)
		if err != nil {
			return errorMsg{err: err}
		}
		return changeViewMsg{view: ViewRoles}
	}
}

func (m rolesModel) View(width, height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Roles"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(subtitleStyle.Render("Loading roles..."))
	} else if len(m.roles) == 0 {
		b.WriteString(subtitleStyle.Render("No roles found. Press 'n' to create one."))
	} else {
		headers := []string{"Name", "Scope", "Active", "Description"}
		colWidths := []int{20, 12, 8, 40}

		headerRow := ""
		for i, h := range headers {
			headerRow += tableHeaderStyle.Width(colWidths[i]).Render(h)
		}
		b.WriteString(headerRow)
		b.WriteString("\n")

		for i, role := range m.roles {
			cursor := "  "
			style := tableCellStyle
			if m.cursor == i {
				cursor = "→ "
				style = selectedStyle
			}

			active := "No"
			if role.IsActive {
				active = "Yes"
			}

			row := cursor
			row += style.Width(colWidths[0] - 2).Render(truncate(role.Name, colWidths[0]-2))
			row += style.Width(colWidths[1]).Render(role.Scope)
			row += style.Width(colWidths[2]).Render(active)
			row += style.Width(colWidths[3]).Render(truncate(role.Description, colWidths[3]))

			b.WriteString(row)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("n: New • Enter: Edit • d: Delete • r: Refresh • Backspace: Menu"))
	return b.String()
}
