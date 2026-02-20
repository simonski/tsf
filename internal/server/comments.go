package server

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/simonski/task/internal/db"
)

type EntityCommentCreateRequest struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Text       string `json:"text"`
}

type EntityCommentUpdateRequest struct {
	Text string `json:"text"`
}

func (s *Server) commentEntityExists(entityType, entityID string) (bool, error) {
	switch entityType {
	case "project":
		var count int
		err := s.db.Conn().QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", entityID).Scan(&count)
		if err != nil {
			return false, err
		}
		return count > 0, nil
	case "task":
		var count int
		err := s.db.Conn().QueryRow("SELECT COUNT(*) FROM tasks WHERE id = ? AND is_deleted = 0", entityID).Scan(&count)
		if err != nil {
			return false, err
		}
		return count > 0, nil
	default:
		return false, nil
	}
}

func (s *Server) handleListComments(w http.ResponseWriter, r *http.Request) {
	entityType := strings.TrimSpace(r.URL.Query().Get("entity_type"))
	entityID := strings.TrimSpace(r.URL.Query().Get("entity_id"))
	includeDeleted := strings.EqualFold(r.URL.Query().Get("include_deleted"), "true")

	if entityType == "" || entityID == "" {
		sendError(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}
	if entityType != "project" && entityType != "task" {
		sendError(w, http.StatusBadRequest, "entity_type must be 'project' or 'task'")
		return
	}
	ok, err := s.commentEntityExists(entityType, entityID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query entity")
		return
	}
	if !ok {
		sendError(w, http.StatusNotFound, "entity not found")
		return
	}

	query := `
		SELECT id, entity_type, entity_id, owner_id, owner_username, text, is_deleted, created_at, updated_at, deleted_at
		FROM comments
		WHERE entity_type = ? AND entity_id = ?
	`
	args := []interface{}{entityType, entityID}
	if !includeDeleted {
		query += " AND is_deleted = 0"
	}
	query += " ORDER BY created_at ASC"

	rows, err := s.db.Conn().Query(query, args...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query comments")
		return
	}
	defer rows.Close()

	comments := []db.EntityComment{}
	for rows.Next() {
		item, err := scanEntityComment(rows)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan comment")
			return
		}
		comments = append(comments, item)
	}

	sendJSON(w, http.StatusOK, comments)
}

func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req EntityCommentCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.EntityType = strings.TrimSpace(req.EntityType)
	req.EntityID = strings.TrimSpace(req.EntityID)
	req.Text = strings.TrimSpace(req.Text)

	if req.EntityType == "" || req.EntityID == "" || req.Text == "" {
		sendError(w, http.StatusBadRequest, "entity_type, entity_id, and text are required")
		return
	}
	if req.EntityType != "project" && req.EntityType != "task" {
		sendError(w, http.StatusBadRequest, "entity_type must be 'project' or 'task'")
		return
	}
	ok, err := s.commentEntityExists(req.EntityType, req.EntityID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query entity")
		return
	}
	if !ok {
		sendError(w, http.StatusNotFound, "entity not found")
		return
	}

	commentID := uuid.New().String()
	_, err = s.db.Conn().Exec(`
		INSERT INTO comments (id, entity_type, entity_id, owner_id, owner_username, text, is_deleted)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`, commentID, req.EntityType, req.EntityID, user.ID, user.Username, req.Text)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create comment")
		return
	}

	_, err = s.db.Conn().Exec(`
		INSERT INTO comment_history (id, comment_id, editor_id, editor_username, action, text)
		VALUES (?, ?, ?, ?, 'create', ?)
	`, uuid.New().String(), commentID, user.ID, user.Username, req.Text)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create comment history")
		return
	}

	comment, err := s.getEntityComment(commentID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "comment not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query comment")
		return
	}
	sendJSON(w, http.StatusCreated, comment)
}

func (s *Server) handleGetComment(w http.ResponseWriter, r *http.Request) {
	commentID := r.PathValue("comment_id")
	comment, err := s.getEntityComment(commentID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "comment not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query comment")
		return
	}
	sendJSON(w, http.StatusOK, comment)
}

