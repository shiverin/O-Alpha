package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/oalpha/internal/alpaca"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/internal/marketdata"
	"github.com/oalpha/pkg/models"
)

// AgentWorkerV2 is the new agent worker that uses the HMM ensemble decision layer.
// It combines MA Crossover and Kalman Filter strategies with regime-aware signal gating.
type AgentWorkerV2 struct {
	ctx            context.Context
	cancelFunc     context.CancelFunc
	alpacaClient   *alpaca.Client
	repo           *db.BarsRepository
	portfolioRepo  *db.PortfolioRepository
	userID         int64
	ensemble       *EnsembleDecisionLayer
	symbol         string
	timeframe      string
	paperTrade     bool
	initialCash    float64
	agentRunID     int64
	account        *PaperAccount
	riskProfile    RiskProfile
	ticker         *time.Ticker
	stopOnce       sync.Once
	doneCh         chan struct{}
	errCh          chan error
	wsConnector    *marketdata.WebSocketConnector
	useWebSocket   bool
	barsMu         sync.RWMutex
	historicalBars []models.Bar
	maxBars        int

	// Telemetry
	lastSignalTime time.Time
	signalCount    int
	lastRegime     MarketRegime
}

// NewAgentWorkerV2 creates a worker with HMM ensemble decision layer.
func NewAgentWorkerV2(
	ctx context.Context,
	alpacaClient *alpaca.Client,
	repo *db.BarsRepository,
	portfolioRepo *db.PortfolioRepository,
	userID int64,
	maStrategy *backtest.MACrossoverStrategy,
	kalmanStrategy *backtest.KalmanStrategy,
	symbol string,
	timeframe string,
	paperTrade bool,
	initialCash float64,
	agentRunID int64,
	riskProfile RiskProfile,
	useWebSocket bool,
) *AgentWorkerV2 {
	wCtx, cancel := context.WithCancel(ctx)

	// Create ensemble with both strategies and HMM detector
	ensemble := NewEnsembleDecisionLayer(
		maStrategy,
		kalmanStrategy,
		50, // 50-bar window for HMM training
		riskProfile,
	)

	worker := &AgentWorkerV2{
		ctx:            wCtx,
		cancelFunc:     cancel,
		alpacaClient:   alpacaClient,
		repo:           repo,
		portfolioRepo:  portfolioRepo,
		userID:         userID,
		ensemble:       ensemble,
		symbol:         symbol,
		timeframe:      timeframe,
		paperTrade:     paperTrade,
		initialCash:    initialCash,
		agentRunID:     agentRunID,
		account:        NewPaperAccount(initialCash),
		riskProfile:    riskProfile,
		doneCh:         make(chan struct{}),
		errCh:          make(chan error, 1),
		useWebSocket:   useWebSocket,
		maxBars:        10000,
		lastRegime:     RegimeMedium,
	}

	if useWebSocket {
		worker.wsConnector = marketdata.NewWebSocketConnector(alpacaClient, []string{symbol}, timeframe)
	}

	return worker
}

// Start warms the ensemble and starts the worker loop.
func (w *AgentWorkerV2) Start() error {
	log.Printf("[Worker] Starting agent for %s (%s) with profile %s", w.symbol, w.timeframe, w.riskProfile.String())

	// Fetch historical data for warmup
	end := time.Now().UTC()
	start := end.Add(-720 * time.Hour) // 30 days

	bars, err := w.alpacaClient.GetBars(w.ctx, w.symbol, w.timeframe, start, end, 10000)
	if err != nil {
		return fmt.Errorf("failed to fetch initial data: %w", err)
	}

	if len(bars) < 50 {
		return fmt.Errorf("insufficient warmup data: have %d, need 50", len(bars))
	}

	// Warm the ensemble (generate initial signals)
	_, _, _, _, err = w.ensemble.EvaluateSignal(w.ctx, bars)
	if err != nil {
		return fmt.Errorf("ensemble warmup failed: %w", err)
	}

	w.setHistoricalBars(bars)

	if w.useWebSocket && w.wsConnector != nil {
		if err := w.wsConnector.Start(w.ctx); err != nil {
			return fmt.Errorf("failed to start WebSocket: %w", err)
		}
		go w.handleWebSocketData()
	} else {
		interval := timeframeToDuration(w.timeframe)
		w.ticker = time.NewTicker(interval)
		go w.runLoop()
	}

	return nil
}

