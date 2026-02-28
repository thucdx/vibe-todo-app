package services

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/repository"
)

// stubChartRepo satisfies the chartRepo interface for testing.
type stubChartRepo struct {
	points []repository.ChartPoint
	err    error
}

func (s *stubChartRepo) ChartData(_ context.Context, _, _ string) ([]repository.ChartPoint, error) {
	return s.points, s.err
}

func newStatsService(stub *stubChartRepo) *StatsService {
	return &StatsService{repo: stub}
}

func TestStatsService_ChartData_InvalidView(t *testing.T) {
	svc := newStatsService(&stubChartRepo{})
	_, err := svc.ChartData(context.Background(), "bad-view", "count")
	if err != apperrors.ErrBadRequest {
		t.Errorf("expected ErrBadRequest for invalid view, got %v", err)
	}
}

func TestStatsService_ChartData_InvalidMetric(t *testing.T) {
	svc := newStatsService(&stubChartRepo{})
	_, err := svc.ChartData(context.Background(), "day", "bad-metric")
	if err != apperrors.ErrBadRequest {
		t.Errorf("expected ErrBadRequest for invalid metric, got %v", err)
	}
}

func TestStatsService_ChartData_ValidCases(t *testing.T) {
	views := []string{"day", "week", "month"}
	metrics := []string{"count", "points"}

	expectedPoints := []repository.ChartPoint{{Label: "Mon", Value: 5}}
	stub := &stubChartRepo{points: expectedPoints}
	svc := newStatsService(stub)

	for _, view := range views {
		for _, metric := range metrics {
			t.Run(view+"/"+metric, func(t *testing.T) {
				got, err := svc.ChartData(context.Background(), view, metric)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(got) != len(expectedPoints) {
					t.Errorf("got %d points, want %d", len(got), len(expectedPoints))
				}
			})
		}
	}
}

func TestStatsService_ChartData_RepoError(t *testing.T) {
	repoErr := errors.New("db failure")
	svc := newStatsService(&stubChartRepo{err: repoErr})
	_, err := svc.ChartData(context.Background(), "day", "count")
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error to propagate, got %v", err)
	}
}
