package ensemble

import (
	"math"

	"github.com/oalpha/internal/backtest"
)

const tradingDaysPerYear = 252.0

type PortfolioRiskConfig struct {
	TargetVolAnnual     float64
	VolLookbackDays     int
	MaxGrossExposure    float64
	MaxNetExposure      float64
	MaxSymbolWeight     float64
	MaxSectorWeight     float64
	MaxDrawdownSoftStop float64
	MaxDrawdownHardStop float64
	FloorVolAnnual      float64
}

type PortfolioRiskState struct {
	PortfolioReturns []float64
	PeakEquity       float64
	CurrentEquity    float64
}

type RiskOverlayResult struct {
	Targets       map[string]backtest.TargetPosition
	RiskScalar    float64
	RealizedVol   float64
	Drawdown      float64
	Reasons       []string
	GrossExposure float64
	NetExposure   float64
}

func DefaultPortfolioRiskConfig() PortfolioRiskConfig {
	return PortfolioRiskConfig{
		TargetVolAnnual:     0.12,
		VolLookbackDays:     60,
		MaxGrossExposure:    1.0,
		MaxNetExposure:      1.0,
		MaxSymbolWeight:     0.15,
		MaxSectorWeight:     0.35,
		MaxDrawdownSoftStop: 0.10,
		MaxDrawdownHardStop: 0.20,
		FloorVolAnnual:      0.02,
	}
}

func ApplyPortfolioRiskOverlay(
	targets map[string]backtest.TargetPosition,
	cfg PortfolioRiskConfig,
	state PortfolioRiskState,
	sectorBySymbol map[string]string,
) RiskOverlayResult {
	cfg = cfg.withDefaults()
	out := cloneTargets(targets)
	realizedVol := realizedPortfolioVol(state.PortfolioReturns, cfg.VolLookbackDays)
	floorVol := cfg.FloorVolAnnual
	if floorVol <= 0 {
		floorVol = 0.02
	}
	denom := math.Max(realizedVol, floorVol)
	riskScalar := 1.0
	reasons := make([]string, 0)
	if cfg.TargetVolAnnual > 0 && denom > 0 {
		riskScalar = math.Min(1.0, cfg.TargetVolAnnual/denom)
		if riskScalar < 1 {
			reasons = append(reasons, "realized_vol_target")
		}
	}

	drawdown := currentDrawdown(state)
	if cfg.MaxDrawdownHardStop > 0 && drawdown >= cfg.MaxDrawdownHardStop {
		riskScalar = 0
		reasons = append(reasons, "hard_drawdown_stop")
	} else if cfg.MaxDrawdownSoftStop > 0 && drawdown >= cfg.MaxDrawdownSoftStop {
		riskScalar *= 0.5
		reasons = append(reasons, "soft_drawdown_stop")
	}

	scaleTargets(out, riskScalar)
	applySymbolCaps(out, cfg.MaxSymbolWeight)
	applySectorCaps(out, sectorBySymbol, cfg.MaxSectorWeight)
	applyGrossNetCaps(out, cfg.MaxGrossExposure, cfg.MaxNetExposure)

	return RiskOverlayResult{
		Targets:       out,
		RiskScalar:    riskScalar,
		RealizedVol:   realizedVol,
		Drawdown:      drawdown,
		Reasons:       reasons,
		GrossExposure: grossExposure(out),
		NetExposure:   netExposure(out),
	}
}

func realizedPortfolioVol(returns []float64, lookback int) float64 {
	if lookback <= 0 {
		lookback = 60
	}
	if len(returns) > lookback {
		returns = returns[len(returns)-lookback:]
	}
	if len(returns) < 2 {
		return 0
	}
	mean := mean(returns)
	var sum float64
	for _, r := range returns {
		d := r - mean
		sum += d * d
	}
	return math.Sqrt(sum/float64(len(returns))) * math.Sqrt(tradingDaysPerYear)
}

