package backtest

// Signal is a trading action.
type Signal int

const (
	SignalHold Signal = iota
	SignalBuy
	SignalSell
)

// MACrossover generates signals when fast SMA crosses slow SMA.
type MACrossover struct {
	FastPeriod int
	SlowPeriod int
}

// Signals returns one signal per bar (index-aligned with closes).
// Signal at index i uses data up to and including bar i.
func (s *MACrossover) Signals(closes []float64) []Signal {
	n := len(closes)
	out := make([]Signal, n)
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
			out[i] = SignalBuy
			position = 1
		} else if fast[i-1] >= slow[i-1] && fast[i] < slow[i] {
			out[i] = SignalSell
			position = 0
		} else if position == 1 {
			out[i] = SignalHold
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