func (s *Server) handleUpdateComment(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	commentID := r.PathValue("comment_id")
	comment, err := s.getEntityComment(commentID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "comment not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query comment")
		return
	}
	if comment.OwnerID != user.ID {
		sendError(w, http.StatusForbidden, "only the comment owner can edit")
		return
	}
	if comment.IsDeleted {
		sendError(w, http.StatusBadRequest, "cannot edit a deleted comment")
		return
	}

	var req EntityCommentUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		sendError(w, http.StatusBadRequest, "text is required")
		return
	}

	_, err = s.db.Conn().Exec(`
		UPDATE comments
		SET text = ?, updated_at = ?
		WHERE id = ?
	`, req.Text, time.Now(), commentID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to update comment")
		return
	}

	_, err = s.db.Conn().Exec(`
		INSERT INTO comment_history (id, comment_id, editor_id, editor_username, action, text)
		VALUES (?, ?, ?, ?, 'edit', ?)
	`, uuid.New().String(), commentID, user.ID, user.Username, req.Text)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create comment history")
		return
	}

	updated, err := s.getEntityComment(commentID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query comment")
		return
	}
	sendJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDeleteComment(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		sendError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	commentID := r.PathValue("comment_id")
	comment, err := s.getEntityComment(commentID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "comment not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query comment")
		return
	}
	if comment.OwnerID != user.ID {
		sendError(w, http.StatusForbidden, "only the comment owner can delete")
		return
	}
	if comment.IsDeleted {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = s.db.Conn().Exec(`
		UPDATE comments
		SET is_deleted = 1, deleted_at = ?, updated_at = ?
		WHERE id = ?
	`, time.Now(), time.Now(), commentID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to delete comment")
		return
	}

	_, err = s.db.Conn().Exec(`
		INSERT INTO comment_history (id, comment_id, editor_id, editor_username, action, text)
		VALUES (?, ?, ?, ?, 'soft_delete', ?)
	`, uuid.New().String(), commentID, user.ID, user.Username, comment.Text)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to create comment history")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetCommentHistory(w http.ResponseWriter, r *http.Request) {
	commentID := r.PathValue("comment_id")

	_, err := s.getEntityComment(commentID)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusNotFound, "comment not found")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query comment")
		return
	}

	rows, err := s.db.Conn().Query(`
		SELECT id, comment_id, editor_id, editor_username, action, text, created_at
		FROM comment_history
		WHERE comment_id = ?
		ORDER BY created_at ASC
	`, commentID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to query comment history")
		return
	}
	defer rows.Close()

	history := []db.EntityCommentHistory{}
	for rows.Next() {
		var item db.EntityCommentHistory
		var createdAt string
		if err := rows.Scan(&item.ID, &item.CommentID, &item.EditorID, &item.EditorUsername, &item.Action, &item.Text, &createdAt); err != nil {
			sendError(w, http.StatusInternalServerError, "failed to scan comment history")
			return
		}
		item.CreatedAt = parseTimestamp(createdAt)
		history = append(history, item)
	}

	sendJSON(w, http.StatusOK, history)
}

func (s *Server) getEntityComment(commentID string) (db.EntityComment, error) {
	var item db.EntityComment
	var isDeleted int
	var createdAt, updatedAt string
	var deletedAt sql.NullString
	err := s.db.Conn().QueryRow(`
		SELECT id, entity_type, entity_id, owner_id, owner_username, text, is_deleted, created_at, updated_at, deleted_at
		FROM comments
		WHERE id = ?
	`, commentID).Scan(
		&item.ID,
		&item.EntityType,
		&item.EntityID,
		&item.OwnerID,
		&item.OwnerUsername,
		&item.Text,
		&isDeleted,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return db.EntityComment{}, err
	}
	item.IsDeleted = isDeleted == 1
	item.CreatedAt = parseTimestamp(createdAt)
	item.UpdatedAt = parseTimestamp(updatedAt)
	if deletedAt.Valid {
		t := parseTimestamp(deletedAt.String)
		item.DeletedAt = &t
	}
	return item, nil
}

func scanEntityComment(scanner interface {
	Scan(dest ...interface{}) error
}) (db.EntityComment, error) {
	var item db.EntityComment
	var isDeleted int
	var createdAt, updatedAt string
	var deletedAt sql.NullString
	err := scanner.Scan(
		&item.ID,
		&item.EntityType,
		&item.EntityID,
		&item.OwnerID,
		&item.OwnerUsername,
		&item.Text,
		&isDeleted,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return db.EntityComment{}, err
	}
	item.IsDeleted = isDeleted == 1
	item.CreatedAt = parseTimestamp(createdAt)
	item.UpdatedAt = parseTimestamp(updatedAt)
	if deletedAt.Valid {
		t := parseTimestamp(deletedAt.String)
		item.DeletedAt = &t
	}
	return item, nil
}
