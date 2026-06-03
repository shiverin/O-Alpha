package agent

import (
	"context"
	"fmt"
	"log"
	"math"

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
	regimeDetector RegimeDetector
	riskOverlay    *RegimeRiskOverlay
	regimeMode     RegimeMode
	regimeConfig   RegimeConfiguration
	positionSizing PositionSizingRules
	riskProfile    RiskProfile

	// State tracking
	lastRegime      MarketRegime
	lastRegimeBars  int
	lastSignalScore float64
}

// NewEnsembleDecisionLayer creates a base MA+Kalman alpha ensemble with an
// HMM risk overlay. The HMM is not allowed to create directional alpha.
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
	hmmDetector := NewHMMRegimeDetector(hmmWindowSize)
	return &EnsembleDecisionLayer{
		maStrategy:     maStrategy,
		kalmanStrategy: kalmanStrategy,
		hmmDetector:    hmmDetector,
		regimeDetector: hmmDetector,
		riskOverlay:    NewRegimeRiskOverlay(DefaultRiskOverlayPolicy()),
		regimeMode:     RegimeModeOverlay,
		regimeConfig:   NewRegimeConfiguration(),
		positionSizing: NewPositionSizingRules(),
		riskProfile:    riskProfile,
		lastRegime:     RegimeMedium,
		lastRegimeBars: 0,
	}
}

func NewEnsembleDecisionLayerForMode(
	maStrategy *backtest.MACrossoverStrategy,
	kalmanStrategy *backtest.KalmanStrategy,
	hmmWindowSize int,
	riskProfile RiskProfile,
	mode RegimeMode,
) (*EnsembleDecisionLayer, error) {
	ensemble := NewEnsembleDecisionLayer(maStrategy, kalmanStrategy, hmmWindowSize, riskProfile)
	switch mode {
	case "", RegimeModeOverlay:
		ensemble.regimeMode = RegimeModeOverlay
	case RegimeModeNone:
		ensemble.regimeMode = RegimeModeNone
		ensemble.hmmDetector = nil
		ensemble.regimeDetector = nil
		ensemble.riskOverlay = nil
	default:
		return nil, fmt.Errorf("unsupported regime mode: %s", mode)
	}
	return ensemble, nil
}

// UpdateRiskProfile changes the position sizing profile
func (e *EnsembleDecisionLayer) UpdateRiskProfile(profile RiskProfile) {
	e.riskProfile = profile
	log.Printf("[Ensemble] Risk profile updated to %s", profile.String())
}

// EvaluateSignal runs the base MA+Kalman alpha engine and returns its
// directional decision. The HMM overlay is intentionally not an alpha source.
// Returns: (signal, confidence, regime, score, error)
func (e *EnsembleDecisionLayer) EvaluateSignal(
	ctx context.Context,
	bars []models.Bar,
) (models.Signal, float64, MarketRegime, float64, error) {
	if len(bars) < 50 {
		return models.SignalHold, 0.0, RegimeMedium, 0.0, fmt.Errorf("insufficient bars for ensemble: need 50, have %d", len(bars))
	}

	regime := RegimeMedium
	regimeConfidence := 1.0
	if e.regimeMode == RegimeModeOverlay && e.regimeDetector != nil {
		var err error
		regime, regimeConfidence, err = e.regimeDetector.Update(bars)
		if err != nil {
			regime = RegimeMedium
			regimeConfidence = 0
			log.Printf("[Ensemble] risk overlay detector update failed: %v", err)
		}
	}

	// Track regime persistence for logging
	if regime != e.lastRegime {
		e.lastRegimeBars = 1
		e.lastRegime = regime
	} else {
		e.lastRegimeBars++
	}

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

	latestMA := maSignals[len(maSignals)-1]
	latestKalman := kalmanSignals[len(kalmanSignals)-1]

	signalScore := e.computeEnsembleScore(latestMA, latestKalman, e.regimeConfig.MediumWeights)
	e.lastSignalScore = signalScore

	finalSignal := e.applyBaseThreshold(signalScore)

	finalConfidence := e.computeFinalConfidence(signalScore, regimeConfidence)

	log.Printf("[Ensemble] BaseAlpha regime=%s(%.2f) MA=%v Kalman=%v Score=%.3f Signal=%v Confidence=%.2f",
		regime.String(), regimeConfidence,
		latestMA, latestKalman,
		signalScore, finalSignal, finalConfidence)

	return finalSignal, finalConfidence, regime, signalScore, nil
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
	return e.applyBaseThreshold(score)
}

