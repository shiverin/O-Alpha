package ml

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

const tradingDaysPerYear = 252.0

type FeatureSpec struct {
	Version          string   `json:"version"`
	Features         []string `json:"features"`
	RequiredLookback int      `json:"required_lookback"`
	ContextSymbols   []string `json:"context_symbols,omitempty"`
}

type FeatureBuildInput struct {
	Symbol                 string
	Bars                   []models.Bar
	BaseOutput             *backtest.StrategyOutput
	ContextBars            map[string][]models.Bar
	SectorSymbol           string
	HMMRegimeProbabilities map[string]float64
}

type FeatureVector struct {
	Symbol   string                 `json:"symbol"`
	Time     time.Time              `json:"time"`
	Names    []string               `json:"names"`
	Values   []float64              `json:"values"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type FeatureBuilder struct {
	spec FeatureSpec
}

func DefaultFeatureSpec() FeatureSpec {
	return FeatureSpec{
		Version:          "ml_meta_label_v1",
		RequiredLookback: 64,
		ContextSymbols:   []string{"SPY", "QQQ", "IWM"},
		Features: []string{
			"fracdiff_close_d0_5",
			"fracdiff_log_close_d0_5",
			"log_ret_1",
			"log_ret_2",
			"log_ret_5",
			"log_ret_10",
			"log_ret_21",
			"rolling_ret_21",
			"rolling_ret_63",
			"distance_to_20d_high",
			"distance_to_20d_low",
			"close_to_vwap_proxy",
			"close_to_close_vol_20",
			"ewma_vol_20",
			"parkinson_vol_20",
			"garman_klass_vol_20",
			"rogers_satchell_vol_20",
			"atr_pct_14",
			"ma_fast_minus_slow_pct",
			"ma_fast_slope",
			"ma_slow_slope",
			"kalman_residual",
			"kalman_zscore",
			"ensemble_score",
			"ensemble_confidence",
			"hmm_regime_probability_low",
			"hmm_regime_probability_medium",
			"hmm_regime_probability_high",
			"spy_ret_1",
			"spy_ret_5",
			"spy_vol_20",
			"qqq_ret_1",
			"qqq_ret_5",
			"iwm_ret_1",
			"iwm_ret_5",
			"sector_etf_ret_1",
			"sector_etf_ret_5",
			"relative_strength_vs_spy_21",
			"beta_to_spy_63",
			"residual_ret_vs_spy_21",
			"volume_z_20",
			"dollar_volume_z_20",
			"order_book_imbalance",
			"signed_volume_imbalance_20",
			"amihud_illiquidity_20",
			"high_low_spread_proxy",
			"turnover_proxy",
			"gap_pct",
		},
	}
}

func NewFeatureBuilder(spec FeatureSpec) *FeatureBuilder {
	if len(spec.Features) == 0 {
		spec = DefaultFeatureSpec()
	}
	if spec.Version == "" {
		spec.Version = "custom"
	}
	return &FeatureBuilder{spec: spec}
}

func (b *FeatureBuilder) Spec() FeatureSpec {
	if b == nil || len(b.spec.Features) == 0 {
		return DefaultFeatureSpec()
	}
	return b.spec
}

func (b *FeatureBuilder) BuildLatest(input FeatureBuildInput) (FeatureVector, error) {
	if len(input.Bars) == 0 {
		return FeatureVector{}, fmt.Errorf("cannot build features from empty bars")
	}
	return b.BuildAt(input, len(input.Bars)-1)
}

func (b *FeatureBuilder) BuildAt(input FeatureBuildInput, index int) (FeatureVector, error) {
	if len(input.Bars) == 0 {
		return FeatureVector{}, fmt.Errorf("cannot build features from empty bars")
	}
	if index < 0 || index >= len(input.Bars) {
		return FeatureVector{}, fmt.Errorf("feature index %d outside bars length %d", index, len(input.Bars))
	}

	spec := b.Spec()
	values := make([]float64, len(spec.Features))
	for i, name := range spec.Features {
		values[i] = cleanFinite(b.computeFeature(name, input, index))
	}

	return FeatureVector{
		Symbol: inputSymbol(input, index),
		Time:   input.Bars[index].Time,
		Names:  append([]string(nil), spec.Features...),
		Values: values,
		Metadata: map[string]interface{}{
			"feature_spec_version": spec.Version,
			"point_in_time_index":  index,
		},
	}, nil
}

func (b *FeatureBuilder) computeFeature(name string, input FeatureBuildInput, index int) float64 {
	eventTime := input.Bars[index].Time

	switch name {
	case "fracdiff_close_d0_5":
		return fractionalDiffClose(input.Bars, index, 0.5, 64, false)
	case "fracdiff_log_close_d0_5":
		return fractionalDiffClose(input.Bars, index, 0.5, 64, true)
	case "log_ret_1":
		return logReturn(input.Bars, index, 1)
	case "log_ret_2":
		return logReturn(input.Bars, index, 2)
	case "log_ret_5":
		return logReturn(input.Bars, index, 5)
	case "log_ret_10":
		return logReturn(input.Bars, index, 10)
	case "log_ret_21", "rolling_ret_21":
		return logReturn(input.Bars, index, 21)
	case "rolling_ret_63":
		return logReturn(input.Bars, index, 63)
	case "distance_to_20d_high":
		return distanceToHigh(input.Bars, index, 20)
	case "distance_to_20d_low":
		return distanceToLow(input.Bars, index, 20)
	case "close_to_vwap_proxy":
		return closeToVWAPProxy(input.Bars, index, 20)
	case "close_to_close_vol_20":
		return closeToCloseVol(input.Bars, index, 20)
	case "ewma_vol_20":
		return ewmaVol(input.Bars, index, 20)
	case "parkinson_vol_20":
		return parkinsonVol(input.Bars, index, 20)
	case "garman_klass_vol_20":
		return garmanKlassVol(input.Bars, index, 20)
	case "rogers_satchell_vol_20":
		return rogersSatchellVol(input.Bars, index, 20)
	case "atr_pct_14":
		return atrPct(input.Bars, index, 14)
	case "ma_fast_minus_slow_pct":
		fast := smaClose(input.Bars, index, 20)
		slow := smaClose(input.Bars, index, 50)
		return safeRatio(fast-slow, slow)
	case "ma_fast_slope":
		return smaSlope(input.Bars, index, 20)
	case "ma_slow_slope":
		return smaSlope(input.Bars, index, 50)
	case "kalman_residual":
		return baseMetadataFloat(input.BaseOutput, "kalman_residual", "residual")
	case "kalman_zscore":
		return baseMetadataFloat(input.BaseOutput, "kalman_zscore", "kalman_z_score", "zscore", "z_score")
	case "ensemble_score":
		if input.BaseOutput != nil && input.BaseOutput.AlphaScore != 0 {
			return input.BaseOutput.AlphaScore
		}
		return baseMetadataFloat(input.BaseOutput, "ensemble_score", "alpha_score", "score")
	case "ensemble_confidence":
		if input.BaseOutput != nil && input.BaseOutput.Confidence != 0 {
			return input.BaseOutput.Confidence
		}
		return baseMetadataFloat(input.BaseOutput, "ensemble_confidence", "confidence")
	case "hmm_regime_probability_low":
		return regimeProbability(input.HMMRegimeProbabilities, "low", "low_vol", "bullish")
	case "hmm_regime_probability_medium":
		return regimeProbability(input.HMMRegimeProbabilities, "medium", "normal", "neutral")
	case "hmm_regime_probability_high":
		return regimeProbability(input.HMMRegimeProbabilities, "high", "high_vol", "crisis", "bearish")
	case "spy_ret_1":
		return contextLogReturn(input.ContextBars, "SPY", eventTime, 1)
	case "spy_ret_5":
		return contextLogReturn(input.ContextBars, "SPY", eventTime, 5)
	case "spy_vol_20":
		return contextVol(input.ContextBars, "SPY", eventTime, 20)
	case "qqq_ret_1":
		return contextLogReturn(input.ContextBars, "QQQ", eventTime, 1)
	case "qqq_ret_5":
		return contextLogReturn(input.ContextBars, "QQQ", eventTime, 5)
	case "iwm_ret_1":
		return contextLogReturn(input.ContextBars, "IWM", eventTime, 1)
	case "iwm_ret_5":
		return contextLogReturn(input.ContextBars, "IWM", eventTime, 5)
	case "sector_etf_ret_1":
		return contextLogReturn(input.ContextBars, input.SectorSymbol, eventTime, 1)
	case "sector_etf_ret_5":
		return contextLogReturn(input.ContextBars, input.SectorSymbol, eventTime, 5)
	case "relative_strength_vs_spy_21":
		return logReturn(input.Bars, index, 21) - contextLogReturn(input.ContextBars, "SPY", eventTime, 21)
	case "beta_to_spy_63":
		return betaToContext(input.Bars, index, input.ContextBars, "SPY", eventTime, 63)
	case "residual_ret_vs_spy_21":
		beta := betaToContext(input.Bars, index, input.ContextBars, "SPY", eventTime, 63)
		return logReturn(input.Bars, index, 21) - beta*contextLogReturn(input.ContextBars, "SPY", eventTime, 21)
	case "volume_z_20":
		return volumeZ(input.Bars, index, 20)
	case "dollar_volume_z_20":
		return dollarVolumeZ(input.Bars, index, 20)
	case "order_book_imbalance":
		return baseMetadataFloat(input.BaseOutput, "order_book_imbalance", "book_imbalance", "l2_imbalance")
	case "signed_volume_imbalance_20":
		return signedVolumeImbalance(input.Bars, index, 20)
	case "amihud_illiquidity_20":
		return amihudIlliquidity(input.Bars, index, 20)
	case "high_low_spread_proxy":
		return highLowSpreadProxy(input.Bars, index)
	case "turnover_proxy":
		return turnoverProxy(input.Bars, index, 20)
	case "gap_pct":
		if index == 0 || input.Bars[index-1].Close <= 0 {
			return 0
		}
		return input.Bars[index].Open/input.Bars[index-1].Close - 1
	default:
		return 0
	}
}

func fractionalDiffClose(bars []models.Bar, index int, d float64, maxLookback int, useLog bool) float64 {
	if index < 0 || index >= len(bars) {
		return 0
	}
	if maxLookback <= 0 {
		maxLookback = 64
	}
	lookback := minInt(maxLookback, index+1)
	weights := fractionalDiffWeights(d, lookback)
	var out float64
	for k := 0; k < lookback; k++ {
		close := bars[index-k].Close
		if close <= 0 {
			continue
		}
		value := close
		if useLog {
			value = math.Log(close)
		}
		out += weights[k] * value
	}
	return out
}

func fractionalDiffWeights(d float64, length int) []float64 {
	if length <= 0 {
		return nil
	}
	weights := make([]float64, length)
	weights[0] = 1
	for k := 1; k < length; k++ {
		weights[k] = -weights[k-1] * (d - float64(k) + 1) / float64(k)
	}
	return weights
}

func inputSymbol(input FeatureBuildInput, index int) string {
	if input.Symbol != "" {
		return strings.ToUpper(input.Symbol)
	}
	if index >= 0 && index < len(input.Bars) {
		return strings.ToUpper(input.Bars[index].Symbol)
	}
	return ""
}

func logReturn(bars []models.Bar, index int, lag int) float64 {
	if lag <= 0 || index-lag < 0 || index >= len(bars) {
		return 0
	}
	current := bars[index].Close
	previous := bars[index-lag].Close
	if current <= 0 || previous <= 0 {
		return 0
	}
	return math.Log(current / previous)
}

func distanceToHigh(bars []models.Bar, index, window int) float64 {
	if index < 0 || index >= len(bars) || bars[index].Close <= 0 {
		return 0
	}
	start := maxInt(0, index-window+1)
	high := 0.0
	for i := start; i <= index; i++ {
		if bars[i].High > high {
			high = bars[i].High
		}
	}
	if high <= 0 {
		return 0
	}
	return bars[index].Close/high - 1
}

func distanceToLow(bars []models.Bar, index, window int) float64 {
	if index < 0 || index >= len(bars) || bars[index].Close <= 0 {
		return 0
	}
	start := maxInt(0, index-window+1)
	low := math.MaxFloat64
	for i := start; i <= index; i++ {
		if bars[i].Low > 0 && bars[i].Low < low {
			low = bars[i].Low
		}
	}
	if low == math.MaxFloat64 || low <= 0 {
		return 0
	}
	return bars[index].Close/low - 1
}

func closeToVWAPProxy(bars []models.Bar, index, window int) float64 {
	if index < 0 || index >= len(bars) || bars[index].Close <= 0 {
		return 0
	}
	start := maxInt(0, index-window+1)
	var weightedClose float64
	var volume float64
	for i := start; i <= index; i++ {
		if bars[i].Volume <= 0 || bars[i].Close <= 0 {
			continue
		}
		v := float64(bars[i].Volume)
		weightedClose += bars[i].Close * v
		volume += v
	}
	if volume <= 0 {
		return 0
	}
	return bars[index].Close/(weightedClose/volume) - 1
}

func closeToCloseVol(bars []models.Bar, index, window int) float64 {
	return annualizedStd(logReturnsEndingAt(bars, index, window))
}

func ewmaVol(bars []models.Bar, index, window int) float64 {
	returns := logReturnsEndingAt(bars, index, window)
	if len(returns) == 0 {
		return 0
	}
	lambda := 2.0 / (float64(window) + 1.0)
	var variance float64
	var weightSum float64
	weight := 1.0
	for i := len(returns) - 1; i >= 0; i-- {
		variance += weight * returns[i] * returns[i]
		weightSum += weight
		weight *= 1 - lambda
	}
	if weightSum <= 0 {
		return 0
	}
	return math.Sqrt(variance/weightSum) * math.Sqrt(tradingDaysPerYear)
}

func parkinsonVol(bars []models.Bar, index, window int) float64 {
	start := maxInt(0, index-window+1)
	var sum float64
	var n int
	for i := start; i <= index && i < len(bars); i++ {
		if bars[i].High <= 0 || bars[i].Low <= 0 {
			continue
		}
		v := math.Log(bars[i].High / bars[i].Low)
		sum += v * v
		n++
	}
	if n == 0 {
		return 0
	}
	variance := sum / (4 * math.Log(2) * float64(n))
	return math.Sqrt(math.Max(0, variance)) * math.Sqrt(tradingDaysPerYear)
}

func garmanKlassVol(bars []models.Bar, index, window int) float64 {
	start := maxInt(0, index-window+1)
	var sum float64
	var n int
	for i := start; i <= index && i < len(bars); i++ {
		bar := bars[i]
		if bar.High <= 0 || bar.Low <= 0 || bar.Open <= 0 || bar.Close <= 0 {
			continue
		}
		hl := math.Log(bar.High / bar.Low)
		co := math.Log(bar.Close / bar.Open)
		sum += 0.5*hl*hl - (2*math.Log(2)-1)*co*co
		n++
	}
	if n == 0 {
		return 0
	}
	return math.Sqrt(math.Max(0, sum/float64(n))) * math.Sqrt(tradingDaysPerYear)
}

func rogersSatchellVol(bars []models.Bar, index, window int) float64 {
	start := maxInt(0, index-window+1)
	var sum float64
	var n int
	for i := start; i <= index && i < len(bars); i++ {
		bar := bars[i]
		if bar.High <= 0 || bar.Low <= 0 || bar.Open <= 0 || bar.Close <= 0 {
			continue
		}
		sum += math.Log(bar.High/bar.Close)*math.Log(bar.High/bar.Open) +
			math.Log(bar.Low/bar.Close)*math.Log(bar.Low/bar.Open)
		n++
	}
	if n == 0 {
		return 0
	}
	return math.Sqrt(math.Max(0, sum/float64(n))) * math.Sqrt(tradingDaysPerYear)
}

func atrPct(bars []models.Bar, index, window int) float64 {
	if index < 0 || index >= len(bars) || bars[index].Close <= 0 {
		return 0
	}
	start := maxInt(0, index-window+1)
	var sum float64
	var n int
	for i := start; i <= index; i++ {
		tr := bars[i].High - bars[i].Low
		if i > 0 {
			tr = math.Max(tr, math.Abs(bars[i].High-bars[i-1].Close))
			tr = math.Max(tr, math.Abs(bars[i].Low-bars[i-1].Close))
		}
		if tr > 0 {
			sum += tr
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return (sum / float64(n)) / bars[index].Close
}

func smaClose(bars []models.Bar, index, period int) float64 {
	if period <= 0 || index-period+1 < 0 || index >= len(bars) {
		return 0
	}
	var sum float64
	for i := index - period + 1; i <= index; i++ {
		sum += bars[i].Close
	}
	return sum / float64(period)
}

func smaSlope(bars []models.Bar, index, period int) float64 {
	current := smaClose(bars, index, period)
	previous := smaClose(bars, index-1, period)
	if previous == 0 {
		return 0
	}
	return current/previous - 1
}

func baseMetadataFloat(output *backtest.StrategyOutput, keys ...string) float64 {
	if output == nil || output.Metadata == nil {
		return 0
	}
	for _, key := range keys {
		if v, ok := output.Metadata[key]; ok {
			return interfaceFloat(v)
		}
	}
	return 0
}

func regimeProbability(values map[string]float64, keys ...string) float64 {
	if values == nil {
		return 0
	}
	for _, key := range keys {
		for actual, value := range values {
			if strings.EqualFold(actual, key) {
				return value
			}
		}
	}
	return 0
}

func contextLogReturn(context map[string][]models.Bar, symbol string, eventTime time.Time, lag int) float64 {
	bars := trimContextBars(context, symbol, eventTime)
	if len(bars) == 0 {
		return 0
	}
	return logReturn(bars, len(bars)-1, lag)
}

func contextVol(context map[string][]models.Bar, symbol string, eventTime time.Time, window int) float64 {
	bars := trimContextBars(context, symbol, eventTime)
	if len(bars) == 0 {
		return 0
	}
	return closeToCloseVol(bars, len(bars)-1, window)
}

func betaToContext(bars []models.Bar, index int, context map[string][]models.Bar, symbol string, eventTime time.Time, window int) float64 {
	left := logReturnsEndingAt(bars, index, window)
	ctxBars := trimContextBars(context, symbol, eventTime)
	if len(ctxBars) == 0 {
		return 0
	}
	right := logReturnsEndingAt(ctxBars, len(ctxBars)-1, window)
	n := minInt(len(left), len(right))
	if n < 2 {
		return 0
	}
	left = left[len(left)-n:]
	right = right[len(right)-n:]
	meanLeft := mean(left)
	meanRight := mean(right)
	var covariance float64
	var variance float64
	for i := 0; i < n; i++ {
		l := left[i] - meanLeft
		r := right[i] - meanRight
		covariance += l * r
		variance += r * r
	}
	if variance == 0 {
		return 0
	}
	return covariance / variance
}

func trimContextBars(context map[string][]models.Bar, symbol string, eventTime time.Time) []models.Bar {
	if context == nil || symbol == "" {
		return nil
	}
	bars, ok := context[strings.ToUpper(symbol)]
	if !ok {
		bars = context[strings.ToLower(symbol)]
	}
	if len(bars) == 0 {
		return nil
	}
	idx := -1
	for i := range bars {
		if bars[i].Time.After(eventTime) {
			break
		}
		idx = i
	}
	if idx < 0 {
		return nil
	}
	return bars[:idx+1]
}

func volumeZ(bars []models.Bar, index, window int) float64 {
	values := make([]float64, 0, window)
	start := maxInt(0, index-window+1)
	for i := start; i <= index && i < len(bars); i++ {
		values = append(values, float64(bars[i].Volume))
	}
	return zscoreLast(values)
}

func dollarVolumeZ(bars []models.Bar, index, window int) float64 {
	values := make([]float64, 0, window)
	start := maxInt(0, index-window+1)
	for i := start; i <= index && i < len(bars); i++ {
		values = append(values, bars[i].Close*float64(bars[i].Volume))
	}
	return zscoreLast(values)
}

func signedVolumeImbalance(bars []models.Bar, index, window int) float64 {
	if index < 0 || index >= len(bars) {
		return 0
	}
	start := maxInt(0, index-window+1)
	var signed float64
	var total float64
	for i := start; i <= index; i++ {
		volume := float64(bars[i].Volume)
		if volume <= 0 {
			continue
		}
		direction := 0.0
		if bars[i].Close > bars[i].Open {
			direction = 1
		} else if bars[i].Close < bars[i].Open {
			direction = -1
		} else if i > 0 {
			if bars[i].Close > bars[i-1].Close {
				direction = 1
			} else if bars[i].Close < bars[i-1].Close {
				direction = -1
			}
		}
		signed += direction * volume
		total += volume
	}
	if total <= 0 {
		return 0
	}
	return signed / total
}

func amihudIlliquidity(bars []models.Bar, index, window int) float64 {
	if index <= 0 || index >= len(bars) {
		return 0
	}
	start := maxInt(1, index-window+1)
	var sum float64
	var n int
	for i := start; i <= index; i++ {
		dollarVolume := bars[i].Close * float64(bars[i].Volume)
		if dollarVolume <= 0 || bars[i-1].Close <= 0 || bars[i].Close <= 0 {
			continue
		}
		ret := math.Abs(math.Log(bars[i].Close / bars[i-1].Close))
		sum += ret / dollarVolume
		n++
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func highLowSpreadProxy(bars []models.Bar, index int) float64 {
	if index < 0 || index >= len(bars) {
		return 0
	}
	bar := bars[index]
	mid := (bar.High + bar.Low) / 2
	if mid <= 0 {
		return 0
	}
	return (bar.High - bar.Low) / mid
}

func turnoverProxy(bars []models.Bar, index, window int) float64 {
	if index < 0 || index >= len(bars) {
		return 0
	}
	start := maxInt(0, index-window+1)
	var sum float64
	var n int
	for i := start; i <= index; i++ {
		if bars[i].Volume > 0 {
			sum += float64(bars[i].Volume)
			n++
		}
	}
	if n == 0 || sum == 0 {
		return 0
	}
	avg := sum / float64(n)
	return float64(bars[index].Volume)/avg - 1
}

func logReturnsEndingAt(bars []models.Bar, index, window int) []float64 {
	if window <= 0 || index <= 0 || index >= len(bars) {
		return nil
	}
	start := maxInt(1, index-window+1)
	out := make([]float64, 0, index-start+1)
	for i := start; i <= index; i++ {
		if bars[i].Close <= 0 || bars[i-1].Close <= 0 {
			continue
		}
		out = append(out, math.Log(bars[i].Close/bars[i-1].Close))
	}
	return out
}

func annualizedStd(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	_, std := meanStd(values)
	return std * math.Sqrt(tradingDaysPerYear)
}

func meanStd(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	m := mean(values)
	var sum float64
	for _, v := range values {
		d := v - m
		sum += d * d
	}
	return m, math.Sqrt(sum / float64(len(values)))
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func zscoreLast(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	m, std := meanStd(values)
	if std == 0 {
		return 0
	}
	return (values[len(values)-1] - m) / std
}

func interfaceFloat(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	default:
		return 0
	}
}

func safeRatio(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

func cleanFinite(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
