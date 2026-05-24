package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/oalpha/internal/alpaca"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/pkg/models"
)

// AgentWorker runs the live agent execution loop
type AgentWorker struct {
	ctx          context.Context
	cancelFunc   context.CancelFunc
	alpacaClient *alpaca.Client
	repo         *db.Repository
	strategy     Strategy
	symbol       string
	timeframe    string
	paperTrade   bool
	initialCash  float64
	account      *PaperAccount
	ticker       *time.Ticker
	doneCh       chan struct{}
	errCh        chan error
}

// NewAgentWorker creates a new agent worker
func NewAgentWorker(
	ctx context.Context,
	alpacaClient *alpaca.Client,
	repo *db.Repository,
	strategy Strategy,
	symbol string,
	timeframe string,
	paperTrade bool,
	initialCash float64,
) *AgentWorker {
	wCtx, cancel := context.WithCancel(ctx)
	return &AgentWorker{
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
	}
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

	// Set ticker based on timeframe
	var interval time.Duration
	switch w.timeframe {
	case "1Min":
		interval = time.Minute
	case "5Min":
		interval = 5 * time.Minute
	case "15Min":
		interval = 15 * time.Minute
	case "1Hour":
		interval = time.Hour
	case "1Day":
		interval = 24 * time.Hour
	default:
		interval = time.Hour // default to 1 hour
	}

	w.ticker = time.NewTicker(interval)

	// Run the worker loop
	go w.runLoop()

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

// processTick fetches new data, generates signals, and executes trades
func (w *AgentWorker) processTick() error {
	// Fetch recent bars (enough for indicator warmup)
	end := time.Now().UTC()
	start := end.Add(-720 * time.Hour) // 30 days

	bars, err := w.alpacaClient.GetBars(w.ctx, w.symbol, w.timeframe, start, end, 10000)
	if err != nil {
		return fmt.Errorf("failed to fetch bars: %w", err)
	}

	if len(bars) == 0 {
		return fmt.Errorf("no bars returned")
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
	if latestSignal.Action != models.Hold {
		// In a real implementation, we would check if we've already acted on this signal
		// For simplicity, we'll act on every signal (in production, you'd track signal IDs)
		err := w.executeTrade(latestSignal, latestBar.Close)
		if err != nil {
			return fmt.Errorf("failed to execute trade: %w", err)
		}
	}

	// Log status
	if w.paperTrade {
		log.Printf("Paper trade - Symbol: %s, Signal: %s, Price: %.2f, Cash: %.2f, Positions: %v",
			w.symbol, latestSignal.Action, latestBar.Close, w.account.Cash, w.account.Positions)
	} else {
		log.Printf("Live trade - Symbol: %s, Signal: %s, Price: %.2f",
			w.symbol, latestSignal.Action, latestBar.Close)
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

	switch signal.Action {
	case models.Buy:
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
	case models.Sell:
		// Sell 50% of current position
		currentPos := w.account.Positions[w.symbol]
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
		return fmt.Errorf("unsupported signal action: %v", signal.Action)
	}

	return nil
}

// executeLiveTrade executes a trade via the Alpaca API
func (w *AgentWorker) executeLiveTrade(signal models.Signal, price float64) error {
	// In a real implementation, this would use the Alpaca Trading API
	// For now, we'll just log the intended action
	log.Printf("Would execute live trade: %s %f shares of %s at $%.2f",
		signal.Action, 0.0, w.symbol, price) // amount calculation omitted for brevity
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