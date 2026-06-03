package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/config"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/internal/ml"
	"github.com/oalpha/pkg/models"
)

type metricSummary struct {
	TotalReturn      float64 `json:"total_return"`
	AnnualizedReturn float64 `json:"annualized_return"`
	Sharpe           float64 `json:"sharpe"`
	Sortino          float64 `json:"sortino"`
	Calmar           float64 `json:"calmar"`
	MaxDrawdown      float64 `json:"max_drawdown"`
	NumTrades        int     `json:"num_trades"`
	WinRate          float64 `json:"win_rate"`
	ProfitFactor     float64 `json:"profit_factor"`
	ExposurePercent  float64 `json:"exposure_percent"`
	Turnover         float64 `json:"turnover"`
	FinalEquity      float64 `json:"final_equity"`
}

type comparisonReport struct {
	GeneratedAt time.Time                `json:"generated_at"`
	Symbol      string                   `json:"symbol"`
	Timeframe   string                   `json:"timeframe"`
	Start       time.Time                `json:"start"`
	End         time.Time                `json:"end"`
	BarCount    int                      `json:"bar_count"`
	Base        metricSummary            `json:"base"`
	MLMeta      metricSummary            `json:"ml_meta"`
	BuyHold     metricSummary            `json:"buy_hold"`
	Diagnostics map[string]interface{}   `json:"diagnostics,omitempty"`
	Artifacts   map[string]string        `json:"artifacts,omitempty"`
	Delta       map[string]metricSummary `json:"-"`
}

type symbolBars struct {
	Symbol string
	Bars   []models.Bar
}

type signalRow struct {
	Symbol string
	Time   time.Time
	Output backtest.StrategyOutput
}

func main() {
	var (
		mode           = flag.String("mode", "export", "mode: export or compare")
		symbol         = flag.String("symbol", "VOO", "primary symbol")
		symbols        = flag.String("symbols", "", "comma-separated training symbols for export")
		allowSingle    = flag.Bool("allow-single-symbol-export", false, "allow export to fall back to --symbol when --symbols is omitted")
		timeframe      = flag.String("timeframe", "1Day", "bar timeframe")
		from           = flag.String("from", "2015-01-01", "inclusive start date YYYY-MM-DD")
		to             = flag.String("to", "2026-06-01", "inclusive end date YYYY-MM-DD")
		feed           = flag.String("feed", "", "market data feed filter; default repo dataset")
		adjustment     = flag.String("adjustment", "", "adjustment filter; default repo dataset")
		source         = flag.String("source", "", "market data source filter; default repo dataset")
		contextSymbols = flag.String("context-symbols", "SPY,QQQ,IWM", "comma-separated context symbols")
		outputDir      = flag.String("output-dir", "", "output directory")
		fastPeriod     = flag.Int("fast", 20, "MA fast period")
		slowPeriod     = flag.Int("slow", 50, "MA slow period")
		initialCash    = flag.Float64("initial-cash", 100000, "initial cash")
		metadataPath   = flag.String("metadata", "", "metadata.json for ML comparison")
		modelPath      = flag.String("model", "", "model.txt for ML comparison when metadata is absent")
	)
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		fatal(err)
	}
	pool, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		fatal(err)
	}
	defer pool.Close()

	start, end, err := resolveDateRange(*from, *to)
	if err != nil {
		fatal(err)
	}
	repo := db.NewBarsRepository(pool)
	primary := strings.ToUpper(strings.TrimSpace(*symbol))
	queryOpts := db.BarQueryOptions{
		Feed:       *feed,
		Adjustment: *adjustment,
		Source:     *source,
	}
	dir := strings.TrimSpace(*outputDir)
	if dir == "" {
		dir = filepath.Join("..", "reports", "batches", time.Now().UTC().Format("2006-01-02")+"_ml_meta_label_"+strings.ToLower(primary))
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fatal(err)
	}

	switch strings.ToLower(strings.TrimSpace(*mode)) {
	case "export":
		trainingSymbols := parseCSV(*symbols)
		if len(trainingSymbols) == 0 {
			if !*allowSingle {
				fatal(fmt.Errorf("export requires --symbols for provenance; pass --allow-single-symbol-export to intentionally export only --symbol=%s", primary))
			}
			trainingSymbols = []string{primary}
		}
		if err := runExport(ctx, repo, trainingSymbols, *timeframe, start, end, queryOpts, parseCSV(*contextSymbols), *fastPeriod, *slowPeriod, dir); err != nil {
			fatal(err)
		}
	case "compare":
		bars, err := repo.GetBarsDataset(ctx, primary, *timeframe, start, end, queryOpts)
		if err != nil {
			fatal(err)
		}
		if len(bars) == 0 {
			fatal(fmt.Errorf("no bars for %s %s between %s and %s", primary, *timeframe, start.Format(time.RFC3339), end.Format(time.RFC3339)))
		}
		if err := runCompare(ctx, repo, primary, *timeframe, start, end, queryOpts, parseCSV(*contextSymbols), bars, *fastPeriod, *slowPeriod, *initialCash, *metadataPath, *modelPath, dir); err != nil {
			fatal(err)
		}
	case "inventory":
		if err := runInventory(ctx, pool, *timeframe, start, end, queryOpts); err != nil {
			fatal(err)
		}
	default:
		fatal(fmt.Errorf("unsupported mode %q", *mode))
	}
}

