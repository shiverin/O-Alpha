package ranker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/ml"
	"github.com/oalpha/pkg/models"
)

const DailyRankerSleeveStrategyName = "daily_lgbm_ranker_sleeve"

type DailyRankerSleeveConfig struct {
	BenchmarkSymbol         string
	CandidateUniverse       string
	ExcludedSymbols         []string
	PointInTimeUniverse     *PointInTimeUniverse
	PointInTimeUniversePath string
	ModelArtifactRoot       string
	ModelVariant            string
	ModelPathsByYear        map[int]string
	RebalanceEveryBars      int
	SleeveFraction          float64
	TopK                    int
	MaxNameWeight           float64
	TurnoverBand            float64
	MinScoreZ               float64
	MaxCandidateVol         float64
	MaxBenchmarkVol         float64
	HighVolScale            float64
	MaxBenchmarkDrawdown    float64
	DrawdownScale           float64
}

type DailyRankerSleeveStrategy struct {
	cfg                DailyRankerSleeveConfig
	universe           []string
	builder            *ml.DailyRankerFeatureBuilder
	predictors         map[int]*ml.LeavesPredictor
	modelMetadata      map[int]map[string]interface{}
	pitUniverse        *PointInTimeUniverse
	pitUniversePath    string
	pitUniverseError   error
	lastRebalanceIndex int
	currentTargets     map[string]backtest.TargetPosition
	lastModelAudit     map[string]interface{}
}

type rankerCandidate struct {
	Symbol string
	Score  float64
	ScoreZ float64
	Vol20  float64
	Rank   int
}

func NewDailyRankerSleeveStrategy(universe []string, cfg DailyRankerSleeveConfig) *DailyRankerSleeveStrategy {
	cfg = cfg.withDefaults()
	pitUniverse := cfg.PointInTimeUniverse
	var pitError error
	if pitUniverse == nil && strings.TrimSpace(cfg.PointInTimeUniversePath) != "" {
		pitUniverse, pitError = LoadPointInTimeUniverse(cfg.PointInTimeUniversePath)
	}
	return &DailyRankerSleeveStrategy{
		cfg:                cfg,
		universe:           normalizeSymbols(universe),
		builder:            ml.NewDailyRankerFeatureBuilder(cfg.BenchmarkSymbol),
		predictors:         make(map[int]*ml.LeavesPredictor),
		modelMetadata:      make(map[int]map[string]interface{}),
		pitUniverse:        pitUniverse,
		pitUniversePath:    strings.TrimSpace(cfg.PointInTimeUniversePath),
		pitUniverseError:   pitError,
		lastRebalanceIndex: -1,
		currentTargets:     make(map[string]backtest.TargetPosition),
	}
}

func DailyRankerModelPaths(root, variant string, years ...int) map[int]string {
	out := make(map[int]string, len(years))
	for _, year := range years {
		out[year] = filepath.Join(root, variant, fmt.Sprintf("%d", year), "model.txt")
	}
	return out
}

