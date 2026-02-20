The SERVER serves both the FRONTEND and the APIs.

The APIs are restful openAPI spec compatible for realtime access to current WORK state.

The server is really only a CRUD manager of state.  It will not make many "decisions" on what to do.

The SERVER is the ONLY component that read/writes the database.

## API

A CRUD api exists to manage tasks creation, it also exposes the openAPI specification.   

Versioning: use an openAPI /v1/ style to start with.

### Entity Endpoints

The server exposes CRUD endpoints for core entities and sub-entities:

- Projects: `/api/v1/projects`
- Project files: `/api/v1/projects/{project_id}/files`
- Project notes: `/api/v1/projects/{project_id}/notes`
- Tasks: `/api/v1/tasks`
- Entity comments (project/task): `/api/v1/comments`

### Comment Model Requirements

Entity comments are attached to either a project or a task and must support:

- owner identity (`owner_id`, `owner_username`)
- timestamps (`created_at`, `updated_at`)
- soft-delete (`is_deleted`, `deleted_at`)
- immutable history (`create`, `edit`, `soft_delete`)

Ownership rule:
- only the comment owner can edit or soft-delete a comment

## Usage

```bash
# runs the server on the default server port
./sf server

# runs the server on the default server port, with explicit database
./sf server -f mydb.db
```

## OpenAPI

Each endpoint must have a fully documentated section in the OpenAPI specification.  

## Authentication

All calls to the server MUST be Basic-Auth OR using a session token once the user has logged in.

The credentials must be encrypted - argon2id.
