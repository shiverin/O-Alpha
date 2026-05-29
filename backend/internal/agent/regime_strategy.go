package agent

import (
	"context"

	"github.com/oalpha/pkg/models"
)

// RegimeType represents a coarse market regime.
type RegimeType string

const (
	RegimeBull     RegimeType = "bull"
	RegimeBear     RegimeType = "bear"
	RegimeVolatile RegimeType = "volatile"
	RegimeNeutral  RegimeType = "neutral"
)

// RegimeSignal pairs a trading signal with its detected regime.
type RegimeSignal struct {
	Signal models.Signal
	Regime RegimeType
}

// RegimeDetectorStrategy combines moving-average momentum with volatility filtering.
type RegimeDetectorStrategy struct {
	fastMA       int
	slowMA       int
	volLookback  int
	volThreshold float64
}

// NewRegimeDetectorStrategy creates a new regime detection strategy.
func NewRegimeDetectorStrategy(fastMA, slowMA, volLookback int, volThreshold float64) *RegimeDetectorStrategy {
	if fastMA <= 0 || slowMA <= 0 || fastMA >= slowMA {
		return &RegimeDetectorStrategy{
			fastMA:       10,
			slowMA:       30,
			volLookback:  20,
			volThreshold: 0.02,
		}
	}
	return &RegimeDetectorStrategy{
		fastMA:       fastMA,
		slowMA:       slowMA,
		volLookback:  volLookback,
		volThreshold: volThreshold,
	}
}

// GenerateSignal detects the current regime and emits index-aligned signals.
func (s *RegimeDetectorStrategy) GenerateSignal(ctx context.Context, bars []models.Bar) ([]models.Signal, error) {
	if len(bars) < s.slowMA {
		return make([]models.Signal, len(bars)), nil
	}

	fastMAValues := calculateSMA(bars, s.fastMA)
	slowMAValues := calculateSMA(bars, s.slowMA)

	volatility := calculateVolatility(bars, s.volLookback)

	signals := make([]models.Signal, len(bars))

	for i := range bars {
		signal := models.SignalHold

		if i >= s.slowMA-1 {
			fastMA := fastMAValues[i]
			slowMA := slowMAValues[i]
			vol := volatility[i]

			var regime RegimeType
			if vol > s.volThreshold {
				regime = RegimeVolatile
			} else if fastMA > slowMA {
				regime = RegimeBull
			} else if fastMA < slowMA {
				regime = RegimeBear
			} else {
				regime = RegimeNeutral
			}

			switch regime {
			case RegimeBull:
				if fastMA > slowMA && fastMAValues[i-1] <= slowMAValues[i-1] {
					signal = models.SignalBuy
				}
			case RegimeBear:
				if fastMA < slowMA && fastMAValues[i-1] >= slowMAValues[i-1] {
					signal = models.SignalSell
				}
			case RegimeVolatile:
				signal = models.SignalHold
			case RegimeNeutral:
				if fastMA > slowMA {
					signal = models.SignalBuy
				} else if fastMA < slowMA {
					signal = models.SignalSell
				}
			}
		}

		signals[i] = signal
	}

	return signals, nil
}

// calculateSMA computes a simple moving average.
func calculateSMA(bars []models.Bar, period int) []float64 {
	sma := make([]float64, len(bars))
	if period <= 0 || len(bars) < period {
		return sma
	}

	var sum float64
	for i := 0; i < len(bars); i++ {
		sum += bars[i].Close
		if i >= period {
			sum -= bars[i-period].Close
		}
		if i >= period-1 {
			sma[i] = sum / float64(period)
		}
	}
	return sma
}

// calculateVolatility computes an average true range proxy.
func calculateVolatility(bars []models.Bar, lookback int) []float64 {
	vol := make([]float64, len(bars))
	if lookback <= 0 || len(bars) < lookback {
		return vol
	}

	for i := lookback; i < len(bars); i++ {
		var trSum float64
		for j := i - lookback + 1; j <= i; j++ {
			high := bars[j].High
			low := bars[j].Low
			prevClose := bars[j-1].Close

			tr := high - low
			if mathAbs(high-prevClose) > tr {
				tr = mathAbs(high - prevClose)
			}
			if mathAbs(low-prevClose) > tr {
				tr = mathAbs(low - prevClose)
			}
			trSum += tr
		}
		avgTR := trSum / float64(lookback)
		vol[i] = avgTR / bars[i].Close
	}
	return vol
}

func mathAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
