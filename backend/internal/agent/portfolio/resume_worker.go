package portfolio

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
)

type PortfolioRunOnceResult struct {
	RunID           int64
	UserID          int64
	StrategyKey     string
	Symbols         []string
	LastRebalanceAt time.Time
}

func (o *PortfolioOrchestrator) RunOnceForAgentRun(ctx context.Context, run db.AgentRunSummary) (PortfolioRunOnceResult, error) {
	result := PortfolioRunOnceResult{
		RunID:  run.ID,
		UserID: run.UserID,
	}
	if o == nil || o.mgr == nil || o.barsRepo == nil || o.agentRepo == nil || o.portfolioRepo == nil {
		return result, fmt.Errorf("portfolio orchestrator is not fully configured")
	}
	if run.ID <= 0 {
		return result, fmt.Errorf("agent run id is required")
	}
	if run.UserID <= 0 {
		return result, fmt.Errorf("agent run %d is missing user_id", run.ID)
	}

	symbols := symbolsFromAgentRun(run)
	if len(symbols) == 0 {
		return result, fmt.Errorf("run %d symbols: parameters are empty", run.ID)
	}
	result.Symbols = symbols

	strategyKey := strings.TrimSpace(run.StrategyKey)
	if strategyKey == "" {
		strategyKey, _ = stringFromMap(run.Parameters, "strategy_key")
	}
	if strategyKey == "" {
		return result, fmt.Errorf("run %d is missing strategy_key", run.ID)
	}
	result.StrategyKey = strategyKey

	timeframe := strings.TrimSpace(run.Timeframe)
	if timeframe == "" {
		timeframe = "1Day"
	}
	strategy, spec, err := NewStrategyFromCatalog(strategyKey, symbols, o.cfg)
	if err != nil {
		return result, fmt.Errorf("run %d strategy: %w", run.ID, err)
	}

	router := o.executionRouterFor(run.UserID, run.ID, run.InitialCash)
	workerKey := fmt.Sprintf("github-worker:%d", run.ID)
	worker, err := o.mgr.StartPortfolioAgent(ctx, workerKey, strategy, symbols, timeframe, run.InitialCash, router)
	if err != nil {
		return result, fmt.Errorf("start one-shot portfolio worker for run %d: %w", run.ID, err)
	}
	defer func() {
		worker.Stop()
		_ = o.mgr.StopPortfolioAgent(workerKey)
	}()

	end := time.Now().UTC()
	start := end.Add(-warmupLookbackFor(timeframe))
	opts := db.BarQueryOptions{AlignMode: backtest.AlignForwardFill, MaxStaleBars: 5}
	if err := worker.LoadInitialBars(start, end, opts); err != nil {
		if refreshErr := o.refreshBarsBeforeEvaluation(ctx, worker, symbols, timeframe, start, end); refreshErr != nil {
			return result, fmt.Errorf("run %d market data refresh: %w", run.ID, refreshErr)
		}
		if retryErr := worker.LoadInitialBars(start, end, opts); retryErr != nil {
			return result, fmt.Errorf("run %d warmup: %w", run.ID, retryErr)
		}
	}
	if !worker.HasBars() {
		return result, fmt.Errorf("run %d has no bars for selected universe/timeframe", run.ID)
	}
	if prices, asOf, err := o.latestIntradayPrices(ctx, symbols); err == nil {
		worker.ApplyLatestPrices(prices, asOf)
	}

	lastRebalance := o.resumeLastRebalance(ctx, run)
	lastRebalance = o.evaluateOnce(ctx, run.UserID, run.ID, worker, router, opts, warmupLookbackFor(timeframe), spec.BenchmarkSymbol, lastRebalance)
	result.LastRebalanceAt = lastRebalance
	return result, nil
}
