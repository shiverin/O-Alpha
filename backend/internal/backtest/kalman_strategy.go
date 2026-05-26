package backtest

import (
	"context"

	"github.com/oalpha/pkg/models"
)

type KalmanStrategy struct {
	kf             *KalmanFilter1D
	lookbackWindow int
	zThreshold     float64
	residuals      []float64
}

// NewKalmanStrategy instantiates the institutional noise-filtering strategy.
func NewKalmanStrategy(qNoise, rNoise float64, lookback int, zThresh float64) *KalmanStrategy {
	return &KalmanStrategy{
		kf:             NewKalmanFilter1D(qNoise, rNoise),
		lookbackWindow: lookback,
		zThreshold:     zThresh,
		residuals:      make([]float64, 0, lookback),
	}
}

// GenerateSignal processes historical price information and decides entry/exit points.
func (s *KalmanStrategy) GenerateSignal(ctx context.Context, bars []models.Bar) ([]models.Signal, error) {
	out := make([]models.Signal, len(bars))
	if len(bars) == 0 {
		return out, nil
	}

	for i, bar := range bars {
		currentEstimate := s.kf.Update(bar.Close)
		currentResidual := bar.Close - currentEstimate

		// Maintain historical rolling window of residuals
		if len(s.residuals) >= s.lookbackWindow {
			s.residuals = s.residuals[1:] // Pop oldest
		}
		s.residuals = append(s.residuals, currentResidual)

		// Insufficient history to establish stable variance threshold
		if len(s.residuals) < s.lookbackWindow {
			out[i] = models.SignalHold
			continue
		}

		// Use meanStd from metrics.go! No duplicate code needed.
		_, stdDev := meanStd(s.residuals)

		if stdDev == 0 {
			out[i] = models.SignalHold
			continue
		}

		// Compute Z-Score of the current bar's price deviation
		zScore := currentResidual / stdDev

		// Mean-Reversion Triggers
		if zScore < -s.zThreshold {
			out[i] = models.SignalBuy
		} else if zScore > s.zThreshold {
			out[i] = models.SignalSell
		} else {
			out[i] = models.SignalHold
		}
	}

	return out, nil
}