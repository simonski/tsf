package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/simonski/task/internal/db"
)

// testServer creates a test server with a temporary database
func testServer(t *testing.T) (*Server, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := database.InitSchema(); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	// Create admin user
	adminHash, _ := db.HashPassword("admin123")
	_, err = database.Conn().Exec(`
		INSERT INTO users (id, username, password_hash, type, is_active)
		VALUES ('admin-id', 'admin', ?, 'human', 1)
	`, adminHash)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create a default project
	_, err = database.Conn().Exec(`
		INSERT INTO projects (id, name, description, status, created_by, updated_by)
		VALUES ('project-1', 'Test Project', 'A test project', 'active', 'admin-id', 'admin-id')
	`)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Create a default role
	_, err = database.Conn().Exec(`
		INSERT INTO roles (id, name, description, goals, scope, created_by, updated_by)
		VALUES ('role-1', 'programmer', 'Programmer role', 'Write code', 'system', 'admin-id', 'admin-id')
	`)
	if err != nil {
		t.Fatalf("Failed to create test role: %v", err)
	}

	server := New(database)

	cleanup := func() {
		database.Close()
	}

	return server, cleanup
}

// doRequest performs an HTTP request against the test server
func doRequest(t *testing.T, s *Server, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	return doRequestAs(t, s, method, path, body, "admin", "admin123")
}

func doRequestAs(t *testing.T, s *Server, method, path string, body interface{}, username, password string) *httptest.ResponseRecorder {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.SetBasicAuth(username, password)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	s.Router().ServeHTTP(rr, req)
	return rr
}

// parseJSON parses JSON response body
func parseJSON(t *testing.T, rr *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(rr.Body.Bytes(), v); err != nil {
		t.Fatalf("Failed to parse JSON: %v, body: %s", err, rr.Body.String())
	}
}

func TestHealthEndpoint(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rr := httptest.NewRecorder()
	s.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	parseJSON(t, rr, &resp)
	if resp["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", resp["status"])
	}
}

