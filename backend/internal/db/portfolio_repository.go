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

// GetExecutionStream fetches paged trade histories for your Activity Console
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

// GetActivePositions calculates floating values and current trade exposure on the fly
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
		// Compute real-time derived quantitative variables safely
		p.Exposure = p.Qty * p.CurrentPrice
		p.UnrealizedPnL = (p.CurrentPrice - p.AvgEntryPrice) * p.Qty
		activePositions = append(activePositions, p)
	}
	return activePositions, rows.Err()
}

// GetLatestSummary fetches the current capital posture details
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

// GetSystemAlerts retrieves recent risk warnings for the banner dock
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
// 📂 internal/db/portfolio_repository.go


// GetPortfolioHistory retrieves the latest N snapshots sorted chronologically for charting
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

	// Initialize as an empty slice instead of nil so it encodes to a clean JSON literal "[]" instead of "null"
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