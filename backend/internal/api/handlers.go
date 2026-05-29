package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oalpha/internal/agent"
	"github.com/oalpha/internal/db"
)

const contextUserIDKey = "userID"

// Handler wires repositories and service coordinators into HTTP endpoints.
type Handler struct {
	repo          *db.BarsRepository
	AgentManager  *agent.AgentManager
	AgentRepo     *db.AgentRepository
	PortfolioRepo *db.PortfolioRepository
}

// NewHandler constructs an HTTP handler with its backing dependencies.
func NewHandler(repo *db.BarsRepository, am *agent.AgentManager, ar *db.AgentRepository, pr *db.PortfolioRepository) *Handler {
	return &Handler{
		repo:          repo,
		AgentManager:  am,
		AgentRepo:     ar,
		PortfolioRepo: pr,
	}
}

// Health returns service status.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}

// deriveUserIDQuery returns the authenticated user ID injected by auth middleware.
func (h *Handler) deriveUserIDQuery(c *gin.Context) (int64, bool) {
	value, exists := c.Get(contextUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authenticated user context is required"})
		return 0, false
	}

	userID, ok := value.(int64)
	if !ok || userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authenticated user context is invalid"})
		return 0, false
	}

	return userID, true
}
