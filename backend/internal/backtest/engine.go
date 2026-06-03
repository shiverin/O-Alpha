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
	costBasis := 0.0
	numTrades := 0
	exposureBars := 0
	turnoverValue := 0.0
	tradePNLs := make([]float64, 0)
	trades := make([]models.Trade, 0)
	equityCurve := make([]models.EquityPoint, 0, len(bars))

	for i := 0; i < len(bars); i++ {
		// Execute pending signal from previous bar at this bar's open.
		if i > 0 {
			cash, shares, costBasis, entryTime, entryPrice, numTrades, turnoverValue, tradePNLs, trades = executeSingleSymbolTarget(
				bars[i],
				outputs[i-1],
				cash,
				shares,
				costBasis,
				entryTime,
				entryPrice,
				numTrades,
				turnoverValue,
				tradePNLs,
				trades,
			)
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
		pnl := proceeds - costBasis
		tradePNLs = append(tradePNLs, pnl)
		trades = append(trades, buildClosedTrade(entryTime, bars[len(bars)-1].Time, entryPrice, bars[len(bars)-1].Close, shares, costBasis, proceeds, pnl))
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

func executeSingleSymbolTarget(
	bar models.Bar,
	output StrategyOutput,
	cash float64,
	shares float64,
	costBasis float64,
	entryTime time.Time,
	entryPrice float64,
	numTrades int,
	turnoverValue float64,
	tradePNLs []float64,
	trades []models.Trade,
) (float64, float64, float64, time.Time, float64, int, float64, []float64, []models.Trade) {
	const epsilon = 1e-9
	if bar.Open <= 0 {
		return cash, shares, costBasis, entryTime, entryPrice, numTrades, turnoverValue, tradePNLs, trades
	}

	var targetWeight float64
	var shouldRebalance bool
	switch output.Signal {
	case models.SignalBuy:
		targetWeight = normalizeSingleSymbolTargetWeight(output)
		shouldRebalance = true
	case models.SignalSell:
		targetWeight = 0
		shouldRebalance = true
	default:
		if output.TargetWeight != 0 {
			targetWeight = normalizeSingleSymbolTargetWeight(output)
			shouldRebalance = true
		}
	}
	if !shouldRebalance {
		return cash, shares, costBasis, entryTime, entryPrice, numTrades, turnoverValue, tradePNLs, trades
	}

	equityAtOpen := cash + shares*bar.Open
	if equityAtOpen <= 0 {
		return cash, shares, costBasis, entryTime, entryPrice, numTrades, turnoverValue, tradePNLs, trades
	}
	targetShares := equityAtOpen * targetWeight / bar.Open
	deltaShares := targetShares - shares
	if deltaShares > epsilon {
		cost := deltaShares * bar.Open
		if cost > cash {
			cost = cash
			deltaShares = cost / bar.Open
		}
		if deltaShares <= epsilon {
			return cash, shares, costBasis, entryTime, entryPrice, numTrades, turnoverValue, tradePNLs, trades
		}
		if shares <= epsilon {
			entryTime = bar.Time
			entryPrice = bar.Open
		}
		cash -= cost
		shares += deltaShares
		costBasis += cost
		turnoverValue += cost
		numTrades++
		return cash, shares, costBasis, entryTime, entryPrice, numTrades, turnoverValue, tradePNLs, trades
	}

	if deltaShares < -epsilon && shares > epsilon {
		sellQty := -deltaShares
		if sellQty > shares {
			sellQty = shares
		}
		proceeds := sellQty * bar.Open
		closedCost := costBasis * (sellQty / shares)
		pnl := proceeds - closedCost
		cash += proceeds
		shares -= sellQty
		costBasis -= closedCost
		if shares <= epsilon {
			shares = 0
			costBasis = 0
		}
		tradePNLs = append(tradePNLs, pnl)
		trades = append(trades, buildClosedTrade(entryTime, bar.Time, entryPrice, bar.Open, sellQty, closedCost, proceeds, pnl))
		turnoverValue += proceeds
		numTrades++
	}
	return cash, shares, costBasis, entryTime, entryPrice, numTrades, turnoverValue, tradePNLs, trades
}

func normalizeSingleSymbolTargetWeight(output StrategyOutput) float64 {
	weight := output.TargetWeight
	if weight == 0 {
		weight = output.PositionSizePct
	}
	if weight <= 0 {
		return 1
	}
	if weight > 1 {
		return 1
	}
	return weight
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
