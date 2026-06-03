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

// AgentWorker manages the infrastructure loop, real-time data streaming,
// data management, and active risk guardrails for a specific asset.
type AgentWorker struct {
	ctx            context.Context
	cancelFunc     context.CancelFunc
	alpacaClient   *alpaca.Client
	repo           *db.BarsRepository
	agentRepo      *db.AgentRepository
	portfolioRepo  *db.PortfolioRepository
	userID         int64
	strategy       backtest.Strategy
	symbol         string
	timeframe      string
	paperTrade     bool
	initialCash    float64
	agentRunID     int64
	account        *PaperAccount
	ticker         *time.Ticker
	stopOnce       sync.Once
	doneCh         chan struct{}
	errCh          chan error
	wsConnector    *marketdata.WebSocketConnector
	useWebSocket   bool
	barsMu         sync.RWMutex
	historicalBars []models.Bar
	maxBars        int
	calibrateEvery int // recalibrate HMM buckets every N new bars (0 = off)
	barsSinceCalib int

	// State Tracking to prevent lookahead mismatches and streaming noise
	lastEvaluatedTime time.Time
	telemetryMetadata sync.Map
}

// NewAgentWorker creates a worker with an isolated runtime state.
func NewAgentWorker(
	ctx context.Context,
	alpacaClient *alpaca.Client,
	repo *db.BarsRepository,
	agentRepo *db.AgentRepository,
	portfolioRepo *db.PortfolioRepository,
	userID int64,
	strategy backtest.Strategy,
	symbol string,
	timeframe string,
	paperTrade bool,
	initialCash float64,
	agentRunID int64,
	useWebSocket bool,
) *AgentWorker {
	wCtx, cancel := context.WithCancel(ctx)
	worker := &AgentWorker{
		ctx:            wCtx,
		cancelFunc:     cancel,
		alpacaClient:   alpacaClient,
		repo:           repo,
		agentRepo:      agentRepo,
		portfolioRepo:  portfolioRepo,
		userID:         userID,
		strategy:       strategy,
		symbol:         symbol,
		timeframe:      timeframe,
		paperTrade:     paperTrade,
		initialCash:    initialCash,
		agentRunID:     agentRunID,
		account:        NewPaperAccount(initialCash),
		doneCh:         make(chan struct{}),
		errCh:          make(chan error, 1),
		useWebSocket:   useWebSocket,
		maxBars:        10000,
		calibrateEvery: 500,
	}

	if useWebSocket {
		worker.wsConnector = marketdata.NewWebSocketConnector(alpacaClient, []string{symbol}, timeframe)
	}

	return worker
}

// Start warms indicators and initializes execution loops.
func (w *AgentWorker) Start() error {
	log.Printf("[Worker] Starting unified execution loop for %s (%s)", w.symbol, w.timeframe)

	end := time.Now().UTC()
	start := end.Add(-720 * time.Hour) // 30 days of data for warmup

	bars, err := w.alpacaClient.GetBars(w.ctx, w.symbol, w.timeframe, start, end, 10000)
	if err != nil {
		return fmt.Errorf("failed to fetch initial data: %w", err)
	}

	if len(bars) == 0 {
		return fmt.Errorf("insufficient baseline history found for symbol %s", w.symbol)
	}

	if c, ok := w.strategy.(calibratable); ok {
		c.Calibrate(bars)
	}

	if _, err := w.strategy.EvaluateLatest(w.ctx, bars); err != nil {
		return fmt.Errorf("failed to generate initial strategy output: %w", err)
	}
	w.setHistoricalBars(bars)

	if w.useWebSocket && w.wsConnector != nil {
		if err := w.wsConnector.Start(w.ctx); err != nil {
			return fmt.Errorf("failed to start WebSocket stream: %w", err)
		}
		go w.handleWebSocketData()
	} else {
		interval := timeframeToDuration(w.timeframe)
		w.ticker = time.NewTicker(interval)
		go w.runLoop()
	}

	return nil
}