// Stop terminates the worker.
func (w *AgentWorkerV2) Stop() {
	w.stopOnce.Do(func() {
		log.Printf("[Worker] Stopping agent for %s (signals: %d)", w.symbol, w.signalCount)
		w.cancelFunc()
		if w.ticker != nil {
			w.ticker.Stop()
		}
		if w.wsConnector != nil {
			w.wsConnector.Stop()
		}
		close(w.doneCh)
	})
}

// runLoop processes bars on a timer schedule.
func (w *AgentWorkerV2) runLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Worker] Panic: %v", r)
			select {
			case w.errCh <- fmt.Errorf("worker panicked: %v", r):
			default:
			}
		}
		close(w.errCh)
	}()

	for {
		select {
		case <-w.ctx.Done():
			log.Printf("[Worker] Context cancelled for %s", w.symbol)
			return
		case <-w.ticker.C:
			if err := w.processTick(); err != nil {
				log.Printf("[Worker] Error: %v", err)
				select {
				case w.errCh <- err:
				default:
				}
			}
		case <-w.doneCh:
			return
		}
	}
}

// handleWebSocketData processes real-time bars.
func (w *AgentWorkerV2) handleWebSocketData() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Worker] WebSocket panic: %v", r)
			select {
			case w.errCh <- fmt.Errorf("websocket panicked: %v", r):
			default:
			}
		}
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		case bar, ok := <-w.wsConnector.Data():
			if !ok {
				return
			}
			if err := w.processNewBar(bar); err != nil {
				log.Printf("[Worker] Bar processing error: %v", err)
				select {
				case w.errCh <- err:
				default:
				}
			}
		case err := <-w.wsConnector.Errors():
			log.Printf("[Worker] WebSocket error: %v", err)
			select {
			case w.errCh <- err:
			default:
			}
		case <-w.doneCh:
			return
		}
	}
}

// processNewBar appends a streamed bar and evaluates the ensemble.
func (w *AgentWorkerV2) processNewBar(bar models.Bar) error {
	w.appendOrUpdateBar(bar)
	historicalBars := w.getHistoricalBarsSnapshot()

	if len(historicalBars) < 50 {
		return nil // Not enough data yet
	}

	return w.evaluateAndExecute(historicalBars)
}

// processTick fetches recent bars and evaluates the ensemble.
func (w *AgentWorkerV2) processTick() error {
	end := time.Now().UTC()
	start := end.Add(-2 * timeframeToDuration(w.timeframe))

	bars, err := w.alpacaClient.GetBars(w.ctx, w.symbol, w.timeframe, start, end, 10000)
	if err != nil {
		return fmt.Errorf("fetch bars failed: %w", err)
	}

	if len(bars) == 0 {
		return nil
	}

	w.mergeBars(bars)
	historicalBars := w.getHistoricalBarsSnapshot()

	if len(historicalBars) < 50 {
		return nil // Not enough data yet
	}

	return w.evaluateAndExecute(historicalBars)
}

