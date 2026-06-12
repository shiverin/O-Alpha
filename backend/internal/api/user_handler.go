package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	agentportfolio "github.com/oalpha/internal/agent/portfolio"
)

type completeOnboardingRequest struct {
	RiskProfile      string `json:"risk_profile"`
	StrategyKey      string `json:"strategy_key"`
	BacktestAccepted bool   `json:"backtest_accepted"`
}

// CompleteUserOnboarding marks the authenticated user's onboarding as complete
// after risk settings and a catalog backtest acceptance have both been supplied.
func (h *Handler) CompleteUserOnboarding(c *gin.Context) {
	var req completeOnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "onboarding completion requires a JSON body"})
		return
	}
	riskProfile, strategyKey, err := validateOnboardingCompletion(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	var hasMatchingSettings bool
	const settingsQuery = `
		SELECT EXISTS (
			SELECT 1
			FROM agent_settings
			WHERE user_id = $1
			  AND risk_profile = $2
		 )`
	if err := h.repo.GetDB().QueryRow(c.Request.Context(), settingsQuery, userID, riskProfile).Scan(&hasMatchingSettings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to verify saved onboarding settings: %w", err).Error()})
		return
	}
	if !hasMatchingSettings {
		c.JSON(http.StatusConflict, gin.H{"error": "saved risk settings are required before onboarding can be completed"})
		return
	}

	const q = `
		UPDATE users 
		SET is_onboarded = true, 
		    updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC' 
		WHERE id = $1`

	tag, err := h.repo.GetDB().Exec(c.Request.Context(), q, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to complete onboarding status flip: %w", err).Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "authenticated user was not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "onboarding_finalized",
		"risk_profile": riskProfile,
		"strategy_key": strategyKey,
	})
}

func validateOnboardingCompletion(req completeOnboardingRequest) (string, string, error) {
	riskProfile := strings.ToLower(strings.TrimSpace(req.RiskProfile))
	strategyKey := strings.ToLower(strings.TrimSpace(req.StrategyKey))
	if !req.BacktestAccepted {
		return "", "", fmt.Errorf("accepted catalog backtest is required before onboarding can be completed")
	}
	if riskProfile != "conservative" && riskProfile != "moderate" && riskProfile != "aggressive" {
		return "", "", fmt.Errorf("risk_profile must be conservative, moderate, or aggressive")
	}
	if strategyKey == "" {
		return "", "", fmt.Errorf("strategy_key is required")
	}
	spec, err := agentportfolio.StrategySpecByKey(strategyKey, nil, agentportfolio.DefaultStrategyCatalogConfig())
	if err != nil {
		return "", "", err
	}
	if spec.RiskProfile != onboardingRiskBucket(riskProfile) {
		return "", "", fmt.Errorf("strategy_key %q does not match %s risk profile", strategyKey, riskProfile)
	}
	return riskProfile, strategyKey, nil
}

func onboardingRiskBucket(riskProfile string) agentportfolio.StrategyRiskProfile {
	switch riskProfile {
	case "conservative":
		return agentportfolio.StrategyRiskLow
	case "aggressive":
		return agentportfolio.StrategyRiskHigh
	default:
		return agentportfolio.StrategyRiskMedium
	}
}
