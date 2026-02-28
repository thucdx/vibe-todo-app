package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// stubStore implements the sessionStore interface for testing.
type stubStore struct {
	sess *models.Session
	err  error
}

func (s *stubStore) Validate(_ context.Context, _ uuid.UUID) (*models.Session, error) {
	return s.sess, s.err
}

func newProtectedRouter(store sessionStore) *gin.Engine {
	r := gin.New()
	r.GET("/protected", Auth(store), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func doReq(r *gin.Engine, cookie string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: cookie})
	}
	r.ServeHTTP(w, req)
	return w
}

func TestAuth_NoCookie(t *testing.T) {
	r := newProtectedRouter(&stubStore{err: apperrors.ErrUnauthorized})
	w := doReq(r, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestAuth_InvalidCookieUUID(t *testing.T) {
	r := newProtectedRouter(&stubStore{err: apperrors.ErrUnauthorized})
	w := doReq(r, "not-a-valid-uuid")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestAuth_ValidateReturnsError(t *testing.T) {
	r := newProtectedRouter(&stubStore{err: apperrors.ErrUnauthorized})
	w := doReq(r, uuid.New().String())
	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want 401", w.Code)
	}
}

func TestAuth_ValidSession(t *testing.T) {
	token := uuid.New()
	sess := &models.Session{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	r := newProtectedRouter(&stubStore{sess: sess})
	w := doReq(r, token.String())
	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}
}

func TestAuth_SessionStoredInContext(t *testing.T) {
	token := uuid.New()
	sess := &models.Session{Token: token, ExpiresAt: time.Now().Add(time.Hour)}
	store := &stubStore{sess: sess}

	r := gin.New()
	var captured any
	r.GET("/check", Auth(store), func(c *gin.Context) {
		captured = c.MustGet("session")
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/check", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token.String()})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", w.Code)
	}
	got, ok := captured.(*models.Session)
	if !ok {
		t.Fatalf("context session is %T, want *models.Session", captured)
	}
	if got.Token != token {
		t.Errorf("stored token %v, want %v", got.Token, token)
	}
}
