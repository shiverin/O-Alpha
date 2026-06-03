package ml

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/oalpha/pkg/models"
)

var DailyRankerFeatureNames = []string{
	"log_ret_1",
	"log_ret_5",
	"log_ret_10",
	"log_ret_21",
	"log_ret_63",
	"log_ret_126",
	"log_ret_252",
	"excess_log_ret_5",
	"excess_log_ret_21",
	"excess_log_ret_63",
	"excess_log_ret_126",
	"vol_20",
	"vol_63",
	"downside_vol_20",
	"return_to_vol_21",
	"return_to_vol_63",
	"beta_63",
	"corr_63",
	"residual_log_ret_21",
	"residual_log_ret_63",
	"distance_to_63d_high",
	"distance_to_63d_low",
	"ma_20_50",
	"ma_50_200",
	"volume_z_20",
	"dollar_volume_z_20",
	"amihud_20",
	"gap_pct",
	"intraday_ret",
	"benchmark_log_ret_21",
	"benchmark_vol_20",
}

type DailyRankerFeatureBuilder struct {
	Benchmark string
}

func NewDailyRankerFeatureBuilder(benchmark string) *DailyRankerFeatureBuilder {
	benchmark = strings.ToUpper(strings.TrimSpace(benchmark))
	if benchmark == "" {
		benchmark = "VOO"
	}
	return &DailyRankerFeatureBuilder{Benchmark: benchmark}
}

func (b *DailyRankerFeatureBuilder) BuildAtTime(
	barsBySymbol map[string][]models.Bar,
	symbol string,
	eventTime time.Time,
) (FeatureVector, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return FeatureVector{}, fmt.Errorf("symbol is required")
	}
	symbolBars := barsBySymbol[symbol]
	if len(symbolBars) == 0 {
		return FeatureVector{}, fmt.Errorf("bars missing for symbol %s", symbol)
	}
	index := rankerBarIndexAt(symbolBars, eventTime)
	if index < 0 {
		return FeatureVector{}, fmt.Errorf("time %s missing for symbol %s", eventTime.Format(time.RFC3339), symbol)
	}
	return b.BuildAt(barsBySymbol, symbol, index)
}

func (b *DailyRankerFeatureBuilder) BuildAt(
	barsBySymbol map[string][]models.Bar,
	symbol string,
	index int,
) (FeatureVector, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return FeatureVector{}, fmt.Errorf("symbol is required")
	}
	symbolBars := barsBySymbol[symbol]
	if len(symbolBars) == 0 {
		return FeatureVector{}, fmt.Errorf("bars missing for symbol %s", symbol)
	}
	if index < 0 || index >= len(symbolBars) {
		return FeatureVector{}, fmt.Errorf("index %d outside bars length %d for %s", index, len(symbolBars), symbol)
	}
	benchmark := b.Benchmark
	if benchmark == "" {
		benchmark = "VOO"
	}
	benchmarkBars := barsBySymbol[benchmark]
	if len(benchmarkBars) == 0 {
		return FeatureVector{}, fmt.Errorf("benchmark bars missing for %s", benchmark)
	}
	benchmarkIndex := rankerBarIndexAt(benchmarkBars, symbolBars[index].Time)
	if benchmarkIndex < 0 {
		return FeatureVector{}, fmt.Errorf("benchmark %s missing time %s", benchmark, symbolBars[index].Time.Format(time.RFC3339))
	}

	values := make([]float64, len(DailyRankerFeatureNames))
	for i, name := range DailyRankerFeatureNames {
		values[i] = cleanFinite(rankerFeatureValue(name, symbolBars, index, benchmarkBars, benchmarkIndex))
	}
	return FeatureVector{
		Symbol: symbol,
		Time:   symbolBars[index].Time,
		Names:  append([]string(nil), DailyRankerFeatureNames...),
		Values: values,
		Metadata: map[string]interface{}{
			"feature_spec_version": "daily_ranker_v1",
			"point_in_time_index":  index,
			"benchmark":            benchmark,
		},
	}, nil
}

