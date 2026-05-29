package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oalpha/internal/auth"
	"github.com/oalpha/internal/db"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService *auth.AuthService
	userRepo    *db.UserRepository
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(authService *auth.AuthService, userRepo *db.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
	}
}

// LoginRequest represents the login request body.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response body.
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse represents the user data returned in auth responses.
type UserResponse struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	IsOnboarded bool   `json:"is_onboarded"`
}

// Login handles user login requests.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	token, user, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	response := LoginResponse{
		Token: token,
		User: UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			IsOnboarded: user.IsOnboarded,
		},
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout requests.
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// ValidateToken validates the provided JWT token.
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var requestBody struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	_, err := h.authService.ValidateToken(requestBody.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// GetCurrentUser returns the currently authenticated user.
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	var token string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		token = authHeader
	}

	userID, err := h.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	user, err := h.userRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	response := UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		IsOnboarded: user.IsOnboarded,
	}

	c.JSON(http.StatusOK, response)
}
