package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

// RiskProfile defines position sizing and signal aggressiveness
type RiskProfile int

const (
	RiskProfileConservative RiskProfile = iota
	RiskProfileModerate
	RiskProfileAggressive
)

// String returns the profile name
func (rp RiskProfile) String() string {
	switch rp {
	case RiskProfileConservative:
		return "Conservative"
	case RiskProfileModerate:
		return "Moderate"
	case RiskProfileAggressive:
		return "Aggressive"
	default:
		return "Unknown"
	}
}

// SignalWeight defines how much each strategy contributes to ensemble decision
type SignalWeight struct {
	MACrossoverWeight float64
	KalmanWeight      float64
}

// RegimeConfiguration maps HMM regimes to signal weights and position sizing
type RegimeConfiguration struct {
	// Weights for each regime state
	LowVolTrendWeights   SignalWeight
	MediumWeights        SignalWeight
	HighVolStressWeights SignalWeight

	// Position sizing scalars (0.0-1.0) per regime
	LowVolTrendScalar   float64
	MediumScalar        float64
	HighVolStressScalar float64

	// Confidence thresholds for signal generation
	BuyThreshold  float64
	SellThreshold float64
}

// NewRegimeConfiguration returns production-grade defaults
func NewRegimeConfiguration() RegimeConfiguration {
	return RegimeConfiguration{
		// Low Vol Trend: favor momentum, suppress mean reversion
		LowVolTrendWeights: SignalWeight{
			MACrossoverWeight: 0.70,
			KalmanWeight:      0.30,
		},
		// Medium: balanced both strategies
		MediumWeights: SignalWeight{
			MACrossoverWeight: 0.50,
			KalmanWeight:      0.50,
		},
		// High Vol Stress: suppress all entries, favor mean reversion exits
		HighVolStressWeights: SignalWeight{
			MACrossoverWeight: 0.20,
			KalmanWeight:      0.80,
		},

		// Position sizing scalars
		LowVolTrendScalar:   1.0,  // Full allocation
		MediumScalar:        0.75, // 75% allocation
		HighVolStressScalar: 0.25, // 25% allocation (capital preservation)

		// Ensemble voting thresholds
		BuyThreshold:  0.5,  // At least 50% vote weight
		SellThreshold: -0.5, // At least 50% vote weight (negative)
	}
}

// PositionSizingRules defines how to scale positions by profile and regime
type PositionSizingRules struct {
	// Base position size (as % of available cash)
	ConservativeBaseSize float64 // e.g., 0.05 = 5% per position
	ModerateBaseSize     float64 // e.g., 0.10 = 10% per position
	AggressiveBaseSize   float64 // e.g., 0.20 = 20% per position

	// Regime multipliers
	LowVolTrendMultiplier   float64 // 1.0 = normal
	MediumMultiplier        float64 // 0.75 = reduced
	HighVolStressMultiplier float64 // 0.25 = minimal
}

// NewPositionSizingRules returns institutional defaults
func NewPositionSizingRules() PositionSizingRules {
	return PositionSizingRules{
		ConservativeBaseSize: 0.05,
		ModerateBaseSize:     0.10,
		AggressiveBaseSize:   0.20,

		LowVolTrendMultiplier:   1.0,
		MediumMultiplier:        0.75,
		HighVolStressMultiplier: 0.25,
	}
}

// EnsembleDecisionLayer aggregates signals from multiple strategies based on regime
type EnsembleDecisionLayer struct {
	maStrategy     *backtest.MACrossoverStrategy
	kalmanStrategy *backtest.KalmanStrategy
	hmmDetector    *HMMRegimeDetector
	regimeConfig   RegimeConfiguration
	positionSizing PositionSizingRules
	riskProfile    RiskProfile

	// State tracking
	lastRegime      MarketRegime
	lastRegimeBars  int
	lastSignalScore float64
}

// NewEnsembleDecisionLayer creates a new ensemble with all three engines
func NewEnsembleDecisionLayer(
	maStrategy *backtest.MACrossoverStrategy,
	kalmanStrategy *backtest.KalmanStrategy,
	hmmWindowSize int,
	riskProfile RiskProfile,
) *EnsembleDecisionLayer {
	if maStrategy == nil {
		maStrategy = backtest.NewMACrossoverStrategy(20, 50)
	}
	if kalmanStrategy == nil {
		kalmanStrategy = backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	}
	return &EnsembleDecisionLayer{
		maStrategy:     maStrategy,
		kalmanStrategy: kalmanStrategy,
		hmmDetector:    NewHMMRegimeDetector(hmmWindowSize),
		regimeConfig:   NewRegimeConfiguration(),
		positionSizing: NewPositionSizingRules(),
		riskProfile:    riskProfile,
		lastRegime:     RegimeMedium,
		lastRegimeBars: 0,
	}
}

