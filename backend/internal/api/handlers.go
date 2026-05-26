package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oalpha/internal/agent"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/pkg/models"
)

// Handler handles systematic HTTP network request parameters.
type Handler struct {
	repo         *db.BarsRepository
	AgentManager *agent.AgentManager
	AgentRepo    *db.AgentRepository
}

// NewHandler builds a clean instance of your endpoint coordinator dependencies.
func NewHandler(repo *db.BarsRepository, am *agent.AgentManager, ar *db.AgentRepository) *Handler {
	return &Handler{
		repo:         repo,
		AgentManager: am,
		AgentRepo:    ar,
	}
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

	// Fix 1: Ensure timeframe has a solid default fallback state
	if req.Timeframe == "" {
		req.Timeframe = "1Day"
	}

	// Fix 2: Pass req.Timeframe into the updated partitioned data query layer
	bars, err := h.repo.GetBars(c.Request.Context(), req.Symbol, req.Timeframe, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(bars) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no bars found for symbol"})
		return
	}

	// Fix 3: Calculate intervals dynamically to match current partitions
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

	// Fix 4: Pass req.Timeframe down to the updated validation report signature
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

type AgentControlRequest struct {
	UserID       int64   `json:"user_id"`
	Symbol       string  `json:"symbol" binding:"required"`
	StrategyType string  `json:"strategy_type" binding:"required"`
	Timeframe    string  `json:"timeframe"`
	InitialCash  float64 `json:"initial_cash"`
	UseWebSocket bool    `json:"use_websocket"`
	QNoise       float64 `json:"q_noise"`
	RNoise       float64 `json:"r_noise"`
	ZThreshold   float64 `json:"z_threshold"`
	FastPeriod   int     `json:"fast_period"`
	SlowPeriod   int     `json:"slow_period"`
}

// LaunchLiveAgent provisions and kicks off a real-time live trading process loop.
func (h *Handler) LaunchLiveAgent(c *gin.Context) {
	var req AgentControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserID == 0 {
		req.UserID = 999
	}
	if req.Timeframe == "" {
		req.Timeframe = "1Hour"
	}
	if req.InitialCash <= 0 {
		req.InitialCash = 50000.0
	}

	// Resolve the desired Strategy strategy pattern mapping
	var strat backtest.Strategy
	switch req.StrategyType {
	case "KALMAN":
		if req.QNoise == 0 {
			req.QNoise = 0.01
		}
		if req.RNoise == 0 {
			req.RNoise = 0.5
		}
		if req.ZThreshold == 0 {
			req.ZThreshold = 2.0
		}
		strat = backtest.NewKalmanStrategy(req.QNoise, req.RNoise, 20, req.ZThreshold)
	default:
		if req.FastPeriod == 0 {
			req.FastPeriod = 10
		}
		if req.SlowPeriod == 0 {
			req.SlowPeriod = 30
		}
		strat = backtest.NewMACrossoverStrategy(req.FastPeriod, req.SlowPeriod)
	}

	// Fix 5: Cleaned up and removed the broken, uncompiled trailing GetBars block here
	err := h.AgentManager.StartAgent(
		c.Request.Context(),
		req.UserID,
		req.Symbol,
		req.Timeframe,
		strat,
		true, // Enforce paper trading flags for safety thresholds in public demos
		req.InitialCash,
		req.UseWebSocket,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "activated", "symbol": req.Symbol})
}

// TerminateLiveAgent stops a matching active process loop instantly.
func (h *Handler) TerminateLiveAgent(c *gin.Context) {
	var req struct {
		UserID int64  `json:"user_id"`
		Symbol string `json:"symbol" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserID == 0 {
		req.UserID = 999
	}

	err := h.AgentManager.StopAgent(req.UserID, req.Symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "terminated", "symbol": req.Symbol})
}

// GetUserSettings evaluates if a configuration state is available for the given user profile.
func (h *Handler) GetUserSettings(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id query parameter is required"})
		return
	}

	var userID int64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
		return
	}

	settings, err := h.AgentRepo.GetAgentSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if settings == nil {
		c.JSON(http.StatusOK, gin.H{"found": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"found": true, "settings": settings})
}

// SaveUserSettings ingests settings payloads and synchronizes active database models.
func (h *Handler) SaveUserSettings(c *gin.Context) {
	var req struct {
		UserID        int64   `json:"user_id" binding:"required"`
		RiskProfile   string  `json:"risk_profile" binding:"required"`
		Leverage      int     `json:"leverage" binding:"required"`
		MaxPositions  int     `json:"max_positions" binding:"required"`
		StopLossPct   float64 `json:"stop_loss_pct" binding:"required"`
		TakeProfitPct float64 `json:"take_profit_pct" binding:"required"`
		RebalanceFreq string  `json:"rebalance_freq" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings := &db.AgentSettings{
		UserID:        req.UserID,
		RiskProfile:   req.RiskProfile,
		Leverage:      req.Leverage,
		MaxPositions:  req.MaxPositions,
		StopLossPct:   req.StopLossPct,
		TakeProfitPct: req.TakeProfitPct,
		RebalanceFreq: req.RebalanceFreq,
	}

	if err := h.AgentRepo.SaveAgentSettings(c.Request.Context(), settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "synchronized"})
}
