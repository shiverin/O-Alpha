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

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	userRepo := db.NewUserRepository(h.repo.GetDB())
	authService := auth.NewAuthService(userRepo, cfg.JWTSecret, 24*time.Hour)
	authHandler := apiAuth.NewAuthHandler(authService, userRepo)

	r.GET("/health", h.Health)

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/validate", authHandler.ValidateToken)
		authGroup.GET("/me", authHandler.GetCurrentUser)
	}

	authenticatedV1 := AuthMiddleware(authService)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/backtest", h.RunBacktest)

		protected := v1.Group("")
		protected.Use(authenticatedV1)

		protected.POST("/agent/start", h.LaunchLiveAgent)
		protected.POST("/agent/stop", h.TerminateLiveAgent)

		protected.GET("/user/settings", h.GetUserSettings)
		protected.POST("/user/settings", h.SaveUserSettings)

		protected.POST("/user/onboarding/complete", h.CompleteUserOnboarding)

		protected.GET("/user/portfolio/summary", h.GetPortfolioSummary)
		protected.GET("/user/portfolio/history", h.GetPortfolioHistory)
		protected.GET("/user/portfolio/positions", h.GetActivePositions)
		protected.GET("/user/portfolio/trades", h.GetExecutionStream)
		protected.GET("/user/portfolio/alerts", h.GetSystemAlerts)
	}

	return r
}
