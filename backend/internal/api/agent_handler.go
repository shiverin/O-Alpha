package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
)

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

	err := h.AgentManager.StartAgent(
		c.Request.Context(),
		req.UserID,
		req.Symbol,
		req.Timeframe,
		strat,
		true,
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
