package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/oalpha/internal/agent/portfolio"
	"github.com/oalpha/internal/alpaca"
	"github.com/oalpha/internal/config"
	"github.com/oalpha/internal/db"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var (
		limit      = flag.Int("limit", 100, "maximum number of active portfolio agent runs to process")
		timeout    = flag.Duration("timeout", 25*time.Minute, "overall worker timeout")
		runTimeout = flag.Duration("run-timeout", 5*time.Minute, "timeout for each individual portfolio run")
	)
	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	sqlDB, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("open database")
	}
	defer sqlDB.Close()

	barsRepo := db.NewBarsRepository(sqlDB)
	agentRepo := db.NewAgentRepository(sqlDB)
	portfolioRepo := db.NewPortfolioRepository(sqlDB)

	unlock, locked, err := agentRepo.TryPortfolioWorkerLock(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("acquire portfolio worker lock")
	}
	if !locked {
		log.Info().Msg("another portfolio worker is already running; exiting")
		return
	}
	defer unlock(context.Background())

	runs, err := agentRepo.ListActivePortfolioAgentRuns(ctx, *limit)
	if err != nil {
		log.Fatal().Err(err).Msg("list active portfolio runs")
	}
	if len(runs) == 0 {
		log.Info().Msg("no active portfolio runs to process")
		return
	}

	alpacaDataClient := alpaca.NewClient(cfg.AlpacaDataURL, cfg.AlpacaAPIKey, cfg.AlpacaAPISecret)
	alpacaTradingClient := alpaca.NewClient(cfg.AlpacaTradingURL, cfg.AlpacaAPIKey, cfg.AlpacaAPISecret)
	manager := portfolio.NewPortfolioAgentManager(barsRepo, alpacaDataClient)
	orchestrator := portfolio.NewPortfolioOrchestrator(
		manager,
		barsRepo,
		agentRepo,
		portfolioRepo,
		alpacaDataClient,
		alpacaTradingClient,
		cfg.PortfolioExecMode,
		portfolio.DefaultStrategyCatalogConfig(),
	)

	var processed, failed int
	for _, run := range runs {
		runCtx, runCancel := context.WithTimeout(ctx, *runTimeout)
		result, err := orchestrator.RunOnceForAgentRun(runCtx, run)
		runCancel()
		if err != nil {
			failed++
			log.Error().Err(err).Int64("run_id", run.ID).Int64("user_id", run.UserID).Msg("portfolio run failed")
			continue
		}
		processed++
		log.Info().
			Int64("run_id", result.RunID).
			Int64("user_id", result.UserID).
			Str("strategy_key", result.StrategyKey).
			Int("symbols", len(result.Symbols)).
			Time("last_rebalance_at", result.LastRebalanceAt).
			Msg("portfolio run processed")
	}

	log.Info().Int("processed", processed).Int("failed", failed).Msg("portfolio worker completed")
	if failed > 0 {
		os.Exit(1)
	}
}
