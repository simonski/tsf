
# Entity: ROLE

## Purpose
Defines a "job description" or persona that shapes how workers approach and execute tasks. Roles provide context, rules, and motivation to LLM-based workers, allowing the same worker to operate as different specialists (e.g., Programmer, Tester, DevOps) based on the assigned role.

## Examples
- Product Owner
- Product Manager  
- Business Analyst
- Programmer
- DevOps Engineer
- Tester
- Release Manager
- Technical Writer
- Security Analyst

## Fields

- `id`: uuid (primary key) - Unique identifier for the role
- `name`: string (required, unique, max 100 chars) - Short name (e.g., "programmer", "tester")
- `description`: text (required) - Multi-paragraph job description explaining the role's purpose and responsibilities
- `goals`: text (required) - Instructions and context for LLM workers operating in this role. This becomes the system prompt that motivates behavior.
- `scope`: enum (system, global, project) - Visibility and applicability of the role
  - `system`: Built-in role, cannot be modified or deleted by users
  - `global`: User-defined role available across all projects
  - `project`: Role specific to a single project
- `project_id`: uuid (nullable, fk to projects.id) - Required if scope=project, null otherwise
- `is_active`: boolean (default true) - Whether this role can be assigned to workers
- `created_at`: datetime (required, immutable) - When role was created
- `updated_at`: datetime (required) - Last modification timestamp
- `created_by`: uuid (required, fk to users.id) - User who created the role
- `updated_by`: uuid (required, fk to users.id) - User who last updated the role

## Relationships

- **Project**: Many-to-one (role.project_id references projects.id) when scope=project
- **Task History**: One-to-many (task_history.role_id references role.id)
- **Creator**: Many-to-one (role.created_by references users.id)
- **Updater**: Many-to-one (role.updated_by references users.id)

## Indexes

- Primary: `id`
- Unique: `name` (within scope - system roles have globally unique names)
- Index: `scope` (for filtering available roles)
- Index: `project_id` (for project-specific role queries)
- Index: `is_active` (for active role queries)

## Constraints

- `name` must be unique for system and global scopes
- `name` must be unique within a project for project-scoped roles
- `project_id` must be null if scope is system or global
- `project_id` must be set if scope is project
- System roles cannot be deleted or have scope changed
- `goals` cannot be empty

## Validation Rules

- Name: 1-100 characters, lowercase, alphanumeric with hyphens/underscores
- Description: Minimum 20 characters
- Rules: Minimum 50 characters (enough for meaningful LLM context)
- Cannot change `created_at` or `created_by` after creation
- Cannot change `scope` after creation

## Usage by Orchestrator

The ORCHESTRATOR assigns both WORK (task) and ROLE to workers. This allows workers to be dynamically repurposed:
- Same worker can function as Programmer on one task, Tester on another
- The `goals` field provides LLM context that motivates appropriate behavior
- Role assignment is recorded in task history for auditability

## Default System Roles

During database initialization, these system roles are created:
- `programmer`: Implements code based on requirements
- `tester`: Writes and executes tests to verify functionality
- `analyst`: Breaks down requirements and creates specifications
- `reviewer`: Reviews work quality and provides feedback

