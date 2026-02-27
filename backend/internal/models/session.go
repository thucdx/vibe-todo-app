package models

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an authenticated user session stored in the database.
type Session struct {
	Token       uuid.UUID `db:"token"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
	LastSeenAt  time.Time `db:"last_seen_at"`
}
