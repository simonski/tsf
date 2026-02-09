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
			SELECT id, name, description, goals, scope, project_id, is_active, created_at, updated_at, created_by, updated_by
			FROM roles WHERE id = ?
		`, roleID).Scan(
			&r.ID, &r.Name, &r.Description, &r.Goals, &r.Scope,
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

// WorkerResponse represents a worker entity in API responses
type WorkerResponse struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	Status      string  `json:"status"`
	LastSeen    string  `json:"last_seen,omitempty"`
	CurrentTask *string `json:"current_task,omitempty"`
}

// handleListWorkers returns all workers
func (s *Server) handleListWorkers(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT u.id, u.username, 
		       COALESCE(h.status, 'offline') as status,
		       h.last_seen,
		       h.task_id
		FROM users u
		LEFT JOIN heartbeats h ON u.id = h.user_id
		WHERE u.type = 'worker' AND u.is_active = 1
		ORDER BY u.username
	`

	rows, err := s.db.Conn().Query(query)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query workers")
		return
	}
	defer rows.Close()

	workers := []WorkerResponse{}
	for rows.Next() {
		var worker WorkerResponse
		var lastSeen sql.NullString
		var taskID sql.NullString

		err := rows.Scan(&worker.ID, &worker.Username, &worker.Status, &lastSeen, &taskID)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan worker")
			return
		}

		if lastSeen.Valid {
			worker.LastSeen = lastSeen.String
		}
		if taskID.Valid {
			worker.CurrentTask = &taskID.String
		}

		workers = append(workers, worker)
	}

	sendJSON(w, http.StatusOK, workers)
}

// handleGetWorker returns a specific worker
func (s *Server) handleGetWorker(w http.ResponseWriter, r *http.Request) {
	workerID := r.PathValue("worker_id")

	query := `
		SELECT u.id, u.username, 
		       COALESCE(h.status, 'offline') as status,
		       h.last_seen,
		       h.task_id
		FROM users u
		LEFT JOIN heartbeats h ON u.id = h.user_id
		WHERE u.id = ? AND u.type = 'worker'
	`

	var worker WorkerResponse
	var lastSeen sql.NullString
	var taskID sql.NullString

	err := s.db.Conn().QueryRow(query, workerID).Scan(
		&worker.ID, &worker.Username, &worker.Status, &lastSeen, &taskID,
	)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "worker not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to get worker")
		return
	}

	if lastSeen.Valid {
		worker.LastSeen = lastSeen.String
	}
	if taskID.Valid {
		worker.CurrentTask = &taskID.String
	}

	sendJSON(w, http.StatusOK, worker)
}

// handleGetWorkerTasks returns all tasks assigned to a worker
func (s *Server) handleGetWorkerTasks(w http.ResponseWriter, r *http.Request) {
	workerID := r.PathValue("worker_id")

	// Verify worker exists
	var exists int
	err := s.db.Conn().QueryRow(`
		SELECT 1 FROM users WHERE id = ? AND type = 'worker'
	`, workerID).Scan(&exists)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "worker not found")
		return
	}

	query := `
		SELECT id, project_id, title, type, description, acceptance_criteria, parent_id, epic_id,
		       depends_on_task_id, status, worker_id, priority, is_complete, completed_at,
		       created_at, updated_at, created_by, updated_by, labels, estimated_effort, actual_effort
		FROM tasks
		WHERE worker_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Conn().Query(query, workerID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query tasks")
		return
	}
	defer rows.Close()

	tasks := []db.Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan task")
			return
		}
		tasks = append(tasks, task)
	}

	sendJSON(w, http.StatusOK, tasks)
}

// handleGetWorkerHistory returns task history for a worker
func (s *Server) handleGetWorkerHistory(w http.ResponseWriter, r *http.Request) {
	workerID := r.PathValue("worker_id")

	// Verify worker exists
	var exists int
	err := s.db.Conn().QueryRow(`
		SELECT 1 FROM users WHERE id = ? AND type = 'worker'
	`, workerID).Scan(&exists)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "worker not found")
		return
	}

	query := `
		SELECT h.id, h.task_id, h.started_at, h.completed_at, h.state, 
		       h.worker_id, h.role_id, h.result, h.summary,
		       t.title as task_title
		FROM task_history h
		LEFT JOIN tasks t ON h.task_id = t.id
		WHERE h.worker_id = ?
		ORDER BY h.started_at DESC
	`

	rows, err := s.db.Conn().Query(query, workerID)
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
		RoleID      string  `json:"role_id"`
		Result      *string `json:"result,omitempty"`
		Summary     *string `json:"summary,omitempty"`
	}

	history := []HistoryEntry{}
	for rows.Next() {
		var entry HistoryEntry
		var completedAt, result, summary, taskTitle sql.NullString

		err := rows.Scan(
			&entry.ID, &entry.TaskID, &entry.StartedAt, &completedAt, &entry.State,
			&entry.WorkerID, &entry.RoleID, &result, &summary, &taskTitle,
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

		history = append(history, entry)
	}

	sendJSON(w, http.StatusOK, history)
}

// WorkerStatsResponse represents worker statistics
type WorkerStatsResponse struct {
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	ActiveTasks    int     `json:"active_tasks"`
	SuccessRate    float64 `json:"success_rate"`
}

// handleGetWorkerStats returns statistics for a worker
func (s *Server) handleGetWorkerStats(w http.ResponseWriter, r *http.Request) {
	workerID := r.PathValue("worker_id")

	// Verify worker exists
	var exists int
	err := s.db.Conn().QueryRow(`
		SELECT 1 FROM users WHERE id = ? AND type = 'worker'
	`, workerID).Scan(&exists)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "worker not found")
		return
	}

	var stats WorkerStatsResponse

	// Get total tasks from history
	s.db.Conn().QueryRow(`
		SELECT COUNT(*) FROM task_history WHERE worker_id = ?
	`, workerID).Scan(&stats.TotalTasks)

	// Get completed tasks
	s.db.Conn().QueryRow(`
		SELECT COUNT(*) FROM task_history 
		WHERE worker_id = ? AND state = 'completed'
	`, workerID).Scan(&stats.CompletedTasks)

	// Get active tasks
	s.db.Conn().QueryRow(`
		SELECT COUNT(*) FROM tasks 
		WHERE worker_id = ? AND is_complete = 0
	`, workerID).Scan(&stats.ActiveTasks)

	// Calculate success rate
	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	sendJSON(w, http.StatusOK, stats)
}

// UpdateWorkerRequest represents a worker update request
type UpdateWorkerRequest struct {
	Status string `json:"status,omitempty"`
}

// handleUpdateWorker updates a worker's status
func (s *Server) handleUpdateWorker(w http.ResponseWriter, r *http.Request) {
	workerID := r.PathValue("worker_id")
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req UpdateWorkerRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Verify worker exists
	var exists int
	err := s.db.Conn().QueryRow(`
		SELECT 1 FROM users WHERE id = ? AND type = 'worker'
	`, workerID).Scan(&exists)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "worker not found")
		return
	}

	// Update heartbeat with new status
	if req.Status != "" {
		err := s.db.UpsertHeartbeat(workerID, req.Status, nil)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to update worker")
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
