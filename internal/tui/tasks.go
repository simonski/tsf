package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonski/task/internal/cli"
)

type tasksModel struct {
	tasks   []Task
	cursor  int
	loading bool
}

func newTasksModel() tasksModel {
	return tasksModel{loading: true}
}

func (m tasksModel) loadTasks(client *cli.Client) tea.Cmd {
	return func() tea.Msg {
		data, err := client.Request("GET", "/api/v1/tasks", nil)
		if err != nil {
			return errorMsg{err: err}
		}
		var tasks []Task
		if err := json.Unmarshal(data, &tasks); err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: tasks}
	}
}

func (m tasksModel) Update(msg tea.Msg, client *cli.Client) (tasksModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		if tasks, ok := msg.data.([]Task); ok {
			m.tasks = tasks
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
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.tasks) > 0 {
				return m, func() tea.Msg {
					return changeViewMsg{view: ViewTaskForm, data: &m.tasks[m.cursor]}
				}
			}
		case "n":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewTaskForm}
			}
		case "d":
			if len(m.tasks) > 0 {
				return m, m.deleteTask(client, m.tasks[m.cursor].ID)
			}
		case "r":
			m.loading = true
			return m, m.loadTasks(client)
		case "esc", "backspace":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewMainMenu}
			}
		}
	}
	return m, nil
}

func (m tasksModel) deleteTask(client *cli.Client, taskID string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Request("DELETE", fmt.Sprintf("/api/v1/tasks/%s", taskID), nil)
		if err != nil {
			return errorMsg{err: err}
		}
		return changeViewMsg{view: ViewTasks}
	}
}

func (m tasksModel) View(width, height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Tasks"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(subtitleStyle.Render("Loading tasks..."))
	} else if len(m.tasks) == 0 {
		b.WriteString(subtitleStyle.Render("No tasks found. Press 'n' to create one."))
	} else {
		headers := []string{"Title", "Status", "Priority", "Project"}
		colWidths := []int{30, 12, 8, 30}

		headerRow := ""
		for i, h := range headers {
			headerRow += tableHeaderStyle.Width(colWidths[i]).Render(h)
		}
		b.WriteString(headerRow)
		b.WriteString("\n")

		for i, task := range m.tasks {
			cursor := "  "
			style := tableCellStyle
			if m.cursor == i {
				cursor = "→ "
				style = selectedStyle
			}

			row := cursor
			row += style.Width(colWidths[0] - 2).Render(truncate(task.Title, colWidths[0]-2))
			row += style.Width(colWidths[1]).Render(task.Status)
			row += style.Width(colWidths[2]).Render(fmt.Sprintf("%d", task.Priority))
			row += style.Width(colWidths[3]).Render(task.ProjectID)

			b.WriteString(row)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("n: New • Enter: Edit • d: Delete • r: Refresh • Backspace: Menu"))
	return b.String()
}
