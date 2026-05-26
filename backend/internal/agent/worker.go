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

// AgentWorker runs the live agent execution loop
type AgentWorker struct {
	ctx            context.Context
	cancelFunc     context.CancelFunc
	alpacaClient   *alpaca.Client
	repo           *db.BarsRepository 
	strategy       backtest.Strategy
	symbol         string
	timeframe      string
	paperTrade     bool
	initialCash    float64
	account        *PaperAccount
	ticker         *time.Ticker
	doneCh         chan struct{}
	errCh          chan error
	wsConnector    *marketdata.WebSocketConnector
	useWebSocket   bool
	barsMu         sync.RWMutex
	historicalBars []models.Bar
	maxBars        int
}

// NewAgentWorker creates a new agent worker
func NewAgentWorker(
	ctx context.Context,
	alpacaClient *alpaca.Client,
	repo *db.BarsRepository,
	strategy backtest.Strategy,
	symbol string,
	timeframe string,
	paperTrade bool,
	initialCash float64,
	useWebSocket bool,
) *AgentWorker {
	wCtx, cancel := context.WithCancel(ctx)
	worker := &AgentWorker{
		ctx:          wCtx,
		cancelFunc:   cancel,
		alpacaClient: alpacaClient,
		repo:         repo,
		strategy:     strategy,
		symbol:       symbol,
		timeframe:    timeframe,
		paperTrade:   paperTrade,
		initialCash:  initialCash,
		account:      NewPaperAccount(initialCash),
		doneCh:       make(chan struct{}),
		errCh:        make(chan error, 1),
		useWebSocket: useWebSocket,
		maxBars:      10000,
	}

	if useWebSocket {
		worker.wsConnector = marketdata.NewWebSocketConnector(alpacaClient, []string{symbol}, timeframe)
	}

	return worker
}

// Start begins the agent worker loop
func (w *AgentWorker) Start() error {
	log.Printf("Starting agent worker for %s (%s)", w.symbol, w.timeframe)

	// Fetch initial data to warm up indicators
	end := time.Now().UTC()
	start := end.Add(-720 * time.Hour) // 30 days of data for warmup

	bars, err := w.alpacaClient.GetBars(w.ctx, w.symbol, w.timeframe, start, end, 10000)
	if err != nil {
		return fmt.Errorf("failed to fetch initial data: %w", err)
	}

	// Generate initial signals
	if _, err := w.strategy.GenerateSignal(w.ctx, bars); err != nil {
		return fmt.Errorf("failed to generate initial signals: %w", err)
	}
	w.setHistoricalBars(bars)

	if w.useWebSocket && w.wsConnector != nil {
		// Use WebSocket for real-time data
		if err := w.wsConnector.Start(w.ctx); err != nil {
			return fmt.Errorf("failed to start WebSocket: %w", err)
		}
		go w.handleWebSocketData()
	} else {
		// Use ticker-based polling (original behavior)
		// Set ticker based on timeframe
		interval := timeframeToDuration(w.timeframe)

		w.ticker = time.NewTicker(interval)

		// Run the worker loop
		go w.runLoop()
	}

	return nil
}

// Stop stops the agent worker
func (w *AgentWorker) Stop() {
	log.Printf("Stopping agent worker for %s", w.symbol)
	w.cancelFunc()
	if w.ticker != nil {
		w.ticker.Stop()
	}
	close(w.doneCh)
}

// runLoop is the main worker loop that fetches data, generates signals, and executes trades
func (w *AgentWorker) runLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Agent worker panicked: %v", r)
			// Attempt to restart or signal error
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
			log.Printf("Agent worker context cancelled for %s", w.symbol)
			return
		case <-w.ticker.C:
			if err := w.processTick(); err != nil {
				log.Printf("Error processing tick: %v", err)
				// Continue processing - don't stop the worker for transient errors
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

// handleWebSocketData processes incoming real-time bar data from WebSocket
func (w *AgentWorker) handleWebSocketData() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WebSocket handler panicked: %v", r)
			select {
			case w.errCh <- fmt.Errorf("websocket handler panicked: %v", r):
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
			// Process real-time bar data
			if err := w.processNewBar(bar); err != nil {
				log.Printf("Error processing new bar: %v", err)
				// Continue processing - don't stop for transient errors
				select {
				case w.errCh <- err:
				default:
				}
			}
		case err := <-w.wsConnector.Errors():
			log.Printf("WebSocket error: %v", err)
			select {
			case w.errCh <- err:
			default:
			}
			// Continue processing - don't stop worker for WS errors (will reconnect)
		case <-w.doneCh:
			return
		}
	}
}

