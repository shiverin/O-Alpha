package panelload

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oalpha/internal/backtest"
	"github.com/oalpha/pkg/models"
)

// LoadPanelFromCSV loads exported OHLCV bars using the same inner-join
// semantics as the alpha-validation CLI.
func LoadPanelFromCSV(path string, symbols []string, timeframe string, start, end time.Time) (backtest.AlignedBars, error) {
	file, err := os.Open(path)
	if err != nil {
		return backtest.AlignedBars{}, fmt.Errorf("open bars csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return backtest.AlignedBars{}, fmt.Errorf("read bars csv header: %w", err)
	}
	columns := headerIndex(header)
	required := []string{"symbol", "time", "open", "high", "low", "close", "volume"}
	for _, name := range required {
		if _, ok := columns[name]; !ok {
			return backtest.AlignedBars{}, fmt.Errorf("bars csv missing %q column", name)
		}
	}

	grouped := make(map[string][]models.Bar, len(symbols))
	wanted := make(map[string]bool, len(symbols))
	for _, symbol := range symbols {
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" {
			continue
		}
		wanted[symbol] = true
		grouped[symbol] = nil
	}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return backtest.AlignedBars{}, fmt.Errorf("read bars csv row: %w", err)
		}
		symbol := strings.ToUpper(strings.TrimSpace(csvValue(record, columns["symbol"])))
		if !wanted[symbol] {
			continue
		}
		timestamp, err := parseCSVTime(csvValue(record, columns["time"]))
		if err != nil {
			return backtest.AlignedBars{}, fmt.Errorf("parse time for %s: %w", symbol, err)
		}
		if timestamp.Before(start) || timestamp.After(end) {
			continue
		}
		open, err := parseCSVFloat(record, columns["open"], "open")
		if err != nil {
			return backtest.AlignedBars{}, err
		}
		high, err := parseCSVFloat(record, columns["high"], "high")
		if err != nil {
			return backtest.AlignedBars{}, err
		}
		low, err := parseCSVFloat(record, columns["low"], "low")
		if err != nil {
			return backtest.AlignedBars{}, err
		}
		closePrice, err := parseCSVFloat(record, columns["close"], "close")
		if err != nil {
			return backtest.AlignedBars{}, err
		}
		volume, err := parseCSVInt(record, columns["volume"], "volume")
		if err != nil {
			return backtest.AlignedBars{}, err
		}
		grouped[symbol] = append(grouped[symbol], models.Bar{
			Time:   timestamp,
			Symbol: symbol,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: volume,
		})
	}
	for symbol := range grouped {
		sort.Slice(grouped[symbol], func(i, j int) bool {
			return grouped[symbol][i].Time.Before(grouped[symbol][j].Time)
		})
	}
	panel, err := backtest.AlignBars(grouped, backtest.AlignmentConfig{
		Mode:       backtest.AlignInnerJoin,
		Timeframe:  timeframe,
		Feed:       "csv",
		Adjustment: "adjusted",
	})
	if err != nil {
		return backtest.AlignedBars{}, err
	}
	panel.Metadata["source"] = "csv"
	panel.Metadata["csv_path"] = path
	return panel, nil
}

func ParseDate(value string, endOfDay bool) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date %q: %w", value, err)
	}
	parsed = parsed.UTC()
	if endOfDay {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return parsed, nil
}

func ParseCSVSymbols(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, part := range parts {
		normalized := strings.ToUpper(strings.TrimSpace(part))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	return out
}

func OrderSymbols(aligned []string, requested []string) []string {
	available := make(map[string]bool, len(aligned))
	for _, symbol := range aligned {
		available[symbol] = true
	}
	out := make([]string, 0, len(aligned))
	seen := make(map[string]bool, len(aligned))
	for _, symbol := range requested {
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" || seen[symbol] || !available[symbol] {
			continue
		}
		out = append(out, symbol)
		seen[symbol] = true
	}
	for _, symbol := range aligned {
		if !seen[symbol] {
			out = append(out, symbol)
		}
	}
	return out
}

func headerIndex(header []string) map[string]int {
	out := make(map[string]int, len(header))
	for index, name := range header {
		out[strings.ToLower(strings.TrimSpace(name))] = index
	}
	return out
}

func csvValue(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func parseCSVTime(value string) (time.Time, error) {
	timestamp, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(value))
	if err == nil {
		return timestamp.UTC(), nil
	}
	timestamp, err = time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err == nil {
		return timestamp.UTC(), nil
	}
	timestamp, err = time.Parse("2006-01-02", strings.TrimSpace(value))
	if err == nil {
		return timestamp.UTC(), nil
	}
	return time.Time{}, err
}

func parseCSVFloat(record []string, index int, name string) (float64, error) {
	value := csvValue(record, index)
	out, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s %q: %w", name, value, err)
	}
	return out, nil
}

func parseCSVInt(record []string, index int, name string) (int64, error) {
	value := csvValue(record, index)
	out, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		return out, nil
	}
	asFloat, floatErr := strconv.ParseFloat(value, 64)
	if floatErr != nil {
		return 0, fmt.Errorf("parse %s %q: %w", name, value, err)
	}
	return int64(asFloat), nil
}
