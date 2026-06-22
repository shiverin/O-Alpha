package portfolio

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/oalpha/internal/db"
)

func TestLastRebalanceFromRuntimeState(t *testing.T) {
	want := time.Date(2026, 6, 12, 4, 0, 0, 123, time.UTC)

	got, ok := lastRebalanceFromRuntimeState(map[string]interface{}{
		"last_rebalance_at": want.Format(time.RFC3339Nano),
	})
	if !ok || !got.Equal(want) {
		t.Fatalf("top-level last rebalance=%v ok=%v, want %v true", got, ok, want)
	}

	got, ok = lastRebalanceFromRuntimeState(map[string]interface{}{
		runtimeSettingsMetadataKey: map[string]interface{}{
			"last_rebalance_at": want.Format(time.RFC3339Nano),
		},
	})
	if !ok || !got.Equal(want) {
		t.Fatalf("nested last rebalance=%v ok=%v, want %v true", got, ok, want)
	}
}

func TestAnnotateLastRebalanceWritesRuntimeState(t *testing.T) {
	last := time.Date(2026, 6, 12, 4, 0, 0, 0, time.UTC)
	state := map[string]interface{}{
		"source": "hmm_risk_overlay",
	}
	settings := RuntimeSettings{
		RiskProfile:        "moderate",
		MaxGrossExposure:   1,
		MaxActivePositions: 5,
		StopLossPct:        2,
		TakeProfitPct:      4,
		RebalanceFreq:      "daily",
		LoadedAt:           last,
		Source:             "agent_settings",
	}

	annotateLastRebalance(state, last, settings)

	gotLast, ok := state["last_rebalance_at"].(time.Time)
	if !ok || !gotLast.Equal(last) {
		t.Fatalf("last_rebalance_at=%#v, want %v", state["last_rebalance_at"], last)
	}
	settingsState, ok := state[runtimeSettingsMetadataKey].(map[string]interface{})
	if !ok {
		t.Fatalf("settings state missing: %#v", state[runtimeSettingsMetadataKey])
	}
	nestedLast, ok := settingsState["last_rebalance_at"].(time.Time)
	if !ok || !nestedLast.Equal(last) {
		t.Fatalf("nested last_rebalance_at=%#v, want %v", settingsState["last_rebalance_at"], last)
	}
	next, ok := settingsState["next_eligible_rebalance_at"].(time.Time)
	if !ok || !next.Equal(last.Add(24*time.Hour)) {
		t.Fatalf("next eligible=%#v, want %v", settingsState["next_eligible_rebalance_at"], last.Add(24*time.Hour))
	}
}

func TestSymbolsFromAgentRun(t *testing.T) {
	run := db.AgentRunSummary{
		Symbol: "VOO",
		Parameters: map[string]interface{}{
			"symbols": []interface{}{" voo ", "AAPL", "VOO", "", "msft"},
		},
	}

	got := symbolsFromAgentRun(run)
	want := []string{"VOO", "AAPL", "MSFT"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("symbols=%v, want %v", got, want)
	}

	got = symbolsFromAgentRun(db.AgentRunSummary{Symbol: "spy"})
	want = []string{"SPY"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fallback symbols=%v, want %v", got, want)
	}
}

func TestResumeLastRebalanceFallsBackToHeartbeat(t *testing.T) {
	heartbeat := time.Date(2026, 6, 15, 8, 43, 9, 0, time.UTC)
	orchestrator := &PortfolioOrchestrator{}

	got := orchestrator.resumeLastRebalance(context.Background(), db.AgentRunSummary{
		ID:              27,
		UserID:          2,
		LastHeartbeatAt: &heartbeat,
	})
	if !got.Equal(heartbeat) {
		t.Fatalf("fallback last rebalance=%v, want heartbeat %v", got, heartbeat)
	}
}
