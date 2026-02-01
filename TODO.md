# TODO

- `task` should render the whole usage
- `make build` should build the binary to ./, not ./bin
- `make` should print all make targets
- merge all binaries to be a SINGLE binary `task` with commands the run them
    task server
    task orchestrator
    task (client)
    task initdb
    task worker

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

### 4. CLI Client [feature/cli-client]
- [ ] Implement all commands from DESIGN_CLI.md
- [ ] Add Basic Auth using environment variables
- [ ] Implement project context management
- [ ] Add JSON and human-readable output modes
- [ ] Write unit tests for CLI logic
- [ ] Write integration tests against running server

### 5. Orchestrator [feature/orchestrator]
- [ ] Create orchestrator daemon
- [ ] Implement registration and heartbeat mechanism
- [ ] Implement task query and routing logic
- [ ] Implement worker assignment logic
- [ ] Implement role assignment logic
- [ ] Support standalone and embedded modes
- [ ] Add tests

### 6. Worker [feature/worker]
- [ ] Create worker daemon
- [ ] Implement registration and heartbeat mechanism
- [ ] Implement work request loop
- [ ] Integrate LLM provider
- [ ] Implement role-based context injection
- [ ] Add error handling and failure reporting
- [ ] Add tests

### 7. Frontend [feature/frontend]
- [ ] Create login/register page
- [ ] Create project dropdown selector
- [ ] Create kanban board visualization
- [ ] Create task detail view
- [ ] Embed in Go binary using go:embed
- [ ] Test frontend integration

### 8. Deployment Infrastructure [feature/deployment]
- [ ] Create Makefile with all targets
- [ ] Create Caddy configuration
- [ ] Create Docker Compose configuration
- [ ] Add volume configuration for SQLite
- [ ] Test local deployment
- [ ] Test Docker deployment

### 9. Documentation
- [ ] Create README.md
- [ ] Create USER_GUIDE.md
- [ ] Update TODO.md as work progresses
