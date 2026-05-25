package api

import (
"time"

"github.com/gin-contrib/cors"
"github.com/gin-gonic/gin"

"github.com/oalpha/internal/auth"
"github.com/oalpha/internal/db"
apiAuth "github.com/oalpha/internal/api/auth"
)

// NewRouter builds the Gin engine with routes and middleware.
func NewRouter(h *Handler) *gin.Engine {
r := gin.New()
r.Use(gin.Recovery(), gin.Logger())

r.Use(cors.New(cors.Config{
AllowOrigins:     []string{"http://localhost:3000"},
AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
ExposeHeaders:    []string{"Content-Length"},
AllowCredentials: true,
MaxAge:           12 * time.Hour,
}))

// Initialize auth service
userRepo := db.NewUserRepository(h.repo.GetDB())
authService := auth.NewAuthService(userRepo, "your-secret-key", 24*time.Hour)
authHandler := apiAuth.NewAuthHandler(authService, userRepo)

r.GET("/health", h.Health)

// Auth routes
auth := r.Group("/auth")
{
auth.POST("/login", authHandler.Login)
auth.POST("/logout", authHandler.Logout)
auth.POST("/validate", authHandler.ValidateToken)
auth.GET("/me", authHandler.GetCurrentUser)
}

v1 := r.Group("/api/v1")
{
v1.POST("/backtest", h.RunBacktest)
}

return r
}
