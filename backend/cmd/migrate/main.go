package main

import (
	"os"
	"strings"

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

	// Try running migrations normally first
	if err := db.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		// If we get a dirty database error, reset and retry
		if strings.Contains(err.Error(), "Dirty database version") {
			log.Warn().Msg("Detected dirty database, resetting migrations...")
			if err := db.ResetMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
				log.Fatal().Err(err).Msg("reset migrations")
			}
			log.Info().Msg("Reset completed successfully")
		} else {
			log.Fatal().Err(err).Msg("run migrations")
		}
	}

	log.Info().Str("path", cfg.MigrationsPath).Msg("migrations applied")
}