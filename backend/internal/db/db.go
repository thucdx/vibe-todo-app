package db

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Connect opens a sqlx connection pool to PostgreSQL.
func Connect(dsn string) (*sqlx.DB, error) {
	database, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(10)
	return database, nil
}

// Migrate runs all pending up-migrations from the migrations/ directory.
// ErrNoChange is treated as success.
func Migrate(database *sqlx.DB) error {
	driver, err := postgres.WithInstance(database.DB, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
