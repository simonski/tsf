# TODO


## Fixed Bugs

- [x] UX: Move website title to the right of hamburger icon in header

- [x] UX: Focus on username field on login page
- [x] UX: Focus on username field on register page

- [x] CLI uses single-hyphen options (-password, -url, -username) with --force as the only double-hyphen exception

- [x] BUG website - cannot login as admin using generated password - FIXED: Updated frontend to use correct endpoint /api/v1/auth/me instead of /api/v1/users/me
- [x] BUG website - cannot register says auth - FIXED: Updated frontend to use correct endpoint /api/v1/auth/register instead of /api/v1/users 

## Completed Improvements

- [x] Add `-password` option to `task initdb` to set custom admin password
- [x] Create sliding panel navigation with hamburger menu (projects, users, settings, logout)
- [x] `task initdb --force` - add --force option to rebuild database (removes and recreates)
- [x] `task` should render the whole usage
- [x] `make build` should build the binary to ./, not ./bin  
- [x] `make` should print all make targets (implemented as `make help`)
- [x] merge all binaries to be a SINGLE binary `task` with commands that run them
    - `task server` - Start HTTP server
    - `task orchestrator` - Start orchestrator daemon
    - `task (client)` - CLI commands (scaffolded)
    - `task initdb` - Initialize database
    - `task worker` - Start worker daemon
- [x] implement and verify the -url option in the CLI calls, the -json option too
- [x] implement 100% of the CLI methods
- [x] BUG FIX: the server now correctly renders the web UI at http://localhost:8080

## Next Priority Tasks

### 1. Database Schema & Initialization [COMPLETED]
- [x] Create SQL schema file with all tables (users, projects, roles, tasks, task_history, project_members, config)
- [x] Implement Go database package with models and connection management
- [x] Implement CLI initdb command
- [x] Add unit tests for database initialization
- [x] Verify all entity definitions match ENTITY_*.md specifications

### 2. OpenAPI Specification [COMPLETED]
- [x] Create /api-specification.yaml with OpenAPI 3.0 spec
- [x] Document all authentication requirements (Basic Auth)
- [x] Document all CRUD endpoints for users, projects, roles, tasks
- [x] Document config management endpoints
- [x] Document worker/orchestrator registration and heartbeat endpoints

### 3. Go Server Backend [COMPLETED]
- [x] Create HTTP server with routing
- [x] Implement Basic Auth middleware
- [x] Implement all CRUD handlers matching OpenAPI spec
- [x] Add database interaction layer
- [x] Add request validation and error handling
- [x] Write unit tests for all handlers
- [x] Write integration tests
- [x] Verify all tests pass with `make test-go`

### 4. CLI Client Infrastructure [COMPLETED]
- [x] Implement HTTP client with Basic Auth using environment variables
- [x] Add config management for server URL and credentials
- [x] Implement project context management
- [x] Add JSON and human-readable output formatting
- [x] Write CLI command infrastructure

### 5. Orchestrator Implementation [COMPLETED]
- [x] Create orchestrator daemon
- [x] Implement registration and heartbeat mechanism
- [x] Implement task query and routing logic
- [x] Implement worker assignment logic (scaffolded, TODO comments for future implementation)
- [x] Implement role assignment logic (scaffolded)
- [x] Support standalone mode with config from server
- [x] Add signal handling for graceful shutdown

### 6. Worker Implementation [COMPLETED]
- [x] Create worker daemon
- [x] Implement registration and heartbeat mechanism
- [x] Implement work request loop
- [x] Integrate task processing (simulated, ready for LLM integration)
- [x] Implement role-based context injection (scaffolded)
- [x] Add error handling and failure reporting
- [x] Add signal handling for graceful shutdown

### 7. Frontend Implementation [COMPLETED]
- [x] Create login/register page
- [x] Create project dropdown selector
- [x] Create kanban board visualization with 4 columns
- [x] Create task detail modal with status updates
- [x] Embed in Go binary using go:embed
- [x] Integrate static file serving in server
- [x] Test frontend integration

### 8. Deployment Infrastructure [COMPLETED]
- [x] Create Dockerfile with multi-stage build
- [x] Create Makefile with docker targets (build, up, down, run-local)
- [x] Create Caddy configuration for reverse proxy
- [x] Create Docker Compose configuration with services
- [x] Add volume configuration for SQLite persistence
- [x] Test Docker build successfully

### 9. Documentation [COMPLETED]
- [x] Create README.md with project overview and quick start
- [x] Create USER_GUIDE.md with comprehensive usage documentation
- [x] Update TODO.md as work progresses
