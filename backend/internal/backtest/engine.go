package backtest

import (
	"fmt"

	"github.com/oalpha/pkg/models"
)

// RunMACrossover simulates a long-only MA crossover strategy.
// Signals generated at bar t execute at bar t+1 open (no look-ahead).
func RunMACrossover(bars []models.Bar, fast, slow int, initialCash float64) (*models.BacktestResult, error) {
	if fast <= 0 || slow <= 0 {
		return nil, fmt.Errorf("periods must be positive")
	}
	if fast >= slow {
		return nil, fmt.Errorf("fast period must be less than slow period")
	}
	if len(bars) < slow+1 {
		return nil, fmt.Errorf("not enough bars: need at least %d", slow+1)
	}
	if initialCash <= 0 {
		initialCash = 100_000
	}

	closes := make([]float64, len(bars))
	opens := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
		opens[i] = b.Open
	}

	strategy := &MACrossover{FastPeriod: fast, SlowPeriod: slow}
	signals := strategy.Signals(closes)

	cash := initialCash
	shares := 0.0
	numTrades := 0
	equityCurve := make([]models.EquityPoint, 0, len(bars))

	for i := 0; i < len(bars); i++ {
		// Execute pending signal from previous bar at this bar's open.
		if i > 0 {
			switch signals[i-1] {
			case SignalBuy:
				if shares == 0 && opens[i] > 0 {
					shares = cash / opens[i]
					cash = 0
					numTrades++
				}
			case SignalSell:
				if shares > 0 {
					cash = shares * opens[i]
					shares = 0
					numTrades++
				}
			}
		}

		equity := cash
		if shares > 0 {
			equity = shares * closes[i]
		}
		equityCurve = append(equityCurve, models.EquityPoint{
			Time:   bars[i].Time,
			Equity: equity,
		})
	}

	// Mark to market at last close if still holding.
	if shares > 0 {
		cash = shares * closes[len(closes)-1]
		shares = 0
	}

	equities := make([]float64, len(equityCurve))
	for i, p := range equityCurve {
		equities[i] = p.Equity
	}
	m := ComputeMetrics(equities)

	return &models.BacktestResult{
		Symbol:      bars[0].Symbol,
		EquityCurve: equityCurve,
		FinalEquity: cash,
		TotalReturn: m.TotalReturn,
		Sharpe:      m.Sharpe,
		Sortino:     m.Sortino,
		MaxDrawdown: m.MaxDrawdown,
		NumTrades:   numTrades,
	}, nil
}
