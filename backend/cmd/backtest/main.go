package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oalpha/internal/agent"
	"github.com/oalpha/internal/agent/evaluation"
	"github.com/oalpha/internal/config"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/pkg/models"
)

func main() {
	var (
		symbol               = flag.String("symbol", "VOO", "symbol to backtest")
		timeframe            = flag.String("timeframe", "1Day", "bar timeframe: 1Min, 5Min, 15Min, 1Hour, or 1Day")
		from                 = flag.String("from", "", "inclusive start date, YYYY-MM-DD")
		to                   = flag.String("to", "", "inclusive end date, YYYY-MM-DD")
		regimeModes          = flag.String("regime-modes", "none,overlay", "comma-separated regime modes: none,overlay")
		workerModes          = flag.String("worker-modes", "worker_overlay,worker_none", "comma-separated worker parity modes; empty disables worker parity")
		trainBars            = flag.Int("train-bars", 126, "walk-forward training bars per fold")
		testBars             = flag.Int("test-bars", 21, "walk-forward test bars per fold")
		stepBars             = flag.Int("step-bars", 21, "walk-forward step bars")
		workerWarmupBars     = flag.Int("worker-warmup-bars", 51, "initial bars used to warm the worker parity simulator")
		workerCalibrateEvery = flag.Int("worker-calibrate-every", 500, "bars between worker parity recalibrations; 0 disables")
		workerMaxBars        = flag.Int("worker-max-bars", 10000, "maximum rolling bars retained by worker parity")
		initialCash          = flag.Float64("initial-cash", 100000, "initial backtest cash")
		minTrades            = flag.Int("min-trades", 10, "minimum trades required for fitted-model promotion")
		outputPath           = flag.String("output", "", "optional JSON report path")
		tradesOutputDir      = flag.String("trades-output-dir", "", "optional directory for CSV trade ledgers")
		printJSON            = flag.Bool("json", false, "print full JSON report to stdout")
		verbose              = flag.Bool("verbose", false, "print per-bar strategy logs")
	)
	flag.Parse()

	if !*verbose {
		log.SetOutput(io.Discard)
	}

	cfg, err := config.Load()
	if err != nil {
		fatal(err)
	}

	ctx := context.Background()
	pool, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		fatal(err)
	}
	defer pool.Close()

	repo := db.NewBarsRepository(pool)
	start, end, err := resolveDateRange(*from, *to)
	if err != nil {
		fatal(err)
	}

	bars, err := repo.GetBars(ctx, *symbol, *timeframe, start, end)
	if err != nil {
		fatal(err)
	}
	if len(bars) == 0 {
		fatal(fmt.Errorf("no bars found for %s %s between %s and %s", *symbol, *timeframe, start.Format(time.RFC3339), end.Format(time.RFC3339)))
	}

	modes, err := parseRegimeModes(*regimeModes)
	if err != nil {
		fatal(err)
	}

	report, err := evaluation.RunWalkForwardRegimeComparison(ctx, bars, modes, evaluation.RegimeComparisonConfig{
		WindowSize:  50,
		TrainBars:   *trainBars,
		TestBars:    *testBars,
		StepBars:    *stepBars,
		InitialCash: *initialCash,
		RiskProfile: agent.RiskProfileModerate,
		MinTrades:   *minTrades,
		Symbol:      strings.ToUpper(strings.TrimSpace(*symbol)),
		Timeframe:   *timeframe,
	})
	if err != nil {
		fatal(err)
	}

	parsedWorkerModes, err := parseWorkerParityModes(*workerModes)
	if err != nil {
		fatal(err)
	}
	if len(parsedWorkerModes) > 0 {
		workerConfig := evaluation.WorkerParityConfig{
			WarmupBars:     *workerWarmupBars,
			MaxBars:        *workerMaxBars,
			CalibrateEvery: *workerCalibrateEvery,
			InitialCash:    *initialCash,
			RiskProfile:    agent.RiskProfileModerate,
			HMMWindowSize:  50,
			Symbol:         strings.ToUpper(strings.TrimSpace(*symbol)),
			Timeframe:      *timeframe,
		}
		report.WorkerResults = make(map[evaluation.WorkerParityMode]evaluation.WorkerParityResult, len(parsedWorkerModes))
		for _, mode := range parsedWorkerModes {
			result, err := evaluation.RunWorkerParityBacktest(ctx, bars, mode, workerConfig)
			if err != nil {
				fatal(err)
			}
			report.WorkerResults[mode] = result
		}
		report.WorkerBuyAndHold, err = evaluation.RunWorkerParityBuyAndHold(bars, workerConfig)
		if err != nil {
			fatal(err)
		}
	}

	if strings.TrimSpace(*tradesOutputDir) != "" {
		if err := writeTradeCSVs(report, *tradesOutputDir); err != nil {
			fatal(err)
		}
	}

	if *outputPath != "" || *printJSON {
		payload, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fatal(err)
		}
		if *outputPath != "" {
			if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
				fatal(err)
			}
			if err := os.WriteFile(*outputPath, payload, 0o644); err != nil {
				fatal(err)
			}
		}
		if *printJSON {
			fmt.Println(string(payload))
			return
		}
	}

	printSummary(report, len(bars), bars[0].Time, bars[len(bars)-1].Time, *outputPath, *tradesOutputDir)
}

