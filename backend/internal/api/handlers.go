package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oalpha/internal/agent"
	"github.com/oalpha/internal/db"
)

// Handler handles systematic HTTP network request parameters.
type Handler struct {
	repo          *db.BarsRepository
	AgentManager  *agent.AgentManager
	AgentRepo     *db.AgentRepository
	PortfolioRepo *db.PortfolioRepository
}

// NewHandler builds a clean instance of your endpoint coordinator dependencies.
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

// deriveUserIDQuery is a private helper method to handle duplicate query param parsing logic gracefully.
func (h *Handler) deriveUserIDQuery(c *gin.Context) (int64, bool) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id verification context parameter is mandatory"})
		return 0, false
	}

	var userID int64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "malformed user_id variable footprint"})
		return 0, false
	}

	return userID, true
}