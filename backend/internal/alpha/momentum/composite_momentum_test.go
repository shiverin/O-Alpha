package momentum

import (
	"context"
	"math"
	"testing"

	"github.com/oalpha/internal/backtest"
)

func TestCompositeMomentumUsesBenchmarkCoreBeforeLookback(t *testing.T) {
	panel := compositeMomentumTestPanel(40)
	strategy := NewCompositeMomentumStrategy(panel.Symbols, DefaultCompositeMomentumConfig())

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 5))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	target := output.Targets["VOO"]
	if math.Abs(target.TargetWeight-1) > 1e-9 {
		t.Fatalf("VOO target=%f, want benchmark core 1.0 before lookback", target.TargetWeight)
	}
}

func TestCompositeMomentumVolCapExcludesHighVolETF(t *testing.T) {
	panel := compositeMomentumTestPanel(180)
	cfg := DefaultCompositeMomentumConfig()
	cfg.Legs = cfg.Legs[:1]
	strategy := NewCompositeMomentumStrategy(panel.Symbols, cfg)

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 150))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if _, ok := output.Targets["SMH"]; ok {
		t.Fatalf("SMH should be excluded by the ETF leg volatility cap")
	}
	if output.Targets["XLU"].TargetWeight <= 0 {
		t.Fatalf("XLU should receive ETF sleeve weight, targets=%v", output.Targets)
	}
	if output.Targets["VOO"].TargetWeight <= 0.70 || output.Targets["VOO"].TargetWeight >= 1.0 {
		t.Fatalf("VOO residual target=%f, want benchmark core residual", output.Targets["VOO"].TargetWeight)
	}
	if output.GrossExposure > 1.0+1e-9 || output.NetExposure > 1.0+1e-9 {
		t.Fatalf("exposures gross=%f net=%f exceed long-only budget", output.GrossExposure, output.NetExposure)
	}
}

func TestCompositeMomentumLiquidityFilterExcludesThinCandidate(t *testing.T) {
	panel := testMomentumPanel(90, map[string]func(int) float64{
		"VOO": func(i int) float64 {
			return 100 * math.Exp(0.0005*float64(i))
		},
		"THIN": func(i int) float64 {
			return 80 * math.Exp(0.0040*float64(i))
		},
		"LIQ": func(i int) float64 {
			return 80 * math.Exp(0.0020*float64(i))
		},
	})
	for i := range panel.Bars["THIN"] {
		panel.Bars["THIN"][i].Volume = 1_000
		panel.Bars["LIQ"][i].Volume = 1_000_000
	}
	cfg := DefaultCompositeMomentumConfig()
	cfg.GlobalMaxNameWeight = 0.20
	cfg.Legs = []CompositeMomentumLegConfig{
		{
			Name:                     "liquidity_gated",
			CandidateUniverse:        "all",
			RankMode:                 "relative_momentum",
			LookbackBars:             21,
			SleeveFraction:           0.20,
			TopK:                     1,
			MaxNameWeight:            0.20,
			MinRelativeMomentum:      0,
			DollarVolumeLookbackBars: 21,
			MinMedianDollarVolume:    10_000_000,
		},
	}
	strategy := NewCompositeMomentumStrategy(panel.Symbols, cfg)

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 70))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if _, ok := output.Targets["THIN"]; ok {
		t.Fatalf("THIN should be excluded by median dollar-volume filter")
	}
	if got := output.Targets["LIQ"].TargetWeight; math.Abs(got-0.20) > 1e-9 {
		t.Fatalf("LIQ target=%f, want liquidity-qualified sleeve 0.20; targets=%v", got, output.Targets)
	}
}

func TestCompositeMomentumHoldsTargetsBetweenRebalances(t *testing.T) {
	panel := compositeMomentumTestPanel(180)
	strategy := NewCompositeMomentumStrategy(panel.Symbols, DefaultCompositeMomentumConfig())
	first, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 150))
	if err != nil {
		t.Fatalf("first EvaluatePortfolioLatest: %v", err)
	}
	if len(first.Targets) == 0 {
		t.Fatalf("expected first call to rebalance")
	}
	second, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 151))
	if err != nil {
		t.Fatalf("second EvaluatePortfolioLatest: %v", err)
	}
	if len(second.Targets) != 0 || second.EngineMetadata["action"] != actionHoldTargets {
		t.Fatalf("expected hold-target output between rebalances, got targets=%d metadata=%v", len(second.Targets), second.EngineMetadata)
	}
}

