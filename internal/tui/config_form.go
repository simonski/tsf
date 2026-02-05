package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonski/task/internal/cli"
)

type configFormModel struct {
	config     *Config
	keyInput   textinput.Model
	valueInput textinput.Model
	descInput  textinput.Model
	focusIndex int
	submitting bool
}

func newConfigFormModel(config *Config) configFormModel {
	keyInput := textinput.New()
	keyInput.Placeholder = "Configuration key"
	keyInput.CharLimit = 100
	keyInput.Width = 50

	valueInput := textinput.New()
	valueInput.Placeholder = "Value"
	valueInput.CharLimit = 1000
	valueInput.Width = 50

	descInput := textinput.New()
	descInput.Placeholder = "Description"
	descInput.CharLimit = 500
	descInput.Width = 50

	if config != nil {
		keyInput.SetValue(config.Key)
		valueInput.SetValue(config.Value)
		descInput.SetValue(config.Description)
	}

	keyInput.Focus()

	return configFormModel{
		config:     config,
		keyInput:   keyInput,
		valueInput: valueInput,
		descInput:  descInput,
		focusIndex: 0,
	}
}

func (m configFormModel) Update(msg tea.Msg, client *cli.Client) (configFormModel, tea.Cmd) {
	if m.submitting {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m.submit(client)
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % 3
			return m.updateFocus()
		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + 3) % 3
			return m.updateFocus()
		case "esc":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewConfig}
			}
		}
	}

	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.keyInput, cmd = m.keyInput.Update(msg)
	case 1:
		m.valueInput, cmd = m.valueInput.Update(msg)
	case 2:
		m.descInput, cmd = m.descInput.Update(msg)
	}

	return m, cmd
}

func (m configFormModel) updateFocus() (configFormModel, tea.Cmd) {
	inputs := []*textinput.Model{
		&m.keyInput,
		&m.valueInput,
		&m.descInput,
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

func (m configFormModel) submit(client *cli.Client) (configFormModel, tea.Cmd) {
	key := strings.TrimSpace(m.keyInput.Value())
	value := strings.TrimSpace(m.valueInput.Value())
	desc := strings.TrimSpace(m.descInput.Value())

	if key == "" {
		return m, func() tea.Msg {
			return errorMsg{err: &ValidationError{"Key is required"}}
		}
	}

	payload := map[string]interface{}{
		"key":         key,
		"value":       value,
		"description": desc,
	}

	m.submitting = true

	return m, func() tea.Msg {
		var err error
		if m.config != nil {
			_, err = client.Request("PUT", fmt.Sprintf("/api/v1/config/%s", m.config.Key), payload)
		} else {
			_, err = client.Request("POST", "/api/v1/config", payload)
		}

		if err != nil {
			return errorMsg{err: err}
		}

		return changeViewMsg{view: ViewConfig}
	}
}

func (m configFormModel) View(width, height int) string {
	var b strings.Builder

	title := "New Configuration"
	if m.config != nil {
		title = "Edit Configuration"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	fields := []string{
		"Key:",
		m.keyInput.View(),
		"",
		"Value:",
		m.valueInput.View(),
		"",
		"Description:",
		m.descInput.View(),
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
