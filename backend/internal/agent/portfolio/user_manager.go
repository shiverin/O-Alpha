package portfolio

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/oalpha/internal/alpaca"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
)

type PortfolioOrchestrator struct {
	mgr           *PortfolioAgentManager
	barsRepo      *db.BarsRepository
	agentRepo     *db.AgentRepository
	portfolioRepo *db.PortfolioRepository
	alpacaClient  *alpaca.Client
	cfg           StrategyCatalogConfig

	mu      sync.Mutex
	running map[int64]*userRun
}

type userRun struct {
	worker      *PortfolioAgentWorker
	runID       int64
	strategyKey string
	spec        StrategySpec
	symbols     []string
}

func NewPortfolioOrchestrator(mgr *PortfolioAgentManager, barsRepo *db.BarsRepository, agentRepo *db.AgentRepository, portfolioRepo *db.PortfolioRepository, alpacaClient *alpaca.Client, cfg StrategyCatalogConfig) *PortfolioOrchestrator {
	return &PortfolioOrchestrator{
		mgr:           mgr,
		barsRepo:      barsRepo,
		agentRepo:     agentRepo,
		portfolioRepo: portfolioRepo,
		alpacaClient:  alpacaClient,
		cfg:           cfg,
		running:       make(map[int64]*userRun),
	}
}

func userKey(userID int64) string {
	return fmt.Sprintf("user:%d", userID)
}

func (o *PortfolioOrchestrator) IsRunningForUser(userID int64) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	_, ok := o.running[userID]
	return ok
}

func (o *PortfolioOrchestrator) Catalog(symbols []string) []StrategySpec {
	return AvailableStrategySpecs(symbols, o.cfg)
}

func (o *PortfolioOrchestrator) SpecByKey(key string, symbols []string) (StrategySpec, error) {
	return StrategySpecByKey(key, symbols, o.cfg)
}

func (o *PortfolioOrchestrator) StartForUser(ctx context.Context, userID, agentRunID int64, strategyKey string, symbols []string, timeframe string, initialCash float64) (StrategySpec, error) {
	o.mu.Lock()
	if _, exists := o.running[userID]; exists {
		o.mu.Unlock()
		return StrategySpec{}, fmt.Errorf("a portfolio agent is already running for this user")
	}
	o.mu.Unlock()

	if timeframe == "" {
		timeframe = "1Day"
	}

	strategy, spec, err := NewStrategyFromCatalog(strategyKey, symbols, o.cfg)
	if err != nil {
		return StrategySpec{}, err
	}

	router := NewDBExecutionRouter(o.portfolioRepo, userID, agentRunID, initialCash)
	worker, err := o.mgr.StartPortfolioAgent(context.Background(), userKey(userID), strategy, symbols, timeframe, initialCash, router)
	if err != nil {
		return StrategySpec{}, err
	}

	end := time.Now().UTC()
	start := end.Add(-warmupLookbackFor(timeframe))
	opts := db.BarQueryOptions{AlignMode: backtest.AlignForwardFill, MaxStaleBars: 5}
	if err := worker.LoadInitialBars(start, end, opts); err != nil {
		log.Printf("[PortfolioOrchestrator] initial DB warmup failed for user %d, trying Alpaca refresh: %v", userID, err)
		if refreshErr := o.refreshBarsBeforeEvaluation(ctx, worker, symbols, timeframe, start, end); refreshErr != nil {
			_ = o.mgr.StopPortfolioAgent(userKey(userID))
			return StrategySpec{}, fmt.Errorf("market data refresh failed: %w", refreshErr)
		}
		if retryErr := worker.LoadInitialBars(start, end, opts); retryErr != nil {
			_ = o.mgr.StopPortfolioAgent(userKey(userID))
			return StrategySpec{}, fmt.Errorf("warmup failed (are bars ingested for the universe?): %w", retryErr)
		}
	}
	if !worker.HasBars() {
		_ = o.mgr.StopPortfolioAgent(userKey(userID))
		return StrategySpec{}, fmt.Errorf("no bars available for the selected universe/timeframe")
	}
	priceCtx, cancelPrices := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPrices()
	if prices, asOf, err := o.latestIntradayPrices(priceCtx, symbols); err != nil {
		log.Printf("[PortfolioOrchestrator] latest price refresh skipped during startup for user %d: %v", userID, err)
	} else {
		worker.ApplyLatestPrices(prices, asOf)
	}
	if err := o.agentRepo.UpdateAgentRunRuntimeState(ctx, agentRunID, worker.RuntimeRegimeState(spec.BenchmarkSymbol)); err != nil {
		_ = o.mgr.StopPortfolioAgent(userKey(userID))
		return StrategySpec{}, fmt.Errorf("runtime state initialization failed: %w", err)
	}

	o.mu.Lock()
	o.running[userID] = &userRun{worker: worker, runID: agentRunID, strategyKey: strategyKey, spec: spec, symbols: append([]string(nil), symbols...)}
	o.mu.Unlock()

	go o.loop(userID, agentRunID, worker, router, timeframe, spec.BenchmarkSymbol)

	_ = o.portfolioRepo.InsertSystemAlert(ctx, userID, "INFO", "Agent started", fmt.Sprintf("%s is now running in paper mode over %d symbols.", spec.DisplayName, len(symbols)), "portfolio_agent", map[string]interface{}{
		"run_id":            agentRunID,
		"strategy_key":      strategyKey,
		"deployment_status": string(spec.DeploymentStatus),
	})

	return spec, nil
}

