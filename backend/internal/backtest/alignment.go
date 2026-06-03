package backtest

import (
	"fmt"
	"sort"
	"time"

	"github.com/oalpha/pkg/models"
)

type AlignMode string

const (
	AlignInnerJoin   AlignMode = "inner_join"
	AlignForwardFill AlignMode = "forward_fill"
)

type AlignmentConfig struct {
	Mode         AlignMode
	MaxStaleBars int
	Timeframe    string
	Feed         string
	Adjustment   string
}

func AlignBars(barsBySymbol map[string][]models.Bar, cfg AlignmentConfig) (AlignedBars, error) {
	if len(barsBySymbol) == 0 {
		return AlignedBars{}, fmt.Errorf("at least one symbol is required")
	}
	if cfg.Mode == "" {
		cfg.Mode = AlignInnerJoin
	}

	symbols := make([]string, 0, len(barsBySymbol))
	for symbol, bars := range barsBySymbol {
		if symbol == "" {
			return AlignedBars{}, fmt.Errorf("symbol cannot be empty")
		}
		if len(bars) == 0 {
			return AlignedBars{}, fmt.Errorf("symbol %s has no bars", symbol)
		}
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	switch cfg.Mode {
	case AlignInnerJoin:
		return alignInnerJoin(symbols, barsBySymbol, cfg)
	case AlignForwardFill:
		return alignForwardFill(symbols, barsBySymbol, cfg)
	default:
		return AlignedBars{}, fmt.Errorf("unsupported align mode: %s", cfg.Mode)
	}
}

func alignInnerJoin(symbols []string, barsBySymbol map[string][]models.Bar, cfg AlignmentConfig) (AlignedBars, error) {
	counts := make(map[time.Time]int)
	bySymbolTime := make(map[string]map[time.Time]models.Bar, len(symbols))
	for _, symbol := range symbols {
		byTime := make(map[time.Time]models.Bar, len(barsBySymbol[symbol]))
		for _, bar := range barsBySymbol[symbol] {
			byTime[bar.Time] = bar
		}
		bySymbolTime[symbol] = byTime
		for t := range byTime {
			counts[t]++
		}
	}

	times := make([]time.Time, 0)
	for t, count := range counts {
		if count == len(symbols) {
			times = append(times, t)
		}
	}
	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	if len(times) == 0 {
		return AlignedBars{}, fmt.Errorf("no common timestamps across %d symbols", len(symbols))
	}

	aligned := newAlignedBars(symbols, times, cfg)
	for _, symbol := range symbols {
		aligned.Bars[symbol] = make([]models.Bar, 0, len(times))
		for _, t := range times {
			aligned.Bars[symbol] = append(aligned.Bars[symbol], bySymbolTime[symbol][t])
		}
	}
	return aligned, nil
}

func alignForwardFill(symbols []string, barsBySymbol map[string][]models.Bar, cfg AlignmentConfig) (AlignedBars, error) {
	timeSet := make(map[time.Time]struct{})
	for _, bars := range barsBySymbol {
		for _, bar := range bars {
			timeSet[bar.Time] = struct{}{}
		}
	}
	times := make([]time.Time, 0, len(timeSet))
	for t := range timeSet {
		times = append(times, t)
	}
	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	if len(times) == 0 {
		return AlignedBars{}, fmt.Errorf("no timestamps to align")
	}

	aligned := newAlignedBars(symbols, times, cfg)
	staleCounts := make(map[string]int)
	for _, symbol := range symbols {
		source := append([]models.Bar(nil), barsBySymbol[symbol]...)
		sort.Slice(source, func(i, j int) bool { return source[i].Time.Before(source[j].Time) })
		aligned.Bars[symbol] = make([]models.Bar, 0, len(times))
		var cursor int
		var last models.Bar
		var hasLast bool
		for _, t := range times {
			for cursor < len(source) && !source[cursor].Time.After(t) {
				last = source[cursor]
				hasLast = true
				cursor++
			}
			if !hasLast {
				return AlignedBars{}, fmt.Errorf("cannot forward-fill %s before first bar", symbol)
			}
			bar := last
			if !bar.Time.Equal(t) {
				staleCounts[symbol]++
				if cfg.MaxStaleBars > 0 && staleCounts[symbol] > cfg.MaxStaleBars {
					return AlignedBars{}, fmt.Errorf("symbol %s exceeds max stale bars", symbol)
				}
				bar.Time = t
			} else {
				staleCounts[symbol] = 0
			}
			aligned.Bars[symbol] = append(aligned.Bars[symbol], bar)
		}
	}
	aligned.Metadata["stale_counts"] = staleCounts
	return aligned, nil
}

func newAlignedBars(symbols []string, times []time.Time, cfg AlignmentConfig) AlignedBars {
	return AlignedBars{
		Times:      times,
		Symbols:    append([]string(nil), symbols...),
		Bars:       make(map[string][]models.Bar, len(symbols)),
		Timeframe:  cfg.Timeframe,
		Feed:       cfg.Feed,
		Adjustment: cfg.Adjustment,
		Metadata:   make(map[string]interface{}),
	}
}
