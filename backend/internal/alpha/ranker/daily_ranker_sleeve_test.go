package ranker

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

func TestDailyRankerSleeveLoadsLambdaRankModelAndAllocatesSleeve(t *testing.T) {
	panel := rankerTestPanel(270)
	modelPath := writeRankerOneLeafModel(t, 0.25)
	modelHash, err := fileSHA256(modelPath)
	if err != nil {
		t.Fatalf("fileSHA256: %v", err)
	}
	strategy := NewDailyRankerSleeveStrategy(panel.Symbols, DailyRankerSleeveConfig{
		BenchmarkSymbol:    "VOO",
		CandidateUniverse:  "stocks",
		ModelArtifactRoot:  "/tmp/ranker_artifacts",
		ModelVariant:       "unit_variant",
		ModelPathsByYear:   map[int]string{2020: modelPath},
		RebalanceEveryBars: 63,
		SleeveFraction:     0.10,
		TopK:               1,
		MaxNameWeight:      0.10,
	})

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), rankerPanelPrefix(panel, 260))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if got := output.Targets["AAA"].TargetWeight; math.Abs(got-0.10) > 1e-9 {
		t.Fatalf("AAA target=%f, want 0.10; targets=%v", got, output.Targets)
	}
	if got := output.Targets["VOO"].TargetWeight; math.Abs(got-0.90) > 1e-9 {
		t.Fatalf("VOO target=%f, want residual benchmark 0.90; targets=%v", got, output.Targets)
	}
	if _, ok := output.Targets["QQQ"]; ok {
		t.Fatalf("QQQ should be excluded from stock universe")
	}
	if output.EngineMetadata["engine"] != DailyRankerSleeveStrategyName {
		t.Fatalf("metadata engine=%v", output.EngineMetadata["engine"])
	}
	if output.EngineMetadata["ranker_model_loaded"] != true {
		t.Fatalf("ranker_model_loaded=%v, want true", output.EngineMetadata["ranker_model_loaded"])
	}
	if output.EngineMetadata["ranker_model_year"] != 2020 {
		t.Fatalf("ranker_model_year=%v, want 2020", output.EngineMetadata["ranker_model_year"])
	}
	if output.EngineMetadata["ranker_model_path"] != modelPath {
		t.Fatalf("ranker_model_path=%v, want %s", output.EngineMetadata["ranker_model_path"], modelPath)
	}
	if output.EngineMetadata["ranker_model_sha256"] != modelHash {
		t.Fatalf("ranker_model_sha256=%v, want %s", output.EngineMetadata["ranker_model_sha256"], modelHash)
	}
	if output.EngineMetadata["ranker_model_variant"] != "unit_variant" {
		t.Fatalf("ranker_model_variant=%v, want unit_variant", output.EngineMetadata["ranker_model_variant"])
	}
	if output.EngineMetadata["ranker_model_artifact_root"] != "/tmp/ranker_artifacts" {
		t.Fatalf("ranker_model_artifact_root=%v, want /tmp/ranker_artifacts", output.EngineMetadata["ranker_model_artifact_root"])
	}
	held, err := strategy.EvaluatePortfolioLatest(context.Background(), rankerPanelPrefix(panel, 261))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest hold: %v", err)
	}
	if held.EngineMetadata["action"] != "hold_targets" {
		t.Fatalf("hold action=%v, want hold_targets", held.EngineMetadata["action"])
	}
	if held.EngineMetadata["ranker_model_sha256"] != modelHash {
		t.Fatalf("hold ranker_model_sha256=%v, want %s", held.EngineMetadata["ranker_model_sha256"], modelHash)
	}
}

