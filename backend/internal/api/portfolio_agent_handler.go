package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oalpha/internal/agent/portfolio"
)

const portfolioStrategyTypeTag = "PORTFOLIO_CATALOG"

var defaultPortfolioUniverse = []string{
	"VOO", "AAPL", "ADBE", "ADP", "AMAT", "AMD", "AMGN", "AMZN", "AVGO", "BKNG",
	"CMCSA", "COST", "CSCO", "DIA", "GILD", "GOOGL", "INTC", "INTU", "ISRG", "IWM",
	"LRCX", "MDLZ", "META", "MSFT", "NFLX", "NVDA", "PEP", "QCOM", "QQQ", "SBUX",
	"SMH", "SPY", "TSLA", "TXN", "VTI", "XLB", "XLE", "XLF", "XLI", "XLK",
	"XLP", "XLU", "XLV", "XLY", "HON", "ABBV", "ABT", "ACN", "GE", "KO",
	"SCHW", "AMT", "AXP", "BA", "BAC", "BLK", "BMY", "C", "CAT", "CI",
	"COP", "CRM", "CVS", "CVX", "DE", "DIS", "ELV", "GS", "HD", "IBM",
	"JNJ", "JPM", "LIN", "LLY", "LOW", "MA", "MCD", "MDT", "MO", "MRK",
	"NEE", "NKE", "NOW", "ORCL", "PFE", "PG", "PLD", "PM", "RTX", "SO",
	"SYK", "T", "TMO", "UNH", "UPS", "USB", "V", "VZ", "WMT", "XOM",
}

type PortfolioAgentRequest struct {
	StrategyKey string   `json:"strategy_key"`
	RiskProfile string   `json:"risk_profile"`
	Symbols     []string `json:"symbols"`
	Timeframe   string   `json:"timeframe"`
	InitialCash float64  `json:"initial_cash"`
}

func riskProfileDefaultStrategy(profile string) string {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case "conservative":
		return "ranker_proxy_h63_low"
	case "moderate":
		return "lgbm_ranker_h63_medium"
	case "aggressive":
		return "composite_momentum_high"
	default:
		return "lgbm_ranker_h63_low"
	}
}

func (h *Handler) LaunchPortfolioAgent(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}
	if h.Portfolio == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "portfolio agent is not configured on this server"})
		return
	}

	var req PortfolioAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	symbols := normalizeUniverse(req.Symbols)
	if len(symbols) == 0 {
		symbols = append([]string(nil), defaultPortfolioUniverse...)
	}

	strategyKey := strings.ToLower(strings.TrimSpace(req.StrategyKey))
	if strategyKey == "" || strategyKey == "auto" {
		strategyKey = riskProfileDefaultStrategy(req.RiskProfile)
	}

	timeframe := strings.TrimSpace(req.Timeframe)
	if timeframe == "" {
		timeframe = "1Day"
	}
	if timeframe != "1Day" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "portfolio catalog agents currently require timeframe 1Day"})
		return
	}

	initialCash := req.InitialCash
	if initialCash <= 0 {
		initialCash = 100000
	}

	spec, err := h.Portfolio.SpecByKey(strategyKey, symbols)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.Portfolio.IsRunningForUser(userID) {
		c.JSON(http.StatusConflict, gin.H{"error": "a portfolio agent is already running; stop it before starting another"})
		return
	}

	normalizedRiskProfile := strings.ToLower(strings.TrimSpace(req.RiskProfile))
	if normalizedRiskProfile != "" && normalizedRiskProfile != "conservative" && normalizedRiskProfile != "moderate" && normalizedRiskProfile != "aggressive" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "risk_profile must be conservative, moderate, or aggressive"})
		return
	}
	if normalizedRiskProfile != "" && spec.RiskProfile != onboardingRiskBucket(normalizedRiskProfile) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "strategy_key does not match requested risk_profile"})
		return
	}

	parameters := map[string]interface{}{
		"strategy_key":      strategyKey,
		"display_name":      spec.DisplayName,
		"family":            spec.Family,
		"risk_profile":      normalizedRiskProfile,
		"deployment_status": string(spec.DeploymentStatus),
		"paper_only":        spec.PaperOnly,
		"benchmark_symbol":  spec.BenchmarkSymbol,
		"symbols":           symbols,
	}
	runID, err := h.AgentRepo.CreateAgentRun(c.Request.Context(), userID, spec.BenchmarkSymbol, portfolioStrategyTypeTag, timeframe, "paper", initialCash, false, parameters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := h.Portfolio.StartForUser(c.Request.Context(), userID, runID, strategyKey, symbols, timeframe, initialCash); err != nil {
		_ = h.AgentRepo.MarkAgentRunFailed(c.Request.Context(), runID, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.AgentRepo.MarkAgentRunRunning(c.Request.Context(), runID); err != nil {
		_ = h.Portfolio.StopForUser(userID)
		_ = h.AgentRepo.MarkAgentRunFailed(c.Request.Context(), runID, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":            "activated",
		"run_id":            runID,
		"strategy_key":      strategyKey,
		"display_name":      spec.DisplayName,
		"deployment_status": string(spec.DeploymentStatus),
		"paper_only":        spec.PaperOnly,
		"symbols":           symbols,
	})
}

func (h *Handler) StopPortfolioAgent(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}
	if h.Portfolio == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "portfolio agent is not configured on this server"})
		return
	}

	if err := h.Portfolio.StopForUser(userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if err := h.AgentRepo.MarkActivePortfolioRunStopped(c.Request.Context(), userID, "user_requested"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.PortfolioRepo.InsertSystemAlert(c.Request.Context(), userID, "INFO", "Agent stopped", "Your portfolio agent was stopped.", "portfolio_agent", nil)

	c.JSON(http.StatusOK, gin.H{"status": "terminated"})
}

func (h *Handler) ListAgents(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}
	runs, err := h.AgentRepo.ListActiveAgentRuns(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"agents": runs})
}

func (h *Handler) GetStrategyCatalog(c *gin.Context) {
	if h.Portfolio == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "portfolio agent is not configured on this server"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"strategies":       h.Portfolio.Catalog(defaultPortfolioUniverse),
		"default_universe": defaultPortfolioUniverse,
		"recommended": gin.H{
			"conservative": riskProfileDefaultStrategy("conservative"),
			"moderate":     riskProfileDefaultStrategy("moderate"),
			"aggressive":   riskProfileDefaultStrategy("aggressive"),
		},
	})
}

func normalizeUniverse(symbols []string) []string {
	seen := make(map[string]struct{}, len(symbols))
	out := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" {
			continue
		}
		if _, ok := seen[symbol]; ok {
			continue
		}
		seen[symbol] = struct{}{}
		out = append(out, symbol)
	}
	return out
}

var _ = portfolio.StrategySpec{}
