package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/simonski/task/internal/db"
)

// TaskRequest represents a task creation/update request
type TaskRequest struct {
	ProjectID          string   `json:"project_id,omitempty"`
	Title              string   `json:"title,omitempty"`
	Type               string   `json:"type,omitempty"`
	Description        string   `json:"description,omitempty"`
	AcceptanceCriteria *string  `json:"acceptance_criteria,omitempty"`
	ParentID           *string  `json:"parent_id,omitempty"`
	EpicID             *string  `json:"epic_id,omitempty"`
	DependsOnTaskID    *string  `json:"depends_on_task_id,omitempty"`
	Priority           string   `json:"priority,omitempty"`
	Labels             []string `json:"labels,omitempty"`
	EstimatedEffort    *int     `json:"estimated_effort,omitempty"`
}

// AssignTaskRequest represents a task assignment request
type AssignTaskRequest struct {
	UserID string `json:"user_id"`
}

// handleListTasks returns tasks with optional filters
func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, project_id, title, type, description, acceptance_criteria, parent_id, epic_id,
		       depends_on_task_id, status, worker_id, priority, is_complete, completed_at,
		       created_at, updated_at, created_by, updated_by, labels, estimated_effort, actual_effort
		FROM tasks
		WHERE 1=1
	`
	args := []interface{}{}

	if projectID := r.URL.Query().Get("project_id"); projectID != "" {
		query += " AND project_id = ?"
		args = append(args, projectID)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if taskType := r.URL.Query().Get("type"); taskType != "" {
		query += " AND type = ?"
		args = append(args, taskType)
	}
	if workerID := r.URL.Query().Get("worker_id"); workerID != "" {
		query += " AND worker_id = ?"
		args = append(args, workerID)
	}
	if isComplete := r.URL.Query().Get("is_complete"); isComplete != "" {
		if isComplete == "true" {
			query += " AND is_complete = 1"
		} else {
			query += " AND is_complete = 0"
		}
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		query += " AND priority = ?"
		args = append(args, priority)
	}

	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := s.db.Conn().Query(query, args...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query tasks")
		return
	}
	defer rows.Close()

	var tasks []db.Task
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

// handleCreateTask creates a new task
func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req TaskRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ProjectID == "" || req.Title == "" || req.Type == "" || req.Description == "" {
		sendError(w, http.StatusBadRequest, "project_id, title, type, and description are required")
		return
	}

	if req.Priority == "" {
		req.Priority = "medium"
	}

	labelsJSON, _ := json.Marshal(req.Labels)

	taskID := uuid.New().String()
	_, err := s.db.Conn().Exec(`
		INSERT INTO tasks (id, project_id, title, type, description, acceptance_criteria, parent_id, epic_id,
		                   depends_on_task_id, priority, created_by, updated_by, labels, estimated_effort)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, taskID, req.ProjectID, req.Title, req.Type, req.Description, req.AcceptanceCriteria, req.ParentID, req.EpicID,
		req.DependsOnTaskID, req.Priority, user.ID, user.ID, string(labelsJSON), req.EstimatedEffort)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	task, err := s.getTaskByID(taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve task")
		return
	}

	sendJSON(w, http.StatusCreated, task)
}

// handleGetTask returns a specific task
func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")

	task, err := s.getTaskByID(taskID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query task")
		return
	}

	sendJSON(w, http.StatusOK, task)
}

// handleUpdateTask updates a task
func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	taskID := r.PathValue("task_id")

	var req TaskRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	labelsJSON, _ := json.Marshal(req.Labels)

	_, err := s.db.Conn().Exec(`
		UPDATE tasks
		SET title = COALESCE(?, title),
		    description = COALESCE(?, description),
		    acceptance_criteria = COALESCE(?, acceptance_criteria),
		    depends_on_task_id = COALESCE(?, depends_on_task_id),
		    priority = COALESCE(?, priority),
		    labels = COALESCE(?, labels),
		    estimated_effort = COALESCE(?, estimated_effort),
		    updated_by = ?,
		    updated_at = ?
		WHERE id = ?
	`, req.Title, req.Description, req.AcceptanceCriteria, req.DependsOnTaskID, req.Priority,
		string(labelsJSON), req.EstimatedEffort, user.ID, time.Now(), taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to update task")
		return
	}

	task, err := s.getTaskByID(taskID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve task")
		return
	}

	sendJSON(w, http.StatusOK, task)
}

