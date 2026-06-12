package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oalpha/internal/agent"
	agentportfolio "github.com/oalpha/internal/agent/portfolio"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/internal/ml"
	"github.com/oalpha/pkg/models"
)

type AgentControlRequest struct {
	Symbol             string  `json:"symbol" binding:"required"`
	StrategyType       string  `json:"strategy_type" binding:"required"`
	Timeframe          string  `json:"timeframe"`
	InitialCash        float64 `json:"initial_cash"`
	UseWebSocket       bool    `json:"use_websocket"`
	QNoise             float64 `json:"q_noise"`
	RNoise             float64 `json:"r_noise"`
	ZThreshold         float64 `json:"z_threshold"`
	FastPeriod         int     `json:"fast_period"`
	SlowPeriod         int     `json:"slow_period"`
	RiskProfile        string  `json:"risk_profile"`
	RegimeMode         string  `json:"regime_mode"`
	RiskOverlayEnabled *bool   `json:"risk_overlay_enabled"`
	BaseStrategyType   string  `json:"base_strategy_type"`
	ModelPath          string  `json:"model_path"`
	MetadataPath       string  `json:"metadata_path"`
	ModelRegistryRoot  string  `json:"model_registry_root"`
	ModelName          string  `json:"model_name"`
	StrategyScope      string  `json:"strategy_scope"`
	MLEnterLong        float64 `json:"ml_enter_long"`
	MLMaxWeight        float64 `json:"ml_max_weight"`
	MLFailOpenOnError  *bool   `json:"ml_fail_open_on_error"`
}

