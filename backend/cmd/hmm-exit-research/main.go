package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oalpha/internal/agent/risk"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/config"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/pkg/models"
)

type researchConfig struct {
	Symbol           string  `json:"symbol"`
	Timeframe        string  `json:"timeframe"`
	From             string  `json:"from"`
	To               string  `json:"to"`
	InitialCash      float64 `json:"initial_cash"`
	WindowBars       int     `json:"window_bars"`
	TrainBars        int     `json:"train_bars"`
	RecalibrateEvery int     `json:"recalibrate_every"`
}

type dataSummary struct {
	BarCount           int       `json:"bar_count"`
	FirstBar           time.Time `json:"first_bar"`
	LastBar            time.Time `json:"last_bar"`
	FirstClose         float64   `json:"first_close"`
	LastClose          float64   `json:"last_close"`
	CloseToCloseReturn float64   `json:"close_to_close_return"`
}

type exitPolicy struct {
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	ExitRule           string  `json:"exit_rule"`
	ReentryRule        string  `json:"reentry_rule"`
	ExitFilter         string  `json:"exit_filter,omitempty"`
	ReentryFilter      string  `json:"reentry_filter,omitempty"`
	FilterWindow       int     `json:"filter_window,omitempty"`
	ExitStressProb     float64 `json:"exit_stress_prob,omitempty"`
	ReentryStressProb  float64 `json:"reentry_stress_prob,omitempty"`
	ExitConfirmBars    int     `json:"exit_confirm_bars"`
	ReentryConfirmBars int     `json:"reentry_confirm_bars"`
}

type variantResult struct {
	Policy          exitPolicy             `json:"policy"`
	Backtest        *models.BacktestResult `json:"backtest"`
	ExcessReturn    float64                `json:"excess_return"`
	DrawdownChange  float64                `json:"drawdown_change"`
	RegimeCounts    map[string]int         `json:"regime_counts"`
	SignalCounts    map[string]int         `json:"signal_counts"`
	CalibrationRuns int                    `json:"calibration_runs"`
	UpdateErrors    int                    `json:"update_errors"`
	RegimeLedger    []regimeLedgerRow      `json:"regime_ledger"`
}

type regimeLedgerRow struct {
	Index            int       `json:"index"`
	Time             time.Time `json:"time"`
	Close            float64   `json:"close"`
	Regime           string    `json:"regime"`
	Confidence       float64   `json:"confidence"`
	ProbLowVol       float64   `json:"prob_low_vol"`
	ProbMedium       float64   `json:"prob_medium"`
	ProbHighStress   float64   `json:"prob_high_stress"`
	Signal           string    `json:"signal"`
	IntendedInMarket bool      `json:"intended_in_market"`
	StressConfirm    int       `json:"stress_confirm"`
	CalmConfirm      int       `json:"calm_confirm"`
	Calibrated       bool      `json:"calibrated"`
	Reason           string    `json:"reason"`
}

type researchReport struct {
	GeneratedAt time.Time              `json:"generated_at"`
	Config      researchConfig         `json:"config"`
	Data        dataSummary            `json:"data"`
	MethodNotes []string               `json:"method_notes"`
	BuyAndHold  *models.BacktestResult `json:"buy_and_hold"`
	Variants    []variantResult        `json:"variants"`
	Artifacts   map[string]string      `json:"artifacts"`
}

