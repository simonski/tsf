# Implementation Plan: Align Code with DESIGN_CLI.md

**Status:** Ready to implement  
**Created:** 2026-02-08  
**Authority:** DESIGN_CLI.md (all specifications must match this document)

## Decisions Made

1. ✅ Task field: `title` (update schema from `title` to match, ENTITY docs updated)
2. ✅ Project primary key: Keep as `id` in DB, use `project_id` in API/CLI
3. ✅ Comments: JSONB column in tasks table: `comments JSONB`
4. ✅ Blocked by: `sf task update -task_id A -blocked_by B` → sets A.depends_on_task_id = B
5. ✅ Session tokens: Store JWT in `~/.config/sf/session.json`

---

## Phase 1: Database Schema Updates (Foundation Layer)

### Task 1.1: Update tasks table
```sql
-- Add new columns
ALTER TABLE tasks ADD COLUMN comments TEXT; -- JSON stored as text in SQLite
ALTER TABLE tasks ADD COLUMN is_deleted INTEGER NOT NULL DEFAULT 0;

-- Update type constraint to match DESIGN_CLI
-- SQLite doesn't support ALTER on CHECK constraints, so this requires table recreation
-- Change: ('epic', 'story', 'task', 'sub-task', 'bug', 'spike')
-- To: ('epic', 'task', 'bug', 'spike', 'chore')
```

**Files to modify:**
- `internal/db/schema.sql` - Update tasks table definition
- `internal/db/models.go` - Add Comments, IsDeleted fields to Task struct

### Task 1.2: Update roles table
```sql
-- Rename rules column to goals
ALTER TABLE roles RENAME COLUMN rules TO goals;
```

**Files to modify:**
- `internal/db/schema.sql` - Rename column
- `internal/db/models.go` - Rename Rules → Goals field

### Task 1.3: Add session tokens table
```sql
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
```

**Files to modify:**
- `internal/db/schema.sql` - Add sessions table
- `internal/db/models.go` - Add Session struct

---

## Phase 2: Server API Implementation (20 endpoints)

### Task 2.1: Worker Management APIs
**New endpoints:**
- `POST /api/v1/workers` - Create worker (admin only)
- `GET /api/v1/workers` - List workers (admin only)
- `GET /api/v1/workers/{worker_id}` - Get worker (admin only)
- `POST /api/v1/workers/{worker_id}/enable` - Enable worker (admin only)
- `POST /api/v1/workers/{worker_id}/disable` - Disable worker (admin only)
- `POST /api/v1/workers/{worker_id}/reset-password` - Reset password (admin only)

**Files to create/modify:**
- `internal/server/workers.go` - Add new handlers
- `internal/server/server.go` - Register routes

### Task 2.2: Task Comment APIs
**Update endpoints:**
- `POST /api/v1/tasks/{task_id}/comments` - Add comment
- `GET /api/v1/tasks/{task_id}/comments` - Get all comments

**Files to modify:**
- `internal/server/tasks.go` - Add comment handlers
- `internal/server/server.go` - Register routes

### Task 2.3: Worker Task APIs
**New endpoints:**
- `POST /api/v1/tasks/request` - Worker requests task (worker only)
- `POST /api/v1/tasks/{task_id}/return` - Worker returns task (worker only)

**Files to modify:**
- `internal/server/tasks.go` - Add handlers
- `internal/server/server.go` - Register routes

### Task 2.4: Task Assignment with Role
**Update endpoint:**
- `POST /api/v1/tasks/{task_id}/assign` - Now accepts `worker_id` AND `role_id`

**Files to modify:**
- `internal/server/tasks.go` - Update handleAssignTask

### Task 2.5: Task Unassign API
**New endpoint:**
- `POST /api/v1/tasks/{task_id}/unassign` - Unassign worker from task

**Files to modify:**
- `internal/server/tasks.go` - Add handler
- `internal/server/server.go` - Register route

### Task 2.6: Task Update with Blocked By
**Update endpoint:**
- `PUT /api/v1/tasks/{task_id}` - Support `blocked_by` field

