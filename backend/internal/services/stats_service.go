package services

import (
	"context"

	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/repository"
)

var validViews   = map[string]bool{"day": true, "week": true, "month": true}
var validMetrics = map[string]bool{"count": true, "points": true}

// StatsService provides aggregated productivity chart data.
type StatsService struct {
	repo *repository.TaskRepo
}

func NewStatsService(repo *repository.TaskRepo) *StatsService {
	return &StatsService{repo: repo}
}

// ChartData returns aggregated data points for the given view and metric.
func (s *StatsService) ChartData(ctx context.Context, view, metric string) ([]repository.ChartPoint, error) {
	if !validViews[view] {
		return nil, apperrors.ErrBadRequest
	}
	if !validMetrics[metric] {
		return nil, apperrors.ErrBadRequest
	}
	return s.repo.ChartData(ctx, view, metric)
}