func main() {
	var (
		symbol           = flag.String("symbol", "VOO", "symbol to test")
		timeframe        = flag.String("timeframe", "1Day", "bar timeframe")
		from             = flag.String("from", "2021-06-03", "inclusive start date, YYYY-MM-DD")
		to               = flag.String("to", "2026-06-03", "inclusive end date, YYYY-MM-DD")
		initialCash      = flag.Float64("initial-cash", 100000, "initial cash")
		windowBars       = flag.Int("window", 50, "HMM rolling observation window")
		trainBars        = flag.Int("train-bars", 252, "walk-forward bars used to calibrate HMM observation buckets")
		recalibrateEvery = flag.Int("recalibrate-every", 252, "bars between walk-forward bucket recalibration; 0 disables after initial fit")
		outputDir        = flag.String("output-dir", "", "output directory; defaults to reports batch folder")
		printSummary     = flag.Bool("summary", true, "print report paths and summary table")
	)
	flag.Parse()

	if *windowBars < 2 {
		fatal(fmt.Errorf("window must be at least 2"))
	}
	if *trainBars < *windowBars {
		fatal(fmt.Errorf("train-bars (%d) must be >= window (%d)", *trainBars, *windowBars))
	}

	cfg, err := config.Load()
	if err != nil {
		fatal(err)
	}

	start, end, err := resolveDateRange(*from, *to)
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
	normalizedSymbol := strings.ToUpper(strings.TrimSpace(*symbol))
	bars, err := repo.GetBars(ctx, normalizedSymbol, *timeframe, start, end)
	if err != nil {
		fatal(err)
	}
	if len(bars) <= *trainBars {
		fatal(fmt.Errorf("need more than train-bars for test: have %d bars, train-bars=%d", len(bars), *trainBars))
	}

	buyHold, err := backtest.RunBuyAndHold(bars, *initialCash)
	if err != nil {
		fatal(err)
	}

	policies := defaultPolicies()
	variants := make([]variantResult, 0, len(policies))
	for _, policy := range policies {
		result, err := runPolicy(bars, policy, buyHold, researchConfig{
			Symbol:           normalizedSymbol,
			Timeframe:        *timeframe,
			From:             start.Format("2006-01-02"),
			To:               end.Format("2006-01-02"),
			InitialCash:      *initialCash,
			WindowBars:       *windowBars,
			TrainBars:        *trainBars,
			RecalibrateEvery: *recalibrateEvery,
		})
		if err != nil {
			fatal(fmt.Errorf("%s: %w", policy.Name, err))
		}
		variants = append(variants, result)
	}

	dir := strings.TrimSpace(*outputDir)
	if dir == "" {
		dir = filepath.Join("..", "reports", "batches", time.Now().UTC().Format("2006-01-02")+"_voo_hmm_exit_research")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fatal(err)
	}

	baseName := fmt.Sprintf("%s_%s_hmm_exit_research", strings.ToLower(normalizedSymbol), strings.ToLower(*timeframe))
	jsonPath := filepath.Join(dir, baseName+".json")
	mdPath := filepath.Join(dir, baseName+".md")
	tradesPath := filepath.Join(dir, baseName+"_trades.csv")
	ledgerPath := filepath.Join(dir, baseName+"_regime_ledger.csv")

	report := researchReport{
		GeneratedAt: time.Now().UTC(),
		Config: researchConfig{
			Symbol:           normalizedSymbol,
			Timeframe:        *timeframe,
			From:             start.Format("2006-01-02"),
			To:               end.Format("2006-01-02"),
			InitialCash:      *initialCash,
			WindowBars:       *windowBars,
			TrainBars:        *trainBars,
			RecalibrateEvery: *recalibrateEvery,
		},
		Data:        summarizeData(bars),
		MethodNotes: methodNotes(),
		BuyAndHold:  buyHold,
		Variants:    variants,
		Artifacts: map[string]string{
			"json":          jsonPath,
			"markdown":      mdPath,
			"trades_csv":    tradesPath,
			"regime_ledger": ledgerPath,
		},
	}

	if err := writeJSON(jsonPath, report); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(mdPath, []byte(report.Markdown()), 0o644); err != nil {
		fatal(err)
	}
	if err := writeTradesCSV(tradesPath, report); err != nil {
		fatal(err)
	}
	if err := writeRegimeLedgerCSV(ledgerPath, report); err != nil {
		fatal(err)
	}

	if *printSummary {
		fmt.Printf("HMM exit research report written:\n  %s\n  %s\n  %s\n  %s\n\n", jsonPath, mdPath, tradesPath, ledgerPath)
		fmt.Printf("%-24s return=%8.2f%% sharpe=%6.3f maxDD=%6.2f%% trades=%3d exposure=%6.2f%%\n",
			"buy_hold",
			buyHold.TotalReturn*100,
			buyHold.Sharpe,
			buyHold.MaxDrawdown*100,
			buyHold.NumTrades,
			buyHold.ExposurePercent*100,
		)
		for _, variant := range variants {
			bt := variant.Backtest
			fmt.Printf("%-24s return=%8.2f%% excess=%8.2f%% sharpe=%6.3f maxDD=%6.2f%% trades=%3d exposure=%6.2f%%\n",
				variant.Policy.Name,
				bt.TotalReturn*100,
				variant.ExcessReturn*100,
				bt.Sharpe,
				bt.MaxDrawdown*100,
				bt.NumTrades,
				bt.ExposurePercent*100,
			)
		}
	}
}