func (e *EnsembleDecisionLayer) applyBaseThreshold(score float64) models.Signal {
	if score >= e.regimeConfig.BuyThreshold {
		return models.SignalBuy
	}
	if score <= e.regimeConfig.SellThreshold {
		return models.SignalSell
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
	if e.regimeMode == RegimeModeNone {
		return 1.0
	}
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
	if e.regimeDetector == nil {
		return [3]float64{0, 1, 0}
	}
	return e.regimeDetector.GetProbabilities()
}

// GetLastSignalScore returns the last computed ensemble score
func (e *EnsembleDecisionLayer) GetLastSignalScore() float64 {
	return e.lastSignalScore
}

// Reset clears internal state (useful for backtesting)
func (e *EnsembleDecisionLayer) Reset() {
	if e.regimeDetector != nil {
		e.regimeDetector.Reset()
	}
	if e.riskOverlay != nil {
		e.riskOverlay.Reset()
	}
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
	basePositionSize := e.getBaseSize()
	adjustedPositionSize := basePositionSize
	var overlayDecision *RegimeOverlayDecision
	if e.regimeMode == RegimeModeOverlay && e.riskOverlay != nil {
		probSlice := []float64{probs[0], probs[1], probs[2]}
		baseExposure := 0.0
		if signal == models.SignalBuy {
			baseExposure = basePositionSize
		}
		decision := e.riskOverlay.Apply(RegimeOverlayInput{
			BaseExposure:      baseExposure,
			PosteriorProbs:    probSlice,
			StateRoles:        []RegimeRiskRole{RegimeRiskLowVol, RegimeRiskNormal, RegimeRiskHighVol},
			ModelHealthy:      true,
			RealizedAnnualVol: annualizedWindowVolatility(bars, e.hmmDetector),
		})
		overlayDecision = &decision
		if signal == models.SignalBuy {
			adjustedPositionSize = decision.AdjustedExposure
			if adjustedPositionSize <= 0 {
				signal = models.SignalHold
			}
		}
	}

	metadata := map[string]interface{}{
		"confidence":             confidence,
		"score":                  score,
		"regime_mode":            e.regimeMode.String(),
		"probability_low":        probs[0],
		"probability_medium":     probs[1],
		"probability_high":       probs[2],
		"base_position_size_pct": basePositionSize,
		"position_size_pct":      adjustedPositionSize,
	}
	if overlayDecision != nil {
		metadata["hmm_overlay_multiplier"] = overlayDecision.Multiplier
		metadata["hmm_overlay_raw_multiplier"] = overlayDecision.RawMultiplier
		metadata["hmm_overlay_role"] = string(overlayDecision.EffectiveRole)
		metadata["hmm_overlay_confidence"] = overlayDecision.Confidence
		metadata["hmm_overlay_reasons"] = overlayDecision.Reasons
		metadata["hmm_overlay_vetoed"] = overlayDecision.Vetoed
	}

	targetWeight := adjustedPositionSize
	if signal != models.SignalBuy {
		targetWeight = 0
	}
	return backtest.StrategyOutput{
		Signal:          signal,
		PositionSizePct: adjustedPositionSize,
		RegimeLabel:     regime.String(),
		Metadata:        metadata,
		AlphaScore:      score,
		Confidence:      confidence,
		TargetWeight:    targetWeight,
		Engine:          "hmm_ensemble",
	}, nil
}

func annualizedWindowVolatility(bars []models.Bar, detector *HMMRegimeDetector) float64 {
	if detector == nil || len(bars) < 2 {
		return 0
	}
	start := len(bars) - detector.WindowSize()
	if start < 0 {
		start = 0
	}
	return RealizedVolatility(bars[start:]) * math.Sqrt(252)
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
