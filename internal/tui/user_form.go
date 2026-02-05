package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonski/task/internal/cli"
)

type userFormModel struct {
	user          *User
	usernameInput textinput.Model
	passwordInput textinput.Model
	typeInput     textinput.Model
	activeInput   textinput.Model
	focusIndex    int
	submitting    bool
}

func newUserFormModel(user *User) userFormModel {
	usernameInput := textinput.New()
	usernameInput.Placeholder = "Username"
	usernameInput.CharLimit = 100
	usernameInput.Width = 50

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Password (leave empty to keep current)"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.CharLimit = 100
	passwordInput.Width = 50

	typeInput := textinput.New()
	typeInput.Placeholder = "Type (human/worker/orchestrator)"
	typeInput.CharLimit = 20
	typeInput.Width = 50

	activeInput := textinput.New()
	activeInput.Placeholder = "Active (true/false)"
	activeInput.CharLimit = 5
	activeInput.Width = 50

	if user != nil {
		usernameInput.SetValue(user.Username)
		typeInput.SetValue(user.Type)
		activeInput.SetValue(fmt.Sprintf("%t", user.IsActive))
	} else {
		typeInput.SetValue("human")
		activeInput.SetValue("true")
	}

	usernameInput.Focus()

	return userFormModel{
		user:          user,
		usernameInput: usernameInput,
		passwordInput: passwordInput,
		typeInput:     typeInput,
		activeInput:   activeInput,
		focusIndex:    0,
	}
}

func (m userFormModel) Update(msg tea.Msg, client *cli.Client) (userFormModel, tea.Cmd) {
	if m.submitting {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m.submit(client)
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % 4
			return m.updateFocus()
		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + 4) % 4
			return m.updateFocus()
		case "esc":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewUsers}
			}
		}
	}

	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.usernameInput, cmd = m.usernameInput.Update(msg)
	case 1:
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	case 2:
		m.typeInput, cmd = m.typeInput.Update(msg)
	case 3:
		m.activeInput, cmd = m.activeInput.Update(msg)
	}

	return m, cmd
}

func (m userFormModel) updateFocus() (userFormModel, tea.Cmd) {
	inputs := []*textinput.Model{
		&m.usernameInput,
		&m.passwordInput,
		&m.typeInput,
		&m.activeInput,
	}

	for i, input := range inputs {
		if i == m.focusIndex {
			input.Focus()
		} else {
			input.Blur()
		}
	}

	return m, nil
}

func (m userFormModel) submit(client *cli.Client) (userFormModel, tea.Cmd) {
	username := strings.TrimSpace(m.usernameInput.Value())
	password := m.passwordInput.Value()
	userType := strings.TrimSpace(m.typeInput.Value())
	active := strings.TrimSpace(m.activeInput.Value()) == "true"

	if username == "" {
		return m, func() tea.Msg {
			return errorMsg{err: &ValidationError{"Username is required"}}
		}
	}

	if m.user == nil && password == "" {
		return m, func() tea.Msg {
			return errorMsg{err: &ValidationError{"Password is required for new users"}}
		}
	}

	payload := map[string]interface{}{
		"username":  username,
		"type":      userType,
		"is_active": active,
	}

	if password != "" {
		payload["password"] = password
	}

	m.submitting = true

	return m, func() tea.Msg {
		var err error
		if m.user != nil {
			_, err = client.Request("PUT", fmt.Sprintf("/api/v1/users/%s", m.user.ID), payload)
		} else {
			_, err = client.Request("POST", "/api/v1/users", payload)
		}

		if err != nil {
			return errorMsg{err: err}
		}

		return changeViewMsg{view: ViewUsers}
	}
}

func (m userFormModel) View(width, height int) string {
	var b strings.Builder

	title := "New User"
	if m.user != nil {
		title = "Edit User"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	fields := []string{
		"Username:",
		m.usernameInput.View(),
		"",
		"Password:",
		m.passwordInput.View(),
		"",
		"Type:",
		m.typeInput.View(),
		"",
		"Active:",
		m.activeInput.View(),
	}

	form := boxStyle.Render(strings.Join(fields, "\n"))
	b.WriteString(form)
	b.WriteString("\n\n")

	if m.submitting {
		b.WriteString(subtitleStyle.Render("Submitting..."))
	} else {
		b.WriteString(helpStyle.Render(formHelp))
	}

	return b.String()
}
