-- Migration 001: Align Database with DESIGN_CLI.md
-- Created: 2026-02-08
-- Purpose: Update schema to match DESIGN_CLI specification

-- This migration updates the database schema to align with DESIGN_CLI.md requirements:
-- 1. Update task types: remove 'story' and 'sub-task', add 'chore'
-- 2. Add 'comments' JSONB field to tasks table
-- 3. Add 'is_deleted' field to tasks table for soft deletes
-- 4. Rename 'rules' to 'goals' in roles table
-- 5. Add sessions table for JWT authentication

-- SQLite doesn't support ALTER TABLE with CHECK constraints or column renaming easily
-- So we need to recreate affected tables

-- =============================================================================
-- STEP 1: Update tasks table
-- =============================================================================

-- Create new tasks table with updated schema
CREATE TABLE IF NOT EXISTS tasks_new (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    title TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('epic', 'task', 'bug', 'spike', 'chore')),
    description TEXT NOT NULL,
    acceptance_criteria TEXT,
    parent_id TEXT,
    epic_id TEXT,
    depends_on_task_id TEXT,
    status TEXT NOT NULL DEFAULT 'idle' CHECK(status IN ('idle', 'active')),
    worker_id TEXT,
    priority TEXT NOT NULL DEFAULT 'medium' CHECK(priority IN ('low', 'medium', 'high', 'critical')),
    is_complete INTEGER NOT NULL DEFAULT 0,
    completed_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    labels TEXT,
    estimated_effort INTEGER,
    actual_effort INTEGER,
    comments TEXT, -- JSON array stored as text
    is_deleted INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE SET NULL,
    FOREIGN KEY (epic_id) REFERENCES tasks(id) ON DELETE SET NULL,
    FOREIGN KEY (depends_on_task_id) REFERENCES tasks(id) ON DELETE SET NULL,
    FOREIGN KEY (worker_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (updated_by) REFERENCES users(id),
    CHECK ((status = 'idle' AND worker_id IS NULL) OR (status = 'active' AND worker_id IS NOT NULL)),
    CHECK ((is_complete = 0 AND completed_at IS NULL) OR (is_complete = 1 AND completed_at IS NOT NULL))
);

-- Copy data from old table to new table
-- Note: Tasks with type 'story' or 'sub-task' will be converted to 'task'
INSERT INTO tasks_new (
    id, project_id, title, 
    type,
    description, acceptance_criteria, parent_id, epic_id, depends_on_task_id,
    status, worker_id, priority, is_complete, completed_at,
    created_at, updated_at, created_by, updated_by,
    labels, estimated_effort, actual_effort, comments, is_deleted
)
SELECT 
    id, project_id, title,
    CASE 
        WHEN type IN ('story', 'sub-task') THEN 'task'
        ELSE type
    END as type,
    description, acceptance_criteria, parent_id, epic_id, depends_on_task_id,
    status, worker_id, priority, is_complete, completed_at,
    created_at, updated_at, created_by, updated_by,
    labels, estimated_effort, actual_effort, 
    NULL as comments,  -- Initialize comments as NULL
    0 as is_deleted    -- Initialize is_deleted as 0 (false)
FROM tasks;

-- Drop old table
DROP TABLE tasks;

-- Rename new table to tasks
ALTER TABLE tasks_new RENAME TO tasks;

-- Recreate indexes on tasks table
CREATE INDEX idx_tasks_project ON tasks(project_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_worker ON tasks(worker_id);
CREATE INDEX idx_tasks_is_complete ON tasks(is_complete);
CREATE INDEX idx_tasks_parent ON tasks(parent_id);
CREATE INDEX idx_tasks_epic ON tasks(epic_id);
CREATE INDEX idx_tasks_depends_on ON tasks(depends_on_task_id);
CREATE INDEX idx_tasks_created_by ON tasks(created_by);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_type ON tasks(type);
CREATE INDEX idx_tasks_is_deleted ON tasks(is_deleted);

-- =============================================================================
-- STEP 2: Update roles table (rename rules to goals)
-- =============================================================================

CREATE TABLE IF NOT EXISTS roles_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    goals TEXT NOT NULL,
    scope TEXT NOT NULL CHECK(scope IN ('system', 'global', 'project')),
    project_id TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (updated_by) REFERENCES users(id),
    CHECK ((scope = 'project' AND project_id IS NOT NULL) OR (scope != 'project' AND project_id IS NULL))
);

-- Copy data from old table to new table
INSERT INTO roles_new (
    id, name, description, goals, scope, project_id, is_active,
    created_at, updated_at, created_by, updated_by
)
SELECT 
    id, name, description, rules as goals, scope, project_id, is_active,
    created_at, updated_at, created_by, updated_by
FROM roles;

-- Drop old table
DROP TABLE roles;

-- Rename new table to roles
ALTER TABLE roles_new RENAME TO roles;

-- Recreate indexes on roles table
CREATE UNIQUE INDEX idx_roles_name_global ON roles(name) WHERE scope IN ('system', 'global');
CREATE UNIQUE INDEX idx_roles_name_project ON roles(name, project_id) WHERE scope = 'project';
CREATE INDEX idx_roles_scope ON roles(scope);
CREATE INDEX idx_roles_project ON roles(project_id);
CREATE INDEX idx_roles_is_active ON roles(is_active);

-- =============================================================================
-- STEP 3: Add sessions table
-- =============================================================================

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

-- =============================================================================
-- STEP 4: Recreate triggers for updated tables
-- =============================================================================

-- Recreate trigger for tasks table
DROP TRIGGER IF EXISTS update_tasks_timestamp;
CREATE TRIGGER update_tasks_timestamp 
    AFTER UPDATE ON tasks
    FOR EACH ROW
    BEGIN
        UPDATE tasks SET updated_at = datetime('now') WHERE id = NEW.id;
    END;

-- Recreate trigger for roles table
DROP TRIGGER IF EXISTS update_roles_timestamp;
CREATE TRIGGER update_roles_timestamp 
    AFTER UPDATE ON roles
    FOR EACH ROW
    BEGIN
        UPDATE roles SET updated_at = datetime('now') WHERE id = NEW.id;
    END;

-- =============================================================================
-- Migration complete
-- =============================================================================
