package momentum

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

const CompositeMomentumStrategyName = "composite_momentum_sleeve"

type CompositeMomentumLegConfig struct {
	Name                     string
	CandidateUniverse        string
	CandidateSymbols         []string
	RankMode                 string
	WeightMode               string
	LookbackBars             int
	SleeveFraction           float64
	TopK                     int
	MaxNameWeight            float64
	MinRelativeMomentum      float64
	MaxVol20                 float64
	DollarVolumeLookbackBars int
	MinMedianDollarVolume    float64
	EdgeExponent             float64
	VolFloor                 float64
}

type CompositeMomentumConfig struct {
	BenchmarkSymbol            string
	RebalanceEveryBars         int
	GlobalMaxNameWeight        float64
	TurnoverBand               float64
	BenchmarkTrendLookbackBars int
	MinBenchmarkTrend          float64
	RiskOffBenchmarkWeight     float64
	Legs                       []CompositeMomentumLegConfig
	RiskOffLegs                []CompositeMomentumLegConfig
}

type CompositeMomentumStrategy struct {
	cfg                CompositeMomentumConfig
	universe           []string
	lastRebalanceIndex int
	currentTargets     map[string]backtest.TargetPosition
}

type compositeCandidate struct {
	Symbol           string
	Score            float64
	RelativeMomentum float64
	AbsoluteMomentum float64
	Vol20            float64
	Rank             int
}

func DefaultCompositeMomentumConfig() CompositeMomentumConfig {
	return CompositeMomentumConfig{
		BenchmarkSymbol:     "VOO",
		RebalanceEveryBars:  21,
		GlobalMaxNameWeight: 0.30,
		Legs: []CompositeMomentumLegConfig{
			{
				Name:                "etf_21",
				CandidateUniverse:   "etfs",
				LookbackBars:        21,
				SleeveFraction:      0.24,
				TopK:                1,
				MaxNameWeight:       0.24,
				MinRelativeMomentum: 0.05,
				MaxVol20:            0.25,
			},
			{
				Name:                "all_126",
				CandidateUniverse:   "all",
				LookbackBars:        126,
				SleeveFraction:      0.05,
				TopK:                5,
				MaxNameWeight:       0.01,
				MinRelativeMomentum: 0.10,
			},
		},
	}
}

func NewCompositeMomentumStrategy(universe []string, cfg CompositeMomentumConfig) *CompositeMomentumStrategy {
	cfg = cfg.withDefaults()
	return &CompositeMomentumStrategy{
		cfg:                cfg,
		universe:           normalizeSymbols(universe),
		lastRebalanceIndex: -1,
		currentTargets:     make(map[string]backtest.TargetPosition),
	}
}

func (s *CompositeMomentumStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	if s == nil {
		return nil, fmt.Errorf("composite momentum strategy is nil")
	}
	previousIndex := s.lastRebalanceIndex
	previousTargets := cloneTargets(s.currentTargets)
	s.lastRebalanceIndex = -1
	s.currentTargets = make(map[string]backtest.TargetPosition)
	defer func() {
		s.lastRebalanceIndex = previousIndex
		s.currentTargets = previousTargets
	}()

	outputs := make([]backtest.PortfolioOutput, len(panel.Times))
	for i := range panel.Times {
		output, err := s.EvaluatePortfolioLatest(ctx, panelPrefix(panel, i+1))
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}
	return outputs, nil
}

func (s *CompositeMomentumStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	_ = ctx
	if s == nil {
		return backtest.PortfolioOutput{}, fmt.Errorf("composite momentum strategy is nil")
	}
	if len(panel.Times) == 0 {
		return backtest.PortfolioOutput{}, fmt.Errorf("aligned panel has no timestamps")
	}
	index := len(panel.Times) - 1
	t := panel.Times[index]
	if !s.hasBenchmark(panel, index) {
		return backtest.PortfolioOutput{
			Time:       t,
			Targets:    map[string]backtest.TargetPosition{},
			CashWeight: 1,
			EngineMetadata: map[string]interface{}{
				"engine": CompositeMomentumStrategyName,
				"reason": "missing_benchmark_bars",
			},
		}, nil
	}
	if !s.shouldRebalance(index) {
		return backtest.PortfolioOutput{
			Time:          t,
			Targets:       map[string]backtest.TargetPosition{},
			GrossExposure: grossExposure(s.currentTargets),
			NetExposure:   netExposure(s.currentTargets),
			CashWeight:    math.Max(0, 1-grossExposure(s.currentTargets)),
			EngineMetadata: map[string]interface{}{
				"engine":    CompositeMomentumStrategyName,
				"rebalance": false,
				"action":    actionHoldTargets,
			},
		}, nil
	}

	targets, metadata := s.buildTargets(panel, index)
	s.lastRebalanceIndex = index
	s.currentTargets = cloneTargets(targets)
	return backtest.PortfolioOutput{
		Time:           t,
		Targets:        targets,
		GrossExposure:  grossExposure(targets),
		NetExposure:    netExposure(targets),
		CashWeight:     math.Max(0, 1-grossExposure(targets)),
		EngineMetadata: metadata,
	}, nil
}

