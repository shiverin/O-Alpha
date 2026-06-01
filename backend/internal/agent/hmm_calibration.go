package agent

import (
	"log"
	"math"
	"sort"

	"github.com/oalpha/pkg/models"
)

// calibratable is implemented by strategies whose thresholds can be recalibrated
// from rolling history. MA/Kalman don't implement it, so they're skipped automatically.
type calibratable interface {
	Calibrate(bars []models.Bar)
}

// Calibrate recomputes the HMM volatility/trend buckets from the actual
// distribution of this symbol+timeframe, fixing the "fixed 1%/2.5%/5%" scale trap.
func (e *EnsembleDecisionLayer) Calibrate(bars []models.Bar) {
	win := e.hmmDetector.windowSize
	if win < 2 || len(bars) <= win {
		return // not enough history; keep current buckets
	}

	vols := make([]float64, 0, len(bars)-win+1)
	trends := make([]float64, 0, len(bars)-win+1)
	for i := win; i <= len(bars); i++ {
		window := bars[i-win : i]
		vols = append(vols, e.hmmDetector.calculateRealizedVolatility(window))
		trends = append(trends, e.hmmDetector.calculateRollingTrend(window))
	}

	sort.Float64s(vols)
	sort.Float64s(trends)

	e.hmmDetector.UpdateBuckets(
		[3]float64{percentile(vols, 25), percentile(vols, 50), percentile(vols, 75)},
		[3]float64{percentile(trends, 25), percentile(trends, 50), percentile(trends, 75)},
	)
	log.Printf("[Calibration] HMM buckets updated: vol=%v trend=%v",
		e.hmmDetector.volatilityBuckets, e.hmmDetector.trendBuckets)
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
