package cointegration

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

func TestKalmanPairStrategyDoesNotImplementSingleSymbolStrategy(t *testing.T) {
	strategy := NewKalmanPairStrategy(DefaultKalmanPairConfig("Y", "X"), nil, nil)
	if _, ok := interface{}(strategy).(backtest.Strategy); ok {
		t.Fatalf("pair strategy must not implement the single-symbol Strategy interface")
	}
}

func TestKalmanPairStrategyPositiveZShortsYLongsX(t *testing.T) {
	panel := syntheticPairPanel(80, 1.2, 0.3, 0)
	shockLatestY(&panel, 0.12)
	cfg := testPairConfig()
	strategy := NewKalmanPairStrategy(cfg, nil, nil)

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panel)
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if output.Targets["Y"].TargetWeight >= 0 {
		t.Fatalf("Y target=%f, want short Y", output.Targets["Y"].TargetWeight)
	}
	if output.Targets["X"].TargetWeight <= 0 {
		t.Fatalf("X target=%f, want long X", output.Targets["X"].TargetWeight)
	}
}

func TestKalmanPairStrategyNegativeZLongsYShortsX(t *testing.T) {
	panel := syntheticPairPanel(80, 1.2, 0.3, 0)
	shockLatestY(&panel, -0.12)
	cfg := testPairConfig()
	strategy := NewKalmanPairStrategy(cfg, nil, nil)

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panel)
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if output.Targets["Y"].TargetWeight <= 0 {
		t.Fatalf("Y target=%f, want long Y", output.Targets["Y"].TargetWeight)
	}
	if output.Targets["X"].TargetWeight >= 0 {
		t.Fatalf("X target=%f, want short X", output.Targets["X"].TargetWeight)
	}
}

func TestKalmanPairStrategyStopZQuarantinesPair(t *testing.T) {
	panel := syntheticPairPanel(80, 1.2, 0.3, 0)
	shockLatestY(&panel, 0.25)
	cfg := testPairConfig()
	cfg.StopZ = 2.0
	strategy := NewKalmanPairStrategy(cfg, nil, nil)

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panel)
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if output.EngineMetadata["reason"] != "stop_z_quarantine" {
		t.Fatalf("reason=%v, want stop_z_quarantine", output.EngineMetadata["reason"])
	}
	if !strategy.State().Quarantined {
		t.Fatalf("pair should be quarantined after stop")
	}
}

func TestKalmanPairStrategyRejectsFailedShortGate(t *testing.T) {
	panel := syntheticPairPanel(80, 1.2, 0.3, 0)
	shockLatestY(&panel, 0.12)
	cfg := testPairConfig()
	cfg.RequireShortable = true
	strategy := NewKalmanPairStrategy(cfg, staticShortGate{shortable: map[string]bool{"Y": false, "X": true}}, nil)

	_, err := strategy.EvaluatePortfolioLatest(context.Background(), panel)
	if err == nil {
		t.Fatalf("expected failed short gate")
	}
}

func TestKalmanPairStrategyFailedRetestBlocksEntry(t *testing.T) {
	panel := syntheticPairPanel(80, 1.2, 0.3, 0)
	shockLatestY(&panel, 0.12)
	cfg := testPairConfig()
	strategy := NewKalmanPairStrategy(cfg, nil, staticRetester{approved: false, reason: "rolling_cointegration_failed"})

	output, err := strategy.EvaluatePortfolioLatest(context.Background(), panel)
	if err != nil {
		t.Fatalf("EvaluatePortfolioLatest: %v", err)
	}
	if len(output.Targets) != 0 || output.EngineMetadata["reason"] != "rolling_cointegration_failed" {
		t.Fatalf("output=%+v, want retest hold", output)
	}
}

func testPairConfig() KalmanPairConfig {
	cfg := DefaultKalmanPairConfig("Y", "X")
	cfg.QAlpha = 1e-6
	cfg.QBeta = 1e-6
	cfg.R = 1e-4
	cfg.EntryZ = 1.0
	cfg.ExitZ = 0.25
	cfg.StopZ = 20.0
	cfg.MaxGrossWeight = 0.20
	cfg.MaxLegWeight = 0.12
	return cfg
}

func syntheticPairPanel(n int, beta, alpha, noise float64) backtest.AlignedBars {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	times := make([]time.Time, n)
	yBars := make([]models.Bar, n)
	xBars := make([]models.Bar, n)
	for i := 0; i < n; i++ {
		times[i] = start.AddDate(0, 0, i)
		logX := math.Log(100) + float64(i)*0.001
		logY := alpha + beta*logX + noise*math.Sin(float64(i))
		xClose := math.Exp(logX)
		yClose := math.Exp(logY)
		xBars[i] = pairBar("X", times[i], xClose)
		yBars[i] = pairBar("Y", times[i], yClose)
	}
	return backtest.AlignedBars{
		Times:   times,
		Symbols: []string{"Y", "X"},
		Bars: map[string][]models.Bar{
			"Y": yBars,
			"X": xBars,
		},
	}
}

func shockLatestY(panel *backtest.AlignedBars, logShock float64) {
	i := len(panel.Times) - 1
	bar := panel.Bars["Y"][i]
	bar.Close *= math.Exp(logShock)
	bar.Open = bar.Close
	bar.High = bar.Close
	bar.Low = bar.Close
	panel.Bars["Y"][i] = bar
}

func pairBar(symbol string, t time.Time, close float64) models.Bar {
	return models.Bar{
		Time:   t,
		Symbol: symbol,
		Open:   close,
		High:   close,
		Low:    close,
		Close:  close,
		Volume: 1_000_000,
	}
}

type staticShortGate struct {
	shortable map[string]bool
}

func (g staticShortGate) IsShortable(ctx context.Context, symbol string, at time.Time) (bool, error) {
	_ = ctx
	_ = at
	return g.shortable[symbol], nil
}

type staticRetester struct {
	approved bool
	reason   string
}

func (r staticRetester) PairStillApproved(ctx context.Context, symbolY, symbolX string, at time.Time) (bool, string, error) {
	_ = ctx
	_ = symbolY
	_ = symbolX
	_ = at
	return r.approved, r.reason, nil
}
