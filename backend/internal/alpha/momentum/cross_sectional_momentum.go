package momentum

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

const (
	StrategyName      = "xsec_momentum"
	RebalanceMonthly  = "monthly"
	RebalanceDaily    = "daily"
	actionHoldTargets = "hold_targets"
)

type CrossSectionalMomentumConfig struct {
	FormationDays         int
	SkipDays              int
	RebalanceFrequency    string
	TopFraction           float64
	MinPositions          int
	MaxPositions          int
	VolLookbackDays       int
	TargetVolAnnual       float64
	MaxSymbolWeight       float64
	MaxSectorWeight       float64
	LongOnly              bool
	MinPrice              float64
	MinMedianDollarVolume float64
	MinDataCompleteness   float64
}

type CrossSectionalMomentumStrategy struct {
	cfg            CrossSectionalMomentumConfig
	universe       []string
	sectorBySymbol map[string]string
	lastRebalance  time.Time
	currentTargets map[string]backtest.TargetPosition
}

func DefaultCrossSectionalMomentumConfig() CrossSectionalMomentumConfig {
	return CrossSectionalMomentumConfig{
		FormationDays:         252,
		SkipDays:              21,
		RebalanceFrequency:    RebalanceMonthly,
		TopFraction:           0.15,
		MinPositions:          10,
		MaxPositions:          25,
		VolLookbackDays:       60,
		TargetVolAnnual:       0.12,
		MaxSymbolWeight:       0.10,
		MaxSectorWeight:       0.35,
		LongOnly:              true,
		MinPrice:              5,
		MinMedianDollarVolume: 5_000_000,
		MinDataCompleteness:   0.98,
	}
}

func NewCrossSectionalMomentumStrategy(
	universe []string,
	cfg CrossSectionalMomentumConfig,
	sectorBySymbol map[string]string,
) *CrossSectionalMomentumStrategy {
	cfg = cfg.withDefaults()
	return &CrossSectionalMomentumStrategy{
		cfg:            cfg,
		universe:       normalizeSymbols(universe),
		sectorBySymbol: normalizeSectorMap(sectorBySymbol),
		currentTargets: make(map[string]backtest.TargetPosition),
	}
}