func runExport(
	ctx context.Context,
	repo *db.BarsRepository,
	symbols []string,
	timeframe string,
	start time.Time,
	end time.Time,
	queryOpts db.BarQueryOptions,
	contextSymbols []string,
	fast int,
	slow int,
	outputDir string,
) error {
	barsPath := filepath.Join(outputDir, "bars.csv")
	signalsPath := filepath.Join(outputDir, "signals.csv")

	trainingSeries := make([]symbolBars, 0, len(symbols))
	allBarsBySymbol := make(map[string][]models.Bar)
	for _, symbol := range symbols {
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" {
			continue
		}
		bars, err := repo.GetBarsDataset(ctx, symbol, timeframe, start, end, queryOpts)
		if err != nil {
			return fmt.Errorf("load training bars for %s: %w", symbol, err)
		}
		if len(bars) == 0 {
			return fmt.Errorf("no bars for training symbol %s %s between %s and %s", symbol, timeframe, start.Format(time.RFC3339), end.Format(time.RFC3339))
		}
		trainingSeries = append(trainingSeries, symbolBars{Symbol: symbol, Bars: bars})
		allBarsBySymbol[symbol] = bars
	}
	if len(trainingSeries) == 0 {
		return fmt.Errorf("at least one training symbol is required")
	}

	for _, contextSymbol := range contextSymbols {
		contextSymbol = strings.ToUpper(strings.TrimSpace(contextSymbol))
		if contextSymbol == "" {
			continue
		}
		if _, exists := allBarsBySymbol[contextSymbol]; exists {
			continue
		}
		contextBars, err := repo.GetBarsDataset(ctx, contextSymbol, timeframe, start, end, queryOpts)
		if err == nil && len(contextBars) > 0 {
			allBarsBySymbol[contextSymbol] = contextBars
		}
	}

	allBars := flattenBars(allBarsBySymbol)
	if err := writeBarsCSV(barsPath, allBars); err != nil {
		return err
	}

	strategy := backtest.NewMACrossoverStrategy(fast, slow)
	var signalRows []signalRow
	nonHold := 0
	buys := 0
	sells := 0
	for _, series := range trainingSeries {
		outputs, err := strategy.GenerateSignals(ctx, series.Bars)
		if err != nil {
			return fmt.Errorf("generate signals for %s: %w", series.Symbol, err)
		}
		for i, output := range outputs {
			switch output.Signal {
			case models.SignalBuy:
				nonHold++
				buys++
			case models.SignalSell:
				nonHold++
				sells++
			}
			signalRows = append(signalRows, signalRow{
				Symbol: series.Symbol,
				Time:   series.Bars[i].Time,
				Output: output,
			})
		}
	}
	if err := writeSignalRowsCSV(signalsPath, signalRows); err != nil {
		return err
	}
	if err := writeExportManifest(filepath.Join(outputDir, "export_manifest.json"), map[string]interface{}{
		"generated_at":       time.Now().UTC(),
		"git_sha":            gitSHA(),
		"export_command":     strings.Join(os.Args, " "),
		"symbols":            symbols,
		"context_symbols":    contextSymbols,
		"dataset":            normalizedDataset(queryOpts),
		"timeframe":          timeframe,
		"start":              start,
		"end":                end,
		"fast_period":        fast,
		"slow_period":        slow,
		"bars_csv":           barsPath,
		"signals_csv":        signalsPath,
		"training_symbols":   len(trainingSeries),
		"total_bars":         len(allBars),
		"signal_rows":        len(signalRows),
		"non_hold_signals":   nonHold,
		"buy_signals":        buys,
		"sell_signals":       sells,
		"single_symbol_only": len(trainingSeries) == 1,
	}); err != nil {
		return err
	}
	fmt.Printf("exported %d training symbols, %d total bars, %d signal rows, %d non-hold signals (BUY=%d SELL=%d)\n  %s\n  %s\n",
		len(trainingSeries), len(allBars), len(signalRows), nonHold, buys, sells, barsPath, signalsPath)
	return nil
}

