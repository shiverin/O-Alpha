package validation

import (
	"fmt"
	"math"

	"github.com/oalpha/internal/backtest"
)

type PromotionConfig struct {
	MinDSR                  float64
	MaxPBO                  float64
	MinOOSTrades            int
	MaxTurnoverIncrease     float64
	MinDrawdownImprovement  float64
	RequireDataQualityPass  bool
	RequireNoLookaheadAudit bool
}

type PromotionDecision struct {
	Promote          bool                      `json:"promote"`
	Reasons          []string                  `json:"reasons"`
	Metrics          backtest.PortfolioMetrics `json:"metrics"`
	BenchmarkMetrics backtest.PortfolioMetrics `json:"benchmark_metrics"`
	DSRPass          bool                      `json:"dsr_pass"`
	PBOPass          bool                      `json:"pbo_pass"`
	CostPass         bool                      `json:"cost_pass"`
	DataQualityPass  bool                      `json:"data_quality_pass"`
	NoLookaheadPass  bool                      `json:"no_lookahead_pass"`
}

func DefaultPromotionConfig() PromotionConfig {
	return PromotionConfig{
		MinDSR:                  0.95,
		MaxPBO:                  0.20,
		MinOOSTrades:            30,
		MaxTurnoverIncrease:     0.15,
		MinDrawdownImprovement:  0.05,
		RequireDataQualityPass:  true,
		RequireNoLookaheadAudit: true,
	}
}

func EvaluatePromotion(
	metrics backtest.PortfolioMetrics,
	benchmark backtest.PortfolioMetrics,
	cfg PromotionConfig,
	dataQualityPass bool,
	noLookaheadPass bool,
) PromotionDecision {
	cfg = cfg.withDefaults()
	decision := PromotionDecision{
		Metrics:          metrics,
		BenchmarkMetrics: benchmark,
		DSRPass:          metrics.DSR >= cfg.MinDSR,
		PBOPass:          metrics.PBO <= cfg.MaxPBO,
		CostPass:         true,
		DataQualityPass:  dataQualityPass,
		NoLookaheadPass:  noLookaheadPass,
	}
	if !decision.DSRPass {
		decision.Reasons = append(decision.Reasons, fmt.Sprintf("DSR %.3f below %.3f", metrics.DSR, cfg.MinDSR))
	}
	if !decision.PBOPass {
		decision.Reasons = append(decision.Reasons, fmt.Sprintf("PBO %.3f above %.3f", metrics.PBO, cfg.MaxPBO))
	}
	if metrics.NumTrades < cfg.MinOOSTrades {
		decision.Reasons = append(decision.Reasons, fmt.Sprintf("OOS trades %d below %d", metrics.NumTrades, cfg.MinOOSTrades))
	}
	if cfg.RequireDataQualityPass && !dataQualityPass {
		decision.Reasons = append(decision.Reasons, "data quality gate failed")
	}
	if cfg.RequireNoLookaheadAudit && !noLookaheadPass {
		decision.Reasons = append(decision.Reasons, "no-lookahead audit failed")
	}
	if turnoverIncrease(metrics.Turnover, benchmark.Turnover) > cfg.MaxTurnoverIncrease &&
		metrics.AnnualReturn <= benchmark.AnnualReturn {
		decision.Reasons = append(decision.Reasons, "turnover increases without return improvement")
	}

	improvesRiskAdjusted := metrics.Sortino > benchmark.Sortino || metrics.Calmar > benchmark.Calmar
	improvesDrawdown := relativeReduction(benchmark.MaxDrawdown, metrics.MaxDrawdown) >= cfg.MinDrawdownImprovement
	preservesReturnWithLowerRisk := metrics.AnnualReturn >= benchmark.AnnualReturn*0.95 && improvesDrawdown
	if !improvesRiskAdjusted && !improvesDrawdown && !preservesReturnWithLowerRisk {
		decision.Reasons = append(decision.Reasons, "no drawdown-adjusted improvement over benchmark")
	}

	decision.Promote = len(decision.Reasons) == 0
	return decision
}

func DeflatedSharpeRatio(observedSharpe float64, n int, skew float64, kurtosisValue float64, numTrials int) float64 {
	benchmarkSharpe := deflatedSharpeBenchmark(numTrials)
	return backtest.ProbabilisticSharpeRatio(observedSharpe, benchmarkSharpe, n, skew, kurtosisValue)
}

func ProbabilityOfBacktestOverfitting(trainScores [][]float64, testScores [][]float64) (float64, error) {
	if len(trainScores) == 0 || len(trainScores) != len(testScores) {
		return 0, fmt.Errorf("train/test score split counts must match and be non-empty")
	}
	var overfit int
	for i := range trainScores {
		if len(trainScores[i]) == 0 || len(trainScores[i]) != len(testScores[i]) {
			return 0, fmt.Errorf("split %d train/test variant counts must match and be non-empty", i)
		}
		winner := argmax(trainScores[i])
		percentile := testRankPercentile(testScores[i], winner)
		if percentile <= 0 || percentile >= 1 {
			continue
		}
		logit := math.Log(percentile / (1 - percentile))
		if logit < 0 {
			overfit++
		}
	}
	return float64(overfit) / float64(len(trainScores)), nil
}

func deflatedSharpeBenchmark(numTrials int) float64 {
	if numTrials <= 1 {
		return 0
	}
	return math.Sqrt(2*math.Log(float64(numTrials))) / math.Sqrt(252)
}

func argmax(values []float64) int {
	winner := 0
	best := values[0]
	for i := 1; i < len(values); i++ {
		if values[i] > best {
			best = values[i]
			winner = i
		}
	}
	return winner
}

func testRankPercentile(values []float64, selected int) float64 {
	rank := 1
	for i, value := range values {
		if i == selected {
			continue
		}
		if value > values[selected] {
			rank++
		}
	}
	return 1 - (float64(rank)-0.5)/float64(len(values))
}

func turnoverIncrease(candidate, benchmark float64) float64 {
	if benchmark <= 0 {
		if candidate <= 0 {
			return 0
		}
		return math.Inf(1)
	}
	return (candidate - benchmark) / benchmark
}

func relativeReduction(before, after float64) float64 {
	if before <= 0 {
		return 0
	}
	return (before - after) / before
}

func (c PromotionConfig) withDefaults() PromotionConfig {
	defaults := DefaultPromotionConfig()
	if c.MinDSR <= 0 {
		c.MinDSR = defaults.MinDSR
	}
	if c.MaxPBO <= 0 {
		c.MaxPBO = defaults.MaxPBO
	}
	if c.MinOOSTrades <= 0 {
		c.MinOOSTrades = defaults.MinOOSTrades
	}
	if c.MaxTurnoverIncrease <= 0 {
		c.MaxTurnoverIncrease = defaults.MaxTurnoverIncrease
	}
	if c.MinDrawdownImprovement <= 0 {
		c.MinDrawdownImprovement = defaults.MinDrawdownImprovement
	}
	return c
}