func TestCompositeMomentumRiskOffUsesDefensiveSleeve(t *testing.T) {
	panel := testMomentumPanel(80, map[string]func(int) float64{
		"VOO": func(i int) float64 {
			return 100 * math.Exp(-0.002*float64(i))
		},
		"XLU": func(i int) float64 {
			return 75 * math.Exp(0.001*float64(i))
		},
		"XLP": func(i int) float64 {
			return 70 * math.Exp(0.0005*float64(i))
		},
		"QQQ": func(i int) float64 {
			return 90 * math.Exp(0.003*float64(i))
		},
	})
	cfg := DefaultCompositeMomentumConfig()
	cfg.BenchmarkTrendLookbackBars = 20
	cfg.MinBenchmarkTrend = 0
	cfg.RiskOffBenchmarkWeight = 0.25
	cfg.GlobalMaxNameWeight = 0.75
	cfg.RiskOffLegs = []CompositeMomentumLegConfig{
		{
			Name:                "defensive",
			CandidateSymbols:    []string{"XLU", "XLP"},
			LookbackBars:        21,
			SleeveFraction:      0.75,
			TopK:                1,
			MaxNameWeight:       0.75,
			MinRelativeMomentum: -0.02,
		},
	}
	strategy := NewCompositeMomentumStrategy(panel.Symbols, cfg)

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 70))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if output.EngineMetadata["risk_off"] != true {
		t.Fatalf("expected risk-off metadata, got %v", output.EngineMetadata)
	}
	if got := output.Targets["VOO"].TargetWeight; math.Abs(got-0.25) > 1e-9 {
		t.Fatalf("VOO target=%f, want fixed risk-off benchmark weight 0.25", got)
	}
	if got := output.Targets["XLU"].TargetWeight; math.Abs(got-0.75) > 1e-9 {
		t.Fatalf("XLU defensive target=%f, want 0.75", got)
	}
	if _, ok := output.Targets["QQQ"]; ok {
		t.Fatalf("QQQ should not be eligible for the risk-off defensive sleeve")
	}
}

func TestCompositeMomentumLowVolRankModeSelectsLowerVolatility(t *testing.T) {
	panel := testMomentumPanel(80, map[string]func(int) float64{
		"VOO": func(i int) float64 {
			return 100 * math.Exp(0.0005*float64(i))
		},
		"LOWV": func(i int) float64 {
			return 60 * math.Exp(0.001*float64(i))
		},
		"HIGHV": func(i int) float64 {
			return 60 * math.Exp(0.001*float64(i)) * (1 + 0.08*math.Sin(float64(i)))
		},
	})
	cfg := DefaultCompositeMomentumConfig()
	cfg.Legs = []CompositeMomentumLegConfig{
		{
			Name:                "low_vol",
			CandidateUniverse:   "all",
			RankMode:            "low_vol",
			LookbackBars:        21,
			SleeveFraction:      0.20,
			TopK:                1,
			MaxNameWeight:       0.20,
			MinRelativeMomentum: -1,
		},
	}
	strategy := NewCompositeMomentumStrategy(panel.Symbols, cfg)

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 70))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if got := output.Targets["LOWV"].TargetWeight; math.Abs(got-0.20) > 1e-9 {
		t.Fatalf("LOWV target=%f, want low-vol sleeve 0.20; targets=%v", got, output.Targets)
	}
	if _, ok := output.Targets["HIGHV"]; ok {
		t.Fatalf("HIGHV should not be selected by low-vol rank mode")
	}
}

func TestCompositeMomentumMeanReversionRankModeSelectsRelativeLoser(t *testing.T) {
	panel := testMomentumPanel(80, map[string]func(int) float64{
		"VOO": func(i int) float64 {
			return 100 * math.Exp(0.001*float64(i))
		},
		"LAG": func(i int) float64 {
			return 60 * math.Exp(0.0001*float64(i))
		},
		"LEAD": func(i int) float64 {
			return 60 * math.Exp(0.003*float64(i))
		},
	})
	cfg := DefaultCompositeMomentumConfig()
	cfg.Legs = []CompositeMomentumLegConfig{
		{
			Name:                "reversal",
			CandidateUniverse:   "all",
			RankMode:            "mean_reversion",
			LookbackBars:        21,
			SleeveFraction:      0.20,
			TopK:                1,
			MaxNameWeight:       0.20,
			MinRelativeMomentum: -1,
		},
	}
	strategy := NewCompositeMomentumStrategy(panel.Symbols, cfg)

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 70))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if got := output.Targets["LAG"].TargetWeight; math.Abs(got-0.20) > 1e-9 {
		t.Fatalf("LAG target=%f, want reversal sleeve 0.20; targets=%v", got, output.Targets)
	}
	if _, ok := output.Targets["LEAD"]; ok {
		t.Fatalf("LEAD should not be selected by mean-reversion rank mode")
	}
}

