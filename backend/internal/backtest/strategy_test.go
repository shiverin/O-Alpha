package backtest

import (
	"context"
	"testing"
	"time"

	"github.com/oalpha/pkg/models"
)

func TestMACrossoverStrategy_GenerateSignal(t *testing.T) {
	// Test case 1: Normal operation with uptrend then downtrend
	closes := []float64{
		10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		19, 18, 17, 16, 15, 14, 13, 12, 11, 10,
	}
	strat := NewMACrossoverStrategy(3, 5)
	// Test with nil bars should return error (not enough bars)
	_, err := strat.GenerateSignal(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil bars")
	} else if err.Error() != "not enough bars: need at least 6" {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test with actual bars
	bars := makeBarsFromCloses(closes, "TEST")
	signals, err := strat.GenerateSignal(context.Background(), bars)
	if err != nil {
		t.Fatalf("GenerateSignal returned error: %v", err)
	}
	if len(signals) != len(bars) {
		t.Fatalf("Expected %d signals, got %d", len(bars), len(signals))
	}

	// Check for at least one buy and one sell signal
	hasBuy := false
	hasSell := false
	for _, sig := range signals {
		if sig == models.SignalBuy {
			hasBuy = true
		}
		if sig == models.SignalSell {
			hasSell = true
		}
	}
	if !hasBuy {
		t.Error("Expected at least one buy signal")
	}
	if !hasSell {
		t.Error("Expected at least one sell signal")
	}

	// Test case 2: Invalid periods
	stratFastNeg := NewMACrossoverStrategy(-1, 5)
	_, err = stratFastNeg.GenerateSignal(context.Background(), bars)
	if err == nil {
		t.Error("Expected error for negative fast period")
	}
	stratSlowNeg := NewMACrossoverStrategy(3, -1)
	_, err = stratSlowNeg.GenerateSignal(context.Background(), bars)
	if err == nil {
		t.Error("Expected error for negative slow period")
	}
	stratFastEqSlow := NewMACrossoverStrategy(5, 5)
	_, err = stratFastEqSlow.GenerateSignal(context.Background(), bars)
	if err == nil {
		t.Error("Expected error for fast period equal to slow period")
	}
	stratFastGtSlow := NewMACrossoverStrategy(6, 5)
	_, err = stratFastGtSlow.GenerateSignal(context.Background(), bars)
	if err == nil {
		t.Error("Expected error for fast period greater than slow period")
	}

	// Test case 3: Not enough bars
	shortCloses := []float64{1, 2, 3}
	shortBars := makeBarsFromCloses(shortCloses, "TEST")
	_, err = NewMACrossoverStrategy(5, 10).GenerateSignal(context.Background(), shortBars)
	if err == nil {
		t.Error("Expected error for insufficient bars")
	}
}

func TestStrategy_Interface(t *testing.T) {
	// Ensure MACrossoverStrategy implements the Strategy interface
	var _ Strategy = (*MACrossoverStrategy)(nil)
}

func TestRunBacktest_WithMACrossoverStrategy(t *testing.T) {
	// Test that RunBacktest works with the Strategy interface
	closes := []float64{
		10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		19, 18, 17, 16, 15, 14, 13, 12, 11, 10,
	}
	bars := makeBarsFromCloses(closes, "TEST")
	strat := NewMACrossoverStrategy(3, 5)

	result, err := RunBacktest(context.Background(), bars, strat, 10_000)
	if err != nil {
		t.Fatalf("RunBacktest returned error: %v", err)
	}
	if len(result.EquityCurve) != len(bars) {
		t.Fatalf("Expected equity curve length %d, got %d", len(bars), len(result.EquityCurve))
	}
	if result.FinalEquity <= 0 {
		t.Fatalf("Expected final equity > 0, got %f", result.FinalEquity)
	}
	// Additional checks: ensure we have some trades
	if result.NumTrades == 0 {
		t.Error("Expected at least one trade")
	}
}

func TestRunBacktest_ErrorConditions(t *testing.T) {
	// Test RunBacktest with nil bars
	_, err := RunBacktest(context.Background(), nil, NewMACrossoverStrategy(3, 5), 10_000)
	if err == nil {
		t.Error("Expected error for nil bars")
	} else if err.Error() != "need at least one bar" {
		t.Fatalf("Unexpected error for nil bars: %v", err)
	}

	// Test RunBacktest with empty bars
	_, err = RunBacktest(context.Background(), []models.Bar{}, NewMACrossoverStrategy(3, 5), 10_000)
	if err == nil {
		t.Error("Expected error for empty bars")
	} else if err.Error() != "need at least one bar" {
		t.Fatalf("Unexpected error for empty bars: %v", err)
	}

	// Test RunBacktest with zero initial cash: should use default initial cash (100_000) and not error
	bars := makeBarsFromCloses([]float64{1, 2, 3, 4, 5}, "TEST")
	_, err = RunBacktest(context.Background(), bars, NewMACrossoverStrategy(2, 3), 0)
	if err != nil {
		t.Fatalf("Unexpected error for zero initial cash: %v", err)
	}
}

// Helper function to create Bar slices from close prices
func makeBarsFromCloses(closes []float64, symbol string) []models.Bar {
	bars := make([]models.Bar, len(closes))
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, close := range closes {
		price := close
		bars[i] = models.Bar{
			Time:   start.Add(time.Duration(i) * time.Hour),
			Symbol: symbol,
			Open:   price,
			High:   price + 1,
			Low:    price - 1,
			Close:  price,
			Volume: 1000,
		}
	}
	return bars
}
