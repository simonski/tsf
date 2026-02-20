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

// ProjectFileRequest represents a project file creation/update request
type ProjectFileRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ProjectFileUpdateRequest represents mutable fields for project file updates
type ProjectFileUpdateRequest struct {
	Name    *string `json:"name,omitempty"`
	Content *string `json:"content,omitempty"`
}

// ProjectNoteRequest represents a project note creation/update request
type ProjectNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ProjectNoteUpdateRequest represents mutable fields for project note updates
type ProjectNoteUpdateRequest struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
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

func (s *Server) projectExists(projectID string) (bool, error) {
	var exists int
	err := s.db.Conn().QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", projectID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *Server) handleListProjectFiles(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	if ok, err := s.projectExists(projectID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	} else if !ok {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	rows, err := s.db.Conn().Query(`
		SELECT id, project_id, name, content, created_at, updated_at, created_by, updated_by
		FROM project_files
		WHERE project_id = ?
		ORDER BY name
	`, projectID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project files")
		return
	}
	defer rows.Close()

	files := []db.ProjectFile{}
	for rows.Next() {
		var item db.ProjectFile
		var createdAt, updatedAt string
		if err := rows.Scan(
			&item.ID,
			&item.ProjectID,
			&item.Name,
			&item.Content,
			&createdAt,
			&updatedAt,
			&item.CreatedBy,
			&item.UpdatedBy,
		); err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan project file")
			return
		}
		item.CreatedAt = parseTimestamp(createdAt)
		item.UpdatedAt = parseTimestamp(updatedAt)
		files = append(files, item)
	}

	sendJSON(w, http.StatusOK, files)
}

func (s *Server) handleCreateProjectFile(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	projectID := r.PathValue("project_id")
	if ok, err := s.projectExists(projectID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	} else if !ok {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	var req ProjectFileRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		sendError(w, http.StatusBadRequest, "name is required")
		return
	}

	fileID := uuid.New().String()
	_, err := s.db.Conn().Exec(`
		INSERT INTO project_files (id, project_id, name, content, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, fileID, projectID, req.Name, req.Content, user.ID, user.ID)
	if err != nil {
		sendError(w, http.StatusBadRequest, "failed to create project file")
		return
	}

	var item db.ProjectFile
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT id, project_id, name, content, created_at, updated_at, created_by, updated_by
		FROM project_files
		WHERE project_id = ? AND id = ?
	`, projectID, fileID).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Name,
		&item.Content,
		&createdAt,
		&updatedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
	)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve project file")
		return
	}
	item.CreatedAt = parseTimestamp(createdAt)
	item.UpdatedAt = parseTimestamp(updatedAt)

	sendJSON(w, http.StatusCreated, item)
}

func (s *Server) handleGetProjectFile(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	fileID := r.PathValue("file_id")
	if ok, err := s.projectExists(projectID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	} else if !ok {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	var item db.ProjectFile
	var createdAt, updatedAt string
	err := s.db.Conn().QueryRow(`
		SELECT id, project_id, name, content, created_at, updated_at, created_by, updated_by
		FROM project_files
		WHERE project_id = ? AND id = ?
	`, projectID, fileID).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Name,
		&item.Content,
		&createdAt,
		&updatedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "project file not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project file")
		return
	}
	item.CreatedAt = parseTimestamp(createdAt)
	item.UpdatedAt = parseTimestamp(updatedAt)

	sendJSON(w, http.StatusOK, item)
}

func (s *Server) handleUpdateProjectFile(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	projectID := r.PathValue("project_id")
	fileID := r.PathValue("file_id")
	if ok, err := s.projectExists(projectID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	} else if !ok {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	var req ProjectFileUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == nil && req.Content == nil {
		sendError(w, http.StatusBadRequest, "at least one field to update required")
		return
	}

	result, err := s.db.Conn().Exec(`
		UPDATE project_files
		SET name = CASE
		        WHEN ? IS NULL OR ? = '' THEN name
		        ELSE ?
		    END,
		    content = CASE
		        WHEN ? IS NULL THEN content
		        ELSE ?
		    END,
		    updated_by = ?,
		    updated_at = ?
		WHERE project_id = ? AND id = ?
	`, req.Name, req.Name, req.Name, req.Content, req.Content, user.ID, time.Now(), projectID, fileID)
	if err != nil {
		sendError(w, http.StatusBadRequest, "failed to update project file")
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		sendError(w, http.StatusNotFound, "project file not found")
		return
	}

	s.handleGetProjectFile(w, r)
}

func (s *Server) handleDeleteProjectFile(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	fileID := r.PathValue("file_id")
	if ok, err := s.projectExists(projectID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	} else if !ok {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	result, err := s.db.Conn().Exec(`
		DELETE FROM project_files
		WHERE project_id = ? AND id = ?
	`, projectID, fileID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to delete project file")
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		sendError(w, http.StatusNotFound, "project file not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListProjectNotes(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	if ok, err := s.projectExists(projectID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	} else if !ok {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	rows, err := s.db.Conn().Query(`
		SELECT id, project_id, title, content, created_at, updated_at, created_by, updated_by
		FROM project_notes
		WHERE project_id = ?
		ORDER BY created_at DESC, title
	`, projectID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project notes")
		return
	}
	defer rows.Close()

	notes := []db.ProjectNote{}
	for rows.Next() {
		var item db.ProjectNote
		var createdAt, updatedAt string
		if err := rows.Scan(
			&item.ID,
			&item.ProjectID,
			&item.Title,
			&item.Content,
			&createdAt,
			&updatedAt,
			&item.CreatedBy,
			&item.UpdatedBy,
		); err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan project note")
			return
		}
		item.CreatedAt = parseTimestamp(createdAt)
		item.UpdatedAt = parseTimestamp(updatedAt)
		notes = append(notes, item)
	}

	sendJSON(w, http.StatusOK, notes)
}

func (s *Server) handleCreateProjectNote(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	projectID := r.PathValue("project_id")
	if ok, err := s.projectExists(projectID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	} else if !ok {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	var req ProjectNoteRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		sendError(w, http.StatusBadRequest, "title is required")
		return
	}

	noteID := uuid.New().String()
	_, err := s.db.Conn().Exec(`
		INSERT INTO project_notes (id, project_id, title, content, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, noteID, projectID, req.Title, req.Content, user.ID, user.ID)
	if err != nil {
		sendError(w, http.StatusBadRequest, "failed to create project note")
		return
	}

	var item db.ProjectNote
	var createdAt, updatedAt string
	err = s.db.Conn().QueryRow(`
		SELECT id, project_id, title, content, created_at, updated_at, created_by, updated_by
		FROM project_notes
		WHERE project_id = ? AND id = ?
	`, projectID, noteID).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Title,
		&item.Content,
		&createdAt,
		&updatedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
	)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to retrieve project note")
		return
	}
	item.CreatedAt = parseTimestamp(createdAt)
	item.UpdatedAt = parseTimestamp(updatedAt)

	sendJSON(w, http.StatusCreated, item)
}

