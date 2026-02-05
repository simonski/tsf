package tui

import (
	"encoding/json"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonski/task/internal/cli"
)

type workersModel struct {
	workers []Worker
	cursor  int
	loading bool
}

func newWorkersModel() workersModel {
	return workersModel{loading: true}
}

func (m workersModel) loadWorkers(client *cli.Client) tea.Cmd {
	return func() tea.Msg {
		// Get heartbeats to show worker status
		data, err := client.Request("GET", "/api/v1/heartbeats", nil)
		if err != nil {
			return errorMsg{err: err}
		}

		var heartbeats []map[string]interface{}
		if err := json.Unmarshal(data, &heartbeats); err != nil {
			return errorMsg{err: err}
		}

		// Convert to Worker structs
		var workers []Worker
		for _, hb := range heartbeats {
			worker := Worker{
				ID:     hb["worker_id"].(string),
				Type:   hb["worker_type"].(string),
				Status: hb["status"].(string),
			}
			workers = append(workers, worker)
		}

		return dataLoadedMsg{data: workers}
	}
}

func (m workersModel) Update(msg tea.Msg, client *cli.Client) (workersModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		if workers, ok := msg.data.([]Worker); ok {
			m.workers = workers
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
			if m.cursor < len(m.workers)-1 {
				m.cursor++
			}
		case "r":
			m.loading = true
			return m, m.loadWorkers(client)
		case "esc", "backspace":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewMainMenu}
			}
		}
	}
	return m, nil
}

func (m workersModel) View(width, height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Workers & Orchestrators"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(subtitleStyle.Render("Loading workers..."))
	} else if len(m.workers) == 0 {
		b.WriteString(subtitleStyle.Render("No active workers found."))
	} else {
		headers := []string{"ID", "Type", "Status"}
		colWidths := []int{35, 20, 15}

		headerRow := ""
		for i, h := range headers {
			headerRow += tableHeaderStyle.Width(colWidths[i]).Render(h)
		}
		b.WriteString(headerRow)
		b.WriteString("\n")

		for i, worker := range m.workers {
			cursor := "  "
			style := tableCellStyle
			if m.cursor == i {
				cursor = "→ "
				style = selectedStyle
			}

			statusStyle := style
			if worker.Status == "active" {
				statusStyle = statusActiveStyle
			} else {
				statusStyle = statusInactiveStyle
			}

			row := cursor
			row += style.Width(colWidths[0] - 2).Render(truncate(worker.ID, colWidths[0]-2))
			row += style.Width(colWidths[1]).Render(worker.Type)
			row += statusStyle.Width(colWidths[2]).Render(worker.Status)

			b.WriteString(row)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("r: Refresh • Backspace: Menu"))
	return b.String()
}
