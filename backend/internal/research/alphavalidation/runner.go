package alphavalidation

import (
	"context"
	"fmt"
	"math"
<<<<<<< HEAD
	"time"

	"github.com/oalpha/internal/backtest"
	btvalidation "github.com/oalpha/internal/backtest/validation"
	"github.com/oalpha/pkg/models"
)

func RunValidation(
	ctx context.Context,
	panel backtest.AlignedBars,
	benchmarks []StrategyFactory,
	candidates []StrategyFactory,
	cfg ValidationConfig,
) (AlphaValidationReport, error) {
	if err := validatePanel(panel); err != nil {
		return AlphaValidationReport{}, err
	}
	cfg = cfg.withDefaults()

	report := AlphaValidationReport{
		GeneratedAt: time.Now().UTC(),
		Symbols:     append([]string(nil), panel.Symbols...),
		Timeframe:   panel.Timeframe,
		Start:       panel.Times[0],
		End:         panel.Times[len(panel.Times)-1],
		BarCount:    len(panel.Times),
		Config:      cfg,
		Notes: []string{
			"All candidate runs use target-weight execution at next-bar open.",
			"Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.",
			"Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.",
		},
	}

	benchmarkResults := make(map[string]ScenarioResult)
	for _, benchmark := range benchmarks {
		if err := validateFactory(benchmark); err != nil {
			return AlphaValidationReport{}, err
		}
		result := runScenario(ctx, panel, benchmark, cfg.CostScenarios[0], cfg)
		if result.Error != "" {
			report.Notes = append(report.Notes, fmt.Sprintf("benchmark %s failed: %s", benchmark.Name, result.Error))
			continue
		}
		benchmarkResults[benchmark.Name] = result
		report.Benchmarks = append(report.Benchmarks, BenchmarkReport{
			Name:    benchmark.Name,
			Metrics: result.Metrics,
		})
	}

	for _, candidate := range candidates {
		candidateReport := runCandidate(ctx, panel, candidate, benchmarkResults, cfg)
		report.Candidates = append(report.Candidates, candidateReport)
	}
	return report, nil
}

func runCandidate(
	ctx context.Context,
	panel backtest.AlignedBars,
	candidate StrategyFactory,
	benchmarkResults map[string]ScenarioResult,
	cfg ValidationConfig,
) CandidateReport {
	report := CandidateReport{
		Name:        candidate.Name,
		Family:      candidate.Family,
		Benchmark:   candidate.Benchmark,
		CostStress:  make([]ScenarioResult, 0, len(cfg.CostScenarios)),
		Diagnostics: make([]string, 0),
	}
	if err := validateFactory(candidate); err != nil {
		report.Primary = ScenarioResult{Scenario: cfg.CostScenarios[0].Name, Error: err.Error()}
		report.Diagnostics = append(report.Diagnostics, err.Error())
		return report
	}
	report.AuditMetadata = auditCandidateMetadata(ctx, panel, candidate)
	benchmark, ok := benchmarkResults[candidate.Benchmark]
	if !ok {
		report.Diagnostics = append(report.Diagnostics, fmt.Sprintf("benchmark %s unavailable; using zero benchmark metrics", candidate.Benchmark))
	}

	for _, scenario := range cfg.CostScenarios {
		result := runScenario(ctx, panel, candidate, scenario, cfg)
		report.CostStress = append(report.CostStress, result)
		if scenario.Name == cfg.CostScenarios[0].Name {
			report.Primary = result
		}
	}
	if report.Primary.Error != "" {
		report.PromotionDecision = btvalidation.PromotionDecision{
			Promote: false,
			Reasons: []string{
				report.Primary.Error,
			},
			Metrics:          report.Primary.Metrics,
			BenchmarkMetrics: benchmark.Metrics,
		}
		return report
	}

	variants := VariantFactories(candidate, panel.Symbols)
	report.WalkForward = runWalkForward(ctx, panel, candidate, cfg)
	pbo, estimated, pboDiagnostics, pboErr := estimatePBO(ctx, panel, variants, cfg)
	report.PBODiagnostics = pboDiagnostics
	if pboErr != nil {
		report.Diagnostics = append(report.Diagnostics, pboErr.Error())
	}

	metrics := report.Primary.Metrics
	metrics.DSR = btvalidation.DeflatedSharpeRatio(metrics.Sharpe, maxInt(2, len(panel.Times)-1), metrics.Skew, metrics.Kurtosis, maxInt(1, len(variants)))
	if estimated {
		metrics.PBO = pbo
		report.PBOEstimated = true
	} else {
		metrics.PBO = 1
		report.Diagnostics = append(report.Diagnostics, "PBO not estimated; promotion fails closed")
	}
	report.Primary.Metrics = metrics
	report.CostStress[0].Metrics = metrics

	promotionCfg := cfg.PromotionConfig
	promotionCfg.MinOOSTrades = cfg.MinOOSTrades
	report.PromotionDecision = btvalidation.EvaluatePromotion(metrics, benchmark.Metrics, promotionCfg, cfg.DataQualityPass, cfg.NoLookaheadPass)
	if !estimated {
		report.PromotionDecision.Promote = false
		report.PromotionDecision.Reasons = append(report.PromotionDecision.Reasons, "PBO was not estimated from walk-forward variants")
		report.PromotionDecision.Metrics = metrics
	}
	for _, reason := range candidateResearchGuardrails(candidate, panel) {
		report.PromotionDecision.Promote = false
		report.PromotionDecision.Reasons = append(report.PromotionDecision.Reasons, reason)
		report.Diagnostics = append(report.Diagnostics, reason)
	}
	return report
}

