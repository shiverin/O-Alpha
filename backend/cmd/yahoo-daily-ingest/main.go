package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/oalpha/internal/config"
	"github.com/oalpha/internal/db"
	"github.com/oalpha/internal/marketdata"
)

func main() {
	var (
		symbols     = flag.String("symbols", "", "comma-separated symbols to ingest")
		from        = flag.String("from", "2015-01-01", "inclusive start date, YYYY-MM-DD")
		to          = flag.String("to", "", "inclusive end date, YYYY-MM-DD; defaults to now")
		timeframe   = flag.String("timeframe", "1Day", "stored timeframe")
		feed        = flag.String("feed", "yahoo", "stored feed")
		adjustment  = flag.String("adjustment", "adj", "stored adjustment")
		source      = flag.String("source", "yahoo_chart", "stored source")
		sleepMillis = flag.Int("sleep-ms", 250, "sleep between symbols to avoid source throttling")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fatal(err)
	}
	start, end, err := resolveDateRange(*from, *to)
	if err != nil {
		fatal(err)
	}
	symbolList := parseCSV(*symbols)
	if len(symbolList) == 0 {
		fatal(fmt.Errorf("at least one symbol is required"))
	}

	ctx := context.Background()
	pool, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		fatal(err)
	}
	defer pool.Close()
	repo := db.NewBarsRepository(pool)
	client := marketdata.YahooDailyClient{}

	var total int64
	for i, symbol := range symbolList {
		bars, err := client.FetchDailyBars(ctx, symbol, start, end)
		if err != nil {
			fatal(fmt.Errorf("%s: %w", symbol, err))
		}
		n, err := repo.InsertBarsDataset(ctx, bars, *timeframe, *feed, *adjustment, *source)
		if err != nil {
			fatal(fmt.Errorf("%s upsert: %w", symbol, err))
		}
		total += n
		first, last := "", ""
		if len(bars) > 0 {
			first = bars[0].Time.Format(time.RFC3339)
			last = bars[len(bars)-1].Time.Format(time.RFC3339)
		}
		fmt.Printf("%-6s fetched=%4d upserted=%4d first=%s last=%s\n", symbol, len(bars), n, first, last)
		if i < len(symbolList)-1 && *sleepMillis > 0 {
			time.Sleep(time.Duration(*sleepMillis) * time.Millisecond)
		}
	}
	fmt.Printf("done symbols=%d upserted=%d feed=%s adjustment=%s source=%s timeframe=%s\n",
		len(symbolList), total, *feed, *adjustment, *source, *timeframe)
}

func resolveDateRange(from, to string) (time.Time, time.Time, error) {
	start, err := parseDate(from, false)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end := time.Now().UTC()
	if strings.TrimSpace(to) != "" {
		parsed, err := parseDate(to, true)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end = parsed
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("to must be after from")
	}
	return start, end, nil
}

func parseDate(value string, endOfDay bool) (time.Time, error) {
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

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, part := range parts {
		symbol := strings.ToUpper(strings.TrimSpace(part))
		if symbol == "" || seen[symbol] {
			continue
		}
		seen[symbol] = true
		out = append(out, symbol)
	}
	return out
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
