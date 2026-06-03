package momentum

import (
	"math"

	"github.com/oalpha/internal/backtest"
)

type WeightedMomentumTarget struct {
	Symbol          string
	Weight          float64
	Volatility      float64
	Rank            int
	Score           float64
	FormationReturn float64
	VolScalar       float64
}

func BuildMomentumTargets(
	panel backtest.AlignedBars,
	index int,
	selected []MomentumScore,
	cfg CrossSectionalMomentumConfig,
	sectorBySymbol map[string]string,
) (map[string]backtest.TargetPosition, map[string]interface{}) {
	cfg = cfg.withDefaults()
	targets := make(map[string]backtest.TargetPosition)
	if len(selected) == 0 {
		return targets, map[string]interface{}{
			"engine":         StrategyName,
			"rebalance":      true,
			"selected_count": 0,
			"reason":         "no_selected_symbols",
		}
	}

	raw := inverseVolWeights(panel, index, selected, cfg)
	capped := applySymbolAndSectorCaps(raw, selected, cfg, sectorBySymbol)
	portfolioVol := estimatePortfolioVol(panel, index, capped, cfg.VolLookbackDays)
	volScalar := 1.0
	if portfolioVol > 0 && cfg.TargetVolAnnual > 0 {
		volScalar = math.Min(1.0, cfg.TargetVolAnnual/portfolioVol)
	}

	for _, score := range selected {
		weight := capped[score.Symbol] * volScalar
		if weight <= 0 {
			continue
		}
		targets[score.Symbol] = backtest.TargetPosition{
			Symbol:       score.Symbol,
			TargetWeight: weight,
			AlphaScore:   clip(score.Score, -3, 3),
			Confidence:   confidenceFromRank(score.Rank, len(selected)),
			Side:         backtest.PositionSideLong,
			Engine:       StrategyName,
			Metadata: map[string]interface{}{
				"rank":             score.Rank,
				"score":            score.Score,
				"formation_return": score.FormationReturn,
				"realized_vol":     score.RealizedVol,
				"vol_scalar":       volScalar,
				"rebalance":        true,
			},
		}
	}

	return targets, map[string]interface{}{
		"engine":               StrategyName,
		"rebalance":            true,
		"selected_count":       len(selected),
		"target_count":         len(targets),
		"portfolio_vol_annual": portfolioVol,
		"vol_scalar":           volScalar,
		"gross_exposure":       sumAbsWeights(targets),
	}
}

func inverseVolWeights(
	panel backtest.AlignedBars,
	index int,
	selected []MomentumScore,
	cfg CrossSectionalMomentumConfig,
) map[string]float64 {
	invVols := make(map[string]float64, len(selected))
	var totalInvVol float64
	for _, score := range selected {
		vol := trailingVolForSymbol(panel, score.Symbol, index, cfg.VolLookbackDays)
		if vol <= 0 {
			vol = score.RealizedVol
		}
		if vol <= 0 {
			continue
		}
		inv := 1 / vol
		invVols[score.Symbol] = inv
		totalInvVol += inv
	}
	weights := make(map[string]float64, len(invVols))
	if totalInvVol <= 0 {
		return weights
	}
	for symbol, inv := range invVols {
		weights[symbol] = inv / totalInvVol
	}
	return weights
}

func applySymbolAndSectorCaps(
	weights map[string]float64,
	selected []MomentumScore,
	cfg CrossSectionalMomentumConfig,
	sectorBySymbol map[string]string,
) map[string]float64 {
	out := make(map[string]float64, len(weights))
	sectorWeights := make(map[string]float64)
	for _, score := range selected {
		weight := weights[score.Symbol]
		if weight <= 0 {
			continue
		}
		weight = math.Min(weight, cfg.MaxSymbolWeight)
		sector := sectorBySymbol[score.Symbol]
		if sector != "" && cfg.MaxSectorWeight > 0 {
			remaining := cfg.MaxSectorWeight - sectorWeights[sector]
			if remaining <= 0 {
				continue
			}
			weight = math.Min(weight, remaining)
			sectorWeights[sector] += weight
		}
		out[score.Symbol] = weight
	}
	return out
}

func trailingVolForSymbol(panel backtest.AlignedBars, symbol string, index, lookback int) float64 {
	bars := panel.Bars[symbol]
	if len(bars) == 0 || index >= len(bars) {
		return 0
	}
	start := maxInt(1, index-lookback+1)
	return annualizedStd(logReturnsBetween(bars, start, index))
}

func estimatePortfolioVol(panel backtest.AlignedBars, index int, weights map[string]float64, lookback int) float64 {
	if len(weights) == 0 || index <= 0 {
		return 0
	}
	start := maxInt(1, index-lookback+1)
	returns := make([]float64, 0, index-start+1)
	for i := start; i <= index; i++ {
		var portfolioReturn float64
		var activeWeight float64
		for symbol, weight := range weights {
			bars := panel.Bars[symbol]
			if len(bars) <= i || bars[i].Close <= 0 || bars[i-1].Close <= 0 {
				continue
			}
			portfolioReturn += weight * math.Log(bars[i].Close/bars[i-1].Close)
			activeWeight += math.Abs(weight)
		}
		if activeWeight > 0 {
			returns = append(returns, portfolioReturn)
		}
	}
	return annualizedStd(returns)
}

func confidenceFromRank(rank, count int) float64 {
	if count <= 0 || rank <= 0 {
		return 0
	}
	return math.Max(0, 1-float64(rank-1)/float64(count))
}

func sumAbsWeights(targets map[string]backtest.TargetPosition) float64 {
	var total float64
	for _, target := range targets {
		total += math.Abs(target.TargetWeight)
	}
	return total
}

func clip(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
