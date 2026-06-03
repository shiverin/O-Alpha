package momentum

import (
	"math"
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

func TestComputeMomentumScoresExcludesSkipWindow(t *testing.T) {
	panel := testMomentumPanel(80, map[string]func(int) float64{
		"AAA": func(i int) float64 { return 100 + float64(i)*0.4 + math.Sin(float64(i))*0.1 },
		"BBB": func(i int) float64 {
			price := 100 + math.Sin(float64(i))*0.2
			if i > 65 {
				price += 50
			}
			return price
		},
	})
	cfg := testMomentumConfig()
	cfg.FormationDays = 20
	cfg.SkipDays = 5

	scores, err := ComputeMomentumScores(panel, []string{"AAA", "BBB"}, 70, cfg)
	if err != nil {
		t.Fatalf("ComputeMomentumScores: %v", err)
	}
	if len(scores) < 2 {
		t.Fatalf("scores=%d, want both symbols", len(scores))
	}
	if scores[0].Symbol != "AAA" {
		t.Fatalf("top symbol=%s, want AAA because BBB's jump is inside skipped window", scores[0].Symbol)
	}
}

func TestSelectTopMomentumRespectsMinAndMaxPositions(t *testing.T) {
	scores := []MomentumScore{
		{Symbol: "A", Score: 5},
		{Symbol: "B", Score: 4},
		{Symbol: "C", Score: 3},
		{Symbol: "D", Score: 2},
	}
	cfg := testMomentumConfig()
	cfg.TopFraction = 0.10
	cfg.MinPositions = 2
	cfg.MaxPositions = 3

	selected := SelectTopMomentum(scores, cfg)
	if len(selected) != 2 {
		t.Fatalf("selected=%d, want min positions 2", len(selected))
	}
}

func testMomentumConfig() CrossSectionalMomentumConfig {
	cfg := DefaultCrossSectionalMomentumConfig()
	cfg.FormationDays = 20
	cfg.SkipDays = 5
	cfg.MinPositions = 1
	cfg.MaxPositions = 3
	cfg.VolLookbackDays = 10
	cfg.MinPrice = 1
	cfg.MinMedianDollarVolume = 0
	cfg.MinDataCompleteness = 0.90
	return cfg
}

func testMomentumPanel(n int, priceFns map[string]func(int) float64) backtest.AlignedBars {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	symbols := make([]string, 0, len(priceFns))
	bars := make(map[string][]models.Bar, len(priceFns))
	for symbol, priceFn := range priceFns {
		symbols = append(symbols, symbol)
		series := make([]models.Bar, n)
		for i := range series {
			close := priceFn(i)
			series[i] = models.Bar{
				Time:   start.AddDate(0, 0, i),
				Symbol: symbol,
				Open:   close,
				High:   close + 1,
				Low:    math.Max(0.01, close-1),
				Close:  close,
				Volume: 1_000_000,
			}
		}
		bars[symbol] = series
	}
	times := make([]time.Time, n)
	for i := range times {
		times[i] = start.AddDate(0, 0, i)
	}
	return backtest.AlignedBars{
		Times:   times,
		Symbols: symbols,
		Bars:    bars,
	}
}
