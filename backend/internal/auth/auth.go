package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/pkg/models"
)

// Claims defines a strongly-typed JWT payload structure to eliminate map casting traps.
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthService handles authentication logic.
type AuthService struct {
	userRepo    *db.UserRepository
	jwtSecret   []byte
	tokenExpiry time.Duration
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo *db.UserRepository, jwtSecret string, tokenExpiry time.Duration) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtSecret:   []byte(jwtSecret),
		tokenExpiry: tokenExpiry,
	}
}

// Login authenticates a user and preserves the current auto-registration behavior.
func (s *AuthService) Login(ctx context.Context, username, password string) (string, *models.User, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, err
	}

	if user == nil {
		newUser := &models.User{Username: username}
		if err := newUser.SetPassword(password); err != nil {
			return "", nil, err
		}
		if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
			return "", nil, err
		}
		user = newUser
	}

	if err := user.ComparePassword(password); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return t, user, nil
}

// ValidateToken validates a JWT token and returns the user ID safely without float64 casting.
func (s *AuthService) ValidateToken(tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	return claims.UserID, nil
}
