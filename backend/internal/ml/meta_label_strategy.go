package ml

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

const MetaLabelEngineName = "ml_meta_label"

type MLMetaLabelStrategy struct {
	Symbol                 string
	BaseStrategy           backtest.Strategy
	FeatureBuilder         *FeatureBuilder
	Predictor              Predictor
	Calibration            CalibrationModel
	Thresholds             MLThresholds
	MaxWeight              float64
	ContextBars            map[string][]models.Bar
	SectorSymbol           string
	HMMRegimeProbabilities map[string]float64
}

func (s *MLMetaLabelStrategy) GenerateSignals(ctx context.Context, bars []models.Bar) ([]backtest.StrategyOutput, error) {
	if len(bars) == 0 {
		return nil, fmt.Errorf("ML meta-label strategy requires bars")
	}
	if s == nil || s.BaseStrategy == nil {
		return nil, fmt.Errorf("ML meta-label strategy requires a base strategy")
	}
	baseOutputs, err := s.BaseStrategy.GenerateSignals(ctx, bars)
	if err != nil {
		return nil, err
	}
	if len(baseOutputs) != len(bars) {
		return nil, fmt.Errorf("base outputs length %d does not match bars length %d", len(baseOutputs), len(bars))
	}

	out := make([]backtest.StrategyOutput, len(baseOutputs))
	for i, base := range baseOutputs {
		decision, err := s.applyAt(bars, i, base)
		if err != nil {
			return nil, err
		}
		out[i] = decision
	}
	return out, nil
}

func (s *MLMetaLabelStrategy) EvaluateLatest(ctx context.Context, bars []models.Bar) (backtest.StrategyOutput, error) {
	if len(bars) == 0 {
		return backtest.StrategyOutput{Signal: models.SignalHold, Engine: MetaLabelEngineName}, fmt.Errorf("ML meta-label strategy requires bars")
	}
	if s == nil || s.BaseStrategy == nil {
		return backtest.StrategyOutput{Signal: models.SignalHold, Engine: MetaLabelEngineName}, fmt.Errorf("ML meta-label strategy requires a base strategy")
	}
	base, err := s.BaseStrategy.EvaluateLatest(ctx, bars)
	if err != nil {
		return backtest.StrategyOutput{}, err
	}
	return s.applyAt(bars, len(bars)-1, base)
}

func (s *MLMetaLabelStrategy) applyAt(bars []models.Bar, index int, base backtest.StrategyOutput) (backtest.StrategyOutput, error) {
	thresholds := s.thresholds()
	switch base.Signal {
	case models.SignalBuy:
		return s.applyBuy(bars, index, base, thresholds)
	case models.SignalSell:
		if thresholds.PassThroughExits {
			return s.decorateBaseDecision(base, "base_exit"), nil
		}
		return backtest.StrategyOutput{
			Signal:      models.SignalHold,
			RegimeLabel: base.RegimeLabel,
			Engine:      MetaLabelEngineName,
			Metadata:    withMeta(base.Metadata, "ml_action", "exit_blocked"),
		}, nil
	default:
		return backtest.StrategyOutput{
			Signal:       models.SignalHold,
			RegimeLabel:  base.RegimeLabel,
			AlphaScore:   0,
			Confidence:   0,
			TargetWeight: 0,
			Engine:       MetaLabelEngineName,
			Metadata:     withMeta(base.Metadata, "ml_action", "base_hold"),
		}, nil
	}
}

