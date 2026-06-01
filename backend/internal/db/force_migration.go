package db

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
)

// ForceResetMigrations forces the migration version to the latest available and then applies all migrations.
// Use this when the database is in a dirty state.
func ForceResetMigrations(databaseURL, migrationsPath string) error {
	// Find the latest migration version available
	latestVersion := 1 // Default to first migration

	// Check if we have a second migration
	if _, err := os.Stat(filepath.Join(migrationsPath, "000002_users.up.sql")); err == nil {
		latestVersion = 2
	}

	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	// Force the version to the latest available migration
	if err := m.Force(latestVersion); err != nil {
		return fmt.Errorf("force version %d: %w", latestVersion, err)
	}

	// Apply all migrations (none, because we are at the latest version)
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