func rankerFeatureValue(name string, bars []models.Bar, index int, benchmarkBars []models.Bar, benchmarkIndex int) float64 {
	logRet21 := rankerLogReturn(bars, index, 21)
	logRet63 := rankerLogReturn(bars, index, 63)
	benchmarkLogRet21 := rankerLogReturn(benchmarkBars, benchmarkIndex, 21)
	benchmarkLogRet63 := rankerLogReturn(benchmarkBars, benchmarkIndex, 63)
	beta63 := rankerBeta(bars, index, benchmarkBars, benchmarkIndex, 63, 20)
	vol20 := rankerRollingStd(rankerLogReturns(bars, index, 20), 5) * math.Sqrt(tradingDaysPerYear)
	vol63 := rankerRollingStd(rankerLogReturns(bars, index, 63), 20) * math.Sqrt(tradingDaysPerYear)

	switch name {
	case "log_ret_1":
		return rankerLogReturn(bars, index, 1)
	case "log_ret_5":
		return rankerLogReturn(bars, index, 5)
	case "log_ret_10":
		return rankerLogReturn(bars, index, 10)
	case "log_ret_21":
		return logRet21
	case "log_ret_63":
		return logRet63
	case "log_ret_126":
		return rankerLogReturn(bars, index, 126)
	case "log_ret_252":
		return rankerLogReturn(bars, index, 252)
	case "excess_log_ret_5":
		return rankerLogReturn(bars, index, 5) - rankerLogReturn(benchmarkBars, benchmarkIndex, 5)
	case "excess_log_ret_21":
		return logRet21 - benchmarkLogRet21
	case "excess_log_ret_63":
		return logRet63 - benchmarkLogRet63
	case "excess_log_ret_126":
		return rankerLogReturn(bars, index, 126) - rankerLogReturn(benchmarkBars, benchmarkIndex, 126)
	case "vol_20":
		return vol20
	case "vol_63":
		return vol63
	case "downside_vol_20":
		return rankerDownsideVol(bars, index, 20, 5)
	case "return_to_vol_21":
		return rankerSafeRatio(logRet21, vol20)
	case "return_to_vol_63":
		return rankerSafeRatio(logRet63, vol63)
	case "beta_63":
		return beta63
	case "corr_63":
		return rankerCorr(bars, index, benchmarkBars, benchmarkIndex, 63, 20)
	case "residual_log_ret_21":
		return logRet21 - beta63*benchmarkLogRet21
	case "residual_log_ret_63":
		return logRet63 - beta63*benchmarkLogRet63
	case "distance_to_63d_high":
		return rankerDistanceToHigh(bars, index, 63, 20)
	case "distance_to_63d_low":
		return rankerDistanceToLow(bars, index, 63, 20)
	case "ma_20_50":
		return rankerRatioMinusOne(rankerSMA(bars, index, 20, 5), rankerSMA(bars, index, 50, 20))
	case "ma_50_200":
		return rankerRatioMinusOne(rankerSMA(bars, index, 50, 20), rankerSMA(bars, index, 200, 60))
	case "volume_z_20":
		return rankerVolumeZ(bars, index, 20, 5)
	case "dollar_volume_z_20":
		return rankerDollarVolumeZ(bars, index, 20, 5)
	case "amihud_20":
		return rankerAmihud(bars, index, 20, 5)
	case "gap_pct":
		if index == 0 || bars[index-1].Close <= 0 {
			return 0
		}
		return bars[index].Open/bars[index-1].Close - 1
	case "intraday_ret":
		if bars[index].Open <= 0 {
			return 0
		}
		return bars[index].Close/bars[index].Open - 1
	case "benchmark_log_ret_21":
		return benchmarkLogRet21
	case "benchmark_vol_20":
		return rankerRollingStd(rankerLogReturns(benchmarkBars, benchmarkIndex, 20), 5) * math.Sqrt(tradingDaysPerYear)
	default:
		return 0
	}
}

func rankerLogReturn(bars []models.Bar, index, lookback int) float64 {
	if lookback <= 0 || index-lookback < 0 || index >= len(bars) {
		return 0
	}
	start := bars[index-lookback].Close
	end := bars[index].Close
	if start <= 0 || end <= 0 {
		return 0
	}
	return math.Log(end / start)
}