func defaultPolicies() []exitPolicy {
	return []exitPolicy{
		{
			Name:               "regime_1_1",
			Description:        "Exit after one High Vol Stress regime print; re-enter after one non-stress print.",
			ExitRule:           "regime",
			ReentryRule:        "non_stress",
			ExitConfirmBars:    1,
			ReentryConfirmBars: 1,
		},
		{
			Name:               "regime_2_3",
			Description:        "Exit after two consecutive High Vol Stress prints; re-enter after three consecutive non-stress prints.",
			ExitRule:           "regime",
			ReentryRule:        "non_stress",
			ExitConfirmBars:    2,
			ReentryConfirmBars: 3,
		},
		{
			Name:               "prob55_40_2_3",
			Description:        "Exit after high-stress posterior >= 55% for two bars; re-enter below 40% and non-stress for three bars.",
			ExitRule:           "probability",
			ReentryRule:        "probability_non_stress",
			ExitStressProb:     0.55,
			ReentryStressProb:  0.40,
			ExitConfirmBars:    2,
			ReentryConfirmBars: 3,
		},
		{
			Name:               "prob45_30_2_3",
			Description:        "Exit after high-stress posterior >= 45% for two bars; re-enter below 30% and non-stress for three bars.",
			ExitRule:           "probability",
			ReentryRule:        "probability_non_stress",
			ExitStressProb:     0.45,
			ReentryStressProb:  0.30,
			ExitConfirmBars:    2,
			ReentryConfirmBars: 3,
		},
		{
			Name:               "lowvol_reentry_2_3",
			Description:        "Exit after high-stress posterior >= 55% for two bars; re-enter only after three Low Vol Trend bars.",
			ExitRule:           "probability",
			ReentryRule:        "low_vol",
			ExitStressProb:     0.55,
			ReentryStressProb:  0.40,
			ExitConfirmBars:    2,
			ReentryConfirmBars: 3,
		},
		{
			Name:               "regime_sma200_1_3",
			Description:        "Exit after one High Vol Stress print only when price is below its 200-day SMA; re-enter after three non-stress bars above the 200-day SMA.",
			ExitRule:           "regime",
			ReentryRule:        "non_stress",
			ExitFilter:         "below_sma",
			ReentryFilter:      "above_sma",
			FilterWindow:       200,
			ExitConfirmBars:    1,
			ReentryConfirmBars: 3,
		},
		{
			Name:               "prob55_sma200_2_3",
			Description:        "Exit after high-stress posterior >= 55% for two bars only when price is below its 200-day SMA; re-enter below 40% and above the 200-day SMA for three bars.",
			ExitRule:           "probability",
			ReentryRule:        "probability_non_stress",
			ExitFilter:         "below_sma",
			ReentryFilter:      "above_sma",
			FilterWindow:       200,
			ExitStressProb:     0.55,
			ReentryStressProb:  0.40,
			ExitConfirmBars:    2,
			ReentryConfirmBars: 3,
		},
		{
			Name:               "prob55_trend50_2_3",
			Description:        "Exit after high-stress posterior >= 55% for two bars only when the trailing 50-day return is negative; re-enter below 40% and positive 50-day return for three bars.",
			ExitRule:           "probability",
			ReentryRule:        "probability_non_stress",
			ExitFilter:         "negative_return",
			ReentryFilter:      "positive_return",
			FilterWindow:       50,
			ExitStressProb:     0.55,
			ReentryStressProb:  0.40,
			ExitConfirmBars:    2,
			ReentryConfirmBars: 3,
		},
	}
}