func TestCompositeMomentumRiskAdjustedEdgeWeightsByConviction(t *testing.T) {
	panel := testMomentumPanel(90, map[string]func(int) float64{
		"VOO": func(i int) float64 {
			return 100 * math.Exp(0.0005*float64(i))
		},
		"FAST": func(i int) float64 {
			return 80 * math.Exp(0.0030*float64(i))
		},
		"SLOW": func(i int) float64 {
			return 80 * math.Exp(0.0018*float64(i))
		},
	})
	cfg := DefaultCompositeMomentumConfig()
	cfg.GlobalMaxNameWeight = 0.30
	cfg.Legs = []CompositeMomentumLegConfig{
		{
			Name:                "risk_weighted",
			CandidateUniverse:   "all",
			RankMode:            "vol_adjusted_momentum",
			WeightMode:          "risk_adjusted_edge",
			LookbackBars:        21,
			SleeveFraction:      0.30,
			TopK:                2,
			MaxNameWeight:       0.30,
			MinRelativeMomentum: 0,
			EdgeExponent:        2,
			VolFloor:            0.05,
		},
	}
	strategy := NewCompositeMomentumStrategy(panel.Symbols, cfg)

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 70))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	fast := output.Targets["FAST"].TargetWeight
	slow := output.Targets["SLOW"].TargetWeight
	if fast <= slow {
		t.Fatalf("FAST target=%f should exceed SLOW target=%f under risk-adjusted edge weighting", fast, slow)
	}
	if got := output.Targets["VOO"].TargetWeight; math.Abs(got+fast+slow-1) > 1e-9 {
		t.Fatalf("targets should remain fully benchmark-funded, VOO=%f FAST=%f SLOW=%f", got, fast, slow)
	}
}

func TestCompositeMomentumTurnoverBandHoldsPriorTargets(t *testing.T) {
	panel := testMomentumPanel(90, map[string]func(int) float64{
		"VOO": func(i int) float64 {
			return 100 * math.Exp(0.0005*float64(i))
		},
		"A": func(i int) float64 {
			return 80 * math.Exp(0.0020*float64(i))
		},
		"B": func(i int) float64 {
			return 80 * math.Exp(0.0018*float64(i))
		},
	})
	cfg := DefaultCompositeMomentumConfig()
	cfg.RebalanceEveryBars = 1
	cfg.TurnoverBand = 0.50
	cfg.Legs = []CompositeMomentumLegConfig{
		{
			Name:                "risk_weighted",
			CandidateUniverse:   "all",
			RankMode:            "vol_adjusted_momentum",
			WeightMode:          "risk_adjusted_edge",
			LookbackBars:        21,
			SleeveFraction:      0.20,
			TopK:                2,
			MaxNameWeight:       0.20,
			MinRelativeMomentum: 0,
			EdgeExponent:        2,
			VolFloor:            0.05,
		},
	}
	strategy := NewCompositeMomentumStrategy(panel.Symbols, cfg)
	first, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 70))
	if err != nil {
		t.Fatalf("first EvaluatePortfolioLatest: %v", err)
	}
	second, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 71))
	if err != nil {
		t.Fatalf("second EvaluatePortfolioLatest: %v", err)
	}
	if second.EngineMetadata["action"] != actionHoldTargets {
		t.Fatalf("expected turnover band to hold prior targets, metadata=%v", second.EngineMetadata)
	}
	if second.Targets["A"].TargetWeight != first.Targets["A"].TargetWeight {
		t.Fatalf("expected held A target, first=%f second=%f", first.Targets["A"].TargetWeight, second.Targets["A"].TargetWeight)
	}
}

func compositeMomentumTestPanel(n int) backtest.AlignedBars {
	return testMomentumPanel(n, map[string]func(int) float64{
		"VOO": func(i int) float64 {
			return 100 * math.Exp(0.0005*float64(i))
		},
		"XLU": func(i int) float64 {
			return 80 * math.Exp(0.0031*float64(i))
		},
		"SMH": func(i int) float64 {
			wobble := 1 + 0.08*math.Sin(float64(i))
			if wobble < 0.50 {
				wobble = 0.50
			}
			return 70 * math.Exp(0.005*float64(i)) * wobble
		},
		"AAPL": func(i int) float64 {
			return 60 * math.Exp(0.0021*float64(i))
		},
		"MSFT": func(i int) float64 {
			return 65 * math.Exp(0.0019*float64(i))
		},
	})
}
