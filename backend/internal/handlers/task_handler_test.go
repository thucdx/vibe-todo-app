package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/models"
)

func TestTaskHandler_List(t *testing.T) {
	task := models.Task{ID: uuid.New(), Title: "Read book"}

	tests := []struct {
		name       string
		query      string
		tasks      []models.Task
		listErr    error
		wantStatus int
	}{
		{
			name:       "returns tasks for date",
			query:      "?date=2026-02-28",
			tasks:      []models.Task{task},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing date param",
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service error",
			query:      "?date=2026-02-28",
			listErr:    fmt.Errorf("db error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubTaskSvc{tasks: tc.tasks, listErr: tc.listErr}
			h := NewTaskHandler(svc)
			r := newRouter(http.MethodGet, "/api/v1/tasks", h.List)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks"+tc.query, nil)
			w := doRequest(r, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestTaskHandler_Create(t *testing.T) {
	task := &models.Task{ID: uuid.New(), Title: "Write tests"}

	tests := []struct {
		name       string
		body       any
		task       *models.Task
		createErr  error
		wantStatus int
	}{
		{
			name:       "creates task",
			body:       map[string]any{"title": "Write tests", "date": "2026-02-28"},
			task:       task,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing required title",
			body:       map[string]any{"date": "2026-02-28"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service error",
			body:       map[string]any{"title": "Write tests", "date": "2026-02-28"},
			createErr:  fmt.Errorf("db error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubTaskSvc{task: tc.task, createErr: tc.createErr}
			h := NewTaskHandler(svc)
			r := newRouter(http.MethodPost, "/api/v1/tasks", h.Create)

			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := doRequest(r, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d (body: %s)", w.Code, tc.wantStatus, w.Body.String())
			}
		})
	}
}

func TestTaskHandler_Update(t *testing.T) {
	id := uuid.New()
	task := &models.Task{ID: id, Title: "Updated"}

	tests := []struct {
		name       string
		paramID    string
		body       any
		task       *models.Task
		updateErr  error
		wantStatus int
	}{
		{
			name:       "updates task",
			paramID:    id.String(),
			body:       map[string]any{"title": "Updated", "date": "2026-02-28"},
			task:       task,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			paramID:    "not-a-uuid",
			body:       map[string]any{"title": "Updated", "date": "2026-02-28"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service error",
			paramID:    id.String(),
			body:       map[string]any{"title": "Updated", "date": "2026-02-28"},
			updateErr:  fmt.Errorf("db error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubTaskSvc{task: tc.task, updateErr: tc.updateErr}
			h := NewTaskHandler(svc)
			r := newRouter(http.MethodPut, "/api/v1/tasks/:id", h.Update)

			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/"+tc.paramID, bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := doRequest(r, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestTaskHandler_ToggleDone(t *testing.T) {
	id := uuid.New()
	task := &models.Task{ID: id, Done: true}

	tests := []struct {
		name       string
		paramID    string
		task       *models.Task
		toggleErr  error
		wantStatus int
	}{
		{
			name:       "toggles done",
			paramID:    id.String(),
			task:       task,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			paramID:    "bad",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "service error",
			paramID:    id.String(),
			toggleErr:  fmt.Errorf("db error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubTaskSvc{task: tc.task, toggleErr: tc.toggleErr}
			h := NewTaskHandler(svc)
			r := newRouter(http.MethodPatch, "/api/v1/tasks/:id/done", h.ToggleDone)

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+tc.paramID+"/done", nil)
			w := doRequest(r, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

func TestTaskHandler_Delete(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name       string
		paramID    string
		deleteErr  error
		wantStatus int
	}{
		{
			name:       "deletes task",
			paramID:    id.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid uuid",
			paramID:    "bad",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found",
			paramID:    id.String(),
			deleteErr:  apperrors.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "service error",
			paramID:    id.String(),
			deleteErr:  fmt.Errorf("db error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubTaskSvc{deleteErr: tc.deleteErr}
			h := NewTaskHandler(svc)
			r := newRouter(http.MethodDelete, "/api/v1/tasks/:id", h.Delete)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+tc.paramID, nil)
			w := doRequest(r, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}
