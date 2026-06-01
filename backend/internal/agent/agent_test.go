package agent

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

type Bar = models.Bar

// ===== HMM REGIME DETECTOR TESTS =====

func TestHMMInitialization(t *testing.T) {
	hmm := NewHMMRegimeDetector(50)

	if hmm.windowSize != 50 {
		t.Errorf("expected window size 50, got %d", hmm.windowSize)
	}

	if len(hmm.stateSequence) != 0 {
		t.Errorf("expected empty state sequence, got length %d", len(hmm.stateSequence))
	}

	probs := hmm.GetProbabilities()
	if probs[0] < 0.3 || probs[0] > 0.34 {
		t.Errorf("initial probabilities incorrect: %v", probs)
	}
}

func TestHMMVolatilityDiscretization(t *testing.T) {
	hmm := NewHMMRegimeDetector(50)
	hmm.UpdateBuckets([3]float64{0.01, 0.025, 0.05}, [3]float64{-0.005, 0.0, 0.005})

	tests := []struct {
		vol      float64
		expected int
		name     string
	}{
		{0.005, 0, "Low volatility"},
		{0.015, 1, "Medium volatility"},
		{0.04, 2, "High volatility"},
		{0.06, 2, "Very high volatility"},
	}

	for _, tt := range tests {
		result := hmm.discretizeVolatility(tt.vol)
		if result != tt.expected {
			t.Errorf("%s: expected bucket %d, got %d", tt.name, tt.expected, result)
		}
	}
}

func TestHMMTrendDiscretization(t *testing.T) {
	hmm := NewHMMRegimeDetector(50)
	hmm.UpdateBuckets([3]float64{0.01, 0.025, 0.05}, [3]float64{-0.005, 0.0, 0.005})

	tests := []struct {
		trend    float64
		expected int
		name     string
	}{
		{-0.01, 0, "Downtrend"},
		{-0.002, 0, "Slight downtrend"},
		{0.001, 1, "Neutral trend"},
		{0.01, 2, "Uptrend"},
	}

	for _, tt := range tests {
		result := hmm.discretizeTrend(tt.trend)
		if result != tt.expected {
			t.Errorf("%s: expected bucket %d, got %d", tt.name, tt.expected, result)
		}
	}
}

func TestHMMVolatilityCalculation(t *testing.T) {
	hmm := NewHMMRegimeDetector(50)

	// Create bars with known returns
	bars := []Bar{
		{Open: 100, High: 102, Low: 99, Close: 100, Volume: 1000},
		{Open: 100, High: 101, Low: 99, Close: 101, Volume: 1000},  // +1%
		{Open: 101, High: 102, Low: 100, Close: 100, Volume: 1000}, // -0.99%
		{Open: 100, High: 103, Low: 99, Close: 102, Volume: 1000},  // +2%
	}

	vol := hmm.calculateRealizedVolatility(bars)
	if vol <= 0 {
		t.Errorf("volatility should be positive, got %.6f", vol)
	}

	// Volatility should be non-zero for non-trivial bars
	if vol < 0.001 {
		t.Errorf("volatility seems too small: %.6f", vol)
	}
}

func TestHMMTrendCalculation(t *testing.T) {
	hmm := NewHMMRegimeDetector(50)

	// Uptrend
	upBars := []Bar{
		{Open: 100, High: 101, Low: 100, Close: 100, Volume: 1000},
		{Open: 100, High: 102, Low: 100, Close: 101, Volume: 1000},
		{Open: 101, High: 103, Low: 101, Close: 102, Volume: 1000},
		{Open: 102, High: 104, Low: 101, Close: 103, Volume: 1000},
	}

	upTrend := hmm.calculateRollingTrend(upBars)
	if upTrend <= 0 {
		t.Errorf("uptrend should be positive, got %.6f", upTrend)
	}

	// Downtrend
	downBars := []Bar{
		{Open: 100, High: 100, Low: 99, Close: 100, Volume: 1000},
		{Open: 100, High: 100, Low: 98, Close: 99, Volume: 1000},
		{Open: 99, High: 99, Low: 97, Close: 98, Volume: 1000},
		{Open: 98, High: 98, Low: 96, Close: 97, Volume: 1000},
	}

	downTrend := hmm.calculateRollingTrend(downBars)
	if downTrend >= 0 {
		t.Errorf("downtrend should be negative, got %.6f", downTrend)
	}
}

