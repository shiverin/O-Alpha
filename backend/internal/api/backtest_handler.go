package api

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oalpha/internal/alpha/cointegration"
	"github.com/oalpha/internal/alpha/momentum"
	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/internal/ml"
	"github.com/oalpha/pkg/models"
)

// RunBacktest executes a backtest for the requested symbol using the specified strategy.
func (h *Handler) RunBacktest(c *gin.Context) {
	var req models.BacktestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.StrategyType = strings.ToUpper(strings.TrimSpace(req.StrategyType))

	// Only enforce MA rules if they are explicitly using MA_CROSSOVER or defaulting to it
	if req.StrategyType == "MA_CROSSOVER" || req.StrategyType == "" {
		fast := req.FastPeriod
		if fast == 0 {
			fast = 10
		}
		slow := req.SlowPeriod
		if slow == 0 {
			slow = 30
		}

		if fast >= slow {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fast_period must be less than slow_period"})
			return
		}
	}

	end := time.Now().UTC()
	start := end.Add(-365 * 24 * time.Hour)
	if req.Start != nil {
		start = req.Start.UTC()
	}
	if req.End != nil {
		end = req.End.UTC()
	}

	if req.Timeframe == "" {
		req.Timeframe = "1Day"
	}
	initialCash := req.InitialCash
	if initialCash <= 0 {
		initialCash = 100000
	}
	if isPortfolioBacktestStrategy(req.StrategyType) {
		h.runPortfolioBacktest(c, req, start, end, initialCash)
		return
	}
	if strings.TrimSpace(req.Symbol) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required for single-symbol backtests"})
		return
	}

	bars, err := h.repo.GetBarsDataset(c.Request.Context(), req.Symbol, req.Timeframe, start, end, db.BarQueryOptions{
		Feed:       req.Feed,
		Adjustment: req.Adjustment,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(bars) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no bars found for symbol"})
		return
	}

	var expectedInterval time.Duration
	switch req.Timeframe {
	case "1Min":
		expectedInterval = time.Minute
	case "5Min":
		expectedInterval = 5 * time.Minute
	case "15Min":
		expectedInterval = 15 * time.Minute
	case "1Hour":
		expectedInterval = time.Hour
	default:
		expectedInterval = 24 * time.Hour
	}

	report, err := h.repo.ValidateData(c.Request.Context(), req.Symbol, req.Timeframe, start, end, expectedInterval)
	if err == nil && report != nil {
		if report.InvalidBars > 0 && (float64(report.InvalidBars)/float64(report.BarCount)) > 0.05 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         "backtest aborted: historical data quality is too poor",
				"invalid_count": report.InvalidBars,
				"total_scanned": report.BarCount,
			})
			return
		}
	}

	strat, err := h.buildSingleSymbolBacktestStrategy(c.Request.Context(), req, bars, start, end)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := backtest.RunBacktest(c.Request.Context(), bars, strat, initialCash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) buildSingleSymbolBacktestStrategy(
	ctx context.Context,
	req models.BacktestRequest,
	bars []models.Bar,
	start time.Time,
	end time.Time,
) (backtest.Strategy, error) {
	switch req.StrategyType {
	case "ML_META_LABEL":
		baseType := strings.ToUpper(strings.TrimSpace(stringParam(req.Parameters, "base_strategy_type", "")))
		if baseType == "" {
			baseType = strings.ToUpper(strings.TrimSpace(stringParam(req.Parameters, "base_strategy", "MA_CROSSOVER")))
		}
		base, err := buildBaseSingleSymbolStrategy(baseType, req)
		if err != nil {
			return nil, err
		}
		predictor, featureSpec, artifactThresholds, calibration, err := loadMLPredictorFromParams(req.Parameters)
		if err != nil {
			return nil, err
		}
		return &ml.MLMetaLabelStrategy{
			Symbol:         strings.ToUpper(strings.TrimSpace(req.Symbol)),
			BaseStrategy:   base,
			FeatureBuilder: ml.NewFeatureBuilder(featureSpec),
			Predictor:      predictor,
			Calibration:    calibration,
			Thresholds:     mlThresholdsFromParams(req.Parameters, artifactThresholds),
			MaxWeight:      floatParam(req.Parameters, "ml_max_weight", 0),
			ContextBars:    h.loadMLContextBars(ctx, strings.ToUpper(strings.TrimSpace(req.Symbol)), req.Timeframe, start, end, req.Feed, req.Adjustment, featureSpec),
		}, nil
	default:
		_ = bars
		return buildBaseSingleSymbolStrategy(req.StrategyType, req)
	}
}

