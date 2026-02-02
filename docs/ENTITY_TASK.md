# Entity: TASK

## Purpose

Represents a unit of work that progresses through a lifecycle until completely finished. Tasks form a hierarchy (epic > story > task > sub-task) and track dependencies, ownership, and work history.

The lifecycle of a task is that it will be worked on in some form until it is completely finished. It will be "passed" between WORKERs as it progresses, where each time it is passed it will have a new entry in its history. Eventually all the WORKERs will complete their jobs and the task will be deemed complete and finished.

## Fields

- `id`: uuid (primary key) - Unique identifier for the task
- `project_id`: uuid (required, fk to projects.id) - The project this task belongs to
- `title`: string (required, max 200 chars) - Single-sentence summary of the task
- `type`: enum (epic, story, task, sub-task, bug, spike) - Classification of work
  - `epic`: Large body of work spanning multiple stories
  - `story`: User-facing feature or requirement
  - `task`: Technical work item
  - `sub-task`: Breakdown of a task
  - `bug`: Defect to be fixed
  - `spike`: Research or investigation
- `description`: text (required) - Multi-paragraph detailed description
- `acceptance_criteria`: text (nullable) - Multi-paragraph definition of done
- `parent_id`: uuid (nullable, fk to tasks.id) - Parent task in hierarchy
- `epic_id`: uuid (nullable, fk to tasks.id) - Top-level epic this task belongs to (denormalized for query efficiency)
- `depends_on_task_id`: uuid (nullable, fk to tasks.id) - Task that must complete before this one can start
- `status`: enum (idle, active) - Current working status
  - `idle`: No worker currently assigned or working on this task
  - `active`: A worker is currently assigned and working on this task
- `worker_id`: uuid (nullable, fk to users.id) - Worker currently assigned to this task
- `priority`: enum (low, medium, high, critical) - Urgency level for orchestrator prioritization
- `is_complete`: boolean (default false) - True only when task is totally finished and accepted
- `completed_at`: datetime (nullable) - When task was marked complete (only set when is_complete=true)
- `created_at`: datetime (required, immutable) - When task was created
- `updated_at`: datetime (required) - Last modification timestamp
- `created_by`: uuid (required, fk to users.id) - User who created the task
- `updated_by`: uuid (required, fk to users.id) - User who last updated the task
- `labels`: text[] (nullable) - Array of tags for categorization and fast lookup
- `estimated_effort`: integer (nullable) - Estimated story points or hours
- `actual_effort`: integer (nullable) - Actual time spent (sum of all work sessions)

## Relationships

- **Project**: Many-to-one (task.project_id references projects.id)
- **Parent**: Many-to-one self-reference (task.parent_id references tasks.id)
- **Children**: One-to-many self-reference (inverse of parent_id)
- **Epic**: Many-to-one self-reference (task.epic_id references tasks.id where type=epic)
- **Dependency**: Many-to-one self-reference (task.depends_on_task_id references tasks.id)
- **Blocked Tasks**: One-to-many (tasks that depend on this one)
- **Worker**: Many-to-one (task.worker_id references users.id)
- **Creator**: Many-to-one (task.created_by references users.id)
- **Updater**: Many-to-one (task.updated_by references users.id)
- **History**: One-to-many (task_history.task_id references tasks.id)

## Task History (Separate Table)

Work sessions are tracked in a separate `task_history` table for better queryability:

- `id`: uuid (primary key) - Unique identifier for this history entry
- `task_id`: uuid (required, fk to tasks.id) - The task this history belongs to
- `started_at`: datetime (required) - When work began on this task
- `completed_at`: datetime (nullable) - When work ended (null if still in progress)
- `state`: enum (success, failure, abandoned, in-progress) - Outcome of this work session
  - `success`: Work completed successfully, task progressed
  - `failure`: Work attempted but failed, task needs rework
  - `abandoned`: Work stopped without completion
  - `in-progress`: Currently active work session
- `worker_id`: uuid (required, fk to users.id) - Worker who performed this work
- `role_id`: uuid (required, fk to roles.id) - Role the worker was operating under
- `notes`: text (nullable) - What was done, learned, or why it failed
- `created_at`: datetime (required, immutable) - When this history entry was created

## Indexes

- Primary: `id`
- Index: `project_id` (for project task listings)
- Index: `status` (for filtering active/idle tasks)
- Index: `worker_id` (for worker's current tasks)
- Index: `is_complete` (for filtering completed tasks)
- Index: `parent_id` (for hierarchy queries)
- Index: `epic_id` (for epic rollup queries)
- Index: `depends_on_task_id` (for dependency resolution)
- Index: `created_by` (for user's created tasks)
- Index: `labels` (GIN index for array search)
- Index: `priority` (for orchestrator prioritization)

## Constraints

- `title` cannot be empty
- `parent_id` cannot reference self (no circular parent)
- `depends_on_task_id` cannot reference self
- `epic_id` must reference a task where type=epic
- `worker_id` must be null when status=idle
- `worker_id` must be set when status=active
- `completed_at` must be null when is_complete=false
- `completed_at` must be set when is_complete=true
- Cannot have circular dependencies (validated at application layer)
- `priority` defaults to `medium`
- `status` defaults to `idle`

## Validation Rules

- Title: 1-200 characters, no leading/trailing whitespace
- Description: Minimum 10 characters
- Parent must be in same project
- Epic must be in same project
- Dependency must be in same project
- Cannot mark complete if depends_on_task is not complete
- Cannot mark complete if any child tasks are not complete
- Cannot change `created_at` or `created_by` after creation
- Worker must have appropriate permissions for the project

## State Transitions

```
Created (idle) → Assigned (active) → Working (active) → Complete (idle, is_complete=true)
                      ↓                     ↓
                    Freed (idle)      Failed → Reassigned (active)
```

## Lifecycle

1. **Created**: Task created with status=idle, worker_id=null
2. **Assigned**: Orchestrator or human assigns worker, status=active, worker_id set, history entry created
3. **Working**: Worker performs work, updates task with progress/notes
4. **Completed/Failed**: Work session ends, history entry updated with state and completed_at
5. **Freed**: Worker releases task, status=idle, worker_id=null (task may be reassigned)
6. **Finished**: Task marked is_complete=true, completed_at set, status=idle, worker_id=null