func (s *CrossSectionalMomentumStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	if s == nil {
		return nil, fmt.Errorf("cross-sectional momentum strategy is nil")
	}
	previousRebalance := s.lastRebalance
	previousTargets := cloneTargets(s.currentTargets)
	s.lastRebalance = time.Time{}
	s.currentTargets = make(map[string]backtest.TargetPosition)
	defer func() {
		s.lastRebalance = previousRebalance
		s.currentTargets = previousTargets
	}()

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

func (s *CrossSectionalMomentumStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	_ = ctx
	if s == nil {
		return backtest.PortfolioOutput{}, fmt.Errorf("cross-sectional momentum strategy is nil")
	}
	if len(panel.Times) == 0 {
		return backtest.PortfolioOutput{}, fmt.Errorf("aligned panel has no timestamps")
	}
	index := len(panel.Times) - 1
	t := panel.Times[index]
	if !s.shouldRebalance(t) {
		return backtest.PortfolioOutput{
			Time:          t,
			Targets:       map[string]backtest.TargetPosition{},
			GrossExposure: grossExposure(s.currentTargets),
			NetExposure:   netExposure(s.currentTargets),
			CashWeight:    math.Max(0, 1-grossExposure(s.currentTargets)),
			EngineMetadata: map[string]interface{}{
				"engine":    StrategyName,
				"rebalance": false,
				"action":    actionHoldTargets,
			},
		}, nil
	}
	if !hasRequiredLookback(index, s.cfg) {
		return backtest.PortfolioOutput{
			Time:       t,
			Targets:    map[string]backtest.TargetPosition{},
			CashWeight: 1,
			EngineMetadata: map[string]interface{}{
				"engine":    StrategyName,
				"rebalance": false,
				"reason":    "insufficient_lookback",
			},
		}, nil
	}

	eligible := FilterUniverse(panel, index, s.candidateUniverse(panel), s.cfg)
	scores, err := ComputeMomentumScores(panel, eligible, index, s.cfg)
	if err != nil {
		return backtest.PortfolioOutput{}, err
	}
	selected := SelectTopMomentum(scores, s.cfg)
	targets, metadata := BuildMomentumTargets(panel, index, selected, s.cfg, s.sectorBySymbol)
	s.lastRebalance = t
	s.currentTargets = cloneTargets(targets)

	return backtest.PortfolioOutput{
		Time:           t,
		Targets:        targets,
		GrossExposure:  grossExposure(targets),
		NetExposure:    netExposure(targets),
		CashWeight:     math.Max(0, 1-grossExposure(targets)),
		EngineMetadata: metadata,
	}, nil
}

func (s *CrossSectionalMomentumStrategy) Universe() []string {
	if s == nil {
		return nil
	}
	return append([]string(nil), s.universe...)
}

func (s *CrossSectionalMomentumStrategy) Name() string {
	return StrategyName
}

func (s *CrossSectionalMomentumStrategy) shouldRebalance(t time.Time) bool {
	if s.lastRebalance.IsZero() {
		return true
	}
	switch strings.ToLower(s.cfg.RebalanceFrequency) {
	case RebalanceDaily:
		return !sameDate(t, s.lastRebalance)
	default:
		return t.Year() != s.lastRebalance.Year() || t.Month() != s.lastRebalance.Month()
	}
}

func (s *CrossSectionalMomentumStrategy) candidateUniverse(panel backtest.AlignedBars) []string {
	if len(s.universe) > 0 {
		return append([]string(nil), s.universe...)
	}
	return append([]string(nil), panel.Symbols...)
}

func (c CrossSectionalMomentumConfig) withDefaults() CrossSectionalMomentumConfig {
	defaults := DefaultCrossSectionalMomentumConfig()
	if c.FormationDays <= 0 {
		c.FormationDays = defaults.FormationDays
	}
	if c.SkipDays < 0 {
		c.SkipDays = defaults.SkipDays
	}
	if c.RebalanceFrequency == "" {
		c.RebalanceFrequency = defaults.RebalanceFrequency
	}
	if c.TopFraction <= 0 {
		c.TopFraction = defaults.TopFraction
	}
	if c.TopFraction > 1 {
		c.TopFraction = 1
	}
	if c.MinPositions <= 0 {
		c.MinPositions = defaults.MinPositions
	}
	if c.MaxPositions <= 0 {
		c.MaxPositions = defaults.MaxPositions
	}
	if c.MaxPositions < c.MinPositions {
		c.MaxPositions = c.MinPositions
	}
	if c.VolLookbackDays <= 0 {
		c.VolLookbackDays = defaults.VolLookbackDays
	}
	if c.TargetVolAnnual <= 0 {
		c.TargetVolAnnual = defaults.TargetVolAnnual
	}
	if c.MaxSymbolWeight <= 0 {
		c.MaxSymbolWeight = defaults.MaxSymbolWeight
	}
	if c.MaxSymbolWeight > 1 {
		c.MaxSymbolWeight = 1
	}
	if c.MaxSectorWeight <= 0 {
		c.MaxSectorWeight = defaults.MaxSectorWeight
	}
	if c.MaxSectorWeight > 1 {
		c.MaxSectorWeight = 1
	}
	if c.MinPrice <= 0 {
		c.MinPrice = defaults.MinPrice
	}
	if c.MinMedianDollarVolume <= 0 {
		c.MinMedianDollarVolume = defaults.MinMedianDollarVolume
	}
	if c.MinDataCompleteness <= 0 {
		c.MinDataCompleteness = defaults.MinDataCompleteness
	}
	if c.MinDataCompleteness > 1 {
		c.MinDataCompleteness = 1
	}
	c.LongOnly = true
	return c
}

func hasRequiredLookback(index int, cfg CrossSectionalMomentumConfig) bool {
	return index-cfg.SkipDays-cfg.FormationDays >= 0 && index-cfg.VolLookbackDays >= 0
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

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
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

func cloneTargets(targets map[string]backtest.TargetPosition) map[string]backtest.TargetPosition {
	out := make(map[string]backtest.TargetPosition, len(targets))
	for symbol, target := range targets {
		out[symbol] = target
	}
	return out
}

func normalizeSymbols(symbols []string) []string {
	out := make([]string, 0, len(symbols))
	seen := make(map[string]bool, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func normalizeSectorMap(sectors map[string]string) map[string]string {
	out := make(map[string]string, len(sectors))
	for symbol, sector := range sectors {
		out[strings.ToUpper(strings.TrimSpace(symbol))] = strings.TrimSpace(sector)
	}
	return out
}
