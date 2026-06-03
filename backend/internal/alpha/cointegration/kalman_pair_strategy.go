package cointegration

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

type KalmanPairStrategy struct {
	cfg              KalmanPairConfig
	filter           *KalmanHedgeFilter
	shortEligibility ShortEligibilityProvider
	retester         PairRetestProvider
}

func NewKalmanPairStrategy(
	cfg KalmanPairConfig,
	shortEligibility ShortEligibilityProvider,
	retester PairRetestProvider,
) *KalmanPairStrategy {
	cfg.SymbolY = strings.ToUpper(strings.TrimSpace(cfg.SymbolY))
	cfg.SymbolX = strings.ToUpper(strings.TrimSpace(cfg.SymbolX))
	cfg = cfg.withDefaults()
	return &KalmanPairStrategy{
		cfg:              cfg,
		filter:           NewKalmanHedgeFilter(cfg),
		shortEligibility: shortEligibility,
		retester:         retester,
	}
}

func (s *KalmanPairStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	if s == nil {
		return nil, fmt.Errorf("kalman pair strategy is nil")
	}
	previousFilter := s.filter
	defer func() { s.filter = previousFilter }()
	s.Reset()

	outputs := make([]backtest.PortfolioOutput, len(panel.Times))
	for i := range panel.Times {
		output, err := s.EvaluatePortfolioLatest(ctx, panelPrefix(panel, i+1))
		if err != nil {
			return nil, err
		}
		outputs[i] = output
	}
	return outputs, nil
}

func (s *KalmanPairStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	if s == nil {
		return backtest.PortfolioOutput{}, fmt.Errorf("kalman pair strategy is nil")
	}
	if err := s.validatePanel(panel); err != nil {
		return backtest.PortfolioOutput{}, err
	}
	update, err := s.processNewBars(panel)
	if err != nil {
		return backtest.PortfolioOutput{}, err
	}
	index := len(panel.Times) - 1
	t := panel.Times[index]
	state := s.filter.State()

	if state.Quarantined || state.PositionState == PairQuarantined {
		return s.holdOutput(t, "pair_quarantined"), nil
	}
	if math.Abs(update.Z) >= s.cfg.StopZ {
		state.PositionState = PairQuarantined
		state.Quarantined = true
		state.NeedsRetest = true
		s.filter.SetState(state)
		return s.flattenOutput(t, "stop_z_quarantine", update), nil
	}
	if state.PositionState != PairFlat && math.Abs(update.Z) <= s.cfg.ExitZ {
		state.PositionState = PairFlat
		s.filter.SetState(state)
		return s.flattenOutput(t, "exit_z_flatten", update), nil
	}
	if state.PositionState != PairFlat {
		return s.holdOutput(t, "hold_open_pair"), nil
	}
	if math.Abs(update.Z) < s.cfg.EntryZ {
		return s.holdOutput(t, "z_below_entry"), nil
	}
	if update.Beta <= 0 {
		return s.holdOutput(t, "non_positive_beta"), nil
	}
	if ok, reason, err := s.retestIfDue(ctx, t); err != nil {
		return backtest.PortfolioOutput{}, err
	} else if !ok {
		return s.holdOutput(t, reason), nil
	}

	targets, positionState := s.entryTargets(update)
	if err := s.validateShortLeg(ctx, t, targets); err != nil {
		return backtest.PortfolioOutput{}, err
	}
	if len(targets) == 0 {
		return s.holdOutput(t, "entry_weights_zero"), nil
	}
	state = s.filter.State()
	state.PositionState = positionState
	s.filter.SetState(state)
	return backtest.PortfolioOutput{
		Time:          t,
		Targets:       targets,
		GrossExposure: grossExposure(targets),
		NetExposure:   netExposure(targets),
		CashWeight:    math.Max(0, 1-grossExposure(targets)),
		EngineMetadata: map[string]interface{}{
			"engine":                   StrategyName,
			"action":                   "enter_pair",
			"z":                        update.Z,
			"alpha":                    update.Alpha,
			"beta":                     update.Beta,
			"position":                 string(positionState),
			"short_live_gate_required": s.cfg.RequireShortable,
		},
	}, nil
}

