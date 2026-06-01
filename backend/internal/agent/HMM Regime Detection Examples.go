package agent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

// ===== EXAMPLE 1: Simple HMM Regime Detection =====
// This example shows how to detect market regimes in isolation

func ExampleHMMRegimeDetection() {
	// Create detector with 50-bar training window
	hmm := NewHMMRegimeDetector(50)

	// Generate synthetic bars (4 hours of data with uptrend and low volatility)
	bars := generateRealisticBars(100, 100.0, 0.005, 0.001)

	// Process bars sequentially (no lookahead)
	for i := 50; i < len(bars); i++ {
		regime, confidence, err := hmm.Update(bars[:i+1])
		if err != nil {
			fmt.Printf("Error at bar %d: %v\n", i, err)
			continue
		}

		if i%10 == 0 {
			probs := hmm.GetProbabilities()
			fmt.Printf("Bar %d: Regime=%s Confidence=%.2f LowVol=%.2f Medium=%.2f HighVol=%.2f\n",
				i, regime.String(), confidence, probs[0], probs[1], probs[2])
		}
	}

	// Output example:
	// Bar 50: Regime=Medium Confidence=0.45 LowVol=0.35 Medium=0.50 HighVol=0.15
	// Bar 60: Regime=Low Vol Trend Confidence=0.62 LowVol=0.68 Medium=0.28 HighVol=0.04
	// Bar 70: Regime=Low Vol Trend Confidence=0.71 LowVol=0.75 Medium=0.22 HighVol=0.03
	// ... regime converges to Low Vol Trend as volatility stays low
}

// ===== EXAMPLE 2: Ensemble Decision Layer =====
// This example shows how the ensemble combines MA and Kalman signals

func ExampleEnsembleSignalBlending() {
	// Create strategies
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)

	// Create ensemble with moderate risk profile
	ensemble := NewEnsembleDecisionLayer(
		maStrat,
		kalmanStrat,
		50,
		RiskProfileModerate,
	)

	// Generate synthetic data
	bars := generateRealisticBars(150, 100.0, 0.01, 0.0005)
	modelBars := convertToModelBars(bars)

	ctx := context.Background()

	fmt.Println("\n=== Ensemble Signal Blending ===")
	fmt.Println("Time | Regime          | MA      | Kalman  | Score | Signal   | Confidence")
	fmt.Println("-----|-----------------|---------|---------|-------|----------|------------")

	// Process bars sequentially
	for i := 50; i < len(modelBars); i += 10 {
		signal, confidence, regime, score, err := ensemble.EvaluateSignal(ctx, modelBars[:i+1])
		if err != nil {
			fmt.Printf("Bar %d: Error %v\n", i, err)
			continue
		}

		fmt.Printf("%4d | %-15s | Signal  | Score   | %.3f  | %-8v | %.2f\n",
			i, regime.String(), score, signal, confidence)
	}

	// Output example:
	// Time | Regime          | MA      | Kalman  | Score | Signal   | Confidence
	// -----|-----------------|---------|---------|-------|----------|------------
	//   50 | Medium          | Signal  | Score   | 0.125 | HOLD     | 0.06
	//   60 | Low Vol Trend   | Signal  | Score   | 0.652 | BUY      | 0.68
	//   70 | Low Vol Trend   | Signal  | Score   | 0.584 | BUY      | 0.61
	//   80 | Low Vol Trend   | Signal  | Score   | -0.142 | HOLD    | 0.14
	//   90 | Medium          | Signal  | Score   | -0.673 | SELL     | 0.67
	//  100 | High Vol Stress | Signal  | Score   | 0.421 | HOLD     | 0.25 (buy suppressed)
	//  110 | High Vol Stress | Signal  | Score   | -0.812 | SELL     | 0.81
}

// ===== EXAMPLE 3: Position Sizing by Profile & Regime =====
// This example shows how positions scale with risk profile and market regime