// evaluateAndExecute runs the ensemble and executes signals with position sizing.
func (w *AgentWorkerV2) evaluateAndExecute(bars []models.Bar) error {
	// Evaluate the ensemble decision layer
	signal, confidence, regime, score, err := w.ensemble.EvaluateSignal(w.ctx, bars)
	if err != nil {
		return fmt.Errorf("ensemble evaluation failed: %w", err)
	}

	w.lastRegime = regime
	latestBar := bars[len(bars)-1]

	// De-duplicate signals (only execute if signal changed or first time)
	if signal == models.SignalHold {
		if w.paperTrade {
			cash, positions := w.account.Snapshot()
			log.Printf("[Worker] HOLD - Symbol: %s, Regime: %s, Cash: %.2f, Positions: %v, Confidence: %.2f",
				w.symbol, regime.String(), cash, positions, confidence)
		}
		return w.persistPortfolioTelemetry(latestBar.Close)
	}

	// Calculate position size based on regime and risk profile
	positionSize := w.ensemble.GetPositionSize(w.account.Cash, regime)

	if err := w.executeTrade(signal, latestBar.Close, positionSize); err != nil {
		return fmt.Errorf("trade execution failed: %w", err)
	}

	w.signalCount++
	w.lastSignalTime = time.Now()

	if w.paperTrade {
		cash, positions := w.account.Snapshot()
		log.Printf("[Worker] %v - Symbol: %s, Regime: %s(%d bars), Price: %.2f, PositionSize: %.2f, Cash: %.2f, Positions: %v, Score: %.3f, Confidence: %.2f",
			signal, w.symbol, regime.String(), w.ensemble.lastRegimeBars, latestBar.Close, positionSize, cash, positions, score, confidence)
	}

	return w.persistPortfolioTelemetry(latestBar.Close)
}

// executeTrade applies the signal with position sizing.
func (w *AgentWorkerV2) executeTrade(signal models.Signal, price float64, positionSize float64) error {
	if w.paperTrade {
		return w.executePaperTrade(signal, price, positionSize)
	}
	return w.executeLiveTrade(signal, price, positionSize)
}

// executePaperTrade applies signal to in-memory account.
func (w *AgentWorkerV2) executePaperTrade(signal models.Signal, price float64, positionSize float64) error {
	if positionSize < price { // Minimum 1 share
		return fmt.Errorf("position size %.2f is less than price %.2f", positionSize, price)
	}

	switch signal {
	case models.SignalBuy:
		currentPos := w.account.GetPosition(w.symbol)
		if currentPos > 0 {
			return nil // Already long
		}

		quantity := positionSize / price
		filledQty, _, err := w.account.Buy(w.ctx, w.symbol, price, quantity)
		if err != nil {
			return fmt.Errorf("buy failed: %w", err)
		}

		if err := w.recordPaperFill("BUY_LONG", price, filledQty); err != nil {
			return err
		}

	case models.SignalSell:
		currentPos := w.account.GetPosition(w.symbol)
		if currentPos <= 0 {
			return nil // No position to sell
		}

		quantity := currentPos * 0.5 // Sell 50% of position
		if quantity < 1.0 {
			quantity = currentPos
		}

		filledQty, _, err := w.account.Sell(w.ctx, w.symbol, price, quantity)
		if err != nil {
			return fmt.Errorf("sell failed: %w", err)
		}

		if err := w.recordPaperFill("SELL_LONG", price, filledQty); err != nil {
			return err
		}
	}

	return nil
}

// executeLiveTrade logs live trades (safety inert until sizing finalized)
func (w *AgentWorkerV2) executeLiveTrade(signal models.Signal, price float64, positionSize float64) error {
	quantity := positionSize / price
	log.Printf("[Worker] Live trade would execute: %v %f shares at $%.2f", signal, quantity, price)
	return nil
}

// recordPaperFill persists trade to database.
func (w *AgentWorkerV2) recordPaperFill(action string, price, qty float64) error {
	if w.portfolioRepo == nil || w.userID <= 0 {
		return nil
	}
	if err := w.portfolioRepo.RecordLongFill(w.ctx, w.userID, w.agentRunID, action, w.symbol, price, qty, 0); err != nil {
		return fmt.Errorf("persist fill: %w", err)
	}
	return nil
}

