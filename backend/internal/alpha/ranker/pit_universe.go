package ranker

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type ConstituentInterval struct {
	Symbol string `json:"symbol"`
	Start  string `json:"start"`
	End    string `json:"end,omitempty"`
}

type PointInTimeUniverse struct {
	Version   string                `json:"version,omitempty"`
	Source    map[string]string     `json:"source,omitempty"`
	Symbols   []string              `json:"symbols,omitempty"`
	Intervals []ConstituentInterval `json:"intervals"`

	intervalsBySymbol map[string][]parsedConstituentInterval
}

type parsedConstituentInterval struct {
	Start  time.Time
	End    time.Time
	HasEnd bool
}

func LoadPointInTimeUniverse(path string) (*PointInTimeUniverse, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read point-in-time universe: %w", err)
	}
	var universe PointInTimeUniverse
	if err := json.Unmarshal(payload, &universe); err != nil {
		return nil, fmt.Errorf("parse point-in-time universe: %w", err)
	}
	if err := universe.BuildIndex(); err != nil {
		return nil, err
	}
	return &universe, nil
}

func NewPointInTimeUniverse(intervals []ConstituentInterval) (*PointInTimeUniverse, error) {
	universe := &PointInTimeUniverse{
		Version:   "pit_constituents_v1",
		Intervals: append([]ConstituentInterval(nil), intervals...),
	}
	if err := universe.BuildIndex(); err != nil {
		return nil, err
	}
	return universe, nil
}

func (u *PointInTimeUniverse) BuildIndex() error {
	if u == nil {
		return nil
	}
	bySymbol := make(map[string][]parsedConstituentInterval)
	symbolSet := make(map[string]bool)
	for _, interval := range u.Intervals {
		symbol := strings.ToUpper(strings.TrimSpace(interval.Symbol))
		if symbol == "" {
			return fmt.Errorf("point-in-time universe interval missing symbol")
		}
		start, err := parseManifestDate(interval.Start)
		if err != nil {
			return fmt.Errorf("parse start for %s: %w", symbol, err)
		}
		parsed := parsedConstituentInterval{Start: start}
		endRaw := strings.TrimSpace(interval.End)
		if endRaw != "" {
			end, err := parseManifestDate(endRaw)
			if err != nil {
				return fmt.Errorf("parse end for %s: %w", symbol, err)
			}
			if end.Before(start) {
				return fmt.Errorf("point-in-time universe interval for %s ends before it starts", symbol)
			}
			parsed.End = end
			parsed.HasEnd = true
		}
		bySymbol[symbol] = append(bySymbol[symbol], parsed)
		symbolSet[symbol] = true
	}
	for symbol := range bySymbol {
		sort.Slice(bySymbol[symbol], func(i, j int) bool {
			return bySymbol[symbol][i].Start.Before(bySymbol[symbol][j].Start)
		})
	}
	u.intervalsBySymbol = bySymbol
	if len(u.Symbols) == 0 {
		u.Symbols = make([]string, 0, len(symbolSet))
		for symbol := range symbolSet {
			u.Symbols = append(u.Symbols, symbol)
		}
		sort.Strings(u.Symbols)
	} else {
		u.Symbols = normalizeSymbols(u.Symbols)
	}
	return nil
}

func (u *PointInTimeUniverse) Active(symbol string, t time.Time) bool {
	if u == nil {
		return true
	}
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	intervals := u.intervalsBySymbol[symbol]
	if len(intervals) == 0 {
		return false
	}
	day := manifestDay(t)
	for _, interval := range intervals {
		if day.Before(interval.Start) {
			continue
		}
		if interval.HasEnd && day.After(interval.End) {
			continue
		}
		return true
	}
	return false
}

func parseManifestDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return manifestDay(t), nil
	}
	return time.Time{}, fmt.Errorf("unsupported date %q", value)
}

func manifestDay(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}
