package portfolio

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/oalpha/internal/agent/risk"
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
	regimeHMM    *risk.HMMRegimeDetector
	riskOverlay  *risk.RegimeRiskOverlay
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
		regimeHMM:    risk.NewHMMRegimeDetector(50),
		riskOverlay:  risk.NewRegimeRiskOverlay(risk.DefaultRiskOverlayPolicy()),
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

func (w *PortfolioAgentWorker) RefreshBarsFromAlpaca(ctx context.Context, start time.Time, end time.Time) (int64, error) {
	if w == nil {
		return 0, fmt.Errorf("portfolio worker is nil")
	}
	if w.repo == nil {
		return 0, fmt.Errorf("portfolio worker requires a bars repository")
	}
	if w.alpacaClient == nil || w.alpacaClient.APIKey() == "" || w.alpacaClient.APISecret() == "" {
		return 0, nil
	}

	barsBySymbol, err := w.alpacaClient.GetBarsMulti(ctx, w.symbols, w.timeframe, start, end, 10000, "iex", "raw")
	if err != nil {
		return 0, err
	}

	var inserted int64
	for _, bars := range barsBySymbol {
		n, err := w.repo.InsertBarsDataset(ctx, bars, w.timeframe, "iex", "raw", "alpaca")
		if err != nil {
			return inserted, err
		}
		inserted += n
	}
	return inserted, nil
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

func (w *PortfolioAgentWorker) RuntimeRegimeState(benchmarkSymbol string) map[string]interface{} {
	if w == nil {
		return nil
	}
	benchmarkSymbol = firstSymbol(normalizeSymbols([]string{benchmarkSymbol}))
	if benchmarkSymbol == "" {
		benchmarkSymbol = firstSymbol(w.symbols)
	}

	w.barsMu.RLock()
	bars := append([]models.Bar(nil), w.bars.Bars[benchmarkSymbol]...)
	w.barsMu.RUnlock()

	state := map[string]interface{}{
		"source":           "hmm_risk_overlay",
		"benchmark_symbol": benchmarkSymbol,
		"model_healthy":    false,
		"regime_label":     "Syncing",
		"overlay_role":     string(risk.RegimeRiskUnknown),
		"updated_at":       time.Now().UTC(),
	}
	if len(bars) == 0 || w.regimeHMM == nil {
		return state
	}

	regime, confidence, err := w.regimeHMM.Update(bars)
	probs := w.regimeHMM.GetProbabilities()
	state["probability_low"] = probs[0]
	state["probability_medium"] = probs[1]
	state["probability_high"] = probs[2]
	if err != nil {
		state["error"] = err.Error()
		return state
	}

	latest := bars[len(bars)-1]
	state["model_healthy"] = true
	state["regime_label"] = regime.String()
	state["confidence"] = confidence
	state["bar_time"] = latest.Time

	if w.riskOverlay != nil {
		volStart := len(bars) - w.regimeHMM.WindowSize()
		if volStart < 0 {
			volStart = 0
		}
		decision := w.riskOverlay.Apply(risk.RegimeOverlayInput{
			Timestamp:         latest.Time,
			BaseExposure:      1,
			PosteriorProbs:    []float64{probs[0], probs[1], probs[2]},
			StateRoles:        []risk.RegimeRiskRole{risk.RegimeRiskLowVol, risk.RegimeRiskNormal, risk.RegimeRiskHighVol},
			ModelHealthy:      true,
			RealizedAnnualVol: risk.RealizedVolatility(bars[volStart:]) * math.Sqrt(252),
		})
		state["overlay_role"] = string(decision.EffectiveRole)
		state["overlay_multiplier"] = decision.Multiplier
		state["overlay_confidence"] = decision.Confidence
		state["overlay_vetoed"] = decision.Vetoed
		state["overlay_reasons"] = decision.Reasons
	}

	return state
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

func firstSymbol(symbols []string) string {
	if len(symbols) == 0 {
		return ""
	}
	return symbols[0]
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