func (s *KalmanPairStrategy) Universe() []string {
	if s == nil {
		return nil
	}
	return []string{s.cfg.SymbolY, s.cfg.SymbolX}
}

func (s *KalmanPairStrategy) Name() string {
	return StrategyName
}

func (s *KalmanPairStrategy) State() KalmanPairState {
	if s == nil || s.filter == nil {
		return KalmanPairState{}
	}
	return s.filter.State()
}

func (s *KalmanPairStrategy) Reset() {
	if s == nil {
		return
	}
	s.filter = NewKalmanHedgeFilter(s.cfg)
}

func (s *KalmanPairStrategy) validatePanel(panel backtest.AlignedBars) error {
	if s.cfg.SymbolY == "" || s.cfg.SymbolX == "" {
		return fmt.Errorf("kalman pair strategy requires symbol_y and symbol_x")
	}
	if len(panel.Times) == 0 {
		return fmt.Errorf("aligned panel has no timestamps")
	}
	if len(panel.Bars[s.cfg.SymbolY]) < len(panel.Times) {
		return fmt.Errorf("panel missing bars for symbol_y %s", s.cfg.SymbolY)
	}
	if len(panel.Bars[s.cfg.SymbolX]) < len(panel.Times) {
		return fmt.Errorf("panel missing bars for symbol_x %s", s.cfg.SymbolX)
	}
	return nil
}

func (s *KalmanPairStrategy) processNewBars(panel backtest.AlignedBars) (KalmanUpdate, error) {
	state := s.filter.State()
	if state.ProcessedBars > len(panel.Times) {
		s.Reset()
		state = s.filter.State()
	}
	var update KalmanUpdate
	for i := state.ProcessedBars; i < len(panel.Times); i++ {
		barY := panel.Bars[s.cfg.SymbolY][i]
		barX := panel.Bars[s.cfg.SymbolX][i]
		if barY.Close <= 0 || barX.Close <= 0 {
			return KalmanUpdate{}, fmt.Errorf("non-positive pair close at %s", panel.Times[i])
		}
		var err error
		update, err = s.filter.Update(math.Log(barX.Close), math.Log(barY.Close))
		if err != nil {
			return KalmanUpdate{}, err
		}
		next := s.filter.State()
		next.ProcessedBars = i + 1
		s.filter.SetState(next)
	}
	return update, nil
}

func (s *KalmanPairStrategy) entryTargets(update KalmanUpdate) (map[string]backtest.TargetPosition, PairPositionState) {
	yAbs, xAbs := pairLegWeights(update.Beta, s.cfg.MaxGrossWeight, s.cfg.MaxLegWeight)
	targets := make(map[string]backtest.TargetPosition, 2)
	if yAbs <= 0 || xAbs <= 0 {
		return targets, PairFlat
	}

	if update.Z > 0 {
		targets[s.cfg.SymbolY] = s.target(s.cfg.SymbolY, -yAbs, -update.Z, update, "short_y")
		targets[s.cfg.SymbolX] = s.target(s.cfg.SymbolX, xAbs, update.Z, update, "long_x")
		return targets, PairShortYLongX
	}
	targets[s.cfg.SymbolY] = s.target(s.cfg.SymbolY, yAbs, -update.Z, update, "long_y")
	targets[s.cfg.SymbolX] = s.target(s.cfg.SymbolX, -xAbs, update.Z, update, "short_x")
	return targets, PairLongYShortX
}

func (s *KalmanPairStrategy) target(symbol string, weight, score float64, update KalmanUpdate, leg string) backtest.TargetPosition {
	side := backtest.PositionSideLong
	if weight < 0 {
		side = backtest.PositionSideShort
	}
	return backtest.TargetPosition{
		Symbol:       symbol,
		TargetWeight: weight,
		AlphaScore:   clip(score, -3, 3),
		Confidence:   math.Min(1, math.Abs(update.Z)/s.cfg.StopZ),
		Side:         side,
		Engine:       StrategyName,
		Metadata: map[string]interface{}{
			"leg":                 leg,
			"z":                   update.Z,
			"alpha":               update.Alpha,
			"beta":                update.Beta,
			"innovation_variance": update.InnovationVariance,
		},
	}
}

