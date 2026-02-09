# Entity: PROJECT

## Purpose
Represents a workspace that contains related tasks. Projects provide organizational boundaries, access control, and git repository integration for context-aware task management.

## Fields

- `id`: TEXT PRIMARY KEY - Unique identifier for the project (database column name is 'id', but referred to as 'project_id' in API/CLI)
- `name`: string (required, max 200 chars, unique) - Single sentence name
- `description`: text (required) - Multi-paragraph description of project goals and context
- `repository`: string (nullable, max 500 chars) - URL to git repository for code context
- `status`: enum (active, inactive) - Indicates if orchestrator should process tasks from this project
  - `active`: Orchestrator will assign and process tasks
  - `inactive`: Tasks remain visible but are not actively processed
- `visibility`: enum (public, internal, private) - Access control level
  - `public`: Any authenticated user can view and contribute
  - `internal`: Any authenticated user can view but only members can contribute
  - `private`: Only members can view or contribute
- `created_at`: datetime (required, immutable) - When project was created
- `updated_at`: datetime (required) - Last modification timestamp
- `created_by`: uuid (required, fk to users.id) - User who created the project
- `updated_by`: uuid (required, fk to users.id) - User who last updated the project

## Relationships

- **Members**: Many-to-many via `project_members` table
  - Links projects to users with specific roles/permissions
  - Only applies to `internal` and `private` projects
  - For `public` projects, membership list is informational only
- **Tasks**: One-to-many (task.project_id references project.id)
  - Cascade behavior: Restrict deletion if tasks exist (must delete tasks first)
- **Creator**: Many-to-one (project.created_by references users.id)
- **Updater**: Many-to-one (project.updated_by references users.id)

## Indexes

- Primary: `id`
- Unique: `name`
- Index: `status` (for orchestrator queries)
- Index: `visibility` (for access control checks)
- Index: `created_by` (for user's projects listing)

## Constraints

- `name` must be unique across all projects
- `name` cannot be empty or whitespace only
- `visibility` defaults to `public`
- `status` defaults to `active`
- `repository` must be valid URL format if provided

## Validation Rules

- Name: 1-200 characters, no leading/trailing whitespace
- Description: Minimum 10 characters
- Repository: Must match URL pattern if not null
- Cannot change `created_at` or `created_by` after creation

## Default Projects

- A project named `default` with `visibility=public` and `status=active` is created during database initialization