func (s *Server) handleGetProjectNote(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	noteID := r.PathValue("note_id")
	if ok, err := s.projectExists(projectID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	} else if !ok {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	var item db.ProjectNote
	var createdAt, updatedAt string
	err := s.db.Conn().QueryRow(`
		SELECT id, project_id, title, content, created_at, updated_at, created_by, updated_by
		FROM project_notes
		WHERE project_id = ? AND id = ?
	`, projectID, noteID).Scan(
		&item.ID,
		&item.ProjectID,
		&item.Title,
		&item.Content,
		&createdAt,
		&updatedAt,
		&item.CreatedBy,
		&item.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "project note not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project note")
		return
	}
	item.CreatedAt = parseTimestamp(createdAt)
	item.UpdatedAt = parseTimestamp(updatedAt)

	sendJSON(w, http.StatusOK, item)
}

func (s *Server) handleUpdateProjectNote(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	projectID := r.PathValue("project_id")
	noteID := r.PathValue("note_id")
	if ok, err := s.projectExists(projectID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	} else if !ok {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	var req ProjectNoteUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == nil && req.Content == nil {
		sendError(w, http.StatusBadRequest, "at least one field to update required")
		return
	}

	result, err := s.db.Conn().Exec(`
		UPDATE project_notes
		SET title = CASE
		        WHEN ? IS NULL OR ? = '' THEN title
		        ELSE ?
		    END,
		    content = CASE
		        WHEN ? IS NULL THEN content
		        ELSE ?
		    END,
		    updated_by = ?,
		    updated_at = ?
		WHERE project_id = ? AND id = ?
	`, req.Title, req.Title, req.Title, req.Content, req.Content, user.ID, time.Now(), projectID, noteID)
	if err != nil {
		sendError(w, http.StatusBadRequest, "failed to update project note")
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		sendError(w, http.StatusNotFound, "project note not found")
		return
	}

	s.handleGetProjectNote(w, r)
}

func (s *Server) handleDeleteProjectNote(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	noteID := r.PathValue("note_id")
	if ok, err := s.projectExists(projectID); err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query project")
		return
	} else if !ok {
		sendError(w, http.StatusNotFound, "project not found")
		return
	}

	result, err := s.db.Conn().Exec(`
		DELETE FROM project_notes
		WHERE project_id = ? AND id = ?
	`, projectID, noteID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to delete project note")
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		sendError(w, http.StatusNotFound, "project note not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
