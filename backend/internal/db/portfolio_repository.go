package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultPaperInitialCash = 100000.0

type TradeLog struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Qty       float64   `json:"qty"`
	Slippage  float64   `json:"slippage"`
	Status    string    `json:"status"`
}

type PositionMetrics struct {
	Symbol        string  `json:"symbol"`
	Qty           float64 `json:"qty"`
	AvgEntryPrice float64 `json:"avg_entry_price"`
	CurrentPrice  float64 `json:"current_price"`
	UnrealizedPnL float64 `json:"unrealized_pnl"`
	Exposure      float64 `json:"exposure"`
}

type PortfolioSummary struct {
	TotalAssetValue       float64   `json:"total_asset_value"`
	ChangePercent24h      float64   `json:"change_percent_24h"`
	ChangeDollar24h       float64   `json:"change_dollar_24h"`
	EstimatedAnnualYield  float64   `json:"estimated_annual_yield"`
	TargetProgressPercent float64   `json:"target_progress_percent"`
	Timestamp             time.Time `json:"timestamp"`
}

type SystemAlert struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AlertType   string    `json:"alert_type"`
	CreatedAt   time.Time `json:"created_at"`
}

type PortfolioRepository struct {
	db *pgxpool.Pool
}

type accountBalance struct {
	ID          int64
	Cash        float64
	RealizedPnL float64
	InitialCash float64
}

func NewPortfolioRepository(db *pgxpool.Pool) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

// RecordLongFill stores a filled long-side order and updates cash, positions, and ledger entries atomically.
func (r *PortfolioRepository) RecordLongFill(ctx context.Context, userID, agentRunID int64, action, symbol string, price, qty, slippage float64) error {
	symbol = normalizeSymbol(symbol)
	action = strings.ToUpper(strings.TrimSpace(action))
	if symbol == "" {
		return fmt.Errorf("fill symbol is required")
	}
	if price <= 0 {
		return fmt.Errorf("fill price must be positive")
	}
	if qty <= 0 {
		return fmt.Errorf("fill quantity must be positive")
	}

	side, err := longActionSide(action)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin portfolio fill transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := ensureAssetTx(ctx, tx, symbol); err != nil {
		return err
	}

	account, err := ensureDefaultPaperAccountTx(ctx, tx, userID, defaultPaperInitialCash)
	if err != nil {
		return err
	}

	var agentRunIDArg *int64
	if agentRunID > 0 {
		agentRunIDArg = &agentRunID
	}

	var orderID int64
	const orderQ = `
		INSERT INTO orders (
			user_id, account_id, agent_run_id, symbol, side, position_side,
			order_type, time_in_force, qty, status, submitted_at, filled_at
		)
		VALUES ($1, $2, $3, $4, $5, 'long', 'market', 'day', $6, 'filled', NOW(), NOW())
		RETURNING id`
	if err := tx.QueryRow(ctx, orderQ, userID, account.ID, agentRunIDArg, symbol, side, qty).Scan(&orderID); err != nil {
		return fmt.Errorf("insert filled order: %w", err)
	}

	var fillID int64
	const fillQ = `
		INSERT INTO fills (
			order_id, user_id, account_id, symbol, side, position_side,
			price, qty, slippage
		)
		VALUES ($1, $2, $3, $4, $5, 'long', $6, $7, $8)
		RETURNING id`
	if err := tx.QueryRow(ctx, fillQ, orderID, userID, account.ID, symbol, side, price, qty, slippage).Scan(&fillID); err != nil {
		return fmt.Errorf("insert fill: %w", err)
	}

	gross := price * qty
	switch action {
	case "BUY_LONG":
		if account.Cash < gross {
			return fmt.Errorf("insufficient cash for %s buy: need %.2f, have %.2f", symbol, gross, account.Cash)
		}
		newCash := account.Cash - gross
		if err := updateAccountCashTx(ctx, tx, userID, account.ID, newCash, 0); err != nil {
			return err
		}
		if err := insertCashLedgerTx(ctx, tx, userID, account.ID, "trade_buy", -gross, newCash, orderID, fillID, fmt.Sprintf("Bought %.8f %s", qty, symbol)); err != nil {
			return err
		}
		if err := upsertBoughtPositionTx(ctx, tx, userID, account.ID, symbol, price, qty); err != nil {
			return err
		}
	case "SELL_LONG":
		realizedPnL, err := reduceSoldPositionTx(ctx, tx, account.ID, symbol, price, qty)
		if err != nil {
			return err
		}
		newCash := account.Cash + gross
		if err := updateAccountCashTx(ctx, tx, userID, account.ID, newCash, realizedPnL); err != nil {
			return err
		}
		if err := insertCashLedgerTx(ctx, tx, userID, account.ID, "trade_sell", gross, newCash, orderID, fillID, fmt.Sprintf("Sold %.8f %s", qty, symbol)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported trade action %s", action)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit portfolio fill transaction: %w", err)
	}
	return nil
}

