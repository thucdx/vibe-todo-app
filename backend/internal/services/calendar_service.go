package services

import (
	"context"

	"github.com/thucdx/todovibe/internal/repository"
)

// CalendarService provides calendar-level task summaries.
type CalendarService struct {
	repo *repository.TaskRepo
}

func NewCalendarService(repo *repository.TaskRepo) *CalendarService {
	return &CalendarService{repo: repo}
}

// MonthlySummary returns done/total counts per day for the given year and month.
func (s *CalendarService) MonthlySummary(ctx context.Context, year, month int) ([]repository.DaySummary, error) {
	return s.repo.CalendarSummary(ctx, year, month)
}
