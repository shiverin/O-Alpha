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

// NewKalmanStrategy creates a Kalman mean-reversion strategy.
func NewKalmanStrategy(qNoise, rNoise float64, lookback int, zThresh float64) *KalmanStrategy {
	return &KalmanStrategy{
		kf:             NewKalmanFilter1D(qNoise, rNoise),
		lookbackWindow: lookback,
		zThreshold:     zThresh,
		residuals:      make([]float64, 0, lookback),
	}
}

// GenerateSignal emits buy/sell signals when residuals breach the configured z-score threshold.
func (s *KalmanStrategy) GenerateSignal(ctx context.Context, bars []models.Bar) ([]models.Signal, error) {
	out := make([]models.Signal, len(bars))
	if len(bars) == 0 {
		return out, nil
	}

	for i, bar := range bars {
		currentEstimate := s.kf.Update(bar.Close)
		currentResidual := bar.Close - currentEstimate

		if len(s.residuals) >= s.lookbackWindow {
			s.residuals = s.residuals[1:]
		}
		s.residuals = append(s.residuals, currentResidual)

		if len(s.residuals) < s.lookbackWindow {
			out[i] = models.SignalHold
			continue
		}

		_, stdDev := meanStd(s.residuals)

		if stdDev == 0 {
			out[i] = models.SignalHold
			continue
		}

		zScore := currentResidual / stdDev

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
