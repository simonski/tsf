package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonski/task/internal/cli"
)

type roleFormModel struct {
	role        *Role
	nameInput   textinput.Model
	descInput   textinput.Model
	rulesInput  textinput.Model
	scopeInput  textinput.Model
	activeInput textinput.Model
	focusIndex  int
	submitting  bool
}

func newRoleFormModel(role *Role) roleFormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Role name"
	nameInput.CharLimit = 100
	nameInput.Width = 50

	descInput := textinput.New()
	descInput.Placeholder = "Description"
	descInput.CharLimit = 500
	descInput.Width = 50

	rulesInput := textinput.New()
	rulesInput.Placeholder = "Rules/Instructions"
	rulesInput.CharLimit = 2000
	rulesInput.Width = 50

	scopeInput := textinput.New()
	scopeInput.Placeholder = "Scope (system/global/project)"
	scopeInput.CharLimit = 20
	scopeInput.Width = 50

	activeInput := textinput.New()
	activeInput.Placeholder = "Active (true/false)"
	activeInput.CharLimit = 5
	activeInput.Width = 50

	if role != nil {
		nameInput.SetValue(role.Name)
		descInput.SetValue(role.Description)
		rulesInput.SetValue(role.Rules)
		scopeInput.SetValue(role.Scope)
		activeInput.SetValue(fmt.Sprintf("%t", role.IsActive))
	} else {
		scopeInput.SetValue("global")
		activeInput.SetValue("true")
	}

	nameInput.Focus()

	return roleFormModel{
		role:        role,
		nameInput:   nameInput,
		descInput:   descInput,
		rulesInput:  rulesInput,
		scopeInput:  scopeInput,
		activeInput: activeInput,
		focusIndex:  0,
	}
}

func (m roleFormModel) Update(msg tea.Msg, client *cli.Client) (roleFormModel, tea.Cmd) {
	if m.submitting {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m.submit(client)
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % 5
			return m.updateFocus()
		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + 5) % 5
			return m.updateFocus()
		case "esc":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewRoles}
			}
		}
	}

	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case 1:
		m.descInput, cmd = m.descInput.Update(msg)
	case 2:
		m.rulesInput, cmd = m.rulesInput.Update(msg)
	case 3:
		m.scopeInput, cmd = m.scopeInput.Update(msg)
	case 4:
		m.activeInput, cmd = m.activeInput.Update(msg)
	}

	return m, cmd
}

func (m roleFormModel) updateFocus() (roleFormModel, tea.Cmd) {
	inputs := []*textinput.Model{
		&m.nameInput,
		&m.descInput,
		&m.rulesInput,
		&m.scopeInput,
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

func (m roleFormModel) submit(client *cli.Client) (roleFormModel, tea.Cmd) {
	name := strings.TrimSpace(m.nameInput.Value())
	desc := strings.TrimSpace(m.descInput.Value())
	rules := strings.TrimSpace(m.rulesInput.Value())
	scope := strings.TrimSpace(m.scopeInput.Value())
	active := strings.TrimSpace(m.activeInput.Value()) == "true"

	if name == "" {
		return m, func() tea.Msg {
			return errorMsg{err: &ValidationError{"Name is required"}}
		}
	}

	payload := map[string]interface{}{
		"name":        name,
		"description": desc,
		"rules":       rules,
		"scope":       scope,
		"is_active":   active,
	}

	m.submitting = true

	return m, func() tea.Msg {
		var err error
		if m.role != nil {
			_, err = client.Request("PUT", fmt.Sprintf("/api/v1/roles/%s", m.role.ID), payload)
		} else {
			_, err = client.Request("POST", "/api/v1/roles", payload)
		}

		if err != nil {
			return errorMsg{err: err}
		}

		return changeViewMsg{view: ViewRoles}
	}
}

func (m roleFormModel) View(width, height int) string {
	var b strings.Builder

	title := "New Role"
	if m.role != nil {
		title = "Edit Role"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	fields := []string{
		"Name:",
		m.nameInput.View(),
		"",
		"Description:",
		m.descInput.View(),
		"",
		"Rules:",
		m.rulesInput.View(),
		"",
		"Scope:",
		m.scopeInput.View(),
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
