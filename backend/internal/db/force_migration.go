package db

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
)

// ForceResetMigrations forces the migration version to the latest available and then applies all migrations.
// Use this when the database is in a dirty state.
func ForceResetMigrations(databaseURL, migrationsPath string) error {
	latestVersion, err := latestMigrationVersion(migrationsPath)
	if err != nil {
		return err
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

func latestMigrationVersion(migrationsPath string) (int, error) {
	path := strings.TrimPrefix(migrationsPath, "file://")
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, fmt.Errorf("read migrations path: %w", err)
	}
	var latest int
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		prefix := strings.SplitN(entry.Name(), "_", 2)[0]
		version, err := strconv.Atoi(prefix)
		if err != nil {
			continue
		}
		if version > latest {
			latest = version
		}
	}
	if latest == 0 {
		return 0, fmt.Errorf("no up migrations found in %s", filepath.Clean(path))
	}
	return latest, nil
}
