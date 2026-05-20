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

	tickInterval := time.Hour
	if d, err := time.ParseDuration(cfg.IngestInterval); err == nil && d > 0 {
		tickInterval = d
	} else {
		log.Warn().Err(err).Str("value", cfg.IngestInterval).Msg("invalid INGEST_INTERVAL, defaulting to 1h")
	}

	if err := db.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		log.Fatal().Err(err).Msg("run migrations")
	}

	sqlDB, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("open database")
	}
	defer sqlDB.Close()

	repo := db.NewRepository(sqlDB)
	client := alpaca.NewClient(cfg.AlpacaDataURL, cfg.AlpacaAPIKey, cfg.AlpacaAPISecret)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Backfill historical data once at startup.
	end := time.Now().UTC()
	start := end.Add(-cfg.IngestLookback)
	if err := ingestAll(ctx, repo, client, cfg.IngestSymbols, cfg.IngestInterval, start, end); err != nil {
		log.Error().Err(err).Msg("initial backfill failed")
	}

	// Periodic ingest per symbol.
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				windowEnd := time.Now().UTC()
				windowStart := windowEnd.Add(-2 * tickInterval)
				if err := ingestAll(ctx, repo, client, cfg.IngestSymbols, cfg.IngestInterval, windowStart, windowEnd); err != nil {
					log.Error().Err(err).Msg("periodic ingest failed")
				}
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("ingest service stopped")
}

func ingestAll(ctx context.Context, repo *db.Repository, client *alpaca.Client, symbols []string, interval string, start, end time.Time) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(symbols))

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			if err := ingestSymbol(ctx, repo, client, sym, interval, start, end); err != nil {
				errCh <- err
				log.Error().Err(err).Str("symbol", sym).Msg("ingest failed")
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

func ingestSymbol(ctx context.Context, repo *db.Repository, client *alpaca.Client, symbol, interval string, start, end time.Time) error {
	bars, err := client.GetBars(ctx, symbol, interval, start, end, 10000)
	if err != nil {
		return err
	}
	n, err := repo.InsertBars(ctx, bars)
	if err != nil {
		return err
	}
	log.Info().Str("symbol", symbol).Int("fetched", len(bars)).Int64("upserted", n).Msg("ingested bars")
	return nil
}
