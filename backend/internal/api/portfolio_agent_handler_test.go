package api

import "testing"

func TestRiskProfileDefaultStrategyMatchesRiskBucket(t *testing.T) {
	cases := map[string]string{
		"conservative": "ranker_proxy_h63_low",
		"moderate":     "lgbm_ranker_h63_medium",
		"aggressive":   "composite_momentum_high",
		"":             "lgbm_ranker_h63_low",
	}

	for profile, want := range cases {
		if got := riskProfileDefaultStrategy(profile); got != want {
			t.Fatalf("riskProfileDefaultStrategy(%q)=%s, want %s", profile, got, want)
		}
	}
}

func TestDefaultPortfolioUniverseUsesRetainedYahoo100Panel(t *testing.T) {
	if len(defaultPortfolioUniverse) != 100 {
		t.Fatalf("default universe size=%d, want 100", len(defaultPortfolioUniverse))
	}
	if defaultPortfolioUniverse[0] != "VOO" {
		t.Fatalf("first default universe symbol=%s, want VOO", defaultPortfolioUniverse[0])
	}

	seen := make(map[string]bool, len(defaultPortfolioUniverse))
	for _, symbol := range defaultPortfolioUniverse {
		if seen[symbol] {
			t.Fatalf("duplicate symbol %s in default universe", symbol)
		}
		seen[symbol] = true
	}
	for _, symbol := range []string{"AAPL", "MSFT", "NVDA", "XOM"} {
		if !seen[symbol] {
			t.Fatalf("default universe missing %s", symbol)
		}
	}
}
