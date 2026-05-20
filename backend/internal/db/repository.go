package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/oalpha/pkg/models"
)

// Repository provides data access for OHLCV bars.
type Repository struct {
	db *sql.DB
}

// NewRepository returns a Repository backed by db.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// InsertBars upserts bars; conflicts on (time, symbol) update OHLCV.
func (r *Repository) InsertBars(ctx context.Context, bars []models.Bar) (int64, error) {
	if len(bars) == 0 {
		return 0, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const q = `
		INSERT INTO bars (time, symbol, open, high, low, close, volume)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (time, symbol) DO UPDATE SET
			open = EXCLUDED.open,
			high = EXCLUDED.high,
			low = EXCLUDED.low,
			close = EXCLUDED.close,
			volume = EXCLUDED.volume`

	stmt, err := tx.PrepareContext(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	var inserted int64
	for _, b := range bars {
		res, err := stmt.ExecContext(ctx, b.Time, b.Symbol, b.Open, b.High, b.Low, b.Close, b.Volume)
		if err != nil {
			return inserted, fmt.Errorf("insert bar %s %s: %w", b.Symbol, b.Time, err)
		}
		n, _ := res.RowsAffected()
		inserted += n
	}

	if err := tx.Commit(); err != nil {
		return inserted, fmt.Errorf("commit: %w", err)
	}
	return inserted, nil
}

// GetBars returns bars for symbol ordered by time ascending.
func (r *Repository) GetBars(ctx context.Context, symbol string, start, end time.Time) ([]models.Bar, error) {
	const q = `
		SELECT time, symbol, open, high, low, close, volume
		FROM bars
		WHERE symbol = $1 AND time >= $2 AND time <= $3
		ORDER BY time ASC`

	rows, err := r.db.QueryContext(ctx, q, symbol, start, end)
	if err != nil {
		return nil, fmt.Errorf("query bars: %w", err)
	}
	defer rows.Close()

	var bars []models.Bar
	for rows.Next() {
		var b models.Bar
		if err := rows.Scan(&b.Time, &b.Symbol, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume); err != nil {
			return nil, fmt.Errorf("scan bar: %w", err)
		}
		bars = append(bars, b)
	}
	return bars, rows.Err()
}

// CountBars returns the number of bars for symbol in [start, end].
func (r *Repository) CountBars(ctx context.Context, symbol string, start, end time.Time) (int64, error) {
	const q = `SELECT COUNT(*) FROM bars WHERE symbol = $1 AND time >= $2 AND time <= $3`
	var n int64
	err := r.db.QueryRowContext(ctx, q, symbol, start, end).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count bars: %w", err)
	}
	return n, nil
}

// DataValidationReport summarizes data quality for a symbol.
type DataValidationReport struct {
	Symbol      string
	BarCount    int64
	FirstBar    time.Time
	LastBar     time.Time
	InvalidBars int
	Gaps        int
}

// ValidateData checks bar integrity and rough continuity.
func (r *Repository) ValidateData(ctx context.Context, symbol string, start, end time.Time, expectedInterval time.Duration) (*DataValidationReport, error) {
	bars, err := r.GetBars(ctx, symbol, start, end)
	if err != nil {
		return nil, err
	}

	report := &DataValidationReport{Symbol: symbol, BarCount: int64(len(bars))}
	if len(bars) == 0 {
		return report, nil
	}

	report.FirstBar = bars[0].Time
	report.LastBar = bars[len(bars)-1].Time

	for _, b := range bars {
		if b.High < b.Low || b.Open < 0 || b.Close < 0 || b.Volume < 0 {
			report.InvalidBars++
		}
		if b.High < b.Open || b.High < b.Close || b.Low > b.Open || b.Low > b.Close {
			report.InvalidBars++
		}
	}

	if expectedInterval > 0 {
		threshold := expectedInterval + expectedInterval/2
		for i := 1; i < len(bars); i++ {
			gap := bars[i].Time.Sub(bars[i-1].Time)
			if gap > threshold {
				report.Gaps++
			}
		}
	}

	return report, nil
}
