package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/simonski/task/internal/cli"
)

// ViewType represents different screens in the TUI
type ViewType int

const (
	ViewLogin ViewType = iota
	ViewMainMenu
	ViewProjects
	ViewProjectForm
	ViewTasks
	ViewTaskForm
	ViewRoles
	ViewRoleForm
	ViewUsers
	ViewUserForm
	ViewConfig
	ViewConfigForm
	ViewWorkers
)

// Model is the main TUI model
type Model struct {
	client   *cli.Client
	config   *cli.Config
	view     ViewType
	width    int
	height   int
	err      error
	quitCtrl int // Track CTRL-C presses

	// Sub-models
	loginModel      loginModel
	mainMenuModel   mainMenuModel
	projectsModel   projectsModel
	projectFormModel projectFormModel
	tasksModel      tasksModel
	taskFormModel   taskFormModel
	rolesModel      rolesModel
	roleFormModel   roleFormModel
	usersModel      usersModel
	userFormModel   userFormModel
	configModel     configModel
	configFormModel configFormModel
	workersModel    workersModel
}

// New creates a new TUI model
func New(config *cli.Config) *Model {
	return &Model{
		config:     config,
		view:       ViewLogin,
		quitCtrl:   0,
		loginModel: newLoginModel(config),
	}
}

// Init initializes the TUI
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle CTRL-C for quit
		if msg.String() == "ctrl+c" {
			m.quitCtrl++
			if m.quitCtrl >= 2 {
				return m, tea.Quit
			}
			return m, nil
		}
		// Reset quit counter on any other key
		m.quitCtrl = 0

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case authSuccessMsg:
		m.client = cli.NewClient(m.config)
		m.view = ViewMainMenu
		m.mainMenuModel = newMainMenuModel()
		return m, nil

	case changeViewMsg:
		return m.changeView(msg.view, msg.data)

	case errorMsg:
		m.err = msg.err
		return m, nil
	}

	// Delegate to sub-model
	return m.updateSubModel(msg)
}

// View renders the TUI
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Show error if present
	var errorView string
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true).
			Padding(1)
		errorView = errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Render current view
	var content string
	switch m.view {
	case ViewLogin:
		content = m.loginModel.View(m.width, m.height)
	case ViewMainMenu:
		content = m.mainMenuModel.View(m.width, m.height)
	case ViewProjects:
		content = m.projectsModel.View(m.width, m.height)
	case ViewProjectForm:
		content = m.projectFormModel.View(m.width, m.height)
	case ViewTasks:
		content = m.tasksModel.View(m.width, m.height)
	case ViewTaskForm:
		content = m.taskFormModel.View(m.width, m.height)
	case ViewRoles:
		content = m.rolesModel.View(m.width, m.height)
	case ViewRoleForm:
		content = m.roleFormModel.View(m.width, m.height)
	case ViewUsers:
		content = m.usersModel.View(m.width, m.height)
	case ViewUserForm:
		content = m.userFormModel.View(m.width, m.height)
	case ViewConfig:
		content = m.configModel.View(m.width, m.height)
	case ViewConfigForm:
		content = m.configFormModel.View(m.width, m.height)
	case ViewWorkers:
		content = m.workersModel.View(m.width, m.height)
	default:
		content = "Unknown view"
	}

	// Combine error and content
	if errorView != "" {
		return errorView + "\n" + content
	}
	return content
}

// updateSubModel delegates updates to the appropriate sub-model
func (m *Model) updateSubModel(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.view {
	case ViewLogin:
		m.loginModel, cmd = m.loginModel.Update(msg)
	case ViewMainMenu:
		m.mainMenuModel, cmd = m.mainMenuModel.Update(msg)
	case ViewProjects:
		m.projectsModel, cmd = m.projectsModel.Update(msg, m.client)
	case ViewProjectForm:
		m.projectFormModel, cmd = m.projectFormModel.Update(msg, m.client)
	case ViewTasks:
		m.tasksModel, cmd = m.tasksModel.Update(msg, m.client)
	case ViewTaskForm:
		m.taskFormModel, cmd = m.taskFormModel.Update(msg, m.client)
	case ViewRoles:
		m.rolesModel, cmd = m.rolesModel.Update(msg, m.client)
	case ViewRoleForm:
		m.roleFormModel, cmd = m.roleFormModel.Update(msg, m.client)
	case ViewUsers:
		m.usersModel, cmd = m.usersModel.Update(msg, m.client)
	case ViewUserForm:
		m.userFormModel, cmd = m.userFormModel.Update(msg, m.client)
	case ViewConfig:
		m.configModel, cmd = m.configModel.Update(msg, m.client)
	case ViewConfigForm:
		m.configFormModel, cmd = m.configFormModel.Update(msg, m.client)
	case ViewWorkers:
		m.workersModel, cmd = m.workersModel.Update(msg, m.client)
	}
	return m, cmd
}

// changeView switches to a different view
func (m *Model) changeView(view ViewType, data interface{}) (tea.Model, tea.Cmd) {
	m.view = view
	m.err = nil

	switch view {
	case ViewMainMenu:
		m.mainMenuModel = newMainMenuModel()
	case ViewProjects:
		m.projectsModel = newProjectsModel()
		return m, m.projectsModel.loadProjects(m.client)
	case ViewProjectForm:
		if project, ok := data.(*Project); ok {
			m.projectFormModel = newProjectFormModel(project)
		} else {
			m.projectFormModel = newProjectFormModel(nil)
		}
	case ViewTasks:
		m.tasksModel = newTasksModel()
		return m, m.tasksModel.loadTasks(m.client)
	case ViewTaskForm:
		if task, ok := data.(*Task); ok {
			m.taskFormModel = newTaskFormModel(task)
		} else {
			m.taskFormModel = newTaskFormModel(nil)
		}
	case ViewRoles:
		m.rolesModel = newRolesModel()
		return m, m.rolesModel.loadRoles(m.client)
	case ViewRoleForm:
		if role, ok := data.(*Role); ok {
			m.roleFormModel = newRoleFormModel(role)
		} else {
			m.roleFormModel = newRoleFormModel(nil)
		}
	case ViewUsers:
		m.usersModel = newUsersModel()
		return m, m.usersModel.loadUsers(m.client)
	case ViewUserForm:
		if user, ok := data.(*User); ok {
			m.userFormModel = newUserFormModel(user)
		} else {
			m.userFormModel = newUserFormModel(nil)
		}
	case ViewConfig:
		m.configModel = newConfigModel()
		return m, m.configModel.loadConfigs(m.client)
	case ViewConfigForm:
		if config, ok := data.(*Config); ok {
			m.configFormModel = newConfigFormModel(config)
		} else {
			m.configFormModel = newConfigFormModel(nil)
		}
	case ViewWorkers:
		m.workersModel = newWorkersModel()
		return m, m.workersModel.loadWorkers(m.client)
	}

	return m, nil
}

// Run starts the TUI
func Run(config *cli.Config) error {
	p := tea.NewProgram(New(config), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
