package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	DatabaseURL        string
	RedisURL           string
	MigrationsPath     string
	HTTPAddr           string
	JWTSecret          string
	CORSAllowedOrigins []string

	AlpacaAPIKey    string
	AlpacaAPISecret string
	AlpacaDataURL   string

	IngestSymbols       []string
	IngestInterval      string
	IngestLookback      time.Duration
	IngestForceBackfill bool
	IngestRunOnce       bool
}

// Load reads configuration from the environment.
func Load() (*Config, error) {
	_ = godotenv.Load(".env", "../.env")

	cfg := &Config{
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		MigrationsPath:      envOr("MIGRATIONS_PATH", "file://migrations"),
		HTTPAddr:            envOr("HTTP_ADDR", ":8080"),
		JWTSecret:           envOr("JWT_SECRET", "dev-change-me"),
		CORSAllowedOrigins:  splitCSV(envOr("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000,http://[::1]:3000,https://o-alpha-tan.vercel.app")),
		AlpacaAPIKey:        os.Getenv("ALPACA_API_KEY"),
		AlpacaAPISecret:     os.Getenv("ALPACA_API_SECRET"),
		AlpacaDataURL:       envOr("ALPACA_DATA_URL", "https://data.alpaca.markets"),
		IngestInterval:      envOr("INGEST_INTERVAL", "1h"),
		IngestForceBackfill: envBool("INGEST_FORCE_BACKFILL"),
		IngestRunOnce:       envBool("INGEST_RUN_ONCE"),
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

func splitCSV(value string) []string {
	var out []string
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "t", "yes", "y", "on":
		return true
	default:
		return false
	}
}
