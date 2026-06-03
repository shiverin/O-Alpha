package cointegration

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type PairCandidate struct {
	SymbolY                   string                 `json:"symbol_y"`
	SymbolX                   string                 `json:"symbol_x"`
	Timeframe                 string                 `json:"timeframe"`
	FormationStart            time.Time              `json:"formation_start"`
	FormationEnd              time.Time              `json:"formation_end"`
	Correlation               float64                `json:"correlation"`
	EngleGrangerPValue        float64                `json:"engle_granger_pvalue"`
	JohansenTraceStat         float64                `json:"johansen_trace_stat"`
	HalfLifeBars              float64                `json:"half_life_bars"`
	Hurst                     float64                `json:"hurst"`
	AvgSpreadBps              float64                `json:"avg_spread_bps"`
	EstimatedRoundTripCostBps float64                `json:"estimated_round_trip_cost_bps"`
	Approved                  bool                   `json:"approved"`
	Status                    string                 `json:"status"`
	NearIdenticalETF          bool                   `json:"near_identical_etf"`
	Metadata                  map[string]interface{} `json:"metadata,omitempty"`
}

type PairSelectionConfig struct {
	MaxEngleGrangerPValue  float64
	MinCorrelation         float64
	MinHalfLifeBars        float64
	MaxHalfLifeBars        float64
	MinMoveCostMultiple    float64
	RejectNearIdenticalETF bool
}

type PairCandidateRepository interface {
	LatestApprovedPair(ctx context.Context, symbolY, symbolX, timeframe string) (*PairCandidate, error)
}

type StaticPairCandidateRepository struct {
	Candidates []PairCandidate
	Config     PairSelectionConfig
}

func DefaultPairSelectionConfig() PairSelectionConfig {
	return PairSelectionConfig{
		MaxEngleGrangerPValue:  0.01,
		MinCorrelation:         0.70,
		MinHalfLifeBars:        2,
		MaxHalfLifeBars:        30,
		MinMoveCostMultiple:    3,
		RejectNearIdenticalETF: true,
	}
}

func (r StaticPairCandidateRepository) LatestApprovedPair(ctx context.Context, symbolY, symbolX, timeframe string) (*PairCandidate, error) {
	_ = ctx
	cfg := r.Config.withDefaults()
	symbolY = strings.ToUpper(strings.TrimSpace(symbolY))
	symbolX = strings.ToUpper(strings.TrimSpace(symbolX))
	var best *PairCandidate
	for i := range r.Candidates {
		candidate := r.Candidates[i].normalized()
		if candidate.SymbolY != symbolY || candidate.SymbolX != symbolX || candidate.Timeframe != timeframe {
			continue
		}
		if ok, _ := candidate.Passes(cfg); !ok {
			continue
		}
		if best == nil || candidate.FormationEnd.After(best.FormationEnd) {
			copyCandidate := candidate
			best = &copyCandidate
		}
	}
	if best == nil {
		return nil, fmt.Errorf("no approved pair candidate found for %s/%s timeframe=%s", symbolY, symbolX, timeframe)
	}
	return best, nil
}

func (c PairCandidate) Passes(cfg PairSelectionConfig) (bool, string) {
	cfg = cfg.withDefaults()
	if !c.Approved || c.Status != "approved" {
		return false, "pair_not_approved"
	}
	if c.Correlation < cfg.MinCorrelation {
		return false, "correlation_below_min"
	}
	if c.EngleGrangerPValue > cfg.MaxEngleGrangerPValue {
		return false, "cointegration_pvalue_above_max"
	}
	if c.HalfLifeBars < cfg.MinHalfLifeBars || c.HalfLifeBars > cfg.MaxHalfLifeBars {
		return false, "half_life_outside_bounds"
	}
	if c.EstimatedRoundTripCostBps > 0 && c.AvgSpreadBps < cfg.MinMoveCostMultiple*c.EstimatedRoundTripCostBps {
		return false, "spread_move_not_cost_feasible"
	}
	if cfg.RejectNearIdenticalETF && c.NearIdenticalETF {
		return false, "near_identical_etf_rejected"
	}
	return true, ""
}

func (c PairCandidate) normalized() PairCandidate {
	c.SymbolY = strings.ToUpper(strings.TrimSpace(c.SymbolY))
	c.SymbolX = strings.ToUpper(strings.TrimSpace(c.SymbolX))
	if c.Status == "" && c.Approved {
		c.Status = "approved"
	}
	return c
}

func (c PairSelectionConfig) withDefaults() PairSelectionConfig {
	defaults := DefaultPairSelectionConfig()
	if c.MaxEngleGrangerPValue <= 0 {
		c.MaxEngleGrangerPValue = defaults.MaxEngleGrangerPValue
	}
	if c.MinCorrelation <= 0 {
		c.MinCorrelation = defaults.MinCorrelation
	}
	if c.MinHalfLifeBars <= 0 {
		c.MinHalfLifeBars = defaults.MinHalfLifeBars
	}
	if c.MaxHalfLifeBars <= 0 {
		c.MaxHalfLifeBars = defaults.MaxHalfLifeBars
	}
	if c.MinMoveCostMultiple <= 0 {
		c.MinMoveCostMultiple = defaults.MinMoveCostMultiple
	}
	if !c.RejectNearIdenticalETF {
		c.RejectNearIdenticalETF = defaults.RejectNearIdenticalETF
	}
	return c
}