func (o *PortfolioOrchestrator) StopForUser(userID int64) error {
	o.mu.Lock()
	run, ok := o.running[userID]
	if ok {
		delete(o.running, userID)
	}
	o.mu.Unlock()
	if !ok {
		return fmt.Errorf("no portfolio agent is running for this user")
	}

	run.worker.Stop()
	_ = o.mgr.StopPortfolioAgent(userKey(userID))
	return nil
}

func (o *PortfolioOrchestrator) loop(userID, agentRunID int64, worker *PortfolioAgentWorker, router *DBExecutionRouter, timeframe string, benchmarkSymbol string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("[PortfolioOrchestrator] loop panic for user %d: %v", userID, recovered)
			_ = o.portfolioRepo.InsertSystemAlert(context.Background(), userID, "CRITICAL", "Agent stopped unexpectedly", "The portfolio agent crashed and was stopped. You can relaunch it.", "portfolio_agent", map[string]interface{}{"run_id": agentRunID})
			_ = o.agentRepo.MarkAgentRunFailed(context.Background(), agentRunID, fmt.Sprintf("panic: %v", recovered))
		}
		o.mu.Lock()
		delete(o.running, userID)
		o.mu.Unlock()
		_ = o.mgr.StopPortfolioAgent(userKey(userID))
	}()

	ctx := worker.Context()
	opts := db.BarQueryOptions{AlignMode: backtest.AlignForwardFill, MaxStaleBars: 5}
	lookback := warmupLookbackFor(timeframe)
	interval := pollIntervalFor(timeframe)
	var lastRebalance time.Time

	lastRebalance = o.evaluateOnce(ctx, userID, agentRunID, worker, router, opts, lookback, benchmarkSymbol, lastRebalance)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lastRebalance = o.evaluateOnce(ctx, userID, agentRunID, worker, router, opts, lookback, benchmarkSymbol, lastRebalance)
		}
	}
}

