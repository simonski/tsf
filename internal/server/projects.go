package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/simonski/task/internal/db"
)

// ProjectRequest represents a project creation/update request
type ProjectRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Repository  *string `json:"repository,omitempty"`
	Status      string  `json:"status,omitempty"`
	Visibility  string  `json:"visibility,omitempty"`
}

// handleListProjects returns all accessible projects
func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	query := `
		SELECT id, name, description, repository, status, visibility, created_at, updated_at, created_by, updated_by
		FROM projects
		WHERE 1=1
	`
	args := []interface{}{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY name"

	rows, err := s.db.Conn().Query(query, args...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query projects")
		return
	}
	defer rows.Close()

	var projects []db.Project
	for rows.Next() {
		var project db.Project
		var repository sql.NullString
		var createdAt, updatedAt string
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&repository,
			&project.Status,
			&project.Visibility,
			&createdAt,
			&updatedAt,
			&project.CreatedBy,
			&project.UpdatedBy,
		)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan project")
			return
		}
		project.CreatedAt = parseTimestamp(createdAt)
		project.UpdatedAt = parseTimestamp(updatedAt)
		if repository.Valid {
			project.Repository = &repository.String
		}
		projects = append(projects, project)
	}

	sendJSON(w, http.StatusOK, projects)
}

// handleCreateProject creates a new project
func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req ProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Description == "" {
		sendError(w, http.StatusBadRequest, "name and description are required")
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}
	if req.Visibility == "" {
		req.Visibility = "public"
	}

	projectID := uuid.New().String()
	_, err := s.db.Conn().Exec(`
		INSERT INTO projects (id, name, description, repository, status, visibility, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, projectID, req.Name, req.Description, req.Repository, req.Status, req.Visibility, user.ID, user.ID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create project")
		return
	}

	var project db.Project
	var repository sql.NullString
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT id, name, description, repository, status, visibility, created_at, updated_at, created_by, updated_by
		FROM projects
		WHERE id = ?
	`, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&repository,
		&project.Status,
		&project.Visibility,
		&createdAt,
		&updatedAt,
		&project.CreatedBy,
		&project.UpdatedBy,
	)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve project")
		return
	}
	project.CreatedAt = parseTimestamp(createdAt)
	project.UpdatedAt = parseTimestamp(updatedAt)
	if repository.Valid {
		project.Repository = &repository.String
	}

	sendJSON(w, http.StatusCreated, project)
}

// handleGetProject returns a specific project
func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")

	var project db.Project
	var repository sql.NullString
	var createdAt, updatedAt string
	err := s.db.Conn().QueryRow(`
		SELECT id, name, description, repository, status, visibility, created_at, updated_at, created_by, updated_by
		FROM projects
		WHERE id = ?
	`, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&repository,
		&project.Status,
		&project.Visibility,
		&createdAt,
		&updatedAt,
		&project.CreatedBy,
		&project.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	}
	project.CreatedAt = parseTimestamp(createdAt)
	project.UpdatedAt = parseTimestamp(updatedAt)
	if repository.Valid {
		project.Repository = &repository.String
	}

	sendJSON(w, http.StatusOK, project)
}

// handleUpdateProject updates a project
func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	projectID := r.PathValue("project_id")

	var req ProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err := s.db.Conn().Exec(`
		UPDATE projects
		SET name = COALESCE(NULLIF(?, ''), name),
		    description = COALESCE(NULLIF(?, ''), description),
		    repository = COALESCE(?, repository),
		    status = COALESCE(NULLIF(?, ''), status),
		    visibility = COALESCE(NULLIF(?, ''), visibility),
		    updated_by = ?,
		    updated_at = ?
		WHERE id = ?
	`, req.Name, req.Description, req.Repository, req.Status, req.Visibility, user.ID, time.Now(), projectID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to update project")
		return
	}

	var project db.Project
	var repository sql.NullString
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT id, name, description, repository, status, visibility, created_at, updated_at, created_by, updated_by
		FROM projects
		WHERE id = ?
	`, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&repository,
		&project.Status,
		&project.Visibility,
		&createdAt,
		&updatedAt,
		&project.CreatedBy,
		&project.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve project")
		return
	}
	project.CreatedAt = parseTimestamp(createdAt)
	project.UpdatedAt = parseTimestamp(updatedAt)
	if repository.Valid {
		project.Repository = &repository.String
	}

	sendJSON(w, http.StatusOK, project)
}

// handleDeleteProject deletes a project
func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")

	// Check if project has tasks
	var count int
	err := s.db.Conn().QueryRow("SELECT COUNT(*) FROM tasks WHERE project_id = ?", projectID).Scan(&count)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to check tasks")
		return
	}
	if count > 0 {
		sendError(w, http.StatusBadRequest, "cannot delete project with tasks")
		return
	}

	result, err := s.db.Conn().Exec("DELETE FROM projects WHERE id = ?", projectID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to delete project")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