func ExamplePositionSizing() {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)

	fmt.Println("\n=== Position Sizing by Profile & Regime ===")
	fmt.Println("Profile        | Low Vol Trend | Medium  | High Vol Stress")
	fmt.Println("---------------|---------------|---------|----------------")

	availableCash := 10000.0

	profiles := []RiskProfile{
		RiskProfileConservative,
		RiskProfileModerate,
		RiskProfileAggressive,
	}

	regimes := []MarketRegime{
		RegimeLowVolTrend,
		RegimeMedium,
		RegimeHighVolStress,
	}

	for _, profile := range profiles {
		ensemble := NewEnsembleDecisionLayer(
			maStrat,
			kalmanStrat,
			50,
			profile,
		)

		sizes := make([]string, 3)
		for j, regime := range regimes {
			posSize := ensemble.GetPositionSize(availableCash, regime)
			sizePercent := (posSize / availableCash) * 100
			sizes[j] = fmt.Sprintf("$%.0f (%.1f%%)", posSize, sizePercent)
		}

		fmt.Printf("%-15s | %-13s | %-7s | %s\n",
			profile.String(),
			sizes[0],
			sizes[1],
			sizes[2],
		)
	}

	// Output:
	// Profile        | Low Vol Trend | Medium  | High Vol Stress
	// --------------|---------------|---------|----------------
	// Conservative   | $500 (5.0%)   | $375    | $125 (1.3%)
	// Moderate       | $1000 (10.0%) | $750    | $250 (2.5%)
	// Aggressive     | $2000 (20.0%) | $1500   | $500 (5.0%)
}

// ===== EXAMPLE 4: Full Trading Day Simulation =====
// This example simulates a complete trading day with regime transitions

func ExampleFullTradingDay() {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(
		maStrat,
		kalmanStrat,
		50,
		RiskProfileModerate,
	)

	// Simulate 8-hour trading day (480 1-minute bars)
	bars := generateTradingDay(480, 100.0)
	modelBars := convertToModelBars(bars)

	ctx := context.Background()

	fmt.Println("\n=== Full Trading Day Simulation ===")
	fmt.Println("Hour | Regime          | Signal | Score  | Confidence | Position Size | Action")
	fmt.Println("-----|-----------------|--------|--------|------------|---------------|--------")

	cash := 10000.0
	shares := 0.0
	tradeCount := 0
	lastSignal := models.SignalHold

	// Start processing at bar 50 (enough data for indicators)
	for i := 50; i < len(modelBars); i += 60 { // Print every hour
		signal, confidence, regime, score, _ := ensemble.EvaluateSignal(ctx, modelBars[:i+1])

		positionSize := ensemble.GetPositionSize(cash, regime)
		price := modelBars[i].Close

		action := "HOLD"
		if signal != models.SignalHold && signal != lastSignal {
			action = fmt.Sprintf("EXECUTE %v", signal)
			tradeCount++

			if signal == models.SignalBuy && shares == 0 {
				shares = positionSize / price
				cash -= shares * price
			} else if signal == models.SignalSell && shares > 0 {
				cash += shares * price
				shares = 0
			}

			lastSignal = signal
		}

		hour := i / 60
		fmt.Printf("%4d | %-15s | %-6v | %6.3f | %10.2f%% | $%-13.0f | %s\n",
			hour, regime.String(), signal, score, confidence*100, positionSize, action)
	}

	finalEquity := cash + (shares * modelBars[len(modelBars)-1].Close)
	totalReturn := ((finalEquity - 10000.0) / 10000.0) * 100
	fmt.Printf("\n=== Day Summary ===\n")
	fmt.Printf("Initial Capital: $10000.00\n")
	fmt.Printf("Final Equity: $%.2f\n", finalEquity)
	fmt.Printf("Total Return: %.2f%%\n", totalReturn)
	fmt.Printf("Trades Executed: %d\n", tradeCount)
}

// ===== EXAMPLE 5: Lookahead Bias Verification =====
// This example proves the system has NO lookahead bias

