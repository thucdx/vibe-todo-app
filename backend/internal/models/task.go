package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Priority represents the urgency level of a task.
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// Task is the database and API representation of a single task.
type Task struct {
	ID        uuid.UUID      `db:"id"         json:"id"`
	Title     string         `db:"title"      json:"title"`
	Date      Date           `db:"date"       json:"date"`
	DueTime   *string        `db:"due_time"   json:"due_time"`  // "HH:MM" or null
	Priority  Priority       `db:"priority"   json:"priority"`
	Tags      pq.StringArray `db:"tags"       json:"tags"`
	Points    int            `db:"points"     json:"points"`
	Done      bool           `db:"done"       json:"done"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
}

// CreateTaskInput is the request body for POST /api/v1/tasks.
type CreateTaskInput struct {
	Title    string   `json:"title"    binding:"required,max=255"`
	Date     string   `json:"date"     binding:"required"` // YYYY-MM-DD
	DueTime  *string  `json:"due_time"`                     // HH:MM or null
	Priority Priority `json:"priority"`
	Tags     []string `json:"tags"`
	Points   int      `json:"points"`
}

// UpdateTaskInput is the request body for PUT /api/v1/tasks/:id.
type UpdateTaskInput struct {
	Title    string   `json:"title"    binding:"required,max=255"`
	Date     string   `json:"date"     binding:"required"` // YYYY-MM-DD
	DueTime  *string  `json:"due_time"`
	Priority Priority `json:"priority"`
	Tags     []string `json:"tags"`
	Points   int      `json:"points"`
}
