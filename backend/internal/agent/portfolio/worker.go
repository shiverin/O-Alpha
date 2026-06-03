package portfolio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/oalpha/internal/alpaca"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/pkg/models"
)

type ExecutionRouter interface {
	ExecutePortfolioTargets(ctx context.Context, output backtest.PortfolioOutput, prices map[string]float64) error
}

type PortfolioPaperAccount struct {
	Cash      float64
	Positions map[string]PortfolioPaperPosition
}

type PortfolioPaperPosition struct {
	LongQty  float64
	ShortQty float64
}

func NewPortfolioPaperAccount(initialCash float64) *PortfolioPaperAccount {
	if initialCash <= 0 {
		initialCash = 100_000
	}
	return &PortfolioPaperAccount{
		Cash:      initialCash,
		Positions: make(map[string]PortfolioPaperPosition),
	}
}

func (a *PortfolioPaperAccount) Equity(prices map[string]float64) float64 {
	if a == nil {
		return 0
	}
	equity := a.Cash
	for symbol, position := range a.Positions {
		price := prices[symbol]
		equity += position.LongQty * price
		equity -= position.ShortQty * price
	}
	return equity
}

type PortfolioAgentWorker struct {
	ctx          context.Context
	cancelFunc   context.CancelFunc
	strategy     backtest.PortfolioStrategy
	symbols      []string
	timeframe    string
	account      *PortfolioPaperAccount
	bars         backtest.AlignedBars
	maxBars      int
	repo         *db.BarsRepository
	alpacaClient *alpaca.Client
	execution    ExecutionRouter
	barsMu       sync.RWMutex
}

func NewPortfolioAgentWorker(
	ctx context.Context,
	strategy backtest.PortfolioStrategy,
	symbols []string,
	timeframe string,
	initialCash float64,
	repo *db.BarsRepository,
	alpacaClient *alpaca.Client,
	execution ExecutionRouter,
) *PortfolioAgentWorker {
	workerCtx, cancel := context.WithCancel(ctx)
	return &PortfolioAgentWorker{
		ctx:          workerCtx,
		cancelFunc:   cancel,
		strategy:     strategy,
		symbols:      append([]string(nil), symbols...),
		timeframe:    timeframe,
		account:      NewPortfolioPaperAccount(initialCash),
		maxBars:      10000,
		repo:         repo,
		alpacaClient: alpacaClient,
		execution:    execution,
	}
}

func (w *PortfolioAgentWorker) Stop() {
	if w != nil && w.cancelFunc != nil {
		w.cancelFunc()
	}
}

func (w *PortfolioAgentWorker) LoadInitialBars(start time.Time, end time.Time, opts db.BarQueryOptions) error {
	if w == nil {
		return fmt.Errorf("portfolio worker is nil")
	}
	if w.repo == nil {
		return fmt.Errorf("portfolio worker requires a bars repository")
	}
	panel, err := w.repo.GetBarsMulti(w.ctx, w.symbols, w.timeframe, start, end, opts)
	if err != nil {
		return err
	}
	w.barsMu.Lock()
	defer w.barsMu.Unlock()
	w.bars = trimPanel(panel, w.maxBars)
	return nil
}

func (w *PortfolioAgentWorker) EvaluateLatest() (backtest.PortfolioOutput, error) {
	if w == nil || w.strategy == nil {
		return backtest.PortfolioOutput{}, fmt.Errorf("portfolio worker requires a strategy")
	}
	w.barsMu.RLock()
	panel := w.bars
	w.barsMu.RUnlock()
	if len(panel.Times) == 0 {
		return backtest.PortfolioOutput{}, fmt.Errorf("portfolio worker has no bars")
	}
	return w.strategy.EvaluatePortfolioLatest(w.ctx, panel)
}

func (w *PortfolioAgentWorker) MergeBars(barsBySymbol map[string][]models.Bar, alignMode backtest.AlignMode) error {
	if w == nil {
		return fmt.Errorf("portfolio worker is nil")
	}
	panel, err := backtest.AlignBars(barsBySymbol, backtest.AlignmentConfig{
		Mode:      alignMode,
		Timeframe: w.timeframe,
	})
	if err != nil {
		return err
	}
	w.barsMu.Lock()
	defer w.barsMu.Unlock()
	w.bars = trimPanel(panel, w.maxBars)
	return nil
}

func trimPanel(panel backtest.AlignedBars, maxBars int) backtest.AlignedBars {
	if maxBars <= 0 || len(panel.Times) <= maxBars {
		return panel
	}
	start := len(panel.Times) - maxBars
	panel.Times = append([]time.Time(nil), panel.Times[start:]...)
	for symbol, bars := range panel.Bars {
		panel.Bars[symbol] = append([]models.Bar(nil), bars[start:]...)
	}
	return panel
}
