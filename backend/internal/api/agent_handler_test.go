package api

import (
	"testing"

	"github.com/oalpha/internal/agent"
	"github.com/oalpha/internal/db"
)

func TestParseRegimeModeDefaultsToOverlay(t *testing.T) {
	mode, label, err := parseRegimeMode("", nil)
	if err != nil {
		t.Fatalf("parse regime mode failed: %v", err)
	}
	if mode != agent.RegimeModeOverlay || label != "overlay" {
		t.Fatalf("expected overlay default, got mode=%s label=%s", mode, label)
	}
}

func TestParseRegimeModeCanDisableOverlay(t *testing.T) {
	enabled := false
	mode, label, err := parseRegimeMode("", &enabled)
	if err != nil {
		t.Fatalf("parse regime mode failed: %v", err)
	}
	if mode != agent.RegimeModeNone || label != "none" {
		t.Fatalf("expected none when overlay disabled, got mode=%s label=%s", mode, label)
	}
}

func TestParseRegimeModeAcceptsExplicitNone(t *testing.T) {
	mode, label, err := parseRegimeMode("none", nil)
	if err != nil {
		t.Fatalf("parse regime mode failed: %v", err)
	}
	if mode != agent.RegimeModeNone || label != "none" {
		t.Fatalf("expected explicit none, got mode=%s label=%s", mode, label)
	}
}

func TestValidateRiskProfileChangeRequiresAcceptedBacktest(t *testing.T) {
	err := validateRiskProfileChangeBacktest(
		&db.AgentSettings{RiskProfile: "moderate"},
		&db.AgentSettings{RiskProfile: "aggressive"},
		saveAgentSettingsRequest{RiskProfile: "aggressive"},
	)
	if err == nil {
		t.Fatalf("expected missing accepted backtest error")
	}
}

func TestValidateRiskProfileChangeRejectsStrategyMismatch(t *testing.T) {
	err := validateRiskProfileChangeBacktest(
		&db.AgentSettings{RiskProfile: "moderate"},
		&db.AgentSettings{RiskProfile: "aggressive"},
		saveAgentSettingsRequest{
			RiskProfile: "aggressive",
			StrategyKey: "ranker_proxy_h63_low",
			BacktestOK:  true,
		},
	)
	if err == nil {
		t.Fatalf("expected strategy/profile mismatch error")
	}
}

func TestValidateRiskProfileChangeAllowsMatchingAcceptedStrategy(t *testing.T) {
	err := validateRiskProfileChangeBacktest(
		&db.AgentSettings{RiskProfile: "moderate"},
		&db.AgentSettings{RiskProfile: "aggressive"},
		saveAgentSettingsRequest{
			RiskProfile: "aggressive",
			StrategyKey: "composite_momentum_high",
			BacktestOK:  true,
		},
	)
	if err != nil {
		t.Fatalf("validateRiskProfileChangeBacktest: %v", err)
	}
}

func TestValidateRiskProfileChangeAllowsInitialSettingsCreate(t *testing.T) {
	err := validateRiskProfileChangeBacktest(
		nil,
		&db.AgentSettings{RiskProfile: "moderate"},
		saveAgentSettingsRequest{RiskProfile: "moderate"},
	)
	if err != nil {
		t.Fatalf("initial settings create should not require separate risk-change validation: %v", err)
	}
}

func TestAgentSettingsFromRequestNormalizesStringFields(t *testing.T) {
	settings := agentSettingsFromRequest(42, saveAgentSettingsRequest{
		RiskProfile:   " Moderate ",
		Leverage:      2,
		MaxPositions:  6,
		StopLossPct:   2.5,
		TakeProfitPct: 5,
		RebalanceFreq: " Daily ",
	})

	if settings.UserID != 42 {
		t.Fatalf("userID=%d, want 42", settings.UserID)
	}
	if settings.RiskProfile != "moderate" {
		t.Fatalf("riskProfile=%q, want moderate", settings.RiskProfile)
	}
	if settings.RebalanceFreq != "daily" {
		t.Fatalf("rebalance=%q, want daily", settings.RebalanceFreq)
	}
}
