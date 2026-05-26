package main

import (
	"os"

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

	// Just run migrations. If they fail, stop immediately.
	if err := db.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		log.Fatal().Err(err).Msg("Migration failed")
	}

	log.Info().Str("path", cfg.MigrationsPath).Msg("migrations applied")
}