func resolveDateRange(from, to string) (time.Time, time.Time, error) {
	end := time.Now().UTC()
	start := end.Add(-730 * 24 * time.Hour)
	if strings.TrimSpace(from) != "" {
		parsed, err := parseDate(from, false)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		start = parsed
	}
	if strings.TrimSpace(to) != "" {
		parsed, err := parseDate(to, true)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end = parsed
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("to must be after from")
	}
	return start, end, nil
}

func parseDate(value string, endOfDay bool) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date %q: %w", value, err)
	}
	parsed = parsed.UTC()
	if endOfDay {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return parsed, nil
}

func parseRegimeModes(value string) ([]agent.RegimeMode, error) {
	var modes []agent.RegimeMode
	for _, raw := range strings.Split(value, ",") {
		mode := agent.RegimeMode(strings.ToLower(strings.TrimSpace(raw)))
		if mode == "" {
			continue
		}
		switch mode {
		case agent.RegimeModeNone, agent.RegimeModeOverlay:
			modes = append(modes, mode)
		default:
			return nil, fmt.Errorf("unsupported regime mode: %s", mode)
		}
	}
	if len(modes) == 0 {
		return nil, fmt.Errorf("at least one regime mode is required")
	}
	return modes, nil
}

func parseWorkerParityModes(value string) ([]evaluation.WorkerParityMode, error) {
	var modes []evaluation.WorkerParityMode
	for _, raw := range strings.Split(value, ",") {
		mode := evaluation.WorkerParityMode(strings.ToLower(strings.TrimSpace(raw)))
		if mode == "" {
			continue
		}
		switch mode {
		case evaluation.WorkerParityOverlay, evaluation.WorkerParityNoHMM:
			modes = append(modes, mode)
		default:
			return nil, fmt.Errorf("unsupported worker parity mode: %s", mode)
		}
	}
	return modes, nil
}

func writeTradeCSVs(report evaluation.RegimeComparisonReport, outputDir string) error {
	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		return nil
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	for mode, result := range report.Results {
		if result.Backtest == nil {
			continue
		}
		path := filepath.Join(outputDir, tradeCSVName(report.Symbol, report.Timeframe, mode.String()))
		if err := writeTradeCSV(path, report.Symbol, report.Timeframe, mode.String(), result.Backtest.Trades); err != nil {
			return err
		}
	}
	if report.BuyAndHold != nil {
		path := filepath.Join(outputDir, tradeCSVName(report.Symbol, report.Timeframe, "buyhold"))
		if err := writeTradeCSV(path, report.Symbol, report.Timeframe, "buyhold", report.BuyAndHold.Trades); err != nil {
			return err
		}
	}
	for mode, result := range report.WorkerResults {
		if result.Backtest == nil {
			continue
		}
		path := filepath.Join(outputDir, tradeCSVName(report.Symbol, report.Timeframe, mode.String()))
		if err := writeTradeCSV(path, report.Symbol, report.Timeframe, mode.String(), result.Backtest.Trades); err != nil {
			return err
		}
	}
	if report.WorkerBuyAndHold != nil {
		path := filepath.Join(outputDir, tradeCSVName(report.Symbol, report.Timeframe, "worker_buyhold"))
		if err := writeTradeCSV(path, report.Symbol, report.Timeframe, "worker_buyhold", report.WorkerBuyAndHold.Trades); err != nil {
			return err
		}
	}
	return nil
}

func tradeCSVName(symbol, timeframe, mode string) string {
	name := strings.ToLower(strings.Join([]string{symbol, timeframe, mode, "trades.csv"}, "_"))
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

func writeTradeCSV(path, symbol, timeframe, mode string, trades []models.Trade) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"symbol", "timeframe", "mode", "entry_time", "exit_time", "entry_price", "exit_price", "quantity", "entry_value", "exit_value", "pnl", "return_pct"}
	if err := writer.Write(header); err != nil {
		return err
	}
	for _, trade := range trades {
		row := []string{
			symbol,
			timeframe,
			mode,
			trade.EntryTime.Format(time.RFC3339),
			trade.ExitTime.Format(time.RFC3339),
			formatFloat(trade.EntryPrice),
			formatFloat(trade.ExitPrice),
			formatFloat(trade.Quantity),
			formatFloat(trade.EntryValue),
			formatFloat(trade.ExitValue),
			formatFloat(trade.PnL),
			formatFloat(trade.ReturnPct),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return writer.Error()
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 6, 64)
}

