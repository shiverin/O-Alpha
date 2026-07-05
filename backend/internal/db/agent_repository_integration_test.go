//go:build integration

package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSaveAgentSettingsRejectsActiveRunIntegration(t *testing.T) {
	ctx := context.Background()
	pool := openAgentRepositoryTestDB(t)
	repo := NewAgentRepository(pool)
	userID := createAgentRepositoryTestUser(t, pool)

	if err := repo.SaveAgentSettings(ctx, testAgentSettings(userID)); err != nil {
		t.Fatalf("save baseline settings: %v", err)
	}
	if _, err := repo.CreateAgentRun(ctx, userID, "VOO", "PORTFOLIO_CATALOG", "1Day", "paper", 100000, false, map[string]interface{}{}); err != nil {
		t.Fatalf("create active run: %v", err)
	}

	changed := testAgentSettings(userID)
	changed.Leverage = 3
	err := repo.SaveAgentSettings(ctx, changed)
	if !errors.Is(err, ErrActiveAgentSettingsLocked) {
		t.Fatalf("save changed settings error=%v, want ErrActiveAgentSettingsLocked", err)
	}
}

func TestSaveAgentSettingsAllowsStoppedRunIntegration(t *testing.T) {
	ctx := context.Background()
	pool := openAgentRepositoryTestDB(t)
	repo := NewAgentRepository(pool)
	userID := createAgentRepositoryTestUser(t, pool)

	if err := repo.SaveAgentSettings(ctx, testAgentSettings(userID)); err != nil {
		t.Fatalf("save baseline settings: %v", err)
	}
	if _, err := repo.CreateAgentRun(ctx, userID, "VOO", "PORTFOLIO_CATALOG", "1Day", "paper", 100000, false, map[string]interface{}{}); err != nil {
		t.Fatalf("create run: %v", err)
	}
	if _, err := repo.MarkActivePortfolioRunStopped(ctx, userID, "test_complete"); err != nil {
		t.Fatalf("stop run: %v", err)
	}

	changed := testAgentSettings(userID)
	changed.MaxPositions = 8
	if err := repo.SaveAgentSettings(ctx, changed); err != nil {
		t.Fatalf("save changed settings after stopped run: %v", err)
	}
}

func TestCreateAgentRunSerializesWithSettingsLockIntegration(t *testing.T) {
	ctx := context.Background()
	pool := openAgentRepositoryTestDB(t)
	repo := NewAgentRepository(pool)
	userID := createAgentRepositoryTestUser(t, pool)

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin lock tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockUserAgentSettingsTx(ctx, tx, userID); err != nil {
		t.Fatalf("lock user settings: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, runErr := repo.CreateAgentRun(ctx, userID, "VOO", "PORTFOLIO_CATALOG", "1Day", "paper", 100000, false, map[string]interface{}{})
		done <- runErr
	}()

	select {
	case err := <-done:
		t.Fatalf("CreateAgentRun completed before settings lock released: %v", err)
	case <-time.After(200 * time.Millisecond):
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit lock tx: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("CreateAgentRun after lock release: %v", err)
	}
}

func openAgentRepositoryTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("OALPHA_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("OALPHA_TEST_DATABASE_URL is required for integration tests")
	}
	migrationsPath := os.Getenv("OALPHA_TEST_MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://../../../migrations"
	}
	if err := RunMigrations(databaseURL, migrationsPath); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func createAgentRepositoryTestUser(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()
	username := fmt.Sprintf("settings_lock_%d", time.Now().UnixNano())
	var userID int64
	const q = `
		INSERT INTO users (username, password_hash, display_name)
		VALUES ($1, 'test-hash', 'Settings Lock Test')
		RETURNING id`
	if err := pool.QueryRow(context.Background(), q, username).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})
	return userID
}

func testAgentSettings(userID int64) *AgentSettings {
	return &AgentSettings{
		UserID:        userID,
		RiskProfile:   "moderate",
		Leverage:      2,
		MaxPositions:  6,
		StopLossPct:   2.5,
		TakeProfitPct: 5,
		RebalanceFreq: "daily",
	}
}
