Use a sqlite backend that is accessed only via the server.

## Database Location








































































- The admin user is always created first, used as creator for projects and roles- All passwords are hashed using argon2id before storage- Roles can reference projects created in projects.md- If a file doesn't exist or is empty, it's silently skipped- Files are processed in order: users, workers, orchestrators, projects, roles## Notes```sf initdb --force --password dev --populate# Combine with other optionssf initdb --populate# Populate database during initialization```bash## Usage- `project_name` - Project name (required for project-scoped roles, empty for system/global)- `scope` - Role scope: `system`, `global`, or `project`- `goals` - Detailed goals/responsibilities for the role- `description` - Brief description- `name` - Role name (unique within scope)**Format:** `name | description | goals | scope | project_name`Creates additional roles.### roles.md- `visibility` - Project visibility: `public`, `internal`, or `private`- `status` - Project status: `active` or `inactive`- `repository` - Git repository URL (optional, can be empty)- `description` - Project description- `name` - Unique project name**Format:** `name | description | repository | status | visibility`Creates additional projects.### projects.md- `password` - Password (will be hashed)- `username` - Unique username**Format:** `username | password`Creates additional orchestrator users.### orchestrators.md- `password` - Password (will be hashed)- `username` - Unique username**Format:** `username | password`Creates additional worker users.### workers.md- `password` - Password (will be hashed)- `type` - User type (must be "human")- `username` - Unique username**Format:** `username | type | password`Creates additional human users.### users.md## FilesEach file uses a pipe-delimited format within markdown code blocks (` ``` `). Lines outside code blocks are ignored, allowing for documentation.## File Format
The default database location is `~/.config/sf/sf.db`, following the XDG Base Directory specification. This provides a standard location for user data that:
- Persists across code updates
- Separates user data from application code
- Is easily backed up
- Can be overridden with the `-f` flag

## Initialize database

```bash
# writes to the location specified with -f
./sf initdb -f /path/to/database.db

# defaults to ~/.config/sf/sf.db
./sf initdb 

# defaults to ~/.config/sf/sf.db, --force overwrites
./sf initdb --force

# set the same password for all users
./sf initdb --password mypassword

# populate database from scripts/initdb/*.md files
./sf initdb --populate

# combine options
./sf initdb --force --password mypassword --populate

```

Note: `sf initdb` is the ONLY command which touches the database directly as it is initialising the database.

Initialisation will:
- Create the database directory if it doesn't exist
- Initialize schema with all required tables
- Create an ADMIN user with a generated password (printed to stdout)
- Create a normal USER with a generated password (printed to stdout)
- Create 1 WORKER with a generated password (printed to stdout)
- Create 1 ORCHESTRATOR with a generated password (printed to stdout)
- Create a default project
- Set up system configuration values

Options:
- `--force` - Remove existing database and create a new one
- `--password <PASSWORD>` - Set the same password for all users (admin, user, worker, orchestrator)
- `--populate` - Populate database from embedded markdown files (compiled into the binary at `internal/db/scripts/initdb/`):
  - `users.md` - Additional human users (format: `username | type | password`)
  - `workers.md` - Additional worker users (format: `username | password`)
  - `orchestrators.md` - Additional orchestrator users (format: `username | password`)
  - `projects.md` - Additional projects (format: `name | description | repository | status | visibility`)
  - `roles.md` - Additional roles (format: `name | description | goals | scope | project_name`)
- `--populate` - Populate database from markdown files in scripts/initdb/

## Populate from Scripts

The `--populate` option reads markdown files from `scripts/initdb/` and creates additional entities:

- `scripts/initdb/users.md` - Create additional human users
- `scripts/initdb/workers.md` - Create additional worker users
- `scripts/initdb/orchestrators.md` - Create additional orchestrator users
- `scripts/initdb/projects.md` - Create additional projects
- `scripts/initdb/roles.md` - Create additional roles (system, global, or project-scoped)

Each file uses a simple pipe-delimited format within markdown code blocks. See the example files for format details.

Initialisation creates default config values in the config table:

```bash
orchestrator.heartbeat=1000
orchestrator.idle=10000
worker.heartbeat=1000
worker.idle=10000
```

creates default project `default`, public.
