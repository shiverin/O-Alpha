package backtest

import (
	"testing"
	"time"

	"github.com/oalpha/pkg/models"
)

func TestRollingSMA(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	sma := rollingSMA(values, 3)
	if sma[2] != 2 {
		t.Fatalf("sma[2] = %v, want 2", sma[2])
	}
	if sma[4] != 4 {
		t.Fatalf("sma[4] = %v, want 4", sma[4])
	}
}

func TestMACrossoverSignals(t *testing.T) {
	// Uptrend then downtrend to force a cross.
	closes := []float64{
		10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		19, 18, 17, 16, 15, 14, 13, 12, 11, 10,
	}
	s := &MACrossover{FastPeriod: 3, SlowPeriod: 5}
	signals := s.Signals(closes)
	hasBuy := false
	hasSell := false
	for _, sig := range signals {
		if sig == SignalBuy {
			hasBuy = true
		}
		if sig == SignalSell {
			hasSell = true
		}
	}
	if !hasBuy {
		t.Fatal("expected at least one buy signal")
	}
	if !hasSell {
		t.Fatal("expected at least one sell signal")
	}
}

func TestRunMACrossover(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := make([]models.Bar, 40)
	for i := range bars {
		price := 100.0 + float64(i%10)
		bars[i] = models.Bar{
			Time:   start.Add(time.Duration(i) * time.Hour),
			Symbol: "TEST",
			Open:   price,
			High:   price + 1,
			Low:    price - 1,
			Close:  price,
			Volume: 1000,
		}
	}

	result, err := RunMACrossover(bars, 3, 5, 10_000)
	if err != nil {
		t.Fatalf("RunMACrossover: %v", err)
	}
	if len(result.EquityCurve) != len(bars) {
		t.Fatalf("equity curve len %d, want %d", len(result.EquityCurve), len(bars))
	}
	if result.FinalEquity <= 0 {
		t.Fatalf("final equity = %v", result.FinalEquity)
	}
}

func TestComputeMetrics(t *testing.T) {
	equity := []float64{100, 105, 103, 110, 108}
	m := ComputeMetrics(equity)
	if m.TotalReturn <= 0 {
		t.Fatalf("total return = %v, want positive", m.TotalReturn)
	}
	if m.MaxDrawdown < 0 {
		t.Fatalf("max drawdown = %v", m.MaxDrawdown)
	}
}
