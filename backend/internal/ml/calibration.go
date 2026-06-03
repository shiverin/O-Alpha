package ml

import (
	"fmt"
	"math"
)

type CalibrationBin struct {
	Lower      float64 `json:"lower"`
	Upper      float64 `json:"upper"`
	Calibrated float64 `json:"calibrated"`
}

type CalibrationModel struct {
	Method    string           `json:"method"`
	Intercept float64          `json:"intercept,omitempty"`
	Slope     float64          `json:"slope,omitempty"`
	Bins      []CalibrationBin `json:"bins,omitempty"`
}

type MLThresholds struct {
	EnterLong        float64 `json:"enter_long"`
	Reduce           float64 `json:"reduce"`
	BetSizingSlope   float64 `json:"bet_sizing_slope"`
	FailOpenOnError  bool    `json:"fail_open_on_error"`
	PassThroughExits bool    `json:"pass_through_exits"`
}

func DefaultMLThresholds() MLThresholds {
	return MLThresholds{
		EnterLong:        0.50,
		Reduce:           0.45,
		BetSizingSlope:   2.0,
		PassThroughExits: true,
	}
}

func (m CalibrationModel) Apply(probability float64) float64 {
	p := clampProbability(probability)
	switch m.Method {
	case "", "none":
		return p
	case "platt", "logistic":
		logit := m.Intercept + m.Slope*safeLogit(p)
		return clampProbability(1 / (1 + math.Exp(-logit)))
	case "histogram", "isotonic":
		for _, bin := range m.Bins {
			if p >= bin.Lower && p <= bin.Upper {
				return clampProbability(bin.Calibrated)
			}
		}
		return p
	default:
		return p
	}
}

func ProbabilityToBetSize(probability, maxWeight, slope float64) float64 {
	return ProbabilityToBetSizeAboveThreshold(probability, 0.5, maxWeight, slope)
}

func ProbabilityToBetSizeAboveThreshold(probability, threshold, maxWeight, slope float64) float64 {
	if maxWeight <= 0 {
		return 0
	}
	if slope <= 0 {
		slope = 2
	}
	threshold = clampProbability(threshold)
	edge := math.Max(0, clampProbability(probability)-threshold)
	return math.Min(maxWeight, slope*edge)
}

func ProbToZScore(probability float64) float64 {
	p := clampProbability(probability)
	z := safeLogit(p) / 1.702
	if z > 3 {
		return 3
	}
	if z < -3 {
		return -3
	}
	return z
}

func ValidateParity(expected, actual []float64, tolerance float64) (float64, error) {
	if len(expected) != len(actual) {
		return 0, fmt.Errorf("parity vectors differ in length: expected %d, actual %d", len(expected), len(actual))
	}
	if tolerance <= 0 {
		tolerance = 1e-6
	}
	maxAbsError := 0.0
	for i := range expected {
		err := math.Abs(expected[i] - actual[i])
		if err > maxAbsError {
			maxAbsError = err
		}
	}
	if maxAbsError > tolerance {
		return maxAbsError, fmt.Errorf("prediction parity max abs error %.12f exceeds tolerance %.12f", maxAbsError, tolerance)
	}
	return maxAbsError, nil
}

func clampProbability(probability float64) float64 {
	if math.IsNaN(probability) || math.IsInf(probability, 0) {
		return 0.5
	}
	if probability < 1e-9 {
		return 1e-9
	}
	if probability > 1-1e-9 {
		return 1 - 1e-9
	}
	return probability
}

func safeLogit(probability float64) float64 {
	p := clampProbability(probability)
	return math.Log(p / (1 - p))
}
