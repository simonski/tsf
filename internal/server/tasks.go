package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
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
	UserID string  `json:"user_id"`
	RoleID *string `json:"role_id,omitempty"`
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

	// Initialize with empty slice to ensure JSON returns [] not null
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

	if req.Title == "" || req.Type == "" {
		sendError(w, http.StatusBadRequest, "title and type are required")
		return
	}

	projectID, err := s.resolveTaskProjectID(req.ProjectID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusBadRequest, "project_id is required and no default project is configured")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to resolve project_id")
		return
	}

	var projectExists int
	err = s.db.Conn().QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&projectExists)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	}
	if projectExists == 0 {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	if req.Priority == "" {
		req.Priority = "medium"
	}

	labelsJSON, _ := json.Marshal(req.Labels)

	taskID := uuid.New().String()
	_, err = s.db.Conn().Exec(`
		INSERT INTO tasks (id, project_id, title, type, description, acceptance_criteria, parent_id, epic_id,
		                   depends_on_task_id, priority, created_by, updated_by, labels, estimated_effort)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, taskID, projectID, req.Title, req.Type, req.Description, req.AcceptanceCriteria, req.ParentID, req.EpicID,
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

// resolveTaskProjectID resolves project aliases for task creation.
// Supports literal "default" by resolving config.default_project_id first,
// then falling back to project name "default".
func (s *Server) resolveTaskProjectID(requested string) (string, error) {
	projectID := strings.TrimSpace(requested)
	if projectID != "" && projectID != "default" {
		return projectID, nil
	}

	var cfgValue string
	err := s.db.Conn().QueryRow("SELECT value FROM config WHERE key = 'default_project_id'").Scan(&cfgValue)
	if err == nil && strings.TrimSpace(cfgValue) != "" {
		return strings.TrimSpace(cfgValue), nil
	}
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	var defaultProjectID string
	err = s.db.Conn().QueryRow("SELECT id FROM projects WHERE name = 'default' LIMIT 1").Scan(&defaultProjectID)
	if err == nil {
		return defaultProjectID, nil
	}
	if err != nil {
		return "", err
	}
	return "", sql.ErrNoRows
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

	// Use json.RawMessage to detect explicit null for depends_on_task_id
	var rawBody map[string]json.RawMessage
	if err := decodeJSON(r, &rawBody); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}

	if v, ok := rawBody["title"]; ok {
		var val string
		if json.Unmarshal(v, &val) == nil && val != "" {
			updates = append(updates, "title = ?")
			args = append(args, val)
		}
	}
	if v, ok := rawBody["description"]; ok {
		var val string
		if json.Unmarshal(v, &val) == nil {
			updates = append(updates, "description = ?")
			args = append(args, val)
		}
	}
	if v, ok := rawBody["acceptance_criteria"]; ok {
		var val *string
		if json.Unmarshal(v, &val) == nil {
			updates = append(updates, "acceptance_criteria = ?")
			args = append(args, val)
		}
	}
	if v, ok := rawBody["depends_on_task_id"]; ok {
		// Check if explicitly null or has a value
		if string(v) == "null" {
			updates = append(updates, "depends_on_task_id = NULL")
		} else {
			var val string
			if json.Unmarshal(v, &val) == nil && val != "" {
				updates = append(updates, "depends_on_task_id = ?")
				args = append(args, val)
			}
		}
	}
	// Support blocked_by as alias for depends_on_task_id
	if v, ok := rawBody["blocked_by"]; ok {
		if string(v) == "null" {
			updates = append(updates, "depends_on_task_id = NULL")
		} else {
			var val string
			if json.Unmarshal(v, &val) == nil && val != "" {
				updates = append(updates, "depends_on_task_id = ?")
				args = append(args, val)
			}
		}
	}
	if v, ok := rawBody["priority"]; ok {
		var val string
		if json.Unmarshal(v, &val) == nil && val != "" {
			updates = append(updates, "priority = ?")
			args = append(args, val)
		}
	}
	if v, ok := rawBody["labels"]; ok {
		var val []string
		if json.Unmarshal(v, &val) == nil {
			labelsJSON, _ := json.Marshal(val)
			updates = append(updates, "labels = ?")
			args = append(args, string(labelsJSON))
		}
	}
	if v, ok := rawBody["estimated_effort"]; ok {
		var val *int
		if json.Unmarshal(v, &val) == nil {
			updates = append(updates, "estimated_effort = ?")
			args = append(args, val)
		}
	}

	if len(updates) == 0 {
		sendError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	// Add updated_by and updated_at
	updates = append(updates, "updated_by = ?", "updated_at = ?")
	args = append(args, user.ID, time.Now())
	args = append(args, taskID)

	query := "UPDATE tasks SET " + joinWithComma(updates) + " WHERE id = ?"
	_, err := s.db.Conn().Exec(query, args...)
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

func joinWithComma(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

// handleDeleteTask soft deletes a task by setting is_deleted flag
func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	taskID := r.PathValue("task_id")

	result, err := s.db.Conn().Exec(`
		UPDATE tasks SET is_deleted = 1, updated_by = ?, updated_at = ? WHERE id = ?
	`, user.ID, time.Now(), taskID)
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

	// Check task exists and is idle, also get dependency info
	var status string
	var dependsOnTaskID sql.NullString
	err = tx.QueryRow("SELECT status, depends_on_task_id FROM tasks WHERE id = ?", taskID).Scan(&status, &dependsOnTaskID)
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

	// Check if blocked by incomplete dependency
	if dependsOnTaskID.Valid {
		var depComplete bool
		err = tx.QueryRow("SELECT is_complete FROM tasks WHERE id = ?", dependsOnTaskID.String).Scan(&depComplete)
		if err == nil && !depComplete {
			sendError(w, http.StatusConflict, "task is blocked by incomplete dependency")
			return
		}
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

	// Create history entry with optional role
	if req.RoleID != nil && *req.RoleID != "" {
		historyID := uuid.New().String()
		s.db.Conn().Exec(`
			INSERT INTO task_history (id, task_id, started_at, state, worker_id, role_id)
			VALUES (?, ?, ?, 'in-progress', ?, ?)
		`, historyID, taskID, time.Now(), req.UserID, *req.RoleID)
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
	now := time.Now()

	// Start transaction
	tx, err := s.db.Conn().Begin()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Update task
	_, err = tx.Exec(`
		UPDATE tasks
		SET status = 'idle', worker_id = NULL, updated_by = ?, updated_at = ?
		WHERE id = ?
	`, user.ID, now, taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to free task")
		return
	}

	// Update the most recent in-progress history entry to abandoned
	tx.Exec(`
		UPDATE task_history
		SET state = 'abandoned', completed_at = ?
		WHERE id = (
			SELECT id FROM task_history 
			WHERE task_id = ? AND state = 'in-progress'
			ORDER BY started_at DESC
			LIMIT 1
		)
	`, now, taskID)
	// Ignore error - may not have a history entry

	if err := tx.Commit(); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to commit transaction")
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

	// Parse optional notes from body
	var body struct {
		Notes string `json:"notes"`
	}
	decodeJSON(r, &body)

	now := time.Now()

	// Start transaction
	tx, err := s.db.Conn().Begin()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Update task
	_, err = tx.Exec(`
		UPDATE tasks
		SET status = 'idle', worker_id = NULL, is_complete = 1, completed_at = ?, updated_by = ?, updated_at = ?
		WHERE id = ?
	`, now, user.ID, now, taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to complete task")
		return
	}

	// Update the most recent in-progress history entry to success
	var notes interface{} = nil
	if body.Notes != "" {
		notes = body.Notes
	}
	_, err = tx.Exec(`
		UPDATE task_history
		SET state = 'success', completed_at = ?, notes = COALESCE(?, notes)
		WHERE id = (
			SELECT id FROM task_history 
			WHERE task_id = ? AND state = 'in-progress'
			ORDER BY started_at DESC
			LIMIT 1
		)
	`, now, notes, taskID)
	// Ignore error - may not have a history entry

	if err := tx.Commit(); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to commit transaction")
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
		var completedAt sql.NullString
		var notes sql.NullString
		var startedAt, createdAt string
		err := rows.Scan(&h.ID, &h.TaskID, &startedAt, &completedAt, &h.State, &h.WorkerID, &h.RoleID, &notes, &createdAt)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan history")
			return
		}
		h.StartedAt = parseTimestamp(startedAt)
		h.CreatedAt = parseTimestamp(createdAt)
		if completedAt.Valid {
			t := parseTimestamp(completedAt.String)
			h.CompletedAt = &t
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
	var completedAt sql.NullString
	var labelsJSON sql.NullString
	var estimatedEffort, actualEffort sql.NullInt64
	var createdAt, updatedAt string

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
		&createdAt,
		&updatedAt,
		&task.CreatedBy,
		&task.UpdatedBy,
		&labelsJSON,
		&estimatedEffort,
		&actualEffort,
	)
	if err != nil {
		return task, err
	}

	task.CreatedAt = parseTimestamp(createdAt)
	task.UpdatedAt = parseTimestamp(updatedAt)
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
		t := parseTimestamp(completedAt.String)
		task.CompletedAt = &t
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

// handleGetTaskDependencies returns the dependency information for a task
func (s *Server) handleGetTaskDependencies(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")

	// Get the task's dependency (what it's blocked by)
	var dependsOnTaskID sql.NullString
	err := s.db.Conn().QueryRow("SELECT depends_on_task_id FROM tasks WHERE id = ?", taskID).Scan(&dependsOnTaskID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query task")
		return
	}

	result := map[string]interface{}{
		"task_id": taskID,
	}

	// If this task is blocked by another, include that info
	if dependsOnTaskID.Valid {
		blockedByTask, err := s.getTaskByID(dependsOnTaskID.String)
		if err == nil {
			result["blocked_by"] = map[string]interface{}{
				"id":          blockedByTask.ID,
				"title":       blockedByTask.Title,
				"status":      blockedByTask.Status,
				"is_complete": blockedByTask.IsComplete,
			}
		}
	}

	// Find tasks that are blocked by this task
	rows, err := s.db.Conn().Query(`
		SELECT id, title, status, is_complete
		FROM tasks
		WHERE depends_on_task_id = ?
	`, taskID)
	if err == nil {
		defer rows.Close()
		var blocking []map[string]interface{}
		for rows.Next() {
			var id, title, status string
			var isComplete bool
			if err := rows.Scan(&id, &title, &status, &isComplete); err == nil {
				blocking = append(blocking, map[string]interface{}{
					"id":          id,
					"title":       title,
					"status":      status,
					"is_complete": isComplete,
				})
			}
		}
		if len(blocking) > 0 {
			result["blocking"] = blocking
		}
	}

	sendJSON(w, http.StatusOK, result)
}

// handleUnassignTask unassigns a task from its current worker
func (s *Server) handleUnassignTask(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	taskID := r.PathValue("task_id")

	// Check task exists
	var workerID sql.NullString
	err := s.db.Conn().QueryRow("SELECT worker_id FROM tasks WHERE id = ?", taskID).Scan(&workerID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query task")
		return
	}

	if !workerID.Valid {
		sendError(w, http.StatusBadRequest, "task is not assigned")
		return
	}

	_, err = s.db.Conn().Exec(`
		UPDATE tasks
		SET status = 'idle', worker_id = NULL, updated_by = ?, updated_at = ?
		WHERE id = ?
	`, user.ID, time.Now(), taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to unassign task")
		return
	}

	task, _ := s.getTaskByID(taskID)
	sendJSON(w, http.StatusOK, task)
}

// CommentRequest represents a comment request
type CommentRequest struct {
	Text   string `json:"text"`
	Author string `json:"author,omitempty"`
}

// handleGetTaskComments returns comments for a task
func (s *Server) handleGetTaskComments(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")

	// Check task exists and get comments
	var commentsJSON sql.NullString
	err := s.db.Conn().QueryRow("SELECT comments FROM tasks WHERE id = ?", taskID).Scan(&commentsJSON)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query task")
		return
	}

	var comments []db.Comment
	if commentsJSON.Valid && commentsJSON.String != "" {
		if err := json.Unmarshal([]byte(commentsJSON.String), &comments); err != nil {
			// If invalid JSON, return empty array
			comments = []db.Comment{}
		}
	}

	sendJSON(w, http.StatusOK, comments)
}

// handleAddTaskComment adds a comment to a task
func (s *Server) handleAddTaskComment(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	taskID := r.PathValue("task_id")

	var req CommentRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Text == "" {
		sendError(w, http.StatusBadRequest, "text is required")
		return
	}

	// Get existing comments
	var commentsJSON sql.NullString
	err := s.db.Conn().QueryRow("SELECT comments FROM tasks WHERE id = ?", taskID).Scan(&commentsJSON)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query task")
		return
	}

	var comments []db.Comment
	if commentsJSON.Valid && commentsJSON.String != "" {
		json.Unmarshal([]byte(commentsJSON.String), &comments)
	}

	// Add new comment
	author := req.Author
	if author == "" {
		author = user.Username
	}
	newComment := db.Comment{
		Text:      req.Text,
		Author:    author,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	comments = append(comments, newComment)

	// Update task with new comments
	commentsData, _ := json.Marshal(comments)
	_, err = s.db.Conn().Exec(`
		UPDATE tasks SET comments = ?, updated_by = ?, updated_at = ? WHERE id = ?
	`, string(commentsData), user.ID, time.Now(), taskID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to add comment")
		return
	}

	sendJSON(w, http.StatusCreated, newComment)
}

// TaskRequestRequest represents a task request by a worker
type TaskRequestRequest struct {
	ProjectID *string `json:"project_id,omitempty"`
	TaskType  *string `json:"task_type,omitempty"`
}

// handleTaskRequest allows workers to request a task
func (s *Server) handleTaskRequest(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if user.Type != "worker" {
		sendError(w, http.StatusForbidden, "only workers can request tasks")
		return
	}

	var req TaskRequestRequest
	if err := decodeJSON(r, &req); err != nil {
		// Body is optional
		req = TaskRequestRequest{}
	}

	// Build query for available task
	query := `
		SELECT id, project_id, title, type, description, acceptance_criteria, parent_id, epic_id,
		       depends_on_task_id, status, worker_id, priority, is_complete, completed_at,
		       created_at, updated_at, created_by, updated_by, labels, estimated_effort, actual_effort
		FROM tasks
		WHERE is_complete = 0 
		  AND is_deleted = 0
		  AND status = 'idle'
		  AND (depends_on_task_id IS NULL 
		       OR depends_on_task_id IN (SELECT id FROM tasks WHERE is_complete = 1))
	`
	args := []interface{}{}

	if req.ProjectID != nil && *req.ProjectID != "" {
		query += " AND project_id = ?"
		args = append(args, *req.ProjectID)
	}
	if req.TaskType != nil && *req.TaskType != "" {
		query += " AND type = ?"
		args = append(args, *req.TaskType)
	}

	query += `
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

	row := s.db.Conn().QueryRow(query, args...)
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
	historyID := uuid.New().String()
	_, err = s.db.Conn().Exec(`
		INSERT INTO task_history (id, task_id, started_at, state, worker_id)
		VALUES (?, ?, ?, 'in-progress', ?)
	`, historyID, task.ID, time.Now(), user.ID)

	task, _ = s.getTaskByID(task.ID)
	sendJSON(w, http.StatusOK, task)
}

// TaskReturnRequest represents a task return request
type TaskReturnRequest struct {
	State   string  `json:"state"` // "completed", "failed", "blocked"
	Summary *string `json:"summary,omitempty"`
	Result  *string `json:"result,omitempty"`
}

// handleTaskReturn allows workers to return a completed task
func (s *Server) handleTaskReturn(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	taskID := r.PathValue("task_id")

	var req TaskReturnRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.State == "" {
		sendError(w, http.StatusBadRequest, "state is required")
		return
	}

	// Verify task is assigned to this worker
	var workerID sql.NullString
	err := s.db.Conn().QueryRow("SELECT worker_id FROM tasks WHERE id = ?", taskID).Scan(&workerID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query task")
		return
	}

	if !workerID.Valid || workerID.String != user.ID {
		sendError(w, http.StatusForbidden, "task not assigned to you")
		return
	}

	// Update task based on state
	if req.State == "completed" {
		_, err = s.db.Conn().Exec(`
			UPDATE tasks
			SET status = 'idle', is_complete = 1, completed_at = ?, updated_by = ?, updated_at = ?
			WHERE id = ?
		`, time.Now(), user.ID, time.Now(), taskID)
	} else {
		_, err = s.db.Conn().Exec(`
			UPDATE tasks
			SET status = 'idle', worker_id = NULL, updated_by = ?, updated_at = ?
			WHERE id = ?
		`, user.ID, time.Now(), taskID)
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to update task")
		return
	}

	// Update task history
	_, err = s.db.Conn().Exec(`
		UPDATE task_history
		SET completed_at = ?, state = ?, summary = ?, result = ?
		WHERE task_id = ? AND worker_id = ? AND completed_at IS NULL
	`, time.Now(), req.State, req.Summary, req.Result, taskID, user.ID)

	task, _ := s.getTaskByID(taskID)
	sendJSON(w, http.StatusOK, task)
}

// handleStatus returns task counts grouped by type and status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			type,
			status,
			COUNT(*) as count
		FROM tasks
		WHERE is_deleted = 0
		GROUP BY type, status
		ORDER BY type, status
	`

	rows, err := s.db.Conn().Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Build nested structure: type -> status -> count
	statusData := make(map[string]map[string]int)
	totals := make(map[string]int)
	grandTotal := 0

	for rows.Next() {
		var taskType, status string
		var count int
		if err := rows.Scan(&taskType, &status, &count); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if statusData[taskType] == nil {
			statusData[taskType] = make(map[string]int)
		}
		statusData[taskType][status] = count
		totals[taskType] += count
		grandTotal += count
	}

	// Build response structure
	response := map[string]interface{}{
		"by_type_and_status": statusData,
		"totals_by_type":     totals,
		"grand_total":        grandTotal,
	}

	sendJSON(w, http.StatusOK, response)
}
