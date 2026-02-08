package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/simonski/task/internal/cli"
)

type loginModel struct {
	config         *cli.Config
	usernameInput  textinput.Model
	passwordInput  textinput.Model
	focusIndex     int
	authenticating bool
}

func newLoginModel(config *cli.Config) loginModel {
	usernameInput := textinput.New()
	usernameInput.Placeholder = "Username"
	usernameInput.Focus()
	usernameInput.CharLimit = 100
	usernameInput.Width = 30

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.CharLimit = 100
	passwordInput.Width = 30

	// Pre-fill from config if available
	if config.Username != "" {
		usernameInput.SetValue(config.Username)
	}
	if config.Password != "" {
		passwordInput.SetValue(config.Password)
	}

	return loginModel{
		config:        config,
		usernameInput: usernameInput,
		passwordInput: passwordInput,
		focusIndex:    0,
	}
}

func (m loginModel) Update(msg tea.Msg) (loginModel, tea.Cmd) {
	if m.authenticating {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m.submit()
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % 2
			return m.updateFocus()
		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + 2) % 2
			return m.updateFocus()
		case "ctrl+r", "r":
			// Switch to register view (only when not typing)
			if m.focusIndex != 0 && m.usernameInput.Value() == "" {
				return m, func() tea.Msg {
					return switchToRegisterMsg{}
				}
			}
		case "esc":
			return m, tea.Quit
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	if m.focusIndex == 0 {
		m.usernameInput, cmd = m.usernameInput.Update(msg)
	} else {
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	}

	return m, cmd
}

func (m loginModel) updateFocus() (loginModel, tea.Cmd) {
	if m.focusIndex == 0 {
		m.usernameInput.Focus()
		m.passwordInput.Blur()
	} else {
		m.usernameInput.Blur()
		m.passwordInput.Focus()
	}
	return m, nil
}

func (m loginModel) submit() (loginModel, tea.Cmd) {
	username := strings.TrimSpace(m.usernameInput.Value())
	password := m.passwordInput.Value()

	if username == "" || password == "" {
		return m, func() tea.Msg {
			return errorMsg{err: &ValidationError{"Username and password are required"}}
		}
	}

	// Update config with credentials
	m.config.Username = username
	m.config.Password = password
	m.authenticating = true

	// Test authentication
	return m, func() tea.Msg {
		client := cli.NewClient(m.config)
		_, err := client.Request("GET", "/api/v1/auth/me", nil)
		if err != nil {
			return errorMsg{err: err}
		}
		return authSuccessMsg{}
	}
}

func (m loginModel) View(width, height int) string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Task Management System")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Login form
	formContent := []string{
		"Username:",
		m.usernameInput.View(),
		"",
		"Password:",
		m.passwordInput.View(),
	}

	form := boxStyle.Width(40).Render(strings.Join(formContent, "\n"))
	b.WriteString(form)
	b.WriteString("\n\n")

	// Help text
	if m.authenticating {
		b.WriteString(subtitleStyle.Render("Authenticating..."))
	} else {
		help := helpStyle.Render("Tab: Switch fields • Enter: Login • Ctrl+R: Register • Esc: Quit")
		b.WriteString(help)
	}

	// Center the content
	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		b.String())
}

// ValidationError is a custom error type for validation
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
