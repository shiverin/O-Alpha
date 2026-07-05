package db

import "testing"

func TestShouldBlockAgentSettingsSaveRejectsChangeWithActiveRun(t *testing.T) {
	current := baselineAgentSettings()
	next := baselineAgentSettings()
	next.Leverage = 3

	if !ShouldBlockAgentSettingsSave(current, next, true) {
		t.Fatalf("expected active run to block leverage change")
	}
}

func TestShouldBlockAgentSettingsSaveRejectsFirstSettingsCreateWithActiveRun(t *testing.T) {
	if !ShouldBlockAgentSettingsSave(nil, baselineAgentSettings(), true) {
		t.Fatalf("expected active run to block first settings create")
	}
}

func TestShouldBlockAgentSettingsSaveAllowsNoopWithActiveRun(t *testing.T) {
	current := baselineAgentSettings()
	next := baselineAgentSettings()
	next.RiskProfile = " Moderate "
	next.RebalanceFreq = " Daily "

	if ShouldBlockAgentSettingsSave(current, next, true) {
		t.Fatalf("expected exact settings no-op to be allowed while active")
	}
}

func TestShouldBlockAgentSettingsSaveAllowsChangeWithoutActiveRun(t *testing.T) {
	current := baselineAgentSettings()
	next := baselineAgentSettings()
	next.MaxPositions = 12

	if ShouldBlockAgentSettingsSave(current, next, false) {
		t.Fatalf("expected inactive run state to allow settings change")
	}
}

func TestAgentSettingsChangedCoversGuardedFields(t *testing.T) {
	cases := map[string]func(*AgentSettings){
		"risk profile":   func(s *AgentSettings) { s.RiskProfile = "aggressive" },
		"leverage":       func(s *AgentSettings) { s.Leverage = 4 },
		"max positions":  func(s *AgentSettings) { s.MaxPositions = 9 },
		"stop loss":      func(s *AgentSettings) { s.StopLossPct = 3.5 },
		"take profit":    func(s *AgentSettings) { s.TakeProfitPct = 6.5 },
		"rebalance freq": func(s *AgentSettings) { s.RebalanceFreq = "weekly" },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			current := baselineAgentSettings()
			next := baselineAgentSettings()
			mutate(next)
			if !AgentSettingsChanged(current, next) {
				t.Fatalf("expected %s change to be detected", name)
			}
		})
	}
}

func baselineAgentSettings() *AgentSettings {
	return &AgentSettings{
		UserID:        7,
		RiskProfile:   "moderate",
		Leverage:      2,
		MaxPositions:  6,
		StopLossPct:   2.5,
		TakeProfitPct: 5,
		RebalanceFreq: "daily",
	}
}