**Files to modify:**
- `internal/server/tasks.go` - Update handleUpdateTask

### Task 2.7: Task Soft Delete
**Update endpoint:**
- `DELETE /api/v1/tasks/{task_id}` - Change to set is_deleted=true

**Files to modify:**
- `internal/server/tasks.go` - Update handleDeleteTask
- All task query endpoints - Add `WHERE is_deleted = 0` filter

### Task 2.8: Role History API
**New endpoint:**
- `GET /api/v1/roles/{role_id}/history` - Get role usage history

**Files to modify:**
- `internal/server/roles.go` - Add handler
- `internal/server/server.go` - Register route

### Task 2.9: User Password Reset API
**New endpoint:**
- `POST /api/v1/users/{username}/reset-password` - Reset user password (admin only)

**Files to modify:**
- `internal/server/users.go` - Add handler
- `internal/server/server.go` - Register route

### Task 2.10: Session Token APIs
**New endpoints:**
- `POST /api/v1/auth/login` - Update to return session token
- `POST /api/v1/auth/logout` - Update to invalidate session token
- Middleware: Support Bearer token authentication

**Files to modify:**
- `internal/server/auth.go` - Add session management
- `internal/server/middleware.go` - Add token validation

---

## Phase 3: CLI Implementation (27 commands)

### Task 3.1: Add Worker Command Group
**New commands:**
```bash
sf worker create -worker_id XXX (-password YYYY)
sf worker list
sf worker enable -worker_id XXX
sf worker disable -worker_id XXX
sf worker reset-password -worker_id XXX -password YYYY
```

**Files to modify:**
- `main.go` - Add handleWorkerCommand function
- `main.go` - Update command routing

### Task 3.2: Standardize Task Flag Names
**Change all occurrences:**
- From: `-id` 
- To: `-task_id`

**Affected commands:**
```bash
sf task get -task_id A
sf task update -task_id A ...
sf task delete -task_id A
sf task claim -task_id A
sf task free -task_id A
sf task assign -task_id A ...
sf task complete -task_id A
sf task history -task_id A
sf task deps -task_id A
```

**Files to modify:**
- `main.go` - Update handleTaskCommand function

### Task 3.3: Add Task Comment Command
**New command:**
```bash
sf task comment -task_id A -comment "The comment"
```

**Files to modify:**
- `main.go` - Add to handleTaskCommand

### Task 3.4: Add Worker Task Commands
**New commands:**
```bash
sf task request   # worker-only
sf task return    # worker-only
```

**Files to modify:**
- `main.go` - Add to handleTaskCommand

### Task 3.5: Update Task Assign Command
**Update command:**
```bash
sf task assign -task_id Y -worker_id XXXXX -role role_id
```

**Files to modify:**
- `main.go` - Update task assign handler to include -role flag

### Task 3.6: Add Task Unassign Command
**New command:**
```bash
sf task unassign -task_id Y -worker_id XXXXX
```

**Files to modify:**
- `main.go` - Add to handleTaskCommand

### Task 3.7: Add Task Blocked By Support
**Update command:**
```bash
sf task update -task_id A -blocked_by B
```

**Files to modify:**
- `main.go` - Add -blocked_by flag to task update

### Task 3.8: Rename Config Delete Command
**Change:**
- From: `sf config delete -key K`
- To: `sf config rm -key KEY`

**Files to modify:**
- `main.go` - Update handleConfigCommand

### Task 3.9: Update Project Commands
**Update commands:**
```bash
sf project create -project_id XXX -name XXX -description XXX
sf project set-default -project_id XXX
sf project get-default
sf project unset-default -project_id XXX
```

**Changes:**
- Add `-project_id` parameter to create
- Rename `set` → `set-default`
- Add `get-default` command
- Rename `unset` → `unset-default`, add `-project_id`

**Files to modify:**
- `main.go` - Update handleProjectCommand

### Task 3.10: Add Role History Command
**New command:**
```bash
sf role history -role_id XXXX
```

**Files to modify:**
- `main.go` - Add to handleRoleCommand

