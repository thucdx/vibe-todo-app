package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/thucdx/todovibe/internal/models"
	"github.com/thucdx/todovibe/internal/repository"
)

// TaskService is a thin delegation layer between handlers and the task repository.
type TaskService struct {
	repo *repository.TaskRepo
}

func NewTaskService(repo *repository.TaskRepo) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) List(ctx context.Context, date string) ([]models.Task, error) {
	return s.repo.ListByDate(ctx, date)
}

func (s *TaskService) Create(ctx context.Context, in models.CreateTaskInput) (*models.Task, error) {
	return s.repo.Create(ctx, in)
}

func (s *TaskService) Update(ctx context.Context, id uuid.UUID, in models.UpdateTaskInput) (*models.Task, error) {
	return s.repo.Update(ctx, id, in)
}

func (s *TaskService) ToggleDone(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	return s.repo.ToggleDone(ctx, id)
}

func (s *TaskService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