func buildBaseSingleSymbolStrategy(strategyType string, req models.BacktestRequest) (backtest.Strategy, error) {
	switch strings.ToUpper(strings.TrimSpace(strategyType)) {
	case "", "MA_CROSSOVER":
		fast := req.FastPeriod
		if fast == 0 {
			fast = 10
		}
		slow := req.SlowPeriod
		if slow == 0 {
			slow = 30
		}
		if fast >= slow {
			return nil, fmt.Errorf("fast_period must be less than slow_period")
		}
		return backtest.NewMACrossoverStrategy(fast, slow), nil
	case "KALMAN":
		q := req.QNoise
		if q == 0 {
			q = 0.01
		}
		r := req.RNoise
		if r == 0 {
			r = 0.5
		}
		z := req.ZThreshold
		if z == 0 {
			z = 2.0
		}
		return backtest.NewKalmanStrategy(q, r, 20, z), nil
	default:
		return nil, fmt.Errorf("unsupported base strategy_type %s for single-symbol ML meta-labeling", strategyType)
	}
}

func loadMLPredictorFromParams(params map[string]interface{}) (*ml.LeavesPredictor, ml.FeatureSpec, ml.MLThresholds, ml.CalibrationModel, error) {
	modelPath := strings.TrimSpace(stringParam(params, "model_path", ""))
	metadataPath := strings.TrimSpace(stringParam(params, "metadata_path", ""))
	registryRoot := strings.TrimSpace(stringParam(params, "model_registry_root", ""))
	modelName := strings.TrimSpace(stringParam(params, "model_name", "ml_meta_label"))
	strategyScope := strings.TrimSpace(stringParam(params, "strategy_scope", ""))

	if metadataPath != "" {
		artifact, err := ml.ReadModelArtifact(metadataPath)
		if err != nil {
			return nil, ml.FeatureSpec{}, ml.MLThresholds{}, ml.CalibrationModel{}, err
		}
		root := filepath.Dir(metadataPath)
		predictor, err := ml.NewLeavesPredictor(artifact.ModelPath(root), artifact.FeatureSpec, artifact.Version())
		if err != nil {
			return nil, ml.FeatureSpec{}, ml.MLThresholds{}, ml.CalibrationModel{}, err
		}
		return predictor, artifact.FeatureSpec, artifact.Thresholds, artifact.Calibration, nil
	}
	if registryRoot != "" {
		predictor, artifact, err := ml.NewModelRegistry(registryRoot).LoadLatestPromotedPredictor(modelName, strategyScope)
		if err != nil {
			return nil, ml.FeatureSpec{}, ml.MLThresholds{}, ml.CalibrationModel{}, err
		}
		return predictor, artifact.FeatureSpec, artifact.Thresholds, artifact.Calibration, nil
	}
	if modelPath == "" {
		return nil, ml.FeatureSpec{}, ml.MLThresholds{}, ml.CalibrationModel{}, fmt.Errorf("ML_META_LABEL requires model_path, metadata_path, or model_registry_root")
	}
	featureSpec := ml.DefaultFeatureSpec()
	predictor, err := ml.NewLeavesPredictor(modelPath, featureSpec, stringParam(params, "model_version", modelPath))
	if err != nil {
		return nil, ml.FeatureSpec{}, ml.MLThresholds{}, ml.CalibrationModel{}, err
	}
	return predictor, featureSpec, ml.DefaultMLThresholds(), ml.CalibrationModel{}, nil
}