func runPolicy(bars []models.Bar, policy exitPolicy, benchmark *models.BacktestResult, cfg researchConfig) (variantResult, error) {
	outputs := make([]backtest.StrategyOutput, len(bars))
	for i := range outputs {
		outputs[i] = backtest.StrategyOutput{
			Signal:      models.SignalHold,
			RegimeLabel: "warmup",
		}
	}
	outputs[0] = backtest.StrategyOutput{
		Signal:          models.SignalBuy,
		PositionSizePct: 1,
		TargetWeight:    1,
		RegimeLabel:     "initial_buy",
		Engine:          "hmm_exit_research",
	}

	detector := risk.NewHMMRegimeDetector(cfg.WindowBars)
	regimeCounts := map[string]int{}
	signalCounts := map[string]int{"buy": 1, "sell": 0, "hold": len(bars) - 1}
	ledger := make([]regimeLedgerRow, 0, len(bars)-1)

	inMarket := true
	stressConfirm := 0
	calmConfirm := 0
	lastCalibrationIndex := -1
	calibrations := 0
	updateErrors := 0

	for i := 1; i < len(bars); i++ {
		reason := "hold"
		calibrated := lastCalibrationIndex >= 0

		if i < cfg.TrainBars || i+1 < cfg.WindowBars {
			outputs[i].RegimeLabel = "warmup"
			ledger = append(ledger, regimeLedgerRow{
				Index:            i,
				Time:             bars[i].Time,
				Close:            bars[i].Close,
				Regime:           "Warmup",
				Signal:           signalName(outputs[i].Signal),
				IntendedInMarket: inMarket,
				Calibrated:       calibrated,
				Reason:           "warmup",
			})
			continue
		}

		if lastCalibrationIndex < 0 || (cfg.RecalibrateEvery > 0 && i-lastCalibrationIndex >= cfg.RecalibrateEvery) {
			start := i + 1 - cfg.TrainBars
			if start < 0 {
				start = 0
			}
			encoder := risk.NewObservationEncoder(cfg.WindowBars)
			if err := encoder.FitBuckets(bars[start : i+1]); err != nil {
				return variantResult{}, fmt.Errorf("fit buckets at %s: %w", bars[i].Time.Format(time.RFC3339), err)
			}
			detector.UpdateBuckets(encoder.VolBuckets, encoder.TrendBuckets)
			lastCalibrationIndex = i
			calibrations++
			calibrated = true
		}

		regime, confidence, err := detector.Update(bars[:i+1])
		if err != nil {
			updateErrors++
			reason = "hmm_update_error"
			ledger = append(ledger, regimeLedgerRow{
				Index:            i,
				Time:             bars[i].Time,
				Close:            bars[i].Close,
				Regime:           "Unknown",
				Signal:           signalName(outputs[i].Signal),
				IntendedInMarket: inMarket,
				Calibrated:       calibrated,
				Reason:           reason,
			})
			continue
		}

		probs := detector.GetProbabilities()
		regimeLabel := regime.String()
		regimeCounts[regimeLabel]++
		outputs[i].RegimeLabel = regimeLabel
		outputs[i].Engine = "hmm_exit_research"
		outputs[i].Metadata = map[string]interface{}{
			"confidence":       confidence,
			"prob_low_vol":     probs[0],
			"prob_medium":      probs[1],
			"prob_high_stress": probs[2],
			"policy":           policy.Name,
		}

		stress := shouldExit(policy, regime, probs[2], bars, i)
		calm := shouldReenter(policy, regime, probs[2], bars, i)
		if stress {
			stressConfirm++
		} else {
			stressConfirm = 0
		}
		if calm {
			calmConfirm++
		} else {
			calmConfirm = 0
		}

		if inMarket && stressConfirm >= minConfirm(policy.ExitConfirmBars) {
			outputs[i].Signal = models.SignalSell
			outputs[i].TargetWeight = 0
			outputs[i].PositionSizePct = 0
			inMarket = false
			reason = "exit_stress"
			signalCounts["sell"]++
			signalCounts["hold"]--
			calmConfirm = 0
		} else if !inMarket && calmConfirm >= minConfirm(policy.ReentryConfirmBars) {
			outputs[i].Signal = models.SignalBuy
			outputs[i].TargetWeight = 1
			outputs[i].PositionSizePct = 1
			inMarket = true
			reason = "reenter_calm"
			signalCounts["buy"]++
			signalCounts["hold"]--
			stressConfirm = 0
		}

		ledger = append(ledger, regimeLedgerRow{
			Index:            i,
			Time:             bars[i].Time,
			Close:            bars[i].Close,
			Regime:           regimeLabel,
			Confidence:       confidence,
			ProbLowVol:       probs[0],
			ProbMedium:       probs[1],
			ProbHighStress:   probs[2],
			Signal:           signalName(outputs[i].Signal),
			IntendedInMarket: inMarket,
			StressConfirm:    stressConfirm,
			CalmConfirm:      calmConfirm,
			Calibrated:       calibrated,
			Reason:           reason,
		})
	}

	result, err := backtest.RunBacktestWithOutputs(bars, outputs, cfg.InitialCash)
	if err != nil {
		return variantResult{}, err
	}
	return variantResult{
		Policy:          policy,
		Backtest:        result,
		ExcessReturn:    result.TotalReturn - benchmark.TotalReturn,
		DrawdownChange:  benchmark.MaxDrawdown - result.MaxDrawdown,
		RegimeCounts:    regimeCounts,
		SignalCounts:    signalCounts,
		CalibrationRuns: calibrations,
		UpdateErrors:    updateErrors,
		RegimeLedger:    ledger,
	}, nil
}

