# Contributing to sf

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to the project.

## Development Setup

### Prerequisites

- Go 1.23 or later
- Git
- SQLite3
- Docker and Docker Compose (optional, for containerized development)

### Getting Started

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/your-username/task.git
   cd task
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Build the project**
   ```bash
   make build
   ```

4. **Initialize the database**
   ```bash
   # With random passwords (printed to stdout)
   ./sf initdb -f ~/.config/sf/sf.db
   
   # Or with a specific password for all users
   ./sf initdb -f ~/.config/sf/sf.db --password dev
   
   # With test data from scripts/initdb/*.md
   ./sf initdb -f ~/.config/sf/sf.db --password dev --populate
   ```

5. **Run tests**
   ```bash
   make test
   ```

## Project Structure

```
.
├── cmd/                    # Command binaries
│   └── task-unified/      # Single unified binary
├── internal/              # Internal packages
│   ├── cli/              # CLI implementation
│   ├── db/               # Database layer
│   ├── orchestrator/     # Orchestrator logic
│   ├── server/           # Server handlers
│   ├── web/              # Embedded web assets
│   └── worker/           # Worker logic
├── web/                   # Web frontend source
├── docs/                  # Documentation
├── api-specification.yaml # OpenAPI specification
└── Makefile              # Build automation
```

## Development Workflow

### Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, idiomatic Go code
   - Follow existing code style and conventions
   - Add tests for new functionality
   - Update documentation as needed

3. **Run tests**
   ```bash
   make test
   ```

4. **Run with coverage**
   ```bash
   make test-go-coverage
   ```

5. **Build and test manually**
   ```bash
   make build
   ./sf server -f test.db -port 8080
   ```

### Code Style

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and concise
- Handle errors explicitly

### Testing

All changes must include appropriate tests:

#### Unit Tests (Go)
- Write unit tests for new functionality
- Maintain or improve code coverage
- Test edge cases and error conditions
- Use table-driven tests where appropriate

Example test structure:
```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"case 1", "input1", "output1", false},
        {"case 2", "input2", "output2", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Feature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Feature() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Feature() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### E2E Tests (Playwright)
For frontend changes, add or update E2E tests:

```bash
# Setup (first time only)
make test-e2e-setup

# Run E2E tests
make test-e2e

# Interactive mode
make test-e2e-ui
```

See [tests/e2e/README.md](tests/e2e/README.md) for detailed E2E testing guidelines.

#### Running Tests

```bash
# All tests (Go + E2E)
make test

# Go tests only
make test-go

# E2E tests only
make test-e2e

# With coverage
make test-go-coverage
```

### Commit Messages

Use clear, descriptive commit messages:

- Start with a verb in imperative mood (Add, Fix, Update, Remove)
- Keep the first line under 50 characters
- Add detailed description if needed after a blank line

Examples:
```
Add heartbeat monitoring endpoint

Implement GET /api/v1/heartbeats endpoint to list all worker and
orchestrator heartbeats for monitoring system health.
```

### Pull Requests

1. **Update your branch**
   ```bash
   git fetch origin
   git rebase origin/main
   ```

2. **Push your changes**
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create a pull request**
   - Provide a clear title and description
   - Reference any related issues
   - Ensure all tests pass
   - Wait for code review

## API Changes

If you're modifying the API:

1. Update `api-specification.yaml` with your changes
2. Update relevant documentation
3. Ensure backward compatibility when possible
4. Document breaking changes clearly

## Documentation

- Update README.md for user-facing changes
- Update USER_GUIDE.md for new features
- Update relevant DESIGN_*.md files for architectural changes
- Add inline code comments for complex logic

## Running Services

### Local Development

```bash
# Run server
./sf server -f ~/.config/sf/sf.db -port 8080

# Run orchestrator (in another terminal)
export SF_USERNAME=orchestrator
export SF_PASSWORD=your-password
./sf orchestrator -url http://localhost:8080

# Run worker (in another terminal)
export SF_USERNAME=worker1
export SF_PASSWORD=your-password
./sf worker -url http://localhost:8080
```

### Docker Development

```bash
# Build and start all services
make docker-up

# View logs
docker-compose logs -f

# Stop services
make docker-down
```

## Common Tasks

### Adding a New API Endpoint

1. Define the endpoint in `api-specification.yaml`
2. Add the route in `internal/server/server.go`
3. Implement the handler in the appropriate file
4. Add database operations if needed
5. Write tests in `internal/server/server_test.go`
6. Update documentation

### Adding a New CLI Command

1. Add command handling in `cmd/task-unified/cli.go`
2. Implement the command logic
3. Update usage text in `printCLIUsage()`
4. Add examples to USER_GUIDE.md

### Database Schema Changes

1. Update `internal/db/schema.sql`
2. Update models in `internal/db/models.go`
3. Add database methods in `internal/db/db.go`
4. Write migration logic if needed
5. Update tests

## Getting Help

- Check existing [documentation](docs/)
- Review [User Guide](USER_GUIDE.md)
- Look at [existing issues](https://github.com/your-org/task/issues)
- Ask questions in discussions

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Help others learn and grow

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
