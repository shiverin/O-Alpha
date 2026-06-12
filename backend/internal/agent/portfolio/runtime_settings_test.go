package portfolio

import (
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
)

func TestRuntimeSettingsFromAgentSettingsMapsSavedControls(t *testing.T) {
	now := time.Date(2026, 6, 12, 4, 0, 0, 0, time.UTC)
	settings := RuntimeSettingsFromAgentSettings(&db.AgentSettings{
		RiskProfile:   "aggressive",
		Leverage:      3,
		MaxPositions:  12,
		StopLossPct:   4.5,
		TakeProfitPct: 9,
		RebalanceFreq: "hourly",
	}, now)

	if settings.RiskProfile != "aggressive" {
		t.Fatalf("risk profile=%s, want aggressive", settings.RiskProfile)
	}
	if settings.MaxGrossExposure != 3 {
		t.Fatalf("max gross=%.2f, want 3", settings.MaxGrossExposure)
	}
	if settings.MaxActivePositions != 12 {
		t.Fatalf("max active=%d, want 12", settings.MaxActivePositions)
	}
	if settings.StopLossPct != 4.5 || settings.TakeProfitPct != 9 {
		t.Fatalf("exit settings=(%.2f, %.2f), want (4.5, 9)", settings.StopLossPct, settings.TakeProfitPct)
	}
	if settings.RebalanceFreq != "hourly" {
		t.Fatalf("rebalance=%s, want hourly", settings.RebalanceFreq)
	}
	if settings.Source != "agent_settings" {
		t.Fatalf("source=%s, want agent_settings", settings.Source)
	}
}

func TestDefaultRuntimeSettingsWhenMissing(t *testing.T) {
	settings := RuntimeSettingsFromAgentSettings(nil, time.Time{})
	if settings.RiskProfile != "moderate" {
		t.Fatalf("risk profile=%s, want moderate", settings.RiskProfile)
	}
	if settings.MaxGrossExposure != 1 {
		t.Fatalf("max gross=%.2f, want 1", settings.MaxGrossExposure)
	}
	if settings.MaxActivePositions != 5 {
		t.Fatalf("max active=%d, want 5", settings.MaxActivePositions)
	}
	if settings.RebalanceFreq != "daily" {
		t.Fatalf("rebalance=%s, want daily", settings.RebalanceFreq)
	}
	if settings.Source != "default" {
		t.Fatalf("source=%s, want default", settings.Source)
	}
}

func TestRuntimeSettingsRebalanceDue(t *testing.T) {
	settings := RuntimeSettings{RebalanceFreq: "daily"}.withBounds()
	last := time.Date(2026, 6, 10, 14, 0, 0, 0, time.UTC)
	if settings.RebalanceDue(last.Add(23*time.Hour), last) {
		t.Fatalf("daily rebalance should not be due before 24h")
	}
	if !settings.RebalanceDue(last.Add(24*time.Hour), last) {
		t.Fatalf("daily rebalance should be due at 24h")
	}
}

func TestApplyRuntimeSettingsCapsPositionsAndGrossExposure(t *testing.T) {
	t0 := time.Date(2026, 6, 12, 14, 0, 0, 0, time.UTC)
	output := backtest.PortfolioOutput{
		Time: t0,
		Targets: map[string]backtest.TargetPosition{
			"VOO":  {Symbol: "VOO", TargetWeight: 0.80, Side: backtest.PositionSideLong},
			"AAPL": {Symbol: "AAPL", TargetWeight: 0.40, AlphaScore: 0.10, Side: backtest.PositionSideLong},
			"MSFT": {Symbol: "MSFT", TargetWeight: 0.30, AlphaScore: 0.20, Side: backtest.PositionSideLong},
			"NVDA": {Symbol: "NVDA", TargetWeight: 0.20, AlphaScore: 0.30, Side: backtest.PositionSideLong},
		},
	}

	settings := RuntimeSettings{
		MaxGrossExposure:   1,
		MaxActivePositions: 2,
		RebalanceFreq:      "daily",
		LoadedAt:           t0,
		Source:             "unit",
	}.withBounds()

	got := applyRuntimeSettingsToOutput(output, "VOO", settings, true, time.Time{})
	if _, ok := got.Targets["NVDA"]; ok {
		t.Fatalf("lowest-weight active target should have been removed: %+v", got.Targets)
	}
	if len(got.Targets) != 3 {
		t.Fatalf("target count=%d, want 3 including benchmark", len(got.Targets))
	}
	if got.GrossExposure > 1.0000001 {
		t.Fatalf("gross exposure %.8f exceeds cap", got.GrossExposure)
	}
	if got.Targets["VOO"].TargetWeight >= 0.80 {
		t.Fatalf("expected VOO target to be scaled under gross cap, got %.4f", got.Targets["VOO"].TargetWeight)
	}
	if got.EngineMetadata[runtimeCadenceMetadataKey] != true {
		t.Fatalf("missing rebalance due metadata: %+v", got.EngineMetadata)
	}
}

func TestApplyRuntimeSettingsSuppressesTargetsWhenCadenceNotDue(t *testing.T) {
	t0 := time.Date(2026, 6, 12, 14, 0, 0, 0, time.UTC)
	output := backtest.PortfolioOutput{
		Time: t0,
		Targets: map[string]backtest.TargetPosition{
			"VOO": {Symbol: "VOO", TargetWeight: 1, Side: backtest.PositionSideLong},
		},
	}
	settings := DefaultRuntimeSettings(t0)
	got := applyRuntimeSettingsToOutput(output, "VOO", settings, false, t0.Add(-time.Hour))
	if len(got.Targets) != 0 {
		t.Fatalf("targets should be suppressed when cadence is not due: %+v", got.Targets)
	}
	if got.EngineMetadata[runtimeCadenceMetadataKey] != false {
		t.Fatalf("expected rebalance_due=false metadata: %+v", got.EngineMetadata)
	}
	if got.EngineMetadata[runtimeSuppressedMetadataKey] == nil {
		t.Fatalf("expected suppression reason metadata: %+v", got.EngineMetadata)
	}
}

func TestRiskExitReasonTriggersStopLossAndTakeProfit(t *testing.T) {
	settings := RuntimeSettings{StopLossPct: 2.5, TakeProfitPct: 5}
	if got := riskExitReason(-2.6, settings); got != "stop_loss" {
		t.Fatalf("stop-loss reason=%q, want stop_loss", got)
	}
	if got := riskExitReason(5.1, settings); got != "take_profit" {
		t.Fatalf("take-profit reason=%q, want take_profit", got)
	}
	if got := riskExitReason(1.2, settings); got != "" {
		t.Fatalf("neutral reason=%q, want empty", got)
	}
}

func TestShouldRecordRebalanceRequiresActualTargets(t *testing.T) {
	if shouldRecordRebalance(backtest.PortfolioOutput{}, true) {
		t.Fatalf("empty output should not consume the rebalance cadence")
	}
	if shouldRecordRebalance(backtest.PortfolioOutput{
		Targets: map[string]backtest.TargetPosition{
			"VOO": {Symbol: "VOO", TargetWeight: 1, Side: backtest.PositionSideLong},
		},
	}, false) {
		t.Fatalf("not-due output should not record a rebalance")
	}
	if !shouldRecordRebalance(backtest.PortfolioOutput{
		Targets: map[string]backtest.TargetPosition{
			"VOO": {Symbol: "VOO", TargetWeight: 1, Side: backtest.PositionSideLong},
		},
	}, true) {
		t.Fatalf("due output with targets should record a rebalance")
	}
}
