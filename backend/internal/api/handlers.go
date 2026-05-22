package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/pkg/models"
)

// Handler holds HTTP dependencies.
type Handler struct {
	repo *db.Repository
}

// NewHandler creates API handlers.
func NewHandler(repo *db.Repository) *Handler {
	return &Handler{repo: repo}
}

// Health returns service status.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}

// RunBacktest executes a backtest for the requested symbol using the specified strategy.
func (h *Handler) RunBacktest(c *gin.Context) {
	var req models.BacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.FastPeriod >= req.SlowPeriod {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fast_period must be less than slow_period"})
		return
	}

	end := time.Now().UTC()
	start := end.Add(-365 * 24 * time.Hour)
	if req.Start != nil {
		start = req.Start.UTC()
	}
	if req.End != nil {
		end = req.End.UTC()
	}

	bars, err := h.repo.GetBars(c.Request.Context(), req.Symbol, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(bars) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no bars found for symbol"})
		return
	}

	initialCash := req.InitialCash
	if initialCash <= 0 {
		initialCash = 100_000
	}

	// Create MA crossover strategy (maintains backward compatibility)
	strat := backtest.NewMACrossoverStrategy(req.FastPeriod, req.SlowPeriod)

	result, err := backtest.RunBacktest(c.Request.Context(), bars, strat, initialCash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