func (o *PortfolioOrchestrator) evaluateOnce(ctx context.Context, userID, agentRunID int64, worker *PortfolioAgentWorker, router *DBExecutionRouter, opts db.BarQueryOptions, lookback time.Duration, benchmarkSymbol string, lastRebalance time.Time) time.Time {
	end := time.Now().UTC()
	start := end.Add(-lookback)
	if err := o.refreshBarsBeforeEvaluation(ctx, worker, worker.Symbols(), worker.timeframe, start, end); err != nil {
		log.Printf("[PortfolioOrchestrator] market data refresh failed for user %d: %v", userID, err)
		return lastRebalance
	}
	if err := worker.LoadInitialBars(start, end, opts); err != nil {
		log.Printf("[PortfolioOrchestrator] reload bars failed for user %d: %v", userID, err)
		return lastRebalance
	}
	if prices, asOf, err := o.latestIntradayPrices(ctx, worker.Symbols()); err != nil {
		log.Printf("[PortfolioOrchestrator] latest price refresh failed for user %d: %v", userID, err)
	} else {
		worker.ApplyLatestPrices(prices, asOf)
	}

	output, err := worker.EvaluateLatest()
	if err != nil {
		log.Printf("[PortfolioOrchestrator] evaluate failed for user %d: %v", userID, err)
		return lastRebalance
	}
	runtimeState := worker.RuntimeRegimeState(benchmarkSymbol)
	settings := o.loadRuntimeSettings(ctx, userID, end)
	rebalanceDue := settings.RebalanceDue(output.Time, lastRebalance)
	output = applyRuntimeSettingsToOutput(output, benchmarkSymbol, settings, rebalanceDue, lastRebalance)
	runtimeState[runtimeSettingsMetadataKey] = output.EngineMetadata[runtimeSettingsMetadataKey]
	runtimeState[runtimeCadenceMetadataKey] = rebalanceDue
	if reason, ok := output.EngineMetadata[runtimeSuppressedMetadataKey]; ok {
		runtimeState[runtimeSuppressedMetadataKey] = reason
	}

	opCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := router.ExecutePortfolioTargetsWithSettings(opCtx, output, worker.LatestPrices(), settings); err != nil {
		log.Printf("[PortfolioOrchestrator] execute failed for user %d: %v", userID, err)
	}
	if shouldRecordRebalance(output, rebalanceDue) {
		lastRebalance = output.Time
		if lastRebalance.IsZero() {
			lastRebalance = end
		}
	}

	if err := o.agentRepo.UpdateAgentRunRuntimeState(opCtx, agentRunID, runtimeState); err != nil {
		log.Printf("[PortfolioOrchestrator] runtime state update failed for run %d: %v", agentRunID, err)
	}

	if err := o.agentRepo.UpdateAgentRunHeartbeat(opCtx, agentRunID); err != nil {
		log.Printf("[PortfolioOrchestrator] heartbeat failed for run %d: %v", agentRunID, err)
	}
	return lastRebalance
}

func shouldRecordRebalance(output backtest.PortfolioOutput, rebalanceDue bool) bool {
	return rebalanceDue && len(output.Targets) > 0
}

func (o *PortfolioOrchestrator) loadRuntimeSettings(ctx context.Context, userID int64, now time.Time) RuntimeSettings {
	if o == nil || o.agentRepo == nil {
		return DefaultRuntimeSettings(now)
	}
	settings, err := o.agentRepo.GetAgentSettings(ctx, userID)
	if err != nil {
		log.Printf("[PortfolioOrchestrator] settings load failed for user %d: %v", userID, err)
		return DefaultRuntimeSettings(now)
	}
	return RuntimeSettingsFromAgentSettings(settings, now)
}

func (o *PortfolioOrchestrator) refreshBarsBeforeEvaluation(ctx context.Context, worker *PortfolioAgentWorker, symbols []string, timeframe string, lookbackStart time.Time, end time.Time) error {
	if o == nil || o.barsRepo == nil || worker == nil || o.alpacaClient == nil || o.alpacaClient.APIKey() == "" || o.alpacaClient.APISecret() == "" {
		return nil
	}

	start := lookbackStart
	latestBySymbol, err := o.barsRepo.GetLatestBarTimes(ctx, symbols, timeframe)
	if err != nil {
		return err
	}
	if len(latestBySymbol) == len(symbols) {
		var oldestLatest time.Time
		for _, symbol := range symbols {
			latest := latestBySymbol[symbol]
			if oldestLatest.IsZero() || latest.Before(oldestLatest) {
				oldestLatest = latest
			}
		}
		if !oldestLatest.IsZero() {
			deltaStart := oldestLatest.Add(-refreshOverlapFor(timeframe))
			if deltaStart.After(start) {
				start = deltaStart
			}
		}
	}

	if !start.Before(end) {
		return nil
	}
	refreshCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	_, err = worker.RefreshBarsFromAlpaca(refreshCtx, start, end)
	return err
}

func (o *PortfolioOrchestrator) latestIntradayPrices(ctx context.Context, symbols []string) (map[string]float64, time.Time, error) {
	if o == nil || o.alpacaClient == nil || o.alpacaClient.APIKey() == "" || o.alpacaClient.APISecret() == "" {
		return nil, time.Time{}, nil
	}
	refreshCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return o.alpacaClient.GetLatestPrices(refreshCtx, symbols)
}
