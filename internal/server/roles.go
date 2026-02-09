package server

import (
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"github.com/simonski/task/internal/db"
)

// RoleRequest represents a role creation/update request
type RoleRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Goals       string  `json:"goals"`
	Scope       string  `json:"scope"`
	ProjectID   *string `json:"project_id,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// handleListRoles returns all accessible roles
func (s *Server) handleListRoles(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, name, description, goals, scope, project_id, is_active, created_at, updated_at, created_by, updated_by
		FROM roles
		WHERE 1=1
	`
	args := []interface{}{}

	if scope := r.URL.Query().Get("scope"); scope != "" {
		query += " AND scope = ?"
		args = append(args, scope)
	}
	if projectID := r.URL.Query().Get("project_id"); projectID != "" {
		query += " AND (scope != 'project' OR project_id = ?)"
		args = append(args, projectID)
	}

	query += " ORDER BY scope, name"

	rows, err := s.db.Conn().Query(query, args...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query roles")
		return
	}
	defer rows.Close()

	var roles []db.Role
	for rows.Next() {
		var role db.Role
		var projectID sql.NullString
		var createdAt, updatedAt string
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.Goals,
			&role.Scope,
			&projectID,
			&role.IsActive,
			&createdAt,
			&updatedAt,
			&role.CreatedBy,
			&role.UpdatedBy,
		)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan role")
			return
		}
		role.CreatedAt = parseTimestamp(createdAt)
		role.UpdatedAt = parseTimestamp(updatedAt)
		if projectID.Valid {
			role.ProjectID = &projectID.String
		}
		roles = append(roles, role)
	}

	sendJSON(w, http.StatusOK, roles)
}

// handleGetRoleHistory returns task history for a specific role
func (s *Server) handleGetRoleHistory(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("role_id")

	// Verify role exists
	var exists int
	err := s.db.Conn().QueryRow(`SELECT 1 FROM roles WHERE id = ?`, roleID).Scan(&exists)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "role not found")
		return
	}

	query := `
		SELECT h.id, h.task_id, h.started_at, h.completed_at, h.state,
		       h.worker_id, h.result, h.summary,
		       t.title as task_title,
		       u.username as worker_name
		FROM task_history h
		LEFT JOIN tasks t ON h.task_id = t.id
		LEFT JOIN users u ON h.worker_id = u.id
		WHERE h.role_id = ?
		ORDER BY h.started_at DESC
	`

	rows, err := s.db.Conn().Query(query, roleID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query history")
		return
	}
	defer rows.Close()

	type HistoryEntry struct {
		ID          string  `json:"id"`
		TaskID      string  `json:"task_id"`
		TaskTitle   string  `json:"task_title,omitempty"`
		StartedAt   string  `json:"started_at"`
		CompletedAt *string `json:"completed_at,omitempty"`
		State       string  `json:"state"`
		WorkerID    string  `json:"worker_id"`
		WorkerName  string  `json:"worker_name,omitempty"`
		Result      *string `json:"result,omitempty"`
		Summary     *string `json:"summary,omitempty"`
	}

	history := []HistoryEntry{}
	for rows.Next() {
		var entry HistoryEntry
		var completedAt, result, summary, taskTitle, workerName sql.NullString

		err := rows.Scan(
			&entry.ID, &entry.TaskID, &entry.StartedAt, &completedAt, &entry.State,
			&entry.WorkerID, &result, &summary, &taskTitle, &workerName,
		)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan history")
			return
		}

		if completedAt.Valid {
			entry.CompletedAt = &completedAt.String
		}
		if result.Valid {
			entry.Result = &result.String
		}
		if summary.Valid {
			entry.Summary = &summary.String
		}
		if taskTitle.Valid {
			entry.TaskTitle = taskTitle.String
		}
		if workerName.Valid {
			entry.WorkerName = workerName.String
		}

		history = append(history, entry)
	}

	sendJSON(w, http.StatusOK, history)
}

