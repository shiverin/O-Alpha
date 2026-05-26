package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oalpha/pkg/models"
)

// UserRepository provides data access for users.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository returns a UserRepository backed by db.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user into the database.
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	const q = `
		INSERT INTO users (username, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	var id int64
	err := r.db.QueryRow(ctx, q, user.Username, user.PasswordHash, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	user.ID = id
	return nil
}

// GetUserByUsername retrieves a user by username.
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	const q = `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1`

	var u models.User
	err := r.db.QueryRow(ctx, q, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select user: %w", err)
	}
	return &u, nil
}

// GetUserByID retrieves a user by ID.
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	const q = `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1`

	var u models.User
	err := r.db.QueryRow(ctx, q, id).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select user by id: %w", err)
	}
	return &u, nil
}
