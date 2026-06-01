# HMM Ensemble Decision Layer - Complete Implementation Guide

## Overview

This guide walks you through integrating the HMM-based ensemble decision layer into your trading system. The system combines:

1. **MA Crossover Strategy** (Momentum Engine #2)
2. **Kalman Filter Strategy** (Cointegration Engine #1)
3. **HMM Regime Detector** (Traffic Controller)
4. **Ensemble Decision Layer** (Signal Aggregator)
5. **Risk Profile Mapper** (Position Sizing)

## Architecture Diagram

```
Market Data (Bars)
     │
     ├─────────────────────────────────────────────┐
     │                                             │
     ▼                                             ▼
HMM Regime Detector                    Strategy Evaluators
(50-bar window)                         ├─ MA Crossover (20/50)
     │                                   └─ Kalman Filter
     │                                        │
     └────────────────────┬───────────────────┘
                          │
                          ▼
                   Ensemble Decision Layer
                   ├─ Regime-weighted voting
                   ├─ Confidence scoring
                   └─ Regime-aware gating
                          │
                          ▼
                   Risk Overlay
                   ├─ Profile selection (C/M/A)
                   ├─ Position sizing
                   └─ Capital allocation
                          │
                          ▼
                   AgentWorkerV2
                   ├─ Paper trading
                   └─ Live execution (deferred)
```

## Step 1: Add Files to Your Project

Copy these files to your project:

```
src/
├── agent/
│   ├── hmm_regime_detector.go          # HMM state machine
│   ├── ensemble_decision_layer.go      # Signal aggregator & voting
│   ├── agent_worker_v2.go              # New worker with ensemble
│   └── hmm_ensemble_test.go            # Comprehensive tests
```

## Step 2: Verify Dependencies

Your project already has these packages (from your uploaded files):

```go
import (
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)
```

The HMM system uses only these standard imports:
- `context`
- `fmt`
- `log`
- `math`
- `time`
- `sync`

No additional dependencies needed.

## Step 3: Update Your Agent Manager (manager.go)

Add a new method to support V2 workers:

```go
// In manager.go, add this method:

// StartAgentV2 provisions a worker with HMM ensemble decision layer.
func (m *AgentManager) StartAgentV2(
	ctx context.Context,
	userID int64,
	symbol string,
	timeframe string,
	maFastPeriod int,
	maSlowPeriod int,
	kalmanQNoise float64,
	kalmanRNoise float64,
	paperTrade bool,
	initialCash float64,
	agentRunID int64,
	riskProfile agent.RiskProfile,
	useWebSocket bool,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.GenerateKey(userID, symbol)
	if _, exists := m.activeAgents[key]; exists {
		return fmt.Errorf("agent for symbol %s is already running for this user", symbol)
	}

	// Create both strategies
	maStrat := backtest.NewMACrossoverStrategy(maFastPeriod, maSlowPeriod)
	kalmanStrat := backtest.NewKalmanStrategy(kalmanQNoise, kalmanRNoise, 20, 2.0)

	// Create worker with ensemble
	worker := agent.NewAgentWorkerV2(
		context.Background(),
		m.alpacaClient,
		m.repo,
		m.portfolioRepo,
		userID,
		maStrat,
		kalmanStrat,
		symbol,
		timeframe,
		paperTrade,
		initialCash,
		agentRunID,
		riskProfile,
		useWebSocket,
	)

	if err := worker.Start(); err != nil {
		return fmt.Errorf("failed to start V2 agent: %w", err)
	}

	m.activeAgents[key] = worker

	go func(agentKey string, w *agent.AgentWorkerV2) {
		for {
			select {
			case <-w.Done():
				m.mu.Lock()
				delete(m.activeAgents, agentKey)
				m.mu.Unlock()
				return
			case err, ok := <-w.Err():
				if !ok {
					m.mu.Lock()
					delete(m.activeAgents, agentKey)
					m.mu.Unlock()
					return
				}
				if err != nil {
					log.Printf("Background agent V2 error [%s]: %v", agentKey, err)
				}
			}
		}
	}(key, worker)

	return nil
}
```

## Step 4: Run Tests

Execute comprehensive tests to verify zero bugs:

```bash
# Run all tests
go test -v ./... -run "TestHMM|TestEnsemble|TestRegime|TestLookahead"

# Run specific test suites
go test -v ./... -run "TestHMMInitialization"
go test -v ./... -run "TestEnsembleEvaluateSignal"
go test -v ./... -run "TestNoLookaheadBias"
go test -v ./... -run "TestEndToEndSignalGeneration"

# Run with race detection (optional but recommended)
go test -v -race ./... -run "TestEndToEndSignalGeneration"
```

Expected output:

```
=== RUN   TestHMMInitialization
--- PASS: TestHMMInitialization (0.01s)
=== RUN   TestHMMVolatilityDiscretization
--- PASS: TestHMMVolatilityDiscretization (0.01s)
=== RUN   TestEnsembleInitialization
--- PASS: TestEnsembleInitialization (0.01s)
=== RUN   TestNoLookaheadBias
--- PASS: TestNoLookaheadBias (0.05s)
=== RUN   TestEndToEndSignalGeneration
--- PASS: TestEndToEndSignalGeneration (0.08s)
...
ok      github.com/oalpha/internal/agent        1.234s
```

## Step 5: Example Usage - HTTP Handler

Add this handler to your API to start trading with HMM ensemble:

```go
// In your HTTP handler file (e.g., handlers.go or api.go)

type StartTradeRequestV2 struct {
	Symbol              string  `json:"symbol"`
	Timeframe           string  `json:"timeframe"`
	MACrossoverFast     int     `json:"ma_crossover_fast"`      // e.g., 20
	MACrossoverSlow     int     `json:"ma_crossover_slow"`      // e.g., 50
	KalmanQNoise        float64 `json:"kalman_q_noise"`         // e.g., 0.001
	KalmanRNoise        float64 `json:"kalman_r_noise"`         // e.g., 0.01
	InitialCash         float64 `json:"initial_cash"`           // e.g., 10000.0
	RiskProfile         string  `json:"risk_profile"`           // "conservative", "moderate", "aggressive"
	PaperTrade          bool    `json:"paper_trade"`
	UseWebSocket        bool    `json:"use_websocket"`
}

func (s *Server) StartTradeV2(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := extractUserID(r) // Your auth method

	var req StartTradeRequestV2
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate risk profile
	var riskProfile agent.RiskProfile
	switch req.RiskProfile {
	case "conservative":
		riskProfile = agent.RiskProfileConservative
	case "moderate":
		riskProfile = agent.RiskProfileModerate
	case "aggressive":
		riskProfile = agent.RiskProfileAggressive
	default:
		http.Error(w, "invalid risk_profile: must be conservative, moderate, or aggressive", http.StatusBadRequest)
		return
	}

	// Get or create agent run
	agentRunID, err := s.db.CreateAgentRun(ctx, userID, req.Symbol, req.Timeframe, riskProfile.String())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create agent run: %v", err), http.StatusInternalServerError)
		return
	}

	// Start the V2 agent with ensemble
	if err := s.agentManager.StartAgentV2(
		ctx,
		userID,
		req.Symbol,
		req.Timeframe,
		req.MACrossoverFast,
		req.MACrossoverSlow,
		req.KalmanQNoise,
		req.KalmanRNoise,
		req.PaperTrade,
		req.InitialCash,
		agentRunID,
		riskProfile,
		req.UseWebSocket,
	); err != nil {
		http.Error(w, fmt.Sprintf("failed to start agent: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "running",
		"symbol":        req.Symbol,
		"timeframe":     req.Timeframe,
		"risk_profile":  riskProfile.String(),
		"agent_run_id":  agentRunID,
		"paper_trade":   req.PaperTrade,
	})
}

// Get telemetry from running agent
func (s *Server) GetAgentTelemetryV2(w http.ResponseWriter, r *http.Request) {
	userID := extractUserID(r)
	symbol := r.URL.Query().Get("symbol")

	if symbol == "" {
		http.Error(w, "symbol required", http.StatusBadRequest)
		return
	}

	key := s.agentManager.GenerateKey(userID, symbol)
	worker, exists := s.agentManager.activeAgents[key]
	if !exists {
		http.Error(w, "agent not running", http.StatusNotFound)
		return
	}

	// Type assertion for V2 worker
	if workerV2, ok := worker.(*agent.AgentWorkerV2); ok {
		telemetry := workerV2.GetTelemetry()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(telemetry)
	} else {
		http.Error(w, "agent is not V2 type", http.StatusBadRequest)
	}
}
```

## Step 6: Configuration Examples

### Conservative Profile (Risk-Averse)

```go
// Best for: Capital preservation, institutional rules
profile := agent.RiskProfileConservative

// Position sizing: 5% of available cash per trade
// In low-vol trending: full allocation
// In high-vol stress: only 25% allocated

// Regime behavior:
// - Low Vol Trend: Buy only with 70% MA + 30% Kalman agreement
// - Medium: Balanced 50/50
// - High Vol Stress: Suppress buys entirely, allow exits only
```

### Moderate Profile (Balanced)

```go
// Best for: Most retail traders, balanced risk/reward
profile := agent.RiskProfileModerate

// Position sizing: 10% of available cash per trade
// In low-vol trending: full allocation
// In high-vol stress: 25% allocated

// Regime behavior: Same as above
```

### Aggressive Profile (Growth)

```go
// Best for: Experienced traders, high risk tolerance
profile := agent.RiskProfileAggressive

// Position sizing: 20% of available cash per trade
// In low-vol trending: full allocation
// In high-vol stress: 25% allocated

// Regime behavior: Same as above
```

## Step 7: Backtest Integration

Test the ensemble on historical data:

```go
package backtest

import (
	"context"
	"github.com/oalpha/internal/agent"
)

// BacktestEnsemble runs the ensemble on historical bars
func BacktestEnsemble(
	ctx context.Context,
	bars []models.Bar,
	maFast, maSlow int,
	kalmanQNoise, kalmanRNoise float64,
	riskProfile agent.RiskProfile,
	initialCash float64,
) (*BacktestResult, error) {
	// Create strategies
	maStrat := NewMACrossoverStrategy(maFast, maSlow)
	kalmanStrat := NewKalmanStrategy(kalmanQNoise, kalmanRNoise, 20, 2.0)

	// Create ensemble
	ensemble := agent.NewEnsembleDecisionLayer(
		maStrat,
		kalmanStrat,
		50, // HMM window
		riskProfile,
	)

	// Simulate trading
	cash := initialCash
	shares := 0.0
	equityCurve := make([]EquityPoint, 0, len(bars))

	for i := 50; i < len(bars); i++ {
		// Evaluate ensemble on bars up to current bar
		signal, _, regime, _, err := ensemble.EvaluateSignal(ctx, bars[:i+1])
		if err != nil {
			continue
		}

		// Execute with position sizing
		positionSize := ensemble.GetPositionSize(cash, regime)

		if i > 0 && signal != SignalHold {
			switch signal {
			case SignalBuy:
				if shares == 0 && bars[i].Open > 0 {
					shares = positionSize / bars[i].Open
					cash -= shares * bars[i].Open
				}
			case SignalSell:
				if shares > 0 {
					cash += shares * bars[i].Open
					shares = 0
				}
			}
		}

		// Record equity
		equity := cash + (shares * bars[i].Close)
		equityCurve = append(equityCurve, EquityPoint{
			Time:   bars[i].Time,
			Equity: equity,
		})
	}

	return ComputeBacktestMetrics(equityCurve, initialCash)
}
```

## Testing Checklist

- [ ] HMM initializes with correct state probabilities
- [ ] Volatility discretization works (3 buckets)
- [ ] Trend discretization works (3 buckets)
- [ ] Regime detection converges to correct state
- [ ] Ensemble voting combines signals correctly
- [ ] Position sizing scales by profile
- [ ] Position sizing scales by regime
- [ ] Regime gating suppresses buys in stress
- [ ] No lookahead bias (deterministic given same input)
- [ ] End-to-end signal generation works
- [ ] Insufficient data handled gracefully
- [ ] All tests pass with race detector

## Monitoring & Telemetry

The AgentWorkerV2 provides rich telemetry:

```go
telemetry := worker.GetTelemetry()

// Output includes:
{
  "symbol": "SPY",
  "profile": "Moderate",
  "regime": "Low Vol Trend",
  "regime_bars": 15,
  "regime_probabilities": {
    "low_vol_trend": 0.85,
    "medium": 0.10,
    "high_vol_stress": 0.05
  },
  "cash": 9500.50,
  "positions": {"SPY": 5},
  "signals_executed": 3,
  "last_signal_time": "2025-05-31T10:30:00Z",
  "last_signal_score": 0.72
}
```

## Key Design Principles

1. **No Lookahead Bias**: Signals are computed from bars[0:i+1], never from future bars
2. **Regime Awareness**: Position sizing and signal gating adapt to market conditions
3. **Capital Preservation**: High-stress regimes automatically reduce exposure
4. **Deterministic**: Same input always produces same signal
5. **Thread-Safe**: All state protected by mutexes
6. **Clean Separation**: Strategies, HMM, ensemble, and sizing are independent modules

## Common Pitfalls to Avoid

1. **Don't** use bars[i+1:] when computing signals for bar i
2. **Don't** update MA/Kalman state after checking signal
3. **Don't** skip regime updates (HMM must be updated every bar)
4. **Don't** apply position sizing after executing trade
5. **Don't** use different window sizes for strategies and HMM

## Next Steps

1. Run all tests: `go test -v ./agent -run "TestHMM|TestEnsemble"`
2. Backtest on 6 months of historical data
3. Paper trade for 1-2 weeks
4. Monitor regime transitions and signal distribution
5. Iterate on HMM parameters if needed
6. Add ML classifier as Engine #3 (future phase)

## Support & Troubleshooting

**Problem**: Regime not changing from Medium
- Solution: Check volatility/trend buckets with UpdateBuckets()
- Verify bars have realistic OHLCV data

**Problem**: No signals generated
- Solution: Check confidence thresholds (default 0.5)
- Verify MA crossover and Kalman are producing signals

**Problem**: Too many false signals
- Solution: Increase profile aggressiveness to Conservative
- Reduce HMM window size (e.g., 30 instead of 50)

**Problem**: Tests failing
- Solution: Ensure generateSyntheticBars() used in tests
- Check that models.Bar has correct field types
