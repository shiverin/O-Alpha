package ml

import (
	"math"
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

func TestFeatureBuilderDoesNotUseFutureBars(t *testing.T) {
	bars := testBars("VOO", 80, 100, 1)
	context := map[string][]models.Bar{
		"SPY": testBars("SPY", 80, 200, 0.5),
	}
	spec := FeatureSpec{
		Version: "test",
		Features: []string{
			"log_ret_5",
			"distance_to_20d_high",
			"spy_ret_1",
			"relative_strength_vs_spy_21",
			"gap_pct",
		},
	}
	builder := NewFeatureBuilder(spec)
	base := backtest.StrategyOutput{AlphaScore: 0.7, Confidence: 0.8}

	before, err := builder.BuildAt(FeatureBuildInput{
		Symbol:      "VOO",
		Bars:        bars,
		BaseOutput:  &base,
		ContextBars: context,
	}, 40)
	if err != nil {
		t.Fatalf("BuildAt before mutation: %v", err)
	}

	futureBars := append([]models.Bar(nil), bars...)
	futureContext := map[string][]models.Bar{"SPY": append([]models.Bar(nil), context["SPY"]...)}
	for i := 41; i < len(futureBars); i++ {
		futureBars[i].Close = 10_000 + float64(i)
		futureBars[i].High = futureBars[i].Close + 100
		futureBars[i].Low = futureBars[i].Close - 100
		futureContext["SPY"][i].Close = 20_000 + float64(i)
	}

	after, err := builder.BuildAt(FeatureBuildInput{
		Symbol:      "VOO",
		Bars:        futureBars,
		BaseOutput:  &base,
		ContextBars: futureContext,
	}, 40)
	if err != nil {
		t.Fatalf("BuildAt after mutation: %v", err)
	}

	if len(before.Values) != len(after.Values) {
		t.Fatalf("feature vector length changed")
	}
	for i := range before.Values {
		if math.Abs(before.Values[i]-after.Values[i]) > 1e-12 {
			t.Fatalf("feature %s used future data: before=%f after=%f", before.Names[i], before.Values[i], after.Values[i])
		}
	}
}

func TestFeatureBuilderIncludesBaseStrategyState(t *testing.T) {
	bars := testBars("VOO", 70, 100, 1)
	builder := NewFeatureBuilder(FeatureSpec{
		Version: "test",
		Features: []string{
			"ensemble_score",
			"ensemble_confidence",
			"kalman_zscore",
			"hmm_regime_probability_high",
		},
	})
	base := backtest.StrategyOutput{
		AlphaScore: 0.42,
		Confidence: 0.77,
		Metadata: map[string]interface{}{
			"kalman_zscore": -1.25,
		},
	}

	vector, err := builder.BuildAt(FeatureBuildInput{
		Symbol:     "VOO",
		Bars:       bars,
		BaseOutput: &base,
		HMMRegimeProbabilities: map[string]float64{
			"crisis": 0.33,
		},
	}, 60)
	if err != nil {
		t.Fatalf("BuildAt: %v", err)
	}

	want := []float64{0.42, 0.77, -1.25, 0.33}
	for i := range want {
		if math.Abs(vector.Values[i]-want[i]) > 1e-12 {
			t.Fatalf("feature %s = %f, want %f", vector.Names[i], vector.Values[i], want[i])
		}
	}
}

func TestFeatureBuilderIncludesStationaryAndMicrostructureFeatures(t *testing.T) {
	bars := testBars("VOO", 90, 100, 0.5)
	builder := NewFeatureBuilder(FeatureSpec{
		Version: "test",
		Features: []string{
			"fracdiff_close_d0_5",
			"fracdiff_log_close_d0_5",
			"order_book_imbalance",
			"signed_volume_imbalance_20",
			"amihud_illiquidity_20",
			"high_low_spread_proxy",
		},
	})
	base := backtest.StrategyOutput{
		Metadata: map[string]interface{}{
			"order_book_imbalance": 0.25,
		},
	}

	vector, err := builder.BuildAt(FeatureBuildInput{
		Symbol:     "VOO",
		Bars:       bars,
		BaseOutput: &base,
	}, 80)
	if err != nil {
		t.Fatalf("BuildAt: %v", err)
	}
	if len(vector.Values) != 6 {
		t.Fatalf("values len=%d, want 6", len(vector.Values))
	}
	if vector.Values[0] == 0 || vector.Values[1] == 0 {
		t.Fatalf("fractional diff features should be populated: %v", vector.Values[:2])
	}
	if math.Abs(vector.Values[2]-0.25) > 1e-12 {
		t.Fatalf("order book imbalance=%f, want 0.25", vector.Values[2])
	}
	for i, value := range vector.Values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("feature %s is not finite: %f", vector.Names[i], value)
		}
	}
}

func testBars(symbol string, n int, start, step float64) []models.Bar {
	baseTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := make([]models.Bar, n)
	for i := range bars {
		close := start + step*float64(i)
		open := close - step/2
		bars[i] = models.Bar{
			Time:   baseTime.AddDate(0, 0, i),
			Symbol: symbol,
			Open:   open,
			High:   close + 1,
			Low:    math.Max(0.01, close-1),
			Close:  close,
			Volume: int64(1_000_000 + i*1_000),
		}
	}
	return bars
}