func runCompare(
	ctx context.Context,
	repo *db.BarsRepository,
	symbol string,
	timeframe string,
	start time.Time,
	end time.Time,
	queryOpts db.BarQueryOptions,
	contextSymbols []string,
	bars []models.Bar,
	fast int,
	slow int,
	initialCash float64,
	metadataPath string,
	modelPath string,
	outputDir string,
) error {
	base := backtest.NewMACrossoverStrategy(fast, slow)
	baseResult, err := backtest.RunBacktest(ctx, bars, base, initialCash)
	if err != nil {
		return err
	}
	buyHold, err := backtest.RunBuyAndHold(bars, initialCash)
	if err != nil {
		return err
	}

	predictor, featureSpec, artifactVersion, thresholds, calibration, err := loadPredictor(metadataPath, modelPath)
	if err != nil {
		return err
	}
	contextBars := make(map[string][]models.Bar)
	for _, contextSymbol := range contextSymbols {
		contextSymbol = strings.ToUpper(strings.TrimSpace(contextSymbol))
		if contextSymbol == "" || contextSymbol == symbol {
			continue
		}
		contextSeries, err := repo.GetBarsDataset(ctx, contextSymbol, timeframe, start, end, queryOpts)
		if err == nil && len(contextSeries) > 0 {
			contextBars[contextSymbol] = contextSeries
		}
	}
	mlStrategy := &ml.MLMetaLabelStrategy{
		Symbol:         symbol,
		BaseStrategy:   backtest.NewMACrossoverStrategy(fast, slow),
		FeatureBuilder: ml.NewFeatureBuilder(featureSpec),
		Predictor:      predictor,
		Calibration:    calibration,
		Thresholds:     thresholds,
		ContextBars:    contextBars,
	}
	mlResult, err := backtest.RunBacktest(ctx, bars, mlStrategy, initialCash)
	if err != nil {
		return err
	}
	baseOutputs, err := base.GenerateSignals(ctx, bars)
	if err != nil {
		return err
	}
	mlOutputs, err := mlStrategy.GenerateSignals(ctx, bars)
	if err != nil {
		return err
	}
	decisionDebugPath := filepath.Join(outputDir, "ml_decision_debug.csv")
	if err := writeDecisionDebugCSV(decisionDebugPath, bars, baseOutputs, mlOutputs); err != nil {
		return err
	}

	report := comparisonReport{
		GeneratedAt: time.Now().UTC(),
		Symbol:      symbol,
		Timeframe:   timeframe,
		Start:       bars[0].Time,
		End:         bars[len(bars)-1].Time,
		BarCount:    len(bars),
		Base:        summarize(baseResult),
		MLMeta:      summarize(mlResult),
		BuyHold:     summarize(buyHold),
		Diagnostics: map[string]interface{}{
			"base_strategy":  "MA_CROSSOVER",
			"fast_period":    fast,
			"slow_period":    slow,
			"dataset":        normalizedDataset(queryOpts),
			"ml_threshold":   thresholds.EnterLong,
			"artifact":       artifactVersion,
			"context_loaded": keys(contextBars),
		},
		Artifacts: map[string]string{
			"metadata":       metadataPath,
			"model":          modelPath,
			"decision_debug": decisionDebugPath,
		},
	}
	jsonPath := filepath.Join(outputDir, "ml_meta_comparison.json")
	if err := writeJSON(jsonPath, report); err != nil {
		return err
	}
	mdPath := filepath.Join(outputDir, "ml_meta_comparison.md")
	if err := writeComparisonMarkdown(mdPath, report); err != nil {
		return err
	}
	fmt.Printf("comparison written:\n  %s\n  %s\n", jsonPath, mdPath)
	printOneLine("base", report.Base)
	printOneLine("ml_meta", report.MLMeta)
	printOneLine("buy_hold", report.BuyHold)
	return nil
}

