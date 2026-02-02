# User Guide

Complete guide to using the Task Management System.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Web Interface](#web-interface)
3. [Command Line Interface](#command-line-interface)
4. [API Usage](#api-usage)
5. [Managing Projects](#managing-projects)
6. [Managing Tasks](#managing-tasks)
7. [Roles and Workers](#roles-and-workers)
8. [Configuration](#configuration)

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
./bin/task-initdb -f ~/.config/task/task.db
```

This creates a SQLite database with:
- Default admin user (username: `admin`, password: `admin123`)
- Sample project
- Sample roles

2. **Start the server:**

```bash
./bin/task-server -f ~/.config/task/task.db -port 8080
```

3. **Access the web UI:**

Open http://localhost:8080 in your browser.

### Changing the Admin Password

**Important:** Change the default admin password immediately!

Using the CLI:
```bash
export TASK_SERVER_URL=http://localhost:8080
export TASK_USERNAME=admin
export TASK_PASSWORD=admin123

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

## Command Line Interface

### Configuration

Set environment variables for authentication:

```bash
export TASK_SERVER_URL=http://localhost:8080
export TASK_USERNAME=admin
export TASK_PASSWORD=admin123
```

Or use command-line flags:

```bash
./bin/task -server http://localhost:8080 -username admin -password admin123 <command>
```

### Project Management

#### List Projects

```bash
./bin/task project list
```

Output:
```
ID                                   | Name          | Description
-------------------------------------|---------------|------------------
550e8400-e29b-41d4-a716-446655440000 | Sample Project| Default project
```

#### Create Project

```bash
./bin/task project create \
  -name "My New Project" \
  -description "Project for feature development"
```

#### View Project Details

```bash
./bin/task project get -id <project-id>
```

#### Update Project

```bash
./bin/task project update \
  -id <project-id> \
  -name "Updated Name" \
  -description "Updated description"
```

#### Delete Project

```bash
./bin/task project delete -id <project-id>
```

### Task Management

#### List Tasks

List all tasks:
```bash
./bin/task task list
```

List tasks for a specific project:
```bash
./bin/task task list -project <project-id>
```

Filter by status:
```bash
./bin/task task list -status todo
./bin/task task list -status in_progress
./bin/task task list -status blocked
./bin/task task list -status completed
```

#### Create Task

```bash
./bin/task task create \
  -title "Implement user authentication" \
  -description "Add login and registration endpoints" \
  -project <project-id> \
  -priority high
```

Priority options: `low`, `medium`, `high` (default: `medium`)

#### View Task Details

```bash
./bin/task task get -id <task-id>
```

#### Update Task

```bash
./bin/task task update \
  -id <task-id> \
  -title "Updated title" \
  -description "Updated description" \
  -status in_progress \
  -priority medium
```

#### Delete Task

```bash
./bin/task task delete -id <task-id>
```

#### Task Actions

Claim a task (assign to yourself):
```bash
./bin/task task claim -id <task-id>
```

Assign task to another user:
```bash
./bin/task task assign -id <task-id> -user <username>
```

Free a task (unassign):
```bash
./bin/task task free -id <task-id>
```

Mark task as complete:
```bash
./bin/task task complete -id <task-id>
```

#### View Task History

```bash
./bin/task task history -id <task-id>
```

Shows all status changes with timestamps.

### User Management

#### List Users

```bash
./bin/task user list
```

#### Create User

```bash
./bin/task user create \
  -username newuser \
  -password securepass123 \
  -type human
```

User types: `human`, `worker`, `orchestrator`

#### Enable/Disable Users

```bash
./bin/task user enable -username <username>
./bin/task user disable -username <username>
```

### Role Management

#### List Roles

```bash
./bin/task role list
```

#### Create Role

```bash
./bin/task role create \
  -name "Senior Developer" \
  -instructions "You are an experienced developer with expertise in Go and Python." \
  -project <project-id>
```

#### Update Role

```bash
./bin/task role update \
  -id <role-id> \
  -name "Updated Role Name" \
  -instructions "Updated instructions"
```

#### Delete Role

```bash
./bin/task role delete -id <role-id>
```

### Output Formats

Get JSON output for scripting:

```bash
./bin/task task list -format json
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
export TASK_PROJECT_ID=550e8400-e29b-41d4-a716-446655440000
```

Then commands default to that project:
```bash
./bin/task task list  # Lists tasks for the default project
```

### Project Members

Add members to a project to control access and visibility:

```bash
./bin/task project add-member \
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
./bin/task role create \
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
./bin/task-worker \
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
./bin/task-orchestrator \
  -server http://localhost:8080 \
  -username orchestrator \
  -password orchpass
```

The orchestrator:
- Monitors task status
- Routes tasks to appropriate workers
- Manages task lifecycle
- Loads configuration dynamically

## Configuration

### System Configuration

Configuration is stored in the database and managed via the CLI.

#### View Configuration

```bash
task config list
```

#### Set Configuration Value

```bash
task config set -key key -val new-value
```

#### Delete Configuration

```bash
task config delete -key key
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
export TASK_DB_PATH=~/.config/task/task.db
export TASK_SERVER_PORT=8080

# CLI
export TASK_SERVER_URL=http://localhost:8080
export TASK_USERNAME=admin
export TASK_PASSWORD=admin123
export TASK_PROJECT_ID=<default-project-id>

# Orchestrator/Worker
export TASK_SERVER_URL=http://localhost:8080
export TASK_USERNAME=orchestrator
export TASK_PASSWORD=orchpass
```

## Troubleshooting

### Server Won't Start

**Error: Database does not exist**
```bash
./bin/task-initdb -f ~/.config/task/task.db
```

**Error: Port already in use**
```bash
# Use a different port
./bin/task-server -port 8081
```

### Authentication Failures

**Error: 401 Unauthorized**

- Verify username and password
- Check environment variables
- Ensure user account is active

### CLI Connection Issues

**Error: Connection refused**

- Verify server is running: `task config list` (or check server process)
- Check `TASK_URL` environment variable
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
task worker -url http://localhost:8080 -username worker1 -password pass1

# Terminal 2
task worker -url http://localhost:8080 -username worker2 -password pass2

# Terminal 3
task worker -url http://localhost:8080 -username worker3 -password pass3
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
cp ~/.config/task/task.db ~/.config/task/task.db.backup

# Restore
cp ~/.config/task/task.db.backup ~/.config/task/task.db
```

For production, use SQLite's backup command:

```bash
sqlite3 ~/.config/task/task.db ".backup ~/.config/task/task.db.backup"
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
