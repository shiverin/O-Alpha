package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Open connects to PostgreSQL using the given DSN and returns a *pgx.Conn.
func Open(databaseURL string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(context.Background()); err != nil {
		_ = conn.Close(context.Background())
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return conn, nil
}
