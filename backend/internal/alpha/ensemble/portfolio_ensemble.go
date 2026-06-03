package ensemble

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

const StrategyName = "multi_engine_ensemble"

type Sleeve struct {
	Name              string
	Strategy          backtest.PortfolioStrategy
	Weight            float64
	RealizedVolAnnual float64
}

type SleeveOutput struct {
	Name              string
	Output            backtest.PortfolioOutput
	Weight            float64
	RealizedVolAnnual float64
}

type MultiEngineEnsembleStrategy struct {
	Sleeves        []Sleeve
	RiskConfig     PortfolioRiskConfig
	RiskState      PortfolioRiskState
	SectorBySymbol map[string]string
}

func NewMultiEngineEnsembleStrategy(
	sleeves []Sleeve,
	riskConfig PortfolioRiskConfig,
	sectorBySymbol map[string]string,
) *MultiEngineEnsembleStrategy {
	return &MultiEngineEnsembleStrategy{
		Sleeves:        append([]Sleeve(nil), sleeves...),
		RiskConfig:     riskConfig.withDefaults(),
		SectorBySymbol: sectorBySymbol,
	}
}

func (s *MultiEngineEnsembleStrategy) GeneratePortfolioSignals(ctx context.Context, panel backtest.AlignedBars) ([]backtest.PortfolioOutput, error) {
	if s == nil {
		return nil, fmt.Errorf("multi-engine ensemble is nil")
	}
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

func (s *MultiEngineEnsembleStrategy) EvaluatePortfolioLatest(ctx context.Context, panel backtest.AlignedBars) (backtest.PortfolioOutput, error) {
	if s == nil {
		return backtest.PortfolioOutput{}, fmt.Errorf("multi-engine ensemble is nil")
	}
	if len(panel.Times) == 0 {
		return backtest.PortfolioOutput{}, fmt.Errorf("aligned panel has no timestamps")
	}
	sleeveOutputs := make([]SleeveOutput, 0, len(s.Sleeves))
	for _, sleeve := range s.Sleeves {
		if sleeve.Strategy == nil {
			continue
		}
		output, err := sleeve.Strategy.EvaluatePortfolioLatest(ctx, panel)
		if err != nil {
			return backtest.PortfolioOutput{}, fmt.Errorf("evaluate sleeve %s: %w", sleeve.Name, err)
		}
		name := sleeve.Name
		if name == "" {
			name = sleeve.Strategy.Name()
		}
		sleeveOutputs = append(sleeveOutputs, SleeveOutput{
			Name:              name,
			Output:            output,
			Weight:            sleeve.Weight,
			RealizedVolAnnual: sleeve.RealizedVolAnnual,
		})
	}
	return CombineSleeveOutputs(panel.Times[len(panel.Times)-1], sleeveOutputs, s.RiskConfig, s.RiskState, s.SectorBySymbol), nil
}

func (s *MultiEngineEnsembleStrategy) Universe() []string {
	seen := make(map[string]bool)
	out := make([]string, 0)
	if s == nil {
		return out
	}
	for _, sleeve := range s.Sleeves {
		if sleeve.Strategy == nil {
			continue
		}
		for _, symbol := range sleeve.Strategy.Universe() {
			if seen[symbol] {
				continue
			}
			seen[symbol] = true
			out = append(out, symbol)
		}
	}
	return out
}

func (s *MultiEngineEnsembleStrategy) Name() string {
	return StrategyName
}

func CombineSleeveOutputs(
	t time.Time,
	sleeves []SleeveOutput,
	riskConfig PortfolioRiskConfig,
	riskState PortfolioRiskState,
	sectorBySymbol map[string]string,
) backtest.PortfolioOutput {
	weights := ComputeSleeveWeights(sleeves)
	rawTargets := make(map[string]backtest.TargetPosition)
	contributions := make(map[string]map[string]float64)
	for _, sleeve := range sleeves {
		weight := weights[sleeve.Name]
		if weight == 0 {
			continue
		}
		for symbol, target := range sleeve.Output.Targets {
			adjustedWeight := weight * target.TargetWeight
			current := rawTargets[symbol]
			current.Symbol = symbol
			current.TargetWeight += adjustedWeight
			current.AlphaScore += weight * clip(target.AlphaScore, -3, 3)
			current.Confidence = math.Max(current.Confidence, target.Confidence)
			current.Engine = StrategyName
			current.RegimeLabel = target.RegimeLabel
			current.Metadata = cloneMetadata(current.Metadata)
			if contributions[symbol] == nil {
				contributions[symbol] = make(map[string]float64)
			}
			contributions[symbol][sleeve.Name] += adjustedWeight
			rawTargets[symbol] = current
		}
	}
	for symbol, target := range rawTargets {
		target.Side = backtest.PositionSideLong
		if target.TargetWeight < 0 {
			target.Side = backtest.PositionSideShort
		}
		target.Metadata = cloneMetadata(target.Metadata)
		target.Metadata["sleeve_contributions"] = contributions[symbol]
		rawTargets[symbol] = target
	}

	risk := ApplyPortfolioRiskOverlay(rawTargets, riskConfig, riskState, sectorBySymbol)
	return backtest.PortfolioOutput{
		Time:          t,
		Targets:       risk.Targets,
		GrossExposure: risk.GrossExposure,
		NetExposure:   risk.NetExposure,
		CashWeight:    math.Max(0, 1-risk.GrossExposure),
		EngineMetadata: map[string]interface{}{
			"engine":         StrategyName,
			"sleeve_count":   len(sleeves),
			"sleeve_weights": weights,
			"risk_scalar":    risk.RiskScalar,
			"realized_vol":   risk.RealizedVol,
			"drawdown":       risk.Drawdown,
			"risk_reasons":   risk.Reasons,
		},
	}
}

func ComputeSleeveWeights(sleeves []SleeveOutput) map[string]float64 {
	weights := make(map[string]float64, len(sleeves))
	var explicitTotal float64
	var invVolTotal float64
	for _, sleeve := range sleeves {
		if sleeve.Weight > 0 {
			explicitTotal += sleeve.Weight
			continue
		}
		if sleeve.RealizedVolAnnual > 0 {
			invVolTotal += 1 / sleeve.RealizedVolAnnual
		}
	}
	if explicitTotal > 0 {
		for _, sleeve := range sleeves {
			if sleeve.Weight > 0 {
				weights[sleeve.Name] = sleeve.Weight / explicitTotal
			}
		}
		return weights
	}
	if invVolTotal > 0 {
		for _, sleeve := range sleeves {
			if sleeve.RealizedVolAnnual > 0 {
				weights[sleeve.Name] = (1 / sleeve.RealizedVolAnnual) / invVolTotal
			}
		}
		return weights
	}
	if len(sleeves) == 0 {
		return weights
	}
	equal := 1 / float64(len(sleeves))
	for _, sleeve := range sleeves {
		weights[sleeve.Name] = equal
	}
	return weights
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
