package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/oalpha/internal/alpaca"
	"github.com/oalpha/internal/config"
	"github.com/oalpha/internal/db"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func parseAlpacaTimeframe(tf string) time.Duration {
	switch tf {
	case "1Min":
		return time.Minute
	case "5Min":
		return 5 * time.Minute
	case "15Min":
		return 15 * time.Minute
	case "1Hour":
		return time.Hour
	case "1Day":
		return 24 * time.Hour
	default:
		if d, err := time.ParseDuration(tf); err == nil {
			return d
		}
		return time.Hour
	}
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}
	if len(cfg.IngestSymbols) == 0 {
		log.Fatal().Msg("INGEST_SYMBOLS is required for ingest")
	}
	if cfg.AlpacaAPIKey == "" {
		log.Fatal().Msg("ALPACA_API_KEY is required for ingest")
	}
	if cfg.AlpacaAPISecret == "" {
		log.Fatal().Msg("ALPACA_API_SECRET is required for ingest")
	}

	tickInterval := parseAlpacaTimeframe(cfg.IngestInterval)
	log.Info().Str("interval", cfg.IngestInterval).Str("parsed_duration", tickInterval.String()).Msg("Smart Delta-Sync scheduling engine online")

	if err := db.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		log.Fatal().Err(err).Msg("run migrations")
	}

	sqlDB, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("open database")
	}
	defer sqlDB.Close()

	repo := db.NewBarsRepository(sqlDB)
	client := alpaca.NewClient(cfg.AlpacaDataURL, cfg.AlpacaAPIKey, cfg.AlpacaAPISecret)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := syncAllSymbolsDelta(ctx, repo, client, cfg.IngestSymbols, cfg.IngestInterval, cfg.IngestLookback); err != nil {
		log.Error().Err(err).Msg("startup delta resync encountered warnings")
	}

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Info().Msg("Cron cycle triggered. Evaluating time delta gaps...")
				if err := syncAllSymbolsDelta(ctx, repo, client, cfg.IngestSymbols, cfg.IngestInterval, cfg.IngestLookback); err != nil {
					log.Error().Err(err).Msg("periodic delta resync failed")
				}
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("ingest service stopped gracefully")
}

// syncAllSymbolsDelta runs one delta sync per configured symbol.
func syncAllSymbolsDelta(ctx context.Context, repo *db.BarsRepository, client *alpaca.Client, symbols []string, interval string, defaultLookback time.Duration) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(symbols))

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			if err := syncSingleSymbolDelta(ctx, repo, client, sym, interval, defaultLookback); err != nil {
				errCh <- err
				log.Error().Err(err).Str("symbol", sym).Msg("delta sync task failed")
			}
		}(symbol)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// syncSingleSymbolDelta starts from the latest stored bar and relies on upserts
// to absorb an overlapping boundary candle.
func syncSingleSymbolDelta(ctx context.Context, repo *db.BarsRepository, client *alpaca.Client, symbol, interval string, defaultLookback time.Duration) error {
	end := time.Now().UTC()
	var start time.Time

	latestTime, found, err := repo.GetLatestBarTime(ctx, symbol, interval)
	if err != nil {
		return err
	}

	if found {
		start = latestTime
		log.Info().Str("symbol", symbol).Time("last_recorded", start).Msg("Delta sync mapping initialized")
	} else {
		start = end.Add(-defaultLookback)
		log.Info().Str("symbol", symbol).Time("lookback_start", start).Msg("No data found. Commencing cold structural initialization")
	}

	if end.Sub(start) < 2*time.Second {
		log.Debug().Str("symbol", symbol).Msg("Asset database timeline matches real-world metrics. Sync skipped")
		return nil
	}

	bars, err := client.GetBars(ctx, symbol, interval, start, end, 10000)
	if err != nil {
		return err
	}

	if len(bars) == 0 {
		return nil
	}

	n, err := repo.InsertBars(ctx, bars, interval)
	if err != nil {
		return err
	}

	log.Info().Str("symbol", symbol).Int("fetched", len(bars)).Int64("upserted", n).Msg("Ingestion synchronization successfully finalized")
	return nil
}
