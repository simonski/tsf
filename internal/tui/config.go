package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/simonski/task/internal/cli"
)

type configModel struct {
	configs []Config
	cursor  int
	loading bool
}

func newConfigModel() configModel {
	return configModel{loading: true}
}

func (m configModel) loadConfigs(client *cli.Client) tea.Cmd {
	return func() tea.Msg {
		data, err := client.Request("GET", "/api/v1/config", nil)
		if err != nil {
			return errorMsg{err: err}
		}
		var configs []Config
		if err := json.Unmarshal(data, &configs); err != nil {
			return errorMsg{err: err}
		}
		return dataLoadedMsg{data: configs}
	}
}

func (m configModel) Update(msg tea.Msg, client *cli.Client) (configModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		if configs, ok := msg.data.([]Config); ok {
			m.configs = configs
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
			if m.cursor < len(m.configs)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.configs) > 0 {
				return m, func() tea.Msg {
					return changeViewMsg{view: ViewConfigForm, data: &m.configs[m.cursor]}
				}
			}
		case "n":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewConfigForm}
			}
		case "d":
			if len(m.configs) > 0 {
				return m, m.deleteConfig(client, m.configs[m.cursor].Key)
			}
		case "r":
			m.loading = true
			return m, m.loadConfigs(client)
		case "esc", "backspace":
			return m, func() tea.Msg {
				return changeViewMsg{view: ViewMainMenu}
			}
		}
	}
	return m, nil
}

func (m configModel) deleteConfig(client *cli.Client, key string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Request("DELETE", fmt.Sprintf("/api/v1/config/%s", key), nil)
		if err != nil {
			return errorMsg{err: err}
		}
		return changeViewMsg{view: ViewConfig}
	}
}

func (m configModel) View(width, height int) string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Configuration"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(subtitleStyle.Render("Loading configuration..."))
	} else if len(m.configs) == 0 {
		b.WriteString(subtitleStyle.Render("No config entries found. Press 'n' to create one."))
	} else {
		headers := []string{"Key", "Value", "Description"}
		colWidths := []int{25, 30, 35}

		headerRow := ""
		for i, h := range headers {
			headerRow += tableHeaderStyle.Width(colWidths[i]).Render(h)
		}
		b.WriteString(headerRow)
		b.WriteString("\n")

		for i, cfg := range m.configs {
			cursor := "  "
			style := tableCellStyle
			if m.cursor == i {
				cursor = "→ "
				style = selectedStyle
			}

			row := cursor
			row += style.Width(colWidths[0] - 2).Render(cfg.Key)
			row += style.Width(colWidths[1]).Render(truncate(cfg.Value, colWidths[1]))
			row += style.Width(colWidths[2]).Render(truncate(cfg.Description, colWidths[2]))

			b.WriteString(row)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("n: New • Enter: Edit • d: Delete • r: Refresh • Backspace: Menu"))
	return b.String()
}