func (s *CompositeMomentumStrategy) Universe() []string {
	if s == nil {
		return nil
	}
	return append([]string(nil), s.universe...)
}

func (s *CompositeMomentumStrategy) Name() string {
	return CompositeMomentumStrategyName
}

func (s *CompositeMomentumStrategy) shouldRebalance(index int) bool {
	if s.lastRebalanceIndex < 0 {
		return true
	}
	if s.cfg.RebalanceEveryBars <= 0 {
		return false
	}
	return index-s.lastRebalanceIndex >= s.cfg.RebalanceEveryBars
}

func (s *CompositeMomentumStrategy) hasBenchmark(panel backtest.AlignedBars, index int) bool {
	bars := panel.Bars[s.cfg.BenchmarkSymbol]
	return len(bars) > index && bars[index].Close > 0
}

func (s *CompositeMomentumStrategy) buildTargets(panel backtest.AlignedBars, index int) (map[string]backtest.TargetPosition, map[string]interface{}) {
	riskOff, benchmarkTrend := s.isRiskOff(panel, index)
	legs := s.cfg.Legs
	benchmarkWeightOverride := math.NaN()
	if riskOff && len(s.cfg.RiskOffLegs) > 0 {
		legs = s.cfg.RiskOffLegs
		benchmarkWeightOverride = s.cfg.RiskOffBenchmarkWeight
	}

	activeWeights := make(map[string]float64)
	selectionRows := make([]map[string]interface{}, 0)
	for _, leg := range legs {
		selected := s.selectLeg(panel, index, leg)
		legWeights := legTargetWeights(selected, leg)
		for _, candidate := range selected {
			weight := legWeights[candidate.Symbol]
			if weight <= 0 {
				continue
			}
			activeWeights[candidate.Symbol] += weight
			selectionRows = append(selectionRows, map[string]interface{}{
				"leg":               leg.Name,
				"rank":              candidate.Rank,
				"symbol":            candidate.Symbol,
				"score":             candidate.Score,
				"relative_momentum": candidate.RelativeMomentum,
				"absolute_momentum": candidate.AbsoluteMomentum,
				"vol_20":            candidate.Vol20,
				"target_weight":     weight,
			})
		}
	}

	activeTotal := capActiveWeights(activeWeights, s.cfg.GlobalMaxNameWeight)
	if !math.IsNaN(benchmarkWeightOverride) && activeTotal+benchmarkWeightOverride > 1 {
		scale := (1 - benchmarkWeightOverride) / activeTotal
		for symbol, weight := range activeWeights {
			activeWeights[symbol] = weight * scale
		}
		activeTotal = capActiveWeights(activeWeights, s.cfg.GlobalMaxNameWeight)
	}

	targets := make(map[string]backtest.TargetPosition)
	for symbol, weight := range activeWeights {
		if weight <= 0 {
			continue
		}
		targets[symbol] = backtest.TargetPosition{
			Symbol:       symbol,
			TargetWeight: weight,
			AlphaScore:   weight,
			Confidence:   1,
			Side:         backtest.PositionSideLong,
			Engine:       CompositeMomentumStrategyName,
			Metadata: map[string]interface{}{
				"rebalance": true,
			},
		}
	}
	benchmarkWeight := math.Max(0, 1-activeTotal)
	if !math.IsNaN(benchmarkWeightOverride) {
		benchmarkWeight = benchmarkWeightOverride
	}
	if benchmarkWeight > 0 {
		targets[s.cfg.BenchmarkSymbol] = backtest.TargetPosition{
			Symbol:       s.cfg.BenchmarkSymbol,
			TargetWeight: benchmarkWeight,
			AlphaScore:   1,
			Confidence:   1,
			Side:         backtest.PositionSideLong,
			Engine:       CompositeMomentumStrategyName,
			Metadata: map[string]interface{}{
				"role":      "benchmark_core",
				"rebalance": true,
			},
		}
	}
	if len(s.currentTargets) > 0 && s.cfg.TurnoverBand > 0 {
		turnover := targetTurnover(s.currentTargets, targets)
		if turnover < s.cfg.TurnoverBand {
			held := cloneTargets(s.currentTargets)
			return held, map[string]interface{}{
				"engine":           CompositeMomentumStrategyName,
				"rebalance":        true,
				"action":           actionHoldTargets,
				"active_weight":    activeWeight(held, s.cfg.BenchmarkSymbol),
				"benchmark":        s.cfg.BenchmarkSymbol,
				"benchmark_weight": targetWeight(held, s.cfg.BenchmarkSymbol),
				"benchmark_trend":  benchmarkTrend,
				"risk_off":         riskOff,
				"selection_rows":   selectionRows,
				"target_count":     len(held),
				"turnover":         turnover,
				"turnover_band":    s.cfg.TurnoverBand,
			}
		}
	}
	return targets, map[string]interface{}{
		"engine":           CompositeMomentumStrategyName,
		"rebalance":        true,
		"active_weight":    activeTotal,
		"benchmark":        s.cfg.BenchmarkSymbol,
		"benchmark_weight": benchmarkWeight,
		"benchmark_trend":  benchmarkTrend,
		"risk_off":         riskOff,
		"selection_rows":   selectionRows,
		"target_count":     len(targets),
		"turnover":         targetTurnover(s.currentTargets, targets),
		"turnover_band":    s.cfg.TurnoverBand,
	}
}

