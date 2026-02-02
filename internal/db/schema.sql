-- Task Management System Database Schema
-- SQLite3 compatible

-- Users table: humans, workers, and orchestrators
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('human', 'worker', 'orchestrator')),
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_type ON users(type);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Projects table: workspaces containing tasks
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    repository TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'inactive')),
    visibility TEXT NOT NULL DEFAULT 'public' CHECK(visibility IN ('public', 'internal', 'private')),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (updated_by) REFERENCES users(id)
);

CREATE INDEX idx_projects_name ON projects(name);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_visibility ON projects(visibility);
CREATE INDEX idx_projects_created_by ON projects(created_by);

-- Project members: many-to-many relationship between projects and users
CREATE TABLE IF NOT EXISTS project_members (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(project_id, user_id)
);

CREATE INDEX idx_project_members_project ON project_members(project_id);
CREATE INDEX idx_project_members_user ON project_members(user_id);

-- Roles table: job descriptions for workers
CREATE TABLE IF NOT EXISTS roles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    rules TEXT NOT NULL,
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

-- Unique constraint: system/global roles have globally unique names
CREATE UNIQUE INDEX idx_roles_name_global ON roles(name) WHERE scope IN ('system', 'global');
-- Unique constraint: project roles have unique names within project
CREATE UNIQUE INDEX idx_roles_name_project ON roles(name, project_id) WHERE scope = 'project';
CREATE INDEX idx_roles_scope ON roles(scope);
CREATE INDEX idx_roles_project ON roles(project_id);
CREATE INDEX idx_roles_is_active ON roles(is_active);

-- Tasks table: units of work
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    title TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('epic', 'story', 'task', 'sub-task', 'bug', 'spike')),
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
    labels TEXT, -- JSON array stored as text
    estimated_effort INTEGER,
    actual_effort INTEGER,
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

-- Task history: work sessions on tasks
CREATE TABLE IF NOT EXISTS task_history (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    started_at TEXT NOT NULL,
    completed_at TEXT,
    state TEXT NOT NULL CHECK(state IN ('success', 'failure', 'abandoned', 'in-progress')),
    worker_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    notes TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (worker_id) REFERENCES users(id),
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

CREATE INDEX idx_task_history_task ON task_history(task_id);
CREATE INDEX idx_task_history_worker ON task_history(worker_id);
CREATE INDEX idx_task_history_role ON task_history(role_id);
CREATE INDEX idx_task_history_state ON task_history(state);
CREATE INDEX idx_task_history_started ON task_history(started_at);

-- Configuration table: system settings
CREATE TABLE IF NOT EXISTS config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_config_key ON config(key);

-- Refresh tokens for JWT authentication
CREATE TABLE IF NOT EXISTS refresh_tokens (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at);

-- Heartbeats table: track worker and orchestrator activity
CREATE TABLE IF NOT EXISTS heartbeats (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    status TEXT NOT NULL,
    task_id TEXT,
    last_seen TEXT NOT NULL DEFAULT (datetime('now')),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX idx_heartbeats_user ON heartbeats(user_id);
CREATE INDEX idx_heartbeats_last_seen ON heartbeats(last_seen);
CREATE INDEX idx_heartbeats_status ON heartbeats(status);

-- Passkey credentials table: store WebAuthn credentials
CREATE TABLE IF NOT EXISTS passkey_credentials (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    credential_id BLOB NOT NULL UNIQUE,
    public_key BLOB NOT NULL,
    attestation_type TEXT NOT NULL,
    aaguid BLOB NOT NULL,
    sign_count INTEGER NOT NULL DEFAULT 0,
    clone_warning INTEGER NOT NULL DEFAULT 0,
    transports TEXT, -- JSON array of transport types
    backup_eligible INTEGER NOT NULL DEFAULT 0,
    backup_state INTEGER NOT NULL DEFAULT 0,
    device_name TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    last_used_at TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_passkey_credentials_user ON passkey_credentials(user_id);
CREATE INDEX idx_passkey_credentials_credential_id ON passkey_credentials(credential_id);
CREATE INDEX idx_passkey_credentials_last_used ON passkey_credentials(last_used_at);

-- WebAuthn sessions table: store temporary session data during registration/authentication
CREATE TABLE IF NOT EXISTS webauthn_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    challenge BLOB NOT NULL,
    user_verification TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    session_type TEXT NOT NULL CHECK(session_type IN ('registration', 'authentication')),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_webauthn_sessions_user ON webauthn_sessions(user_id);
CREATE INDEX idx_webauthn_sessions_expires ON webauthn_sessions(expires_at);
CREATE INDEX idx_webauthn_sessions_challenge ON webauthn_sessions(challenge);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_users_timestamp 
    AFTER UPDATE ON users
    FOR EACH ROW
    BEGIN
        UPDATE users SET updated_at = datetime('now') WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_projects_timestamp 
    AFTER UPDATE ON projects
    FOR EACH ROW
    BEGIN
        UPDATE projects SET updated_at = datetime('now') WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_roles_timestamp 
    AFTER UPDATE ON roles
    FOR EACH ROW
    BEGIN
        UPDATE roles SET updated_at = datetime('now') WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_tasks_timestamp 
    AFTER UPDATE ON tasks
    FOR EACH ROW
    BEGIN
        UPDATE tasks SET updated_at = datetime('now') WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_config_timestamp 
    AFTER UPDATE ON config
    FOR EACH ROW
    BEGIN
        UPDATE config SET updated_at = datetime('now') WHERE key = NEW.key;
    END;