func (s *KalmanPairStrategy) flattenOutput(t time.Time, reason string, update KalmanUpdate) backtest.PortfolioOutput {
	targets := map[string]backtest.TargetPosition{
		s.cfg.SymbolY: s.target(s.cfg.SymbolY, 0, 0, update, "flatten_y"),
		s.cfg.SymbolX: s.target(s.cfg.SymbolX, 0, 0, update, "flatten_x"),
	}
	return backtest.PortfolioOutput{
		Time:          t,
		Targets:       targets,
		GrossExposure: 0,
		NetExposure:   0,
		CashWeight:    1,
		EngineMetadata: map[string]interface{}{
			"engine": StrategyName,
			"action": "flatten_pair",
			"reason": reason,
			"z":      update.Z,
		},
	}
}

func (s *KalmanPairStrategy) holdOutput(t time.Time, reason string) backtest.PortfolioOutput {
	state := s.filter.State()
	return backtest.PortfolioOutput{
		Time:       t,
		Targets:    map[string]backtest.TargetPosition{},
		CashWeight: 1,
		EngineMetadata: map[string]interface{}{
			"engine":   StrategyName,
			"action":   "hold_targets",
			"reason":   reason,
			"z":        state.LastZ,
			"position": string(state.PositionState),
		},
	}
}

func (s *KalmanPairStrategy) retestIfDue(ctx context.Context, t time.Time) (bool, string, error) {
	if s.retester == nil {
		return true, "", nil
	}
	state := s.filter.State()
	if !state.LastRetest.IsZero() && t.Sub(state.LastRetest) < s.cfg.RetestCadence && !state.NeedsRetest {
		return true, "", nil
	}
	ok, reason, err := s.retester.PairStillApproved(ctx, s.cfg.SymbolY, s.cfg.SymbolX, t)
	if err != nil {
		return false, "", err
	}
	state.LastRetest = t
	state.NeedsRetest = !ok
	s.filter.SetState(state)
	if !ok {
		if reason == "" {
			reason = "pair_retest_failed"
		}
		return false, reason, nil
	}
	return true, "", nil
}

func (s *KalmanPairStrategy) validateShortLeg(ctx context.Context, t time.Time, targets map[string]backtest.TargetPosition) error {
	if !s.cfg.RequireShortable {
		return nil
	}
	if s.shortEligibility == nil {
		return fmt.Errorf("short eligibility provider is required when RequireShortable is true")
	}
	for symbol, target := range targets {
		if target.TargetWeight >= 0 {
			continue
		}
		ok, err := s.shortEligibility.IsShortable(ctx, symbol, t)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("short leg %s is not shortable/easy-to-borrow", symbol)
		}
	}
	return nil
}

func pairLegWeights(beta, gross, maxLeg float64) (float64, float64) {
	if beta <= 0 || gross <= 0 || maxLeg <= 0 {
		return 0, 0
	}
	xRatio := math.Abs(beta)
	yWeight := gross / (1 + xRatio)
	xWeight := gross - yWeight
	return math.Min(yWeight, maxLeg), math.Min(xWeight, maxLeg)
}

func grossExposure(targets map[string]backtest.TargetPosition) float64 {
	var gross float64
	for _, target := range targets {
		gross += math.Abs(target.TargetWeight)
	}
	return gross
}

func netExposure(targets map[string]backtest.TargetPosition) float64 {
	var net float64
	for _, target := range targets {
		net += target.TargetWeight
	}
	return net
}

func clip(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func panelPrefix(panel backtest.AlignedBars, length int) backtest.AlignedBars {
	if length > len(panel.Times) {
		length = len(panel.Times)
	}
	out := backtest.AlignedBars{
		Times:      append([]time.Time(nil), panel.Times[:length]...),
		Symbols:    append([]string(nil), panel.Symbols...),
		Bars:       make(map[string][]models.Bar, len(panel.Bars)),
		Timeframe:  panel.Timeframe,
		Feed:       panel.Feed,
		Adjustment: panel.Adjustment,
		Metadata:   panel.Metadata,
	}
	for symbol, bars := range panel.Bars {
		if length > len(bars) {
			out.Bars[symbol] = append([]models.Bar(nil), bars...)
			continue
		}
		out.Bars[symbol] = append([]models.Bar(nil), bars[:length]...)
	}
	return out
}
