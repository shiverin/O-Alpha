package portfolio

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oalpha/internal/alpaca"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
)

type AlpacaPaperBroker interface {
	PlaceOrder(ctx context.Context, req *alpaca.OrderRequest) (*alpaca.OrderResponse, error)
	GetOrder(ctx context.Context, orderID string) (*alpaca.OrderResponse, error)
	GetOrderByClientOrderID(ctx context.Context, clientOrderID string) (*alpaca.OrderResponse, error)
}

type AlpacaPaperExecutionRouter struct {
	repo             *db.PortfolioRepository
	broker           AlpacaPaperBroker
	userID           int64
	agentRunID       int64
	initialCash      float64
	qtyEpsilon       float64
	minTradeNotional float64
	pollInterval     time.Duration
	pollTimeout      time.Duration
}

func NewAlpacaPaperExecutionRouter(repo *db.PortfolioRepository, broker AlpacaPaperBroker, userID, agentRunID int64, initialCash float64) *AlpacaPaperExecutionRouter {
	if initialCash <= 0 {
		initialCash = 100_000
	}
	return &AlpacaPaperExecutionRouter{
		repo:             repo,
		broker:           broker,
		userID:           userID,
		agentRunID:       agentRunID,
		initialCash:      initialCash,
		qtyEpsilon:       1e-6,
		minTradeNotional: 1.0,
		pollInterval:     500 * time.Millisecond,
		pollTimeout:      15 * time.Second,
	}
}

func (r *AlpacaPaperExecutionRouter) ExecutePortfolioTargets(ctx context.Context, output backtest.PortfolioOutput, prices map[string]float64) error {
	return r.ExecutePortfolioTargetsWithSettings(ctx, output, prices, DefaultRuntimeSettings(output.Time))
}