// handleDeleteTask deletes a task
func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")

	result, err := s.db.Conn().Exec("DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleClaimTask claims a task for the authenticated user
func (s *Server) handleClaimTask(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	taskID := r.PathValue("task_id")

	// Start transaction
	tx, err := s.db.Conn().Begin()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Check task is idle
	var status string
	err = tx.QueryRow("SELECT status FROM tasks WHERE id = ?", taskID).Scan(&status)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query task")
		return
	}

	if status != "idle" {
		sendError(w, http.StatusBadRequest, "task is not available")
		return
	}

	// Update task
	_, err = tx.Exec(`
		UPDATE tasks
		SET status = 'active', worker_id = ?, updated_by = ?, updated_at = ?
		WHERE id = ?
	`, user.ID, user.ID, time.Now(), taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to claim task")
		return
	}

	// Create history entry
	historyID := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO task_history (id, task_id, started_at, state, worker_id, role_id)
		VALUES (?, ?, ?, 'in-progress', ?, (SELECT id FROM roles WHERE name = 'programmer' AND scope = 'system' LIMIT 1))
	`, historyID, taskID, time.Now(), user.ID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create history")
		return
	}

	if err := tx.Commit(); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to commit transaction")
		return
	}

	task, _ := s.getTaskByID(taskID)
	sendJSON(w, http.StatusOK, task)
}

// handleAssignTask assigns a task to a specific user
func (s *Server) handleAssignTask(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	taskID := r.PathValue("task_id")

	var req AssignTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" {
		sendError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	_, err := s.db.Conn().Exec(`
		UPDATE tasks
		SET status = 'active', worker_id = ?, updated_by = ?, updated_at = ?
		WHERE id = ?
	`, req.UserID, user.ID, time.Now(), taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to assign task")
		return
	}

	task, err := s.getTaskByID(taskID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve task")
		return
	}

	sendJSON(w, http.StatusOK, task)
}

// handleFreeTask releases a task
func (s *Server) handleFreeTask(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	taskID := r.PathValue("task_id")

	_, err := s.db.Conn().Exec(`
		UPDATE tasks
		SET status = 'idle', worker_id = NULL, updated_by = ?, updated_at = ?
		WHERE id = ?
	`, user.ID, time.Now(), taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to free task")
		return
	}

	task, err := s.getTaskByID(taskID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve task")
		return
	}

	sendJSON(w, http.StatusOK, task)
}

// handleCompleteTask marks a task as complete
func (s *Server) handleCompleteTask(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	taskID := r.PathValue("task_id")

	_, err := s.db.Conn().Exec(`
		UPDATE tasks
		SET status = 'idle', worker_id = NULL, is_complete = 1, completed_at = ?, updated_by = ?, updated_at = ?
		WHERE id = ?
	`, time.Now(), user.ID, time.Now(), taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to complete task")
		return
	}

	task, err := s.getTaskByID(taskID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve task")
		return
	}

	sendJSON(w, http.StatusOK, task)
}

// handleGetTaskHistory returns task work history
func (s *Server) handleGetTaskHistory(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")

	rows, err := s.db.Conn().Query(`
		SELECT id, task_id, started_at, completed_at, state, worker_id, role_id, notes, created_at
		FROM task_history
		WHERE task_id = ?
		ORDER BY started_at DESC
	`, taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query history")
		return
	}
	defer rows.Close()

	var history []db.TaskHistory
	for rows.Next() {
		var h db.TaskHistory
		var completedAt sql.NullTime
		var notes sql.NullString
		err := rows.Scan(&h.ID, &h.TaskID, &h.StartedAt, &completedAt, &h.State, &h.WorkerID, &h.RoleID, &notes, &h.CreatedAt)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan history")
			return
		}
		if completedAt.Valid {
			h.CompletedAt = &completedAt.Time
		}
		if notes.Valid {
			h.Notes = &notes.String
		}
		history = append(history, h)
	}

	sendJSON(w, http.StatusOK, history)
}

// getTaskByID retrieves a task by ID
func (s *Server) getTaskByID(taskID string) (db.Task, error) {
	row := s.db.Conn().QueryRow(`
		SELECT id, project_id, title, type, description, acceptance_criteria, parent_id, epic_id,
		       depends_on_task_id, status, worker_id, priority, is_complete, completed_at,
		       created_at, updated_at, created_by, updated_by, labels, estimated_effort, actual_effort
		FROM tasks
		WHERE id = ?
	`, taskID)
	return scanTask(row)
}

// scanTask scans a task row
func scanTask(scanner interface {
	Scan(dest ...interface{}) error
}) (db.Task, error) {
	var task db.Task
	var acceptanceCriteria, parentID, epicID, dependsOnTaskID, workerID sql.NullString
	var completedAt sql.NullTime
	var labelsJSON sql.NullString
	var estimatedEffort, actualEffort sql.NullInt64

	err := scanner.Scan(
		&task.ID,
		&task.ProjectID,
		&task.Title,
		&task.Type,
		&task.Description,
		&acceptanceCriteria,
		&parentID,
		&epicID,
		&dependsOnTaskID,
		&task.Status,
		&workerID,
		&task.Priority,
		&task.IsComplete,
		&completedAt,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.CreatedBy,
		&task.UpdatedBy,
		&labelsJSON,
		&estimatedEffort,
		&actualEffort,
	)
	if err != nil {
		return task, err
	}

	if acceptanceCriteria.Valid {
		task.AcceptanceCriteria = &acceptanceCriteria.String
	}
	if parentID.Valid {
		task.ParentID = &parentID.String
	}
	if epicID.Valid {
		task.EpicID = &epicID.String
	}
	if dependsOnTaskID.Valid {
		task.DependsOnTaskID = &dependsOnTaskID.String
	}
	if workerID.Valid {
		task.WorkerID = &workerID.String
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	if labelsJSON.Valid && labelsJSON.String != "" {
		json.Unmarshal([]byte(labelsJSON.String), &task.Labels)
	}
	if estimatedEffort.Valid {
		effort := int(estimatedEffort.Int64)
		task.EstimatedEffort = &effort
	}
	if actualEffort.Valid {
		effort := int(actualEffort.Int64)
		task.ActualEffort = &effort
	}

	return task, nil
}
