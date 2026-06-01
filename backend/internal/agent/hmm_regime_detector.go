package agent

import (
	"fmt"
	"math"
	"time"

	"github.com/oalpha/pkg/models"
)

// MarketRegime represents the current market state
type MarketRegime int

const (
	RegimeLowVolTrend   MarketRegime = iota // State 0: Low volatility, trending
	RegimeMedium                            // State 1: Medium volatility, neutral
	RegimeHighVolStress                     // State 2: High volatility, stress
)

// String returns the regime name
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

// HMMRegimeDetector implements a Hidden Markov Model for regime detection
// States:
//
//	0: Low Volatility + Trending (bull)
//	1: Medium Volatility + Neutral (range)
//	2: High Volatility + Stress (bear)
type HMMRegimeDetector struct {
	// State transition matrix [current_state][next_state]
	transitionMatrix [3][3]float64

	// Emission probabilities for observable: volatility [state][vol_bucket]
	volatilityEmissionMatrix [3][3]float64

	// Emission probabilities for observable: trend [state][trend_bucket]
	trendEmissionMatrix [3][3]float64

	// Hidden state sequence (for Viterbi decoding)
	stateSequence []MarketRegime

	// Observation probabilities (smoothed)
	obsProbs [][]float64

	// Volatility buckets: [low, medium, high]
	volatilityBuckets [3]float64

	// Trend buckets: [down, neutral, up]
	trendBuckets [3]float64

	// Current state probability distribution
	stateProbabilities [3]float64

	// Window size for rolling statistics
	windowSize int

	// Minimum bars required before generating regime signal
	minBarsRequired int

	lastProcessedTime time.Time
	cachedRegime      MarketRegime
	cachedConfidence  float64
}

// NewHMMRegimeDetector creates a new HMM detector with institutional defaults
func NewHMMRegimeDetector(windowSize int) *HMMRegimeDetector {
	detector := &HMMRegimeDetector{
		windowSize:      windowSize,
		minBarsRequired: windowSize,
		stateSequence:   make([]MarketRegime, 0, 1000),
		obsProbs:        make([][]float64, 0, 1000),
		stateProbabilities: [3]float64{
			0.33, // Initial equal probability
			0.34,
			0.33,
		},
	}

	// Transition probabilities: regime persistence with drift possibility
	detector.transitionMatrix = [3][3]float64{
		{0.80, 0.15, 0.05}, // From LowVolTrend: likely stay, drift to Medium, rare jump to HighVol
		{0.25, 0.50, 0.25}, // From Medium: symmetric to all states
		{0.10, 0.30, 0.60}, // From HighVolStress: likely stay, possible drift to Medium
	}

	// Volatility emissions (realized volatility percentile)
	// Low state emits low volatility
	detector.volatilityEmissionMatrix = [3][3]float64{
		{0.70, 0.25, 0.05}, // LowVolTrend: mostly low vol, some medium, rare high
		{0.20, 0.60, 0.20}, // Medium: balanced
		{0.05, 0.25, 0.70}, // HighVolStress: mostly high vol, some medium, rare low
	}

	// Trend emissions (rolling momentum)
	// Low state emits uptrend
	detector.trendEmissionMatrix = [3][3]float64{
		{0.10, 0.20, 0.70}, // LowVolTrend: strong uptrend
		{0.33, 0.34, 0.33}, // Medium: balanced
		{0.60, 0.30, 0.10}, // HighVolStress: downtrend
	}

	// Default bucket thresholds (will be updated based on rolling percentiles)
	detector.volatilityBuckets = [3]float64{0.01, 0.025, 0.05}
	detector.trendBuckets = [3]float64{-0.005, 0.0, 0.005}

	return detector
}

// ObservationKey combines volatility and trend into a single observable
type ObservationKey int

const (
	ObsLowVolDown ObservationKey = iota
	ObsLowVolNeutral
	ObsLowVolUp
	ObsMedVolDown
	ObsMedVolNeutral
	ObsMedVolUp
	ObsHighVolDown
	ObsHighVolNeutral
	ObsHighVolUp
)

// UpdateBuckets recalibrates volatility and trend thresholds based on recent statistics
func (hmm *HMMRegimeDetector) UpdateBuckets(volatilityPercentiles [3]float64, trendPercentiles [3]float64) {
	// volatilityPercentiles should be [25th, 50th, 75th] percentile of realized vol
	// trendPercentiles should be [25th, 50th, 75th] percentile of returns
	hmm.volatilityBuckets = [3]float64{
		volatilityPercentiles[0], // Low threshold
		volatilityPercentiles[1], // Medium threshold
		volatilityPercentiles[2], // High threshold
	}
	hmm.trendBuckets = [3]float64{
		trendPercentiles[0],
		trendPercentiles[1],
		trendPercentiles[2],
	}
}

// discretizeVolatility converts continuous volatility to bucket index
func (hmm *HMMRegimeDetector) discretizeVolatility(vol float64) int {
	if vol < hmm.volatilityBuckets[0] {
		return 0 // Low
	} else if vol < hmm.volatilityBuckets[1] {
		return 1 // Medium
	}
	return 2 // High
}

// discretizeTrend converts continuous trend to bucket index
func (hmm *HMMRegimeDetector) discretizeTrend(trend float64) int {
	if trend < hmm.trendBuckets[1] {
		return 0 // Down
	} else if trend < hmm.trendBuckets[2] {
		return 1 // Neutral
	}
	return 2 // Up
}

