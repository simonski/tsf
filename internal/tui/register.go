package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/simonski/task/internal/cli"
)

type registerModel struct {
	config        *cli.Config
	usernameInput textinput.Model
	passwordInput textinput.Model
	confirmInput  textinput.Model
	focusIndex    int
	registering   bool
	showError     bool
	errorMessage  string
}

func newRegisterModel(config *cli.Config) registerModel {
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

	confirmInput := textinput.New()
	confirmInput.Placeholder = "Confirm Password"
	confirmInput.EchoMode = textinput.EchoPassword
	confirmInput.CharLimit = 100
	confirmInput.Width = 30

	return registerModel{
		config:        config,
		usernameInput: usernameInput,
		passwordInput: passwordInput,
		confirmInput:  confirmInput,
		focusIndex:    0,
	}
}

func (m registerModel) Update(msg tea.Msg) (registerModel, tea.Cmd) {
	if m.registering {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m.submit()
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % 3
			return m.updateFocus()
		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + 3) % 3
			return m.updateFocus()
		case "esc":
			// Return to login view
			return m, func() tea.Msg {
				return switchToLoginMsg{}
			}
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	if m.focusIndex == 0 {
		m.usernameInput, cmd = m.usernameInput.Update(msg)
	} else if m.focusIndex == 1 {
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	} else {
		m.confirmInput, cmd = m.confirmInput.Update(msg)
	}

	return m, cmd
}

func (m registerModel) updateFocus() (registerModel, tea.Cmd) {
	if m.focusIndex == 0 {
		m.usernameInput.Focus()
		m.passwordInput.Blur()
		m.confirmInput.Blur()
	} else if m.focusIndex == 1 {
		m.usernameInput.Blur()
		m.passwordInput.Focus()
		m.confirmInput.Blur()
	} else {
		m.usernameInput.Blur()
		m.passwordInput.Blur()
		m.confirmInput.Focus()
	}
	return m, nil
}

func (m registerModel) submit() (registerModel, tea.Cmd) {
	username := strings.TrimSpace(m.usernameInput.Value())
	password := m.passwordInput.Value()
	confirm := m.confirmInput.Value()

	// Validation
	if username == "" {
		m.showError = true
		m.errorMessage = "Username is required"
		return m, nil
	}

	if password == "" {
		m.showError = true
		m.errorMessage = "Password is required"
		return m, nil
	}

	if len(password) < 8 {
		m.showError = true
		m.errorMessage = "Password must be at least 8 characters"
		return m, nil
	}

	if password != confirm {
		m.showError = true
		m.errorMessage = "Passwords do not match"
		return m, nil
	}

	m.registering = true
	m.showError = false

	// Register user
	return m, func() tea.Msg {
		client := cli.NewClient(m.config)
		body := map[string]string{
			"username": username,
			"password": password,
			"type":     "human",
		}
		_, err := client.Request("POST", "/api/v1/auth/register", body)
		if err != nil {
			return errorMsg{err: err}
		}
		// After successful registration, automatically login
		m.config.Username = username
		m.config.Password = password
		return registrationSuccessMsg{}
	}
}

func (m registerModel) View(width, height int) string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Register New Account")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Register form
	formContent := []string{
		"Username:",
		m.usernameInput.View(),
		"",
		"Password:",
		m.passwordInput.View(),
		"",
		"Confirm Password:",
		m.confirmInput.View(),
	}

	form := boxStyle.Width(40).Render(strings.Join(formContent, "\n"))
	b.WriteString(form)
	b.WriteString("\n\n")

	// Error message
	if m.showError {
		errorText := errorStyle.Render("Error: " + m.errorMessage)
		b.WriteString(errorText)
		b.WriteString("\n\n")
	}

	// Help text
	if m.registering {
		b.WriteString(subtitleStyle.Render("Creating account..."))
	} else {
		help := helpStyle.Render("Tab: Switch fields • Enter: Register • Esc: Back to login")
		b.WriteString(help)
	}

	// Center the content
	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		b.String())
}
