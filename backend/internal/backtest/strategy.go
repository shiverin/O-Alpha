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
}

type Strategy interface {
	// GenerateSignals processes historical data slices (ideal for Backtesting engines)
	GenerateSignals(ctx context.Context, bars []models.Bar) ([]StrategyOutput, error)

	// EvaluateLatest computes the decision context for the absolute trailing edge bar (Live/Paper loop)
	EvaluateLatest(ctx context.Context, bars []models.Bar) (StrategyOutput, error)
}