### Task 3.11: Add User Reset Password Command
**New command:**
```bash
sf user reset-password -username XXX -password YYYY
```

**Files to modify:**
- `main.go` - Add to handleUserCommand

### Task 3.12: Implement Session Token Storage
**Features:**
- Login stores token to `~/.config/sf/session.json`
- Subsequent commands read token from file
- Send as Bearer token in Authorization header
- Fallback to Basic Auth if no token

**Files to modify:**
- `internal/cli/config.go` - Add session token management
- `internal/cli/client.go` - Add token authentication
- `main.go` - Update handleLoginCommand

---

## Phase 4: Database Migration

### Task 4.1: Create Migration Script
**Create:**
- `internal/db/migrations/001_align_with_design_cli.sql`

**Contents:**
1. Add is_deleted to tasks
2. Add comments to tasks
3. Rename roles.rules to roles.goals
4. Update task types enum
5. Create sessions table

**Files to create:**
- Migration script

### Task 4.2: Update Database Initialization
**Update:**
- `internal/db/init.go` - Apply new schema
- Create default session config values

**Files to modify:**
- `internal/db/init.go`

---

## Phase 5: Testing & Documentation

### Task 5.1: Update Unit Tests
**Files to update:**
- `internal/db/db_test.go`
- `internal/server/*_test.go`
- Add tests for new endpoints

### Task 5.2: Update Integration Tests
**Files to update:**
- `tests/e2e/specs/*.spec.ts`

### Task 5.3: Update Documentation
**Files to update:**
- `USER_GUIDE.md` - Document all new commands
- `README.md` - Update examples
- `docs/DESIGN_SERVER.md` - Document new APIs

---

## Execution Order (Dependency-Based)

**Week 1: Foundation**
1. Phase 1 (Tasks 1.1, 1.2, 1.3) - Schema updates
2. Phase 4, Task 4.1 - Migration script
3. Phase 4, Task 4.2 - Update init

**Week 2: Server APIs**
4. Phase 2, Task 2.10 - Session tokens (needed by all)
5. Phase 2, Task 2.1 - Worker APIs
6. Phase 2, Task 2.7 - Soft delete (affects queries)
7. Phase 2, Tasks 2.2-2.6, 2.8-2.9 - Remaining APIs

**Week 3: CLI**
8. Phase 3, Task 3.12 - Session token storage (needed by all)
9. Phase 3, Task 3.2 - Standardize flags
10. Phase 3, Tasks 3.1, 3.3-3.11 - All commands

**Week 4: Testing & Polish**
11. Phase 5, Tasks 5.1-5.3 - Tests and docs

---

## Checklist Summary

### Database (3 tasks)
- [ ] Update tasks table (comments, is_deleted, types)
- [ ] Update roles table (rules → goals)
- [ ] Add sessions table

### Server APIs (20 endpoints)
- [ ] 6 worker management endpoints
- [ ] 2 task comment endpoints
- [ ] 2 worker task endpoints (request/return)
- [ ] 1 task assignment update (add role)
- [ ] 1 task unassign endpoint
- [ ] 1 task update (blocked_by support)
- [ ] 1 task soft delete update
- [ ] 1 role history endpoint
- [ ] 1 user password reset endpoint
- [ ] 4 session token endpoints/updates

### CLI Commands (27 items)
- [ ] 5 worker commands
- [ ] 10 task command updates (flag names)
- [ ] 1 task comment command
- [ ] 2 worker task commands
- [ ] 1 task assign update
- [ ] 1 task unassign command
- [ ] 1 task blocked_by support
- [ ] 1 config rename (delete → rm)
- [ ] 3 project command updates
- [ ] 1 role history command
- [ ] 1 user reset password command
- [ ] 1 session token implementation

### Testing & Docs (3 tasks)
- [ ] Update unit tests
- [ ] Update integration tests
- [ ] Update documentation

---

## Total: 53 Implementation Tasks

**Estimated Time:** 3-4 weeks for complete implementation
**Priority:** High - Critical for system consistency
**Risk:** Medium - Schema changes require careful migration
