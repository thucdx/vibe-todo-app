package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thucdx/todovibe/internal/repository"
)

func TestStatsHandler_Chart(t *testing.T) {
	points := []repository.ChartPoint{
		{Label: "Mon", Value: 3},
		{Label: "Tue", Value: 7},
	}

	tests := []struct {
		name       string
		query      string
		points     []repository.ChartPoint
		err        error
		wantStatus int
		wantLen    int
	}{
		{
			name:       "returns chart data with defaults",
			query:      "",
			points:     points,
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:       "explicit view and metric",
			query:      "?view=week&metric=points",
			points:     points,
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:       "bad request from service (invalid view/metric)",
			query:      "?view=bad&metric=count",
			err:        errBadReq,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service error",
			query:      "?view=day&metric=count",
			err:        fmt.Errorf("db failure"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubStatsSvc{points: tc.points, err: tc.err}
			h := NewStatsHandler(svc)
			r := newRouter(http.MethodGet, "/api/v1/stats", h.Chart)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/stats"+tc.query, nil)
			w := doRequest(r, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d (body: %s)", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantLen > 0 {
				var result []repository.ChartPoint
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if len(result) != tc.wantLen {
					t.Errorf("got %d points, want %d", len(result), tc.wantLen)
				}
			}
		})
	}
}