func printSummary(report evaluation.RegimeComparisonReport, barCount int, first, last time.Time, outputPath string, tradesOutputDir string) {
	fmt.Printf("Walk-forward regime comparison: %s %s\n", report.Symbol, report.Timeframe)
	fmt.Printf("Bars: %d | first=%s | last=%s | folds=%d\n",
		barCount, first.Format(time.RFC3339), last.Format(time.RFC3339), report.FoldCount)
	if outputPath != "" {
		fmt.Printf("JSON report: %s\n", outputPath)
	}
	if strings.TrimSpace(tradesOutputDir) != "" {
		fmt.Printf("Trade ledgers: %s\n", tradesOutputDir)
	}
	fmt.Println()
	fmt.Println("Mode         Return    AnnRet    Sharpe   Sortino  Calmar   MaxDD    Trades  PF      Exposure")

	modes := make([]agent.RegimeMode, 0, len(report.Results))
	for mode := range report.Results {
		modes = append(modes, mode)
	}
	sort.Slice(modes, func(i, j int) bool {
		return modes[i].String() < modes[j].String()
	})
	for _, mode := range modes {
		result := report.Results[mode]
		bt := result.Backtest
		fmt.Printf("%-12s %8.2f%% %8.2f%% %8.3f %8.3f %7.3f %7.2f%% %7d %7.3f %8.2f%%\n",
			mode.String(),
			bt.TotalReturn*100,
			bt.AnnualizedReturn*100,
			bt.Sharpe,
			bt.Sortino,
			bt.Calmar,
			bt.MaxDrawdown*100,
			bt.NumTrades,
			bt.ProfitFactor,
			bt.ExposurePercent*100,
		)
	}
	if report.BuyAndHold != nil {
		bt := report.BuyAndHold
		fmt.Printf("%-12s %8.2f%% %8.2f%% %8.3f %8.3f %7.3f %7.2f%% %7d %7.3f %8.2f%%\n",
			"buyhold",
			bt.TotalReturn*100,
			bt.AnnualizedReturn*100,
			bt.Sharpe,
			bt.Sortino,
			bt.Calmar,
			bt.MaxDrawdown*100,
			bt.NumTrades,
			bt.ProfitFactor,
			bt.ExposurePercent*100,
		)
	}

	if len(report.WorkerResults) > 0 {
		fmt.Println()
		fmt.Println("Worker-Parity Modes")
		fmt.Println("Mode               Return    AnnRet    Sharpe   Sortino  Calmar   MaxDD    Trades  PF      Exposure  Warmup  EvalBars")
		workerModes := make([]evaluation.WorkerParityMode, 0, len(report.WorkerResults))
		for mode := range report.WorkerResults {
			workerModes = append(workerModes, mode)
		}
		sort.Slice(workerModes, func(i, j int) bool {
			return workerModes[i].String() < workerModes[j].String()
		})
		for _, mode := range workerModes {
			result := report.WorkerResults[mode]
			bt := result.Backtest
			fmt.Printf("%-18s %8.2f%% %8.2f%% %8.3f %8.3f %7.3f %7.2f%% %7d %7.3f %8.2f%% %7d %8d\n",
				mode.String(),
				bt.TotalReturn*100,
				bt.AnnualizedReturn*100,
				bt.Sharpe,
				bt.Sortino,
				bt.Calmar,
				bt.MaxDrawdown*100,
				bt.NumTrades,
				bt.ProfitFactor,
				bt.ExposurePercent*100,
				result.WarmupBars,
				result.EvaluatedBars,
			)
		}
		if report.WorkerBuyAndHold != nil {
			bt := report.WorkerBuyAndHold
			fmt.Printf("%-18s %8.2f%% %8.2f%% %8.3f %8.3f %7.3f %7.2f%% %7d %7.3f %8.2f%% %7s %8d\n",
				"worker_buyhold",
				bt.TotalReturn*100,
				bt.AnnualizedReturn*100,
				bt.Sharpe,
				bt.Sortino,
				bt.Calmar,
				bt.MaxDrawdown*100,
				bt.NumTrades,
				bt.ProfitFactor,
				bt.ExposurePercent*100,
				"-",
				len(bt.EquityCurve),
			)
		}
	}

	fmt.Println()
	fmt.Printf("Promotion decision: promote_overlay=%t | reason=%s\n", report.Promotion.PromoteOverlay, report.Promotion.Reason)
	fmt.Printf("Gates: drawdown_improvement=%.2f%% calmar_improved=%t sortino_improved=%t sharpe_deterioration=%.2f%% return_deterioration=%.2f%% turnover_increase=%.2f%% trade_count_ok=%t\n",
		report.Promotion.DrawdownImprovement*100,
		report.Promotion.CalmarImproved,
		report.Promotion.SortinoImproved,
		report.Promotion.SharpeDeterioration*100,
		report.Promotion.ReturnDeterioration*100,
		report.Promotion.TurnoverIncrease*100,
		report.Promotion.TradeCountOK,
	)
	if report.BuyAndHold != nil {
		fmt.Println()
		fmt.Println("Active strategy excess return vs buy-and-hold:")
		for _, mode := range modes {
			result := report.Results[mode]
			fmt.Printf("  %-12s %+0.2f%%\n", mode.String(), (result.Backtest.TotalReturn-report.BuyAndHold.TotalReturn)*100)
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
