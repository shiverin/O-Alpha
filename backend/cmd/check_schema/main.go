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
	
	// Check schema_migrations table structure
	rows, err := db.Query(context.Background(), "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'schema_migrations'")
	if err != nil {
		log.Fatalf("Failed to get schema_migrations structure: %v", err)
	}
	defer rows.Close()
	
	fmt.Println("schema_migrations table structure:")
	for rows.Next() {
		var columnName, dataType string
		if err := rows.Scan(&columnName, &dataType); err != nil {
			log.Fatalf("Failed to scan column info: %v", err)
		}
		fmt.Printf("  %s: %s\n", columnName, dataType)
	}
	
	// Check current data
	fmt.Println("\nschema_migrations table data:")
	rows, err = db.Query(context.Background(), "SELECT version, dirty FROM schema_migrations ORDER BY version")
	if err != nil {
		log.Fatalf("Failed to read schema_migrations data: %v", err)
	}
	defer rows.Close()
	
	hasData := false
	for rows.Next() {
		hasData = true
		var version uint64
		var dirty bool
		if err := rows.Scan(&version, &dirty); err != nil {
			log.Fatalf("Failed to scan migration data: %v", err)
		}
		fmt.Printf("  version: %d, dirty: %v\n", version, dirty)
	}
	
	if !hasData {
		fmt.Println("  (no rows)")
	}
}