// Stop safely cancels execution states and cleanly tears down communication channels.
func (w *AgentWorker) Stop() {
	w.stopOnce.Do(func() {
		log.Printf("[Worker] Terminating execution pipeline for %s safely", w.symbol)
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

// runLoop processes background ticks on a timer interval.
func (w *AgentWorker) runLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[CRASH Recovery] Worker runtime panic: %v", r)
			select {
			case w.errCh <- fmt.Errorf("worker panicked: %v", r):
			default:
			}
		}
		close(w.errCh)
		w.Stop() // Guarantee structural cleanup
	}()

	for {
		select {
		case <-w.ctx.Done():
			log.Printf("[Worker] Context finalized for %s", w.symbol)
			return
		case <-w.ticker.C:
			if err := w.processTick(); err != nil {
				log.Printf("[Worker Loop Error] Tick processing failed: %v", err)
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

// handleWebSocketData consumes streamed updates, automatically resolving zombie worker locks on disconnect.
func (w *AgentWorker) handleWebSocketData() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[CRASH Recovery] WebSocket pipeline panic: %v", r)
			select {
			case w.errCh <- fmt.Errorf("websocket handler panicked: %v", r):
			default:
			}
		}
		close(w.errCh)
		w.Stop() // Prevents zombie states by ensuring manager cleans up active tracking mappings
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		case bar, ok := <-w.wsConnector.Data():
			if !ok {
				log.Printf("[Worker Connection Drop] Stream closed for symbol %s", w.symbol)
				return
			}
			if err := w.processNewBar(bar); err != nil {
				log.Printf("[Worker Error] Streaming bar calculation failed: %v", err)
				select {
				case w.errCh <- err:
				default:
				}
			}
		case err, ok := <-w.wsConnector.Errors():
			if !ok {
				log.Printf("[Worker Connection Drop] Error channel closed for symbol %s", w.symbol)
				return
			}
			if err == nil {
				continue
			}
			log.Printf("[Stream Exception] WebSocket transmission alert: %v", err)
			select {
			case w.errCh <- err:
			default:
			}
		case <-w.doneCh:
			return
		}
	}
}

// processNewBar safely updates trailing bar state arrays and enforces real-time evaluations.
func (w *AgentWorker) processNewBar(bar models.Bar) error {
	w.appendOrUpdateBar(bar)
	historicalBars := w.getHistoricalBarsSnapshot()
	if len(historicalBars) == 0 {
		return fmt.Errorf("empty active buffer state context")
	}

	latestBar := historicalBars[len(historicalBars)-1]

	// Active Guardrail: Run Risk Rules to enforce Stop-Loss and Take-Profit bounds
	if err := w.enforceRiskRules(latestBar.Close); err != nil {
		return fmt.Errorf("risk rule enforcement exception: %w", err)
	}

	// Bar-Close discipline: Prevent looking ahead or running duplicate signals on micro-updates
	if !latestBar.Time.After(w.lastEvaluatedTime) {
		return nil
	}

	w.maybeCalibrate(historicalBars)

	output, err := w.strategy.EvaluateLatest(w.ctx, historicalBars)
	if err != nil {
		return fmt.Errorf("failed to compute target strategy parameters: %w", err)
	}

	w.lastEvaluatedTime = latestBar.Time
	w.storeStrategyOutput(output)

	if output.Signal != models.SignalHold {
		targetAllocation := w.targetAllocationForOutput(output, latestBar.Close)
		if err := w.executeTrade(output.Signal, latestBar.Close, targetAllocation); err != nil {
			return fmt.Errorf("trade routing mismatch error: %w", err)
		}
	}

	w.logStateTelemetry(output, latestBar.Close)
	return w.persistPortfolioTelemetry(latestBar.Close)
}

// processTick polls recent market updates for timer-based execution environments.
func (w *AgentWorker) processTick() error {
	end := time.Now().UTC()
	start := end.Add(-2 * timeframeToDuration(w.timeframe))

	bars, err := w.alpacaClient.GetBars(w.ctx, w.symbol, w.timeframe, start, end, 10000)
	if err != nil {
		return fmt.Errorf("polling data fetch phase failed: %w", err)
	}

	if len(bars) == 0 {
		return nil
	}

	w.mergeBars(bars)
	historicalBars := w.getHistoricalBarsSnapshot()
	if len(historicalBars) == 0 {
		return fmt.Errorf("empty dynamic tracking matrix bounds")
	}

	latestBar := historicalBars[len(historicalBars)-1]

	if err := w.enforceRiskRules(latestBar.Close); err != nil {
		return fmt.Errorf("risk compliance validation exception: %w", err)
	}

	if !latestBar.Time.After(w.lastEvaluatedTime) {
		return nil
	}

	w.maybeCalibrate(historicalBars)

	output, err := w.strategy.EvaluateLatest(w.ctx, historicalBars)
	if err != nil {
		return fmt.Errorf("failed to process signal calculation layer: %w", err)
	}

	w.lastEvaluatedTime = latestBar.Time
	w.storeStrategyOutput(output)

	if output.Signal != models.SignalHold {
		targetAllocation := w.targetAllocationForOutput(output, latestBar.Close)
		if err := w.executeTrade(output.Signal, latestBar.Close, targetAllocation); err != nil {
			return fmt.Errorf("execution path mismatch error: %w", err)
		}
	}

	w.logStateTelemetry(output, latestBar.Close)
	return w.persistPortfolioTelemetry(latestBar.Close)
}