func (s *MLMetaLabelStrategy) applyBuy(bars []models.Bar, index int, base backtest.StrategyOutput, thresholds MLThresholds) (backtest.StrategyOutput, error) {
	predictor := s.Predictor
	if predictor == nil {
		if thresholds.FailOpenOnError {
			return s.decorateBaseDecision(base, "predictor_missing_fail_open"), nil
		}
		return backtest.StrategyOutput{}, fmt.Errorf("ML meta-label strategy requires a predictor")
	}

	builder := s.FeatureBuilder
	if builder == nil {
		builder = NewFeatureBuilder(DefaultFeatureSpec())
	}
	vector, err := builder.BuildAt(FeatureBuildInput{
		Symbol:                 s.Symbol,
		Bars:                   bars,
		BaseOutput:             &base,
		ContextBars:            contextBarsWithPrimary(s.ContextBars, s.Symbol, bars),
		SectorSymbol:           s.SectorSymbol,
		HMMRegimeProbabilities: s.HMMRegimeProbabilities,
	}, index)
	if err != nil {
		if thresholds.FailOpenOnError {
			return s.decorateBaseDecision(base, "feature_error_fail_open"), nil
		}
		return backtest.StrategyOutput{}, err
	}

	rawProbability, err := predictor.PredictProba(vector.Values)
	if err != nil {
		if thresholds.FailOpenOnError {
			return s.decorateBaseDecision(base, "predict_error_fail_open"), nil
		}
		return backtest.StrategyOutput{}, err
	}
	probability := s.Calibration.Apply(rawProbability)

	metadata := cloneMetadata(base.Metadata)
	metadata["ml_action"] = "entry_filter"
	metadata["p_success"] = probability
	metadata["p_success_raw"] = rawProbability
	metadata["calibration_method"] = s.Calibration.Method
	metadata["model_version"] = predictor.Version()
	metadata["feature_spec_version"] = builder.Spec().Version
	metadata["base_signal"] = int(base.Signal)
	metadata["base_engine"] = base.Engine
	metadata["base_target_weight"] = baseTargetWeight(base)

	if probability < thresholds.EnterLong {
		metadata["ml_decision"] = "veto_buy"
		return backtest.StrategyOutput{
			Signal:       models.SignalHold,
			RegimeLabel:  base.RegimeLabel,
			AlphaScore:   ProbToZScore(probability),
			Confidence:   probability,
			TargetWeight: 0,
			Engine:       MetaLabelEngineName,
			Metadata:     metadata,
		}, nil
	}

	maxWeight := s.effectiveMaxWeight(base)
	size := ProbabilityToBetSizeAboveThreshold(probability, thresholds.EnterLong, maxWeight, thresholds.BetSizingSlope)
	if size <= 0 {
		size = math.Min(maxWeight, baseTargetWeight(base))
	}
	metadata["ml_decision"] = "accept_buy"
	return backtest.StrategyOutput{
		Signal:          models.SignalBuy,
		PositionSizePct: size,
		RegimeLabel:     base.RegimeLabel,
		AlphaScore:      ProbToZScore(probability),
		Confidence:      probability,
		TargetWeight:    size,
		Engine:          MetaLabelEngineName,
		Metadata:        metadata,
	}, nil
}

func contextBarsWithPrimary(contextBars map[string][]models.Bar, symbol string, bars []models.Bar) map[string][]models.Bar {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		if len(bars) == 0 {
			return contextBars
		}
		symbol = strings.ToUpper(strings.TrimSpace(bars[len(bars)-1].Symbol))
	}
	if symbol == "" {
		return contextBars
	}
	if existing, ok := contextBars[symbol]; ok && len(existing) > 0 {
		return contextBars
	}
	out := make(map[string][]models.Bar, len(contextBars)+1)
	for key, value := range contextBars {
		out[strings.ToUpper(strings.TrimSpace(key))] = value
	}
	out[symbol] = bars
	return out
}

func (s *MLMetaLabelStrategy) decorateBaseDecision(base backtest.StrategyOutput, action string) backtest.StrategyOutput {
	baseEngine := base.Engine
	base.Engine = MetaLabelEngineName
	base.Metadata = withMeta(base.Metadata, "ml_action", action)
	if base.Metadata != nil {
		base.Metadata["base_engine"] = baseEngine
	}
	return base
}

func (s *MLMetaLabelStrategy) thresholds() MLThresholds {
	thresholds := s.Thresholds
	defaults := DefaultMLThresholds()
	if thresholds.EnterLong <= 0 {
		thresholds.EnterLong = defaults.EnterLong
	}
	if thresholds.Reduce <= 0 {
		thresholds.Reduce = defaults.Reduce
	}
	if thresholds.BetSizingSlope <= 0 {
		thresholds.BetSizingSlope = defaults.BetSizingSlope
	}
	if !thresholds.PassThroughExits {
		thresholds.PassThroughExits = defaults.PassThroughExits
	}
	return thresholds
}

func (s *MLMetaLabelStrategy) effectiveMaxWeight(base backtest.StrategyOutput) float64 {
	baseWeight := baseTargetWeight(base)
	maxWeight := s.MaxWeight
	if maxWeight <= 0 {
		maxWeight = baseWeight
	}
	if maxWeight <= 0 {
		maxWeight = 0.10
	}
	if baseWeight > 0 {
		return math.Min(maxWeight, baseWeight)
	}
	return maxWeight
}

func baseTargetWeight(output backtest.StrategyOutput) float64 {
	if output.TargetWeight > 0 {
		return output.TargetWeight
	}
	if output.PositionSizePct > 0 {
		if output.PositionSizePct > 1 {
			return 1
		}
		return output.PositionSizePct
	}
	return 0.10
}

func cloneMetadata(metadata map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(metadata)+4)
	for k, v := range metadata {
		out[k] = v
	}
	return out
}

func withMeta(metadata map[string]interface{}, key string, value interface{}) map[string]interface{} {
	out := cloneMetadata(metadata)
	out[key] = value
	return out
}
