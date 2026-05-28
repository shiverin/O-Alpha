package api

import (
	"net/http"

	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPortfolioSummary fetches high-level balance snap postures for main analytics widgets.
func (h *Handler) GetPortfolioSummary(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	summary, err := h.PortfolioRepo.GetLatestSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if summary == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no portfolio history snapshots discovered"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetActivePositions returns open exposures and dynamic calculations for table charts.
func (h *Handler) GetActivePositions(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	positions, err := h.PortfolioRepo.GetActivePositions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, positions)
}

// GetExecutionStream isolates historical trade logging for the Activity Console grid.
func (h *Handler) GetExecutionStream(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	limit := 50

	if queryLimit := c.Query("limit"); queryLimit != "" {
		if parsed, err := strconv.Atoi(queryLimit); err == nil {
			limit = parsed
		}
	}

	trades, err := h.PortfolioRepo.GetExecutionStream(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trades)
}

// GetSystemAlerts parses the security warning log stack for dashboard telemetry components.
func (h *Handler) GetSystemAlerts(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	limit := 10
	if queryLimit := c.Query("limit"); queryLimit != "" {
		if parsed, err := strconv.Atoi(queryLimit); err == nil {
			limit = parsed
		}
	}

	alerts, err := h.PortfolioRepo.GetSystemAlerts(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// 📂 internal/api/portfolio_handler.go

// GetPortfolioHistory handles requests for historical equity coordinates
func (h *Handler) GetPortfolioHistory(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	// Default lookback point resolution for sparklines
	limit := 30
	if queryLimit := c.Query("limit"); queryLimit != "" {
		if parsed, err := strconv.Atoi(queryLimit); err == nil {
			limit = parsed
		}
	}

	history, err := h.PortfolioRepo.GetPortfolioHistory(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}