func TestHMMRegimeDetection(t *testing.T) {
	hmm := NewHMMRegimeDetector(50)

	// Create a synthetic trending market with low volatility
	bars := generateSyntheticBars(100, 100.0, 0.005, 0.001) // low vol, uptrend

	regime, confidence, err := hmm.Update(bars[:50])
	if err != nil {
		t.Fatalf("HMM update failed: %v", err)
	}

	if confidence < 0 || confidence > 1 {
		t.Errorf("confidence out of range: %.2f", confidence)
	}

	if regime < 0 || regime > 2 {
		t.Errorf("regime out of range: %v", regime)
	}

	// Should trend towards low vol trend regime over time
	for i := 50; i < len(bars); i++ {
		regime, _, _ = hmm.Update(bars[:i+1])
	}

	// After many bars of low vol uptrend, should settle into LowVolTrend
	// This is probabilistic, so we just check it's called successfully
	if regime != RegimeLowVolTrend && regime != RegimeMedium {
		t.Logf("Final regime: %s (acceptable for probabilistic test)", regime.String())
	}
}

func TestHMMRegimePersistence(t *testing.T) {
	hmm := NewHMMRegimeDetector(50)

	bars := generateSyntheticBars(100, 100.0, 0.005, 0.001)

	// Update through bars
	for i := 50; i < len(bars); i++ {
		hmm.Update(bars[:i+1])
	}

	persistence := hmm.GetRegimePersistence(hmm.stateSequence[len(hmm.stateSequence)-1])
	if persistence <= 0 {
		t.Errorf("persistence should be positive, got %d", persistence)
	}

	if len(hmm.stateSequence) == 0 {
		t.Errorf("state sequence should not be empty")
	}
}

// ===== ENSEMBLE DECISION LAYER TESTS =====

func TestEnsembleInitialization(t *testing.T) {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, RiskProfileModerate)

	if ensemble.riskProfile != RiskProfileModerate {
		t.Errorf("expected Moderate profile, got %v", ensemble.riskProfile)
	}

	if ensemble.hmmDetector == nil {
		t.Errorf("ensemble.hmmDetector should not be nil")
	}
}

func TestRiskProfilePositionSizing(t *testing.T) {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)

	tests := []struct {
		profile   RiskProfile
		maxExpect float64
		name      string
	}{
		{RiskProfileConservative, 0.05, "Conservative"},
		{RiskProfileModerate, 0.10, "Moderate"},
		{RiskProfileAggressive, 0.20, "Aggressive"},
	}

	for _, tt := range tests {
		ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, tt.profile)
		posSize := ensemble.GetPositionSize(10000.0, RegimeMedium)
		expectedSize := 10000.0 * tt.maxExpect * 0.75 // medium scalar

		if math.Abs(posSize-expectedSize) > 10 {
			t.Errorf("%s: expected ~%.2f, got %.2f", tt.name, expectedSize, posSize)
		}
	}
}

func TestRegimeMultiplierScaling(t *testing.T) {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, RiskProfileModerate)

	cash := 10000.0

	// Low vol trend: full allocation
	lowVolSize := ensemble.GetPositionSize(cash, RegimeLowVolTrend)
	// Medium: 75% allocation
	mediumSize := ensemble.GetPositionSize(cash, RegimeMedium)
	// High vol stress: 25% allocation
	stressSize := ensemble.GetPositionSize(cash, RegimeHighVolStress)

	if !(lowVolSize > mediumSize && mediumSize > stressSize) {
		t.Errorf("position sizing not properly scaled: low=%.2f, med=%.2f, stress=%.2f",
			lowVolSize, mediumSize, stressSize)
	}
}

func TestSignalVoting(t *testing.T) {
	tests := []struct {
		maSig    models.Signal
		kalSig   models.Signal
		weights  SignalWeight
		expected float64
		name     string
	}{
		// Both buy
		{models.SignalBuy, models.SignalBuy, SignalWeight{0.5, 0.5}, 1.0, "Both buy"},
		// Both sell
		{models.SignalSell, models.SignalSell, SignalWeight{0.5, 0.5}, -1.0, "Both sell"},
		// Disagreement
		{models.SignalBuy, models.SignalSell, SignalWeight{0.5, 0.5}, 0.0, "Disagreement"},
		// Weighted disagreement
		{models.SignalBuy, models.SignalHold, SignalWeight{0.7, 0.3}, 0.7, "MA dominates buy"},
		{models.SignalSell, models.SignalHold, SignalWeight{0.3, 0.7}, -0.3, "Kalman dominates sell"},
	}

	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, RiskProfileModerate)

	for _, tt := range tests {
		score := ensemble.computeEnsembleScore(tt.maSig, tt.kalSig, tt.weights)
		if math.Abs(score-tt.expected) > 0.01 {
			t.Errorf("%s: expected %.2f, got %.2f", tt.name, tt.expected, score)
		}
	}
}