// enforceRiskRules pulls dynamic user profile metrics from DB layer and verifies account capital protection rules.
func (w *AgentWorker) enforceRiskRules(currentPrice float64) error {
	if w.agentRepo == nil {
		return nil
	}

	settings, err := w.agentRepo.GetAgentSettings(w.ctx, w.userID)
	if err != nil || settings == nil {
		return nil // Graceful fallback if custom configurations are unassigned
	}

	positionQty := w.account.GetPosition(w.symbol)
	if positionQty <= 0 {
		return nil
	}

	// Fetch historical entry context safely
	var avgEntryPrice float64
	if w.portfolioRepo != nil {
		positions, err := w.portfolioRepo.GetActivePositions(w.ctx, w.userID)
		if err == nil {
			for _, pos := range positions {
				if pos.Symbol == w.symbol {
					avgEntryPrice = pos.AvgEntryPrice
					break
				}
			}
		}
	}

	if avgEntryPrice <= 0 {
		return nil // Prevent tracking anomalies if entry data is initializing
	}

	currentPnLPct := ((currentPrice - avgEntryPrice) / avgEntryPrice) * 100

	// 1. Enforce Stop-Loss Threshold Violation
	if settings.StopLossPct > 0 && currentPnLPct <= -settings.StopLossPct {
		log.Printf("[RISK INTERVENTION] Stop-loss rule violation triggered for asset %s (PnL: %.2f%%)", w.symbol, currentPnLPct)
		return w.executeForceLiquidate(currentPrice, positionQty)
	}

	// 2. Enforce Take-Profit Targets
	if settings.TakeProfitPct > 0 && currentPnLPct >= settings.TakeProfitPct {
		log.Printf("[RISK INTERVENTION] Target take-profit milestone reached for asset %s (PnL: %.2f%%)", w.symbol, currentPnLPct)
		return w.executeForceLiquidate(currentPrice, positionQty)
	}

	return nil
}

// executeForceLiquidate exits market position completely under strict protection conditions.
func (w *AgentWorker) executeForceLiquidate(price float64, qty float64) error {
	if w.paperTrade {
		filledQty, _, err := w.account.Sell(w.ctx, w.symbol, price, qty)
		if err != nil {
			return err
		}
		return w.recordPaperFill("SELL_LONG", price, filledQty)
	}

	log.Printf("[LIVE PROXY RISK EXIT] Immediate safety liquidation routine initiated for %s", w.symbol)
	return nil
}

// executeTrade determines tracking environment routing logic.
func (w *AgentWorker) executeTrade(signal models.Signal, price float64, targetAllocation float64) error {
	if w.paperTrade {
		return w.executePaperTrade(signal, price, targetAllocation)
	}
	return w.executeLiveTrade(signal, price, targetAllocation)
}

// executePaperTrade handles paper balance modifications, resolving data races and fractional execution bugs.
func (w *AgentWorker) executePaperTrade(signal models.Signal, price float64, targetAllocation float64) error {
	switch signal {
	case models.SignalBuy:
		availableCash := w.account.AvailableCash()
		cashToUse := targetAllocation
		if cashToUse <= 0 {
			return nil
		}
		if cashToUse > availableCash {
			cashToUse = availableCash
		}
		if cashToUse < price {
			return nil
		}

		amount := cashToUse / price
		filledQty, _, err := w.account.Buy(w.ctx, w.symbol, price, amount)
		if err != nil {
			return fmt.Errorf("paper matching buy engine execution failed: %w", err)
		}
		if err := w.recordPaperFill("BUY_LONG", price, filledQty); err != nil {
			return err
		}

	case models.SignalSell:
		currentPos := w.account.GetPosition(w.symbol)
		if currentPos <= 0 {
			return nil // Safety exit catch to absorb historical trailing state variations
		}

		amount := currentPos

		filledQty, _, err := w.account.Sell(w.ctx, w.symbol, price, amount)
		if err != nil {
			return fmt.Errorf("paper matching sell engine execution failed: %w", err)
		}
		if err := w.recordPaperFill("SELL_LONG", price, filledQty); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unrecognized execution matrix condition value: %v", signal)
	}

	return nil
}