func candidateResearchGuardrails(candidate StrategyFactory, panel backtest.AlignedBars) []string {
	switch candidate.Family {
	case "xsec_momentum":
		if len(panel.Symbols) < 50 {
			return []string{fmt.Sprintf("xsec universe size %d below research minimum 50", len(panel.Symbols))}
		}
	case "kalman_cointegration":
		return []string{"pair sleeve requires offline approved cointegration candidate and live shortability gate before promotion"}
	}
	return nil
}

func runScenario(
	ctx context.Context,
	panel backtest.AlignedBars,
	factory StrategyFactory,
	scenario CostScenario,
	cfg ValidationConfig,
) ScenarioResult {
	return runScenarioWithMetricStart(ctx, panel, factory, scenario, cfg, 0)
}

func runScenarioWithMetricStart(
	ctx context.Context,
	panel backtest.AlignedBars,
	factory StrategyFactory,
	scenario CostScenario,
	cfg ValidationConfig,
	metricStart int,
) ScenarioResult {
	result := ScenarioResult{Scenario: scenario.Name}
	strategy := factory.New()
	if strategy == nil {
		result.Error = fmt.Sprintf("strategy factory %s returned nil", factory.Name)
		return result
	}
	backtestResult, err := backtest.RunPortfolioBacktest(ctx, panel, strategy, backtest.PortfolioBacktestConfig{
		InitialCash:      cfg.InitialCash,
		AllowShorts:      factory.AllowShorts,
		MaxGrossExposure: cfg.MaxGrossExposure,
		MaxNetExposure:   cfg.MaxNetExposure,
		MaxSymbolWeight:  cfg.MaxSymbolWeight,
		CostModel:        scenario.CostModel,
	})
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if metricStart > 0 {
		return trimScenarioResult(panel, backtestResult, scenario.Name, metricStart)
	}
	result.Metrics = backtestResult.Metrics
	result.FinalEquity = finalEquity(backtestResult.EquityCurve)
	result.NumTrades = len(backtestResult.Trades)
	for _, trade := range backtestResult.Trades {
		result.FeesPaid += trade.Fees
		result.SlippageCost += trade.SlippageCost
	}
	return result
}

