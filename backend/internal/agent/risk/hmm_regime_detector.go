package risk

import (
	"fmt"
	"time"

	"github.com/oalpha/pkg/models"
)

// MarketRegime represents the current market state.
type MarketRegime int

const (
	RegimeLowVolTrend   MarketRegime = iota // State 0: low volatility, constructive trend
	RegimeMedium                            // State 1: normal/uncertain market state
	RegimeHighVolStress                     // State 2: high volatility or stress
)

func (mr MarketRegime) String() string {
	switch mr {
	case RegimeLowVolTrend:
		return "Low Vol Trend"
	case RegimeMedium:
		return "Medium"
	case RegimeHighVolStress:
		return "High Vol Stress"
	default:
		return "Unknown"
	}
}

// RegimeDetector exposes posterior risk-state estimates. It is intentionally
// used as a risk overlay input, not as an alpha source.
type RegimeDetector interface {
	Name() string
	Reset()
	Update(bars []models.Bar) (MarketRegime, float64, error)
	GetProbabilities() [3]float64
}

// HMMRegimeDetector is a fixed-parameter HMM risk posterior filter. It does not
// fit models and does not emit alpha signals.
type HMMRegimeDetector struct {
	name       string
	encoder    ObservationEncoder
	transition [3][3]float64
	emission   [3][9]float64

	stateSequence      []MarketRegime
	obsProbs           [][]float64
	volatilityBuckets  [3]float64
	trendBuckets       [3]float64
	stateProbabilities [3]float64

	windowSize      int
	minBarsRequired int

	lastProcessedTime time.Time
	cachedRegime      MarketRegime
	cachedConfidence  float64
}

func NewHMMRegimeDetector(windowSize int) *HMMRegimeDetector {
	encoder := NewObservationEncoder(windowSize)
	detector := &HMMRegimeDetector{
		name:               "risk_overlay_hmm",
		encoder:            encoder,
		transition:         riskOverlayTransitionMatrix(),
		emission:           riskOverlayEmissionMatrix(),
		stateSequence:      make([]MarketRegime, 0, 1000),
		obsProbs:           make([][]float64, 0, 1000),
		volatilityBuckets:  encoder.VolBuckets,
		trendBuckets:       encoder.TrendBuckets,
		stateProbabilities: [3]float64{0.30, 0.50, 0.20},
		windowSize:         encoder.WindowSize,
		minBarsRequired:    encoder.WindowSize,
		cachedRegime:       RegimeMedium,
	}
	return detector
}

func (hmm *HMMRegimeDetector) Name() string {
	if hmm.name == "" {
		return "risk_overlay_hmm"
	}
	return hmm.name
}

func riskOverlayTransitionMatrix() [3][3]float64 {
	return [3][3]float64{
		{0.82, 0.16, 0.02},
		{0.18, 0.66, 0.16},
		{0.03, 0.27, 0.70},
	}
}

func riskOverlayEmissionMatrix() [3][9]float64 {
	// Symbols are volatility bucket * 3 + trend bucket.
	return [3][9]float64{
		// Low-vol trend: low-vol up/neutral observations dominate.
		{0.03, 0.14, 0.46, 0.02, 0.09, 0.20, 0.01, 0.02, 0.03},
		// Normal: broad middle, intentionally less opinionated.
		{0.08, 0.12, 0.12, 0.10, 0.18, 0.14, 0.08, 0.10, 0.08},
		// High-vol stress: high volatility and downtrend dominate.
		{0.04, 0.03, 0.02, 0.10, 0.08, 0.04, 0.40, 0.20, 0.09},
	}
}

func (hmm *HMMRegimeDetector) UpdateBuckets(volatilityPercentiles [3]float64, trendPercentiles [3]float64) {
	hmm.volatilityBuckets = volatilityPercentiles
	hmm.trendBuckets = trendPercentiles
	hmm.encoder.VolBuckets = volatilityPercentiles
	hmm.encoder.TrendBuckets = trendPercentiles
}

func (hmm *HMMRegimeDetector) WindowSize() int {
	return hmm.windowSize
}

func (hmm *HMMRegimeDetector) Buckets() ([3]float64, [3]float64) {
	return hmm.volatilityBuckets, hmm.trendBuckets
}

func (hmm *HMMRegimeDetector) DiscretizeVolatility(vol float64) int {
	return hmm.discretizeVolatility(vol)
}

func (hmm *HMMRegimeDetector) discretizeVolatility(vol float64) int {
	hmm.encoder.VolBuckets = hmm.volatilityBuckets
	return hmm.encoder.DiscretizeVolatility(vol)
}

func (hmm *HMMRegimeDetector) DiscretizeTrend(trend float64) int {
	return hmm.discretizeTrend(trend)
}

