# Terminal User Interface (TUI)

A bubbletea/lipgloss-infused TUI (Terminal User Interface) that allows a user to CRUD-manage all the entities that exist in the backend, via the SERVER api calls.

## Usage

`task tui` will invoke the TUI.

It should use all the same features as the CLI/Server.

## Navigation

Navigation is via:
- Arrow keys or WASD keys for moving up/down
- Space or Enter for selection
- Backspace or Escape for going back
- Tab/Shift-Tab for moving between form fields
- CTRL-C twice quits the TUI back to the terminal

## Implementation

### Technology Stack
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework following The Elm Architecture
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal UI components
- [Bubbles](https://github.com/charmbracelet/bubbles) - Common TUI components (text input, etc.)

### Architecture

The TUI is organized into several components:

1. **Main Model** (`tui.go`) - Coordinates view switching and manages sub-models
2. **Login Screen** (`login.go`) - Authentication form
3. **Main Menu** (`mainmenu.go`) - Primary navigation hub
4. **Entity Views** - List and form views for each entity type:
   - Projects (`projects.go`, `project_form.go`)
   - Tasks (`tasks.go`, `task_form.go`)
   - Roles (`roles.go`, `role_form.go`)
   - Users (`users.go`, `user_form.go`)
   - Config (`config.go`, `config_form.go`)
   - Workers (`workers.go`) - View only

### Features

#### Authentication
- Login screen with username/password inputs
- Credentials can be pre-filled from environment variables
- API authentication test before proceeding to main menu

#### Main Menu
- Six primary sections: Projects, Tasks, Roles, Users, Config, Workers
- Each option includes a description
- Keyboard-driven navigation

#### Entity Management
All entities support:
- **List View**: Table display with pagination support
- **Create**: Form for creating new entities
- **Edit**: Form pre-filled with existing data
- **Delete**: Immediate deletion with confirmation
- **Refresh**: Reload data from server

#### Form Handling
- Tab/Shift-Tab navigation between fields
- Enter to submit
- Escape to cancel
- Real-time validation
- Loading indicators during submission

#### Styling
- Consistent color scheme throughout
- Selected items highlighted
- Status indicators (active/inactive, success/error)
- Bordered boxes for forms
- Responsive to terminal size

### API Integration

The TUI reuses the existing CLI client (`internal/cli/client.go`) for all API communication, ensuring consistent behavior across interfaces. All API calls are authenticated using Basic Auth credentials provided during login.
