package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

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
		}
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

	if req.Timeframe == "" {
		req.Timeframe = "1Day"
	}

	bars, err := h.repo.GetBars(c.Request.Context(), req.Symbol, req.Timeframe, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(bars) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no bars found for symbol"})
		return
	}

	var expectedInterval time.Duration
	switch req.Timeframe {
	case "1Min":
		expectedInterval = time.Minute
	case "5Min":
		expectedInterval = 5 * time.Minute
	case "15Min":
		expectedInterval = 15 * time.Minute
	case "1Hour":
		expectedInterval = time.Hour
	default:
		expectedInterval = 24 * time.Hour
	}

	report, err := h.repo.ValidateData(c.Request.Context(), req.Symbol, req.Timeframe, start, end, expectedInterval)
	if err == nil && report != nil {
		if report.InvalidBars > 0 && (float64(report.InvalidBars)/float64(report.BarCount)) > 0.05 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         "backtest aborted: historical data quality is too poor",
				"invalid_count": report.InvalidBars,
				"total_scanned": report.BarCount,
			})
			return
		}
	}

	initialCash := req.InitialCash
	if initialCash <= 0 {
		initialCash = 100000
	}

	var strat backtest.Strategy

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
