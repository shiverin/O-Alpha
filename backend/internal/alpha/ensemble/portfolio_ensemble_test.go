package ensemble

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
)

func TestComputeSleeveWeightsUsesInverseVol(t *testing.T) {
	weights := ComputeSleeveWeights([]SleeveOutput{
		{Name: "low_vol", RealizedVolAnnual: 0.10},
		{Name: "high_vol", RealizedVolAnnual: 0.20},
	})
	if math.Abs(weights["low_vol"]-2.0/3.0) > 1e-9 {
		t.Fatalf("low_vol weight=%f, want 2/3", weights["low_vol"])
	}
	if math.Abs(weights["high_vol"]-1.0/3.0) > 1e-9 {
		t.Fatalf("high_vol weight=%f, want 1/3", weights["high_vol"])
	}
}

func TestCombineSleeveOutputsNetsSameSymbol(t *testing.T) {
	timestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	output := CombineSleeveOutputs(timestamp, []SleeveOutput{
		{
			Name:   "momentum",
			Weight: 1,
			Output: outputWithTarget(timestamp, "AAPL", 0.10),
		},
		{
			Name:   "cointegration",
			Weight: 1,
			Output: outputWithTarget(timestamp, "AAPL", -0.03),
		},
	}, PortfolioRiskConfig{TargetVolAnnual: 1, MaxSymbolWeight: 1, MaxGrossExposure: 1, MaxNetExposure: 1}, PortfolioRiskState{}, nil)

	got := output.Targets["AAPL"].TargetWeight
	want := 0.035
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("net target=%f, want %f", got, want)
	}
	contrib := output.Targets["AAPL"].Metadata["sleeve_contributions"].(map[string]float64)
	if math.Abs(contrib["momentum"]-0.05) > 1e-9 || math.Abs(contrib["cointegration"]+0.015) > 1e-9 {
		t.Fatalf("unexpected contributions: %v", contrib)
	}
}

func TestMultiEngineEnsembleStrategyEvaluatesSleeves(t *testing.T) {
	timestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	panel := backtest.AlignedBars{Times: []time.Time{timestamp}}
	strategy := NewMultiEngineEnsembleStrategy([]Sleeve{
		{Name: "a", Strategy: staticPortfolioStrategy{symbols: []string{"AAPL"}, output: outputWithTarget(timestamp, "AAPL", 0.10)}, Weight: 1},
		{Name: "b", Strategy: staticPortfolioStrategy{symbols: []string{"MSFT"}, output: outputWithTarget(timestamp, "MSFT", 0.20)}, Weight: 1},
	}, PortfolioRiskConfig{TargetVolAnnual: 1, MaxSymbolWeight: 1, MaxGrossExposure: 1, MaxNetExposure: 1}, nil)

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panel)
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if len(output.Targets) != 2 {
		t.Fatalf("targets=%d, want 2", len(output.Targets))
	}
}

func outputWithTarget(t time.Time, symbol string, weight float64) backtest.PortfolioOutput {
	side := backtest.PositionSideLong
	if weight < 0 {
		side = backtest.PositionSideShort
	}
	return backtest.PortfolioOutput{
		Time: t,
		Targets: map[string]backtest.TargetPosition{
			symbol: {
				Symbol:       symbol,
				TargetWeight: weight,
				AlphaScore:   weight * 10,
				Confidence:   0.8,
				Side:         side,
				Engine:       "test",
			},
		},
	}
}

type staticPortfolioStrategy struct {
	symbols []string
	output  backtest.PortfolioOutput
}

func (s staticPortfolioStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	_ = ctx
	outputs := make([]backtest.PortfolioOutput, len(panel.Times))
	for i := range outputs {
		outputs[i] = s.output
	}
	return outputs, nil
}

func (s staticPortfolioStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	_ = ctx
	_ = panel
	return s.output, nil
}

func (s staticPortfolioStrategy) Universe() []string {
	return s.symbols
}

func (s staticPortfolioStrategy) Name() string {
	return "static"
}
