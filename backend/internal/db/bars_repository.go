package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oalpha/pkg/models"
)

// BarsRepository provides data access for OHLCV bars.
type BarsRepository struct {
	db *pgxpool.Pool
}

// GetDB returns the underlying database connection pool.
func (r *BarsRepository) GetDB() *pgxpool.Pool {
	return r.db
}

// NewBarsRepository returns a BarsRepository backed by db.
func NewBarsRepository(db *pgxpool.Pool) *BarsRepository {
	return &BarsRepository{db: db}
}

// InsertBars upserts bars; conflicts on (time, symbol, timeframe) update OHLCV.
func (r *BarsRepository) InsertBars(ctx context.Context, bars []models.Bar, timeframe string) (int64, error) {
	if len(bars) == 0 {
		return 0, nil
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO bars (time, symbol, timeframe, open, high, low, close, volume)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (time, symbol, timeframe) DO UPDATE SET
			open = EXCLUDED.open,
			high = EXCLUDED.high,
			low = EXCLUDED.low,
			close = EXCLUDED.close,
			volume = EXCLUDED.volume`

	const assetQ = `
		INSERT INTO assets (symbol, name, asset_class)
		VALUES ($1, $1, 'equity')
		ON CONFLICT (symbol) DO NOTHING`

	var inserted int64
	seenAssets := make(map[string]struct{})
	batch := &pgx.Batch{}
	for _, b := range bars {
		symbol := normalizeSymbol(b.Symbol)
		if symbol == "" {
			return inserted, fmt.Errorf("bar symbol is required")
		}
		if _, seen := seenAssets[symbol]; !seen {
			if _, err := tx.Exec(ctx, assetQ, symbol); err != nil {
				return inserted, fmt.Errorf("upsert asset %s: %w", symbol, err)
			}
			seenAssets[symbol] = struct{}{}
		}
		batch.Queue(q, b.Time, symbol, timeframe, b.Open, b.High, b.Low, b.Close, b.Volume)
	}

	br := tx.SendBatch(ctx, batch)
	for range bars {
		_, err := br.Exec()
		if err != nil {
			return inserted, fmt.Errorf("batch execution failed: %w", err)
		}
		inserted++
	}
	if err := br.Close(); err != nil {
		return inserted, fmt.Errorf("close batch: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return inserted, fmt.Errorf("commit: %w", err)
	}
	return inserted, nil
}

// GetBars returns bars for symbol and timeframe ordered by time ascending.
func (r *BarsRepository) GetBars(ctx context.Context, symbol, timeframe string, start, end time.Time) ([]models.Bar, error) {
	symbol = normalizeSymbol(symbol)
	const q = `
		SELECT time, symbol, open, high, low, close, volume
		FROM bars
		WHERE symbol = $1 AND timeframe = $2 AND time >= $3 AND time <= $4
		ORDER BY time ASC`

	rows, err := r.db.Query(ctx, q, symbol, timeframe, start, end)
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
func (r *BarsRepository) CountBars(ctx context.Context, symbol string, start, end time.Time) (int64, error) {
	symbol = normalizeSymbol(symbol)
	const q = `SELECT COUNT(*) FROM bars WHERE symbol = $1 AND time >= $2 AND time <= $3`
	var n int64
	err := r.db.QueryRow(ctx, q, symbol, start, end).Scan(&n)
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
func (r *BarsRepository) ValidateData(ctx context.Context, symbol, timeframe string, start, end time.Time, expectedInterval time.Duration) (*DataValidationReport, error) {
	bars, err := r.GetBars(ctx, symbol, timeframe, start, end)
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

// SaveStrategyConfig persists a strategy configuration and returns its generated id.
func (r *BarsRepository) SaveStrategyConfig(ctx context.Context, userID int64, name, strategyType string, parameters map[string]interface{}) (int64, error) {
	paramsBytes, err := json.Marshal(parameters)
	if err != nil {
		return 0, fmt.Errorf("marshal strategy parameters: %w", err)
	}

	var id int64
	const q = `
        INSERT INTO strategy_configs (user_id, name, strategy_type, parameters)
        VALUES ($1, $2, $3, $4)
        RETURNING id`

	err = r.db.QueryRow(ctx, q, userID, name, strategyType, paramsBytes).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert strategy config: %w", err)
	}
	return id, nil
}

// SaveBacktestRun persists a backtest execution with its metrics.
func (r *BarsRepository) SaveBacktestRun(ctx context.Context, userID, configID int64, req models.BacktestRequest, result *models.BacktestResult) error {
	symbol := normalizeSymbol(req.Symbol)
	if symbol == "" {
		return fmt.Errorf("backtest symbol is required")
	}
	strategyType := strings.ToUpper(strings.TrimSpace(req.StrategyType))
	if strategyType == "" {
		strategyType = "MA_CROSSOVER"
	}

	endDate := time.Now().UTC()
	startDate := endDate.Add(-365 * 24 * time.Hour)
	if req.Start != nil {
		startDate = req.Start.UTC()
	}
	if req.End != nil {
		endDate = req.End.UTC()
	}

	equityCurveBytes, err := json.Marshal(result.EquityCurve)
	if err != nil {
		return fmt.Errorf("marshal equity curve: %w", err)
	}

	var configIDArg *int64
	if configID > 0 {
		configIDArg = &configID
	}

	params := map[string]interface{}{
		"strategy_type": strategyType,
		"q_noise":       req.QNoise,
		"r_noise":       req.RNoise,
		"z_threshold":   req.ZThreshold,
		"fast_period":   req.FastPeriod,
		"slow_period":   req.SlowPeriod,
	}
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("marshal backtest parameters: %w", err)
	}

	const assetQ = `
		INSERT INTO assets (symbol, name, asset_class)
		VALUES ($1, $1, 'equity')
		ON CONFLICT (symbol) DO NOTHING`
	if _, err := r.db.Exec(ctx, assetQ, symbol); err != nil {
		return fmt.Errorf("upsert backtest asset: %w", err)
	}

	const q = `
        INSERT INTO backtest_runs (
            user_id, strategy_config_id, strategy_type, symbol, timeframe,
            start_time, end_time, initial_cash,
            final_equity, total_return_pct, annual_return_pct, sharpe_ratio, sortino_ratio, max_drawdown_pct, num_trades,
            parameters, equity_curve, created_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 0, $11, $12, $13, $14, $15, $16, NOW())`

	_, err = r.db.Exec(ctx, q,
		userID, configIDArg, strategyType, symbol, req.Timeframe,
		startDate, endDate, req.InitialCash,
		result.FinalEquity, result.TotalReturn*100, result.Sharpe, result.Sortino, result.MaxDrawdown*100, result.NumTrades,
		paramsBytes, equityCurveBytes,
	)
	if err != nil {
		return fmt.Errorf("insert backtest run: %w", err)
	}
	return nil
}

// GetLatestBarTime returns the latest stored timestamp for a symbol/timeframe pair.
func (r *BarsRepository) GetLatestBarTime(ctx context.Context, symbol, timeframe string) (time.Time, bool, error) {
	symbol = normalizeSymbol(symbol)
	const q = `
		SELECT max(time) 
		FROM bars 
		WHERE symbol = $1 AND timeframe = $2`

	var latestTime *time.Time
	err := r.db.QueryRow(ctx, q, symbol, timeframe).Scan(&latestTime)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("query max bar time: %w", err)
	}

	if latestTime == nil {
		return time.Time{}, false, nil
	}

	return *latestTime, true, nil
}

func normalizeSymbol(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}
