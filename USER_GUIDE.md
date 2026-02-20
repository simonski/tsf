# User Guide

Complete guide to using sf.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Web Interface](#web-interface)
3. [Terminal User Interface (TUI)](#terminal-user-interface-tui)
4. [Command Line Interface](#command-line-interface)
5. [API Usage](#api-usage)
6. [Managing Projects](#managing-projects)
7. [Managing Tasks](#managing-tasks)
8. [Roles and Workers](#roles-and-workers)
9. [Comments](#comments)
10. [Configuration](#configuration)

## Getting Started

### Installation

#### Using Pre-built Binaries

Download the latest release from the releases page and extract the binaries:

```bash
tar -xzf task-*.tar.gz
cd task
```

#### Building from Source

```bash
git clone <repository-url>
cd task
make build
```

### First-Time Setup

1. **Initialize the database:**

```bash
./sf initdb -f ~/.config/sf/sf.db
```

This creates a SQLite database with:
- Default admin user (username: `admin`, password: `admin123`)
- Sample project
- Sample roles

2. **Start the server:**

```bash
./sf server -f ~/.config/sf/sf.db -port 8080
```

3. **Access the web UI:**

Open http://localhost:8080 in your browser.

### Changing the Admin Password

**Important:** Change the default admin password immediately!

Using the CLI:
```bash
export SF_SERVER_URL=http://localhost:8080
export SF_USERNAME=admin
export SF_PASSWORD=admin123

# Change password via direct database update or API
# Note: Password management via CLI is not yet implemented
# Use the web UI or API to change passwords
```

## Web Interface

### Login

1. Navigate to http://localhost:8080
2. Enter your username and password
3. Click "Login"

Your credentials are stored locally for convenience.

### Register New Account

1. Click the "Register" tab
2. Enter username and password (minimum 8 characters)
3. Confirm password
4. Click "Register"

You'll be automatically logged in after registration.

### Project Selection

After logging in:
1. Use the project dropdown in the header to select a project
2. The kanban board will load tasks for the selected project

### Passkey Authentication (Passwordless Login)

sf supports passkey authentication, providing a more secure and convenient login experience using your device's biometric sensors (fingerprint, face recognition) or hardware security keys.

#### Registering a Passkey

1. Log in with your username and password
2. Click the hamburger menu (☰) and select "Settings"
3. Click "Add New Passkey"
4. Enter an optional name for your passkey (e.g., "My Laptop", "iPhone")
5. Follow your browser's prompts to create the passkey using:
   - Fingerprint sensor
   - Face recognition
   - Device PIN
   - Hardware security key (YubiKey, etc.)

Your passkey is now registered and can be used for future logins.

#### Signing in with a Passkey

1. On the login page, click "Sign in with Passkey"
2. Optionally enter your username (or leave blank for discoverable credentials)
3. Follow your browser's prompts to authenticate using your passkey

No password needed!

#### Managing Passkeys

In the Settings screen, you can:
- View all registered passkeys
- See when each passkey was created and last used
- Delete passkeys you no longer use

**Note:** Passkeys are tied to your device and browser. Register multiple passkeys if you access the system from different devices.

### Kanban Board

The board displays four columns:

- **To Do**: Tasks ready to be started
- **In Progress**: Tasks currently being worked on
- **Blocked**: Tasks that are stuck or waiting
- **Completed**: Finished tasks

Each task card shows:
- Title
- Description (truncated)
- Priority (high/medium/low)
- Creation date

### Viewing Task Details

Click any task card to open the task detail modal showing:
- Full description
- Current status
- Priority level
- Creation date
- Assigned user

### Updating Task Status

1. Click a task to open the modal
2. Change the status dropdown
3. Click "Save Changes"
4. The task moves to the appropriate column

## Terminal User Interface (TUI)

The TUI provides a full-featured terminal-based interface for managing tasks.

### Launching the TUI

```bash
# Using environment variables for credentials
export SF_SERVER_URL=http://localhost:8080
export SF_USERNAME=admin
export SF_PASSWORD=admin123

./sf tui
```

Or with flags:

```bash
./sf tui -url http://localhost:8080
```

### First-Time Login and Registration

When you launch the TUI, you'll see the login screen.

#### Logging In

1. Enter your username (Tab to move to next field)
2. Enter your password
3. Press Enter to login

#### Registering a New Account

1. From the login screen, press **Ctrl+R** to switch to registration
2. Enter a username
3. Enter a password (minimum 8 characters)
4. Confirm your password
5. Press Enter to register

You'll be automatically logged in after successful registration.

To go back to login from registration, press **Esc**.

### Navigation

**General Navigation:**
- **Tab/Shift+Tab**: Move between fields or options
- **Enter**: Select or submit
- **Esc**: Go back or cancel
- **Ctrl+C** (twice): Quit the application

**In List Views:**
- **↑/↓ or j/k**: Navigate items
- **n**: Create new item
- **e**: Edit selected item
- **d**: Delete selected item
- **r**: Refresh list
- **Backspace**: Return to main menu

### Main Menu

After login, you'll see the main menu with six options:

1. **Projects**: Manage projects and their details
2. **Tasks**: View and manage tasks across all projects
3. **Roles**: Define and manage role-based access
4. **Users**: User account management
5. **Config**: System configuration
6. **Workers**: Monitor active workers (view-only)

Use arrow keys to select an option and press Enter.

### Managing Entities

All entity screens (Projects, Tasks, Roles, Users, Config) follow the same pattern:

#### Viewing Lists

- Navigate items with arrow keys or j/k
- View details of the selected item in the table

#### Creating Items

1. Press **n** for "New"
2. Fill in the form fields (use Tab to move between fields)
3. Press Enter to submit
4. Press Esc to cancel

#### Editing Items

1. Select an item from the list
2. Press **e** for "Edit"
3. Modify fields as needed
4. Press Enter to save
5. Press Esc to cancel

#### Deleting Items

1. Select an item from the list
2. Press **d** for "Delete"
3. The item is removed immediately

#### Refreshing

Press **r** to reload the current list from the server.

### Workers View

The Workers screen shows real-time information about active workers:

- Worker ID
- Status
- Last heartbeat time
- Active tasks

This is a view-only screen. Use arrow keys to browse workers.

### Tips

- **Credentials**: If you set `SF_USERNAME` and `SF_PASSWORD` environment variables, they'll auto-fill on the login screen
- **Quick Navigation**: Use keyboard shortcuts consistently across all screens for efficient workflow
- **Error Messages**: Any errors will display at the top of the screen in red

## Command Line Interface

### Authentication

#### Login

The easiest way to authenticate is using the `login` command:

```bash
./sf login
```

You'll be prompted for username and password. Your credentials will be saved to `~/.config/sf/credentials.json` for future use.

You can also provide credentials via flags:

```bash
./sf login -username admin -password admin123
```

#### Register a New Account

```bash
./sf register
```

You'll be prompted for username and password. After successful registration, your credentials are automatically saved.

Or with flags:

```bash
./sf register -username myuser -password mypassword
```

#### Using Environment Variables

Alternatively, set environment variables:

```bash
export SF_SERVER_URL=http://localhost:8080
export SF_USERNAME=admin
export SF_PASSWORD=admin123
```

#### Credential Priority

Credentials are loaded in this order:
1. Command-line flags (`-username`, `-password`)
2. Environment variables (`SF_USERNAME`, `SF_PASSWORD`)
3. Saved credentials file (`~/.config/sf/credentials.json`)

### Configuration

Or use command-line flags:

```bash
./sf -server http://localhost:8080 -username admin -password admin123 <command>
```

### Project Management

#### List Projects

```bash
./sf project list
```

Output:
```
ID                                   | Name          | Description
-------------------------------------|---------------|------------------
550e8400-e29b-41d4-a716-446655440000 | Sample Project| Default project
```

#### Create Project

```bash
./sf project create \
  -name "My New Project" \
  -description "Project for feature development"
```

#### View Project Details

```bash
./sf project get -id <project-id>
```

#### Update Project

```bash
./sf project update \
  -id <project-id> \
  -name "Updated Name" \
  -description "Updated description"
```

#### Delete Project

```bash
./sf project delete -id <project-id>
```

#### Project Files and Notes

Project files and notes are project-scoped artifacts managed from the CLI.

```bash
# files
./sf project file list -project_id <project-id>
./sf project file create "README.md" -project_id <project-id> -content "# Intro"
./sf project file get <file-id> -project_id <project-id>
./sf project file update <file-id> -project_id <project-id> -content "# Updated"
./sf project file rm <file-id> -project_id <project-id>

# notes
./sf project note list -project_id <project-id>
./sf project note create "Kickoff" -project_id <project-id> -content "Initial decisions"
./sf project note get <note-id> -project_id <project-id>
./sf project note update <note-id> -project_id <project-id> -content "Revised decisions"
./sf project note rm <note-id> -project_id <project-id>
```

### Task Management

#### List Tasks

List all tasks:
```bash
./sf task list
```

List tasks for a specific project:
```bash
./sf task list -project <project-id>
```

Filter by status:
```bash
./sf task list -status todo
./sf task list -status in_progress
./sf task list -status blocked
./sf task list -status completed
```

#### Create Task

```bash
./sf task create \
  -title "Implement user authentication" \
  -description "Add login and registration endpoints" \
  -project <project-id> \
  -priority high
```

Priority options: `low`, `medium`, `high` (default: `medium`)

#### View Task Details

```bash
./sf task get -id <task-id>
```

#### Update Task

```bash
./sf task update \
  -id <task-id> \
  -title "Updated title" \
  -description "Updated description" \
  -status in_progress \
  -priority medium
```

#### Delete Task

```bash
./sf task delete -id <task-id>
```

#### Task Actions

Claim a task (assign to yourself):
```bash
./sf task claim -id <task-id>
```

Assign task to another user:
```bash
./sf task assign -id <task-id> -user <username>
```

Free a task (unassign):
```bash
./sf task free -id <task-id>
```

Mark task as complete:
```bash
./sf task complete -id <task-id>
```

#### View Task History

```bash
./sf task history -id <task-id>
```

Shows all status changes with timestamps.

### User Management

#### List Users

```bash
./sf user list
```

#### Create User

```bash
./sf user create \
  -username newuser \
  -password securepass123 \
  -type human
```

User types: `human`, `worker`, `orchestrator`

#### Enable/Disable Users

```bash
./sf user enable -username <username>
./sf user disable -username <username>
```

### Role Management

#### List Roles

```bash
./sf role list
```

#### Create Role

```bash
./sf role create \
  -name "Senior Developer" \
  -instructions "You are an experienced developer with expertise in Go and Python." \
  -project <project-id>
```

#### Update Role

```bash
./sf role update \
  -id <role-id> \
  -name "Updated Role Name" \
  -instructions "Updated instructions"
```

#### Delete Role

```bash
./sf role delete -id <role-id>
```

### Output Formats

Get JSON output for scripting:

```bash
./sf task list -format json
```

Default output is human-readable tables.

## API Usage

> **Note:** For most operations, use the `task` CLI commands shown throughout this guide. The following curl examples are provided for direct API access, automation, or integration with other tools.

All API endpoints require Basic Authentication.

### Authentication

Include credentials in every request:

```bash
curl -u username:password http://localhost:8080/api/v1/endpoint
```

### Common Endpoints

#### Health Check (No Auth Required)

```bash
curl http://localhost:8080/api/v1/health
```

Response:
```json
{
  "status": "ok"
}
```

#### List Projects

```bash
curl -u admin:admin123 http://localhost:8080/api/v1/projects
```

#### Create Project

```bash
curl -u admin:admin123 -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New Project",
    "description": "Project description"
  }'
```

#### List Tasks

```bash
curl -u admin:admin123 http://localhost:8080/api/v1/tasks
```

#### Create Task

```bash
curl -u admin:admin123 -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Task title",
    "description": "Task description",
    "status": "todo",
    "priority": "medium",
    "project_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

#### Update Task Status

```bash
curl -u admin:admin123 -X PUT http://localhost:8080/api/v1/tasks/<task-id> \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Task title",
    "description": "Task description",
    "status": "in_progress",
    "priority": "high",
    "project_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

### API Reference

See [api-specification.yaml](../api-specification.yaml) for complete API documentation with all endpoints, request/response formats, and error codes.

## Managing Projects

### Project Workflow

1. **Create a project** for each major initiative or product
2. **Add team members** (users) to projects
3. **Create roles** to define AI worker behavior for the project
4. **Create tasks** within the project
5. **Track progress** using the kanban board

### Project Context

The CLI can maintain a default project context:

```bash
export SF_PROJECT_ID=550e8400-e29b-41d4-a716-446655440000
```

Then commands default to that project:
```bash
./sf task list  # Lists tasks for the default project
```

### Project Members

Add members to a project to control access and visibility:

```bash
./sf project add-member \
  -project <project-id> \
  -username <username>
```

## Managing Tasks

### Task Lifecycle

Tasks move through these states:

1. **todo** - Initial state, ready to be started
2. **in_progress** - Work has begun
3. **blocked** - Stuck, waiting for something
4. **completed** - Work is finished

### Task Priority

Set priority to help workers and users focus:

- **high** - Urgent, needs immediate attention
- **medium** - Normal priority (default)
- **low** - Can be deferred

### Task Assignment

Tasks can be:
- **Unassigned** - Available for anyone
- **Claimed** - Assigned to the current user
- **Assigned** - Assigned to a specific user/worker

### Best Practices

- Write clear, actionable task titles
- Include acceptance criteria in descriptions
- Set appropriate priority levels
- Keep tasks focused and scoped
- Update status regularly
- Use task history to track decisions

## Roles and Workers

### Understanding Roles

Roles define how AI workers should behave when processing tasks:

- Each role has **instructions** (system prompt)
- Roles can be **project-specific** or **global**
- Workers use role instructions as context

### Creating Effective Roles

```bash
./sf role create \
  -name "Code Reviewer" \
  -instructions "You are a senior developer reviewing code for quality, security, and best practices. Provide constructive feedback." \
  -project <project-id>
```

Role instruction tips:
- Be specific about expertise and approach
- Define quality standards
- Specify output format expectations
- Include domain knowledge

### Running a Worker

Workers automatically request and process tasks:

```bash
./sf worker \
  -server http://localhost:8080 \
  -username worker1 \
  -password workerpass
```

The worker will:
1. Register with the server
2. Send periodic heartbeats
3. Request available work
4. Process tasks (with LLM)
5. Report completion

### Running the Orchestrator

The orchestrator manages task distribution:

```bash
./sf orchestrator \
  -server http://localhost:8080 \
  -username orchestrator \
  -password orchpass
```

The orchestrator:
- Monitors task status
- Routes tasks to appropriate workers
- Manages task lifecycle
- Loads configuration dynamically

## Comments

`sf comment` provides social-style comments for both projects and tasks.

Each comment stores:
- owner
- timestamps
- entity attachment (`project` or `task`)
- edit history
- soft-delete status

```bash
# project comments
./sf comment list -project_id <project-id>
./sf comment create -project_id <project-id> -text "Looks good"

# task comments
./sf comment list -task_id <task-id>
./sf comment create -task_id <task-id> -text "Needs test coverage"

# comment lifecycle
./sf comment get <comment-id>
./sf comment update <comment-id> -text "Updated text"
./sf comment history <comment-id>
./sf comment rm <comment-id>   # soft delete
```

## Configuration

### System Configuration

Configuration is stored in the database and managed via the CLI.

#### View Configuration

```bash
sf config list
```

#### Set Configuration Value

```bash
sf config set -key key -val new-value
```

#### Delete Configuration Key

```bash
sf config delete -key key
```

### Configuration Keys

Common configuration options:

| Key | Description | Default |
|-----|-------------|---------|
| `orchestrator.interval` | How often orchestrator polls tasks | `30s` |
| `worker.interval` | How often workers request work | `5s` |
| `heartbeat.timeout` | Worker heartbeat timeout | `60s` |

### Environment Variables

Server configuration:

```bash
# Server
export SF_DB_PATH=~/.config/sf/sf.db
export SF_SERVER_PORT=8080

# CLI
export SF_SERVER_URL=http://localhost:8080
export SF_USERNAME=admin
export SF_PASSWORD=admin123
export SF_PROJECT_ID=<default-project-id>

# Orchestrator/Worker
export SF_SERVER_URL=http://localhost:8080
export SF_USERNAME=orchestrator
export SF_PASSWORD=orchpass
```

## Troubleshooting

### Server Won't Start

**Error: Database does not exist**
```bash
./sf initdb -f ~/.config/sf/sf.db
```

**Error: Port already in use**
```bash
# Use a different port
./sf server -port 8081
```

### Authentication Failures

**Error: 401 Unauthorized**

- Verify username and password
- Check environment variables
- Ensure user account is active

### CLI Connection Issues

**Error: Connection refused**

- Verify server is running: `task config list` (or check server process)
- Check `SF_URL` environment variable
- Verify firewall/network settings

### Task Not Appearing

- Verify project ID is correct
- Check task status filter
- Ensure you have access to the project
- Refresh the web interface

### Worker Not Receiving Tasks

- Check worker heartbeat is successful
- Verify orchestrator is running
- Check task status (should be `todo` or `in_progress`)
- Review server logs

## Advanced Usage

### Running Multiple Workers

Scale horizontally by running multiple worker instances:

```bash
# Terminal 1
sf worker -url http://localhost:8080 -username worker1 -password pass1

# Terminal 2
sf worker -url http://localhost:8080 -username worker2 -password pass2

# Terminal 3
sf worker -url http://localhost:8080 -username worker3 -password pass3
```

### Batch Operations

Use the CLI for batch operations:

```bash
# Create multiple tasks
for title in "Task 1" "Task 2" "Task 3"; do
  task task create -title "$title" -project_id "<project-id>"
done
```

### Monitoring

Monitor worker health via server logs and worker output. Workers send heartbeats automatically to maintain their active status.

### Backup and Restore

Backup the SQLite database regularly:

```bash
# Backup
cp ~/.config/sf/sf.db ~/.config/sf/sf.db.backup

# Restore
cp ~/.config/sf/sf.db.backup ~/.config/sf/sf.db
```

For production, use SQLite's backup command:

```bash
sqlite3 ~/.config/sf/sf.db ".backup ~/.config/sf/sf.db.backup"
```

## Tips and Best Practices

1. **Start Small**: Begin with one project and a few tasks
2. **Clear Titles**: Use descriptive, actionable task titles
3. **Regular Updates**: Keep task status current
4. **Use Priorities**: Help focus work on what matters
5. **Define Roles**: Create specific roles for different work types
6. **Monitor Workers**: Check heartbeats and logs regularly
7. **Backup Database**: Regular backups prevent data loss
8. **Secure Passwords**: Change defaults immediately
9. **Scale Workers**: Add workers as workload grows
10. **Review History**: Use task history to understand decisions

## Getting Help

- Check this guide for common tasks
- Review [README.md](../README.md) for architecture overview
- See [API Specification](../api-specification.yaml) for API details
- Check design docs in [docs/](../docs/) for system internals
