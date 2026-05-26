package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	apiAuth "github.com/oalpha/internal/api/auth"
	"github.com/oalpha/internal/auth"
	"github.com/oalpha/internal/config"
	"github.com/oalpha/internal/db"
)

// NewRouter builds the Gin engine with routes and middleware.
func NewRouter(h *Handler, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	// Configure Cross-Origin Resource Sharing (CORS) rules for the frontend dashboard interface
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize authentication tracking services natively from handler connections
	userRepo := db.NewUserRepository(h.repo.GetDB())
	authService := auth.NewAuthService(userRepo, cfg.JWTSecret, 24*time.Hour)
	authHandler := apiAuth.NewAuthHandler(authService, userRepo)

	// Global System Health Telemetry Verification Endpoint
	r.GET("/health", h.Health)

	// User Session Authentication Groups
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/validate", authHandler.ValidateToken)
		authGroup.GET("/me", authHandler.GetCurrentUser)
	}

	// Quantitative Core Analytical Endpoints (V1 Public Demo Scope)
	v1 := r.Group("/api/v1")
	{
		// Historical Engine Simulation Interface
		v1.POST("/backtest", h.RunBacktest)

		// Live Agent Manager Orchestration Controllers
		v1.POST("/agent/start", h.LaunchLiveAgent)
		v1.POST("/agent/stop", h.TerminateLiveAgent)
	}

	return r
}
