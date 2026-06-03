package momentum

import (
	"fmt"
	"math"
	"sort"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

const tradingDaysPerYear = 252.0

type MomentumScore struct {
	Symbol          string  `json:"symbol"`
	Rank            int     `json:"rank"`
	Score           float64 `json:"score"`
	FormationReturn float64 `json:"formation_return"`
	RealizedVol     float64 `json:"realized_vol"`
}

func ComputeMomentumScores(
	panel backtest.AlignedBars,
	symbols []string,
	index int,
	cfg CrossSectionalMomentumConfig,
) ([]MomentumScore, error) {
	cfg = cfg.withDefaults()
	if index < 0 || index >= len(panel.Times) {
		return nil, fmt.Errorf("index %d outside panel length %d", index, len(panel.Times))
	}
	scores := make([]MomentumScore, 0, len(symbols))
	for _, symbol := range normalizeSymbols(symbols) {
		bars := panel.Bars[symbol]
		if len(bars) <= index {
			continue
		}
		score, ok := momentumScoreForSymbol(symbol, bars, index, cfg)
		if ok {
			scores = append(scores, score)
		}
	}
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score == scores[j].Score {
			return scores[i].Symbol < scores[j].Symbol
		}
		return scores[i].Score > scores[j].Score
	})
	for i := range scores {
		scores[i].Rank = i + 1
	}
	return scores, nil
}

func SelectTopMomentum(scores []MomentumScore, cfg CrossSectionalMomentumConfig) []MomentumScore {
	cfg = cfg.withDefaults()
	if len(scores) == 0 {
		return nil
	}
	count := int(math.Floor(float64(len(scores)) * cfg.TopFraction))
	if len(scores) >= cfg.MinPositions && count < cfg.MinPositions {
		count = cfg.MinPositions
	}
	if count < 1 {
		count = 1
	}
	if count > cfg.MaxPositions {
		count = cfg.MaxPositions
	}
	if count > len(scores) {
		count = len(scores)
	}
	return append([]MomentumScore(nil), scores[:count]...)
}

func momentumScoreForSymbol(symbol string, bars []models.Bar, index int, cfg CrossSectionalMomentumConfig) (MomentumScore, bool) {
	formationEnd := index - cfg.SkipDays
	formationStart := formationEnd - cfg.FormationDays
	if formationStart < 0 || formationEnd >= len(bars) || formationEnd <= formationStart {
		return MomentumScore{}, false
	}
	startClose := bars[formationStart].Close
	endClose := bars[formationEnd].Close
	if startClose <= 0 || endClose <= 0 {
		return MomentumScore{}, false
	}
	formationReturn := math.Log(endClose / startClose)
	returns := logReturnsBetween(bars, formationStart+1, formationEnd)
	vol := annualizedStd(returns)
	if vol <= 0 {
		return MomentumScore{}, false
	}
	return MomentumScore{
		Symbol:          symbol,
		Score:           formationReturn / vol,
		FormationReturn: formationReturn,
		RealizedVol:     vol,
	}, true
}

func logReturnsBetween(bars []models.Bar, start, end int) []float64 {
	if start < 1 {
		start = 1
	}
	if end >= len(bars) {
		end = len(bars) - 1
	}
	if end < start {
		return nil
	}
	out := make([]float64, 0, end-start+1)
	for i := start; i <= end; i++ {
		if bars[i].Close <= 0 || bars[i-1].Close <= 0 {
			continue
		}
		out = append(out, math.Log(bars[i].Close/bars[i-1].Close))
	}
	return out
}

func annualizedStd(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	mean := mean(values)
	var sum float64
	for _, value := range values {
		d := value - mean
		sum += d * d
	}
	return math.Sqrt(sum/float64(len(values))) * math.Sqrt(tradingDaysPerYear)
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
