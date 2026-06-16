package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oalpha/internal/agent"
	"github.com/oalpha/internal/agent/portfolio"
	"github.com/oalpha/internal/alpaca"
	"github.com/oalpha/internal/api"
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

	if err := db.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		log.Fatal().Err(err).Msg("run migrations")
	}

	sqlDB, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("open database")
	}
	defer sqlDB.Close()

	repo := db.NewBarsRepository(sqlDB)
	agentRepo := db.NewAgentRepository(sqlDB)
	portfolioRepo := db.NewPortfolioRepository(sqlDB)

	alpacaClient := alpaca.NewClient(cfg.AlpacaDataURL, cfg.AlpacaAPIKey, cfg.AlpacaAPISecret)
	alpacaTradingClient := alpaca.NewClient(cfg.AlpacaTradingURL, cfg.AlpacaAPIKey, cfg.AlpacaAPISecret)

	agentManager := agent.NewAgentManager(alpacaClient, repo, agentRepo, portfolioRepo)
	portfolioManager := portfolio.NewPortfolioAgentManager(repo, alpacaClient)
	portfolioOrchestrator := portfolio.NewPortfolioOrchestrator(
		portfolioManager,
		repo,
		agentRepo,
		portfolioRepo,
		alpacaClient,
		alpacaTradingClient,
		cfg.PortfolioExecMode,
		portfolio.DefaultStrategyCatalogConfig(),
	)

	h := api.NewHandler(repo, agentManager, agentRepo, portfolioRepo, portfolioOrchestrator, alpacaClient)
	r := api.NewRouter(h, cfg)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", cfg.HTTPAddr).Msg("starting API server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	go func() {
		resumeCtx, resumeCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		result, err := portfolioOrchestrator.ResumeActiveRuns(resumeCtx)
		resumeCancel()
		if err != nil {
			log.Error().Err(err).Int("resumed", result.Resumed).Int("skipped", result.Skipped).Int("failed", result.Failed).Msg("portfolio agent resume completed with failures")
		} else if result.Resumed > 0 || result.Skipped > 0 {
			log.Info().Int("resumed", result.Resumed).Int("skipped", result.Skipped).Msg("portfolio agent resume completed")
		}

		reconcileCtx, reconcileCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer reconcileCancel()
		if reclaimed, err := agentRepo.MarkOrphanedAgentRunsFailed(reconcileCtx, 30*time.Minute); err != nil {
			log.Error().Err(err).Msg("orphaned non-portfolio agent run reconciliation failed")
		} else if reclaimed > 0 {
			log.Info().Int64("reclaimed", reclaimed).Msg("reclaimed orphaned non-portfolio agent runs")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("API server stopped")
}
