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

	// Try running migrations normally first
	if err := db.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		// If we get an error, try force reset
		log.Warn().Msg("Migration failed, attempting force reset...")
		if err := db.ForceResetMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
			log.Fatal().Err(err).Msg("force reset migrations")
		}
		log.Info().Msg("Force reset completed successfully")
	}

	log.Info().Str("path", cfg.MigrationsPath).Msg("migrations applied")
}