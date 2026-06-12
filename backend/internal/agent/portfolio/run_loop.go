package portfolio

import (
	"context"
	"time"
)

func (w *PortfolioAgentWorker) Context() context.Context {
	return w.ctx
}

func (w *PortfolioAgentWorker) LatestPrices() map[string]float64 {
	w.barsMu.RLock()
	defer w.barsMu.RUnlock()

	prices := make(map[string]float64, len(w.bars.Bars))
	for symbol, bars := range w.bars.Bars {
		if len(bars) == 0 {
			continue
		}
		prices[symbol] = bars[len(bars)-1].Close
	}
	return prices
}

func (w *PortfolioAgentWorker) ApplyLatestPrices(prices map[string]float64, asOf time.Time) {
	if w == nil || len(prices) == 0 {
		return
	}
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}

	w.barsMu.Lock()
	defer w.barsMu.Unlock()

	for symbol, price := range prices {
		if price <= 0 {
			continue
		}
		bars := w.bars.Bars[symbol]
		if len(bars) == 0 {
			continue
		}
		latest := &bars[len(bars)-1]
		if asOf.Before(latest.Time) {
			continue
		}
		if price > latest.High {
			latest.High = price
		}
		if price < latest.Low {
			latest.Low = price
		}
		latest.Close = price
		w.bars.Bars[symbol] = bars
	}
}

func (w *PortfolioAgentWorker) HasBars() bool {
	w.barsMu.RLock()
	defer w.barsMu.RUnlock()
	return len(w.bars.Times) > 0
}

func (w *PortfolioAgentWorker) Symbols() []string {
	return append([]string(nil), w.symbols...)
}

func pollIntervalFor(timeframe string) time.Duration {
	switch timeframe {
	case "1Min":
		return time.Minute
	case "5Min":
		return 5 * time.Minute
	case "15Min":
		return 15 * time.Minute
	case "1Hour":
		return 5 * time.Minute
	case "1Day":
		return time.Hour
	default:
		return time.Hour
	}
}

func warmupLookbackFor(timeframe string) time.Duration {
	switch timeframe {
	case "1Min", "5Min", "15Min":
		return 30 * 24 * time.Hour
	case "1Hour":
		return 120 * 24 * time.Hour
	case "1Day":
		// Live catalog strategies need roughly 252-260 trading bars of context.
		// Loading five years for the full default universe makes launch feel stuck
		// against a remote DB, so keep startup to a recent live warmup window.
		return 390 * 24 * time.Hour
	default:
		return 365 * 24 * time.Hour
	}
}

func refreshOverlapFor(timeframe string) time.Duration {
	switch timeframe {
	case "1Min":
		return 10 * time.Minute
	case "5Min":
		return time.Hour
	case "15Min":
		return 3 * time.Hour
	case "1Hour":
		return 6 * time.Hour
	case "1Day":
		return 7 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}
