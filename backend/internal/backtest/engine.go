package backtest

import (
	"context"
	"fmt"

	"github.com/oalpha/pkg/models"
)

// RunBacktest simulates a strategy on historical data.
// Signals generated at bar t execute at bar t+1 open (no look-ahead).
func RunBacktest(ctx context.Context, bars []models.Bar, strat Strategy, initialCash float64) (*models.BacktestResult, error) {
	if len(bars) < 1 {
		return nil, fmt.Errorf("need at least one bar")
	}
	if initialCash <= 0 {
		initialCash = 100_000
	}

	outputs, err := strat.GenerateSignals(ctx, bars)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signals: %w", err)
	}

	if len(outputs) != len(bars) {
		return nil, fmt.Errorf("signals length (%d) does not match bars length (%d)", len(outputs), len(bars))
	}

	cash := initialCash
	shares := 0.0
	numTrades := 0
	equityCurve := make([]models.EquityPoint, 0, len(bars))

	for i := 0; i < len(bars); i++ {
		// Execute pending signal from previous bar at this bar's open.
		if i > 0 {
			switch outputs[i-1].Signal {
			case models.SignalBuy:
				if shares == 0 && bars[i].Open > 0 {
					allocationPct := outputs[i-1].PositionSizePct
					if allocationPct <= 0 || allocationPct > 1 {
						allocationPct = 1
					}
					allocation := cash * allocationPct
					shares = allocation / bars[i].Open
					cash -= allocation
					numTrades++
				}
			case models.SignalSell:
				if shares > 0 {
					cash = shares * bars[i].Open
					shares = 0
					numTrades++
				}
			}
		}

		equity := cash
		if shares > 0 {
			equity = shares * bars[i].Close
		}
		equityCurve = append(equityCurve, models.EquityPoint{
			Time:   bars[i].Time,
			Equity: equity,
		})
	}

	if shares > 0 {
		cash = shares * bars[len(bars)-1].Close
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
