package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/models"
)

// taskServicer is the subset of services.TaskService used by TaskHandler.
type taskServicer interface {
	List(ctx context.Context, date string) ([]models.Task, error)
	Create(ctx context.Context, in models.CreateTaskInput) (*models.Task, error)
	Update(ctx context.Context, id uuid.UUID, in models.UpdateTaskInput) (*models.Task, error)
	ToggleDone(ctx context.Context, id uuid.UUID) (*models.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// TaskHandler handles task CRUD endpoints.
type TaskHandler struct {
	svc taskServicer
}

func NewTaskHandler(svc taskServicer) *TaskHandler {
	return &TaskHandler{svc: svc}
}

// List godoc
// GET /api/v1/tasks?date=YYYY-MM-DD
func (h *TaskHandler) List(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date query param is required"})
		return
	}
	tasks, err := h.svc.List(c.Request.Context(), date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// Create godoc
// POST /api/v1/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	var in models.CreateTaskInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

// Update godoc
// PUT /api/v1/tasks/:id
func (h *TaskHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	var in models.UpdateTaskInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

// ToggleDone godoc
// PATCH /api/v1/tasks/:id/done
func (h *TaskHandler) ToggleDone(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	task, err := h.svc.ToggleDone(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

// Delete godoc
// DELETE /api/v1/tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if err == apperrors.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