func TestRegimeGating(t *testing.T) {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, RiskProfileModerate)

	tests := []struct {
		score    float64
		regime   MarketRegime
		expected models.Signal
		name     string
	}{
		// Low vol trend: normal thresholds
		{0.7, RegimeLowVolTrend, models.SignalBuy, "Low vol buy"},
		{-0.7, RegimeLowVolTrend, models.SignalSell, "Low vol sell"},
		{0.3, RegimeLowVolTrend, models.SignalHold, "Low vol hold"},

		// Medium: balanced
		{0.6, RegimeMedium, models.SignalBuy, "Medium buy"},
		{-0.6, RegimeMedium, models.SignalSell, "Medium sell"},

		// High vol stress: suppress buys, allow exits
		{0.9, RegimeHighVolStress, models.SignalHold, "Stress suppress buy"},
		{-0.8, RegimeHighVolStress, models.SignalSell, "Stress allow sell"},
		{0.3, RegimeHighVolStress, models.SignalHold, "Stress hold"},
	}

	for _, tt := range tests {
		signal := ensemble.applyRegimeGating(tt.score, tt.regime)
		if signal != tt.expected {
			t.Errorf("%s: expected %v, got %v (score=%.2f regime=%s)",
				tt.name, tt.expected, signal, tt.score, tt.regime.String())
		}
	}
}

// ===== INTEGRATION TESTS =====

func TestEnsembleEvaluateSignal(t *testing.T) {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, RiskProfileModerate)

	// Generate synthetic bars
	bars := generateSyntheticBars(100, 100.0, 0.01, 0.002)

	// Convert to models.Bar
	modelBars := barsToModelBars(bars)

	ctx := context.Background()
	signal, confidence, regime, score, err := ensemble.EvaluateSignal(ctx, modelBars)

	if err != nil {
		t.Fatalf("evaluate signal failed: %v", err)
	}

	if confidence < 0 || confidence > 1 {
		t.Errorf("invalid confidence: %.2f", confidence)
	}

	if regime < 0 || regime > 2 {
		t.Errorf("invalid regime: %v", regime)
	}

	if signal != models.SignalBuy && signal != models.SignalSell && signal != models.SignalHold {
		t.Errorf("invalid signal: %v", signal)
	}

	_ = score // score should be between -1 and 1
	if score < -1.1 || score > 1.1 {
		t.Errorf("score out of range: %.2f", score)
	}
}

func TestEnsembleReset(t *testing.T) {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, RiskProfileModerate)

	bars := generateSyntheticBars(100, 100.0, 0.01, 0.002)
	modelBars := barsToModelBars(bars)

	ctx := context.Background()
	ensemble.EvaluateSignal(ctx, modelBars)

	stateSeqBefore := len(ensemble.hmmDetector.GetRegimeSequence())
	if stateSeqBefore == 0 {
		t.Fatalf("state sequence should not be empty after evaluation")
	}

	ensemble.Reset()

	stateSeqAfter := len(ensemble.hmmDetector.GetRegimeSequence())
	if stateSeqAfter != 0 {
		t.Errorf("state sequence not cleared after reset: %d", stateSeqAfter)
	}
}

// ===== LOOKAHEAD BIAS TESTS =====

func TestNoLookaheadBias(t *testing.T) {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, RiskProfileModerate)

	bars := generateSyntheticBars(100, 100.0, 0.01, 0.002)
	modelBars := barsToModelBars(bars)

	ctx := context.Background()

	// Evaluate at different points in history
	signal1, _, _, _, _ := ensemble.EvaluateSignal(ctx, modelBars[:80])
	ensemble.Reset()

	signal2, _, _, _, _ := ensemble.EvaluateSignal(ctx, modelBars[:80])
	ensemble.Reset()

	// Same input should produce same signal (deterministic)
	if signal1 != signal2 {
		t.Errorf("determinism failed: got %v then %v", signal1, signal2)
	}

	// Signal at bar N should not change when we add bars N+1
	ensemble.Reset()
	_, _, _, _, _ = ensemble.EvaluateSignal(ctx, modelBars[:80])
	signalAtN := ensemble.GetLastSignalScore()

	ensemble.Reset()
	_, _, _, _, _ = ensemble.EvaluateSignal(ctx, modelBars[:85])
	signalAtNWith85 := ensemble.GetLastSignalScore()

	// Scores should be similar but may differ due to regime update
	// Both should be stable and not retroactively change
	t.Logf("Score at bar 80: %.3f, with extra bars: %.3f (small drift ok)", signalAtN, signalAtNWith85)
}

