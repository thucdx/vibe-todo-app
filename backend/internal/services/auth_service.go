package services

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
	apperrors "github.com/thucdx/todovibe/internal/errors"
	"github.com/thucdx/todovibe/internal/models"
	"github.com/thucdx/todovibe/internal/repository"
)

const sessionTTL = 7 * 24 * time.Hour

// AuthService handles PIN management and session lifecycle.
type AuthService struct {
	settings *repository.SettingsRepo
	sessions *repository.SessionRepo
}

func NewAuthService(settings *repository.SettingsRepo, sessions *repository.SessionRepo) *AuthService {
	return &AuthService{settings: settings, sessions: sessions}
}

// IsPINConfigured returns true if a PIN hash has been stored.
func (s *AuthService) IsPINConfigured(ctx context.Context) bool {
	_, err := s.settings.Get(ctx, models.PinHashKey)
	return err == nil
}

// SetupPIN hashes and stores a new PIN on first use.
func (s *AuthService) SetupPIN(ctx context.Context, pin string) error {
	if s.IsPINConfigured(ctx) {
		return apperrors.ErrPINAlreadySet
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.settings.Set(ctx, models.PinHashKey, string(hash))
}

// ResetPIN overwrites the stored PIN hash and invalidates all existing sessions.
func (s *AuthService) ResetPIN(ctx context.Context, newPIN string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPIN), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.settings.Set(ctx, models.PinHashKey, string(hash)); err != nil {
		return err
	}
	return s.sessions.DeleteAll(ctx)
}

// Login verifies the PIN and, on success, creates and returns a new session token.
func (s *AuthService) Login(ctx context.Context, pin string) (string, error) {
	hash, err := s.settings.Get(ctx, models.PinHashKey)
	if err != nil {
		return "", apperrors.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin)); err != nil {
		return "", apperrors.ErrInvalidPIN
	}
	sess, err := s.sessions.Create(ctx, sessionTTL)
	if err != nil {
		return "", err
	}
	return sess.Token.String(), nil
}

// Logout deletes the session identified by the given token string.
func (s *AuthService) Logout(ctx context.Context, tokenStr string) {
	// Best-effort: parse and delete — ignore errors.
	from, err := parseUUID(tokenStr)
	if err != nil {
		return
	}
	_ = s.sessions.Delete(ctx, from)
}
