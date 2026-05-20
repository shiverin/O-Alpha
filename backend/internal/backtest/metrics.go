package backtest

import (
	"math"
)

// Metrics holds performance statistics.
type Metrics struct {
	Sharpe      float64
	Sortino     float64
	MaxDrawdown float64
	TotalReturn float64
}

const tradingDaysPerYear = 252.0

// ComputeMetrics calculates risk/return stats from an equity curve.
func ComputeMetrics(equity []float64) Metrics {
	if len(equity) < 2 {
		return Metrics{}
	}

	returns := make([]float64, 0, len(equity)-1)
	for i := 1; i < len(equity); i++ {
		if equity[i-1] == 0 {
			continue
		}
		returns = append(returns, (equity[i]-equity[i-1])/equity[i-1])
	}
	if len(returns) == 0 {
		return Metrics{}
	}

	totalReturn := (equity[len(equity)-1] - equity[0]) / equity[0]
	sharpe := annualizedSharpe(returns)
	sortino := annualizedSortino(returns)
	maxDD := maxDrawdown(equity)

	return Metrics{
		Sharpe:      sharpe,
		Sortino:     sortino,
		MaxDrawdown: maxDD,
		TotalReturn: totalReturn,
	}
}

func annualizedSharpe(returns []float64) float64 {
	mu, sigma := meanStd(returns)
	if sigma == 0 {
		return 0
	}
	return (mu / sigma) * math.Sqrt(tradingDaysPerYear)
}

func annualizedSortino(returns []float64) float64 {
	mu := mean(returns)
	downside := downsideDeviation(returns)
	if downside == 0 {
		return 0
	}
	return (mu / downside) * math.Sqrt(tradingDaysPerYear)
}

func maxDrawdown(equity []float64) float64 {
	peak := equity[0]
	maxDD := 0.0
	for _, e := range equity {
		if e > peak {
			peak = e
		}
		if peak > 0 {
			dd := (peak - e) / peak
			if dd > maxDD {
				maxDD = dd
			}
		}
	}
	return maxDD
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func meanStd(values []float64) (float64, float64) {
	mu := mean(values)
	if len(values) < 2 {
		return mu, 0
	}
	var sumSq float64
	for _, v := range values {
		d := v - mu
		sumSq += d * d
	}
	return mu, math.Sqrt(sumSq / float64(len(values)-1))
}

func downsideDeviation(returns []float64) float64 {
	var sumSq float64
	var n int
	for _, r := range returns {
		if r < 0 {
			sumSq += r * r
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return math.Sqrt(sumSq / float64(n))
}
