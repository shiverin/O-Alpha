package alphavalidation

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/oalpha/internal/agent"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

type AllocationPlan struct {
	TargetWeights   map[string]float64     `json:"target_weights"`
	ActiveSymbols   []string               `json:"active_symbols,omitempty"`
	ActiveSleevePct float64                `json:"active_sleeve_pct"`
	RegimeLabel     string                 `json:"regime_label,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type BenchmarkRankerProxyStrategy struct {
	config        VariantConfig
	benchmarkBars []models.Bar
	activeBars    map[string][]models.Bar
}

type rankedSymbol struct {
	symbol string
	score  float64
}

type timeKey int64

func NewBenchmarkRankerProxyStrategy(config VariantConfig, barsBySymbol map[string][]models.Bar) (*BenchmarkRankerProxyStrategy, error) {
	benchmarkBars := barsBySymbol[config.BenchmarkSymbol]
	if len(benchmarkBars) == 0 {
		return nil, fmt.Errorf("benchmark bars for %s are required", config.BenchmarkSymbol)
	}
	activeBars := alignActiveBars(benchmarkBars, barsBySymbol, config.ActiveSymbols)
	return &BenchmarkRankerProxyStrategy{
		config:        config,
		benchmarkBars: append([]models.Bar(nil), benchmarkBars...),
		activeBars:    activeBars,
	}, nil
}

func (s *BenchmarkRankerProxyStrategy) ActiveSymbols() []string {
	symbols := make([]string, 0, len(s.activeBars))
	for symbol := range s.activeBars {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	return symbols
}

func (s *BenchmarkRankerProxyStrategy) ActiveBars() map[string][]models.Bar {
	out := make(map[string][]models.Bar, len(s.activeBars))
	for symbol, bars := range s.activeBars {
		out[symbol] = append([]models.Bar(nil), bars...)
	}
	return out
}

func (s *BenchmarkRankerProxyStrategy) BenchmarkBars() []models.Bar {
	return append([]models.Bar(nil), s.benchmarkBars...)
}

func (s *BenchmarkRankerProxyStrategy) GeneratePlans(_ context.Context, trainEnd, end int) ([]AllocationPlan, error) {
	if trainEnd <= 0 || end <= 0 || end > len(s.benchmarkBars) || trainEnd > end {
		return nil, fmt.Errorf("invalid train/end indexes: trainEnd=%d end=%d len=%d", trainEnd, end, len(s.benchmarkBars))
	}

	plans := make([]AllocationPlan, end)
	selected := make([]string, 0, s.config.MaxPositions)
	maxLookback := maxIntSlice(s.config.LookbackBars)

	var detector *agent.HMMRegimeDetector
	var overlay *agent.RegimeRiskOverlay
	if s.config.UseRiskOverlay {
		windowSize := s.config.VolWindow
		if windowSize < 20 {
			windowSize = 20
		}
		detector = agent.NewHMMRegimeDetector(windowSize)
		encoder := agent.NewObservationEncoder(detector.WindowSize())
		if trainEnd >= detector.WindowSize() {
			if err := encoder.FitBuckets(s.benchmarkBars[:trainEnd]); err == nil {
				detector.UpdateBuckets(encoder.VolBuckets, encoder.TrendBuckets)
			}
		}
		overlay = agent.NewRegimeRiskOverlay(s.config.RiskPolicy)
	}

	for i := 0; i < end; i++ {
		plan := AllocationPlan{
			TargetWeights: map[string]float64{s.config.BenchmarkSymbol: 1.0},
			Metadata:      make(map[string]interface{}),
		}
		if i < maxLookback || len(s.activeBars) == 0 {
			plans[i] = plan
			continue
		}
		if shouldRebalance(i, maxLookback, s.config.RebalanceBars) {
			ranked := s.rankAt(i)
			selected = applyTurnoverBuffer(selected, ranked, s.config.MaxPositions, s.config.TurnoverBufferRanks)
		}

		activeSleeve := s.config.ActiveSleevePct
		regimeLabel := "benchmark_core"
		if s.config.UseRiskOverlay && detector != nil && overlay != nil {
			_, confidence, err := detector.Update(s.benchmarkBars[:i+1])
			probs := detector.GetProbabilities()
			decision := overlay.Apply(agent.RegimeOverlayInput{
				BaseExposure:      activeSleeve,
				PosteriorProbs:    []float64{probs[0], probs[1], probs[2]},
				StateRoles:        []agent.RegimeRiskRole{agent.RegimeRiskLowVol, agent.RegimeRiskNormal, agent.RegimeRiskHighVol},
				ModelHealthy:      err == nil,
				RealizedAnnualVol: annualizedVolatility(s.benchmarkBars, i+1, detector.WindowSize()),
				PeakEquity:        rollingPeakClose(s.benchmarkBars[:i+1]),
				CurrentEquity:     s.benchmarkBars[i].Close,
			})
			activeSleeve = decision.AdjustedExposure
			regimeLabel = string(decision.EffectiveRole)
			plan.Metadata["overlay_confidence"] = confidence
			plan.Metadata["overlay_multiplier"] = decision.Multiplier
			plan.Metadata["overlay_reasons"] = decision.Reasons
		}
		if len(selected) == 0 || activeSleeve <= 0 {
			plan.RegimeLabel = regimeLabel
			plans[i] = plan
			continue
		}

		benchmarkWeight := 1.0 - activeSleeve
		if benchmarkWeight < 0 {
			benchmarkWeight = 0
		}
		weights := map[string]float64{s.config.BenchmarkSymbol: benchmarkWeight}
		perSymbol := activeSleeve / float64(len(selected))
		for _, symbol := range selected {
			weights[symbol] = perSymbol
		}
		plan.TargetWeights = weights
		plan.ActiveSymbols = append([]string(nil), selected...)
		plan.ActiveSleevePct = activeSleeve
		plan.RegimeLabel = regimeLabel
		plan.Metadata["selected_symbols"] = append([]string(nil), selected...)
		plans[i] = plan
	}
	return plans, nil
}

func (s *BenchmarkRankerProxyStrategy) rankAt(index int) []rankedSymbol {
	ranked := make([]rankedSymbol, 0, len(s.activeBars))
	for symbol, bars := range s.activeBars {
		score := weightedLookbackScore(bars, index, s.config.LookbackBars, s.config.LookbackWeights)
		ranked = append(ranked, rankedSymbol{symbol: symbol, score: score})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			return ranked[i].symbol < ranked[j].symbol
		}
		return ranked[i].score > ranked[j].score
	})
	return ranked
}

func RunPortfolioBacktest(benchmarkSymbol string, benchmarkBars []models.Bar, activeBars map[string][]models.Bar, plans []AllocationPlan, initialCash, transactionCostBPS float64) (*models.BacktestResult, error) {
	if len(benchmarkBars) == 0 {
		return nil, fmt.Errorf("benchmark bars are required")
	}
	if len(plans) != len(benchmarkBars) {
		return nil, fmt.Errorf("plans length %d does not match bars length %d", len(plans), len(benchmarkBars))
	}

	cash := cashBase(initialCash)
	holdings := make(map[string]float64)
	equityCurve := make([]models.EquityPoint, 0, len(benchmarkBars))
	turnoverValue := 0.0
	feesPaid := 0.0
	numTrades := 0
	exposureBars := 0
	tradeRate := transactionCostBPS / 10_000.0

	for i := range benchmarkBars {
		if i > 0 {
			prices := buildOpenPriceMap(benchmarkSymbol, benchmarkBars, activeBars, i)
			equityAtOpen := cash + holdingsValue(holdings, prices)
			target := plans[i-1].TargetWeights
			for _, symbol := range unionSymbols(holdings, target) {
				price := prices[symbol]
				if price <= 0 {
					continue
				}
				currentShares := holdings[symbol]
				currentValue := currentShares * price
				targetValue := equityAtOpen * target[symbol]
				delta := targetValue - currentValue
				if delta >= 0 {
					continue
				}
				sharesToSell := math.Min(currentShares, (-delta)/price)
				if sharesToSell <= 0 {
					continue
				}
				proceeds := sharesToSell * price
				fee := proceeds * tradeRate
				cash += proceeds - fee
				remaining := currentShares - sharesToSell
				if remaining <= 1e-9 {
					delete(holdings, symbol)
				} else {
					holdings[symbol] = remaining
				}
				turnoverValue += proceeds
				feesPaid += fee
				numTrades++
			}
			for _, symbol := range unionSymbols(holdings, target) {
				price := prices[symbol]
				if price <= 0 {
					continue
				}
				currentShares := holdings[symbol]
				currentValue := currentShares * price
				targetValue := equityAtOpen * target[symbol]
				delta := targetValue - currentValue
				if delta <= 0 {
					continue
				}
				maxBuyValue := cash
				if tradeRate > 0 {
					maxBuyValue = cash / (1 + tradeRate)
				}
				buyValue := math.Min(delta, maxBuyValue)
				if buyValue <= 0 {
					continue
				}
				sharesToBuy := buyValue / price
				if sharesToBuy <= 0 {
					continue
				}
				fee := buyValue * tradeRate
				cash -= buyValue + fee
				holdings[symbol] = currentShares + sharesToBuy
				turnoverValue += buyValue
				feesPaid += fee
				numTrades++
			}
		}

		closePrices := buildClosePriceMap(benchmarkSymbol, benchmarkBars, activeBars, i)
		equity := cash + holdingsValue(holdings, closePrices)
		if holdingsValue(holdings, closePrices) > 0 {
			exposureBars++
		}
		equityCurve = append(equityCurve, models.EquityPoint{Time: benchmarkBars[i].Time, Equity: equity})
	}

	values := make([]float64, len(equityCurve))
	for i, point := range equityCurve {
		values[i] = point.Equity
	}
	metrics := backtest.ComputeMetrics(values)
	return &models.BacktestResult{
		Symbol:           benchmarkSymbol,
		EquityCurve:      equityCurve,
		FinalEquity:      equityCurve[len(equityCurve)-1].Equity,
		TotalReturn:      metrics.TotalReturn,
		AnnualizedReturn: metrics.AnnualizedReturn,
		Sharpe:           metrics.Sharpe,
		Sortino:          metrics.Sortino,
		Calmar:           metrics.Calmar,
		MaxDrawdown:      metrics.MaxDrawdown,
		NumTrades:        numTrades,
		ExposurePercent:  float64(exposureBars) / float64(len(equityCurve)),
		Turnover:         turnoverValue / cashBase(initialCash),
		FeesPaid:         feesPaid,
		SlippageCost:     feesPaid,
	}, nil
}

func weightedLookbackScore(bars []models.Bar, index int, lookbacks []int, weights []float64) float64 {
	if index <= 0 || len(lookbacks) == 0 {
		return 0
	}
	total := 0.0
	weightTotal := 0.0
	for i, lookback := range lookbacks {
		if lookback <= 0 || index < lookback {
			continue
		}
		weight := 1.0
		if i < len(weights) {
			weight = weights[i]
		}
		past := bars[index-lookback].Close
		current := bars[index].Close
		if past <= 0 || current <= 0 {
			continue
		}
		total += weight * (current/past - 1.0)
		weightTotal += weight
	}
	if weightTotal == 0 {
		return 0
	}
	return total / weightTotal
}

func applyTurnoverBuffer(previous []string, ranked []rankedSymbol, maxPositions, buffer int) []string {
	if maxPositions <= 0 || len(ranked) == 0 {
		return nil
	}
	rankIndex := make(map[string]int, len(ranked))
	for i, item := range ranked {
		rankIndex[item.symbol] = i
	}
	keepLimit := maxPositions + buffer
	kept := make([]string, 0, maxPositions)
	for _, symbol := range previous {
		if idx, ok := rankIndex[symbol]; ok && idx < keepLimit {
			kept = append(kept, symbol)
		}
	}
	sort.Slice(kept, func(i, j int) bool { return rankIndex[kept[i]] < rankIndex[kept[j]] })
	if len(kept) > maxPositions {
		kept = kept[:maxPositions]
	}
	seen := make(map[string]struct{}, len(kept))
	for _, symbol := range kept {
		seen[symbol] = struct{}{}
	}
	for _, item := range ranked {
		if len(kept) >= maxPositions {
			break
		}
		if _, ok := seen[item.symbol]; ok {
			continue
		}
		kept = append(kept, item.symbol)
	}
	return kept
}

func shouldRebalance(index, maxLookback, rebalanceBars int) bool {
	if rebalanceBars <= 0 {
		rebalanceBars = 1
	}
	if index == maxLookback {
		return true
	}
	return index > maxLookback && (index-maxLookback)%rebalanceBars == 0
}

func alignActiveBars(benchmarkBars []models.Bar, barsBySymbol map[string][]models.Bar, symbols []string) map[string][]models.Bar {
	benchmarkTimes := make([]timeKey, len(benchmarkBars))
	for i, bar := range benchmarkBars {
		benchmarkTimes[i] = timeToKey(bar.Time)
	}
	out := make(map[string][]models.Bar)
	for _, symbol := range symbols {
		bars := barsBySymbol[symbol]
		if len(bars) == 0 {
			continue
		}
		byTime := make(map[timeKey]models.Bar, len(bars))
		for _, bar := range bars {
			byTime[timeToKey(bar.Time)] = bar
		}
		aligned := make([]models.Bar, 0, len(benchmarkBars))
		missing := false
		for _, ts := range benchmarkTimes {
			bar, ok := byTime[ts]
			if !ok {
				missing = true
				break
			}
			aligned = append(aligned, bar)
		}
		if !missing {
			out[symbol] = aligned
		}
	}
	return out
}

func buildOpenPriceMap(benchmarkSymbol string, benchmarkBars []models.Bar, activeBars map[string][]models.Bar, index int) map[string]float64 {
	prices := map[string]float64{benchmarkSymbol: benchmarkBars[index].Open}
	for symbol, bars := range activeBars {
		prices[symbol] = bars[index].Open
	}
	return prices
}

func buildClosePriceMap(benchmarkSymbol string, benchmarkBars []models.Bar, activeBars map[string][]models.Bar, index int) map[string]float64 {
	prices := map[string]float64{benchmarkSymbol: benchmarkBars[index].Close}
	for symbol, bars := range activeBars {
		prices[symbol] = bars[index].Close
	}
	return prices
}

func holdingsValue(holdings map[string]float64, prices map[string]float64) float64 {
	total := 0.0
	for symbol, qty := range holdings {
		price := prices[symbol]
		if price <= 0 || qty == 0 {
			continue
		}
		total += qty * price
	}
	return total
}

func unionSymbols(holdings map[string]float64, target map[string]float64) []string {
	set := make(map[string]struct{}, len(holdings)+len(target))
	for symbol := range holdings {
		set[symbol] = struct{}{}
	}
	for symbol := range target {
		set[symbol] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for symbol := range set {
		out = append(out, symbol)
	}
	sort.Strings(out)
	return out
}

func annualizedVolatility(bars []models.Bar, end, windowSize int) float64 {
	start := end - windowSize
	if start < 0 {
		start = 0
	}
	return agent.RealizedVolatility(bars[start:end]) * math.Sqrt(252)
}

func rollingPeakClose(bars []models.Bar) float64 {
	peak := 0.0
	for _, bar := range bars {
		if bar.Close > peak {
			peak = bar.Close
		}
	}
	return peak
}

func maxIntSlice(values []int) int {
	maxValue := 0
	for _, value := range values {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func cashBase(initialCash float64) float64 {
	if initialCash <= 0 {
		return 100_000
	}
	return initialCash
}

func sliceBars(bars []models.Bar, start, end int) []models.Bar {
	return append([]models.Bar(nil), bars[start:end]...)
}

func sliceBarsMap(barsBySymbol map[string][]models.Bar, start, end int) map[string][]models.Bar {
	out := make(map[string][]models.Bar, len(barsBySymbol))
	for symbol, bars := range barsBySymbol {
		out[symbol] = sliceBars(bars, start, end)
	}
	return out
}

func chainBacktestResults(results []*models.BacktestResult, initialCash float64) *models.BacktestResult {
	if len(results) == 0 {
		return &models.BacktestResult{}
	}
	currentEquity := cashBase(initialCash)
	combinedCurve := make([]models.EquityPoint, 0)
	totalTrades := 0
	totalFees := 0.0
	totalTurnover := 0.0
	exposureBars := 0
	totalBars := 0
	for _, result := range results {
		if result == nil || len(result.EquityCurve) == 0 {
			continue
		}
		base := result.EquityCurve[0].Equity
		if base <= 0 {
			base = cashBase(initialCash)
		}
		for _, point := range result.EquityCurve {
			factor := point.Equity / base
			combinedCurve = append(combinedCurve, models.EquityPoint{Time: point.Time, Equity: currentEquity * factor})
		}
		currentEquity = combinedCurve[len(combinedCurve)-1].Equity
		totalTrades += result.NumTrades
		totalFees += result.FeesPaid
		totalTurnover += result.Turnover
		exposureBars += int(result.ExposurePercent * float64(len(result.EquityCurve)))
		totalBars += len(result.EquityCurve)
	}
	values := make([]float64, len(combinedCurve))
	for i, point := range combinedCurve {
		values[i] = point.Equity
	}
	metrics := backtest.ComputeMetrics(values)
	exposurePercent := 0.0
	if totalBars > 0 {
		exposurePercent = float64(exposureBars) / float64(totalBars)
	}
	return &models.BacktestResult{
		EquityCurve:      combinedCurve,
		FinalEquity:      currentEquity,
		TotalReturn:      metrics.TotalReturn,
		AnnualizedReturn: metrics.AnnualizedReturn,
		Sharpe:           metrics.Sharpe,
		Sortino:          metrics.Sortino,
		Calmar:           metrics.Calmar,
		MaxDrawdown:      metrics.MaxDrawdown,
		NumTrades:        totalTrades,
		ExposurePercent:  exposurePercent,
		Turnover:         totalTurnover,
		FeesPaid:         totalFees,
		SlippageCost:     totalFees,
	}
}

func timeToKey(t time.Time) timeKey {
	return timeKey(t.UTC().UnixNano())
}
