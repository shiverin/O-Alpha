package evaluation

import (
	"context"
	"fmt"
	"time"

	"github.com/oalpha/internal/agent"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

type WorkerParityMode string

const (
	WorkerParityOverlay WorkerParityMode = "worker_overlay"
	WorkerParityNoHMM   WorkerParityMode = "worker_none"
)

func (m WorkerParityMode) String() string {
	return string(m)
}

type WorkerParityConfig struct {
	WarmupBars     int
	MaxBars        int
	CalibrateEvery int
	InitialCash    float64
	RiskProfile    agent.RiskProfile
	HMMWindowSize  int
	Symbol         string
	Timeframe      string
}

type WorkerParityResult struct {
	Mode               WorkerParityMode
	Backtest           *models.BacktestResult
	WarmupBars         int
	EvaluatedBars      int
	RegimeDistribution map[string]int
}

type workerParityLot struct {
	Time     time.Time
	Price    float64
	Quantity float64
}

type calibratable interface {
	Calibrate(bars []models.Bar)
}

func DefaultWorkerParityConfig() WorkerParityConfig {
	return WorkerParityConfig{
		WarmupBars:     51,
		MaxBars:        10000,
		CalibrateEvery: 500,
		InitialCash:    100_000,
		RiskProfile:    agent.RiskProfileModerate,
		HMMWindowSize:  50,
	}
}

func RunWorkerParityBacktest(ctx context.Context, bars []models.Bar, mode WorkerParityMode, config WorkerParityConfig) (WorkerParityResult, error) {
	config = config.withDefaults()
	if len(bars) <= config.WarmupBars {
		return WorkerParityResult{}, fmt.Errorf("insufficient bars for worker parity: have %d, need more than warmup %d", len(bars), config.WarmupBars)
	}

	strategy, err := newWorkerParityStrategy(mode, config)
	if err != nil {
		return WorkerParityResult{}, err
	}

	history := append([]models.Bar(nil), bars[:config.WarmupBars]...)
	if c, ok := strategy.(calibratable); ok {
		c.Calibrate(history)
	}
	if _, err := strategy.EvaluateLatest(ctx, history); err != nil {
		return WorkerParityResult{}, fmt.Errorf("worker warmup evaluation failed: %w", err)
	}

	account := agent.NewPaperAccount(config.InitialCash)
	var numTrades int
	var turnoverValue float64
	var exposureBars int
	var barsSinceCalib int
	openLots := make([]workerParityLot, 0)
	trades := make([]models.Trade, 0)
	tradePNLs := make([]float64, 0)
	equityCurve := make([]models.EquityPoint, 0, len(bars)-config.WarmupBars)
	regimeDistribution := make(map[string]int)

	for _, bar := range bars[config.WarmupBars:] {
		history = appendOrUpdateHistoricalBar(history, bar, config.MaxBars)

		barsSinceCalib++
		if config.CalibrateEvery > 0 && barsSinceCalib >= config.CalibrateEvery {
			barsSinceCalib = 0
			if c, ok := strategy.(calibratable); ok {
				c.Calibrate(history)
			}
		}

		output, err := strategy.EvaluateLatest(ctx, history)
		if err != nil {
			return WorkerParityResult{}, fmt.Errorf("worker strategy evaluation failed at %s: %w", bar.Time, err)
		}
		label := output.RegimeLabel
		if label == "" {
			label = "UNKNOWN"
		}
		regimeDistribution[label]++

		switch output.Signal {
		case models.SignalBuy:
			targetAllocation := workerParityTargetAllocation(ctx, account, config.Symbol, bar.Close, output)
			filledQty, cost, err := workerParityBuy(ctx, account, config.Symbol, bar.Close, targetAllocation)
			if err != nil {
				return WorkerParityResult{}, err
			}
			if filledQty > 0 {
				openLots = append(openLots, workerParityLot{Time: bar.Time, Price: bar.Close, Quantity: filledQty})
				turnoverValue += cost
				numTrades++
			}
		case models.SignalSell:
			currentQty := account.GetPosition(config.Symbol)
			if currentQty > 0 {
				amount := currentQty
				filledQty, proceeds, err := account.Sell(ctx, config.Symbol, bar.Close, amount)
				if err != nil {
					return WorkerParityResult{}, fmt.Errorf("worker parity sell: %w", err)
				}
				closedTrades := closeWorkerParityLots(&openLots, filledQty, bar.Time, bar.Close)
				for _, trade := range closedTrades {
					tradePNLs = append(tradePNLs, trade.PnL)
				}
				trades = append(trades, closedTrades...)
				turnoverValue += proceeds
				numTrades++
			}
		}

		equity := account.Equity(ctx, map[string]float64{config.Symbol: bar.Close})
		if account.GetPosition(config.Symbol) > 0 {
			exposureBars++
		}
		equityCurve = append(equityCurve, models.EquityPoint{Time: bar.Time, Equity: equity})
	}

	result := buildWorkerParityBacktestResult(config.Symbol, equityCurve, config.InitialCash, numTrades, tradePNLs, turnoverValue, exposureBars, trades)
	return WorkerParityResult{
		Mode:               mode,
		Backtest:           result,
		WarmupBars:         config.WarmupBars,
		EvaluatedBars:      len(equityCurve),
		RegimeDistribution: regimeDistribution,
	}, nil
}

func RunWorkerParityBuyAndHold(bars []models.Bar, config WorkerParityConfig) (*models.BacktestResult, error) {
	config = config.withDefaults()
	if len(bars) <= config.WarmupBars {
		return nil, fmt.Errorf("insufficient bars for worker buy-and-hold: have %d, need more than warmup %d", len(bars), config.WarmupBars)
	}
	account := agent.NewPaperAccount(config.InitialCash)
	equityCurve := make([]models.EquityPoint, 0, len(bars)-config.WarmupBars)
	var turnoverValue float64
	var numTrades int

	first := bars[config.WarmupBars]
	if first.Close > 0 {
		qty := account.AvailableCash() / first.Close
		if qty > 0 {
			_, cost, err := account.Buy(context.Background(), config.Symbol, first.Close, qty)
			if err != nil {
				return nil, err
			}
			turnoverValue += cost
			numTrades++
		}
	}

	for _, bar := range bars[config.WarmupBars:] {
		equity := account.Equity(context.Background(), map[string]float64{config.Symbol: bar.Close})
		equityCurve = append(equityCurve, models.EquityPoint{Time: bar.Time, Equity: equity})
	}

	return buildWorkerParityBacktestResult(config.Symbol, equityCurve, config.InitialCash, numTrades, nil, turnoverValue, len(equityCurve), nil), nil
}

func newWorkerParityStrategy(mode WorkerParityMode, config WorkerParityConfig) (backtest.Strategy, error) {
	switch mode {
	case WorkerParityOverlay:
		return agent.NewEnsembleDecisionLayer(nil, nil, config.HMMWindowSize, config.RiskProfile), nil
	case WorkerParityNoHMM:
		return agent.NewEnsembleDecisionLayerForMode(nil, nil, config.HMMWindowSize, config.RiskProfile, agent.RegimeModeNone)
	default:
		return nil, fmt.Errorf("unsupported worker parity mode: %s", mode)
	}
}

func appendOrUpdateHistoricalBar(history []models.Bar, bar models.Bar, maxBars int) []models.Bar {
	if len(history) > 0 && !bar.Time.After(history[len(history)-1].Time) {
		history[len(history)-1] = bar
		return history
	}
	history = append(history, bar)
	if maxBars > 0 && len(history) > maxBars {
		history = history[len(history)-maxBars:]
	}
	return history
}

func workerParityBuy(ctx context.Context, account *agent.PaperAccount, symbol string, price float64, targetAllocation float64) (float64, float64, error) {
	availableCash := account.AvailableCash()
	cashToUse := targetAllocation
	if cashToUse <= 0 {
		return 0, 0, nil
	}
	if cashToUse > availableCash {
		cashToUse = availableCash
	}
	if cashToUse < price {
		return 0, 0, nil
	}
	amount := cashToUse / price
	filledQty, cost, err := account.Buy(ctx, symbol, price, amount)
	if err != nil {
		return 0, 0, fmt.Errorf("worker parity buy: %w", err)
	}
	return filledQty, cost, nil
}

func workerParityTargetAllocation(ctx context.Context, account *agent.PaperAccount, symbol string, price float64, output backtest.StrategyOutput) float64 {
	if output.Signal != models.SignalBuy || price <= 0 {
		return 0
	}
	targetWeight := output.TargetWeight
	if targetWeight <= 0 {
		targetWeight = output.PositionSizePct
	}
	targetWeight = normalizePositionSizePct(targetWeight)
	equity := account.Equity(ctx, map[string]float64{symbol: price})
	targetValue := equity * targetWeight
	currentValue := account.GetPosition(symbol) * price
	delta := targetValue - currentValue
	if delta <= 0 {
		return 0
	}
	return delta
}

func closeWorkerParityLots(openLots *[]workerParityLot, quantity float64, exitTime time.Time, exitPrice float64) []models.Trade {
	const epsilon = 1e-9
	remaining := quantity
	lots := *openLots
	trades := make([]models.Trade, 0)

	for remaining > epsilon && len(lots) > 0 {
		lot := lots[0]
		closeQty := lot.Quantity
		if closeQty > remaining {
			closeQty = remaining
		}

		entryValue := closeQty * lot.Price
		exitValue := closeQty * exitPrice
		pnl := exitValue - entryValue
		trade := models.Trade{
			EntryTime:  lot.Time,
			ExitTime:   exitTime,
			EntryPrice: lot.Price,
			ExitPrice:  exitPrice,
			Quantity:   closeQty,
			EntryValue: entryValue,
			ExitValue:  exitValue,
			PnL:        pnl,
		}
		if lot.Price > 0 {
			trade.ReturnPct = (exitPrice / lot.Price) - 1
		}
		trades = append(trades, trade)

		remaining -= closeQty
		lot.Quantity -= closeQty
		if lot.Quantity > epsilon {
			lots[0] = lot
		} else {
			lots = lots[1:]
		}
	}

	*openLots = lots
	return trades
}

func normalizePositionSizePct(sizePct float64) float64 {
	if sizePct <= 0 {
		return 0.1
	}
	if sizePct > 1 {
		return 1
	}
	return sizePct
}

func buildWorkerParityBacktestResult(symbol string, equityCurve []models.EquityPoint, initialCash float64, numTrades int, tradePNLs []float64, turnoverValue float64, exposureBars int, trades []models.Trade) *models.BacktestResult {
	if len(equityCurve) == 0 {
		return &models.BacktestResult{Symbol: symbol}
	}
	equities := make([]float64, len(equityCurve))
	for i, point := range equityCurve {
		equities[i] = point.Equity
	}
	metrics := backtest.ComputeMetrics(equities)
	stats := workerParityTradeStats(tradePNLs)
	return &models.BacktestResult{
		Symbol:           symbol,
		EquityCurve:      equityCurve,
		Trades:           trades,
		FinalEquity:      equityCurve[len(equityCurve)-1].Equity,
		TotalReturn:      metrics.TotalReturn,
		AnnualizedReturn: metrics.AnnualizedReturn,
		Sharpe:           metrics.Sharpe,
		Sortino:          metrics.Sortino,
		Calmar:           metrics.Calmar,
		MaxDrawdown:      metrics.MaxDrawdown,
		NumTrades:        numTrades,
		ProfitFactor:     stats.ProfitFactor,
		WinRate:          stats.WinRate,
		AverageWin:       stats.AverageWin,
		AverageLoss:      stats.AverageLoss,
		AverageTrade:     stats.AverageTrade,
		ExposurePercent:  float64(exposureBars) / float64(len(equityCurve)),
		Turnover:         turnoverValue / initialCash,
	}
}

type workerTradeStats struct {
	ProfitFactor float64
	WinRate      float64
	AverageWin   float64
	AverageLoss  float64
	AverageTrade float64
}

func workerParityTradeStats(pnls []float64) workerTradeStats {
	if len(pnls) == 0 {
		return workerTradeStats{}
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
	stats := workerTradeStats{
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

func (c WorkerParityConfig) withDefaults() WorkerParityConfig {
	defaults := DefaultWorkerParityConfig()
	if c.WarmupBars <= 0 {
		c.WarmupBars = defaults.WarmupBars
	}
	if c.MaxBars <= 0 {
		c.MaxBars = defaults.MaxBars
	}
	if c.CalibrateEvery < 0 {
		c.CalibrateEvery = defaults.CalibrateEvery
	}
	if c.InitialCash <= 0 {
		c.InitialCash = defaults.InitialCash
	}
	if c.RiskProfile < agent.RiskProfileConservative || c.RiskProfile > agent.RiskProfileAggressive {
		c.RiskProfile = defaults.RiskProfile
	}
	if c.HMMWindowSize <= 0 {
		c.HMMWindowSize = defaults.HMMWindowSize
	}
	if c.Symbol == "" {
		c.Symbol = "UNKNOWN"
	}
	return c
}
