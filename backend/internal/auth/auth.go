package auth

import (
	"context"
	"errors"
	"time"

	"github.com/oalpha/internal/db"
	"github.com/golang-jwt/jwt/v5"
)

// AuthService handles authentication logic.
type AuthService struct {
	userRepo *db.UserRepository
	jwtSecret []byte
	tokenExpiry time.Duration
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo *db.UserRepository, jwtSecret string, tokenExpiry time.Duration) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtSecret: []byte(jwtSecret),
		tokenExpiry: tokenExpiry,
	}
}

// Login authenticates a user with email and password.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	if err := user.ComparePassword(password); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(s.tokenExpiry).Unix()

	t, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return t, nil
}

// ValidateToken validates a JWT token and returns the user ID if valid.
func (s *AuthService) ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64UserID, ok := claims["user_id"].(float64); ok {
			return int64(float64UserID), nil
		}
		return 0, errors.New("invalid user ID in token")
	}

	return 0, errors.New("invalid token")
}