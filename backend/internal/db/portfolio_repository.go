package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

func NewPortfolioRepository(db *pgxpool.Pool) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

// RecordLongFill stores a filled long-side trade and updates the user's open position atomically.
func (r *PortfolioRepository) RecordLongFill(ctx context.Context, userID int64, action, symbol string, price, qty, slippage float64) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin portfolio fill transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const tradeQ = `
		INSERT INTO trades (user_id, action, symbol, price, qty, slippage, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'FILLED')`
	if _, err := tx.Exec(ctx, tradeQ, userID, action, symbol, price, qty, slippage); err != nil {
		return fmt.Errorf("insert trade fill: %w", err)
	}

	switch action {
	case "BUY_LONG":
		const upsertPositionQ = `
			INSERT INTO positions (user_id, symbol, qty, avg_entry_price, current_price)
			VALUES ($1, $2, $3, $4, $4)
			ON CONFLICT (user_id, symbol) DO UPDATE SET
				qty = positions.qty + EXCLUDED.qty,
				avg_entry_price = CASE
					WHEN positions.qty + EXCLUDED.qty <= 0 THEN EXCLUDED.avg_entry_price
					ELSE ((positions.qty * positions.avg_entry_price) + (EXCLUDED.qty * EXCLUDED.avg_entry_price)) / (positions.qty + EXCLUDED.qty)
				END,
				current_price = EXCLUDED.current_price`
		if _, err := tx.Exec(ctx, upsertPositionQ, userID, symbol, qty, price); err != nil {
			return fmt.Errorf("upsert bought position: %w", err)
		}
	case "SELL_LONG":
		const reducePositionQ = `
			UPDATE positions
			SET qty = qty - $4, current_price = $3
			WHERE user_id = $1 AND symbol = $2`
		tag, err := tx.Exec(ctx, reducePositionQ, userID, symbol, price, qty)
		if err != nil {
			return fmt.Errorf("reduce sold position: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("cannot sell missing position %s", symbol)
		}

		const deleteFlatPositionQ = `
			DELETE FROM positions
			WHERE user_id = $1 AND symbol = $2 AND qty <= 0`
		if _, err := tx.Exec(ctx, deleteFlatPositionQ, userID, symbol); err != nil {
			return fmt.Errorf("delete flat position: %w", err)
		}
	default:
		return fmt.Errorf("unsupported trade action %s", action)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit portfolio fill transaction: %w", err)
	}
	return nil
}

// MarkPositionPrice updates the current price used for unrealized P&L on an open position.
func (r *PortfolioRepository) MarkPositionPrice(ctx context.Context, userID int64, symbol string, currentPrice float64) error {
	const q = `
		UPDATE positions
		SET current_price = $3
		WHERE user_id = $1 AND symbol = $2`
	if _, err := r.db.Exec(ctx, q, userID, symbol, currentPrice); err != nil {
		return fmt.Errorf("mark position price: %w", err)
	}
	return nil
}

// SavePortfolioSnapshot records account equity with a best-effort 24h change baseline.
func (r *PortfolioRepository) SavePortfolioSnapshot(ctx context.Context, userID int64, totalAssetValue, targetValue float64) error {
	previousValue, err := r.snapshotComparisonValue(ctx, userID)
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

	targetProgress := 0.0
	if targetValue > 0 {
		targetProgress = (totalAssetValue / targetValue) * 100
	}
	targetProgress = clampNumericPct(targetProgress)

	const q = `
		INSERT INTO portfolio_snapshots (
			user_id,
			total_asset_value,
			change_percent_24h,
			change_dollar_24h,
			estimated_annual_yield,
			target_progress_percent
		)
		VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := r.db.Exec(ctx, q, userID, totalAssetValue, changePercent, changeDollar, 0, targetProgress); err != nil {
		return fmt.Errorf("insert portfolio snapshot: %w", err)
	}
	return nil
}

func (r *PortfolioRepository) snapshotComparisonValue(ctx context.Context, userID int64) (float64, error) {
	const dayOldQ = `
		SELECT total_asset_value
		FROM portfolio_snapshots
		WHERE user_id = $1 AND timestamp <= NOW() - INTERVAL '24 hours'
		ORDER BY timestamp DESC
		LIMIT 1`

	var value float64
	err := r.db.QueryRow(ctx, dayOldQ, userID).Scan(&value)
	if err == nil {
		return value, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("query 24h portfolio baseline: %w", err)
	}

	const latestQ = `
		SELECT total_asset_value
		FROM portfolio_snapshots
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT 1`
	err = r.db.QueryRow(ctx, latestQ, userID).Scan(&value)
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

// GetExecutionStream returns the latest persisted trades.
func (r *PortfolioRepository) GetExecutionStream(ctx context.Context, userID int64, limit int) ([]TradeLog, error) {
	const q = `
		SELECT id, timestamp, action, symbol, price, qty, slippage, status
		FROM trades
		WHERE user_id = $1
		ORDER BY timestamp DESC
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
		WHERE user_id = $1`

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

	// Return [] instead of null for empty history responses.
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
