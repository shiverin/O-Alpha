package api

import (
	"net/http"

	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPortfolioSummary fetches the latest portfolio snapshot for dashboard views.
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

// GetActivePositions returns currently open portfolio positions.
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

// GetExecutionStream returns the user's latest persisted trade records.
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
	limit = clampLimit(limit, 50, 500)

	trades, err := h.PortfolioRepo.GetExecutionStream(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trades)
}

// GetSystemAlerts returns recent user-scoped alerts.
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
	limit = clampLimit(limit, 10, 100)

	alerts, err := h.PortfolioRepo.GetSystemAlerts(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetPortfolioHistory returns snapshots sorted chronologically for charting.
func (h *Handler) GetPortfolioHistory(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	limit := 30
	if queryLimit := c.Query("limit"); queryLimit != "" {
		if parsed, err := strconv.Atoi(queryLimit); err == nil {
			limit = parsed
		}
	}
	limit = clampLimit(limit, 30, 365)

	history, err := h.PortfolioRepo.GetPortfolioHistory(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

func clampLimit(value, fallback, max int) int {
	if value <= 0 {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}
