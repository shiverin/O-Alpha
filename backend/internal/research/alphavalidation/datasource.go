package alphavalidation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/oalpha/internal/db"
	"github.com/oalpha/pkg/models"
)

type DataSource interface {
	LoadBars(ctx context.Context, symbols []string, timeframe string, window ValidationWindow) (map[string][]models.Bar, error)
}

type CSVDataSource struct {
	Path string
}

func (s CSVDataSource) LoadBars(_ context.Context, symbols []string, _ string, window ValidationWindow) (map[string][]models.Bar, error) {
	return LoadBarsCSV(s.Path, symbols, window)
}

type DBDataSource struct {
	Repo *db.BarsRepository
}

func (s DBDataSource) LoadBars(ctx context.Context, symbols []string, timeframe string, window ValidationWindow) (map[string][]models.Bar, error) {
	if s.Repo == nil {
		return nil, fmt.Errorf("database repository is required")
	}
	loaded := make(map[string][]models.Bar, len(symbols))
	for _, symbol := range symbols {
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" {
			continue
		}
		bars, err := s.Repo.GetBars(ctx, symbol, timeframe, window.From, window.To)
		if err != nil {
			return nil, err
		}
		loaded[symbol] = bars
	}
	return loaded, nil
}

func ResolveValidationWindow(from, to string) (ValidationWindow, error) {
	if strings.TrimSpace(from) == "" || strings.TrimSpace(to) == "" {
		return ValidationWindow{}, fmt.Errorf("from and to dates are required")
	}
	start, err := time.Parse("2006-01-02", strings.TrimSpace(from))
	if err != nil {
		return ValidationWindow{}, fmt.Errorf("parse from date: %w", err)
	}
	end, err := time.Parse("2006-01-02", strings.TrimSpace(to))
	if err != nil {
		return ValidationWindow{}, fmt.Errorf("parse to date: %w", err)
	}
	start = start.UTC()
	end = end.UTC().Add(24*time.Hour - time.Nanosecond)
	if !end.After(start) {
		return ValidationWindow{}, fmt.Errorf("to date must be after from date")
	}
	return ValidationWindow{From: start, To: end}, nil
}
