package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/models"
)

// SessionRepo manages session records in the database.
type SessionRepo struct {
	db *sqlx.DB
}

func NewSessionRepo(db *sqlx.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

// Create inserts a new session with the given TTL and returns it.
func (r *SessionRepo) Create(ctx context.Context, ttl time.Duration) (*models.Session, error) {
	now := time.Now()
	s := &models.Session{
		Token:      uuid.New(),
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
		LastSeenAt: now,
	}
	const q = `
		INSERT INTO sessions (token, created_at, expires_at, last_seen_at)
		VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, q, s.Token, s.CreatedAt, s.ExpiresAt, s.LastSeenAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Validate looks up a session by token, checks expiry, and slides the expiry window forward.
func (r *SessionRepo) Validate(ctx context.Context, token uuid.UUID) (*models.Session, error) {
	var s models.Session
	const q = `SELECT * FROM sessions WHERE token = $1 AND expires_at > NOW()`
	if err := r.db.GetContext(ctx, &s, q, token); err != nil {
		return nil, apperrors.ErrUnauthorized
	}
	// Slide the 7-day expiry window
	const upd = `UPDATE sessions SET last_seen_at = NOW(), expires_at = NOW() + INTERVAL '7 days' WHERE token = $1`
	_, _ = r.db.ExecContext(ctx, upd, token)
	return &s, nil
}

// Delete removes a session by token.
func (r *SessionRepo) Delete(ctx context.Context, token uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}

// DeleteAll removes all sessions (used when resetting PIN).
func (r *SessionRepo) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions`)
	return err
}