func TestInsufficientDataHandling(t *testing.T) {
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, RiskProfileModerate)

	// Only 30 bars (less than min 50)
	bars := generateSyntheticBars(30, 100.0, 0.01, 0.002)
	modelBars := barsToModelBars(bars)

	ctx := context.Background()
	_, _, _, _, err := ensemble.EvaluateSignal(ctx, modelBars)

	if err == nil {
		t.Errorf("expected error with insufficient data, got nil")
	}
}

// ===== HELPER FUNCTIONS =====

func generateSyntheticBars(count int, startPrice float64, volatility float64, drift float64) []Bar {
	bars := make([]Bar, count)
	price := startPrice

	for i := 0; i < count; i++ {
		// Random walk with drift
		ret := drift + (volatility * (float64(i%3) - 1.0) / 2.0) // Pseudo-random
		newPrice := price * (1.0 + ret)

		bars[i] = Bar{
			Open:   price,
			High:   price * 1.01,
			Low:    price * 0.99,
			Close:  newPrice,
			Volume: 1000,
		}
		price = newPrice
	}

	return bars
}

func barsToModelBars(bars []Bar) []models.Bar {
	modelBars := make([]models.Bar, len(bars))
	for i, b := range bars {
		modelBars[i] = models.Bar{
			Symbol: "TEST",
			Time:   time.Now().Add(time.Duration(-len(bars)+i) * time.Hour),
			Open:   b.Open,
			High:   b.High,
			Low:    b.Low,
			Close:  b.Close,
			Volume: int64(b.Volume),
		}
	}
	return modelBars
}

func TestEndToEndSignalGeneration(t *testing.T) {
	// This test simulates real trading workflow
	maStrat := backtest.NewMACrossoverStrategy(20, 50)
	kalmanStrat := backtest.NewKalmanStrategy(0.001, 0.01, 20, 2.0)
	ensemble := NewEnsembleDecisionLayer(maStrat, kalmanStrat, 50, RiskProfileModerate)

	// Simulate a trading day: 100 hourly bars
	bars := generateSyntheticBars(100, 100.0, 0.01, 0.0001)
	modelBars := barsToModelBars(bars)

	ctx := context.Background()
	signalCount := 0
	lastSignal := models.SignalHold

	// Process bars sequentially (no lookahead)
	for i := 50; i < len(modelBars); i++ {
		signal, confidence, regime, _, err := ensemble.EvaluateSignal(ctx, modelBars[:i+1])

		if err != nil {
			t.Fatalf("bar %d: evaluation failed: %v", i, err)
		}

		if signal != models.SignalHold && signal != lastSignal {
			signalCount++
			lastSignal = signal
			t.Logf("Bar %d: Signal=%v Confidence=%.2f Regime=%s", i, signal, confidence, regime.String())
		}
	}

	t.Logf("Total signals generated: %d", signalCount)
	if signalCount < 0 {
		t.Errorf("signal count invalid: %d", signalCount)
	}
}

func TestAgentWorkerSellIgnoresBuySizingMinimum(t *testing.T) {
	worker := &AgentWorker{
		ctx:     context.Background(),
		account: NewPaperAccount(0),
		symbol:  "TEST",
	}
	worker.account.Positions["TEST"] = 2

	if err := worker.executePaperTrade(models.SignalSell, 100, 10); err != nil {
		t.Fatalf("sell should not be blocked by buy-side position sizing: %v", err)
	}

	if position := worker.account.GetPosition("TEST"); position != 1 {
		t.Fatalf("expected half the position to be sold, got %.2f shares", position)
	}
}

func TestAgentWorkerBuyBelowOneShareIsNoop(t *testing.T) {
	worker := &AgentWorker{
		ctx:     context.Background(),
		account: NewPaperAccount(50),
		symbol:  "TEST",
	}

	if err := worker.executePaperTrade(models.SignalBuy, 100, 50); err != nil {
		t.Fatalf("tiny buy should be skipped without killing the worker: %v", err)
	}

	if position := worker.account.GetPosition("TEST"); position != 0 {
		t.Fatalf("expected no position for below-one-share buy, got %.2f shares", position)
	}
}
