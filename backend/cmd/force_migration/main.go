package main

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	migrationsPath := "file://../migrations"

	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		log.Fatalf("create migrator: %v", err)
	}
	defer m.Close()

	// Force the migration version to 0 to clean the dirty state
	if err := m.Force(0); err != nil {
		log.Fatalf("force version: %v", err)
	}
	log.Println("Forced migration version to 0")

	// Now run the migrations
	if err := m.Up(); err != nil {
		log.Fatalf("migrate up: %v", err)
	}
	log.Println("Migrations applied")
}