// MarkPositionPrice updates the current price used for unrealized P&L on open positions.
func (r *PortfolioRepository) MarkPositionPrice(ctx context.Context, userID int64, symbol string, currentPrice float64) error {
	symbol = normalizeSymbol(symbol)
	if symbol == "" {
		return fmt.Errorf("position symbol is required")
	}
	if currentPrice < 0 {
		return fmt.Errorf("position price cannot be negative")
	}

	const q = `
		UPDATE positions
		SET current_price = $3
		WHERE user_id = $1 AND symbol = $2 AND qty > 0`
	if _, err := r.db.Exec(ctx, q, userID, symbol, currentPrice); err != nil {
		return fmt.Errorf("mark position price: %w", err)
	}
	return nil
}

// SavePortfolioSnapshot records canonical account equity from persisted cash and positions.
func (r *PortfolioRepository) SavePortfolioSnapshot(ctx context.Context, userID int64, _ float64, targetValue float64) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin portfolio snapshot transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	account, err := ensureDefaultPaperAccountTx(ctx, tx, userID, defaultPaperInitialCash)
	if err != nil {
		return err
	}

	var positionsValue, unrealizedPnL float64
	const positionsQ = `
		SELECT
			COALESCE(SUM(qty * current_price), 0),
			COALESCE(SUM((current_price - avg_entry_price) * qty), 0)
		FROM positions
		WHERE account_id = $1 AND position_side = 'long' AND qty > 0`
	if err := tx.QueryRow(ctx, positionsQ, account.ID).Scan(&positionsValue, &unrealizedPnL); err != nil {
		return fmt.Errorf("query portfolio mark-to-market: %w", err)
	}

	totalAssetValue := account.Cash + positionsValue
	previousValue, err := snapshotComparisonValueTx(ctx, tx, account.ID)
	if err != nil {
		return err
	}

	changeDollar := 0.0
	changePercent := 0.0
	if previousValue > 0 {
		changeDollar = totalAssetValue - previousValue
		changePercent = (changeDollar / previousValue) * 100
	}
	changePercent = clampNumericPct(changePercent)

	if targetValue <= 0 {
		targetValue = account.InitialCash
	}
	targetProgress := 0.0
	if targetValue > 0 {
		targetProgress = (totalAssetValue / targetValue) * 100
	}
	targetProgress = clampNumericPct(targetProgress)

	const q = `
		INSERT INTO portfolio_snapshots (
			user_id,
			account_id,
			cash_value,
			positions_value,
			total_asset_value,
			realized_pnl,
			unrealized_pnl,
			change_percent_24h,
			change_dollar_24h,
			estimated_annual_yield,
			target_progress_percent
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	if _, err := tx.Exec(
		ctx,
		q,
		userID,
		account.ID,
		account.Cash,
		positionsValue,
		totalAssetValue,
		account.RealizedPnL,
		unrealizedPnL,
		changePercent,
		changeDollar,
		0,
		targetProgress,
	); err != nil {
		return fmt.Errorf("insert portfolio snapshot: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit portfolio snapshot transaction: %w", err)
	}
	return nil
}

func snapshotComparisonValueTx(ctx context.Context, tx pgx.Tx, accountID int64) (float64, error) {
	const dayOldQ = `
		SELECT total_asset_value
		FROM portfolio_snapshots
		WHERE account_id = $1 AND timestamp <= NOW() - INTERVAL '24 hours'
		ORDER BY timestamp DESC
		LIMIT 1`

	var value float64
	err := tx.QueryRow(ctx, dayOldQ, accountID).Scan(&value)
	if err == nil {
		return value, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("query 24h portfolio baseline: %w", err)
	}

	const latestQ = `
		SELECT total_asset_value
		FROM portfolio_snapshots
		WHERE account_id = $1
		ORDER BY timestamp DESC
		LIMIT 1`
	err = tx.QueryRow(ctx, latestQ, accountID).Scan(&value)
	if err == nil {
		return value, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return 0, fmt.Errorf("query latest portfolio baseline: %w", err)
}

func clampNumericPct(value float64) float64 {
	if value > 999.99 {
		return 999.99
	}
	if value < -999.99 {
		return -999.99
	}
	return value
}

// GetExecutionStream returns the latest persisted fills with order status.
func (r *PortfolioRepository) GetExecutionStream(ctx context.Context, userID int64, limit int) ([]TradeLog, error) {
	const q = `
		SELECT
			f.id,
			f.filled_at,
			CASE
				WHEN f.side = 'buy' AND f.position_side = 'long' THEN 'BUY_LONG'
				WHEN f.side = 'sell' AND f.position_side = 'long' THEN 'SELL_LONG'
				WHEN f.side = 'sell' AND f.position_side = 'short' THEN 'SELL_SHORT'
				WHEN f.side = 'buy' AND f.position_side = 'short' THEN 'COVER_SHORT'
				ELSE upper(f.side)
			END,
			f.symbol,
			f.price,
			f.qty,
			f.slippage,
			upper(o.status)
		FROM fills f
		JOIN orders o ON o.id = f.order_id
		WHERE f.user_id = $1
		ORDER BY f.filled_at DESC
		LIMIT $2`

	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("query execution stream: %w", err)
	}
	defer rows.Close()

	var stream []TradeLog
	for rows.Next() {
		var t TradeLog
		if err := rows.Scan(&t.ID, &t.Timestamp, &t.Action, &t.Symbol, &t.Price, &t.Qty, &t.Slippage, &t.Status); err != nil {
			return nil, err
		}
		stream = append(stream, t)
	}
	return stream, rows.Err()
}

// GetActivePositions returns open positions with derived exposure and P&L.
func (r *PortfolioRepository) GetActivePositions(ctx context.Context, userID int64) ([]PositionMetrics, error) {
	const q = `
		SELECT symbol, qty, avg_entry_price, current_price
		FROM positions
		WHERE user_id = $1 AND qty > 0
		ORDER BY updated_at DESC`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("query positions failed: %w", err)
	}
	defer rows.Close()

	var activePositions []PositionMetrics
	for rows.Next() {
		var p PositionMetrics
		if err := rows.Scan(&p.Symbol, &p.Qty, &p.AvgEntryPrice, &p.CurrentPrice); err != nil {
			return nil, err
		}
		p.Exposure = p.Qty * p.CurrentPrice
		p.UnrealizedPnL = (p.CurrentPrice - p.AvgEntryPrice) * p.Qty
		activePositions = append(activePositions, p)
	}
	return activePositions, rows.Err()
}

// GetLatestSummary returns the latest portfolio snapshot.
func (r *PortfolioRepository) GetLatestSummary(ctx context.Context, userID int64) (*PortfolioSummary, error) {
	const q = `
		SELECT total_asset_value, change_percent_24h, change_dollar_24h, estimated_annual_yield, target_progress_percent, timestamp
		FROM portfolio_snapshots
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT 1`

	var s PortfolioSummary
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&s.TotalAssetValue, &s.ChangePercent24h, &s.ChangeDollar24h, &s.EstimatedAnnualYield, &s.TargetProgressPercent, &s.Timestamp,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query snapshot failed: %w", err)
	}
	return &s, nil
}

// GetSystemAlerts returns recent risk alerts.
func (r *PortfolioRepository) GetSystemAlerts(ctx context.Context, userID int64, limit int) ([]SystemAlert, error) {
	const q = `
		SELECT id, title, description, alert_type, created_at
		FROM system_alerts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("query system alerts failed: %w", err)
	}
	defer rows.Close()

	var alerts []SystemAlert
	for rows.Next() {
		var a SystemAlert
		if err := rows.Scan(&a.ID, &a.Title, &a.Description, &a.AlertType, &a.CreatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// GetPortfolioHistory retrieves the latest snapshots sorted chronologically for charting.
func (r *PortfolioRepository) GetPortfolioHistory(ctx context.Context, userID int64, limit int) ([]PortfolioSummary, error) {
	const q = `
		SELECT total_asset_value, change_percent_24h, change_dollar_24h, estimated_annual_yield, target_progress_percent, timestamp
		FROM (
			SELECT total_asset_value, change_percent_24h, change_dollar_24h, estimated_annual_yield, target_progress_percent, timestamp
			FROM portfolio_snapshots
			WHERE user_id = $1
			ORDER BY timestamp DESC
			LIMIT $2
		) sub
		ORDER BY timestamp ASC`

	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("query portfolio history failed: %w", err)
	}
	defer rows.Close()

	history := make([]PortfolioSummary, 0)
	for rows.Next() {
		var s PortfolioSummary
		err := rows.Scan(
			&s.TotalAssetValue,
			&s.ChangePercent24h,
			&s.ChangeDollar24h,
			&s.EstimatedAnnualYield,
			&s.TargetProgressPercent,
			&s.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("scan portfolio snapshot row failed: %w", err)
		}
		history = append(history, s)
	}

	return history, rows.Err()
}

func longActionSide(action string) (string, error) {
	switch action {
	case "BUY_LONG":
		return "buy", nil
	case "SELL_LONG":
		return "sell", nil
	default:
		return "", fmt.Errorf("unsupported trade action %s", action)
	}
}

func ensureAssetTx(ctx context.Context, tx pgx.Tx, symbol string) error {
	symbol = normalizeSymbol(symbol)
	if symbol == "" {
		return fmt.Errorf("asset symbol is required")
	}

	const q = `
		INSERT INTO assets (symbol, name, asset_class)
		VALUES ($1, $1, 'equity')
		ON CONFLICT (symbol) DO NOTHING`
	if _, err := tx.Exec(ctx, q, symbol); err != nil {
		return fmt.Errorf("upsert asset %s: %w", symbol, err)
	}
	return nil
}

func ensureDefaultPaperAccountTx(ctx context.Context, tx pgx.Tx, userID int64, initialCash float64) (*accountBalance, error) {
	if initialCash <= 0 {
		initialCash = defaultPaperInitialCash
	}

	const selectQ = `
		SELECT id, cash_balance, realized_pnl, initial_cash
		FROM accounts
		WHERE user_id = $1
			AND account_type = 'paper'
			AND provider = 'internal'
			AND provider_account_id = ''
		ORDER BY id
		LIMIT 1
		FOR UPDATE`

	var account accountBalance
	err := tx.QueryRow(ctx, selectQ, userID).Scan(&account.ID, &account.Cash, &account.RealizedPnL, &account.InitialCash)
	if err == nil {
		return &account, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("query paper account: %w", err)
	}

	const insertQ = `
		INSERT INTO accounts (
			user_id, account_type, provider, provider_account_id,
			initial_cash, cash_balance
		)
		VALUES ($1, 'paper', 'internal', '', $2, $2)
		RETURNING id, cash_balance, realized_pnl, initial_cash`
	if err := tx.QueryRow(ctx, insertQ, userID, initialCash).Scan(&account.ID, &account.Cash, &account.RealizedPnL, &account.InitialCash); err != nil {
		return nil, fmt.Errorf("insert paper account: %w", err)
	}

	const ledgerQ = `
		INSERT INTO cash_ledger (
			user_id, account_id, event_type, amount, balance_after, description
		)
		VALUES ($1, $2, 'initial_deposit', $3, $3, 'Initial paper cash provisioning')`
	if _, err := tx.Exec(ctx, ledgerQ, userID, account.ID, account.Cash); err != nil {
		return nil, fmt.Errorf("insert initial cash ledger entry: %w", err)
	}

	return &account, nil
}

func updateAccountCashTx(ctx context.Context, tx pgx.Tx, userID, accountID int64, newCash, realizedPnLDelta float64) error {
	const q = `
		UPDATE accounts
		SET cash_balance = $3,
			realized_pnl = realized_pnl + $4
		WHERE id = $1 AND user_id = $2`
	tag, err := tx.Exec(ctx, q, accountID, userID, newCash, realizedPnLDelta)
	if err != nil {
		return fmt.Errorf("update account cash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("paper account %d was not found for user %d", accountID, userID)
	}
	return nil
}

func insertCashLedgerTx(ctx context.Context, tx pgx.Tx, userID, accountID int64, eventType string, amount, balanceAfter float64, orderID, fillID int64, description string) error {
	const q = `
		INSERT INTO cash_ledger (
			user_id, account_id, event_type, amount, balance_after,
			related_order_id, related_fill_id, description
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := tx.Exec(ctx, q, userID, accountID, eventType, amount, balanceAfter, orderID, fillID, description); err != nil {
		return fmt.Errorf("insert cash ledger entry: %w", err)
	}
	return nil
}

func upsertBoughtPositionTx(ctx context.Context, tx pgx.Tx, userID, accountID int64, symbol string, price, qty float64) error {
	const q = `
		INSERT INTO positions (
			user_id, account_id, symbol, position_side, qty,
			avg_entry_price, current_price
		)
		VALUES ($1, $2, $3, 'long', $4, $5, $5)
		ON CONFLICT (account_id, symbol, position_side) DO UPDATE SET
			qty = positions.qty + EXCLUDED.qty,
			avg_entry_price = CASE
				WHEN positions.qty + EXCLUDED.qty <= 0 THEN EXCLUDED.avg_entry_price
				ELSE ((positions.qty * positions.avg_entry_price) + (EXCLUDED.qty * EXCLUDED.avg_entry_price)) / (positions.qty + EXCLUDED.qty)
			END,
			current_price = EXCLUDED.current_price`
	if _, err := tx.Exec(ctx, q, userID, accountID, symbol, qty, price); err != nil {
		return fmt.Errorf("upsert bought position: %w", err)
	}
	return nil
}

func reduceSoldPositionTx(ctx context.Context, tx pgx.Tx, accountID int64, symbol string, price, qty float64) (float64, error) {
	const selectQ = `
		SELECT qty, avg_entry_price
		FROM positions
		WHERE account_id = $1 AND symbol = $2 AND position_side = 'long'
		FOR UPDATE`

	var heldQty, avgEntryPrice float64
	err := tx.QueryRow(ctx, selectQ, accountID, symbol).Scan(&heldQty, &avgEntryPrice)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("cannot sell missing position %s", symbol)
		}
		return 0, fmt.Errorf("query sold position: %w", err)
	}
	if heldQty+1e-8 < qty {
		return 0, fmt.Errorf("cannot sell %.8f %s with only %.8f held", qty, symbol, heldQty)
	}

	realizedPnL := (price - avgEntryPrice) * qty
	remainingQty := heldQty - qty
	if remainingQty <= 1e-8 {
		const deleteQ = `
			DELETE FROM positions
			WHERE account_id = $1 AND symbol = $2 AND position_side = 'long'`
		if _, err := tx.Exec(ctx, deleteQ, accountID, symbol); err != nil {
			return 0, fmt.Errorf("delete closed position: %w", err)
		}
		return realizedPnL, nil
	}

	const updateQ = `
		UPDATE positions
		SET qty = $3,
			current_price = $4,
			realized_pnl = realized_pnl + $5
		WHERE account_id = $1 AND symbol = $2 AND position_side = 'long'`
	if _, err := tx.Exec(ctx, updateQ, accountID, symbol, remainingQty, price, realizedPnL); err != nil {
		return 0, fmt.Errorf("reduce sold position: %w", err)
	}
	return realizedPnL, nil
}
