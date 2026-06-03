package ml

import (
	"context"
	"fmt"
	"testing"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

func TestMLMetaLabelStrategyAcceptsHighProbabilityBuy(t *testing.T) {
	bars := testBars("VOO", 80, 100, 1)
	base := &scriptedBaseStrategy{outputs: repeatOutput(len(bars), backtest.StrategyOutput{
		Signal:          models.SignalBuy,
		PositionSizePct: 0.10,
		TargetWeight:    0.10,
		AlphaScore:      1,
		Confidence:      1,
		Engine:          "base",
	})}
	strategy := &MLMetaLabelStrategy{
		Symbol:       "VOO",
		BaseStrategy: base,
		Predictor:    constantPredictor{p: 0.60, version: "model-v1"},
		MaxWeight:    0.20,
		Thresholds:   DefaultMLThresholds(),
		FeatureBuilder: NewFeatureBuilder(FeatureSpec{
			Version:  "test",
			Features: []string{"log_ret_1", "ensemble_score"},
		}),
	}

	out, err := strategy.EvaluateLatest(context.Background(), bars)
	if err != nil {
		t.Fatalf("EvaluateLatest: %v", err)
	}
	if out.Signal != models.SignalBuy {
		t.Fatalf("signal=%v, want buy", out.Signal)
	}
	if out.TargetWeight <= 0 || out.TargetWeight > 0.10 {
		t.Fatalf("target weight=%f, want positive and capped by base target", out.TargetWeight)
	}
	if out.Engine != MetaLabelEngineName {
		t.Fatalf("engine=%s, want %s", out.Engine, MetaLabelEngineName)
	}
	if out.Metadata["model_version"] != "model-v1" {
		t.Fatalf("missing model version metadata")
	}
}

func TestMLMetaLabelStrategyDefaultThresholdIsPointFive(t *testing.T) {
	bars := testBars("VOO", 80, 100, 1)
	base := &scriptedBaseStrategy{outputs: repeatOutput(len(bars), backtest.StrategyOutput{
		Signal:       models.SignalBuy,
		TargetWeight: 0.10,
		Engine:       "base",
	})}
	strategy := &MLMetaLabelStrategy{
		Symbol:       "VOO",
		BaseStrategy: base,
		Predictor:    constantPredictor{p: 0.51, version: "model-v1"},
		FeatureBuilder: NewFeatureBuilder(FeatureSpec{
			Version:  "test",
			Features: []string{"log_ret_1"},
		}),
	}

	out, err := strategy.EvaluateLatest(context.Background(), bars)
	if err != nil {
		t.Fatalf("EvaluateLatest: %v", err)
	}
	if out.Signal != models.SignalBuy {
		t.Fatalf("signal=%v, want buy at p>=0.50", out.Signal)
	}
}

func TestMLMetaLabelStrategyVetoesLowProbabilityBuy(t *testing.T) {
	bars := testBars("VOO", 80, 100, 1)
	base := &scriptedBaseStrategy{outputs: repeatOutput(len(bars), backtest.StrategyOutput{
		Signal:          models.SignalBuy,
		PositionSizePct: 0.10,
		TargetWeight:    0.10,
		Engine:          "base",
	})}
	strategy := &MLMetaLabelStrategy{
		Symbol:       "VOO",
		BaseStrategy: base,
		Predictor:    constantPredictor{p: 0.40, version: "model-v1"},
		Thresholds:   DefaultMLThresholds(),
		FeatureBuilder: NewFeatureBuilder(FeatureSpec{
			Version:  "test",
			Features: []string{"log_ret_1"},
		}),
	}

	out, err := strategy.EvaluateLatest(context.Background(), bars)
	if err != nil {
		t.Fatalf("EvaluateLatest: %v", err)
	}
	if out.Signal != models.SignalHold || out.TargetWeight != 0 {
		t.Fatalf("signal=%v target=%f, want hold with zero target", out.Signal, out.TargetWeight)
	}
	if out.Metadata["ml_decision"] != "veto_buy" {
		t.Fatalf("metadata decision=%v, want veto_buy", out.Metadata["ml_decision"])
	}
}

func TestMLMetaLabelStrategyAppliesCalibrationBeforeThreshold(t *testing.T) {
	bars := testBars("VOO", 80, 100, 1)
	base := &scriptedBaseStrategy{outputs: repeatOutput(len(bars), backtest.StrategyOutput{
		Signal:       models.SignalBuy,
		TargetWeight: 0.10,
		Engine:       "base",
	})}
	strategy := &MLMetaLabelStrategy{
		Symbol:       "VOO",
		BaseStrategy: base,
		Predictor:    constantPredictor{p: 0.40, version: "model-v1"},
		Calibration: CalibrationModel{
			Method:    "platt",
			Intercept: 1.0,
			Slope:     1.0,
		},
		Thresholds: MLThresholds{
			EnterLong:        0.50,
			Reduce:           0.45,
			BetSizingSlope:   2.0,
			PassThroughExits: true,
		},
		FeatureBuilder: NewFeatureBuilder(FeatureSpec{
			Version:  "test",
			Features: []string{"log_ret_1"},
		}),
	}

	out, err := strategy.EvaluateLatest(context.Background(), bars)
	if err != nil {
		t.Fatalf("EvaluateLatest: %v", err)
	}
	if out.Signal != models.SignalBuy {
		t.Fatalf("signal=%v, want calibrated probability to accept buy", out.Signal)
	}
	raw, ok := out.Metadata["p_success_raw"].(float64)
	if !ok || raw != 0.40 {
		t.Fatalf("raw probability metadata=%v, want 0.40", out.Metadata["p_success_raw"])
	}
	calibrated, ok := out.Metadata["p_success"].(float64)
	if !ok || calibrated <= 0.50 {
		t.Fatalf("calibrated probability metadata=%v, want >0.50", out.Metadata["p_success"])
	}
}

func TestMLMetaLabelStrategyDoesNotCreateAlphaFromHold(t *testing.T) {
	bars := testBars("VOO", 80, 100, 1)
	base := &scriptedBaseStrategy{outputs: repeatOutput(len(bars), backtest.StrategyOutput{
		Signal: models.SignalHold,
		Engine: "base",
	})}
	strategy := &MLMetaLabelStrategy{
		Symbol:       "VOO",
		BaseStrategy: base,
		Predictor:    constantPredictor{p: 0.99, version: "model-v1"},
	}

	out, err := strategy.EvaluateLatest(context.Background(), bars)
	if err != nil {
		t.Fatalf("EvaluateLatest: %v", err)
	}
	if out.Signal != models.SignalHold || out.TargetWeight != 0 {
		t.Fatalf("signal=%v target=%f, want base hold preserved", out.Signal, out.TargetWeight)
	}
}

func TestMLMetaLabelStrategyPassesThroughBaseExit(t *testing.T) {
	bars := testBars("VOO", 80, 100, 1)
	base := &scriptedBaseStrategy{outputs: repeatOutput(len(bars), backtest.StrategyOutput{
		Signal:       models.SignalSell,
		TargetWeight: 0,
		Engine:       "base",
	})}
	strategy := &MLMetaLabelStrategy{
		Symbol:       "VOO",
		BaseStrategy: base,
		Predictor:    constantPredictor{p: 0.99, version: "model-v1"},
		Thresholds:   DefaultMLThresholds(),
	}

	out, err := strategy.EvaluateLatest(context.Background(), bars)
	if err != nil {
		t.Fatalf("EvaluateLatest: %v", err)
	}
	if out.Signal != models.SignalSell {
		t.Fatalf("signal=%v, want sell", out.Signal)
	}
	if out.Metadata["ml_action"] != "base_exit" {
		t.Fatalf("metadata action=%v, want base_exit", out.Metadata["ml_action"])
	}
}

type constantPredictor struct {
	p       float64
	version string
	err     error
}

func (p constantPredictor) PredictProba(features []float64) (float64, error) {
	if p.err != nil {
		return 0, p.err
	}
	if len(features) == 0 {
		return 0, fmt.Errorf("empty features")
	}
	return p.p, nil
}

func (p constantPredictor) Version() string {
	return p.version
}

type scriptedBaseStrategy struct {
	outputs []backtest.StrategyOutput
}

func (s *scriptedBaseStrategy) GenerateSignals(ctx context.Context, bars []models.Bar) ([]backtest.StrategyOutput, error) {
	if len(s.outputs) != len(bars) {
		return nil, fmt.Errorf("scripted outputs length %d does not match bars length %d", len(s.outputs), len(bars))
	}
	return append([]backtest.StrategyOutput(nil), s.outputs...), nil
}

func (s *scriptedBaseStrategy) EvaluateLatest(ctx context.Context, bars []models.Bar) (backtest.StrategyOutput, error) {
	if len(s.outputs) == 0 {
		return backtest.StrategyOutput{}, fmt.Errorf("no scripted outputs")
	}
	return s.outputs[len(s.outputs)-1], nil
}

func repeatOutput(n int, output backtest.StrategyOutput) []backtest.StrategyOutput {
	out := make([]backtest.StrategyOutput, n)
	for i := range out {
		out[i] = output
	}
	return out
}
