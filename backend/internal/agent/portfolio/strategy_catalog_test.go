package portfolio

import (
	"context"
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

func TestAvailableStrategySpecsCoversRiskBuckets(t *testing.T) {
	symbols := catalogTestSymbols()
	specs := AvailableStrategySpecs(symbols, StrategyCatalogConfig{})
	if len(specs) < 9 {
		t.Fatalf("specs=%d, want at least 9", len(specs))
	}
	seenKeys := make(map[string]bool)
	buckets := make(map[StrategyRiskProfile]int)
	for _, spec := range specs {
		if spec.Key == "" {
			t.Fatalf("empty strategy key: %+v", spec)
		}
		if seenKeys[spec.Key] {
			t.Fatalf("duplicate strategy key %s", spec.Key)
		}
		seenKeys[spec.Key] = true
		buckets[spec.RiskProfile]++
		if spec.DeploymentStatus == "" {
			t.Fatalf("strategy %s missing deployment status", spec.Key)
		}
		if !spec.PaperOnly {
			t.Fatalf("strategy %s should default to paper-only", spec.Key)
		}
	}
	for _, bucket := range []StrategyRiskProfile{StrategyRiskLow, StrategyRiskMedium, StrategyRiskHigh} {
		if buckets[bucket] < 3 {
			t.Fatalf("risk bucket %s has %d specs, want at least 3", bucket, buckets[bucket])
		}
	}
	promoted, err := StrategySpecByKey("lgbm_ranker_h63_medium", symbols, StrategyCatalogConfig{})
	if err != nil {
		t.Fatalf("StrategySpecByKey: %v", err)
	}
	if !promoted.PromotedCheckpoint || promoted.RiskProfile != StrategyRiskMedium {
		t.Fatalf("promoted h63 spec mismatch: %+v", promoted)
	}
}

func TestNewStrategyFromCatalogBuildsEverySpec(t *testing.T) {
	symbols := catalogTestSymbols()
	cfg := StrategyCatalogConfig{ModelArtifactRoot: "/tmp/ranker_artifacts"}
	for _, spec := range AvailableStrategySpecs(symbols, cfg) {
		strategy, builtSpec, err := NewStrategyFromCatalog(spec.Key, symbols, cfg)
		if err != nil {
			t.Fatalf("NewStrategyFromCatalog(%s): %v", spec.Key, err)
		}
		if strategy == nil {
			t.Fatalf("NewStrategyFromCatalog(%s) returned nil strategy", spec.Key)
		}
		if builtSpec.Key != spec.Key {
			t.Fatalf("built spec key=%s, want %s", builtSpec.Key, spec.Key)
		}
		if strategy.Name() != spec.Key {
			t.Fatalf("strategy name=%s, want %s", strategy.Name(), spec.Key)
		}
	}
}

func TestLabeledStrategyAddsAgentMetadata(t *testing.T) {
	t0 := time.Date(2026, 1, 2, 14, 30, 0, 0, time.UTC)
	strategy := &labeledPortfolioStrategy{
		inner: staticPortfolioStrategy{
			output: backtest.PortfolioOutput{
				Time: t0,
				Targets: map[string]backtest.TargetPosition{
					"VOO": {Symbol: "VOO", TargetWeight: 1},
				},
			},
		},
		spec: StrategySpec{
			Key:                "unit_strategy",
			DisplayName:        "Unit Strategy",
			Family:             "unit",
			RiskProfile:        StrategyRiskLow,
			DeploymentStatus:   StrategyStatusPaperOnly,
			PromotedCheckpoint: false,
			PaperOnly:          true,
		},
	}
	output, err := strategy.EvaluatePortfolioLatest(context.Background(), oneBarPanel())
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if output.EngineMetadata["agent_strategy_key"] != "unit_strategy" {
		t.Fatalf("missing strategy key metadata: %+v", output.EngineMetadata)
	}
	if output.EngineMetadata["agent_strategy_risk_profile"] != "low" {
		t.Fatalf("missing risk profile metadata: %+v", output.EngineMetadata)
	}
	if output.EngineMetadata["agent_strategy_paper_only"] != true {
		t.Fatalf("missing paper-only metadata: %+v", output.EngineMetadata)
	}
}

func TestStartCatalogPortfolioAgent(t *testing.T) {
	manager := NewPortfolioAgentManager(nil, nil)
	worker, spec, err := manager.StartCatalogPortfolioAgent(
		context.Background(),
		"paper-low",
		"ranker_proxy_h63_low",
		catalogTestSymbols(),
		"1Day",
		100000,
		StrategyCatalogConfig{},
		nil,
	)
	if err != nil {
		t.Fatalf("StartCatalogPortfolioAgent: %v", err)
	}
	if worker == nil {
		t.Fatalf("worker is nil")
	}
	if spec.RiskProfile != StrategyRiskLow {
		t.Fatalf("risk profile=%s, want low", spec.RiskProfile)
	}
	if _, _, err := manager.StartCatalogPortfolioAgent(context.Background(), "paper-low", "ranker_proxy_h63_low", catalogTestSymbols(), "1Day", 100000, StrategyCatalogConfig{}, nil); err == nil {
		t.Fatalf("expected duplicate key error")
	}
}

type staticPortfolioStrategy struct {
	output backtest.PortfolioOutput
}

func (s staticPortfolioStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	out := make([]backtest.PortfolioOutput, len(panel.Times))
	for i := range out {
		out[i] = s.output
		out[i].Time = panel.Times[i]
	}
	return out, nil
}

func (s staticPortfolioStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	return s.output, nil
}

func (s staticPortfolioStrategy) Universe() []string {
	return []string{"VOO"}
}

func (s staticPortfolioStrategy) Name() string {
	return "static"
}

func catalogTestSymbols() []string {
	return []string{
		"VOO", "AAPL", "MSFT", "NVDA", "AMAT", "INTC", "LRCX", "JNJ", "PG", "COST",
		"QQQ", "SPY", "IWM", "VTI", "XLU", "XLP", "XLV", "XLE", "XLF", "XLK",
	}
}

func oneBarPanel() backtest.AlignedBars {
	t0 := time.Date(2026, 1, 2, 14, 30, 0, 0, time.UTC)
	return backtest.AlignedBars{
		Times:   []time.Time{t0},
		Symbols: []string{"VOO"},
		Bars: map[string][]models.Bar{
			"VOO": {{Time: t0, Symbol: "VOO", Close: 100}},
		},
	}
}
