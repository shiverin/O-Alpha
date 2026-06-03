package ml

import (
	"fmt"
	"math"

	"github.com/dmitryikh/leaves"
)

type Predictor interface {
	PredictProba(features []float64) (float64, error)
	Version() string
}

type LeavesPredictor struct {
	modelPath          string
	version            string
	featureSpec        FeatureSpec
	model              *leaves.Ensemble
	loadTransformation bool
}

type ParitySample struct {
	Features            []float64 `json:"features"`
	ExpectedProbability float64   `json:"expected_probability"`
}

func NewLeavesPredictor(modelPath string, featureSpec FeatureSpec, version string) (*LeavesPredictor, error) {
	return newLeavesPredictor(modelPath, featureSpec, version, true)
}

func NewRawLeavesPredictor(modelPath string, featureSpec FeatureSpec, version string) (*LeavesPredictor, error) {
	return newLeavesPredictor(modelPath, featureSpec, version, false)
}

func newLeavesPredictor(modelPath string, featureSpec FeatureSpec, version string, loadTransformation bool) (*LeavesPredictor, error) {
	if modelPath == "" {
		return nil, fmt.Errorf("model path is required")
	}
	if len(featureSpec.Features) == 0 {
		featureSpec = DefaultFeatureSpec()
	}
	model, err := leaves.LGEnsembleFromFile(modelPath, loadTransformation)
	if err != nil {
		return nil, fmt.Errorf("load LightGBM model with leaves: %w", err)
	}
	if model.NFeatures() > 0 && len(featureSpec.Features) != model.NFeatures() {
		return nil, fmt.Errorf("feature spec has %d features but model expects %d", len(featureSpec.Features), model.NFeatures())
	}
	if version == "" {
		version = modelPath
	}
	return &LeavesPredictor{
		modelPath:          modelPath,
		version:            version,
		featureSpec:        featureSpec,
		model:              model,
		loadTransformation: loadTransformation,
	}, nil
}

func (p *LeavesPredictor) PredictProba(features []float64) (float64, error) {
	if p == nil || p.model == nil {
		return 0, fmt.Errorf("leaves predictor is not initialized")
	}
	if !p.loadTransformation {
		return 0, fmt.Errorf("probability prediction is unavailable for raw-score predictor")
	}
	score, err := p.predictWithModel(p.model, features)
	if err != nil {
		return 0, err
	}
	return clampProbability(score), nil
}

func (p *LeavesPredictor) PredictRaw(features []float64) (float64, error) {
	if p == nil || p.model == nil {
		return 0, fmt.Errorf("leaves predictor is not initialized")
	}
	return p.predictWithModel(p.model.EnsembleWithRawPredictions(), features)
}

func (p *LeavesPredictor) predictWithModel(model *leaves.Ensemble, features []float64) (float64, error) {
	if p == nil || p.model == nil {
		return 0, fmt.Errorf("leaves predictor is not initialized")
	}
	if model == nil {
		return 0, fmt.Errorf("leaves model is not initialized")
	}
	if model.NFeatures() > 0 && len(features) != model.NFeatures() {
		return 0, fmt.Errorf("feature vector has %d features but model expects %d", len(features), model.NFeatures())
	}
	for i, value := range features {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return 0, fmt.Errorf("feature %d is not finite", i)
		}
	}

	predictions := make([]float64, model.NOutputGroups())
	if len(predictions) == 0 {
		return 0, fmt.Errorf("model has no output groups")
	}
	if err := model.PredictDense(features, 1, len(features), predictions, 0, 1); err != nil {
		return 0, fmt.Errorf("predict with leaves: %w", err)
	}
	score := predictions[0]
	if math.IsNaN(score) || math.IsInf(score, 0) {
		return 0, fmt.Errorf("model returned non-finite score")
	}
	return score, nil
}

func (p *LeavesPredictor) Version() string {
	if p == nil {
		return ""
	}
	return p.version
}

func (p *LeavesPredictor) FeatureSpec() FeatureSpec {
	if p == nil {
		return DefaultFeatureSpec()
	}
	return p.featureSpec
}

func ValidatePredictorParity(predictor Predictor, samples []ParitySample, tolerance float64) (float64, error) {
	if predictor == nil {
		return 0, fmt.Errorf("predictor is required")
	}
	expected := make([]float64, len(samples))
	actual := make([]float64, len(samples))
	for i, sample := range samples {
		p, err := predictor.PredictProba(sample.Features)
		if err != nil {
			return 0, fmt.Errorf("predict parity sample %d: %w", i, err)
		}
		expected[i] = sample.ExpectedProbability
		actual[i] = p
	}
	return ValidateParity(expected, actual, tolerance)
}
