package evaluation

import (
	"context"
	"fmt"
	"math"

	"github.com/oalpha/internal/agent"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

type RegimeComparisonConfig struct {
	WindowSize  int
	TrainBars   int
	TestBars    int
	StepBars    int
	InitialCash float64
	RiskProfile agent.RiskProfile
	MinTrades   int
	Symbol      string
	Timeframe   string
}

type RegimeComparisonReport struct {
	Symbol           string
	Timeframe        string
	FoldCount        int
	Results          map[agent.RegimeMode]RegimeComparisonResult
	BuyAndHold       *models.BacktestResult
	WorkerResults    map[WorkerParityMode]WorkerParityResult
	WorkerBuyAndHold *models.BacktestResult
	Promotion        RegimePromotionDecision
}

type RegimeComparisonResult struct {
	Mode               agent.RegimeMode
	Backtest           *models.BacktestResult
	RegimeDistribution map[string]int
	RejectedModels     int
	RejectReasons      map[string]int
}

type RegimePromotionDecision struct {
	PromoteOverlay      bool
	Reason              string
	DrawdownImprovement float64
	CalmarImproved      bool
	SortinoImproved     bool
	SharpeDeterioration float64
	ReturnDeterioration float64
	TurnoverIncrease    float64
	TradeCountOK        bool
}

func DefaultRegimeComparisonConfig() RegimeComparisonConfig {
	return RegimeComparisonConfig{
		WindowSize:  50,
		TrainBars:   6 * 30,
		TestBars:    30,
		StepBars:    30,
		InitialCash: 100_000,
		RiskProfile: agent.RiskProfileModerate,
		MinTrades:   10,
	}
}

func RunWalkForwardRegimeComparison(
	ctx context.Context,
	bars []models.Bar,
	modes []agent.RegimeMode,
	config RegimeComparisonConfig,
) (RegimeComparisonReport, error) {
	config = config.withDefaults()
	if len(modes) == 0 {
		modes = []agent.RegimeMode{agent.RegimeModeNone, agent.RegimeModeOverlay}
	}
	if len(bars) < config.TrainBars+config.TestBars {
		return RegimeComparisonReport{}, fmt.Errorf("insufficient bars for walk-forward comparison: have %d, need %d", len(bars), config.TrainBars+config.TestBars)
	}

	accumulators := make(map[agent.RegimeMode]*regimeModeAccumulator, len(modes))
	for _, mode := range modes {
		accumulators[mode] = newRegimeModeAccumulator(mode)
	}

	fold := 0
	var outOfSampleBars []models.Bar
	for start := 0; start+config.TrainBars+config.TestBars <= len(bars); start += config.StepBars {
		trainEnd := start + config.TrainBars
		testEnd := trainEnd + config.TestBars
		train := bars[start:trainEnd]
		test := bars[trainEnd:testEnd]
		history := bars[start:testEnd]
		outOfSampleBars = append(outOfSampleBars, test...)

		for _, mode := range modes {
			strategy, err := buildRegimeStrategyForFold(mode, train, config)
			if err != nil {
				return RegimeComparisonReport{}, err
			}

			outputs, err := strategy.GenerateSignals(ctx, history)
			if err != nil {
				return RegimeComparisonReport{}, fmt.Errorf("generate %s fold %d signals: %w", mode, fold, err)
			}
			if len(outputs) != len(history) {
				return RegimeComparisonReport{}, fmt.Errorf("%s fold %d produced %d outputs for %d bars", mode, fold, len(outputs), len(history))
			}

			testOutputs := outputs[len(train):]
			acc := accumulators[mode]
			acc.bars = append(acc.bars, test...)
			acc.outputs = append(acc.outputs, testOutputs...)
			for _, output := range testOutputs {
				label := output.RegimeLabel
				if label == "" {
					label = "UNKNOWN"
				}
				acc.regimeDistribution[label]++
			}
		}
		fold++
	}

	results := make(map[agent.RegimeMode]RegimeComparisonResult, len(modes))
	for _, mode := range modes {
		acc := accumulators[mode]
		result, err := backtest.RunBacktestWithOutputs(acc.bars, acc.outputs, config.InitialCash)
		if err != nil {
			return RegimeComparisonReport{}, fmt.Errorf("backtest %s comparison outputs: %w", mode, err)
		}
		results[mode] = RegimeComparisonResult{
			Mode:               mode,
			Backtest:           result,
			RegimeDistribution: acc.regimeDistribution,
			RejectReasons:      make(map[string]int),
		}
	}

	buyAndHold, err := backtest.RunBuyAndHold(outOfSampleBars, config.InitialCash)
	if err != nil {
		return RegimeComparisonReport{}, fmt.Errorf("backtest buy-and-hold baseline: %w", err)
	}

	report := RegimeComparisonReport{
		Symbol:     config.Symbol,
		Timeframe:  config.Timeframe,
		FoldCount:  fold,
		Results:    results,
		BuyAndHold: buyAndHold,
		Promotion:  evaluateOverlayPromotion(results, config),
	}
	return report, nil
}

type regimeModeAccumulator struct {
	mode               agent.RegimeMode
	bars               []models.Bar
	outputs            []backtest.StrategyOutput
	regimeDistribution map[string]int
}

func newRegimeModeAccumulator(mode agent.RegimeMode) *regimeModeAccumulator {
	return &regimeModeAccumulator{
		mode:               mode,
		regimeDistribution: make(map[string]int),
	}
}

func buildRegimeStrategyForFold(
	mode agent.RegimeMode,
	train []models.Bar,
	config RegimeComparisonConfig,
) (*agent.EnsembleDecisionLayer, error) {
	switch mode {
	case agent.RegimeModeNone:
		return agent.NewEnsembleDecisionLayerForMode(nil, nil, config.WindowSize, config.RiskProfile, agent.RegimeModeNone)
	case "", agent.RegimeModeOverlay:
		strategy := agent.NewEnsembleDecisionLayer(nil, nil, config.WindowSize, config.RiskProfile)
		strategy.Calibrate(train)
		return strategy, nil
	default:
		return nil, fmt.Errorf("unsupported regime mode: %s", mode)
	}
}

func evaluateOverlayPromotion(results map[agent.RegimeMode]RegimeComparisonResult, config RegimeComparisonConfig) RegimePromotionDecision {
	none, hasNone := results[agent.RegimeModeNone]
	overlay, hasOverlay := results[agent.RegimeModeOverlay]
	if !hasNone || !hasOverlay {
		return RegimePromotionDecision{Reason: "promotion requires none and overlay results"}
	}
	if none.Backtest == nil || overlay.Backtest == nil {
		return RegimePromotionDecision{Reason: "promotion requires completed backtests"}
	}

	drawdownImprovement := relativeReduction(none.Backtest.MaxDrawdown, overlay.Backtest.MaxDrawdown)
	calmarImproved := overlay.Backtest.Calmar > none.Backtest.Calmar
	sortinoImproved := overlay.Backtest.Sortino > none.Backtest.Sortino
	sharpeDeterioration := relativeDeterioration(none.Backtest.Sharpe, overlay.Backtest.Sharpe)
	returnDeterioration := relativeDeterioration(none.Backtest.TotalReturn, overlay.Backtest.TotalReturn)
	turnoverIncrease := relativeIncrease(overlay.Backtest.Turnover, none.Backtest.Turnover)
	tradeCountOK := overlay.Backtest.NumTrades >= config.MinTrades

	decision := RegimePromotionDecision{
		DrawdownImprovement: drawdownImprovement,
		CalmarImproved:      calmarImproved,
		SortinoImproved:     sortinoImproved,
		SharpeDeterioration: sharpeDeterioration,
		ReturnDeterioration: returnDeterioration,
		TurnoverIncrease:    turnoverIncrease,
		TradeCountOK:        tradeCountOK,
	}

	switch {
	case drawdownImprovement < 0.05:
		decision.Reason = "max drawdown improvement is below 5 percent"
	case !calmarImproved && !sortinoImproved:
		decision.Reason = "neither Calmar nor Sortino improved versus no-HMM"
	case sharpeDeterioration > 0.05:
		decision.Reason = "Sharpe deteriorated by more than 5 percent"
	case returnDeterioration > 0.10:
		decision.Reason = "total return deteriorated by more than 10 percent"
	case turnoverIncrease > 0.15:
		decision.Reason = "turnover increased by more than 15 percent"
	case !tradeCountOK:
		decision.Reason = "trade count is below the configured minimum"
	default:
		decision.PromoteOverlay = true
		decision.Reason = "overlay passed risk-first promotion gates"
	}
	return decision
}

func (c RegimeComparisonConfig) withDefaults() RegimeComparisonConfig {
	defaults := DefaultRegimeComparisonConfig()
	if c == (RegimeComparisonConfig{}) {
		return defaults
	}
	if c.WindowSize <= 0 {
		c.WindowSize = defaults.WindowSize
	}
	if c.TrainBars <= 0 {
		c.TrainBars = defaults.TrainBars
	}
	if c.TestBars <= 0 {
		c.TestBars = defaults.TestBars
	}
	if c.StepBars <= 0 {
		c.StepBars = c.TestBars
	}
	if c.InitialCash <= 0 {
		c.InitialCash = defaults.InitialCash
	}
	if c.RiskProfile < agent.RiskProfileConservative || c.RiskProfile > agent.RiskProfileAggressive {
		c.RiskProfile = defaults.RiskProfile
	}
	if c.MinTrades <= 0 {
		c.MinTrades = defaults.MinTrades
	}
	return c
}

func relativeReduction(baseline, candidate float64) float64 {
	if baseline <= 0 {
		if candidate <= 0 {
			return 0
		}
		return math.Inf(-1)
	}
	return (baseline - candidate) / baseline
}

func relativeDeterioration(baseline, candidate float64) float64 {
	if candidate >= baseline {
		return 0
	}
	denom := math.Abs(baseline)
	if denom < 1e-9 {
		return math.Inf(1)
	}
	return (baseline - candidate) / denom
}

func relativeIncrease(candidate, baseline float64) float64 {
	if baseline <= 0 {
		if candidate <= 0 {
			return 0
		}
		return math.Inf(1)
	}
	return (candidate - baseline) / baseline
}
