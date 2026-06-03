package momentum

import (
	"testing"

	"github.com/oalpha/internal/backtest"
)

func TestBuildMomentumTargetsAppliesSymbolAndSectorCaps(t *testing.T) {
	panel := testMomentumPanel(80, map[string]func(int) float64{
		"AAA": func(i int) float64 { return 100 + float64(i)*0.5 },
		"BBB": func(i int) float64 { return 90 + float64(i)*0.4 },
		"CCC": func(i int) float64 { return 80 + float64(i)*0.3 },
	})
	cfg := testMomentumConfig()
	cfg.MaxSymbolWeight = 0.20
	cfg.MaxSectorWeight = 0.30
	cfg.TargetVolAnnual = 1.0
	selected := []MomentumScore{
		{Symbol: "AAA", Rank: 1, Score: 3, FormationReturn: 0.2, RealizedVol: 0.1},
		{Symbol: "BBB", Rank: 2, Score: 2, FormationReturn: 0.1, RealizedVol: 0.1},
		{Symbol: "CCC", Rank: 3, Score: 1, FormationReturn: 0.1, RealizedVol: 0.1},
	}

	targets, _ := BuildMomentumTargets(panel, 70, selected, cfg, map[string]string{
		"AAA": "tech",
		"BBB": "tech",
		"CCC": "health",
	})
	assertWeightCap(t, targets, 0.20)
	sectorWeights := map[string]float64{
		"tech":   targets["AAA"].TargetWeight + targets["BBB"].TargetWeight,
		"health": targets["CCC"].TargetWeight,
	}
	if sectorWeights["tech"] > 0.30+1e-9 {
		t.Fatalf("tech sector weight=%f exceeds cap", sectorWeights["tech"])
	}
}

func assertWeightCap(t *testing.T, targets map[string]backtest.TargetPosition, cap float64) {
	t.Helper()
	for symbol, target := range targets {
		if target.TargetWeight > cap+1e-9 {
			t.Fatalf("%s target weight=%f exceeds cap=%f", symbol, target.TargetWeight, cap)
		}
	}
}