func (hmm *HMMRegimeDetector) discretizeTrend(trend float64) int {
	hmm.encoder.TrendBuckets = hmm.trendBuckets
	return hmm.encoder.DiscretizeTrend(trend)
}

func (hmm *HMMRegimeDetector) Update(bars []models.Bar) (MarketRegime, float64, error) {
	if len(bars) < hmm.minBarsRequired {
		return RegimeMedium, 0.0, fmt.Errorf("insufficient bars: have %d, need %d", len(bars), hmm.minBarsRequired)
	}

	latestBar := bars[len(bars)-1]
	if !latestBar.Time.IsZero() && !hmm.lastProcessedTime.IsZero() && !latestBar.Time.After(hmm.lastProcessedTime) {
		return hmm.cachedRegime, hmm.cachedConfidence, nil
	}

	startIdx := len(bars) - hmm.windowSize
	if startIdx < 0 {
		startIdx = 0
	}
	windowBars := bars[startIdx:]

	volatilityBucket := hmm.discretizeVolatility(hmm.calculateRealizedVolatility(windowBars))
	trendBucket := hmm.discretizeTrend(hmm.calculateRollingTrend(windowBars))
	regime, confidence := hmm.forwardFilterStep(volatilityBucket, trendBucket)

	if !latestBar.Time.IsZero() {
		hmm.lastProcessedTime = latestBar.Time
	}
	hmm.cachedRegime = regime
	hmm.cachedConfidence = confidence
	return regime, confidence, nil
}

func (hmm *HMMRegimeDetector) CalculateRealizedVolatility(bars []models.Bar) float64 {
	return hmm.calculateRealizedVolatility(bars)
}

func (hmm *HMMRegimeDetector) calculateRealizedVolatility(bars []models.Bar) float64 {
	return realizedVolatility(bars)
}

func (hmm *HMMRegimeDetector) CalculateRollingTrend(bars []models.Bar) float64 {
	return hmm.calculateRollingTrend(bars)
}

func (hmm *HMMRegimeDetector) calculateRollingTrend(bars []models.Bar) float64 {
	return rollingTrend(bars)
}

func (hmm *HMMRegimeDetector) forwardFilterStep(volBucket, trendBucket int) (MarketRegime, float64) {
	symbol := ObservationSymbol(volBucket, trendBucket)
	var predicted [3]float64
	for next := 0; next < hmmNumStates; next++ {
		for prev := 0; prev < hmmNumStates; prev++ {
			predicted[next] += hmm.stateProbabilities[prev] * hmm.transition[prev][next]
		}
	}

	var total float64
	for state := 0; state < hmmNumStates; state++ {
		predicted[state] *= hmm.emission[state][symbol]
		total += predicted[state]
	}
	if total <= 0 {
		hmm.stateProbabilities = [3]float64{0.30, 0.50, 0.20}
	} else {
		for state := 0; state < hmmNumStates; state++ {
			hmm.stateProbabilities[state] = predicted[state] / total
		}
	}

	regime, confidence := mostLikelyRiskRegime(hmm.stateProbabilities)
	hmm.stateSequence = append(hmm.stateSequence, regime)
	hmm.obsProbs = append(hmm.obsProbs, hmm.stateProbabilities[:])
	return regime, confidence
}

func (hmm *HMMRegimeDetector) GetRegimeSequence() []MarketRegime {
	return hmm.stateSequence
}

func (hmm *HMMRegimeDetector) GetProbabilities() [3]float64 {
	return hmm.stateProbabilities
}

func (hmm *HMMRegimeDetector) GetRegimePersistence(regime MarketRegime) int {
	if len(hmm.stateSequence) == 0 {
		return 0
	}
	count := 0
	for i := len(hmm.stateSequence) - 1; i >= 0; i-- {
		if hmm.stateSequence[i] == regime {
			count++
			continue
		}
		break
	}
	return count
}

func (hmm *HMMRegimeDetector) Reset() {
	hmm.stateSequence = hmm.stateSequence[:0]
	hmm.obsProbs = hmm.obsProbs[:0]
	hmm.stateProbabilities = [3]float64{0.30, 0.50, 0.20}
	hmm.lastProcessedTime = time.Time{}
	hmm.cachedRegime = RegimeMedium
	hmm.cachedConfidence = 0
}

func mostLikelyRiskRegime(probs [3]float64) (MarketRegime, float64) {
	maxState := 0
	maxProb := probs[0]
	for i := 1; i < hmmNumStates; i++ {
		if probs[i] > maxProb {
			maxState = i
			maxProb = probs[i]
		}
	}
	return MarketRegime(maxState), maxProb
}