// handleCreateRole creates a new role
func (s *Server) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req RoleRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Description == "" || req.Goals == "" || req.Scope == "" {
		sendError(w, http.StatusBadRequest, "name, description, goals, and scope are required")
		return
	}

	if req.Scope != "global" && req.Scope != "project" {
		sendError(w, http.StatusBadRequest, "scope must be 'global' or 'project'")
		return
	}

	if req.Scope == "project" && (req.ProjectID == nil || *req.ProjectID == "") {
		sendError(w, http.StatusBadRequest, "project_id is required for project-scoped roles")
		return
	}

	roleID := uuid.New().String()
	_, err := s.db.Conn().Exec(`
		INSERT INTO roles (id, name, description, goals, scope, project_id, is_active, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, roleID, req.Name, req.Description, req.Goals, req.Scope, req.ProjectID, true, user.ID, user.ID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create role")
		return
	}

	var role db.Role
	var projectID sql.NullString
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT id, name, description, goals, scope, project_id, is_active, created_at, updated_at, created_by, updated_by
		FROM roles
		WHERE id = ?
	`, roleID).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Goals,
		&role.Scope,
		&projectID,
		&role.IsActive,
		&createdAt,
		&updatedAt,
		&role.CreatedBy,
		&role.UpdatedBy,
	)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve role")
		return
	}
	role.CreatedAt = parseTimestamp(createdAt)
	role.UpdatedAt = parseTimestamp(updatedAt)
	if projectID.Valid {
		role.ProjectID = &projectID.String
	}

	sendJSON(w, http.StatusCreated, role)
}

// handleGetRole returns a specific role
func (s *Server) handleGetRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("role_id")

	var role db.Role
	var projectID sql.NullString
	var createdAt, updatedAt string
	err := s.db.Conn().QueryRow(`
		SELECT id, name, description, goals, scope, project_id, is_active, created_at, updated_at, created_by, updated_by
		FROM roles
		WHERE id = ?
	`, roleID).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Goals,
		&role.Scope,
		&projectID,
		&role.IsActive,
		&createdAt,
		&updatedAt,
		&role.CreatedBy,
		&role.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "role not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query role")
		return
	}
	role.CreatedAt = parseTimestamp(createdAt)
	role.UpdatedAt = parseTimestamp(updatedAt)
	if projectID.Valid {
		role.ProjectID = &projectID.String
	}

	sendJSON(w, http.StatusOK, role)
}

// handleUpdateRole updates a role
func (s *Server) handleUpdateRole(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	roleID := r.PathValue("role_id")

	// Check if it's a system role
	var scope string
	err := s.db.Conn().QueryRow("SELECT scope FROM roles WHERE id = ?", roleID).Scan(&scope)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "role not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query role")
		return
	}
	if scope == "system" {
		sendError(w, http.StatusBadRequest, "cannot modify system roles")
		return
	}

	var req RoleRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err = s.db.Conn().Exec(`
		UPDATE roles
		SET name = COALESCE(?, name),
		    description = COALESCE(?, description),
		    goals = COALESCE(?, goals),
		    is_active = COALESCE(?, is_active),
		    updated_by = ?
		WHERE id = ?
	`, req.Name, req.Description, req.Goals, req.IsActive, user.ID, roleID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to update role")
		return
	}

	var role db.Role
	var projectID sql.NullString
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT id, name, description, goals, scope, project_id, is_active, created_at, updated_at, created_by, updated_by
		FROM roles
		WHERE id = ?
	`, roleID).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Goals,
		&role.Scope,
		&projectID,
		&role.IsActive,
		&createdAt,
		&updatedAt,
		&role.CreatedBy,
		&role.UpdatedBy,
	)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve role")
		return
	}
	role.CreatedAt = parseTimestamp(createdAt)
	role.UpdatedAt = parseTimestamp(updatedAt)
	if projectID.Valid {
		role.ProjectID = &projectID.String
	}

	sendJSON(w, http.StatusOK, role)
}

// handleDeleteRole deletes a role
func (s *Server) handleDeleteRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("role_id")

	// Check if it's a system role
	var scope string
	err := s.db.Conn().QueryRow("SELECT scope FROM roles WHERE id = ?", roleID).Scan(&scope)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "role not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query role")
		return
	}
	if scope == "system" {
		sendError(w, http.StatusBadRequest, "cannot delete system roles")
		return
	}

	// Check if role is in use
	var count int
	err = s.db.Conn().QueryRow("SELECT COUNT(*) FROM task_history WHERE role_id = ?", roleID).Scan(&count)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to check role usage")
		return
	}
	if count > 0 {
		sendError(w, http.StatusBadRequest, "cannot delete role in use")
		return
	}

	result, err := s.db.Conn().Exec("DELETE FROM roles WHERE id = ?", roleID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to delete role")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		sendError(w, http.StatusNotFound, "role not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
