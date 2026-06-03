package backtest

import "math"

type PortfolioMetrics struct {
	TotalReturn      float64 `json:"total_return"`
	AnnualReturn     float64 `json:"annual_return"`
	AnnualVol        float64 `json:"annual_vol"`
	Sharpe           float64 `json:"sharpe"`
	Sortino          float64 `json:"sortino"`
	Calmar           float64 `json:"calmar"`
	MaxDrawdown      float64 `json:"max_drawdown"`
	ProfitFactor     float64 `json:"profit_factor"`
	WinRate          float64 `json:"win_rate"`
	AvgWin           float64 `json:"avg_win"`
	AvgLoss          float64 `json:"avg_loss"`
	NumTrades        int     `json:"num_trades"`
	Turnover         float64 `json:"turnover"`
	AvgGrossExposure float64 `json:"avg_gross_exposure"`
	AvgNetExposure   float64 `json:"avg_net_exposure"`
	TailRatio        float64 `json:"tail_ratio"`
	Skew             float64 `json:"skew"`
	Kurtosis         float64 `json:"kurtosis"`
	DSR              float64 `json:"dsr"`
	PBO              float64 `json:"pbo"`
}

func ComputePortfolioMetrics(equity []float64, pnls []float64, grossExposures []float64, netExposures []float64, turnover float64) PortfolioMetrics {
	base := ComputeMetrics(equity)
	returns := equityReturns(equity)
	stats := computeTradeStats(pnls)
	return PortfolioMetrics{
		TotalReturn:      base.TotalReturn,
		AnnualReturn:     base.AnnualizedReturn,
		AnnualVol:        annualizedVol(returns),
		Sharpe:           base.Sharpe,
		Sortino:          base.Sortino,
		Calmar:           base.Calmar,
		MaxDrawdown:      base.MaxDrawdown,
		ProfitFactor:     stats.ProfitFactor,
		WinRate:          stats.WinRate,
		AvgWin:           stats.AverageWin,
		AvgLoss:          stats.AverageLoss,
		NumTrades:        len(pnls),
		Turnover:         turnover,
		AvgGrossExposure: mean(grossExposures),
		AvgNetExposure:   mean(netExposures),
		TailRatio:        tailRatio(returns),
		Skew:             skewness(returns),
		Kurtosis:         kurtosis(returns),
	}
}

func ProbabilisticSharpeRatio(observedSharpe float64, benchmarkSharpe float64, n int, skew float64, kurtosisValue float64) float64 {
	if n < 2 {
		return 0
	}
	denom := math.Sqrt(1 - skew*observedSharpe + ((kurtosisValue-1)/4)*observedSharpe*observedSharpe)
	if denom <= 0 || math.IsNaN(denom) || math.IsInf(denom, 0) {
		return 0
	}
	z := (observedSharpe - benchmarkSharpe) * math.Sqrt(float64(n-1)) / denom
	return normalCDF(z)
}

func equityReturns(equity []float64) []float64 {
	if len(equity) < 2 {
		return nil
	}
	out := make([]float64, 0, len(equity)-1)
	for i := 1; i < len(equity); i++ {
		if equity[i-1] <= 0 {
			continue
		}
		out = append(out, (equity[i]-equity[i-1])/equity[i-1])
	}
	return out
}

func annualizedVol(returns []float64) float64 {
	_, sigma := meanStd(returns)
	return sigma * math.Sqrt(tradingDaysPerYear)
}

func skewness(values []float64) float64 {
	if len(values) < 3 {
		return 0
	}
	mu, sigma := meanStd(values)
	if sigma == 0 {
		return 0
	}
	var sum float64
	for _, value := range values {
		z := (value - mu) / sigma
		sum += z * z * z
	}
	return sum / float64(len(values))
}

func kurtosis(values []float64) float64 {
	if len(values) < 4 {
		return 0
	}
	mu, sigma := meanStd(values)
	if sigma == 0 {
		return 0
	}
	var sum float64
	for _, value := range values {
		z := (value - mu) / sigma
		sum += z * z * z * z
	}
	return sum / float64(len(values))
}

func tailRatio(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sortFloat64s(sorted)
	lower := percentileSorted(sorted, 5)
	upper := percentileSorted(sorted, 95)
	if lower == 0 {
		return 0
	}
	return math.Abs(upper / lower)
}

func percentileSorted(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}
	rank := (p / 100) * float64(n-1)
	lo := int(math.Floor(rank))
	hi := int(math.Ceil(rank))
	if lo == hi {
		return sorted[lo]
	}
	frac := rank - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

func normalCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}

func sortFloat64s(values []float64) {
	for i := 1; i < len(values); i++ {
		key := values[i]
		j := i - 1
		for j >= 0 && values[j] > key {
			values[j+1] = values[j]
			j--
		}
		values[j+1] = key
	}
}