func rankerLogReturns(bars []models.Bar, index, window int) []float64 {
	if window <= 0 || index <= 0 || index >= len(bars) {
		return nil
	}
	start := index - window + 1
	if start < 1 {
		start = 1
	}
	values := make([]float64, 0, index-start+1)
	for i := start; i <= index; i++ {
		if bars[i-1].Close <= 0 || bars[i].Close <= 0 {
			continue
		}
		values = append(values, math.Log(bars[i].Close/bars[i-1].Close))
	}
	return values
}

func rankerRollingStd(values []float64, minPeriods int) float64 {
	if len(values) < minPeriods || len(values) < 2 {
		return 0
	}
	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))
	sumSq := 0.0
	for _, value := range values {
		diff := value - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(values)-1))
}

func rankerDownsideVol(bars []models.Bar, index, window, minPeriods int) float64 {
	values := rankerLogReturns(bars, index, window)
	negative := make([]float64, 0, len(values))
	for _, value := range values {
		if value < 0 {
			negative = append(negative, value)
		}
	}
	return rankerRollingStd(negative, minPeriods) * math.Sqrt(tradingDaysPerYear)
}

func rankerBeta(bars []models.Bar, index int, benchmarkBars []models.Bar, benchmarkIndex, window, minPeriods int) float64 {
	x, y := rankerAlignedReturns(bars, index, benchmarkBars, benchmarkIndex, window)
	if len(x) < minPeriods || len(x) < 2 {
		return 0
	}
	_, varianceY := rankerCovVar(x, y)
	if varianceY == 0 {
		return 0
	}
	covariance, _ := rankerCovVar(x, y)
	return covariance / varianceY
}

func rankerCorr(bars []models.Bar, index int, benchmarkBars []models.Bar, benchmarkIndex, window, minPeriods int) float64 {
	x, y := rankerAlignedReturns(bars, index, benchmarkBars, benchmarkIndex, window)
	if len(x) < minPeriods || len(x) < 2 {
		return 0
	}
	covariance, varianceY := rankerCovVar(x, y)
	varianceX := rankerSampleVariance(x)
	if varianceX <= 0 || varianceY <= 0 {
		return 0
	}
	return covariance / math.Sqrt(varianceX*varianceY)
}

func rankerAlignedReturns(bars []models.Bar, index int, benchmarkBars []models.Bar, benchmarkIndex, window int) ([]float64, []float64) {
	valuesX := make([]float64, 0, window)
	valuesY := make([]float64, 0, window)
	for offset := window - 1; offset >= 0; offset-- {
		i := index - offset
		j := benchmarkIndex - offset
		if i <= 0 || j <= 0 || i >= len(bars) || j >= len(benchmarkBars) {
			continue
		}
		if bars[i-1].Close <= 0 || bars[i].Close <= 0 || benchmarkBars[j-1].Close <= 0 || benchmarkBars[j].Close <= 0 {
			continue
		}
		valuesX = append(valuesX, math.Log(bars[i].Close/bars[i-1].Close))
		valuesY = append(valuesY, math.Log(benchmarkBars[j].Close/benchmarkBars[j-1].Close))
	}
	return valuesX, valuesY
}

func rankerCovVar(x, y []float64) (float64, float64) {
	if len(x) != len(y) || len(x) < 2 {
		return 0, 0
	}
	meanX := rankerMean(x)
	meanY := rankerMean(y)
	var cov float64
	for i := range x {
		cov += (x[i] - meanX) * (y[i] - meanY)
	}
	return cov / float64(len(x)-1), rankerSampleVariance(y)
}

func rankerSampleVariance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	mean := rankerMean(values)
	var sumSq float64
	for _, value := range values {
		diff := value - mean
		sumSq += diff * diff
	}
	return sumSq / float64(len(values)-1)
}

