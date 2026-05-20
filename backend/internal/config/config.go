package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	DatabaseURL    string
	RedisURL       string
	MigrationsPath string
	HTTPAddr       string

	AlpacaAPIKey    string
	AlpacaAPISecret string
	AlpacaDataURL   string

	IngestSymbols  []string
	IngestInterval string
	IngestLookback time.Duration
}

// Load reads configuration from the environment.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		RedisURL:       os.Getenv("REDIS_URL"),
		MigrationsPath: envOr("MIGRATIONS_PATH", "file://migrations"),
		HTTPAddr:       envOr("HTTP_ADDR", ":8080"),
		AlpacaAPIKey:    os.Getenv("ALPACA_API_KEY"),
		AlpacaAPISecret: os.Getenv("ALPACA_API_SECRET"),
		AlpacaDataURL:   envOr("ALPACA_DATA_URL", "https://data.alpaca.markets"),
		IngestInterval:  envOr("INGEST_INTERVAL", "1h"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	symbols := strings.TrimSpace(os.Getenv("INGEST_SYMBOLS"))
	if symbols != "" {
		for _, s := range strings.Split(symbols, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				cfg.IngestSymbols = append(cfg.IngestSymbols, strings.ToUpper(s))
			}
		}
	}

	lookback := envOr("INGEST_LOOKBACK", "17520h")
	d, err := time.ParseDuration(lookback)
	if err != nil {
		return nil, fmt.Errorf("INGEST_LOOKBACK: %w", err)
	}
	cfg.IngestLookback = d

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