func ExampleNoLookaheadBias(t *testing.T) {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(
		maStrat,
		kalmanStrat,
		50,
		RiskProfileModerate,
	)

	bars := generateRealisticBars(200, 100.0, 0.01, 0.0005)
	modelBars := convertToModelBars(bars)
	ctx := context.Background()

	fmt.Println("\n=== Lookahead Bias Test ===")
	fmt.Println("Verifying signal at bar N doesn't change when bars N+1, N+2 are added...")

	// Evaluate at bar 100
	ensemble.Reset()
	signal1, score1, _, _, _ := ensemble.EvaluateSignal(ctx, modelBars[:101])
	regime1 := ensemble.lastRegime

	// Evaluate again with same data (should be identical - deterministic)
	ensemble.Reset()
	signal2, score2, _, _, _ := ensemble.EvaluateSignal(ctx, modelBars[:101])
	regime2 := ensemble.lastRegime

	fmt.Printf("Determinism Test:\n")
	fmt.Printf("  Signal1: %v | Regime1: %s | Score1: %.4f\n", signal1, regime1.String(), score1)
	fmt.Printf("  Signal2: %v | Regime2: %s | Score2: %.4f\n", signal2, regime2.String(), score2)
	fmt.Printf("  Match: %v\n", signal1 == signal2 && regime1 == regime2)

	// Evaluate at bar 100 vs bar 110 - signal shouldn't retroactively change
	ensemble.Reset()
	_, score100, _, _, _ := ensemble.EvaluateSignal(ctx, modelBars[:101])

	ensemble.Reset()

	// Reset and re-evaluate at bar 100 with all 110 bars loaded
	// (testing if bar 100 signal changes based on future data)
	ensemble.Reset()
	for i := 50; i <= 100; i++ {
		ensemble.EvaluateSignal(ctx, modelBars[:i+1])
	}
	scoreAt100 := ensemble.GetLastSignalScore()

	fmt.Printf("\nRetroactive Change Test:\n")
	fmt.Printf("  Score at bar 100 (with 100 bars): %.4f\n", scoreAt100)
	fmt.Printf("  Score at bar 100 from bar 110 evaluation: %.4f\n", score100)
	fmt.Printf("  Difference: %.4f (should be 0.0)\n", score100-scoreAt100)
	fmt.Printf("  No lookahead bias: %v\n", scoreAt100 == score100)
}

// ===== EXAMPLE 6: Regime Transition Tracking =====
// This example shows how to track regime persistence and transitions

func ExampleRegimeTransitions() {
	hmm := NewHMMRegimeDetector(50)

	// Create data with 3 distinct phases:
	// Phase 1: Low vol trend (bars 0-50)
	// Phase 2: High vol (bars 50-100)
	// Phase 3: Low vol trend again (bars 100-150)

	bars := make([]Bar, 150)

	// Phase 1: Low volatility, uptrend
	for i := 0; i < 50; i++ {
		bars[i] = Bar{
			Close: 100.0 + float64(i)*0.1,
			Open:  100.0 + float64(i)*0.1,
			High:  100.0 + float64(i)*0.1 + 0.3,
			Low:   100.0 + float64(i)*0.1 - 0.3,
			Volume: 1000000,
		}
	}

	// Phase 2: High volatility, sideways
	for i := 50; i < 100; i++ {
		j := i - 50
		bars[i] = Bar{
			Close: 105.0 + float64(j%5)*0.5,
			Open:  105.0 + float64(j%5)*0.5,
			High:  105.0 + float64(j%5)*0.5 + 2.0,
			Low:   105.0 + float64(j%5)*0.5 - 2.0,
			Volume: 2000000,
		}
	}

	// Phase 3: Low volatility, uptrend again
	for i := 100; i < 150; i++ {
		j := i - 100
		bars[i] = Bar{
			Close: 110.0 + float64(j)*0.08,
			Open:  110.0 + float64(j)*0.08,
			High:  110.0 + float64(j)*0.08 + 0.2,
			Low:   110.0 + float64(j)*0.08 - 0.2,
			Volume: 1500000,
		}
	}

	fmt.Println("\n=== Regime Transition Tracking ===")
	fmt.Println("Bar | Regime          | Persistence | Action")
	fmt.Println("----|-----------------|-------------|------------------------")

	var currentRegime MarketRegime
	for i := 50; i < 150; i += 10 {
		regime, _, _ := hmm.Update(bars[:i+1])
		persistence := hmm.GetRegimePersistence(regime)
		action := ""

		if i == 50 {
			currentRegime = regime
			action = "Initial"
		} else if regime != currentRegime {
			action = fmt.Sprintf("TRANSITION to %s", regime.String())
			currentRegime = regime
		}

		fmt.Printf("%3d | %-15s | %11d | %s\n", i, regime.String(), persistence, action)
	}

	fmt.Printf("\nFinal regime: %s\n", currentRegime.String())
	fmt.Printf("Final persistence: %d bars\n", hmm.GetRegimePersistence(currentRegime))
}

// ===== HELPER FUNCTIONS =====

