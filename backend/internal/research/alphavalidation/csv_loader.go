package alphavalidation

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/oalpha/pkg/models"
)

func LoadBarsCSV(path string, symbols []string, window ValidationWindow) (map[string][]models.Bar, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open bars csv: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}
	indexes := make(map[string]int, len(header))
	for i, name := range header {
		indexes[strings.ToLower(strings.TrimSpace(name))] = i
	}
	required := []string{"time", "symbol", "open", "high", "low", "close", "volume"}
	for _, field := range required {
		if _, ok := indexes[field]; !ok {
			return nil, fmt.Errorf("bars csv missing required column %q", field)
		}
	}

	allowed := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		allowed[strings.ToUpper(strings.TrimSpace(symbol))] = struct{}{}
	}

	out := make(map[string][]models.Bar, len(allowed))
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read bars csv row: %w", err)
		}
		symbol := strings.ToUpper(strings.TrimSpace(record[indexes["symbol"]]))
		if _, ok := allowed[symbol]; !ok {
			continue
		}
		ts, err := parseCSVTime(record[indexes["time"]])
		if err != nil {
			return nil, fmt.Errorf("parse csv time for %s: %w", symbol, err)
		}
		if ts.Before(window.From) || ts.After(window.To) {
			continue
		}
		openPrice, err := strconv.ParseFloat(strings.TrimSpace(record[indexes["open"]]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse open for %s: %w", symbol, err)
		}
		highPrice, err := strconv.ParseFloat(strings.TrimSpace(record[indexes["high"]]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse high for %s: %w", symbol, err)
		}
		lowPrice, err := strconv.ParseFloat(strings.TrimSpace(record[indexes["low"]]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse low for %s: %w", symbol, err)
		}
		closePrice, err := strconv.ParseFloat(strings.TrimSpace(record[indexes["close"]]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse close for %s: %w", symbol, err)
		}
		volume, err := strconv.ParseInt(strings.TrimSpace(record[indexes["volume"]]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse volume for %s: %w", symbol, err)
		}
		out[symbol] = append(out[symbol], models.Bar{
			Time:   ts,
			Symbol: symbol,
			Open:   openPrice,
			High:   highPrice,
			Low:    lowPrice,
			Close:  closePrice,
			Volume: volume,
		})
	}
	for symbol := range out {
		sort.Slice(out[symbol], func(i, j int) bool { return out[symbol][i].Time.Before(out[symbol][j].Time) })
	}
	return out, nil
}

func parseCSVTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	layouts := []string{time.RFC3339, "2006-01-02", "2006-01-02 15:04:05", "2006-01-02 15:04:05-07:00"}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format %q", value)
}