func shouldExit(policy exitPolicy, regime risk.MarketRegime, highStressProb float64, bars []models.Bar, index int) bool {
	if !marketFilterPass(policy.ExitFilter, policy.FilterWindow, bars, index) {
		return false
	}
	switch policy.ExitRule {
	case "regime":
		return regime == risk.RegimeHighVolStress
	case "probability":
		return highStressProb >= policy.ExitStressProb
	default:
		return false
	}
}

func shouldReenter(policy exitPolicy, regime risk.MarketRegime, highStressProb float64, bars []models.Bar, index int) bool {
	if !marketFilterPass(policy.ReentryFilter, policy.FilterWindow, bars, index) {
		return false
	}
	switch policy.ReentryRule {
	case "non_stress":
		return regime != risk.RegimeHighVolStress
	case "probability_non_stress":
		return regime != risk.RegimeHighVolStress && highStressProb <= policy.ReentryStressProb
	case "low_vol":
		return regime == risk.RegimeLowVolTrend && highStressProb <= policy.ReentryStressProb
	default:
		return false
	}
}

func marketFilterPass(filter string, window int, bars []models.Bar, index int) bool {
	switch filter {
	case "":
		return true
	case "below_sma":
		avg, ok := trailingSMA(bars, index, window)
		return ok && bars[index].Close < avg
	case "above_sma":
		avg, ok := trailingSMA(bars, index, window)
		return ok && bars[index].Close > avg
	case "negative_return":
		ret, ok := trailingReturn(bars, index, window)
		return ok && ret < 0
	case "positive_return":
		ret, ok := trailingReturn(bars, index, window)
		return ok && ret > 0
	default:
		return true
	}
}

func trailingSMA(bars []models.Bar, index int, window int) (float64, bool) {
	if window <= 0 || index < 0 || index+1 < window {
		return 0, false
	}
	start := index + 1 - window
	sum := 0.0
	for i := start; i <= index; i++ {
		sum += bars[i].Close
	}
	return sum / float64(window), true
}

func trailingReturn(bars []models.Bar, index int, window int) (float64, bool) {
	if window <= 0 || index-window < 0 {
		return 0, false
	}
	startClose := bars[index-window].Close
	if startClose <= 0 {
		return 0, false
	}
	return bars[index].Close/startClose - 1, true
}

func minConfirm(v int) int {
	if v < 1 {
		return 1
	}
	return v
}

func summarizeData(bars []models.Bar) dataSummary {
	if len(bars) == 0 {
		return dataSummary{}
	}
	first := bars[0]
	last := bars[len(bars)-1]
	closeReturn := 0.0
	if first.Close > 0 {
		closeReturn = last.Close/first.Close - 1
	}
	return dataSummary{
		BarCount:           len(bars),
		FirstBar:           first.Time,
		LastBar:            last.Time,
		FirstClose:         first.Close,
		LastClose:          last.Close,
		CloseToCloseReturn: closeReturn,
	}
}

func methodNotes() []string {
	return []string{
		"Benchmark is VOO buy-and-hold through the same backtest engine.",
		"All active variants start by buying VOO, then go to cash on HMM high-volatility stress, and re-enter VOO when the policy's calm condition is confirmed.",
		"Signals are generated after a bar closes and execute at the next bar open, matching the existing single-symbol backtest engine.",
		"HMM observation buckets are calibrated only from historical bars available at that point in the walk-forward timeline.",
		"The HMM state is high-volatility stress, not a pure bearish-direction classifier.",
		"No explicit transaction-cost or slippage model is added by this research runner.",
	}
}