func rankerDistanceToHigh(bars []models.Bar, index, window, minPeriods int) float64 {
	closes := rankerCloseWindow(bars, index, window)
	if len(closes) < minPeriods {
		return 0
	}
	high := closes[0]
	for _, closePrice := range closes[1:] {
		if closePrice > high {
			high = closePrice
		}
	}
	if high <= 0 {
		return 0
	}
	return bars[index].Close/high - 1
}

func rankerDistanceToLow(bars []models.Bar, index, window, minPeriods int) float64 {
	closes := rankerCloseWindow(bars, index, window)
	if len(closes) < minPeriods {
		return 0
	}
	low := closes[0]
	for _, closePrice := range closes[1:] {
		if closePrice < low {
			low = closePrice
		}
	}
	if low <= 0 {
		return 0
	}
	return bars[index].Close/low - 1
}

func rankerSMA(bars []models.Bar, index, window, minPeriods int) float64 {
	closes := rankerCloseWindow(bars, index, window)
	if len(closes) < minPeriods {
		return 0
	}
	return rankerMean(closes)
}

func rankerCloseWindow(bars []models.Bar, index, window int) []float64 {
	if window <= 0 || index < 0 || index >= len(bars) {
		return nil
	}
	start := index - window + 1
	if start < 0 {
		start = 0
	}
	values := make([]float64, 0, index-start+1)
	for i := start; i <= index; i++ {
		if bars[i].Close > 0 {
			values = append(values, bars[i].Close)
		}
	}
	return values
}

func rankerVolumeZ(bars []models.Bar, index, window, minPeriods int) float64 {
	values := make([]float64, 0, window)
	start := index - window + 1
	if start < 0 {
		start = 0
	}
	for i := start; i <= index && i < len(bars); i++ {
		values = append(values, float64(bars[i].Volume))
	}
	if len(values) < minPeriods {
		return 0
	}
	std := rankerRollingStd(values, minPeriods)
	if std == 0 {
		return 0
	}
	return (float64(bars[index].Volume) - rankerMean(values)) / std
}

func rankerDollarVolumeZ(bars []models.Bar, index, window, minPeriods int) float64 {
	values := make([]float64, 0, window)
	start := index - window + 1
	if start < 0 {
		start = 0
	}
	for i := start; i <= index && i < len(bars); i++ {
		dollarVolume := bars[i].Close * float64(bars[i].Volume)
		if dollarVolume <= 0 {
			continue
		}
		values = append(values, math.Log(dollarVolume))
	}
	if len(values) < minPeriods {
		return 0
	}
	std := rankerRollingStd(values, minPeriods)
	if std == 0 {
		return 0
	}
	currentDollarVolume := bars[index].Close * float64(bars[index].Volume)
	if currentDollarVolume <= 0 {
		return 0
	}
	return (math.Log(currentDollarVolume) - rankerMean(values)) / std
}

func rankerAmihud(bars []models.Bar, index, window, minPeriods int) float64 {
	values := make([]float64, 0, window)
	start := index - window + 1
	if start < 1 {
		start = 1
	}
	for i := start; i <= index && i < len(bars); i++ {
		dollarVolume := bars[i].Close * float64(bars[i].Volume)
		if dollarVolume <= 0 || bars[i-1].Close <= 0 || bars[i].Close <= 0 {
			continue
		}
		values = append(values, math.Abs(bars[i].Close/bars[i-1].Close-1)/dollarVolume)
	}
	if len(values) < minPeriods {
		return 0
	}
	return rankerMean(values)
}

func rankerMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func rankerSafeRatio(numerator, denominator float64) float64 {
	if denominator == 0 || math.IsNaN(denominator) || math.IsInf(denominator, 0) {
		return 0
	}
	return numerator / denominator
}

func rankerRatioMinusOne(numerator, denominator float64) float64 {
	if denominator == 0 || math.IsNaN(denominator) || math.IsInf(denominator, 0) {
		return 0
	}
	return numerator/denominator - 1
}

func rankerBarIndexAt(bars []models.Bar, t time.Time) int {
	t = t.UTC()
	i := sort.Search(len(bars), func(i int) bool {
		return !bars[i].Time.UTC().Before(t)
	})
	if i < len(bars) && bars[i].Time.UTC().Equal(t) {
		return i
	}
	return -1
}