func trimScenarioResult(
	panel backtest.AlignedBars,
	result *backtest.PortfolioBacktestResult,
	scenarioName string,
	metricStart int,
) ScenarioResult {
	out := ScenarioResult{Scenario: scenarioName}
	if result == nil || len(result.EquityCurve) == 0 {
		out.Error = "empty portfolio result"
		return out
	}
	if metricStart < 0 {
		metricStart = 0
	}
	if metricStart >= len(result.EquityCurve) {
		out.Error = fmt.Sprintf("metric start %d beyond equity curve length %d", metricStart, len(result.EquityCurve))
		return out
	}

	equities := make([]float64, 0, len(result.EquityCurve)-metricStart)
	for _, point := range result.EquityCurve[metricStart:] {
		equities = append(equities, point.Equity)
	}
	gross := make([]float64, 0, len(result.PositionCurve)-metricStart)
	net := make([]float64, 0, len(result.PositionCurve)-metricStart)
	for _, snapshot := range result.PositionCurve[metricStart:] {
		gross = append(gross, snapshot.GrossExposure)
		net = append(net, snapshot.NetExposure)
	}

	startTime := result.EquityCurve[metricStart].Time
	if metricStart < len(panel.Times) {
		startTime = panel.Times[metricStart]
	}
	var turnoverValue float64
	pnls := make([]float64, 0)
	for _, trade := range result.Trades {
		if trade.Time.Before(startTime) {
			continue
		}
		out.NumTrades++
		out.FeesPaid += trade.Fees
		out.SlippageCost += trade.SlippageCost
		turnoverValue += math.Abs(trade.Notional)
		if isClosingFill(trade) {
			pnls = append(pnls, trade.RealizedPnL)
		}
	}
	turnover := 0.0
	if len(equities) > 0 && equities[0] > 0 {
		turnover = turnoverValue / equities[0]
	}
	out.Metrics = backtest.ComputePortfolioMetrics(equities, pnls, gross, net, turnover)
	out.FinalEquity = equities[len(equities)-1]
	return out
}

func isClosingFill(trade backtest.SimulatedTrade) bool {
	return (trade.PositionSide == backtest.PositionSideLong && trade.Side == backtest.OrderSideSell) ||
		(trade.PositionSide == backtest.PositionSideShort && trade.Side == backtest.OrderSideBuy)
}

