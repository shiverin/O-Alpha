package momentum

import "testing"

func TestEvaluateUniverseAppliesPriceLiquidityAndCompletenessFilters(t *testing.T) {
	panel := testMomentumPanel(40, map[string]func(int) float64{
		"GOOD":  func(i int) float64 { return 50 + float64(i)*0.1 },
		"CHEAP": func(i int) float64 { return 2 },
		"MISS":  func(i int) float64 { return 40 + float64(i)*0.1 },
	})
	panel.Bars["MISS"][35].Close = 0
	panel.Bars["MISS"][36].Close = 0
	cfg := testMomentumConfig()
	cfg.FormationDays = 20
	cfg.SkipDays = 5
	cfg.MinPrice = 5
	cfg.MinMedianDollarVolume = 10_000
	cfg.MinDataCompleteness = 0.98

	candidates := EvaluateUniverse(panel, 39, []string{"GOOD", "CHEAP", "MISS"}, cfg)
	reasons := make(map[string]string)
	eligible := make(map[string]bool)
	for _, candidate := range candidates {
		reasons[candidate.Symbol] = candidate.Reason
		eligible[candidate.Symbol] = candidate.Eligible
	}
	if !eligible["GOOD"] {
		t.Fatalf("GOOD should be eligible: %+v", candidates)
	}
	if reasons["CHEAP"] != "price_below_min" {
		t.Fatalf("CHEAP reason=%s, want price_below_min", reasons["CHEAP"])
	}
	if reasons["MISS"] != "insufficient_data_completeness" {
		t.Fatalf("MISS reason=%s, want insufficient_data_completeness", reasons["MISS"])
	}
}