func runInventory(ctx context.Context, pool *pgxpool.Pool, timeframe string, start time.Time, end time.Time, queryOpts db.BarQueryOptions) error {
	timeframe = strings.TrimSpace(timeframe)
	includeAll := timeframe == "" || strings.EqualFold(timeframe, "all")
	dataset := normalizedDataset(queryOpts)
	query := `
		SELECT symbol, timeframe, COUNT(*) AS n, MIN(time) AS first_bar, MAX(time) AS last_bar
		FROM bars
		WHERE time >= $1 AND time <= $2
			AND feed = $3 AND adjustment = $4 AND source = $5`
	args := []interface{}{start, end, dataset["feed"], dataset["adjustment"], dataset["source"]}
	if !includeAll {
		query += ` AND timeframe = $6`
		args = append(args, timeframe)
	}
	query += `
		GROUP BY symbol, timeframe
		ORDER BY n DESC, symbol ASC, timeframe ASC`

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Printf("symbol,timeframe,bars,first,last\n")
	for rows.Next() {
		var symbol string
		var tf string
		var count int64
		var first time.Time
		var last time.Time
		if err := rows.Scan(&symbol, &tf, &count, &first, &last); err != nil {
			return err
		}
		fmt.Printf("%s,%s,%d,%s,%s\n", symbol, tf, count, first.UTC().Format(time.RFC3339), last.UTC().Format(time.RFC3339))
	}
	return rows.Err()
}

func loadPredictor(metadataPath, modelPath string) (ml.Predictor, ml.FeatureSpec, string, ml.MLThresholds, ml.CalibrationModel, error) {
	if strings.TrimSpace(metadataPath) != "" {
		artifact, err := ml.ReadModelArtifact(metadataPath)
		if err != nil {
			return nil, ml.FeatureSpec{}, "", ml.MLThresholds{}, ml.CalibrationModel{}, err
		}
		predictor, err := ml.NewLeavesPredictor(artifact.ModelPath(filepath.Dir(metadataPath)), artifact.FeatureSpec, artifact.Version())
		if err != nil {
			return nil, ml.FeatureSpec{}, "", ml.MLThresholds{}, ml.CalibrationModel{}, err
		}
		return predictor, artifact.FeatureSpec, artifact.Version(), artifact.Thresholds, artifact.Calibration, nil
	}
	if strings.TrimSpace(modelPath) == "" {
		return nil, ml.FeatureSpec{}, "", ml.MLThresholds{}, ml.CalibrationModel{}, fmt.Errorf("compare requires --metadata or --model")
	}
	featureSpec := ml.DefaultFeatureSpec()
	predictor, err := ml.NewLeavesPredictor(modelPath, featureSpec, modelPath)
	if err != nil {
		return nil, ml.FeatureSpec{}, "", ml.MLThresholds{}, ml.CalibrationModel{}, err
	}
	return predictor, featureSpec, modelPath, ml.DefaultMLThresholds(), ml.CalibrationModel{}, nil
}

