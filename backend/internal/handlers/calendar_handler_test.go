package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thucdx/todovibe/internal/repository"
)

func TestCalendarHandler_Summary(t *testing.T) {
	summary := []repository.DaySummary{
		{Date: "2026-02-01", Done: 2, Total: 5},
		{Date: "2026-02-15", Done: 1, Total: 1},
	}

	tests := []struct {
		name       string
		query      string
		summary    []repository.DaySummary
		err        error
		wantStatus int
		wantLen    int
	}{
		{
			name:       "returns summary with explicit year+month",
			query:      "?year=2026&month=2",
			summary:    summary,
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:       "uses current year/month when params absent",
			query:      "",
			summary:    summary,
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:       "invalid year param",
			query:      "?year=not-a-year&month=2",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid month param",
			query:      "?year=2026&month=abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service error",
			query:      "?year=2026&month=2",
			err:        fmt.Errorf("db failure"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubCalSvc{summary: tc.summary, err: tc.err}
			h := NewCalendarHandler(svc)
			r := newRouter(http.MethodGet, "/api/v1/calendar", h.Summary)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/calendar"+tc.query, nil)
			w := doRequest(r, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d (body: %s)", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantLen > 0 {
				var result []repository.DaySummary
				if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if len(result) != tc.wantLen {
					t.Errorf("got %d entries, want %d", len(result), tc.wantLen)
				}
			}
		})
	}
}
