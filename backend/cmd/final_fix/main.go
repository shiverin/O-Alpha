package main

import (
	"context"
	"fmt"
	"log"

	"github.com/oalpha/internal/db"
)

func main() {
	databaseURL := "postgresql://postgres.aojfydtbdvjumcsgywxa:shiverin123%26Z@aws-1-ap-southeast-2.pooler.supabase.com:5432/postgres"
	fmt.Printf("Connecting to: %s\n", databaseURL)
	
	db, err := db.Open(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()
	
	fmt.Println("✅ Connected to database successfully")
	
	// Ensure schema_migrations table exists with correct structure
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT NOT NULL,
			dirty BOOLEAN NOT NULL DEFAULT FALSE,
			PRIMARY KEY (version)
		);
	`
	
	_, err = db.Exec(context.Background(), createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}
	fmt.Println("✅ Ensured schema_migrations table exists")
	
	// Clear any existing incorrect entries
	_, err = db.Exec(context.Background(), "DELETE FROM schema_migrations")
	if err != nil {
		log.Fatalf("Failed to clear schema_migrations: %v", err)
	}
	fmt.Println("✅ Cleared existing schema_migrations entries")
	
	// Insert the correct migration versions based on what we actually have applied
	// Since we manually applied migrations 1 and 2 earlier, we should mark them as applied
	_, err = db.Exec(context.Background(), "INSERT INTO schema_migrations (version, dirty) VALUES (1, false), (2, false)")
	if err != nil {
		log.Fatalf("Failed to insert migration versions: %v", err)
	}
	fmt.Println("✅ Inserted correct migration versions (1 and 2)")
	
	// Verify the fix
	rows, err := db.Query(context.Background(), "SELECT version, dirty FROM schema_migrations ORDER BY version")
	if err != nil {
		log.Fatalf("Failed to read schema_migrations: %v", err)
	}
	defer rows.Close()
	
	fmt.Println("\nFinal schema_migrations state:")
	for rows.Next() {
		var version uint64
		var dirty bool
		if err := rows.Scan(&version, &dirty); err != nil {
			log.Fatalf("Failed to scan migration row: %v", err)
		}
		fmt.Printf("  Version: %d, Dirty: %v\n", version, dirty)
	}
	
	fmt.Println("\n🎉 Migration system is now ready!")
}
