package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system.
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	PasswordHash string    `json:"-"` // never JSON encoded
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SetPassword hashes the password and stores it in the PasswordHash field.
func (u *User) SetPassword(password string) error {
	if len(password) == 0 {
		return errors.New("password should not be empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// ComparePassword compares the hashed password with the provided password.
func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}