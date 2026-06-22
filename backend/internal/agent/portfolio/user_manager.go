package portfolio

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	paperBroker   AlpacaPaperBroker
	executionMode string
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

type ResumePortfolioRunsResult struct {
	Resumed int
	Skipped int
	Failed  int
}

type startPortfolioRunOptions struct {
	initialLastRebalance time.Time
	markResumed          bool
}

type portfolioExecutionRouter interface {
	ExecutePortfolioTargets(ctx context.Context, output backtest.PortfolioOutput, prices map[string]float64) error
	ExecutePortfolioTargetsWithSettings(ctx context.Context, output backtest.PortfolioOutput, prices map[string]float64, settings RuntimeSettings) error
}

const (
	PortfolioExecutionInternal    = "internal"
	PortfolioExecutionAlpacaPaper = "alpaca_paper"
)

func NewPortfolioOrchestrator(mgr *PortfolioAgentManager, barsRepo *db.BarsRepository, agentRepo *db.AgentRepository, portfolioRepo *db.PortfolioRepository, alpacaClient *alpaca.Client, paperBroker AlpacaPaperBroker, executionMode string, cfg StrategyCatalogConfig) *PortfolioOrchestrator {
	executionMode = normalizePortfolioExecutionMode(executionMode)
	return &PortfolioOrchestrator{
		mgr:           mgr,
		barsRepo:      barsRepo,
		agentRepo:     agentRepo,
		portfolioRepo: portfolioRepo,
		alpacaClient:  alpacaClient,
		paperBroker:   paperBroker,
		executionMode: executionMode,
		cfg:           cfg,
		running:       make(map[int64]*userRun),
	}
}

func normalizePortfolioExecutionMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case PortfolioExecutionAlpacaPaper:
		return PortfolioExecutionAlpacaPaper
	default:
		return PortfolioExecutionInternal
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

func (o *PortfolioOrchestrator) ExecutionMode() string {
	if o == nil {
		return PortfolioExecutionInternal
	}
	return o.executionMode
}

func (o *PortfolioOrchestrator) SpecByKey(key string, symbols []string) (StrategySpec, error) {
	return StrategySpecByKey(key, symbols, o.cfg)
}

func (o *PortfolioOrchestrator) StartForUser(ctx context.Context, userID, agentRunID int64, strategyKey string, symbols []string, timeframe string, initialCash float64) (StrategySpec, error) {
	spec, err := o.startForUser(ctx, userID, agentRunID, strategyKey, symbols, timeframe, initialCash, startPortfolioRunOptions{})
	if err != nil {
		return StrategySpec{}, err
	}

	_ = o.portfolioRepo.InsertSystemAlert(ctx, userID, "INFO", "Agent started", fmt.Sprintf("%s is now running in %s mode over %d symbols.", spec.DisplayName, o.executionMode, len(symbols)), "portfolio_agent", map[string]interface{}{
		"run_id":            agentRunID,
		"strategy_key":      strategyKey,
		"deployment_status": string(spec.DeploymentStatus),
		"execution_mode":    o.executionMode,
	})

	return spec, nil
}

