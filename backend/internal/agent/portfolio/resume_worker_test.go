package portfolio

import (
	"testing"
	"time"

	"github.com/oalpha/internal/db"
)

func TestSymbolsFromAgentRunNormalizesStringSlices(t *testing.T) {
	got := symbolsFromAgentRun(db.AgentRunSummary{
		Parameters: map[string]interface{}{
			"symbols": []interface{}{"aapl", " MSFT ", "", "aapl"},
		},
	})
	want := []string{"AAPL", "MSFT"}
	if len(got) != len(want) {
		t.Fatalf("symbols length mismatch: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("symbol[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLastRebalanceFromRuntimeStateParsesSettingsMetadata(t *testing.T) {
	want := time.Date(2026, 6, 16, 3, 15, 0, 0, time.UTC)
	got, ok := lastRebalanceFromRuntimeState(map[string]interface{}{
		runtimeSettingsMetadataKey: map[string]interface{}{
			"last_rebalance_at": want.Format(time.RFC3339Nano),
		},
	})
	if !ok {
		t.Fatalf("last rebalance was not found")
	}
	if !got.Equal(want) {
		t.Fatalf("last rebalance = %s, want %s", got, want)
	}
}

func TestAnnotateLastRebalanceStoresNextEligibleTime(t *testing.T) {
	state := map[string]interface{}{}
	last := time.Date(2026, 6, 16, 3, 15, 0, 0, time.UTC)
	settings := DefaultRuntimeSettings(last)
	settings.RebalanceFreq = "daily"

	annotateLastRebalance(state, last, settings)

	applied, ok := state[runtimeSettingsMetadataKey].(map[string]interface{})
	if !ok {
		t.Fatalf("settings metadata missing: %+v", state)
	}
	gotLast, ok := timeFromRuntimeValue(applied["last_rebalance_at"])
	if !ok || !gotLast.Equal(last) {
		t.Fatalf("last rebalance = %s, want %s", gotLast, last)
	}
	gotNext, ok := timeFromRuntimeValue(applied["next_eligible_rebalance_at"])
	if !ok || !gotNext.Equal(last.Add(24*time.Hour)) {
		t.Fatalf("next eligible = %s, want %s", gotNext, last.Add(24*time.Hour))
	}
}