// LaunchLiveAgent starts a user-scoped paper trading worker.
func (h *Handler) LaunchLiveAgent(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	var req AgentControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))
	req.StrategyType = strings.ToUpper(strings.TrimSpace(req.StrategyType))
	if req.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}
	if req.Timeframe == "" {
		req.Timeframe = "1Hour"
	}
	if req.InitialCash <= 0 {
		req.InitialCash = 100000.0
	}

	var strat backtest.Strategy
	useHMMEnsemble := false
	riskProfile := agent.RiskProfileModerate
	regimeMode := agent.RegimeModeOverlay
	switch req.StrategyType {
	case "ML_META_LABEL":
		baseReq := models.BacktestRequest{
			Symbol:     req.Symbol,
			QNoise:     req.QNoise,
			RNoise:     req.RNoise,
			ZThreshold: req.ZThreshold,
			FastPeriod: req.FastPeriod,
			SlowPeriod: req.SlowPeriod,
			Parameters: mlAgentParams(req),
		}
		baseType := strings.ToUpper(strings.TrimSpace(req.BaseStrategyType))
		if baseType == "" {
			baseType = "MA_CROSSOVER"
		}
		base, err := buildBaseSingleSymbolStrategy(baseType, baseReq)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		predictor, featureSpec, artifactThresholds, calibration, err := loadMLPredictorFromParams(baseReq.Parameters)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		strat = &ml.MLMetaLabelStrategy{
			Symbol:         req.Symbol,
			BaseStrategy:   base,
			FeatureBuilder: ml.NewFeatureBuilder(featureSpec),
			Predictor:      predictor,
			Calibration:    calibration,
			Thresholds:     mlThresholdsFromParams(baseReq.Parameters, artifactThresholds),
			MaxWeight:      req.MLMaxWeight,
		}
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
	case "MA_CROSSOVER":
		if req.FastPeriod == 0 {
			req.FastPeriod = 10
		}
		if req.SlowPeriod == 0 {
			req.SlowPeriod = 30
		}
		if req.FastPeriod >= req.SlowPeriod {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fast_period must be less than slow_period"})
			return
		}
		strat = backtest.NewMACrossoverStrategy(req.FastPeriod, req.SlowPeriod)
	case "HMM", "HMM_ENSEMBLE":
		req.StrategyType = "HMM_ENSEMBLE"
		useHMMEnsemble = true
		if req.FastPeriod == 0 {
			req.FastPeriod = 20
		}
		if req.SlowPeriod == 0 {
			req.SlowPeriod = 50
		}
		if req.FastPeriod >= req.SlowPeriod {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fast_period must be less than slow_period"})
			return
		}
		if req.QNoise == 0 {
			req.QNoise = 0.001
		}
		if req.RNoise == 0 {
			req.RNoise = 0.01
		}
		if req.ZThreshold == 0 {
			req.ZThreshold = 2.0
		}
		var err error
		riskProfile, req.RiskProfile, err = parseRiskProfile(req.RiskProfile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		regimeMode, req.RegimeMode, err = parseRegimeMode(req.RegimeMode, req.RiskOverlayEnabled)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "strategy_type must be KALMAN, MA_CROSSOVER, HMM_ENSEMBLE, or ML_META_LABEL"})
		return
	}

	if h.AgentManager.IsAgentRunning(userID, req.Symbol) {
		c.JSON(http.StatusConflict, gin.H{"error": "agent is already running for this symbol"})
		return
	}

	parameters := map[string]interface{}{
		"strategy_type": req.StrategyType,
		"q_noise":       req.QNoise,
		"r_noise":       req.RNoise,
		"z_threshold":   req.ZThreshold,
		"fast_period":   req.FastPeriod,
		"slow_period":   req.SlowPeriod,
		"risk_profile":  req.RiskProfile,
		"regime_mode":   req.RegimeMode,
	}
	if req.StrategyType == "ML_META_LABEL" {
		parameters["base_strategy_type"] = req.BaseStrategyType
		parameters["model_path"] = req.ModelPath
		parameters["metadata_path"] = req.MetadataPath
		parameters["model_registry_root"] = req.ModelRegistryRoot
		parameters["model_name"] = req.ModelName
		parameters["strategy_scope"] = req.StrategyScope
		parameters["ml_enter_long"] = req.MLEnterLong
		parameters["ml_max_weight"] = req.MLMaxWeight
		if req.MLFailOpenOnError != nil {
			parameters["ml_fail_open_on_error"] = *req.MLFailOpenOnError
		}
	}
	if req.RiskOverlayEnabled != nil {
		parameters["risk_overlay_enabled"] = *req.RiskOverlayEnabled
	}
	runID, err := h.AgentRepo.CreateAgentRun(
		c.Request.Context(),
		userID,
		req.Symbol,
		req.StrategyType,
		req.Timeframe,
		"paper",
		req.InitialCash,
		req.UseWebSocket,
		parameters,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if useHMMEnsemble {
		err = h.AgentManager.StartAgentV2(
			c.Request.Context(),
			userID,
			req.Symbol,
			req.Timeframe,
			req.FastPeriod,
			req.SlowPeriod,
			req.QNoise,
			req.RNoise,
			req.ZThreshold,
			true,
			req.InitialCash,
			runID,
			riskProfile,
			regimeMode,
			req.UseWebSocket,
		)
	} else {
		err = h.AgentManager.StartAgent(
			c.Request.Context(),
			userID,
			req.Symbol,
			req.Timeframe,
			strat,
			true,
			req.InitialCash,
			runID,
			req.UseWebSocket,
		)
	}
	if err != nil {
		_ = h.AgentRepo.MarkAgentRunFailed(c.Request.Context(), runID, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.AgentRepo.MarkAgentRunRunning(c.Request.Context(), runID); err != nil {
		_ = h.AgentManager.StopAgent(userID, req.Symbol)
		_ = h.AgentRepo.MarkAgentRunFailed(c.Request.Context(), runID, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "activated", "symbol": req.Symbol, "run_id": runID})
}

func mlAgentParams(req AgentControlRequest) map[string]interface{} {
	params := map[string]interface{}{
		"model_path":          req.ModelPath,
		"metadata_path":       req.MetadataPath,
		"model_registry_root": req.ModelRegistryRoot,
		"model_name":          req.ModelName,
		"strategy_scope":      req.StrategyScope,
		"ml_max_weight":       req.MLMaxWeight,
	}
	if req.ModelName == "" {
		params["model_name"] = "ml_meta_label"
	}
	if req.MLEnterLong > 0 {
		params["enter_long"] = req.MLEnterLong
	}
	if req.MLFailOpenOnError != nil {
		params["fail_open_on_error"] = *req.MLFailOpenOnError
	}
	return params
}

// TerminateLiveAgent stops a user-scoped worker.
func (h *Handler) TerminateLiveAgent(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	var req struct {
		Symbol string `json:"symbol" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))
	if req.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	err := h.AgentManager.StopAgent(userID, req.Symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err := h.AgentRepo.MarkLatestAgentRunStopped(c.Request.Context(), userID, req.Symbol, "user_requested"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "terminated", "symbol": req.Symbol})
}

// GetUserSettings returns saved agent settings for the authenticated user.
func (h *Handler) GetUserSettings(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
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

// SaveUserSettings validates and persists agent settings for the authenticated user.
func (h *Handler) SaveUserSettings(c *gin.Context) {
	var req struct {
		RiskProfile   string  `json:"risk_profile" binding:"required"`
		Leverage      int     `json:"leverage" binding:"required"`
		MaxPositions  int     `json:"max_positions" binding:"required"`
		StopLossPct   float64 `json:"stop_loss_pct" binding:"required"`
		TakeProfitPct float64 `json:"take_profit_pct" binding:"required"`
		RebalanceFreq string  `json:"rebalance_freq" binding:"required"`
		StrategyKey   string  `json:"strategy_key"`
		BacktestOK    bool    `json:"backtest_accepted"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}
	if !validAgentSettings(req.RiskProfile, req.Leverage, req.MaxPositions, req.StopLossPct, req.TakeProfitPct, req.RebalanceFreq) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent settings are outside supported bounds"})
		return
	}

	currentSettings, err := h.AgentRepo.GetAgentSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if currentSettings != nil && currentSettings.RiskProfile != req.RiskProfile {
		activeRuns, err := h.AgentRepo.ListActiveAgentRuns(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if len(activeRuns) > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "stop the running agent before changing risk_profile"})
			return
		}
		if !req.BacktestOK {
			c.JSON(http.StatusBadRequest, gin.H{"error": "accepted catalog backtest is required before changing risk_profile"})
			return
		}
		strategyKey := strings.ToLower(strings.TrimSpace(req.StrategyKey))
		if strategyKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "strategy_key is required before changing risk_profile"})
			return
		}
		spec, err := agentportfolio.StrategySpecByKey(strategyKey, nil, agentportfolio.DefaultStrategyCatalogConfig())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if spec.RiskProfile != onboardingRiskBucket(req.RiskProfile) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "strategy_key does not match requested risk_profile"})
			return
		}
	}

	settings := &db.AgentSettings{
		UserID:        userID,
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

func parseRiskProfile(value string) (agent.RiskProfile, string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "moderate":
		return agent.RiskProfileModerate, "moderate", nil
	case "conservative":
		return agent.RiskProfileConservative, "conservative", nil
	case "aggressive":
		return agent.RiskProfileAggressive, "aggressive", nil
	default:
		return agent.RiskProfileModerate, "", fmt.Errorf("risk_profile must be conservative, moderate, or aggressive")
	}
}

