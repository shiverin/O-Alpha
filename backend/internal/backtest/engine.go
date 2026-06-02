package backtest

import (
	"context"
	"fmt"
	"time"

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

	return RunBacktestWithOutputs(bars, outputs, initialCash)
}

func RunBacktestWithOutputs(bars []models.Bar, outputs []StrategyOutput, initialCash float64) (*models.BacktestResult, error) {
	if len(bars) < 1 {
		return nil, fmt.Errorf("need at least one bar")
	}
	if len(outputs) != len(bars) {
		return nil, fmt.Errorf("signals length (%d) does not match bars length (%d)", len(outputs), len(bars))
	}
	if initialCash <= 0 {
		initialCash = 100_000
	}

	cash := initialCash
	shares := 0.0
	entryTime := bars[0].Time
	entryPrice := 0.0
	entryValue := 0.0
	numTrades := 0
	exposureBars := 0
	turnoverValue := 0.0
	tradePNLs := make([]float64, 0)
	trades := make([]models.Trade, 0)
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
					entryTime = bars[i].Time
					entryPrice = bars[i].Open
					entryValue = allocation
					turnoverValue += allocation
					numTrades++
				}
			case models.SignalSell:
				if shares > 0 {
					proceeds := shares * bars[i].Open
					cash += proceeds
					pnl := proceeds - entryValue
					tradePNLs = append(tradePNLs, pnl)
					trades = append(trades, buildClosedTrade(entryTime, bars[i].Time, entryPrice, bars[i].Open, shares, entryValue, proceeds, pnl))
					turnoverValue += proceeds
					shares = 0
					entryPrice = 0
					entryValue = 0
					numTrades++
				}
			}
		}

		equity := cash + shares*bars[i].Close
		if shares > 0 {
			exposureBars++
		}
		equityCurve = append(equityCurve, models.EquityPoint{
			Time:   bars[i].Time,
			Equity: equity,
		})
	}

	if shares > 0 {
		proceeds := shares * bars[len(bars)-1].Close
		cash += proceeds
		pnl := proceeds - entryValue
		tradePNLs = append(tradePNLs, pnl)
		trades = append(trades, buildClosedTrade(entryTime, bars[len(bars)-1].Time, entryPrice, bars[len(bars)-1].Close, shares, entryValue, proceeds, pnl))
		turnoverValue += proceeds
		numTrades++
	}

	equities := make([]float64, len(equityCurve))
	for i, p := range equityCurve {
		equities[i] = p.Equity
	}
	m := ComputeMetrics(equities)
	tradeStats := computeTradeStats(tradePNLs)
	exposurePercent := float64(exposureBars) / float64(len(bars))
	turnover := turnoverValue / initialCash

	return &models.BacktestResult{
		Symbol:           bars[0].Symbol,
		EquityCurve:      equityCurve,
		Trades:           trades,
		FinalEquity:      cash,
		TotalReturn:      m.TotalReturn,
		AnnualizedReturn: m.AnnualizedReturn,
		Sharpe:           m.Sharpe,
		Sortino:          m.Sortino,
		Calmar:           m.Calmar,
		MaxDrawdown:      m.MaxDrawdown,
		NumTrades:        numTrades,
		ProfitFactor:     tradeStats.ProfitFactor,
		WinRate:          tradeStats.WinRate,
		AverageWin:       tradeStats.AverageWin,
		AverageLoss:      tradeStats.AverageLoss,
		AverageTrade:     tradeStats.AverageTrade,
		ExposurePercent:  exposurePercent,
		Turnover:         turnover,
	}, nil
}

func buildClosedTrade(entryTime, exitTime time.Time, entryPrice, exitPrice, quantity, entryValue, exitValue, pnl float64) models.Trade {
	trade := models.Trade{
		EntryTime:  entryTime,
		ExitTime:   exitTime,
		EntryPrice: entryPrice,
		ExitPrice:  exitPrice,
		Quantity:   quantity,
		EntryValue: entryValue,
		ExitValue:  exitValue,
		PnL:        pnl,
	}
	if entryPrice > 0 {
		trade.ReturnPct = (exitPrice / entryPrice) - 1
	}
	return trade
}

// RunBuyAndHold buys the asset at the first executable bar and holds it until
// the backtest engine's final liquidation. It uses the same one-bar execution
// latency as active strategies for a fair comparison.
func RunBuyAndHold(bars []models.Bar, initialCash float64) (*models.BacktestResult, error) {
	if len(bars) < 1 {
		return nil, fmt.Errorf("need at least one bar")
	}
	outputs := make([]StrategyOutput, len(bars))
	outputs[0] = StrategyOutput{
		Signal:          models.SignalBuy,
		PositionSizePct: 1.0,
		RegimeLabel:     "BUY_AND_HOLD",
	}
	for i := 1; i < len(outputs); i++ {
		outputs[i] = StrategyOutput{Signal: models.SignalHold, RegimeLabel: "BUY_AND_HOLD"}
	}
	return RunBacktestWithOutputs(bars, outputs, initialCash)
}

type tradeStats struct {
	ProfitFactor float64
	WinRate      float64
	AverageWin   float64
	AverageLoss  float64
	AverageTrade float64
}

func computeTradeStats(pnls []float64) tradeStats {
	if len(pnls) == 0 {
		return tradeStats{}
	}

	var grossProfit float64
	var grossLoss float64
	var wins int
	var losses int
	var sumWins float64
	var sumLosses float64
	var sum float64
	for _, pnl := range pnls {
		sum += pnl
		if pnl > 0 {
			wins++
			sumWins += pnl
			grossProfit += pnl
		} else if pnl < 0 {
			losses++
			sumLosses += pnl
			grossLoss += -pnl
		}
	}

	stats := tradeStats{
		WinRate:      float64(wins) / float64(len(pnls)),
		AverageTrade: sum / float64(len(pnls)),
	}
	if wins > 0 {
		stats.AverageWin = sumWins / float64(wins)
	}
	if losses > 0 {
		stats.AverageLoss = sumLosses / float64(losses)
	}
	if grossLoss > 0 {
		stats.ProfitFactor = grossProfit / grossLoss
	}
	return stats
}