func (s *CompositeMomentumStrategy) isRiskOff(panel backtest.AlignedBars, index int) (bool, float64) {
	if s.cfg.BenchmarkTrendLookbackBars <= 0 {
		return false, 0
	}
	trend, ok := logReturnAt(panel.Bars[s.cfg.BenchmarkSymbol], index, s.cfg.BenchmarkTrendLookbackBars)
	if !ok {
		return false, 0
	}
	return trend < s.cfg.MinBenchmarkTrend, trend
}

func (s *CompositeMomentumStrategy) selectLeg(panel backtest.AlignedBars, index int, leg CompositeMomentumLegConfig) []compositeCandidate {
	if leg.LookbackBars <= 0 || index-leg.LookbackBars < 0 {
		return nil
	}
	benchmarkBars := panel.Bars[s.cfg.BenchmarkSymbol]
	benchmarkMomentum, ok := logReturnAt(benchmarkBars, index, leg.LookbackBars)
	if !ok {
		return nil
	}
	candidates := make([]compositeCandidate, 0)
	for _, symbol := range s.candidateUniverse(panel, leg) {
		if symbol == s.cfg.BenchmarkSymbol || isBenchmarkProxy(symbol) {
			continue
		}
		bars := panel.Bars[symbol]
		absoluteMomentum, ok := logReturnAt(bars, index, leg.LookbackBars)
		if !ok {
			continue
		}
		if leg.MinMedianDollarVolume > 0 {
			liquidityLookback := leg.DollarVolumeLookbackBars
			if liquidityLookback <= 0 {
				liquidityLookback = 63
			}
			start := maxInt(0, index-liquidityLookback+1)
			if medianDollarVolume(bars, start, index) < leg.MinMedianDollarVolume {
				continue
			}
		}
		vol20 := annualizedStd(logReturnsBetween(bars, maxInt(1, index-20+1), index))
		if leg.MaxVol20 > 0 && vol20 > leg.MaxVol20 {
			continue
		}
		relativeMomentum := absoluteMomentum - benchmarkMomentum
		if relativeMomentum < leg.MinRelativeMomentum {
			continue
		}
		candidates = append(candidates, compositeCandidate{
			Symbol:           symbol,
			Score:            legScore(leg, relativeMomentum, vol20),
			RelativeMomentum: relativeMomentum,
			AbsoluteMomentum: absoluteMomentum,
			Vol20:            vol20,
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Symbol < candidates[j].Symbol
		}
		return candidates[i].Score > candidates[j].Score
	})
	limit := leg.TopK
	if limit <= 0 {
		limit = 1
	}
	if limit > len(candidates) {
		limit = len(candidates)
	}
	for i := 0; i < limit; i++ {
		candidates[i].Rank = i + 1
	}
	return append([]compositeCandidate(nil), candidates[:limit]...)
}