// UpdateRiskProfile changes the position sizing profile
func (e *EnsembleDecisionLayer) UpdateRiskProfile(profile RiskProfile) {
	e.riskProfile = profile
	log.Printf("[Ensemble] Risk profile updated to %s", profile.String())
}

// EvaluateSignal runs all three engines and returns a composite decision
// Returns: (signal, confidence, regime, score, error)
func (e *EnsembleDecisionLayer) EvaluateSignal(
	ctx context.Context,
	bars []models.Bar,
) (models.Signal, float64, MarketRegime, float64, error) {
	if len(bars) < 50 {
		return models.SignalHold, 0.0, RegimeMedium, 0.0, fmt.Errorf("insufficient bars for ensemble: need 50, have %d", len(bars))
	}

	// Step 1: Detect regime
	regime, regimeConfidence, err := e.hmmDetector.Update(bars)
	if err != nil {
		return models.SignalHold, 0.0, RegimeMedium, 0.0, fmt.Errorf("HMM update failed: %w", err)
	}

	// Track regime persistence for logging
	if regime != e.lastRegime {
		e.lastRegimeBars = 1
		e.lastRegime = regime
	} else {
		e.lastRegimeBars++
	}

	// Step 2: Generate signals from both strategies
	maSignals, err := e.maStrategy.GenerateSignal(ctx, bars)
	if err != nil {
		return models.SignalHold, 0.0, regime, 0.0, fmt.Errorf("ma strategy failed: %w", err)
	}

	kalmanSignals, err := e.kalmanStrategy.GenerateSignal(ctx, bars)
	if err != nil {
		return models.SignalHold, 0.0, regime, 0.0, fmt.Errorf("kalman strategy failed: %w", err)
	}

	if len(maSignals) == 0 || len(kalmanSignals) == 0 {
		return models.SignalHold, 0.0, regime, 0.0, fmt.Errorf("no signals generated")
	}

	// Get latest signals
	latestMA := maSignals[len(maSignals)-1]
	latestKalman := kalmanSignals[len(kalmanSignals)-1]

	// Step 3: Apply regime-based weighting
	weights := e.getWeightsForRegime(regime)

	// Step 4: Ensemble voting
	signalScore := e.computeEnsembleScore(latestMA, latestKalman, weights)
	e.lastSignalScore = signalScore

	// Step 5: Regime-aware gating
	finalSignal := e.applyRegimeGating(signalScore, regime)

	// Step 6: Compute final confidence (account for regime + ensemble agreement)
	finalConfidence := e.computeFinalConfidence(signalScore, regimeConfidence)

	log.Printf("[Ensemble] Regime=%s(%.2f) MA=%v Kalman=%v Score=%.3f Signal=%v Confidence=%.2f",
		regime.String(), regimeConfidence,
		latestMA, latestKalman,
		signalScore, finalSignal, finalConfidence)

	return finalSignal, finalConfidence, regime, signalScore, nil
}

// getWeightsForRegime returns signal weights based on market regime
func (e *EnsembleDecisionLayer) getWeightsForRegime(regime MarketRegime) SignalWeight {
	switch regime {
	case RegimeLowVolTrend:
		return e.regimeConfig.LowVolTrendWeights
	case RegimeMedium:
		return e.regimeConfig.MediumWeights
	case RegimeHighVolStress:
		return e.regimeConfig.HighVolStressWeights
	default:
		return e.regimeConfig.MediumWeights
	}
}

// computeEnsembleScore converts signals into weighted vote (-1.0 to +1.0)
func (e *EnsembleDecisionLayer) computeEnsembleScore(
	maSignal models.Signal,
	kalmanSignal models.Signal,
	weights SignalWeight,
) float64 {
	// Convert signals to votes
	maVote := signalToVote(maSignal)
	kalmanVote := signalToVote(kalmanSignal)

	// Weighted average
	score := (maVote * weights.MACrossoverWeight) + (kalmanVote * weights.KalmanWeight)

	return score
}

// signalToVote converts Signal to numerical vote
func signalToVote(signal models.Signal) float64 {
	switch signal {
	case models.SignalBuy:
		return 1.0
	case models.SignalSell:
		return -1.0
	default:
		return 0.0
	}
}

