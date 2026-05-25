package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oalpha/internal/db"
)

// Notice we dropped the "auth." prefix here because we are ALREADY in the auth package!
type AuthHandler struct {
	authService *AuthService 
	userRepo    *db.UserRepository
}

func NewAuthHandler(authService *AuthService, userRepo *db.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Registration coming soon!"})
}