func (s *CompositeMomentumStrategy) candidateUniverse(panel backtest.AlignedBars, leg CompositeMomentumLegConfig) []string {
	if len(panel.Symbols) == 0 {
		return nil
	}
	symbols := leg.CandidateSymbols
	if len(symbols) == 0 {
		symbols = s.universe
	}
	if len(symbols) == 0 {
		symbols = panel.Symbols
	}
	out := make([]string, 0, len(symbols))
	for _, symbol := range normalizeSymbols(symbols) {
		switch strings.ToLower(leg.CandidateUniverse) {
		case "etfs":
			if !isETFSymbol(symbol) {
				continue
			}
		case "stocks":
			if isETFSymbol(symbol) {
				continue
			}
		}
		out = append(out, symbol)
	}
	return out
}

func candidateSymbolOverride(symbols []string) []string {
	normalized := normalizeSymbols(symbols)
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func (c CompositeMomentumConfig) withDefaults() CompositeMomentumConfig {
	defaults := DefaultCompositeMomentumConfig()
	c.BenchmarkSymbol = strings.ToUpper(strings.TrimSpace(c.BenchmarkSymbol))
	if c.BenchmarkSymbol == "" {
		c.BenchmarkSymbol = defaults.BenchmarkSymbol
	}
	if c.RebalanceEveryBars <= 0 {
		c.RebalanceEveryBars = defaults.RebalanceEveryBars
	}
	if c.GlobalMaxNameWeight <= 0 {
		c.GlobalMaxNameWeight = defaults.GlobalMaxNameWeight
	}
	if c.GlobalMaxNameWeight > 1 {
		c.GlobalMaxNameWeight = 1
	}
	if c.TurnoverBand < 0 {
		c.TurnoverBand = 0
	}
	if c.TurnoverBand > 1 {
		c.TurnoverBand = 1
	}
	if len(c.Legs) == 0 {
		c.Legs = defaults.Legs
	}
	for i := range c.Legs {
		c.Legs[i] = c.Legs[i].withDefaults(i)
	}
	if c.RiskOffBenchmarkWeight < 0 {
		c.RiskOffBenchmarkWeight = 0
	}
	if c.RiskOffBenchmarkWeight > 1 {
		c.RiskOffBenchmarkWeight = 1
	}
	for i := range c.RiskOffLegs {
		c.RiskOffLegs[i] = c.RiskOffLegs[i].withDefaults(i)
	}
	return c
}

func (c CompositeMomentumLegConfig) withDefaults(index int) CompositeMomentumLegConfig {
	if c.Name == "" {
		c.Name = fmt.Sprintf("leg_%d", index+1)
	}
	c.CandidateSymbols = candidateSymbolOverride(c.CandidateSymbols)
	c.RankMode = strings.ToLower(strings.TrimSpace(c.RankMode))
	if c.RankMode == "" {
		c.RankMode = "relative_momentum"
	}
	c.WeightMode = strings.ToLower(strings.TrimSpace(c.WeightMode))
	if c.WeightMode == "" {
		c.WeightMode = "equal"
	}
	c.CandidateUniverse = strings.ToLower(strings.TrimSpace(c.CandidateUniverse))
	if c.CandidateUniverse == "" {
		c.CandidateUniverse = "all"
	}
	if c.LookbackBars <= 0 {
		c.LookbackBars = 21
	}
	if c.SleeveFraction < 0 {
		c.SleeveFraction = 0
	}
	if c.SleeveFraction > 1 {
		c.SleeveFraction = 1
	}
	if c.TopK <= 0 {
		c.TopK = 1
	}
	if c.MaxNameWeight <= 0 {
		c.MaxNameWeight = c.SleeveFraction / float64(c.TopK)
	}
	if c.MaxNameWeight > 1 {
		c.MaxNameWeight = 1
	}
	if c.EdgeExponent <= 0 {
		c.EdgeExponent = 1
	}
	if c.VolFloor <= 0 {
		c.VolFloor = 1e-6
	}
	return c
}

func legScore(leg CompositeMomentumLegConfig, relativeMomentum float64, vol20 float64) float64 {
	switch strings.ToLower(leg.RankMode) {
	case "low_vol", "low_volatility":
		return -vol20
	case "mean_reversion", "relative_reversal":
		return -relativeMomentum
	case "vol_adjusted_momentum":
		if vol20 <= 1e-9 {
			return relativeMomentum
		}
		return relativeMomentum / vol20
	default:
		return relativeMomentum
	}
}

func capActiveWeights(weights map[string]float64, maxNameWeight float64) float64 {
	var total float64
	for symbol, weight := range weights {
		if maxNameWeight > 0 {
			weight = math.Min(weight, maxNameWeight)
		}
		if weight < 0 {
			weight = 0
		}
		weights[symbol] = weight
		total += weight
	}
	return total
}

func legTargetWeights(selected []compositeCandidate, leg CompositeMomentumLegConfig) map[string]float64 {
	switch strings.ToLower(leg.WeightMode) {
	case "score", "score_weighted", "risk_adjusted_edge", "edge_over_vol":
		return scoreLegWeights(selected, leg)
	default:
		return equalLegWeights(selected, leg)
	}
}

func equalLegWeights(selected []compositeCandidate, leg CompositeMomentumLegConfig) map[string]float64 {
	out := make(map[string]float64, len(selected))
	if len(selected) == 0 || leg.SleeveFraction <= 0 {
		return out
	}
	weight := leg.SleeveFraction / float64(len(selected))
	if leg.MaxNameWeight > 0 {
		weight = math.Min(weight, leg.MaxNameWeight)
	}
	for _, candidate := range selected {
		out[candidate.Symbol] = weight
	}
	return out
}

func scoreLegWeights(selected []compositeCandidate, leg CompositeMomentumLegConfig) map[string]float64 {
	out := make(map[string]float64, len(selected))
	if len(selected) == 0 || leg.SleeveFraction <= 0 {
		return out
	}
	raw := make(map[string]float64, len(selected))
	var total float64
	for _, candidate := range selected {
		edge := candidate.RelativeMomentum - leg.MinRelativeMomentum
		if strings.Contains(leg.WeightMode, "score") {
			edge = math.Max(0, candidate.Score)
		}
		if edge <= 0 {
			continue
		}
		conviction := math.Pow(edge, leg.EdgeExponent)
		denominator := math.Max(candidate.Vol20, leg.VolFloor)
		weight := conviction / denominator
		if weight <= 0 || math.IsNaN(weight) || math.IsInf(weight, 0) {
			continue
		}
		raw[candidate.Symbol] = weight
		total += weight
	}
	if total <= 0 {
		return out
	}
	for symbol, value := range raw {
		weight := leg.SleeveFraction * value / total
		if leg.MaxNameWeight > 0 {
			weight = math.Min(weight, leg.MaxNameWeight)
		}
		if weight > 0 {
			out[symbol] = weight
		}
	}
	return out
}

func targetTurnover(previous map[string]backtest.TargetPosition, next map[string]backtest.TargetPosition) float64 {
	if len(previous) == 0 && len(next) == 0 {
		return 0
	}
	seen := make(map[string]struct{}, len(previous)+len(next))
	var turnover float64
	for symbol, target := range previous {
		seen[symbol] = struct{}{}
		turnover += math.Abs(target.TargetWeight - targetWeight(next, symbol))
	}
	for symbol, target := range next {
		if _, ok := seen[symbol]; ok {
			continue
		}
		turnover += math.Abs(target.TargetWeight)
	}
	return turnover
}

func targetWeight(targets map[string]backtest.TargetPosition, symbol string) float64 {
	if targets == nil {
		return 0
	}
	return targets[strings.ToUpper(strings.TrimSpace(symbol))].TargetWeight
}

func activeWeight(targets map[string]backtest.TargetPosition, benchmark string) float64 {
	benchmark = strings.ToUpper(strings.TrimSpace(benchmark))
	var total float64
	for symbol, target := range targets {
		if symbol == benchmark {
			continue
		}
		if target.TargetWeight > 0 {
			total += target.TargetWeight
		}
	}
	return total
}

func logReturnAt(bars []models.Bar, index, lookback int) (float64, bool) {
	if index-lookback < 0 || index >= len(bars) {
		return 0, false
	}
	start := bars[index-lookback].Close
	end := bars[index].Close
	if start <= 0 || end <= 0 {
		return 0, false
	}
	return math.Log(end / start), true
}

func isBenchmarkProxy(symbol string) bool {
	switch strings.ToUpper(symbol) {
	case "VOO", "SPY":
		return true
	default:
		return false
	}
}

func isETFSymbol(symbol string) bool {
	switch strings.ToUpper(symbol) {
	case "DIA", "IWM", "QQQ", "SMH", "SPY", "VTI", "VOO", "XLB", "XLC", "XLE", "XLF", "XLI", "XLK", "XLP", "XLRE", "XLU", "XLV", "XLY":
		return true
	default:
		return false
	}
}