// Update processes a new bar and updates regime estimate
// bars: full bar history up to current bar
// Returns: (current regime, confidence [0,1], error)
func (hmm *HMMRegimeDetector) Update(bars []models.Bar) (MarketRegime, float64, error) {
	if len(bars) < hmm.minBarsRequired {
		return RegimeMedium, 0.0, fmt.Errorf("insufficient bars: have %d, need %d", len(bars), hmm.minBarsRequired)
	}

	latestBar := bars[len(bars)-1]
	if !latestBar.Time.IsZero() && !hmm.lastProcessedTime.IsZero() && !latestBar.Time.After(hmm.lastProcessedTime) {
		return hmm.cachedRegime, hmm.cachedConfidence, nil
	}

	// Extract rolling window
	startIdx := len(bars) - hmm.windowSize
	if startIdx < 0 {
		startIdx = 0
	}
	windowBars := bars[startIdx:]

	// Calculate volatility: realized volatility over window
	volatility := hmm.calculateRealizedVolatility(windowBars)
	volatilityBucket := hmm.discretizeVolatility(volatility)

	// Calculate trend: rolling momentum (return-based)
	trend := hmm.calculateRollingTrend(windowBars)
	trendBucket := hmm.discretizeTrend(trend)

	// Forward filter pass with observation likelihood
	regime, confidence := hmm.forwardFilterStep(volatilityBucket, trendBucket)

	if !latestBar.Time.IsZero() {
		hmm.lastProcessedTime = latestBar.Time
	}
	hmm.cachedRegime = regime
	hmm.cachedConfidence = confidence

	return regime, confidence, nil
}

// calculateRealizedVolatility computes realized volatility from intrabar returns
func (hmm *HMMRegimeDetector) calculateRealizedVolatility(bars []models.Bar) float64 {
	if len(bars) < 2 {
		return 0.0
	}

	var sumSquaredReturns float64
	for i := 1; i < len(bars); i++ {
		if bars[i-1].Close <= 0 {
			continue
		}
		logReturn := math.Log(bars[i].Close / bars[i-1].Close)
		sumSquaredReturns += logReturn * logReturn
	}

	variance := sumSquaredReturns / float64(len(bars)-1)
	if variance < 0 {
		variance = 0
	}
	return math.Sqrt(variance)
}

// calculateRollingTrend computes momentum as average return over window
func (hmm *HMMRegimeDetector) calculateRollingTrend(bars []models.Bar) float64 {
	if len(bars) < 2 {
		return 0.0
	}

	firstClose := bars[0].Close
	lastClose := bars[len(bars)-1].Close
	if firstClose <= 0 {
		return 0.0
	}

	// Simple return ratio
	return (lastClose - firstClose) / firstClose
}

func (hmm *HMMRegimeDetector) forwardFilterStep(volBucket, trendBucket int) (MarketRegime, float64) {
	// Observation likelihood: combine volatility and trend emissions
	observationLikelihoods := [3]float64{
		hmm.volatilityEmissionMatrix[0][volBucket] * hmm.trendEmissionMatrix[0][trendBucket],
		hmm.volatilityEmissionMatrix[1][volBucket] * hmm.trendEmissionMatrix[1][trendBucket],
		hmm.volatilityEmissionMatrix[2][volBucket] * hmm.trendEmissionMatrix[2][trendBucket],
	}

	// Update state probabilities using Bayes rule with transition matrix
	newStateProbs := [3]float64{0, 0, 0}
	var totalProb float64

	for currentState := 0; currentState < 3; currentState++ {
		// Sum over all possible previous states
		for prevState := 0; prevState < 3; prevState++ {
			transitionProb := hmm.transitionMatrix[prevState][currentState]
			newStateProbs[currentState] += hmm.stateProbabilities[prevState] * transitionProb
		}
		// Multiply by observation likelihood
		newStateProbs[currentState] *= observationLikelihoods[currentState]
		totalProb += newStateProbs[currentState]
	}

	// Normalize
	if totalProb > 0 {
		for i := 0; i < 3; i++ {
			newStateProbs[i] /= totalProb
		}
	}

	hmm.stateProbabilities = newStateProbs

	// Find most likely state
	maxProb := newStateProbs[0]
	maxState := 0
	for i := 1; i < 3; i++ {
		if newStateProbs[i] > maxProb {
			maxProb = newStateProbs[i]
			maxState = i
		}
	}

	hmm.stateSequence = append(hmm.stateSequence, MarketRegime(maxState))
	hmm.obsProbs = append(hmm.obsProbs, newStateProbs[:])

	return MarketRegime(maxState), maxProb
}

// GetRegimeSequence returns the detected regime sequence
func (hmm *HMMRegimeDetector) GetRegimeSequence() []MarketRegime {
	return hmm.stateSequence
}

// GetProbabilities returns current state probability distribution
func (hmm *HMMRegimeDetector) GetProbabilities() [3]float64 {
	return hmm.stateProbabilities
}

// GetRegimePersistence returns how many consecutive bars in current regime
func (hmm *HMMRegimeDetector) GetRegimePersistence(regime MarketRegime) int {
	if len(hmm.stateSequence) == 0 {
		return 0
	}

	count := 0
	for i := len(hmm.stateSequence) - 1; i >= 0; i-- {
		if hmm.stateSequence[i] == regime {
			count++
		} else {
			break
		}
	}
	return count
}

// Reset clears the state sequence (useful for backtesting)
func (hmm *HMMRegimeDetector) Reset() {
	hmm.stateSequence = hmm.stateSequence[:0]
	hmm.obsProbs = hmm.obsProbs[:0]
	hmm.stateProbabilities = [3]float64{0.33, 0.34, 0.33}
	hmm.lastProcessedTime = time.Time{}
	hmm.cachedRegime = RegimeMedium
	hmm.cachedConfidence = 0
}
