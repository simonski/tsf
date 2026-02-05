package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonski/task/internal/cli"
)

type projectsModel struct {
	projects []Project
	cursor   int
	loading  bool
}

func newProjectsModel() projectsModel {
	return projectsModel{
		projects: []Project{},
		loading:  true,
	}
}

func (m projectsModel) loadProjects(client *cli.Client) tea.Cmd {
	return func() tea.Msg {
		data, err := client.Request("GET", "/api/v1/projects", nil)
		if err != nil {
			return errorMsg{err: err}
		}

		var projects []Project
		if err := json.Unmarshal(data, &projects); err != nil {
			return errorMsg{err: err}
		}

		return dataLoadedMsg{data: projects}
	}
}

func (m projectsModel) Update(msg tea.Msg, client *cli.Client) (projectsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		if projects, ok := msg.data.([]Project); ok {
			m.projects = projects
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
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case "enter", " ":
			// Edit selected project
			if len(m.projects) > 0 {
				return m, func() tea.Msg {
					return changeViewMsg{
						view: ViewProjectForm,
						data: &m.projects[m.cursor],
					}
				}
			}
		case "n":
			// New project
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewProjectForm}
			}
		case "d":
			// Delete selected project
			if len(m.projects) > 0 {
				return m, m.deleteProject(client, m.projects[m.cursor].ID)
			}
		case "r":
			// Refresh
			m.loading = true
			return m, m.loadProjects(client)
		case "esc", "backspace":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewMainMenu}
			}
		}
	}

	return m, nil
}

func (m projectsModel) deleteProject(client *cli.Client, projectID string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Request("DELETE", fmt.Sprintf("/api/v1/projects/%s", projectID), nil)
		if err != nil {
			return errorMsg{err: err}
		}
		// Reload projects
		return changeViewMsg{view: ViewProjects}
	}
}

func (m projectsModel) View(width, height int) string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Projects")
	b.WriteString(title)
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(subtitleStyle.Render("Loading projects..."))
	} else if len(m.projects) == 0 {
		b.WriteString(subtitleStyle.Render("No projects found. Press 'n' to create one."))
	} else {
		// Render table
		headers := []string{"Name", "Status", "Visibility", "Description"}
		colWidths := []int{20, 12, 12, 40}

		// Header row
		headerRow := ""
		for i, h := range headers {
			headerRow += tableHeaderStyle.Width(colWidths[i]).Render(h)
		}
		b.WriteString(headerRow)
		b.WriteString("\n")

		// Data rows
		for i, project := range m.projects {
			cursor := "  "
			style := tableCellStyle
			if m.cursor == i {
				cursor = "→ "
				style = selectedStyle
			}

			row := cursor
			row += style.Width(colWidths[0] - 2).Render(truncate(project.Name, colWidths[0]-2))
			row += style.Width(colWidths[1]).Render(project.Status)
			row += style.Width(colWidths[2]).Render(project.Visibility)
			row += style.Width(colWidths[3]).Render(truncate(project.Description, colWidths[3]))

			b.WriteString(row)
			b.WriteString("\n")
		}
	}

	// Help
	b.WriteString("\n")
	help := helpStyle.Render("n: New • Enter: Edit • d: Delete • r: Refresh • Backspace: Menu")
	b.WriteString(help)

	return b.String()
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