func currentDrawdown(state PortfolioRiskState) float64 {
	if state.PeakEquity <= 0 || state.CurrentEquity <= 0 || state.CurrentEquity >= state.PeakEquity {
		return 0
	}
	return 1 - state.CurrentEquity/state.PeakEquity
}

func scaleTargets(targets map[string]backtest.TargetPosition, scalar float64) {
	for symbol, target := range targets {
		target.TargetWeight *= scalar
		target.Metadata = cloneMetadata(target.Metadata)
		target.Metadata["risk_scalar"] = scalar
		targets[symbol] = target
	}
}

func applySymbolCaps(targets map[string]backtest.TargetPosition, maxWeight float64) {
	if maxWeight <= 0 {
		return
	}
	for symbol, target := range targets {
		if math.Abs(target.TargetWeight) <= maxWeight {
			continue
		}
		if target.TargetWeight > 0 {
			target.TargetWeight = maxWeight
		} else {
			target.TargetWeight = -maxWeight
		}
		targets[symbol] = target
	}
}

func applySectorCaps(targets map[string]backtest.TargetPosition, sectorBySymbol map[string]string, maxSectorWeight float64) {
	if maxSectorWeight <= 0 || len(sectorBySymbol) == 0 {
		return
	}
	sectorWeights := make(map[string]float64)
	for symbol, target := range targets {
		sector := sectorBySymbol[symbol]
		if sector == "" {
			continue
		}
		remaining := maxSectorWeight - sectorWeights[sector]
		if remaining <= 0 {
			target.TargetWeight = 0
			targets[symbol] = target
			continue
		}
		absWeight := math.Abs(target.TargetWeight)
		if absWeight > remaining {
			target.TargetWeight = math.Copysign(remaining, target.TargetWeight)
			absWeight = remaining
		}
		sectorWeights[sector] += absWeight
		targets[symbol] = target
	}
}

func applyGrossNetCaps(targets map[string]backtest.TargetPosition, maxGross, maxNet float64) {
	if maxGross > 0 {
		gross := grossExposure(targets)
		if gross > maxGross {
			scaleTargets(targets, maxGross/gross)
		}
	}
	if maxNet > 0 {
		net := math.Abs(netExposure(targets))
		if net > maxNet {
			scaleTargets(targets, maxNet/net)
		}
	}
}

func grossExposure(targets map[string]backtest.TargetPosition) float64 {
	var gross float64
	for _, target := range targets {
		gross += math.Abs(target.TargetWeight)
	}
	return gross
}

func netExposure(targets map[string]backtest.TargetPosition) float64 {
	var net float64
	for _, target := range targets {
		net += target.TargetWeight
	}
	return net
}

func cloneTargets(targets map[string]backtest.TargetPosition) map[string]backtest.TargetPosition {
	out := make(map[string]backtest.TargetPosition, len(targets))
	for symbol, target := range targets {
		target.Metadata = cloneMetadata(target.Metadata)
		out[symbol] = target
	}
	return out
}

func cloneMetadata(metadata map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(metadata)+2)
	for key, value := range metadata {
		out[key] = value
	}
	return out
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func (c PortfolioRiskConfig) withDefaults() PortfolioRiskConfig {
	defaults := DefaultPortfolioRiskConfig()
	if c.TargetVolAnnual <= 0 {
		c.TargetVolAnnual = defaults.TargetVolAnnual
	}
	if c.VolLookbackDays <= 0 {
		c.VolLookbackDays = defaults.VolLookbackDays
	}
	if c.MaxGrossExposure <= 0 {
		c.MaxGrossExposure = defaults.MaxGrossExposure
	}
	if c.MaxNetExposure <= 0 {
		c.MaxNetExposure = defaults.MaxNetExposure
	}
	if c.MaxSymbolWeight <= 0 {
		c.MaxSymbolWeight = defaults.MaxSymbolWeight
	}
	if c.MaxSectorWeight <= 0 {
		c.MaxSectorWeight = defaults.MaxSectorWeight
	}
	if c.FloorVolAnnual <= 0 {
		c.FloorVolAnnual = defaults.FloorVolAnnual
	}
	return c
}
