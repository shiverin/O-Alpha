package main

import (
	"context"
	"fmt"
	"log"
	"os"

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
	
	// Check schema_migrations table
	var version uint64
	var dirty bool
	err = db.QueryRow(context.Background(), "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version, &dirty)
	if err != nil {
		log.Fatalf("Failed to read schema_migrations: %v", err)
	}
	fmt.Printf("Current migration version: %d, dirty: %v\n", version, dirty)
	
	// List all applied migrations
	rows, err := db.Query(context.Background(), "SELECT version, dirty FROM schema_migrations ORDER BY version")
	if err != nil {
		log.Fatalf("Failed to list migrations: %v", err)
	}
	defer rows.Close()
	
	fmt.Println("Applied migrations:")
	for rows.Next() {
		var v uint64
		var d bool
		if err := rows.Scan(&v, &d); err != nil {
			log.Fatalf("Failed to scan migration row: %v", err)
		}
		fmt.Printf("  Version: %d, Dirty: %v\n", v, d)
	}
	
	// Check if our migration files exist
	fmt.Println("\nChecking migration files:")
	files := []string{"000001_init.up.sql", "000001_init.down.sql", "000002_users.up.sql", "000002_users.down.sql"}
	for _, f := range files {
		if _, err := os.Stat("./migrations/" + f); err == nil {
			fmt.Printf("  ✅ %s exists\n", f)
		} else {
			fmt.Printf("  ❌ %s missing\n", f)
		}
	}
}
