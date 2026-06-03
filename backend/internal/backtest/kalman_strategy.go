package backtest

import (
	"context"

	"github.com/oalpha/pkg/models"
)

type KalmanStrategy struct {
	kf             *KalmanFilter1D
	lookbackWindow int
	zThreshold     float64
}

// NewKalmanStrategy creates a Kalman mean-reversion strategy.
func NewKalmanStrategy(qNoise, rNoise float64, lookback int, zThresh float64) *KalmanStrategy {
	return &KalmanStrategy{
		kf:             NewKalmanFilter1D(qNoise, rNoise),
		lookbackWindow: lookback,
		zThreshold:     zThresh,
	}
}

// GenerateSignal emits buy/sell signals when residuals breach the configured z-score threshold.
func (s *KalmanStrategy) GenerateSignal(ctx context.Context, bars []models.Bar) ([]models.Signal, error) {
	out := make([]models.Signal, len(bars))
	if len(bars) == 0 {
		return out, nil
	}

	kf := NewKalmanFilter1D(s.kf.ProcessNoise, s.kf.MeasurementNoise)
	residuals := make([]float64, 0, s.lookbackWindow)

	for i, bar := range bars {
		currentEstimate := kf.Update(bar.Close)
		currentResidual := bar.Close - currentEstimate

		if len(residuals) >= s.lookbackWindow {
			residuals = residuals[1:]
		}
		residuals = append(residuals, currentResidual)

		if len(residuals) < s.lookbackWindow {
			out[i] = models.SignalHold
			continue
		}

		_, stdDev := meanStd(residuals)

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

func (s *KalmanStrategy) GenerateSignals(ctx context.Context, bars []models.Bar) ([]StrategyOutput, error) {
	signals, err := s.GenerateSignal(ctx, bars)
	if err != nil {
		return nil, err
	}

	out := make([]StrategyOutput, len(signals))
	for i, signal := range signals {
		out[i] = StrategyOutput{
			Signal:          signal,
			PositionSizePct: 0.10,
			RegimeLabel:     "NORMAL",
			AlphaScore:      signalToSignedScore(signal),
			Confidence:      confidenceFromSignal(signal),
			TargetWeight:    targetWeightFromSignal(signal, 0.10),
			Engine:          "kalman",
		}
	}
	return out, nil
}

func (s *KalmanStrategy) EvaluateLatest(ctx context.Context, bars []models.Bar) (StrategyOutput, error) {
	outputs, err := s.GenerateSignals(ctx, bars)
	if err != nil || len(outputs) == 0 {
		return StrategyOutput{Signal: models.SignalHold}, err
	}
	return outputs[len(outputs)-1], nil
}
