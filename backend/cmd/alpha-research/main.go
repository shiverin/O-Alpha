package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/config"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/internal/research/alphavalidation"
	"github.com/oalpha/internal/research/panelload"
)

func main() {
	var (
		symbols      = flag.String("symbols", "VOO,SPY,QQQ,IWM", "comma-separated symbols; first symbol is single-symbol target and pair Y, second is pair X")
		strategies   = flag.String("strategies", "all", "comma-separated strategies: all,ma,kalman,xsec,pair")
		timeframe    = flag.String("timeframe", "1Day", "bar timeframe")
		from         = flag.String("from", "", "inclusive start date, YYYY-MM-DD")
		to           = flag.String("to", "", "inclusive end date, YYYY-MM-DD")
		feed         = flag.String("feed", "", "market data feed filter; default repo dataset")
		adjustment   = flag.String("adjustment", "", "adjustment filter; default repo dataset")
		source       = flag.String("source", "", "market data source filter; default repo dataset")
		barsCSV      = flag.String("bars-csv", "", "optional exported bars CSV; bypasses database when set")
		initialCash  = flag.Float64("initial-cash", 100000, "initial cash")
		trainBars    = flag.Int("train-bars", 756, "walk-forward train bars")
		testBars     = flag.Int("test-bars", 126, "walk-forward test bars")
		stepBars     = flag.Int("step-bars", 126, "walk-forward step bars")
		minTrades    = flag.Int("min-trades", 30, "minimum OOS trades for promotion")
		maxGross     = flag.Float64("max-gross", 1, "portfolio max gross exposure")
		maxNet       = flag.Float64("max-net", 1, "portfolio max net exposure")
		maxSymbol    = flag.Float64("max-symbol-weight", 1, "portfolio max symbol weight")
		outputDir    = flag.String("output-dir", "", "output directory; defaults to reports batch folder")
		printSummary = flag.Bool("summary", true, "print markdown summary path and promotion table")
	)
	flag.Parse()

	ctx := context.Background()
	start, end, err := resolveDateRange(*from, *to)
	if err != nil {
		fatal(err)
	}
	symbolList := parseCSV(*symbols)
	if len(symbolList) == 0 {
		fatal(fmt.Errorf("at least one symbol is required"))
	}

	var panel backtest.AlignedBars
	if strings.TrimSpace(*barsCSV) != "" {
		panel, err = loadPanelFromCSV(*barsCSV, symbolList, *timeframe, start, end)
		if err != nil {
			fatal(err)
		}
	} else {
		cfg, err := config.Load()
		if err != nil {
			fatal(err)
		}
		pool, err := db.Open(cfg.DatabaseURL)
		if err != nil {
			fatal(err)
		}
		defer pool.Close()

		repo := db.NewBarsRepository(pool)
		panel, err = repo.GetBarsMulti(ctx, symbolList, *timeframe, start, end, db.BarQueryOptions{
			Feed:       *feed,
			Adjustment: *adjustment,
			Source:     *source,
		})
		if err != nil {
			fatal(err)
		}
	}
	if len(panel.Times) == 0 {
		fatal(fmt.Errorf("no aligned bars found for %s", strings.Join(symbolList, ",")))
	}
	panel.Symbols = orderSymbols(panel.Symbols, symbolList)

	validationCfg := alphavalidation.DefaultValidationConfig()
	validationCfg.InitialCash = *initialCash
	validationCfg.TrainBars = *trainBars
	validationCfg.TestBars = *testBars
	validationCfg.StepBars = *stepBars
	validationCfg.MinOOSTrades = *minTrades
	validationCfg.MaxGrossExposure = *maxGross
	validationCfg.MaxNetExposure = *maxNet
	validationCfg.MaxSymbolWeight = *maxSymbol
	validationCfg.DataQualityPass = true
	validationCfg.NoLookaheadPass = true

	benchmarks := alphavalidation.BenchmarkFactories(panel.Symbols)
	candidates := alphavalidation.CandidateFactories(panel.Symbols, parseCSV(*strategies))
	if len(candidates) == 0 {
		fatal(fmt.Errorf("no candidate strategies selected for symbols=%s strategies=%s", strings.Join(panel.Symbols, ","), *strategies))
	}

	report, err := alphavalidation.RunValidation(ctx, panel, benchmarks, candidates, validationCfg)
	if err != nil {
		fatal(err)
	}

	dir := strings.TrimSpace(*outputDir)
	if dir == "" {
		dir = filepath.Join("..", "reports", "batches", time.Now().UTC().Format("2006-01-02")+"_alpha_validation")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fatal(err)
	}
	base := reportBaseName(panel.Symbols, *timeframe)
	jsonPath := filepath.Join(dir, base+".json")
	mdPath := filepath.Join(dir, base+".md")
	if err := writeJSON(jsonPath, report); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(mdPath, []byte(report.Markdown()), 0o644); err != nil {
		fatal(err)
	}

	if *printSummary {
		fmt.Printf("alpha validation report written:\n  %s\n  %s\n\n", jsonPath, mdPath)
		for _, candidate := range report.Candidates {
			reason := "pass"
			if len(candidate.PromotionDecision.Reasons) > 0 {
				reason = candidate.PromotionDecision.Reasons[0]
			}
			fmt.Printf("%-28s promote=%-5t return=%7.2f%% sharpe=%6.3f dsr=%5.3f pbo=%5.3f reason=%s\n",
				candidate.Name,
				candidate.PromotionDecision.Promote,
				candidate.Primary.Metrics.TotalReturn*100,
				candidate.Primary.Metrics.Sharpe,
				candidate.Primary.Metrics.DSR,
				candidate.Primary.Metrics.PBO,
				reason,
			)
		}
	}
}

func resolveDateRange(from, to string) (time.Time, time.Time, error) {
	end := time.Now().UTC()
	start := end.AddDate(-10, 0, 0)
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

func loadPanelFromCSV(path string, symbols []string, timeframe string, start, end time.Time) (backtest.AlignedBars, error) {
	return panelload.LoadPanelFromCSV(path, symbols, timeframe, start, end)
}

func parseDate(value string, endOfDay bool) (time.Time, error) {
	return panelload.ParseDate(value, endOfDay)
}

func parseCSV(value string) []string {
	return panelload.ParseCSVSymbols(value)
}

func orderSymbols(aligned []string, requested []string) []string {
	return panelload.OrderSymbols(aligned, requested)
}

func reportBaseName(symbols []string, timeframe string) string {
	joined := strings.ToLower(strings.Join(symbols, "_"))
	joined = strings.ReplaceAll(joined, "/", "_")
	if len(joined) > 140 {
		digest := sha1.Sum([]byte(joined))
		hash := hex.EncodeToString(digest[:])[:12]
		joined = fmt.Sprintf("%s_%dsymbols_%s", joined[:120], len(symbols), hash)
	}
	return fmt.Sprintf("%s_%s_alpha_validation", joined, strings.ToLower(timeframe))
}

func writeJSON(path string, value interface{}) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