func (r researchReport) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s HMM Exit Research\n\n", r.Config.Symbol)
	fmt.Fprintf(&b, "Generated: %s\n\n", r.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "Data: %d %s bars from %s to %s. Close-to-close return: %s.\n\n",
		r.Data.BarCount,
		r.Config.Timeframe,
		r.Data.FirstBar.Format("2006-01-02"),
		r.Data.LastBar.Format("2006-01-02"),
		formatPct(r.Data.CloseToCloseReturn),
	)

	b.WriteString("## Method\n\n")
	for _, note := range r.MethodNotes {
		fmt.Fprintf(&b, "- %s\n", note)
	}
	b.WriteString("\n")

	bestReturn := -1
	bestDrawdown := -1
	for i := range r.Variants {
		if bestReturn < 0 || r.Variants[i].Backtest.TotalReturn > r.Variants[bestReturn].Backtest.TotalReturn {
			bestReturn = i
		}
		if bestDrawdown < 0 || r.Variants[i].Backtest.MaxDrawdown < r.Variants[bestDrawdown].Backtest.MaxDrawdown {
			bestDrawdown = i
		}
	}

	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "Benchmark buy-and-hold return is %s with Sharpe %.3f and max drawdown %s.\n\n",
		formatPct(r.BuyAndHold.TotalReturn),
		r.BuyAndHold.Sharpe,
		formatPct(r.BuyAndHold.MaxDrawdown),
	)
	if bestReturn >= 0 {
		v := r.Variants[bestReturn]
		fmt.Fprintf(&b, "Best HMM exit variant by total return is `%s` at %s, excess %s versus buy-and-hold.\n\n",
			v.Policy.Name,
			formatPct(v.Backtest.TotalReturn),
			formatPct(v.ExcessReturn),
		)
	}
	if bestDrawdown >= 0 {
		v := r.Variants[bestDrawdown]
		fmt.Fprintf(&b, "Lowest HMM exit max drawdown is `%s` at %s, a %s drawdown change versus buy-and-hold.\n\n",
			v.Policy.Name,
			formatPct(v.Backtest.MaxDrawdown),
			formatPct(v.DrawdownChange),
		)
	}

	b.WriteString("| Strategy | Total Return | Excess vs B&H | Ann. Return | Sharpe | Sortino | Calmar | Max DD | Trades | Exposure | Turnover |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	fmt.Fprintf(&b, "| buy_hold | %s | %s | %s | %.3f | %.3f | %.3f | %s | %d | %s | %.2f |\n",
		formatPct(r.BuyAndHold.TotalReturn),
		formatPct(0),
		formatPct(r.BuyAndHold.AnnualizedReturn),
		r.BuyAndHold.Sharpe,
		r.BuyAndHold.Sortino,
		r.BuyAndHold.Calmar,
		formatPct(r.BuyAndHold.MaxDrawdown),
		r.BuyAndHold.NumTrades,
		formatPct(r.BuyAndHold.ExposurePercent),
		r.BuyAndHold.Turnover,
	)
	for _, v := range r.Variants {
		bt := v.Backtest
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %.3f | %.3f | %.3f | %s | %d | %s | %.2f |\n",
			v.Policy.Name,
			formatPct(bt.TotalReturn),
			formatPct(v.ExcessReturn),
			formatPct(bt.AnnualizedReturn),
			bt.Sharpe,
			bt.Sortino,
			bt.Calmar,
			formatPct(bt.MaxDrawdown),
			bt.NumTrades,
			formatPct(bt.ExposurePercent),
			bt.Turnover,
		)
	}
	b.WriteString("\n")

	b.WriteString("## Policy Details\n\n")
	b.WriteString("| Policy | Description | Re-entry rule | Filter | Sells | Buys | Calibration runs | Update errors |\n")
	b.WriteString("|---|---|---|---|---:|---:|---:|---:|\n")
	for _, v := range r.Variants {
		filter := v.Policy.ExitFilter
		if filter == "" {
			filter = "none"
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s/%d | %d | %d | %d | %d |\n",
			v.Policy.Name,
			v.Policy.Description,
			v.Policy.ReentryRule,
			filter,
			v.Policy.FilterWindow,
			v.SignalCounts["sell"],
			v.SignalCounts["buy"],
			v.CalibrationRuns,
			v.UpdateErrors,
		)
	}
	b.WriteString("\n")

	b.WriteString("## Trade Ledgers\n\n")
	for _, v := range r.Variants {
		fmt.Fprintf(&b, "### %s\n\n", v.Policy.Name)
		if len(v.Backtest.Trades) == 0 {
			b.WriteString("No closed trades.\n\n")
			continue
		}
		b.WriteString("| # | Entry | Exit | Entry Price | Exit Price | Return | PnL |\n")
		b.WriteString("|---:|---|---|---:|---:|---:|---:|\n")
		for i, trade := range v.Backtest.Trades {
			fmt.Fprintf(&b, "| %d | %s | %s | %.2f | %.2f | %s | %.2f |\n",
				i+1,
				trade.EntryTime.Format("2006-01-02"),
				trade.ExitTime.Format("2006-01-02"),
				trade.EntryPrice,
				trade.ExitPrice,
				formatPct(trade.ReturnPct),
				trade.PnL,
			)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func writeJSON(path string, value interface{}) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}

func writeTradesCSV(path string, report researchReport) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	w := csv.NewWriter(file)
	defer w.Flush()
	if err := w.Write([]string{"strategy", "trade_number", "entry_time", "exit_time", "entry_price", "exit_price", "quantity", "entry_value", "exit_value", "pnl", "return_pct"}); err != nil {
		return err
	}
	for _, variant := range report.Variants {
		for i, trade := range variant.Backtest.Trades {
			row := []string{
				variant.Policy.Name,
				fmt.Sprintf("%d", i+1),
				trade.EntryTime.Format(time.RFC3339),
				trade.ExitTime.Format(time.RFC3339),
				fmt.Sprintf("%.6f", trade.EntryPrice),
				fmt.Sprintf("%.6f", trade.ExitPrice),
				fmt.Sprintf("%.8f", trade.Quantity),
				fmt.Sprintf("%.2f", trade.EntryValue),
				fmt.Sprintf("%.2f", trade.ExitValue),
				fmt.Sprintf("%.2f", trade.PnL),
				fmt.Sprintf("%.8f", trade.ReturnPct),
			}
			if err := w.Write(row); err != nil {
				return err
			}
		}
	}
	return w.Error()
}

