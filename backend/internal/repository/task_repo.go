package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/models"
)

// TaskRepo handles all task-related database operations.
type TaskRepo struct {
	db *sqlx.DB
}

func NewTaskRepo(db *sqlx.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

// DaySummary holds the done/total task counts for a single calendar day.
type DaySummary struct {
	Date  string `db:"date"  json:"date"`
	Done  int    `db:"done"  json:"done"`
	Total int    `db:"total" json:"total"`
}

// ChartPoint holds a single data point for the stats chart.
type ChartPoint struct {
	Label string `db:"label" json:"label"`
	Value int    `db:"value" json:"value"`
}

// ListByDate returns all tasks for a given date, sorted by status then priority then due_time.
func (r *TaskRepo) ListByDate(ctx context.Context, date string) ([]models.Task, error) {
	const q = `
		SELECT * FROM tasks
		WHERE date = $1::date
		ORDER BY
			done ASC,
			CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END,
			due_time ASC NULLS LAST,
			created_at ASC`
	tasks := make([]models.Task, 0)
	if err := r.db.SelectContext(ctx, &tasks, q, date); err != nil {
		return nil, fmt.Errorf("listByDate %q: %w", date, err)
	}
	return tasks, nil
}

// Create inserts a new task and returns the persisted record.
func (r *TaskRepo) Create(ctx context.Context, in models.CreateTaskInput) (*models.Task, error) {
	priority := in.Priority
	if priority == "" {
		priority = models.PriorityMedium
	}
	points := in.Points
	if points <= 0 {
		points = 1
	}
	tags := pq.StringArray(in.Tags)
	if tags == nil {
		tags = pq.StringArray{}
	}
	const q = `
		INSERT INTO tasks (title, date, due_time, priority, tags, points)
		VALUES ($1, $2::date, $3::time, $4, $5, $6)
		RETURNING *`
	var t models.Task
	err := r.db.QueryRowxContext(ctx, q, in.Title, in.Date, in.DueTime, priority, tags, points).StructScan(&t)
	if err != nil {
		return nil, fmt.Errorf("createTask: %w", err)
	}
	return &t, nil
}

// Update modifies all mutable fields on a task and returns the updated record.
func (r *TaskRepo) Update(ctx context.Context, id uuid.UUID, in models.UpdateTaskInput) (*models.Task, error) {
	priority := in.Priority
	if priority == "" {
		priority = models.PriorityMedium
	}
	points := in.Points
	if points <= 0 {
		points = 1
	}
	tags := pq.StringArray(in.Tags)
	if tags == nil {
		tags = pq.StringArray{}
	}
	const q = `
		UPDATE tasks
		SET title = $1, date = $2::date, due_time = $3::time, priority = $4, tags = $5, points = $6
		WHERE id = $7
		RETURNING *`
	var t models.Task
	err := r.db.QueryRowxContext(ctx, q, in.Title, in.Date, in.DueTime, priority, tags, points, id).StructScan(&t)
	if err != nil {
		return nil, fmt.Errorf("updateTask %s: %w", id, err)
	}
	return &t, nil
}

// ToggleDone flips the done status of a task and returns the updated record.
func (r *TaskRepo) ToggleDone(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	const q = `UPDATE tasks SET done = NOT done WHERE id = $1 RETURNING *`
	var t models.Task
	err := r.db.QueryRowxContext(ctx, q, id).StructScan(&t)
	if err != nil {
		return nil, fmt.Errorf("toggleDone %s: %w", id, err)
	}
	return &t, nil
}

// DeleteAll removes all tasks.
func (r *TaskRepo) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tasks`)
	return err
}

// Delete removes a task by ID.
func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleteTask %s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// CalendarSummary returns done/total task counts grouped by day for the given month.
func (r *TaskRepo) CalendarSummary(ctx context.Context, year, month int) ([]DaySummary, error) {
	const q = `
		SELECT
			TO_CHAR(date, 'YYYY-MM-DD') AS date,
			COUNT(*) FILTER (WHERE done) AS done,
			COUNT(*) AS total
		FROM tasks
		WHERE EXTRACT(YEAR  FROM date) = $1
		  AND EXTRACT(MONTH FROM date) = $2
		GROUP BY date
		ORDER BY date`
	rows := make([]DaySummary, 0)
	if err := r.db.SelectContext(ctx, &rows, q, year, month); err != nil {
		return nil, fmt.Errorf("calendarSummary %d-%02d: %w", year, month, err)
	}
	return rows, nil
}

// ChartData returns aggregated chart points for the given view and metric.
// view: "day" | "week" | "month"   metric: "count" | "points"
// Values are validated before this call so string interpolation is safe.
func (r *TaskRepo) ChartData(ctx context.Context, view, metric string) ([]ChartPoint, error) {
	var valueExpr string
	if metric == "points" {
		valueExpr = "COALESCE(SUM(points) FILTER (WHERE done), 0)"
	} else {
		valueExpr = "COUNT(*) FILTER (WHERE done)"
	}

	var groupExpr, labelExpr, interval string
	switch view {
	case "week":
		groupExpr = "DATE_TRUNC('week', date)"
		labelExpr = "TO_CHAR(DATE_TRUNC('week', date), 'Mon DD')"
		interval = "12 weeks"
	case "month":
		groupExpr = "DATE_TRUNC('month', date)"
		labelExpr = "TO_CHAR(DATE_TRUNC('month', date), 'Mon YYYY')"
		interval = "12 months"
	default: // day
		groupExpr = "date"
		labelExpr = "TO_CHAR(date, 'Mon DD')"
		interval = "30 days"
	}

	// view and metric are validated to a fixed allowed set before this call — no injection risk.
	q := fmt.Sprintf(`
		SELECT %s AS label, %s AS value
		FROM tasks
		WHERE date >= (CURRENT_DATE - INTERVAL '%s')
		GROUP BY %s
		ORDER BY %s`,
		labelExpr, valueExpr, interval, groupExpr, groupExpr)

	points := make([]ChartPoint, 0)
	if err := r.db.SelectContext(ctx, &points, q); err != nil {
		return nil, fmt.Errorf("chartData view=%s metric=%s: %w", view, metric, err)
	}
	return points, nil
}