// applyRegimeGating applies regime-aware signal filtering
func (e *EnsembleDecisionLayer) applyRegimeGating(score float64, regime MarketRegime) models.Signal {
	switch regime {
	case RegimeLowVolTrend:
		// Low vol trending: normal thresholds, favor momentum
		if score >= e.regimeConfig.BuyThreshold {
			return models.SignalBuy
		}
		if score <= e.regimeConfig.SellThreshold {
			return models.SignalSell
		}

	case RegimeMedium:
		// Medium: balanced
		if score >= e.regimeConfig.BuyThreshold {
			return models.SignalBuy
		}
		if score <= e.regimeConfig.SellThreshold {
			return models.SignalSell
		}

	case RegimeHighVolStress:
		// High vol stress: suppress buy signals, only allow mean-reversion sells
		if score <= e.regimeConfig.SellThreshold*0.8 { // Lower threshold for exits
			return models.SignalSell
		}
		// Suppress buy signals entirely in stress regime
		return models.SignalHold
	}

	return models.SignalHold
}

// computeFinalConfidence combines regime and ensemble agreement
func (e *EnsembleDecisionLayer) computeFinalConfidence(score float64, regimeConfidence float64) float64 {
	// Absolute score indicates agreement
	scoreConfidence := absoluteValue(score) * 0.7 // Score contrib: 70%
	regimeWeighting := regimeConfidence * 0.3     // Regime contrib: 30%

	return (scoreConfidence + regimeWeighting)
}

// GetPositionSize computes the position size based on profile, regime, and available cash
func (e *EnsembleDecisionLayer) GetPositionSize(
	availableCash float64,
	regime MarketRegime,
) float64 {
	baseSize := e.getBaseSize()
	regimeMultiplier := e.getRegimeMultiplier(regime)

	return availableCash * baseSize * regimeMultiplier
}

// getBaseSize returns position size % based on risk profile
func (e *EnsembleDecisionLayer) getBaseSize() float64 {
	switch e.riskProfile {
	case RiskProfileConservative:
		return e.positionSizing.ConservativeBaseSize
	case RiskProfileModerate:
		return e.positionSizing.ModerateBaseSize
	case RiskProfileAggressive:
		return e.positionSizing.AggressiveBaseSize
	default:
		return e.positionSizing.ModerateBaseSize
	}
}

// getRegimeMultiplier returns the scalar for current regime
func (e *EnsembleDecisionLayer) getRegimeMultiplier(regime MarketRegime) float64 {
	switch regime {
	case RegimeLowVolTrend:
		return e.positionSizing.LowVolTrendMultiplier
	case RegimeMedium:
		return e.positionSizing.MediumMultiplier
	case RegimeHighVolStress:
		return e.positionSizing.HighVolStressMultiplier
	default:
		return e.positionSizing.MediumMultiplier
	}
}

// GetRegimeInfo returns current regime and persistence
func (e *EnsembleDecisionLayer) GetRegimeInfo() (MarketRegime, int) {
	return e.lastRegime, e.lastRegimeBars
}

// GetStateProbabilities returns HMM state probabilities
func (e *EnsembleDecisionLayer) GetStateProbabilities() [3]float64 {
	return e.hmmDetector.GetProbabilities()
}

// GetLastSignalScore returns the last computed ensemble score
func (e *EnsembleDecisionLayer) GetLastSignalScore() float64 {
	return e.lastSignalScore
}

// Reset clears internal state (useful for backtesting)
func (e *EnsembleDecisionLayer) Reset() {
	e.hmmDetector.Reset()
	e.lastRegime = RegimeMedium
	e.lastRegimeBars = 0
	e.lastSignalScore = 0.0
}

// Helper function
func absoluteValue(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func (e *EnsembleDecisionLayer) EvaluateLatest(ctx context.Context, bars []models.Bar) (backtest.StrategyOutput, error) {
	signal, confidence, regime, score, err := e.EvaluateSignal(ctx, bars)
	if err != nil {
		return backtest.StrategyOutput{Signal: models.SignalHold}, err
	}

	probs := e.GetStateProbabilities()
	return backtest.StrategyOutput{
		Signal:          signal,
		PositionSizePct: e.getBaseSize() * e.getRegimeMultiplier(regime),
		RegimeLabel:     regime.String(),
		Metadata: map[string]interface{}{
			"confidence":         confidence,
			"score":              score,
			"probability_low":    probs[0],
			"probability_medium": probs[1],
			"probability_high":   probs[2],
		},
	}, nil
}

// GenerateSignals evaluates rolling history sequentially to support HMM backtesting
func (e *EnsembleDecisionLayer) GenerateSignals(ctx context.Context, bars []models.Bar) ([]backtest.StrategyOutput, error) {
	e.Reset()
	out := make([]backtest.StrategyOutput, len(bars))
	for i := range bars {
		if i < 50 { // Minimum lookback warmup boundary
			out[i] = backtest.StrategyOutput{Signal: models.SignalHold}
			continue
		}
		res, err := e.EvaluateLatest(ctx, bars[:i+1])
		if err != nil {
			return nil, err
		}
		out[i] = res
	}
	return out, nil
}
