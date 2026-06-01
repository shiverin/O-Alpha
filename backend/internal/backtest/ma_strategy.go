package backtest

import (
	"context"
	"fmt"

	"github.com/oalpha/pkg/models"
)

// MACrossoverStrategy implements the Strategy interface using MA crossover logic.
type MACrossoverStrategy struct {
	FastPeriod int
	SlowPeriod int
}

// NewMACrossoverStrategy creates a new MA crossover strategy.
func NewMACrossoverStrategy(fast, slow int) *MACrossoverStrategy {
	return &MACrossoverStrategy{
		FastPeriod: fast,
		SlowPeriod: slow,
	}
}

// GenerateSignal returns signals based on MA crossover.
// Signal at index i uses data up to and including bar i.
func (s *MACrossoverStrategy) GenerateSignal(_ context.Context, bars []models.Bar) ([]models.Signal, error) {
	if s.FastPeriod <= 0 || s.SlowPeriod <= 0 {
		return nil, fmt.Errorf("periods must be positive")
	}
	if s.FastPeriod >= s.SlowPeriod {
		return nil, fmt.Errorf("fast period must be less than slow period")
	}
	if len(bars) < s.SlowPeriod+1 {
		return nil, fmt.Errorf("not enough bars: need at least %d", s.SlowPeriod+1)
	}

	closes := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
	}

	return s.Signals(closes), nil
}

// Signals returns one signal per bar (index-aligned with closes).
// Signal at index i uses data up to and including bar i.
func (s *MACrossoverStrategy) Signals(closes []float64) []models.Signal {
	n := len(closes)
	out := make([]models.Signal, n)
	if s.FastPeriod <= 0 || s.SlowPeriod <= 0 || s.FastPeriod >= s.SlowPeriod {
		return out
	}
	if n < s.SlowPeriod {
		return out
	}

	fast := rollingSMA(closes, s.FastPeriod)
	slow := rollingSMA(closes, s.SlowPeriod)

	var position int // 0 flat, 1 long
	for i := s.SlowPeriod; i < n; i++ {
		if fast[i-1] <= slow[i-1] && fast[i] > slow[i] {
			out[i] = models.SignalBuy
			position = 1
		} else if fast[i-1] >= slow[i-1] && fast[i] < slow[i] {
			out[i] = models.SignalSell
			position = 0
		} else if position == 1 {
			out[i] = models.SignalHold
		}
	}
	return out
}

// rollingSMA computes simple moving average; indices before period-1 are zero.
func rollingSMA(values []float64, period int) []float64 {
	n := len(values)
	out := make([]float64, n)
	if period <= 0 || n < period {
		return out
	}

	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	out[period-1] = sum / float64(period)

	for i := period; i < n; i++ {
		sum += values[i] - values[i-period]
		out[i] = sum / float64(period)
	}
	return out
}

func (s *MACrossoverStrategy) GenerateSignals(ctx context.Context, bars []models.Bar) ([]StrategyOutput, error) {
	signals, err := s.GenerateSignal(ctx, bars) // Uses your existing logic
	if err != nil {
		return nil, err
	}

	out := make([]StrategyOutput, len(signals))
	for i, signal := range signals {
		out[i] = StrategyOutput{
			Signal:          signal,
			PositionSizePct: 0.10,
			RegimeLabel:     "NORMAL",
		}
	}
	return out, nil
}

func (s *MACrossoverStrategy) EvaluateLatest(ctx context.Context, bars []models.Bar) (StrategyOutput, error) {
	outputs, err := s.GenerateSignals(ctx, bars)
	if err != nil || len(outputs) == 0 {
		return StrategyOutput{Signal: models.SignalHold}, err
	}

	return outputs[len(outputs)-1], nil
}
