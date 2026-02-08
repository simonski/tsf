# TODO

- implement the API and client calls to manage ALL Entities via CRUD calls.  Once this is complete we can say the entire ticketing and entity management system is finished as there will be
  - openAPI APIs implemented
  - database schemas completed
  - client API calls via the terminal testable
Once this is complete we can then implement WORKERS and the ORCHESTRATOR


## Recent Improvements (Feb 2026)

- [x] **Terminal User Interface (TUI)** - Implemented interactive TUI for task management (commit: 62c1fc1)
  - Added bubbletea, lipgloss, and bubbles dependencies for TUI framework
  - Created internal/tui package with complete TUI implementation
  - Implemented login screen with authentication
  - Added main menu navigation with 6 primary sections (Projects, Tasks, Roles, Users, Config, Workers)
  - Implemented full CRUD operations for all entities
  - Added form handling with Tab/Shift-Tab navigation
  - Implemented keyboard-driven navigation (arrows/WASD, space, esc)
  - Added consistent styling with lipgloss
  - Integrated with existing CLI client for API communication
  - Added 'sf tui' command to unified binary
  - Updated README.md and DESIGN_TUI.md with comprehensive documentation

- [x] **Config Admin Panel** - Added admin-only config panel with CRUD functionality (commit: 9b80b44)
  - Added Config navigation link (visible to admin only)
  - Created config screen with table display showing key/value/description
  - Implemented add, edit, and delete operations for config entries
  - Added CSS styling for config table
  - Config panel is only accessible to users with username 'admin'

- [x] Fixed Playwright navigation test class assertion (/active/ → /open/)

- [x] Added comprehensive Requirements section to README with installation instructions 
- [x] Fixed `make test` hanging after Playwright tests complete (added proper cleanup)
- [x] Upgraded Docker image to Go 1.24 (matches local Go 1.24.2)
- [x] Implemented Playwright E2E testing infrastructure
- [x] Added authentication, navigation, and project management tests
- [x] Updated Makefile with E2E test targets
- [x] Created comprehensive E2E testing documentation
- [x] Fixed passkey credential verification bug (challenge storage and retrieval)
- [x] Implemented passkey (WebAuthn) authentication for passwordless login
- [x] Added database schema for passkey credentials and sessions
- [x] Created passkey registration and authentication endpoints
- [x] Updated frontend with passkey support in login and settings screens
- [x] Added comprehensive tests for passkey functionality
- [x] Updated API specification with passkey endpoints
- [x] Updated documentation with passkey usage instructions
- [x] Verified argon2id is already implemented for secure password hashing (not bcrypt)
- [x] Added .env.example file for easier environment setup
- [x] Enhanced .gitignore with additional common patterns
- [x] Created comprehensive CONTRIBUTING.md for contributors
- [x] Created quickstart.sh script for easy onboarding
- [x] Updated README with quickstart instructions
- [x] Fixed go.mod indirect dependency issue for jwt package
- [x] Implemented heartbeat storage in database with proper schema and methods
- [x] Enhanced task assignment logic with intelligent priority-based selection
- [x] Implemented role fetching and assignment for worker task requests
- [x] Added GET /api/v1/heartbeats endpoint for monitoring worker/orchestrator health
- [x] Updated API specification with heartbeats endpoint and schema
- [x] Updated documentation to use single `task` binary consistently
- [x] Replaced curl examples with CLI commands in documentation
- [x] All tests passing, build successful

## Completed

✅ bug ux: the overview once loaded no longer shows hte panel when the hamburger is clicked

✅ add an overview link in the settings panel - the page should show the configuration and workers in a threejs 2d graph

## Next Tasks

## Fixed Bugs

- [x] UX: Move website title to the right of hamburger icon in header

- [x] UX: Focus on username field on login page
- [x] UX: Focus on username field on register page

- [x] CLI uses single-hyphen options (-password, -url, -username) with --force as the only double-hyphen exception

- [x] BUG website - cannot login as admin using generated password - FIXED: Updated frontend to use correct endpoint /api/v1/auth/me instead of /api/v1/users/me
- [x] BUG website - cannot register says auth - FIXED: Updated frontend to use correct endpoint /api/v1/auth/register instead of /api/v1/users 

## Completed Improvements

- [x] Add `-password` option to `sf initdb` to set custom admin password
- [x] Create sliding panel navigation with hamburger menu (projects, users, settings, logout)
- [x] `sf initdb --force` - add --force option to rebuild database (removes and recreates)
- [x] `task` should render the whole usage
- [x] `make build` should build the binary to ./, not ./bin  
- [x] `make` should print all make targets (implemented as `make help`)
- [x] merge all binaries to be a SINGLE binary `task` with commands that run them
    - `sf server` - Start HTTP server
    - `sf orchestrator` - Start orchestrator daemon
    - `task (client)` - CLI commands (scaffolded)
    - `sf initdb` - Initialize database
    - `sf worker` - Start worker daemon
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
