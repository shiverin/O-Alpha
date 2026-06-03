package momentum

import (
	"math"
	"sort"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

type UniverseCandidate struct {
	Symbol             string  `json:"symbol"`
	Price              float64 `json:"price"`
	MedianDollarVolume float64 `json:"median_dollar_volume"`
	DataCompleteness   float64 `json:"data_completeness"`
	Eligible           bool    `json:"eligible"`
	Reason             string  `json:"reason,omitempty"`
}

func FilterUniverse(
	panel backtest.AlignedBars,
	index int,
	symbols []string,
	cfg CrossSectionalMomentumConfig,
) []string {
	candidates := EvaluateUniverse(panel, index, symbols, cfg)
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Eligible {
			out = append(out, candidate.Symbol)
		}
	}
	return out
}

func EvaluateUniverse(
	panel backtest.AlignedBars,
	index int,
	symbols []string,
	cfg CrossSectionalMomentumConfig,
) []UniverseCandidate {
	cfg = cfg.withDefaults()
	out := make([]UniverseCandidate, 0, len(symbols))
	for _, symbol := range normalizeSymbols(symbols) {
		bars := panel.Bars[symbol]
		candidate := UniverseCandidate{Symbol: symbol}
		if len(bars) <= index {
			candidate.Reason = "missing_bars"
			out = append(out, candidate)
			continue
		}
		candidate.Price = bars[index].Close
		if candidate.Price < cfg.MinPrice {
			candidate.Reason = "price_below_min"
			out = append(out, candidate)
			continue
		}
		start := maxInt(0, index-cfg.FormationDays-cfg.SkipDays+1)
		candidate.DataCompleteness = dataCompleteness(bars, start, index)
		if candidate.DataCompleteness < cfg.MinDataCompleteness {
			candidate.Reason = "insufficient_data_completeness"
			out = append(out, candidate)
			continue
		}
		candidate.MedianDollarVolume = medianDollarVolume(bars, start, index)
		if candidate.MedianDollarVolume < cfg.MinMedianDollarVolume {
			candidate.Reason = "insufficient_liquidity"
			out = append(out, candidate)
			continue
		}
		candidate.Eligible = true
		out = append(out, candidate)
	}
	return out
}

func dataCompleteness(bars []models.Bar, start, end int) float64 {
	if end < start || start < 0 || end >= len(bars) {
		return 0
	}
	var valid int
	for i := start; i <= end; i++ {
		if bars[i].Close > 0 && !math.IsNaN(bars[i].Close) && !math.IsInf(bars[i].Close, 0) {
			valid++
		}
	}
	return float64(valid) / float64(end-start+1)
}

func medianDollarVolume(bars []models.Bar, start, end int) float64 {
	values := make([]float64, 0, maxInt(0, end-start+1))
	for i := start; i <= end && i < len(bars); i++ {
		if bars[i].Close <= 0 || bars[i].Volume <= 0 {
			continue
		}
		values = append(values, bars[i].Close*float64(bars[i].Volume))
	}
	if len(values) == 0 {
		return 0
	}
	sort.Float64s(values)
	mid := len(values) / 2
	if len(values)%2 == 1 {
		return values[mid]
	}
	return (values[mid-1] + values[mid]) / 2
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
