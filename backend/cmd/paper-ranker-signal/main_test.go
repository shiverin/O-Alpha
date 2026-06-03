package main

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

func TestParseYearsDeduplicatesAndSorts(t *testing.T) {
	years, err := parseYears("2026, 2024,2025,2024")
	if err != nil {
		t.Fatalf("parseYears returned error: %v", err)
	}
	if got, want := joinYears(years), "2024,2025,2026"; got != want {
		t.Fatalf("years=%s, want %s", got, want)
	}
}

func TestBuildReportUsesLastNonEmptyTargets(t *testing.T) {
	t0 := time.Date(2024, 1, 2, 14, 30, 0, 0, time.UTC)
	t1 := t0.Add(24 * time.Hour)
	panel := backtest.AlignedBars{
		Times:     []time.Time{t0, t1},
		Symbols:   []string{"VOO", "AAPL"},
		Timeframe: "1Day",
		Bars: map[string][]models.Bar{
			"VOO": {
				{Time: t0, Symbol: "VOO", Close: 100},
				{Time: t1, Symbol: "VOO", Close: 101},
			},
			"AAPL": {
				{Time: t0, Symbol: "AAPL", Close: 10},
				{Time: t1, Symbol: "AAPL", Close: 11},
			},
		},
	}
	outputs := []backtest.PortfolioOutput{
		{
			Time: t0,
			Targets: map[string]backtest.TargetPosition{
				"VOO": {
					Symbol:       "VOO",
					TargetWeight: 0.95,
					AlphaScore:   1,
					Confidence:   1,
					Side:         backtest.PositionSideLong,
					Engine:       "daily_lgbm_ranker_sleeve",
					Metadata:     map[string]interface{}{"role": "benchmark_core", "rebalance": true},
				},
				"AAPL": {
					Symbol:       "AAPL",
					TargetWeight: 0.05,
					AlphaScore:   2.4,
					Confidence:   1,
					Side:         backtest.PositionSideLong,
					Engine:       "daily_lgbm_ranker_sleeve",
					Metadata:     map[string]interface{}{"role": "active_sleeve", "rebalance": true},
				},
			},
			GrossExposure: 1,
			NetExposure:   1,
			EngineMetadata: map[string]interface{}{
				"action":                     "rebalance",
				"ranker_model_loaded":        true,
				"ranker_model_sha256":        "abc123",
				"ranker_model_variant":       "stocks_h63_s15_top3_reb63_z10",
				"ranker_model_feature_count": 31,
				"selection_rows": []map[string]interface{}{
					{"symbol": "AAPL", "rank": 1, "score": 2.4, "score_z": 1.2, "vol_20": 0.20},
				},
			},
		},
		{
			Time:           t1,
			Targets:        map[string]backtest.TargetPosition{},
			GrossExposure:  1,
			NetExposure:    1,
			EngineMetadata: map[string]interface{}{"action": "hold_targets", "ranker_model_sha256": "abc123"},
		},
	}
	report, err := buildReport(signalArgs{
		BarsCSV:           "bars.csv",
		Symbols:           []string{"VOO", "AAPL"},
		Timeframe:         "1Day",
		Benchmark:         "VOO",
		ModelArtifactRoot: "fold_artifacts",
		ModelVariant:      "stocks_h63_s15_top3_reb63_z10",
	}, panel, outputs, t1)
	if err != nil {
		t.Fatalf("buildReport returned error: %v", err)
	}
	if !report.PaperOnly || report.OrdersEnabled || report.BrokerConnected {
		t.Fatalf("paper flags mismatch: paper=%t orders=%t broker=%t", report.PaperOnly, report.OrdersEnabled, report.BrokerConnected)
	}
	if report.TargetSource != "last_non_empty_target" {
		t.Fatalf("target source=%s, want last_non_empty_target", report.TargetSource)
	}
	if report.LastRebalanceTime == nil || !report.LastRebalanceTime.Equal(t0) {
		t.Fatalf("last rebalance time=%v, want %v", report.LastRebalanceTime, t0)
	}
	if len(report.Targets) != 2 {
		t.Fatalf("targets=%d, want 2", len(report.Targets))
	}
	if report.Targets[0].Symbol != "VOO" || report.Targets[1].Symbol != "AAPL" {
		t.Fatalf("targets not benchmark-first then active: %+v", report.Targets)
	}
	if report.Targets[1].ModelScore == nil || *report.Targets[1].ModelScore != 2.4 {
		t.Fatalf("active model score not copied from selection metadata: %+v", report.Targets[1])
	}
	markdown := report.Markdown()
	if !strings.Contains(markdown, "Orders enabled: `false`") || !strings.Contains(markdown, "ranker_model_sha256") {
		t.Fatalf("markdown missing paper/model audit fields:\n%s", markdown)
	}
}

func joinYears(years []int) string {
	parts := make([]string, 0, len(years))
	for _, year := range years {
		parts = append(parts, strconv.Itoa(year))
	}
	return strings.Join(parts, ",")
}
