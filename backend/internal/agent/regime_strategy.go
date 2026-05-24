package agent

import (
	"context"

	"github.com/oalpha/pkg/models"
)

// RegimeType represents different market regimes
type RegimeType string

const (
	RegimeBull   RegimeType = "bull"
	RegimeBear   RegimeType = "bear"
	RegimeVolatile RegimeType = "volatile"
	RegimeNeutral RegimeType = "neutral"
)

// RegimeSignal extends the basic Signal with regime information
type RegimeSignal struct {
	Signal  models.Signal
	Regime  RegimeType
}

// RegimeDetectorStrategy implements a simple regime detection strategy
// based on price momentum and volatility
type RegimeDetectorStrategy struct {
	fastMA  int // Fast moving average period
	slowMA  int // Slow moving average period
	volLookback int // Volatility lookback period
	volThreshold float64 // Volatility threshold for regime classification
}

// NewRegimeDetectorStrategy creates a new regime detection strategy
func NewRegimeDetectorStrategy(fastMA, slowMA, volLookback int, volThreshold float64) *RegimeDetectorStrategy {
	if fastMA <= 0 || slowMA <= 0 || fastMA >= slowMA {
		return &RegimeDetectorStrategy{
			fastMA:  10,
			slowMA:  30,
			volLookback: 20,
			volThreshold: 0.02, // 2% daily volatility threshold
		}
	}
	return &RegimeDetectorStrategy{
		fastMA:     fastMA,
		slowMA:     slowMA,
		volLookback: volLookback,
		volThreshold: volThreshold,
	}
}

// GenerateSignal implements the Strategy interface
// It detects market regime and generates appropriate signals
func (s *RegimeDetectorStrategy) GenerateSignal(ctx context.Context, bars []models.Bar) ([]models.Signal, error) {
	if len(bars) < s.slowMA {
		// Not enough data to compute signals
		return make([]models.Signal, len(bars)), nil
	}

	// Calculate moving averages
	fastMAValues := calculateSMA(bars, s.fastMA)
	slowMAValues := calculateSMA(bars, s.slowMA)

	// Calculate volatility (using ATR-like measure)
	volatility := calculateVolatility(bars, s.volLookback)

	signals := make([]models.Signal, len(bars))

	for i := range bars {
		// Default to hold
		var signal models.Signal = models.SignalHold

		if i >= s.slowMA-1 { // We have enough data for both MAs
			fastMA := fastMAValues[i]
			slowMA := slowMAValues[i]
			vol := volatility[i]

			// Determine regime
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

			// Generate signal based on regime
			switch regime {
			case RegimeBull:
				if fastMA > slowMA && fastMAValues[i-1] <= slowMAValues[i-1] {
					// Golden cross - bullish signal
					signal = models.SignalBuy
				}
			case RegimeBear:
				if fastMA < slowMA && fastMAValues[i-1] >= slowMAValues[i-1] {
					// Death cross - bearish signal
					signal = models.SignalSell
				}
			case RegimeVolatile:
				// In volatile regime, reduce position sizes or stay neutral
				signal = models.SignalHold
			case RegimeNeutral:
				// In neutral regime, follow the trend if any
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

// calculateSMA computes simple moving average
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

// calculateVolatility computes a simple volatility measure (average true range proxy)
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
		// Normalize by price to get percentage volatility
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