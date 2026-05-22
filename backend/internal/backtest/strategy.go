package backtest

import (
	"context"

	"github.com/oalpha/pkg/models"
)

// Strategy defines the interface for generating trading signals.
// Implementations can be based on various algorithms (MA crossover, regime detection, pairs trading, etc.).
type Strategy interface {
	// GenerateSignal returns a signal for the given bar data.
	// The signal at index i should be based on data up to and including bar i.
	// Signals are executed at the next bar's open to avoid look-ahead bias.
	GenerateSignal(ctx context.Context, bars []models.Bar) ([]models.Signal, error)
}