package portfolio

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
)

type DBExecutionRouter struct {
	repo             *db.PortfolioRepository
	userID           int64
	agentRunID       int64
	initialCash      float64
	qtyEpsilon       float64
	minTradeNotional float64
}

func NewDBExecutionRouter(repo *db.PortfolioRepository, userID, agentRunID int64, initialCash float64) *DBExecutionRouter {
	if initialCash <= 0 {
		initialCash = 100_000
	}
	return &DBExecutionRouter{
		repo:             repo,
		userID:           userID,
		agentRunID:       agentRunID,
		initialCash:      initialCash,
		qtyEpsilon:       1e-6,
		minTradeNotional: 1.0,
	}
}

func (r *DBExecutionRouter) ExecutePortfolioTargets(ctx context.Context, output backtest.PortfolioOutput, prices map[string]float64) error {
	return r.ExecutePortfolioTargetsWithSettings(ctx, output, prices, DefaultRuntimeSettings(output.Time))
}

func (r *DBExecutionRouter) ExecutePortfolioTargetsWithSettings(ctx context.Context, output backtest.PortfolioOutput, prices map[string]float64, settings RuntimeSettings) error {
	if r == nil || r.repo == nil {
		return fmt.Errorf("execution router not configured")
	}

	riskExited, firstErr := r.executeRiskExits(ctx, output, prices, settings)
	if output.EngineMetadata != nil {
		if due, ok := output.EngineMetadata[runtimeCadenceMetadataKey].(bool); ok && !due {
			if err := r.markAndSnapshot(ctx, prices); err != nil && firstErr == nil {
				firstErr = err
			}
			return firstErr
		}
	}
	if len(output.Targets) == 0 {
		if err := r.markAndSnapshot(ctx, prices); err != nil && firstErr == nil {
			firstErr = err
		}
		return firstErr
	}

	cash, positions, err := r.repo.GetAccountState(ctx, r.userID)
	if err != nil {
		return fmt.Errorf("read account state: %w", err)
	}

	equity := cash
	for symbol, qty := range positions {
		equity += qty * prices[symbol]
	}
	if equity <= 0 {
		equity = r.initialCash
	}

	desired := make(map[string]float64, len(output.Targets))
	for symbol, target := range output.Targets {
		if riskExited[symbol] {
			continue
		}
		if target.Side == backtest.PositionSideShort || target.TargetWeight <= 0 {
			continue
		}
		price := prices[symbol]
		if price <= 0 {
			continue
		}
		desired[symbol] = (equity * target.TargetWeight) / price
	}

	barUnix := output.Time.Unix()
	if barUnix <= 0 {
		barUnix = 0
	}
	traded := 0

	for symbol, currentQty := range positions {
		targetQty := desired[symbol]
		if targetQty >= currentQty-r.qtyEpsilon {
			continue
		}
		price := prices[symbol]
		if price <= 0 {
			continue
		}
		sellQty := currentQty - targetQty
		if sellQty*price < r.minTradeNotional {
			continue
		}
		clientOrderID := fmt.Sprintf("%d:%s:%d:SELL", r.agentRunID, symbol, barUnix)
		if err := r.repo.RecordLongFillKeyed(ctx, r.userID, r.agentRunID, clientOrderID, "SELL_LONG", symbol, price, sellQty, 0); err != nil {
			log.Printf("[PortfolioExec] sell %s failed: %v", symbol, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		traded++
	}

	type buyLeg struct {
		symbol string
		qty    float64
		weight float64
	}
	buys := make([]buyLeg, 0, len(desired))
	for symbol, targetQty := range desired {
		delta := targetQty - positions[symbol]
		if delta <= r.qtyEpsilon {
			continue
		}
		buys = append(buys, buyLeg{
			symbol: symbol,
			qty:    delta,
			weight: output.Targets[symbol].TargetWeight,
		})
	}
	sort.Slice(buys, func(i, j int) bool { return buys[i].weight > buys[j].weight })

	for _, leg := range buys {
		price := prices[leg.symbol]
		if price <= 0 || leg.qty*price < r.minTradeNotional {
			continue
		}
		clientOrderID := fmt.Sprintf("%d:%s:%d:BUY", r.agentRunID, leg.symbol, barUnix)
		if err := r.repo.RecordLongFillKeyed(ctx, r.userID, r.agentRunID, clientOrderID, "BUY_LONG", leg.symbol, price, leg.qty, 0); err != nil {
			log.Printf("[PortfolioExec] buy %s failed: %v", leg.symbol, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		traded++
	}

	if err := r.markAndSnapshot(ctx, prices); err != nil {
		if firstErr == nil {
			firstErr = err
		}
	}

	if traded > 0 {
		_ = r.repo.InsertSystemAlert(ctx, r.userID, "INFO", "Rebalance executed", fmt.Sprintf("Agent rebalanced %d position(s) at %s.", traded, output.Time.Format("2006-01-02 15:04")), "portfolio_agent", map[string]interface{}{
			"run_id":   r.agentRunID,
			"legs":     traded,
			"bar_time": output.Time,
		})
	}

	return firstErr
}

func (r *DBExecutionRouter) markAndSnapshot(ctx context.Context, prices map[string]float64) error {
	var firstErr error
	for symbol, price := range prices {
		if price <= 0 {
			continue
		}
		if err := r.repo.MarkPositionPrice(ctx, r.userID, symbol, price); err != nil {
			log.Printf("[PortfolioExec] mark price %s failed: %v", symbol, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	if err := r.repo.SavePortfolioSnapshot(ctx, r.userID, 0, r.initialCash); err != nil {
		log.Printf("[PortfolioExec] snapshot failed: %v", err)
		if firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (r *DBExecutionRouter) executeRiskExits(ctx context.Context, output backtest.PortfolioOutput, prices map[string]float64, settings RuntimeSettings) (map[string]bool, error) {
	exited := make(map[string]bool)
	positions, err := r.repo.GetActivePositions(ctx, r.userID)
	if err != nil {
		return exited, fmt.Errorf("read active positions for risk exits: %w", err)
	}
	if len(positions) == 0 {
		return exited, nil
	}

	barUnix := output.Time.Unix()
	if barUnix <= 0 {
		barUnix = 0
	}

	var firstErr error
	for _, position := range positions {
		price := prices[position.Symbol]
		if price <= 0 {
			price = position.CurrentPrice
		}
		if price <= 0 || position.AvgEntryPrice <= 0 || position.Qty <= 0 {
			continue
		}

		pnlPct := ((price - position.AvgEntryPrice) / position.AvgEntryPrice) * 100
		reason := riskExitReason(pnlPct, settings)
		if reason == "" {
			continue
		}

		clientOrderID := fmt.Sprintf("%d:%s:%d:RISK_EXIT:%s", r.agentRunID, position.Symbol, barUnix, reason)
		if err := r.repo.RecordLongFillKeyed(ctx, r.userID, r.agentRunID, clientOrderID, "SELL_LONG", position.Symbol, price, position.Qty, 0); err != nil {
			log.Printf("[PortfolioExec] risk exit %s %s failed: %v", reason, position.Symbol, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		exited[position.Symbol] = true

		_ = r.repo.InsertSystemAlert(ctx, r.userID, "WARNING", riskExitTitle(reason), riskExitDescription(position.Symbol, pnlPct, reason), "portfolio_agent", map[string]interface{}{
			"run_id":          r.agentRunID,
			"symbol":          position.Symbol,
			"reason":          reason,
			"pnl_pct":         math.Round(pnlPct*100) / 100,
			"stop_loss_pct":   settings.StopLossPct,
			"take_profit_pct": settings.TakeProfitPct,
			"bar_time":        output.Time,
		})
	}
	return exited, firstErr
}

func riskExitReason(pnlPct float64, settings RuntimeSettings) string {
	switch {
	case settings.StopLossPct > 0 && pnlPct <= -settings.StopLossPct:
		return "stop_loss"
	case settings.TakeProfitPct > 0 && pnlPct >= settings.TakeProfitPct:
		return "take_profit"
	default:
		return ""
	}
}

func riskExitTitle(reason string) string {
	switch reason {
	case "take_profit":
		return "Take-profit exit triggered"
	default:
		return "Stop-loss exit triggered"
	}
}

func riskExitDescription(symbol string, pnlPct float64, reason string) string {
	switch reason {
	case "take_profit":
		return fmt.Sprintf("%s was closed after reaching %.2f%% unrealized P&L.", symbol, pnlPct)
	default:
		return fmt.Sprintf("%s was closed after reaching %.2f%% unrealized P&L.", symbol, pnlPct)
	}
}
