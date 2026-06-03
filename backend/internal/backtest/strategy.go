package backtest

import (
	"context"

	"github.com/oalpha/pkg/models"
)

// StrategyOutput wraps trade signals with metadata for execution and telemetry
type StrategyOutput struct {
	Signal          models.Signal
	PositionSizePct float64                // Target sizing (e.g., 0.10 for 10% cash allocation)
	RegimeLabel     string                 // Human-readable regime for tracking
	Metadata        map[string]interface{} // Open slot for HMM probabilities or indicator scores

	// Continuous output fields for target-weight engines. PositionSizePct stays
	// for legacy callers; new engines should prefer TargetWeight.
	AlphaScore   float64
	Confidence   float64
	TargetWeight float64
	Engine       string
}

type Strategy interface {
	// GenerateSignals processes historical data slices (ideal for Backtesting engines)
	GenerateSignals(ctx context.Context, bars []models.Bar) ([]StrategyOutput, error)

	// EvaluateLatest computes the decision context for the absolute trailing edge bar (Live/Paper loop)
	EvaluateLatest(ctx context.Context, bars []models.Bar) (StrategyOutput, error)
}

func confidenceFromSignal(signal models.Signal) float64 {
	if signal == models.SignalHold {
		return 0
	}
	return 1
}

func targetWeightFromSignal(signal models.Signal, defaultWeight float64) float64 {
	switch signal {
	case models.SignalBuy:
		if defaultWeight <= 0 {
			return 1
		}
		if defaultWeight > 1 {
			return 1
		}
		return defaultWeight
	case models.SignalSell:
		return 0
	default:
		return 0
	}
}
