package momentum

import (
	"context"
	"testing"
)

func TestCrossSectionalMomentumRebalancesMonthly(t *testing.T) {
	panel := testMomentumPanel(100, map[string]func(int) float64{
		"AAA": func(i int) float64 { return 100 + float64(i)*0.5 },
		"BBB": func(i int) float64 { return 100 + float64(i)*0.3 + float64(i%3)*0.2 },
		"CCC": func(i int) float64 { return 100 + float64(i)*0.2 + float64(i%5)*0.2 },
	})
	cfg := testMomentumConfig()
	cfg.RebalanceFrequency = RebalanceMonthly
	cfg.TopFraction = 1
	cfg.TargetVolAnnual = 1.0
	strategy := NewCrossSectionalMomentumStrategy(panel.Symbols, cfg, nil)

	outputs, err := strategy.GeneratePortfolioSignals(context.Background(), panel)
	if err != nil {
		t.Fatalf("GeneratePortfolioSignals: %v", err)
	}
	rebalancedMonths := make(map[string]bool)
	for _, output := range outputs {
		if len(output.Targets) == 0 {
			continue
		}
		key := output.Time.Format("2006-01")
		if rebalancedMonths[key] {
			t.Fatalf("multiple rebalances in month %s", key)
		}
		rebalancedMonths[key] = true
		if output.EngineMetadata["rebalance"] != true {
			t.Fatalf("target output missing rebalance metadata")
		}
	}
	if len(rebalancedMonths) < 2 {
		t.Fatalf("monthly rebalances=%d, want at least 2", len(rebalancedMonths))
	}
}

func TestCrossSectionalMomentumHoldTargetsBetweenRebalances(t *testing.T) {
	panel := testMomentumPanel(40, map[string]func(int) float64{
		"AAA": func(i int) float64 { return 100 + float64(i)*0.5 },
		"BBB": func(i int) float64 { return 100 + float64(i)*0.3 + float64(i%3)*0.2 },
	})
	cfg := testMomentumConfig()
	cfg.RebalanceFrequency = RebalanceMonthly
	strategy := NewCrossSectionalMomentumStrategy(panel.Symbols, cfg, nil)

	first, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 27))
	if err != nil {
		t.Fatalf("first EvaluatePortfolioLatest: %v", err)
	}
	if len(first.Targets) == 0 {
		t.Fatalf("expected first eligible evaluation to rebalance")
	}
	second, err := strategy.EvaluatePortfolioLatest(context.Background(), panelPrefix(panel, 28))
	if err != nil {
		t.Fatalf("second EvaluatePortfolioLatest: %v", err)
	}
	if len(second.Targets) != 0 || second.EngineMetadata["action"] != actionHoldTargets {
		t.Fatalf("second output should hold targets, got targets=%d metadata=%v", len(second.Targets), second.EngineMetadata)
	}
}