func TestDailyRankerSleeveHoldsBenchmarkWhenYearModelMissing(t *testing.T) {
	panel := rankerTestPanel(270)
	strategy := NewDailyRankerSleeveStrategy(panel.Symbols, DailyRankerSleeveConfig{
		BenchmarkSymbol:    "VOO",
		CandidateUniverse:  "stocks",
		RebalanceEveryBars: 63,
		SleeveFraction:     0.10,
		TopK:               1,
		MaxNameWeight:      0.10,
	})

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), rankerPanelPrefix(panel, 260))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if got := output.Targets["VOO"].TargetWeight; math.Abs(got-1) > 1e-9 {
		t.Fatalf("VOO target=%f, want benchmark-only fallback", got)
	}
	if output.EngineMetadata["reason"] != "missing_year_model" {
		t.Fatalf("reason=%v, want missing_year_model", output.EngineMetadata["reason"])
	}
	if output.EngineMetadata["ranker_model_loaded"] != false {
		t.Fatalf("ranker_model_loaded=%v, want false", output.EngineMetadata["ranker_model_loaded"])
	}
	if output.EngineMetadata["ranker_model_year"] != 2020 {
		t.Fatalf("ranker_model_year=%v, want 2020", output.EngineMetadata["ranker_model_year"])
	}
}

func TestDailyRankerSleeveExcludesConfiguredSymbols(t *testing.T) {
	panel := rankerTestPanel(270)
	modelPath := writeRankerOneLeafModel(t, 0.25)
	strategy := NewDailyRankerSleeveStrategy(panel.Symbols, DailyRankerSleeveConfig{
		BenchmarkSymbol:    "VOO",
		CandidateUniverse:  "stocks",
		ExcludedSymbols:    []string{"aaa"},
		ModelPathsByYear:   map[int]string{2020: modelPath},
		RebalanceEveryBars: 63,
		SleeveFraction:     0.10,
		TopK:               1,
		MaxNameWeight:      0.10,
	})

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), rankerPanelPrefix(panel, 260))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if _, ok := output.Targets["AAA"]; ok {
		t.Fatalf("AAA should be excluded from the active sleeve: targets=%v", output.Targets)
	}
	if got := output.Targets["BBB"].TargetWeight; math.Abs(got-0.10) > 1e-9 {
		t.Fatalf("BBB target=%f, want 0.10 after AAA exclusion; targets=%v", got, output.Targets)
	}
}

func TestDailyRankerSleeveFiltersPointInTimeUniverse(t *testing.T) {
	panel := rankerTestPanel(270)
	modelPath := writeRankerOneLeafModel(t, 0.25)
	pitUniverse, err := NewPointInTimeUniverse([]ConstituentInterval{
		{Symbol: "AAA", Start: "2019-01-01", End: "2020-03-31"},
		{Symbol: "BBB", Start: "2019-01-01"},
	})
	if err != nil {
		t.Fatalf("NewPointInTimeUniverse: %v", err)
	}
	strategy := NewDailyRankerSleeveStrategy(panel.Symbols, DailyRankerSleeveConfig{
		BenchmarkSymbol:     "VOO",
		CandidateUniverse:   "stocks",
		PointInTimeUniverse: pitUniverse,
		ModelPathsByYear:    map[int]string{2020: modelPath},
		RebalanceEveryBars:  63,
		SleeveFraction:      0.10,
		TopK:                1,
		MaxNameWeight:       0.10,
	})

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), rankerPanelPrefix(panel, 260))
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if _, ok := output.Targets["AAA"]; ok {
		t.Fatalf("AAA should be inactive in the point-in-time universe: targets=%v", output.Targets)
	}
	if got := output.Targets["BBB"].TargetWeight; math.Abs(got-0.10) > 1e-9 {
		t.Fatalf("BBB target=%f, want 0.10 after PIT filter; targets=%v", got, output.Targets)
	}
	if output.EngineMetadata["point_in_time_universe"] != true {
		t.Fatalf("metadata point_in_time_universe=%v, want true", output.EngineMetadata["point_in_time_universe"])
	}
}

