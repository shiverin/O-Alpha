package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/lib/pq"
)

// ResetMigrations drops the schema_migrations table and reapplies all migrations.
// Use this when the database is in a dirty state and you want to start fresh.
func ResetMigrations(databaseURL, migrationsPath string) error {
	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	// Drop the schema_migrations table to reset migration state
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if _, err := db.Exec("DROP TABLE IF EXISTS schema_migrations"); err != nil {
		return fmt.Errorf("drop schema_migrations: %w", err)
	}

	// Apply all migrations from scratch
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}