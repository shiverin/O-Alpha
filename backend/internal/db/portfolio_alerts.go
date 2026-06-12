package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *PortfolioRepository) InsertSystemAlert(ctx context.Context, userID int64, alertType, title, description, source string, metadata map[string]interface{}) error {
	alertType = strings.ToUpper(strings.TrimSpace(alertType))
	switch alertType {
	case "INFO", "WARNING", "CRITICAL":
	default:
		alertType = "INFO"
	}
	if source == "" {
		source = "agent"
	}
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal alert metadata: %w", err)
	}

	const q = `
		INSERT INTO system_alerts (user_id, title, description, alert_type, source, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := r.db.Exec(ctx, q, userID, title, description, alertType, source, metaBytes); err != nil {
		return fmt.Errorf("insert system alert: %w", err)
	}
	return nil
}

func (r *PortfolioRepository) GetAccountState(ctx context.Context, userID int64) (float64, map[string]float64, error) {
	const cashQ = `
		SELECT cash_balance
		FROM accounts
		WHERE user_id = $1
			AND account_type = 'paper'
			AND provider = 'internal'
			AND provider_account_id = ''
		ORDER BY id
		LIMIT 1`

	var cash float64
	err := r.db.QueryRow(ctx, cashQ, userID).Scan(&cash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, map[string]float64{}, nil
		}
		return 0, nil, fmt.Errorf("query account cash: %w", err)
	}

	const posQ = `
		SELECT symbol, qty
		FROM positions
		WHERE user_id = $1 AND position_side = 'long' AND qty > 0`
	rows, err := r.db.Query(ctx, posQ, userID)
	if err != nil {
		return 0, nil, fmt.Errorf("query account positions: %w", err)
	}
	defer rows.Close()

	positions := make(map[string]float64)
	for rows.Next() {
		var symbol string
		var qty float64
		if err := rows.Scan(&symbol, &qty); err != nil {
			return 0, nil, err
		}
		positions[symbol] = qty
	}
	return cash, positions, rows.Err()
}

func (r *PortfolioRepository) RecordLongFillKeyed(ctx context.Context, userID, agentRunID int64, clientOrderID string, action, symbol string, price, qty, slippage float64) error {
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

	if clientOrderID != "" {
		var exists bool
		if err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE client_order_id = $1)`, clientOrderID).Scan(&exists); err != nil {
			return fmt.Errorf("dedupe check: %w", err)
		}
		if exists {
			return nil
		}
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
	var clientOrderIDArg *string
	if clientOrderID != "" {
		clientOrderIDArg = &clientOrderID
	}

	var orderID int64
	const orderQ = `
		INSERT INTO orders (
			user_id, account_id, agent_run_id, symbol, side, position_side,
			order_type, time_in_force, qty, status, client_order_id, submitted_at, filled_at
		)
		VALUES ($1, $2, $3, $4, $5, 'long', 'market', 'day', $6, 'filled', $7, NOW(), NOW())
		RETURNING id`
	if err := tx.QueryRow(ctx, orderQ, userID, account.ID, agentRunIDArg, symbol, side, qty, clientOrderIDArg).Scan(&orderID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil
		}
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
