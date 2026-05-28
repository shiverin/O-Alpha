package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oalpha/internal/agent"
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

	// Initialize database access and tracking repositories
	repo := db.NewBarsRepository(sqlDB)
	agentRepo := db.NewAgentRepository(sqlDB)
	portfolioRepo := db.NewPortfolioRepository(sqlDB) // Injected database table layer

	// Instantiate the market data provider client link
	alpacaClient := alpaca.NewClient(cfg.AlpacaDataURL, cfg.AlpacaAPIKey, cfg.AlpacaAPISecret)

	// Initialize execution supervisor manager loops
	agentManager := agent.NewAgentManager(alpacaClient, repo)

	// Build HTTP resource coordinators
	h := api.NewHandler(repo, agentManager, agentRepo, portfolioRepo)

	// ✅ RESTORED: Passes both the handler instance and config context parameters cleanly
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
