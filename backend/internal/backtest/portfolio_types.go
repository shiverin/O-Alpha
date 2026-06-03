package backtest

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/oalpha/pkg/models"
)

// AlignedBars is a point-in-time, symbol-aligned OHLCV panel.
type AlignedBars struct {
	Times      []time.Time
	Symbols    []string
	Bars       map[string][]models.Bar
	Timeframe  string
	Feed       string
	Adjustment string
	Metadata   map[string]interface{}
}

type PositionSide string

const (
	PositionSideLong  PositionSide = "long"
	PositionSideShort PositionSide = "short"
)

type TargetPosition struct {
	Symbol       string                 `json:"symbol"`
	TargetWeight float64                `json:"target_weight"`
	AlphaScore   float64                `json:"alpha_score"`
	Confidence   float64                `json:"confidence"`
	Side         PositionSide           `json:"side"`
	Engine       string                 `json:"engine"`
	RegimeLabel  string                 `json:"regime_label"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type PortfolioOutput struct {
	Time           time.Time                 `json:"time"`
	Targets        map[string]TargetPosition `json:"targets"`
	GrossExposure  float64                   `json:"gross_exposure"`
	NetExposure    float64                   `json:"net_exposure"`
	CashWeight     float64                   `json:"cash_weight"`
	EngineMetadata map[string]interface{}    `json:"engine_metadata,omitempty"`
}

type PortfolioStrategy interface {
	GeneratePortfolioSignals(ctx context.Context, panel AlignedBars) ([]PortfolioOutput, error)
	EvaluatePortfolioLatest(ctx context.Context, panel AlignedBars) (PortfolioOutput, error)
	Universe() []string
	Name() string
}

type SingleSymbolPortfolioAdapter struct {
	Symbol string
	Strat  Strategy
}

func (a *SingleSymbolPortfolioAdapter) GeneratePortfolioSignals(ctx context.Context, panel AlignedBars) ([]PortfolioOutput, error) {
	if a == nil || a.Strat == nil {
		return nil, fmt.Errorf("single-symbol adapter requires a strategy")
	}
	bars, ok := panel.Bars[a.Symbol]
	if !ok {
		return nil, fmt.Errorf("panel missing symbol %s", a.Symbol)
	}
	outputs, err := a.Strat.GenerateSignals(ctx, bars)
	if err != nil {
		return nil, err
	}
	out := make([]PortfolioOutput, len(outputs))
	for i, strategyOutput := range outputs {
		timeValue := bars[i].Time
		if i < len(panel.Times) {
			timeValue = panel.Times[i]
		}
		out[i] = strategyOutputToPortfolioOutput(a.Symbol, timeValue, strategyOutput)
	}
	return out, nil
}

func (a *SingleSymbolPortfolioAdapter) EvaluatePortfolioLatest(ctx context.Context, panel AlignedBars) (PortfolioOutput, error) {
	if a == nil || a.Strat == nil {
		return PortfolioOutput{}, fmt.Errorf("single-symbol adapter requires a strategy")
	}
	bars, ok := panel.Bars[a.Symbol]
	if !ok || len(bars) == 0 {
		return PortfolioOutput{}, fmt.Errorf("panel missing symbol %s", a.Symbol)
	}
	out, err := a.Strat.EvaluateLatest(ctx, bars)
	if err != nil {
		return PortfolioOutput{}, err
	}
	timeValue := bars[len(bars)-1].Time
	if len(panel.Times) > 0 {
		timeValue = panel.Times[len(panel.Times)-1]
	}
	return strategyOutputToPortfolioOutput(a.Symbol, timeValue, out), nil
}

func (a *SingleSymbolPortfolioAdapter) Universe() []string {
	if a == nil || a.Symbol == "" {
		return nil
	}
	return []string{a.Symbol}
}

func (a *SingleSymbolPortfolioAdapter) Name() string {
	return "single_symbol_adapter"
}

func strategyOutputToPortfolioOutput(symbol string, t time.Time, out StrategyOutput) PortfolioOutput {
	targets := make(map[string]TargetPosition)
	weight := targetWeightFromStrategyOutput(out)
	if out.Signal == models.SignalHold && weight == 0 {
		return PortfolioOutput{
			Time:           t,
			Targets:        targets,
			CashWeight:     1,
			EngineMetadata: map[string]interface{}{"action": "hold_targets"},
		}
	}
	targets[symbol] = TargetPosition{
		Symbol:       symbol,
		TargetWeight: weight,
		AlphaScore:   scoreFromStrategyOutput(out),
		Confidence:   confidenceFromStrategyOutput(out),
		Side:         sideForWeight(weight),
		Engine:       engineFromStrategyOutput(out),
		RegimeLabel:  out.RegimeLabel,
		Metadata:     out.Metadata,
	}
	return PortfolioOutput{
		Time:          t,
		Targets:       targets,
		GrossExposure: math.Abs(weight),
		NetExposure:   weight,
		CashWeight:    math.Max(0, 1-math.Abs(weight)),
	}
}

func targetWeightFromStrategyOutput(out StrategyOutput) float64 {
	if out.TargetWeight != 0 {
		return clampWeight(out.TargetWeight)
	}
	switch out.Signal {
	case models.SignalBuy:
		return normalizePositionSizePct(out.PositionSizePct)
	case models.SignalSell:
		return 0
	default:
		return 0
	}
}

func scoreFromStrategyOutput(out StrategyOutput) float64 {
	if out.AlphaScore != 0 {
		return out.AlphaScore
	}
	return signalToSignedScore(out.Signal)
}

func confidenceFromStrategyOutput(out StrategyOutput) float64 {
	if out.Confidence != 0 {
		return clamp01(out.Confidence)
	}
	return metadataFloat(out.Metadata, "confidence")
}

func engineFromStrategyOutput(out StrategyOutput) string {
	if out.Engine != "" {
		return out.Engine
	}
	return "single_symbol_adapter"
}

func sideForWeight(weight float64) PositionSide {
	if weight < 0 {
		return PositionSideShort
	}
	return PositionSideLong
}

func signalToSignedScore(signal models.Signal) float64 {
	switch signal {
	case models.SignalBuy:
		return 1
	case models.SignalSell:
		return -1
	default:
		return 0
	}
}

func metadataFloat(metadata map[string]interface{}, key string) float64 {
	if metadata == nil {
		return 0
	}
	switch value := metadata[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	default:
		return 0
	}
}

func normalizePositionSizePct(sizePct float64) float64 {
	if math.IsNaN(sizePct) || math.IsInf(sizePct, 0) || sizePct < 0 {
		return 0
	}
	if sizePct == 0 {
		return 1
	}
	if sizePct > 1 {
		return 1
	}
	return sizePct
}

func clampWeight(weight float64) float64 {
	if math.IsNaN(weight) || math.IsInf(weight, 0) {
		return 0
	}
	if weight > 1 {
		return 1
	}
	if weight < -1 {
		return -1
	}
	return weight
}

func clamp01(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
