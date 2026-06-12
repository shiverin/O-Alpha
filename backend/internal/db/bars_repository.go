package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oalpha/internal/backtest"
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

type BarQueryOptions struct {
	Feed         string
	Adjustment   string
	Source       string
	AlignMode    backtest.AlignMode
	MaxStaleBars int
}

// InsertBars upserts raw IEX Alpaca bars for backward-compatible callers.
func (r *BarsRepository) InsertBars(ctx context.Context, bars []models.Bar, timeframe string) (int64, error) {
	return r.InsertBarsDataset(ctx, bars, timeframe, "iex", "raw", "alpaca")
}

// InsertBarsDataset upserts bars for a specific feed/adjustment/source dataset.
func (r *BarsRepository) InsertBarsDataset(ctx context.Context, bars []models.Bar, timeframe string, feed string, adjustment string, source string) (int64, error) {
	if len(bars) == 0 {
		return 0, nil
	}
	feed, adjustment, source = normalizeDataset(feed, adjustment, source)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO bars (time, symbol, timeframe, feed, adjustment, source, open, high, low, close, volume)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (time, symbol, timeframe, feed, adjustment, source) DO UPDATE SET
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
		batch.Queue(q, b.Time, symbol, timeframe, feed, adjustment, source, b.Open, b.High, b.Low, b.Close, b.Volume)
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

// GetBars returns default raw IEX Alpaca bars for symbol and timeframe ordered by time ascending.
func (r *BarsRepository) GetBars(ctx context.Context, symbol, timeframe string, start, end time.Time) ([]models.Bar, error) {
	return r.GetBarsDataset(ctx, symbol, timeframe, start, end, BarQueryOptions{})
}

// GetBarsDataset returns one symbol from one immutable market-data dataset.
func (r *BarsRepository) GetBarsDataset(ctx context.Context, symbol, timeframe string, start, end time.Time, opts BarQueryOptions) ([]models.Bar, error) {
	symbol = normalizeSymbol(symbol)
	feed, adjustment, source := normalizeDataset(opts.Feed, opts.Adjustment, opts.Source)
	const q = `
		SELECT time, symbol, open, high, low, close, volume
		FROM bars
		WHERE symbol = $1 AND timeframe = $2 AND time >= $3 AND time <= $4
			AND feed = $5 AND adjustment = $6 AND source = $7
		ORDER BY time ASC`

	rows, err := r.db.Query(ctx, q, symbol, timeframe, start, end, feed, adjustment, source)
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

func (r *BarsRepository) GetBarsMulti(
	ctx context.Context,
	symbols []string,
	timeframe string,
	start time.Time,
	end time.Time,
	opts BarQueryOptions,
) (backtest.AlignedBars, error) {
	normalizedSymbols := normalizeSymbols(symbols)
	if len(normalizedSymbols) == 0 {
		return backtest.AlignedBars{}, fmt.Errorf("at least one symbol is required")
	}
	feed, adjustment, source := normalizeDataset(opts.Feed, opts.Adjustment, opts.Source)
	const q = `
		SELECT time, symbol, open, high, low, close, volume
		FROM bars
		WHERE symbol = ANY($1) AND timeframe = $2 AND time >= $3 AND time <= $4
			AND feed = $5 AND adjustment = $6 AND source = $7
		ORDER BY time ASC, symbol ASC`

	rows, err := r.db.Query(ctx, q, normalizedSymbols, timeframe, start, end, feed, adjustment, source)
	if err != nil {
		return backtest.AlignedBars{}, fmt.Errorf("query multi-symbol bars: %w", err)
	}
	defer rows.Close()

	grouped := make(map[string][]models.Bar, len(normalizedSymbols))
	for _, symbol := range normalizedSymbols {
		grouped[symbol] = nil
	}
	for rows.Next() {
		var b models.Bar
		if err := rows.Scan(&b.Time, &b.Symbol, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume); err != nil {
			return backtest.AlignedBars{}, fmt.Errorf("scan multi-symbol bar: %w", err)
		}
		grouped[b.Symbol] = append(grouped[b.Symbol], b)
	}
	if err := rows.Err(); err != nil {
		return backtest.AlignedBars{}, err
	}

	alignMode := opts.AlignMode
	if alignMode == "" {
		alignMode = backtest.AlignInnerJoin
	}
	return backtest.AlignBars(grouped, backtest.AlignmentConfig{
		Mode:         alignMode,
		MaxStaleBars: opts.MaxStaleBars,
		Timeframe:    timeframe,
		Feed:         feed,
		Adjustment:   adjustment,
	})
}

// CountBars returns the number of bars for symbol in [start, end].
func (r *BarsRepository) CountBars(ctx context.Context, symbol string, start, end time.Time) (int64, error) {
	symbol = normalizeSymbol(symbol)
	const q = `SELECT COUNT(*) FROM bars WHERE symbol = $1 AND time >= $2 AND time <= $3 AND feed = 'iex' AND adjustment = 'raw' AND source = 'alpaca'`
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
	timeframe := strings.TrimSpace(req.Timeframe)
	if timeframe == "" {
		timeframe = "1Day"
	}
	initialCash := req.InitialCash
	if initialCash <= 0 {
		initialCash = defaultPaperInitialCash
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
		"feed":          req.Feed,
		"adjustment":    req.Adjustment,
	}
	for key, value := range req.Parameters {
		params[key] = value
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
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW())`

	_, err = r.db.Exec(ctx, q,
		userID, configIDArg, strategyType, symbol, timeframe,
		startDate, endDate, initialCash,
		result.FinalEquity, result.TotalReturn*100, result.AnnualizedReturn*100, result.Sharpe, result.Sortino, result.MaxDrawdown*100, result.NumTrades,
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
		WHERE symbol = $1 AND timeframe = $2 AND feed = 'iex' AND adjustment = 'raw' AND source = 'alpaca'`

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

// GetLatestBarTimes returns the latest stored raw Alpaca timestamp per symbol.
func (r *BarsRepository) GetLatestBarTimes(ctx context.Context, symbols []string, timeframe string) (map[string]time.Time, error) {
	normalizedSymbols := normalizeSymbols(symbols)
	out := make(map[string]time.Time, len(normalizedSymbols))
	if len(normalizedSymbols) == 0 {
		return out, nil
	}

	const q = `
		SELECT symbol, max(time)
		FROM bars
		WHERE symbol = ANY($1) AND timeframe = $2
			AND feed = 'iex' AND adjustment = 'raw' AND source = 'alpaca'
		GROUP BY symbol`

	rows, err := r.db.Query(ctx, q, normalizedSymbols, timeframe)
	if err != nil {
		return nil, fmt.Errorf("query latest bar times: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var symbol string
		var latest time.Time
		if err := rows.Scan(&symbol, &latest); err != nil {
			return nil, fmt.Errorf("scan latest bar time: %w", err)
		}
		out[normalizeSymbol(symbol)] = latest
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func normalizeSymbol(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}

func normalizeSymbols(symbols []string) []string {
	out := make([]string, 0, len(symbols))
	seen := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		normalized := normalizeSymbol(symbol)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeDataset(feed string, adjustment string, source string) (string, string, string) {
	feed = strings.ToLower(strings.TrimSpace(feed))
	adjustment = strings.ToLower(strings.TrimSpace(adjustment))
	source = strings.ToLower(strings.TrimSpace(source))
	if feed == "" {
		feed = "iex"
	}
	if adjustment == "" {
		adjustment = "raw"
	}
	if source == "" {
		source = "alpaca"
	}
	return feed, adjustment, source
}
