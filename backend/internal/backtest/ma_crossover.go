package backtest

import (
	"github.com/oalpha/pkg/models"
)

// MACrossover generates signals when fast SMA crosses slow SMA.
type MACrossover struct {
	FastPeriod int
	SlowPeriod int
}

// Signals returns one signal per bar (index-aligned with closes).
// Signal at index i uses data up to and including bar i.
func (s *MACrossover) Signals(closes []float64) []models.Signal {
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
