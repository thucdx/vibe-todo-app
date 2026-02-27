package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thucdx/todovibe/internal/repository"
)

const SessionCookieName = "session_token"

// Auth validates the session cookie on every protected request.
// On success it stores the session in the Gin context under the key "session".
func Auth(sessions *repository.SessionRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(SessionCookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		token, err := uuid.Parse(cookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		sess, err := sessions.Validate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("session", sess)
		c.Next()
	}
}
