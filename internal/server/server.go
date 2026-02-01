package server

import (
	"net/http"

	"github.com/simonski/task/internal/db"
)

// Server represents the HTTP server
type Server struct {
	db *db.DB
}

// New creates a new server instance
func New(database *db.DB) *Server {
	return &Server{
		db: database,
	}
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

	return mux
}
