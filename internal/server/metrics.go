package server

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// MetricsResponse contains server metrics
type MetricsResponse struct {
	Tasks     TaskMetrics     `json:"tasks"`
	Users     UserMetrics     `json:"users"`
	Projects  ProjectMetrics  `json:"projects"`
	System    SystemMetrics   `json:"system"`
	Timestamp time.Time       `json:"timestamp"`
}

// TaskMetrics contains task-related metrics
type TaskMetrics struct {
	Total        int            `json:"total"`
	ByStatus     map[string]int `json:"by_status"`
	ByPriority   map[string]int `json:"by_priority"`
	Completed    int            `json:"completed"`
	InProgress   int            `json:"in_progress"`
	Blocked      int            `json:"blocked"`
	AvgEffort    float64        `json:"avg_estimated_effort"`
	CompletedToday int          `json:"completed_today"`
}

// UserMetrics contains user-related metrics
type UserMetrics struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Inactive int `json:"inactive"`
	Human    int `json:"human"`
	Agent    int `json:"agent"`
}

// ProjectMetrics contains project-related metrics
type ProjectMetrics struct {
	Total int `json:"total"`
}

// SystemMetrics contains system metrics
type SystemMetrics struct {
	UptimeSeconds  int64   `json:"uptime_seconds"`
	GoRoutines     int     `json:"goroutines"`
	HeapAllocMB    float64 `json:"heap_alloc_mb"`
	HeapInUseMB    float64 `json:"heap_inuse_mb"`
	NumGC          uint32  `json:"num_gc"`
}

var serverStartTime = time.Now()

// handleMetrics returns server metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := MetricsResponse{
		Timestamp: time.Now(),
	}

	// Task metrics
	metrics.Tasks.ByStatus = make(map[string]int)
	metrics.Tasks.ByPriority = make(map[string]int)

	rows, err := s.db.Conn().Query("SELECT status, COUNT(*) FROM tasks GROUP BY status")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			if rows.Scan(&status, &count) == nil {
				metrics.Tasks.ByStatus[status] = count
				metrics.Tasks.Total += count
				if status == "active" {
					metrics.Tasks.InProgress = count
				}
			}
		}
	}

	rows, err = s.db.Conn().Query("SELECT priority, COUNT(*) FROM tasks GROUP BY priority")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var priority string
			var count int
			if rows.Scan(&priority, &count) == nil {
				metrics.Tasks.ByPriority[priority] = count
			}
		}
	}

	s.db.Conn().QueryRow("SELECT COUNT(*) FROM tasks WHERE is_complete = 1").Scan(&metrics.Tasks.Completed)
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM tasks WHERE depends_on_task_id IS NOT NULL AND is_complete = 0").Scan(&metrics.Tasks.Blocked)
	s.db.Conn().QueryRow("SELECT AVG(estimated_effort) FROM tasks WHERE estimated_effort IS NOT NULL").Scan(&metrics.Tasks.AvgEffort)
	
	// Tasks completed today
	today := time.Now().Format("2006-01-02")
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM tasks WHERE is_complete = 1 AND completed_at >= ?", today).Scan(&metrics.Tasks.CompletedToday)

	// User metrics
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM users").Scan(&metrics.Users.Total)
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM users WHERE is_active = 1").Scan(&metrics.Users.Active)
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM users WHERE is_active = 0").Scan(&metrics.Users.Inactive)
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM users WHERE type = 'human'").Scan(&metrics.Users.Human)
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM users WHERE type = 'agent'").Scan(&metrics.Users.Agent)

	// Project metrics
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM projects").Scan(&metrics.Projects.Total)

	// System metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	metrics.System.UptimeSeconds = int64(time.Since(serverStartTime).Seconds())
	metrics.System.GoRoutines = runtime.NumGoroutine()
	metrics.System.HeapAllocMB = float64(memStats.HeapAlloc) / 1024 / 1024
	metrics.System.HeapInUseMB = float64(memStats.HeapInuse) / 1024 / 1024
	metrics.System.NumGC = memStats.NumGC

	sendJSON(w, http.StatusOK, metrics)
}

