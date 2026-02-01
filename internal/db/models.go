package db

import (
	"time"
)

// User represents a human, worker, or orchestrator user
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`    // Never expose password hash in JSON
	Type         string    `json:"type"` // human, worker, orchestrator
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Project represents a workspace containing tasks
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Repository  *string   `json:"repository,omitempty"`
	Status      string    `json:"status"`     // active, inactive
	Visibility  string    `json:"visibility"` // public, internal, private
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by"`
	UpdatedBy   string    `json:"updated_by"`
}

// ProjectMember represents project membership
type ProjectMember struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Role represents a job description for workers
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Rules       string    `json:"rules"`
	Scope       string    `json:"scope"` // system, global, project
	ProjectID   *string   `json:"project_id,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by"`
	UpdatedBy   string    `json:"updated_by"`
}

// Task represents a unit of work
type Task struct {
	ID                 string     `json:"id"`
	ProjectID          string     `json:"project_id"`
	Title              string     `json:"title"`
	Type               string     `json:"type"` // epic, story, task, sub-task, bug, spike
	Description        string     `json:"description"`
	AcceptanceCriteria *string    `json:"acceptance_criteria,omitempty"`
	ParentID           *string    `json:"parent_id,omitempty"`
	EpicID             *string    `json:"epic_id,omitempty"`
	DependsOnTaskID    *string    `json:"depends_on_task_id,omitempty"`
	Status             string     `json:"status"` // idle, active
	WorkerID           *string    `json:"worker_id,omitempty"`
	Priority           string     `json:"priority"` // low, medium, high, critical
	IsComplete         bool       `json:"is_complete"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	CreatedBy          string     `json:"created_by"`
	UpdatedBy          string     `json:"updated_by"`
	Labels             []string   `json:"labels,omitempty"`
	EstimatedEffort    *int       `json:"estimated_effort,omitempty"`
	ActualEffort       *int       `json:"actual_effort,omitempty"`
}

// TaskHistory represents a work session on a task
type TaskHistory struct {
	ID          string     `json:"id"`
	TaskID      string     `json:"task_id"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	State       string     `json:"state"` // success, failure, abandoned, in-progress
	WorkerID    string     `json:"worker_id"`
	RoleID      string     `json:"role_id"`
	Notes       *string    `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Config represents a system configuration setting
type Config struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
