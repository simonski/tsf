package server

import (
	"database/sql"
	"net/http"

	"github.com/simonski/task/internal/db"
)

// WorkerRequestResponse represents a worker request response
type WorkerRequestResponse struct {
	Task interface{} `json:"task,omitempty"`
	Role interface{} `json:"role,omitempty"`
}

// HeartbeatRequest represents a heartbeat request
type HeartbeatRequest struct {
	Status string  `json:"status"`
	TaskID *string `json:"task_id,omitempty"`
}

// handleWorkerRequest handles worker work requests
func (s *Server) handleWorkerRequest(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if user.Type != "worker" {
		sendError(w, http.StatusForbidden, "only workers can request work")
		return
	}

	// Find an available task that is:
	// 1. Not complete (is_complete = 0)
	// 2. Status is 'idle' (not assigned to any worker)
	// 3. Has no blocking dependencies or dependencies are complete
	// Priority: critical > high > medium > low
	query := `
		SELECT id, project_id, title, type, description, acceptance_criteria, parent_id, epic_id,
		       depends_on_task_id, status, worker_id, priority, is_complete, completed_at,
		       created_at, updated_at, created_by, updated_by, labels, estimated_effort, actual_effort
		FROM tasks
		WHERE is_complete = 0 
		  AND status = 'idle'
		  AND (depends_on_task_id IS NULL 
		       OR depends_on_task_id IN (SELECT id FROM tasks WHERE is_complete = 1))
		ORDER BY 
			CASE priority 
				WHEN 'critical' THEN 1
				WHEN 'high' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'low' THEN 4
			END,
			created_at ASC
		LIMIT 1
	`

	row := s.db.Conn().QueryRow(query)
	task, err := scanTask(row)
	if err != nil {
		// No work available
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Assign task to this worker
	_, err = s.db.Conn().Exec(`
		UPDATE tasks 
		SET status = 'active', worker_id = ?, updated_by = ?
		WHERE id = ?
	`, user.ID, user.ID, task.ID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to assign task")
		return
	}

	// Create task history entry
	historyID := task.ID + "-history-" + user.ID

	// Fetch an appropriate role for this task
	// Priority: project-scoped role > global role > system role
	var roleID string
	roleQuery := `
		SELECT id FROM roles 
		WHERE is_active = 1 
		  AND (
			(scope = 'project' AND project_id = ?) OR
			(scope = 'global') OR
			(scope = 'system')
		  )
		ORDER BY 
			CASE scope 
				WHEN 'project' THEN 1
				WHEN 'global' THEN 2
				WHEN 'system' THEN 3
			END
		LIMIT 1
	`
	err = s.db.Conn().QueryRow(roleQuery, task.ProjectID).Scan(&roleID)
	if err != nil {
		// Use a placeholder if no role found
		roleID = "default-role"
	}

	_, err = s.db.Conn().Exec(`
		INSERT INTO task_history (id, task_id, started_at, state, worker_id, role_id)
		VALUES (?, ?, datetime('now'), 'in-progress', ?, ?)
	`, historyID, task.ID, user.ID, roleID)
	if err != nil {
		// Log error but don't fail the request
		// The task was already assigned
	}

	// Fetch the role details to return to worker
	var role *db.Role
	if roleID != "default-role" {
		var r db.Role
		var projectID sql.NullString
		var createdAt, updatedAt string
		err = s.db.Conn().QueryRow(`
			SELECT id, name, description, rules, scope, project_id, is_active, created_at, updated_at, created_by, updated_by
			FROM roles WHERE id = ?
		`, roleID).Scan(
			&r.ID, &r.Name, &r.Description, &r.Rules, &r.Scope,
			&projectID, &r.IsActive, &createdAt, &updatedAt,
			&r.CreatedBy, &r.UpdatedBy,
		)
		if err == nil {
			r.CreatedAt = parseTimestamp(createdAt)
			r.UpdatedAt = parseTimestamp(updatedAt)
			if projectID.Valid {
				r.ProjectID = &projectID.String
			}
			role = &r
		}
	}

	// Return the assigned task with role information
	response := WorkerRequestResponse{
		Task: task,
		Role: role,
	}

	sendJSON(w, http.StatusOK, response)
}

// handleWorkerHeartbeat handles worker heartbeat
func (s *Server) handleWorkerHeartbeat(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req HeartbeatRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Store heartbeat in database
	if err := s.db.UpsertHeartbeat(user.ID, req.Status, req.TaskID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to store heartbeat")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleOrchestratorHeartbeat handles orchestrator heartbeat
func (s *Server) handleOrchestratorHeartbeat(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req HeartbeatRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Store heartbeat in database
	if err := s.db.UpsertHeartbeat(user.ID, req.Status, req.TaskID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to store heartbeat")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleListHeartbeats returns all recent heartbeats for monitoring
func (s *Server) handleListHeartbeats(w http.ResponseWriter, r *http.Request) {
	heartbeats, err := s.db.ListHeartbeats()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to list heartbeats")
		return
	}

	sendJSON(w, http.StatusOK, heartbeats)
}