func writeRegimeLedgerCSV(path string, report researchReport) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	w := csv.NewWriter(file)
	defer w.Flush()
	if err := w.Write([]string{"strategy", "index", "time", "close", "regime", "confidence", "prob_low_vol", "prob_medium", "prob_high_stress", "signal", "intended_in_market", "stress_confirm", "calm_confirm", "calibrated", "reason"}); err != nil {
		return err
	}
	for _, variant := range report.Variants {
		for _, row := range variant.RegimeLedger {
			record := []string{
				variant.Policy.Name,
				fmt.Sprintf("%d", row.Index),
				row.Time.Format(time.RFC3339),
				fmt.Sprintf("%.6f", row.Close),
				row.Regime,
				fmt.Sprintf("%.8f", row.Confidence),
				fmt.Sprintf("%.8f", row.ProbLowVol),
				fmt.Sprintf("%.8f", row.ProbMedium),
				fmt.Sprintf("%.8f", row.ProbHighStress),
				row.Signal,
				fmt.Sprintf("%t", row.IntendedInMarket),
				fmt.Sprintf("%d", row.StressConfirm),
				fmt.Sprintf("%d", row.CalmConfirm),
				fmt.Sprintf("%t", row.Calibrated),
				row.Reason,
			}
			if err := w.Write(record); err != nil {
				return err
			}
		}
	}
	return w.Error()
}

func resolveDateRange(from, to string) (time.Time, time.Time, error) {
	end := time.Now().UTC()
	start := end.AddDate(-5, 0, 0)
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

func signalName(signal models.Signal) string {
	switch signal {
	case models.SignalBuy:
		return "buy"
	case models.SignalSell:
		return "sell"
	default:
		return "hold"
	}
}

func formatPct(v float64) string {
	return fmt.Sprintf("%.2f%%", v*100)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
