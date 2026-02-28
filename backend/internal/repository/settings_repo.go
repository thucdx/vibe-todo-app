package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// SettingsRepo handles key-value app settings in the database.
type SettingsRepo struct {
	db *sqlx.DB
}

func NewSettingsRepo(db *sqlx.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

// Get retrieves the value for a given settings key.
func (r *SettingsRepo) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.GetContext(ctx, &value, `SELECT value FROM settings WHERE key = $1`, key)
	return value, err
}

// Set upserts a key-value pair in settings.
func (r *SettingsRepo) Set(ctx context.Context, key, value string) error {
	const q = `
		INSERT INTO settings (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`
	_, err := r.db.ExecContext(ctx, q, key, value)
	return err
}

// Delete removes a single settings key row.
func (r *SettingsRepo) Delete(ctx context.Context, key string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM settings WHERE key = $1`, key)
	return err
}