func (r *AlpacaPaperExecutionRouter) ExecutePortfolioTargetsWithSettings(ctx context.Context, output backtest.PortfolioOutput, prices map[string]float64, settings RuntimeSettings) error {
	if r == nil || r.repo == nil || r.broker == nil {
		return fmt.Errorf("alpaca paper execution router not configured")
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
		if err := r.submitPollAndRecord(ctx, clientOrderID, "SELL_LONG", symbol, sellQty); err != nil {
			log.Printf("[AlpacaPaperExec] sell %s failed: %v", symbol, err)
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
		if err := r.submitPollAndRecord(ctx, clientOrderID, "BUY_LONG", leg.symbol, leg.qty); err != nil {
			log.Printf("[AlpacaPaperExec] buy %s failed: %v", leg.symbol, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		traded++
	}

	if err := r.markAndSnapshot(ctx, prices); err != nil && firstErr == nil {
		firstErr = err
	}
	if traded > 0 {
		_ = r.repo.InsertSystemAlert(ctx, r.userID, "INFO", "Broker paper rebalance executed", fmt.Sprintf("Alpaca paper confirmed %d rebalance order(s) at %s.", traded, output.Time.Format("2006-01-02 15:04")), "portfolio_agent", map[string]interface{}{
			"run_id":   r.agentRunID,
			"legs":     traded,
			"bar_time": output.Time,
			"broker":   "alpaca_paper",
		})
	}
	return firstErr
}

func (r *AlpacaPaperExecutionRouter) markAndSnapshot(ctx context.Context, prices map[string]float64) error {
	var firstErr error
	for symbol, price := range prices {
		if price <= 0 {
			continue
		}
		if err := r.repo.MarkPositionPrice(ctx, r.userID, symbol, price); err != nil {
			log.Printf("[AlpacaPaperExec] mark price %s failed: %v", symbol, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	if err := r.repo.SavePortfolioSnapshot(ctx, r.userID, 0, r.initialCash); err != nil {
		log.Printf("[AlpacaPaperExec] snapshot failed: %v", err)
		if firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (r *AlpacaPaperExecutionRouter) executeRiskExits(ctx context.Context, output backtest.PortfolioOutput, prices map[string]float64, settings RuntimeSettings) (map[string]bool, error) {
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
		if err := r.submitPollAndRecord(ctx, clientOrderID, "SELL_LONG", position.Symbol, position.Qty); err != nil {
			log.Printf("[AlpacaPaperExec] risk exit %s %s failed: %v", reason, position.Symbol, err)
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
			"pnl_pct":         pnlPct,
			"stop_loss_pct":   settings.StopLossPct,
			"take_profit_pct": settings.TakeProfitPct,
			"bar_time":        output.Time,
			"broker":          "alpaca_paper",
		})
	}
	return exited, firstErr
}

func (r *AlpacaPaperExecutionRouter) submitPollAndRecord(ctx context.Context, clientOrderID, action, symbol string, qty float64) error {
	side, err := alpacaPaperLongActionSide(action)
	if err != nil {
		return err
	}
	if qty <= 0 {
		return fmt.Errorf("order quantity must be positive")
	}

	order, lookupErr := r.broker.GetOrderByClientOrderID(ctx, clientOrderID)
	if lookupErr != nil {
		order, err = r.broker.PlaceOrder(ctx, &alpaca.OrderRequest{
			Symbol:        symbol,
			Qty:           &qty,
			Side:          side,
			Type:          "market",
			TimeInForce:   "day",
			ClientOrderID: clientOrderID,
		})
		if err != nil {
			return fmt.Errorf("submit alpaca paper order: %w", err)
		}
	}
	if order == nil || order.ID == "" {
		return fmt.Errorf("alpaca paper order did not return an id")
	}

	order, err = r.pollOrder(ctx, order)
	if err != nil {
		return err
	}
	return r.recordConfirmedFill(ctx, clientOrderID, action, symbol, order)
}

func alpacaPaperLongActionSide(action string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(action)) {
	case "BUY_LONG":
		return "buy", nil
	case "SELL_LONG":
		return "sell", nil
	default:
		return "", fmt.Errorf("unsupported trade action %s", action)
	}
}

func (r *AlpacaPaperExecutionRouter) pollOrder(ctx context.Context, order *alpaca.OrderResponse) (*alpaca.OrderResponse, error) {
	deadline := time.Now().Add(r.pollTimeout)
	for {
		if orderHasRecordableFill(order) || orderIsTerminal(order.Status) {
			return order, nil
		}
		if r.pollTimeout <= 0 || !time.Now().Before(deadline) {
			if filledQty(order) > 0 {
				return order, nil
			}
			return nil, fmt.Errorf("alpaca paper order %s did not fill before timeout; status=%s", order.ID, order.Status)
		}
		wait := r.pollInterval
		if wait <= 0 {
			wait = 10 * time.Millisecond
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
		next, err := r.broker.GetOrder(ctx, order.ID)
		if err != nil {
			return nil, fmt.Errorf("poll alpaca paper order: %w", err)
		}
		order = next
	}
}

func (r *AlpacaPaperExecutionRouter) recordConfirmedFill(ctx context.Context, clientOrderID, action, symbol string, order *alpaca.OrderResponse) error {
	qty := filledQty(order)
	if qty <= 0 {
		status := strings.TrimSpace(order.Status)
		if status == "" {
			status = "unknown"
		}
		if order.RejectedReason != "" {
			return fmt.Errorf("alpaca paper order %s produced no fill; status=%s reason=%s", order.ID, status, order.RejectedReason)
		}
		return fmt.Errorf("alpaca paper order %s produced no fill; status=%s", order.ID, status)
	}
	price := filledAvgPrice(order)
	if price <= 0 {
		return fmt.Errorf("alpaca paper order %s filled %.8f %s without a valid average price", order.ID, qty, symbol)
	}

	if err := r.repo.RecordLongFillKeyed(ctx, r.userID, r.agentRunID, clientOrderID, action, symbol, price, qty, 0); err != nil {
		return fmt.Errorf("mirror alpaca paper fill: %w", err)
	}
	if !strings.EqualFold(order.Status, "filled") {
		_ = r.repo.InsertSystemAlert(ctx, r.userID, "WARNING", "Partial broker paper fill", fmt.Sprintf("Alpaca paper reported %.8f shares filled for %s order %s with status %s.", qty, symbol, order.ID, order.Status), "portfolio_agent", map[string]interface{}{
			"run_id":            r.agentRunID,
			"symbol":            symbol,
			"provider_order_id": order.ID,
			"client_order_id":   clientOrderID,
			"status":            order.Status,
			"filled_qty":        qty,
			"broker":            "alpaca_paper",
		})
	}
	return nil
}

func orderHasRecordableFill(order *alpaca.OrderResponse) bool {
	return order != nil && strings.EqualFold(order.Status, "filled") && filledQty(order) > 0 && filledAvgPrice(order) > 0
}

func orderIsTerminal(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "filled", "canceled", "expired", "rejected", "done_for_day":
		return true
	default:
		return false
	}
}

func filledQty(order *alpaca.OrderResponse) float64 {
	if order == nil {
		return 0
	}
	return parseOrderFloat(order.FilledQty)
}

func filledAvgPrice(order *alpaca.OrderResponse) float64 {
	if order == nil {
		return 0
	}
	return parseOrderFloat(order.FilledAvgPrice)
}

func parseOrderFloat(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}