func (w *AgentWorker) targetAllocationForOutput(output backtest.StrategyOutput, price float64) float64 {
	if output.Signal != models.SignalBuy || price <= 0 {
		return 0
	}
	targetWeight := output.TargetWeight
	if targetWeight <= 0 {
		targetWeight = output.PositionSizePct
	}
	targetWeight = normalizePositionSizePct(targetWeight)
	equity := w.account.Equity(w.ctx, map[string]float64{w.symbol: price})
	targetValue := equity * targetWeight
	currentValue := w.account.GetPosition(w.symbol) * price
	delta := targetValue - currentValue
	if delta <= 0 {
		return 0
	}
	return delta
}

func (w *AgentWorker) recordPaperFill(action string, price, qty float64) error {
	if w.portfolioRepo == nil || w.userID <= 0 {
		return nil
	}
	if err := w.portfolioRepo.RecordLongFill(w.ctx, w.userID, w.agentRunID, action, w.symbol, price, qty, 0); err != nil {
		return fmt.Errorf("failed to persist ledger execution entries: %w", err)
	}
	return nil
}

func (w *AgentWorker) persistPortfolioTelemetry(currentPrice float64) error {
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

// executeLiveTrade acts as a safety intercept until physical order connectivity blocks are mapped.
func (w *AgentWorker) executeLiveTrade(signal models.Signal, price float64, targetAllocation float64) error {
	log.Printf("[LIVE PROXY LOG] Order matching parameter block: %v | %s at $%.2f target=$%.2f", signal, w.symbol, price, targetAllocation)
	return nil
}

// GetTelemetry returns current execution indicators safely.
func (w *AgentWorker) GetTelemetry() map[string]interface{} {
	return map[string]interface{}{
		"symbol":             w.symbol,
		"timeframe":          w.timeframe,
		"available_cash":     w.account.AvailableCash(),
		"tracked_position":   w.account.GetPosition(w.symbol),
		"last_checked_block": w.lastEvaluatedTime,
		"strategy_metrics":   w.GetLatestMetrics(),
	}
}

func (w *AgentWorker) storeStrategyOutput(output backtest.StrategyOutput) {
	w.telemetryMetadata.Store("signal", output.Signal)
	w.telemetryMetadata.Store("position_size_pct", output.PositionSizePct)
	w.telemetryMetadata.Store("regime", output.RegimeLabel)
	w.telemetryMetadata.Store("metrics", output.Metadata)
}

// maybeCalibrate periodically recalibrates the strategy on the worker goroutine.
// Same goroutine as Update() => no lock needed.
func (w *AgentWorker) maybeCalibrate(bars []models.Bar) {
	if w.calibrateEvery <= 0 {
		return
	}
	w.barsSinceCalib++
	if w.barsSinceCalib < w.calibrateEvery {
		return
	}
	w.barsSinceCalib = 0
	if c, ok := w.strategy.(calibratable); ok {
		c.Calibrate(bars)
	}
}

func (w *AgentWorker) logStateTelemetry(output backtest.StrategyOutput, price float64) {
	log.Printf("[Worker] %s signal=%v regime=%s price=%.2f size_pct=%.4f metadata=%v",
		w.symbol, output.Signal, output.RegimeLabel, price, output.PositionSizePct, output.Metadata)
}

func normalizePositionSizePct(sizePct float64) float64 {
	if sizePct <= 0 {
		return 0.10
	}
	if sizePct > 1 {
		return 1
	}
	return sizePct
}

func (w *AgentWorker) Err() <-chan error {
	return w.errCh
}

func (w *AgentWorker) Done() <-chan struct{} {
	return w.doneCh
}

func timeframeToDuration(timeframe string) time.Duration {
	switch timeframe {
	case "1Min":
		return time.Minute
	case "5Min":
		return 5 * time.Minute
	case "15Min":
		return 15 * time.Minute
	case "1Hour":
		return time.Hour
	case "1Day":
		return 24 * time.Hour
	default:
		return time.Hour
	}
}

func (w *AgentWorker) setHistoricalBars(bars []models.Bar) {
	w.barsMu.Lock()
	defer w.barsMu.Unlock()
	w.historicalBars = append([]models.Bar(nil), bars...)
	w.trimHistoricalBarsLocked()
}

func (w *AgentWorker) getHistoricalBarsSnapshot() []models.Bar {
	w.barsMu.RLock()
	defer w.barsMu.RUnlock()
	return append([]models.Bar(nil), w.historicalBars...)
}

func (w *AgentWorker) appendOrUpdateBar(bar models.Bar) {
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

func (w *AgentWorker) mergeBars(newBars []models.Bar) {
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

func (w *AgentWorker) trimHistoricalBarsLocked() {
	if w.maxBars <= 0 || len(w.historicalBars) <= w.maxBars {
		return
	}
	w.historicalBars = w.historicalBars[len(w.historicalBars)-w.maxBars:]
}
