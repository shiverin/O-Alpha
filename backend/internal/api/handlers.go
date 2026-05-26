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

	// Only enforce MA rules if they are explicitly using MA_CROSSOVER or defaulting to it
	if req.StrategyType == "MA_CROSSOVER" || req.StrategyType == "" {
		fast := req.FastPeriod
		if fast == 0 {
			fast = 10
		} // Safe defaults
		slow := req.SlowPeriod
		if slow == 0 {
			slow = 30
		}

		if fast >= slow {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fast_period must be less than slow_period"})
			return
		}
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

	var strat backtest.Strategy

	// Dynamically route the strategy based on frontend input
	switch req.StrategyType {
	case "KALMAN":
		q := req.QNoise
		if q == 0 {
			q = 0.01
		}
		r := req.RNoise
		if r == 0 {
			r = 0.5
		}
		z := req.ZThreshold
		if z == 0 {
			z = 2.0
		}

		strat = backtest.NewKalmanStrategy(q, r, 20, z)

	case "MA_CROSSOVER":
		fast := req.FastPeriod
		if fast == 0 {
			fast = 10
		}
		slow := req.SlowPeriod
		if slow == 0 {
			slow = 30
		}
		strat = backtest.NewMACrossoverStrategy(fast, slow)

	default:
		// Default to MA crossover if nothing is specified for backward compatibility
		fast := req.FastPeriod
		if fast == 0 {
			fast = 10
		}
		slow := req.SlowPeriod
		if slow == 0 {
			slow = 30
		}
		strat = backtest.NewMACrossoverStrategy(fast, slow)
	}

	result, err := backtest.RunBacktest(c.Request.Context(), bars, strat, initialCash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
