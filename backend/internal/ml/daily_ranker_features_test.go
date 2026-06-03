package ml

import (
	"math"
	"testing"
	"time"

	"github.com/oalpha/pkg/models"
)

func TestDailyRankerFeatureBuilderDoesNotUseFutureBars(t *testing.T) {
	bars := map[string][]models.Bar{
		"VOO":  dailyRankerTestBars("VOO", 320, 100, 0.0010),
		"AAPL": dailyRankerTestBars("AAPL", 320, 80, 0.0015),
	}
	builder := NewDailyRankerFeatureBuilder("VOO")

	before, err := builder.BuildAt(bars, "AAPL", 260)
	if err != nil {
		t.Fatalf("BuildAt before mutation: %v", err)
	}

	mutated := map[string][]models.Bar{
		"VOO":  append([]models.Bar(nil), bars["VOO"]...),
		"AAPL": append([]models.Bar(nil), bars["AAPL"]...),
	}
	for i := 261; i < len(mutated["AAPL"]); i++ {
		mutated["AAPL"][i].Close = 10_000 + float64(i)
		mutated["AAPL"][i].Volume = 1
		mutated["VOO"][i].Close = 20_000 + float64(i)
	}

	after, err := builder.BuildAt(mutated, "AAPL", 260)
	if err != nil {
		t.Fatalf("BuildAt after mutation: %v", err)
	}
	for i := range before.Values {
		if math.Abs(before.Values[i]-after.Values[i]) > 1e-12 {
			t.Fatalf("feature %s used future data: before=%f after=%f", before.Names[i], before.Values[i], after.Values[i])
		}
	}
}

func TestDailyRankerFeatureBuilderComputesCoreFeatures(t *testing.T) {
	bars := map[string][]models.Bar{
		"VOO":  dailyRankerTestBars("VOO", 320, 100, 0.0010),
		"MSFT": dailyRankerTestBars("MSFT", 320, 120, 0.0020),
	}
	builder := NewDailyRankerFeatureBuilder("VOO")

	vector, err := builder.BuildAt(bars, "MSFT", 260)
	if err != nil {
		t.Fatalf("BuildAt: %v", err)
	}
	values := dailyRankerFeatureMap(vector)
	if math.Abs(values["log_ret_21"]-0.042) > 1e-12 {
		t.Fatalf("log_ret_21=%f, want 0.042", values["log_ret_21"])
	}
	if math.Abs(values["excess_log_ret_21"]-0.021) > 1e-12 {
		t.Fatalf("excess_log_ret_21=%f, want 0.021", values["excess_log_ret_21"])
	}
	if values["vol_20"] > 1e-10 {
		t.Fatalf("constant log returns should have near-zero vol_20, got %f", values["vol_20"])
	}
	if math.Abs(values["benchmark_log_ret_21"]-0.021) > 1e-12 {
		t.Fatalf("benchmark_log_ret_21=%f, want 0.021", values["benchmark_log_ret_21"])
	}
}

func dailyRankerFeatureMap(vector FeatureVector) map[string]float64 {
	out := make(map[string]float64, len(vector.Names))
	for i, name := range vector.Names {
		out[name] = vector.Values[i]
	}
	return out
}

func dailyRankerTestBars(symbol string, n int, start, logStep float64) []models.Bar {
	startTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := make([]models.Bar, n)
	for i := range bars {
		closePrice := start * math.Exp(logStep*float64(i))
		bars[i] = models.Bar{
			Time:   startTime.AddDate(0, 0, i),
			Symbol: symbol,
			Open:   closePrice * 0.999,
			High:   closePrice * 1.01,
			Low:    closePrice * 0.99,
			Close:  closePrice,
			Volume: int64(1_000_000 + i*1_000),
		}
	}
	return bars
}
