package auth

import (
	"context"
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
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse represents the login response body.
type LoginResponse struct {
	Token string `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse represents the user data returned in auth responses.
type UserResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

// Login handles user login requests.
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	token, err := h.authService.Login(context.Background(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Get user info for response
	user, err := h.userRepo.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user info"})
		return
	}

	response := LoginResponse{
		Token: token,
		User: UserResponse{
			ID:   user.ID,
			Email: user.Email,
		},
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout requests.
// @Summary Logout user
// @Description Invalidate user session (client should remove token)
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT auth system, logout is handled client-side
	// by removing the token. This endpoint exists for completeness.
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// ValidateToken validates the provided JWT token.
// @Summary Validate token
// @Description Check if the provided token is valid
// @Tags auth
// @Accept json
// @Produce json
// @Param token body map[string]string true "JWT token"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/validate [post]
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
// @Summary Get current user
// @Description Returns the currently authenticated user based on JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	// Extract token from Bearer token format
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

	user, err := h.userRepo.GetUserByID(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	response := UserResponse{
		ID:   user.ID,
		Email: user.Email,
	}

	c.JSON(http.StatusOK, response)
}