func writeBarsCSV(path string, bars []models.Bar) error {
	sort.SliceStable(bars, func(i, j int) bool {
		if bars[i].Symbol != bars[j].Symbol {
			return bars[i].Symbol < bars[j].Symbol
		}
		return bars[i].Time.Before(bars[j].Time)
	})
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"symbol", "time", "open", "high", "low", "close", "volume"}); err != nil {
		return err
	}
	for _, bar := range bars {
		if err := writer.Write([]string{
			bar.Symbol,
			bar.Time.UTC().Format(time.RFC3339),
			floatString(bar.Open),
			floatString(bar.High),
			floatString(bar.Low),
			floatString(bar.Close),
			fmt.Sprintf("%d", bar.Volume),
		}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func writeSignalRowsCSV(path string, rows []signalRow) error {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Symbol != rows[j].Symbol {
			return rows[i].Symbol < rows[j].Symbol
		}
		return rows[i].Time.Before(rows[j].Time)
	})
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"symbol", "time", "signal", "alpha_score", "confidence", "target_weight", "ensemble_score", "ensemble_confidence"}); err != nil {
		return err
	}
	for _, row := range rows {
		output := row.Output
		if err := writer.Write([]string{
			row.Symbol,
			row.Time.UTC().Format(time.RFC3339),
			signalName(output.Signal),
			floatString(output.AlphaScore),
			floatString(output.Confidence),
			floatString(output.TargetWeight),
			floatString(output.AlphaScore),
			floatString(output.Confidence),
		}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func writeDecisionDebugCSV(path string, bars []models.Bar, baseOutputs []backtest.StrategyOutput, mlOutputs []backtest.StrategyOutput) error {
	if len(baseOutputs) != len(bars) || len(mlOutputs) != len(bars) {
		return fmt.Errorf("decision debug length mismatch bars=%d base=%d ml=%d", len(bars), len(baseOutputs), len(mlOutputs))
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	writer := csv.NewWriter(file)
	if err := writer.Write([]string{
		"symbol",
		"time",
		"base_signal",
		"ml_signal",
		"p_success",
		"ml_decision",
		"ml_action",
		"base_target_weight",
		"ml_target_weight",
	}); err != nil {
		return err
	}
	for i, bar := range bars {
		base := baseOutputs[i]
		mlOut := mlOutputs[i]
		if base.Signal == models.SignalHold && mlOut.Signal == models.SignalHold && metadataString(mlOut.Metadata, "ml_decision") == "" {
			continue
		}
		if err := writer.Write([]string{
			bar.Symbol,
			bar.Time.UTC().Format(time.RFC3339),
			signalName(base.Signal),
			signalName(mlOut.Signal),
			metadataFloatString(mlOut.Metadata, "p_success"),
			metadataString(mlOut.Metadata, "ml_decision"),
			metadataString(mlOut.Metadata, "ml_action"),
			floatString(base.TargetWeight),
			floatString(mlOut.TargetWeight),
		}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func writeJSON(path string, value interface{}) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}

func writeExportManifest(path string, value map[string]interface{}) error {
	return writeJSON(path, value)
}

func writeComparisonMarkdown(path string, report comparisonReport) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# ML Meta-Label Comparison\n\n")
	fmt.Fprintf(&b, "- Symbol: `%s`\n", report.Symbol)
	fmt.Fprintf(&b, "- Timeframe: `%s`\n", report.Timeframe)
	fmt.Fprintf(&b, "- Period: `%s` to `%s`\n", report.Start.Format("2006-01-02"), report.End.Format("2006-01-02"))
	fmt.Fprintf(&b, "- Bars: `%d`\n\n", report.BarCount)
	fmt.Fprintf(&b, "| Strategy | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Exposure | Turnover |\n")
	fmt.Fprintf(&b, "|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	writeMetricRow(&b, "base_ma", report.Base)
	writeMetricRow(&b, "ml_meta", report.MLMeta)
	writeMetricRow(&b, "buy_hold", report.BuyHold)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeMetricRow(b *strings.Builder, name string, m metricSummary) {
	fmt.Fprintf(b, "| %s | %.2f%% | %.2f%% | %.3f | %.3f | %.3f | %.2f%% | %d | %.2f%% | %.3f |\n",
		name,
		m.TotalReturn*100,
		m.AnnualizedReturn*100,
		m.Sharpe,
		m.Sortino,
		m.Calmar,
		m.MaxDrawdown*100,
		m.NumTrades,
		m.ExposurePercent*100,
		m.Turnover,
	)
}

func summarize(result *models.BacktestResult) metricSummary {
	if result == nil {
		return metricSummary{}
	}
	return metricSummary{
		TotalReturn:      result.TotalReturn,
		AnnualizedReturn: result.AnnualizedReturn,
		Sharpe:           result.Sharpe,
		Sortino:          result.Sortino,
		Calmar:           result.Calmar,
		MaxDrawdown:      result.MaxDrawdown,
		NumTrades:        result.NumTrades,
		WinRate:          result.WinRate,
		ProfitFactor:     result.ProfitFactor,
		ExposurePercent:  result.ExposurePercent,
		Turnover:         result.Turnover,
		FinalEquity:      result.FinalEquity,
	}
}

func resolveDateRange(from, to string) (time.Time, time.Time, error) {
	start, err := parseDate(from, false)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseDate(to, true)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("to must be after from")
	}
	return start, end, nil
}

func parseDate(value string, endOfDay bool) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, err
	}
	parsed = parsed.UTC()
	if endOfDay {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return parsed, nil
}

func parseCSV(value string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0)
	for _, item := range strings.Split(value, ",") {
		normalized := strings.ToUpper(strings.TrimSpace(item))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func normalizedDataset(opts db.BarQueryOptions) map[string]string {
	feed := strings.ToLower(strings.TrimSpace(opts.Feed))
	adjustment := strings.ToLower(strings.TrimSpace(opts.Adjustment))
	source := strings.ToLower(strings.TrimSpace(opts.Source))
	if feed == "" {
		feed = "iex"
	}
	if adjustment == "" {
		adjustment = "raw"
	}
	if source == "" {
		source = "alpaca"
	}
	return map[string]string{
		"feed":       feed,
		"adjustment": adjustment,
		"source":     source,
	}
}

func flattenBars(grouped map[string][]models.Bar) []models.Bar {
	symbols := make([]string, 0, len(grouped))
	for symbol := range grouped {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	var out []models.Bar
	for _, symbol := range symbols {
		out = append(out, grouped[symbol]...)
	}
	return out
}

func signalName(signal models.Signal) string {
	switch signal {
	case models.SignalBuy:
		return "BUY"
	case models.SignalSell:
		return "SELL"
	default:
		return "HOLD"
	}
}

func floatString(value float64) string {
	return fmt.Sprintf("%.12g", value)
}

func metadataString(metadata map[string]interface{}, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}

func metadataFloatString(metadata map[string]interface{}, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case float64:
		return floatString(v)
	case float32:
		return floatString(float64(v))
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprint(v)
	}
}

func keys(values map[string][]models.Bar) []string {
	out := make([]string, 0, len(values))
	for key := range values {
		out = append(out, key)
	}
	return out
}

func printOneLine(name string, m metricSummary) {
	fmt.Printf("%-8s return=%7.2f%% sharpe=%6.3f calmar=%6.3f maxDD=%6.2f%% trades=%d\n",
		name,
		m.TotalReturn*100,
		m.Sharpe,
		m.Calmar,
		m.MaxDrawdown*100,
		m.NumTrades,
	)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func gitSHA() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