// processNewBar handles incoming real-time bar data
func (w *AgentWorker) processNewBar(bar models.Bar) error {
	w.appendOrUpdateBar(bar)
	historicalBars := w.getHistoricalBarsSnapshot()
	if len(historicalBars) == 0 {
		return fmt.Errorf("no bars available for signal generation")
	}

	// Generate signals with updated data
	signals, err := w.strategy.GenerateSignal(w.ctx, historicalBars)
	if err != nil {
		return fmt.Errorf("failed to generate signals: %w", err)
	}

	if len(signals) == 0 {
		return fmt.Errorf("no signals generated")
	}

	// Get the latest signal (for the most recent bar)
	latestSignal := signals[len(signals)-1]
	latestBar := historicalBars[len(historicalBars)-1]

	// Execute trade based on signal (only act on new signals to avoid overtrading)
	if latestSignal != models.SignalHold {
		// In a real implementation, we would check if we've already acted on this signal
		// For simplicity, we'll act on every signal (in production, you'd track signal IDs)
		err := w.executeTrade(latestSignal, latestBar.Close)
		if err != nil {
			return fmt.Errorf("failed to execute trade: %w", err)
		}
	}

	// Log status
	if w.paperTrade {
		cash, positions := w.account.Snapshot()
		log.Printf("Paper trade - Symbol: %s, Signal: %v, Price: %.2f, Cash: %.2f, Positions: %v",
			w.symbol, latestSignal, latestBar.Close, cash, positions)
	} else {
		log.Printf("Live trade - Symbol: %s, Signal: %v, Price: %.2f",
			w.symbol, latestSignal, latestBar.Close)
	}

	return nil
}

// processTick fetches new data, generates signals, and executes trades
func (w *AgentWorker) processTick() error {
	// Fetch only a small recent window and merge it into in-memory history.
	end := time.Now().UTC()
	start := end.Add(-2 * timeframeToDuration(w.timeframe))

	bars, err := w.alpacaClient.GetBars(w.ctx, w.symbol, w.timeframe, start, end, 10000)
	if err != nil {
		return fmt.Errorf("failed to fetch bars: %w", err)
	}

	if len(bars) == 0 {
		return nil
	}

	w.mergeBars(bars)
	bars = w.getHistoricalBarsSnapshot()
	if len(bars) == 0 {
		return fmt.Errorf("no bars available for signal generation")
	}

	// Generate signals
	signals, err := w.strategy.GenerateSignal(w.ctx, bars)
	if err != nil {
		return fmt.Errorf("failed to generate signals: %w", err)
	}

	if len(signals) == 0 {
		return fmt.Errorf("no signals generated")
	}

	// Get the latest signal (for the most recent bar)
	latestSignal := signals[len(signals)-1]
	latestBar := bars[len(bars)-1]

	// Execute trade based on signal (only act on new signals to avoid overtrading)
	if latestSignal != models.SignalHold {
		// In a real implementation, we would check if we've already acted on this signal
		// For simplicity, we'll act on every signal (in production, you'd track signal IDs)
		err := w.executeTrade(latestSignal, latestBar.Close)
		if err != nil {
			return fmt.Errorf("failed to execute trade: %w", err)
		}
	}

	// Log status
	if w.paperTrade {
		cash, positions := w.account.Snapshot()
		log.Printf("Paper trade - Symbol: %s, Signal: %v, Price: %.2f, Cash: %.2f, Positions: %v",
			w.symbol, latestSignal, latestBar.Close, cash, positions)
	} else {
		log.Printf("Live trade - Symbol: %s, Signal: %v, Price: %.2f",
			w.symbol, latestSignal, latestBar.Close)
	}

	return nil
}

// executeTrade executes a trade based on the signal
func (w *AgentWorker) executeTrade(signal models.Signal, price float64) error {
	if w.paperTrade {
		return w.executePaperTrade(signal, price)
	}
	return w.executeLiveTrade(signal, price)
}

// executePaperTrade executes a trade in the paper account
func (w *AgentWorker) executePaperTrade(signal models.Signal, price float64) error {
	var amount float64
	var err error

	switch signal {
	case models.SignalBuy:
		// Use 10% of available cash for each trade
		cashToUse := w.account.Cash * 0.1
		if cashToUse < price {
			return fmt.Errorf("insufficient cash for minimum trade")
		}
		amount = cashToUse / price
		_, _, err = w.account.Buy(w.ctx, w.symbol, price, amount)
		if err != nil {
			return fmt.Errorf("paper buy failed: %w", err)
		}
	case models.SignalSell:
		// Sell 50% of current position
		currentPos := w.account.GetPosition(w.symbol)
		if currentPos <= 0 {
			return fmt.Errorf("no position to sell")
		}
		amount = currentPos * 0.5
		if amount < 1.0 { // Minimum 1 share (assuming fractional shares allowed)
			amount = 1.0
		}
		_, _, err = w.account.Sell(w.ctx, w.symbol, price, amount)
		if err != nil {
			return fmt.Errorf("paper sell failed: %w", err)
		}
	default:
		return fmt.Errorf("unsupported signal action: %v", signal)
	}

	return nil
}

// executeLiveTrade executes a trade via the Alpaca API
func (w *AgentWorker) executeLiveTrade(signal models.Signal, price float64) error {
	// In a real implementation, this would use the Alpaca Trading API
	// For now, we'll just log the intended action
	log.Printf("Would execute live trade: %v %f shares of %s at $%.2f",
		signal, 0.0, w.symbol, price) // amount calculation omitted for brevity
	return nil
}

// Err returns the error channel for the worker
func (w *AgentWorker) Err() <-chan error {
	return w.errCh
}

// Done returns a channel that closes when the worker stops
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
