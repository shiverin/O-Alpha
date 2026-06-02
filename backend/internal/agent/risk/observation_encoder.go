package risk

import (
	"fmt"
	"math"
	"sort"

	"github.com/oalpha/pkg/models"
)

const (
	hmmNumStates  = 3
	hmmNumSymbols = 9
)

// ObservationEncoder converts rolling bar windows into the shared 9-symbol HMM
// emission space: volatility bucket * 3 + trend bucket.
type ObservationEncoder struct {
	WindowSize   int        `json:"window_size"`
	VolBuckets   [3]float64 `json:"vol_buckets"`
	TrendBuckets [3]float64 `json:"trend_buckets"`
}

func NewObservationEncoder(windowSize int) ObservationEncoder {
	if windowSize < 2 {
		windowSize = 2
	}
	return ObservationEncoder{
		WindowSize:   windowSize,
		VolBuckets:   [3]float64{0.01, 0.025, 0.05},
		TrendBuckets: [3]float64{-0.005, 0.0, 0.005},
	}
}

// FitBuckets calibrates thresholds from training data only. Callers should not
// pass validation/test bars here.
func (e *ObservationEncoder) FitBuckets(bars []models.Bar) error {
	if e.WindowSize < 2 {
		return fmt.Errorf("window size must be at least 2")
	}
	if len(bars) < e.WindowSize {
		return fmt.Errorf("insufficient bars for bucket fitting: have %d, need %d", len(bars), e.WindowSize)
	}

	vols := make([]float64, 0, len(bars)-e.WindowSize+1)
	trends := make([]float64, 0, len(bars)-e.WindowSize+1)
	for i := e.WindowSize; i <= len(bars); i++ {
		window := bars[i-e.WindowSize : i]
		vols = append(vols, realizedVolatility(window))
		trends = append(trends, rollingTrend(window))
	}

	sort.Float64s(vols)
	sort.Float64s(trends)

	e.VolBuckets = [3]float64{
		percentile(vols, 25),
		percentile(vols, 50),
		percentile(vols, 75),
	}
	e.TrendBuckets = [3]float64{
		percentile(trends, 25),
		percentile(trends, 50),
		percentile(trends, 75),
	}
	return nil
}

func (e ObservationEncoder) EncodeWindow(window []models.Bar) int {
	volBucket := e.DiscretizeVolatility(realizedVolatility(window))
	trendBucket := e.DiscretizeTrend(rollingTrend(window))
	return ObservationSymbol(volBucket, trendBucket)
}

func (e ObservationEncoder) EncodeSequence(bars []models.Bar) []int {
	if e.WindowSize < 2 || len(bars) < e.WindowSize {
		return nil
	}

	observations := make([]int, 0, len(bars)-e.WindowSize+1)
	for i := e.WindowSize; i <= len(bars); i++ {
		observations = append(observations, e.EncodeWindow(bars[i-e.WindowSize:i]))
	}
	return observations
}

func (e ObservationEncoder) DiscretizeVolatility(vol float64) int {
	if vol < e.VolBuckets[0] {
		return 0
	}
	if vol < e.VolBuckets[1] {
		return 1
	}
	return 2
}

func (e ObservationEncoder) DiscretizeTrend(trend float64) int {
	if trend < e.TrendBuckets[1] {
		return 0
	}
	if trend < e.TrendBuckets[2] {
		return 1
	}
	return 2
}

func ObservationSymbol(volBucket, trendBucket int) int {
	if volBucket < 0 {
		volBucket = 0
	}
	if volBucket > 2 {
		volBucket = 2
	}
	if trendBucket < 0 {
		trendBucket = 0
	}
	if trendBucket > 2 {
		trendBucket = 2
	}
	return volBucket*3 + trendBucket
}

func volBucketOfSymbol(symbol int) int {
	if symbol < 0 {
		return 0
	}
	if symbol >= hmmNumSymbols {
		return 2
	}
	return symbol / 3
}

// percentile does linear interpolation on a slice that is already sorted ascending.
func percentile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}
	rank := (p / 100.0) * float64(n-1)
	lo := int(math.Floor(rank))
	hi := int(math.Ceil(rank))
	if lo == hi {
		return sorted[lo]
	}
	frac := rank - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

func RealizedVolatility(bars []models.Bar) float64 {
	return realizedVolatility(bars)
}

func realizedVolatility(bars []models.Bar) float64 {
	if len(bars) < 2 {
		return 0
	}

	var sumSquaredReturns float64
	var observations int
	for i := 1; i < len(bars); i++ {
		if bars[i-1].Close <= 0 || bars[i].Close <= 0 {
			continue
		}
		logReturn := math.Log(bars[i].Close / bars[i-1].Close)
		sumSquaredReturns += logReturn * logReturn
		observations++
	}
	if observations == 0 {
		return 0
	}

	variance := sumSquaredReturns / float64(observations)
	if variance < 0 {
		variance = 0
	}
	return math.Sqrt(variance)
}

func RollingTrend(bars []models.Bar) float64 {
	return rollingTrend(bars)
}

func rollingTrend(bars []models.Bar) float64 {
	if len(bars) < 2 {
		return 0
	}

	firstClose := bars[0].Close
	lastClose := bars[len(bars)-1].Close
	if firstClose <= 0 {
		return 0
	}
	return (lastClose - firstClose) / firstClose
}
