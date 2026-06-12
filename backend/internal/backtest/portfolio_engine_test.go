package backtest

import (
	"context"
	"testing"
	"time"

	"github.com/oalpha/pkg/models"
)

func TestRunPortfolioBacktestLongTargetPreservesCash(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	panel := oneSymbolPanel("TEST", []models.Bar{
		{Time: t0, Symbol: "TEST", Open: 100, High: 100, Low: 100, Close: 100},
		{Time: t0.Add(time.Hour), Symbol: "TEST", Open: 100, High: 110, Low: 100, Close: 110},
	})
	strategy := &scriptedPortfolioStrategy{
		symbols: []string{"TEST"},
		outputsByLength: map[int]PortfolioOutput{
			1: targetOutput(t0, "TEST", 0.10),
		},
	}

	result, err := RunPortfolioBacktest(context.Background(), panel, strategy, PortfolioBacktestConfig{InitialCash: 10000})
	if err != nil {
		t.Fatalf("portfolio backtest: %v", err)
	}
	if got := result.EquityCurve[1].Equity; got != 10100 {
		t.Fatalf("expected equity 10100, got %.2f", got)
	}
	if result.PositionCurve[1].GrossExposure < 0.10 || result.PositionCurve[1].GrossExposure > 0.11 {
		t.Fatalf("unexpected gross exposure %.4f", result.PositionCurve[1].GrossExposure)
	}
}

func TestRunPortfolioBacktestShortUsesExplicitLiability(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	panel := oneSymbolPanel("TEST", []models.Bar{
		{Time: t0, Symbol: "TEST", Open: 100, High: 100, Low: 100, Close: 100},
		{Time: t0.Add(time.Hour), Symbol: "TEST", Open: 100, High: 100, Low: 90, Close: 90},
	})
	strategy := &scriptedPortfolioStrategy{
		symbols: []string{"TEST"},
		outputsByLength: map[int]PortfolioOutput{
			1: targetOutput(t0, "TEST", -0.10),
		},
	}

	result, err := RunPortfolioBacktest(context.Background(), panel, strategy, PortfolioBacktestConfig{InitialCash: 10000, AllowShorts: true})
	if err != nil {
		t.Fatalf("portfolio backtest: %v", err)
	}
	snapshot := result.PositionCurve[1]
	position := snapshot.Positions["TEST"]
	if position.ShortQty <= 0 || position.LongQty != 0 {
		t.Fatalf("expected explicit short qty only, got %+v", position)
	}
	if snapshot.Equity != 10100 {
		t.Fatalf("expected short equity 10100 after price falls to 90, got %.2f", snapshot.Equity)
	}
}

func TestRunPortfolioBacktestExecutesTargetOnlyOnce(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	panel := oneSymbolPanel("TEST", []models.Bar{
		{Time: t0, Symbol: "TEST", Open: 100, High: 100, Low: 100, Close: 100},
		{Time: t0.Add(24 * time.Hour), Symbol: "TEST", Open: 100, High: 100, Low: 100, Close: 100},
		{Time: t0.Add(48 * time.Hour), Symbol: "TEST", Open: 120, High: 120, Low: 120, Close: 120},
		{Time: t0.Add(72 * time.Hour), Symbol: "TEST", Open: 80, High: 80, Low: 80, Close: 80},
	})
	strategy := &scriptedPortfolioStrategy{
		symbols: []string{"TEST"},
		outputsByLength: map[int]PortfolioOutput{
			1: targetOutput(t0, "TEST", 0.10),
		},
	}

	result, err := RunPortfolioBacktest(context.Background(), panel, strategy, PortfolioBacktestConfig{
		InitialCash: 10000,
		CostModel:   CostModel{},
	})
	if err != nil {
		t.Fatalf("portfolio backtest: %v", err)
	}
	if got := len(result.Trades); got != 1 {
		t.Fatalf("expected one next-bar fill, got %d trades: %+v", got, result.Trades)
	}
	if got := result.Trades[0].Notional; got != 1000 {
		t.Fatalf("expected initial rebalance notional 1000, got %.2f", got)
	}
}

func TestRunPortfolioBacktestEmitsProgress(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	panel := oneSymbolPanel("TEST", []models.Bar{
		{Time: t0, Symbol: "TEST", Open: 100, High: 100, Low: 100, Close: 100},
		{Time: t0.Add(24 * time.Hour), Symbol: "TEST", Open: 100, High: 110, Low: 100, Close: 110},
		{Time: t0.Add(48 * time.Hour), Symbol: "TEST", Open: 110, High: 120, Low: 110, Close: 120},
	})
	strategy := &scriptedPortfolioStrategy{
		symbols: []string{"TEST"},
		outputsByLength: map[int]PortfolioOutput{
			1: targetOutput(t0, "TEST", 0.10),
		},
	}

	var progress []PortfolioBacktestProgress
	result, err := RunPortfolioBacktest(context.Background(), panel, strategy, PortfolioBacktestConfig{
		InitialCash: 10000,
		ProgressCallback: func(update PortfolioBacktestProgress) error {
			progress = append(progress, update)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("portfolio backtest: %v", err)
	}
	if len(progress) != len(result.EquityCurve) {
		t.Fatalf("progress updates=%d, equity points=%d", len(progress), len(result.EquityCurve))
	}
	for i, update := range progress {
		if update.Index != i {
			t.Fatalf("progress[%d].Index=%d", i, update.Index)
		}
		if update.Total != len(panel.Times) {
			t.Fatalf("progress[%d].Total=%d", i, update.Total)
		}
		if update.Point != result.EquityCurve[i] {
			t.Fatalf("progress[%d].Point=%+v, want %+v", i, update.Point, result.EquityCurve[i])
		}
	}
	if got := progress[len(progress)-1].Percent; got != 1 {
		t.Fatalf("final progress percent=%v, want 1", got)
	}
}

type scriptedPortfolioStrategy struct {
	symbols         []string
	outputsByLength map[int]PortfolioOutput
}

func (s *scriptedPortfolioStrategy) GeneratePortfolioSignals(ctx context.Context, panel AlignedBars) ([]PortfolioOutput, error) {
	outputs := make([]PortfolioOutput, len(panel.Times))
	for i := range outputs {
		outputs[i] = s.outputsByLength[i+1]
	}
	return outputs, nil
}

func (s *scriptedPortfolioStrategy) EvaluatePortfolioLatest(ctx context.Context, panel AlignedBars) (PortfolioOutput, error) {
	return s.outputsByLength[len(panel.Times)], nil
}

func (s *scriptedPortfolioStrategy) Universe() []string {
	return s.symbols
}

func (s *scriptedPortfolioStrategy) Name() string {
	return "scripted"
}

func oneSymbolPanel(symbol string, bars []models.Bar) AlignedBars {
	times := make([]time.Time, len(bars))
	for i, bar := range bars {
		times[i] = bar.Time
	}
	return AlignedBars{
		Times:   times,
		Symbols: []string{symbol},
		Bars: map[string][]models.Bar{
			symbol: bars,
		},
		Timeframe: "1Hour",
	}
}

func targetOutput(t time.Time, symbol string, weight float64) PortfolioOutput {
	return PortfolioOutput{
		Time: t,
		Targets: map[string]TargetPosition{
			symbol: {
				Symbol:       symbol,
				TargetWeight: weight,
				Engine:       "test",
			},
		},
	}
}