func (o *PortfolioOrchestrator) ResumeActiveRuns(ctx context.Context) (ResumePortfolioRunsResult, error) {
	var result ResumePortfolioRunsResult
	if o == nil || o.agentRepo == nil {
		return result, nil
	}

	runs, err := o.agentRepo.ListResumablePortfolioRuns(ctx)
	if err != nil {
		return result, err
	}

	seenUsers := make(map[int64]struct{}, len(runs))
	var failures []string
	for _, run := range runs {
		if _, seen := seenUsers[run.UserID]; seen {
			result.Skipped++
			continue
		}
		seenUsers[run.UserID] = struct{}{}

		if o.IsRunningForUser(run.UserID) {
			result.Skipped++
			continue
		}

		if err := o.ResumeRun(ctx, run); err != nil {
			result.Failed++
			failures = append(failures, fmt.Sprintf("run %d/user %d: %v", run.ID, run.UserID, err))
			continue
		}
		result.Resumed++
	}

	if len(failures) > 0 {
		return result, fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return result, nil
}

func (o *PortfolioOrchestrator) ResumeRun(ctx context.Context, run db.AgentRunSummary) error {
	if run.UserID <= 0 {
		return fmt.Errorf("resumable portfolio run %d is missing user_id", run.ID)
	}
	strategyKey := strings.ToLower(strings.TrimSpace(run.StrategyKey))
	if strategyKey == "" {
		strategyKey, _ = stringFromMap(run.Parameters, "strategy_key")
	}
	if strategyKey == "" {
		err := fmt.Errorf("resumable portfolio run %d is missing strategy_key", run.ID)
		o.markResumeFailed(ctx, run, err)
		return err
	}

	symbols := symbolsFromAgentRun(run)
	if len(symbols) == 0 {
		err := fmt.Errorf("resumable portfolio run %d is missing symbols", run.ID)
		o.markResumeFailed(ctx, run, err)
		return err
	}

	lastRebalance := o.resumeLastRebalance(ctx, run)
	spec, err := o.startForUser(ctx, run.UserID, run.ID, strategyKey, symbols, run.Timeframe, run.InitialCash, startPortfolioRunOptions{
		initialLastRebalance: lastRebalance,
		markResumed:          true,
	})
	if err != nil {
		o.markResumeFailed(ctx, run, err)
		return err
	}

	if o.portfolioRepo != nil {
		_ = o.portfolioRepo.InsertSystemAlert(ctx, run.UserID, "INFO", "Agent resumed", fmt.Sprintf("%s resumed after backend restart in %s mode.", spec.DisplayName, o.executionMode), "portfolio_agent", map[string]interface{}{
			"run_id":              run.ID,
			"strategy_key":        strategyKey,
			"execution_mode":      o.executionMode,
			"last_rebalance_at":   lastRebalance,
			"last_rebalance_seed": !lastRebalance.IsZero(),
		})
	}
	return nil
}

func (o *PortfolioOrchestrator) markResumeFailed(ctx context.Context, run db.AgentRunSummary, cause error) {
	reason := fmt.Sprintf("resume_failed: %v", cause)
	if o.agentRepo != nil {
		_ = o.agentRepo.MarkAgentRunFailed(ctx, run.ID, reason)
	}
	if o.portfolioRepo != nil {
		_ = o.portfolioRepo.InsertSystemAlert(ctx, run.UserID, "CRITICAL", "Agent resume failed", "The portfolio agent could not be resumed after backend restart. You can relaunch it.", "portfolio_agent", map[string]interface{}{
			"run_id": run.ID,
			"reason": reason,
		})
	}
}

func (o *PortfolioOrchestrator) startForUser(ctx context.Context, userID, agentRunID int64, strategyKey string, symbols []string, timeframe string, initialCash float64, opts startPortfolioRunOptions) (StrategySpec, error) {
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

	router := o.executionRouterFor(userID, agentRunID, initialCash)
	worker, err := o.mgr.StartPortfolioAgent(context.Background(), userKey(userID), strategy, symbols, timeframe, initialCash, router)
	if err != nil {
		return StrategySpec{}, err
	}

	end := time.Now().UTC()
	start := end.Add(-warmupLookbackFor(timeframe))
	barOpts := db.BarQueryOptions{AlignMode: backtest.AlignForwardFill, MaxStaleBars: 5}
	if err := worker.LoadInitialBars(start, end, barOpts); err != nil {
		log.Printf("[PortfolioOrchestrator] initial DB warmup failed for user %d, trying Alpaca refresh: %v", userID, err)
		if refreshErr := o.refreshBarsBeforeEvaluation(ctx, worker, symbols, timeframe, start, end); refreshErr != nil {
			_ = o.mgr.StopPortfolioAgent(userKey(userID))
			return StrategySpec{}, fmt.Errorf("market data refresh failed: %w", refreshErr)
		}
		if retryErr := worker.LoadInitialBars(start, end, barOpts); retryErr != nil {
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
	runtimeState := worker.RuntimeRegimeState(spec.BenchmarkSymbol)
	annotateLastRebalance(runtimeState, opts.initialLastRebalance, o.loadRuntimeSettings(ctx, userID, time.Now().UTC()))
	if err := o.agentRepo.UpdateAgentRunRuntimeState(ctx, agentRunID, runtimeState); err != nil {
		_ = o.mgr.StopPortfolioAgent(userKey(userID))
		return StrategySpec{}, fmt.Errorf("runtime state initialization failed: %w", err)
	}
	if opts.markResumed {
		if err := o.agentRepo.MarkAgentRunResumed(ctx, agentRunID); err != nil {
			_ = o.mgr.StopPortfolioAgent(userKey(userID))
			return StrategySpec{}, err
		}
	}

	o.mu.Lock()
	o.running[userID] = &userRun{worker: worker, runID: agentRunID, strategyKey: strategyKey, spec: spec, symbols: append([]string(nil), symbols...)}
	o.mu.Unlock()

	go o.loop(userID, agentRunID, worker, router, timeframe, spec.BenchmarkSymbol, opts.initialLastRebalance)

	return spec, nil
}

func (o *PortfolioOrchestrator) executionRouterFor(userID, agentRunID int64, initialCash float64) portfolioExecutionRouter {
	if o.executionMode == PortfolioExecutionAlpacaPaper && o.paperBroker != nil {
		return NewAlpacaPaperExecutionRouter(o.portfolioRepo, o.paperBroker, userID, agentRunID, initialCash)
	}
	return NewDBExecutionRouter(o.portfolioRepo, userID, agentRunID, initialCash)
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

func (o *PortfolioOrchestrator) loop(userID, agentRunID int64, worker *PortfolioAgentWorker, router portfolioExecutionRouter, timeframe string, benchmarkSymbol string, initialLastRebalance time.Time) {
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
	lastRebalance := initialLastRebalance.UTC()

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

func (o *PortfolioOrchestrator) evaluateOnce(ctx context.Context, userID, agentRunID int64, worker *PortfolioAgentWorker, router portfolioExecutionRouter, opts db.BarQueryOptions, lookback time.Duration, benchmarkSymbol string, lastRebalance time.Time) time.Time {
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
	annotateLastRebalance(runtimeState, lastRebalance, settings)
	runtimeState[runtimeCadenceMetadataKey] = rebalanceDue
	if reason, ok := output.EngineMetadata[runtimeSuppressedMetadataKey]; ok {
		runtimeState[runtimeSuppressedMetadataKey] = reason
	} else {
		delete(runtimeState, runtimeSuppressedMetadataKey)
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

func (o *PortfolioOrchestrator) resumeLastRebalance(ctx context.Context, run db.AgentRunSummary) time.Time {
	if last, ok := lastRebalanceFromRuntimeState(run.RuntimeState); ok {
		return last.UTC()
	}
	if o != nil && o.portfolioRepo != nil {
		last, err := o.portfolioRepo.LatestPortfolioRebalanceTime(ctx, run.UserID, run.ID)
		if err != nil {
			log.Printf("[PortfolioOrchestrator] latest rebalance fallback failed for run %d: %v", run.ID, err)
		} else if !last.IsZero() {
			return last.UTC()
		}
	}
	if run.LastHeartbeatAt != nil && !run.LastHeartbeatAt.IsZero() {
		return run.LastHeartbeatAt.UTC()
	}
	return time.Now().UTC()
}

func annotateLastRebalance(runtimeState map[string]interface{}, lastRebalance time.Time, settings RuntimeSettings) {
	if runtimeState == nil {
		return
	}
	lastRebalance = lastRebalance.UTC()
	if !lastRebalance.IsZero() {
		runtimeState["last_rebalance_at"] = lastRebalance
		settings.LastRebalanceAt = lastRebalance
		settings.NextEligibleRebalance = lastRebalance.Add(settings.RebalanceInterval())
	} else {
		delete(runtimeState, "last_rebalance_at")
		settings.LastRebalanceAt = time.Time{}
		settings.NextEligibleRebalance = time.Time{}
	}
	runtimeState[runtimeSettingsMetadataKey] = settings.ToRuntimeState()
}

func lastRebalanceFromRuntimeState(runtimeState map[string]interface{}) (time.Time, bool) {
	if len(runtimeState) == 0 {
		return time.Time{}, false
	}
	if last, ok := timeFromRuntimeValue(runtimeState["last_rebalance_at"]); ok {
		return last, true
	}
	if settings, ok := runtimeState[runtimeSettingsMetadataKey].(map[string]interface{}); ok {
		return timeFromRuntimeValue(settings["last_rebalance_at"])
	}
	return time.Time{}, false
}

func timeFromRuntimeValue(value interface{}) (time.Time, bool) {
	switch typed := value.(type) {
	case time.Time:
		if typed.IsZero() {
			return time.Time{}, false
		}
		return typed.UTC(), true
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return time.Time{}, false
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05 -0700 MST", "2006-01-02 15:04:05Z07:00"} {
			parsed, err := time.Parse(layout, trimmed)
			if err == nil && !parsed.IsZero() {
				return parsed.UTC(), true
			}
		}
		return time.Time{}, false
	default:
		return time.Time{}, false
	}
}

func symbolsFromAgentRun(run db.AgentRunSummary) []string {
	if len(run.Parameters) > 0 {
		if raw, ok := run.Parameters["symbols"]; ok {
			switch typed := raw.(type) {
			case []string:
				return normalizeSymbols(typed)
			case []interface{}:
				symbols := make([]string, 0, len(typed))
				for _, item := range typed {
					if symbol, ok := item.(string); ok {
						symbols = append(symbols, symbol)
					}
				}
				return normalizeSymbols(symbols)
			case string:
				return normalizeSymbols(strings.Split(typed, ","))
			}
		}
	}
	return normalizeSymbols([]string{run.Symbol})
}

func stringFromMap(values map[string]interface{}, key string) (string, bool) {
	if len(values) == 0 {
		return "", false
	}
	value, ok := values[key]
	if !ok {
		return "", false
	}
	switch typed := value.(type) {
	case string:
		trimmed := strings.TrimSpace(typed)
		return trimmed, trimmed != ""
	default:
		return "", false
	}
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