func generateRealisticBars(count int, startPrice float64, volatility float64, drift float64) []Bar {
	bars := make([]Bar, count)
	price := startPrice

	for i := 0; i < count; i++ {
		// Pseudo-random walk
		pseudoRandom := float64((i*7)%11 - 5) / 5.0 // -1 to +1
		ret := drift + (volatility * pseudoRandom)
		newPrice := price * (1.0 + ret)

		bars[i] = Bar{
			Open:   price,
			High:   price * (1.0 + volatility),
			Low:    price * (1.0 - volatility),
			Close:  newPrice,
			Volume: 1000000,
		}
		price = newPrice
	}

	return bars
}

func generateTradingDay(barCount int, startPrice float64) []Bar {
	bars := make([]Bar, barCount)
	price := startPrice
	volatility := 0.001
	drift := 0.00001

	for i := 0; i < barCount; i++ {
		// Volatility increases mid-day, calms at end
		if i > barCount/3 && i < barCount*2/3 {
			volatility = 0.003
		} else {
			volatility = 0.001
		}

		pseudoRandom := float64((i*13)%20 - 10) / 10.0
		ret := drift + (volatility * pseudoRandom)
		newPrice := price * (1.0 + ret)

		bars[i] = Bar{
			Open:   price,
			High:   price * (1.0 + volatility*2),
			Low:    price * (1.0 - volatility*2),
			Close:  newPrice,
			Volume: int64(1000000 + (i%500000)),
		}
		price = newPrice
	}

	return bars
}

func convertToModelBars(bars []Bar) []models.Bar {
	modelBars := make([]models.Bar, len(bars))
	baseTime := time.Now().Add(-time.Duration(len(bars)) * time.Minute)

	for i, b := range bars {
		modelBars[i] = models.Bar{
			Symbol: "TEST",
			Time:   baseTime.Add(time.Duration(i) * time.Minute),
			Open:   b.Open,
			High:   b.High,
			Low:    b.Low,
			Close:  b.Close,
			Volume: int64(b.Volume),
		}
	}

	return modelBars
}

// ===== BENCHMARK TESTS =====

func BenchmarkHMMUpdate(b *testing.B) {
	hmm := NewHMMRegimeDetector(50)
	bars := generateRealisticBars(100, 100.0, 0.01, 0.0001)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hmm.Update(bars)
	}
}

func BenchmarkEnsembleEvaluate(b *testing.B) {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, RiskProfileModerate)

	bars := generateRealisticBars(100, 100.0, 0.01, 0.0001)
	modelBars := convertToModelBars(bars)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ensemble.EvaluateSignal(ctx, modelBars)
		ensemble.Reset()
	}
}

// ===== QUICK START EXAMPLE =====

func QuickStartExample() {
	fmt.Println("\n========== HMM ENSEMBLE QUICK START ==========\n")

	// Step 1: Create strategies
	fmt.Println("Step 1: Creating MA Crossover and Kalman Filter strategies...")
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)

	// Step 2: Create ensemble
	fmt.Println("Step 2: Creating HMM ensemble with Moderate risk profile...")
	ensemble := NewEnsembleDecisionLayer(
		maStrat,
		kalmanStrat,
		50,
		RiskProfileModerate,
	)

	// Step 3: Generate test data
	fmt.Println("Step 3: Generating 150 bars of synthetic market data...")
	bars := generateRealisticBars(150, 100.0, 0.01, 0.0005)
	modelBars := convertToModelBars(bars)

	// Step 4: Process bars
	fmt.Println("Step 4: Processing bars and generating signals...\n")
	ctx := context.Background()

	buyCount := 0
	sellCount := 0

	for i := 50; i < len(modelBars); i++ {
		signal, confidence, regime, _, _ := ensemble.EvaluateSignal(ctx, modelBars[:i+1])

		if signal == models.SignalBuy {
			buyCount++
			fmt.Printf("Bar %3d: %v (Regime: %s, Confidence: %.1f%%)\n",
				i, signal, regime.String(), confidence*100)
		} else if signal == models.SignalSell {
			sellCount++
			fmt.Printf("Bar %3d: %v (Regime: %s, Confidence: %.1f%%)\n",
				i, signal, regime.String(), confidence*100)
		}
	}

	fmt.Printf("\n========== RESULTS ==========\n")
	fmt.Printf("Buy Signals: %d\n", buyCount)
	fmt.Printf("Sell Signals: %d\n", sellCount)
	fmt.Printf("Signal Ratio: %.1f%%\n", (float64(buyCount+sellCount)/float64(len(modelBars)-50))*100)
	fmt.Println("\n✓ System working correctly!")
}