func (s *DailyRankerSleeveStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	if s == nil {
		return nil, fmt.Errorf("daily ranker strategy is nil")
	}
	previousIndex := s.lastRebalanceIndex
	previousTargets := cloneTargets(s.currentTargets)
	previousModelAudit := cloneInterfaceMap(s.lastModelAudit)
	s.lastRebalanceIndex = -1
	s.currentTargets = make(map[string]backtest.TargetPosition)
	s.lastModelAudit = nil
	defer func() {
		s.lastRebalanceIndex = previousIndex
		s.currentTargets = previousTargets
		s.lastModelAudit = previousModelAudit
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

func (s *DailyRankerSleeveStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	_ = ctx
	if s == nil {
		return backtest.PortfolioOutput{}, fmt.Errorf("daily ranker strategy is nil")
	}
	if len(panel.Times) == 0 {
		return backtest.PortfolioOutput{}, fmt.Errorf("aligned panel has no timestamps")
	}
	if s.pitUniverseError != nil {
		return backtest.PortfolioOutput{}, s.pitUniverseError
	}
	index := len(panel.Times) - 1
	t := panel.Times[index]
	if !s.hasBenchmark(panel, index) {
		return s.benchmarkOnlyOutput(t, "missing_benchmark_bars"), nil
	}
	if !s.shouldRebalance(index) {
		metadata := map[string]interface{}{
			"engine":    DailyRankerSleeveStrategyName,
			"rebalance": false,
			"action":    "hold_targets",
		}
		mergeMetadata(metadata, s.lastModelAudit)
		return backtest.PortfolioOutput{
			Time:           t,
			Targets:        map[string]backtest.TargetPosition{},
			GrossExposure:  grossExposure(s.currentTargets),
			NetExposure:    grossExposure(s.currentTargets),
			CashWeight:     math.Max(0, 1-grossExposure(s.currentTargets)),
			EngineMetadata: metadata,
		}, nil
	}

	predictor, ok, err := s.predictorForYear(t.Year())
	if err != nil {
		return backtest.PortfolioOutput{}, err
	}
	if !ok {
		output := s.rebalanceToBenchmark(t, index, "missing_year_model")
		audit := s.modelAuditMetadata(t.Year())
		mergeMetadata(output.EngineMetadata, audit)
		s.lastModelAudit = audit
		return output, nil
	}

	targets, metadata, err := s.buildTargets(panel, index, predictor)
	if err != nil {
		return backtest.PortfolioOutput{}, err
	}
	audit := s.modelAuditMetadata(t.Year())
	mergeMetadata(metadata, audit)
	s.lastRebalanceIndex = index
	s.currentTargets = cloneTargets(targets)
	s.lastModelAudit = audit
	return backtest.PortfolioOutput{
		Time:           t,
		Targets:        targets,
		GrossExposure:  grossExposure(targets),
		NetExposure:    grossExposure(targets),
		CashWeight:     math.Max(0, 1-grossExposure(targets)),
		EngineMetadata: metadata,
	}, nil
}

func (s *DailyRankerSleeveStrategy) Universe() []string {
	if s == nil {
		return nil
	}
	return append([]string(nil), s.universe...)
}

func (s *DailyRankerSleeveStrategy) Name() string {
	return DailyRankerSleeveStrategyName
}

func (s *DailyRankerSleeveStrategy) predictorForYear(year int) (*ml.LeavesPredictor, bool, error) {
	modelPath := strings.TrimSpace(s.cfg.ModelPathsByYear[year])
	if modelPath == "" {
		return nil, false, nil
	}
	if predictor, ok := s.predictors[year]; ok {
		return predictor, true, nil
	}
	predictor, err := ml.NewRawLeavesPredictor(modelPath, ml.FeatureSpec{
		Version:  "daily_ranker_v1",
		Features: ml.DailyRankerFeatureNames,
	}, modelPath)
	if err != nil {
		return nil, false, fmt.Errorf("load daily ranker model for %d: %w", year, err)
	}
	modelHash, err := fileSHA256(modelPath)
	if err != nil {
		return nil, false, fmt.Errorf("hash daily ranker model for %d: %w", year, err)
	}
	s.predictors[year] = predictor
	s.modelMetadata[year] = map[string]interface{}{
		"ranker_model_loaded":               true,
		"ranker_model_sha256":               modelHash,
		"ranker_model_feature_spec_version": "daily_ranker_v1",
		"ranker_model_feature_count":        len(ml.DailyRankerFeatureNames),
	}
	return predictor, true, nil
}

func (s *DailyRankerSleeveStrategy) modelAuditMetadata(year int) map[string]interface{} {
	out := map[string]interface{}{
		"ranker_model_year":                 year,
		"ranker_model_loaded":               false,
		"ranker_model_feature_spec_version": "daily_ranker_v1",
		"ranker_model_feature_count":        len(ml.DailyRankerFeatureNames),
	}
	if s == nil {
		return out
	}
	if root := strings.TrimSpace(s.cfg.ModelArtifactRoot); root != "" {
		out["ranker_model_artifact_root"] = root
	}
	if variant := strings.TrimSpace(s.cfg.ModelVariant); variant != "" {
		out["ranker_model_variant"] = variant
	}
	if modelPath := strings.TrimSpace(s.cfg.ModelPathsByYear[year]); modelPath != "" {
		out["ranker_model_path"] = modelPath
	}
	mergeMetadata(out, s.modelMetadata[year])
	return out
}

func (s *DailyRankerSleeveStrategy) buildTargets(
	panel backtest.AlignedBars,
	index int,
	predictor *ml.LeavesPredictor,
) (map[string]backtest.TargetPosition, map[string]interface{}, error) {
	candidates, err := s.scoreCandidates(panel, index, predictor)
	if err != nil {
		return nil, nil, err
	}
	candidates = filterAndRank(candidates, s.cfg)
	risk := benchmarkRiskState(panel.Bars[s.cfg.BenchmarkSymbol], index, s.cfg)
	activeWeights := candidateWeights(candidates, s.cfg, risk.Scale)
	targets := completeWithBenchmark(activeWeights, s.cfg.BenchmarkSymbol)
	if len(s.currentTargets) > 0 && s.cfg.TurnoverBand > 0 {
		turnover := targetTurnover(s.currentTargets, targets)
		if turnover < s.cfg.TurnoverBand {
			held := cloneTargets(s.currentTargets)
			return held, map[string]interface{}{
				"engine":        DailyRankerSleeveStrategyName,
				"rebalance":     true,
				"action":        "hold_targets",
				"turnover":      turnover,
				"turnover_band": s.cfg.TurnoverBand,
			}, nil
		}
	}
	metadata := map[string]interface{}{
		"engine":                 DailyRankerSleeveStrategyName,
		"rebalance":              true,
		"candidate_count":        len(candidates),
		"selection_rows":         selectionMetadata(candidates, activeWeights),
		"active_weight":          activeWeight(targets, s.cfg.BenchmarkSymbol),
		"benchmark":              s.cfg.BenchmarkSymbol,
		"benchmark_weight":       targetWeight(targets, s.cfg.BenchmarkSymbol),
		"sleeve_scale":           risk.Scale,
		"benchmark_vol_20":       risk.BenchmarkVol20,
		"benchmark_drawdown":     risk.BenchmarkDrawdown,
		"risk_reasons":           risk.Reasons,
		"turnover":               targetTurnover(s.currentTargets, targets),
		"turnover_band":          s.cfg.TurnoverBand,
		"point_in_time_universe": s.pitUniverse != nil,
	}
	if s.pitUniversePath != "" {
		metadata["point_in_time_universe_path"] = s.pitUniversePath
	}
	return targets, metadata, nil
}

func (s *DailyRankerSleeveStrategy) scoreCandidates(
	panel backtest.AlignedBars,
	index int,
	predictor *ml.LeavesPredictor,
) ([]rankerCandidate, error) {
	barsBySymbol := panel.Bars
	eventTime := panel.Times[index]
	out := make([]rankerCandidate, 0)
	for _, symbol := range s.candidateUniverse(panel, eventTime) {
		vector, err := s.builder.BuildAtTime(barsBySymbol, symbol, eventTime)
		if err != nil {
			continue
		}
		score, err := predictor.PredictRaw(vector.Values)
		if err != nil {
			return nil, fmt.Errorf("score %s at %s: %w", symbol, eventTime.Format(time.RFC3339), err)
		}
		if !isFinite(score) {
			continue
		}
		vol20 := featureValue(vector, "vol_20")
		if s.cfg.MaxCandidateVol > 0 && vol20 > s.cfg.MaxCandidateVol {
			continue
		}
		out = append(out, rankerCandidate{
			Symbol: symbol,
			Score:  score,
			Vol20:  vol20,
		})
	}
	return out, nil
}

func (s *DailyRankerSleeveStrategy) candidateUniverse(panel backtest.AlignedBars, eventTime time.Time) []string {
	symbols := s.universe
	if len(symbols) == 0 {
		symbols = panel.Symbols
	}
	excluded := symbolSet(s.cfg.ExcludedSymbols)
	out := make([]string, 0, len(symbols))
	for _, symbol := range normalizeSymbols(symbols) {
		if symbol == s.cfg.BenchmarkSymbol || isBenchmarkProxy(symbol) || excluded[symbol] {
			continue
		}
		if s.pitUniverse != nil && !s.pitUniverse.Active(symbol, eventTime) {
			continue
		}
		switch strings.ToLower(s.cfg.CandidateUniverse) {
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

func (s *DailyRankerSleeveStrategy) hasBenchmark(panel backtest.AlignedBars, index int) bool {
	bars := panel.Bars[s.cfg.BenchmarkSymbol]
	return len(bars) > index && bars[index].Close > 0
}

func (s *DailyRankerSleeveStrategy) shouldRebalance(index int) bool {
	if s.lastRebalanceIndex < 0 {
		return true
	}
	return index-s.lastRebalanceIndex >= s.cfg.RebalanceEveryBars
}

func (s *DailyRankerSleeveStrategy) benchmarkOnlyOutput(t time.Time, reason string) backtest.PortfolioOutput {
	targets := benchmarkOnlyTargets(s.cfg.BenchmarkSymbol, reason)
	return backtest.PortfolioOutput{
		Time:          t,
		Targets:       targets,
		GrossExposure: 1,
		NetExposure:   1,
		CashWeight:    0,
		EngineMetadata: map[string]interface{}{
			"engine": DailyRankerSleeveStrategyName,
			"reason": reason,
		},
	}
}

func (s *DailyRankerSleeveStrategy) rebalanceToBenchmark(t time.Time, index int, reason string) backtest.PortfolioOutput {
	s.lastRebalanceIndex = index
	targets := benchmarkOnlyTargets(s.cfg.BenchmarkSymbol, reason)
	s.currentTargets = cloneTargets(targets)
	return backtest.PortfolioOutput{
		Time:          t,
		Targets:       targets,
		GrossExposure: 1,
		NetExposure:   1,
		CashWeight:    0,
		EngineMetadata: map[string]interface{}{
			"engine":    DailyRankerSleeveStrategyName,
			"rebalance": true,
			"reason":    reason,
		},
	}
}

func filterAndRank(candidates []rankerCandidate, cfg DailyRankerSleeveConfig) []rankerCandidate {
	if len(candidates) == 0 {
		return nil
	}
	mean, std := scoreMoments(candidates)
	filtered := make([]rankerCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if std > 0 {
			candidate.ScoreZ = (candidate.Score - mean) / std
		}
		if cfg.MinScoreZ > 0 && candidate.ScoreZ < cfg.MinScoreZ {
			continue
		}
		filtered = append(filtered, candidate)
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Score == filtered[j].Score {
			return filtered[i].Symbol < filtered[j].Symbol
		}
		return filtered[i].Score > filtered[j].Score
	})
	limit := cfg.TopK
	if limit > len(filtered) {
		limit = len(filtered)
	}
	for i := 0; i < limit; i++ {
		filtered[i].Rank = i + 1
	}
	return append([]rankerCandidate(nil), filtered[:limit]...)
}

func scoreMoments(candidates []rankerCandidate) (float64, float64) {
	var sum float64
	for _, candidate := range candidates {
		sum += candidate.Score
	}
	mean := sum / float64(len(candidates))
	var variance float64
	for _, candidate := range candidates {
		diff := candidate.Score - mean
		variance += diff * diff
	}
	return mean, math.Sqrt(variance / float64(len(candidates)))
}

func candidateWeights(candidates []rankerCandidate, cfg DailyRankerSleeveConfig, riskScale float64) map[string]float64 {
	if len(candidates) == 0 {
		return nil
	}
	effectiveSleeve := clamp01(cfg.SleeveFraction * riskScale)
	effectiveMaxName := math.Min(cfg.MaxNameWeight, effectiveSleeve)
	raw := make(map[string]float64, len(candidates))
	for _, candidate := range candidates {
		rankWeight := float64(cfg.TopK - candidate.Rank + 1)
		if rankWeight <= 0 {
			rankWeight = 1
		}
		raw[candidate.Symbol] = rankWeight / math.Max(0.05, candidate.Vol20)
	}
	return capWeightBudget(raw, effectiveSleeve, effectiveMaxName)
}

func capWeightBudget(rawWeights map[string]float64, budget float64, maxWeight float64) map[string]float64 {
	budget = math.Max(0, budget)
	maxWeight = math.Max(0, maxWeight)
	if budget <= 0 || maxWeight <= 0 || len(rawWeights) == 0 {
		return nil
	}
	positive := make(map[string]float64)
	for symbol, weight := range rawWeights {
		if weight > 0 && isFinite(weight) {
			positive[symbol] = weight
		}
	}
	if len(positive) == 0 {
		return nil
	}
	capped := make(map[string]float64)
	remaining := make(map[string]struct{}, len(positive))
	for symbol := range positive {
		remaining[symbol] = struct{}{}
	}
	remainingBudget := math.Min(budget, maxWeight*float64(len(positive)))
	for len(remaining) > 0 && remainingBudget > 1e-12 {
		var totalRaw float64
		for symbol := range remaining {
			totalRaw += positive[symbol]
		}
		if totalRaw <= 0 {
			equal := remainingBudget / float64(len(remaining))
			for symbol := range remaining {
				allocation := math.Min(maxWeight, equal)
				capped[symbol] += allocation
				remainingBudget -= allocation
				delete(remaining, symbol)
			}
			break
		}
		progressed := false
		for symbol := range remaining {
			proposed := remainingBudget * positive[symbol] / totalRaw
			if proposed >= maxWeight {
				capped[symbol] = maxWeight
				remainingBudget -= maxWeight
				delete(remaining, symbol)
				progressed = true
			}
		}
		if !progressed {
			for symbol := range remaining {
				capped[symbol] = remainingBudget * positive[symbol] / totalRaw
			}
			break
		}
	}
	for symbol, weight := range capped {
		if weight <= 1e-9 {
			delete(capped, symbol)
		}
	}
	return capped
}

func completeWithBenchmark(activeWeights map[string]float64, benchmark string) map[string]backtest.TargetPosition {
	targets := make(map[string]backtest.TargetPosition)
	var activeTotal float64
	for symbol, weight := range activeWeights {
		if weight <= 0 {
			continue
		}
		targets[symbol] = targetPosition(symbol, weight, weight, "active_sleeve")
		activeTotal += weight
	}
	benchmarkWeight := math.Max(0, 1-activeTotal)
	targets[benchmark] = targetPosition(benchmark, benchmarkWeight, 1, "benchmark_core")
	return targets
}

func benchmarkOnlyTargets(benchmark, reason string) map[string]backtest.TargetPosition {
	target := targetPosition(benchmark, 1, 1, "benchmark_core")
	target.Metadata["reason"] = reason
	return map[string]backtest.TargetPosition{benchmark: target}
}

func targetPosition(symbol string, weight float64, alphaScore float64, role string) backtest.TargetPosition {
	return backtest.TargetPosition{
		Symbol:       symbol,
		TargetWeight: weight,
		AlphaScore:   alphaScore,
		Confidence:   1,
		Side:         backtest.PositionSideLong,
		Engine:       DailyRankerSleeveStrategyName,
		Metadata: map[string]interface{}{
			"role":      role,
			"rebalance": true,
		},
	}
}

type riskState struct {
	Scale             float64
	BenchmarkVol20    float64
	BenchmarkDrawdown float64
	Reasons           []string
}

func benchmarkRiskState(bars []models.Bar, index int, cfg DailyRankerSleeveConfig) riskState {
	state := riskState{Scale: 1}
	if len(bars) == 0 || index < 0 || index >= len(bars) {
		return state
	}
	returns := logReturnsBetween(bars, maxInt(1, index-20+1), index)
	if len(returns) >= 2 {
		state.BenchmarkVol20 = populationStd(returns) * math.Sqrt(252)
	}
	start := maxInt(0, index-63+1)
	var high float64
	for i := start; i <= index; i++ {
		if bars[i].Close > high {
			high = bars[i].Close
		}
	}
	if high > 0 && bars[index].Close > 0 {
		state.BenchmarkDrawdown = bars[index].Close/high - 1
	}
	if cfg.MaxBenchmarkVol > 0 && state.BenchmarkVol20 > cfg.MaxBenchmarkVol {
		state.Scale *= clamp01(cfg.HighVolScale)
		state.Reasons = append(state.Reasons, "high_benchmark_vol")
	}
	if cfg.MaxBenchmarkDrawdown > 0 && state.BenchmarkDrawdown < -cfg.MaxBenchmarkDrawdown {
		state.Scale *= clamp01(cfg.DrawdownScale)
		state.Reasons = append(state.Reasons, "benchmark_drawdown")
	}
	return state
}

func selectionMetadata(candidates []rankerCandidate, weights map[string]float64) []map[string]interface{} {
	rows := make([]map[string]interface{}, 0, len(candidates))
	for _, candidate := range candidates {
		rows = append(rows, map[string]interface{}{
			"rank":          candidate.Rank,
			"symbol":        candidate.Symbol,
			"score":         candidate.Score,
			"score_z":       candidate.ScoreZ,
			"vol_20":        candidate.Vol20,
			"target_weight": weights[candidate.Symbol],
		})
	}
	return rows
}

func featureValue(vector ml.FeatureVector, name string) float64 {
	for i, featureName := range vector.Names {
		if featureName == name && i < len(vector.Values) {
			return vector.Values[i]
		}
	}
	return 0
}

func (c DailyRankerSleeveConfig) withDefaults() DailyRankerSleeveConfig {
	c.BenchmarkSymbol = strings.ToUpper(strings.TrimSpace(c.BenchmarkSymbol))
	if c.BenchmarkSymbol == "" {
		c.BenchmarkSymbol = "VOO"
	}
	c.CandidateUniverse = strings.ToLower(strings.TrimSpace(c.CandidateUniverse))
	if c.CandidateUniverse == "" {
		c.CandidateUniverse = "stocks"
	}
	c.ExcludedSymbols = normalizeSymbols(c.ExcludedSymbols)
	c.PointInTimeUniversePath = strings.TrimSpace(c.PointInTimeUniversePath)
	c.ModelArtifactRoot = strings.TrimSpace(c.ModelArtifactRoot)
	c.ModelVariant = strings.TrimSpace(c.ModelVariant)
	if c.RebalanceEveryBars <= 0 {
		c.RebalanceEveryBars = 63
	}
	if c.SleeveFraction <= 0 {
		c.SleeveFraction = 0.10
	}
	c.SleeveFraction = clamp01(c.SleeveFraction)
	if c.TopK <= 0 {
		c.TopK = 3
	}
	if c.MaxNameWeight <= 0 {
		c.MaxNameWeight = c.SleeveFraction / float64(c.TopK)
	}
	c.MaxNameWeight = clamp01(c.MaxNameWeight)
	c.TurnoverBand = clamp01(c.TurnoverBand)
	if c.HighVolScale < 0 {
		c.HighVolScale = 0
	}
	if c.DrawdownScale < 0 {
		c.DrawdownScale = 0
	}
	return c
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func mergeMetadata(dst map[string]interface{}, src map[string]interface{}) {
	if dst == nil || len(src) == 0 {
		return
	}
	for key, value := range src {
		dst[key] = value
	}
}

func cloneInterfaceMap(in map[string]interface{}) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func symbolSet(symbols []string) map[string]bool {
	out := make(map[string]bool, len(symbols))
	for _, symbol := range normalizeSymbols(symbols) {
		out[symbol] = true
	}
	return out
}

func panelPrefix(panel backtest.AlignedBars, n int) backtest.AlignedBars {
	if n < 0 {
		n = 0
	}
	if n > len(panel.Times) {
		n = len(panel.Times)
	}
	out := backtest.AlignedBars{
		Times:      append([]time.Time(nil), panel.Times[:n]...),
		Symbols:    append([]string(nil), panel.Symbols...),
		Bars:       make(map[string][]models.Bar, len(panel.Bars)),
		Timeframe:  panel.Timeframe,
		Feed:       panel.Feed,
		Adjustment: panel.Adjustment,
		Metadata:   panel.Metadata,
	}
	for symbol, bars := range panel.Bars {
		limit := n
		if limit > len(bars) {
			limit = len(bars)
		}
		out.Bars[symbol] = append([]models.Bar(nil), bars[:limit]...)
	}
	return out
}

func targetTurnover(previous map[string]backtest.TargetPosition, next map[string]backtest.TargetPosition) float64 {
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
	return targets[symbol].TargetWeight
}

func activeWeight(targets map[string]backtest.TargetPosition, benchmark string) float64 {
	var active float64
	for symbol, target := range targets {
		if symbol == benchmark {
			continue
		}
		active += target.TargetWeight
	}
	return active
}

func grossExposure(targets map[string]backtest.TargetPosition) float64 {
	var gross float64
	for _, target := range targets {
		gross += math.Abs(target.TargetWeight)
	}
	return gross
}

func cloneTargets(targets map[string]backtest.TargetPosition) map[string]backtest.TargetPosition {
	out := make(map[string]backtest.TargetPosition, len(targets))
	for symbol, target := range targets {
		out[symbol] = target
	}
	return out
}

func logReturnsBetween(bars []models.Bar, start, end int) []float64 {
	if start < 1 {
		start = 1
	}
	if end >= len(bars) {
		end = len(bars) - 1
	}
	if start > end {
		return nil
	}
	out := make([]float64, 0, end-start+1)
	for i := start; i <= end; i++ {
		if bars[i-1].Close <= 0 || bars[i].Close <= 0 {
			continue
		}
		out = append(out, math.Log(bars[i].Close/bars[i-1].Close))
	}
	return out
}

func populationStd(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, value := range values {
		sum += value
	}
	mean := sum / float64(len(values))
	var variance float64
	for _, value := range values {
		diff := value - mean
		variance += diff * diff
	}
	return math.Sqrt(variance / float64(len(values)))
}

func normalizeSymbols(symbols []string) []string {
	out := make([]string, 0, len(symbols))
	seen := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
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

func isBenchmarkProxy(symbol string) bool {
	switch strings.ToUpper(symbol) {
	case "SPY", "VOO":
		return true
	default:
		return false
	}
}

func isETFSymbol(symbol string) bool {
	switch strings.ToUpper(symbol) {
	case "DIA", "IWM", "QQQ", "SMH", "SPY", "VOO", "VTI", "XLB", "XLC", "XLE", "XLF", "XLI", "XLK", "XLP", "XLRE", "XLU", "XLV", "XLY":
		return true
	default:
		return false
	}
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
