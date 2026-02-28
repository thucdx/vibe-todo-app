package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/models"
	"github.com/thucdx/todovibe/internal/repository"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newRouter builds a bare gin engine with the provided handler function registered.
func newRouter(method, path string, h gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Handle(method, path, h)
	return r
}

// doRequest performs a test HTTP request against a gin router.
func doRequest(r *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ─── auth stubs ──────────────────────────────────────────────────────────────

type stubAuthSvc struct {
	configured bool
	setupErr   error
	loginToken string
	loginErr   error
}

func (s *stubAuthSvc) IsPINConfigured(_ context.Context) bool        { return s.configured }
func (s *stubAuthSvc) SetupPIN(_ context.Context, _ string) error    { return s.setupErr }
func (s *stubAuthSvc) Login(_ context.Context, _ string) (string, error) {
	return s.loginToken, s.loginErr
}
func (s *stubAuthSvc) Logout(_ context.Context, _ string) {}

type stubSession struct {
	sess *models.Session
	err  error
}

func (s *stubSession) Validate(_ context.Context, _ uuid.UUID) (*models.Session, error) {
	return s.sess, s.err
}

// ─── task stubs ──────────────────────────────────────────────────────────────

type stubTaskSvc struct {
	tasks  []models.Task
	task   *models.Task
	listErr  error
	createErr error
	updateErr error
	toggleErr error
	deleteErr error
}

func (s *stubTaskSvc) List(_ context.Context, _ string) ([]models.Task, error) {
	return s.tasks, s.listErr
}
func (s *stubTaskSvc) Create(_ context.Context, _ models.CreateTaskInput) (*models.Task, error) {
	return s.task, s.createErr
}
func (s *stubTaskSvc) Update(_ context.Context, _ uuid.UUID, _ models.UpdateTaskInput) (*models.Task, error) {
	return s.task, s.updateErr
}
func (s *stubTaskSvc) ToggleDone(_ context.Context, _ uuid.UUID) (*models.Task, error) {
	return s.task, s.toggleErr
}
func (s *stubTaskSvc) Delete(_ context.Context, _ uuid.UUID) error {
	return s.deleteErr
}

// ─── calendar stubs ──────────────────────────────────────────────────────────

type stubCalSvc struct {
	summary []repository.DaySummary
	err     error
}

func (s *stubCalSvc) MonthlySummary(_ context.Context, _, _ int) ([]repository.DaySummary, error) {
	return s.summary, s.err
}

// ─── stats stubs ─────────────────────────────────────────────────────────────

type stubStatsSvc struct {
	points []repository.ChartPoint
	err    error
}

func (s *stubStatsSvc) ChartData(_ context.Context, _, _ string) ([]repository.ChartPoint, error) {
	return s.points, s.err
}

// errBadReq is a convenience alias used by stats handler tests.
var errBadReq = apperrors.ErrBadRequest