// handlePrometheusMetrics returns metrics in Prometheus format
func (s *Server) handlePrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Task metrics
	var total, completed, inProgress, blocked int
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM tasks").Scan(&total)
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM tasks WHERE is_complete = 1").Scan(&completed)
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'active'").Scan(&inProgress)
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM tasks WHERE depends_on_task_id IS NOT NULL AND is_complete = 0").Scan(&blocked)

	fmt.Fprintf(w, "# HELP task_total Total number of tasks\n")
	fmt.Fprintf(w, "# TYPE task_total gauge\n")
	fmt.Fprintf(w, "task_total %d\n", total)

	fmt.Fprintf(w, "# HELP task_completed_total Number of completed tasks\n")
	fmt.Fprintf(w, "# TYPE task_completed_total gauge\n")
	fmt.Fprintf(w, "task_completed_total %d\n", completed)

	fmt.Fprintf(w, "# HELP task_in_progress Number of tasks in progress\n")
	fmt.Fprintf(w, "# TYPE task_in_progress gauge\n")
	fmt.Fprintf(w, "task_in_progress %d\n", inProgress)

	fmt.Fprintf(w, "# HELP task_blocked Number of blocked tasks\n")
	fmt.Fprintf(w, "# TYPE task_blocked gauge\n")
	fmt.Fprintf(w, "task_blocked %d\n", blocked)

	// Task by status
	rows, _ := s.db.Conn().Query("SELECT status, COUNT(*) FROM tasks GROUP BY status")
	if rows != nil {
		defer rows.Close()
		fmt.Fprintf(w, "# HELP task_by_status Tasks by status\n")
		fmt.Fprintf(w, "# TYPE task_by_status gauge\n")
		for rows.Next() {
			var status string
			var count int
			if rows.Scan(&status, &count) == nil {
				fmt.Fprintf(w, "task_by_status{status=\"%s\"} %d\n", status, count)
			}
		}
	}

	// Task by priority
	rows, _ = s.db.Conn().Query("SELECT priority, COUNT(*) FROM tasks GROUP BY priority")
	if rows != nil {
		defer rows.Close()
		fmt.Fprintf(w, "# HELP task_by_priority Tasks by priority\n")
		fmt.Fprintf(w, "# TYPE task_by_priority gauge\n")
		for rows.Next() {
			var priority string
			var count int
			if rows.Scan(&priority, &count) == nil {
				fmt.Fprintf(w, "task_by_priority{priority=\"%s\"} %d\n", priority, count)
			}
		}
	}

	// User metrics
	var totalUsers, activeUsers int
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM users WHERE is_active = 1").Scan(&activeUsers)

	fmt.Fprintf(w, "# HELP user_total Total number of users\n")
	fmt.Fprintf(w, "# TYPE user_total gauge\n")
	fmt.Fprintf(w, "user_total %d\n", totalUsers)

	fmt.Fprintf(w, "# HELP user_active Number of active users\n")
	fmt.Fprintf(w, "# TYPE user_active gauge\n")
	fmt.Fprintf(w, "user_active %d\n", activeUsers)

	// Project metrics
	var totalProjects int
	s.db.Conn().QueryRow("SELECT COUNT(*) FROM projects").Scan(&totalProjects)

	fmt.Fprintf(w, "# HELP project_total Total number of projects\n")
	fmt.Fprintf(w, "# TYPE project_total gauge\n")
	fmt.Fprintf(w, "project_total %d\n", totalProjects)

	// System metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	fmt.Fprintf(w, "# HELP process_uptime_seconds Server uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE process_uptime_seconds counter\n")
	fmt.Fprintf(w, "process_uptime_seconds %d\n", int64(time.Since(serverStartTime).Seconds()))

	fmt.Fprintf(w, "# HELP go_goroutines Number of goroutines\n")
	fmt.Fprintf(w, "# TYPE go_goroutines gauge\n")
	fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine())

	fmt.Fprintf(w, "# HELP go_heap_alloc_bytes Heap allocation in bytes\n")
	fmt.Fprintf(w, "# TYPE go_heap_alloc_bytes gauge\n")
	fmt.Fprintf(w, "go_heap_alloc_bytes %d\n", memStats.HeapAlloc)

	fmt.Fprintf(w, "# HELP go_gc_cycles_total Number of GC cycles\n")
	fmt.Fprintf(w, "# TYPE go_gc_cycles_total counter\n")
	fmt.Fprintf(w, "go_gc_cycles_total %d\n", memStats.NumGC)
}
