package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonski/task/internal/cli"
)

type projectFormModel struct {
	project     *Project
	nameInput   textinput.Model
	descInput   textinput.Model
	repoInput   textinput.Model
	statusInput textinput.Model
	visInput    textinput.Model
	focusIndex  int
	submitting  bool
}

func newProjectFormModel(project *Project) projectFormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Project name"
	nameInput.CharLimit = 100
	nameInput.Width = 50

	descInput := textinput.New()
	descInput.Placeholder = "Description"
	descInput.CharLimit = 500
	descInput.Width = 50

	repoInput := textinput.New()
	repoInput.Placeholder = "Repository URL (optional)"
	repoInput.CharLimit = 200
	repoInput.Width = 50

	statusInput := textinput.New()
	statusInput.Placeholder = "Status (active/inactive)"
	statusInput.CharLimit = 20
	statusInput.Width = 50

	visInput := textinput.New()
	visInput.Placeholder = "Visibility (public/internal/private)"
	visInput.CharLimit = 20
	visInput.Width = 50

	// Pre-fill if editing
	if project != nil {
		nameInput.SetValue(project.Name)
		descInput.SetValue(project.Description)
		if project.Repository != nil {
			repoInput.SetValue(*project.Repository)
		}
		statusInput.SetValue(project.Status)
		visInput.SetValue(project.Visibility)
	} else {
		statusInput.SetValue("active")
		visInput.SetValue("private")
	}

	nameInput.Focus()

	return projectFormModel{
		project:     project,
		nameInput:   nameInput,
		descInput:   descInput,
		repoInput:   repoInput,
		statusInput: statusInput,
		visInput:    visInput,
		focusIndex:  0,
	}
}

func (m projectFormModel) Update(msg tea.Msg, client *cli.Client) (projectFormModel, tea.Cmd) {
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
				return changeViewMsg{view: ViewProjects}
			}
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case 1:
		m.descInput, cmd = m.descInput.Update(msg)
	case 2:
		m.repoInput, cmd = m.repoInput.Update(msg)
	case 3:
		m.statusInput, cmd = m.statusInput.Update(msg)
	case 4:
		m.visInput, cmd = m.visInput.Update(msg)
	}

	return m, cmd
}

func (m projectFormModel) updateFocus() (projectFormModel, tea.Cmd) {
	inputs := []*textinput.Model{
		&m.nameInput,
		&m.descInput,
		&m.repoInput,
		&m.statusInput,
		&m.visInput,
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

func (m projectFormModel) submit(client *cli.Client) (projectFormModel, tea.Cmd) {
	name := strings.TrimSpace(m.nameInput.Value())
	desc := strings.TrimSpace(m.descInput.Value())
	repo := strings.TrimSpace(m.repoInput.Value())
	status := strings.TrimSpace(m.statusInput.Value())
	vis := strings.TrimSpace(m.visInput.Value())

	if name == "" {
		return m, func() tea.Msg {
			return errorMsg{err: &ValidationError{"Name is required"}}
		}
	}

	payload := map[string]interface{}{
		"name":        name,
		"description": desc,
		"status":      status,
		"visibility":  vis,
	}

	if repo != "" {
		payload["repository"] = repo
	}

	m.submitting = true

	return m, func() tea.Msg {
		var err error
		if m.project != nil {
			// Update existing
			_, err = client.Request("PUT", fmt.Sprintf("/api/v1/projects/%s", m.project.ID), payload)
		} else {
			// Create new
			_, err = client.Request("POST", "/api/v1/projects", payload)
		}

		if err != nil {
			return errorMsg{err: err}
		}

		return changeViewMsg{view: ViewProjects}
	}
}

func (m projectFormModel) View(width, height int) string {
	var b strings.Builder

	// Title
	title := "New Project"
	if m.project != nil {
		title = "Edit Project"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Form fields
	fields := []string{
		"Name:",
		m.nameInput.View(),
		"",
		"Description:",
		m.descInput.View(),
		"",
		"Repository:",
		m.repoInput.View(),
		"",
		"Status:",
		m.statusInput.View(),
		"",
		"Visibility:",
		m.visInput.View(),
	}

	form := boxStyle.Render(strings.Join(fields, "\n"))
	b.WriteString(form)
	b.WriteString("\n\n")

	// Help
	if m.submitting {
		b.WriteString(subtitleStyle.Render("Submitting..."))
	} else {
		help := helpStyle.Render(formHelp)
		b.WriteString(help)
	}

	return b.String()
}
