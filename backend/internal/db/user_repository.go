package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oalpha/pkg/models"
)

// ErrUsernameTaken is returned when a registration request hits a unique constraint conflict.
var ErrUsernameTaken = errors.New("username already taken")

// UserRepository provides data access for users.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository returns a UserRepository backed by db.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a user profile record and handles atomic provisioning of a baseline portfolio state.
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	// Begin an isolated database transaction to guarantee cross-table atomicity
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin user registration transaction: %w", err)
	}

	// Ensure clean rollback protection if an execution error occurs before Commit() runs
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// User defaults are owned and managed cleanly by the database schema
	const userQuery = `
		INSERT INTO users (
			username, 
			password_hash, 
			display_name
		)
		VALUES ($1, $2, $3)
		RETURNING id, is_onboarded, created_at, updated_at`

	var id int64
	err = tx.QueryRow(ctx, userQuery, user.Username, user.PasswordHash, user.Username).Scan(
		&id,
		&user.IsOnboarded,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		// Detect unique constraint violations (Code 23505) and surface clean domain semantics
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUsernameTaken
		}
		return fmt.Errorf("insert user profile identity failed: %w", err)
	}

	// Assign the generated database reference parameters back into our structural model tracking object
	user.ID = id

	// 🚀 Clean DB-driven initialization. All metrics, asset balances, and timestamp
	// allocations are generated dynamically on the database engine.
	const portfolioQuery = `
		INSERT INTO portfolio_snapshots (user_id)
		VALUES ($1)`

	_, err = tx.Exec(ctx, portfolioQuery, id)
	if err != nil {
		return fmt.Errorf("failed to provision baseline account portfolio snapshot: %w", err)
	}

	// Atomically commit all records across both tables to disk simultaneously
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit atomic account provisioning transaction: %w", err)
	}

	return nil
}

// GetUserByUsername retrieves a user by username, including their active onboarding status.
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	const q = `
        SELECT id, username, password_hash, is_onboarded, created_at, updated_at
        FROM users
        WHERE username = $1`

	var u models.User
	err := r.db.QueryRow(ctx, q, username).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&u.IsOnboarded,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { // High-performance sentinel checking
			return nil, nil
		}
		return nil, fmt.Errorf("select user by username failed: %w", err)
	}
	return &u, nil
}

// GetUserByID retrieves a user by ID, including their active onboarding status.
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	const q = `
        SELECT id, username, password_hash, is_onboarded, created_at, updated_at
        FROM users
        WHERE id = $1`

	var u models.User
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&u.IsOnboarded,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { // High-performance sentinel checking
			return nil, nil
		}
		return nil, fmt.Errorf("select user by id failed: %w", err)
	}
	return &u, nil
}