func mlThresholdsFromParams(params map[string]interface{}, thresholds ml.MLThresholds) ml.MLThresholds {
	if thresholds.EnterLong <= 0 {
		thresholds = ml.DefaultMLThresholds()
	}
	if value := floatParam(params, "enter_long", thresholds.EnterLong); value > 0 {
		thresholds.EnterLong = value
	}
	if value := floatParam(params, "reduce", thresholds.Reduce); value > 0 {
		thresholds.Reduce = value
	}
	if value := floatParam(params, "bet_sizing_slope", thresholds.BetSizingSlope); value > 0 {
		thresholds.BetSizingSlope = value
	}
	thresholds.FailOpenOnError = boolParam(params, "fail_open_on_error", thresholds.FailOpenOnError)
	thresholds.PassThroughExits = boolParam(params, "pass_through_exits", thresholds.PassThroughExits)
	return thresholds
}

func (h *Handler) loadMLContextBars(
	ctx context.Context,
	symbol string,
	timeframe string,
	start time.Time,
	end time.Time,
	feed string,
	adjustment string,
	featureSpec ml.FeatureSpec,
) map[string][]models.Bar {
	if h == nil || h.repo == nil {
		return nil
	}
	contextSymbols := featureSpec.ContextSymbols
	if len(contextSymbols) == 0 {
		contextSymbols = ml.DefaultFeatureSpec().ContextSymbols
	}
	out := make(map[string][]models.Bar)
	for _, contextSymbol := range contextSymbols {
		contextSymbol = strings.ToUpper(strings.TrimSpace(contextSymbol))
		if contextSymbol == "" || contextSymbol == symbol {
			continue
		}
		bars, err := h.repo.GetBarsDataset(ctx, contextSymbol, timeframe, start, end, db.BarQueryOptions{
			Feed:       feed,
			Adjustment: adjustment,
		})
		if err == nil && len(bars) > 0 {
			out[contextSymbol] = bars
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (h *Handler) runPortfolioBacktest(c *gin.Context, req models.BacktestRequest, start, end time.Time, initialCash float64) {
	symbols := normalizeRequestSymbols(req)
	if len(symbols) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "portfolio backtests require at least two symbols"})
		return
	}
	panel, err := h.repo.GetBarsMulti(c.Request.Context(), symbols, req.Timeframe, start, end, db.BarQueryOptions{
		Feed:       req.Feed,
		Adjustment: req.Adjustment,
		AlignMode:  backtest.AlignInnerJoin,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(panel.Times) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no aligned bars found for symbols"})
		return
	}

	strategy, allowShorts, err := buildPortfolioStrategy(req, panel.Symbols)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg := backtest.PortfolioBacktestConfig{
		InitialCash:      initialCash,
		AllowShorts:      allowShorts,
		MaxGrossExposure: floatParam(req.Parameters, "max_gross_exposure", 1),
		MaxNetExposure:   floatParam(req.Parameters, "max_net_exposure", 1),
		MaxSymbolWeight:  floatParam(req.Parameters, "max_symbol_weight", 1),
		CostModel: backtest.CostModel{
			DefaultSpreadBps:   floatParam(req.Parameters, "default_spread_bps", backtest.DefaultCostModel().DefaultSpreadBps),
			SlippageBps:        floatParam(req.Parameters, "slippage_bps", backtest.DefaultCostModel().SlippageBps),
			BorrowFeeBpsAnnual: floatParam(req.Parameters, "borrow_fee_bps_annual", backtest.DefaultCostModel().BorrowFeeBpsAnnual),
			SECFeesBpsSell:     floatParam(req.Parameters, "sec_fees_bps_sell", backtest.DefaultCostModel().SECFeesBpsSell),
		},
	}
	result, err := backtest.RunPortfolioBacktest(c.Request.Context(), panel, strategy, cfg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func buildPortfolioStrategy(req models.BacktestRequest, symbols []string) (backtest.PortfolioStrategy, bool, error) {
	switch req.StrategyType {
	case "XSEC_MOMENTUM":
		cfg := momentum.DefaultCrossSectionalMomentumConfig()
		cfg.FormationDays = intParam(req.Parameters, "formation_days", cfg.FormationDays)
		cfg.SkipDays = intParam(req.Parameters, "skip_days", cfg.SkipDays)
		cfg.TopFraction = floatParam(req.Parameters, "top_fraction", cfg.TopFraction)
		cfg.MinPositions = intParam(req.Parameters, "min_positions", cfg.MinPositions)
		cfg.MaxPositions = intParam(req.Parameters, "max_positions", cfg.MaxPositions)
		cfg.VolLookbackDays = intParam(req.Parameters, "vol_lookback_days", cfg.VolLookbackDays)
		cfg.TargetVolAnnual = floatParam(req.Parameters, "target_vol_annual", cfg.TargetVolAnnual)
		cfg.MaxSymbolWeight = floatParam(req.Parameters, "max_symbol_weight", cfg.MaxSymbolWeight)
		cfg.MaxSectorWeight = floatParam(req.Parameters, "max_sector_weight", cfg.MaxSectorWeight)
		cfg.MinPrice = floatParam(req.Parameters, "min_price", cfg.MinPrice)
		cfg.MinMedianDollarVolume = floatParam(req.Parameters, "min_median_dollar_volume", cfg.MinMedianDollarVolume)
		cfg.MinDataCompleteness = floatParam(req.Parameters, "min_data_completeness", cfg.MinDataCompleteness)
		cfg.RebalanceFrequency = stringParam(req.Parameters, "rebalance_frequency", cfg.RebalanceFrequency)
		return momentum.NewCrossSectionalMomentumStrategy(symbols, cfg, nil), false, nil
	case "KALMAN_COINTEGRATION":
		symbolY := strings.ToUpper(strings.TrimSpace(stringParam(req.Parameters, "symbol_y", "")))
		symbolX := strings.ToUpper(strings.TrimSpace(stringParam(req.Parameters, "symbol_x", "")))
		if symbolY == "" && len(symbols) > 0 {
			symbolY = symbols[0]
		}
		if symbolX == "" && len(symbols) > 1 {
			symbolX = symbols[1]
		}
		if symbolY == "" || symbolX == "" || symbolY == symbolX {
			return nil, false, fmt.Errorf("KALMAN_COINTEGRATION requires distinct symbol_y and symbol_x")
		}
		cfg := cointegration.DefaultKalmanPairConfig(symbolY, symbolX)
		cfg.QAlpha = floatParam(req.Parameters, "q_alpha", cfg.QAlpha)
		cfg.QBeta = floatParam(req.Parameters, "q_beta", cfg.QBeta)
		cfg.R = floatParam(req.Parameters, "r", cfg.R)
		cfg.EntryZ = floatParam(req.Parameters, "entry_z", cfg.EntryZ)
		cfg.ExitZ = floatParam(req.Parameters, "exit_z", cfg.ExitZ)
		cfg.StopZ = floatParam(req.Parameters, "stop_z", cfg.StopZ)
		cfg.MaxGrossWeight = floatParam(req.Parameters, "max_pair_gross_weight", cfg.MaxGrossWeight)
		cfg.MaxLegWeight = floatParam(req.Parameters, "max_leg_weight", cfg.MaxLegWeight)
		return cointegration.NewKalmanPairStrategy(cfg, nil, nil), true, nil
	default:
		return nil, false, fmt.Errorf("unsupported portfolio strategy_type %s", req.StrategyType)
	}
}

func isPortfolioBacktestStrategy(strategyType string) bool {
	switch strings.ToUpper(strings.TrimSpace(strategyType)) {
	case "XSEC_MOMENTUM", "KALMAN_COINTEGRATION":
		return true
	default:
		return false
	}
}

func normalizeRequestSymbols(req models.BacktestRequest) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(req.Symbols)+1)
	for _, symbol := range append(req.Symbols, req.Symbol) {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func floatParam(params map[string]interface{}, key string, fallback float64) float64 {
	if params == nil {
		return fallback
	}
	switch value := params[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	default:
		return fallback
	}
}

func intParam(params map[string]interface{}, key string, fallback int) int {
	value := floatParam(params, key, float64(fallback))
	if value <= 0 {
		return fallback
	}
	return int(value)
}

func stringParam(params map[string]interface{}, key string, fallback string) string {
	if params == nil {
		return fallback
	}
	if value, ok := params[key].(string); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func boolParam(params map[string]interface{}, key string, fallback bool) bool {
	if params == nil {
		return fallback
	}
	switch value := params[key].(type) {
	case bool:
		return value
	case string:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true", "1", "yes", "y":
			return true
		case "false", "0", "no", "n":
			return false
		}
	case float64:
		return value != 0
	case int:
		return value != 0
	}
	return fallback
}
