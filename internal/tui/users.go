package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonski/task/internal/cli"
)

type usersModel struct {
	users   []User
	cursor  int
	loading bool
}

func newUsersModel() usersModel {
	return usersModel{loading: true}
}

func (m usersModel) loadUsers(client *cli.Client) tea.Cmd {
	return func() tea.Msg {
		data, err := client.Request("GET", "/api/v1/users", nil)
		if err != nil {
			return errorMsg{err: err}
		}
		var users []User
		if err := json.Unmarshal(data, &users); err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: users}
	}
}

func (m usersModel) Update(msg tea.Msg, client *cli.Client) (usersModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		if users, ok := msg.data.([]User); ok {
			m.users = users
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
			if m.cursor < len(m.users)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.users) > 0 {
				return m, func() tea.Msg {
					return changeViewMsg{view: ViewUserForm, data: &m.users[m.cursor]}
				}
			}
		case "n":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewUserForm}
			}
		case "d":
			if len(m.users) > 0 {
				return m, m.deleteUser(client, m.users[m.cursor].ID)
			}
		case "r":
			m.loading = true
			return m, m.loadUsers(client)
		case "esc", "backspace":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewMainMenu}
			}
		}
	}
	return m, nil
}

func (m usersModel) deleteUser(client *cli.Client, userID string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Request("DELETE", fmt.Sprintf("/api/v1/users/%s", userID), nil)
		if err != nil {
			return errorMsg{err: err}
		}
		return changeViewMsg{view: ViewUsers}
	}
}

func (m usersModel) View(width, height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Users"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(subtitleStyle.Render("Loading users..."))
	} else if len(m.users) == 0 {
		b.WriteString(subtitleStyle.Render("No users found. Press 'n' to create one."))
	} else {
		headers := []string{"Username", "Type", "Active"}
		colWidths := []int{25, 15, 10}

		headerRow := ""
		for i, h := range headers {
			headerRow += tableHeaderStyle.Width(colWidths[i]).Render(h)
		}
		b.WriteString(headerRow)
		b.WriteString("\n")

		for i, user := range m.users {
			cursor := "  "
			style := tableCellStyle
			if m.cursor == i {
				cursor = "→ "
				style = selectedStyle
			}

			active := "No"
			if user.IsActive {
				active = "Yes"
			}

			row := cursor
			row += style.Width(colWidths[0] - 2).Render(user.Username)
			row += style.Width(colWidths[1]).Render(user.Type)
			row += style.Width(colWidths[2]).Render(active)

			b.WriteString(row)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("n: New • Enter: Edit • d: Delete • r: Refresh • Backspace: Menu"))
	return b.String()
}
