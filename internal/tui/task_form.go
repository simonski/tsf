package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonski/task/internal/cli"
)

type taskFormModel struct {
	task          *Task
	titleInput    textinput.Model
	descInput     textinput.Model
	projectInput  textinput.Model
	statusInput   textinput.Model
	priorityInput textinput.Model
	focusIndex    int
	submitting    bool
}

func newTaskFormModel(task *Task) taskFormModel {
	titleInput := textinput.New()
	titleInput.Placeholder = "Task title"
	titleInput.CharLimit = 200
	titleInput.Width = 50

	descInput := textinput.New()
	descInput.Placeholder = "Description"
	descInput.CharLimit = 1000
	descInput.Width = 50

	projectInput := textinput.New()
	projectInput.Placeholder = "Project ID"
	projectInput.CharLimit = 50
	projectInput.Width = 50

	statusInput := textinput.New()
	statusInput.Placeholder = "Status (open/in_progress/done)"
	statusInput.CharLimit = 20
	statusInput.Width = 50

	priorityInput := textinput.New()
	priorityInput.Placeholder = "Priority (0-10)"
	priorityInput.CharLimit = 2
	priorityInput.Width = 50

	if task != nil {
		titleInput.SetValue(task.Title)
		descInput.SetValue(task.Description)
		projectInput.SetValue(task.ProjectID)
		statusInput.SetValue(task.Status)
		priorityInput.SetValue(fmt.Sprintf("%d", task.Priority))
	} else {
		statusInput.SetValue("open")
		priorityInput.SetValue("5")
		projectInput.SetValue("default")
	}

	titleInput.Focus()

	return taskFormModel{
		task:          task,
		titleInput:    titleInput,
		descInput:     descInput,
		projectInput:  projectInput,
		statusInput:   statusInput,
		priorityInput: priorityInput,
		focusIndex:    0,
	}
}

func (m taskFormModel) Update(msg tea.Msg, client *cli.Client) (taskFormModel, tea.Cmd) {
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
				return changeViewMsg{view: ViewTasks}
			}
		}
	}

	var cmd tea.Cmd
	switch m.focusIndex {
	case 0:
		m.titleInput, cmd = m.titleInput.Update(msg)
	case 1:
		m.descInput, cmd = m.descInput.Update(msg)
	case 2:
		m.projectInput, cmd = m.projectInput.Update(msg)
	case 3:
		m.statusInput, cmd = m.statusInput.Update(msg)
	case 4:
		m.priorityInput, cmd = m.priorityInput.Update(msg)
	}

	return m, cmd
}

func (m taskFormModel) updateFocus() (taskFormModel, tea.Cmd) {
	inputs := []*textinput.Model{
		&m.titleInput,
		&m.descInput,
		&m.projectInput,
		&m.statusInput,
		&m.priorityInput,
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

func (m taskFormModel) submit(client *cli.Client) (taskFormModel, tea.Cmd) {
	title := strings.TrimSpace(m.titleInput.Value())
	desc := strings.TrimSpace(m.descInput.Value())
	projectID := strings.TrimSpace(m.projectInput.Value())
	status := strings.TrimSpace(m.statusInput.Value())
	priority, _ := strconv.Atoi(strings.TrimSpace(m.priorityInput.Value()))

	if title == "" || projectID == "" {
		return m, func() tea.Msg {
			return errorMsg{err: &ValidationError{"Title and Project ID are required"}}
		}
	}

	payload := map[string]interface{}{
		"title":       title,
		"description": desc,
		"project_id":  projectID,
		"status":      status,
		"priority":    priority,
	}

	m.submitting = true

	return m, func() tea.Msg {
		var err error
		if m.task != nil {
			_, err = client.Request("PUT", fmt.Sprintf("/api/v1/tasks/%s", m.task.ID), payload)
		} else {
			_, err = client.Request("POST", "/api/v1/tasks", payload)
		}

		if err != nil {
			return errorMsg{err: err}
		}

		return changeViewMsg{view: ViewTasks}
	}
}

func (m taskFormModel) View(width, height int) string {
	var b strings.Builder

	title := "New Task"
	if m.task != nil {
		title = "Edit Task"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	fields := []string{
		"Title:",
		m.titleInput.View(),
		"",
		"Description:",
		m.descInput.View(),
		"",
		"Project ID:",
		m.projectInput.View(),
		"",
		"Status:",
		m.statusInput.View(),
		"",
		"Priority:",
		m.priorityInput.View(),
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
