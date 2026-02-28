package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/models"
)

func TestAuthHandler_Status(t *testing.T) {
	validToken := uuid.New()
	validSession := &models.Session{Token: validToken, ExpiresAt: time.Now().Add(time.Hour)}

	tests := []struct {
		name       string
		svc        *stubAuthSvc
		session    *stubSession
		cookie     string
		wantStatus int
		wantBody   map[string]any
	}{
		{
			name:       "not configured, no cookie",
			svc:        &stubAuthSvc{configured: false},
			session:    &stubSession{err: apperrors.ErrUnauthorized},
			cookie:     "",
			wantStatus: http.StatusOK,
			wantBody:   map[string]any{"configured": false, "authenticated": false},
		},
		{
			name:       "configured, no cookie",
			svc:        &stubAuthSvc{configured: true},
			session:    &stubSession{err: apperrors.ErrUnauthorized},
			cookie:     "",
			wantStatus: http.StatusOK,
			wantBody:   map[string]any{"configured": true, "authenticated": false},
		},
		{
			name:       "configured, valid cookie",
			svc:        &stubAuthSvc{configured: true},
			session:    &stubSession{sess: validSession},
			cookie:     validToken.String(),
			wantStatus: http.StatusOK,
			wantBody:   map[string]any{"configured": true, "authenticated": true},
		},
		{
			name:       "configured, invalid cookie uuid",
			svc:        &stubAuthSvc{configured: true},
			session:    &stubSession{err: apperrors.ErrUnauthorized},
			cookie:     "not-a-uuid",
			wantStatus: http.StatusOK,
			wantBody:   map[string]any{"configured": true, "authenticated": false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewAuthHandler(tc.svc, tc.session)
			r := newRouter(http.MethodGet, "/api/v1/auth/status", h.Status)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/status", nil)
			if tc.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "session_token", Value: tc.cookie})
			}
			w := doRequest(r, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			for k, v := range tc.wantBody {
				if body[k] != v {
					t.Errorf("body[%q]: got %v, want %v", k, body[k], v)
				}
			}
		})
	}
}

func TestAuthHandler_Setup(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		setupErr   error
		wantStatus int
	}{
		{
			name:       "success",
			body:       map[string]string{"pin": "1234"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "pin already set",
			body:       map[string]string{"pin": "1234"},
			setupErr:   apperrors.ErrPINAlreadySet,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "missing pin field",
			body:       map[string]string{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "pin too short (< 4 chars)",
			body:       map[string]string{"pin": "123"},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubAuthSvc{setupErr: tc.setupErr}
			h := NewAuthHandler(svc, &stubSession{})
			r := newRouter(http.MethodPost, "/api/v1/auth/setup", h.Setup)

			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/setup", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := doRequest(r, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d (body: %s)", w.Code, tc.wantStatus, w.Body.String())
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		loginToken string
		loginErr   error
		wantStatus int
		wantCookie bool
	}{
		{
			name:       "success sets cookie",
			body:       map[string]string{"pin": "1234"},
			loginToken: uuid.New().String(),
			wantStatus: http.StatusOK,
			wantCookie: true,
		},
		{
			name:       "invalid pin",
			body:       map[string]string{"pin": "wrong"},
			loginErr:   apperrors.ErrInvalidPIN,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing body",
			body:       map[string]string{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := &stubAuthSvc{loginToken: tc.loginToken, loginErr: tc.loginErr}
			h := NewAuthHandler(svc, &stubSession{})
			r := newRouter(http.MethodPost, "/api/v1/auth/login", h.Login)

			b, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := doRequest(r, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
			hasCookie := w.Result().Header.Get("Set-Cookie") != ""
			if tc.wantCookie && !hasCookie {
				t.Error("expected Set-Cookie header, got none")
			}
			if !tc.wantCookie && hasCookie {
				t.Error("unexpected Set-Cookie header")
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	h := NewAuthHandler(&stubAuthSvc{}, &stubSession{})
	r := newRouter(http.MethodPost, "/api/v1/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := doRequest(r, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	// Cookie should be cleared (MaxAge = -1).
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "session_token" && c.MaxAge >= 0 {
			t.Errorf("expected cookie to be cleared, got MaxAge=%d", c.MaxAge)
		}
	}
}
