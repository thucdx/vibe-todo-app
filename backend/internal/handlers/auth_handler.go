package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/middleware"
	"github.com/thucdx/todovibe/internal/repository"
	"github.com/thucdx/todovibe/internal/services"
)

// AuthHandler handles authentication-related endpoints.
type AuthHandler struct {
	svc      *services.AuthService
	sessions *repository.SessionRepo
}

func NewAuthHandler(svc *services.AuthService, sessions *repository.SessionRepo) *AuthHandler {
	return &AuthHandler{svc: svc, sessions: sessions}
}

// Status godoc
// GET /api/v1/auth/status
// This endpoint sits outside the Auth middleware group, so it manually validates
// the session cookie to return the correct `authenticated` state.
func (h *AuthHandler) Status(c *gin.Context) {
	configured := h.svc.IsPINConfigured(c.Request.Context())

	authenticated := false
	if cookie, err := c.Cookie(middleware.SessionCookieName); err == nil {
		if token, err := uuid.Parse(cookie); err == nil {
			if _, err := h.sessions.Validate(c.Request.Context(), token); err == nil {
				authenticated = true
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"configured":    configured,
		"authenticated": authenticated,
	})
}

// Setup godoc
// POST /api/v1/auth/setup
func (h *AuthHandler) Setup(c *gin.Context) {
	var body struct {
		PIN string `json:"pin" binding:"required,min=4,max=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SetupPIN(c.Request.Context(), body.PIN); err != nil {
		if err == apperrors.ErrPINAlreadySet {
			c.JSON(http.StatusConflict, gin.H{"error": "PIN already configured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set PIN"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "PIN created"})
}

// Login godoc
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var body struct {
		PIN string `json:"pin" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.svc.Login(c.Request.Context(), body.PIN)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid PIN"})
		return
	}
	maxAge := int(7 * 24 * time.Hour / time.Second)
	c.SetCookie(middleware.SessionCookieName, token, maxAge, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged in"})
}

// Logout godoc
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	cookie, _ := c.Cookie(middleware.SessionCookieName)
	h.svc.Logout(c.Request.Context(), cookie)
	c.SetCookie(middleware.SessionCookieName, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
