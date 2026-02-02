## General Rules

@XXX means "read this file XXX"

- typing "continue" means reread 
    @docs/RULES.md
    @docs/DESIGN.md
    @/TODO.md 
  to continue creating and extending the project.
- typing "propose", "suggest", "next" means reread 
    @docs/RULES.md
    @docs/DESIGN.md
    @/TODO.md
  to continue creating and extending the project but DO NOT DO the work - suggest what should be done next
- Maintain a top-level USER_GUIDE.md that is to explain all use cases and functionality.   
- Maintain a top-level README.md such that it explains in simple terms what the project is and how to build it.

- NEVER write to /tmp

- when starting a new feature
  - create a feature branch from develop
  - always use `make build` for building
  - always use `make test` for testing
  - push to the feature branch first
  - only merge the feature is complete and once all tests pass via `make test`

- maintain the /TODO.md explaining the tasks carried out and to be carried out.  
  - Pick from the first TODO and work on that until it is complete
  - each todo should be in its own feature branch off develop
  - once complete the feature should be merged with the TODO updated with the last hash on that todo
  - when there are no more TODOs, read the docs/DESIGN.md for more work
- limit use of emojis in responses
- don't compliment or apologise.  No hyperbole please.
- any other generated documentation write to docs/generated/

## Git, Testing

### Branch Naming and Workflow

- use git
- **ALWAYS** create feature branches with the naming pattern: `feature/<descriptive-name>`
  - Correct: `feature/go`, `feature/user-auth`, `feature/3d-rendering`
  - Wrong: `go`, `auth`, `new-feature`
- **ALWAYS** create feature branches from `develop`: `git checkout develop && git pull && git checkout -b feature/<name>`
- merge hierarchy: feature → develop → main
- only merge to develop once the feature is complete and all tests pass in the feature branch
- always run `make test` before committing
- never commit code that does not pass tests
- ensure the test suite passes before committing code
- strive for comprehensive test coverage
- do not merge if tests are failing
- do not stop if a test fails - fix the failing test before you say you've finished

### Git Commands Reference
```bash
# Create new feature branch (ALWAYS use this pattern)
git checkout develop
git pull
git checkout -b feature/<descriptive-name>

# Push feature branch
git push -u origin feature/<descriptive-name>
```

## Makefile

- there should be a Makefile (make build clean test)
- `make` on its own should print a usage

## Documentation

- any generated markdown should be written to docs/generated/.... 
- except for /README.md and /TODO.md

##  Testing

Test harness - all backend/database code / all backend APIs should have a unit and integration test suite before testing the javascript frontend.  

Available test targets:
- `make test` - Run all test suites (JavaScript + Go)
- `make test-go` - Run all Go backend tests (unit + integration)
- `make test-go-unit` - Run Go unit tests only (auth, models, database, handlers)
- `make test-integration` - Run Go integration tests only
- `make test-go-coverage` - Run Go tests with coverage report

All test targets should be run on new features. Do not test "conversationally" - use and extend the test harnesses and do not commit code that is causing failing tests.

## APIS

- /api-specification.yaml openAPI spec MUST be maintained and up to date with all APIs described.   The specification should be used with the implementation of the server APIS.
- 100% of the APIS should have unit and integration tests for both the go backend, the tui client and the javascript front end
