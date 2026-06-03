package backtest

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/oalpha/pkg/models"
)

type RebalanceMode string

const (
	RebalanceOnSignal RebalanceMode = "on_signal"
)

type PortfolioBacktestConfig struct {
	InitialCash      float64
	AllowShorts      bool
	MaxGrossExposure float64
	MaxNetExposure   float64
	MaxSymbolWeight  float64
	CostModel        CostModel
	RebalanceMode    RebalanceMode
}

type PortfolioBacktestResult struct {
	Symbols        []string               `json:"symbols"`
	EquityCurve    []models.EquityPoint   `json:"equity_curve"`
	Trades         []SimulatedTrade       `json:"trades"`
	PositionCurve  []PortfolioSnapshot    `json:"position_curve"`
	Metrics        PortfolioMetrics       `json:"metrics"`
	EngineMetadata map[string]interface{} `json:"engine_metadata,omitempty"`
}

type SimulatedTrade struct {
	Time         time.Time              `json:"time"`
	Symbol       string                 `json:"symbol"`
	Side         OrderSide              `json:"side"`
	PositionSide PositionSide           `json:"position_side"`
	Quantity     float64                `json:"quantity"`
	Price        float64                `json:"price"`
	Notional     float64                `json:"notional"`
	Fees         float64                `json:"fees"`
	SpreadCost   float64                `json:"spread_cost"`
	SlippageCost float64                `json:"slippage_cost"`
	RealizedPnL  float64                `json:"realized_pnl"`
	TargetWeight float64                `json:"target_weight"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type PortfolioSnapshot struct {
	Time          time.Time                            `json:"time"`
	Equity        float64                              `json:"equity"`
	Cash          float64                              `json:"cash"`
	GrossExposure float64                              `json:"gross_exposure"`
	NetExposure   float64                              `json:"net_exposure"`
	Positions     map[string]PortfolioPositionSnapshot `json:"positions"`
}

type PortfolioPositionSnapshot struct {
	Symbol           string  `json:"symbol"`
	LongQty          float64 `json:"long_qty"`
	ShortQty         float64 `json:"short_qty"`
	MarketValue      float64 `json:"market_value"`
	GrossMarketValue float64 `json:"gross_market_value"`
	NetWeight        float64 `json:"net_weight"`
}

type portfolioPosition struct {
	LongQty       float64
	LongCost      float64
	ShortQty      float64
	ShortProceeds float64
}

func RunPortfolioBacktest(
	ctx context.Context,
	panel AlignedBars,
	strat PortfolioStrategy,
	cfg PortfolioBacktestConfig,
) (*PortfolioBacktestResult, error) {
	if strat == nil {
		return nil, fmt.Errorf("portfolio strategy is required")
	}
	if err := validatePanel(panel); err != nil {
		return nil, err
	}
	cfg = cfg.withDefaults()

	cash := cfg.InitialCash
	positions := make(map[string]portfolioPosition)
	pendingTargets := make(map[string]TargetPosition)
	trades := make([]SimulatedTrade, 0)
	tradePNLs := make([]float64, 0)
	equityCurve := make([]models.EquityPoint, 0, len(panel.Times))
	positionCurve := make([]PortfolioSnapshot, 0, len(panel.Times))
	grossExposures := make([]float64, 0, len(panel.Times))
	netExposures := make([]float64, 0, len(panel.Times))
	var turnoverValue float64

	for i := range panel.Times {
		if i > 0 && len(pendingTargets) > 0 {
			openPrices := pricesAt(panel, i, true)
			var fills []SimulatedTrade
			var pnls []float64
			cash, positions, fills, pnls = executePortfolioRebalance(panel.Times[i], cash, positions, openPrices, pendingTargets, cfg)
			trades = append(trades, fills...)
			tradePNLs = append(tradePNLs, pnls...)
			for _, fill := range fills {
				turnoverValue += math.Abs(fill.Notional)
			}
			pendingTargets = nil
		}

		closePrices := pricesAt(panel, i, false)
		snapshot := buildPortfolioSnapshot(panel.Times[i], cash, positions, closePrices)
		equityCurve = append(equityCurve, models.EquityPoint{Time: panel.Times[i], Equity: snapshot.Equity})
		positionCurve = append(positionCurve, snapshot)
		grossExposures = append(grossExposures, snapshot.GrossExposure)
		netExposures = append(netExposures, snapshot.NetExposure)

		if i < len(panel.Times)-1 {
			prefix := panelPrefix(panel, i+1)
			output, err := strat.EvaluatePortfolioLatest(ctx, prefix)
			if err != nil {
				return nil, fmt.Errorf("evaluate portfolio strategy at %s: %w", panel.Times[i], err)
			}
			if len(output.Targets) > 0 {
				pendingTargets = sanitizeTargets(output.Targets, cfg)
			}
		}
	}

	equities := make([]float64, len(equityCurve))
	for i, point := range equityCurve {
		equities[i] = point.Equity
	}
	turnover := 0.0
	if cfg.InitialCash > 0 {
		turnover = turnoverValue / cfg.InitialCash
	}

	return &PortfolioBacktestResult{
		Symbols:       append([]string(nil), panel.Symbols...),
		EquityCurve:   equityCurve,
		Trades:        trades,
		PositionCurve: positionCurve,
		Metrics:       ComputePortfolioMetrics(equities, tradePNLs, grossExposures, netExposures, turnover),
		EngineMetadata: map[string]interface{}{
			"cost_model":        cfg.CostModel,
			"allow_shorts":      cfg.AllowShorts,
			"rebalance_mode":    cfg.RebalanceMode,
			"max_gross":         cfg.MaxGrossExposure,
			"max_net":           cfg.MaxNetExposure,
			"max_symbol_weight": cfg.MaxSymbolWeight,
		},
	}, nil
}

func executePortfolioRebalance(
	t time.Time,
	cash float64,
	positions map[string]portfolioPosition,
	openPrices map[string]float64,
	targets map[string]TargetPosition,
	cfg PortfolioBacktestConfig,
) (float64, map[string]portfolioPosition, []SimulatedTrade, []float64) {
	const epsilon = 1e-9
	equity := portfolioEquity(cash, positions, openPrices)
	if equity <= 0 {
		return cash, positions, nil, nil
	}

	trades := make([]SimulatedTrade, 0)
	pnls := make([]float64, 0)
	for _, symbol := range rebalanceSymbols(positions, targets) {
		price := openPrices[symbol]
		if price <= 0 {
			continue
		}
		targetWeight := 0.0
		if target, ok := targets[symbol]; ok {
			targetWeight = target.TargetWeight
		}
		pos := positions[symbol]

		if targetWeight >= 0 {
			if pos.ShortQty > epsilon {
				var fills []SimulatedTrade
				var realized []float64
				cash, pos, fills, realized = coverShort(t, symbol, cash, pos, pos.ShortQty, price, targetWeight, cfg.CostModel)
				trades = append(trades, fills...)
				pnls = append(pnls, realized...)
			}
			targetQty := equity * targetWeight / price
			delta := targetQty - pos.LongQty
			if delta > epsilon {
				fillNotional := delta * price
				fill := cfg.CostModel.ApplyFill(OrderSideBuy, price, fillNotional, symbol)
				fee := fill.Commission + fill.SECFees
				qty := fillNotional / price
				cash -= qty*fill.Price + fee
				pos.LongQty += qty
				pos.LongCost += qty*fill.Price + fee
				trades = append(trades, simulatedFill(t, symbol, OrderSideBuy, PositionSideLong, qty, fill, 0, targetWeight))
			} else if delta < -epsilon {
				sellQty := math.Min(-delta, pos.LongQty)
				var fills []SimulatedTrade
				var realized []float64
				cash, pos, fills, realized = sellLong(t, symbol, cash, pos, sellQty, price, targetWeight, cfg.CostModel)
				trades = append(trades, fills...)
				pnls = append(pnls, realized...)
			}
		} else if cfg.AllowShorts {
			if pos.LongQty > epsilon {
				var fills []SimulatedTrade
				var realized []float64
				cash, pos, fills, realized = sellLong(t, symbol, cash, pos, pos.LongQty, price, targetWeight, cfg.CostModel)
				trades = append(trades, fills...)
				pnls = append(pnls, realized...)
			}
			targetShortQty := equity * math.Abs(targetWeight) / price
			delta := targetShortQty - pos.ShortQty
			if delta > epsilon {
				fillNotional := delta * price
				fill := cfg.CostModel.ApplyFill(OrderSideSell, price, fillNotional, symbol)
				fee := fill.Commission + fill.SECFees
				qty := fillNotional / price
				proceeds := qty*fill.Price - fee
				cash += proceeds
				pos.ShortQty += qty
				pos.ShortProceeds += proceeds
				trades = append(trades, simulatedFill(t, symbol, OrderSideSell, PositionSideShort, qty, fill, 0, targetWeight))
			} else if delta < -epsilon {
				coverQty := math.Min(-delta, pos.ShortQty)
				var fills []SimulatedTrade
				var realized []float64
				cash, pos, fills, realized = coverShort(t, symbol, cash, pos, coverQty, price, targetWeight, cfg.CostModel)
				trades = append(trades, fills...)
				pnls = append(pnls, realized...)
			}
		}

		if pos.LongQty <= epsilon && pos.ShortQty <= epsilon {
			delete(positions, symbol)
		} else {
			positions[symbol] = pos
		}
	}
	return cash, positions, trades, pnls
}

func sellLong(t time.Time, symbol string, cash float64, pos portfolioPosition, qty float64, price float64, targetWeight float64, costModel CostModel) (float64, portfolioPosition, []SimulatedTrade, []float64) {
	if qty <= 0 || pos.LongQty <= 0 {
		return cash, pos, nil, nil
	}
	qty = math.Min(qty, pos.LongQty)
	fillNotional := qty * price
	fill := costModel.ApplyFill(OrderSideSell, price, fillNotional, symbol)
	fee := fill.Commission + fill.SECFees
	proceeds := qty*fill.Price - fee
	closedCost := pos.LongCost * (qty / pos.LongQty)
	pnl := proceeds - closedCost
	cash += proceeds
	pos.LongQty -= qty
	pos.LongCost -= closedCost
	if pos.LongQty <= 1e-9 {
		pos.LongQty = 0
		pos.LongCost = 0
	}
	return cash, pos, []SimulatedTrade{simulatedFill(t, symbol, OrderSideSell, PositionSideLong, qty, fill, pnl, targetWeight)}, []float64{pnl}
}

func coverShort(t time.Time, symbol string, cash float64, pos portfolioPosition, qty float64, price float64, targetWeight float64, costModel CostModel) (float64, portfolioPosition, []SimulatedTrade, []float64) {
	if qty <= 0 || pos.ShortQty <= 0 {
		return cash, pos, nil, nil
	}
	qty = math.Min(qty, pos.ShortQty)
	fillNotional := qty * price
	fill := costModel.ApplyFill(OrderSideBuy, price, fillNotional, symbol)
	fee := fill.Commission + fill.SECFees
	cost := qty*fill.Price + fee
	entryProceeds := pos.ShortProceeds * (qty / pos.ShortQty)
	pnl := entryProceeds - cost
	cash -= cost
	pos.ShortQty -= qty
	pos.ShortProceeds -= entryProceeds
	if pos.ShortQty <= 1e-9 {
		pos.ShortQty = 0
		pos.ShortProceeds = 0
	}
	return cash, pos, []SimulatedTrade{simulatedFill(t, symbol, OrderSideBuy, PositionSideShort, qty, fill, pnl, targetWeight)}, []float64{pnl}
}

func simulatedFill(t time.Time, symbol string, side OrderSide, positionSide PositionSide, qty float64, fill FillPrice, realizedPnL float64, targetWeight float64) SimulatedTrade {
	return SimulatedTrade{
		Time:         t,
		Symbol:       symbol,
		Side:         side,
		PositionSide: positionSide,
		Quantity:     qty,
		Price:        fill.Price,
		Notional:     qty * fill.Price,
		Fees:         fill.Commission + fill.SECFees,
		SpreadCost:   fill.SpreadCost,
		SlippageCost: fill.SlippageCost,
		RealizedPnL:  realizedPnL,
		TargetWeight: targetWeight,
	}
}

func sanitizeTargets(targets map[string]TargetPosition, cfg PortfolioBacktestConfig) map[string]TargetPosition {
	out := make(map[string]TargetPosition, len(targets))
	for symbol, target := range targets {
		weight := clampWeight(target.TargetWeight)
		if !cfg.AllowShorts && weight < 0 {
			weight = 0
		}
		if cfg.MaxSymbolWeight > 0 && math.Abs(weight) > cfg.MaxSymbolWeight {
			weight = math.Copysign(cfg.MaxSymbolWeight, weight)
		}
		target.TargetWeight = weight
		target.Symbol = symbol
		out[symbol] = target
	}
	scaleTargetsForExposure(out, cfg)
	return out
}

func scaleTargetsForExposure(targets map[string]TargetPosition, cfg PortfolioBacktestConfig) {
	var gross, net float64
	for _, target := range targets {
		gross += math.Abs(target.TargetWeight)
		net += target.TargetWeight
	}
	scale := 1.0
	if cfg.MaxGrossExposure > 0 && gross > cfg.MaxGrossExposure {
		scale = math.Min(scale, cfg.MaxGrossExposure/gross)
	}
	if cfg.MaxNetExposure > 0 && math.Abs(net) > cfg.MaxNetExposure {
		scale = math.Min(scale, cfg.MaxNetExposure/math.Abs(net))
	}
	if scale >= 1 {
		return
	}
	for symbol, target := range targets {
		target.TargetWeight *= scale
		targets[symbol] = target
	}
}

func buildPortfolioSnapshot(t time.Time, cash float64, positions map[string]portfolioPosition, prices map[string]float64) PortfolioSnapshot {
	equity := portfolioEquity(cash, positions, prices)
	snapshot := PortfolioSnapshot{
		Time:      t,
		Equity:    equity,
		Cash:      cash,
		Positions: make(map[string]PortfolioPositionSnapshot, len(positions)),
	}
	var grossValue, netValue float64
	for symbol, pos := range positions {
		price := prices[symbol]
		longValue := pos.LongQty * price
		shortValue := pos.ShortQty * price
		net := longValue - shortValue
		gross := longValue + shortValue
		grossValue += gross
		netValue += net
		weight := 0.0
		if equity > 0 {
			weight = net / equity
		}
		snapshot.Positions[symbol] = PortfolioPositionSnapshot{
			Symbol:           symbol,
			LongQty:          pos.LongQty,
			ShortQty:         pos.ShortQty,
			MarketValue:      net,
			GrossMarketValue: gross,
			NetWeight:        weight,
		}
	}
	if equity > 0 {
		snapshot.GrossExposure = grossValue / equity
		snapshot.NetExposure = netValue / equity
	}
	return snapshot
}

func portfolioEquity(cash float64, positions map[string]portfolioPosition, prices map[string]float64) float64 {
	equity := cash
	for symbol, pos := range positions {
		price := prices[symbol]
		equity += pos.LongQty * price
		equity -= pos.ShortQty * price
	}
	return equity
}

func pricesAt(panel AlignedBars, index int, open bool) map[string]float64 {
	prices := make(map[string]float64, len(panel.Symbols))
	for _, symbol := range panel.Symbols {
		bar := panel.Bars[symbol][index]
		if open {
			prices[symbol] = bar.Open
		} else {
			prices[symbol] = bar.Close
		}
	}
	return prices
}

func panelPrefix(panel AlignedBars, length int) AlignedBars {
	prefix := panel
	prefix.Times = append([]time.Time(nil), panel.Times[:length]...)
	prefix.Bars = make(map[string][]models.Bar, len(panel.Bars))
	for symbol, bars := range panel.Bars {
		prefix.Bars[symbol] = append([]models.Bar(nil), bars[:length]...)
	}
	return prefix
}

func rebalanceSymbols(positions map[string]portfolioPosition, targets map[string]TargetPosition) []string {
	seen := make(map[string]struct{}, len(positions)+len(targets))
	symbols := make([]string, 0, len(positions)+len(targets))
	for symbol := range positions {
		seen[symbol] = struct{}{}
		symbols = append(symbols, symbol)
	}
	for symbol := range targets {
		if _, ok := seen[symbol]; ok {
			continue
		}
		symbols = append(symbols, symbol)
	}
	return symbols
}

func validatePanel(panel AlignedBars) error {
	if len(panel.Times) == 0 {
		return fmt.Errorf("panel requires at least one timestamp")
	}
	if len(panel.Symbols) == 0 {
		return fmt.Errorf("panel requires at least one symbol")
	}
	for _, symbol := range panel.Symbols {
		bars := panel.Bars[symbol]
		if len(bars) != len(panel.Times) {
			return fmt.Errorf("symbol %s has %d bars for %d timestamps", symbol, len(bars), len(panel.Times))
		}
	}
	return nil
}

func (c PortfolioBacktestConfig) withDefaults() PortfolioBacktestConfig {
	if c.InitialCash <= 0 {
		c.InitialCash = 100_000
	}
	if c.MaxGrossExposure <= 0 {
		c.MaxGrossExposure = 1
	}
	if c.MaxNetExposure <= 0 {
		c.MaxNetExposure = 1
	}
	if c.MaxSymbolWeight <= 0 {
		c.MaxSymbolWeight = 1
	}
	if c.RebalanceMode == "" {
		c.RebalanceMode = RebalanceOnSignal
	}
	return c
}
