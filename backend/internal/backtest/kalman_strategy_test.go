package backtest

import (
	"context"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/oalpha/pkg/models"
)

func TestKalmanFilter1D_Smoothing(t *testing.T) {
	// Setup filter with small process noise and high measurement noise (heavy smoothing)
	kf := NewKalmanFilter1D(0.01, 1.0)

	// Seed random for synthetic noise generation
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	trueValue := 100.0
	var unfilteredSumVariance float64
	var filteredSumVariance float64

	// Feed a flat line corrupted by massive white noise
	for i := 0; i < 100; i++ {
		noise := (rng.Float64() - 0.5) * 10.0 // Noise up to +/- 5
		measurement := trueValue + noise
		estimate := kf.Update(measurement)

		unfilteredSumVariance += math.Abs(measurement - trueValue)
		filteredSumVariance += math.Abs(estimate - trueValue)
	}

	if filteredSumVariance >= unfilteredSumVariance {
		t.Errorf("Kalman Filter failed to smooth variance. Raw variance: %f, Filtered: %f", unfilteredSumVariance, filteredSumVariance)
	}
}

func TestKalmanStrategy_SignalTrigger(t *testing.T) {
	strategy := NewKalmanStrategy(0.01, 0.5, 5, 2.0)
	ctx := context.Background()

	// Create baseline history (stable pricing)
	bars := []models.Bar{
		{Close: 100.0}, {Close: 100.1}, {Close: 100.0}, {Close: 99.9}, {Close: 100.0},
	}

	// Baseline processing
	_, _ = strategy.GenerateSignal(ctx, bars)

	// Introduce a massive, sudden structural retail flush (anomaly)
	crashBars := append(bars, models.Bar{Close: 90.0})
	signals, err := strategy.GenerateSignal(ctx, crashBars)
	if err != nil {
		t.Fatalf("GenerateSignal failed: %v", err)
	}

	if len(signals) == 0 {
		t.Fatal("Expected at least one signal")
	}

	lastSignal := signals[len(signals)-1]
	if lastSignal != models.SignalBuy {
		t.Errorf("Expected SignalBuy on sharp statistical deviation down, got: %v", lastSignal)
	}
}

func TestKalmanStrategy_GenerateSignalPerBar(t *testing.T) {
	strategy := NewKalmanStrategy(0.01, 0.5, 5, 2.0)
	ctx := context.Background()

	bars := []models.Bar{
		{Close: 100.0}, {Close: 100.1}, {Close: 100.0}, {Close: 99.9},
		{Close: 100.0}, {Close: 90.0}, // Sharp drop → should trigger buy
	}

	signals, err := strategy.GenerateSignal(ctx, bars)
	if err != nil {
		t.Fatalf("GenerateSignal failed: %v", err)
	}

	if len(signals) != len(bars) {
		t.Errorf("Expected %d signals, got %d", len(bars), len(signals))
	}

	// Signals should be indexes 0-4: Hold (not enough lookback), last: Buy
	for i, sig := range signals[:len(signals)-1] {
		if sig != models.SignalHold {
			t.Errorf("Expected Hold at index %d, got %v", i, sig)
		}
	}

	if signals[len(signals)-1] != models.SignalBuy {
		t.Errorf("Expected Buy on last bar, got %v", signals[len(signals)-1])
	}
}
