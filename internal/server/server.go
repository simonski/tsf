package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/simonski/task/internal/db"
)

// Server represents the HTTP server
type Server struct {
	db    *db.DB
	webFS embed.FS
}

// New creates a new server instance
func New(database *db.DB) *Server {
	return &Server{
		db: database,
	}
}

// SetWebFS sets the embedded web filesystem
func (s *Server) SetWebFS(webFS embed.FS) {
	s.webFS = webFS
}

// Router returns the HTTP router
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// Health check (no auth required)
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	// Auth endpoints
	mux.HandleFunc("POST /api/v1/auth/register", s.handleRegisterUser)

	// User endpoints
	mux.HandleFunc("GET /api/v1/users", s.withAuth(s.handleListUsers))
	mux.HandleFunc("GET /api/v1/users/{user_id}", s.withAuth(s.handleGetUser))
	mux.HandleFunc("POST /api/v1/users/{username}/enable", s.withAuth(s.withAdmin(s.handleEnableUser)))
	mux.HandleFunc("POST /api/v1/users/{username}/disable", s.withAuth(s.withAdmin(s.handleDisableUser)))

	// Project endpoints
	mux.HandleFunc("GET /api/v1/projects", s.withAuth(s.handleListProjects))
	mux.HandleFunc("POST /api/v1/projects", s.withAuth(s.handleCreateProject))
	mux.HandleFunc("GET /api/v1/projects/{project_id}", s.withAuth(s.handleGetProject))
	mux.HandleFunc("PUT /api/v1/projects/{project_id}", s.withAuth(s.handleUpdateProject))
	mux.HandleFunc("DELETE /api/v1/projects/{project_id}", s.withAuth(s.handleDeleteProject))

	// Role endpoints
	mux.HandleFunc("GET /api/v1/roles", s.withAuth(s.handleListRoles))
	mux.HandleFunc("POST /api/v1/roles", s.withAuth(s.handleCreateRole))
	mux.HandleFunc("GET /api/v1/roles/{role_id}", s.withAuth(s.handleGetRole))
	mux.HandleFunc("PUT /api/v1/roles/{role_id}", s.withAuth(s.handleUpdateRole))
	mux.HandleFunc("DELETE /api/v1/roles/{role_id}", s.withAuth(s.handleDeleteRole))

	// Task endpoints
	mux.HandleFunc("GET /api/v1/tasks", s.withAuth(s.handleListTasks))
	mux.HandleFunc("POST /api/v1/tasks", s.withAuth(s.handleCreateTask))
	mux.HandleFunc("GET /api/v1/tasks/{task_id}", s.withAuth(s.handleGetTask))
	mux.HandleFunc("PUT /api/v1/tasks/{task_id}", s.withAuth(s.handleUpdateTask))
	mux.HandleFunc("DELETE /api/v1/tasks/{task_id}", s.withAuth(s.handleDeleteTask))
	mux.HandleFunc("POST /api/v1/tasks/{task_id}/claim", s.withAuth(s.handleClaimTask))
	mux.HandleFunc("POST /api/v1/tasks/{task_id}/assign", s.withAuth(s.handleAssignTask))
	mux.HandleFunc("POST /api/v1/tasks/{task_id}/free", s.withAuth(s.handleFreeTask))
	mux.HandleFunc("POST /api/v1/tasks/{task_id}/complete", s.withAuth(s.handleCompleteTask))
	mux.HandleFunc("GET /api/v1/tasks/{task_id}/history", s.withAuth(s.handleGetTaskHistory))

	// Worker endpoints
	mux.HandleFunc("POST /api/v1/workers/request", s.withAuth(s.handleWorkerRequest))
	mux.HandleFunc("POST /api/v1/workers/heartbeat", s.withAuth(s.handleWorkerHeartbeat))

	// Orchestrator endpoints
	mux.HandleFunc("POST /api/v1/orchestrator/heartbeat", s.withAuth(s.handleOrchestratorHeartbeat))

	// Config endpoints
	mux.HandleFunc("GET /api/v1/config", s.withAuth(s.handleListConfig))
	mux.HandleFunc("GET /api/v1/config/{key}", s.withAuth(s.handleGetConfig))
	mux.HandleFunc("PUT /api/v1/config/{key}", s.withAuth(s.withAdmin(s.handleSetConfig)))
	mux.HandleFunc("DELETE /api/v1/config/{key}", s.withAuth(s.withAdmin(s.handleDeleteConfig)))

	// Static file serving - always register these routes
	// Check if webFS has content by trying to read directory
	entries, err := fs.ReadDir(s.webFS, ".")
	hasWebFS := err == nil && len(entries) > 0
	log.Printf("WebFS has content: %v (entries: %d, err: %v)", hasWebFS, len(entries), err)

	if hasWebFS {
		// Serve static files and root from embedded FS
		fileServer := http.FileServer(http.FS(s.webFS))
		mux.Handle("GET /static/", fileServer)

		// Serve index.html for root and SPA routes
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			// For root and non-API paths, serve index.html
			if r.URL.Path == "/" || r.URL.Path == "/index.html" {
				data, err := fs.ReadFile(s.webFS, "index.html")
				if err != nil {
					log.Printf("Error reading index.html: %v", err)
					http.Error(w, "Not Found", http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(data)
				return
			}
			// For other paths, try to serve static file or fall back to index.html (SPA)
			_, err := fs.Stat(s.webFS, r.URL.Path[1:]) // Remove leading /
			if err != nil {
				// File not found, serve index.html for SPA routing
				data, err := fs.ReadFile(s.webFS, "index.html")
				if err != nil {
					http.Error(w, "Not Found", http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(data)
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	} else {
		// No web content, serve simple message at root
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Task Management Server - API available at /api/v1/"))
		})
	}

	return mux
}
