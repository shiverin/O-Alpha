package agent

import (
	"log"

	"github.com/oalpha/pkg/models"
)

// calibratable is implemented by strategies whose thresholds can be recalibrated
// from rolling history. MA/Kalman don't implement it, so they're skipped automatically.
type calibratable interface {
	Calibrate(bars []models.Bar)
}

// Calibrate recomputes the overlay HMM volatility/trend buckets from the actual
// distribution of this symbol+timeframe, fixing the "fixed 1%/2.5%/5%" scale trap.
func (e *EnsembleDecisionLayer) Calibrate(bars []models.Bar) {
	if e.regimeMode == RegimeModeNone || e.hmmDetector == nil {
		return
	}
	encoder := NewObservationEncoder(e.hmmDetector.WindowSize())
	if len(bars) <= encoder.WindowSize {
		return // not enough history; keep current buckets
	}
	if err := encoder.FitBuckets(bars); err != nil {
		return
	}
	e.hmmDetector.UpdateBuckets(encoder.VolBuckets, encoder.TrendBuckets)
	log.Printf("[Calibration] overlay HMM buckets updated: vol=%v trend=%v",
		encoder.VolBuckets, encoder.TrendBuckets)
}