func TestLoadPointInTimeUniverse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pit.json")
	payload := `{"version":"pit_constituents_v1","intervals":[{"symbol":"AAA","start":"2020-01-01","end":"2020-12-31"},{"symbol":"BBB","start":"2021-01-01"}]}`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("write PIT manifest: %v", err)
	}
	universe, err := LoadPointInTimeUniverse(path)
	if err != nil {
		t.Fatalf("LoadPointInTimeUniverse: %v", err)
	}
	if !universe.Active("AAA", time.Date(2020, 6, 1, 14, 30, 0, 0, time.UTC)) {
		t.Fatalf("AAA should be active in 2020")
	}
	if universe.Active("AAA", time.Date(2021, 6, 1, 14, 30, 0, 0, time.UTC)) {
		t.Fatalf("AAA should be inactive after 2020")
	}
	if !universe.Active("BBB", time.Date(2026, 6, 1, 14, 30, 0, 0, time.UTC)) {
		t.Fatalf("BBB should remain active with open-ended interval")
	}
}

func rankerTestPanel(n int) backtest.AlignedBars {
	symbols := []string{"VOO", "AAA", "BBB", "QQQ"}
	start := time.Date(2020, 1, 2, 14, 30, 0, 0, time.UTC)
	panel := backtest.AlignedBars{
		Symbols:   symbols,
		Times:     make([]time.Time, n),
		Bars:      make(map[string][]models.Bar, len(symbols)),
		Timeframe: "1Day",
	}
	for i := 0; i < n; i++ {
		panel.Times[i] = start.AddDate(0, 0, i)
	}
	priceFns := map[string]func(int) float64{
		"VOO": func(i int) float64 { return 100 * math.Exp(0.0005*float64(i)) },
		"AAA": func(i int) float64 { return 50 * math.Exp(0.0010*float64(i)) },
		"BBB": func(i int) float64 { return 70 * math.Exp(0.0008*float64(i)) },
		"QQQ": func(i int) float64 { return 80 * math.Exp(0.0015*float64(i)) },
	}
	for _, symbol := range symbols {
		bars := make([]models.Bar, n)
		for i := 0; i < n; i++ {
			closePrice := priceFns[symbol](i)
			openPrice := closePrice * 0.999
			bars[i] = models.Bar{
				Time:   panel.Times[i],
				Symbol: symbol,
				Open:   openPrice,
				High:   closePrice * 1.01,
				Low:    closePrice * 0.99,
				Close:  closePrice,
				Volume: 1_000_000,
			}
		}
		panel.Bars[symbol] = bars
	}
	return panel
}

func rankerPanelPrefix(panel backtest.AlignedBars, n int) backtest.AlignedBars {
	if n > len(panel.Times) {
		n = len(panel.Times)
	}
	out := panel
	out.Times = append([]time.Time(nil), panel.Times[:n]...)
	out.Bars = make(map[string][]models.Bar, len(panel.Bars))
	for symbol, bars := range panel.Bars {
		out.Bars[symbol] = append([]models.Bar(nil), bars[:n]...)
	}
	return out
}

func writeRankerOneLeafModel(t *testing.T, leafValue float64) string {
	t.Helper()
	payload := strings.Join([]string{
		"tree",
		"version=v2",
		"num_class=1",
		"num_tree_per_iteration=1",
		"label_index=0",
		"max_feature_idx=0",
		"objective=lambdarank",
		"feature_names=f0",
		"feature_infos=[0:1]",
		"tree_sizes=128",
		"",
		"Tree=0",
		"num_leaves=1",
		"num_cat=0",
		"split_feature=",
		"split_gain=",
		"threshold=",
		"decision_type=",
		"left_child=",
		"right_child=",
		"leaf_value=" + strconv.FormatFloat(leafValue, 'g', -1, 64),
		"leaf_count=",
		"internal_value=",
		"internal_count=",
		"shrinkage=1",
		"",
		"end of trees",
		"",
	}, "\n")
	path := filepath.Join(t.TempDir(), "model.txt")
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("write model: %v", err)
	}
	return path
}