func parseRegimeMode(value string, riskOverlayEnabled *bool) (agent.RegimeMode, string, error) {
	raw := strings.ToLower(strings.TrimSpace(value))
	if raw == "" && riskOverlayEnabled != nil {
		if *riskOverlayEnabled {
			raw = string(agent.RegimeModeOverlay)
		} else {
			raw = string(agent.RegimeModeNone)
		}
	}

	switch raw {
	case "", "overlay", "risk_overlay", "hmm_overlay", "on", "enabled", "true":
		return agent.RegimeModeOverlay, string(agent.RegimeModeOverlay), nil
	case "none", "off", "disabled", "false", "no_hmm":
		return agent.RegimeModeNone, string(agent.RegimeModeNone), nil
	default:
		return agent.RegimeModeOverlay, "", fmt.Errorf("regime_mode must be overlay or none")
	}
}

func validAgentSettings(riskProfile string, leverage, maxPositions int, stopLossPct, takeProfitPct float64, rebalanceFreq string) bool {
	switch riskProfile {
	case "conservative", "moderate", "aggressive":
	default:
		return false
	}
	switch rebalanceFreq {
	case "hourly", "daily", "weekly", "monthly":
	default:
		return false
	}
	if leverage < 1 || leverage > 10 {
		return false
	}
	if maxPositions < 1 || maxPositions > 100 {
		return false
	}
	if stopLossPct <= 0 || stopLossPct > 100 {
		return false
	}
	if takeProfitPct <= 0 || takeProfitPct > 100 {
		return false
	}
	return true
}