func TestAuthRequired(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// Request without auth should fail
	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	rr := httptest.NewRecorder()
	s.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestProjectCRUD(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// List projects
	rr := doRequest(t, s, "GET", "/api/v1/projects", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("List projects failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	var projects []map[string]interface{}
	parseJSON(t, rr, &projects)
	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}

	// Create project
	rr = doRequest(t, s, "POST", "/api/v1/projects", map[string]string{
		"name":        "New Project",
		"description": "A new project",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create project failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	var project map[string]interface{}
	parseJSON(t, rr, &project)
	projectID := project["id"].(string)
	if project["name"] != "New Project" {
		t.Errorf("Expected name 'New Project', got %v", project["name"])
	}

	// Get project
	rr = doRequest(t, s, "GET", "/api/v1/projects/"+projectID, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Get project failed: %d", rr.Code)
	}

	// Update project
	rr = doRequest(t, s, "PUT", "/api/v1/projects/"+projectID, map[string]string{
		"name": "Updated Project",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("Update project failed: %d", rr.Code)
	}

	parseJSON(t, rr, &project)
	if project["name"] != "Updated Project" {
		t.Errorf("Expected name 'Updated Project', got %v", project["name"])
	}

	// Delete project
	rr = doRequest(t, s, "DELETE", "/api/v1/projects/"+projectID, nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("Delete project failed: %d", rr.Code)
	}
}

func TestTaskCRUD(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// Create task
	rr := doRequest(t, s, "POST", "/api/v1/tasks", map[string]interface{}{
		"project_id":  "project-1",
		"title":       "Test Task",
		"description": "A test task",
		"type":        "task",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create task failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	var task map[string]interface{}
	parseJSON(t, rr, &task)
	taskID := task["id"].(string)
	if task["title"] != "Test Task" {
		t.Errorf("Expected title 'Test Task', got %v", task["title"])
	}
	if task["status"] != "idle" {
		t.Errorf("Expected status 'idle', got %v", task["status"])
	}

	// Get task
	rr = doRequest(t, s, "GET", "/api/v1/tasks/"+taskID, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Get task failed: %d", rr.Code)
	}

	// Update task
	rr = doRequest(t, s, "PUT", "/api/v1/tasks/"+taskID, map[string]interface{}{
		"title":    "Updated Task",
		"priority": "high",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("Update task failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	parseJSON(t, rr, &task)
	if task["title"] != "Updated Task" {
		t.Errorf("Expected title 'Updated Task', got %v", task["title"])
	}

	// List tasks
	rr = doRequest(t, s, "GET", "/api/v1/tasks", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("List tasks failed: %d", rr.Code)
	}

	var tasks []map[string]interface{}
	parseJSON(t, rr, &tasks)
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// Delete task
	rr = doRequest(t, s, "DELETE", "/api/v1/tasks/"+taskID, nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("Delete task failed: %d", rr.Code)
	}
}

func TestTaskWorkflow(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// Create task
	rr := doRequest(t, s, "POST", "/api/v1/tasks", map[string]interface{}{
		"project_id":  "project-1",
		"title":       "Workflow Task",
		"description": "Testing workflow",
		"type":        "task",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create task failed: %d", rr.Code)
	}

	var task map[string]interface{}
	parseJSON(t, rr, &task)
	taskID := task["id"].(string)

	// Claim task
	rr = doRequest(t, s, "POST", "/api/v1/tasks/"+taskID+"/claim", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Claim task failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	parseJSON(t, rr, &task)
	if task["status"] != "active" {
		t.Errorf("Expected status 'active', got %v", task["status"])
	}

	// Check history
	rr = doRequest(t, s, "GET", "/api/v1/tasks/"+taskID+"/history", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Get history failed: %d", rr.Code)
	}

	var history []map[string]interface{}
	parseJSON(t, rr, &history)
	if len(history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(history))
	}
	if history[0]["state"] != "in-progress" {
		t.Errorf("Expected state 'in-progress', got %v", history[0]["state"])
	}

	// Complete task
	rr = doRequest(t, s, "POST", "/api/v1/tasks/"+taskID+"/complete", map[string]string{
		"notes": "Task completed successfully",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("Complete task failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	parseJSON(t, rr, &task)
	if task["is_complete"] != true {
		t.Errorf("Expected is_complete true, got %v", task["is_complete"])
	}

	// Check history updated
	rr = doRequest(t, s, "GET", "/api/v1/tasks/"+taskID+"/history", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Get history failed: %d", rr.Code)
	}

	parseJSON(t, rr, &history)
	if history[0]["state"] != "success" {
		t.Errorf("Expected state 'success', got %v", history[0]["state"])
	}
}

func TestProjectFileAndNoteCRUD(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// Create project file
	rr := doRequest(t, s, "POST", "/api/v1/projects/project-1/files", map[string]string{
		"name":    "README.md",
		"content": "hello",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create project file failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	var file map[string]interface{}
	parseJSON(t, rr, &file)
	fileID := file["id"].(string)

	// List project files
	rr = doRequest(t, s, "GET", "/api/v1/projects/project-1/files", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("List project files failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	var files []map[string]interface{}
	parseJSON(t, rr, &files)
	if len(files) != 1 {
		t.Fatalf("Expected 1 project file, got %d", len(files))
	}

	// Update project file
	rr = doRequest(t, s, "PUT", "/api/v1/projects/project-1/files/"+fileID, map[string]string{
		"content": "updated",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("Update project file failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	// Get project file
	rr = doRequest(t, s, "GET", "/api/v1/projects/project-1/files/"+fileID, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Get project file failed: %d, body: %s", rr.Code, rr.Body.String())
	}
	parseJSON(t, rr, &file)
	if file["content"] != "updated" {
		t.Fatalf("Expected updated content, got %v", file["content"])
	}

	// Delete project file
	rr = doRequest(t, s, "DELETE", "/api/v1/projects/project-1/files/"+fileID, nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("Delete project file failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	// Create project note
	rr = doRequest(t, s, "POST", "/api/v1/projects/project-1/notes", map[string]string{
		"title":   "Kickoff",
		"content": "Start here",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create project note failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	var note map[string]interface{}
	parseJSON(t, rr, &note)
	noteID := note["id"].(string)

	// List project notes
	rr = doRequest(t, s, "GET", "/api/v1/projects/project-1/notes", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("List project notes failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	var notes []map[string]interface{}
	parseJSON(t, rr, &notes)
	if len(notes) != 1 {
		t.Fatalf("Expected 1 project note, got %d", len(notes))
	}

	// Update project note
	rr = doRequest(t, s, "PUT", "/api/v1/projects/project-1/notes/"+noteID, map[string]string{
		"content": "Updated note",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("Update project note failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	// Get project note
	rr = doRequest(t, s, "GET", "/api/v1/projects/project-1/notes/"+noteID, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Get project note failed: %d, body: %s", rr.Code, rr.Body.String())
	}
	parseJSON(t, rr, &note)
	if note["content"] != "Updated note" {
		t.Fatalf("Expected updated note content, got %v", note["content"])
	}

	// Delete project note
	rr = doRequest(t, s, "DELETE", "/api/v1/projects/project-1/notes/"+noteID, nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("Delete project note failed: %d, body: %s", rr.Code, rr.Body.String())
	}
}

func TestProjectFileAndNoteValidation(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// Missing file name
	rr := doRequest(t, s, "POST", "/api/v1/projects/project-1/files", map[string]string{
		"content": "x",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for missing file name, got %d", rr.Code)
	}

	// Unknown project
	rr = doRequest(t, s, "POST", "/api/v1/projects/nope/files", map[string]string{
		"name": "x",
	})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for unknown project file create, got %d", rr.Code)
	}

	// Create file then duplicate by name in same project
	rr = doRequest(t, s, "POST", "/api/v1/projects/project-1/files", map[string]string{
		"name":    "dup.txt",
		"content": "a",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected 201 for create file, got %d", rr.Code)
	}
	rr = doRequest(t, s, "POST", "/api/v1/projects/project-1/files", map[string]string{
		"name":    "dup.txt",
		"content": "b",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for duplicate file name, got %d", rr.Code)
	}

	// Get/Update/Delete unknown file IDs
	rr = doRequest(t, s, "GET", "/api/v1/projects/project-1/files/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for missing file get, got %d", rr.Code)
	}
	rr = doRequest(t, s, "PUT", "/api/v1/projects/project-1/files/missing", map[string]string{
		"content": "x",
	})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for missing file update, got %d", rr.Code)
	}
	rr = doRequest(t, s, "DELETE", "/api/v1/projects/project-1/files/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for missing file delete, got %d", rr.Code)
	}

	// Missing note title
	rr = doRequest(t, s, "POST", "/api/v1/projects/project-1/notes", map[string]string{
		"content": "x",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for missing note title, got %d", rr.Code)
	}

	// Unknown project note create
	rr = doRequest(t, s, "POST", "/api/v1/projects/nope/notes", map[string]string{
		"title": "x",
	})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for unknown project note create, got %d", rr.Code)
	}

	// Get/Update/Delete unknown note IDs
	rr = doRequest(t, s, "GET", "/api/v1/projects/project-1/notes/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for missing note get, got %d", rr.Code)
	}
	rr = doRequest(t, s, "PUT", "/api/v1/projects/project-1/notes/missing", map[string]string{
		"content": "x",
	})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for missing note update, got %d", rr.Code)
	}
	rr = doRequest(t, s, "DELETE", "/api/v1/projects/project-1/notes/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for missing note delete, got %d", rr.Code)
	}
}

func TestEntityCommentCRUDHistoryAndOwnership(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// Create a task to comment on
	rr := doRequest(t, s, "POST", "/api/v1/tasks", map[string]interface{}{
		"project_id":  "project-1",
		"title":       "Commented task",
		"description": "Task for comment tests",
		"type":        "task",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create task failed: %d, body: %s", rr.Code, rr.Body.String())
	}
	var task map[string]interface{}
	parseJSON(t, rr, &task)
	taskID := task["id"].(string)

	// Create comment on task
	rr = doRequest(t, s, "POST", "/api/v1/comments", map[string]string{
		"entity_type": "task",
		"entity_id":   taskID,
		"text":        "first",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create comment failed: %d, body: %s", rr.Code, rr.Body.String())
	}
	var comment map[string]interface{}
	parseJSON(t, rr, &comment)
	commentID := comment["id"].(string)

	// Update as owner
	rr = doRequest(t, s, "PUT", "/api/v1/comments/"+commentID, map[string]string{
		"text": "edited",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("Update comment failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	// History should have create+edit
	rr = doRequest(t, s, "GET", "/api/v1/comments/"+commentID+"/history", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Get comment history failed: %d, body: %s", rr.Code, rr.Body.String())
	}
	var history []map[string]interface{}
	parseJSON(t, rr, &history)
	if len(history) != 2 {
		t.Fatalf("Expected 2 history entries, got %d", len(history))
	}

	// Add a second user and verify non-owner cannot edit/delete
	userHash, _ := db.HashPassword("user123")
	_, err := s.db.Conn().Exec(`
		INSERT INTO users (id, username, password_hash, type, is_active)
		VALUES ('user-2', 'user2', ?, 'human', 1)
	`, userHash)
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}
	rr = doRequestAs(t, s, "PUT", "/api/v1/comments/"+commentID, map[string]string{
		"text": "hijack",
	}, "user2", "user123")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 for non-owner edit, got %d, body: %s", rr.Code, rr.Body.String())
	}
	rr = doRequestAs(t, s, "DELETE", "/api/v1/comments/"+commentID, nil, "user2", "user123")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 for non-owner delete, got %d, body: %s", rr.Code, rr.Body.String())
	}

	// Soft-delete as owner
	rr = doRequest(t, s, "DELETE", "/api/v1/comments/"+commentID, nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("Delete comment failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	// Default list excludes deleted
	rr = doRequest(t, s, "GET", "/api/v1/comments?entity_type=task&entity_id="+taskID, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("List comments failed: %d, body: %s", rr.Code, rr.Body.String())
	}
	var comments []map[string]interface{}
	parseJSON(t, rr, &comments)
	if len(comments) != 0 {
		t.Fatalf("Expected 0 non-deleted comments, got %d", len(comments))
	}

	// Include deleted returns the comment
	rr = doRequest(t, s, "GET", "/api/v1/comments?entity_type=task&entity_id="+taskID+"&include_deleted=true", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("List comments (include_deleted) failed: %d, body: %s", rr.Code, rr.Body.String())
	}
	parseJSON(t, rr, &comments)
	if len(comments) != 1 {
		t.Fatalf("Expected 1 comment with include_deleted, got %d", len(comments))
	}
}

func TestEntityCommentProjectAndValidation(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// Create comment on project
	rr := doRequest(t, s, "POST", "/api/v1/comments", map[string]string{
		"entity_type": "project",
		"entity_id":   "project-1",
		"text":        "project-level comment",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected 201 for project comment create, got %d, body: %s", rr.Code, rr.Body.String())
	}
	var comment map[string]interface{}
	parseJSON(t, rr, &comment)
	commentID := comment["id"].(string)

	// List project comments
	rr = doRequest(t, s, "GET", "/api/v1/comments?entity_type=project&entity_id=project-1", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for project comments list, got %d", rr.Code)
	}
	var comments []map[string]interface{}
	parseJSON(t, rr, &comments)
	if len(comments) != 1 {
		t.Fatalf("Expected 1 project comment, got %d", len(comments))
	}

	// Delete and verify include_deleted filter
	rr = doRequest(t, s, "DELETE", "/api/v1/comments/"+commentID, nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("Expected 204 for comment soft-delete, got %d", rr.Code)
	}
	rr = doRequest(t, s, "GET", "/api/v1/comments?entity_type=project&entity_id=project-1", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for project comments list post-delete, got %d", rr.Code)
	}
	parseJSON(t, rr, &comments)
	if len(comments) != 0 {
		t.Fatalf("Expected 0 non-deleted comments, got %d", len(comments))
	}
	rr = doRequest(t, s, "GET", "/api/v1/comments?entity_type=project&entity_id=project-1&include_deleted=true", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for include_deleted list, got %d", rr.Code)
	}
	parseJSON(t, rr, &comments)
	if len(comments) != 1 {
		t.Fatalf("Expected 1 deleted comment with include_deleted=true, got %d", len(comments))
	}

	// Validation / unknown entity cases
	rr = doRequest(t, s, "POST", "/api/v1/comments", map[string]string{
		"entity_type": "invalid",
		"entity_id":   "x",
		"text":        "x",
	})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 invalid entity_type, got %d", rr.Code)
	}
	rr = doRequest(t, s, "POST", "/api/v1/comments", map[string]string{
		"entity_type": "task",
		"entity_id":   "missing-task",
		"text":        "x",
	})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 missing entity, got %d", rr.Code)
	}
	rr = doRequest(t, s, "GET", "/api/v1/comments?entity_type=task&entity_id=missing-task", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 listing comments for missing entity, got %d", rr.Code)
	}
}

func TestTaskDependencies(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// Create dependency task
	rr := doRequest(t, s, "POST", "/api/v1/tasks", map[string]interface{}{
		"project_id":  "project-1",
		"title":       "Dependency Task",
		"description": "This task blocks another",
		"type":        "task",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create dependency task failed: %d", rr.Code)
	}

	var depTask map[string]interface{}
	parseJSON(t, rr, &depTask)
	depTaskID := depTask["id"].(string)

	// Create blocked task
	rr = doRequest(t, s, "POST", "/api/v1/tasks", map[string]interface{}{
		"project_id":         "project-1",
		"title":              "Blocked Task",
		"description":        "This task is blocked",
		"type":               "task",
		"depends_on_task_id": depTaskID,
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create blocked task failed: %d", rr.Code)
	}

	var blockedTask map[string]interface{}
	parseJSON(t, rr, &blockedTask)
	blockedTaskID := blockedTask["id"].(string)

	// Try to claim blocked task - should fail
	rr = doRequest(t, s, "POST", "/api/v1/tasks/"+blockedTaskID+"/claim", nil)
	if rr.Code != http.StatusConflict {
		t.Errorf("Expected 409 Conflict, got %d", rr.Code)
	}

	// Claim and complete dependency task
	rr = doRequest(t, s, "POST", "/api/v1/tasks/"+depTaskID+"/claim", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Claim dependency task failed: %d", rr.Code)
	}

	rr = doRequest(t, s, "POST", "/api/v1/tasks/"+depTaskID+"/complete", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Complete dependency task failed: %d", rr.Code)
	}

	// Now claim blocked task - should succeed
	rr = doRequest(t, s, "POST", "/api/v1/tasks/"+blockedTaskID+"/claim", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Claim blocked task failed after dependency complete: %d, body: %s", rr.Code, rr.Body.String())
	}

	// Check dependencies endpoint
	rr = doRequest(t, s, "GET", "/api/v1/tasks/"+depTaskID+"/dependencies", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Get dependencies failed: %d", rr.Code)
	}

	var deps map[string]interface{}
	parseJSON(t, rr, &deps)
	if deps["blocking"] == nil {
		t.Error("Expected blocking array in dependencies")
	}
}

func TestJWTAuth(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// Login
	loginBody := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	bodyBytes, _ := json.Marshal(loginBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Login failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	var loginResp map[string]interface{}
	parseJSON(t, rr, &loginResp)
	token := loginResp["token"].(string)
	refreshToken := loginResp["refresh_token"].(string)

	if token == "" {
		t.Error("Expected JWT token in response")
	}
	if refreshToken == "" {
		t.Error("Expected refresh token in response")
	}

	// Use JWT token to access protected endpoint
	req = httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	s.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("JWT auth failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	var userResp map[string]interface{}
	parseJSON(t, rr, &userResp)
	if userResp["username"] != "admin" {
		t.Errorf("Expected username 'admin', got %v", userResp["username"])
	}

	// Refresh token
	refreshBody := map[string]string{
		"refresh_token": refreshToken,
	}
	bodyBytes, _ = json.Marshal(refreshBody)
	req = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	s.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Token refresh failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	var refreshResp map[string]interface{}
	parseJSON(t, rr, &refreshResp)
	if refreshResp["token"] == "" {
		t.Error("Expected new JWT token in refresh response")
	}
}

func TestMetricsEndpoint(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// Create some tasks for metrics
	rr := doRequest(t, s, "POST", "/api/v1/tasks", map[string]interface{}{
		"project_id":  "project-1",
		"title":       "Task 1",
		"description": "Test task 1",
		"type":        "task",
		"priority":    "high",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create task 1 failed: %d, body: %s", rr.Code, rr.Body.String())
	}
	rr = doRequest(t, s, "POST", "/api/v1/tasks", map[string]interface{}{
		"project_id":  "project-1",
		"title":       "Task 2",
		"description": "Test task 2",
		"type":        "task",
		"priority":    "low",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create task 2 failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	// Get metrics
	rr = doRequest(t, s, "GET", "/api/v1/metrics", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Get metrics failed: %d", rr.Code)
	}

	var metrics map[string]interface{}
	parseJSON(t, rr, &metrics)

	tasks := metrics["tasks"].(map[string]interface{})
	if tasks["total"].(float64) != 2 {
		t.Errorf("Expected 2 tasks, got %v", tasks["total"])
	}

	// Test Prometheus endpoint (no auth)
	req := httptest.NewRequest("GET", "/metrics", nil)
	rr = httptest.NewRecorder()
	s.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Prometheus metrics failed: %d", rr.Code)
	}

	body := rr.Body.String()
	if !bytes.Contains([]byte(body), []byte("task_total")) {
		t.Error("Expected task_total metric in Prometheus output")
	}
}

func TestUserManagement(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// List users
	rr := doRequest(t, s, "GET", "/api/v1/users", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("List users failed: %d", rr.Code)
	}

	var users []map[string]interface{}
	parseJSON(t, rr, &users)
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	// Register new user
	regBody := map[string]string{
		"username": "testuser",
		"password": "testpass123",
		"type":     "human",
	}
	bodyBytes, _ := json.Marshal(regBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	s.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Register user failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	// Disable user (admin only)
	rr = doRequest(t, s, "POST", "/api/v1/users/testuser/disable", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Disable user failed: %d", rr.Code)
	}

	// Enable user
	rr = doRequest(t, s, "POST", "/api/v1/users/testuser/enable", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Enable user failed: %d", rr.Code)
	}
}

func TestRoleCRUD(t *testing.T) {
	s, cleanup := testServer(t)
	defer cleanup()

	// List roles
	rr := doRequest(t, s, "GET", "/api/v1/roles", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("List roles failed: %d", rr.Code)
	}

	// Create role
	rr = doRequest(t, s, "POST", "/api/v1/roles", map[string]string{
		"name":        "tester",
		"description": "Tester role",
		"goals":       "{}",
		"scope":       "global",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("Create role failed: %d, body: %s", rr.Code, rr.Body.String())
	}

	var role map[string]interface{}
	parseJSON(t, rr, &role)
	roleID := role["id"].(string)

	// Get role
	rr = doRequest(t, s, "GET", "/api/v1/roles/"+roleID, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Get role failed: %d", rr.Code)
	}

	// Delete role
	rr = doRequest(t, s, "DELETE", "/api/v1/roles/"+roleID, nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("Delete role failed: %d", rr.Code)
	}
}