// persistPortfolioTelemetry updates portfolio snapshot.
func (w *AgentWorkerV2) persistPortfolioTelemetry(currentPrice float64) error {
	if !w.paperTrade || w.portfolioRepo == nil || w.userID <= 0 {
		return nil
	}

	if err := w.portfolioRepo.MarkPositionPrice(w.ctx, w.userID, w.symbol, currentPrice); err != nil {
		return err
	}

	prices := map[string]float64{w.symbol: currentPrice}
	equity := w.account.Equity(w.ctx, prices)
	if err := w.portfolioRepo.SavePortfolioSnapshot(w.ctx, w.userID, equity, w.initialCash); err != nil {
		return err
	}
	return nil
}

// Err returns the error channel.
func (w *AgentWorkerV2) Err() <-chan error {
	return w.errCh
}

// Done returns a channel that closes when the worker stops.
func (w *AgentWorkerV2) Done() <-chan struct{} {
	return w.doneCh
}

// GetTelemetry returns current worker statistics.
func (w *AgentWorkerV2) GetTelemetry() map[string]interface{} {
	regime, regimeBars := w.ensemble.GetRegimeInfo()
	probs := w.ensemble.GetStateProbabilities()
	cash, positions := w.account.Snapshot()

	return map[string]interface{}{
		"symbol":              w.symbol,
		"profile":             w.riskProfile.String(),
		"regime":              regime.String(),
		"regime_bars":         regimeBars,
		"regime_probabilities": map[string]float64{
			"low_vol_trend":     probs[0],
			"medium":            probs[1],
			"high_vol_stress":   probs[2],
		},
		"cash":               cash,
		"positions":          positions,
		"signals_executed":   w.signalCount,
		"last_signal_time":   w.lastSignalTime,
		"last_signal_score":  w.ensemble.GetLastSignalScore(),
	}
}

// BarManagement methods (same as V1)

func (w *AgentWorkerV2) setHistoricalBars(bars []models.Bar) {
	w.barsMu.Lock()
	defer w.barsMu.Unlock()
	w.historicalBars = append([]models.Bar(nil), bars...)
	w.trimHistoricalBarsLocked()
}

func (w *AgentWorkerV2) getHistoricalBarsSnapshot() []models.Bar {
	w.barsMu.RLock()
	defer w.barsMu.RUnlock()
	return append([]models.Bar(nil), w.historicalBars...)
}

func (w *AgentWorkerV2) appendOrUpdateBar(bar models.Bar) {
	w.barsMu.Lock()
	defer w.barsMu.Unlock()

	n := len(w.historicalBars)
	if n > 0 && w.historicalBars[n-1].Time.Equal(bar.Time) {
		w.historicalBars[n-1] = bar
	} else {
		w.historicalBars = append(w.historicalBars, bar)
	}
	w.trimHistoricalBarsLocked()
}

func (w *AgentWorkerV2) mergeBars(newBars []models.Bar) {
	w.barsMu.Lock()
	defer w.barsMu.Unlock()

	if len(w.historicalBars) == 0 {
		w.historicalBars = append([]models.Bar(nil), newBars...)
		w.trimHistoricalBarsLocked()
		return
	}

	indexByTime := make(map[time.Time]int, len(w.historicalBars))
	for i, bar := range w.historicalBars {
		indexByTime[bar.Time] = i
	}

	for _, bar := range newBars {
		if i, ok := indexByTime[bar.Time]; ok {
			w.historicalBars[i] = bar
			continue
		}
		w.historicalBars = append(w.historicalBars, bar)
		indexByTime[bar.Time] = len(w.historicalBars) - 1
	}

	w.trimHistoricalBarsLocked()
}

func (w *AgentWorkerV2) trimHistoricalBarsLocked() {
	if w.maxBars <= 0 || len(w.historicalBars) <= w.maxBars {
		return
	}
	w.historicalBars = w.historicalBars[len(w.historicalBars)-w.maxBars:]
}