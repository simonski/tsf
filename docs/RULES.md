## Workflow Commands

**@XXX** means "read this file XXX"

### Continue Command
Typing **"continue"** means reread the following files to continue creating and extending the project:
- `@docs/RULES.md`
- `@docs/DESIGN.md`
- `@/TODO.md`

### Propose/Suggest/Next Commands
Typing **"propose"**, **"suggest"**, or **"next"** means reread the following files to suggest what should be done next, but **DO NOT DO** the work:
- `@docs/RULES.md`
- `@docs/DESIGN.md`
- `@/TODO.md`

## Project Documentation

**MUST** maintain these top-level documentation files:
- `README.md` - Explains in simple terms what the project is and how to build it
- `USER_GUIDE.md` - Explains all use cases and functionality
- `TODO.md` - Explains tasks carried out and to be carried out

**Generated documentation:**
- Write generated markdown to `docs/generated/`
- Exception: `README.md` and `TODO.md` remain at project root

## TODO Workflow

1. **ALWAYS** pick the first incomplete TODO from `TODO.md`
2. Work on that TODO until complete
3. Each TODO should be in its own feature branch off `develop`
4. Once complete, merge the feature and update `TODO.md` with the merge commit hash
5. When no more TODOs exist, read `docs/DESIGN.md` for additional work

## Git Workflow

### Branch Naming
**ALWAYS** use these branch naming patterns:

| Branch Type | Pattern | Example |
|-------------|---------|---------|
| Feature | `feature/<descriptive-name>` | `feature/user-auth` |
| Bugfix | `bugfix/<descriptive-name>` | `bugfix/login-error` |
| Hotfix | `hotfix/<descriptive-name>` | `hotfix/security-patch` |

**Incorrect:** `go`, `auth`, `new-feature` (missing prefix)

### Branch Creation
**ALWAYS** create feature branches from `develop`:

```bash
# Create new feature branch
git checkout develop
git pull
git checkout -b feature/<descriptive-name>

# Push feature branch
git push -u origin feature/<descriptive-name>
```

### Merge Process
**Branch hierarchy:** `feature` → `develop` → `main`

**Complete merge workflow:**
1. Ensure all tests pass: `make test`
2. Rebase feature onto latest develop:
   ```bash
   git checkout feature/<name>
   git fetch origin
   git rebase origin/develop
   ```
3. Verify tests still pass after rebase: `make test`
4. Merge into develop:
   ```bash
   git checkout develop
   git pull
   git merge --ff-only feature/<name>
   ```
5. Update `TODO.md` with the merge commit hash from develop
6. Push develop:
   ```bash
   git push origin develop
   ```
7. Delete feature branch:
   ```bash
   git branch -d feature/<name>
   git push origin --delete feature/<name>
   ```

**Rules:**
- **ONLY** merge when feature is complete and all tests pass
- **NEVER** merge failing tests into develop
- **ALWAYS** rebase feature branches before merging to maintain clean history

## Build System

### Makefile Requirements
- **MUST** have a `Makefile` with these targets: `build`, `clean`, `test`
- Running `make` alone **MUST** print usage information
- **ALWAYS** use `make build` for building
- **ALWAYS** use `make test` for testing

### Build Workflow
When starting a new feature or TODO:
1. Create feature branch from `develop`
2. Use `make build` for building
3. Use `make test` for testing
4. Fix all tests before committing
5. Push to feature branch
6. Only merge when feature is complete and all tests pass

## Testing

### Testing Requirements
- **ALWAYS** run `make test` before committing
- **NEVER** commit code that does not pass tests
- **NEVER** merge if tests are failing
- **DO NOT** stop if a test fails - fix the failing test before finishing
- Strive for comprehensive test coverage

### Test Harness
All backend/database code and backend APIs **MUST** have unit and integration test suites before testing the JavaScript frontend.

### Available Test Targets
```bash
make test                 # Run all test suites (JavaScript + Go)
make test-go              # Run all Go backend tests (unit + integration)
make test-go-unit         # Run Go unit tests only (auth, models, database, handlers)
make test-integration     # Run Go integration tests only
make test-go-coverage     # Run Go tests with coverage report
```

**All test targets should be run on new features.**

### Testing Guidelines
- Use and extend the test harnesses
- **DO NOT** test "conversationally" (iterative manual testing via chat)
- **NEVER** commit code that causes failing tests
- Fix tests immediately when they fail

## API Development

### OpenAPI Specification
- `api-specification.yaml` **MUST** be maintained and kept up to date
- **ALL** APIs **MUST** be described in the specification
- The specification **MUST** be used with the implementation of server APIs

### API Testing
- 100% of APIs **MUST** have unit and integration tests for:
  - Go backend
  - Go client
  - JavaScript frontend

## Communication Style

- Limit use of emojis in responses
- Don't compliment or apologize
- No hyperbole
