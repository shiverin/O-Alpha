package ml

import (
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

func TestLabelEventsUsesFirstTouchedBarrier(t *testing.T) {
	bars := barsFromCloses("VOO", []float64{100, 102, 98, 101})
	events := []PrimaryEvent{{
		Symbol:     "VOO",
		Time:       bars[0].Time,
		Index:      0,
		Side:       1,
		Price:      100,
		Volatility: 0.01,
	}}
	labels, err := LabelEvents(bars, events, TripleBarrierConfig{
		HorizonBars:       3,
		ProfitTakeVolMult: 1,
		StopLossVolMult:   1,
	})
	if err != nil {
		t.Fatalf("LabelEvents: %v", err)
	}
	if labels[0].Label != 1 || labels[0].Barrier != BarrierProfitTake {
		t.Fatalf("label=%d barrier=%s, want profit take", labels[0].Label, labels[0].Barrier)
	}
	if !labels[0].EndTime.Equal(bars[1].Time) {
		t.Fatalf("end time=%s, want first profit bar %s", labels[0].EndTime, bars[1].Time)
	}
}

func TestLabelEventsStopLossAndVertical(t *testing.T) {
	stopBars := barsFromCloses("VOO", []float64{100, 99, 103})
	stopEvents := []PrimaryEvent{{
		Symbol:     "VOO",
		Time:       stopBars[0].Time,
		Index:      0,
		Side:       1,
		Price:      100,
		Volatility: 0.01,
	}}
	stopLabels, err := LabelEvents(stopBars, stopEvents, TripleBarrierConfig{
		HorizonBars:       2,
		ProfitTakeVolMult: 2,
		StopLossVolMult:   1,
	})
	if err != nil {
		t.Fatalf("LabelEvents stop: %v", err)
	}
	if stopLabels[0].Label != 0 || stopLabels[0].Barrier != BarrierStopLoss {
		t.Fatalf("label=%d barrier=%s, want stop loss", stopLabels[0].Label, stopLabels[0].Barrier)
	}

	verticalBars := barsFromCloses("VOO", []float64{100, 100.2, 100.3, 100.4})
	verticalEvents := []PrimaryEvent{{
		Symbol:     "VOO",
		Time:       verticalBars[0].Time,
		Index:      0,
		Side:       1,
		Price:      100,
		Volatility: 0.10,
	}}
	verticalLabels, err := LabelEvents(verticalBars, verticalEvents, TripleBarrierConfig{
		HorizonBars:       2,
		ProfitTakeVolMult: 1,
		StopLossVolMult:   1,
	})
	if err != nil {
		t.Fatalf("LabelEvents vertical: %v", err)
	}
	if verticalLabels[0].Label != 0 || verticalLabels[0].Barrier != BarrierVertical {
		t.Fatalf("label=%d barrier=%s, want vertical", verticalLabels[0].Label, verticalLabels[0].Barrier)
	}
	if !verticalLabels[0].EndTime.Equal(verticalBars[2].Time) {
		t.Fatalf("end time=%s, want vertical horizon %s", verticalLabels[0].EndTime, verticalBars[2].Time)
	}
}

func TestGeneratePrimaryEventsUsesLongEntriesOnly(t *testing.T) {
	bars := barsFromCloses("VOO", []float64{100, 101, 102, 103, 104})
	outputs := []backtest.StrategyOutput{
		{Signal: models.SignalHold},
		{Signal: models.SignalBuy, AlphaScore: 0.8},
		{Signal: models.SignalSell},
		{Signal: models.SignalBuy, AlphaScore: 0.6},
		{Signal: models.SignalHold},
	}
	events, err := GeneratePrimaryEvents(bars, outputs, TripleBarrierConfig{MinEventSpacingBars: 2})
	if err != nil {
		t.Fatalf("GeneratePrimaryEvents: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("events=%d, want 2", len(events))
	}
	if events[0].Index != 1 || events[1].Index != 3 {
		t.Fatalf("event indices=%d,%d want 1,3", events[0].Index, events[1].Index)
	}
}

func barsFromCloses(symbol string, closes []float64) []models.Bar {
	start := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := make([]models.Bar, len(closes))
	for i, close := range closes {
		bars[i] = models.Bar{
			Time:   start.AddDate(0, 0, i),
			Symbol: symbol,
			Open:   close,
			High:   close,
			Low:    close,
			Close:  close,
			Volume: 1_000_000,
		}
	}
	return bars
}