func runWalkForward(
	ctx context.Context,
	panel backtest.AlignedBars,
	factory StrategyFactory,
	cfg ValidationConfig,
) []WindowResult {
	splits, err := btvalidation.GenerateMLWalkForwardSplits(len(panel.Times), btvalidation.MLWalkForwardConfig{
		TrainSize: cfg.TrainBars,
		TestSize:  cfg.TestBars,
		StepSize:  cfg.StepBars,
	})
	if err != nil {
		return []WindowResult{{Error: err.Error()}}
	}
	out := make([]WindowResult, 0, len(splits))
	for _, split := range splits {
		window := WindowResult{
			Fold:       split.Fold,
			TrainStart: split.TrainStart,
			TrainEnd:   split.TrainEnd,
			TestStart:  split.TestStart,
			TestEnd:    split.TestEnd,
		}
		trainResult := runScenario(ctx, slicePanel(panel, split.TrainStart, split.TrainEnd), factory, cfg.CostScenarios[0], cfg)
		testPanel := slicePanel(panel, split.TrainStart, split.TestEnd)
		testResult := runScenarioWithMetricStart(ctx, testPanel, factory, cfg.CostScenarios[0], cfg, split.TestStart-split.TrainStart)
		window.Train = trainResult.Metrics
		window.Test = testResult.Metrics
		if trainResult.Error != "" {
			window.Error = "train: " + trainResult.Error
		}
		if testResult.Error != "" {
			if window.Error != "" {
				window.Error += "; "
			}
			window.Error += "test: " + testResult.Error
		}
		out = append(out, window)
=======
	"sort"
	"strings"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

type RunnerConfig struct {
	Symbols     []string
	Timeframe   string
	Window      ValidationWindow
	TrainBars   int
	TestBars    int
	StepBars    int
	MinTrades   int
	InitialCash float64
}

func (c RunnerConfig) withDefaults() RunnerConfig {
	if c.TrainBars <= 0 {
		c.TrainBars = 756
	}
	if c.TestBars <= 0 {
		c.TestBars = 252
	}
	if c.StepBars <= 0 {
		c.StepBars = 126
	}
	if c.MinTrades <= 0 {
		c.MinTrades = 30
	}
	if c.InitialCash <= 0 {
		c.InitialCash = 100_000
	}
	c.Symbols = normalizeSymbols(c.Symbols)
	return c
}

func RunValidation(ctx context.Context, source DataSource, strategyFamily string, cfg RunnerConfig) (ValidationReport, error) {
	cfg = cfg.withDefaults()
	family, err := ResolveFamily(strategyFamily)
	if err != nil {
		return ValidationReport{}, err
	}
	barsBySymbol, err := source.LoadBars(ctx, cfg.Symbols, cfg.Timeframe, cfg.Window)
	if err != nil {
		return ValidationReport{}, err
	}
	benchmarkBars, ok := barsBySymbol["VOO"]
	if !ok || len(benchmarkBars) == 0 {
		return ValidationReport{}, fmt.Errorf("benchmark symbol VOO is required in loaded bars")
	}
	windows, err := BuildWalkForwardWindows(len(benchmarkBars), cfg.TrainBars, cfg.TestBars, cfg.StepBars)
	if err != nil {
		return ValidationReport{}, err
	}
	buyHold, err := backtest.RunBuyAndHold(benchmarkBars, cfg.InitialCash)
	if err != nil {
		return ValidationReport{}, err
	}

	candidates := make([]CandidateReport, 0, len(family.VariantFactories))
	foldScores := make(map[string][]float64, len(family.VariantFactories))
	input := BuildInput{BenchmarkSymbol: "VOO", Symbols: cfg.Symbols}
	for _, factory := range family.VariantFactories {
		variant := factory.Build(input)
		candidate, scores, err := runVariant(ctx, variant, barsBySymbol, windows, cfg)
		if err != nil {
			return ValidationReport{}, fmt.Errorf("run variant %s: %w", variant.Name, err)
		}
		candidates = append(candidates, candidate)
		foldScores[variant.Name] = scores
	}

	pbo := EstimatePBO(foldScores, "walk-forward candidate score")
	for i := range candidates {
		candidates[i].PBO = &pbo
		candidates[i].Promotion = EvaluatePromotion(&candidates[i], buyHold, cfg.MinTrades)
	}
	primaryIndex := -1
	for i := range candidates {
		if candidates[i].Strategy == family.PrimaryVariant {
			primaryIndex = i
			break
		}
	}
	if primaryIndex > 0 {
		primary := candidates[primaryIndex]
		copy(candidates[1:primaryIndex+1], candidates[0:primaryIndex])
		candidates[0] = primary
	}

	return ValidationReport{
		GeneratedAt:       time.Now().UTC(),
		StrategyFamily:    family.Name,
		RequestedStrategy: family.PrimaryVariant,
		BenchmarkSymbol:   "VOO",
		Symbols:           append([]string(nil), cfg.Symbols...),
		Timeframe:         cfg.Timeframe,
		Window:            cfg.Window,
		TrainBars:         cfg.TrainBars,
		TestBars:          cfg.TestBars,
		StepBars:          cfg.StepBars,
		MinTrades:         cfg.MinTrades,
		InitialCash:       cfg.InitialCash,
		BuyHold:           buyHold,
		Candidates:        candidates,
	}, nil
}

func runVariant(ctx context.Context, variant VariantConfig, barsBySymbol map[string][]models.Bar, windows []WalkForwardWindow, cfg RunnerConfig) (CandidateReport, []float64, error) {
	strategy, err := NewBenchmarkRankerProxyStrategy(variant, barsBySymbol)
	if err != nil {
		return CandidateReport{}, nil, err
	}
	benchmarkBars := strategy.BenchmarkBars()
	activeBars := strategy.ActiveBars()

	foldResults := make([]FoldResult, 0, len(windows))
	foldScores := make([]float64, 0, len(windows))
	combinedResults := make([]*models.BacktestResult, 0, len(windows))
	for _, window := range windows {
		foldBenchmark := sliceBars(benchmarkBars, 0, window.TestEnd)
		foldActive := sliceBarsMap(activeBars, 0, window.TestEnd)
		foldStrategy, err := NewBenchmarkRankerProxyStrategy(variant, mergeBars(variant.BenchmarkSymbol, foldBenchmark, foldActive))
		if err != nil {
			return CandidateReport{}, nil, err
		}
		plans, err := foldStrategy.GeneratePlans(ctx, window.TrainEnd, window.TestEnd)
		if err != nil {
			return CandidateReport{}, nil, err
		}
		testPlans := append([]AllocationPlan(nil), plans[window.TestStart:window.TestEnd]...)
		testBenchmark := sliceBars(benchmarkBars, window.TestStart, window.TestEnd)
		testActive := sliceBarsMap(activeBars, window.TestStart, window.TestEnd)
		foldResult, err := RunPortfolioBacktest(variant.BenchmarkSymbol, testBenchmark, testActive, testPlans, cfg.InitialCash, variant.TransactionCostBPS)
		if err != nil {
			return CandidateReport{}, nil, err
		}
		combinedResults = append(combinedResults, foldResult)
		buyHold, err := backtest.RunBuyAndHold(testBenchmark, cfg.InitialCash)
		if err != nil {
			return CandidateReport{}, nil, err
		}
		score := scoreResult(foldResult, buyHold)
		foldScores = append(foldScores, score)
		foldResults = append(foldResults, FoldResult{
			Fold:    window.Fold,
			Train:   ValidationWindow{From: benchmarkBars[window.TrainStart].Time, To: benchmarkBars[window.TrainEnd-1].Time},
			Test:    ValidationWindow{From: benchmarkBars[window.TestStart].Time, To: benchmarkBars[window.TestEnd-1].Time},
			Score:   score,
			Result:  foldResult,
			BuyHold: buyHold,
		})
	}

	result := chainBacktestResults(combinedResults, cfg.InitialCash)
	result.Symbol = variant.BenchmarkSymbol
	stresses := make([]CostScenarioResult, 0, 3)
	for _, costBPS := range []float64{variant.TransactionCostBPS, variant.TransactionCostBPS * 2, variant.TransactionCostBPS * 3} {
		stressResults := make([]*models.BacktestResult, 0, len(windows))
		for _, window := range windows {
			foldBenchmark := sliceBars(benchmarkBars, 0, window.TestEnd)
			foldActive := sliceBarsMap(activeBars, 0, window.TestEnd)
			foldStrategy, err := NewBenchmarkRankerProxyStrategy(variant, mergeBars(variant.BenchmarkSymbol, foldBenchmark, foldActive))
			if err != nil {
				return CandidateReport{}, nil, err
			}
			plans, err := foldStrategy.GeneratePlans(ctx, window.TrainEnd, window.TestEnd)
			if err != nil {
				return CandidateReport{}, nil, err
			}
			testPlans := append([]AllocationPlan(nil), plans[window.TestStart:window.TestEnd]...)
			testBenchmark := sliceBars(benchmarkBars, window.TestStart, window.TestEnd)
			testActive := sliceBarsMap(activeBars, window.TestStart, window.TestEnd)
			stress, err := RunPortfolioBacktest(variant.BenchmarkSymbol, testBenchmark, testActive, testPlans, cfg.InitialCash, costBPS)
			if err != nil {
				return CandidateReport{}, nil, err
			}
			stressResults = append(stressResults, stress)
		}
		stressResult := chainBacktestResults(stressResults, cfg.InitialCash)
		stresses = append(stresses, CostScenarioResult{
			Name:               fmt.Sprintf("%.0fbps", costBPS),
			TransactionCostBPS: costBPS,
			FinalEquity:        stressResult.FinalEquity,
			TotalReturn:        stressResult.TotalReturn,
			AnnualizedReturn:   stressResult.AnnualizedReturn,
			Sharpe:             stressResult.Sharpe,
			Sortino:            stressResult.Sortino,
			Calmar:             stressResult.Calmar,
			MaxDrawdown:        stressResult.MaxDrawdown,
			NumTrades:          stressResult.NumTrades,
			Turnover:           stressResult.Turnover,
			FeesPaid:           stressResult.FeesPaid,
		})
	}

	return CandidateReport{
		Strategy:            variant.Name,
		Family:              variant.Family,
		Variant:             variant.Name,
		Description:         variant.Description,
		BenchmarkSymbol:     variant.BenchmarkSymbol,
		ActiveSymbols:       append([]string(nil), variant.ActiveSymbols...),
		LookbackBars:        append([]int(nil), variant.LookbackBars...),
		LookbackWeights:     append([]float64(nil), variant.LookbackWeights...),
		MaxPositions:        variant.MaxPositions,
		RebalanceBars:       variant.RebalanceBars,
		TurnoverBufferRanks: variant.TurnoverBufferRanks,
		ActiveSleevePct:     variant.ActiveSleevePct,
		TransactionCostBPS:  variant.TransactionCostBPS,
		Result:              result,
		CostScenarios:       stresses,
		FoldResults:         foldResults,
	}, foldScores, nil
}

func mergeBars(benchmarkSymbol string, benchmarkBars []models.Bar, activeBars map[string][]models.Bar) map[string][]models.Bar {
	out := make(map[string][]models.Bar, len(activeBars)+1)
	out[benchmarkSymbol] = append([]models.Bar(nil), benchmarkBars...)
	for symbol, bars := range activeBars {
		out[symbol] = append([]models.Bar(nil), bars...)
>>>>>>> 3ea6d428 (Alpha research)
	}
	return out
}

<<<<<<< HEAD
func estimatePBO(
	ctx context.Context,
	panel backtest.AlignedBars,
	variants []StrategyFactory,
	cfg ValidationConfig,
) (float64, bool, []PBODiagnostic, error) {
	if len(variants) < 2 {
		return 1, false, nil, fmt.Errorf("PBO requires at least two variants")
	}
	splits, err := btvalidation.GenerateMLWalkForwardSplits(len(panel.Times), btvalidation.MLWalkForwardConfig{
		TrainSize: cfg.TrainBars,
		TestSize:  cfg.TestBars,
		StepSize:  cfg.StepBars,
	})
	if err != nil {
		return 1, false, nil, err
	}
	trainScores := make([][]float64, 0, len(splits))
	testScores := make([][]float64, 0, len(splits))
	diagnostics := make([]PBODiagnostic, 0, len(splits))
	for _, split := range splits {
		trainRow := make([]float64, 0, len(variants))
		testRow := make([]float64, 0, len(variants))
		nameRow := make([]string, 0, len(variants))
		for _, variant := range variants {
			train := runScenario(ctx, slicePanel(panel, split.TrainStart, split.TrainEnd), variant, cfg.CostScenarios[0], cfg)
			testPanel := slicePanel(panel, split.TrainStart, split.TestEnd)
			test := runScenarioWithMetricStart(ctx, testPanel, variant, cfg.CostScenarios[0], cfg, split.TestStart-split.TrainStart)
			if train.Error != "" || test.Error != "" {
				continue
			}
			trainRow = append(trainRow, selectionMetric(train.Metrics))
			testRow = append(testRow, selectionMetric(test.Metrics))
			nameRow = append(nameRow, variant.Name)
		}
		if len(trainRow) >= 2 && len(trainRow) == len(testRow) {
			trainScores = append(trainScores, trainRow)
			testScores = append(testScores, testRow)
			diagnostics = append(diagnostics, buildPBODiagnostic(split, nameRow, trainRow, testRow))
		}
	}
	if len(trainScores) == 0 {
		return 1, false, diagnostics, fmt.Errorf("PBO could not form valid variant train/test score rows")
	}
	pbo, err := btvalidation.ProbabilityOfBacktestOverfitting(trainScores, testScores)
	return pbo, err == nil, diagnostics, err
}

func selectionMetric(metrics backtest.PortfolioMetrics) float64 {
	if math.IsNaN(metrics.Calmar) || math.IsInf(metrics.Calmar, 0) || metrics.Calmar == 0 {
		return metrics.Sharpe
	}
	return metrics.Calmar
}

func buildPBODiagnostic(
	split btvalidation.WalkForwardSplit,
	names []string,
	trainScores []float64,
	testScores []float64,
) PBODiagnostic {
	winner := maxScoreIndex(trainScores)
	testRank := descendingRank(testScores, winner)
	testWinner := maxScoreIndex(testScores)
	return PBODiagnostic{
		Fold:             split.Fold,
		TrainStart:       split.TrainStart,
		TrainEnd:         split.TrainEnd,
		TestStart:        split.TestStart,
		TestEnd:          split.TestEnd,
		Winner:           names[winner],
		WinnerTrainScore: trainScores[winner],
		WinnerTestScore:  testScores[winner],
		WinnerTestRank:   testRank,
		TestWinner:       names[testWinner],
		TestWinnerScore:  testScores[testWinner],
		VariantCount:     len(trainScores),
		Overfit:          pboOverfitFlag(testRank, len(testScores)),
	}
}

func maxScoreIndex(values []float64) int {
	winner := 0
	best := values[0]
	for i := 1; i < len(values); i++ {
		if values[i] > best {
			winner = i
			best = values[i]
		}
	}
	return winner
}

func descendingRank(values []float64, selected int) int {
	rank := 1
	for i, value := range values {
		if i == selected {
			continue
		}
		if value > values[selected] {
			rank++
		}
	}
	return rank
}

func pboOverfitFlag(rank int, variantCount int) bool {
	if variantCount <= 0 {
		return false
	}
	percentile := 1 - (float64(rank)-0.5)/float64(variantCount)
	if percentile <= 0 || percentile >= 1 {
		return false
	}
	logit := math.Log(percentile / (1 - percentile))
	return logit < 0
}

func slicePanel(panel backtest.AlignedBars, start, end int) backtest.AlignedBars {
	if start < 0 {
		start = 0
	}
	if end > len(panel.Times) {
		end = len(panel.Times)
	}
	if end < start {
		end = start
	}
	out := backtest.AlignedBars{
		Times:      append([]time.Time(nil), panel.Times[start:end]...),
		Symbols:    append([]string(nil), panel.Symbols...),
		Bars:       make(map[string][]models.Bar, len(panel.Bars)),
		Timeframe:  panel.Timeframe,
		Feed:       panel.Feed,
		Adjustment: panel.Adjustment,
		Metadata:   panel.Metadata,
	}
	for symbol, bars := range panel.Bars {
		out.Bars[symbol] = append([]models.Bar(nil), bars[start:end]...)
	}
	return out
}

func validatePanel(panel backtest.AlignedBars) error {
	if len(panel.Times) < 2 {
		return fmt.Errorf("validation requires at least two aligned bars")
	}
	if len(panel.Symbols) == 0 {
		return fmt.Errorf("validation requires at least one symbol")
	}
	for _, symbol := range panel.Symbols {
		if len(panel.Bars[symbol]) != len(panel.Times) {
			return fmt.Errorf("panel symbol %s has %d bars but times has %d", symbol, len(panel.Bars[symbol]), len(panel.Times))
		}
	}
	return nil
}

func auditCandidateMetadata(ctx context.Context, panel backtest.AlignedBars, factory StrategyFactory) map[string]interface{} {
	if factory.New == nil {
		return nil
	}
	strategy := factory.New()
	if strategy == nil {
		return nil
	}
	var fallback map[string]interface{}
	var loadedFallback map[string]interface{}
	for i := range panel.Times {
		output, err := strategy.EvaluatePortfolioLatest(ctx, slicePanel(panel, 0, i+1))
		if err != nil {
			continue
		}
		metadata := compactAuditMetadata(output.EngineMetadata)
		if len(metadata) == 0 {
			continue
		}
		fallback = metadata
		if value, ok := metadata["ranker_model_loaded"].(bool); ok && value {
			loadedFallback = metadata
			if positiveAuditNumber(metadata["candidate_count"]) || positiveAuditNumber(metadata["active_weight"]) {
				return metadata
			}
		}
	}
	if len(loadedFallback) > 0 {
		return loadedFallback
	}
	return fallback
}

func compactAuditMetadata(metadata map[string]interface{}) map[string]interface{} {
	if len(metadata) == 0 {
		return nil
	}
	out := make(map[string]interface{})
	for key, value := range metadata {
		if key == "" || key == "selection_rows" {
			continue
		}
		if compact, ok := compactAuditValue(value); ok {
			out[key] = compact
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func compactAuditValue(value interface{}) (interface{}, bool) {
	switch v := value.(type) {
	case nil:
		return nil, false
	case string:
		if v == "" {
			return nil, false
		}
		return v, true
	case bool:
		return v, true
	case int:
		return v, true
	case int64:
		return v, true
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, false
		}
		return v, true
	case []string:
		if len(v) == 0 {
			return nil, false
		}
		return append([]string(nil), v...), true
	case []interface{}:
		items := make([]interface{}, 0, len(v))
		for _, item := range v {
			compact, ok := compactAuditValue(item)
			if !ok {
				return nil, false
			}
			items = append(items, compact)
		}
		if len(items) == 0 {
			return nil, false
		}
		return items, true
	default:
		return nil, false
	}
}

func positiveAuditNumber(value interface{}) bool {
	switch v := value.(type) {
	case int:
		return v > 0
	case int64:
		return v > 0
	case float64:
		return v > 0 && !math.IsNaN(v) && !math.IsInf(v, 0)
	default:
		return false
	}
}

func finalEquity(equity []models.EquityPoint) float64 {
	if len(equity) == 0 {
		return 0
	}
	return equity[len(equity)-1].Equity
}

func (c ValidationConfig) withDefaults() ValidationConfig {
	defaults := DefaultValidationConfig()
	if c.InitialCash <= 0 {
		c.InitialCash = defaults.InitialCash
	}
	if c.TrainBars <= 0 {
		c.TrainBars = defaults.TrainBars
	}
	if c.TestBars <= 0 {
		c.TestBars = defaults.TestBars
	}
	if c.StepBars <= 0 {
		c.StepBars = defaults.StepBars
	}
	if c.MinOOSTrades <= 0 {
		c.MinOOSTrades = defaults.MinOOSTrades
	}
	if c.NumTrials <= 0 {
		c.NumTrials = defaults.NumTrials
	}
	if c.MaxGrossExposure <= 0 {
		c.MaxGrossExposure = defaults.MaxGrossExposure
	}
	if c.MaxNetExposure <= 0 {
		c.MaxNetExposure = defaults.MaxNetExposure
	}
	if c.MaxSymbolWeight <= 0 {
		c.MaxSymbolWeight = defaults.MaxSymbolWeight
	}
	if len(c.CostScenarios) == 0 {
		c.CostScenarios = defaults.CostScenarios
	}
	if c.PromotionConfig.MinDSR <= 0 {
		c.PromotionConfig = defaults.PromotionConfig
	}
	return c
}
=======
func scoreResult(candidate, buyHold *models.BacktestResult) float64 {
	if candidate == nil {
		return math.Inf(-1)
	}
	score := candidate.Sharpe + candidate.Sortino + candidate.Calmar - candidate.MaxDrawdown
	if buyHold != nil {
		score += (candidate.Sharpe - buyHold.Sharpe) + (candidate.Sortino - buyHold.Sortino) + (candidate.Calmar - buyHold.Calmar)
		score += buyHold.MaxDrawdown - candidate.MaxDrawdown
	}
	return score
}

func normalizeSymbols(symbols []string) []string {
	set := make(map[string]struct{}, len(symbols))
	out := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" {
			continue
		}
		if _, ok := set[symbol]; ok {
			continue
		}
		set[symbol] = struct{}{}
		out = append(out, symbol)
	}
	sort.Strings(out)
	return out
}
>>>>>>> 3ea6d428 (Alpha research)
