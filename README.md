# tsf

A kanban-style task management system designed for software development workflows, with support for human users and automated AI workers.

## Overview

tsf is a comprehensive solution for managing software development tasks through a lifecycle approach. It provides:

- **Web UI**: Modern, responsive kanban board interface
- **REST API**: Complete OpenAPI-compatible REST API
- **CLI**: Terminal-based task management
- **Worker System**: Automated task processing with LLM integration
- **Orchestration**: Intelligent task routing and assignment
- **Single Binary**: All components embedded in one Go binary

## Features

- **Multi-User Support**: Human users, workers, and orchestrators with role-based access
- **Project Management**: Organize work into projects with team members
- **Kanban Workflow**: Visual board with To Do, In Progress, Blocked, and Completed columns
- **Task Lifecycle**: Track task history and status changes
- **Role-Based Context**: Define roles with specific instructions for AI workers
- **Basic Authentication**: Secure access with username/password
- **SQLite Database**: Lightweight, embedded database with no external dependencies
- **Docker Support**: Easy deployment with Docker Compose

## Architecture

```
┌─────────────┐
│   Web UI    │──┐
└─────────────┘  │
                 │    ┌──────────┐    ┌──────────┐
┌─────────────┐  ├───▶│  Server  │───▶│ Database │
│     CLI     │──┤    └──────────┘    │ (SQLite) │
└─────────────┘  │         ▲          └──────────┘
                 │         │
┌─────────────┐  │         │
│ Orchestrator│──┘         │
└─────────────┘            │
                           │
┌─────────────┐            │
│  Worker 1   │────────────┤
└─────────────┘            │
                           │
┌─────────────┐            │
│  Worker 2   │────────────┤
└─────────────┘            │
                           │
┌─────────────┐            │
│  Worker N   │────────────┘
└─────────────┘
```

## Quick Start

### Automated Setup

For the fastest setup, use the quickstart script:

```bash
./quickstart.sh
```

This will:
1. Build the binary
2. Initialize the database
3. Display credentials and next steps

### Prerequisites

- Go 1.23 or later (for building from source)
- Docker and Docker Compose (for containerized deployment)

### Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd task

# Build the binary
make build

# Initialize database
task initdb -f ~/.config/task/task.db

# Start server
task server -f ~/.config/task/task.db -port 8080
```

The web UI will be available at http://localhost:8080

Default credentials:
- Username: `admin`
- Password: `admin123`

### Using Docker

```bash
# Build Docker image
make docker

# Start all services
make docker-up

# Server available at http://localhost:8080
```

### Using CLI

```bash
# Set server URL and credentials
export TASK_SERVER_URL=http://localhost:8080
export TASK_USERNAME=admin
export TASK_PASSWORD=admin123

# List projects
task project list

# Create a task
task task create -title "Implement feature" -description "Add new functionality"

# List tasks
task task list
```

## Components

The `task` binary provides multiple subcommands for different operational modes:

### Server

The server mode (`task server`) provides:
- RESTful API endpoints (see [API Specification](api-specification.yaml))
- Static file serving for web UI
- Basic authentication middleware
- Database management

**Usage:**
```bash
task server -f <database-path> -port <port>
```

### Orchestrator

The orchestrator:
- Continuously monitors task status
- Routes tasks to appropriate workers
- Manages task lifecycle
- Loads configuration from server

**Usage:**
```bash
task orchestrator -url <server-url> -username <username> -password <password>
```

### Worker

Workers:
- Request work from the server
- Process tasks (with LLM integration ready)
- Report completion status
- Send heartbeat to maintain availability

**Usage:**
```bash
task worker -url <server-url> -username <username> -password <password>
```

### CLI

Command-line interface for:
- Project management
- Task CRUD operations
- User management
- Configuration

See [User Guide](USER_GUIDE.md) for detailed CLI documentation.

## Configuration

Configuration is stored in the database and managed via CLI:

```bash
# List all configuration
task config list

# Set configuration value
task config set -key orchestrator.interval -val 60s

# Get configuration value  
task config get -key orchestrator.interval

# Delete configuration
task config delete -key orchestrator.interval
```

Common configuration keys:
- `orchestrator.interval`: Orchestrator polling interval (default: 30s)
- `worker.interval`: Worker polling interval (default: 5s)

## Database Schema

The system uses SQLite with the following main tables:

- **users**: User accounts (human, worker, orchestrator)
- **projects**: Projects containing tasks
- **tasks**: Individual work items
- **task_history**: Task status change history
- **roles**: Role definitions with instructions
- **project_members**: Project team membership
- **config**: System configuration key-value pairs

See [Database Design](docs/DESIGN_DATABASE.md) for detailed schema.

## API Documentation

Full API documentation is available in [api-specification.yaml](api-specification.yaml).

Key endpoints:
- `POST /api/v1/users` - Register user
- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create project
- `GET /api/v1/tasks` - List tasks
- `POST /api/v1/tasks` - Create task
- `PUT /api/v1/tasks/{id}` - Update task
- `POST /api/v1/workers/request` - Request work (workers)
- `POST /api/v1/workers/heartbeat` - Send heartbeat

## Development

### Project Structure

```
.
├── cmd/                    # Command binaries
│   └── task-unified/      # Single unified binary (task)
├── internal/              # Internal packages
│   ├── cli/              # CLI implementation
│   ├── db/               # Database layer
│   ├── orchestrator/     # Orchestrator logic
│   ├── server/           # Server handlers
│   ├── web/              # Embedded web assets
│   └── worker/           # Worker logic
├── web/                   # Web frontend source
│   ├── index.html
│   └── static/
│       ├── app.js
│       └── style.css
├── docs/                  # Documentation
├── Dockerfile            # Docker build
├── docker-compose.yml    # Docker Compose config
└── Makefile              # Build automation
```

### Building

```bash
# Build the binary
make build

# Run tests
make test

# Run tests with coverage
make test-go-coverage

# Clean build artifacts
make clean
```

### Testing

```bash
# Run all tests (Go + E2E)
make test

# Run Go tests only
make test-go

# Run E2E tests only (requires Node.js)
make test-e2e-setup  # First time only
make test-e2e

# Run E2E tests interactively
make test-e2e-ui

# Run with coverage
make test-go-coverage
# Opens coverage.html in browser
```

See [tests/e2e/README.md](tests/e2e/README.md) for more E2E testing details.

## Deployment

### Local Deployment

```bash
make run-local
```

This will:
1. Build all binaries
2. Initialize database at `~/.config/task/task.db`
3. Start server on port 8080

### Docker Deployment

```bash
# Build and start all services
make docker-up

# View logs
docker-compose logs -f

# Stop services
make docker-down
```

Services:
- **server**: REST API and web UI (port 8080)
- **orchestrator**: Task orchestration daemon
- **caddy**: Reverse proxy (ports 80/443)

### Production Considerations

- Change default admin password immediately
- Use HTTPS (Caddy handles this automatically)
- Back up SQLite database regularly
- Monitor orchestrator and worker health via heartbeat API
- Set appropriate configuration values for intervals
- Consider using external authentication service
- Scale workers horizontally as needed

## Contributing

1. Create a feature branch from `develop`
2. Make changes
3. Run tests: `make test`
4. Commit changes
5. Merge to `develop` using `--no-ff`
6. For releases, merge `develop` to `main`

## License

[License details to be added]

## Support

For issues, questions, or contributions:
- See [User Guide](USER_GUIDE.md) for detailed usage
- Check [Design Documentation](docs/) for architecture details
- Review [API Specification](api-specification.yaml) for API reference
