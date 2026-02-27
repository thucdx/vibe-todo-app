package services_test

import (
	"context"
	"testing"

	"github.com/thucdx/todovibe/internal/config"
	"github.com/thucdx/todovibe/internal/db"
	"github.com/thucdx/todovibe/internal/repository"
	"github.com/thucdx/todovibe/internal/services"
)

// testDSN returns the DSN from environment for integration tests.
// Tests are skipped when no database is available.
func testDSN() string {
	cfg := config.Config{DBDSN: "postgres://todovibe:changeme@localhost:5432/todovibe?sslmode=disable"}
	return cfg.DBDSN
}

func TestAuthService_SetupAndLogin(t *testing.T) {
	database, err := db.Connect(testDSN())
	if err != nil {
		t.Skipf("no test database available: %v", err)
	}
	defer database.Close()

	settingsRepo := repository.NewSettingsRepo(database)
	sessionRepo  := repository.NewSessionRepo(database)
	svc          := services.NewAuthService(settingsRepo, sessionRepo)
	ctx          := context.Background()

	// Clean state
	_, _ = database.ExecContext(ctx, "DELETE FROM sessions")
	_, _ = database.ExecContext(ctx, "DELETE FROM settings WHERE key = 'pin_hash'")

	// Should not be configured yet
	if svc.IsPINConfigured(ctx) {
		t.Fatal("expected PIN not configured before setup")
	}

	// Setup PIN
	if err := svc.SetupPIN(ctx, "1234"); err != nil {
		t.Fatalf("SetupPIN failed: %v", err)
	}

	if !svc.IsPINConfigured(ctx) {
		t.Fatal("expected PIN to be configured after setup")
	}

	// Second setup should fail with ErrPINAlreadySet
	if err := svc.SetupPIN(ctx, "5678"); err == nil {
		t.Fatal("expected error on second SetupPIN call")
	}

	// Login with correct PIN
	token, err := svc.Login(ctx, "1234")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty session token")
	}

	// Login with wrong PIN
	_, err = svc.Login(ctx, "wrong")
	if err == nil {
		t.Fatal("expected error for wrong PIN")
	}
}

func TestAuthService_Reset(t *testing.T) {
	database, err := db.Connect(testDSN())
	if err != nil {
		t.Skipf("no test database available: %v", err)
	}
	defer database.Close()

	settingsRepo := repository.NewSettingsRepo(database)
	sessionRepo  := repository.NewSessionRepo(database)
	svc          := services.NewAuthService(settingsRepo, sessionRepo)
	ctx          := context.Background()

	_, _ = database.ExecContext(ctx, "DELETE FROM sessions")
	_, _ = database.ExecContext(ctx, "DELETE FROM settings WHERE key = 'pin_hash'")

	_ = svc.SetupPIN(ctx, "1234")
	_ = svc.ResetPIN(ctx, "9999")

	// Old PIN should no longer work
	_, err = svc.Login(ctx, "1234")
	if err == nil {
		t.Fatal("expected error logging in with old PIN after reset")
	}

	// New PIN should work
	_, err = svc.Login(ctx, "9999")
	if err != nil {
		t.Fatalf("expected login to succeed with new PIN: %v", err)
	}
}
