package agent

import "github.com/oalpha/internal/agent/risk"

type MarketRegime = risk.MarketRegime
type RegimeDetector = risk.RegimeDetector
type HMMRegimeDetector = risk.HMMRegimeDetector
type RegimeRiskRole = risk.RegimeRiskRole
type RiskOverlayPolicy = risk.RiskOverlayPolicy
type RegimeOverlayInput = risk.RegimeOverlayInput
type RegimeOverlayDecision = risk.RegimeOverlayDecision
type RegimeRiskOverlay = risk.RegimeRiskOverlay
type ObservationEncoder = risk.ObservationEncoder

const (
	RegimeLowVolTrend   = risk.RegimeLowVolTrend
	RegimeMedium        = risk.RegimeMedium
	RegimeHighVolStress = risk.RegimeHighVolStress

	RegimeRiskLowVol  = risk.RegimeRiskLowVol
	RegimeRiskNormal  = risk.RegimeRiskNormal
	RegimeRiskHighVol = risk.RegimeRiskHighVol
	RegimeRiskCrisis  = risk.RegimeRiskCrisis
	RegimeRiskUnknown = risk.RegimeRiskUnknown
)

var (
	NewHMMRegimeDetector     = risk.NewHMMRegimeDetector
	NewObservationEncoder    = risk.NewObservationEncoder
	NewRegimeRiskOverlay     = risk.NewRegimeRiskOverlay
	DefaultRiskOverlayPolicy = risk.DefaultRiskOverlayPolicy
	RealizedVolatility       = risk.RealizedVolatility
	RollingTrend             = risk.RollingTrend
)
